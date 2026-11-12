Original-source oracle, run at repository root:

```
cc -w -O2 -DABC_USE_STDINT_H -DUSE_SYSTEM_MEMORY_MANAGEMENT \
  -ffunction-sections -fdata-sections -Isrc \
  abcgo/opt/fxu/test_vectors/oracle.c src/opt/fxu/*.c \
  src/base/abc/abcSop.c src/misc/mem/mem.c src/misc/util/utilSort.c \
  src/misc/util/utilBridge.c -Wl,--gc-sections -lm -o /tmp/fxu-oracle
/tmp/fxu-oracle 0 > abcgo/opt/fxu/test_vectors/single.txt
/tmp/fxu-oracle 1 > abcgo/opt/fxu/test_vectors/double.txt
/tmp/fxu-oracle 2 > abcgo/opt/fxu/test_vectors/complement.txt
```

The harness supplies input vectors and required output storage. All extraction,
allocation, SOP and sorting implementation comes from unchanged ABC sources.
