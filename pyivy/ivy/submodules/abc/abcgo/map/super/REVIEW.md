# Native supergate precomputation

The complete `map/super` package now builds normally. Technology enumeration, AND2/inverter enumeration, Pareto pruning, legacy/tree serialization, source-shaped library entry points, and the registered `super`/`super2` commands are native Go. The original GENERIC source was an empty template; the writer source's format documentation is implemented by the native serializer. No executable C drafts are hidden, excluded or replaced by stubs.

## Readability and compactness pass

One recursive ordered-combination enumerator replaces the six nearly identical C nested-loop cases. It retains distinct fanin objects, area pruning, source root/fanin complexity heuristics after level one, input-inverter selection and final delay checks. The optional `Exhaustive` setting disables only the source's heuristic complexity pruning. `Mio_CollectRoots`, gate truth evaluation and pin-delay composition use the actual existing MIO library, with no duplicate parser or expression engine.

Truth classes use full uint64 keys. Each class owns a small area/per-input-delay Pareto frontier; frozen generation frontiers and immutable child references replace intrusive hash lists and recycled C objects. Slice lengths replace allocated/used-capacity fields. Standard Go ownership removes both manual memory managers. Search counters are int64; explicit options control depth, area, delay, class count, deadline and context. `MaxGates` follows the original class-oriented stopping rule, now counted by actual truth classes (including elementary variables) instead of colliding hash buckets.

AND2 enumeration uses explicit `{Gate, Inverted}` edges instead of tagged addresses. Its complement-class table is local to the manager, and formatting uses dynamic owned strings instead of static per-level buffers. A saturated function space is reused without repeating an exhaustive pair search. Zero-input AND enumeration produces the constant class safely.

Serialization shares one header implementation and local postorder numbering. Selected physical gates require consecutive used inputs, as in C. Referenced subgates that are no longer Pareto roots still receive unmarked tree rows. Native `MarshalTree` emits ordinary text; the source-compatible `Vec_Str_t` interface retains its terminating NUL. Output paths retain the input library directory, matching the original command. Native headers omit volatile timestamps, making repeated output reproducible.

## Separate source/bug pass

Reviewed the complete C sources and every native branch after the readability pass. Fixed and regressed the following issues:

- Elementary truth tables are finalized before table insertion. C inserted variables under zero keys and changed their truth values afterward.
- Truth keys include both words directly, eliminating XOR-hash aliases and the associated linked-list predecessor hazards. Functions below six inputs are masked to their actual minterm width.
- A candidate must survive all frontier comparisons before any dominated points are removed. Rejected candidates cannot partially edit the frontier.
- Removed gates remain valid when referenced by existing parents or the frozen generation array. C recycled these objects during enumeration, permitting retained pointer corruption. Unused slice-tail references are cleared.
- The final composed delay is checked explicitly. C admitted the first child beyond its candidate delay cutoff and did not enforce the final bound.
- The input-inverter restriction checks elementary variables, not zero maximum delay; a physical zero-delay gate can still receive an output inverter.
- Areas/delays and option bounds reject invalid/nonfinite data; the global class limit stops later levels rather than being repeatedly exceeded.
- Empty output libraries and unnamed manually created managers serialize valid headers instead of indexing the last member of an empty array or shifting metadata lines.
- Serialization numbers are local and unbounded Go integers, repairing repeated-write omissions and the source's 24-bit field limit. Cycles and invalid tree arities are reported.
- Equal AND/formula comparands compare equal; C's strict comparator contract was violated on identical values. Formula buffers have no five-level or fixed-character limit.
- The advertised `-o` command switch now writes the old-format file; C parsed it without passing it into generation. Command validation reports malformed values instead of accepting `atoi` truncation.

Generation maintains tree-area accounting even when immutable child structures are shared. Pin delay dominance retains C's missing-variable sentinel semantics. Root/fanin heuristic pruning is intentionally preserved by default; therefore this is not advertised as unconstrained exhaustive technology enumeration.

## Validation

Normal full-package gates pass:

- `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -cover ./map/super` (76.5% coverage).
- `GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/super`.

Tests use `test_vectors/tiny.genlib`. They independently evaluate composed Boolean truth, area and pin delays; preserve incomparable Pareto points; reject dominated points; check retained children; enforce limits/cancellation; cover six-input tables and zero-delay gates; validate empty/cyclic trees; and check repeated/concurrent serialization. Command tests exercise both actual registered commands, old-format output, source-compatible string termination and lifecycle cleanup. All output artifacts use temporary directories.

Actual original ABC executable tests compare complete selected truth/area/all-pin-delay frontiers for two-input one/two-level and three-input one-level technology libraries. Separate original `super2` runs match the full complement-class truth sets at two-input one/two-level and three-input two-level. The independent tree parser reconstructs every referenced row and validates numbering and selected-root counts. These oracles executed and passed in this workspace; they skip only when the original executable is unavailable. No tests touch experiments.
