import { defineStore } from 'pinia';
import { useEngineStore } from './engineStore.js';

export const useSessionStore = defineStore('session', {
  state: () => ({
    sessionId: '',
    mode: 'pdr',
    status: 'Ready',
    statusLevel: '',
    loading: false,
    loadingMessage: 'Loading...',
    events: [],
    unsubscribeEvents: null,
  }),
  actions: {
    setStatus(message, level = '') {
      this.status = message;
      this.statusLevel = level;
    },
    showLoading(message = 'Loading...') {
      this.loading = true;
      this.loadingMessage = message || 'Loading...';
    },
    hideLoading() {
      this.loading = false;
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
