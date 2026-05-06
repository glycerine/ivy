import { IvyHttpClient } from '../api/ivyHttpClient.js';
import { IvyEngine } from './ivyEngine.js';

function toFormDataFromModel(model = {}) {
  const formData = new FormData();
  if (model.file) {
    formData.append('file', model.file);
    return formData;
  }
  const content = model.content || '';
  const filename = model.filename || model.path || 'model.ivy';
  const blob = new Blob([content], { type: 'text/plain' });
  formData.append('file', blob, filename);
  return formData;
}

export class HostedGoEngine extends IvyEngine {
  constructor({
    client = new IvyHttpClient(),
    eventSourceFactory = (url) => new EventSource(url),
  } = {}) {
    super();
    this.client = client;
    this.eventSourceFactory = eventSourceFactory;
    this.sessionId = null;
    this.eventSource = null;
  }

  sessionPath(suffix) {
    if (!this.sessionId) {
      throw new Error('No Ivy session has been created yet');
    }
    return `/api/session/${this.sessionId}${suffix}`;
  }

  async requestSession(suffix, options = {}) {
    return this.client.request(this.sessionPath(suffix), options);
  }

  async fetchSession(suffix, options = {}) {
    return this.client.fetch(this.sessionPath(suffix), options);
  }

  async createSession() {
    const data = await this.client.request('/api/session/new', { method: 'POST' });
    this.sessionId = data.session_id;
    return this.sessionId;
  }

  async loadModel(model = {}) {
    return this.client.request(this.sessionPath('/load'), {
      method: 'POST',
      body: toFormDataFromModel(model),
    });
  }

  async getArg({ sheetId } = {}) {
    const suffix = sheetId ? `/arg?sheet=${encodeURIComponent(sheetId)}` : '/arg';
    return this.client.request(this.sessionPath(suffix));
  }

  async getConcept({ sheetId, nodeId, stateId } = {}) {
    const params = new URLSearchParams();
    if (sheetId) params.set('sheet', sheetId);
    const selectedNode = nodeId != null ? nodeId : stateId;
    if (selectedNode != null) params.set('node', String(selectedNode));
    const query = params.toString();
    return this.client.request(this.sessionPath(`/concept${query ? `?${query}` : ''}`));
  }

  async getMenus() {
    return this.client.request(this.sessionPath('/menus'));
  }

  async check({ mode, ...options } = {}) {
    const body = { ...options };
    if (mode) body.mode = mode;
    return this.client.request(this.sessionPath('/check'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  async runAction(action = {}) {
    return this.client.request(this.sessionPath('/action'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(action),
    });
  }

  async runArgAction(action = {}) {
    return this.client.request(this.sessionPath('/arg/action'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(action),
    });
  }

  async getToggles() {
    return this.client.request(this.sessionPath('/toggles'));
  }

  async setToggles(toggles = {}) {
    return this.client.request(this.sessionPath('/toggles'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(toggles),
    });
  }

  subscribeEvents(onEvent, onError) {
    if (!this.sessionId) {
      throw new Error('No Ivy session has been created yet');
    }
    this.closeEvents();
    const source = this.eventSourceFactory(this.sessionPath('/events'));
    this.eventSource = source;
    source.onmessage = (event) => {
      if (!onEvent) return;
      try {
        onEvent(JSON.parse(event.data));
      } catch (_err) {
        onEvent(event.data);
      }
    };
    source.onerror = (event) => {
      if (onError) onError(event);
    };
    return () => this.closeEvents();
  }

  closeEvents() {
    if (this.eventSource && typeof this.eventSource.close === 'function') {
      this.eventSource.close();
    }
    this.eventSource = null;
  }
}
