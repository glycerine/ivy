# SOP helpers: native port and review

All defined functions in `src/base/abc/abcSop.c` are ported in `abcSop.go`.
Covers remain immutable strings with fixed-width `variables + 3` rows. Constants
remain ` 0\n` and ` 1\n`. Standard tags 0/1 and special XOR tags x/n retain their
source phase and predicate behavior. Constructors preserve literal ordering,
including the intentional binary-text MSB-first versus hexadecimal-text LSB-first
variable order. MV encoders/decoders retain their BLIF-MV text exactly.

## Native API contracts

Arena parameters are real `*mem.Mem_Flex_t`; returned strings own their storage
independently of arena lifetime. C's complement mutators return replacement
strings in Go: callers must assign the return value. Complement flags and truth
buffers use slices; ISOP covers use the native `*vec.Vec_Int_t`. Truth conversion
uses uint64 words, and Big conversion keeps explicit variable tables and scratch
buffers. Counts, phases, and predicates retain int32 return values. Source NULL
string results map to empty strings: this includes all-zero nonzero-width minterm
construction and the source binary-text constant rejection. Constructors retain
zero-variable constant normalization.

The source's standard truth-conversion routines do not interpret x/n as parity;
they compute the ordinary cube union and apply the output phase. This behavior
is preserved. Network strashing handles special XOR tags separately. ISOP uses
the actual native Kit implementation, with source variable/cube encoding and
optional output complement. Invalid malformed inputs now fail safely; Check
also rejects invalid input-pattern characters that C's checker did not inspect.

## Readability and compactness pass

Completed separately from the correctness pass. Repeated one-cube constructors
share a small builder; all word-width truth APIs share one evaluator. Strings
replace byte-pointer arithmetic and manual arenas. uint32 cube encodings avoid
signed shifts at the sixteenth literal. Named helpers keep source ordering and
phase rules visible. There are no unsafe conversions, global scratch buffers,
closure-generated output functions, stubs, or hidden source files.

## Independent bug pass

Checked all constructor polarity conventions, complemented covers, x/n tag
swapping, duplicate ordinary/phase-encoded buffer and inverter predicates,
source cube/literal counts, binary/hex ordering, ISOP packed two-bit literals,
truth-table replication below six inputs, the seventh variable's high word,
and caller-provided Big scratch dimensions. Registration and transforms retain
independent immutable string values. Source zero-length/NULL cases are mapped
explicitly instead of dereferencing C-style pointers.

Validation includes constructor truth tables through six variables, 300 random
ISOP and cube-encoding round trips through seven variables, eight-input Big
truth, output/input complementation, invalid covers and textual inputs, and MV
codecs. An actual unchanged-C object oracle checks 500 random covers: all count,
phase, buffer/inverter/AND/OR/XOR predicates, truth words, output complements,
and variable complements match exactly. The oracle actually ran during review;
missing compilers/object files produce visible skips.

Focused normal/race and vet commands:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race base/abc/abcSop.go base/abc/abcSop_test.go base/abc/abcSop_reference_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet base/abc/abcSop.go base/abc/abcSop_test.go base/abc/abcSop_reference_test.go
```

Native Ntk consumer tests cover original-C network variants, metadata/local
function-manager duplication, sequential terminal conversion, and permutations.
The ordinary `go -C abcgo test ./base/abc` gate was run and still fails at
unrelated `abcHieGia.go` syntax. Focused results do not claim the package is green.
