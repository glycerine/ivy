# FRAIG coverage notes

The FRAIG package reached 78.7% statement coverage with
`FRAIG_C_DIFFERENTIAL=1 go test -race -coverprofile=... ./proof/fraig`.
The percentage includes opt-in compiled-C rehash, compression, and allocator
comparisons. It is not a completeness or correctness guarantee.

The tests include exhaustive small SAT truth tables, source-specific XOR clause
polarity, tagged-vector ordering, feedback covering at the 20-node class
boundary, both sparse and normal compression, and memory recycle/restart. The
compiled-C compression suite also compares a six-node class and internal cone
across two rounds and a permanent-word boundary. A
five-pigeon/four-hole instance exercises a real conflict-budget unknown and is
compared with compiled C. A 4,096-node cone and repeated mixed-cone feedback
exercise bounded depth and compression stress. An eight-bit ripple adder is
checked against integer addition for 1,024 deterministic input pairs.
The public API tests also replay a miter counterexample, and two successive
simulation-buffer resizes check preserved PI words, internal hashes and masks.

Manager allocator accounting checks typed node-pool and pooled simulation
storage through initial allocation, chunk growth, and a simulation resize. The
8-bit and 16-bit ripple adders check arithmetic graphs against integer addition
for deterministic input sets. The 16-bit case exercises larger pooled node and
simulation storage, but these workloads do not replace end-to-end execution
through an ABC consumer or compiled-C workload comparisons.

Still missing are representative end-to-end workloads through a ported ABC
consumer; wall-clock timeout outcomes during real search; broader sparse and
larger-class C-versus-Go compression comparisons; and stress beyond the bounded deep-cone
and feedback tests. Some zero-coverage exported functions
are diagnostic accessors or alternate/static C paths that production FRAIG does
not normally call. A branch count cannot establish whether those paths preserve
all observable side effects.
