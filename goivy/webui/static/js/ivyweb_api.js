/**
 * IvyAPI - Backend communication layer for the Ivy verification UI.
 *
 * Handles all HTTP requests to the Go backend JSON API and
 * Server-Sent Events (SSE) for real-time updates.
 */

'use strict';

class IvyAPI {
    /**
     * @param {string} baseURL - Base URL for API requests (default: current origin)
     */
    constructor(baseURL) {
        this.baseURL = baseURL || '';
        this.sessionId = null;
        this.eventSource = null;
    }

    /**
     * Internal helper for fetch with JSON response handling.
     * @param {string} path - API path (appended to baseURL)
     * @param {object} options - fetch options
     * @returns {Promise<object>} parsed JSON response
     */
    async _request(path, options) {
        var opts = options || {};
        var url = this.baseURL + path;
        var resp = await fetch(url, opts);
        if (!resp.ok) {
            var text = '';
            try { text = await resp.text(); } catch (e) { /* ignore */ }
            throw new Error('API error ' + resp.status + ': ' + (text || resp.statusText));
        }
        var contentType = resp.headers.get('content-type') || '';
        if (contentType.indexOf('application/json') !== -1) {
            return resp.json();
        }
        return resp.text();
    }

    /**
     * Create a new verification session.
     * @returns {Promise<string>} session ID
     */
    async createSession() {
        var data = await this._request('/api/session/new', { method: 'POST' });
        this.sessionId = data.session_id;
        return this.sessionId;
    }

    /**
     * Load an .ivy file into the current session.
     * @param {File} file - File object from file input
     * @returns {Promise<object>} load result
     */
    async loadFile(file) {
        var formData = new FormData();
        formData.append('file', file);
        return this._request('/api/session/' + this.sessionId + '/load', {
            method: 'POST',
            body: formData,
        });
    }

    /**
     * Re-upload editor content so the server recompiles before a check.
     * @param {string} content - current editor text
     * @param {string} filename - file name for the server
     * @returns {Promise<object>} load result
     */
    async reloadContent(content, filename) {
        var blob = new Blob([content], { type: 'text/plain' });
        var formData = new FormData();
        formData.append('file', blob, filename || 'model.ivy');
        return this._request('/api/session/' + this.sessionId + '/load', {
            method: 'POST',
            body: formData,
        });
    }

    /**
     * Get the current Analysis Reachability Graph.
     * @returns {Promise<object>} ARG graph data (Cytoscape elements JSON)
     */
    async getARG() {
        return this._request('/api/session/' + this.sessionId + '/arg');
    }

    /**
     * Get browser menu descriptors supplied by the Go UI model.
     * @returns {Promise<object>} menu descriptors keyed by graph region
     */
    async getMenus() {
        return this._request('/api/session/' + this.sessionId + '/menus');
    }

    /**
     * Get the concept graph for the currently selected ARG node.
     * @param {string} [nodeId] - Optional ARG node ID to view
     * @returns {Promise<object>} concept graph data (Cytoscape elements JSON)
     */
    async getConceptGraph(nodeId, sheetId) {
        var path = '/api/session/' + this.sessionId + '/concept';
        var params = [];
        if (sheetId) {
            params.push('sheet=' + encodeURIComponent(sheetId));
        }
        if (nodeId) {
            params.push('node=' + encodeURIComponent(nodeId));
        }
        if (params.length > 0) {
            path += '?' + params.join('&');
        }
        return this._request(path);
    }

    /**
     * Get the proof graph/stack.
     * @returns {Promise<object>} proof graph data (Cytoscape elements JSON)
     */
    async getProofGraph() {
        return this._request('/api/session/' + this.sessionId + '/proof');
    }

    /**
     * Execute a generic action on the session.
     * @param {string} action - Action name
     * @param {object} args - Action arguments
     * @returns {Promise<object>} action result
     */
    async executeAction(action, args) {
        return this._request('/api/session/' + this.sessionId + '/action', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action: action, args: args || {} }),
        });
    }

    /**
     * Split a concept by another concept.
     * @param {string} concept - Concept node to split
     * @param {string} splitBy - Concept/label to split by
     * @returns {Promise<object>}
     */
    async splitConcept(concept, splitBy) {
        return this._request('/api/session/' + this.sessionId + '/concept/split', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ concept: concept, split_by: splitBy }),
        });
    }

    /**
     * Suppose a concept is empty.
     * @param {string} concept - Concept node ID
     * @returns {Promise<object>}
     */
    async supposeEmpty(concept) {
        return this._request('/api/session/' + this.sessionId + '/concept/empty', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ concept: concept }),
        });
    }

    /**
     * Remove a concept from the domain.
     * @param {string} concept - Concept node or edge ID
     * @returns {Promise<object>}
     */
    async removeConcept(concept) {
        return this._request('/api/session/' + this.sessionId + '/concept/remove', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ concept: concept }),
        });
    }

    /**
     * Undo the last concept domain change.
     * @returns {Promise<object>}
     */
    async undo() {
        return this._request('/api/session/' + this.sessionId + '/concept/undo', {
            method: 'POST',
        });
    }

    /**
     * Materialize a concept node (create a concrete witness).
     * @param {string} concept - Concept node ID
     * @returns {Promise<object>}
     */
    async materializeNode(concept) {
        return this._request('/api/session/' + this.sessionId + '/concept/materialize', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ concept: concept, type: 'node' }),
        });
    }

    /**
     * Materialize a concept edge (positive or negative).
     * @param {string} concept - Concept edge ID
     * @param {boolean} positive - true for positive witness, false for negative
     * @returns {Promise<object>}
     */
    async materializeEdge(relation, source, target, positive) {
        return this._request('/api/session/' + this.sessionId + '/concept/materialize', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ relation: relation, source: source, target: target, type: 'edge', positive: positive }),
        });
    }

    /**
     * Add a projection concept.
     * @param {string} name - Projection name
     * @param {string} concept - Base concept
     * @returns {Promise<object>}
     */
    async addProjection(name, concept) {
        return this._request('/api/session/' + this.sessionId + '/concept/projection', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: name, concept: concept }),
        });
    }

    /**
     * Run a verification check in the selected mode.
     * @param {string} mode - Verification mode (concrete, abstract, bounded, induction, pdr)
     * @returns {Promise<object>} check result
     */
    async runCheck(mode) {
        return this._request('/api/session/' + this.sessionId + '/check', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ mode: mode }),
        });
    }

    /**
     * Reset the concept domain.
     * @returns {Promise<object>}
     */
    async resetDomain() {
        return this._request('/api/session/' + this.sessionId + '/concept/reset', {
            method: 'POST',
        });
    }

    /**
     * Switch to diagram concept domain.
     * @returns {Promise<object>}
     */
    async diagramDomain() {
        return this._request('/api/session/' + this.sessionId + '/concept/diagram', {
            method: 'POST',
        });
    }

    /**
     * Get edge and label toggle configuration from server.
     * @returns {Promise<object>} toggle configuration
     */
    async getToggles() {
        return this._request('/api/session/' + this.sessionId + '/toggles');
    }

    /**
     * Update edge/label visibility on the server.
     * @param {object} toggles - Toggle state object
     * @returns {Promise<object>}
     */
    async setToggles(toggles) {
        return this._request('/api/session/' + this.sessionId + '/toggles', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(toggles),
        });
    }

    /**
     * Execute an ARG node action (view state, execute action, mark, cover, safety check).
     * @param {string} nodeId - ARG node ID
     * @param {string} action - Action name
     * @param {object} [args] - Optional additional arguments
     * @returns {Promise<object>}
     */
    async argNodeAction(nodeId, action, args) {
        return this._request('/api/session/' + this.sessionId + '/arg/action', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ node: nodeId, action: action, args: args || {} }),
        });
    }

    /**
     * Execute a proof goal action.
     * @param {string} goalId - Proof goal ID
     * @param {string} action - Action name
     * @returns {Promise<object>}
     */
    async proofGoalAction(goalId, action) {
        return this._request('/api/session/' + this.sessionId + '/proof/action', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ goal: goalId, action: action }),
        });
    }

    /**
     * Save the current session state.
     * @returns {Promise<Blob>} session file blob
     */
    async saveSession() {
        var resp = await fetch(this.baseURL + '/api/session/' + this.sessionId + '/save');
        if (!resp.ok) {
            throw new Error('Save failed: ' + resp.statusText);
        }
        return resp.blob();
    }

    /**
     * Connect to Server-Sent Events for real-time updates.
     * Uses manual reconnect with exponential backoff instead of
     * EventSource's default auto-reconnect, which can storm
     * indefinitely and destabilize browser tests.
     *
     * @param {function} onEvent - Callback receiving parsed event objects
     */
    connectEvents(onEvent) {
        this.disconnectEvents();
        this._sseRetries = 0;
        this._sseMaxRetries = 5;
        this._sseClosed = false;
        this._sseOnEvent = onEvent;
        this._sseConnect();
    }

    /**
     * Internal: open one EventSource connection.
     * On error, reconnects with exponential backoff up to _sseMaxRetries.
     */
    _sseConnect() {
        if (this._sseClosed || !this.sessionId) return;

        var url = this.baseURL + '/api/session/' + this.sessionId + '/events';
        var es = new EventSource(url);
        this.eventSource = es;
        var self = this;

        es.onopen = function () {
            self._sseRetries = 0;  // reset on successful connect
        };

        es.onmessage = function (e) {
            try {
                var event = JSON.parse(e.data);
                self._sseOnEvent(event);
            } catch (err) {
                console.error('Failed to parse SSE event:', err, e.data);
            }
        };

        es.onerror = function () {
            es.close();  // stop browser auto-reconnect
            if (self._sseClosed) return;
            self._sseRetries++;
            if (self._sseRetries > self._sseMaxRetries) {
                console.error('SSE: max retries exceeded, giving up');
                if (self.onConnectionLost) {
                    self.onConnectionLost();
                }
                return;
            }
            var delay = Math.min(1000 * Math.pow(2, self._sseRetries - 1), 16000);
            console.warn('SSE: reconnect attempt ' + self._sseRetries +
                         '/' + self._sseMaxRetries + ' in ' + delay + 'ms');
            self._sseTimer = setTimeout(function () {
                self._sseConnect();
            }, delay);
        };
    }

    /**
     * Disconnect from Server-Sent Events.
     */
    disconnectEvents() {
        this._sseClosed = true;
        if (this._sseTimer) {
            clearTimeout(this._sseTimer);
            this._sseTimer = null;
        }
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }
    }
}
