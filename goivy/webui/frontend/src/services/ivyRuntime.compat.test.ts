import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  IvyRuntime,
  configureIvyRuntimeDependencies,
  resetIvyRuntimeDependencies,
} from './ivyRuntime.ts';
import { FakeAPI, FakeControls, FakeGraph, makePersist } from '../test/fakes.ts';

function makeRuntime(overrides = {}) {
  resetIvyRuntimeDependencies();
  configureIvyRuntimeDependencies({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
    ...overrides,
  });
  return new IvyRuntime();
}

function installSheetDom() {
  document.body.innerHTML = [
    '<div id="sheet-area">',
    '  <div id="tab-bar"><button class="sheet-tab active" data-sheet="sheet-1"><span>Sheet 1</span></button></div>',
    '  <div id="sheet-1" class="sheet-content active">',
    '    <div class="sheet-columns">',
    '      <div class="sheet-left">',
    '        <div class="sheet-main">',
    '          <div id="arg-panel" class="panel">',
    '            <div class="panel-header">',
    '              <div class="panel-header-actions">',
    '                <div class="dropdown ui-mode-only ui-mode-reachability">',
    '                  <span class="panel-menu" data-dropdown="arg-action-menu">Action</span>',
    '                  <div id="arg-action-menu" class="dropdown-content">',
    '                    <a href="#" id="arg-recalculate-all">Recalculate all</a>',
    '                    <a href="#" id="arg-show-reachable">Show reachable states</a>',
    '                  </div>',
    '                </div>',
    '              </div>',
    '            </div>',
    '            <div id="arg-graph" class="graph-container"></div>',
    '          </div>',
    '          <div id="divider" class="divider"></div>',
    '          <div id="concept-panel" class="panel"><div id="concept-graph" class="graph-container"></div></div>',
    '        </div>',
    '        <div id="info-panel" class="info-panel">',
    '          <div id="info-header" class="info-header" title="Drag to resize Details">Details</div>',
    '          <div id="info-content"></div>',
    '        </div>',
    '      </div>',
    '      <div id="divider2" class="divider"></div>',
    '      <div id="state-panel" class="panel"></div>',
    '    </div>',
    '  </div>',
    '</div>',
  ].join('');
}

function installStaticMenuDom() {
  document.body.innerHTML = [
    '<div id="menubar">',
    '  <div class="menu-group">',
    '    <div class="dropdown">',
    '      <span class="panel-menu" data-dropdown="file-menu">File</span>',
    '      <div id="file-menu" class="dropdown-content">',
    '        <a href="#" id="file-load">Open .ivy spec...</a>',
    '      </div>',
    '    </div>',
    '  </div>',
    '</div>',
    '<div id="sheet-1" class="sheet-content active">',
    '  <div id="arg-panel" class="panel">',
    '    <div class="panel-header">',
    '      <strong class="column-title">Reachability Graph</strong>',
    '      <div class="panel-header-actions">',
    '        <div class="dropdown ui-mode-only ui-mode-cti">',
    '          <span class="panel-menu" data-dropdown="arg-inv-menu">Invariant</span>',
    '          <div id="arg-inv-menu" class="dropdown-content">',
    '            <a href="#" id="arg-check-induction">Check induction</a>',
    '          </div>',
    '        </div>',
    '        <div class="dropdown ui-mode-only ui-mode-reachability">',
    '          <span class="panel-menu" data-dropdown="arg-action-menu">Action</span>',
    '          <div id="arg-action-menu" class="dropdown-content">',
    '            <a href="#" id="arg-recalculate-all">Recalculate all</a>',
    '          </div>',
    '        </div>',
    '      </div>',
    '    </div>',
    '  </div>',
    '  <div id="concept-panel" class="panel">',
    '    <div class="panel-header">',
    '      <strong class="column-title">Concept graph</strong>',
    '      <div class="panel-header-actions">',
    '        <div class="dropdown ui-mode-only ui-mode-cti">',
    '          <span class="panel-menu" data-dropdown="conj-menu">Conjecture</span>',
    '          <div id="conj-menu" class="dropdown-content">',
    '            <a href="#" id="conj-undo">Undo</a>',
    '          </div>',
    '        </div>',
    '        <div class="dropdown ui-mode-only ui-mode-reachability">',
    '          <span class="panel-menu" data-dropdown="reach-action-menu">Action</span>',
    '          <div id="reach-action-menu" class="dropdown-content">',
    '            <a href="#" id="conj-reach-undo">Undo</a>',
    '          </div>',
    '        </div>',
    '        <div class="dropdown">',
    '          <span class="panel-menu" data-dropdown="view-menu">View</span>',
    '          <div id="view-menu" class="dropdown-content">',
    '            <a href="#" id="view-add-relation">Add relation</a>',
    '          </div>',
    '        </div>',
    '      </div>',
    '    </div>',
    '  </div>',
    '</div>',
  ].join('');
}

function visibleMenuLabels(scope: Element | Document, mode: 'cti' | 'reachability') {
  return Array.from(scope.querySelectorAll('.panel-menu'))
    .filter((el) => {
      const modeOnly = el.closest('.ui-mode-only');
      return !modeOnly || modeOnly.classList.contains(`ui-mode-${mode}`);
    })
    .map((el) => (el.textContent || '').trim())
    .filter(Boolean);
}

function duplicateMenuLabels(labels: string[]) {
  const counts = new Map<string, number>();
  for (const label of labels) {
    counts.set(label, (counts.get(label) || 0) + 1);
  }
  return Array.from(counts)
    .filter(([, count]) => count > 1)
    .map(([label]) => label)
    .sort();
}

function duplicatedVisibleMenuLabels(scope: Element | Document, mode: 'cti' | 'reachability') {
  return duplicateMenuLabels(visibleMenuLabels(scope, mode));
}

function setOffsetWidth(element: Element, width: number) {
  Object.defineProperty(element, 'offsetWidth', {
    configurable: true,
    value: width,
  });
}

function setOffsetHeight(element: Element, height: number) {
  Object.defineProperty(element, 'offsetHeight', {
    configurable: true,
    value: height,
  });
}

afterEach(() => {
  resetIvyRuntimeDependencies();
  document.body.innerHTML = '';
  document.documentElement.style.removeProperty('--ivy-graph-background');
  document.documentElement.style.removeProperty('--ivy-relation-name-color');
  vi.useRealTimers();
});

describe('ivyRuntime compatibility behavior', () => {
  it('uses one divider path for the Concept and State boundary', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="sheet-workspace" class="sheet-workspace">',
      '    <div id="sheet-pages" class="sheet-pages">',
      '      <div id="sheet-1" class="sheet-content active">',
      '        <div class="sheet-columns">',
      '          <div id="arg-panel" class="sheet-pane"></div>',
      '          <div id="divider" class="divider" data-resize-target="previous"></div>',
      '          <div id="concept-panel" class="sheet-pane" style="min-width: 120px"></div>',
      '          <div id="divider2" class="divider" data-resize-target="previous"></div>',
      '          <div id="state-panel" class="sheet-pane" style="min-width: 120px"></div>',
      '        </div>',
      '      </div>',
      '    </div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.argGraph = new FakeGraph();
    runtime.conceptGraph = new FakeGraph();
    runtime._refreshEditorLayout = vi.fn();
    const columns = document.querySelector('.sheet-columns')!;
    const concept = document.getElementById('concept-panel')!;
    const state = document.getElementById('state-panel')!;
    const divider = document.getElementById('divider2')!;
    setOffsetWidth(columns, 1000);
    setOffsetWidth(concept, 260);
    setOffsetWidth(state, 220);

    runtime.setupResizer();
    divider.dispatchEvent(new MouseEvent('mousedown', { clientX: 300, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientX: 250, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(concept.style.width).toBe('210px');
    expect(concept.style.flex).toBe('0 0 210px');
    expect(state.style.width).toBe('');
    expect(state.style.flex).toBe('');
  });

  it('moves the editor boundary by resizing the rightmost active sheet panes', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="sheet-workspace" class="sheet-workspace">',
      '    <div id="sheet-pages" class="sheet-pages">',
      '      <div id="sheet-1" class="sheet-content active has-analysis-history">',
      '        <div class="proof-crg-pane">',
      '          <div class="proof-goal-column" data-resizable-pane="proof-goals"></div>',
      '          <div class="divider proof-crg-column-divider" data-resize-target="previous"></div>',
      '          <div class="crg-column" data-resizable-pane="crg-transition"></div>',
      '        </div>',
      '        <div class="sheet-columns">',
      '          <div id="arg-panel" class="sheet-pane"></div>',
      '          <div id="divider" class="divider" data-resize-target="previous"></div>',
      '          <div id="concept-panel" class="sheet-pane"></div>',
      '          <div id="divider2" class="divider" data-resize-target="previous"></div>',
      '          <div id="state-panel" class="sheet-pane"></div>',
      '        </div>',
      '      </div>',
      '    </div>',
      '    <div id="editor-left-resize-handle"></div>',
      '    <div id="editor-panel" class="sheet-pane" style="min-width: 120px"></div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.argGraph = new FakeGraph();
    runtime.conceptGraph = new FakeGraph();
    runtime._refreshEditorLayout = vi.fn();
    const workspace = document.getElementById('sheet-workspace')!;
    const editor = document.getElementById('editor-panel')!;
    const divider = document.getElementById('editor-left-resize-handle')!;
    const state = document.getElementById('state-panel')!;
    const crg = document.querySelector('.crg-column')!;
    const log = vi.spyOn(console, 'log').mockImplementation(() => {});
    setOffsetWidth(workspace, 1200);
    setOffsetWidth(editor, 420);
    setOffsetWidth(state, 220);
    setOffsetWidth(crg, 260);

    runtime.setupResizer();
    divider.dispatchEvent(new MouseEvent('mousedown', { clientX: 600, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientX: 550, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(state.style.width).toBe('170px');
    expect(state.style.flex).toBe('0 0 170px');
    expect(crg.style.width).toBe('210px');
    expect(crg.style.flex).toBe('0 0 210px');
    expect(editor.style.width).toBe('');
    expect(editor.style.flex).toBe('');
    expect(log).toHaveBeenCalledTimes(1);
    log.mockRestore();
  });

  it('does not shrink CRG / transition when State/relations is already at minimum width', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="sheet-workspace" class="sheet-workspace">',
      '    <div id="sheet-pages" class="sheet-pages">',
      '      <div id="sheet-1" class="sheet-content active has-analysis-history">',
      '        <div class="proof-crg-pane">',
      '          <div class="proof-goal-column" data-resizable-pane="proof-goals"></div>',
      '          <div class="divider proof-crg-column-divider" data-resize-target="previous"></div>',
      '          <div class="crg-column" data-resizable-pane="crg-transition"></div>',
      '        </div>',
      '        <div class="sheet-columns">',
      '          <div id="arg-panel" class="sheet-pane"></div>',
      '          <div id="divider" class="divider" data-resize-target="previous"></div>',
      '          <div id="concept-panel" class="sheet-pane"></div>',
      '          <div id="divider2" class="divider" data-resize-target="previous"></div>',
      '          <div id="state-panel" class="sheet-pane" style="min-width: 220px"></div>',
      '        </div>',
      '      </div>',
      '    </div>',
      '    <div id="editor-left-resize-handle"></div>',
      '    <div id="editor-panel" class="sheet-pane" style="min-width: 120px"></div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.argGraph = new FakeGraph();
    runtime.conceptGraph = new FakeGraph();
    runtime._refreshEditorLayout = vi.fn();
    const workspace = document.getElementById('sheet-workspace')!;
    const editor = document.getElementById('editor-panel')!;
    const divider = document.getElementById('editor-left-resize-handle')!;
    const state = document.getElementById('state-panel')!;
    const crg = document.querySelector('.crg-column')!;
    const log = vi.spyOn(console, 'log').mockImplementation(() => {});
    setOffsetWidth(workspace, 1200);
    setOffsetWidth(editor, 420);
    setOffsetWidth(state, 220);
    setOffsetWidth(crg, 260);

    runtime.setupResizer();
    divider.dispatchEvent(new MouseEvent('mousedown', { clientX: 600, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientX: 550, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(state.style.width).toBe('220px');
    expect(state.style.flex).toBe('0 0 220px');
    expect(crg.style.width).toBe('260px');
    expect(crg.style.flex).toBe('0 0 260px');
    expect(log).toHaveBeenCalledTimes(1);
    log.mockRestore();
  });

  it('does not shrink State/relations when CRG / transition is already at minimum width', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="sheet-workspace" class="sheet-workspace">',
      '    <div id="sheet-pages" class="sheet-pages">',
      '      <div id="sheet-1" class="sheet-content active has-analysis-history">',
      '        <div class="proof-crg-pane">',
      '          <div class="proof-goal-column" data-resizable-pane="proof-goals"></div>',
      '          <div class="divider proof-crg-column-divider" data-resize-target="previous"></div>',
      '          <div class="crg-column" data-resizable-pane="crg-transition" style="min-width: 260px"></div>',
      '        </div>',
      '        <div class="sheet-columns">',
      '          <div id="arg-panel" class="sheet-pane"></div>',
      '          <div id="divider" class="divider" data-resize-target="previous"></div>',
      '          <div id="concept-panel" class="sheet-pane"></div>',
      '          <div id="divider2" class="divider" data-resize-target="previous"></div>',
      '          <div id="state-panel" class="sheet-pane" style="min-width: 220px"></div>',
      '        </div>',
      '      </div>',
      '    </div>',
      '    <div id="editor-left-resize-handle"></div>',
      '    <div id="editor-panel" class="sheet-pane" style="min-width: 120px"></div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.argGraph = new FakeGraph();
    runtime.conceptGraph = new FakeGraph();
    runtime._refreshEditorLayout = vi.fn();
    const workspace = document.getElementById('sheet-workspace')!;
    const editor = document.getElementById('editor-panel')!;
    const divider = document.getElementById('editor-left-resize-handle')!;
    const state = document.getElementById('state-panel')!;
    const crg = document.querySelector('.crg-column')!;
    const log = vi.spyOn(console, 'log').mockImplementation(() => {});
    setOffsetWidth(workspace, 1200);
    setOffsetWidth(editor, 420);
    setOffsetWidth(state, 260);
    setOffsetWidth(crg, 260);

    runtime.setupResizer();
    divider.dispatchEvent(new MouseEvent('mousedown', { clientX: 600, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientX: 550, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(state.style.width).toBe('260px');
    expect(state.style.flex).toBe('0 0 260px');
    expect(crg.style.width).toBe('260px');
    expect(crg.style.flex).toBe('0 0 260px');
    expect(log).toHaveBeenCalledTimes(1);
    log.mockRestore();
  });

  it('keeps the editor width fixed while logging geometry after left-edge drags', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="sheet-workspace" class="sheet-workspace">',
      '    <div id="sheet-pages" class="sheet-pages">',
      '      <div id="sheet-1" class="sheet-content active">',
      '        <div class="sheet-columns">',
      '          <div id="state-panel" class="sheet-pane"></div>',
      '        </div>',
      '      </div>',
      '    </div>',
      '    <div id="editor-left-resize-handle"></div>',
      '    <div id="editor-panel" class="sheet-pane" style="min-width: 120px">',
      '      <div class="panel-header"></div>',
      '    </div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.argGraph = new FakeGraph();
    runtime.conceptGraph = new FakeGraph();
    runtime._refreshEditorLayout = vi.fn();
    const workspace = document.getElementById('sheet-workspace')!;
    const editor = document.getElementById('editor-panel')!;
    const state = document.getElementById('state-panel')!;
    const handle = document.getElementById('editor-left-resize-handle')!;
    const log = vi.spyOn(console, 'log').mockImplementation(() => {});
    setOffsetWidth(workspace, 1200);
    setOffsetWidth(editor, 420);
    setOffsetWidth(state, 220);
    Object.defineProperty(editor, 'getBoundingClientRect', {
      configurable: true,
      value: () => {
        const left = 400 + (Number.parseFloat(state.style.width) || state.offsetWidth);
        const width = editor.offsetWidth;
        return { left, right: left + width, top: 0, bottom: 600, width, height: 600 };
      },
    });

    runtime.setupResizer();
    handle.dispatchEvent(new MouseEvent('mousedown', { clientX: 602, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientX: 552, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(state.style.width).toBe('170px');
    expect(state.style.flex).toBe('0 0 170px');
    expect(editor.style.width).toBe('');
    expect(editor.style.flex).toBe('');
    expect(log).toHaveBeenLastCalledWith(
      '[ivyweb editor-left-resize-handle] editor pane geometry',
      expect.objectContaining({ left: 570, right: 990, width: 420, height: 600, styleWidth: '', flex: '' }),
    );

    setOffsetWidth(state, 170);
    handle.dispatchEvent(new MouseEvent('mousedown', { clientX: 602, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientX: 642, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(state.style.width).toBe('210px');
    expect(state.style.flex).toBe('0 0 210px');
    expect(editor.style.width).toBe('');
    expect(editor.style.flex).toBe('');
    expect(log).toHaveBeenCalledTimes(2);
    expect(log).toHaveBeenLastCalledWith(
      '[ivyweb editor-left-resize-handle] editor pane geometry',
      expect.objectContaining({ left: 610, right: 1030, width: 420, height: 600, styleWidth: '', flex: '' }),
    );
    log.mockRestore();
  });

  it('adds the editor resize handle when older markup lacks it', () => {
    document.body.innerHTML = '<div id="editor-panel"><div class="panel-header"></div></div>';
    const runtime = makeRuntime();

    const handle = runtime._ensureEditorResizeHandle();

    expect(handle).not.toBeNull();
    expect(document.querySelector('#editor-left-resize-handle + #editor-panel')).toBe(document.getElementById('editor-panel'));
    expect(handle.getAttribute('aria-label')).toBe('Resize editor');
  });

  it('normalizes old editor divider markup to one boundary handle', () => {
    document.body.innerHTML = [
      '<div id="sheet-workspace">',
      '  <div id="divider3" class="divider" data-resize-target="next"></div>',
      '  <div id="editor-panel">',
      '    <div id="editor-left-resize-handle"></div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    const editor = document.getElementById('editor-panel')!;

    const handle = runtime._ensureEditorResizeHandle();

    expect(document.getElementById('divider3')).toBeNull();
    expect(document.querySelectorAll('#editor-left-resize-handle')).toHaveLength(1);
    expect(handle.nextElementSibling).toBe(editor);
    expect(editor.querySelector('#editor-left-resize-handle')).toBeNull();
    expect(handle.getAttribute('class')).toBeNull();
    expect(handle.getAttribute('data-resize-target')).toBeNull();
  });

  it('wraps analysis history and Proof/CRG panes in a hidden proof goal wrapper by default', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="sheet-1" class="sheet-content active">',
      '    <div class="sheet-columns"></div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-1';
    runtime.sheets = { 'sheet-1': { id: 'sheet-1' } };

    const history = runtime._ensureAnalysisHistoryControls('sheet-1');
    const pane = runtime._ensureProofGoalPane('sheet-1');
    const wrapper = document.querySelector('.proof-goal-wrapper') as HTMLElement;
    const rowDivider = document.querySelector('.proof-crg-row-divider') as HTMLElement;
    const closeButton = wrapper.querySelector('.proof-goal-close-btn') as HTMLButtonElement;

    expect(wrapper).not.toBeNull();
    expect(wrapper.hidden).toBe(true);
    expect(wrapper.classList.contains('is-visible')).toBe(false);
    expect(wrapper.nextElementSibling).toBe(document.querySelector('.sheet-columns'));
    expect(history.parentElement).toBe(wrapper);
    expect(pane.parentElement).toBe(wrapper);
    expect(rowDivider.parentElement).toBe(wrapper);
    expect(rowDivider.previousElementSibling).toBe(pane);
    expect(closeButton).not.toBeNull();
    expect(closeButton.parentElement?.firstElementChild).toBe(closeButton);
    expect(closeButton.title).toBe('Close proof goal windows');
    expect(closeButton.getAttribute('aria-label')).toBe('Close proof goal windows');

    runtime._setProofGoalWrapperVisible('sheet-1', true);
    expect(wrapper.hidden).toBe(false);
    closeButton.click();
    expect(wrapper.hidden).toBe(true);
    expect(wrapper.classList.contains('is-visible')).toBe(false);
  });

  it('adds and drags the divider between Proof goals and CRG / transition', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="sheet-1" class="sheet-content active">',
      '    <div class="sheet-columns"></div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-1';
    runtime.sheets = { 'sheet-1': { id: 'sheet-1' } };
    runtime.argGraph = new FakeGraph();
    runtime.conceptGraph = new FakeGraph();
    runtime._refreshEditorLayout = vi.fn();

    runtime._ensureProofGoalPane('sheet-1');
    runtime._setProofGoalWrapperVisible('sheet-1', true);
    const pane = document.querySelector('.proof-crg-pane')!;
    const goalColumn = document.querySelector('.proof-goal-column')!;
    const crgColumn = document.querySelector('.crg-column')!;
    const divider = document.querySelector('.proof-crg-column-divider')!;
    setOffsetWidth(pane, 760);
    setOffsetWidth(goalColumn, 280);

    runtime.setupResizer();
    divider.dispatchEvent(new MouseEvent('mousedown', { clientX: 300, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientX: 360, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(goalColumn.getAttribute('data-resizable-pane')).toBe('proof-goals');
    expect(crgColumn.getAttribute('data-resizable-pane')).toBe('crg-transition');
    expect(goalColumn.style.width).toBe('340px');
    expect(goalColumn.style.flex).toBe('0 0 340px');
    expect(runtime.sheets['sheet-1'].proofGraph.resize).toHaveBeenCalled();
  });

  it('adds and drags the divider between Proof/CRG and the graph panes', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="sheet-1" class="sheet-content active">',
      '    <div class="sheet-columns"></div>',
      '  </div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-1';
    runtime.sheets = { 'sheet-1': { id: 'sheet-1' } };
    runtime.argGraph = new FakeGraph();
    runtime.conceptGraph = new FakeGraph();
    runtime._refreshEditorLayout = vi.fn();

    runtime._ensureProofGoalPane('sheet-1');
    const sheet = document.getElementById('sheet-1')!;
    const pane = document.querySelector('.proof-crg-pane')!;
    const divider = document.querySelector('.proof-crg-row-divider')!;
    const columns = document.querySelector('.sheet-columns')!;
    setOffsetHeight(sheet, 720);
    setOffsetHeight(pane, 148);

    runtime._setupSheetRowResizer();
    divider.dispatchEvent(new MouseEvent('mousedown', { clientY: 200, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientY: 250, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(pane.getAttribute('data-resizable-row')).toBe('proof-crg');
    expect(divider.getAttribute('data-resize-axis')).toBe('y');
    expect(columns.parentElement).toBe(sheet);
    expect(pane.style.height).toBe('198px');
    expect(pane.style.flex).toBe('0 0 198px');
    expect(runtime.sheets['sheet-1'].proofGraph.resize).toHaveBeenCalled();
  });

  it('adds draggable Proof/CRG dividers to Reachable states sheets', () => {
    installSheetDom();
    const runtime = makeRuntime();
    runtime.setupResizer();
    runtime._setupSheetRowResizer();

    const sheetId = runtime.addSheet('Reachable states', 'sheet-2', { reachabilityOnly: true });
    const sheet = document.getElementById(sheetId)!;
    const wrapper = sheet.querySelector('.proof-goal-wrapper') as HTMLElement;
    const pane = sheet.querySelector('.proof-crg-pane') as HTMLElement;
    const goalColumn = sheet.querySelector('.proof-goal-column') as HTMLElement;
    const columnDivider = sheet.querySelector('.proof-crg-column-divider') as HTMLElement;
    const rowDivider = sheet.querySelector('.proof-crg-row-divider') as HTMLElement;
    setOffsetWidth(pane, 760);
    setOffsetWidth(goalColumn, 280);
    setOffsetHeight(sheet, 720);
    setOffsetHeight(pane, 148);

    expect(sheet.classList.contains('reachability-only-sheet')).toBe(true);
    expect(sheet.classList.contains('has-analysis-history')).toBe(true);
    expect(wrapper.hidden).toBe(true);
    expect(wrapper.nextElementSibling).toBe(sheet.querySelector('.sheet-columns'));
    expect(rowDivider.previousElementSibling).toBe(pane);
    expect(rowDivider.parentElement).toBe(wrapper);
    expect(columnDivider).not.toBeNull();
    expect(rowDivider).not.toBeNull();
    runtime._setProofGoalWrapperVisible(sheetId, true);

    columnDivider.dispatchEvent(new MouseEvent('mousedown', { clientX: 300, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientX: 360, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(goalColumn.style.width).toBe('340px');
    expect(goalColumn.style.flex).toBe('0 0 340px');

    rowDivider.dispatchEvent(new MouseEvent('mousedown', { clientY: 200, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientY: 250, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(pane.style.height).toBe('198px');
    expect(pane.style.flex).toBe('0 0 198px');
    expect(runtime.sheets[sheetId].proofGraph.resize).toHaveBeenCalled();
  });

  it('maps the graph background slider and flips relation-name text at channel 181', () => {
    document.body.innerHTML = '<input id="graph-background-slider" type="range" min="0" max="100" value="0">';
    const runtime = makeRuntime();
    const slider = document.getElementById('graph-background-slider') as HTMLInputElement;

    runtime._setupGraphBackgroundSlider();
    expect(document.documentElement.style.getPropertyValue('--ivy-graph-background')).toBe('rgb(0, 0, 0)');
    expect(document.documentElement.style.getPropertyValue('--ivy-relation-name-color')).toBe('rgb(255, 255, 255)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-color')).toBe('rgb(0, 0, 0)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-fill')).toBe('0%');
    expect(slider.title).toBe('Graph background: rgb(0, 0, 0)');

    slider.value = '50';
    slider.dispatchEvent(new Event('input'));
    expect(document.documentElement.style.getPropertyValue('--ivy-graph-background')).toBe('rgb(128, 128, 128)');
    expect(document.documentElement.style.getPropertyValue('--ivy-relation-name-color')).toBe('rgb(255, 255, 255)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-color')).toBe('rgb(128, 128, 128)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-fill')).toBe('50%');
    expect(slider.title).toBe('Graph background: rgb(128, 128, 128)');

    expect(runtime._relationNameColorForGraphBackground((180 / 255) * 100)).toBe('rgb(255, 255, 255)');
    expect(runtime._relationNameColorForGraphBackground((181 / 255) * 100)).toBe('rgb(0, 0, 255)');

    slider.value = '71';
    slider.dispatchEvent(new Event('input'));
    expect(document.documentElement.style.getPropertyValue('--ivy-graph-background')).toBe('rgb(181, 181, 181)');
    expect(document.documentElement.style.getPropertyValue('--ivy-relation-name-color')).toBe('rgb(0, 0, 255)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-color')).toBe('rgb(181, 181, 181)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-fill')).toBe('71%');
    expect(slider.title).toBe('Graph background: rgb(181, 181, 181)');

    slider.value = '100';
    slider.dispatchEvent(new Event('input'));
    expect(document.documentElement.style.getPropertyValue('--ivy-graph-background')).toBe('rgb(255, 255, 255)');
    expect(document.documentElement.style.getPropertyValue('--ivy-relation-name-color')).toBe('rgb(0, 0, 255)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-color')).toBe('rgb(255, 255, 255)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-fill')).toBe('100%');
    expect(slider.title).toBe('Graph background: rgb(255, 255, 255)');
  });

  it('shows a pointer readout while dragging the graph background slider, then removes it on release', () => {
    vi.useFakeTimers();
    document.body.innerHTML = '<input id="graph-background-slider" type="range" min="0" max="100" value="50">';
    const runtime = makeRuntime();
    const slider = document.getElementById('graph-background-slider') as HTMLInputElement;

    runtime._setupGraphBackgroundSlider();
    slider.dispatchEvent(new MouseEvent('pointerdown', { clientX: 20, clientY: 30, bubbles: true }));

    let readout = document.querySelector('.graph-background-readout') as HTMLElement | null;
    expect(readout?.textContent).toBe('rgb(128, 128, 128)');
    expect(readout?.classList.contains('visible')).toBe(true);
    expect(readout?.style.left).toBe('34px');
    expect(readout?.style.top).toBe('46px');

    slider.value = '75';
    slider.dispatchEvent(new Event('input', { bubbles: true }));
    readout = document.querySelector('.graph-background-readout') as HTMLElement | null;
    expect(readout?.textContent).toBe('rgb(191, 191, 191)');
    expect(slider.title).toBe('Graph background: rgb(191, 191, 191)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-color')).toBe('rgb(191, 191, 191)');
    expect(slider.style.getPropertyValue('--ivy-graph-slider-fill')).toBe('75%');

    document.dispatchEvent(new MouseEvent('pointerup', { clientX: 44, clientY: 55, bubbles: true }));
    expect(document.querySelector('.graph-background-readout')).toBeNull();
  });

  it('shows a direct DOM toast notification', () => {
    const runtime = makeRuntime();

    runtime._showToast('Connection lost', 'error', { persistent: true, className: 'custom-toast' });

    const toast = document.querySelector('.ivy-toast');
    expect(toast.textContent).toBe('Connection lost');
    expect(toast.classList.contains('ivy-toast-floating')).toBe(true);
    expect(toast.classList.contains('ivy-toast-error')).toBe(true);
    expect(toast.classList.contains('custom-toast')).toBe(true);
  });

  it('opens external tutorial links outside the iframe while preserving local links', () => {
    window.history.replaceState(null, '', '/static/tutorial/kenmcmil.github.io/ivy/examples/sht/table.html');
    const runtime = makeRuntime();
    const tutorialDoc = document.implementation.createHTMLDocument('tutorial');
    tutorialDoc.body.innerHTML = [
      '<a id="external" href="http://dl.acm.org/citation.cfm?id=359108">ACM</a>',
      '<a id="local" href="table.ivy">Local</a>',
    ].join('');
    const open = vi.spyOn(window, 'open').mockImplementation(() => null);

    try {
      expect(runtime._externalTutorialLinkUrl(
        tutorialDoc.getElementById('local'),
        { location: { href: window.location.href } },
      )).toBe('');
      runtime._installTutorialExternalLinkInterceptor({ contentDocument: tutorialDoc });

      const event = new MouseEvent('click', { bubbles: true, cancelable: true });
      const dispatched = tutorialDoc.getElementById('external')!.dispatchEvent(event);

      expect(dispatched).toBe(false);
      expect(open).toHaveBeenCalledWith(
        'http://dl.acm.org/citation.cfm?id=359108',
        '_blank',
        'noopener,noreferrer',
      );
    } finally {
      open.mockRestore();
    }
  });

  it('opens a large crash report dialog when the browser WASM Go runtime exits', () => {
    const runtime = makeRuntime();
    runtime.textDialog = vi.fn();

    runtime.handleEvent({
      type: 'browser_wasm_runtime_crash',
      data: {
        timestamp: '2026-05-19T01:02:03.000Z',
        reason: 'go.run rejected',
        message: 'panic: fake crash',
        stack: 'Error: fake crash\n    at go.run',
        recent_output: '[stderr] panic: fake crash\n[stderr] goroutine 1 [running]\n',
        asset_base_url: '/static/wasm/',
        generation: 3,
      },
    });

    expect(runtime.controls.lastStatus).toEqual({
      message: 'Browser WASM Go runtime exited; crash report opened',
      kind: 'error',
    });
    expect(runtime.controls.lastInfo.shortInfo).toBe('Browser WASM Go runtime crashed');
    expect(runtime.textDialog).toHaveBeenCalledWith(
      'Browser WASM Go runtime crashed',
      expect.stringContaining('Ivy will try to restart it'),
      expect.stringContaining('goroutine 1 [running]'),
      expect.objectContaining({ readOnly: true, rows: 28, cols: 120 }),
    );
  });

  it('reloads edited model content before model-dependent API calls', async () => {
    let apiInstance: any = null;
    class ReloadingAPI extends FakeAPI {
      constructor() {
        super();
        apiInstance = this;
        this.reloadContent = vi.fn(async () => ({ status: 'ok', isolates: ['fresh_iso'], isolate: 'fresh_iso' }));
        this.executeAction = vi.fn(async () => ({ status: 'ran' }));
      }
    }
    const runtime = makeRuntime({ IvyAPI: ReloadingAPI });
    runtime.cmEditor = { getValue: vi.fn(() => 'type client\n') };
    runtime._persistedFileName = 'client.ivy';
    runtime.activeIsolate = 'iso_client';
    runtime.availableIsolates = ['iso_client'];

    runtime._invalidateModelState('test-edit');
    expect(runtime.activeIsolate).toBe('');
    expect(runtime.availableIsolates).toEqual([]);
    const result = await runtime.api.executeAction('diagram', {});

    expect(result).toEqual({ status: 'ran' });
    expect(apiInstance.reloadContent).toHaveBeenCalledWith('type client\n', 'client.ivy', { isolate: '' });
    expect(apiInstance.executeAction).toHaveBeenCalledWith('diagram', {});
    expect(apiInstance.reloadContent.mock.invocationCallOrder[0]).toBeLessThan(
      apiInstance.executeAction.mock.invocationCallOrder[0],
    );
    expect(runtime._modelStateInvalid).toBe(false);
    expect(runtime.activeIsolate).toBe('fresh_iso');
    expect(runtime.availableIsolates).toEqual(['fresh_iso']);
  });

  it('ignores delayed editor change events when the content still matches the freshly loaded model', () => {
    const runtime = makeRuntime();
    const invalidateModelState = vi.fn();
    runtime.uiDataStore = { invalidateModelState };
    runtime.cmEditor = { getValue: vi.fn(() => 'type client\n') };
    runtime._loadedModelContent = 'type client\n';

    expect(runtime._invalidateModelState('editor-change')).toBe(false);

    expect(runtime._modelStateInvalid).toBe(false);
    expect(invalidateModelState).not.toHaveBeenCalled();
  });

  it('invalidates model state when the editor content actually changes', () => {
    const runtime = makeRuntime();
    const invalidateModelState = vi.fn();
    runtime.uiDataStore = { invalidateModelState };
    runtime.cmEditor = { getValue: vi.fn(() => 'type server\n') };
    runtime._loadedModelContent = 'type client\n';

    expect(runtime._invalidateModelState('editor-change')).toBe(true);

    expect(runtime._modelStateInvalid).toBe(true);
    expect(invalidateModelState).toHaveBeenCalledTimes(1);
  });

  it('ignores delayed editor change events caused by programmatic file loads', () => {
    const runtime = makeRuntime();
    const invalidateModelState = vi.fn();
    runtime.uiDataStore = { invalidateModelState };
    runtime._loadedModelContent = 'previous model';
    runtime._modelStateInvalid = false;
    runtime.cmEditor = {
      setValue: vi.fn(),
      getValue: vi.fn(() => 'new model'),
    };
    runtime._updateReopenLastFileButton = vi.fn();

    runtime.setEditorContent('new model');

    expect(runtime._invalidateModelState('editor-change')).toBe(false);
    expect(runtime._modelStateInvalid).toBe(false);
    expect(invalidateModelState).not.toHaveBeenCalled();
  });

  it('does not reload model content for event-trace-only actions', async () => {
    let apiInstance: any = null;
    class EventAPI extends FakeAPI {
      constructor() {
        super();
        apiInstance = this;
        this.reloadContent = vi.fn(async () => ({ status: 'ok' }));
        this.executeAction = vi.fn(async () => ({ status: 'events' }));
      }
    }
    const runtime = makeRuntime({ IvyAPI: EventAPI });
    runtime.cmEditor = { getValue: vi.fn(() => 'edited') };

    runtime._invalidateModelState('test-edit');
    await runtime.api.executeAction('events_find', { query: 'send' });

    expect(apiInstance.reloadContent).not.toHaveBeenCalled();
    expect(apiInstance.executeAction).toHaveBeenCalledWith('events_find', { query: 'send' });
    expect(runtime._modelStateInvalid).toBe(true);
  });

  it('opens job control and toggles browser/remote submission mode', () => {
    document.body.innerHTML = [
      '<button id="btn-toggle-job-control"></button>',
      '<section id="job-control-page" aria-hidden="true"></section>',
      '<button id="job-control-close"></button>',
      '<div id="job-control-backend-status" hidden></div>',
      '<button id="job-submission-toggle" class="job-mode-toggle is-browser" data-mode="browser"></button>',
      '<span id="job-submission-label"></span>',
    ].join('');
    const runtime = makeRuntime();

    runtime._setupJobControlHandlers();
    expect(document.getElementById('btn-toggle-job-control').classList.contains('job-submission-browser')).toBe(true);
    expect(document.getElementById('btn-toggle-job-control').getAttribute('data-job-submission-mode')).toBe('browser');
    expect(document.getElementById('job-submission-label').textContent).toBe('run in browser');
    expect(document.getElementById('job-control-backend-status').hidden).toBe(true);

    document.getElementById('btn-toggle-job-control').click();

    expect(document.getElementById('job-control-page').classList.contains('open')).toBe(true);
    expect(document.getElementById('job-control-page').getAttribute('aria-hidden')).toBe('false');
    expect(document.getElementById('btn-toggle-job-control').classList.contains('active')).toBe(true);

    document.getElementById('job-submission-toggle').click();
    expect(runtime.jobSubmissionMode).toBe('remote');
    expect(document.getElementById('job-submission-toggle').classList.contains('is-remote')).toBe(true);
    expect(document.getElementById('btn-toggle-job-control').classList.contains('job-submission-browser')).toBe(false);
    expect(document.getElementById('btn-toggle-job-control').classList.contains('job-submission-remote')).toBe(true);
    expect(document.getElementById('btn-toggle-job-control').getAttribute('data-job-submission-mode')).toBe('remote');
    expect(document.getElementById('job-submission-label').textContent).toBe('run on remote');

    document.getElementById('job-submission-toggle').click();
    expect(runtime.jobSubmissionMode).toBe('browser');
    expect(document.getElementById('job-submission-toggle').classList.contains('is-browser')).toBe(true);
    expect(document.getElementById('btn-toggle-job-control').classList.contains('job-submission-browser')).toBe(true);
    expect(document.getElementById('btn-toggle-job-control').classList.contains('job-submission-remote')).toBe(false);
    expect(document.getElementById('btn-toggle-job-control').getAttribute('data-job-submission-mode')).toBe('browser');
    expect(document.getElementById('job-submission-label').textContent).toBe('run in browser');

    document.getElementById('job-control-close').click();
    expect(document.getElementById('job-control-page').classList.contains('open')).toBe(false);
  });

  it('binds every visible static File menu item to a controller command', () => {
    vi.useFakeTimers();
    document.body.innerHTML = [
      '<input id="file-input" type="file">',
      '<input id="event-file-input" type="file">',
      '<input id="analysis-state-file-input" type="file">',
      '<button id="btn-toggle-tutorial"></button>',
      '<a id="file-load" href="#"></a>',
      '<a id="file-open-event-trace" href="#"></a>',
      '<a id="file-save-as" href="#"></a>',
      '<a id="file-download" href="#"></a>',
      '<a id="file-save-analysis-state" href="#"></a>',
      '<a id="file-load-analysis-state" href="#"></a>',
      '<a id="file-save-invariant" href="#"></a>',
      '<a id="file-new" href="#"></a>',
      '<div id="file-menu" class="dropdown-content"></div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.setupDropdownMenus = vi.fn();
    runtime._bindStaticMenuActions = vi.fn();
    runtime.attachGraphEventHandlers = vi.fn();
    runtime.setUIMode = vi.fn();
    runtime.closeAllDropdowns = vi.fn();
    runtime.chooseAndLoadModelFile = vi.fn(async () => true);
    runtime.chooseAndLoadEventTraceFile = vi.fn(async () => true);
    runtime.saveAs = vi.fn();
    runtime.downloadModel = vi.fn();
    runtime.saveAnalysisState = vi.fn();
    runtime.chooseAndLoadAnalysisStateFile = vi.fn();
    runtime.saveInvariant = vi.fn();
    runtime.newModel = vi.fn();
    runtime.runAction = vi.fn();

    runtime.setupEventHandlers();

    document.getElementById('file-load')!.click();
    document.getElementById('file-open-event-trace')!.click();
    document.getElementById('file-save-as')!.click();
    document.getElementById('file-download')!.click();
    document.getElementById('file-save-analysis-state')!.click();
    document.getElementById('file-load-analysis-state')!.click();
    document.getElementById('file-save-invariant')!.click();
    document.getElementById('file-new')!.click();
    vi.advanceTimersByTime(50);

    expect(runtime.chooseAndLoadModelFile).toHaveBeenCalledTimes(1);
    expect(runtime.chooseAndLoadEventTraceFile).toHaveBeenCalledTimes(1);
    expect(runtime.saveAs).toHaveBeenCalledTimes(1);
    expect(runtime.downloadModel).toHaveBeenCalledTimes(1);
    expect(runtime.saveAnalysisState).toHaveBeenCalledTimes(1);
    expect(runtime.chooseAndLoadAnalysisStateFile).toHaveBeenCalledTimes(1);
    expect(runtime.saveInvariant).toHaveBeenCalledTimes(1);
    expect(runtime.newModel).toHaveBeenCalledTimes(1);
    expect(runtime.runAction).not.toHaveBeenCalled();
  });

  it('clicks every descriptor File menu item through controller commands', async () => {
    document.body.innerHTML = [
      '<div id="sheet-2" class="sheet-content active">',
      '  <div id="arg-panel"><div class="panel-header"><div class="panel-header-actions"></div></div></div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-2';
    runtime.closeAllDropdowns = vi.fn();
    runtime.save = vi.fn(async () => true);
    runtime.saveAnalysisState = vi.fn(async () => ({ ok: true }));
    runtime.saveAbstraction = vi.fn(async () => ({ ok: true }));
    runtime.removeSheet = vi.fn();
    runtime.closeCurrentFile = vi.fn(async () => true);
    runtime.runAction = vi.fn();

    runtime.renderMenuRegion('arg', [{
      type: 'menu',
      label: 'File',
      items: [
        { type: 'button', label: 'Save', action: 'save_model', dispatch: 'action', enabled: true },
        { type: 'button', label: 'Save analysis state', action: 'save_analysis_state', dispatch: 'action', enabled: true },
        { type: 'button', label: 'Save abstraction', action: 'save_abstraction', dispatch: 'action', enabled: true },
        { type: 'separator', label: '---' },
        { type: 'button', label: 'Remove tab', action: 'remove_tab', dispatch: 'action', enabled: true },
        { type: 'button', label: 'Exit', action: 'exit', dispatch: 'action', enabled: true },
      ],
    }]);

    for (const link of document.querySelectorAll('[data-dynamic-menu-region="arg"] [data-menu-action]')) {
      (link as HTMLElement).click();
    }
    await Promise.resolve();

    expect(runtime.save).toHaveBeenCalledTimes(1);
    expect(runtime.saveAnalysisState).toHaveBeenCalledTimes(1);
    expect(runtime.saveAbstraction).toHaveBeenCalledTimes(1);
    expect(runtime.removeSheet).toHaveBeenCalledWith('sheet-2');
    expect(runtime.closeCurrentFile).toHaveBeenCalledTimes(1);
    expect(runtime.runAction).not.toHaveBeenCalled();
  });

  it('requests menu descriptors for the active sheet and workflow', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-7';
    runtime.uiMode = 'reachability';
    runtime.api.getMenus = vi.fn(async () => ({ arg: [], concept: [] }));
    runtime.renderMenuRegion = vi.fn();

    await runtime.loadMenuDescriptors();

    expect(runtime.api.getMenus).toHaveBeenCalledWith({
      sheetId: 'sheet-7',
      uiMode: 'reachability',
    });
    expect(runtime.renderMenuRegion).toHaveBeenCalledWith('arg', []);
    expect(runtime.renderMenuRegion).toHaveBeenCalledWith('concept', []);
  });

  it('refreshes menu descriptors when the workflow mode changes', async () => {
    document.body.innerHTML = [
      '<select id="ui-mode-select">',
      '  <option value="cti">CTI</option>',
      '  <option value="reachability">reachability</option>',
      '</select>',
    ].join('');
    const runtime = makeRuntime();
    runtime.controls.setStatus = vi.fn();
    runtime.loadMenuDescriptors = vi.fn(async () => ({ ok: true }));

    runtime.setUIMode('reachability', { announce: true });
    await Promise.resolve();

    expect(document.body.getAttribute('data-ui-mode')).toBe('reachability');
    expect(runtime.loadMenuDescriptors).toHaveBeenCalledTimes(1);
  });

  it('renders descriptor menus into the active sheet panel', () => {
    document.body.innerHTML = [
      '<div id="sheet-1" class="sheet-content">',
      '  <div id="arg-panel"><div class="panel-header"><div class="panel-header-actions"></div></div></div>',
      '</div>',
      '<div id="sheet-2" class="sheet-content active">',
      '  <div id="arg-panel"><div class="panel-header"><div class="panel-header-actions"></div></div></div>',
      '</div>',
    ].join('');
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-2';

    runtime.renderMenuRegion('arg', [{
      type: 'menu',
      label: 'Action',
      items: [{ type: 'button', label: 'Reach', action: 'reach', dispatch: 'action', enabled: true }],
    }]);

    expect(document.querySelector('#sheet-1 [data-dynamic-menu-region="arg"]')).toBeNull();
    expect(document.querySelector('#sheet-2 [data-dynamic-menu-region="arg"]')?.textContent).toContain('Action');
  });

  it('does not duplicate top-level menu labels when reachability descriptors render', () => {
    installStaticMenuDom();
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-1';
    runtime.uiMode = 'reachability';

    runtime.renderMenuRegion('arg', [
      { type: 'menu', label: 'File', items: [{ type: 'button', label: 'Save', action: 'save_model', enabled: true }] },
      { type: 'menu', label: 'Mode', items: [{ type: 'button', label: 'Pdr', action: 'mode_pdr', enabled: true }] },
      { type: 'menu', label: 'Action', items: [{ type: 'button', label: 'Show reachable states', action: 'show_reachable', enabled: true }] },
    ]);
    runtime.renderMenuRegion('concept', [
      { type: 'menu', label: 'Action', items: [{ type: 'button', label: 'Undo', action: 'undo', enabled: true }] },
      { type: 'menu', label: 'View', items: [{ type: 'button', label: 'Add relation', action: 'add_relation', enabled: true }] },
    ]);

    const argMenuScope = document.createElement('div');
    argMenuScope.append(
      document.getElementById('menubar')!.cloneNode(true),
      document.getElementById('arg-panel')!.cloneNode(true),
    );

    expect({
      reachabilityGraph: duplicatedVisibleMenuLabels(argMenuScope, 'reachability'),
      conceptGraph: duplicatedVisibleMenuLabels(document.getElementById('concept-panel')!, 'reachability'),
    }).toEqual({
      reachabilityGraph: [],
      conceptGraph: [],
    });
  });

  it('does not duplicate top-level menu labels when CTI descriptors render', () => {
    installStaticMenuDom();
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-1';
    runtime.uiMode = 'cti';

    runtime.renderMenuRegion('arg', [
      { type: 'menu', label: 'File', items: [{ type: 'button', label: 'Save invariant', action: 'save_conjectures', enabled: true }] },
      { type: 'menu', label: 'Invariant', items: [{ type: 'button', label: 'Check induction', action: 'check_inductiveness', enabled: true }] },
    ]);
    runtime.renderMenuRegion('concept', [
      { type: 'menu', label: 'Conjecture', items: [{ type: 'button', label: 'Undo', action: 'undo', enabled: true }] },
      { type: 'menu', label: 'View', items: [{ type: 'button', label: 'Add relation', action: 'add_relation', enabled: true }] },
    ]);

    const argMenuScope = document.createElement('div');
    argMenuScope.append(
      document.getElementById('menubar')!.cloneNode(true),
      document.getElementById('arg-panel')!.cloneNode(true),
    );

    expect({
      reachabilityGraph: duplicatedVisibleMenuLabels(argMenuScope, 'cti'),
      conceptGraph: duplicatedVisibleMenuLabels(document.getElementById('concept-panel')!, 'cti'),
    }).toEqual({
      reachabilityGraph: [],
      conceptGraph: [],
    });
  });

  it('sends CTI mode to Diagram and applies the returned pre-state label', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-2';
    runtime.getUIMode = vi.fn(() => 'cti');
    runtime.api.executeAction = vi.fn(async () => ({
      status: 'diagrammed',
      message: 'Diagram complete.',
      concept: {
        sheet_id: 'sheet-2',
        elements: [],
        cti_state_label: 'CTI pre-state 0',
      },
    }));
    runtime.applyConceptSnapshot = vi.fn();

    await runtime.diagramCurrentState();

    expect(runtime.api.executeAction).toHaveBeenCalledWith('diagram', {
      sheet_id: 'sheet-2',
      ui_mode: 'cti',
    });
    expect(runtime.applyConceptSnapshot).toHaveBeenCalledWith('sheet-2', {
      sheet_id: 'sheet-2',
      elements: [],
      cti_state_label: 'CTI pre-state 0',
    });
  });

  it('explains remote backend switch failures in the job control panel', async () => {
    class BrowserAPI extends FakeAPI {
      constructor() {
        super({ kind: 'browser-wasm' });
        this.createSession = vi.fn(async () => 'browser-s1');
        this.disconnectEvents = vi.fn();
      }
    }
    class UnreachableRemoteAPI extends FakeAPI {
      constructor() {
        super({ kind: 'hosted-go' });
        this.createSession = vi.fn(async () => {
          throw new Error('connection refused');
        });
      }
    }

    document.body.innerHTML = [
      '<button id="btn-toggle-job-control"></button>',
      '<div id="job-control-backend-status" hidden></div>',
      '<button id="job-submission-toggle" class="job-mode-toggle is-browser" data-mode="browser"></button>',
      '<span id="job-submission-label"></span>',
      '<span id="session-id"></span>',
    ].join('');
    const runtime = makeRuntime({
      BrowserIvyAPI: BrowserAPI,
      IvyAPI: UnreachableRemoteAPI,
    });
    runtime._jobSubmissionReady = true;

    await runtime._switchJobSubmissionBackend('remote');

    expect(runtime.jobSubmissionMode).toBe('browser');
    expect(runtime.api.kind).toBe('browser-wasm');
    expect(document.getElementById('job-submission-toggle').classList.contains('is-browser')).toBe(true);
    expect(document.getElementById('btn-toggle-job-control').getAttribute('data-job-submission-mode')).toBe('browser');
    expect(document.getElementById('job-control-backend-status').hidden).toBe(false);
    expect(document.getElementById('job-control-backend-status').textContent).toBe('server unreachable');
  });

  it('stays on the remote backend when only the model reload fails on a reachable server', async () => {
    class BrowserAPI extends FakeAPI {
      constructor() {
        super({ kind: 'browser-wasm' });
        this.createSession = vi.fn(async () => 'browser-s1');
        this.disconnectEvents = vi.fn();
      }
    }
    class ReachableRemoteAPI extends FakeAPI {
      constructor() {
        super({ kind: 'hosted-go' });
        this.sessionId = 'remote-s1';
        this.createSession = vi.fn(async () => 'remote-s1');
        // Server is reachable but rejects the model (e.g. a compile error).
        this.reloadContent = vi.fn(async () => {
          throw new Error('API error 400: syntax error');
        });
      }
    }

    document.body.innerHTML = [
      '<button id="btn-toggle-job-control"></button>',
      '<div id="job-control-backend-status" hidden></div>',
      '<button id="job-submission-toggle" class="job-mode-toggle is-browser" data-mode="browser"></button>',
      '<span id="job-submission-label"></span>',
      '<span id="session-id"></span>',
    ].join('');
    const runtime = makeRuntime({
      BrowserIvyAPI: BrowserAPI,
      IvyAPI: ReachableRemoteAPI,
      IvyPersist: makePersist({ getSessionIdFromURL: () => '' }),
    });
    runtime._jobSubmissionReady = true;
    runtime._persistedFileContent = 'type client\n';
    runtime._persistedFileName = 'client.ivy';

    await runtime._switchJobSubmissionBackend('remote');

    // The session was created, so the server is reachable: stay on remote.
    expect(runtime._apiMode).toBe('remote');
    expect(runtime.api.kind).toBe('hosted-go');
    // Do NOT falsely blame the (reachable) server.
    expect(document.getElementById('job-control-backend-status').textContent).toBe('');
    // Surface the real model error instead.
    expect(runtime.controls.lastStatus.kind).toBe('error');
    expect(runtime.controls.lastStatus.message).toContain('reloading the model failed');
    expect(runtime.controls.lastStatus.message).toContain('syntax error');
  });

  it('confirms before clearing saved localStorage session data from job control', async () => {
    document.body.innerHTML = [
      '<button id="job-clear-saved-sessions"></button>',
    ].join('');
    const clearSavedSessions = vi.fn(() => 3);
    const runtime = makeRuntime({
      IvyPersist: {
        clearSavedSessions,
      },
    });

    runtime._setupJobControlHandlers();
    document.getElementById('job-clear-saved-sessions')?.click();

    expect(document.querySelector('.dialog-message')?.textContent).toBe('Really delete all browser localStorage sessions?');
    const buttons = Array.from(document.querySelectorAll('[data-ivy-dialog-button]')) as HTMLButtonElement[];
    expect(buttons.map((button) => button.textContent)).toEqual(['Delete', 'Cancel']);
    expect(document.activeElement).toBe(buttons[1]);

    buttons[1].click();
    await Promise.resolve();
    expect(clearSavedSessions).not.toHaveBeenCalled();
    expect(runtime.controls.lastStatus).toEqual({
      message: 'Clear saved session data cancelled',
      kind: 'warning',
    });

    document.getElementById('job-clear-saved-sessions')?.click();
    const deleteButton = document.querySelector('[data-ivy-dialog-button]') as HTMLButtonElement;
    deleteButton.click();
    await Promise.resolve();

    expect(clearSavedSessions).toHaveBeenCalledTimes(1);
    expect(runtime.controls.lastStatus).toEqual({
      message: 'Deleted browser localStorage sessions',
      kind: 'success',
    });
  });

  it('accepts the default integer dialog value with Enter', async () => {
    const runtime = makeRuntime();

    const resultPromise = runtime.integerDialog('Bounded check', 'Number of steps to check:', 10, { min: 0 });
    const input = document.querySelector('[data-ivy-dialog-int]') as HTMLInputElement;
    const buttons = Array.from(document.querySelectorAll('[data-ivy-dialog-button]'));

    expect(input.value).toBe('10');
    expect(buttons.map((button) => button.textContent)).toEqual(['OK', 'Cancel']);

    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    await expect(resultPromise).resolves.toBe(10);
    expect(document.querySelector('[data-ivy-dialog]')).toBeNull();
  });

  it('submits entry dialogs with Return', async () => {
    const runtime = makeRuntime();

    const resultPromise = runtime.entryDialog('Remember graph', 'Name:', '', {});
    const input = document.querySelector('[data-ivy-dialog-entry]') as HTMLInputElement;
    input.value = 'goal-from-return';

    let resolved = false;
    resultPromise.then(() => { resolved = true; });
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await Promise.resolve();

    const resolvedByReturn = resolved;
    if (!resolvedByReturn) {
      const ok = Array.from(document.querySelectorAll('[data-ivy-dialog-button]'))
        .find((button) => button.textContent === 'OK') as HTMLButtonElement;
      ok.click();
      await resultPromise;
    }
    expect(resolvedByReturn).toBe(true);
    await expect(resultPromise).resolves.toBe('goal-from-return');
  });

  it('can return listbox selection indices for Tk-compatible callers', async () => {
    const runtime = makeRuntime();

    const singlePromise = runtime.listboxDialog('Pick one', 'Choice:', [
      { label: 'Alpha', value: 'alpha' },
      { label: 'Beta', value: 'beta' },
    ], { returnIndex: true });
    const single = document.querySelector('[data-ivy-dialog-list]') as HTMLSelectElement;
    single.selectedIndex = 1;
    let ok = Array.from(document.querySelectorAll('[data-ivy-dialog-button]'))
      .find((button) => button.textContent === 'OK') as HTMLButtonElement;
    ok.click();
    await expect(singlePromise).resolves.toBe(1);

    const multiPromise = runtime.listboxDialog('Pick many', 'Choices:', [
      { label: 'Alpha', value: 'alpha' },
      { label: 'Beta', value: 'beta' },
      { label: 'Gamma', value: 'gamma' },
    ], { multiple: true, returnIndex: true });
    const multi = document.querySelector('[data-ivy-dialog-list]') as HTMLSelectElement;
    Array.from(multi.options).forEach((option, index) => {
      option.selected = index === 0 || index === 2;
    });
    ok = Array.from(document.querySelectorAll('[data-ivy-dialog-button]'))
      .find((button) => button.textContent === 'OK') as HTMLButtonElement;
    ok.click();
    await expect(multiPromise).resolves.toEqual([0, 2]);
  });

  it('cancels listbox and button-list dialogs with Tk-compatible values', async () => {
    const runtime = makeRuntime();

    const singlePromise = runtime.listboxDialog('Pick one', 'Choice:', ['a', 'b'], {});
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await expect(singlePromise).resolves.toBeNull();

    const multiPromise = runtime.listboxDialog('Pick many', 'Choices:', ['a', 'b'], { multiple: true });
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await expect(multiPromise).resolves.toEqual([]);

    const buttonPromise = runtime.buttonListDialog('Choose', 'Continue?', [
      { label: 'Go', value: 'go' },
    ]);
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await expect(buttonPromise).resolves.toBeNull();
  });

  it('keeps integer dialogs open for out-of-range input', async () => {
    const runtime = makeRuntime();

    const resultPromise = runtime.integerDialog('Bounded check', 'Number of steps to check:', 1, { min: 0, max: 3 });
    const input = document.querySelector('[data-ivy-dialog-int]') as HTMLInputElement;
    let resolved = false;
    resultPromise.then(() => { resolved = true; });

    input.value = '4';
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await Promise.resolve();

    expect(resolved).toBe(false);
    expect(document.querySelector('[data-ivy-dialog-error]')?.textContent).toBe('Enter a value at most 3.');
    expect(document.querySelector('[data-ivy-dialog]')).not.toBeNull();

    input.value = '3';
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));
    await expect(resultPromise).resolves.toBe(3);
    expect(document.querySelector('[data-ivy-dialog]')).toBeNull();
  });

  it('preseeds dialog answers for deterministic command tests', async () => {
    const runtime = makeRuntime();
    runtime.preseedDialogAnswers([
      { kind: 'entry', value: 'goal-a' },
      { kind: 'integer', value: 7 },
      { kind: 'listbox', value: 'choice-b' },
      { kind: 'listbox', value: [0, 2] },
      { kind: 'buttonList', value: 'save' },
    ]);

    await expect(runtime.entryDialog('Remember graph', 'Name:', '', {})).resolves.toBe('goal-a');
    await expect(runtime.integerDialog('Bounded check', 'Bound:', 3, { min: 0 })).resolves.toBe(7);
    await expect(runtime.listboxDialog('Pick one', 'Choice:', ['choice-a', 'choice-b'], {})).resolves.toBe('choice-b');
    await expect(runtime.listboxDialog('Pick many', 'Choices:', [0, 1, 2], { multiple: true })).resolves.toEqual([0, 2]);
    await expect(runtime.buttonListDialog('Unsaved changes', 'Save first?', [
      { label: 'Save', value: 'save' },
      { label: 'Discard', value: 'discard' },
    ])).resolves.toBe('save');

    expect(document.querySelector('[data-ivy-dialog]')).toBeNull();
  });

  it('passes preseeded entry answers through real commands', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-9';
    runtime.api.executeAction = vi.fn(async () => ({ status: 'ok' }));
    runtime.preseedDialogAnswers([{ kind: 'entry', value: 'saved-goal' }]);

    await runtime.rememberGraph();

    expect(runtime.api.executeAction).toHaveBeenCalledWith('remember', {
      name: 'saved-goal',
      sheet_id: 'sheet-9',
    });
    expect(document.querySelector('[data-ivy-dialog]')).toBeNull();
  });

  it('detects when the focused editor should own Emacs Ctrl-S search', () => {
    const runtime = makeRuntime();
    runtime.cmEditor = {
      hasFocus: vi.fn(() => true),
      getOption: vi.fn(() => 'emacs'),
    };

    expect(runtime._editorHasFocusedEmacsKeymap()).toBe(true);

    runtime.cmEditor.getOption.mockReturnValue('vim');
    expect(runtime._editorHasFocusedEmacsKeymap()).toBe(false);

    runtime.cmEditor.hasFocus.mockReturnValue(false);
    runtime.cmEditor.getOption.mockReturnValue('emacs');
    expect(runtime._editorHasFocusedEmacsKeymap()).toBe(false);
  });

  it('renders isolate choices and appends the active isolate to sheet tabs', () => {
    document.body.innerHTML = [
      '<div id="isolate-menu-wrapper" class="dropdown isolate-menu" hidden>',
      '  <span id="isolate-menu-title" class="panel-menu" data-dropdown="isolate-menu">isolate</span>',
      '  <div id="isolate-menu" class="dropdown-content"></div>',
      '</div>',
      '<div id="tab-bar"><button class="sheet-tab active" data-sheet="sheet-1"><span>Sheet 1</span></button></div>',
    ].join('');
    const runtime = makeRuntime();

    runtime.setIsolates(['cf_backup', 'cf_live'], 'cf_live');

    expect((document.getElementById('isolate-menu-wrapper') as HTMLElement).hidden).toBe(false);
    expect(document.getElementById('isolate-menu-title')?.textContent).toBe('cf_live');
    expect(document.getElementById('isolate-menu-title')?.getAttribute('title')).toBe('choose isolate');
    expect(document.querySelector('#isolate-menu .dropdown-heading')?.textContent).toBe('choose isolate:');
    expect(Array.from(document.querySelectorAll('#isolate-menu a')).map((item) => item.textContent)).toEqual([
      'cf_backup',
      'cf_live',
    ]);
    expect(document.querySelector('.sheet-tab span')?.textContent).toBe('Sheet 1 · cf_live');

    runtime.setIsolates(['cf_backup', 'cf_live'], 'cf_backup');
    expect(document.querySelector('.sheet-tab span')?.textContent).toBe('Sheet 1 · cf_backup');
  });

  it('shows no_isolates_found for specs without declared isolates', () => {
    document.body.innerHTML = [
      '<div id="isolate-menu-wrapper" class="dropdown isolate-menu" hidden>',
      '  <span id="isolate-menu-title" class="panel-menu" data-dropdown="isolate-menu">isolate</span>',
      '  <div id="isolate-menu" class="dropdown-content"></div>',
      '</div>',
      '<div id="tab-bar"><button class="sheet-tab active" data-sheet="sheet-1"><span>Sheet 1</span></button></div>',
    ].join('');
    const runtime = makeRuntime();

    runtime.setIsolates(['no_isolates_found'], 'no_isolates_found');

    expect((document.getElementById('isolate-menu-wrapper') as HTMLElement).hidden).toBe(false);
    expect(document.getElementById('isolate-menu-title')?.textContent).toBe('no_isolates_found');
    expect(document.getElementById('isolate-menu-title')?.getAttribute('title')).toBe('choose isolate');
    expect(document.querySelector('#isolate-menu .dropdown-heading')?.textContent).toBe('choose isolate:');
    expect(Array.from(document.querySelectorAll('#isolate-menu a')).map((item) => item.textContent)).toEqual(['no_isolates_found']);
    expect(document.querySelector('.sheet-tab span')?.textContent).toBe('Sheet 1');
  });

  it('does not keep an active isolate that is absent from the current isolate list', () => {
    document.body.innerHTML = [
      '<div id="isolate-menu-wrapper" class="dropdown isolate-menu" hidden>',
      '  <span id="isolate-menu-title" class="panel-menu" data-dropdown="isolate-menu">isolate</span>',
      '  <div id="isolate-menu" class="dropdown-content"></div>',
      '</div>',
      '<div id="tab-bar"><button class="sheet-tab active" data-sheet="sheet-1"><span>Sheet 1</span></button></div>',
    ].join('');
    const runtime = makeRuntime();

    runtime.setIsolates(['service', 'protocol'], 'dramc_nb2');

    expect(runtime.activeIsolate).toBe('');
    expect(runtime.availableIsolates).toEqual(['protocol', 'service']);
    expect(document.getElementById('isolate-menu-title')?.textContent).toBe('isolate');
    expect(document.querySelector('.sheet-tab span')?.textContent).toBe('Sheet 1');
  });

  it('opens reachability-only sheets without a concept graph runtime', () => {
    vi.useFakeTimers();
    installSheetDom();
    const runtime = makeRuntime();
    runtime.recalculateAll = vi.fn();
    runtime.showReachableStates = vi.fn();

    const sheetId = runtime.addSheet('Sheet 3', 'sheet-3', { reachabilityOnly: true });
    const sheet = document.getElementById('sheet-3');

    expect(sheetId).toBe('sheet-3');
    expect(sheet.classList.contains('reachability-only-sheet')).toBe(true);
    expect(sheet.getAttribute('data-sheet-layout')).toBe('reachability-only');
    expect(runtime.sheets['sheet-3'].reachabilityOnly).toBe(true);
    expect(runtime.sheets['sheet-3'].argGraph).toBeTruthy();
    expect(runtime.sheets['sheet-3'].conceptGraph).toBeNull();

    const actionTrigger = sheet.querySelector('[data-dropdown="arg-action-menu"]') as HTMLElement;
    actionTrigger.click();

    const actionDropdown = actionTrigger.closest('.dropdown');
    expect(actionDropdown?.classList.contains('open')).toBe(true);
    expect(sheet.querySelector('#arg-action-menu')?.textContent).toContain('Recalculate all');
    expect(sheet.querySelector('#arg-action-menu')?.textContent).toContain('Show reachable states');

    (sheet.querySelector('#arg-recalculate-all') as HTMLElement).click();
    vi.advanceTimersByTime(50);

    expect(runtime.recalculateAll).toHaveBeenCalledOnce();
  });

  it('recalculates the active sheet concept graph from the backend payload', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-2';
    const concept = { elements: [{ group: 'nodes', data: { id: 'node' } }] };
    runtime.api.executeAction = vi.fn(async () => ({ sheet_id: 'sheet-2', concept }));
    runtime.applyConceptSnapshot = vi.fn();
    runtime.refreshConceptGraph = vi.fn();

    await runtime.recalculateGraph();

    expect(runtime.api.executeAction).toHaveBeenCalledWith('recalculate', { sheet_id: 'sheet-2' });
    expect(runtime.applyConceptSnapshot).toHaveBeenCalledWith('sheet-2', concept);
    expect(runtime.refreshConceptGraph).not.toHaveBeenCalled();
    expect(runtime.controls.lastStatus).toEqual({ message: 'Recalculated', kind: 'success' });
  });

  it('keeps concrete state edges visible after clicking an ARG state', async () => {
    installSheetDom();
    const runtime = makeRuntime();
    runtime.argGraph = new FakeGraph('arg-graph');
    runtime.conceptGraph = new FakeGraph('concept-graph');
    runtime.registerSheet('sheet-1', runtime.argGraph, runtime.conceptGraph);
    runtime.sheets['sheet-1'].reachabilityOnly = true;
    runtime.uiDataModel.sheets['sheet-1'].reachabilityOnly = true;
    runtime._applyEdgeVisibility = vi.fn();
    runtime.api.getConceptGraph = vi.fn(async () => ({
      selected_node: 'state_1',
      state_label: '1',
      elements: [
        { group: 'nodes', data: { id: 'n0', obj: '0:client', label: '0:client', cluster: 'client' } },
        { group: 'nodes', data: { id: 'n1', obj: '1:client', label: '1:client', cluster: 'client' } },
        { group: 'nodes', data: { id: 'n2', obj: '0:server', label: '0:server', cluster: 'server' } },
        {
          group: 'edges',
          data: { id: 'e0', obj: 'link', source: 'n0', target: 'n2', source_obj: '0:client', target_obj: '0:server' },
          classes: 'all_to_all',
        },
        {
          group: 'edges',
          data: { id: 'e1', obj: 'link', source: 'n1', target: 'n2', source_obj: '1:client', target_obj: '0:server' },
          classes: 'all_to_all',
        },
      ],
      relations: ['link(X,Y)'],
      toggles: {
        edges: { link: { all_to_all: true } },
        labels: {},
      },
    }));

    await runtime.onArgNodeClick({ id: 'n1', obj: 'state_1', label: '1' }, 'sheet-1');

    expect(runtime.api.getConceptGraph).toHaveBeenCalledWith('state_1', 'sheet-1');
    const latestElements = runtime.conceptGraph.update.mock.calls.at(-1)[0];
    expect(latestElements.filter((element) => element.group === 'edges' && element.data.obj === 'link')).toHaveLength(2);
    const visibility = runtime._applyEdgeVisibility.mock.calls.at(-1)[1].edgeVisibilityById;
    expect(visibility.e0).toBe(true);
    expect(visibility.e1).toBe(true);
  });

  it('resizes the active reachability-only sheet details pane', () => {
    installSheetDom();
    const runtime = makeRuntime();
    runtime.setupDetailsResizer();
    runtime.addSheet('Error trace', 'sheet-2', { reachabilityOnly: true });

    const rootPanel = document.querySelector('#sheet-1 .info-panel') as HTMLElement;
    const tracePanel = document.querySelector('#sheet-2 .info-panel') as HTMLElement;
    const traceHeader = document.querySelector('#sheet-2 .info-header') as HTMLElement;
    const traceLeft = document.querySelector('#sheet-2 .sheet-left') as HTMLElement;

    expect(rootPanel.id).toBe('info-panel');
    expect(tracePanel.id).toBe('info-panel-2');
    Object.defineProperty(tracePanel, 'offsetHeight', { configurable: true, value: 120 });
    Object.defineProperty(traceLeft, 'offsetHeight', { configurable: true, value: 500 });

    traceHeader.dispatchEvent(new MouseEvent('mousedown', { clientY: 400, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mousemove', { clientY: 300, bubbles: true }));
    document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));

    expect(tracePanel.style.flex).toBe('0 0 220px');
    expect(tracePanel.style.height).toBe('220px');
    expect(rootPanel.style.height).toBe('');
  });

  it('refreshes file-loaded events through the primary analysis sheet, not the active reachability sheet', async () => {
    installSheetDom();
    const runtime = makeRuntime();
    runtime.argGraph = new FakeGraph('arg-graph');
    runtime.conceptGraph = new FakeGraph('concept-graph');
    runtime.registerSheet('sheet-1', runtime.argGraph, runtime.conceptGraph);
    runtime.api = {
      getARG: vi.fn(async () => ({ elements: [{ data: { id: 's0' } }] })),
      getConceptGraph: vi.fn(async () => ({
        elements: [{ data: { id: 'rel' } }],
        relations: ['link(X,Y)'],
      })),
    };

    runtime.addSheet('Reachable states', 'sheet-2', { reachabilityOnly: true });
    expect(runtime.activeSheetId).toBe('sheet-2');

    await runtime.refreshAfterLoad({ isolates: ['cf_live'], isolate: 'cf_live' });

    expect(runtime.activeSheetId).toBe('sheet-1');
    expect(runtime.sheets['sheet-1'].reachabilityOnly).toBe(false);
    expect(runtime.uiDataModel.sheets['sheet-1'].reachabilityOnly).toBe(false);
    expect(document.getElementById('sheet-1')?.classList.contains('reachability-only-sheet')).toBe(false);
    expect(document.getElementById('sheet-1')?.hasAttribute('data-sheet-layout')).toBe(false);
    expect(runtime.uiDataModel.sheets['sheet-1'].concept?.relations).toEqual(['link(X,Y)']);
    expect(runtime.sheets['sheet-2']).toBeUndefined();
  });

  it('defers state relation redraws until the current model load generation commits', () => {
    installSheetDom();
    document.getElementById('state-panel')!.innerHTML = [
      '<table id="state-checkbox-table">',
      '  <thead><tr><th>+</th><th>?</th><th>-</th><th>T</th><th>Relation</th></tr></thead>',
      '  <tbody id="state-checkbox-body"></tbody>',
      '</table>',
    ].join('');
    const runtime = makeRuntime();
    runtime.argGraph = new FakeGraph('arg-graph');
    runtime.conceptGraph = new FakeGraph('concept-graph');
    runtime.registerSheet('sheet-1', runtime.argGraph, runtime.conceptGraph);
    runtime.applyConceptSnapshot('sheet-1', {
      elements: [{ data: { id: 'old-node' } }],
      relations: ['old_relation'],
    });
    expect(document.querySelector('[data-state-toggle-relation]')?.textContent).toBe('old_relation');

    const load = runtime._beginModelLoad({ reason: 'test-load', filename: 'ord_live.ivy', content: 'ivy source' });
    runtime.applyConceptSnapshot('sheet-1', { elements: [], relations: [] });
    expect(document.querySelector('[data-state-toggle-relation]')?.textContent).toBe('old_relation');

    runtime._commitModelLoad(load, {
      argData: { elements: [{ data: { id: 'fresh-state' } }] },
      conceptData: { elements: [{ data: { id: 'fresh-node' } }], relations: ['fresh_relation'] },
      content: 'ivy source',
    });
    runtime._finishModelLoad(load);

    expect(document.querySelector('[data-state-toggle-relation]')?.textContent).toBe('fresh_relation');
    expect(runtime.uiDataModel.sheets['sheet-1'].concept?.relations).toEqual(['fresh_relation']);
  });

  it('allocates frontend-only sheet ids outside the backend sheet namespace', () => {
    installSheetDom();
    const runtime = makeRuntime();

    expect(runtime.nextLocalSheetId('trace')).toBe('trace-1');
    runtime.addSheet('Trace', 'trace-2', { reachabilityOnly: true });
    expect(runtime.nextLocalSheetId('trace')).toBe('trace-3');
  });

  it('reuses an existing sheet when backend reachable-state id collides with an old local trace id', async () => {
    installSheetDom();
    const runtime = makeRuntime();
    runtime.addSheet('Error trace', 'sheet-2', { reachabilityOnly: true });
    runtime.api.executeAction = vi.fn(async () => ({
      sheet_id: 'sheet-2',
      arg: {
        elements: [
          { group: 'nodes', data: { id: 'state_0', obj: 'state_0', label: '0' } },
        ],
      },
    }));

    await runtime.showReachableStates();

    expect(runtime.api.executeAction).toHaveBeenCalledWith('show_reachable', {});
    expect(runtime.activeSheetId).toBe('sheet-2');
    expect(runtime.tabLabelForSheet('sheet-2')).toBe('Reachable states');
    expect(runtime.sheets['sheet-2'].reachabilityOnly).toBe(true);
    expect(runtime.controls.lastStatus).toEqual({
      message: 'Reachable states opened',
      kind: 'success',
    });
  });

  it('runs PDR step through the active sheet and applies returned concept model state', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-7';
    runtime.api.executeAction = vi.fn(async () => ({
      sheet_id: 'sheet-7',
      message: 'PDR step diagrammed the predecessor goal.',
      concept: { sheet_id: 'sheet-7', graph: {}, elements: [] },
    }));
    runtime.applyConceptSnapshot = vi.fn();
    runtime.refreshConceptGraph = vi.fn();

    const result = await runtime.pdrStep();

    expect(runtime.api.executeAction).toHaveBeenCalledWith('pdr_step', { sheet_id: 'sheet-7' });
    expect(runtime.applyConceptSnapshot).toHaveBeenCalledWith('sheet-7', {
      sheet_id: 'sheet-7',
      graph: {},
      elements: [],
    });
    expect(runtime.refreshConceptGraph).not.toHaveBeenCalled();
    expect(runtime.controls.lastStatus).toEqual({
      message: 'PDR step diagrammed the predecessor goal.',
      kind: 'success',
    });
    expect(result.message).toBe('PDR step diagrammed the predecessor goal.');
  });

  it('resumes PDR step after interactive UPDR literal and core selections', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-7';
    runtime.api.executeAction = vi.fn(async (action, args) => {
      if (action === 'pdr_step' && !args.selection && !args.core) {
        return {
          sheet_id: 'sheet-7',
          status: 'needs_input',
          message: 'Choose which literals to take as the refutation goal',
          dialog: {
            type: 'select_multiple',
            title: 'Generalize Diagram',
            prompt: 'Choose which literals to take as the refutation goal',
            arg: 'selection',
            options: [
              { label: 'p(X)', value: 'p(X)' },
              { label: 'q(X)', value: 'q(X)' },
            ],
          },
          resume_action: 'pdr_step',
          resume_args: {
            sheet_id: 'sheet-7',
            interaction_id: 'iupdr-1',
          },
        };
      }
      if (action === 'pdr_step' && args.selection && !args.core) {
        return {
          sheet_id: 'sheet-7',
          status: 'needs_input',
          message: 'Choose the literals to use',
          dialog: {
            type: 'updr_select_core',
            title: 'Refinement',
            prompt: 'Choose the literals to use',
            options: [
              { label: 'p(X)', value: 'p(X)' },
              { label: 'q(X)', value: 'q(X)' },
            ],
          },
          resume_action: 'pdr_step',
          resume_args: {
            sheet_id: 'sheet-7',
            interaction_id: 'iupdr-1',
          },
        };
      }
      return {
        sheet_id: 'sheet-7',
        status: 'reversed',
        message: 'Refined with user selected core',
        concept: { sheet_id: 'sheet-7', graph: {}, elements: ['selected'] },
      };
    });
    runtime.listboxDialog = vi.fn(async (title) => (title === 'Refinement' ? ['q(X)'] : ['p(X)']));
    runtime.applyConceptSnapshot = vi.fn();
    runtime.refreshConceptGraph = vi.fn();

    const result = await runtime.pdrStep();

    expect(runtime.listboxDialog).toHaveBeenCalledWith(
      'Generalize Diagram',
      'Choose which literals to take as the refutation goal',
      [
        { label: 'p(X)', value: 'p(X)' },
        { label: 'q(X)', value: 'q(X)' },
      ],
      expect.objectContaining({ multiple: true, okLabel: 'OK' }),
    );
    expect(runtime.listboxDialog).toHaveBeenCalledWith(
      'Refinement',
      'Choose the literals to use',
      [
        { label: 'p(X)', value: 'p(X)' },
        { label: 'q(X)', value: 'q(X)' },
      ],
      expect.objectContaining({ multiple: true, okLabel: 'OK' }),
    );
    expect(runtime.api.executeAction).toHaveBeenNthCalledWith(1, 'pdr_step', { sheet_id: 'sheet-7' });
    expect(runtime.api.executeAction).toHaveBeenNthCalledWith(2, 'pdr_step', {
      sheet_id: 'sheet-7',
      interaction_id: 'iupdr-1',
      selection: ['p(X)'],
    });
    expect(runtime.api.executeAction).toHaveBeenNthCalledWith(3, 'pdr_step', {
      sheet_id: 'sheet-7',
      interaction_id: 'iupdr-1',
      core: ['q(X)'],
    });
    expect(runtime.applyConceptSnapshot).toHaveBeenCalledWith('sheet-7', {
      sheet_id: 'sheet-7',
      graph: {},
      elements: ['selected'],
    });
    expect(runtime.controls.lastStatus).toEqual({
      message: 'Refined with user selected core',
      kind: 'success',
    });
    expect(result?.message).toBe('Refined with user selected core');
  });

  it('shows eliminated conjectures after one-step reach', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-7';
    runtime.api.executeAction = vi.fn(async () => ({
      reachable: true,
      eliminated_conjectures_message: 'The following conjectures have been eliminated:',
      eliminated_conjectures: ['p', 'q'],
    }));
    runtime.refreshConceptGraph = vi.fn(async () => undefined);
    runtime.listboxDialog = vi.fn(async () => null);

    await runtime.reachStep();

    expect(runtime.api.executeAction).toHaveBeenCalledWith('reach', { sheet_id: 'sheet-7' });
    expect(runtime.listboxDialog).toHaveBeenCalledWith(
      'Reach',
      'The following conjectures have been eliminated:',
      [
        { label: 'p', value: 'p' },
        { label: 'q', value: 'q' },
      ],
      { cancel: false },
    );
    expect(runtime.controls.lastStatus).toEqual({
      message: 'Reach complete',
      kind: 'success',
    });
  });

  it('submits accepted PDR interpolants through a Refine dialog', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-7';
    runtime.api.executeAction = vi.fn(async (action) => {
      if (action === 'pdr_step') {
        return {
          sheet_id: 'sheet-7',
          status: 'refinement_suggested',
          message: 'The pre-state is vacuous. The following predicate can be used to prove your goal in the post-state:',
          interpolant: 'p(X)',
          refinement_action: 'refine_with_interpolant',
          refinement_kind: 'predicate',
          concept: { sheet_id: 'sheet-7', graph: {}, elements: [] },
        };
      }
      return {
        sheet_id: 'sheet-7',
        message: 'Refinement applied.',
        concept: { sheet_id: 'sheet-7', graph: {}, elements: ['refined'] },
      };
    });
    runtime.textDialog = vi.fn(async () => 'p(X)');
    runtime.applyConceptSnapshot = vi.fn();
    runtime.refreshConceptGraph = vi.fn();

    await runtime.pdrStep();

    expect(runtime.textDialog).toHaveBeenCalledWith(
      'ivyweb',
      'The pre-state is vacuous. The following predicate can be used to prove your goal in the post-state:',
      'p(X)',
      expect.objectContaining({ okLabel: 'Refine', cancel: true, primaryFirst: true }),
    );
    expect(runtime.api.executeAction).toHaveBeenCalledWith('refine_with_interpolant', {
      sheet_id: 'sheet-7',
      interpolant: 'p(X)',
    });
    expect(runtime.applyConceptSnapshot).toHaveBeenCalledWith('sheet-7', {
      sheet_id: 'sheet-7',
      graph: {},
      elements: ['refined'],
    });
  });

  it('leaves the graph alone when Diagram Domain would be empty', async () => {
    const runtime = makeRuntime();
    runtime.activeSheetId = 'sheet-1';
    runtime.api.executeAction = vi.fn(async () => ({
      sheet_id: 'sheet-1',
      type: 'diagram_domain_empty',
      status: 'warning',
      message: "no first-order constants in 'client_server_example.ivy' found. Diagram Domain would give an empty graph. Leaving existing graph alone.",
    }));
    runtime.applyConceptSnapshot = vi.fn();
    runtime.refreshConceptGraph = vi.fn();

    const result = await runtime.diagramDomain();

    expect(runtime.api.executeAction).toHaveBeenCalledWith('diagram_domain', { sheet_id: 'sheet-1' });
    expect(runtime.applyConceptSnapshot).not.toHaveBeenCalled();
    expect(runtime.refreshConceptGraph).not.toHaveBeenCalled();
    expect(runtime.controls.lastStatus).toEqual({
      message: "no first-order constants in 'client_server_example.ivy' found. Diagram Domain would give an empty graph. Leaving existing graph alone.",
      kind: 'warning',
    });
    expect(result.type).toBe('diagram_domain_empty');
  });
});
