package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Mf_CutPrintOne(c []int32) {
	fmt.Printf("%d {", Mf_CutSize(c))
	for _, id := range c[1 : 1+Mf_CutSize(c)] {
		fmt.Printf(" %d", id)
	}
	fmt.Println(" }")
}
func Mf_CubeLit(cube, variable int32) int32 { return (cube >> (2 * variable)) & 3 }
func Mf_ManCountLits(cubes []int32, nCubes, nVars int32) int32 {
	count := nCubes
	for _, cube := range cubes[:nCubes] {
		for k := int32(0); k < nVars; k++ {
			if Mf_CubeLit(cube, k) != 0 {
				count++
			}
		}
	}
	return count
}
func Mf_ManDeriveCnfs(p *Mf_Man_t, nVars, nClauses, nLits *int32) *Vec_Int_t {
	if p.pPars.nLutSize > 8 {
		panic("gia: CNF requires LUT size at most 8")
	}
	count := vec.Vec_IntSize(&p.vCnfSizes)
	lits := vec.Vec_IntStart(count)
	cnfs := vec.Vec_IntStartFull(count)
	var cubes [512]int32
	generate := func(id, n int32, check bool) {
		truth := vec.Vec_MemReadEntry(p.vTtMem, int(id))
		var num int32
		if p.pPars.nLutSize <= 6 {
			num = util.Abc_Tt6Cnf(truth[0], n, cubes[:])
		} else {
			num = util.Abc_Tt8Cnf(truth, n, cubes[:])
		}
		if check && num != vec.Vec_IntEntry(&p.vCnfSizes, int(id)) {
			panic("gia: inconsistent CNF size")
		}
		vec.Vec_IntWriteEntry(lits, int(id), Mf_ManCountLits(cubes[:], num, n))
		vec.Vec_IntWriteEntry(cnfs, int(id), int32(vec.Vec_IntSize(cnfs)))
		vec.Vec_IntPush(cnfs, num)
		for _, cube := range cubes[:num] {
			vec.Vec_IntPush(cnfs, cube)
		}
	}
	generate(0, 0, false)
	generate(1, 1, false)
	*nVars = 1 + Gia_ManCiNum(p.pGia) + Gia_ManCoNum(p.pGia)
	*nClauses = 1 + 2*Gia_ManCoNum(p.pGia)
	*nLits = 1 + 4*Gia_ManCoNum(p.pGia)
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Mf_ObjMapRefNum(p, obj.id) == 0 {
			continue
		}
		cut := Mf_ObjCutBest(p, obj.id)
		id := int32(uint32(Mf_CutFunc(cut)) >> 1)
		if vec.Vec_IntEntry(cnfs, int(id)) == -1 {
			generate(id, Mf_CutSize(cut), true)
		}
		*nVars++
		*nClauses += vec.Vec_IntEntry(&p.vCnfSizes, int(id))
		*nLits += vec.Vec_IntEntry(lits, int(id))
	}
	return cnfs
}
