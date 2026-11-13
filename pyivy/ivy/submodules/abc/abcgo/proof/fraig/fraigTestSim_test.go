package fraig

// The C bit simulator assumes a nonconstant root; the oracle handles constant
// edges directly so randomized networks can still be checked exhaustively.
func fraigTestSimulateBitNode(p *Fraig_Man_t, edge *Fraig_Node_t, model []int32) int32 {
	if Fraig_Regular(edge) == p.pConst1 {
		if Fraig_IsComplement(edge) {
			return 0
		}
		return 1
	}
	return Fraig_ManSimulateBitNode(p, edge, model)
}
