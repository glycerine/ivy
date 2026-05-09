import { expect, test } from '@playwright/test';
import fs from 'node:fs/promises';
import http from 'node:http';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const webvueDir = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const nodeModulesDir = path.join(webvueDir, 'node_modules');
const srcDir = path.join(webvueDir, 'src');
const staticDir = path.join(webvueDir, 'static');
const solverWasmFile = process.env.SMT_SOLVER_ROUND_TRIP_WASM;
const errorBoundaryWasmFile = process.env.SMT_ERROR_BOUNDARY_WASM;

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

async function runSmtWasip1Fixture(page, wasmFile, exportNames) {
  await page.goto(baseURL);

  return await page.evaluate(async ({ wasmFile, exportNames }) => {
    const assetBaseURL = window.location.origin;
    const workerSource = `
      const assetBaseURL = ${JSON.stringify(assetBaseURL)};
      const wasmFile = ${JSON.stringify(wasmFile)};
      const exportNames = ${JSON.stringify(exportNames)};

      self.onmessage = async () => {
        let z3;

        try {
          const { createSmtZ3Imports } = await import(assetBaseURL + '/src/workers/smtZ3Imports.js');
          const { createGoIvyWasiP1 } = await import(assetBaseURL + '/src/workers/goivyWasiP1.js');

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
          const wasi = createGoIvyWasiP1({
            args: [wasmFile],
            includeRoot: '.',
            includeTree: { root: '.', files: [] },
            stdout() {},
            stderr() {},
          });
          const imports = {
            smt_z3: createSmtZ3Imports({ z3, getGoMemory: () => wasmMemory }),
            wasi_snapshot_preview1: wasi.wasiImport,
          };

          const response = await fetch(assetBaseURL + '/' + wasmFile, { cache: 'no-store' });
          if (!response.ok) {
            throw new Error('could not fetch ' + wasmFile + ': ' + response.status);
          }

          const bytes = await response.arrayBuffer();
          const { instance } = await WebAssembly.instantiate(bytes, imports);
          wasmMemory = instance.exports.memory;

          const startCode = wasi.start(instance);
          if (startCode !== 0) {
            throw new Error(wasmFile + ' _start exited with code ' + startCode);
          }

          imports.smt_z3.__z3ClearErrors();

          const results = {};
          for (const exportName of exportNames) {
            const exported = instance.exports[exportName];
            if (typeof exported !== 'function') {
              throw new Error('SMT wasip1 fixture is missing ' + exportName);
            }
            results[exportName] = exported() | 0;
          }

          self.postMessage({
            ok: true,
            results,
            z3Errors: imports.smt_z3.__z3ErrorEvents(),
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
  }, { wasmFile, exportNames });
}

test('Big Go WASI smt package solves a tiny query through Z3 wasm in a worker', async ({ page }) => {
  test.skip(!solverWasmFile, 'SMT_SOLVER_ROUND_TRIP_WASM is generated by the smt Go web test');

  const result = await runSmtWasip1Fixture(page, solverWasmFile, [
    'smt_solver_sat_true',
    'smt_solver_unsat_true_and_not_true',
    'smt_solver_uninterpreted_sort_name_ok',
  ]);

  expect(result.results.smt_solver_sat_true).toBe(1);
  expect(result.results.smt_solver_unsat_true_and_not_true).toBe(-1);
  expect(result.results.smt_solver_uninterpreted_sort_name_ok).toBe(1);
});

test('Big Go WASI smt package records Z3 errors through the browser callback boundary', async ({ page }) => {
  test.skip(!errorBoundaryWasmFile, 'SMT_ERROR_BOUNDARY_WASM is generated by the smt Go web test');

  const result = await runSmtWasip1Fixture(page, errorBoundaryWasmFile, [
    'smt_z3_invalid_bv_sort_returns_to_caller',
  ]);

  expect(result.results.smt_z3_invalid_bv_sort_returns_to_caller).toBe(1);
  expect(result.z3Errors.map((event) => event.code)).toContain(3);
  expect(result.z3Errors.some((event) => /invalid argument/i.test(event.message))).toBe(true);
});
