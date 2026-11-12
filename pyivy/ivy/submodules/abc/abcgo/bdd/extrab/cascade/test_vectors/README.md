# Non-strict encoding oracle

`encoding_oracle.c` links the original ABC `extraBddCas.c` implementation and
checks every seventeenth three-input partition of four columns (3,856 instances).
It hashes the original direct-wire optimization count and the number of allowed
codes for each input assignment: `6b850130b3287765`.

Concrete code numbers are pointer-hash traversal dependent, so the Go fixture
checks invariant multiplicities and optimization counts. It also independently
checks totality, code uniqueness, and column separation.

Compile `extraBddCas.c`, `extraBddMisc.c`, and `extraUtilMisc.c` into the three
function-section objects as described in `bdd/cas/test_vectors/README.md`, then:

```sh
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/cascade/test_vectors/encoding_oracle.c /tmp/cas-extra.o /tmp/cas-misc.o /tmp/cas-util.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/cas-encoding-oracle
/tmp/cas-encoding-oracle
```

The original process-global `HHTable2` caches pointer addresses after a manager
has been destroyed. The oracle clears that cache before each fresh manager to
prevent recycled addresses from contaminating independent cases. This does not
change the encoding algorithm; a fresh process per case has the same effect.
The Go implementation uses call-local maps.
