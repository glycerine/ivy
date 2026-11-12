package dec

import (
	"github.com/glycerine/abcgo/misc/mvc"
	"strings"
)

func sopRows(sop string) ([]string, int, bool) {
	var cubes []string
	n := -1
	complement := false
	for _, line := range strings.Split(strings.TrimSpace(sop), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) == 1 {
			if fields[0] == "0" {
				return nil, 0, false
			}
			if fields[0] == "1" {
				return []string{""}, 0, false
			}
			panic("dec: invalid SOP")
		}
		cube, out := fields[0], fields[1]
		if out != "0" && out != "1" {
			panic("dec: invalid SOP output")
		}
		if n < 0 {
			n = len(cube)
			complement = out == "0"
		}
		if len(cube) != n || complement != (out == "0") {
			panic("dec: inconsistent SOP")
		}
		for _, c := range cube {
			if c != '0' && c != '1' && c != '-' {
				panic("dec: invalid SOP literal")
			}
		}
		cubes = append(cubes, cube)
	}
	if n < 0 {
		n = 0
	}
	return cubes, n, complement
}
func Dec_Factor(sop string) *Dec_Graph_t {
	rows, n, compl := sopRows(sop)
	if len(rows) == 0 {
		return Dec_GraphCreateConst0()
	}
	for _, row := range rows {
		if strings.Trim(row, "-") == "" {
			if compl {
				return Dec_GraphCreateConst0()
			}
			return Dec_GraphCreateConst1()
		}
	}
	cover := Dec_ConvertSopToMvc(sop)
	mvc.Mvc_CoverContain(cover)
	mvc.Mvc_CoverInverse(cover)
	g := Dec_GraphCreate(int32(n))
	g.eRoot = Dec_Factor_rec(g, cover)
	if compl {
		Dec_GraphComplement(g)
	}
	return g
}
func Dec_Factor_rec(g *Dec_Graph_t, c *mvc.Mvc_Cover_t) Dec_Edge_t {
	if mvc.Mvc_CoverReadCubeNum(c) == 0 {
		panic("dec: empty factoring cover")
	}
	divisor := mvc.Mvc_CoverDivisor(c)
	if divisor == nil {
		return Dec_FactorTrivial(g, c)
	}
	quotient, _ := mvc.Mvc_CoverDivideInternal(c, divisor)
	if mvc.Mvc_CoverReadCubeNum(quotient) == 1 {
		return Dec_FactorLF_rec(g, c, quotient)
	}
	mvc.Mvc_CoverMakeCubeFree(quotient)
	divisor, remainder := mvc.Mvc_CoverDivideInternal(c, quotient)
	if mvc.Mvc_CoverIsCubeFree(divisor) {
		a, b := Dec_Factor_rec(g, divisor), Dec_Factor_rec(g, quotient)
		product := Dec_GraphAddNodeAnd(g, a, b)
		if mvc.Mvc_CoverReadCubeNum(remainder) == 0 {
			return product
		}
		return Dec_GraphAddNodeOr(g, product, Dec_Factor_rec(g, remainder))
	}
	return Dec_FactorLF_rec(g, c, mvc.Mvc_CoverCommonCubeCover(divisor))
}
func Dec_FactorLF_rec(g *Dec_Graph_t, c, simple *mvc.Mvc_Cover_t) Dec_Edge_t {
	divisor := mvc.Mvc_CoverBestLiteralCover(c, simple)
	quotient, remainder := mvc.Mvc_CoverDivideByLiteral(c, divisor)
	lit := Dec_FactorTrivialCube(g, divisor, mvc.Mvc_CoverReadCubeHead(divisor), nil)
	product := Dec_GraphAddNodeAnd(g, lit, Dec_Factor_rec(g, quotient))
	if mvc.Mvc_CoverReadCubeNum(remainder) == 0 {
		return product
	}
	return Dec_GraphAddNodeOr(g, product, Dec_Factor_rec(g, remainder))
}
func Dec_FactorTrivial(g *Dec_Graph_t, c *mvc.Mvc_Cover_t) Dec_Edge_t {
	var edges []Dec_Edge_t
	for cube := c.Cubes.Head; cube != nil; cube = cube.Next {
		edges = append(edges, Dec_FactorTrivialCube(g, c, cube, nil))
	}
	return Dec_FactorTrivialTree_rec(g, edges, int32(len(edges)), 1)
}
func Dec_FactorTrivialCube(g *Dec_Graph_t, c *mvc.Mvc_Cover_t, cube *mvc.Mvc_Cube_t, scratch []int32) Dec_Edge_t {
	var edges []Dec_Edge_t
	for i := 0; i < c.NBits; i++ {
		if mvc.Mvc_CubeBitValue(cube, i) {
			edges = append(edges, Dec_EdgeCreate(int32(i/2), int32(i%2)))
		}
	}
	return Dec_FactorTrivialTree_rec(g, edges, int32(len(edges)), 0)
}
func Dec_FactorTrivialTree_rec(g *Dec_Graph_t, edges []Dec_Edge_t, count, or int32) Dec_Edge_t {
	if count <= 0 || int(count) > len(edges) {
		panic("dec: empty factoring tree")
	}
	if count == 1 {
		return edges[0]
	}
	split := count / 2
	a := Dec_FactorTrivialTree_rec(g, edges[:split], split, or)
	b := Dec_FactorTrivialTree_rec(g, edges[split:count], count-split, or)
	if or != 0 {
		return Dec_GraphAddNodeOr(g, a, b)
	}
	return Dec_GraphAddNodeAnd(g, a, b)
}
func Dec_ConvertSopToMvc(sop string) *mvc.Mvc_Cover_t {
	rows, n, _ := sopRows(sop)
	cover := mvc.Mvc_CoverAlloc(mvc.Mvc_ManagerStart(), n*2)
	for _, row := range rows {
		cube := mvc.Mvc_CubeAlloc(cover)
		mvc.Mvc_CoverAddCubeTail(cover, cube)
		mvc.Mvc_CubeBitFill(cube)
		for i, c := range row {
			if c == '0' {
				mvc.Mvc_CubeBitRemove(cube, i*2+1)
			} else if c == '1' {
				mvc.Mvc_CubeBitRemove(cube, i*2)
			}
		}
	}
	return cover
}
func Dec_FactorVerify(sop string, g *Dec_Graph_t) int32 {
	rows, n, compl := sopRows(sop)
	native := g.Native()
	if !native.Constant && native.LeafCount != n {
		return 0
	}
	inputs := make([]bool, n)
	var check func(int) bool
	check = func(i int) bool {
		if i < n {
			inputs[i] = false
			if !check(i + 1) {
				return false
			}
			inputs[i] = true
			return check(i + 1)
		}
		want := false
		for _, row := range rows {
			match := true
			for j, c := range row {
				if c != '-' && (c == '1') != inputs[j] {
					match = false
					break
				}
			}
			want = want || match
		}
		if compl {
			want = !want
		}
		return native.Evaluate(inputs) == want
	}
	return decBool(check(0))
}
