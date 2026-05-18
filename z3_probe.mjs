import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const [, , jsPath, wasmPath] = process.argv;
if (!jsPath || !wasmPath) {
  console.error('usage: node z3_probe.mjs z3-api.js z3-api.wasm');
  process.exit(2);
}

function loadZ3Glue(file) {
  const source = fs.readFileSync(file, 'utf8');
  const moduleObject = { exports: {} };
  return new Function(
    'module',
    'exports',
    'require',
    '__dirname',
    '__filename',
    source + '\nreturn module.exports;',
  )(moduleObject, moduleObject.exports, require, path.dirname(file), file);
}

function u32(mod, ptr, value) {
  mod.HEAPU32[ptr >>> 2] = value >>> 0;
}

const initZ3 = loadZ3Glue(jsPath);
const z3 = await initZ3({
  print: (text) => console.log(text),
  printErr: (text) => console.error(text),
  locateFile(file) {
    if (file === 'z3-api.wasm') return wasmPath;
    return path.join(path.dirname(jsPath), file);
  },
});

try {
  const cfg = z3._Z3_mk_config();
  const ctx = z3._Z3_mk_context(cfg);
  z3._Z3_del_config(cfg);

  const symU = z3._Z3_mk_string_symbol(ctx, 'U');
  const sortU = z3._Z3_mk_uninterpreted_sort(ctx, symU);
  const boolSort = z3._Z3_mk_bool_sort(ctx);

  const domain = z3._malloc(4);
  u32(z3, domain, sortU);
  const pred = z3._Z3_mk_func_decl(ctx, z3._Z3_mk_string_symbol(ctx, 'P'), 1, domain, boolSort);

  const x = z3._Z3_mk_const(ctx, z3._Z3_mk_string_symbol(ctx, 'x'), sortU);
  const args = z3._malloc(4);
  u32(z3, args, x);
  const body = z3._Z3_mk_app(ctx, pred, 1, args);

  const bound = z3._malloc(4);
  u32(z3, bound, z3._Z3_to_app(ctx, x));
  const forall = z3._Z3_mk_forall_const(ctx, 0, 1, bound, 0, 0, body);

  const solver = z3._Z3_mk_solver(ctx);
  z3._Z3_solver_inc_ref(ctx, solver);
  z3._Z3_solver_assert(ctx, solver, forall);
  const result = z3._Z3_solver_check(ctx, solver);
  console.log(`check-result=${result}`);
} catch (err) {
  console.error(`threw=${err && err.stack ? err.stack : String(err)}`);
  process.exit(1);
}
