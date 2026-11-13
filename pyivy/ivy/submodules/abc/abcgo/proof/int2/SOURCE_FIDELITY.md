# INT2 source fidelity

The checked-in `src/proof/int2` implementation is incomplete. In
`int2Core.c`, `Int2_ManUnroll` refers to undeclared `pAig`, the prefix and
suffix constructors refer to undeclared locals, and `Int2_ManComputePreimage`
uses a manager `p` that is not a parameter. The interpolation driver contains
undeclared image/map variables, an invalid member-access token, a call with a
different argument count from its definition, and a reference to a manager
field that is not declared. It cannot be built as a C reference in this state.

The Go package retains an exact, bounded reachable-state fallback in
`Int2_ManPerformInterpolation`; it is not a port of the interpolation driver.
It is limited to small transition systems and must not be taken as evidence of
equivalent performance, resource handling, or an interpolation certificate.
Reconstructing the missing driver requires a reliable source or a separately
specified algorithm, rather than assuming its absent control flow.

Source search: the other local ABC checkout has the same `int2Core.c` bytes
(SHA-256 `99f3fa643b19dd3e65c317ed6f47ba8c845145be8091a34fe111c1ed03837272`).
Its available git history has one commit for that path,
`b910cba3e2e065d444d4a7b1bcbf81227417fc11` (2014-01-28,
"Initial new interpolation code"). The current workspace's available history
has only the vendoring commit for that path. No complete driver was found in
these local histories. Attempts to inspect the Berkeley ABC upstream driver and
its history through the web tool failed, so those sources remain unverified;
this search does not establish that no complete implementation exists elsewhere.

Independent components have source-based repairs and tests: timeframe
construction preserves complemented edges and maps constants and SAT
variables; JF CNF results use their actual Go type; refinement selects
justifying CIs using their values and priorities; suffix construction uses
the property for the initial image; and preimage blocking translates between
prefix and suffix SAT variable maps. These repairs do not establish that the
unavailable full interpolation algorithm has been restored.
