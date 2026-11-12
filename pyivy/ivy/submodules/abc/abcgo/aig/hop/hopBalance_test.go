package hop

import "testing"

func balanceEvaluate(p *Hop_Man_t, edge *Hop_Obj_t, assignment int) bool {
	values := map[*Hop_Obj_t]bool{p.pConst1: true}
	for i := int32(0); i < Vec_PtrSize(p.vPis); i++ {
		values[Vec_PtrArray(p.vPis)[i].(*Hop_Obj_t)] = assignment&(1<<uint(i)) != 0
	}
	var eval func(*Hop_Obj_t) bool
	eval = func(edge *Hop_Obj_t) bool {
		node := Hop_Regular(edge)
		value, known := values[node]
		if !known {
			switch node.Type {
			case AIG_AND:
				value = eval(node.pFanin0) && eval(node.pFanin1)
			case AIG_EXOR:
				value = eval(node.pFanin0) != eval(node.pFanin1)
			case AIG_PO:
				value = eval(node.pFanin0)
			default:
				tPanicUnexpectedBalanceType(node.Type)
			}
			values[node] = value
		}
		return value != (Hop_IsComplement(edge) != 0)
	}
	return eval(edge)
}
func tPanicUnexpectedBalanceType(typ Hop_Type_t) { panic("unexpected Hop type in balance evaluator") }
func TestHopBalanceChain(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	chain := Hop_IthVar(p, 0)
	for i := int32(1); i < 8; i++ {
		chain = Hop_And(p, chain, Hop_IthVar(p, i))
	}
	Hop_ObjCreatePo(p, chain)
	Hop_ObjCreatePo(p, Hop_Not(chain))
	for _, update := range []int32{0, 1} {
		balanced := Hop_ManBalance(p, update)
		if Vec_PtrSize(balanced.vPis) != 8 || Vec_PtrSize(balanced.vPos) != 2 || Hop_ManCheck(balanced) == 0 {
			t.Fatal("invalid balanced manager")
		}
		if level := Hop_ObjLevel(Hop_ObjFanin0(Vec_PtrArray(balanced.vPos)[0].(*Hop_Obj_t))); level != 3 {
			t.Fatalf("balanced depth %d, want 3", level)
		}
		for assignment := 0; assignment < 256; assignment++ {
			for output := int32(0); output < 2; output++ {
				old := Vec_PtrArray(p.vPos)[output].(*Hop_Obj_t)
				next := Vec_PtrArray(balanced.vPos)[output].(*Hop_Obj_t)
				if balanceEvaluate(p, old, assignment) != balanceEvaluate(balanced, next, assignment) {
					t.Fatalf("assignment %d output %d", assignment, output)
				}
			}
		}
		Hop_ManStop(balanced)
	}
}
func TestHopBalanceContradictorySupergate(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	node := Hop_And(p, Hop_And(p, a, b), Hop_And(p, Hop_Not(a), c))
	Hop_ObjCreatePo(p, node)
	q := Hop_ManBalance(p, 1)
	defer Hop_ManStop(q)
	if Hop_ObjChild0(Vec_PtrArray(q.vPos)[0].(*Hop_Obj_t)) != Hop_ManConst0(q) {
		t.Fatal("contradictory supergate was not simplified")
	}
	for _, node := range []*Hop_Obj_t{a, b, c, Hop_Regular(node)} {
		if node.fMarkB != 0 {
			t.Fatal("leaf mark leaked")
		}
	}
}
func TestHopBalanceSharedAndConstantOutputs(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	shared := Hop_And(p, a, b)
	Hop_ObjCreatePo(p, Hop_And(p, shared, c))
	Hop_ObjCreatePo(p, Hop_Or(p, shared, Hop_Not(c)))
	Hop_ObjCreatePo(p, Hop_ManConst0(p))
	Hop_ObjCreatePo(p, Hop_ManConst1(p))
	q := Hop_ManBalance(p, 1)
	defer Hop_ManStop(q)
	for assignment := 0; assignment < 8; assignment++ {
		for output := int32(0); output < 4; output++ {
			if balanceEvaluate(p, Vec_PtrArray(p.vPos)[output].(*Hop_Obj_t), assignment) != balanceEvaluate(q, Vec_PtrArray(q.vPos)[output].(*Hop_Obj_t), assignment) {
				t.Fatal(assignment, output)
			}
		}
	}
}
func TestHopBalanceReusesHashPair(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	p.fRefCount = 0
	a, b, c := Hop_IthVar(p, 0), Hop_IthVar(p, 1), Hop_IthVar(p, 2)
	Hop_And(p, a, c)
	cone := Vec_PtrAlloc(3)
	Vec_PtrPush(cone, a)
	Vec_PtrPush(cone, b)
	Vec_PtrPush(cone, c)
	Hop_NodeBalancePermute(p, cone, 0, 0)
	if Vec_PtrArray(cone)[1].(*Hop_Obj_t) != a {
		t.Fatal("did not prioritize existing pair")
	}
	result := Hop_NodeBalanceBuildSuper(p, cone, AIG_AND, 1)
	if Hop_ObjLevel(Hop_Regular(result)) != 2 {
		t.Fatal("unexpected balanced level")
	}
}
