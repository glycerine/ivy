# MUX mapping parity oracle

The genlib and BLIF inputs in this directory are used by
`amapMap_native_test.go`. They encode the same three-input mux and include
AND2, INV, XOR2, and MUX2 cells with unit area and unit pin delays.

Native C ABC command:

```sh
abc -c 'read_blif test_vectors/amap_mux.blif; strash; read_genlib test_vectors/amap_mux.genlib; amap -m -v; write_blif /tmp/amap_mux_c_output.blif'
```

The C output is `amap_mux_c_output.blif`: ABC selects one MUX2 cell. The Go
test enables `fUseMuxes` (the equivalent of `amap -m`) and asserts that the
serialized mapped network contains MUX2.
