# Frontier representation and direct simulation callers

Review the five routines in `src/aig/gia/giaFront.c` against `giaFront.go`.
Retain the source rotating slot allocator, cross-cut sizing, object traversal,
reference lifetimes and inverse slot-to-object transform. Also inspect and
repair direct gate/terminal consumption in `giaSim.c`, `giaSwitch.c` and
`giaTsim.c` ports. Those three larger source families still need their own full
reviews; this document closes only the frontier representation and its direct
consumer boundary.

## Failing evidence and repairs

A buffer contributes one reference in both C and Go reference construction,
but frontier counting and allocation decremented its two identical stored pins.
The minimal buffer regression first fails on input3/output1 after slot reuse;
cross-cut reports3 instead of4, and reference scratch underflows
(`/tmp/abcgo-front-buffer-red.log`). Decrement the logical buffer edge once in
both loops. The fixture's --buffer mode executes unmodified C, reproduces its
wrong output against the independent a/(a AND b) formula and separately checks
that corrected Go preserves all four rows. No corrected-source oracle is
silently substituted for this case.

A reordered ordinary interface changes output truth after conversion
(`/tmp/abcgo-front-interface-red.log`). Physical object IDs survive, but append
order need not equal logical port order. Copy logical CI/CO IDs, restore each
terminal's logical index and retain the owned interface through the existing
helper. Do not read Value as a copy literal: it holds a frontier slot here.
Word and ternary callers also incorrectly enumerated logical ports in physical
order; a separate regression fails before those caller repairs
(`/tmp/abcgo-front-reordered-consumers-red.log`). Switching already used the
terminal indices and needs no such change.

Mixed real XOR/MUX/buffer graphs expose lost MUX metadata, unreleased control
references, inflated cross-cuts and changed roundtrip/simulation truth
(`/tmp/abcgo-front-typed-red.log`). Preserve owned MUX controls as original
object literals, representation flags and typed counts. A control object's
Value gives its slot when the MUX consumes it. Reference counting and release
include this third logical edge.

For graphs with real XORs, retain the existing ordering-based XOR tag by
commutatively ordering compressed data slots, then restoring canonical object
pin order in the inverse transform. Ordinary graphs keep exact source slot
records. MUX data pin roles remain unchanged. This needs no new manager fields,
per-node gate-kind arrays, lowered object IDs or duplicate persistent state.
Switching counts continue to refer to the original nodes.

The shared word loop implements the same Boolean gate operations for ordinary
and switching simulation. Ternary evaluation uses the existing algebra on the
source's three-AND XOR/MUX expansions. In particular an unknown MUX control with
both data pins one retains the source expansion's conservative unknown result;
no stronger alternative ternary algorithm is substituted.

## Source reconciliation and independent coverage

| Source routines | Coverage |
| --- | --- |
| FrontFindNext | Every nonfull occupancy mask/start position for slot counts1..8, independent minimum circular-distance selection, unchanged unrelated slots; invalid dimensions/indices and full frontier reject rather than looping forever. |
| FrontTransform | All192 exact native cut/front/restored records over64 ordinary graphs, physical pin fields on64 mixed typed graphs, exhaustive before/after scalar truth and owned MUX metadata. |
| CrossCutSimple | Independent last-use interval maximum, all reference scratch zero after complete sweep; shared/dangling nodes, constants, buffers, all four repeated-fanin phases in simple mode, XORs and three-input MUX lifetimes. |
| Front | Exact ordinary native slots/fields; sequential dimensions, reordered ports, property/constraint roles, reset states, comments and independently owned interface names; constant-true control and empty/direct-output boundaries. |
| FrontTest | Exact successful diagnostic on empty and direct/constant output graphs, ordinary native roundtrip fields and typed physical pin checks. |

`test_vectors/frontier_review.c` includes actual original giaFront.c and links
original ABC objects through the shared source-oracle helper. These tests need
Linux/GCC and the built ABC objects and skip when unavailable. Default native
mode is unmodified source on64 raw ordinary DAGs, including constants, dangling
CIs/nodes, complemented edges and sequential dimensions. The original buffer
counterexample is a separate required native mode.

Independent cross-cut checking uses last-use intervals, rather than repeating
the production decrement loop. Scalar graph truth reads logical CI positions
and actual gate kinds; it uses no frontier slots or reference scratch.

Sixteen mixed graphs exercise all three direct callers at every four-input
Boolean assignment.96 multiword/count modes cover one/two/five words, both one
counts and transition counts, and two consecutive rounds. Every output and
original-node switching accumulator is checked against independent scalar truth.
The second round verifies cumulative counts and repeated slot reuse.

For unknowns, expand real gates without hashing or simplification, then
independently enumerate possible Boolean values at each ordinary AND. All1296
ternary input cases match, with and without exposing every internal typed node
as an output. This preserves the source expansion's precision instead of
comparing to an optimized or more powerful evaluator. MUX and interface tests
mutate destination storage to prove ownership.

## Compactness and measured loop cost

Share the two word simulators' node loops. The first shared implementation
selected gate kind per word and measured222.9/455.5/2361 ns per round at
one/eight/64 words, versus165.6/343.1/1536 for the previous ordinary loop
(`/tmp/abcgo-front-word-benchmark-before.log`,
`/tmp/abcgo-front-word-benchmark-after.log`). Move selection outside the word
loop and apply complement masks. The revised loop measures190.3/304.5/1209 ns
(`/tmp/abcgo-front-word-benchmark-hoisted.log`). All versions allocate zero bytes
and zero objects per round in this benchmark. These are single-run measurements
on the same24-node graph, not a general performance guarantee: the one-word
case retains some typed-dispatch overhead; the longer word loops improve.

Keep readable fixed-size fanin scratch in cross-cut counting and the source
slot allocation strategy. Do not add caches or a second graph representation
solely to remove a short dispatch branch. Metadata copies own their storage.
Normal graph traversal requires inverse transformation first; compressed
frontier fanin differences are slots, not ordinary backward object differences.
The inverse retains nFront, as C does. Full/invalid allocator rejection is a
bounded Go guard, not a replacement allocation strategy.

## Validation

Final focused caller/frontier rerun, frozen full module race tests and GIA vet
are pending. This family review does not complete the remaining GIA inventory
or the sequential audit of the original131 packages.
