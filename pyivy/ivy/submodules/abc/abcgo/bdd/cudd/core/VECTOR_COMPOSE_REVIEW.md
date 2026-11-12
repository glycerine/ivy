# Simultaneous BDD vector-composition review

The generic primitive needed by extrab's fast translation-space computation is
implemented once in `cuddVectorCompose.go`, with original CUDD attribution and
license header. Root CUDD's larger compose family remains separately incomplete.

Readability pass: one memoized structural walk reconstructs each original BDD
node using its corresponding substitution and authoritative ITE. It never
recursively substitutes into newly inserted functions, so cyclic substitutions
are simultaneous rather than order-dependent. No new unique table exists.

Bug pass: validates manager/family ownership for the input and every substitution,
requires the full manager-sized vector, handles constants and complemented
edges, and keeps substitution indices stable under manager order. Tests check
variable swaps/cycles, all 256 three-variable functions against direct Boolean
substitution, short vectors and foreign roots. Original extrab Auto fast-path
oracle additionally exercises the primitive in the exact source algorithm.

Full core race/vet and the extrab Auto regression gate are required.
