import { describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from './helpers/fakes.mjs';

function installEventDom() {
  document.body.innerHTML = [
    '<div id="sheet-area">',
    '  <div id="tab-bar"><button class="sheet-tab active" data-sheet="sheet-1"><span>Sheet 1</span></button></div>',
    '  <div id="sheet-1" class="sheet-content active"></div>',
    '</div>',
  ].join('');
}

function makeEventApp() {
  installEventDom();
  const IvyApp = loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
    IvyPersist: makePersist(),
  });
  const app = new IvyApp();
  app.controls = new FakeControls();
  app.setupTabs();
  return app;
}

describe('IvyApp event trace viewer', () => {
  it('opens an event sheet with lazy child expansion and selectable rows', () => {
    const app = makeEventApp();

    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [
        { text: 'root(a)', address: '0', subs: [{ text: 'child(a)', address: '0/0' }] },
      ],
      patterns: [],
    });

    expect(document.querySelector('.sheet-tab[data-sheet="events-1"]').textContent).toContain('Trace');
    expect(document.querySelector('[data-event-address="0"]').textContent).toContain('root(a)');
    expect(document.querySelector('[data-event-address="0/0"]')).toBeNull();

    document.querySelector('[data-event-toggle="0"]').click();
    expect(document.querySelector('[data-event-address="0/0"]').textContent).toContain('child(a)');

    document.querySelector('[data-event-address="0/0"]').click();
    expect(app.sheets['events-1'].selectedEventAddress).toBe('0/0');
    expect(document.querySelector('[data-event-address="0/0"]').classList.contains('selected')).toBe(true);
  });

  it('filters through the backend and opens the returned event sheet', async () => {
    const app = makeEventApp();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [{ text: 'call(a)', address: '0' }],
      patterns: [],
    });
    app.api = {
      executeAction: vi.fn(async () => ({
        sheet_id: 'events-2',
        label: 'Sheet 1',
        events: [{ text: 'call(a)', address: '0' }],
        patterns: [],
      })),
    };

    await app.filterEventTrace('call(a)');

    expect(app.api.executeAction).toHaveBeenCalledWith('events_filter', {
      sheet_id: 'events-1',
      pattern: 'call(a)',
    });
    expect(document.querySelector('.sheet-tab[data-sheet="events-2"]').textContent).toContain('Sheet 1');
    expect(document.querySelector('#events-2 [data-event-address="0"]').textContent).toContain('call(a)');
  });

  it('finds from the selected anchor and uncovers nested matches', async () => {
    const app = makeEventApp();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [
        { text: 'root(a)', address: '0', subs: [{ text: 'target(a)', address: '0/0' }] },
        { text: 'other(a)', address: '1' },
      ],
      patterns: [],
    });
    app.selectEventTraceRow('events-1', '0');
    app.api = {
      executeAction: vi.fn(async () => ({ address: '0/0', event: { text: 'target(a)', address: '0/0' } })),
    };

    await app.findEventTrace('target', false);

    expect(app.api.executeAction).toHaveBeenCalledWith('events_find', {
      sheet_id: 'events-1',
      pattern: 'target',
      reverse: false,
      anchor: '0',
    });
    expect(document.querySelector('[data-event-address="0/0"]').classList.contains('selected')).toBe(true);
  });

  it('loads a raw event trace through the backend parser', async () => {
    const app = makeEventApp();
    app.api = {
      executeAction: vi.fn(async () => ({
        sheet_id: 'events-raw',
        label: 'Sheet 0',
        events: [{ text: 'root(a)', address: '0' }],
        patterns: [],
      })),
    };
    const file = new File(['root(a);'], 'trace.iev', { type: 'text/plain' });

    await app.loadEventTraceFile(file);

    expect(app.api.executeAction).toHaveBeenCalledWith('events_parse', {
      content: 'root(a);',
      filename: 'trace.iev',
    });
    expect(document.querySelector('.sheet-tab[data-sheet="events-raw"]').textContent).toContain('trace.iev');
  });

  it('keeps event pattern state backend-authoritative on add failure and success', async () => {
    const app = makeEventApp();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [],
      patterns: ['old'],
    });
    app.api = {
      executeAction: vi.fn(async () => {
        throw new Error('syntax error');
      }),
    };

    await expect(app.addEventPattern('events-1', 'call(')).rejects.toThrow('syntax error');
    expect(app.sheets['events-1'].patterns).toEqual(['old']);
    expect(document.querySelector('.event-pattern-list').textContent).toContain('old');

    app.api.executeAction = vi.fn(async () => ({ patterns: ['old', 'call(a)'] }));
    await app.addEventPattern('events-1', 'call(a)');
    expect(app.sheets['events-1'].patterns).toEqual(['old', 'call(a)']);
  });

  it('keeps event pattern state unchanged when remove/load fail', async () => {
    const app = makeEventApp();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [],
      patterns: ['first', 'second'],
    });
    const select = document.querySelector('.event-pattern-list');
    select.selectedIndex = 1;
    app.api = {
      executeAction: vi.fn(async () => {
        throw new Error('backend refused');
      }),
    };

    await expect(app.removeSelectedEventPattern('events-1')).rejects.toThrow('backend refused');
    expect(app.sheets['events-1'].patterns).toEqual(['first', 'second']);

    await expect(app.loadEventPatterns('events-1', 'broken(')).rejects.toThrow('backend refused');
    expect(app.sheets['events-1'].patterns).toEqual(['first', 'second']);
  });

  it('uses backend content when saving event patterns', async () => {
    const app = makeEventApp();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [],
      patterns: ['local'],
    });
    app.api = {
      executeAction: vi.fn(async () => ({ content: 'server\n' })),
    };
    app.downloadTextFile = vi.fn();

    const content = await app.saveEventPatterns('events-1');

    expect(content).toBe('server\n');
    expect(app.downloadTextFile).toHaveBeenCalledWith('event_patterns.pats', 'server\n', 'text/plain');
  });

  it('surfaces backend event filter errors without opening a sheet', async () => {
    const app = makeEventApp();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [],
      patterns: [],
    });
    app.api = {
      executeAction: vi.fn(async () => {
        throw new Error('syntax error');
      }),
    };

    const result = await app.filterEventTrace('call(');

    expect(result).toBeNull();
    expect(app.controls.lastStatus).toEqual({ message: 'Filter failed: syntax error', kind: 'error' });
    expect(document.querySelector('.sheet-tab[data-sheet="events-2"]')).toBeNull();
  });

  it('selects event rows whose addresses would be unsafe CSS selectors', () => {
    const app = makeEventApp();
    app.openEventTraceSheet('Trace', {
      sheet_id: 'events-1',
      events: [{ text: 'odd(a)', address: '0"]', subs: [] }],
      patterns: [],
    });

    expect(() => app.selectEventTraceRow('events-1', '0"]')).not.toThrow();
    const row = Array.from(document.querySelectorAll('.event-row')).find(
      (el) => el.getAttribute('data-event-address') === '0"]',
    );
    expect(row.classList.contains('selected')).toBe(true);
  });

  it('rejects invalid event sheet ids before they reach selectors or HTML', () => {
    const app = makeEventApp();

    expect(() => app.openEventTraceSheet('Trace', {
      sheet_id: 'events"bad',
      events: [],
      patterns: [],
    })).toThrow(/invalid sheet id/);
  });

  it('updates an existing event tab label when replacing sheet data', () => {
    const app = makeEventApp();
    app.openEventTraceSheet('Old Trace', {
      sheet_id: 'events-1',
      events: [],
      patterns: [],
    });

    app.openEventTraceSheet('New Trace', {
      sheet_id: 'events-1',
      events: [{ text: 'new(a)', address: '0' }],
      patterns: [],
    }, 'events-1');

    expect(document.querySelector('.sheet-tab[data-sheet="events-1"]').textContent).toContain('New Trace');
    expect(document.querySelector('#events-1 [data-event-address="0"]').textContent).toContain('new(a)');
  });
});
