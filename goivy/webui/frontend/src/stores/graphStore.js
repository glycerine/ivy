import { defineStore } from 'pinia';
import { useEngineStore } from './engineStore.js';

export const useGraphStore = defineStore('graph', {
  state: () => ({
    activeSheetId: 'sheet-1',
    argBySheet: {},
    conceptBySheet: {},
    selectedArgNode: null,
    selectedArgNodeBySheet: {},
    toggles: {},
  }),
  actions: {
    setActiveSheet(sheetId) {
      if (!sheetId) return;
      this.activeSheetId = String(sheetId);
    },
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
    async refreshArg(sheetId = this.activeSheetId) {
      const data = await useEngineStore().engine.getARG({ sheetId });
      this.argBySheet = { ...this.argBySheet, [sheetId]: data };
      return data;
    },
    async refreshConcept({ sheetId = this.activeSheetId, stateId } = {}) {
      const data = await useEngineStore().engine.getConceptGraph(stateId, sheetId);
      this.conceptBySheet = { ...this.conceptBySheet, [sheetId]: data };
      return data;
    },
    async refreshToggles() {
      this.toggles = await useEngineStore().engine.getToggles();
      return this.toggles;
    },
    selectArgNode(nodeId, sheetId = this.activeSheetId) {
      this.selectedArgNode = nodeId;
      if (sheetId) {
        this.selectedArgNodeBySheet = { ...this.selectedArgNodeBySheet, [sheetId]: nodeId };
      }
    },
  },
});
