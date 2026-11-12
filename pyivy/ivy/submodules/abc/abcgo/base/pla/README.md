# ABC packed PLA manager

Parse(io.Reader,name) and ReadFile(name) return a Pla_Man_t and an error.
Pla_ReadPla(name) retains the C-style nil-on-failure wrapper. The model exposes
Cube input/output word slices with DASH=0, ZERO=1, ONE=2, FULL=3. Evaluate reports
explicit matching-row 0/1 membership; it does not borrow Espresso's semantics or
process-global variable geometry. Write and Pla_WritePla return write errors.

Source algorithms include the native random/prime/sorter generators, pairwise
and hashed cube relations, simple minterm/edge cover, literal occurrence
conversion and two-literal fast extraction. Fxch_ManStart/FastExtract expose the
extracted network; the source Pla_ManPerformFxch compatibility routine consumes
its temporary arrays and discards that network, as the original did.

Original-C oracle reproduction from repository root:

```
gcc -O2 -ffunction-sections -fdata-sections -DABC_USE_STDINT_H -I src \
  abcgo/base/pla/test_vectors/native_oracle.c src/base/pla/plaRead.c \
  src/base/pla/plaWrite.c src/base/pla/plaMan.c src/base/pla/plaHash.c \
  -Wl,--gc-sections -o /tmp/pla-native-oracle
/tmp/pla-native-oracle
/tmp/pla-native-oracle abcgo/base/pla/test_vectors/fdr.pla
/tmp/pla-native-oracle abcgo/base/pla/test_vectors/fd.pla
```

The harness supplies deterministic filename/timestamp/random dependencies while
compiling the original parser/manager/hash sources. Their complete packed outputs
are retained as Go test goldens. The no-argument run tests random generation.

```
gcc -O3 -flto -ffunction-sections -fdata-sections -DABC_USE_STDINT_H -I src \
  abcgo/base/pla/test_vectors/native_fxch_oracle.c \
  -Wl,--gc-sections -o /tmp/pla-native-fxch
/tmp/pla-native-fxch
```

The latter harness includes the unchanged experimental plaFxch.c, compares its
complete extracted state, and uses LTO to discard disabled diagnostic/ABC UI
branches. No native tools are needed to run ordinary Go tests.
