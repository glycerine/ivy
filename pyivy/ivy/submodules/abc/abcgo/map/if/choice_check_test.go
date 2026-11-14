package if_

import "testing"

func TestDanglingChoices(t *testing.T) {
	var pars If_Par_t
	Gia_ManSetIfParsDefault(&pars)
	pars.NLutSize = 4
	p := If_ManStart(&pars)
	defer If_ManStop(p)
	a, b, c := If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p)
	root := If_ManCreateAnd(p, a, b)
	alternative := If_ManCreateAnd(p, a, c)
	dangling := If_ManCreateAnd(p, b, c)
	root.pEquiv = alternative
	If_ManCreateCo(p, root)
	got := ifDanglingChoices(p)
	if len(got) != 1 || got[0] != dangling.Id {
		t.Fatalf("dangling IDs = %v, want [%d]", got, dangling.Id)
	}
	for i := int32(0); i < If_ManObjNum(p); i++ {
		if If_ManObj(p, i).fDriver != 0 {
			t.Fatalf("object %d retains a traversal mark", i)
		}
	}
	If_ManCreateCo(p, dangling)
	if got := ifDanglingChoices(p); len(got) != 0 {
		t.Fatalf("unexpected dangling nodes: %v", got)
	}
}
