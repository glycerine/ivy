package abc

// Hand-checked skeleton of src/map/if/ifTest.c.

// ParThData is Par_ThData_t.
type ParThData struct {
	p      *GiaManager
	Id     int
	Status int
}

// Gia_ParTestObj (static inline)
// C: word * Gia_ParTestObj( Gia_Man_t * p, int Id )
func (p *GiaManager) giaParTestObj(Id int) []uint64 {
	panic("TODO, stubbed")
}

// Gia_ParTestAlloc (static inline)
// C: void Gia_ParTestAlloc( Gia_Man_t * p, int nWords )
func (p *GiaManager) giaParTestAlloc(nWords int) {
	panic("TODO, stubbed")
}

// Gia_ParTestFree (static inline)
// C: void Gia_ParTestFree( Gia_Man_t * p )
func (p *GiaManager) giaParTestFree() {
	panic("TODO, stubbed")
}

// Gia_ParComputeSignature
// C: void Gia_ParComputeSignature( Gia_Man_t * p, int nWords )
func (p *GiaManager) GiaParComputeSignature(nWords int) {
	panic("TODO, stubbed")
}

// Gia_ParTestSimulateInit
// C: void Gia_ParTestSimulateInit( Gia_Man_t * p )
func (p *GiaManager) GiaParTestSimulateInit() {
	panic("TODO, stubbed")
}

// Gia_ParTestSimulateObj
// C: void Gia_ParTestSimulateObj( Gia_Man_t * p, int Id )
func (p *GiaManager) GiaParTestSimulateObj(Id int) {
	panic("TODO, stubbed")
}

// Gia_ParTestSimulate
// C: void Gia_ParTestSimulate( Gia_Man_t * p, int nWords )
func (p *GiaManager) GiaParTestSimulate(nWords int) {
	panic("TODO, stubbed")
}

// Gia_ManCreateFaninCounts
// C: Vec_Int_t * Gia_ManCreateFaninCounts( Gia_Man_t * p )
func (p *GiaManager) GiaManCreateFaninCounts() *IntVector {
	panic("TODO, stubbed")
}

// Gia_ParWorkerThread (pArg is a void * that holds a Par_ThData_t *)
// C: void * Gia_ParWorkerThread( void * pArg )
func GiaParWorkerThread(pArg *ParThData) any {
	panic("TODO, stubbed")
}

// Gia_ParTestSimulate2
// C: void Gia_ParTestSimulate2( Gia_Man_t * p, int nWords, int nProcs )
func (p *GiaManager) GiaParTestSimulate2(nWords int, nProcs int) {
	panic("TODO, stubbed")
}

// Gia_ParTest
// C: void Gia_ParTest( Gia_Man_t * p, int nWords, int nProcs )
func (p *GiaManager) GiaParTest(nWords int, nProcs int) {
	panic("TODO, stubbed")
}
