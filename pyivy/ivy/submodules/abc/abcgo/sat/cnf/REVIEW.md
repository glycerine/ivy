# CNF review

The correctness review covered graph encoding, output constraints and phases,
cut composition and mapping, clause metadata, frame loading, variable lifting,
packed covers, and DIMACS I/O. Fixes and regressions are recorded in the commit
history. Seeded graph and cut oracles compare small cases against independent
Boolean evaluation; frame tests exercise initial state and a latch transition.

The readability pass consolidated serialization and leaf traversal and uses
the shared ISOP implementation for cut costs. `EncodeAIG` uses a direct
Tseitin representation; `Cnf_Derive` uses ABC's mapped cut-based encoding.
Direct comparisons against C cover hundreds of mapped AND/inverter graphs.

The source BSAT solver can report SAT for an inconsistent formula when pending
root units and assumptions interact. The `EncodeAIG` randomized regression
records two such results reproduced by the original C solver. See the BSAT
porting notes for the source behavior and its implications.

Known limitations needing further work: fast cuts with more than six leaves
support AND supergates but reject other cones. Some retained compatibility helpers
remain lightly exercised. The current tests are finite and do not prove all
graph representations or optimization paths correct.

Validation: `GOCACHE=/tmp/abcgo-cache go test -race -count=1 ./sat/cnf` and
`GOCACHE=/tmp/abcgo-cache go vet ./sat/cnf`; CNF/BMC integration was also run
with the race detector after changes to shared encoding and output semantics.
