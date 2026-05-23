# ivy2cpp TODO3: remaining whole-hog Python parity

Created: 2026-05-23

This file tracks what remains after AUDIT2 items 031-046 were
implemented and the Python/Go oracle harness was added.

The oracle harness is live:

```sh
env XTRACE_OFF=1 ORACLE_TEST=1 go test ./ivy2cpp -run TestOracleSingle -v -count=1
```

Current result: all 14 oracle fixtures run, and all are still marked
`EXPECTED_FAIL` in `testdata/oracle/STATUS.md`. After fixing the initial
class-name/file-name mismatch in the harness, the first concrete
divergence for every fixture is in the shared runtime scaffold.

## TODO3-001 - Make the Go runtime scaffold token-equivalent to Python

Observed first divergence:

- Go `*.cpp` starts with `<algorithm>`, `<fstream>`, `<iostream>`, ...
- Python `*.cpp` starts with `<sstream>`, `<algorithm>`, blank line,
  `<iostream>`, ...

Required work:

- Align `emitRuntimeImplPreamble` in `runtime.go` with Python
  `ivy_to_cpp.py` output order:
  - `#include <sstream>`
  - `#include <algorithm>`
  - blank line
  - `#include <iostream>`
  - no `<fstream>` in the impl preamble unless Python emits it for a
    specific target
  - conditional `#include <cstdint>` block exactly like Python
  - `#include "ivy_threads.hpp"` after `__ivy_exit`, not in the header
  - `#include "ivy_value.hpp"` after lock/unlock definitions for impl
    target
- Align `emitHeaderPreamble` with Python:
  - Python does not emit `#pragma once` in the observed oracle output.
  - Python header starts with `_HAS_ITERATOR_DEBUGGING`, `ivy_hash.hpp`,
    `typedef std::string __strlit`, externs, then thunk/hash_thunk.
  - The Go header currently emits broad C++ system includes and
    `ivy_threads.hpp` up front.
- Align support struct text:
  - Python `thunk` has no virtual destructor in the observed output.
  - Python `hash_thunk` destructor keeps the commented-out delete block.
  - Python `hash_thunk` has no `operator==` in the observed output.
  - Iterator type spellings and spacing differ.
- Align class skeleton text and member order:
  - Python emits `int ___ivy_choose(...)` before assertion hooks.
  - Python target=impl output currently includes `_generating`; Go does
    not for target=impl.
  - Python method argument names differ (`timeout`, unnamed booleans).

Acceptance:

- `ORACLE_TEST=1 go test ./ivy2cpp -run TestOracleSingle/empty.ivy -v -count=1`
  must pass with `empty.ivy` promoted to `PASS`.

## TODO3-002 - Promote oracle fixtures one by one after runtime parity

Once `empty.ivy` passes, rerun:

```sh
env XTRACE_OFF=1 ORACLE_TEST=1 go test ./ivy2cpp -run TestOracleSingle -v -count=1
```

For each fixture:

- Fix the next token divergence in the Go generator, preferring Python's
  exact `ivy_to_cpp.py` behavior over local style.
- Add or update a focused unit test for that divergence.
- Promote the fixture from `EXPECTED_FAIL` to `PASS` in
  `testdata/oracle/STATUS.md`.
- Keep going until all 14 fixtures are `PASS`.

## TODO3-003 - Turn the oracle status policy into a hard gate

After all fixtures are `PASS`:

- Make `TestOracleSingle` fail if any fixture remains `EXPECTED_FAIL`.
- Run and fix the slow hooks:

```sh
env XTRACE_OFF=1 SLOW_CPP_TEST=1 go test ./ivy2cpp -run 'TestOracleCompileGo|TestOracleCompilePython|TestOracleSemanticEquivalence' -count=1
```

- Add `.in` transcripts for exported-action REPL fixtures so semantic
  equivalence is exercised instead of skipped.
