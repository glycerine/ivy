import { defineStore } from 'pinia';

export const useDropdownStore = defineStore('dropdowns', {
  state: () => ({
    openId: '',
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
  },
});
