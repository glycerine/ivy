# MF mapped-CNF review

Ported giaMf.c CNF template derivation and emission against Mapping's shared
Network/Cut/Truths/Refs model. DeriveCNF preserves CO-first variable numbering,
reverse mapped-node numbering and emission, constant then CI numbering, optional
leading output-OR clause, output-before-input LUT literal ordering, constant
clause last, and source flat mapping/first-clause metadata. Synthetic COs support
networks that expose only output drivers; explicit COs retain source IDs.

## Compactness and readability review

No second graph or cut representation is introduced. Function templates are
cached locally by interned truth ID. Shared literal, initialization and flat
mapping helpers replace pointer arithmetic. Clause slices replace C's packed
pointer/sentinel arrays; counts follow their actual lengths. Inputs are borrowed,
while clauses, IDs and mapping payloads are independently owned. LUT-width
selection uses six-variable conversion directly and pads seven-variable truth
storage for the eight-variable utility.

## Bug review

Reviewed complemented outputs/cuts, zero-support cuts, unused gates, sparse
object IDs, empty output OR, small LUT truth widths, missing leaves and truth
storage, and mapping payload lengths for up to eight leaves. The source mapping
truth copy tests parity of iFunc (the function index), rather than the cut's
function literal complement. The port uses Cut.Function's low bit; exhaustive
phase tests cover this correction. Constant-cut flat mapping retains the
source's explicitly unsupported case and returns an error. Object-ID mode sizes
the variable bound to include the largest used ID, fixing the source's compact
count bound when unused objects create gaps. Empty output-OR clauses are valid
unsatisfiable clauses and do not index clause[0] during metadata construction.

## Source and semantic tests

An AND AIG generated as binary AIGER was passed to the repository's original ABC
executable with `&r -s /tmp/mf_and.aig; &write_cnf -K 8 /tmp/mf_and.cnf`.
All seven emitted clauses and the six-variable bound match exactly after the
ABC-literal-to-DIMACS conversion (DIMACS variable is internal variable plus one).
The expected clause sequence is checked in cnf_test.go. All assignments verify
both cut phases in compact and object-ID modes, including flat mapping truth
polarity. Directed tests cover synthetic constant outputs, an empty output OR,
small LUT storage and constant-cut clauses/mapping rejection. Entire mf package
passes race tests and vet together with its mapping/cut tests.

## Independent wide-cut follow-up

Added 24 deterministic random seven/eight-leaf functions, each in both cut
complement phases. Every leaf assignment and both root values are checked
against the exact truth relation; all flat mapping truth words are checked too.
A separate test compiles the actual original utilTruth.h/abc_global.h C headers
in a temporary directory, queries Abc_Tt8Cnf for eight further random wide
functions, and compares every ordered output-first literal of every clause in
both phases. The C oracle runs when a compiler is available; the exhaustive Go
semantic checks always run. No stored fixtures or testdata directory is needed.
This review found no additional CNF semantic/ordering defects; both wide paths
and all phase combinations pass together with the mapping/IOA tests.
