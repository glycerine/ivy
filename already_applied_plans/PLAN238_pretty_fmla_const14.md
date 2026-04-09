# Port Python's `cls.__str__ = pretty_fmla` monkey-patch to Go for all 14 logic types

Created: 2026-04-09 (afternoon session)

## Context

Cross-language log comparison (`TestOrdLive`) fails at line 236447 of
`~/ivy/goivy/log.red`:

```
go : XTRACE: ivy_solver.py:417 numeral_to_z3() ENTER num=0
py : XTRACE: ivy_solver.py:417 numeral_to_z3() ENTER num=0:index
```

The Python side prints the numeral with its sort annotation (`0:index`); Go
prints only the name (`0`). This is one symptom of a broader, structural
mismatch: Python `ivy_logic.py` monkey-patches `__str__` over **14 logic
types** so that `str(x)` for any of them returns the result of `pretty_fmla(x)`
(which is `drop_annotations(False, set()).ugly(0)`). Go's `String()` methods
for those same 14 types currently return ad-hoc Go-style debug formats and
do not match Python.

This plan ports Python's monkey-patch to Go uniformly: every one of the 14
types gets `String() = PrettyFmla(self)`, mirroring Python verbatim.

## Source-of-truth Python

`~/ivy/pyivy/ivy/ivy/ivy_logic.py:1440-1446`:

```python
def pretty_fmla(self):
    d = self.drop_annotations(False,set())
    return d.ugly(0)

for cls in [lg.Eq, lg.Not, lg.And, lg.Or, lg.Implies, lg.Iff, lg.Ite,
            lg.ForAll, lg.Exists, lg.Apply, lg.Var, lg.Const, lg.Lambda,
            lg.NamedBinder]:
    cls.__str__ = pretty_fmla
```

The 14 types: **Eq, Not, And, Or, Implies, Iff, Ite, ForAll, Exists, Apply,
Var, Const, Lambda, NamedBinder**.

`pretty_fmla` is mechanically equivalent to a recursive ugly-print rooted at
the node, with sort annotations dropped where they would be inferable. The
per-type `ugly` lambdas are defined at `ivy_logic.py:1315-1355` and the
per-type `drop_annotations` at `ivy_logic.py:1363-1437`.

For the immediate divergence (numeral Const at `numeral_to_z3` ENTER) the
relevant `ugly` is at `ivy_logic.py:1317-1319`:

```python
lg.Const.ugly = (lambda self,prec: (self.name+':'+self.sort.name)
                    if show_numeral_sorts and self.is_numeral()
                       and not isinstance(self.sort,lg.TopSort)
                 else self.name)
```

So `str(Const('0', IndexSort))` evaluates to `'0:index'` because:
1. `Const.__str__` is monkey-patched to `pretty_fmla`.
2. `pretty_fmla(c)` calls `c.drop_annotations(False, set())`, which for Const
   is a no-op (`const_drop_annotations` only acts when `inferred_sort=True`,
   per `ivy_logic.py:1373-1378`). Returns `c` unchanged.
3. Then `c.ugly(0)` returns `name+':'+sort.name` because `c.is_numeral()` is
   true and the sort is not TopSort.

## Go-side current state

The helpers needed to match Python's `pretty_fmla` already exist and are
correct mechanical ports:

- `logic/pretty.go:14-17` `PrettyFmla(n Expr) string` — calls
  `dropAnnotations(n, false, ...)` then `ugly(d, 0)`. This is the direct
  port of Python's `pretty_fmla`.
- `logic/pretty.go:39-92` `ugly(n Expr, prec int) string` dispatches on
  type and handles all 14 of the monkey-patched types (plus extras like
  Globally, Eventually, Cond, Definition that are NOT in Python's list).
- `logic/pretty.go:96-104` `varUgly`, `:108-115` `constUgly`, `:117-165`
  `appUgly`, `:168-174` `notUgly`, `:177-195` `naryUgly`, `:199-206`
  `naryParen`, `:209-219` `quantUgly` — each mechanically ports the
  matching Python `ugly` lambda.
- `logic/pretty.go:226-345` `dropAnnotations` — mechanical port of
  Python's per-type `drop_annotations` methods, including the no-op
  `case *Const` at line 238-242 and `case *Variable` at line 228-236.

So `PrettyFmla(x)` for any of the 14 types already produces the same
formatted string Python's monkey-patched `__str__` does. The remaining gap
is purely that Go's per-type `String()` methods do not call `PrettyFmla`.

### Current Go `String()` methods to be replaced

| Type        | File                  | Line | Current body (paraphrase)                              |
|-------------|-----------------------|------|--------------------------------------------------------|
| Variable    | `logic/term.go`       | 53   | `return v.Name`                                        |
| Const       | `logic/term.go`       | 102  | `return c.Name`                                        |
| Apply       | `logic/term.go`       | 212  | `fmt.Sprintf("%s(%s)", a.Func, …)` (recursive)         |
| Eq          | `logic/formula.go`    | 54   | `fmt.Sprintf("(%s == %s)", e.T1, e.T2)`                |
| Ite         | `logic/formula.go`    | 87   | `fmt.Sprintf("Ite(%s, %s, %s)", t.Cond, t.Then, t.Else)`|
| Not         | `logic/formula.go`    | 111  | `(t1 != t2)` for Not(Eq), else `Not(body)`             |
| And         | `logic/formula.go`    | 270  | `fmt.Sprintf("And(%s)", nodeSliceStr(a.Terms))`        |
| Or          | `logic/formula.go`    | 301  | `fmt.Sprintf("Or(%s)", nodeSliceStr(o.Terms))`         |
| Implies     | `logic/formula.go`    | 327  | `fmt.Sprintf("Implies(%s, %s)", i.T1, i.T2)`           |
| Iff         | `logic/formula.go`    | 351  | `fmt.Sprintf("Iff(%s, %s)", i.T1, i.T2)`               |
| ForAll      | `logic/formula.go`    | 387  | `fmt.Sprintf("(ForAll %s. %s)", varSortList(...), …)`  |
| Exists      | `logic/formula.go`    | 424  | `fmt.Sprintf("(Exists %s. %s)", varSortList(...), …)`  |
| Lambda      | `logic/formula.go`    | 455  | `fmt.Sprintf("(Lambda %s. %s)", varSortList(...), …)`  |
| NamedBinder | `logic/formula.go`    | 505  | `fmt.Sprintf("($%s%s %s. %s)", …)`                     |

All 14 will be replaced with `return PrettyFmla(self)`.

## Plan

### Step 1 — Replace `String()` for all 14 types

For each row in the table above, replace the body with
`return PrettyFmla(<receiver>)` and add a one-line doc comment that points
at `ivy_logic.py:1444-1446`. Example for Const:

```go
// String matches Python ivy_logic.py:1444-1446 which monkey-patches
// lg.Const.__str__ = pretty_fmla. PrettyFmla calls
// drop_annotations(False, set()).ugly(0); for a Const this returns
// "name:sortName" for numerals with non-TopSort, else just "name".
func (c *Const) String() string { return PrettyFmla(c) }
```

The same one-line replacement applies to all 14, with the receiver and
type name varying. The existing helper bodies (`varUgly`, `constUgly`,
`appUgly`, `naryUgly`, `naryParen`, `notUgly`, `quantUgly`) already encode
the correct per-type formatting, so no other logic changes are needed.

**Recursion safety:** `PrettyFmla(x)` → `dropAnnotations(x, false, ...)`
→ `ugly(d, 0)` → per-type helper → recursive `ugly(...)` calls on children.
None of these helpers call `.String()` — they go through `ugly` directly,
or read `.Name`/`.CSort` fields. The only `fmt.Sprint` fallback is in
`appUgly` at `pretty.go:127` for an unrecognized `a.Func` type, which is
not on a hot path and is recursion-bounded. No infinite recursion.

### Step 2 — Update `logic/formula_test.go` golden assertions

`logic/formula_test.go` currently asserts on the old Go-style format. After
Step 1, those assertions will fail with the new Python-style output. Each
needs the expected string updated. The affected lines (and what they
become — final values to be confirmed by running the test once):

| Line | Old expected                                   | New expected (Python style)                         |
|------|------------------------------------------------|-----------------------------------------------------|
| 16   | `"(X == Y)"`                                   | `"X = Y"`                                           |
| 66   | `"(X != Y)"`                                   | `"X ~= Y"`                                          |
| 72   | `"Not(And((X == Y)))"`                         | `"~ (X = Y)"` (note: trailing space matches `nary_paren`) |
| 98   | `"And(leq(X, Y), leq(Y, X))"`                  | `"(leq(X, Y) & leq(Y, X))"`                         |
| 106  | `"Or(leq(X, Y), leq(Y, X))"`                   | `"leq(X, Y) | leq(Y, X)"` (or with parens, depends on prec) |
| 122  | `"And()"`                                      | `"true"`                                            |
| 125  | `"Or()"`                                       | `"false"`                                           |
| 142  | `"Implies(leq(X, Y), leq(Y, X))"`              | `"leq(X, Y) -> leq(Y, X)"`                          |
| 150  | `"Iff(leq(X, Y), leq(Y, X))"`                  | `"leq(X, Y) <-> leq(Y, X)"`                         |
| 224  | `"Ite((X == Y), X, Y)"`                        | `"(X if X = Y else Y)"`                             |

The `t.Log(...)` lines for ForAll/Exists/Lambda (171, 201, 212) are
informational only — no change needed beyond the format being different.

The Globally / Eventually / WhenOperator / Cond / Definition tests at
lines 251, 260, 273, 289 are **not affected** — those types are not in
Python's monkey-patch list, so their Go `String()` stays as-is.

> **Note:** The exact "new expected" strings above are best-effort
> derivations from reading the helper code. Run the test once after Step 1
> and copy the actual output (it will be the correct Python-mirror output).
> Do not hand-tune these — match what `PrettyFmla` actually emits, since
> that is by construction the same as Python.

### Step 3 — Spot-check other test packages for golden-string sensitivity

The packages below contain string literals that mention Go-style names like
`And(`, `Or(`, etc., but most are likely constructor invocations (e.g.,
`NewAnd(...)`) rather than golden assertions. Skim each for any
`.String() ==` or `t.Errorf` comparing against the old Go-style format and
update as needed:

- `z3bridge/solver_test.go`
- `z3bridge/z3bridge_test.go`
- `z3bridge/translate2_test.go`
- `z3bridge/z3_utils_test.go`
- `z3bridge/solver2_test.go`
- `logicutil/logicutil_test.go`
- `logicparser/logicparser_test.go`
- `lalr_logicparser/lalr_parser_test.go`

Update only assertions that compare a `*logic.<14-type>.String()` against a
literal in old Go-style format. Leave constructor calls, sexp literals, and
parser-input strings untouched.

### Step 4 — What NOT to change (scope discipline)

- `Const.Repr()` (term.go:104-105), `Variable.Repr()` (term.go:58-63),
  `Apply.Repr()` (term.go:226-235): these are on a different code path
  (`ReprExpr`/`ReprNode`) and correspond to Python `__repr__` (not
  `__str__`). Out of scope.
- The 5 types **not** in Python's monkey-patch list — Globally, Eventually,
  WhenOperator, Cond, Definition — keep their existing Go-style `String()`.
- No Config struct changes are needed: `show_variable_sorts` and
  `show_numeral_sorts` are package-level constants in Python set to True
  and never flipped on this path. The Go pretty.go helpers already behave
  as if they are True (which is correct).
- Do not change any of the existing helpers in `logic/pretty.go` — they
  are already correct mechanical ports.
- Do not consolidate or remove any `String()` methods. Each of the 14
  keeps its own one-line method that calls `PrettyFmla(self)`.

## Files to modify

1. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/term.go`
   — replace `String()` bodies for `Variable` (line 53), `Const` (line 102),
   `Apply` (line 212).

2. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/formula.go`
   — replace `String()` bodies for `Eq` (54), `Ite` (87), `Not` (111),
   `And` (270), `Or` (301), `Implies` (327), `Iff` (351), `ForAll` (387),
   `Exists` (424), `Lambda` (455), `NamedBinder` (505).

3. `/Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/formula_test.go`
   — update golden expected strings as listed in Step 2.

4. (Possibly) other `_test.go` files listed in Step 3, only if a hard-coded
   golden assertion compares against the old format.

## Verification

1. Build the test binary the golden harness depends on:
   ```
   cd /Users/jaten/ivy/goivy/cmd/goivy_check && go build -o /Users/jaten/go/bin/goivy_check_xtrace
   ```
   The harness rebuilds this automatically; mentioned for manual repro.

2. Run the unit tests in the logic package first (smallest blast radius):
   ```
   cd /Users/jaten/ivy/goivy/logic && go test ./...
   ```
   Update any golden strings in `formula_test.go` to the actual output.

3. Run the full test suites that touch logic types:
   ```
   cd /Users/jaten/ivy/goivy && go test ./z3bridge/... ./logicutil/... ./logicparser/... ./lalr_logicparser/...
   ```
   Fix any other golden assertions surfaced.

4. Run the failing golden-path test that prompted this fix:
   ```
   cd /Users/jaten/ivy/goivy/parser && go test -run TestOrdLive -v -timeout 30m
   ```
   Expected: the divergence at line 236447 (`num=0` vs `num=0:index`) is
   resolved. The test should advance past 236447. Inspect the new
   `~/ivy/goivy/log.red` to confirm both sides print `num=0:index` for the
   numeral_to_z3 ENTER trace.

5. If a new divergence appears at some later trace line, that becomes the
   next mechanical-port item — out of scope for this plan.
