#!/bin/sh
# Additional real AIG integration; requires the AIG package itself to build.
set -eu
exec "$(dirname "$0")/test_ported_core.sh" "$@" \
  aig/gia/aig_synthesis.go aig/gia/giaBalAig.go aig/gia/giaBidec.go aig/gia/giaCSat.go aig/gia/circuit_sat_test.go aig/gia/circuit_sat_oracle_test.go aig/gia/bidec_test.go aig/gia/giaBalLut.go aig/gia/giaBalLut_test.go aig/gia/giaBalLut_reference_test.go aig/gia/giaBalMap.go aig/gia/balance_window_test.go aig/gia/balance_test.go aig/gia/balance_oracle_test.go aig/gia/dup_muxes.go aig/gia/giaMuxes.go aig/gia/muxes_test.go aig/gia/normalize.go aig/gia/aig_synthesis_test.go aig/gia/aig_conversion.go aig/gia/aig_representatives.go \
  aig/gia/aig_conversion_test.go
