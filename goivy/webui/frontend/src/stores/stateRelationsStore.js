import { defineStore } from 'pinia';

const emptyColumns = {
  all_to_all: false,
  edge_unknown: false,
  none_to_none: false,
  transitive: false,
};

const edgeColumns = ['all_to_all', 'edge_unknown', 'none_to_none', 'transitive'];
const labelColumns = [
  ['all_to_all', 'node_necessarily'],
  ['edge_unknown', 'node_maybe'],
  ['none_to_none', 'node_necessarily_not'],
];

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
    toggleSnapshot: (state) => {
      const toggles = {};
      state.rows.forEach((row) => {
        edgeColumns.forEach((column) => {
          toggles[`${row.name}|${column}`] = Boolean(row.checked[column]);
        });
      });
      return toggles;
    },
    visibilitySnapshot: (state) => {
      const edges = {};
      const labels = {};
      state.rows.forEach((row) => {
        edges[row.name] = {};
        edgeColumns.forEach((column) => {
          edges[row.name][column] = Boolean(row.checked[column]);
        });
        labels[row.name] = {};
        labelColumns.forEach(([edgeColumn, labelColumn]) => {
          labels[row.name][labelColumn] = Boolean(row.checked[edgeColumn]);
        });
      });
      return { edges, labels };
    },
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
    applyToggleSnapshot(toggles = {}) {
      this.rows.forEach((row) => {
        edgeColumns.forEach((column) => {
          const key = `${row.name}|${column}`;
          if (Object.prototype.hasOwnProperty.call(toggles, key)) {
            row.checked[column] = Boolean(toggles[key]);
          }
        });
      });
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
