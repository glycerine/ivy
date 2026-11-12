# Separate Tree/Merge bug review

Native source comparison and tests establish these corrected source bugs:

- Dss_NtkCreate("1") creates zero in C. The native bug harness prints
  constant1=0; the Go constant-one network retains the complemented zero edge.
- The C string merger handles constant identities before applying permutations.
  Its 1 AND a with a->c returns a. Go applies mappings first and returns c.
- A nested negated prime loses its explicit phase when C overwrites fCompl with
  canonicalization's output phase. The source bug harness produces
  9600960096009600 for (!96{abc}d); the correct Go result is
  6900690069006900. Both parsing and rebuilt-manager truth are checked.
- The C shared-map initializer {-1} marks only the first element absent; the Go
  map explicitly rejects an alleged shared variable absent from operand zero.

Separate review of the new implementation found a prime hash-lookup mismatch:
insertion cropped the truth to its fanin width, while lookup included unrelated
caller scratch-tail words. A minimal red test observed the missed lookup; key
construction now consistently uses the meaningful prime truth only. The test
retains that regression. Owned results also prevent later queries from changing
previous permutations/tables, and constant merges initialize their result map.

Validation evidence:

- Every one of the256 three-variable truth functions passes decomposition,
  parsing, prime canonicalization, structural permutation, rebuilding and exact
  truth reconstruction.
- Five hundred independently randomized shared/permuted/complemented formula
  merges match direct Boolean evaluation.
- Nine hundred complete merge strings match unchanged original C: all400
  operand/phase combinations of the fixture pool plus500 reproducible requests.
- Native structural IDs, printed shape, truth tables and input permutations match.
  Native shared merge ID8, permutation[5,2,0] and truth0808080808080808 match.
- Regressions cover cache hit/miss behavior, MPM shared mapping, prime-size limits,
  extended composition, malformed syntax, repeated network variables, brace
  flattening and output ownership.

The combined DSD/DSD2/Tree/Merge explicit-file group passes race tests and vet.
The complete opt/dau package still contains unrelated drafts; the latest full
package check stopped in dauGia.go. No such drafts are hidden by this family.

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race opt/dau/dauDsd.go opt/dau/dauDsd2.go opt/dau/dauTree.go opt/dau/dauMerge.go opt/dau/dau_tree_types.go opt/dau/dau.h.go opt/dau/hierarchical.go opt/dau/truth_counts.go opt/dau/dauCount.go opt/dau/tree_merge_test.go opt/dau/dsd_native_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet opt/dau/dauDsd.go opt/dau/dauDsd2.go opt/dau/dauTree.go opt/dau/dauMerge.go opt/dau/dau_tree_types.go opt/dau/dau.h.go opt/dau/hierarchical.go opt/dau/truth_counts.go opt/dau/dauCount.go opt/dau/tree_merge_test.go opt/dau/dsd_native_test.go
```
