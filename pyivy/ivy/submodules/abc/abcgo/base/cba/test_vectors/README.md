# CBA C-source probes

These probes call the original ABC implementation, and the corresponding Go
unit tests check the same results. They do not imply that the CBA package is
fully translated or builds yet.

- `cba_constants.c`: binary, x/z, hex, decimal, signed-fill, and omitted-width
  constant decoding (`cbaConstants_test.go`).
- `cba_names.c`: natural ordering of ASCII object names (`cbaNames_test.go`).
- `cba_fons.c`: constant-fon encoding (`cbaFons_test.go`).
- `cba_verilog_names.c`: Verilog identifier checks (`cbaVerilogNames_test.go`).
- `cba_verilog_chars.c` and `cba_blif_chars.c`: character classifications for
  all 256 byte values. The two source parsers intentionally classify newline
  differently.

After building ABC's C objects, a probe can be compiled from `abcgo` with:

```sh
gcc -c -DABC_USE_STDINT_H=1 -I../src base/cba/test_vectors/PROBE.c -o /tmp/cba_probe.o
gcc -c -DABC_USE_STDINT_H=1 -I../src ../src/misc/st/st.c -o /tmp/cba_st.o
g++ -o /tmp/cba_probe /tmp/cba_probe.o /tmp/cba_st.o $(find ../src -name '*.o' ! -path '*/base/main/main.o' ! -path '*/misc/st/st.o') -lm -ldl -lpthread -lz -lreadline
/tmp/cba_probe
```

The probes which include an entire source file to access static functions
require excluding that file's object from the link: exclude `cbaWriteVer.o`
for `cba_verilog_names.c`, `cbaReadVer.o` for `cba_verilog_chars.c`, and
`cbaReadBlif.o` for `cba_blif_chars.c`.
