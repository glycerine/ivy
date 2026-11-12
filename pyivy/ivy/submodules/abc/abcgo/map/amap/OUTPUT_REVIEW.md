# Amap mapped-network output review

`amapOutput.go` converts the mapper's selected matches into an ordered list of
input, gate, and output records. The source's trailing memory-manager entry is
preserved for the existing ABC adapter contract, while record and fanin storage
are Go-owned slices.

## Bug and readability pass

- Every serialized gate allocates its own fanin slice and copies its gate name;
  records do not alias another record's input list.
- Matched fanins use the C macro's combined complement rule:
  `set input complement XOR cut fanin complement`. The expected mapped
  polarity is compared to the fanin's selected polarity before choosing its
  direct or inverter record index.
- PI negative-polarity uses and opposite-polarity node uses require a real
  inverter gate and fail with an assertion if the mapping asks for one but the
  library lacks it.
- Constant-driven outputs select the library's constant-zero/constant-one gate
  based on the PO edge complement. Ordinary outputs select the direct or
  opposite-polarity driver record.
- The empty memory marker is still appended because the current ABC adapter
  removes and stops that final entry. The structs and slices themselves are
  ordinary Go allocations and remain valid if the marker is stopped.

The input/PO serializer test exercises a small manager. Its explicit-file-set
race and vet gates are documented in the parent work log; they are not a
substitute for `go test ./map/amap` or the external ABC adapter build.
