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
