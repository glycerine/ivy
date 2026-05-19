import { describe, expect, it, vi } from 'vitest';
import { populateRecentFiles, recentFileItems } from './recentFileService.ts';

describe('recentFileService', () => {
  it('deduplicates recent files and adds path context for colliding basenames', () => {
    const sessions = [
      { id: 's1', fileName: 'client.ivy', filePath: '/tmp/a/client.ivy', timestamp: 1 },
      { id: 's2', fileName: 'client.ivy', filePath: '/tmp/b/client.ivy', timestamp: 2 },
      { id: 's3', fileName: 'client.ivy', filePath: '/tmp/a/client.ivy', timestamp: 3 },
      { id: 'ignored', fileName: '(unnamed)' },
    ];

    const items = recentFileItems(sessions, (path) => `...${path.slice(-6)}`);

    expect(items.map((item) => item.id)).toEqual(['s1', 's2']);
    expect(items[0].label).toContain('...nt.ivy');
    expect(items[1].label).toContain('...nt.ivy');
  });

  it('keeps the six most recent files in storage order', () => {
    const sessions = [
      { id: 'current', fileName: 'zeta.ivy', filePath: '/tmp/zeta.ivy', timestamp: 7 },
      { id: 's6', fileName: 'alpha.ivy', filePath: '/tmp/alpha.ivy', timestamp: 6 },
      { id: 's5', fileName: 'beta.ivy', filePath: '/tmp/beta.ivy', timestamp: 5 },
      { id: 's4', fileName: 'gamma.ivy', filePath: '/tmp/gamma.ivy', timestamp: 4 },
      { id: 's3', fileName: 'delta.ivy', filePath: '/tmp/delta.ivy', timestamp: 3 },
      { id: 's2', fileName: 'epsilon.ivy', filePath: '/tmp/epsilon.ivy', timestamp: 2 },
      { id: 's1', fileName: 'old.ivy', filePath: '/tmp/old.ivy', timestamp: 1 },
    ];

    const items = recentFileItems(sessions, (path) => path);

    expect(items.map((item) => item.id)).toEqual(['current', 's6', 's5', 's4', 's3', 's2']);
  });

  it('renders recent files into the DOM and opens the selected session', () => {
    document.body.innerHTML = '<div id="file-recent-list"></div>';
    const app = {
      loadRecentSession: vi.fn(),
    };
    const persist = {
      listSessions: vi.fn(() => [
        { id: 's1', fileName: 'client.ivy', filePath: 'client.ivy' },
      ]),
      truncatePath: vi.fn((path) => path),
    };

    populateRecentFiles(app, persist, { doc: document });

    const item = document.querySelector('[data-session-id="s1"]');
    expect(item.textContent).toBe('client.ivy');
    item.click();
    expect(app.loadRecentSession).toHaveBeenCalledWith('s1');
  });
});
