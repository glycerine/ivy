package if_

import "testing"

func TestAbciParameterAccessorsPreserveValues(t *testing.T) {
	var pars If_Par_t_
	pars.SetLUTSize(9)
	if got := pars.LUTSize(); got != 9 {
		t.Fatalf("LUTSize() = %d, want 9", got)
	}

	pars.SetCutMinFlag(2)
	if got := pars.CutMinFlag(); got != 2 {
		t.Fatalf("CutMinFlag() = %d, want 2", got)
	}
	pars.SetVerboseFlag(-3)
	if got := pars.VerboseFlag(); got != -3 {
		t.Fatalf("VerboseFlag() = %d, want -3", got)
	}
	pars.SetExpRedFlag(4)
	if got := pars.ExpRedFlag(); got != 4 {
		t.Fatalf("ExpRedFlag() = %d, want 4", got)
	}

	lib := &If_LibLut_t{}
	pars.SetLUTLibrary(lib)
	if got := pars.LUTLibrary(); got != lib {
		t.Fatalf("LUTLibrary() = %p, want %p", got, lib)
	}
	pars.SetLUTStructure("6:2")
	if got := pars.LUTStructure(); got != "6:2" {
		t.Fatalf("LUTStructure() = %q, want %q", got, "6:2")
	}
}
