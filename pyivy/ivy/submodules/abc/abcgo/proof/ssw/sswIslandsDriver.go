package ssw

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ssw_SecWithSimilaritySweep(p0, p1 *aig.Graph, pairs *vec.Vec_Int_t, pars *Ssw_Pars_t) *aig.Graph {
	if pars == nil {
		pars = new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(pars)
	}
	originalDeadline := pars.deadline
	if pars.TimeLimit > 0 {
		limit := time.Now().Add(time.Duration(pars.TimeLimit) * time.Second)
		if pars.deadline.IsZero() || limit.Before(pars.deadline) {
			pars.deadline = limit
		}
	}
	defer func() { pars.deadline = originalDeadline }()
	if sswDeadlineExpired(pars) {
		return nil
	}
	Ssw_MatchingStart(p0, p1, pairs)
	if pars.nIsleDist != 0 {
		Ssw_MatchingExtend(p0, p1, pars.nIsleDist, pars.fVerbose)
	}
	Ssw_MatchingComplete(p0, p1)
	Ssw_MatchingComplete(p1, p0)
	all := Ssw_MatchingPairs(p0, p1)
	miter := sswIslandMiter(p0, p1)
	matches := Ssw_MatchingMiter(miter, p0, p1, all)
	man := Ssw_ManCreate(miter, pars)
	if pars.fPartSigCorr != 0 {
		man.ppClasses = Ssw_ClassesPreparePairsSimple(miter, matches)
	} else {
		man.ppClasses = Ssw_ClassesPrepare(miter, pars.nFramesK, pars.fLatchCorr, pars.fConstCorr, pars.fOutputCorr, pars.nMaxLevs, pars.fVerbose)
	}
	if pars.fDumpSRInit != 0 {
		if pars.fPartSigCorr != 0 {
			reduced := Ssw_SpeculativeReduction(man)
			if err := aig.Aig_ManDumpBlif(reduced, "srm_part.blif", nil, nil); err != nil {
				panic(err)
			}
			fmt.Println("Speculatively reduced miter is saved in file \"srm_part.blif\".")
		} else {
			fmt.Println("Dumping speculative miter is possible only for partial signal correspondence (switch \"-c\").")
		}
	}
	man.pSml = Ssw_SmlStart(miter, 0, 1+pars.nFramesAddSim, 1)
	Ssw_ClassesSetData(man.ppClasses, man.pSml, func(data any, e aig.Edge) uint32 { return Ssw_SmlObjHashWord(data.(*Ssw_Sml_t), e) }, func(data any, e aig.Edge) int32 { return Ssw_SmlObjIsConstWord(data.(*Ssw_Sml_t), e) }, func(data any, a, b aig.Edge) int32 { return Ssw_SmlObjsAreEqualWord(data.(*Ssw_Sml_t), a, b) })
	result := Ssw_SignalCorrespondenceRefine(man)
	Ssw_ManStop(man)
	return result
}

func Ssw_SecWithSimilarityPairs(p0, p1 *aig.Graph, pairs *vec.Vec_Int_t, pars *Ssw_Pars_t) int32 {
	start := time.Now()
	if pars == nil {
		pars = new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(pars)
	}
	result := Ssw_SecWithSimilaritySweep(p0, p1, pairs, pars)
	if result == nil {
		return -1
	}
	defer aig.Aig_ManStop(result)
	status := Ssw_MiterStatus(result, 1)
	switch status {
	case 1:
		fmt.Print("Verification successful.  ")
	case 0:
		fmt.Print("Verification failed with a counter-example.  ")
	default:
		fmt.Printf("Verification UNDECIDED. The number of remaining regs = %d (total = %d).  ", aig.Aig_ManRegNum(result), aig.Aig_ManRegNum(p0)+aig.Aig_ManRegNum(p1))
	}
	fmt.Printf("Time = %9.2f sec\n", time.Since(start).Seconds())
	return status
}

func Ssw_SecWithSimilarity(p0, p1 *aig.Graph, pars *Ssw_Pars_t) int32 {
	if pars == nil {
		pars = new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(pars)
	}
	if pars.fVerbose != 0 {
		fmt.Println("Performing sequential verification using structural similarity.")
	}
	var a, b *aig.Graph
	if p1 == nil {
		if saig.Saig_ManDemiterSimpleDiff(p0, &a, &b) == 0 {
			fmt.Println("Demitering has failed.")
			return -1
		}
	} else {
		a, b = aig.Aig_ManDupSimple(p0), aig.Aig_ManDupSimple(p1)
	}
	if aig.Aig_ManRegNum(a) <= 0 || aig.Aig_ManRegNum(b) <= 0 || saig.Saig_ManPiNum(a) != saig.Saig_ManPiNum(b) || saig.Saig_ManPoNum(a) != saig.Saig_ManPoNum(b) {
		panic("ssw: incompatible sequential graphs")
	}
	pairs := saig.Saig_StrSimPerformMatching(a, b, 0, pars.fVerbose, nil)
	return Ssw_SecWithSimilarityPairs(a, b, pairs, pars)
}
