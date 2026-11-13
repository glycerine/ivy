package abc

import "github.com/glycerine/abcgo/misc/vec"

// SetSequentialCounterexamples transfers the per-output witness vector to n.
// Entries can be nil or a non-nil status marker as well as a witness.
func (n *Network) SetSequentialCounterexamples(models *vec.Vec_Ptr_t) {
	n.seqModelVec = models
}

// SequentialCounterexamples returns the network's borrowed per-output vector.
func (n *Network) SequentialCounterexamples() *vec.Vec_Ptr_t { return n.seqModelVec }

// TakeSequentialCounterexamples transfers the vector out of the network.
func (n *Network) TakeSequentialCounterexamples() *vec.Vec_Ptr_t {
	models := n.seqModelVec
	n.seqModelVec = nil
	return models
}
