# Network core review

Readability/idiomatic pass: reviewed explicit Edge values, Go-owned object
storage, stable ID holes, fanin/fanout multiplicity, names via misc/nm, and the
dependency-free DecNetwork adapter. Converted the original abc.h public type,
edge, fanin and object-selection APIs in abc.go onto this representation. Replaced
abcFanio.go with Go implementations of its original connection, patching,
insertion, fanout-transfer and replacement routines. No tagged-pointer tricks,
unsafe memory or hidden/build-excluded draft source was introduced.

Separate bug pass: found stale structural hashes after AddFanin, PO objects
accepted as PatchFanin sources, and replacements capable of introducing cycles
or mutating before rejection of constant deletion. Added red tests, fixed hash
invalidation and source checks, and preflighted replacements. Tests cover
combinational cycles versus sequential latch boundaries, duplicate edge counts,
replacement complements, cross-network edges, structural hashing, truth-table
evaluation, original API fanin mutations and insertion. Focused original/core
file tests pass. Full-package verification is still impossible because the
remaining generated ABC algorithms are unported; this is not a package completion
claim.

Original API continuation: replaced the entire abcObj.go draft as well. All
original object routines now have Go implementations: allocation/recycling,
creation/deletion, recursive dangling deletion, object/box duplication, regular-
fanin cloning, name lookup/net interning, Boolean constant/buffer/inverter/
AND/OR/XOR/MUX creation, recognition and local-function polarity changes.
Boolean functionality uses Go functions, with CloneForNetwork for specialized
mutable managers; fanin vector arguments become slices. Other function-manager
representations still need conversion by their owners. This is documented API
adaptation rather than a claim that BDD/Hop manager migration is finished.

Readability pass on this continuation: consolidated node construction and typed
name lookup, retained meaningful ABC entry-point names, and replaced vector and
manual ownership operations with slices/GC. Separate semantic bug pass reviewed
fanout multiplicity, deleting connected objects, complement composition, box
terminal orientation, dangling-cone sharing and clone-without-function semantics.
Regression tests exhaustively evaluate binary gates and MUX inputs, check
constant/input complement behavior, names, net interning, box duplication,
allocation/recycle and connected/dangling deletion. Original/core selected-file
tests pass with -race; the rest of base/abc remains explicitly unported.

abcRefs.go continuation: all 14 original reference/MFFC routines now ported.
Scratch reference counters preserve the original zero-transition algorithm,
including duplicate fanins and complemented-edge stopping, without truncating
live fanout slices. The original fanout-count accessor exposes virtual counts
while dereferenced. Optional vector outputs are pointers to Go slices. Added
traversal IDs and wrap handling; collection/labeling preserves the original
boundary-versus-interior label distinction. Counts restore automatically when
balanced, and cut leaves add temporary reference barriers.

Readability pass: consolidated paired recursive reference/dereference walks and
cone/support/label collection while retaining original entry-point names.
Documented scratch-counter ownership and slice adaptation. Coordinated with the
AIG owner; base/abc owns object-network refs and does not duplicate tagged-pointer
storage from aig/aig.

Separate bug pass: found invalid cut leaves could increment earlier leaves and
leave scratch state modified before rejecting a foreign leaf. Observed a red
regression and moved all leaf validation before mutation. Also initialized the
network traversal epoch to one so fresh objects are unvisited, and tested epoch
wrap clearing. Regression coverage includes shared diamond cones, external
fanouts, cut boundaries, duplicate fanins, constants, complemented stopping,
interior/boundary labeling, reference restoration, immutable graph topology,
foreign-leaf rejection and traversal wrap. Selected original/core files pass
-race. Remaining generated base/abc files still prevent package-wide compilation.

abcAig.go continuation: replaced the next package-blocking generated file with
all 39 original manager entry points. The native ABC object manager retains
network object identities; it does not alias a second aig.Graph. Coordinated
edge/model APIs with aig/aig's owner. Ported AND structural hashing, Boolean
OR/XOR/MUX creation and lookup, balanced/sequential/implication miters, dangling
cleanup, hash rebuilding, queued replacement with fanout re-strashing,
identity-preserving fanout recreation, cascading deletion with pinned pending
replacement targets, forward/reverse level updates, phase computation, choice
checks, update-list lifetime/reset and diagnostics. Container outputs now use
slices or pointers to slices; absent lookup returns Edge{Object:nil}. Explicit
Choice and Phase fields replace overloaded C fields/tagging. Level recomputation
uses topological traversal instead of manual per-level work queues.

Readability pass: consolidated complemented-edge normalization, single-node
construction, lookup variants, update collection and level removal helpers.
Preserved ABC entry-point names and kept manager/network responsibilities clear.
Separate bug pass found that graph-scanning hash fallback could rediscover a
node explicitly removed from the structural table. Observed a red test and
added manager-local exclusions, respecting them through lookup, rehash, cleanup
and collision diagnostics. Foreign-input recreation and dependent replacements
are rejected before mutation. Regression coverage includes alternate XOR/MUX
lookup phases, choice cycles/fanout restrictions, complemented output rewiring,
fanout simplification and identity preservation, miter/update-list behavior,
cleanup and 80 randomized 4-input replacement truth tables. Selected-file tests
pass -race. Full package now advances past abcAig.go and fails at the still-
unported abcBarBuf.go draft; no remaining draft source was hidden.

Current coherent API group now includes original object/fanin APIs, reference-
cone operations and the AIG manager. The next blocker, abcBarBuf.go, contains
hierarchical module flattening/reconstruction as well as barrier-latch/buffer
round trips. It requires a real design/module representation and ordered
PI/PO/box-terminal mappings; replacing it with buffer-only shortcuts would lose
its primary functionality, so it remains explicitly unported pending that work.

Hierarchy continuation supersedes the earlier pending-barrier note: hierarchy.go
now provides the real Design/module model and ordered terminal identity storage,
and all 14 original abcBarBuf.go routines are ported. Separate readability and
bug/fidelity reviews are in HIERARCHY_REVIEW.md. End-to-end nested flattening,
module reconstruction and barrier/buffer truth-table tests pass with the rest of
the original/core -race tests. The next full-package blocker is abcBlifMv.go.
