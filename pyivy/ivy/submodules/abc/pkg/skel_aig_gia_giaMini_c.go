package abc

// Hand-checked skeleton of src/aig/gia/giaMini.c.

// Gia_ObjFromMiniFanin0Copy
// C: int Gia_ObjFromMiniFanin0Copy( Gia_Man_t * pGia, Vec_Int_t * vCopies, Mini_Aig_t * p, int Id )
func (pGia *GiaManager) GiaObjFromMiniFanin0Copy(vCopies *IntVector, p *MiniAIG, Id int) int {
	panic("TODO, stubbed")
}

// Gia_ObjFromMiniFanin1Copy
// C: int Gia_ObjFromMiniFanin1Copy( Gia_Man_t * pGia, Vec_Int_t * vCopies, Mini_Aig_t * p, int Id )
func (pGia *GiaManager) GiaObjFromMiniFanin1Copy(vCopies *IntVector, p *MiniAIG, Id int) int {
	panic("TODO, stubbed")
}

// Gia_ManFromMiniAig
// C: Gia_Man_t * Gia_ManFromMiniAig( Mini_Aig_t * p, Vec_Int_t ** pvCopies )
func GiaManFromMiniAig(p *MiniAIG, pvCopies **IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManToMiniAig
// C: Mini_Aig_t * Gia_ManToMiniAig( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManToMiniAig() *MiniAIG {
	panic("TODO, stubbed")
}

// Abc_FrameGiaInputMiniAig (p is a void * that holds a Mini_Aig_t *)
// C: void Abc_FrameGiaInputMiniAig( Abc_Frame_t * pAbc, void * p )
func (pAbc *MainFrame) AbcFrameGiaInputMiniAig(p *MiniAIG) {
	panic("TODO, stubbed")
}

// Abc_FrameGiaOutputMiniAig (returns a void * that holds a Mini_Aig_t *)
// C: void * Abc_FrameGiaOutputMiniAig( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) AbcFrameGiaOutputMiniAig() *MiniAIG {
	panic("TODO, stubbed")
}

// Gia_ManReadMiniAig
// C: Gia_Man_t * Gia_ManReadMiniAig( char * pFileName )
func GiaManReadMiniAig(pFileName string) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManWriteMiniAig
// C: void Gia_ManWriteMiniAig( Gia_Man_t * pGia, char * pFileName )
func (pGia *GiaManager) GiaManWriteMiniAig(pFileName string) {
	panic("TODO, stubbed")
}

// Gia_ManFromMiniLut
// C: Gia_Man_t * Gia_ManFromMiniLut( Mini_Lut_t * p, Vec_Int_t ** pvCopies )
func GiaManFromMiniLut(p *MiniLUT, pvCopies **IntVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFindComplLuts
// C: Vec_Bit_t * Gia_ManFindComplLuts( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManFindComplLuts() *BitVector {
	panic("TODO, stubbed")
}

// Gia_ManToMiniLut
// C: Mini_Lut_t * Gia_ManToMiniLut( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaManToMiniLut() *MiniLUT {
	panic("TODO, stubbed")
}

// Gia_ManToMiniLutAttr (pMiniLut is a void * that holds a Mini_Lut_t *)
// C: char * Gia_ManToMiniLutAttr( Gia_Man_t * pGia, void * pMiniLut )
func (pGia *GiaManager) GiaManToMiniLutAttr(pMiniLut *MiniLUT) string {
	panic("TODO, stubbed")
}

// Abc_FrameGiaInputMiniLut (p is a void * that holds a Mini_Lut_t *)
// C: void Abc_FrameGiaInputMiniLut( Abc_Frame_t * pAbc, void * p )
func (pAbc *MainFrame) AbcFrameGiaInputMiniLut(p *MiniLUT) {
	panic("TODO, stubbed")
}

// Abc_FrameGiaOutputMiniLut (returns a void * that holds a Mini_Lut_t *)
// C: void * Abc_FrameGiaOutputMiniLut( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) AbcFrameGiaOutputMiniLut() *MiniLUT {
	panic("TODO, stubbed")
}

// Abc_FrameGiaOutputMiniLutAttr (pMiniLut is a void * that holds a Mini_Lut_t *)
// C: char * Abc_FrameGiaOutputMiniLutAttr( Abc_Frame_t * pAbc, void * pMiniLut )
func (pAbc *MainFrame) AbcFrameGiaOutputMiniLutAttr(pMiniLut *MiniLUT) string {
	panic("TODO, stubbed")
}

// Gia_ManReadMiniLut
// C: Gia_Man_t * Gia_ManReadMiniLut( char * pFileName )
func GiaManReadMiniLut(pFileName string) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManWriteMiniLut
// C: void Gia_ManWriteMiniLut( Gia_Man_t * pGia, char * pFileName )
func (pGia *GiaManager) GiaManWriteMiniLut(pFileName string) {
	panic("TODO, stubbed")
}

// Gia_ManMapMiniLut2MiniAig
// C: int * Gia_ManMapMiniLut2MiniAig( Gia_Man_t * p, Gia_Man_t * p1, Gia_Man_t * p2, Vec_Int_t * vMap1, Vec_Int_t * vMap2 )
func (p *GiaManager) GiaManMapMiniLut2MiniAig(p1 *GiaManager, p2 *GiaManager, vMap1 *IntVector, vMap2 *IntVector) []int {
	panic("TODO, stubbed")
}

// Gia_ManNameMapVerify
// C: void Gia_ManNameMapVerify( Gia_Man_t * p, Gia_Man_t * p1, Gia_Man_t * p2, Vec_Int_t * vMap1, Vec_Int_t * vMap2, int * pMap )
func (p *GiaManager) GiaManNameMapVerify(p1 *GiaManager, p2 *GiaManager, vMap1 *IntVector, vMap2 *IntVector, pMap []int) {
	panic("TODO, stubbed")
}

// Abc_FrameReadMiniLutNameMapping
// C: int * Abc_FrameReadMiniLutNameMapping( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) AbcFrameReadMiniLutNameMapping() []int {
	panic("TODO, stubbed")
}

// Gia_ManMapEquivAfterScorr
// C: Vec_Int_t * Gia_ManMapEquivAfterScorr( Gia_Man_t * p, Vec_Int_t * vMap )
func (p *GiaManager) GiaManMapEquivAfterScorr(vMap *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Abc_FrameReadMiniAigEquivClasses
// C: int * Abc_FrameReadMiniAigEquivClasses( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) AbcFrameReadMiniAigEquivClasses() []int {
	panic("TODO, stubbed")
}

// Gia_MiniAigReduce
// C: Gia_Man_t * Gia_MiniAigReduce( Mini_Aig_t * p, int * pEquivs )
func GiaMiniAigReduce(p *MiniAIG, pEquivs []int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_MiniAigMiter
// C: Gia_Man_t * Gia_MiniAigMiter( Mini_Aig_t * p, int * pEquivs )
func GiaMiniAigMiter(p *MiniAIG, pEquivs []int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_MiniAigVerify
// C: void Gia_MiniAigVerify( Abc_Frame_t * pAbc, char * pFileName )
func (pAbc *MainFrame) GiaMiniAigVerify(pFileName string) {
	panic("TODO, stubbed")
}

// Gia_MiniAigSuperGates_rec
// C: void Gia_MiniAigSuperGates_rec( Mini_Aig_t * p, int iObj, Vec_Int_t * vRes, Vec_Int_t * vMap )
func GiaMiniAigSuperGates_rec(p *MiniAIG, iObj int, vRes *IntVector, vMap *IntVector) {
	panic("TODO, stubbed")
}

// Gia_MiniAigSuperGates
// C: Vec_Wec_t * Gia_MiniAigSuperGates( Mini_Aig_t * p )
func GiaMiniAigSuperGates(p *MiniAIG) *IntLevelVector {
	panic("TODO, stubbed")
}

// Gia_MiniAigSuperPrintDouble
// C: void Gia_MiniAigSuperPrintDouble( Vec_Int_t * p, int nPis )
func GiaMiniAigSuperPrintDouble(p *IntVector, nPis int) {
	panic("TODO, stubbed")
}

// Gia_MiniAigSuperMerge
// C: int Gia_MiniAigSuperMerge( Vec_Int_t * p, int nPis )
func GiaMiniAigSuperMerge(p *IntVector, nPis int) int {
	panic("TODO, stubbed")
}

// Gia_MiniAigSuperPreprocess
// C: int Gia_MiniAigSuperPreprocess( Mini_Aig_t * p, Vec_Wec_t * vSuper, int nPis, int fVerbose )
func GiaMiniAigSuperPreprocess(p *MiniAIG, vSuper *IntLevelVector, nPis int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Gia_MiniAigSuperDeriveGia
// C: Gia_Man_t * Gia_MiniAigSuperDeriveGia( Vec_Wec_t * p, int nPis, int Multi )
func GiaMiniAigSuperDeriveGia(p *IntLevelVector, nPis int, Multi int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_MiniAigSuperDerive
// C: Gia_Man_t * Gia_MiniAigSuperDerive( char * pFileName, int fVerbose )
func GiaMiniAigSuperDerive(pFileName string, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_MiniAigProcessFile
// C: Vec_Int_t * Gia_MiniAigProcessFile()
func GiaMiniAigProcessFile() *IntVector {
	panic("TODO, stubbed")
}

// Gia_MiniAigGenerate_rec
// C: void Gia_MiniAigGenerate_rec( Mini_Aig_t * p, Vec_Int_t * vTriples, int iObj, Vec_Int_t * vDefs, Vec_Int_t * vMap )
func GiaMiniAigGenerate_rec(p *MiniAIG, vTriples *IntVector, iObj int, vDefs *IntVector, vMap *IntVector) {
	panic("TODO, stubbed")
}

// Gia_MiniAigGenerateFromFile
// C: void Gia_MiniAigGenerateFromFile()
func GiaMiniAigGenerateFromFile() {
	panic("TODO, stubbed")
}
