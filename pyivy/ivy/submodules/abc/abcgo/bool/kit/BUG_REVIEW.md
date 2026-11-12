# Independent Kit bug review

This pass was performed independently of the package author's algorithm port
and misc_avl's compactness/readability audit. Source edits were coordinated with
the package author before changing kitGraph.go or kitCloud.go.

## Reproduced defects and fixes

1. **Truth-table input paths used parent-directory digits as the variable count.**
   `Kit_TruthFindVarNum` scanned the entire path, inherited from the C helper.
   A regression creates a real binary three-variable table under `run123/` in a
   temporary directory, verifies the count and calls `Kit_TruthTest`. Before the
   fix it inferred an unsupported count (or -1 when temporary-path digits
   overflowed int32). The helper now scans `filepath.Base(name)`. The real file
   load and graph-statistics path succeeds.

2. **Cloud scratch reuse exposed stale trailing truth words.**
   `kitCloudStore` only grew slices. After a six-variable truth (two uint32 words)
   followed by a two-variable truth using the same store, the returned result
   remained two words long and contained stale data (`66666666 99999999`).
   The helper now re-slices every active row to the exact current word count,
   using capacity for reuse. The regression checks both returned shape and data.

Both regressions were observed failing before their source fixes and pass after.

## Additional independent semantic checks

Read and checked truth movement/cofactor implementations, graph construction and
AIG bridge, SOP division scratch/ownership behavior, PLA conversion, LUT-network
truth/arrival propagation, Cloud lifecycle/canonicalization, packed MUX topology,
and cofactor-support composition. No duplicate state or test-only runtime path
was introduced by the fixes.

Added an independent exhaustive reference for cofactor supports over all 256
three-variable Boolean functions. It builds both direct and reversed-variable
Cloud programs, runs `Kit_TruthCofSupports`, and compares every variable/phase
support with a direct minterm-flip reference. This checks the two-pass support
merge and variable orientation without using the implementation under test to
compute expected results.

Existing algorithm tests remain active, including exhaustive three-variable
ISOP/factoring and DSD cases, random multiword care-set ISOPs and packed Cloud
round trips. Validation commands: `go test -race ./bool/kit -count=1` and
`go vet ./bool/kit` from abcgo. No starting-point files were changed.
