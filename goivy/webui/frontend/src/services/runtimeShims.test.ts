import { afterEach, describe, expect, it } from 'vitest';
import { IvyControlsShim } from './runtimeShims.ts';

afterEach(() => {
  document.body.innerHTML = '';
});

describe('IvyControlsShim', () => {
  it('writes graph details to the active sheet details pane', () => {
    document.body.innerHTML = [
      '<div id="sheet-1" class="sheet-content">',
      '  <div class="info-panel"><div id="info-content">root details</div></div>',
      '</div>',
      '<div id="sheet-3" class="sheet-content active reachability-only-sheet">',
      '  <div class="info-panel"><div id="info-content-3">trace details</div></div>',
      '</div>',
    ].join('');
    const controls = new IvyControlsShim({});

    controls.showInfo('state 0', ['transition init']);

    expect(document.getElementById('info-content')?.textContent).toBe('root details');
    expect(document.getElementById('info-content-3')?.textContent).toBe('state 0\ntransition init');
    expect(document.getElementById('info-content-3')?.getAttribute('data-ivy-details-kind')).toBe('selection');

    controls.clearInfo();

    expect(document.getElementById('info-content')?.textContent).toBe('root details');
    expect(document.getElementById('info-content-3')?.textContent).toBe('Select a node or edge to see details');
    expect(document.getElementById('info-content-3')?.getAttribute('data-ivy-details-kind')).toBe('placeholder');
  });
});
