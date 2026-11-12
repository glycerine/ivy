package gia

import "sync"

// Component services live in packages that depend on GIA and register their
// implementations when those packages are loaded.
var giaComponentServices struct {
	sync.RWMutex
	recording     func() bool
	computeAdders func(*Gia_Man_t, **Vec_Int_t, int32) *Vec_Int_t
	showFile      func(string)
}

func RegisterRecordingState(f func() bool) {
	giaComponentServices.Lock()
	giaComponentServices.recording = f
	giaComponentServices.Unlock()
}

func giaRecordingActive() bool {
	giaComponentServices.RLock()
	f := giaComponentServices.recording
	giaComponentServices.RUnlock()
	return f != nil && f()
}

func RegisterAdderRecognizer(f func(*Gia_Man_t, **Vec_Int_t, int32) *Vec_Int_t) {
	giaComponentServices.Lock()
	giaComponentServices.computeAdders = f
	giaComponentServices.Unlock()
}

func RegisterDotViewer(f func(string)) {
	giaComponentServices.Lock()
	giaComponentServices.showFile = f
	giaComponentServices.Unlock()
}

func giaShowServices(adders bool) (func(*Gia_Man_t, **Vec_Int_t, int32) *Vec_Int_t, func(string)) {
	giaComponentServices.RLock()
	compute, show := giaComponentServices.computeAdders, giaComponentServices.showFile
	giaComponentServices.RUnlock()
	if show == nil {
		panic("gia: DOT viewer is not registered")
	}
	if adders && compute == nil {
		panic("gia: adder recognizer is not registered")
	}
	return compute, show
}
