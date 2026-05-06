/**
 * IvyApp - Main application for the Ivy Interactive Verification web UI.
 *
 * Orchestrates the API, graph views, and UI controls. Handles all
 * user interactions including ARG/concept graph clicks, context menus,
 * file loading, mode selection, and verification checks.
 */

import { ARG_STYLE as DEFAULT_ARG_STYLE, CONCEPT_STYLE as DEFAULT_CONCEPT_STYLE, IvyGraph as DefaultIvyGraph } from './services/graphRuntime.js';
import { IvyAPIShim as DefaultIvyAPI, IvyControlsShim as DefaultIvyControls } from './legacyRuntimeGlobals.js';
import { createIvyPersist } from './services/persistenceService.js';
import { registerCommand, registerControllerCommands } from './services/commandRegistry.js';
import {
    connectSessionEvents,
    createLegacyApi,
    createSession,
    updateSessionDisplay as updateSessionDisplayViaService,
} from './services/sessionService.js';
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
} from './services/editorService.js';
import { initializeLegacyCodeMirror } from './codeMirrorEditor.js';
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
} from './services/fileService.js';
import {
    loadRecentSession as loadRecentSessionViaService,
    populateRecentFiles as populateRecentFilesViaService,
} from './services/recentFileService.js';
import {
    currentSheet as currentSheetViaService,
    graphElementsSnapshot as graphElementsSnapshotViaService,
    installGraphStoreHook as installGraphStoreHookViaService,
    refreshGraphsAndEditorLayout,
    refreshLayoutAfterVuePatch,
    registerSheet as registerSheetViaService,
} from './services/graphService.js';
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
} from './services/conceptVisibilityService.js';
import { populateConstraintFacts as populateConstraintFactsViaService } from './services/detailsService.js';
import {
    assertValidSheetId as assertValidSheetIdViaService,
    isValidSheetId as isValidSheetIdViaService,
    isVisualOnlySheet as isVisualOnlySheetViaService,
    setVisualOnlySheet as setVisualOnlySheetViaService,
    sheetExists as sheetExistsViaService,
    sheetTab as sheetTabViaService,
    switchSheet as switchSheetViaService,
    visualOnlyMessage as visualOnlyMessageViaService,
} from './services/sheetService.js';
import {
    activeEventSheet as activeEventSheetViaService,
    filterEventTrace as filterEventTraceViaService,
    findEventTrace as findEventTraceViaService,
    loadEventTraceFile as loadEventTraceFileViaService,
    lookupEventTrace as lookupEventTraceViaService,
    readFileText as readFileTextViaService,
    selectEventTraceRow as selectEventTraceRowViaService,
    selectedEventPattern as selectedEventPatternViaService,
    toggleEventTraceNode as toggleEventTraceNodeViaService,
    uncoverEventTraceAddress as uncoverEventTraceAddressViaService,
} from './services/eventTraceService.js';
import {
    addCheckResultViewActions as addCheckResultViewActionsViaService,
    autoCheckUsedRelations,
    boundedCheck as boundedCheckViaService,
    checkInduction as checkInductionViaService,
    runCheck as runCheckViaService,
    showCheckResult as showCheckResultViaService,
} from './services/checkService.js';
import {
    executeAndRefresh,
    refreshConceptGraph as refreshConceptGraphViaService,
    runAction as runActionViaService,
} from './services/analysisActionService.js';
import {
    closeAllDropdowns as closeAllDropdownsViaService,
    dispatchMenuDescriptorAction as dispatchMenuDescriptorActionViaService,
    flashAndClose as flashAndCloseViaService,
} from './services/menuService.js';
import {
    analysisStateLimits as analysisStateLimitsViaService,
    buildAnalysisState as buildAnalysisStateViaService,
    removeAnalysisStateExtraSheets as removeAnalysisStateExtraSheetsViaService,
    validateAnalysisStateEvents as validateAnalysisStateEventsViaService,
    validateAnalysisStateGraphPayload as validateAnalysisStateGraphPayloadViaService,
    validateAnalysisStateObject as validateAnalysisStateObjectViaService,
    validateAnalysisStateSheet as validateAnalysisStateSheetViaService,
} from './services/analysisStateService.js';
import {
    executeArgEdgeAction as executeArgEdgeActionViaService,
    executeArgNodeAction as executeArgNodeActionViaService,
    prepareArgNodeActionArgs as prepareArgNodeActionArgsViaService,
} from './services/argActionService.js';
import {
    addProjection as addProjectionViaService,
    executeConceptEdgeAction as executeConceptEdgeActionViaService,
    executeConceptNodeAction as executeConceptNodeActionViaService,
    materializeEdge as materializeEdgeViaService,
    materializeNode as materializeNodeViaService,
    removeConcept as removeConceptViaService,
    selectConceptNode as selectConceptNodeViaService,
    splatterNode as splatterNodeViaService,
    splitConcept as splitConceptViaService,
    supposeEmpty as supposeEmptyViaService,
} from './services/conceptActionService.js';
import {
    buttonListDialog as buttonListDialogViaService,
    entryDialog as entryDialogViaService,
    integerDialog as integerDialogViaService,
    listboxDialog as listboxDialogViaService,
    okCancelDialog as okCancelDialogViaService,
    okDialog as okDialogViaService,
    textDialog as textDialogViaService,
} from './services/dialogService.js';

const defaultLegacyAppDependencies = {
    IvyAPI: DefaultIvyAPI,
    IvyControls: DefaultIvyControls,
    IvyGraph: DefaultIvyGraph,
    IvyPersist: createIvyPersist(globalThis.window),
    ARG_STYLE: DEFAULT_ARG_STYLE,
    CONCEPT_STYLE: DEFAULT_CONCEPT_STYLE,
    CodeMirror: globalThis.window && globalThis.window.CodeMirror,
};
let legacyAppDeps = { ...defaultLegacyAppDependencies };

function configureLegacyAppDependencies(overrides) {
    legacyAppDeps = { ...defaultLegacyAppDependencies, ...(overrides || {}) };
    return legacyAppDeps;
}

function resetLegacyAppDependencies() {
    legacyAppDeps = { ...defaultLegacyAppDependencies };
    return legacyAppDeps;
}

class IvyApp {
    constructor() {
        this.api = this.createApi();
        this.controls = new legacyAppDeps.IvyControls(this.api);
        this.argGraph = null;
        this.conceptGraph = null;
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
        return createLegacyApi({
            fallbackApiFactory: function () {
                return new legacyAppDeps.IvyAPI();
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
        var savedState = legacyAppDeps.IvyPersist.load();

        // Always need a server session for API calls.
        try {
            await createSession(this.api, { controls: this.controls });
        } catch (e) {
            // createSession has already reported the failure through controls.
        }

        // If restoring, keep the saved session's URL hash.
        // If fresh, set the new session ID in the URL.
        if (!savedState || !savedState.fileContent) {
            legacyAppDeps.IvyPersist.setSessionIdInURL(this.api.sessionId);
        }

        // Show the persisted session ID (from URL hash), not the server session ID.
        // These can differ because the server ID increments on restart while
        // the persisted ID is stable across reloads.
        this.updateSessionDisplay(legacyAppDeps.IvyPersist.getSessionIdFromURL() || this.api.sessionId);

        // Create Cytoscape graph instances
        this.argGraph = new legacyAppDeps.IvyGraph('arg-graph', legacyAppDeps.ARG_STYLE);
        this.conceptGraph = new legacyAppDeps.IvyGraph('concept-graph', legacyAppDeps.CONCEPT_STYLE);
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
        if (!(window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.layoutResizersHandled === 'function' &&
            window.__ivyVueBridge.layoutResizersHandled())) {
            this.setupResizer();
            this.setupResizer2();
            this.setupResizer3();
            this.setupResizerH();
            this.setupDetailsResizer();
        }
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.initializeEditor === 'function') {
            this.cmEditor = window.__ivyVueBridge.initializeEditor(this);
        } else if (modelEditor) {
            var codeMirror = legacyAppDeps.CodeMirror || window.CodeMirror;
            this.cmEditor = initializeLegacyCodeMirror({
                legacyApp: this,
                editorStore: { keymap: this.getEditorKeymap() },
                doc: document,
                codeMirror: codeMirror,
            });
            // Keymap radio button switching fallback for non-Vue test harnesses.
            if (!(window.__ivyVueBridge &&
                typeof window.__ivyVueBridge.editorKeymapHandled === 'function' &&
                window.__ivyVueBridge.editorKeymapHandled())) {
                var radios = document.querySelectorAll('input[name="keymap"]');
                for (var i = 0; i < radios.length; i++) {
                    radios[i].addEventListener('change', function () {
                        self.setEditorKeymap(this.value);
                    });
                }
            }
        }

        // Restore saved session if available (survives page reload).
        if (savedState && savedState.fileContent) {
            console.log('IvyPersist: restoring session', savedState.sessionId, savedState.fileName);
            var restored = await legacyAppDeps.IvyPersist.restore(this, savedState);
            if (restored) {
                // Keep the URL hash from the saved session (don't overwrite)
                legacyAppDeps.IvyPersist.setFileName(savedState.fileName, savedState.filePath);
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
            legacyAppDeps.IvyPersist.save(self);
        });

        // Hook into setStatus: auto-save whenever a 'success' status is set.
        var origSetStatus = this.controls.setStatus.bind(this.controls);
        this.controls.setStatus = function (msg, level) {
            origSetStatus(msg, level);
            if (level === 'success') {
                legacyAppDeps.IvyPersist.save(self);
            }
        };
    }

    // --- Editor helpers (CodeMirror) ---

    updateSessionDisplay(sessionId) {
        updateSessionDisplayViaService(sessionId);
    }

    _setVueLayoutSize(method, value) {
        var bridge = window.__ivyVueBridge;
        if (bridge && typeof bridge[method] === 'function') {
            bridge[method](value);
            return true;
        }
        return false;
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.editorSaveSheenHandled === 'function' && window.__ivyVueBridge.editorSaveSheenHandled()) {
            return { vueHandled: true };
        }
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
        rememberLastOpenFile(this, legacyAppDeps.IvyPersist);
    }

    _updateReopenLastFileButton() {
        updateReopenLastFileButton(this);
    }

    async reopenLastFile() {
        return reopenLastFileViaService(this, legacyAppDeps.IvyPersist);
    }

    async _ensureFileHandleWritable() {
        return ensureFileHandleWritable(this);
    }

    async _readFileHandleContent() {
        return readFileHandleContent(this);
    }

    async _restoreFileHandleForCurrentFile() {
        return restoreFileHandleForCurrentFile(this, legacyAppDeps.IvyPersist);
    }

    async _confirmNoExternalChangeBeforeSave(content) {
        return confirmNoExternalChangeBeforeSaveViaService(this, content, legacyAppDeps.IvyPersist);
    }

    _mergeDiskVersionIntoEditBuffer(baseContent, editorContent, diskContent) {
        return mergeDiskVersionIntoEditBuffer(baseContent, editorContent, diskContent);
    }

    showExternalChangeDialog() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'buttons',
                title: 'File changed on disk',
                message: 'The file "' + (this._persistedFileName || 'model') + '" has changed outside IvyWeb. Choose how to handle the current editor buffer.',
                escapeValue: 'do-nothing',
                buttons: [
                    { label: 'Do nothing', value: 'do-nothing' },
                    { label: 'Revert to on-disk version', value: 'reload' },
                    { label: 'Merge disk version into edit buffer', value: 'merge' },
                    { label: 'Overwrite on-disk with edited buffer', value: 'overwrite', danger: true },
                ],
            });
        }
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'buttons',
                title: 'File has changed',
                message: 'File "' + (this._persistedFileName || 'model') + '" has changed. Save before closing?',
                escapeValue: 'cancel',
                buttons: [
                    { label: 'Cancel the close', value: 'cancel' },
                    { label: 'Save', value: 'save' },
                    { label: 'Discard Edits', value: 'discard', danger: true },
                ],
            });
        }
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setSaveAsNoticeVisible === 'function') {
            window.__ivyVueBridge.setSaveAsNoticeVisible(true);
            return;
        }
        var notice = document.getElementById('save-as-explain-notice');
        if (notice) {
            notice.style.display = 'block';
        }
    }

    hideSaveAsExplanationNotice() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setSaveAsNoticeVisible === 'function') {
            window.__ivyVueBridge.setSaveAsNoticeVisible(false);
            return;
        }
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

    installGraphStoreHook(sheetId, kind, graph) {
        installGraphStoreHookViaService(sheetId, kind, graph);
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
        var vueSuppressesNativeGraphMenu = window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.graphContextMenuSuppressionHandled === 'function' &&
            window.__ivyVueBridge.graphContextMenuSuppressionHandled();
        if (!vueSuppressesNativeGraphMenu) {
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
        var vueHandlesStaticCommands = window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.staticCommandHandlersHandled === 'function' &&
            window.__ivyVueBridge.staticCommandHandlersHandled();
        var vueHandlesFileInputs = window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.fileInputHandlersHandled === 'function' &&
            window.__ivyVueBridge.fileInputHandlersHandled();

        if (!vueHandlesFileInputs && fileInput) fileInput.addEventListener('change', function () {
            if (fileInput.files && fileInput.files.length > 0) {
                self.loadFile(fileInput.files[0]);
                fileInput.value = ''; // reset for re-selection of same file
            }
        });

        if (!vueHandlesFileInputs && eventFileInput) {
            eventFileInput.addEventListener('change', function () {
                if (eventFileInput.files && eventFileInput.files.length > 0) {
                    self.loadEventTraceFile(eventFileInput.files[0]);
                    eventFileInput.value = '';
                }
            });
        }

        if (!vueHandlesFileInputs && analysisStateFileInput) {
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

        if (!vueHandlesStaticCommands) {
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
        }

        if (!(window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.globalInteractionsHandled === 'function' &&
            window.__ivyVueBridge.globalInteractionsHandled())) {
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
        }

        this.attachGraphEventHandlers(this.argGraph, this.conceptGraph, 'sheet-1');
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
            if (!(activePanel.id === 'arg-panel' && self._setVueLayoutSize('setArgPanelWidth', newWidth))) {
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
            if (!self._setVueLayoutSize('setStatePanelWidth', newWidth)) {
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
            if (!self._setVueLayoutSize('setEditorWidth', newWidth)) {
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
        if (window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.tabClicksHandled === 'function' &&
            window.__ivyVueBridge.tabClicksHandled()) {
            return;
        }
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
        var bridge = window.__ivyVueBridge;
        var vueTabs = bridge && typeof bridge.activateSheetTab === 'function';
        if (vueTabs) {
            bridge.activateSheetTab(sheetId);
        }

        // Deactivate all tabs and sheets
        var tabs = vueTabs ? [] : document.querySelectorAll('.sheet-tab');
        var sheets = document.querySelectorAll('.sheet-content');
        for (var i = 0; i < tabs.length; i++) tabs[i].classList.remove('active');
        for (var i = 0; i < sheets.length; i++) {
            if (!vueTabs || sheets[i].__ivyVueRenderedSheet) {
                sheets[i].classList.remove('active');
            }
        }
        // Activate the target
        var tab = this.sheetTab(sheetId);
        var sheet = document.getElementById(sheetId);
        if (tab && !vueTabs) tab.classList.add('active');
        if (sheet && (!vueTabs || sheet.__ivyVueRenderedSheet)) sheet.classList.add('active');
        if (this.sheets && this.sheets[sheetId]) {
            this.activeSheetId = sheetId;
            if (this.sheets[sheetId].type !== 'events') {
                this.argGraph = this.sheets[sheetId].argGraph;
                this.conceptGraph = this.sheets[sheetId].conceptGraph;
                this.selectedArgNode = this.sheets[sheetId].selectedArgNode;
            }
        }
        // Resize graphs in the newly visible sheet
        if (this.argGraph) this.argGraph.resize();
        if (this.conceptGraph) this.conceptGraph.resize();
        if (bridge && typeof bridge.setActiveGraphSheet === 'function') {
            bridge.setActiveGraphSheet(sheetId);
        }
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

        // Create tab button with close X (Sheet 1 never has X)
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.upsertSheetTab === 'function') {
            window.__ivyVueBridge.upsertSheetTab({ id: sheetId, label: label, closable: true, type: 'analysis' });
        } else {
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
        }

        // Create sheet content (clone structure from sheet-1)
        var newSheet = null;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.createAnalysisSheetShell === 'function') {
            newSheet = window.__ivyVueBridge.createAnalysisSheetShell({ id: sheetId, counter: this._sheetCounter });
        }
        if (!newSheet) {
            var template = document.getElementById('sheet-1');
            newSheet = template.cloneNode(true);
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
        }

        var graphs = newSheet.querySelectorAll('.graph-container');
        var graphIds = [];
        for (var i = 0; i < graphs.length; i++) {
            graphs[i].innerHTML = '';
            graphIds.push(graphs[i].id);
        }

        var argGraph = new legacyAppDeps.IvyGraph(graphIds[0], legacyAppDeps.ARG_STYLE);
        var conceptGraph = new legacyAppDeps.IvyGraph(graphIds[1], legacyAppDeps.CONCEPT_STYLE);
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
        var vueEventSheets = window.__ivyVueBridge && typeof window.__ivyVueBridge.upsertEventTraceSheet === 'function';
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.upsertSheetTab === 'function') {
            window.__ivyVueBridge.upsertSheetTab({ id: sheetId, label: tabLabel, closable: true, type: 'events' });
        } else {
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
        }

        var sheet = existingSheet;
        if (!sheet && !vueEventSheets) {
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
        if (vueEventSheets) {
            window.__ivyVueBridge.upsertEventTraceSheet({
                id: sheetId,
                label: tabLabel,
                events: events,
                patterns: patterns,
                selectedEventAddress: data.selected_address || null,
            });
        } else {
            this.renderEventTraceSheet(sheetId);
        }
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setEventTraceExpanded === 'function') {
            var expanded = false;
            if (typeof window.__ivyVueBridge.isEventTraceExpanded === 'function') {
                expanded = !!window.__ivyVueBridge.isEventTraceExpanded(sheetId, address);
            }
            window.__ivyVueBridge.setEventTraceExpanded(sheetId, address, !expanded);
            return;
        }
        var row = this.eventTraceRow(sheetId, address);
        var li = row ? row.closest('.event-tree-node') : null;
        var sheetState = this.sheets && this.sheets[sheetId];
        if (!li || !sheetState) return;
        var existing = li.querySelector(':scope > ul.event-tree-list');
        var toggle = row.querySelector('.event-toggle');
        if (existing) {
            existing.remove();
            if (toggle) toggle.textContent = '+';
            return;
        }
        var ev = this.lookupEventTrace(sheetState.events, address);
        if (!ev || !ev.subs || ev.subs.length === 0) return;
        var list = document.createElement('ul');
        list.className = 'event-tree-list';
        for (var i = 0; i < ev.subs.length; i++) {
            list.appendChild(this.renderEventTreeNode(ev.subs[i], sheetId, address + '/' + i));
        }
        li.appendChild(list);
        if (toggle) toggle.textContent = '-';
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
        var sheetState = this.sheets && this.sheets[sheetId];
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.selectEventTraceRow === 'function') {
            if (sheetState) sheetState.selectedEventAddress = address;
            window.__ivyVueBridge.selectEventTraceRow(sheetId, address);
            return;
        }
        this.uncoverEventTraceAddress(sheetId, address);
        var sheet = document.getElementById(sheetId);
        if (!sheetState || !sheet) return;
        var rows = sheet.querySelectorAll('.event-row.selected');
        for (var i = 0; i < rows.length; i++) rows[i].classList.remove('selected');
        var row = this.eventTraceRow(sheetId, address);
        if (row) {
            row.classList.add('selected');
            if (typeof row.scrollIntoView === 'function') {
                row.scrollIntoView({ block: 'nearest' });
            }
        }
        sheetState.selectedEventAddress = address;
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

    applyEventPatternResult(sheetId, result, fallbackPatterns) {
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return;
        if (result && Array.isArray(result.patterns)) {
            sheet.patterns = result.patterns.slice();
        } else if (fallbackPatterns) {
            sheet.patterns = fallbackPatterns.slice();
        }
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateEventPatterns === 'function') {
            window.__ivyVueBridge.updateEventPatterns(sheetId, sheet.patterns || []);
            return;
        }
        this.renderEventPatternList(sheetId);
    }

    renderEventPatternList(sheetId) {
        var sheet = document.getElementById(sheetId);
        var sheetState = this.sheets && this.sheets[sheetId];
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateEventPatterns === 'function') {
            if (sheetState) {
                window.__ivyVueBridge.updateEventPatterns(sheetId, sheetState.patterns || []);
            }
            return;
        }
        var select = sheet ? sheet.querySelector('.event-pattern-list') : null;
        if (!select || !sheetState) return;
        select.innerHTML = '';
        for (var i = 0; i < (sheetState.patterns || []).length; i++) {
            var option = document.createElement('option');
            option.value = sheetState.patterns[i];
            option.textContent = sheetState.patterns[i];
            select.appendChild(option);
        }
    }

    selectedEventPattern(sheetId) {
        return selectedEventPatternViaService(this, sheetId);
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getSelectedEventPattern === 'function') {
            return window.__ivyVueBridge.getSelectedEventPattern(sheetId) || '';
        }
        var sheet = document.getElementById(sheetId);
        var select = sheet ? sheet.querySelector('.event-pattern-list') : null;
        return select && select.value ? select.value : '';
    }

    async addEventPattern(sheetId, pattern) {
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
        var sheet = this.sheets && this.sheets[sheetId];
        var sheetEl = document.getElementById(sheetId);
        var select = sheetEl ? sheetEl.querySelector('.event-pattern-list') : null;
        var idx = -1;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getSelectedEventPatternIndex === 'function') {
            idx = window.__ivyVueBridge.getSelectedEventPatternIndex(sheetId);
        } else if (select) {
            idx = select.selectedIndex;
        }
        if (!sheet || idx < 0) return;
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_remove_pattern', { sheet_id: sheetId, index: idx });
            this.applyEventPatternResult(sheetId, result);
            return;
        }
        sheet.patterns.splice(idx, 1);
        this.renderEventPatternList(sheetId);
    }

    async clearEventPatterns(sheetId) {
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
        var sheetState = this.sheets && this.sheets[sheetId];
        var bridge = window.__ivyVueBridge;
        var vueTabs = bridge && typeof bridge.removeSheetTab === 'function';
        var vueOwnsSheetDom = vueTabs && sheetState && sheetState.type === 'events';

        if (tab && !vueTabs) tab.remove();
        if (bridge && typeof bridge.removeRenderedSheet === 'function') {
            bridge.removeRenderedSheet(sheetId);
        }
        if (sheet && !vueOwnsSheetDom) sheet.remove();
        if (this.sheets) {
            delete this.sheets[sheetId];
        }
        if (vueTabs) {
            bridge.removeSheetTab(sheetId);
        }

        // If the closed tab was active, switch to Sheet 1
        if (wasActive) {
            this.switchSheet('sheet-1');
        }
    }

    toggleTutorial(flash) {
        var bridge = window.__ivyVueBridge;
        if (bridge && typeof bridge.setTutorialVisible === 'function') {
            var tutorialForState = document.getElementById('tutorial-container');
            var visible = tutorialForState ? tutorialForState.style.display !== 'none' : true;
            if (typeof bridge.isTutorialVisible === 'function') {
                visible = !!bridge.isTutorialVisible();
            }
            var nextVisible = !visible;
            bridge.setTutorialVisible(nextVisible);
            if (!nextVisible && flash) {
                if (typeof bridge.flashTutorialButton === 'function') {
                    bridge.flashTutorialButton(1200);
                } else {
                    var vueFallbackButton = document.getElementById('btn-toggle-tutorial');
                    if (vueFallbackButton) {
                        vueFallbackButton.classList.add('btn-flash');
                        setTimeout(function () { vueFallbackButton.classList.remove('btn-flash'); }, 1200);
                    }
                }
            }
            this._refreshLayoutAfterVuePatch();
            return;
        }

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

    _refreshLayoutAfterVuePatch() {
        refreshLayoutAfterVuePatch(this);
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
            if (!self._setVueLayoutSize('setTutorialHeight', newHeight)) {
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
            if (!(activePanel.id === 'info-panel' && self._setVueLayoutSize('setDetailsHeight', newHeight))) {
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
        if (window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.tutorialUrlBarHandled === 'function' &&
            window.__ivyVueBridge.tutorialUrlBarHandled()) {
            return;
        }
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
        return populateStateCheckboxesViaService(this, conceptData);
        // Store concept data for node label rendering
        this._lastConceptData = conceptData;
        var tbody = document.getElementById('state-checkbox-body');
        this._hydrateBackendToggleState(conceptData);

        // Use the relations list from the server (edges + node_labels).
        // This matches Python's Graph.relation_ids.
        var names = [];
        if (conceptData && conceptData.relations) {
            names = conceptData.relations.slice().sort();
        }
        var self = this;

        var appendLegacyRows = function () {
            if (!tbody) return;
            tbody.innerHTML = '';
            for (var i = 0; i < names.length; i++) {
                (function(name) {
                    var tr = document.createElement('tr');

                    // + checkbox (all_to_all)
                    var td1 = document.createElement('td');
                    var cb1 = document.createElement('input');
                    cb1.type = 'checkbox';
                    cb1.name = name;
                    cb1.value = 'all_to_all';
                    cb1.title = 'Show definite edges (' + name + ')';
                    cb1.checked = self._toggleChecked(name, 'all_to_all');
                    cb1.addEventListener('change', function() {
                        self.onEdgeToggle(name, 'all_to_all', cb1.checked);
                    });
                    td1.appendChild(cb1);
                    tr.appendChild(td1);

                    // ? checkbox (unknown)
                    var td2 = document.createElement('td');
                    var cb2 = document.createElement('input');
                    cb2.type = 'checkbox';
                    cb2.name = name;
                    cb2.value = 'edge_unknown';
                    cb2.title = 'Show unknown edges (' + name + ')';
                    cb2.checked = self._toggleChecked(name, 'edge_unknown');
                    cb2.addEventListener('change', function() {
                        self.onEdgeToggle(name, 'edge_unknown', cb2.checked);
                    });
                    td2.appendChild(cb2);
                    tr.appendChild(td2);

                    // - checkbox (none_to_none)
                    var td3 = document.createElement('td');
                    var cb3 = document.createElement('input');
                    cb3.type = 'checkbox';
                    cb3.name = name;
                    cb3.value = 'none_to_none';
                    cb3.title = 'Show absent edges (' + name + ')';
                    cb3.checked = self._toggleChecked(name, 'none_to_none');
                    cb3.addEventListener('change', function() {
                        self.onEdgeToggle(name, 'none_to_none', cb3.checked);
                    });
                    td3.appendChild(cb3);
                    tr.appendChild(td3);

                    // T checkbox (transitive reduction)
                    var td4 = document.createElement('td');
                    var cb4 = document.createElement('input');
                    cb4.type = 'checkbox';
                    cb4.name = name;
                    cb4.value = 'transitive';
                    cb4.title = 'Transitive reduction (' + name + ')';
                    cb4.checked = self._toggleChecked(name, 'transitive');
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
        };

        var vueRows = names.map(function (name) {
            return {
                name: name,
                checked: {
                    all_to_all: self._toggleChecked(name, 'all_to_all'),
                    edge_unknown: self._toggleChecked(name, 'edge_unknown'),
                    none_to_none: self._toggleChecked(name, 'none_to_none'),
                    transitive: self._toggleChecked(name, 'transitive')
                }
            };
        });
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateStateRelations === 'function') {
            window.__ivyVueBridge.updateStateRelations(vueRows, function (name, displayClass, checked) {
                self.onEdgeToggle(name, displayClass, checked);
            });
            this._applyEdgeVisibility();
            this._applyNodeLabels();
            this.populateConstraintFacts(conceptData);
            return;
        }

        if (!tbody) return;
        appendLegacyRows();

        this._applyEdgeVisibility();
        this._applyNodeLabels();
        this.populateConstraintFacts(conceptData);
    }

    populateConstraintFacts(conceptData) {
        return populateConstraintFactsViaService(this, conceptData);
        var facts = (conceptData && Array.isArray(conceptData.facts)) ? conceptData.facts : [];
        var self = this;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateConstraintFacts === 'function') {
            window.__ivyVueBridge.updateConstraintFacts(facts, async function (index, selected) {
                await self.api.executeAction('set_fact_selection', {
                    index: index,
                    selected: selected,
                });
            });
            return;
        }
        var info = document.getElementById('info-content');
        if (!info) return;
        info.innerHTML = '';
        if (facts.length === 0) {
            info.textContent = 'Select a node or edge to see details';
            return;
        }

        var title = document.createElement('div');
        title.className = 'constraint-facts-title';
        title.textContent = 'Constraints:';
        info.appendChild(title);

        facts.forEach(function(fact, offset) {
            var index = typeof fact.index === 'number' ? fact.index : offset;
            var row = document.createElement('button');
            row.type = 'button';
            row.className = 'constraint-fact';
            row.setAttribute('data-constraint-fact', String(index));
            row.setAttribute('aria-pressed', fact.selected ? 'true' : 'false');
            row.textContent = fact.text || '';
            if (!fact.selected) {
                row.classList.add('inactive');
            }
            row.addEventListener('click', async function() {
                var selected = row.classList.contains('inactive');
                row.classList.toggle('inactive', !selected);
                row.setAttribute('aria-pressed', selected ? 'true' : 'false');
                try {
                    await self.api.executeAction('set_fact_selection', {
                        index: index,
                        selected: selected,
                    });
                } catch (e) {
                    row.classList.toggle('inactive', selected);
                    row.setAttribute('aria-pressed', selected ? 'false' : 'true');
                    self.controls.setStatus('Fact selection failed: ' + e.message, 'error');
                }
            });
            info.appendChild(row);
        });
    }

    _hydrateBackendToggleState(conceptData) {
        return hydrateBackendToggleState(this, conceptData);
        this._edgeVisibility = {};
        this._labelVisibility = {};
        var toggles = (conceptData && conceptData.toggles) || {};
        var edges = toggles.edges || {};
        var labels = toggles.labels || {};
        for (var edge in edges) {
            if (edges.hasOwnProperty(edge)) {
                this._edgeVisibility[edge] = Object.assign({
                    all_to_all: false,
                    edge_unknown: false,
                    none_to_none: false,
                    transitive: false,
                }, edges[edge]);
            }
        }
        for (var label in labels) {
            if (labels.hasOwnProperty(label)) {
                this._labelVisibility[label] = Object.assign({
                    node_necessarily: false,
                    node_maybe: false,
                    node_necessarily_not: false,
                }, labels[label]);
            }
        }
    }

    _toggleChecked(name, displayClass) {
        return toggleChecked(this, name, displayClass);
        var base = name.split('(')[0];
        var vis = this._edgeVisibility[name] || this._edgeVisibility[base];
        if (vis && Object.prototype.hasOwnProperty.call(vis, displayClass)) {
            return !!vis[displayClass];
        }
        var labelKeyMap = {
            all_to_all: 'node_necessarily',
            edge_unknown: 'node_maybe',
            none_to_none: 'node_necessarily_not',
        };
        var labelKey = labelKeyMap[displayClass];
        var labelVis = this._labelVisibility[name] || this._labelVisibility[base];
        if (labelKey && labelVis && Object.prototype.hasOwnProperty.call(labelVis, labelKey)) {
            return !!labelVis[labelKey];
        }
        return false;
    }

    /**
     * Handle edge visibility toggle change.
     */
    async onEdgeToggle(edgeName, displayClass, checked) {
        return onEdgeToggleViaService(this, edgeName, displayClass, checked);
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
        try {
            await this.api.setToggles({
                edge: edgeName,
                display_class: displayClass,
                value: checked
            });
            await this.refreshConceptGraph();
        } catch (e) {
            console.error('Toggle error:', e);
        }
    }

    /**
     * Apply edge visibility based on checkbox state.
     * An edge is shown if its display class checkbox is checked.
     * Matches Python cy_render.py line 346:
     *   if widget.edge_display_checkboxes[edge][classes[0]].value is False: skip
     */
    _applyEdgeVisibility(conceptGraph) {
        return applyEdgeVisibility(this, conceptGraph);
        var graph = conceptGraph || this.conceptGraph;
        if (!graph || !graph.cy) return;
        var self = this;
        graph.cy.edges().forEach(function(edge) {
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
        return findEdgeVisibility(this, obj, label);
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
        return applyNodeLabels(this);
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
            var nodeID = node.data('obj') || '';
            if (!nodeID) return;
            var sortName = node.data('cluster') || node.data('sort') || nodeID;
            var topLabel = node.data('display_label') || sortName;

            var labelParts = [topLabel];

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
                var necKey = 'node_label|node_necessarily|' + nodeID + '|' + baseLabelName;
                var necNotKey = 'node_label|node_necessarily_not|' + nodeID + '|' + baseLabelName;
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
                var displayLabelName = self._displayConceptName(baseLabelName);
                if (prefix === '?') {
                    labelParts.push(prefix + displayLabelName);
                } else {
                    labelParts.push(prefix + displayLabelName);
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

    _displayConceptName(name) {
        return displayConceptName(name);
        if (typeof name === 'string' && name.charAt(0) === '=') {
            var body = name.slice(1);
            var idx = body.lastIndexOf(':');
            if (idx > 0) {
                return '=' + body.slice(0, idx);
            }
        }
        return name;
    }

    /**
     * Update the state label to show which ARG node is selected.
     */
    updateStateLabel(nodeId) {
        return updateStateLabelViaService(nodeId);
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateStateLabel === 'function') {
            window.__ivyVueBridge.updateStateLabel(nodeId);
            return;
        }
        var label = document.getElementById('state-label');
        if (label) {
            label.textContent = 'State: ' + (nodeId != null ? nodeId : '—');
        }
    }

    /**
     * Set up keyboard shortcuts.
     */
    setupKeyboardShortcuts() {
        if (window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.globalInteractionsHandled === 'function' &&
            window.__ivyVueBridge.globalInteractionsHandled()) {
            return;
        }
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.selectArgNode === 'function') {
            window.__ivyVueBridge.selectArgNode(sheetId, nodeData.id);
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
        var actions = [
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
        var actions = [
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
        return loadModelFile(this, file, legacyAppDeps.IvyPersist);
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
        return saveModel(this, legacyAppDeps.IvyPersist);
    }

    downloadModelForUnsupportedSave(content) {
        return downloadModelForUnsupportedSaveViaService(this, content, legacyAppDeps.IvyPersist);
    }

    async saveAs(options) {
        return saveModelAs(this, legacyAppDeps.IvyPersist, { options: options || {} });
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getSheetTabLabel === 'function') {
            var bridgeLabel = window.__ivyVueBridge.getSheetTabLabel(sheetId);
            if (bridgeLabel) return bridgeLabel;
        }
        var tab = this.sheetTab(sheetId);
        var label = tab ? tab.querySelector('span') : null;
        return label ? label.textContent : sheetId;
    }

    getMode() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getMode === 'function') {
            return window.__ivyVueBridge.getMode() || 'pdr';
        }
        var modeEl = document.getElementById('mode-select');
        return modeEl ? modeEl.value : 'pdr';
    }

    setMode(mode) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setMode === 'function') {
            window.__ivyVueBridge.setMode(mode);
            return;
        }
        var modeEl = document.getElementById('mode-select');
        if (modeEl && mode) modeEl.value = mode;
    }

    buildAnalysisState() {
        return buildAnalysisStateViaService(this, legacyAppDeps.IvyPersist);
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
            toggles: legacyAppDeps.IvyPersist._getToggles ? legacyAppDeps.IvyPersist._getToggles() : {},
            sheets: sheets,
        };
    }

    async saveAnalysisState() {
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
        if (!file) return false;
        if (typeof file.size === 'number' && file.size > this.analysisStateLimits().maxFileBytes) {
            throw new Error('analysis state file too large');
        }
        var text = await this.readFileText(file);
        return this.loadAnalysisStateObject(JSON.parse(text));
    }

    async loadAnalysisStateObject(state) {
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
            legacyAppDeps.IvyPersist._setToggles(state.toggles);
        }
        if (state.activeSheetId && this.sheetExists(state.activeSheetId)) {
            this.switchSheet(state.activeSheetId);
        } else {
            this.switchSheet('sheet-1');
        }
        legacyAppDeps.IvyPersist.setFileName(this._persistedFileName, this._persistedFilePath);
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
    async newModel(options) {
        return newModelViaService(this, legacyAppDeps.IvyPersist, { options: options || {} });
    }

    /**
     * Populate the recent files list in the File dropdown.
     * Deduplicates by fileName+contentLength, keeping the most recent.
     * Shows truncated path context when file names collide.
     */
    populateRecentFiles() {
        populateRecentFilesViaService(this, legacyAppDeps.IvyPersist);
    }

    /**
     * Load a recent session by its saved session ID.
     */
    async loadRecentSession(savedSessionId) {
        return loadRecentSessionViaService(this, legacyAppDeps.IvyPersist, savedSessionId);
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
        if (!result || !result.trace_arg) return;
        var self = this;
        var openTrace = function () {
            self.openARGSheet('Error trace', result.trace_arg, result.trace_sheet_id);
        };

        var appendLegacyButton = function () {
            if (document.querySelector('[data-check-view-trace]')) return;
            var info = document.getElementById('info-content');
            if (!info) return;
            var button = document.createElement('button');
            button.type = 'button';
            button.className = 'btn small';
            button.setAttribute('data-check-view-trace', 'true');
            button.textContent = 'View error trace';
            button.addEventListener('click', openTrace);
            info.appendChild(document.createElement('br'));
            info.appendChild(button);
        };

        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setCheckTraceAction === 'function') {
            window.__ivyVueBridge.setCheckTraceAction(openTrace);
            return;
        }
        appendLegacyButton();
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.closeDropdownMenus === 'function') {
            window.__ivyVueBridge.closeDropdownMenus();
            return;
        }
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
        var bridge = window.__ivyVueBridge;
        if (bridge && typeof bridge.flashMenuItem === 'function' && el && el.id) {
            bridge.flashMenuItem(el.id, 50);
            setTimeout(function () {
                self.closeAllDropdowns();
                if (callback) callback();
            }, 50);
            return;
        }
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
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateMenuRegion === 'function') {
            var self = this;
            window.__ivyVueBridge.updateMenuRegion(region, menus || [], function (item) {
                return self.dispatchMenuDescriptorAction(region, item);
            });
            return;
        }
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
    async runAction(actionName, args, options) {
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

    _addDialogButton(dialog, label, callback, extraClass) {
        var btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'dialog-btn' + (extraClass ? ' ' + extraClass : '');
        btn.textContent = label || 'OK';
        btn.setAttribute('data-ivy-dialog-button', 'true');
        btn.addEventListener('click', callback);
        dialog.buttons.appendChild(btn);
        return btn;
    }

    _finishDialog(dialog, cleanup, resolve, value) {
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
        var bridgedOk = okDialogViaService(title, message);
        if (bridgedOk !== undefined) return bridgedOk;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({ type: 'ok', title: title, message: message });
        }
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
        var bridgedOkCancel = okCancelDialogViaService(title, message);
        if (bridgedOkCancel !== undefined) return bridgedOkCancel;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({ type: 'okCancel', title: title, message: message });
        }
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
        var bridgedText = textDialogViaService(title, message, text, options);
        if (bridgedText !== undefined) return bridgedText;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'text',
                title: title,
                message: message,
                text: text,
                options: options || {},
            });
        }
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
        var bridgedEntry = entryDialogViaService(title, message, initialValue, options);
        if (bridgedEntry !== undefined) return bridgedEntry;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'entry',
                title: title,
                message: message,
                initialValue: initialValue,
                options: options || {},
            });
        }
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
        var bridgedInteger = integerDialogViaService(title, message, initialValue, options);
        if (bridgedInteger !== undefined) return bridgedInteger;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'integer',
                title: title,
                message: message,
                initialValue: initialValue,
                options: options || {},
            });
        }
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
        var bridgedListbox = listboxDialogViaService(title, message, items, options);
        if (bridgedListbox !== undefined) return bridgedListbox;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'listbox',
                title: title,
                message: message,
                items: items || [],
                options: options || {},
            });
        }
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
        var bridgedButtons = buttonListDialogViaService(title, message, buttons);
        if (bridgedButtons !== undefined) return bridgedButtons;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'buttons',
                title: title,
                message: message,
                buttons: buttons || [],
            });
        }
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
        try {
            this.controls.setStatus('Choosing conjectures...');
            var choicesResult = await this.api.executeAction('get_conjectures', {});
            var conjectures = (choicesResult && choicesResult.conjectures) || [];
            var choices = conjectures.map(function (conj, index) {
                var label = conj.label ? '[' + conj.label + '] ' + conj.formula : conj.formula;
                return { label: label || String(index), value: index };
            });
            var indices = await this.listboxDialog('Weaken', 'Choose conjectures to remove:', choices, {
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
    _showToast(message, level, options) {
        options = options || {};
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showToast === 'function') {
            return window.__ivyVueBridge.showToast(message, level, options);
        }
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
function startIvyApp() {
    if (window.ivyApp) {
        return window.ivyApp;
    }
    window.ivyApp = new IvyApp();
    var unregisterControllerCommands = registerControllerCommands(window.ivyApp);
    var fileCommandUnregisters = [
        registerCommand('file.load', window.ivyApp.chooseAndLoadModelFile.bind(window.ivyApp)),
        registerCommand('file.save', window.ivyApp.save.bind(window.ivyApp)),
        registerCommand('file.saveAs', window.ivyApp.saveAs.bind(window.ivyApp)),
        registerCommand('file.download', window.ivyApp.downloadModel.bind(window.ivyApp)),
        registerCommand('file.close', window.ivyApp.closeCurrentFile.bind(window.ivyApp)),
        registerCommand('file.new', window.ivyApp.newModel.bind(window.ivyApp)),
        registerCommand('file.reopenLast', window.ivyApp.reopenLastFile.bind(window.ivyApp)),
    ];
    window.ivyApp._unregisterCommands = function () {
        unregisterControllerCommands();
        for (var i = 0; i < fileCommandUnregisters.length; i++) {
            fileCommandUnregisters[i]();
        }
    };
    window.ivyApp.init();
    return window.ivyApp;
}

export { IvyApp, configureLegacyAppDependencies, resetLegacyAppDependencies, startIvyApp };
