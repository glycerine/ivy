package abc

import "github.com/glycerine/abcgo/aig/aig"

// TakeInductionCounterexample moves the induction algorithm's compact witness
// into the network's typed sequence-model slot. Its compact dimensions differ
// from a full initialized sequential trace, so no trace-layout conversion is
// performed. The source graph no longer owns the witness after this handoff.
func (n *Network) TakeInductionCounterexample(graph *aig.Graph) {
	if n == nil || graph == nil {
		panic("abc: nil induction witness handoff")
	}
	n.Model = nil
	n.SequenceModel = graph.TakeInductionCounterexample()
}
