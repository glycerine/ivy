import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useDetailsStore } from './detailsStore.js';

describe('detailsStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('combines short and long details like the legacy controls', () => {
    const details = useDetailsStore();

    details.setDetails({ shortInfo: 'State 0', longInfo: ['x = 1', 'y = 2'] });

    expect(details.text).toBe('State 0\nx = 1\ny = 2');
  });

  it('clears legacy trace actions when new details arrive', () => {
    const details = useDetailsStore();
    details.setTraceAction(() => {});

    details.setDetails({ shortInfo: 'New selection' });

    expect(details.traceActionVisible).toBe(false);
    expect(details.traceActionCallback).toBeNull();
  });

  it('stores constraint facts and toggles through a callback', async () => {
    const details = useDetailsStore();
    const callback = vi.fn(async () => {});

    details.setConstraintFacts([
      { index: 3, text: 'link(a,b)', selected: true },
    ], callback);
    await details.toggleFact(3);

    expect(details.facts[0].selected).toBe(false);
    expect(callback).toHaveBeenCalledWith(3, false);
  });
});
