package fraig

import "testing"

func TestChoiceTraversalAndLevels(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	ab := Fraig_NodeAnd(p, a, b)
	bc := Fraig_NodeAnd(p, b, c)
	primary := Fraig_NodeAnd(p, ab, c)
	alternative := Fraig_NodeAnd(p, a, bc)
	Fraig_NodeSetChoice(p, primary, alternative)
	Fraig_ManSetPo(p, primary)
	if primary.pNextE != alternative || alternative.pRepr != primary {
		t.Fatal("choice link missing")
	}
	without := Fraig_CountNodes(p, 0)
	with := Fraig_CountNodes(p, 1)
	if with <= without {
		t.Fatalf("choice traversal %d <= %d", with, without)
	}
	Fraig_MappingSetChoiceLevels(p, 0)
	if primary.Level != alternative.Level {
		t.Fatal("choice levels differ")
	}
	Fraig_MappingSetChoiceLevels(p, 1)
	if primary.Level != alternative.Level {
		t.Fatal("maximum choice levels differ")
	}
}
func TestFanoutRemovePivotAndLast(t *testing.T) {
	a, b := &Fraig_Node_t{}, &Fraig_Node_t{}
	first := &Fraig_Node_t{p1: a, p2: b}
	second := &Fraig_Node_t{p1: a, p2: b}
	Fraig_NodeAddFaninFanout(a, first)
	Fraig_NodeAddFaninFanout(a, second)
	Fraig_NodeRemoveFaninFanout(a, first)
	if a.pFanPivot != second || Fraig_NodeGetFanoutNum(a) != 1 {
		t.Fatal("pivot removal lost successor")
	}
	Fraig_NodeRemoveFaninFanout(a, second)
	if a.pFanPivot != nil || Fraig_NodeGetFanoutNum(a) != 0 {
		t.Fatal("last fanout removal failed")
	}
}
