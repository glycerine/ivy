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

test('Z3 wasm computes a legacy interpolant in a browser runtime', async ({ page }) => {
  await page.goto(baseURL);
  await page.addScriptTag({ url: `${baseURL}/z3-471-api.js` });

  const result = await page.evaluate(async () => {
    function writeCString(z3, value) {
      const byteLength = new TextEncoder().encode(value).length + 1;
      const ptr = z3._malloc(byteLength);
      z3.stringToUTF8(value, ptr, byteLength);
      return ptr;
    }

    function mkBoolVar(z3, ctx, name) {
      const symbolName = writeCString(z3, name);
      const symbol = z3._Z3_mk_string_symbol(ctx, symbolName);
      z3._free(symbolName);
      return z3._Z3_mk_const(ctx, symbol, z3._Z3_mk_bool_sort(ctx));
    }

    function mkAnd(z3, ctx, args) {
      const argsPtr = z3._malloc(4 * args.length);
      for (let i = 0; i < args.length; i += 1) {
        z3.HEAP32[(argsPtr >> 2) + i] = args[i];
      }
      const ast = z3._Z3_mk_and(ctx, args.length, argsPtr);
      z3._free(argsPtr);
      return ast;
    }

    const z3 = await window.initZ3({
      locateFile(file) {
        if (file === 'z3-api.wasm') {
          return '/z3-471-api.wasm';
        }
        return file;
      },
    });

    const cfg = z3._Z3_mk_config();
    const ctx = z3._Z3_mk_interpolation_context(cfg);
    z3._Z3_del_config(cfg);

    const predA = mkBoolVar(z3, ctx, 'PredA');
    const predB = mkBoolVar(z3, ctx, 'PredB');
    const predC = mkBoolVar(z3, ctx, 'PredC');
    const markedLeft = z3._Z3_mk_interpolant(ctx, mkAnd(z3, ctx, [predA, predB]));
    const right = mkAnd(z3, ctx, [z3._Z3_mk_not(ctx, predB), predC]);
    const formula = mkAnd(z3, ctx, [markedLeft, right]);
    const interpolationOut = z3._malloc(4);
    const modelOut = z3._malloc(4);
    z3.HEAP32[interpolationOut >> 2] = 0;
    z3.HEAP32[modelOut >> 2] = 0;

    try {
      const status = z3._Z3_compute_interpolant(ctx, formula, 0, interpolationOut, modelOut);
      const interpolationVector = z3.HEAP32[interpolationOut >> 2];
      const vectorSize = interpolationVector ? z3._Z3_ast_vector_size(ctx, interpolationVector) : 0;
      const interpolant = interpolationVector ? z3._Z3_ast_vector_get(ctx, interpolationVector, 0) : 0;
      const interpolantText = interpolant ? z3.UTF8ToString(z3._Z3_ast_to_string(ctx, interpolant)) : '';

      return { status, vectorSize, interpolantText };
    } finally {
      z3._free(interpolationOut);
      z3._free(modelOut);
      z3._Z3_del_context(ctx);
    }
  });

  expect(result.status).toBe(-1);
  expect(result.vectorSize).toBe(1);
  expect(result.interpolantText.length).toBeGreaterThan(0);
});
