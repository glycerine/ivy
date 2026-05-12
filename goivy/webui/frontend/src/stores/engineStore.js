import { defineStore } from 'pinia';
import { markRaw } from 'vue';
import { HostedGoIvyApiAdapter } from '../engines/index.js';

export const useEngineStore = defineStore('engine', {
  state: () => ({
    kind: 'hosted-go',
    engine: markRaw(new HostedGoIvyApiAdapter()),
  }),
  actions: {
    useHostedGo(options = {}) {
      this.kind = 'hosted-go';
      this.engine = markRaw(new HostedGoIvyApiAdapter(options));
      return this.engine;
    },
    setEngine(kind, engine) {
      this.kind = kind;
      this.engine = markRaw(engine);
      return this.engine;
    },
  },
});
