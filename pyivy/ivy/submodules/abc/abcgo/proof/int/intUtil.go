package int

// Line-by-line port of src/proof/int/intUtil.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the property fails in the initial state.]

**********************************************************************
*/
func Inter_ManCheckInitialState(p *Aig_Man_t) int32 {
	var pCnf *Cnf_Dat_t
	var pObj Aig_Obj_t
	var pSat *sat_solver
	var i, status int32
	//abctime clk = Abc_Clock();
	pCnf = cnf.Cnf_Derive(p, saig.Saig_ManRegNum(p))
	pSat = cnf.Cnf_DataWriteIntoSolver(pCnf, 1, 1)
	if pSat == nil {
		cnf.Cnf_DataFree(pCnf)
		return 0
	}
	status = pSat.SolveC(nil, 0, 0, 0, 0)
	//ABC_PRT( "Time", Abc_Clock() - clk );
	if status == int32(bsat.Satisfiable) {
		seqModel := util.Abc_CexAlloc(aig.Aig_ManRegNum(p), saig.Saig_ManPiNum(p), 1)
		for i = 0; i < saig.Saig_ManPiNum(p); i++ {
			pObj = aig.Aig_ManCi(p, i)
			if pSat.VarValueC(pCnf.VarNums[aig.Aig_ObjId(pObj)]) != 0 {
				util.Abc_InfoSetBit(seqModel.PData, aig.Aig_ManRegNum(p)+i)
			}
		}
		p.SetCecCounterexample(seqModel)
	}
	cnf.Cnf_DataFree(pCnf)
	return b2i(status == int32(bsat.Satisfiable))
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the property holds in all states.]

**********************************************************************
*/
func Inter_ManCheckAllStates(p *Aig_Man_t) int32 {
	var pCnf *Cnf_Dat_t
	var pSat *sat_solver
	var status int32
	clk := util.Abc_Clock()
	pCnf = cnf.Cnf_Derive(p, saig.Saig_ManRegNum(p))
	pSat = cnf.Cnf_DataWriteIntoSolver(pCnf, 1, 0)
	cnf.Cnf_DataFree(pCnf)
	if pSat == nil {
		return 1
	}
	status = pSat.SolveC(nil, 0, 0, 0, 0)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	return b2i(status == int32(bsat.Unsatisfiable))
}

func b2i(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
