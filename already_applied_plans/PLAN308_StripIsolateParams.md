# Plan: Fix StripIsolateParams skipping isolate params already in Sig.Symbols

Created: 2026-04-15 ~04:00 UTC

## Context

The previous BindingMap ordering fix (xtrace line 845791) is applied and working. Running `make golden` now shows a new divergence at **xtrace line 845853** in the Level 2 CanonSnapshot (temporal subgoal's CheckFragment):

```
845853  go : params=[]
        py : params=[(Symbol name:_T sort:(UninterpretedSort name:lclock))]
```

Go has empty `params` while Python has the isolate parameter `_T` of sort `lclock`. The Level 1 CanonSnapshot matched (both had `_T`), so `_T` IS being added during Level 1 `StripIsolateParams`. The issue is that it's being added only to the Level 1 module — but when Level 2 creates `fakeMod` via `mod.Copy()`, the copy should preserve `_T`.

Investigation reveals the root cause is in `StripIsolateParams` itself: the `_T` parameter is **skipped entirely** because its symbol already exists in `mod.Sig.Symbols` by the time Step 4 runs.

## Root Cause

In `StripIsolateParams` (`isolate/strip.go:694-723`), Step 4 adds isolate parameters to `mod.Params`. Two `continue` statements cause the append to be **skipped** when the symbol already exists:

```go
// Line 704-709: Go code
if alreadyAdded {
    continue           // ← BUG: skips mod.Params append
}
if _, exists := mod.Sig.Symbols.Get2(paramName); exists {
    continue           // ← BUG: skips mod.Params append
}
// ... only reaches mod.Params = append(...) if NEITHER fires
```

Python's equivalent (`ivy_isolate.py:472-487`) **always** appends to `mod.params`:

```python
add_map = dict((s.name, s) for s in strip_added_symbols)
for s in isolate.params():
    # ...
    if s.rep not in add_map:                              # NOT in strip-added
        sym = ivy_logic.add_symbol(s.rep, mod.sig.sorts[s.sort])  # get-or-create
        mod.params.append(sym)                            # ALWAYS append
    else:                                                 # IS in strip-added
        mod.params.append(add_map[s.rep])                 # ALWAYS append
    mod.param_defaults.append(None)                       # ALWAYS append
```

Python's `add_symbol` (logic.py:929-943) internally handles the "already exists" case by returning the existing symbol — it does NOT skip the `mod.params.append`. The `add_map` branch also always appends.

**Why `_T` triggers this**: During earlier strip processing, `_T`'s symbol gets registered in `mod.Sig.Symbols`. When Step 4 reaches `_T` in `isolate.params()`, Go's line 709 `continue` fires because the symbol exists, and `mod.Params` never gets `_T`. Python's `add_symbol` returns the existing symbol and `mod.params.append(sym)` runs normally.

## Changes Required

### 1. `isolate/strip.go:694-723` — Always append to mod.Params

Replace the current Step 4 logic:

```go
			// Check if already added via StripAddedSymbols
			alreadyAdded := false
			for _, added := range isoCfg.StripAddedSymbols {
				if added.Name == paramName {
					alreadyAdded = true
					break
				}
			}

			if mod.Sig != nil {
				if alreadyAdded {
					// Use existing symbol
					continue
				}
				if _, exists := mod.Sig.Symbols.Get2(paramName); exists {
					continue
				}
				if paramSort != nil {
					mod.Sig.Symbols.Set(paramName, &il.SymbolEntry{Name: paramName, Sort: paramSort})
					newSym := lg.NewConst(paramName, paramSort)
					mod.Params = append(mod.Params, newSym)
					mod.ParamDefaults = append(mod.ParamDefaults, nil)
				} else if s, ok := mod.Sig.Sorts.Get2(paramName); ok {
					newSym := lg.NewConst(paramName, s)
					mod.Sig.Symbols.Set(paramName, &il.SymbolEntry{Name: paramName, Sort: s})
					mod.Params = append(mod.Params, newSym)
					mod.ParamDefaults = append(mod.ParamDefaults, nil)
				}
			}
```

With code matching Python's two-branch structure:

```go
			if mod.Sig != nil {
				// Python: add_map = dict((s.name,s) for s in strip_added_symbols)
				// Python: if s.rep not in add_map:
				//             sym = ivy_logic.add_symbol(s.rep, mod.sig.sorts[s.sort])
				//             mod.params.append(sym)
				//         else:
				//             mod.params.append(add_map[s.rep])
				//         mod.param_defaults.append(None)
				var addedSym *lg.Const
				for _, added := range isoCfg.StripAddedSymbols {
					if added.Name == paramName {
						addedSym = added
						break
					}
				}
				if addedSym != nil {
					// Python: mod.params.append(add_map[s.rep])
					mod.Params = append(mod.Params, addedSym)
				} else if paramSort != nil {
					// Python: sym = ivy_logic.add_symbol(s.rep, mod.sig.sorts[s.sort])
					// add_symbol adds to sig if new, returns existing if already there.
					if _, exists := mod.Sig.Symbols.Get2(paramName); !exists {
						mod.Sig.Symbols.Set(paramName, &il.SymbolEntry{Name: paramName, Sort: paramSort})
					}
					mod.Params = append(mod.Params, lg.NewConst(paramName, paramSort))
				} else if s, ok := mod.Sig.Sorts.Get2(paramName); ok {
					if _, exists := mod.Sig.Symbols.Get2(paramName); !exists {
						mod.Sig.Symbols.Set(paramName, &il.SymbolEntry{Name: paramName, Sort: s})
					}
					mod.Params = append(mod.Params, lg.NewConst(paramName, s))
				}
				mod.ParamDefaults = append(mod.ParamDefaults, nil)
			}
```

Key differences from old code:
- **No `continue` statements** — all paths fall through to append
- `addedSym != nil` branch: appends the strip-added symbol (was: `continue`)
- Symbol-exists check: only guards `Sig.Symbols.Set`, not the params append
- `mod.ParamDefaults` append moved outside branches (Python always appends it)

## Key files

- `isolate/strip.go:694-723` — the lines to change (Step 4 of StripIsolateParams)
- `ivy_isolate.py:472-487` — Python source of truth (strip_isolate params loop)
- `ivy_logic.py:929-943` — Python `add_symbol` (get-or-create semantics)

## Verification

1. `go build ./...` — compiles clean
2. `make golden` — should advance past line 845853
3. `go test ./...` — should pass
