# Section 6 (ivy_compiler.py) TDD Work Plan

Based on /Users/jaten/go/src/github.com/glycerine/goivy/AUDIT18MARCH.md §6.2 (14 stubs) and §6.3 (18 behavioral differences) — 32 items total.

---

## Batch A — Small validation/guard fixes (low risk, quick wins)  (DONE)

- [ ] §6.3 #20: `compile_while_action` — validate no action calls in conditions
- [ ] §6.3 #21: `pullArgs` — raise IvyError instead of silent return
- [ ] §6.3 #25: `export`/`import_` — add `check_is_action` validation
- [ ] §6.3 #26: `attribute` — add validation of object existence / attribute name
- [ ] §6.3 #28: `mixin` — validate mixee is `init` or in `top_context.actions`

These are all small guard clauses / error checks. ~5 items, tightly scoped.

---

## Batch B — Compilation expression fixes  (DONE)

- [ ] §6.3 #15: `compile_action_def` — `prm:` prefix substitution
- [ ] §6.3 #16: `compile_action_def` — free-variable check in call args
- [ ] §6.3 #17: `compile_local` — assignment-as-declaration sort inference
- [ ] §6.3 #18: `compile_call` — field-reference fallback + param count validation
- [ ] §6.3 #32: `ExprContext.extract()` — `LocalAction` wrapping

These all touch the expression/action compilation path and are logically related.

---

## Batch C — Type system / sort infrastructure  (DONE)

- [ ] §6.2 #1: `FixConstructors` — adjust constructor domain sorts from destructor sorts
- [ ] §6.2 #3: `CreateConstructorSchemata` — constructor existence axiom schemata
- [ ] §6.3 #22: Ghost sorts — track via `ghost_sorts`
- [ ] §6.3 #23: Struct destructors — create with intermediate Variable parameter
- [ ] §6.3 #24: `variant` — populate `Supertypes` in addition to `Variants`

All related to the type/sort system. They interact with each other (constructors, destructors, variants).

---

## Batch D — Graph algorithms (Tarjan + dependencies)  (DONE)

- [ ] §6.2 #2: `CreateSortOrder` — topological sort with Tarjan SCC, error on cycles
- [ ] §6.2 #11: `TarjanArcs` — full Tarjan SCC instead of only filtering self-loops
- [ ] §6.2 #12: `GetSymbolDependencies` — transitive symbol dependencies (not just direct)

These share the Tarjan algorithm. Implement once, use in both places.

---

## Batch E — If/while existential compilation (DONE)

- [ ] §6.3 #19: `compile_if_action` — `Some`/`SomeMinMax` existential-if handling
- [ ] §6.2 #14: `compile_thunk_action` — subtype/destructor/substitution logic

Both deal with complex action compilation with existential/thunk semantics.

---

## Batch F — Conjecture/proof pipeline (DONE)

- [ ] §6.2 #5: `CheckDefinitions` — separate defs from props, check redefinition, detect cycles
- [ ] §6.2 #7: `CreateConjActions` — determine which actions must preserve each conjecture
- [ ] §6.3 #29: `ConjSetup` pass 2 — implement real conjecture setup (currently all no-op)

All part of the conjecture processing pipeline.

---

## Batch G — Native/compile_native + ARG pass 3

- [ ] §6.3 #27: `native` — call `compile_native_def` instead of storing raw node
- [ ] §6.3 #30: `ARGSetup` pass 3 — handle exports, delegates, progress properties

---

## Solo Items (treat individually due to size/complexity)

### Solo 1: `CheckPropertiesPass` (§6.2 #6) (DONE)
Full proof checking with `prover.AdmitProposition`. Touches the prover subsystem,
needs careful understanding of the proof-checking loop, and has deep interactions
with the `check/` package. Biggest single item.

### Solo 2: `ApplyAssertProof` (§6.2 #9)
Generate subgoals from ProofChecker. Tightly coupled to the proof system but
distinct from CheckPropertiesPass. Needs its own test harness for subgoal generation.

### Solo 3: `InferParameters` (§6.2 #10)
Parameter extension with body rewriting. Touches parameter resolution, AST rewriting,
and action compilation. Subtle enough to warrant focused attention.

### Solo 4: `AttachProofs` (§6.2 #4)
Matching labeled proofs to properties/conjectures. Depends on understanding the
proof labeling system end-to-end.

### Solo 5: `HandleTemporals` (§6.2 #8)
Labeling actions with isolate membership for temporal properties. Standalone
temporal logic concern.

### Solo 6: `TheoremToProperty` (§6.2 #13)
Full skolemization + sort renaming. Algorithmically non-trivial and self-contained.

### Solo 7: Main `ivy_compile` integration (§6.3 #31)
Missing `check_instantiations`, `add_to_hierarchy`, progress symbol removal,
`type_check`. This is integration glue — do it LAST, after the pieces it calls
are implemented.

---

## Recommended Order

```
A → B → C → D → E → F → solos → G → #31 (last)
```

Rationale:
- **A first**: quick wins, builds momentum, hardens error paths
- **B next**: core compilation fixes that many later items depend on
- **C**: type system must be correct before proof/conjecture work
- **D**: graph algorithms are self-contained, needed by sort order and dependency analysis
- **E**: existential compilation builds on B's expression fixes
- **F**: conjecture pipeline needs types (C) and definitions (D) working
- **Solos**: proof-related solos (1-4) form a natural sequence; temporals (5) and TheoremToProperty (6) are independent
- **G + #31 last**: integration glue after all pieces are in place
