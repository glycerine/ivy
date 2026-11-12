package kit

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Kit_DsdVerify(ntk *Kit_DsdNtk_t, truth []uint32, nVars int32) bool {
	if int32(ntk.nVars) != nVars {
		return false
	}
	out := make([]uint32, Kit_TruthWordNum(nVars))
	Kit_DsdTruth(ntk, out)
	valid := Kit_TruthIsEqual(out, truth, nVars) != 0
	if !valid {
		fmt.Println("Verification failed.")
	}
	return valid
}
func Kit_DsdTest(truth []uint32, nVars int32) {
	ntk := Kit_DsdDecomposeExpand(truth, nVars)
	Kit_DsdPrint(os.Stdout, ntk)
	fmt.Println()
	Kit_DsdVerify(ntk, truth, nVars)
}
func Kit_DsdEval(truth []uint32, nVars, lutSize int32) int32 {
	ntk := Kit_DsdDecompose(truth, nVars)
	if !Kit_DsdVerify(ntk, truth, nVars) {
		return -1
	}
	return Kit_DsdCountLuts(ntk, lutSize)
}
func Kit_DsdTestCofs(ntk *Kit_DsdNtk_t, truth []uint32) int32 {
	n := int32(ntk.nVars)
	low, high := make([]uint32, Kit_TruthWordNum(n)), make([]uint32, Kit_TruthWordNum(n))
	result := int32(0)
	for v := int32(0); v < n; v++ {
		Kit_TruthCofactor0New(low, truth, n, v)
		Kit_TruthCofactor1New(high, truth, n, v)
		a, b := Kit_DsdDecomposeExpand(low, n), Kit_DsdDecomposeExpand(high, n)
		fmt.Printf("Cof%d0: ", v)
		Kit_DsdPrint(os.Stdout, a)
		fmt.Printf("\nCof%d1: ", v)
		Kit_DsdPrint(os.Stdout, b)
		fmt.Println()
		if n == 4 && Kit_DsdCheckVar4Dec2(a, b) != 0 {
			result = 1
		}
	}
	return result
}
func Kit_DsdPrecompute4Vars() error {
	file, err := os.Open("5npn/npn4.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	index, nonDsd, implementable := 0, 0, 0
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 6 {
			return fmt.Errorf("kit: malformed NPN class line %d", index+1)
		}
		value, err := strconv.ParseUint(line[2:6], 16, 16)
		if err != nil {
			return err
		}
		truth := []uint32{uint32(value) * 0x00010001}
		ntk := Kit_DsdDecomposeExpand(truth, 4)
		if Kit_DsdFindLargeBox(ntk, 3) != 0 {
			ret := Kit_DsdTestCofs(ntk, truth)
			fmt.Printf("%3d : Non-DSD function %04x implementable=%d\n", index, value, ret)
			nonDsd++
			implementable += int(ret)
		}
		if !Kit_DsdVerify(ntk, truth, 4) {
			return fmt.Errorf("kit: verification failed for NPN class %d", index)
		}
		index++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	fmt.Printf("non-DSD = %d implementable = %d\n", nonDsd, implementable)
	return nil
}
func Kit_DsdCofactoringGetVars(networks []*Kit_DsdNtk_t, count int32, out []int32) int32 {
	seen := map[int32]bool{}
	n := int32(0)
	for _, ntk := range networks[:count] {
		for _, obj := range ntk.pNodes {
			if obj.Type != KIT_DSD_PRIME || obj.nFans == 3 {
				continue
			}
			for _, lit := range obj.pFans[:obj.nFans] {
				v := int32(lit >> 1)
				if v >= int32(ntk.nVars) || seen[v] {
					continue
				}
				if n >= int32(len(out)) {
					panic("kit: short cofactor variable result")
				}
				out[n] = v
				n++
				seen[v] = true
			}
		}
	}
	return n
}
func kitDsdCofs(tables [][]uint32, nVars, v int32) [][]uint32 {
	out := make([][]uint32, 0, 2*len(tables))
	for _, truth := range tables {
		a, b := make([]uint32, Kit_TruthWordNum(nVars)), make([]uint32, Kit_TruthWordNum(nVars))
		Kit_TruthCofactor0New(a, truth, nVars, v)
		Kit_TruthCofactor1New(b, truth, nVars, v)
		out = append(out, a, b)
	}
	return out
}
func Kit_DsdCofactoring(truth []uint32, nVars int32, variables []int32, limit, verbose int32) int32 {
	if limit < 0 || limit >= 5 {
		panic("kit: cofactoring supports at most four levels")
	}
	tables := [][]uint32{append([]uint32(nil), truth[:Kit_TruthWordNum(nVars)]...)}
	for step := int32(0); step < limit; step++ {
		networks := make([]*Kit_DsdNtk_t, len(tables))
		for i, table := range tables {
			networks[i] = Kit_DsdDecompose(table, nVars)
		}
		candidates := make([]int32, nVars)
		count := Kit_DsdCofactoringGetVars(networks, int32(len(networks)), candidates)
		if count == 0 {
			return step
		}
		best, primeBest, supportBest := int32(-1), int32(KIT_INFINITY), int32(KIT_INFINITY)
		for _, v := range candidates[:count] {
			var prime, support int32
			for _, cof := range kitDsdCofs(tables, nVars, v) {
				ntk := Kit_DsdDecompose(cof, nVars)
				prime = max(prime, Kit_DsdNonDsdSizeMax(ntk))
				support += Kit_TruthSupportSize(cof, nVars)
			}
			if prime < primeBest || prime == primeBest && support < supportBest {
				best, primeBest, supportBest = v, prime, support
			}
		}
		if variables != nil {
			if step >= int32(len(variables)) {
				panic("kit: short cofactor variable buffer")
			}
			variables[step] = best
		}
		tables = kitDsdCofs(tables, nVars, best)
		if verbose != 0 {
			for i, cof := range tables {
				fmt.Printf("Cof%d%d: ", step+1, i)
				Kit_DsdPrint(os.Stdout, Kit_DsdDecomposeExpand(cof, nVars))
				fmt.Println()
			}
		}
	}
	return limit
}
func Kit_DsdPrintCofactors(truth []uint32, nVars, level, verbose int32) {
	if level < 0 || level > 4 || level > nVars {
		panic("kit: invalid cofactor print level")
	}
	var chosen []int32
	var visit func(int32)
	visit = func(start int32) {
		if int32(len(chosen)) < level {
			for v := start; v < nVars; v++ {
				chosen = append(chosen, v)
				visit(v + 1)
				chosen = chosen[:len(chosen)-1]
			}
			return
		}
		tables := [][]uint32{truth}
		var names []string
		for _, v := range chosen {
			names = append(names, string(rune('a'+v)))
			tables = kitDsdCofs(tables, nVars, v)
		}
		fmt.Printf("Variables { %s }\n", strings.Join(names, " "))
		var prime, support int32
		for i, cof := range tables {
			ntk := Kit_DsdDecomposeExpand(cof, nVars)
			prime = max(prime, Kit_DsdNonDsdSizeMax(ntk))
			support += Kit_TruthSupportSize(cof, nVars)
			if verbose != 0 {
				fmt.Printf("Cof%d%d: ", level, i)
				Kit_DsdPrint(os.Stdout, ntk)
				fmt.Println()
			}
		}
		fmt.Printf("Max = %2d. Supps = %2d.\n", prime, support)
	}
	visit(0)
}
