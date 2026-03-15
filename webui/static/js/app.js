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
        } catch (e) {
            this.controls.setStatus('Failed to create session: ' + e.message, 'error');
            console.error('Session creation failed:', e);
            // Continue anyway - graphs can still be created for when server comes up
        }

        // Create Cytoscape graph instances
        this.argGraph = new IvyGraph('arg-graph', ARG_STYLE);
        this.conceptGraph = new IvyGraph('concept-graph', CONCEPT_STYLE);

        // Wire up all event handlers
        this.setupEventHandlers();
        this.setupResizer();
        this.setupResizer2();
        this.setupKeyboardShortcuts();

        // Connect to SSE for real-time updates
        if (this.api.sessionId) {
            this.api.connectEvents(this.handleEvent.bind(this));
        }

        this.controls.setStatus('Ready');
    }

    /**
     * Set up all DOM and graph event handlers.
     */
    setupEventHandlers() {
        var self = this;

        // --- File Load ---
        var fileInput = document.getElementById('file-input');
        document.getElementById('btn-load').addEventListener('click', function () {
            fileInput.click();
        });
        fileInput.addEventListener('change', function () {
            if (fileInput.files.length > 0) {
                self.loadFile(fileInput.files[0]);
                fileInput.value = ''; // reset for re-selection of same file
            }
        });

        // --- Save ---
        document.getElementById('btn-save').addEventListener('click', function () {
            self.saveSession();
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

        // --- Click anywhere to dismiss context menu ---
        document.addEventListener('click', function () {
            self.controls.hideContextMenu();
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

        // ARG background click: clear info
        this.argGraph.onBackgroundClick(function () {
            self.controls.clearInfo();
            self.controls.hideContextMenu();
        });

        // --- Concept Graph Events ---

        // Concept node left-click: show info
        this.conceptGraph.onNodeClick(function (nodeData) {
            self.controls.showInfo(nodeData.short_info, nodeData.long_info);
        });

        // Concept node right-click: context menu (split, empty, remove, materialize)
        this.conceptGraph.onNodeRightClick(function (nodeData, pos) {
            self.onConceptNodeRightClick(nodeData, pos);
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
        // Send toggle state to server
        this.api.setToggles({
            edge: edgeName,
            display_class: displayClass,
            value: checked
        }).then(function() {
            // Optionally refresh concept graph
        }).catch(function(e) {
            console.error('Toggle error:', e);
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
            // Default ARG node actions
            var defaultActions = [
                { name: 'Execute Action...', id: 'execute' },
                { name: 'Mark', id: 'mark' },
                { name: 'Cover', id: 'cover' },
                { name: 'Safety Check', id: 'safety_check' },
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
            // Default concept node actions
            actions.push({
                name: 'Remove',
                id: 'remove',
                callback: function () {
                    self.removeConcept(nodeData.obj || nodeData.id);
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
            this.populateStateCheckboxes(conceptData);
            // Update state label
            this.updateStateLabel(0);
            this.controls.setStatus('Loaded: ' + file.name, 'success');
        } catch (e) {
            this.controls.setStatus('Load failed: ' + e.message, 'error');
            console.error('File load error:', e);
        } finally {
            this.controls.hideLoading();
        }
    }

    /**
     * Save the current session.
     */
    async saveSession() {
        this.controls.setStatus('Saving...');
        try {
            var blob = await this.api.saveSession();
            // Trigger download
            var url = URL.createObjectURL(blob);
            var a = document.createElement('a');
            a.href = url;
            a.download = 'ivy_session.ivy';
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            this.controls.setStatus('Saved', 'success');
        } catch (e) {
            this.controls.setStatus('Save failed: ' + e.message, 'error');
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
            this.controls.setStatus('Undo failed: ' + e.message, 'error');
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
