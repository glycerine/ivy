#!/bin/sh
# Focused real-source tests; other SAIG families are ported independently.
set -eu
cd "$(dirname "$0")/../.."
export GOCACHE="${GOCACHE:-/tmp/abcgo-cache}"
exec go test "$@" \
 aig/saig/saigSimFast.go aig/saig/saigSimSeq.go aig/saig/saigPhase.go \
 aig/saig/saigSimFast_test.go aig/saig/saigSimSeq_test.go \
 aig/saig/saigSimSeq_reference_test.go aig/saig/saigPhase_test.go \
 aig/saig/saigPhase_reference_test.go
