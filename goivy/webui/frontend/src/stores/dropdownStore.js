import { defineStore } from 'pinia';

export const useDropdownStore = defineStore('dropdowns', {
  state: () => ({
    openId: '',
    flashingItemId: '',
  }),
  getters: {
    isOpen: (state) => (id) => state.openId === id,
  },
  actions: {
    toggle(id) {
      this.openId = this.openId === id ? '' : id;
      return this.openId === id;
    },
    closeAll() {
      this.openId = '';
    },
    flashItem(id, durationMs = 50) {
      if (!id) return;
      this.flashingItemId = String(id);
      window.setTimeout(() => {
        if (this.flashingItemId === String(id)) {
          this.flashingItemId = '';
        }
      }, durationMs);
    },
  },
});
