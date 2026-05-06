import { defineStore } from 'pinia';

const defaultDetails = 'Select a node or edge to see details';

export const useDetailsStore = defineStore('details', {
  state: () => ({
    text: defaultDetails,
    traceActionVisible: false,
    traceActionCallback: null,
  }),
  actions: {
    setDetails({ shortInfo = '', longInfo = '' } = {}) {
      let lines = [];
      if (shortInfo) lines.push(shortInfo);
      if (longInfo) {
        if (Array.isArray(longInfo)) {
          lines = lines.concat(longInfo);
        } else {
          lines.push(longInfo);
        }
      }
      this.text = lines.join('\n') || defaultDetails;
      this.traceActionVisible = false;
      this.traceActionCallback = null;
    },
    clear() {
      this.text = defaultDetails;
      this.traceActionVisible = false;
      this.traceActionCallback = null;
    },
    setTraceAction(action) {
      this.traceActionCallback = typeof action === 'function' ? action : null;
      this.traceActionVisible = Boolean(this.traceActionCallback);
    },
    runTraceAction() {
      if (this.traceActionCallback) {
        this.traceActionCallback();
      }
    },
  },
});
