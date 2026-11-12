# RPO port review

The package ports ABC's read-polarity-once factorization and literal-cluster algorithms from `src/bool/rpo`, originally by Mayler G. A. Martins and Vinicius Callegaro.

## Compactness, idiomatic Go, and readability pass

Performed separately from the bug review below.

- Truth tables use `[]uint32` in the existing KIT stretched layout. Expressions use immutable Go strings; cluster and edge collections use slices.
- Transitions use word masks or paired word blocks rather than allocating three C cofactor temporaries. A single source snapshot preserves aliased input/output behavior.
- AND/OR grouping checks inspect words directly without temporary complement/intersection arrays.
- Recursive branches explicitly restore their two cluster slots before returning. Candidate ordering retains endpoint-degree cost and minimum-endpoint tie breaking.
- C bitfield edges become ordinary named Go fields. Literal fields are exported as `Function`, `Transition`, and `Expression` for consumers in other packages.
- Kept source-corresponding public names; small representation helpers are private. No dependency on KIT/vector packages, no cgo, unsafe, stubs, or hidden source files.

## Separate bug review

Found and addressed the following issues with focused coverage:

1. Clearing a transition output before reading aliased input destroyed the source truth table. A snapshot now precedes writes. `TestTransitionSupportsAliasedOutput` was observed failing before the fix.
2. The original constant-literal constructor called its transition assignment “wrong but no effect” and aliased caller-owned truth memory. Constants now have a zero transition and own their function copy. `TestConstantLiteralTransitionAndOwnership` was observed failing before the transition fix and checks ownership.
3. The C truth printer iterated by the variable count instead of the number of words. The port iterates the packed truth-table word count.
4. The source candidate loop included self pairs despite allocating for distinct pairs. The port uses distinct endpoints and dynamically sized slices, keeping the active-cluster-count invariant even if both AND and OR candidates exist.

Reviewed threshold semantics: positive limits bound recursion calls, while zero/negative limits remain unlimited as in ordinary source use. Reviewed returned-cluster ownership and restoration on success/failure.

## Validation

- Full package tests: `GOCACHE=/tmp/abcgo-cache go -C abcgo test ./bool/rpo -count=1`.
- Race tests: `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./bool/rpo -count=1`.
- Exhaustive three-variable coverage: all 256 input tables; successful expressions evaluated independently against every assignment.
- Manual differential test against compiled original C: all 256 tables match exact expression strings and threshold failures. The reference harness avoids freeing the original constant constructor's borrowed stack memory, which otherwise aborts.
- Multiword factorization and assignment-by-assignment transition checks through eight variables, both polarities, all variable positions.
- Constants, negation, AND/OR/nested groups, edge tie breaking, recursion thresholds, cluster restoration, and aliased buffers.

No experiment directories were tested.
