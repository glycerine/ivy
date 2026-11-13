package int

// Interpolation manager state from proof/int/intInt.h. The graph, CNF, and
// vector types use the corresponding Go-owned representations.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Cnf_Dat_t = cnf.Cnf_Dat_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type sat_solver = bsat.Solver
type abctime = int64

type Inter_Man_t struct {
	pAig, pAigTrans                                                   *Aig_Man_t
	pCnfAig                                                           *Cnf_Dat_t
	pInter                                                            *Aig_Man_t
	pCnfInter                                                         *Cnf_Dat_t
	pFrames                                                           *Aig_Man_t
	pCnfFrames                                                        *Cnf_Dat_t
	vVarsAB                                                           *Vec_Int_t
	pInterNew                                                         *Aig_Man_t
	vInters                                                           *Vec_Ptr_t
	nFrames, nConfCur, nConfLimit, fVerbose                           int32
	pFileName                                                         string
	timeRwr, timeCnf, timeSat, timeInt, timeEqu, timeOther, timeTotal abctime
}

