import { expect, test } from '@playwright/test';
import fs from 'node:fs/promises';
import http from 'node:http';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const webvueDir = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const nodeModulesDir = path.join(webvueDir, 'node_modules');
const srcDir = path.join(webvueDir, 'src');
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
      let filePath;

      if (url.pathname === '/') {
        filePath = path.join(staticDir, 'index.html');
      } else if (url.pathname.startsWith('/node_modules/')) {
        filePath = path.normalize(path.join(webvueDir, url.pathname));

        if (!filePath.startsWith(nodeModulesDir + path.sep)) {
          response.writeHead(403);
          response.end('forbidden');
          return;
        }
      } else if (url.pathname.startsWith('/src/')) {
        filePath = path.normalize(path.join(webvueDir, url.pathname));

        if (!filePath.startsWith(srcDir + path.sep)) {
          response.writeHead(403);
          response.end('forbidden');
          return;
        }
      } else {
        filePath = path.normalize(path.join(staticDir, url.pathname));

        if (!filePath.startsWith(staticDir + path.sep)) {
          response.writeHead(403);
          response.end('forbidden');
          return;
        }
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

test('Big Go WASI Ivy wasm calls Z3 wasm through JavaScript in a worker', async ({ page }) => {
  await page.goto(baseURL);

  const results = await page.evaluate(async ({ roundTripCount }) => {
    const assetBaseURL = window.location.origin;
    const workerSource = `
      const assetBaseURL = ${JSON.stringify(assetBaseURL)};

      async function instantiateProbe(z3, ctx, probe) {
        let wasmMemory;
        let boolSortCalls = 0;
        let lastSortId = 0;
        const { createSmtZ3Imports } = await import(assetBaseURL + '/src/workers/smtZ3Imports.js');
        const { createGoIvyWasiP1 } = await import(assetBaseURL + '/src/workers/goivyWasiP1.js');

        function recordBoolSort(z3ctx, sort) {
          boolSortCalls += 1;
          lastSortId = z3._Z3_get_sort_id(z3ctx, sort) >>> 0;
          return sort >>> 0;
        }

        const smtZ3Imports = createSmtZ3Imports({ z3, getGoMemory: () => wasmMemory });
        const smtMkBoolSort = smtZ3Imports.Z3_mk_bool_sort;
        smtZ3Imports.Z3_mk_bool_sort = (z3ctx) => recordBoolSort(z3ctx, smtMkBoolSort(z3ctx));
        const wasi = createGoIvyWasiP1({
          args: [probe.file],
          includeRoot: '.',
          includeTree: { root: '.', files: [] },
          stdout() {},
          stderr() {},
        });

        const imports = {
          smt_z3: smtZ3Imports,
          goivy_z3: {
            bool_sort() {
              recordBoolSort(ctx, z3._Z3_mk_bool_sort(ctx));
              return lastSortId;
            },
          },
          wasi_snapshot_preview1: wasi.wasiImport,
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

        if (probe.compiler === 'biggo') {
          startExitCode = wasi.start(instance);
          startReturned = true;
          if (startExitCode !== 0) {
            throw new Error(probe.file + ' _start exited with code ' + startExitCode);
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

  expect(results.map((result) => result.compiler)).toEqual(['biggo']);

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
