# Sequential duplication review

`saigDup.go` ports the active duplication and counterexample algorithms from
`src/aig/saig/saigDup.c`. A source-ID-to-destination-edge map replaces `pData`
casts. AND, native XOR, buffers, complemented edges, and forward-ID dependencies
are handled recursively. Register interfaces retain trailing ordering.

## Readability and compactness pass

One graph copier implements the repeated source duplication loops. OR-output,
phase-coordinate, equivalence-miter, abstraction, trim, and cone construction
retain their own explicit interface policies. Source-private recursive helpers
are consolidated into this copier and cone DFS. One replay engine implements
all counterexample consumers; its frame visitor expands PI-plus-state traces in
one linear pass, rather than repeatedly replaying prefixes. Go owns storage.
The source's disabled-CUDD zero-result/no-op placeholders are not added.

## Independent bug pass

Checked the original loops and frame boundaries separately from refactoring.
Trim retains the source all-PIs-unused special case and does not clean dangling
logic. Abstraction orders abstracted states before retained states, closes
selected next-state cones, removes dead registers, and remaps `InputNumbers`
provenance after cleanup. Cone selection follows registers transitively in first
DFS-discovery order. Phase duplication flips both state inputs and next-state
outputs. Equivalence outputs include the source phase correction.

Replay initializes stored register values, consumes frame-major PI bits, and
returns the state before the final transition or the next state as requested.
NoClear writes every live node's final-frame MarkB; other entry points clean
marks. Dimensions and packed-bit capacity are validated. Cex expansion records
PI and current-state values before each frame's transition. Tests cover those
boundaries and 200 independent seven-frame simulations with AND/XOR graphs.

`Graph.InputNumbers` adds the source CI-provenance metadata. The native cone API
accepts `[]int32` plus the source count instead of an unbounded C pointer.
Abstraction accepts repeated calls without mutating source traversal marks.

Focused commands (package-wide compilation still encounters unrelated SAIG drafts):

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test aig/saig/saig.go aig/saig/saigDup.go aig/saig/saigDup_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race aig/saig/saig.go aig/saig/saigDup.go aig/saig/saigDup_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet aig/saig/saig.go aig/saig/saigDup.go aig/saig/saigDup_test.go
```
