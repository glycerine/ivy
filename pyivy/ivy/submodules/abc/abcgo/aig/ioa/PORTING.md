# IOA AIGER I/O

IOA is independent of the AIG manager package and imports only Go's standard
library. `AigerProvider` exports a model snapshot; this avoids an import cycle
when an AIG manager itself needs to write a recording. AIG owns conversion
between its graph and this literal model.

`Read`/`Marshal` and `Write` return errors. Original named entry points are
retained with Go slices and error results. `Ioa_ReadAiger` returns `*Model`.
The compatibility `check` argument remains accepted; decoding always validates
references, definitions, cycles, resets and symbol indices.

Supported encodings are standard binary `aig`, ASCII `aag`, and ABC's delta-coded
`aig2`. Standard encodings preserve AIGER 1.9 latch initialization, bad-state
properties, constraints, justice and fairness sections, symbols, comments and
ABC model names. The compact variant returns an error for nonzero/uninitialized
reset or liveness metadata that its original format cannot encode.

`Model` literals use 0=false and 1=true. Latch `Init` is 0, 1 or its own literal
(uninitialized). `Comment` is the payload after the comment marker line,
excluding that line's newline. `Normalize` returns a dense topological copy
without mutating the source or dropping dangling gates. ASCII supports sparse
IDs and arbitrary acyclic gate ordering. The decoder limits header-directed
allocation to 2^24 objects and rejects overflowing/truncated delta integers.

Reference specifications:
- https://fmv.jku.at/aiger/FORMAT
- https://fmv.jku.at/papers/BiereHeljankoWieringa-FMV-TR-11-2.pdf

## Separate review passes

The compactness/readability pass kept counts derived from slices (only MaxVar
is stored because sparse ASCII headers can contain unused IDs), removed the
unused assertion helper, and eliminated redundant full validation in binary
serialization. Models are validated and normalized without altering managers'
scratch fields or constraint polarities.

The bug-review pass added malformed-input regressions, zero-entry compact lists,
integer/delta overflow checks, combinational cycle detection, reset and symbol
checks, sparse normalization tests and comment-payload roundtrip coverage. A
failing regression caught an unwanted newline being added to comment payloads;
serialization/parsing now preserve payload bytes, including embedded zero bytes.

Tests cover fixtures, full extended metadata, model/file/memory roundtrips,
short writes and malformed inputs. An optional original ABC binary check reads
and re-emits both standard and compact output, then compares all circuit input
assignments. A fuzz run passed more than 556,000 inputs before the review fixes;
additional fuzz/race runs verify the final version. Production code has no C
runtime or compiler dependency.
