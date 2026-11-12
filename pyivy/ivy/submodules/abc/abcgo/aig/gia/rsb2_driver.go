package gia

import "github.com/glycerine/abcgo/misc/vec"

func Abc_ResubNodeToTry(tried *Vec_Int_t, first, last int32) int32 {
	seen := make(map[int32]bool, vec.Vec_IntSize(tried))
	for _, id := range vec.Vec_IntArray(tried) {
		seen[id] = true
	}
	for id := last - 1; id >= first; id-- {
		if !seen[id] {
			return id
		}
	}
	return -1
}

// Abc_ResubComputeWindow optimizes an Rsb2 array and returns nil when no
// replacement was found. The second result is the number of replacements.
func Abc_ResubComputeWindow(objects []int32, nObjs, divsMax, levelIncrease, useXor, useZeroCost, debug, verbose int32) ([]int32, int32) {
	p := Gia_Rsb2ManAlloc()
	defer Gia_Rsb2ManFree(p)
	Gia_Rsb2ManStart(p, objects, nObjs, divsMax, levelIncrease, useXor, useZeroCost, debug, verbose)
	changes := int32(0)
	tried := vec.Vec_IntAlloc(0)
	copies := vec.Vec_IntAlloc(0)
	for {
		node := Abc_ResubNodeToTry(tried, p.nPis+1, p.iFirstPo)
		if node < 0 {
			break
		}
		Gia_Rsb2ManDivs(p, node)
		divisors := vec.Vec_PtrAlloc(len(p.divisors))
		for _, d := range p.divisors {
			vec.Vec_PtrPush(divisors, d)
		}
		manager := Gia_ResbAlloc(1)
		Gia_ManResubPerform(manager, divisors, 1, max(0, p.nMffc-1), divsMax, 0, useXor, debug, verbose)
		if vec.Vec_IntSize(manager.vGates) == 0 {
			vec.Vec_IntPush(tried, node)
			Gia_ResbFree(manager)
			continue
		}
		source := vec.Vec_IntAlloc(len(p.objects))
		for _, lit := range p.objects {
			vec.Vec_IntPush(source, lit)
		}
		divIDs := vec.Vec_IntAlloc(len(p.divisorIDs))
		for _, id := range p.divisorIDs {
			vec.Vec_IntPush(divIDs, id)
		}
		replacement := Gia_Rsb2ManInsert(p.nPis, p.nPos, source, node, manager.vGates, divIDs, copies)
		Gia_ResbFree(manager)
		remapped := make([]int32, 0, vec.Vec_IntSize(tried))
		for _, id := range vec.Vec_IntArray(tried) {
			value := vec.Vec_IntEntry(copies, int(id))
			if value > 2*p.nPis {
				remapped = append(remapped, value>>1)
			}
		}
		updated := append([]int32(nil), vec.Vec_IntArray(replacement)...)
		Gia_Rsb2ManStart(p, updated, int32(len(updated)/2), divsMax, levelIncrease, useXor, useZeroCost, debug, verbose)
		vec.Vec_IntClear(tried)
		for _, id := range remapped {
			vec.Vec_IntPush(tried, id)
		}
		changes++
	}
	if changes == 0 {
		return nil, 0
	}
	return append([]int32(nil), p.objects...), changes
}
func Abc_ResubComputeWindow2(objects []int32, nObjs, divsMax, levelIncrease, useXor, useZeroCost, debug, verbose int32) ([]int32, int32) {
	_, _, _, _, _, _ = divsMax, levelIncrease, useXor, useZeroCost, debug, verbose
	if nObjs < 0 || len(objects) != int(2*nObjs) {
		panic("gia: invalid Rsb2 window")
	}
	return append([]int32(nil), objects...), 0
}
func Gia_ManResub2Test(g *Gia_Man_t) *Gia_Man_t {
	objects := Gia_ManToResub(g)
	updated, _ := Abc_ResubComputeWindow(objects, g.nObjs, 1000, -1, 0, 0, 0, 0)
	if updated == nil {
		return Gia_ManDup(g)
	}
	result := Gia_ManFromResub(updated, int32(len(updated)/2), Gia_ManCiNum(g))
	result.pName = g.pName
	return result
}
