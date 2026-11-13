# RWR C fixtures

`manager_results.txt` records the forest size and every class member's ID and
truth table from `manager_oracle.c`, which invokes the original C
`Rwr_ManStart(0)` after `Abc_Start`. From the ABC repository root after
building its C object files, regenerate it with:

```sh
gcc -c -DABC_USE_STDINT_H=1 -I src abcgo/opt/rwr/test_vectors/manager_oracle.c -o /tmp/rwr_oracle.o
gcc -c -DABC_USE_STDINT_H=1 -I src src/misc/st/st.c -o /tmp/rwr_st.o
g++ -o /tmp/rwr_oracle /tmp/rwr_oracle.o /tmp/rwr_st.o $(find src -name '*.o' ! -path '*/base/main/main.o' ! -path '*/misc/st/st.o') -lm -ldl -lpthread -lz -lreadline
/tmp/rwr_oracle > abcgo/opt/rwr/test_vectors/manager_results.txt
```

The Go test compares all class lengths, member order and truth tables, as well
as the full forest size. It does not establish equivalence of the ABC network
rewrite adapter or the source precomputation mode.

`library_oracle.c` includes the original `rwrLib.c` to reach its static
candidate evaluator. It records complemented-AND insertion, dominance rejection,
equal-cost ordering, shared XOR volume and source manager counters. Regenerate
it with the commands above using `library_oracle.c` and additionally excluding
`*/opt/rwr/rwrLib.o` from the object list; write the output to
`library_results.txt`.
