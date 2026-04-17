# PLAN: Fix Python Ivy backward compatibility for `#lang ivy1.1`

Created: 2026-04-17, 06:15

## Context

Python Ivy (`ivy_check`) crashes with `YaccError: Unable to build parser` when opening any `#lang ivy1.1` file (e.g., `tilelink1.ivy`). The parser was built for version 1.7 at import time. When `read_module` detects `#lang ivy1.1`, it calls `set_string_version('1.1')`, clears all grammar rules, and reloads `ivy_parser`. During the reload, PLY yacc rejects the grammar because of an undefined nonterminal.

## Root cause

**`ivy_parser.py:676`** — the module definition grammar rule uses the `optwith` nonterminal:

```python
def p_top_module_atom_eq_lcb_top_rcb(p):
    'top : top MODULE modulestart modcat atom optwith EQ LCB top RCB moduleend'
```

This rule is **unconditional** (defined at module scope, no version guard).

But the `optwith` nonterminal itself (lines 2259-2266) is defined **inside** the `if not (iu.get_numeric_version() <= [1,1]):` block (starting at line 2158). For version 1.1, `optwith` is never defined.

PLY yacc detects this at parser-build time: `"Symbol 'optwith' used, but not defined as a token or a rule"`, sets `errors = True`, and raises `YaccError('Unable to build parser')`. The error message is silenced by `errorlog=yacc.NullLogger()`, making the failure opaque.

Verified: adding `optwith` as an unconditional definition makes the parser build succeed for version 1.1.

## Fix

Move the `p_optwith` and `p_optwith_with_callatoms` definitions **outside** the version guard (from inside `if not (iu.get_numeric_version() <= [1,1]):` to the unconditional module scope).

This is safe because:
- `optwith` has an epsilon production (`optwith :` → empty list) so it always matches
- The `WITH` keyword is removed from the lexer's reserved words for version ≤1.1 (via `LexerVersion.__enter__`, line 267-268), so `optwith : WITH callatoms` can never fire at parse time for version 1.1
- The grammar rule at line 676 already works correctly for both versions — for 1.1, `optwith` will always reduce to the empty epsilon production

### Specific change

In `/Users/jaten/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/ivy_parser.py`:

**Before** (simplified):
```python
if not (iu.get_numeric_version() <= [1,1]):      # line 2158
    ...
    def p_optwith(p):                              # line 2259
        'optwith : '
        p[0] = []
    def p_optwith_with_callatoms(p):               # line 2263
        'optwith : WITH callatoms'
        p[0] = p[2]
    ...
```

**After**: Move `p_optwith` and `p_optwith_with_callatoms` before the version guard (or after the guard at module scope). Unindent from 4-space to 0-space. The two functions become unconditional.

### Also fix: `UNPROVABLETRIGGER` token bug

`ivy_lexer.py` lines 47-48 have a missing comma:
```python
    'UNPROVABLE'
    'TRIGGER'
```
Python string concatenation produces `'UNPROVABLETRIGGER'` instead of two separate tokens. This doesn't block the fix but is a latent bug (PLY warns: "Token 'UNPROVABLETRIGGER' defined, but not used").

## Go side — no change needed

Go's `parser/grammar_v17.y` already has `optwith` defined **unconditionally** (line 3481). The Go approach is the same pattern we're applying to Python: a single grammar with lexer-level version gating. For v1.1, `lexer.go:174` removes `with` from reserved words so `TOK_WITH` is never emitted, and `optwith` always reduces to epsilon.

Go also has version-specific **logic parser** grammars (`lalr_logicparser/v12/grammar_v12.y`, `v16/grammar_v16.y`) for formulas/terms, but the main parser grammar is version-agnostic and already correct.

## Critical file

- `/Users/jaten/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/ivy_parser.py` — move `optwith` rules out of version guard

## Secondary file (optional)

- `/Users/jaten/go/src/github.com/glycerine/ivy/pyivy/ivy/ivy/ivy_lexer.py` — fix missing comma on line 47

## Verification

```bash
cd ~/ivy/ivy-lang-examples/examples/tilelink
ivy_check tilelink1.ivy
```

Should no longer crash with `YaccError`. It should proceed to actually parse and check the file (may still have other issues, but the parser build must succeed).

Also verify version 1.7 still works:
```bash
cd ~/ivy/goivy
make tlb   # uses tlb.ivy which is version 1.7
```
