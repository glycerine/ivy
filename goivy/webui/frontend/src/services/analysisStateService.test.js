import { describe, expect, it, vi } from 'vitest';
import {
  analysisStateLimits,
  buildAnalysisState,
  removeAnalysisStateExtraSheets,
  validateAnalysisStateObject,
} from './analysisStateService.js';

describe('analysisStateService', () => {
  it('builds a serializable analysis state from explicit app dependencies', () => {
    const app = {
      sheets: {
        'sheet-1': {
          type: 'analysis',
          selectedArgNode: 'n0',
          argGraph: { cy: { json: () => ({ elements: ['arg'] }) } },
          conceptGraph: { cy: { json: () => ({ elements: ['concept'] }) } },
        },
      },
      tabLabelForSheet: vi.fn(() => 'ARG'),
      graphElementsSnapshot: (graph) => graph.cy.json().elements,
      _persistedFileName: 'client.ivy',
      _persistedFilePath: '/tmp/client.ivy',
      _editorContent: () => 'ivy',
      getMode: () => 'pdr',
      activeSheetId: 'sheet-1',
      selectedArgNode: 'n0',
      _edgeVisibility: { link: { all_to_all: true } },
      _labelVisibility: {},
    };
    const persist = {
      _getToggles: vi.fn(() => { return { 'link|all_to_all': true }; }),
    };

    expect(buildAnalysisState(app, persist)).toMatchObject({
      analysis_state_format: 'ivyweb-json',
      fileName: 'client.ivy',
      fileContent: 'ivy',
      sheets: [
        {
          id: 'sheet-1',
          arg: { elements: ['arg'] },
          concept: { elements: ['concept'] },
        },
      ],
    });
  });

  it('validates sheet shape and limits', () => {
    const app = {
      analysisStateLimits,
      isValidSheetId: (id) => /^[A-Za-z]/.test(id),
    };

    expect(() => validateAnalysisStateObject(app, {
      analysis_state_format: 'ivyweb-json',
      activeSheetId: 'sheet-1',
      sheets: [{ id: 'sheet-1', type: 'analysis', arg: { elements: [] }, concept: { elements: [] } }],
    })).not.toThrow();

    expect(() => validateAnalysisStateObject(app, {
      analysis_state_format: 'ivyweb-json',
      sheets: [{ id: '1-bad', type: 'analysis' }],
    })).toThrow('invalid sheet id');
  });

  it('removes every restored sheet except the root analysis sheet', () => {
    const app = {
      sheets: { 'sheet-1': {}, 'sheet-2': {}, events: {} },
      removeSheet: vi.fn(),
    };

    removeAnalysisStateExtraSheets(app);

    expect(app.removeSheet).toHaveBeenCalledWith('sheet-2');
    expect(app.removeSheet).toHaveBeenCalledWith('events');
  });
});
