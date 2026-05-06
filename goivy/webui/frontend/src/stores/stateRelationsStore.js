import { defineStore } from 'pinia';

const emptyColumns = {
  all_to_all: false,
  edge_unknown: false,
  none_to_none: false,
  transitive: false,
};

export const useStateRelationsStore = defineStore('stateRelations', {
  state: () => ({
    stateLabel: '0',
    loaded: false,
    rows: [],
    toggleCallback: null,
  }),
  getters: {
    hasRows: (state) => state.rows.length > 0,
    showPlaceholder: (state) => state.loaded && state.rows.length === 0,
  },
  actions: {
    setStateLabel(value) {
      this.stateLabel = value == null ? '\u2014' : String(value);
    },
    setRows(rows = [], toggleCallback = null) {
      this.loaded = true;
      this.rows = rows.map((row) => ({
        name: row.name,
        checked: { ...emptyColumns, ...(row.checked || {}) },
      }));
      this.toggleCallback = typeof toggleCallback === 'function' ? toggleCallback : null;
    },
    clear() {
      this.loaded = false;
      this.rows = [];
      this.toggleCallback = null;
      this.stateLabel = '0';
    },
    toggle(name, displayClass, checked) {
      const row = this.rows.find((candidate) => candidate.name === name);
      if (row) {
        row.checked[displayClass] = checked;
      }
      if (this.toggleCallback) {
        this.toggleCallback(name, displayClass, checked);
      }
    },
  },
});
