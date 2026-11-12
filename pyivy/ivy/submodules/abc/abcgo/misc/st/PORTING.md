# Symbol tables port

Both `st` and `stmm` are implemented using Go-managed linked bucket entries,
retaining caller-supplied comparison and hash functions. Go GC replaces the
memory-managed variant's allocation pool. No C pointer casts are used.

Exported `Table` methods provide lookup, insert, direct duplicate insertion,
delete, shallow entry copy, clear, foreach with stop/delete, generators, and
stable mutable value slots (`Find`/`FindOrAdd`). `New`, `StringCompare/Hash`,
`NumericCompare/Hash`, and `PointerCompare/Hash` are available across packages.
The original-spelling `st__*`/`stmm_*` functions remain local compatibility entry
points for ported code in this package. Keys/values use `any`; pointer/numeric
identity hash functions accept integer and actual pointer values. String keys
use content comparison and hashing. Hash callbacks must return equivalent
buckets for comparator-equivalent keys.

Growth, reorder-on-lookup, duplicate entries and iterator behavior follow the
original chained tables. Copies duplicate entries but retain key/value object
references. Updating a stable value slot survives rehashing; deletion/clear
invalidates its association with the table. As in C, structural modification
while using a generator is not supported.

`go test ./misc/st` passes, covering growth/replacement, value slots, copies,
foreach deletion/stopping, generators, string and pointer identity, and stmm
cleanup. Source references remain in `src/misc/st`.
