package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
	"testing"
)

func TestSemiFormalManagerRetainsSourceAssertion(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(r.(string), "unfinished") {
			t.Fatalf("unexpected panic %v", r)
		}
	}()
	Ssw_SemManStart(nil, 100, 0)
}
func TestSemiFormalPatternCapture(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 2)
	p := &Ssw_Sem_t{pMan: &Ssw_Man_t{pAig: g, pPatWords: []uint32{2}}, nPatternsAlloc: 32, nPatterns: 1, vPatterns: vec.Vec_PtrAllocSimInfo(2, 1)}
	Ssw_ManFilterBmcSavePattern(p)
	if p.nPatterns != 2 || vec.Vec_PtrEntry(p.vPatterns, 0).([]uint32)[0] != 2 || vec.Vec_PtrEntry(p.vPatterns, 1).([]uint32)[0] != 0 {
		t.Fatal("bad saved pattern")
	}
}
