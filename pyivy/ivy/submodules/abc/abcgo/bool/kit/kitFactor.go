package kit

import "fmt"

const KIT_FACTOR_MEM_LIMIT = 1 << 20

func Kit_SopFactor(cover []uint32, compl, nVars int32, memory *[]uint32) *Kit_Graph_t {
	if nVars < 0 || nVars >= 16 {
		panic("kit: factoring supports 0..15 variables")
	}
	if len(cover) == 0 {
		g := Kit_GraphCreateConst0()
		if compl != 0 {
			Kit_GraphComplement(g)
		}
		return g
	}
	if len(cover) == 1 && cover[0] == 0 {
		g := Kit_GraphCreateConst1()
		if compl != 0 {
			Kit_GraphComplement(g)
		}
		return g
	}
	var sop Kit_Sop_t
	Kit_SopCreateInverse(&sop, cover, 2*nVars, memory)
	g := Kit_GraphCreate(nVars)
	g.eRoot = Kit_SopFactor_rec(g, &sop, 2*nVars, memory)
	if compl != 0 {
		Kit_GraphComplement(g)
	}
	return g
}
func Kit_SopFactor_rec(g *Kit_Graph_t, sop *Kit_Sop_t, nLits int32, memory *[]uint32) Kit_Edge_t {
	if sop.cubeNum() <= 0 {
		panic("kit: cannot factor an empty intermediate cover")
	}
	var div, quo, rem, common Kit_Sop_t
	if Kit_SopDivisor(&div, sop, nLits, memory) == 0 {
		return Kit_SopFactorTrivial(g, sop, nLits)
	}
	Kit_SopDivideInternal(sop, &div, &quo, &rem, memory)
	if quo.cubeNum() <= 0 {
		panic("kit: empty quotient")
	}
	if quo.cubeNum() == 1 {
		return Kit_SopFactorLF_rec(g, sop, &quo, nLits, memory)
	}
	Kit_SopMakeCubeFree(&quo)
	Kit_SopDivideInternal(sop, &quo, &div, &rem, memory)
	if Kit_SopIsCubeFree(&div) != 0 {
		a := Kit_SopFactor_rec(g, &div, nLits, memory)
		b := Kit_SopFactor_rec(g, &quo, nLits, memory)
		result := Kit_GraphAddNodeAnd(g, a, b)
		if rem.cubeNum() == 0 {
			return result
		}
		r := Kit_SopFactor_rec(g, &rem, nLits, memory)
		return Kit_GraphAddNodeOr(g, result, r)
	}
	Kit_SopCommonCubeCover(&common, &div, memory)
	return Kit_SopFactorLF_rec(g, sop, &common, nLits, memory)
}
func Kit_SopFactorLF_rec(g *Kit_Graph_t, sop, simple *Kit_Sop_t, nLits int32, memory *[]uint32) Kit_Edge_t {
	if simple.cubeNum() != 1 {
		panic("kit: expected a simple divisor")
	}
	var div, quo, rem Kit_Sop_t
	Kit_SopBestLiteralCover(&div, sop, simple.pCubes[0], nLits, memory)
	Kit_SopDivideByCube(sop, &div, &quo, &rem, memory)
	a := Kit_SopFactorTrivialCube(g, div.pCubes[0], nLits)
	b := Kit_SopFactor_rec(g, &quo, nLits, memory)
	result := Kit_GraphAddNodeAnd(g, a, b)
	if rem.cubeNum() == 0 {
		return result
	}
	r := Kit_SopFactor_rec(g, &rem, nLits, memory)
	return Kit_GraphAddNodeOr(g, result, r)
}
func Kit_SopFactorTrivialCube_rec(g *Kit_Graph_t, cube uint32, start, end int32) Kit_Edge_t {
	var literals []int32
	for i := start; i < end; i++ {
		if cube&(1<<uint(i)) != 0 {
			literals = append(literals, i)
		}
	}
	if len(literals) == 0 {
		panic("kit: empty intermediate cube")
	}
	if len(literals) == 1 {
		return Kit_EdgeCreate(literals[0]/2, literals[0]%2)
	}
	middle := literals[len(literals)/2]
	a := Kit_SopFactorTrivialCube_rec(g, cube, start, middle)
	b := Kit_SopFactorTrivialCube_rec(g, cube, middle, end)
	return Kit_GraphAddNodeAnd(g, a, b)
}
func Kit_SopFactorTrivialCube(g *Kit_Graph_t, cube uint32, nLits int32) Kit_Edge_t {
	return Kit_SopFactorTrivialCube_rec(g, cube, 0, nLits)
}
func Kit_SopFactorTrivial_rec(g *Kit_Graph_t, cubes []uint32, nCubes, nLits int32) Kit_Edge_t {
	if nCubes < 1 {
		panic("kit: no cubes to factor")
	}
	if nCubes == 1 {
		return Kit_SopFactorTrivialCube(g, cubes[0], nLits)
	}
	middle := nCubes / 2
	a := Kit_SopFactorTrivial_rec(g, cubes[:middle], middle, nLits)
	b := Kit_SopFactorTrivial_rec(g, cubes[middle:nCubes], nCubes-middle, nLits)
	return Kit_GraphAddNodeOr(g, a, b)
}
func Kit_SopFactorTrivial(g *Kit_Graph_t, sop *Kit_Sop_t, nLits int32) Kit_Edge_t {
	return Kit_SopFactorTrivial_rec(g, sop.pCubes, sop.cubeNum(), nLits)
}
func Kit_FactorTest(truth []uint32, nVars int32) {
	var cover, memory []uint32
	result := Kit_TruthIsop(truth, nVars, &cover, 0)
	if result != 0 {
		panic("kit: direct ISOP failed")
	}
	g := Kit_SopFactor(cover, 0, nVars, &memory)
	fmt.Printf("Vars = %2d. Cubes = %3d. FFNodes = %3d. FF_memory = %3d.\n", nVars, len(cover), Kit_GraphNodeNum(g), len(memory))
}
