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

`acec_zero_roots.c` invokes `Acec_ProduceBox` on a single-input buffer graph.
With assertions enabled, ABC aborts in `Acec_OrderTreeRoots` because no XOR
root exists; the Go test checks the corresponding panic. Build it with the
same commands, substituting its filename.

`acec_polynomial_order.c` builds a two-full-adder chain and prints the
complete `Gia_PolynReorder` result. The Go regression checks those source
values, including entries for the output objects. Use the same C build commands
with this filename to regenerate it.

`abc_acec_results.txt` records the ABC CLI differential cases; set
`go test -run TestABCDifferential ./proof/acec` always compares Go against the
recorded results. Set `ABC_BINARY` to the absolute path of the compiled ABC
executable to also rerun ABC against the generated AIGs.
