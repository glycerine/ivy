import { describe, expect, it } from 'vitest';
import { createGoIvyTinyGoWasiP1, GOIVY_WASI_ERRNO } from './goivyTinyGoWasiP1.js';

const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder('utf-8', { fatal: false });

function makeHarness() {
  const memory = new WebAssembly.Memory({ initial: 1 });
  const stdout = [];
  const stderr = [];
  const includeTree = {
    root: '/virtual/ivy/include',
    files: [
      { path: '1.8/order.ivy', data: 'module order = {}\n' },
      { path: 'a.txt', data: 'alpha' },
    ],
  };
  const wasi = createGoIvyTinyGoWasiP1({
    args: ['goivy_check_jswasm'],
    env: ['GOIVY_INCLUDE=' + includeTree.root],
    includeRoot: includeTree.root,
    includeTree,
    stdout(data) { stdout.push(new Uint8Array(data)); },
    stderr(data) { stderr.push(new Uint8Array(data)); },
  });
  wasi.setInstance({ exports: { memory } });
  return { memory, wasi: wasi.wasiImport, stdout, stderr, includeTree };
}

function view(h) {
  return new DataView(h.memory.buffer);
}

function bytes(h) {
  return new Uint8Array(h.memory.buffer);
}

function readU32(h, ptr) {
  return view(h).getUint32(ptr, true);
}

function writeString(h, ptr, value) {
  const data = textEncoder.encode(value);
  bytes(h).set(data, ptr);
  return data.byteLength;
}

function readString(h, ptr, len) {
  return textDecoder.decode(bytes(h).subarray(ptr, ptr + len));
}

function writeIovecs(h, ptr, iovecs) {
  const v = view(h);
  for (let i = 0; i < iovecs.length; i += 1) {
    v.setUint32(ptr + i * 8, iovecs[i].ptr, true);
    v.setUint32(ptr + i * 8 + 4, iovecs[i].len, true);
  }
}

describe('goivy TinyGo WASI preview1 shim', () => {
  it('exports only the WASI functions imported by the TinyGo goivy artifact', () => {
    const h = makeHarness();
    expect(Object.keys(h.wasi).sort()).toEqual([
      'fd_close',
      'fd_fdstat_get',
      'fd_prestat_dir_name',
      'fd_prestat_get',
      'fd_read',
      'fd_seek',
      'fd_write',
      'path_open',
      'proc_exit',
      'random_get',
    ]);
  });

  it('lets wasi-libc discover exactly one preopen before fd 4 returns BADF', () => {
    const h = makeHarness();
    const rootLen = textEncoder.encode(h.includeTree.root).byteLength;

    expect(h.wasi.fd_prestat_get(3, 0)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU32(h, 4)).toBe(rootLen);
    expect(h.wasi.fd_prestat_dir_name(3, 16, rootLen)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readString(h, 16, rootLen)).toBe(h.includeTree.root);
    expect(h.wasi.fd_prestat_get(4, 64)).toBe(GOIVY_WASI_ERRNO.BADF);
  });

  it('opens and reads include files through wasi-libc path_open/fd_read', () => {
    const h = makeHarness();
    const pathPtr = 64;
    const openedFdPtr = 128;
    const pathLen = writeString(h, pathPtr, 'a.txt');

    expect(h.wasi.path_open(3, 0, pathPtr, pathLen, 0, 0n, 0n, 0, openedFdPtr)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const fd = readU32(h, openedFdPtr);
    expect(fd).toBe(4);

    writeIovecs(h, 160, [{ ptr: 192, len: 8 }]);
    expect(h.wasi.fd_read(fd, 160, 1, 184)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU32(h, 184)).toBe(5);
    expect(readString(h, 192, 5)).toBe('alpha');
  });
});
