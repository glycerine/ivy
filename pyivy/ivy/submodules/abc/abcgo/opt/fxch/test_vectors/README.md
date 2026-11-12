Regenerate from repository root using unchanged source algorithms:

```
cc -w -O2 -DABC_USE_STDINT_H -ffunction-sections -fdata-sections -Isrc \
 abcgo/opt/fxch/test_vectors/oracle.c src/opt/fxch/*.c \
 src/aig/gia/giaUtil.c src/misc/util/utilSort.c src/misc/util/utilBridge.c \
 -Wl,--gc-sections -lm -o /tmp/fxch-oracle
/tmp/fxch-oracle 0 > abcgo/opt/fxch/test_vectors/single.txt
/tmp/fxch-oracle 1 > abcgo/opt/fxch/test_vectors/double.txt
/tmp/fxch-oracle 2 > abcgo/opt/fxch/test_vectors/mux.txt
```

The harness supplies only input construction, result printing and ABC application's
logging hooks (bridge-mode disabled). Extraction, MUX normalization, weights,
subcube hashing, random keys and vector code come from original source.
