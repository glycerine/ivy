#!/bin/sh
# Additional real AIG integration; requires the AIG package itself to build.
set -eu
exec "$(dirname "$0")/test_ported_core.sh" "$@" \
  aig/gia/aig_synthesis.go aig/gia/aig_synthesis_test.go aig/gia/aig_conversion.go aig/gia/aig_representatives.go \
  aig/gia/aig_conversion_test.go
