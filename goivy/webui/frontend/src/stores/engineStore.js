import { defineStore } from 'pinia';
import { HostedGoEngine, WanixEngine } from '../engines/index.js';

export const useEngineStore = defineStore('engine', {
  state: () => ({
    kind: 'hosted-go',
    engine: new HostedGoEngine(),
  }),
  actions: {
    useHostedGo(options = {}) {
      this.kind = 'hosted-go';
      this.engine = new HostedGoEngine(options);
      return this.engine;
    },
    useWanix() {
      this.kind = 'wanix';
      this.engine = new WanixEngine();
      return this.engine;
    },
    setEngine(kind, engine) {
      this.kind = kind;
      this.engine = engine;
      return this.engine;
    },
  },
});
