package res

// The C-shaped translation of resSat.c was not executable Go: it depended on
// untranslated Abc_Ntk_t pointer bookkeeping and called the disabled
// Sat_SolverDoubleClauses path. The original remains at src/opt/res/resSat.c.
// The usable graph-backed SAT implementation is in resGraphSAT.go; it encodes
// aig.Graph nodes with BSAT and supports constrained simulation and exact
// equivalence checking. Interpolant proof generation remains unavailable until
// clause partition duplication has a supported, tested implementation.
