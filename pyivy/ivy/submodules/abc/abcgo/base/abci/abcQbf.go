package abci

import (
	"fmt"
	"math/rand"
	"time"

	aiggraph "github.com/glycerine/abcgo/aig/aig"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Abc_NtkQbf solves the bounded synthesis problem exists p forall x M(p,x)
// using the counterexample-guided synthesis loop from ABC.
func Abc_NtkQbf(pNtk *Abc_Ntk_t, nPars, nItersMax, fDumpCnf, fVerbose int32) {
	assert((Abc_NtkIsStrash(pNtk) != 0), "Abc_NtkIsStrash(pNtk)")
	assert((Abc_NtkIsComb(pNtk) != 0), "Abc_NtkIsComb(pNtk)")
	assert(Abc_NtkPoNum(pNtk) == 1, "Abc_NtkPoNum(pNtk) == 1")
	assert(nPars > 0 && nPars < Abc_NtkPiNum(pNtk), "nPars > 0 && nPars < Abc_NtkPiNum(pNtk)")
	nInputs := Abc_NtkPiNum(pNtk) - nPars
	start := time.Now()

	if fDumpCnf != 0 {
		man := Abc_NtkToDar(pNtk, 0, 0)
		if man == nil {
			return
		}
		defer aiggraph.Aig_ManStop(man)
		data := cnf.Cnf_Derive(man, 0)
		if data == nil {
			return
		}
		defer cnf.Cnf_DataFree(data)
		varMap := vec.Vec_IntStart(data.Variables)
		for i, input := range man.Inputs() {
			if int32(i) >= nPars {
				break
			}
			id := aiggraph.Aig_ObjId(input)
			if int(id) >= 0 && int(id) < len(data.VarNums) {
				variable := data.VarNums[id]
				if variable >= 0 && int(variable) < vec.Vec_IntSize(varMap) {
					vec.Vec_IntWriteEntry(varMap, int(variable), 1)
				}
			}
		}
		forall, exists := vec.Vec_IntAlloc(int(nPars)), vec.Vec_IntAlloc(int(nInputs))
		defer vec.Vec_IntFree(varMap)
		defer vec.Vec_IntFree(forall)
		defer vec.Vec_IntFree(exists)
		for i := int32(0); i < int32(vec.Vec_IntSize(varMap)); i++ {
			if vec.Vec_IntEntry(varMap, int(i)) != 0 {
				vec.Vec_IntPush(forall, i)
			} else {
				vec.Vec_IntPush(exists, i)
			}
		}
		filename := extra.Extra_FileNameGenericAppend(baseabc.Abc_NtkSpec(pNtk), ".qdimacs")
		cnf.Cnf_DataWriteIntoFile(data, filename, 0, forall, exists)
		fmt.Printf("The 2QBF formula was written into file %q.\n", filename)
		return
	}

	piValues := vec.Vec_IntStart(int(Abc_NtkPiNum(pNtk)))
	defer vec.Vec_IntFree(piValues)
	rng := rand.New(rand.NewSource(time.Now().Unix()))
	for i := nPars; i < Abc_NtkPiNum(pNtk); i++ {
		vec.Vec_IntWriteEntry(piValues, int(i), int32(rng.Intn(2)))
	}
	Abc_NtkVectorClearPars(piValues, nPars)
	synthesis := Abc_NtkMiterCofactor(pNtk, piValues)
	if synthesis == nil {
		return
	}
	if fVerbose != 0 {
		fmt.Printf("Iter %2d : AIG = %6d  ", 0, baseabc.Abc_NtkNodeNum(synthesis))
		Abc_NtkVectorPrintVars(pNtk, piValues, nPars)
		fmt.Println()
	}
	found := int32(0)
	var iteration int32
	for iteration = 0; iteration < nItersMax; iteration++ {
		var started time.Time
		started = time.Now()
		ret := Abc_NtkDSat(synthesis, 0, 0, 0, 0, 0, 1, 0, 0, 0)
		if ret == 0 {
			Abc_NtkModelToVector(synthesis, piValues)
		}
		if ret == 1 {
			break
		}
		if ret == -1 {
			fmt.Println("Synthesis timed out.")
			break
		}
		synthesisTime := time.Since(started)

		Abc_NtkVectorClearVars(pNtk, piValues, nPars)
		verification := Abc_NtkMiterCofactor(pNtk, piValues)
		if verification == nil {
			break
		}
		baseabc.Abc_ObjXorFaninC(Abc_NtkPo(verification, 0), 0)
		ret = Abc_NtkMiterSat(verification, 0, 0, 0, nil, nil)
		if ret == 0 {
			Abc_NtkModelToVector(verification, piValues)
		}
		Abc_NtkDelete(verification)
		if ret == 1 {
			found = 1
			break
		}
		if ret == -1 {
			fmt.Println("Verification timed out.")
			break
		}

		Abc_NtkVectorClearPars(piValues, nPars)
		synthesisNext := Abc_NtkMiterCofactor(pNtk, piValues)
		if synthesisNext == nil {
			break
		}
		previous := synthesis
		synthesis = Abc_NtkMiterAnd(previous, synthesisNext, 0, 0)
		Abc_NtkDelete(synthesisNext)
		Abc_NtkDelete(previous)
		if synthesis == nil {
			break
		}
		if fVerbose != 0 {
			fmt.Printf("Iter %2d : AIG = %6d  ", iteration+1, baseabc.Abc_NtkNodeNum(synthesis))
			Abc_NtkVectorPrintVars(pNtk, piValues, nPars)
			fmt.Printf("  Syn = %9.2f sec\n", synthesisTime.Seconds())
		}
	}
	Abc_NtkDelete(synthesis)
	if found != 0 {
		zeros := vec.Vec_IntCountZero(piValues)
		fmt.Printf("Parameters: ")
		Abc_NtkVectorPrintPars(piValues, nPars)
		fmt.Printf("  Statistics: 0=%d 1=%d\n", zeros, vec.Vec_IntSize(piValues)-zeros)
		fmt.Printf("Solved after %d iterations.  ", iteration)
	} else if iteration == nItersMax {
		fmt.Printf("Quit after %d iterations.  ", nItersMax)
	} else {
		fmt.Printf("Implementation does not exist.  ")
	}
	fmt.Printf("Total runtime = %9.2f sec\n", time.Since(start).Seconds())
}

func Abc_NtkModelToVector(pNtk *Abc_Ntk_t, vPiValues *Vec_Int_t) {
	for i := int32(0); i < Abc_NtkPiNum(pNtk); i++ {
		Vec_IntWriteEntry(vPiValues, i, pNtk.Model[i])
	}
}

func Abc_NtkVectorClearPars(vPiValues *Vec_Int_t, nPars int32) {
	for i := int32(0); i < nPars; i++ {
		Vec_IntWriteEntry(vPiValues, i, -1)
	}
}

func Abc_NtkVectorClearVars(pNtk *Abc_Ntk_t, vPiValues *Vec_Int_t, nPars int32) {
	for i := nPars; i < Abc_NtkPiNum(pNtk); i++ {
		Vec_IntWriteEntry(vPiValues, i, -1)
	}
}

func Abc_NtkVectorPrintPars(vPiValues *Vec_Int_t, nPars int32) {
	for i := int32(0); i < nPars; i++ {
		fmt.Printf("%d", Vec_IntEntry(vPiValues, i))
	}
}

func Abc_NtkVectorPrintVars(pNtk *Abc_Ntk_t, vPiValues *Vec_Int_t, nPars int32) {
	for i := nPars; i < Abc_NtkPiNum(pNtk); i++ {
		fmt.Printf("%d", Vec_IntEntry(vPiValues, i))
	}
}
