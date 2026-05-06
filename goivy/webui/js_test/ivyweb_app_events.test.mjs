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
});
