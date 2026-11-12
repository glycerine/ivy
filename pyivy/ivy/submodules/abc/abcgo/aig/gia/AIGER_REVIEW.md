# AIGER codec review

Scope: `giaAiger.go`, `giaAigerExt.go`, and their tests. This is a completed-file review; the complete GIA package still contains unrelated draft translations.

## Compactness and readability pass

The parser, graph serialization, and ABC extension encoding use typed slices and checked byte cursors rather than pointer arithmetic. Existing variable-length and big-endian primitives are reused. Wire renumbering and metadata remapping are centralized; writers do not overwrite manager object scratch values. Go errors replace source assertions on malformed external data. Native `ParseAiger`/`MarshalAiger` APIs expose options; source-named adapters retain ABC-specific behavior. Timing/hierarchy records are preserved as explicit serialized `AigerTimingData`, with a marshaling interface for native timing managers.

The API deliberately returns errors for file/codec operations and owned slices for encoded records. No C allocator or unsafe dependency remains. Known extension tags decode to native manager fields; unknown tags retain their payload. Renumbering unknown object-indexed metadata is refused instead of corrupting it.

## Separate semantic and bug pass

Reviewed header bounds, sparse ASCII IDs, compact deltas, complementary literals, sequential ordering, standard latch initial values, constraints, symbols, metadata renumbering, and malformed extension lengths. Fixed property-count signed overflow, placement/switching retention, stale native-field extension serialization, and equivalence proof flags when a new representative is selected. Legacy reader wrappers now implement ABC's reset-one inversion and reset-X initialization-control circuit, and name-based `constraint:`/`invariant:` interpretation. Native parsing preserves standard AIGER reset metadata and output names without these legacy rewrites. As in the C duplication helpers, those legacy graph rewrites discard auxiliary indexed metadata. Partial writers preserve selected latch reset metadata and reject nil selections.

Unsupported source features (justice/fairness headers and specialized XOR/MUX wire nodes) return explicit errors. Compact `aig2` rejects nonzero or unknown reset values, which it cannot represent. There are no known pending correctness findings in this scope.

## Validation

Focused real-core tests pass with the race detector. Tests include standard/compact/ASCII sequential round trips, sparse inputs, all ABC extension records, interleaved-ID remapping, exact original-C equivalence/mapping codec differentials, and bidirectional interoperability with the repository's compiled ABC binary. Regression tests cover property overflow, reset-one/reset-X normalization, named constraints/invariants, metadata proof weakening, truncated records, and 3,000 deterministic malformed byte inputs. Full package compilation remains blocked by unrelated draft files.
