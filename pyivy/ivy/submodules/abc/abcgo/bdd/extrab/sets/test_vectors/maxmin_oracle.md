# Original MaxMin fixture

From repository root (existing original CUDD objects required):

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddSet.c -o /tmp/extrab-set.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddMaxMin.c -o /tmp/extrab-maxmin.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/sets/test_vectors/maxmin_oracle.c /tmp/extrab-set.o /tmp/extrab-maxmin.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/extrab-maxmin-oracle
/tmp/extrab-maxmin-oracle
```

Expected checksum: `d5ec309d9c0db5a5`.

The generator covers every pair of the 256 three-element set families. Ordinary
pairwise union/intersection products receive general families. Original fused
maximal/minimal union and maximal dot product receive antichains, the prerequisite
the original public wrappers fail to enforce. Go's public wrappers normalize
arbitrary families before invoking the same recursions. Separate independent
finite-set tests cover all general inputs and the source shortcut defect.

The fixture also includes original minimal-family extraction. Native generators
are never run by normal Go tests; there are no executed stubs or alternative
native algorithms.
