# Native simulation port

## Scope and interfaces

All eleven original algorithm/header files are native Go. The package imports
AIG's `Graph` and leaf Satoko, without importing ABC networks or FRAIG. Original
public `Sim_*` / `Sym_*` entry points remain, with typed slices, bool options and
manager-aware `Graph, Edge` arguments replacing pointer/vector casts. `Info` rows
are owned uint32 words. Structural support is indexed by node ID; functional
support and symmetry matrices are indexed by CO, exactly as the implementation
of the C routines (despite the structural routine's misleading synopsis).

Implemented families:

* Structural support union and functional-support simulation, unresolved target
  queues, SAT witnesses and distance-one counterexample neighborhoods.
* Implication-supergate structural symmetry proofs, compatibility propagation,
  single-flip simulation squares, bitset matrix rejection, exact cofactor SAT,
  equivalence-class proof/disproof propagation and total symmetry counts.
* Word simulation with complemented edges, mixed baseline/alternate images,
  unary transfers, bit predicates/differences/new bits and pair statistics.
* Random/model sequential simulation, initial state and frame-to-frame RO/RI
  transfer, and independent-pattern switching estimates `2p(1-p)`.

`Symmetries` additionally returns the complete per-CO matrix rather than only a
count. Diagonals are true; non-diagonal results include only functional-support
variables, matching the original counting convention. Register CIs/COs trail the
ordinary inputs/outputs. Graphs remain read-only throughout the algorithms.

## Source comparison and deliberate differences

`test_vectors/sim/oracle.c` includes the actual source `simUtils.c`,
`simSwitch.c`, `simSupp.c`, `simSymStr.c`, `simSymSim.c` and bit-matrix code.
Its real ABC network/object/vector fixtures compare 200 word-operation cases,
200 symmetry-square rejection cases (including >32 CIs), and 100 randomly
constructed AND-DAG structural-support/symmetry cases. No mock algorithm or
replacement C implementation is used. ELF section collection discards unrelated
ABC application dependencies. A second test uses the original `abc` executable
for 12 complete functional-support and simulation/FRAIG symmetry-count cases.
It skips that extra integration test when the executable is absent; the compiled
source-stage oracle remains independent of the executable.

The old ABC/FRAIG cofactor miter is replaced by an exact two-copy Satoko CNF miter
with shared unchanged CI variables. Queries have no conflict limit; unresolved
results panic instead of being classified as proofs or counterexamples. This
changes the SAT backend and scheduling, not dependency or swap semantics. All
256 three-variable Boolean functions and 40 random four-input, multi-output
AND/XOR graphs are checked against exhaustive evaluation, including independently
replayed SAT witnesses and full symmetry workflow results.

The source's global libc `rand` sequence is replaced by per-manager local AIG RNG
state; standalone `Sim_UtilSetRandom` uses Go's concurrency-safe generator.
Results are deterministic per manager and concurrent analyses do not share RNG or
graph marks. Exact random word streams are intentionally not libc-compatible.

The separate `Sim_ComputeFunSuppSourceWithSolver` entry point preserves the
`simSupp.c` random-round, target iteration, SAT witness, and FIFO replay schedule.
The C-facing `abci.Sim_ComputeFunSupp` supplies the original cofactor-miter/FRAIG
proof callback. Its private glibc-compatible random generator reproduces the
`srand(0xABC)` stream without changing libc's process-global RNG state; that
process-global side effect is the remaining compatibility difference. The
no-callback entry point uses the package's exact cofactor solver for native Go
callers.

Native Graph stores zero-initialized registers. The compatibility sequential
wrappers use that convention; `SequentialSimulation` accepts explicit initial
0/1/unknown values, with unknown states randomized as in C. Explicit XOR and
buffer nodes are simulated correctly; XOR is a conservative structural-proof
boundary and exact SAT completes its symmetry pairs. DFS order handles forward
references introduced by rewiring, dangling nodes and deleted-ID holes. Constant
words are initialized, unlike the original switching routine's uninitialized
constant row. Switching deliberately rounds sample counts to whole 32-bit words
as C does; it is an independent-pattern estimate, not temporal activity.

## Separate design/readability review

Reviewed data ownership, API meanings, algorithm boundaries and hot loops after
the initial native port. Changes made during this pass:

* Replaced full alternate-image copying for every CI with source-style
  transitive-fanout recomputation and mixed baseline/alternate fanin reads.
* Replaced pair-by-pair simulation rejection with the original word-parallel
  column/row bitset OR algorithm.
* Reused one merged class list for non-symmetry propagation, removed a second
  complete support computation from the count wrapper, and used the O(1) graph
  object-slot count instead of copying the node table merely to measure it.
* Kept pair indices as a typed `Pair`, avoiding C's 16-bit packing restriction;
  avoided allocator pools, node scratch unions and float-through-int type punning.

Storage remains bounded by the source's simulation rows and per-output quadratic
symmetry matrices. The source's 1000 random-square rounds are retained. Large
circuits can still require expensive exact SAT completion; no approximation is
returned as an exact result.

## Separate bug review

Reviewed complement polarity, cofactor variable sharing, witness reconstruction,
class closure, matrix padding, sequential layout/initialization, empty graphs,
wide neighborhoods and retained manager state. Found and fixed a native bug:
`Sim_UtilCountSuppSizes` indexed absent structural data in lightweight managers.
`TestLightweightManagerSupportStatistics` was observed failing with an index
panic, then passed after the missing-data case was handled.

Regression coverage additionally checks:

* Positive and negative cofactor SAT results and replay of both support and swap
  witnesses, including rare 24-CI minterms that require SAT completion.
* Non-support variables from absorption and full class closure without random
  rejection (two parity variables versus two conjunction variables).
* >64-word FIFO neighborhoods safely truncate flips instead of writing beyond
  a row; 70-CI support retains the correct last-word bit.
* Forward-ID buffer rewiring, mixed-image simulation and explicit XOR gates.
* Zero-CI constants, zero frames, zero/one/unknown latch initialization,
  consecutive frame storage, multi-word state transfer and switching constants.
* Parallel support/symmetry/sequential analyses leave the graph unchanged under
  the race detector.

## Validation

Run from the repository root (no experiments included):

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -count=1 ./opt/sim
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 ./opt/sim
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./opt/sim
```

All three ordinary package gates pass. No source drafts are hidden, and no
production stubs or unsafe pointer emulation remain in this package.
