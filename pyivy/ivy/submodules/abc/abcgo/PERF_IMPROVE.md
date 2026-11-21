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

### Structural hash insertion and reproducible C workloads

C reserves the collision-link vector to the bucket count and retains the insertion slot if appending a gate cannot move the links. Ported both operations. The hot lookup now indexes the two integer slices directly. A forced single-bucket test checks the three colliding entries across link growth, including the insertion immediately after reallocation. The hash tests include a comparison with the original native primitives.

With one Go execution thread (`-cpu=1`), five repetitions of 4,096 unique hash insertions, each followed by a commuted hit, fell from 388–394 µs to 333–341 µs and from 502,464 to 474,064 B/op. The single-thread i10 load improved from 421–431 µs to 378–382 µs. These are separate before/after measurements with the same thread setting. A fresh monotonic-wall-clock C run took 122.5 µs for the hash workload, 127.9 µs for i10, and 5.18 µs for the cone. The Go benchmarks report wall time too. Full GIA tests at the first cone-change snapshot passed in 491 s. Targeted hash tests passed at this snapshot.

The reference driver lives at `abcgo/aig/gia/test_vectors/perf_gia.c`. Build and run from the ABC directory on Linux with the default native library configuration:

```sh
make libabc.a
cc -O2 -DABC_USE_STDINT_H -Isrc -c abcgo/aig/gia/test_vectors/perf_gia.c -o /tmp/perf_gia.o
c++ /tmp/perf_gia.o libabc.a -lreadline -lpthread -lm -ldl -o /tmp/perf_gia
/tmp/perf_gia load i10.aig
/tmp/perf_gia hash
/tmp/perf_gia cone
cd abcgo
go test ./aig/gia -run '^$' -bench 'Benchmark(HashAndChain|AigerLoadI10)$' -cpu=1 -count=5 -benchmem
```

For the timings above, the driver was linked directly against the existing native object files with the executable entry point omitted.

### Count-only traversal on every i10 output

Size queries now use a stack of object IDs and mark each dependency when first queued. This avoids loading a shared object again to discover it was already visited. The ordered collectors still produce their original postorder. The count tests cover traversal marks, overlapping roots, typed gates, constants, and a deep graph with more pending siblings than the local stack can hold.

BenchmarkI10OutputCones sums cone sizes separately across all 224 outputs. The native driver (`perf_gia cones i10.aig`) counted 23,773 AND visits per sweep in 169.4 µs. At one Go execution thread, the old walk took 355–360 µs and 91 allocations per sweep. The new walk verified the same visit count in 121–123 µs with zero allocations/op reported. For the 1,024-AND chain, three Go cone timings were 4.58–4.59 µs against C's 5.18 µs. Go is ahead of C on both of these counting workloads. File loading and hash construction still trail C.

The module-wide run started at commit 450d13fd9 passed 130 packages, with three additional packages reporting no tests. It predates the single-pass reader and the new traversal.

### Satoko CDCL propagation

BenchmarkPigeonhole7Into6 constructs the 42-variable pigeonhole formula through the public clause API, solves it, checks UNSAT and destroys the solver. The C driver uses identical variable numbering, clause insertion order and polarity. Compile from the ABC directory with `cc -O2 -DABC_USE_STDINT_H -Isrc abcgo/sat/satoko/test_vectors/perf_pigeonhole.c src/sat/satoko/{solver,solver_api,cnf_reader}.c -lm -o /tmp/perf_pigeonhole` in bash. Run the Go equivalent with `go test ./sat/satoko -run '^$' -bench BenchmarkPigeonhole7Into6 -cpu=1 -count=3` from abcgo.

Go initially took 10.89 ms/op, about 1.28 MB/op and 9,366 allocations/op. Propagation accounted for 70% cumulative CPU. Hashing the (literal, clause ID) key for blocker lookups was the largest flat sample. C took 2.57 ms/op, with 57% flat CPU in solver_propagate in a separate gprof build. Porting C's watcher pair and binary prefix puts the blocker next to its clause ID and avoids scanning long clauses for binary implications. This also aligns marked-cone blocker checks and inspection accounting with the C propagation loop. A regression test moves one watch, then exercises the opposite watch with and without cone restriction.

Three Go runs now take 4.76–4.79 ms/op, 793 KB/op and 9,306 allocations/op. Before and after, Go took 695 conflicts and 892 decisions per solve. C took 710 conflicts and 875 decisions. Both return UNSAT, but their search traces differ. The race-enabled Satoko subtree and integration tests passed, including the optional native comparison of 900 formula/assumption cases. Investigation of the remaining search cost includes conflict analysis and restart timing.

### Match the native restart trace

Instrumented temporary builds of both solvers to record each decision and each learnt clause with its backtrack level and LBD (number of distinct decision levels). The first difference was event 469, after 204 learnt clauses. Go restarted while another conflict was waiting to propagate. C first finishes propagation, then checks the integer moving average against the global average. The revised search follows that ordering and uses the existing bounded queue. Database reduction likewise runs at a propagation fixed point and checks the conflict threshold before counting learnt clauses.

The revised Go and C traces compared byte-for-byte equal over all 1,584 events. Both solved the benchmark with 710 conflicts, 875 decisions and three search starts. The regression test checks those counters. Three Go runs took 4.49–4.52 ms/op; the differential and integration tests passed.

### Reuse Satoko's conflict-analysis memory

Ported the C timestamp array for counting distinct decision levels and performing binary resolution. Stamps grow to cover the current index domain, including duplicate assumptions that push decision depth beyond the number of variables. Explicitly clear on generation wrap. On its own this change reduced the seven-pigeon run to 4.01–4.02 ms and 8,035 allocations/op.

Learnt-clause minimization now uses Satoko's iterative reason traversal. Tentative marks roll back if a proof fails and remain available to later literals if it succeeds. The solver reuses scratch for learnt literals, the proof stack, tagged variables and last-level variables. Added regressions for mark rollback and retention, deep decision levels and stamp wrap. The race-enabled Satoko subtree, native differential and integration tests passed. Rechecking all 1,584 trace records still matched C exactly. Three final runs took 3.42–3.43 ms, 411 KB and 2,512 allocations per solve.

### End-to-end mapped SAT on i10

BenchmarkI10MappedSAT loads i10, coarsens GIA, maps cuts with an eight-input LUT limit, generates the output-disjunction CNF, initializes and simplifies Satoko, and solves. The matched native workload is `perf_gia sat i10.aig`. C took 21.26 ms and returned SAT with zero conflicts. Its flat profile put 37% in Mf_ObjMergeOrder and 27% in Abc_Tt6Isop. The initial Go benchmark took 171.7 ms/op and allocated 25.8 MB/op in 460,878 allocations. More than half the CPU samples were in the per-assignment truth expansion.

MF now uses the same truth expansion and minimum-base utilities as giaMf.c. The shared Go utilities swap low variables with word masks, swap high variables in blocks of words, and test dependence by comparing cofactors a word at a time. A random-table test checks swapped assignments through ten variables. The MF differential suite now compares the entire i10 CNF against native ABC, including clause and literal order. MF and utility package tests passed.

Three Go pipeline measurements took 56.5–57.1 ms/op, about 22.8 MB/op and 406,782–406,783 allocations/op. The updated allocation profile points to temporary truth tables and leaf arrays during candidate evaluation.

### Reduce candidate allocation in MF

The eight-input LUT limit bounds truth tables at four words. The mapper now expands the two or three input truths in local arrays, computes the result in the first array, and copies a truth into permanent storage only when interning a new function. A four-word map key replaces serialization to a byte string. Tests check that reusing a candidate buffer preserves older functions, including ones that differ only in the last word.

With truth scratch and reused exact-area traversal scratch, i10 took 48.5–48.8 ms and about 159,200 allocations. Moving leaf-array scratch into each candidate evaluation regressed allocations to about 328,735. The profile showed those arrays escaped. Reusing them across candidate pairs of the same node, preparing fanins once and retaining owned copies of candidate leaves reduced the final run to 43.8–44.2 ms, 11.7 MB and about 100,620 allocations. Utility and MF race tests passed, including full i10 native CNF parity.

The module-wide run started at the Satoko watcher checkpoint passed 130 packages, with three packages reporting no tests.

### Exercise the synthesis demo through its CLI

Built src/demo.c against the same native objects. On i10 the native resynthesis goes from 2,396 ANDs at depth 37 after the first balance to 1,851 at depth 35. C's final equivalence proof passed. One run took 0.13 s for rewriting and 0.22 s for verification. The original Go executable failed on its first read command. Linked the actual I/O and ABC command modules to frame initialization and shutdown. The BLIF dispatcher now converts an AIG or logic network to a netlist using ABC's conversion, and file-read AIGER networks inherit the source filename and default port names when needed.

Three regression tests failed against the preceding revision via a Go build overlay. They reproduced the unknown read command, failure to write a shared AND to BLIF, and missing AIGER network and port names. All three passed with the fixes, as did the root and base/io package race tests. The small NAND CLI test exercises the complete demo with real command handlers. The full module run at the preceding MF checkpoint passed 130 packages, with three packages having no tests.

Native CEC independently checked a BLIF after each Go i10 stage. Reading and both initial balancing passes preserved equivalence. The first rewrite -l changed output V1613(0) from one to zero on all-zero inputs. Subsequent passes retained errors. Go's CEC also panicked on a prover parameter type assertion for the nontrivial miter. These findings preclude a valid end-to-end timing comparison. A diagnostic Go profile across all synthesis stages with intermediate BLIF writes put 52% of CPU in Network.lookupAnd, which scans the object table on a hash miss.

### Check nontrivial miters through the CEC command path

A regression test compares the two associations of a three-input AND. Before the fix, its nonconstant structural miter panicked in Ivy proving when CEC passed a proof/fraig parameter record to the base/abci prover. CEC, sensitivity analysis and redundancy removal now use the command-facing parameter type and their original limits. Both proving entry points accept the concrete type, so the compiler checks every caller. The regression now reports equivalence. The exhaustive two-input prover tests passed alongside it.

### Preserve Ivy FRAIG representatives during refinement

The Go i10 prover reached an unmapped simulation representative. A six-input unit test exposed the underlying class-list error without running a SAT solver. After splitting a class into three pairs, its heads pointed to later members. The pattern scores were [1, 2] where successive source-order splits give [2, 1]. Both assertions failed before the fix. The Go refinement now partitions the linked list in place like ivyFraig.c, and iterates over the remaining nonmatching group. This also eliminates the map and grouped slices formerly created for every refinement. The complete Ivy package passed under the race detector.

### Repair incremental rewriting and measure the synthesis path

The zero-input regression on i10 failed before the rewrite fix. Two initial balances followed by rewrite -l changed V512 from true to false. C computes cuts from live fanins on demand, frees cached cuts when removing an AND from its hash table, and counts a hash hit in the current MFFC as a cost while still sharing the actual AND during reconstruction. The Go pass now follows those rules. The regression covers all 224 output bits. Two more red checks caught incorrect use of MarkA in prospective cone cost and unnecessary duplication of a marked AND. Fixing sharing also exposed a double polarity adjustment and an unordered merge in DSD restructuring. Its four-input OR depth regression failed at depth three and now passes at two.

Native CEC proved the final Go i10 BLIF equivalent to the input. Go produces 1,838 ANDs at depth 33 versus C's 1,851 at depth 35. Both Go and native CEC now report equivalence. With concurrent tests running, one complete Go demo took 27.82 seconds, including 9.31 seconds in rewriting and 18.31 seconds in verification. Eight native demo profiles put the largest flat samples in SAT propagation (21%), solver search (7%) and cut merging (5%).

BenchmarkI10ReadBalanceRewrite reads the input network, balances twice, runs the first rewrite -l, records node count and depth and releases the network. Two Go iterations at one CPU took 2.99 seconds/op, 930 MB/op and 937,243 allocations/op. The native driver in base/abci/test_vectors/perf_rewrite.c took 42.52 ms/op over 30 iterations. The Go and C stages finish at depth 36 with 2,004 and 2,039 ANDs respectively. These are equivalent synthesis tasks with different internal choices. Go's stage profile attributed 55% cumulative CPU time to lookupAnd and 93% of allocation bytes to initializing reference counts. These are the next performance targets.

### Remove whole-network scans during rewrite lookup and cone costing

Temporary MFFC reference counts now keep signed adjustments only for touched objects. The old implementation repeatedly copied every object's fanout count into a new map. A red regression also showed that the snapshot hid edges attached while an adjustment was active. It now returns the live count plus the adjustment. In the same read/balance/rewrite stage this change alone cut allocation from 930 MB to 69.1 MB and runtime from 2.99 s to 1.76 s.

On a structural hash miss the network now checks the shorter of the two input fanout lists. Every matching AND must appear in both. The test covers duplicate unhashed ANDs, opposite edge polarity, rewiring that changes list order, and a manager excluding the first match. Three final one-thread trials took 166–171 ms/op, with 2,004 ANDs at depth 36 throughout. This stage was 42.52 ms/op in C. Its native flat profile favored Rwr_NodeRewrite (18%), Abc_AigAndLookup (16%) and Dec_GraphToNetworkCount (11%). The complete Go demo reached 1.83 s in one run, with 0.64 s rewriting and 1.10 s verification, and proved equivalence. Focused race tests passed; the preceding live-cut snapshot also passed complete base/abc, base/abci and opt/rwr race tests.

### Avoid copying network tables to count live objects

Native ABC keeps separate live-object counts and an upper bound for object IDs. Go now reads its existing per-kind and terminal indexes directly for counts and indexed primary-port access. A sparse-ID simulation regression failed before the fix: the AND after a deleted ID produced zero on two true inputs. Verification now walks the ID range and advances the traversal epoch as in C. The new simulation and sparse-count tests passed under the race detector.

Three one-thread i10 read/balance/rewrite trials took 148–150 ms/op and 49.3 MB/op, down from 166–171 ms and 69.1 MB. All trials preserved the 2,004 ANDs at depth 36.

### Match cut enumeration order under a budget

With two cut slots, the Go enumerator returned [1,2,3] twice. A new regression reproduced the second cut coming from parents (0,1). The C loop enumerates each small right cut against all full left cuts first, which retains [1,2,4] from (1,0). The regression passed after restoring this loop order.

Walk the size-ordered fanin lists directly, use their signatures to rule out incompatible full cuts and track retained cut counts during dominance filtering. Bounded scratch lets oversize merges fail without allocating a leaf slice. Tests check the sorted union and independent ownership across all cut limits from 3 through 12, including modulo-31 signature collisions. Three i10 stage runs took 124.6–126.9 ms/op, 42.9 MB and 599,077 allocations, versus 148–150 ms, 49.3 MB and 859,246 allocations previously. The result still had 2,004 ANDs at depth 36. The earlier full module run covering the live-rewrite repairs passed all 130 tested packages, with three reporting no tests.

### Record the canonical form of each output polarity

The three-input AND regression for Cut_TruthNCanonicize initially returned 0x01010101 for both polarities and left the direct phase at zero. Native cutTruth.c canonizes the direct and complemented truth independently and stores the first minimizing input phase for each. Reusing the Go port of Extra_TruthCanonFastN gives direct 0x01010101 at phase seven and complement 0x7f7f7f7f at phase zero. The cut and extra package race tests passed. The preceding enumeration changes passed complete cut, rwr and abci race tests.

### Expand cut truth tables a word at a time

Cut truth computation now places each child's support using adjacent wordwise swaps, combines the two expanded functions, then replicates across unused variables. Tests check each assignment of randomly generated cuts through twelve variables, including compact inputs, all input/output phase flags and both AND and XOR. The shared adjacent-swap, stretch and shrink routines work directly in the caller's output storage. An in-place eight-variable round-trip regression initially measured 34 heap allocations, and now measures zero. Partially overlapping input/output slices and disjoint buffers also pass.

The native driver in opt/cut/test_vectors/perf_truth.c uses the same leaf placements, functions and complemented first edge as BenchmarkCutTruthMerge. A Go regression compares whole-table hashes against the native results for all three widths. Times are elapsed nanoseconds per merge.

| Variables | Previous Go | Wordwise Go | Native C |
| ---: | ---: | ---: | ---: |
| 4 | 480–498 | 60.9–61.0 | 41.8 |
| 7 | 3,771–3,989 | 152.2–152.4 | 120.9 |
| 12 | 294,756–299,156 | 2,662–2,834 | 2,796 |

The one-thread i10 read/balance/rewrite stage took 120.9–123.4 ms over three trials, with 2,004 ANDs and depth 36. Cut and extra package race tests passed, including the native small-function truth and phase fixtures.

### Score the rewrite manager's own decomposition graphs

The rewrite path previously rebuilt its canonicalization tables, decoded a second forest and constructed a second set of candidate graphs on the first node. A winning four-input XOR gave gain two while the initialized manager's winning node still had zero score, gain and added nodes. A separate regression showed the score report omitted candidate gains and cleanup erased the class count while retaining the candidate scores. Both tests failed on the previous implementation.

The evaluator now uses the manager's own class lists, graphs, permutation table and per-node scores. Score reporting and clearing follow rwrEva.c. The native class-order fixture checks all 222 classes against the actual manager, and an independent test decoder checks the candidate graph functions. Complete rewrite package race tests passed.

The i10 read/balance/rewrite stage reached 103.8–107.7 ms/op, 36.2 MB and 571,640 allocations, down from 120.9–123.4 ms and 42.9 MB. The complete Go demo finished in 1.37 s in one run, including 0.41 s rewriting and 0.89 s verification. Native CEC independently proved the exported BLIF equivalent to i10.

### Prove the same nontrivial equivalence problem in both ports

Added BenchmarkI10NativeResynthesisCEC and the matching base/abci/test_vectors/perf_cec.c. The inputs are the original i10 and the named AIGER produced by the native demo's synthesis commands. Generate the second file from the checkout root with `./abc -c 'read_aiger i10.aig; balance; balance; rewrite -l; rewrite -lz; balance; rewrite -lz; balance; write_aiger -s abcgo/base/abci/test_vectors/i10_native_demo.aig'`. Both drivers load the input pair outside the timer and build a fresh miter, check that structural hashing alone does not solve it, prove it under default CEC parameters and delete it each iteration.

Both programs built miters with 5,051 ANDs and proved equivalence. Native took 210.4 ms/op across 18 calls. Go took 596.8–603.5 ms/op, 165–166 MB/op and roughly four million allocations/op. In the native gprof run, SAT propagation accounted for 40% of samples and solver search for another 12%. In Go, propagation took 18% cumulatively, rewriting 33%, mapped CNF derivation 12% and refactoring 12%. Go sampled allocations included boxing ABC edges, cut enumeration and comparing generic vector entries.

### Label rewrite cones with live references on the nodes

The concrete rewrite adapter's cut-boundary test initially saw one reference on a pinned leaf (expected two) and failed to label the unpinned input when its last reference disappeared. Applying temporary fanout increments directly, advancing the traversal epoch and calling Abc_NodeMffcLabelAig reproduces the native labeling. A signed count adjustment on each object replaces the temporary map. The field uses space freed by grouping Boolean fields. A four-AND reconvergent cone regressed at two heap allocations on the previous revision and now labels without heap scratch.

The i10 read/balance/rewrite stage measured 90.7–92.6 ms, 31.7 MB and 504,270 allocations. The fixed-pair equivalence benchmark measured 560.4–569.2 ms and 151–152 MB across two trials, still proving the same 5,051-AND miter. Complete network and rewrite package race tests passed.

### Size bsat variable storage for the actual formula

The new bsat benchmark and native driver both solve seven pigeons in six holes with 604 conflicts, 788 decisions and 7,409 propagations. A regression fixes those native search counters. Before optimization, Go used 6.23 MB and 6,050 allocations per solve. Native averaged 2.10 ms in an initial 700-iteration profile dominated by propagation (62.5%). Go initially took 5.53–5.69 ms, spending 20% of sampled CPU in integer-key map lookups and 95% of allocated bytes on preallocating variable arrays.

A capacity regression failed because the 42-variable instance reserved space for 50,000 variables and 100,000 watch lists. Variable storage now grows geometrically from the requested size. The test checks repeated growth to 1,000 variables and preservation of an earlier implication. Allocation fell to 327 KB per solve. During high system load a sequential trial pinned both Go binaries and C to CPU 12 for 300 iterations each. Elapsed times were 7.83 ms (previous Go), 5.64 ms (new Go) and 2.70 ms (C); process CPU times per solve were about 7.73, 5.60 and 2.73 ms. The fixed i10 proof used about 140 MB in this version. Complete bsat, fracomb, fra and Ivy race tests passed.

### Look up bsat clauses by page and word offset

The solver now resolves a handle through its page and aligned offset. Learned-clause compaction validates every relocation before installing the surviving pointers. Two regressions failed on the previous revision. A handle with its high bit set decoded to a negative page, and a rejected relocation overwrote the first survivor's ID before reporting the second survivor's invalid ID. Both tests pass after the fixes. Further coverage checks stale handles after rollback and restart. The complete bsat and proof package race tests passed.

In consecutive 300-solve trials on CPU 12, Go fell from 5.23 to 3.26 ms per pigeonhole proof. Native took 2.17 ms with identical search counters. The direct index increased allocation from 327 KB to 438 KB per solve, with 6,050 allocations in both versions. The 5,051-AND i10 proof averaged 551 ms and 142 MB over four iterations.

### Reuse conflict and backtracking storage in bsat

The allocation profile attributed 55% of sampled objects to conflict analysis and 13% to backtracking. Conflict analysis now writes directly into the search's reusable learned clause and uses the solver's stack to minimize it. Backtracking restores the heap in reverse order from the propagated segment of the existing trail. Watch vectors acquire backing storage on their first insertion.

New allocation tests first measured four allocations per repeated conflict, one per backtrack and four reserved entries per unwatched literal. All pass after the changes. The initial trail implementation also panicked when an assumption contradicted a pending root unit. A one-variable regression reproduced the failure before the correction and verifies a subsequent unrestricted solve. Complete race tests passed for bsat, proof, fra, fracomb and Ivy.

The seven-pigeon solve now allocates 301,256 bytes in 1,967 objects, down from 438,112 bytes in 6,050 objects, with identical search counters. Under concurrent system load, alternating 500-solve wall times ranged from 8.52–9.79 ms before and 8.10–9.30 ms after. The timing ranges overlap. The fixed i10 proof averaged 132.5 MB and 3.74 million objects over five calls.

### Build four-input NPN classes by swapping cofactors

Building the canonicalization tables spent 85% of Go's sampled CPU time in Transform. Native's two permutation functions took 87% of flat samples. Go now permutes each representative once per column and computes its 16 input phases by swapping the cofactors of each permuted variable. Visiting the resulting table in ABC's phase and permutation order preserves its chosen witnesses. All four result arrays still belong to each caller.

Consecutive 300-iteration runs on CPU 12 measured 4.986 ms for previous Go, 0.835 ms for revised Go and 10.817 ms for the original uninstrumented C source compiled with GCC -O3. Both Go revisions allocate 327,680 bytes in four arrays. The native driver is misc/extra/npn/test_vectors/perf_npn.c. Exhaustive reconstruction, the native table digest and independent mutation of all output arrays passed. Race tests passed for npn, dec, rwr and rwt.

### Count mapped CNF references during traversal

The fixed i10 proof spent 7.4% of its sampled CPU time searching a vector for nodes already visited in its CNF mapping. Regression tests also found zero area for a seven-clause mapping, a complemented root in the collected gates, reversed DAR traversal order, stale reference counts and ignored supergate marks. The area-only scan allocated and filled a vector that C leaves null. The tests failed before changing the scan.

The traversal now counts incoming mapped references on each regular object, collects a gate only on its first visit and sums the actual cut costs. It handles the native DAR preorder, both transferred-cut orders, marked supergates and recursive calls that continue an existing traversal. Top-level CNF derivations restore structural fanout counts before returning. Complete cnf, pdr, bmc and fracomb race tests passed. The preceding whole-module checkpoint passed all 130 tested packages, with three packages reporting no tests.

The 1,024-gate binary chain in sat/cnf/test_vectors/perf_scan.c has 3,072 clauses of mapped area. Go's scan fell from 14.13 ms and 524,766 allocations to 89.2 µs and 964 allocations. The native driver took 15.1 µs. Consecutive eight-proof runs pinned to CPU 12 measured 521.8 ms and 3.73 million allocations before the change and 474.3 ms and 2.27 million allocations after. Both proved the 5,051-AND i10 miter.

### Restore the alternate mapped CNF used by PDR

The public alternate derivation produced six variables, eight clauses and an empty mapping for a complemented three-input AND. Direct calls to original C with both truth-minimization settings produced seven variables, seven clauses, fifteen literals and a twelve-entry cut map. A regression checks the exact clause order, object ranges and mapping. A separate red test showed that the alternate writer ignored a supplied five-input ISOP. Both writers now share clause emission, and both alternate derivation entry points select DAR cuts using the requested minimization setting. An explicit inverting buffer had returned nil in both mapped encodings. Its new one-leaf cut preserves its polarity.

The mapped format exposed a PDR model-preparation panic on an internal AND absorbed by a cut. A focused test reproduced the panic. Dynamic clauses now use AIG object IDs directly. When a model consumer requests an absorbed physical gate, the frame solver installs its definition so the object can be read after solving. Tests cover absorbed buffers, ANDs and XORs on every input assignment. Complete CNF, PDR, abstraction and bounded-model-checking race tests passed, including the PDR differential against local ABC.

BenchmarkPairedShiftRegisters and proof/pdr/test_vectors/perf_shift.c construct two twelve-stage shift registers driven by a shared unconstrained input. Both prove the final stages always agree through frame 12. Sequential 150-proof runs on CPU 12 took 9.04 ms for the previous Go encoding, 8.81 ms for the restored mapping and 18.29 ms for C. The mapped Go path used 2.52 MB and 8,373 allocations per proof. Native spent 51% of samples loading SOP covers and 19% in SAT search. Go spent 28% in the SAT decision random-number generator, and 77% of its sampled allocation bytes were in variable-array growth.

### Reserve PDR frame variables together

Building a PDR frame solver allocated 118 objects in a new regression. Reserving all 500 initial variables in one call passes the allocation bound and preserves an existing larger solver's SAT model. For the default integral seed, BSAT now reduces the exact integer product modulo 2³¹−1 with a fold. Tests compare the state and result bit for bit with the original recurrence over 10,000 steps per seed, including fractional and signed-zero cases. The standalone generator took 7.14 ns per draw, down from 10.85 ns.

Consecutive 350-proof runs pinned to CPU 12 took 8.73 ms and 2.50 MB per paired-register proof before the changes and 7.38 ms and 1.40 MB after. Both reached frame 12. The seven-pigeon solve measured 3.04 and 3.07 ms respectively, with 604 conflicts, 788 decisions and 7,409 propagations in both runs. Complete BSAT and PDR race tests passed.

### Expand four-input DAR cut truths by support placement

A native driver and a Go test build the same 768-gate layered graph. The test compares a digest of every active slot, its physical position, leaves, value and truth against the C implementation with support minimization both on and off. A separate regression found that both whole-graph Go entry points retained an unused two-gate cone. Native removes it before counting references and scoring cuts. Both Go entry points now do the same.

On the baseline, the Go CPU profile attributed 33% of samples to merging cuts, including sorting and allocation. A rejected union allocated four objects in a red allocation test. Bounded insertion now rejects it without allocating. A small table expands compact functions onto each subset of four variables, and support minimization compares word-sized cofactors. Tests cover every truth function at every four-variable placement and exhaustive support reduction.

Across 200 passes pinned to CPU 12, the public physical-slot benchmark fell from 1,193 to 875 µs with minimization and from 4,109 to 2,184 µs with full support. The original C driver took 142 and 596 µs respectively. C retains the physical slots in node-owned storage. The public Go benchmark includes a separate mutable copy for the caller. Native sampling attributed 43% to computing node cuts, 29% to dominance checks and 21% to ordered merge. DAR, CNF, FRA and PDR race tests passed. The previous alternate-CNF checkpoint passed all 130 tested packages in the complete module.

### Store DAR cut boundaries inline and stream them to CNF

Each cached slot now holds four leaves and a signature. A bounded pool allocates slots as needed. Merging sorted boundaries also determines both truth-table placements. CNF evaluates candidates by value and copies the chosen cut once. Callers of Slots and Cuts retain independently editable slices. A cached-root regression initially measured ten and fourteen allocations for those APIs. Both now need at most two, including their leaf storage.

With 300 passes, the 768-gate caller-owned enumeration reached 500 µs with minimization and 985 µs with full support, at 1,634 allocations in either case. For the node-owned operation, 500 passes on CPU 11 took 343 and 869 µs in Go, versus 144 and 605 µs in C. Go allocated 29 objects per pass. Ten fixed i10 proofs on CPU 12 averaged 437 ms and 2.01 million allocations, compared with 454 ms and 2.27 million allocations on the pre-DAR comparison build.

Two tests exposed incorrect area-flow choices. Replacing an earlier AIG node with a gate that depends on a later object made CNF read a zero area for that fanin. The mapper now visits fanins first. Already topologically numbered graphs need only the returned slice, down from 23 allocations in the regression. A second circuit tied at cost 40 on two cuts with different reference sums. Original C truncated the sums to its 11-bit field, choosing the two-input cut with value four over the three-input cut with value two. Go previously chose the latter. Both regressions passed after the fixes. Full race tests passed for AIG, SAIG, DAR, CSW, CNF, SSW, abstraction, FRA and PDR.

### Preserve clause identity when loading SAT frames

Go used to insert all frame clauses before linking the registers and reversed the order of the two implications. A red proof-store test recorded different root clause IDs. The loader now reserves variables together, inserts latch equalities in native order before each new frame, and uses one temporary buffer to lift clauses. The test checks all six proof roots and the original CNF after loading. A second red test found that Cnf_DataWriteIntoSolver2 returned the regular solver. It now returns the proof-logging Solver2, and an unsatisfiable four-clause formula yields a four-root resolution core.

The CNF allocator had lost its graph pointer and both declared dimensions, and duplicate flip offsets were applied only once. Regression tests reproduced all three discrepancies. Clause copies now share a bounded backing array across clauses in each copy. Each clause has its own capacity boundary and both public and private duplicates retain their independent data. Packed import of 8,192 ternary clauses fell from 277 to 70 µs and from 8,193 allocations to two.

The paired native driver is sat/cnf/test_vectors/perf_loading.c. Across 500 single-core runs, full Go data duplication took 269 µs and five allocations, down from 590 µs and 16,387 allocations. Native took 8.45 µs and makes one literal copy. Loading and disposing a regular solver took 1.75 ms in Go, down from 2.07 ms, compared with 1.17 ms in C. The Go profile attributed 54% of sampled allocated objects to BSAT clause storage and 27% to copying input clauses for normalization. Native sampling put 36% of flat samples in Sat_MemAppend. CNF, BSAT, ACB, ABCI, PDR, BMC, abstraction and FRA race tests passed.

### Store short SAT clauses with their literal arrays

Normalizing a repeated 66-literal input allocated a temporary slice on every call. The solver now reuses its private normalization vector as in C. A separate test measured 1,043 allocations for 512 short clauses. Both regular and learned streams now allocate bounded blocks of short records containing the literals and optional ID. The block size grows with actual short-clause use. Tests exercise independent input and neighbor storage, retained clause pointers, IDs and restarts. A high-bit watch-tag test also exposed negative decoded literals for valid Go variables at and above 2²⁹. An unsigned shift now preserves all encodable bits.

In 750 loads of the 8,192-clause workload on CPU 12, Go averaged 1.45 ms and 6,220 allocations, down from 1.75 ms and 30,727 allocations. The seven-pigeon solve took 2.95 ms and 300,432 bytes in 1,814 allocations, with the same native search counters. An earlier block-sizing trial used 308,944 bytes because long learned clauses advanced its growth estimate. BSAT, CNF, BMC and PDR race tests passed.

### Count minterms across sparse and wide BDDs

The public Go counting routines returned zero for a function of variables 6 and 7 when asked for its count over two variables. Native CUDD returned one, and returned 0.25 when asked for the count over zero variables. Both ordinary and extended Go counters now scale by the requested variable count independently of the indices. A second red test found that the core lost the sole assignment of a complemented 80-variable cube. Following complemented cofactors directly preserves that assignment. For widths beyond the double exponent range, the extended calculation also recovers a single assignment among 1,200 variables. The ADD background setter now accepts manager-owned constants. With infinity as background, ordinary ADD zero counts as 16 at width four in both C and Go.

The paired queens fixtures construct row-major BDDs using the same operation order and reference lifetimes. Five queens gives ten solutions and 167 DAG nodes. Six queens gives four and 130. The six-queen baseline took 14.66 ms, 8.18 MB and 20,402 allocations per solve in Go across 80 runs. C took 2.60 ms across 100 runs. The Go CPU profile attributed 86% cumulatively to the general ITE recursion, and its computed map accounted for 63% of allocated bytes. Native sampling put 42% of flat time in unique-node lookup and 17% in the dedicated AND recursion. All 13 BDD packages passed under the race detector after the counting fixes. The preceding complete module run passed all 130 tested packages.

### Use CUDD's two-operand conjunction recursion

AND now handles terminals and complemented cofactors directly, orders the operands for commutative cache reuse and consults the computed table when either operand's regular reference count differs from one. OR uses the same recursion through De Morgan's law. The context-aware calls retain cancellation checks inside the recursion. A regression for swapped AND followed by its dual OR initially took eight lookups and one hit. It now takes two lookups and two hits.

Cache hits for BDD ITE, ADD ITE and generic node operations previously returned released nodes with too few child references. Tests now check that each hit restores the children while leaving the returned root's reference count at zero. A constant-only lookup preserves its native read-only behavior. Separate red tests found leaked references after failed bounded AND and ISOP operations. Three retries raised two surviving nodes' reference counts by three each. The two operations now undo the references acquired by new and revived nodes.

The six-queen Go workload took 5.95 ms and 3.26 MB in 180 pinned runs, down from 14.66 ms and 8.18 MB. All 13 BDD packages passed under the race detector, including exhaustive three-variable fused abstraction and reorder tests. The updated profile attributed 68% of CPU time and 89% of allocated bytes to canonical node lookup and construction.

### Count incoming decision-diagram references in one pass

Checking external ownership after releasing the six-queen root scanned every canonical node once for each node being checked. The original Go profile placed 99.7% of samples in the per-node reference query. The whole-manager check now counts incoming arcs once across BDD, ADD and ZDD nodes. Garbage collection uses the same count, and BDD reordering retains each external root's multiplicity across trial permutations. A red test also reported three saturated nodes as leaking. The check now follows CUDD's rule for saturated reference counts.

Three checks averaged 1.084 s each before the change. Afterward, 3,000 checks averaged 223 µs with zero allocated bytes on the released graph. Native averaged 49.6 µs in 6,000 checks. Both returned zero outstanding roots. The native queens driver accepts a final refs argument to time the check on a single constructed and released graph.

A regular handle and its complement now share one Go allocation. The allocation regression originally measured 1,085 allocations to create 512 projection pairs. The full queens build now uses 10,159 allocations and 3.10 MB, down from 20,158 and 3.26 MB. Its 6.08 ms runtime was close to the previous 5.95 ms measurement. All 13 BDD packages passed under the race detector.

### Honor CUDD table sizes and memory hints

Four native initialization cases exposed fixed defaults in the Go constructor. With three BDD and two ZDD variables, a unique-table request of nine and a cache request of 70, C reports 96 unique slots and 64 cache slots. Go reported 1,536 and 262,144. The constructors now round the unique subtable size up and the computed cache size down using the native minimums. A 1 MiB memory hint yields a hard cache threshold of 8,738 entries and a loose unique-table threshold of 5,242 on the 64-bit build. Resetting either threshold uses ABC's default 64 MiB data limit.

Tests also reproduced wrong slot counts when growing variables through the public APIs. Inserting a BDD or ADD variable at the top now moves the previous levels' capacities along with their variables. Aligning six ZDD variables updates the inventory to 160 slots. Growing an empty eight-slot unique table to ten variables doubles the cache once, to eight slots. Both growth cases were compared with native CUDD. All 13 BDD packages and misc/extra passed their race suites in the sizing run.

### Share bounded computed slots across CUDD operations

The old caches let 64 insertions occupy 64 entries in a four-slot manager. The ZDD and constant-only lookups also missed entries inserted by the matching operation. A public flush retained BDD results, a shuffle retained generic results from the old variable order, and a closed manager still returned cached nodes. Each failure was reproduced in a unit test before the port.

BDD conjunction, BDD/ADD ITE and external operators now use the same direct-mapped slot space, with exact operand and operation checks on a hit. Sparse storage allocates only touched slots. The native miss threshold can double the slot count and migrate valid entries while retaining lifetime hit and lookup counters. Flushing invalidates values, drops their Go pointers and preserves the used-slot statistic. Tests check collisions, actual returned values, a four-to-eight-slot growth after ten hits, live entries across resizing and cleanup on quit.

The six-queen graph ended with 3,092 valid entries in 16,384 slots. A first version that retained a copy of each hash took 5.73 ms and 3.36 MB. Recomputing hashes only at resize and releasing manager maps on quit took 5.39 ms and 3.29 MB across 400 runs. The previous unbounded implementation took 6.08 ms and 3.10 MB. All BDD packages and misc/extra passed under the race detector, and both CUDD packages passed again after compacting the entries.

### Count exact minterms using CUDD digits

With infinity as the ADD background, Go's exact counter returned zero for ADD zero and for a variable choosing between those two values. The native terminal count at width four is 16 and the choice has eight assignments. The extended counter also lost the complement of the designated background when that background was BDD one. The extended regression follows the complement calculation in cuddUtil.c. The native debug build reached an assertion in EpdSubtract3 on this particular input, so its output was unavailable for comparison.

The exact count now memoizes one base-2^32 value per regular node. A path makes at most one decision per manager variable, so entire low digits that must stay zero can be restored at the root. The tests compare all 256 three-variable functions against independently scaled truth-table populations at 11 widths. Comparing APA digit slices also dropped from ten allocations to zero in a red allocation regression.

For five queens scaled to 4,121 variables, Go took 23.2 µs and 19,784 bytes per count, down from 137 µs and 316,360 bytes. The checked native driver took 33.5 µs over 24,000 counts, with 65% of its samples in digit shifting and 24% in addition. For six queens at 4,096 variables Go took 28.6 µs and 21,208 bytes, down from 171 µs and 247,624 bytes. The Go profile puts 52% of cumulative samples in memo table growth. In the native six-queen case, two 32-bit operands in Cudd_ApaAdd wrap before widening. At width 36 the native answer was hexadecimal 70000004 for the known four solutions. The Go benchmark checks the full exact result against four shifted to the requested width. All 13 BDD packages and misc/extra passed under the race detector.

### Match CUDD live-node counts and manager pins

A mixed manager with three BDD and two ZDD variables reported nine application BDD/ADD nodes in Go on initialization. Native CUDD reports one. Its seven BDD/ADD keys include isolated projections and unused constants. The ZDD live count starts at four, including both terminals. Go now maintains the dead BDD/ADD key count as canonical nodes release or reclaim their children, and the public counters distinguish these inventories. Nine native snapshots cover creating, releasing, reclaiming and collecting both a conjunction and a ZDD singleton. The mixed manager's garbage collection threshold is 384, up from the erroneous reported value of one.

Red tests also showed that releasing a manager-pinned projection removed two references to one, and that exact lookup of a previously released ADD constant did not revive it. Both paths now retain the correct references. The peak count survives both dereference and collection when the peak was never explicitly queried. The raw node allocator previously admitted ten live keys under a nine-key limit and rejected ten live keys under a ten-key limit after fifteen conjunctions had died. Both cases are now covered.

The six-queen construction took 5.71 ms with 10,141 allocations after correcting the pin lifetimes and accounting. Checking the released graph took 237 µs and zero allocated bytes. A separate three-run experiment moved the child pointers first in the Go map key. Its 5.45–5.65 ms times and identical allocation counts did not justify the layout change on their own.

### Partition canonical decision nodes by variable

Three table implementations ran against the same six-queen construction. Partitioning Go maps by variable reduced allocation to 2.74 MB and took 5.30–5.37 ms. Per-variable arrays with linear probing took 4.48–4.56 ms and 3.07 MB. Reusing the location from a lookup when inserting a missing node took 4.27–4.35 ms in three runs of 500. The final 600-run check during the full-module build took 4.92 ms. The preceding global map took 5.71 ms and 3.29 MB. Native took 2.60 ms on the original paired baseline. The Go profile still places 62% of cumulative CPU time in canonical node lookup and construction.

The selected tables store one key inline for an isolated projection, retain the original child identities during graph rewiring, and allow deletion during iteration. Tests force 96 distinct keys to share a hash and exercise 1,280 canonical functions through growth, partial collection and reinsertion. A profiling run exposed two new allocations in the reference audit. A red allocation test now covers this path. Passing the incoming-arc map from the caller restored zero heap allocations, with the six-queen audit taking 205 µs.

Releasing the three-variable ZDD universe left five live nodes in Go. Native releases three decisions and keeps its two terminals. Both implementations now reproduce the six native free, rebuild and external-ownership snapshots. Additional red tests exposed two out-of-range reads during the exported Go free/rebuild lifecycle. Reading an absent level now returns nil, and requesting a projection can rebuild its universe. The prior full-module race run completed with all 130 tested packages passing, including GIA in 744 seconds.

### Pack CUDD cache records and count projection references

In C, AND on two isolated projections performs no cache lookups. With an additional owner for each projection it performs one lookup per call and hits on the second and third calls. Go produced the opposite results in all six cases. The cache policy now accounts for the manager's implicit reference. A separate red diagnostic test found that the consistency checker accepted a missing or misdirected inverse on both a projection and the one terminal. It now checks canonical pairs. A regular and complemented handle share one reciprocal pointer and a phase flag, saving one pointer per handle.

Correcting the cache policy increased the number of occupied slots at the end of the six-queen solve from 3,092 to 7,061 out of 16,384. Storing those records directly in a Go map took 5.42–5.51 ms and 3.76 MB. A fully dense array after promotion took 4.94–5.15 ms and 3.70 MB. The selected layout stores the first four records inline, then uses blocks of 64. Its slot index starts as a sparse map of integers and becomes a direct integer array when enough slots have been used. The comparison trial took 3.74–3.88 ms and 2.85 MB. Tests verify all 160 entries across index promotion, flush and rehash, including removal of invalidated node references.

In a fresh back-to-back run on CPU 12, C took 1.68 ms over 500 solves and profiled Go took 4.64 ms over 800. Both returned four solutions and 130 DAG nodes. Native flat samples attributed 39% to unique-node interning and 28% to conjunction recursion. Go attributed 59% cumulatively to canonical-node lookup and construction, including 22% in handle allocation. The earlier C baseline was 2.60 ms. All 13 BDD packages, misc/extra and base/abci passed the race detector at this checkpoint.

### Allocate mapped CNF clauses in two flat literal arrays

Both drivers wrote 2,051 SAT variables, 3,074 clauses and 7,170 literals for 1,024 binary ANDs. Native took 82.7 µs per write over 8,000 iterations. Its largest flat costs were SOP conversion (43%) and counting literals (31%). Go initially took 368 µs, 571,928 bytes and 6,172 allocations. A 128-gate regression measured 792 allocations for the regular writer and 793 for the object-numbered writer.

The writers now count first, reserve both clause tables and emit capacity-bounded slices of contiguous literals. A single bulk copy preserves the private version of each clause. A byte lookup decodes each ternary SOP cube. Two covering tests exercise every input byte and all 16 signed literal-pair positions. Eliminating per-clause allocations took the benchmark to 193–194 µs. Building both clause tables during emission took 173–175 µs, with 34.9% of CPU samples still in garbage collection's object scan. The final compact-truth version took 182–196 µs, 286,568 bytes and 13 allocations across three 6,000-iteration runs.

A red SAT test found that compact two-input truth 1 rejected its true row, inputs 0 and 0. Both writers, the mapping export and the small-cut lookup now repeat the significant truth bits over unused inputs. The SAT test checks all 16 functions against every assignment in both output formats. Another regression reproduced the loss of an explicitly assigned cut on a fresh manager and after its graph grew. Cut storage now grows for valid new object IDs.

Native area-flow measurements also exposed two scoring errors. A two-input AND with an internal leaf of area 83 and two fanouts scored 71 in C and 44 in Go. Empty and two-input supergates scored 100 and 341 in C versus 0 and 41 in Go. The Go helpers now return the native scores. The main DAR mapping path already used the SOP score and 11-bit fanout tie break.

All nine affected packages in the final CNF race run passed, including both writers, PDR, BMC, abstraction, DAR and the ABC integration package.

### Count rewrite candidates through the owning AIG hash

A fresh i10 proof profile attributed 23% of Go's allocated objects to boxing complemented decomposition operands. A miss on an AND of two inverted inputs allocated two objects. Passing phases into a typed lookup eliminates both. A generic adapter still runs the original counting path. Tests compare both paths on hits, misses, both input phases, root reuse, constant and operand reductions, deletion-cone membership and both budgets.

Regressions reproduced a deleted hash entry reappearing in decomposition lookup, construction and balancing. These operations now consult the owning AIG manager. A second test showed that a nontrivial pair still returned an AND after either operand's effective fanout count reached zero. The count now follows the native check after trivial reductions. An Ivy rewrite with two missing intermediate ANDs panicked when a typed nil passed as an interface value reached the next lookup. It now counts two new nodes and rejects a one-node budget. Both the balance search and the complemented decomposition miss have zero-allocation regressions.

Both languages proved the same 5,051-AND miter equivalent. C took 221.6 ms per proof over 14 iterations, with 37.5% of flat samples in SAT propagation and 9.4% in solver search. Go initially took 438–452 ms, 120 MB and 1.95–1.97 million allocations. The phased lookup trial took 396–402 ms, 109 MB and 1.27–1.30 million allocations. After updating effective fanout checks, stable sorting and scratch handling in balancing, two seven-iteration runs took 408–411 ms, 107–108 MB and 1.22–1.24 million allocations. Race tests passed for the six affected decomposition, ABC, rewrite and Ivy packages.

### Evaluate only the selected refactoring cone

The i10 profile attributed 6.8% of flat CPU samples to saving, clearing and restoring every object mark for each local truth table. The new paired cone benchmark has three gates in a network with 4,096 unrelated inputs. Both drivers require the same 64-bit XOR truth and three visited gates. Go took 13.9 µs and allocated a 4,864-byte mark array per call. Native took 36.2 ns across 20 million calls, with 54% of samples in truth evaluation and 42% in collection.

A local function table now serves as the visit set, preserving caller marks and the traversal epoch. A two-pass version took 180 ns. Evaluating each gate on its first depth-first visit took 129–130 ns with zero allocations. A two-word regression checks all 128 assignments, node order, shared fanins and a preexisting mark.

Separate red tests exposed incorrect refactoring gains and deletion-cone marks. A constant replacement counted two removable gates with a cut leaf held, although all three ANDs disappear. A nonconstant replacement marked pinned inputs as deletable, advanced the traversal epoch twice and changed the caller's marks. Refactoring now uses the full MFFC size for a constant and the native temporary leaf references plus one labeling traversal for a nonconstant. The standalone ABC helper and all six affected packages passed their race tests.

The complete i10 equivalence proof took 357–363 ms with two-pass cone evaluation. The final single-pass code took 363–376 ms and about 99 MB during the full-module test build. Each proof still checked the 5,051-AND miter. The preceding version took 408–411 ms and 107–108 MB. The native proof measurement for this pair was 221.6 ms.

### Check the i10 network and its ordered object lists

The fresh i10 check benchmark validates 2,675 ANDs per iteration, with file checks enabled in C. Native took 382 µs per check over 3,000 iterations. Its largest flat costs were checking each object's connections (26%) and traversing for cycles (13%). Go initially took 4.70 ms, 2.32 MB and 673 allocations. Grouping expected destinations by source allowed one exact comparison of every reciprocal pin, bringing this to 1.13 ms. Reusing that scratch for a cycle check and indexing the object lists by ID brought it to 446–448 µs. Cycle failures still report the original DFS back edge.

The final checker verifies all per-kind and combined lists in one pass, compares CI and CO names in a shared table, and enumerates stored name IDs without sorting. The AIG check recognizes a cached AND after checking that the pair is irreducible and both inputs have effective fanouts. Three 3,000-iteration runs took 314–322 µs, 134,335 bytes and seven allocations. The full i10 proof took 341–348 ms and 83 MB for the 5,051-AND miter, versus 363–376 ms and 99 MB before this change. C previously took 222 ms for the proof.

Red tests reproduced a duplicate object occupying two ID slots, a removed AND accepted by the whole-network check, missing or duplicate gate entries, and malformed box lists that could pass validation or panic. A separate red test caught the standalone latch checker dereferencing a nil object. The retained tests also cover exchanged reciprocal pins with equal degrees, a high-degree gate with duplicate pins, stale names, direct CI-to-CO namesakes and cached reducible ANDs. All nine affected packages passed their race tests. The full module run started at the preceding refactoring checkpoint passed all 130 tested packages, including GIA in 465 seconds.

### Retain DAR choices and allocate only populated cut slots

The paired 1,024-AND chain now measures full CNF derivation with the same 1,710 clauses and 4,442 literals in C and Go. Native took 446 µs over 1,000 derivations, with 38% of CPU samples computing DAR cuts and 16% scoring area flow. Go initially took 1.19–1.20 ms, 1.39 MB and 3,458 allocations. Half the allocated bytes reserved DAR slots, including ten slots for every input.

A red lookup test found that the prepared manager always returned a nil DAR choice. Another showed that replacing a CNF cut changed the supposedly original DAR scan to area five and kept an absorbed AND referenced. The manager now retains the original chosen leaves, truth, area flow and 11-bit tie value. DAR scanning and cut creation consult that choice. Explicit transfer copies referenced choices and drops absorbed nodes. Tests cover distinct managers with overlapping IDs, mutable CNF cuts, complemented lookups, zero-leaf contradictions and freeing one manager independently.

Packing initial CNF cuts into per-gate records brought allocation down to 387 objects. DAR now reserves slots after visiting both children and keeps only the prefix touched by the merge. Holes inside it retain their original physical positions and values. The public physical-slot API still returns the requested number. A 64-slot test originally cached 64 records for both an input and a two-input AND. It now caches one and two respectively. Both native 768-gate slot digests still match.

Complete chain derivation now takes 934–938 µs, 847,120 bytes and 372 allocations. On the layered graph, node-owned enumeration took 255–330 µs with support minimization and 647–661 µs with full support. The preceding node-owned Go measurements were 343 and 869 µs, and native took 144 and 605 µs. Eight affected packages passed their race tests on the combined changes.

### Validate compact AIGs through their existing hash tables

After compacting the DAR slots, 16% of the full chain derivation profile was in compact AIG validation. On the same 1,024-AND chain, the Go checker took 150–151 µs, 75,304 bytes and 22 allocations. Native Aig_ManCheck took 14.5 µs across 180,000 checks, with 39% of flat samples in table lookup and 27% in hashing. The Go check also audits fanout counts, cached levels and acyclicity.

The Go checker now verifies the identity of each gate in its owning table and compares the gate counts with the table sizes. This detects duplicate and stale entries without building another pair of hash tables. Strictly increasing IDs across all structural edges prove acyclicity on the ordinary path. Graphs containing backward edges still receive a full cycle check. Three 8,000-pass measurements took 48.6–49.9 µs, 20,736 bytes and two allocations. Whole chain CNF derivation took 806–846 µs with 352 allocations.

Red tests reproduced malformed PI pins, an extra PO pin, missing and duplicate interfaces, a deleted constant, and an AND stored with both phases of one regular fanin. All now fail validation. A separate test showed that reordering legal repeated raw outputs rejected an exact permutation. The reorder check now verifies multiplicities, preserving complemented and repeated raw drivers. All eleven affected packages in the race run passed.

### Fold singly referenced CNF cuts and verify the composed truth

Five red postprocessing cases exposed the Go implementation's cost and reference rules. It compared a folded cut with the root cost alone, tried leaves in their stored order and could fold an unreferenced root. It now visits eligible leaves in stable cost order, permits a fold up to the combined cost of both cuts, rejects saturated covers and updates references while removing the absorbed cut. Tests compare the resulting counts with a fresh mapped scan and check that each root absorbs at most one leaf.

A red four-input test found that truth 0x0018 cost seven clauses instead of the native minimum of six. Cut composition now uses the precomputed SOP table for small results. The regression checks all 65,536 functions. Another red test exposed reversed arguments in the support-placement helper. A separate eight-variable test checks 256 rows with overlapping supports and a substituted variable in the middle of the top truth table.

On the three-leaf and seven-leaf AND compositions, Go initially took 1.97–2.04 µs and 7.93–8.99 µs respectively. Compiling the input row toggles once per union and using small SOP lookup brought this to 144–148 ns with three allocations and 3.71–3.72 µs with 62 allocations. The seven-leaf case spends most of its remaining work deriving the two ISOP covers.

The original C driver took 65 ns and 877 ns but produced different Boolean functions. Its three-leaf truth was 0x14 in the low byte, differing from the expected AND on three of eight rows. Its seven-leaf truth differed on eight of 128 rows. In cnfCut.c, the native routine decrements the leaf count before calling Cnf_CutTruth, whose address calculation uses that count. A temporary one-word pointer correction changed both results and still failed to recover the correct tables. The Go regressions check every assignment against the AND definition. The native measurement driver reports its mismatches explicitly. This optional C routine is outside the standard CNF derivation path. Nine affected packages passed their race tests.

### Restore scalar ISOP recursion and reuse its private store

The native reference forms 256 pseudo-random five-variable covers per batch, trying both polarities on alternating functions. Both languages produce digest aa1cc4a627bf6337 over the selected polarity, cube count and every cube in order. C took 312 µs per batch over 9,000 batches. Its profile attributed 64% of samples to scalar recursion and 27% to the two cofactor operations. The original Go code delegated each scalar recursive call to the slice-based implementation. It took 1.49–1.50 ms, 608 KB and 37,024 allocations per batch.

A red test showed that the scalar entry point exhausted its store for a zero cover with zero words available, a constant cube with one word available and a single literal with two words available. All three fit in the native KIT implementation. Another case with seven declared variables and one essential variable verifies that the truth fills all four output words. Scalar recursion now holds cofactors and intermediate truth values in registers. The wide routine reuses its reserved output words for temporary cofactors, and both paths append cubes in native order.

The scalar implementation alone took 492–496 µs and 1,662 allocations per batch. Reusing the private recursion store across calls brought that to 234–236 µs, 96 bytes and two allocations per 256 functions. Stores larger than 256 KiB are released. Caller-owned results survive reuse, including concurrent cover and incompletely specified truth-table requests. Seven-leaf CNF composition fell from 3.72 µs and 62 allocations to 1.68–1.74 µs and 12 allocations. All eleven affected packages passed their race tests.

### Accept existing unary AIG buffer records

The full module suite exposed failures in hierarchy flattening and SBD observability after the AIG checker change. Both callers use a buffer with a single driver and Go's zero value in the unused second pin. A new unit test failed on construction, reconnection and a historical raw buffer record. The checker rejected all three and the second-fanin accessor reported the constant object as a driver. Buffer construction and reconnection now store the absent-edge sentinel, and accessors recognize existing unary records. The new red test and both affected integration tests passed after the fix. All four affected packages passed under the race detector.

The subsequent full module run passed all 130 packages with tests after the buffer fix.

### Report native CUDD occupied buckets and per-level density

A four-stage native fixture grows BDD, ZDD and numeric constant tables independently. The final native manager has 192 buckets, of which 69 are occupied. Go originally reported 214 occupied buckets out of 152 and returned the same value for theoretical occupancy. The regression failed at every stage. Go now counts nonempty native hash buckets using the ABC build's node identity and complement-bit hash. BDD and ADD keys share a logical level, ZDD keys use separate levels and numeric constants use their IEEE-754 words. The expected fraction uses each level's key count and width separately. Explicitly dereferenced keys remain in both diagnostics until collection.

Additional coverage exercises the native pre-insertion density threshold, an existing-key hit, a permuted mixed BDD/ADD level and a sparse table with two billion logical buckets. A red mixed-family test showed the shared subtable staying at eight slots when a new ADD key should have grown it to sixteen. Capacity changes update the computed-cache limits. Go's underlying open-addressed storage still grows according to its own probe density. The native decision to collect under dead-node pressure needs a separate adaptation for Go's unpinned intermediate results.

The C driver took 222 ns per actual/expected query pair. The Go reproduction took 1.81–2.18 µs and three allocations initially. Small stack bitmaps brought this to 1.58–1.85 µs and zero allocations. The Go profile attributed 40% cumulatively to marking occupied buckets, 23% to iteration over the constant map and 13% to exponentiation. Six-queens construction took 3.86–4.16 ms across three samples. All thirteen BDD package race suites passed.

### Build variable interactions once per BDD or ADD root

The old interaction builder walked every BDD node as a separate root, built two maps per search and enumerated the support pairs repeatedly for each nested cone. It also omitted ADD roots. A red regression identified eight missing pairs across two ADDs with a shared tail. A separate red test added four variables to a manager and observed existing pairs (2,4) and (9,14) reappear at (1,12) and (6,8). The old packed matrix interpreted its bits using the current variable count.

The builder now scans levels from the top, searches each root once and revisits a shared subgraph when collecting a different root's support. A byte of existing node padding holds the local and global visit state. Consecutive variable indices within a support can be written to each packed row as bit ranges. Indexing rows by the larger index keeps pair positions stable as variables are added, and removal discards the corresponding rows. A 139-variable test compares the whole matrix twice against 25 independently constructed mixed supports in reverse level order.

For 32 independent 24-variable cones, 8,832 true pairs, Go went from 2.01–2.02 ms and 5,001 allocations to 28.6–29.3 µs and 14 allocations. On a 128-variable chain, 8,128 pairs, it went from 3.13 ms and 2,054 allocations to 6.08–6.12 µs and 18 allocations. The intermediate root-search version still set individual pairs and took 59 µs and 36 µs. The original C routine took 857 µs and 65.2 µs in the long runs. All pairs were also checked in separate C and Go runs. C profiles attributed 96% and 71% respectively to interaction initialization itself. All thirteen BDD package race suites passed.

### Validate and compact groups of retiring BDD variables

Red tests showed that destroying a BDD subtable could invalidate a referenced ADD, while a released BDD parent prevented valid removal. Replacement variables inherited the binding, type, pair index and lazy group of their predecessors. Further failures covered a saturated isolated projection, an uncounted collection and a pre-collection hook veto that left the manager partially resized. The native fixture confirms rejection for a live ADD or BDD parent, successful removal after dereference and one counted collection.

Go checks the live keys at each selected BDD/ADD level and validates the entire request before changing a projection. A saturated projection requires a scan for live parents. Collection respects active Go temporaries and removes explicitly released keys. One pass compacts the level widths and both permutations, retaining any capacity reserved by InitTable. Removed handles lose their manager, surviving handles keep their identity and newly allocated variables start with default attributes.

Inserting and removing 128 variables at level 64 in a 256-variable manager initially took 74.0–74.4 µs, 60.7 KB and 153 allocations per cycle. Table and collection changes brought this to 19.4–19.6 µs and 128 allocations. Initializing fresh projection pairs in groups of up to 128, then clearing the retired level records together, brought it to 10.9–11.0 µs and one 18,432-byte allocation. C took 10.2–10.8 µs. A test covers reuse of a preexisting projection inside a batch, partial removal and reserved capacity. Six-queens construction took 3.75–3.82 ms and 10,410 allocations after the change. All thirteen BDD package race suites passed.

### Collect released decision nodes under unique-table pressure

A red BDD test filled fourteen levels with 448 explicitly released keys. The old Go table doubled its width and retained 447 dead keys after reusing one as a child. Native C collected once, retained the eight-bucket level and finished with 31 BDD/ADD keys including constants. Red ZDD and numeric tests showed that even an existing-key lookup can trigger collection in those families. Disabling GC preserves the original keys and expands the table. The Go implementation now matches all four native cases, preserves unpinned intermediate Go results and removes cache entries referring to retired handles.

Additional red tests covered a collection hook that interns the same key as its caller, a failed bounded allocation that grew a subtable during rollback and an overbroad slowdown that affected constants, ZDDs and BDD variables created after the slowdown. A successful rehash restores the native density threshold for the affected BDD level. A separate red memory test found 256 probe slots retained for one surviving projection and another 256 for seven nodes. Sweeping directly in probe order permits inline storage of a lone survivor and downsizing of sparse backing arrays.

Go initially took 100–105 µs for the complete pressure workload, including manager creation. Sweeping in probe order brought this to 84.6–85.0 µs and about 115 KB per manager. C took 35.2 µs, with 71% of its sampled time in table initialization. The initial Go profile attributed 37% cumulatively to allocation. The thirteen BDD package race suites and base/abci passed on the final implementation.

The full-module run started with automatic table-pressure collection passed all 130 packages with tests.

### Register present/next-state maps for BDD reachability

Both network and AIG initializers built the initial cube without registering the state-variable swap advertised by their C counterparts. The two native public entry points were missing. Once the core API was present, separate red integration tests showed each initializer still leaving the manager unmapped. Both reachability drivers now register once and reuse mapped functions across image iterations. Direct callers that supply their own initial cube still get the default latch swap when needed.

Tests cover complement phase, replacement and invalid registration, computed-cache lifetime, heap shuffling, variable creation and subtable destruction. All 256 three-variable functions are checked for an overlapping mapping. The initial registration preserves unrelated cached operations, subsequent registrations flush them and variables created later map to themselves.

Mapping all 256 eight-bit state cubes originally took 277–279 µs, 30.7 KB and 768 allocations per batch through general permutation. The registered version took 10.9–11.1 µs and zero allocations. C took 7.62 µs. Its profile attributed 24% of samples to unary cache lookup and 20% to mapping recursion. The Go profile also identified cache lookup and canonical node probing.

A six-bit enabled counter reaches its first all-ones state at frame 63. Both Go image modes replay the returned trace. Partitioned mode went from 2.02–2.09 ms and 5,202 allocations to 1.89–1.99 ms and 4,778 allocations. Monolithic mode went from 1.79–1.80 ms and 4,976 allocations to 1.62–1.72 ms and 4,553 allocations. The original C partitioned solver took 675 µs. Its monolithic solver returned unknown before image iteration because bbrReach.c tests the partitioned tree pointer after constructing the monolithic tree. A temporary native comparison build checked the selected pointer, replayed the same trace and took 515 µs. The repository's C source remains the benchmark reference. In the new Go profile, fused conjunction and abstraction account for 46% of samples and counterexample construction for 52%, including backward images. The thirteen BDD packages, base/abci and misc/extra passed their race tests.

### Restore numerical ADD composition and ordered substitutions

Seven retained tests failed against the preceding implementation. Applying weights 3 and −2 to leaves 5 and 7 returned 5 where C computes 1. A nested case returned 7 instead of −14. Sequential substitution followed variable index in the wrong direction, the internal entry ignored its remaining cube, and four APIs accepted foreign-manager inputs on terminal shortcuts. Two permutation regressions covered an incomplete final triple and listing the same variable swap twice.

The composition routines validate all participating edges and use the manager's current level order. Simultaneous composition stops below the deepest substitution. Sequential composition applies the deepest replacement first. Weighted composition multiplies each child by its own weight and sums the products. The native driver and the Go benchmark both check the result on a reverse-ordered manager. A separate Go test checks all eight assignments with nested, overlapping, zero and negative weights.

The first correct implementation took 841–855 ns and 18 allocations per weighted composition. Its profile attributed 93% of time to generic arithmetic application, including allocations for operand pointers passed to callbacks. Giving the two arithmetic operations distinct computed-cache keys and passing operands directly reduced this to 109–111 ns and zero allocations. Original C took 106 ns over three million calls. The thirteen BDD package race tests passed.

### Fuse cached conjunction and existential abstraction

A bounded image operation returned nil after adding a unique key and advancing the next ID. A second red test showed an inner operation overriding its parent's tighter ceiling. A third found that existential, universal and XOR abstraction returned nil for a negative cube without setting the native invalid-argument error. A fourth showed that swapping the operands of a repeated image query produced zero cache hits. All four tests now pass. Independent truth tables check existential and universal abstraction on all 256 three-variable functions and all eight cubes after shuffling levels; the existing 524,288-case test checks the fused result.

Fused image recursion and existential abstraction now share the manager's computed table across calls, skip irrelevant cube levels and stop when a quantified branch determines the result. Universal abstraction uses the complement dual. Limited image operations share the same rollback and nested-ceiling logic as limited conjunction. Cache hits still check the caller's cancellation context.

The commuted query fell from 449–459 ns to 39–44 ns, with zero allocations. Original C took 9.5 ns with one lookup and one hit, and spent 62% of its CPU samples in the cache lookup. On the six-bit counter, partitioned Go took 1.72–1.84 ms on a same-core repeat (1.79–2.15 ms in the profiled run), down from 1.89–1.99 ms. Allocations fell from 4,778 to 4,152, while bytes rose from 1.04 MB to 1.12 MB. Monolithic Go fell from 1.62–1.72 ms to 1.36–1.46 ms and from 4,553 to 3,879 allocations. Conjunction and abstraction still account for 38% of the counter profile, and trace replay for 17%. The thirteen BDD packages, base/abci and misc/extra passed their race tests.

### Reuse the AIG evaluation order and buffers during trace replay

Two red replay tests exposed incorrect handling of trace dimensions. A graph with one register and an additional output returned two next-state bits. An oversized trace with an int32-wrapped bit count began simulating its first frame before discovering invalid data. Next-state extraction now reads exactly the registered latch inputs and validates the bit count with 64-bit arithmetic before evaluating a frame.

The old trace replay copied the graph and its terminal lists, computed an all-object order and allocated evaluation buffers on every frame. It spent 33% of its samples obtaining the order and 31% allocating. Replay now obtains the order once and uses the AIG's scalar simulation loop with reusable inputs and values. Tests check dirty buffer reuse, rewritten logic with non-topological IDs, a deletion hole, a dangling AND, explicit XORs, inverted buffer fanins and final graph marks. Existing randomized tests compare all final outputs and registers to independent graph evaluation.

Replaying the six-bit counter's 64 frames took 117–123 µs, 176,136 bytes and 705 allocations before the change. It now takes 8.85–9.12 µs, 272 bytes and five allocations. C took 6.92 µs, with 99% of its samples in the replay function. The full partitioned Go solver took 1.49–1.55 ms on a same-core repeat and the monolithic solver took 1.10–1.13 ms. Both replay the trace twice and now use 1,400 fewer allocations per solve. AIG, SAIG, BBR, PDR, SSW and base/abci passed their race tests.

### Pack sparse computed-cache positions into integer arrays

A red allocation test found two heap allocations when inserting four computed records, although the records themselves fit inline. The first four slot numbers now also fit inline. Larger sparse indexes use open-addressed 64-bit words that contain both the logical cache slot and its packed record position. Dense indexes retain direct integer lookup. Cache hits compare the stored key in place. A 300-entry strided-slot test covers sparse growth, unrelated misses, partial invalidation, and replacement of both live and invalidated entries.

The Go map implementation took 39–44 ns for the commuted image hit. The packed index took 28.3–28.4 ns in the final trial, with zero allocations in both cases. The six-queen solve took 3.69–3.76 ms and 10,396 allocations, compared with the earlier 3.75–3.82 ms and 10,410 allocations. A batch of 256 state maps stayed at 10.9–11.0 µs. The C references for these workloads were 9.5 ns, 1.68–2.60 ms and 7.62 µs respectively.

Across paired 650-solve runs on the same core, partitioned reachability went from 1.50–1.67 ms to 1.35–1.37 ms and used 33 fewer allocations. Monolithic reachability went from 1.10 ms to 1.03–1.09 ms and used 22 fewer allocations. Bytes per solve stayed near 768 KB in the partitioned case and fell from 682 KB to 672 KB in the monolithic case. All thirteen BDD package race suites, base/abci and misc/extra passed.

### Preserve the structural AND hash across network edits

Seven red cases showed that an unrelated edit could replace the newest duplicate AND with the oldest. Other failures covered a deletion callback seeing the removed entry, losing the native predecessor after deletion or object reuse, bucket order after rehash, accepting a cached non-node, and unchanged pin or data writes displacing a valid entry. A miss on an excluded node could repopulate the network's hash and spuriously resize an empty manager. The native fixture checks insertion, deletion, output editing, rehash and resize, including growth from three to eleven bins on the eighth insertion.

Network edits now invalidate the affected fanin pair. A raw node rewired into a cached pair still participates in the existing lowest-object-ID fallback. Explicitly registered duplicates retain their insertion order, reverse on rehash, and reveal their predecessor when removed. Unique pairs use the normal hash entry alone. Deletion unlinks the node before calling the cut callback. The tests distinguish edited functions, complemented pins, excluded nodes, and the source ordering of registered duplicates.

Toggling an output and looking up its shared-input AND took 4.71–4.79 µs in Go. Fanout scanning accounted for 98% of the old profile. The updated lookup took 60–65 ns and zero allocations. C took 5.4–5.5 ns and spent 94% of its samples in lookup. A paired i10 CEC run with 5,051 miter ANDs went from 327 ms to 307–310 ms; C previously took 222 ms. Network checking was within the prior range at 324 µs versus 318–325 µs. The base/abc, base/abci, base/io, bool/dec and map/scl race tests passed.

The full-module snapshot containing cached BDD abstraction and reusable trace replay passed all 130 packages with tests, including GIA in 463 seconds.

### Retire collected decision handles and group short-lived allocations

Four red correctness tests exposed dangling Go handles after explicit collection, reordering, variable deletion and bounded rollback. A collected BDD could revive references to its old children. A rolled-back handle could enter a manager after its numeric ID had been reused. Releasing a newly created key before rollback also dropped two unrelated external pins from one to zero. The manager now detaches both handle phases and their child links at retirement, and accounts for child references once.

The pressure fixture previously allocated 541 objects for a batch containing 448 released keys. Allocating every decision pair in groups of 31 brought that down to 111. After collecting 4,096 keys with one survivor every 64, however, each manager retained 287 KB versus 33 KB with individual allocations. Unconditional groups of four still retained 58 KB. The selected version starts a four-pair group only when the most recently issued pair has released its children. A 40-manager comparison retained 33.2 KB in the strided case and 41.5 KB after six queens, versus 33.0 KB and 40.5 KB for individual pairs. The retained-heap benchmark in cuddNodeLifetime_test.go reproduces both workloads with a fixed 40 iterations.

Pressure collection improved from 83.1–84.0 µs and 541 allocations to 76.7–79.1 µs and 205 allocations, with 115,200 allocated bytes in each case. C took 35.2 µs. Six queens took 3.76–3.79 ms versus 3.67–3.72 ms in this comparison, with allocations falling from 10,396 to 10,308. The six-bit reachability solver took 1.46–1.53 ms partitioned and 1.02–1.07 ms monolithic. The thirteen BDD packages, base/abci and misc/extra passed their race tests.
