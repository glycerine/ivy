# RSB integration review

## Algorithm/API group ported

Ported the complete active `rsbMan.c` and `rsbDec6.c` group: reachability-pattern
primitives, single- and multiword divisor consistency checks, precise conflicting
minterm pairs, counterexample recording/initialization, the guided search over
singletons/pairs/triples/quadruples, selected-fanin access, old-fanin verification,
six-variable resubstitution, verification and diagnostics. The original maximum
64-counterexample search bound and candidate enumeration/tie order are retained.

Vectors become Go slices, counts/dimensions use int32, truth words use uint64.
`Rsb_ManGetFanins`/`GetFaninsOld` return borrowed slices; `SetFaninsOld` supplies
an explicit copying setter because appending to a returned Go slice cannot
update its owner's length. The general engine accepts multiword truth tables;
the public `PerformResub6` wrapper checks its six-variable/single-word contract.
The only package dependency is misc/util for diagnostic ISOP construction.

## Separate compactness, idiomatic Go and readability audit

Reviewed the implementation after the algorithm port. The four nearly identical
C nested-loop stages share one bounded subset enumerator, preserving their exact
ordering (outer highest candidate ascending, remaining IDs descending in each
selected tuple). Pattern encoding uses one recursive helper behind the familiar
Try0..Try5 APIs. Temporary pattern/divisor arrays use bounded Go storage; manager
scratch slices derive logical dimensions from their lengths, with no C vector
size/capacity mirrors. Configuration bounds and the local counterexample count
remain because they represent actual algorithm budgets, not duplicate lengths.

## Separate bug and boundary review

Reviewed CEX orientation across words, partial reachability/don't-cares, reused
scratch storage, all four subset sizes, the 64-bit shift boundary, failure results,
verification and verbose diagnostics. Fixes/regressions:

- Failed searches clear selected fanins instead of leaving a successful prior
  result or the final attempted candidate visible. Output truth pointers remain
  untouched on failure; a repeated-manager regression checks both properties.
- The source verifier prints an error but always returns 1. The Go verifier now
  returns 0 for a mismatched composition; a deliberately corrupted result is
  checked alongside a valid decomposition.
- Verbose diagnostic ISOP receives a real cube counter instead of the source's
  NULL pointer, which its recursive ISOP base case increments. A verbose-path
  regression verifies successful printing without a crash.
- Constant functions do not have true/false distinguishing minterms. CEX
  initialization returns zero rather than indexing a negative bit position.
- Decomposition budgets above four retain the source's effective four-input
  search behavior; they do not accidentally invoke an unsupported five-input
  pattern routine.

## Validation evidence

Compiled unchanged original rsbMan.c/rsbDec6.c with LTO into a native oracle and
hashed status, both output truth words, fanin counts and exact ordered fanin IDs
for all 256 three-variable functions. The Go full-search hash matches exactly.
A second oracle compiled the unchanged original DecTry/DecCheck routines and
verified 1,024 cases including exact ordered counterexamples; its permanent Go
hash regression also matches. C compilers are not required to run Go tests.

Independent tests cover minterm reconstruction for all 256 three-variable
functions, the original six-input AND/OR example, four-input parity selection,
100 random four-divisor compositions, multiword/cross-word consistency, all
pattern arities, old fanins, failure reuse, and exactly 64 CEX records without
bit overflow. Normal tests, package vet and combined RSB/Kit/util/BDC race tests
pass. No starting-point files or git operations are involved.
An additional independent partition oracle checks both reachable patterns and
conflict witnesses for random 7-, 8-, and 9-variable tables with 1..4 divisors,
including both arbitrary targets and deliberately decomposable compositions.
