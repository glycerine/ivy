package abc

import "github.com/glycerine/abcgo/misc/vec"

// Abc_NtkReadSeqModelVec returns the network's borrowed per-output witness vector.
func Abc_NtkReadSeqModelVec(n *Network) *vec.Vec_Ptr_t { return n.seqModelVec }

// Abc_NtkSetSeqModelVec transfers ownership of the per-output witness vector.
func Abc_NtkSetSeqModelVec(n *Network, models *vec.Vec_Ptr_t) { n.seqModelVec = models }
