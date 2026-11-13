package aig

// Aig_ManSetRegNumForConstruction records the register count while a graph
// is being assembled, after all inputs but before all outputs are present.
// Call Aig_ManSetRegNum with the same count after adding the outputs to check
// that the finished sequential interface is valid.
func Aig_ManSetRegNumForConstruction(p *Graph, n int32) {
	if n < 0 || int(n) > p.InputCount() {
		panic("aig: invalid register count for graph under construction")
	}
	p.nRegs = n
}
