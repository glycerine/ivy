package reo

// unitStorage owns stable backing arrays. Keeping slices prevents Go's garbage
// collector from losing chunks while the reordering engine links their nodes.
// nUnitsUsed follows REO's operation counter: bulk plane recycling deliberately
// does not decrement it (the C callers reset it between reordering operations).
type unitStorage struct {
	pUnitFreeList                           *reo_unit
	pMemChunks                              [][]reo_unit
	nMemChunks, nMemChunksAlloc, nUnitsUsed int32
}

func (s *unitStorage) get() *reo_unit {
	if s.pUnitFreeList == nil {
		s.grow()
	}
	u := s.pUnitFreeList
	s.pUnitFreeList = u.Next
	s.nUnitsUsed++
	return u
}

func (s *unitStorage) recycle(u *reo_unit) {
	u.Next = s.pUnitFreeList
	s.pUnitFreeList = u
	s.nUnitsUsed--
}

func (s *unitStorage) recyclePlane(plane *reo_plane) {
	if plane.pHead == nil {
		return
	}
	tail := plane.pHead
	for tail.Next != nil {
		tail = tail.Next
	}
	tail.Next = s.pUnitFreeList
	s.pUnitFreeList = plane.pHead
	*plane = reo_plane{}
}

func (s *unitStorage) stop() {
	// Clear the free-list root as well as chunk owners: no retained node points
	// into storage from a previous dispenser lifetime.
	s.pUnitFreeList = nil
	s.pMemChunks = nil
	s.nMemChunks = 0
	s.nUnitsUsed = 0
}

func addUnitToPlane(plane *reo_plane, u *reo_unit) {
	u.Next = plane.pHead
	plane.pHead = u
	plane.statsNodes++
}

func (s *unitStorage) grow() {
	if s.pUnitFreeList != nil {
		panic("reo: growing nonempty free list")
	}
	if s.nMemChunks >= s.nMemChunksAlloc {
		panic("reo: unit storage capacity exhausted")
	}
	chunk := make([]reo_unit, unitChunkSize)
	for i := 0; i+1 < len(chunk); i++ {
		chunk[i].Next = &chunk[i+1]
	}
	s.pMemChunks = append(s.pMemChunks, chunk)
	s.pUnitFreeList = &chunk[0]
	s.nMemChunks++
}

// REO_CHUNK_SIZE from reo.h. Separate from draft constants so this component
// can be exercised independently while dependent CUDD code is being ported.
const unitChunkSize = 5000
