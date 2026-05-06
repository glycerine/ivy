import { defineStore } from 'pinia';
import { useEngineStore } from './engineStore.js';

export const SESSION_MODES = ['induction', 'pdr', 'concrete', 'abstract', 'bounded'];

export const useSessionStore = defineStore('session', {
  state: () => ({
    sessionId: '',
    mode: 'pdr',
    loadedFileDisplay: '',
    loadedFileTitle: '',
    status: 'Ready',
    statusLevel: '',
    loading: false,
    loadingMessage: 'Loading...',
    saveAsNoticeVisible: false,
    events: [],
    unsubscribeEvents: null,
  }),
  actions: {
    setStatus(message, level = '') {
      this.status = message;
      this.statusLevel = level;
    },
    setMode(mode) {
      const next = String(mode || '');
      this.mode = SESSION_MODES.includes(next) ? next : 'pdr';
    },
    setLoadedFile(fileName = '', filePath = '') {
      const display = filePath || fileName || '';
      this.loadedFileDisplay = display;
      this.loadedFileTitle = display;
    },
    showLoading(message = 'Loading...') {
      this.loading = true;
      this.loadingMessage = message || 'Loading...';
    },
    hideLoading() {
      this.loading = false;
    },
    setSaveAsNoticeVisible(visible) {
      this.saveAsNoticeVisible = Boolean(visible);
    },
    async createSession() {
      const engine = useEngineStore().engine;
      this.setStatus('Initializing...');
      this.sessionId = await engine.createSession();
      this.setStatus('Ready');
      return this.sessionId;
    },
    subscribeEvents() {
      const engine = useEngineStore().engine;
      if (this.unsubscribeEvents) {
        this.unsubscribeEvents();
        this.unsubscribeEvents = null;
      }
      this.unsubscribeEvents = engine.subscribeEvents(
        (event) => this.events.push(event),
        () => this.setStatus('Server connection lost', 'error'),
      );
      return this.unsubscribeEvents;
    },
    closeEvents() {
      if (this.unsubscribeEvents) {
        this.unsubscribeEvents();
      }
      this.unsubscribeEvents = null;
    },
  },
});
