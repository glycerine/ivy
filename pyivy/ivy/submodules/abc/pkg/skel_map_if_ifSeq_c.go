package abc

// Hand-checked skeleton of src/map/if/ifSeq.c.

// If_ManPrepareMappingSeq
// C: void If_ManPrepareMappingSeq( If_Man_t * p )
func (p *IfMan) IfManPrepareMappingSeq() {
	panic("TODO, stubbed")
}

// If_ManCollectLatches_rec
// C: void If_ManCollectLatches_rec( If_Obj_t * pObj, Vec_Ptr_t * vLatches )
func IfManCollectLatches_rec(pObj *IfObj, vLatches *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_ManCollectLatches
// C: Vec_Ptr_t * If_ManCollectLatches( If_Man_t * p )
func (p *IfMan) IfManCollectLatches() *PointerVector[IfObj] {
	panic("TODO, stubbed")
}

// If_ManPerformMappingRoundSeq
// C: int If_ManPerformMappingRoundSeq( If_Man_t * p, int nIter )
func (p *IfMan) IfManPerformMappingRoundSeq(nIter int) int {
	panic("TODO, stubbed")
}

// If_ManBinarySearchPeriod
// C: int If_ManBinarySearchPeriod( If_Man_t * p )
func (p *IfMan) IfManBinarySearchPeriod() int {
	panic("TODO, stubbed")
}

// If_ManBinarySearch_rec
// C: int If_ManBinarySearch_rec( If_Man_t * p, int FiMin, int FiMax )
func (p *IfMan) IfManBinarySearch_rec(FiMin int, FiMax int) int {
	panic("TODO, stubbed")
}

// If_ManPerformMappingSeqPost
// C: void If_ManPerformMappingSeqPost( If_Man_t * p )
func (p *IfMan) IfManPerformMappingSeqPost() {
	panic("TODO, stubbed")
}

// If_ManPerformMappingSeq
// C: int If_ManPerformMappingSeq( If_Man_t * p )
func (p *IfMan) IfManPerformMappingSeq() int {
	panic("TODO, stubbed")
}
