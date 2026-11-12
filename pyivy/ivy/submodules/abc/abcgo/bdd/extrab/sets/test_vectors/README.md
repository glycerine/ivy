# Original set-family oracle

From repository root, with original CUDD objects available:

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddSet.c -o /tmp/extrab-set.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddMaxMin.c -o /tmp/extrab-maxmin.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/sets/test_vectors/set_oracle.c /tmp/extrab-set.o /tmp/extrab-maxmin.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/extrab-set-oracle
/tmp/extrab-set-oracle
```

Expected: `697c168ae20d22d5`.

The generator invokes the five original set filters for every pair of the 256
three-element set families, plus original empty-membership/exactly-one predicates
and maximal-family extraction. Each resulting family is hashed as its eight-bit
membership table. The Go fixture also compares every result against independent
ordinary set inclusion, under identity and nonidentity ZDD order.

Function sections discard unrelated original MaxMin functions; no executed
algorithm is stubbed or replaced. Native compilation is optional and never occurs
in normal Go tests.
