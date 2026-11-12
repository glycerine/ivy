# Dec07 / Dec08 / Dec10 review

Scope: `ifDec07.go`, `ifDec08.go`, `ifDec10.go`, `ifDec_family_test.go`,
`ifDec_c_test.go`, and `test_vectors/decomposition*`. Shared headers, mapper
commands and other decomposition families were not changed.

## Separate readability/compactness pass

The source's bounded cofactor-matrix algorithms are expressed with native word
slices and explicit variable lists. Dec07 enumerates four-input bound sets and
three-input free sets, including their shared variables for five/six inputs.
Each shared-variable partition admits at most two distinct cofactor patterns;
these directly produce the two LUT4 tables. The source 64-bit configuration
layout is preserved. An equivalent valid configuration may differ from the
source's permutation-dependent choice or pruning of unused inputs.

Dec08 and Dec10 share the signed-cofactor-count/compatible-bound-set search with
three- and four-variable bound sets respectively. The evolving permutation
order is retained for the source's fourth sharing-variable candidate. Their
original active code implements acceptance only: `fDerive` does not generate
configuration records in those C functions, and the port retains that contract.

Common support, cofactor, composition, permutation, minimum-base and MUX helpers
replace duplicated C-static masks, byte tables and unsafe uint32/word casts.
Truth tables are caller-owned and remain unchanged by structural queries.
Diagnostic configuration printing reports tables and selectors without requiring
an external DSD pretty-printer. Production code uses no C, unsafe or stubs.

## Separate bug review

- Independent reconstruction checks every derived LUT4 configuration, including
  all shared/free pin assignments and output polarity.
- Minimum-base operations preserve the compacted support labels and repeat truth
  bits correctly across the full word storage.
- Coffactors above variable five preserve correct word/block offsets; aliased
  adjacent swaps use temporary storage and round-trip correctly.
- The uint32 input adapter explicitly packs words and handles the repeated
  five-variable half-word representation without pointer aliasing.
- Structural cut checks reject nonminimal full support instead of passing it
  into the original internal full-support assertions.
- Zero-leaf acceptance needs no truth-buffer access.
- Configuration verification uses all four selector bits. The C debug verifier
  masked selectors to three bits even for ten-variable inputs, confusing input
  nine with input one (and preventing later LUT references above index seven).
- Invalid configuration references, short truth/cofactor buffers and out-of-range
  permutations report explicit Go panics instead of reading uninitialized or
  out-of-bounds C storage. Unused six-input selector slots have deterministic
  values rather than uninitialized temporary words.

## Validation

Tests independently reconstruct 240 generated five/six/seven-input cascades,
exercise support shrinking, MUX selection, low/high-variable cofactors, swaps,
truth adapters, nonminimal support, parity acceptance and four-bit selectors.
Recorded source/parity examples live in `test_vectors/decomposition_cases.json`.

The optional `TestDecOriginalCComparison` compiles the actual three original
source files into separate C translation units with minimal truth/diagnostic
scaffolding. It compares **960 decomposition decisions** over random truth
tables, parity and generated cascades, and passed here without skipping. C
assertions are disabled for this oracle because the original Dec10 helper
asserts more than six inputs although its public routine permits six. Production
source is not modified, and C is used only by this optional test oracle.

Exact **selected-file** gates, including the existing mapper/header/cut/truth/
count/library suite and source cut oracle:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 $(cat abcgo/map/if/test_vectors/decomposition-files.txt)
GOCACHE=/tmp/abcgo-cache go -C abcgo vet $(cat abcgo/map/if/test_vectors/decomposition-files.txt)
```

Both passed. These are labeled subset gates: the full `go test ./map/if` gate
must wait for the remaining separately owned IF drafts; this review makes no
whole-package completion claim.
