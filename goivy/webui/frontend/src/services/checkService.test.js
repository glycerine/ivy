import { describe, expect, it, vi } from 'vitest';
import { addCheckResultViewActions, autoCheckUsedRelations, showCheckResult } from './checkService.js';

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

  it('wires the view-trace action through Vue when available', () => {
    const app = {
      openARGSheet: vi.fn(),
    };
    const bridge = {
      setCheckTraceAction: vi.fn(),
    };
    const result = {
      trace_arg: { elements: [] },
      trace_sheet_id: 'trace-1',
    };

    addCheckResultViewActions(app, result, { bridge });
    bridge.setCheckTraceAction.mock.calls[0][0]();

    expect(app.openARGSheet).toHaveBeenCalledWith('Error trace', result.trace_arg, 'trace-1');
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
