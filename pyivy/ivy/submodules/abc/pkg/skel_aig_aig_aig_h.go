package abc

// Hand-written skeleton of types in src/aig/aig/aig.h that have no Go port
// yet (declarations only).

// AigCut is Aig_Cut_t: the cut used to represent node in the AIG.
type AigCut struct {
	pNext    *AigCut // the next cut in the table
	Cost     int     // the cost of the cut
	uSign    uint32  // cut signature
	iNode    int     // the node, for which it is the cut
	nCutSize int16   // the number of bytes in the cut
	nLeafMax int8    // the maximum number of fanins
	nFanins  int8    // the current number of fanins
	pFanins  []int   // int pFanins[0] : flexible array member, the fanins (followed by the truth table)
}
