# ivy2cpp TODO4: full target=test token parity

Created: 2026-05-23

The emitted tester loop has been ported to Python's literal `vector<gen *>`
dispatch shape, and generated tester semantic fixtures now run through the
oracle harness.

Remaining full-file `target=test` token parity work:

- Port the target=test Z3 scaffold from Go's `ivy_go_z3.hpp` helper style to
  Python's emitted `ivy_z3_gen.hpp` / `ivy_z3_helpers.hpp` ordering and calls.
- Align the remaining REPL/test main scaffold tokens outside the generated
  test loop, including one-line condition bodies and Python's init flow.
- `variant_simple.ivy` is explicitly skipped in generated tester semantic
  fixtures because Python's own `target=test` generator raises a Z3 sort
  mismatch before emitting C++ for that fixture.

The broader full-file token oracle is available as an opt-in diagnostic:

```sh
env XTRACE_OFF=1 ORACLE_TEST=1 ORACLE_TEST_TARGET_TEST=1 go test ./ivy2cpp -run TestOracleSingleTargetTest -v -count=1
```
