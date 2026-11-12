package proof

// Lit is the MiniSAT-style packed literal used in proof-log clauses.
// Variable v's positive literal is v<<1; its negation is the low bit set.
type Lit uint32

type lit = Lit

type Pr_Man_t = Pr_Man_t_
