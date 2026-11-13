package fraig

import "testing"

func TestNodeClassificationAndOutputPolarity(t *testing.T) {
	constant := &Fraig_Node_t{Num: 0, NumPi: -1}
	input := &Fraig_Node_t{Num: 1, NumPi: 0}
	and := &Fraig_Node_t{Num: 2, NumPi: -1, p1: input, p2: constant}
	if Fraig_NodeIsConst(Fraig_Not(constant)) != 1 || Fraig_NodeIsVar(Fraig_Not(input)) != 1 || Fraig_NodeIsAnd(Fraig_Not(and)) != 1 {
		t.Fatal("classification ignored regular node")
	}
	if Fraig_NodeReadOne(and) != input || Fraig_NodeReadTwo(and) != constant {
		t.Fatal("incorrect fanin")
	}
	p := &Fraig_Man_t{vOutputs: Fraig_NodeVecAlloc(0)}
	Fraig_ManSetPo(p, Fraig_Not(and))
	if and.fNodePo != 1 || Fraig_ManReadOutputNum(p) != 1 || Fraig_ManReadOutputs(p)[0] != Fraig_Not(and) {
		t.Fatal("output phase or node mark lost")
	}
}

func TestNodeChoiceLinks(t *testing.T) {
	old, prior, next := &Fraig_Node_t{}, &Fraig_Node_t{}, &Fraig_Node_t{}
	old.pNextE = prior
	Fraig_NodeSetChoice(nil, old, next)
	if old.pNextE != next || next.pNextE != prior || next.pRepr != old {
		t.Fatal("choice links incorrect")
	}
}

// fraigApi.c fills all missing input indices and permits an output node to be
// registered more than once, including with opposite edge polarity.
func TestInputIndexExpansionAndDuplicateOutputs(t *testing.T) {
	p := Fraig_ManCreate(nil)
	defer Fraig_ManFree(p)
	if Fraig_ManReadIthVar(p, -1) != nil || Fraig_ManReadInputNum(p) != 0 {
		t.Fatal("negative input request changed manager")
	}
	fourth := Fraig_ManReadIthVar(p, 3)
	if fourth == nil || Fraig_ManReadInputNum(p) != 4 {
		t.Fatal("missing inputs were not allocated")
	}
	for i := int32(0); i < 4; i++ {
		if got := Fraig_ManReadIthVar(p, i); got != Fraig_ManReadInputs(p)[i] || got.NumPi != i {
			t.Fatalf("input %d: %+v", i, got)
		}
	}
	if Fraig_ManReadIthVar(p, 3) != fourth || Fraig_ManReadInputNum(p) != 4 {
		t.Fatal("existing input was recreated")
	}
	Fraig_ManSetPo(p, fourth)
	Fraig_ManSetPo(p, Fraig_Not(fourth))
	outputs := Fraig_ManReadOutputs(p)
	if len(outputs) != 2 || outputs[0] != fourth || outputs[1] != Fraig_Not(fourth) || fourth.fNodePo != 1 {
		t.Fatal("duplicate output or edge polarity lost")
	}
}
