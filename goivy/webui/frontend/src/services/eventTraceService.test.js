import { describe, expect, it, vi } from 'vitest';
import {
  activeEventSheet,
  addEventPattern,
  applyEventPatternResult,
  filterEventTrace,
  findEventTrace,
  loadEventPatterns,
  loadEventTraceFile,
  lookupEventTrace,
  readFileText,
  removeSelectedEventPattern,
  renderEventPatternList,
  saveEventPatterns,
  selectEventTraceRow,
  selectedEventPattern,
  toggleEventTraceNode,
} from './eventTraceService.js';

function makePatternApp(overrides = {}) {
  const app = {
    sheets: {
      'events-1': { id: 'events-1', type: 'events', patterns: ['old'], events: [], selectedEventAddress: '' },
    },
    activeSheetId: 'events-1',
    activeEventSheet() {
      return activeEventSheet(this);
    },
    applyEventPatternResult(sheetId, result, fallbackPatterns) {
      return applyEventPatternResult(this, sheetId, result, fallbackPatterns, { bridge: null });
    },
    renderEventPatternList: vi.fn(),
    openEventTraceSheet: vi.fn(),
    readFileText,
    downloadTextFile: vi.fn(),
    controls: { setStatus: vi.fn() },
    visualOnlyMessage: vi.fn(() => 'visual only'),
    api: {
      executeAction: vi.fn(),
    },
    ...overrides,
  };
  return app;
}

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

  it('filters through the backend, opens returned event sheets, and reports filter errors', async () => {
    const app = makePatternApp();
    app.api.executeAction = vi.fn(async () => ({
      sheet_id: 'events-2',
      label: 'Sheet 1',
      events: [{ text: 'call(a)', address: '0' }],
      patterns: [],
    }));

    await filterEventTrace(app, 'call(a)');

    expect(app.api.executeAction).toHaveBeenCalledWith('events_filter', {
      sheet_id: 'events-1',
      pattern: 'call(a)',
    });
    expect(app.openEventTraceSheet).toHaveBeenCalledWith('Sheet 1', expect.objectContaining({ sheet_id: 'events-2' }), 'events-2');

    app.api.executeAction = vi.fn(async () => {
      throw new Error('syntax error');
    });
    const result = await filterEventTrace(app, 'call(');

    expect(result).toBeNull();
    expect(app.controls.setStatus).toHaveBeenCalledWith('Filter failed: syntax error', 'error');
  });

  it('loads a raw event trace through the backend parser', async () => {
    const app = makePatternApp({
      api: {
        executeAction: vi.fn(async () => ({
          sheet_id: 'events-raw',
          label: 'Sheet 0',
          events: [{ text: 'root(a)', address: '0' }],
          patterns: [],
        })),
      },
    });
    const file = new File(['root(a);'], 'trace.iev', { type: 'text/plain' });

    await loadEventTraceFile(app, file);

    expect(app.api.executeAction).toHaveBeenCalledWith('events_parse', {
      content: 'root(a);',
      filename: 'trace.iev',
    });
    expect(app.openEventTraceSheet).toHaveBeenCalledWith('trace.iev', expect.objectContaining({ sheet_id: 'events-raw' }), 'events-raw');
  });

  it('keeps event pattern state backend-authoritative on add failure and success', async () => {
    const app = makePatternApp();
    app.api.executeAction = vi.fn(async () => {
      throw new Error('syntax error');
    });

    await expect(addEventPattern(app, 'events-1', 'call(')).rejects.toThrow('syntax error');
    expect(app.sheets['events-1'].patterns).toEqual(['old']);

    app.api.executeAction = vi.fn(async () => ({ patterns: ['old', 'call(a)'] }));
    await addEventPattern(app, 'events-1', 'call(a)');

    expect(app.sheets['events-1'].patterns).toEqual(['old', 'call(a)']);
  });

  it('keeps event pattern state unchanged when remove/load fail', async () => {
    document.body.innerHTML = '<div id="events-1"><select class="event-pattern-list"><option>first</option><option selected>second</option></select></div>';
    const app = makePatternApp({
      sheets: {
        'events-1': { id: 'events-1', type: 'events', patterns: ['first', 'second'] },
      },
    });
    app.api.executeAction = vi.fn(async () => {
      throw new Error('backend refused');
    });

    await expect(removeSelectedEventPattern(app, 'events-1', { bridge: null, doc: document })).rejects.toThrow('backend refused');
    expect(app.sheets['events-1'].patterns).toEqual(['first', 'second']);

    await expect(loadEventPatterns(app, 'events-1', 'broken(')).rejects.toThrow('backend refused');
    expect(app.sheets['events-1'].patterns).toEqual(['first', 'second']);
  });

  it('uses backend content when saving event patterns', async () => {
    const app = makePatternApp();
    app.api.executeAction = vi.fn(async () => ({ content: 'server\n' }));

    const content = await saveEventPatterns(app, 'events-1');

    expect(content).toBe('server\n');
    expect(app.downloadTextFile).toHaveBeenCalledWith('event_patterns.pats', 'server\n', 'text/plain');
  });

  it('renders and updates event pattern lists through Vue or DOM fallback', () => {
    const app = makePatternApp();
    const bridge = {
      updateEventPatterns: vi.fn(),
    };

    applyEventPatternResult(app, 'events-1', { patterns: ['server'] }, null, { bridge });
    expect(app.sheets['events-1'].patterns).toEqual(['server']);
    expect(bridge.updateEventPatterns).toHaveBeenCalledWith('events-1', ['server']);

    document.body.innerHTML = '<div id="events-1"><select class="event-pattern-list"></select></div>';
    renderEventPatternList(app, 'events-1', { bridge: null, doc: document });
    expect(Array.from(document.querySelectorAll('option')).map((option) => option.textContent)).toEqual(['server']);
  });

  it('selects event rows whose addresses would be unsafe CSS selectors', () => {
    document.body.innerHTML = '<div id="events-1"><div class="event-row" data-event-address="0&quot;]"></div></div>';
    const app = makePatternApp({
      eventTraceRow(sheetId, address) {
        const sheet = document.getElementById(sheetId);
        return Array.from(sheet.querySelectorAll('.event-row')).find((row) => row.getAttribute('data-event-address') === address);
      },
      uncoverEventTraceAddress: vi.fn(),
    });

    expect(() => selectEventTraceRow(app, 'events-1', '0"]', { bridge: null, doc: document })).not.toThrow();
    expect(document.querySelector('.event-row').classList.contains('selected')).toBe(true);
  });
});
