import { describe, expect, it, vi } from 'vitest';
import { populateRecentFiles, recentFileItems } from './recentFileService.js';

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
