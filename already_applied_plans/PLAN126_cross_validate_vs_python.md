# Detach old parser: cross-validate LALR against Python Ivy

Created: 2026-03-28

## Context

The old hand-rolled recursive descent parser (`goivy/parser/`) is only used in 2 test files for cross-validation against the LALR parser. We want to replace it with the actual Python Ivy parser as the oracle, which is both the source of truth AND better test coverage.

Current references to `goivy/parser`:
- `lalr_logicparser/lalr_parser_test.go` — `parseHW()`, `parseHWAction()`, `crossValidate()`
- `lalr_logicparser/fuzz_crossval_test.go` — `parseHWFull()`

No production code imports `goivy/parser`.

## Design

### New Python helper: `pytesthelper/ivy_expr_shape.py`

A small Python script that:
1. Accepts a formula string on stdin (or as argument)
2. Wraps it as `axiom [_crossval] FORMULA` to make it parseable as a full Ivy file
3. Parses with the real `ivy_parser.parse()`
4. Extracts the formula from the resulting AxiomDecl's LabeledFormula
5. Outputs the canonical `astShape` string matching Go's format

The `astShape` format for Python AST nodes:

```python
def ast_shape(node):
    """Produce shape string matching Go's astShape() format."""
    name = type(node).__name__
    if name == 'Symbol':
        return 'Atom(%s)' % node.rep       # normalize Symbol→Atom
    if name == 'Variable':
        return 'Var(%s)' % node.rep
    if name == 'Atom':
        if not node.args:
            return 'Atom(%s)' % node.rep
        return 'Atom(%s,[%s])' % (node.rep, ','.join(ast_shape(a) for a in node.args))
    # ... etc for And, Or, Not, Implies, Forall, etc.
```

Key: And/Or flattening must match Go's `flattenAnd`/`flattenOr` behavior.

For **action bodies**, wrap as: `action _crossval = { ACTION_BODY }` and extract the body.

The script reuses the Z3 mocking pattern from the existing `ivy_ast_dump.py`.

### Go test changes

**`lalr_logicparser/lalr_parser_test.go`:**

1. Remove `import "github.com/glycerine/goivy/parser"`
2. Delete `parseHW()` and `parseHWAction()`
3. Add `parsePython(input string) (string, error)` — invokes `ivy_expr_shape.py` with the input, returns the shape string
4. Rewrite `crossValidate()`:
   ```go
   func crossValidate(t *testing.T, input string, version lexer.Version) {
       lalr, lalrErr := Parse(input, version)
       pyShape, pyErr := parsePython(input)
       // compare astShape(lalr) vs pyShape
   }
   ```
5. Rewrite `crossValidateAction()` similarly — pass input to Python as action body
6. Add `TestMain` or init to check `pythonAvailable()` and skip if not

**`lalr_logicparser/fuzz_crossval_test.go`:**

1. Remove `import "github.com/glycerine/goivy/parser"`
2. Delete `parseHWFull()`
3. Rewrite cross-validation to use `parsePython()` instead
4. For fuzz tests, Python invocation per-case may be slow — consider batch mode or accept slower fuzz runs

### Performance: Long-running subprocess + batching

The old hand-written parser was in-process (fast). Python invocation via `exec.Command` adds ~100ms startup per call. We use two strategies to stay fast:

**1. Long-running Python subprocess** — Start Python once per test binary, communicate via stdin/stdout pipes. The Python script runs a read-eval-print loop:

- Go writes a formula line to stdin
- Python parses it, writes shape line to stdout
- Go reads the shape line
- Repeat until Go closes stdin

Protocol:
```
Go → Python:  EXPR formula_text\n       (parse as expression)
Go → Python:  ACTION action_body\n      (parse as action body)
Python → Go:  OK shape_string\n         (success)
Python → Go:  ERR error_message\n       (parse error)
```

The subprocess is started lazily on first use and kept alive for the duration of the test binary via `TestMain` or `sync.Once`.

**2. Batch mode** — For random/fuzz tests that generate many formulas upfront, collect them all and send as a batch:

```
Go → Python:  BATCH\n
Go → Python:  EXPR formula1\n
Go → Python:  EXPR formula2\n
Go → Python:  ...
Go → Python:  END\n
Python → Go:  OK shape1\n
Python → Go:  OK shape2\n
Python → Go:  ...
```

Both modes use the same Python script and protocol — batch mode is just a convenience for tests that generate all inputs before comparing.

**Go helper**: A `PythonOracle` struct in a test helper file manages the subprocess lifecycle, provides `ParseExpr(input) (string, error)` and `ParseAction(input) (string, error)` methods, and handles cleanup via `Close()`.

## Files to create/modify

1. **Create `pytesthelper/ivy_expr_shape.py`** — Python script that parses expressions and outputs astShape-compatible canonical forms
2. **Update `pytesthelper/embed.go`** — Add `//go:embed ivy_expr_shape.py`
3. **Modify `lalr_logicparser/lalr_parser_test.go`** — Remove old parser import, add Python invocation, rewrite crossValidate
4. **Modify `lalr_logicparser/fuzz_crossval_test.go`** — Remove old parser import, rewrite to use Python oracle

## Verification

1. `go build ./...` — compiles without `goivy/parser` import
2. `cd lalr_logicparser && go test -v -run TestCrossValidation` — cross-validation passes against Python
3. `cd lalr_logicparser && go test -v -run TestRandom` — random formula tests pass
4. `grep -r 'goivy/parser' --include='*.go'` outside `parser/` — zero hits
5. `make golden` — unaffected
