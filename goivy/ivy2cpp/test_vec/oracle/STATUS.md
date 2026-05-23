# ivy2cpp oracle fixture status

The oracle harness reads this table to decide whether a fixture must
pass today (`PASS`), is a known parity gap (`EXPECTED_FAIL`), or is
temporarily disabled (`SKIP`). `ORACLE_TEST=1` runs token parity against
Python; `SLOW_CPP_TEST=1` adds compile and transcript hooks.

| Fixture | Status | Notes |
| --- | --- | --- |
| empty.ivy | PASS | Baseline runtime scaffold matches the Python oracle. |
| basic_assign.ivy | PASS | Simple mutable-state codegen matches the Python oracle. |
| forall_assign.ivy | PASS | Two-phase assignment and post-isolate sort helper emission match the Python oracle. |
| bv_arithmetic.ivy | PASS | Declared bitvector helper functions and hash-thunk initialization match the Python oracle. |
| enum_dispatch.ivy | PASS | Enum action dispatch, return initialization, and helper placement match the Python oracle. |
| range_bounds.ivy | PASS | Range bounds and nondet initialization match the Python oracle. |
| destructor_record.ivy | PASS | Record/destructor layout, restored exported formals, and helper placement match the Python oracle. |
| variant_simple.ivy | PASS | Tracks non-recursive variant layout and conversion parity. |
| variant_recursive.ivy | PASS | Tracks recursive variant layout parity. |
| hash_thunk_assign.ivy | PASS | Tracks large-domain hash thunk assignment parity. |
| native_block.ivy | PASS | Tracks native header/member/init/action antiquote parity. |
| callback_thunk.ivy | PASS | Tracks native callback thunk parity. |
| progress_property.ivy | PASS | Tracks progress/rely tick generation parity. |
| isolate_two_parts.ivy | PASS | Tracks multi-isolate extraction parity. |
