# INT2 source fidelity

The checked-in `src/proof/int2` implementation is incomplete. In
`int2Core.c`, `Int2_ManUnroll` refers to undeclared `pAig`, the prefix and
suffix constructors refer to undeclared locals, and `Int2_ManComputePreimage`
uses a manager `p` that is not a parameter. The interpolation driver contains
undeclared image/map variables, an invalid member-access token, a call with a
different argument count from its definition, and a reference to a manager
field that is not declared. It cannot be built as a C reference in this state.

`Int2_ManPerformInterpolation` returns `INT2_TIME_OUT` (the existing undecided
status) for every input: the driver is unsupported until a reliable source is
available. The previously invented bounded reachability fallback was removed;
this entry point reports neither a proof nor a counterexample.

Source search: the other local ABC checkout has the same `int2Core.c` bytes
(SHA-256 `99f3fa643b19dd3e65c317ed6f47ba8c845145be8091a34fe111c1ed03837272`).
Its available git history has one commit for that path,
`b910cba3e2e065d444d4a7b1bcbf81227417fc11` (2014-01-28,
"Initial new interpolation code"). The current workspace's available history
has only the vendoring commit for that path. No complete driver was found in
these local histories. The current [Berkeley ABC upstream driver](https://github.com/berkeley-abc/abc/blob/master/src/proof/int2/int2Core.c)
has the same SHA-256 as the local file. The GitHub commits API likewise reports
only the initial commit for this path. This search does not establish that no
complete implementation exists elsewhere.

A subsequent GitHub API search compared default branches of 471 distinct
publicly listed forks: 464 had the same file hash and seven did not have the
path. For example, the [Yosys fork](https://raw.githubusercontent.com/YosysHQ/abc/yosys-experimental/src/proof/int2/int2Core.c)
and [OpenROAD fork](https://raw.githubusercontent.com/The-OpenROAD-Project/abc/master/src/proof/int2/int2Core.c)
match. This sample does not cover every fork, non-default branch, or independent
repository; no complete alternative was found in the sampled default branches.

The C refinement code tests whether a register CI is a CO, treats priority
`Value == 1` as the true-AND test, compares Boolean marks when choosing a
fanin, and assigns infinite priority to the constant. These behaviors are
retained and can panic or choose a wrong cube; they are not correctness claims.
The C SOP loops increment the index and then read `i+k`, skipping some literals;
the Go helpers retain that indexing. A nil SOP is dereferenced by the source
loop and the Go helper also panics; a non-nil empty SOP produces false.

The C preimage helper refers to an undeclared manager and has no reliable
prefix/suffix mapping contract. `Int2_ManComputePreimage` therefore returns
nil with the existing undecided status. The suffix constructor also returns
nil without a map or graph: its clause loop references undeclared `k` and
`vSop`, and CNF is derived into `pNew` but read from `pSuff`. The C code
does not provide a reliable clause or CNF contract to port. No suffix,
preimage, or interpolation claim is made.

Other source helpers also have limits. `Int2_ManUnroll` uses the
undeclared name `pAig` throughout its C body; `Int2_ManPreparePrefix` reads
the CNF from `pPref` despite deriving it into `pNew`, and refers to undeclared
locals. The source-named Go entry points now return nil and produce no graph
or solver; prefix also clears its optional output map. The separately named
`Int2_GoUnroll` and `Int2_GoPreparePrefix` retain the inferred Go convenience
implementations, whose tests do not establish equivalence to executable C.
`Int2_ManCreateFrames` in `int2Bmc.c` likewise refers to undeclared `f`,
`iObj`, and `pFrames` and has inconsistent SAT-manager arguments. Its Go
implementation is reconstructed and has no executable C oracle.
The manager constructor in `int2Int.h` calls `Gia_ManHashAlloc(pFrames)`
without declaring `pFrames`; the Go constructor infers `pGiaPref` there.
In particular, the C AND branch passes the two mapped fanin literals to
`Gia_ManHashAnd` without applying the source fanin complement bits; the Go
implementation currently applies those bits. This is a difference in the
source text, not an executable comparison, because the C function does not
compile as written.

The source text does specify a work stack of `(frame, object)` pairs, skips
already mapped objects, maps a frame-zero register output to zero, follows a
later register output to the previous frame's register input, and applies the
CO complement when copying its fanin. These individual branches can be
compared mechanically with the Go code. It also describes hashing two mapped
fanin literals and adding a SAT AND when a new node is created; as written, it
does not apply the original AND fanin complement bits.

Other operations cannot be established from the C body: its constant setup
tests undeclared `f`, writes an undeclared `iObj`, and passes a GIA pointer to
`sat_solver_addclause`; its PI and AND branches use absent `pFrames` state;
and its SAT AND call also passes the GIA pointer rather than a SAT solver.
The separately named `Int2_GoCreateFrames` instead initializes constant zero in every new frame map, asserts
the SAT constant unit on each call, allocates SAT variables for appended PIs,
and uses `pGiaPref` and `pSatPref`. Those choices are reconstructions, not
source-confirmed behavior. The source-named `Int2_ManCreateFrames` has a void
signature, so it explicitly panics before changing state to report that its
source contract is unsupported. Individual matching branches in the Go-only
helper do not validate the full operation against C.

The Go `Int2_ManCheckFrames` returns -1 for missing frame/object mappings,
whereas the C helper directly indexes the vectors. The Go `Int2_ManCheckBmc`
returns undecided for an unconstructed output unless another output is known
to fail; C instead passes the unchecked mapping to its SAT call. The tests for
these Go outcomes describe defensive behavior, not an established C contract.
Likewise, `Int2_ManSetupBmcSolver` checks whether adding an output buffer
fails and returns nil, while the C code ignores that return value.
The separate SAT bridge follows the C helper distinction: adding an
inconsistent AND clause triggers an assertion-equivalent panic, whereas adding
an inconsistent buffer clause returns zero.

The graph-copy and frame helpers follow the checked-in C loop, which treats
all `Gia_ManForEachAnd` nodes as ordinary ANDs. This changes native XOR/MUX
semantics. A compiled C probe (`test_vectors/native_xor_copy.c`) confirms a
native XOR of two inputs copies as their AND; a Go regression checks the same
result across all input assignments. `Int2_ManCreateFrames` has an incomplete
C body as noted above, so the analogous incremental behavior remains inferred.
The Go-only zero-frame unroller returns the all-zero initial state. The
source-named frame initializer instead reads each register-input fanin's
existing `Value` when no frames are visited, as the C body does; callers must
not assume that stale or uninitialized source values describe an initial state.
