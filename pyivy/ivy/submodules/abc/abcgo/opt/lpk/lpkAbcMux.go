package lpk

import (
	kit "github.com/glycerine/abcgo/bool/kit"
	util "github.com/glycerine/abcgo/misc/util"
)

/**CFile****************************************************************

  FileName    [lpkAbcMux.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    [LUT-decomposition based on recursive MUX decomposition.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkAbcMux.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "lpkInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Checks the possibility of MUX decomposition.]

	Description [Returns the best variable to use for MUX decomposition.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Lpk_MuxAnalize_Res Lpk_Res_t
var spStatic_Lpk_MuxAnalize_pRes *Lpk_Res_t = &spStatic_Lpk_MuxAnalize_Res

func Lpk_MuxAnalize(pMan *Lpk_Man_t, p *Lpk_Fun_t) *Lpk_Res_t {

	var nSuppSize0, nSuppSize1, nSuppSizeS, nSuppSizeL int32
	var Var, Area, Polarity, Delay, Delay0, Delay1, DelayA, DelayB int32
	*spStatic_Lpk_MuxAnalize_pRes = Lpk_Res_t{}
	assert(p.uSupp == kit.Kit_BitMask(int32(p.nVars)), "p.uSupp == kit.Kit_BitMask(int32(p.nVars))")
	assert(p.fSupports != 0, "p.fSupports")
	// derive the delay and area after MUX-decomp with each var - and find the best var
	spStatic_Lpk_MuxAnalize_pRes.Variable = -1
	// inlining of macro Lpk_SuppForEachVar
	for Var = 0; Var < 16; Var++ {
		if (p.uSupp & (1 << Var)) == 0 {
		} else {
			nSuppSize0 = kit.Kit_WordCountOnes(p.puSupps[2*Var+0])
			nSuppSize1 = kit.Kit_WordCountOnes(p.puSupps[2*Var+1])
			assert(nSuppSize0 < int32(p.nVars), "nSuppSize0 < int32(p.nVars)")
			assert(nSuppSize1 < int32(p.nVars), "nSuppSize1 < int32(p.nVars)")
			if nSuppSize0 < 1 || nSuppSize1 < 1 {
				continue
			}
			//printf( "%d %d    ", nSuppSize0, nSuppSize1 );
			if nSuppSize0 <= int32(p.nLutK)-2 && nSuppSize1 <= int32(p.nLutK)-2 {
				// include cof var into 0-block
				DelayA = Lpk_SuppDelay(p.puSupps[2*Var+0]|(1<<Var), p.pDelays[:])
				DelayB = Lpk_SuppDelay(p.puSupps[2*Var+1], p.pDelays[:])
				Delay0 = util.Abc_MaxInt(DelayA, DelayB+1)
				// include cof var into 1-block
				DelayA = Lpk_SuppDelay(p.puSupps[2*Var+1]|(1<<Var), p.pDelays[:])
				DelayB = Lpk_SuppDelay(p.puSupps[2*Var+0], p.pDelays[:])
				Delay1 = util.Abc_MaxInt(DelayA, DelayB+1)
				// get the best delay
				Delay = util.Abc_MinInt(Delay0, Delay1)
				Area = 2
				Polarity = 0
				if Delay == Delay1 {
					Polarity = 1
				}

			} else if nSuppSize0 <= int32(p.nLutK)-2 {
				DelayA = Lpk_SuppDelay(p.puSupps[2*Var+0]|(1<<Var), p.pDelays[:])
				DelayB = Lpk_SuppDelay(p.puSupps[2*Var+1], p.pDelays[:])
				Delay = util.Abc_MaxInt(DelayA, DelayB+1)
				Area = 1 + Lpk_LutNumLuts(nSuppSize1, int32(p.nLutK))
				Polarity = 0

			} else if nSuppSize1 <= int32(p.nLutK)-2 {
				DelayA = Lpk_SuppDelay(p.puSupps[2*Var+1]|(1<<Var), p.pDelays[:])
				DelayB = Lpk_SuppDelay(p.puSupps[2*Var+0], p.pDelays[:])
				Delay = util.Abc_MaxInt(DelayA, DelayB+1)
				Area = 1 + Lpk_LutNumLuts(nSuppSize0, int32(p.nLutK))
				Polarity = 1

			} else if nSuppSize0 <= int32(p.nLutK) {
				DelayA = Lpk_SuppDelay(p.puSupps[2*Var+1]|(1<<Var), p.pDelays[:])
				DelayB = Lpk_SuppDelay(p.puSupps[2*Var+0], p.pDelays[:])
				Delay = util.Abc_MaxInt(DelayA, DelayB+1)
				Area = 1 + Lpk_LutNumLuts(nSuppSize1+2, int32(p.nLutK))
				Polarity = 1

			} else if nSuppSize1 <= int32(p.nLutK) {
				DelayA = Lpk_SuppDelay(p.puSupps[2*Var+0]|(1<<Var), p.pDelays[:])
				DelayB = Lpk_SuppDelay(p.puSupps[2*Var+1], p.pDelays[:])
				Delay = util.Abc_MaxInt(DelayA, DelayB+1)
				Area = 1 + Lpk_LutNumLuts(nSuppSize0+2, int32(p.nLutK))
				Polarity = 0

			} else {
				// include cof var into 0-block
				DelayA = Lpk_SuppDelay(p.puSupps[2*Var+0]|(1<<Var), p.pDelays[:])
				DelayB = Lpk_SuppDelay(p.puSupps[2*Var+1], p.pDelays[:])
				Delay0 = util.Abc_MaxInt(DelayA, DelayB+1)
				// include cof var into 1-block
				DelayA = Lpk_SuppDelay(p.puSupps[2*Var+1]|(1<<Var), p.pDelays[:])
				DelayB = Lpk_SuppDelay(p.puSupps[2*Var+0], p.pDelays[:])
				Delay1 = util.Abc_MaxInt(DelayA, DelayB+1)
				// get the best delay
				Delay = util.Abc_MinInt(Delay0, Delay1)
				if Delay == Delay0 {
					Area = Lpk_LutNumLuts(nSuppSize0+2, int32(p.nLutK)) + Lpk_LutNumLuts(nSuppSize1, int32(p.nLutK))
				} else {
					Area = Lpk_LutNumLuts(nSuppSize1+2, int32(p.nLutK)) + Lpk_LutNumLuts(nSuppSize0, int32(p.nLutK))
				}
				Polarity = 0
				if Delay == Delay1 {
					Polarity = 1
				}
			}
			// find the best variable
			if Delay > int32(p.nDelayLim) {
				continue
			}
			if Area > int32(p.nAreaLim) {
				continue
			}
			nSuppSizeS = util.Abc_MinInt(nSuppSize0+2*(1-Polarity), nSuppSize1+2*Polarity)
			nSuppSizeL = util.Abc_MaxInt(nSuppSize0+2*(1-Polarity), nSuppSize1+2*Polarity)
			if nSuppSizeL > int32(p.nVars) {
				continue
			}
			if spStatic_Lpk_MuxAnalize_pRes.Variable == -1 || spStatic_Lpk_MuxAnalize_pRes.AreaEst > Area ||
				(spStatic_Lpk_MuxAnalize_pRes.AreaEst == Area && spStatic_Lpk_MuxAnalize_pRes.nSuppSizeS+spStatic_Lpk_MuxAnalize_pRes.nSuppSizeL > nSuppSizeS+nSuppSizeL) ||
				(spStatic_Lpk_MuxAnalize_pRes.AreaEst == Area && spStatic_Lpk_MuxAnalize_pRes.nSuppSizeS+spStatic_Lpk_MuxAnalize_pRes.nSuppSizeL == nSuppSizeS+nSuppSizeL && spStatic_Lpk_MuxAnalize_pRes.DelayEst > Delay) {
				spStatic_Lpk_MuxAnalize_pRes.Variable = Var
				spStatic_Lpk_MuxAnalize_pRes.Polarity = Polarity
				spStatic_Lpk_MuxAnalize_pRes.AreaEst = Area
				spStatic_Lpk_MuxAnalize_pRes.DelayEst = Delay
				spStatic_Lpk_MuxAnalize_pRes.nSuppSizeS = nSuppSizeS
				spStatic_Lpk_MuxAnalize_pRes.nSuppSizeL = nSuppSizeL
			}
		}
	}
	if spStatic_Lpk_MuxAnalize_pRes.Variable == -1 {
		return nil
	} else {
		return spStatic_Lpk_MuxAnalize_pRes
	}
}

/*
*Function*************************************************************

	Synopsis    [Transforms the function decomposed by the MUX decomposition.]

	Description [Returns the best variable to use for MUX decomposition.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MuxSplit(pMan *Lpk_Man_t, p *Lpk_Fun_t, Var int32, Pol int32) *Lpk_Fun_t {
	var pNew *Lpk_Fun_t
	var pTruth []uint32 = Lpk_FunTruth(p, 0)
	var pTruth0 []uint32 = Lpk_FunTruth(p, 1)
	var pTruth1 []uint32 = Lpk_FunTruth(p, 2)
	//    unsigned uSupp;
	var iVarVac int32
	assert(Var >= 0 && Var < int32(p.nVars), "Var >= 0 && Var < int32(p.nVars)")
	assert(p.nAreaLim >= 2, "p.nAreaLim >= 2")
	assert(p.uSupp == kit.Kit_BitMask(int32(p.nVars)), "p.uSupp == kit.Kit_BitMask(int32(p.nVars))")
	kit.Kit_TruthCofactor0New(pTruth0, pTruth, int32(p.nVars), Var)
	kit.Kit_TruthCofactor1New(pTruth1, pTruth, int32(p.nVars), Var)
	/*
	   uSupp = kit.Kit_TruthSupport( pTruth, p->nVars );
	   Extra_PrintBinary( stdout, &uSupp, 16 ); printf( "\n" );
	   uSupp = kit.Kit_TruthSupport( pTruth0, p->nVars );
	   Extra_PrintBinary( stdout, &uSupp, 16 ); printf( "\n" );
	   uSupp = kit.Kit_TruthSupport( pTruth1, p->nVars );
	   Extra_PrintBinary( stdout, &uSupp, 16 ); printf( "\n\n" );
	*/
	// derive the new component
	if Pol != 0 {
		pNew = Lpk_FunDup(p, pTruth0)
	} else {
		pNew = Lpk_FunDup(p, pTruth1)
	}
	// update the support of the old component
	if Pol != 0 {
		p.uSupp = kit.Kit_TruthSupport(pTruth1, int32(p.nVars))
	} else {
		p.uSupp = kit.Kit_TruthSupport(pTruth0, int32(p.nVars))
	}
	p.uSupp |= (1 << Var)
	// update the truth table of the old component
	iVarVac = kit.Kit_WordFindFirstBit(^p.uSupp)
	assert(iVarVac < int32(p.nVars), "iVarVac < int32(p.nVars)")
	p.uSupp |= (1 << iVarVac)
	kit.Kit_TruthIthVar(pTruth, int32(p.nVars), iVarVac)
	if Pol != 0 {
		kit.Kit_TruthMuxVar(pTruth, pTruth, pTruth1, int32(p.nVars), Var)
	} else {
		kit.Kit_TruthMuxVar(pTruth, pTruth0, pTruth, int32(p.nVars), Var)
	}
	assert(p.uSupp == kit.Kit_TruthSupport(pTruth, int32(p.nVars)), "p.uSupp == kit.Kit_TruthSupport(pTruth, int32(p.nVars))")
	// set the decomposed variable
	p.pFanins[iVarVac] = byte(pNew.Id)
	p.pDelays[iVarVac] = int32(p.nDelayLim) - 1
	// support minimize both
	p.fSupports = 0
	Lpk_FunSuppMinimize(p)
	Lpk_FunSuppMinimize(pNew)
	// update delay and area requirements
	pNew.nDelayLim = p.nDelayLim - 1
	if pNew.nVars <= pNew.nLutK {
		pNew.nAreaLim = 1
		p.nAreaLim = p.nAreaLim - 1

	} else if p.nVars <= p.nLutK {
		pNew.nAreaLim = p.nAreaLim - 1
		p.nAreaLim = 1

	} else if p.nVars < pNew.nVars {
		pNew.nAreaLim = p.nAreaLim/2 + p.nAreaLim%2
		p.nAreaLim = p.nAreaLim/2 - p.nAreaLim%2

	} else { // if ( pNew->nVars < p->nVars )

		pNew.nAreaLim = p.nAreaLim/2 - p.nAreaLim%2
		p.nAreaLim = p.nAreaLim/2 + p.nAreaLim%2
	}
	pNew.fMark = 1
	return pNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
