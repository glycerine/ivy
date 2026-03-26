# Plan: Fix bracket-stripping in `labelname` rule (golden test divergence at 61368)

**Created:** 2026-03-25T22:15

## Context

Golden test (`make golden`) diverges at line 61368. Go emits `rep:"[cfabric_pio_fair_ax]"` while Python emits `rep:"cfabric_pio_fair_ax"`. Brackets are not being stripped from label names.

**Root cause:** Python's `p_optlabel_label` does `Atom(p[1][1:-1],[])` — stripping the brackets from the LABEL token. In Go, the `labelname` grammar rule assembles `"[" + sym + "]"` into a TokenInfo, and most consumers pass `$N.Val` raw without stripping.

## Analysis

Go has **9 rules** that consume `labelname`. Python **always** strips brackets with `[1:-1]`.

| Rule location | Currently strips? |
|---|---|
| `labeledfmla: labelname fmla` (line 2231) | YES — manual if/else |
| `TOK_THUNK labelname ...` (line 4349) | YES — `strings.Trim` |
| `optlabel: labelname` (line 2339) | **NO** |
| `top TOK_THEOREM labelname ...` (line 1006) | **NO** |
| `top TOK_PROOF labelname ...` (line 1022) | **NO** |
| `optproof: TOK_PROOF labelname ...` (line 2385) | **NO** |
| `TOK_INSTANTIATE labelname ...` (line 4948) | **NO** |
| `TOK_INSTANTIATE labelname ... WITH ...` (line 4961) | **NO** |
| `proofstep: TOK_PROOF labelname ...` (line 5063) | **NO** |

**7 of 9 consumers are missing bracket stripping.**

## Fix Strategy

Strip brackets **once** at the source — in the `labelname` rule itself. Then remove redundant stripping from the 2 rules that already do it.

### Change 1: `labelname` rule (line 2249-2260)

**Before:**
```go
labelname:
    TOK_LB SYMBOLx TOK_RB
    {
        xtracer.Trace("parser.p_LABEL_LB_SYMBOL_RB ENTER (LABEL)")
        // Python: LABEL = p[1] + p[2] + p[3] → "[sym]"
        $$ = TokenInfo{Val: "[" + $2.Val + "]", Line: $1.Line}
    }
    | TOK_LABEL
    {
        xtracer.Trace("parser.p_labelname__label ENTER (labelname)")
        $$ = $1
    }
```

**After:**
```go
labelname:
    TOK_LB SYMBOLx TOK_RB
    {
        xtracer.Trace("parser.p_LABEL_LB_SYMBOL_RB ENTER (LABEL)")
        // Python: all LABEL consumers strip brackets with [1:-1].
        // Strip here at the source so all consumers get the bare name.
        $$ = TokenInfo{Val: $2.Val, Line: $1.Line}
    }
    | TOK_LABEL
    {
        xtracer.Trace("parser.p_labelname__label ENTER (labelname)")
        // TOK_LABEL comes from lexer with brackets "[name]"; strip them
        // to match Python's p[N][1:-1] pattern applied by all consumers.
        val := $1.Val
        if len(val) >= 2 && val[0] == '[' && val[len(val)-1] == ']' {
            val = val[1 : len(val)-1]
        }
        $$ = TokenInfo{Val: val, Line: $1.Line}
    }
```

### Change 2: Remove redundant stripping in `labeledfmla` (line 2231)

**Before:**
```go
name := $1.Val
if len(name) >= 2 && name[0] == '[' && name[len(name)-1] == ']' {
    name = name[1 : len(name)-1]
}
lf := acfg(v17lex).NewLabeledFormula(acfg(v17lex).NewAtom(name), $2)
```

**After:**
```go
lf := acfg(v17lex).NewLabeledFormula(acfg(v17lex).NewAtom($1.Val), $2)
```

### Change 3: Remove redundant stripping in `TOK_THUNK` rule (line 4354)

**Before:**
```go
labelStr := strings.Trim($2.Val, "[]")
label := acfg(v17lex).NewAtom(labelStr)
```

**After:**
```go
label := acfg(v17lex).NewAtom($2.Val)
```

## Files to Modify

| File | Change |
|------|--------|
| `lalr_full/grammar_v17.y:2249-2260` | Strip brackets in `labelname` rule |
| `lalr_full/grammar_v17.y:2235-2238` | Remove redundant stripping in `labeledfmla` |
| `lalr_full/grammar_v17.y:4354` | Remove redundant stripping in `TOK_THUNK` |
| `lalr_full/grammar_v17.go` | Regenerate via `go generate` |

## Verification

```bash
cd ~/goivy && go generate ./lalr_full/ && go build ./... && make golden
```

Golden test should advance past line 61368.
