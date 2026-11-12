# Irredundancy and cofactor review

Owned files: irred.go, cofactor.go, irred_test.go. Public package-internal C
entry names are retained for the Espresso consumers. Pointer-punned cube lists
use the shared cubeList contract. Cover rows remain borrowed; each list owns its
cofactor mask and row-vector slice. Coverage excludes precisely the queried row
identity, so equal duplicate rows remain available to cover one another.

## Compactness and readability pass

Reviewed independently of the correctness checks. One shared constructor builds
one-, two-, and three-cover lists. One unate-reduction helper serves Boolean
tautology and covering-table derivation. The current redundant column is an
explicit recursion argument, replacing the C global Rp_current. Recursive
children are constructed before releasing their parent. No extra graph model,
manual pointer sentinel, allocator, or eager minterm enumeration is used in the
implementation. The original binate Shannon recursion, unate pruning, component
partition, relative-essential classification, covering matrix, and minimum-cover
selection remain intact. Debug-only progress printing was omitted.

## Bug-review pass

Checked empty covers, duplicate row identity, multiple-valued variable geometry,
cofactor alias/lifetime, recursion termination, relative-essential flags, row
order, and table column identities. The source simplify_cubelist loop starts at
index three with an initially empty destination, so it drops the first row and
all singleton inputs. The replacement retains the first normalized row and
removes only subsequent duplicates; a singleton/duplicate regression covers it.
The source multiple-valued split balancing updates maxactive from the running
number of active parts rather than the zero count; the implementation uses the
actual maximum per-part zero count, matching the documented balancing criterion.
Tests cover multiple-valued count and split metadata.

Independent tests compare 150 randomly generated mixed-valued covers with direct
minterm enumeration for both tautology and single-cube coverage, and 100 random
irredundant selections for unchanged rows and equivalent coverage modulo don't
cares. Directed tests check duplicate candidate selection, relative-essential
marking, disjoint partition branches, and independently owned cofactor lists.
No oracle calls use the implementation's own recursive coverage to compute the
expected result.

Validation uses the actual converted Espresso source files, explicitly selected
because other package files are still being ported. This subset passes normal
and race tests and vet. Whole-package integration remains the responsibility of
the coordinating Espresso owners while unrelated drafts contain C syntax.
