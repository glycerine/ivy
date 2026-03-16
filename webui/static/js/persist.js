/**
 * IvyPersist - Client-side state persistence for the Ivy web UI.
 *
 * Saves session state to localStorage so that accidental page reloads
 * don't lose work. Supports multiple concurrent users/sessions by
 * keying on session ID (each browser has its own localStorage, and
 * each tab gets its own session ID).
 *
 * State saved:
 *   - Session ID
 *   - File name and content (for re-upload on restore)
 *   - Selected ARG node
 *   - Concept graph selected nodes
 *   - State checkbox toggle values
 *   - Mode selection
 *   - Timestamp
 *
 * Storage layout in localStorage:
 *   "ivy_sessions"        -> JSON array of session IDs (most recent first)
 *   "ivy_sess_{id}"       -> JSON object with full session state
 *   "ivy_last_session"    -> ID of the most recently active session
 */

'use strict';

var IvyPersist = {

    MAX_SESSIONS: 1000, // effectively unlimited — never discard user data

    /**
     * Save the current app state to localStorage.
     * Called automatically after file loads, state changes, etc.
     * @param {IvyApp} app - The app instance
     */
    save: function (app) {
        if (!app || !app.api || !app.api.sessionId) return;
        var sid = app.api.sessionId;
        try {
            var state = {
                sessionId: sid,
                timestamp: Date.now(),
                fileName: app._persistedFileName || '',
                filePath: app._persistedFilePath || app._persistedFileName || '',
                fileContent: app._persistedFileContent || '',
                selectedArgNode: app.selectedArgNode || null,
                mode: IvyPersist._getMode(),
                selectedConceptNodes: IvyPersist._getSelectedConceptNodes(app),
                toggles: IvyPersist._getToggles(),
                argElements: IvyPersist._getCyElements(app.argGraph),
                conceptElements: IvyPersist._getCyElements(app.conceptGraph),
                conceptRelations: app._persistedConceptRelations || null,
            };
            localStorage.setItem('ivy_sess_' + sid, JSON.stringify(state));
            localStorage.setItem('ivy_last_session', sid);
            IvyPersist._updateSessionList(sid);
        } catch (e) {
            console.warn('IvyPersist.save failed:', e);
        }
    },

    /**
     * Try to restore state from a previous session after page reload.
     * Returns the saved state object, or null if nothing to restore.
     * @returns {object|null}
     */
    load: function () {
        try {
            // Check URL hash first — allows multiple tabs with different sessions.
            // URL format: http://host:port/#sessionId
            var sid = IvyPersist.getSessionIdFromURL();
            if (!sid) {
                sid = localStorage.getItem('ivy_last_session');
            }
            if (!sid) return null;
            var raw = localStorage.getItem('ivy_sess_' + sid);
            if (!raw) return null;
            var state = JSON.parse(raw);
            // Sessions are permanent — never expire.
            return state;
        } catch (e) {
            console.warn('IvyPersist.load failed:', e);
            return null;
        }
    },

    /**
     * Load a specific session by ID.
     * @param {string} sid - Session ID
     * @returns {object|null}
     */
    loadSession: function (sid) {
        try {
            var raw = localStorage.getItem('ivy_sess_' + sid);
            if (!raw) return null;
            return JSON.parse(raw);
        } catch (e) {
            return null;
        }
    },

    /**
     * List all saved sessions, newest first.
     * @returns {Array<{id: string, fileName: string, timestamp: number}>}
     */
    listSessions: function () {
        try {
            var raw = localStorage.getItem('ivy_sessions');
            if (!raw) return [];
            var ids = JSON.parse(raw);
            var sessions = [];
            for (var i = 0; i < ids.length; i++) {
                var s = IvyPersist.loadSession(ids[i]);
                if (s) {
                    sessions.push({
                        id: ids[i],
                        fileName: s.fileName || '(unnamed)',
                        filePath: s.filePath || s.fileName || '',
                        timestamp: s.timestamp || 0,
                    });
                }
            }
            return sessions;
        } catch (e) {
            return [];
        }
    },

    /**
     * Delete a saved session.
     * @param {string} sid
     */
    deleteSession: function (sid) {
        try {
            localStorage.removeItem('ivy_sess_' + sid);
            var raw = localStorage.getItem('ivy_sessions');
            if (raw) {
                var ids = JSON.parse(raw);
                ids = ids.filter(function (id) { return id !== sid; });
                localStorage.setItem('ivy_sessions', JSON.stringify(ids));
            }
            var last = localStorage.getItem('ivy_last_session');
            if (last === sid) {
                localStorage.removeItem('ivy_last_session');
            }
        } catch (e) {
            console.warn('IvyPersist.deleteSession failed:', e);
        }
    },

    /**
     * Restore app state from a saved state object.
     * This re-uploads the file to the server and restores UI state.
     * @param {IvyApp} app
     * @param {object} state - saved state from load()
     */
    restore: async function (app, state) {
        if (!state || !state.fileContent) return false;

        app.controls.setStatus('Restoring session...');
        try {
            // Re-upload the file to the server to rebuild compiled module
            var blob = new Blob([state.fileContent], { type: 'text/plain' });
            var file = new File([blob], state.fileName || 'restored.ivy');
            var result = await app.api.loadFile(file);

            // Store for future saves
            app._persistedFileName = state.fileName;
            app._persistedFileContent = state.fileContent;

            // Refresh graphs from server (rebuilt from re-uploaded file)
            var argData = await app.api.getARG();
            if (argData && argData.elements) {
                app.argGraph.update(argData.elements, argData.positions);
            }
            var conceptData = await app.api.getConceptGraph();
            if (conceptData && conceptData.elements) {
                app.conceptGraph.update(conceptData.elements, conceptData.positions);
            }
            app._persistedConceptRelations = conceptData;
            app.populateStateCheckboxes(conceptData);

            // Restore mode
            if (state.mode) {
                var sel = document.getElementById('mode-select');
                if (sel) sel.value = state.mode;
            }

            // Restore selected concept nodes
            if (state.selectedConceptNodes && app.conceptGraph && app.conceptGraph.cy) {
                for (var i = 0; i < state.selectedConceptNodes.length; i++) {
                    var node = app.conceptGraph.cy.getElementById(state.selectedConceptNodes[i]);
                    if (node.length > 0) {
                        node.addClass('selected_node');
                    }
                }
            }

            // Restore toggles
            if (state.toggles) {
                IvyPersist._setToggles(state.toggles);
            }

            // Restore selected ARG node
            if (state.selectedArgNode) {
                app.selectedArgNode = state.selectedArgNode;
            }

            // Update file name display and URL
            IvyPersist.setFileName(state.fileName);
            IvyPersist.setSessionIdInURL(app.api.sessionId);

            app.controls.setStatus('Restored: ' + (state.fileName || 'session'), 'success');
            return true;
        } catch (e) {
            console.error('IvyPersist.restore failed:', e);
            app.controls.setStatus('Restore failed: ' + e.message, 'error');
            return false;
        }
    },

    // --- Internal helpers ---

    _updateSessionList: function (sid) {
        var raw = localStorage.getItem('ivy_sessions');
        var ids = raw ? JSON.parse(raw) : [];
        // Move sid to front, dedup
        ids = ids.filter(function (id) { return id !== sid; });
        ids.unshift(sid);
        // Trim old sessions
        while (ids.length > IvyPersist.MAX_SESSIONS) {
            var old = ids.pop();
            localStorage.removeItem('ivy_sess_' + old);
        }
        localStorage.setItem('ivy_sessions', JSON.stringify(ids));
    },

    _getMode: function () {
        var sel = document.getElementById('mode-select');
        return sel ? sel.value : 'concrete';
    },

    _getSelectedConceptNodes: function (app) {
        var ids = [];
        if (app.conceptGraph && app.conceptGraph.cy) {
            app.conceptGraph.cy.nodes('.selected_node').forEach(function (n) {
                ids.push(n.id());
            });
        }
        return ids;
    },

    _getToggles: function () {
        var toggles = {};
        var table = document.getElementById('state-checkbox-body');
        if (!table) return toggles;
        var inputs = table.querySelectorAll('input[type="radio"], input[type="checkbox"]');
        for (var i = 0; i < inputs.length; i++) {
            var inp = inputs[i];
            if (inp.name) {
                toggles[inp.name + '|' + inp.value] = inp.checked;
            }
        }
        return toggles;
    },

    _setToggles: function (toggles) {
        var table = document.getElementById('state-checkbox-body');
        if (!table) return;
        var inputs = table.querySelectorAll('input[type="radio"], input[type="checkbox"]');
        for (var i = 0; i < inputs.length; i++) {
            var inp = inputs[i];
            var key = inp.name + '|' + inp.value;
            if (key in toggles) {
                inp.checked = toggles[key];
            }
        }
    },

    _getCyElements: function (graph) {
        if (!graph || !graph.cy) return null;
        return graph.cy.json().elements;
    },

    /**
     * Get session ID from URL hash fragment.
     * URL format: http://host:port/#sessionId
     * @returns {string|null}
     */
    getSessionIdFromURL: function () {
        var hash = window.location.hash;
        if (hash && hash.length > 1) {
            return hash.substring(1); // strip leading '#'
        }
        return null;
    },

    /**
     * Set the session ID in the URL hash fragment (without page reload).
     * @param {string} sid
     */
    setSessionIdInURL: function (sid) {
        if (sid) {
            window.history.replaceState(null, '', '#' + sid);
        }
    },

    /**
     * Update the file name display in the menubar.
     * @param {string} fileName
     */
    setFileName: function (fileName) {
        var el = document.getElementById('loaded-file');
        if (el) {
            el.textContent = fileName || '';
        }
    },

    /**
     * Truncate a file path to at most maxChars, preserving the end
     * (closest to the file) and adding "..." at the front.
     * Shows at least the parent directory.
     * @param {string} path
     * @param {number} maxChars
     * @returns {string}
     */
    truncatePath: function (path, maxChars) {
        if (!path || path.length <= maxChars) return path || '';
        // Strip the filename itself — we just want the directory context
        var sep = path.lastIndexOf('/');
        if (sep < 0) sep = path.lastIndexOf('\\');
        var dir = (sep >= 0) ? path.substring(0, sep) : '';
        if (!dir) return '';
        if (dir.length <= maxChars) return dir;
        return '...' + dir.substring(dir.length - maxChars + 3);
    },
};
