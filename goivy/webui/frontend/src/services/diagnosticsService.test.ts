import { afterEach, describe, expect, it, vi } from 'vitest';
import { registerCommand, resetCommandRegistry } from './commandRegistry.ts';
import { createIvyDiagnostics, installIvyDiagnostics } from './diagnosticsService.ts';

afterEach(() => {
  delete window.__ivyDiagnostics;
  resetCommandRegistry();
});

describe('diagnosticsService', () => {
  it('exposes runtime snapshots without exposing the app as the API', () => {
    const graph = {
      cy: {
        nodes: () => ({ length: 2 }),
        elements: () => ({
          jsons: () => [{ data: { id: 'n1' } }],
        }),
      },
    };
    const runtime = {
      api: { sessionId: 'sess-1' },
      conceptGraph: graph,
      cmEditor: { getValue: () => 'type client' },
      activeSheetId: 'sheet-1',
      sheets: { 'sheet-1': { label: 'main' } },
    };
    const diagnostics = createIvyDiagnostics({
      services: { runtime: () => runtime },
    });

    expect(diagnostics.sessionId()).toBe('sess-1');
    expect(diagnostics.graphNodeCount()).toBe(2);
    expect(diagnostics.graphSnapshot()).toEqual([{ data: { id: 'n1' } }]);
    expect(diagnostics.editorContent()).toBe('type client');
    expect(diagnostics.sheetSnapshot().activeSheetId).toBe('sheet-1');
  });

  it('installs and removes the browser diagnostics bridge', () => {
    const remove = installIvyDiagnostics({
      win: window,
      services: { runtime: () => null },
    });

    expect(window.__ivyDiagnostics).toBeDefined();
    remove();
    expect(window.__ivyDiagnostics).toBeUndefined();
  });

  it('routes diagnostic commands through the command registry', () => {
    const command = vi.fn(() => 'ok');
    registerCommand('demo.command', command);
    const diagnostics = createIvyDiagnostics({
      services: { runtime: () => null },
    });

    expect(diagnostics.runCommand('demo.command', 1)).toBe('ok');
    expect(command).toHaveBeenCalledWith(1);
  });
});
