# Plan: Fix all missing fields in module.Copy()

**Created**: 2026-04-02 18:45

## Context

`module.Copy()` is missing 5 fields that Python's `copy.copy()` would copy. All fields must be copied for correctness.

## Missing fields

| Field | Type | Struct line | Fix |
|-------|------|-------------|-----|
| `Instantiations` | `[]Instantiation` | 51 | `c.Instantiations = append([]Instantiation{}, m.Instantiations...)` |
| `IsolateInfo` | `*IsolateInfo` | 55 | `c.IsolateInfo = m.IsolateInfo` (shallow, same as Python copy.copy) |
| `IsolateProof` | `ast.Node` | 57 | `c.IsolateProof = m.IsolateProof` |
| `InitCond` | `*co.Clauses` | 129 | `c.InitCond = m.InitCond` |
| `Name` | `string` | 138 | `c.Name = m.Name` |

## File

`module/module.go` — add 5 lines in `Copy()`, after the existing slice/map copies (around line 328).
