function jsonPost(body = {}) {
  return {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  };
}

export class LegacyApiAdapter {
  constructor(engine) {
    this.engine = engine;
    this.onConnectionLost = null;
    this._sseRetries = 0;
    this._sseMaxRetries = 5;
    this._sseClosed = true;
    this._sseTimer = null;
    this._sseOnEvent = null;
    this._unsubscribe = null;
  }

  get sessionId() {
    return this.engine.sessionId;
  }

  set sessionId(value) {
    this.engine.sessionId = value;
  }

  get eventSource() {
    return this.engine.eventSource;
  }

  set eventSource(value) {
    this.engine.eventSource = value;
  }

  async createSession() {
    return this.engine.createSession();
  }

  async loadFile(file) {
    return this.engine.loadModel({ file });
  }

  async reloadContent(content, filename) {
    return this.engine.loadModel({ content, filename: filename || 'model.ivy' });
  }

  async getARG() {
    return this.engine.getArg();
  }

  async getMenus() {
    return this.engine.getMenus();
  }

  async getConceptGraph(nodeId, sheetId) {
    return this.engine.getConcept({ nodeId, sheetId });
  }

  async getProofGraph() {
    return this.engine.requestSession('/proof');
  }

  async executeAction(action, args) {
    return this.engine.runAction({ action, args: args || {} });
  }

  async splitConcept(concept, splitBy) {
    return this.engine.requestSession('/concept/split', jsonPost({ concept, split_by: splitBy }));
  }

  async supposeEmpty(concept) {
    return this.engine.requestSession('/concept/empty', jsonPost({ concept }));
  }

  async removeConcept(concept) {
    return this.engine.requestSession('/concept/remove', jsonPost({ concept }));
  }

  async undo() {
    return this.engine.requestSession('/concept/undo', { method: 'POST' });
  }

  async materializeNode(concept) {
    return this.engine.requestSession('/concept/materialize', jsonPost({ concept, type: 'node' }));
  }

  async materializeEdge(relation, source, target, positive) {
    return this.engine.requestSession('/concept/materialize', jsonPost({
      relation,
      source,
      target,
      type: 'edge',
      positive,
    }));
  }

  async addProjection(name, concept) {
    return this.engine.requestSession('/concept/projection', jsonPost({ name, concept }));
  }

  async runCheck(mode, options = {}) {
    return this.engine.check({ mode, ...options });
  }

  async resetDomain() {
    return this.engine.requestSession('/concept/reset', { method: 'POST' });
  }

  async diagramDomain() {
    return this.engine.requestSession('/concept/diagram', { method: 'POST' });
  }

  async getToggles() {
    return this.engine.getToggles();
  }

  async setToggles(toggles) {
    return this.engine.setToggles(toggles);
  }

  async argNodeAction(nodeId, action, args) {
    return this.engine.runArgAction({ node: nodeId, action, args: args || {} });
  }

  async proofGoalAction(goalId, action) {
    return this.engine.requestSession('/proof/action', jsonPost({ goal: goalId, action }));
  }

  async saveSession() {
    const response = await this.engine.fetchSession('/save');
    if (!response.ok) {
      throw new Error(`Save failed: ${response.statusText || response.status || 'unknown error'}`);
    }
    return response.blob();
  }

  connectEvents(onEvent) {
    this.disconnectEvents();
    if (typeof this.engine.eventSourceFactory !== 'function' || typeof this.engine.sessionPath !== 'function') {
      this._unsubscribe = this.engine.subscribeEvents(onEvent, () => {
        if (this.onConnectionLost) this.onConnectionLost();
      });
      return;
    }
    this._sseRetries = 0;
    this._sseClosed = false;
    this._sseOnEvent = onEvent;
    this._sseConnect();
  }

  _sseConnect() {
    if (this._sseClosed || !this.sessionId) return;
    const source = this.engine.eventSourceFactory(this.engine.sessionPath('/events'));
    this.eventSource = source;

    source.onopen = () => {
      this._sseRetries = 0;
    };

    source.onmessage = (event) => {
      try {
        if (this._sseOnEvent) this._sseOnEvent(JSON.parse(event.data));
      } catch (err) {
        console.error('Failed to parse SSE event:', err, event.data);
      }
    };

    source.onerror = () => {
      source.close();
      if (this._sseClosed) return;
      this._sseRetries += 1;
      if (this._sseRetries > this._sseMaxRetries) {
        console.error('SSE: max retries exceeded, giving up');
        if (this.onConnectionLost) this.onConnectionLost();
        return;
      }
      const delay = Math.min(1000 * Math.pow(2, this._sseRetries - 1), 16000);
      console.warn(`SSE: reconnect attempt ${this._sseRetries}/${this._sseMaxRetries} in ${delay}ms`);
      this._sseTimer = window.setTimeout(() => {
        this._sseConnect();
      }, delay);
    };
  }

  disconnectEvents() {
    this._sseClosed = true;
    if (this._unsubscribe) {
      this._unsubscribe();
      this._unsubscribe = null;
    }
    if (this._sseTimer) {
      window.clearTimeout(this._sseTimer);
      this._sseTimer = null;
    }
    if (this.eventSource && typeof this.eventSource.close === 'function') {
      this.eventSource.close();
    }
    this.eventSource = null;
  }
}
