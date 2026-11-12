package gia

import "fmt"

func Gia_ManArePrintCube(p *Gia_ManAre_t, s *Gia_StaAre_t) {
	fmt.Printf("%4d %4d :  Prev %4d   %p   ", p.iStaCur, p.nStas-1, Gia_Ptr2Int(s.iPrev), s)
	var zeros, ones, dashes int
	for i := int32(0); i < p.pAig.nRegs; i++ {
		if Gia_StaHasValue0(s, i) != 0 {
			fmt.Print("0")
			zeros++
		} else if Gia_StaHasValue1(s, i) != 0 {
			fmt.Print("1")
			ones++
		} else {
			fmt.Print("-")
			dashes++
		}
	}
	fmt.Printf("  0 =%3d  1 =%3d  - =%3d\n", zeros, ones, dashes)
}
func Gia_ManAreListCountListUsed(p *Gia_ManAre_t, root Gia_PtrAre_t) int32 {
	var n int32
	for s := Gia_ManAreSta(p, root); Gia_StaIsGood(p, s) != 0; s = Gia_StaNext(p, s) {
		n += Gia_StaIsUsed(s)
	}
	return n
}
func Gia_ManAreListCountUsed_rec(p *Gia_ManAre_t, root Gia_PtrAre_t, tree int32) int32 {
	if tree == 0 {
		return Gia_ManAreListCountListUsed(p, root)
	}
	obj := Gia_ManAreObj(p, root)
	return Gia_ManAreListCountUsed_rec(p, obj.F[0], Gia_ObjHasBranch0(obj)) + Gia_ManAreListCountUsed_rec(p, obj.F[1], Gia_ObjHasBranch1(obj)) + Gia_ManAreListCountUsed_rec(p, obj.F[2], Gia_ObjHasBranch2(obj))
}
func Gia_ManAreListCountUsed(p *Gia_ManAre_t) int32 {
	return Gia_ManAreListCountUsed_rec(p, p.Root, p.fTree)
}
func Gia_ManArePrintListUsed(p *Gia_ManAre_t, root Gia_PtrAre_t) int32 {
	for s := Gia_ManAreSta(p, root); Gia_StaIsGood(p, s) != 0; s = Gia_StaNext(p, s) {
		if Gia_StaIsUsed(s) != 0 {
			Gia_ManArePrintCube(p, s)
		}
	}
	return 1
}
func Gia_ManArePrintUsed_rec(p *Gia_ManAre_t, root Gia_PtrAre_t, tree int32) int32 {
	if tree == 0 {
		return Gia_ManArePrintListUsed(p, root)
	}
	o := Gia_ManAreObj(p, root)
	return Gia_ManArePrintUsed_rec(p, o.F[0], Gia_ObjHasBranch0(o)) + Gia_ManArePrintUsed_rec(p, o.F[1], Gia_ObjHasBranch1(o)) + Gia_ManArePrintUsed_rec(p, o.F[2], Gia_ObjHasBranch2(o))
}
func Gia_ManArePrintUsed(p *Gia_ManAre_t) int32 { return Gia_ManArePrintUsed_rec(p, p.Root, p.fTree) }
func Gia_ManAreFindBestVar(p *Gia_ManAre_t, root Gia_PtrAre_t) int32 {
	best, weightBest := int32(-1), int32(-1)
	for v := int32(0); v < p.pAig.nRegs; v++ {
		var n0, n1, n2 int32
		for s := Gia_ManAreSta(p, root); Gia_StaIsGood(p, s) != 0; s = Gia_StaNext(p, s) {
			if Gia_StaIsUnused(s) != 0 {
				continue
			}
			if Gia_StaHasValue0(s, v) != 0 {
				n0++
			} else if Gia_StaHasValue1(s, v) != 0 {
				n1++
			} else {
				n2++
			}
		}
		if n0 == 0 && n1 == 0 || n0 == 0 && n2 == 0 || n1 == 0 && n2 == 0 {
			continue
		}
		weight := n0 + n1 - max(n0-n1, n1-n0)
		if weight > weightBest {
			best, weightBest = v, weight
		}
	}
	if best < 0 {
		Gia_ManArePrintListUsed(p, root)
		panic("gia: best cube variable not found")
	}
	return best
}
func Gia_ManAreRebalance(p *Gia_ManAre_t, root *Gia_PtrAre_t) {
	if Gia_Ptr2Int(*root) == 0 {
		panic("gia: cannot rebalance empty list")
	}
	obj := Gia_ManAreCreateObj(p)
	obj.iVar = uint32(Gia_ManAreFindBestVar(p, *root))
	for id := *root; Gia_Ptr2Int(id) != 0; {
		s := Gia_ManAreSta(p, id)
		next := s.iNext
		if Gia_StaIsUsed(s) != 0 {
			index := 2
			if Gia_StaHasValue0(s, int32(obj.iVar)) != 0 {
				index = 0
			} else if Gia_StaHasValue1(s, int32(obj.iVar)) != 0 {
				index = 1
			}
			s.iNext = obj.F[index]
			obj.F[index] = id
			switch index {
			case 0:
				obj.nStas0++
			case 1:
				obj.nStas1++
			case 2:
				obj.nStas2++
			}
		}
		id = next
	}
	*root = Gia_Int2Ptr(uint32(p.nObjs - 1))
	p.fTree = 1
}
func Gia_ManAreCompress(p *Gia_ManAre_t, root *Gia_PtrAre_t) {
	if Gia_Ptr2Int(*root) == 0 {
		panic("gia: cannot compress empty list")
	}
	current := *root
	*root = Gia_PtrAre_t{}
	for Gia_Ptr2Int(current) != 0 {
		s := Gia_ManAreSta(p, current)
		next := s.iNext
		if Gia_StaIsUsed(s) != 0 {
			s.iNext = *root
			*root = current
		}
		current = next
	}
}
func Gia_ManAreCubeCheckList(p *Gia_ManAre_t, root *Gia_PtrAre_t, s *Gia_StaAre_t) int32 {
	for cube := Gia_ManAreSta(p, *root); Gia_StaIsGood(p, cube) != 0; cube = Gia_StaNext(p, cube) {
		p.nChecks++
		if Gia_StaIsUnused(cube) != 0 || Gia_StaAreDisjoint(s, cube, p.nWords) != 0 {
			continue
		}
		if Gia_StaAreContain(cube, s, p.nWords) != 0 {
			Gia_ManAreRycycleSta(p, s)
			return 0
		}
		if Gia_StaAreContain(s, cube, p.nWords) != 0 {
			Gia_StaSetUnused(cube)
			continue
		}
		v := Gia_StaAreSharpVar(s, cube, p.nWords)
		if v < 0 {
			continue
		}
		if Gia_StaHasValue0(cube, v) != 0 {
			Gia_StaSetValue1(s, v)
		} else {
			Gia_StaSetValue0(s, v)
		}
	}
	return 1
}
func Gia_ManAreCubeAddToList(p *Gia_ManAre_t, root *Gia_PtrAre_t, s *Gia_StaAre_t) {
	s.iNext = *root
	*root = Gia_Int2Ptr(uint32(p.nStas - 1))
	if s != Gia_ManAreSta(p, *root) {
		panic("gia: list insert requires newest cube")
	}
}
