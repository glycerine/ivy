# Independent CUDD ZDD oracle

`zdd_setops_oracle.c` runs against the original ABC CUDD implementation, not the
Go implementation. It constructs every family of subsets of three variables,
then visits all 65,536 ordered pairs. It folds union, intersection, difference,
one ITE triple per pair (`h = families[(a+3*b)&255]`), and inclusion into an FNV
checksum. The family encoding is bit `set` for each member set; graph pointers
and allocation order do not affect the output.

The original C implementation produced `7292090e7f59522c`, matched by Go. This
value is checked by `TestZDDOriginalCUDDOracleChecksum`. Normal Go tests do not
compile or execute this C file and require no C toolchain/runtime.

From the ABC checkout, with its original C object files built, reproduce with:

```sh
cc -DABC_USE_STDINT_H -I src \
  abcgo/bdd/cudd/core/testdata/zdd_setops_oracle.c \
  src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o \
  src/misc/st/st.o src/misc/extra/extraUtilUtil.o \
  -lm -o /tmp/abcgo_cudd_oracle
/tmp/abcgo_cudd_oracle
```

The harness links genuine original functions, including supporting utility code;
it supplies no stubbed CUDD or allocator functions.

## Cover product/division oracles

`zdd_cover_oracle.c` covers all 512 families of the nine consistent cubes on two
Boolean variables (four ZDD literals), paired with each of 511 nonempty divisors.
It checks unate product, paired-literal product, divide, divideF, weakDiv and
weakDivF. Every temporary C result is referenced and released according to CUDD's
public contract. Its combined checksum is `1c45d75814740e31`.

`zdd_cover_random_oracle.c` adds 2,000 deterministic three-variable cover pairs.
Inputs include native-mask products, single-literal divisors and dense random
covers; the native mask preparation is independent of the CUDD algorithms.
The LCG output is mixed with its upper bits so the two operand streams do not
have fixed opposite parity. The combined checksum is `5eb39e097b2af27d`.
This larger logical-result oracle disables CUDD's garbage collection through its
real public API; it introduces no stubs and leaves logical operations unchanged.

`zdd_cover_reordered_oracle.c` repeats those cases after original CUDD's real
ZDD heap shuffle to level order `[4,5,0,1,2,3]`, preserving positive/negative
literal adjacency. The checksum is again `5eb39e097b2af27d`. Go verifies both
orders using fresh managers; this fixture does not claim that Go ZDD dynamic
heap reordering has been ported.

Reproduce each with the earlier link command, substituting the desired C source
and executable name. `TestZDDCoverOriginalCUDDOracle` and
`TestZDDCoverThreeVariableOriginalOracle` check the fixed combined and per-operation
checksums entirely in Go. Divide and divideF have identical source recursion in
this CUDD version apart from their cache tags and share the Go engine. WeakDivF's
separate earlier-dividend-variable path is retained. Matching these fixtures does
not establish that the weak-division variants are universally interchangeable.

The three cover-harness operation orders are:
`unate product`, `product`, `divide`, `divideF`, `weakDiv`, `weakDivF`.
Checksums fold each numeric set-family bit mask as one unsigned 64-bit value,
then multiply by FNV's prime; the C and Go loops use that same specified order.
The complete two-variable fixture checks 1,569,792 operation outputs. The two
three-variable runs add 12,000 outputs each, including the independent original
C heap-order change in the reordered harness.

# Independent CUDD dynamic reordering oracle

`reorder_oracle.c` constructs every three-variable Boolean function, references
its root, then runs either real `Cudd_ReduceHeap(SIFT)` or `SYMM_SIFT`. The 512-row
`reorder_oracle.txt` fixture records initial/final DAG sizes, success status,
post-reorder truth table, resulting order, symmetric-variable mask and group
count. The Go test compares all semantic/size/group fields. Permutation ties are
not compared because C qsort tie ordering is unspecified.

This oracle caught the public return-value discrepancy: the source comment says
1+symmetry count, but the actual CUDD tree dispatcher returns 1 on public success.

Reproduce from the ABC checkout (using the same original object files above):

```sh
cc -DABC_USE_STDINT_H -I src \
  abcgo/bdd/cudd/core/testdata/reorder_oracle.c \
  src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o \
  src/misc/st/st.o src/misc/extra/extraUtilUtil.o \
  -lm -o /tmp/abcgo_reorder_oracle
/tmp/abcgo_reorder_oracle > abcgo/bdd/cudd/core/testdata/reorder_oracle.txt
```

## ISOP, cover conversion and logical complement oracle

`zdd_isop_oracle.c` checks all 6,561 three-variable lower/upper intervals. Each
assignment is independently OFF, don't-care, or ON, so lower is always contained
in upper. For each interval it hashes the chosen BDD truth table, exact ZDD cube
family, converted BDD truth table, BDD-only ISOP truth table, logical complement's
exact cube family, and the complement's converted truth table.

The combined original-C checksum is `23c2e19ddcb24343` for the identity BDD order
and `f1ff6f7e29aa5823` after the real CUDD BDD shuffle `[2,0,1]`. Changing only the
ZDD level order leaves those logical outputs unchanged, including both a pair-
group permutation and reversed adjacent literal pairs. The permanent Go tests
check six combinations, including independently different BDD/ZDD level orders.

Use the earlier link command with `zdd_isop_oracle.c`, adding these optional
compiler flags before compilation:

- `-DREORDER_BDD`: original CUDD BDD shuffle to `[2,0,1]`.
- `-DREORDER_ZDD`: original CUDD ZDD shuffle to `[4,5,0,1,2,3]`.
- `-DREORDER_ZDD -DREVERSE_LITERAL_PAIRS`: ZDD order `[1,0,3,2,5,4]`.

Both BDD and ZDD flags may be combined. The oracle disables CUDD GC through its
public API and follows the reference protocol for all retained operands/results.
The Go verification requires no C compiler, C object or runtime library. Source
files `cuddZddIsop.c` and `cuddZddFuncs.c` were also freshly compiled into temporary
objects and substituted for the cached originals: identity and combined BDD/
reversed-literal configurations reproduced the same fixed checksums.

Go's ISOP implementation is additionally checked against native truth bounds and
cube irredundancy, not only against a checksum. BDD-only mode is checked for
absence of ZDD allocation, and invalid intervals must leave output handles intact.
