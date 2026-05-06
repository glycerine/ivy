import { afterEach, describe, expect, it, vi } from 'vitest';
import { loadIvyPersist } from './helpers/load_browser_scripts.mjs';

afterEach(() => {
  delete window.__ivyVueBridge;
  document.body.innerHTML = '';
});

describe('IvyPersist state relation snapshots', () => {
  it('uses the Vue bridge for relation toggles and visibility', () => {
    const IvyPersist = loadIvyPersist();
    window.__ivyVueBridge = {
      getStateRelationToggles: vi.fn(() => ({ 'link|all_to_all': true })),
      setStateRelationToggles: vi.fn(),
      buildStateRelationVisibility: vi.fn(() => ({
        edges: { link: { all_to_all: true } },
        labels: { link: { node_necessarily: true } },
      })),
    };

    expect(IvyPersist._getToggles()).toEqual({ 'link|all_to_all': true });

    IvyPersist._setToggles({ 'link|edge_unknown': true });
    expect(window.__ivyVueBridge.setStateRelationToggles).toHaveBeenCalledWith({ 'link|edge_unknown': true });
    expect(IvyPersist._buildVisibilityFromCheckboxes()).toEqual({
      edges: { link: { all_to_all: true } },
      labels: { link: { node_necessarily: true } },
    });
  });

  it('keeps the named-checkbox fallback for non-Vue harnesses', () => {
    const IvyPersist = loadIvyPersist();
    document.body.innerHTML = [
      '<table><tbody id="state-checkbox-body">',
      '  <tr>',
      '    <td><input type="checkbox" name="link" value="all_to_all" checked></td>',
      '    <td><input type="checkbox" name="link" value="edge_unknown"></td>',
      '    <td><input type="checkbox" name="link" value="none_to_none" checked></td>',
      '    <td><input type="checkbox" name="link" value="transitive"></td>',
      '    <td class="name-col"><a href="#">link</a></td>',
      '  </tr>',
      '</tbody></table>',
    ].join('');

    expect(IvyPersist._getToggles()).toEqual({
      'link|all_to_all': true,
      'link|edge_unknown': false,
      'link|none_to_none': true,
      'link|transitive': false,
    });

    IvyPersist._setToggles({ 'link|edge_unknown': true, 'link|none_to_none': false });

    expect(IvyPersist._buildVisibilityFromCheckboxes()).toEqual({
      edges: {
        link: {
          all_to_all: true,
          edge_unknown: true,
          none_to_none: false,
          transitive: false,
        },
      },
      labels: {
        link: {
          node_necessarily: true,
          node_maybe: true,
          node_necessarily_not: false,
        },
      },
    });
  });
});
