package abci

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sim"
)

func Abc_NtkSymmetries(n *Abc_Ntk_t, useBdds, naive, reorder, verbose int32) {
	if useBdds != 0 || naive != 0 {
		Abc_NtkSymmetriesUsingBdds(n, naive, reorder, verbose)
		return
	}
	Abc_NtkSymmetriesUsingSandS(n, verbose)
}

func Abc_NtkSymmetriesUsingSandS(n *Abc_Ntk_t, verbose int32) {
	if n == nil {
		return
	}
	if abc.Abc_NtkIsStrash(n) == 0 {
		panic("Abc_NtkSymmetriesUsingSandS requires a strashed network")
	}
	graph := abc.Abc_NtkStrashToGia(n)
	count := sim.Sim_ComputeTwoVarSymms(graph, verbose != 0)
	fmt.Printf("The total number of symmetries is %d.\n", count)
}

func Abc_NtkSymmetriesUsingBdds(n *Abc_Ntk_t, naive, reorder, verbose int32) {
	if n == nil {
		return
	}
	clk := util.Abc_Clock()
	dd := Abc_NtkBuildGlobalBdds(n, 10000000, 1, reorder, 0, verbose)
	if dd == nil {
		return
	}
	fmt.Printf("Shared BDD size = %d nodes.\n", Abc_NtkSizeOfGlobalBdds(n))
	core.Cudd_AutodynDisable(dd)
	core.Cudd_zddVarsFromBddVars(dd, 2)
	bddTime := util.Abc_Clock() - clk
	clk = util.Abc_Clock()
	Ntk_NetworkSymmsBdd(dd, n, naive, verbose)
	symmTime := util.Abc_Clock() - clk
	Abc_NtkFreeGlobalBdds(n, 1)
	abcNtkSymmPrintBddStats(naive, reorder, bddTime, symmTime)
}

func abcNtkSymmPrintBddStats(naive, reorder int32, bddTime, symmTime int64) {
	fmt.Println("Statistics of BDD-based symmetry detection:")
	fmt.Printf("Algorithm = %s. Reordering = %s. Garbage collection = yes.\n", abcSelect(naive, "naive", "fast"), abcSelect(reorder, "yes", "no"))
	util.Abc_PrintTime(1, "Constructing BDDs", bddTime)
	util.Abc_PrintTime(1, "Computing symms  ", symmTime)
	util.Abc_PrintTime(1, "TOTAL            ", bddTime+symmTime)
}

func Ntk_NetworkSymmsBdd(dd *core.DdManager, n *Abc_Ntk_t, naive, verbose int32) {
	if dd == nil || n == nil {
		return
	}
	var symms, supports int
	for _, co := range n.Objects() {
		if !co.IsCO() {
			continue
		}
		f := Abc_ObjGlobalBdd(co)
		if f == nil {
			continue
		}
		supports += core.Cudd_SupportSize(dd, f)
		if core.Cudd_IsConstant(f) {
			continue
		}
		var info *extrab.Extra_SymmInfo_t
		if naive != 0 {
			info = extrab.Extra_SymmPairsComputeNaive(dd, f)
		} else {
			info = extrab.Extra_SymmPairsCompute(dd, f)
		}
		if info == nil {
			continue
		}
		symms += info.NSymms
		if verbose != 0 {
			fmt.Printf("Output %6s (%d): ", co.Name(), info.NSymms)
			Ntk_NetworkSymmsPrint(n, info)
		}
		extrab.Extra_SymmPairsDissolve(info)
	}
	fmt.Printf("Total number of vars in functional supports = %8d.\n", supports)
	fmt.Printf("Total number of two-variable symmetries     = %8d.\n", symms)
}

func Ntk_NetworkSymmsPrint(n *Abc_Ntk_t, info *extrab.Extra_SymmInfo_t) {
	if n == nil || info == nil {
		fmt.Println()
		return
	}
	names := abc.Abc_NtkCollectCioNames(n, 0)
	taken := make([]bool, info.NVars)
	start := true
	for i := 0; i < info.NVars; i++ {
		if taken[i] {
			continue
		}
		group := []string{}
		for k := 0; k < info.NVars; k++ {
			if k == i || !info.Symms[i][k] {
				continue
			}
			if k < len(taken) && taken[k] {
				continue
			}
			if start {
				if info.Vars[i] >= 0 && info.Vars[i] < len(names) {
					group = append(group, names[info.Vars[i]])
				}
				taken[i] = true
				start = false
			}
			if info.Vars[k] >= 0 && info.Vars[k] < len(names) {
				group = append(group, names[info.Vars[k]])
			}
			taken[k] = true
		}
		if len(group) != 0 {
			fmt.Printf("  { %s }", strings.Join(group, " "))
			start = true
		}
	}
	fmt.Println()
}

func Ntk_SymTryRandomFlips(fun, npn []uint64, nVars int32) {
	rand := [...]int32{17290, 20203, 19027, 12035, 14687, 10920, 10413, 261, 2072, 16899, 4480, 6192, 3978, 8343, 745, 1370}
	words := int(util.Abc_TtWordNum(nVars))
	tmp := make([]uint64, words)
	util.Abc_TtCopy(tmp, fun, int32(words), 0)
	for _, v := range rand {
		if nVars > 1 {
			util.Abc_TtFlip(tmp, int32(words), v%(nVars-1))
		}
	}
	if util.Abc_TtCompareRev(npn, tmp, int32(words)) == 1 {
		panic("Ntk_SymTryRandomFlips: canonical truth table mismatch")
	}
}

func Ntk_SymFunDeriveNpn(fun []uint64, nVars int32, comp []int32) {
	words := int(util.Abc_TtWordNum(nVars))
	best := make([]uint64, words)
	tmp := make([]uint64, words)
	util.Abc_TtCopy(best, fun, int32(words), 1)
	if util.Abc_TtCompareRev(best, fun, int32(words)) == 1 {
		util.Abc_TtCopy(best, fun, int32(words), 0)
	}
	for _, v := range comp {
		util.Abc_TtFlip(fun, int32(words), v)
		if util.Abc_TtCompareRev(best, fun, int32(words)) == 1 {
			util.Abc_TtCopy(best, fun, int32(words), 0)
		}
		util.Abc_TtNot(fun, int32(words))
		if util.Abc_TtCompareRev(best, fun, int32(words)) == 1 {
			util.Abc_TtCopy(best, fun, int32(words), 0)
		}
	}
	copy(tmp, best)
	copy(fun, tmp)
}

func Ntk_SymFunGenerate(nVars, verbose int32) {
	if nVars < 1 || nVars > 16 {
		panic("Ntk_SymFunGenerate: variable count must be in [1,16]")
	}
	fmt.Printf("Generating truth tables of all symmetric functions of %d variables.\n", nVars)
	comp := extra.Extra_GreyCodeSchedule(nVars)
	mem := vec.Vec_MemAllocForTTSimple(int(nVars))
	for m := int32(0); m < (1 << (nVars + 1)); m++ {
		ones := make([]byte, nVars+1)
		for k := int32(0); k <= nVars; k++ {
			ones[k] = byte('0' + ((m >> k) & 1))
		}
		fun := util.Abc_TtSymFunGenerate(string(ones), nVars)
		if nVars < 6 {
			fun[0] = util.Abc_Tt6Stretch(fun[0], nVars)
		}
		if verbose != 0 {
			fmt.Printf("%s : ", string(ones))
			extra.Extra_PrintHex(&symmStdout{}, symmWords32(fun), nVars)
		}
		Ntk_SymFunDeriveNpn(fun, nVars, comp)
		class := vec.Vec_MemHashInsert(mem, fun)
		if verbose != 0 {
			fmt.Printf(" : NPN ")
			extra.Extra_PrintHex(&symmStdout{}, symmWords32(fun), nVars)
			fmt.Printf("  Class %3d\n", class)
		}
	}
	fmt.Printf("The number of different NPN classes is %d.\n", vec.Vec_MemEntryNum(mem))
	vec.Vec_MemHashFree(mem)
	vec.Vec_MemFree(mem)
}

type symmStdout struct{}

func (*symmStdout) Write(p []byte) (int, error) { return fmt.Print(string(p)) }
func symmWords32(words []uint64) []uint32 {
	out := make([]uint32, 2*len(words))
	for i, w := range words {
		out[2*i], out[2*i+1] = uint32(w), uint32(w>>32)
	}
	return out
}
