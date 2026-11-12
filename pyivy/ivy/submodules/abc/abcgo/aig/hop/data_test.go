package hop

import "testing"

func TestConversionScratchIsSharedByComplement(t *testing.T) {
	manager := Hop_ManStart()
	defer Hop_ManStop(manager)
	obj := Hop_IthVar(manager, 0)
	opposite := Hop_Not(obj)
	value := new(int)
	Hop_ObjSetData(opposite, value)
	Hop_ObjSetIData(obj, 42)
	if Hop_ObjData(obj) != value || Hop_ObjData(opposite) != value || Hop_ObjIData(opposite) != 42 {
		t.Fatal("complement conversion scratch diverged")
	}
}
