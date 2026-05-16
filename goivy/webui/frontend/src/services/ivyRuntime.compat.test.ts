import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  IvyRuntime,
  configureIvyRuntimeDependencies,
  resetIvyRuntimeDependencies,
} from './ivyRuntime.ts';
import { FakeAPI, FakeControls, FakeGraph } from '../test/fakes.ts';

function makeRuntime() {
  resetIvyRuntimeDependencies();
  configureIvyRuntimeDependencies({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
  return new IvyRuntime();
}

afterEach(() => {
  resetIvyRuntimeDependencies();
  document.body.innerHTML = '';
  vi.useRealTimers();
});

describe('ivyRuntime compatibility behavior', () => {
  it('shows a direct DOM toast notification', () => {
    const runtime = makeRuntime();

    runtime._showToast('Connection lost', 'error', { persistent: true, className: 'custom-toast' });

    const toast = document.querySelector('.ivy-toast');
    expect(toast.textContent).toBe('Connection lost');
    expect(toast.classList.contains('ivy-toast-floating')).toBe(true);
    expect(toast.classList.contains('ivy-toast-error')).toBe(true);
    expect(toast.classList.contains('custom-toast')).toBe(true);
  });

  it('opens job control and toggles browser/remote submission mode', () => {
    document.body.innerHTML = [
      '<button id="btn-toggle-job-control"></button>',
      '<section id="job-control-page" aria-hidden="true"></section>',
      '<button id="job-control-close"></button>',
      '<button id="job-submission-toggle" class="job-mode-toggle is-browser" data-mode="browser"></button>',
      '<span id="job-submission-label"></span>',
    ].join('');
    const runtime = makeRuntime();

    runtime._setupJobControlHandlers();
    expect(document.getElementById('btn-toggle-job-control').classList.contains('job-submission-browser')).toBe(true);
    expect(document.getElementById('btn-toggle-job-control').getAttribute('data-job-submission-mode')).toBe('browser');
    expect(document.getElementById('job-submission-label').textContent).toBe('run in browser');

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
    expect(Array.from(document.querySelectorAll('#isolate-menu a')).map((item) => item.textContent)).toEqual([
      'cf_backup',
      'cf_live',
    ]);
    expect(document.querySelector('.sheet-tab span')?.textContent).toBe('Sheet 1 · cf_live');

    runtime.setIsolates(['cf_backup', 'cf_live'], 'cf_backup');
    expect(document.querySelector('.sheet-tab span')?.textContent).toBe('Sheet 1 · cf_backup');
  });

  it('opens reachability-only sheets without a concept graph runtime', () => {
    vi.useFakeTimers();
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
      '        <div id="info-panel" class="info-panel"><div id="info-content"></div></div>',
      '      </div>',
      '      <div id="divider2" class="divider"></div>',
      '      <div id="state-panel" class="panel"></div>',
      '    </div>',
      '  </div>',
      '</div>',
    ].join('');
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
});
