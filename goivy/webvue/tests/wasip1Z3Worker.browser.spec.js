import { expect, test } from '@playwright/test';
import fs from 'node:fs/promises';
import http from 'node:http';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const webvueDir = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const staticDir = path.join(webvueDir, 'static');

const contentTypes = new Map([
  ['.html', 'text/html; charset=utf-8'],
  ['.js', 'text/javascript; charset=utf-8'],
  ['.json', 'application/json; charset=utf-8'],
  ['.wasm', 'application/wasm'],
]);

const roundTripCount = 100;

let server;
let baseURL;

test.beforeAll(async () => {
  server = http.createServer(async (request, response) => {
    try {
      const url = new URL(request.url ?? '/', 'http://127.0.0.1');
      const pathname = url.pathname === '/' ? '/index.html' : url.pathname;
      const filePath = path.normalize(path.join(staticDir, pathname));

      if (!filePath.startsWith(staticDir + path.sep)) {
        response.writeHead(403);
        response.end('forbidden');
        return;
      }

      const data = await fs.readFile(filePath);
      response.writeHead(200, {
        'cache-control': 'no-store',
        'content-type': contentTypes.get(path.extname(filePath)) ?? 'application/octet-stream',
      });
      response.end(data);
    } catch {
      response.writeHead(404);
      response.end('not found');
    }
  });

  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  baseURL = `http://127.0.0.1:${server.address().port}`;
});

test.afterAll(async () => {
  if (!server) {
    return;
  }

  await new Promise((resolve, reject) => {
    server.close((error) => (error ? reject(error) : resolve()));
  });
});

test('TinyGo and Big Go WASI Ivy wasm call Z3 wasm through JavaScript in a worker', async ({ page }) => {
  await page.goto(baseURL);

  const results = await page.evaluate(async ({ roundTripCount }) => {
    const assetBaseURL = window.location.origin;
    const workerSource = `
      const assetBaseURL = ${JSON.stringify(assetBaseURL)};

      class WasiExit extends Error {
        constructor(code) {
          super('WASI exit ' + code);
          this.name = 'WasiExit';
          this.code = code >>> 0;
        }
      }

      function createWasiImports(getMemory) {
        const errnoSuccess = 0;
        const errnoBadf = 8;
        const textDecoder = new TextDecoder();

        function bytes() {
          const memory = getMemory();
          if (!memory) {
            throw new Error('WASI import used before wasm memory was available');
          }
          return new Uint8Array(memory.buffer);
        }

        function view() {
          const memory = getMemory();
          if (!memory) {
            throw new Error('WASI import used before wasm memory was available');
          }
          return new DataView(memory.buffer);
        }

        function writeU32(ptr, value) {
          view().setUint32(ptr >>> 0, value >>> 0, true);
        }

        function writeU64(ptr, value) {
          view().setBigUint64(ptr >>> 0, BigInt(value), true);
        }

        function zero(ptr, length) {
          bytes().fill(0, ptr >>> 0, (ptr >>> 0) + (length >>> 0));
        }

        function copyRandom(ptr, length) {
          const heap = bytes();
          let offset = ptr >>> 0;
          let remaining = length >>> 0;
          while (remaining > 0) {
            const chunkLength = Math.min(remaining, 65536);
            const chunk = heap.subarray(offset, offset + chunkLength);
            crypto.getRandomValues(chunk);
            offset += chunkLength;
            remaining -= chunkLength;
          }
        }

        return {
          args_get() {
            return errnoSuccess;
          },
          args_sizes_get(argcPtr, argvBufSizePtr) {
            writeU32(argcPtr, 0);
            writeU32(argvBufSizePtr, 0);
            return errnoSuccess;
          },
          clock_time_get(clockId, precision, timePtr) {
            void clockId;
            void precision;
            writeU64(timePtr, BigInt(Date.now()) * 1000000n);
            return errnoSuccess;
          },
          environ_get() {
            return errnoSuccess;
          },
          environ_sizes_get(countPtr, bufSizePtr) {
            writeU32(countPtr, 0);
            writeU32(bufSizePtr, 0);
            return errnoSuccess;
          },
          fd_close() {
            return errnoSuccess;
          },
          fd_fdstat_get(fd, statPtr) {
            void fd;
            zero(statPtr, 24);
            return errnoSuccess;
          },
          fd_fdstat_set_flags() {
            return errnoSuccess;
          },
          fd_prestat_dir_name() {
            return errnoBadf;
          },
          fd_prestat_get() {
            return errnoBadf;
          },
          fd_read(fd, iovsPtr, iovsLen, nreadPtr) {
            void fd;
            void iovsPtr;
            void iovsLen;
            writeU32(nreadPtr, 0);
            return errnoSuccess;
          },
          fd_seek(fd, offset, whence, newOffsetPtr) {
            void fd;
            void offset;
            void whence;
            writeU64(newOffsetPtr, 0n);
            return errnoSuccess;
          },
          fd_write(fd, iovsPtr, iovsLen, nwrittenPtr) {
            const dataView = view();
            const heap = bytes();
            let written = 0;
            let output = '';

            for (let i = 0; i < iovsLen; i += 1) {
              const iov = (iovsPtr >>> 0) + i * 8;
              const ptr = dataView.getUint32(iov, true);
              const len = dataView.getUint32(iov + 4, true);
              written += len;
              if (fd === 1 || fd === 2) {
                output += textDecoder.decode(heap.subarray(ptr, ptr + len));
              }
            }

            if (output.length) {
              console[fd === 2 ? 'error' : 'log'](output.replace(/\\n$/, ''));
            }
            writeU32(nwrittenPtr, written);
            return errnoSuccess;
          },
          path_open() {
            return errnoBadf;
          },
          poll_oneoff(inPtr, outPtr, nsubscriptions, neventsPtr) {
            void inPtr;
            void outPtr;
            void nsubscriptions;
            writeU32(neventsPtr, 0);
            return errnoSuccess;
          },
          proc_exit(code) {
            throw new WasiExit(code);
          },
          random_get(bufPtr, bufLen) {
            copyRandom(bufPtr, bufLen);
            return errnoSuccess;
          },
          sched_yield() {
            return errnoSuccess;
          },
        };
      }

      async function instantiateProbe(z3, ctx, probe) {
        let wasmMemory;
        let boolSortCalls = 0;
        let lastSortId = 0;
        const imports = {
          goivy_z3: {
            bool_sort() {
              boolSortCalls += 1;
              const sort = z3._Z3_mk_bool_sort(ctx);
              lastSortId = z3._Z3_get_sort_id(ctx, sort) >>> 0;
              return lastSortId;
            },
          },
          wasi_snapshot_preview1: createWasiImports(() => wasmMemory),
        };

        const response = await fetch(assetBaseURL + '/' + probe.file, { cache: 'no-store' });
        if (!response.ok) {
          throw new Error('could not fetch ' + probe.file + ': ' + response.status);
        }

        const bytes = await response.arrayBuffer();
        const { instance } = await WebAssembly.instantiate(bytes, imports);
        wasmMemory = instance.exports.memory;

        const hasInitialize = typeof instance.exports._initialize === 'function';
        const hasStart = typeof instance.exports._start === 'function';
        let startExitCode = null;
        let startReturned = false;
        if (probe.compiler === 'tinygo' && hasInitialize) {
          instance.exports._initialize();
        } else if (probe.compiler === 'biggo') {
          try {
            instance.exports._start();
            startReturned = true;
          } catch (error) {
            if (error instanceof WasiExit) {
              startExitCode = error.code;
            } else {
              throw error;
            }
          }
        }

        const initializedByStart = boolSortCalls > 0;
        const checksum = initializedByStart
          ? instance.exports.ivy_probe_checksum() >>> 0
          : instance.exports.ivy_probe_init() >>> 0;
        const sortId = instance.exports.ivy_probe_z3_bool_sort_id() >>> 0;
        const cachedChecksum = instance.exports.ivy_probe_checksum() >>> 0;
        const roundTrip = instance.exports.ivy_probe_round_trip_bool_sort_id;
        if (typeof roundTrip !== 'function') {
          throw new Error(probe.file + ' is missing ivy_probe_round_trip_bool_sort_id; run make z3-wasip1-browser-artifacts');
        }

        const roundTripStartCalls = boolSortCalls;
        let roundTripLastSortId = 0;
        const roundTripStartedAt = performance.now();
        for (let i = 0; i < ${roundTripCount}; i += 1) {
          roundTripLastSortId = roundTrip() >>> 0;
        }
        const roundTripElapsedMs = performance.now() - roundTripStartedAt;
        const roundTripCalls = boolSortCalls - roundTripStartCalls;

        return {
          boolSortCalls,
          cachedChecksum,
          checksum,
          compiler: probe.compiler,
          hasInitialize,
          hasStart,
          initializedByStart,
          lastSortId,
          roundTripCalls,
          roundTripCount: ${roundTripCount},
          roundTripElapsedMs,
          roundTripLastSortId,
          roundTripMeanUs: (roundTripElapsedMs * 1000) / ${roundTripCount},
          sortId,
          startExitCode,
          startReturned,
        };
      }

      self.onmessage = async () => {
        let z3;
        let ctx = 0;

        try {
          importScripts(assetBaseURL + '/z3-471-api.js');
          if (typeof initZ3 !== 'function') {
            throw new Error('z3-471-api.js did not expose initZ3 in the worker');
          }

          z3 = await initZ3({
            locateFile(file) {
              if (file === 'z3-api.wasm') {
                return assetBaseURL + '/z3-471-api.wasm';
              }
              return assetBaseURL + '/' + file;
            },
          });

          const cfg = z3._Z3_mk_config();
          ctx = z3._Z3_mk_context_rc(cfg);
          z3._Z3_del_config(cfg);

          const probes = [
            { compiler: 'tinygo', file: 'ivy-tinygo-wasip1-probe.wasm' },
            { compiler: 'biggo', file: 'ivy-biggo-wasip1-probe.wasm' },
          ];

          const results = [];
          for (const probe of probes) {
            results.push(await instantiateProbe(z3, ctx, probe));
          }

          self.postMessage({ ok: true, results });
        } catch (error) {
          self.postMessage({
            ok: false,
            error: error && error.stack ? error.stack : String(error),
          });
        } finally {
          if (z3 && ctx) {
            z3._Z3_del_context(ctx);
          }
        }
      };
    `;

    const workerURL = URL.createObjectURL(new Blob([workerSource], { type: 'text/javascript' }));
    const worker = new Worker(workerURL);

    try {
      return await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error('WASI Ivy/Z3 worker probe timed out'));
        }, 30000);

        worker.onerror = (error) => {
          clearTimeout(timeout);
          reject(new Error(error.message || 'worker error'));
        };

        worker.onmessage = (event) => {
          clearTimeout(timeout);
          if (!event.data.ok) {
            reject(new Error(event.data.error));
            return;
          }
          resolve(event.data.results);
        };

        worker.postMessage({ command: 'run' });
      });
    } finally {
      worker.terminate();
      URL.revokeObjectURL(workerURL);
    }
  }, { roundTripCount });

  expect(results.map((result) => result.compiler)).toEqual(['tinygo', 'biggo']);

  for (const result of results) {
    expect(result.boolSortCalls).toBeGreaterThan(0);
    expect(result.lastSortId).toBeGreaterThan(0);
    expect(result.sortId).toBe(result.lastSortId);
    expect(result.checksum).toBe(result.cachedChecksum);
    expect(result.checksum).not.toBe(result.sortId);
    expect(result.hasInitialize || result.hasStart).toBe(true);
    expect(result.roundTripCalls).toBe(roundTripCount);
    expect(result.roundTripLastSortId).toBe(result.lastSortId);
    expect(result.roundTripElapsedMs).toBeGreaterThan(0);
    if (result.compiler === 'biggo') {
      expect([null, 0]).toContain(result.startExitCode);
    }
  }

  for (const result of results) {
    console.log(
      `${result.compiler}: ${result.roundTripCount} round trips in `
      + `${result.roundTripElapsedMs.toFixed(3)} ms `
      + `(${result.roundTripMeanUs.toFixed(2)} us/round trip)`,
    );
  }
});
