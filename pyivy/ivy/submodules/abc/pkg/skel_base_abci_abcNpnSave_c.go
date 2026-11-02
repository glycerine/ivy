package abc

// Hand-checked skeleton of src/base/abci/abcNpnSave.c.

// NpnObj is Npn_Obj_t.
type NpnObj struct {
	uTruth uint64 // word : truth table
	Count  int    // occurrences
	iNext  int    // next entry
}

// NpnMan is Npn_Man_t.
type NpnMan struct {
	pBuffer     []NpnObj // Npn_Obj_t * : all NPN entries
	pBins       []int    // int * : hash table
	nBins       int      // hash table size
	nBufferSize int      // buffer size
	nEntries    int      // entry count
}

// Npn_ManObj (static)
// C: Npn_Obj_t * Npn_ManObj( Npn_Man_t * p, int i )
func (p *NpnMan) npnManObj(i int) *NpnObj {
	panic("TODO, stubbed")
}

// Npn_ManObjNum (static)
// C: int Npn_ManObjNum( Npn_Man_t * p, Npn_Obj_t * pObj )
func (p *NpnMan) npnManObjNum(pObj *NpnObj) int {
	panic("TODO, stubbed")
}

// Npn_TruthPermute_rec
// C: void Npn_TruthPermute_rec( char * pStr, int mid, int end )
func NpnTruthPermute_rec(pStr []byte, mid int, end int) {
	panic("TODO, stubbed")
}

// Npn_TruthHasVar (static)
// C: int Npn_TruthHasVar( word t, int v )
func npnTruthHasVar(t uint64, v int) int {
	panic("TODO, stubbed")
}

// Npn_TruthSupport (static)
// C: int Npn_TruthSupport( word t )
func npnTruthSupport(t uint64) int {
	panic("TODO, stubbed")
}

// Npn_TruthSuppSize (static)
// C: int Npn_TruthSuppSize( word t, int nVars )
func npnTruthSuppSize(t uint64, nVars int) int {
	panic("TODO, stubbed")
}

// Npn_TruthIsMinBase (static)
// C: int Npn_TruthIsMinBase( word t )
func npnTruthIsMinBase(t uint64) int {
	panic("TODO, stubbed")
}

// Npn_TruthPadWord
// C: word Npn_TruthPadWord( word uTruth, int nVars )
func NpnTruthPadWord(uTruth uint64, nVars int) uint64 {
	panic("TODO, stubbed")
}

// Npn_TruthCountOnes (static)
// C: int Npn_TruthCountOnes( word t )
func npnTruthCountOnes(t uint64) int {
	panic("TODO, stubbed")
}

// Npn_TruthChangePhase (static)
// C: word Npn_TruthChangePhase( word t, int v )
func npnTruthChangePhase(t uint64, v int) uint64 {
	panic("TODO, stubbed")
}

// Npn_TruthSwapAdjacentVars (static)
// C: word Npn_TruthSwapAdjacentVars( word t, int v )
func npnTruthSwapAdjacentVars(t uint64, v int) uint64 {
	panic("TODO, stubbed")
}

// Npn_TruthCanon (static)
// C: word Npn_TruthCanon( word t, int nVars, int * pPhase )
func npnTruthCanon(t uint64, nVars int, pPhase *int) uint64 {
	panic("TODO, stubbed")
}

// Npn_ManHash (static)
// C: int Npn_ManHash( Npn_Man_t * p, word uTruth )
func (p *NpnMan) npnManHash(uTruth uint64) int {
	panic("TODO, stubbed")
}

// Npn_ManResize
// C: void Npn_ManResize( Npn_Man_t * p )
func (p *NpnMan) NpnManResize() {
	panic("TODO, stubbed")
}

// Npn_ManAdd
// C: Npn_Obj_t * Npn_ManAdd( Npn_Man_t * p, word uTruth )
func (p *NpnMan) NpnManAdd(uTruth uint64) *NpnObj {
	panic("TODO, stubbed")
}

// Npn_ManRead
// C: void Npn_ManRead( Npn_Man_t * p, char * pFileName )
func (p *NpnMan) NpnManRead(pFileName string) {
	panic("TODO, stubbed")
}

// Npn_ManCompareEntries (static)
// C: int Npn_ManCompareEntries( Npn_Obj_t ** pp1, Npn_Obj_t ** pp2 )
func npnManCompareEntries(pp1 **NpnObj, pp2 **NpnObj) int {
	panic("TODO, stubbed")
}

// Npn_ManWrite
// C: void Npn_ManWrite( Npn_Man_t * p, char * pFileName )
func (p *NpnMan) NpnManWrite(pFileName string) {
	panic("TODO, stubbed")
}

// Npn_ManStart
// C: Npn_Man_t * Npn_ManStart( char * pFileName )
func NpnManStart(pFileName string) *NpnMan {
	panic("TODO, stubbed")
}

// Npn_ManStop
// C: void Npn_ManStop( Npn_Man_t * p )
func (p *NpnMan) NpnManStop() {
	panic("TODO, stubbed")
}

// Npn_ManClean
// C: void Npn_ManClean()
func NpnManClean() {
	panic("TODO, stubbed")
}

// Npn_ManLoad
// C: void Npn_ManLoad( char * pFileName )
func NpnManLoad(pFileName string) {
	panic("TODO, stubbed")
}

// Npn_ManSave
// C: void Npn_ManSave( char * pFileName )
func NpnManSave(pFileName string) {
	panic("TODO, stubbed")
}

// Npn_ManSaveOne
// C: void Npn_ManSaveOne( unsigned * puTruth, int nVars )
func NpnManSaveOne(puTruth []uint32, nVars int) {
	panic("TODO, stubbed")
}
