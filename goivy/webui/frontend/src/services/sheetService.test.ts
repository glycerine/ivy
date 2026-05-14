import { describe, expect, it, vi } from 'vitest';
import {
  assertValidSheetId,
  isValidSheetId,
  isVisualOnlySheet,
  setVisualOnlySheet,
  sheetExists,
  switchSheet,
  visualOnlyMessage,
} from './sheetService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';

describe('sheetService', () => {
  it('validates sheet ids and visual-only state', () => {
    expect(isValidSheetId('sheet-1')).toBe(true);
    expect(isValidSheetId('1-sheet')).toBe(false);
    expect(() => assertValidSheetId('1-sheet')).toThrow('invalid sheet id');

    const app = { sheets: { s1: { visualOnly: false } } };
    setVisualOnlySheet(app, 's1', true);
    expect(isVisualOnlySheet(app, 's1')).toBe(true);
    expect(visualOnlyMessage('events')).toContain('event trace');
  });

  it('detects sheets by state, DOM content, or tabs', () => {
    document.body.innerHTML = [
      '<button class="sheet-tab" data-sheet="tabbed"></button>',
      '<div id="dom-sheet"></div>',
    ].join('');
    const app = { sheets: { stateSheet: {} } };

    expect(sheetExists(app, 'stateSheet', document)).toBe(true);
    expect(sheetExists(app, 'dom-sheet', document)).toBe(true);
    expect(sheetExists(app, 'tabbed', document)).toBe(true);
  });

  it('switches active analysis sheets and resizes graphs', () => {
    document.body.innerHTML = [
      '<button class="sheet-tab active" data-sheet="sheet-1"></button>',
      '<button class="sheet-tab" data-sheet="sheet-2"></button>',
      '<div id="sheet-1" class="sheet-content active"></div>',
      '<div id="sheet-2" class="sheet-content"></div>',
    ].join('');
    const app = {
      activeSheetId: 'sheet-1',
      sheetTab(id) {
        return document.querySelector(`[data-sheet="${id}"]`);
      },
      sheets: {
        'sheet-2': {
          type: 'analysis',
          argGraph: { resize: vi.fn() },
          conceptGraph: { resize: vi.fn() },
          selectedArgNode: 'n2',
        },
      },
    };

    switchSheet(app, 'sheet-2', { bridge: null, doc: document });

    expect(app.activeSheetId).toBe('sheet-2');
    expect(document.getElementById('sheet-2').classList.contains('active')).toBe(true);
    expect(app.selectedArgNode).toBe('n2');
    expect(app.argGraph.resize).toHaveBeenCalled();
  });

  it('switches event sheets without touching analysis graphs or changing model type', () => {
    document.body.innerHTML = [
      '<button class="sheet-tab active" data-sheet="sheet-1"></button>',
      '<button class="sheet-tab" data-sheet="events-1"></button>',
      '<div id="sheet-1" class="sheet-content active"></div>',
      '<div id="events-1" class="sheet-content"></div>',
    ].join('');
    const uiDataModel = new UIDataModel();
    const analysisArg = { resize: vi.fn() };
    const analysisConcept = { resize: vi.fn() };
    const app = {
      activeSheetId: 'sheet-1',
      argGraph: analysisArg,
      conceptGraph: analysisConcept,
      uiDataModel,
      sheetTab(id) {
        return document.querySelector(`[data-sheet="${id}"]`);
      },
      sheets: {
        'sheet-1': { type: 'analysis', argGraph: analysisArg, conceptGraph: analysisConcept },
        'events-1': { type: 'events' },
      },
    };

    switchSheet(app, 'events-1', { doc: document });

    expect(app.activeSheetId).toBe('events-1');
    expect(app.argGraph).toBeNull();
    expect(app.conceptGraph).toBeNull();
    expect(analysisArg.resize).not.toHaveBeenCalled();
    expect(analysisConcept.resize).not.toHaveBeenCalled();
    expect(app.uiDataModel.sheets['events-1'].type).toBe('events');
  });
});
