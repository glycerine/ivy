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
