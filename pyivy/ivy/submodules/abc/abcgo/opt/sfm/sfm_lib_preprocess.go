package sfm

import "github.com/glycerine/abcgo/map/mio"

// Sfm_DecCreateCnf builds local gate clauses in library order.
func Sfm_DecCreateCnf(sizes []int, truths []uint64) [][][]int {
	if len(sizes) != len(truths) {
		panic("sfm: inconsistent library dimensions")
	}
	result := make([][][]int, len(sizes))
	for i, n := range sizes {
		result[i] = Sfm_TruthToCnf(truths[i], nil, n)
	}
	return result
}

// Sfm_LibPreprocess assigns gate indices and derives their local CNFs.
func Sfm_LibPreprocess(lib *mio.Mio_Library_t) ([]int, []uint64, [][][]int, []*mio.Mio_Gate_t) {
	var sizes []int
	var truths []uint64
	var gates []*mio.Mio_Gate_t
	for gate := mio.Mio_LibraryReadGates(lib); gate != nil; gate = mio.Mio_GateReadNext(gate) {
		mio.Mio_GateSetValue(gate, int32(len(gates)))
		sizes = append(sizes, int(mio.Mio_GateReadPinNum(gate)))
		truths = append(truths, mio.Mio_GateReadTruth(gate))
		gates = append(gates, gate)
	}
	return sizes, truths, Sfm_DecCreateCnf(sizes, truths), gates
}
