import { beforeEach, describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import {
  FakeAPI,
  FakeControls,
  FakeGraph,
  installSaveDom,
  makeApp,
} from './helpers/fakes.mjs';

function loadApp() {
  return loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
}

beforeEach(() => {
  installSaveDom();
});

describe('IvyApp source browsing', () => {
  it('loads returned source into the editor and scrolls to the backend line', async () => {
    const IvyApp = loadApp();
    const app = makeApp(IvyApp);
    app.api = {
      argNodeAction: vi.fn(async () => ({
        status: 'ok',
        file: 'sample.ivy',
        lineno: 2,
        source: 'line1\naction go = {}\n',
      })),
    };
    app.setEditorContent = vi.fn();
    app.scrollEditorToLine = vi.fn();

    await app.executeArgEdgeAction({ source_obj: 'state_0', target_obj: 'state_1' }, 'view_source');

    expect(app.api.argNodeAction).toHaveBeenCalledWith('state_0', 'view_source', { target: 'state_1', sheet_id: 'sheet-1' });
    expect(app.setEditorContent).toHaveBeenCalledWith('line1\naction go = {}\n');
    expect(app.scrollEditorToLine).toHaveBeenCalledWith(2);
    expect(app.controls.lastInfo.shortInfo).toContain('sample.ivy');
  });
});
