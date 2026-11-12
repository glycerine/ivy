# Native Tree/Merge oracles

Run from repository root. These harnesses use the original C sources, not the Go
implementation. Ordinary Go tests use checked-in results and require no C tools.

String merge fixture (900 exact output rows):

```
gcc -O3 -flto -ffunction-sections -fdata-sections -DABC_USE_STDINT_H -I src \
  abcgo/opt/dau/test_vectors/native_merge_oracle.c \
  src/opt/dau/dauDsd.c src/opt/dau/dauMerge.c \
  -Wl,--gc-sections -o /tmp/dau-native-merge
/tmp/dau-native-merge
```

Its stdout is retained as native_merges.tsv. The first400 rows enumerate every
operand/phase combination; the final500 come from its explicit LCG request stream.

Structural manager fixture:

```
gcc -O3 -flto -ffunction-sections -fdata-sections -DABC_USE_STDINT_H -I src \
  abcgo/opt/dau/test_vectors/native_tree_oracle.c \
  src/misc/mem/mem.c src/opt/dau/dauDsd.c \
  -Wl,--gc-sections -o /tmp/dau-native-tree
/tmp/dau-native-tree
```

The shared merge and bug harnesses require the original merger and canonicalizer:

```
gcc -O3 -flto -ffunction-sections -fdata-sections -DABC_USE_STDINT_H -I src \
  abcgo/opt/dau/test_vectors/native_shared_tree_oracle.c \
  src/misc/mem/mem.c src/opt/dau/dauDsd.c src/opt/dau/dauMerge.c \
  src/opt/dau/dauCanon.c -Wl,--gc-sections -o /tmp/dau-native-shared-tree
/tmp/dau-native-shared-tree
```

Substitute native_tree_bug_oracle.c to reproduce the three source defects recorded
in TREE_MERGE_BUG_REVIEW.md. The native harnesses include original dauTree.c to
access its private types; they do not replace its allocator or decomposition.
