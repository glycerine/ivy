package abc

// Hand-checked skeleton of src/aig/gia/giaMem.c.

// GiaMmFixed is Gia_MmFixed_t.
type GiaMmFixed struct {
	nEntrySize    int      // the size of one entry
	nEntriesAlloc int      // the total number of entries allocated
	nEntriesUsed  int      // the number of entries in use
	nEntriesMax   int      // the max number of entries in use
	pEntriesFree  []byte   // char * : the linked list of free entries
	nChunkSize    int      // the size of one chunk
	nChunksAlloc  int      // the maximum number of memory chunks
	nChunks       int      // the current number of memory chunks
	pChunks       [][]byte // char ** : the allocated memory
	nMemoryUsed   int      // memory used in the allocated entries
	nMemoryAlloc  int      // memory allocated
}

// GiaMmFlex is Gia_MmFlex_t.
type GiaMmFlex struct {
	nEntriesUsed int      // the number of entries allocated
	pCurrent     []byte   // char * : the current pointer to free memory
	pEnd         []byte   // char * : the first entry outside the free memory
	nChunkSize   int      // the size of one chunk
	nChunksAlloc int      // the maximum number of memory chunks
	nChunks      int      // the current number of memory chunks
	pChunks      [][]byte // char ** : the allocated memory
	nMemoryUsed  int      // memory used in the allocated entries
	nMemoryAlloc int      // memory allocated
}

// GiaMmStep is Gia_MmStep_t.
type GiaMmStep struct {
	nMems        int           // the number of fixed memory managers employed
	pMems        []*GiaMmFixed // Gia_MmFixed_t ** : memory managers: 2^1 words, 2^2 words, etc
	nMapSize     int           // the size of the memory array
	pMap         []*GiaMmFixed // Gia_MmFixed_t ** : maps the number of bytes into its memory manager
	nChunksAlloc int           // the maximum number of memory chunks
	nChunks      int           // the current number of memory chunks
	pChunks      [][]byte      // char ** : the allocated memory
}

// Gia_MmFixedStart
// C: Gia_MmFixed_t * Gia_MmFixedStart( int nEntrySize, int nEntriesMax )
func GiaMmFixedStart(nEntrySize int, nEntriesMax int) *GiaMmFixed {
	panic("TODO, stubbed")
}

// Gia_MmFixedStop
// C: void Gia_MmFixedStop( Gia_MmFixed_t * p, int fVerbose )
func (p *GiaMmFixed) GiaMmFixedStop(fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_MmFixedEntryFetch
// C: char * Gia_MmFixedEntryFetch( Gia_MmFixed_t * p )
func (p *GiaMmFixed) GiaMmFixedEntryFetch() []byte {
	panic("TODO, stubbed")
}

// Gia_MmFixedEntryRecycle
// C: void Gia_MmFixedEntryRecycle( Gia_MmFixed_t * p, char * pEntry )
func (p *GiaMmFixed) GiaMmFixedEntryRecycle(pEntry []byte) {
	panic("TODO, stubbed")
}

// Gia_MmFixedRestart
// C: void Gia_MmFixedRestart( Gia_MmFixed_t * p )
func (p *GiaMmFixed) GiaMmFixedRestart() {
	panic("TODO, stubbed")
}

// Gia_MmFixedReadMemUsage
// C: int Gia_MmFixedReadMemUsage( Gia_MmFixed_t * p )
func (p *GiaMmFixed) GiaMmFixedReadMemUsage() int {
	panic("TODO, stubbed")
}

// Gia_MmFixedReadMaxEntriesUsed
// C: int Gia_MmFixedReadMaxEntriesUsed( Gia_MmFixed_t * p )
func (p *GiaMmFixed) GiaMmFixedReadMaxEntriesUsed() int {
	panic("TODO, stubbed")
}

// Gia_MmFlexStart
// C: Gia_MmFlex_t * Gia_MmFlexStart()
func GiaMmFlexStart() *GiaMmFlex {
	panic("TODO, stubbed")
}

// Gia_MmFlexStop
// C: void Gia_MmFlexStop( Gia_MmFlex_t * p, int fVerbose )
func (p *GiaMmFlex) GiaMmFlexStop(fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_MmFlexEntryFetch
// C: char * Gia_MmFlexEntryFetch( Gia_MmFlex_t * p, int nBytes )
func (p *GiaMmFlex) GiaMmFlexEntryFetch(nBytes int) []byte {
	panic("TODO, stubbed")
}

// Gia_MmFlexRestart
// C: void Gia_MmFlexRestart( Gia_MmFlex_t * p )
func (p *GiaMmFlex) GiaMmFlexRestart() {
	panic("TODO, stubbed")
}

// Gia_MmFlexReadMemUsage
// C: int Gia_MmFlexReadMemUsage( Gia_MmFlex_t * p )
func (p *GiaMmFlex) GiaMmFlexReadMemUsage() int {
	panic("TODO, stubbed")
}

// Gia_MmStepStart
// C: Gia_MmStep_t * Gia_MmStepStart( int nSteps )
func GiaMmStepStart(nSteps int) *GiaMmStep {
	panic("TODO, stubbed")
}

// Gia_MmStepStop
// C: void Gia_MmStepStop( Gia_MmStep_t * p, int fVerbose )
func (p *GiaMmStep) GiaMmStepStop(fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_MmStepEntryFetch
// C: char * Gia_MmStepEntryFetch( Gia_MmStep_t * p, int nBytes )
func (p *GiaMmStep) GiaMmStepEntryFetch(nBytes int) []byte {
	panic("TODO, stubbed")
}

// Gia_MmStepEntryRecycle
// C: void Gia_MmStepEntryRecycle( Gia_MmStep_t * p, char * pEntry, int nBytes )
func (p *GiaMmStep) GiaMmStepEntryRecycle(pEntry []byte, nBytes int) {
	panic("TODO, stubbed")
}

// Gia_MmStepReadMemUsage
// C: int Gia_MmStepReadMemUsage( Gia_MmStep_t * p )
func (p *GiaMmStep) GiaMmStepReadMemUsage() int {
	panic("TODO, stubbed")
}
