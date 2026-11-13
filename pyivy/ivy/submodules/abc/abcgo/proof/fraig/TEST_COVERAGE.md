# FRAIG coverage notes

The FRAIG package reached 78.2% statement coverage with
`FRAIG_C_DIFFERENTIAL=1 go test -race -coverprofile=... ./proof/fraig`.
The percentage includes opt-in compiled-C rehash, compression, and allocator
comparisons. It is not a completeness or correctness guarantee.

The tests include exhaustive small SAT truth tables, source-specific XOR clause
polarity, tagged-vector ordering, feedback covering at the 20-node class
boundary, both sparse and normal compression, and memory recycle/restart. A
five-pigeon/four-hole instance exercises a real conflict-budget unknown and is
compared with compiled C. A 4,096-node cone and repeated mixed-cone feedback
exercise bounded depth and compression stress.

Still missing are representative end-to-end workloads through a ported ABC
consumer; wall-clock timeout outcomes during real search; larger and repeated
C-versus-Go compression comparisons; and stress beyond the bounded deep-cone
and feedback tests. Some zero-coverage exported functions
are diagnostic accessors or alternate/static C paths that production FRAIG does
not normally call. A branch count cannot establish whether those paths preserve
all observable side effects.
