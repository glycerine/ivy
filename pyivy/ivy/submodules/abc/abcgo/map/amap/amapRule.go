package amap

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/bool/kit"
)

// Amap_CreateRulesPrime builds the mux-rule products for a 3-input prime DSD.
func Amap_CreateRulesPrime(p *Amap_Lib_t, a, b, c *Vec_Int_t) *Vec_Int_t {
	result := Vec_IntAlloc(Vec_IntSize(a) * Vec_IntSize(b) * Vec_IntSize(c))
	for _, x := range Vec_IntArray(a) {
		for _, y := range Vec_IntArray(b) {
			for _, z := range Vec_IntArray(c) {
				node := Amap_LibFindMux(p, x, y, z)
				if node < 0 {
					node = Amap_LibCreateMux(p, x, y, z)
				}
				Vec_IntPushUnique(result, Abc_Var2Lit(node, 0))
			}
		}
	}
	return result
}

func Amap_CreateRulesTwo(p *Amap_Lib_t, result, a, b *Vec_Int_t, xor int32) {
	for _, x := range Vec_IntArray(a) {
		for _, y := range Vec_IntArray(b) {
			node := Amap_LibFindNode(p, x, y, xor)
			if node < 0 {
				node = Amap_LibCreateNode(p, x, y, xor)
			}
			Vec_IntPushUnique(result, Abc_Var2Lit(node, 0))
		}
	}
}

func Amap_CreateCheckAllZero(groups *Vec_Ptr_t) int32 {
	for i := int32(0); i < Vec_PtrSize(groups); i++ {
		values, ok := Vec_PtrEntry(groups, i).(*Vec_Int_t)
		if !ok || Vec_IntSize(values) != 1 || Vec_IntEntry(values, 0) != 0 {
			return 0
		}
	}
	return 1
}

// Amap_CreateRulesVector_rec enumerates binary associations of DSD components.
// For the all-zero basis it uses canonical prefix/suffix splits; otherwise it
// visits each nontrivial subset partition once.
func Amap_CreateRulesVector_rec(lib *Amap_Lib_t, groups *Vec_Ptr_t, xor int32) *Vec_Int_t {
	n := Vec_PtrSize(groups)
	if n == 0 {
		return Vec_IntAlloc(0)
	}
	if n == 1 {
		v, _ := Vec_PtrEntry(groups, 0).(*Vec_Int_t)
		if v == nil {
			return Vec_IntAlloc(0)
		}
		return Vec_IntDup(v)
	}
	result := Vec_IntAlloc(16)
	makeGroup := func(indices []int) *Vec_Ptr_t {
		out := Vec_PtrAlloc(int32(len(indices)))
		for _, index := range indices {
			Vec_PtrPush(out, Vec_PtrEntry(groups, int32(index)))
		}
		return out
	}
	combine := func(left, right []int) {
		lg, rg := makeGroup(left), makeGroup(right)
		defer Vec_PtrFree(lg)
		defer Vec_PtrFree(rg)
		a := Amap_CreateRulesVector_rec(lib, lg, xor)
		b := Amap_CreateRulesVector_rec(lib, rg, xor)
		defer Vec_IntFree(a)
		defer Vec_IntFree(b)
		Amap_CreateRulesTwo(lib, result, a, b, xor)
	}
	if Amap_CreateCheckAllZero(groups) != 0 {
		for split := n - 1; split >= 1; split-- {
			left, right := make([]int, split), make([]int, n-split)
			for i := range left {
				left[i] = i
			}
			for i := range right {
				right[i] = int(split) + i
			}
			combine(left, right)
		}
		return result
	}
	// Keep the first item on the left, removing mirror duplicates.
	limit := 1 << uint(n)
	for mask := 1; mask < limit-1; mask++ {
		if mask&1 == 0 {
			continue
		}
		left, right := make([]int, 0, n), make([]int, 0, n)
		for i := int32(0); i < n; i++ {
			if mask&(1<<uint(i)) != 0 {
				left = append(left, int(i))
			} else {
				right = append(right, int(i))
			}
		}
		if len(right) != 0 {
			combine(left, right)
		}
	}
	return result
}

func Amap_CreateRulesFromDsd_rec(lib *Amap_Lib_t, ntk *kit.Kit_DsdNtk_t, lit int32) *Vec_Int_t {
	if kit.Kit_DsdLitIsLeaf(ntk, lit) != 0 {
		return Vec_IntStartNatural(1)
	}
	obj := kit.Kit_DsdNtkObj(ntk, Abc_Lit2Var(lit))
	if obj == nil {
		return nil
	}
	groups := Vec_PtrAlloc(int32(kit.Kit_DsdObjFanNum(obj)))
	defer Vec_PtrFree(groups)
	for _, fan := range kit.Kit_DsdObjFans(obj) {
		values := Amap_CreateRulesFromDsd_rec(lib, ntk, Abc_LitRegular(int32(fan)))
		if values == nil {
			return nil
		}
		if Abc_LitIsCompl(int32(fan)) != 0 {
			for i, value := range Vec_IntArray(values) {
				if value > 0 {
					Vec_IntWriteEntry(values, int32(i), Abc_LitNot(value))
				}
			}
		}
		Vec_PtrPush(groups, values)
	}
	defer func() {
		for i := int32(0); i < Vec_PtrSize(groups); i++ {
			if v, ok := Vec_PtrEntry(groups, i).(*Vec_Int_t); ok {
				Vec_IntFree(v)
			}
		}
	}()
	var result *Vec_Int_t
	switch obj.Type {
	case kit.KIT_DSD_AND:
		result = Amap_CreateRulesVector_rec(lib, groups, 0)
	case kit.KIT_DSD_XOR:
		result = Amap_CreateRulesVector_rec(lib, groups, 1)
	case kit.KIT_DSD_PRIME:
		if len(kit.Kit_DsdObjFans(obj)) != 3 {
			return nil
		}
		a, _ := Vec_PtrEntry(groups, 0).(*Vec_Int_t)
		b, _ := Vec_PtrEntry(groups, 1).(*Vec_Int_t)
		c, _ := Vec_PtrEntry(groups, 2).(*Vec_Int_t)
		result = Amap_CreateRulesPrime(lib, a, b, c)
	default:
		return nil
	}
	return result
}

func Amap_CreateRulesFromDsd(lib *Amap_Lib_t, ntk *kit.Kit_DsdNtk_t) *Vec_Int_t {
	if ntk == nil || kit.Kit_DsdNtkVarNum(ntk) < 2 {
		return nil
	}
	root := kit.Kit_DsdNtkRootLit(ntk)
	result := Amap_CreateRulesFromDsd_rec(lib, ntk, Abc_LitRegular(root))
	if result == nil {
		return nil
	}
	if Abc_LitIsCompl(root) != 0 {
		for i, value := range Vec_IntArray(result) {
			Vec_IntWriteEntry(result, int32(i), Abc_LitNot(value))
		}
	}
	return result
}

// Amap_CreateCheckEqual_rec reports identical ordered DSD subgraphs with the
// same operator and edge polarities.
func Amap_CreateCheckEqual_rec(ntk *kit.Kit_DsdNtk_t, left, right int32) int32 {
	if kit.Kit_DsdLitIsLeaf(ntk, left) != 0 || kit.Kit_DsdLitIsLeaf(ntk, right) != 0 {
		if kit.Kit_DsdLitIsLeaf(ntk, left) != 0 && kit.Kit_DsdLitIsLeaf(ntk, right) != 0 && left == right {
			return 1
		}
		return 0
	}
	a, b := kit.Kit_DsdNtkObj(ntk, Abc_Lit2Var(left)), kit.Kit_DsdNtkObj(ntk, Abc_Lit2Var(right))
	if a == nil || b == nil || a.Type != b.Type || a.Type == kit.KIT_DSD_PRIME || kit.Kit_DsdObjFanNum(a) != kit.Kit_DsdObjFanNum(b) {
		return 0
	}
	af, bf := kit.Kit_DsdObjFans(a), kit.Kit_DsdObjFans(b)
	for i := range af {
		if (af[i]&1) != (bf[i]&1) || Amap_CreateCheckEqual_rec(ntk, int32(af[i])&^1, int32(bf[i])&^1) == 0 {
			return 0
		}
	}
	return 1
}

func Amap_CreateCheckAsym_rec(ntk *kit.Kit_DsdNtk_t, lit int32, syms **Vec_Int_t) {
	if kit.Kit_DsdLitIsLeaf(ntk, lit) != 0 {
		return
	}
	obj := kit.Kit_DsdNtkObj(ntk, Abc_Lit2Var(lit))
	if obj == nil {
		return
	}
	fans := kit.Kit_DsdObjFans(obj)
	for _, fan := range fans {
		Amap_CreateCheckAsym_rec(ntk, Abc_LitRegular(int32(fan)), syms)
	}
	if obj.Type != kit.KIT_DSD_AND && obj.Type != kit.KIT_DSD_XOR {
		return
	}
	for i := 0; i < len(fans); i++ {
		for j := i + 1; j < len(fans); j++ {
			a, b := int32(fans[i]), int32(fans[j])
			if (a&1) == (b&1) || kit.Kit_DsdLitIsLeaf(ntk, a) == 0 || kit.Kit_DsdLitIsLeaf(ntk, b) == 0 {
				continue
			}
			if *syms == nil {
				*syms = Vec_IntAlloc(8)
			}
			Vec_IntPush(*syms, (Abc_Lit2Var(a)<<8)|Abc_Lit2Var(b))
		}
	}
}

func Amap_CreateCheckAsym(ntk *kit.Kit_DsdNtk_t, syms **Vec_Int_t) {
	if ntk != nil {
		Amap_CreateCheckAsym_rec(ntk, Abc_LitRegular(kit.Kit_DsdNtkRootLit(ntk)), syms)
	}
}

func Amap_CreateRulesForGate(lib *Amap_Lib_t, gate *Amap_Gat_t) {
	if gate == nil || gate.nPins < 2 || len(gate.pFunc) == 0 {
		return
	}
	ntk := kit.Kit_DsdDecomposeMux(gate.pFunc, int32(gate.nPins), 2)
	if ntk == nil {
		return
	}
	defer kit.Kit_DsdNtkFree(ntk)
	if gate.nPins == 2 && (gate.pFunc[0] == 0x66666666 || gate.pFunc[0] == ^uint32(0x66666666)) {
		lib.fHasXor = 1
	}
	if kit.Kit_DsdNonDsdSizeMax(ntk) == 3 {
		lib.fHasMux, gate.fMux = 1, 1
	}
	ntk = kit.Kit_DsdExpand(ntk)
	if ntk == nil {
		return
	}
	defer kit.Kit_DsdNtkFree(ntk)
	if lib.fVerbose != 0 {
		fmt.Printf("\nProcessing library gate %4d: %10s ", gate.Id, gate.pName)
		kit.Kit_DsdPrint(os.Stdout, ntk)
	}
	syms := (*Vec_Int_t)(nil)
	Amap_CreateCheckAsym(ntk, &syms)
	if syms != nil {
		defer Vec_IntFree(syms)
	}
	nodes := Amap_CreateRulesFromDsd(lib, ntk)
	if nodes == nil {
		return
	}
	defer Vec_IntFree(nodes)
	for _, lit := range Vec_IntArray(nodes) {
		if lit <= 1 {
			continue
		}
		node := Amap_LibNod(lib, Abc_Lit2Var(lit))
		set := &Amap_Set_t{iGate: gate.Id, fInv: uint32(Abc_LitIsCompl(lit)), nIns: gate.nPins}
		if Amap_LibDeriveGatePerm(lib, gate, ntk, node, set.Ins[:]) == 0 {
			if lib.fVerbose != 0 {
				fmt.Printf("Could not prepare gate %q\n", gate.pName)
			}
			continue
		}
		set.pNext, node.pSets = node.pSets, set
		lib.nSets++
		if syms == nil {
			continue
		}
		for i := int32(0); i < Vec_IntSize(syms); i++ {
			entry := Vec_IntEntry(syms, i)
			x, y := entry&0xff, (entry>>8)&0xff
			if x < 0 || y < 0 || x >= int32(gate.nPins) || y >= int32(gate.nPins) {
				continue
			}
			copySet := amapSymmetryCopy(set, int(x), int(y))
			copySet.pNext = node.pSets
			node.pSets = copySet
			lib.nSets++
		}
	}
}

// amapSymmetryCopy swaps variable assignments while retaining the phase at
// each destination pin. Symmetry pairs discovered from complemented DSD
// leaves require this phase-preserving permutation rather than a byte swap.
func amapSymmetryCopy(set *Amap_Set_t, x, y int) *Amap_Set_t {
	copySet := *set
	pinX, pinY := int32(set.Ins[x]), int32(set.Ins[y])
	copySet.Ins[x] = byte(Abc_Var2Lit(Abc_Lit2Var(pinY), Abc_LitIsCompl(pinX)))
	copySet.Ins[y] = byte(Abc_Var2Lit(Abc_Lit2Var(pinX), Abc_LitIsCompl(pinY)))
	return &copySet
}

func Amap_LibCreateRules(lib *Amap_Lib_t, veryVerbose int32) {
	if lib == nil {
		return
	}
	lib.fVerbose = veryVerbose
	lib.vRules, lib.vRulesX = Vec_PtrAlloc(100), Vec_PtrAlloc(100)
	lib.vRules3 = Vec_IntAlloc(100)
	Amap_LibCreateVar(lib)
	if lib.vSelect == nil {
		return
	}
	for _, item := range Vec_PtrArray(lib.vSelect) {
		gate, ok := item.(*Amap_Gat_t)
		if !ok || gate == nil || gate.nPins < 2 {
			continue
		}
		if len(gate.pFunc) == 0 {
			fmt.Printf("Amap_LibCreateRules(): skipping gate %s (%s): no truth table\n", gate.pName, gate.pForm)
			continue
		}
		Amap_CreateRulesForGate(lib, gate)
	}
	lib.pRules = Amap_LibLookupTableAlloc(lib.vRules, 0)
	lib.pRulesX = Amap_LibLookupTableAlloc(lib.vRulesX, 0)
}
