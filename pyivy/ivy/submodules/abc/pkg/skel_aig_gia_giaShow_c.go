package abc

// Hand-checked skeleton of src/aig/gia/giaShow.c.

// Gia_ShowPath
// C: void Gia_ShowPath( Gia_Man_t * p, char * pFileName )
func (p *GiaManager) GiaShowPath(pFileName string) {
	panic("TODO, stubbed")
}

// Gia_WriteDotAigSimple
// C: void Gia_WriteDotAigSimple( Gia_Man_t * p, char * pFileName, Vec_Int_t * vBold )
func (p *GiaManager) GiaWriteDotAigSimple(pFileName string, vBold *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ShowAddOut
// C: int Gia_ShowAddOut( Vec_Int_t * vAdds, Vec_Int_t * vMapAdds, int Node )
func GiaShowAddOut(vAdds *IntVector, vMapAdds *IntVector, Node int) int {
	panic("TODO, stubbed")
}

// Gia_WriteDotAig
// C: void Gia_WriteDotAig( Gia_Man_t * p, char * pFileName, Vec_Int_t * vBold, Vec_Int_t * vAdds, Vec_Int_t * vXors, Vec_Int_t * vMapAdds, Vec_Int_t * vMapXors, Vec_Int_t * vOrder )
func (p *GiaManager) GiaWriteDotAig(pFileName string, vBold *IntVector, vAdds *IntVector, vXors *IntVector, vMapAdds *IntVector, vMapXors *IntVector, vOrder *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ShowMapAdds
// C: Vec_Int_t * Gia_ShowMapAdds( Gia_Man_t * p, Vec_Int_t * vAdds, int fFadds, Vec_Int_t * vBold )
func (p *GiaManager) GiaShowMapAdds(vAdds *IntVector, fFadds int, vBold *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ShowMapXors
// C: Vec_Int_t * Gia_ShowMapXors( Gia_Man_t * p, Vec_Int_t * vXors )
func (p *GiaManager) GiaShowMapXors(vXors *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ShowCollectObjs_rec
// C: int Gia_ShowCollectObjs_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vAdds, Vec_Int_t * vXors, Vec_Int_t * vMapAdds, Vec_Int_t * vMapXors, Vec_Int_t * vOrder )
func (p *GiaManager) GiaShowCollectObjs_rec(pObj *GiaObject, vAdds *IntVector, vXors *IntVector, vMapAdds *IntVector, vMapXors *IntVector, vOrder *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ShowCollectObjs
// C: Vec_Int_t * Gia_ShowCollectObjs( Gia_Man_t * p, Vec_Int_t * vAdds, Vec_Int_t * vXors, Vec_Int_t * vMapAdds, Vec_Int_t * vMapXors )
func (p *GiaManager) GiaShowCollectObjs(vAdds *IntVector, vXors *IntVector, vMapAdds *IntVector, vMapXors *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ShowProcess
// C: void Gia_ShowProcess( Gia_Man_t * p, char * pFileName, Vec_Int_t * vBold, Vec_Int_t * vAdds, Vec_Int_t * vXors, int fFadds )
func (p *GiaManager) GiaShowProcess(pFileName string, vBold *IntVector, vAdds *IntVector, vXors *IntVector, fFadds int) {
	panic("TODO, stubbed")
}

// Gia_ManShow
// C: void Gia_ManShow( Gia_Man_t * pMan, Vec_Int_t * vBold, int fAdders, int fFadds, int fPath )
func (pMan *GiaManager) GiaManShow(vBold *IntVector, fAdders int, fFadds int, fPath int) {
	panic("TODO, stubbed")
}
