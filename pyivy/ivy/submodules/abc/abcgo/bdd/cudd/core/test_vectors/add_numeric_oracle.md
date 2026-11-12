# Original-C ADD numeric fixture

Expected IEEE-configured checksum: `dab9cf682104fafe`.

The original prebuilt ABC CUDD objects use the optional non-IEEE infinity sentinel
`1e302`. Core Go deliberately uses IEEE infinities. To reproduce the same original
C algorithm under CUDD's IEEE configuration, rebuild its constant-table and
initialization objects and use the remaining original objects unchanged:

```bash
cc -DABC_USE_STDINT_H -DHAVE_IEEE_754 -I src -c src/bdd/cudd/cuddTable.c -o /tmp/cudd-ieee-table.o
cc -DABC_USE_STDINT_H -DHAVE_IEEE_754 -I src -c src/bdd/cudd/cuddInit.c -o /tmp/cudd-ieee-init.o
bdd_oracle_objects=()
for bdd_oracle_object in src/bdd/cudd/*.o; do
 case "$bdd_oracle_object" in
  */cuddTable.o|*/cuddInit.o) continue ;;
 esac
 bdd_oracle_objects+=("$bdd_oracle_object")
done
cc -DABC_USE_STDINT_H -DHAVE_IEEE_754 -I src abcgo/bdd/cudd/core/test_vectors/add_numeric_oracle.c /tmp/cudd-ieee-table.o /tmp/cudd-ieee-init.o "${bdd_oracle_objects[@]}" src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/cudd-add-numeric-ieee-oracle
/tmp/cudd-add-numeric-ieee-oracle
```

The non-IEEE build produces `9b9f115f3cf36db6`; its only differing operation here
is reciprocal-of-zero, represented as `1e302` rather than IEEE positive infinity.
The Go fixture uses the IEEE checksum. It exercises real original abstractions,
extrema, reciprocal and bit extraction; there are no native stubs. Float bits
are hashed in an explicit little-endian sequence, including nil-result markers.
Normal Go tests require no C compiler and do not execute this generator.
