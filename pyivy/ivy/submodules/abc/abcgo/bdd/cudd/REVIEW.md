# Root CUDD forwarding/integration review

## Compactness and readability pass

Completed original modules forward their public API names and graph types to
`bdd/cudd/core`. Root files contain aliases, not second copies of graph
algorithms. This preserves the original module/API names while giving completed
consumers an importable, tested implementation. The authoritative implementation
and its original source notices are in core. `cuddInt.go` no longer defines a
second manager struct; the public root manager/node types alias core exactly.

## Bug and integration pass

- Compiled the completed root forwarding files together with explicit file
  selection to verify exported aliases and callback/type identity resolve.
- Fixed generated internal ZDD aliases to preserve the original `cuddZdd*`
  capitalization, so later original-module ports can use the expected names.
- Concrete behavior is tested once in core, including original-signature typed
  callbacks; aliases add no behavior or separate caches.
- Reviewed the package-level status independently of those narrow checks:
  `go -C abcgo test ./bdd/cudd` still fails in unported `cuddAPI.go` at retained
  preprocessing directives and C declarations. Root CUDD is not marked complete.

The root Go source inventory still contains approximately 68,000 lines, including
unported ADD/ZDD algorithms, dynamic reordering, exports and other API modules.
All remaining source files stay visible to the compiler; no build exclusions,
hidden originals, or placeholder algorithms make the root package appear green.
See `core/REVIEW.md` for the separate algorithm/readability review passes and
regressions of each completed core API group.

## Partial ZDD cover module update

The six completed public cover product/division APIs and their recursive names
now forward to core. Pointer-output cofactor helpers are small status-convention
adapters over checked core utilities. The completed prefix was formatted and
compiled separately against the real core to check forwarding types. The original
public/internal ZDD complement draft bodies remain in `cuddZddFuncs.go`, clearly
marked, because their BDD-cover/ISOP dependencies remain unported. No fake logical
complement or build exclusion has been supplied to make the root package green.

## ISOP completion update

The prior partial-cover-module note is superseded: original ISOP, BDD-only ISOP,
cover-to-BDD conversion and logical ZDD complement now forward to oracle-verified
core implementations. The retained complement drafts are removed because their
actual dependencies have been implemented, not because they were hidden or
replaced with a placeholder. The complete ISOP/cover forwarding modules are
formatted and compile together in the selected-file root audit. Root CUDD as a
whole still has other unported modules and is not marked complete.
