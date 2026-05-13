import { defineStore } from 'pinia';
import { useEngineStore } from './engineStore.js';

export const useGraphStore = defineStore('graph', {
  state: () => ({
    argBySheet: {},
    conceptBySheet: {},
    toggles: {},
  }),
  actions: {
    applyGraphSnapshot(sheetId, kind, snapshot = {}) {
      if (!sheetId || (kind !== 'arg' && kind !== 'concept')) return;
      const target = kind === 'arg' ? 'argBySheet' : 'conceptBySheet';
      this[target] = {
        ...this[target],
        [sheetId]: {
          elements: Array.isArray(snapshot.elements) ? snapshot.elements : [],
          positions: snapshot.positions || null,
          updatedAt: snapshot.updatedAt || Date.now(),
        },
      };
    },
    async refreshArg(sheetId = 'sheet-1') {
      const data = await useEngineStore().engine.getARG({ sheetId });
      this.argBySheet = { ...this.argBySheet, [sheetId]: data };
      return data;
    },
    async refreshConcept({ sheetId = 'sheet-1', stateId } = {}) {
      const data = await useEngineStore().engine.getConceptGraph(stateId, sheetId);
      this.conceptBySheet = { ...this.conceptBySheet, [sheetId]: data };
      return data;
    },
    async refreshToggles() {
      this.toggles = await useEngineStore().engine.getToggles();
      return this.toggles;
    },
  },
});
