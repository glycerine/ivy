# Packed-word sorting dependency review (reviewed)

This is a dependency of the sequential GIA isomorphism review, not a completed
audit of the whole utility package. Source scope: utilSort.c:511..698 selection
sort, two-way/three-way recursive routines and their wrappers. The original
131-package audit remains active, with2/131 complete.

## Regression and restoration

Go used one middle-pivot two-way sorter for both recursive algorithms and both
wrappers. It sorted low32 keys correctly but changed high32 payload order for
equal keys. GIA uses that payload order to assign tie labels, feeding canonical
PI ordering. New hand-traced Go-only goldens fail before repair:
/tmp/abcgo-iso-sort-payload-red.log. A separate GIA regression fails on its second
packed class member (/tmp/abcgo-iso-canonical-ties-red.log); after repair it also
checks tie-label transfer and the complete canonical PI permutation.

Restore source last-word pivots, preincrement/decrement scans, swaps and the
selection-sort cutoff below11 words. The standalone three-way routines retain
the exact source p/q equal-key moves, including strict k<p and k>q bounds.
Source QuickSort3's wrapper actually calls the two-way routines; preserve that
behavior rather than selecting an algorithm from its name.

The smaller disjoint partition is recursed into while the larger is iterated.
This bounds Go stack use with the same per-partition source swaps and pivot
choices: the order of processing disjoint ranges cannot change their final
packed words. Keep Go's safe empty/base-case check before reading a pivot;
C reads pData[r] before its base check, which can read out of bounds.

## Coverage and verification

Eight exact payload cases cover both wrappers, direct two-way calls and direct
three-way subranges, in both directions. Independent key sorting and payload
bijections check576 cases across sizes0/1/2/9/10/11/12/31/32/33/129/513,
uniform/ascending/descending/duplicate/random/wide keys, both directions and
all four entry modes. Verify range guards and entire words with high-bit payload
IDs. Expected goldens are hand traces; tests neither read nor compile C files.

Frozen full Go-only module exits0:130 tested packages pass,3 commands have no
tests. Full util race1.427s, focused GIA isomorphism race1.046s and util/GIA vet
exit0. Logs /tmp/abcgo-iso-sort-final-{module.jsonl,util-race.log,gia-race.log,
vet.log}. All handles are terminal. Other utility sorting APIs and full package
review are still pending; broader GIA isomorphism review remains open.
