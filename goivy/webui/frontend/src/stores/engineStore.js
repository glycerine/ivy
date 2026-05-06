import { defineStore } from 'pinia';
import { markRaw } from 'vue';
import { HostedGoEngine, WanixEngine } from '../engines/index.js';

export const useEngineStore = defineStore('engine', {
  state: () => ({
    kind: 'hosted-go',
    engine: markRaw(new HostedGoEngine()),
  }),
  actions: {
    useHostedGo(options = {}) {
      this.kind = 'hosted-go';
      this.engine = markRaw(new HostedGoEngine(options));
      return this.engine;
    },
    useWanix() {
      this.kind = 'wanix';
      this.engine = markRaw(new WanixEngine());
      return this.engine;
    },
    setEngine(kind, engine) {
      this.kind = kind;
      this.engine = markRaw(engine);
      return this.engine;
    },
  },
});
