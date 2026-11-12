# FXCH design and compactness review

Reviewed the complete package independently of the bug pass. This implements the
2016 Bruno Schmitt/UFRGS FXCH algorithm from src/opt/fxch; it does not alias FXU,
Kit, or the different base/pla extraction experiment.

Rows carry their owner ID followed by sorted ABC literals. Initial identical
cubes are grouped across outputs. Go sets replace packed output masks, supporting
arbitrary output counts without manually cached word lengths. Single-cube costs
are multiplied by output membership; double-cube costs by common output count.
The actual one/two-literal-deletion subcube enumeration, literal-sum hash keys,
full residual comparison, normalized two-literal/4-literal MUX divisors, De Morgan
sharing and float32 depth-biased weights are retained. The MWC random hash-key
generator has manager-local state rather than a shared global RNG.

The public source function spellings remain available with native Go slices and
pointers for slice mutation. Cubes, heap/table sizes, fan counts and variable
counts derive from slices; unused original counters, translation scratch arrays,
pool fields and cached capacity are removed. Levels encode real algorithm state.
The subcube table maps complete sum hashes to slices; full residual comparison
protects against collisions. Memory reporting includes allocated descriptor slices.

Candidate maximum selection scans stable divisor insertion order. Candidates are
rebuilt after rewrites rather than maintained through the original incremental
queue/subcube delete bookkeeping. This simplifies the intertwined update invariants
but can change subsequent equal-weight tie order and large-network performance.
It does not promise identical multistep extraction traces or source throughput.
The exported removal operation still performs the actual inverse divisor-weight
and pair-membership updates and tests collision-specific deletion.

Grouped-output extraction implements exact and partial-overlap rewrites, preserving
nonshared memberships. SCC cleanup removes memberships only where a subsuming cube
is present for the same output. Node definitions remain ordinary visible rows.
The FXCHNetwork boundary supports a concrete graph adapter without importing an
unfinished ABC graph package. Such adapter implementation is outside this package.
