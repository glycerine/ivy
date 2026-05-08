import { createRequire } from 'node:module';
import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const require = createRequire(import.meta.url);
const testDir = path.dirname(fileURLToPath(import.meta.url));
const webvueDir = path.resolve(testDir, '../..');
const staticDir = path.join(webvueDir, 'static');
const z3GluePath = path.join(staticDir, 'z3-471-api.js');
const z3WasmPath = path.join(staticDir, 'z3-471-api.wasm');
const z3ExportListPath = path.join(staticDir, 'z3-wasm-exported-functions.json');

const wasmModuleNameExpectedByGlue = 'z3-api.wasm';

function readExportedFunctions() {
  return JSON.parse(fs.readFileSync(z3ExportListPath, 'utf8'));
}

function missingExports(requiredExports) {
  const exported = new Set(readExportedFunctions());
  return requiredExports.filter((name) => !exported.has(name));
}

function loadZ3Factory() {
  const source = fs.readFileSync(z3GluePath, 'utf8');
  const moduleObject = { exports: {} };
  const evaluateGlue = new Function(
    'module',
    'exports',
    'require',
    '__dirname',
    '__filename',
    `${source}\nreturn module.exports;`,
  );
  return evaluateGlue(moduleObject, moduleObject.exports, require, staticDir, z3GluePath);
}

async function instantiateZ3() {
  const initZ3 = loadZ3Factory();
  return initZ3({
    locateFile(file) {
      if (file === wasmModuleNameExpectedByGlue) {
        return z3WasmPath;
      }
      return path.join(staticDir, file);
    },
  });
}

function writeCString(z3, value) {
  const byteLength = Buffer.byteLength(value, 'utf8') + 1;
  const ptr = z3._malloc(byteLength);
  z3.stringToUTF8(value, ptr, byteLength);
  return ptr;
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
      '_Z3_get_error_msg',
      '_Z3_get_parser_error',
      '_Z3_parse_smtlib2_string',
      '_Z3_interrupt',
    ])).toEqual([]);
  });

  it('documents the legacy interpolation exports still missing from the wasm artifact', () => {
    expect(missingExports([
      '_Z3_mk_interpolation_context',
      '_Z3_mk_interpolant',
      '_Z3_compute_interpolant',
    ])).toEqual([
      '_Z3_mk_interpolation_context',
      '_Z3_mk_interpolant',
      '_Z3_compute_interpolant',
    ]);
  });

  it('instantiates Z3 wasm and runs a synchronous solver check', async () => {
    const z3 = await instantiateZ3();
    const cfg = z3._Z3_mk_config();
    const ctx = z3._Z3_mk_context_rc(cfg);
    z3._Z3_del_config(cfg);

    try {
      const symbolName = writeCString(z3, 'p');
      const symbol = z3._Z3_mk_string_symbol(ctx, symbolName);
      z3._free(symbolName);

      const boolSort = z3._Z3_mk_bool_sort(ctx);
      const p = z3._Z3_mk_const(ctx, symbol, boolSort);
      z3._Z3_inc_ref(ctx, p);

      const notP = z3._Z3_mk_not(ctx, p);
      z3._Z3_inc_ref(ctx, notP);

      const solver = z3._Z3_mk_solver(ctx);
      z3._Z3_solver_inc_ref(ctx, solver);
      z3._Z3_solver_assert(ctx, solver, p);
      z3._Z3_solver_assert(ctx, solver, notP);

      expect(z3._Z3_solver_check(ctx, solver)).toBe(-1);
      expect(z3.UTF8ToString(z3._Z3_solver_to_string(ctx, solver))).toContain('(assert');
    } finally {
      z3._Z3_del_context(ctx);
    }
  }, 30000);

  it('documents that the current build cannot safely use the SMT-LIB2 parser path yet', () => {
    const script = `
      const fs = require('fs');
      const path = require('path');
      const source = fs.readFileSync(${JSON.stringify(z3GluePath)}, 'utf8');
      const moduleObject = { exports: {} };
      const initZ3 = new Function(
        'module',
        'exports',
        'require',
        '__dirname',
        '__filename',
        source + '\\nreturn module.exports;',
      )(moduleObject, moduleObject.exports, require, ${JSON.stringify(staticDir)}, ${JSON.stringify(z3GluePath)});

      (async () => {
        const z3 = await initZ3({
          locateFile(file) {
            return file === ${JSON.stringify(wasmModuleNameExpectedByGlue)}
              ? ${JSON.stringify(z3WasmPath)}
              : path.join(${JSON.stringify(staticDir)}, file);
          },
        });
        const cfg = z3._Z3_mk_config();
        const ctx = z3._Z3_mk_context_rc(cfg);
        z3._Z3_del_config(cfg);
        const solver = z3._Z3_mk_solver(ctx);
        z3._Z3_solver_inc_ref(ctx, solver);
        const text = '(set-logic QF_UF)\\n(declare-const p Bool)\\n(assert p)\\n(assert (not p))\\n';
        const ptr = z3._malloc(Buffer.byteLength(text, 'utf8') + 1);
        z3.stringToUTF8(text, ptr, Buffer.byteLength(text, 'utf8') + 1);
        z3._Z3_solver_from_string(ctx, solver, ptr);
        console.log(z3._Z3_solver_check(ctx, solver));
      })();
    `;

    const result = spawnSync(process.execPath, ['-e', script], {
      encoding: 'utf8',
      timeout: 30000,
    });

    expect(result.status).not.toBe(0);
    expect(`${result.stderr}\n${result.stdout}`).toContain('Exception thrown, but exception catching is not enabled');
  }, 30000);
});
