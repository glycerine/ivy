package fraig

import "testing"

func TestAddChoicesPreservesOriginalOutput(t *testing.T) {
	m := &Fraig_Man_t{nWordsRand: 1, nWordsDyna: 1, vNodes: Fraig_NodeVecAlloc(8), vInputs: Fraig_NodeVecAlloc(8), vOutputs: Fraig_NodeVecAlloc(1), pTableS: Fraig_HashTableCreate(4)}
	m.pConst1 = &Fraig_Node_t{Num: 0, NumPi: -1}
	Fraig_NodeVecPush(m.vNodes, m.pConst1)
	for i := int32(0); i < 3; i++ {
		n := &Fraig_Node_t{Num: i + 1, NumPi: i, puSimR: []uint32{uint32(1) << i}, puSimD: []uint32{0}}
		Fraig_NodeVecPush(m.vNodes, n)
		Fraig_NodeVecPush(m.vInputs, n)
	}
	a, b, c := m.vInputs.pArray[0], m.vInputs.pArray[1], m.vInputs.pArray[2]
	root := Fraig_NodeAnd(m, Fraig_NodeAnd(m, a, b), c)
	Fraig_NodeVecPush(m.vOutputs, root)
	before := Fraig_ManReadNodeNum(m)
	Fraig_ManAddChoices(m, 0, 10)
	if m.vOutputs.pArray[0] != root || Fraig_ManReadNodeNum(m) <= before {
		t.Fatal("associative choices missing or output changed")
	}
}
