package abci

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
)

type mvMan struct {
	nInputs, nFuncs int
	dd              *core.DdManager
	bValues         [15][4]*core.DdNode
	bValueDcs       [15][4]*core.DdNode
	bFuncs          [15]*core.DdNode
}

// Abc_MvExperiment reads input.pla and reports/decomposes its multi-valued functions.
func Abc_MvExperiment() {
	p := &mvMan{dd: core.Cudd_Init(32, 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0), nFuncs: 15, nInputs: 9}
	if p.dd == nil {
		panic("Abc_MvExperiment: cannot create BDD manager")
	}
	defer extrab.Extra_StopManager(p.dd)
	Abc_MvRead(p)
	Abc_MvPrintStats(p)
	Abc_MvDecompose(p)
	Abc_MvDeref(p)
}

// Abc_MvPrintStats prints DAG and support sizes for each output and value set.
func Abc_MvPrintStats(p *mvMan) {
	for i := 0; i < 15; i++ {
		fmt.Printf("%2d : ", i)
		fmt.Printf("%3d (%2d)    ", core.Cudd_DagSize(p.bFuncs[i])-1, core.Cudd_SupportSize(p.dd, p.bFuncs[i]))
		for v := 0; v < 4; v++ {
			fmt.Printf("%d = %3d (%2d)  ", v, core.Cudd_DagSize(p.bValues[i][v])-1, core.Cudd_SupportSize(p.dd, p.bValues[i][v]))
		}
		fmt.Println()
	}
}

// Abc_MvReadCube translates a '-'/'0'/'1' input cube. Variables occupy DD indices 29 down.
func Abc_MvReadCube(dd *core.DdManager, line string, nVars int) *core.DdNode {
	if dd == nil || nVars < 0 || nVars > 30 || len(line) < nVars {
		panic("Abc_MvReadCube: invalid manager or cube width")
	}
	cube := core.Cudd_ReadOne(dd)
	core.Cudd_Ref(cube)
	for i := 0; i < nVars; i++ {
		var variable *core.DdNode
		switch line[i] {
		case '-':
			continue
		case '0':
			variable = core.Cudd_Not(core.Cudd_bddIthVar(dd, 29-i))
		case '1':
			variable = core.Cudd_bddIthVar(dd, 29-i)
		default:
			panic("Abc_MvReadCube: invalid cube character")
		}
		temp := cube
		cube = core.Cudd_bddAnd(dd, temp, variable)
		if cube == nil {
			panic("Abc_MvReadCube: BDD conjunction failed")
		}
		core.Cudd_Ref(cube)
		core.Cudd_RecursiveDeref(dd, temp)
	}
	core.Cudd_Deref(cube)
	return cube
}

// Abc_MvRead reads the source's fixed input.pla format.
func Abc_MvRead(p *mvMan) {
	file, err := os.Open("input.pla")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := abcMvReadFrom(p, file); err != nil {
		panic(err)
	}
}

func abcMvReadFrom(p *mvMan, input *os.File) error {
	return abcMvReadScanner(p, bufio.NewScanner(input))
}

func abcMvReadScanner(p *mvMan, scanner *bufio.Scanner) error {
	scanner.Buffer(make([]byte, 1000), 1000)
	dd := p.dd
	cubeSum := core.Cudd_ReadLogicZero(dd)
	core.Cudd_Ref(cubeSum)
	for i := 0; i < 15; i++ {
		for v := 0; v < 4; v++ {
			p.bValues[i][v] = core.Cudd_ReadLogicZero(dd)
			p.bValueDcs[i][v] = core.Cudd_ReadLogicZero(dd)
			core.Cudd_Ref(p.bValues[i][v])
			core.Cudd_Ref(p.bValueDcs[i][v])
		}
	}
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if line[0] == '.' {
			if len(line) > 1 && line[1] == 'e' {
				break
			}
			continue
		}
		if len(line) < 49 {
			return fmt.Errorf("Abc_MvRead: short PLA row (%d chars)", len(line))
		}
		cube := Abc_MvReadCube(dd, line, 18)
		core.Cudd_Ref(cube)
		outputs := line[19:]
		for i := 0; i < 15; i++ {
			a, b := outputs[2*i], outputs[2*i+1]
			if a == '-' && b == '-' {
				for v := 0; v < 4; v++ {
					p.bValueDcs[i][v] = abcMvOr(dd, p.bValueDcs[i][v], cube)
				}
				continue
			}
			v := -1
			switch string([]byte{a, b}) {
			case "00":
				v = 0
			case "10":
				v = 1
			case "01":
				v = 2
			case "11":
				v = 3
			default:
				return fmt.Errorf("Abc_MvRead: invalid output value %q", []byte{a, b})
			}
			p.bValues[i][v] = abcMvOr(dd, p.bValues[i][v], cube)
		}
		cubeSum = abcMvOr(dd, cubeSum, cube)
		core.Cudd_RecursiveDeref(dd, cube)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for i := 0; i < 15; i++ {
		for v := 0; v < 4; v++ {
			if p.bValues[i][v] == core.Cudd_Not(core.Cudd_ReadOne(dd)) {
				continue
			}
			p.bValues[i][v] = abcMvOr(dd, p.bValues[i][v], p.bValueDcs[i][v])
			p.bValues[i][v] = abcMvOr(dd, p.bValues[i][v], core.Cudd_Not(cubeSum))
		}
	}
	all := core.Cudd_CountMinterm(dd, core.Cudd_ReadOne(dd), 32)
	fmt.Printf("Domain = %5.2f %%.\n", 100.0*core.Cudd_CountMinterm(dd, cubeSum, 32)/all)
	core.Cudd_RecursiveDeref(dd, cubeSum)
	for i := 0; i < 15; i++ {
		p.bFuncs[i] = core.Cudd_ReadLogicZero(dd)
		core.Cudd_Ref(p.bFuncs[i])
		for v := 0; v < 4; v++ {
			var0 := core.Cudd_NotCond(core.Cudd_bddIthVar(dd, 30), v&1 == 0)
			var1 := core.Cudd_NotCond(core.Cudd_bddIthVar(dd, 31), v&2 == 0)
			cube := core.Cudd_bddAnd(dd, var0, var1)
			core.Cudd_Ref(cube)
			prod := core.Cudd_bddAnd(dd, p.bValues[i][v], cube)
			core.Cudd_Ref(prod)
			core.Cudd_RecursiveDeref(dd, cube)
			p.bFuncs[i] = abcMvOr(dd, p.bFuncs[i], prod)
			core.Cudd_RecursiveDeref(dd, prod)
		}
	}
	return nil
}

func abcMvOr(dd *core.DdManager, old, add *core.DdNode) *core.DdNode {
	result := core.Cudd_bddOr(dd, old, add)
	if result == nil {
		panic("Abc_MvRead: BDD disjunction failed")
	}
	core.Cudd_Ref(result)
	core.Cudd_RecursiveDeref(dd, old)
	return result
}

// Abc_MvDeref releases the functions and value sets owned by the manager.
func Abc_MvDeref(p *mvMan) {
	for i := 0; i < 15; i++ {
		for v := 0; v < 4; v++ {
			core.Cudd_RecursiveDeref(p.dd, p.bValues[i][v])
			core.Cudd_RecursiveDeref(p.dd, p.bValueDcs[i][v])
		}
		core.Cudd_RecursiveDeref(p.dd, p.bFuncs[i])
	}
}

// Abc_MvDecompose counts distinct cofactors for each output and every pair of 4-valued inputs.
func Abc_MvDecompose(p *mvMan) {
	dd := p.dd
	bVar0 := core.Cudd_bddIthVar(dd, 30)
	bVar1 := core.Cudd_bddIthVar(dd, 31)
	bCube := core.Cudd_bddAnd(dd, bVar0, bVar1)
	core.Cudd_Ref(bCube)
	for k := 0; k < p.nFuncs; k++ {
		fmt.Printf("FUNCTION %d\n", k)
		for i1 := 0; i1 < p.nFuncs; i1++ {
			for i2 := i1 + 1; i2 < p.nFuncs; i2++ {
				var cofs [16]*core.DdNode
				for v1 := 0; v1 < 4; v1++ {
					var0 := core.Cudd_NotCond(core.Cudd_bddIthVar(dd, 29-2*i1), v1&1 == 0)
					var1 := core.Cudd_NotCond(core.Cudd_bddIthVar(dd, 28-2*i1), v1&2 == 0)
					cube1 := core.Cudd_bddAnd(dd, var0, var1)
					core.Cudd_Ref(cube1)
					for v2 := 0; v2 < 4; v2++ {
						var0 := core.Cudd_NotCond(core.Cudd_bddIthVar(dd, 29-2*i2), v2&1 == 0)
						var1 := core.Cudd_NotCond(core.Cudd_bddIthVar(dd, 28-2*i2), v2&2 == 0)
						cube2 := core.Cudd_bddAnd(dd, var0, var1)
						core.Cudd_Ref(cube2)
						cube12 := core.Cudd_bddAnd(dd, cube1, cube2)
						core.Cudd_Ref(cube12)
						cofs[v1*4+v2] = core.Cudd_Cofactor(dd, p.bFuncs[k], cube12)
						core.Cudd_Ref(cofs[v1*4+v2])
						core.Cudd_RecursiveDeref(dd, cube12)
						core.Cudd_RecursiveDeref(dd, cube2)
					}
					core.Cudd_RecursiveDeref(dd, cube1)
				}
				unique := make(map[*core.DdNode]struct{}, 16)
				for _, cofactor := range cofs {
					unique[cofactor] = struct{}{}
				}
				fmt.Printf("%d ", len(unique))
				for _, cofactor := range cofs {
					core.Cudd_RecursiveDeref(dd, cofactor)
				}
				fmt.Println()
			}
		}
	}
	core.Cudd_RecursiveDeref(dd, bCube)
}
