# ABC network and graph integration review

## Readability and compactness pass

The network bridge keeps ABC's CI/CO ordering and names while translating the
strashed network to the explicit `aig.Graph` API. It uses maps only at the
network boundary; the resubstitution search itself keeps stable graph edges.
The rebuilt ABC network is fresh, so an unsuccessful rewrite never partially
mutates the caller's network.

## Independent bug pass

Checked constant-one mapping, complemented AND inputs, object topological
ordering, CI and CO ordering, unsupported non-AIG nodes, graph holes after
rewrites, explicit XOR and buffer graph nodes on reconstruction, output names,
network name/spec preservation, structural hashing, and source non-mutation.
The integration test exhaustively compares two output functions over all
three-input assignments and verifies input/output names and output counts.

## API boundary

The bridge accepts structurally hashed ABC AIG networks with combinational
object functions. It returns a new network and does not rewrite the input
network in place. The normal test currently cannot execute because the
`base/abc` package has unrelated syntax errors; this test must run once that
package is buildable.
