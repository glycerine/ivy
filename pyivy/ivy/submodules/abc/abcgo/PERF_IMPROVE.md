# ABC performance investigation

## Starting inventory

The C source has ten subsystem families. The Go module mirrors each family.

| Family | C subdirectories | Responsibilities |
| --- | ---: | --- |
| aig | 7 | Graph representations including MiniAIG, Hop, AIG and GIA, traversal, simulation and graph I/O |
| base | 14 | Network and word-level models, shell, file formats and command dispatch |
| bdd | 9 | Decision diagrams, reordering and decomposition |
| bool | 7 | Boolean functions, truth tables, covers and decomposition |
| map | 9 | LUT and standard-cell technology mapping, libraries and timing |
| misc | 16 | Vectors, allocators, hashing, utilities and compression |
| opt | 20 | Logic optimization including rewriting, factoring and retiming |
| phys | 1 | Physical design |
| proof | 12 | Equivalence checking, abstraction and other proof engines |
| sat | 13 | SAT solvers and solver interfaces |

The Go commands are abc-demo (read, balance, rewrite, write, verify) and abc-starter (parallel shell command runner). The integration package connects MiniAIG and SAT. The checkout contains 54 Go benchmark functions, including focused graph traversal, mapping, simulation and truth-table workloads. Historical build results in PORTING.md are dated.

## Baseline workloads

The bundled C executable reads i10.aig with 257 inputs, 224 outputs, 2,675 AND nodes and depth 50 using the command `&r i10.aig; &ps`. This circuit supplies a shared input for an AIGER load benchmark. BenchmarkDfsSupportWalks exercises support, input collection, cone size and sequential collection on a 1,024-AND chain. Measure both allocation volume and runtime. Candidate experiments include graph construction allocation locality, traversal scratch storage and integer decoding. Select changes using profiles.

### Counting a shared GIA cone

On the Threadripper 3960X (linux/amd64, Go 1.27.3), three runs of the existing 1,024-AND cone benchmark took 33.15–34.08 µs/op, 25,200 B/op and 11 allocations. The CPU profile across all four DFS benchmarks attributed 58.5% cumulative CPU to giaDfsCollect. A C driver linked against the checkout's object files constructs the same chain, calls Gia_ManConeSize 200,000 times and checks the sum is 204,800,000. It took 5.08 µs/op of process CPU. A gprof flat profile of that driver attributed all 100 samples to Gia_ManConeSize_rec.

Count ANDs at first visit when the caller only needs a count. Ordered node collection still emits on the return visit. Eight local stack slots cover small traversal frontiers. The revised Go cone benchmark takes about 10 µs/op with zero heap allocations. Support also drops from one allocation to zero. The added test exercises nested and repeated roots plus a constant. The C and Go timing methods differ (process CPU and elapsed respectively), so small timing differences should be checked in matched runs.

### Loading i10 through Gia_AigerRead

BenchmarkAigerLoadI10 reads the checked-in file through the file API, requests default structural hashing, verifies graph dimensions and stops the manager each iteration. A matched C file-API driver did 4,000 iterations at 135 µs/op elapsed. The gprof flat profile put 27% of samples in Gia_ManHashFind, 24% in Gia_AigerReadFromMemory and 20% in Gia_ManHasDangling. Go initially took 2.15 ms/op, 807 KB/op and 8,081 allocations/op. Its CPU profile showed map operations throughout the parser and hash pass. Gia_ManAppendObj accounted for 49% of sampled allocation bytes.

Binary AIGER variables have the same IDs as the GIA objects during initial decoding. Using these IDs directly and sharing a single definition/remapping table for sparse ASCII inputs reduced allocation to 654 KB/op and 8,023 allocations/op. Three elapsed measurements were 1.38, 1.63 and 1.23 ms/op. Tests cover sparse and dense numbering, polarity and undefined outputs. Other package tests ran concurrently, so use isolated repetitions for precise time comparisons. Manager object allocation is the next candidate.

### Allocate GIA objects in bounded blocks

Gia_ManAppendObj now takes the next zero-initialized slot from a block of up to 256 objects. The manager still exposes stable object pointers and the ID-indexed pointer slice. The cap limits speculative allocation for graphs that start with a large capacity hint. A regression test holds a CI pointer through multiple blocks and a pointer-table reallocation.

On i10, three measurements took 0.90–1.08 ms/op and 1,737 allocations/op, down from 8,023. Total allocated bytes stayed around 654 KB/op. The reduction is in heap object count. The targeted pointer and variable-numbering tests passed; a module-wide test run was started for this cross-cutting change.

### Hash gates during AIGER decoding

The C reader hashes each gate as it decodes it, then checks the finished graph for dangling logic. The former Go reader built an unoptimized graph and copied reachable gates into a second manager. Reachability before simplification missed gates made dangling by a constant fold. The Go reader now uses the GIA structural hash during decoding, creates a binary-literal remap only on the first changed gate, and runs cleanup on the simplified graph before installing metadata. Object-indexed extensions still require original IDs. Simple mode also applies the constant and equal-input simplifications found in C.

A two-output fixture checks a gate made dangling by a later simplification. Calling the C file API directly produced (AND count, output driver literals) of (1, 1, 7) in default mode, (2, 1, 9) in simple mode, and (3, 9, 11) with strashing skipped. The Go regression test matches each result and checks indexed metadata. Driver marking follows backward edges in one scan, including a real MUX's third input. A separate test starts with stale marks on every object.

Three i10 reads took 652–701 µs/op, 343 KB/op and 1,687 allocations/op. The initial single-manager experiment reached 854–875 µs/op with 7,261 allocations/op before removing temporary fanin slices from driver marking. The AIGER and dangling-graph tests passed. Reproduce the Go benchmark from abcgo with `go test ./aig/gia -run '^$' -bench BenchmarkAigerLoadI10 -benchmem -count=3`.
