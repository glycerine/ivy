/**
 * IvyApp - Main application for the Ivy Interactive Verifier web UI.
 *
 * Orchestrates the API, graph views, and UI controls. Handles all
 * user interactions including ARG/concept graph clicks, context menus,
 * file loading, mode selection, and verification checks.
 */

'use strict';

class IvyApp {
    constructor() {
        this.api = new IvyAPI();
        this.controls = new IvyControls(this.api);
        this.argGraph = null;
        this.conceptGraph = null;
        this.selectedArgNode = null;
        // Edge/label visibility state, matching Python's edge_display_checkboxes.
        // Keys: edgeName, values: {all_to_all: bool, edge_unknown: bool, none_to_none: bool, transitive: bool}
        // Default: all false (edges hidden until checkbox is checked).
        this._edgeVisibility = {};
    }

    /**
     * Initialize the application: create session, build graphs, wire events.
     */
    async init() {
        this.controls.setStatus('Initializing...');

        try {
            await this.api.createSession();
            var sessionEl = document.getElementById('session-id');
            if (sessionEl) {
                sessionEl.textContent = 'Session: ' + this.api.sessionId;
            }
            // Set session ID in URL hash for multi-tab support
            IvyPersist.setSessionIdInURL(this.api.sessionId);
        } catch (e) {
            this.controls.setStatus('Failed to create session: ' + e.message, 'error');
            console.error('Session creation failed:', e);
            // Continue anyway - graphs can still be created for when server comes up
        }

        // Create Cytoscape graph instances
        this.argGraph = new IvyGraph('arg-graph', ARG_STYLE);
        this.conceptGraph = new IvyGraph('concept-graph', CONCEPT_STYLE);

        // Health check: verify graphs initialized correctly.
        // Catches silent failures from bad stylesheet data() mappers.
        this.argGraph.healthCheck();
        this.conceptGraph.healthCheck();

        // Hook concept graph updates to auto-apply edge visibility.
        // Matches Python: edges are hidden by default, shown only when
        // the corresponding checkbox (+/?/-) in the state panel is checked.
        var self = this;
        var origUpdate = this.conceptGraph.update.bind(this.conceptGraph);
        this.conceptGraph.update = function(elements, positions) {
            origUpdate(elements, positions);
            self._applyEdgeVisibility();
        };

        // Wire up all event handlers
        this.setupEventHandlers();
        this.setupResizer();
        this.setupResizer2();
        this.setupKeyboardShortcuts();

        // Connect to SSE for real-time updates
        if (this.api.sessionId) {
            this.api.connectEvents(this.handleEvent.bind(this));
        }

        // Try to restore state from a previous session (survives page reload).
        // URL hash takes priority: allows multiple tabs with different sessions.
        var savedState = IvyPersist.load();
        if (savedState && savedState.fileContent) {
            console.log('IvyPersist: restoring session', savedState.sessionId, savedState.fileName);
            var restored = await IvyPersist.restore(this, savedState);
            if (restored) {
                IvyPersist.setSessionIdInURL(this.api.sessionId);
                IvyPersist.setFileName(savedState.fileName);
                this.controls.setStatus('Restored: ' + (savedState.fileName || 'session'), 'success');
            } else {
                this.controls.setStatus('Ready');
            }
        } else {
            this.controls.setStatus('Ready');
        }

        // Auto-save: on beforeunload (catches reload, tab close, navigation)
        // and after any successful operation (debounced).
        var self = this;
        window.addEventListener('beforeunload', function () {
            IvyPersist.save(self);
        });

        // Hook into setStatus: auto-save whenever a 'success' status is set.
        var origSetStatus = this.controls.setStatus.bind(this.controls);
        this.controls.setStatus = function (msg, level) {
            origSetStatus(msg, level);
            if (level === 'success') {
                IvyPersist.save(self);
            }
        };
    }

    /**
     * Set up all DOM and graph event handlers.
     */
    setupEventHandlers() {
        var self = this;

        // --- File Menu ---
        var fileInput = document.getElementById('file-input');

        // File > Load...
        document.getElementById('file-load').addEventListener('click', function (e) {
            e.preventDefault();
            self.closeAllDropdowns(e);
            fileInput.click();
        });
        fileInput.addEventListener('change', function () {
            if (fileInput.files.length > 0) {
                self.loadFile(fileInput.files[0]);
                fileInput.value = ''; // reset for re-selection of same file
            }
        });

        // File > Save as... (uses File System Access API to write to a chosen path)
        document.getElementById('file-save-as').addEventListener('click', function (e) {
            e.preventDefault();
            self.closeAllDropdowns(e);
            self.saveAs();
        });

        // File > Download current model (browser download)
        document.getElementById('file-download').addEventListener('click', function (e) {
            e.preventDefault();
            self.closeAllDropdowns(e);
            self.downloadModel();
        });

        // File > New Model
        document.getElementById('file-new').addEventListener('click', function (e) {
            e.preventDefault();
            self.closeAllDropdowns(e);
            self.newModel();
        });

        // File > Save Invariant
        document.getElementById('file-save-invariant').addEventListener('click', function (e) {
            e.preventDefault();
            self.closeAllDropdowns(e);
            self.saveInvariant();
        });

        // --- Check ---
        document.getElementById('btn-check').addEventListener('click', function () {
            self.runCheck();
        });

        // --- Undo ---
        document.getElementById('btn-undo').addEventListener('click', function () {
            self.doUndo();
        });

        // --- Reset Domain ---
        document.getElementById('btn-reset-domain').addEventListener('click', function () {
            self.resetDomain();
        });

        // --- Diagram Domain ---
        document.getElementById('btn-diagram-domain').addEventListener('click', function () {
            self.diagramDomain();
        });

        // --- Dropdown Menus (panel header) ---
        this.setupDropdownMenus();

        // --- ARG Panel Menu Items (File, Invariant) ---
        this.bindMenuAction('arg-save-abs', function () { self.saveAbstraction(); });
        this.bindMenuAction('arg-check-induction', function () { self.checkInduction(); });
        this.bindMenuAction('arg-bounded-check', function () { self.boundedCheck(); });
        this.bindMenuAction('arg-diagram', function () { self.diagramDomain(); });
        this.bindMenuAction('arg-weaken', function () { self.weakenInvariant(); });
        this.bindMenuAction('arg-save-invariant', function () { self.saveInvariant(); });

        // --- Concept Panel Menu Items (Conjecture, View) ---
        this.bindMenuAction('conj-undo', function () { self.doUndo(); });
        this.bindMenuAction('conj-redo', function () { self.doRedo(); });
        this.bindMenuAction('conj-pdr-step', function () { self.pdrStep(); });
        this.bindMenuAction('conj-concrete', function () { self.concreteStep(); });
        this.bindMenuAction('conj-gather', function () { self.gatherFacts(); });
        this.bindMenuAction('conj-reverse', function () { self.reverseStep(); });
        this.bindMenuAction('conj-path-reach', function () { self.pathReach(); });
        this.bindMenuAction('conj-reach', function () { self.reachStep(); });
        this.bindMenuAction('conj-conjecture', function () { self.makeConjecture(); });
        this.bindMenuAction('conj-backtrack', function () { self.backtrack(); });
        this.bindMenuAction('conj-recalculate', function () { self.recalculateGraph(); });
        this.bindMenuAction('conj-diagram', function () { self.diagramDomain(); });
        this.bindMenuAction('conj-remember', function () { self.rememberGraph(); });
        this.bindMenuAction('conj-export', function () { self.exportConjecture(); });
        this.bindMenuAction('view-add-relation', function () { self.addRelationFromString(); });

        // --- Click anywhere to dismiss context menu and dropdowns ---
        document.addEventListener('click', function (e) {
            self.controls.hideContextMenu();
            self.closeAllDropdowns(e);
        });

        // --- Prevent browser context menu on graph containers ---
        document.getElementById('arg-graph').addEventListener('contextmenu', function (e) {
            e.preventDefault();
        });
        document.getElementById('concept-graph').addEventListener('contextmenu', function (e) {
            e.preventDefault();
        });

        // --- ARG Graph Events ---

        // ARG node left-click: view state in concept graph
        this.argGraph.onNodeClick(function (nodeData) {
            self.onArgNodeClick(nodeData);
        });

        // ARG node right-click: context menu
        this.argGraph.onNodeRightClick(function (nodeData, pos) {
            self.onArgNodeRightClick(nodeData, pos);
        });

        // ARG edge click: show info
        this.argGraph.onEdgeClick(function (edgeData) {
            self.controls.showInfo(edgeData.short_info, edgeData.long_info);
        });

        // ARG edge right-click: context menu (matches Python ivy_ui.py get_edge_actions)
        this.argGraph.onEdgeRightClick(function (edgeData, pos) {
            self.onArgEdgeRightClick(edgeData, pos);
        });

        // ARG background click: clear info
        this.argGraph.onBackgroundClick(function () {
            self.controls.clearInfo();
            self.controls.hideContextMenu();
        });

        // --- Concept Graph Events ---

        // Concept node right-click: context menu (split, empty, remove, materialize)
        this.conceptGraph.onNodeRightClick(function (nodeData, pos) {
            self.onConceptNodeRightClick(nodeData, pos);
        });

        // Concept node left-click: toggle selection independently per node.
        // Matches Python Tk: click selects (fills interior gray), click again deselects (white).
        this.conceptGraph.onNodeClick(function (nodeData, evt) {
            try {
                var node = evt.target;
                var name = nodeData.obj || nodeData.id;
                if (node.hasClass('selected_node')) {
                    node.removeClass('selected_node');
                    self.controls.setStatus('Deselected: ' + name);
                    self.controls.clearInfo();
                } else {
                    node.addClass('selected_node');
                    self.selectedConceptNode = name;
                    self.controls.setStatus('Selected: ' + name);
                    self.controls.showInfo(nodeData.short_info, nodeData.long_info);
                }
            } catch (e) {
                console.error('concept node click error:', e);
            }
        });

        // Concept edge left-click: show info
        this.conceptGraph.onEdgeClick(function (edgeData) {
            self.controls.showInfo(edgeData.short_info, edgeData.long_info);
        });

        // Concept edge right-click: context menu (remove, materialize +/-)
        this.conceptGraph.onEdgeRightClick(function (edgeData, pos) {
            self.onConceptEdgeRightClick(edgeData, pos);
        });

        // Concept background click: clear
        this.conceptGraph.onBackgroundClick(function () {
            self.controls.clearInfo();
            self.controls.hideContextMenu();
        });
    }

    /**
     * Set up the resizable divider between ARG and concept panels.
     */
    setupResizer() {
        var divider = document.getElementById('divider');
        var argPanel = document.getElementById('arg-panel');
        var container = document.getElementById('main-container');
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;

        divider.addEventListener('mousedown', function (e) {
            isDragging = true;
            startX = e.clientX;
            startWidth = argPanel.offsetWidth;
            divider.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = e.clientX - startX;
            var newWidth = startWidth + dx;
            var containerWidth = container.offsetWidth;
            // Clamp between 150px and container - 200px
            newWidth = Math.max(150, Math.min(newWidth, containerWidth - 200));
            argPanel.style.flex = '0 0 ' + newWidth + 'px';
            // Trigger resize on graphs
            self.argGraph.resize();
            self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                divider.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                // Final resize
                self.argGraph.resize();
                self.conceptGraph.resize();
            }
        });
    }

    /**
     * Set up the resizable second divider between concept and state panels.
     */
    setupResizer2() {
        var divider2 = document.getElementById('divider2');
        if (!divider2) return;
        var statePanel = document.getElementById('state-panel');
        var container = document.getElementById('main-container');
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;

        divider2.addEventListener('mousedown', function (e) {
            isDragging = true;
            startX = e.clientX;
            startWidth = statePanel.offsetWidth;
            divider2.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = startX - e.clientX; // reversed: drag left = wider
            var newWidth = startWidth + dx;
            newWidth = Math.max(140, Math.min(newWidth, container.offsetWidth - 300));
            statePanel.style.flex = '0 0 ' + newWidth + 'px';
            self.argGraph.resize();
            self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                divider2.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                self.argGraph.resize();
                self.conceptGraph.resize();
            }
        });
    }

    /**
     * Populate the state checkbox table (right pane) with edge/relation names.
     * Each row has checkboxes for: + (all_to_all), ? (unknown), - (none_to_none), T (transitive)
     * and the relation name.
     */
    populateStateCheckboxes(conceptData) {
        var tbody = document.getElementById('state-checkbox-body');
        if (!tbody) return;
        tbody.innerHTML = '';

        // Use the relations list from the server (edges + node_labels).
        // This matches Python's Graph.relation_ids.
        var names = [];
        if (conceptData && conceptData.relations) {
            names = conceptData.relations.slice().sort();
        }
        var self = this;

        for (var i = 0; i < names.length; i++) {
            (function(name) {
                var tr = document.createElement('tr');

                // + checkbox (all_to_all)
                var td1 = document.createElement('td');
                var cb1 = document.createElement('input');
                cb1.type = 'checkbox';
                cb1.title = 'Show definite edges (' + name + ')';
                cb1.addEventListener('change', function() {
                    self.onEdgeToggle(name, 'all_to_all', cb1.checked);
                });
                td1.appendChild(cb1);
                tr.appendChild(td1);

                // ? checkbox (unknown)
                var td2 = document.createElement('td');
                var cb2 = document.createElement('input');
                cb2.type = 'checkbox';
                cb2.title = 'Show unknown edges (' + name + ')';
                cb2.addEventListener('change', function() {
                    self.onEdgeToggle(name, 'edge_unknown', cb2.checked);
                });
                td2.appendChild(cb2);
                tr.appendChild(td2);

                // - checkbox (none_to_none)
                var td3 = document.createElement('td');
                var cb3 = document.createElement('input');
                cb3.type = 'checkbox';
                cb3.title = 'Show absent edges (' + name + ')';
                cb3.addEventListener('change', function() {
                    self.onEdgeToggle(name, 'none_to_none', cb3.checked);
                });
                td3.appendChild(cb3);
                tr.appendChild(td3);

                // T checkbox (transitive reduction)
                var td4 = document.createElement('td');
                var cb4 = document.createElement('input');
                cb4.type = 'checkbox';
                cb4.title = 'Transitive reduction (' + name + ')';
                cb4.addEventListener('change', function() {
                    self.onEdgeToggle(name, 'transitive', cb4.checked);
                });
                td4.appendChild(cb4);
                tr.appendChild(td4);

                // Name column
                var td5 = document.createElement('td');
                td5.className = 'name-col';
                var a = document.createElement('a');
                a.textContent = name;
                a.href = '#';
                a.addEventListener('click', function(e) {
                    e.preventDefault();
                    // Clicking the name could highlight related edges
                });
                td5.appendChild(a);
                tr.appendChild(td5);

                tbody.appendChild(tr);
            })(names[i]);
        }

        // If no edges found, show a placeholder
        if (names.length === 0 && conceptData) {
            var tr = document.createElement('tr');
            var td = document.createElement('td');
            td.colSpan = 5;
            td.style.color = '#666';
            td.style.fontStyle = 'italic';
            td.textContent = 'No relations loaded';
            tr.appendChild(td);
            tbody.appendChild(tr);
        }
    }

    /**
     * Handle edge visibility toggle change.
     */
    onEdgeToggle(edgeName, displayClass, checked) {
        // Track visibility state client-side (matches Python edge_display_checkboxes)
        if (!this._edgeVisibility[edgeName]) {
            this._edgeVisibility[edgeName] = {
                all_to_all: false, edge_unknown: false, none_to_none: false, transitive: false
            };
        }
        this._edgeVisibility[edgeName][displayClass] = checked;

        // Apply visibility to concept graph edges immediately (no server round-trip).
        this._applyEdgeVisibility();

        // Also inform server for persistence
        this.api.setToggles({
            edge: edgeName,
            display_class: displayClass,
            value: checked
        }).catch(function(e) {
            console.error('Toggle error:', e);
        });
    }

    /**
     * Apply edge visibility based on checkbox state.
     * An edge is shown if its display class checkbox is checked.
     * Matches Python cy_render.py line 346:
     *   if widget.edge_display_checkboxes[edge][classes[0]].value is False: skip
     */
    _applyEdgeVisibility() {
        if (!this.conceptGraph || !this.conceptGraph.cy) return;
        var self = this;
        this.conceptGraph.cy.edges().forEach(function(edge) {
            var edgeName = edge.data('obj') || edge.data('label') || '';
            var vis = self._edgeVisibility[edgeName];
            if (!vis) {
                // No checkbox state → hide by default (matching Python)
                edge.style('display', 'none');
                return;
            }
            // Edge class determines which checkbox controls it
            var classes = (edge.classes() || '').split(' ');
            var show = false;
            for (var i = 0; i < classes.length; i++) {
                var cls = classes[i].trim();
                if (cls && vis[cls]) {
                    show = true;
                    break;
                }
            }
            // Also show if edge_unknown is checked and edge has no specific class
            if (!show && vis.edge_unknown) {
                if (classes.indexOf('edge_unknown') >= 0 || classes.length === 0) {
                    show = true;
                }
            }
            edge.style('display', show ? 'element' : 'none');
        });
    }

    /**
     * Update the state label to show which ARG node is selected.
     */
    updateStateLabel(nodeId) {
        var label = document.getElementById('state-label');
        if (label) {
            label.textContent = 'State: ' + (nodeId != null ? nodeId : '—');
        }
    }

    /**
     * Set up keyboard shortcuts.
     */
    setupKeyboardShortcuts() {
        var self = this;
        document.addEventListener('keydown', function (e) {
            // Ctrl+Z or Cmd+Z: Undo
            if ((e.ctrlKey || e.metaKey) && e.key === 'z') {
                e.preventDefault();
                self.doUndo();
            }
            // Escape: hide context menu
            if (e.key === 'Escape') {
                self.controls.hideContextMenu();
            }
        });
    }

    // ================================================================
    // ARG Graph Interactions
    // ================================================================

    /**
     * Handle left-click on an ARG node: load its concept graph.
     */
    async onArgNodeClick(nodeData) {
        this.selectedArgNode = nodeData.id;
        this.argGraph.highlightNode(nodeData.id);
        this.controls.showInfo(nodeData.short_info, nodeData.long_info);
        this.updateStateLabel(nodeData.label || nodeData.id);
        this.controls.setStatus('Loading concept graph for state ' + (nodeData.label || nodeData.id) + '...');

        try {
            var result = await this.api.getConceptGraph(nodeData.obj || nodeData.id);
            if (result && result.elements) {
                this.conceptGraph.update(result.elements, result.positions);
            }
            // Build toggles if toggle info is provided
            if (result && result.edge_names) {
                this.controls.buildEdgeToggles(result.edge_names, this.onEdgeToggleChange.bind(this));
            }
            if (result && result.label_names) {
                this.controls.buildLabelToggles(result.label_names, this.onLabelToggleChange.bind(this));
            }
            this.controls.setStatus('Viewing state ' + (nodeData.label || nodeData.id));
        } catch (e) {
            this.controls.setStatus('Error loading concept graph: ' + e.message, 'error');
            console.error('Concept graph load error:', e);
        }
    }

    /**
     * Handle right-click on an ARG node: show context menu.
     * Actions: view state, execute action, mark, cover, safety check.
     */
    onArgNodeRightClick(nodeData, pos) {
        var self = this;
        var actions = [
            { header: 'State ' + (nodeData.label || nodeData.id) },
            {
                name: 'View State',
                id: 'view_state',
                callback: function () { self.onArgNodeClick(nodeData); },
            },
            { separator: true },
        ];

        // Add server-provided actions if available
        if (nodeData.actions && Array.isArray(nodeData.actions)) {
            for (var i = 0; i < nodeData.actions.length; i++) {
                var act = nodeData.actions[i];
                (function (action) {
                    actions.push({
                        name: action[0] || action.name,
                        id: action.id || action[0],
                        callback: function () {
                            self.executeArgNodeAction(nodeData, action);
                        },
                    });
                })(act);
            }
        } else {
            // Default ARG node actions — matches Python ivy_ui.py node_commands()
            var defaultActions = [
                { name: 'Check safety', id: 'check_safety' },
                { name: 'Extend', id: 'extend' },
                { name: 'Mark', id: 'mark' },
                { name: 'Cover by marked', id: 'cover' },
                { name: 'Join with marked', id: 'join' },
                { name: 'Try conjecture', id: 'try_conjecture' },
                { name: 'Try remembered goal', id: 'try_remembered' },
                { name: 'Delete', id: 'delete' },
            ];
            for (var j = 0; j < defaultActions.length; j++) {
                (function (act) {
                    actions.push({
                        name: act.name,
                        id: act.id,
                        callback: function () {
                            self.executeArgNodeAction(nodeData, act);
                        },
                    });
                })(defaultActions[j]);
            }
        }

        // Offset from the graph container position
        var graphContainer = document.getElementById('arg-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Execute an ARG node action via the API.
     */
    async executeArgNodeAction(nodeData, action) {
        var actionName = action.id || action[0] || action.name;
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            var result = await this.api.argNodeAction(nodeData.obj || nodeData.id, actionName);
            if (result && result.arg) {
                this.argGraph.update(result.arg.elements, result.arg.positions);
            }
            if (result && result.concept) {
                this.conceptGraph.update(result.concept.elements, result.concept.positions);
            }
            this.controls.setStatus('Action complete: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Action failed: ' + e.message, 'error');
            console.error('ARG action error:', e);
        }
    }

    /**
     * Handle right-click on an ARG edge: show context menu.
     * Matches Python ivy_ui.py get_edge_actions: Dismiss, Recalculate, Step in, View Source.
     */
    onArgEdgeRightClick(edgeData, pos) {
        var self = this;
        var label = edgeData.label || edgeData.obj || '';
        var actions = [
            { header: 'Transition: ' + label },
            {
                name: 'Dismiss',
                id: 'dismiss',
                callback: function () { self.controls.hideContextMenu(); },
            },
            {
                name: 'Recalculate',
                id: 'recalculate_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'recalculate'); },
            },
            {
                name: 'Step in',
                id: 'decompose_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'decompose'); },
            },
            {
                name: 'View Source',
                id: 'view_source_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'view_source'); },
            },
        ];
        this.controls.showContextMenu(pos.x, pos.y, actions);
    }

    /**
     * Execute an ARG edge action via the API.
     */
    async executeArgEdgeAction(edgeData, actionName) {
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            var result = await this.api.argNodeAction(
                edgeData.source_obj || edgeData.source || edgeData.obj,
                actionName,
                { target: edgeData.target_obj || edgeData.target }
            );
            if (result && result.arg) {
                this.argGraph.update(result.arg.elements, result.arg.positions);
            }
            if (result && result.source) {
                // View Source: show source code in info panel
                this.controls.showInfo('Source: ' + (result.file || ''), result.source);
            }
            this.controls.setStatus('Done: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Edge action failed: ' + e.message, 'error');
            console.error('ARG edge action error:', e);
        }
    }

    // ================================================================
    // Concept Graph Interactions
    // ================================================================

    /**
     * Handle right-click on a concept node: show context menu.
     * Actions: split by X (for each possible split), suppose empty, remove, materialize.
     */
    onConceptNodeRightClick(nodeData, pos) {
        var self = this;
        var actions = [
            { header: nodeData.label || nodeData.obj || nodeData.id },
        ];

        // Server-provided actions (split by X, remove, empty, materialize, add projection)
        if (nodeData.actions && Array.isArray(nodeData.actions)) {
            for (var i = 0; i < nodeData.actions.length; i++) {
                var act = nodeData.actions[i];
                (function (action) {
                    var actionName = action[0] || action.name;
                    actions.push({
                        name: actionName,
                        id: actionName,
                        callback: function () {
                            self.executeConceptNodeAction(nodeData, action);
                        },
                    });
                })(act);
            }
        } else {
            // Default concept node actions — matches Python tk_graph_ui.py
            var conceptId = nodeData.obj || nodeData.id;
            var isSelected = false;
            if (self.conceptGraph && self.conceptGraph.cy) {
                var nd = self.conceptGraph.cy.nodes().filter(function (n) {
                    return n.data('obj') === conceptId || n.id() === conceptId;
                });
                isSelected = nd.length > 0 && nd.hasClass('selected_node');
            }
            actions.push({
                name: isSelected ? 'Unselect' : 'Select',
                id: 'select',
                callback: function () {
                    self.selectConceptNode(conceptId);
                },
            });
            actions.push({
                name: 'Suppose Empty',
                id: 'suppose_empty',
                callback: function () {
                    self.supposeEmpty(nodeData.obj || nodeData.id);
                },
            });
            actions.push({
                name: 'Materialize',
                id: 'materialize',
                callback: function () {
                    self.materializeNode(nodeData.obj || nodeData.id);
                },
            });
            actions.push({
                name: 'Splatter',
                id: 'splatter',
                callback: function () {
                    self.splatterNode(nodeData.obj || nodeData.id);
                },
            });
            actions.push({ separator: true });
            actions.push({
                name: 'Remove',
                id: 'remove',
                callback: function () {
                    self.removeConcept(nodeData.obj || nodeData.id);
                },
            });
        }

        var graphContainer = document.getElementById('concept-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Handle right-click on a concept edge: show context menu.
     * Actions: remove, materialize +, materialize -.
     */
    onConceptEdgeRightClick(edgeData, pos) {
        var self = this;
        var conceptId = edgeData.obj || edgeData.id;
        var actions = [
            { header: edgeData.label || conceptId },
        ];

        if (edgeData.actions && Array.isArray(edgeData.actions)) {
            for (var i = 0; i < edgeData.actions.length; i++) {
                var act = edgeData.actions[i];
                (function (action) {
                    var actionName = action[0] || action.name;
                    actions.push({
                        name: actionName,
                        id: actionName,
                        callback: function () {
                            self.executeConceptEdgeAction(edgeData, action);
                        },
                    });
                })(act);
            }
        } else {
            actions.push({
                name: 'Remove',
                id: 'remove',
                callback: function () { self.removeConcept(conceptId); },
            });
            actions.push({
                name: 'Materialize +',
                id: 'materialize_pos',
                callback: function () { self.materializeEdge(conceptId, true); },
            });
            actions.push({
                name: 'Materialize \u2013',
                id: 'materialize_neg',
                callback: function () { self.materializeEdge(conceptId, false); },
            });
            actions.push({ separator: true });
            actions.push({
                name: 'Suppose Empty',
                id: 'empty_edge',
                callback: function () { self.supposeEmpty(conceptId); },
            });
            actions.push({
                name: 'Dematerialize',
                id: 'dematerialize',
                callback: function () {
                    self.executeConceptEdgeAction(edgeData, { id: 'dematerialize' });
                },
            });
        }

        var graphContainer = document.getElementById('concept-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Execute a concept node action dispatched from the context menu.
     */
    async executeConceptNodeAction(nodeData, action) {
        var actionName = (action[0] || action.name || '').toLowerCase();

        if (actionName === 'remove') {
            return this.removeConcept(nodeData.obj || nodeData.id);
        }
        if (actionName === 'suppose empty') {
            return this.supposeEmpty(nodeData.obj || nodeData.id);
        }
        if (actionName === 'materialize') {
            return this.materializeNode(nodeData.obj || nodeData.id);
        }
        if (actionName.indexOf('split by ') === 0) {
            var splitBy = actionName.substring('split by '.length);
            return this.splitConcept(nodeData.obj || nodeData.id, splitBy);
        }
        if (actionName.indexOf('add ') === 0) {
            var projName = actionName.substring('add '.length);
            return this.addProjection(projName, nodeData.obj || nodeData.id);
        }

        // Fallback: generic action execution
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            var result = await this.api.executeAction(actionName, {
                concept: nodeData.obj || nodeData.id,
            });
            await this.refreshConceptGraph();
            this.controls.setStatus('Done: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Error: ' + e.message, 'error');
        }
    }

    /**
     * Execute a concept edge action dispatched from the context menu.
     */
    async executeConceptEdgeAction(edgeData, action) {
        var actionName = (action[0] || action.name || '').toLowerCase();
        var conceptId = edgeData.obj || edgeData.id;

        if (actionName === 'remove') {
            return this.removeConcept(conceptId);
        }
        if (actionName === 'materialize +') {
            return this.materializeEdge(conceptId, true);
        }
        if (actionName === 'materialize -' || actionName === 'materialize \u2013') {
            return this.materializeEdge(conceptId, false);
        }

        // Fallback
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            await this.api.executeAction(actionName, { concept: conceptId });
            await this.refreshConceptGraph();
            this.controls.setStatus('Done: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Error: ' + e.message, 'error');
        }
    }

    // ================================================================
    // Concept Domain Operations
    // ================================================================

    async splitConcept(concept, splitBy) {
        this.controls.setStatus('Splitting ' + concept + ' by ' + splitBy + '...');
        try {
            await this.api.splitConcept(concept, splitBy);
            await this.refreshConceptGraph();
            this.controls.setStatus('Split complete', 'success');
        } catch (e) {
            this.controls.setStatus('Split failed: ' + e.message, 'error');
        }
    }

    async supposeEmpty(concept) {
        this.controls.setStatus('Supposing ' + concept + ' is empty...');
        try {
            await this.api.supposeEmpty(concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Suppose empty applied', 'success');
        } catch (e) {
            this.controls.setStatus('Suppose empty failed: ' + e.message, 'error');
        }
    }

    async removeConcept(concept) {
        this.controls.setStatus('Removing concept...');
        try {
            await this.api.removeConcept(concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Concept removed', 'success');
        } catch (e) {
            this.controls.setStatus('Remove failed: ' + e.message, 'error');
        }
    }

    async materializeNode(concept) {
        this.controls.setStatus('Materializing node...');
        try {
            await this.api.materializeNode(concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Node materialized', 'success');
        } catch (e) {
            this.controls.setStatus('Materialize failed: ' + e.message, 'error');
        }
    }

    async materializeEdge(concept, positive) {
        var dir = positive ? '+' : '\u2013';
        this.controls.setStatus('Materializing edge (' + dir + ')...');
        try {
            await this.api.materializeEdge(concept, positive);
            await this.refreshConceptGraph();
            this.controls.setStatus('Edge materialized (' + dir + ')', 'success');
        } catch (e) {
            this.controls.setStatus('Materialize failed: ' + e.message, 'error');
        }
    }

    async addProjection(name, concept) {
        this.controls.setStatus('Adding projection ' + name + '...');
        try {
            await this.api.addProjection(name, concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Projection added', 'success');
        } catch (e) {
            this.controls.setStatus('Add projection failed: ' + e.message, 'error');
        }
    }

    /**
     * Select/mark a concept node for edge materialization.
     * Matches Python tk_graph_ui.py select action.
     */
    selectConceptNode(conceptId) {
        this.selectedConceptNode = conceptId;
        // Toggle selected_node class (same visual as left-click)
        if (this.conceptGraph && this.conceptGraph.cy) {
            var node = this.conceptGraph.cy.nodes().filter(function (n) {
                return n.data('obj') === conceptId || n.id() === conceptId;
            });
            if (node.length > 0) {
                if (node.hasClass('selected_node')) {
                    node.removeClass('selected_node');
                    this.controls.setStatus('Deselected: ' + conceptId);
                } else {
                    node.addClass('selected_node');
                    this.controls.setStatus('Selected: ' + conceptId);
                }
            }
        }
    }

    /**
     * Splatter a concept node — materialize all universe elements of its sort.
     * Matches Python tk_graph_ui.py splatter action.
     */
    async splatterNode(conceptId) {
        this.controls.setStatus('Splattering ' + conceptId + '...');
        try {
            await this.api.executeAction('splatter', { concept: conceptId });
            await this.refreshConceptGraph();
            this.controls.setStatus('Splattered: ' + conceptId, 'success');
        } catch (e) {
            this.controls.setStatus('Splatter failed: ' + e.message, 'error');
        }
    }

    // ================================================================
    // Toggle Handlers
    // ================================================================

    /**
     * Called when an edge visibility toggle checkbox changes.
     */
    async onEdgeToggleChange(edgeName, className, checked) {
        try {
            var toggles = {
                type: 'edge',
                edge: edgeName,
                class: className,
                visible: checked,
            };
            await this.api.setToggles(toggles);
            await this.refreshConceptGraph();
        } catch (e) {
            console.error('Toggle update error:', e);
        }
    }

    /**
     * Called when a label visibility toggle checkbox changes.
     */
    async onLabelToggleChange(labelName, className, checked) {
        try {
            var toggles = {
                type: 'label',
                label: labelName,
                class: className,
                visible: checked,
            };
            await this.api.setToggles(toggles);
            await this.refreshConceptGraph();
        } catch (e) {
            console.error('Toggle update error:', e);
        }
    }

    // ================================================================
    // Top-Level Actions
    // ================================================================

    /**
     * Load an .ivy file.
     */
    async loadFile(file) {
        this.controls.showLoading('Loading ' + file.name + '...');
        this.controls.setStatus('Loading file: ' + file.name + '...');
        try {
            // Read file content for persistence before uploading
            var self = this;
            var reader = new FileReader();
            var contentPromise = new Promise(function (resolve) {
                reader.onload = function () { resolve(reader.result); };
                reader.readAsText(file);
            });
            var fileContent = await contentPromise;
            self._persistedFileName = file.name;
            self._persistedFilePath = file.webkitRelativePath || file.name;
            self._persistedFileContent = fileContent;

            var result = await this.api.loadFile(file);
            // Refresh ARG
            var argData = await this.api.getARG();
            if (argData && argData.elements) {
                this.argGraph.update(argData.elements, argData.positions);
            }
            // Refresh concept graph and populate state checkbox pane
            var conceptData = await this.api.getConceptGraph();
            if (conceptData && conceptData.elements) {
                this.conceptGraph.update(conceptData.elements, conceptData.positions);
            }
            this._persistedConceptRelations = conceptData;
            this.populateStateCheckboxes(conceptData);
            // Update state label and file name display
            this.updateStateLabel(0);
            IvyPersist.setFileName(file.name);
            this.controls.setStatus('Loaded: ' + file.name, 'success');

            // Auto-save after file load
            IvyPersist.save(this);
        } catch (e) {
            this.controls.setStatus('Load failed: ' + e.message, 'error');
            console.error('File load error:', e);
        } finally {
            this.controls.hideLoading();
        }
    }

    /**
     * Save invariant/conjectures to a .ivy file.
     * Matches Python ivy_ui_cti.py save_conjectures:
     * exports conjectures as "invariant [label] formula" lines.
     */
    async saveInvariant() {
        this.controls.setStatus('Saving invariant...');
        try {
            var result = await this.api.executeAction('get_conjectures', {});
            var conjectures = (result && result.conjectures) || [];

            // If the server doesn't return conjectures yet, gather from
            // the concept graph facts as a fallback.
            if (conjectures.length === 0) {
                var facts = await this.api.executeAction('gather', {});
                conjectures = (facts && facts.conjectures) || [];
            }

            // Build Ivy-format output matching Python's _write_conj
            var lines = ['# This file was generated by Ivy.\n'];
            if (conjectures.length > 0) {
                lines.push('# conjectures\n');
                for (var i = 0; i < conjectures.length; i++) {
                    var c = conjectures[i];
                    if (c.label) {
                        lines.push('invariant [' + c.label + '] ' + c.formula + '\n');
                    } else {
                        lines.push('invariant ' + c.formula + '\n');
                    }
                }
            } else {
                // No conjectures from server — export from the loaded file's
                // invariant/conjecture declarations if available.
                if (this._persistedFileContent) {
                    var srcLines = this._persistedFileContent.split('\n');
                    lines.push('# invariants from ' + (this._persistedFileName || 'model') + '\n');
                    for (var j = 0; j < srcLines.length; j++) {
                        var line = srcLines[j].trim();
                        if (line.match(/^(invariant|conjecture)\b/)) {
                            lines.push(line + '\n');
                        }
                    }
                }
                if (lines.length <= 2) {
                    lines.push('# (no invariants found)\n');
                }
            }

            var text = lines.join('');
            var suggestedName = (this._persistedFileName || 'invariant').replace(/\.ivy$/, '') + '_invariant.ivy';

            // Use File System Access API to let user choose save location
            if (window.showSaveFilePicker) {
                var handle = await window.showSaveFilePicker({
                    suggestedName: suggestedName,
                    types: [{
                        description: 'Ivy files',
                        accept: { 'text/plain': ['.ivy'] },
                    }],
                });
                var writable = await handle.createWritable();
                await writable.write(text);
                await writable.close();
                this.controls.setStatus('Invariant saved: ' + handle.name, 'success');
            } else {
                // Fallback: browser download
                var blob = new Blob([text], { type: 'text/plain' });
                var url = URL.createObjectURL(blob);
                var a = document.createElement('a');
                a.href = url;
                a.download = suggestedName;
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
                this.controls.setStatus('Invariant downloaded: ' + suggestedName, 'success');
            }
        } catch (e) {
            if (e.name === 'AbortError') {
                this.controls.setStatus('Save invariant cancelled');
            } else {
                this.controls.setStatus('Save invariant failed: ' + e.message, 'error');
            }
        }
    }

    /**
     * Download current model as a browser download.
     */
    async downloadModel() {
        this.controls.setStatus('Downloading...');
        try {
            var content = this._persistedFileContent || '';
            if (!content) {
                this.controls.setStatus('No model loaded to download', 'error');
                return;
            }
            var blob = new Blob([content], { type: 'text/plain' });
            var url = URL.createObjectURL(blob);
            var a = document.createElement('a');
            a.href = url;
            a.download = this._persistedFileName || 'model.ivy';
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            this.controls.setStatus('Downloaded: ' + a.download, 'success');
        } catch (e) {
            this.controls.setStatus('Download failed: ' + e.message, 'error');
        }
    }

    /**
     * Save as... — uses the File System Access API (showSaveFilePicker)
     * to let the user choose a disk path. Remembers the file handle
     * for subsequent saves.
     */
    async saveAs() {
        var content = this._persistedFileContent || '';
        if (!content) {
            this.controls.setStatus('No model loaded to save', 'error');
            return;
        }
        try {
            if (!window.showSaveFilePicker) {
                // Fallback for browsers without File System Access API
                this.controls.setStatus('Save as... not supported in this browser — use Download instead', 'error');
                return;
            }
            var handle = await window.showSaveFilePicker({
                suggestedName: this._persistedFileName || 'model.ivy',
                types: [{
                    description: 'Ivy files',
                    accept: { 'text/plain': ['.ivy'] },
                }],
            });
            var writable = await handle.createWritable();
            await writable.write(content);
            await writable.close();

            // Remember the handle and path for future saves
            this._fileHandle = handle;
            this._persistedFileName = handle.name;
            IvyPersist.setFileName(handle.name);
            this.controls.setStatus('Saved: ' + handle.name, 'success');
        } catch (e) {
            if (e.name === 'AbortError') {
                // User cancelled the dialog
                this.controls.setStatus('Save cancelled');
            } else {
                this.controls.setStatus('Save as failed: ' + e.message, 'error');
            }
        }
    }

    // Keep saveSession as an alias for downloadModel (used by ARG panel binding)
    async saveSession() { return this.downloadModel(); }

    /**
     * Start a new model. Saves any existing state first, then clears everything.
     */
    async newModel() {
        // Save current state before clearing
        if (this._persistedFileContent) {
            IvyPersist.save(this);
        }

        // Create a fresh server session
        try {
            await this.api.createSession();
            var sessionEl = document.getElementById('session-id');
            if (sessionEl) {
                sessionEl.textContent = 'Session: ' + this.api.sessionId;
            }
            IvyPersist.setSessionIdInURL(this.api.sessionId);

            // Reconnect SSE
            if (this.api.sessionId) {
                this.api.connectEvents(this.handleEvent.bind(this));
            }
        } catch (e) {
            this.controls.setStatus('Failed to create session: ' + e.message, 'error');
            return;
        }

        // Clear graphs
        this.argGraph.cy.elements().remove();
        this.conceptGraph.cy.elements().remove();

        // Clear state
        this._persistedFileName = '';
        this._persistedFileContent = '';
        this._persistedConceptRelations = null;
        this.selectedArgNode = null;

        // Clear UI
        IvyPersist.setFileName('');
        var tbody = document.getElementById('state-checkbox-body');
        if (tbody) tbody.innerHTML = '';

        this.controls.setStatus('New model — load an .ivy file to begin', 'success');
    }

    /**
     * Populate the recent files list in the File dropdown.
     * Deduplicates by fileName+contentLength, keeping the most recent.
     * Shows truncated path context when file names collide.
     */
    populateRecentFiles() {
        var container = document.getElementById('file-recent-list');
        if (!container) return;
        container.innerHTML = '';

        var sessions = IvyPersist.listSessions();
        if (sessions.length === 0) {
            var empty = document.createElement('a');
            empty.href = '#';
            empty.textContent = '(no recent files)';
            empty.style.color = '#666';
            empty.style.pointerEvents = 'none';
            container.appendChild(empty);
            return;
        }

        // Deduplicate: keep only the most recent entry per (fileName, contentLength).
        var seen = {};
        var unique = [];
        for (var i = 0; i < sessions.length; i++) {
            var sess = sessions[i];
            if (!sess.fileName || sess.fileName === '(unnamed)') continue;
            // Load full state to get content length for dedup key
            var state = IvyPersist.loadSession(sess.id);
            var contentLen = (state && state.fileContent) ? state.fileContent.length : 0;
            var dedupKey = sess.fileName + '|' + contentLen;
            if (seen[dedupKey]) continue;
            seen[dedupKey] = true;
            unique.push(sess);
        }

        // Check for duplicate basenames to decide if path context is needed.
        var baseNameCount = {};
        for (var k = 0; k < unique.length; k++) {
            var bn = unique[k].fileName;
            baseNameCount[bn] = (baseNameCount[bn] || 0) + 1;
        }

        // Show up to 10 recent files
        var self = this;
        for (var j = 0; j < unique.length && j < 10; j++) {
            (function (s) {
                var displayName = s.fileName;
                // If basename appears more than once, show path context
                if (baseNameCount[s.fileName] > 1 && s.filePath) {
                    displayName = s.fileName + '  ' + IvyPersist.truncatePath(s.filePath, 15);
                }
                var link = document.createElement('a');
                link.href = '#';
                link.textContent = displayName;
                if (s.timestamp) {
                    var date = new Date(s.timestamp);
                    link.title = (s.filePath || s.fileName) + '\nLast used: ' + date.toLocaleString();
                }
                link.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.closeAllDropdowns(e);
                    self.loadRecentSession(s.id);
                });
                container.appendChild(link);
            })(unique[j]);
        }
    }

    /**
     * Load a recent session by its saved session ID.
     */
    async loadRecentSession(savedSessionId) {
        // Save current state first
        if (this._persistedFileContent) {
            IvyPersist.save(this);
        }

        var state = IvyPersist.loadSession(savedSessionId);
        if (!state || !state.fileContent) {
            this.controls.setStatus('Could not load session: no saved data', 'error');
            return;
        }

        // Create a fresh server session for this restore
        try {
            await this.api.createSession();
        } catch (e) {
            this.controls.setStatus('Failed to create session: ' + e.message, 'error');
            return;
        }

        // Reconnect SSE
        if (this.api.sessionId) {
            this.api.connectEvents(this.handleEvent.bind(this));
        }

        var restored = await IvyPersist.restore(this, state);
        if (restored) {
            IvyPersist.setSessionIdInURL(this.api.sessionId);
            IvyPersist.setFileName(state.fileName);
            var sessionEl = document.getElementById('session-id');
            if (sessionEl) {
                sessionEl.textContent = 'Session: ' + this.api.sessionId;
            }
            this.controls.setStatus('Loaded: ' + state.fileName, 'success');
        } else {
            this.controls.setStatus('Restore failed', 'error');
        }
    }

    /**
     * Run a verification check in the currently selected mode.
     */
    async runCheck() {
        var mode = document.getElementById('mode-select').value;
        this.controls.showLoading('Running ' + mode + ' check...');
        this.controls.setStatus('Running ' + mode + ' check...');
        try {
            var result = await this.api.runCheck(mode);
            this.showCheckResult(result);
        } catch (e) {
            this.controls.setStatus('Check failed: ' + e.message, 'error');
            console.error('Check error:', e);
        } finally {
            this.controls.hideLoading();
        }
    }

    /**
     * Display the result of a verification check.
     */
    showCheckResult(result) {
        if (!result) return;

        if (result.status === 'pass' || result.status === 'ok') {
            this.controls.setStatus('Check PASSED', 'success');
            this.controls.showInfo('Verification Result', 'PASSED: ' + (result.message || 'All properties hold.'));
        } else if (result.status === 'fail' || result.status === 'counterexample') {
            this.controls.setStatus('Check FAILED - counterexample found', 'error');
            this.controls.showInfo('Verification Result', 'FAILED: ' + (result.message || 'Counterexample found.'));
            // Update ARG if counterexample adds states
            if (result.arg) {
                this.argGraph.update(result.arg.elements, result.arg.positions);
            }
        } else {
            this.controls.setStatus('Check result: ' + result.status);
            this.controls.showInfo('Verification Result', result.message || JSON.stringify(result));
        }
    }

    /**
     * Undo the last concept domain change.
     */
    async doUndo() {
        this.controls.setStatus('Undoing...');
        try {
            await this.api.undo();
            await this.refreshConceptGraph();
            this.controls.setStatus('Undo complete', 'success');
        } catch (e) {
            // "nothing to undo" is not an error — just a no-op.
            if (e.message && e.message.indexOf('nothing to undo') >= 0) {
                this.controls.setStatus('Nothing to undo');
            } else {
                this.controls.setStatus('Undo failed: ' + e.message, 'error');
            }
        }
    }

    /**
     * Reset the concept domain to initial state.
     */
    async resetDomain() {
        this.controls.setStatus('Resetting domain...');
        try {
            await this.api.resetDomain();
            await this.refreshConceptGraph();
            this.controls.setStatus('Domain reset', 'success');
        } catch (e) {
            this.controls.setStatus('Reset failed: ' + e.message, 'error');
        }
    }

    /**
     * Switch to diagram concept domain.
     */
    async diagramDomain() {
        this.controls.setStatus('Switching to diagram domain...');
        try {
            await this.api.diagramDomain();
            await this.refreshConceptGraph();
            this.controls.setStatus('Diagram domain active', 'success');
        } catch (e) {
            this.controls.setStatus('Diagram domain failed: ' + e.message, 'error');
        }
    }

    /**
     * Refresh the concept graph from the server.
     */
    async refreshConceptGraph() {
        if (!this.selectedArgNode) return;
        try {
            var result = await this.api.getConceptGraph(this.selectedArgNode);
            if (result && result.elements) {
                this.conceptGraph.update(result.elements, result.positions);
            }
        } catch (e) {
            console.error('Concept graph refresh error:', e);
        }
    }

    // ================================================================
    // Server-Sent Events Handler
    // ================================================================

    /**
     * Handle an SSE event from the server.
     * Event types: arg_updated, concept_updated, status, check_result, error,
     *              toggles, proof_updated
     */
    handleEvent(event) {
        if (!event || !event.type) return;

        switch (event.type) {
            case 'arg_updated':
                if (event.data && event.data.elements) {
                    this.argGraph.update(event.data.elements, event.data.positions);
                }
                break;

            case 'concept_updated':
                if (event.data && event.data.elements) {
                    this.conceptGraph.update(event.data.elements, event.data.positions);
                }
                if (event.data && event.data.edge_names) {
                    this.controls.buildEdgeToggles(event.data.edge_names, this.onEdgeToggleChange.bind(this));
                }
                if (event.data && event.data.label_names) {
                    this.controls.buildLabelToggles(event.data.label_names, this.onLabelToggleChange.bind(this));
                }
                break;

            case 'status':
                if (event.data && event.data.message) {
                    this.controls.setStatus(event.data.message, event.data.level);
                }
                break;

            case 'check_result':
                this.controls.hideLoading();
                this.showCheckResult(event.data);
                break;

            case 'error':
                this.controls.hideLoading();
                this.controls.setStatus('Error: ' + (event.data.message || 'Unknown error'), 'error');
                if (event.data && event.data.message) {
                    this.controls.showInfo('Error', event.data.message);
                }
                break;

            case 'toggles':
                // Server is providing toggle configuration
                if (event.data && event.data.edge_names) {
                    this.controls.buildEdgeToggles(event.data.edge_names, this.onEdgeToggleChange.bind(this));
                }
                if (event.data && event.data.label_names) {
                    this.controls.buildLabelToggles(event.data.label_names, this.onLabelToggleChange.bind(this));
                }
                break;

            case 'proof_updated':
                // Future: update proof graph view
                break;

            case 'file_loaded':
                // File was loaded on the server; refresh graphs and state pane
                this.refreshAfterLoad(event.data);
                break;

            case 'check_started':
                this.controls.setStatus('Verification running...', 'info');
                break;

            case 'check_completed':
                var resultMsg = 'Check complete';
                if (event.data && event.data.result) {
                    resultMsg += ': ' + event.data.result;
                }
                this.controls.setStatus(resultMsg, 'success');
                break;

            case 'action_started':
                if (event.data && event.data.action) {
                    this.controls.setStatus('Running: ' + event.data.action + '...', 'info');
                }
                break;

            case 'action_completed':
                if (event.data && event.data.action) {
                    this.controls.setStatus('Done: ' + event.data.action, 'success');
                }
                break;

            default:
                // Silently ignore unknown event types
                break;
        }
    }

    // ================================================================
    // Dropdown Menu Infrastructure
    // ================================================================

    /**
     * Set up panel header dropdown menus (click to toggle).
     */
    setupDropdownMenus() {
        var self = this;
        var dropdowns = document.querySelectorAll('.dropdown > .panel-menu');
        for (var i = 0; i < dropdowns.length; i++) {
            (function (trigger) {
                trigger.addEventListener('click', function (e) {
                    e.stopPropagation();
                    var parent = trigger.parentElement;
                    var wasOpen = parent.classList.contains('open');
                    // Close all dropdowns first
                    var all = document.querySelectorAll('.dropdown.open');
                    for (var j = 0; j < all.length; j++) {
                        all[j].classList.remove('open');
                    }
                    if (!wasOpen) {
                        parent.classList.add('open');
                        // Populate recent files when File menu opens
                        var dropdownId = trigger.getAttribute('data-dropdown');
                        if (dropdownId === 'file-menu') {
                            self.populateRecentFiles();
                        }
                    }
                });
            })(dropdowns[i]);
        }
    }

    /**
     * Close all open dropdown menus.
     */
    closeAllDropdowns(e) {
        // Don't close if clicking inside a dropdown
        if (e && e.target && e.target.closest && e.target.closest('.dropdown-content')) {
            return;
        }
        var all = document.querySelectorAll('.dropdown.open');
        for (var j = 0; j < all.length; j++) {
            all[j].classList.remove('open');
        }
    }

    /**
     * Bind a menu item by ID to a callback, with dropdown auto-close.
     */
    bindMenuAction(id, callback) {
        var el = document.getElementById(id);
        if (!el) return;
        el.addEventListener('click', function (e) {
            e.preventDefault();
            e.stopPropagation();
            // Close all dropdowns
            var all = document.querySelectorAll('.dropdown.open');
            for (var j = 0; j < all.length; j++) {
                all[j].classList.remove('open');
            }
            callback();
        });
    }

    // ================================================================
    // Verification Operations (Invariant menu)
    // Matches Python ivy_ui_cti.py
    // ================================================================

    /**
     * Check inductiveness of current conjectures.
     * Matches Python ivy_ui_cti.py check_inductiveness().
     */
    async checkInduction() {
        this.controls.setStatus('Checking induction...');
        try {
            var result = await this.api.runCheck('induction');
            this.controls.setStatus('Induction check: ' + (result.result || 'done'), 'success');
        } catch (e) {
            this.controls.setStatus('Induction check failed: ' + e.message, 'error');
        }
    }

    /**
     * Run bounded model checking.
     * Matches Python ivy_ui_cti.py bounded_check().
     */
    async boundedCheck() {
        this.controls.setStatus('Running bounded check...');
        try {
            var result = await this.api.runCheck('bounded');
            this.controls.setStatus('Bounded check: ' + (result.result || 'done'), 'success');
        } catch (e) {
            this.controls.setStatus('Bounded check failed: ' + e.message, 'error');
        }
    }

    /**
     * Weaken the current invariant.
     * Matches Python ivy_ui_cti.py weaken().
     */
    async weakenInvariant() {
        this.controls.setStatus('Weakening invariant...');
        try {
            var result = await this.api.executeAction('weaken', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Invariant weakened', 'success');
        } catch (e) {
            this.controls.setStatus('Weaken failed: ' + e.message, 'error');
        }
    }

    /**
     * Save the current abstraction to a file.
     * Matches Python ivy_ui.py save_abstraction().
     */
    async saveAbstraction() {
        this.controls.setStatus('Saving abstraction...');
        try {
            // Build abstraction content from concept session state
            var result = await this.api.executeAction('save_abstraction', {});
            var content = (result && result.content) || '';
            if (!content) {
                // Fallback: serialize the concept graph elements as text
                if (this.conceptGraph && this.conceptGraph.cy) {
                    var elements = this.conceptGraph.cy.json().elements;
                    content = '# Ivy concept abstraction\n# Generated by Ivy web UI\n\n';
                    content += JSON.stringify(elements, null, 2) + '\n';
                } else {
                    content = '# (no abstraction data available)\n';
                }
            }

            var suggestedName = (this._persistedFileName || 'abstraction').replace(/\.ivy$/, '') + '_abstraction.ivy';

            if (window.showSaveFilePicker) {
                var handle = await window.showSaveFilePicker({
                    suggestedName: suggestedName,
                    types: [{
                        description: 'Ivy files',
                        accept: { 'text/plain': ['.ivy'] },
                    }],
                });
                var writable = await handle.createWritable();
                await writable.write(content);
                await writable.close();
                this.controls.setStatus('Abstraction saved: ' + handle.name, 'success');
            } else {
                // Fallback: browser download
                var blob = new Blob([content], { type: 'text/plain' });
                var url = URL.createObjectURL(blob);
                var a = document.createElement('a');
                a.href = url;
                a.download = suggestedName;
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
                this.controls.setStatus('Abstraction downloaded: ' + suggestedName, 'success');
            }
        } catch (e) {
            if (e.name === 'AbortError') {
                this.controls.setStatus('Save abstraction cancelled');
            } else {
                this.controls.setStatus('Save abstraction failed: ' + e.message, 'error');
            }
        }
    }

    // ================================================================
    // Conjecture Menu Operations
    // Matches Python ivy_graph_ui.py GraphWidget menus
    // ================================================================

    /**
     * Redo the last undone operation.
     */
    async doRedo() {
        this.controls.setStatus('Redo...');
        try {
            await this.api.executeAction('redo', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Redo complete', 'success');
        } catch (e) {
            this.controls.setStatus('Redo failed: ' + e.message, 'error');
        }
    }

    /**
     * Perform one step of PDR strengthening.
     * Matches Python ivy_graph_ui.py pdr_step().
     */
    async pdrStep() {
        this.controls.setStatus('PDR step...');
        try {
            var result = await this.api.executeAction('pdr_step', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('PDR step complete', 'success');
        } catch (e) {
            this.controls.setStatus('PDR step failed: ' + e.message, 'error');
        }
    }

    /**
     * Show concrete model.
     * Matches Python ivy_graph_ui.py concrete().
     */
    async concreteStep() {
        this.controls.setStatus('Computing concrete model...');
        try {
            var result = await this.api.executeAction('concrete', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Concrete model computed', 'success');
        } catch (e) {
            this.controls.setStatus('Concrete failed: ' + e.message, 'error');
        }
    }

    /**
     * Gather facts from the current state.
     * Matches Python ivy_graph_ui.py gather().
     */
    async gatherFacts() {
        this.controls.setStatus('Gathering facts...');
        try {
            var result = await this.api.executeAction('gather', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Facts gathered', 'success');
        } catch (e) {
            this.controls.setStatus('Gather failed: ' + e.message, 'error');
        }
    }

    /**
     * Compute reverse image.
     * Matches Python ivy_graph_ui.py reverse().
     */
    async reverseStep() {
        this.controls.setStatus('Computing reverse...');
        try {
            var result = await this.api.executeAction('reverse', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Reverse complete', 'success');
        } catch (e) {
            this.controls.setStatus('Reverse failed: ' + e.message, 'error');
        }
    }

    /**
     * Compute reachable states along a path.
     * Matches Python ivy_graph_ui.py path_reach().
     */
    async pathReach() {
        this.controls.setStatus('Computing path reachability...');
        try {
            var result = await this.api.executeAction('path_reach', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Path reach complete', 'success');
        } catch (e) {
            this.controls.setStatus('Path reach failed: ' + e.message, 'error');
        }
    }

    /**
     * Compute reachable states.
     * Matches Python ivy_graph_ui.py reach().
     */
    async reachStep() {
        this.controls.setStatus('Computing reachability...');
        try {
            var result = await this.api.executeAction('reach', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Reach complete', 'success');
        } catch (e) {
            this.controls.setStatus('Reach failed: ' + e.message, 'error');
        }
    }

    /**
     * Generate a conjecture from the current state.
     * Matches Python ivy_graph_ui.py conjecture().
     */
    async makeConjecture() {
        this.controls.setStatus('Generating conjecture...');
        try {
            var result = await this.api.executeAction('conjecture', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Conjecture generated', 'success');
        } catch (e) {
            this.controls.setStatus('Conjecture failed: ' + e.message, 'error');
        }
    }

    /**
     * Backtrack to a previous checkpoint.
     * Matches Python ivy_graph_ui.py backtrack().
     */
    async backtrack() {
        this.controls.setStatus('Backtracking...');
        try {
            await this.api.executeAction('backtrack', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Backtracked', 'success');
        } catch (e) {
            this.controls.setStatus('Backtrack failed: ' + e.message, 'error');
        }
    }

    /**
     * Recalculate the current concept graph.
     * Matches Python ivy_graph_ui.py recalculate().
     */
    async recalculateGraph() {
        this.controls.setStatus('Recalculating...');
        try {
            await this.api.executeAction('recalculate', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Recalculated', 'success');
        } catch (e) {
            this.controls.setStatus('Recalculate failed: ' + e.message, 'error');
        }
    }

    /**
     * Remember the current graph for later use.
     * Matches Python ivy_graph_ui.py remember().
     */
    async rememberGraph() {
        this.controls.setStatus('Remembering graph...');
        try {
            await this.api.executeAction('remember', {});
            this.controls.setStatus('Graph remembered', 'success');
        } catch (e) {
            this.controls.setStatus('Remember failed: ' + e.message, 'error');
        }
    }

    /**
     * Export the current conjecture.
     * Matches Python ivy_graph_ui.py export().
     */
    async exportConjecture() {
        this.controls.setStatus('Exporting conjecture...');
        try {
            var result = await this.api.executeAction('export', {});
            this.controls.setStatus('Conjecture exported', 'success');
        } catch (e) {
            this.controls.setStatus('Export failed: ' + e.message, 'error');
        }
    }

    /**
     * Add a relation from a user-entered string.
     * Matches Python ivy_graph_ui.py add_concept_from_string().
     */
    addRelationFromString() {
        var input = prompt('Enter a relation formula (e.g., r(X,Y)):');
        if (input) {
            this.api.executeAction('add_relation', { formula: input }).then(
                function () {
                    this.refreshConceptGraph();
                }.bind(this)
            ).catch(function (e) {
                this.controls.setStatus('Add relation failed: ' + e.message, 'error');
            }.bind(this));
        }
    }

    /**
     * Refresh graphs and state pane after a file_loaded SSE event.
     * The event.data contains {filename, sorts, relations, actions}.
     */
    async refreshAfterLoad(data) {
        try {
            // Refresh ARG
            var argData = await this.api.getARG();
            if (argData && argData.elements) {
                this.argGraph.update(argData.elements, argData.positions);
            }
            // Refresh concept graph and populate state checkbox pane
            var conceptData = await this.api.getConceptGraph();
            if (conceptData && conceptData.elements) {
                this.conceptGraph.update(conceptData.elements, conceptData.positions);
            }
            this.populateStateCheckboxes(conceptData);
            this.updateStateLabel(0);
        } catch (e) {
            console.error('refreshAfterLoad error:', e);
        }
    }
}

// ================================================================
// Initialize on DOM ready
// ================================================================
document.addEventListener('DOMContentLoaded', function () {
    window.ivyApp = new IvyApp();
    window.ivyApp.init();
});
