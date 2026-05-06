import { defineStore } from 'pinia';

export const useLayoutStore = defineStore('layout', {
  state: () => ({
    tutorialVisible: true,
    detailsHeight: 160,
    editorWidth: 520,
  }),
  actions: {
    setTutorialVisible(visible) {
      this.tutorialVisible = Boolean(visible);
    },
    setDetailsHeight(height) {
      this.detailsHeight = Math.max(80, Number(height) || 80);
    },
    setEditorWidth(width) {
      this.editorWidth = Math.max(200, Number(width) || 200);
    },
  },
});
