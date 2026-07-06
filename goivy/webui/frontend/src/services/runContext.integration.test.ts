import { describe, expect, it, vi } from 'vitest';
import { executeArgNodeAction } from './argActionService.ts';
import { checkInduction, ctiConceptAction } from './checkService.ts';
import { executeConceptNodeAction } from './conceptActionService.ts';
import { activeEventSheet, filterEventTrace } from './eventTraceService.ts';

function makeControls() {
  return {
    setStatus: vi.fn(),
    showLoading: vi.fn(),
    hideLoading: vi.fn(),
  };
}

function makeErrorApp(prefix: string, overrides = {}) {
  return {
    activeSheetId: 'sheet-1',
    controls: makeControls(),
    textDialog: vi.fn(async () => null),
    api: {},
    ...overrides,
    expectedRunContextMessage: `${prefix}: backend exploded`,
  };
}

async function expectRunContextError(app: any, run: () => Promise<any>) {
  const result = await run();

  expect(result).toBeNull();
  expect(app.controls.setStatus).toHaveBeenCalledWith(app.expectedRunContextMessage, 'error');
  expect(app.controls.showLoading).toHaveBeenCalled();
  expect(app.controls.hideLoading).toHaveBeenCalled();
  expect(document.body.hasAttribute('data-ivy-run-context')).toBe(false);
  expect(document.body.getAttribute('aria-busy')).not.toBe('true');
  expect(app.textDialog).toHaveBeenCalledWith(
    'ivyweb',
    'Ivy error',
    app.expectedRunContextMessage,
    expect.objectContaining({ readOnly: true, okLabel: 'OK', cancel: false }),
  );
}

describe('run context error handling', () => {
  it('shows modal Ivy errors and restores ready state across command families', async () => {
    const inductionApp = makeErrorApp('Induction check failed', {
      api: {
        runCheck: vi.fn(async () => {
          throw new Error('backend exploded');
        }),
      },
    });
    await expectRunContextError(inductionApp, () => checkInduction(inductionApp));

    const argApp = makeErrorApp('Action failed', {
      isVisualOnlySheet: vi.fn(() => false),
      prepareArgNodeActionArgs: vi.fn(async (_node, _action, args) => args),
      api: {
        argNodeAction: vi.fn(async () => {
          throw new Error('backend exploded');
        }),
      },
      visualOnlyMessage: vi.fn(() => 'visual only'),
    });
    await expectRunContextError(argApp, () => executeArgNodeAction(argApp, { id: 'state_0' }, { id: 'extend' }, 'sheet-1'));

    const conceptApp = makeErrorApp('Concept action failed', {
      api: {
        executeAction: vi.fn(async () => {
          throw new Error('backend exploded');
        }),
      },
      refreshConceptGraph: vi.fn(),
    });
    await expectRunContextError(conceptApp, () => executeConceptNodeAction(conceptApp, { id: 'server' }, { id: 'highlight' }));

    const ctiApp = makeErrorApp('CTI action failed', {
      api: {
        executeAction: vi.fn(async () => {
          throw new Error('backend exploded');
        }),
      },
      refreshConceptGraph: vi.fn(),
    });
    await expectRunContextError(ctiApp, () => ctiConceptAction(ctiApp, 'cti_gather'));

    const eventApp = makeErrorApp('Filter failed', {
      activeSheetId: 'events-1',
      sheets: {
        'events-1': { id: 'events-1', type: 'events', events: [] },
      },
      activeEventSheet() {
        return activeEventSheet(this);
      },
      visualOnlyMessage: vi.fn(() => 'visual only'),
      api: {
        executeAction: vi.fn(async () => {
          throw new Error('backend exploded');
        }),
      },
      openEventTraceSheet: vi.fn(),
    });
    await expectRunContextError(eventApp, () => filterEventTrace(eventApp, 'call('));
  });
});
