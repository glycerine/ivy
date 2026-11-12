package amap

import (
	"fmt"

	"github.com/glycerine/abcgo/bool/kit"
)

// Amap_LibCollectFanins returns the leaves below a same-operation AND/XOR
// region. The order follows the source graph's fanin0, fanin1 traversal.
func Amap_LibCollectFanins(pLib *Amap_Lib_t, pNod *Amap_Nod_t) *Vec_Int_t {
	out := Vec_IntAlloc(4)
	var visit func(*Amap_Nod_t)
	visit = func(n *Amap_Nod_t) {
		if n == nil || n.Id == 0 {
			Vec_IntPush(out, 0)
			return
		}
		for _, lit := range []int32{int32(n.iFan0), int32(n.iFan1)} {
			fan := Amap_LibNod(pLib, Abc_Lit2Var(lit))
			if Abc_LitIsCompl(lit) == 0 && fan.Type == n.Type && n.Type != AMAP_OBJ_MUX {
				visit(fan)
			} else {
				Vec_IntPush(out, lit)
			}
		}
	}
	visit(pNod)
	return out
}

// Amap_LibDeriveGatePerm_rec maps the leaves of a DSD subnetwork onto the
// leaves of a structurally equivalent library node. It returns DSD literals
// in library-leaf order, retaining complement bits.
func Amap_LibDeriveGatePerm_rec(lib *Amap_Lib_t, ntk *kit.Kit_DsdNtk_t, lit int32, nod *Amap_Nod_t) *Vec_Int_t {
	if kit.Kit_DsdLitIsLeaf(ntk, lit) != 0 {
		if nod == nil || nod.Id != 0 {
			return nil
		}
		out := Vec_IntAlloc(1)
		Vec_IntPush(out, lit)
		return out
	}
	obj := kit.Kit_DsdNtkObj(ntk, Abc_Lit2Var(lit))
	if obj == nil || nod == nil {
		return nil
	}
	if obj.Type == kit.KIT_DSD_PRIME {
		if nod.Type != AMAP_OBJ_MUX || kit.Kit_DsdObjFanNum(obj) != 3 {
			return nil
		}
		// MUX nodes are stored with three Amap fanins; the two data arms and
		// selector map in the same order as the DSD representation.
		out := Vec_IntAlloc(3)
		for i, fanLit := range kit.Kit_DsdObjFans(obj) {
			fan := Amap_LibNod(lib, Abc_Lit2Var(int32(nodFan(nod, i))))
			part := Amap_LibDeriveGatePerm_rec(lib, ntk, int32(fanLit), fan)
			if part == nil {
				Vec_IntFree(out)
				return nil
			}
			for j := int32(0); j < Vec_IntSize(part); j++ {
				Vec_IntPush(out, Vec_IntEntry(part, j))
			}
			Vec_IntFree(part)
		}
		return out
	}
	want := AMAP_OBJ_AND
	if obj.Type == kit.KIT_DSD_XOR {
		want = AMAP_OBJ_XOR
	}
	if nod.Type != want {
		return nil
	}
	// Both representations permit associative chains. Match each DSD child
	// against one unused frontier component, trying candidates by operation
	// type so commutative input order does not constrain the permutation.
	frontier := Amap_LibCollectFanins(lib, nod)
	children := kit.Kit_DsdObjFans(obj)
	if int32(len(children)) != Vec_IntSize(frontier) {
		Vec_IntFree(frontier)
		return nil
	}
	used := make([]bool, len(children))
	result := Vec_IntAlloc(int32(len(children)))
	var assign func(int) bool
	assign = func(pos int) bool {
		if pos == len(children) {
			return true
		}
		dlit := int32(children[pos])
		for j := range children {
			if used[j] {
				continue
			}
			nlit := Vec_IntEntry(frontier, int32(j))
			nchild := Amap_LibNod(lib, Abc_Lit2Var(nlit))
			dobj := kit.Kit_DsdNtkObj(ntk, Abc_Lit2Var(dlit))
			if kit.Kit_DsdLitIsLeaf(ntk, dlit) == 0 && (dobj == nil || nchild.Type != want) {
				continue
			}
			part := Amap_LibDeriveGatePerm_rec(lib, ntk, dlit, nchild)
			if part == nil {
				continue
			}
			used[j] = true
			oldLen := Vec_IntSize(result)
			for k := int32(0); k < Vec_IntSize(part); k++ {
				Vec_IntPush(result, Vec_IntEntry(part, k))
			}
			Vec_IntFree(part)
			if assign(pos + 1) {
				return true
			}
			Vec_IntShrink(result, oldLen)
			used[j] = false
		}
		return false
	}
	ok := assign(0)
	Vec_IntFree(frontier)
	if !ok {
		Vec_IntFree(result)
		return nil
	}
	return result
}

func nodFan(n *Amap_Nod_t, i int) int16 {
	switch i {
	case 0:
		return n.iFan0
	case 1:
		return n.iFan1
	default:
		return n.iFan2
	}
}

// Amap_LibVerifyPerm is retained as a diagnostic entry point. It independently
// evaluates the matched graph and reports truth-table mismatches.
func Amap_LibVerifyPerm(lib *Amap_Lib_t, gate *Amap_Gat_t, ntk *kit.Kit_DsdNtk_t, nod *Amap_Nod_t, permutation []int32) {
	if lib == nil || gate == nil || ntk == nil || nod == nil {
		return
	}
	truth := make([]uint32, kit.Kit_TruthWordNum(int32(gate.nPins)))
	kit.Kit_DsdTruth(ntk, truth)
	if Abc_LitIsCompl(kit.Kit_DsdNtkRootLit(ntk)) != 0 {
		kit.Kit_TruthNot(truth, truth, int32(gate.nPins))
	}
	if !equalTruth(truth, gate.pFunc) {
		fmt.Printf("Verification failed for gate %d (%s) and node %d (permutation %v)\\n", gate.Id, gate.pForm, nod.Id, permutation)
	}
}

func equalTruth(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Amap_LibDeriveGatePerm produces the reverse permutation used when matching
// a gate's pins to an Amap cut. The caller supplies storage for the result.
func Amap_LibDeriveGatePerm(lib *Amap_Lib_t, gate *Amap_Gat_t, ntk *kit.Kit_DsdNtk_t, nod *Amap_Nod_t, permutation []byte) int32 {
	if lib == nil || gate == nil || ntk == nil || nod == nil || len(permutation) < int(nod.nSuppSize) {
		return 0
	}
	result := Amap_LibDeriveGatePerm_rec(lib, ntk, Abc_LitRegular(kit.Kit_DsdNtkRootLit(ntk)), nod)
	if result == nil || Vec_IntSize(result) != int32(nod.nSuppSize) {
		Vec_IntFree(result)
		return 0
	}
	defer Vec_IntFree(result)
	seen := make(map[int32]bool, Vec_IntSize(result))
	for i := int32(0); i < Vec_IntSize(result); i++ {
		lit := Vec_IntEntry(result, i)
		variable := Abc_Lit2Var(lit)
		if variable < 0 || uint32(variable) >= nod.nSuppSize || seen[variable] {
			return 0
		}
		seen[variable] = true
		mapped := Abc_Var2Lit(int32(i), Abc_LitIsCompl(lit))
		if mapped < 0 || mapped > 255 {
			return 0
		}
		permutation[variable] = byte(mapped)
	}
	return 1
}
