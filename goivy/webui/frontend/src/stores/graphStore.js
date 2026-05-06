import { defineStore } from 'pinia';
import { useEngineStore } from './engineStore.js';

export const useGraphStore = defineStore('graph', {
  state: () => ({
    activeSheetId: 'sheet-1',
    argBySheet: {},
    conceptBySheet: {},
    selectedArgNode: null,
    toggles: {},
  }),
  actions: {
    async refreshArg(sheetId = this.activeSheetId) {
      const data = await useEngineStore().engine.getArg({ sheetId });
      this.argBySheet = { ...this.argBySheet, [sheetId]: data };
      return data;
    },
    async refreshConcept({ sheetId = this.activeSheetId, stateId } = {}) {
      const data = await useEngineStore().engine.getConcept({ sheetId, stateId });
      this.conceptBySheet = { ...this.conceptBySheet, [sheetId]: data };
      return data;
    },
    async refreshToggles() {
      this.toggles = await useEngineStore().engine.getToggles();
      return this.toggles;
    },
    selectArgNode(nodeId) {
      this.selectedArgNode = nodeId;
    },
  },
});
