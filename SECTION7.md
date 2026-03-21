# Section 7 Analysis: ivy_solver.py → solver/

## Overview

27 audit items (9 MISSING, 8 STUB, 10 BEHAVIORAL_DIFFERENCE) covering the solver's Z3 integration layer.
The Go solver package is substantially ported (~186 KB across 8 source files) but has gaps in
sort handling, encoding, model extraction, and quantifier constraint placement.

---

## Grouping Strategy

Items are organized into 6 batches by functional cohesion. Within each batch, items share
data structures, caches, or call chains — fixing one often requires or simplifies fixing others.

---

## Batch A: Sort & Type System (5 items) — DO TOGETHER for cohesion

These items all touch the sort-to-Z3 mapping layer (`z3_sorts`, `z3_sorts_inv`, sort creation,
and cardinality). They share caches and conversion paths.

| Item | Category | Summary |
|------|----------|---------|
| 7.1#2 | MISSING | `uninterpretedsort()`/`functionsort()`/`enumeratedsort()` — sort creation with caching + `z3_sorts_inv` reverse map |
| 7.1#3 | MISSING | `sorts()` — BV, array, strbv, intbv, int, real, nat, strlit sort dispatcher |
| 7.1#7 | MISSING | `sort_from_z3()` — reverse map from Z3 sort ID to Ivy sort |
| 7.2#1 | STUB | `RangeSortBounds()` — returns hardcoded `(0, MaxInt32)` instead of parsing real bounds |
| 7.3#11 | BEHAV_DIFF | `sort_card` — only handles EnumeratedSort; missing BV (`2**size`), range, datatype cardinality |

**Why together:** All 5 operate on the sort mapping layer. `z3_sorts_inv` (7.1#2) is consumed by
`sort_from_z3` (7.1#7). The `sorts()` dispatcher (7.1#3) calls `uninterpretedsort`/`enumeratedsort` (7.1#2).
`sort_card` (7.3#11) needs the sort type info from `sorts()`. `RangeSortBounds` (7.2#1) feeds into
range sort construction in `sorts()`.

**Complexity:** Medium-high. Core infrastructure that other batches depend on. Do this batch FIRST.

**Python lines:** ~111-230, 357-369, 905-907

---

## Batch B: Term/Formula Conversion & Quantifiers (7 items) — DO TOGETHER for cohesion

The core conversion pipeline: how Ivy terms and formulas become Z3 expressions. These items
are tightly coupled through the conversion recursion and quantifier handling.

| Item | Category | Summary |
|------|----------|---------|
| 7.1#4 | MISSING | `atom_to_z3()` — enumerated equality encoding, polymorphic macro expansion |
| 7.1#5 | MISSING | `term_to_z3()` — variable name caching with `:sort_name` suffix convention |
| 7.1#6 | MISSING | `formula_to_z3_int()` — Def True/False simplification, quantifier constraints for nat/range |
| 7.3#1 | BEHAV_DIFF | Quantifier bound constraints — Python adds INSIDE quantifier body, Go adds at clause level |
| 7.3#5 | BEHAV_DIFF | `numeral_to_z3` — Python clamps out-of-range; Go doesn't clamp |
| 7.3#7 | BEHAV_DIFF | `my_eq` — missing True/False simplification for boolean equality |
| 7.3#9 | BEHAV_DIFF | Variable naming — Python caches as `v.rep + ':' + v.sort.name`; Go convention unknown |

**Why together:** `formula_to_z3_int` calls `atom_to_z3` which calls `term_to_z3`. Quantifier
constraints (7.3#1) are generated inside `formula_to_z3_int`'s ForAll/Exists handling. Variable
naming (7.3#9) affects `term_to_z3` caching. `my_eq` (7.3#7) is used in `atom_to_z3`. Numeral
clamping (7.3#5) is called from `term_to_z3`.

**Complexity:** HIGH. This is the most complex batch — the core conversion engine. Requires careful
reading of Python lines 414-650. Do this batch SECOND (after Batch A, since it depends on sort infrastructure).

**Python lines:** ~376-650, 83-98, 509-533

---

## Batch C: Binary Encoding (4 items) — DO TOGETHER for cohesion

Enumerated sort encoding as binary bit-vectors. Self-contained subsystem.

| Item | Category | Summary |
|------|----------|---------|
| 7.2#2 | STUB | `EncodeTerm()` — only handles cardinality; missing Ite, constructors, variables |
| 7.2#3 | STUB | `EncodeEquality()` — simple Eq instead of binary encoding with `gebin()` |
| 7.2#7 | STUB | `bfeToZ3()` — missing IntSort input, BV size overflow, zero-width, zero-extension |
| 7.3#8 | BEHAV_DIFF | `gebin` — binary predicate encoding vs simple integer Ge comparison |

**Why together:** `EncodeEquality` calls `gebin`. `EncodeTerm` produces the bit vectors that
`EncodeEquality` compares. `bfeToZ3` is the bit-field-extract counterpart. All deal with
binary/BV encoding.

**Complexity:** Medium. Self-contained math. Can be done in any order relative to other batches
(no dependencies on A or B for correctness, though A's sort infrastructure helps).

**Python lines:** ~1563-1640, 174-212

---

## Batch D: Model Extraction & Simplification (6 items) — DO TOGETHER for cohesion

How Z3 models are converted back to Ivy clauses. Tightly coupled pipeline.

| Item | Category | Summary |
|------|----------|---------|
| 7.1#8 | MISSING | `HerbrandModel.universes()` — returns dict from sorts to universe elements |
| 7.2#4 | STUB | `ClausesModelToDiagram` — creates trivial `X=X` equalities (always true) |
| 7.2#5 | STUB | `ClausesModelToClauses` — only handles 0-arity constants; missing functions, relations |
| 7.3#2 | BEHAV_DIFF | `clause_model_simp` — returns single true literal vs keeps all true/unknown literals |
| 7.3#4 | BEHAV_DIFF | `model_if_none` sort size search — simultaneous vs independent minimization |
| 7.3#6 | BEHAV_DIFF | `filter_redundant_facts` — activation literals vs push/pop |

**Why together:** `ClausesModelToClauses` (7.2#5) calls `model_if_none` (7.3#4) which calls
`clause_model_simp` (7.3#2). `ClausesModelToDiagram` (7.2#4) calls `filter_redundant_facts` (7.3#6)
and `ClausesModelToClauses`. `HerbrandModel.universes()` (7.1#8) provides the data that all
model-to-clauses functions consume.

**Complexity:** HIGH. Second most complex batch. The model extraction pipeline is where verification
results become human-readable counterexamples.

**Python lines:** ~824-903 (HerbrandModel), 1062-1080 (simp), 1135-1175 (model_if_none),
1292-1550 (model facts/diagram), 1373-1500 (model_to_clauses/diagram)

---

## Batch E: Solver Infrastructure (4 items) — Small items, DO TOGETHER

Smaller, mostly independent fixes to solver-level functions.

| Item | Category | Summary |
|------|----------|---------|
| 7.1#1 | MISSING | `clear()` — reset global Z3 caches (z3_sorts, z3_predicates, z3_constants, z3_functions) |
| 7.2#6 | STUB | `CheckNativeCompatSym()` — only checks sort compatibility, no runtime type verification |
| 7.2#8 | STUB | `SolverName()` — returns empty string for z3_builtins instead of raising error |
| 7.3#10 | BEHAV_DIFF | `check_sequence` — missing reporter with start/end callbacks, early abort |

**Why together:** These are all solver-level infrastructure. None are deeply entangled with each
other, but they're all small enough to batch. `clear()` is trivial. `SolverName` and
`CheckNativeCompatSym` are straightforward fixes. `check_sequence` reporter is a bit more work
but self-contained.

**Complexity:** Low-medium. Good warm-up or cool-down batch.

**Python lines:** ~228-240 (clear), 60-80 (solver_name), 326-348 (check_native), 966-1005 (check_sequence)

---

## Batch F: Clause Simplification (1 item) — DO ALONE due to complexity

| Item | Category | Summary |
|------|----------|---------|
| 7.3#3 | BEHAV_DIFF | `clauses_case` unit resolution — Python performs `UnitRes` propagation between rounds; Go skips entirely |

**Why alone:** Unit resolution requires importing/implementing the `ivy_unitres` module's algorithm.
This is a self-contained but algorithmically complex fix. It needs careful study of how Python's
`UnitRes` class works and how it interacts with the clause decomposition loop.

**Complexity:** Medium-high. Algorithmic (unit resolution propagation). May require porting parts
of `ivy_unitres.py` if not already ported.

**Python lines:** ~1031-1060 (clauses_case), plus ivy_unitres.py

---

## Recommended Execution Order

```
1. Batch E  (4 items) — Solver infrastructure — Low complexity warm-up
2. Batch A  (5 items) — Sort system          — Foundation for B and D
3. Batch C  (4 items) — Binary encoding      — Self-contained, medium
4. Batch B  (7 items) — Conversion pipeline  — High complexity, needs A
5. Batch D  (6 items) — Model extraction     — High complexity, needs A+B
6. Batch F  (1 item)  — Unit resolution      — Independent, needs research
```

**Rationale:** E warms up with easy wins. A establishes sort infrastructure. C is independent
and can be done early. B needs A's sorts. D needs both A and B. F is independent but requires
researching ivy_unitres.

---

## Item Cross-Reference (all 27 items)

| Batch | Items | Count |
|-------|-------|-------|
| A — Sort System | 7.1#2, 7.1#3, 7.1#7, 7.2#1, 7.3#11 | 5 |
| B — Conversion | 7.1#4, 7.1#5, 7.1#6, 7.3#1, 7.3#5, 7.3#7, 7.3#9 | 7 |
| C — Encoding | 7.2#2, 7.2#3, 7.2#7, 7.3#8 | 4 |
| D — Model | 7.1#8, 7.2#4, 7.2#5, 7.3#2, 7.3#4, 7.3#6 | 6 |
| E — Infrastructure | 7.1#1, 7.2#6, 7.2#8, 7.3#10 | 4 |
| F — Unit Resolution | 7.3#3 | 1 |
| **Total** | | **27** |

Note: 7.1#9 (arrsel/arrupd/arrcst) is marked **FIXED** in the audit and excluded from this plan.
