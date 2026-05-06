import { afterEach, describe, expect, it } from 'vitest';
import { loadIvyControls } from './helpers/load_browser_scripts.mjs';

afterEach(() => {
  document.body.innerHTML = '';
});

describe('IvyControls legacy toggle containers', () => {
  it('does not crash when Vue no longer renders old edge/label toggle containers', () => {
    const IvyControls = loadIvyControls();
    const controls = new IvyControls({});

    expect(() => controls.buildEdgeToggles(['link'], () => {})).not.toThrow();
    expect(() => controls.buildLabelToggles(['semaphore'], () => {})).not.toThrow();
    expect(controls.edgeToggles).toEqual({});
    expect(controls.labelToggles).toEqual({});
  });
});
