import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useEventTraceStore } from './eventTraceStore.js';

describe('eventTraceStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('normalizes events, expands selected ancestors, and tracks patterns', () => {
    const traces = useEventTraceStore();

    traces.upsertSheet({
      id: 'events-1',
      label: 'Trace',
      events: [{ text: 'root', subs: [{ text: 'child' }] }],
      patterns: ['root', 'child'],
      selectedEventAddress: '0/0',
    });
    traces.setSelectedPatternIndex('events-1', 1);

    expect(traces.sheetById('events-1').events[0].address).toBe('0');
    expect(traces.sheetById('events-1').events[0].subs[0].address).toBe('0/0');
    expect(traces.isExpanded('events-1', '0')).toBe(true);
    expect(traces.selectedPattern('events-1')).toBe('child');

    traces.setPatterns('events-1', ['root']);
    expect(traces.selectedPatternIndex('events-1')).toBe(-1);
  });
});
