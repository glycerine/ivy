package abci

import base_abc "github.com/glycerine/abcgo/base/abc"

import (
	"fmt"
	"math"
	"os"

	"github.com/glycerine/abcgo/proof/fraig"
)

// Prove_Params_t is the command-facing form of ABC's proving options. The
// original C fields are package local in proof/fraig, so the integration layer
// keeps the source names here and translates only the parameters it consumes.
type Prove_Params_t struct {
	fUseFraiging, fUseRewriting, fUseBdds, fVerbose int32
	nItersMax                                       int32
	nMiteringLimitStart                             int32
	nMiteringLimitMulti                             float32
	nRewritingLimitStart                            int32
	nRewritingLimitMulti                            float32
	nFraigingLimitStart                             int32
	nFraigingLimitMulti                             float32
	nBddSizeLimit                                   int32
	fBddReorder                                     int32
	nMiteringLimitLast                              int32
	nTotalBacktrackLimit, nTotalInspectLimit        int64
	nTotalBacktracksMade, nTotalInspectsMade        int64
}

// Prove_ParamsSetDefault matches proof/fraig/fraigMan.c field-for-field.
func Prove_ParamsSetDefault(p *Prove_Params_t) {
	*p = Prove_Params_t{fUseFraiging: 1, fUseRewriting: 1, nItersMax: 6,
		nMiteringLimitStart: 5000, nMiteringLimitMulti: 2,
		nRewritingLimitStart: 3, nRewritingLimitMulti: 1,
		nFraigingLimitStart: 2, nFraigingLimitMulti: 8,
		nBddSizeLimit: 1000000, fBddReorder: 1}
}

// Abc_NtkMiterProve attempts to prove a one-output strashed miter. It returns
// -1 when resource limits prevent a result, 0 for SAT, and 1 for UNSAT.
func Abc_NtkMiterProve(ppNtk **Abc_Ntk_t, pPars any) int32 {
	p := pPars.(*Prove_Params_t)
	n := *ppNtk
	assert(Abc_NtkIsStrash(n) != 0, "Abc_NtkIsStrash(pNtk)")
	assert(Abc_NtkPoNum(n) == 1, "Abc_NtkPoNum(pNtk) == 1")
	ret := int32(-1)
	var satConfs, satInspects, inspectLimit int64
	if p.fVerbose != 0 {
		fmt.Printf("RESOURCE LIMITS: Iterations = %d. Rewriting = %s. Fraiging = %s.\n", p.nItersMax, abcYesNo(p.fUseRewriting), abcYesNo(p.fUseFraiging))
		fmt.Printf("Miter = %d (%3.1f).  Rwr = %d (%3.1f).  Fraig = %d (%3.1f).  Last = %d.\n", p.nMiteringLimitStart, p.nMiteringLimitMulti, p.nRewritingLimitStart, p.nRewritingLimitMulti, p.nFraigingLimitStart, p.nFraigingLimitMulti, p.nMiteringLimitLast)
	}
	if p.fUseRewriting == 0 && p.fUseFraiging == 0 {
		clk := Abc_Clock()
		ret = Abc_NtkMiterSat(n, int64(p.nMiteringLimitLast), 0, 0, nil, nil)
		Abc_NtkMiterPrint(n, "SAT solving", clk, p.fVerbose)
		*ppNtk = n
		return ret
	}
	for iter := int32(0); iter < p.nItersMax; iter++ {
		if p.fVerbose != 0 {
			fmt.Printf("ITERATION %2d : Confs = %6d. FraigBTL = %3d. \n", iter+1, int32(float64(p.nMiteringLimitStart)*math.Pow(float64(p.nMiteringLimitMulti), float64(iter))), int32(float64(p.nFraigingLimitStart)*math.Pow(float64(p.nFraigingLimitMulti), float64(iter))))
			_, _ = fmt.Fprintln(os.Stdout)
		}
		clk := Abc_Clock()
		inspectLimit = 0
		if p.nTotalInspectLimit != 0 {
			inspectLimit = p.nTotalInspectLimit - p.nTotalInspectsMade
		}
		ret = Abc_NtkMiterSat(n, int64(float64(p.nMiteringLimitStart)*math.Pow(float64(p.nMiteringLimitMulti), float64(iter))), inspectLimit, 0, &satConfs, &satInspects)
		Abc_NtkMiterPrint(n, "SAT solving", clk, p.fVerbose)
		if ret >= 0 {
			break
		}
		p.nTotalBacktracksMade += satConfs
		p.nTotalInspectsMade += satInspects
		if resourceLimitReached(p) {
			fmt.Println("Reached global limit on conflicts/inspects. Quitting.")
			*ppNtk = n
			return -1
		}
		if p.fUseRewriting != 0 {
			clk = Abc_Clock()
			counter := int32(float64(p.nRewritingLimitStart) * math.Pow(float64(p.nRewritingLimitMulti), float64(iter)))
			for {
				Abc_NtkRewrite(n, 0, 0, 0, 0, 0)
				if ret = Abc_NtkMiterIsConstant(n); ret >= 0 {
					break
				}
				counter--
				if counter == 0 {
					break
				}
				Abc_NtkRefactor(n, 10, 16, 0, 0, 0, 0)
				if ret = Abc_NtkMiterIsConstant(n); ret >= 0 {
					break
				}
				counter--
				if counter == 0 {
					break
				}
				tmp := n
				n = Abc_NtkBalance(tmp, 0, 0, 0)
				Abc_NtkDelete(tmp)
				if ret = Abc_NtkMiterIsConstant(n); ret >= 0 {
					break
				}
				counter--
				if counter == 0 {
					break
				}
			}
			Abc_NtkMiterPrint(n, "Rewriting  ", clk, p.fVerbose)
		}
		if p.fUseFraiging != 0 {
			clk = Abc_Clock()
			inspectLimit = 0
			if p.nTotalInspectLimit != 0 {
				inspectLimit = p.nTotalInspectLimit - p.nTotalInspectsMade
			}
			tmp := n
			n = Abc_NtkMiterFraig(tmp, int32(float64(p.nFraigingLimitStart)*math.Pow(float64(p.nFraigingLimitMulti), float64(iter))), inspectLimit, &ret, new(int32), &satConfs, &satInspects)
			Abc_NtkDelete(tmp)
			Abc_NtkMiterPrint(n, "FRAIGing   ", clk, p.fVerbose)
			if ret >= 0 {
				break
			}
			p.nTotalBacktracksMade += satConfs
			p.nTotalInspectsMade += satInspects
			if resourceLimitReached(p) {
				fmt.Println("Reached global limit on conflicts/inspects. Quitting.")
				*ppNtk = n
				return -1
			}
		}
	}
	if ret < 0 && p.fUseBdds != 0 && proveBddEnabled {
		if p.fVerbose != 0 {
			fmt.Printf("Attempting BDDs with node limit %d ...\n", p.nBddSizeLimit)
		}
		clk := Abc_Clock()
		if collapsed, result, ok := proveBddFallback(n, p); ok {
			Abc_NtkDelete(n)
			n, ret = collapsed, result
		}
		Abc_NtkMiterPrint(n, "BDD building", clk, p.fVerbose)
	}
	if ret < 0 {
		if p.fVerbose != 0 {
			fmt.Printf("Attempting SAT with conflict limit %d ...\n", p.nMiteringLimitLast)
		}
		clk := Abc_Clock()
		inspectLimit = 0
		if p.nTotalInspectLimit != 0 {
			inspectLimit = p.nTotalInspectLimit - p.nTotalInspectsMade
		}
		ret = Abc_NtkMiterSat(n, int64(p.nMiteringLimitLast), inspectLimit, 0, nil, nil)
		Abc_NtkMiterPrint(n, "SAT solving", clk, p.fVerbose)
	}
	if ret == 0 && Abc_NtkReadModel(n) == nil {
		Abc_NtkSetModel(n, make([]int32, Abc_NtkCiNum(n)))
	}
	*ppNtk = n
	return ret
}

func resourceLimitReached(p *Prove_Params_t) bool {
	return (p.nTotalBacktrackLimit != 0 && p.nTotalBacktracksMade >= p.nTotalBacktrackLimit) || (p.nTotalInspectLimit != 0 && p.nTotalInspectsMade >= p.nTotalInspectLimit)
}

func Abc_NtkMiterFraig(n *Abc_Ntk_t, btLimit int32, inspectLimit int64, ret, fails *int32, confs, inspects *int64) *Abc_Ntk_t {
	words := int32(32)
	denom := Abc_NtkNodeNum(n) + Abc_NtkCiNum(n)
	if denom > 0 && (int64(1)<<27)/int64(denom) < int64(words) {
		words = int32((int64(1) << 27) / int64(denom))
	}
	params := fraig.Fraig_Params_t{}
	fraig.Fraig_ParamsSetDefault(&params)
	params.NPatsRand, params.NPatsDyna = words*32, words*32
	params.NBTLimit, params.NSeconds, params.FTryProve, params.FDoSparse = btLimit, -1, 0, 1
	params.FVerbose, params.NInspLimit = 0, inspectLimit
	man := fraig.Abc_NtkToFraig(n, &params, 0, 0).(*fraig.Fraig_Man_t)
	fraig.Fraig_ManProveMiter(man)
	*ret = fraig.Fraig_ManCheckMiter(man)
	newNtk := Abc_NtkFromFraig(man, n)
	if *ret == 0 {
		model := fraig.Fraig_ManReadModel(man)
		Abc_NtkSetModel(newNtk, append([]int32(nil), model...))
	}
	*fails = fraig.Fraig_ManReadSatFails(man)
	*confs = int64(fraig.Fraig_ManReadConflicts(man))
	*inspects = int64(fraig.Fraig_ManReadInspects(man))
	fraig.Fraig_ManFree(man)
	return newNtk
}

func Abc_NtkMiterPrint(n *Abc_Ntk_t, label string, clk int64, verbose int32) {
	if verbose == 0 {
		return
	}
	level := base_abc.Abc_NtkLevel(n)
	if Abc_NtkIsStrash(n) != 0 {
		level = Abc_AigLevel(n)
	}
	fmt.Printf("Nodes = %7d.  Levels = %4d.  ", Abc_NtkNodeNum(n), level)
	ABC_PRT(label, Abc_Clock()-clk)
}

func Abc_NtkMiterRwsat(n *Abc_Ntk_t) *Abc_Ntk_t {
	Abc_NtkRewrite(n, 0, 0, 0, 0, 0)
	tmp := n
	n = Abc_NtkBalance(tmp, 0, 0, 0)
	Abc_NtkDelete(tmp)
	Abc_NtkRewrite(n, 0, 0, 0, 0, 0)
	Abc_NtkRefactor(n, 10, 16, 0, 0, 0, 0)
	return n
}
