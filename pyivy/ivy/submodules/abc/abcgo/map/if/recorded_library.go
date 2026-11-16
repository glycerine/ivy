package if_

import "sync"

// RecordedLibrary connects the source ABC recording backend without an import
// cycle: ABC owns the library, while its delay and reconstruction routines use IF.
type RecordedLibrary struct {
	DelayCost func(*If_Man_t, *If_Cut_t, *If_Obj_t) int32
	BuildGia  func(*Gia_Man_t, *If_Man_t, *If_Cut_t, *Vec_Int_t, int32) int32
}

var ifRecordedLibrary struct {
	sync.RWMutex
	backend RecordedLibrary
}

// RegisterRecordedLibrary installs the source library cost and GIA builders.
func RegisterRecordedLibrary(backend RecordedLibrary) {
	ifRecordedLibrary.Lock()
	ifRecordedLibrary.backend = backend
	ifRecordedLibrary.Unlock()
}

func recordedLibraryBackend() RecordedLibrary {
	ifRecordedLibrary.RLock()
	backend := ifRecordedLibrary.backend
	ifRecordedLibrary.RUnlock()
	return backend
}

func If_CutDelayRecCost3(p *If_Man_t, cut *If_Cut_t, obj *If_Obj_t) int32 {
	cost := recordedLibraryBackend().DelayCost
	if cost == nil {
		panic("if: recorded-library delay backend is not registered")
	}
	return cost(p, cut, obj)
}
