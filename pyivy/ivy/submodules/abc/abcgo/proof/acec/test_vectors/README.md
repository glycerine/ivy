# ABC source fixtures

`acec_source_results.txt` is the output from `acec_source_harness.c` using
ABC's compiled C routines. From the `abcgo` directory, after building ABC and
its object files, regenerate the source output with:

```sh
gcc -c -DABC_USE_STDINT_H=1 -I../src proof/acec/test_vectors/acec_source_harness.c -o /tmp/acec_harness.o
gcc -c -DABC_USE_STDINT_H=1 -I../src ../src/misc/st/st.c -o /tmp/acec_st.o
g++ -o /tmp/acec_harness /tmp/acec_harness.o /tmp/acec_st.o $(find ../src -name '*.o' ! -path '*/base/main/main.o' ! -path '*/misc/st/st.o') -lm -ldl -lpthread -lz -lreadline
/tmp/acec_harness
```

The mask is the set of marked object IDs. The normalized result records its
input/output counts; the Go test additionally exhaustively checks that its own
seven-input function is preserved. The harness does not compare C and Go
normalized truth tables, so it does not establish their functional equality.
The tied-cut case records the C `qsort` and matching
result for three XOR cuts and two majority cuts with the same leaf tuple.
Because C does not guarantee stable ordering for equal sort keys, the recorded
ordering is for the library/platform used to build ABC.

`acec_single_output.c` uses the same compile and link commands with its own
filename. It records `box=1`: source box detection can succeed on a graph with
one output. This does not establish that the later `Acec_Solve` path is safe
for that graph; the source unconditionally patches its last two outputs after
matching boxes.

`acec_ambiguous_polarity.c` builds a reconvergent XOR and invokes
`Acec_DetectAdditional`. With ABC assertions enabled it aborts on inconsistent
polarity, matching the precondition enforced by the Go port. Build it with the
same commands, substituting its filename.

`acec_one_root.c` and `acec_single_adder.c` both reach the same assertion
because their roots do not form a linked chain. `acec_disconnected_roots.c`
shows that C retains the first linked chain and omits a disconnected root.

`acec_zero_roots.c` invokes `Acec_ProduceBox` on a single-input buffer graph.
With assertions enabled, ABC aborts in `Acec_OrderTreeRoots` because no XOR
root exists; the Go test checks the corresponding panic. Build it with the
same commands, substituting its filename.

`acec_polynomial_order.c` builds a two-full-adder chain and prints the
complete `Gia_PolynReorder` result. The Go regression checks those source
values, including entries for the output objects. Use the same C build commands
with this filename to regenerate it.

`acec_npn_order.c` records the canonical truth and ordered orbit for a
three-input truth table. Run it with no arguments and with one argument to
exercise both an empty and a pre-populated destination vector.

`acec_missing_box.c` invokes normalization on a graph with no arithmetic box;
the C implementation dereferences the missing box. `acec_large_coefficients.c`
checks signed exponent carries and cancellations without overflowing C `int`.
It includes `acecPolyn.c` to reach a static helper, so use the same build
commands but additionally exclude `*/proof/acec/acecPolyn.o` at link time.

`abc_acec_results.txt` records the ABC CLI differential cases; set
`go test -run TestABCDifferential ./proof/acec` always compares Go against the
recorded results. Set `ABC_BINARY` to the absolute path of the compiled ABC
executable to also rerun ABC against the generated AIGs.

`acec_booth_oracle.c` generates `acec_booth_results.txt` using the same
compile/link commands above, substituting its filename. It compares two Booth
recognizers on 96 ordered constructions: 32 input-polarity combinations for
each of a valid Booth-one shape, a different-control negative case, and a
valid Booth-two shape. The fixture records both return values and all output
slots, including their values on failure. Intermediate AND calls are sequenced
explicitly to avoid C's unspecified argument evaluation order changing node
IDs. The Go regression builds the same GIA operations and compares every row.

`acec_multiplier3_oracle.c` compares deterministic 3×3 multiplier graphs
through C `Acec_Solve`. Invoke it with `-1` (equivalent), `0` (invert output
0), or `4` (invert output 4). Freshly compiled `src/proof/acec/*.c` routines,
linked with the remaining ABC object files, return statuses `1`, `0`, and `1`
respectively. The last result is the documented source false equivalence: the
matched path discards the final two outputs. The Go test exhaustively checks
the fixture's multiplier truth table and compares those statuses; set
`ACEC_MULTIPLIER3_ORACLE` to a compiled oracle path to rerun C during the test.
The non-ACEC object files used for this comparison must also match the checkout.
