import { defineStore } from 'pinia';

function normalizeEntries(items = []) {
  return items.map((item) => {
    if (typeof item === 'object' && item !== null) {
      return { label: item.label || String(item.value), value: item.value };
    }
    return { label: String(item), value: item };
  });
}

export const useDialogStore = defineStore('dialogs', {
  state: () => ({
    active: null,
  }),
  actions: {
    open(config = {}) {
      return new Promise((resolve) => {
        const type = config.type || 'ok';
        const entries = type === 'listbox' ? normalizeEntries(config.items || []) : [];
        this.active = {
          ...config,
          type,
          title: config.title || 'ivyweb',
          message: config.message || '',
          options: config.options || {},
          entries,
          inputValue: String(config.initialValue == null ? (config.text || '') : config.initialValue),
          selectedIndex: '',
          selectedIndices: [],
          error: '',
          resolve,
        };
      });
    },
    clearError() {
      if (this.active) this.active.error = '';
    },
    setInputValue(value) {
      if (this.active) {
        this.active.inputValue = value;
        this.clearError();
      }
    },
    setSelectedIndex(value) {
      if (this.active) {
        this.active.selectedIndex = value;
      }
    },
    setSelectedIndices(value) {
      if (this.active) {
        this.active.selectedIndices = Array.isArray(value) ? value : [];
      }
    },
    finish(value) {
      const active = this.active;
      this.active = null;
      if (active && typeof active.resolve === 'function') {
        active.resolve(value);
      }
    },
    cancelValue() {
      if (!this.active) return null;
      return this.active.type === 'listbox' && this.active.options.multiple ? [] : null;
    },
    cancel() {
      this.finish(this.cancelValue());
    },
    escape() {
      if (!this.active) return;
      if (this.active.type === 'ok') {
        this.finish(true);
      } else if (this.active.type === 'okCancel') {
        this.finish(false);
      } else if (this.active.escapeValue !== undefined) {
        this.finish(this.active.escapeValue);
      } else {
        this.cancel();
      }
    },
    submit() {
      if (!this.active) return;
      const active = this.active;
      const opts = active.options || {};

      if (active.type === 'ok') {
        this.finish(true);
        return;
      }
      if (active.type === 'okCancel') {
        this.finish(true);
        return;
      }
      if (active.type === 'text' || active.type === 'entry') {
        this.finish(active.inputValue);
        return;
      }
      if (active.type === 'integer') {
        const raw = String(active.inputValue || '').trim();
        const value = Number(raw);
        if (raw === '' || !Number.isInteger(value)) {
          active.error = 'Enter an integer.';
          return;
        }
        if (opts.min != null && value < opts.min) {
          active.error = `Enter a value at least ${opts.min}.`;
          return;
        }
        if (opts.max != null && value > opts.max) {
          active.error = `Enter a value at most ${opts.max}.`;
          return;
        }
        this.finish(value);
        return;
      }
      if (active.type === 'listbox') {
        if (opts.multiple) {
          this.finish(active.selectedIndices.map((index) => active.entries[Number(index)]).filter(Boolean).map((entry) => entry.value));
          return;
        }
        const entry = active.entries[Number(active.selectedIndex)];
        this.finish(entry ? entry.value : null);
      }
    },
    chooseButton(button) {
      this.finish(button ? button.value : null);
    },
  },
});
