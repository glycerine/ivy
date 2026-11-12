package gia

import "testing"

func TestComponentServiceRegistration(t *testing.T) {
	giaComponentServices.RLock()
	recording, compute, show := giaComponentServices.recording, giaComponentServices.computeAdders, giaComponentServices.showFile
	giaComponentServices.RUnlock()
	defer func() { RegisterRecordingState(recording); RegisterAdderRecognizer(compute); RegisterDotViewer(show) }()
	RegisterRecordingState(nil)
	if giaRecordingActive() {
		t.Fatal("unregistered recording library is active")
	}
	RegisterRecordingState(func() bool { return true })
	if !giaRecordingActive() {
		t.Fatal("recording state provider was not used")
	}
	called := false
	RegisterDotViewer(func(string) { called = true })
	RegisterAdderRecognizer(func(*Gia_Man_t, **Vec_Int_t, int32) *Vec_Int_t { return Vec_IntAlloc(0) })
	gotCompute, gotShow := giaShowServices(true)
	if gotCompute(nil, nil, 0) == nil {
		t.Fatal("adder recognizer not used")
	}
	gotShow("example.dot")
	if !called {
		t.Fatal("DOT viewer not used")
	}
}
