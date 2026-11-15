# False-path detection and rewriting

Review all ten routines in `src/aig/gia/giaFalse.c`, implemented by
`false_path.go` and the typed representation boundary in `false_lower.go`.
Keep the original two-frame path/sides encoding, BSAT final-conflict selection,
controlling-value reconstruction, output priority and iterative driver order.

## Repairs and red evidence

The previous Go implementation used Satoko and deletion-based assumption-core
minimization. Both can select different rewrite intervals from the source's one
BSAT call. Native single-path comparison first fails at seed15, including an
object-count difference (`/tmp/abcgo-false-core-red.log`). Use the existing BSAT
port and its owned final conflict; preserve source clause order, including XORs.
Restore the source reverse-cone SAT-variable numbering in object Value scratch,
removing the redundant map (`/tmp/abcgo-false-scratch-red.log`).

Empty-hook rebuilding was reading real XOR/MUX/buffer gates as ordinary ANDs,
and lost reordered ports and sequential/interface metadata
(`/tmp/abcgo-false-rebuild-red.log`). Reuse the shared gate-copy helper for
unhooked nodes and the existing port-remapping/interface-copy helpers. Hooks
remain ordinary-AND certificates. At the algorithm boundary, lower typed and
simple representations once into an owned ordinary graph; preserve root
mapping before cleanup, logical port order and output polarity. The 128-case
typed driver regression is initially red
(`/tmp/abcgo-false-typed-driver-red.log`).

Stable sorting changed source heap tie order (five equal outputs must visit
0,4,3,2,1). Reuse Vec_Que with source weights/push order rather than a new sort
(`/tmp/abcgo-false-priority-red.log`). Keep source span<20, which can admit a
21-node hook: testing hook length<=20 was incorrect
(`/tmp/abcgo-false-interval-red.log`). Keep conflict size, original path length
and rewritten interval length distinct in diagnostics. Restore exact source
report order and timing-field width
(`/tmp/abcgo-false-reports-red.log`).

## Demonstrated source correctness defect

Typed lowering exposes an ordinary-AND batch rewrite defect also present in C.
A reduced three-input graph computes F = d ? (a XOR c) : !c. Original C batch
rewriting changes F(1,1,1) from false to true. The regression fails before repair
(`/tmp/abcgo-false-batch-reconvergence-red.log`). The original PI-transition
assumption can certify an impossible timing transition without certifying a
Boolean cut when side cones reconverge on that PI.

When the batch conflict uses assumption0, solve again without that condition,
as the original single-path criterion does. Reuse the same BSAT solver, learnt
clauses, absolute deadline and total conflict cap. Unknown never yields a hook.
This is an explicit source bug repair, not a substitute solver or an
approximation of its answers. C itself uses nTimeOut for both seconds and a
conflict budget; preserve that valid nonnegative resource policy.

The C fixture defaults to unmodified source. Its separate --reconvergence mode
must reproduce the wrong original result and verify the corrected result
against an independent eight-row formula. Only --safe-batch/--drivers/--budget
apply an explicit SAT-call bridge with the identical safety recheck. These
comparisons are labeled corrected-source comparisons; single-path, heap and
verbose comparisons still run the unmodified source. No solver answers are
stubbed or supplied by the tests.

## Ten-routine reconciliation

| C routines | Evidence |
| --- | --- |
| FalseRebuildOne | Controlling polarity, temporary Value restoration, source path diagnostics, native single/batch rebuilt graphs; independent truth for each accepted certificate. |
| FalseRebuild, FalseRebuildPath | Exact native result counts/outputs; hook/nonhook paths, typed empty hooks, reordered ports and independently owned interface names, resets, roles and comments. |
| CollectPath_rec, CollectPath | Exact native path IDs and fanin1 tie choice, CO driver lookup, a 6,003-node path and constant/CI-only boundaries. |
| CheckFalseOne | Exact corrected-source hooks and full Value scratch for 64 graphs/two outputs; 128 individual queries with exhaustive truth for every accepted hook; deterministic budget exhaustion with no collected hook. |
| CheckFalse2 | Original heap tie order and all verbose modes; corrected-source full-driver matrix with overlap, slack, budget and sequential dimensions. |
| CheckOne | 64 unmodified native graph/path/rebuild records, exact final-core interval behavior and diagnostic formatting, typed root mapping. |
| CheckFalseAll, CheckFalse | Source attempts/changes and final output truth in the 576-case driver matrix; 128 mixed typed driver cases and independently owned sequential interfaces. |

The native fixture includes actual giaFalse.c and links original ABC objects.
The shared oracle helper requires Linux/GCC and those built objects and skips
when unavailable. Independent Go scalar checks cover all assignments of each
small input graph and do not reuse SAT encodings or rebuild scratch. The native
resource test independently confirms Unknown with zero hooks at seed12/budget1;
unlimited mode decides the same query. The deep path checks the source path
ordering and complete rebuilt output truth without timing-dependent sleeps.

## Compactness and contracts

Share a small query result containing original path, owned final conflict and
status; keep single/batch reporting differences explicit. Reuse one rebuild
loop, with its source-specific path-numbering switch. Batch interval insertion
uses the owned original path slice directly and avoids allocating rejected
hooks. SAT variable IDs live in existing Value scratch, as in C. Solvers and
temporary typed graphs are disposed with defer. No persistent duplicate fields,
new cache or invented SAT engine are added.

Low-level path/rebuild helpers retain the source precondition of ordinary AND
path IDs and valid hooks. Expanded typed hooks cannot be mapped blindly back
onto original typed object IDs. Top-level drivers handle typed inputs at their
ordinary representation boundary. The singleton manual-hook behavior retains
the source constant-zero case; normal accepted safety certificates contain an
AND transition and do not produce that case. Node-indexed source payloads are
not copied as though IDs survived a rewrite.

## Validation

All focused false-path tests pass in19.197s
(`/tmp/abcgo-false-final-focused.log`). Final frozen module race and GIA vet
checks are pending. This family does not complete the remaining GIA inventory
or the sequential review of the original131 packages.
