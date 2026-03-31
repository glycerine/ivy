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
     * Get the current Analysis Reachability Graph.
     * @returns {Promise<object>} ARG graph data (Cytoscape elements JSON)
     */
    async getARG() {
        return this._request('/api/session/' + this.sessionId + '/arg');
    }

    /**
     * Get the concept graph for the currently selected ARG node.
     * @param {string} [nodeId] - Optional ARG node ID to view
     * @returns {Promise<object>} concept graph data (Cytoscape elements JSON)
     */
    async getConceptGraph(nodeId) {
        var path = '/api/session/' + this.sessionId + '/concept';
        if (nodeId) {
            path += '?node=' + encodeURIComponent(nodeId);
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
    async materializeEdge(concept, positive) {
        return this._request('/api/session/' + this.sessionId + '/concept/materialize', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ concept: concept, type: 'edge', positive: positive }),
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
     * @param {function} onEvent - Callback receiving parsed event objects
     */
    connectEvents(onEvent) {
        if (this.eventSource) {
            this.eventSource.close();
        }
        var url = this.baseURL + '/api/session/' + this.sessionId + '/events';
        this.eventSource = new EventSource(url);

        this.eventSource.onmessage = function (e) {
            try {
                var event = JSON.parse(e.data);
                onEvent(event);
            } catch (err) {
                console.error('Failed to parse SSE event:', err, e.data);
            }
        };

        this.eventSource.onerror = function () {
            console.warn('SSE connection error, will auto-reconnect');
        };
    }

    /**
     * Disconnect from Server-Sent Events.
     */
    disconnectEvents() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }
    }
}
