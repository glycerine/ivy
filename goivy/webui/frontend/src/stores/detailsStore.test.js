import { beforeEach, describe, expect, it } from 'vitest';
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
});
