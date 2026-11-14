package abci

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
)

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// Abc_Ntk_t * Abc_NtkDarFold2( Abc_Ntk_t * pNtk, int fCompl, int fVerbose , int);

func Abc_NtkDarUnfold2(pNtk *Abc_Ntk_t, nFrames int32, nConfs int32, nProps int32, fStruct int32, fOldAlgo int32, fVerbose int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan *Aig_Man_t
	var pTemp *Aig_Man_t = nil
	var typeII_cnt int32 = 0
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	if fStruct != 0 {
		panic("assertion failed: 0") // The C implementation asserts here.
	}
	pTemp = pMan
	pMan = saig.Saig_ManDupUnfoldConstrsFunc2(pTemp, nFrames, nConfs, nProps, fOldAlgo, fVerbose, &typeII_cnt)
	aig.Aig_ManStop(pTemp)
	if pMan == nil {
		return nil
	}
	//    typeII_cnt = pMan->nConstrsTypeII;
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	pNtkAig.Name = extra.Extra_UtilStrsav(pMan.Name)
	pNtkAig.Spec = extra.Extra_UtilStrsav(pMan.Spec)
	aig.Aig_ManStop(pMan)

	return pNtkAig //Abc_NtkDarFold2(pNtkAig, 0, fVerbose, typeII_cnt);

	//return pNtkAig;
}

/*
*Function*************************************************************

	Synopsis    [Performs BDD-based reachability analysis.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkDarFold2(pNtk *Abc_Ntk_t, fCompl int32, fVerbose int32, typeII_cnt int32) *Abc_Ntk_t {
	var pNtkAig *Abc_Ntk_t
	var pMan, pTemp *Aig_Man_t
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	pMan = Abc_NtkToDar(pNtk, 0, 1)
	if pMan == nil {
		return nil
	}
	pTemp = pMan
	pMan = saig.Saig_ManDupFoldConstrsFunc2(pTemp, fCompl, fVerbose, typeII_cnt)
	aig.Aig_ManStop(pTemp)
	pNtkAig = Abc_NtkFromAigPhase(pMan)
	pNtkAig.Name = extra.Extra_UtilStrsav(pMan.Name)
	pNtkAig.Spec = extra.Extra_UtilStrsav(pMan.Spec)
	aig.Aig_ManStop(pMan)
	return pNtkAig
}
