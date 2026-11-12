package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcCexCare.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Computing care set of the counter-example.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcCexCare.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Takes CEX and its care-set. Returns care-set of all objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexCareExtendToObjects(p *Gia_Man_t, pCex *Abc_Cex_t, pCexCare *Abc_Cex_t) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var pObj *Gia_Obj_t;
    var i, k int32;
    assert(pCex.nPis == pCexCare.nPis, "pCex.nPis == pCexCare.nPis");
    assert(pCex.nRegs == pCexCare.nRegs, "pCex.nRegs == pCexCare.nRegs");
    assert(pCex.nBits == pCexCare.nBits, "pCex.nBits == pCexCare.nBits");
    // start the counter-example
    pNew = Abc_CexAlloc( pCex.nRegs, Gia_ManObjNum(p), pCex.iFrame + 1 );
    pNew.iFrame = pCex.iFrame;
    pNew.iPo    = pCex.iPo;
    // initialize terminary simulation
    Gia_ObjTerSimSet0( Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachRi
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+k)); k++  {
        Gia_ObjTerSimSet0( pObj );
    }
    for  i = 0; i <= pCex.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            if !Abc_InfoHasBit( pCexCare.pData, pCexCare.nRegs + i * pCexCare.nPis + k ) {
                Gia_ObjTerSimSetX( pObj );
            } else if Abc_InfoHasBit( pCex.pData, pCex.nRegs + i * pCex.nPis + k ) {
                Gia_ObjTerSimSet1( pObj );
            } else {
                Gia_ObjTerSimSet0( pObj ); }
        }
        // inlining of macro Gia_ManForEachRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            Gia_ObjTerSimRo( p, pObj );
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            Gia_ObjTerSimAnd( pObj ); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            Gia_ObjTerSimCo( pObj );
        }
        // add cares to the exdended counter-example
        // inlining of macro Gia_ManForEachObj
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {
            if !Gia_ObjTerSimGetX(pObj) {
                Abc_InfoSetBit( pNew.pData, pNew.nRegs + pNew.nPis * i + k ); }
        }
    }
    pObj = Gia_ManPo( p, pCex.iPo );
    assert(Gia_ObjTerSimGet1(pObj), "Gia_ObjTerSimGet1(pObj)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Forward propagation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexCarePropagateFwdOne(p *Gia_Man_t, pCex *Abc_Cex_t, f int32, vPriosIn *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var Prio, Prio0, Prio1 int32;
    var i, Phase0, Phase1 int32;
    assert(Vec_IntSize(vPriosIn) == pCex.nPis * (pCex.iFrame + 1), "Vec_IntSize(vPriosIn) == pCex.nPis * (pCex.iFrame + 1)");
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Vec_IntEntry( vPriosIn, f * pCex.nPis + i );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Prio0  = Abc_Lit2Var(Gia_ObjFanin0(pObj).Value);
        Prio1  = Abc_Lit2Var(Gia_ObjFanin1(pObj).Value);
        Phase0 = Abc_LitIsCompl(Gia_ObjFanin0(pObj).Value) ^ Gia_ObjFaninC0(pObj);
        Phase1 = Abc_LitIsCompl(Gia_ObjFanin1(pObj).Value) ^ Gia_ObjFaninC1(pObj);
        if Phase0 && (Phase1 != 0) {
            Prio = Abc_MinInt(Prio0, Prio1);
        } else if Phase0 {
            Prio = Prio1;
        } else if Phase1 != 0 {
            Prio = Prio0;
        } else { // if ( !Phase0 && !Phase1 )
            Prio = Abc_MaxInt(Prio0, Prio1); }
        pObj.Value = Abc_Var2Lit( Prio, Phase0 && Phase1 );
        pObj.fPhase = 0;
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Abc_LitNotCond( Gia_ObjFanin0(pObj).Value, Gia_ObjFaninC0(pObj) );
    }
}
func Bmc_CexCarePropagateFwd(p *Gia_Man_t, pCex *Abc_Cex_t, vPriosIn *Vec_Int_t, vPriosFf *Vec_Int_t) {
    var pObjRo, pObjRi *Gia_Obj_t;
    var i int32;
    var f int32;
    var ValueMax int32 = Abc_Var2Lit( pCex.nPis * (pCex.iFrame + 1), 0 );
    Gia_ManConst0( p ).Value = ValueMax;
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObjRi.Value = ValueMax;
    }
    Vec_IntClear( vPriosFf );
    for  f = 0; f <= pCex.iFrame; f++  {
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            Vec_IntPush( vPriosFf, (pObjRo.Value = pObjRi.Value) );
        }
        Bmc_CexCarePropagateFwdOne( p, pCex, f, vPriosIn );
    }
}

/**Function*************************************************************

  Synopsis    [Backward propagation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexCarePropagateBwdOne(p *Gia_Man_t, pCex *Abc_Cex_t, f int32, pCexMin *Abc_Cex_t) {
    var pObj, pFan0, pFan1 *Gia_Obj_t;
    var i, Phase0, Phase1 int32;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fPhase = 0;
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if pObj.fPhase {
            Gia_ObjFanin0(pObj).fPhase = 1; }
    }
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        if !pObj.fPhase {
            continue; }
        pFan0 = Gia_ObjFanin0(pObj);
        pFan1 = Gia_ObjFanin1(pObj);
        Phase0 = Abc_LitIsCompl(pFan0.Value) ^ Gia_ObjFaninC0(pObj);
        Phase1 = Abc_LitIsCompl(pFan1.Value) ^ Gia_ObjFaninC1(pObj);
        if Phase0 && (Phase1 != 0) {
            pFan0.fPhase = 1;
            pFan1.fPhase = 1;
        
        } else if Phase0 {
            pFan1.fPhase = 1;
        } else if Phase1 != 0 {
            pFan0.fPhase = 1;
        } else { // if ( !Phase0 && !Phase1 )
        
            if pFan0.fPhase || pFan1.fPhase {
                continue; }
            if Gia_ObjIsPi(p, pFan0) {
                pFan0.fPhase = 1;
            } else if Gia_ObjIsPi(p, pFan1) {
                pFan1.fPhase = 1;
//            else if ( Gia_ObjIsAnd(pFan0) && Txs_ObjIsJust(p, pFan0) )
//                pFan0->fPhase = 1;
//            else if ( Gia_ObjIsAnd(pFan1) && Txs_ObjIsJust(p, pFan1) )
//                pFan1->fPhase = 1;
            } else {
                if Abc_Lit2Var(pFan0.Value) > Abc_Lit2Var(pFan1.Value) {
                    pFan0.fPhase = 1;
                } else {
                    pFan1.fPhase = 1; }
            }
        }
    }
    }
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if pObj.fPhase {
            Abc_InfoSetBit( pCexMin.pData, pCexMin.nRegs + pCexMin.nPis * f + i ); }
    }
}
func Bmc_CexCarePropagateBwd(p *Gia_Man_t, pCex *Abc_Cex_t, vPriosIn *Vec_Int_t, vPriosFf *Vec_Int_t) *Abc_Cex_t {
    var pCexMin *Abc_Cex_t;
    var pObjRo, pObjRi *Gia_Obj_t;
    var f, i int32;
    pCexMin = Abc_CexAlloc( pCex.nRegs, pCex.nPis, pCex.iFrame + 1 );
    pCexMin.iPo    = pCex.iPo;
    pCexMin.iFrame = pCex.iFrame;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObjRi) = Gia_ManCo(p, i)); i++  {
        pObjRi.fPhase = 0;
    }
    for  f = pCex.iFrame; f >= 0; f--  {
        Gia_ManPo(p, pCex.iPo).fPhase = int32((f == pCex.iFrame));
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            pObjRo.Value = Vec_IntEntry( vPriosFf, f * pCex.nRegs + i );
        }
        Bmc_CexCarePropagateFwdOne( p, pCex, f, vPriosIn );
        Bmc_CexCarePropagateBwdOne( p, pCex, f, pCexMin );
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            pObjRi.fPhase = pObjRo.fPhase;
        }
    }
    return pCexMin;
}

/**Function*************************************************************

  Synopsis    [Computes the care set of the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexCareTotal(pCexes **Abc_Cex_t, nCexes int32) *Abc_Cex_t {
    var i int32;
    var k int32;
    var nWords int32 = Abc_BitWordNum( pCexes[0].nBits );
    var pCexMin *Abc_Cex_t  = Abc_CexAlloc( pCexes[0].nRegs, pCexes[0].nPis, pCexes[0].iFrame + 1 );
    pCexMin.iPo    = pCexes[0].iPo;
    pCexMin.iFrame = pCexes[0].iFrame;
    for  i = 0; i < nWords; i++  {
        pCexMin.pData[i] = pCexes[0].pData[i];
        for  k = 1; k < nCexes; k++  {
            pCexMin.pData[i] &= pCexes[k].pData[i];
        }
    }
    return pCexMin;
}
func Bmc_CexCareMinimizeAig(p *Gia_Man_t, nRealPis int32, pCex *Abc_Cex_t, nTryCexes int32, fCheck int32, fVerbose int32) *Abc_Cex_t {
    //int nTryCexes = 4; // belongs to range [1;4]
    var pCexBest *Abc_Cex_t;
    var pCexMin [4]*Abc_Cex_t = [4]*Abc_Cex_t{nil};
    var k, f, i, nOnesBest, nOnesCur, Counter int32;
    var vPriosIn, vPriosFf *Vec_Int_t;
    if pCex.nPis != Gia_ManPiNum(p) {
        fmt.Printf("Given CEX does to have same number of inputs as the AIG.\n");
        return nil;
    }
    if pCex.nRegs != Gia_ManRegNum(p) {
        fmt.Printf("Given CEX does to have same number of flops as the AIG.\n");
        return nil;
    }
    if !(pCex.iPo >= 0 && pCex.iPo < Gia_ManPoNum(p)) {
        fmt.Printf("Given CEX has PO whose index is out of range for the AIG.\n");
        return nil;
    }
    assert(pCex.nPis == Gia_ManPiNum(p), "pCex.nPis == Gia_ManPiNum(p)");
    assert(pCex.nRegs == Gia_ManRegNum(p), "pCex.nRegs == Gia_ManRegNum(p)");
    assert(pCex.iPo >= 0 && pCex.iPo < Gia_ManPoNum(p), "pCex.iPo >= 0 && pCex.iPo < Gia_ManPoNum(p)");
    if fVerbose != 0 {
        fmt.Printf("Original :    ");
        Bmc_CexPrint( pCex, nRealPis, 0 );
    }
    vPriosIn = Vec_IntAlloc( pCex.nPis  * (pCex.iFrame + 1) );
    vPriosFf = Vec_IntAlloc( pCex.nRegs * (pCex.iFrame + 1) );
    for  k = 0; k < nTryCexes; k++  {
        Counter = 0;
        Vec_IntFill( vPriosIn, pCex.nPis * (pCex.iFrame + 1), 0 );
/*
        if ( k == 0 )
        {
            for ( f = 0; f <= pCex->iFrame; f++ )
                for ( i = nRealPis; i < Gia_ManPiNum(p); i++ )
                    Vec_IntWriteEntry( vPriosIn, f * pCex->nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex->pData, pCex->nRegs + pCex->nPis * f + i)) );
            for ( f = 0; f <= pCex->iFrame; f++ )
                for ( i = 0; i < nRealPis; i++ )
                    Vec_IntWriteEntry( vPriosIn, f * pCex->nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex->pData, pCex->nRegs + pCex->nPis * f + i)) );
        }
        else if ( k == 1 )
        {
            for ( f = pCex->iFrame; f >= 0; f-- )
                for ( i = nRealPis; i < Gia_ManPiNum(p); i++ )
                    Vec_IntWriteEntry( vPriosIn, f * pCex->nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex->pData, pCex->nRegs + pCex->nPis * f + i)) );
            for ( f = pCex->iFrame; f >= 0; f-- )
                for ( i = 0; i < nRealPis; i++ )
                    Vec_IntWriteEntry( vPriosIn, f * pCex->nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex->pData, pCex->nRegs + pCex->nPis * f + i)) );
        }
        else if ( k == 2 )
        {
            for ( f = 0; f <= pCex->iFrame; f++ )
                for ( i = Gia_ManPiNum(p) - 1; i >= nRealPis; i-- )
                    Vec_IntWriteEntry( vPriosIn, f * pCex->nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex->pData, pCex->nRegs + pCex->nPis * f + i)) );
            for ( f = 0; f <= pCex->iFrame; f++ )
                for ( i = nRealPis - 1; i >= 0; i-- )
                    Vec_IntWriteEntry( vPriosIn, f * pCex->nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex->pData, pCex->nRegs + pCex->nPis * f + i)) );
        }
        else if ( k == 3 )
        {
            for ( f = pCex->iFrame; f >= 0; f-- )
                for ( i = Gia_ManPiNum(p) - 1; i >= nRealPis; i-- )
                    Vec_IntWriteEntry( vPriosIn, f * pCex->nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex->pData, pCex->nRegs + pCex->nPis * f + i)) );
            for ( f = pCex->iFrame; f >= 0; f-- )
                for ( i = nRealPis - 1; i >= 0; i-- )
                    Vec_IntWriteEntry( vPriosIn, f * pCex->nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex->pData, pCex->nRegs + pCex->nPis * f + i)) );
        }
        else assert( 0 );
*/
        if k == 0 {
            for  f = pCex.iFrame; f >= 0; f--  {
                for  i = Gia_ManPiNum(p) - 1; i >= nRealPis; i--  {
                    Vec_IntWriteEntry( vPriosIn, f * pCex.nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i)) );
                }
            }
            for  f = pCex.iFrame; f >= 0; f--  {
                for  i = nRealPis - 1; i >= 0; i--  {
                    Vec_IntWriteEntry( vPriosIn, f * pCex.nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i)) );
                }
            }
        
        } else if k == 1 {
            for  f = pCex.iFrame; f >= 0; f--  {
                for  i = Gia_ManPiNum(p) - 1; i >= nRealPis; i--  {
                    Vec_IntWriteEntry( vPriosIn, f * pCex.nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i)) );
                }
            }
            for  f = pCex.iFrame; f >= 0; f--  {
                for  i = 0; i < nRealPis; i++  {
                    Vec_IntWriteEntry( vPriosIn, f * pCex.nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i)) );
                }
            }
        
        } else if k == 2 {
            for  f = pCex.iFrame; f >= 0; f--  {
                for  i = nRealPis; i < Gia_ManPiNum(p); i++  {
                    Vec_IntWriteEntry( vPriosIn, f * pCex.nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i)) );
                }
            }
            for  f = pCex.iFrame; f >= 0; f--  {
                for  i = nRealPis - 1; i >= 0; i--  {
                    Vec_IntWriteEntry( vPriosIn, f * pCex.nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i)) );
                }
            }
        
        } else if k == 3 {
            for  f = pCex.iFrame; f >= 0; f--  {
                for  i = nRealPis; i < Gia_ManPiNum(p); i++  {
                    Vec_IntWriteEntry( vPriosIn, f * pCex.nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i)) );
                }
            }
            for  f = pCex.iFrame; f >= 0; f--  {
                for  i = 0; i < nRealPis; i++  {
                    Vec_IntWriteEntry( vPriosIn, f * pCex.nPis + i, Abc_Var2Lit(Counter++, Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i)) );
                }
            }
        
        } else { assert(false, "0"); }

        assert(Counter == pCex.nPis * (pCex.iFrame + 1), "Counter == pCex.nPis * (pCex.iFrame + 1)");
        Bmc_CexCarePropagateFwd( p, pCex, vPriosIn, vPriosFf );
        assert(Vec_IntSize(vPriosFf) == pCex.nRegs * (pCex.iFrame + 1), "Vec_IntSize(vPriosFf) == pCex.nRegs * (pCex.iFrame + 1)");
        if !Abc_LitIsCompl(Gia_ManPo(p, pCex.iPo).Value) {
            fmt.Printf("Counter-example is invalid.\n");
            Vec_IntFree( vPriosIn );
            Vec_IntFree( vPriosFf );
            return nil;
        }
        pCexMin[k] = Bmc_CexCarePropagateBwd( p, pCex, vPriosIn, vPriosFf );
        if fVerbose != 0 {
            if k == 0 {
                fmt.Printf("PI-  PPI-:    ");
            } else if k == 1 {
                fmt.Printf("PI+  PPI-:    ");
            } else if k == 2 {
                fmt.Printf("PI-  PPI+:    ");
            } else if k == 3 {
                fmt.Printf("PI+  PPI+:    ");
            } else { assert(false, "0"); }
            Bmc_CexPrint( pCexMin[k], nRealPis, 0 );
        }
    }
    Vec_IntFree( vPriosIn );
    Vec_IntFree( vPriosFf );
    // select the best one
    pCexBest  = pCexMin[0];
    nOnesBest = Abc_CexCountOnes(pCexMin[0]);
    for  k = 1; k < nTryCexes; k++  {
        if pCexMin[k] == nil {
            continue; }
        nOnesCur = Abc_CexCountOnes(pCexMin[k]);
        if nOnesBest > nOnesCur {
            nOnesBest = nOnesCur;
            pCexBest  = pCexMin[k];
        }
    }
    if fVerbose != 0 {
        //Abc_Cex_t * pTotal = Bmc_CexCareTotal( pCexMin, nTryCexes );
        fmt.Printf("Final    :    ");
        Bmc_CexPrint( pCexBest, nRealPis, 0 );
        //printf( "Total    :    " );
        //Bmc_CexPrint( pTotal, nRealPis, 0 );
        //Abc_CexFreeP( &pTotal );
    }
    for  k = 0; k < nTryCexes; k++  {
        if pCexMin[k] && pCexBest != pCexMin[k] {
            Abc_CexFreeP( &pCexMin[k] ); }
    }
    // verify and return
    if !Bmc_CexVerify( p, pCex, pCexBest ) {
        fmt.Printf("Counter-example verification has failed.\n");
    } else if fCheck != 0 { 
        fmt.Printf("Counter-example verification succeeded.\n"); }
    return pCexBest;
}
func Bmc_CexCareMinimize(p *Aig_Man_t, nRealPis int32, pCex *Abc_Cex_t, nTryCexes int32, fCheck int32, fVerbose int32) *Abc_Cex_t {
    var pGia *Gia_Man_t  = Gia_ManFromAigSimple( p );
    var pCexMin *Abc_Cex_t  = Bmc_CexCareMinimizeAig( pGia, nRealPis, pCex, nTryCexes, fCheck, fVerbose );
    Gia_ManStop( pGia );
    return pCexMin;
}
func Bmc_CexCareSatBasedMinimize(p *Aig_Man_t, nRealPis int32, pCex *Abc_Cex_t, fHighEffort int32, fCheck int32, fVerbose int32) *Abc_Cex_t {
    var pGia *Gia_Man_t  = Gia_ManFromAigSimple( p );
    var pCexMin *Abc_Cex_t  = Bmc_CexCareSatBasedMinimizeAig( pGia, pCex, fHighEffort, fVerbose );
    if pCexMin == nil {
        Gia_ManStop( pGia );
        return nil;
    }
    // unfortunately, cannot verify - ternary simulation does not work
    Gia_ManStop( pGia );
    return pCexMin;
}

/**Function*************************************************************

  Synopsis    [Verifies the care set of the counter-example.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexCareVerify(p *Aig_Man_t, pCex *Abc_Cex_t, pCexMin *Abc_Cex_t, fVerbose int32) {
    var pGia *Gia_Man_t  = Gia_ManFromAigSimple( p );
    if fVerbose != 0 {
        fmt.Printf("Original :    ");
        Bmc_CexPrint( pCex, Gia_ManPiNum(pGia), 0 );
        fmt.Printf("Minimized:    ");
        Bmc_CexPrint( pCexMin, Gia_ManPiNum(pGia), 0 );
    }
    if !Bmc_CexVerify( pGia, pCex, pCexMin ) {
        fmt.Printf("Counter-example verification has failed.\n");
    } else { 
        fmt.Printf("Counter-example verification succeeded.\n"); }
    Gia_ManStop( pGia );
}
/*
    {
        Aig_Man_t * pAig = Abc_NtkToDar( pNtk, 0, 1 );
        Abc_Cex_t * pCex = Bmc_CexCareMinimize( pAig, 3*Saig_ManPiNum(pAig)/4, 4, pAbc->pCex, 1, 1 );
        Aig_ManStop( pAig );
        Abc_CexFree( pCex );
    }
*/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

