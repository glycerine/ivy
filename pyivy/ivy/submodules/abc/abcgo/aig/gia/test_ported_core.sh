#!/bin/sh
# Focused verification only: the complete GIA package still contains C drafts.
set -eu
cd "$(dirname "$0")/../.."
export GOCACHE="${GOCACHE:-/tmp/abcgo-cache}"
exec go test "$@" \
  aig/gia/object.go aig/gia/object_links.go aig/gia/timing_metadata.go aig/gia/timing_levels.go aig/gia/timing_test.go aig/gia/utility_graph.go aig/gia/distance.go aig/gia/recognize.go aig/gia/recognize_test.go aig/gia/utilities.go aig/gia/mffc.go aig/gia/mffc_test.go aig/gia/giaDfs.go aig/gia/dfs_test.go aig/gia/mapping_accessors.go aig/gia/gia.h.go aig/gia/dup_dfs.go aig/gia/giaSatoko.go aig/gia/mf_driver.go aig/gia/satoko_driver_test.go aig/gia/identity.go aig/gia/ternary.go aig/gia/append_simplified.go aig/gia/class_metadata.go aig/gia/scratch_arrays.go aig/gia/levels.go aig/gia/special_gates.go aig/gia/object_metadata.go aig/gia/literal.go \
  aig/gia/truth.go aig/gia/truth_test.go \
  aig/gia/giaAiger.go aig/gia/giaAigerExt.go aig/gia/giaAiger_test.go aig/gia/giaAigerExt_test.go \
  aig/gia/giaTruth.go aig/gia/gia_truth_port_test.go aig/gia/gia_truth_c_oracle_test.go \
  aig/gia/aiger_io.go aig/gia/aiger_vec.go aig/gia/aiger_io_test.go \
  aig/gia/vector_aliases.go aig/gia/manager_types.go aig/gia/manager.go \
  aig/gia/network.go aig/gia/manager_accessors.go aig/gia/manager_auxiliary.go \
  aig/gia/choice_miter.go aig/gia/choice_miter_test.go aig/gia/metadata.go aig/gia/random.go aig/gia/counterexample.go aig/gia/metadata_test.go aig/gia/sat.go aig/gia/sat_test.go aig/gia/representatives.go aig/gia/fanout.go aig/gia/references.go aig/gia/bridge.go aig/gia/slv.go \
  aig/gia/giaMem.go aig/gia/memory_test.go aig/gia/kit_manager.go aig/gia/support.go aig/gia/giaAgi.go aig/gia/manager_test.go \
  aig/gia/graph.go aig/gia/graph_objects.go aig/gia/graph_test.go \
  aig/gia/enumerate.go aig/gia/enumerate_test.go aig/gia/release_test.go
