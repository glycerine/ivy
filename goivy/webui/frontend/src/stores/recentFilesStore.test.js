import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useRecentFilesStore } from './recentFilesStore.js';

describe('recentFilesStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.useRealTimers();
  });

  it('stores recent file rows and dispatches selected ids', () => {
    const recent = useRecentFilesStore();
    const loader = vi.fn();

    recent.setItems([{ id: 'saved-1', label: 'client_server_example.ivy' }], loader);
    recent.load(recent.items[0]);

    expect(recent.items[0].label).toBe('client_server_example.ivy');
    expect(loader).toHaveBeenCalledWith('saved-1');
  });

  it('tracks the flashed recent-file row declaratively', () => {
    const recent = useRecentFilesStore();

    recent.setItems([{ id: 'saved-1', label: 'client_server_example.ivy' }]);
    recent.flash('saved-1');
    expect(recent.flashingId).toBe('saved-1');

    recent.clearFlash();
    expect(recent.flashingId).toBeNull();
  });
});
