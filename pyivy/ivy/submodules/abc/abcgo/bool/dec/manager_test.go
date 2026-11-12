package dec

import "testing"

func TestManagerTables(t *testing.T) {
	m := Dec_ManStart()
	if len(m.puCanons) != 65536 || len(m.pPhases) != 65536 || len(m.pPerms) != 65536 || len(m.pMap) != 65536 {
		t.Fatal("missing tables")
	}
	classes := map[byte]bool{}
	for _, c := range m.pMap {
		classes[c] = true
	}
	if len(classes) != 222 {
		t.Fatal("classes", len(classes))
	}
	Dec_ManStop(m)
	if m.pMvcMem != nil || m.puCanons != nil || m.vCubes != nil {
		t.Fatal("manager retained storage")
	}
	Dec_ManStop(nil)
}
