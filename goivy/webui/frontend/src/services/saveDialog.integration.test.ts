import { afterEach, describe, expect, it, vi } from 'vitest';
import { exportConjecture } from './analysisActionService.ts';
import { saveAnalysisState } from './analysisStateService.ts';
import { saveEventPatterns } from './eventTraceService.ts';
import { saveModelAs } from './fileService.ts';
import {
  IvyRuntime,
  configureIvyRuntimeDependencies,
  resetIvyRuntimeDependencies,
} from './ivyRuntime.ts';
import { FakeAPI, FakeControls, FakeGraph, makePersist, makeWritableHandle } from '../test/fakes.ts';

function pickerHandle(name = 'saved.txt') {
  return makeWritableHandle({ name });
}

function pickerFor(handle) {
  return {
    showSaveFilePicker: vi.fn(async () => handle),
  };
}

function makeRuntime(overrides = {}) {
  resetIvyRuntimeDependencies();
  configureIvyRuntimeDependencies({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
  const runtime = new IvyRuntime();
  Object.assign(runtime, overrides);
  return runtime;
}

afterEach(() => {
  resetIvyRuntimeDependencies();
  delete (window as any).showSaveFilePicker;
  document.body.innerHTML = '';
});

describe('save dialog metadata', () => {
  it('uses explicit picker descriptors for model, analysis state, invariant, abstraction, DOT, and event patterns', async () => {
    const modelHandle = pickerHandle('client.ivy');
    const modelWin = pickerFor(modelHandle);
    const modelPersist = makePersist({ saveFileHandle: vi.fn(async () => undefined), setFileName: vi.fn() });
    const modelApp = {
      _persistedFileName: 'client.ivy',
      _persistedFilePath: 'client.ivy',
      _persistedFileContent: 'old',
      _savedFileContent: 'old',
      _fileHandle: null,
      _editorContent: vi.fn(() => 'new'),
      _updateEditorLabel: vi.fn(),
      _showSaveProgress: vi.fn(() => null),
      _hideSaveProgress: vi.fn(),
      downloadModelForUnsupportedSave: vi.fn(),
      showSaveAsExplanationNotice: vi.fn(),
      hideSaveAsExplanationNotice: vi.fn(),
      controls: { setStatus: vi.fn() },
    };
    await saveModelAs(modelApp, modelPersist, { win: modelWin });

    expect(modelWin.showSaveFilePicker).toHaveBeenCalledWith({
      suggestedName: 'client.ivy',
      types: [{ description: 'Ivy model files', accept: { 'text/plain': ['.ivy'] } }],
    });

    const state = { analysis_state_format: 'ivyweb-json', sheets: [] };
    const analysisHandle = pickerHandle('client.ivyweb.json');
    const analysisWin = pickerFor(analysisHandle);
    await saveAnalysisState({
      _persistedFileName: 'client.ivy',
      buildAnalysisState: vi.fn(() => state),
      downloadTextFile: vi.fn(),
      controls: { setStatus: vi.fn() },
    }, { win: analysisWin });

    expect(analysisWin.showSaveFilePicker).toHaveBeenCalledWith({
      suggestedName: 'client.ivyweb.json',
      types: [{ description: 'IvyWeb analysis state files', accept: { 'application/json': ['.ivyweb.json', '.json'] } }],
    });

    const invariantHandle = pickerHandle('client_invariant.ivy');
    const invariantPicker = vi.fn(async () => invariantHandle);
    Object.defineProperty(window, 'showSaveFilePicker', { configurable: true, value: invariantPicker });
    const invariantRuntime = makeRuntime({
      _persistedFileName: 'client.ivy',
      api: { executeAction: vi.fn(async () => ({ content: 'invariant c1 true\n' })) },
    });
    await invariantRuntime.saveInvariant();

    expect(invariantPicker).toHaveBeenCalledWith({
      suggestedName: 'client_invariant.ivy',
      types: [{ description: 'Ivy invariant files', accept: { 'text/plain': ['.ivy'] } }],
    });

    const abstractionHandle = pickerHandle('client_abstraction.ivy');
    const abstractionPicker = vi.fn(async () => abstractionHandle);
    Object.defineProperty(window, 'showSaveFilePicker', { configurable: true, value: abstractionPicker });
    const abstractionRuntime = makeRuntime({
      _persistedFileName: 'client.ivy',
      api: { executeAction: vi.fn(async () => ({ content: 'relation r\n' })) },
    });
    await abstractionRuntime.saveAbstraction();

    expect(abstractionPicker).toHaveBeenCalledWith({
      suggestedName: 'client_abstraction.ivy',
      types: [{ description: 'Ivy abstraction files', accept: { 'text/plain': ['.ivy'] } }],
    });

    const dotHandle = pickerHandle('concept_graph.dot');
    const dotWin = pickerFor(dotHandle);
    await exportConjecture({
      activeSheetId: 'sheet-1',
      api: { executeAction: vi.fn(async () => ({ content: 'digraph G {}\n', filename: 'concept_graph.dot' })) },
      downloadTextFile: vi.fn(),
      controls: { setStatus: vi.fn() },
    }, { win: dotWin, doc: document });

    expect(dotWin.showSaveFilePicker).toHaveBeenCalledWith({
      suggestedName: 'concept_graph.dot',
      types: [{ description: 'Graphviz DOT files', accept: { 'text/vnd.graphviz': ['.dot'] } }],
    });

    const eventHandle = pickerHandle('event_patterns.pats');
    const eventWin = pickerFor(eventHandle);
    await saveEventPatterns({
      sheets: { events: { id: 'events', patterns: ['call(a)'] } },
      api: { executeAction: vi.fn(async () => ({ content: 'call(a)\n' })) },
      downloadTextFile: vi.fn(),
    }, 'events', { win: eventWin });

    expect(eventWin.showSaveFilePicker).toHaveBeenCalledWith({
      suggestedName: 'event_patterns.pats',
      types: [{ description: 'Ivy event pattern files', accept: { 'text/plain': ['.pats'] } }],
    });
  });

  it('keeps explicit fallback filenames and MIME types when picker support is unavailable', async () => {
    const state = { analysis_state_format: 'ivyweb-json', sheets: [] };
    const analysisApp = {
      _persistedFileName: 'client.ivy',
      buildAnalysisState: vi.fn(() => state),
      downloadTextFile: vi.fn(),
      controls: { setStatus: vi.fn() },
    };
    await saveAnalysisState(analysisApp, { win: {} });
    expect(analysisApp.downloadTextFile).toHaveBeenCalledWith(
      'client.ivyweb.json',
      `${JSON.stringify(state, null, 2)}\n`,
      'application/json',
    );

    const invariantRuntime = makeRuntime({
      _persistedFileName: 'client.ivy',
      api: { executeAction: vi.fn(async () => ({ content: 'invariant c1 true\n' })) },
      downloadTextFile: vi.fn(),
    });
    await invariantRuntime.saveInvariant();
    expect(invariantRuntime.downloadTextFile).toHaveBeenCalledWith('client_invariant.ivy', 'invariant c1 true\n', 'text/plain');

    const abstractionRuntime = makeRuntime({
      _persistedFileName: 'client.ivy',
      api: { executeAction: vi.fn(async () => ({ content: 'relation r\n' })) },
      downloadTextFile: vi.fn(),
    });
    await abstractionRuntime.saveAbstraction();
    expect(abstractionRuntime.downloadTextFile).toHaveBeenCalledWith('client_abstraction.ivy', 'relation r\n', 'text/plain');

    const exportApp = {
      activeSheetId: 'sheet-1',
      api: { executeAction: vi.fn(async () => ({ content: 'digraph G {}\n', filename: 'concept_graph.dot' })) },
      downloadTextFile: vi.fn(),
      controls: { setStatus: vi.fn() },
    };
    await exportConjecture(exportApp, { win: {}, doc: document });
    expect(exportApp.downloadTextFile).toHaveBeenCalledWith('concept_graph.dot', 'digraph G {}\n', 'text/vnd.graphviz');

    const eventApp = {
      sheets: { events: { id: 'events', patterns: ['call(a)'] } },
      api: { executeAction: vi.fn(async () => ({ content: 'call(a)\n' })) },
      downloadTextFile: vi.fn(),
    };
    await saveEventPatterns(eventApp, 'events', { win: {} });
    expect(eventApp.downloadTextFile).toHaveBeenCalledWith('event_patterns.pats', 'call(a)\n', 'text/plain');
  });
});
