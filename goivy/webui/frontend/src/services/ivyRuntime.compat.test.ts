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
});
