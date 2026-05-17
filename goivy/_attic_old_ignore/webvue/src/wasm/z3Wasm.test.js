import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { describe, expect, it } from 'vitest';

const testDir = path.dirname(fileURLToPath(import.meta.url));
const webvueDir = path.resolve(testDir, '../..');
const staticDir = path.join(webvueDir, 'static');
const z3GluePath = path.join(staticDir, 'z3-471-api.js');
const z3WasmPath = path.join(staticDir, 'z3-471-api.wasm');
const z3ExportListPath = path.join(staticDir, 'z3-wasm-exported-functions.json');
const smtZ3ImportsURL = pathToFileURL(path.join(webvueDir, 'src/workers/smtZ3Imports.js')).href;

const wasmModuleNameExpectedByGlue = 'z3-api.wasm';
const legacyInterpolationExports = [
  '_Z3_mk_interpolant',
  '_Z3_mk_interpolation_context',
  '_Z3_get_interpolant',
  '_Z3_compute_interpolant',
  '_Z3_interpolation_profile',
  '_Z3_read_interpolation_problem',
  '_Z3_check_interpolant',
  '_Z3_write_interpolation_problem',
];

function readExportedFunctions() {
  return JSON.parse(fs.readFileSync(z3ExportListPath, 'utf8'));
}

function missingExports(requiredExports) {
  const exported = new Set(readExportedFunctions());
  return requiredExports.filter((name) => !exported.has(name));
}

function runNodeWithWasmExceptionHandling(script) {
  const result = spawnSync(process.execPath, ['--experimental-wasm-exnref', '-e', script], {
    cwd: webvueDir,
    encoding: 'utf8',
    timeout: 30000,
  });
  const optionFailure = result.status === 9
    && /bad option|unrecognized|not allowed/.test(result.stderr || '');

  if (!optionFailure) {
    return result;
  }

  return spawnSync(process.execPath, ['-e', script], {
    cwd: webvueDir,
    encoding: 'utf8',
    timeout: 30000,
  });
}

function expectNodeScriptToPass(result) {
  expect(result.error).toBeUndefined();
  if (result.status !== 0) {
    throw new Error(`Node Z3 harness failed with status ${result.status}.\nstdout:\n${result.stdout}\nstderr:\n${result.stderr}`);
  }
}

function z3NodeScript(body) {
  return `
    const fs = require('fs');
    const path = require('path');
    const staticDir = ${JSON.stringify(staticDir)};
    const z3GluePath = ${JSON.stringify(z3GluePath)};
    const z3WasmPath = ${JSON.stringify(z3WasmPath)};
    const wasmModuleNameExpectedByGlue = ${JSON.stringify(wasmModuleNameExpectedByGlue)};
    const source = fs.readFileSync(z3GluePath, 'utf8');
    const moduleObject = { exports: {} };
    const initZ3 = new Function(
      'module',
      'exports',
      'require',
      '__dirname',
      '__filename',
      source + '\\nreturn module.exports;',
    )(moduleObject, moduleObject.exports, require, staticDir, z3GluePath);

    function writeCString(z3, value) {
      const byteLength = Buffer.byteLength(value, 'utf8') + 1;
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

    (async () => {
      const z3 = await initZ3({
        locateFile(file) {
          if (file === wasmModuleNameExpectedByGlue) {
            return z3WasmPath;
          }
          return path.join(staticDir, file);
        },
      });
      ${body}
    })().catch((err) => {
      console.error(err && err.stack ? err.stack : err);
      process.exit(1);
    });
  `;
}

describe('Z3 wasm artifact', () => {
  it('keeps the generated export list beside the Z3 wasm artifacts', () => {
    expect(fs.existsSync(z3GluePath)).toBe(true);
    expect(fs.existsSync(z3WasmPath)).toBe(true);
    expect(fs.existsSync(z3ExportListPath)).toBe(true);

    const exportedFunctions = readExportedFunctions();
    expect(exportedFunctions.length).toBeGreaterThan(600);
    expect(exportedFunctions).toContain('_Z3_solver_check');
  });

  it('exports the non-interpolation C API calls needed by the first Go-wasm bridge spike', () => {
    expect(missingExports([
      '_malloc',
      '_free',
      '_Z3_mk_config',
      '_Z3_del_config',
      '_Z3_mk_context_rc',
      '_Z3_del_context',
      '_Z3_mk_string_symbol',
      '_Z3_mk_bool_sort',
      '_Z3_mk_const',
      '_Z3_mk_not',
      '_Z3_mk_and',
      '_Z3_inc_ref',
      '_Z3_dec_ref',
      '_Z3_mk_solver',
      '_Z3_solver_inc_ref',
      '_Z3_solver_dec_ref',
      '_Z3_solver_assert',
      '_Z3_solver_from_string',
      '_Z3_solver_check',
      '_Z3_solver_push',
      '_Z3_solver_pop',
      '_Z3_solver_reset',
      '_Z3_solver_to_string',
      '_Z3_solver_get_model',
      '_Z3_model_inc_ref',
      '_Z3_model_dec_ref',
      '_Z3_model_eval',
      '_Z3_model_to_string',
      '_Z3_solver_check_assumptions',
      '_Z3_solver_get_unsat_core',
      '_Z3_ast_vector_size',
      '_Z3_ast_vector_get',
      '_Z3_ast_vector_inc_ref',
      '_Z3_ast_vector_dec_ref',
      '_Z3_mk_params',
      '_Z3_params_inc_ref',
      '_Z3_params_dec_ref',
      '_Z3_solver_set_params',
      '_Z3_get_error_code',
      '_Z3_set_error_handler',
      '_Z3_get_error_msg',
      '_Z3_get_parser_error',
      '_Z3_parse_smtlib2_string',
      '_Z3_interrupt',
    ])).toEqual([]);
  });

  it('exports the legacy interpolation C API needed by Ivy', () => {
    expect(missingExports(legacyInterpolationExports)).toEqual([]);
  });

  it('instantiates Z3 wasm with native Wasm EH and runs a synchronous solver check', () => {
    const result = runNodeWithWasmExceptionHandling(z3NodeScript(`
      const cfg = z3._Z3_mk_config();
      const ctx = z3._Z3_mk_context_rc(cfg);
      z3._Z3_del_config(cfg);

      try {
        const p = mkBoolVar(z3, ctx, 'p');
        z3._Z3_inc_ref(ctx, p);

        const notP = z3._Z3_mk_not(ctx, p);
        z3._Z3_inc_ref(ctx, notP);

        const solver = z3._Z3_mk_solver(ctx);
        z3._Z3_solver_inc_ref(ctx, solver);
        z3._Z3_solver_assert(ctx, solver, p);
        z3._Z3_solver_assert(ctx, solver, notP);

        console.log('solver check', z3._Z3_solver_check(ctx, solver));
        console.log(z3.UTF8ToString(z3._Z3_solver_to_string(ctx, solver)));
      } finally {
        z3._Z3_del_context(ctx);
      }
    `));

    expectNodeScriptToPass(result);
    expect(result.stdout).toContain('solver check -1');
    expect(result.stdout).toContain('(assert');
  }, 30000);

  it('accepts BigInt int64 values through the smt Z3 import bridge', () => {
    const result = runNodeWithWasmExceptionHandling(z3NodeScript(`
      const { createSmtZ3Imports } = await import(${JSON.stringify(smtZ3ImportsURL)});
      const cfg = z3._Z3_mk_config();
      const ctx = z3._Z3_mk_context_rc(cfg);
      z3._Z3_del_config(cfg);

      try {
        const bridge = createSmtZ3Imports({ z3, getGoMemory: () => null });
        const intSort = bridge.Z3_mk_int_sort(ctx);
        const cases = [
          [1n, '1'],
          [-1n, '(- 1)'],
          [2147483648n, '2147483648'],
          [-2147483649n, '(- 2147483649)'],
        ];

        for (const [value, expected] of cases) {
          const ast = bridge.Z3_mk_int64(ctx, value, intSort);
          const text = z3.UTF8ToString(z3._Z3_ast_to_string(ctx, ast));
          if (text !== expected) {
            throw new Error('Z3_mk_int64(' + value + ') rendered as ' + text + ', expected ' + expected);
          }
        }

        console.log('smt bigint int64 bridge ok');
      } finally {
        z3._Z3_del_context(ctx);
      }
    `));

    expectNodeScriptToPass(result);
    expect(result.stdout).toContain('smt bigint int64 bridge ok');
  }, 30000);

  it('computes a legacy interpolation result through the JavaScript glue', () => {
    const result = runNodeWithWasmExceptionHandling(z3NodeScript(`
      for (const name of ${JSON.stringify(legacyInterpolationExports)}) {
        if (typeof z3[name] !== 'function') {
          throw new Error(name + ' is not callable through the Emscripten glue');
        }
      }

      const cfg = z3._Z3_mk_config();
      const modelKey = writeCString(z3, 'model');
      const modelValue = writeCString(z3, 'true');
      z3._Z3_set_param_value(cfg, modelKey, modelValue);
      z3._free(modelKey);
      z3._free(modelValue);

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
        if (status !== -1) {
          throw new Error('expected unsat interpolation status -1, got ' + status);
        }
        if (!interpolationVector) {
          throw new Error('Z3_compute_interpolant did not return an AST vector');
        }

        const vectorSize = z3._Z3_ast_vector_size(ctx, interpolationVector);
        if (vectorSize !== 1) {
          throw new Error('expected one interpolant, got ' + vectorSize);
        }

        const interpolant = z3._Z3_ast_vector_get(ctx, interpolationVector, 0);
        const interpolantText = z3.UTF8ToString(z3._Z3_ast_to_string(ctx, interpolant));
        if (!interpolantText.length) {
          throw new Error('interpolant string was empty');
        }
        console.log('interpolation compute ok', status, vectorSize, interpolantText);
      } finally {
        z3._free(interpolationOut);
        z3._free(modelOut);
        z3._Z3_del_context(ctx);
      }
    `));

    expectNodeScriptToPass(result);
    expect(result.stdout).toContain('interpolation compute ok -1 1');
  }, 30000);

  it('runs a minimal SMT-LIB2 parser path without corrupting wasm memory', () => {
    const result = runNodeWithWasmExceptionHandling(z3NodeScript(`
        const cfg = z3._Z3_mk_config();
        const ctx = z3._Z3_mk_context_rc(cfg);
        z3._Z3_del_config(cfg);
        const solver = z3._Z3_mk_solver(ctx);
        z3._Z3_solver_inc_ref(ctx, solver);
        const text = '(set-logic QF_UF)\\n(declare-const p Bool)\\n(assert p)\\n(assert (not p))\\n';
        const ptr = z3._malloc(Buffer.byteLength(text, 'utf8') + 1);
        try {
          z3.stringToUTF8(text, ptr, Buffer.byteLength(text, 'utf8') + 1);
          z3._Z3_solver_from_string(ctx, solver, ptr);
          console.log('solver from string check', z3._Z3_solver_check(ctx, solver));
        } finally {
          z3._free(ptr);
          z3._Z3_solver_dec_ref(ctx, solver);
          z3._Z3_del_context(ctx);
        }
    `));

    expectNodeScriptToPass(result);
    expect(result.stdout).toContain('solver from string check -1');
  }, 30000);
});
