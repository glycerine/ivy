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

  it('feeds the Vue recent-files store through the bridge', () => {
    const app = {
      loadRecentSession: vi.fn(),
    };
    const bridge = {
      updateRecentFiles: vi.fn(),
    };
    const persist = {
      listSessions: vi.fn(() => [
        { id: 's1', fileName: 'client.ivy', filePath: 'client.ivy' },
      ]),
      truncatePath: vi.fn((path) => path),
    };

    populateRecentFiles(app, persist, { bridge });

    expect(bridge.updateRecentFiles).toHaveBeenCalledWith(
      [{ id: 's1', label: 'client.ivy', title: '' }],
      expect.any(Function),
    );
    bridge.updateRecentFiles.mock.calls[0][1]('s1');
    expect(app.loadRecentSession).toHaveBeenCalledWith('s1');
  });
});
