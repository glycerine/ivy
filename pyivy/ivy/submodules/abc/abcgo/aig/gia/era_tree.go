package gia

import "github.com/glycerine/abcgo/misc/vec"

func giaAreBranchNode(o *Gia_ObjAre_t, index int) bool {
	switch index {
	case 0:
		return Gia_ObjHasBranch0(o) != 0
	case 1:
		return Gia_ObjHasBranch1(o) != 0
	default:
		return Gia_ObjHasBranch2(o) != 0
	}
}
func giaAreCubeBranch(o *Gia_ObjAre_t, s *Gia_StaAre_t) int {
	if Gia_StaHasValue0(s, int32(o.iVar)) != 0 {
		return 0
	}
	if Gia_StaHasValue1(s, int32(o.iVar)) != 0 {
		return 1
	}
	return 2
}
func Gia_ManAreCubeCheckTree_rec(p *Gia_ManAre_t, o *Gia_ObjAre_t, s *Gia_StaAre_t) int32 {
	check := func(i int) int32 {
		if giaAreBranchNode(o, i) {
			return Gia_ManAreCubeCheckTree_rec(p, Gia_ManAreObj(p, o.F[i]), s)
		}
		return Gia_ManAreCubeCheckList(p, &o.F[i], s)
	}
	branch := giaAreCubeBranch(o, s)
	if branch != 2 && check(branch) == 0 {
		return 0
	}
	return check(2)
}
func Gia_ManAreCubeAddToTree_rec(p *Gia_ManAre_t, o *Gia_ObjAre_t, s *Gia_StaAre_t) {
	branch := giaAreCubeBranch(o, s)
	if giaAreBranchNode(o, branch) {
		Gia_ManAreCubeAddToTree_rec(p, Gia_ManAreObj(p, o.F[branch]), s)
		return
	}
	Gia_ManAreCubeAddToList(p, &o.F[branch], s)
	counts := [3]*uint32{&o.nStas0, &o.nStas1, &o.nStas2}
	count := counts[branch]
	*count++
	if *count == MAX_CUBE_NUM {
		*count = uint32(Gia_ManAreListCountListUsed(p, o.F[branch]))
		if *count < MAX_CUBE_NUM/2 {
			Gia_ManAreCompress(p, &o.F[branch])
		} else {
			Gia_ManAreRebalance(p, &o.F[branch])
			*count = 0
		}
	}
}
func Gia_ManAreCubeCollectList(p *Gia_ManAre_t, root *Gia_PtrAre_t, s *Gia_StaAre_t) int32 {
	for id := Gia_Ptr2Int(*root); id != 0; {
		cube := Gia_ManAreStaInt(p, int32(id))
		if Gia_StaIsUsed(cube) != 0 && Gia_StaAreDisjoint(s, cube, p.nWords) == 0 {
			if int32(id) <= p.iStaCur {
				vec.Vec_IntPush(p.vCubesA, int32(id))
			} else {
				vec.Vec_IntPush(p.vCubesB, int32(id))
			}
		}
		id = Gia_Ptr2Int(cube.iNext)
	}
	return 1
}
func Gia_ManAreCubeCollectTree_rec(p *Gia_ManAre_t, o *Gia_ObjAre_t, s *Gia_StaAre_t) int32 {
	collect := func(i int) int32 {
		if giaAreBranchNode(o, i) {
			return Gia_ManAreCubeCollectTree_rec(p, Gia_ManAreObj(p, o.F[i]), s)
		}
		return Gia_ManAreCubeCollectList(p, &o.F[i], s)
	}
	branch := giaAreCubeBranch(o, s)
	if branch != 2 && collect(branch) == 0 {
		return 0
	}
	return collect(2)
}
func Gia_ManAreCubeCheckTree(p *Gia_ManAre_t, s *Gia_StaAre_t) int32 {
	if p.fTree == 0 {
		panic("gia: expected cube tree")
	}
	for {
		vec.Vec_IntClear(p.vCubesA)
		vec.Vec_IntClear(p.vCubesB)
		Gia_ManAreCubeCollectTree_rec(p, Gia_ManAreObj(p, p.Root), s)
		retry := false
		for _, list := range []*Vec_Int_t{p.vCubesA, p.vCubesB} {
			for _, id := range vec.Vec_IntArray(list) {
				cube := Gia_ManAreStaInt(p, id)
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
				retry = true
				break
			}
			if retry {
				break
			}
		}
		if !retry {
			return 1
		}
	}
}
func Gia_ManAreCubeProcess(p *Gia_ManAre_t, s *Gia_StaAre_t) int32 {
	p.nChecks = 0
	if p.fTree == 0 && p.nStas == MAX_CUBE_NUM {
		Gia_ManAreRebalance(p, &p.Root)
	}
	if p.fTree != 0 {
		result := Gia_ManAreCubeCheckTree(p, s)
		if result != 0 {
			Gia_ManAreCubeAddToTree_rec(p, Gia_ManAreObj(p, p.Root), s)
		}
		return result
	}
	result := Gia_ManAreCubeCheckList(p, &p.Root, s)
	if result != 0 {
		Gia_ManAreCubeAddToList(p, &p.Root, s)
	}
	return result
}
