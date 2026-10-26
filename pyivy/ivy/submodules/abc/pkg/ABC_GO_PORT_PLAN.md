# ABC C/C++ to Go port plan

Planning baseline: local source surveyed on 2026-10-26. The initial planning
phase introduced no Go source; this is now the living implementation plan and
records completed validation checkpoints as the port proceeds. Paths beginning
with `src/` are relative to the ABC repository; Go paths are relative to `pkg/`
unless stated otherwise.

## 1. Objective and binding implementation choices

Port the local ABC `src` tree to maintainable, robust Go, with a native Go runtime and no required C toolchain or cgo. The C source may be read while translating, but no checked-in test, generator, or retained oracle may require compiling or executing it. Prioritize the verification path used by Ivy, then expand to the rest of ABC. Finishing the Ivy path is an intermediate milestone, not completion of the full source port.

* Put coupled code in **one package `abc` at the root of `pkg/`**. Do not reproduce the C directory hierarchy as a Go import hierarchy. Use descriptive filename prefixes to preserve source provenance and searchability.
* Extract a separate package only for an obviously independent leaf with no dependency on `abc` and a concrete maintenance benefit. The default is no extraction. In particular, graph managers, SAT adapters, CNF, PDR, rewriting, mapping, timing, BDD integration, and command registration can all stay in `abc`.
* Prefer concrete structs, typed slices, integer IDs, explicit enums, and function callbacks. Do not introduce Go interfaces for solvers, graphs, passes, plugins, storage, or testing. Existing standard-library boundary interfaces such as `error` or `io.Writer` may be used where useful; they are not a reason to build an internal interface architecture.
* When dictionary iteration order matters, use `github.com/glycerine/insdict` **`insdict.Dict`**, already required at v0.14.6. Do not implement another ordered dictionary. Plain maps are appropriate only when enumeration cannot affect results, serialization, diagnostics, or heuristic choices.
* Do not use git commands: the user exclusively manages repository history. Inventory and permanent-dataset identity must use file content hashes and manifests instead.

`pkg/go.mod` currently declares module `github.com/glycerine/abc-go`, Go
`1.27.3`, direct `insdict` v0.14.6 and indirect xxhash/v2 v2.3.0. Preserve
these declarations unless an implementation slice requires a reviewed change.
Do not silently lower the Go version or upgrade dependencies.

## 2. What the survey found

### Scale and build boundaries

A recursive count of `.c`, `.h`, `.cpp`, and `.cc` files under the present `src` tree found **1,290 C files (862,725 lines), 316 headers (85,733 lines), and 19 C++ files (10,834 lines)**: 1,625 files and 959,292 physical lines, including comments and large data tables. This is a major multi-stage port, not a small C wrapper conversion. Counts include sources outside the default build and header-only implementations.

The top-level `Makefile` is the primary build manifest: `MODULES` includes base, graph, Boolean, optimization, mapping, SAT, and proof modules, with BDD modules enabled unless `ABC_USE_NO_CUDD` is set. Each `module.make` selects translation units. `CMakeLists.txt` obtains its source list and flags from make's `cmake_info`; it is not an independent complete inventory. The library excludes `src/base/main/main.c`. Readline, pthreads, CUDD, architecture definitions, extension directories, and PIC affect build behavior. `Makefile.emscripten` is another configuration to inventory, not a substitute for native coverage.

Presence on disk does not imply inclusion in the default build. Examples of additional trees include `sat/bsat2`, `proof/int2`, `phys/place`, `map/fpga`, `misc/espresso`, and `opt/fsim`. Account for them explicitly. Header-only code in `misc/vec`, `misc/hash`, `aig/miniaig`, `bool/deco`, and other modules is executable logic, not just declarations. `src/demo.c`, `starter.c`, `generic.c/.h`, and `template.c` also need dispositions outside the two-level module inventory below.

### Functional map

| Source family | Role and port implications |
| --- | --- |
| `misc/vec`, `util`, `mem`, `hash`, `st`, `nm` | Vectors, literal/bit operations, counters, names, allocation and hashing. Replace allocator mechanisms carefully; retain algorithmic ordering and identity. Not all `misc` code is independent. |
| `aig/aig`, `saig` | Pointer-oriented AIGs and sequential interpretation; foundational to the existing PDR path. Preserve object categories, complemented edges, register conventions, and mappings. |
| `aig/gia` | Large scalable AIG manager, compact fanin differences, extensive transformation/verification integration. Distinct from `aig/aig`, not a drop-in alias. |
| `aig/hop`, `ivy`, `miniaig`, `ioa` | Other graph representations and AIG I/O. ABC's `aig/ivy` is an ABC graph subsystem, not the Python Ivy project. |
| `base/abc`, `abci`, `io`, `cmd`, `main` | General networks, algorithm orchestration, readers/writers, command interpreter, global framework. `base/abci` is a major integration hub. |
| Other `base/*` | Word-level and hierarchical designs, Verilog, PLA, ESOP and related representations; schedule after the verification spine but retain full coverage. |
| `sat/bsat`, other `sat/*` | Several SAT solvers, CNF generation, bounded checking and related support. Glucose variants and bsat2 include C++, despite the shorthand “C port.” |
| `proof/pdr`, `cec`, `ssw`, `fra*`, `abs`, `int*`, `live`, etc. | Safety checking, equivalence, abstraction, interpolation, sequential reasoning and liveness. Result and witness semantics are critical. |
| `bool/*`, `opt/*` | Truth tables, decomposition, cuts, rewriting, resynthesis, retiming, simulation and optimization. Large precomputed tables need integrity checks. |
| `map/*`, `misc/tim` | LUT/cell mapping, technology libraries and timing; preserving function alone is insufficient to establish mapped/timing correctness. |
| `bdd/*` | CUDD and supporting BDD algorithms. Significant independent milestone with complement pointers, canonicalization, reference semantics, reordering and cache subtleties. |
| `misc/zlib`, `bzlib`; `phys/place` | Bundled compression and placement. Assess standard-library replacements for compression; placement is a distinct late feature family. |

### Evidence for one coupled package

Textual include sampling shows `aig/gia` depends on optimization (`opt/dau`), SAT, timing, extra utilities, and `base/main`; `base/abci` bridges nearly everything. `src/sat/cnf/cnf.h` directly includes `opt/dar/darInt.h` and uses DAR cuts. `src/proof/pdr/pdrInt.h` includes SAIG, CNF, hash utilities, GIA/AIG conversions and, by default, `sat/bsat/satSolver.h`. PDR implementation files also use BMC and frame services. Thus even apparently low-level CNF/PDR modules are not clean import leaves. These are observed includes, not a complete preprocessed call graph. Resolve actual function dependencies per work slice; do not delay the port to perfect that graph.

### Ivy integration contract

The adjacent `../../ivy/ivy_mc.py` contains `ABCModelChecker.cmd`, which invokes:

```
abc -c 'read_aiger <input>; pdr; write_aiger_cex <output>'
```

`check_isolate` emits AAG, converts it with the existing `aigtoaig`, runs ABC, checks the process status, and recognizes a proof by finding `Property proved` in stdout. Otherwise it attempts to consume a witness through `aiger_witness_to_ivy_trace2`. This makes output wording, witness format, failure behavior, latch/input order, and process exit status part of the current integration contract. Unknown must never be printed as proved. Because the existing caller treats non-proof output as a witness path, specify and test a safe unknown/error behavior before switching executables; a caller change may be needed later and is outside this planning-only edit.

`wasm-smoke-test.js` already contains BLIF/strash and AIGER/PDR proof and counterexample examples. `wasm-aiger-compare.js`, `i10.aig`, `src/demo.c`, `readmeaig`, and `abc.rc` provide additional seeds and workflow evidence. These scripts are useful starting points, not a comprehensive suite. The local README explicitly says there is no comprehensive regression test.

### Key representation facts

* `Aig_Obj_t` contains pointer-tagged fanins, unions, reference/level bitfields, traversal marks, and temporary data. Its manager uses constant **one** as the distinguished constant object.
* `Gia_Obj_t` stores compact fanin differences and complement bits; `Gia_Man_t` has many optional metadata/scratch arrays. Its constant convention differs from the AIG manager. Document and test conversion polarity rather than sharing a raw ID assumption.
* `readmeaig` requires all primary inputs before latch outputs, and primary outputs before latch inputs. Register counts partition CI/CO arrays. Cleanup can return a new graph and requires mapping remaps.
* `abc_global.h` defines `word` as 64-bit and uses low-bit literal complement operations. SAT, graph and external AIGER literal spaces must remain distinct even when their encodings look similar.
* PDR has per-output state, budgets, seeds, callbacks, abstraction options and ternary simulation, as well as multiple CNF paths. Preserve these incrementally with explicit capability tracking.

## 3. Destination layout and code organization

Recommended initial layout (future files, not created now):

| Destination | Contents |
| --- | --- |
| root of `pkg/`, package `abc` | All coupled library implementation and concrete public API. |
| `cmd/abcgo/` | Thin command-line entry point; parsing and framework behavior stay in `abc` where reusable. |
| `test_vectors/` | Small checked-in regression circuits, CNFs, expected semantic results, minimized fuzz inputs, provenance. |
| `test_vectors/oracle/` | Reviewed immutable datasets and small expected outputs, hash-verified without C or machine-specific paths. |
| `tools/` | Go-independent inventory or data-maintenance scripts only; C helpers and native-executable test drivers are forbidden. |

Filename examples: `aig_aig_manager.go`, `aig_gia_hash.go`, `sat_bsat_solver.go`, `sat_cnf_derive.go`, `proof_pdr_core.go`, `base_io_aiger.go`. Match tests by area with `_test.go`. Prefix types and helpers sufficiently to avoid collisions across the large package; do not export every C symbol. Keep a symbol/source cross-reference so C names remain searchable. Large package does not mean one enormous file.

Begin with both AIG and GIA concrete managers as required, not a universal graph interface or an early unification rewrite. Use conversion functions returning explicit old/new object, input, latch and property mappings. Port only the GIA subset required by the first slice initially; other GIA algorithms follow later. Use concrete SAT solver types; if alternate backends eventually need runtime selection, an enum plus explicit dispatch or typed adapters suffices.

## 4. Translation rules and correctness contracts

### Types, literals and indexing

Choose named integer types for graph IDs, graph literals, SAT variables/literals, clause references and frame indices. Specify width, maximum value, invalid value and whether zero is valid. Prefer fixed-width unsigned values for packed formats/bit operations and `int` for checked slice indexes. Check conversions and allocation arithmetic before allocating; do not let attacker-controlled counts wrap. Never cast `-1` into a valid large unsigned ID accidentally.

Use explicit `(id << 1) | complement` style graph literals with checked constructors and XOR-one complement where appropriate. Translate pointer tagging to literal values, never to `unsafe.Pointer` tricks. Do not infer AIG's constant polarity from GIA or AIGER. Keep SAT result tri-states distinct from graph truth values and ternary simulation values; translate each C API's numeric conventions at its boundary.

Represent C bitfields using explicit masks/accessors or ordinary typed fields first. Go's complement and shifts differ from C promotion rules: fix widths before bit operations, mask truth-table tail bits, handle shifts at word boundaries and signed right shifts explicitly. Use `math/bits` where it preserves the intended operation. Endianness and C struct layout are not serialization specifications.

### Ownership and mutation

Use stable IDs into manager-owned slices for dense graphs and clause arenas. Appending can relocate a slice: do not retain pointers to its elements across growth. Object deletion, compaction and slot reuse need documented invalidation or generation rules. Test stale handles deliberately. Expose clones or read-only access conventions where callers could mutate invariant-bearing slices.

Replace unions/`void *` and generic temporary fields with typed scratch arrays or concrete dedicated structs. Give scratch state a phase owner and lifecycle; nested calls must not overwrite outer state. Restore marks on early exit and clear traversal arrays safely on counter rollover. Distinguish nil/uninitialized from valid empty state.

Garbage collection replaces C allocation mechanics, not semantic reference counts used for graph fanouts, reclamation decisions, or BDD canonicalization. Preserve those semantics. Begin without `sync.Pool` or custom unsafe arenas; optimize allocation only after correctness and profiles. Explicit reset/close operations should clear large references and have documented reuse behavior.

### Ordering and reproducibility

Use insdict for command/name registries and order-sensitive dictionaries; use insertion order based on deterministic source traversal. For composite comparable keys, inspect v0.14.6 hash support and supply a stable field-wise hash with `NewDictFunc` when necessary; avoid raw struct memory/padding hashing. Test update, delete/reinsert and clone behavior against the exact dependency. Treat ordered dictionaries as mutable shared objects: do not copy them by value or mutate during iteration without checking documented semantics.

Dense arrays and explicit queues are better than dictionaries for numbered objects. Plain-map iteration must not determine SAT variable numbering, clause order, cut tie-breaking, output order or traversal order. Preserve C traversal/tie-break rules when they influence heuristics. Use an explicit seeded PRNG; port the C PRNG if exact sequence compatibility matters. Fix seed and logical work budgets in reproducibility tests. Wall-clock-limited searches can differ legitimately.

### API, errors and execution

Use concrete options and result structs. Provide explicit safe/proved, unsafe/counterexample, and unknown statuses, plus termination reasons such as conflict/frame/time limits or cancellation. Invalid input and internal errors must not masquerade as a verification answer. A SAT model proves existence only after evaluation; an UNSAT result for a bounded query proves only that bound.

Keep all mutable frame state, libraries, RNGs, diagnostics and registries under explicit instance ownership. Replace `Abc_Start`/`Abc_Stop` global initialization with construction and scoped lifetime. Do not launch automatic background goroutines. A manager/solver may initially be single-owner; independent instances must be isolatable and race-free. Use concrete callbacks for stop/failure notifications with defined reentrancy rules. A standard `context.Context` boundary is optional; no custom cancellation interface is needed.

Retain internal invariant checks as development assertions or explicit validation routines. Malformed external data returns errors, not assertion panics. Library code must not call `os.Exit`. Commands call typed library operations. Implement only supported command options and explicitly reject the rest until ported; no silent no-op compatibility stubs.

### Formats and semantic obligations

AIGER requires explicit handling of binary deltas/varints, ASCII records, maximum literal bounds, symbols/comments, latch resets, bad properties, constraints, justice and fairness where supported. Verify this snapshot's actual reader support before promising a feature; reject unsupported sections explicitly. Never silently reinterpret liveness as safety or unknown latch initialization as zero. Bound input bytes, counts, nesting and decompression expansion. Decode with a cursor and checked EOF/overflow at every step; the unchecked pointer decoder in `ioaReadAig.c` is not a safe parser template.

Specify input and latch ordering, bit packing and frame numbering for counterexamples using `utilCex`, relevant AIGER writers, and Ivy's decoder. Replay on the original transition system after any transformation, honoring assumptions, initialization and property selection. Output files must not leave stale witnesses from an earlier run. Use temporary output plus successful completion/rename where appropriate.

For each pass, classify whether it preserves combinational function, sequential traces, equisatisfiability, mapped timing intent, or an abstraction relation. These require different validators. Never apply a combinational equivalence check as the sole oracle for retiming or state abstraction.

## 5. Inventory and tracking before implementation

Create a source manifest during the first implementation milestone covering every source/header, `module.make`, relevant configuration and generated data asset. Record SHA-256, size, default/config-specific inclusion, exported symbols, important inline helpers/macros, dependencies, destination, tests and disposition. Do not modify C originals for the Go port. Translate observations into reviewed permanent datasets or independent Go models; do not retain a C generator or executable as part of the test system.

Every item gets a status: unstarted, in progress, ported, validated, replaced, example-only, or explicitly deferred with reason. “Not in default make” is not a completion status. A standard-library replacement needs behavioral tests and a compatibility note. Unreachable code must be demonstrated per configuration; a successful link is insufficient. Track exposed command/option coverage separately from internal symbol coverage.

For each small work slice: read the C implementation plus relevant headers and callers; record preconditions and mutation effects; add independent expected results; translate using the rules above; run dataset/property tests; update the manifest. Do not preserve undefined behavior as compatibility. When historical behavior and independent semantics disagree, minimize the case and record the adjudicated behavior in a permanent dataset and Go test.

The C tree is translation input, not a permanent test service. Do not create or
retain per-feature C capture programs, differential executables, cgo bridges,
compiler-driving scripts or dataset refresh recipes that execute ABC. A port
slice is not test-complete until its behavior is covered by Go-only unit tests,
an independent bounded Go model, exhaustive enumeration, certificate/witness
replay, round-trip or algebraic invariants, and relevant fuzzing. Fixed datasets
under `test_vectors/` are allowed when they preserve reviewed compatibility
facts, but their expected meaning must be documented and checked independently
in Go. They are immutable regression evidence, not output that maintainers are
expected to regenerate from C. If a dataset cannot be understood or defended
without rerunning a C oracle, replace it with a specified Go model or remove it
before declaring the source slice validated.

## 6. Ordered implementation milestones and exit gates

### M0 — Reproducible baseline and harness design

Freeze a content-hashed source manifest and a stand-alone dataset manifest. Record dataset schemas, seeds, modeled invariants, intentional corrections and fixture hashes. Every retained row must be consumable by Go alone and supported by an independent Go model, exhaustive check, replay/certificate check, algebraic invariant or reviewed fixed expectation. Inventory actual Ivy commands/options and witness parsing. Review existing WASM fixtures and establish provenance for a broader corpus without adding a native runtime dependency.

Exit: all tests run with only the Go toolchain and repository datasets; every discrepancy has saved input and metadata; unsupported capabilities and current source identity are explicit. No test or ported algorithm depends on a C compiler, ABC archive, ABC executable, cgo, or subprocess oracle.

### M1 — Foundations and representation contracts

Port required utility/literal/bit operations, vector behavior using slices, names/hashes, counterexample storage, typed options/results and manager lifecycles. Port required header inline code, not just `.c` bodies. Define graph/SAT literal domains and check routines before dependent algorithms.

Exit: exhaustive small bit/truth tests, boundary/overflow tests, ordered-dictionary adapter tests, reset/reuse tests, and deterministic repeat runs pass. No pointer arithmetic or broad `any` containers.

### M2 — Graph spine and AIGER

Port AIG construction, hashing, traversal, duplication, cleanup and sequential CI/CO conventions. Add required GIA construction/conversions without attempting all GIA features. Implement binary AIGER reading for Ivy and witness serialization; ASCII support is useful for fixtures but must be documented separately from C parity. Add semantic simulator and small explicit-state reference checker in tests.

Exit: small graphs exhaustively simulate; conversions and cleanup preserve traces and mappings; parser mutation fuzzing is bounded; AIGER round trips preserve properties and initialization. BLIF and `i10.aig` become additional gates as their readers become available, not prerequisites for initial AIGER PDR.

### M3 — SAT and CNF

Port the `sat/bsat` backend actually used by PDR first: literal/assignment conventions, clause arena, watched literals, propagation, decision trail, conflict analysis, learning, backtracking, assumptions, incremental addition, restarts, reductions, seeds and limits. Other bsat variants and proof/interpolation support remain tracked work.

For CNF, implement a simple independently testable Tseitin encoder as a reference path. Port ABC's required optimized `sat/cnf` path and DAR/Boolean helpers when reaching parity; the simple encoder is not a claim that cut-based CNF has been ported. Keep precise graph-variable and auxiliary-variable mappings. Compare existential projection, not arbitrary auxiliary model bits.

Exit: exhaustive tiny SAT formulas and incremental operation sequences match brute force; every SAT model is validated; assumptions/cores and repeated solves are checked; CNF matches graph truth for all small assignments, with target output assertions explicit. Unknown/budget behavior is tested.

### M4 — Bounded safety and PDR

Implement bounded unrolling/simulation needed to cross-check sequential behavior, then PDR cubes/obligations, per-frame solvers, relative induction, generalization, clause pushing/subsumption, convergence, invariant extraction, counterexample construction and budget handling. Follow actual PDR call dependencies including ternary simulation, dynamic CNF and BMC helpers. Start with a documented default-option safety subset and explicitly reject unsupported non-default options. Expand option coverage before claiming PDR parity.

Exit: safe/unsafe/unknown agree semantically on tiny reachable state spaces; every returned counterexample replays; exported invariants pass independent initiation, consecution and safety checks. Test initial-frame bugs, deep bugs, multiple outputs, vacuous constraints, unconstrained inputs, no latches, no inputs and cancellation.

### M5 — Ivy-compatible executable slice

Add the thin `cmd/abcgo` entry point and required command parsing/framework state, `read_aiger`, `pdr`, and `write_aiger_cex` options. Preserve the observed proof marker and witness contract. Check quoting, semicolon sequences, paths with spaces, errors and repeated commands. Do not implement arbitrary shell execution to emulate commands. Specify `abc.rc`/startup handling and avoid ambient user startup files in deterministic tests.

Exit: reviewed permanent datasets and Go-only replay/certificate checks feed Ivy's real witness decoder correctly on curated Ivy examples. Safe returns no witness, unsafe returns a valid witness, and unknown/errors cannot be mistaken for a proof or a successful unsafe answer. Any migration-time C comparison is manual and disposable; it must not become a checked-in test gate, generator, helper, or prerequisite. Do not replace user binaries automatically.

### M6 — Combinational synthesis and equivalence

Expand general `base/abc` networks, BLIF/SOP support, HOP/IVY, Boolean truth/decomposition helpers, cuts, DAR/RWR and other rewrite/resubstitution engines. Port CEC/FRAIG and simulation support. Make the `src/demo.c`/README rewriting flow a semantic regression, then broaden options and graph representations.

Exit: every pass validates its structural invariants and proves or exhaustively establishes equivalence on small cases. Larger corpus miters are independently checked; expected optimization counts are secondary diagnostics unless algorithm identity is intended.

### M7 — Sequential verification and transformations

Port remaining BMC, sequential simulation, SSW/FRA, DCH/CEC variants, interpolation, abstraction/refinement, retiming, liveness and remaining PDR features according to the coverage ledger. Port additional solvers (`satoko`, `xsat`, `msat`, Glucose variants, bsat2 and others) as concrete implementations with backend-specific tests, not empty adapters to bsat.

Exit: explicit-state small-model oracles, proof/witness validation and transformation mapping checks cover each feature. Abstraction counterexamples undergo refinement/concretization; liveness gets accepting-cycle/fairness validation and cannot inherit a safety-only gate.

### M8 — Mapping, timing, word-level and hierarchy

Port `map/*`, technology library parsers, `misc/tim`, remaining word-level/hierarchical `base/*`, Verilog and formats, and placement. Keep related code in `abc`. Respect signedness, bit widths, overflow, four-state/unknown semantics where represented, hierarchy connectivity, clock/reset semantics, library units and timing constraints.

Exit: tiny word-level expressions match exhaustive bit-vector evaluation; mapped netlists simulate against their inputs; pin truth tables, area accounting and timing on hand-calculable examples agree with expected values. Exercise empty/malformed libraries and unresolved modules.

### M9 — BDDs, bundled facilities, tail completion and optimization

Port CUDD and supporting BDD operations deliberately, including variable ordering/reordering, unique/cache tables, complemented references and lifecycle behavior. Keep them in `abc` unless an independent leaf is demonstrably easy. Evaluate zlib/gzip standard-library equivalents and `compress/bzip2` reading; Go's standard library has no bzip2 writer, so explicitly port or select a reviewed concrete dependency for required writing. Verify ABC-specific wrapper semantics and concatenated/truncated streams. Handle all omitted/default-disabled trees, auxiliary sources and demonstration programs through the ledger.

Exit: no unaccounted source/header/command/configuration remains. Any deferred functionality is named and means full port completion has not been claimed. Run complete semantic, fuzz and performance gates before optimizing. Profile hot paths, allocations and GC, then optimize one measured bottleneck at a time with the same correctness corpus.

## 7. Comprehensive unit and property test matrix

Tests belong beside implementation in the big package so they can inspect internal invariants; add a small external-package API suite later. Use concrete test helpers and independent slow models rather than mock interfaces. Every bug fix adds a minimized permanent regression.

| Area | Required unit/property cases |
| --- | --- |
| Literals and words | Complement involution, variable round trip, all small literals, constants, invalid/sentinel values, 31/32/63/64 boundaries, tail masks, checked multiply/add and index conversion. |
| Slices/arenas/dictionaries | Empty/singleton/growth/reset/clone; deletion/reuse, stale IDs, nested scratch use; collision-heavy keys, deterministic insertion/update/delete/reinsert order, independent managers. |
| AIG/GIA/network | AND identities, complements, fanin canonicalization, duplicate hashing, shared/reconvergent cones, dangling cleanup, fanout/refcounts, cycles rejected, levels/topological order, traversal rollover, conversion polarity and CI/CO mappings. |
| Truth/decomposition/cuts | Exhaustive small truth tables, cofactors, support, permutation/phase transformations, NPN transforms, ISOP/SOP evaluation, cut leaf order and composed truth, table integrity. |
| SAT | Empty formula SAT, empty clause UNSAT, units, duplicates, tautologies, contradictory assumptions, absent variables, repeated solve/add/backtrack cycles, restarts/reduction/recycling, learned clauses, cores, budget interruption and reuse. |
| CNF | Every primitive truth table, constant outputs, complemented edges, shared nodes, multiple outputs, output polarity, unused nodes/vars, projections and mapping after cleanup. |
| Sequential/PDR | Counter/reset/shift-register models, safe invariants, delayed bugs, dead ends, constraints, multiple properties, arbitrary/zero/one initialization, frame zero and latchless cases, generalization soundness, convergence and per-output results. |
| Transformations | Exhaustive combinational outputs or sequential trace relations as appropriate; name/property/latch remaps; choices/equivalence classes; metadata invalidation after structural changes. |
| I/O and commands | Truncation at each byte boundary, varint overflow, count mismatches, forward refs, invalid latch resets, duplicate names, symbols/comments, unsupported sections/options, escaped paths, output errors, stale CEX prevention and exact Ivy proof marker. |
| BDD | Canonical equality, complement/cofactor/apply/quantify, zero/one, variable permutation/reordering and cache invalidation, reference lifecycle, exhaustive truth comparison on small supports. |
| Mapping/hierarchy/timing | Cell pin polarity/order, malformed libraries, hierarchy resolution, signed extension/truncation, known small mapped circuits, timing/area hand calculations and deterministic ties. |
| Lifecycle | Repeated create/run/reset/drop, cancellation at multiple phases, independent concurrent instances under race detector, no leaked goroutines/file descriptors, bounded retained heap across reuse. |

For cut stores, expose a concrete owned lifecycle with explicit failure and result-transfer semantics. The current fixed-graph `GiaCutStore` validates and snapshots the graph, initializes references before ranking, advances one gate at a time and transfers the completed result once. Verify paused/resumed computation against native batch cut sets, truth insertion IDs and statistics. Check cancellation inside a merge, resource failures after partial interning, abandoned-store cleanup and graph mutation isolation. Keep live-graph append/reference updates and reusable allocation/reset as separate outstanding work: fixed reference costs must not silently change while ranked cuts survive.

CI-support truth APIs must distinguish an oversized structural support from the all-zero truth word. `CITruth6` retains sorted CI object IDs and returns an explicit Fits flag alongside the native zero sentinel for more than six inputs. Verify structural support independently of Boolean support: cancellation can remove functional dependence while the CI must remain in structural support. Propagate support masks independently and exhaustively compare truth values on source assignments, including constants, COs, buffers, native XOR/mux and unused register-output CIs. Keep collection iterative and preserve graph scratch.

Boundary truth APIs must retain the specific C entrypoint's duplicate-leaf convention. `BoundaryTruth6` ports `Gia_ObjComputeTruth6`, where each scratch write overwrites the previous one and a repeated leaf takes its last slot. `CutTruth` follows the separate collector's first-slot convention. Preserve unused slot positions during deduplication/embedding, and compare both conventions explicitly with pinned native vectors. Automatic lookup of a LUT boundary from mapping storage remains separate from evaluating an explicit structural boundary.

Represent native LUT mappings with concrete owned packed storage. `GiaLUTMapping` keeps object-indexed offsets and `[count, fanins..., signed mux footer]` records. Iteration follows object order regardless of record insertion order; zero-object constant records are queryable but skipped by native LUT statistics/iteration. Validate imports before indexing: cap the complete storage, reject malformed/overlapping referenced records and non-topological fanins, and preserve unreferenced stale records produced by replacement. Compare packed storage, mux markers, empty maximum-size sentinel (-1), complete object levels and CO maximum levels against C, including dangling LUTs deeper than outputs. Mapping truth uses the stored ordered boundary. Selection heuristics, mapping2, timing boxes and graph-edit propagation remain separate unfinished work.

Mapped reference counts include every CO edge and every mapped fanin slot, including duplicate slots and dangling LUTs. The native `pRefs` and `pLutRefs` representations contain the same counts; return one owned array. Overlap is the sum of unique internal nodes per LUT cone minus all source AND nodes, and can be negative for partial mappings. `GiaLUTMapping.Analyze` preserves this definition, reuses epoch-marked private scratch, checks aggregate cone-visit work, and returns no partial state on cancellation or uncovered CI boundaries. Differentially check both native overlap implementations; use independent reverse topological reachability to check shared-node counting and deep iterative traversal.

Mapping connectivity verification must follow the native dependency graph, not substitute structural cone coverage for it. `GiaLUTMapping.Verify` starts at all buffer drivers (including dangling buffers) and CO drivers, requires records for reached non-buffer gates, then follows those records' fanins. Constants, CIs and buffers terminate each dependency path. Unrelated dangling gates may remain unmapped. Return a Boolean for incomplete mapping, with errors reserved for invalid context, budgets and cancellation; keep diagnostic printing out of library APIs. Test all mapping subsets, both direct and flattened boundaries, buffer chains and CI-only outputs against C and an independent bottom-up model. `Analyze` and truth evaluation separately establish structural-boundary validity and functional behavior.

Cut truth oracles must respect the leaf domain. CI leaves are independent, so compare all free leaf assignments. Composed functional cuts with internal leaves can retain correlations from the source graph; verify them by exhaustive CI assignments and the resulting leaf values. The pinned `test_vectors/gia_cut_enumeration_regression.json` demonstrates a native cut implementing XOR(a, g) where g = !a & b: direct structural-boundary evaluation gives OR(a, g), and the two agree on every realizable source assignment because a and g cannot both be true. Keep both representations and the native function ID in the regression. Whole-mapped-CNF projection tests must include the defining logic of internal leaves.

Small independent models are essential: use truth enumeration, a direct expression evaluator, brute-force SAT, explicit BFS over reachable states, and a straightforward transition simulator. They should not call the production encoder or reuse the same graph evaluator for both sides of an equivalence assertion.

## 8. Differential testing and proof validation

Use three complementary stand-alone checks: reviewed permanent datasets for compatibility, independent exhaustive small Go models for semantics, and mathematical certificate/witness checks where available. Dataset agreement alone can preserve a historical bug. Neither equal node counts nor matching formatted output proves correctness.

References below to “native rows”, “native vectors”, or comparisons with C
describe the provenance or behavior being ported; they do not authorize a C
program in `pkg/tools/` or a test-time differential run. Convert such evidence
once into reviewed data under `test_vectors/`, then make the lasting oracle a Go
model/property/certificate check. New work must normally derive expectations
directly in Go. `pkg/tools/` is limited to maintained inventory or dataset
housekeeping scripts that neither compile nor execute native code; generated
executables, object files, caches and helper sources do not belong there.

For each generated case, save seed, generation parameters, raw input, normalized semantic problem, feature flags, Go source and dataset hashes, limits and results. Keep generation and checking in Go. Exclude timestamps, memory addresses and timing from permanent comparisons while retaining semantic result markers and parseable artifacts. Treat parse errors, resource limits and cancellation as distinct outcomes.

* SAT: check the returned assignment directly. For UNSAT on small formulas enumerate assignments; for larger cases use an independently checked proof when the backend supports a suitable certificate. Do not assume every ABC backend can emit the same proof format.
* CNF: for every small primary assignment, existentially quantify auxiliary variables and compare with the original Boolean relation, including assumptions/output assertions.
* CEC/rewrite: exhaustively evaluate tiny problems; otherwise build a miter and solve/check it independently. A timeout is inconclusive.
* Safety: replay counterexamples on the original model. Validate candidate invariant `I` with `Init => I`, `I ∧ T => I'`, and `I => Safe`, respecting constraints. Use an independent small checker or separately encoded queries; the same buggy encoder cannot be the only validator.
* Sequential transformations: compare trace semantics with explicit remapping of state and inputs; bounded agreement alone is not a general proof.
* Liveness: validate loop closure and acceptance/fairness on lasso witnesses, and use independent small SCC analysis for expected answers.

Different valid SAT assignments, witnesses, graph numbering and proof traces are allowed. First compare semantics; use exact traces only in targeted deterministic tests. A timeout or resource limit on either side is not evidence of disagreement in truth value. Any safe-versus-unsafe disagreement blocks promotion until minimized and explained.

## 9. Native Go fuzzing program

Use Go's built-in fuzzing after implementation, with separate named targets so campaigns are reproducible and failures minimize well. Keep every fuzz loop entirely in process and compare against independent Go models, invariants, certificates or permanent datasets.

| Planned target | Input and assertions |
| --- | --- |
| `FuzzAigerRead` | Raw bytes, valid seeds plus mutations; bounded parse, no panic/hang, checked allocation, structural validity on success. |
| `FuzzAigerRoundTrip` | Structured valid circuits; serialize/parse and compare semantics, initialization, constraints, properties and symbols that the format preserves. |
| `FuzzWitnessReadReplay` | Witness bytes with a small model; errors stay errors, accepted witnesses have consistent dimensions and replay to the selected failure. |
| `FuzzGraphOperations` | Operation stream constructing, hashing, cloning, cleaning and converting bounded graphs; check invariants and independent truth after every mutation. |
| `FuzzSATIncremental` | Bounded add-clause/assume/solve/reset stream; compare each solve with brute force and check models/cores and reuse after interruption. |
| `FuzzCNFProjection` | Small generated graphs/assignments; compare existential CNF projection with direct Boolean evaluation. |
| `FuzzTruthAndCuts` | Tables, leaf permutations and decompositions; evaluate reconstruction and bounds independently. |
| `FuzzSequentialPDR` | Tiny valid machines, constraints and initialization; compare with complete reachable-state search, replay bugs and check invariants. |
| `FuzzRewriteEquivalence` | Bounded random DAGs and pass options; preserve function/declared sequential relation plus metadata mappings. |
| `FuzzBDDOperations` | Small Boolean operation sequences, quantify/reorder/reset; compare with truth-table model and canonical invariants. |
| `FuzzFormatAndCommand` | Each additional parser and command tokenizer separately; bounded nesting/output, no host shell execution, valid error state. |

Set explicit bounds before allocation or expensive solving. Example starting limits for semantic fuzzers: at most 8 Boolean variables for brute-force SAT, 6 primary inputs for exhaustive combinational checks, and 4 latches/3 inputs for explicit-state sequential checking. These are tunable test limits, not production limits. Bound operation streams, frames, clauses and solver conflicts as well as input size. Raw parser fuzzing should exercise large declared counts without allocating them. Use deterministic work budgets in assertions; use a process watchdog for hangs and wall-clock limits only as infrastructure.

Seeds include constants, empty/one-node graphs, complemented outputs, duplicate/tautological clauses, disconnected cones, self/reset latches, maximal legal varints, every EOF boundary, existing WASM proof/bug examples, and minimized historic failures. Maintain a fast checked-in seed set per target and a larger external campaign corpus. Preserve original failing bytes and seed on shrink; structured generation complements raw-byte mutations rather than replacing them.

Fuzzer failure classes: unexpected panic, nontermination, excessive allocation, invalid successful parse, changed semantics, wrong SAT/PDR status, invalid witness/invariant, nondeterministic output under fixed conditions, and state corruption after errors. A successful rejection of malformed input is not a failure. Investigate invariant-check failures even when final results happen to agree.

## 10. Verification cadence, CI and performance

After Go code exists, run from `pkg/` using the module's declared toolchain:

```
go test ./...
go vet ./...
go test -race . -run '<changed-lane tests>'
# At milestones (normally every third integrated wave) and releases:
go test -race ./...
go test -shuffle=on -count=10 ./...
go test -run '^$' -fuzz '^FuzzAigerRead$' -fuzztime=60s .
go test -run '^$' -bench . -benchmem ./...
```

These are planned commands, not checks performed in the original planning phase. Run one fuzz target per invocation. Add tests as each feature lands; do not wait until the final milestone. Every test, including dataset compatibility tests, must run without an ABC executable, archive, header tree or C compiler. A skipped native job is not a supported test mode.

Per change: affected unit/property tests, targeted race/independent-instance
tests for every changed lane, package-wide ordinary regression, static checks,
short relevant fuzz smoke, and permanent dataset fixtures. Run the full package
race suite at major milestones, normally every third integrated parallel wave,
and whenever shared synchronization, lifecycle, or cross-cutting concurrent
state changes; also run it before releases. The most recent full-race checkpoint
was wave 33, so the next scheduled periodic checkpoint is wave 36. Before
milestone promotion: full unit suite, the applicable race gate, shuffled repeats
and feature-specific stand-alone model/dataset checks. Nightly: rotate all fuzz
targets with recorded worker/time budgets and a larger corpus. Weekly/release:
extended campaigns (hours per high-risk target), supported OS/architecture
matrix and larger industrial/Ivy benchmarks, all through Go-owned parsers,
replayers and certificate checkers.

### Parallel implementation waves

Use parallel agents only for source units with disjoint implementation/test
files and no unresolved shared API design. The supervising agent owns
assignment boundaries, dependency review, all shared documentation and the
authoritative coverage ledger. Workers must not edit `PORT_COVERAGE.json`,
`PORT_STATUS.md`, `ABC_GO_PORT_PLAN.md`, module metadata or another worker's
files. A worker reports the exact source definition inventory, intentional
semantic corrections, owned files and targeted test evidence.

Run at most the available concurrency limit: one supervising integrator plus
the remaining slots as implementation workers. More assignments proceed in
successive waves rather than increasing shared-filesystem contention. Give
each worker exclusive Go/test filenames and, only when needed, an exclusive
subdirectory below `test_vectors/`. Workers use no git operations, C/cgo/native
oracles or permanent `testdata/` corpora.

Parallel implementation ends after wave 19 to control token expenditure.
Beginning with wave 20, perform only one porting or review task at a time: no
concurrent implementation lanes and no background reviewer while another lane
is active. Preserve the same independent source review, focused tests, ledger
discipline and integration gates; execute those phases serially in the single
active thread.

The supervisor independently reviews every returned implementation against the
complete source, checks definition-to-ledger equality, resolves cross-unit API
or naming conflicts and rejects incomplete slices. Targeted worker tests are
necessary but never sufficient for promotion. After combining a wave, run the
full ordinary package suite, relevant 5,000-execution Go-only fuzz/model
campaigns, vet, targeted race tests for every changed lane and native-source/
package-header/`testdata`/symlink/tool allowlist audits. Run the full package
race suite on the milestone cadence above rather than after every wave. Update
the ledger, plan and status only after the combined state passes every
applicable gate. A failed worker or integration gate leaves that source
unstarted or partial; it must not be hidden by promoting the other results.

Measure correctness first, then parse throughput, solver conflicts/propagations, PDR frames/obligations, graph node count, allocations, peak/retained heap and runtime on fixed hardware/seeds/configurations. Keep small, medium and large cases. Logical work counts distinguish heuristic drift from implementation slowdown. Establish numeric performance budgets from baseline measurements in M0/M3/M4; do not invent a universal C-to-Go ratio. Profile before changing layout, inline helpers or allocation strategy. No performance gain justifies a verification soundness regression.

## 11. Completion criteria and risk management

The main risks are unsound proof answers, literal polarity mistakes, lost initial-state/constraint semantics, remapping errors, incremental solver corruption, scratch aliasing, nondeterministic heuristics, parser resource exhaustion, and undercounting optional/third-party features. The milestones above pair each with an independent semantic check and a tracked scope entry.

A work slice is complete only when its source/macros/options are accounted for, its invariants are documented, small independent semantics agree, relevant permanent fixtures pass, fuzz seeds pass, and limitations are explicit. A release of the Ivy slice additionally passes real caller/witness tests. A claim of the full ABC port requires the complete source/configuration ledger closed, including alternate solvers, BDDs, disabled modules, auxiliary files and replacements; otherwise label the release by its supported subset.

Do not estimate calendar duration from physical line count. After M1–M3, measure representative translation/review/testing throughput and dependency expansion to estimate subsequent milestones. Keep algorithms recognizable until validated. Broader refactoring or extracting more packages is optional future maintenance, never a prerequisite to making forward progress.

## Appendix A. Survey inventory by source directory

The table below was generated from this checkout's physical source files. “Lines” includes comments and generated tables; it is a sizing aid, not an effort estimate or build-selected count. A complete per-file/symbol/configuration ledger remains an M0 deliverable.

| Source directory | C files | Headers | C++ files | Lines |
| --- | ---: | ---: | ---: | ---: |
| `src/aig/aig` | 36 | 1 | 0 | 20,234 |
| `src/aig/gia` | 105 | 3 | 0 | 95,514 |
| `src/aig/hop` | 12 | 2 | 0 | 3,990 |
| `src/aig/ioa` | 3 | 1 | 0 | 1,297 |
| `src/aig/ivy` | 25 | 2 | 0 | 16,075 |
| `src/aig/miniaig` | 0 | 4 | 0 | 2,085 |
| `src/aig/saig` | 32 | 1 | 0 | 18,807 |
| `src/base/abc` | 24 | 2 | 0 | 25,743 |
| `src/base/abci` | 87 | 0 | 0 | 112,911 |
| `src/base/acb` | 10 | 3 | 0 | 9,243 |
| `src/base/bac` | 17 | 2 | 0 | 7,949 |
| `src/base/cba` | 9 | 3 | 0 | 8,868 |
| `src/base/cmd` | 10 | 2 | 0 | 6,171 |
| `src/base/exor` | 6 | 1 | 0 | 3,959 |
| `src/base/io` | 31 | 2 | 0 | 21,787 |
| `src/base/main` | 8 | 4 | 0 | 2,770 |
| `src/base/pla` | 9 | 1 | 0 | 3,087 |
| `src/base/test` | 1 | 0 | 0 | 70 |
| `src/base/ver` | 6 | 1 | 0 | 4,344 |
| `src/base/wlc` | 20 | 1 | 0 | 17,683 |
| `src/base/wln` | 8 | 1 | 0 | 2,708 |
| `src/bdd/bbr` | 5 | 1 | 0 | 2,477 |
| `src/bdd/cas` | 2 | 1 | 0 | 1,843 |
| `src/bdd/cudd` | 62 | 2 | 0 | 73,555 |
| `src/bdd/dsd` | 6 | 2 | 0 | 3,909 |
| `src/bdd/epd` | 1 | 1 | 0 | 1,543 |
| `src/bdd/extrab` | 11 | 1 | 0 | 13,424 |
| `src/bdd/llb` | 25 | 2 | 0 | 12,612 |
| `src/bdd/mtr` | 2 | 2 | 0 | 1,468 |
| `src/bdd/reo` | 9 | 1 | 0 | 3,472 |
| `src/bool/bdc` | 5 | 2 | 0 | 2,760 |
| `src/bool/dec` | 5 | 1 | 0 | 2,011 |
| `src/bool/deco` | 0 | 1 | 0 | 714 |
| `src/bool/kit` | 15 | 2 | 0 | 11,569 |
| `src/bool/lucky` | 7 | 2 | 0 | 3,186 |
| `src/bool/rpo` | 1 | 2 | 0 | 741 |
| `src/bool/rsb` | 2 | 2 | 0 | 1,018 |
| `src/map/amap` | 13 | 2 | 0 | 6,141 |
| `src/map/cov` | 8 | 2 | 0 | 4,293 |
| `src/map/fpga` | 14 | 2 | 0 | 6,541 |
| `src/map/if` | 28 | 2 | 0 | 19,607 |
| `src/map/mapper` | 18 | 2 | 0 | 9,163 |
| `src/map/mio` | 8 | 3 | 0 | 5,830 |
| `src/map/mpm` | 12 | 3 | 0 | 5,280 |
| `src/map/scl` | 11 | 5 | 0 | 12,077 |
| `src/map/super` | 5 | 2 | 0 | 2,765 |
| `src/misc/avl` | 1 | 1 | 0 | 712 |
| `src/misc/bar` | 1 | 1 | 0 | 273 |
| `src/misc/bbl` | 1 | 1 | 0 | 1,803 |
| `src/misc/bzlib` | 7 | 2 | 0 | 5,202 |
| `src/misc/espresso` | 40 | 7 | 0 | 15,444 |
| `src/misc/extra` | 22 | 1 | 0 | 14,434 |
| `src/misc/hash` | 0 | 5 | 0 | 1,417 |
| `src/misc/mem` | 1 | 2 | 0 | 951 |
| `src/misc/mvc` | 16 | 1 | 0 | 4,920 |
| `src/misc/nm` | 2 | 2 | 0 | 778 |
| `src/misc/parse` | 3 | 2 | 0 | 1,285 |
| `src/misc/st` | 2 | 2 | 0 | 1,468 |
| `src/misc/tim` | 6 | 2 | 0 | 2,157 |
| `src/misc/util` | 9 | 10 | 0 | 9,396 |
| `src/misc/vec` | 0 | 15 | 0 | 10,940 |
| `src/misc/zlib` | 15 | 11 | 0 | 12,263 |
| `src/opt/cgt` | 5 | 2 | 0 | 1,716 |
| `src/opt/csw` | 5 | 2 | 0 | 1,286 |
| `src/opt/cut` | 11 | 3 | 0 | 5,673 |
| `src/opt/dar` | 11 | 2 | 0 | 17,093 |
| `src/opt/dau` | 15 | 2 | 0 | 13,115 |
| `src/opt/dsc` | 1 | 1 | 0 | 616 |
| `src/opt/fret` | 4 | 1 | 0 | 4,404 |
| `src/opt/fsim` | 6 | 2 | 0 | 1,933 |
| `src/opt/fxch` | 4 | 1 | 0 | 2,147 |
| `src/opt/fxu` | 12 | 2 | 0 | 5,835 |
| `src/opt/lpk` | 12 | 2 | 0 | 4,878 |
| `src/opt/mfs` | 12 | 2 | 0 | 4,427 |
| `src/opt/nwk` | 16 | 3 | 0 | 7,391 |
| `src/opt/res` | 9 | 2 | 0 | 3,778 |
| `src/opt/ret` | 8 | 1 | 0 | 3,174 |
| `src/opt/rwr` | 8 | 1 | 0 | 3,011 |
| `src/opt/rwt` | 3 | 1 | 0 | 1,346 |
| `src/opt/sbd` | 10 | 2 | 0 | 6,083 |
| `src/opt/sfm` | 10 | 2 | 0 | 6,070 |
| `src/opt/sim` | 10 | 1 | 0 | 3,216 |
| `src/phys/place` | 11 | 4 | 0 | 4,684 |
| `src/proof/abs` | 18 | 3 | 0 | 13,420 |
| `src/proof/acec` | 21 | 2 | 0 | 10,427 |
| `src/proof/cec` | 20 | 2 | 0 | 13,914 |
| `src/proof/dch` | 10 | 2 | 0 | 3,102 |
| `src/proof/fra` | 18 | 1 | 0 | 11,546 |
| `src/proof/fraig` | 13 | 2 | 0 | 7,500 |
| `src/proof/int` | 11 | 2 | 0 | 2,957 |
| `src/proof/int2` | 4 | 2 | 0 | 1,250 |
| `src/proof/live` | 9 | 0 | 0 | 7,532 |
| `src/proof/pdr` | 12 | 2 | 0 | 7,757 |
| `src/proof/ssc` | 5 | 2 | 0 | 1,975 |
| `src/proof/ssw` | 21 | 2 | 0 | 11,272 |
| `src/sat/bmc` | 34 | 1 | 0 | 22,665 |
| `src/sat/bsat` | 15 | 9 | 0 | 17,237 |
| `src/sat/bsat2` | 0 | 17 | 7 | 5,688 |
| `src/sat/cnf` | 10 | 1 | 0 | 9,018 |
| `src/sat/csat` | 1 | 1 | 0 | 1,004 |
| `src/sat/glucose` | 0 | 20 | 6 | 8,213 |
| `src/sat/glucose2` | 0 | 22 | 6 | 9,152 |
| `src/sat/lsat` | 0 | 1 | 0 | 132 |
| `src/sat/msat` | 14 | 2 | 0 | 5,478 |
| `src/sat/proof` | 1 | 1 | 0 | 1,333 |
| `src/sat/psat` | 0 | 2 | 0 | 60 |
| `src/sat/satoko` | 3 | 19 | 0 | 4,315 |
| `src/sat/xsat` | 3 | 8 | 0 | 3,118 |

## Appendix B. Local evidence and reading map

Use the local files as the primary reference for this snapshot rather than assuming behavior from a newer upstream version:

* Build and lifecycle: `Makefile`, `CMakeLists.txt`, `Makefile.emscripten`, every `module.make`, `src/base/main/mainInit.c`, `mainFrame.c`, `main.h`, `src/base/cmd/`, and `abc.rc`.
* Core representations: `src/misc/util/abc_global.h`, `utilCex.h`, `src/misc/vec/`, `src/aig/aig/aig.h`, `src/aig/saig/saig.h`, `src/aig/gia/gia.h`, `giaAig.h`, `src/base/abc/abc.h`, and `readmeaig`.
* First verification slice: `src/aig/ioa/ioaReadAig.c`, `src/base/io/` AIGER readers/writers, `src/sat/bsat/satSolver.h` and implementation, `src/sat/cnf/cnf.h` and implementation, `src/opt/dar/darInt.h`, and `src/proof/pdr/pdr.h`, `pdrInt.h`, and implementation.
* Compatibility fixtures: `../../ivy/ivy_mc.py`, especially `ABCModelChecker`, `check_isolate` and witness decoding; `wasm-smoke-test.js`, `wasm-aiger-compare.js`, `src/demo.c`, `i10.aig`, and `README.md`.
* Dependency policy: `pkg/go.mod`, `pkg/go.sum`, and the cached v0.14.6 insdict source/API. The exact cached version was inspected for `Dict`, typed constructors and accessors; no dependency upgrade is needed for this plan.

This is a structural survey with targeted implementation/header and caller inspection, not a line-by-line correctness audit of all 959,292 lines. No oracle binaries were rebuilt or benchmarked, and no algorithm parity is claimed at the planning stage.

## Appendix C. Initial execution evidence

After completing the plan, the updated task goal requested implementation as well. Initial M0 support artifacts were added while retaining the no-Go-source boundary of the original planning request:

* `tools/source_manifest.py` generates `test_vectors/oracle/source_manifest.json`, recording content hashes, byte/line counts and textual module membership. It does not execute make or infer active build configurations. The per-symbol and progress ledger remains outstanding.
* The original native smoke driver was temporary planning scaffolding. It has
  been removed; its frozen data is integrity-checked but is not used as runtime
  evidence. Current witness tests use Go replay and certificate checks.
* The installed Go toolchain reports `go1.27.3 linux/amd64`, matching `go.mod`. No Go build/test/fuzz campaign has been run because there is no Go implementation yet.

Refresh the planning-only source inventory from the repository root while the
original source tree is still present:

```
python3 pkg/tools/source_manifest.py > pkg/test_vectors/oracle/source_manifest.json
```

Source-manifest output is deterministic for unchanged input files. It is a port
coverage aid, not a test dependency. The remaining M0 work includes the
configuration/symbol coverage ledger and broader reviewed fixtures. Later
milestones remain incomplete.

## Appendix D. Implementation updates and user conventions

The current implementation state and verified source coverage are recorded in
[PORT_STATUS.md](PORT_STATUS.md). Appendix C describes the initial planning
baseline, not the current implementation. The full roadmap and completion gates
remain in force; the initial Go slice does not constitute the full ABC port.

Permanent reviewed vectors, stand-alone model evidence and minimized regressions
belong in `test_vectors/`, per the user's instruction. Go's automatic
`testdata/fuzz/` output must be promoted explicitly into permanent vectors and
ordinary regression tests before relying on it for retained coverage. The
entire Go suite must work after the original ABC source tree and all C build
tools are removed. The Go module, including `pkg/tools`, must contain no C or
C++ helper source. The version-2 dataset manifest records only fixture hashes
and policy; package tests reject missing/untracked data, reintroduced native
helper source and permanent-vector directories named `testdata`. See
`test_vectors/README.md`.

Per the user's source-format instruction, Go files begin directly with their
`package` declaration. Keep useful C source-reference comments after the package
statement, and keep generated headers out of package documentation.


Mapping transfer now uses `GiaLUTMapping.Transfer` with an explicit owned-by-
caller object correspondence instead of mutable graph Values. Preserve native
source-root iteration, skipped constant/dropped roots, removed dropped fanins,
phase-discarded fanin/footer IDs, signed mux footers and last-root replacement
with stale packed records. Reject complemented surviving roots, invalid target
IDs, non-topological fanins and dropped footers of surviving LUTs. Cap complete
packed output, including stale records. Test native packed parity and caller
ownership; equivalence of the correspondence needs separate graph validation.
Packing/timing transfer and mapping selection remain distinct pending work.


Packing storage/transfer now uses concrete immutable `GiaLUTPacking` tied to
its source mapping. Preserve native `[group count, size, IDs..., ...]` records
with group sizes 1–3, repeated members, group/slot order and phase-discarding
transfer. Validate exact group counts/record lengths and source/target LUT
membership before exposing data. Dropped packed members reject; target aliases
remain valid as in C. Return statistics without printing and preserve the empty
maximum-size sentinel -1. Cap full storage, test caller/export ownership and
cancellation, and compare native transferred records plus independent slot
substitution. Packing selection/conversion and timing transfer remain pending.


Native two-slot edge storage is now represented by owned `GiaEdgeAssignment`.
Preserve zero as empty sentinel, duplicate/self-edge slot use and independent
directed overflow (which can create asymmetric rows). Export pairs by scanning
objects and slots for nonzero lower neighbors; do not silently repair symmetry
or deduplicate. Packing groups of two add one undirected edge; groups of three
add first-to-third then second-to-third edges in native direction order.
Return skipped directed insertions without printing. Test complete native rows,
exported pair order, repeated packing members, sentinel zero, exact work caps
and cancellation. Delay evaluation and edge optimization remain separate work.


Native no-box edge timing now uses `GiaEdgeAssignment.EvaluateDelay` with
plain packed mappings. Preserve root-directed slot lookup (including sentinel
zero), delays 2/10, at most four fanin slots, zero initial unmapped objects and
zero CO delay-array entries. The returned maximum scans CO drivers, excluding
dangling LUTs. Edge count truncates positive directed slots divided by two,
even for asymmetric rows. Return owned delays, bound total evaluated fanin
slots and check addition. Compare full arrays and maxima to C and an independent
recursive model. Box timing and mapping2 must remain explicit pending paths.


Greedy no-box edge selection now uses `GiaLUTMapping.SelectEdges`. Preserve
first-level skipping, fanin-order running maximum/runner-up tracking, maximum
slot multiplicity, optional two-critical-slot selection and fanin capacity
checks. Duplicate slots count separately, and directed insertions retain the
native overflow policy. Compare full native rows and delay arrays, not just
the output maximum. Test distinct two-way ties, fanout capacity saturation,
deep chains, empty/oversized LUT rejection and bounded cancellation. Independently
reevaluate selected rows through the timing evaluator. Timing boxes, mapping2
and the window optimizer remain pending.


Reverse edge selection now uses `GiaLUTMapping.SelectEdgesReverse` and
`GiaReverseEdgeTiming`. Preserve descending LUT order, zero-initialized
reverse unit distances without CO seeds, critical-fanout tie multiplicities,
one/two-fanout capacity checks, duplicate-slot insertion and distance propagation
to every mapped fanin. Maximum is over CIs and includes dangling mapped logic.
Keep this result distinct from forward 2/10 timing. Compare native full rows
and distances; independently propagate weighted fanout paths (selected edge
weight zero, ordinary weight one). Test deep no-CO chains, exact work caps
and cancellation. Window optimization and alternate/timing-box paths remain
pending.


The immutable alternate mapping prerequisite is now `GiaLUTNetwork`, converted
from packed mappings with native four-slot bounds and ordered, duplicate-
preserving fanin/fanout lists. Empty fanin lists are not LUTs in this
representation. `EvaluateEdges` computes forward/reverse unit costs (0 for
marked edges, 1 otherwise), with maxima over all mapped LUTs including dangling
logic; CI reverse entries remain zero. `CriticalLevels` groups F+R==Maximum
by forward distance and safely checks C's sum/nonempty-positive-level assertions.
Compare full native fanouts, arrays and groups. Count both passes against the
work budget, preserve caller ownership, and independently check recursive path
maxima. Mutable window updates and the actual improvement loop remain pending.


Native unit-edge improvement is now `GiaLUTMapping.OptimizeEdges`. Preserve
empty initial rows, edge limits one/two, critical fanin-before-fanout attempts,
capacity/assertion checks, increasing critical-level population and original
level order on ties (made explicitly stable in Go, matching the pinned C
reference). Keep strict-delay backup saves, the 50-iteration stagnation rule
and 10,000-iteration native cap. Return the restored backup, not the last
plateau assignment. Charge construction, both timing passes, candidate scans
and final backup reevaluation against shared work. Caller iteration caps and
cancellation must return no partial result. Compare complete restored rows
and both timing arrays to C; fuzz reevaluation, capacity and delay nonregression.
Mutable window update/evaluation APIs and timing-box integration remain pending.


Window replacement/full evaluation now uses `GiaLUTNetwork.WithWindow` and
`EvaluateWindow`, with network `SelectEdges` sharing the packed greedy engine.
Empty lists remove LUTs; unique nonconstant roots, topological fanins and
four-slot widths are checked. Keep the source immutable instead of temporarily
swapping shared lists. Rebuild derived fanouts for a consistent proposal (C's
full window evaluator only consumes replacement fanins). Explicit network
context represents mapping2-only behavior: C prefers packed mappings when
both representations remain present. Storage and selection work budgets apply
per pass. Compare complete native rows/timing and original restoration; fuzz
fanouts against independent rebuilding. Proposed Boolean equivalence still
requires a separate check. Incremental window evaluation remains pending.


Directed edge mutation now uses atomic owned `GiaEdgeAssignment.WithEdits`
batches. Preserve native insertion/overflow, first-slot removal compaction,
second-slot clearing and zero/duplicate semantics. Reciprocal rows do not
change implicitly. Missing-neighbor removal rejects the C assertion precondition.
Bound arena/edit counts, preflight skipped-count overflow and discard the
proposal on any invalid edit or cancellation. Exhaustive native slot vectors
and randomized sequences must verify the source remains unchanged after a
later edit fails. These primitives prepare incremental timing/window work.


Incremental packed-mapping recomputation now uses
`GiaLUTMapping.RecomputeEdges`. Preserve reversed structural TFO postorder,
skipped unmapped gates, retained prior delays, and per-LUT row clearing without
removing stale reciprocal entries elsewhere. Report the native affected-LUT
maximum separately from the global CO maximum. Copy prior state and reject
invalid dimensions/negative delays, bounds or cancellation without exposing
partial updates. Share the per-node greedy kernel with full evaluation and
pin complete C incremental state, not just timing maxima. This explicit packed
context matches C's representation priority; it does not imply that temporary
mapping2 replacements alter packed records. Mapping2-aware incremental proposal
integration and timing boxes remain pending.


Timing-manager foundations now use concrete owned `TimManager` terminal
storage. This first layer has no boxes and treats all CIs/COs as PIs/POs.
Preserve exact float32 bits, native eternity, default -1 associations, shared
traversal marks and epoch-zero setter distinctions: repeated CO required writes
are allowed at zero while CI required/CO arrival writes reject. Preserve
enable/disable checks and 30-bit traversal rollover. Bulk required-time writes
must preflight all terminals before mutation. Check native terminal states
including signed zero, infinities, NaN payloads and subnormals, plus randomized
operation sequences/clone ownership. Boxes, delay-table evaluation and timing-
box graph integration remain separate required layers; do not claim the native
boxed GetCiArrival/GetCoRequired paths from these no-box getters.


Timing-box records/associations now use owned `TimBox` entries in `TimManager`.
Box inputs are CO ranges; outputs are CI ranges. Preserve insertion-order
overwrite for overlapping associations, pin positions, table/copy/black
metadata and current/previous pin labels (including -1 at epoch zero). Reject
nonpositive box dimensions and out-of-range/cap violations before mutation.
Native primary counts use first/last box boundaries, while bulk initialization
scans all unassociated terminals; test these separately, including overlaps.
Clones must own box records and terminals. Boxed timing getters remain
explicitly unsupported until delay-table propagation is implemented.


Owned sparse timing delay tables and boxed arrival propagation are now
implemented. Preserve output-major storage, disabled -1e9 arcs, dimensions,
aggregate table/value caps and deep cloning. Boxed arrival uses current-epoch
cache entries (including epoch zero), stored input arrivals even if marks are
stale, all-output marking and C's integer maximum helper: sums are float32,
then truncate toward zero to int32. Reject nonfinite/out-of-int32 conversion
before changing any output or box marks. Table changes require epoch advance
to invalidate cached outputs. Pin reviewed compatibility fractional/negative/disabled cases
and separately test C-undefined inputs for safe rejection.

Required-time propagation needs special source auditing: `Tim_ManGetCoRequired`
checks `pDelays[k]` while iterating output k, but subtracts `pDelays[i]` for
input i. Thus the disabled-arc guard can address a different input column and
can read beyond an input row when outputs exceed inputs. Do not assume this
is the intended delay-matrix semantics or use undefined native reads as a
correctness oracle. Preserve regression evidence and establish explicit safe
required-time semantics before integrating graph timing boxes. The Go path now uses corrected per-arc guarding, as described below.


Boxed required-time propagation is now implemented with the corrected actual
input-arc guard. Preserve native epoch-zero/current caching, stored output
requirements, float32 subtraction, comparison-based Abc_MinFloat behavior
(including NaN/tie selection), all-input marking and owned manager state.
The original native routine is invoked only for safe shapes in regression
oracles: 80 defined cases, 11 differing results due to the guard bug. Compare
all 128 shapes to a separately computed corrected C matrix reference, including
48 shapes that would permit out-of-row native guard reads. Keep both native
and corrected vectors so this intentional behavioral fix stays explicit. Test
caching, rectangular shapes, no-arc columns, clone independence and atomic
shape failure; fuzz against independent finite column minima. Graph timing
integration remains pending.


GIA timing-box order is now `GiaManager.OrderWithBoxes`, with explicit
normalized-layout and contiguous timing-partition checks. Preserve native
constant/PI staging, input-cone postorder, all box input COs before output CIs,
and primary-output tail. Omit dangling gates. Iterative private marks must
reject future-box CI dependencies without partial order or source mutation.
Include buffer/XOR/mux-control dependencies safely; choice-sibling traversal
remains pending. Pin exact native orders and ordinary dependency failures.
The native diagnostic for a final-box future CI uses `Tim_ManBoxFindFromCiNum`'s
-2 result as a box index and aborts: exclude those native calls, but test Go's
safe rejection explicitly. Bound returned entries and exercise deep cones,
metadata overlap/gap rejection and cancellation before graph timing integration.


Boxed graph edge-delay evaluation now uses
`GiaEdgeAssignment.EvaluateDelayWithBoxes` and returns an owned updated
`TimManager`. Preserve staged order, epoch advance, checked CI arrival-to-int32
truncation, signed 2/10 LUT arithmetic, CO-driver arrival updates and zero
graph CO delay entries. Dangling LUTs remain unevaluated. Without boxes,
native fallback ignores terminal overrides and leaves epochs unchanged.
Charge copied delay values, order entries, mapped fanin slots and actual box
arcs to a shared budget. Compare full native delay arrays and manager arrival
bits, and check atomic cancellation/ownership. Boxed edge selection, mapping2
evaluation and broader timing integration remain pending.


Boxed greedy edge selection now uses `GiaLUTMapping.SelectEdgesWithBoxes`,
sharing staged traversal with boxed evaluation and the per-node native greedy
kernel with unboxed/incremental selection. Preserve empty starting rows,
signed CI truncation, possible PI-critical edges, tie/capacity rules and cloned
manager updates. Check C's nonnegative-candidate assertion explicitly when
negative PI arrivals leave all candidates below zero. Pin full selected rows,
graph timing and terminal arrival bits for defined native cases; separately
test assertion/undefined arithmetic rejection. No-box fallback ignores
terminal overrides and leaves epochs unchanged. Reevaluate selected edges
and fuzz capacity/nonregression/ownership. Mapping2 boxed paths remain pending.


Alternate-mapping boxed evaluation/selection now use
`GiaLUTNetwork.EvaluateDelayWithBoxes` and `SelectEdgesWithBoxes`. Preserve
empty-list non-LUT semantics via private adapters to shared kernels. Native
oracles must clear packed mapping after conversion so they exercise mapping2
branches, rather than accidentally rechecking packed priority. Compare full
rows, timing and terminal float bits; test removed LUTs, duplicate fanins and
window proposals against independently rebuilt mappings. Adapter storage and
boxed evaluation retain their documented work bounds and source ownership.
Mapping2 incremental integration and reverse graph timing remain pending.


SAT-LUT unboxed timing now uses `GiaLUTNetwork.UnitTiming`: zero CI arrivals,
unit LUT delays, delayStart lower bound and a maximum over all mapped LUTs,
including dangling logic. Initialize required arrays to 1e9, downdate CO
drivers by the computed maximum, then propagate required-1 through every
mapped LUT, including unobservable cones. Preserve untouched CO entries.
Pin full native arrays and use independent recursive forward-path and reverse
fanout-minimum models. Bound both passes together and test deep chains.

Before porting boxed `Sbl_ManCreateTiming`, correct/audit its traversal-index
reuse: the reverse outer `Gia_ManForEachObjVecReverse(..., k)` contains an inner
`Vec_IntForEachEntry(..., k)`, overwriting the outer position. This is separate
from the already-corrected delay-table guard bug. Preserve native regression
evidence and compare a corrected boxed reference rather than silently trusting
skipped/repeated traversal. Boxed SAT-LUT timing now uses the corrected implementation described below.


Boxed SAT-LUT timing now uses `GiaLUTNetwork.UnitTimingWithBoxes` with a
cloned manager and two traversal epochs. Preserve native visited-LUT maximum/
delayStart behavior, integer CI/CO conversions, CO-driver required assignment,
float terminal values and zero/eternity graph CO entries. Use separate reverse
traversal/fanin indices and corrected per-arc guards. Retain the original
native required arrays to demonstrate counter-clobber regressions, and compare
complete state against a corrected C routine with both fixes explicit. Bound
copied tables, both staged passes, fanin slots and actual box arcs together;
validate fallback epoch behavior and reject invalid conversions atomically.
Fuzz independent boxed recurrences, disabled arcs, fractional requirements and
source ownership. SAT remapping and broader integration remain pending.


Decoded unit-window timing now uses `GiaLUTNetwork.EvaluateUnitWindow`.
Keep original required constraints/reference maximum and recompute only the
provided nodes in proposal topological order, including empty-list zero
arrivals. Scan roots in caller order; on the first late root follow the first
critical fanin at every step and emit negated node-selection variables indexed
by window order, independent of GIA IDs. Reject missing/constant critical
boundaries that violate native assertions. Bound original timing, proposal
slots and path scans together. Pin reviewed compatibility decisions, full arrays and clause
order, test reordered independent nodes and feed clauses into the concrete
SAT solver. Decoding cuts, proving Boolean/closure validity, edge-mode timing
and the complete remapping search remain pending.


Selected cut decoding now uses `GiaWindowCut` and `DecodeGiaWindowCuts`.
`Node` is a destination window ordinal; `Leaves[0/1]` correspond to native
CutI1/I2 and `Nodes[0/1]` to CutN1/N2. Expand lower then upper leaf bits, then
lower then upper node bits, ascending within each word. Preserve raw IDs,
repeated values and later-selected-cut replacement for a destination. Only
selected records are interpreted; unselected sentinel entries remain legal.
Reject empty solutions and out-of-range selected indices/destinations/bits
safely. Budget all input entries and every expanded bit, including overwritten
cuts. Decoding alone does not establish topology, width, closure or Boolean
validity. Pin reviewed compatibility bit-boundary/full-mask vectors and integrate decoded lists
with timing-conflict clauses. Candidate generation/CNF/search remain pending.


Structural SAT-window clauses now use `BuildGiaWindowCNF`. Reserve node
variables first, preserve the caller's auxiliary prefix and allocate a variable
for every cut record, including unused leaf and final-unit records. For each
node emit its negative literal followed by positive proper-cut literals;
exclude the final unit entry. Each proper cut implies its root, then each
internal fanin in low/high ascending mask order. Preserve native literal and
clause order before solver normalization. Do not add an at-most-one constraint:
the native formulation allows multiple proper cuts to be selected. Unit-record
metadata is unused here. Bound variables, clauses and aggregate literals and
cancel without exposing partial artifacts. Pin reviewed compatibility clauses, exhaustively
check small assignment semantics and verify multiple-cut satisfiability with
the concrete solver. Candidate validity, polarity hints, root/cardinality
constraints and the complete remapping loop remain separate pending work.


Cardinality networks now use `BuildGiaCardinalityCNF`, porting the native
pairwise recursion in `giaSatMap.c`. Reserve power-of-two inputs first;
allocate two variables per comparator. Preserve native half-sorter forward
implications (OR-to-high and AND-to-low), recursive call order and extra merge
implications. Outputs are threshold variables: negating output k constrains
at most k selected inputs. They are not exact equivalence signals. Preflight
variable storage, cap total clauses/literals and check cancellation without
returning partial artifacts. Pin reviewed compatibility variable allocation/output/clause
order and independently solve all small assignments at every threshold.
Connect padded input assumptions, root requirements, concrete solver
lifecycle/bookmarks and incremental bounds in the remaining remapper work.


`GiaWindowSolver` now connects fixed cut-selection windows to the concrete
Go SAT solver. Build cardinality first and use its complete variable count as
the structural cut-variable prefix. Enforce aggregate clause/literal budgets
across both builders. Own roots and compiled clauses; caller mutations must
not affect repeated queries. Preserve native assumption order: negative
threshold output, negative unused padded inputs, positive roots including
duplicates. A capacity-wide bound requires no threshold. Retain learned
clauses between queries and test both tightening and loosening after UNSAT.
Expose complete SAT results and ascending selected node/cut indices only on
SAT; do not claim a particular model matches C heuristics. Pin reviewed compatibility status
and assumption sequences, then test cancellation and independent chain-size
expectations. Candidate generation/validation, polarity preferences, timing
conflict feedback, native rollback and graph replacement remain required
before claiming complete SAT remapping.


`GiaManager.GenerateWindowCuts` now derives SAT-window candidates from an
explicit ordered leaf boundary and topological internal-node list. Limit both
to 128 unique disjoint IDs; require every structural fanin to have an earlier
window ordinal. Reject native muxes because the source two-fanin routine omits
the control dependency. Preserve leaf singleton/sentinel records, outer-first
Cartesian enumeration, existing-subset rejection, stable removal of supersets,
and the final internal unit cut. Native requested widths 1..4 use feasibility
four and 5..6 use six; preserve this unusual behavior explicitly. Count pair
merges and dominance comparisons against a shared work budget; cap retained
records plus transient candidates and return no partial result. Pin complete
native sequences through both mask halves, verify emitted graph boundaries
independently and solve generated windows. Mapping-root derivation, current
mapping cut lookup/polarity, timing feedback and graph updates remain required
for full remapping.


`GiaLUTNetwork.WindowMappingState` now derives root variables and current
mapping preferences from generated cuts. Skip CI roots but preserve remaining
root order/duplicates and require mapped internal membership. Keep every
mapped window node in the initial solution even when exact cut lookup fails.
Build four masks from current fanins, deduplicate through bits, preserve the
native skip for unset external fanins, and choose the first exact mask match
including unit records. This lookup is not closure/equivalence validation.
Pin complete native root/initial/preference arrays for both matching and
missing cuts. `GiaWindowSolver.SetMappingPreferences` translates node/cut
ordinals into SAT variables. `SatSolver.SetPreferredVariables` resets all
saved phases then marks preferred variables true; validate inputs first so
errors leave phases intact. Test actual phase bits and generated-window SAT.
Timing-conflict feedback, Boolean/closure validation, candidate application
and complete remapping remain pending.


`GiaLUTNetwork.SearchUnitWindow` now connects explicit-window candidate
generation, root/current mapping state, phase hints, cardinality/structural
CNF, selected-cut decoding and unit timing feedback. Use a private solver and
return owned proposals; do not mutate graph or mapping during search. Begin
at the current mapped-node count, retry the same bound after a late proposal,
append the native negative node-path clause, and tighten only after timing
acceptance. Retain the best accepted proposal when a later solve is UNSAT or
unknown; expose the native default ten-iteration cutoff separately. Preserve
native selected-cut-count improvement checks. SAT budgets apply per solve,
analysis work caps per decode/timing operation, and bounded iterations limit
the combined search. Propagate cancellation throughout preparation and solves.
Native node-path exclusions are heuristic: an UNSAT result after timing
exclusions is not a proof of timing-feasible optimality because cuts may
change while node selections remain fixed. Do not expose it as such. Replay
native timing fixtures through the real feedback loop with fixed candidate
cuts, independently verify accepted generated proposals and test saved state,
iteration limits and source immutability. The current search covers unboxed
four-input unit delay. Window discovery, boxed/edge timing, Boolean/closure
validation and atomic application remain required for the complete remapper.


`GiaLUTNetwork.ValidateFunctions` now certifies complete mapped dependencies,
CO-driver coverage and structural boundaries for every mapped LUT, including
dangling logic. All nonconstant/non-CI mapped fanins must themselves be mapped;
unit timing does not implicitly propagate through an unmapped buffer. Derive
packed truth words from the source graph with private epoch-indexed scratch
and iterative traversal. Preserve last-slot duplicate semantics, ignore constant
zero as a symbolic variable, and include phases, buffers, XORs and mux control
cones. Bound arena scans, fanin slots and aggregate cone traversal together.
Because these functions come from complete structural cuts, composing the
closed mapping preserves graph outputs without trusting externally supplied
truth tables. Verify composition exhaustively on small random graphs.

`WithValidatedUnitWindow` returns an owned replacement after validating both
source and proposal, rebuilding fanouts and comparing complete global unit
maxima. This catches escaping-root timing regressions even when a root list is
incomplete. Limits apply separately to each validation/storage/timing phase.
All failures expose no replacement. `RemapUnitWindow` now connects search to
checked replacement construction when an improvement exists. Pin actual native
UpdateMapping fanin/reference arrays and native truth words, and verify caller
mutation cannot alter returned storage. This explicit-window unit remapper
still requires automatic window discovery, native lifecycle/scheduling and
boxed/edge paths before claiming complete SAT-LUT feature parity.


`GiaLUTNetwork.DiscoverWindow` now ports native pivot window growth from
`giaSplit.c`. Distinguish selected mapped roots (`Nodes`) from the union of
their structural cone gates (`Ands`). Build static structural fanouts in object
order and mapped reference counts including duplicated slots and CO edges.
Temporarily dereference selected LUT fanins when selecting candidates. Preserve
zero-external-reference input priority, first maximum-overlap ties, reverse
odd-window fanout preference, the >=5 external-reference skip and first
single-fanout shortcut. Stop at the first candidate exceeding the remaining
AND limit. Final leaves come from structural boundary edges; roots are cone
gates with remaining external mapped references. Sort every returned vector.
Use private marks/reference storage and iterative traversals; reject native
muxes because source cone collection omits controls. Validate the source
mapping and bound discovery work separately. Pin complete native vectors and
rejections, then feed discovered windows to the explicit remapper and verify
replacement outputs exhaustively. Persistent manager reuse, native packed/
mapping2 conversion lifecycle, pivot scheduling and boxed/edge remapping
remain required for complete SAT-LUT parity.


`GiaLUTNetwork.RemapUnitPass` now schedules checked unit remapping over an
entire network. Scan currently mapped object IDs in ascending order rather
than snapshotting the initial pivot list: replacements may remove later pivots
or introduce later mapped roots. Preserve the native sorted-AND-vector window
hash across improvements; hash before native large/small filters, and skip
windows below ten internal gates. Respect the requested improvement cap (zero
means unlimited). Use a lookup-only map for window membership and an ordered
attempt slice for reporting; no dictionary iteration determines execution.
Return a complete owned network/functions/timing only after the entire pass
succeeds. Bound total pivot attempts and distinct stored windows, validate all
nested option dimensions even on skipped windows, and propagate cancellation.
Per-window work budgets remain separate; pivot/window caps bound their total
multiplicity. Pin full native pass counts, independently compose native and Go
mappings against the graph and verify timing/source immutability. Do not require
identical model choices or claim timing-feasible optimality. The current pass
rebuilds discovery/SAT state per pivot; native manager reuse, cardinality
bookmarks, packed/mapping2 lifecycle and boxed/edge modes remain pending.


Native mapping conversion now uses `GiaLUTMapping.AlternateNetwork` and
`GiaLUTNetwork.PackedMapping`. Preserve arbitrary stored fanin widths and
ordered duplicate slots; keep the existing four-input `Network` edge adapter
restricted. Native conversion drops empty and constant-zero records, stale
packed storage and signed footer metadata. Repack nonempty records in root-ID
order with root-ID footers. Bound the complete packed header/records before
allocation and separately cap copied fanin slots; validate topology and check
cancellation without exposing partial output. Pin reviewed compatibility whole arrays for
reordered/signed/empty/wide input and check canonical idempotence and ownership.
The internal edge selection adapter shares this checked emitter.

`GiaLUTMapping.RemapUnitPass` now joins owned packed-to-network conversion,
checked whole-network unit remapping and native canonical repacking. The
`MaxPackedEntries` option caps final packed storage; failure exposes neither a
partial mapping nor a partial pass. Test output equivalence, source immutability
and output-cap/cancellation rejection. Native destructive manager attachment,
cardinality bookmark reuse and boxed/edge modes still require implementation;
the conversion lifecycle itself is now available through owned values.


`SatSolver.Bookmark/Rollback` now support retained owned checkpoints. Normalize
to propagated root state before bookmarking; native C requires a drained queue.
Save every mutable solver slice, including each clause's literals and each watch
bucket, without recursively capturing earlier bookmarks. Rollback restores
variables, formula/learned prefix, assignments/reasons, watches/units, heap and
heuristic state, retaining the bookmark for repeated use. Nonempty-prefix work
statistics remain cumulative; empty-prefix statistics/random seed reset. The
owned snapshot safely restores prefix clauses even if later reduction deleted
them. This matches logical prefix behavior, not native memory-arena or exact
heuristic-state replay. Guard solve callbacks and clear checkpoints on Reset.
Pin reviewed compatibility queue-valid scenarios and exhaustively check assignments after
repeated contradictory/learned tails, then run solver invariant checks.

Window solvers now bookmark cardinality before adding cut variables.
`ResetWindow` compiles and validates a new window against combined prefix/CNF
budgets, stages a rollback privately and commits only on success. Reuse the same
power-of-two capacity; reset removes prior structural and learned clauses.
Test alternating chain/empty windows, UNSAT-to-SAT recovery and failed reset
preserving the previous query. Whole-pass solver reuse and persistent discovery
scratch still need wiring; snapshot allocation costs must be measured before
claiming native arena efficiency.


Whole-pass SAT reuse is now wired through a private concrete solver session.
Choose one power-of-two capacity from the configured window limit, matching
native `Sbl_ManAlloc`, and compile cardinality lazily for the first active
window. Rebind later windows with `ResetWindow`, then reapply current mapping
phase preferences. Retain the immutable prefix checkpoint while discarding
all prior structural/learned/feedback clauses. The standalone explicit-window
entry still chooses capacity from its own node count. Expose successful prefix
build/reset counters to test actual reuse rather than infer it from results.
Pin full native two-cone passes with different window sizes and verify composed
outputs; separately force an exclusion to UNSAT and prove it disappears at
reset. Preparation failures must leave the current window usable. Current
rollback uses owned snapshots, not native arena efficiency. Persistent discovery
scratch and boxed/edge scheduling remain pending.


Whole-pass discovery now uses a private persistent concrete session. Keep
structural static fanouts, terminal masks and an owned graph image fixed across
pivots; reject topology/source changes, including same-size CO patching. Reuse
arena mark/seen/reference scratch, clearing transient state before each query
so cancellation and oversized-pivot exits cannot leak marks. Cache mapped
reference prefixes by immutable network identity. A replacement network must
pass complete validation and reference counting before committing its prefix;
a rejected refresh leaves the previous mapping usable. Include graph-image
comparison, refresh work and scratch reset in the discovery budget. Revalidate
source/proposal functions on mapping changes rather than every unchanged pivot.
Expose static-build/reference-refresh/query counts to prove reuse. Replay all
native discovery vectors through one session per graph/mapping and test
cancellation recovery and invalid refresh explicitly. Whole-network tests must
show reference refreshes after real accepted replacements. SAT and discovery
session wiring is now present; native arena allocation efficiency and boxed/
edge remapping remain pending.


`GiaLUTNetwork.EvaluateEdgeWindow` now ports native edge feedback: temporarily
replace mapping2 lists on owned storage, recompute full greedy two-edge timing,
and compare the global CO-driver maximum with the explicit signed limit.
On rejection emit negative selection literals for every nonempty proposed
window entry in caller order. This differs from the unit critical-path clause;
keep the rules separate inside the shared concrete search loop. Pin reviewed compatibility
acceptance, complete delays, edge slots and clauses, including removed entries
and reordered independent nodes. Raw timing evaluation does not prove closure.

`SearchEdgeWindow` and `RemapEdgeWindow` now join candidate generation/SAT to
edge feedback and checked replacement. Validate complete source/proposal
functions, dependency/output closure and edge timing. Do not add a unit-delay
condition, which would change the edge objective. Explicit-window maximum zero
is used verbatim; automatic baseline selection from an existing assignment is
part of the remaining whole-network edge entry point. Test actual rejected
search steps, feasible first-level delay ten, zero-bound exhaustion, cancellation
and output composition. Native exclusion clauses remain heuristic. Whole-pass
edge scheduling, packed edge outputs, prior-edge baseline handling and boxed
search are still required for full edge-remapping parity.


Whole-network edge remapping now uses `GiaLUTNetwork.RemapEdgePass`, with a
packed wrapper on `GiaLUTMapping`. Share the concrete pivot schedule, repeated
window filter and persistent SAT/discovery sessions with the unit pass, while
keeping edge timing and checked replacement distinct. Maximum zero derives a
baseline by evaluating the supplied prior assignment against the original
mapping; do not substitute fresh greedy selection. Preserve owned prior rows
when no improvement is accepted. An unchanged result may remain above an
explicit infeasible limit and must not be reported as satisfying that limit.
For improvements reselect rows for the saved accepted mapping and reevaluate
its final delay before returning. Native search leaves trial edge rows globally
installed: 12 pinned no-improvement cases replace the original assignment.
Correct this state leak in the Go owned-value API and retain the native rows
as regression evidence. Pin full native final counts/baselines/mappings, verify
output composition and delay limits independently, and test packed output caps,
source ownership and cancellation. Boxed remapping and native arena efficiency
remain pending; the unboxed whole-network edge path is now connected.


Boxed edge-window feedback now uses `EvaluateEdgeWindowWithBoxes`, sharing
the concrete edge evaluator with an optional timing manager. Evaluate the
owned proposed mapping with staged `SelectEdgesWithBoxes`; box output arrivals
must reflect changed upstream LUTs before evaluating downstream LUTs. Return
an owned timing clone/order with edge rows and the native all-nonempty-node
exclusion. Preserve no-box fallback semantics (ignore arrival overrides and
leave epochs unchanged). Pin reviewed compatibility acceptance/clauses, full edges/delays and
terminal float bits for chained boxes, buffers and collapsed windows.

`SearchEdgeWindowWithBoxes` and `RemapEdgeWindowWithBoxes` now connect this
feedback to the shared SAT loop and checked replacement. Keep an explicit
signed delay maximum, clone timing state per candidate and validate complete
source/proposal structural functions and dependency/output closure before
exposing mapping, edges or timing updates. Check accepted output composition,
source immutability and cancellation. Do not route boxed edge search through
unit-delay constraints. Whole-network boxed baseline handling/scheduling and
boxed unit-window feedback with corrected required propagation remain pending.


Boxed unit feedback now uses `EvaluateUnitWindowWithBoxes`: first compute
corrected boxed original timing, then recompute native local candidate arrivals
against its required constraints. Preserve fixed original box-output CI arrivals
during local evaluation, first violated root and first-critical-fanin clauses.
The returned manager represents original propagation, not an applied proposal.
Budgets apply separately to boxed original timing and local candidate work.
Pin full corrected C arrays/clauses/terminal bits; retain original native bug
regressions separately so the reference correction is explicit.

`SearchUnitWindowWithBoxes` and `RemapUnitWindowWithBoxes` now join this feedback
to concrete SAT and checked owned replacement. Application must independently
recompute complete staged original/proposal timing and compare global maxima:
local evaluation with fixed box outputs or incomplete root lists alone cannot
certify global delay. Validate structural functions/dependency/output closure
and return the proposal's owned timing manager only after every check succeeds.
Test explicit global timing regression rejection, accepted output composition,
source immutability, cancellation and no-box fallback. Native local exclusions
remain heuristic. Whole-network boxed scheduling and baseline propagation are
the remaining integration layer.


Whole-network boxed scheduling is now connected through
`RemapUnitPassWithBoxes` and `RemapEdgePassWithBoxes`, including packed wrappers.
Reuse the same pivot order, window hashes and persistent SAT/discovery sessions
as unboxed passes. Keep a private current timing manager and move mapping,
accepted edge rows and timing state together after each checked improvement.
Unit mode uses corrected required propagation and performs a final staged
comparison with the original maximum. Edge mode derives automatic zero limits
from the supplied prior rows through boxes and reevaluates the final assignment
against the returned mapping. Preserve prior rows when unchanged; trial timing
must never mutate the caller manager. Limits remain per phase/window with total
pivot/window caps. Every failure returns no partial packed/network/timing result.
Pin full native/corrected-reference passes with active >=10-gate windows on
both sides of boxes, capped/unlimited improvements and different prior rows.
Compare complete final mappings, counts, maxima and terminal float bits; verify
output composition, source ownership, no-box fallback and atomic late failure.
Do not infer full SAT-LUT parity from these four-input passes: native unit LUT
widths five/six, remaining public surface and broader design validation remain
required, along with the rest of the ABC roadmap.


Unit remapping now exposes `GiaUnitWindowSearchOptions.LutSize`: zero defaults
to four, native requests 1..4 use four-input feasibility, and 5..6 use six-input
feasibility. Preserve those source buckets rather than silently making requested
five a strict five-input limit. General window storage, local unit timing and
structural function validation now support six slots, including all elementary
truth variables and last-duplicate semantics. Packed unit entry points use
`AlternateNetwork` so wide inputs are preserved. Keep the four-input edge
adapter and reject wider edge search explicitly before candidate generation.
Pin complete native wider passes/counts/truth words for structural and already
wide initial mappings. Verify independent output composition, actual six-input
boxed packed results and unchanged-source behavior when narrower timing is
infeasible. Remaining source/public/diagnostic parity, representative larger
benchmarks and the broader ABC roadmap still require completion.


SAT-LUT audit found native `fDelay=false` missing from the initial Go entry
points. `AreaOnly` now exposes it through shared search and unit/edge/boxed/
packed passes. Skip timing exclusions entirely while still checking structural
functions, dependency/output closure, ownership and budgets. Unit application
may trade delay for area; do not accidentally route it through strict delay
validation. Whole edge passes retain the original supplied rows across mapping
changes and evaluate actual final delay with those rows, matching native no-
timing behavior. Do not derive an automatic zero delay constraint in this mode.
`TimingConstrained` distinguishes an enforced edge limit from metrics-only
reporting. Standalone edge replacement without supplied prior rows still builds
an owned assignment but does not enforce its maximum in area-only mode. Pin
native cases that actually worsen unit delay, verify prior-row retention and
compare boxed metrics and output composition. This closes the fDelay option
gap; optimized CNF, remaining command/API parity and the rest of the roadmap
remain unfinished.

The final `giaSatLut.c` surface audit adds `GiaManager.FullWindow` for the
unused whole-graph window path, preserving CI, AND-object and CO-driver order
with owned bounded storage. `FormatGiaWindowCut` preserves the exact ascending
four-mask diagnostic spelling and returns its population without stdout.
`GiaUnitRemapPassResult.Stats` and `GiaRemapPassStats.Format` retain stable
logical pass counters; nondeterministic wall-clock breakdowns move to Go
benchmarks. Go ownership and reusable concrete sessions replace manager
allocation, cleaning and destruction. With these additions every active
function in `giaSatLut.c` is represented by the unit/edge, boxed/unboxed,
four/six-input, timing/area-only and diagnostic paths above. Independent graph
ordering, Boolean composition, timing recurrence, SAT semantics, ownership,
limit/cancellation and fuzz tests validate the source without a native runtime.
The remainder of ABC and optimized CNF are still separate unfinished scope.

Optimized CNF work now has the native four-variable SOP database and cube
primitives. `CNFSOP4Cover` embeds the exact 75-byte `s_Data4` string payload,
decodes the native alphabet/delimiters and reverses ternary digit order into
packed ISOP pairs. Keep native cover order and return owned copies from a
synchronized immutable cache. Pin every cover and literal count, then prove
all 65536 functions independently over their sixteen assignments. Record both
production asset and immutable fixture hashes. Reject malformed/incomplete data
without partial tables; native padding after the final function is ignored.

`CNFISOPCubeLiterals` matches native negated-cube emission: pair 1 produces a
positive leaf literal, pair 2 a negative one, and 0/3 contribute no literal.
Preserve supplied variable order/duplicates, ignore high unused pairs and
validate only referenced variables. A uint32 cube supports sixteen pairs.
Check native order/counts and independently verify the emitted clause is the
negation of the cube. Next connect these primitives to `Cnf_ManWriteCnf`,
preserving reserved variable zero, output modes, mapped-node/CI/constant order,
exact clause emission and graph-to-variable mappings. DAR cut derivation and
CNF area mapping remain required before claiming the optimized encoder.


`WriteAigMappedCNF4` now ports the native small-cut writer for explicit mapped
cuts. Use reserved SAT variable zero, then exposed output variables, mapped roots
in caller order, CIs and the constant-one variable. Leave unused graph entries
at -1. Emit positive-function SOP implications first, then negative-function
implications, constant-one unit and native output assertions/equivalences.
Expose either every combinational output or exactly the sequential latch-input
suffix; reject invalid modes. Preserve raw clauses and duplicate leaf variables
without solver normalization. Validate stretched truth tables, mapped leaf
availability, prior-ID boundaries and bounded output storage. Return owned
immutable CNF metadata without mutating the graph or caller cuts.

The writer trusts supplied cut truth, matching native layering; it is not yet
a graph-to-optimized-CNF encoder. Pin exact native numbering/clauses including
mapped-order permutations and constant functions, exhaustively check projected
CI/output relations and independently verify valid mappings against AIG
simulation. Cover asserted outputs, unconstrained exposed variables, sequential
suffixes and unused variable zero. Wide-cut ISOP covers, cut derivation and
native CNF area mapping remain required before full optimized encoder parity.

## Supplied wide-cut CNF implementation checkpoint

`TruthTable.ISOP` ports native Kit zero/one/shared recursion for up to sixteen
inputs and preserves ordered cubes. Pair 1 requires a false variable, pair 2 a
true variable. Optional dual polarity selects fewer cubes, then fewer literals,
retaining direct polarity on ties and constants. Owned cumulative allocation and
work budgets replace C arena thresholds. A limited complement attempt retains
the completed direct cover; cancellation returns no result. Wrapped limit errors
must be recognized with `errors.Is`.

`WriteAigMappedCNF` consumes explicitly supplied cut truth tables. Cuts through
four inputs use the exact native SOP database; larger cuts use direct ISOP
covers of the function and its complement, regardless of the caller's TryBoth
option. Preserve native variable numbering and polarity/clause order. Validate
shape, dependency assignment and resource bounds before publishing owned output.
Supplied cut functions are trusted; automatic cut derivation, truth computation
and native area mapping must still be implemented before claiming an automatic
optimized encoder. Keep the four-input entry point compatible.

Regression gates include 196 pinned native ISOP cases with exhaustive cover
evaluation through sixteen inputs, 156 native wide writer cases, mixed small/wide
cuts, SAT projection exhaustive through eight inputs and sampled above eight,
constant/tie cases, allocation/work limits, complement fallback and cancellation.
Fuzz both cover semantics and projected CNF semantics. Permanent native vectors
are `test_vectors/oracle/kit_isop.json` and `cnf_mapped_wide.json`; fixture hashes are recorded in reference_build.json.

The initial `Cnf_DeriveMapping` area-flow pass is available as
`SelectCNFAreaFlow4` for supplied nontrivial candidates. Preserve native
10*(positive SOP cube count + negative SOP cube count), unsigned uint32 flow
addition, signed int32 leaf-flow division by structural reference count and eleven-bit
reference-sum truncation after each leaf. Smaller flow wins; ties favor larger
reference sum and otherwise retain the first candidate. Process ANDs in object
order and keep graph reference counts unchanged. Results include every AND,
so output-cone filtering is still required before mapped CNF emission.

Native exhaustive cost vectors cover all 65,536 truths. Tests exercise recursive
leaf flows, tie direction, owned storage, malformed truth shape and limits;
fuzz arbitrary candidate costs/ranking. Remaining integration work is native
DAR cut enumeration/order/pruning, cut transfer, mapping scan and postprocessing.

## Reachable CNF mapping scan checkpoint

`ScanCNFMapping4` preserves `Cnf_ManScanMapping` traversal: process CO drivers
in output order, increment each encountered signal's mapping references, expand
each AND only once, and visit leaves in supplied order. Preorder is used for
optimized CNF emission, postorder for mapping-array traversal. Area sums each
reachable cut's positive/negative SOP cube costs once. References include
constants, CIs and repeated leaves. Return separate owned counters rather than
overwriting structural AIG reference counts. Use an explicit bounded event
stack so deep input cones do not depend on Go recursion depth. Validate all
supplied cut shapes and require a selected cut for every reachable AND.

`WriteAigAreaFlowCNF4` connects supplied candidate selection, preorder scan and
raw clause emission. This provides a complete supplied-candidate path, while
automatic native DAR cut generation remains required. Native cut transfer's
leaf order and truth are retained directly in owned Go cut values; native arena
allocation is unnecessary. Supergate marks, optimized postprocessing and native
mapping serialization still require separate implementation.

Validation includes 64 pinned native scans in both orders, exhaustive SAT
projection for their supplied functions, integrated area-flow selection/emission,
shared and duplicate leaves, constants, unreachable cuts, missing/duplicate
roots, limits/cancellation, ownership, a 10,000-cut chain, and fuzz comparison
against an independent small recursive scan.

`ScanAigMapping4` separately covers native `Aig_ManScanMapping`, whose selected
DAR best cuts may be replaced by structural supergates. Process CO drivers in
order, increment owned mapping references before terminal/revisit checks and
collect every first AND in preorder. For selected supergates, flatten positive
same-type AND edges until a complement, different type or dynamically shared
node boundary; deduplicate identical signed leaves, collapse opposite-polarity
leaves to the constant-false empty boundary, and charge leaf count plus one.
Otherwise charge the selected truth's two SOP covers and recurse through cut
leaves. Keep structural graph references and mark fields unchanged.

The implementation uses explicit bounded stacks for both mapping and supergate
walks, lookup-only signed-boundary membership, 26-bit-compatible reference
overflow checks and signed-int-compatible area bounds. Validate unique best cuts
and selected supergate roots, permit supergates without a cut, and require a cut
for every other reachable AND. All failures and cancellation return no partial
result. Thirty-two native cases pin complete area/order/reference output; tests
also cover nested/repeated/constant/output-sharing cases, a contradictory
supergate and a 10,000-node chain, while fuzzing all supergate subsets.

## Automatic four-input optimized CNF checkpoint

`ComputeDARCuts4` now implements the DAR enumerator used by native optimized
CNF derivation. Clone and clean the graph, retaining IDs and tombstones; prepare
constant/CI unit cuts; process ANDs in object order. Preserve fixed cut slots
and their iteration order. Eviction chooses the first unused slot or the first
minimum-value cut among >=3 leaves, then >=2, then any. Eviction occurs before
merge success is known. Apply signature population pruning, sorted leaf union,
dominance filtering, aligned truth conjunction with complemented fanins, optional
support minimization and a second dominance pass. Reference-based values cap at
1000, favor constant/literal cuts with 1001 and penalize four singly referenced
leaves. A constant/literal result ends enumeration for that root.

Use bounded slots (default ten, supported 2..255), total slot and pair budgets,
cancellation and owned results. Nontrivial candidates exclude slot zero and
retain object/slot order. Reject explicit buffers/XOR until their native handling
is ported. No custom interfaces or mutable global manager are needed.

`DeriveAigCNF4` connects enumeration, area-flow selection, owned cut transfer,
preorder scan and clause emission. Defaults now implement the native automatic
four-input optimized CNF path on supported AIGs. Larger cuts in the supplied-cut
writer remain separate. Remaining scope includes supergate marks, alternate
derivation variants, mapping serialization, optimized postprocessing and
large-design verification/performance validation.

Validation gates: 192 pinned native candidate lists across slot pressure and
minimization modes; independent exhaustive cut evaluation; 24 complete native
CNF comparisons; exhaustive six-input/two-output SAT projection against original
AIG behavior; cleanup/ownership, malformed options, budgets/cancellation; fuzz
cut semantics and the complete automatic encoder. Maintain fixture
hashes with permanent vectors in test_vectors/.

## Packed CNF mapping serialization checkpoint

`WriteCNFMappingArray4` ports the native offset-prefix/five-entry format. Keep
caller cut order, zero offsets for unmapped objects, signed int32 truth bits
(repeat the uint16 truth in both halves), and four leaf slots padded with -1.
Preserve constants and literal cuts, arbitrary supplied cut order, and negative
truth entries. Validate root uniqueness, shape, stretched truth, leaf IDs and
packed allocation bounds; do not require caller-supplied cuts to cover outputs.

`DeriveAigCNFMappingArray4` runs automatic DAR generation and initial area-flow
selection, then postorder mapping scan and serialization. Keep its traversal
separate from the preorder optimized-CNF writer. Owned cleanup preserves IDs and
source graph storage. All 24 default native vectors pin complete packed arrays;
layout, padding, signed bits, no-cut graphs, invalid/duplicate roots, limits,
cancellation and ownership have tests, plus arbitrary truth/order fuzzing.
Alternate mapping formats and derivation entry points remain in scope.

## Object-ID CNF derivation checkpoint

`WriteAigObjectCNF4` and the supplied wide-cut `WriteAigObjectCNF` port
`Cnf_ManWriteCnfOther`: object IDs are SAT IDs, constant one uses variable zero,
all COs (including latch inputs) receive equivalence clauses, and final variable
count equals the allocated object span. Emit mapped cuts first, CO pairs next,
constant-one unit last. Preserve each mapped object's contiguous clause start
and count, CI count zero/start -1, and unmapped start/count -1. Return owned
metadata through copying accessors. Reuse the ordered cover emitter and translate
its temporary numbering; apply the caller's variable budget to final object
span rather than the temporary reserved-zero numbering.

`DeriveAigObjectCNF4` matches native alternate automatic derivation with optional
support minimization and a preorder packed mapping attached. Keep this preorder
mapping distinct from standalone mapping-array derivation's postorder. Source
cleanup remains on an owned clone. Tests compare 48 full native clauses/ranges/
mappings, prove their input/output relations exhaustively, check sequential CO
exposure, empty graph/variable-zero clauses, bounds/cancellation and metadata
ownership. Supplied wide-cut projection is exhaustive through eight inputs and
sampled above eight; native wide object-ID ordering remains a separate gate.
Fuzz complete alternate derivation and verify clause ownership ranges as well
as SAT semantics.

Together these default, mapping-array and object-ID paths account for every
active derivation in `cnfCore.c`. `Cnf_ManPrepare`, `Cnf_ManRead` and
`Cnf_ManFree` only manage a mutable global workspace and timing counters; Go
uses independent owned per-call state instead. Owned cut values replace arena
transfer/free and leave the source graph unchanged. `Cnf_Derive_old` is wholly
inside `#if 0`. Fresh end-to-end Go-only derivation fuzzing passed 558,554
executions, so `cnfCore.c` is promoted to validated. Supergate handling, other
CNF source files and larger-design validation remain separate subsystem scope.

## Correct cut composition prerequisite for postprocessing

`ComposeCNFCuts` implements Boolean substitution with sorted unique leaf union,
truth alignment, preserved redundant support and positive/negative SOP or direct
ISOP cost, saturated at 127. The default result-input cap is nine, matching
native nMergeLimit=10 including its temporary substituted variable; explicit
caps allow sixteen. Allocation/work limits and cancellation expose no partial
result, and neither source truth nor leaves may change. Dual-polarity ISOP
optimization is disabled for these costs.

The native `Cnf_CutCompose` has three verified semantic defects. Removing a leaf
before computing its truth pointer reads a leaf ID instead of the truth table.
Moving the substituted variable to the union's last position before stretching
it there again misaligns support. Treating packed cofactors as adjacent truth
words fails below five result inputs. Preserve original native regression
outputs, but use correct substitution in Go. A corrected comparison was used
only to freeze the reviewed dataset; it is not retained or executed. The
correction changes those three operations while retaining merge, truth
operations and cost computation. All 240 fixtures match that dataset and independent
exhaustive truth evaluation; 237 original first words differ. Do not describe
these cases as unchanged native parity.

Required gates also include constant/literal/shared/disjoint functions, source
ownership, malformed/unsorted/duplicate leaves, absent substituted root, merge
caps, work/ISOP budgets, cancellation and a sixteen-input exhaustive composition.
Fuzz arbitrary top/fanin functions against independent substitution. Port
`Cnf_ManPostprocess` selection order and mapping-reference updates on top of this
correct composition primitive; keep automatic recovery an explicit option.

## CNF postprocessing recovery checkpoint

`PostprocessCNFMapping4` now ports the native one-pass selection/reference logic
over supplied four-input mappings, with corrected cut composition. Initialize
owned mapping references from output reachability. Visit roots in increasing
object order; stable-sort fanins by clause cost, treating terminals as zero.
Only singly referenced AND fanins may be absorbed, and at most one absorption
is accepted for each root. Skip failed proposal limits, sentinel cost 127 and
cost increases; equal combined cost is accepted. On acceptance dereference the
old top and fanin leaves, reference composed leaves, remove the fanin mapping
and require its mapping reference to become zero. Do not recursively dereference
structural AIG nodes or alter graph counters. Return surviving preorder cuts,
owned references, area and change count. Global work/cancellation failures must
return no partial mapping.

`DeriveAigPostprocessedCNF4` explicitly connects automatic DAR/area-flow mapping
with one recovery pass and the wide-cut emitter. Keep recovery explicit because
the native default derivation does not enable the pass. Newly composed cuts can
exceed four inputs and use direct ISOP covers; standalone packed four-input
mapping serialization cannot represent these wider cuts.

Validation compares 48 full mappings to the native pass redirected only to the
corrected composition reference, then proves original-AIG semantics exhaustively.
Check stable ordering, one absorption per root, updated repeated/terminal refs,
source/output ownership, merge-limit skip versus global attempt-limit abort,
cancellation and nonincreasing area. Fuzz arbitrary structural mappings through
recovery and SAT projection. Automatic integration also receives semantic tests.
Owned cuts make `Cnf_ManTransferCuts`/`Cnf_ManFreeCuts` unnecessary: their only
source behavior is moving pointers through `Aig_Obj.pData` and restarting/freeing
a flex arena. `Cnf_ManPostprocess_old` hardcodes null best-cut pointers and zero
new cost while all producing calls are commented out, so it is a dead diagnostic
scaffold rather than a second pass. A fresh Go-only recovery/SAT campaign passed
553,467 executions. These dispositions account for every source definition and
promote `cnfPost.c` to validated. Supergate handling, other derivation variants
and larger-design validation remain separate CNF subsystem work.

## Owned CNF transforms and solver loading checkpoint

`AigMappedCNF.Clone` deep-copies clauses, variable mappings and signal literals.
`Shifted` returns independent storage with a base added to every assigned object
variable and literal. Preserve -1 unmapped entries and literal signs. Unlike
`Cnf_DataLift`, increase the returned variable span by base so the result can
be loaded directly into a solver; document and test this deliberate API change.

`TransformPolarity` matches native per-variable AIG zero-input phase flags,
optionally excluding CO phase flags. Flip clauses by variable phase; keep object
variable numbers but also flip returned input/output literals so original signal
meanings survive. Validate graph/mapping span, dimensions, variable/literal bounds,
work allocation caps and cancellation. Transform twice must restore the relation
and signal mappings. Copying/transformation never mutates source clauses/graph.

Mapped CNF `LoadInto` reuses the concrete incremental solver loading path with
a variable base. Different bases must form independent copies; callers shift
assumptions by the same literal offset. Failure/cancellation during incremental
loading can leave added clauses, following existing AddClause semantics. No
custom solver interface is needed. This ports the single-copy loading portion;
native multiframe/latch initialization remains unfinished.

Gates include all 48 native polarity/lift clause/mapping comparisons, exhaustive
original-AIG SAT projection, phase involution, owned storage, allocation/variable
limits, cancellation, two copies in one solver and object-ID constant variable
zero. Fuzz arbitrary graph phases and base offsets against SAT projection.

## Multiframe CNF loading checkpoint

`AigMappedCNF.UnrollFrames` now builds the native timeframe relation in owned
storage. Preserve native order: frame-zero clauses; for each later frame, two
equivalence clauses per register linking previous LI to current LO, then that
frame's clauses; finally optional negative units for all first-frame LOs. Frame
stride equals source variable count. Apply an optional base offset consistently.
Require positive frame count, matching graph/signal spans and assigned LI CO
variables when multiple frames are requested. Preserve already asserted source
outputs; do not add separate bad-state constraints.

Return variable/frame counts, copied clauses and phase-aware input/output
literals by frame. Link the semantic LI/LO literals so phase-transformed inputs
remain correct. Initialize semantic register values to zero. Budget frame count,
object span, total variables, clauses and literals before allocation; cancellation
or invalid dimensions return no partial frame relation. Source graph/CNF never
changes, replacing the native loader's temporary literal shifts/restoration.
`AigCNFFrames.LoadInto` uses the concrete solver path; caller retains ownership
even if clauses prove inconsistency. Alternate native solver loader variants
and initialization policies beyond this native all-zero flag remain in scope.

Tests compare 96 complete immutable clause-stream rows. Exhaustively enumerate input sequences and
initial states through four frames for six transition patterns and one/two
registers, verify all latch-input outputs, and reject corrupted linked states.
Also cover base offsets, phase changes, mapping ownership, missing LI variables,
resource caps and cancellation; fuzz the full trajectory relation.

## Output and XOR CNF constraints checkpoint

Ported native output OR/AND and XOR constraint helpers with pure clause builders
and concrete solver additions. Output helpers include every CO, including latch
inputs, require assigned CO variables and preserve phase-aware signal literals
and base offsets. Return one OR clause or one unit per output; empty AND is true.
Native empty OR aborts on a solver assertion; Go deliberately returns the empty
clause and reports inconsistency through AddClause instead. Validate the whole
output mapping before incremental solver mutation.

`CNFXORClauses` and `AddCNFXOR` preserve native four-clause order for
result = a XOR b. Require positive valid variable IDs, permit aliases and check
solver capacity before additions. Solver additions retain standard incremental
partial-mutation behavior if a later clause proves inconsistent.

Tests match 36 immutable clause-stream rows and exhaustively prove all 27
alias triples, output constraints across zero..four COs, phases and offsets.
Fuzz arbitrary XOR IDs/alias combinations against independent clause evaluation.
Native proof-enabled solver2 wrappers remain separate unfinished work.

## Graph-free CNF and selected literal positions checkpoint

`CNFFormula` now owns checked graph-free clauses, including zero-variable empty
formulas and empty contradictory clauses. Expose copied clauses and concrete
solver loading; validate variable/literal spans, dimensions, allocation limits
and cancellation. Constructor input must never remain aliased.

`CollectFlipPositions` preserves native flattened literal order, with empty
clauses contributing zero positions and absent variables producing empty lists.
`LiftAndFlipPositions` applies only the supplied occurrences: flip sign, then
add the variable offset. Duplicate positions are sequential operations, not a
set. Validate every position before processing; preserve all other literals and
clause boundaries. Return owned output and expand its variable span as needed,
an explicit difference from native unchanged nVars. MaxLiterals bounds source
literal count and requested operation count; check variable capacity after each
operation, including repeated shifts. Failure/cancellation returns no result
and leaves source storage untouched.

Tests match 84 native collections/transforms across absent variables, offset
zero, duplicates and nonmonotonic positions. Independent fuzz applies selected
operations to a separate flat reference and checks source preservation. Include
constructor/output ownership, invalid positions/literals, variable/work caps,
zero-variable formulas and empty-clause solver behavior.

## Bounded DIMACS text output checkpoint

`CNFFormula.DIMACS`, mapped CNF and frame-relation wrappers now preserve native
plain text formatting: comment/header, quantifier order, literal spaces, clause
zero terminators and final blank line. Default numbering adds one to zero-based
SAT variables; native Readable mode preserves raw IDs and its variable-zero
diagnostic ambiguity. Keep that mode explicit and avoid claiming it is standard
DIMACS. Nil prefixes omit lines, while nonnil empty prefixes emit a terminated
empty line. Preserve caller order/duplicates; validate each prefix variable.

Return owned bytes rather than opening a path. Enforce source limits, output
byte budget and cancellation without exposing partial bytes. Integer formatting
uses bounded scratch storage; protect empty clauses and zero-variable formulas.
Filesystem writes/errors and native compressed output remain separate work.

Tests match ten exact native text fixtures. Also check explicit empty prefixes,
invalid IDs, exact byte limits, cancellation and mapped/frame wrappers. Fuzz
standard text by independently parsing header/literals and reconstructing every
clause boundary. Native raw-ID diagnostics receive exact-byte regression tests.

## Compressed CNF and file output checkpoint

`DIMACSGZIP` now serializes native-format text then writes a deterministic Go
gzip stream through a concrete bounded buffer. Keep text and compressed byte
budgets separate. Check cancellation during text generation, chunked compression
and buffer flush/close; return no bytes after any error. Use zero/default gzip
header metadata and preserve exact decompressed text. Native zlib and Go can
produce different compressed bytes, so parity gates compare decompressed content
and separately test Go determinism. No custom interface type is introduced.

`WriteDIMACSFile` safely selects compression by a case-sensitive .gz suffix,
including short paths. Serialize/validate fully and check cancellation before
creating or truncating a destination. Use native create/truncate behavior and
creation permissions filtered by the OS umask. Return OS errors instead of
printing a message. A failure after an OS write starts may leave partial output;
this API does not promise atomic file replacement. Existing destinations must
remain unchanged on serialization/limit/cancellation failures before opening.
Mapped CNF and frame relations expose matching wrappers.

Gates include ten native gzip fixtures with exact decompressed text, deterministic
Go output, exact compressed caps, separate text caps, compression cancellation,
wrapper equivalence and fuzz roundtrips. File tests use temporary directories
for plain/gzip/short paths, case-sensitive suffixes, destination preservation
on serialization failure and directory/missing-parent errors.

## Object-owned literal polarity metadata checkpoint

`AigObjectCNF.LiteralPolarities` ports native per-owner/head-polarity sign
aggregation. Negative occurrences contribute 1, positive occurrences 2; mixed
signs yield 3. Record one byte per clause with two-bit masks for its first four
tail positions. Include all supported tail occurrences in collection, but
explicitly ignore output slots beyond four rather than relying on truncated
shifts. Clear touched variable masks after each owner. Unit clauses yield zero.

Check object-sized metadata, valid contiguous ranges, nonempty clauses, matching
head owner, no overlap, complete clause coverage, supported width and literal
bounds. Return owned bytes and no partial result on cancellation/failure. Tests
compare all 48 native byte arrays, exercise mixed signs, separate head polarities,
wider tails and malformed ranges. Fuzz independently rescans same-owner clauses
to verify each packed mask. Other CNF utilities remain in scope.

## Checked plain CNF reading checkpoint

`ParseCNFDIMACS` now provides strict standard DIMACS parsing by default and an
explicit NativeLines mode for supported native line semantics. Strict mode
allows multiple clauses per line and clauses across lines, including zero-
variable/zero-clause formulas, and requires exactly the declared clause count.
NativeLines requires positive declared counts, one zero-terminated clause per
body line, ignores tokens after the first zero and accepts actual/declaration
count differences without printing warnings. Preserve literal/duplicate/empty-
clause order in both modes. These compatibility semantics do not include native
malformed-number atoi coercion or crashes on missing headers.

Use bounded byte scanning rather than allocating all lines/tokens. Validate
ASCII decimal signs/digits/overflow, one complete header, literal span, zero
termination and byte/line/variable/clause/literal budgets. Reject quantified
prefixes in this plain-CNF API. Return owned formulas and no partial results
on syntax errors, limits or cancellation. `ReadCNFDIMACSFile` bounds chunks
before appending/allocating and returns OS errors; it does not implicitly unzip
input, matching the native reader's scope.

Tests match ten native line-mode results, exercise strict multiline/multiple-
clause syntax, zero counts, malformed/truncated/overflow input, duplicate headers,
quantified prefixes, resource limits, cancellation and temporary-file read
bounds/errors. Fuzz arbitrary bytes in both modes; validate every successful
formula and strict standard roundtrip. Remaining work includes file-solving
wrappers and other CNF utilities.

## Typed CNF solving and file wrapper checkpoint

`CNFFormula.Solve`, `SolveCNFDIMACS` and `SolveCNFDIMACSFile` now connect checked
source loading to an owned concrete SAT solver. `CNFSolveResult.NativeCode`
preserves native SAT=0, UNSAT=1 and unknown=-1. Return full SAT models plus
optional owned final-N-variable suffixes matching native nPis extraction.
Validate suffix size before use and independently check every source clause
before publishing a SAT assignment. No model or suffix may escape on UNKNOWN,
UNSAT, resource errors or cancellation. Native arbitrary model choices need
not equal Go's, but both must satisfy every clause.

Apply formula/reader resource limits and cancellation through parsing, loading,
search and verification. Search work budgets begin after loading as in native
code. Cancellation at any stage returns typed UNKNOWN/SatCancelled rather
than claiming SAT/UNSAT. Filesystem/syntax/resource errors remain explicit
errors. Zero-variable formulas and empty clauses must work. Keep logging out
of the library; CLI diagnostics remain separate. Native learned-database start/
delta/ratio tuning and proof-solver variants are still unfinished.

Validation matches 33 native default outcomes and replays both models. Tests
cover suffix values/ownership, missing/bad files, malformed text, trivial and
search UNSAT, conflict/decision limits, cancellation while loading and zero-
variable cases. Fuzz small formulas against complete assignment enumeration
and verify returned models.

## CNF diagnostic text and graph-variable collection checkpoint

`CNFFormula.DiagnosticText` now ports `Cnf_DataPrint` separately from the
interchange-oriented DIMACS writer. Preserve the short header, absence of a
comment and clause-zero terminators, trailing spaces, empty-clause blank lines
and final blank line. In readable raw-ID mode, write the sign separately so a
complemented variable zero remains `-0`; ordinary signed integer conversion
would silently collapse this native diagnostic distinction. Provide the same
owned formatter on mapped, simple and multiframe concrete CNF values. Bound
source structure and output bytes, check every literal, honor cancellation and
return no partial bytes.

`AigMappedCNF.CIVariables`, `PIVariables` and `COVariables` cover the native
graph-order SAT-number collectors. The PI compatibility spelling intentionally
includes all CIs, including sequential latch outputs, because the C routine does.
Preserve `-1` for unassigned objects, return owned slices, validate the graph and
object-variable span, and enforce object/cancellation limits. Do not introduce a
generic CNF interface: each concrete representation keeps direct wrappers.

Validation compares four exact native diagnostic byte streams, including empty
formulas/clauses and readable `-0`. Tests cover wrapper equivalence, owned
collector results, exposed and asserted CO maps, malformed mappings, exact byte
bounds and cancellation. Fuzz arbitrary clause streams in both numbering modes
against an independent formatter. Go constructors, owned slices and garbage
collection replace C's raw allocation/free lifecycle; manager-specific scratch
arenas and proof-solver-specific wrappers remain separate work.

## Fast CNF cone prerequisites checkpoint

`CollectFastCNFLeaves` preserves fanin-zero-first DFS, expanding the root even
if marked. Stop on marked boundaries; optional complement stopping retains
signed literals and uses exact signed uniqueness. Otherwise regularize edges
and deduplicate by ID. Require explicitly bounded terminals. Use iterative
stacks with event/object limits and first-encounter output slices, not ordered
map iteration or recursion proportional to graph depth.

`CollectFastCNFVolume` treats supplied regular leaves as visited boundaries and
returns unique internal ANDs in native postorder. `DeriveFastCNFCutTruth` assigns
ordered six-variable patterns to leaves and computes complemented ANDs in that
postorder. Match native <=6 leaves and <=256 total scratch entries. Constant
leaves are abstract variables, later constrained by the graph's constant unit.
Validate unique/disjoint leaves/nodes and assigned dependencies rather than
reading stale object iData. Keep marks, traversal IDs, iData and static truth
storage out of caller graph state; return owned data and atomic errors.

Tests compare 24 native cone vectors, regular/signed order and node order, plus
truth where supported. Independently evaluate all six-CI assignments, fuzz
cones, check malformed boundaries/dependencies, limits/cancellation and a
10,000-node chain. Remaining fast-path work is mux recognition/marking,
supergate-versus-ISOP clause emission, size counting and reverse-node variable
assignment/complete derivation.

## Fast CNF mux recognition and marking checkpoint

`AigManager.RecognizeMux` now preserves native structural mux/XOR-type detection
and matching priority: common opposite-polarity grandchildren at (1,1), (0,0),
(0,1), (1,0). Return positive control and complemented data literals describing
the regular root's function; do not confuse it with a complemented mux output.
Use a private checked-graph implementation inside graph-wide passes.

`MarkFastCNF` ports native CI/CO marks, nonoverlapping mux groups, shared-node
and complemented-edge boundaries, followed by support recovery. Preserve
else-before-then mux recovery, equal-data reference handling, first eligible
nonmux leaf absorption and support sentinel six. Keep all marks/support scratch
owned, reject unsupported explicit buffers/XOR nodes and bound pass/cone events.
Return native CO-driver and AND-gate rule-violation counts rather than printing
or silently claiming a usable mapping. Complete fast derivation must check
these diagnostics before emitting invalid variable references.

Internal leaf traversal now shares prevalidated graph/mark storage and uses
lookup-only uniqueness with ordered append, avoiding per-cone object-sized
clears and repeated graph validation. Diagnostics use owned visitation epochs.
Tests match 32 native mark arrays and all 361 mux records, plus independent
exhaustive mux simulation, source preservation, invalid roots, resource limits
and cancellation; fuzz mixed traces. Fixture graph construction explicitly
sequences C AND calls so compiler argument order cannot permute compared IDs.
The marking checkpoint feeds the complete derivation below. Diagnostic-only
size counting and broader scale/performance qualification remain separate work.

## Complete fast CNF derivation checkpoint

`DeriveAigFastCNF` now composes validation, `MarkFastCNF`, internal cone
collection and truth derivation without mutating the source graph. It preserves
native allocation order: reserve variable zero, assign exposed CO variables,
walk marked AND objects in reverse object order, assign CIs, then assign the
constant-one object. Accept only the native output modes: no exposed outputs,
all combinational outputs, or the register-count suffix for sequential graphs.

For each marked root in reverse object order, collect regular leaves and the
internal postorder. If every internal edge that crosses an unmarked boundary is
positive, collect signed supergate leaves and emit the one wide implication plus
one binary implication per leaf. Otherwise derive the ordered truth word over at
most six leaves and emit the direct positive cover followed by the complement
cover. Force native direct-cover selection rather than the optional smaller-of-
both-polarities heuristic. Preserve cube order and literal polarity. Emit
asserted-driver units or exposed-CO equivalences after all node clauses, then the
constant-one unit last. Return the existing owned `AigMappedCNF` representation,
not a native arena view or a new interface.

Validate every variable lookup before emission and reject nonzero marking
diagnostics, unsupported explicit buffers/XOR objects, complex cones wider than
six leaves and malformed dependencies. Enforce independent object, traversal,
variable, clause, literal and ISOP bounds with overflow-safe accounting. Combine
all cancellation hooks and return no partial result on any error. Keep internal
lookup maps lookup-only for membership/indexing; output order always comes from
explicit graph walks and slices.

Validation compares 96 complete immutable dataset rows: exact variable counts, complete
object-variable arrays and exact raw clause streams across asserted, all-exposed
and sequential suffix modes. Independently project each result through the SAT
solver for every six-CI assignment and every requested two-output valuation.
Check graph immutability, constant-only graphs, every resource limit, immediate
cancellation and a 1,000-input iterative supergate. Fuzz mixed AND/XOR/MUX
construction and all three output modes, checking the full original-AIG SAT
relation rather than only self-consistency. Permanent oracle data remains under
`test_vectors/`; Go's ephemeral fuzz corpus is not used as a permanent fixture.

`CountAigFastCNFNodes` now covers `Cnf_CutCountClauses` and
`Cnf_CountCnfSize` without printing. It returns a concrete
`FastCNFNodeSize{MarkedObjects, Clauses}` and deliberately does not call those
values the final variable/clause totals: native counting excludes output and
constant clauses. It recomputes owned marking, preserves direct ISOP order,
uses the same bounded iterative cones, returns atomically and matches all 32
native count pairs. Preserve the native corner case where a contradictory
small signed AND supergate has a smaller regularized truth-cover count than the
clauses actually emitted; a dedicated regression pins the one-versus-four
result and fuzzing treats the diagnostic as a lower bound.

Retain explicit follow-up gates for large real designs, performance/memory
profiling, unsupported larger complex cones and native arena allocation-layout
concerns; none of these gaps invalidate the supported owned derivation contract.

## Core AIG DFS, support and level checkpoint

`AigManager.DFS` now covers native CO-rooted node-only/full preorder and
postorder using explicit event stacks. Preserve constant-first full traversal,
CO order, fanin-zero-first descent, first-encounter sharing and the distinction
between traversing CO objects versus their regular drivers. As in native
`Aig_ManDfs`, require all applicable live objects to be reachable; translate
its assertion on dangling internals into `ErrInvariant`. `DFSAll` appends only
otherwise unused CIs after CO postorder and rejects dangling internals.

`DFSArray` retains the constant and explicit CO roots, while `DFSNodes` omits
the constant and replaces CO roots with drivers. `Support` accepts multiple
roots, replaces COs with their drivers, ignores constants and returns unique CIs
in native fanin encounter order. `SupportSize`, `DAGSize`, `MaxOutputLevel` and
`LevelBuckets` provide the corresponding concrete queries. `CollectCut` marks
supplied regular leaves in owned scratch and returns unique internal ANDs in
native fanin-zero-first postorder, excluding boundaries; duplicate or incomplete
boundaries reject. Level buckets retain
object order and reject live objects above the maximum CO-driver level rather
than reproducing a native assertion. `VerifyTopologicalOrder` uses the stronger
Go manager check, covering topological fanins together with hash, phase, level,
partition and reference invariants.

All operations use owned boolean/event storage, object and visit caps,
cancellation and atomic error returns; no traversal IDs or graph marks are
modified. Twenty-four immutable fixtures compare every ordered output and scalar.
Additional tests cover buffers, invalid roots, dangling graphs and 10,000-node
chains. Fuzz mixed AND/XOR/MUX traces, proving unique postorder, fanin precedence,
support cardinality, DAG/cut counts and source immutability. Remaining `aigDfs.c`
work is timing-box dependency handling in topology verification and choice levels.

## AIG cone transfer and CI composition checkpoint

`TransferAigCone` now ports `Aig_Transfer` and its recursive worker using an
explicit postorder. Preserve the native first-`nVars` CI mapping through
destination `IthVar` creation, complement propagation at every fanin and root,
structural hashing, the constant shortcut and same-manager no-op. Reject a cone
whose support lies beyond the mapped CI prefix instead of following stale C
scratch data. Explicit buffers are outside the native worker's safe contract
and return a typed error.

`AigManager.ComposeAigInput` now ports `Aig_Compose` and its worker. Map every
ordinary CI to itself, map exactly the indexed CI to the supplied literal and
reconstruct only the root cone in native fanin order. The replacement remains
opaque during reconstruction, so a replacement function may contain the same
CI without creating a graph cycle. Preserve root and replacement phases and
structural-hash reuse.

Both operations prevalidate graph ownership and topology, use owned mapping and
event storage, enforce object/event limits and cancellation, and construct on a
private clone. Commit the clone only after the result and full manager check
succeed; all error paths leave caller-owned managers unchanged. This deliberate
transactional strengthening replaces native scratch marks/data and partial
mutation without narrowing successful behavior.

Validation compares 32 pinned native records containing exact returned
literals and every resulting object/fanin. Exhaustively compare all 64 input
assignments in each record. Add constant and same-manager shortcuts, automatic
destination CI creation, malformed-prefix/buffer/literal/index cases, exact
limits, cancellation, ownership checks, an internal self-dependent replacement
and a 10,000-node iterative cone. Fuzz arbitrary hashed AIG construction and
prove transfer equivalence plus substitution semantics for every four-input
assignment. Keep the oracle and reviewed vectors under `test_vectors/`.

## Reverse fanout DFS checkpoint

`AigManager.ReverseDFS` now ports `Aig_ManDfsReverse` without installing the
native manager's persistent fanout database. Build private fanout slices by
scanning live objects in ascending ID order, which preserves the insertion
order produced by `Aig_ManFanoutStart`. Pre-mark every CO, then visit each
AND/buffer in object order and emit it after all reachable unmarked fanouts.
This retains native fanout-first postorder and includes dangling logic while
excluding terminals and COs.

`ConeANDMask` replaces `Aig_ConeMark_rec`, `Aig_ConeCleanAndMark_rec`,
`Aig_ConeCountAndMark_rec` and `Aig_ConeUnmark_rec` with owned state. It returns
the same unique regular-AND membership mask; its population is the native
count. There is no persistent mark to clear and no generic Go object-data field
to null, so clean/unmark become lifetime properties of the returned slice.

Use an explicit event stack and owned seen/fanout storage. Validate the graph
before traversal, enforce object/event limits and cancellation across fanout
construction and traversal, and return no partial order on failure. Do not add
fanout fields to `AigManager` or expose an interface merely for this query.

Validation compares all 32 native orders and complete cone masks/counts across
shared, complemented, multi-output and deliberately dangling logic. Add explicit-buffer behavior,
exact limit/cancellation failures, source immutability and a 10,000-node chain.
Fuzz arbitrary AND/buffer construction, proving every applicable object occurs
exactly once and every internal fanout precedes its fanin. Keep the oracle and
reviewed fixture permanently under `test_vectors/`.

## AIG choice-chain traversal and no-box levels checkpoint

Add optional concrete object-indexed choice successors to `AigManager`; do not
introduce a graph or metadata interface. `SetChoiceChains` accepts owned
representative-first ID lists and atomically validates AND-only membership,
single predecessors, disjoint chains and acyclicity. Extend append, clone,
cleanup and `Check` so alternatives survive graph maintenance and newly created
objects receive empty successor slots. Functional equivalence remains an
explicit caller proof obligation, as with native `pEquivs`.

Update ordinary node/full DFS so postorder visits a successor before structural
fanins and preorder reproduces native's unusual rule: append the current node,
visit its successor with postorder DFS, then continue fanins in preorder.
`DFSChoices` separately follows fanin zero, fanin one and successor before
emitting each node. Use iterative events and owned seen state throughout.

`ComputeChoiceLevels` returns an object-indexed owned level vector and maximum
without overwriting `AigObject.Level`, whose structural invariant remains
checked independently. In the no-box path, CI/constant levels are zero, COs
copy driver levels, ANDs add one to their maximum fanin and then take the
maximum with the successor level. Reject combined structural/choice dependency
cycles rather than accepting traversal-state artifacts. Handle timing-box CI
dependencies in the explicit checkpoint below.

Replace the printing-only `Aig_SupportSizeTest` diagnostic with
`SmallSupportANDCount` and a general concrete
`CountANDsWithSupportAtMost`. Reuse one bounded visit budget across all AND
roots and visitation epochs across support walks; return the count rather than
writing timing or diagnostic text.

Validation compares 32 native phase cases for node/full preorder and postorder,
dedicated choice DFS, every object level and the maximum. Add atomic invalid
chain tests, the exact native fixed-16 count plus an independent wide-threshold
case, clone/cleanup/growth ownership, limits and cancellation, dependency
cycles and a 10,001-alternative chain. Fuzz choice phases while proving order,
level recurrences and exhaustive equivalence of all alternatives. Keep the
oracle and reviewed fixture under `test_vectors/`.

## AIG timing-box topology and choice-level checkpoint

Extend the concrete AIG traversal path with an explicit `*TimManager`; do not
store a persistent timing pointer on `AigManager` and do not introduce an
interface. Require the timing manager's CI/CO counts to match the graph and
validate both directions of every box-terminal association before traversal.
Reject overlapping, stale, out-of-range or position-inconsistent metadata as a
typed invariant error instead of relying on native assertions.

`VerifyTopologicalOrderWithBoxes` scans objects in ID order after the ordinary
manager check. When it reaches a box-output CI, every CO in that box's input
range must already have been seen. Account for every object and dependency
check in one visit budget, honor cancellation, and never mutate graph or timing
traversal marks.

`ComputeChoiceLevelsWithBoxes` shares the iterative dependency engine used by
no-box choice levels. A box-output CI depends on all input COs and receives one
plus their maximum level; COs copy their driver, ANDs add one to the maximum
fanin level and then take the maximum with their choice successor. Include
dangling box outputs by starting walks from all COs and then all CIs. Detect
cycles across structural edges, choice successors and box arcs and return no
partial level vector. A nil optional timing manager retains the no-box path.

Validation pins 32 native records covering phase variation and every valid or
invalid placement of two chained output CIs relative to their input COs. Compare
the complete object level vector, maximum and topology result. Add malformed
association, overlap, dimension, no-box, exact limit, cancellation, ownership
and self-dependent-box tests plus a 10,000-box iterative chain. Fuzz short box
chains with independently selected valid/invalid placements and prove both the
topology predicate and level recurrence. Store the reviewed fixture under
`test_vectors/`, record source/fixture hashes, and only then promote
`src/aig/aig/aigDfs.c` to validated coverage.

## AIG ordered fanout-index checkpoint

Port `aigFanout.c` as a concrete owned `AigFanoutIndex`, not an interface and
not a mutable pointer stored on `AigManager`. Build by scanning live objects in
ascending ID order and appending fanin zero before fanin one, retaining each
edge's fanout object and pin. Preserve this deterministic order in queries and
return copies so callers cannot mutate the index accidentally. Enforce native
`Aig_ManFanoutStart`'s no-buffer precondition; the separate reverse-DFS path can
continue to build a buffer-aware private index where needed.

Represent native `FanoutStop` by ordinary ownership lifetime. Provide concrete
add/remove operations for staged graph-editing algorithms: graph-derived forms
select the native fanin pin, while explicit-edge forms let a future transaction
update an index after staging structural changes. Removal keeps the order of
all survivors and reinsertion appends at the tail, matching the circular list.
Reject duplicate, missing, out-of-range and invalid-pin operations atomically
instead of permitting native list corruption.

Validation pins 32 native snapshots before removal, after removing a middle
edge and after reinsertion, comparing every object's complete ordered list and
pin numbers. Add clone/query ownership, malformed mutation, buffer, exact
limit, cancellation and 10,000-fanout tests. Fuzz arbitrary structurally hashed
AIGs; independently rebuild expected lists from all object fanins, compare
every reference count and total edge count, then remove/reinsert a selected
edge and verify stable order. Keep the fixture under `test_vectors/`, record
source and fixture hashes, and promote `src/aig/aig/aigFanout.c` only after
full tests, race testing and vet pass.

## AIG cut-truth checkpoint

Port `aigTruth.c` using the concrete `TruthTable` representation and owned
object-indexed scratch. `AigManager.CutTruth` accepts an ordered boundary of CI
or internal AND IDs, binds the first occurrence of a repeated leaf, collects
the remaining cone in fanin-zero-first postorder and evaluates every packed
word. Treat native AIG constant one correctly, apply complement bits while
reading fanins, support a complemented result literal and require every reached
CI to be covered. Do not reuse `AigObject` data fields or caller-owned work
buffers.

Expose the single-node operation as `AigCutTruthOne` so both native branches are
represented even while ordinary Go AIG construction lowers XORs: AND combines
independently complemented inputs, while a dedicated XOR object combines two
regular inputs. Reject complemented dedicated-XOR inputs because native XOR
nodes canonicalize their fanins and carry any phase on the returned literal.

Bound object discovery, object-indexed mappings and the sum of internal/result
truth words before allocation; check cancellation during discovery, evaluation
and result transfer. Errors return neither a table nor a partial node order.
Permit constants and root-as-leaf cases as safe owned extensions, while
rejecting CO/buffer leaves, uncovered inputs and malformed literals.

Validation pins complete native node orders and truth words for 32 cuts at 2,
5, 8 and 10 leaves, including internal boundaries and multiword tables. Pin all
four complemented-AND cases and the native dedicated-XOR branch separately.
Add constant, complemented root, repeated leaf, malformed boundary, exact
resource, cancellation and ownership tests. Fuzz arbitrary six-input hashed
AIGs and compare every assignment with independent direct simulation. Keep the
fixture under `test_vectors/`, record reproducible hashes and promote
`src/aig/aig/aigTruth.c` only after full tests, race testing and vet pass.

## AIG dynamic-order checkpoint

Port `aigOrder.c` as a concrete owned `AigOrder`, not an interface and not a
pointer stored on `AigManager`. Mirror the native circular list with object zero
as sentinel. Initialize it with every live AND in ascending object order and
enforce native's no-buffer precondition. Return copied sequences and support
independent cloning so callers cannot alias the order accidentally; ordinary
ownership lifetime replaces `Aig_ManOrderStop`.

Preserve cursor semantics exactly. Insert a new AND immediately before `Next`.
Removal repairs `Previous` backward and decrements the processed count when it
removes that cursor; it repairs `Next` forward independently, including the
case where both cursors name the same object. Keep `Advance` separate from the
iterator increment: native `Aig_ObjOrderAdvance` moves only `Previous`, while a
concrete `Refresh` performs the surrounding macro's update of `Next` from the
successor of `Previous`. A `Step` convenience may perform both only when no
intervening edit is required. Reject duplicate, absent, sentinel, wrong-type
and out-of-range operations atomically.

Validation pins complete sequence and cursor/count state for 32 native traces
covering four insertion positions, refresh-selected insertion, equal-cursor
removal, predecessor removal, successor removal and iteration completion. Add
empty-order, clone ownership, malformed mutation, native buffer, exact limit,
cancellation and 10,000-node tests. Fuzz arbitrary hashed AIGs, compare initial
order with an independent ascending-ID scan, perform deterministic current or
processed-node removals and compare the final sequence and counts with a slice
model. Keep vectors under `test_vectors/`, record reproducible hashes and only
promote `src/aig/aig/aigOrder.c` after full tests, race testing and vet pass.

## AIG timeframe-expansion checkpoint

Port `aigFrames.c` as the concrete `AigManager.ExpandFrames` operation and
return an owned `AigFrameExpansion`; do not introduce a graph interface or use
native object scratch. Preserve the native source-object-major mapping layout,
including complemented internal references and regular CO object IDs. Expose
copied complete mappings plus checked object/frame lookup. After cleanup,
invalidate every entry whose target became a tombstone, matching the native
callers that explicitly null pointers to deleted objects without exposing an
unsafe dangling-pointer state.

Create primary-input CIs frame-major, followed by frame-zero latch-state CIs
when initialization is disabled. Map initialized latch outputs to constant
zero. Rebuild ANDs in source object order through structural hashing and carry
all latch-input drivers simultaneously to the next frame's latch outputs.
When outputs are requested, emit every frame's primary outputs unless
enlargement selects only the final frame. When registers are requested, emit a
final latch-input CO suffix driven by the last frame normally or frame zero in
enlargement mode, then install the source register count. Copy Name and Spec,
but retain native's zero constraint count.

Validate the source before work and reject native-unsupported buffers and
dedicated XOR objects explicitly. Treat zero frames and a register suffix that
cannot fit the constructed CI partition as dimension errors rather than
executing native out-of-bounds behavior. Bound the target object arena and the
complete mapping allocation before exposure, check cancellation throughout,
and discard the private target on any failure so the source and caller-visible
state remain unchanged.

Validation pins all 48 native combinations of one through three frames and the
initialization, output, register and enlargement flags. Compare total and
primary CI/CO counts, register count, every live or tombstoned object's type,
fanin literals, reference count, level and phase, and every cleanup-filtered
map entry. Add ownership, copied-map, invalid lookup, buffer rejection, exact
resource, cancellation and 10,000-node tests. Fuzz arbitrary hashed
combinational AIGs over three frames and independently compare each frame's
outputs with direct source evaluation. Keep the reviewed fixture under
`test_vectors/`, record reproducible source/fixture and fixture hashes,
and promote `src/aig/aig/aigFrames.c` only after full, race and vet gates pass.

## AIG justification and ternary-simulation checkpoint

Port the core of `aigJust.c` with concrete owned slices. `FindJustification`
must reproduce `Aig_ObjFindSatAssign`'s eight deterministic heuristic starts,
its value-conflict checks on reconvergent objects, fanin-zero/fanin-one order,
and CI-ordinal SAT literal convention. Preserve the final attempt's literals
on failure because the native vector remains populated. Also preserve and
document the native constant-one quirk: the recursive routine reports success
for either requested value without adding a literal.

Use an explicit continuation stack rather than recursive Go calls. Each retry
gets fresh object-indexed assignment state, while a single bounded work budget
and cancellation callback cover all eight attempts. Return owned literals and
never write traversal IDs, marks or generic object data. Accept CO roots by
folding their edge phase into the requested value. Reject reached buffer or
dedicated-XOR nodes with a typed unsupported error, but do not reject an
otherwise valid cone merely because an unrelated buffer exists elsewhere.

Implement `TernarySimulate` with the native 1/2/3 zero/one/unknown encoding,
last-wins duplicate CI assignments, complemented-edge negation and Kleene AND.
Its iterative cone walk must retain the native fanin-zero short circuit: once
that input is zero, do not visit fanin one. Validate literal ordinals and bound
all visits. Do not expose an interface for ternary storage.

Validation pins 32 native phase/root cases over contradictory reconvergent
cones, including regular nodes, CO roots, successes, failures, exact ordered
support literals and every object's value under duplicate partial assignments.
Add source ownership, malformed input, reached/disconnected buffer, exact
limit, cancellation, constant and 10,000-node tests. Fuzz arbitrary hashed
AIGs against an independent Kleene evaluator and require every successful
nonconstant justification to force the requested value. Keep the reviewed
fixture under `test_vectors/` with reproducible hashes. Initially retain
partial status until `aigPack.c` provides the candidate/pattern operations
required by `Aig_ManJustExperiment`; then complete and promote the file through
the packed-pattern checkpoint below without substituting a reduced experiment.

## AIG memory-arena checkpoint

Port all three `aigMem.c` allocators as concrete owned Go types rather than an
allocator interface. `AigFixedArena` returns pointer tokens containing borrowed
byte slices, preserves native chunk sizing, ascending allocation within a new
chunk, LIFO recycle, first-chunk restart, allocated-byte accounting and the
lifetime high-water mark. Use a reversed free stack internally so native order
is retained without an O(n) prepend. Reject foreign, stale and double recycle
operations instead of corrupting an intrusive free list.

`AigFlexArena` suballocates exact requested byte lengths from owned chunks,
including zero-length requests and permanent chunk-size growth after an
oversized request. Restart invalidates every borrowed block and retains the
first chunk. Preserve the native nominal post-restart statistics, but always
check the first chunk's actual Go slice capacity before reuse; native instead
extends `pEnd` using the enlarged nominal size and can address past that first
allocation.

`AigStepArena` constructs the same 8, 16, 32, ... fixed classes and maps every
request through `4 << steps`. A zero request returns nil. Requests above the map
size receive individually retained blocks whose recycle operation is a no-op,
and reported usage remains the sum of fixed-class allocated bytes, exactly as
in C. Track oversized bytes separately for safety limits even though native
statistics omit them. Return concrete copied class statistics in place of
verbose stop-time printing.

All arenas must enforce aggregate byte and block limits before allocation and
invalidate borrowed tokens on recycle, restart or stop. Validation pins fixed
reuse identity and pre/post-restart statistics; every flexible allocation's
entries, nominal chunk size, chunk count and byte totals; and every step class,
oversized count and internal fixed statistic. Add malformed constructor,
lifetime, foreign/wrong-size recycle, limit and stopped-arena tests. Fuzz both
arbitrary fixed fetch/recycle traces and mixed flexible/step allocation and
restart traces. Keep the reviewed fixture under `test_vectors/`, record
reproducible hashes, and promote `aigMem.c` only after full, race and vet gates.

## AIG MFFC checkpoint

Port all eight operations in `aigMffc.c` as concrete `AigManager` methods and
owned `AigMFFC` results. Use copied object reference counts for every
dereference/reference transaction so no caller-visible count, traversal mark
or generic scratch field changes. Replace the recursive C routines with
explicit continuation frames, preserving fanin-zero/fanin-one order, native
counting of AND nodes, buffer pass-through during reference transactions and
the exact depth boundary `Level <= LevelMin`.

`MFFCSupport` returns the native first-visit ordered boundary and internal-node
count, including a CI root as a zero-size singleton support. `MFFCLabel`
returns an object-indexed owned membership slice. When probabilities are
present, evaluate each unsuffixed-constant expression with double
intermediates and round on each float store, matching C rather than relying on
Go float32 expression rounding. `MFFCLabelCut` temporarily increments every
listed leaf reference, including duplicates, then removes those protections
after the balanced transaction. `ExtendMFFCCut` selects only maximum-level
leaves, chooses the first leaf with the strictly smallest dereferenced cone,
skips that object while collecting the replacement boundary and restores both
the leaf and root transactions before returning.

Validate managers and IDs before work, reject negative limits, bound all
iterative events with one shared budget and cancellation callback, and expose
no partial result on failure. Preserve native level-zero buffer boundary
behavior. If support traversal actually reaches a positive-level buffer,
return a typed unsupported error instead of triggering the native assertion.
Verify exact reference restoration before exposing any successful result.

Validation pins 32 native phase and external-reference variants, comparing
depth-limited sizes and support order, complete labels, protected-cut labels,
float power values and cut-extension existence/order. Add CI-root, malformed
leaf, nil-manager, exact-limit, cancellation, ownership, boundary/reached
buffer and 10,000-node iterative tests. Fuzz hashed AIGs against an independent
recursive reference-count transaction for size and ordered support, and check
label size against independent membership cardinality. Keep the permanent fixture under `test_vectors/`, register source and fixture hashes, and promote `src/aig/aig/aigMffc.c` only after clean fixture
fixture integrity verification plus full, race and vet gates pass.

## AIG reconvergence-window checkpoint

Port both routines in `aigWin.c` as the concrete
`AigManager.FindReconvergentCut` operation returning owned leaf and visitation
slices. Replace native `fMarkA` scratch with an object-indexed Boolean slice so
the source manager remains immutable and concurrent read-only callers do not
share traversal state. Preserve initial fanin-zero/fanin-one order, stable
frontier removal and fanin append order.

For each frontier scan, assign CIs the native unexpandable cost. For an AND,
count its unmarked fanins; accept cost zero or one regardless of fanout, but
reject cost two when the node's structural reference count exceeds the caller's
fanout limit. Select the first strictly cheaper node, replace equal-cost choices
only with a deeper node, and stop the scan immediately on cost zero. Stop the
whole construction if no node is expandable or if replacing the best node
would exceed the size limit. Keep the complete first-mark visitation order as
well as the final frontier.

Validate the graph and root before allocating result state. Require a frontier
capacity of at least two and a nonnegative fanout limit. Replace native
assertions for non-AND roots and reached buffers with typed errors. Bound every
frontier examination with a shared work limit and cancellation callback, and
discard partial slices on failure.

Validation pins 32 native phase, external-reference, size-limit and
fanout-limit combinations with exact frontier and visitation arrays. Add nil,
invalid, CI and buffer roots, malformed limits, exact work exhaustion,
cancellation, result ownership and a 10,000-node reconvergent-chain test. Fuzz
arbitrary hashed AIGs and compare both ordered arrays against an independent
frontier model while enforcing the requested cut-size bound. Keep the reviewed
fixture under `test_vectors/`, register reproducible source and fixture hashes,
and promote `src/aig/aig/aigWin.c` only after clean
fixture integrity verification plus full, race and vet gates pass.

## AIG packed-pattern and completed justification checkpoint

Port `aigPack.c` as a concrete `AigPatternPack`; do not introduce allocator,
simulation or pattern-store interfaces. Keep object signatures, CI patterns
and CI care masks as owned `[]uint64`. Give every pack its own copy of the
native multiply-with-carry RNG state so its words and reset/advance behavior
match C without retaining the native process-global interference. Preserve the
left shift that reserves bit zero as the all-zero assignment.

Simulate constant one, CIs, ANDs and COs in object order with exact complemented
edge behavior. Reject buffers and dedicated XOR objects rather than silently
leaving their signatures zero. Select internal nodes whose signatures are
constant or differ from a constant in one bit. Preserve care insertion exactly:
first reject any bit conflicting with an already cared literal without
mutation, then apply literals sequentially so duplicate-variable assignments
retain native last-wins behavior. Search bits 1 through 63 in order, increment
the total counter once per requested pattern and increment skipped only when no
bit accepts it.

Replace native diagnostic printing and destructor output with copied pattern,
care and signature slices plus a concrete statistics value. Count every care
bit and retain the 33-bin histogram indexed by the smaller of the signature's
one and zero populations. Validate manager identity and dimensions on every
operation, reject graph growth after pack construction, enforce object and
pattern limits, check cancellation during long simulation/randomization and
discard partial public results on failure.

Complete `Aig_ManJustExperiment` as `RunJustificationExperiment`. Start a pack,
select rare candidates in native object order, request zero for phase-one nodes
and one for phase-zero nodes, retain separate success/failure counters and the
total successful literal count, and add each successful assignment to the
pack. Return the final candidates, counters, patterns, care, signatures and
statistics instead of timing/printing side effects. Native callers must install
CI/CO ordinals before justification; Go managers already maintain stable CI
ordinals directly.

Validation pins 32 native graphs with exact reset RNG words, base and final
signatures, rare candidates, every candidate's post-attempt literal vector,
success/failure counts, literal totals, care masks, packed patterns and both
histograms. Add nil/malformed options, buffer rejection, stale graph detection,
copied-result ownership, duplicate/conflicting literals, all-bits-conflicting
skip behavior, exact object/candidate/pattern limits and cancellation. Fuzz
arbitrary hashed AIGs and compare each of all 64 packed output bits with direct
Boolean evaluation; rerun the independent justification semantic fuzzer. Keep
the reviewed fixture under `test_vectors/`, register all
source and fixture hashes and promote both `aigPack.c` and `aigJust.c` only
after immutable-fixture verification plus full, race and vet gates pass.

## Complete AIG operation checkpoint

Finish `aigOper.c` in the existing concrete `AigManager` rather than adding an
operation interface. Add generic AND/XOR dispatch, lookup-only canonical AND,
the explicit `Mux2` spelling, adjacent-pair and two-slice miters, and canonical
N-variable AND/OR/XOR constructors. Preserve the native balanced split in
multi-operations and miter reduction; return the OR of pairwise differences.
Do not mutate caller-owned pair or node slices.

Retain the port-wide choice to lower XOR into ordinary structurally hashed AND
nodes and to omit the native optional additional-strashing rewrite mode. These
are representation optimizations rather than distinct Boolean operations, and
a single representation avoids infecting every downstream algorithm with
dedicated-XOR modes. Within that representation, reproduce exact default
native object creation order. In particular, the pinned GCC build evaluates
the independent arguments in lowered XOR, direct MUX and majority expressions
right-to-left; preserve that observable ordering for stable object IDs. Tests
or callers that require an explicitly sequenced structural fixture must build
its intermediate ANDs explicitly.

Validate every input literal before mutation. Build multi-step miter and
N-variable operations on a manager clone, enforce a shared high-level operation
budget and cancellation callback, and commit only on success. Reject empty or
mismatched miter inputs, negative variable counts, invalid operation kinds and
malformed literals with typed errors. Replace the standalone `Aig_MuxTest`
random BLIF-writing diagnostic with deterministic semantic unit and fuzz tests.

Validation pins 32 native phase sequences and compares all returned literals,
lookup hits/misses, CI arrays and every object's type, fanins, reference count,
level and phase. Add lookup nonmutation, caller-slice ownership, exact budget,
cancellation and atomic-error tests; exhaustively verify zero-through-seven-
variable constructors and exercise 10,000-input constructors and miters. Fuzz
arbitrary pairwise miters against exhaustive four-input evaluation and rerun
the mixed-operation truth-mask fuzzer. Keep the permanent fixture
under `test_vectors/`, register reproducible hashes, and promote `aigOper.c`
only after immutable-fixture verification plus full, race and vet gates pass.

## Stable-ID-safe AIG object-mutation checkpoint

Port the `aigObj.c` operations that fit the concrete append-only `AigManager`
without introducing an object-management interface. Existing CI, CO, buffer
and structurally hashed AND construction collectively implement native create
and connect behavior. Add recursive internal-cone deletion, ordered-output
deletion, output-driver patching and an owned string formatter. Retain deleted
IDs as permanent zero-value tombstones and rebuild neither IDs nor caller-held
literals.

Perform every multi-object mutation on a clone and commit only after a complete
invariant check. Reject referenced deletion roots, terminals and choice-chain
members. Make recursive deletion iterative, process fanin zero before fanin
one, update the structural hash and reference counts exactly once, and bound
work with a deletion limit and cancellation callback. When patching an output,
increment the replacement reference before deleting the old cone so replacing
a root with one of its descendants preserves that descendant. Keep primary
output, constraint and register-input partitions consistent when an output is
deleted.

Do not copy native `Aig_ObjDeletePo`'s stale `vCos` entry into Go: the ordered
output slice is authoritative and must contain only live COs. Compare native
and Go at the complete object-arena level for this operation, then separately
test Go's stronger slice and partition invariants. Replace `Aig_ObjPrint`'s
stdout side effect with exact owned formatting and typed errors.

Validation pins 32 phase seeds with separate recursive-delete, output-patch and
output-delete managers, comparing every object type, literal, reference count,
level and phase. Add atomic malformed/limit/cancellation cases, choice-member
protection, descendant replacement, exact formatting, partition accounting and
a 10,000-input iterative deletion. Fuzz both patched-output semantics and
dangling-cone deletion. Keep the permanent fixture under
`test_vectors/`, register source and fixture hashes, and run full, race and
vet gates plus byte-identical clean fixture verification.

Complete the remaining `aigObj.c` replacement paths with a concrete owned
rebuild rather than weakening `AigManager.Check`. Native
`Aig_NodeFixBufferFanins`, `Aig_ManPropagateBuffers` and `Aig_ObjReplace` can
temporarily create old-ID buffers whose fanins have newer IDs. The Go
equivalent must instead substitute the replacement throughout a newly owned
graph, bypass every buffer chain, preserve CI/CO order and sequential
partitions, structurally hash rebuilt ANDs, and return an explicit old-to-new
literal map. This makes the unavoidable remapping visible while keeping every
observable Go graph topological.

Reject replacement targets that are buffers, the replaced object itself or
transitive users of it. Reject choice graphs until choice remapping has a
separately specified contract. Bound both source/result objects and iterative
dependency events, honor cancellation and expose no partial result. Test
higher-ID replacements (the motivating native forward-reference case), phase
composition through multi-buffer chains, unrelated buffers, constants/CIs,
terminal and register/constraint order, malformed targets, cycles, limits,
cancellation and a 10,000-node cone. Exhaustively compare rewritten Boolean
outputs and fuzz against an independent formula model. Promote `aigObj.c` only
after focused, full, race and vet gates pass.

## BSAT clause/proof-store checkpoint

Port `sat/bsat/satStore.c` and the store-owned portion of `satStore.h` as a
concrete ordered `SatProofStore`; keep the header partial until its four
interpolation manager families are implemented. Preserve encoded-literal
insertion sorting, rejection of duplicate variables regardless of phase,
stable clause IDs/order, variable maxima, single-empty tracking, root/A prefix
marking, last-unit A-to-B transfer, visit metadata, 64-KiB chunk/report
accounting and the `p vars clauses roots clausesA` text format.

Use copied clause snapshots and ordered slices instead of linked/raw pointers;
future interpolators should own their watch/reason indexes. Replace stdout,
silent file failure and assertions with errors and byte/variable/clause/literal
limits. Correct the native loader defects by starting insertion at zero rather
than the declared count and restoring root/A flags from header prefix counts.
Test sorted ownership, same-variable rejection, empty-clause atomicity,
mark/change behavior, exact chunk rollover, exact marshal text, comments,
malformed headers/counts/literals, limits, file round trips and lifecycle.
Fuzz constructed stores through marshal/parse against the owned clause model
without C, generated fixtures or retained fuzz output.

## BSAT solver utility checkpoint

Port all of `sat/bsat/satUtil.c` onto the concrete Go `SatSolver`; do not create
a duplicate solver2 type merely to retain parallel diagnostic entry points.
Preserve exact clause formatting in standard one-based DIMACS and native
zero-based/raw-ID modes, original non-learned clauses, zero-level assertions,
ordered assumption units, header counts and the final blank line. Serialize to
owned bytes before checked file output; apply byte bounds and cancellation and
never mutate the solver while normalizing its root snapshot.

Return exact first- and second-generation statistics strings rather than
ignoring a stream argument and printing globally. Preserve packed-word counter
conversion, expose solver2 per-variable memory accounting with explicit
pointer/activity widths, and project requested variables from an owned SAT
result with bounds. Represent the source's unconditional-assert
`Sat_SolverDoubleClauses` as atomic `ErrUnsupported`; do not implement its
disabled comment body or partially mutate the solver.

Test exact standard/raw bytes, root propagation, assumptions, explicit
inconsistency, parse/file round trips, both statistics layouts, packed counter
boundaries, 32/64-bit memory arithmetic, model projections, all errors, limits,
cancellation and disabled-operation atomicity. Fuzz generated formulas and
assumptions by solving the original snapshot and a freshly parsed Go-only
snapshot and comparing status, with no C oracle or retained generated corpus.

## BSAT AIG interpolation checkpoint

Port `sat/bsat/satInterA.c` and `satInterB.c` together as a concrete
`SatAigInterpolator` with explicit standard and polarity-mux modes. Share the
owned proof-replay engine and root AIG initialization. In standard mode,
preserve OR for A-local pivots and AND for shared/B pivots. In mux mode,
preserve OR for A-local pivots, AND for B-local pivots and the precise Mux0 or
Mux1 orientation selected by the shared pivot's polarity in the reason clause.

Accept a copied caller order for shared SAT variables or infer variable order;
when an order is supplied, reject it if any actual common variable is omitted.
Return an owned checked `AigManager` with ordered CIs and one CO, and clean
dangling construction without remapping the returned driver. Replace the A
variant's absolute deadline with synchronous cancellation and apply explicit
replay/AIG-object limits to both modes. Return errors instead of assertions,
unchecked allocation or partial graphs.

Represent both `ManDeriveClauses` helpers with concrete partition-CNF AIGs,
preserving shared-input order and the native reserved input gap before local
SAT-variable positions. Test standard/mux shared pivots, A-local OR, B-local
AND, A-only/B-only constants, learned replay, explicit-order coverage,
ownership/reuse, exact partition input mapping, graph invariants, every limit,
cancellation and lifecycle. Fuzz both modes under variable/polarity/clause/global
order changes, exhaustively checking both Craig conditions and expected
semantics without C, generated native fixtures or retained fuzz output.

## BSAT truth-table interpolation checkpoint

Port all of `sat/bsat/satInter.c` as a concrete `SatTruthInterpolator` sharing
the owned BSAT proof-replay state. Preserve inference of common A/B variables
in variable order, explicit caller-selected global-variable order, the source's
eight-global-variable bound, A-root initialization as the disjunction of its
shared literals, B-root initialization as one, OR combination for A-local
pivots, AND combination for shared/B pivots and final-empty-clause extraction.

Return a copied canonical `uint64` truth table with its ordered SAT variables
and exact meaningful bit count instead of an aliased `unsigned*`. Preserve
assignment-index bit order and provide exact owned high-to-low binary formatting
in place of stdout. Reuse concrete proof state without interfaces; keep truth
tables as an optional owned hook on resolution rather than duplicating watched
propagation. Bound truth words and replay resources, support cancellation and
reject duplicate/out-of-range globals, malformed partitions and invalid proof
traces without partial results.

Test inferred and explicit ordering, A-only/B-only constant partitions,
learned-unit replay, copied configuration/results, exact formatting, all limits,
cancellation, reuse and lifecycle. Exercise the full eight-variable/256-bit
boundary. For every constructed case, independently enumerate assignments to
prove both Craig conditions. Fuzz variable identities, polarities, root order
and explicit global order against that model, using no C oracle or retained
generated corpus.

## BSAT proof replay and UNSAT-core checkpoint

Port all of `sat/bsat/satInterP.c` as a concrete `SatProofCoreManager` over
owned `SatProofStore` snapshots. Preserve watched-clause propagation, root
unit processing, negated learned-clause assumptions, proof numbering,
antecedent construction, resolution verification, stronger-resolvent
shrinking and selection between root-clause and learned-clause cores. Keep
watch/reason indexes local to the replay manager rather than adding pointer
links to stored clauses.

Use iterative deterministic postorder traversal for core extraction so deeply
learned proofs cannot overflow the Go stack. Verify root cores through the
native Go SAT solver, and return exact owned strings for clause, resolvent,
interpolant-placeholder and BMC learned-core formatting instead of writing to
stdout or caller-owned `FILE` values. Bound variables, clauses, literals,
resolution steps and antecedent references; support cancellation and reject
malformed prefix flags, missing/floating empty clauses, invalid proof reasons
and unproved learned clauses without exposing partial results.

Test direct root conflicts, propagation conflicts, learned-unit replay,
stronger resolvents, exact traversal order, exact formatting, SAT verification,
malformed traces, every limit, cancellation, reuse and lifecycle. Fuzz both
contradictory-unit families and learned proofs under variable renaming,
polarity changes and root reordering, verifying every extracted root core with
the Go solver. Use no C oracle, generated native fixture or retained fuzz data.

## BSAT legacy vector and literal checkpoint

Port all of `sat/bsat/satVec.h` with concrete types and no interfaces. Preserve
the integer and opaque-pointer vectors' initial capacity four, exact 3/2 growth
sequence, ordered push, shrink-only resize, integer pop and first-match removal.
Use `uintptr` for the old `void*` payload. Return copied whole-vector views and
provide checked indexed read/write in place of exposing mutable backing
pointers; replace missing-element assertions, allocation failure and deletes
with errors, explicit limits and idempotent stop.

Reuse the package's typed `SatLit`, `SatVar` and `SatClauseHandle` rather than
adding integer aliases. Cover construction, conditional sign, negation,
variable/sign extraction, signed one-based DIMACS print/read and bounded
validity, including zero, minimum-integer and maximum-variable errors. Add a
concrete three-valued Boolean and a complete legacy statistics value. Unit
tests must pin the 4,6,9,12,18 growth sequence, duplicate removal order,
shrink/pop, indexed mutation, copied ownership, opaque handles, all constants
and literal round trips. Fuzz integer-vector operations against a slice model;
use no C oracle or retained generated corpus.

## BSAT streaming trace checkpoint

Port all three functions in `sat/bsat/satTrace.c` as a concrete owned streaming
writer. Preserve creation/truncation at start, the exact 40-space first-line
placeholder, signed one-based DIMACS literal rendering, ordered clause/root
counts, zero terminators and stop-time rewrite of
`p <variables> <clauses> <roots>` without truncating the remaining padding.
Keep writes manually driven as in the native API; do not silently couple trace
output to every Go solver clause mutation.

Return stdio errors, bound clause/literal accumulation and make invalid stop
arguments atomic so callers can retry. Provide a concrete solver helper for
the final variable count without adding an interface. Test exact final bytes,
positive/negative/empty clauses, mixed root flags, placeholder padding,
creation errors, limits, invalid literals, repeated stop and recovery from bad
stop arguments using temporary files only. Fuzz the pure clause formatter
against an independent signed-literal model; retain no generated corpus or C
oracle. Promote only after focused, full, race and vet gates pass.

## BSAT proof-information bitset checkpoint

Port all of `sat/bsat/satProof2.h` as a concrete owned proof-information
manager, without interfaces or exposed vector pointers. Preserve 64-bit word
rounding, flat fixed-width records keyed from the first learned-clause ID,
grow-at-exact-capacity doubling, record-preserving zero extension, shrink before
the named clause, saved-clause order (including duplicates), compaction and
the source's first-new-ID update rules.

Represent a proof chain as OR of prior learned records and bits for problem
clauses. Preserve ignored learned references before the tracked prefix, direct
root IDs when no map exists, mapped negative exclusions, the zero chain-stop
result and descending UNSAT-core order. Own/copy the root-ID mapping and add an
explicit checked update operation for later abstraction growth instead of
retaining the native caller-vector alias. Return copied record/mapping views;
replace assertions and frees with errors, limits and idempotent stop, and
report actual Go-owned capacity rather than C vector capacity.

Test width boundaries, exact-capacity growth, record preservation, direct and
mapped chains, mapping ownership/updates, skipped roots, descending cores,
shrink, duplicate saved records, empty/nonempty compaction and lifecycle.
Fuzz random root/prior-record unions and saved-record compaction against an
independent bitset model. No native proof logger, generated fixture or C
subprocess is permitted; promote only after focused, full, race and vet gates.

## BSAT text resolution-checker checkpoint

Port the three functions in build-disabled `sat/bsat/satChecker.c` as a
concrete signed-DIMACS resolution checker independent of the larger proof DAG
subsystem. Replace stdout clause printing with exact returned formatting.
Preserve the resolver's nested complementary-pair count (including duplicate
multiplicity), exactly-one-pivot requirement, removal of both pivot phases and
first-occurrence unique union in antecedent order.

Parse the source's proof grammar strictly: sequential one-based clause IDs,
root records terminated by two zeros, `* antecedent antecedent 0` resolution
records, backward-only references and final empty-clause reporting. Return
structured counts and copied clauses; replace assertions and silent file-open
failure with errors. Bound input bytes, lines, clauses and total stored
literals before allocation. Construct complete proofs directly in Go and test
nonempty conclusions, malformed IDs/references/pivots/trailers, duplicate
clauses, limits and temporary-file behavior. Fuzz both resolution against an
independent pair-count/union model and arbitrary proof bytes for panic freedom.
No native checker, generated C fixture or subprocess is permitted.

## BSAT paged clause-store checkpoint

Complete `sat/bsat/satClause.h` with a concrete owned clause store rather than
exposing page pointers or creating an interface. Preserve the exact even/odd
problem/learned page encoding, even handles and low-bit tagged binary literals,
integer storage rounding, page rollover, hidden ID word, first-problem scratch
skip, ordered iterators and all three learned-clause defaults. Keep the existing
solver's concrete clause values while making the header's store independently
usable by later solver parity work.

Represent direct bitfield/pointer access with copied clause snapshots and
checked metadata/ID mutations. Preserve bookmark/rollback, restart, current
handles, memory formulas and the unusual shrink behavior that rewinds storage
without decrementing entry counts. Implement learned compaction as the native
mandatory two-pass protocol: the dry run writes future handles into retained
clause IDs, the move pass verifies those relocations, packs clauses, assigns
dense IDs and repairs a bookmark even when its pivot is removed. Ordered slices
must determine iteration; any handle map is lookup-only.

Test tagged literals, exact size/page boundary arithmetic, blank scratch and
plus-one entries, copied ownership, multiple pages, rollback into reused pages,
shrink count behavior, lifecycle, limits and formatting without C. Independently
model compaction positions, clauses removed on both sides of a bookmark and
post-compaction rollback, then fuzz randomized clause sizes/marks against that
model. Promote the header only after focused, full, race and vet gates pass.

## BSAT allocator checkpoint

Port `sat/bsat/satMem.c` and `satMem.h` together as three concrete owned arena
types, without interfaces or pointer-shaped integer handles. Preserve the
fixed allocator's exact 1,024-entry/64-KiB chunk rule (with an eight-entry
minimum), LIFO recycling, peak statistics and first-chunk restart. Preserve
the flexible allocator's 64-KiB bump chunks, zero-length entries and doubled
oversized chunk growth. Preserve the step allocator's 8-byte power-of-two
classes, direct oversized allocations, no-op oversized recycle, restart and
class-only memory-usage result.

Use checked borrowed block objects so recycle, restart and stop invalidate
access deterministically; reject foreign blocks, double recycling, mismatched
sizes and invalid dimensions. Bound bytes and block counts before allocating,
return errors in place of assertions/allocation failure, and expose copied
statistics rather than verbose stdout. Unit tests must derive every threshold,
class and accounting value directly in Go, including boundary sizes, multi-
chunk growth, LIFO identity, large-block retention and lifetime invalidation.
Fuzz fixed and flexible/step state machines against independent counters and
allocation invariants. No C helper, compiler, native runtime oracle or
generated fixture is permitted. Promote both entries only after focused,
full, race and vet gates pass.

## BSAT solver2 public gate-encoding checkpoint

Begin `sat/bsat/satSolver2.h` with its five self-contained public gate
constructors. Return concrete owned clause specifications rather than hiding
partition and caller ID metadata inside an interface. Preserve exact native
clause count, clause order, literal order and complement orientation for
constant, buffer, AND, XOR and constraint encodings. Reject zero IDs and
invalid variables with checked errors; clone all returned literal storage.

Provide an atomic composition helper for the existing concrete `SatSolver`.
Validate the complete batch before mutation, load through an owned trial state,
commit logical UNSAT results, and expose no partial mutation on API errors.
Keep `PartA` and external ID fields in the specifications because the current
CDCL core does not consume proof/interpolation metadata. Do not pretend this
completes the solver2 state, clause accessors, model/final-conflict queries,
limits, activity controls, bookmark helpers or core solver API.

Unit tests must pin every literal sequence and metadata field, prove caller and
clone ownership, exhaust all assignments for every encoding and solve the
loaded formulas under complete assumptions. Fuzz complement/partition choices
against independent AND/XOR Boolean relations without native code or fixtures.

Execution checkpoint: all five builders and the atomic loader are concrete and
tested. Exhaustive units and a fresh 5,000-case Go-only fuzz campaign pass. The
uncached full suite passes in 10.139s, race in 58.334s, and vet is clean,
promoting `satSolver2.h` from unstarted to partial while retaining explicit
remaining gaps.

Continue the same header through the public state/query surface by extending
the shared concrete CDCL solver, not by duplicating it. Retain an owned model
only after SAT and an owned final conflict only after UNSAT; clear both before a
new solve and whenever a valid formula mutation makes the old result stale.
Expose checked model values/literals, copied conflict clauses, variable/live-
original/learned/conflict counts, an absolute runtime deadline, a replaceable
learned-clause limit and activity clearing with native increment reset. A zero
learned limit disables reduction. Include retained answers and controls in
owned bookmark snapshots and invariant checking.

Use a distinct runtime stop reason and test an already-expired deadline rather
than relying on timing races. Cover SAT-to-UNSAT result replacement, copied
conflicts, mutation invalidation, zero-variable SAT, bookmark restoration,
limit replacement, activity/heap invariants and nil/error paths. Fuzz retained
results against each solve result and run the ordinary, race and vet gates.
Document that the shared solver's live-original count excludes normalized
tautologies and root-satisfied clauses; exact solver2 arena call-count and
handle compatibility remains a later facade obligation.

Execution checkpoint: retained models/conflicts and all listed controls now
participate in solve, mutation, reset, bookmark and invariant lifecycles. A
fresh independent 5,000-case retained-result fuzz campaign passes. The latest
uncached full suite passes in 10.127s, race in 58.917s, and vet is clean. The
header remains partial for its exact arena facade, packed clause helpers and
remaining solver2 declarations.

## BSAT solver2 exact caller-arena checkpoint

Add one concrete `SatSolver2` facade over the shared CDCL engine rather than
forking propagation and analysis. Preserve every successful caller clause in a
separate owned arena with stable one-based handles, including full sorted
tautologies and clauses already true at root that the logical core normalizes
away. Deduplicate ordinary repeated literals as native insertion does. Retain
external IDs and PartA bits, expose copied clause snapshots, and allow clause/
variable metadata mutation only before the first solve. Keep variable and
clause limits explicit.

Single and batch insertion must use private cloned core/facade state and commit
only on success; a contradiction or API error exposes no partial clause,
variable, watch or assignment changes. Gate specifications load through this
arena so every generated clause receives a handle and metadata. A facade
bookmark must pair the core checkpoint with copied caller clauses and variable
partitions; rollback removes tail clauses/variables without forgetting that a
solve occurred. Stop is idempotent.

Test exact normalized records, root-satisfied and tautological count divergence
from the logical core, snapshot ownership, metadata timing, gate handles and
truth, contradictory insertion, atomic batches, limits, rollback and lifecycle.
Fuzz arbitrary tautological records and require stable sequential handles,
exact facade counts, zero logical-core clauses, satisfiability and invariants.
No native oracle or generated fixture is permitted.

Execution checkpoint: the concrete facade now closes the original-clause call-
count and handle gap while reusing the validated CDCL search path. Exact units
pass and the standalone arena fuzzer completed 5,003 executions. The uncached
full suite passes in 10.261s, race in 58.715s, and vet is clean. `satSolver2.c`
is promoted from unstarted to partial; packed learned proof references and
proof integration remain explicit gaps.

## BSAT solver2 root-watch cleanup checkpoint

Port `sat_solver2_check_watched` as a non-printing atomic operation on the
shared solver. First establish complete level-zero propagation on a private
clone. For each watch bucket in literal order, retain unsatisfied clauses and
remove entries for a clause only when one of its literals is true at decision
level zero. Remove both entries, count each clause once by original/learned
kind, count every removed watch entry and commit only after complete validation.

Record a concrete unwatched flag in the clause rather than weakening checks.
The solver invariant must require exactly zero watches and a currently true
level-zero literal for such a clause, while ordinary clauses still require one
watch per watched literal. Include the flag and rebuilt watch buckets in owned
bookmarks. Later solves and clause additions must remain correct because root
assignments are permanent within that state. Expose the same operation through
the concrete solver2 facade.

Unit tests must distinguish a root-satisfied clause from an undecided clause,
pin exact unique/entry counts, verify idempotence, solve after pruning, restore
all watches on rollback and cover nil/stopped errors. Fuzz batches of binary
clauses sharing a later root unit; require every expected pair of entries to be
removed, every formula to remain satisfiable and all invariants to hold without
native code.

Execution checkpoint: atomic root-watch pruning, invariants, facade exposure
and rollback are implemented. The standalone fuzzer passed 5,000 executions
with up to 128 clauses. The uncached full suite passes in 10.370s, race in
58.821s, and vet is clean, covering `sat_solver2_check_watched` while leaving
proof integration as the main remaining gap.

## BSAT solver2 owned-memory accounting checkpoint

Port `sat_solver2_memory` as a structured owned report rather than a floating-
point byte total. Use actual Go type sizes and allocated backing capacities,
matching the source's capacity-based intent. Account separately for the solver
struct, clause headers and literal arrays, watch-list headers, optional watch
entry arrays, variable/heap metadata, trail/unit arrays, retained results and
the complete owned core bookmark. Preserve the native `fAll` distinction by
always counting watch headers but counting per-literal entry arrays only when
requested.

Extend the same report through `SatSolver2` with facade struct, exact caller-
clause headers/literals, variable partitions and facade-bookmark storage. Keep
proof bytes explicit and zero until a proof manager is attached; do not hide
that remaining dependency or estimate garbage-collector/allocator metadata.
Use checked errors for nil, stopped or uninitialized state.

Unit tests must prove exact component summation, optional-watch deltas, growth
after variables/clauses, both bookmark layers, facade-over-core deltas, zero
proof bytes and lifecycle errors. Fuzz arbitrary exact tautological arenas and
require nondecreasing total owned capacity, internally consistent totals and
correct facade literal/count storage without native code.

Execution checkpoint: core and facade memory reports now cover all owned Go
storage and retain the watch-entry switch. The standalone capacity fuzzer
passed 5,000 executions. The uncached full suite passes in 10.409s, race in
58.737s, and vet is clean, covering both solver2 memory queries for the
then-current no-proof state. The later proof-information checkpoint extends
the same report with exact attached-manager ownership.

## BSAT solver2 learned-reduction completion checkpoint

Replace the generic learned-clause deletion heuristic with the exact active
policy in `sat_solver2_reducedb`. Track start, delta, ratio and reduction count;
on each call advance the current limit with checked saturation. Score learned
clauses as capped inverse LBD in the high bits plus shifted activity, select the
ascending cutoff for the configured top ratio and preserve every cutoff tie.
Also preserve the native recent suffix test, unit clauses and clauses locked as
variable reasons. Skip the whole deletion if fewer than one sixth of the old
learned database is eligible.

When deletion proceeds, compact the complete owned clause arena. Build an
old-to-new handle map, retain original and selected learned clauses in their
existing order, remap every variable reason and unit handle, and rewrite each
watch bucket in place order while dropping deleted handles. Update current
learned/deleted counts atomically. Keep facade caller handles independent of
core compaction. Expose a structured reduction result and a checked policy
setter through both concrete solver types; a zero learned limit still disables
automatic reduction.

Unit tests must pin cutoff selection, locked-clause preservation/remapping,
compacted identities, limit progression, skip behavior, facade composition,
post-compaction invariants and satisfiability. Fuzz independently generated
LBD/activity arrays and ratios against a separate selection model, then solve
the compacted formula. No native oracle is required.

Execution checkpoint: native scoring, threshold gates and full handle
compaction are concrete. Exact units pass and the independent reduction fuzzer
completed 5,015 executions. The uncached full suite passes in 10.290s, race in
58.855s, and vet is clean.

## BSAT solver2 proof-information attachment checkpoint

Attach the already validated concrete `SatProofInfoManager` to `SatSolver2` as
the source's optional `pPrf2` mode. Require attachment after the exact facade
formula is loaded and before the first solve, so every logical problem clause
can retain its stable caller handle and every subsequently learned clause can
receive one dense proof ID. Do not add an interface or a native logger.

For each learned or final conflict, collect the complete implication-clause
dependency closure in deterministic order. OR original caller-handle bits and
prior learned records into a new manager record. A conservative dependency
superset is acceptable for this bitset mode, but it must never produce an
unsound core: expose a result only after rebuilding the selected facade
clauses in a fresh Go solver and proving them UNSAT under the assumptions from
the last solve. Preserve the manager's descending handle order. Empty cores
are valid only when the assumptions alone are contradictory.

Repeated solves must remove the prior non-clause final record before search.
Owned bookmarks must clone the manager, assumptions and final-record state.
When learned reduction compacts clauses, compact proof records in the same
retained order and rewrite every surviving learned proof ID. Account for exact
owned proof storage in memory reports and validate manager record counts,
unique dense IDs and inactive chains in the solver invariant. Freeze formula
insertion after attachment so a fixed bit width cannot silently omit roots.

Test a four-clause learned UNSAT derivation, assumption conflicts, contradictory
assumptions with an empty root core, repeated SAT/UNSAT solves, proof limits,
bookmark/rollback, memory accounting and learned-database compaction. Fuzz
small binary formulas against exhaustive assignment enumeration; independently
enumerate every returned core. No C helper, executable, compiler, generated
fixture or retained fuzz corpus is permitted.

Execution checkpoint: the concrete optional attachment, final-record lifecycle,
verified facade-handle cores, bookmark ownership, proof memory and learned
compaction are implemented. Exact tests pass and a fresh 5,000-execution
Go-only small-formula fuzz campaign passes. The full suite passes in 10.497s,
race in 61.345s, and vet is clean. This closes `pPrf2` core attachment while
leaving the next packed `pPrf1` attachment checkpoint, tagged binary clause
storage and remaining proof/interpolation entry points explicit partial gaps.

## BSAT solver2 exact resolution-proof attachment checkpoint

Attach the validated concrete `SatProofDAG` as solver2's optional packed
`pPrf1` replacement. Copy exact facade roots and variable partition state only
after formula loading and before search; freeze later formula/variable growth.
Store a dense DAG node handle on every learned clause rather than reproducing
packed pointer/flag words.

Reconstruct each learned and final clause from the live implication graph as a
real resolution chain. Begin with the conflict clause and repeatedly resolve
the highest-trail literal absent from the target against its reason. Use the
standalone checked Go resolver for every step and reject a chain that has zero
or multiple complementary pairs, repeats a resolved variable or fails to reach
the exact target. Start problem references from copied, unsimplified facade
clauses so root-level shortening in the CDCL arena is justified by explicit
unit-resolution steps. Carry each pivot's variable-partition bit.

Full verification must replay the reachable DAG using only copied facade roots
and require the final resolvent to equal the retained final-conflict clause.
Core extraction converts reachable zero-based roots back to stable facade
handles and then independently proves that subset UNSAT under the last
assumptions. Directly contradictory assumptions legitimately have no formula
proof node. Clone all DAG/root/assumption state in bookmarks, account for it in
proof memory, and stop it idempotently. Learned-database compaction must reduce
the DAG to nodes reachable from surviving learned clauses and atomically remap
their proof handles; antecedent nodes for deleted clauses remain when needed.

Test learned empty-clause derivation, assumption clauses, root-simplified
originals, partition flags, simultaneous pPrf1/pPrf2 logging, repeated solves,
bookmark rollback, limits, assumption-only conflicts, memory and forced search
reduction. Fuzz small formulas against exhaustive satisfiability; replay every
UNSAT proof and exhaustively validate each derived root core. Use no native
source, compiler, runtime oracle, generated fixture or retained fuzz corpus.

Execution checkpoint: exact checked chain reconstruction, full DAG replay,
independently verified facade cores, lifecycle/memory ownership and atomic proof
reduction are concrete. Exact tests pass, including a reduced pigeonhole proof,
and a fresh 5,000-execution Go-only fuzz campaign passes. The full suite passes
in 10.510s, race in 59.548s, and vet is clean. Packed proof checking/core
attachment is complete.

## BSAT solver2 attached interpolation and diagnostics checkpoint

Complete `Sat_ProofInterpolant` and `Sat_ProofInterpolantTruth` without a new
interpolation engine. Require a checked exact final empty-clause DAG with no
assumptions, collect its reachable roots/nodes, reorder used problem roots into
stable A-then-B groups and replay every verified resolvent in topological order
into an owned `SatProofStore`. Feed that store to the validated concrete AIG and
truth interpolators. Freeze root and variable partition metadata when proof
logging attaches. Reject assumption proofs because assumptions are not caller
partition clauses.

Port `sat_solver2_simplify` as checked root propagation and
`sat_solver2_verify` as retained-model validation over both logical and exact
facade clauses. Return errors rather than printing failed clauses. Treat copied
facade `Clause` lookup as the owned replacement for raw diagnostic handle
inspection. Dense typed clause/proof handles intentionally replace packed
pointers while preserving observable ordering and identity.

Test an UNSAT A/B split exhaustively for `A => I`, unsatisfiability of `B && I`
and AIG/truth agreement. Fuzz variable IDs, polarities and explicit shared order
for both representations. Test pre/post-solve simplify/model lifecycle and
deliberate model corruption. Run full, race, vet and repository policy gates.

Execution checkpoint: reachable proof-store replay, both solver-attached
interpolants, checked simplify/model diagnostics and owned clause inspection are
concrete. A fresh 5,000-execution interpolation campaign passes. The final full
suite passes in 10.429s, race in 59.416s, and vet is clean. Every active function
and public declaration in `satSolver2.c` and `satSolver2.h` has a concrete tested
Go path, promoting both files from partial to validated.

## BSAT solver2 interpolation composition checkpoint

Account for all six live definitions in `sat/bsat/satSolver2i.c` without
creating a second solver hierarchy. Absorb `Int2_ManStart`, stop, result read,
chain start and chain resolve into the concrete standard `SatAigInterpolator`:
retain copied global-variable order and inverse classification, A-root
disjunctions of shared literals, B-root constant one, learned-result reuse,
A-local OR and shared/B AND. Existing owned cleanup, limits, cancellation and
lifecycle rules apply; do not add interfaces or expose mutable clause scratch.

Represent `Gia_ManInterTest` by a checked one-output combinational graph
transformation. Copy only the output cone into a private structurally hashed
GIA, retain all original inputs in order, append one claimed-output input and
emit equality between the copied output and that claim. Support constants,
buffers, ANDs, XORs and muxes according to their Boolean semantics; reject
sequential or non-single-output input. Bound source mapping and result objects,
poll cancellation and leave the source unchanged on every path.

Unit tests must exhaust the relation for mixed node kinds, constants and
dangling logic and cover ownership, malformed shapes, exact resource failures
and cancellation. Fuzz arbitrary small graphs and exhaustively compare every
source assignment and both claimed values against independent source
evaluation. No native source, compiler, runtime oracle or generated C fixture
is permitted.

Execution checkpoint: standard incremental rules compose through the already
validated `SatAigInterpolator`, and `InterpolateOutputRelation` supplies the
self-relation diagnostic as an owned GIA transformation. Exact units and a
fresh independent 5,000-case graph fuzz campaign pass. The uncached full suite
passes in 10.217s, race in 58.872s, and vet is clean, promoting
`satSolver2i.c` to validated without a native oracle.

## BSAT packed proof-DAG checkpoint

Port the enabled surface of `sat/bsat/satProof.c` as one concrete owned proof
DAG. Replace packed `(handle << 2) | flags` values with a tagged reference that
names exactly one earlier proof node or problem-clause index and carries the
partition-pivot bit. Dense zero-based Go handles may replace byte/word offsets,
but preserve antecedent order and the source's exact record-word calculation.
Return copied nodes; resizing may only shrink a nonempty antecedent list.

Implement both stable postorder and ascending-handle reachability without
recursion, unique used counts, first-encounter core collection and sorted core
derivation. Proof reduction must be atomic: collect reachable nodes, build a
complete old-to-new mapping, clone and remap every node antecedent, then update
roots and the pivot together. Clear transient labels during compaction. Reject
invalid forward references, absent pivots and malformed roots with checked
errors. Bound nodes, total antecedents and clause IDs, poll cancellation, and
make close/stop idempotent. Use slices for all observable iteration order and
do not introduce an interface.

Verify reachable derivations by resolving clauses with the standalone Go proof
checker. Unit tests must cover exact record sizing, ownership, labels, shrinking,
postorder and sorted collection, duplicate roots, core ordering, remapping,
resolution success/failure, atomic errors, exact limits, cancellation, lifecycle
and at least a 10,000-node chain. Fuzz acyclic DAGs against an independent Go
reachability/core model and require reduction to preserve the core. No native
source, compiler, runtime oracle or generated C fixture may be used or retained.
The source-commented alternate reductions and checker/interpolator blocks under
`#if 0` are not live entry points; rely on the separately validated concrete Go
checker and truth/AIG interpolators rather than duplicating dormant APIs.

Execution checkpoint: `SatProofDAG` now provides the complete enabled behavior
with typed references, iterative deterministic traversal, atomic remapping,
standalone resolution verification, resource limits and cancellation. Exact and
deep units pass, and a fresh independent-model fuzz campaign passed 5,000
executions while checking core preservation through reduction. The uncached
full suite passes in 10.328s, race in 58.537s, and vet is clean, promoting
`satProof.c` to validated without C source or a native runtime oracle.

## BSAT truth interning checkpoint

Port `sat/bsat/satTruth.c` and `satTruth.h` together as an owned concrete truth
store. Preserve one-through-sixteen-variable elementary tables, constant-zero
and constant-one handles, first-bit complement normalization, XOR-one signed
handles, insertion-stable deduplication and copied function/variable reads.
Replace Vec_Set page offsets with documented dense even handles and bucket
rehashing with a lookup-only map that is never iterated; slices remain the sole
source of deterministic IDs. Add explicit word/entry bounds and checked close
semantics, returning statistics rather than printing during release.

Represent every header inline operation with concrete uint64 slices: equality,
zero/one tests, clear/fill/not, copy/AND/OR, negated operands and conditional
negation. Dimension errors must be atomic and returned slices must retain the
native in-place contract. Test exact variable bits at every supported width,
constant and complement insertion, deduplication, caller/storage ownership,
limits and lifecycle. Fuzz normalized signed insertion against an independent
word model without C, then run full, race and vet gates before promoting both
inventory entries.

## Decomposition-manager lifecycle checkpoint

Port `bool/dec/decMan.c` as a concrete owned manager around the existing MVC
manager, private cube/literal scratch and the exact four-input NPN tables used
by later factoring code. Preserve `Extra_Truth4VarNPN`'s permutation generation
order, ascending representative discovery, phase/output-complement encoding,
permutation indices and class bytes—including observable sparse class entries
rather than silently normalizing them. Cache only immutable generated state and
copy every table into each manager; avoid interfaces and shared mutable arrays.

Expose checked lookup and owned table snapshots for verification. Replace
native frees with checked close semantics. Exhaustively compare all 65,536
canonical entries with an independent NPN enumerator and reconstruct every
truth from its stored phase/permutation. Test multiple-manager isolation, nil
and repeated close, exact constant behavior and fuzz arbitrary orbit
transformations without C or permanent generated code. Promote only after
focused, full, race and vet gates pass.

## SFM mitigation placeholder checkpoint

Account for every active function in `opt/sfm/sfmMit.c`, which is part of the
default SFM module despite containing only temporary placeholders. Use a
concrete `SFMMit` pointer type, concrete integer/byte slices and `uintptr` only
for library, network, object and gate types whose subsystems are not yet
ported. Do not introduce an interface or pretend that a live timing manager
exists: start must always return nil, all readers/decisions must return zero,
and lifecycle/load/timing mutations must remain no-ops.

Test nil and nonzero opaque handles, negative and positive integers, duplicate
slice entries and overlapping slice arguments. Preserve every caller-owned
slice byte-for-byte and fuzz the full placeholder family without a native
oracle. Record all thirteen definitions and verify that the source remains in
the default `module.make`; promote only after focused, full, race and vet gates
pass.

## GIA DFS and diagnostic completion checkpoint

Complete `giaDfs.c` with concrete owned traversal results rather than exposing
manager traversal epochs or process-global diagnostics. Cover increasing-level
and descending-level grouping, sequential closure through discovered register
inputs, TFI and ordered-static-fanout TFO postorder, CI support, leaf-bounded
AND cones, all-gate collection, combined gate/CI postorder, and union
support/cone cardinalities. Traversals must be iterative, deduplicate roots,
preserve native encoded-fanin and terminal ordering, and explicitly traverse
native mux controls while leaving source scratch unchanged.

Treat the three test drivers as behavior to validate, not APIs for elapsed time
or stdout. A Go-only regression must execute whole-output shared DFS, count
gates below a structural-support threshold and collect every primary output's
sequential closure. Retain ordered immutable datasets for the algorithmic paths;
add independent constructor-ledger/model fuzzing, malformed-root and exact
bound checks, buffers/extended nodes and deep cones. Promote only after every
active definition has a disposition and focused, full, race and vet gates pass.

## AIG direct/reverse timing checkpoint

Port `aigTiming.c` as a concrete owned `AigTiming` value rather than attaching
mutable timing vectors and level buckets to `AigManager`. Copy object metadata
and direct levels into the snapshot, compute the output-driver maximum plus the
requested margin, and compute internal reverse levels from fanouts. Stable
topological IDs permit a descending iterative pass: every internal object has
minimum reverse level one, and each internal fanin receives its fanout's level
plus one. CIs, COs and constant one retain reverse level zero.

Expose copied forward/reverse arrays, checked per-object reverse and required
levels, clearing, fresh reverse-level calculation, atomic refresh and a
non-printing verifier. Required level is `maximum + 1 - reverse`, including
negative slack for sufficiently deep dangling cones. Preserve native's
fanout-storage precondition by rejecting buffers. Replace both native
incremental update routines with full atomic snapshot refresh: public Go graph
edits already commit with valid direct levels, and a complete recomputation
avoids manager-attached scratch and partially updated level buckets.

Validation pins 32 native graphs across complemented phases and margins zero
through three, comparing every direct, reverse, required and newly recomputed
level on shared, reconvergent, output-driven and dangling cones. Add nil,
negative, buffer, exact-limit and cancellation cases; prove returned slices are
owned, failed refresh is atomic, clearing is detected and refresh repairs it,
exercise graph growth, and use a 10,000-input chain. Fuzz arbitrary hashed AIGs
against an independent fanout recurrence. Keep the fixture under
`test_vectors/`, record source and fixture hashes, and promote only after
immutable-fixture verification plus full, race and vet gates pass.

## AIG structural-hash checkpoint

Keep the production structural hash as the concrete lookup-only Go map already
owned by `AigManager`; do not introduce a table interface or emulate native
linked buckets in graph operations. Add the `Aig_TableLookupTwo` spelling as a
nonmutating lookup with native constant, identical and opposite-literal cases.
Construction and recursive deletion continue to provide insertion/deletion,
and `AndCount` remains the authoritative live-entry count.

Replace print-only profiling with an owned `AigHashProfile`. Accept an explicit
positive table size and reproduce native's XOR/multiply hash formula and exact
bucket occupancies for every live AND. This also makes native resize observable:
profile the same graph at its old and new sizes without exposing or controlling
Go runtime map capacity. Validate the manager first, bound the object span,
check cancellation on every object and return no partial profile on failure.

Validation pins 32 phase cases with all trivial, existing and missing lookup
results, live counts, and nonempty bucket/count pairs before and after native
resize. Add invalid literal and dimension cases, exact limits, cancellation,
owned-result mutation and recursive-deletion cleanup. Fuzz arbitrary hashed
construction traces and verify every retained lookup plus total bucket count.
Record fixture provenance and run immutable-fixture verification, full, race and
vet gates.

Treat `Aig_TableClear` as a private representation boundary, not a public
manager mutation. Native temporarily destroys the table while leaving hashed
nodes live; allowing that state in the ordinary Go manager would break `Check`
and all subsequent structural construction. Forward retiming therefore copies
into its private unhashed net, performs raw intermediate construction there,
and dependency-normalizes into a new fully hashed checked manager before any
result escapes. With that lifecycle validated by the retiming oracle and
semantic tests, all `aigTable.c` behavior is represented without adding an
unsafe public state or a table interface.

## Fixed AIG diagnostic checkpoint

Port the sole `aigTest.c` routine as `BuildAigProcedureTest`, returning a
concrete independently owned `AigManager`. Preserve its three CIs, MUX, AND,
two ordered COs, one-register suffix and cleanup/check sequence exactly. Do not
retain the hard-coded `aig_test_file.blif` write; callers can choose their own
serialization and destination using the eventual I/O layer.

Pin the complete native arena, CI/CO vectors, register count and check result.
Compare every object type, literal, reference count, level and phase, prove two
calls return independent storage, and exhaustively evaluate all eight input
assignments against direct MUX/AND semantics. This constructor has no inputs to
fuzz; its underlying MUX, AND, manager and serializer surfaces retain their own
dedicated fuzz campaigns. Record fixture hashes and promote only after
clean immutable-fixture verification plus full, race and vet gates pass.

## AIG sequential-cleanup checkpoint

Port the reachability half of `aigScl.c` directly on concrete `AigManager`
storage. Starting from all primary outputs, traverse fanins iteratively. Primary
inputs and constant one are fixed boundaries; encountering a latch-output CI
adds its corresponding latch-input CO, closing transitively over registers.
Use an ID-indexed register table so closure remains linear even for wide state.

Also port `Aig_ManRemap` as an owned CI-substitution transform. Require one
signed source CI or constant per CI, precreate every destination CI so
cross-register substitutions are valid, collapse buffers, structurally rebuild
ANDs and recreate COs in order. Preserve name/spec, register and constraint
partitions. Reject choice metadata rather than silently discarding it; enforce
object/work limits and cancellation before exposing a result.

Full cleanup operates on a clone: disconnect and tombstone unused latch-input
COs, sweep newly dangling logic in reverse topological order, then tombstone
the now-unreferenced latch-output CIs. Rebuild CI and CO slices as primary
prefixes followed by retained registers in source order. Basic cleanup keeps
all register slots, atomically patches each unreachable latch input to constant
zero, and sweeps dangling logic. Preserve C's return convention: basic cleanup
reports only the final sweep count, not recursive nodes deleted while patching.

Validate 32 native dependency graphs in both modes, comparing CI/CO arrays and
every arena slot, fanin, reference count, level and phase. Native full cleanup
can report `Aig_ManCheck` failure because its terminal removals do not update
global created/deleted counters; record that observation but require the Go
result's structural invariants to pass. Add nil, choice, malformed-limit,
cancellation, ownership and 10,000-level tests. Fuzz arbitrary three-register
dependency cycles against independently computed closure and one-step Boolean
semantics. Keep fixtures under `test_vectors/`, record hashes and require clean
fixture integrity verification plus full, race and vet gates.

Keep `aigScl.c` partial after this checkpoint. Duplicate-latch analysis and
reduction, SCC reporting, partitioned sweeping and the high-level
constant/equivalent-latch pipeline remain separate work. Choice-aware cleanup
must be added together with those representation transforms rather than
silently discarding equivalence chains.

## AIG forward-retiming checkpoint

Port all three routines in `aigRetF.c` as one concrete `AigManager` transform.
Preserve the recursive routine's observable fanin-zero-first short circuit with
an iterative stack, compute the register frontier without manager scratch, and
perform each requested forward-retiming step using un-hashed intermediate ANDs
and explicit buffers. Normalize the resulting forward references with a
dependency-driven copy into a fresh structurally valid manager. The public
operation must own its result and leave the source unchanged on success,
invalid input, cancellation or an exact resource-limit failure. Keep the
options and result concrete; do not introduce an interface.

Do not use the pinned entry point blindly as an oracle. Its normal register
setup writes CIO IDs into the same union field that `Aig_ManRetimeMark` asserts
is null, and after a successful step its `Aig_ManDupOrdered` reads future CIs
before they have been copied. Preserve the original retiming code in an
stand-alone regression dataset but avoid the unrelated CIO setup conflict and
substitute only a final duplicator that precreates retained CIs before
DFS-copying CO cones. Hash the exact `aigRetF.c` source and fixture, and explain
the correction in both the coverage ledger and permanent vector documentation.

Pin 32 phase and one-through-three-step cases, comparing terminal order,
register count and every object field. Independently enumerate transition
languages after one warmup frame so arbitrary values of newly introduced state
cannot create false mismatches, and compare longer warmed traces in fuzzing.
Add nil, malformed, combinational, object-limit, cancellation, source-ownership
and 10,000-level tests. A zero Go step limit is a bounded default rather than
native's accidental do-while single step. Promote the source only after clean
fixture integrity verification and focused, full, race and vet gates pass.

## AIG DOT-rendering checkpoint

Port `aigShow.c` as concrete deterministic byte generation plus explicit file
output. Reproduce native DOT text exactly: header, empty level spine, title and
statistics blocks, output/input/constant shapes, object-ID ordering, rank
placement, invisible ordering edges, complemented dotted edges, ordinary bold
edges and selected filled nodes. Accept the HAIG flag for source compatibility,
but document that the active native implementation also ignores it because all
HAIG-specific rendering blocks are commented out.

Replace temporary `fMarkB` and CO-level mutation with owned Boolean selection
and local level computation. Preserve the native default rejection above 200
AND nodes, add explicit object/byte bounds and cancellation, and return typed
errors instead of printing. The `Aig_ManShow` equivalent derives `.dot` by
replacing the final filename extension, writes the file and returns its path;
it must not launch an external viewer or hide a subprocess side effect.

Pin byte-exact native output for all 16 relevant phase combinations on a graph
containing constant output, dangling logic, ordered COs and one/two bold nodes.
Add file/path/error/immutability tests, a 10,000-node case and fuzzing that
checks determinism and complete node emission across varied chains. Store the
fixture under `test_vectors/`, hash fixture and exact C
source, and require clean fixture verification plus full, race and vet gates before
promotion.

## AIG sequential ternary-state checkpoint

Port the deterministic core of `aigTsim.c` using the existing concrete
`AigTernary` encoding. Preserve uint32-overflow state hashing and the 128-prime
cycle, packed two-bit register states, zero initialization, unknown primary
inputs, complemented Kleene AND propagation, 300 exact transitions and native
widening thereafter. Compute every latch input before committing any latch
output so all registers advance simultaneously. Retain the repeated state in
the owned state collection, union all collected states, and return copied
register classifications and CI substitutions instead of manager scratch or
printed diagnostics. A nil mapping denotes the native no-constants result.

Bound object visits, rounds and actual retained state words, poll cancellation,
reject unsupported buffer/XOR storage and leave the source unchanged on every
path. Do not preallocate from an untrusted round limit. Pin 32 native mappings
covering fixed zeros, delayed changes, toggles, complements, unknowns and null
results, plus 32 hash records spanning zero through 140 words. Add exact-limit,
combinational, simultaneous-update and 10,000-node tests. Fuzz small sequential
machines and exhaustively enumerate their Boolean reachable states to prove
every reported ternary constant sound.

Keep `aigTsim.c` partial after this checkpoint. Native print-only formatting,
unused random ternary helpers and multivalued `Saig_MvManSimulate` integration
remain separate work. The ordinary `Aig_ManConstReduce` branch composes owned
ternary analysis, CI remapping and full sequential cleanup until stable,
requiring register-count progress and returning an original-register
projection. Validate 32 complete final arenas and fuzz 32-frame trace
equivalence from zero state.

## Empty translation-unit checkpoint

Do not manufacture Go files or exported names for C translation units that
contain no declarations, definitions, data or executable behavior after
comments and preprocessing scaffolding. Before marking one validated, inspect
its complete active text, require that it contains only includes and namespace
delimiters, and prove that its local SHA-256 equals both the inventory and the
pinned-reference source. Record the decision directly on the exact coverage
entry with empty symbol and Go-file lists.

The audited batches contain 49 such files across `aig`, `base`, `bdd`, `bool`,
`map`, `misc`, `opt`, `proof`, `sat` and the source root. Combined with
`aig_.c`, 50
empty translation units are intentionally represented by absence rather than
boilerplate Go files. Recheck this classification whenever the pinned source
identity changes.

Keep callable native stubs distinct from empty translation units. When a file
exports an empty routine or a constant-return routine, provide a concrete Go
spelling if its signature can be represented without pulling in unported
manager types. Preserve the exact no-op/constant behavior, test nontrivial
arguments where present, and document that this is compatibility behavior—not
an implementation of the subsystem suggested by the filename. Defer stubs
whose signatures require unported concrete managers rather than weakening them
to `any` or an interface.

The concrete stub batch also covers all callable bodies in `giaSim5.c`,
`extraUtilCfs.c`, `extraUtilGen.c`, `plaMerge.c`, `mpm.c`, `abcRunGen.c`,
`abcEco.c`, `cbaCba.c` and `giaIiff.c`. Use `*GiaManager` where the concrete
graph already exists, fixed string arrays for filename lists, and `uintptr`
only for opaque frame, simulation-data, PLA, CBA and vector pointers that are
never inspected. Pin varied reviewed records, test graph nonmutation and fuzz
every argument. Freeze results for the exact pinned `mpm.c` source even when the
default archive omits it; never retain temporary comparison code. A null or
constant result is validation only for these exact pinned bodies and must be
re-audited if their source identity changes.

The same callable-stub batch covers both empty lifecycle hooks in `base/test/test.c`.
Represent their uninspected ABC frame as `uintptr`, exercise zero and varied
opaque values in unit and fuzz tests, and require a reviewed immutable dataset,
independent invariants and source-hash evidence before validating the file.

## Fast-simulation defaults checkpoint

Port both complete initializers in `opt/fsim/fsimCore.c` as constructors for
concrete Go value records. Preserve every integer default and translate semantic
zero/one flags to booleans; do not introduce a parameter interface or expose
shared mutable defaults. Pin every native field in one permanent oracle and
test repeated-construction isolation. Because the constructors accept no input,
record that fuzzing has no meaningful domain rather than manufacturing one.
Require exact source/fixture hashes and focused, full, race and vet
gates before validating the source.

Port the complete ternary state engine in `opt/fsim/fsimTsim.c` over the concrete
`AigManager`. Preserve two-bit zero/one/unknown values, unknown primary inputs,
zero register initialization, complemented edges, AND semantics, packed state
hashing with uint32 overflow, the 128-prime cycle, and termination after retaining
the first repeated state. Return independent register-order state slices and an
owned diagnostic string rather than native linked allocations and stdout writes.
Use deterministic bucket slices for lookup, with explicit state-word, state-count
and cancellation bounds.

The pinned FSim files are not in the default archive and no longer compile
against current AIG names. If historical comparison is needed during porting,
freeze the result as a reviewed dataset and remove all temporary compatibility
code immediately. Correct their uninitialized ternary storage by zero-initializing
FSim allocations: the XOR setters require zero backing bits, so heap-history
behavior is not a compatibility target. Avoid graphs whose CI traversal violates
the legacy PI-first assumption. Pin fixed points, PI unknown propagation,
complement cycles, hashes and the 16-to-17-register packing boundary. Fuzz small
sequential graphs and require unique retained states before the repeated final
state, plus exact source/fixture hashes and all gates.

## MVC manager checkpoint

Begin the multi-valued-cover package with the complete allocation lifecycle in
`misc/mvc/mvcMan.c`. Use a concrete noncopyable `MvcManager` owner that later
cube and cover values can reference. Let Go allocation replace the four native
fixed-size pools; do not expose allocator callbacks or an interface. Make close
checked, returning errors for nil and repeated release instead of reproducing
invalid native frees.

Pin multiple simultaneous native managers and require all manager/pool pointers
to be non-null and independent before reverse-order release. Test independent Go
owners, successful close, use-after-close detection and nil/double-close errors;
fuzz varied manager counts and early-close order. Require exact source/fixture hashes, clean dataset integrity verification and full, race and vet gates before
validating the source. Cube/cover allocation and algorithms remain later MVC
checkpoints rather than fabricated manager behavior.

Next implement all of `misc/mvc/mvcCube.c` with a concrete `MvcCube` owning a
packed `[]uint32`, plus `Mvc_CoverAlloc` from `mvcCover.c` as the checked fixed-
width layout constructor. Use deterministic zeroed Go storage in place of the
native pool's indeterminate payload, while preserving 32-bit word/unused-tail
calculation, owned duplication, bit access, tail masking and the exact adjacent-
pair transform in `Mvc_CubeBitRemoveDcs`. Checked close replaces pool recycling
and rejects nil, repeated, stale-manager and dimension misuse.

Pin zero through 127-bit native layouts and transformed word strings across all
word boundaries. Verify clone isolation and invalid accesses, then fuzz arbitrary
widths through 512 bits against an independent per-pair model. Validate
`mvcCube.c` only after all gates; keep `mvcCover.c` partial until clone/dup/free,
mask, literal-array and cube-array operations are implemented.

Then complete `mvcApi.c`, `mvcCover.c` and `mvcList.c` as one cohesive ownership
layer. Store native list order in a concrete `[]*MvcCube`; do not introduce a
list interface or expose setters that can separate count, head and tail. Provide
checked head/tail insertion, ownership-transferring deletion, empty clone,
ordered deep duplicate, close, reusable mask and literal-count allocation, and
list-to-array/array-to-list conversion through an exact atomic pointer
permutation. A listed cube and the cover-retained scratch mask must reject direct
close, while deletion transfers a cube back to the caller. Preserve the two
empty literal-update bodies and the native zero-bit tautology result.

Pin full workflows at zero, two, 31, 33 and 65 bits, including direct native
list helpers, accessors/setters, order reversal, clone/duplicate, allocation
reuse, deletion, empty/tautology construction and binary-buffer recognition.
Add ownership and failure-atomicity unit tests plus model-based fuzzing of
arbitrary cube permutations and deep copies. Promote all three sources only
after clean fixture verification, source/fixture hashes, full tests,
race and vet pass.

Complete `misc/mvc/mvcCompare.c` next with five concrete `MvcCube` comparison
methods. Preserve high-word-first unsigned integer order, descending population
before integer tie-breaking, selected-bit order, unselected-bit order using the
native OR-with-mask form, and unselected-then-selected partition order. Compute
population directly from packed words instead of carrying the native cached
`nOnes` header. Validate equal dimensions and live ownership. Return zero for
equal partition inputs rather than reproducing the native assertion; this makes
the operation a total comparator without changing valid unequal calls.

Pin every comparator over deterministic triples at one-, two- and multiword
boundaries through 127 bits. Independently test zero-bit/equal behavior and
dimension/lifetime failures, and fuzz arbitrary packed triples through 512 bits
against direct word and popcount models, including antisymmetry. Require clean
dataset verification, all hashes and full, race and vet gates before promotion.

Port `misc/mvc/mvcSort.c` and `mvcContain.c` together. Replace the arbitrary C
comparison callback with a concrete `MvcCubeOrder` enum covering exactly the
five MVC comparators; do not add an interface. Preserve the recursive half split
and right-half-first merge behavior on equal keys, since pointer order among
duplicates and mask-equivalent cubes is observable. Containment must sort by
descending population/integer order, remove adjacent duplicates, then remove
every later subset covered by each larger-population group while closing only
the removed owned cubes. Direct popcount replaces the cached size pass from
`mvcUtils.c`, leaving that larger source partial.

Pin all five pointer-identity sort orders plus exact containment survivors at
twelve one-, two- and multiword widths. Add invalid mode/mask and failure-
atomicity tests, zero-bit duplicate handling, removal lifetime checks and
idempotence. Fuzz arbitrary covers against an independent maximal-set model,
proving that survivors are unique, pairwise noncontained, drawn from the input
and cover every original cube. Require clean fixture verification, all source/
fixture hashes, full tests, race and vet before promotion.

Complete `misc/mvc/mvcLits.c` with concrete frequency and selection methods.
Preserve highest-index scanning for any repeated literal, lowest-index tie
selection for best/worst repeated literals, first-cube/first-bit order, total
population and the native `Mvc_CoverIsOneLiteral` quirk that accepts one cube
with zero or one asserted bit. Return an owned frequency slice. Best-literal
cover construction must clone layout and ownership, require exactly one mask
cube and reject the native negative-bit-shift precondition when no selected
literal repeats.

Pin masked and unmasked results across zero-, one-, two- and multiword widths,
including total counts, selected one-cube covers and the five at-most-one
predicate cases. Add empty cover, invalid dimension/lifetime and invalid-simple-
cover tests. Fuzz arbitrary covers and masks against an independent frequency
model, checking every selection, total, first literal, predicate and constructed
cover. Require exact hashes, clean dataset integrity verification, full tests, race and vet.

Complete `misc/mvc/mvcOpAlg.c` as concrete checked cover operations and add
`Mvc_CoverSupport` from `mvcUtils.c`. Preserve Cartesian OR order for algebraic
multiplication, source order for subtraction, support-disjointness and native
directional equality: every left cube must occur on the right, while extra
right cubes and multiplicity do not matter. Accept overlapping multiplication
as native computation does, but omit its incidental global stdout warning.
Return an owned support cube rather than exposing or mutating cover scratch.

Pin product/subtraction counts and word digests, disjoint and overlapping
supports, directional equality and duplicate-insensitive membership at nine
widths through 127 bits. Unit-test ownership, dimensions, closed values and
empty covers. Fuzz arbitrary bounded covers against independent word-set,
Cartesian-OR and ordered-filter models, including overlapping supports.
Promote `mvcOpAlg.c` only after exact source/fixture hashes, clean
dataset verification, full tests, race and vet pass. Keep `mvcUtils.c` partial
until its common-cube, pair-difference, remap, column, inverse, cofactor,
expansion, transpose and remaining routines are ported.

Complete the active surface of `misc/mvc/mvcOpBool.c` with a concrete
`BooleanOr` cover method. Preserve exact left-then-right list order, duplicate
and contained cubes, and independent deep-copy storage; do not introduce
containment or algebraic simplification. The native Boolean AND and equality
sketches are excluded by `#if 0`, so account for them in coverage notes without
inventing callable Go APIs.

Pin zero through 127-bit native concatenation counts, complete ordered digests
and unchanged source-cover digests. Unit-test empty inputs, dimension and
lifetime errors, source/result mutation isolation, and retained duplicates.
Fuzz arbitrary packed covers against a direct concatenation and deep-copy model.
Require exact source/fixture hashes, clean fixture verification,
full tests, race and vet before validating the source.

Complete the compiled surface of `misc/mvc/mvcPrint.c` with concrete
string-returning cube and cover methods rather than a writer interface or
global stdout mutation. Preserve exact headings, raw bit order, newline and
literal-count formatting. For binary rendering, preserve pair mappings
`01 -> 0`, `10 -> 1`, `11 -> -`, ignored odd tail bits and the iterator's
early termination at a `00` pair. Do not expose the multivalued routines under
native `#if 0`.

Pin a complete native transcript at zero-, one-, two- and multiword boundaries
through 127 bits, including allocated literal arrays, zero-bit blank cube lines
and early binary termination. Unit-test closed owners and malformed retained
scratch; fuzz arbitrary packed cubes against independent per-bit and per-pair
models. Require exact source/fixture hashes, clean native
fixture integrity verification, full tests, race and vet before validating the source.

Before algebraic division, port the common-cube/support dependency slice from
`misc/mvc/mvcUtils.c`: support intersection, binary support size/membership,
common-cube extraction, cube-free checking and normalization, one-cube cover
construction and directional support containment. Return owned cubes rather
than exposing the retained native mask. Preserve all-ones intersection for an
empty positive-width cover, floor division of odd bit counts into binary
variables, and in-place cube order during normalization.

Pin per-variable membership, common/intersection cube digests, cube-free state,
normalized cover order, common-cube covers and both containment directions at
paired word boundaries through 126 bits. Unit-test empty covers, invalid
variables, dimensions and closed owners. Fuzz arbitrary empty/nonempty covers
and odd widths against independent union/intersection models. Keep
`mvcUtils.c` partial for pair differences, remapping, inverse/cofactor/flip,
quantification, expansion, transpose and remaining transforms; require exact
hashes, clean fixture verification and full/race/vet gates for this checkpoint.

Finish the active `misc/mvc/mvcUtils.c` surface with ordered uint8 cube-pair
differences, fill-and-copy remapping, in-place inversion, owned don't-care
removal, cofactor, adjacent binary flip, universal quantification, matrix
transpose and unused-tail validation. Preserve duplicate/reordered remap
columns, `-1` all-one columns, uint8 difference truncation, cube order and the
native quantification filter. Convert adjacency, range and nonzero-pair asserts
to checked errors. Report dirty tail state directly rather than reproducing a
diagnostic print routine that always returns one.

Pin every transform and ordered digest across one-, two- and multiword widths
through 126 bits. Fuzz arbitrary covers against independent population,
column, pair, complement and transpose models, including empty covers and
cross-word tails. Functions under native `#if 0` remain non-APIs. Promote the
whole source only after both utility oracle sets, exact hashes, clean native
fixture integrity verification, full tests, race and vet pass.

Close the MVC build manifest only after every one of its fifteen listed C
implementation units is validated in the single `abc` package; Go package
compilation replaces per-object make enumeration. Audit `mvc.h` separately.
Credit its packed constants, active concrete ownership/layout concepts,
bit/list/iteration macros and declarations to their tested Go representations,
but keep the header partial while it advertises missing-source legacy
Boolean/complement/sharp/reshape/merge APIs, dormant `Mvc_Data_t` state and
low-level allocator entry points outside the manifest. Do not fabricate those
APIs merely to close a header ledger row.

## NM name-manager checkpoint

Port `misc/nm/nmApi.c` and `nmTable.c` as one concrete checked `NameManager`.
Use `insdict.Dict` for both ID-to-entry and name-to-ordered-IDs storage; do not
expose a map, iterator interface or native hash buckets. Preserve unique IDs,
suffix concatenation, duplicate names across types, type `-1` lookup, two-type
fallback, and native namesake ordering: new duplicates follow the anchor, while
deleting the anchor rotates the oldest namesake into the anchor position.
Return owned strings/slices and make close, duplicate ID, missing deletion,
invalid ID/type and use-after-close checked errors instead of stdout messages,
assertions or arena lifetime coupling.

Replace native bucket resizing/profile output with deterministic dictionary
growth and a concrete owned profile summary. `Nm_ManReturnNameIds` incorrectly
returns only the head of each occupied ID bucket, dropping colliding entries;
the Go method must return every live ID in insertion order and document this
reliability correction. Pin reviewed compatibility store/find/type/unique-name/delete workflows,
namesake anchor rotation and resize-stable results, plus a fixture that exposes
the lossy native ID enumeration. Unit-test ownership, close and error atomicity;
fuzz operation sequences against an independent ordered model, including
duplicate names and repeated anchor/non-anchor deletion. Require exact source,
header and fixture hashes, clean fixture verification, full tests,
race and vet before promoting both C files; keep headers partial until all
advertised active declarations and internal representations are accounted for.

Checkpoint result: the implementation uses two concrete `insdict.Dict` indexes
and an explicit ordered-ID group per name. The complete public and internal
header surfaces are accounted for without interfaces or exposed buckets. The
immutable fixture covers every workflow above and demonstrates 37 returned bucket
heads for 64 live IDs; Go returns all 64 in insertion order. Independent
operation-model fuzzing passed 102,829 executions, clean fixture verification is byte-
identical, and full tests (7.373s), race (49.482s) and vet all pass. Therefore
`nmApi.c`, `nmTable.c`, `nm.h`, `nmInt.h` and `module.make` are validated.

## PARSE equation-parser checkpoint

Port the two implementation units selected by `misc/parse/module.make` as one
concrete Boolean-equation parser over the existing `AigManager` and `AigLit`
types. `parseEqn.c` depends on the not-yet-ported HOP manager only as its result
representation; preserve its grammar and Boolean semantics while compiling the
validated expression into the already owned, structurally hashed AIG. Fold the
two fixed-size raw stacks from `parseStack.c` into typed Go slices internal to
the parser. Do not expose a generic stack, parser interface, token interface or
writer callback.

Preserve constants `0`/`1`, prefix `!`, binary `*`, `^`, `+`, precedence
`! > * > ^ > +`, left associativity, arbitrary ASCII whitespace, explicit
parentheses, exact variable-name lookup and the native `a!b` shorthand for
`a*!b`. Reject adjacency otherwise, including a variable followed directly by
an opening parenthesis. Preserve first-match behavior for duplicate variable
names. Return an `ErrFormat`-wrapping byte-offset diagnostic instead of writing
to a `FILE *`, silently overflowing a 1,000-entry stack or following native
underflow into null HOP operands. Reject embedded NULs explicitly. Provide
concrete token/depth limits and cancellation, and parse to an owned postfix
program before touching graph state. Evaluate on a cloned manager and commit
only after success so syntax, cancellation, limit and construction failures are
atomic.

Pin reviewed compatibility accept/reject outcomes and complete truth tables for constants,
variables, every operator, precedence, associativity, nested negation,
whitespace, the `a!b` shorthand and malformed forms. Compare successful Go
results independently over all assignments, verify exact CI mapping and source
graph preservation on failure, and test duplicate/invalid variable names,
limits, cancellation, deep nesting and large flat expressions. Fuzz arbitrary
formula bytes against an independent recursive Boolean evaluator for accepted
programs and require that every returned literal passes exhaustive simulation.
Keep `parseCore.c` and public `parse.h` unstarted until the CUDD-backed BDD
surface is replaced; validate `parseInt.h` only when both typed stack concepts
and all its active declarations are accounted for. Require clean native
fixture integrity verification, exact source/header/fixture hashes, full tests, race
and vet before promoting the manifest units.

Checkpoint result: `ParseBooleanFormula` parses to a concrete postfix program,
evaluates on an `AigManager` clone and commits only on success. Seventeen native
truth tables and nine rejection outcomes match exactly; duplicate/invalid-name,
atomicity, cancellation, token/depth, 10,000-level and 5,000-term tests pass.
Independent semantic fuzzing passed 33,630 executions, dataset verification is
byte-identical, and full tests (7.332s), race (49.947s) and vet pass. Therefore
`parseEqn.c`, `parseStack.c`, `parseInt.h` and `module.make` are validated;
`parseCore.c` and `parse.h` remain unstarted for the later CUDD replacement.

## ST deterministic-table checkpoint

Port `misc/st/st.c` and `stmm.c` together as one concrete generic
`StableTable[K comparable, V any]` in package `abc`. Both C units implement the
same logical table; `stmm` differs only by recycling entries through a fixed
allocator and by exposing `clean`. Ordinary Go ownership and one checked
`Clear` operation replace both allocation strategies. Back the live-key index
with `insdict.Dict`; do not expose a map, comparison interface, hash interface,
raw `char *`, or `char **` slot.

Preserve insert-versus-replace results, lookup, delete with the stored key and
value, shallow-value cloning, clear/close, direct insertion, find-or-add and
find-without-add. Although direct insertion is normally called only for absent
keys, preserve its native duplicate behavior: newest direct duplicates shadow
older entries, deleting the newest reveals the preceding value, and iteration
can observe every live duplicate. Represent the native mutable record pointer
with a concrete checked `StableTableSlot[K,V]` referencing a heap-stable entry;
deletion, clear and close invalidate it. Provide a concrete cursor plus an owned
entry snapshot and a callback-based continue/stop/delete traversal; reject
external structural mutation during cursor/traversal use instead of allowing
dangling native generators.

Define Go enumeration as deterministic live insertion order, including direct
duplicates. Native bucket order, rehash reversal and move-to-front lookup are
incidental physical behavior and must not leak into the Go port; the native
`reorder_flag`, density and growth factor therefore need no public Go knobs.
Document this intentional ordering correction and test it across growth,
deletion, reinsertion and clone. Retain checked standalone spellings of the
portable string, integer and shifted-address hash calculations for oracle and
diagnostic parity, while normal table operation relies on `insdict` hashing and
Go comparable equality.

Pin both native variants through insert/update, lookup, find/find-or-add slot
mutation, direct duplicates, reveal-on-delete, copy, generator, foreach-delete,
rehash and `stmm_clean` workflows. Record native generator order only as
evidence of the replaced bucket semantics; compare Go enumeration against the
specified insertion-order model. Unit-test zero values, ownership, stale slots
and cursors, stop/delete traversal, nil/closed receivers and hash boundaries.
Fuzz arbitrary operation traces against an independent duplicate-aware ordered
model and repeatedly verify clone isolation and every visible key/value. Require
exact source/header/fixture hashes, clean fixture verification, full tests,
race and vet before validating both C files, both headers and `module.make`.

Checkpoint result: one `insdict.Dict`-indexed generic table now covers both
allocation variants, with concrete checked slots, cursors, snapshots and
continue/stop/delete actions. Native insert/update, direct duplicate, slot,
delete-reveal, clone, traversal, clean and hash workflows match; native bucket
order is recorded while Go deterministically uses insertion order. Independent
duplicate-aware fuzzing passed 104,762 executions, clean fixture verification is byte-
identical, and full tests (7.286s), race (49.211s) and vet pass. Therefore
`st.c`, `st.h`, `stmm.c`, `stmm.h` and `module.make` are validated.

## TIM hierarchy/timing checkpoint

Treat `misc/tim` as one concrete timing-manager family, but close its source
units independently. Preserve CI/CO identity, the distinction between primary
terminals and box pins, output-major float32 delay tables, native epoch caching,
30-bit traversal rollover and the unusual `Abc_MaxInt` truncation in arrival
propagation. Keep terminal, box, table and serialization records concrete; no
library/timing interface is needed. Continue returning owned snapshots and
checked errors in place of borrowed pointers, assertions and diagnostic prints.

The version-1 `timDump.c` stream is a deliberately narrow format. Integers are
signed 32-bit big-endian fields, while float32 values use the pinned native
host byte order (little-endian in the reference build). The stream does not
store box terminal indices, so save only canonical managers whose box inputs
are consecutive from CO zero and whose outputs are consecutive after the PI
prefix. It also omits black-box flags and cannot represent sparse table slots;
reject those states rather than silently changing them. Load into a private
manager and publish only after checking the version, every dimension/ID,
aggregate resource caps, exact consumption and canonical PI/PO totals. Test
hierarchy-only and full streams separately, exact native load/resave bytes,
all float bit patterns, every truncation boundary, trailing data, malformed
counts and valid generated managers. Keep permanent vectors in `test_vectors/`.

Finish the remaining family in dependency order. The exact
`Tim_ManBoxFindFromCiNum` boundary behavior is now covered for every PI/box
position, including the native final-box `-2` sentinel. Next port manager
duplication, unit-delay conversion, trim, reduce and two-manager alignment with explicit
old/new terminal and box maps. Preserve source ordering and copy IDs, but make
all failures atomic. Port library-derived table construction only after the
concrete library structures it consumes exist; do not introduce an interface
as a placeholder. Return owned arrival/required slices in place of raw arrays.
Replace print routines with deterministic strings or structured statistics.
Validate `timMan.c` only when every active symbol and correction is covered;
validate `timInt.h`, `tim.h` and `module.make` last when their full
type/declaration/build surfaces are accounted for.

Checkpoint result: all nine `timTime.c` operations and all seven `timTrav.c`
operations are validated, including native/corrected boxed timing fixtures.
`timDump.c` now round-trips empty, no-box and two-box native streams byte-for-
byte in both modes. Malformed-stream fuzzing passed 2,157,361 executions and
valid-manager fuzzing passed 1,305,440; clean fixture verification, exact hashes, full
tests (7.410s), race (52.445s) and vet pass. `timBox.c` is also complete after
the final CI-boundary lookup matched 294 native cases; association fuzzing
passed 307,916 executions and its final full tests (7.373s), race (49.361s) and
vet pass. Manager duplication, unit-delay conversion, primary timing
extraction, trim/reduce/alignment, atomic library configuration, deterministic
formatting/statistics and Go-owned cleanup subsequently completed `timMan.c`.
Its model, configuration and formatting fuzzers pass. The public/internal
headers and `module.make` are validated, so the complete active `misc/tim`
module is accounted for.

Complete `misc/mvc/mvcDivide.c` as a concrete algebraic-division family.
Preserve single-literal and single-cube source order; for multi-cube divisors,
preserve native in-place sorting of both inputs, outside-support grouping,
under-support merge order, failed-group remainder order and quotient extraction
from the final match. Return two nil covers without error for the native outer
count/support rejections. Reject empty and zero-literal divisors instead of
following native invalid indexing, and use full-width containment rather than
the fixed two-word macro in the single-cube path.

Pin sorted input digests, quotient/remainder digests, both specialized paths,
in-place literal reduction and wrapper rejections across two- and multiword
widths through 127 bits. Model-fuzz specialized paths and use constructed
dividend/divisor products to verify general algebraic reconstruction. Check
dimensions, lifetimes, invalid literals, source isolation and closed removed
cubes. Treat the unused static print verifier as test behavior rather than a Go
API; require exact hashes, clean fixture verification, full tests, race and vet
before validating the source.

Complete `misc/mvc/mvcDivisor.c` after division and common-cube normalization.
Preserve the initial any-repeated-literal rejection, independent source
duplication, lowest-index least-frequency repeated-literal choice, in-place
literal quotient order, common-cube removal after each step and nil results for
single/no-repeat covers. Implement the tail-recursive zero-kernel routine as a
loop so large covers cannot consume the Go call stack.

Pin multi-stage product kernels and unchanged-source digests across word
boundaries through 127 bits, plus both native nil paths. Unit-test closed owners
and result independence; fuzz arbitrary covers against an independent
frequency/filter/intersection model. Require exact source/fixture
hashes, clean fixture verification, full tests, race and vet before validation.

## MPM simple LUT-library checkpoint

Port both routines in `map/mpm/mpmLib.c` as one concrete owned value and a
checked constructor. Preserve the native twelve-variable maximum, unit area of
20, unit delay of 1, zero-size behavior, false variable-pin-delay flag and zero
padding outside each row's active input pins. Reject sizes outside zero through
twelve before allocating or indexing. Let Go ownership replace the native free
routine; do not add an explicit destructive method or an interface.

Pin all thirteen supported sizes with complete area and delay matrices from an
immutable reviewed fixture. Check the exact source and fixture hashes, and require clean immutable-fixture verification. Add
independent table-shape assertions, invalid-size tests, mutation-isolation tests
and fuzzing over all constructor inputs. Promote the source only after focused,
full, race and vet gates pass.

## IF specialized-match compatibility checkpoint

Treat `map/if/ifMatch2.c` as callable compatibility behavior, not an empty
translation unit and not a functioning technology matcher. Give both table
lifecycle routines concrete no-op Go functions. Preserve the cell lookup's
unconditional result of one and its complete lack of argument mutation. Since
the native `void *` manager is neither inspected nor dereferenced, represent it
as `uintptr`; represent truth storage as `[]uint32` and text as `string`. Do not
introduce an interface or prematurely define the unfinished IF manager.

Pin reviewed compatibility calls bracketed by setup/setdown across zero and nonzero manager
values, varied truth words, negative and positive counts and multiple strings.
Hash the exact source and permanent vector, require byte-exact
fixture integrity verification, and fuzz the constant result plus truth immutability. Promote the
whole source only after focused, full, race and vet gates pass.

## Lucky arbitrary-variable swap checkpoint

Account for all of `bool/lucky/luckySwapIJ.c` through the concrete owned
`TruthTable.SwapVariables` operation. Preserve identity, either argument order,
within-word exchange, mixed low-variable/high-word exchange and whole word-block
exchange. Validate dimensions before shifting or indexing, retain canonical
small-table padding, leave source storage unchanged and return owned results.
Use Go-managed scratch/exchange rather than reproducing the native case-3 heap
allocation, which is not freed by that routine.

Pin deterministic aggregate results for every identity and increasing variable
pair through twelve inputs, with separate digests proving execution of all three
native storage branches. Independently fuzz Boolean assignment permutation so a
matching packed implementation error cannot validate itself. Include reversed
pairs, invalid dimensions, ownership and large-width boundaries in the shared
truth-transform tests. Require exact source and fixture hashes, clean fixture
fixture integrity verification, and focused, full, race and vet gates before promotion.

## ANSI color diagnostic checkpoint

Port the configured non-Windows branch of `misc/util/utilColor.c` as a concrete
function returning owned text. Preserve every header, row, field width, ANSI
escape, space and newline, including the native first two rows' separate style
selector. Avoid implicit stdout and terminal-state mutation. Keep the Windows
console-attribute branch explicitly build-inactive rather than emulating it on
non-Windows systems.

Capture the complete raw native stream and pin its byte length and SHA-256 in a
permanent vector. Independently assert foreground-row, sample, reset, prefix,
suffix and deterministic-output structure. The function has no arguments, so
document why input fuzzing is inapplicable. Require exact source and fixture
hashes, clean dataset integrity verification, and focused, full, race and vet gates
before promotion.

## Process and temporary-file checkpoint

Port every declaration in `misc/util/utilSignal.c` and `utilSignal.h` with
concrete functions. Preserve the pinned non-Windows `/bin/sh` command behavior
and its `system()` wait-status encoding because existing ABC callers distinguish
zero, nonzero and start failure. Do not hide command execution behind an
interface.

Return an owned `*os.File`, path and checked error for temporary creation rather
than a raw descriptor and allocated `char **`. Preserve caller prefix/suffix,
unique exclusive creation and mode 0600, but deliberately replace native's
`mkstemp`, unlink and suffix-reopen sequence with one atomic `CreateTemp`
operation. Conditional removal must return filesystem failures instead of
discarding them. Pin reviewed compatibility exit/property results, test invalid directories,
leave/remove behavior and duplicate removal, and fuzz prefixes, suffixes and
payload round trips. Require source/header identity, clean dataset integrity verification,
full tests, race and vet.

## Gzip file-lifecycle checkpoint

Port `gzlib.c`, `gzread.c`, `gzwrite.c` and `gzclose.c` through concrete reader
and writer types rather than an interface. Preserve path-based read, truncate
and append construction; pre-initialization buffer choice; transparent raw
fallback; concatenated gzip members; buffer/byte/string/line I/O; guaranteed
bounded multiple-byte LIFO pushback; rewind; start/current seek; deferred
beyond-EOF read seek;
logical tell; effective compressed descriptor offset; descriptor-backed open
with ownership; EOF/error clearing; forward zero-fill writes; compression
levels; flush; and typed close. Add explicit uncompressed-byte limits and
cancellation. Compile-time reader/writer types replace native close-mode
dispatch.

Pin a native state-machine trace and complete compressed streams covering
zero-fill seek, append members, direct raw input, tell, pushback, line limits,
EOF clearing, rewind, flush and both close paths. Add a descriptor trace covering
nonzero starting offsets, ownership transfer, compressed offsets and lazy far
seek realization. Decode native streams in Go and fuzz file-backed round trips
across levels, split writes and gaps. The standard Go and pinned zlib encoders
need not produce identical compressed byte counts; require each reported offset
to equal its real descriptor position. Keep `gzlib.c`, `gzread.c` and
`gzread.c` may validate once all public read routines and the trailing-raw
transition pass; a separate bounded pushback stack is an acceptable safe
replacement for private native output-buffer occupancy. Keep `gzlib.c` and
`gzwrite.c` partial until exact ignored mode/strategy parsing, formatted writes,
strategies and midstream parameter/flush modes are implemented. `gzclose.c` may
validate once both typed close paths and misuse behavior pass. Require exact
source/fixture hashes, clean dataset integrity verification, full tests, race and vet.

## Zlib utility checkpoint

Port the portable public queries in `misc/zlib/zutil.c`: bundled version,
reference-build compile flags and defined error strings. Treat compile flags as
pinned compatibility metadata rather than a description of Go integer and
pointer types. Return an explicit validity Boolean for error codes outside -6
through 2 instead of reproducing the native out-of-bounds table access. Let Go
ownership replace the active allocation/free callbacks; do not create public
allocator shims. Record debug, 16-bit, Windows CE and fallback-memory branches
as inactive only after verifying the pinned compile configuration.

Pin the exact version, flags and all defined error messages in an immutable
dataset. Fuzz the full integer error-code domain for checked behavior. Require
exact source and fixture hashes, clean fixture integrity verification, and focused, full,
race and vet gates before promoting the complete configured source.

## Bzip2 constant-table checkpoint

Represent the complete `misc/bzlib/crctable.c` table as a fixed `[256]uint32`
generated from the native MSB-first polynomial recurrence. Keep package storage
unexported and return the array by value so callers cannot mutate shared state.
Do not paste a second mutable table or introduce an interface.

Pin a canonical big-endian digest and endpoints from the unchanged native
table, then independently check every Go entry against the eight-step
`0x04c11db7` recurrence and verify ownership isolation. This data-only leaf has
no fuzz input domain. Require source/fixture hashes, clean
fixture integrity verification, full tests, race and vet before promotion.

Port `misc/bzlib/randtable.c` as a fixed `[512]int32` with an owned-copy
accessor. Compare every entry with native ABC, independently check the table's
50-through-999 range and verify mutation isolation. Do not expose mutable
package data. Like the CRC table, this constant has no fuzz input domain and
must pass exact source/provenance and all standard gates.

## Bzip2 Huffman checkpoint

Port all of `misc/bzlib/huffman.c` as concrete slice/array operations: exact
length-limited heap construction, canonical code assignment and the fixed
23-entry limit/base decoder with owned permutation storage. Preserve native
strict comparisons and tie order, zero-frequency normalization, low-byte depth
tracking, maximum-length rescaling and symbol-order traversal. Reject negative
or overflowing frequencies, impossible length bounds and oversized alphabets
instead of reproducing assertions or nontermination.

Pin reviewed compatibility aggregate results across two through 258 symbols, including a
strongly skewed case that enters the rescaling loop. Independently verify every
code decodes through the generated tables and that no code prefixes another.
Fuzz arbitrary bounded frequency vectors and length limits; require no partial
result on error, clean dataset integrity verification, all hashes, full tests, race and
vet.

Begin central `misc/bzlib/bzlib.c` coverage with `BZ2_indexIntoF`, the
small-mode decompressor's cumulative-frequency lookup. Require exactly 257
nondecreasing boundaries and return a checked error for malformed storage
instead of reading unchecked memory. Preserve native bucket-zero behavior below
the first boundary, highest-matching behavior with duplicates and bucket 255
above the terminal total. Pin reviewed compatibility boundary probes and fuzz against an
independent linear scan. Keep the source partial until its complete streaming,
buffer/file, error/version and private state-machine surface is implemented.

Next expose the pinned `BZ2_bzlibVersion` metadata and replace
`BZ2_bzBuffToBuffDecompress` with Go's concrete `compress/bzip2` reader. Return
an owned slice instead of accepting caller storage, impose a checked output cap,
poll a concrete cancellation callback between decoder reads and expose no partial
bytes on format, truncation, limit or cancellation failure. A zero cap selects a
documented safe default; negative caps are invalid. Pin reviewed compatibility streams for empty,
scalar, text, repeated, full-byte-range and large deterministic binary inputs at
both low and high compressor levels. Require exact native decoding, exact and
one-byte-short boundaries where representable, malformed/truncated cases,
ownership isolation and adversarial bounded fuzzing. This does not cover buffer
compression or either streaming state machine, so `bzlib.c` remains partial.

## Zlib CRC-32 checkpoint

Port the complete public surface of `misc/zlib/crc32.c`: arbitrary-initial
updates, the IEEE lookup table and ordinary/64-bit-length concatenation. Use
Go's concrete standard IEEE updater in place of the native endian- and
alignment-specific loops, return an owned `[256]uint32` table rather than a raw
static pointer, and retain the GF(2) matrix combination semantics. Use a
`uint64` second length so negative native lengths are excluded by construction.

Pin the full native table by digest, updates for several initial states around
the one/eight/32-byte paths, and combinations with ordinary and greater-than-
terabyte lengths. Independently compare with Go's IEEE checksum, test table
ownership and fuzz incremental versus whole-buffer and combined versus
concatenated results. Require exact source and fixture hashes, clean fixture
fixture integrity verification, and focused, full, race and vet gates before promotion.

## Zlib memory-buffer checkpoint

Replace `misc/zlib/compress_.c` and `uncompr.c` with concrete standard-library
zlib operations in the main package. Preserve compression levels -1 and zero
through nine, and reproduce the exact `compressBound` arithmetic with overflow
checks. Return owned byte slices instead of caller-sized destination buffers.
Require explicit compressed/decompressed byte caps, cancellation polling and no
partial result on level, format, checksum, truncation, allocation or limit
failure. Keep gzip file lifecycle and the remainder of bundled zlib separate.

Pin reviewed compatibility bounds and streams across empty, scalar, text, repeated, binary and
full-byte-range inputs and representative levels. Require Go to decode every
native stream and require each Go stream to round-trip within the native bound;
do not require byte-identical compression because different valid deflate
encodings are compatible. Add exact-cap and one-byte-short boundaries, invalid
levels, malformed/truncated streams, source/output ownership, expansion caps,
round-trip fuzzing and adversarial decoder fuzzing. Promote both source files
only after exact source and fixture hashes, clean fixture integrity verification, and
focused, full, race and vet gates pass.

## Zlib Adler-32 checkpoint

Port all of `misc/zlib/adler32.c` as concrete uint32 streaming and combination
functions. Preserve arbitrary valid prior states, the native nil-buffer query
result, non-nil zero-length normalization, one-byte behavior, chunked modular
reduction at 5,552 bytes and both ordinary and 64-bit-length concatenation.
Accept a `uint64` second length so the Go API excludes the native routine's
undefined semantic domain of negative byte lengths.

Pin reviewed compatibility updates for several initial states and lengths immediately around
each algorithm branch, plus combinations with ordinary and greater-than-
terabyte lengths. Independently compare initial-one checksums with Go's Adler-32
implementation and fuzz incremental versus whole-buffer and combined versus
concatenated results. Require exact source and fixture hashes, clean fixture
fixture integrity verification, and focused, full, race and vet gates before promotion.

## DAU Quick population-count checkpoint

Port all active behavior in `opt/dau/dauCount.c` without preserving a native
runtime oracle. Generate the exact 256-by-32 packed byte-lane lookup in Go and
expose an owned snapshot plus stable language-neutral text when callers need to
inspect or regenerate the dataset; do not reproduce a C initializer or stdout
side effect. Preserve the recursive Quick selector/counting structure while
explicitly correcting the source's eight-input packed-total overflow and its
high-variable sibling-cofactor overwrite. Use unsigned counts, numeric byte
extraction and owned result vectors to eliminate signed overflow, host aliasing
and endian dependencies.

Replace the source's `srand`/`rand` assertion drivers with deterministic Go-only
unit tests and fuzzing. Check all 8,192 generated lanes, snapshot independence,
stable formatting, nil and zero-input conventions, every small boundary, the
known overflow/overwrite regressions and pinned permanent datasets under
`test_vectors/`. Compare Quick results with an independently implemented
assignment/bit model, including packed-repeat scaling and recursive widths;
never compile or invoke C during tests. Promote the source only after focused,
full, race and vet gates pass and fixture hashes are verified.

## DAU divisor-statistics checkpoint

Port all active behavior in `opt/dau/dauDivs.c` as
`DauDSDDivisors`. For each ordered variable pair, move the pair into the low
truth-table positions and scan every four-bit cofactor in packed order. Count
digits 1/14, 2/13, 4/11, 8/7 and 6/9 as the five complemented divisor classes.
Return owned rows and provide deterministic formatting instead of writing
stdout. Represent the demonstration routine's effective early-return behavior
explicitly without depending on an unported parser or inventing side effects.

Preflight pair count and aggregate digit work, poll concrete cancellation and
return no partial result. Unit-test every class, its complement, zero/one-input
domains, formatting, ownership and exact bounds. Fuzz against an independent
assignment-level model that reconstructs each four-bit cofactor without using
the production variable-swap algorithm. Promote `dauDivs.c` only after focused,
full, race and vet gates pass; the other DAU sources and headers remain separate
unfinished scope.

## DAU DSD formula checkpoint

Port the formula-manipulation front of `opt/dau/dauDsd.c` as an immutable,
indexed concrete syntax arena rather than an interface-valued tree. Parse the
source syntax for complement, variables, AND, XOR, three-input MUX, hexadecimal
prime truth functions and the six-variable evaluator's special
`<local-mux>{outer-fanins}` form. Bound bytes, nodes, nesting, variables and
truth storage; poll concrete cancellation and replace the source assertions
with atomic errors. Preserve delimiter matching, lexical variable discovery,
variable-sequence normalization, source gate-cost weights, packed truth-table
bit order, prime composition and local binding semantics. Deterministically
break normalization ties that valid disjoint-support formulas normally exclude.

Unit-test every operator, constants and complements, nested costs, hexadecimal
endianness, mux-local substitution, normalization idempotence, malformed and
mismatched syntax, each resource limit and cancellation. Fuzz arbitrary parser
input for panic freedom and stable render/reparse behavior. Separately generate
valid formulas and require normalized and original formulas to have identical
truth tables across every assignment. Port `Dau_DsdPerform` with its exact
support minimization, output factoring, pair substitution and best-variable MUX
fallback order. Exhaustively decompose all four-variable functions and reparse
and evaluate every result; fuzz unrestricted six-variable truth words with the
same independent round trip. Provide direct checked variable permutation and a
seeded random-permutation convenience operation. Preserve the source's mapping
and per-position swap schedule, but specify a concrete local generator rather
than retaining process-global, libc-dependent `rand()` state. Fuzz permutation
semantics against independent assignment remapping. Keep permanent compatibility datasets in
`test_vectors/` if later decomposition work needs them; never add a C generator
or runtime native oracle. Classify `dauDsd.c` as partial until the general
decomposition manager, level analysis, brace removal/merge and the remaining
machinery are implemented and independently tested.

Execution checkpoint: general non-split decomposition now uses the verified
truth-table output, phased-pair and mux-input kernels while retaining the
source's descending scan order, minimum-base swaps and owned prime truth
spelling through twelve variables. Its result reports the largest irreducible
block and reparses through the independent formula evaluator. Exhaustive zero-
through three-variable coverage, wide/nested unit cases and more than two
million semantic fuzz executions pass without C. One-step analysis and
split-prime output now preserve natural priority, minimum combined cofactor
support, the source local-MUX binding spelling, all three return statuses and
the no-recursive-split rule for outer MUX cofactors; another 1.5 million
semantic fuzz executions pass. Level priorities, explicit associative brace
removal and source diagnostics remain, so `dauDsd.c` stays partial.

The level-aware path now carries source levels through synthesized pair/MUX
definitions, including the extra recursive-definition level, and reproduces
the source selection sort's non-stable ties. Generated formulas run through the
owned associative brace pass. The source's duplicated specialized-six and wide
managers are deliberately consolidated into the general owned engine: direct
support checks replace its private pair cache and owned expressions replace its
fixed definition/write arrays without changing observable decomposition order.
Formula results replace the print wrappers. No-op, fixed-vector and external-
file benchmark drivers are covered by Go regressions and fuzzing rather than
retained runtime APIs or ambient inputs; the active Test44 and Test888 formulas
have explicit semantic regressions. These dispositions account for the
remaining machinery and promote `dauDsd.c` to validated.

## DAU formula-merge checkpoint

Port `opt/dau/dauMerge.c` in two reviewable layers. First implement associative
brace removal as an owned checked rewrite: flatten only uncomplemented ANDs
inside ANDs and XORs inside XORs, transfer a complemented nested XOR's phase to
one flattened fanin, cancel adjacent complement pairs, and retain prime and MUX
boundaries including mux-local binding lists. Bound input/output bytes, nodes
and depth through the common DSD options, poll cancellation, and return no
partial text. Unit-test exact forms, phase edge cases, ownership, malformed
syntax, limits and idempotence; fuzz valid formulas and compare complete truth
tables before and after without invoking C.

Then port the complete merge pipeline: permutation-aware variable replacement,
shared-variable presence and renumbering, pure-subtree status analysis,
definition extraction/inlining, conjunction decomposition, original-variable
restoration, associative simplification and deterministic normalization. Use
concrete owned slices and builders instead of fixed buffers, static return
storage, global counters or timing state. Define checked limits for variables,
definitions, intermediate bytes, truth words and decomposition work. Test
constants, disjoint inputs, every status class, complemented operands, primes,
mux-local blocks, permutations, failed one-step decomposition and atomic
errors. Fuzz merged semantics against independently evaluating both inputs,
applying their permutations/complements and ANDing the results. An exact
bounded prime is an acceptable temporary semantic fallback for reduced
functions above six variables, but keep `dauMerge.c` partial until the general
decomposer restores the source's wider structural spelling and the remaining
diagnostic/global artifacts have explicit dispositions.

Execution checkpoint: the temporary formatter and wide-prime fallback have
been removed; every reduced conjunction now uses the general decomposer. The
integrated merge fuzzer passed more than 2.2 million assignment-exhaustive
executions. Debug printers are reachable only under the source's local false
verbosity flag and are intentionally not process-global Go APIs; timing globals
are omitted, and the early-return demo has no observable result. All operational
symbols are covered, so `dauMerge.c` is validated.

## DAU non-DSD set checkpoint

Port `opt/dau/dauNonDsd.c` around a concrete two-bit variable-set value whose
roles are free, unique or shared. Provide checked role access, counts, parsing
and deterministic formatting; correct the source reader's uppercase offset so
uppercase letters map from `A` without an undefined negative shift. Preserve
bound-size and numeric-mask order, persistent variable-position swaps,
shared-subset order, the cap of three shared variables during discovery, and
the source containment relation. Implement decomposability through the exact
criterion: for each shared assignment, varying unique variables may expose at
most two distinct free-variable cofactors. Use owned dynamic storage, explicit
work/result limits and cancellation instead of fixed arrays and assertions.

For a selected set, return owned composition and decomposition truth tables,
the free/unique/shared original-variable maps and an exhaustive reconstruction
operation. Derive deterministic DSD formula pairs by mapping compact inputs
back to original variables and reserving the next variable as the selector;
return data rather than printing. Unit-test role encoding, source order,
containment, disjoint and genuinely shared examples, nondecomposable sets,
ownership and every bound. Fuzz discovery against an independent cofactor-key
model and require every omitted eligible set to be covered by an accepted set;
fuzz truth and formula reconstruction over every assignment. Return
`Dau_DecPerform6`'s complete shared-minterm polarity family as owned pairs,
fixing phase bit zero to remove global-selector duplicates and explicitly
bounding the source's overflow-prone larger shifts. Keep the source partial
until combined diagnostic disposition and wider general-decomposer structural
spelling are accounted for without a C runtime oracle.

Execution checkpoint: both component truth tables now use the general bounded
DSD decomposer through twelve inputs. A seven-input parity component proves
wide structural decomposition, and an independent Go-only composition fuzzer
checks arbitrary seven-input components embedded in valid eight-input source
functions. Deterministic formatters plus owned set, truth-table, formula and
polarity results replace every report path; exhaustive recomposition replaces
`Dau_DecVerify`. The three hard-coded source demos are ordinary regression/fuzz
inputs rather than runtime APIs. With those explicit dispositions, every active
source function is accounted for and `dauNonDsd.c` is validated without C or
stdout transcript oracles.

## DAU NPN enumeration checkpoint

Port `opt/dau/dauNpn.c` in layers. Begin with its reusable pruning kernel:
count directly symmetric input pairs, mark each second variable that has an
earlier symmetric partner, and classify each input flip as equivalent either
to output complementation or another input flip. Return owned Boolean vectors
plus checked compact masks rather than assuming a machine-width integer. Bound
comparisons and poll cancellation. Unit-test constants, symmetric operators,
unused inputs, asymmetric phases, dimensions and ownership; fuzz every result
against direct assignment remapping.

Replace `Dau_TruthEnum`'s fixed-path binary side effect with an owned exact NPN
class table for the tractable exhaustive range through four variables. Visit
normalized functions in ascending order and close each orbit under output
negation, all input phases and every permutation. Return ascending class
representatives and canonical lookup; bound classes/transforms and reject
larger explosive domains explicitly. Pin the established 1, 2, 4, 14 and 222
class counts and fuzz canonical invariance under generated transformations.
Then port network/function-by-node enumeration using the common concrete truth
and canonicalization types, retaining source gate-candidate order, support and
symmetry pruning, node frontiers and reduction modes. Replace table reads,
writes, dumps and progress printing with explicit data codecs/formatters under
caller control. Keep `dauNpn.c` partial until persistence, exact-minimal class
integration, both enumerators and all report/demo behavior have explicit tested
Go dispositions without native runtime or generated C oracles.

Execution checkpoint: the exact class table, pruning analysis, legacy network
frontier through all 222 four-input NPN orbits, and exact three/four-input
function frontier are implemented with deterministic `insdict.Dict` interning.
The two enumerators retain their distinct source candidate schedules, including
the legacy optional two-node branch and the newer reduced-final-level rule.
Their permanent tests use exhaustive Go orbit closure, assignment models,
frontier invariants and fuzzing. A versioned bounded little-endian codec now
replaces the native table save/read stream; owned exact-class summaries and
deterministic hexadecimal rows replace report and dump side effects without
reviving `tableW*.data` as an input dependency. The function frontier and
exact-summary path now extend deliberately through six inputs: support sizes
through four use the exhaustive lookup, while five and six use bounded exact
`MinimizeNPN` with shared work limits and cancellation. Wide-orbit model fuzzing
checks invariance under independently generated phases, permutations and output
complements. This completes the checkpoint and validates `dauNpn.c` without
silently presenting a small-domain implementation as full source coverage.

## DAU functional-library enumeration checkpoint

Port `opt/dau/dauNpn2.c` as two concrete layers. First parse its deliberately
small fully parenthesized Boolean grammar (`a` through `f`, prefix `~`, binary
`&` and `^`) into an immutable bounded AST. Reject trailing operands and
unbalanced delimiters that the assertion-based parser could ignore. Provide
exact truth evaluation, deterministic normalization, atomic GIA lowering and
caller-owned byte-slice library verification/construction; do not retain file
paths or C line readers as runtime dependencies. Test the embedded source
formula with an independent assignment model, malformed/limit/ownership cases,
round trips and all-assignment parser-to-GIA fuzzing.

Then port the DTT enumeration manager with concrete owned slices: exact NPN
orbit collection, deterministic truth interning, node/delay frontiers,
minimal-node and multiplicity accounting, transform composition/recovery, and
recursive formula-library generation. Replace stdout distributions and output
files with owned records and formatters. Bound classes, implementations,
frontier candidates, formulas and recursion; poll cancellation throughout.
Cross-check small domains against exhaustive truth/orbit models and replay every
generated formula through the first-layer evaluator. Keep `dauNpn2.c` partial
until this second layer and every diagnostic/demo disposition are tested.

Execution checkpoint: the formula/GIA/library front end and DTT core are now
implemented. The core retains the cyclic Gray-phase and adjacent-permutation
walk, five gate types, ordinary node-sum and delay-max frontier schedules,
minimal class costs, structural node counts, and the intentionally delayed
per-pair visited commit used by multiplicity mode. Deterministic `insdict.Dict`
state replaces hash-bin iteration. Exact three-input closure, independent orbit
sets, replayed fanins, ownership, work limits, cancellation and fuzzing are
tested. Packed NPN transform composition/recovery and bounded recursive
formula-library expansion complete the checkpoint below.

The packed transform and concrete formula steps are now implemented. Each
enumerated implementation retains the exact left-fanin permutation, input
phases and output normalization used by the source-order orbit walk. Transform
composition/recovery and all ten complemented gate spellings have assignment
models. Bounded canonical formulas and complete retained-implementation
libraries replay through the first-layer evaluator and fuzz successfully. The
final dumper behavior is now implemented without retaining its file or C
runtime: implementations are grouped by canonical truth in source order, the
first implementation for each unordered fanin-class pair is retained, and
recursive generation cross-products all alternatives of both fanins.
Canonical variable spelling is generated rather than assumed, and input
transforms are applied by checked formula substitution. Tests prove
deterministic output, a fixed root with multiple distinct recursive
substitutions, complete semantic replay, count/byte/depth/transform limits and
cancellation. Focused, full, race and vet gates pass, so this checkpoint
promotes `dauNpn2.c` to validated.

## DAU array checkpoint

Port all active behavior in build-disabled `opt/dau/dauArray.c` as a concrete,
owned pre-order node array in the common package. Preserve variable, fanin,
depth, offset, data, type, complement and unused concepts without reproducing C
bitfields or aliasing caller storage. Construct arrays from the checked DSD
parser; reject mux-local bindings because the source record has no binding
field. Validate the complete fanin walk, recompute depths on owned copies,
preserve the source per-node AND/XOR/MUX/prime cost model, and flatten nested
XORs plus non-complemented nested ANDs exactly as the source compactor does.
Return the source diagnostic form as a string instead of using stdout.

Test constants, complements, every node kind, nested fanin/depth metadata,
costs, XOR/AND flattening differences, receiver immutability, malformed arrays
and unsupported bindings. Fuzz generated valid formulas through construction,
depth reconstruction, formatting and compaction invariants. Promote the source
only after focused, full, race and vet gates pass.

## DAU alternate DSD checkpoint

Treat build-disabled `opt/dau/dauDsd2.c` as unfinished source rather than
silently normalizing its invalid preprocessor line, undeclared locals, missing
semicolons and placeholder object-recording comments. Port coherent operations
in reviewable slices. Begin with contiguous-bound-set recognition: for
`[first,last)`, prove that every bound minterm selects one of at most two
low-variable cofactors with the same selector mask in every outer partition,
and return an owned reduced truth table. For `[0,bound)`, recognize and return
the explicit `D(bound) XOR P(outer)` decomposition. Provide exact inverse
composition for both representations, checked dimensions/word limits and
atomic cancellation.

Exhaust normalized selector masks, reconstruct every accepted unit input,
reject third-cofactor and inconsistent-mask cases, and fuzz independently
generated reduced/phase functions through analyze-compose equality. Keep the
source partial until half-unshuffle, one/two-variable canonicalization, object
construction and all intended placeholder recording behavior have concrete,
independently tested Go definitions.

## DAU GIA lowering checkpoint

Port the formula-lowering front of `opt/dau/dauGia.c` onto the existing native
GIA manager. Parse once into the concrete DSD arena and append constants,
complements, variables, multi-input AND/XOR, MUX, mux-local binding and prime
nodes to an owned graph clone. Offer balanced or direct reduction and ordinary
or native XOR/MUX gates. Preserve source fanin order for equal-level balancing,
use structural hashing during construction, and restore the caller-visible hash
state. Compose prime truth functions through checked Shannon recursion up to
twelve inputs instead of retaining the C cover-workspace dependency. Bound the
complete object arena, poll cancellation throughout, and return no partial graph
or root on any error.

Test every operator across balance/native/prehashed mode combinations, constants,
phases, local bindings, primes, object exhaustion, invalid literals, malformed
formulas, cancellation, graph ownership and hash-state preservation. Fuzz valid
generated formulas and compare every assignment of the resulting GIA root with
the independent DSD truth evaluator. Keep `dauGia.c` partial until truth-to-DSD
dispatch, statistics and complete mapped-network derivation are implemented and
tested.

Execution checkpoint: `WithDauTruth` performs bounded general DSD dispatch and
lowers onto an owned graph while accepting unused trailing cut leaves.
`DeriveDauGia` validates the LUT mapping, computes every ordered boundary truth,
uses the source's always-balanced mapped path, preserves native gates, buffers,
CI/CO and sequential partitions plus metadata, omits unmapped logic and cleans
the result. Owned statistics and a deterministic formatter replace the three
process-global counters and reset-on-print side effect. Arbitrary truth and
random mapped-network fuzzers exhaustively compare all inputs without C. All
source definitions are accounted for, promoting `dauGia.c` to validated.

## DAU canonicalization completion checkpoint

The complete `opt/dau/dauCanon.c` translation unit is now accounted for. The
Go package provides direct packed counts plus independent Slow and Fast
cofactor-population strategies, checked small-word normalization, phase and
signature refinement, symmetry/tied-group reduction, simple/full/SCC and
cost-aware enumeration, hierarchical caches, exact grouped minimization and
the balanced-output wrapper. Concrete owned state, explicit work limits and
cancellation replace fixed arrays, static scratch and borrowed manager memory.
Forward source-indexed transformations reconstruct retained results.

The final symbol audit also records non-algorithmic dispositions: commented
duplicate cofactor comparators use the active `CompareCofactors` implementation;
debug-only canonical reconstruction, state copying and configuration checks are
covered by Go invariants, ownership and reconstruction tests; the adjacent-swap
diagnostic is an ordinary unit/fuzz property; and `Abc_TgResetGroup1` has no
caller in the C tree, so it has no runtime API. The called general reset path is
implemented. Slow-count fuzzing passed 1,244,318 Go-only executions and all
source definitions, private helpers, state types and fixed tables are ledgered,
promoting `dauCanon.c` to validated without generated C oracles.

With all seven translation units declared by `opt/dau/dau.h` validated, the
header itself is also complete. Its constants, type numbering and predicates
have direct checked Go forms; every operation declaration maps to the concrete
canonicalization, Quick-count, DSD, GIA, merge, non-DSD or tree API. Concrete
methods and option/result records replace `TtCanonicizeFunc` without adding an
interface. Go ownership replaces forward-declared allocation/free lifetimes,
and deterministic strings/summaries replace FILE/stdout entry points. The
duplicate hierarchical declaration maps to one method. The ledger therefore
promotes `dau.h` to validated rather than leaving aggregate header coverage
artificially partial after its defining sources are complete.

The excluded experimental `opt/dau/dauDsd2.c` is also fully dispositioned. It
is not selected by `module.make` and is syntactically and semantically
unfinished: invalid preprocessor directives, undeclared locals, missing
semicolons/return, a mismatched test variable, and comment placeholders instead
of object recording/removal mean its one-/two-variable manager sketches have no
executable C behavior. The coherent contiguous-range algorithms remain useful
and are ported as owned decompositions with exact inverse composition, bounds,
limits and cancellation. Independent reconstruction properties and Go-only
fuzz campaigns (1,074,337 general-range and 709,741 low-bound executions)
validate them. Half-unshuffle and private phase/swap/object helpers are recorded
as fragments of the nonterminating sketches, not invented public behavior.
This promotes `dauDsd2.c` to validated without making the Go package depend on
an imagined repaired C implementation.

## Counterexample utility completion checkpoint

The complete `misc/util/utilCex.c` implementation and `utilCex.h` public surface
are now represented by the concrete owned `Counterexample` type. Checked
constructors cover zero/full/trivial/value and combinational-model traces;
cloning versus input-only duplication is explicit. Packed access, population,
permutation, phase/temporal/unknown-initial transforms and checked merge retain
their defined source semantics. A register-aware combinational constructor
corrects C's latent missing input offset for nonzero registers; the sole C call
uses zero. Merge rejects the nonzero overloaded `part.iPo` mode because C's
allocation and copy lengths disagree and fail its final assertion.

Owned statistics expose total/register/real/auxiliary populations, and
deterministic strings replace all stdout diagnostics with defined empty-input
percentages. Go lifetime replaces `Free`, `FreeP` and the pointer-one sentinel.
Unit tests cover exact dimensions, packed boundaries, formatting and the
register-offset correction; independent Go-only layout/statistics/format fuzzing
passed 1,192,461 executions. Both files are promoted to validated without
interfaces or runtime native oracles.

## AIG manager checking completion checkpoint

`AigManager.Check` now accounts for all of `aig/aig/aigCheck.c`. It enforces
the source CI/CO fanin rules, internal canonical fanin ordering, hash identity
and entry totals, and live object counts, plus stronger Go invariants for
topological order, terminal slices, levels, exact phases, references,
register/constraint partitions, tombstones and choice chains. Checked errors
replace stdout diagnostics and assertions. `Aig_ManCheckPhase` is integrated
into every check. `Aig_ManCheckMarkA` has no separate API because persistent
marks are absent from `AigObject`; traversals own local mark storage, making the
reset assertion structural. Corruption tests cover each category and the
Go-only operation fuzzer invokes the checker after every mutation, clone and
cleanup; a fresh campaign passed 219,298 executions. The source is therefore
promoted to validated.

## DAU enumeration checkpoint

Port `opt/dau/dauEnum.c` as deterministic owned formula sets in the common
package. Preserve variable-partition loops, insertion order, star-marker
propagation, complement eligibility, symmetry exclusions and variable lifting.
Replace both static buffers and unconstrained result growth with checked formula
and aggregate-result limits plus concrete cancellation. Reject shifts beyond
the variable alphabet. Return the dump representation as a deterministic
string rather than writing the source demo's fixed path or process-global
stdout; persistence remains a caller decision.

Pin the complete zero- through three-variable sets and order. Parse and evaluate
every four-variable result, test byte/result exact limits, cancellation,
alphabet overflow, formatting and result ownership, and fuzz selection across
the enumerated corpus through independent parser/truth validation. Promote the
source only after focused, full, race and vet gates pass.

## DAU core template checkpoint

Treat `opt/dau/dauCore.c` according to its active behavior rather than its
suggestive placeholder names. Its temporary manager only retains graph and
parameter pointers, no optimization body exists, and the entry point always
returns one after freeing the shell. Represent this as one concrete Go
operation returning success; rely on ordinary Go lifetime instead of exporting
allocation/free functions, accept the same nil and parameter domains, and do
not invent validation or transformations. Unit-test unconditional success and
prove that a supplied graph is unchanged across boundary parameter values.
Promote this source only after focused, full, race and vet gates pass.

## DAU public-header and build-manifest checkpoint

Port the behavioral parts of `opt/dau/dau.h` as concrete package declarations:
the three public bounds, exact DSD node-kind numbers, constant predicates and a
checked variable-token decoder. Keep the header partial while any declared
implementation surface remains incomplete; forward declarations and the C
callback typedef do not justify Go interfaces. Audit `opt/dau/module.make`
separately as build metadata, proving that each selected translation unit has a
ledger disposition and that nonmembers remain in the full inventory.

Execution checkpoint: the bounds, kind numbering and token helpers have Go-only
unit coverage. The reserved zero kind corrects the earlier Go enum offset while
leaving symbolic consumers unchanged. The header remains partial because its
general DSD, GIA and tree-manager declarations are not complete. All twelve
default module sources are ledgered, so `module.make` is validated without a
runtime Go artifact. Focused, full, race and vet gates pass.

## DAU structural tree-manager checkpoint

Port `opt/dau/dauTree.c` in dependency order rather than treating its parser as
the whole source. Begin with an owned concrete literal/node manager, exact
complement encoding, support accounting, recursive structural comparison,
stable ordering, deterministic interning and the AND/XOR/MUX/prime operation
core. Replace flex allocation and hash bins with bounded slices and
`insdict.Dict`. Evaluate complete truth tables through explicit leaf mappings
so every normalization is checked independently. Then add the AND permutation
path, temporary network parse/transform/rebuild, deterministic formatting,
nondecomposable traversal, owned cache entries, shared maps, Boolean
decomposition and final merge. Explicitly disposition the disabled legacy
block; do not count it as active behavior silently.

Execution checkpoint: the owned manager and no-permutation operation core are
implemented. Source comparator behavior includes the subtle equality of
complemented variable leaves, retaining stable input order. AND/XOR flattening,
XOR phase folding, MUX normalization, prime ownership, structural interning and
the separate AND support-permutation path are covered. Deterministic formulas
round-trip through the Go parser/evaluator; prime detection, truth replay,
limits and cancellation have Go-only unit and fuzz coverage. Atomic formula
import now replaces the temporary create/transform/rebuild network and returns
its depth-first support mapping. It accepts zero/one-input primes, rejects
unsupported mux-local bindings and repeated support without mutation, and
cross-checks the independent parser truth. Owned shared-support maps preserve
phase composition and second-leaf order, while transparent classification
covers every active node kind. Merge constants and the disjoint-support path
now preserve source ordering, support-limit accounting and local-to-global
phase translation, with exact intersection validation and atomic failure.
The optional cache owns structural-local results, remaps them on hits, enforces
an entry cap and exposes separate shared/unshared counters. Shared-support merge
now derives phase-correct local coordinates, evaluates both functions, invokes
the general decomposer, enforces the nondecomposable limit and atomically imports
the structural result. Shared cache reuse and owned result mappings are covered;
single-leaf and structured overlap fuzzers jointly passed more than three
million executions. Owned manager summaries, bounded deterministic formatting
and independently allocated/stretched prime truth datasets replace active
report files and stdout. Prime construction now uses the validated
`Abc_TtCanonicize` Go pipeline and transfers canonical input permutation/phases
and output phase into the structural literal. Import preserves leaf mappings
through every normalization; a minimized `[(ac)b]` regression records the
stale-order bug found by fuzzing. `MergeAnd` plus the optional owned cache is the
concrete MPM wrapper replacement, while summaries replace hash/cache profiles.
Active demos are regression cases; the complete 6-feasible experiment and its
timing driver are inside `#if 0` and deliberately have no runtime API. All
active definitions are ledgered, promoting `dauTree.c` to validated. Full,
race and vet gates pass, and fuzzing leaves no `testdata/` artifact.

## Lucky simple exact-minimization checkpoint

Audit `bool/lucky/luckySimple.c` as one complete translation unit rather than
leaving it partial for behavior owned by other Lucky files. Represent its
Johnson-Trotter permutation walk and reflected-Gray phase walk with owned,
bounded schedules. Implement unrestricted NPN minimization and fixed-adjacent-
group minimization with the source mixed-radix transition order, including
output complementation and the singleton-only meaning of the input-flip flag.
Use Go lifetime for allocation/free helpers and shared packed-truth operations
for the three minimum-copy helpers. Correct the one-input dummy swap, which is
outside the valid variable domain and can corrupt packed padding.

Execution checkpoint: all fourteen active definitions have concrete Go paths
or ownership/helper dispositions. Forty unrestricted and 320 grouped permanent
cases cover compatibility, while independent traversal properties prove that
every permutation and phase is visited exactly once and a recursive Boolean
model checks exact unrestricted and grouped minima. Limits, cancellation,
ownership, word-boundary swaps and no-op carries are covered. Exact hierarchical
canonicalization consumes the grouped API. A fresh Go-only fuzz campaign passed
1,035,339 executions, so `luckySimple.c` is validated; other Lucky translation
units remain independently tracked.

## GIA cleanup-output checkpoint

Complete the ordinary-metadata path of `aig/gia/giaScl.c` by implementing
`Gia_ManCleanupOutputs` separately from selected-output cone extraction. Mark
through every original CO and buffer, then omit the requested CO prefix while
copying all marked internal objects; this intentionally retains gates used only
by skipped outputs. Return an owned source-object mapping, preserve native node
kinds/name/spec and a valid constraint suffix, and replace assertions for
sequential graphs, invalid counts and constraint splitting with checked errors.

Execution checkpoint: focused tests distinguish skipped-only cones, dangling
buffer cones and genuinely unreachable nodes, exhaustively compare every
remaining output assignment, verify source immutability/mappings/metadata and
cover zero-prefix and invalid requests. A fresh Go-only fuzz campaign passed
1,722,360 executions. `giaScl.c` remains partial because its combinational mark
recursion also follows optional `pNexts` equivalence and `pSibls` choice links;
the Go GIA manager must own and validate those metadata relations before this
translation unit can be promoted.

## GIA MFFC/dominator completion checkpoint

Finish `aig/gia/giaMffc.c` as a complete source. Use owned reference counters
for dereference/reference/restore MFFC discovery, owned dominator arrays for the
reverse closest-dominator forest, and explicit cut boundaries for the recursive
zero-through-five-variable cofactor derivation. Include buffer and native mux
control edges and preserve source boundary encounter order without mutating
manager scratch.

Replace `Gia_ManComputeDomsTry` rather than retaining its process-global
diagnostics: return bounded owned candidate records containing the source-loop
object and effective root, MFFC size/boundary/reference deltas in its strict
selection-sort order, and the effective `k=0` derived cone. Do not write the
fixed-name AIGER dump or call the unrelated external collapse experiment;
persistence and follow-on analysis belong to callers.

Execution checkpoint: all eleven definitions have concrete Go implementations
or are absorbed into owned storage. Native fixtures pin reconvergent MFFC,
dominator and cofactor cases; independent consumer-closure/intersection models
check MFFC and dominators; exhaustive recursive cut replay checks each candidate
cone. Invalid roots/cuts, dangling logic, extended nodes, five-variable limits,
ownership, caps and cancellation are covered. A fresh Go-only candidate fuzz
campaign passed 1,648,524 executions, promoting `giaMffc.c` to validated.

## CNF area-flow mapping completion checkpoint

Close `sat/cnf/cnfMap.c` around its actual compiled surface. Preserve active
topological cut selection with exact SOP cost scaling, integer per-reference
flow division, unsigned accumulation, eleven-bit leaf-reference sums, the
larger-reference-sum tie break and first-candidate exact ties. Represent the
compiled but uncalled supergate formula as a checked pure helper over regularized
literals; reject malformed flow arrays, zero-reference internal leaves and
overflow rather than inheriting assertion/wraparound behavior.

Disposition the contiguous exact-area section explicitly: `Aig_CutDeref`,
`Aig_CutRef`, `Cnf_CutArea`, `Cnf_CutCompare`, `Cnf_ObjFindBestCut`,
`Cnf_CutAssignArea` and `Cnf_ManMapForCnf` are all inside `#if 0`, and their
only apparent iterative call is inside the separately disabled old core
derivation. Do not claim that experimental pass as active behavior. Compose
automatic Go CNF construction from the validated DAR generation, active
area-flow selection, mapping scan, postprocess and writer APIs.

Execution checkpoint: permanent data covers all 65,536 truth costs/reference
sums; selection properties check candidate order, tie rules and ownership;
supergate unit tests cover complemented leaves, terminals, dimensions,
references and overflow. A fresh independent arithmetic-model fuzz campaign
passed 1,934,279 executions, promoting `cnfMap.c` to validated.

## Float priority-queue completion checkpoint

Complete `misc/vec/vecQue.h` as an owned bounded `FloatPriorityQueue`. Preserve
float32 max-heap comparisons, equal-priority movement/child ties, ID-priority
fallback, sparse membership, growth, clear/top/pop/update behavior and heap-to-
inverse-order consistency. Replace the borrowed pointer-to-pointer cost array
with copied priorities and explicit atomic setters/rebuilds; use Go lifetime
instead of both free helpers and return checked errors instead of assertions.

Return `Vec_QuePrint`'s level-by-level ID and rounded-priority layout as a
deterministic string. Treat `Vec_QueTest` as a demonstration superseded by
ordinary and model-fuzz tests, not a runtime queue API. Strengthen the native
dense-prefix checker to validate arbitrary sparse IDs, ownership and malformed
state.

Execution checkpoint: permanent rows pin ID-priority, equal-priority, updated-
priority and infinity/signed-zero heap orders. Unit tests cover exact formatting,
ownership, cloning, nil/empty/error behavior, NaN rejection, caps and deliberate
invariant corruptions. A Go-only reference-model fuzzer checks every operation
and passed 193,094 executions, promoting `vecQue.h` to validated.

## CNF cut lifecycle completion checkpoint

Finish `sat/cnf/cnfCut.c` around owned values rather than recreating its flex
arena. Treat `AigCNFCut` construction in composition/postprocessing as
`Cnf_CutAlloc/Create`, Go lifetime as `Cnf_CutFree`, and return the exact
`Cnf_CutPrint` leaf layout as a string. Implement dereference/reference/update
over owned mapping-reference arrays. Absorb merge, truth-position and temporary
remove/insert helpers into pure atomic substitution so caller cuts never mutate.

Retain the documented correction to the source truth-pointer, support movement
and packed small-table cofactor defects. Reviewed original results remain
regression evidence, not the desired semantics. Cover sorted/overlapping
supports, both polarity costs, saturation, sixteen inputs, ownership, malformed
cuts, limits, cancellation and exact formatting. Execution checkpoint: all
twelve definitions have implementations or explicit owned-lifetime/internal-
helper dispositions; 240 reviewed corrected cases and exhaustive assignment
checks pass, and fresh Go-only substitution fuzzing passed 1,539,092 executions,
promoting `cnfCut.c` to validated.

## Kit ISOP completion checkpoint

Complete `bool/kit/kitIsop.c` with one concrete owned interval engine rather
than separate public implementations for its two native recursive widths.
`TruthTable.ISOP` is the equal on/allowed case; `ISOPInterval` accepts a required
on-set and an allowed on/don't-care set, verifies containment, and returns a
cover between them. Preserve ordered zero/one/shared decomposition through
sixteen variables and choose a complementary-interval cover only for fewer
cubes or equal cubes with fewer literals, retaining direct ties.

Return the selected cover's exact owned truth table and deterministic PLA-style
text instead of native output parameters and stdout. Intentionally derive that
table from the selected cubes, avoiding `fReturnTt`'s ambiguous last attempted
polarity when both covers are explored. Use cumulative allocation/work limits,
poll cancellation, retain a completed direct result if only the optional
complement attempt reaches its limit, and otherwise expose no partial result.

Execution checkpoint: all six definitions are implemented or absorbed into
the shared engine and formatter. Permanent equal-set rows pin exact native cube
order, phase and counts, while Go-only unit properties exhaust every assignment
of representative intervals and cover constants, invalid containment,
ownership, limits and exact formatting. Fresh standalone interval fuzzing
passed 1,811,974 executions, promoting `kitIsop.c` to validated without any C
source, compiler or runtime oracle under `pkg`.

## Packed bit-vector completion checkpoint

Treat `misc/vec/vecBit.h` as one owned bounded `BitVector`, accounting for all
27 inline functions rather than preserving unsafe pointer aliases. Constructors,
rounded allocation, logical size, entry access, growth/fill, autoextending
get/set, stack operations, population count and reset retain packed LSB-first
semantics. `Words` returns an owned logical snapshot, `ReleaseWords` transfers
the packed storage, garbage collection absorbs `Vec_BitFree/FreeP`, and the five
iteration macros become explicit indexed or range loops.

Retain the reviewed Go corrections: `FillExtra` must avoid the source's boundary
word writes and full-width shifts, clone must support unaligned logical sizes,
released/snapshotted tail padding must be deterministic, and `StorageBytes`
reports the actual packed payload instead of multiplying bit capacity by an
integer width. Bounds/allocation failures are checked and atomic.

Execution checkpoint: every inline definition has a concrete implementation or
ownership/language disposition. Permanent rows pin native packed transitions;
boundary tests cross every adjacent 32-bit edge and verify limits, zero/nil
behavior, clone/snapshot/transfer ownership and corrected cases. A fresh
standalone boolean-slice model fuzz campaign passed 208,121 operation sequences,
promoting `vecBit.h` to validated without C source or a runtime native oracle.

## Float-vector completion checkpoint

Complete `misc/vec/vecFlt.h` as one concrete owned bounded `FloatVector`.
Preserve exact/minimum allocation, all-one-byte full initialization, copy/clone/
move/release/erase, float32 arithmetic and max-update, growth/fill/stack/search,
IEEE equality/extrema and ascending/descending ordering. Replace the adopting
array constructor with a copying constructor, borrowed `ArrayP/EntryP` aliases
with snapshots and checked writes, both free helpers with Go lifetime, the two
private comparators with `Sort`, and four iteration macros with ordinary loops.

Return `Vec_FltPrint`'s finite six-decimal layout as a deterministic string.
Permit Go-defined nonfinite spellings rather than inheriting platform C-library
text. Preserve NaN payloads in storage and copying, but reject NaN ordered
insertion/sorting atomically because the source comparator is not a total order;
use stable numeric sorting so signed-zero encounter order is deterministic.

Execution checkpoint: all 42 function definitions have implementations or
explicit ownership/internal-helper dispositions. Permanent transitions pin
native capacity and float-bit behavior; unit tests cover ownership, limits,
NaNs, signed zero, infinities, extrema and exact formatting. Fresh standalone
fuzzing passed 794,222 stateful operation sequences and 1,983,556 raw IEEE
cases, promoting `vecFlt.h` to validated without C source or a native oracle.

## Four-input CNF table completion checkpoint

Close `sat/cnf/cnfData.c` around its only active definition. Decode the embedded
`s_Data4` payload once into immutable native-ordered four-input SOP covers and
return owned packed copies to callers. Preserve the source's trailing-padding
tolerance while rejecting malformed symbols, separators, sizes and incomplete
tables; do not expose mutable global table memory.

Explicitly disposition `Aig_ManDeriveCnfTest` and `Aig_ManDeriveCnfTest2`: both
are wholly under `#if 0` and are table-development/debug programs that print or
consume an external candidate file. Permanent payload integrity plus exhaustive
decoder semantics supersede that offline workflow; they are not runtime APIs.

Execution checkpoint: permanent data matches every one of the 65,536 native
covers and counts, and independent exhaustive assignment evaluation proves each
cover realizes its truth ID. Malformed decoder inputs, ownership and concurrent
initialization are unit-tested. Fresh standalone cover fuzzing passed 1,994,033
cases, promoting `cnfData.c` to validated without generated C oracles.

## Fast CNF completion checkpoint

Complete all twelve active definitions in `sat/cnf/cnfFast.c` with owned local
state. Iterative leaf and volume traversal must preserve fanin-zero-first order,
signed versus regular first-encounter uniqueness and postorder. Derive truth
with native six-variable patterns, and preserve fast marking's mux groups,
CI/CO, sharing and complemented-edge boundaries plus support recovery.

Assign variables in reverse object order and emit flattened signed AND
supergates or ordered direct ISOP clauses, followed by output/equivalence and
constant clauses. Return diagnostic marked-object/node-clause counts rather
than printing, including the source's deliberate omission of output/constant
clauses and its contradictory-supergate undercount. Replace graph marks, iData
and global scratch with owned slices; return checked errors for marking-rule
violations, unsupported explicit object kinds and complex cuts above the
source's asserted six-leaf bound. Enforce object, visit, variable, clause,
literal and ISOP limits with cancellation and no partial result.

Execution checkpoint: native cone, marking, clause-stream and object-variable
fixtures match, while independent assignment checks and exhaustive SAT
projection prove complete formulas. Deep traversal, malformed boundaries,
source immutability, ownership, limits and cancellation are covered. Fresh
standalone fuzzing passed 1,212,559 cone cases, 1,073,059 marking cases and
504,660 complete derivations, promoting `cnfFast.c` to validated without C.

## CNF writer completion checkpoint

Complete all nine active definitions in `sat/cnf/cnfWrite.c`. Preserve ternary
SOP conversion and literal counting, packed cube emission, mapped-CNF numbering
and clause polarity order, mapping-array layout, object-ID clause ranges and
both four-input and wide supplied-cut paths. Automatic entry points compose the
validated DAR generation, selection and scan stages; supplied truth remains a
caller contract as in C.

Add concrete owned simple encoders for `Cnf_DeriveSimple` and
`Cnf_DeriveSimpleForRetiming`. Retain variable zero as unused, native CO/AND/CI/
constant allocation order, three-clause AND encoding and constant-one unit.
When any outputs are exposed, preserve the source's unusual allocation of SAT
variables for every CO even though earlier asserted outputs use driver units;
equate only the final exposed suffix. The retiming form equates every CO and
adds a negative unit for each. Return meaningful driver literals for asserted
outputs and CO literals for exposed outputs. Reject unsupported explicit buffer
or XOR objects and enforce object/variable/clause/literal limits plus
cancellation without partial output.

Execution checkpoint: exact hand-derived streams pin both zero/nonzero exposure
numbering and retiming clauses. Solver-backed tests exhaust input assignments,
including globally contradictory asserted outputs, exposed-output equivalence,
retiming semantics, ownership, malformed requests and every resource gate.
Fresh standalone fuzzing passed 1,404,802 cases. Together with existing native
mapped/object fixtures and exhaustive SAT projection, this promotes
`cnfWrite.c` to validated without a C runtime oracle.

## Utility sorting completion checkpoint

Complete all 24 active definitions in `misc/util/utilSort.c` as concrete slice
operations. Preserve integer merge/selection thresholds, signed and unsigned
cost permutations, alternating equal-cost merge behavior, low-32-bit word
comparisons with high-word payloads, ascending/descending selection bases and
both two-way and three-way partition helpers. Retain the observable source quirk
that `Abc_QuickSort3` calls the two-way recursion, while separately exposing the
otherwise-unreferenced three-way helper behavior. The qsort-backed variant may
leave equal-key order unspecified as C does.

Quick-cost sorting must compare the uint32 bit patterns of signed costs and
return the packed high-word indices. Replace comparator subtraction with direct
comparisons and make empty input safe, correcting overflow and out-of-bounds
source hazards. Replace the process-global multiply-with-carry RNG with an owned
`ABCRandom`; preserve native seeds, reset-before-advance, sequence values and
`Abc_RandomW`'s duplicated halves when reset is requested. Treat `SortTest` and
`QuickSortTest` as benchmarks superseded by unit/model tests, and explicitly
exclude the alternate implementation wholly under `#if 0`.

Execution checkpoint: exact units pin merge ties, signed/unsigned costs,
payload preservation, every wrapper, RNG sequence/reset/ownership and corrected
empty behavior. A standalone fuzzer compares integer results to an independent
sort, validates low-word order and complete multisets for every algorithm, and
checks both cost permutations; a fresh campaign passed 516,998 cases. This
promotes `utilSort.c` to validated without C source or a native runtime oracle.

## GIA fanout-index completion checkpoint

Complete all ten external functions and six inline helpers in
`aig/gia/giaFanout.c` with owned indexes rather than manager-attached pointer
storage. Ordinary static snapshots preserve one target per incoming edge in
object/slot order, including duplicate edges and mux controls. A concrete
mutable `GiaDynamicFanouts` preserves the source's two structural fanin slots,
tail insertion and exact removal; require callers to name and validate the slot
instead of inheriting `Gia_ObjWhatFanin` ambiguity. Accessors return owned edge
records containing target and slot.

For mapped fanouts, count and append LUT edges by ascending root and record
fanin order, retain duplicate fanins, then append CO-driver edges in CO order.
Use owned offset/target storage to absorb both fanout-map allocation helpers and
Go lifetime for stop/free paths. Replace `Gia_ManStaticFanoutTest` stdout with
ordinary exact/model tests. Bound objects/edges, poll cancellation and expose
no partial index.

Execution checkpoint: units verify ordinary slots, list-tail reinsertion,
duplicate mapped fanins, CO ordering, ownership, invalid mutations, limits and
cancellation. A standalone independent graph-scan fuzzer checks every initial
edge and slot, then removes/reinserts all edges; a fresh campaign passed
1,813,845 cases. This promotes `giaFanout.c` to validated without C or mutable
manager scratch.

## GIA enable-analysis and removal completion checkpoint

Complete all ten active definitions in `aig/gia/giaEnable.c`. Preserve the
active immediate-fanin supergate collector and separately expose the recursive
helper's complemented-edge/CI boundaries and left-to-right DFS uniqueness,
even though the source wrapper comments out that recursive call. Replace
frequency and detector stdout with owned result records plus deterministic
formatting; preserve strict thresholds, category order, object-ID order and
cross-category deduplication.

Keep frame transfer and initialized unroll/cofactor on their existing explicit
mapping paths. For `Gia_ManRemoveEnables2`, recognize native or structural
muxes, require a branch whose regular object is the corresponding RO, apply the
RI phase to the non-feedback branch, retain unmatched RIs and return normalized
update controls and recognition state. Port the older `Gia_ManRemoveEnables`
as a separate OR-of-two-AND recognizer: preserve exact feedback-literal tests,
first-match branch order and one-based equality classes for control literals.
Return classes explicitly instead of adding hidden `vFlopClasses` state to the
manager. Rebuild into owned graphs, preserve terminal order/name/spec/registers/
constraints and extended-node semantics, and enforce mapping/object limits and
cancellation without partial results.

Execution checkpoint: all ten definitions now have concrete Go behavior or an
owned diagnostic/lifecycle disposition. Unit tests exhaust all 32 combinations
of control, data, feedback and RI phase/branch placement over every input/state
assignment; they also cover legacy shared-control classes, unmatched-driver
retention, recursive order, exact formatting, ownership, malformed bounds and
cancellation. A standalone Boolean-model fuzz campaign passed 2,311,228
executions. The uncached full suite passes in 10.455s, race in 59.326s, and vet
is clean, promoting `giaEnable.c` to validated without a C runtime oracle.

## GIA partial-pattern verification completion checkpoint

Complete all three external functions and both inline mark accessors in
`aig/gia/giaPat.c` using owned slices. Collect the regular driver cone once in
fanin postorder, including CIs and gates but excluding the CO wrapper and
constant boundary. Preserve unique first traversal, CI-ordinal assignments,
unknown defaults, later duplicate assignment precedence, edge phases and
ternary AND semantics. Return the final ternary value and definite-one status
instead of writing a failure message; source marks must remain untouched.

Use iterative traversal with object/value bounds and cancellation. Reject a
constant root as the source asserts. Extend the intended plain-AND domain
safely to owned GIA buffers, XORs and muxes using their actual ternary semantics
and include mux controls in the cone, rather than silently evaluating an
extended object as an AND. Keep assignments and returned cones caller-owned.

Execution checkpoint: exhaustive unit tests cover every partial assignment and
value combination for AND, buffer, XOR and mux roots, plus exact postorder,
duplicate overwrite, ownership, malformed roots/indices, limits and traversal/
simulation cancellation. A standalone Boolean-completion fuzzer checks that a
definite ternary answer never contradicts any concrete completion; a fresh
campaign passed 1,861,405 executions. The uncached full suite passes in
10.070s, race in 58.598s, and vet is clean, promoting `giaPat.c` to validated
without a C runtime oracle.

## Sequential AIG cleanup-diagnostic completion checkpoint

Complete both functions in `aig/saig/saigScl.c` as non-printing owned queries.
For useless-register reporting, scan primary outputs for regular register-output
drivers with exactly one total graph reference, and scan register inputs for
regular primary-input drivers with exactly one total graph reference. Return
both counts and the register total without assigning CIO IDs or modifying the
manager.

For `Saig_ManReportComplements`, preserve behavior rather than the suggestive
name: scan RI drivers in order, ignore literal phase, and increment once for
every occurrence of a regular driver after its first occurrence. Thus three
RIs sharing one regular driver return two, not three pair combinations. Replace
`fMarkA` writes/cleanup with a local Boolean slice; reject the asserted
zero-register case and enforce object limits/cancellation atomically.

Execution checkpoint: units distinguish total reference counts from terminal
roles, pin phase-insensitive repeated-driver behavior, prove source ownership
and cover malformed limits/zero registers/cancellation. A standalone fuzzer
builds arbitrary PI/RO-driven PO/RI selections and compares both reports to an
independent terminal/reference model; a fresh campaign passed 1,853,877
executions. The uncached full suite passes in 10.189s, race in 58.627s, and vet
is clean, promoting `saigScl.c` to validated without a C runtime oracle.

## Paged word-memory completion checkpoint

Complete `misc/vec/vecMem.h` around the existing concrete `WordMemory`. Retain
fixed-width uint64 records, page/entry metrics, stable insertion IDs, uppercase
high-word-first dumps, truth-table width/seeds/mux patterns and deterministic
owned snapshots. Keep reviewed corrections: newly exposed gaps are zeroed,
empty and exact-boundary shrink release the right pages, hashed records reject
mutation, truncated keys disappear, and all failed operations are atomic.

Map language/lifecycle helpers explicitly: Go garbage collection absorbs
`Free/FreeP`, `Clone` supplies the source's empty `Dup` stub, ordinary loops
replace `Vec_MemForEachEntry`, `DumpDigit` folds into the formatter, and the
empty `Fill/Clean` source bodies require no state change. Lookup-only maps
replace bucket links/profile/resize without iteration-dependent IDs. Preserve
`Vec_MemHashKey` separately with uint32 overflow, little-endian low/high halves
and its eight prime multipliers. Replace the truth-table file/stdout wrapper
with an owned result containing exact text, filename convention and summary.

Execution checkpoint: every macro/function has concrete behavior or an
explicit language/internal-stub disposition. Unit tests pin native transitions,
stable IDs, truth patterns, dump layout/summary, exact hash arithmetic,
ownership, bounds and corrected hazards. Stateful page/intern and truth-memory
fuzzers remain in the suite; a fresh independent hash campaign passed 2,036,171
executions. The uncached full suite passes in 10.263s, race in 59.059s, and vet
is clean, promoting `vecMem.h` to validated without C source or runtime oracles.

## BSAT solver3 completion checkpoint

Port `sat/bsat/satSolver3.c` and `satSolver3.h` as a concrete `SatSolver3`
facade over the existing owned `SatSolver` CDCL engine. Do not duplicate the
watched-clause, first-UIP, restart, learned-reduction or final-conflict engine:
solver3 is another MiniSat-family storage/heuristic variant, while the Go core
already supplies those semantics with checked ownership. The facade must own
its pushed-assumption stack, configured relative and absolute work limits,
seed, lifecycle state and single bookmark. It must not introduce an interface.

The public lifecycle and diagnostic surface includes construction with the
default or checked explicit seed, monotonic variable-count growth, atomic
clause insertion, root simplification, ordinary solving, model/final-conflict
access, original/learned/conflict/assigned counts, owned memory accounting,
activity ordering, polarity controls, random-decision control, runtime limits,
bookmark/rollback, restart and stop. Restart discards the formula, learned
state, pushed assumptions, result and statistics while retaining the facade's
configured variable bound; restart-with-seed installs the supplied checked
seed. Go lifetime replaces delete/delete-pointer helpers. DIMACS, statistics,
model projection, trace and proof-store operations reuse the already validated
concrete helpers rather than receiving solver3-specific copies. The disabled
clause-doubling diagnostic remains an explicit unsupported operation.

Preserve solver3 push/pop behavior separately from `Solve`: each push adds one
decision-level assumption and performs propagation only, never an unrestricted
search. Recompute the pushed prefix on an owned trial core so failure is
reported with a copied final conflict and no partially committed facade state;
pop removes exactly one committed assumption. Ordinary solve composes the
committed prefix with call-local assumptions and always returns the core to its
root state. Relative conflict/propagation limits begin at the current cumulative
statistics, absolute limits cap them, and the tighter remaining budget wins.
Already-exhausted absolute limits return unknown without work. Keep limits
uint64 and reject negative/overflowing compatibility inputs at boundaries.

Implement lexical SAT over an ordered literal vector without recursive facade
mutation. Literal polarity encodes the requested starting bit (complemented is
zero, regular is one). First test the exact lower assignment. If it is absent,
try zero-to-one pivots from least significant toward most significant while
holding the preceding prefix equal; the first satisfiable pivot is the least
larger prefix. Greedily minimize every suffix bit by testing zero before one.
Return an owned literal vector and final SAT model, leave the caller's input and
pushed prefix unchanged, reject duplicate variables, and return UNSAT when no
assignment at or above the requested word exists. Resource exhaustion remains
unknown rather than being misread as absence of a prefix.

Port both assumption minimizers with explicit semantics. The ordinary variant
accepts a formula known UNSAT under the supplied assumptions and returns an
order-preserving irredundant subset using deterministic deletion queries;
unknown queries conservatively retain the literal. The specialized variant
also probes the complement of a removable singleton and may commit a proven
unit consequence, but all additions must be staged and atomic. Preserve the
source requirement that the marked result is nonempty when the base formula is
SAT. Correct the two dead source loops whose conditions use `k > nLitsL` and
`k > nResL` immediately after initializing smaller `k`; do not claim those
loops provide unit learning. Reject satisfiable input rather than silently
returning an arbitrary subset.

Port every inline gate builder as a pure ordered-clause constructor plus an
atomic loader: constant, buffer, enabled buffer, complemented-output AND, XOR,
mux, four-to-one mux, XOR-AND, constraint and half sorter. Preserve literal and
clause order, including the mux's four-clause equal-data shortcut and the exact
selector convention of the four-to-one mux. Builders validate all variables
before allocation and return owned clauses. Loader failure leaves the solver
unchanged. Exhaust every Boolean assignment for each polarity/alias pattern;
compare satisfying projections to independent Go truth functions rather than
to stored native output.

Testing is entirely stand-alone. Unit tests cover push propagation without
search, failed-push atomicity, nested pop, root conflict, composed assumptions,
relative/absolute limit precedence, restart/seed, activity and polarity order,
bookmark restoration, memory ownership, models and final conflicts. Lexical
tests exhaust every formula and starting word over small ordered domains and
compare to direct assignment enumeration. Minimizer tests prove returned
subsets are UNSAT, order-preserving and irredundant and check unknown-limit
conservatism. Stateful fuzzing compares clauses, pushes, pops, restarts and
solves to a small exhaustive formula model; dedicated gate, lexical and
minimization fuzzers use independent Boolean enumeration. Retain no C helper,
compiler, executable, generated native fixture or fuzz corpus. Promote both
files only after focused tests, the uncached full suite, race testing, vet, the
package-header check and the stand-alone-oracle guard all pass.

Execution checkpoint: the complete active source and header surface now has a
concrete path or explicit Go-lifetime/reuse disposition. Lexical search is
exhaustive over every three-variable truth subset and passed a fresh 5,000-case
arbitrary-CNF campaign. Sound unit-learning minimization passed 5,001 cases;
stateful push/pop/solve passed 5,000 cases against direct enumeration. Every
gate assignment, ordered clauses, lifecycle, limits, ownership, utility, trace
and proof-store delegation are covered. The complete suite passes in 10.696s,
race in 59.280s, and vet is clean, promoting `satSolver3.c` and `satSolver3.h`
to validated without native source or runtime oracles.

## Legacy BSAT solver completion checkpoint

`src/sat/bsat/satSolver.c` and `satSolver.h` share the validated owned CDCL
engine and solver3 facade instead of creating a second maintenance-heavy solver.
The concrete `SatSolverSession` alias retains one package and no interfaces.
It covers construction and monotonic growth, original/learned clause insertion,
watched propagation, first-UIP learning and minimization, models/final conflicts,
push/pop, relative and absolute work limits, lexical SAT, both assumption
minimizers, learned reduction, restart, bookmark/rollback, polarity, random
decisions, counts, memory, DIMACS/statistics/model helpers, tracing, proof-store
lifecycle and all inline gate encodings.

Legacy-only state has explicit owned APIs. Model projection copies its variable
list and appends values after each SAT result. Run-ID cancellation composes with
ordinary stop callbacks. Clause diagnostics return copied clauses rather than
printing. Attached proof stores atomically retain original clauses and one empty
UNSAT marker, including change-last behavior. Learned-policy configuration sets
the current start limit plus delta and percentage; a zero Luby interval disables
periodic restarts. Activity-factor arrays and active-variable lists are copied,
validated atomically and applied at odd-numbered search starts. The optional
global-variable mask applies the source's factor-three bump at every search
start to marked active variables. Factor arithmetic is checked/saturating,
updates the live heap, participates in owned bookmarks and memory accounting,
and grows with zero factors when variables are added.

No compatibility hook is retained for `pCnfFunc`: its only compiled block is
guarded by `TEST_CNF_LOAD`, while the enqueue version is commented out. The Go
solver always performs safe level-zero propagation and never performs the
source's destructive database simplification, so `fSkipSimplify` has no useful
state to expose. Go ownership replaces arena handles, pointer-tagged binary
clauses and delete-pointer helpers. Assertions and stdout diagnostics become
checked errors and copied diagnostic records.

Standalone tests cover caller-storage independence, atomic invalid-policy and
factor rejection, restart disabling, learned-policy state, model projection,
run cancellation, proof-store recording/change-last, diagnostics, lifecycle and
the exact odd/every-search bias cadence. An independent recurrence fuzzer covers
5,000 combinations of factor, active/global masks and repeated searches. These
join the solver3 exhaustive three-variable lexical catalog, exhaustive gate
truth assignments, sound minimizer checks, direct-enumeration solver models and
5,000-case lexical, push/pop and minimization fuzz campaigns. Promotion requires
the uncached full suite, race suite, vet, first-line package policy and native-
oracle guard to remain clean. The checkpoint passed those gates: complete suite
10.504s, race 59.954s, with vet, formatting and policy guards clean; both source
files are validated in the authoritative ledger.


## CNF utility completion checkpoint

`src/sat/cnf/cnfUtil.c` is complete across its nine active definitions. The two
mapping families remain concrete and separate: supplied `Cnf_Cut_t` scans own
preorder/postorder roots, ordinary four-input cuts, dynamic references and SOP
area, while DAR-best-cut scans retain their preorder-only result. Both implement
`fMarkB` structural-supergate traversal without mutating graph marks or reference
counters. Supergate roots remain in mapped order but outside the ordinary cut
list, matching the source's node-vector result while keeping later Go emission
typed. Traversals are iterative, bounded and cancellable.

CI and CO SAT-variable collectors copy graph-order mappings, including native
unassigned `-1` entries. Literal-polarity derivation retains native head-owner
bits and four packed tail sign positions. The checked reader supports strict
token DIMACS and the source's line-oriented compatibility mode with explicit
byte/line/object/literal bounds and cancellation; malformed atoi/null behavior
is deliberately returned as an error. QDIMACS prefixes and gzip remain outside
this plain-reader function and have separate package paths.

Formula, byte and file solve entry points own a concrete BSAT session and retain
native SAT/UNSAT/undecided codes, conflict budgets and final-variable model
suffixes. Nonzero learned start, delta and percentage options override defaults
atomically. `fShowPattern` returns a copied full model, and `fVerbose` returns
deterministic formula counts and solver statistics rather than coupling library
code to stdout or wall-clock timing. Every exposed SAT model is independently
checked against all clauses before return.

Permanent rows cover native scan order/reference/area, polarity bytes, reader
quirks and 33 default solve outcomes. Independent tests cover supergate
reachability, ownership, exact limits, invalid policies, patterns/diagnostics,
zero-variable and trivially inconsistent formulas. Fresh Go-only scan and solve
fuzz campaigns each passed 5,000 executions. The complete suite passes in
10.475s, race in 59.835s, and vet/formatting are clean, promoting `cnfUtil.c` to
validated without a compiler, executable or C runtime oracle.


## CNF manager/data completion checkpoint

`src/sat/cnf/cnfMan.c` is complete without introducing a stateful compatibility
manager. `Cnf_ManStart` and `Cnf_ManStop` only acquire/release the four-input SOP
table, flexible cut storage and temporary truth/ISOP buffers. The Go port reuses
the validated immutable lazy SOP table and allocates cut, truth and composition
state per operation with explicit bounds and cancellation. This removes shared
scratch lifetime and arena aliasing while preserving every observable algorithm.

Owned `AigMappedCNF` and `CNFFormula` values cover allocate/free, deep clone,
literal/object lift, selected-occurrence flip/lift and graph-phase transforms.
Diagnostic and DIMACS serializers preserve raw/one-based numbering, quantifier
lines, empty clauses and native text order; gzip has deterministic Go framing
with identical decompressed content. File wrappers serialize before create,
return filesystem failures and safely handle short names. Go values and garbage
collection replace explicit frees.

`AigCNFFrames` preserves native per-frame clause order, register-input to next-
frame register-output equivalences, optional zero initialization, base offsets
and source immutability. It now constructs or atomically loads both legacy BSAT
and concrete solver2 sessions. Solver2 assigns stable positive source-order IDs
instead of anonymous zero IDs, retains copied clauses and remains eligible for
proof attachment after loading. Trivial inconsistency returns no partial solver.
The solver2 output-OR path likewise stages variable growth and its copied clause
atomically. Legacy output OR, all-output AND and four-clause XOR paths retain
their existing checked concrete implementations.

All 48 lift/polarity cases, 96 native frame streams, 36 constraint/XOR cases,
84 selected-flip cases, DIMACS/gzip/file fixtures and diagnostic streams remain
covered. Frame trajectories now execute exhaustively through both solver
generations, including initialization and deliberately disconnected-state
negation. A fresh 5,000-case Go-only frame campaign checks both solvers against
an independent transition recurrence. The complete suite passes in 10.539s,
race in 60.190s, and vet/formatting are clean, promoting `cnfMan.c` to validated.


## Ranked BMC unrolling completion checkpoint

Complete `src/sat/bmc/bmcUnroll.c` with one concrete incremental
`GiaRankedUnroller`. Replace packed integer handles and variable-size C objects
with typed per-node records and bounded circular `GiaLit` histories. Discover
tents and fanin ranks iteratively from primary outputs through register
boundaries, preserving frame-major primary-input lookup and primary-output
emission. Own a checked source snapshot so caller mutation cannot affect an
active manager. Make multi-frame additions transactional by applying them to an
owned clone and committing only after all object, mapping and cancellation
checks succeed. Return copied rank profiles and statistics instead of printing;
Go ownership replaces explicit allocation/free. Retain the existing simple
unroller as an independent implementation and support native buffer, XOR and
mux nodes in the ranked path.

Test only in Go. Exhaustively compare ranked and simple results for a sequential
machine through four frames and every input trace. Test one-frame incremental
snapshots, exact object/mapping/rank limits, cancellation atomicity, lifecycle,
source independence and native-node truth tables. Fuzz generated mixed
AND/XOR/mux graphs against the simple implementation without retaining fuzz
artifacts. Keep the package guard rejecting native source/header files and any
permanent `testdata` directory. This checkpoint passed 5,004 fuzz executions,
the complete suite in 10.792s, race in 60.607s and vet, promoting
`bmcUnroll.c` to validated.


## RSB manager ownership checkpoint

Port all four active definitions in `src/bool/rsb/rsbMan.c` as one concrete,
noncopyable `RsbManager` in the common package. Preserve the source's 20-leaf
and 200-divisor construction bounds, `nDivMax+150` counterexample capacity,
packed truth-word capacity and five independent scratch families. Use typed
`uint64`, `uint32` and `int` slices rather than generic vector handles. Replace
the two mutable-vector accessors with copied snapshots and explicit atomic
replacement methods so callers cannot alias manager storage. Make close
idempotent and reject use after close. Keep decomposition algorithms in the
remaining RSB sources separately tracked; this manager establishes their owned
state without introducing an interface.

Test every supported leaf count and both upper boundaries, exact scratch
capacities, invalid dimensions, failed-update atomicity, caller/result aliasing,
nil use and repeated close. Fuzz arbitrary divisor-index sequences against a
copied slice model without retaining generated artifacts. This checkpoint
passed 5,000 Go-only fuzz executions, the complete suite in 10.536s, race in
60.705s and vet, promoting `rsbMan.c` to validated.


## Command flag-table checkpoint

Port all three active definitions in `src/base/cmd/cmdFlag.c` as a concrete
owned `CommandFlags`. Use `insdict.Dict[string,string]` because later command
diagnostics and persistence may enumerate the table and require deterministic
insertion order. Preserve read-not-found, replacement, delete-not-found and the
source's null-value-to-empty-string update behavior. Bound new entries while
allowing replacement at exact capacity, and make failed insertion atomic.
Return copied ordered entry records rather than exposing dictionary storage.
Keep the broader command frame and parser separately tracked.

Test empty keys, missing reads/deletes, null and ordinary updates, replacement
position, delete/reinsert order, exact capacity, invalid construction and nil
receivers. Fuzz mixed updates, reads and deletes against an independent ordered
slice model. This checkpoint passed 5,000 Go-only fuzz executions, the complete
suite in 10.659s, race in 60.541s and vet, promoting `cmdFlag.c` to validated.


## Command alias-table checkpoint

Port all four active definitions in `src/base/cmd/cmdAlias.c` as a concrete
owned `CommandAliases`. Store aliases in `insdict.Dict[string,CommandAlias]` so
enumeration remains deterministic. Copy each argument vector on insertion and
on retrieval, preserve first-token lookup and unchanged fallback for an unknown
command, and return the source's fixed-width printable row as a string rather
than writing through a frame-owned stream. Reject duplicate aliases as a
checked invariant error, bound new entries atomically and let Go lifetime plus
table deletion replace explicit free.

Test formatting, empty but non-null names, invalid empty expansions, duplicate
and exact-capacity failures, caller/result ownership and delete/reinsert order.
Fuzz add, lookup and delete operations against an independent ordered-slice
model. This checkpoint passed 5,000 Go-only fuzz executions, the complete suite
in 10.568s, race in 60.525s and vet, promoting `cmdAlias.c` to validated.


## Command history checkpoint

Port all four active definitions in `src/base/cmd/cmdHist.c` as an owned
`CommandHistory`. Preserve batch suppression, removal of one trailing newline,
the source's exact command-prefix and trailing-question filters, and the rule
that a duplicate among the last ten entries moves to the end while an older
duplicate is appended. Bound entry count and command bytes. Replace the
platform-only implicit `abc.history` access with atomic byte loading and
caller-selected file paths on every platform; return printable history text
instead of writing to process-global stdout.

Test every filter, recent/old duplicates, copied entries, negative/exact
limits, batch behavior, atomic malformed loads, selected-tail serialization,
formatting and temporary-file round trips. Fuzz arbitrary admitted commands
against an independent ordered-slice model. This checkpoint passed 5,000
Go-only fuzz executions, the complete suite in 10.569s, race in 60.597s and
vet, promoting `cmdHist.c` to validated.


## Bi-decomposition function-cache checkpoint

Port all four active definitions in `src/bool/bdc/bdcTable.c` as the concrete
owned `BdcTable`. Preserve support-indexed newest-first insertion, direct-match
precedence over complemented matches, on/off restoration, constant-one
function zero and selective occupied-bucket clearing. Combine native table
clearing with owned function-arena reset so Go callers cannot retain dangling
pointers. Validate support masks, dimensions, disjoint ISFs and the unique-mask
subset; bound function storage, copy all truth data and make cancellation
atomic. Represent complemented function pointers with the typed `BdcLit`, not
an interface or pointer tagging.

Prove containment exhaustively for all three-variable truth functions and all
disjoint on/off pairs using an independent assignment relation. Test direct and
complement lookup order, constants, clear/reprepare, ownership, malformed ISFs,
exact limits and cancellation. Fuzz arbitrary cache contents and ISFs against
an independent newest-first slice model. This checkpoint passed 5,012 Go-only
fuzz executions, the complete suite in 15.551s, race in 81.827s and vet,
promoting `bdcTable.c` to validated without native source or a runtime oracle.


## FRAIG prime-schedule checkpoint

Port the complete `s_FraigPrimes` data definition in
`src/proof/fraig/fraigPrime.c` as package-owned fixed storage for the later
FRAIG simulation hashes. Generate the mathematically defined first 1,024
primes rather than retaining a copied native table. Provide checked indexed
access and copied full snapshots; do not expose mutable package storage.

Compare every generated entry with an independently implemented sieve, pin
representative positions and the terminal value, and test ownership and all
index boundaries. Fuzz arbitrary indices while independently checking both
primality and ordinal rank. A migration-only audit confirmed all 1,024 source
entries are precisely this sequence; retained tests require only Go. This
checkpoint passed 5,000 fuzz executions, the complete suite in 15.680s, race
in 85.682s and vet, promoting `fraigPrime.c` to validated without a C oracle.


## MSAT propagation-queue checkpoint

Port all six definitions in `src/sat/msat/msatQueue.c` as the concrete owned
`MsatQueue`. Preserve its fixed allocation and first/last cursor behavior,
including the non-circular rule that extracted slots are not appendable again
until `Clear`. Replace assertion failure at capacity with `ErrLimit`, replace
the overloaded `-1` empty result with a typed found flag, and use idempotent
close for manual free. Keep arbitrary integer literals valid because the source
queue itself imposes no literal-domain check; return copied pending snapshots.

Test FIFO order including a literal value of `-1`, zero and exact capacities,
post-extraction exhaustion, clear epochs, ownership, closed/nil errors and
idempotent lifecycle. Fuzz arbitrary insert/extract/clear sequences against an
independent fixed-array cursor model. This checkpoint passed 5,000 Go-only
fuzz executions, the complete suite in 15.984s, race in 86.554s and vet,
promoting `msatQueue.c` to validated without native source or runtime oracles.


## MSAT integer-vector checkpoint

Port every definition in `src/sat/msat/msatVec.c` as the concrete bounded
`MsatIntVector`, establishing the storage used by later MSAT clause creation,
reason analysis and solver state. Preserve the native 1-through-15 capacity
rounding to 16, exact explicit growth, 16-or-double push growth, fill, shrink,
clear, uniqueness return convention, ordered-unique bubbling, pop and signed
ascending/descending sorting. Replace assertion failures with typed errors and
manual free with idempotent close.

Model raw-array ownership explicitly: Take clears the caller's slice header,
Transfer moves storage between vectors, Release moves the active prefix to the
caller, Clone preserves capacity without aliasing, and ordinary Values reads
copy. Test every operation, exact capacity transitions, zero/small configured
limits, move/reuse behavior, ordering, aliases and failures. Fuzz arbitrary
push/pop/clear/fill/sort/unique/ordered/shrink sequences against an independent
slice model. This checkpoint passed 5,010 Go-only fuzz executions, the complete
suite in 15.355s, race in 86.676s and vet, promoting `msatVec.c` to validated
without C source, native oracles or permanent `testdata/` output.


## MSAT clause-vector checkpoint

Port all 11 definitions in `src/sat/msat/msatClauseVec.c` as the concrete
bounded `MsatClauseVector`, establishing stable clause-reference storage for
the learned/problem databases and watched lists. Preserve native positive
capacities below 16 rounding to 16, exact explicit Grow, 16-or-double Push,
shrink, clear, pop and indexed replacement/read behavior. Keep nil references
representable. Return copied pointer-array snapshots so callers retain clause
identity without gaining structural mutation access; replace assertions with
typed errors and manual free with idempotent close.

Test all operations, exact growth boundaries, nil entries, pointer identity,
snapshot aliasing, zero/small limits and closed/nil failures. Fuzz arbitrary
push/pop/clear/set/shrink/read sequences against an independent pointer-slice
model. This checkpoint passed 5,000 Go-only fuzz executions, the complete suite
in 15.886s, race in 85.125s and vet, promoting `msatClauseVec.c` to validated
without C source, native oracles or retained generated fixtures.


## MSAT learned-clause sorting checkpoint

Port every definition in `src/sat/msat/msatSort.c` directly over the concrete
`MsatClause` identity and `MsatClauseVector`. Preserve the fixed 91648253 seed,
the native multiply/modulo pseudo-random pivot recurrence, selection sorting at
sizes up to 15, recursive two-way partitioning and float32 comparison rules.
Do not add a stable tie-break: equal activities intentionally follow native
pointer permutation, and NaN comparisons retain their source behavior. Validate
all clause references before mutation so malformed input fails atomically.

Test every size through both sorting paths, deterministic duplicates and equal
runs, pivot recurrence, infinities, NaN termination, vector integration,
pointer permutation and invalid/closed clauses. Fuzz finite activities against
independent ascending-value and identity-permutation oracles while repeating
each run to prove deterministic pointer order. This checkpoint passed 5,000
Go-only fuzz executions, the complete suite in 15.778s, race in 85.416s and
vet, promoting `msatSort.c` to validated without C source or runtime oracles.


## MSAT activity-order heap checkpoint

Port all 15 definitions in build-selected `src/sat/msat/msatOrderH.c` as the
concrete `MsatVariableOrder`; keep the unbuilt `msatOrderJ.c` as a separate
later scope item. Preserve one-based heap/index arrays, strict-greater activity
comparison, source-order Clean without heapification, destructive top removal,
lazy discarding of assigned variables during selection, no-op assignment heap
updates, reinsertion on unassignment, upward-only activity updates and exact
up/down percolation. Own assignment and activity arrays in this concrete type
so `msatActivity.c` can integrate directly without an interface or callback.

Add checked bounds, atomic malformed-cone rejection, uniform nonnegative
activity rescaling, copied heap/activity diagnostics and idempotent close. Test
fixed tie order, maximal selection, lazy assigned removal, reinsertion,
resizing, rescaling, exact limits, aliases, invariants and lifecycle. Fuzz
arbitrary bumps, assignments, unassignments, selections and rescaling while an
independent model checks maximal unassigned choice, memberships, activities and
the heap after every operation. This checkpoint passed 5,000 Go-only fuzz
executions, the complete suite in 15.836s, race in 86.042s and vet, promoting
`msatOrderH.c` to validated without C source or runtime oracles.


## MSAT activity checkpoint

Port all six definitions in `src/sat/msat/msatActivity.c` as one concrete
`MsatActivityState` sharing the already owned `MsatVariableOrder` and
`MsatClauseVector`; use no interface or callback layer. Preserve negative
variable decay as static-order mode, float64 variable increment multiplication,
heap update after bumps, uniform 1e-100 rescaling when activity exceeds 1e100,
float32 clause storage/addition, clause increment decay and learned-database
1e-20 rescaling before a bump exceeding 1e20.

Validate configuration and finite arithmetic, preflight whole-database clause
rescaling so errors are atomic, copy diagnostic snapshots and detach lifecycle
without owning or closing the shared heap/vector. Test static and ordinary
variable modes, decision-order effects, both automatic thresholds, explicit
rescaling, float32 rounding, external and learned clauses, malformed data,
aliases and lifecycle. Fuzz mixed variable/clause bumps, decays and rescales
against an independent float64/float32 recurrence while checking the heap after
every operation. This checkpoint passed 5,000 Go-only fuzz executions, the
complete suite in 15.711s, race in 81.450s and vet, promoting `msatActivity.c`
to validated without C source or runtime oracles.


## MSAT clause checkpoint

Port all 21 active definitions in `src/sat/msat/msatClause.c` as concrete
`MsatClause` operations coordinated by `MsatClauseContext`. Share the existing
decision heap, activity state and stable clause-vector identity directly; do
not introduce solver, watch-list or activity interfaces. Preserve top-level
original-clause reduction, learned second-watch selection by highest decision
level, clause/variable bumps, complement-indexed watch lists, two-watch
propagation, unit/conflict reporting, false-tail simplification, complemented
reason construction, locking, fields, removal and diagnostic rendering.

Intentionally repair the C file's documented polarity bug: identical literals
collapse and opposite polarities make an original clause tautologically
satisfied. Keep this compatibility difference explicit in the coverage ledger.
Use copied caller inputs and literal snapshots, checked bounds and assignments,
finite activity writes, atomic activity rollback on failed learned creation and
idempotent close. Clause creation owns neither the problem nor learned database
append, matching the source caller boundary; it owns watch installation.

Test empty, unit, satisfied, tautological, duplicate, assigned and stored
original clauses; learned watch ordering and activity; all fields and copied
views; satisfied, moved-watch and unit propagation; simplification; reason and
locking; duplicate watch buckets; removal, formatting, exact limits, failures
and lifecycle. Fuzz original normalization against an independent canonical
model plus exhaustive truth evaluation on every consistent four-variable
assignment. Separately fuzz watched propagation against an independent state
transition model. Both campaigns passed 5,000 Go-only executions, the complete
suite in 15.615s, race in 81.064s and vet, promoting `msatClause.c` to validated
without C source, runtime oracles or retained generated fixtures.


## MSAT search-state and complete API checkpoint

Complete all remaining definitions in `msatSolverApi.c` together with all 13
active operations in `msatSolverSearch.c`, using the same concrete CDCL engine
as the core checkpoint. Expose copied decision levels, reasons, assignments,
models, trails, learned clauses and watched clause IDs; concrete memory stats,
seen epochs, allocation counters, factors, adjacency, cone and used-variable
state; and checked lifecycle, marking and Prepare operations. Stable integer
clause IDs and value snapshots replace borrowed pointer arrays.

Add an optional concrete enqueue mask inside the shared solver state. It is nil
for every existing BSAT user, while MSAT installs its owned variables-used
slice so both explicit assumptions and watched-clause implications preserve
`Msat_SolverEnqueue`'s cone filtering. Clone the mask with bookmarks and trial
states; do not introduce a callback or solver interface. Prepare clears
transient assignments and installs exactly the requested decision order while
retaining the clause databases, matching the source's incremental-cone role.

Represent assumption, propagation, cancellation, conflict analysis and learned
recording as checked operations over the same trail/reason/watch machinery used
by solving. Delegate reduction to the validated atomic learned compactor.
Implement learned removal and recently-added-original rollback by rebuilding
from every retained live original core clause, including units, then resync
stable original IDs; never reconstruct only the public multi-literal list.

Test direct propagation conflicts, first-UIP clauses and backtrack levels,
learned unit/nonunit recording, copied learned/watch/reason/trail views,
reduction/removal, marked rollback, Prepare order, disabled-variable
implications, adjacency/seen/allocation/memory state and every malformed-state
path. Fuzz incremental assumptions and watched unit propagation against an
independent clause-scanning model. The model passed 5,022 executions, the full
suite in 15.768s, race in 81.839s and vet, promoting both
`msatSolverSearch.c` and `msatSolverApi.c` to validated without C source,
runtime oracles or retained generated fixtures.


## MSAT J-boundary ordering checkpoint

Port the build-disabled `src/sat/msat/msatOrderJ.c` as a separate concrete
`MsatJBoundaryOrder`, not as an interface layered over the build-selected
activity heap. Own the variable activity, assignment, cone, adjacency and ring
state so the alternative can be tested without exposing or borrowing
`MsatSolver` internals. Preserve circular insertion order, root movement after
removal, adjacency-order insertion and strict-greater activity selection so
ties select the earliest boundary member.

Define the J-boundary semantically as unassigned cone variables with an
assigned cone neighbor. Keep outside-cone assignments isolated, return copied
state views, validate every endpoint and duplicate, and replace pointer-ring
assertions with structural and semantic checks. Cone, adjacency and size are
configuration-time operations that reject active assignments or a nonempty
boundary atomically; resizing preserves surviving quiescent activity, cone and
valid adjacency state while clearing the ring. Close is idempotent.

Test empty/singleton/multi-entry transitions, stable ties, root removal,
assignment and backtracking, outside-cone isolation, resizing, ownership,
invalid operations and lifecycle. Fuzz mixed assignments, unassignments and
activity writes against an independent insertion-ordered state model while
checking the boundary after every operation. The model passed 5,000 Go-only
executions, the complete suite in 15.490s, race in 81.962s and vet, promoting
`msatOrderJ.c` to validated without a C oracle or retained generated fixture.


## MSAT header and build-manifest reconciliation checkpoint

After validating every selected MSAT translation unit, reconcile
`src/sat/msat/module.make`, `msat.h` and `msatInt.h` as first-class ledger
entries. Confirm that the manifest's exact 13 selected sources each have a
validated concrete Go implementation, while the independently validated
`msatOrderJ.c` alternative remains correctly classified as build-disabled.

Map the public status/literal macros, solver/vector types and every linked
external declaration to the concrete Go surface. Map internal solver fields,
statistics, search parameters, clauses, vectors, queue, ordering and memory
managers to owned Go state and copied accessors rather than reproducing C
layouts. Search the whole source snapshot before treating declaration-only
symbols as requirements: `Msat_SolverReadTruth`, proof/variable-map/solution
accessors, the obsolete `Msat_VarHeap` family, and the fake/truth clause
helpers have declarations but no definitions. Record that disposition instead
of fabricating behavior with no source semantics.

Ledger parsing, source hashes and every referenced Go filename were checked;
all selected MSAT source, both headers, the manifest and the disabled ordering
alternative are now validated. The global inventory is 230 validated, 40
partial and 1,487 unstarted files. Full behavioral gates remain those recorded
for the concrete MSAT implementations above.


## LSAT declaration-only header disposition

Audit `src/sat/lsat/solver.h` separately from implemented SAT engines. The file
contains declarations only: no inline algorithms or constant definitions, no
manifest selects it, no source includes it, and a whole-tree symbol search
finds no definition for its construction, clause, solve, value, model or
conflict API. Do not manufacture a wrapper with guessed semantics merely to
mirror those names. Record the header as a validated no-runtime disposition;
the concrete validated bsat and MSAT ports supply actual Go SAT behavior. This
moves the inventory to 231 validated, 40 partial and 1,486 unstarted files.

The same evidence rule applies to truly empty headers. `aig/gia/giaIiff.h`
contains only guards and headings, with no declaration, inline logic, data,
include or caller, so it has a validated no-runtime disposition. The inventory
thereafter is 232 validated, 40 partial and 1,485 unstarted files.


## GIA SAT-mapping manager completion checkpoint

Complete `src/aig/gia/giaSatMap.c` over the validated concrete SAT solver and
pairwise cardinality network. Replace its untyped `void *` call into the Nf
mapper with an explicit owned `GiaSatMapProblem`: object-literal count,
complemented-input variables, roots, cuts, typed fanin dependencies, gates,
areas and the current selection. This keeps extraction and SAT optimization
separable without introducing an interface or borrowing mapper vectors.

Validate that every literal polarity owns at least one cut, current selections
are unique, roots are selected, and every selected cut's internal or input-
inverter dependency is also selected. Preserve cardinality variable layout,
cut-variable allocation, per-object implication clauses, paired final-cut
exclusivity, fanin clauses, preferred-true variable order, negative threshold,
negative padding and positive root assumptions. Decode satisfiable models into
ordered objects, inverters, cuts and gate IDs with overflow-checked area.

Expose single-bound solving and the source's decreasing-bound optimization as
owned results. Stop safely on UNKNOWN rather than repeating the same query
forever as the diagnostic C loop can. Test raw clause suffixes, solution
closure, bound transitions, optimization, areas, caller ownership, budgets,
cancellation and lifecycle. Fuzz chain dependencies and optional input
inverters against an independent minimum-cardinality model. The model passed
5,000 executions, the full suite in 15.546s, race in 82.292s and vet, promoting
`giaSatMap.c` from partial to validated. The inventory is now 233 validated, 39
partial and 1,485 unstarted files.


## BMC timeframe-builder checkpoint

Complete all five active definitions in `src/sat/bmc/bmcBmc.c`. The concrete
`BuildBmcFrames` operation emits a zero-initialized combinational image with
frame-major primary inputs and outputs, preserving structural hashing and the
source's cumulative unique-cone threshold: a positive bound is checked only
after every output in a frame has been created, so the returned image may
cross the threshold but never contains a partial frame. Iterative owned marks
replace traversal IDs and recursion. Hard object/mapping limits, cancellation,
source validation and atomic failure wrap the source's unchecked allocation.

Use the existing concrete `AigerModel.BoundedCheck` for the per-output SAT loop
and checked witness extraction. Its reference Tseitin encoding, explicit latch
initialization and positive constraint-prefix semantics replace the C-only
optimized-CNF, rewrite/cofactor and Satoko selection switches with one native
Go path. The SAT solver's owned model replaces `Sat2_SolverGetModel`; no backend
interface or native runtime is retained.

Test exact completed-frame cutoff, duplicate-output cone accounting, source
ownership, constraints-as-outputs, limits, cancellation and a 10,000-AND cone.
Exhaustively compare a three-frame sequential machine over every primary-input
trace against independent step-by-step simulation. Fuzz random sequential AIGs
against the same model for 5,000 executions. The focused tests and fuzzing pass,
the full suite passes in 15.744s, race in 82.188s and vet is clean, promoting
`bmcBmc.c` from partial to validated. The inventory is now 234 validated, 38
partial and 1,485 unstarted files.


## BMC lazy-CNF checkpoint

Complete all six active definitions in `src/sat/bmc/bmcLoad.c` with one
concrete owned `GiaLazyCNF`, not a callback interface. Preserve the source's
reserved SAT variable zero, constant-zero negative unit, first-use object
variable allocation and exact positive and negative directional AND clauses.
Expose atomic single-direction and complete-cone loading with copied clause
views, checked limits, cancellation, statistics and idempotent close.

Replace the solver's untyped on-demand callback with counterexample-guided
native refinement. In lazy mode, load the requested output direction, solve,
directly compare every mapped gate assignment with its fanins, load the missing
direction of the first inconsistent gate and repeat. Accept SAT only after the
candidate also evaluates true in the owned GIA snapshot. UNSAT from any loaded
directional subset is sound because every emitted clause is implied by the
full AND relation. Eager mode loads both directions throughout the cone.

Pin exact clause and variable order, ownership, repeated directions, eager and
lazy counts, constants, limits, cancellation, atomic rollback and lifecycle.
Exhaustively solve all 256 three-input Boolean functions against direct truth
enumeration and fuzz plain GIA cones against the same independent evaluator.
The fuzz campaign passed 5,000 executions, the full suite in 15.816s, race in
81.762s and vet, promoting `bmcLoad.c` from unstarted to validated. The
inventory is now 235 validated, 38 partial and 1,484 unstarted files.


## BMC packed-generation checkpoint

Complete all six active definitions in `src/sat/bmc/bmcGen.c`. Return an owned
object-major `GiaPackedSimulation` rather than storing `vSims` and `nSimWords`
in graph scratch. Preserve the exact reset multiply-with-carry random stream,
CI word order, complemented-fanin AND evaluation and CO propagation. Add
checked word/entry limits, cancellation, copied whole-arena/object views and
bit access. Reject buffer, native-XOR and mux objects because the source loop
classifies all of them as `Gia_ObjIsAnd` and would mis-evaluate their semantics.

Represent `Gia_ManTestSatEnum` honestly as a diagnostic result. Its collected
simulation variables never affect solving because the polarity calls are
commented out; therefore build one output-constrained native formula and solve
it once per packed sample, returning the output bits, ordered statuses, SAT
count and final statistics instead of stdout and wall-clock timing. Do not
claim distinct assignment enumeration.

Pin exact RNG words and arena layout, independently scalar-evaluate every bit,
check reset repeatability/source ownership, constants, repeated SAT/UNSAT and
cancelled UNKNOWN outcomes, bounds, malformed inputs, unsupported nodes and a
10,000-gate cone. Fuzz random plain GIA cones and compare every packed sample
with direct graph evaluation. The campaign passed 5,000 executions, the full
suite in 15.778s, race in 81.598s and vet, promoting `bmcGen.c` from unstarted
to validated. The inventory is now 236 validated, 38 partial and 1,483
unstarted files.

## BMC initial-state maximization checkpoint

Complete all four active definitions in `src/sat/bmc/bmcMaxi.c` with concrete
owned Go APIs. Represent the six initialization integers as `GiaInitialClass`:
known zero/one, relaxed zero/one and fixed zero/one. Unfold a sequential GIA
for a checked positive frame count using concrete control/data CIs, frame-major
primary inputs, ordinary-AND hashing and final register-input outputs. Preserve
the source distinction between fixed unfolding and variable unfolding, but do
not mutate source object scratch. Reject buffer, XOR and mux extensions because
the source's `Gia_ManForEachAnd` path would interpret them as ordinary ANDs.

Implement maximization by building fixed and variable unfolds, a semantic GIA
miter, explicit GIA-to-AIG conversion and owned reference CNF. Add one clause
requiring an output difference, initially assume false every control for a
known-zero or known-one register, and repeatedly replace assumptions with the
negation of the solver's final UNSAT conflict until its cardinality reaches a
fixed point. Preserve fixed classes and relax only known classes omitted from
the final core. Return copied classes, retained register indices, SAT status,
stop reason, iteration count and statistics instead of stdout and timings.
Bound graph objects, mappings, variables, clauses, literals and iterations;
support cancellation without retaining solver or graph state.

Exhaust every one of the 36 class pairs in both unfolding modes and all 64
control/data/two-frame assignments against independent `GiaManager.Step`
evaluation. Check known/fixed/relaxed core outcomes by exhaustively comparing
the generalized unfold with its fixed baseline, plus source ownership, copied
views, malformed inputs, unsupported objects, resource limits, cancellation
and a 10,000-gate cone. A standalone fuzzer generates small sequential AND
graphs and independently exhausts every generalized input assignment. It
passed 5,054 executions; the complete suite passed in 15.721s, race in 82.361s
and vet. This promotes `bmcMaxi.c` from unstarted to validated. The inventory
is now 237 validated, 38 partial and 1,482 unstarted files.


## BMC insertion-based initial-state checkpoint

Complete all ten active definitions in `src/sat/bmc/bmcInse.c`, including the
three inline `Gia_ParTest` arena helpers, with concrete owned Go state. Store
separate object-major definitely-zero and definitely-one bitsets, preserving
the source's ternary AND rules, complemented-edge lane swapping, RI-to-RO
transfer and exact reset `Gia_ManRandomW` stream. Do not borrow `pData`,
`iData`, object marks or another simulation interface. Reject extended buffer,
XOR and mux objects because the source loop interprets every GIA AND-category
object as an ordinary AND.

For each frame, simulate every packed sample, count the binary next-state bits,
select the earliest pattern with maximum count, append that pattern's primary
inputs and reinitialize the RI target from its ternary next state. Replay the
selected trace from the concrete starting state and encode the final register
as known zero/one when its target remains unknown or fixed zero/one when the
target is binary. A nil starting state reproduces `Gia_ManInseTest`'s explicit
all-zero assumption; a supplied state preserves `Gia_ManInsePerform`'s useful
lower-level input. Return the trace, target, pattern indices and unknown counts
instead of stdout and timing. Omit the wrapper's unused timeout and simulation
arguments. Add checked frame/word/arena/trace limits, cancellation and safe
zero-primary-input handling around assertion-based source behavior.

Compare packed selection to an independent model that evaluates every sample
through `EvaluateTernary`, including exact random lane polarity, earliest-tie
selection and concrete trace replay. Test zero frames, constant/fixed results,
explicit and default starts, ownership, malformed dimensions, every resource
cap, cancellation, unsupported extended nodes, zero PIs and a 10,000-gate
cone. Fuzz small sequential AND graphs against the independent model; 5,000
executions passed. The complete suite passed in 15.756s, race in 82.333s and
vet. This promotes `bmcInse.c` from unstarted to validated. The inventory is
now 238 validated, 38 partial and 1,481 unstarted files.


## BMC mesh-embedding checkpoint

Complete all seven active definitions in each of `src/sat/bmc/bmcMesh.c` and
`bmcMesh2.c` through one concrete owned Go implementation. The two sources
have identical mesh CNF and refinement logic; they differ only in Satoko versus
bsat calls and one explicit topological assertion. Reuse the owned native SAT
solver directly rather than adding a backend interface. Preserve the four
cell-layout helpers, complete-model value access, and the lazy one-hot helper
that adds every pairwise negative clause among currently true variables.

Allocate each cell's contiguous time, graph, 16 configuration and used
variables in y-major cell order. Preserve time-zero boundary cells, PI-only
boundary graphs, the interior time-zero prohibition, graph-implies-used,
used-implies-some-time/configuration, four buffer directions, twelve ordered
node direction pairs and predecessor time/graph implications. Preserve the
source's final condition on graph item `G-1` at `(1,1)` rather than silently
changing it to the CO driver. Re-solve after each complete-model one-hot
refinement round. Return concrete row-major cells, buffer/node counts, SAT
status/reason/statistics and construction/refinement counts instead of stdout
and wall-clock timing.

Replace the fixed 102×102 arrays and assertions with checked 3..100 dimensions,
2..100 latency, at most 100 graph items, canonical combinational GIA topology,
and variable/clause/literal/iteration limits plus cancellation. Reject constant
fanins, extended buffers/XORs/muxes, sequential graphs and object-order gaps
that the native index arithmetic cannot represent. Test PI buffering, one
node, a short impossible chain, successful multi-level 4×4 routing, the `G-1`
target quirk, exact refinement pairs, source ownership, every limit and
cancellation. Fuzz canonical 3×3/latency-2 graphs against the independent fact
that only a PI or a node with two PI fanins can occupy the sole interior cell,
and validate each SAT placement structurally. The campaign passed 5,000
executions; the complete suite passed in 15.573s, race in 82.623s and vet. This
promotes both mesh files from unstarted to validated. The inventory is now 240
validated, 38 partial and 1,479 unstarted files.


## BMC assignment-justification enumeration checkpoint

Complete all three active definitions in `src/sat/bmc/bmcEnum.c`. Derive a
phase-specific justification by evaluating every CI/ordinary AND, marking all
CO drivers, and walking ANDs backward: equal effective phases require both
fanin justifications and become AND for phase one or OR for phase zero, while
mixed phases retain only the controlling-zero fanin. Build an owned GIA with
one CI per source CI, preserve name/spec and the register partition, and avoid
source phase, mark and `Value` scratch. Use local safe simplification where the
raw native append would assert on collapsed duplicate evidence. Reject extended
AND kinds and any result that actually depends on the source's broken constant
temporary value, which is literal zero rather than a true equality predicate.

Represent `Gia_ManDeriveOneTest2` as an explicit all-zero wrapper. Replace the
stdout enumeration driver with a concrete result and configurable selected
output/phase. Eagerly encode the selected original output, obtain each SAT CI
assignment, derive the complete whole-output justification, then append exact
Tseitin definitions sharing the original CI variables. Carry the negation of
the selected derived root as an assumption on all later solves so the complete
covered formula is excluded, not just its originating model. Return copied
assignments and observed outputs, owned justification graphs, completion,
status, stop reason, statistics and construction counts.

Add per-graph objects/mappings, total retained objects, SAT variables/clauses/
literals and assignment-count limits plus cancellation. Exhaust all source and
candidate assignments for a small sequential multi-output graph, proving each
derived output is true at its origin and implies the matching observed phase.
For both XOR phases, prove that later records satisfy no earlier root and that
the formula union covers exactly all requested assignments. Cover ownership,
partial-limit reporting, solver UNKNOWN, malformed options, unsupported
constants/extensions and a 10,000-gate cone. Fuzz random plain GIA graphs and
exhaust every candidate assignment against the independent implication rule;
5,000 executions passed. The complete suite passed in 15.740s, race in 82.270s
and vet. This promotes `bmcEnum.c` from unstarted to validated. The inventory
is now 241 validated, 38 partial and 1,478 unstarted files.


## BMC learned-core checkpoint

Complete all six active definitions in `src/sat/bmc/bmcBCore.c`. Replace the
unchecked whitespace pivot reader with bounded parsing and file loading, and
return invalid object IDs rather than printing them. Expand each valid object
ID across the object/frame variable map in SAT-variable order. Preserve the
source's dummy variable zero, output variable one, descending timeframe order,
per-frame constant-zero variable, fanin-zero-first DFS assignment, PO/RI root
order, paired constraint clauses, inter-frame RI buffers, ordinary-AND Tseitin
clauses and auxiliary variables representing register outputs at frame -1.
Keep all traversal and variable state owned rather than borrowing GIA marks or
`Value` scratch. Reject extended buffers, XORs and muxes instead of reproducing
the source's ordinary-AND misencoding.

Use the concrete Go CDCL solver with exact resolution-DAG recording; do not add
a solver or proof interface. On UNSAT, independently replay the final proof,
derive and re-solve its original-clause core, then expose every reachable
learned clause in proof-handle order including the final empty clause. Render
these clauses with object(frame) literals and offer checked concrete file
output instead of stdout. Preserve pivot variables as returned data. The ABC
bsat field accepts the list but has no read site in this source snapshot, so it
has no search or proof effect to emulate.

Test exact pivot parsing/collection, variable maps and clause counts; exhaust
the combinational projection under every CI assignment; check zero-initialized
alternating-register reachability across multiple frames; independently solve
every reported root core; replay and format the learned cone; cover SAT,
UNSAT, UNKNOWN, malformed dimensions, all resource limits, cancellation,
ownership, unsupported extensions and a 10,000-gate cone. Fuzz random plain
GIA graphs against exhaustive Boolean satisfiability with no C oracle; 5,000
executions passed. The complete suite passed in 15.883s, race in 82.174s and
vet. This promotes `bmcBCore.c` from unstarted to validated. The inventory is
now 242 validated, 38 partial and 1,477 unstarted files.


## BMC SOP-expansion checkpoint

Complete all three active definitions in `src/sat/bmc/bmcExpand.c` with owned
`SOPCover`, `SOPNode` and `SOPNetwork` values. Preserve the raw cover's ordinary
`0`/`1` and special parity `n`/`x` output codes. For the lower-level expansion,
map cover columns to caller SAT variables, visit cubes in storage order and try
removing literals from the last column to the first, retaining a literal exactly
when the remaining cube intersects the preconstrained offset. Return an atomic
UNKNOWN result on SAT budget exhaustion rather than reproducing the source's
partially cleared cube.

For graph-backed expansion, use the concrete Go GIA-to-AIG, reference CNF and
CDCL path without introducing an interface. Check both output phases first and
canonicalize constants. Otherwise constrain the offset, expand every cube and
perform the source-equivalent direct-order full irredundancy pass. Support an
explicit permutation from local SOP columns to global GIA CIs. The network
wrapper must preserve repeated/shared output-driver updates, clear fanins when
an output becomes constant, and reproduce `Abc_NtkSortSops`: weighted cube
sort, literal-count plus lexicographic column sort, lexical cube sort,
singleton-literal cube-order permutation and the final lexical cube sort for
every node.

Test parsing/formatting and parity evaluation, exact reverse literal removal,
constant phases, direct-order redundancy, arbitrary column order, shared owned
state and the complete five-stage sorting result. Cover malformed covers and
networks, graph/CNF/cube limits, construction cancellation, atomic SAT UNKNOWN,
source ownership and a 10,000-gate cone. Fuzz one- through four-variable truth
tables starting from their minterms and compare exact output cubes to an
independent exhaustive greedy/irredundancy model; 5,000 executions passed.
The complete suite passed in 15.708s, race in 82.129s and vet. This promotes
`bmcExpand.c` from unstarted to validated. The inventory is now 243 validated,
38 partial and 1,476 unstarted files.


## BMC multi-output proving checkpoint

Represent all seven active definitions in `src/sat/bmc/bmcMulti.c` with
concrete owned Go APIs while keeping the file partial until its two large
dependencies are exact. `ProcessMultiProveOutputs` must validate the complete
transfer before mutation, move counterexamples without alias-producing clones,
clear transferred input slots, place each witness by original output ID and
return both unresolved local indexes and a compacted original-output map.
Implement exact GIA/AIG syntactic constant-zero counters. Replace stdout and
wall-clock report text with structured stage, topology, solved-percentage and
relative-size data.

Use a concrete `AigerModel.MultiProve` orchestration rather than solver or
simulation interfaces. Preserve the source SIM/BMC/optional-SYN stage order,
solve-all output removal, iteration seed schedule, increasing local effort,
gap termination and the 1,000-iteration default. Replace approximate process
time with deterministic frame, iteration, simulation-work and SAT-work limits,
checked overflow and synchronous cancellation. Keep the original model
immutable and validate every discovered trace against it. Strengthen the
source's rarity pointer-one markers and reduced-cone-coordinate witnesses by
returning complete original-input/register counterexamples in original output
slots. Keep constraint-prefix semantics explicit and return constant-zero proof
status separately from unresolved outputs instead of relying on nil ambiguity.

The first partial implementation uses deterministic sequential random
simulation and a semantics-preserving AIG-to-GIA cleanup/structural-rehash
synthesis boundary. Complete the checkpoint by porting `Ssw_RarSimulate`'s
rarity groups, bins, pattern ancestry, best-pattern transfer and restarts, then
port the old-algorithm `Gia_ManAigSyn2` path used here: area balancing, JF cut
mapping and the second balance. Only after the exact dependencies are present
should native schedule/shape fixtures be added under `test_vectors/` and
`bmcMulti.c` promoted to validated.

Require unit coverage for atomic ownership, output-map ordering, constant
drivers, report denominators, sequential behavior, constraints, synthesis
equivalence, original-coordinate witness replay, source immutability, malformed
options, every logical budget and cancellation point. Fuzz output-map transfer
against an independent slice model and fuzz multi-prove results by replaying
every witness and independently checking every constant-zero claim. The first
two Go-only campaigns passed 5,000 executions each; the full suite passed in
15.703s, the race suite in 81.379s and vet passed. The inventory is now 243
validated, 39 partial and 1,475 unstarted files.


## BMC ECO patch checkpoint

Complete all five active definitions in `src/sat/bmc/bmcEco.c` with concrete
owned GIA, CNF and SAT types. Build the ECO miter with shared golden CIs plus
one replacement-function PI; select the last old PO's regular driver as the
replacement root; copy the golden graph, the old graph with positive root
substitution, and the old graph with negative root substitution. Emit one
onset mismatch per golden CO, the OR of all offset mismatches, then requested
fan signals using their offset mapping. Preserve the source's flattening of
sequential CI/CO terminals into a combinational result, source names, hashing
and cleanup without borrowing `Value` scratch. Reject buffers, native XORs and
muxes rather than reproducing ordinary-AND misencoding.

Use the concrete reference Tseitin encoder and CDCL solver. Assert every onset
mismatch false and the offset mismatch true. Add a fresh selector, repeatedly
solve replacement-root true with selector true, capture each divisor model,
query root false under that divisor assignment, and either report an
impossible projection or turn the UNSAT assumption core into a returned cube
and selector-guarded blocking clause. Preserve the source's one-million
per-query conflict default. Return owned status, stop reason, cubes and SAT
statistics; expose direct cube evaluation and replace the hardcoded file/stdout
test driver with ordinary Go tests.

Bound objects, mappings, variables, clauses, literals, iterations, cubes and
cube literals; check overflow and cancellation; do not return partial cubes on
errors. Exhaust all 16 two-input truth tables and every assignment, verify an
impossible divisor projection, miter output truth, sequential flattening,
source immutability, malformed/extended graphs, every exact limit and a
10,000-gate cone. Fuzz complete patch truth independently for 5,000 executions.
The full suite passed in 15.698s, the race suite in 81.802s and vet passed.
This promotes `bmcEco.c` from unstarted to validated; inventory is now 244
validated, 39 partial and 1,474 unstarted files.


## BMC counterexample-chain checkpoint

Complete all ten active definitions in `src/sat/bmc/bmcChain.c` with concrete
owned GIA, AIG, CNF, SAT and counterexample values. Use the replay-checked
bounded checker to find the first failure. Verify its trace through the failing
frame, retain the register state at the start of that frame, and phase-recode
both ROs and RIs so zero in the new graph denotes that reached state. Reuse the
exact `FlipRegisters` transformation instead of duplicating mark-based scratch
logic. Preserve names, specs, terminal order, register count and extended-gate
semantics without modifying the source graph.

Build the current-state combinational problem by substituting zero for ROs,
retaining PIs and POs, structurally hashing and cleaning. Derive one concrete
reference Tseitin relation and incrementally ask whether each syntactically
nonzero PO can be true. Return a zero-initialized one-frame counterexample for
each SAT answer and replay it before publication. Stop on UNKNOWN instead of
silently treating a resource-limited query as UNSAT. Replace selected POs by
false atomically, preserve RIs and unselected terminal order, then clean.

Retain the outer BMC/move/current-state-SAT/cleanup sequence, failing-frame
depth sum and 10,000-iteration default. Return owned structured segments rather
than the source's flat mixed-coordinate pointer vector: each segment records
its path, reached-state phase, failed outputs and one-frame witnesses. Return
the final graph, counts, deterministic stop reason and accumulated SAT work in
place of stdout and elapsed time. Reject environment constraints because the
source's current-state phase treats every PO as a property and has no sound
constraint handling; do not silently narrow or reinterpret them.

Unit tests pin start-of-failing-frame rather than post-transition movement,
phase-recoded transition equivalence, zero-state propagation, exhaustive
output satisfiability, witness replay, source immutability, exact cleanup and a
two-state/two-segment complete chain. Test finite and iteration stops, every
owned resource boundary, cancellation and malformed inputs. Fuzz random
two-input extended-gate DAGs and compare the complete failed-output set with
exhaustive evaluation; 5,039 executions passed. The final full suite passed in
15.770s, the race suite in 81.941s, and vet passed. This promotes `bmcChain.c`
from unstarted to validated; inventory is now 245 validated, 39 partial and
1,473 unstarted files.


## BMC counterexample-toolbox checkpoint

Complete the 21 externally meaningful definitions in
`src/sat/bmc/bmcCexTools.c` with concrete owned graphs, traces, masks and
statistics. Count only frame input columns when reporting input ones. Replace
stdout and append-only statistics files with deterministic strings and typed
results. Build initialized timeframe logic through the selected final property
using the shared GIA unroller. Preserve the ordinary-AND monotone justification
recurrences for observed one, two controlling zeros and one controlling zero;
reject native extended gates in this specialized builder rather than treating
them as ordinary ANDs.

Expand each trace into PI/RO values and a parallel implication mask. Carry
actual and implied register inputs simultaneously across frames without graph
marks. Traverse backward from the final property and cared next-state inputs,
supporting both all-controlling and first-controlling choices. Add explicit
buffer, XOR and mux sensitization for Go graphs. Verify returned masks through
conservative ternary simulation. Determine essential PIs through exact
single-bit perturbation of the expanded trace, retain cared state columns, run
the second care pass with essentials fixed and independently verify the final
mask. Return the full pipeline as one owned analysis value.

Return reverse care-node counts instead of printing averages. Replace the two
hardcoded MNIST file helpers with a bounded parser for caller-provided standard
IDX image bytes, retaining little-bit-order pixel expansion. Check object,
mapping, frame and bit budgets before allocation and poll cancellation through
every traversal. Preserve caller graphs and counterexamples.

Test sequential state/implication propagation, initialized unrolling, all
three AND justification recurrences, native XOR/mux care, insufficient masks,
formatting/statistics, malformed IDX data, every resource class, cancellation
and a 10,000-node iterative cone. Exhaust every satisfying assignment of all
255 nonzero two-input truth tables and compare essential bits with direct
flips. Fuzz another 5,000 cases and exhaust every assignment admitted by care
and minimized masks. The final full suite passed in 15.779s, the race suite in
81.611s, and vet passed. This promotes `bmcCexTools.c` from unstarted to
validated; inventory is now 246 validated, 39 partial and 1,472 unstarted
files.


## BMC unate counterexample-annotation checkpoint

Complete all three active definitions and packed annotation helpers in
`src/sat/bmc/bmcCexMin2.c`. Replace graph-owned two-bit `vTruths` storage with
an owned per-frame/per-object annotation containing the concrete value and
justification membership. Simulate the supplied initialized trace, verify its
selected final PO, then traverse backward through COs, ordinary ANDs and
RI-to-RO boundaries. Preserve deterministic fanin-zero selection when both
controlling inputs are zero, and preserve the option to justify both. Record
an exact private structural snapshot so stale annotations fail after mutation.

Build the same monotone unate fixedness networks in both modes. Counterexample
analysis exposes justified start-frame ROs as variables; refinement fixes the
real PI prefix and exposes later pseudo-PIs. Preserve use-all candidate CIs,
path constants, observed-one AND, observed-zero OR, single-fanin passthrough,
RI transfer, structural hashing and cleanup. Reject buffers, native XORs and
muxes because this source defines no sound recurrence for them.

Replace `Gia_ManCexMin`'s print-only loop, optional `temp.aig` overwrite and
unconditional `NULL` return with owned frame indexes, annotations and networks.
Add checked object, mapping, frame and bit limits plus synchronous cancellation.
Test exact sequential annotations, both modes, use-all behavior, first-zero and
maximum-zero policies, provenance rejection, source ownership, every resource
class and a 10,000-node cone. Fuzz random two-input functions and independently
enumerate every completion allowed by an accepted fixedness set; 5,000
executions passed. The final full suite passed in 15.855s, the race suite in
82.361s, and vet passed. This promotes `bmcCexMin2.c` from unstarted to
validated; inventory is now 247 validated, 39 partial and 1,471 unstarted
files.


## AIG priority-selected counterexample-reason checkpoint

Complete all 13 active definitions in `src/sat/bmc/bmcCexMin1.c` with concrete
owned AIG, trace, reason and graph results. Collect each frame's CIs by starting
at the selected final PO, then following only next-frame ROs through their
matching previous-frame RIs. Preserve root order, fanin-zero-first DFS order
and uniqueness without traversal IDs or recursion.

Derive phase/priority tables in ascending frame order. Preserve PI discovery
order, inherited LI-to-LO priority, true-AND minimum priority, the sole
controlling-zero priority and double-zero maximum priority. Represent both
source priority-band variants explicitly: registers below frame PIs in the
main path and registers above all PIs in the diagnostic alternative. Support
actual nonzero initial register values. Verify the selected final PO from the
derived tables rather than relying on debug-only phase scratch.

Collect ordered frame-local PI or register reason literals by following the
selected controlling fanin, both true fanins, and the higher-priority fanin on
a double zero. Retain exact structural provenance. Duplicate the source AIG
with one PO per post-zero-frame register cube followed by all original RIs;
preserve name and register count, support buffers, reject source constraints,
and clean dangling logic. Replace the source driver that only prints, writes
`aigcube.aig` and returns `NULL` with an owned reason analysis and cube graph,
without claiming to have generated a reduced counterexample.

Bound objects, traversal events, frames, entries, literals and output growth,
and poll cancellation throughout. Test exact sequential terms and inherited
priorities, every controlling-zero case, both priority bands, PI/state reason
modes, positive/negative cube outputs, nonzero initial states, buffers,
provenance, all limits and a 10,000-node iterative cone. Fuzz random two-input
functions and enumerate every assignment compatible with each selected reason;
5,000 executions passed. The final full suite passed in 15.623s, the race suite
in 82.022s, and vet passed. This promotes `bmcCexMin1.c` from unstarted to
validated; inventory is now 248 validated, 39 partial and 1,470 unstarted
files.


## MSAT memory-manager checkpoint

Port all 15 active definitions in `src/sat/msat/msatMem.c` as concrete owned
fixed, flexible and hierarchical step arenas. Do not expose unsafe pointers or
an allocator interface. Return stable concrete handles with bounded writable
byte views, and invalidate them deterministically after recycle, restart or
close. Preserve the fixed manager's 1,024-entry versus 64-KiB chunk selection,
minimum eight-entry chunks, LIFO free list, peak counters, first-chunk restart
and verbose diagnostic content. Preserve the MSAT flexible manager's distinct
4-KiB initial chunk, monotonic packing and permanent two-times-request growth.

For step allocation, preserve 8-byte power-of-two classes, the `4 << steps`
map boundary, zero-size nil result, class-only memory-usage total and direct
oversized allocation. Oversized recycle must invalidate and release the handle
immediately, matching this source rather than retaining it like a region arena.
Add checked byte/block limits and exact size/owner validation around behavior
that relied on assertions and correctly repeated size arguments in C.

Test every chunk-size boundary, cross-chunk growth, LIFO identity, restart
order and stale handles; flexible zero/edge/large packing; every step-class
boundary; immediate large release; diagnostics, counters, resource limits,
foreign/double/wrong-size recycle and lifecycle. Fuzz fixed fetch/recycle/
restart sequences against an independent state model, flexible packing against
an independent chunk-offset recurrence, and step routing against the minimal
power-of-two class rule. Each campaign passed 5,000 Go-only executions, the
complete suite in 15.482s, race in 81.596s and vet, promoting `msatMem.c` to
validated without C source, runtime oracles or retained generated fixtures.


## MSAT concrete solver core, I/O and reader checkpoint

Introduce one concrete `MsatSolver` over the validated owned Go CDCL engine;
do not create a solver interface or duplicate a second untested search engine.
First cover the complete behavioral surfaces of `msatSolverCore.c`,
`msatSolverIo.c` and `msatRead.c`, while moving `msatSolverApi.c` to partial
until its search-state arrays and Prepare operation can be integrated with
`msatSolverSearch.c`. Preserve MSAT's -1/0/1 statuses, literal-encoded models,
variable growth, stored-clause numbering, original-clause field identity,
assumptions, root simplification, progress and cumulative statistics.

Use exact checked conflict and wall-time budgets instead of allowing the C
restart loop to overshoot them. Define progress on zero variables as zero rather
than propagating a division NaN. Return copied assignments, models, conflicts,
factors and clause lists; validate configurations, capacities, literals and
lifecycle atomically. Keep the incomplete API ledger explicit: decision levels,
reasons, watches, learned views, memory, seen IDs, allocation counters,
adjacency/cone/used-variable state and Prepare remain coupled to search.

Replace stdout assignment/clause/activity/statistics printing with exact owned
diagnostic strings. Serialize original and learned clauses plus level-zero
assertions, allow timestamp injection for deterministic tests, and complete
serialization before touching a file. Correct the inconsistent-database writer
edge case by emitting an empty clause, preventing UNSAT from becoming SAT.
Fold DIMACS tokenization over the validated strict bounded CNF parser, retaining
MSAT solver construction and root simplification while replacing `exit()` and
unchecked whole-file allocation with errors and byte/line/object limits.

Test construction, zero-capacity growth, resize/clean/close, normalization,
numbering and Type-A marking, factor ownership, root progress, SAT, UNSAT,
assumption conflicts, copied results, all six truth patterns, exact diagnostics,
timestamped DIMACS, file I/O, parse round trips and malformed inputs. Fuzz up to
four-variable formulas against exhaustive truth enumeration and reparse/re-solve
every generated DIMACS stream. The model passed 5,008 executions, the complete
suite in 15.740s, race in 82.244s and vet. This promotes `msatSolverCore.c`,
`msatSolverIo.c` and `msatRead.c` to validated and `msatSolverApi.c` to partial,
without C source, runtime oracles or retained generated fixtures.


## BMC counterexample-depth target checkpoint

Complete all six active definitions in `src/sat/bmc/bmcCexDepth.c` with
concrete owned GIA and counterexample values. For target enlargement, preserve
the reversed frame-block PI convention, arbitrary starting register state,
frame-to-frame RI/RO transfer, PO-zero selection, structural hashing and
cleanup. Semantically copy supported extended gates instead of treating them
as ordinary ANDs. Quantify every bounded PI, retain only register variables,
and append the resulting reachability predicate as a new property. Return the
owned graph instead of printing statistics or writing `miter3.aig`.

Consolidate the two depth-network routines behind an explicit pruning option.
Use the expanded PI/RO trace from the validated counterexample tools, create a
polarity-adjusted variable for each register at the selected start frame, fix
the later primary inputs, and preserve the observed-value recurrence: AND for
two true fanins, OR for two false fanins, and the sole controlling formula for
a mixed pair. Carry formulas and impliedness through RIs to the next-frame ROs.
Reject buffers, XORs and muxes because this source supplies no sound depth-cone
recurrence for them. Build the last requested nonzero-start cones in descending
order and append them as register-only properties. Compose this operation with
the verified care analysis and return the analysis, starts, cones and graph;
do not retain the `miter2.aig` side effect.

Bound objects, mappings, frames, bits and cone count and poll cancellation.
Test exact reversed-frame and existential semantics, both pruning variants,
descending selection, metadata and source ownership, extended-gate boundaries,
invalid traces, every resource class and a 10,000-gate iterative cone. Fuzz
small sequential truth tables and independently replay every register witness
accepted by each cone. A 5,000-execution Go-only campaign passed, as did the
complete suite in 15.758s, race in 81.838s and vet. This promotes
`bmcCexDepth.c` from unstarted to validated; inventory is now 249 validated,
39 partial and 1,469 unstarted files.


## BMC counterexample bad-state-cut checkpoint

Complete all eight active definitions in `src/sat/bmc/bmcCexCut.c` using
concrete owned annotations and graphs. Consolidate the recursive and
nonrecursive justification routines into an iterative frame/object traversal.
Preserve concrete values, variable initial registers, fixed primary inputs and
constants, fixedness recurrence, deterministic first-fanin choice when both
AND inputs control zero, CO propagation and RI-to-RO transfer. Store structural
provenance and do not borrow graph marks or scratch values. Reject buffers,
XORs and muxes because the source supplies only an ordinary-AND recurrence.

Represent the target-state driver with an explicit concrete mode: a cube of
justified register values at the end frame, one monotone cone from the end
frame through the failure, or a multi-output collection for every terminal
frame through the failure. Return the begin-frame state, ordered frames and
owned combinational property. When a sequential result is requested, replace
the transition system's original properties and phase-recode the begin state
to zero using the validated concrete GIA operations. Convert AIG input and
output through the validated concrete conversion path; omit verbose printing.

Bound object, mapping, frame and bit work and poll cancellation. Test exact
annotation rows, deterministic double-zero choice, all construction modes,
multi-output ordering, sequential recoding, AIG conversion, fully implied
terminal failures, ownership, invalid ranges, every resource class and a
10,000-gate iterative cone. Fuzz small sequential truth tables and independently
replay every state accepted by cube and terminal-cone results. A 5,000-execution
Go-only campaign passed, as did the full suite in 16.003s, race in 81.717s and
vet. This promotes `bmcCexCut.c` from unstarted to validated; inventory is now
250 validated, 39 partial and 1,468 unstarted files.


## BMC priority counterexample-care checkpoint

Complete all ten active definitions in `src/sat/bmc/bmcCexCare.c` with concrete
owned priorities, care masks and adapter results. Preserve all four active
PI/PPI orderings, forward frame/register transfer and the source priority rule:
minimum for a true AND, the controlling-zero input for a mixed AND and maximum
for two zeros. In reverse order preserve CO roots, marked RI-to-RO propagation,
both true fanins, the sole controlling zero, PI preference on double zero,
already-marked suppression, greater-priority selection and the second-fanin
tie. Select the first sparsest verified candidate. Support actual nonzero
initial states and avoid graph scratch.

Extend PI care to per-object knownness with bounded ternary simulation and
return owned intersections. Implement AIG entry points through concrete GIA
conversion and return Boolean verification rather than stdout. The declared
`Bmc_CexCareSatBasedMinimizeAig` dependency has no definition in the surveyed
source tree; replace its wrapper contract explicitly with a deterministic
implication-aware low-effort pass and the validated care/essential/minimized
high-effort pipeline, verifying both. Reject extended gates only in the
ordinary-AND priority algorithm; the replacement care pipeline retains its
validated semantic extended-gate handling.

Bound object, mapping, frame and bit work and poll cancellation. Test every
gate recurrence and ordering, deterministic double-zero choice, nonzero initial
state, object knownness, mask ownership, AIG adapters, both replacement effort
modes, invalid dimensions, every resource class and a 10,000-gate cone. Fuzz
small sequential truth tables and exhaustively enumerate every completion
allowed by each care mask. A 5,000-execution Go-only campaign passed, as did the
full suite in 15.620s, race in 81.920s and vet. This promotes `bmcCexCare.c`
from unstarted to validated; inventory is now 251 validated, 39 partial and
1,467 unstarted files.


## BMC incremental initial-state-equivalence checkpoint

Complete all seven active definitions in `src/sat/bmc/bmcBmci.c` with concrete
owned unrolling and SAT results. Preserve constant versus independent symbolic
initial-register literals, shared PI blocks for paired trajectories,
ordinary-AND hashing, RI transfer, per-register XOR queries, the source's
preferred-side UNSAT merge and fixed-point rule. Return the cumulative graph,
final paired states, frame statistics, aggregate SAT work and exact unknown
reason; do not retain print-only progress or the wrapper's unconditional one.

Consolidate `Cnf_DeriveGiaRemapped`, `Cnf_DataLiftGia`, and the recursive part
extraction into checked cumulative-graph conversion and Tseitin encoding per
query. This intentionally drops incremental clause reuse but must preserve the
same Boolean assumptions and frame/register merge sequence. Bound frames,
objects, mappings, variables, clauses and literals; retain logical SAT budgets
and synchronous cancellation instead of relying only on wall-clock timeout.
Reject buffers, XORs and muxes because the source unfolds them as ordinary ANDs.

Test absorption-based logical equivalence that structural hashing cannot see,
persistent differences, immediate shared-input convergence, symbolic and
constant states, every resource class, SAT unknown reporting, cancellation and
a 10,000-gate iterative cone. Fuzz small sequential truth tables and enumerate
both initial values and all shared input sequences independently. A
5,000-execution Go-only campaign passed, as did the full suite in 15.806s, race
in 82.334s and vet. This promotes `bmcBmci.c` from unstarted to validated;
inventory is now 252 validated, 39 partial and 1,466 unstarted files.


## Incremental GIA bounded-checking checkpoint

Complete all 13 active definitions in `src/sat/bmc/bmcBmcG.c` by consolidating
their behavior over the validated concrete Go unroller, AIG conversion,
Tseitin CNF and CDCL solver. Preserve zero initial registers, frame-major
unfolding, property order, environment constraints, first-SAT termination,
model-to-trace decoding, witness replay, finite-bound versus unknown status,
logical budgets, cancellation and frame/property completion callbacks.

Represent manager start/stop, recursive unfolding/collection, frame-to-SAT
mapping, incremental CNF addition and the one-process driver through owned Go
values. Treat frame batching and eliminated-variable CNF as performance
choices; do not create another solver hierarchy or interface. Replace print
helpers with owned result/statistics/callback values. Reject buffers, XORs and
muxes because the source path would interpret them as ordinary ANDs.

Test the first witness and frame, trace replay, exact callback order, finite
safe bounds, constraints, selected property order, ownership, every resource
class, cancellation and a 10,000-gate cone. Fuzz small sequential truth tables
and independently enumerate every input sequence. A 5,000-execution Go-only
campaign passed, as did the full suite in 15.948s, race in 82.272s and vet.
This promotes `bmcBmcG.c` from unstarted to validated; inventory is now 253
validated, 39 partial and 1,465 unstarted files.


## M-inductiveness checkpoint

Complete all eight active definitions in `src/sat/bmc/bmcICheck.c` with a
concrete owned self-composition and selected-register result. For exact M,
start the two copies in arbitrary independent states, share primary inputs,
require all primary outputs and selected next-state functions to agree for M
transitions, and ask whether an output or selected next-state function can
differ on transition M+1. Preserve SAT, UNSAT and unknown distinctly.

Retain both source minimizers. The core mode starts with every selector positive
or negative and repeatedly removes positive assumptions absent from the final
UNSAT conflict. The search mode carries its selected set through M=1..Frames
and greedily attempts each removal. Preserve natural, reverse,
backward-topological and reverse-backward ordering. Backward traversal must
start at POs, visit fanin zero before fanin one, continue through the RIs of
newly found ROs, and append dangling registers naturally.

Consolidate remapped-CNF derivation, literal lifting and solver assembly into
checked GIA-to-AIG conversion, reference Tseitin encoding and the concrete CDCL
solver. This intentionally replaces mutable CNF offsets while preserving the
Boolean selector query. Return owned selected sets, exact status/reason/frame
and aggregate statistics instead of printing. Bound graph, mapping, variable,
clause and literal work, poll cancellation, and reject extended gates that the
source would misread as ordinary ANDs.

Test exact M semantics against an independent paired-state enumerator, both
core starts, order-sensitive natural/reverse/backward searches, ownership,
source immutability, every resource class, SAT unknowns, cancellation and a
10,000-gate cone. Fuzz generated output and transition truth tables while
enumerating every initial-state pair and shared input sequence. A 5,000-run
Go-only campaign passed, as did the full suite in 15.958s, race in 83.596s and
vet. This promotes `bmcICheck.c` from unstarted to validated; inventory is now
254 validated, 39 partial and 1,464 unstarted files.


## Interval BMC and ternary-prefix checkpoint

Complete all 25 active definitions in `src/sat/bmc/bmcBmc2.c`; exclude the
block-commented `Saig_BmcIntervalExplore_rec`. Preserve two useful operations.
First, simulate from zero registers with unknown PIs using the exact
three-valued complement/AND recurrence, carry RIs to ROs, retain frame/object
values and stop at the first PO not proved zero. Return owned eager values
rather than exposing zero/unvisited storage ambiguity.

Second, solve bounded targets in frame-major then property-major order. Honor
the start frame, treat every PO as a target, apply per-target and cumulative
conflict limits, add an entailed negative unit after each UNSAT target, stop at
the first SAT target, map it to the exact frame/output and replay the decoded
zero-initialized witness. Do not reinterpret AIGER constraint metadata because
the source does not distinguish it in this path.

Represent frame lookup/set, child mapping, manager lifetime, recursive
construction, interval extraction, stitched variable CNF loading and target
solving through the validated owned unroller, AIG conversion, Tseitin relation
and concrete CDCL solver. Keep node batching as a performance hint for the
single bounded arena. Do not duplicate the Satoko backend, which has no
separate Boolean contract. Replace debug target-PO mutation and printing with
the materialized frame-major relation, structured results and callbacks.

Bound frames, objects, mappings, variables, clauses, literals and ternary
entries; reserve clause/literal capacity for every possible UNSAT unit; use
logical SAT budgets and synchronous cancellation. Test ternary stopping and
completion, target order and skips, constraint-marked POs, blocking, replay,
callbacks, ownership, source immutability, all limits and a 10,000-gate cone.
Fuzz both exhaustive concrete traces and independent possible-value-set
structural ternary simulation. Both 5,000-run Go-only campaigns passed, as did
the full suite in 15.680s, race in 82.827s and vet. This promotes `bmcBmc2.c`
from unstarted to validated; inventory is now 255 validated, 39 partial and
1,463 unstarted files.


## Tent/rank and all-output BMC checkpoint

Complete all 22 unique active symbols and 23 conditional bodies in
`src/sat/bmc/bmcBmcS.c`; count the two mutually exclusive
`Bmcs_ManPerformMulti` bodies as one symbol. Preserve PO-rooted tent traversal,
RI-to-RO expansion, first-tent assignment, cross-tent candidate marking,
constant-zero candidate clearing, deterministic boundary order and rank-delay
profiles. Return owned object-indexed arrays and slices rather than graph marks,
Value scratch and parallel vectors.

Represent `Bmc_SuperBuildTents` through the owned schedule and ranked unroller:
use a private checked source snapshot, circular rank slots, atomic frame append
and semantic snapshots instead of emitting a temporary retimed graph/map.
Require exhaustive equivalence with ordinary unrolling. Reject extended gates
in the source-compatible analysis because the C traversal treats them as
ordinary ANDs.

Preserve the bounded checker’s zero initial state, frame-major/property-major
order, every-PO targets, finite bound, first SAT target, exact frame/output,
definitive-only callbacks and replayed counterexample. Consolidate manager
lifetime, per-frame copies, recursive unfold/collect, clean-CNF extraction and
solver loading into the validated interval unroller, GIA/AIG conversion,
Tseitin relation and concrete CDCL solver. Keep frame batch and worker count as
performance options; use deterministic solving instead of retaining pthread
worker state or a nondeterministic portfolio. Replace printing with structured
status and statistics.

Test exact tent/candidate metadata, constant zero, ownership, source
immutability, all-PO ordering, one/four-worker equality, batching, callback
order including unknown suppression, replay, all limits, cancellation and a
10,000-gate cone. Fuzz tent analysis against an independent DFS and BMC against
exhaustive sequential truth tables. Both 5,000-run Go-only campaigns passed,
as did the full suite in 15.856s, race in 83.338s and vet. This promotes
`bmcBmcS.c` from unstarted to validated; inventory is now 256 validated, 39
partial and 1,462 unstarted files.


## Ternary-guided sparse AND BMC checkpoint

Complete all 27 active definitions in `src/sat/bmc/bmcBmcAnd.c`. Preserve the
zero-register/unknown-PI ternary recurrence, RI-to-RO state transfer,
first-unknown frame, explicit and adaptive stopping, reverse selection through
only unknown controlling fanins, register-input leaf propagation, early
backward cutoff, forward sparse construction, frame-major POs and frame/PI
provenance. Use owned eager values and iterative walks instead of packed words,
recursive traversal, graph marks, phases or `Value` scratch.

Consolidate cone duplication, variable assignment, old/new CNF addition,
manager lifecycle, output scans, counterexample generation and the unroller,
old-CNF and internal drivers over the concrete GIA-to-AIG converter, Tseitin
relation and CDCL solver. Preserve zero state, frame/property query order,
first-SAT termination, finite exclusion versus unknown and exact witness replay.
Ignore future sparse PIs when translating a shorter early-frame witness.
Reject known-one POs because the source sparse builder unsoundly maps them to
zero; reject constraint metadata and extended gates rather than silently
changing the Boolean contract. Replace printing, frame dumps, runtime globals
and retry mutation with structured results, checked caps and cancellation.

Test exact ternary frames, adaptive bounds, cutoff behavior, known controlling
values, provenance ownership, source immutability, replay, finite exclusion,
all error surfaces and a 10,000-gate cone. Fuzz arbitrary one-register output
and transition truth tables against exhaustive input-sequence reachability. A
5,000-execution campaign passed, as did the full suite in 16.155s, race in
82.737s and vet. This promotes `bmcBmcAnd.c` from unstarted to validated;
inventory is now 257 validated, 39 partial and 1,461 unstarted files.


## Scalable solve-all BMC checkpoint

Complete all 46 active definitions in `src/sat/bmc/bmcBmc3.c`. Preserve packed
ternary complement/AND semantics, zero registers, unknown PIs, RI/RO transfer,
both register-convergence and first-unknown-output stopping policies, and
cross-frame unknown-cone counts. Replace destructive packed-word clearing with
owned eager rows and iterative `(frame, object)` visitation.

Preserve PO-rooted sequential DFS sections, LI-to-LO layer expansion,
complemented-output supergate collection, regular-ID uniqueness, reference
counts and increasing reference/ID sorting. Reuse the exact DAR four-input
mapping array and SOP database for truth cofactoring/reduction, mapping
references, mapped CNF and deterministic BLIF formatting rather than adding a
second cut mapper.

Consolidate manager allocation, per-frame literal tables, recursive and
iterative CNF construction, duplicate-cut hashing, three solver backends,
literal queries and counterexample generation into the concrete owned unroller,
Tseitin relation and CDCL solver. Preserve zero state, start frames,
frame/property order, solve-all skipping, same-model output discovery, entailed
UNSAT units, jump-on-unknown limits, finite-state completion, callbacks and
replayed witnesses. Use logical budgets and synchronous cancellation instead
of wall clocks. Explicitly retain dropped outputs as unknown rather than
reproducing the source timeout path's unsafe success answer. Do not invent
behavior for `fDropSatOuts` or `nPisAbstract`, which are initialized but never
read in this translation unit.

Test every analysis, mapping, formatting and solver path, ownership, all limits,
cancellation and 10,000-node cones. Fuzz ternary rows against an independent
possible-set evaluator, truth reduction against cofactors and solve-all results
against exhaustive sequential reachability. All three 5,000-execution campaigns
passed, as did the full suite in 16.203s, race in 82.921s and vet. This promotes
`bmcBmc3.c` from unstarted to validated; inventory is now 258 validated, 39
partial and 1,460 unstarted files.


## BMC public-header checkpoint

Reconcile every public declaration in `src/sat/bmc/bmc.h` after the underlying
BMC implementation files are validated. Represent `Bmc_EsPar_t`,
`Saig_ParBmc_t`, `Bmc_AndPar_t`, `Bmc_BCorePar_t`, `Bmc_MulPar_t` and
`Bmc_ParFf_t` as concrete Go parameter values, preserving every field without
introducing an options interface. Pin the exact inline synthesis default and
the distinct command defaults used for ordinary AND BMC, super BMC, bounded
core and multi-output proving. Preserve callbacks as concrete function fields.

Map the opaque `Unr_Man_t` lifetime to owned `GiaUnrolling` state, and audit all
22 extern declarations against the validated concrete APIs from `bmcBCore.c`,
`bmcBmc.c`, `bmcBmc2.c`, `bmcBmc3.c`, `bmcBmcAnd.c`, `bmcCexCare.c`,
`bmcCexCut.c`, `bmcCexMin1.c`, `bmcCexTools.c`, `bmcICheck.c` and
`bmcUnroll.c`. Keep backend selectors, wall-clock controls and diagnostic file
fields as configuration data where necessary; solving continues to use logical
budgets and owned results. Do not let header validation overclaim the separately
tracked exact-synthesis or fault algorithms.

Stand-alone tests pin all nonzero/sentinel defaults, zero fields, constructor
independence and callback types. The complete suite passes in 15.862s, race in
84.004s and vet is clean. This promotes `bmc.h` from unstarted to validated;
inventory is now 259 validated, 39 partial and 1,459 unstarted files.


## SAT-guided SOP-collapse checkpoint

Complete all 18 active definitions in `src/sat/bmc/bmcClp.c` as two concrete,
owned capabilities. Replace the packed single-pattern perturbation and random
stdout experiment with exact `OutputSensitivity` results and caller-supplied
`AnalyzeOutputSensitivity` pattern statistics. Preserve baseline-output,
single-CI derivative, output/input partition and co-derivative meanings while
removing the source's 63-input packed shift and fixed 64-variable arrays.

Consolidate the multiple old/experimental solver entry points into
`CollapseGiaOutputToSOP` over the validated GIA-to-AIG conversion, Tseitin
relation and CDCL solver. Preserve alternating on/off-set enumeration,
first-completed phase selection, canonical lexicographic and reverse CI order,
the canonical flipped-region guard, reverse literal-removal order,
complement-set expansion checks, cube blocking, constant covers and direct
full irredundancy. Treat quick conflict-core reduction and the secondary
expander as performance variants of the same prime-implicant operation rather
than exposing pointer-mutating solver APIs.

Return owned covers, tri-state status and cumulative statistics. Bound every
object, mapping, variable, clause, literal, cube, sample and evaluation surface;
use synchronous cancellation and logical SAT work instead of wall clocks,
global profile counters and printing. Tests cover exact derivatives and tables,
both phases, constants, every ordering mode, determinism, ownership,
irredundancy, all limits and a 10,000-gate raw chain. Two Go-only fuzz campaigns
passed 5,000 executions against exhaustive three-input truth tables. The full
suite passes in 16.152s, race in 83.993s and vet is clean. This promotes
`bmcClp.c` from unstarted to validated; inventory is now 260 validated, 39
partial and 1,458 unstarted files.


## FX sharing-extraction checkpoint

Complete all 20 active definitions in `src/sat/bmc/bmcFx.c`. Represent divisor
type, up to three signed cover-column literals and accumulated cost as concrete
owned values. Preserve sorted cube difference, the two-residual cap,
shared-base cost, every within-cube AND pair, OR as an inverted-input AND,
regular-input XOR and normalized mux control/branches. Preserve insertion
identity and stable decreasing-cost ranking, selecting only costs greater than
one. Use a lookup-only map without iterating it; deterministic slices replace
the native chained hash table and stdout reports.

Implement `Bmc_FxSolve` as onset-only enumeration over correlated candidate
signals, preserving forward greedy assumption removal, final conflict-core
shrinking, cube blocking and tri-state SAT results. Implement
`Bmc_FxAddClauses` with whole-batch validation and exact AND/XOR/mux clauses,
assigning variables in reverse selected-divisor order. Preserve
`Bmc_FxComputeOne` as repeated cover/discover/add iterations in which every
prior derived signal remains available. Return the diagnostic multi-output
onset/offset trials as owned covers and count tables instead of printing.

Expose candidate and per-iteration evaluators so correlated covers can be
replayed from original inputs. Test every gate truth table, classifier, exact
cost and tie rule, iteration, output analysis, ownership, malformed input, all
bounds, cancellation and a 10,000-gate raw chain. Two Go-only fuzz campaigns
passed 5,000 executions: deterministic divisor ranking and exhaustive replay of
every intermediate cover for arbitrary three-input functions. The complete
suite passes in 16.409s, race in 84.430s and vet is clean. This promotes
`bmcFx.c` from partial to validated; inventory is now 261 validated, 39 partial
and 1,457 unstarted files.


## Fault-cardinality checkpoint

Begin `src/sat/bmc/bmcFault.c` with its independent parameter/cardinality
surface. Implement its recursive pairwise sorting network, general odd/even
transposition network and tree at-most-one encoding as one concrete operation
over `SatSolver`. Preserve exact OR/AND/buffer relations, descending thresholds,
at-most versus exact constraints and deterministic variable allocation. Define
safe `k=0` and `k=n` behavior instead of retaining source assertions. Validate
all input variables and make construction atomic across dimensions, variable/
clause/literal limits and cancellation; retain solver bookmarks and return
owned statistics rather than mutating caller vectors into unusable storage.

Use no C oracle. Exhaust every input assignment through eight variables for
both sorting networks and through nine for the tree against an independent Go
popcount model. Test pre-existing formulas, contradiction, rollback,
determinism, ownership and failures, then run 5,000 fuzz executions. This ports
`Gia_ParFfSetDefault` and the eight `Cnf_AddCardin*` definitions, moving
`bmcFault.c` from unstarted to partial. Formula parsing/realization, CNF
remapping/lifting, fault unfold/cofactor transforms, test persistence and the
fault campaign remain explicit completion gates. Inventory becomes 261
validated, 40 partial and 1,456 unstarted files. The complete suite passes in
16.497s, the race suite in 85.571s, and vet is clean.


## Fault-formula checkpoint

Continue `bmcFault.c` with its formula derivation, validation/transformation,
token traversal, diagnostic rendering, GIA realization and formula-unfold
definitions. Replace pointer ranges and in-place case conversion with a concrete
postorder AST. Preserve two functional variables and the declared eight
parameters, correcting the source validator's accidental four-parameter range.
Reject missing, ambiguous and trailing syntax instead of allowing the recursive
realizer to ignore it. Support grouped negation so formulas derived from
complemented internal nodes remain semantically valid.

Use the same AST for exhaustive Boolean evaluation and GIA realization.
Structural derivation must cover constants, complements, ordinary AND, native
XOR and mux under explicit byte/node/cancellation bounds. Formula unfold applies
only to ordinary AND source gates, either globally or at direct register-input
drivers; reject extended gates rather than misinterpreting them. Preallocate
parameter PIs before ROs so the resulting sequential graph retains its PI/RO
partition, metadata and exact source-gate/parameter correspondence.

Validate entirely in Go: exhaustive truth tables for parsing and realization,
derived-formula replay against GIA evaluation, full and register-only unfold
replay, ownership, atomic failure and resource tests, plus 5,000 fuzz
executions against an independent Boolean model. This accounts for ten more
active definitions. CNF remapping/lifting, the four remaining fault unfolds,
fault cofactoring, persistence and the campaign remain required before
`bmcFault.c` can become validated. The complete suite passes in 16.478s, the
race suite in 85.772s, and vet is clean.


## Structural fault-transform checkpoint

Port `Gia_ManFaultUnfold`, `Gia_ManStuckAtUnfold`, `Gia_ManFlipUnfold`,
`Gia_ManFOFUnfold` and `Gia_ManFaultCofactor` as concrete owned graph
transformations. Preserve source gate order and every allocated control slot,
including inactive mapping entries. Implement stuck-zero/stuck-one precedence,
conditional output inversion and the four-entry two-input truth-table mux tree.
Correct the source's last-entry off-by-one condition so all four entries of the
final gate can be selected, and pin that difference with a semantic test.

For single-frame sequential transforms, allocate controls between ordinary PIs
and ROs before rebuilding gates so register metadata remains meaningful.
Input cofactoring retains CI shape while replacing a checked PI prefix with
constants. The temporal transform explicitly constructs frame-zero RO/PI
values, frame-one PIs and either per-gate or per-RI freeze controls, returning a
combinational two-frame relation as the source does. Reject native XOR/mux/
buffer gates rather than silently treating their storage as ordinary AND.

Exhaust every small input/control assignment against independent recurrence and
truth-table models, cover all four temporal modes, sequential partitions,
metadata, malformed maps, resource limits and cancellation, and run 5,000
Go-only fuzz executions. These five definitions bring `bmcFault.c` to 24
covered definitions. CNF remapping/lifting, persistence/result analysis and the
fault campaign remain required. The complete suite passes in 16.539s, the race
suite in 85.391s, and vet is clean.


## Fault-CNF checkpoint

Port `Cnf_DeriveGiaRemapped` and `Cnf_DataLiftGia` as an owned GIA relation
rather than exposing mutable C CNF arrays. Derive the full combinational
relation through the checked GIA-to-AIG lowering and reference Tseitin encoder,
then compose the lowering map with AIG object literals to obtain a literal for
every GIA object. Preserve complements, structural aliases, native XOR/mux
lowering, CI order and CO aliases; optimized cut layout is not observable in
these helpers and must not become part of their contract.

Make lifting immutable. Shift every clause, object, input and output literal,
preserve polarity, explicitly grow the variable span, and check integer/literal
overflow, resource limits and cancellation before returning a result. This
removes the source helper's reliance on a separately enlarged solver and avoids
partial mutation.

Exhaustively validate both output polarities under every input assignment on a
graph containing complements, native XOR and mux, then repeat projection checks
after a nonzero lift. Check exact shift arithmetic, object mapping, ownership,
limits and cancellation, and run 5,000 Go-only SAT-projection fuzz executions.
These two definitions bring `bmcFault.c` to 26 covered definitions;
persistence/result analysis and campaign orchestration remain required. The
complete suite passes in 16.669s, the race suite in 84.550s, and vet is clean.


## Fault-pattern and reporting checkpoint

Port the pattern dump/read, delayed-pattern expansion, result reporting and
extra-input duplication helpers without ambient files or graph scratch. Make
the durable core a pair of bounded byte codecs accepting only 0/1 and documented
ASCII whitespace; provide optional checked path wrappers for callers that
explicitly request filesystem I/O. Expand `[RO0, PI0, PI1]` rows into two
`[PI,state]` lines by replaying the owned transition relation.

Return solver/graph run statistics as a concrete record instead of appending a
hard-coded process-global file. When adding PIs to a sequential graph, rebuild
CI order as original PIs, added PIs, then ROs; preserve extended gates, outputs,
registers, constraints, names and source ownership rather than appending CIs
after the RO suffix.

Test exact formatting, whitespace/error parsing, explicit temporary-file
round trips, sequential delay expansion, duplicate truth, metadata, limits and
cancellation. Run 5,000 Go-only codec fuzz executions. These six definitions
bring `bmcFault.c` to 32 covered definitions; untestable-fault enumeration,
parameter analysis, incremental cofactor addition, new-fault comparison,
preparation and the campaign remain required. The complete suite passes in
16.377s, the race suite in 85.338s, and vet is clean.


## Reusable fault-campaign SAT checkpoint

Port parameter analysis, untestable-assignment enumeration and per-pattern
cofactor addition before assembling the top-level driver. Parameter analysis
must retain source-order positive assumptions, deactivate proven-UNSAT
parameters and skip active parameters already true in a returned model. Return
owned maps and counters; use logical SAT budgets and cancellation rather than
fixed conflict constants, timing and stdout.

Enumerate parameter projections on a cloned solver, block each complete model
deterministically, and return enabled indices plus final tri-state status rather
than mutating the campaign solver or writing a fixed file. Adding a pattern
must be atomic: cofactor the functional PI prefix, derive and lift its full CNF,
load it at the current variable boundary, assert detection-OR or all-zero
equality, and connect every remaining instance PI to its base parameter with
two equivalence clauses.

Test UNSAT pruning, same-model coverage, complete and bounded enumeration,
caller-solver isolation, every tiny pattern/detection combination, ownership,
limits, cancellation and rollback on failure. Run 5,000 Go-only incremental
instance fuzz executions. These three definitions bring `bmcFault.c` to 35
covered definitions; new-fault comparison, preparation and the top-level
campaign remain required.


## Complete fault-campaign workflow checkpoint

Finish `bmcFault.c` by composing the previously validated pieces into concrete
preparation, new-fault comparison and top-level operations. Preparation must
support all five source algorithms, build a dimension-checked golden/faulty
miter, derive and load its full relation, assert an observable difference,
apply inactive-parameter and basic/general cardinality constraints, and ingest
seed or all-zero/all-one start patterns with source-compatible warm-up and
prefix truncation. Use explicit Go path input only when requested; persistent
outputs are owned bytes and records rather than fixed process-global files.

New-fault comparison must require detection by the candidate pattern, exclude
faults already detected by every prior pattern, enumerate complete bounded
parameter projections and return enabled indices. The top-level campaign must
generate distinguishing functional patterns until UNSAT/unknown/limit,
periodically analyze parameters and rebuild from retained patterns, support
new-fault mode, check or enumerate untestable/rectifying assignments, encode
ordinary or delayed patterns and return exact run statistics. Use logical SAT
budgets and synchronous cancellation rather than wall-clock deadlines.

Test every algorithm, cardinality mode, seed/start-pattern path, analysis and
rebuild, explicit input, output encoding, iteration and assignment limit,
new/untestable result, ownership and error surface entirely in Go. Independently
exhaust the tiny stuck-at fault space and prove that every detectable allowed
assignment is covered by the generated suite; run 5,000 fuzz executions. This
accounts for `Gia_ManFaultDumpNewFaults`, `Gia_ManFaultPrepare` and
`Gia_ManFaultTest`, promotes all 38 active definitions in `bmcFault.c` to
validated, and changes inventory to 262 validated, 39 partial and 1,456
unstarted files. The focused suite, full suite, race suite and vet pass.


## Exact-synthesis semantic and topology foundation checkpoint

Begin the build-disabled but in-scope `bmcMaj.c` exact-synthesis cluster with a
shared owned representation for its majority, two-input Exa and generalized
Exa3 managers. Preserve strict-majority truth, normalized LUT truth variables,
the majority manager's constant and cascade candidate exceptions, every LUT
candidate range, fewer-variable restrictions, source SAT-variable numbering
and deterministic fanout-use lists. Keep the distinct source object domains
explicit while avoiding three duplicate pointer-heavy managers.

Represent a candidate as a concrete topological network of MAJ3 or 2-through-6
input LUT gates with explicit constants and output complementation. Provide
checked exhaustive evaluation and the exact first/middle-layer mismatch rule
needed by the later CEGAR loops. Return owned descriptions and BLIF bytes rather
than printing or writing a derived global filename. Bound complete object,
variable, assignment and text work, reject duplicate/forward fanins and wide
truths, preserve caller ownership and support synchronous cancellation.

Test exact topology layouts and variable IDs, majority truth through ten
inputs, arbitrary LUT semantics, output complementation, mismatch selection,
formatting, BLIF rows, ownership and all limits entirely in Go. Fuzz 5,000
two-level networks against an independent truth lookup. This accounts for 17
of 36 active `bmcMaj.c` definitions and moves the file from unstarted to
partial; manager/SAT allocation, topology and minterm CNF, lazy cache paths and
the three complete exact-synthesis loops remain. Inventory becomes 262
validated, 40 partial and 1,455 unstarted files.

Complete the non-lazy path by building exact-one pin selection, adjacent-pin
symmetry, optional node-order, mandatory-use and normalized two-input function
clauses over the concrete Go solver. Append each minterm atomically with pin
connection equivalences and exact MAJ3/LUT functionality, using fresh checked
evaluation variables. Decode every SAT model into the semantic network and
iterate on the deterministic mismatch until exact, UNSAT, unknown or a logical
limit. Normalize targets whose zero minterm is one and complement only the
returned output, matching Exa behavior.

Exhaustively project every normalized two-input target after each added
minterm; test majority, target complementation, `OnlyAND`, SAT/UNSAT/iteration
stops, atomic rollback and cancellation. Run separate 5,000-execution fuzz
campaigns for minterm CNF and the final CEGAR result. This brings coverage to
33 of 36 active definitions. Keep the file partial until the alternate Exa3
incremental tuple-cache operations (`Exa3_ManIsUsed2`, `Exa3_ManIsUsed3`, and
`Exa3_ManAddCnf2`) are implemented and tested rather than treating the complete
non-lazy path as a substitute.

Complete that alternate path with preallocated per-minterm/node value
variables and a lookup-only cache keyed by `(minterm,node,selected fanins)`.
Guard every lazy function clause by the current model's connection variables,
skip only previously emitted tuples, preserve tried/added counts, permit a
semantic minterm to recur when the topology changes, and make clause/cache
commit atomic. Support the source LUT2/LUT3 cache domains and return explicit
errors for other lazy arities.

Test first insertion and cache hits, fixed evaluation-variable numbering,
limits, cancellation, rollback, target complementation, SAT/UNSAT/iteration
results and exact decoded truth. Fuzz the complete lazy CEGAR loop 5,000 times
against the independent truth model. This accounts for the final three active
definitions and promotes all 36 definitions in `bmcMaj.c` to validated.
Inventory becomes 263 validated, 39 partial and 1,455 unstarted files. The
focused, full, race and vet gates pass.

## Sparse exact-synthesis topology completion checkpoint

Complete `bmcMaj2.c` by reusing the validated majority/LUT semantic, structural
CNF, minterm CNF, decode and CEGAR layers for its duplicate definitions. Add a
concrete sparse-majority connection record and a reproducible generator that
gives each non-output internal node a later consumer before inserting distinct
extra edges. Require an explicit seed, bound objects and extra-edge capacity,
poll cancellation and fail finite impossible requests; do not reproduce the
source's process-global clock-seeded random state or unbounded retry loops.

Build the sparse topology with the hardwired first MAJ3 gate, optional cascade
pin, fixed internal connections, input-only residual fanin domains, optional
third-pin constants, exact variable order and fanout-use clauses. Preserve the
dormant `Maj_ManConnect2` table as owned data rather than executable global
state. Represent Exa3's optional initial polarity as an ordered concrete list
of function/connection variables and an explicit solver operation, retaining
cyclic primary-input selection and the previous-node chain. Since the source
allocation call is commented out, keep this initialization opt-in.

Test exact sparse pin domains and variable uniqueness, deterministic same-seed
generation, mandatory consumers, distinct edges, capacity errors, fixed-table
data, cancellation, phase ordering and synthesized majority truth entirely in
Go. Replace the source's hardcoded infinite-retry test driver with a bounded
stand-alone synthesis test. Fuzz 5,000 bounded sparse topologies against
independent structural invariants. This accounts for all 36 active definitions
and promotes `bmcMaj2.c` to validated without C source or a runtime native
oracle. Inventory becomes 264 validated, 39 partial and 1,454 unstarted files.

## Level-constrained exact-majority checkpoint

Begin `bmcMaj3.c` with its Maj3 manager rather than conflating its unordered
fanin subsets with the ordered-pin engines. Represent the caller's top-down
level counts as an owned topology. Preserve reversed object allocation, the
fixed first three-input gate, fixed top fanins, one fixed cascade edge between
intermediate levels, lower-level-only selectable sources and source variable
numbering. Validate positive level counts, the single output level, the native
32-object bound and all source/variable uniqueness.

Build initial fanin and mandatory-fanout clauses over the concrete solver.
Append each minterm atomically with fixed input/target values, fresh internal
values and every selected-source pair implication for MAJ3 semantics. Decode a
model's unordered subset into an `ExactNetwork`. Preserve the source's lazy
cardinality loop: a node with fewer than three selected sources receives its
complete currently-false clause, while a node with too many receives the
source-sized prefix of currently-true literals. Iterate until all nodes have
exactly three fanins before semantic mismatch evaluation.

Test exact level/object/variable maps, fanin and fanout clauses, both lazy
directions, decode, every five-input minterm, complete `{1,2,1}` synthesis,
ownership, limits and atomic cancellation entirely in Go. Fuzz 5,000 bounded
level distributions against independent structural invariants. This accounts
for the 16 Maj3 definitions and moves `bmcMaj3.c` from unstarted to partial;
the 27 Zyx topology/function/enumeration/verifier definitions remain. Inventory
becomes 264 validated, 40 partial and 1,453 unstarted files.

## Generic subset exact-synthesis completion checkpoint

Finish the Zyx half of `bmcMaj3.c` with a concrete subset-selector topology,
not ordered pins. Preserve the three contiguous variable regions for node
functions, every node/source topology bit and every minterm/object value.
Install unit clauses for normalized function bit zero, invalid/future edges,
the top predecessor, majority's first three inputs, primary-input assignments
and final target values. Retain mandatory fanin/fanout clauses and LUT2
nondegeneracy/`OnlyAND` restrictions.

Refine SAT models lazily to exactly the requested fanin count. Once cardinality
is valid, preserve adjacent-node lexicographic topology/function ordering and
symmetric-input first-use ordering with model-blocking clauses. Incremental
mode must cache `(minterm,node,fanins)` for LUT2/LUT3 and each selected MAJ3
pair using a lookup-only map; eager mode must enumerate every bounded source
combination. Add topology-only solution blocking, concrete decode, semantic
mismatch selection, complementation and bounded multi-solution enumeration.
Expose counts and solver statistics instead of diagnostic stdout.

Replace the derived filename writer and native file tester with owned compact
solution bytes, a strict metadata-name parser and a dataset verifier that
reconstructs/evaluates every network. Store permanent vectors only under
`test_vectors/`. Bound variables, assignments, combinations, clauses, bytes
and solutions; poll cancellation and commit clause/cache changes atomically.

Test exact variable IDs and fixed values, both cardinality directions,
topology/function/symmetry ordering, majority-pair and LUT-tuple cache hits,
eager clauses, topology blocking, majority/LUT synthesis, complementation,
enumeration, text round trips, malformed data, ownership and every limit in
Go. Run separate 5,000-execution fuzz campaigns for subset synthesis and the
dataset verifier, in addition to the level-topology campaign. A permanent
compact XOR vector replaces C-generated fixtures. This accounts for the final
27 definitions and promotes all 43 active definitions in `bmcMaj3.c` to
validated without C or a runtime native oracle. Inventory becomes 265
validated, 39 partial and 1,453 unstarted files. The complete suite passes in
16.427s, the race suite in 84.628s, and vet is clean.

Reconcile `src/sat/bmc/module.make` after this cluster: its exact 32 selected
translation units are all ledgered, 31 validated and `bmcMulti.c` partial. Move
the manifest from unstarted to partial; Go package compilation replaces native
object enumeration. Do not promote it until the final selected source is
validated. Inventory becomes 265 validated, 40 partial and 1,452 unstarted.

## IF cut-composition cache execution checkpoint

Port both active definitions in `src/map/if/ifCache.c` as a concrete owned
cache. Preserve canonical DSD operand order, duplicate observations,
first-occurrence tuple IDs and per-shared-input entry/unique statistics. Use
`insdict.Dict` with an explicit deterministic record hash, and return typed
statistics instead of reproducing diagnostic stdout.

Validate canonicalization, duplicates, percentages, empty behavior, ownership,
record limits, atomic failures, cancellation and invalid receivers with Go-only
unit tests. Fuzz 5,000 bounded record streams against an independent slice
model. This promotes `ifCache.c` to validated without a generated native
fixture or runtime C oracle. Inventory becomes 266 validated, 40 partial and
1,451 unstarted files. The complete suite passes in 16.368s, the race suite in
84.776s, and vet is clean.

## IF cut decomposition execution checkpoint

Port all five active definitions in `src/map/if/ifCheck.c` over the concrete
owned AIG representation. Replace temporary object marks with an iterative
cut-boundary reachability walk. Compute fixed-leaf cofactors with owned truth
tables, retaining the source's seven-leaf bound, leaf order and zero-before-one
scan. Replace static counters and diagnostic stdout with a concrete checker and
typed structural, reconvergent-conflict and functional statistics.

Test direct and reconvergent structures, a function with no constant cofactor,
statistics, malformed boundaries, work/storage limits, atomic cancellation and
a 10,000-node path. Fuzz 5,000 bounded AIGs, comparing functional results with
exhaustive Boolean enumeration and structural results with an independent
graph walk. This promotes `ifCheck.c` to validated without a generated native
fixture or runtime C oracle. Inventory becomes 267 validated, 40 partial and
1,450 unstarted files. The complete suite passes in 16.401s, the race suite in
85.264s, and vet is clean.

## GIA support-pair experiment execution checkpoint

Port all four active definitions in `src/aig/gia/giaSupMin.c` as one owned
output-pair analysis operation. Consolidate Start/Stop into Go lifetime because
the allocated truth workspaces are unused in this snapshot. Replace diagnostic
printing with typed records while preserving pair order, complemented-driver
regularization, AND-like care validation, continuation after invalid care, and
the data/care support and cone cardinalities.

Test exact reports, invalid care, source immutability, output bounds,
pair/object limits, atomic cancellation and a 10,001-gate iterative cone. Fuzz
5,000 bounded graphs against an independent traversal model. This promotes
`giaSupMin.c` to validated without a generated native fixture or runtime C
oracle. Inventory becomes 268 validated, 40 partial and 1,449 unstarted files.
The complete suite passes in 16.576s, the race suite in 84.957s, and vet is
clean.

## GIA LUT packing execution checkpoint

Port all three active definitions in `src/aig/gia/giaPack.c` over owned LUT
mapping and packing records. Preserve output-rooted BFS/reversal separately
from all-LUT CO-distance propagation, including dangling LUTs only in the
latter. Use the validated exact `Abc_MergeSortCost` permutation so equal-cost
ordering retains source behavior.

Preserve the packing algorithm's 32-fanin bound, descending selection-sort
swaps, strict latest-input condition, block capacities, direct/routed delay
recurrence and reverse block fill. Return typed order, delay, block and packing
results instead of graph scratch mutation and stdout. Generalize the shared
packing representation for caller-selected wider blocks while making
three-member-only consumers reject unsupported widths explicitly.

Test reachable/dangling order, exact groups and delays, wide blocks, ownership,
malformed mappings, limits and atomic cancellation. Fuzz 5,000 mapped DAGs
against independent collection and packing models. This promotes `giaPack.c`
to validated without a generated native fixture or runtime C oracle. Inventory
becomes 269 validated, 40 partial and 1,448 unstarted files. The complete suite
passes in 16.481s, the race suite in 85.353s, and vet is clean.

## GIA technology-independent synthesis execution checkpoint

Port all five active definitions in `src/aig/gia/giaTis.c` on their intended
ordinary-AND graph domain. Preserve object-order hashing, recognized mux/XOR
construction, the source rule that retains an ordinary root when both mux
pattern children have multiple references, CI/CO and sequential metadata, and
final cleanup. Reject pre-existing native mux/XOR/buffer objects rather than
silently applying ordinary-AND semantics to their compact encodings.

Replace recursive MFFC collection and borrowed traversal IDs with an iterative
owned traversal that preserves first-encounter leaves, fanin postorder,
mux-control edges and unconditional root inclusion. Treat constants as safe
leaves. Replace vector allocation, object printing and stdout drivers with an
owned converted graph plus typed MFFC records in deterministic AND-then-CO
order; omit terminal output drivers because they have no gate MFFC and reach a
C assertion.

Test mux/XOR truth exhaustively, the doubly shared exception, metadata,
cleanup, exact report order, malformed roots, unsupported native inputs,
object/work/report bounds, atomic cancellation and a 10,001-gate iterative
cone. Fuzz 5,000 bounded ordinary graphs, comparing converted truth tables and
MFFC orders against independent Go models. This promotes `giaTis.c` to
validated without a generated native fixture or runtime C oracle. Inventory
becomes 270 validated, 40 partial and 1,447 unstarted files. The complete suite
passes in 16.441s, the race suite in 85.469s, and vet is clean.

## GIA MiniSat sorting execution checkpoint

Port all 15 definitions in `src/aig/gia/giaSort.c` into concrete owned integer
and float32 operations. Preserve the 15-entry selection-sort cutoff, middle
pivot, strict comparisons and Hoare split. For floats, apply every exact swap
to a caller-provided or generated permutation; retain the source comparison
behavior for duplicates, signed zero, infinities and NaNs rather than inventing
a total order. Use bounded work copies so cancellation never exposes a partial
sort, and do not add interfaces.

Consolidate the file-local comparators and duplicate callback/inlined integer
paths into one concrete integer operation. Replace the C-rand test generator,
sortedness assertion and fixed 100-million-entry stdout benchmark with bounded
deterministic Go unit, property and fuzz tests. Cover exact selection and
partition orders, custom permutations, arbitrary float bit patterns, malformed
dimensions, entry limits, atomic cancellation, empty slices and a 100,000-item
case. Fuzz 5,000 or more inputs against a separate literal algorithm model.

This promotes `giaSort.c` to validated without a generated native fixture or
runtime C oracle. Inventory becomes 271 validated, 40 partial and 1,446
unstarted files. The complete suite passes in 16.344s, the race suite in
85.794s, and vet is clean.

## GIA structural-unateness execution checkpoint

Port all six definitions in `src/aig/gia/giaUnate.c` as a concrete typed
analysis over selected CI/CO ordinals. Preserve the source's four-state class
encoding, bitwise meet, polarity flip, complemented fanins, output phases,
buffers and conservative reconvergence. Replace recursive traversal-ID memo
state and sparse phase-literal vectors with owned object-order dataflow. Define
nil as all terminals and explicit ordered selections, including repeats,
without interfaces. Reject native XOR/mux nodes because the source reads their
packed fields as ordinary ANDs and omits mux controls.

Return positive, negative, binate and independent matrices plus typed support,
unate and binate statistics instead of either stdout test driver. Document that
this is structural rather than exact functional unateness: reconvergent paths
of both polarities remain binate even if Boolean cancellation removes the
dependency.

Test every class, selected/repeated terminals, complemented outputs, buffers,
reconvergent cancellation, malformed IDs, ownership, object/status-cell
limits, atomic cancellation and a 10,001-gate path. Fuzz 5,000 bounded graphs
against an independent signed-path model, and exhaustively verify that every
positive, negative and independent result implies the corresponding Boolean
property. This promotes `giaUnate.c` to validated without a generated native
fixture or runtime C oracle. Inventory becomes 272 validated, 40 partial and
1,445 unstarted files. The complete suite passes in 16.558s, the race suite in
85.044s, and vet is clean.

## GIA experimental isomorphism-signature execution checkpoint

Port all 12 definitions in `src/aig/gia/giaIso3.c` as two separate concrete
owned analyses. For Iso3, preserve all fixed node, complement and edge-direction
seeds; simultaneous saved rounds; uint32 wraparound accumulation; unique-count
history; and the repeated-adjacent-count stopping rule. For Iso4, reuse the
validated reverse-level ordering and preserve the exact ABC RNG reset/discard,
per-level complement salts, CO overwrite and reverse fanin propagation from
explicit zero scratch.

Replace manager Value mutation, global random state, timing, data-vector
allocation and stdout test drivers with owned signature/level/value results and
typed per-level multiset comparisons. Reject native XOR/mux graphs because the
source experiments hash only two packed fanins and omit mux controls. State
clearly that these are fingerprints rather than a complete graph-isomorphism
decision procedure.

Test exact initial/refined signatures, convergence history, isomorphic CI
relabeling, fixed random propagation, level comparisons, malformed dimensions,
object/entry/iteration limits, atomic cancellation, ownership and a
10,001-gate graph. Fuzz 5,000 bounded graphs against independent models for
initialization, edge accumulation, reverse levels and RNG propagation. This
promotes `giaIso3.c` to validated without a generated native fixture or runtime
C oracle. Inventory becomes 273 validated, 40 partial and 1,444 unstarted
files. The complete suite passes in 16.427s, the race suite in 87.351s, and vet
is clean.

## GIA most-forward-retiming execution checkpoint

Port all six definitions in `src/aig/gia/giaRetime.c` as bounded concrete Go
operations. Preserve shared autonomous-state memoization, its deterministic
cycle break, source object-order marking and cut discovery, optional
same-register-class crossing, all-zero phase normalization, sequential
duplication, structural hashing, cleanup and repeated rounds through the first
zero-move round. Keep register classes explicit and owned; return typed cut,
phase, fixed-register and move records instead of manager-private scratch or
stdout. Reject native XOR/mux encodings rather than misreading their packed
fields as ordinary ANDs.

Use only stand-alone Go validation. Exhaustively prove one-round output and
next-state correspondence for every small old state and primary input under
the reported cut encoding. Compare multi-round zero-state traces cycle by
cycle, and cover exact cut order, class barriers, complemented phases,
sequential cycles, metadata, ownership, malformed inputs, bounds, atomic
cancellation and a 10,001-gate iterative cone. Fuzz at least 5,000 bounded
sequential graphs and rerun the exhaustive encoding proof for each case. Do
not create a native helper or retain fuzz output under `testdata/`.

This promotes `giaRetime.c` to validated without a generated native fixture or
runtime C oracle. Inventory becomes 274 validated, 40 partial and 1,443
unstarted files. The checkpoint passed 5,000 fuzz executions, the complete
suite in 16.349s, the race suite in 85.619s, and vet.

## GIA regular-expression-automata execution checkpoint

Port all 18 definitions in `src/aig/gia/giaRex.c` into concrete bounded Go
operations. Preserve byte-oriented regex preprocessing, dedicated-parenthesis
union handling, insertion/sorted alphabet modes, source state numbering,
epsilon traversal, balanced exactly-one input gating, acceptance/register
ordering, cleanup and known-one initial-state normalization. Replace transition,
reachability, simulation and walk stdout with owned metadata, acceptance
vectors and typed walk records. Use a caller seed instead of ambient C random
state and add no interfaces.

Preserve the Hacker's Delight transpose exactly, including its reversed-axis
lane convention, and expose parallel one-hot stepping plus structural initial
mask discovery with explicit 64-input/register checks. Validate regex language
semantics independently with anchored Go regexps; exhaust exactly-one logic;
model every transpose bit; compare packed stepping against scalar sequential
evaluation from arbitrary states; replay every walk; and cover ordering,
whitespace, malformed syntax, limits, cancellation and a deep iterative input.
Run at least 5,000 fuzz cases each for language and packed-step equivalence.

This promotes `giaRex.c` to validated without a generated native fixture or C
runtime oracle. Inventory becomes 275 validated, 40 partial and 1,442
unstarted files. Both fuzz campaigns passed 5,000 executions, the complete
suite passed in 16.089s, the race suite in 86.053s, and vet is clean.

## GIA solver3-integration execution checkpoint

Port all five wrapper definitions in `src/aig/gia/giaSat3.c` using the concrete
solver3 and native owned CNF structures. Preserve the aggregate output-OR
clause, independent source-order CO-cone mode, all-CI cone boundary, initial
simplification and conflict accounting. Replace status and timing stdout with
typed statuses, stop reasons, GIA-CI models, statistics and CNF dimensions;
provide explicit session ownership and no interfaces.

Use the validated reference GIA-to-AIG-to-CNF relation until the separately
tracked eight-input Mf mapper is ported. This validates the wrapper's Boolean
and lifecycle behavior without pretending that reference Tseitin clauses are
the Mf clause structure. Exhaustively compare aggregate and per-output answers,
replay SAT models, cover extended gates, selected disjunctions, contradictory
roots, construction/solve cancellation, all resource limits and a deep cone,
then fuzz at least 5,000 bounded graphs against direct truth enumeration.

This promotes `giaSat3.c` to validated while leaving the Mf mapping sources at
their honest existing dispositions. Inventory becomes 276 validated, 40
partial and 1,441 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 16.598s, the race suite in 86.302s, and vet is
clean.

## Compact AGI support-experiment execution checkpoint

Port all 37 definitions in `src/aig/gia/giaAgi.c` on the representation's
exercised ordinary-AND domain. Use concrete typed IDs, literals, objects and a
bounded manager while preserving object order, complement bits, CI/CO and
sequential terminal roles, metadata, explicit construction, conversion and
lifecycle behavior. Keep the source's dormant third-fanin extended-node path
explicitly unsupported because its allocator never creates the required
storage; do not invent safe semantics and call them source parity.

Replace recursive traversal IDs and stdout timing with iterative owned
per-node support sizes, threshold statistics and a typed comparison against
the independent GIA support-set implementation. Validate identity-preserving
conversion, all-assignment truth, exact supports, constants and CO wrappers,
sequential ordering, malformed construction, object/visit limits, cancellation
and a deep cone. Fuzz at least 5,000 bounded ordinary graphs, comparing both
truth and support sizes entirely in Go.

This promotes `giaAgi.c` to validated on its executable domain. Inventory
becomes 277 validated, 40 partial and 1,440 unstarted files. The fuzz campaign
passed 5,000 executions, the complete suite passed in 16.551s, the race suite
in 85.899s, and vet is clean.

## GIA Satoko-wrapper execution checkpoint

Port all nine definitions in `src/aig/gia/giaSatoko.c` as concrete owned Go
operations. Replace the recursive mapping traversal with iterative unique DFS
postorder over checked packed records. Preserve aggregate output-disjunction,
source-order split-cone and shared incremental-assumption modes, along with
conflict accounting and explicit solver lifecycle. Replace print/timing reports
with typed statuses, reasons, projected GIA-CI models, statistics, dimensions
and deterministic cone-variable metadata. Parse DIMACS from caller-owned bytes
with bounded variables, clauses, literals and scanner storage; reject malformed
headers, literals, unterminated clauses and count mismatches.

Use the validated standalone Tseitin relation and concrete Go CDCL solver until
the independently tracked Mf mapping and Satoko internals are ported. Validate
the wrapper's Boolean relation without claiming Satoko branching, clause-layout
or cone-scheduling parity. Accept a mapping leaf root safely, detect malformed
offsets and cycles, and provide cancellation during construction, traversal
and solving without interfaces.

Exhaustively compare every output and aggregate result with direct GIA truth,
replay every model, pin exact mapping postorder and sharing, exercise constants,
lowered extended gates, selected outputs, lifecycle, every limit and a deep
iterative cone. Cover DIMACS comments, headers, signs, multiline clauses,
empty clauses, SAT/UNSAT and malformed forms entirely in Go. Fuzz at least
5,000 bounded graphs across all three modes and retain no native helper or
`testdata/` output.

This promotes `giaSatoko.c` to validated while leaving the Mf mapper and Satoko
algorithm sources at their honest separate dispositions. Inventory becomes 278
validated, 40 partial and 1,439 unstarted files. The campaign passed 5,000 fuzz
executions, the complete suite passed in 16.241s, the race suite in 85.611s,
vet is clean, and package-header and standalone-oracle audits pass.

## GIA frontier-representation execution checkpoint

Port all five definitions in `src/aig/gia/giaFront.c` with one concrete owned
frontier representation. Compute value-reference liveness without manager
scratch, preserve source-order peak crosscut, cyclic first-free slot allocation
and the source's float32 ten-percent capacity margin. Store object kinds, slot
fanins and complement phases as typed fields, then decode slots back into exact
compact GIA object differences. Replace the print-only comparison driver with a
Boolean exact-object-and-metadata round trip.

Preserve constants, CI/CO ordering, ordinary ANDs, native two-fanin XORs,
dangling logic, object IDs, sequential partitions, constraints, Name and Spec.
Explicitly reject buffers and explicit third-edge mux controls because the C
routine unconditionally decrements two fanins for every AND-like object and
cannot encode those extensions safely. Bound objects and frontier width, make
cancellation atomic, return owned snapshots, and diagnose malformed slot
records instead of relying on assertions.

Pin exact slot reuse and width on a hand graph; compare peak liveness with an
independent future-use model; exhaust truth assignments; cover dangling logic,
sequential metadata, malformed records, every limit, cancellation, unsupported
objects and a deep iterative graph. Fuzz at least 5,000 bounded AND/XOR graphs,
checking the independent peak, exact compact words and exhaustive truth without
a native helper or retained `testdata/` output.

This promotes `giaFront.c` to validated on its safe executable domain.
Inventory becomes 279 validated, 40 partial and 1,438 unstarted files. The
campaign passed 5,000 fuzz executions, the complete suite passed in 16.495s,
the race suite in 85.866s, vet is clean, and package-header and standalone-
oracle audits pass.

## GIA custom-AIGER-extension execution checkpoint

Port all 14 definitions in `src/aig/gia/giaAigerExt.c` as bounded byte codecs
over concrete owned equivalence, LUT-mapping and packing values. Preserve the
constant equivalence class, representative/member deltas, proved bits and ABC
unsigned-varint layout. Preserve compact mapping root/fanin order and signed
delta direction, complete packed storage in the simple form, documented big-
endian counts/maximum LUT size/records, and signed big-endian packing words.
Specify the C simple codec's host-memory format as deterministic little-endian
int32 in Go, matching deployed x86 artifacts rather than retaining host
dependence.

Reject malformed length prefixes, truncated and overflowing varints, invalid
delta movement, duplicate/non-topological mapping records, incorrect documented
maxima, partial fixed words and trailing bytes. Bound objects, entries and bytes,
check cancellation throughout, clone all inputs/results and add no interfaces.
Pin exact encoded bytes with independent byte-level expectations, round-trip
empty and populated forms, exercise signed boundaries and every error surface,
then fuzz equivalence and mapping codecs for at least 5,000 cases each, including
mutated streams, entirely in Go.

This promotes `giaAigerExt.c` to validated. Inventory becomes 280 validated,
40 partial and 1,437 unstarted files. Both fuzz campaigns passed 5,000
executions, the complete suite passed in 16.134s, the race suite in 85.658s,
vet is clean, and package-header and standalone-oracle audits pass.

## GIA hierarchical-LUT-packing execution checkpoint

Port all ten definitions in `src/aig/gia/giaIff.c` as one concrete owned
session over the existing LUT mapping and timing manager. Preserve unique
source-order merged boundaries, one/two/three-LUT arrival candidates, float32
input-count delay lookup, ordered strict-improvement ties and the pair choice
when degree-three search finds no further improvement. Propagate CI/CO timing,
return per-object arrivals/matches and typed single/pair/triple/max-delay
statistics instead of printing.

Select reachable packing groups in CO order and dependency postorder, retaining
the source's absorbed-child group layout. Use an iterative work stack, cloned
inputs, explicit lifecycle, object/mapping/group/visit bounds and cancellation.
Support the command wrapper's automatic 4/6-LUT and degree-two/three library
selection without storing hidden mutable mapping, timing or packing state on
`GiaManager`, and add no interfaces.

Pin exact boundaries, timing, pair/triple ties and packing order; cover timing
ownership, auto configuration, malformed mapping dependencies, lifecycle,
every resource limit, cancellation and a deep iterative chain. Fuzz at least
5,000 bounded mapped DAGs against an independent set-based arrival/match model
and validate every returned packing entirely in Go.

This promotes `giaIff.c` to validated. Inventory becomes 281 validated, 40
partial and 1,436 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 16.399s, the race suite in 85.744s, vet is clean,
and package-header and standalone-oracle audits pass.

## GIA packed-simulation and dataset execution checkpoint

Port all 13 definitions in `src/aig/gia/giaGen.c` as bounded concrete Go
operations. Preserve CI-major input initialization, object-order ordinary-AND
propagation, complement handling, the reset/seed random stream and its
source-specific binary sample-count behavior. Replace direct filesystem writes
with owned named byte artifacts and specify native integer headers as
little-endian int32. Replace file reads, stdout statistics and timing output
with bounded byte parsing, typed comparisons and deterministic JSON bytes. Add
no interfaces and reject native buffers, XORs and muxes because this source
does not distinguish them from ordinary ANDs during simulation.

Validate explicit packed words bit-for-bit with independent Boolean evaluation.
Decode every generated binary sample and re-evaluate its output; parse every
generated PLA row and do the same. Pin artifact names, headers, stream reset and
continuation, validation/test separation, ownership, malformed rows, dimensions,
limits and cancellation. Fuzz random ordinary graphs against direct Go truth
evaluation and fuzz PLA bytes against a canonical Go-only round trip for at
least 5,000 executions each. Do not create a C capture helper, compiler-driving
test, runtime ABC dependency or retained fuzz corpus; the permanent boundary
test must continue rejecting native source and `testdata/` anywhere in `pkg`.

This promotes `giaGen.c` to validated. Inventory becomes 282 validated, 40
partial and 1,435 unstarted files. Both fuzz campaigns passed 5,000 executions,
the complete suite passed in 16.712s, the race suite in 86.084s, vet is clean,
and package-header and standalone-oracle audits pass.

## Espresso minimum-column-cover execution checkpoint

Port the sole operation in `src/misc/espresso/sminterf.c` on the concrete MVC
packed-cover representation: each cube is a sparse-matrix row, and the result
is an owned bit mask selecting columns that hit every row. Replace Espresso's
global cube dimensions and sparse-matrix allocation with checked per-cover
state. Preserve deterministic valid-cover behavior while intentionally
strengthening the native non-branching heuristic call to exact minimum
cardinality under explicit row, column, dominance-comparison and search-node
budgets. Use row dominance, a greedy upper bound, a disjoint-row lower bound
and deterministic coverage-first branching; add no interfaces.

Pin empty, impossible, dominated, duplicate, tied, wide, ownership, limit and
cancellation cases. Exhaust every matrix family through three columns and
compare minimum cardinality with direct subset enumeration. Fuzz at least 5,000
bounded matrices through eight columns against the same independent brute-force
model, checking both optimality and that every row is hit. Retain no generated
corpus or native oracle.

This promotes `sminterf.c` to validated. Inventory becomes 283 validated, 40
partial and 1,434 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 16.743s, the race suite in 86.108s, vet is clean,
and package-header and standalone-oracle audits pass.

## Espresso sparse-dominance execution checkpoint

Port both operations in `src/misc/espresso/dominate.c` on the concrete owned
MVC cover. Row dominance must preserve source row identities, least-population
column selection, strict-superset deletion and equal-size higher-row tie
removal. Column dominance must preserve dynamic row lengths, least-population
row selection, unit or caller weights, coverage-superset deletion and the
equal-size higher-column preference. Return owned reduced covers, active-column
masks and removal counts instead of exposing mutable linked sparse objects; add
no interfaces.

Use packed-word row populations and subset checks while retaining dynamic
column membership. Bound rows, columns and dominance checks, poll cancellation,
reject empty sparse rows and malformed weights, and keep input/output ownership
independent. Pin exact deletion order, weighted and unweighted cases, duplicate
ties, unused columns and all failure surfaces. Exhaust every matrix family
through three columns and fuzz at least 5,000 matrices through eight columns,
proving by brute force that unweighted and weighted minimum cover cost is
unchanged. Use no native oracle or retained generated corpus.

This promotes `dominate.c` to validated. Inventory becomes 284 validated, 40
partial and 1,433 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 16.525s, the race suite in 86.283s, vet is clean,
and package-header and standalone-oracle audits pass.

## Espresso sparse block-partition execution checkpoint

Port `sm_block_partition` and its row/column traversal and copy helpers from
`src/misc/espresso/part.c` on the concrete MVC cover. Traverse the bipartite
row/column graph iteratively from the first row, preserve original row order
and present-column identities, and split the first component from all remaining
components. Preserve the empty-matrix answer and the source's early connected
answer when every present column is reached, even if disconnected empty rows
remain. Return owned covers and typed identity metadata instead of mutable
sparse-matrix pointers; add no interfaces.

Bound rows, columns, nonzero entries and visits; poll cancellation during graph
construction, traversal and result construction; close partial owned results
atomically. Pin connected, disconnected, multi-component, empty-row and
all-columns-reached cases, ownership, every failure surface and a deep iterative
component. Fuzz at least 5,000 bounded matrices against an independent
fixed-point bipartite component model, checking covers and all row/column IDs
without a native oracle or retained corpus.

This promotes `part.c` to validated. Inventory becomes 285 validated, 40
partial and 1,432 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 16.612s, the race suite in 86.350s, vet is clean,
and package-header and standalone-oracle audits pass.

## Espresso minimum-cover solution execution checkpoint

Port all seven operations in `src/misc/espresso/solution.c` as concrete owned
solution state over MVC covers. Preserve zero-cost construction, release,
independent duplication, unit/caller-weighted selection, accept-side covered-
row deletion, reject-side column and newly empty-row deletion, lower-cost best
choice and first-argument ties. Make native ownership transfer explicit when
choosing the better solution. Return owned transformed covers and commit accept
atomically rather than exposing pointer-linked sparse mutation; add no
interfaces.

Reject invalid dimensions, duplicate selection, negative weights and cost
overflow; bound rows and nonzero result entries and poll cancellation during
transformation. Pin lifecycle, clone ownership, weighted costs, tie ownership,
accept/reject matrices, absent-column no-ops, empty-row extension behavior and
every failure surface. Fuzz at least 5,000 bounded matrices against an
independent bit-mask state model, retaining neither minimized cases nor a native
oracle.

This promotes `solution.c` to validated. Inventory becomes 286 validated, 40
partial and 1,431 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 16.970s, the race suite in 86.585s, vet is clean,
and package-header and standalone-oracle audits pass.

## Espresso Gimpel-reduction execution checkpoint

Port `gimpel_reduce` from `src/misc/espresso/gimpel.c` over concrete MVC covers
and cover solutions. Preserve first eligible two-column selection, secondary
residual capture, merging that residual into second-column rows, deletion of
the two columns and primary/secondary rows, recursive-cover reconstruction and
the residual-intersection choice between the original columns. Clone base
selection state and preserve strict bound behavior. Use the validated exact Go
minimum-cover engine for the reduced problem and add no interfaces.

Make the native call-order assumptions explicit: dominance must have removed a
secondary row containing both primary columns, and essential selection must
have removed a singleton secondary residual. Unsafe raw matrices are
non-applicable. Bound rows, columns, entries and reduced search, poll
cancellation, and return only owned state. Pin both reconstruction branches,
base ownership, bounds, hidden preconditions and every failure surface. Fuzz at
least 5,000 bounded matrices and prove every applied result both covers all
original rows and matches brute-force minimum cardinality. Retain no minimized
fuzz case or native oracle.

This promotes `gimpel.c` to validated. Inventory becomes 287 validated, 40
partial and 1,430 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 16.840s, the race suite in 86.226s, vet is clean,
and package-header and standalone-oracle audits pass.

## Espresso maximal-independent-row execution checkpoint

Port both active definitions in `src/misc/espresso/indep.c` over concrete MVC
covers. Build the row-intersection graph without pointer-linked sparse matrix
state. Preserve original row identities, source-order ties, minimum active
closed-neighborhood degree, closed-neighborhood removal and unit or minimum
original-row-column weight. Return owned sorted row identities and expose the
implicit greedy selection order for deterministic testing; add no interfaces.

Reject empty rows and invalid weights, bound dimensions, pair checks, adjacency,
visits and cost overflow, and poll cancellation throughout graph construction
and selection. Pin exact tie order, weighted cost, ownership, atomic errors,
all resource limits and a case wider than one machine word. Exhaustively compare
every row family through three columns with an independent graph model and
brute-force minimum cover, proving independence, maximality and the lower-bound
property. Fuzz at least 5,000 matrices through eight columns against the same
Go-only model without retaining a corpus or native oracle.

This promotes `indep.c` to validated. Inventory becomes 288 validated, 40
partial and 1,429 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 16.675s, the race suite in 85.435s, vet is clean,
and package-header and standalone-oracle audits pass.

## Espresso weighted minimum-cover execution checkpoint

Port `mincov.c` and both minimum-cover headers over the concrete MVC cover,
solution, dominance, partition, Gimpel and independent-row layers. Expose unit
or caller-supplied nonnegative costs, owned sorted selected columns, total cost
and concrete deterministic statistics. Exact mode must be a bounded minimum-
cost search; heuristic mode must preserve the no-branch valid-cover contract.
Replace wall-clock/debug globals, pointer ownership and aborts with explicit
results and errors; add no interfaces.

Use row dominance, a deterministic cost-normalized greedy incumbent, weighted
disjoint-row bounds and coverage-ordered branching. Check zero weights, maximum
integer costs and overflow. Bound rows, columns, pair preprocessing and search
nodes, and poll cancellation throughout. Pin a greedy-suboptimal matrix,
weighted choices, ownership, all failures and empty behavior. Exhaust every
matrix through three columns under all `{0,1,2}` weight vectors against brute
force, then fuzz at least 5,000 matrices through eight columns while replaying
both exact and heuristic answers without a native oracle or retained corpus.

This promotes `mincov.c`, `mincov.h` and `mincov_int.h` to validated. Inventory
becomes 291 validated, 40 partial and 1,426 unstarted files. The fuzz campaign
passed 5,000 executions, the complete suite passed in 16.744s, the race suite
in 86.423s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso sparse-matrix foundation execution checkpoint

Port `matrix.c`, `rows.c`, `cols.c` and both sparse headers as concrete owned
sorted row, column and matrix types. Preserve ascending numeric order, unique
insertion, bidirectional membership, lookup/removal, empty-header deletion,
row and transposed-column copy, longest first-tie selection, vector containment,
intersection, lexical comparison, modular hashing and deterministic text
diagnostics. Replace intrusive links, free lists and generic user pointers with
owned slices and concrete integer metadata; add no interfaces.

Provide bounded byte-slice parsers for pair text and the source compressed
format, canonical text output, source-style formatting, clone/close lifecycle
and a structural validator. Reject negative IDs, malformed counts, truncated
words and high bits beyond the declared final compressed column. Pin every
operation, metadata, ownership, display, error and limit. Fuzz at least 5,000
arbitrary mutation sequences while checking both indexes after every step, and
fuzz at least 5,000 compressed inputs through Go-only canonical round trips.

This promotes `matrix.c`, `rows.c`, `cols.c`, `sparse.h` and `sparse_int.h` to
validated. Inventory becomes 296 validated, 40 partial and 1,421 unstarted
files. Both fuzz campaigns passed 5,000 executions, the complete suite passed
in 16.917s, the race suite in 86.352s, vet is clean, and package-header and
standalone-oracle audits pass.

## Espresso packed-set foundation execution checkpoint

Port all of `set.c` as concrete owned packed sets and set families. Preserve
32-bit element numbering, rank/flag metadata, every Boolean operation and
predicate, active compaction, source replacement-with-last deletion, owned
copy/join/consuming append, counts, restricted weighting, column insertion,
deletion, copy and compression, transpose, arbitrary permutation and exact
text diagnostics. Replace packed pointer aliasing, allocator garbage lists and
static buffers with checked owned slices; add no interfaces.

Provide bounded parsers and canonical writers for both the native hexadecimal
family format and bit matrices. Preserve the source distinction between
metadata-clearing rebuild transforms and the metadata-preserving same-word
append fast path. Reject malformed loop metadata, tail bits, dimensions,
truncation and excessive resources. Exhaust every pairwise set operation
through eight bits, pin all family ordering, metadata, ownership, transforms,
formats and failures, then run separate 5,000-case Go-only operation and
family-round-trip/permutation fuzz campaigns.

This promotes `set.c` to validated. Inventory becomes 297 validated, 40 partial
and 1,420 unstarted files. Both fuzz campaigns passed 5,000 executions, the
refreshed complete suite passed in 17.351s, the refreshed race suite in 88.660s,
vet is clean, and package-header and standalone-oracle audits pass.

## Espresso variable-aware cube-set execution checkpoint

Port all of `setc.c` over `EspressoSet` with an immutable concrete cube layout
that records binary and multiple-valued part ranges. Replace global cube masks
and word-specific bit tricks while preserving full-row tests, zero/exact/capped
distance, force-lower, consensus, sole-active-variable and common-active tests.
Preserve Rank-first descending/ascending order, reverse-word lexical order,
masked distance-one comparison and nullable descending order; add no interfaces.

Validate binary arity, dimensions and total parts, and return owned results.
Pin mixed binary/MV variables, wide-word precedence, masks, nil order,
ownership and every failure. Exhaust every pair of arbitrary cubes in a
seven-part layout against independent per-variable distance and consensus
models, then fuzz at least 5,000 cases in a wider mixed layout while checking
distance, capped distance, consensus and force-lower without a native oracle.

This promotes `setc.c` to validated. Inventory becomes 298 validated, 40
partial and 1,419 unstarted files. The fuzz campaign passed 5,000 executions,
the complete suite passed in 17.106s, the race suite in 92.777s, vet is clean,
and package-header and standalone-oracle audits pass.

## Espresso containment execution checkpoint

Port all of `contain.c` over concrete owned `EspressoSetFamily` values. Preserve
the source's cardinality-derived Rank overwrite, descending containment order
by reverse packed words, ascending reverse-containment order, stable duplicate
survival, indexed source-row remapping, contained union and masked distance
merging. Use the variable-aware cube layout for distance-one masks. Replace
pointer-array sorting/list conversion and consuming native inputs with owned
Go results; add no interfaces.

Use bounded quadratic subset checking rather than reproducing the intrusive
pointer-optimized removal routines, with explicit family/check limits and
cancellation. Pin exact ordering, metadata, duplicates, strict containment,
remapping, cross-family union, masked merging, ownership and failures. Compare
containment and distance merging against separate Go-only models in at least
5,000 fuzz executions each, with no compiled or runtime native oracle.

This promotes `contain.c` and all 18 active definitions to validated. Inventory
becomes 299 validated, 40 partial and 1,418 unstarted files. Both fuzz campaigns
passed 5,000 executions, the complete suite passed in 17.363s, the race suite
in 87.773s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso minterm-map execution checkpoint

Port all of `map.c` over concrete `EspressoCubeLayout`, `EspressoSet` and
`EspressoSetFamily` values. Preserve mixed-radix minterm numbering with variable
zero least significant, output-major map offsets, exact 16-by-16 coordinate
order and spacing, and the source's linear page-count behavior above eight
binary inputs. Replace process-global recursion scratch with local owned state
and stdout output with a returned string; add no interfaces.

Check layout/family dimensions and product overflow before allocation. Bound
cubes, generated minterms and traversal visits, and support synchronous
cancellation without exposing partial results. Pin all 256 coordinate entries,
mixed-radix ordering, exact formatted output, ownership and each failure. Fuzz
at least 5,000 arbitrary binary covers against a separate assignment-level Go
model without C source, a compiler or a runtime oracle.

This promotes `map.c` and all three active definitions to validated. Inventory
becomes 300 validated, 40 partial and 1,417 unstarted files. The fuzzer passed
5,000 executions, the complete suite passed in 17.159s, the race suite in
88.771s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso cover-cost execution checkpoint

Port all of `cvrmisc.c` using concrete cover-cost values, cube layouts and an
owned totals accumulator. Count missing binary and dense-MV parts, present
sparse-MV and output parts, prime flags and total literals exactly as the
source's empty-cofactor `massive_count` path. Use Go value copies for
`copy_cost`. Replace static formatting buffers and stdout with exact returned
cost, size and trace strings, and replace fatal exits with checked errors; add
no interfaces.

Accept elapsed milliseconds from callers rather than consulting a process
clock, preserving native two-decimal truncation deterministically. Update
elapsed/call totals only after cost validation succeeds. Check dimensions,
arithmetic, cube/check limits and cancellation. Pin sparse/dense accounting,
prime counts, both formatting branches, exact reports, ownership, totals and
atomic failures. Fuzz at least 5,000 arbitrary mixed covers against a separate
per-part Go model without a native oracle.

This promotes `cvrmisc.c` and all eight active definitions to validated.
Inventory becomes 301 validated, 40 partial and 1,416 unstarted files. The
fuzzer passed 5,000 executions, the complete suite passed in 17.236s, the race
suite in 88.683s, vet is clean, and package-header and standalone-oracle audits
pass.

## Espresso cofactor execution checkpoint

Port all of `cofactor.c` with a concrete owned cube-list type rather than
pointer/sentinel arrays. Preserve one-, two- and three-family list order,
cumulative complement masks, general distance-zero cofactoring, explicit
source-pointer exclusion, single-variable cofactoring and owned unlisting.
Make the single-variable precondition explicit and retain general semantics for
malformed empty products; add no interfaces.

Return owned per-part/per-variable zero counts, active and unate metadata and
the exact source best-variable tie rules, including its executable MV balance
value. Preserve ordered floor-half binate splitting. Implement intended stable
masked sorting and deduplication while correcting `simplify_cubelist`'s dormant
off-by-two truncation defect. Bound cubes, variable checks and merge
comparisons, and support cancellation without partial results. Pin list
ownership, metadata, exclusion, masks, all count fields, split order,
simplification and failures. Fuzz at least 5,000 arbitrary cube lists against
separate Go models for cofactoring, specialized equivalence, counting,
unlisting and masked deduplication, retaining minimized discoveries as Go seeds
rather than `testdata/` or native fixtures.

This promotes `cofactor.c` and all nine active definitions to validated.
Inventory becomes 302 validated, 40 partial and 1,415 unstarted files. The
fuzzer passed 5,000 executions, the complete suite passed in 17.317s, the race
suite in 88.005s, vet is clean, and package-header and standalone-oracle audits
pass.

## Espresso cube-structure execution checkpoint

Port all of `cubestr.c` as immutable owned cube-layout construction. Preserve
binary part-size normalization, absolute sizing and symbolic markers for
negative MV declarations, variable/output counts, part and packed-word ranges,
per-variable/binary/MV masks, full/empty cubes, default sparsity and the binary
input-word mask. Keep Go word indexes explicitly zero based because packed set
metadata lives in fields rather than a header word.

Use Go ownership instead of explicit teardown. Replace the native one-slot
save/restore globals with independently usable immutable layouts and deep
clones so nesting and concurrent use are safe; produce former `cdata` arrays
on demand through concrete count analysis. Pin symbolic inputs, normalization,
wide word boundaries, every derived range/mask, output rules, sparse defaults,
clone ownership and errors. Fuzz at least 5,000 arbitrary layouts while proving
contiguous nonoverlapping variable masks and exact binary/MV/full partitioning,
without C source or a runtime oracle.

This promotes `cubestr.c` and all four active definitions to validated.
Inventory becomes 303 validated, 40 partial and 1,414 unstarted files. The
fuzzer passed 5,000 executions, the complete suite passed in 17.404s, the race
suite in 88.833s, vet is clean, and package-header and standalone-oracle audits
pass.

## Espresso binary cube-sizer execution checkpoint

Port all of `cubehack.c` with a concrete owned, concurrency-safe cube sizer and
a concrete minimizer-flags record. Preserve exact all-binary logical layouts,
same-size reuse, explicit undefinition, independent snapshots, counts above
100 and all ten default flag assignments. Omit the native <=100-input oversized
allocation because it is an internal memory-reuse strategy with no logical
effect; add no interfaces or process-global mutable state.

Reject negative and over-capacity dimensions atomically. Pin zero, ordinary,
100-boundary and larger layouts, repeated definitions, snapshot ownership,
undefinition, each default, invalid inputs and concurrent access. Fuzz at least
5,000 define/cautious-define/undefine sequences while checking exact dimensions
and binary/full-mask equality, without C source or a runtime oracle.

This promotes `cubehack.c` and all four active definitions to validated.
Inventory becomes 304 validated, 40 partial and 1,413 unstarted files. The
fuzzer passed 5,000 executions, the complete suite passed in 17.219s, the race
suite in 89.933s, vet is clean, and package-header and standalone-oracle audits
pass.

## Espresso sharp algebra execution checkpoint

Port all of `sharp.c` as concrete cube-layout cover algebra. Preserve ordinary
and disjoint cube sharp, balanced cube-by-cover recursion, both cover-sharp
orientations, cover intersection, canonical containment, unchanged-cube
metadata and ordered pairwise-disjoint cover construction. Normalize malformed
empty products according to their set semantics. Replace the fixed 500-cube
batching with a bounded canonical containment pass of identical cover meaning,
and replace debug timing/stdout with deterministic returned values; add no
interfaces.

Use one shared operation state across nested recursion. Charge every inspected
part, cube pair and generated intermediate cube, support cancellation, and
expose no partial result on failure. Pin cube order, metadata, ownership, empty
cases, every operation orientation, disjointness, limits and errors. Fuzz at
least 5,000 arbitrary small covers against independent exhaustive assignment
sets for difference, intersection and disjointization without C source or a
runtime oracle.

This promotes `sharp.c` and all ten active definitions to validated. Inventory
becomes 305 validated, 40 partial and 1,412 unstarted files. The fuzzer passed
5,000 executions, the complete suite passed in 17.084s, the race suite in
88.787s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso prime-consensus execution checkpoint

Port all of `primes.c` over concrete owned cube lists and layouts. Preserve
empty, singleton and full-row special cases, common-zero-column factoring,
active/unate analysis, exact best-variable splitting, single-variable branch
cofactors, split-cofactor filtering, distance-one cross-branch consensus and
canonical containment. Keep inputs owned rather than consuming pointer lists.
Replace fixed 500-cube batching with one bounded canonical containment pass of
identical cover meaning; add no interfaces.

Share recursion-depth, operation and generated-cube budgets across every
special scan, count analysis, cofactor, wide-MV intersection, consensus and
containment operation, with atomic cancellation. Pin all special cases,
factoring, unate and mixed-MV behavior, recursive parity, ownership and failure
classes. Independently enumerate every valid small cube, select implicants and
retain only maximal ones, then fuzz at least 5,000 arbitrary covers through
four variables against that model without a native oracle.

This promotes `primes.c` and all four active definitions to validated.
Inventory becomes 306 validated, 40 partial and 1,411 unstarted files. The
fuzzer passed 5,000 executions, the complete suite passed in 17.319s, the race
suite in 89.953s, vet is clean, and package-header and standalone-oracle audits
pass.

## Espresso essential-prime execution checkpoint

Port all of `essen.c` over concrete layouts and owned cover results. Preserve
distance-zero containment suppression, MV per-variable merge, binary
intersection fallback, distance-one consensus, explicit row exclusion and the
sharp-consensus plus don't-care essentiality criterion. Preserve
`NONESSEN`/`RELESSEN` gating, stable order and `ACTIVE` metadata, returning
essential, remaining-ON and augmented-DC covers rather than mutating caller
pointers; add no interfaces.

Make the source precondition explicit in tests: ON rows are primes of ON∪DC
and intersect ON. Share resource accounting with nested sharp traversal, bound
generated cubes and packed consensus work, and support atomic cancellation.
Pin mixed binary/MV construction, every distance case, flags, order, ownership
and failures. Independently enumerate assignments, construct maximal primes of
arbitrary ON∪DC functions, filter ON-intersecting rows and fuzz at least 5,000
extractions against unique-ON-minterm essentiality without a native oracle.

This promotes `essen.c` and all four active definitions to validated. Inventory
becomes 307 validated, 40 partial and 1,410 unstarted files. The fuzzer passed
5,000 executions, the complete suite passed in 17.232s, the race suite in
88.335s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso verification and PLA-core execution checkpoint

Port all of `verify.c` using a concrete owned PLA record and structured
verification results. Preserve both cover-containment directions modulo the old
don't-care set and return counterexample covers instead of printing. Perform
strict deterministic label association across ON/DC/OFF, rejecting missing or
ambiguous labels and returning a new PLA on the reference layout rather than
mutating caller state. Preserve the source PLA comparison direction. Return
three pairwise overlaps and unspecified space from consistency checking; add no
interfaces.

Share one bounded sharp-work state across nested differences/intersections and
support atomic cancellation. Preserve filename/type metadata and independent
label ownership. Pin directional witnesses, permutation, ownership,
consistency witnesses, limits and errors. Fuzz at least 5,000 arbitrary small
PLAs against independent assignment sets for both verification directions,
every overlap, universe coverage and unspecified minterms without a native
oracle.

This promotes `verify.c` and all four active definitions to validated.
Inventory becomes 308 validated, 40 partial and 1,409 unstarted files. The
fuzzer passed 5,000 executions, the complete suite passed in 17.397s, the race
suite in 88.967s, vet is clean, and package-header and standalone-oracle audits
pass.

## Espresso output-equivalence execution checkpoint

Port both definitions in `equiv.c` using concrete covers and typed pair
records. Implement `check_equiv` as semantic mutual containment, so cube order,
duplication and representation do not affect equality. For each multi-output
OFF set, select rows containing the output part and make the output variable
unrestricted. Preserve stable lexicographic pair order and the source's direct
equivalence preference. Detect inversion by proving the two OFF covers are
disjoint and jointly exhaustive; this is exactly the source complement
comparison and does not couple this unit to a general complement engine. Return
owned labels, filling absent output names with source-compatible defaults, and
do not mutate the PLA or print diagnostics.

Share one bounded sharp-work state across all cofactors, containment checks,
intersections and universe differences, with synchronous cancellation. Pin
semantic equality, representation independence, polarity, ordering, labels,
ownership, empty results, errors and limits. Independently generate arbitrary
small output truth vectors, encode them as multi-output OFF covers and fuzz at
least 5,000 analyses against direct equality and bitwise complementation with
no native oracle or retained corpus.

This promotes `equiv.c` and both active definitions to validated. Inventory
becomes 309 validated, 40 partial and 1,408 unstarted files. The fuzzer passed
5,000 executions, the complete suite passed in 17.253s, the race suite in
88.101s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso complement and simplification execution checkpoint

Port all twelve definitions in `compl.c` as concrete owned cover operations.
Compute exact complementation as universe sharp cover, reusing the validated
sharp traversal to absorb the private recursive special cases, split/merge,
distance-one and lifting implementations. Implement `simp_comp` as a paired
complement and exact double complement, and `simplify` as its simplified side.
This preserves Boolean semantics while deliberately replacing global scratch,
debug recursion levels and heuristic branch-specific cube ordering with a
deterministic representation; add no interfaces.

Use one shared work state for both paired results so cube and operation limits
are aggregate, cancellation is synchronous and errors expose no partial cover.
Pin empty and full functions, mixed binary/MV complements, pair consistency,
ownership, errors and resource limits. Independently enumerate assignments for
arbitrary small covers and fuzz at least 5,000 complement/simplification pairs,
requiring the complement truth to be universe minus source and the simplified
truth to equal source without a native oracle or retained corpus.

This promotes `compl.c` and all twelve definitions to validated. Inventory
becomes 310 validated, 40 partial and 1,407 unstarted files. The fuzzer passed
5,000 executions, the complete suite passed in 17.269s, the race suite in
88.882s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso reduction execution checkpoint

Port all six definitions in `reduce.c` with concrete owned covers. Compute SCCC
by complementing a cover with the validated sharp engine and taking the
positional supercube, absorbing recursive split, merge, cube and special-case
helpers. Reduce a target cube by taking target sharp (other current ON union
DC), then the positional supercube of that essential region. Preserve source
sequential replacement, empty reductions, `ACTIVE`/`PRIME` transitions and the
two distance and column score orderings. Replace the unsafe process-global
alternating toggle with an explicit concrete order option and stable ties; add
no interfaces.

Share one aggregate work state across ordering, sharp traversal and supercube
construction, with synchronous cancellation and atomic errors. Pin SCCC,
single-cube reduction, both orderings, flags, ownership, ON-union-DC
preservation, dimensions and limits. Independently enumerate essential
assignments for arbitrary small covers and fuzz at least 5,000 reductions in
each ordering without a native oracle or retained corpus. Re-run affected
neighboring fuzz models whenever their independent encoders are corrected.

This promotes `reduce.c` and all six definitions to validated. Inventory
becomes 311 validated, 40 partial and 1,406 unstarted files. The reduction
fuzzer passed 5,000 executions and the corrected output-equivalence fuzzer
passed 5,028. The complete suite passed in 17.380s, the race suite in 88.556s,
vet is clean, and package-header and standalone-oracle audits pass.

## Espresso unate-family execution checkpoint

Port all nine definitions in `unate.c` with concrete owned set families.
Implement unate complementation and the private covered/select helpers as
deterministic minimal-transversal enumeration, retaining every
inclusion-minimal hitting set and stable first duplicates. Preserve
`exact_minimum_cover` as all minimal covers despite its historical name.
Implement unate intersection as nonempty pairwise intersection followed by
inclusion-maximal containment, with optional maximum-cardinality filtering.

Replace global `cdata` during positional-cover mapping with an owned record
containing the immutable layout, active-part/unate-variable association and
part-zero counts. Reject binate inputs, return owned mapped families and map
results back without aliasing. Bound generated sets and operations, check
cancellation synchronously and expose no partial results; add no interfaces.
Pin empty/impossible inputs, exact and largest-only intersections, mapping
metadata, ownership, complement round trips, dimensions and limits. Compare
arbitrary small families against exhaustive minimal hitting sets and
intersections for at least 5,000 fuzz executions, and separately fuzz monotone
cover map/complement/inverse-map semantics for every assignment without a
native oracle or retained corpus.

This promotes `unate.c` and all nine definitions to validated. Inventory
becomes 312 validated, 40 partial and 1,405 unstarted files. The family fuzzer
passed 5,031 executions and the mapping fuzzer passed 5,000. The complete suite
passed in 17.259s, the race suite in 88.777s, vet is clean, and package-header
and standalone-oracle audits pass.

## Espresso exact-minimization execution checkpoint

Port all four definitions in `exact.c` with concrete owned minimization
results. Generate all primes of ON∪DC through the validated consensus engine,
discard DC-only primes and solve a bounded deterministic prime-subset problem.
In cube mode minimize cube count then source-compatible literal cost, thereby
subsuming the heuristic sparse post-pass. In literal mode minimize literal cost
then cube count. Preserve `exact_cover=false` as a deterministic irredundant
heuristic result without claiming optimality. Require orthogonal ON/DC/OFF
inputs and add no interfaces.

Return the owned prime inventory, selected indices, objective and final cover
instead of `dump_irredundant` filename and stdout side effects. Bound prime
generation, packed work, recursion depth and subset nodes, check cancellation
synchronously and expose no partial result. Pin both exact objectives,
heuristic and empty functions, ownership, malformed partitions, dimensions and
every limit. Independently enumerate every valid positional cube for arbitrary
small Boolean functions, derive maximal primes, test every prime subset and
fuzz at least 5,000 cases against both objective optima and assignment truth
without a native oracle or retained corpus.

This promotes `exact.c` and all four definitions to validated. Inventory
becomes 313 validated, 40 partial and 1,404 unstarted files. The fuzzer passed
5,037 executions, the complete suite passed in 17.267s, the race suite in
88.980s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso irredundancy execution checkpoint

Port all ten definitions in `irred.c` with concrete owned results. Replace
recursive tautology/cofactor special cases with exact complement and sharp
checks. Preserve the source two-stage essential/redundant and
totally/partially-redundant classification, retaining original row indices.
Derive an exact covering table by partitioning uncovered care space according
to partially redundant candidate membership, with stable deduplicated rows and
no process-global current-column state.

Return both a source-order marked cover and its active owned subset, preserving
`ACTIVE` and `RELESSEN`. Make deterministic greedy-plus-deletion selection the
default source-compatible heuristic and offer bounded exact
minimum-cardinality transversal selection without interfaces. Validate every
family/index invariant; bound cubes, membership rows and operations; check
cancellation synchronously; expose no partial result. Pin classification,
table rows, flags, ownership, tautology, both selectors, dimensions and limits.
Fuzz at least 5,000 arbitrary small covers against assignment-level E/Rt/Rp
and membership-table models and prove ON∪DC preservation in both modes without
a native oracle or retained corpus.

This promotes `irred.c` and all ten definitions to validated. Inventory
becomes 314 validated, 40 partial and 1,403 unstarted files. The fuzzer passed
5,023 executions, the complete suite passed in 17.095s, the race suite in
88.624s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso expansion execution checkpoint

Port all twelve definitions in `expand.c` with concrete owned expansion and
prime-enumeration results. Replace mutable blocking/covering matrices and the
private essential-part, feasible-selection, most-frequent and minimum-cover
scratch paths with exact prime enumeration of the OFF-set complement. Select a
prime containing each source cube by maximum eligible ON-row coverage and then
source-compatible literal cost. Preserve ascending column-inner-product cover
ordering, caller-marked prime skipping, later nonprime coverage, covered-row
deletion and `PRIME`/`ACTIVE`/`NONESSEN` metadata.

For nonsparse expansion, restrict exact global primes back to the source cube
on sparse variables, yielding maximal dense-only candidates. Return every
prime covering a cube and aggregate per-source primes with stable ordering and
marked-prime passthrough. Bound cubes, operations and prime recursion depth;
check cancellation and ON/OFF orthogonality; return owned atomic results; add
no interfaces. Pin exact prime sets, dense-only maximality, flags, ownership,
semantics and every failure class. Fuzz at least 5,000 arbitrary small Boolean
functions in full and nonsparse modes, proving exact truth, OFF avoidance,
one-part maximality and exhaustive primes covering a chosen minterm without a
native oracle or retained corpus.

This promotes `expand.c` and all twelve definitions to validated. Inventory
becomes 315 validated, 40 partial and 1,402 unstarted files. The fuzzer passed
5,003 executions, the complete suite passed in 17.266s, the race suite in
87.888s, vet is clean, and package-header and standalone-oracle audits pass.

## Espresso sparse-cleanup execution checkpoint

Port both definitions in `sparse.c` with concrete owned cover operations.
Implement `mv_reduce` by forming a single-part cofactor for every part of every
sparse variable, applying the validated irredundancy classifier, and deleting
the original part when the corresponding cofactor row is redundant. Preserve
the source exception for a full non-output variable, remove cubes left empty
in any sparse variable, and retain `ACTIVE`/`PRIME` transitions. Permit exact
or deterministic heuristic irredundancy without interfaces.

Implement `make_sparse` as alternating sparse reduction and validated
nonsparse expansion, stopping when source-compatible literal cost does not
change or after the single pass requested by the caller. Replace the implicit
loop bound with an explicit iteration limit. Bound cubes, operations,
irredundancy rows and expansion depth; check cancellation synchronously and
return owned atomic results. Pin mixed binary/MV cofactors, truth, OFF
exclusion, flags, ownership, dimensions and every failure class. Independently
enumerate arbitrary small truth tables and fuzz at least 5,000 cases, requiring
ON-union-DC preservation, OFF exclusion, nonincreasing cost and a nonempty part
set for every retained sparse variable, without a native oracle or retained
corpus.

This promotes `sparse.c` and both definitions to validated. Inventory becomes
316 validated, 40 partial and 1,401 unstarted files. The fuzzer passed 5,000
executions, the complete suite passed in 17.259s, the race suite in 88.133s,
vet is clean, and the package contains no native source, header, `testdata/` or
runtime oracle dependency.

## Espresso GASP execution checkpoint

Port all six definitions in `gasp.c` using the validated reduction, expansion,
prime-enumeration and irredundancy engines. Implement `reduce_gasp` as
independent maximal reduction of every row against the unchanged original ON
union DC cover, preserving order and prime metadata. Treat an already
redundant row as a checked invariant failure instead of terminating the
process.

Implement `expand1_gasp` and `expand_gasp` by deriving each other row's
essential region after substituting the first reduced row, then using exact
OFF-complement primes to decide whether their union has a feasible expansion.
Deduplicate candidate unions and expand them to primes. Compose `irred_gasp`,
`last_gasp` and `super_gasp` from owned joins, validated irredundancy and
all-prime enumeration. Permit DC space to overlap either minimized polarity,
but reject active ON/blocking OFF intersection. Add no interfaces.

Share reduction work across each reduction or candidate-generation phase;
bound cubes, operations, irredundancy rows and prime depth; check cancellation
synchronously and return no partial result. Pin independent reduction, flags,
ownership, candidate expansion, workflow semantics, malformed partitions and
every limit. Generate arbitrary small truth tables, first expand and make the
ON cover irredundant as required by GASP, then fuzz at least 5,000 last- and
super-GASP runs against independent assignment truth with no native oracle or
retained corpus.

This promotes `gasp.c` and all six definitions to validated. Inventory becomes
317 validated, 40 partial and 1,400 unstarted files. The fuzzer passed 5,001
executions, the complete suite passed in 17.337s, the race suite in 87.887s,
vet is clean, and native-source, package-header and `testdata/` audits pass.

## Espresso main minimizer execution checkpoint

Port the sole `espresso.c` definition as a concrete bounded orchestration
method. Preserve optional onset simplification and output unraveling, clear
prime marks before the initial expansion, extract an irredundant cover, and
optionally remove essentials into the working DC cover. Repeat reduction,
expansion and irredundancy until cube count stabilizes, then perturb with last
or super GASP and continue only for a lexicographic cube/literal improvement.
Restore essentials and run sparse cleanup unless disabled.

Replace every process-global strategy flag with `EspressoMinimizeOptions` and
return cost, iteration, unravel and fallback metadata instead of trace/timing
stdout. Retain the source fallback that reruns without output unraveling if
the result grows; strengthen the owned API by returning the original cover if
any other strategy unexpectedly grows it. Bound iterations, cubes,
operations, irredundancy rows and prime depth per phase, check cancellation
synchronously and add no interfaces.

Pin every strategy branch, both reduction orders, output unraveling,
ownership, malformed partitions and all resource errors. Independently encode
arbitrary small truth tables and fuzz at least 5,000 combinations of onset
recomputation, essential extraction, single expansion, both GASP strategies,
sparse cleanup and exact/heuristic irredundancy. Require ON-union-DC
preservation, OFF exclusion, non-growth and exact returned-cost agreement with
no native oracle or retained corpus.

This promotes `espresso.c` and its sole definition to validated. Inventory
becomes 318 validated, 40 partial and 1,399 unstarted files. The fuzzer passed
5,034 executions, the complete suite passed in 17.448s, the race suite in
89.114s, vet is clean, and native-source, package-header and `testdata/` audits
pass.

## Espresso cover-manipulation execution checkpoint

Port all twenty definitions in `cvrm.c` using concrete owned covers and cube
lists. Implement bounded Cartesian unraveling over an inclusive variable
range. Preserve lexical, cardinality, column-inner-product and reduction
scores with deterministic stable sorting. Replace process-global random order
with a caller-seeded shuffle. Preserve `cubelist_partition` as first-seed
common-active-variable closure while returning both owned lists and their
source indices.

Implement output cofactor by selecting rows containing a relative output part
and making that field unrestricted; implement the inverse by restricting rows
back to one part. Replace `foreach_output_function` and every `so_*` callback
plus global `Fmin`/`phase` scratch with a concrete `MinimizeOutputs` loop.
Support heuristic and exact per-output minimization, optional comparison with
the complemented function, positive-polarity ties and returned phase/count
metadata. Add no interfaces.

Bound generated Cartesian rows, sort comparisons, component checks, output
cofactors, cubes and operations; check cancellation synchronously and expose
no partial result. Pin masks, score order, stable ties, ownership, seeded
shuffle, partitions, cofactors, both minimizers, phase choices and all errors.
Fuzz arbitrary small positional covers against assignment truth for at least
5,000 executions, and independently fuzz multi-output truth tables through
both strategies and both polarity modes for at least 5,000 executions, without
a native oracle or retained corpus.

This promotes `cvrm.c` and all twenty definitions to validated. Inventory
becomes 319 validated, 40 partial and 1,398 unstarted files. Both fuzzers passed
5,000 executions, the complete suite passed in 17.421s, the race suite in
88.691s, vet is clean, and native-source, package-header and `testdata/` audits
pass.

## Espresso output-phase execution checkpoint

Port all ten active definitions in `opo.c` with owned PLA and phase values.
Construct the double-phase characteristic function for an output suffix,
preserving fixed positive prefixes, duplicating the remaining ON/OFF columns,
and copying DC columns into both polarities. Apply a chosen phase by splitting
each ON/OFF row into positive and negative output subsets while leaving DC
unchanged.

Replace `opo`, `opo_recur` and `opo_leaf` static recursion with bounded direct
phase enumeration. For each phase, form the union of primes carrying the
chosen polarity; minimizing that union is exactly equivalent to maximizing
the source expression's excludable-prime intersection. Preserve positive-first
ties. Compose doubled minimization and selection into `FindOutputPhase`, then
support both one-shot and repeated single-output commitment before final
minimization. Replace `opoall` with exhaustive enumeration over an inclusive
output range and retain strict cube-count improvement. Add no interfaces.

Replace process-global OPO flags, phase pointers, recursion depth and stdout
with options and returned phase, evaluation and cost metadata. Bound phases,
cubes, operations and exact subset search; check cancellation synchronously
and expose no partial result. Pin doubled characteristic truth, phase
application, carrier-subset choice, heuristic and exact minimization, repeated
commitment, exhaustive enumeration, ownership and every failure class. Fuzz
arbitrary small multi-output truth tables for at least 5,000 transformation
cases and independently fuzz at least 5,000 complete assignments against
positive-or-complement truth without a native oracle or retained corpus.

This promotes `opo.c` and all ten active definitions to validated; the two
`#if 0` definitions remain inactive. Inventory becomes 320 validated, 40
partial and 1,397 unstarted files. The transformation fuzzer passed 5,026
executions and the assignment fuzzer passed 5,003. The complete suite passed
in 17.403s, the race suite in 88.578s, vet is clean, and native-source,
package-header and `testdata/` audits pass.

## Espresso variable-pairing execution checkpoint

Port all sixteen active definitions in `pair.c` with concrete owned pairing,
PLA and result values. Replace disjoint binary pairs by four-valued variables
using the source's exact joint-part encoding. Preserve unpaired binary order,
then pair order, then existing multiple-valued variables; preserve labels and
cube metadata, and force new pair variables nonsparse. Leave all inputs owned
and unchanged.

Preserve deterministic maximum-cardinality pairing enumeration, including one
unpaired variable for odd dimensions. Implement exhaustive and greedy maximum-
cost selection over signed matrices, including checked accumulation. Retain
weak collapsing cost, strong reduce/expand/irredundant cost, and complete
heuristic and exact cost modes. Compose selected pairings with heuristic,
exact and output-phase minimization. Replace global pair arrays, callbacks,
printing and allocation/free helpers with concrete options, returned metadata,
owned clones and formatting. Add no interfaces.

Bound enumerated pairings, cubes, operations, rows, depth and exact-search
nodes; poll concrete cancellation and expose no partial mutation. Pin exact
parts, layouts, labels, truth, recursion order, greedy/exhaustive choices,
formatting, ownership, every strategy, signed costs, overflow and all resource
errors. Fuzz transformations against an independent assignment evaluator and
exhaustive best matching against an independent enumeration for at least 5,000
executions each. Use no native source, generated helper, executable or runtime
C oracle; permanent vectors, if later needed, belong only in `test_vectors/`.

This promotes `pair.c` and all sixteen active definitions to validated.
The transform and matching campaigns passed 5,006 and 5,000 executions. The
complete suite passed in 17.208s, the race suite in 89.505s, vet is clean, and
native-source, package-header and `testdata/` audits pass. Inventory becomes
321 validated, 40 partial and 1,396 unstarted files.

## Espresso PLA-output execution checkpoint

Port all eighteen active definitions in `cvrout.c` as concrete bounded
formatting methods. Preserve standard ON/DC/OFF PLA headers, type directives,
product counts, TPLA `.e` behavior, ordinary `.end` behavior, binary and
multiple-valued cube fields, configurable output maps, labels and phase
annotations. Preserve equation output and its 72-column source wrapping,
PLEASURE labels/groups/expanded rows, KISS symbolic fields and numeric and
symbolic constraint reports.

Return owned strings or bytes in place of `FILE`, stdout and the `pc1`/`pc2`
static buffers. Generate default labels into an owned result rather than
mutating the PLA. Return diagnostic cover/cubelist renderings instead of using
the verbose global. Retain stable projected-pattern order and duplicate
weights without an iteration-ordered map dependency. Add no interfaces.

Bound cubes, visited parts and emitted bytes, poll concrete cancellation and
return no partial output on error. Pin every output mode, exact headers and
rows, label generation, phases, equation wrapping, PLEASURE grouping, KISS
validation, constraint weights, diagnostic output, ownership and every limit.
Fuzz arbitrary cube layouts against an independent formatter for at least
5,000 executions, then independently decode complete marshaled PLAs and prove
every emitted cube equals its source for another 5,000. Use no native oracle,
generated helper or retained corpus.

This promotes `cvrout.c` and all eighteen active definitions to validated. The
cube and complete-marshal campaigns passed 5,000 and 5,009 executions. The
complete suite passed in 17.235s, the race suite in 88.152s, vet is clean, and
native-source, package-header and `testdata/` audits pass. Inventory becomes
322 validated, 40 partial and 1,395 unstarted files.

## Espresso PLA-input partial execution checkpoint

Implement the ordinary portion of all eleven `cvrin.c` definitions with a
concrete bounded byte parser and owned parse result. Support `.i`, `.o`, `.mv`,
`.p`, `.e`, `.end`, `.kiss`, `.type`, `.ilb`, `.ob`, `.label`, `.phase`,
`.pair`, comments, unknown directives and product rows. Preserve binary,
symbolic-MV and F/D/R product encodings, label insertion, warning-and-skip for
malformed products, KISS state/output merging, missing-cover construction,
positive or explicit phase and pair postprocessing.

Parse `.symbolic` and `.symbolic-output` into concrete owned reference/label
records. Until their delegated `hack.c` layout transformations are ported,
default parsing must return `ErrUnsupported`; an explicit defer option may
return the pre-mapping PLA plus metadata for inspection and testing. Return an
owned summary in place of stdout. Reuse existing PLA construction and Go
lifetime for `new_PLA`, `PLA_labels` and `free_PLA`. Add no interfaces.

Bound input bytes, lines, variables, parts, cubes, tokens and operations and
poll concrete cancellation. Pin every directive, exact cover rows, labels,
phases, KISS, pairing, computed polarity covers, summaries, warnings and all
limits. Fuzz at least 5,000 complete Go marshal/parse round trips and 5,000
arbitrary byte streams for panic freedom and successful-result invariants,
without native source or retained fuzz output.

This moves `cvrin.c` from unstarted to partial. The round-trip and robustness
campaigns passed 5,000 and 5,030 executions. Physical-line continuation of
product terms and the two symbolic mappings remain before validation; implement
those mappings with `hack.c`, then remove the defer-only limitation and promote
both units together. The complete suite passed in 17.235s, the race suite in
93.001s, vet is clean, and native-source, package-header and `testdata/` audits
pass. Inventory becomes 322 validated, 41 partial and 1,394 unstarted files.

## Espresso PLA-input completion and symbolic-hack execution checkpoint

Complete `cvrin.c` by treating physical newlines inside a product as ignored
whitespace with one returned warning and by applying parsed symbolic directives
through concrete `hack.c` transformations. Keep explicit deferral as an
inspection feature, not the default semantic path. Re-run both parser fuzz
models after integration.

Port `map_symbolic`, `map_symbolic_cover`, `form_bitvector` and label repair as
one owned `ApplySymbolicInputs` transformation. Preserve unique referenced-
binary deletion, old MV order, appended symbolic MV order, source MSB-first
combination numbering, unconstrained values and user/default labels across all
three covers. Port `map_output_symbolic`, `find_inputs` and label repair as
`ApplySymbolicOutputs`, preserving inserted one-hot groups, recursive F/R
cofactors, DC-driven onset recomputation, removal of old output parts and quick
contain/distance-one cleanup. Pin the source's all-zero-output omission.

Port `map_dcset` as `MapDontCareVariable`: retain the four label prefixes,
cofactor-disagreement DC function, separation of independent ON cubes and
variable/label removal. Return a consistent owned OFF cover rather than leaving
storage at the old width. Bound cubes, operations and symbolic depth, poll
cancellation and add no interfaces. Fuzz input mapping, output mapping and
named-DC truth independently for at least 5,000 executions each.

This promotes `cvrin.c` and all eleven definitions to validated and moves
`hack.c` from unstarted to partial with seven of nine active definitions. The
parser campaigns passed 5,000 and 5,015 executions; input-symbolic,
output-symbolic and named-DC campaigns each passed 5,000. `fsm_simplify` and
`disassemble_fsm` remain. The complete suite passed in 17.414s, the race suite
in 89.015s, vet is clean, and native-source, package-header and `testdata/`
audits pass. Inventory becomes 323 validated, 41 partial and 1,393 unstarted
files.

## Espresso symbolic-hack completion execution checkpoint

Complete the remaining `fsm_simplify` and `disassemble_fsm` definitions with a
concrete `DisassembleFSM` operation. Partition transitions into every-state to
a selected next state, nowhere, and each specific present/next-state pair in
deterministic source order. Minimize each partition independently, unravel the
present-state field for nowhere transitions, and return owned statistics,
diagnostics, regrouped PLA and KISS bytes instead of stdout or global scratch.
Represent a machine with no external outputs without a zero-sized trailing Go
part. Add no interfaces.

Bound cubes, operations, rows, depth, minimizer iterations and emitted bytes;
poll cancellation synchronously and return no partial semantic result on
failure. Pin exact partition statistics, diagnostics and KISS rows, ownership,
zero-output layout and every error class. Fuzz at least 5,000 valid generated
FSM covers against an independent Go relation model that exhaustively compares
each next-state and external-output part before and after disassembly.

This promotes `hack.c` and all nine active definitions to validated. The
standalone relation campaign passed 5,000 executions. The complete suite passed
in 17.248s, the race suite in 88.741s, vet is clean, and the module-wide audit
confirms there is no native source, compiler invocation, executable oracle or
permanent `testdata/`. `pkg/tools/` contains only `source_manifest.py`, and
permanent compatibility evidence remains immutable under `test_vectors/`.
Inventory becomes 324 validated, 40 partial and 1,393 unstarted files.

## Espresso global-state replacement execution checkpoint

Replace every declaration in `globals.c` without introducing writable package
globals. Put driver flags and source defaults in an owned
`EspressoRuntimeOptions`, project supported controls into the narrower concrete
algorithm options, and replace the source's process-global random reduction
order with a deterministic per-run seed. Return owned format descriptors for
all 20 `pla_types` spellings
and provide deterministic lookup with and without the source's leading dash.
Allocate timing names, elapsed values and call counts per driver instance.

Use the existing `EspressoPLA` and parser options for filenames, immutable
`EspressoCubeLayout` for cube metadata, and operation-local work structs for
temporary cube/cdata analysis. Replace `bit_count[256]` with width-correct
`math/bits` operations. Test defaults, projection, every descriptor, copy
ownership and concurrent isolation; exhaust all 256 byte populations and fuzz
combined population/lookup behavior for at least 5,000 executions. Add no
interfaces and no native oracle.

This validates all 27 declarations in `globals.c`. The standalone Go campaign
passed 5,000 executions. The complete suite passed in 17.551s, the race suite
in 88.774s, vet is clean, and native-source and permanent-`testdata/` audits
pass. Inventory becomes 325 validated, 40 partial and 1,392 unstarted files.

## Espresso driver-table execution checkpoint

Port `main.h` as immutable concrete Go metadata before implementing the driver.
Give every source command a typed kind and preserve the ordered 43-row table,
including aliases, PLA arity and OFF/DC requirements. Preserve all debug names
and exact bit masks. Replace the behavior table's pointers to writable globals
with a pure function that takes and returns `EspressoRuntimeOptions`. Keep
source-case-sensitive lookup and return owned descriptor copies. Add no
interfaces.

Pin exact order, aliases, masks, arities, cover requirements, option effects,
unknown-name errors and copy ownership. Fuzz arbitrary names against fresh
descriptor snapshots and pure option application for at least 5,000
executions. Do not promote `main.c` until parsing and command dispatch exist.

This validates `main.h`. The standalone Go campaign passed 5,000 executions.
The complete suite passed in 17.282s, the race suite in 89.914s, vet is clean,
and standalone-oracle audits pass. Inventory becomes 326 validated, 40 partial
and 1,391 unstarted files.

## Espresso driver-parser execution checkpoint

Begin `main.c` with a bounded library parser rather than a process-exiting CLI.
Preserve legacy `-do`/`-out` rewriting, bare behavior switches, source input-
type precedence and modern attached/separate option arguments. Return one owned
concrete configuration containing command metadata, I/O types, runtime flags,
strategy, range and positional inputs. Check command arity, numeric syntax,
argument/byte bounds and cancellation without mutating caller input. Return
deterministic usage text rather than writing stdout. Add no interfaces.

Pin defaults, legacy/modern equivalence, input-type precedence, debug masks,
ranges, strategies, ownership, errors and 76-column usage. Fuzz arbitrary
argument text for at least 5,000 executions, checking successful configuration
invariants and immutable input. Keep `main.c` partial until PLA loading,
dispatch, runtime summary, output and many-PLA behavior are implemented.

This moves `main.c` from unstarted to partial with six definition surfaces
represented. The standalone parser campaign passed 5,000 executions. The full
suite passed in 17.578s, the race suite in 88.774s, vet is clean, and standalone-
oracle audits pass. Inventory becomes 326 validated, 41 partial and 1,390
unstarted files.

## Espresso standalone-driver completion execution checkpoint

Complete `main.c` as a side-effect-free library driver. Dispatch every command
kind to the existing concrete Go implementation and return owned PLAs,
decisions, witnesses, phase/pair metadata and diagnostics. Compose parsing,
required-cover PLA loading, execution and bounded output marshaling over owned
bytes. Preserve print suppression and the `many` loop as batch processing.
Replace stdout usage/runtime reports with returned strings and caller-supplied
elapsed values. Keep files, stdin and process exit in a future thin CLI rather
than embedding them in the package. Add no interfaces.

Test every command group, Boolean semantics, transformation ownership,
phase/pair behavior, verification, end-to-end byte parsing/output, batch mode,
suppression, runtime formatting, errors, limits and cancellation. Run separate
independent Boolean-dispatch and byte-driver fuzz models for at least 5,000
executions each, with no C oracle or retained fuzz corpus.

This validates all nine `main.c` definition surfaces. Both standalone Go
campaigns passed 5,000 executions. The complete suite passed in 17.249s, the
race suite in 88.845s, vet is clean, and standalone-oracle audits pass. All 39
translation units named by `src/misc/espresso/module.make` are individually
validated, promoting that manifest too. Inventory becomes 328 validated, 40
partial and 1,389 unstarted files.

## Espresso header completion execution checkpoint

Finish the Espresso directory by treating `espresso.h` as executable contract
surface. Preserve all numeric PLA, flag, debug, Boolean, cube-input and runtime
constants with typed Go definitions. Map packed set/header macros to checked
`EspressoSet` operations, cube/cdata globals to immutable layouts and local
analysis, and cost/pair/symbolic/PLA records to owned concrete structs. Account
for every function declaration through its already validated translation unit;
do not reproduce C ABI layout or pointer arithmetic.

Pin constants, source layout normalization, masks, tail bits and ownership, and
fuzz packed widths/content for at least 5,000 executions. Give `util_old.h` an
explicit disposition after proving it has no include edge and is absent from
the module manifest: Go ownership, standard-library functionality, concrete
argument parsing, explicit errors and supplied durations replace its obsolete
platform, allocation, fatal, getopt and timing macros.

This validates both remaining headers. The packed-contract campaign passed
5,000 executions. The complete suite passed in 17.122s, the race suite in
90.156s, vet is clean, and standalone-oracle audits pass. Every tracked item in
`src/misc/espresso/` is validated. Inventory becomes 330 validated, 40 partial
and 1,387 unstarted files.

## Extra progress-bar execution checkpoint

Port all five routines in `src/misc/extra/extraUtilProgress.c` as one concrete,
owned state machine. Preserve the fixed 78-column layout, source threshold
truncation, initial position, seven-column steps, at-most-one update per call,
completion saturation, label placement and carriage-return clearing frame.
Model disabled progress as a nil handle and batch mode as state advancement
with suppressed output. Return owned strings instead of accepting a writer or
holding `FILE*`; do not introduce an interface or global frame/output state.

Check item totals and output bounds, make rendering failures atomic, and expose
an explicit closed lifecycle in place of manual allocation/free. Pin exact
frames, recurrence states, disabled/batch behavior, completion, cleanup,
limits and error atomicity. Fuzz arbitrary totals and update sequences against
an independent recurrence model for at least 5,000 executions.

This validates `extraUtilProgress.c`. The Go-only campaign passed 5,002 fuzz
executions. The complete suite passed in 17.451s, the race suite in 89.482s,
vet is clean, and standalone-oracle audits pass. Inventory becomes 331
validated, 40 partial and 1,386 unstarted files.

## Bar progress facade execution checkpoint

Recognize `src/misc/bar/bar.c` as the same five-routine implementation as
`extraUtilProgress.c` and reuse the owned `ExtraProgressBar` engine through a
concrete `BarProgress` alias. Preserve source-shaped start, update and stop
functions, including disabled nil-handle no-ops. Do not duplicate the state
machine and do not add a writer interface or global output state.

Map `bar.h`'s enabled macro to the typed value `BarProgressUse`, its opaque
record to the concrete alias, and its inline threshold shortcut to the shared
update early return without exposing field layout. Pin the wrapper constant,
nil behavior, threshold fast path, exact completion frame and cleanup. Reuse
the independent 5,002-execution recurrence-model campaign because the C bodies
and Go engine are identical.

This validates `bar.c`, `bar.h`, and the manifest that selects only `bar.c`.
The complete suite passed in 17.504s, the race suite in 89.620s, vet is clean,
and standalone-oracle audits pass. Inventory becomes 334 validated, 40 partial
and 1,383 unstarted files.

## Glucose constant-contract execution checkpoint

Port the identical five-value surfaces in `sat/glucose/Constants.h` and
`sat/glucose2/Constants.h` once as typed constants plus a concrete value-return
configuration. Preserve the three enabled feature selectors, removal ratio
`2`, and blocking-restart threshold `10000`. Port `sat/glucose2/CGlucose.h`'s
sole selector as the exact typed value `Glucose2CircuitExperiment == 1`.

Pin every finite value and independent configuration copies with fixed Go-only
tests. Keep each remaining Glucose/Glucose2 file at its actual ledger status;
configuration-header completion must not be used to claim solver behavior.
Add no interface and no generated or native fixture.

This validates both constant headers and `CGlucose.h`. The complete suite
passed in 17.235s, the race suite in 88.512s, vet is clean, and standalone-
oracle audits pass. Inventory becomes 337 validated, 40 partial and 1,380
unstarted files.

## Extra packed bit-matrix execution checkpoint

Port all 16 routines in `misc/extra/extraUtilBitMatrix.c` into one concrete
owned packed matrix. Use `uint32` words to preserve source indexing exactly.
Keep upper/lower symmetric normalization, diagonal behavior, mutation/lookup
counters, raw row OR, simultaneous row union, strict-upper counting,
full-storage disjointness and the clique-equivalence predicate distinct. Return
the triangular rendering as bounded text instead of printing. Add no interface.

Provide copied row snapshots, checked cleanup and an explicit close lifecycle.
Preflight dimensions, allocation, work and output sizes; poll cancellation and
return no partial scan result or counter change on failure. Test exact packing,
both orientations, formatting, counters, row operations, disjointness, clique
positive/negative cases, ownership, cleanup, closure and all error classes.
Fuzz arbitrary relation mutations and queries against an independent Boolean
matrix model for at least 5,000 executions.

This validates `extraUtilBitMatrix.c`. The model campaign passed 5,000 Go-only
executions. The complete suite passed in 17.431s, the race suite in 89.283s,
vet is clean, and standalone-oracle audits pass. Inventory becomes 338
validated, 40 partial and 1,379 unstarted files.

## Constant multiplier-accumulator execution checkpoint

Port all ten surfaces in `misc/extra/extraUtilMacc.c` around one concrete owned
recipe plan. Preserve signed low-bit indexing, trivial/shift sentinels,
adder-level traversal and ordering, first-solution retention, packed add/sub
recipes and negative completion. Validate every coefficient by an independent
recipe traversal and cap the format at the source encoding's 15-bit operand
width. Return copied plan entries and add no interface.

Return bounded deterministic strings for direct multiplier/MACC specifications
and synthesized multiplier, narrow-accumulator and wide-accumulator modules.
Make timestamps explicit caller data. Replace the two file-writing test drivers
with ordered concrete generated-file rows and Go semantic tests. Deduplicate
shared declarations so generated modules remain usable. Check dimensions,
entry/work/output limits, cancellation and timestamp syntax atomically.

Exhaust all signed coefficients over several bit widths, pin exact text and
every template/zero mode, verify batch naming/order and ownership, and fuzz
plans against direct signed arithmetic for at least 5,000 executions without a
native oracle or retained generated corpus.

This validates `extraUtilMacc.c`. The arithmetic campaign passed 5,000 Go-only
executions. The complete suite passed in 17.165s, the race suite in 93.508s,
vet is clean, and standalone-oracle audits pass. Inventory becomes 339
validated, 40 partial and 1,378 unstarted files.

## Extra canonical BDD execution checkpoint

Port all 25 routines in `misc/extra/extraUtilMult.c` as a concrete owned BDD
manager. Preserve false/true literals, low-bit complementation, variable-node
IDs, wrapping hash constants, chained unique insertion, regular else-edge
normalization, direct-mapped cache replacement, numeric argument ordering,
canonical AND and De Morgan OR. Use transaction logs so node, unique-table,
cache and statistic changes roll back together on limit or cancellation.

Replace shared mark bytes with operation-local visitation while preserving
single, union-shared and per-root-summed counting modes. Return bounded exact
else-first SOP text, example results and concrete statistics rather than
printing. Make close explicit. Convert plain topological GIA AND graphs into
owned BDDs with complement phases and ordered outputs, returning optional SOPs
and the source driver's per-output summed count. Reject buffer/XOR/mux graphs
whose semantics the C loop would silently lose. Add no interface.

Exhaust the four-variable example and GIA bridge against direct Boolean
evaluation. Pin canonical identities, cache behavior, exact text, both count
modes, lifecycle, capacity, cancellation and transaction atomicity. Fuzz
arbitrary AND/OR/complement sequences against packed truth tables for at least
5,000 executions without native code or permanent generated data.

This validates `extraUtilMult.c`. The truth-table campaign passed 5,000 Go-only
executions. The complete suite passed in 17.358s, the race suite in 88.033s,
vet is clean, and standalone-oracle audits pass. Inventory becomes 340
validated, 40 partial and 1,377 unstarted files.

## Permanent stand-alone oracle policy checkpoint

Treat removal of native oracle generators as permanent, not as a temporary
porting convenience. No test may read the upstream `src` tree, invoke ABC, run
a compiler, use cgo, or depend on a regenerated native result. Keep permanent
inputs only under `test_vectors/`, hash-pin them in
`test_vectors/oracle/reference_build.json`, and make each dataset-backed test
self-justifying with an independent Go model, exhaustive bounded enumeration,
an algebraic or structural invariant, certificate replay, or a round trip.

Keep `pkg/tools/` on an exact allowlist containing only
`source_manifest.py`. That script is planning evidence for source identity, is
not an oracle, and is never run by the Go tests. Reject every additional tool,
even a Python script, until it has been shown not to restore a compiler, ABC
executable, source-generator, or refresh-from-C workflow. Continue rejecting
C/C++ sources and headers, symlinks, subprocess and generator hooks in tests,
and permanent `testdata/` directories across the whole module.

## Extra majority-network enumeration execution checkpoint

Port all 14 routines in `misc/extra/extraUtilMaj.c` as one concrete bounded
manager. Replace the eager ten-million-record allocation with owned dynamic
growth while retaining source indices and metadata. Hash canonical truth
tables after stretching them to the manager width so functions with different
declared input counts deduplicate exactly as they do in `Vec_Mem`. Add no
interface and do not expose borrowed table storage.

Preserve adjacent-symmetry group boundaries, group bit removal/insertion and
unpacking, cumulative one- and three-variable placement, and the source's
little-endian memory comparison. Implement expansion as selected-input MAJ3
substitution and reduction as tying the selected pair to one fresh input. Run
both through the validated permutation-only canonicalizer, then record groups,
predecessor, history and node count. Return the fully symmetric weight polar
and odd-input majority match as concrete data.

Replace printing, timing and truth-file dumping with bounded stage records,
optional owned table snapshots and diagnostics. Enforce object, truth-word,
work and output limits, cancellation and explicit close. Pin exact group
algebra, placement, MAJ3 expansion/reduction/deduplication, metadata,
formatting, enumeration, ownership and every error class. Fuzz arbitrary
expand/reduce histories for at least 5,000 executions against independent
direct substitution and exhaustive permutation-orbit membership.

This validates `extraUtilMaj.c`. The independent campaign passed 5,000 Go-only
executions. Inventory becomes 341 validated, 40 partial and 1,376 unstarted
files, with no native oracle or retained corpus.

## Extra memory-manager execution checkpoint

Port all 18 routines in `misc/extra/extraUtilMemory.c` through concrete fixed,
flexible and stepped arenas. Reuse the package's validated checked byte-arena
core for fixed free-list and flexible bump behavior instead of duplicating the
allocator. Preserve fixed chunk sizing, LIFO reuse, first-chunk restart,
persistent peaks and exact statistics. Configure the Extra flexible facade for
its source-specific 4-KiB chunk and twice-request growth rule.

Build the stepped facade from 8/16/... fixed classes with the exact
`4 << steps` map boundary. Preserve nil zero-byte allocation, retained large
blocks, no-op large recycling and the source's class-only memory-usage result.
Defer storage for high classes until fetch so an unused class does not consume
or prematurely fail the global byte budget. Use checked concrete block handles
that invalidate on fixed recycle/restart or close; reject foreign, duplicate or
wrong-sized recycling.

Return exact fixed/flexible statistics text instead of printing. Bound bytes,
live blocks and output, poll cancellation, and make close explicit. Test every
chunk threshold, multi-chunk/reuse/restart path, peak and accounting field,
exact formatting, flexible fill/resize/zero path, step class boundary, large
retention, high-class deferral, ownership, errors and lifecycle. Fuzz mixed
fixed/flexible/step allocation state against direct accounting for at least
5,000 executions without native code or permanent generated data.

This validates `extraUtilMemory.c`. The mixed allocator campaign passed 5,000
Go-only executions. Inventory becomes 342 validated, 40 partial and 1,375
unstarted files.

## Extra token-reader execution checkpoint

Port all nine routines in `misc/extra/extraUtilReader.c` as a concrete owned
reader with byte and file constructors. Replace the mutable 4-MiB pointer
window and reload relocation with an owned input buffer, but preserve token
group boundaries, character-class overwrite precedence, special `//` handling,
ordinary slash fallthrough, clean delimiters, empty-group skipping, one-based
line numbers, byte position, EOF behavior and unterminated-comment rejection.

Return copied `ExtraTokenBatch` values instead of pointers into mutable parser
storage. Reject NUL bytes rather than silently presenting truncated C strings.
Check file size around loading, cap input bytes, tokens, token bytes and parser
operations, poll cancellation, keep errors position-atomic and make close
explicit. Add no interface.

Test exact mixed-comment batches, overlapping classifications, slash cases,
empty groups, positions/lines, final tokens, file loading, copied ownership,
missing/oversized/NUL/unterminated input, every limit, cancellation and close.
Fuzz arbitrary bounded alphabets for at least 5,000 executions against an
independent parser model without native code or retained generated data.

This validates `extraUtilReader.c`. The parser-model campaign passed 5,000
Go-only executions. Inventory becomes 343 validated, 40 partial and 1,374
unstarted files.

## Extra support-minimization execution checkpoint

Port all 22 definitions in `misc/extra/extraUtilSupp.c` over copied `uint64`
rows. Preserve numeric M-out-of-N and even-weight mask order, first-occurrence
ordered unique XOR pairs, and the duplicate-preserving full difference matrix.
Use map membership only, never map iteration, so no ordered dictionary is
needed.

Represent the greedy parity tag as explicit owned Boolean state. Preserve
profile counts, first-maximum ties, strict merit improvement, selection,
filtering and matrix order. Reject zero difference rows rather than entering
the source's non-terminating outer loop. Return verification codes and pairwise
collision counts. Preserve the second solver's singleton protection,
minimum-double selection, tie order and variable removal, returning concrete
active/removed state and profiles.

Return bounded low-bit-first mask and profile text. Parse MIN input with the
exact marker, delimiters, literals, cube separators, INPUT stop, count and
width rules; provide byte/file entry points and a combined returned workflow
instead of print-only tests. Bound rows, pairs, work, input/output, poll
cancellation and copy every caller-visible slice. Add no interface.

Test exact combinatorial order, unique and duplicate differences, greedy state
transitions, deterministic matrix output, collision verification, reduction
ties/irredundancy, formatting, valid/malformed MIN inputs, ownership, limits
and cancellation. Fuzz unique cube sets for at least 5,000 executions, proving
linear-code injectivity and that every retained reduction variable is necessary
without native code or permanent generated data.

This validates `extraUtilSupp.c`. The independent campaign passed 5,000
Go-only executions. Inventory becomes 344 validated, 40 partial and 1,373
unstarted files.

## Extra SIS-utility execution checkpoint

Port all 13 definition surfaces in `misc/extra/extraUtilUtil.c` with concrete
caller-owned state. Replace global getopt variables with `ExtraOptionParser`,
copy its argument vector, and preserve clusters, attached/separate required
arguments, option indices, reset, `--`, non-option termination and structured
unknown/missing-argument results. Add no interface.

Preserve the 64-MiB fallback, exact millisecond formatting, nullable copied
strings and the source's deliberately disabled tilde expansion. Implement
file-mode checking and ordered colon-path search with exact dot handling,
`x`-as-read and fopen-compatible write side effects. Bound candidate count and
path bytes and poll cancellation. Replace assertion/exit allocation failure
and its mutable callback pointer with an ordinary typed error. Use portable Go
runtime user-CPU metrics for both native clock variants and retain the no-op
memory test.

Test every cluster/argument/index transition, reset and malformed option;
formatting boundaries; string ownership; all supported modes; missing files;
search order; create/truncate behavior; limits; cancellation; error policy and
CPU monotonicity. Fuzz at least 5,000 bounded argument vectors against a
separate Go parser model, without native code or permanent generated data.

This validates `extraUtilUtil.c`. The independent parser-model campaign passed
5,000 Go-only executions. The full suite passes in 17.113s, the race suite in
89.547s, vet is clean, and all stand-alone policy audits pass. Inventory
becomes 345 validated, 40 partial and 1,372 unstarted files.

## Extra truth-utility execution checkpoint

Port all 20 routines in `misc/extra/extraUtilTruth.c` through the existing
concrete owned `TruthTable` representation. Reuse its variable construction,
swaps, support, cofactors, quantification, mux, phase transform and symmetry
operations. Add thin source-specific helpers only for phase-mask
stretch/shrink, cofactor recombination, minimum cofactor-support overlap,
32-bit packed counts, the prime-weighted hash and the older semicanonical
policy. Add no interface.

Preserve increasing selected-variable order, first-minimum overlap ties, the
five-support cutoff, `32/-1` sentinel, replicated small-table count scale,
wrapping uint32 multiplication/XOR, balanced-output assignment-zero tie,
input-phase orientation and stable adjacent signature sorting. Return owned
tables plus both permutation directions and explicit phases so callers can
replay the result. Reject unsafe active-input shrink and replace fixed static
scratch with checked dynamic storage.

Test all widths around packed-word boundaries, elementary patterns, each swap
case, arbitrary phase embeddings, shrink validation, support, both cofactors,
quantifiers, mux reconstruction, direct/phased symmetry, overlap ties,
source-scaled populations, exact hash, semicanonical ties, ownership and every
error class. Fuzz at least 5,000 arbitrary truth tables against a separate
assignment model and replay all returned transformation metadata without
native code or generated fixtures.

This validates `extraUtilTruth.c`. The independent model campaign passed 5,021
Go-only executions. The full suite passes in 17.318s, the race suite in
87.950s, vet is clean, and all stand-alone policy audits pass. Inventory
becomes 346 validated, 40 partial and 1,371 unstarted files.

## Extra finite-ZDD truncation execution checkpoint

Port all public routines and private vector definitions in
`misc/extra/extraZddTrunc.c` as an immutable concrete finite-subset family.
Normalize each subset with sorted unique variables, deduplicate union results,
and sort families by increasing-index ZDD exclude-before-include traversal.
Represent a ZDD variable as the one-subset family containing that input. Add no
interface and retain no CUDD dependency.

Preserve the recursive truncator's product of included-variable probabilities,
strict below-threshold pruning, equality inclusion and traversal order. Return
both copied subset slices and the original leading-count, `-1`-terminated flat
encoding. Replace the private growable integer vector with ordinary owned Go
slices. Return exact print text and the built-in example workflow instead of
using stdout or initializing a native manager. Bound variables, subsets,
elements and encoded output; validate all probabilities; poll cancellation.

Test unsorted and duplicate elements, duplicate subsets, empty/singleton
families, canonical traversal, copied ownership, the exact source example,
flat encoding, text, equality boundaries, invalid dimensions/probabilities,
all limits and cancellation. Fuzz at least 5,000 bounded families against an
independent exhaustive bit-mask model whose bit-reversed ordinal order derives
ZDD traversal without reusing the implementation comparator.

This validates `extraZddTrunc.c`. The independent exhaustive campaign passed
5,000 Go-only executions. The full suite passes in 17.119s, the race suite in
89.851s, vet is clean, and all stand-alone policy audits pass. Inventory
becomes 347 validated, 40 partial and 1,370 unstarted files.

## Extra module-manifest execution checkpoint

Reconcile `misc/extra/module.make` after the first 12 of its 20 selected
translation units are validated. Record exactly those selected filenames and
leave `Canon`, `Cube`, `Dsd`, `Enum`, `File`, `Misc`, `Path` and `Perm` as
outstanding. Mark the manifest partial; do not treat Go package compilation as
proof that unported selected C behavior exists. Inventory becomes 347
validated, 41 partial and 1,369 unstarted files.

## Extra input-phase canonicalization execution checkpoint

Port both canonicalization routines, the generated lookup data and all three
diagnostic drivers in `misc/extra/extraUtilCanon.c` through an exhaustive
concrete phase-orbit operation. Support zero through six inputs and retain the
source rule that at least three phase bits are enumerated, even when some are
unused. Select the numerically smallest packed truth and return all minimizing
phase masks in deterministic ascending order. Add no interface.

Generate the 256 three-input rows on demand instead of copying static arrays.
Return owned truth/phase results rather than pointers into recursive static
storage. Reject active inputs beyond the declared real width, bound candidate
visits and poll cancellation. Replace print/test drivers with returned rows and
exhaustive Go verification.

Check all 256 three-input and 65,536 four-input functions against an independent
assignment-level phase transform. Test unused-input multiplicity, six-input
embedding, ownership, active-high rejection, limits, cancellation and invalid
dimensions. Fuzz at least 5,000 arbitrary zero-through-six-input functions,
checking the minimum and every returned phase without native code or fixtures.

This validates `extraUtilCanon.c`. The exhaustive three/four-input campaigns
and 5,000-case model fuzzer pass. The Extra manifest reaches 13 of 20 selected
units. The full suite passes in 17.802s, the race suite in 90.669s, vet is
clean, and all stand-alone policy audits pass. Inventory becomes 348 validated,
41 partial and 1,368 unstarted files.

## Extra cube-group enumeration execution checkpoint

Port both enumerators and all private helpers in `misc/extra/extraUtilCube.c`
as concrete 24-sticker permutations. Preserve the three literal quarter-turn
swap lists, their first/second/third powers, `result[i] = state[perm[i]]`, exact
first-nine seeding, later generator order, BFS frontiers and convergence. Add
no interface.

Offer full-permutation and seven-corner-signature deduplication modes. Preserve
the signature's selected positions, corner/orientation maps and five-bit field
packing, but use a sparse lookup-only map instead of a 4-GiB bitmap. Return
layer totals, optional copied states and exact state text. Bound states/depth,
poll cancellation and return structured errors rather than printing and
abandoning fixed storage.

Test generator fourth powers, permutation validation, signatures, formatting,
ownership, limits and cancellation. Reproduce published layer totals in both
modes, retain an opt-in complete 3,674,160-state regression, and fuzz at least
5,000 bounded group walks/full-vs-signature enumerations while proving every
state is a permutation and signatures remain collision-free.

This validates `extraUtilCube.c`. The full group converged to 3,674,160 states
in 7.909s and the differential campaign passed 5,000 Go-only executions. The
Extra manifest reaches 14 of 20 selected units. Inventory becomes 349
validated, 41 partial and 1,367 unstarted files. The full suite passes in
17.588s, the race suite in 91.028s, vet is clean, and all stand-alone policy
audits pass.

## Extra file/text utility execution checkpoint

Port all 33 routines in `misc/extra/extraUtilFile.c` as concrete owned name,
filesystem, binary/hex and text operations. Preserve last-dot rules, static
buffer results as newly owned strings, `>`/backslash path correction,
basename/path/design extraction, strict suffix matching and ordered similar
name probing. Use concrete `*os.File` parameters rather than an interface.

Preserve file reads' unconditional newline, uint32 binary wrapping, both bit
orders, little-word/right-to-left hex parsing, every case/prefix/reverse format
and special 0/1-variable strings. Reject the old print-hex undefined-shift
widths. Return timestamp/symbol/string output. Preserve sorting's CR and empty
line behavior. Simulate 999-byte `fgets` and the line-number routine's
`strlen-2` replacement, but reject memory-unsafe chunks. Add byte/line/output
limits, cancellation, metadata race checks and explicit errors.

Test every path edge, suffix order, empty/missing/combined file, appended
newline, exact binary and hex form, malformed digit, sorting, CRLF/LF numbering
quirks, long chunks, ownership, every limit and cancellation. Fuzz at least
5,000 binary/hex/string cases against round-trip and direct filtering models,
without native code or generated fixtures.

This validates `extraUtilFile.c`. The independent campaign passed 5,000
Go-only executions. The Extra manifest reaches 15 of 20 selected units.
The full suite passes in 17.686s, the race suite in 90.410s, vet is clean, and
all stand-alone policy audits pass. Inventory becomes 350 validated, 41 partial
and 1,366 unstarted files.

## Extra function-enumeration execution checkpoint

Port the three fixed problems and both engines in `misc/extra/extraUtilEnum.c`.
Represent row minterms, transposed variable columns and targets as owned uint32
slices. Preserve the first engine's cost boundaries, `a+b+1` decomposition,
operand ordering, AND-before-OR insertion, global truth deduplication, target
order and recursive expression formatting.

Represent the second engine with one concrete bounded manager retaining all
fanin, phase, polarity, reference, top, level and uint64 truth arrays. Preserve
same-level/new-level recursion, reference feasibility, NPN skips, four AND
polarities, optional XOR, complement equivalence and full cheaper-pair filter.
Return counters and optionally copied completed networks instead of printing.
Use lookup-only maps, add no interface, bound attempts/nodes/builds/finished
results and poll cancellation.

Test literal fixed-data transposition and a substantial built-in-problem layer
prefix; exact simple AND/OR expressions; AND-only and XOR primitive counts;
network reevaluation, single-top property, feasibility, deterministic counters,
ownership, limits and cancellation. Fuzz at least 5,000 arbitrary small truth
problems and two/three-input structural searches, independently reevaluating
every node and completed network without native code or fixtures.

This validates `extraUtilEnum.c`. The second fixed problem reaches 116,764
unique truths at five gates, and the independent campaign passed 5,000 Go-only
executions. The Extra manifest reaches 16 of 20 selected units. Inventory
becomes 351 validated, 41 partial and 1,365 unstarted files. The full suite
passes in 17.930s, the race suite in 91.548s, vet is clean, and all stand-alone
policy audits pass.

## Extra grid-path execution checkpoint

Port all 19 definitions in `misc/extra/extraUtilPath.c` as concrete graph,
frontier, exact-count, GIA and SAT-result values. Preserve the monotone grid's
horizontal/vertical variable IDs and dynamic program. Preserve the undirected
grid's edge order, first/last lifetimes, live-frontier order, packed
degree/component states, endpoint/internal degree constraints, cycle rejection,
component renaming, state insertion order and terminal encodings. Use exact
integer path counts instead of accumulating doubles, and add no interface.

Derive the path GIA through memoized muxes over edge variables. Return path
drawings and suggested AIG filenames rather than printing or writing files.
Recreate the SAT experiment with deterministic lexicographic model selection,
blocking clauses and returned sampled assignments/statistics. Define the
zero-span case, reject more than 64 inputs where the source's packed minterm is
not representable, and bound states, transitions, objects, SAT storage and
operations with cancellation.

Test exact edge/lifetime/frontier order and known path counts. Exhaustively
compare every 2x2 and 3x3 assignment to an independent Go simple-path
recognizer and to the derived GIA. Exhaust the monotone GIA through size two.
Pin rendering, filenames, lexical SAT models, limits, cancellation and errors.
Fuzz at least 5,000 bounded grid assignments against the standalone predicate,
without C code, native executables or retained oracle data.

This validates `extraUtilPath.c`. Counts 2, 12 and 184 are pinned through 4x4,
the independent campaign passes 5,000 executions, and the Extra manifest
reaches 17 of 20 selected units. Inventory becomes 352 validated, 41 partial
and 1,364 unstarted files. The full suite passes in 18.042s, the race suite in
90.815s, vet is clean, and all stand-alone policy audits pass.

## Extra permutation-ZDD execution checkpoint

Port the canonical manager and every active operation in
`misc/extra/extraUtilPerm.c`. Preserve terminal/variable IDs, uint32 hash
arithmetic, zero-suppressed unique creation, direct-mapped cache replacement,
node creation order and lexicographic transposition-variable mapping. Use a
concrete manager with bounded nodes, cache and operations; make exported
failures transactional and return exact path counts. Add no interface.

Preserve the literal difference, union, minimal-union, intersection,
cofactor, threshold, dot-product, size-six minimal-product, permutation and
permutation-product recurrences, including observable source quirks. Return
owned traversed subsets, permutations, transpositions and formatted text.
Replace the GIA and cube print drivers with structured bounded results.

Test canonical identities and node IDs, exact families, path/shared-node
counts, source intersection behavior, cofactors, thresholds, products,
selection-sort conversion, reverse reconstruction and group composition.
Reproduce the cube's early cumulative path and node sequences. Test ownership,
limits, cancellation and atomicity. Fuzz at least 5,000 arbitrary small
families against independent explicit set models without native code or
fixtures.

This validates `extraUtilPerm.c`. The explicit-family campaign passes 5,000
executions; cube path totals 1/10/64/385 and node totals 0/63/443/2018 match.
The Extra manifest reaches 18 of 20 selected units. Inventory becomes 353
validated, 41 partial and 1,363 unstarted files. The full suite passes in
17.725s, the race suite in 94.361s, vet is clean, and all stand-alone policy
audits pass.

## Extra six-variable DSD execution checkpoint

Port the complete active `misc/extra/extraUtilDsd.c` manager without retaining
its missing `dsdfuncs6.dat` runtime dependency. Preserve all 595 class records
in a permanent `test_vectors/` table and embed that table in the package.
Validate its shape and semantics during concrete manager construction.

Recreate the source permutation and Gray schedules, partial packed-permutation
completion, sequential class permutations, output normalization, exact DSD
configuration lookup, AND/XOR composition, phase handling, cut rewriting,
statistics, CNF costs and divisor collection. Derive configurations from exact
truth transformations instead of shipping the multi-megabyte native hash
file. Replace singleton lifecycle and print-only drivers with owned managers
and structured results; add limits and cancellation, with no interface.

Test all 595 direct representatives and supports, schedule cycles, partial
permutations, arbitrary transformed representatives, configuration replay,
compute/cut truth, statistics, costs, divisors and errors. Fuzz at least 5,000
class/permutation/phase/output combinations and reconstruct every answer
without C code or generated native artifacts.

This validates `extraUtilDsd.c`. The independent replay campaign passes 5,000
executions. The Extra manifest reaches 19 of 20 selected units. Inventory
becomes 354 validated, 41 partial and 1,362 unstarted files. The full suite
passes in 17.691s, the race suite in 93.829s, vet is clean, and all standalone
policy audits pass.

## Extra miscellaneous utility execution checkpoint

Port every active definition in `misc/extra/extraUtilMisc.c` while sharing the
already validated Gray schedule, adjacent permutation schedule, six-variable
bit transforms and 64-by-64 transpose. Use checked concrete functions with
owned results for logarithms, powers, combinations, radix digits, popcount,
factorials and the source's recursive permutation order. Add no interface.

Preserve compact minterm permutation in both directions, input polarization,
and brute-force N, NN, P, NP and NPN minima. Return owned complete phase tables
and a complete four-variable NPN classification with replayable phase and
permutation metadata. Replace the literal 4/5/6/8-variable support tables and
their stdout generator with one algorithmic support compactor, retaining the
source direction and singleton-variable quirks. Return the eight elementary
truth tables as data and preserve stable equal-cost sorting.

Preserve exact six-variable minimization over two output phases, all 720
permutations and all 64 Gray phases. Preserve the popcount-selected heuristic
and ten-round adjacent-pair descent. Replace native files and print drivers
with a strict sixteen-hex-digit parser, bounded power rows, bit strings and
matrix strings. Preserve the reversed bit convention of the raw 32-by-32
primitive and the conventional semantics of the rectangular wrappers. Replace
the C contiguous-row allocator with ordinary owned slices and consolidate
lookup constructors, comparison callbacks, pointer variants and self-test
drivers into the operations they support.

Test exact recursive permutation order, arithmetic ownership and errors,
permutation reversal, polarization, all 3-input N classes, all 65,536
four-input NPN rows by metadata replay, all support masks through seven inputs,
stable sorting, exact six-input orbit invariance, parsing, formatting, power
quirks and every transpose bit. Fuzz at least 5,000 arbitrary truth/phase/
permutation/transpose combinations against standalone Go models without C
code, generated native fixtures or retained fuzz output.

This validates `extraUtilMisc.c` and completes all 20 selected Extra
translation units. The independent model campaign passes 5,000 executions.
Inventory becomes 356 validated, 40 partial and 1,361 unstarted files. The
full suite passes in 18.263s, the race suite in 94.263s, vet is clean, and all
stand-alone policy audits pass.

## Header-only hash-map execution checkpoint

Port the five headers under `misc/hash` as concrete typed maps inside package
`abc`; the empty `module.make` selects no separate native object. Preserve the
default integer hash with explicit 32-bit overflow, the string hash with
uint32 base-6 accumulation, long-key remainder hashing, collision-chain tail
insertion, overwrite without reordering, conditional zero creation, removal,
size and bin-then-chain iteration order.

Use separate int, float32, opaque `uintptr`, string and long-key maps. Support
the general source callback mode with concrete hash/comparison function types,
not an interface, and retain the originally inserted key when a later
equivalent key overwrites its data. Immutable strings replace optional native
key freeing. Use checked concrete cells for the source APIs that returned
mutable pointers; removal and close must invalidate them. Return owned entry
snapshots, enforce entry limits and reject invalid callback bins.

Test 32-bit overflow hashing at signed extremes, string hashing, collisions,
chain order, overwrite/create behavior, all data types, custom equivalence
callbacks, removal return values, ownership, limits, close and stale handles.
Fuzz at least 5,000 arbitrary operation streams against independent int and
float Go models. Use no interface, native helper, generated fixture or Go-map
iteration as an ordering oracle.

This validates `hash.h`, `hashInt.h`, `hashFlt.h`, `hashPtr.h`, `hashGen.h` and
their header-only manifest. The state-machine campaign passes 5,000
executions. Inventory becomes 362 validated, 40 partial and 1,355 unstarted
files. The full suite passes in 18.269s, the race suite in 95.542s, vet is
clean, and all stand-alone policy audits pass.

## Memory-manager execution checkpoint

Consolidate `misc/mem/mem.c` and `mem.h` with the validated concrete byte-arena
core already used by Extra and SAT. Preserve fixed chunk geometry, LIFO
recycling, first-chunk restart, peak counts, flexible 4-KiB and twice-request
growth, requested-byte usage, stepped power-of-two classes, large-block
retention and class-only memory reporting. Add the missing flexible restart,
invalidating every previous block and using the retained chunk's real extent
when source chunk growth would otherwise make restart overrun it.

Port the distinct `mem2.h` page allocators as concrete flex, fixed and stepped
handle managers. Preserve page-mask encoding, offset one, page-boundary rules,
0xFF initialization, fixed handle order, LIFO reuse, packed step size tags,
current/peak/total counters and word memory totals. Preserve the unusual flex
release rule that frees the preceding page and the numeric handle alias after
recycle/reuse. Replace returned word pointers with checked read, write and
owned snapshot operations. Reject a geometry that cannot create the mandatory
offset-one entry instead of reaching the native empty free-list pop.

Test exact handle sequences, fill and reuse, page release, packed routing,
statistics, byte-arena restart, ownership, invalidation, limits, cancellation
and lifecycle. Fuzz at least 5,000 arbitrary fixed fetch/write/recycle streams
against an independent handle/value model without native code or retained
fixtures.

This validates `mem.c`, `mem.h`, `mem2.h` and their one-unit manifest. The
fixed-handle state-machine campaign passes 5,000 executions. Inventory becomes
366 validated, 40 partial and 1,351 unstarted files. The full suite passes in
18.305s, the race suite in 94.220s, vet is clean, and all stand-alone policy
audits pass.

## Ranked Boolean parser execution checkpoint

Port `misc/parse/parseCore.c` and its public declaration as a concrete bounded
parser returning the existing owned packed `TruthTable`; do not retain the
CUDD dependency or introduce an interface. Preserve constants, both negation
forms, implicit/explicit AND, XOR, OR, equivalence, forward/reverse
implication, source precedence, left associativity and exact-name lookup.
Represent each rank as another variable block, with bare names in the highest
block and each bracket pair lowering a complete subexpression one block.

Use dynamically sized typed stacks rather than the native fixed stack. Reject
mismatched delimiter kinds and rank underflow explicitly. Return errors rather
than printing diagnostics, own all returned data, and bound tokens, nesting,
nodes and truth words with cancellation throughout parsing and evaluation.

Exhaust every operator on all small assignments and independently model
ranked expressions over all 64 six-variable assignments. Pin precedence,
implicit multiplication, constants, whitespace, duplicate names, ownership,
malformed relations/delimiters, underflow, limits and cancellation. Generate
at least 5,000 bounded formulas from fuzz bytes and compare every assignment
to a separate closure model. Retain no generated corpus and invoke no native
helper.

This validates `parseCore.c` and `parse.h`, completing the parse directory.
The independent semantic campaign passes 5,000 executions. Inventory becomes
368 validated, 40 partial and 1,349 unstarted files. The full suite passes in
18.406s, the race suite in 96.117s, vet is clean, and native-source,
package-header, `testdata/` and tool-allowlist audits pass.

## AVL-tree execution checkpoint

Port `misc/avl/avl.c`, `avl.h` and the documented API as one concrete generic
AVL tree using a comparator function, not a runtime interface. Preserve
duplicate insertion down the right branch, first-match lookup/deletion,
rightmost-predecessor replacement, height/balance maintenance, first/last,
find-or-add value slots, counts, both traversal directions, generators,
callbacks, integer comparison and invariant checking.

Replace the fixed 50-entry native path arrays with balanced recursive or
dynamic traversal. Make value slots checked and invalidate them on deletion
or close. Use owned generator snapshots so later insertions are excluded and
deletion during generation is memory-safe. Keep optional deterministic
postorder close callbacks for callers needing explicit cleanup, while relying
on Go ownership by default. Bound entries, support cancellation and return
errors rather than printing diagnostics.

Test all four rotation forms, deep ordering, duplicates, repeated deletion,
slots, first/last, forward/backward generation, traversal mutation,
postorder close, stale handles, limits, cancellation and injected invariant
failures. Fuzz at least 5,000 operation streams against an independent
key-to-value-multiset model and check balance, heights, ordering and counts
after every operation without native code or retained fuzz output.

This validates all three AVL inventory entries. The independent multiset
campaign passes 5,000 executions. After the later ledger correction described
below, inventory becomes 369 validated, 40 partial and 1,348 unstarted files.
The full suite passes in 18.519s, the race suite in
94.943s, vet is clean, and all standalone policy audits pass.

## TIM inventory-reconciliation checkpoint

Close the public `misc/tim/tim.h` and `module.make` entries only after every
selected implementation unit is independently validated and every public
declaration maps to the concrete owned manager surface. Re-audit construction,
boxes, persistence, duplication/transforms, configuration, timing propagation,
traversal, counts, delay tables and deterministic diagnostics; use Go
ownership for Stop/StopP and owned results for pointer/stdout APIs.

The audit finds no missing declaration or selected source: all five manifest
units and the internal header were already validated, and the implementation
and its model/fuzz evidence cover the complete public header. The two stale
inventory entries are therefore validated without adding a duplicate facade
or interface. After the later ledger correction described below, inventory
becomes 371 validated, 40 partial and 1,346 unstarted
files; existing package gates remain applicable because executable code is
unchanged.

## Binary-BLIF execution checkpoint

Port `misc/bbl/bblif.c`, its public header and one-unit manifest as one concrete
owned network manager. Preserve object creation order and declared fanin order,
CI/CO constraints, internal SOP functions, source selection-sort cube order,
exact canonical SOP deduplication, marks, statistics, original IDs and
bidirectional object/fanin iteration. Replace borrowed objects with owned views
and correct the aborting native original-ID accessor by returning retained data.

Implement the exact pinned binary stream as three little-endian
length-prefixed vectors. Preserve the 12-byte object header, packed CI/CO/box/
mark/fanin fields, signed relative object offsets, word-counted padded SOP
records and read-time sequential ID rewrite. Decode only complete bounded
streams; reject malformed handles, offsets, flags, padding, topology and
trailing bytes. Return binary/text data instead of writing paths. Return
structured statistics and a checked half-adder manager instead of stdout/demo
side effects. Bound objects, IDs, fanins, aggregate SOP/binary storage and
truth words with cancellation.

Preserve the 32-bit truth/SOP helpers through sixteen variables: low-bit
variable order, small-table stretching, constants, minterm enumeration and the
source's final-cube off-set complement convention. Test a hand-derived exact
binary stream, read renumbering and stable re-encoding, forward/backward fanin
offsets, deduplication, text, ownership, malformed construction, limits,
cancellation and every truncation. Fuzz at least 5,000 truth round trips and
5,000 arbitrary byte streams without native code or retained fuzz output.

This validates all three BBL entries. The two independent campaigns pass 5,000
executions each. A source-path audit also moves five misplaced AVL/TIM records
onto their intended files and corrects three older false AIG promotions while
assigning the finished DSD work to `extraUtilDsd.c`. Inventory becomes 374
validated, 40 partial and 1,343 unstarted files. The full suite passes in
18.518s, the race suite in 94.696s, vet is clean, and all standalone policy
audits pass.

## Packed extended-float and MVC-header execution checkpoint

Port the header-only positive extended-float formats in
`misc/util/utilFloat.h` and `utilDouble.h` as concrete checked 32-bit and
64-bit values. Preserve the 16-bit exponent fields, normalized 16-bit or
48-bit mantissas, packed integer ordering, all named constants, IEEE-754
conversion truncation, source addition alignment, overflow saturation and
underflow-to-one division. Preserve the 64-bit source's four-term split
multiplication exactly; do not substitute ordinary floating-point arithmetic
or a differently rounded full-width product.

Test exact bit patterns, constant values, round trips, representability
boundaries, invalid encodings, wide exponent differences, carry
normalization, overflow and division underflow. Pin the source's exact
`1e100` value. Fuzz at least 5,000 input pairs against independent integer and
`big.Int` arithmetic models, with no native helper and no retained generated
corpus.

Close `misc/mvc/mvc.h` only after cross-checking every declaration and macro
against the source files actually present in this snapshot and the complete
concrete `misc_mvc_*.go` implementation. Do not invent APIs for declarations
whose named implementation units are absent. Record those declarations as
historical inventory-only surface, while requiring all source-backed manager,
cover, cube, list, packed-bit and algorithm declarations to remain covered by
the existing unit and fuzz tests.

This validates both extended-float headers and changes the final MVC header
from partial to validated. The independent float campaign passes 5,000
executions. Inventory becomes 377 validated, 39 partial and 1,341 unstarted
files. The full suite passes in 18.521s, the race suite in 93.824s, vet is
clean, and all standalone policy audits pass. `pkg/tools/` retains only the
source-manifest script.

## Utility memory-recorder execution checkpoint

Port `misc/util/utilMem.c` and `utilMem.h` as one explicit concrete recorder,
not as writable package globals. Replace raw pointer identity with a stable
caller-assigned unsigned identity. Preserve inactive pass-through recording,
allocation/free event accumulation, numeric sorting, duplicate multiset
pairing, start/quit state, the interrupt flag and expression-friendly returned
identities. Bound each recording window by an explicit event limit.

Do not reproduce the source's undefined recycle implementation: it asserts
both vectors are null and then dereferences them, while its final size
assertions contradict successful allocation/free pairing. Validate the
intended multiset operation before committing any state, reject a free without
a matching allocation, and return sorted unreleased identities in owned
storage. This lets a caller perform concrete cleanup or diagnostics without
attempting to manually free Go memory. Normalize the header/source
Start/Quit/Recycle naming mismatch as methods on the recorder.

Test inactive and active operation, duplicate IDs, arbitrary order, numeric
sorting, ownership, unmatched-free atomicity and retry, empty recycle windows,
event limits, lifecycle errors and interrupt transitions. Fuzz at least 5,000
bounded event streams against an independent allocation-count multiset without
native code or retained fuzz output.

This validates both utility-memory inventory entries. The independent campaign
passes 5,000 executions. Inventory becomes 379 validated, 39 partial and 1,339
unstarted files. The full suite passes in 18.513s, the race suite in 94.070s,
vet is clean, and the standalone-policy gates pass.

## Utility name-table execution checkpoint

Port `misc/util/utilNam.c` and `utilNam.h` as a concrete append-only unique
string table distinct from the typed, deletable `misc/nm` manager. Reserve ID
zero and assign every nonempty name a stable ID in first-insertion order.
Preserve exact and bounded-prefix find/intern behavior, ID-to-name access,
ordered iteration, cross-manager ID mapping, duplicate-sensitive common-name
counts and first-unique input order.

Keep an owned ordered name slice as the source of observable iteration. A plain
map may replace the internal chained bins because it is lookup-only and is
never enumerated. Implement both public source hashes exactly with uint32
overflow, signed-byte promotion, the fixed 128-prime cycle, Jenkins mixing and
their distinct bounded/NUL stopping rules. Return logical owned storage
statistics and deterministic text rather than exposing Go allocation details
or writing a caller path.

Use Go lifetime and Close instead of manual reference counting. Replace the
borrowed formatting buffer with owned strings; callers should format a string
and then call Intern rather than introducing an exported any-typed variadic
method. Bound names and total stored bytes, reject empty/embedded-NUL names,
and return errors for invalid IDs, prefixes, limits and closed state.

Test constants and hash vectors, first insertion and repeated insertion,
prefixes, lookup misses, ID zero, ordered snapshots, ownership, rendering,
storage counts, self/cross ID maps, repeated-ID common reports, first-unique
order, limits and lifecycle. Fuzz at least 5,000 bounded operation streams
against an independent ordered string-to-ID model, including cross-table maps,
without native code or retained fuzz output.

This validates both utility-name inventory entries. The independent campaign
passes 5,000 executions. Inventory becomes 381 validated, 39 partial and 1,337
unstarted files. The full suite passes in 18.057s, the race suite in 94.589s,
vet is clean, and the standalone-policy gates pass.

## Binary bridge protocol execution checkpoint

Port all active behavior in `misc/util/utilBridge.c` as concrete owned package,
netlist and result values. Preserve the exact 24-byte decimal package header,
all message type/value constants, AIGER unsigned integers, bridge const-one
numbering, CI/gate numbering, ordinary AND fanins, latch-driver encoding,
property count and complemented property outputs. Reject GIA buffer/XOR/mux
extensions that the source wire grammar cannot represent.

Implement the incoming body grammar independently from the outgoing body. The
source serializer ends immediately after properties, whereas its parser
requires version one, zero fairness and zero constraints before constructing
register inputs. Preserve this asymmetry explicitly. Parse the three-package
input sequence, retain owned envelope packages for diagnostics, and parse the
middle payload into an independently owned checked GIA plus latch-initial
values.

Implement holds, unknown and concrete counterexample result messages with
exact status/vector bytes, property/depth varints, frame-major inputs and the
source's zero initial-register vector. Bound package bytes, graph objects,
inputs, outputs and counterexample allocation before construction. Replace
assertions, short reads, process diagnostics and emergency exit with typed
errors; return bytes instead of mutating a `FILE*`.

Test exact package/body/result vectors, payload ownership, every truncation,
invalid headers, varint overflow, malformed fanins and trailers, limits,
unsupported extended nodes and three-package trailing data. Compare a parsed
sequential fixture to its source graph over every input/state assignment. Fuzz
at least 5,000 arbitrary byte strings through both package and netlist parsers,
requiring every success to pass graph validation and stable package re-encode,
without native code or retained output.

Also disposition `misc/util/utilFile.c` without adding redundant wrappers.
Reuse the atomic prefix/suffix temporary-file implementation already shared
with `utilSignal.c`; preserve exclusive mode-0600 read/write behavior while
removing the unlink/reopen race. Use ordinary owned Go formatting at future
call sites instead of exposing C-style variadic allocation, and omit the
private platform clock and commented demonstration because they carry no
independent behavior.

This validates the bridge and utility-file entries. The bridge parser campaign
passes 5,045 executions. Inventory becomes 383 validated, 39 partial and 1,335
unstarted files. The full suite passes in 18.187s, the race suite in 95.263s,
vet is clean, and the standalone-policy gates pass.

## Utility ISOP/ESOP execution checkpoint

Complete `misc/util/utilIsop.c` by composing its recursive ISOP entry points
with the existing generalized truth-table implementation and adding the
remaining concrete utility operations. Preserve OR/XOR cover construction and
optional complementation, recursive ESOP selection with lexicographic
cube/literal cost, the exact equal-cost drop order, greedy interval expansion,
the source's whole-tail cube comparison, and selector-tagged positive/negative
ISOP-CNF construction. Keep the native two-bit-per-variable cube encoding and
return owned truth tables and cube slices.

Bound variables, cubes, elements and visits before allocation, check malformed
cube digits and interval containment, and support cancellation. Return typed
errors rather than asserting. Use deterministic string formatting rather than
stdout and concrete structs rather than interfaces. Do not broaden a uint32
cube beyond the representable selector width merely to mimic undefined native
shifts.

Exhaust all 256 three-variable ESOP functions, all 6,561 three-variable
on/allowed intervals and every supported three-variable CNF relation. Pin the
source-sensitive cost/tie order and test ordinary, exclusive and complemented
covers, malformed cubes, unsupported widths, ownership, cancellation and each
resource limit. Fuzz bounded truth tables against direct Go evaluation for at
least 5,000 executions. Reuse the 196 permanent recursive-ISOP records under
`test_vectors/`; never compile C or retain native output as an oracle.

Disposition the surrounding non-runtime utility files at the same boundary.
`abc_namespaces.h` is replaced by the package namespace because all of its
definitions are C/C++ preprocessor linkage macros. `util_hack.h` maps its null,
CPU-time, soft-limit and allocation-failure aliases to Go `nil` and the
validated concrete extra-utility operations; omit the writable global
out-of-memory callback. Validate `module.make` only after each of its eight
selected translation units has an independently validated ledger entry.

This validates the utility ISOP source, namespace header, compatibility header
and build manifest. The standalone campaign passes 5,000 executions. Inventory
becomes 387 validated, 39 partial and 1,331 unstarted files. The full suite
passes in 18.468s, the race suite in 93.673s, vet is clean, and the
native-source, package-header, `testdata/`, symlink and tool-allowlist audits
pass. `pkg/tools/` contains only `source_manifest.py`.

## Integer pair/triple hash execution checkpoint

Port `misc/vec/vecHash.h` as one concrete arity-fixed integer tuple interner.
Pair managers key on `(Data0, Data1)` and retain `Data2` as mutable payload;
triple managers key on all three values. Preserve stable one-based IDs in
first-insertion order, the first payload on duplicate pair insertion, exact
uint32 hash overflow, source prime-sized growth, collision-chain order, direct
ID-to-entry access, pair payload increment/decrement/set operations and bin
profile order.

Fix arity at construction and reject pair/triple API mixing. In particular,
never expose triple `Data2` through the pair mutation API because that would
move a key without rehashing it. Use ordinary shared Go pointers and garbage
collection instead of manual Ref/Deref; provide a checked Close operation for
callers that need explicit lifecycle. Return owned entry and profile slices
instead of borrowed vector slots.

Require source-width signed 32-bit key/payload values and preserve hash
arithmetic by explicit uint32 conversion. Bound entries and bins, check prime
growth before allocation, make resizing atomic, support cancellation and
replace assertions with typed errors. Preserve the actual duplicate-insertion
behavior—return the existing positive ID—even though the header's local
demonstration incorrectly asserts that this result is zero.

Test exact pair/triple hashes across signed boundary values and many table
sizes; stable IDs; duplicate payload retention; triple distinction; old/new
counter return conventions; mutation; lookup; forced resizing; chain profiles;
owned snapshots; invalid IDs and arity; overflow boundaries; limits;
cancellation; validation and lifecycle. Fuzz at least 5,000 bounded operation
streams against independent insertion-ordered pair and triple models without
native code or retained fuzz output.

This validates the complete vector hash header. The combined model campaign
passes 5,000 executions. Inventory becomes 388 validated, 39 partial and 1,330
unstarted files. The full suite passes in 18.542s, the race suite in 93.903s,
vet is clean, and the native-source, package-header, `testdata/`, symlink and
tool-allowlist audits pass. `pkg/tools/` contains only `source_manifest.py`.

## Managed attribute-vector execution checkpoint

Port `misc/vec/vecAtt.h` as a generic concrete typed attribute vector. Use
`[]*T` storage and typed manager/start/free function callbacks; generic
constraints remain compile-time only and must not introduce runtime interface
storage or dispatch. Preserve all 15 attribute tags, zero-size default
capacity, exact doubling-or-index-plus-ten growth, direct writes when no
starter is configured, lazy entry construction, deletion, clearing, manager
access and retained-versus-freed manager shutdown.

Return owned snapshots rather than exposing the internal mutable slot slice.
Retain pointed-to object identity, and confirm that no active source caller
depends on direct `Vec_AttArray` mutation. Preserve the source asymmetry in
which FreeEntry invokes the object callback only with a nonnil manager while
Clear and Free invoke it regardless. Reject a manager-free request without a
configured callback rather than calling a null function pointer.

Check negative indices and capacities, bound growth before allocation, support
cancellation and return typed errors in place of assertions. Preserve direct
write overwrite semantics and repeated lazy starts only when a starter returns
nil. Use ordinary Go ownership rather than untyped pointers.

Test every attribute constant, both initial capacity paths, exact explicit and
automatic growth, sparse direct writes, lazy memoization, deletion/recreation,
callback order, nil-manager behavior, manager retention/freeing, overwrite,
snapshot ownership, limits, cancellation and closed state. Fuzz at least 5,000
bounded operations against an independent slot/capacity model without native
code or retained fuzz output.

This validates the complete attribute-vector header. The slot-model campaign
passes 5,000 executions. Inventory becomes 389 validated, 39 partial and 1,329
unstarted files. The full suite passes in 18.462s, the race suite in 93.387s,
vet is clean, and the native-source, package-header, `testdata/`, symlink and
tool-allowlist audits pass. `pkg/tools/` contains only `source_manifest.py`.

## Paged primary/shadow word-arena execution checkpoint

Port `misc/vec/vecSet.h` as one concrete paged word arena, not as a mathematical
set or a generic storage interface. Preserve page sizes of `2^n` uint64 words,
the two per-page primary/shadow metadata positions, page/shift handle encoding,
stable record handles, signed 32-bit integer packing, strict page rollover,
explicit entry counts, restart/reuse and both current cursors.

Implement primary integer append, zeroed byte fetches, string saving, primary
and shadow shrink, shadow reservation, shadow-limit commit, direct checked word
writes and overlap-safe record copying. Replace the iteration macro with
concrete fixed- or callback-sized handle collection. Provide the source's
handle/current/shadow/all-page memory calculations and its 64-bit-ABI report
measure. Constructor and Close cover both embedded and heap C allocation forms.

Return owned word/int/byte snapshots rather than interior pointers. Define
little-endian two-int-per-word packing explicitly. Zero unwritten storage and
odd trailing halves rather than preserving uninitialized bytes. Reject a
record that reaches the page's final word because the source's weak assertion
allows a current handle that aliases offset zero of the next page. Bound pages,
entries and bytes before allocation, support cancellation, make rollover
atomic and use typed errors.

Test exact handle geometry, packing boundaries, zero-length records, rollover,
source-shaped memory totals, bytes and embedded-NUL strings, mutable writes,
owned snapshots, fixed and variable iteration, both cursors, overlapping
compaction, entry-count rewrite, shrink, restart, reuse, stale metadata reset,
invalid handles, limits, cancellation and Close. Fuzz at least 5,000 append and
select/compact/replay streams against an independent model without native code
or retained fuzz output.

This validates the complete paged-set header. The combined primary/shadow model
campaign passes 5,000 executions. Inventory becomes 390 validated, 39 partial
and 1,328 unstarted files. The full suite passes in 18.126s, the race suite in
93.475s, vet is clean, and the native-source, package-header, `testdata/`,
symlink and tool-allowlist audits pass. `pkg/tools/` contains only
`source_manifest.py`.

## Vector hash-interner execution checkpoint

Port all three manager families in `misc/vec/vecHsh.h` as concrete Go types.
The single-integer manager interns signed int32 values. The fixed-record manager
owns a flat fixed-width backing array but retains the first backing index for
each unique record. The variable-vector manager owns arbitrary int32 slices.
Preserve zero-based IDs, first-insertion order, duplicate results,
resize-before-lookup timing and the source's prepend-based chain rebuild.

Implement the source-width little-endian Jenkins hash, its weighted-byte
alternative, the vector prime-sum hash and the exact ABC prime schedule where
one is accepted and two advances to three. Split every uint64 word into low and
high int32 halves for word-record helpers. Reproduce batch int and word ID maps
and the prebuilt unique-word manager, converting its uniqueness assertion into
a typed error.

Return owned backing and entry snapshots instead of exposing mutable source
arrays or scratch Vec_Int headers. Provide explicit append-and-intern for the
dynamic fixed-record use case. Preserve empty variable vectors and source
header/even-padding accounting. Bound entries, bins and stored integers,
support cancellation, validate every chain and use checked Close lifecycle.

Test every hash against an independent byte/arithmetic model across signed
boundary values and table sizes. Pin source examples, stable IDs, retained
source indices, dynamic append, word splitting, duplicate word rejection,
empty and natural vectors, padding, ownership, resize, lookup, limits,
cancellation and Close. Fuzz at least 5,000 mixed operations against independent
ordered models for all three managers without native code or retained output.

This validates the complete vector-hash header. The combined model campaign
passes 5,000 executions. Inventory becomes 391 validated, 39 partial and 1,327
unstarted files. The full suite passes in 18.613s, the race suite in 94.258s,
vet is clean, and the native-source, package-header, `testdata/`, symlink and
tool-allowlist audits pass. `pkg/tools/` contains only `source_manifest.py`.

## Typed nested-vector execution checkpoint

Port all 60 tracked definitions in `misc/vec/vecVec.h` as the concrete
compile-time generic `NestedVector[T comparable]`. The constraint is used only
to implement source pointer/integer identity for unique insertion; do not store
or dispatch through runtime interface values. Collapse the pointer and integer
C entry points into typed instantiations while retaining concrete int-only
sorting and formatting helpers.

Preserve the source outer minimum-eight capacity and exact requested expansion,
the inner minimum-sixteen/doubling schedule, automatically created empty
levels, append order, duplicate suppression, total entry count, deep clone,
clear-with-capacity-retention and source-ABI memory accounting. Return owned
level and traversal snapshots instead of exposing mutable inner slices. Replace
all start/stop, forward/reverse, pointer/int and paired-level macros with
checked `Levels`, `Entries` and `ZipNestedVectorLevels` results in the same
orders.

Use stable sorting to make equal-size/equal-key order deterministic where C
`qsort` leaves ties unspecified. Reject first-integer sorting when any level is
empty instead of reading outside its array. Replace stdout printing with exact
owned text. Consolidate Free, FreeP and Erase into checked Close because Go has
one ownership path. Bound levels, entries and retained inner capacity before
allocation, support cancellation, keep failures atomic and replace assertions
with typed errors.

Test outer and inner capacity transitions, empty expansion, pointer identity,
integer uniqueness, entry and level access, every traversal direction/range,
paired traversal, clone/snapshot ownership, clearing, stable size/first-key
sorting, exact formatting, ABI memory totals, invalid dimensions, all limits,
cancellation, atomic failure and closed lifecycle. Run an independent
nested-slice state model and the Go fuzzer for at least 5,000 mixed-operation
executions each, without native code or retained generated output.

This validates the complete nested-vector header. The deterministic model and
Go fuzz campaigns each pass 5,000 executions. Inventory becomes 392 validated,
39 partial and 1,326 unstarted files. The full suite passes in 18.709s, the race
suite in 94.895s, vet is clean, and the native-source, package-header,
`testdata/`, symlink and tool-allowlist audits pass. `pkg/tools/` contains only
`source_manifest.py`.

## Embedded integer-level vector execution checkpoint

Port all 59 tracked definitions in `misc/vec/vecWec.h` as a distinct concrete
`IntLevelVector`. Do not alias it to `NestedVector[int]`: the source embeds
`Vec_Int_t` headers across the entire outer capacity, so dormant levels,
header movement, insertion and compaction have observable semantics that the
pointer-owning vector does not share.

Preserve ordinary minimum-eight and exact allocation, exact Grow/Init,
sparse-Push growth by `max(2*size, level+1)`, PushLevel and InsertLevel
minimum-sixteen/doubling growth, and minimum-sixteen/doubling inner integer
storage. Retain dormant level contents when Init reduces the visible size;
make Shrink clear removed logical entries while retaining their allocations,
and make Clear empty active levels before setting visible size to zero. Preserve
unique insertion, deep duplication, active total/used counts and the source's
embedded-header memory calculation.

Return owned level/array snapshots and use checked integer level IDs instead of
borrowed `Vec_Int_t *` values and pointer subtraction. Replace every selected,
range, reverse, double-level and two-vector iterator with `Levels`,
`LevelsByIndex`, `LevelPairs` and `ZipIntLevelVectors`. Make size, first-entry
and last-entry sorts stable for deterministic ties; reject empty boundary-key
levels rather than reading out of bounds. Return exact ordinary and literal
text instead of printing.

Implement equivalence-class construction, nontrivial class/entry counts,
first-entry collection and pointer-vector conversion with owned Go values.
Represent the source's temporary high-capacity-bit mark as a separate Boolean,
including single set/toggle and atomic bulk mark/unmark with duplicate and state
validation. Compact nonempty levels in order and release empty storage. Use one
checked Close path for Zero/Erase/Free/FreeP. Bound levels, total retained
entries and inner capacities before allocation, support cancellation, validate
the embedded-slot state and keep failures atomic.

Test ordinary/exact and every outer/inner growth path, sparse push, dormant
Init resurrection, destructive Shrink, Clear, appended and inserted levels,
snapshot/clone ownership, traversal orders, pairing, uniqueness, all sorts,
exact formatting, equivalence classes, marks, compaction, memory, limits,
cancellation and lifecycle. Run both an independent embedded-slot state model
and the Go fuzzer for at least 5,000 mixed-operation executions without native
code or retained generated output.

This validates the complete embedded integer-level header. The deterministic
model and Go fuzz campaigns each pass 5,000 executions. Inventory becomes 393
validated, 39 partial and 1,325 unstarted files. The full suite passes in
18.820s, the race suite in 96.043s, vet is clean, and the native-source,
package-header, `testdata/`, symlink and tool-allowlist audits pass.
`pkg/tools/` contains only `source_manifest.py`.

## First supervised parallel-wave integration checkpoint

Apply the parallel-wave protocol to fourteen disjoint worker lanes, then require
independent source review before ledger promotion. Do not accept worker-local
compilation or fuzzing as sufficient evidence. The review phase must reconcile
actual definition lines rather than call-like name matches and must be allowed
to downgrade a broad submission to partial.

Validate the complete string, word and pointer vector headers at 68, 71 and 67
definitions respectively. Preserve native capacity schedules, owned transfer,
signed-byte and word behavior, typed generic pointer identity, simulation rows,
binary/hex codecs and deterministic safe replacements for raw pointer access.
Run separate 5,000-execution mutation and codec/simulation fuzzers. Independent
review must cover append growth, record capacity, nil memory, self-append,
iterator bounds, cancellation and removal of runtime-interface sorting and
formatting.

Validate all 31 semantic definitions in `misc/extra/extraUtilPrime.c` with
owned prime, pair, PLA and cube operations. Direct collision-free lookup may
absorb its static hash machinery. Correct the dormant distance-two item count
to match the actual `a < b` loop, bound overflow, and keep pair/artifact/cube
oracles independent of production helpers. Validate all 64 `cnf.h` items plus
its nine-unit build manifest; explicitly disposition the four allocation and
lifecycle declarations absorbed by Go ownership.

Complete FSIM manager/frontier/codec/simulation behavior across its remaining
six tracked files. Use exact owned `ABCRandom` defaults and an independent
random-topology scalar simulator. Increase the frontier to
`maximumCrossCut+2`: the source formula can loop forever because slot zero is
reserved and an AND result is allocated before its fanins are released. Treat
zero time as unlimited, correct witness pattern indexing, and replace global
RNG/stdout/graph-attached witnesses with owned results.

Complete the six-file BDC cluster only after pinning strong/weak OR and AND,
support minimization, reuse, UpdateRight absorption, first-support MUX choice,
care/divisors, node accounting and source statistics. Reuse the concrete BDC
table. Replace enormous process-global SPFD experiments and fixed files with
bounded search and owned datasets while retaining uint32 hashing, internal-node
overlap counts and exact level construction. Require two 5,000-execution fuzz
campaigns and independent decision tests.

Validate the six-file supergate cluster across all 64 active semantic
definitions. Preserve native tried counts, branch/truth order, complement
normalization, delay/area propagation, existing-class Pareto alternatives and
command options. Exact-truth lookup, deterministic ties and owned serializers
replace global hash/comparator/file state; correct the six-input area typo.

Validate the two unfold wrappers, clock-gate implication source, NWK checker,
rewrite temporary generator and AMAP manager/core units after independent
review. Record that automatic unfold classification and the remaining AMAP
graph/match/output algorithms stay in their own source units. Keep the six
Verilog files partial: owned streams/formulas and structural parsing are useful,
but mapped genlib gates, undefined-blackbox inference/attachment, full ABC
network finalization and hierarchy reporting remain unported.

Promote only after the integrated ordinary suite, package-wide race suite, vet,
policy scans and every new 5,000-execution fuzz target pass. This checkpoint
ends at 423 validated, 45 partial and 1,289 unstarted files; ordinary tests pass
in 22.357s and the race suite in 117.660s.

## Second supervised parallel-wave integration checkpoint

Expand the protocol to sixteen disjoint implementation lanes under a
20-sub-agent ceiling, retaining independent review and coordinator ownership of
the ledger/status/plan. Validate complete integer-vector behavior, the Extra
public header, DECO, DSC, SAT proof recording, EPD, RPO, RWT, MTR and the active
six-variable RSB surface. Require exact inventory reconciliation, owned Go
state, atomic failures, limits/cancellation, and at least 5,000 executions for
every new model/fuzz target.

For RWT, pin the exact 3,562-word static stream, all 135 practical classes,
222-bucket retained order and all 1,139 decomposition truths. For SAT proof
recording, require independent replay that does not reuse the production
resolver. For MTR, distinguish unrestricted basic trees from interval-ordered
group trees and stop interval semantics at terminal cut points. For EPD, test
subnormal decimal normalization, extreme exponent differences, signed zero and
the package's intentional non-IEEE exceptional conventions against an
independent high-precision model.

Retain honest partial states where a semantic replacement is not source parity.
CSAT lacks IvyProve and eight declared gates; PSAT lacks the MiniSat 1.14p
implementation; CAS lacks CUDD symmetry reordering and non-strict encoding;
INT2 uses finite-state reachability rather than the incomplete SAT-interpolation
sources; and FRET has only invariant/min-cut/frontier scaffolding. Leave
`fretInit.c` unstarted. In structural Verilog, validate formula, parse and stream
units, but keep the core/header partial until complete Mio truth/timing objects
and exact API/ABI behavior exist.

Independent review in this wave corrected union representation, DSC match
direction, vector growth/atomicity/order, proof root and pivot semantics, CAS
selector significance, EPD subnormal/exception behavior, MTR group validation,
RWT lifecycle/rollover, and Verilog bit-formal/blackbox handling. Promote only
after the integrated ordinary suite, full race suite, vet, source-hash and
policy audits pass. This checkpoint ends at 464 validated, 59 partial and 1,234
unstarted files; ordinary tests pass in 27.290s and the race suite in 138.659s.

## Third supervised parallel-wave integration checkpoint

Use twenty disjoint implementation clusters with independent review while the
coordinator retains the coverage ledger and shared documentation. Extend
stand-alone Go coverage across the root examples, DEC, retiming, rewriting,
interpolation, MiniAIG, CSW, XSAT, clock gating, FXCH, PLA, IOA, WLN, SSC, DCH,
Lucky, SIM, RES, HOP and AMAP. Keep each source file partial when the tested Go
surface is a concrete substitute for, rather than a complete port of, its ABC
network, cut, timing, CUDD, interpolation or solver integration.

Require independent review of parser rejection and limits, normalized Boolean
representations, transactional mutation, graph ownership and lifecycle,
retiming preimages, rewrite rollback, interpolation constraints and budgets,
MiniAIG binary layouts, CSW signatures, SAT literal normalization, clock-gating
polarity and ODCs, PLA consensus, latch initialization, word-level signedness,
sequential object limits, representative-aware choice rebuilding, structural
symmetry, truth bounds and mapping phase/tie behavior. Corrections in all of
these areas must land before ledger promotion.

Validate complete active CSW, IOA and DCH behavior and the independently
testable DEC, PLA, HOP, WLN, Lucky, SIM, retiming, rewriting, interpolation,
MiniAIG, SSC, RES and AMAP surfaces. Leave integration-heavy headers and units
partial with exact caveats. Use only Go-owned models and datasets; add no
C/cgo/native process oracle and no permanent `testdata/` corpus.

Promote only after the frozen combined tree passes the package suite, race
suite, vet, all source/hash and package-policy audits, and at least 5,000
executions for each of 26 new fuzz targets. This checkpoint ends at 553
validated, 149 partial and 1,055 unstarted files; ordinary tests pass in
31.820s and the race suite in 162.056s.

## Fourth supervised parallel-wave integration checkpoint

Complete the twenty implementation lanes already dispatched before reducing
the continuing concurrency limit to one supervisor plus three sub-agents.
Independently review AIG cuts/factoring/canonical recording, Ivy and SAIG
foundations, KIT truth/SOP/PLA/Cloud behavior, the CUT kernel, MIO, MPM, COV,
EXOR, placement, CUDD arithmetic/Boolean operations, SATOKO, live-property
transformations and ACEC polynomial utilities. Keep shared documentation and
the authoritative ledger coordinator-owned.

Require source-definition reconciliation and final-tree tests rather than
accepting worker-local compilation. Review must cover cancellation rollback,
solver limit/error distinctions, packed-index overflow, ambiguous formulas,
permutation grouping, option propagation and exact deterministic formatting.
The review phase corrects all seven of these issue classes and adds focused
regressions before promotion.

Validate the complete independently testable AIG cut/factoring, KIT, live
transformation and algebraic foundations. Retain partial dispositions for
headers and units whose exact source behavior requires missing frame/network,
ADD, sequential-cut, native arena/heuristic or wider subsystem orchestration.
Leave COV network adapters outside the promoted set. All permanent validation
remains Go-only and uses owned models or datasets.

Promote only after the frozen combined tree passes ordinary, race and vet
gates; all 23 new fuzz targets run for at least 5,000 executions; and source
hash, aggregate identity, Go-file reference, formatting, package-header,
native-source, `testdata/`, symlink and tool-allowlist audits pass. This
checkpoint ends at 592 validated, 214 partial and 951 unstarted files; ordinary
tests pass in 43.241s and the race suite in 218.083s.

## Fifth supervised parallel-wave integration checkpoint

Continue under the reduced concurrency limit of one supervisor and at most
three sub-agents. Use three disjoint lanes for the foundational ABC
object/network representation, FRAIG owned data structures and DSD
decomposition. Retain coordinator ownership of the ledger and shared status
documents, and independently review each lane before promotion.

For ABC networks, require stable object identity, terminal/box/latch
partitions, phase-correct fanin replacement and insertion, exact fanout
multiplicity, constant-one invariants and transactional cancellation. For
FRAIG, require complemented-node identity, deterministic vector/pool behavior,
structural and bounded functional interning, source-shaped simulation masks,
fanout transfer rollback and overflow-safe accounting. For DSD, require owned
complemented nodes, OR/XOR/prime reconstruction, terminal-free traversal,
source-order decomposable variables, large-prime blocking and transactional
BDD conversion.

Keep broader network transformations, representation-specific object
functions, FRAIG SAT/feedback/choice orchestration and full CUDD
containment/care/cube-remapping/reference mechanics partial. Preserve the
exhaustive DSD campaign while rotating bounded manager state so the complete
race suite remains practical rather than dropping semantic coverage.

Promote only after ordinary, race and vet gates, independent model campaigns,
coordinator reruns of all three fuzz targets, and package/source policy audits
pass. This checkpoint ends at 596 validated, 232 partial and 929 unstarted
files; ordinary tests pass in 50.356s and the race suite in 256.353s.

## Sixth supervised parallel-wave integration checkpoint

Continue with three sub-agents on disjoint ABC analysis, remaining FRAIG
orchestration and FXU sparse-structure lanes. Independently review each lane
before coordinator-owned ledger promotion, preserving the four-thread total.

For ABC analysis, pin terminal name ordering, combinational latch comparison,
latch-triple shape, transparent-net traversal/levels, internal-node ordering,
reference multiplicity and MFFC boundaries. For FRAIG, require phase-correct
choice classes, root-aware TFI, normalized MUX controls, unique source-shaped
supergates, SAT/feedback budgets, composed cancellation and accurate query
statistics. For FXU, require derived-state invalidation, valid noncontained
cube pairs, transactional work accounting, bounded scans/singles and complete
heap/pair provenance validation.

Keep representation-library/hierarchy/GIA/choice checks, raw traversal-label
mutation, native FRAIG heuristic/SAT-cone scheduling, feedback compression,
FXU raw pointer storage, complement lookahead and wider network
conversion/update/printing partial. Validate all concrete replacements with
independent adjacency, Boolean and sparse-incidence models.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 599 validated, 247 partial and 911 unstarted files; ordinary tests pass
in 51.737s and the race suite in 263.703s.

## Seventh supervised parallel-wave integration checkpoint

Continue with three sub-agents on disjoint ABC metadata/SOP, remaining FXU flow
and Gordian placement lanes. Independently review all three before
coordinator-owned ledger promotion.

For ABC metadata, require terminal partition-aware name comparison,
transactional name/latch changes, source-sequential self-feed removal,
phase-preserving fan ordering and overflow-safe owned SOP algebra. For FXU,
require one cumulative work budget across rebuilds, cancellation-aware
normalization, immediate definition-capacity checks and semantic/cost replay.
For Gordian placement, require consistent internal/leaf completion state,
incremental tree reopening, fixed core geometry, bounded clique-edge growth,
mid-net cancellation and retained QP limits.

Keep native name/flex-manager lifecycle, advanced latch transformations,
ABC-network FXU adapters, complement lookahead, native incremental rings,
external placement partitioning, sparse-array reuse, dormant experiments and
standalone process/file drivers partial. Validate concrete replacements with
independent SOP, cover/truth and partition/conservation models.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 601 validated, 263 partial and 893 unstarted files; ordinary tests pass
in 52.080s and the race suite in 264.661s.

## Eighth supervised parallel-wave integration checkpoint

Continue with three sub-agents on disjoint ABC functional/conversion, ABC
utility/library and CUDD ADD lanes, with independent review before promotion.

For ABC conversion, require malformed latch/box rejection, bounded
duplicate-fanin truth reconstruction, phase/partition/name preservation,
shared-root uniqueness and transactional mappings. For utilities, require
source-shaped node statistics and crosscut event order, dangling-cone safety,
transactional barrier mutation and bounded deterministic reports. For ADDs,
require canonical numeric terminals, bounded memoized DAG queries, cumulative
generation budgets and rollback across apply, ITE, abstraction, extrema,
inverse/negation/log/round and Walsh/residue construction.

Keep BDD/ZDD/GIA/Mio/map/BLIF-MV representations, native hash-manager and
sequential-box variants, hierarchy/library-manager integration, broad
synthesis/viewer behavior, arbitrary C callback apply and source-specific
residue update quirks partial. Validate concrete semantics with independent
truth/network round trips, graph/report models and dense numeric tables.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 607 validated, 273 partial and 877 unstarted files; ordinary tests pass
in 53.261s and the race suite in 266.866s.

## Ninth supervised parallel-wave integration checkpoint

Continue with three sub-agents on remaining ABC hierarchy/MV behavior, the
CUDD manager/cache/API foundation and a disjoint CUDD Boolean-analysis cluster.
Require independent review before coordinator-owned promotion.

For hierarchy, require absent-default MV semantics, exact copied domains,
XOR/XNOR marker phases, unique names/ports, full-design recursion checks,
atomic aggregate limits and bounded serialization/flattened names. For the
CUDD foundation, require source-aligned variable roles, canonical cache keys,
constructor cleanup, complemented-reference sharing and exact build
inventories. For Boolean analysis, require direct essential implications,
constant sentinel behavior, deduplicated literal intersections, true restrict
semantics and rollback of discarded canonical nodes.

Keep native ABC hierarchy mutation, black-box/latch/HOP/file/SAT workflows, Au
ABI and CBLIF/ternary paths partial. Keep CUDD ZDD, dynamic reordering and
collection, hooks/streams, raw callback/ABI caches, weighted paths, ADD
analysis and approximate/iterative heuristic suites partial. Validate concrete
semantics with independent hierarchy, truth and manager-state models.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 612 validated, 294 partial and 851 unstarted files; ordinary tests pass
in 53.072s and the race suite in 274.193s.

## Tenth supervised parallel-wave integration checkpoint

Continue with exactly three sub-agents on disjoint CUDD approximation,
reordering and ZDD lanes. Replace each completed implementation lane with an
independent reviewer before coordinator-owned promotion, preserving the
one-supervisor/three-sub-agent limit throughout.

For approximation, require subset/superset contracts, compatible projection,
threshold preservation, clipping, priority/signature/Hamming semantics,
cumulative limits and cancellation. For reordering, require canonical semantic
cost models, bounded rebuild allocation, cycle detection, stable sifting,
explicit ranges, value-owned linear maps, interactions and deterministic
heuristic ties. For ZDDs, require zero-suppressed canonical identity, exhaustive
set algebra, DAG-based count/support, terminal-correct size, bounded family
expansion and transactional cumulative BDD transfers.

Keep exact native approximation rankings and simultaneous clipping recursion,
in-place unique-table reordering, full group trees, dynamic collection,
source-specific annealing/genetic schedules, paired-literal binate ZDD
operations, printing, multiplicities and minimal irredundant ISOP partial.
Validate all concrete behavior with independent dense truth, permutation and
set-family models plus rollback, cancellation and resource regressions.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 616 validated, 312 partial and 829 unstarted files; ordinary tests pass
in 70.886s and the race suite in 382.453s.

## Eleventh supervised parallel-wave integration checkpoint

Continue with exactly three disjoint lanes for CUDD bridge/I/O formats, ADD
matrix operations and ZDD reordering utilities. Replace implementation agents
one-for-one with independent reviewers and leave spare slots idle until the
checkpoint closes.

For bridge/I/O behavior, require fixed-order ADD/BDD conversions, cumulative
transactional manager transfer, bounded sparse and RUA parsing, additive
duplicates, validated export graphs, portable identifiers and preflighted
output growth. For matrices, require all four public arithmetic, times-plus,
min-plus and outer-sum operations, noncontiguous and duplicate variable
handling, explicit infinity/NaN decisions and rollback. For ZDD reordering,
require canonical order costs, malformed-DAG rejection, bounded decode,
deterministic searches/groups/symmetry/transforms, cancellation and owned path
results.

Keep dynamic cross-order restarts and shared nodes, byte-identical streams and
layouts, extended matrix classes, caller-array mutation, native unique-table
swaps, group trees, hooks, collection and address-based output partial. Promote
matrix multiplication as a semantic adaptation only after independent dense
models cover every operation; retain the broader I/O and ZDD units as partial.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 617 validated, 321 partial and 819 unstarted files; ordinary tests pass
in 72.309s and the race suite in 385.561s.

## Twelfth supervised parallel-wave integration checkpoint

Continue with exactly three disjoint BDD lanes: BBR reachability, the ExtraBDD
utility cluster and REO multi-output reordering. Replace each implementation
agent with an independent source reviewer and begin no later wave until the
combined tree is frozen and promoted.

For BBR, require bounded partitioned image/preimage, explicit-state onion-ring
reachability, deterministic shortest counterexamples, replay, state-set BDD
construction, overflow-safe memory and cumulative cancellation. For ExtraBDD,
require canonical-manager validation, half-open cube ranges, source-aligned
empty-family behavior, broad Boolean and family operations, translation,
profiles, symmetry/unateness, thresholds, Gray/K-map, deterministic budgets and
non-expanding membership. For REO, require terminal-aware objectives, compact
support planes, cumulative profiles, stable sift/shuffle/swap behavior and
bounded canonical unit-DAG transfer.

Keep native ABC/CUDD frame mutation, quantification schedulers, scratch storage,
bit-packed counterexamples, equation/encoding/cover paths, paired-literal ZDD
representations, remap-up, complemented-pointer sharing and native allocation,
hash, reference and table mechanics partial. Validate standalone semantics with
explicit transition systems, exhaustive truth/family models and canonical
multi-output order models.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 622 validated, 346 partial and 789 unstarted files; ordinary tests pass
in 73.238s and the race suite in 385.398s.

## Thirteenth supervised parallel-wave integration checkpoint

Continue with exactly three disjoint base-network lanes covering ACB, BAC and
CBA. Replace each implementation agent one-for-one with an independent reviewer
and leave later work unopened until combined promotion finishes.

For ACB, require owned networks/objects/ranges/attributes, truth and structural
analysis, source-aligned interface levels and critical paths, transactional
normalization, canonical set algebra, aggregate ownership and cumulative
limits. For BAC, require stable hierarchical word networks, an explicitly
bounded primitive subset, exact binary preflights, strict structural
Verilog/scalar BLIF subsets, atomic construction, bounded output and reconciled
metadata. For CBA, require exact type numbering, exclusive design ownership,
stored-shape validation, broad checked word semantics, direct-output/constants,
strict declarations/cubes/trailing input and bounded Boolean format round trips.

Keep full command/frame integration, native byte/pointer layouts, additional
primitive families, complete BLIF/SystemVerilog grammars and hierarchy,
mapped/RAM/sequential constructs, advanced blasting/insertion, SAT/GIA and
synthesis heuristics partial. Validate owned semantics using exhaustive small
truth/word models, structural round trips and malformed-input campaigns.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 628 validated, 378 partial and 751 unstarted files; ordinary tests pass
in 73.046s and the race suite in 412.147s.

## Fourteenth supervised parallel-wave integration checkpoint

Continue with exactly three lanes covering the remaining base command, main
and I/O clusters. Replace implementations one-for-one with independent reviews
and freeze the combined tree before any ledger promotion.

For commands, require deterministic registry/session behavior, source-aligned
parsing, aliases, flags, history, nested sources, builtins, configuration,
atomic multi-name mutations, aggregate accounting and nil/cancellation safety.
For main, require owned frame/runtime lifecycle, stable network identities,
transactional history updates, libraries, verification/log/session state,
snapshots, source-shaped CLI ordering and cumulative copy/text limits. For I/O,
require canonical Boolean/latch graphs, source-aligned BLIF off-set/default and
constant semantics, strict bounded format subsets, cumulative parse/evaluation/
format budgets and explicit rejection of silent sequential loss.

Keep native process globals, filesystem/shell discovery, dynamic plugins,
opaque exchange APIs, streams and interactive behavior partial. Credit only
source-faithful I/O units: leave generic JSON, BAF, BBLIF, BLIF-AIG, BLIF-MV,
DSD, EDIF and Bookshelf-specific implementations unstarted when the Go layer
does not implement their actual semantics.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 631 validated, 419 partial and 707 unstarted files; ordinary tests pass
in 74.024s and the race suite in 392.377s.

## Fifteenth supervised parallel-wave integration checkpoint

Continue with exactly three disjoint lanes for WLC word-level circuits, Ivy AIG
extensions and NWK optimization networks. Replace implementations one-for-one
with independent reviews and freeze the combined tree before promotion.

For WLC, require exact type values, owned ranges/flops/networks, checked word
and sequential semantics, strict source shapes, topology/inventory integrity,
cumulative graph-analysis limits, cancellation/snapshots/lifecycle and a
loss-rejecting structural-Verilog bridge. For Ivy, require complemented-edge
correctness, acyclic validation, bounded attributes, cuts/truth/MFFC, mapping,
verified choices/equivalence, <=6-variable canonical truth, sequential memory,
balance growth and cumulative limits. For NWK, require constants, source-order
traversals, graph reconciliation, truth/support/MFFC, numeric timing, output-
rooted depth, derived strash/cut limits and bounded merge ownership.

Keep WLC abstraction/simulation/memory/NDR/UI/SMT and command paths unstarted
when absent. Keep native Ivy DSD/RWT/SAT/CUDD/FRAIG/HAIG/mapping engines and
NWK NTL/IF/HOP/GIA/FRA/SSW, flow/retiming, bidec/LUT/pin/power integrations
partial. Validate concrete semantics with exhaustive small word/truth models,
canonical graph models, malformed-state campaigns and transactional limits.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 634 validated, 461 partial and 662 unstarted files; ordinary tests pass
in 74.286s and the race suite in 392.900s.

## Sixteenth supervised parallel-wave integration checkpoint

Continue with exactly three disjoint proof lanes for ABS abstraction, CEC
equivalence and SSW sequential sweeping. Replace implementations one-for-one
with independent reviewers and freeze the combined tree before promotion.

For ABS, require validated explicit-state systems, stable projection IDs,
existential reachability, monotone transactional refinement, exact supplied-
witness and proof replay, real/spurious discrimination, bounded conversions and
cumulative construction/search limits. For CEC, require strict AIG validation,
output-aligned miters, exact combinational checking, shortest paired-state BFS,
honest bounded Unknown, trailing-frame rejection, replayable witnesses, output
classes and cumulative ownership/accounting. For SSW, require partition-kind
validation, deterministic simulation/signatures/classes, bounded equivalence,
phase-aware verified sweep mappings, exact signature bits, atomic cancellation
and cumulative class/pattern/result limits.

Keep ABS RPM dominator/reparameterization algorithms and SSW constraint,
dynamic, filter, island, pair, partition, rarity, semi and uniqueness workflows
unstarted when absent. Keep scalable native GIA/AIG/SAT/PDR/CUDD/fraiging/CNF
engines, proof objects, commands and exact heuristics partial. Test soundness
with exhaustive transition systems, tampered-witness rejection and replay.

Promote only after ordinary, race and vet gates, coordinator reruns of the
three fuzz targets and all package/source policy audits pass. This checkpoint
ends at 637 validated, 512 partial and 608 unstarted files; ordinary tests pass
in 74.629s and the race suite in 395.767s.

## Seventeenth supervised parallel-wave integration checkpoint

Continue with exactly three disjoint proof lanes for FRA fraiging, ACEC
arithmetic equivalence and the remaining PDR boundaries. Replace implementations
one-for-one with independent reviewers and freeze the combined tree before
promotion.

For FRA, require exact zero-state and per-frame replay, final-divergence checks,
bounded proof labelling, deterministic signatures/classes, verified replacement
mappings, atomic cancellation and cumulative transition/SAT/sweep limits. For
ACEC, require bounded signed and unsigned polynomial arithmetic, canonical
inputs, exact equivalence, transactional ownership, cumulative expansion limits
and callback-safe concurrent lifecycle handling. For PDR, require exact
observation partitions, state-only clauses and invariants, atomic generalize/
subsume/push/archive operations, non-compounding statistics, strict witness
replay and cumulative construction/work limits.

Keep FRA clause, hotspot, implication, independent-verification and partition
workflows unstarted when absent. Keep native ABC graph/CNF/SAT integration,
scalable native detectors, rewriters, timing/NPN/slicing engines and PDR CNF/GIA
solver construction explicit partial or unstarted gaps. Test concrete semantics
with exhaustive small models, malformed-state campaigns, tampered replay,
concurrency and fixed-count Go-only fuzzing.

Promote only after ordinary, race and vet gates, coordinator reruns of the three
fuzz targets and all package/source policy audits pass. This checkpoint ends at
640 validated, 546 partial and 571 unstarted files; ordinary tests pass in
74.289s and the race suite in 394.329s. All three fuzz targets pass exact
5,000-execution coordinator reruns, and the Go-only source and policy audits are
clean.

## Eighteenth supervised parallel-wave integration checkpoint

Continue with exactly three disjoint lanes for bounded ABCI truth analysis,
remaining GIA simulation/replay/cones and SAIG supplied-trace simulation and
switching. Replace implementations one-for-one with independent reviewers and
freeze the combined tree before promotion.

For ABCI, require immutable <=16-variable truth tables, exact activity/pair
measures, deterministic ordering, exact single-variable quantification and
unateness, a bounded six-real-variable decomposition search and explicit
corrections for native undefined/inflated behavior. For GIA, require exact
constant/complement/buffer/XOR/mux/CI/CO semantics, latch partitions, supplied-
frame timing, strict counterexample replay, deterministic failure order,
explicit-pattern statistics and sequentially closed deterministic cone groups.
For SAIG, require exact supplied-trace replay, all-object probabilities and
2*p*(1-p) activity, immutable results, exact inventory and bounded orbit helper
tests without claiming native ternary phase abstraction.

Keep ABCI network/BDD/SAT/command workflows, GIA random packed managers,
counterexample minimization/extension and heuristic cone/signature workflows,
and SAIG random packed/recycling/ternary/control/CEX flows unstarted when absent.
Require cumulative retained-plus-temporary byte/work limits, overflow-safe
preflights, responsive cancellation, atomic errors, concurrent read-only tests,
independent exhaustive models and fixed-count Go-only fuzzing.

Promote only after the full ordinary suite, vet, targeted race tests for all
three lanes, coordinator reruns of all three fuzz targets and every source/policy
audit pass. This checkpoint ends at 642 validated, 555 partial and 560 unstarted
files; ordinary tests pass in 74.600s and targeted race tests in 1.078s. Each
fuzzer passes exactly 5,000 executions. Under the documented milestone cadence,
wave 20 is the next scheduled full-package race checkpoint.

## Nineteenth and final parallel-wave integration checkpoint

Use the final three parallel lanes for IF mapping, classic Mapper and LLB
partition scheduling. Replace each implementation one-for-one with independent
source review, then freeze the combined tree. After this checkpoint all porting
and review phases execute serially in one active thread.

For IF, require bounded normalized truth with separate complement phase,
ordered cut union/expanded AND, modulo-31 signatures, dominance and cost orders,
stable pin/truth permutation, a strict owned LUT-library subset and variable-pin
arrival matching. For Mapper, require <=6-leaf Num-ordered merge, exact leaf
phase, source-shaped match ordering, consistent rise/fall/worst timing, complete
two-word elementary truths and stretched phase transforms. For LLB, require
strict supplied PI/current-state/internal/next-state incidence shape, stable
last-use ordering, native-shaped greedy scheduling and exact quantification/
attraction clustering over supplied partitions.

Keep full IF/Mapper cut enumeration, filtering, matching, libraries, required-
time propagation and mapping passes unstarted when absent. Keep LLB native
headers/managers, AIG group derivation, CUDD image/reachability, SAT/GIA,
reporting and commands unstarted. Require exact inventory, immutable ownership,
deterministic ties, responsive cancellation, atomic zero errors, cumulative
work/object/retained-plus-temporary byte limits, concurrent read-only tests,
independent exhaustive models and fixed-count Go-only fuzzing.

Promote only after the full ordinary suite, vet, targeted race tests for all
three lanes, exact 5,000-execution reruns of all fuzzers and every source/policy
audit pass. This checkpoint ends at 645 validated, 568 partial and 544 unstarted
files; ordinary tests pass in 74.519s and targeted race tests in 1.111s. Wave 19
is the final parallel wave. Wave 20 is the next scheduled full-package race
milestone and all its implementation, review and integration work is serial.

## Twentieth serial-wave integration checkpoint

Begin strictly serial porting with the remaining live-property cluster. Cover
only source-specific portable behavior from `monotone.c` and
`disjunctiveMonotone.c`: pending/hint substring discovery in PO order, remaining
monotone/consequent filtering, ordered vector difference and source-shaped
unique insertion. Validate the complete nine-unit `module.make` inventory.

Require owned results, nonnegative indices, native order and duplicate behavior,
cumulative item/work/result-byte limits, responsive mid-operation cancellation,
atomic zero errors, concurrent immutable callers and an independent Go-only
fuzz model. Keep monotone/disjunctive AIG tester construction, PDR proof search,
recursive level generation, network mutation, printing/files and commands
unstarted when absent; helper coverage must not imply those workflows.

Promote only after focused ordinary and race tests, exactly 5,000 fuzzer
executions, the full ordinary suite, vet, the scheduled full-package race suite
and all source/policy audits pass. This checkpoint ends at 646 validated, 570
partial and 541 unstarted files; ordinary tests pass in 74.996s, focused race
tests in 1.058s and the milestone full race suite in 397.104s. All work is
performed in the single active supervisor thread.

## Twenty-first serial-wave integration checkpoint

Continue serially with `kitCloud.c`. Connect the existing owned TruthTable and
CloudManager using canonical Shannon expansion in native variable order,
complemented edges and Cloud unique-table sharing. Reconstruct exact truth
semantics from a Cloud root without retaining or exposing manager storage, and
validate the complete twelve-unit KIT build inventory.

Require strict manager/truth dimensional checks, bounded variable count,
operation limits, responsive traversal cancellation, exact manager rollback on
all failed construction paths, owned reconstructed tables, complement sharing,
concurrent independent-manager tests, exhaustive small functions and a Go-only
fuzz model. Keep packed `Kit_Mux_t` serialization, scratch-vector stores,
general truth composition, cofactor-support propagation and unrelated KIT AIG,
BDD, DSD, factor, graph and HOP engines unstarted when absent.

Promote only after focused ordinary and targeted race tests, exactly 5,000
fuzzer executions, the full ordinary suite, vet and all source/policy audits
pass. This checkpoint ends at 647 validated, 571 partial and 539 unstarted
files; ordinary tests pass in 74.797s and focused race tests in 1.070s. The
wave-20 full-package race result remains the current milestone gate. All work
is performed serially in the supervisor thread.

## Twenty-second serial-wave integration checkpoint

Continue serially with the concrete core of `kitGraph.c`. Port bounded graph,
constant and literal construction; complemented edges; atomic AND, DeMorgan OR,
both XOR forms and both MUX forms; exact native <=5-input truth evaluation;
maximum graph level; and per-leaf depth with the native absent sentinel.

Require private node storage, strict backward-edge validation, immutable edge
values, mutex-protected mutation/read phases, atomic capacity failure for
three-node operations, node/work/temporary-byte limits, traversal cancellation,
concurrent frozen-reader tests and an independent Boolean operation-stream fuzz
model. Keep ISOP/factor truth-to-graph conversion, packed statistics arrays and
file parsing/testing utilities unstarted when absent.

Promote only after focused ordinary and targeted race tests, the exact
5,000-execution fuzz invocation, the full ordinary suite, vet and all source/
policy audits pass. This checkpoint ends at 647 validated, 572 partial and 538
unstarted files; ordinary tests pass in 74.458s and focused race tests in
1.025s. The wave-20 full-package race remains the current milestone gate, and
all work remains serial.

## Twenty-third serial-wave integration checkpoint

Continue serially with the graph bridge in `kitAig.c`. Map constant, literal
and internal KIT graph roots into the existing Go AIG manager, preserving every
leaf literal, graph-edge complement and root complement while using the AIG
manager's structural hashing.

Return an owned manager clone so conversion is transactional and the caller's
source remains untouched on both success and failure. Require strict source/
fanin/graph validation, object/work/retained-byte limits, cancellation before
allocation and during traversal, constants and complemented literals, exact
truth agreement, concurrent independent-result tests and a Go-only fuzz model.
Keep `Kit_TruthToAig` unstarted until source-shaped ISOP/factoring truth-to-graph
conversion exists; do not substitute generic truth synthesis for that path.

Promote only after focused ordinary and targeted race tests, exactly 5,000
fuzzer executions, the full ordinary suite, vet and all source/policy audits
pass. This checkpoint ends at 647 validated, 573 partial and 537 unstarted
files; ordinary tests pass in 74.953s and focused race tests in 1.026s. The
wave-20 full-package race remains the current milestone gate, and all work
remains serial.

## Twenty-fourth serial-wave integration checkpoint

Continue serially with the portable representative-table contract in
`aigRepr.c`. Use a separately owned table rather than overloading the existing
AIG choice-successor chains. Cover bounded lifecycle, strict lower-ID creation,
lower-pair selection, direct and transitive lookup, clearing, counting and the
phase correction used when resolving a directly mapped representative. Also
validate the exact ordered 31-translation-unit AIG `module.make` inventory.

Snapshot object type and phase metadata at construction so the representative
table does not retain its source manager. Reject buffered managers as the
source does, validate representative endpoints, preserve atomic errors, bound
objects/work/retained bytes, poll cancellation and protect shared table access.
Test exact inventory ownership, direct and transitive chains, clearing,
phase/complement combinations, invalid inputs, limits and concurrent frozen
readers. Fuzz mixed set/clear traces against an independent Go-only array model.
Keep representative transfer, duplication, remapping, TFI checks, rehash,
valid-choice marking and mapped-class transfer unstarted when absent.

Promote only after focused ordinary and targeted race tests, exactly 5,000
fuzzer executions, the full ordinary suite, vet, a catch-up milestone full race
suite and all source/policy audits pass. This checkpoint ends at 648 validated,
574 partial and 535 unstarted files; ordinary tests pass in 74.840s, focused
race tests in 1.054s and the milestone full race suite in 404.827s. All work
remains serial in the supervisor thread.

## Twenty-fifth serial-wave integration checkpoint

Continue serially with both functions in `src/base/io/ioWriteBblif.c`. Convert
the concrete `ABCNetwork` SOP-logic representation directly into the existing
owned `BblManager`; do not route through or claim support for the separate
`IONetwork` abstraction. Preserve original possibly sparse object IDs, CI/CO
order, node fanin order, standard and native `x`/`n` SOP text, and exactly the
reachable-node postorder of `Abc_NtkDfs` with `fCollectAll` disabled. Return
owned deterministic binary bytes in place of process file mutation.

Reject non-SOP/non-logic networks, sequential boxes, nets, malformed terminal
shapes, complemented logic edges, cycles and malformed SOPs with typed errors.
Require atomic failure, source ownership, object/ID/fanin/SOP/work/construction-
byte/binary-byte limits and responsive cancellation. Pin conversion against a
separately constructed BBLIF manager and parse the exact bytes back. Cover
dangling-node omission, sparse IDs, native XOR markers, malformed cases, every
limit and concurrent frozen callers. Fuzz bounded DAGs and exhaustively compare
all three-input assignments using independent ABC-SOP and BBLIF evaluators.

Promote only after focused ordinary and targeted race tests, exactly 5,000
fuzzer executions, the full ordinary suite, vet and all source/policy audits
pass. This checkpoint ends at 649 validated, 574 partial and 534 unstarted
files; ordinary tests pass in 75.015s and focused race tests in 1.069s. The
wave-24 full-package race remains the current milestone gate, and all work
remains serial in the supervisor thread.

## Twenty-sixth serial-wave integration checkpoint

Continue serially with the first exact half of `src/base/io/ioReadBblif.c`:
`Bbl_ManToAbc`, `Bbl_ManDfs_rec` and `Bbl_ManDfs`. Reconstruct the concrete
owned `ABCNetwork` SOP-logic representation in BBLIF object order, create all
objects before edges so forward references work, retain function and fanin
order, copy the design name and assign source-shaped dummy terminal names.
Return the fanin-first order of all LUTs, including dangling LUTs. Use private
traversal state rather than consuming persistent manager mark bits so frozen
calls remain deterministic and concurrency-safe.

Reject cycles, boxes, output fanins and malformed object/function shapes.
Require atomic failure, owned results, object/ID/fanin/work/byte limits and
responsive cancellation. Test exact order and names, forward references,
dangling nodes, source ownership, cycles, every limit and concurrent callers.
Fuzz bounded BBLIF DAGs and exhaustively compare all three-input assignments
between an independent BBLIF evaluator and the reconstructed network converted
through the existing AIG boundary. Keep `Bbl_ManToAig`, `Bbl_ManVerify`,
`Bbl_ManTest` and `Io_ReadBblif` explicitly unimplemented until their factoring,
CEC and top-level read contracts are ported and reviewed.

Promote the source only to partial after focused ordinary and targeted race
tests, exactly 5,000 fuzzer executions, the full ordinary suite, vet and all
source/policy audits pass. This checkpoint ends at 649 validated, 575 partial
and 533 unstarted files; ordinary tests pass in 75.129s and focused race tests
in 1.058s. The wave-24 full-package race remains the current milestone gate,
and all work remains serial in the supervisor thread.

## Twenty-seventh serial-wave integration checkpoint

Complete the remaining active contracts in `src/base/io/ioReadBblif.c`:
`Bbl_ManToAig`, `Bbl_ManVerify`, `Bbl_ManTest` and `Io_ReadBblif`. Cache the
native-order `Dec_Factor` result by BBLIF function handle, translate its concrete
factor graph into the existing structurally hashed AIG, create outputs in source
order and remove dangling logic. Return the existing concrete CEC result instead
of printing verification output. Replace filename, timing and process-file
behavior with owned binary input and an in-memory semantic round trip.

Require strict BBLIF and factor-graph validation, source ownership, atomic
failure, object/ID/fanin/SOP/binary/work/byte limits, responsive cancellation
and concurrent frozen callers. Test constants and exact AND behavior, shared
functions, dangling cleanup, names, corrupt bytes, every limit, equivalent and
inequivalent CEC outcomes and counterexample replay. Fuzz bounded ABC DAGs,
marshal and read them, then independently convert and exhaustively compare all
three-input assignments without a C runtime oracle.

Promote `ioReadBblif.c` to validated only after focused ordinary and targeted
race tests, exactly 5,000 round-trip fuzzer executions, the full ordinary suite,
vet, the scheduled full-package race suite and all source/policy audits pass.
This checkpoint ends at 650 validated, 574 partial and 533 unstarted files;
ordinary tests pass in 74.777s, targeted race tests in 1.079s and the full race
suite in 393.575s. All subsequent work remains strictly single-threaded.

## Twenty-eighth serial-wave integration checkpoint

Port the complete generic nested-JSON utility in `src/base/io/ioJson.c` without
substituting the existing `IONetwork` JSON interchange. Represent the source
`Abc_Nam_t` and `Vec_Wec_t` model as one concrete immutable document containing
insertion-ordered interned scalar strings and a forward-owned array/object
arena. Preserve brace/bracket preprocessing, the source whitespace-delimited
token rules, all-scalars-as-strings semantics, object/array entry order, exact
pretty-print indentation, every fixed neural-layer label and ordered extraction.
Use owned input/output bytes for read, write, extraction and the read-test round
trip rather than filenames or process state.

Require matched nesting, even object pairs, one rooted tree, forward child
ownership, atomic errors, input/output/object/per-node-entry/depth/work/retained-
byte limits and responsive cancellation. Test preprocessing bytes, arena shape,
string interning, deterministic and idempotent formatting, all layer mappings,
exact extraction, malformed structures, ownership, every limit and concurrent
frozen readers. Fuzz arbitrary bounded bytes through parse, format, reparse and
identical reformat without a C runtime oracle.

Promote `ioJson.c` from unstarted to validated only after focused tests, exactly
5,000 fuzzer executions, targeted race tests, the full ordinary suite, vet and
all source/policy audits pass. This checkpoint ends at 651 validated, 574
partial and 532 unstarted files; ordinary tests pass in 75.077s and targeted
race tests in 1.026s. Wave 27's full race remains the current milestone gate,
and all work remains strictly single-threaded.

## Twenty-ninth serial-wave integration checkpoint

Port `src/base/io/ioReadBaf.c` and `ioWriteBaf.c` as one coherent Binary AIG
Format boundary. Use an immutable concrete model containing the owned AIG,
separate PI and PO name vectors, complete latch name triples and two-bit latch
initialization. Keep the vectors separate so the format's permitted PO-name
overlap is not lost to the concrete ABC network's unique-name index. Provide
checked model/network bridges for representable names, but never silently
rename exact format metadata. Replace process files with owned bytes, native-
endian unsigned words with deterministic little-endian encoding, progress bars
with cancellation, and the process-clock comment with stable text.

Require strict NUL fields, nonnegative decimal counts, exact header/body
alignment, backward BAF fanins, complemented-edge decoding, structural hashing,
CI/CO register partitioning, latch initialization, atomic errors and bounded
input/output/object/name/work/retained bytes. Reject graph buffers and metadata
the format cannot encode. Independently decode fixed output words, exhaustively
compare all small sequential assignments, test source ownership, exact metadata,
network round trips, corrupt/truncated inputs, every limit and concurrent frozen
readers. Fuzz arbitrary bounded bytes through parse, canonical write, reparse
and semantic comparison without C.

Promote both BAF translation units only after focused tests, exactly 5,000
fuzzer executions, targeted race tests, the full ordinary suite, vet and all
source/policy audits pass. This checkpoint ends at 653 validated, 574 partial
and 530 unstarted files; ordinary tests pass in 75.165s and targeted race tests
in 1.027s. Wave 27's full race remains the current milestone gate, and all work
remains strictly single-threaded.

## Thirtieth serial-wave integration checkpoint

Port all active helpers in `src/base/io/ioReadDsd.c` into the concrete SOP
network boundary. Preserve balanced-parenthesis end discovery, homogeneous
top-level `*`/`+`/`,` splitting, repeated complements, the source preprocessing
that wraps the formula and doubles every original parenthesis, comma expansion,
parent-before-child node allocation, default AND/comma groups, the special
multi-input XOR marker and hexadecimal prime truth tables. Reproduce native
hex-digit variable inference, minterm/cube ordering, sparse lowercase-variable
counting, fake PI names, network name `dsd` and output name `F`. Normalize only
the empty zero-truth cover into the concrete SOP type's equivalent explicit
off-cover.

Require immutable input, atomic errors, matched complete groups, uniform
separators, prime arity agreement and formula/object/edge/fanin/depth/cube/work/
retained-byte limits with responsive cancellation. Pin exact helper outputs,
object counts and source order, node SOP text, hexadecimal cube order and sparse
names. Exhaustively compare representative formulas through the independent AIG
boundary, test malformed input and every limit, and exercise concurrent
independent readers. Fuzz generated nested formulas against truth words computed
during generation, without C.

Promote `ioReadDsd.c` only after focused tests, the exact 5,000-execution fuzz
invocation, targeted race tests, the full ordinary suite, vet, the scheduled
full-package race milestone and all source/policy audits pass. This checkpoint
ends at 654 validated, 574 partial and 529 unstarted files; ordinary tests pass
in 75.080s, targeted race tests in 1.027s and the full race suite in 395.900s.
The authoritative audit also moves the already validated Wave 28 JSON metadata
from an accidentally matched `aigInter.c` entry to `ioJson.c`, preserving the
same counts. All work remains strictly single-threaded.

## Thirty-first serial-wave integration checkpoint

Disposition `src/base/io/ioReadEdif.c` according to its actual reachable
contract. `Io_ReadEdif` unconditionally reports that the parser does not work
and returns before file access, option use, allocation or the call to its
file-local parser. Represent that public boundary as a typed unsupported result
over owned bytes, independent of check flags, malformed options and cancellation
callbacks. Demonstrate from the source call graph and pinned manifest hash that
`Io_ReadEdifNetwork` has no other caller and lies after the unconditional
return; do not turn dead, self-described broken code into invented EDIF support.

Test nil, empty, arbitrary and plausible EDIF bytes, both check modes, invalid
limits, untouched cancellation and concurrent calls. Fuzz arbitrary bytes,
flags and limits to prove the result remains unconditional and side-effect-free.
Keep the distinction explicit in status and coverage documentation so a
validated disabled disposition cannot be mistaken for format support.

Promote `ioReadEdif.c` only after focused tests, exactly 5,000 fuzzer
executions, targeted race tests, the full ordinary suite, vet and all source/
policy audits pass. This checkpoint ends at 655 validated, 574 partial and 528
unstarted files; ordinary tests pass in 74.775s and targeted race tests in
1.025s. Wave 30's full race remains the current milestone gate, and all work
remains strictly single-threaded.

## Thirty-second serial-wave integration checkpoint

Port the complete `src/map/fpga/fpgaTruth.c` contract with a concrete immutable
cut DAG. Retain the complement bit on each of two cut edges and the cut's output
phase. Accept an explicit ordered leaf boundary, matching `ppLeaves`, and derive
an owned native packed truth table rather than exposing a CUDD manager or BDD
node. Traverse iteratively in deterministic fanin-first postorder, memoize each
shared cut once, and compute structural volume as the unique reached non-leaf
count including the root. Do not retain the source's pointer-truncated `uSign`
cache, mutate `fMark`, print volume, or introduce a BDD/runtime interface.

Require two or more unique nonnil leaves, a non-leaf root, two inputs for every
reached internal cut, acyclic structure, atomic errors, cut/combined-word limits
and responsive cancellation. Pin exact exhaustive semantics, edge and output
complements, shared-DAG volume, ownership, malformed boundaries, missing inputs,
cycles, every limit, cancellation and a 10,000-node chain that proves recursion
has been removed. Exercise concurrent evaluation of one frozen DAG. Fuzz bounded
acyclic DAGs and compare every assignment with an independent Boolean model.

Promote `fpgaTruth.c` only after focused tests, exactly 5,000 fuzzer executions,
targeted race tests, the full ordinary suite, vet and all source/policy audits
pass. This checkpoint ends at 656 validated, 574 partial and 527 unstarted
files; ordinary tests pass in 75.585s and targeted race tests in 1.042s. Wave
33's full race is now the current milestone gate. Porting continues with only
the single supervisor thread active.

## Thirty-third serial-wave integration checkpoint

Port all three definitions in `src/opt/mfs/mfsWin.c` over the concrete
`ABCNetwork`. Accept the source's already computed object levels as an explicit
ID-indexed slice rather than silently trusting or mutating object metadata.
Starting from one logic node, select it as a root when its fanout multiplicity
exceeds the supplied limit, it drives any combinational output, or any fanout's
level is strictly greater than the start level plus the TFO window. Otherwise
visit node fanouts in source insertion order. Replace traversal-ID mutation and
recursive stack growth with private iterative state while preserving first-visit
DFS order and duplicate-edge fanout multiplicity.

Reject malformed or negative bounds, missing/stale level dimensions, non-node
starts, negative live-object levels and fanout paths that violate the source's
node/CO assumptions. Bound distinct visits and retained roots, poll cancellation
within both node and fanout traversal, return no partial roots on failure, and
leave both network and levels untouched. Pin strict cutoff semantics, CO and
fanout-limit roots, deterministic branch order, malformed cases, every limit,
every cancellation point, concurrency and a 10,000-node chain. Fuzz bounded
acyclic fanout graphs against an independent recursive model without C.

Promote `mfsWin.c` only after focused tests, exactly 5,000 fuzzer executions,
targeted race tests, the full ordinary suite, vet, the scheduled full-package
race milestone and all source/policy audits pass. This checkpoint ends at 657
validated, 574 partial and 526 unstarted files; ordinary tests pass in 75.745s,
targeted race tests in 1.062s and the full race suite in 398.582s. Porting
continues with only the single supervisor thread active.

## Thirty-fourth serial-wave integration checkpoint

Complete `src/aig/ivy/ivyFraig.c` as a native deterministic reduction and proof
engine. Use packed simulation to partition complement-normalized candidates,
then certify constant, equivalent and complemented merges through the concrete
native SAT orchestrator over full support. Return an owned reduced graph,
complete validity-tagged source mapping, certified classes, statistics and
replayable miter model. Proof mode must distinguish proved, counterexample and
unknown and replay every model. Replace the source's optional CUDD fallback with
complete unbounded native SAT; do not retain a runtime backend dependency. Pin
all native defaults and account for all 87 named definitions in the translation
unit, including source helpers consolidated into the semantic engine.
The completed coverage record must be `validated` with no residual partial
symbol marker; retain the exact definition inventory in Go so completeness
remains reviewable without consulting a runtime C oracle.

Port `src/base/abci/abcOrder.c` with transactional explicit and reverse CI
permutations plus stable support-first discovery followed by unused CIs. Permit
semantic terminal ordering in `ABCNetwork` validation while continuing to
require exact duplicate-free CI/CO sets. Test malformed permutations and owned
name bytes, source immutability, limits, cancellation and concurrent frozen
readers.

With full IVY FRAIG available, port `src/base/abci/abcBmc.c` by composing the
concrete ABC-to-AIG bridge, sequential timeframe expansion and FRAIG reduction.
Preserve symbolic versus constant-zero initial state and the source's unbounded
per-node SAT setting. Keep this API a reduction/reporting operation rather than
inventing a property result. Replace stdout diagnostics with owned aggregate and
per-frame structural counts, and cover initialization, limits, cancellation,
ownership and races.

Promote all three translation units only after fixed semantic/replay tests,
exact 5,000-execution CI-order and BMC fuzz campaigns, the IVY FRAIG campaign's
5,002 scheduled executions including baseline, targeted race gates, the full
ordinary suite, vet and source/policy audits pass. This checkpoint ends at 660
validated, 573 partial and 524 unstarted files. Ordinary tests pass in 75.533s;
the ordering/FRAIG race gate passes in 1.064s and the BMC race test itself in
1.029s. Wave 33's 398.582s full-package race remains current, with the next
scheduled full race at Wave 36. Porting continues strictly single-threaded.

## Thirty-fifth serial-wave integration checkpoint

Port all eleven definitions in `src/base/abci/abcXsim.c` over the concrete
strashed `ABCNetwork`. Preserve the source zero/one/unknown numeric encoding,
complemented-fanin inversion, conservative ternary AND, terminal and latch
order, symbolic input/state options and the exact reset `Gia_ManRandom` bit
stream. Use private value storage and return owned per-frame PI/latch/PO rows
instead of consuming object `pCopy` scratch and printing. Retain the source's
otherwise unused nonzero three-valued random helper and expose deterministic
0/1/x formatting as owned text.

Represent `Abc_NtkCycleInitState` as a pure operation returning an owned cloned
network with reached latch initializations. Zero frames preserve the source
state; random and symbolic input prefixes must update latches identically while
the input network stays frozen. Reject non-strashed graphs, malformed latch or
terminal shapes and unsupported boxes. Bound frames, objects, retained trace
entries and work, poll cancellation through traversal and simulation, and
return no partial trace or clone on failure.

Promote `abcXsim.c` only after exact random/trace unit tests, ownership, every
initialization and zero-frame behavior, malformed inputs, limits, cancellation,
concurrent frozen callers, exactly 5,000 generated-DAG fuzz executions against
an independent possibility-set model, targeted race tests, the full ordinary
suite, vet and all source/policy audits pass. This checkpoint ends at 661
validated, 573 partial and 523 unstarted files. The race gate passes in 1.069s
and ordinary tests in 74.948s. Wave 33's 398.582s full race remains current;
Wave 36 is the next scheduled full-package race. Porting remains strictly
single-threaded.

## Thirty-sixth serial-wave integration checkpoint

Port all ten definitions in `src/base/abci/abcMini.c` using the concrete native
`MiniAIG`, `AigManager` and `ABCNetwork` representations. In both directions,
preserve MiniAIG constant polarity, complemented fanins, structural hashing,
CI/CO order, dangling cleanup and the final-register-count PI/PO partition.
Replace `iTemp` and copy-field mutation with owned validity-checked mappings.
Retain the `MiniAig` network name and deterministic dummy PI/PO names; collapse
representable unary buffers without changing semantics.

Replace the opaque native frame calls with a concrete synchronized owner that
supports MiniAIG input, output, snapshot retrieval and flop-count repartitioning
without exposing mutable network state. Replace the process-file test routine
with an in-memory binary marshal/parse conversion round trip. Bound objects,
mapping storage and work, poll cancellation throughout conversion, and return
no partially installed frame or output on failure.

Promote `abcMini.c` only after fixed exhaustive combinational and sequential
tests, register/order/name and cleanup checks, ownership, malformed inputs,
limits, cancellation, concurrent direct and shared-frame callers, exactly 5,000
generated-DAG fuzz executions with exhaustive assignment comparison, targeted
race tests, the ordinary full suite, vet, the scheduled package-wide race and
all policy audits pass. This checkpoint ends at 662 validated, 573 partial and
522 unstarted files. The final targeted race passes in 1.028s, ordinary tests
in 75.143s and the scheduled full race in 395.682s. The next full-package race
is scheduled for Wave 39; work remains strictly single-threaded.

## Thirty-seventh serial-wave integration checkpoint

Port both definitions in `src/base/abci/abcLog.c` as owned-byte operations.
Preserve the three-line `snl_SAT`, `snl_UNSAT` and `snl_UNK` grammar, bug-free
depth, engine token, optional property and counterexample-frame fields, initial
register state and frame-major input bits. Replace filename I/O and stdout
diagnostics with deterministic bytes and typed errors, and clone all retained
counterexample state.

Correct two unsafe native parser behaviors explicitly: paired `NULL` payloads
remain nil rather than becoming a fabricated zero-width counterexample, and
characters outside the representable binary alphabet fail instead of shifting
later bits by silent omission. Support register-only, zero-input traces. Reject
malformed field counts, numeric ranges, mixed NULL/data payloads and inconsistent
frame dimensions. Bound bytes, frames, bits and engine text and poll
cancellation without returning partial records.

Promote `abcLog.c` only after exact byte and round-trip tests for every status,
optional SAT fields, zero-input traces, ownership, malformed records, every
limit, cancellation, concurrent readers, exactly 5,000 generated-record fuzz
executions and exactly 5,000 arbitrary-byte parser executions, targeted race
tests, the ordinary suite, vet and all policy audits pass. This checkpoint ends
at 663 validated, 573 partial and 521 unstarted files. The targeted race gate
passes in 1.026s and ordinary tests in 74.983s. Wave 36's 395.682s full race
remains current; Wave 39 is next. Work remains strictly single-threaded.

## Thirty-eighth serial-wave integration checkpoint

Port the sole definition in `src/sat/bsat2/AbcApi.cpp` without inventing solver
behavior absent from the source. `Abc_CallMiniSat22` constructs an unused local
solver, does not inspect `Cnf_Dat_t`, and returns `-1` for every argument. Expose
that exact dormant contract as `ABCCallMiniSat22`, accept nil as the native code
does, and define the result constant explicitly. Do not redirect the call to the
working Go SAT solver because that would change the API's observable behavior
and obscure that the upstream entry point is only a stub.

Pin nil and nonnil results, CNF immutability and concurrent frozen callers.
Generate bounded owned CNFs for exactly 5,000 fuzz executions and confirm that
every call returns the native constant without retaining or changing formula
state. Promote `AbcApi.cpp` only after focused tests, the targeted race gate,
the ordinary full suite, vet and all source/policy audits pass. This checkpoint
ends at 664 validated, 573 partial and 520 unstarted files. The targeted race
passes in 1.024s and ordinary tests in 74.744s. Wave 36's full race remains
current; Wave 39 is next. Work remains strictly single-threaded.

## Thirty-ninth serial-wave integration checkpoint

Port both definitions in `src/opt/sbd/sbd.c` as checked owned operations.
Preserve `Sbd_CountConfigVars` level order, degree constraint, suffix-based
formula and its zero result when the level population is below `nVars - 1`.
Generalize the fixed `Sbd_CountTopos` experiment to explicit input, node and
degree arguments while retaining its breadth-by-new-generation traversal.
Intern level vectors with `insdict.Dict` so uniqueness and insertion order are
deterministic, and return owned stage/configuration records instead of stdout.

Reject empty or malformed configurations and invalid experiment dimensions;
detect every integer overflow. Bound levels, unique configurations, retained
level entries and work, poll cancellation throughout enumeration and counting,
and return no partial result on failure. Pin the original 9/10/3 experiment's
stage totals, exact small insertion order, ownership, all limits, cancellation
and concurrent callers. Fuzz exactly 5,000 bounded experiments against an
independent model using linear duplicate detection rather than the production
dictionary.

Promote `sbd.c` only after focused tests, the targeted race gate, ordinary full
suite, vet, the scheduled full-package race and all source/policy audits pass.
This checkpoint ends at 665 validated, 573 partial and 519 unstarted files. The
final targeted race passes in 1.032s, ordinary tests in 74.961s and the full race in
396.785s. Wave 42 is the next full-package race; work remains strictly
single-threaded.

## Fortieth serial-wave integration checkpoint

Port both definitions in `src/bdd/llb/llb2Dump.c` as reentrant owned
operations. Preserve `Llb_ManGetDummyName` zero padding and the reached-state
dump's `ff` input names, `Reached` output and Boolean function. Compose the
existing native Go BDD manager and BLIF exporter, returning bytes instead of
writing a named process file. Keep the fixed-order manager unchanged instead
of performing the source's in-place heap reorder; this retains dump semantics
and permits concurrent readers of frozen managers.

Reject invalid prefixes, dimensions, managers, literals and model names. Bound
bytes, BDD nodes and work, poll cancellation during naming and export, return no
partial output and ensure every result owns its storage. Pin exact names and
BLIF fragments, exhaustive truth-table equivalence, ownership, every error and
limit, cancellation and concurrent readers. Fuzz exactly 5,000 bounded truth
tables using standalone semantic evaluation. Add direct shared-exporter
regressions for the newline between `.outputs` and `.names` and for enforcing
node limits at discovery time.

Promote `llb2Dump.c` only after focused LLB/CUDD tests, the targeted race gate,
the ordinary full suite, vet and all source/policy audits pass. This checkpoint
ends at 666 validated, 573 partial and 518 unstarted files. The targeted race
passes in 1.064s, ordinary tests in 75.557s and vet is clean. Wave 39's
396.785s full-package race remains current; Wave 42 is next. Work remains
strictly single-threaded.

## Forty-first serial-wave integration checkpoint

Port both definitions in `src/bdd/llb/llb2Bad.c` over the concrete sequential
AIG and canonical BDD managers. Traverse the primary-output cones in postorder,
translate constants, CIs, buffers, ANDs and complemented edges, and disjoin all
primary outputs. Preserve the native BDD variable layout exactly: latch outputs
occupy the first variables and primary inputs follow, even though `AigManager`
stores its CIs in the reverse partition order. Existentially abstract precisely
the primary-input variables in `Llb_BddQuantifyPis`.

Validate graph and manager dimensions, perform BDD construction in one
transaction, and return no root after failure. Bound AIG objects, traversal
events, BDD operations and memory; poll cancellation through traversal,
construction and abstraction. Treat the native timeout parameter according to
its actual source behavior—the timed conjunction is commented out—while
providing explicit checked cancellation instead.

Promote `llb2Bad.c` only after exhaustive independent AIG-versus-BDD semantics,
constants, empty outputs, deep iterative buffer traversal, malformed inputs,
limits, cancellation, rollback, concurrent independent managers, exactly 5,000
generated two-output truth-table fuzz executions, the targeted race gate, the
ordinary full suite, vet and all source/policy audits pass. This checkpoint ends
at 667 validated, 573 partial and 517 unstarted files. The targeted race passes
in 1.070s, ordinary tests in 75.853s and vet is clean. Wave 39's 396.785s
full-package race remains current; Wave 42 is next. Work remains strictly
single-threaded.

## Forty-second serial-wave integration checkpoint

Port all three definitions in `src/opt/sbd/sbdCnf.c` over concrete truth-table,
ISOP and SAT-literal types. Convert a zero- through six-input packed truth word
into owned clauses for the direct and complement covers, preserving native cube
order, negated cube literals and the final output literal. Return the native
zero-based diagnostic form as bounded owned text. Translate each local literal
through an owned fanin-variable map and complement literals whose mapped
variable equals the pivot.

Replace the signed-byte/sentinel representation with structured clauses.
Deliberately omit `Sbd_TranslateCnf`'s final unused empty vector level because
passing it to a solver would assert an empty clause. Correct the constant
shortcut for nonzero input counts so it constrains the actual output variable;
retain the exact zero-input encoding. Validate canonical packed truth, every
literal and map entry, dimensions and CNF structure. Bound clauses, literals,
text, ISOP storage and work; poll cancellation and expose no partial result.

Promote `sbdCnf.c` only after exact constant/order/text tests, exhaustive
zero-through-six-input clause equivalence, translation semantics, ownership,
malformed inputs, every limit, cancellation, concurrent callers, exactly 5,000
standalone truth/translation fuzz executions, the targeted race gate, ordinary
full suite, vet, the scheduled full-package race and all source/policy audits
pass. This checkpoint ends at 668 validated, 573 partial and 516 unstarted
files. The targeted race passes in 1.057s, ordinary tests in 75.120s and the
full race in 395.760s. Wave 45 is the next full-package race; work remains
strictly single-threaded.

## Forty-third serial-wave integration checkpoint

Port all five definitions in `src/opt/sbd/sbdPath.c` over concrete
`GiaManager`, `GiaLUTMapping` and `TimManager` values. Compute ordinary mapped
LUT levels and canonical timing-box levels, then reproduce the source's
slack-one critical-path recursion with output-major traversal and one global
first-visit mark. Treat each non-`TimNoArc` box entry as a unit dependency, as
the source does, rather than introducing a different weighted-delay algorithm.

Retain the observable idempotence of `Sbc_ManAddInternalToPath`: the source
marks mapped LUT boundaries before its recursive visit reaches an already
marked root, so it does not add internal AIG nodes. Return an owned clone rather
than mutating caller storage. Replace delay-trace printing with owned path and
aggregate statistics. Use iterative traversal, validate graph/mapping/timing
dimensions, bound objects and visits, poll cancellation and return no partial
result on failure.

Promote `sbdPath.c` only after exact ordinary and timing-box paths, ownership,
malformed inputs, every limit, cancellation, a 10,000-LUT chain, reconvergent
output-order behavior, concurrent shared frozen readers and an independent
generated-DAG model fuzzer passing exactly 5,000 executions. The targeted race
gate passes in 1.071s, ordinary tests in 75.100s and vet is clean. This
checkpoint ends at 669 validated, 573 partial and 515 unstarted files. Wave
42's 395.760s full-package race remains current; Wave 45 is next. Work remains
strictly single-threaded.

## Forty-fourth serial-wave integration checkpoint

Port all three definitions in `src/base/abci/abcBidec.c` by composing the
concrete native `HopManager`, `TruthTable` and `BdcManager`. Convert each local
HOP root to its packed truth table, run BDC decomposition, assign constant and
PI copies, then rebuild every complemented two-input BDC node through HOP
structural hashing. Provide both the per-node operation and the source's
network-level loop over an owned sequence of local roots. Preserve its
15-input cutoff and return owned per-node and aggregate gain statistics instead
of printing them.

Reproduce the power-aware care rule exactly: compute `2*p*(1-p)`; below 0.4,
retain the incompletely specified care set, while at or above 0.4 complete
don't-cares to zero when `p <= 0.5` and to one otherwise before full-care
decomposition. Never mutate the caller's care table. Validate roots,
dimensions, probabilities and truth storage; bound roots, truth words, BDC
nodes and new nodes; poll cancellation; and roll back all HOP nodes created by
a failed node or batch.

Promote `abcBidec.c` only after exact cared and completed semantics, ownership,
wide-node retention, malformed inputs, every limit, cancellation, batch
atomicity, a 10,000-node cone, concurrent independent managers and an
independent generated-HOP model fuzzer passing exactly 5,000 executions. The
targeted race gate passes in 1.070s, ordinary tests in 75.182s and vet is clean.
This checkpoint ends at 670 validated, 573 partial and 514 unstarted files.
Wave 42's 395.760s full-package race remains current; Wave 45 is next. Work
remains strictly single-threaded.

## Forty-fifth serial-wave integration checkpoint

Port both definitions in `src/base/abci/abcReorder.c` over concrete owned local
nodes containing fanin IDs and packed truth tables. Replace the optional CUDD
backend with the validated deterministic native BDD ordering engine. Before
ordering, tie repeated fanin IDs and remove functionally unused variables, as
the source network preprocessing does. After ordering, permute the fanin list
and apply the inverse source-to-target truth permutation so evaluation by
external signal ID remains identical. Return owned per-node and aggregate
statistics rather than mutating opaque data or writing verbose diagnostics.

Support source truth widths through the packed table boundary and perform
minimum-base reduction before imposing the dense ordering engine's explicit
20-variable retained-support limit. Validate names, fanin IDs, dimensions and
truth storage; bound network nodes, original fanins, truth words, reorder
evaluations and passes; poll cancellation throughout table construction and
search; and expose no partial network result after failure.

Promote `abcReorder.c` only after a fixed strict-cost reduction, exhaustive
external-signal equivalence, duplicate tying, minimum-base and constant cases,
wide-small-support handling, ownership, malformed inputs, every limit,
cancellation, concurrent immutable readers and an independent model fuzzer
passing exactly 5,000 executions. The targeted race gate passes in 1.072s,
ordinary tests in 75.418s, vet is clean, and the scheduled full-package race
passes in 407.673s. This checkpoint ends at 671 validated, 573 partial and 513
unstarted files. Wave 48 is the next full-package race; work remains strictly
single-threaded.

## Forty-sixth serial-wave integration checkpoint

Port all three definitions in `src/base/abci/abcSense.c` by composing the
concrete `AigManager`, simple Tseitin encoder and native CDCL solver. Build two
cofactor copies for each selected CI without recursion. XOR each ordinary
output's cofactors, OR those differences, and finally AND the result with both
cofactors of the last output, preserving the source's special constraint
interpretation. Retain the selected CI as an unused miter input, matching the
source terminal boundary, and return a checked owned one-output AIG.

For network sensitivity, solve the asserted miter once per CI. Return explicit
SAT, UNSAT and unknown records plus owned source-CI models; replay every SAT
model before exposing it. Include SAT and resource-limited unknown inputs in
the sensitive list, exactly as the source does. Use complete native SAT for
unbounded calls rather than retaining the multi-engine Ivy dependency. Reject
sequential graphs, validate every graph and literal, bound inputs, objects,
work, variables, clauses and literals, combine cancellation with SAT limits,
and return no partial result after failure.

Promote `abcSense.c` only after exhaustive miter truth, an independent
cofactor-sensitivity model, two-sided constraints, model replay and ownership,
malformed inputs, every limit, cancellation, a retained 10,000-node cone,
concurrent frozen readers and exactly 5,000 single-core generated-AIG fuzz
executions. The targeted race gate passes in 1.048s, ordinary tests in 75.438s
and vet is clean. This checkpoint ends at 672 validated, 573 partial and 512
unstarted files. Wave 45's 407.673s full-package race remains current; Wave 48
is next. Work remains strictly single-threaded.

## Forty-seventh serial-wave integration checkpoint

Port both definitions in `src/aig/saig/saigOutDec.c` by composing the concrete
`AigManager`, simple Tseitin encoder and native CDCL solver. Collect the first
property output's cone in native postorder. Test phase-relative one-literal
candidates first, mark successful nodes, then test unmarked pairs in the
source's exact three polarity modes. Treat only UNSAT as a prime; count and
skip bounded unknown results. Return owned cubes and typed query statistics
instead of source stdout diagnostics.

Duplicate the graph without source scratch: create all CIs, internal nodes,
original primary outputs, prime-cube outputs and latch inputs in that order,
then restore register and constraint counts. Bound graph traversal, candidates,
pairs, SAT queries, cubes, CNF dimensions and aggregate work; combine caller
cancellation with SAT cancellation and return no partial result after failure.

Promote `saigOutDec.c` only after independent exhaustive truth-table checks of
prime order and polarity, decomposed-output replay, sequential partition and
ownership checks, malformed inputs, every limit, cancellation, a retained
10,000-node graph, concurrent frozen readers and exactly 5,000 single-core
generated-AIG fuzz executions. The targeted race gate passes in 1.179s,
ordinary tests in 75.027s and vet is clean. This checkpoint ends at 673
validated, 573 partial and 511 unstarted files. Wave 45's 407.673s full-package
race remains current; Wave 48 is next. Work remains strictly single-threaded.

## Forty-eighth serial-wave integration checkpoint

Port all five definitions in `src/aig/saig/saigDual.c` using the concrete
`AigManager`. Encode each dual rail as definitely-zero and definitely-one,
with both false representing unknown. Preserve selected-PI dual expansion,
ordinary PI encoding, global and per-register initialization choices,
complemented-fanin care construction, AND propagation, zero/one/X violation
checks, PO-versus-RI miter selection, output complementation and doubled latch
ordering. Preserve the source's zero-check priority if both flags are set.

Replace the mutating property-blocking API with atomic owned reconstruction,
while retaining the exact transient primary-output boundary used by the C
routine as delay-register COs are appended. Rebuild in topological order so
the result satisfies the Go manager's stronger checked invariant. Bound source
and destination objects plus aggregate work, poll cancellation throughout and
return no partial graph after failure.

Promote `saigDual.c` only after exhaustive independent ternary evaluation of
every option combination and raw rail assignment, independent blocking-output
evaluation, sequential partition and ownership checks, malformed inputs,
every limit, cancellation, a retained 10,000-gate graph, concurrent frozen
readers and exactly 5,000 single-core combined-transform fuzz executions. The
targeted race gate passes in 1.115s, ordinary tests in 75.854s, vet is clean
and the scheduled full-package race passes in 398.659s. This checkpoint ends
at 674 validated, 573 partial and 510 unstarted files. Wave 51 is the next
full-package race; work remains strictly single-threaded.

## Forty-ninth serial-wave integration checkpoint

Port all six definitions in `src/aig/saig/saigIoa.c` as byte-oriented,
bounded input/output for the source's restricted structural sequential-BLIF
dialect. Preserve padded object and terminal namespaces, source-shaped header
comments and statistics, PI/PO/latch declarations, conditional constant-one
emission, object-number fanin references, AND topology, complemented cube
polarity and PO-before-LI ordering. Provide memory APIs for deterministic tests
and path wrappers for the source file entry points.

Replace static token buffers, unchecked numeric arrays and assertion/print
failure paths with owned tokens, checked references and typed errors. Accept
the exact emitted grammar, validate regular AND results and reconstruct a
checked manager without carrying constraints the format cannot represent.
Bound input/output bytes, objects, tokens, token size and aggregate work; poll
cancellation throughout and expose no partial graph or byte slice on failure.

Promote `saigIoa.c` only after byte-stable memory and path round trips,
exhaustive Boolean output comparison, exact name/metadata checks, malformed
input coverage, every limit, cancellation, a retained 10,000-gate graph,
concurrent frozen readers and exactly 5,000 single-core combined arbitrary-
parser/generated-round-trip fuzz executions. The targeted race gate passes in
1.278s, ordinary tests in 75.431s and vet is clean. This checkpoint ends at
675 validated, 573 partial and 509 unstarted files. Wave 48's 398.659s
full-package race remains current; Wave 51 is next. Work remains strictly
single-threaded.

## Fiftieth serial-wave integration checkpoint

Port all six definitions in `src/aig/saig/saigIsoFast.c` as bounded fast
structural signatures and candidate output classes. Preserve the packed 16-bit
node data, native bitfield truncation, modulo-four fanout counters,
complement/level fanin ordering, dynamic LO-to-LI traversal, postorder
first-occurrence order and exact ABC cost permutation. Reproduce the pinned
little-endian bytewise integer-vector `memcmp` used for sorting and grouping;
do not replace it with ordinary numeric slice comparison.

Use iterative cone traversal to remove recursion depth risk. Return owned
per-output signatures and the classes the experimental C driver prints and
frees before returning `NULL`. Reject buffers explicitly because the native
collector asserts that every visited nonterminal is an AND. Bound objects,
visits, dynamic roots, signature entries and aggregate work, poll cancellation
throughout and return no partial analysis after failure.

Promote `saigIsoFast.c` only after manually derived packed-word checks,
independent sequential class construction, constant-output and ownership
coverage, malformed graphs, every limit, cancellation, a retained 10,000-gate
cone, concurrent frozen readers and exactly 5,000 single-core generated-
sequential-AIG fuzz executions against a separate recursive model. The
targeted race gate passes in 1.072s, ordinary tests in 75.652s and vet is clean.
This checkpoint ends at 676 validated, 573 partial and 508 unstarted files.
Wave 48's 398.659s full-package race remains current; Wave 51 is next. Work
remains strictly single-threaded.

## Fifty-first serial-wave integration checkpoint

Port `Fra_InvariantVerify`, the sole definition in
`src/proof/fra/fraIndVer.c`, as a bounded owned K=1 invariant verifier. Map
source AIG object literals through the native CNF encoding without mutating
the caller's clauses. For the base case, constrain latch outputs to zero and
solve the negation of every invariant clause. For induction, compose two
uninitialized copies of the relation, equate each first-frame latch input to
its second-frame latch output, assert every invariant clause in frame zero and
solve each clause's negation in frame one. Preserve SAT and unknown as explicit
failures, clone and replay-check SAT models, and distinguish an inconsistent
induction premise from successful verification.

Return `ErrUnsupported` outside K=1 rather than carrying forward the source's
accidental success return. Bound objects, variables, clauses, literals and
aggregate work with overflow-safe accounting; honor both operation and SAT
cancellation. Promote the source only after independent exhaustive state/input
enumeration, base/induction/model/ownership checks, empty-clause inconsistency,
deterministic unknown, malformed input and every resource boundary, a retained
10,000-gate graph, sixteen concurrent readers and a 5,000x single-worker
generated sequential-AIG fuzz campaign. The targeted race gate passes in
1.454s, ordinary tests in 75.274s, vet is clean and the scheduled full-package
race passes in 398.681s. This checkpoint ends at 677 validated, 573 partial
and 507 unstarted files. Wave 54 is the next scheduled full-package race; work
remains strictly single-threaded.
