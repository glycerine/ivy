#!/bin/sh
# Regenerates default.pgo, the CPU profile `go build` uses for profile-guided
# optimization of the abc binary (about 3-5% on mapping, synthesis and BDD
# workloads). Run from abcgo/ after hot code changes; needs ../i10.aig and
# .scratch/bench/mul64.aig.
set -e
T=$(mktemp -d)
go build -o "$T/abc" ./cmd/abc
S="read .scratch/bench/mul64.aig; strash; if"
ABCGO_CPUPROFILE=$T/p1.prof "$T/abc" -c "$S; $S; $S; $S; $S; $S; $S; $S; $S; $S" >/dev/null
ABCGO_CPUPROFILE=$T/p2.prof "$T/abc" -c "read .scratch/bench/mul64.aig; strash; balance; rewrite; refactor; resub; dc2; if -K 6; mfs; lutpack" >/dev/null
ABCGO_CPUPROFILE=$T/p3.prof "$T/abc" -c "read ../i10.aig; collapse; sop; strash; fraig; dch; if; mfs" >/dev/null
go tool pprof -proto "$T/p1.prof" "$T/p2.prof" "$T/p3.prof" > cmd/abc/default.pgo
rm -r "$T"
