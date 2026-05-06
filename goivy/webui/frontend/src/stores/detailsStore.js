import { defineStore } from 'pinia';

const defaultDetails = 'Select a node or edge to see details';

export const useDetailsStore = defineStore('details', {
  state: () => ({
    text: defaultDetails,
    facts: [],
    factCallback: null,
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
      this.facts = [];
      this.factCallback = null;
      this.traceActionVisible = false;
      this.traceActionCallback = null;
    },
    clear() {
      this.text = defaultDetails;
      this.facts = [];
      this.factCallback = null;
      this.traceActionVisible = false;
      this.traceActionCallback = null;
    },
    setConstraintFacts(facts = [], callback = null) {
      this.facts = facts.map((fact, offset) => ({
        index: typeof fact.index === 'number' ? fact.index : offset,
        text: fact.text || '',
        selected: fact.selected !== false,
      }));
      this.factCallback = typeof callback === 'function' ? callback : null;
      this.text = this.facts.length > 0 ? '' : defaultDetails;
      this.traceActionVisible = false;
      this.traceActionCallback = null;
    },
    async toggleFact(index) {
      const fact = this.facts.find((candidate) => candidate.index === index);
      if (!fact) return;
      const previous = fact.selected;
      fact.selected = !previous;
      if (!this.factCallback) return;
      try {
        await this.factCallback(index, fact.selected);
      } catch (err) {
        fact.selected = previous;
        throw err;
      }
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
