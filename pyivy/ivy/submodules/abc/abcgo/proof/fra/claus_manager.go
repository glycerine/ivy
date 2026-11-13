package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Clu_Man_t holds the clause generation and induction state from fraClaus.c.
type Clu_Man_t struct {
	nFrames, nPref, nClausesMax, nLutSize, nLevels, nCutsMax, nBatches int32
	fStepUp, fTarget, fVerbose, fVeryVerbose                           int32
	nSimWords, nSimWordsPref, nSimFrames, nBTLimit                     int32
	pAig                                                               *aig.Graph
	pSatMain, pSatBmc                                                  *bsat.Solver
	pCnf                                                               *cnf.Cnf_Dat_t
	fFail, fFiltering, fNothingNew                                     int32
	vLits, vClauses, vCosts                                            *vec.Vec_Int_t
	nClauses, nCuts, nOneHots, nOneHotsProven                          int32
	vLitsProven, vClausesProven                                        *vec.Vec_Int_t
	models                                                             *clausModels
}
