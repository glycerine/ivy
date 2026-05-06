import { defineStore } from 'pinia';

function normalizeEvents(events = [], prefix = '') {
  return (events || []).map((event, index) => {
    const address = event && event.address != null
      ? String(event.address)
      : (prefix === '' ? String(index) : `${prefix}/${index}`);
    return {
      ...(event || {}),
      address,
      text: (event && event.text) || '',
      subs: normalizeEvents((event && event.subs) || [], address),
    };
  });
}

function parentAddresses(address) {
  const parts = String(address || '').split('/');
  const parents = [];
  for (let index = 0; index < parts.length - 1; index += 1) {
    parents.push(parts.slice(0, index + 1).join('/'));
  }
  return parents;
}

export const useEventTraceStore = defineStore('eventTrace', {
  state: () => ({
    sheets: {},
    expandedBySheet: {},
    selectedPatternIndexBySheet: {},
  }),
  getters: {
    sheetList: (state) => Object.values(state.sheets),
    sheetById: (state) => (sheetId) => state.sheets[sheetId] || null,
    isExpanded: (state) => (sheetId, address) => !!(state.expandedBySheet[sheetId] && state.expandedBySheet[sheetId][address]),
    selectedPatternIndex: (state) => (sheetId) => state.selectedPatternIndexBySheet[sheetId] ?? -1,
    selectedPattern: (state) => (sheetId) => {
      const sheet = state.sheets[sheetId];
      const index = state.selectedPatternIndexBySheet[sheetId] ?? -1;
      return sheet && index >= 0 ? (sheet.patterns[index] || '') : '';
    },
  },
  actions: {
    upsertSheet(payload = {}) {
      if (!payload.id) return;
      const id = String(payload.id);
      this.sheets[id] = {
        id,
        label: payload.label || id,
        events: normalizeEvents(payload.events || []),
        patterns: Array.isArray(payload.patterns) ? payload.patterns.slice() : [],
        selectedEventAddress: payload.selectedEventAddress || payload.selected_address || '',
      };
      if (!this.expandedBySheet[id]) this.expandedBySheet[id] = {};
      if (this.sheets[id].selectedEventAddress) {
        this.selectEvent(id, this.sheets[id].selectedEventAddress);
      }
    },
    removeSheet(sheetId) {
      delete this.sheets[sheetId];
      delete this.expandedBySheet[sheetId];
      delete this.selectedPatternIndexBySheet[sheetId];
    },
    reset() {
      this.sheets = {};
      this.expandedBySheet = {};
      this.selectedPatternIndexBySheet = {};
    },
    setExpanded(sheetId, address, expanded) {
      if (!this.expandedBySheet[sheetId]) this.expandedBySheet[sheetId] = {};
      this.expandedBySheet[sheetId][address] = !!expanded;
    },
    selectEvent(sheetId, address) {
      const sheet = this.sheets[sheetId];
      if (!sheet) return;
      sheet.selectedEventAddress = address || '';
      parentAddresses(address).forEach((parent) => this.setExpanded(sheetId, parent, true));
    },
    setPatterns(sheetId, patterns = []) {
      const sheet = this.sheets[sheetId];
      if (!sheet) return;
      sheet.patterns = Array.isArray(patterns) ? patterns.slice() : [];
      const selected = this.selectedPatternIndexBySheet[sheetId] ?? -1;
      if (selected >= sheet.patterns.length) {
        this.selectedPatternIndexBySheet[sheetId] = -1;
      }
    },
    setSelectedPatternIndex(sheetId, index) {
      this.selectedPatternIndexBySheet[sheetId] = Number.isInteger(index) ? index : -1;
    },
  },
});
