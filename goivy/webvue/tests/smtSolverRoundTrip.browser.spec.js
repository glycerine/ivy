import { expect, test } from '@playwright/test';
import fs from 'node:fs/promises';
import http from 'node:http';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const webvueDir = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const staticDir = path.join(webvueDir, 'static');
const wasmFile = process.env.SMT_SOLVER_ROUND_TRIP_WASM ?? 'smt-wasip1-roundtrip.wasm';

const contentTypes = new Map([
  ['.html', 'text/html; charset=utf-8'],
  ['.js', 'text/javascript; charset=utf-8'],
  ['.wasm', 'application/wasm'],
]);

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

test('Big Go WASI smt package solves a tiny query through Z3 wasm in a worker', async ({ page }) => {
  await page.goto(baseURL);

  const result = await page.evaluate(async ({ wasmFile }) => {
    const assetBaseURL = window.location.origin;
    const workerSource = `
      const assetBaseURL = ${JSON.stringify(assetBaseURL)};
      const wasmFile = ${JSON.stringify(wasmFile)};

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

        function memoryBytes() {
          const memory = getMemory();
          if (!memory) {
            throw new Error('WASI import used before wasm memory was available');
          }
          return new Uint8Array(memory.buffer);
        }

        function memoryView() {
          const memory = getMemory();
          if (!memory) {
            throw new Error('WASI import used before wasm memory was available');
          }
          return new DataView(memory.buffer);
        }

        function writeU32(ptr, value) {
          memoryView().setUint32(ptr >>> 0, value >>> 0, true);
        }

        function writeU64(ptr, value) {
          memoryView().setBigUint64(ptr >>> 0, BigInt(value), true);
        }

        function zero(ptr, length) {
          memoryBytes().fill(0, ptr >>> 0, (ptr >>> 0) + (length >>> 0));
        }

        function copyRandom(ptr, length) {
          const heap = memoryBytes();
          let offset = ptr >>> 0;
          let remaining = length >>> 0;
          while (remaining > 0) {
            const chunkLength = Math.min(remaining, 65536);
            crypto.getRandomValues(heap.subarray(offset, offset + chunkLength));
            offset += chunkLength;
            remaining -= chunkLength;
          }
        }

        return {
          args_get() { return errnoSuccess; },
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
          environ_get() { return errnoSuccess; },
          environ_sizes_get(countPtr, bufSizePtr) {
            writeU32(countPtr, 0);
            writeU32(bufSizePtr, 0);
            return errnoSuccess;
          },
          fd_close() { return errnoSuccess; },
          fd_fdstat_get(fd, statPtr) {
            void fd;
            zero(statPtr, 24);
            return errnoSuccess;
          },
          fd_fdstat_set_flags() { return errnoSuccess; },
          fd_prestat_dir_name() { return errnoBadf; },
          fd_prestat_get() { return errnoBadf; },
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
            const dataView = memoryView();
            const heap = memoryBytes();
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
          path_open() { return errnoBadf; },
          poll_oneoff(inPtr, outPtr, nsubscriptions, neventsPtr) {
            void inPtr;
            void outPtr;
            void nsubscriptions;
            writeU32(neventsPtr, 0);
            return errnoSuccess;
          },
          proc_exit(code) { throw new WasiExit(code); },
          random_get(bufPtr, bufLen) {
            copyRandom(bufPtr, bufLen);
            return errnoSuccess;
          },
          sched_yield() { return errnoSuccess; },
        };
      }

      self.onmessage = async () => {
        let z3;

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

          let wasmMemory;
          const imports = {
            smt_z3: {
              context_new() {
                const cfg = z3._Z3_mk_config();
                const ctx = z3._Z3_mk_context_rc(cfg);
                z3._Z3_del_config(cfg);
                return ctx >>> 0;
              },
              context_close(ctx) {
                z3._Z3_del_context(ctx);
              },
              bool_val(ctx, val) {
                return (val ? z3._Z3_mk_true(ctx) : z3._Z3_mk_false(ctx)) >>> 0;
              },
              not(ctx, expr) {
                return z3._Z3_mk_not(ctx, expr) >>> 0;
              },
              solver_new(ctx) {
                const solver = z3._Z3_mk_solver(ctx);
                z3._Z3_solver_inc_ref(ctx, solver);
                return solver >>> 0;
              },
              solver_assert(ctx, solver, expr) {
                z3._Z3_solver_assert(ctx, solver, expr);
              },
              solver_check(ctx, solver) {
                return z3._Z3_solver_check(ctx, solver) | 0;
              },
            },
            wasi_snapshot_preview1: createWasiImports(() => wasmMemory),
          };

          const response = await fetch(assetBaseURL + '/' + wasmFile, { cache: 'no-store' });
          if (!response.ok) {
            throw new Error('could not fetch ' + wasmFile + ': ' + response.status);
          }

          const bytes = await response.arrayBuffer();
          const { instance } = await WebAssembly.instantiate(bytes, imports);
          wasmMemory = instance.exports.memory;

          try {
            instance.exports._start();
          } catch (error) {
            if (!(error instanceof WasiExit) || error.code !== 0) {
              throw error;
            }
          }

          const satTrue = instance.exports.smt_solver_sat_true;
          const unsatTrueAndNotTrue = instance.exports.smt_solver_unsat_true_and_not_true;
          if (typeof satTrue !== 'function' || typeof unsatTrueAndNotTrue !== 'function') {
            throw new Error('SMT wasip1 round-trip fixture is missing solver exports');
          }

          self.postMessage({
            ok: true,
            satTrue: satTrue() | 0,
            unsatTrueAndNotTrue: unsatTrueAndNotTrue() | 0,
          });
        } catch (error) {
          self.postMessage({
            ok: false,
            error: error && error.stack ? error.stack : String(error),
          });
        }
      };
    `;

    const workerURL = URL.createObjectURL(new Blob([workerSource], { type: 'text/javascript' }));
    const worker = new Worker(workerURL);

    try {
      return await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error('smt WASI/Z3 worker test timed out'));
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
          resolve(event.data);
        };

        worker.postMessage({ command: 'run' });
      });
    } finally {
      worker.terminate();
      URL.revokeObjectURL(workerURL);
    }
  }, { wasmFile });

  expect(result.satTrue).toBe(1);
  expect(result.unsatTrueAndNotTrue).toBe(-1);
});
