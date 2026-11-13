# WLN source fixture

`wln_fanins.c` calls the original WLN allocation, fanin insertion and access
routines. With assertions enabled it confirms the writable inline metadata
slot, the 0/4/8/16 capacity sequence, and fanin order across both growth
transitions. The Go test `TestFaninVectorSourceStorageTransitions` checks the
same cases.

To regenerate, build ABC's C object files, then from the `abcgo` directory:

```sh
gcc -c -DABC_USE_STDINT_H=1 -I../src base/wln/test_vectors/wln_fanins.c -o /tmp/wln_fanins.o
gcc -c -DABC_USE_STDINT_H=1 -I../src ../src/misc/st/st.c -o /tmp/wln_st.o
g++ -o /tmp/wln_fanins /tmp/wln_fanins.o /tmp/wln_st.o $(find ../src -name '*.o' ! -path '*/base/main/main.o' ! -path '*/misc/st/st.o') -lm -ldl -lpthread -lz -lreadline
/tmp/wln_fanins
```
