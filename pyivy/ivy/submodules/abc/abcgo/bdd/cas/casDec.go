// Port of casDec.c, Alan Mishchenko, ABC (2005).
package cas

import (
	"fmt"
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	extra "github.com/glycerine/abcgo/bdd/extrab/cascade"
	"io"
	"math/bits"
	"os"
)

// LUT records the source algorithm's input split, column multiplicity and
// encoding relation. Previous outputs occupy the top nInsP manager levels.
type LUT struct {
	nIns, nInsP, nCols, nMulti, nSimple, Level int
	pbCols, pbCodes, paNodes                   []*DdNode
	bRelation                                  *DdNode
}

// Decompose constructs the original level-cut LUT cascade. The manager order
// stays fixed while encoding variables are appended below its functional vars.
func Decompose(dd *DdManager, f *DdNode, nNames, lutSize int) (luts []*LUT, codes []*DdNode, err error) {
	if dd == nil || f == nil || nNames < 0 || nNames > len(dd.Vars) || lutSize < 2 || lutSize > 31 {
		return nil, nil, fmt.Errorf("cas: invalid decomposition parameters")
	}
	if !cudd.Cudd_IsAdd(f) || cudd.Cudd_addBddPattern(dd, f) == nil {
		return nil, nil, fmt.Errorf("cas: decomposition expects a 0/1 ADD")
	}
	seen := map[*DdNode]bool{}
	var check func(*DdNode) bool
	check = func(n *DdNode) bool {
		if seen[n] {
			return true
		}
		seen[n] = true
		if cudd.Cudd_IsConstant(n) {
			return n.Value == 0 || n.Value == 1
		}
		return dd.Perm[n.Index] < nNames && check(n.T) && check(n.E)
	}
	if !check(f) {
		return nil, nil, fmt.Errorf("cas: ADD is not Boolean or references unnamed inputs")
	}
	profile := extra.Extra_ProfileWidth(dd, f)
	remaining, previous := nNames, 0
	for {
		p := &LUT{nInsP: previous, Level: nNames - remaining}
		last := remaining+previous <= lutSize
		if last {
			p.nIns = remaining + previous
			p.nCols = 2
			p.nMulti = 1
			remaining = 0
		} else {
			p.nIns = lutSize
			consumed := lutSize - previous
			if consumed <= 0 {
				return nil, nil, fmt.Errorf("cas: LUT size is too small")
			}
			remaining -= consumed
			p.nCols = profile[nNames-remaining]
			p.nMulti = bits.Len(uint(p.nCols - 1))
		}
		if p.nMulti >= lutSize {
			return nil, nil, fmt.Errorf("cas: LUT size is too small")
		}
		previous = p.nMulti
		luts = append(luts, p)
		if last {
			break
		}
	}
	codes = make([]*DdNode, lutSize)
	for i := range codes {
		codes[i] = cudd.Cudd_bddNewVar(dd)
	}
	nodes, paths := []*DdNode{f}, []*DdNode{dd.One}
	for i, p := range luts {
		cut := p.Level + p.nIns - p.nInsP
		result := extra.Extra_bddNodePathsUnderCutArray(dd, nodes, paths, cut)
		p.paNodes, p.pbCols = result.Nodes, result.Columns
		p.nCols = len(result.Nodes)
		inputVars := make([]*DdNode, 0, p.nIns)
		for v := 0; v < p.nInsP; v++ {
			inputVars = append(inputVars, cudd.Cudd_bddIthVar(dd, dd.InvPerm[v]))
		}
		for v := 0; v < p.nIns-p.nInsP; v++ {
			inputVars = append(inputVars, cudd.Cudd_bddIthVar(dd, dd.InvPerm[p.Level+v]))
		}
		cube := extra.Extra_bddBitsToCube(dd, (1<<len(inputVars))-1, len(inputVars), inputVars, true)
		if i == len(luts)-1 {
			// Unlike the C implementation, constant functions need only one column.
			// Construct the final bit directly from every true leaf's path.
			value := dd.LogicZero
			for c, n := range p.paNodes {
				if n.Value == 1 {
					value = cudd.Cudd_bddOr(dd, value, p.pbCols[c])
				}
			}
			p.bRelation = cudd.Cudd_bddXnor(dd, codes[0], value)
		} else {
			p.bRelation, p.nSimple = extra.Extra_bddEncodingNonStrict(dd, p.pbCols, cube, codes[:p.nMulti])
		}
		if p.bRelation == nil {
			return nil, nil, fmt.Errorf("cas: failed to encode stage %d", i)
		}
		permutation := identity(len(dd.Vars))
		for v := 0; v < p.nMulti; v++ {
			permutation[codes[v].Index] = dd.InvPerm[v]
		}
		p.pbCodes = make([]*DdNode, p.nCols)
		for c, col := range p.pbCols {
			code := cudd.Cudd_bddAndAbstract(dd, p.bRelation, col, cube)
			p.pbCodes[c] = cudd.Cudd_bddPermute(dd, code, permutation)
		}
		nodes, paths = p.paNodes, p.pbCodes
	}
	return luts, codes, nil
}

// stageFunctions extracts each encoded output bit by quantifying the other
// encoding variables, exactly as WriteLUTSintoBLIFfile in the C source.
func stageFunctions(dd *DdManager, p *LUT, codes []*DdNode) []*DdNode {
	cube := extra.Extra_bddBitsToCube(dd, (1<<p.nMulti)-1, p.nMulti, codes, true)
	funcs := make([]*DdNode, p.nMulti)
	for o := range funcs {
		cof := cudd.Cudd_Cofactor(dd, p.bRelation, codes[o])
		funcs[o] = cudd.Cudd_bddExistAbstract(dd, cof, cube)
	}
	return funcs
}

// VerifyCascade symbolically substitutes previous-stage outputs. This replaces
// the source CLI's external cec command with a direct canonical BDD check.
func VerifyCascade(dd *DdManager, f *DdNode, luts []*LUT, codes []*DdNode) bool {
	var previous []*DdNode
	for _, p := range luts {
		current := stageFunctions(dd, p, codes)
		// Simultaneous substitution uses fresh code variables to avoid replacing
		// variables introduced by an earlier substitution in this same stage.
		for i := range current {
			permutation := identity(len(dd.Vars))
			for k := range previous {
				permutation[dd.InvPerm[k]] = int(codes[k].Index)
			}
			current[i] = cudd.Cudd_bddPermute(dd, current[i], permutation)
			for k, g := range previous {
				current[i] = cudd.Cudd_bddCompose(dd, current[i], g, int(codes[k].Index))
			}
		}
		previous = current
	}
	return len(previous) == 1 && previous[0] == cudd.Cudd_addBddPattern(dd, f)
}
func WriteLUTSintoBLIFfile(w io.Writer, dd *DdManager, luts []*LUT, nLuts int, codes []*DdNode, names []string, nNames int, fileName string) {
	for i, p := range luts[:nLuts] {
		fmt.Fprintf(w, "#----------------- LUT #%d ----------------------\n", i)
		local := make([]string, len(dd.Vars))
		for v := 0; v < p.nInsP; v++ {
			local[dd.InvPerm[v]] = fmt.Sprintf("LUT%02d_%02d", i-1, v)
		}
		for v := 0; v < p.nIns-p.nInsP; v++ {
			index := dd.InvPerm[p.Level+v]
			local[index] = names[index]
		}
		for o, f := range stageFunctions(dd, p, codes) {
			output := fmt.Sprintf("LUT%02d_%02d", i, o)
			if i == nLuts-1 {
				output = "F"
			}
			WriteDDintoBLIFfile(w, f, output, fmt.Sprintf("L%02d_%02d_", i, o), local)
		}
	}
}
func CreateDecomposedNetwork(dd *DdManager, f *DdNode, names []string, nNames int, fileName string, lutSize int, check, verbose bool) int {
	if nNames < 0 || nNames > len(names) {
		return 0
	}
	luts, codes, err := Decompose(dd, f, nNames, lutSize)
	if err != nil {
		if verbose {
			fmt.Println(err)
		}
		return 0
	}
	if check {
		if !VerifyCascade(dd, f, luts, codes) {
			return 0
		}
		file, err := os.Create(fileName)
		if err != nil {
			return 0
		}
		writeHeader(file, fileName, names[:nNames])
		WriteLUTSintoBLIFfile(file, dd, luts, len(luts), codes, names, nNames, fileName)
		fmt.Fprintln(file, ".end")
		if file.Close() != nil {
			return 0
		}
	}
	if verbose {
		outputs, simple := 0, 0
		for i, p := range luts {
			outputs += p.nMulti
			simple += p.nSimple
			fmt.Printf("Stage %3d: In = %3d  InP = %3d  Cols = %5d  Multi = %2d  Simple = %2d  Level = %3d\n", i+1, p.nIns, p.nInsP, p.nCols, p.nMulti, p.nSimple, p.Level)
		}
		fmt.Printf("LUTs: Total = %5d. Final = %5d. Simple = %5d.\n", outputs, outputs-simple, simple)
	}
	return 1
}
