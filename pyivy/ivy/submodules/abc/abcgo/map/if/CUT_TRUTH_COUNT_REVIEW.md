# Cut, truth and log-counter algorithm port

Scope: all active algorithm families in `ifCut.c`, `ifTruth.c` and `ifCount.h`, implemented in `ifCut.go`, `ifTruth.go`, `ifCount.go`. Shared manager/header ownership remains with the map/if lead. No experiments or generated-source exclusions were added.

## Design and readability review

Cut union uses a single native-slice merge path, preserving ordered and first-input-order variants, signatures, shared masks and all permutation maps. Failed bounded merges leave the destination intact. Source sorting priorities are explicit metric sequences; power and regular modes retain their separate tie ordering. Filtering and incremental sorting retain the scratch candidate slot and first-cut protection behavior.

Area/edge/power reference operations share one traversal while preserving zero-to-one and one-to-zero fanout rules. A local journal restores counts if malformed input causes a panic. Cone verification follows choices with cycle detection; traversal scratch stays local.

Truth tables use owned word slices and per-support Vec_Mem interning. Projection explicitly tracks leaf identities and canonical input phases; support minimization retains compact leaf order. The source canonicalizer is reused through the buildable `opt/dau/canon` leaf package. Disjoint canonical pair caching retains the source key/result/permutation semantics. ISOP complement metadata uses the manager's typed companion slice instead of modifying vector capacity bits.

Log counting shares queue movement logic across plain arrival times, generated AND/XOR networks and packed pin-delay profiles. Packed delays and word counters retain their original representation. AIG-array verification uses native truth storage instead of aliasing an integer vector as a word vector.

## Separate bug review

Checked scratch-slot ownership, bounded merge rollback, duplicate/shared leaves, six ranking modes, flow division, shared-DAG reference rollback, complemented normal and canonical truth tables, support removal, cache phases and IDs, orbit generation, logarithmic carry ordering, profile saturation, truth-array topology and large shifts.

Regression findings were reproduced and corrected:

- Lazy truth caches initially lacked the source-reserved constant and variable IDs. Initialization now uses Vec_MemAllocForTT, preserving pair-cache eligibility and ID conventions.
- Zero-variable truth orbit generation discarded high truth bits. Permutations now preserve assignment bits outside the permuted support.
- Dominance filtering initially omitted the source assertion tying its candidate to the cut-set scratch slot. An explicit check now rejects invalid ownership before modifying the set.

The native 32-bit log counter handles shifts of 32 or more explicitly, retaining discarded-bit stickiness instead of relying on undefined C shifts. Empty-cut average references return zero. Reference mutations roll back on invalid cones, and failed merges do not expose partial destination/permutation updates.

## Validation

- 2,000 seeded bounded cut unions checked against an independent set-union oracle, including shared-variable permutation maps and destination preservation on failure.
- Shared-cone area ref/deref evaluation, reversible accounting and malformed-underflow rollback.
- 500 seeded logarithmic sum sequences checked against independent exact uint64 sums, plus balanced pin-profile and generated-AIG truth checks.
- 100 seeded normal/canonical truth compositions, all assignments, complemented support elimination, disjoint pair-cache reuse across different leaf IDs, pin permutation and NPN orbit checks.
- Original C differential oracle: 1,500 seeded mixed cases comparing merged leaves/signatures/shared masks/permutation maps, source log-counter queue states, and every power/regular ranking mode.

Focused actual-source command (the rest of map/if is still being integrated):

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race \
 map/if/if.go map/if/ifNativeTypes.go map/if/ifCut.go \
 map/if/ifTruth.go map/if/ifCount.go map/if/assert.go \
 map/if/ifCut_native_test.go map/if/ifTruth_native_test.go \
 map/if/ifCount_native_test.go map/if/ifCut_c_test.go
```

The same explicit file set passes vet. The optional original-C oracle needs a C compiler and supported ELF linker; production algorithms remain pure Go.
