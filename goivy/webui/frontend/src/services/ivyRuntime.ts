/**
 * IvyRuntime - Service runtime for the Ivy Interactive Verification web UI.
 *
 * Orchestrates the API, graph views, and UI controls. Handles all
 * user interactions including ARG/concept graph clicks, context menus,
 * file loading, mode selection, and verification checks.
 */

import { ARG_STYLE as DEFAULT_ARG_STYLE, CONCEPT_STYLE as DEFAULT_CONCEPT_STYLE, IvyGraph as DefaultIvyGraph } from './graphRuntime.ts';
import { IvyAPIShim as DefaultIvyAPI, IvyControlsShim as DefaultIvyControls } from './runtimeShims.ts';
import { createIvyPersist } from './persistenceService.ts';
import {
    connectSessionEvents,
    createIvyApi,
    createSession,
    updateSessionDisplay as updateSessionDisplayViaService,
} from './sessionService.ts';
import {
    editorContent,
    editorDirty,
    editorTextElement,
    getEditorKeymap as getEditorKeymapViaService,
    refreshEditorLayout,
    scrollEditorToLine as scrollEditorToLineViaService,
    setEditorContent as setEditorContentViaService,
    setEditorKeymap as setEditorKeymapViaService,
    updateEditorLabel as updateEditorLabelViaService,
} from './editorService.ts';
import { initializeCodeMirrorEditor } from '../codeMirrorEditor.ts';
import { UIDataModel } from '../models/uiDataModel.ts';
import {
    chooseAndLoadModelFile as chooseAndLoadModelFileViaService,
    closeCurrentFile as closeCurrentFileViaService,
    confirmNoExternalChangeBeforeSave as confirmNoExternalChangeBeforeSaveViaService,
    downloadModel as downloadModelViaService,
    downloadModelForUnsupportedSave as downloadModelForUnsupportedSaveViaService,
    downloadTextFile as downloadTextFileViaService,
    ensureFileHandleWritable,
    loadModelFile,
    mergeDiskVersionIntoEditBuffer,
    newModel as newModelViaService,
    readFileHandleContent,
    rememberLastOpenFile,
    reopenLastFile as reopenLastFileViaService,
    restoreFileHandleForCurrentFile,
    saveModel,
    saveModelAs,
    updateReopenLastFileButton,
} from './fileService.ts';
import {
    loadRecentSession as loadRecentSessionViaService,
    populateRecentFiles as populateRecentFilesViaService,
} from './recentFileService.ts';
import {
    currentSheet as currentSheetViaService,
    graphElementsSnapshot as graphElementsSnapshotViaService,
    refreshGraphsAndEditorLayout,
    refreshLayoutAfterPatch,
    registerSheet as registerSheetViaService,
} from './graphService.ts';
import {
    applyEdgeVisibility,
    applyNodeLabels,
    displayConceptName,
    findEdgeVisibility,
    hydrateBackendToggleState,
    onEdgeToggle as onEdgeToggleViaService,
    onEdgeToggleChange as onEdgeToggleChangeViaService,
    onLabelToggleChange as onLabelToggleChangeViaService,
    populateStateCheckboxes as populateStateCheckboxesViaService,
    toggleChecked,
    updateStateLabel as updateStateLabelViaService,
} from './conceptVisibilityService.ts';
import { populateConstraintFacts as populateConstraintFactsViaService } from './detailsService.ts';
import {
    assertValidSheetId as assertValidSheetIdViaService,
    isValidSheetId as isValidSheetIdViaService,
    isVisualOnlySheet as isVisualOnlySheetViaService,
    setVisualOnlySheet as setVisualOnlySheetViaService,
    sheetExists as sheetExistsViaService,
    sheetTab as sheetTabViaService,
    switchSheet as switchSheetViaService,
    visualOnlyMessage as visualOnlyMessageViaService,
} from './sheetService.ts';
import {
    activeEventSheet as activeEventSheetViaService,
    addEventPattern as addEventPatternViaService,
    applyEventPatternResult as applyEventPatternResultViaService,
    clearEventPatterns as clearEventPatternsViaService,
    filterEventTrace as filterEventTraceViaService,
    findEventTrace as findEventTraceViaService,
    loadEventPatterns as loadEventPatternsViaService,
    loadEventTraceFile as loadEventTraceFileViaService,
    lookupEventTrace as lookupEventTraceViaService,
    readFileText as readFileTextViaService,
    removeSelectedEventPattern as removeSelectedEventPatternViaService,
    renderEventPatternList as renderEventPatternListViaService,
    saveEventPatterns as saveEventPatternsViaService,
    selectEventTraceRow as selectEventTraceRowViaService,
    selectedEventPattern as selectedEventPatternViaService,
    toggleEventTraceNode as toggleEventTraceNodeViaService,
    uncoverEventTraceAddress as uncoverEventTraceAddressViaService,
} from './eventTraceService.ts';
import {
    addCheckResultViewActions as addCheckResultViewActionsViaService,
    autoCheckUsedRelations,
    boundedCheck as boundedCheckViaService,
    checkInduction as checkInductionViaService,
    ctiConceptAction as ctiConceptActionViaService,
    runCheck as runCheckViaService,
    showCheckResult as showCheckResultViaService,
    weakenInvariant as weakenInvariantViaService,
} from './checkService.ts';
import {
    executeAndRefresh,
    exportConjecture as exportConjectureViaService,
    refreshConceptGraph as refreshConceptGraphViaService,
    rememberGraph as rememberGraphViaService,
    runAction as runActionViaService,
} from './analysisActionService.ts';
import {
    closeAllDropdowns as closeAllDropdownsViaService,
    dispatchMenuDescriptorAction as dispatchMenuDescriptorActionViaService,
    flashAndClose as flashAndCloseViaService,
} from './menuService.ts';
import {
    analysisStateLimits as analysisStateLimitsViaService,
    buildAnalysisState as buildAnalysisStateViaService,
    loadAnalysisStateFile as loadAnalysisStateFileViaService,
    loadAnalysisStateObject as loadAnalysisStateObjectViaService,
    removeAnalysisStateExtraSheets as removeAnalysisStateExtraSheetsViaService,
    saveAnalysisState as saveAnalysisStateViaService,
    validateAnalysisStateEvents as validateAnalysisStateEventsViaService,
    validateAnalysisStateGraphPayload as validateAnalysisStateGraphPayloadViaService,
    validateAnalysisStateObject as validateAnalysisStateObjectViaService,
    validateAnalysisStateSheet as validateAnalysisStateSheetViaService,
} from './analysisStateService.ts';
import {
    executeArgEdgeAction as executeArgEdgeActionViaService,
    executeArgNodeAction as executeArgNodeActionViaService,
    prepareArgNodeActionArgs as prepareArgNodeActionArgsViaService,
} from './argActionService.ts';
import {
    addProjection as addProjectionViaService,
    addRelationFromString as addRelationFromStringViaService,
    executeConceptEdgeAction as executeConceptEdgeActionViaService,
    executeConceptNodeAction as executeConceptNodeActionViaService,
    materializeEdge as materializeEdgeViaService,
    materializeEdgeFromSelected as materializeEdgeFromSelectedViaService,
    materializeNode as materializeNodeViaService,
    removeConcept as removeConceptViaService,
    selectConceptNode as selectConceptNodeViaService,
    splatterNode as splatterNodeViaService,
    splitConcept as splitConceptViaService,
    supposeEmpty as supposeEmptyViaService,
} from './conceptActionService.ts';
const defaultRuntimeDependencies = {
    IvyAPI: DefaultIvyAPI,
    IvyControls: DefaultIvyControls,
    IvyGraph: DefaultIvyGraph,
    IvyPersist: createIvyPersist(globalThis.window),
    ARG_STYLE: DEFAULT_ARG_STYLE,
    CONCEPT_STYLE: DEFAULT_CONCEPT_STYLE,
    CodeMirror: globalThis.window && globalThis.window.CodeMirror,
};
let runtimeDeps = { ...defaultRuntimeDependencies };

function configureIvyRuntimeDependencies(overrides) {
    runtimeDeps = { ...defaultRuntimeDependencies, ...(overrides || {}) };
    return runtimeDeps;
}

function resetIvyRuntimeDependencies() {
    runtimeDeps = { ...defaultRuntimeDependencies };
    return runtimeDeps;
}

class IvyRuntime {
    [key: string]: any;

    constructor() {
        this.api = this.createApi();
        this.controls = new runtimeDeps.IvyControls(this.api);
        this.argGraph = null;
        this.conceptGraph = null;
        this.uiDataModel = new UIDataModel();
        this.sheets = {};
        this.activeSheetId = 'sheet-1';
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
        this._lastClosedFileHandle = null;
        this._lastClosedSessionId = '';
        this._lastClosedFileName = '';
        // Content as last written to disk; used to detect unsaved changes.
        this._savedFileContent = null;
        this._saveInProgress = false;
        this._saveProgressSheen = null;
        this.currentBound = 10;
    }

    createApi() {
        return createIvyApi({
            fallbackApiFactory: function () {
                return new runtimeDeps.IvyAPI();
            },
        });
    }

    /**
     * Initialize the application: create session, build graphs, wire events.
     */
    async init() {
        var self = this;
        this.controls.setStatus('Initializing...');

        // Check for a saved session BEFORE creating a new server session.
        // This prevents the URL session ID from incrementing on every reload.
        var savedState = runtimeDeps.IvyPersist.load();

        // Always need a server session for API calls.
        try {
            await createSession(this.api, { controls: this.controls });
        } catch (e) {
            // createSession has already reported the failure through controls.
        }

        // If restoring, keep the saved session's URL hash.
        // If fresh, set the new session ID in the URL.
        if (!savedState || !savedState.fileContent) {
            runtimeDeps.IvyPersist.setSessionIdInURL(this.api.sessionId);
        }

        // Show the persisted session ID (from URL hash), not the server session ID.
        // These can differ because the server ID increments on restart while
        // the persisted ID is stable across reloads.
        this.updateSessionDisplay(runtimeDeps.IvyPersist.getSessionIdFromURL() || this.api.sessionId);
        this.uiDataModel.setSessionMetadata({
            id: runtimeDeps.IvyPersist.getSessionIdFromURL() || this.api.sessionId || '',
        });

        // Create Cytoscape graph instances
        this.argGraph = new runtimeDeps.IvyGraph('arg-graph', runtimeDeps.ARG_STYLE);
        this.conceptGraph = new runtimeDeps.IvyGraph('concept-graph', runtimeDeps.CONCEPT_STYLE);
        this.registerSheet('sheet-1', this.argGraph, this.conceptGraph);

        // Health check: verify graphs initialized correctly.
        this.argGraph.healthCheck();
        this.conceptGraph.healthCheck();

        // Hook concept graph updates to auto-apply edge visibility.
        // Matches Python: edges are hidden by default, shown only when
        // the corresponding checkbox (+/?/-) in the state panel is checked.
        // Wire up all event handlers
        this.setupEventHandlers();
        await this.loadMenuDescriptors();
        this.setupTabs();
        this.setupResizer();
        this.setupResizer2();
        this.setupResizer3();
        this.setupResizerH();
        this.setupDetailsResizer();
        this.setupTutorialUrlBar();
        this.setupKeyboardShortcuts();

        // Connect to SSE for real-time updates
        if (this.api.sessionId) {
            connectSessionEvents(this.api, this.handleEvent.bind(this), function () {
                self.controls.setStatus('Server connection lost', 'error');
                self._showToast('Connection to server lost. Check that the server is running and reload the page.', 'error');
            });
        }

        // Initialize CodeMirror on the model editor textarea.
        // Must happen BEFORE restore so setEditorContent() can call cmEditor.setValue().
        var modelEditor = document.getElementById('model-editor');
        if (modelEditor) {
            var codeMirror = runtimeDeps.CodeMirror || window.CodeMirror;
            this.cmEditor = initializeCodeMirrorEditor({
                runtime: this,
                keymap: this.getEditorKeymap(),
                doc: document,
                codeMirror: codeMirror,
            });
            var radios = document.querySelectorAll('input[name="keymap"]');
            for (var i = 0; i < radios.length; i++) {
                radios[i].addEventListener('change', function () {
                    self.setEditorKeymap(this.value);
                });
            }
        }

        // Restore saved session if available (survives page reload).
        if (savedState && savedState.fileContent) {
            console.log('IvyPersist: restoring session', savedState.sessionId, savedState.fileName);
            var restored = await runtimeDeps.IvyPersist.restore(this, savedState);
            if (restored) {
                // Keep the URL hash from the saved session (don't overwrite)
                runtimeDeps.IvyPersist.setFileName(savedState.fileName, savedState.filePath);
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
            runtimeDeps.IvyPersist.save(self);
        });

        // Hook into setStatus: auto-save whenever a 'success' status is set.
        var origSetStatus = this.controls.setStatus.bind(this.controls);
        this.controls.setStatus = function (msg, level) {
            origSetStatus(msg, level);
            if (level === 'success') {
                runtimeDeps.IvyPersist.save(self);
            }
        };
    }

    // --- Editor helpers (CodeMirror) ---

    updateSessionDisplay(sessionId) {
        updateSessionDisplayViaService(sessionId);
    }

    _setLayoutSize(method, value) {
        var target = null;
        if (method === 'setArgPanelWidth') target = document.getElementById('arg-panel');
        if (method === 'setStatePanelWidth') target = document.getElementById('state-panel');
        if (method === 'setEditorWidth') target = document.getElementById('editor-panel');
        if (method === 'setTutorialHeight') target = document.getElementById('tutorial-container');
        if (method === 'setDetailsHeight') target = document.getElementById('info-panel');
        if (!target) return false;
        var size = Math.max(0, Number(value) || 0);
        target.style.flex = '0 0 ' + size + 'px';
        if (method === 'setTutorialHeight' || method === 'setDetailsHeight') {
            target.style.height = size + 'px';
        }
        return true;
    }

    setEditorContent(content) {
        setEditorContentViaService(this, content);
    }

    _updateEditorLabel() {
        updateEditorLabelViaService(this, {
            updateReopenLastFileButton: this._updateReopenLastFileButton.bind(this),
        });
    }

    _refreshEditorLayout() {
        refreshEditorLayout(this);
    }

    _editorContent() {
        return editorContent(this);
    }

    _editorDirty() {
        return editorDirty(this);
    }

    _showSaveProgress(message) {
        this._hideSaveProgress();
        this._saveInProgress = true;
        this._updateEditorLabel();
        this._saveProgressSheen = this._showSaveEditorSheen();
        return true;
    }

    _editorTextElement() {
        return editorTextElement(this);
    }

    _showSaveEditorSheen() {
        var target = this._editorTextElement();
        if (!target || typeof target.getBoundingClientRect !== 'function') return null;
        var rect = target.getBoundingClientRect();
        if (!rect || rect.width <= 0 || rect.height <= 0) return null;
        var sheen = document.createElement('div');
        sheen.className = 'ivy-save-editor-sheen';
        sheen.setAttribute('aria-hidden', 'true');
        sheen.style.cssText = 'position:fixed;z-index:9999;pointer-events:none;' +
            'background:rgba(128,128,128,0.2);';
        sheen.style.left = Math.round(rect.left) + 'px';
        sheen.style.top = Math.round(rect.top) + 'px';
        sheen.style.width = Math.round(rect.width) + 'px';
        sheen.style.height = Math.round(rect.height) + 'px';
        document.body.appendChild(sheen);
        return sheen;
    }

    _hideSaveProgress() {
        if (this._saveProgressSheen && this._saveProgressSheen.parentNode) {
            this._saveProgressSheen.parentNode.removeChild(this._saveProgressSheen);
        }
        this._saveProgressSheen = null;
        if (this._saveInProgress) {
            this._saveInProgress = false;
            this._updateEditorLabel();
        }
    }

    _rememberLastOpenFile() {
        rememberLastOpenFile(this, runtimeDeps.IvyPersist);
    }

    _updateReopenLastFileButton() {
        updateReopenLastFileButton(this);
    }

    async reopenLastFile() {
        return reopenLastFileViaService(this, runtimeDeps.IvyPersist);
    }

    async _ensureFileHandleWritable() {
        return ensureFileHandleWritable(this);
    }

    async _readFileHandleContent() {
        return readFileHandleContent(this);
    }

    async _restoreFileHandleForCurrentFile() {
        return restoreFileHandleForCurrentFile(this, runtimeDeps.IvyPersist);
    }

    async _confirmNoExternalChangeBeforeSave(content) {
        return confirmNoExternalChangeBeforeSaveViaService(this, content, runtimeDeps.IvyPersist);
    }

    _mergeDiskVersionIntoEditBuffer(baseContent, editorContent, diskContent) {
        return mergeDiskVersionIntoEditBuffer(baseContent, editorContent, diskContent);
    }

    showExternalChangeDialog() {
        return this.buttonListDialog(
            'File changed on disk',
            'The file "' + (this._persistedFileName || 'model') + '" has changed outside IvyWeb. Choose how to handle the current editor buffer.',
            [
                { label: 'Do nothing', value: 'do-nothing' },
                { label: 'Revert to on-disk version', value: 'reload' },
                { label: 'Merge disk version into edit buffer', value: 'merge' },
                { label: 'Overwrite on-disk with edited buffer', value: 'overwrite', danger: true },
            ],
        );
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
        return this.buttonListDialog(
            'File has changed',
            'File "' + (this._persistedFileName || 'model') + '" has changed. Save before closing?',
            [
                { label: 'Cancel the close', value: 'cancel' },
                { label: 'Save', value: 'save' },
                { label: 'Discard Edits', value: 'discard', danger: true },
            ],
        );
        var self = this;
        return new Promise(function (resolve) {
            var overlay = document.getElementById('dirty-close-dialog-overlay');
            var msg = document.getElementById('dirty-close-dialog-message');
            var cancel = document.getElementById('dirty-close-cancel');
            var save = document.getElementById('dirty-close-save');
            var discard = document.getElementById('dirty-close-discard');
            if (!overlay || !msg || !cancel || !save || !discard) {
                resolve('cancel');
                return;
            }
            msg.textContent = 'File "' + (self._persistedFileName || 'model') + '" has changed. Save before closing?';
            overlay.style.display = 'flex';
            cancel.focus();

            var done = function (choice) {
                overlay.style.display = 'none';
                cancel.removeEventListener('click', onCancel);
                save.removeEventListener('click', onSave);
                discard.removeEventListener('click', onDiscard);
                document.removeEventListener('keydown', onKeyDown);
                resolve(choice);
            };
            var onCancel = function () { done('cancel'); };
            var onSave = function () { done('save'); };
            var onDiscard = function () { done('discard'); };
            var onKeyDown = function (e) {
                if (e.key === 'Enter' || e.key === 'Escape') {
                    e.preventDefault();
                    done('cancel');
                }
            };
            cancel.addEventListener('click', onCancel);
            save.addEventListener('click', onSave);
            discard.addEventListener('click', onDiscard);
            document.addEventListener('keydown', onKeyDown);
        });
    }

    showSaveAsExplanationNotice() {
        var notice = document.getElementById('save-as-explain-notice');
        if (notice) {
            notice.style.display = 'block';
        }
    }

    hideSaveAsExplanationNotice() {
        var notice = document.getElementById('save-as-explain-notice');
        if (notice) {
            notice.style.display = 'none';
        }
    }

    scrollEditorToLine(lineno) {
        scrollEditorToLineViaService(this, lineno);
    }

    currentSheet() {
        return currentSheetViaService(this);
    }

    registerSheet(sheetId, argGraph, conceptGraph) {
        registerSheetViaService(this, sheetId, argGraph, conceptGraph);
    }

    acceptArgSnapshot(sheetId, payload) {
        if (!this.uiDataModel) return null;
        return this.uiDataModel.acceptArgSnapshot(sheetId || this.activeSheetId || 'sheet-1', payload || {});
    }

    acceptConceptSnapshot(sheetId, payload) {
        if (!this.uiDataModel) return null;
        return this.uiDataModel.acceptConceptSnapshot(sheetId || this.activeSheetId || 'sheet-1', payload || {});
    }

    isVisualOnlySheet(sheetId) {
        return isVisualOnlySheetViaService(this, sheetId);
    }

    setVisualOnlySheet(sheetId, visualOnly) {
        return setVisualOnlySheetViaService(this, sheetId, visualOnly);
    }

    visualOnlyMessage(kind) {
        return visualOnlyMessageViaService(kind);
    }

    installConceptGraphVisibilityHook(conceptGraph) {
        if (!conceptGraph || conceptGraph._ivyVisibilityHooked) return;
        var self = this;
        var origUpdate = conceptGraph.update.bind(conceptGraph);
        conceptGraph.update = function(elements, positions) {
            origUpdate(elements, positions);
            if (conceptGraph.cy && typeof conceptGraph.cy.edges === 'function') {
                self._applyEdgeVisibility(conceptGraph);
            }
        };
        conceptGraph._ivyVisibilityHooked = true;
    }

    attachGraphEventHandlers(argGraph, conceptGraph, sheetId) {
        var self = this;
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        if (argGraph && argGraph.containerId) {
            var argEl = document.getElementById(argGraph.containerId);
            if (argEl && !argEl._ivyContextSuppressed) {
                argEl.addEventListener('contextmenu', function (e) { e.preventDefault(); });
                argEl._ivyContextSuppressed = true;
            }
        }
        if (conceptGraph && conceptGraph.containerId) {
            var conceptEl = document.getElementById(conceptGraph.containerId);
            if (conceptEl && !conceptEl._ivyContextSuppressed) {
                conceptEl.addEventListener('contextmenu', function (e) { e.preventDefault(); });
                conceptEl._ivyContextSuppressed = true;
            }
        }

        argGraph.onNodeClick(function (nodeData) {
            self.onArgNodeClick(nodeData, sheetId);
        });
        argGraph.onNodeRightClick(function (nodeData, pos) {
            self.onArgNodeRightClick(nodeData, pos, sheetId);
        });
        argGraph.onEdgeClick(function (edgeData) {
            self.controls.showInfo(edgeData.short_info, edgeData.long_info);
        });
        argGraph.onEdgeRightClick(function (edgeData, pos) {
            self.onArgEdgeRightClick(edgeData, pos, sheetId);
        });
        argGraph.onBackgroundClick(function () {
            self.controls.clearInfo();
            self.controls.hideContextMenu();
        });

        conceptGraph.onNodeRightClick(function (nodeData, pos) {
            self.onConceptNodeRightClick(nodeData, pos);
        });
        conceptGraph.onNodeClick(function (nodeData, evt) {
            try {
                var node = evt.target;
                var name = nodeData.obj || nodeData.id;
                if (node.hasClass('selected_node')) {
                    node.removeClass('selected_node');
                    node.unselect();
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
        conceptGraph.onEdgeClick(function (edgeData, evt) {
            var edge = evt.target;
            var name = edgeData.obj || edgeData.label || edgeData.id;
            if (edge.hasClass('selected_edge')) {
                edge.removeClass('selected_edge');
                edge.unselect();
                edge.removeStyle('line-color target-arrow-color source-arrow-color width');
                self.controls.setStatus('Deselected: ' + name);
                self.controls.clearInfo();
            } else {
                edge.addClass('selected_edge');
                self.controls.setStatus('Selected: ' + name);
                self.controls.showInfo(edgeData.short_info, edgeData.long_info);
            }
        });
        conceptGraph.onEdgeRightClick(function (edgeData, pos) {
            self.onConceptEdgeRightClick(edgeData, pos);
        });
        conceptGraph.onBackgroundClick(function () {
            self.controls.clearInfo();
            self.controls.hideContextMenu();
        });
    }

    async chooseAndLoadModelFile() {
        return chooseAndLoadModelFileViaService(this);
    }

    async chooseAndLoadEventTraceFile() {
        var eventFileInput = document.getElementById('event-file-input');
        if (window.showOpenFilePicker) {
            try {
                var handles = await window.showOpenFilePicker({
                    types: [{ description: 'Ivy event traces', accept: { 'text/plain': ['.iev', '.txt'] } }],
                    multiple: false,
                });
                var file = await handles[0].getFile();
                await this.loadEventTraceFile(file);
            } catch (ex) {
                if (ex.name !== 'AbortError') {
                    this.controls.setStatus('Event trace load failed: ' + ex.message, 'error');
                }
            }
        } else if (eventFileInput) {
            eventFileInput.click();
        }
    }

    chooseAndLoadAnalysisStateFile() {
        var analysisStateFileInput = document.getElementById('analysis-state-file-input');
        if (analysisStateFileInput) {
            analysisStateFileInput.click();
        }
    }

    /**
     * Set up all DOM and graph event handlers.
     */
    setupEventHandlers() {
        var self = this;

        // --- File Menu ---
        var fileInput = document.getElementById('file-input');
        var eventFileInput = document.getElementById('event-file-input');
        var analysisStateFileInput = document.getElementById('analysis-state-file-input');

        if (fileInput) fileInput.addEventListener('change', function () {
            if (fileInput.files && fileInput.files.length > 0) {
                self.loadFile(fileInput.files[0]);
                fileInput.value = ''; // reset for re-selection of same file
            }
        });

        if (eventFileInput) {
            eventFileInput.addEventListener('change', function () {
                if (eventFileInput.files && eventFileInput.files.length > 0) {
                    self.loadEventTraceFile(eventFileInput.files[0]);
                    eventFileInput.value = '';
                }
            });
        }

        if (analysisStateFileInput) {
            analysisStateFileInput.addEventListener('change', async function () {
                if (analysisStateFileInput.files && analysisStateFileInput.files.length > 0) {
                    try {
                        await self.loadAnalysisStateFile(analysisStateFileInput.files[0]);
                    } catch (ex) {
                        self.controls.setStatus('Load analysis state failed: ' + ex.message, 'error');
                    }
                    analysisStateFileInput.value = '';
                }
            });
        }

        // File > Load...
        // Use showOpenFilePicker when available so we get a writable FileSystemFileHandle,
        // enabling Ctrl+S to save directly without re-prompting. Fall back to <input> otherwise.
        var loadModel = document.getElementById('file-load');
        if (loadModel) loadModel.addEventListener('click', function (e) {
            e.preventDefault();
            self.flashAndClose(this, async function () {
                await self.chooseAndLoadModelFile();
            });
        });

        var eventTraceOpen = document.getElementById('file-open-event-trace');
        if (eventTraceOpen && eventFileInput) {
            eventTraceOpen.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, async function () {
                    await self.chooseAndLoadEventTraceFile();
                });
            });
        }

            // File > Save as... (uses File System Access API to write to a chosen path)
            var saveAs = document.getElementById('file-save-as');
            if (saveAs) saveAs.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, function () { self.saveAs(); });
            });

            // File > Download current model (browser download)
            var download = document.getElementById('file-download');
            if (download) download.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, function () { self.downloadModel(); });
            });

            var saveAnalysis = document.getElementById('file-save-analysis-state');
            if (saveAnalysis) {
                saveAnalysis.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.flashAndClose(this, function () { self.saveAnalysisState(); });
                });
            }
            var loadAnalysis = document.getElementById('file-load-analysis-state');
            if (loadAnalysis && analysisStateFileInput) {
                loadAnalysis.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.flashAndClose(this, function () { self.chooseAndLoadAnalysisStateFile(); });
                });
            }

            // File > New Model
            var newModel = document.getElementById('file-new');
            if (newModel) newModel.addEventListener('click', function (e) {
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
            var reopenLast = document.getElementById('file-reopen-last');
            if (reopenLast) {
                reopenLast.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.reopenLastFile();
                });
            }

            // File > Save Invariant
            var saveInvariant = document.getElementById('file-save-invariant');
            if (saveInvariant) saveInvariant.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, function () { self.saveInvariant(); });
            });

            // --- Check ---
            document.getElementById('btn-check').addEventListener('click', function () {
                self.runCheck();
            });

            // --- Show reachable states ---
            document.getElementById('btn-show-reachable').addEventListener('click', function () {
                self.showReachableStates();
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

            this._setupJobControlHandlers();

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
            this.bindMenuAction('conj-cti-gather', function () { self.ctiConceptAction('cti_gather'); });
            this.bindMenuAction('conj-cti-minimize', function () { self.ctiConceptAction('cti_minimize'); });
            this.bindMenuAction('conj-cti-check-sufficient', function () { self.ctiConceptAction('cti_check_sufficient'); });
            this.bindMenuAction('conj-cti-check-inductive', function () { self.ctiConceptAction('cti_check_inductive'); });
            this.bindMenuAction('conj-cti-strengthen', function () { self.ctiConceptAction('cti_strengthen'); });
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

        this.attachGraphEventHandlers(this.argGraph, this.conceptGraph, 'sheet-1');
    }

    _setupJobControlHandlers() {
        var self = this;
        var toggleButton = document.getElementById('btn-toggle-job-control');
        var closeButton = document.getElementById('job-control-close');
        var modeToggle = document.getElementById('job-submission-toggle');
        if (toggleButton && !toggleButton._ivyJobControlBound) {
            toggleButton.addEventListener('click', function () {
                self._toggleJobControlPage();
            });
            toggleButton._ivyJobControlBound = true;
        }
        if (closeButton && !closeButton._ivyJobControlBound) {
            closeButton.addEventListener('click', function () {
                self._setJobControlOpen(false);
            });
            closeButton._ivyJobControlBound = true;
        }
        if (modeToggle && !modeToggle._ivyJobControlBound) {
            modeToggle.addEventListener('click', function () {
                var nextMode = modeToggle.getAttribute('data-mode') === 'remote' ? 'browser' : 'remote';
                self._setJobSubmissionMode(nextMode);
            });
            modeToggle._ivyJobControlBound = true;
        }
        this._setJobSubmissionMode(this.jobSubmissionMode || 'browser');
    }

    _toggleJobControlPage() {
        var page = document.getElementById('job-control-page');
        this._setJobControlOpen(!(page && page.classList.contains('open')));
    }

    _setJobControlOpen(open) {
        var page = document.getElementById('job-control-page');
        var button = document.getElementById('btn-toggle-job-control');
        if (!page) return;
        page.classList.toggle('open', !!open);
        page.setAttribute('aria-hidden', open ? 'false' : 'true');
        if (button) {
            button.classList.toggle('active', !!open);
            button.setAttribute('aria-expanded', open ? 'true' : 'false');
        }
    }

    _setJobSubmissionMode(mode) {
        var normalized = mode === 'remote' ? 'remote' : 'browser';
        this.jobSubmissionMode = normalized;
        var toggle = document.getElementById('job-submission-toggle');
        if (!toggle) return normalized;
        toggle.setAttribute('data-mode', normalized);
        toggle.setAttribute('aria-pressed', normalized === 'remote' ? 'true' : 'false');
        toggle.classList.toggle('is-browser', normalized === 'browser');
        toggle.classList.toggle('is-remote', normalized === 'remote');
        toggle.title = 'Job submission: ' + normalized;
        return normalized;
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
            if (!(activePanel.id === 'arg-panel' && self._setLayoutSize('setArgPanelWidth', newWidth))) {
                activePanel.style.flex = '0 0 ' + newWidth + 'px';
            }
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
            if (!self._setLayoutSize('setStatePanelWidth', newWidth)) {
                rightSection.style.flex = '0 0 ' + newWidth + 'px';
            }
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
        var sheetArea = document.getElementById('sheet-area');
        var editorPanel = document.getElementById('editor-panel');
        var topRow = document.getElementById('top-row');
        if (!sheetArea || !editorPanel || !topRow) return;
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;
        var minEditorWidth = 200;
        var minSheetAreaWidth = 400;

        divider3.addEventListener('mousedown', function (e) {
            isDragging = true;
            startX = e.clientX;
            startWidth = editorPanel.offsetWidth;
            sheetArea.style.flex = '1 1 auto';
            divider3.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = startX - e.clientX; // drag left = editor wider
            var newWidth = startWidth + dx;
            var dividerWidth = divider3.offsetWidth || 4;
            var maxW = Math.max(minEditorWidth, topRow.offsetWidth - dividerWidth - minSheetAreaWidth);
            newWidth = Math.max(minEditorWidth, Math.min(newWidth, maxW));
            if (!self._setLayoutSize('setEditorWidth', newWidth)) {
                editorPanel.style.flex = '0 0 ' + newWidth + 'px';
            }
            if (self.argGraph) self.argGraph.resize();
            if (self.conceptGraph) self.conceptGraph.resize();
            self._refreshEditorLayout();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                divider3.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                if (self.argGraph) self.argGraph.resize();
                if (self.conceptGraph) self.conceptGraph.resize();
                self._refreshEditorLayout();
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

    isValidSheetId(sheetId) {
        return isValidSheetIdViaService(sheetId);
    }

    assertValidSheetId(sheetId) {
        return assertValidSheetIdViaService(sheetId);
    }

    sheetTab(sheetId) {
        return sheetTabViaService(sheetId);
    }

    sheetExists(sheetId) {
        return sheetExistsViaService(this, sheetId);
    }

    eventTraceRow(sheetId, address) {
        var sheet = document.getElementById(sheetId);
        if (!sheet) return null;
        var rows = sheet.querySelectorAll('.event-row');
        for (var i = 0; i < rows.length; i++) {
            if (rows[i].getAttribute('data-event-address') === address) {
                return rows[i];
            }
        }
        return null;
    }

    /**
     * Switch to a sheet by ID.
     */
    switchSheet(sheetId) {
        return switchSheetViaService(this, sheetId);
    }

    /**
     * Add a new sheet tab (matches Python ui_parent.add(art)).
     * Called on "Step into" / "Decompose" actions.
     * @param {string} [label] - Tab label (default: "Sheet N")
     * @returns {string} The new sheet ID
     */
    addSheet(label, preferredSheetId) {
        this._sheetCounter++;
        var sheetId = preferredSheetId || ('sheet-' + this._sheetCounter);
        while (!preferredSheetId && this.sheetExists(sheetId)) {
            this._sheetCounter++;
            sheetId = 'sheet-' + this._sheetCounter;
        }
        this.assertValidSheetId(sheetId);
        if (preferredSheetId && this.sheetExists(sheetId)) {
            throw new Error('duplicate sheet id: ' + sheetId);
        }
        var match = /^sheet-(\d+)$/.exec(sheetId);
        if (match) {
            this._sheetCounter = Math.max(this._sheetCounter, parseInt(match[1], 10));
        }
        label = label || ('Sheet ' + this._sheetCounter);

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
        var fallbackGraphs = newSheet.querySelectorAll('.graph-container');
        for (var g = 0; g < fallbackGraphs.length; g++) {
            fallbackGraphs[g].innerHTML = '';
            fallbackGraphs[g].id = fallbackGraphs[g].id + '-' + this._sheetCounter;
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

        var graphs = newSheet.querySelectorAll('.graph-container');
        var graphIds = [];
        for (var i = 0; i < graphs.length; i++) {
            graphs[i].innerHTML = '';
            graphIds.push(graphs[i].id);
        }

        var argGraph = new runtimeDeps.IvyGraph(graphIds[0], runtimeDeps.ARG_STYLE);
        var conceptGraph = new runtimeDeps.IvyGraph(graphIds[1], runtimeDeps.CONCEPT_STYLE);
        argGraph.healthCheck();
        conceptGraph.healthCheck();
        this.registerSheet(sheetId, argGraph, conceptGraph);
        this.attachGraphEventHandlers(argGraph, conceptGraph, sheetId);

        // Switch to the new sheet
        this.switchSheet(sheetId);
        this.controls.setStatus('Opened: ' + label);
        return sheetId;
    }

    openARGSheet(label, argData, preferredSheetId) {
        var sheetId = this.addSheet(label, preferredSheetId);
        this.acceptArgSnapshot(sheetId, argData || {});
        var sheetState = this.sheets && this.sheets[sheetId];
        if (sheetState && sheetState.argGraph && argData && argData.elements) {
            sheetState.argGraph.update(argData.elements, argData.positions);
        }
        return sheetId;
    }

    openEventTraceSheet(label, data, preferredSheetId) {
        data = data || {};
        var events = data.events || [];
        var patterns = data.patterns || [];
        this._sheetCounter = this._sheetCounter || 1;
        this._sheetCounter++;
        var sheetId = preferredSheetId || data.sheet_id || ('events-' + this._sheetCounter);
        this.assertValidSheetId(sheetId);
        var sheetArea = document.getElementById('sheet-area');
        if (!sheetArea) return '';

        var existingTab = this.sheetTab(sheetId);
        var existingSheet = document.getElementById(sheetId);
        var tabLabel = label || data.label || 'Events';
        var tabBar = document.getElementById('tab-bar');
        if (!tabBar) return '';
        if (!existingTab) {
            var tabBtn = document.createElement('button');
            tabBtn.className = 'sheet-tab';
            tabBtn.setAttribute('data-sheet', sheetId);
            var labelSpan = document.createElement('span');
            labelSpan.textContent = tabLabel;
            tabBtn.appendChild(labelSpan);
            var closeBtn = document.createElement('span');
            closeBtn.className = 'tab-close';
            closeBtn.textContent = '\u00D7';
            closeBtn.title = 'Close tab';
            tabBtn.appendChild(closeBtn);
            tabBar.appendChild(tabBtn);
        } else {
            var existingLabel = existingTab.querySelector('span');
            if (existingLabel) {
                existingLabel.textContent = tabLabel;
            }
        }

        var sheet = existingSheet;
        if (!sheet) {
            sheet = document.createElement('div');
            sheet.id = sheetId;
            sheet.className = 'sheet-content event-sheet';
            sheetArea.appendChild(sheet);
        }

        this.sheets[sheetId] = {
            id: sheetId,
            type: 'events',
            events: events,
            patterns: patterns.slice(),
            selectedEventAddress: data.selected_address || null,
        };
        this.renderEventTraceSheet(sheetId);
        this.switchSheet(sheetId);
        this.controls.setStatus('Opened: ' + (label || data.label || 'Events'));
        return sheetId;
    }

    async loadEventTraceFile(file) {
        return loadEventTraceFileViaService(this, file);
        if (!file) return null;
        this.controls.setStatus('Loading event trace...');
        try {
            var text = await this.readFileText(file);
            var result = await this.api.executeAction('events_parse', { content: text, filename: file.name || '' });
            this.openEventTraceSheet(file.name || result.label || 'Event trace', result || {}, result && result.sheet_id);
            this.controls.setStatus('Loaded event trace: ' + (file.name || 'trace'), 'success');
            return result;
        } catch (e) {
            this.controls.setStatus('Event trace load failed: ' + e.message, 'error');
            throw e;
        }
    }

    readFileText(file) {
        return readFileTextViaService(file);
        if (file && typeof file.text === 'function') {
            return file.text();
        }
        return new Promise(function (resolve, reject) {
            var reader = new FileReader();
            reader.onload = function () { resolve(String(reader.result || '')); };
            reader.onerror = function () { reject(reader.error || new Error('failed to read file')); };
            reader.readAsText(file);
        });
    }

    renderEventTraceSheet(sheetId) {
        var sheetState = this.sheets && this.sheets[sheetId];
        var sheet = document.getElementById(sheetId);
        if (!sheetState || !sheet) return;
        sheet.classList.add('event-sheet');
        sheet.innerHTML = [
            '<div class="event-viewer">',
            '  <div class="event-tree-panel panel">',
            '    <div class="panel-header">',
            '      <span class="panel-title">Events</span>',
            '      <button class="menu-btn event-filter-btn" type="button">Filter...</button>',
            '      <button class="menu-btn event-find-fwd-btn" type="button">&gt;&gt;</button>',
            '      <button class="menu-btn event-find-rev-btn" type="button">&lt;&lt;</button>',
            '    </div>',
            '    <div class="event-tree" data-event-tree="' + sheetId + '"></div>',
            '  </div>',
            '  <div class="event-pattern-panel panel">',
            '    <div class="panel-header"><span class="panel-title">Patterns</span></div>',
            '    <select class="event-pattern-list" size="8"></select>',
            '    <div class="event-pattern-buttons">',
            '      <button class="menu-btn event-pattern-rev" type="button">&lt;&lt;</button>',
            '      <button class="menu-btn event-pattern-fwd" type="button">&gt;&gt;</button>',
            '      <button class="menu-btn event-pattern-add" type="button">+</button>',
            '      <button class="menu-btn event-pattern-remove" type="button">-</button>',
            '    </div>',
            '    <div class="event-pattern-buttons">',
            '      <button class="menu-btn event-pattern-save" type="button">Save</button>',
            '      <button class="menu-btn event-pattern-load" type="button">Load</button>',
            '      <button class="menu-btn event-pattern-clear" type="button">Clear</button>',
            '    </div>',
            '  </div>',
            '</div>',
        ].join('');

        var tree = sheet.querySelector('.event-tree');
        this.renderEventTree(tree, sheetState.events || [], sheetId, '');
        this.renderEventPatternList(sheetId);
        this.attachEventTraceHandlers(sheetId);
        if (sheetState.selectedEventAddress) {
            this.selectEventTraceRow(sheetId, sheetState.selectedEventAddress);
        }
    }

    renderEventTree(container, events, sheetId, prefix) {
        if (!container) return;
        container.innerHTML = '';
        var list = document.createElement('ul');
        list.className = 'event-tree-list';
        for (var i = 0; i < (events || []).length; i++) {
            list.appendChild(this.renderEventTreeNode(events[i], sheetId, prefix === '' ? String(i) : prefix + '/' + i));
        }
        container.appendChild(list);
    }

    renderEventTreeNode(ev, sheetId, fallbackAddress) {
        ev = ev || {};
        var address = ev.address || fallbackAddress;
        ev.address = address;
        var li = document.createElement('li');
        li.className = 'event-tree-node';
        li.setAttribute('data-event-node', address);

        var row = document.createElement('div');
        row.className = 'event-row';
        row.setAttribute('data-event-address', address);

        var hasSubs = ev.subs && ev.subs.length > 0;
        var toggle = document.createElement('button');
        toggle.className = 'event-toggle';
        toggle.type = 'button';
        toggle.textContent = hasSubs ? '+' : '';
        toggle.disabled = !hasSubs;
        if (hasSubs) toggle.setAttribute('data-event-toggle', address);
        row.appendChild(toggle);

        var text = document.createElement('span');
        text.className = 'event-text';
        text.textContent = ev.text || '';
        row.appendChild(text);
        li.appendChild(row);

        row.addEventListener('click', this.selectEventTraceRow.bind(this, sheetId, address));
        if (hasSubs) {
            toggle.addEventListener('click', function (e) {
                e.stopPropagation();
                this.toggleEventTraceNode(sheetId, address);
            }.bind(this));
        }
        return li;
    }

    attachEventTraceHandlers(sheetId) {
        var sheet = document.getElementById(sheetId);
        if (!sheet) return;
        var self = this;
        var filter = sheet.querySelector('.event-filter-btn');
        if (filter) filter.addEventListener('click', async function () {
            var pat = await self.entryDialog('Filter events', 'Pattern:', '', { okLabel: 'Filter' });
            if (pat !== null) await self.filterEventTrace(pat);
        });
        var findFwd = sheet.querySelector('.event-find-fwd-btn');
        if (findFwd) findFwd.addEventListener('click', async function () {
            var pat = await self.entryDialog('Find event', 'Pattern:', '', { okLabel: 'Find' });
            if (pat !== null) await self.findEventTrace(pat, false);
        });
        var findRev = sheet.querySelector('.event-find-rev-btn');
        if (findRev) findRev.addEventListener('click', async function () {
            var pat = await self.entryDialog('Find event', 'Pattern:', '', { okLabel: 'Find' });
            if (pat !== null) await self.findEventTrace(pat, true);
        });
        var patRev = sheet.querySelector('.event-pattern-rev');
        if (patRev) patRev.addEventListener('click', function () {
            var pat = self.selectedEventPattern(sheetId);
            if (pat) self.findEventTrace(pat, true);
        });
        var patFwd = sheet.querySelector('.event-pattern-fwd');
        if (patFwd) patFwd.addEventListener('click', function () {
            var pat = self.selectedEventPattern(sheetId);
            if (pat) self.findEventTrace(pat, false);
        });
        var patAdd = sheet.querySelector('.event-pattern-add');
        if (patAdd) patAdd.addEventListener('click', async function () {
            var pat = await self.entryDialog('Add pattern', 'Pattern:', '', { okLabel: 'Add' });
            if (pat !== null && pat !== '') await self.addEventPattern(sheetId, pat);
        });
        var patRemove = sheet.querySelector('.event-pattern-remove');
        if (patRemove) patRemove.addEventListener('click', function () {
            self.removeSelectedEventPattern(sheetId);
        });
        var patSave = sheet.querySelector('.event-pattern-save');
        if (patSave) patSave.addEventListener('click', function () {
            self.saveEventPatterns(sheetId);
        });
        var patLoad = sheet.querySelector('.event-pattern-load');
        if (patLoad) patLoad.addEventListener('click', async function () {
            var text = await self.textDialog('Load patterns', 'Paste patterns:', '', { okLabel: 'Load' });
            if (text !== null) await self.loadEventPatterns(sheetId, text);
        });
        var patClear = sheet.querySelector('.event-pattern-clear');
        if (patClear) patClear.addEventListener('click', function () {
            self.clearEventPatterns(sheetId);
        });
    }

    toggleEventTraceNode(sheetId, address) {
        return toggleEventTraceNodeViaService(this, sheetId, address);
    }

    lookupEventTrace(events, address) {
        return lookupEventTraceViaService(events, address);
        if (!address && address !== '0') return null;
        var parts = String(address).split('/');
        var current = null;
        var list = events || [];
        for (var i = 0; i < parts.length; i++) {
            var idx = Number(parts[i]);
            if (!Number.isInteger(idx) || idx < 0 || idx >= list.length) return null;
            current = list[idx];
            list = current.subs || [];
        }
        return current;
    }

    uncoverEventTraceAddress(sheetId, address) {
        return uncoverEventTraceAddressViaService(this, sheetId, address);
        var parts = String(address || '').split('/');
        var prefix = '';
        for (var i = 0; i < parts.length - 1; i++) {
            prefix = prefix === '' ? parts[i] : prefix + '/' + parts[i];
            if (!this.eventTraceRow(sheetId, prefix)) break;
            if (!this.eventTraceRow(sheetId, prefix + '/' + parts[i + 1])) {
                this.toggleEventTraceNode(sheetId, prefix);
            }
        }
    }

    selectEventTraceRow(sheetId, address) {
        return selectEventTraceRowViaService(this, sheetId, address);
    }

    activeEventSheet() {
        return activeEventSheetViaService(this);
        var sheet = this.sheets && this.sheets[this.activeSheetId];
        return sheet && sheet.type === 'events' ? sheet : null;
    }

    async filterEventTrace(pattern) {
        return filterEventTraceViaService(this, pattern);
        var sheet = this.activeEventSheet();
        if (!sheet) {
            this.controls.setStatus('No event sheet selected', 'error');
            return null;
        }
        if (sheet.visualOnly) {
            this.controls.setStatus(this.visualOnlyMessage('events'), 'warning');
            return null;
        }
        try {
            var result = await this.api.executeAction('events_filter', {
                sheet_id: sheet.id,
                pattern: pattern,
            });
            var label = (result && result.label) || 'Filtered events';
            this.openEventTraceSheet(label, result || {}, result && result.sheet_id);
            return result;
        } catch (e) {
            this.controls.setStatus('Filter failed: ' + e.message, 'error');
            return null;
        }
    }

    async findEventTrace(pattern, reverse) {
        return findEventTraceViaService(this, pattern, reverse);
        var sheet = this.activeEventSheet();
        if (!sheet) {
            this.controls.setStatus('No event sheet selected', 'error');
            return null;
        }
        if (sheet.visualOnly) {
            this.controls.setStatus(this.visualOnlyMessage('events'), 'warning');
            return null;
        }
        var args = {
            sheet_id: sheet.id,
            pattern: pattern,
            reverse: !!reverse,
            anchor: sheet.selectedEventAddress || '',
        };
        var result;
        try {
            result = await this.api.executeAction('events_find', args);
        } catch (e) {
            this.controls.setStatus('Find failed: ' + e.message, 'error');
            return null;
        }
        if (!result || !result.address) {
            this.controls.setStatus('Pattern not found', 'error');
            return result;
        }
        this.selectEventTraceRow(sheet.id, result.address);
        return result;
    }

    applyEventPatternResult(sheetId, result, fallbackPatterns = undefined) {
        return applyEventPatternResultViaService(this, sheetId, result, fallbackPatterns);
    }

    renderEventPatternList(sheetId) {
        return renderEventPatternListViaService(this, sheetId);
    }

    selectedEventPattern(sheetId) {
        return selectedEventPatternViaService(this, sheetId);
    }

    async addEventPattern(sheetId, pattern) {
        return addEventPatternViaService(this, sheetId, pattern);
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return;
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_add_pattern', { sheet_id: sheetId, pattern: pattern });
            this.applyEventPatternResult(sheetId, result);
            return;
        }
        sheet.patterns = (sheet.patterns || []).concat([pattern]);
        this.renderEventPatternList(sheetId);
    }

    async removeSelectedEventPattern(sheetId) {
        return removeSelectedEventPatternViaService(this, sheetId);
    }

    async clearEventPatterns(sheetId) {
        return clearEventPatternsViaService(this, sheetId);
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return;
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_clear_patterns', { sheet_id: sheetId });
            this.applyEventPatternResult(sheetId, result, []);
            return;
        }
        sheet.patterns = [];
        this.renderEventPatternList(sheetId);
    }

    async loadEventPatterns(sheetId, text) {
        return loadEventPatternsViaService(this, sheetId, text);
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return;
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_load_patterns', { sheet_id: sheetId, patterns: text });
            this.applyEventPatternResult(sheetId, result);
            return;
        }
        var patterns = String(text || '').split(/\r?\n/).map(function (line) { return line.trim(); }).filter(Boolean);
        sheet.patterns = sheet.patterns.concat(patterns);
        this.renderEventPatternList(sheetId);
    }

    async saveEventPatterns(sheetId) {
        return saveEventPatternsViaService(this, sheetId);
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return '';
        var content = (sheet.patterns || []).join('\n');
        if (content !== '') content += '\n';
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_save_patterns', { sheet_id: sheetId });
            if (result && typeof result.content === 'string') {
                content = result.content;
            }
        }
        this.downloadTextFile('event_patterns.pats', content, 'text/plain');
        return content;
    }

    /**
     * Remove a sheet tab and its content.
     * If the removed sheet was active, switch to Sheet 1.
     * Sheet 1 cannot be removed.
     */
    removeSheet(sheetId) {
        if (sheetId === 'sheet-1') return; // never remove Sheet 1

        var tab = this.sheetTab(sheetId);
        var sheet = document.getElementById(sheetId);
        var wasActive = this.activeSheetId === sheetId || (tab && tab.classList.contains('active'));

        if (tab) tab.remove();
        if (sheet) sheet.remove();
        if (this.sheets) {
            delete this.sheets[sheetId];
        }
        if (this.uiDataModel) {
            this.uiDataModel.removeSheet(sheetId);
        }

        // If the closed tab was active, switch to Sheet 1
        if (wasActive) {
            this.switchSheet('sheet-1');
        }
    }

    toggleTutorial(flash = false) {
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
        this._refreshEditorLayout();
    }

    _refreshGraphsAndEditorLayout() {
        refreshGraphsAndEditorLayout(this);
    }

    _refreshLayoutAfterPatch() {
        refreshLayoutAfterPatch(this);
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
            if (!self._setLayoutSize('setTutorialHeight', newHeight)) {
                tutorial.style.flex = '0 0 ' + newHeight + 'px';
            }
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

    setupDetailsResizer() {
        var sheetArea = document.getElementById('sheet-area');
        if (!sheetArea) return;
        var self = this;
        var isDragging = false;
        var startY = 0;
        var startHeight = 0;
        var activeHeader = null;
        var activePanel = null;
        var activeSheetLeft = null;
        var activeSheetMain = null;

        sheetArea.addEventListener('mousedown', function (e) {
            var header = e.target.closest('.info-header, #info-header');
            if (!header || !sheetArea.contains(header)) return;
            var panel = header.closest('.info-panel') || header.parentElement;
            var sheetLeft = panel ? panel.closest('.sheet-left') : null;
            if (!panel || !sheetLeft) return;

            isDragging = true;
            startY = e.clientY;
            startHeight = panel.offsetHeight;
            activeHeader = header;
            activePanel = panel;
            activeSheetLeft = sheetLeft;
            activeSheetMain = sheetLeft.querySelector('.sheet-main');
            header.classList.add('active');
            document.body.style.cursor = 'row-resize';
            document.body.style.userSelect = 'none';
            var graphs = sheetLeft.querySelectorAll('.graph-container');
            for (var i = 0; i < graphs.length; i++) graphs[i].style.pointerEvents = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging || !activePanel || !activeSheetLeft) return;
            var dy = startY - e.clientY;
            var minDetailsHeight = 72;
            var minMainHeight = self._detailsResizerMinimumMainHeight(activeSheetLeft);
            var maxHeight = Math.max(minDetailsHeight, activeSheetLeft.offsetHeight - minMainHeight);
            var newHeight = Math.max(minDetailsHeight, Math.min(startHeight + dy, maxHeight));
            if (activeSheetMain) {
                activeSheetMain.style.minHeight = minMainHeight + 'px';
            }
            if (!(activePanel.id === 'info-panel' && self._setLayoutSize('setDetailsHeight', newHeight))) {
                activePanel.style.flex = '0 0 ' + newHeight + 'px';
                activePanel.style.height = newHeight + 'px';
            }
            if (self.argGraph) self.argGraph.resize();
            if (self.conceptGraph) self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (!isDragging) return;
            isDragging = false;
            if (activeHeader) activeHeader.classList.remove('active');
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
            if (activeSheetLeft) {
                var graphs = activeSheetLeft.querySelectorAll('.graph-container');
                for (var i = 0; i < graphs.length; i++) graphs[i].style.pointerEvents = '';
            }
            if (self.argGraph) self.argGraph.resize();
            if (self.conceptGraph) self.conceptGraph.resize();
            activeHeader = null;
            activePanel = null;
            activeSheetLeft = null;
            activeSheetMain = null;
        });
    }

    _detailsResizerMinimumMainHeight(sheetLeft) {
        var fallback = 44;
        if (!sheetLeft) return fallback;
        var sheetMain = sheetLeft.querySelector('.sheet-main');
        if (!sheetMain) return fallback;
        var headers = sheetMain.querySelectorAll('.panel-header');
        var height = fallback;
        for (var i = 0; i < headers.length; i++) {
            height = Math.max(height, headers[i].offsetHeight || 0);
        }
        return height;
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

        function navigateTo(url, forceReload = false) {
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
        return populateStateCheckboxesViaService(this, conceptData);
    }

    populateConstraintFacts(conceptData) {
        return populateConstraintFactsViaService(this, conceptData);
    }

    _hydrateBackendToggleState(conceptData) {
        return hydrateBackendToggleState(this, conceptData);
    }

    _toggleChecked(name, displayClass) {
        return toggleChecked(this, name, displayClass);
    }

    async onEdgeToggle(edgeName, displayClass, checked) {
        return onEdgeToggleViaService(this, edgeName, displayClass, checked);
    }

    _applyEdgeVisibility(conceptGraph) {
        return applyEdgeVisibility(this, conceptGraph);
    }

    _findEdgeVisibility(obj, label) {
        return findEdgeVisibility(this, obj, label);
    }

    _applyNodeLabels() {
        return applyNodeLabels(this);
    }

    _displayConceptName(name) {
        return displayConceptName(name);
    }

    updateStateLabel(nodeId) {
        return updateStateLabelViaService(nodeId);
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
    async onArgNodeClick(nodeData, sheetId) {
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        var sheet = this.sheets && this.sheets[sheetId];
        var argGraph = (sheet && sheet.argGraph) || this.argGraph;
        var conceptGraph = (sheet && sheet.conceptGraph) || this.conceptGraph;
        if (sheetId !== this.activeSheetId && this.sheets && this.sheets[sheetId]) {
            this.switchSheet(sheetId);
            sheet = this.sheets[sheetId];
            argGraph = sheet.argGraph;
            conceptGraph = sheet.conceptGraph;
        }
        this.selectedArgNode = nodeData.id;
        if (sheet) {
            sheet.selectedArgNode = nodeData.id;
        }
        if (this.uiDataModel) {
            this.uiDataModel.setSelectedArgNode(sheetId, nodeData.id);
        }
        if (argGraph && typeof argGraph.highlightNode === 'function') {
            argGraph.highlightNode(nodeData.id);
        }
        this.controls.showInfo(nodeData.short_info, nodeData.long_info);
        this.updateStateLabel(nodeData.label || nodeData.id);
        if (sheet && sheet.visualOnly) {
            this.controls.setStatus(this.visualOnlyMessage('analysis'), 'warning');
            return;
        }
        this.controls.setStatus('Loading concept graph for state ' + (nodeData.label || nodeData.id) + '...');

        try {
            var result = await this.api.getConceptGraph(nodeData.obj || nodeData.id, sheetId);
            if (result && result.elements) {
                this.acceptConceptSnapshot(sheetId, result);
                conceptGraph.update(result.elements, result.positions);
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
    onArgNodeRightClick(nodeData, pos, sheetId) {
        var self = this;
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        var sheet = this.sheets && this.sheets[sheetId];
        if (sheet && sheet.visualOnly) {
            this.controls.setStatus(this.visualOnlyMessage('analysis'), 'warning');
            return;
        }
        var actions: any[] = [
            { header: 'State ' + (nodeData.label || nodeData.id) },
            {
                name: 'View State',
                id: 'view_state',
                callback: function () { self.onArgNodeClick(nodeData, sheetId); },
            },
            { separator: true },
        ];

        // Add server-provided actions if available
        if (nodeData.actions && Array.isArray(nodeData.actions)) {
            for (var i = 0; i < nodeData.actions.length; i++) {
                var act = nodeData.actions[i];
                var actionLabel = act.label || act[0] || act.name || '';
                var actionID = act.action || act.id || act[0] || '';
                if (actionLabel === '---') {
                    actions.push({ separator: true });
                    continue;
                }
                if (!actionID) {
                    actions.push({ header: actionLabel });
                    continue;
                }
                (function (action) {
                    actions.push({
                        name: action.label || action[0] || action.name,
                        id: action.action || action.id || action[0],
                        callback: function () {
                            self.executeArgNodeAction(nodeData, action, sheetId);
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
                            self.executeArgNodeAction(nodeData, act, sheetId);
                        },
                    });
                })(defaultActions[j]);
            }
        }

        // Offset from the graph container position
        var graphContainer = document.getElementById((sheet && sheet.argGraph && sheet.argGraph.containerId) || 'arg-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Execute an ARG node action via the API.
     */
    async executeArgNodeAction(nodeData, action, sheetId) {
        return executeArgNodeActionViaService(this, nodeData, action, sheetId);
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        if (this.isVisualOnlySheet(sheetId)) {
            this.controls.setStatus(this.visualOnlyMessage('analysis'), 'warning');
            return null;
        }
        var actionName = action.action || action.id || action[0] || action.name;
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            var args = Object.assign({}, action.args || {});
            args.sheet_id = sheetId;
            args = await this.prepareArgNodeActionArgs(nodeData, actionName, args, sheetId);
            if (args === null) {
                this.controls.setStatus('Action cancelled: ' + actionName, 'warning');
                return;
            }
            var result = await this.api.argNodeAction(nodeData.obj || nodeData.id, actionName, args);
            var sheet = this.sheets && this.sheets[sheetId];
            var argGraph = (sheet && sheet.argGraph) || this.argGraph;
            var conceptGraph = (sheet && sheet.conceptGraph) || this.conceptGraph;
            if (result && result.arg) {
                argGraph.update(result.arg.elements, result.arg.positions);
            }
            if (result && result.concept) {
                conceptGraph.update(result.concept.elements, result.concept.positions);
            }
            this.controls.setStatus('Action complete: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Action failed: ' + e.message, 'error');
            console.error('ARG action error:', e);
        }
    }

    async prepareArgNodeActionArgs(nodeData, actionName, args, sheetId) {
        return prepareArgNodeActionArgsViaService(this, nodeData, actionName, args, sheetId);
        if (actionName === 'try_conjecture' && !args.conjecture) {
            var conjChoices = await this.api.argNodeAction(nodeData.obj || nodeData.id, 'try_conjecture_choices', { sheet_id: sheetId });
            var selectedConj = await this.listboxDialog(
                'Try conjecture',
                'Choose a conjecture to prove:',
                conjChoices.choices || [],
                { cancel: true }
            );
            if (selectedConj == null) return null;
            args.conjecture = selectedConj;
        } else if (actionName === 'try_remembered' && !args.goal) {
            var goalChoices = await this.api.argNodeAction(nodeData.obj || nodeData.id, 'try_remembered_choices', { sheet_id: sheetId });
            var selectedGoal = await this.listboxDialog(
                'Try remembered goal',
                'Choose a remembered goal:',
                goalChoices.choices || [],
                { cancel: true }
            );
            if (selectedGoal == null) return null;
            args.goal = selectedGoal;
        }
        return args;
    }

    /**
     * Handle right-click on an ARG edge: show context menu.
     * Matches Python ivy_ui.py get_edge_actions: Dismiss, Recalculate, Step in, View Source.
     */
    onArgEdgeRightClick(edgeData, pos, sheetId) {
        var self = this;
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        var sheet = this.sheets && this.sheets[sheetId];
        var label = edgeData.label || edgeData.obj || '';
        var actions: any[] = [
            { header: 'Transition: ' + label },
            {
                name: 'Dismiss',
                id: 'dismiss',
                callback: function () { self.controls.hideContextMenu(); },
            },
            {
                name: 'Recalculate',
                id: 'recalculate_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'recalculate', sheetId); },
            },
            {
                name: 'Step in',
                id: 'decompose_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'decompose', sheetId); },
            },
            {
                name: 'View Source',
                id: 'view_source_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'view_source', sheetId); },
            },
        ];
        var graphContainer = document.getElementById((sheet && sheet.argGraph && sheet.argGraph.containerId) || 'arg-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Execute an ARG edge action via the API.
     */
    async executeArgEdgeAction(edgeData, actionName, sheetId) {
        return executeArgEdgeActionViaService(this, edgeData, actionName, sheetId);
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            var result = await this.api.argNodeAction(
                edgeData.source_obj || edgeData.source || edgeData.obj,
                actionName,
                { target: edgeData.target_obj || edgeData.target, sheet_id: sheetId }
            );
            if (actionName === 'decompose' && result && result.decomposed) {
                // Decompose: open a new tab with the sub-ARG
                var label = 'Step: ' + (edgeData.label || actionName);
                this.openARGSheet(label, result.sub_arg, result.sheet_id);
            }
            if (result && result.arg) {
                var sheet = this.sheets && this.sheets[sheetId];
                var argGraph = (sheet && sheet.argGraph) || this.argGraph;
                argGraph.update(result.arg.elements, result.arg.positions);
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
        var actions: any[] = [
            { header: nodeData.label || nodeData.obj || nodeData.id },
        ];

        // Server-provided actions (split by X, remove, empty, materialize, add projection)
        if (nodeData.actions && Array.isArray(nodeData.actions)) {
            for (var i = 0; i < nodeData.actions.length; i++) {
                var act = nodeData.actions[i];
                var label = act.label || act[0] || act.name || '';
                var id = act.action || act.id || act[0] || '';
                if (label === '---') {
                    actions.push({ separator: true });
                    continue;
                }
                if (!id) {
                    actions.push({ header: label });
                    continue;
                }
                (function (action) {
                    var actionName = action.label || action[0] || action.name;
                    actions.push({
                        name: actionName,
                        id: action.action || action.id || action[0] || actionName,
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
                name: 'Materialize edge',
                id: 'materialize_from_selected',
                callback: function () {
                    self.materializeEdgeFromSelected(nodeData.obj || nodeData.id);
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
        var actions: any[] = [
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
                callback: function () { self.materializeEdge(edgeData, true); },
            });
            actions.push({
                name: 'Materialize \u2013',
                id: 'materialize_neg',
                callback: function () { self.materializeEdge(edgeData, false); },
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
        return executeConceptNodeActionViaService(this, nodeData, action);
        var actionID = action.action || action.id || action[0] || action.name || '';
        var actionName = actionID.toLowerCase();
        var actionArgs = action.args || {};

        if (actionName === 'remove') {
            return this.removeConcept(nodeData.obj || nodeData.id);
        }
        if (actionName === 'suppose empty') {
            return this.supposeEmpty(nodeData.obj || nodeData.id);
        }
        if (actionName === 'materialize') {
            return this.materializeNode(nodeData.obj || nodeData.id);
        }
        if (actionName === 'materialize edge' || actionName === 'materialize_from_selected') {
            return this.materializeEdgeFromSelected(nodeData.obj || nodeData.id);
        }
        if (actionName.indexOf('split by ') === 0) {
            var splitBy = actionName.substring('split by '.length);
            return this.splitConcept(nodeData.obj || nodeData.id, splitBy);
        }
        if (actionName.indexOf('add ') === 0) {
            var projName = actionName.substring('add '.length);
            return this.addProjection(projName, nodeData.obj || nodeData.id);
        }
        if (actionName === 'add_projection') {
            return this.addProjection(actionArgs.name, actionArgs.concept || actionArgs.name);
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
        return executeConceptEdgeActionViaService(this, edgeData, action);
        var actionName = (action[0] || action.name || '').toLowerCase();
        var conceptId = edgeData.obj || edgeData.id;

        if (actionName === 'remove') {
            return this.removeConcept(conceptId);
        }
        if (actionName === 'materialize +') {
            return this.materializeEdge(edgeData, true);
        }
        if (actionName === 'materialize -' || actionName === 'materialize \u2013') {
            return this.materializeEdge(edgeData, false);
        }
        if (actionName === 'dematerialize') {
            return this.materializeEdge(edgeData, false);
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
        return splitConceptViaService(this, concept, splitBy);
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
        return supposeEmptyViaService(this, concept);
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
        return removeConceptViaService(this, concept);
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
        return materializeNodeViaService(this, concept);
        this.controls.setStatus('Materializing node...');
        try {
            await this.api.materializeNode(concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Node materialized', 'success');
        } catch (e) {
            this.controls.setStatus('Materialize failed: ' + e.message, 'error');
        }
    }

    async materializeEdge(edgeOrConcept, positive) {
        return materializeEdgeViaService(this, edgeOrConcept, positive);
        var dir = positive ? '+' : '\u2013';
        var relation = edgeOrConcept;
        var source = '';
        var target = '';
        if (edgeOrConcept && typeof edgeOrConcept === 'object') {
            relation = edgeOrConcept.obj || edgeOrConcept.label || edgeOrConcept.id;
            source = edgeOrConcept.source_obj || edgeOrConcept.source || '';
            target = edgeOrConcept.target_obj || edgeOrConcept.target || '';
        }
        this.controls.setStatus('Materializing edge (' + dir + ')...');
        try {
            await this.api.materializeEdge(relation, source, target, positive);
            await this.refreshConceptGraph();
            this.controls.setStatus('Edge materialized (' + dir + ')', 'success');
        } catch (e) {
            this.controls.setStatus('Materialize failed: ' + e.message, 'error');
        }
    }

    async addProjection(name, concept) {
        return addProjectionViaService(this, name, concept);
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
        return selectConceptNodeViaService(this, conceptId);
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

    async materializeEdgeFromSelected(targetConceptId) {
        return materializeEdgeFromSelectedViaService(this, targetConceptId);
        var sourceConceptId = this.selectedConceptNode;
        if (!sourceConceptId) {
            this.controls.setStatus('Select a source node first', 'warning');
            return;
        }
        var data = this._lastConceptData || {};
        var edgeSorts = data.edge_sorts || {};
        var relations = Object.keys(edgeSorts).filter(function (rel) {
            var sorts = edgeSorts[rel] || [];
            return sorts.length >= 2 && sorts[0] === sourceConceptId && sorts[1] === targetConceptId;
        });
        if (relations.length === 0 && Array.isArray(data.edges)) {
            relations = data.edges.slice();
        }
        if (relations.length === 0) {
            this.controls.setStatus('No matching binary relations', 'warning');
            return;
        }
        var selected = await this.listboxDialog(
            'Materialize edge',
            'Materialize this relation from selected node:',
            relations.map(function (rel) { return { label: rel, value: rel }; }),
            { cancel: true }
        );
        if (selected == null) {
            this.controls.setStatus('Materialize edge cancelled', 'warning');
            return;
        }
        this.controls.setStatus('Materializing edge ' + selected + '...');
        try {
            await this.api.materializeEdge(selected, sourceConceptId, targetConceptId, true);
            await this.refreshConceptGraph();
            this.controls.setStatus('Edge materialized (+)', 'success');
        } catch (e) {
            this.controls.setStatus('Materialize edge failed: ' + e.message, 'error');
        }
    }

    /**
     * Splatter a concept node — materialize all universe elements of its sort.
     * Matches Python tk_graph_ui.py splatter action.
     */
    async splatterNode(conceptId) {
        return splatterNodeViaService(this, conceptId);
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
        return onEdgeToggleChangeViaService(this, edgeName, className, checked);
        try {
            var toggles = {
                edge: edgeName,
                display_class: className,
                value: checked,
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
        return onLabelToggleChangeViaService(this, labelName, className, checked);
        try {
            var toggles = {
                label: labelName,
                display_class: className,
                value: checked,
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
        return loadModelFile(this, file, runtimeDeps.IvyPersist);
    }

    /**
     * Save invariant/conjectures to a .ivy file.
     * Matches Python ivy_ui_cti.py save_conjectures:
     * exports conjectures as "invariant [label] formula" lines.
     */
    async saveInvariant() {
        this.controls.setStatus('Saving invariant...');
        try {
            var result = await this.api.executeAction('save_invariant', {});
            var text = (result && result.content) || '';
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
        return downloadModelViaService(this);
    }

    downloadTextFile(filename, content, mimeType) {
        return downloadTextFileViaService(filename, content, mimeType);
    }

    /**
     * Save as... — uses the File System Access API (showSaveFilePicker)
     * to let the user choose a disk path. Remembers the file handle
     * for subsequent saves.
     */
    async save() {
        return saveModel(this, runtimeDeps.IvyPersist);
    }

    downloadModelForUnsupportedSave(content) {
        return downloadModelForUnsupportedSaveViaService(this, content, runtimeDeps.IvyPersist);
    }

    async saveAs(options = {}) {
        return saveModelAs(this, runtimeDeps.IvyPersist, { options: options || {} });
    }

    // Keep saveSession as an alias for downloadModel (used by ARG panel binding)
    async saveSession() { return this.downloadModel(); }

    graphElementsSnapshot(graph) {
        return graphElementsSnapshotViaService(graph);
    }

    getEditorKeymap() {
        return getEditorKeymapViaService();
    }

    setEditorKeymap(keymap) {
        setEditorKeymapViaService(this, keymap);
    }

    tabLabelForSheet(sheetId) {
        var tab = this.sheetTab(sheetId);
        var label = tab ? tab.querySelector('span') : null;
        return label ? label.textContent : sheetId;
    }

    getMode() {
        var modeEl = document.getElementById('mode-select');
        return modeEl ? modeEl.value : 'pdr';
    }

    setMode(mode) {
        var modeEl = document.getElementById('mode-select');
        if (modeEl && mode) modeEl.value = mode;
    }

    buildAnalysisState() {
        return buildAnalysisStateViaService(this, runtimeDeps.IvyPersist);
        var sheets = [];
        var ids = Object.keys(this.sheets || {});
        ids.sort(function (a, b) {
            if (a === 'sheet-1') return -1;
            if (b === 'sheet-1') return 1;
            return a.localeCompare(b);
        });
        for (var i = 0; i < ids.length; i++) {
            var sheetId = ids[i];
            var sheet = this.sheets[sheetId];
            if (!sheet) continue;
            if (sheet.type === 'events') {
                sheets.push({
                    id: sheetId,
                    type: 'events',
                    label: this.tabLabelForSheet(sheetId),
                    events: sheet.events || [],
                    patterns: sheet.patterns || [],
                    selectedEventAddress: sheet.selectedEventAddress || null,
                });
            } else {
                sheets.push({
                    id: sheetId,
                    type: 'analysis',
                    label: this.tabLabelForSheet(sheetId),
                    selectedArgNode: sheet.selectedArgNode || null,
                    arg: { elements: this.graphElementsSnapshot(sheet.argGraph), positions: null },
                    concept: { elements: this.graphElementsSnapshot(sheet.conceptGraph), positions: null },
                });
            }
        }
        return {
            analysis_state_format: 'ivyweb-json',
            analysis_state_version: 1,
            python_a2g_equivalent: false,
            fileName: this._persistedFileName || '',
            filePath: this._persistedFilePath || this._persistedFileName || '',
            fileContent: this._editorContent ? this._editorContent() : (this._persistedFileContent || ''),
            mode: this.getMode(),
            activeSheetId: this.activeSheetId || 'sheet-1',
            selectedArgNode: this.selectedArgNode || null,
            edgeVisibility: this._edgeVisibility || {},
            labelVisibility: this._labelVisibility || {},
            toggles: runtimeDeps.IvyPersist._getToggles ? runtimeDeps.IvyPersist._getToggles() : {},
            sheets: sheets,
        };
    }

    async saveAnalysisState() {
        return saveAnalysisStateViaService(this);
        try {
            var state = this.buildAnalysisState();
            var text = JSON.stringify(state, null, 2) + '\n';
            var suggestedName = (this._persistedFileName || 'ivy_analysis').replace(/\.ivy$/, '') + '.ivyweb.json';
            if (window.showSaveFilePicker) {
                var handle = await window.showSaveFilePicker({
                    suggestedName: suggestedName,
                    types: [{ description: 'IvyWeb analysis state', accept: { 'application/json': ['.json'] } }],
                });
                var writable = await handle.createWritable();
                await writable.write(text);
                await writable.close();
                this.controls.setStatus('Analysis state saved: ' + handle.name, 'success');
            } else {
                this.downloadTextFile(suggestedName, text, 'application/json');
                this.controls.setStatus('Analysis state downloaded: ' + suggestedName, 'success');
            }
            return state;
        } catch (e) {
            if (e.name === 'AbortError') {
                this.controls.setStatus('Save analysis state cancelled');
            } else {
                this.controls.setStatus('Save analysis state failed: ' + e.message, 'error');
            }
            return null;
        }
    }

    async loadAnalysisStateFile(file) {
        return loadAnalysisStateFileViaService(this, file);
        if (!file) return false;
        if (typeof file.size === 'number' && file.size > this.analysisStateLimits().maxFileBytes) {
            throw new Error('analysis state file too large');
        }
        var text = await this.readFileText(file);
        return this.loadAnalysisStateObject(JSON.parse(text));
    }

    async loadAnalysisStateObject(state) {
        return loadAnalysisStateObjectViaService(this, state, runtimeDeps.IvyPersist);
        this.validateAnalysisStateObject(state);
        if (!state || state.analysis_state_format !== 'ivyweb-json') {
            throw new Error('unsupported analysis state format');
        }
        this._persistedFileName = state.fileName || '';
        this._persistedFilePath = state.filePath || state.fileName || '';
        this._persistedFileContent = state.fileContent || '';
        this._savedFileContent = this._persistedFileContent;
        this._edgeVisibility = state.edgeVisibility || {};
        this._labelVisibility = state.labelVisibility || {};
        this.selectedArgNode = state.selectedArgNode || null;

        if (this.setEditorContent) {
            this.setEditorContent(this._persistedFileContent);
        }
        if (state.mode) this.setMode(state.mode);
        if (this.api && this.api.reloadContent && this._persistedFileContent) {
            await this.api.reloadContent(this._persistedFileContent, this._persistedFileName || 'restored.ivy');
        }

        this.removeAnalysisStateExtraSheets();
        var sheets = state.sheets || [];
        for (var i = 0; i < sheets.length; i++) {
            var sheet = sheets[i];
            if (sheet.type === 'events') {
                this.openEventTraceSheet(sheet.label || 'Events', {
                    sheet_id: sheet.id,
                    events: sheet.events || [],
                    patterns: sheet.patterns || [],
                    selected_address: sheet.selectedEventAddress || null,
                }, sheet.id);
                this.setVisualOnlySheet(sheet.id, true);
                continue;
            }
            if (sheet.id === 'sheet-1') {
                if (sheet.arg && sheet.arg.elements && this.argGraph) {
                    this.argGraph.update(sheet.arg.elements, sheet.arg.positions || undefined);
                }
                if (sheet.concept && sheet.concept.elements && this.conceptGraph) {
                    this.conceptGraph.update(sheet.concept.elements, sheet.concept.positions || undefined);
                }
                if (this.sheets && this.sheets['sheet-1']) {
                    this.sheets['sheet-1'].selectedArgNode = sheet.selectedArgNode || null;
                    this.sheets['sheet-1'].visualOnly = true;
                }
            } else if (sheet.arg && sheet.arg.elements) {
                this.openARGSheet(sheet.label || sheet.id, sheet.arg, sheet.id);
                var opened = this.sheets && this.sheets[sheet.id];
                if (opened && opened.conceptGraph && sheet.concept && sheet.concept.elements) {
                    opened.conceptGraph.update(sheet.concept.elements, sheet.concept.positions || undefined);
                }
                if (opened) {
                    opened.selectedArgNode = sheet.selectedArgNode || null;
                    opened.visualOnly = true;
                }
            }
        }

        if (state.toggles) {
            runtimeDeps.IvyPersist._setToggles(state.toggles);
        }
        if (state.activeSheetId && this.sheetExists(state.activeSheetId)) {
            this.switchSheet(state.activeSheetId);
        } else {
            this.switchSheet('sheet-1');
        }
        runtimeDeps.IvyPersist.setFileName(this._persistedFileName, this._persistedFilePath);
        this.controls.setStatus('Visual analysis state loaded: ' + (this._persistedFileName || 'state'), 'warning');
        return true;
    }

    analysisStateLimits() {
        return analysisStateLimitsViaService();
        return {
            maxFileBytes: 25 * 1024 * 1024,
            maxSheets: 100,
            maxGraphElements: 50000,
            maxEvents: 100000,
            maxEventDepth: 200,
        };
    }

    validateAnalysisStateObject(state) {
        return validateAnalysisStateObjectViaService(this, state);
        if (!state || state.analysis_state_format !== 'ivyweb-json') {
            throw new Error('unsupported analysis state format');
        }
        var limits = this.analysisStateLimits();
        var sheets = state.sheets || [];
        if (!Array.isArray(sheets)) {
            throw new Error('analysis state sheets must be an array');
        }
        if (sheets.length > limits.maxSheets) {
            throw new Error('too many sheets in analysis state');
        }
        if (state.activeSheetId && !this.isValidSheetId(state.activeSheetId)) {
            throw new Error('invalid active sheet id: ' + state.activeSheetId);
        }
        var seen = {};
        for (var i = 0; i < sheets.length; i++) {
            this.validateAnalysisStateSheet(sheets[i], seen, limits);
        }
    }

    validateAnalysisStateSheet(sheet, seen, limits) {
        return validateAnalysisStateSheetViaService(sheet, seen, limits, this.isValidSheetId.bind(this));
        if (!sheet || typeof sheet !== 'object') {
            throw new Error('invalid sheet entry');
        }
        if (!this.isValidSheetId(sheet.id)) {
            throw new Error('invalid sheet id: ' + sheet.id);
        }
        if (seen[sheet.id]) {
            throw new Error('duplicate sheet id: ' + sheet.id);
        }
        seen[sheet.id] = true;
        if (sheet.type !== 'analysis' && sheet.type !== 'events') {
            throw new Error('invalid sheet type: ' + sheet.type);
        }
        if (sheet.type === 'events') {
            if (sheet.events != null && !Array.isArray(sheet.events)) {
                throw new Error('event sheet events must be an array');
            }
            if (sheet.patterns != null && !Array.isArray(sheet.patterns)) {
                throw new Error('event sheet patterns must be an array');
            }
            var count = { events: 0 };
            this.validateAnalysisStateEvents(sheet.events || [], 0, count, limits);
            return;
        }
        this.validateAnalysisStateGraphPayload(sheet.arg, 'arg', limits);
        this.validateAnalysisStateGraphPayload(sheet.concept, 'concept', limits);
    }

    validateAnalysisStateGraphPayload(graph, name, limits) {
        return validateAnalysisStateGraphPayloadViaService(graph, name, limits);
        if (!graph) return;
        if (graph.elements != null && !Array.isArray(graph.elements)) {
            throw new Error(name + ' graph elements must be an array');
        }
        if (graph.elements && graph.elements.length > limits.maxGraphElements) {
            throw new Error(name + ' graph has too many elements');
        }
    }

    validateAnalysisStateEvents(events, depth, count, limits) {
        return validateAnalysisStateEventsViaService(events, depth, count, limits);
        if (!Array.isArray(events)) {
            throw new Error('event children must be an array');
        }
        if (depth > limits.maxEventDepth) {
            throw new Error('event tree too deep');
        }
        for (var i = 0; i < events.length; i++) {
            count.events++;
            if (count.events > limits.maxEvents) {
                throw new Error('too many events in analysis state');
            }
            var ev = events[i] || {};
            if (ev.address != null && !/^\d+(\/\d+)*$/.test(String(ev.address))) {
                throw new Error('invalid event address: ' + ev.address);
            }
            if (ev.subs != null) {
                this.validateAnalysisStateEvents(ev.subs, depth + 1, count, limits);
            }
        }
    }

    removeAnalysisStateExtraSheets() {
        return removeAnalysisStateExtraSheetsViaService(this);
        var ids = Object.keys(this.sheets || {});
        for (var i = 0; i < ids.length; i++) {
            if (ids[i] !== 'sheet-1') {
                this.removeSheet(ids[i]);
            }
        }
    }

    async closeCurrentFile() {
        return closeCurrentFileViaService(this);
    }

    /**
     * Start a new model. Saves any existing state first, then clears everything.
     */
    async newModel(options = {}) {
        return newModelViaService(this, runtimeDeps.IvyPersist, { options: options || {} });
    }

    /**
     * Populate the recent files list in the File dropdown.
     * Deduplicates by fileName+contentLength, keeping the most recent.
     * Shows truncated path context when file names collide.
     */
    populateRecentFiles() {
        populateRecentFilesViaService(this, runtimeDeps.IvyPersist);
    }

    /**
     * Load a recent session by its saved session ID.
     */
    async loadRecentSession(savedSessionId) {
        return loadRecentSessionViaService(this, runtimeDeps.IvyPersist, savedSessionId);
    }

    /**
     * Run a verification check in the currently selected mode.
     */
    async runCheck() {
        return runCheckViaService(this);
        var mode = this.getMode();
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
                await this._autoCheckUsedRelations(result.used_relations);
            }
            this.showCheckResult(result);
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
    async _autoCheckUsedRelations(relationNames) {
        return autoCheckUsedRelations(this, relationNames);
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
                    await this.onEdgeToggle(name, 'all_to_all', true);
                }
            }
        }
    }

    /**
     * Display the result of a verification check.
     */
    showCheckResult(result) {
        return showCheckResultViaService(this, result);
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
            var failDetails = result.message || 'Counterexample found.';
            if (result.failed_conjecture) {
                failDetails += '\n\n' + result.failed_conjecture;
            }
            if (result.counterexample_details) {
                failDetails += '\n\n' + result.counterexample_details;
            } else if (result.counterexample_trace) {
                failDetails += '\n\nCounterexample trace:\n' + result.counterexample_trace;
            }
            this.controls.setStatus('Check FAILED' + mode + z3note + ' - counterexample found', 'error');
            this.controls.showInfo('Verification Result', 'FAILED' + z3note + ': ' + failDetails);
            this.addCheckResultViewActions(result);
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

    addCheckResultViewActions(result) {
        return addCheckResultViewActionsViaService(this, result);
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
            var result = await this.api.diagramDomain();
            if (result && result.concept) {
                this.updateConceptGraph(result.concept);
            } else {
                await this.refreshConceptGraph();
            }
            this.controls.setStatus('Diagram domain active', 'success');
        } catch (e) {
            this.controls.setStatus('Diagram domain failed: ' + e.message, 'error');
        }
    }

    /**
     * Refresh the concept graph from the server.
     */
    async refreshConceptGraph() {
        return refreshConceptGraphViaService(this);
        try {
            var result = await this.api.getConceptGraph(this.selectedArgNode);
            if (result && result.elements) {
                this.conceptGraph.update(result.elements, result.positions);
            }
            if (result) {
                this.populateStateCheckboxes(result);
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
                    if (event.data.status && event.data.status !== 'ok') {
                        this.controls.setStatus('Action failed: ' + event.data.status, 'error');
                    } else {
                        this.controls.setStatus('Done: ' + event.data.action, 'success');
                    }
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
        return closeAllDropdownsViaService();
        var all = document.querySelectorAll('.dropdown.open');
        for (var j = 0; j < all.length; j++) {
            all[j].classList.remove('open');
        }
    }

    /**
     * Flash a menu item (macOS Cocoa style invert) then close dropdowns and invoke callback.
     */
    flashAndClose(el, callback) {
        return flashAndCloseViaService(this, el, callback);
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

    async loadMenuDescriptors() {
        try {
            var menus = await this.api.getMenus();
            this.renderMenuRegion('arg', menus.arg || []);
            this.renderMenuRegion('concept', menus.concept || []);
        } catch (e) {
            this.controls.setStatus('Menu load failed: ' + e.message, 'error');
        }
    }

    renderMenuRegion(region, menus) {
        var panel = region === 'arg' ? document.getElementById('arg-panel') : document.getElementById('concept-panel');
        if (!panel) return;
        var header = panel.querySelector('.panel-header');
        if (!header) return;
        var old = header.querySelector('[data-dynamic-menu-region="' + region + '"]');
        if (old) old.remove();
        var menuRow = header.querySelector('.panel-header-actions') || header;

        var root = document.createElement('div');
        root.className = 'dynamic-menu-root';
        root.setAttribute('data-dynamic-menu-region', region);
        menuRow.appendChild(root);

        for (var i = 0; i < menus.length; i++) {
            this.renderMenuDescriptor(root, region, menus[i], i);
        }
    }

    renderMenuDescriptor(root, region, menu, index) {
        var self = this;
        var dropdown = document.createElement('div');
        dropdown.className = 'dropdown';

        var contentId = 'dynamic-' + region + '-' + index + '-' + (menu.label || 'menu').toLowerCase().replace(/[^a-z0-9]+/g, '-');
        var label = document.createElement('span');
        label.className = 'panel-menu';
        label.textContent = menu.label || 'Menu';
        label.setAttribute('data-dropdown', contentId);
        label.addEventListener('click', function (e) {
            e.preventDefault();
            e.stopPropagation();
            var wasOpen = dropdown.classList.contains('open');
            self.closeAllDropdowns();
            if (!wasOpen) dropdown.classList.add('open');
        });
        dropdown.appendChild(label);

        var content = document.createElement('div');
        content.id = contentId;
        content.className = 'dropdown-content';
        dropdown.appendChild(content);

        var items = menu.items || [];
        for (var i = 0; i < items.length; i++) {
            var item = items[i];
            if (item.type === 'separator') {
                var sep = document.createElement('div');
                sep.className = 'dropdown-sep';
                content.appendChild(sep);
                continue;
            }
            var link = document.createElement('a');
            link.href = '#';
            link.textContent = item.label || item.action || '';
            link.setAttribute('data-menu-action', item.action || '');
            link.setAttribute('data-menu-dispatch', item.dispatch || '');
            if (item.enabled === false) {
                link.classList.add('disabled');
                link.setAttribute('aria-disabled', 'true');
            }
            link.addEventListener('click', function (e) {
                e.preventDefault();
                e.stopPropagation();
                var actionItem = this.__ivyMenuItem;
                self.closeAllDropdowns();
                self.dispatchMenuDescriptorAction(region, actionItem);
            });
            link.__ivyMenuItem = item;
            content.appendChild(link);
        }

        root.appendChild(dropdown);
    }

    dispatchMenuDescriptorAction(region, item) {
        return dispatchMenuDescriptorActionViaService(this, region, item);
        if (!item || item.enabled === false) return Promise.resolve({ ok: false, error: 'disabled action' });
        if (item.dispatch === 'action') {
            return this.runAction(item.action, {}, {
                runningMessage: 'Running: ' + item.action + '...',
                successMessage: 'Done: ' + item.action,
            });
        }
        return this.runAction(item.action, {});
    }

    /**
     * Shared browser action runner, matching Python's run_context shape:
     * show progress, surface backend errors, and return a structured outcome.
     */
    async runAction(actionName, args, options = undefined) {
        return runActionViaService(this, actionName, args, options);
        var opts = options || {};
        var runningMessage = opts.runningMessage || ('Running: ' + actionName + '...');
        var successMessage = opts.successMessage || ('Done: ' + actionName);
        var failurePrefix = opts.failurePrefix || 'Action failed';
        if (opts.showLoading !== false) {
            this.controls.showLoading(runningMessage);
        }
        this.controls.setStatus(runningMessage, 'info');
        try {
            var result = await this.api.executeAction(actionName, args || {});
            if (opts.successStatus !== false) {
                this.controls.setStatus(successMessage, 'success');
            }
            return { ok: true, result: result };
        } catch (e) {
            var message = failurePrefix + ': ' + e.message;
            this.controls.setStatus(message, 'error');
            if (opts.showDialog) {
                this.showTextDialog('ivyweb', failurePrefix, e.message);
            }
            return { ok: false, error: e.message };
        } finally {
            if (opts.showLoading !== false) {
                this.controls.hideLoading();
            }
        }
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
        return checkInductionViaService(this);
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

    _createDialog(title, message) {
        var overlay = document.createElement('div');
        overlay.className = 'dialog-overlay';
        overlay.setAttribute('data-ivy-dialog', 'true');
        overlay.style.display = 'flex';

        var box = document.createElement('div');
        box.className = 'dialog-box';
        overlay.appendChild(box);

        var titleEl = document.createElement('div');
        titleEl.className = 'dialog-title';
        titleEl.textContent = title || 'ivyweb';
        box.appendChild(titleEl);

        var messageEl = document.createElement('div');
        messageEl.className = 'dialog-message';
        messageEl.textContent = message || '';
        box.appendChild(messageEl);

        var body = document.createElement('div');
        body.className = 'dialog-body';
        box.appendChild(body);

        var error = document.createElement('div');
        error.className = 'dialog-message dialog-error';
        error.setAttribute('data-ivy-dialog-error', 'true');
        error.style.display = 'none';
        box.appendChild(error);

        var buttons = document.createElement('div');
        buttons.className = 'dialog-buttons';
        box.appendChild(buttons);

        document.body.appendChild(overlay);
        return { overlay: overlay, body: body, buttons: buttons, error: error };
    }

    _setDialogError(dialog, message) {
        dialog.error.textContent = message || '';
        dialog.error.style.display = message ? 'block' : 'none';
    }

    _addDialogButton(dialog, label, callback, extraClass = '') {
        var btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'dialog-btn' + (extraClass ? ' ' + extraClass : '');
        btn.textContent = label || 'OK';
        btn.setAttribute('data-ivy-dialog-button', 'true');
        btn.addEventListener('click', callback);
        dialog.buttons.appendChild(btn);
        return btn;
    }

    _finishDialog(dialog, cleanup, resolve, value = null) {
        if (cleanup) cleanup();
        dialog.overlay.remove();
        resolve(value);
    }

    _installDialogEscape(dialog, finish) {
        var handler = function (e) {
            if (e.key === 'Escape') {
                finish();
            }
        };
        document.addEventListener('keydown', handler);
        return function () {
            document.removeEventListener('keydown', handler);
        };
    }

    okDialog(title, message) {
        var self = this;
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, true);
            });
            self._addDialogButton(dialog, 'OK', function () {
                self._finishDialog(dialog, cleanup, resolve, true);
            });
        });
    }

    okCancelDialog(title, message) {
        var self = this;
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, false);
            });
            self._addDialogButton(dialog, 'Cancel', function () {
                self._finishDialog(dialog, cleanup, resolve, false);
            });
            self._addDialogButton(dialog, 'OK', function () {
                self._finishDialog(dialog, cleanup, resolve, true);
            });
        });
    }

    textDialog(title, message, text, options) {
        var self = this;
        var opts = options || {};
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var textarea = document.createElement('textarea');
            textarea.className = 'dialog-text';
            textarea.rows = opts.rows || 4;
            textarea.cols = opts.cols || 80;
            textarea.value = text || '';
            textarea.readOnly = !!opts.readOnly;
            textarea.setAttribute('data-ivy-dialog-text', 'true');
            dialog.body.appendChild(textarea);

            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, null);
            });
            if (opts.cancel) {
                self._addDialogButton(dialog, 'Cancel', function () {
                    self._finishDialog(dialog, cleanup, resolve, null);
                });
            }
            self._addDialogButton(dialog, opts.okLabel || 'OK', function () {
                self._finishDialog(dialog, cleanup, resolve, textarea.value);
            });
            textarea.focus();
            textarea.select();
        });
    }

    /**
     * Backward-compatible display-only text dialog.
     */
    showTextDialog(title, message, text) {
        this.textDialog(title, message, text, { readOnly: false, okLabel: 'OK' });
    }

    entryDialog(title, message, initialValue, options) {
        var self = this;
        var opts = options || {};
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var input = document.createElement('input');
            input.type = 'text';
            input.className = 'dialog-input';
            input.value = initialValue || '';
            input.setAttribute('data-ivy-dialog-entry', 'true');
            dialog.body.appendChild(input);

            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, null);
            });
            if (opts.cancel !== false) {
                self._addDialogButton(dialog, 'Cancel', function () {
                    self._finishDialog(dialog, cleanup, resolve, null);
                });
            }
            self._addDialogButton(dialog, opts.okLabel || 'OK', function () {
                self._finishDialog(dialog, cleanup, resolve, input.value);
            });
            input.focus();
            input.select();
        });
    }

    integerDialog(title, message, initialValue, options) {
        var self = this;
        var opts = options || {};
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var input = document.createElement('input');
            input.type = 'number';
            input.className = 'dialog-input';
            input.value = String(initialValue == null ? '' : initialValue);
            if (opts.min != null) input.min = String(opts.min);
            if (opts.max != null) input.max = String(opts.max);
            input.setAttribute('data-ivy-dialog-int', 'true');
            dialog.body.appendChild(input);

            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, null);
            });
            if (opts.cancel !== false) {
                self._addDialogButton(dialog, 'Cancel', function () {
                    self._finishDialog(dialog, cleanup, resolve, null);
                });
            }
            self._addDialogButton(dialog, opts.okLabel || 'OK', function () {
                var raw = input.value.trim();
                var value = Number(raw);
                if (raw === '' || !Number.isInteger(value)) {
                    self._setDialogError(dialog, 'Enter an integer.');
                    return;
                }
                if (opts.min != null && value < opts.min) {
                    self._setDialogError(dialog, 'Enter a value at least ' + opts.min + '.');
                    return;
                }
                if (opts.max != null && value > opts.max) {
                    self._setDialogError(dialog, 'Enter a value at most ' + opts.max + '.');
                    return;
                }
                self._finishDialog(dialog, cleanup, resolve, value);
            });
            input.focus();
            input.select();
        });
    }

    listboxDialog(title, message, items, options) {
        var self = this;
        var opts = options || {};
        var entries = (items || []).map(function (item) {
            if (typeof item === 'object' && item !== null) {
                return { label: item.label || String(item.value), value: item.value };
            }
            return { label: String(item), value: item };
        });
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var select = document.createElement('select');
            select.className = 'dialog-input dialog-listbox';
            select.size = opts.size || Math.min(Math.max(entries.length, 2), 12);
            select.multiple = !!opts.multiple;
            select.setAttribute('data-ivy-dialog-list', 'true');
            entries.forEach(function (entry, index) {
                var opt = document.createElement('option');
                opt.value = String(entry.value);
                opt.setAttribute('data-ivy-dialog-index', String(index));
                opt.textContent = entry.label;
                select.appendChild(opt);
            });
            dialog.body.appendChild(select);

            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, opts.multiple ? [] : null);
            });
            if (opts.cancel !== false) {
                self._addDialogButton(dialog, 'Cancel', function () {
                    self._finishDialog(dialog, cleanup, resolve, opts.multiple ? [] : null);
                });
            }
            self._addDialogButton(dialog, opts.okLabel || 'OK', function () {
                if (opts.multiple) {
                    var selected = Array.from(select.selectedOptions).map(function (opt) {
                        return entries[Number(opt.getAttribute('data-ivy-dialog-index'))].value;
                    });
                    self._finishDialog(dialog, cleanup, resolve, selected);
                    return;
                }
                var selectedOption = select.selectedOptions[0];
                var idx = selectedOption ? Number(selectedOption.getAttribute('data-ivy-dialog-index')) : -1;
                self._finishDialog(dialog, cleanup, resolve, idx >= 0 ? entries[idx].value : null);
            });
            select.focus();
        });
    }

    buttonListDialog(title, message, buttons) {
        var self = this;
        var entries = buttons || [];
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, null);
            });
            entries.forEach(function (entry) {
                var label = entry.label || String(entry.value);
                self._addDialogButton(dialog, label, function () {
                    self._finishDialog(dialog, cleanup, resolve, entry.value);
                }, entry.danger ? 'dialog-btn-danger' : '');
            });
        });
    }

    /**
     * Run bounded model checking.
     * Matches Python ivy_ui_cti.py bounded_check().
     */
    async boundedCheck() {
        return boundedCheckViaService(this);
        try {
            var bound = await this.integerDialog('Bounded check', 'Enter bound:', this.currentBound, {
                min: 1,
                okLabel: 'OK',
            });
            if (bound === null) {
                this.controls.setStatus('Bounded check cancelled');
                return;
            }
            this.currentBound = bound;
            this.controls.setStatus('Running bounded check...');
            var result = await this.api.runCheck('bounded', { bound: bound });
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
        return weakenInvariantViaService(this);
        try {
            this.controls.setStatus('Choosing conjectures...');
            var choicesResult = await this.api.executeAction('get_conjectures', {});
            var conjectures = (choicesResult && choicesResult.conjectures) || [];
            var choices = conjectures.map(function (conj, index) {
                var label = conj.label ? '[' + conj.label + '] ' + conj.formula : conj.formula;
                return { label: label || String(index), value: index };
            });
            var indices: any = await this.listboxDialog('Weaken', 'Choose conjectures to remove:', choices, {
                multiple: true,
                okLabel: 'Weaken',
            });
            if (!indices || indices.length === 0) {
                this.controls.setStatus('Weaken cancelled');
                return;
            }
            this.controls.setStatus('Weakening invariant...');
            var result = await this.api.executeAction('weaken', { indices: indices });
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
        return executeAndRefresh(this, { action: 'redo', running: 'Redo...', success: 'Redo complete', failure: 'Redo failed' });
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
        return executeAndRefresh(this, { action: 'pdr_step', running: 'PDR step...', success: 'PDR step complete', failure: 'PDR step failed' });
        this.controls.setStatus('PDR step...');
        try {
            var result = await this.api.executeAction('pdr_step', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('PDR step complete', 'success');
        } catch (e) {
            this.controls.setStatus('PDR step failed: ' + e.message, 'error');
        }
    }

    async showReachableStates() {
        this.controls.setStatus('Opening reachable states...');
        try {
            var result = await this.api.executeAction('show_reachable', {});
            this.openARGSheet('Reachable states', result.arg, result.sheet_id);
            this.controls.setStatus('Reachable states opened', 'success');
        } catch (e) {
            this.controls.setStatus('Show reachable states failed: ' + e.message, 'error');
        }
    }

    /**
     * Show concrete model.
     * Matches Python ivy_graph_ui.py concrete().
     */
    async concreteStep() {
        this.controls.setStatus('Computing concrete model...');
        try {
            var result = await this.api.executeAction('concrete', { sheet_id: this.activeSheetId || 'sheet-1' });
            if (result && result.concept) {
                this.updateConceptGraph(result.concept);
            } else {
                await this.refreshConceptGraph();
            }
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
            var result = await this.api.executeAction('gather', { sheet_id: this.activeSheetId || 'sheet-1' });
            if (result && result.concept) {
                this.updateConceptGraph(result.concept);
            } else {
                await this.refreshConceptGraph();
            }
            this.controls.setStatus('Facts gathered', 'success');
        } catch (e) {
            this.controls.setStatus('Gather failed: ' + e.message, 'error');
        }
    }

    async ctiConceptAction(actionName) {
        return ctiConceptActionViaService(this, actionName);
        this.controls.setStatus('Running CTI action...');
        try {
            var result = await this.api.executeAction(actionName, { sheet_id: this.activeSheetId || 'sheet-1' });
            if (result && result.concept) {
                this.updateConceptGraph(result.concept);
            } else {
                await this.refreshConceptGraph();
            }
            this.controls.setStatus((result && result.message) || 'CTI action complete', 'success');
        } catch (e) {
            this.controls.setStatus('CTI action failed: ' + e.message, 'error');
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
            var result = await this.api.executeAction('path_reach', { sheet_id: this.activeSheetId || 'sheet-1' });
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
            var result = await this.api.executeAction('reach', { sheet_id: this.activeSheetId || 'sheet-1' });
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
            await this.api.executeAction('backtrack', { sheet_id: this.activeSheetId || 'sheet-1' });
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
        return rememberGraphViaService(this);
        this.controls.setStatus('Remembering graph...');
        try {
            var name = await this.entryDialog('Remember graph', 'Enter a name for this goal:', '', { okLabel: 'Remember' });
            if (name === null) {
                this.controls.setStatus('Remember cancelled', 'warning');
                return;
            }
            await this.api.executeAction('remember', { name: name, sheet_id: this.activeSheetId || 'sheet-1' });
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
        return exportConjectureViaService(this);
        this.controls.setStatus('Exporting conjecture...');
        try {
            var result = await this.api.executeAction('export', { sheet_id: this.activeSheetId || 'sheet-1' });
            var content = (result && result.content) || '';
            var filename = (result && result.filename) || 'concept_graph.dot';
            if (content) {
                if (window.showSaveFilePicker) {
                    var handle = await window.showSaveFilePicker({
                        suggestedName: filename,
                        types: [{
                            description: 'DOT files',
                            accept: { 'text/vnd.graphviz': ['.dot'] },
                        }],
                    });
                    var writable = await handle.createWritable();
                    await writable.write(content);
                    await writable.close();
                } else {
                    var blob = new Blob([content], { type: (result && result.mime_type) || 'text/vnd.graphviz' });
                    var url = URL.createObjectURL(blob);
                    var a = document.createElement('a');
                    a.href = url;
                    a.download = filename;
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
                    URL.revokeObjectURL(url);
                }
            }
            this.controls.setStatus('Graph exported', 'success');
        } catch (e) {
            this.controls.setStatus('Export failed: ' + e.message, 'error');
        }
    }

    /**
     * Add a relation from a user-entered string.
     * Matches Python ivy_graph_ui.py add_concept_from_string().
     */
    async addRelationFromString() {
        return addRelationFromStringViaService(this);
        var input = await this.entryDialog(
            'Add relation',
            'Add a relation [example: p(X,a,Y)]:',
            '',
            { okLabel: 'Add' }
        );
        if (!input) return;
        try {
            await this.api.executeAction('add_relation', { formula: input });
            await this.refreshConceptGraph();
            this.controls.setStatus('Relation added', 'success');
        } catch (e) {
            this.controls.setStatus('Add relation failed: ' + e.message, 'error');
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
                this.acceptArgSnapshot(this.activeSheetId || 'sheet-1', argData);
                this.argGraph.update(argData.elements, argData.positions);
            }
            // Refresh concept graph and populate state checkbox pane
            var conceptData = await this.api.getConceptGraph();
            if (conceptData && conceptData.elements) {
                this.acceptConceptSnapshot(this.activeSheetId || 'sheet-1', conceptData);
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
    _showToast(message, level, options: any = {}) {
        options = options || {};
        var toast = document.createElement('div');
        toast.className = 'ivy-toast ivy-toast-floating ivy-toast-' + (level || 'info');
        if (options.className) {
            toast.className += ' ' + options.className;
        }
        toast.textContent = message;
        toast.onclick = function () { toast.remove(); };
        document.body.appendChild(toast);
        if (!options.persistent) {
            setTimeout(function () { toast.remove(); }, 10000);
        }
        return toast;
    }
}

// ================================================================
// Initialize on DOM ready unless a framework shell owns boot timing.
// ================================================================
let activeIvyRuntime = null;

function startIvyRuntime() {
    if (activeIvyRuntime) {
        return activeIvyRuntime;
    }
    activeIvyRuntime = new IvyRuntime();
    activeIvyRuntime.init();
    return activeIvyRuntime;
}

function stopIvyRuntime() {
    activeIvyRuntime = null;
}

export { IvyRuntime, configureIvyRuntimeDependencies, resetIvyRuntimeDependencies, startIvyRuntime, stopIvyRuntime };
