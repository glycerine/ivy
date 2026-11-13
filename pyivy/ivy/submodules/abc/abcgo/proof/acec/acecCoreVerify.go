package acec

// Class diagnostics from src/proof/acec/acecCore.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

const acecTruthUnused uint64 = 0x1234567812345678

func Acec_VerifyClasses(p *Gia_Man_t, lits, reprs *Vec_Wec_t) {
	supp := vec.Vec_IntAlloc(100)
	temp := vec.Vec_WrdStart(int(gia.Gia_ManObjNum(p)))
	truths := make([][]uint64, vec.Vec_WecSize(lits))
	overs := 0
	for i := range truths {
		for _, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(lits, i)) {
			truth := gia.Gia_ObjComputeTruth6Cis(p, lit, supp, temp)
			if vec.Vec_IntSize(supp) > 6 {
				overs++
				truths[i] = append(truths[i], acecTruthUnused)
				continue
			}
			n := util.Abc_Tt6MinBase(&truth, vec.Vec_IntArray(supp), int32(vec.Vec_IntSize(supp)))
			vec.Vec_IntShrink(supp, int(n))
			if n > 5 {
				overs++
				truths[i] = append(truths[i], acecTruthUnused)
			} else {
				truths[i] = append(truths[i], truth)
			}
		}
	}
	if overs > 0 {
		fmt.Printf("Detected %d oversize support nodes.\n", overs)
	}
	errors := 0
	for i := 0; i < vec.Vec_WecSize(reprs); i++ {
		level := vec.Vec_IntArray(vec.Vec_WecEntry(reprs, i))
		for k, a := range level {
			for j := k + 1; j < len(level); j++ {
				b := level[j]
				x, y := truths[i][k], truths[i][j]
				if x == acecTruthUnused || y == acecTruthUnused {
					continue
				}
				if (a == b && x != y) || (a == b^1 && x != ^y) {
					errors++
					fmt.Printf("Rank %d:  Lit %d and %d do not pass verification.\n", i, k, j)
				}
			}
		}
	}
	if errors > 0 {
		fmt.Printf("Total errors in equivalence classes = %d.\n", errors)
	}
}
func Acec_MatchPrintEquivLits(p *Gia_Man_t, lits *Vec_Wec_t, cost []int32, verbose int32) {
	fmt.Println("Leaf literals and their classes:")
	for i := 0; i < vec.Vec_WecSize(lits); i++ {
		level := vec.Vec_IntArray(vec.Vec_WecEntry(lits, i))
		if len(level) == 0 {
			continue
		}
		fmt.Printf("Rank %2d : %2d  ", i, len(level))
		for _, lit := range level {
			sign := "+"
			if lit&1 != 0 {
				sign = "-"
			}
			fmt.Printf("%s%d(%d) ", sign, lit>>1, cost[lit>>1]^(lit&1))
		}
		fmt.Println()
	}
	if verbose == 0 {
		return
	}
	supp := vec.Vec_IntAlloc(100)
	temp := vec.Vec_WrdStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; i < vec.Vec_WecSize(lits); i++ {
		level := vec.Vec_IntArray(vec.Vec_WecEntry(lits, i))
		if len(level) == 0 {
			continue
		}
		for _, lit := range level {
			truth := gia.Gia_ObjComputeTruth6Cis(p, lit, supp, temp)
			fmt.Printf("Rank = %4d : Obj = %4d  ", i, lit>>1)
			if vec.Vec_IntSize(supp) > 6 {
				fmt.Printf("Supp = %d.\n", vec.Vec_IntSize(supp))
				continue
			}
			n := util.Abc_Tt6MinBase(&truth, vec.Vec_IntArray(supp), int32(vec.Vec_IntSize(supp)))
			vec.Vec_IntShrink(supp, int(n))
			if n > 5 {
				fmt.Printf("Supp = %d.\n", n)
				continue
			}
			fmt.Printf("%x  ", truth)
			vec.Vec_IntPrint(supp)
		}
		fmt.Println()
	}
}
