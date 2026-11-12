package abcgo

import "testing"

func TestTemplateBorrowsHandlesAndReturnsOriginalResult(t *testing.T) {
	aig := &struct{ value int }{7}
	parameters := &Xyz_ParTry_t{Par: 42}
	manager := Xyz_ManTryAlloc(aig, parameters)
	if manager.pAig != aig || manager.pPar != parameters {
		t.Fatal("borrowed identity changed")
	}
	Xyz_ManTryFree(manager)
	Xyz_ManTryFree(nil)
	if Xyz_ManPerform(aig, parameters) != 1 || Xyz_ManPerform(nil, nil) != 1 {
		t.Fatal("source result changed")
	}
	if aig.value != 7 || parameters.Par != 42 {
		t.Fatal("borrowed data modified")
	}
}
