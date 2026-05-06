import { defineStore } from 'pinia';

let nextToastId = 1;

export const useToastStore = defineStore('toasts', {
  state: () => ({
    items: [],
  }),
  actions: {
    show(message, level = 'info', options = {}) {
      const id = nextToastId++;
      const toast = {
        id,
        message: String(message || ''),
        level: level || 'info',
        className: options.className || '',
        persistent: Boolean(options.persistent),
      };
      this.items.push(toast);
      if (!toast.persistent) {
        const timeoutMs = Number(options.timeoutMs) || 10000;
        window.setTimeout(() => {
          this.remove(id);
        }, timeoutMs);
      }
      return id;
    },
    remove(id) {
      this.items = this.items.filter((toast) => toast.id !== id);
    },
    clear() {
      this.items = [];
    },
  },
});
