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

  it('opens reachability-only sheets without a concept graph runtime', () => {
    document.body.innerHTML = [
      '<div id="sheet-area">',
      '  <div id="tab-bar"><button class="sheet-tab active" data-sheet="sheet-1"><span>Sheet 1</span></button></div>',
      '  <div id="sheet-1" class="sheet-content active">',
      '    <div class="sheet-columns">',
      '      <div class="sheet-left">',
      '        <div class="sheet-main">',
      '          <div id="arg-panel" class="panel"><div id="arg-graph" class="graph-container"></div></div>',
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

    const sheetId = runtime.addSheet('Sheet 3', 'sheet-3', { reachabilityOnly: true });
    const sheet = document.getElementById('sheet-3');

    expect(sheetId).toBe('sheet-3');
    expect(sheet.classList.contains('reachability-only-sheet')).toBe(true);
    expect(sheet.getAttribute('data-sheet-layout')).toBe('reachability-only');
    expect(runtime.sheets['sheet-3'].reachabilityOnly).toBe(true);
    expect(runtime.sheets['sheet-3'].argGraph).toBeTruthy();
    expect(runtime.sheets['sheet-3'].conceptGraph).toBeNull();
  });
});
