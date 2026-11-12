# Ivy port progress

Completed API groups include attributes, central node/manager representation,
header edge/node/init helpers, paged memory, basic manager lifecycle, structural
hashing, object construction/deletion, Boolean operations, canonicalization,
fanout tracking, validation and combinational balancing. These groups are tested
together independently of the remaining drafts. Whole-package compilation is
still blocked by unported replacement/manager/history/DFS/checking and advanced
mapping/fraiging/rewriting/sequential routines. No draft file is hidden or
excluded from the package build.

Attribute integer values are `int32`; pointer attributes are Go `any` values.
Managed fixed-size attributes use zeroed byte slices instead of raw allocator
addresses. Slot cells keep `FetchSpot` addresses stable across table growth.
`SetCleanup` replaces callback fields with typed functions. Managers are not
concurrency-safe; independent managers may be used concurrently.

## Separate review passes

The compactness/readability review eliminated duplicated array size/capacity
fields and C pointer/integer reinterpretation. Table lengths are slice lengths,
unused slots are nil, and integer cells allocate only when needed. The managed
mode and attribute size remain because they describe ownership and entry shape.

The bug review covered original-header defects: growth only doubled once even
when an ID lay far beyond that size; initialization repeatedly wrote the
requested slot instead of clearing all new slots; and managed fetch allocated a
new buffer when one already existed. The Go implementation grows to include the
requested ID, relies on zero-initialized slices and preserves existing buffers.
It also rejects negative IDs and kind mismatches, normalizes typed nil pointers,
and makes stop/cleanup idempotent under callback reentrancy.

Focused validation (does not imply the whole Ivy package is ported):

```
go test attr.go attr_test.go
go test -race attr.go attr_test.go
```

Tests exercise far-ID growth, old slot pointers after growth, integer/pointer
updates, managed buffer initialization/persistence, cleanup ordering and invalid
operations. No draft files are hidden or excluded from the package build.

## Core representation, lifecycle and table review

Regular nodes own state; complemented pointers are cached Go proxies referencing
regular nodes. This retains pointer equality without unsafe pointer tagging.
Integer edges are uint32 literals with a complement bit. Node IDs and traversal
fields retain int32 APIs. Manager vectors are typed slices, with sizes/capacities
derived directly; no local vector wrappers or extra dependencies are needed.

Memory consists of owned Go node pages and a free-slot slice. Fetch/recycle clears
node state while preserving the cached complement identity. Stop releases graph
references and invalidates ownership; foreign, duplicate, live/published, ghost
and constant-node recycling is rejected. There is no address arithmetic or
alignment trick. Constant and ghost identities follow the original manager.

Basic lifecycle routines live in ivyLife.go. Their already-ported duplicate
bodies were removed from ivyMan.go; its other routines remain visible drafts.
StartFrom copies only the constant/input boundary and establishes pEquiv maps,
matching the original function's scope.

The structural table uses a Go map of the complete structural key: both signed
fanin pointers, object type and latch initial state. Built-in hashing replaces
manual open addressing/prime-table allocation. Replacement updates preserve the
new object's identity, and reject absent/duplicate or structurally inconsistent
entries. Resize rebuilds the map; profile reports its entry count rather than
C-specific bucket-chain statistics.

The compactness/readability pass removed duplicated table/vector allocation
metadata and cast-based memory operations. The bug-review pass corrected the
original fanout-polarity helper reading polarity from the driving node instead
of the fanout edge, protected signed edge/leaf shifts, preserved complement proxy
identity through clean/overwrite/recycle, and tightened memory/hash ownership
checks. Tests cover phase/init algebra, bit boundaries, page expansion, manager
isolation, lifecycle, mapping, table key distinctions and identity transfer.

Earlier core/table checks (the current combined suite also needs the construction and balance files listed below):

```
go test ivy.go ivyMem.go ivyLife.go ivyTable.go ivyCore_test.go ivyTable_test.go
go test -race ivy.go ivyMem.go ivyLife.go ivyTable.go ivyCore_test.go ivyTable_test.go
```

A lifecycle review regression also caught manager-owned references surviving when
StopMemory had been called before ManStop. ManStop now always clears graph/table
references even when the memory pages were already stopped. Combined attribute,
core, lifecycle and table tests pass together under the race detector.

## Balance and adjacent construction APIs

Algebraic balancing is now ported, along with real Boolean constructors,
canonical AND/XOR/latch construction, node creation/connection/deletion,
intrusive fanout tracking, cleanup, buffer resolution, traversal IDs, graph
validation and cycle detection. History creation and representative lookup are
ported; the remaining history algorithms are still drafts. Matching completed
bodies were moved out of draft object/manager/utility/check/DFS/history files;
all unported bodies remain visible. Balancing explicitly requires a valid
combinational graph, as the original algorithm cannot balance latch leaves.

The separate readability pass uses pointer-stable scratch slices for recursive
supergates, derives their sizes from len, shares the actual PI constructor with
StartFrom, and shares cycle-DFS state across roots to keep validation linear.
It retains the original level-order pairing and structural-sharing heuristic.

The separate bug pass fixed native XOR supergate multiplicity: duplicate leaves
cancel, opposite leaves toggle parity, and rebuilding XOR cones must preserve
operand multiplicity. The original generic AND-style deduplication was wrong
for native XORs. Zero/singleton cones are handled directly, and pairing checks
same-regular operands before creating a hash ghost. Object creation also now
uses XOR rather than AND when computing a native XOR's phase.

Regressions compare all 256 assignments of an eight-input skewed AND before and
after balancing and require depth 3. They cover both level-update options,
complemented/constant outputs, contradictory ANDs, native XOR repeated/opposite
operands, Boolean/MUX/majority truth tables, fanout insertion and recursive cone
deletion. Tests use the real constructors and storage, with no test-only stubs.
The focused core, table, operation and balance suite passes; whole-package
compilation still fails in the visible unported ivyCheck.go/ivyCut.go routines
and later drafts. Replacement/buffer propagation, the remaining manager
transformations, full DFS/level maintenance, full history, mapping, cuts,
fraiging, rewriting and sequential optimization remain unfinished.
