import { defineStore } from 'pinia';

const ROOT_SHEET = {
  id: 'sheet-1',
  label: 'Sheet 1',
  closable: false,
  type: 'analysis',
};

export const useSheetStore = defineStore('sheets', {
  state: () => ({
    tabs: [ROOT_SHEET],
    activeSheetId: ROOT_SHEET.id,
  }),
  getters: {
    labelFor: (state) => (sheetId) => {
      const tab = state.tabs.find((candidate) => candidate.id === sheetId);
      return tab ? tab.label : '';
    },
  },
  actions: {
    upsertTab(tab) {
      if (!tab || !tab.id) return;
      const next = {
        id: String(tab.id),
        label: tab.label || tab.id,
        closable: tab.closable !== false && tab.id !== ROOT_SHEET.id,
        type: tab.type || 'analysis',
      };
      const index = this.tabs.findIndex((existing) => existing.id === next.id);
      if (index >= 0) {
        this.tabs[index] = { ...this.tabs[index], ...next };
      } else {
        this.tabs.push(next);
      }
    },
    activateTab(sheetId) {
      if (!sheetId) return;
      this.activeSheetId = String(sheetId);
    },
    removeTab(sheetId) {
      if (!sheetId || sheetId === ROOT_SHEET.id) return;
      this.tabs = this.tabs.filter((tab) => tab.id !== sheetId);
      if (this.activeSheetId === sheetId) {
        this.activeSheetId = ROOT_SHEET.id;
      }
    },
    resetTabs() {
      this.tabs = [ROOT_SHEET];
      this.activeSheetId = ROOT_SHEET.id;
    },
  },
});
