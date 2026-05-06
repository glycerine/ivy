import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { isReactive } from 'vue';
import { useEngineStore } from './engineStore.js';

describe('engineStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('keeps engine service objects out of Vue reactivity proxies', () => {
    const store = useEngineStore();

    expect(isReactive(store.engine)).toBe(false);

    const replacement = { createSession: async () => 's1' };
    store.setEngine('test', replacement);

    expect(store.engine).toBe(replacement);
    expect(isReactive(store.engine)).toBe(false);
  });
});
