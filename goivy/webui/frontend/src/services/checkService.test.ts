import { describe, expect, it, vi } from 'vitest';
import {
  addCheckResultViewActions,
  autoCheckUsedRelations,
  boundedCheck,
  ctiConceptAction,
  showCheckResult,
  weakenInvariant,
} from './checkService.ts';

describe('checkService', () => {
  it('adds CTI details and trace actions for failed checks', () => {
    const app = {
      addCheckResultViewActions: vi.fn(),
      argGraph: { update: vi.fn() },
      controls: {
        setStatus: vi.fn(),
        showInfo: vi.fn(),
      },
    };

    showCheckResult(app, {
      result: 'fail',
      mode: 'induction',
      z3_contacted: true,
      message: 'Counterexample found.',
      failed_conjecture: 'bad conjecture',
      counterexample_details: 'frame details',
      arg: { elements: [1], positions: {} },
    });

    expect(app.controls.setStatus).toHaveBeenCalledWith(
      'Check FAILED (induction) [Z3: yes] - counterexample found',
      'error',
    );
    expect(app.controls.showInfo.mock.calls[0][1]).toContain('frame details');
    expect(app.addCheckResultViewActions).toHaveBeenCalled();
    expect(app.argGraph.update).toHaveBeenCalledWith([1], {});
  });

  it('adds a direct DOM view-trace action', () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const app = {
      openARGSheet: vi.fn(),
    };
    const result = {
      trace_arg: { elements: [] },
      trace_sheet_id: 'trace-1',
    };

    addCheckResultViewActions(app, result, { doc: document });
    document.querySelector('[data-check-view-trace]').click();

    expect(app.openARGSheet).toHaveBeenCalledWith('Error trace', result.trace_arg, 'trace-1');
  });

  it('prompts for a bounded-check bound and sends it to the backend', async () => {
    const app = {
      currentBound: 3,
      integerDialog: vi.fn(async () => 7),
      api: {
        runCheck: vi.fn(async () => ({ result: 'pass' })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await boundedCheck(app);

    expect(app.integerDialog).toHaveBeenCalledWith('Bounded check', 'Enter bound:', 3, {
      min: 1,
      okLabel: 'OK',
    });
    expect(app.api.runCheck).toHaveBeenCalledWith('bounded', { bound: 7 });
    expect(app.currentBound).toBe(7);
  });

  it('prompts for conjectures before weakening', async () => {
    const app = {
      refreshConceptGraph: vi.fn(),
      listboxDialog: vi.fn(async () => [0]),
      api: {
        executeAction: vi.fn(async (action) => {
          if (action === 'get_conjectures') {
            return {
              conjectures: [
                { label: 'c0', formula: 'p(X)' },
                { label: 'c1', formula: 'q(X)' },
              ],
            };
          }
          return { removed: ['p(X)'], removed_count: 1 };
        }),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await weakenInvariant(app);

    expect(app.api.executeAction).toHaveBeenNthCalledWith(1, 'get_conjectures', {});
    expect(app.listboxDialog).toHaveBeenCalledWith(
      'Weaken',
      'Choose conjectures to remove:',
      [
        { label: '[c0] p(X)', value: 0 },
        { label: '[c1] q(X)', value: 1 },
      ],
      { multiple: true, okLabel: 'Weaken' },
    );
    expect(app.api.executeAction).toHaveBeenNthCalledWith(2, 'weaken', { indices: [0] });
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Invariant weakened', 'success');
  });

  it('dispatches CTI concept graph actions with the active sheet id', async () => {
    const app = {
      activeSheetId: 'sheet-2',
      refreshConceptGraph: vi.fn(),
      api: {
        executeAction: vi.fn(async () => ({ ok: true, message: 'strengthened' })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await ctiConceptAction(app, 'cti_strengthen');

    expect(app.api.executeAction).toHaveBeenCalledWith('cti_strengthen', { sheet_id: 'sheet-2' });
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('strengthened', 'success');
  });

  it('auto-checks used relation rows', async () => {
    document.body.innerHTML = [
      '<table><tbody id="state-checkbox-body">',
      '<tr><td class="name-col"><a>link(X,Y)</a></td><td><input type="checkbox"></td></tr>',
      '</tbody></table>',
    ].join('');
    const row = document.querySelector('tr');
    row.appendChild(row.firstElementChild);
    const app = {
      onEdgeToggle: vi.fn(),
    };

    await autoCheckUsedRelations(app, ['link'], document);

    expect(document.querySelector('input').checked).toBe(true);
    expect(app.onEdgeToggle).toHaveBeenCalledWith('link(X,Y)', 'all_to_all', true);
  });
});
