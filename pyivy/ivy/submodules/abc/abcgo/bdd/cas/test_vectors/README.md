# Original-C fixture

`encoding_oracle.c` calls `GetSingleOutputFunction` from the original
`src/bdd/cas/casCore.c`, with original CUDD and extra helpers. It evaluates 4,096
three-output instances over all 32 assignments and prints FNV-1a checksum
`3702fd9a65c8d325`. The Go fixture stores only that expected checksum.

To regenerate from the repository root (existing original CUDD/MTR/EPD/ST object
files are required):

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/cas/casCore.c -o /tmp/cas-core.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddCas.c -o /tmp/cas-extra.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddMisc.c -o /tmp/cas-misc.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/misc/extra/extraUtilMisc.c -o /tmp/cas-util.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/cas/test_vectors/encoding_oracle.c /tmp/cas-core.o /tmp/cas-extra.o /tmp/cas-misc.o /tmp/cas-util.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/cas-source-oracle
/tmp/cas-source-oracle
```

Function sections discard unrelated executable/CLI dependencies without replacing
any executed algorithms. Go tests do not compile or run this program.
