# Native balancing, hashing, MUX and window review

This records completed source families, not full GIA package completion.
The ordinary package gate currently reaches the visible giaCSat2.go draft.

## Compactness and readability pass

- giaBalAig.go ports all delay balancing and DAM area extraction functions.
  DAM lists use size-prefixed slices with integer offsets, retaining source
  ordering while avoiding invalidated pointers when stores grow. The indexed
  source-compatible Vec_Que heap still supplies divisor priorities and tie order.
- giaHash.go retains linked buckets and source canonicalization, source
  additional-strashing rules, and actual hit/miss accounting. Hash slots are
  reacquired after appending a gate. No second graph or generic SAT replacement
  was introduced. Kit's AndLiteral shares an active native table; its inactive
  convenience cache is released when native hashing starts/stops.
- giaMuxes.go ports profiling, MUX reference accounting/structure serialization,
  source restructuring and cofactor-class synthesis. String maps replace C name
  managers, keeping first-seen class order. Native Kit generates class-code logic.
- giaBalMap.go owns window marking/extraction/insertion and depth selection.
  balance_window_drivers.go retains the real SOP/DSD synthesis calls; those
  dependent GiaIf engines remain visibly unfinished. The drivers were not
  replaced by ordinary AIG balancing.
- normalize.go ports normalization and box-interface counters needed by these
  drivers, using actual misc/tim partitions.

## Separate bug-review pass and fixes

- Reacquire hash bucket slots after backing-vector growth and retain source
  unsigned arithmetic in bucket computation. Skip explicit buffers in hash
  insertion/resize (the source resize invariant excludes them).
- Source additional-strashing can compare complemented null C pointers. Go uses
  an explicit opposite-edge predicate, preventing a nil complement panic without
  treating two nil children as complementary.
- Retain float64 priority arithmetic followed by float32 storage, as in C.
- Copy DAM node lists before extending stores. A removed divisor occurrence must
  update weights only for the pairs actually still present in its supergate.
- Preserve interface names, register resets, constraints and bad-output partition
  through coarsening/uncoarsening and balancing. Node-indexed metadata requires
  its own remapping and is not blindly copied.
- The original three-MUX restructuring selector expression is valid for canonical
  regular then edges. The port keeps that expression and additionally preserves
  then-edge phase for directly appended noncanonical MUXs. It does not claim the
  original canonical algorithm is algebraically wrong.
- Window insertion now retains sequential register/constraint metadata. Standalone
  extraction retains selected CI/constant outputs that the source's else-if
  omitted. Temporary marks are cleared.
- Class truth generation uses Kit's actual repeated 32-bit small-table convention.
  MUX structure keys retain source zero-padded IDs and NUL termination.

## Evidence

- 48 actual original ABC `&b` cases (12 seeded random DAGs, ordinary/real-gate
  preprocessing, delay/DAM modes): identical AND counts and logic depth.
- 30 random networks, four balancing modes, all 32 input assignments; serial
  16-input AND depth falls from 15 to 4.
- Exhaustive additional-strash pair/phase truth; real-gate hash simplifications,
  phase canonicalization and reuse; hash growth/resize and Kit interoperability.
- 16 nested-MUX phase combinations times all 64 assignments; MFFC/dump reference
  restoration; cofactor compression replay and typed structural-profile codes.
- Window external shared-fanout roundtrip, zero-depth extraction, register and
  constraint retention, and absolute/percentage depth selection.
- The maintained real-native AIG suite passes race testing. Whole-package race
  and vet cannot yet run past the remaining visible parser drafts.
