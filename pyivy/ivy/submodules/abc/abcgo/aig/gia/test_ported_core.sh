#!/bin/sh
# Focused verification only: the complete GIA package still contains C drafts.
set -eu
cd "$(dirname "$0")/../.."
export GOCACHE="${GOCACHE:-/tmp/abcgo-cache}"
exec go test "$@" \
  aig/gia/object.go aig/gia/object_metadata.go aig/gia/literal.go \
  aig/gia/truth.go aig/gia/truth_test.go \
  aig/gia/aiger_io.go aig/gia/aiger_vec.go aig/gia/aiger_io_test.go \
  aig/gia/vector_aliases.go aig/gia/manager_types.go aig/gia/manager.go \
  aig/gia/network.go aig/gia/manager_accessors.go aig/gia/manager_auxiliary.go \
  aig/gia/sat.go aig/gia/sat_test.go aig/gia/representatives.go aig/gia/fanout.go aig/gia/references.go aig/gia/bridge.go aig/gia/slv.go \
  aig/gia/giaMem.go aig/gia/memory_test.go aig/gia/kit_manager.go aig/gia/support.go aig/gia/giaAgi.go aig/gia/manager_test.go \
  aig/gia/graph.go aig/gia/graph_objects.go aig/gia/graph_test.go \
  aig/gia/enumerate.go aig/gia/enumerate_test.go aig/gia/release_test.go
