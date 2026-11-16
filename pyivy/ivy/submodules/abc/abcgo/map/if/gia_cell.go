package if_

// Gia_ManFromIfLogicFindCell is disabled in ABC by an unconditional assert.
// The source keeps a reconstruction body after the assertion, but no call can
// reach it. Keep the same unconditional failure without unused Go machinery.
func Gia_ManFromIfLogicFindCell(p *If_Man_t, dst, temp *Gia_Man_t, cut *If_Cut_t, cell *Ifn_Ntk_t, lutMax int32, leaves, lits, cover, mapping, records, configs *Vec_Int_t) int32 {
	panic("if: programmable-cell reconstruction is disabled in ABC")
}
