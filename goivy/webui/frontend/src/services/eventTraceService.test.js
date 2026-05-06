import { describe, expect, it, vi } from 'vitest';
import {
  activeEventSheet,
  findEventTrace,
  lookupEventTrace,
  readFileText,
  selectedEventPattern,
  toggleEventTraceNode,
} from './eventTraceService.js';

describe('eventTraceService', () => {
  it('reads files and looks up nested event addresses', async () => {
    await expect(readFileText({ text: () => Promise.resolve('trace') })).resolves.toBe('trace');
    const events = [{ text: 'root', subs: [{ text: 'child' }] }];
    expect(lookupEventTrace(events, '0/0')).toEqual({ text: 'child' });
    expect(lookupEventTrace(events, '0/99')).toBeNull();
  });

  it('uses the Vue bridge to toggle expanded event nodes and selected patterns', () => {
    const bridge = {
      isEventTraceExpanded: vi.fn(() => false),
      setEventTraceExpanded: vi.fn(),
      getSelectedEventPattern: vi.fn(() => 'p(X)'),
    };

    toggleEventTraceNode({}, 'events-1', '0/1', { bridge });

    expect(bridge.setEventTraceExpanded).toHaveBeenCalledWith('events-1', '0/1', true);
    expect(selectedEventPattern({}, 'events-1', { bridge })).toBe('p(X)');
  });

  it('runs find against the active event sheet and selects the result', async () => {
    const app = {
      activeSheetId: 'events-1',
      sheets: {
        'events-1': { id: 'events-1', type: 'events', selectedEventAddress: '0' },
      },
      activeEventSheet() {
        return activeEventSheet(this);
      },
      api: {
        executeAction: vi.fn(async () => ({ address: '0/1' })),
      },
      selectEventTraceRow: vi.fn(),
      controls: {
        setStatus: vi.fn(),
      },
      visualOnlyMessage: vi.fn(),
    };

    await expect(findEventTrace(app, 'pattern', true)).resolves.toEqual({ address: '0/1' });
    expect(app.api.executeAction).toHaveBeenCalledWith('events_find', {
      sheet_id: 'events-1',
      pattern: 'pattern',
      reverse: true,
      anchor: '0',
    });
    expect(app.selectEventTraceRow).toHaveBeenCalledWith('events-1', '0/1');
  });
});
