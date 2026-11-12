# Timing/size foundation review

Scope: `sclSize.c.go`, `sclSize.h.go`, `sclLoad.go`, `sclUtil.go`,
`sclTiming_test.go`, and `test_vectors/timing_chain.json`. Other SCL source
areas remain separately owned. No draft files are excluded or renamed.

## Separate readability/compactness review

The port keeps ABC's float32 arrival, slew, departure, load and sizing evaluation
algorithms. `SC_Man` uses typed slices and local scheduling sets instead of raw
arrays, traversal marks and pointer arithmetic. Timing-object IDs may be sparse
and need not be topological: the manager derives a topological ordering and
levels from the supplied graph. A maximum-output scan replaces the original
output-priority queue. Source cone/load backups and rollback evaluations remain
available to the separately ported sizing optimizers.

`TimingNetwork` supplies topology without copying it or importing incomplete
`base/abc` source. `NetworkMapping` owns SCL IDs and MIO assignments. The
optional `SclCommitMapping() error` adapter commits converted MIO functions back
to a host graph; adapters must validate before changing host state. Native
network implementations can use the mapping directly. The external ABC graph
adapter remains an integration task for the graph owner.

Reports use plain portable text, propagate write failures, and avoid terminal
color APIs and zero-denominator diagnostics. Stats baselines belong to each
manager rather than process-global static C state. Releasing a Go manager leaves
the network-owned mapping valid.

## Separate bug review and regressions

- Validate normalized ps/fF library units and mapped cell arities.
- Detect missing/duplicate IDs, absent terminals, inconsistent edge multiplicity
  and combinational cycles instead of assuming IDs encode topological order.
- Compute both phases for positive, negative and non-unate timing arcs; missing
  arcs return errors.
- Load-sensitive incremental/full timing regressions exposed stale input
  departures when a changed output load changes an upstream arc but leaves that
  gate's own departure unchanged. Forward-recomputed gates now schedule their
  fanins for reverse recomputation.
- Incremental departure calculation applies the same temporary buffer-tree
  load clamp/logarithmic delay adjustment as full timing.
- Sparse changed-ID holes are rejected. Level checking actually checks the
  manager's cached levels; updates rebuild valid levels.
- Wire-load tables validate shapes/anchors, interpolate explicit zero values,
  sort anchors, extrapolate with the source slope, and handle empty/zero-fanout
  models without indexing entry one of a one-element vector.
- Barrier nodes behave as ideal wires for timing and pass their downstream load
  back to the driver, preserving timing without requiring a temporary graph copy.
- Buffer splitting accounts for every repeated pin occurrence.
- Timing cleanup restores temporarily clamped physical loads, and cone/load
  backup helpers preserve both phases and slew values.
- Min/max sizing and MIO conversions report missing gates instead of writing
  invalid pointers; legacy timing-constraint updates are staged and use the
  supplied frame. Maximum-size lookup cannot loop forever on a malformed ring.

## Tests and limits

The JSON fixture under `test_vectors` drives actual sparse-topology timing,
critical-path, reverse-departure and load calculations with bilinear tables.
Tests also cover 40 deterministic randomized incremental-vs-full recomputations,
phase handling, barriers, wire interpolation/extrapolation, buffer estimates,
backup/restore, malformed topology, missing arcs, sparse holes and stale levels.
The combined gate includes the separate reader/interpolation original-C oracle,
constraints/commands tests and bool_dec's real downsize workflow regression.

Exact combined review gate (the list contains the converted reader/command/
downsize files; the extra timing test must be included):

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 $(cat abcgo/map/scl/test_vectors/timing-files.txt)
GOCACHE=/tmp/abcgo-cache go -C abcgo vet $(cat abcgo/map/scl/test_vectors/timing-files.txt)
```

The full `go -C abcgo test ./map/scl` package gate still encounters generated
buffer/optimization source outside this review's scope. This document does not
claim package completion. Incremental timing retains the source's 0.1 phase
change propagation tolerance; the reported randomized differential uses changes
large enough to propagate. Graph mutations and manager operations require
caller serialization.
