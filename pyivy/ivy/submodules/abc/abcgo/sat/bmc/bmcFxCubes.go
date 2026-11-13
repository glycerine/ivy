package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Div_CubePrintOne(cube *vec.Vec_Int_t, text *vec.Vec_Str_t, nvars int32) {
	vec.Vec_StrFill(text, int(nvars), '-')
	for _, lit := range vec.Vec_IntArray(cube) {
		c := byte('1')
		if lit&1 != 0 {
			c = '0'
		}
		vec.Vec_StrWriteEntry(text, int(lit>>1), c)
	}
	fmt.Println(string(vec.Vec_StrArray(text)))
}
func Div_CubePrint(cubes *vec.Vec_Wec_t, nvars int32) {
	text := vec.Vec_StrAlloc(int(nvars))
	for i := 0; i < vec.Vec_WecSize(cubes); i++ {
		Div_CubePrintOne(vec.Vec_WecEntry(cubes, i), text, nvars)
	}
}

// Div_CubePairs scores literal pairs and the divisors shared by cube pairs.
func Div_CubePairs(cubes *vec.Vec_Wec_t, nvars, nDivs int32) *vec.Vec_Int_t {
	_ = nvars
	table := Tab_TabAlloc(5)
	for i := 0; i < vec.Vec_WecSize(cubes); i++ {
		a := vec.Vec_WecEntry(cubes, i)
		av := vec.Vec_IntArray(a)
		for j := 0; j < len(av); j++ {
			for k := j + 1; k < len(av); k++ {
				Tab_TabHashAdd(table, []int32{av[j], av[k], TAB_UNUSED}, int32(DIV_AND), 1)
			}
		}
		for j := i + 1; j < vec.Vec_WecSize(cubes); j++ {
			var la, lb [2]int32
			var lits [3]int32
			var phase int32
			common := Div_FindDiv(a, vec.Vec_WecEntry(cubes, j), la[:], lb[:])
			if common < 0 {
				continue
			}
			kind := Div_FindType(la[:], lb[:], lits[:], &phase)
			if kind >= DIV_AND && kind <= DIV_MUX {
				Tab_TabHashAdd(table, lits[:], int32(kind), common)
			}
		}
	}
	fmt.Printf("Divisors = %d.\n", table.nBins)
	return Tab_TabFindBest(table, nDivs)
}
