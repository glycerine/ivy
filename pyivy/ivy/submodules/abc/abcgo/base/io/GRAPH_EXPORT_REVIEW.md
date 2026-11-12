# Graph export family

The active DOT, GML, adjacency-list/CellNet, and Bookshelf algorithms from
`src/base/io/ioWrite{Dot,Gml,List,Book}.c` are native Go. File entry points retain
source names and return errors. Writer-based entry points expose error-aware
rendering; selected DOT objects use native object slices. Production registry
handlers are installed for DOT/GML/LIST/BOOK.

DOT retains ranked layout, forward/reverse levels, constants, terminal/latch
shapes, highlighted selections, complemented edges, and sequential terminal
collapse. SOP, mapped-gate, and BDD labels are derived without converting the
source network's representation. GML retains source consumer-to-fanin edge
orientation and arrow-at-first-end convention. Lists retain zero latency and
optional HOST arcs. CellNet retains latch-input cell IDs and source pin numbers.

Bookshelf exports the actual six-file placement family: cell dimensions/areas,
terminal and internal pins, nets, alternating row orientations, breadth-first
pad ordering, perimeter distribution, and initial core-cell placement. SOP and
mapped cells retain source sizing rules, latches have width six, and logic/BDD
conversion uses an independently duplicated network. Hierarchy/EXDC modules are
qualified and merged under one set of counts rather than appending invalid
repeated file headers. Source driver-I/sink-O pin conventions remain unchanged.

## Readability and compactness pass

Completed independently from bug review. One explicit placement model replaces
repeated Bookshelf node-name reconstruction and raw pointer arrays. Separate
DOT views share rendering and retain their interface policy. Generic buffering
keeps validation failures from truncating output files. Go-owned maps replace
borrowed node marks, level mutations, and static SOP buffers. Escaping is explicit
for DOT and GML. Errors propagate through reader/writer registry contracts.

## Independent bug pass

Checked all retained graph endpoints, complement propagation across BI/BO latch
terminals, selected-set boundaries, duplicate pin occurrences, pad counts, cell
names, and layout dimension calculations. Fixed a missing BDD scratch pointer
caught by a direct BDD-label regression. Source marks, levels, BDD identity, and
function-manager type stay unchanged after visualization.

Corrected inherited source faults: Bookshelf pad bookkeeping indexed a
terminal-count allocation by arbitrary object IDs; disconnected/one/zero-pad
cases could fail or omit pads; zero area caused division by zero; layout call
arguments could interchange width/height. Native maps, deterministic fallback,
and positive dimension bounds fix these cases. Terminal records use standard
width/height/terminal ordering. Actual EXDC data replaces the source's accidental
second copy of the main network. CellNet repeated fanins emit each pin once.
GML defines constant and box-terminal endpoints when needed. DOT filters external
selection endpoints and safely handles labels beyond C's static 1000-byte buffer.

## Validation and package status

Focused tests cover GML/list connectivity and HOST arcs, sequential DOT collapse,
reverse and selected views, escaping, source-state preservation, BDD labeling and
Bookshelf conversion, writer/file errors, foreign/empty selections, repeated
CellNet pins, disconnected pads, empty layouts, and Bookshelf file/pin counts.
The actual unchanged-C oracle compares GML nodes/edges and adjacency HOST output
exactly except timestamp headers, and exercises original Bookshelf cell/net
counts and sizes. It does not claim parity with the unsafe original pad allocator.
The C oracle actually ran successfully; missing compilers/objects are visible skips.

Focused real-source commands:

```
abcgo/base/io/test_graph_exports.sh
abcgo/base/io/test_graph_exports.sh -race
GRAPH_CHECK_MODE=vet abcgo/base/io/test_graph_exports.sh
```

The harness copies the native ABC dependency list and production registry into a
temporary module. It uses actual implementations and leaves original drafts
visible. It is not evidence that the whole base/io package builds.

The ordinary gate was run:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./base/io
```

It currently fails in imported `base/abc/abcLatch.go` at line 30 and later
syntax errors (the earlier `abcHieNew.go` frontier moved during integration). No ordinary base/io success is claimed.
