# Complete Parser Gap Analysis: Python ivy_parser.py → Go parser/

**Created: 2026-03-22**

## Context

The Go parser must faithfully port ALL grammar rules from the Python Ivy parser. This document catalogs every missing grammar rule by systematically cross-referencing all ~240 Python grammar productions against the Go parser's capabilities.

## Methodology

1. Extracted all grammar rules from `ivy_parser.py` (~200 rules) and `ivy_logic_parser.py` (~40 rules)
2. Cataloged all Go parser functions and token dispatch in `parser/decl.go`, `parser/action.go`, `parser/expr.go`, `parser/parser.go`
3. Found all lexer keywords NOT referenced by any parser file
4. Cross-referenced every Python rule against Go capabilities

---

## CATEGORY 1: Missing Top-Level Declaration Rules

### 1.1 `module isolate` and `module object` (CRITICAL)
**Python** (line 573): `top : top MODULE modulestart modcat atom optwith EQ LCB top RCB moduleend`
- `modcat` can be empty, `OBJECT`, or `ISOLATE`
- When `modcat == "isolate"`: automatically creates inner `IsolateDecl` with `iso` label and `("common",)` attribute
- `optwith` allows `with callatom, callatom, ...` clause

**Go** (`parseModuleDecl`, line 608): Only handles `module name(params) = { body }` — no `modcat`, no `optwith`.

**Fix**: Add `modcat` lookahead after `MODULE` keyword. If next token is `OBJECT` or `ISOLATE`, consume it, parse name+optwith, then handle the isolate-specific IsolateDecl injection.

**Used in**: `ord_live.ivy` lines 1757, 2968

### 1.2 `subclass` Declaration (MISSING)
**Python** (line 649): `top : top SUBCLASS objsym OF atype EQ LCB optdotdotdot top RCB objectend`

**Go**: `SUBCLASS` token exists in lexer but NO parser rule handles it.

**Fix**: Add `case lexer.SUBCLASS:` to `parseTopLevel()` dispatch, implement `parseSubclassDecl()` following the object decl pattern with an additional `OF atype` clause.

### 1.3 `field` Declaration (MISSING)
**Python** (line 906): `symdecl : FIELD tterms`
- Treated same as `DESTRUCTOR tterms` — creates DestructorDecl nodes

**Go**: `FIELD` token exists but NO parser rule handles it.

**Fix**: Add `case lexer.FIELD:` to `parseTopLevel()` dispatch, delegate to existing `parseDestructorDecl()` (since Python treats them identically).

### 1.4 `concept` Declaration (MISSING)
**Python** (line 1463): `top : top CONCEPT cdefns`
- Where `cdefns` is concept definition list

**Go**: `CONCEPT` token NOT handled.

**Fix**: Add `case lexer.CONCEPT:` and implement `parseConceptDecl()`.

### 1.5 `state` Declaration (MISSING)
**Python** (line 2056): `top : top STATE SYMBOL EQ state_expr`
- Legacy construct for state expressions

**Go**: `STATE` token NOT handled.

**Fix**: Add `case lexer.STATE:` and implement `parseStateDecl()`.

### 1.6 `update ... from ... ` Declaration (MISSING)
**Python** (line 1478): `top : top UPDATE apps FROM apps upaxes`
- Where `upaxes` includes `params`, `requires`, `ensures`, `modifies`

**Go**: `UPDATE`, `FROM`, `MODIFIES`, `REQUIRES`, `ENSURES`, `PARAMS` tokens all NOT handled.

**Fix**: Add `case lexer.UPDATE:` and implement `parseUpdateDecl()` with full `upaxes` sub-grammar.

### 1.7 `method` as Action Variant (MISSING)
**Python** (line 1767): `actmeth : METHOD` — allows `method` as alternative keyword for `action`
**Python** (line 2663): `callatom : METHOD` — allows `method` in call position

**Go**: `METHOD` token NOT handled in parser.

**Fix**: Add `case lexer.METHOD:` to `parseTopLevel()` dispatch, delegate to `parseActionDecl()`.

### 1.8 `init` with `labeledfmla` (v1.0-1.6, PARTIAL)
**Python** (line 1468): `top : top INIT labeledfmla` — old-style init as formula
- Go only handles `init { body }` not `init fmla`

**Fix**: In `parseInitDecl()`, check if next token is `LCB`; if not, parse as labeled formula.

### 1.9 `definition` with `optexplicit` (v1.7+, PARTIAL)
**Python** (line 1429): `top : top optexplicit DEFINITION optlabel gdefn optproof`
- The `explicit` modifier is handled by our recent fix (delegates to `parseDefinitionDecl`), but `optlabel` before the definition body and `optproof` after need verification.

**Status**: Partially handled. Verify `parseDefinitionDecl` handles `[label] defn proof { ... }` fully.

---

## CATEGORY 2: Missing Action Statement Rules

### 2.1 `set` Action (MISSING)
**Python** (line 2451): `simpleact : SET lit`
- Sets a literal (for constraint languages)

**Go**: Not handled.

**Fix**: Add `case lexer.SET:` to `parseStatement()`.

### 2.2 `null` Assignment (MISSING)
**Python** (line 2632): `simpleact : term DOT SYMBOL ASSIGN NULL`
- Assigns null to a field

**Go**: Not handled.

**Fix**: Handle `NULL` on the RHS of assignments in `parseExprStatement()`.

### 2.3 `ensures` as Action Statement (MISSING)
**Python** (line 2387): `simpleact : ENSURES labeledfmla`
- `ensures` as a standalone action statement (different from `ensure`)

**Go**: `ENSURES` token NOT used in parser. Note: `ensure` IS handled — this is the older `ensures` keyword.

**Fix**: Add `case lexer.ENSURES:` to `parseStatement()`, delegate to `parseEnsureAction()`.

### 2.4 `requires` as Action Statement (MISSING)
**Python** (line 1693 context): `requires : REQUIRES fmla` (in update pattern context)

**Go**: `REQUIRES` token NOT used. Similar to `ensures` above.

**Fix**: Add `case lexer.REQUIRES:` handling.

---

## CATEGORY 3: Missing Expression/Formula Rules

### 3.1 `process` keyword (MISSING)
**Python**: No direct grammar rule found for `process` keyword.

**Status**: Likely a v1.8+ keyword reserved but not yet used in grammar. Can defer.

### 3.2 `template` keyword (MISSING)
**Python**: No direct grammar rule found.

**Status**: Likely a v1.8+ keyword reserved but not yet used. Can defer.

### 3.3 `fresh` Modifier in Schema Declarations (MISSING)
**Python** (lines 685, 693, 701):
```
schdecl : FRESH FUNCTION funs
schdecl : FRESH INDIV funs
schdecl : FRESH RELATION rels
```

**Go**: `parseSchemaBody()` handles `FUNCTION`, `INDIV`, `RELATION` in schema bodies, but does NOT check for `FRESH` modifier before them.

**Fix**: In `parseSchemaBody()`, add `FRESH` lookahead before `FUNCTION`/`INDIV`/`RELATION`.

### 3.4 `entry` State Expression (MISSING)
**Python** (line 3055): `state_expr : ENTRY`

**Go**: `ENTRY` token NOT handled.

**Fix**: Part of state expression grammar (see 1.5 above).

### 3.5 Concept Space Expressions (MISSING)
**Python** (lines 2962-3024): Complex concept space grammar with `expr`, `exprterm`, `prod`, `sum` rules.

**Go**: Not implemented. The concept space grammar is used for counterexample-guided abstraction refinement (CEGAR).

**Fix**: Implement concept space expression grammar as a separate parser function.

---

## CATEGORY 4: Incomplete Existing Rules

### 4.1 `axiom` Without `optexplicit opttemporal` Handling (v1.7+)
**Python** (line 466): `top : top optexplicit opttemporal AXIOM lgprop`

**Go**: `parseAxiomDeclMulti()` parses `axiom [label] fmla [proof]` but does NOT handle `explicit` or `temporal` modifiers on axioms when they appear via the AXIOM token path (only handles them through the EXPLICIT/TEMPORAL dispatch).

**Status**: The Go parser handles `explicit axiom` and `temporal axiom` through the `parseExplicitDeclMulti` and `parseTemporalDeclMulti` paths. Verify these paths are complete.

### 4.2 `object` with `...` Continuation (PARTIAL)
**Python** (line 633): `top : top OBJECT objsym objectargs EQ LCB optdotdotdot top RCB objectend`
- `optdotdotdot` allows `...` at start of object body for continuation

**Go**: `parseObjectDeclMulti()` exists. Verify it handles `...` continuation.

### 4.3 `module` with `optwith` (MISSING)
**Python** (line 573): `atom optwith EQ` — the module can have a `with callatoms` clause before `=`.

**Go**: `parseModuleDecl()` does NOT handle `with` clause.

**Fix**: Add `with` parsing before `=` in `parseModuleDecl()`.

### 4.4 `callatom : METHOD` (MISSING)
**Python** (line 2663): `callatom : METHOD` — `method` as a call atom

**Go**: Not handled.

**Fix**: In `parseCallatom()`, add `METHOD` as valid start token.

---

## CATEGORY 5: Already Handled (Verified)

These were potential gaps but are actually correctly implemented:

- ✅ `explicit temporal property` — Fixed in this session
- ✅ `explicit temporal axiom` — Fixed in this session
- ✅ `explicit property/axiom/invariant` — Fixed in this session
- ✅ `temporal property/axiom` — Fixed in this session
- ✅ Schema bodies in axioms (`lgprop` → `parseLabeledFmla` handles `LCB` → `parseSchemaBody()`)
- ✅ `autoinstance` declarations
- ✅ `global`, `common`, `specification`, `implementation` blocks
- ✅ `proof { tactic ... with { definition ...; invariant ... } }` — Both braced and unbraced
- ✅ `showgoals` in proof blocks
- ✅ `if some ... minimizing` expressions
- ✅ `isolate X = { body } with callatoms`
- ✅ `trusted isolate`
- ✅ `extract` declarations
- ✅ All arithmetic operators (+, -, *, /)
- ✅ Ternary `X if cond else Y`
- ✅ `globally`, `eventually` temporal operators
- ✅ `forall`, `exists` quantifiers
- ✅ Named binders ($)
- ✅ Sort annotations (expr : type)
- ✅ `for` loops, `while` loops, `if` statements
- ✅ `debug` statements
- ✅ `var` declarations in actions
- ✅ `local`, `let` actions
- ✅ `call` actions
- ✅ All mixin forms (before/after/around/implement)
- ✅ Schema declarations and theorem declarations
- ✅ Attribute declarations
- ✅ `private { ... }` blocks
- ✅ `nativequote` (<<<...>>>)
- ✅ `variant ... of`
- ✅ `interpret`
- ✅ `delegate`
- ✅ `progress`, `rely`, `mixord`
- ✅ `destructor`, `constructor`

---

## Implementation Priority

### P0 — Blocks Real Files (Fix First)
1. **1.1**: `module isolate` / `module object` / `module ... with` — Blocks `ord_live.ivy`
2. **1.3**: `field` declaration — Used in `include/1.8/numbers.ivy` and struct types
3. **3.3**: `fresh` modifier in schemas — Used in theorem/schema proofs

### P1 — Complete Coverage (Fix Second)
4. **1.2**: `subclass` declaration
5. **1.7**: `method` as action keyword
6. **2.1**: `set` action
7. **2.2**: `null` assignment
8. **2.3**: `ensures` action statement
9. **2.4**: `requires` action statement

### P2 — Legacy/Advanced (Fix Third)
10. **1.4**: `concept` declarations (CEGAR)
11. **1.5**: `state` declarations (legacy)
12. **1.6**: `update ... from ...` patterns (update axioms)
13. **3.4**: `entry` state expression (legacy)
14. **3.5**: Concept space expressions (CEGAR)
15. **1.8**: `init` with formula (v1.0-1.6 legacy)
16. **4.4**: `callatom : METHOD`

---

## Files to Modify

| File | Changes |
|------|---------|
| `parser/decl.go` | Add `SUBCLASS`, `FIELD`, `CONCEPT`, `STATE`, `UPDATE`, `METHOD` to `parseTopLevel()` dispatch; extend `parseModuleDecl()` for `modcat` and `optwith`; add `parseSubclassDecl()`, `parseConceptDecl()`, `parseStateDecl()`, `parseUpdateDecl()`; add `FRESH` handling in `parseSchemaBody()` |
| `parser/action.go` | Add `SET`, `ENSURES`, `REQUIRES` to `parseStatement()` dispatch; handle `NULL` in RHS |
| `parser/expr.go` | Add `METHOD` to `parseCallatom()` |
| `parser/parser.go` | Concept space expression grammar (if implementing P2) |

## Verification

1. `go build ./parser/` compiles
2. `go test ./parser/...` passes (no regressions)
3. `./goivy_check ord_live.ivy` parses past all currently-failing lines
4. Grep all `.ivy` files in `ivy-lang-examples/` for each fixed keyword, verify they parse
5. Run `go test ./...` for full test suite
