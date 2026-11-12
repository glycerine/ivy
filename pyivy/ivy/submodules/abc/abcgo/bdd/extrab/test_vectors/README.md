# Original-C extrab fixtures

These generators execute genuine original ABC algorithms. Go tests contain the
expected checksums and do not require a native compiler.

## Unateness

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddUnate.c -o /tmp/extrab-unate.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddSymm.c -o /tmp/extrab-symm.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddMisc.c -o /tmp/extrab-misc.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/unate_oracle.c /tmp/extrab-unate.o /tmp/extrab-symm.o /tmp/extrab-misc.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/extrab-unate-oracle
/tmp/extrab-unate-oracle
```

Expected: `94e5759806dcde0c` (all 256 three-variable functions; cofactor checks,
fast info fields, and slow info count).

## Bounded threshold search

```sh
cc -DABC_USE_STDINT_H -DABC_USE_CUDD -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddThresh.c -o /tmp/extrab-threshold.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/threshold_oracle.c /tmp/extrab-threshold.o -lm -o /tmp/extrab-threshold-oracle
/tmp/extrab-threshold-oracle
```

Expected: `23d92fb53fb7aab7` (all three-/four-variable functions; Chow values,
returned threshold, mutated truth table, and weights).

## Inequality heuristic

```sh
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/threshold_heuristic_oracle.c /tmp/extrab-threshold.o -lm -o /tmp/extrab-threshold-heuristic-oracle
/tmp/extrab-threshold-heuristic-oracle
```

Expected: `14b0eda9522e02d5` (all nonconstant three-/four-variable functions).
The adapter executes the original Chow sort/grouping and inequality weight search.
For its monotone functions, complete prime covers equal minimal true/maximal false
assignments; the adapter enumerates those in original SOP traversal order. It
avoids ABC's global frame/allocator interface and the C heuristic's buffer-overrun
bug, without replacing the original weight-search algorithm or truncating covers.

Function sections discard unrelated original CLI dependencies. There are no
native stubs. Floating infinity configuration does not affect these integer and
Boolean fixtures.
