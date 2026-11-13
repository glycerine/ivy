# CNF review

The correctness review covered graph encoding, output constraints and phases,
cut composition and mapping, clause metadata, frame loading, variable lifting,
packed covers, and DIMACS I/O. Fixes and regressions are recorded in the commit
history. Seeded graph and cut oracles compare small cases against independent
Boolean evaluation; frame tests exercise initial state and a latch transition.

The readability pass consolidated serialization and leaf traversal and uses
the shared ISOP implementation for cut costs. The encoder uses a direct
Tseitin representation; its internal cut mapping and clause optimization are
not a claim of byte-for-byte or size equivalence with ABC's mapped encoder.

Known limitations needing further work: `Cnf_CutCreate` does not implement
ABC's DAR-selected-cut interface; fast cuts with more than six leaves support
AND supergates but reject other cones. Some retained compatibility helpers
remain lightly exercised. The current tests are finite and do not prove all
graph representations or optimization paths correct.

Validation: `GOCACHE=/tmp/abcgo-cache go test -race -count=1 ./sat/cnf` and
`GOCACHE=/tmp/abcgo-cache go vet ./sat/cnf`; CNF/BMC integration was also run
with the race detector after changes to shared encoding and output semantics.
