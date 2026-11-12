package ivy

import "testing"

func ivyBalanceEvaluate(p *Ivy_Man_t, bits int) []bool {
	values := map[*Ivy_Obj_t]bool{p.pConst1: true}
	for i, n := range p.vPis {
		values[n] = bits&(1<<i) != 0
	}
	var eval func(*Ivy_Obj_t) bool
	eval = func(edge *Ivy_Obj_t) bool {
		n := Ivy_Regular(edge)
		value, known := values[n]
		if !known {
			a := eval(n.pFanin0)
			switch n.Type {
			case IVY_AND:
				value = a && eval(n.pFanin1)
			case IVY_EXOR:
				value = a != eval(n.pFanin1)
			case IVY_BUF, IVY_PO:
				value = a
			default:
				panic("unsupported evaluation node")
			}
			values[n] = value
		}
		if Ivy_IsComplement(edge) != 0 {
			return !value
		}
		return value
	}
	out := make([]bool, len(p.vPos))
	for i, n := range p.vPos {
		out[i] = eval(n)
	}
	return out
}
func TestBalanceSkewAndTruthAndDepth(t *testing.T) {
	for _, update := range []int32{0, 1} {
		p := Ivy_ManStart()
		inputs := make([]*Ivy_Obj_t, 8)
		for i := range inputs {
			inputs[i] = Ivy_ObjCreatePi(p)
		}
		root := inputs[0]
		for _, n := range inputs[1:] {
			root = Ivy_And(p, root, n)
		}
		Ivy_ObjCreatePo(p, root)
		Ivy_ObjCreatePo(p, Ivy_Not(root))
		Ivy_ObjCreatePo(p, p.pConst1)
		balanced := Ivy_ManBalance(p, update)
		if level := Ivy_Regular(Ivy_ObjChild0(balanced.vPos[0])).Level; level != 3 {
			t.Fatalf("balanced depth %d, want 3", level)
		}
		for bits := 0; bits < 256; bits++ {
			a, b := ivyBalanceEvaluate(p, bits), ivyBalanceEvaluate(balanced, bits)
			for i := range a {
				if a[i] != b[i] {
					t.Fatalf("truth differs at %d output %d", bits, i)
				}
			}
		}
		for _, n := range p.vObjs {
			if n != nil && n.fMarkB != 0 {
				t.Fatal("balance leaked marks")
			}
		}
	}
}
func TestBalanceNativeXorCancellation(t *testing.T) {
	for _, opposite := range []bool{false, true} {
		p := Ivy_ManStart()
		a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
		first := Ivy_CanonExor(p, a, b)
		second := Ivy_CanonExor(p, first, c)
		repeated := a
		if opposite {
			repeated = Ivy_Not(a)
		}
		root := Ivy_CanonExor(p, second, repeated)
		Ivy_ObjCreatePo(p, root)
		result := Ivy_ManBalance(p, 1)
		for bits := 0; bits < 8; bits++ {
			if ivyBalanceEvaluate(p, bits)[0] != ivyBalanceEvaluate(result, bits)[0] {
				t.Fatalf("XOR cancellation differs for opposite=%v bits=%d", opposite, bits)
			}
		}
	}
}
func TestBalanceContradictoryAnd(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	root := Ivy_And(p, Ivy_And(p, a, b), Ivy_Not(a))
	Ivy_ObjCreatePo(p, root)
	result := Ivy_ManBalance(p, 1)
	if Ivy_ObjChild0(result.vPos[0]) != Ivy_ManConst0(result) {
		t.Fatal("contradiction did not collapse")
	}
}
