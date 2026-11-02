package abc

// Hand-checked skeleton of src/phys/place/libhmetis.h.

// HMETIS_PartRecursive (static, empty body in the C source; edgecnt is the hMETIS scalar edge-cut output)
// C: void HMETIS_PartRecursive(int nvtxs, int nhedges, int *vwgts, int *eptr, int *eind, int *hewgts, int nparts, int nbfactor, int *options, int *part, int *edgecnt )
func hmetis_PartRecursive(nvtxs int, nhedges int, vwgts []int, eptr []int, eind []int, hewgts []int, nparts int, nbfactor int, options []int, part []int, edgecnt *int) {
	panic("TODO, stubbed")
}

// HMETIS_PartKway (static, empty body in the C source; edgecnt is the hMETIS scalar edge-cut output)
// C: void HMETIS_PartKway(int nvtxs, int nhedges, int *vwgts, int *eptr, int *eind, int *hewgts, int nparts, int nbfactor, int *options, int *part, int *edgecnt )
func hmetis_PartKway(nvtxs int, nhedges int, vwgts []int, eptr []int, eind []int, hewgts []int, nparts int, nbfactor int, options []int, part []int, edgecnt *int) {
	panic("TODO, stubbed")
}
