package goivy

import (
	"strings"
	"testing"
)

// Python check_with_parameters (ivy_isolate.py): present is an ordered
// dict (present names, then verified), and the first unknown name raises
// IvyError(isolate, "{} is not an object, action, sort, definition,
// interpreted function or property"). With "with zz, aa" the error names
// zz, located at the isolate. goivy iterated a Go map, without location.
func TestCheckWithParametersReportsFirstUnknownName(t *testing.T) {
	for i := 0; i < 30; i++ {
		m := mkModuleWithSig()
		cfg := NewAstConfig()
		iso := cfg.NewIsolateDef([]Node{cfg.NewAtom("iso"), cfg.NewAtom("zz"), cfg.NewAtom("aa")}, 2)
		iso.SetLineno(Location{Filename: "f.ivy", Line: 7})
		m.Isolates.Set("iso", iso)
		err := CheckWithParameters(m, "iso")
		if err == nil || !strings.HasPrefix(err.Error(), "f.ivy: line 7: error: zz is not an object") {
			t.Fatalf("run %d: got %v, want zz first at the isolate's location", i, err)
		}
	}
}
