# Deep Dive: Ivy 1.8 Support in goivy

**Created: 2026-03-22**

## Context

The Go port (goivy) declares `IvyLatestLanguageVersion = "1.7"` while include/1.8/ library files exist in the codebase. We need to determine whether the Go port truly supports Ivy 1.8, and if not, plan the work to make it real.

---

## Analysis: What Does Python Ivy 1.8 Actually Change?

I exhaustively grep'd every `version_le`, `get_numeric_version`, and `get_string_version` call across all Python `.py` files. Here is every version-gated code path in the Python Ivy codebase and whether it distinguishes 1.8 from 1.7:

### Version Gates That Treat 1.7 and 1.8 Identically

**ALL of these checks are binary "≤ 1.6 vs > 1.6" or "≥ 1.7 vs < 1.7" — they fire the same way for both 1.7 and 1.8:**

| File | Check | Meaning |
|------|-------|---------|
| `ivy_utils.py:574` | `get_numeric_version() <= [1,1]` | compose character |
| `ivy_utils.py:576` | `get_numeric_version() <= [1,2]` | polymorphism flag |
| `ivy_utils.py:577` | `get_numeric_version() <= [1,5]` | polymorphic macros |
| `ivy_utils.py:578` | `get_numeric_version() <= [1,6]` | ghost init forbid |
| `ivy_lexer.py:263-298` | 7 version gates from ≤[1,0] to ≤[1,7] | keyword availability |
| `ivy_parser.py` | ~30 checks, ALL ≤[1,0] through ≤[1,6] | grammar rule gating |
| `ivy_logic_parser.py` | ~12 checks, ALL ≤[1,0] through ≤[1,6] | logic grammar |
| `ivy_logic.py:323,1121` | `≤ [1,2]` | polymorphism |
| `ivy_ast.py:674,1110` | `≤ [1,6]` | ensures semantics |
| `ivy_actions.py:376` | `≤ [1,6]` | ensures behavior |
| `ivy_compiler.py:1738` | `version_le("1.7", version)` means ≥1.7 | action interference check |
| `ivy_compiler.py:2090,2221` | `≤ "1.6"` | isolate creation |
| `ivy_isolate.py` (10 checks) | ALL either `≤ "1.6"` or `≥ "1.7"` | isolate compilation |
| `ivy_theory.py:97` | `version_le("1.7", version)` means ≥1.7 | theory schema selection |
| `ivy_to_cpp.py:3546,5800,5838,5858,5982` | ALL `≤ "1.6"` | C++ code gen |

### The ONE Check That Distinguishes 1.8 from 1.7

```python
# ivy_to_cpp.py:5866-5868
iso.compile_with_invariants.set("true" if target.get()=='test'
                                and not iu.version_le(iu.get_string_version(),"1.7")
                                else "false")
```

This sets `compile_with_invariants = true` ONLY for version > 1.7 (i.e., 1.8+) when the compilation target is `test`. This affects `ivy_isolate.py:1097`:
```python
if not create_imports.get() or compile_with_invariants.get():
    # don't strip conjectures during isolate compilation
```

**This is a C++ code generation feature** — when compiling for test targets in 1.8, conjectures (invariants) are preserved in the compiled output rather than being stripped. This is NOT relevant to goivy's verification focus (goivy doesn't do C++ code generation).

### Lexer Keywords for Version > 1.7

When version > 1.7, these keywords become available (were gated behind `≤ [1,7]`):
- `global`, `common`, `debug`, `field`, `for`, `process`, `subclass`, `template`, `whenfirst`, `whenlast`, `whennext`, `whenprev`, `unprovable`, `trigger`

And `requires`/`ensures` are REMOVED (replaced by `require`/`ensure` which were added in 1.7).

### Standard Library Changes (include/1.8/ vs include/1.7/)

One NEW file: `numbers.ivy` — provides bit vectors, natural numbers, exponentiation, numeric iterators.

15 MODIFIED files with changes like:
- `#lang ivy1.7` → `#lang ivy` (no version number — inherits from parent)
- Heavy use of `global { ... }` blocks
- `autoinstance` declarations
- `class` definitions with `field` members
- Modernized C++ threading (for `ivy_to_cpp` targets, not verification)

---

## Status of Go Port Support

### What WORKS Correctly for 1.8

| Feature | Go Implementation | Status |
|---------|-------------------|--------|
| Version parsing `#lang ivy1.8` | `ivyinit.go:85-97` — parses any version string | ✅ Works |
| `#lang ivy` (no version) | `ivyinit.go:87` — keeps current version if empty | ✅ Works |
| SetStringVersion flags | `names.go:163-180` — all flags set correctly for 1.8 | ✅ Works |
| Keyword gating | `lexer.go:204-214` — `vle(1,7)` is false for 1.8, enabling correct keywords | ✅ Works |
| Include directory resolution | `names.go:233-274` — finds smallest dir ≥ version | ✅ Works |
| Parser routing | `lalr_parser.go:15-24` — routes to v17 parser for ≥1.7 | ✅ Works |
| Theory schemas | `theory.go:111-188` — uses TheorySchemas17 for ≥1.7 | ✅ Works |
| `global` keyword + parser rule | `parser/decl.go:177-178` — `parseGlobalDecl` exists | ✅ Works |
| `autoinstance` keyword + parser | `parser/decl.go:167-168` — `parseAutoInstanceDecl` exists | ✅ Works |
| `class` keyword + parser | `parser/decl.go:87-88` — handled as object variant | ✅ Works |
| `common` keyword + parser | `parser/decl.go:171-172` — `parseCommonBlock` exists | ✅ Works |
| `field` keyword in struct parsing | `parser/parser.go:674-679` — struct fields parsed | ✅ Works |
| `for` loops | `parser/action.go:94` — parsed | ✅ Works |
| `debug` statements | `parser/action.go:106` — parsed | ✅ Works |
| `unprovable` invariants | `parser/decl.go:179-181` — handled | ✅ Works |
| All compiler version checks | All use `≤ 1.6` or `≥ 1.7` — same result for 1.7 and 1.8 | ✅ Works |
| All isolate version checks | Same pattern | ✅ Works |

### What's WRONG

| Issue | Impact | Fix |
|-------|--------|-----|
| `IvyLatestLanguageVersion = "1.7"` | Default version is 1.7, not 1.8. Files using `#lang ivy1.8` work fine since they explicitly set the version. But the declared "latest" is wrong. | Change to `"1.8"` |
| `ivyLanguageVersion = "1.7"` default | Default language version should be 1.8 to match Python's behavior if we update the latest. | Change to `"1.8"` |

### What's NOT Relevant

The `compile_with_invariants` feature (the ONLY code-level 1.8 vs 1.7 distinction) is in `ivy_to_cpp.py` — C++ code generation. goivy does not port `ivy_to_cpp`, so this does not apply.

---

## Plan

### Step 1: Update Version Constants

**File: `/Users/jaten/go/src/github.com/glycerine/goivy/ivyutils/names.go`**

```go
// Line 128: Change default version
var ivyLanguageVersion = "1.8"   // was "1.7"

// Line 147: Change latest version
var IvyLatestLanguageVersion = "1.8"  // was "1.7"
```

Also update the default initial values of the version-dependent flags to match what `SetStringVersion("1.8")` would produce:
```go
// Line 133: Already true (correct for > 1.2) ✅
var IvyHavePolymorphism = true

// Line 138: Should be true (correct for > 1.5), currently false ❌
var IvyUsePolymorphicMacros = true   // was false

// Line 143: Should be true (correct for > 1.6), currently false ❌
var IvyForbidGhostInit = true   // was false
```

**Rationale for flag changes**: The defaults should match what `SetStringVersion(ivyLanguageVersion)` would produce. If the default version is "1.8", then `IvyUsePolymorphicMacros` (> 1.5) and `IvyForbidGhostInit` (> 1.6) should both be `true`. Currently they're `false`, which means code that runs before `SetStringVersion` is called would see wrong defaults. This matches Python where the defaults are set inline at module load time:
```python
ivy_language_version = '1.7'  # but then set_string_version is called
ivy_use_polymorphic_macros = False  # Python also has this wrong until set_string_version runs
```

Actually wait — Python has the same issue. Python's defaults are also `False` for these flags, and they only get set to `True` when `set_string_version()` is called. So to stay faithful to the Python port, we should ONLY change the version strings, not the default flag values. The flags get set correctly when `SetStringVersion` is called during file parsing.

**Revised Step 1**: Only change the two version strings:
- `ivyLanguageVersion`: `"1.7"` → `"1.8"`
- `IvyLatestLanguageVersion`: `"1.7"` → `"1.8"`

### Step 2: Verify `stdIncludeDir` Cache Reset

**File: `/Users/jaten/go/src/github.com/glycerine/goivy/ivyutils/names.go`**

The `stdIncludeDir` is cached after first resolution. If the version changes mid-session (e.g., processing a file that includes another file with a different version), the cached directory won't update. The Python code has the same behavior (it also caches). This is pre-existing and not specific to 1.8 — no change needed.

### Step 3: Verify Go Parser Handles All 1.8 Constructs

Already verified above — the Go parser has rules for `global`, `autoinstance`, `class`, `common`, `field`, `for`, `debug`, `unprovable`, and `trigger`. All the version > 1.7 keywords have corresponding parser implementations.

**No parser changes needed.**

### Step 4: Update Tests

If there are any tests that assert `IvyLatestLanguageVersion == "1.7"`, update them.

**File to check:** `/Users/jaten/go/src/github.com/glycerine/goivy/theory/theory_test.go` — has `VersionLE("1.8", "1.7")` test, which should still pass (1.8 is NOT ≤ 1.7).

### Step 5: Smoke Test

Run the goivy parser/compiler on a `#lang ivy1.8` file that uses 1.8 constructs to verify end-to-end:
1. Parse a simple file with `#lang ivy1.8`, `global { ... }`, `autoinstance`
2. Verify it parses and compiles without error
3. Try parsing the include/1.8/numbers.ivy file

---

## Verification

1. `go test ./...` — all existing tests pass
2. Parse `ivy-lang-examples/ivy/include/1.8/numbers.ivy` with version 1.8
3. Parse a test file from `ivy-lang-examples/` that uses `#lang ivy1.8`
4. Confirm `GetStringVersion()` returns `"1.8"` by default

---

## Summary

**The Go port already has nearly complete functional support for Ivy 1.8.** The parser handles all 1.8 keywords, the version-gated code paths all treat 1.7 and 1.8 identically (except one C++ codegen feature we don't port), and the include directory resolution works. The only real fix needed is updating two version string constants from `"1.7"` to `"1.8"`.
