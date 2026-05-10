import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, describe, expect, test } from 'vitest';

import { createGoIvyTinyGoWasiP1, GOIVY_WASI_ERRNO } from './goivyTinyGoWasiP1.js';

const enc = new TextEncoder();
const dec = new TextDecoder();

let tmpDirs = [];

afterEach(() => {
  for (const dir of tmpDirs) {
    fs.rmSync(dir, { recursive: true, force: true });
  }
  tmpDirs = [];
});

function newHarness(options = {}) {
  const memory = new WebAssembly.Memory({ initial: 1 });
  const wasi = createGoIvyTinyGoWasiP1({
    includeRoot: '/include',
    includeTree: { files: [] },
    ...options,
  });
  wasi.setInstance({ exports: { memory } });
  const mem = new Uint8Array(memory.buffer);
  const view = new DataView(memory.buffer);
  return { wasi, mem, view };
}

function putString(mem, ptr, text) {
  const data = enc.encode(text);
  mem.set(data, ptr);
  return data.byteLength;
}

function putIovec(view, ptr, dataPtr, len) {
  view.setUint32(ptr, dataPtr, true);
  view.setUint32(ptr + 4, len, true);
}

describe('goivy TinyGo WASI preview1 host', () => {
  test('keeps browser mode limited to the in-memory include preopen', () => {
    const { wasi } = newHarness();
    expect(wasi.wasiImport.fd_prestat_get(4, 1024)).toBe(GOIVY_WASI_ERRNO.BADF);
  });

  test('can read and write host files when tinynode passes Node fs hooks', () => {
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'goivy-tinygo-wasi-'));
    tmpDirs.push(dir);
    const readPath = path.join(dir, 'read.ivy');
    const writePath = path.join(dir, 'write.log');
    fs.writeFileSync(readPath, 'hello tinygo wasi fs');

    const { wasi, mem, view } = newHarness({
      nodeFilesystem: fs,
      nodePath: path,
      hostPreopenPath: '/',
      hostPreopenName: '/',
    });

    let rel = readPath.slice(1);
    let len = putString(mem, 1024, rel);
    let ret = wasi.wasiImport.path_open(4, 0, 1024, len, 0, 2n, 0n, 0, 2048);
    expect(ret).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const readFd = view.getUint32(2048, true);

    putIovec(view, 3000, 4000, 64);
    ret = wasi.wasiImport.fd_read(readFd, 3000, 1, 5000);
    expect(ret).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const nread = view.getUint32(5000, true);
    expect(dec.decode(mem.subarray(4000, 4000 + nread))).toBe('hello tinygo wasi fs');
    expect(wasi.wasiImport.fd_close(readFd)).toBe(GOIVY_WASI_ERRNO.SUCCESS);

    rel = writePath.slice(1);
    len = putString(mem, 1100, rel);
    const oflagsCreateTrunc = 1 | 8;
    const rightsFdWrite = 1n << 6n;
    ret = wasi.wasiImport.path_open(4, 0, 1100, len, oflagsCreateTrunc, rightsFdWrite, 0n, 0, 2048);
    expect(ret).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const writeFd = view.getUint32(2048, true);

    const data = enc.encode('written through wasi');
    mem.set(data, 4100);
    putIovec(view, 3100, 4100, data.byteLength);
    ret = wasi.wasiImport.fd_write(writeFd, 3100, 1, 5100);
    expect(ret).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(wasi.wasiImport.fd_close(writeFd)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(fs.readFileSync(writePath, 'utf8')).toBe('written through wasi');
  });
});
