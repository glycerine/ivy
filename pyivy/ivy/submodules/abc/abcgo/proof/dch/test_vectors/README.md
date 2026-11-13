# Independent original DCH oracle

`dch_oracle.c` invokes original ABC `Dch_DeriveTotalAig`/`Dch_ComputeChoices` on
associative and redundant-support snapshots. Each fixture row is:

    case total_ands result_ands choice_heads output_truth_mask

All three rows match exact counts and every truth assignment in the Go test.
The harness links real original functions; no solver or allocation stubs.

Reproduce in an ABC checkout with the original C object files built:

```sh
cc -DABC_USE_STDINT_H -I src -c \
  abcgo/proof/dch/test_vectors/dch_oracle.c -o /tmp/abcgo-dch-oracle.o
find src -name '*.o' ! -path 'src/base/main/main.o' > /tmp/abcgo-dch-objects.rsp
g++ /tmp/abcgo-dch-oracle.o @/tmp/abcgo-dch-objects.rsp \
  -lm -ldl -lpthread -lreadline -o /tmp/abcgo-dch-oracle
/tmp/abcgo-dch-oracle > abcgo/proof/dch/test_vectors/dch_oracle.txt
```

Go tests only read the fixture; the resulting port has no C runtime dependency.

`dch_choice_order.c` builds three associative implementations of a four-input
AND and derives their choice chain. Its output is `11(4,10) 9(5,8) 7(1,6)`
both with no arguments and with one argument (enabling redundant-support
filtering). Reproduce it with the same compile/link steps, substituting its
filename, and run the resulting executable both ways. The Go regression checks
the full ordered chain, including the DFS-renumbered fanins.
