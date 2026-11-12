# Original bridge fixture

From repository root, with original CUDD object files available:

```sh
cc -DABC_USE_STDINT_H -I src abcgo/bdd/cudd/core/test_vectors/bridge_oracle.c src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/cudd-bridge-oracle
/tmp/cudd-bridge-oracle
```

Expected checksum: `2c12e60c54a5bb25`.

The program invokes original CUDD public bridge APIs, not Go port code. Eight
numeric leaves cover negative/fractional numbers, zero, positive numbers, and
signed 32-bit boundaries. It hashes nine inclusive/strict threshold pairs, six
intervals (including NaN bounds), and all 32 integer bits. C's undefined integer
conversions and invalid shift counts are excluded; Go tests check those reject
with nil. Normal Go tests store the checksum and require no C compiler.
