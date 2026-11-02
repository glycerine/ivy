package abc

// Hand-checked skeleton of src/bdd/llb/llb2Image.c.

// Llb_ImgSupports
// C: Vec_Ptr_t * Llb_ImgSupports( Aig_Man_t * p, Vec_Ptr_t * vDdMans, Vec_Int_t * vStart, Vec_Int_t * vStop, int fAddPis, int fVerbose )
func (p *AigManager) LlbImgSupports(vDdMans *PointerVector[CuddFoundation], vStart *IntVector, vStop *IntVector, fAddPis int, fVerbose int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Llb_ImgSchedule
// C: void Llb_ImgSchedule( Vec_Ptr_t * vSupps, Vec_Ptr_t ** pvQuant0, Vec_Ptr_t ** pvQuant1, int fVerbose )
func LlbImgSchedule(vSupps *PointerVector[IntVector], pvQuant0 **PointerVector[IntVector], pvQuant1 **PointerVector[IntVector], fVerbose int) {
	panic("TODO, stubbed")
}

// Llb_ImgPartition
// C: DdManager * Llb_ImgPartition( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper, abctime TimeTarget )
func (p *AigManager) LlbImgPartition(vLower *PointerVector[AigObject], vUpper *PointerVector[AigObject], TimeTarget int64) *CuddFoundation {
	panic("TODO, stubbed")
}

// Llb_ImgComputeCube
// C: DdNode * Llb_ImgComputeCube( Aig_Man_t * pAig, Vec_Int_t * vNodeIds, DdManager * dd )
func (pAig *AigManager) LlbImgComputeCube(vNodeIds *IntVector, dd *CuddFoundation) *DdNode {
	panic("TODO, stubbed")
}

// Llb_ImgQuantifyFirst
// C: void Llb_ImgQuantifyFirst( Aig_Man_t * pAig, Vec_Ptr_t * vDdMans, Vec_Ptr_t * vQuant0, int fVerbose )
func (pAig *AigManager) LlbImgQuantifyFirst(vDdMans *PointerVector[CuddFoundation], vQuant0 *PointerVector[IntVector], fVerbose int) {
	panic("TODO, stubbed")
}

// Llb_ImgQuantifyReset
// C: void Llb_ImgQuantifyReset( Vec_Ptr_t * vDdMans )
func LlbImgQuantifyReset(vDdMans *PointerVector[CuddFoundation]) {
	panic("TODO, stubbed")
}

// Llb_ImgComputeImage
// C: DdNode * Llb_ImgComputeImage( Aig_Man_t * pAig, Vec_Ptr_t * vDdMans, DdManager * dd, DdNode * bInit, Vec_Ptr_t * vQuant0, Vec_Ptr_t * vQuant1, Vec_Int_t * vDriRefs, abctime TimeTarget, int fBackward, int fReorder, int fVerbose )
func (pAig *AigManager) LlbImgComputeImage(vDdMans *PointerVector[CuddFoundation], dd *CuddFoundation, bInit *DdNode, vQuant0 *PointerVector[IntVector], vQuant1 *PointerVector[IntVector], vDriRefs *IntVector, TimeTarget int64, fBackward int, fReorder int, fVerbose int) *DdNode {
	panic("TODO, stubbed")
}
