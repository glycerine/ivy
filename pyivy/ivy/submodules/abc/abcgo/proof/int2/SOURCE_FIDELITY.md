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

Independent components have source-based repairs and tests: timeframe
construction preserves complemented edges and maps constants and SAT
variables; JF CNF results use their actual Go type; refinement selects
justifying CIs using their values and priorities; suffix construction uses
the property for the initial image; and preimage blocking translates between
prefix and suffix SAT variable maps. These repairs do not establish that the
unavailable full interpolation algorithm has been restored.

The source's cube refinement asserts finite priority even for a constant-true
suffix. The Go implementation handles that case with an empty justifying cube;
both refinement and preimage behavior have regression tests.

