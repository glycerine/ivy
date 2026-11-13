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
