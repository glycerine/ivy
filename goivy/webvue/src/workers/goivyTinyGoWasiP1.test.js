import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { afterEach, describe, expect, test } from 'vitest';

import { createGoIvyTinyGoWasiP1, GOIVY_WASI_ERRNO } from './goivyTinyGoWasiP1.js';

const enc = new TextEncoder();
const dec = new TextDecoder();
const webvueDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');
const goivyRoot = path.resolve(webvueDir, '..');
const RIGHTS_FD_READ = 1n << 1n;
const RIGHTS_FD_WRITE = 1n << 6n;
const RIGHTS_PATH_OPEN = 1n << 13n;

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

function scratchBytes(wasi, mem, text, ptr = 1024) {
  const len = putString(mem, ptr, text);
  const h = wasi.goivyFsImport.__scratch_bytes_begin(len);
  for (let offset = 0; offset < len;) {
    let word = 0;
    let n = Math.min(4, len - offset);
    for (let i = 0; i < n; i += 1) {
      word |= mem[ptr + offset + i] << (8 * i);
    }
    wasi.goivyFsImport.__scratch_bytes_write(h, offset, word >>> 0, n);
    offset += n;
  }
  return { h, len };
}

function readBytesHandle(wasi, h) {
  const len = wasi.goivyFsImport.bytes_len(h);
  const out = new Uint8Array(len);
  for (let offset = 0; offset < len; offset += 4) {
    const word = wasi.goivyFsImport.bytes_word(h, offset);
    for (let i = 0; i < 4 && offset + i < len; i += 1) {
      out[offset + i] = (word >>> (8 * i)) & 0xff;
    }
  }
  wasi.goivyFsImport.bytes_release(h);
  return out;
}

function maybeTinyGo() {
  const tinygo = process.env.TINYGO || '/usr/local/bin/tinygo';
  try {
    const output = execFileSync(tinygo, ['version'], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] });
    if (output.includes('go1.26')) {
      return { tinygo, skip: 'TinyGo 0.40.1 rejects Go 1.26 for fresh wasm builds' };
    }
    return { tinygo, skip: '' };
  } catch (error) {
    return { tinygo, skip: 'tinygo is not available: ' + (error && error.message ? error.message : String(error)) };
  }
}

async function runTinyGoProbe(wasmPath, readPath, writePath, stdout, stderr) {
  const wasmExecPath = path.join(goivyRoot, 'webvue', 'static', 'wasm_exec_tinygo_0.40.1.js');
  const wasmExecSource = fs.readFileSync(wasmExecPath, 'utf8');
  new Function(wasmExecSource + '\n//# sourceURL=' + wasmExecPath)();
  const go = new globalThis.Go();
  go.argv = ['tinyfsprobe', readPath, writePath];
  go.env = {};
  go.exit = (code) => {
    if (code !== 0) {
      stderr(enc.encode('[tinyfsprobe] exit code ' + code + '\n'));
    }
  };

  let wasmMemory;
  const wasi = createGoIvyTinyGoWasiP1({
    args: go.argv,
    env: [],
    stdout,
    stderr,
    procExit: go.importObject.wasi_snapshot_preview1.proc_exit,
    nodeFilesystem: fs,
    nodePath: path,
    hostPreopenPath: '/',
    hostPreopenName: '/',
  });
  go.importObject.wasi_snapshot_preview1 = wasi.wasiImport;
  go.importObject.goivy_fs = wasi.goivyFsImport;
  const result = await WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject);
  const instance = result.instance;
  wasi.setInstance(instance);
  wasmMemory = instance.exports.mem || instance.exports.memory;
  if (!wasmMemory) {
    throw new Error('tinyfsprobe wasm does not export memory');
  }
  await go.run(instance);
}

describe('goivy TinyGo WASI preview1 host', () => {
  test('keeps browser mode limited to the in-memory include preopen', () => {
    const { wasi, mem, view } = newHarness({
      includeRoot: '/include',
      includeTree: { files: [{ path: 'order.ivy', data: '#lang ivy1.8\n' }] },
    });

    const len = putString(mem, 1024, 'order.ivy');
    expect(wasi.wasiImport.path_filestat_get(3, 0, 1024, len, 2048)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(view.getUint8(2048 + 16)).toBe(4);
    expect(Number(view.getBigUint64(2048 + 32, true))).toBe('#lang ivy1.8\n'.length);

    expect(wasi.wasiImport.fd_prestat_get(3, 3000)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const preopenNameLen = view.getUint32(3004, true);
    expect(preopenNameLen).toBe('/include'.length);
    expect(wasi.wasiImport.fd_prestat_dir_name(3, 3100, 64)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(dec.decode(mem.subarray(3100, 3100 + preopenNameLen))).toBe('/include');
    expect(wasi.wasiImport.fd_prestat_dir_name(3, 3200, preopenNameLen - 1)).toBe(GOIVY_WASI_ERRNO.NAMETOOLONG);

    expect(wasi.wasiImport.fd_fdstat_get(3, 3300)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const baseRights = view.getBigUint64(3300 + 8, true);
    const inheritingRights = view.getBigUint64(3300 + 16, true);
    expect(baseRights & RIGHTS_PATH_OPEN).toBe(RIGHTS_PATH_OPEN);
    expect(baseRights & RIGHTS_FD_READ).toBe(0n);
    expect(inheritingRights & RIGHTS_FD_READ).toBe(RIGHTS_FD_READ);
    expect(inheritingRights & RIGHTS_FD_WRITE).toBe(0n);

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

    expect(wasi.wasiImport.fd_fdstat_get(4, 7000)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const hostBaseRights = view.getBigUint64(7000 + 8, true);
    const hostInheritingRights = view.getBigUint64(7000 + 16, true);
    expect(hostBaseRights & RIGHTS_PATH_OPEN).toBe(RIGHTS_PATH_OPEN);
    expect(hostBaseRights & RIGHTS_FD_READ).toBe(0n);
    expect(hostInheritingRights & RIGHTS_FD_READ).toBe(RIGHTS_FD_READ);
    expect(hostInheritingRights & RIGHTS_FD_WRITE).toBe(RIGHTS_FD_WRITE);

    let rel = readPath.slice(1);
    let len = putString(mem, 1024, rel);
    let ret = wasi.wasiImport.path_open(4, 0, 1024, len, 0, hostInheritingRights & RIGHTS_FD_READ, hostInheritingRights, 0, 2048);
    expect(ret).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const readFd = view.getUint32(2048, true);
    expect(wasi.wasiImport.fd_filestat_get(readFd, 6000)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(view.getUint8(6000 + 16)).toBe(4);
    expect(Number(view.getBigUint64(6000 + 32, true))).toBe('hello tinygo wasi fs'.length);

    putIovec(view, 3000, 4000, 64);
    ret = wasi.wasiImport.fd_read(readFd, 3000, 1, 5000);
    expect(ret).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const nread = view.getUint32(5000, true);
    expect(dec.decode(mem.subarray(4000, 4000 + nread))).toBe('hello tinygo wasi fs');
    expect(wasi.wasiImport.fd_close(readFd)).toBe(GOIVY_WASI_ERRNO.SUCCESS);

    rel = writePath.slice(1);
    len = putString(mem, 1100, rel);
    expect(wasi.wasiImport.path_filestat_get(4, 0, 1100, len, 6000)).toBe(GOIVY_WASI_ERRNO.NOENT);
    const oflagsCreateTrunc = 1 | 8;
    ret = wasi.wasiImport.path_open(4, 0, 1100, len, oflagsCreateTrunc, hostInheritingRights & RIGHTS_FD_WRITE, hostInheritingRights, 0, 2048);
    expect(ret).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    const writeFd = view.getUint32(2048, true);

    const data = enc.encode('written through wasi');
    mem.set(data, 4100);
    putIovec(view, 3100, 4100, data.byteLength);
    ret = wasi.wasiImport.fd_write(writeFd, 3100, 1, 5100);
    expect(ret).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(wasi.wasiImport.fd_close(writeFd)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(fs.readFileSync(writePath, 'utf8')).toBe('written through wasi');
    expect(wasi.wasiImport.path_filestat_get(4, 0, 1100, len, 6000)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(Number(view.getBigUint64(6000 + 32, true))).toBe('written through wasi'.length);
  });

  test('goivy_fs host import can stat, read, and write host files directly', () => {
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'goivy-tinygo-hostfs-'));
    tmpDirs.push(dir);
    const readPath = path.join(dir, 'input.ivy');
    const writePath = path.join(dir, 'output.ivy');
    fs.writeFileSync(readPath, 'host fs payload');

    const { wasi, mem } = newHarness({
      includeRoot: '/goivy-virtual-include',
      includeTree: { files: [{ path: '1.8/order.ivy', data: '#lang ivy1.8\ninclude payload' }] },
      nodeFilesystem: fs,
      nodePath: path,
      hostPreopenPath: '/',
      hostPreopenName: '/',
    });

    let pathScratch = scratchBytes(wasi, mem, readPath);
    expect(wasi.goivyFsImport.stat(pathScratch.h, pathScratch.len)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(wasi.goivyFsImport.last_is_dir()).toBe(0);
    expect(wasi.goivyFsImport.last_size_lo()).toBe('host fs payload'.length);
    const readHandleID = wasi.goivyFsImport.read_file(pathScratch.h, pathScratch.len);
    expect(readHandleID).not.toBe(0);
    expect(dec.decode(readHandleID === 0 ? new Uint8Array(0) : readBytesHandle(wasi, readHandleID))).toBe('host fs payload');

    const outPathScratch = scratchBytes(wasi, mem, writePath);
    const payloadScratch = scratchBytes(wasi, mem, 'host fs write payload');
    expect(wasi.goivyFsImport.write_file(outPathScratch.h, outPathScratch.len, payloadScratch.h, payloadScratch.len)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(fs.readFileSync(writePath, 'utf8')).toBe('host fs write payload');
    expect(wasi.goivyFsImport.stat(outPathScratch.h, outPathScratch.len)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(wasi.goivyFsImport.last_size_lo()).toBe('host fs write payload'.length);

    const dirScratch = scratchBytes(wasi, mem, dir);
    const dirHandleID = wasi.goivyFsImport.read_dir(dirScratch.h, dirScratch.len);
    expect(dirHandleID).not.toBe(0);
    expect(dec.decode(readBytesHandle(wasi, dirHandleID))).toContain('f\tinput.ivy\n');

    const includeRootScratch = scratchBytes(wasi, mem, '/goivy-virtual-include');
    const includeRootHandleID = wasi.goivyFsImport.read_dir(includeRootScratch.h, includeRootScratch.len);
    expect(includeRootHandleID).not.toBe(0);
    expect(dec.decode(readBytesHandle(wasi, includeRootHandleID))).toContain('d\t1.8\n');

    const includeScratch = scratchBytes(wasi, mem, '/goivy-virtual-include/1.8/order.ivy');
    expect(wasi.goivyFsImport.stat(includeScratch.h, includeScratch.len)).toBe(GOIVY_WASI_ERRNO.SUCCESS);
    expect(wasi.goivyFsImport.last_size_lo()).toBe('#lang ivy1.8\ninclude payload'.length);
    const includeHandleID = wasi.goivyFsImport.read_file(includeScratch.h, includeScratch.len);
    expect(includeHandleID).not.toBe(0);
    expect(dec.decode(readBytesHandle(wasi, includeHandleID))).toBe('#lang ivy1.8\ninclude payload');
  });

  test('TinyGo wasm probe can stat, read, and write through goivy_fs', async () => {
    const { tinygo, skip } = maybeTinyGo();
    if (skip) {
      console.warn('skipping TinyGo wasm probe: ' + skip);
      return;
    }

    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'goivy-tinygo-probe-'));
    tmpDirs.push(dir);
    const readPath = path.join(dir, 'input.ivy');
    const writePath = path.join(dir, 'output.ivy');
    const wasmPath = path.join(dir, 'tinyfsprobe.wasm');
    fs.writeFileSync(readPath, 'probe payload');

    execFileSync(tinygo, [
      'build',
      '-panic=trap',
      '-gc=precise',
      '-no-debug',
      '-ldflags',
      `-X main.readPath=${readPath} -X main.writePath=${writePath} -X main.dirPath=${dir}`,
      '-o',
      wasmPath,
      './cmd/tinyfsprobe',
    ], {
      cwd: goivyRoot,
      env: {
        ...process.env,
        HOME: os.tmpdir(),
        GOCACHE: path.join(os.tmpdir(), 'goivy-tinygo-gocache'),
        GOOS: 'js',
        GOARCH: 'wasm',
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    const stdoutChunks = [];
    const stderrChunks = [];
    await runTinyGoProbe(
      wasmPath,
      readPath,
      writePath,
      (data) => stdoutChunks.push(new Uint8Array(data)),
      (data) => stderrChunks.push(new Uint8Array(data)),
    );

    const stdoutText = stdoutChunks.map((chunk) => dec.decode(chunk)).join('');
    const stderrText = stderrChunks.map((chunk) => dec.decode(chunk)).join('');
    expect(stderrText).toBe('');
    expect(stdoutText).toContain('OK read=13 wrote=25');
    expect(fs.readFileSync(writePath, 'utf8')).toBe('tinyfsprobe:probe payload');
  }, 120000);
});
