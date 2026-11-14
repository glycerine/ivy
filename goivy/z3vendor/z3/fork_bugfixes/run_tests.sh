#!/bin/bash
# Regression vectors for the bug fixes described in ../FORK_BUGFIXES.md.
#
# usage: fork_bugfixes/run_tests.sh [path/to/z3]
#
# Runs every test_vectors/*.smt2 through z3 (60 s limit each) and compares
# stdout+stderr with the matching .expected file. Without the fixes:
#   relevancy_update_scopes  wrong models (p false although p is asserted)
#   pareto_model_completion  never terminates (timeout)
#   qe_rec_free_boolean      segmentation fault
#   qe_rec_real_disjunction  unsound result (the existential becomes false)
#   pb_strengthening_proof   free(): invalid pointer / abort
#   bv_smul_noovfl_fold      wrong folds (#xa*#x3 false, #b1*#b1 true)
#
# The C++ unit test for the same fixes is src/test/fork_bugfixes.cpp
# (run: test-z3 fork_bugfixes).

here="$(cd "$(dirname "$0")" && pwd)"
z3="${1:-}"
if [ -z "$z3" ]; then
    for c in "$here/../native.build/z3" "$here/../build/z3" "$here/../../native_lib/z3"; do
        if [ -x "$c" ]; then z3="$c"; break; fi
    done
fi
if [ -z "$z3" ] || [ ! -x "$z3" ]; then
    echo "z3 binary not found; pass its path as the first argument" >&2
    exit 2
fi
# The vectors run from test_vectors/, so make the path absolute.
z3="$(cd "$(dirname "$z3")" && pwd)/$(basename "$z3")"
echo "using $z3"

pass=0
fail=0
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
for f in "$here"/test_vectors/*.smt2; do
    name="$(basename "$f" .smt2)"
    (cd "$here/test_vectors" && timeout 60 "$z3" "$name.smt2") > "$tmp" 2>&1
    rc=$?
    if [ $rc -eq 0 ] && cmp -s "$tmp" "$here/test_vectors/$name.expected"; then
        echo "PASS $name"
        pass=$((pass + 1))
    else
        echo "FAIL $name (exit $rc)"
        diff "$here/test_vectors/$name.expected" "$tmp" | head -20
        fail=$((fail + 1))
    fi
done
echo "$pass passed, $fail failed"
[ $fail -eq 0 ]
