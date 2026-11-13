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
