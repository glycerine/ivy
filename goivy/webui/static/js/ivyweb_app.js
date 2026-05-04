/**
 * IvyApp - Main application for the Ivy Interactive Verification web UI.
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
        // Edge visibility state, matching Python's edge_display_checkboxes.
        // Keys: edgeName, values: {all_to_all: bool, edge_unknown: bool, none_to_none: bool, transitive: bool}
        // Default: all false (edges hidden until checkbox is checked).
        this._edgeVisibility = {};
        // Node label visibility state, matching Python's node_label_display_checkboxes.
        // Keys: labelName, values: {node_necessarily: bool, node_maybe: bool, node_necessarily_not: bool}
        // Maps to checkbox columns: + → node_necessarily, ? → node_maybe, - → node_necessarily_not
        this._labelVisibility = {};
        // Concept data from last server response (for node/label sort info)
        this._lastConceptData = null;
        // File System Access API handle for in-place saves (Ctrl+S).
        this._fileHandle = null;
        // Content as last written to disk; used to detect unsaved changes.
        this._savedFileContent = null;
    }

    /**
     * Initialize the application: create session, build graphs, wire events.
     */
    async init() {
        this.controls.setStatus('Initializing...');

        // Check for a saved session BEFORE creating a new server session.
        // This prevents the URL session ID from incrementing on every reload.
        var savedState = IvyPersist.load();

        // Always need a server session for API calls.
        try {
            await this.api.createSession();
        } catch (e) {
            this.controls.setStatus('Failed to create session: ' + e.message, 'error');
            console.error('Session creation failed:', e);
        }

        // If restoring, keep the saved session's URL hash.
        // If fresh, set the new session ID in the URL.
        if (!savedState || !savedState.fileContent) {
            IvyPersist.setSessionIdInURL(this.api.sessionId);
        }

        // Show the persisted session ID (from URL hash), not the server session ID.
        // These can differ because the server ID increments on restart while
        // the persisted ID is stable across reloads.
        var displaySessionId = IvyPersist.getSessionIdFromURL() || this.api.sessionId;
        var sessionEl = document.getElementById('session-id');
        if (sessionEl) {
            sessionEl.textContent = 'Session: ' + displaySessionId;
        }

        // Create Cytoscape graph instances
        this.argGraph = new IvyGraph('arg-graph', ARG_STYLE);
        this.conceptGraph = new IvyGraph('concept-graph', CONCEPT_STYLE);

        // Health check: verify graphs initialized correctly.
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
        this.setupTabs();
        this.setupResizer();
        this.setupResizer2();
        this.setupResizer3();
        this.setupResizerH();
        this.setupTutorialUrlBar();
        this.setupKeyboardShortcuts();

        // Connect to SSE for real-time updates
        if (this.api.sessionId) {
            this.api.connectEvents(this.handleEvent.bind(this));
            this.api.onConnectionLost = function () {
                self.controls.setStatus('Server connection lost', 'error');
                self._showToast('Connection to server lost. Check that the server is running and reload the page.', 'error');
            };
        }

        // Initialize CodeMirror on the model editor textarea.
        // Must happen BEFORE restore so setEditorContent() can call cmEditor.setValue().
        var modelEditor = document.getElementById('model-editor');
        if (modelEditor) {
            this.cmEditor = CodeMirror.fromTextArea(modelEditor, {
                lineNumbers: true,
                keyMap: 'sublime',
                tabSize: 4,
                indentUnit: 4,
                lineWrapping: false,
                matchBrackets: true,
                extraKeys: {
                    'Ctrl-Z': 'undo',
                    'Ctrl-Y': 'redo',
                    'Ctrl-Shift-Z': 'redo',
                }
            });
            // Sync edits back to persisted content and update dirty marker.
            this.cmEditor.on('change', function () {
                self._persistedFileContent = self.cmEditor.getValue();
                self._updateEditorLabel();
            });
            // Keymap radio button switching.
            var radios = document.querySelectorAll('input[name="keymap"]');
            for (var i = 0; i < radios.length; i++) {
                radios[i].addEventListener('change', function () {
                    self.cmEditor.setOption('keyMap', this.value);
                });
            }
        }

        // Restore saved session if available (survives page reload).
        if (savedState && savedState.fileContent) {
            console.log('IvyPersist: restoring session', savedState.sessionId, savedState.fileName);
            var restored = await IvyPersist.restore(this, savedState);
            if (restored) {
                // Keep the URL hash from the saved session (don't overwrite)
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

    // --- Editor helpers (CodeMirror) ---

    setEditorContent(content) {
        this._persistedFileContent = content;
        this._savedFileContent = content;
        if (this.cmEditor) {
            this.cmEditor.setValue(content);
        }
        this._updateEditorLabel();
    }

    _updateEditorLabel() {
        var editorLabel = document.getElementById('model-editor-label');
        if (!editorLabel) return;
        var name = this._persistedFileName || '';
        var current = this.cmEditor ? this.cmEditor.getValue() : (this._persistedFileContent || '');
        var dirty = current !== (this._savedFileContent || '');
        if (!name) {
            editorLabel.textContent = 'Model: ' + (dirty ? '** ' : '') + '(unsaved file)';
        } else if (dirty) {
            editorLabel.textContent = 'Model: ** ' + name;
        } else {
            editorLabel.textContent = 'Model: ' + name + ' [saved]';
        }
    }

    _editorContent() {
        return this.cmEditor ? this.cmEditor.getValue() : (this._persistedFileContent || '');
    }

    _editorDirty() {
        return this._editorContent() !== (this._savedFileContent || '');
    }

    async _ensureFileHandleWritable() {
        if (!this._fileHandle) return false;
        if (!this._fileHandle.queryPermission || !this._fileHandle.requestPermission) {
            return true;
        }
        var opts = { mode: 'readwrite' };
        var perm = await this._fileHandle.queryPermission(opts);
        if (perm === 'granted') return true;
        perm = await this._fileHandle.requestPermission(opts);
        return perm === 'granted';
    }

    async _readFileHandleContent() {
        if (!this._fileHandle) return null;
        var file = await this._fileHandle.getFile();
        return await file.text();
    }

    async _confirmNoExternalChangeBeforeSave(content) {
        if (!this._fileHandle) return 'ok';
        var diskContent = await this._readFileHandleContent();
        var lastSaved = this._savedFileContent || '';
        if (diskContent === lastSaved || diskContent === content) {
            return 'ok';
        }
        var choice = await this.showExternalChangeDialog();
        if (choice === 'overwrite') {
            return 'overwrite';
        }
        if (choice === 'reload') {
            this.setEditorContent(diskContent);
            this._persistedFileContent = diskContent;
            this._savedFileContent = diskContent;
            IvyPersist.save(this);
            this.controls.setStatus('Reverted to on-disk version: ' + (this._persistedFileName || 'model'), 'success');
        } else if (choice === 'merge') {
            var merged = this._mergeDiskVersionIntoEditBuffer(lastSaved, content, diskContent);
            this._persistedFileContent = merged;
            this._savedFileContent = diskContent;
            if (this.cmEditor) {
                this.cmEditor.setValue(merged);
            }
            this._updateEditorLabel();
            IvyPersist.save(this);
            this.controls.setStatus('Merged disk changes into editor buffer; resolve conflict markers before saving', 'warning');
        } else {
            this.controls.setStatus('Save cancelled: file changed on disk', 'warning');
        }
        return 'skip';
    }

    _mergeDiskVersionIntoEditBuffer(baseContent, editorContent, diskContent) {
        if (editorContent === baseContent) return diskContent;
        if (diskContent === baseContent) return editorContent;
        return [
            '<<<<<<< EDIT BUFFER',
            editorContent.replace(/\s*$/, ''),
            '||||||| LAST SAVED',
            baseContent.replace(/\s*$/, ''),
            '=======',
            diskContent.replace(/\s*$/, ''),
            '>>>>>>> ON DISK',
            ''
        ].join('\n');
    }

    showExternalChangeDialog() {
        var self = this;
        return new Promise(function (resolve) {
            var overlay = document.getElementById('external-change-dialog-overlay');
            var msg = document.getElementById('external-change-dialog-message');
            var doNothing = document.getElementById('external-change-do-nothing');
            var reload = document.getElementById('external-change-reload');
            var merge = document.getElementById('external-change-merge');
            var overwrite = document.getElementById('external-change-overwrite');
            if (!overlay || !msg || !doNothing || !reload || !merge || !overwrite) {
                resolve('do-nothing');
                return;
            }
            msg.textContent = 'The file "' + (self._persistedFileName || 'model') + '" has changed outside IvyWeb. Choose how to handle the current editor buffer.';
            overlay.style.display = 'flex';
            doNothing.focus();

            var done = function (choice) {
                overlay.style.display = 'none';
                doNothing.removeEventListener('click', onDoNothing);
                reload.removeEventListener('click', onReload);
                merge.removeEventListener('click', onMerge);
                overwrite.removeEventListener('click', onOverwrite);
                document.removeEventListener('keydown', onKeyDown);
                resolve(choice);
            };
            var onDoNothing = function () { done('do-nothing'); };
            var onReload = function () { done('reload'); };
            var onMerge = function () { done('merge'); };
            var onOverwrite = function () { done('overwrite'); };
            var onKeyDown = function (e) {
                if (e.key === 'Escape' || e.key === 'Enter') {
                    e.preventDefault();
                    done('do-nothing');
                }
            };
            doNothing.addEventListener('click', onDoNothing);
            reload.addEventListener('click', onReload);
            merge.addEventListener('click', onMerge);
            overwrite.addEventListener('click', onOverwrite);
            document.addEventListener('keydown', onKeyDown);
        });
    }

    showDirtyCloseDialog() {
        var self = this;
        return new Promise(function (resolve) {
            var overlay = document.getElementById('dirty-close-dialog-overlay');
            var msg = document.getElementById('dirty-close-dialog-message');
            var save = document.getElementById('dirty-close-save');
            var discard = document.getElementById('dirty-close-discard');
            if (!overlay || !msg || !save || !discard) {
                resolve('save');
                return;
            }
            msg.textContent = 'File "' + (self._persistedFileName || 'model') + '" has changed. Save before closing?';
            overlay.style.display = 'flex';
            save.focus();

            var done = function (choice) {
                overlay.style.display = 'none';
                save.removeEventListener('click', onSave);
                discard.removeEventListener('click', onDiscard);
                document.removeEventListener('keydown', onKeyDown);
                resolve(choice);
            };
            var onSave = function () { done('save'); };
            var onDiscard = function () { done('discard'); };
            var onKeyDown = function (e) {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    done('save');
                } else if (e.key === 'Escape') {
                    e.preventDefault();
                    done('discard');
                }
            };
            save.addEventListener('click', onSave);
            discard.addEventListener('click', onDiscard);
            document.addEventListener('keydown', onKeyDown);
        });
    }

    scrollEditorToLine(lineno) {
        if (this.cmEditor) {
            var line = lineno - 1;
            this.cmEditor.setCursor(line, 0);
            this.cmEditor.setSelection(
                {line: line, ch: 0},
                {line: line, ch: this.cmEditor.getLine(line).length}
            );
            this.cmEditor.scrollIntoView({line: line, ch: 0}, 50);
            this.cmEditor.focus();
        }
    }

    /**
     * Set up all DOM and graph event handlers.
     */
    setupEventHandlers() {
        var self = this;

        // --- File Menu ---
        var fileInput = document.getElementById('file-input');

        // File > Load...
        // Use showOpenFilePicker when available so we get a writable FileSystemFileHandle,
        // enabling Ctrl+S to save directly without re-prompting. Fall back to <input> otherwise.
        document.getElementById('file-load').addEventListener('click', function (e) {
            e.preventDefault();
            self.flashAndClose(this, async function () {
                if (window.showOpenFilePicker) {
                    try {
                        var handles = await window.showOpenFilePicker({
                            types: [{ description: 'Ivy files', accept: { 'text/plain': ['.ivy'] } }],
                            multiple: false,
                        });
                        var handle = handles[0];
                        var file = await handle.getFile();
                        self._fileHandle = handle;
                        await self.loadFile(file);
                    } catch (ex) {
                        if (ex.name !== 'AbortError') {
                            self.controls.setStatus('Load failed: ' + ex.message, 'error');
                        }
                    }
                } else {
                    fileInput.click();
                }
            });
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
            self.flashAndClose(this, function () { self.saveAs(); });
        });

        // File > Download current model (browser download)
        document.getElementById('file-download').addEventListener('click', function (e) {
            e.preventDefault();
            self.flashAndClose(this, function () { self.downloadModel(); });
        });

        // File > New Model
        document.getElementById('file-new').addEventListener('click', function (e) {
            e.preventDefault();
            self.flashAndClose(this, function () { self.newModel(); });
        });

        var closeCurrent = document.getElementById('file-close-current');
        if (closeCurrent) {
            closeCurrent.addEventListener('click', function (e) {
                e.preventDefault();
                self.closeCurrentFile();
            });
        }

        // File > Save Invariant
        document.getElementById('file-save-invariant').addEventListener('click', function (e) {
            e.preventDefault();
            self.flashAndClose(this, function () { self.saveInvariant(); });
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

        // --- Toggle Tutorial ---
        document.getElementById('btn-toggle-tutorial').addEventListener('click', function () {
            self.toggleTutorial();
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
            if (!e.target.closest('.dropdown')) {
                self.closeAllDropdowns();
            }
        });

        // --- Escape key closes open dropdowns; Ctrl+S saves ---
        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape') {
                self.closeAllDropdowns();
            }
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                self.save();
            }
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
                    node.unselect(); // clear Cytoscape's built-in :selected state
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

        // Concept edge left-click: toggle selection (matches Python Tk behavior).
        this.conceptGraph.onEdgeClick(function (edgeData, evt) {
            var edge = evt.target;
            var name = edgeData.obj || edgeData.label || edgeData.id;
            if (edge.hasClass('selected_edge')) {
                edge.removeClass('selected_edge');
                edge.unselect();
                // Remove inline styles so class-based styles work on next select
                edge.removeStyle('line-color target-arrow-color source-arrow-color width');
                self.controls.setStatus('Deselected: ' + name);
                self.controls.clearInfo();
            } else {
                edge.addClass('selected_edge');
                self.controls.setStatus('Selected: ' + name);
                self.controls.showInfo(edgeData.short_info, edgeData.long_info);
            }
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
     * Set up resizable dividers between ARG and concept panels.
     * Uses event delegation on the sheet area so it works for ALL tabs,
     * including dynamically created ones.
     */
    setupResizer() {
        var sheetArea = document.getElementById('sheet-area');
        if (!sheetArea) return;
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;
        var activeDivider = null;
        var activePanel = null;
        var activeContainer = null;

        // Delegation: any .divider inside a .sheet-main starts a drag
        sheetArea.addEventListener('mousedown', function (e) {
            var div = e.target;
            if (!div.classList.contains('divider')) return;
            var container = div.parentElement; // .sheet-main
            var panel = div.previousElementSibling; // ARG panel (left of divider)
            if (!container || !panel) return;

            isDragging = true;
            activeDivider = div;
            activePanel = panel;
            activeContainer = container;
            startX = e.clientX;
            startWidth = panel.offsetWidth;
            div.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            var canvases = document.querySelectorAll('.graph-container');
            for (var i = 0; i < canvases.length; i++) canvases[i].style.pointerEvents = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = e.clientX - startX;
            var newWidth = startWidth + dx;
            var containerWidth = activeContainer ? activeContainer.offsetWidth : 800;
            newWidth = Math.max(150, Math.min(newWidth, containerWidth - 200));
            activePanel.style.flex = '0 0 ' + newWidth + 'px';
            if (self.argGraph) self.argGraph.resize();
            if (self.conceptGraph) self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                if (activeDivider) activeDivider.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                var canvases = document.querySelectorAll('.graph-container');
                for (var i = 0; i < canvases.length; i++) canvases[i].style.pointerEvents = '';
                if (self.argGraph) self.argGraph.resize();
                if (self.conceptGraph) self.conceptGraph.resize();
                activeDivider = null;
                activePanel = null;
                activeContainer = null;
            }
        });
    }

    /**
     * Set up the resizable second divider between concept and state panels.
     */
    /**
     * Resizer for divider2: between sheet-left (ARG+Concept+Details) and state-panel.
     * Dragging left makes state panel wider; dragging right makes left wider.
     */
    setupResizer2() {
        var divider2 = document.getElementById('divider2');
        if (!divider2) return;
        var rightSection = document.getElementById('state-panel');
        var topRow = divider2.parentElement;
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;

        divider2.addEventListener('mousedown', function (e) {
            isDragging = true;
            startX = e.clientX;
            startWidth = rightSection.offsetWidth;
            divider2.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = startX - e.clientX; // drag left = right-section wider
            var newWidth = startWidth + dx;
            var maxW = topRow ? topRow.offsetWidth - 300 : 800;
            newWidth = Math.max(200, Math.min(newWidth, maxW));
            rightSection.style.flex = '0 0 ' + newWidth + 'px';
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
     * Set up the resizable divider between State pane and Editor/Tutorial pane.
     */
    /**
     * Resizer for divider3: between sheet-area and editor-panel.
     * Dragging left makes editor wider; dragging right makes sheet area wider.
     */
    setupResizer3() {
        var divider3 = document.getElementById('divider3');
        if (!divider3) return;
        var statePanel = document.getElementById('sheet-area');
        var rightSection = document.getElementById('top-row');
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;

        divider3.addEventListener('mousedown', function (e) {
            isDragging = true;
            startX = e.clientX;
            startWidth = statePanel.offsetWidth;
            divider3.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = e.clientX - startX; // drag right = state wider
            var newWidth = startWidth + dx;
            var maxW = rightSection ? rightSection.offsetWidth - 100 : 400;
            newWidth = Math.max(100, Math.min(newWidth, maxW));
            statePanel.style.flex = '0 0 ' + newWidth + 'px';
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                divider3.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
            }
        });
    }

    /**
     * Populate the state checkbox table (right pane) with edge/relation names.
     * Each row has checkboxes for: + (all_to_all), ? (unknown), - (none_to_none), T (transitive)
     * and the relation name.
     */
    /**
     * Set up the tutorial URL bar: Go button and Enter key navigate the iframe.
     */
    /**
     * Resizer for divider-h: horizontal divider between top row and tutorial BiB.
     * Dragging up makes tutorial taller; dragging down makes top row taller.
     */
    /**
     * Toggle the tutorial BiB panel visibility.
     */
    /**
     * Set up tab bar click handlers.
     * Matches Python tix.NoteBook: clicking a tab shows that sheet.
     */
    setupTabs() {
        var self = this;
        this._sheetCounter = 1;
        var tabBar = document.getElementById('tab-bar');
        if (!tabBar) return;
        tabBar.addEventListener('click', function (e) {
            // Close button clicked?
            if (e.target.classList.contains('tab-close')) {
                var tab = e.target.parentElement;
                var sheetId = tab.getAttribute('data-sheet');
                self.removeSheet(sheetId);
                return;
            }
            var tab = e.target.closest('.sheet-tab');
            if (!tab) return;
            self.switchSheet(tab.getAttribute('data-sheet'));
        });
    }

    /**
     * Switch to a sheet by ID.
     */
    switchSheet(sheetId) {
        // Deactivate all tabs and sheets
        var tabs = document.querySelectorAll('.sheet-tab');
        var sheets = document.querySelectorAll('.sheet-content');
        for (var i = 0; i < tabs.length; i++) tabs[i].classList.remove('active');
        for (var i = 0; i < sheets.length; i++) sheets[i].classList.remove('active');
        // Activate the target
        var tab = document.querySelector('.sheet-tab[data-sheet="' + sheetId + '"]');
        var sheet = document.getElementById(sheetId);
        if (tab) tab.classList.add('active');
        if (sheet) sheet.classList.add('active');
        // Resize graphs in the newly visible sheet
        if (this.argGraph) this.argGraph.resize();
        if (this.conceptGraph) this.conceptGraph.resize();
    }

    /**
     * Add a new sheet tab (matches Python ui_parent.add(art)).
     * Called on "Step into" / "Decompose" actions.
     * @param {string} [label] - Tab label (default: "Sheet N")
     * @returns {string} The new sheet ID
     */
    addSheet(label) {
        this._sheetCounter++;
        var sheetId = 'sheet-' + this._sheetCounter;
        label = label || ('Sheet ' + this._sheetCounter);

        // Create tab button with close X (Sheet 1 never has X)
        var tabBar = document.getElementById('tab-bar');
        var tabBtn = document.createElement('button');
        tabBtn.className = 'sheet-tab';
        tabBtn.setAttribute('data-sheet', sheetId);
        var labelSpan = document.createElement('span');
        labelSpan.textContent = label;
        tabBtn.appendChild(labelSpan);
        var closeBtn = document.createElement('span');
        closeBtn.className = 'tab-close';
        closeBtn.textContent = '\u00D7'; // ×
        closeBtn.title = 'Close tab';
        tabBtn.appendChild(closeBtn);
        tabBar.appendChild(tabBtn);

        // Create sheet content (clone structure from sheet-1)
        var template = document.getElementById('sheet-1');
        var newSheet = template.cloneNode(true);
        newSheet.id = sheetId;
        newSheet.classList.remove('active');
        // Clear graph containers (they'll be initialized fresh)
        var graphs = newSheet.querySelectorAll('.graph-container');
        for (var i = 0; i < graphs.length; i++) {
            graphs[i].innerHTML = '';
            graphs[i].id = graphs[i].id + '-' + this._sheetCounter;
        }
        // Clear info panel
        var info = newSheet.querySelector('#info-content');
        if (info) {
            info.id = 'info-content-' + this._sheetCounter;
            info.textContent = 'Select a node or edge to see details';
        }
        var infoHeader = newSheet.querySelector('#info-header');
        if (infoHeader) infoHeader.id = 'info-header-' + this._sheetCounter;
        // Insert before the tutorial container
        var sheetArea = document.getElementById('sheet-area');
        sheetArea.appendChild(newSheet);

        // Switch to the new sheet
        this.switchSheet(sheetId);
        this.controls.setStatus('Opened: ' + label);
        return sheetId;
    }

    /**
     * Remove a sheet tab and its content.
     * If the removed sheet was active, switch to Sheet 1.
     * Sheet 1 cannot be removed.
     */
    removeSheet(sheetId) {
        if (sheetId === 'sheet-1') return; // never remove Sheet 1

        var tab = document.querySelector('.sheet-tab[data-sheet="' + sheetId + '"]');
        var sheet = document.getElementById(sheetId);
        var wasActive = tab && tab.classList.contains('active');

        if (tab) tab.remove();
        if (sheet) sheet.remove();

        // If the closed tab was active, switch to Sheet 1
        if (wasActive) {
            this.switchSheet('sheet-1');
        }
    }

    toggleTutorial(flash) {
        var tutorial = document.getElementById('tutorial-container');
        var dividerH = document.getElementById('divider-h');
        var btn = document.getElementById('btn-toggle-tutorial');
        if (!tutorial || !btn) return;

        if (tutorial.style.display === 'none') {
            // Show
            tutorial.style.display = '';
            if (dividerH) dividerH.style.display = '';
            btn.textContent = 'Hide Tutorial';
        } else {
            // Hide
            tutorial.style.display = 'none';
            if (dividerH) dividerH.style.display = 'none';
            btn.textContent = 'Show Tutorial';
            // Flash the button to alert the user where to find it again
            if (flash) {
                btn.classList.add('btn-flash');
                setTimeout(function () { btn.classList.remove('btn-flash'); }, 1200);
            }
        }
        // Resize graphs to fill the reclaimed/reduced space
        if (this.argGraph) this.argGraph.resize();
        if (this.conceptGraph) this.conceptGraph.resize();
    }

    setupResizerH() {
        var dividerH = document.getElementById('divider-h');
        if (!dividerH) return;
        var tutorial = document.getElementById('tutorial-container');
        var outerContainer = document.getElementById('outer-container');
        var iframe = document.getElementById('tutorial-iframe');
        var self = this;
        var isDragging = false;
        var startY = 0;
        var startHeight = 0;

        dividerH.addEventListener('mousedown', function (e) {
            isDragging = true;
            startY = e.clientY;
            startHeight = tutorial.offsetHeight;
            dividerH.classList.add('active');
            document.body.style.cursor = 'row-resize';
            document.body.style.userSelect = 'none';
            // Block iframe from stealing mouse events during drag
            if (iframe) iframe.style.pointerEvents = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dy = startY - e.clientY;
            var newHeight = startHeight + dy;
            var maxH = outerContainer ? outerContainer.offsetHeight - 100 : 600;
            newHeight = Math.max(80, Math.min(newHeight, maxH));
            tutorial.style.flex = '0 0 ' + newHeight + 'px';
            self.argGraph.resize();
            self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                dividerH.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                if (iframe) iframe.style.pointerEvents = '';
                self.argGraph.resize();
                self.conceptGraph.resize();
            }
        });
    }

    setupTutorialUrlBar() {
        var urlInput = document.getElementById('tutorial-url');
        var iframe = document.getElementById('tutorial-iframe');
        var backBtn = document.getElementById('tutorial-back');
        var fwdBtn = document.getElementById('tutorial-fwd');
        var reloadBtn = document.getElementById('tutorial-reload');
        var closeBtn = document.getElementById('tutorial-close');
        if (!urlInput || !iframe) return;

        // Close button: hide tutorial and flash the "Show Tutorial" button
        var self = this;
        if (closeBtn) {
            closeBtn.addEventListener('click', function () {
                self.toggleTutorial(true);
            });
        }

        // Track navigation history.
        var history = [urlInput.value.trim()];
        var historyIdx = 0;

        function navigateTo(url, forceReload) {
            if (!url) return;
            if (!url.match(/^https?:\/\//) && !url.startsWith('/')) {
                url = 'https://' + url;
            }
            if (forceReload) {
                // Force reload: about:blank trick bypasses browser cache
                iframe.src = 'about:blank';
                setTimeout(function () { iframe.src = url; }, 0);
            } else {
                // Normal navigation: browser cache can serve the page offline
                iframe.src = url;
            }
            if (historyIdx < history.length - 1) {
                history = history.slice(0, historyIdx + 1);
            }
            history.push(url);
            historyIdx = history.length - 1;
            urlInput.value = url;
            updateNavButtons();
        }

        function updateNavButtons() {
            if (backBtn) backBtn.disabled = (historyIdx <= 0);
            if (fwdBtn) fwdBtn.disabled = (historyIdx >= history.length - 1);
        }

        // Enter key navigates
        urlInput.addEventListener('keydown', function (e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                navigateTo(urlInput.value.trim());
            }
        });

        // Back button
        if (backBtn) {
            backBtn.addEventListener('click', function () {
                if (historyIdx > 0) {
                    historyIdx--;
                    var url = history[historyIdx];
                    urlInput.value = url;
                    iframe.src = url;
                    updateNavButtons();
                }
            });
        }

        // Forward button
        if (fwdBtn) {
            fwdBtn.addEventListener('click', function () {
                if (historyIdx < history.length - 1) {
                    historyIdx++;
                    var url = history[historyIdx];
                    urlInput.value = url;
                    iframe.src = url;
                    updateNavButtons();
                }
            });
        }

        // Reload button — force reload (bypasses cache, for when server is back)
        if (reloadBtn) {
            reloadBtn.addEventListener('click', function () {
                var url = history[historyIdx];
                if (url) {
                    navigateTo(url, true);
                }
            });
        }

        // Update URL bar when iframe navigates, and cache successful pages.
        iframe.addEventListener('load', function () {
            try {
                var newUrl = iframe.contentWindow.location.href;
                // Convert full URL to path for cleaner display
                if (newUrl && newUrl !== 'about:blank') {
                    try {
                        var u = new URL(newUrl);
                        newUrl = u.pathname + u.search + u.hash;
                    } catch (e) {}
                    if (history[historyIdx] !== newUrl) {
                        if (historyIdx < history.length - 1) {
                            history = history.slice(0, historyIdx + 1);
                        }
                        history.push(newUrl);
                        historyIdx = history.length - 1;
                    }
                    urlInput.value = newUrl;
                }
            } catch (e) {
                // Cross-origin or error page — don't cache
            }
            updateNavButtons();
        });

        updateNavButtons();
    }

    populateStateCheckboxes(conceptData) {
        // Store concept data for node label rendering
        this._lastConceptData = conceptData;
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
        // Track visibility state client-side (matches Python edge/node_label display_checkboxes)
        // Python maps checkbox columns to keys:
        //   For edges:  + → all_to_all, ? → edge_unknown, - → none_to_none, T → transitive
        //   For labels: + → node_necessarily, ? → node_maybe, - → node_necessarily_not
        if (!this._edgeVisibility[edgeName]) {
            this._edgeVisibility[edgeName] = {
                all_to_all: false, edge_unknown: false, none_to_none: false, transitive: false
            };
        }
        this._edgeVisibility[edgeName][displayClass] = checked;

        // Also track as a node label (Python does both: set_checkbox sets BOTH dicts).
        // Store under the bare name (e.g. "semaphore") since _applyNodeLabels
        // looks up by bare node_label names, not the parameterized relation
        // names like "semaphore(X)" that the checkboxes display.
        var labelKeyMap = { 'all_to_all': 'node_necessarily', 'edge_unknown': 'node_maybe', 'none_to_none': 'node_necessarily_not' };
        var labelKey = labelKeyMap[displayClass];
        if (labelKey) {
            var bareName = edgeName.split('(')[0];
            if (!this._labelVisibility[bareName]) {
                this._labelVisibility[bareName] = { node_necessarily: false, node_maybe: false, node_necessarily_not: false };
            }
            this._labelVisibility[bareName][labelKey] = checked;
        }

        // Apply visibility to edges and node labels
        this._applyEdgeVisibility();
        this._applyNodeLabels();

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
            // Try multiple keys to match: the edge obj, label, and
            // formatted versions like "link(X,Y)" that checkboxes use.
            var obj = edge.data('obj') || '';
            var label = edge.data('label') || '';
            var vis = self._findEdgeVisibility(obj, label);
            if (!vis) {
                // No checkbox state → hide by default (matching Python)
                edge.style('display', 'none');
                return;
            }
            // Check if ANY of the edge's classes has its checkbox checked.
            // Cytoscape classes() returns an array.
            var classList = edge.classes();
            var show = false;
            for (var i = 0; i < classList.length; i++) {
                if (vis[classList[i]]) {
                    show = true;
                    break;
                }
            }
            edge.style('display', show ? 'element' : 'none');
        });
    }

    /**
     * Find edge visibility entry. Checkboxes use display names like "link(X,Y)"
     * while edge data uses bare names like "link". Try both.
     */
    _findEdgeVisibility(obj, label) {
        var ev = this._edgeVisibility;
        // Direct match on obj or label
        if (ev[obj]) return ev[obj];
        if (ev[label]) return ev[label];
        // Checkbox names may include params: "link(X,Y)" — try matching
        // by prefix before the "("
        for (var key in ev) {
            if (ev.hasOwnProperty(key)) {
                var base = key.split('(')[0];
                if (base === obj || base === label) {
                    return ev[key];
                }
            }
        }
        return null;
    }

    /**
     * Apply node label text based on checkbox state.
     * Matches Python cy_render.py render_concept_graph lines 127-146:
     *   For each sort node, check each node_label. If the label's checkbox
     *   is checked, add the label text (with prefix) to the node's display.
     *   Prefix: + → plain, ? → "?suffix", - → "¬prefix"
     */
    /**
     * Apply node label text based on checkbox state and abstract value.
     * Matches Python cy_render.py render_concept_graph lines 127-146 EXACTLY:
     *
     * For each sort node and each node_label:
     *   1. Check if label's sort matches this node's sort (label_sorts map)
     *   2. Determine k from abstract_value:
     *      - if a['node_label|node_necessarily|node|label'] → k = 'node_necessarily'
     *      - elif a['node_label|node_necessarily_not|node|label'] → k = 'node_necessarily_not'
     *      - else → k = 'node_maybe'
     *   3. Check if checkbox[label][k] is checked → if not, skip
     *   4. Display with prefix: '' for +, '?' for ?, '¬' for -
     */
    _applyNodeLabels() {
        if (!this.conceptGraph || !this.conceptGraph.cy) return;
        if (!this._lastConceptData) return;

        var labelPrefixes = {
            'node_necessarily': '',
            'node_maybe': '?',
            'node_necessarily_not': '\u00AC'
        };

        var nodeLabels = this._lastConceptData.node_labels || [];
        var labelSorts = this._lastConceptData.label_sorts || {};
        var abstractValue = this._lastConceptData.abstract_value || {};

        var self = this;
        this.conceptGraph.cy.nodes().forEach(function (node) {
            var sortName = node.data('obj') || '';
            if (!sortName) return;

            var labelParts = [sortName];

            for (var i = 0; i < nodeLabels.length; i++) {
                var labelName = nodeLabels[i];
                var baseLabelName = labelName.split('(')[0];

                // Step 1: Check if this label belongs to this sort node.
                // Python: only adds label to nodes whose sort matches the label's sort.
                var labelSort = labelSorts[labelName] || labelSorts[baseLabelName];
                if (labelSort && labelSort !== sortName) continue;

                // Step 2: Determine k from abstract_value (Z3 result).
                // Python cy_render.py lines 132-137.
                var k;
                var necKey = 'node_label|node_necessarily|' + sortName + '|' + baseLabelName;
                var necNotKey = 'node_label|node_necessarily_not|' + sortName + '|' + baseLabelName;
                if (abstractValue[necKey]) {
                    k = 'node_necessarily';
                } else if (abstractValue[necNotKey]) {
                    k = 'node_necessarily_not';
                } else {
                    k = 'node_maybe';
                }

                // Step 3: Check if the checkbox for this k is checked.
                // Python: widget.node_label_display_checkboxes[label_name][k].value
                var vis = self._labelVisibility[labelName] || self._labelVisibility[baseLabelName];
                if (!vis || !vis[k]) continue;

                // Step 4: Display with prefix.
                var prefix = labelPrefixes[k];
                if (prefix === '?') {
                    labelParts.push(prefix + baseLabelName);
                } else {
                    labelParts.push(prefix + baseLabelName);
                }
            }

            var newLabel = labelParts.join('\n');
            if (node.data('label') !== newLabel) {
                node.data('label', newLabel);
                var lines = labelParts.length;
                var h = Math.max(50, 30 + lines * 20);
                node.data('height', h);
            }
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
        var graphContainer = document.getElementById('arg-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
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
            if (actionName === 'decompose' && result && result.decomposed) {
                // Decompose: open a new tab with the sub-ARG
                var label = 'Step: ' + (edgeData.label || actionName);
                var sheetId = this.addSheet(label);
                // Populate the new sheet's ARG with the decomposed sub-graph
                if (result.sub_arg && result.sub_arg.elements) {
                    var sheet = document.getElementById(sheetId);
                    if (sheet) {
                        var argContainer = sheet.querySelector('.graph-container');
                        if (argContainer) {
                            var subGraph = new IvyGraph(argContainer.id, ARG_STYLE);
                            subGraph.update(result.sub_arg.elements);
                        }
                    }
                }
            }
            if (result && result.arg) {
                this.argGraph.update(result.arg.elements, result.arg.positions);
            }
            if (result && result.source && actionName === 'view_source') {
                // Show source in the model editor and scroll to the action line.
                // Matches Python ivy_ui.py view_source_edge → browse(filename, lineno).
                this.setEditorContent(result.source);
                if (result.lineno) {
                    this.scrollEditorToLine(result.lineno);
                }
                this.controls.showInfo(
                    'Source: ' + (result.file || '') + (result.lineno ? ' line ' + result.lineno : ''),
                    ''
                );
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
            if (self._fileHandle) {
                await IvyPersist.saveFileHandle(self);
            }

            // Populate the model editor with the file content (also marks clean via setEditorContent)
            this.setEditorContent(fileContent);
            var editorLabel = document.getElementById('model-editor-label');
            if (editorLabel) {
                editorLabel.textContent = 'Model: ' + file.name;
            }

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
            var suggestedName = (this._persistedFileName || 'model').replace(/\.ivy$/, '') + '_invariant.ivy';

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
    async save() {
        var content = this._editorContent();
        var dirty = this._editorDirty();
        if (this._fileHandle) {
            try {
                var writableAllowed = await this._ensureFileHandleWritable();
                if (!writableAllowed) {
                    this.controls.setStatus('Save permission denied', 'error');
                    return false;
                }
                var saveDecision = await this._confirmNoExternalChangeBeforeSave(content);
                if (saveDecision === 'skip') {
                    return false;
                }
                if (!dirty && saveDecision !== 'overwrite') {
                    this._updateEditorLabel();
                    return true;
                }
                var writable = await this._fileHandle.createWritable();
                await writable.write(content);
                await writable.close();
                this._persistedFileContent = content;
                this._savedFileContent = content;
                this._updateEditorLabel();
                this.controls.setStatus('Saved: ' + this._persistedFileName, 'success');
                return true;
            } catch (e) {
                this.controls.setStatus('Save failed: ' + e.message, 'error');
                return false;
            }
        } else {
            if (!dirty) {
                this._updateEditorLabel();
                return true;
            }
            return await this.saveAs();
        }
    }

    async saveAs() {
        var content = this._editorContent();
        if (!content) {
            this.controls.setStatus('No model loaded to save', 'error');
            return false;
        }
        try {
            if (!window.showSaveFilePicker) {
                // Fallback for browsers without File System Access API
                this.controls.setStatus('Save as... not supported in this browser — use Download instead', 'error');
                return false;
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
            this._persistedFilePath = handle.name;
            this._persistedFileContent = content;
            this._savedFileContent = content;
            await IvyPersist.saveFileHandle(this);
            IvyPersist.setFileName(handle.name);
            this._updateEditorLabel();
            this.controls.setStatus('Saved: ' + handle.name, 'success');
            return true;
        } catch (e) {
            if (e.name === 'AbortError') {
                // User cancelled the dialog
                this.controls.setStatus('Save cancelled');
            } else {
                this.controls.setStatus('Save as failed: ' + e.message, 'error');
            }
            return false;
        }
    }

    // Keep saveSession as an alias for downloadModel (used by ARG panel binding)
    async saveSession() { return this.downloadModel(); }

    async closeCurrentFile() {
        if (this._editorDirty()) {
            var choice = await this.showDirtyCloseDialog();
            if (choice === 'save') {
                var saved = await this.save();
                if (!saved) {
                    return;
                }
            }
        }
        await this.newModel({ skipSaveCurrent: true });
    }

    /**
     * Start a new model. Saves any existing state first, then clears everything.
     */
    async newModel(options) {
        options = options || {};
        // Save current state before clearing
        if (!options.skipSaveCurrent && this._persistedFileContent) {
            IvyPersist.save(this);
        }

        // Create a fresh server session
        try {
            await this.api.createSession();
            var sessionEl = document.getElementById('session-id');
            if (sessionEl) {
                sessionEl.textContent = 'Session: ' + (IvyPersist.getSessionIdFromURL() || this.api.sessionId);
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
        this._persistedFilePath = '';
        this._persistedFileContent = '';
        this._persistedConceptRelations = null;
        this._fileHandle = null;
        this._savedFileContent = null;
        this.selectedArgNode = null;

        // Clear UI
        this.setEditorContent('');
        var editorLabel = document.getElementById('model-editor-label');
        if (editorLabel) {
            editorLabel.textContent = 'Model: (unsaved file)';
        }
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

        // Deduplicate: keep only the most recent entry per file path.
        // sessions is newest-first, so the first occurrence of each path is the most recent.
        var seen = {};
        var unique = [];
        for (var i = 0; i < sessions.length; i++) {
            var sess = sessions[i];
            if (!sess.fileName || sess.fileName === '(unnamed)') continue;
            var dedupKey = sess.filePath || sess.fileName;
            if (seen[dedupKey]) continue;
            seen[dedupKey] = true;
            unique.push(sess);
        }
        unique.sort(function (a, b) {
            return (a.fileName || '').localeCompare(b.fileName || '');
        });

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
                    self.flashAndClose(this, function () { self.loadRecentSession(s.id); });
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
            // Keep the URL set by restore() (state.sessionId) — do NOT clobber it with
            // api.sessionId, which resets to s1/s2/... on every server restart and would
            // overwrite unrelated historical sessions stored under those same IDs.
            IvyPersist.setFileName(state.fileName);
            var sessionEl = document.getElementById('session-id');
            if (sessionEl) {
                sessionEl.textContent = 'Session: ' + (IvyPersist.getSessionIdFromURL() || this.api.sessionId);
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
        this.controls.setStatus('Recompiling editor content...');
        try {
            // Always recompile from the current editor content so the
            // server checks exactly what the user sees, not a stale cache.
            var editorContent = this.cmEditor ? this.cmEditor.getValue() : this._persistedFileContent;
            if (editorContent) {
                await this.api.reloadContent(editorContent, this._persistedFileName || 'model.ivy');
            }
            this.controls.setStatus('Running ' + mode + ' check...');
            var result = await this.api.runCheck(mode);
            this.showCheckResult(result);

            // After check: refresh ARG to show counterexample states.
            var argData = await this.api.getARG();
            if (argData && argData.elements) {
                this.argGraph.update(argData.elements, argData.positions);
            }

            // After check: refresh concept graph to pick up new abstract_value.
            // Matches Python: view_state() → set_parent_state() → recompute().
            var conceptData = await this.api.getConceptGraph();
            if (conceptData && conceptData.elements) {
                this._lastConceptData = conceptData;
                this.conceptGraph.update(conceptData.elements, conceptData.positions);
            }

            // Auto-check "+" for used relations.
            // Matches Python ivy_ui_cti.py show_used_relations():
            // checks "+" for any relation whose formula mentions constants from the CTI.
            if (result && result.used_relations) {
                this._autoCheckUsedRelations(result.used_relations);
            }
        } catch (e) {
            this.controls.setStatus('Check failed: ' + e.message, 'error');
            console.error('Check error:', e);
        } finally {
            this.controls.hideLoading();
        }
    }

    /**
     * Auto-check the "+" checkbox for used relations after finding a CTI.
     * Matches Python ivy_ui_cti.py show_used_relations → show_relation(rel, '+').
     * @param {Array<string>} relationNames - names of relations to auto-check
     */
    _autoCheckUsedRelations(relationNames) {
        if (!relationNames || relationNames.length === 0) return;
        var usedSet = {};
        for (var i = 0; i < relationNames.length; i++) {
            usedSet[relationNames[i]] = true;
        }
        // Find checkbox rows and auto-check "+" for matching relations
        var tbody = document.getElementById('state-checkbox-body');
        if (!tbody) return;
        var rows = tbody.querySelectorAll('tr');
        for (var r = 0; r < rows.length; r++) {
            var nameCell = rows[r].querySelector('.name-col a');
            if (!nameCell) continue;
            var name = nameCell.textContent.trim();
            var baseName = name.split('(')[0];
            if (usedSet[name] || usedSet[baseName]) {
                // Auto-check the "+" checkbox (first checkbox, index 0)
                var inputs = rows[r].querySelectorAll('input[type="checkbox"]');
                if (inputs.length > 0 && !inputs[0].checked) {
                    inputs[0].checked = true;
                    // Trigger the toggle handler
                    this.onEdgeToggle(name, 'all_to_all', true);
                }
            }
        }
    }

    /**
     * Display the result of a verification check.
     */
    showCheckResult(result) {
        if (!result) return;

        // The backend returns {status:"ok", result:"pass"/"fail"/"error", ...}.
        // Use result.result (the verification outcome), not result.status (HTTP status).
        var verdict = result.result || result.status;
        var z3note = result.z3_contacted ? ' [Z3: yes]' : ' [Z3: no]';
        var mode = result.mode ? ' (' + result.mode + ')' : '';

        if (verdict === 'pass') {
            this.controls.setStatus('Check PASSED' + mode + z3note, 'success');
            this.controls.showInfo('Verification Result', 'PASSED' + z3note + ': ' + (result.message || 'All properties hold.'));
        } else if (verdict === 'fail') {
            this.controls.setStatus('Check FAILED' + mode + z3note + ' - counterexample found', 'error');
            this.controls.showInfo('Verification Result', 'FAILED' + z3note + ': ' + (result.message || 'Counterexample found.'));
            if (result.arg) {
                this.argGraph.update(result.arg.elements, result.arg.positions);
            }
        } else if (verdict === 'error') {
            this.controls.setStatus('Check ERROR' + mode + z3note, 'error');
            this.controls.showInfo('Verification Error', result.message || 'Unknown error');
        } else {
            this.controls.setStatus('Check result: ' + verdict + z3note);
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
    closeAllDropdowns() {
        var all = document.querySelectorAll('.dropdown.open');
        for (var j = 0; j < all.length; j++) {
            all[j].classList.remove('open');
        }
    }

    /**
     * Flash a menu item (macOS Cocoa style invert) then close dropdowns and invoke callback.
     */
    flashAndClose(el, callback) {
        var self = this;
        el.classList.add('menu-flash');
        setTimeout(function () {
            el.classList.remove('menu-flash');
            self.closeAllDropdowns();
            if (callback) callback();
        }, 50);
    }

    /**
     * Bind a menu item by ID to a callback, with dropdown auto-close.
     */
    bindMenuAction(id, callback) {
        var self = this;
        var el = document.getElementById(id);
        if (!el) return;
        el.addEventListener('click', function (e) {
            e.preventDefault();
            e.stopPropagation();
            self.flashAndClose(this, callback);
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
        this.controls.showLoading('Checking inductiveness...');
        try {
            var result = await this.api.runCheck('induction');
            if (result.result === 'fail' && result.failed_conjecture) {
                // Show dialog matching Python ivy_ui_cti.py:
                // "The following conjecture is not relatively inductive:"
                this.showTextDialog(
                    'ivyweb',
                    result.message || 'The following conjecture is not relatively inductive:',
                    result.failed_conjecture
                );
                this.controls.setStatus('Induction check: not inductive');
            } else if (result.result === 'pass') {
                // Success — show the invariant in a dialog
                this.showTextDialog(
                    'ivyweb',
                    'Inductive invariant found:',
                    result.message.replace('Inductive invariant found:\n', '')
                );
                this.controls.setStatus('Induction check: PASSED', 'success');
            } else {
                this.controls.setStatus('Induction check: ' + (result.message || result.result));
            }
        } catch (e) {
            this.controls.setStatus('Induction check failed: ' + e.message, 'error');
        } finally {
            this.controls.hideLoading();
        }
    }

    /**
     * Show a text dialog (matches Python ivy_ui_util.py text_dialog).
     * Displays a title, message, and editable text area with an OK button.
     * @param {string} title - Dialog title (e.g., "ivyweb")
     * @param {string} message - Message text above the text area
     * @param {string} text - Content for the text area (editable, selectable)
     */
    showTextDialog(title, message, text) {
        var overlay = document.getElementById('text-dialog-overlay');
        document.getElementById('text-dialog-title').textContent = title || 'ivyweb';
        document.getElementById('text-dialog-message').textContent = message || '';
        var textarea = document.getElementById('text-dialog-text');
        textarea.value = text || '';
        overlay.style.display = 'flex';

        // Select all text for easy copying
        textarea.focus();
        textarea.select();

        // OK button closes the dialog
        var okBtn = document.getElementById('text-dialog-ok');
        var handler = function () {
            overlay.style.display = 'none';
            okBtn.removeEventListener('click', handler);
        };
        okBtn.addEventListener('click', handler);

        // Also close on Escape key
        var escHandler = function (e) {
            if (e.key === 'Escape') {
                overlay.style.display = 'none';
                document.removeEventListener('keydown', escHandler);
            }
        };
        document.addEventListener('keydown', escHandler);
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

    /**
     * Show a non-modal toast notification. Auto-dismisses after 10s
     * or on click. Appends to document body so it floats over everything.
     */
    _showToast(message, level) {
        var toast = document.createElement('div');
        toast.className = 'ivy-toast ivy-toast-' + (level || 'info');
        toast.textContent = message;
        toast.style.cssText = 'position:fixed;top:20px;right:20px;z-index:10000;' +
            'padding:12px 20px;border-radius:6px;max-width:400px;cursor:pointer;' +
            'font-size:14px;box-shadow:0 4px 12px rgba(0,0,0,0.3);';
        if (level === 'error') {
            toast.style.background = '#d32f2f';
            toast.style.color = '#fff';
        } else {
            toast.style.background = '#333';
            toast.style.color = '#fff';
        }
        toast.onclick = function () { toast.remove(); };
        document.body.appendChild(toast);
        setTimeout(function () { toast.remove(); }, 10000);
    }
}

// ================================================================
// Initialize on DOM ready
// ================================================================
document.addEventListener('DOMContentLoaded', function () {
    window.ivyApp = new IvyApp();
    window.ivyApp.init();
});
