# Fix golden divergence: compileLabeledFormula type flow

Created: 2026-03-28

## Context

`make golden` diverges because `compileLabeledFormula` misuses `lg.Definition` as a carrier for (label, formula), losing the `LabeledFormula` identity and metadata. Python stores the compiled `LabeledFormula` directly in `AssertAction.args[0]` via duck typing; Go needs an explicit design.

## Approach: Add LF field to action types

Store the compiled `*ast.LabeledFormula` alongside the `lg.Expr` formula. This matches Python's pattern with stronger Go typing:

- `Formula lg.Expr` — the unwrapped logic expression (always set)
- `LF *ast.LabeledFormula` — the container with metadata (nil when absent)
- Python's `isinstance(fmla, LabeledFormula)` → Go's `a.LF != nil`

## Changes

### 1. `actions/action.go` — Add LF field

Add `LF *ast.LabeledFormula` to: AssertAction, AssumeAction, RequiresAction, EnsuresAction (any action that can wrap a LabeledFormula in Python).

### 2. `compiler/action.go` — Use ThingLF for LabeledFormula

In `CompileAssertFormula` and `CompileAssumeFormula`, when node is `*ast.LabeledFormula`:
- Call `c.ThingLF(lf)` instead of `c.Thing(node)` — gets PRESERVE clone trace
- Extract `compiledLF.Formula.(lg.Expr)` as the formula
- Set `res.LF = compiledLF`

### 3. `compiler/compiler.go` — Delete compileLabeledFormula

Remove `compileLabeledFormula` entirely. In CompileNode's `case *ast.LabeledFormula`, call `CompileLF` inline and return the extracted formula.

### 4. Audit downstream consumers

Search for any code that currently receives a `lg.Definition` from the LabeledFormula compilation path and expects to unwrap it. Update those sites.

## Verification

1. `go build ./...`
2. `cd ~/goivy && make golden` — advance past current divergence
3. `cd ~/goivy && make test` — existing tests pass
