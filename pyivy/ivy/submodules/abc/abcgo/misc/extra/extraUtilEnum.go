package extra

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [extraUtilEnum.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Function enumeration.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilEnum.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

type enumDataset struct {
	variables           int32
	minterms, functions []string
}

var enumDatasets = []enumDataset{
	{8, []string{"1-1-1-1-", "1-1--11-", "1-1-1--1", "1-1--1-1", "-11-1-1-", "-11--11-", "-11-1--1", "-11--1-1", "1--11-1-", "1--1-11-", "1--11--1", "1--1-1-1", "-1-11-1-", "-1-1-11-", "-1-11--1", "-1-1-1-1"}, []string{"1111101011111010", "0000010100000101", "1111110010101001", "0000001101010110", "1111111111001101", "0000000000110010", "1111111111111110", "0000000000000001"}},
	{10, []string{"1-1---1---", "1-1----1--", "1-1-----1-", "1-1------1", "1--1--1---", "1--1---1--", "1--1----1-", "1--1-----1", "1---1-1---", "1---1--1--", "1---1---1-", "1---1----1", "1----11---", "1----1-1--", "1----1--1-", "1----1---1", "-11---1---", "-11----1--", "-11-----1-", "-11------1", "-1-1--1---", "-1-1---1--", "-1-1----1-", "-1-1-----1", "-1--1-1---", "-1--1--1--", "-1--1---1-", "-1--1----1", "-1---11---", "-1---1-1--", "-1---1--1-", "-1---1---1"}, []string{"11111110110010001110110010000000", "00000001001101110001001101111111", "10000001001001000001001001001000", "01001000000100101000000100100100", "00100100100000010100100000010010", "00010010010010000010010010000001", "11111111111111111111000000000000"}},
	{8, []string{"1---1---", "1----1--", "1-----1-", "1------1", "-1--1---", "-1---1--", "-1----1-", "-1-----1", "--1-1---", "--1--1--", "--1---1-", "--1----1", "---11---", "---1-1--", "---1--1-", "---1---1"}, []string{"1111111011001000", "0000000100110111", "1000000100100100", "0100100000010010", "0010010010000001", "0001001001001000", "1111111111111111"}},
}

func enumDecodeDataset(id int, vars, mints, funcs []uint32) (int32, int32, int32) {
	data := enumDatasets[id]
	vars = vars[:data.variables]
	mints = mints[:len(data.minterms)]
	funcs = funcs[:len(data.functions)]
	for i, mint := range data.minterms {
		for k := int32(0); k < data.variables; k++ {
			if mint[k] == '1' {
				mints[i] |= 1 << uint(k)
				vars[k] |= 1 << uint(i)
			}
		}
	}
	for i, function := range data.functions {
		for k, ch := range function {
			if ch == '1' {
				funcs[i] |= 1 << uint(k)
			}
		}
	}
	return data.variables, int32(len(data.minterms)), int32(len(data.functions))
}
func Abc_GetFirst(vars, mints, funcs []uint32) (int32, int32, int32) {
	return enumDecodeDataset(0, vars, mints, funcs)
}
func Abc_GetSecond(vars, mints, funcs []uint32) (int32, int32, int32) {
	return enumDecodeDataset(1, vars, mints, funcs)
}
func Abc_GetThird(vars, mints, funcs []uint32) (int32, int32, int32) {
	return enumDecodeDataset(2, vars, mints, funcs)
}
func enumPrintNode(out io.Writer, gates *vec.Vec_Int_t, node, vars int32) {
	if node < 0 {
		panic("negative enumeration node")
	}
	if node < vars {
		fmt.Fprintf(out, "%c", 'a'+node)
		return
	}
	a, b := vec.Vec_IntEntry(gates, 2*int(node)), vec.Vec_IntEntry(gates, 2*int(node)+1)
	op := ""
	if a >= b {
		op = "+"
	}
	if a > b {
		a, b = b, a
	}
	if a >= node || b >= node {
		panic("non-topological enumeration gate")
	}
	if a < vars {
		enumPrintNode(out, gates, a, vars)
	} else {
		fmt.Fprint(out, "(")
		enumPrintNode(out, gates, a, vars)
		fmt.Fprint(out, ")")
	}
	fmt.Fprint(out, op)
	if b < vars {
		enumPrintNode(out, gates, b, vars)
	} else {
		fmt.Fprint(out, "(")
		enumPrintNode(out, gates, b, vars)
		fmt.Fprint(out, ")")
	}
}
func Abc_EnumPrint_rec(gates *vec.Vec_Int_t, node, vars int32) {
	enumPrintNode(os.Stdout, gates, node, vars)
}
func Abc_EnumPrint(gates *vec.Vec_Int_t, node, vars int32) {
	Abc_EnumPrint_rec(gates, node, vars)
	fmt.Println()
}

type enumFormulaMatch struct{ index, gates int }
type enumFormulaResult struct {
	truths     []uint32
	gates      *vec.Vec_Int_t
	gateCounts []int
	found      []enumFormulaMatch
}

func enumFormulaAppend(result *enumFormulaResult, seen map[uint32]int, truth uint32, a, b int) bool {
	if _, exists := seen[truth]; exists {
		return false
	}
	seen[truth] = len(result.truths)
	result.truths = append(result.truths, truth)
	vec.Vec_IntPush(result.gates, int32(a))
	vec.Vec_IntPush(result.gates, int32(b))
	return true
}
func enumFunctions(variables, targets []uint32, maxGates int, verbose bool) enumFormulaResult {
	if maxGates < 0 {
		panic("negative enumeration depth")
	}
	result := enumFormulaResult{gates: vec.Vec_IntAlloc(100), gateCounts: make([]int, maxGates+2), found: make([]enumFormulaMatch, len(targets))}
	for i := range result.found {
		result.found[i].index = -1
	}
	seen := map[uint32]int{}
	done := 0
	for i, value := range variables {
		if _, exists := seen[value]; !exists {
			seen[value] = i
		}
		result.truths = append(result.truths, value)
		vec.Vec_IntPush(result.gates, -1)
		vec.Vec_IntPush(result.gates, -1)
	}
	for target, value := range targets {
		if index, exists := seen[value]; exists {
			result.found[target] = enumFormulaMatch{index, 0}
			done++
		}
	}
	result.gateCounts[1] = len(variables)
	started := time.Now()
	for cost := 1; cost <= maxGates && (len(targets) == 0 || done < len(targets)); cost++ {
		for a := 0; a < cost; a++ {
			b := cost - 1 - a
			if b < a {
				continue
			}
			if verbose {
				fmt.Printf("Trying %d + %d + 1 = %d\n", a, b, cost)
			}
			for i := result.gateCounts[a]; i < result.gateCounts[a+1]; i++ {
				for k := result.gateCounts[b]; k < result.gateCounts[b+1]; k++ {
					if i >= k {
						continue
					}
					for operation := 0; operation < 2; operation++ {
						truth := result.truths[i] & result.truths[k]
						fan0, fan1 := i, k
						if operation != 0 {
							truth = result.truths[i] | result.truths[k]
							fan0, fan1 = k, i
						}
						if !enumFormulaAppend(&result, seen, truth, fan0, fan1) {
							continue
						}
						index := len(result.truths) - 1
						for target, value := range targets {
							if result.found[target].index >= 0 || truth != value {
								continue
							}
							result.found[target] = enumFormulaMatch{index, cost}
							done++
							if verbose {
								fmt.Printf("Found function %d with %d gates: ", target, cost)
								Abc_EnumPrint(result.gates, int32(index), int32(len(variables)))
							}
						}
					}
				}
			}
		}
		result.gateCounts[cost+1] = len(result.truths)
		if verbose {
			fmt.Printf("Finished %d gates. Truths = %10d. Time = %v\n", cost, len(result.truths), time.Since(started))
		}
	}
	return result
}
func Abc_EnumerateFunctions(max int32) {
	var variables, mints, functions [100]uint32
	n, _, count := Abc_GetSecond(variables[:], mints[:], functions[:])
	enumFunctions(variables[:n], functions[:count], int(max), true)
}

const ABC_ENUM_MAX = 16

var enumElementaryTruths = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

type Abc_EnuMan_t_ struct {
	nVars, nVarsFree, fVerbose, fUseXor, nNodeMax, nNodes, nTops int32
	pFans0, pFans1, fCompl0, fCompl1, Polar, pRefs, pLevel       [ABC_ENUM_MAX]int32
	pTruths                                                      [ABC_ENUM_MAX]uint64
	nTries, nBuilds, nFinished                                   uint64
}
type Abc_EnuMan_t = Abc_EnuMan_t_

func Abc_EnumRef(p *Abc_EnuMan_t, i int32) {
	if p.pRefs[i] < 0 {
		panic("negative enumeration reference count")
	}
	if p.pRefs[i] == 0 {
		p.nTops--
	}
	p.pRefs[i]++
}
func Abc_EnumDeref(p *Abc_EnuMan_t, i int32) {
	if p.pRefs[i] <= 0 {
		panic("enumeration reference underflow")
	}
	p.pRefs[i]--
	if p.pRefs[i] == 0 {
		p.nTops++
	}
}
func Abc_EnumRefNode(p *Abc_EnuMan_t, i int32) {
	Abc_EnumRef(p, p.pFans0[i])
	Abc_EnumRef(p, p.pFans1[i])
	p.nTops++
	p.nNodes++
	if i >= p.nNodes {
		panic("invalid enumeration node")
	}
}
func Abc_EnumDerefNode(p *Abc_EnuMan_t, i int32) {
	if i >= p.nNodes {
		panic("invalid enumeration node")
	}
	Abc_EnumDeref(p, p.pFans0[i])
	Abc_EnumDeref(p, p.pFans1[i])
	p.nTops--
	p.nNodes--
}
func Abc_EnumPrintOne(p *Abc_EnuMan_t) {
	value := p.pTruths[p.nNodes-1]
	kit.Kit_DsdPrintFromTruth([]uint32{uint32(value), uint32(value >> 32)}, p.nVars)
	for i := p.nVars; i < p.nNodes; i++ {
		if p.Polar[i] == 4 {
			fmt.Printf("  %c=%c+%c", 'a'+i, 'a'+p.pFans0[i], 'a'+p.pFans1[i])
		} else {
			a, b := "", ""
			if p.fCompl0[i] != 0 {
				a = "!"
			}
			if p.fCompl1[i] != 0 {
				b = "!"
			}
			fmt.Printf("  %c=%s%c%s%c", 'a'+i, a, 'a'+p.pFans0[i], b, 'a'+p.pFans1[i])
		}
	}
	fmt.Println()
}
func Abc_EnumEquiv(a, b uint64) int32 {
	if a == b || a == ^b {
		return 1
	}
	return 0
}
func Abc_EnumerateFilter(p *Abc_EnuMan_t) int32 {
	n := p.nNodes
	i, k := p.pFans0[n], p.pFans1[n]
	if i >= k {
		panic("unordered enumeration fanins")
	}
	truth := p.pTruths[n]
	if Abc_EnumEquiv(truth, 0) != 0 {
		return 1
	}
	for v := int32(0); v < n; v++ {
		if Abc_EnumEquiv(truth, p.pTruths[v]) != 0 {
			return 1
		}
	}
	for a := int32(0); a <= i; a++ {
		for b := a + 1; b <= k; b++ {
			if a == i && b == k {
				continue
			}
			x, y := p.pTruths[a], p.pTruths[b]
			if Abc_EnumEquiv(truth, x&y) != 0 || Abc_EnumEquiv(truth, x&^y) != 0 || Abc_EnumEquiv(truth, ^x&y) != 0 || Abc_EnumEquiv(truth, ^x&^y) != 0 || p.fUseXor != 0 && Abc_EnumEquiv(truth, x^y) != 0 {
				return 1
			}
		}
	}
	return 0
}
func Abc_EnumerateFuncs_rec(p *Abc_EnuMan_t, newLevel, firstNode int32) {
	if p.nNodes == p.nNodeMax {
		if p.nTops != 1 {
			panic("enumerated graph has multiple outputs")
		}
		if p.fVerbose != 0 {
			Abc_EnumPrintOne(p)
		}
		p.nFinished++
		return
	}
	n := p.nNodes
	referencedFans := p.nNodeMax - n + 1 - p.nTops
	high0, high1, low0, c0 := firstNode, n, int32(0), int32(0)
	if newLevel == 0 {
		high0, high1, low0, c0 = p.pFans1[n-1], firstNode, p.pFans0[n-1], p.Polar[n-1]
	}
	level := p.pLevel[high0]
	if p.nTops <= 0 || p.nTops > p.nNodeMax-n+1 {
		panic("invalid enumeration frontier")
	}
	for k := high0; k < high1; k++ {
		if referencedFans == 0 && p.pRefs[k] > 0 {
			continue
		}
		if p.pRefs[k] > 0 {
			referencedFans--
		}
		if referencedFans < 0 {
			panic("enumeration fanout budget underflow")
		}
		low := int32(0)
		if k == high0 {
			low = low0
		}
		for i := low; i < k; i++ {
			if referencedFans == 0 && p.pRefs[i] > 0 {
				continue
			}
			if level == 0 && p.pRefs[i] == 0 && p.pRefs[k] == 0 && (i+1 != k || i > 0 && p.pRefs[i-1] == 0) {
				continue
			}
			if p.pLevel[k] == 0 && p.pRefs[k] == 0 && p.pRefs[i] != 0 && k > 0 && p.pRefs[k-1] == 0 {
				continue
			}
			start := int32(0)
			if k == high0 && i == low0 && newLevel == 0 {
				start = c0 + 1
			}
			for c := start; c < 4+p.fUseXor; c++ {
				if p.pLevel[i] == 0 && p.pRefs[i] == 0 && c&1 != 0 || p.pLevel[k] == 0 && p.pRefs[k] == 0 && c&2 != 0 {
					continue
				}
				p.nTries++
				p.pFans0[n], p.pFans1[n] = i, k
				p.fCompl0[n], p.fCompl1[n] = c&1, c>>1&1
				p.Polar[n] = c
				x, y := p.pTruths[i], p.pTruths[k]
				if c == 4 {
					p.pTruths[n] = x ^ y
				} else {
					if c&1 != 0 {
						x = ^x
					}
					if c&2 != 0 {
						y = ^y
					}
					p.pTruths[n] = x & y
				}
				if Abc_EnumerateFilter(p) != 0 {
					continue
				}
				p.nBuilds++
				if level != max(p.pLevel[i], p.pLevel[k]) {
					panic("inconsistent enumeration level")
				}
				p.pLevel[n] = level + 1
				Abc_EnumRefNode(p, n)
				nextFirst := firstNode
				if newLevel != 0 {
					nextFirst = n
				}
				Abc_EnumerateFuncs_rec(p, 0, nextFirst)
				Abc_EnumDerefNode(p, n)
				if n != p.nNodes {
					panic("enumeration backtracking mismatch")
				}
			}
		}
		if p.pRefs[k] > 0 {
			referencedFans++
		}
	}
	if newLevel == 0 {
		Abc_EnumerateFuncs_rec(p, 1, firstNode)
	}
}
func enumAigGraphs(vars, gates, verbose int32) *Abc_EnuMan_t {
	if vars < 2 || vars > 6 || gates <= 0 || gates >= ABC_ENUM_MAX-vars {
		panic("invalid AIG enumeration dimensions")
	}
	if vars > gates+1 {
		return nil
	}
	p := &Abc_EnuMan_t{nVars: vars, nNodeMax: vars + gates, nNodes: vars, nTops: vars, fVerbose: verbose}
	copy(p.pTruths[:], enumElementaryTruths[:vars])
	Abc_EnumerateFuncs_rec(p, 1, 0)
	if p.nNodes != vars || p.nTops != vars {
		panic("enumeration state not restored")
	}
	return p
}
func Abc_EnumerateFuncs(vars, gates, verbose int32) {
	if vars > gates+1 {
		fmt.Printf("The gate count %d is not enough to have functions with %d inputs.\n", gates, vars)
		return
	}
	started := time.Now()
	p := enumAigGraphs(vars, gates, verbose)
	fmt.Printf("Vars = %d. Gates = %d. Tries = %d. Builds = %d. Finished = %d. Time = %v\n", vars, gates, p.nTries, p.nBuilds, p.nFinished, time.Since(started))
}
