# BLIF-MV review

## Readability / idiomatic pass

Replaced abcBlifMv.go's complete 13-API group: MV variable attributes and decimal
scanning, normalized numeric MV-cover parsing/strashing, binary naming and
logarithmic interface encoding, MV skeleton encoder/decoder construction,
processed combinational logic insertion, Boolean/SOP-to-MV conversion, packed
SOP conversion, and the original cube-count cost estimator. Attribute maps,
owned strings/slices and typed MVTable values replace flexible arenas, void
pointer casts and node-copy fields overloaded with arrays. Existing object
copy operations clone MV metadata. Tables implement CloneForNetwork.

mv_table.go exposes ParseMVTable, canonical formatting and relational evaluation
of defaults, finite sets, wildcard and input-equality constructs. mv_evaluate.go
provides multi-valued cone evaluation without collapsing nondeterminism. This
parses the normalized per-node cover grammar handled by abcBlifMv.c, not the
separate .model/.mv file reader owned by base/io. Skeleton insertion directly
inlines the processed module into its real encoding/decoding skeleton rather
than leaving a black box or substituting an approximate function.

## Separate bug / fidelity pass

The review found deleted graph objects retained in the MV attribute map. Added
a red deletion regression, removed attributes during Network.DeleteObject, and
verified the regression green. Temporary encoding arrays now clear on both
successful and failed strashing. Free-variable PIs move before register CIs,
with explicit register input/output pairing. Local equality comparisons iterate
the input domain (not the output domain as in the C loop), regression-tested
with ternary inputs and binary output. Fixed-default reconstruction excludes
that default value from coverage, matching C semantics.

Tests compare table parse/format roundtrips and evaluations, literal sets and
complements, input and output equality/defaults, malformed tables, binary AIG
encoding over all valid values, sequential free-variable ordering/connectivity,
skeleton/logic insertion over every ternary input, both SOP polarities,
attributes/free/deletion and packed-SOP cost behavior. Combined original/core
selected-file tests pass with -race.

The source's logarithmic encoding and constant policy are retained: binary
nondeterministic constants get fresh inputs; higher-domain wildcard constants
choose zero. General relations remain explicit in EvaluateMV. Converting an
opaque Go Boolean function to an exact cover enumerates assignments and returns
failure above 20 inputs; existing SOP covers convert directly without that
enumeration bound. No approximate large-function cover is emitted. The separate
BLIF-MV file parser and remaining base/abc algorithms are still unported.

IO integration adds SetMVVariable with deep-copy/validation and explicit symbolic
binary-domain retention. Regressions cover ownership, invalid update atomicity,
clearing annotations and deleted objects; the legacy binary-default heuristic
in Abc_NtkSetMvVarValues remains available independently.
