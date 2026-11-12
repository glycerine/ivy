# LUT cut balancing

`giaBalLut.go` ports the active algorithms of `src/aig/gia/giaBalLut.c`.
The public `Gia_ManBalanceLut` still coarsens AND-encoded XOR/MUX structures,
balances using bounded cut sets, and expands real XOR/MUX nodes back into an AIG.
It preserves output and register ordering. LUT sizes are bounded by six leaves
and cut capacities by eight, with at most capacity-minus-one retained cuts.
A three-input MUX requires a LUT large enough for its fanins, as in the source;
size-two LUTs are useful for binary-gate networks.

Cuts retain sorted leaves, 64-bit signatures, delay, and source function IDs.
Their ordered unions, containment filtering, delay/leaf-count cost, and insertion
tie ordering match C. Unit cuts raise the level; existing cuts permit grouping at
the current LUT level. Pair selection checks higher-level packing, then equal
lowest-level packing, and finally proximity, in the source order. AND/XOR
supergates retain the source fanout-three boundary and soft 50-leaf collection
limit. Constant/duplicate simplification and structural hashing remain active.
The source's commented-out truth-minimization calls remain inactive here too.

## Readability and compactness pass

Completed independently of the correctness pass. Native slices replace fixed
sentinel-filled cut arrays, byte layout assumptions, pointer-vector permutations,
and packed pair indices. One sorted-union routine handles both two- and
three-input merges. One dominance/insertion helper preserves tie behavior.
Manager state is explicit; source process-global temporary unit cuts are local
values, and graph `pData` is not borrowed. Source-local Gia_* helpers are renamed
`giaLut*` so AIG balancing retains its separate algorithms. Source Bal_* helper
interfaces use native slices/bools and pair results rather than pointer outputs.

## Independent bug pass

Compared every live source stage, including target-delay equality, retained cut
capacity, equal-cost insertion, reconstruction ordering, XOR parity, MUX control
ordering, existing hashed nodes, and final cleanup. Tests caught the native
uncoarsening helper dropping constraint metadata: the final LUT entry point now
restores constraints and IO names explicitly. This also repairs the original LUT
pipeline's loss of constraint counts. Registers and identity are retained.

Validation includes 100 random AND/XOR/MUX networks on all 32 input assignments,
all pairs of nonempty six-variable supports, cut containment/tie/dominance tests,
a 12-input chain over all 4096 assignments, constants, binary-size LUTs, invalid
capacity checks, and sequential constraint metadata. The original ABC executable
runs `&blut -n` on eight six-input networks with LUT sizes 3/4/6 and cut capacities
2/8. All 48 cases match output truth tables, AIG node counts, and AIG depth.
This oracle actually ran during review; absence of the executable is an explicit
skip. No test relies on experimental outputs.

Focused normal/race tests and vet pass using the native GIA bridge source list.
For example:

```
abcgo/aig/gia/test_aig_bridge.sh -race -run TestLut aig/gia/giaBalLut.go aig/gia/giaBalLut_test.go aig/gia/giaBalLut_reference_test.go
```

The ordinary package gate was also run:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./aig/gia
```

It still fails at unrelated `giaBalMap.go` syntax errors. Focused verification is
not a claim that the whole GIA package builds. All remaining drafts stay visible;
no build constraints or placeholder implementations were introduced.
