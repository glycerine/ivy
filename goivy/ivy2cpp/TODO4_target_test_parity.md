# ivy2cpp TODO4: full target=test token parity

Created: 2026-05-23

Status: completed 2026-05-23.

The emitted tester loop, Z3 scaffold, REPL/test main scaffold, and generated
tester semantic fixtures now run through the oracle harness with Python
`ivy_to_cpp.py` token parity for the supported fixtures.

- `variant_simple.ivy` is explicitly skipped in generated tester semantic
  fixtures because Python's own `target=test` generator raises a Z3 sort
  mismatch before emitting C++ for that fixture.

The broader full-file token oracle is available as an opt-in diagnostic and
passes with 13 fixtures plus the documented Python-side tester skip:

```sh
env XTRACE_OFF=1 ORACLE_TEST=1 ORACLE_TEST_TARGET_TEST=1 go test ./ivy2cpp -run TestOracleSingleTargetTest -v -count=1
```
