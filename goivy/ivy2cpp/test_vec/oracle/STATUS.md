# ivy2cpp oracle fixture status

The oracle harness reads this table to decide whether a fixture must
pass today (`PASS`), is a known parity gap (`EXPECTED_FAIL`), or is
temporarily disabled (`SKIP`). `ORACLE_TEST=1` runs token parity against
Python; `SLOW_CPP_TEST=1` adds compile and transcript hooks.

| Fixture | Status | Notes |
| --- | --- | --- |
| empty.ivy | EXPECTED_FAIL | Baseline parity is intentionally enforced by the oracle harness before being promoted. |
| basic_assign.ivy | EXPECTED_FAIL | Tracks simple mutable-state codegen parity. |
| forall_assign.ivy | EXPECTED_FAIL | Tracks two-phase assignment and thunk lowering parity. |
| bv_arithmetic.ivy | EXPECTED_FAIL | Tracks bitvector xor/shift/neg operator parity. |
| enum_dispatch.ivy | EXPECTED_FAIL | Tracks enum action dispatch and return parsing parity. |
| range_bounds.ivy | EXPECTED_FAIL | Tracks range type bounds and argument conversion parity. |
| destructor_record.ivy | EXPECTED_FAIL | Tracks record/destructor type layout parity. |
| variant_simple.ivy | EXPECTED_FAIL | Tracks non-recursive variant layout and conversion parity. |
| variant_recursive.ivy | EXPECTED_FAIL | Tracks recursive variant layout parity. |
| hash_thunk_assign.ivy | EXPECTED_FAIL | Tracks large-domain hash thunk assignment parity. |
| native_block.ivy | EXPECTED_FAIL | Tracks native header/member/init/action antiquote parity. |
| callback_thunk.ivy | EXPECTED_FAIL | Tracks native callback thunk parity. |
| progress_property.ivy | EXPECTED_FAIL | Tracks progress/rely tick generation parity. |
| isolate_two_parts.ivy | EXPECTED_FAIL | Tracks multi-isolate extraction parity. |
