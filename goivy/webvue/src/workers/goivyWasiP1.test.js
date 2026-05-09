import { describe, expect, it } from 'vitest';
import { createGoIvyWasiP1, GOIVY_WASI_ERRNO } from './goivyWasiP1.js';

const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder('utf-8', { fatal: false });

const FILETYPE_DIRECTORY = 3;
const FILETYPE_REGULAR_FILE = 4;
const OFLAGS_DIRECTORY = 1 << 1;

function makeHarness(options = {}) {
  const memory = new WebAssembly.Memory({ initial: 1 });
  const stdout = [];
  const stderr = [];
  const heapProfiles = [];
  const debug = [];
  const includeTree = options.includeTree || {
    root: '/virtual/ivy/include',
    files: [
      { path: '1.8/order.ivy', data: 'module order = {}\n' },
      { path: 'a.txt', data: 'alpha' },
      { path: 'long-name-for-truncation.txt', data: 'long' },
      { path: 'sub/child.ivy', data: 'child' },
    ],
  };
  const wasi = createGoIvyWasiP1({
    args: options.args || ['goivy_check_wasip1', 'input.ivy'],
    env: options.env || ['GOIVY_INCLUDE=' + includeTree.root],
    includeRoot: options.includeRoot || includeTree.root,
    includeTree,
    stdout(data) { stdout.push(new Uint8Array(data)); },
    stderr(data) { stderr.push(new Uint8Array(data)); },
    heapProfile(data) { heapProfiles.push(new Uint8Array(data)); },
    debug(text) { debug.push(String(text)); },
  });
  wasi.initialize({ exports: { memory } });
  return { memory, wasi: wasi.wasiImport, stdout, stderr, heapProfiles, debug };
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

function readU64(h, ptr) {
  return view(h).getBigUint64(ptr, true);
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

function parseDirents(h, ptr, used) {
  const out = [];
  const v = view(h);
  let cursor = ptr;
  const end = ptr + used;
  while (cursor + 24 <= end) {
    const next = v.getBigUint64(cursor, true);
    const ino = v.getBigUint64(cursor + 8, true);
    const namelen = v.getUint32(cursor + 16, true);
    const filetype = v.getUint8(cursor + 20);
    const nameStart = cursor + 24;
    const nameEnd = nameStart + namelen;
    out.push({
      next,
      ino,
      namelen,
      filetype,
      name: nameEnd <= end ? readString(h, nameStart, namelen) : null,
    });
    if (nameEnd > end) {
      break;
    }
    cursor = nameEnd;
  }
  return out;
}

function mulberry32(seed) {
  let x = seed >>> 0;
  return function random() {
    x += 0x6D2B79F5;
    let t = x;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function pick(rng, values) {
  return values[Math.floor(rng() * values.length)];
}

describe('goivy WASI preview1 shim', () => {
  it('reports args and env, and returns EFAULT for out-of-bounds memory like wazero', () => {
    const h = makeHarness({
      args: ['goivy_check_wasip1', 'demo.ivy'],
      env: ['GOIVY_INCLUDE=/virtual/ivy/include', 'X=1'],
    });

    expect(h.wasi.args_sizes_get(0, 4)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU32(h, 0)).toBe(2);
    expect(readU32(h, 4)).toBe('goivy_check_wasip1'.length + 1 + 'demo.ivy'.length + 1);

    expect(h.wasi.args_get(8, 32)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const arg0 = readU32(h, 8);
    const arg1 = readU32(h, 12);
    expect(readString(h, arg0, 'goivy_check_wasip1'.length)).toBe('goivy_check_wasip1');
    expect(readString(h, arg1, 'demo.ivy'.length)).toBe('demo.ivy');

    expect(h.wasi.environ_sizes_get(128, 132)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU32(h, 128)).toBe(2);
    expect(h.wasi.args_sizes_get(65535, 4)).toBe(GOIVY_WASI_ERRNO.FAULT);
    expect(h.debug.some((line) => line.includes('args_sizes_get'))).toBe(true);
  });

  it('exposes the include directory preopen and file stats with wazero-compatible errors', () => {
    const h = makeHarness();
    const root = '/virtual/ivy/include';
    const rootLen = textEncoder.encode(root).byteLength;

    expect(h.wasi.fd_prestat_get(3, 0)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU32(h, 4)).toBe(rootLen);
    expect(h.wasi.fd_prestat_dir_name(3, 16, rootLen)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readString(h, 16, rootLen)).toBe(root);
    expect(h.wasi.fd_prestat_dir_name(3, 16, rootLen + 1)).toBe(GOIVY_WASI_ERRNO.NAMETOOLONG);

    const pathPtr = 128;
    const pathLen = writeString(h, pathPtr, '1.8/order.ivy');
    expect(h.wasi.path_filestat_get(3, 0, pathPtr, pathLen, 256)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(view(h).getUint8(256 + 16)).toBe(FILETYPE_REGULAR_FILE);
    expect(readU64(h, 256 + 32)).toBe(BigInt('module order = {}\n'.length));

    const badPathLen = writeString(h, pathPtr, '/1.8/order.ivy');
    expect(h.wasi.path_filestat_get(3, 0, pathPtr, badPathLen, 256)).toBe(GOIVY_WASI_ERRNO.PERM);
  });

  it('opens, reads, seeks, and rejects invalid paths the same way wazero does for our read-only tree', () => {
    const h = makeHarness();
    const pathPtr = 64;
    const openedFdPtr = 128;

    let pathLen = writeString(h, pathPtr, 'a.txt');
    expect(h.wasi.path_open(3, 0, pathPtr, pathLen, 0, 0n, 0n, 0, openedFdPtr)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const fd = readU32(h, openedFdPtr);
    expect(fd).toBeGreaterThanOrEqual(5);

    writeIovecs(h, 160, [{ ptr: 192, len: 8 }]);
    expect(h.wasi.fd_read(fd, 160, 1, 184)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU32(h, 184)).toBe(5);
    expect(readString(h, 192, 5)).toBe('alpha');
    expect(h.wasi.fd_seek(fd, 0n, 0, 200)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU64(h, 200)).toBe(0n);

    expect(h.wasi.fd_seek(3, 0n, 0, 200)).toBe(GOIVY_WASI_ERRNO.ISDIR);
    expect(h.wasi.path_open(3, 0, pathPtr, 0, 0, 0n, 0n, 0, openedFdPtr)).toBe(GOIVY_WASI_ERRNO.INVAL);

    pathLen = writeString(h, pathPtr, '/a.txt');
    expect(h.wasi.path_open(3, 0, pathPtr, pathLen, 0, 0n, 0n, 0, openedFdPtr)).toBe(GOIVY_WASI_ERRNO.PERM);
    pathLen = writeString(h, pathPtr, '../a.txt');
    expect(h.wasi.path_open(3, 0, pathPtr, pathLen, 0, 0n, 0n, 0, openedFdPtr)).toBe(GOIVY_WASI_ERRNO.PERM);

    pathLen = writeString(h, pathPtr, 'sub');
    expect(h.wasi.path_open(3, 0, pathPtr, pathLen, 0, 0n, 0n, 0, openedFdPtr)).toBe(GOIVY_WASI_ERRNO.ISDIR);
    expect(h.wasi.path_open(3, 0, pathPtr, pathLen, OFLAGS_DIRECTORY, 0n, 0n, 0, openedFdPtr)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU32(h, openedFdPtr)).toBeGreaterThanOrEqual(5);
  });

  it('implements fd_readdir cookies and truncation in the wazero style', () => {
    const h = makeHarness();

    expect(h.wasi.fd_readdir(3, 0, 23, 0n, 512)).toBe(GOIVY_WASI_ERRNO.INVAL);
    expect(h.wasi.fd_readdir(3, 0, 24, 0n, 512)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(readU32(h, 512)).toBe(24);
    const truncated = parseDirents(h, 0, 24);
    expect(truncated).toHaveLength(1);
    expect(truncated[0]).toMatchObject({ next: 1n, namelen: 1, filetype: FILETYPE_DIRECTORY, name: null });

    bytes(h).fill(0, 0, 512);
    expect(h.wasi.fd_readdir(3, 0, 192, 0n, 512)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const used = readU32(h, 512);
    const entries = parseDirents(h, 0, used);
    expect(entries.map((entry) => entry.name)).toEqual(['.', '..', '1.8', 'a.txt', 'long-name-for-truncation.txt', 'sub']);
    expect(entries[2]).toMatchObject({ next: 3n, filetype: FILETYPE_DIRECTORY });
    expect(entries[3]).toMatchObject({ filetype: FILETYPE_REGULAR_FILE });

    expect(h.wasi.fd_readdir(3, 0, 192, 99n, 512)).toBe(GOIVY_WASI_ERRNO.NOENT);
  });

  it('fuzzes path and directory calls so malformed guest pointers return errno instead of host throws', () => {
    const h = makeHarness();
    const rng = mulberry32(0x1EAFCAFE);
    const paths = [
      '',
      '.',
      'a.txt',
      'sub',
      'sub/child.ivy',
      'missing.ivy',
      '../escape',
      '/absolute',
      'sub/../a.txt',
      'sub/child.ivy/',
      'nul\0byte',
      'very/'.repeat(32) + 'deep.ivy',
    ];
    const memSize = bytes(h).byteLength;

    for (let i = 0; i < 1000; i += 1) {
      const path = pick(rng, paths);
      const encoded = textEncoder.encode(path);
      const pathPtr = Math.floor(rng() * (memSize + 64));
      const pathLen = rng() < 0.2 ? Math.floor(rng() * (memSize + 64)) : encoded.byteLength;
      const resultPtr = Math.floor(rng() * (memSize + 64));
      const bufPtr = Math.floor(rng() * (memSize + 64));
      const bufLen = Math.floor(rng() * 160);
      const cookie = BigInt(Math.floor(rng() * 16));
      const oflags = rng() < 0.25 ? OFLAGS_DIRECTORY : 0;

      if (pathPtr + encoded.byteLength <= memSize) {
        bytes(h).set(encoded, pathPtr);
      }

      expect(() => h.wasi.path_filestat_get(3, 0, pathPtr, pathLen, resultPtr)).not.toThrow();
      expect(() => h.wasi.path_open(3, 0, pathPtr, pathLen, oflags, 0n, 0n, 0, resultPtr)).not.toThrow();
      expect(() => h.wasi.fd_readdir(3, bufPtr, bufLen, cookie, resultPtr)).not.toThrow();

      expect(typeof h.wasi.path_filestat_get(3, 0, pathPtr, pathLen, resultPtr)).toBe('number');
      expect(typeof h.wasi.path_open(3, 0, pathPtr, pathLen, oflags, 0n, 0n, 0, resultPtr)).toBe('number');
      expect(typeof h.wasi.fd_readdir(3, bufPtr, bufLen, cookie, resultPtr)).toBe('number');
    }
  });
});
