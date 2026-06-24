# Go Toolchain Test Corpus

Imported from `/usr/local/go1.26.4/test`.

This directory is a copied corpus. A small selected subset is wired into the
active Go-junior test runner through `gojr/tests/goToolchainCorpus.test.ts`.

The importer copies Go files from the Go distribution `test/` tree when they
do not lexically use unsupported Go-junior concurrency/recover constructs:

- `go`
- `select`
- `chan`
- `<-`
- `recover`

Comments and string/rune literals are stripped before applying that filter, so
ordinary prose and import paths do not cause false skips. Non-Go support files
are copied unchanged.

See `MANIFEST.json` for exact included/skipped file lists.

Current active coverage executes a small selected `// run` set first:

- `alias1.go`
- `bigmap.go`
- `align.go`
- `char_lit.go`
- `clear.go`
- `decl.go`
- `defer.go`
- `ddd.go`
- `floatcmp.go`
- `helloworld.go`
- `for.go`
- `closure1.go`
- `closure2.go`
- `compos.go`
- `const.go`
- `const3.go`
- `const8.go`
- `func.go`
- `func4.go`
- `func6.go`
- `func7.go`
- `func8.go`
- `if.go`
- `intcvt.go`
- `divide.go`
- `initcomma.go`
- `range3.go`
- `range4.go`
- `typeswitch1.go`
- `varinit.go`
- `mapclear.go`
- `map.go`
- `method3.go`
- `method7.go`
- `newexpr.go`
- `print.go`
- `string_lit.go`
- `abi/convF_criteria.go`
- `abi/convT64_criteria.go`
- `abi/defer_aggregate.go`
- `abi/double_nested_addressed_struct.go`
- `abi/double_nested_struct.go`
- `abi/f_ret_z_not.go`
- `abi/named_results.go`
- `iota.go`
- `literal.go`
- `ken/interbasic.go`
- `ken/interfun.go`
- `ken/array.go`
- `ken/complit.go`
- `ken/for.go`
- `ken/intervar.go`
- `ken/litfun.go`
- `ken/range.go`
- `ken/ptrvar.go`
- `ken/robfor.go`
- `ken/sliceslice.go`
- `ken/simparray.go`
- `ken/simpbool.go`
- `ken/simpconv.go`
- `ken/simpfun.go`
- `ken/strvar.go`
- `ken/string.go`

This smoke set is deliberately small so new failures point at specific missing
Go semantics. It should grow as the runtime gains more standard-library and
dynamic-type fidelity.
