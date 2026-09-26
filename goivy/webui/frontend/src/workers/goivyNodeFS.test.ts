import { afterEach, describe, expect, it } from 'vitest';
import { installGoIvyNodeFS } from './goivyNodeFS.js';

function stat(path: string): Promise<any> {
  return new Promise((resolve, reject) => {
    (globalThis as any).fs.stat(path, (err, result) => {
      if (err) reject(err);
      else resolve(result);
    });
  });
}

function open(path: string): Promise<number> {
  return new Promise((resolve, reject) => {
    (globalThis as any).fs.open(path, 0, 0, (err, fd) => {
      if (err) reject(err);
      else resolve(fd);
    });
  });
}

function read(fd: number, length: number): Promise<string> {
  const buffer = new Uint8Array(length);
  return new Promise((resolve, reject) => {
    (globalThis as any).fs.read(fd, buffer, 0, length, 0, (err, n) => {
      if (err) reject(err);
      else resolve(new TextDecoder().decode(buffer.subarray(0, n)));
    });
  });
}

describe('goivy Node fs shim', () => {
  const originalFs = (globalThis as any).fs;
  const originalProcess = (globalThis as any).process;
  const originalPath = (globalThis as any).path;

  afterEach(() => {
    (globalThis as any).fs = originalFs;
    (globalThis as any).process = originalProcess;
    (globalThis as any).path = originalPath;
  });

  it('mounts authorized project files beside the bundled include tree', async () => {
    const fsHost = installGoIvyNodeFS({
      includeRoot: '/include',
      includeTree: { files: [{ path: 'stdlib.ivy', data: '#lang ivy1.7\n' }] },
      projectRoot: '/project',
    });

    fsHost.setProjectFiles([
      { path: 'raft_no_assume_test.ivy', data: '#lang ivy1.6\ninclude raft_no_assume\n' },
      { path: 'raft_no_assume.ivy', data: '#lang ivy1.6\n' },
    ]);

    await expect(stat('/include/stdlib.ivy')).resolves.toMatchObject({ size: '#lang ivy1.7\n'.length });
    await expect(stat('/project/raft_no_assume.ivy')).resolves.toMatchObject({ size: '#lang ivy1.6\n'.length });
    const fd = await open('/project/raft_no_assume_test.ivy');
    await expect(read(fd, 128)).resolves.toBe('#lang ivy1.6\ninclude raft_no_assume\n');
  });
});
