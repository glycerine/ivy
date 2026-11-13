# Abstraction source inventory

This is an index for checking the active ABC C implementation against the Go
port. The source files are `src/proof/abs` in the parent repository. Go code
is split by algorithm; the representative files below are entry points, not an
exhaustive list of helpers or a claim of instruction-by-instruction equivalence.
Tests and remaining fidelity limits are recorded in `PORTING_REVIEW.md`.

| C source | Main Go implementation | Active behavior |
| --- | --- | --- |
| `abs.c` | `absTypes.go`, `options.go` | Declarations and configuration; the C file has no active functions. |
| `absDup.c` | `absDup.go`, `aig/gia/man_abstraction.go` | Gate/flop maps, collecting and duplicating abstractions. |
| `absGla.c` | `absGlaDriver.go`, `absGla*.go` | Current gate abstraction, CNF, solving, refinement, reporting, and bridge. |
| `absGlaOld.c` | `absGlaOldDriver.go`, `absGlaOld*.go` | Original mapped-cut gate abstraction and refinement. |
| `absIter.c` | `absIter.go` | Bounded abstraction shrinking. |
| `absOldCex.c` | `absOldCba*.go`, `absOldCex*.go` | Counterexample-based abstraction and ternary verification. |
| `absOldRef.c` | `absOldRef*.go`, `absOldParams.go` | Legacy flop abstraction, BMC/BDD/PDR selection and refinement. |
| `absOldSat.c` | `absOldSat*.go` | Sequential unrolling, SAT and care-set refinement. |
| `absOldSim.c` | `absOldSim*.go`, `absOldTernary.go` | Legacy simulation and justification. |
| `absOut.c` | `absOut.go`, `absCex*.go`, `absTransformFlops.go` | One-layer and fast-forward refinement, witness replay and state transforms. |
| `absPth.c` | `absPth.go`, `absPthControl.go`, `absPthPdr.go` | Concurrent optional PDR prover and cancellation. |
| `absRef.c`, `absRef.h` | `absRef.h.go`, `rnm_*.go` | Ternary refinement manager, collection, simulation and justification. |
| `absRefJ.c`, `absRefJ.h` | `rf2_*.go`, `refine_cover.go` | Alternate justification/refinement manager and cover operations. |
| `absRefSelect.c` | `absRefSelect.go` | Refinement filtering and fanout selection. |
| `absRpm.c`, `absRpmOld.c` | `absRpm*.go` | Reparameterization, support, dominators, and truth computations. |
| `absUtil.c`, `abs.h` | `absUtil.go`, `absTypes.go` | Defaults, VTA/GLA/FLA conversion, and shared records. |
| `absVta.c` | `absVta.go`, `vta_*.go` | Variable-timeframe abstraction, solver, core, rollback and reporting. |

A source-name scan found that the active non-macro function names are present
except for the following translated internal helpers: CBA/SAT recursive cone
collection shares `oldUnrollWithCex`; the CBA/SAT reason recursion shares
`oldRefinementReason`; ternary mark accessors and simulation are represented by
local values in `Saig_ManCexVerifyUsingTernary`; and `Abs_ProverThread` is
represented by `absRunPDR` and `absProverControl`. C's `Saig_ObjCexMinPrint`
is referenced only inside a commented debug block. `Pdr_ManSolve_test` is an
unused C test routine; it is not a solver implementation. The UNSAT-core
refinement prototype and its SAT-variable helpers in `absRef.c` are under
`#if 0` and are not compiled.

The C `ABC_USE_PTHREADS` conditional has a no-op prover branch when pthreads
are disabled; the Go implementation provides the concurrent PDR path. GLA
uses a labeled full proof graph for core extraction, including when C selects
its alternate proof mode. Standalone VTA requires an AND-inverter graph; its
native XOR/MUX guard returns unknown without changing object classes. See
`PORTING_REVIEW.md` for the source VTA mask assertion and retained-witness
boundary behavior.

`Abs_Par_t` options are exposed by `Options`: frame and conflict bounds,
learned-clause limits, timeout, ratio and restart settings, terminal variables,
rollback, fanout/layer/new-refinement selection, skip/simple/hash modes,
proof mode, dumping, concurrent prover and verbosity. `LegacyOptions` exposes
the corresponding legacy abstraction parameters. Mapping an option into a
record does not by itself prove equivalent runtime behavior: source C itself
initializes some legacy fields without using them in this package, and the
full-proof mode difference above remains open.

A declaration-name scan of `abs.h` found 29 public declarations. After the
collection and bounded-shrink tests, every non-diagnostic declaration is named
by at least one `proof/abs` test. The three exceptions are output-only
`Gia_ManPrintFlopClasses`, `Gia_ManPrintGateClasses`, and
`Gia_ManPrintObjClasses`, which delegate to `aig/gia` printing helpers and do
not have direct output-format tests in this package. A name occurrence does
not establish full branch coverage; resource limits, large proofs and
cross-package differential testing remain coverage gaps.
