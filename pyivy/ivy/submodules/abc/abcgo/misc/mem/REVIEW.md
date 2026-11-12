# Memory allocator review

Readability and idiomatic-Go pass: replaced raw pointer free lists and page casts
with typed byte/word slices, retained the original allocator API families and
handle encoding, removed an unused Step field, and avoided repeated entry lookups
while recycling words. The package uses the standard library only.

Separate bug pass fixed neighboring allocations being exposed through excess
slice capacity, restarted arenas retaining released chunks in slice backing
arrays, and the maximum fixed word-entry size producing an empty free list.
Normal handle ordering remains unchanged.

Regression tests cover append isolation, released backing references, maximum-size
entries, stable growth, byte/word recycling, restart, size-class mapping, encoded
handles, and flexible-page release.

Validation: `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./misc/mem` passed.
