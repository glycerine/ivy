# Original complete-primes oracle

From repository root, using existing original CUDD objects:

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddMisc.c -o /tmp/extrab-misc.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/primes/test_vectors/primes_oracle.c /tmp/extrab-misc.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/extrab-primes-oracle
/tmp/extrab-primes-oracle
```

Expected checksum: `5394c07cdac40638`.

For every three-variable function, the program invokes original
`Extra_zddPrimes` and hashes membership of every consistent ternary literal cube.
The Go test also independently enumerates mathematical prime implicants.
No native compiler/library is required at Go test time.
