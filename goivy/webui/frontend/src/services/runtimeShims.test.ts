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

  it('does not repeat details when short_info and long_info are identical', () => {
    document.body.innerHTML = '<div class="sheet-content active"><div class="info-panel"><div id="info-content"></div></div></div>';
    const controls = new IvyControlsShim({});

    controls.showInfo('State 0', 'State 0');

    expect(document.getElementById('info-content')?.textContent).toBe('State 0');
  });

  it('deduplicates repeated long_info lines while keeping distinct details', () => {
    document.body.innerHTML = '<div class="sheet-content active"><div class="info-panel"><div id="info-content"></div></div></div>';
    const controls = new IvyControlsShim({});

    controls.showInfo('State 0', ['State 0', 'init', 'init']);

    expect(document.getElementById('info-content')?.textContent).toBe('State 0\ninit');
  });

  it('does not replace rendered transition constraints with a raw clause dump', () => {
    document.body.innerHTML = [
      '<div class="sheet-content active">',
      '  <div class="info-panel">',
      '    <div id="info-content" data-ivy-details-kind="constraints">',
      '      <div class="constraint-context-label">connect(0:client, 0:server)</div>',
      '      <div class="constraint-facts-title">Constraints:</div>',
      '      <button class="constraint-fact">link(0,0)</button>',
      '    </div>',
      '  </div>',
      '</div>',
    ].join('');
    const controls = new IvyControlsShim({});

    controls.showInfo('Clauses{defs=[__semaphore(V0) = true]}', 'Clauses{defs=[link(V0,V1) = false]}');

    const text = document.getElementById('info-content')?.textContent || '';
    expect(text).toContain('connect(0:client, 0:server)');
    expect(text).toContain('Constraints:');
    expect(text).toContain('link(0,0)');
    expect(text).not.toContain('Clauses{');
    expect(document.getElementById('info-content')?.getAttribute('data-ivy-details-kind')).toBe('constraints');
  });
});
