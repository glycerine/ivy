package bsat

import "testing"

func TestReconstructedUnsatCore(t *testing.T) {
	cnf := partitionedProof(t, [][]lit{{0, 2}, {1, 2}}, [][]lit{{3}, {4, 6}})
	p := Intp_ManAlloc()
	core := Intp_ManUnsatCore(p, cnf, 0, 0)
	if p.lastErr != nil || len(core) != 3 {
		t.Fatal(core, p.lastErr)
	}
	Intp_ManUnsatCoreVerify(cnf, core)
	for _, id := range core {
		if id == 3 {
			t.Fatal("irrelevant root included", core)
		}
	}
	learned := Intp_ManUnsatCore(p, cnf, 1, 0)
	if p.lastErr != nil || len(learned) == 0 || learned[len(learned)-1] != cnf.pEmpty.Id {
		t.Fatal(learned, p.lastErr)
	}
	for _, id := range learned {
		if id < cnf.nRoots {
			t.Fatal("root in learned-only core")
		}
	}
}
