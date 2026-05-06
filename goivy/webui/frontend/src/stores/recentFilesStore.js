import { defineStore } from 'pinia';

export const useRecentFilesStore = defineStore('recentFiles', {
  state: () => ({
    items: [],
    loader: null,
    flashingId: null,
  }),
  actions: {
    setItems(items = [], loader = null) {
      this.items = Array.isArray(items) ? items : [];
      this.loader = typeof loader === 'function' ? loader : null;
      this.flashingId = null;
    },
    flash(id) {
      this.flashingId = id == null ? null : String(id);
    },
    clearFlash() {
      this.flashingId = null;
    },
    load(item) {
      if (item && this.loader) {
        this.loader(item.id);
      }
    },
  },
});
