const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder();

function withZ3CString(z3, value, fn) {
  const bytes = textEncoder.encode(value);
  const ptr = z3._malloc(bytes.length + 1);
  z3.HEAPU8.set(bytes, ptr);
  z3.HEAPU8[ptr + bytes.length] = 0;
  try {
    return fn(ptr);
  } finally {
    z3._free(ptr);
  }
}

function withZ3HandleArray(z3, handles, fn) {
  if (handles.length === 0) {
    return fn(0);
  }
  const ptr = z3._malloc(handles.length * 4);
  for (let i = 0; i < handles.length; i += 1) {
    z3.HEAPU32[(ptr >>> 2) + i] = handles[i] >>> 0;
  }
  try {
    return fn(ptr);
  } finally {
    z3._free(ptr);
  }
}

function makeStringTable() {
  let next = 1;
  const byHandle = new Map();

  return {
    add(value) {
      return this.addBytes(textEncoder.encode(value));
    },
    addBytes(value) {
      const bytes = new Uint8Array(value);
      const handle = next;
      next += 1;
      byHandle.set(handle, bytes);
      return handle;
    },
    bytes(handle) {
      return byHandle.get(handle >>> 0) ?? new Uint8Array(0);
    },
    release(handle) {
      byHandle.delete(handle >>> 0);
    },
  };
}

function z3CStringBytes(z3, handle) {
  const ptr = handle >>> 0;
  if (!ptr) {
    return new Uint8Array(0);
  }
  const heap = z3.HEAPU8;
  if (!heap || ptr >= heap.length) {
    throw new Error('Z3 returned string pointer outside wasm heap: ' + ptr);
  }
  let end = ptr;
  while (end < heap.length && heap[end] !== 0) {
    end += 1;
  }
  if (end >= heap.length) {
    throw new Error('Z3 returned unterminated string pointer: ' + ptr);
  }
  return new Uint8Array(heap.subarray(ptr, end));
}

function makeScratchTable() {
  let next = 1;
  const byHandle = new Map();

  function add(kind, data) {
    const handle = next;
    next += 1;
    byHandle.set(handle, { kind, data });
    return handle;
  }

  function take(handle, kind) {
    const key = handle >>> 0;
    const entry = byHandle.get(key);
    if (!entry) {
      if (kind === 'bytes') {
        return new Uint8Array(0);
      }
      return new Uint32Array(0);
    }
    byHandle.delete(key);
    if (entry.kind !== kind) {
      throw new Error(`scratch handle ${key} is ${entry.kind}, not ${kind}`);
    }
    return entry.data;
  }

  return {
    beginBytes(size) {
      return add('bytes', new Uint8Array(size >>> 0));
    },
    writeBytes(handle, offset, word, n) {
      const entry = byHandle.get(handle >>> 0);
      if (!entry || entry.kind !== 'bytes') {
        throw new Error('bad byte scratch handle ' + handle);
      }
      const start = offset >>> 0;
      const count = n >>> 0;
      for (let i = 0; i < count; i += 1) {
        entry.data[start + i] = (word >>> (8 * i)) & 0xff;
      }
    },
    beginU32(size) {
      return add('u32', new Uint32Array(size >>> 0));
    },
    writeU32(handle, index, value) {
      const entry = byHandle.get(handle >>> 0);
      if (!entry || entry.kind !== 'u32') {
        throw new Error('bad u32 scratch handle ' + handle);
      }
      entry.data[index >>> 0] = value >>> 0;
    },
    takeBytes(handle, len) {
      const data = take(handle, 'bytes');
      return data.subarray(0, len >>> 0);
    },
    takeU32(handle, len) {
      const data = take(handle, 'u32');
      return Array.from(data.subarray(0, len >>> 0), (v) => v >>> 0);
    },
  };
}

function makeErrorTracker(z3, z3StringText, onError) {
  let callbackPtr = 0;
  const events = [];

  function ensureCallback() {
    if (callbackPtr) {
      return callbackPtr;
    }
    if (typeof z3.addFunction !== 'function') {
      throw new Error('Z3 wasm glue does not export addFunction for Z3_set_error_handler');
    }

    callbackPtr = z3.addFunction((ctx, code) => {
      const event = { ctx: ctx >>> 0, code: code | 0, message: '' };
      try {
        if (typeof z3._Z3_get_error_msg === 'function') {
          event.message = z3StringText(z3._Z3_get_error_msg(ctx, code));
        }
      } catch {
        event.message = '';
      }
      events.push(event);
      if (typeof onError === 'function') {
        try {
          onError({ ...event });
        } catch {
          // Error reporting must never become the Z3 error behavior.
        }
      }
    }, 'vii');

    return callbackPtr;
  }

  return {
    install(ctx) {
      if (typeof z3._Z3_set_error_handler !== 'function') {
        throw new Error('Z3 wasm does not export Z3_set_error_handler');
      }
      z3._Z3_set_error_handler(ctx, ensureCallback());
    },
    clear() {
      events.length = 0;
    },
    events() {
      return events.map((event) => ({ ...event }));
    },
  };
}

export function createSmtZ3Imports({ z3, getGoMemory, onError }) {
  const strings = makeStringTable();
  const scratch = makeScratchTable();
  let lastU32Results = [];
  let lastEnumConsts = [];
  let lastEnumTesters = [];
  void getGoMemory;

  function z3StringText(handle) {
    return textDecoder.decode(z3CStringBytes(z3, handle));
  }

  function z3StringHandle(handle) {
    return strings.addBytes(z3CStringBytes(z3, handle));
  }

  const errors = makeErrorTracker(z3, z3StringText, onError);

  function symbolFromScratch(ctx, handle, len) {
    return withZ3CString(z3, textDecoder.decode(scratch.takeBytes(handle, len)), (z3Ptr) => (
      z3._Z3_mk_string_symbol(ctx, z3Ptr) >>> 0
    ));
  }

  function z3StringAstFromScratch(ctx, handle, len) {
    return withZ3CString(z3, textDecoder.decode(scratch.takeBytes(handle, len)), (z3Ptr) => (
      z3._Z3_mk_string(ctx, z3Ptr) >>> 0
    ));
  }

  function paramValueFromScratch(cfg, keyHandle, keyLen, valueHandle, valueLen) {
    const key = textDecoder.decode(scratch.takeBytes(keyHandle, keyLen));
    const value = textDecoder.decode(scratch.takeBytes(valueHandle, valueLen));
    withZ3CString(z3, key, (z3Key) => {
      withZ3CString(z3, value, (z3Value) => {
        z3._Z3_set_param_value(cfg, z3Key, z3Value);
      });
    });
  }

  function mkNAry(name, ctx, len, handle) {
    const handles = scratch.takeU32(handle, len);
    return withZ3HandleArray(z3, handles, (z3Ptr) => z3[`_${name}`](ctx, len >>> 0, z3Ptr) >>> 0);
  }

  function mkFuncDecl(ctx, sym, domainLen, domainHandle, rangeSort) {
    const domain = scratch.takeU32(domainHandle, domainLen);
    return withZ3HandleArray(z3, domain, (z3Ptr) => (
      z3._Z3_mk_func_decl(ctx, sym, domain.length, z3Ptr, rangeSort) >>> 0
    ));
  }

  function mkApp(ctx, funcDecl, argLen, argHandle) {
    const args = scratch.takeU32(argHandle, argLen);
    return withZ3HandleArray(z3, args, (z3Ptr) => (
      z3._Z3_mk_app(ctx, funcDecl, args.length, z3Ptr) >>> 0
    ));
  }

  function mkEnumerationSort(ctx, name, len, elemsHandle) {
    const elems = scratch.takeU32(elemsHandle, len);
    return withZ3HandleArray(z3, elems, (z3ElemsPtr) => {
      const constsZ3Ptr = z3._malloc((len >>> 0) * 4);
      const testersZ3Ptr = z3._malloc((len >>> 0) * 4);
      try {
        const sort = z3._Z3_mk_enumeration_sort(ctx, name, len, z3ElemsPtr, constsZ3Ptr, testersZ3Ptr) >>> 0;
        lastEnumConsts = [];
        lastEnumTesters = [];
        for (let i = 0; i < (len >>> 0); i += 1) {
          lastEnumConsts.push(z3.HEAPU32[(constsZ3Ptr >>> 2) + i] >>> 0);
          lastEnumTesters.push(z3.HEAPU32[(testersZ3Ptr >>> 2) + i] >>> 0);
        }
        return sort;
      } finally {
        z3._free(constsZ3Ptr);
        z3._free(testersZ3Ptr);
      }
    });
  }

  function mkQuantifierConst(name, ctx, weight, len, boundHandle, numPatterns, patternsHandle, body) {
    const bound = scratch.takeU32(boundHandle, len);
    return withZ3HandleArray(z3, bound, (z3BoundPtr) => {
      void patternsHandle;
      return z3[`_${name}`](ctx, weight, bound.length, z3BoundPtr, numPatterns, 0, body) >>> 0;
    });
  }

  function substitute(ctx, expr, len, fromHandle, toHandle) {
    const from = scratch.takeU32(fromHandle, len);
    const to = scratch.takeU32(toHandle, len);
    return withZ3HandleArray(z3, from, (z3FromPtr) => (
      withZ3HandleArray(z3, to, (z3ToPtr) => (
        z3._Z3_substitute(ctx, expr, from.length, z3FromPtr, z3ToPtr) >>> 0
      ))
    ));
  }

  function checkAssumptions(ctx, solver, len, assumptionsHandle) {
    const assumptions = scratch.takeU32(assumptionsHandle, len);
    return withZ3HandleArray(z3, assumptions, (z3Ptr) => (
      z3._Z3_solver_check_assumptions(ctx, solver, assumptions.length, z3Ptr) | 0
    ));
  }

  function modelEval(ctx, model, expr, completion) {
    const outPtr = z3._malloc(4);
    try {
      const ok = z3._Z3_model_eval(ctx, model, expr, completion ? 1 : 0, outPtr) ? 1 : 0;
      lastU32Results = [z3.HEAPU32[outPtr >>> 2] >>> 0];
      return ok;
    } finally {
      z3._free(outPtr);
    }
  }

  function computeInterpolant(ctx, pattern, params) {
    const interpOutPtr = z3._malloc(4);
    const modelOutPtr = z3._malloc(4);
    try {
      const result = z3._Z3_compute_interpolant(ctx, pattern, params, interpOutPtr, modelOutPtr) | 0;
      lastU32Results = [
        z3.HEAPU32[interpOutPtr >>> 2] >>> 0,
        z3.HEAPU32[modelOutPtr >>> 2] >>> 0,
      ];
      return result;
    } finally {
      z3._free(interpOutPtr);
      z3._free(modelOutPtr);
    }
  }

  return {
    __scratch_bytes_begin(size) { return scratch.beginBytes(size); },
    __scratch_bytes_write(handle, offset, word, n) { scratch.writeBytes(handle, offset, word, n); },
    __scratch_u32_begin(size) { return scratch.beginU32(size); },
    __scratch_u32_write(handle, index, value) { scratch.writeU32(handle, index, value); },
    __last_u32_result(index) { return lastU32Results[index >>> 0] >>> 0; },
    __last_enum_const(index) { return lastEnumConsts[index >>> 0] >>> 0; },
    __last_enum_tester(index) { return lastEnumTesters[index >>> 0] >>> 0; },
    Z3_string_len(handle) { return strings.bytes(handle).length >>> 0; },
    Z3_string_word(handle, offset) {
      const bytes = strings.bytes(handle);
      const start = offset >>> 0;
      let word = 0;
      for (let i = 0; i < 4 && start + i < bytes.length; i += 1) {
        word |= bytes[start + i] << (8 * i);
      }
      return word >>> 0;
    },
    Z3_string_release(handle) { strings.release(handle); },
    Z3_mk_config() { return z3._Z3_mk_config() >>> 0; },
    Z3_del_config(cfg) { z3._Z3_del_config(cfg); },
    Z3_set_param_value_bytes: paramValueFromScratch,
    Z3_mk_context_rc(cfg) { return z3._Z3_mk_context_rc(cfg) >>> 0; },
    Z3_mk_interpolation_context(cfg) { return z3._Z3_mk_interpolation_context(cfg) >>> 0; },
    Z3_set_error_handler(ctx) { errors.install(ctx); },
    Z3_get_error_code(ctx) { return z3._Z3_get_error_code(ctx) | 0; },
    Z3_get_error_msg(ctx, code) { return z3StringHandle(z3._Z3_get_error_msg(ctx, code)); },
    __z3ClearErrors() { errors.clear(); },
    __z3ErrorEvents() { return errors.events(); },
    Z3_del_context(ctx) { z3._Z3_del_context(ctx); },
    Z3_inc_ref(ctx, ast) { z3._Z3_inc_ref(ctx, ast); },
    Z3_dec_ref(ctx, ast) { z3._Z3_dec_ref(ctx, ast); },
    Z3_mk_string_symbol_bytes: symbolFromScratch,
    Z3_get_symbol_string(ctx, sym) { return z3StringHandle(z3._Z3_get_symbol_string(ctx, sym)); },
    Z3_mk_bool_sort(ctx) { return z3._Z3_mk_bool_sort(ctx) >>> 0; },
    Z3_mk_uninterpreted_sort(ctx, sym) { return z3._Z3_mk_uninterpreted_sort(ctx, sym) >>> 0; },
    Z3_mk_int_sort(ctx) { return z3._Z3_mk_int_sort(ctx) >>> 0; },
    Z3_mk_real_sort(ctx) { return z3._Z3_mk_real_sort(ctx) >>> 0; },
    Z3_mk_bv_sort(ctx, width) { return z3._Z3_mk_bv_sort(ctx, width) >>> 0; },
    Z3_mk_string_sort(ctx) { return z3._Z3_mk_string_sort(ctx) >>> 0; },
    Z3_mk_array_sort(ctx, domain, range) { return z3._Z3_mk_array_sort(ctx, domain, range) >>> 0; },
    Z3_get_array_sort_domain(ctx, sort) { return z3._Z3_get_array_sort_domain(ctx, sort) >>> 0; },
    Z3_get_array_sort_range(ctx, sort) { return z3._Z3_get_array_sort_range(ctx, sort) >>> 0; },
    Z3_get_sort_name(ctx, sort) { return z3._Z3_get_sort_name(ctx, sort) >>> 0; },
    Z3_sort_to_ast(ctx, sort) { return z3._Z3_sort_to_ast(ctx, sort) >>> 0; },
    Z3_get_sort_kind(ctx, sort) { return z3._Z3_get_sort_kind(ctx, sort) | 0; },
    Z3_get_bv_sort_size(ctx, sort) { return z3._Z3_get_bv_sort_size(ctx, sort) >>> 0; },
    Z3_is_eq_sort(ctx, a, b) { return z3._Z3_is_eq_sort(ctx, a, b) ? 1 : 0; },
    Z3_mk_true(ctx) { return z3._Z3_mk_true(ctx) >>> 0; },
    Z3_mk_false(ctx) { return z3._Z3_mk_false(ctx) >>> 0; },
    Z3_mk_const(ctx, sym, sort) { return z3._Z3_mk_const(ctx, sym, sort) >>> 0; },
    Z3_mk_int64(ctx, value, sort) { return z3._Z3_mk_int64(ctx, value, sort) >>> 0; },
    Z3_mk_string_bytes: z3StringAstFromScratch,
    Z3_mk_not(ctx, expr) { return z3._Z3_mk_not(ctx, expr) >>> 0; },
    Z3_mk_and(ctx, len, ptr) { return mkNAry('Z3_mk_and', ctx, len, ptr); },
    Z3_mk_or(ctx, len, ptr) { return mkNAry('Z3_mk_or', ctx, len, ptr); },
    Z3_mk_implies(ctx, a, b) { return z3._Z3_mk_implies(ctx, a, b) >>> 0; },
    Z3_mk_iff(ctx, a, b) { return z3._Z3_mk_iff(ctx, a, b) >>> 0; },
    Z3_mk_eq(ctx, a, b) { return z3._Z3_mk_eq(ctx, a, b) >>> 0; },
    Z3_mk_ite(ctx, c, t, e) { return z3._Z3_mk_ite(ctx, c, t, e) >>> 0; },
    Z3_mk_add(ctx, len, ptr) { return mkNAry('Z3_mk_add', ctx, len, ptr); },
    Z3_mk_sub(ctx, len, ptr) { return mkNAry('Z3_mk_sub', ctx, len, ptr); },
    Z3_mk_mul(ctx, len, ptr) { return mkNAry('Z3_mk_mul', ctx, len, ptr); },
    Z3_mk_div(ctx, a, b) { return z3._Z3_mk_div(ctx, a, b) >>> 0; },
    Z3_mk_gt(ctx, a, b) { return z3._Z3_mk_gt(ctx, a, b) >>> 0; },
    Z3_mk_lt(ctx, a, b) { return z3._Z3_mk_lt(ctx, a, b) >>> 0; },
    Z3_mk_ge(ctx, a, b) { return z3._Z3_mk_ge(ctx, a, b) >>> 0; },
    Z3_mk_le(ctx, a, b) { return z3._Z3_mk_le(ctx, a, b) >>> 0; },
    Z3_mk_bvand(ctx, a, b) { return z3._Z3_mk_bvand(ctx, a, b) >>> 0; },
    Z3_mk_bvor(ctx, a, b) { return z3._Z3_mk_bvor(ctx, a, b) >>> 0; },
    Z3_mk_bvnot(ctx, e) { return z3._Z3_mk_bvnot(ctx, e) >>> 0; },
    Z3_mk_bvadd(ctx, a, b) { return z3._Z3_mk_bvadd(ctx, a, b) >>> 0; },
    Z3_mk_bvsub(ctx, a, b) { return z3._Z3_mk_bvsub(ctx, a, b) >>> 0; },
    Z3_mk_bvmul(ctx, a, b) { return z3._Z3_mk_bvmul(ctx, a, b) >>> 0; },
    Z3_mk_bvudiv(ctx, a, b) { return z3._Z3_mk_bvudiv(ctx, a, b) >>> 0; },
    Z3_mk_bvshl(ctx, a, b) { return z3._Z3_mk_bvshl(ctx, a, b) >>> 0; },
    Z3_mk_bvlshr(ctx, a, b) { return z3._Z3_mk_bvlshr(ctx, a, b) >>> 0; },
    Z3_mk_bvashr(ctx, a, b) { return z3._Z3_mk_bvashr(ctx, a, b) >>> 0; },
    Z3_mk_bvxor(ctx, a, b) { return z3._Z3_mk_bvxor(ctx, a, b) >>> 0; },
    Z3_mk_concat(ctx, a, b) { return z3._Z3_mk_concat(ctx, a, b) >>> 0; },
    Z3_mk_extract(ctx, hi, lo, e) { return z3._Z3_mk_extract(ctx, hi, lo, e) >>> 0; },
    Z3_mk_bv2int(ctx, e, signed) { return z3._Z3_mk_bv2int(ctx, e, signed ? 1 : 0) >>> 0; },
    Z3_mk_int2bv(ctx, width, e) { return z3._Z3_mk_int2bv(ctx, width, e) >>> 0; },
    Z3_mk_bvult(ctx, a, b) { return z3._Z3_mk_bvult(ctx, a, b) >>> 0; },
    Z3_mk_bvule(ctx, a, b) { return z3._Z3_mk_bvule(ctx, a, b) >>> 0; },
    Z3_mk_bvugt(ctx, a, b) { return z3._Z3_mk_bvugt(ctx, a, b) >>> 0; },
    Z3_mk_bvuge(ctx, a, b) { return z3._Z3_mk_bvuge(ctx, a, b) >>> 0; },
    Z3_mk_select(ctx, array, index) { return z3._Z3_mk_select(ctx, array, index) >>> 0; },
    Z3_mk_store(ctx, array, index, value) { return z3._Z3_mk_store(ctx, array, index, value) >>> 0; },
    Z3_mk_const_array(ctx, domain, value) { return z3._Z3_mk_const_array(ctx, domain, value) >>> 0; },
    Z3_mk_func_decl: mkFuncDecl,
    Z3_func_decl_to_ast(ctx, fd) { return z3._Z3_func_decl_to_ast(ctx, fd) >>> 0; },
    Z3_mk_app: mkApp,
    Z3_mk_enumeration_sort: mkEnumerationSort,
    Z3_to_app(ctx, ast) { return z3._Z3_to_app(ctx, ast) >>> 0; },
    Z3_mk_forall_const(ctx, weight, len, boundPtr, numPatterns, patternsPtr, body) {
      return mkQuantifierConst('Z3_mk_forall_const', ctx, weight, len, boundPtr, numPatterns, patternsPtr, body);
    },
    Z3_mk_exists_const(ctx, weight, len, boundPtr, numPatterns, patternsPtr, body) {
      return mkQuantifierConst('Z3_mk_exists_const', ctx, weight, len, boundPtr, numPatterns, patternsPtr, body);
    },
    Z3_substitute: substitute,
    Z3_ast_to_string(ctx, ast) { return z3StringHandle(z3._Z3_ast_to_string(ctx, ast)); },
    Z3_get_ast_id(ctx, ast) { return z3._Z3_get_ast_id(ctx, ast) >>> 0; },
    Z3_get_ast_kind(ctx, ast) { return z3._Z3_get_ast_kind(ctx, ast) | 0; },
    Z3_get_bool_value(ctx, ast) { return z3._Z3_get_bool_value(ctx, ast) | 0; },
    Z3_is_eq_ast(ctx, a, b) { return z3._Z3_is_eq_ast(ctx, a, b) ? 1 : 0; },
    Z3_get_index_value(ctx, ast) { return z3._Z3_get_index_value(ctx, ast) >>> 0; },
    Z3_get_numeral_string(ctx, ast) { return z3StringHandle(z3._Z3_get_numeral_string(ctx, ast)); },
    Z3_get_sort(ctx, ast) { return z3._Z3_get_sort(ctx, ast) >>> 0; },
    Z3_get_app_num_args(ctx, app) { return z3._Z3_get_app_num_args(ctx, app) >>> 0; },
    Z3_get_app_arg(ctx, app, index) { return z3._Z3_get_app_arg(ctx, app, index) >>> 0; },
    Z3_get_app_decl(ctx, app) { return z3._Z3_get_app_decl(ctx, app) >>> 0; },
    Z3_get_decl_name(ctx, fd) { return z3._Z3_get_decl_name(ctx, fd) >>> 0; },
    Z3_get_decl_kind(ctx, fd) { return z3._Z3_get_decl_kind(ctx, fd) | 0; },
    Z3_get_arity(ctx, fd) { return z3._Z3_get_arity(ctx, fd) >>> 0; },
    Z3_get_domain(ctx, fd, i) { return z3._Z3_get_domain(ctx, fd, i) >>> 0; },
    Z3_get_range(ctx, fd) { return z3._Z3_get_range(ctx, fd) >>> 0; },
    Z3_is_quantifier_forall(ctx, ast) { return z3._Z3_is_quantifier_forall(ctx, ast) ? 1 : 0; },
    Z3_get_quantifier_num_bound(ctx, ast) { return z3._Z3_get_quantifier_num_bound(ctx, ast) >>> 0; },
    Z3_get_quantifier_bound_name(ctx, ast, i) { return z3._Z3_get_quantifier_bound_name(ctx, ast, i) >>> 0; },
    Z3_get_quantifier_bound_sort(ctx, ast, i) { return z3._Z3_get_quantifier_bound_sort(ctx, ast, i) >>> 0; },
    Z3_get_quantifier_body(ctx, ast) { return z3._Z3_get_quantifier_body(ctx, ast) >>> 0; },
    Z3_mk_solver(ctx) { return z3._Z3_mk_solver(ctx) >>> 0; },
    Z3_mk_solver_for_logic(ctx, logic) { return z3._Z3_mk_solver_for_logic(ctx, logic) >>> 0; },
    Z3_solver_inc_ref(ctx, solver) { z3._Z3_solver_inc_ref(ctx, solver); },
    Z3_solver_dec_ref(ctx, solver) { z3._Z3_solver_dec_ref(ctx, solver); },
    Z3_solver_assert(ctx, solver, expr) { z3._Z3_solver_assert(ctx, solver, expr); },
    Z3_solver_check(ctx, solver) { return z3._Z3_solver_check(ctx, solver) | 0; },
    Z3_solver_push(ctx, solver) { z3._Z3_solver_push(ctx, solver); },
    Z3_solver_pop(ctx, solver, n) { z3._Z3_solver_pop(ctx, solver, n); },
    Z3_solver_reset(ctx, solver) { z3._Z3_solver_reset(ctx, solver); },
    Z3_solver_to_string(ctx, solver) { return z3StringHandle(z3._Z3_solver_to_string(ctx, solver)); },
    Z3_solver_get_assertions(ctx, solver) { return z3._Z3_solver_get_assertions(ctx, solver) >>> 0; },
    Z3_solver_get_model(ctx, solver) { return z3._Z3_solver_get_model(ctx, solver) >>> 0; },
    Z3_solver_check_assumptions: checkAssumptions,
    Z3_solver_get_unsat_core(ctx, solver) { return z3._Z3_solver_get_unsat_core(ctx, solver) >>> 0; },
    Z3_solver_set_params(ctx, solver, params) { z3._Z3_solver_set_params(ctx, solver, params); },
    Z3_ast_vector_inc_ref(ctx, vec) { z3._Z3_ast_vector_inc_ref(ctx, vec); },
    Z3_ast_vector_dec_ref(ctx, vec) { z3._Z3_ast_vector_dec_ref(ctx, vec); },
    Z3_ast_vector_size(ctx, vec) { return z3._Z3_ast_vector_size(ctx, vec) >>> 0; },
    Z3_ast_vector_get(ctx, vec, i) { return z3._Z3_ast_vector_get(ctx, vec, i) >>> 0; },
    Z3_model_inc_ref(ctx, model) { z3._Z3_model_inc_ref(ctx, model); },
    Z3_model_dec_ref(ctx, model) { z3._Z3_model_dec_ref(ctx, model); },
    Z3_model_eval: modelEval,
    Z3_model_get_num_sorts(ctx, model) { return z3._Z3_model_get_num_sorts(ctx, model) >>> 0; },
    Z3_model_get_sort(ctx, model, i) { return z3._Z3_model_get_sort(ctx, model, i) >>> 0; },
    Z3_model_get_sort_universe(ctx, model, sort) { return z3._Z3_model_get_sort_universe(ctx, model, sort) >>> 0; },
    Z3_model_to_string(ctx, model) { return z3StringHandle(z3._Z3_model_to_string(ctx, model)); },
    Z3_mk_params(ctx) { return z3._Z3_mk_params(ctx) >>> 0; },
    Z3_params_inc_ref(ctx, params) { z3._Z3_params_inc_ref(ctx, params); },
    Z3_params_dec_ref(ctx, params) { z3._Z3_params_dec_ref(ctx, params); },
    Z3_params_set_bool(ctx, params, key, value) { z3._Z3_params_set_bool(ctx, params, key, value ? 1 : 0); },
    Z3_params_set_uint(ctx, params, key, value) { z3._Z3_params_set_uint(ctx, params, key, value >>> 0); },
    Z3_params_set_symbol(ctx, params, key, value) { z3._Z3_params_set_symbol(ctx, params, key, value); },
    Z3_mk_interpolant(ctx, expr) { return z3._Z3_mk_interpolant(ctx, expr) >>> 0; },
    Z3_compute_interpolant: computeInterpolant,
  };
}
