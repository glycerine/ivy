package saig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigSynch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Computation of synchronizing sequence.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigSynch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//        0  1  x
//       00 01 11
// 0  00 00 00 00
// 1  01 00 01 11
// x  11 00 11 11

func Saig_SynchNot(w uint32) uint32 {
    return w^((~(w&(w>>1)))&0x55555555);
}
func Saig_SynchAnd(u uint32, w uint32) uint32 {
    return (u&w)|((((u&(u>>1)&w&~(w>>1))|(w&(w>>1)&u&~(u>>1)))&0x55555555)<<1);
}
func Saig_SynchRandomBinary() uint32 { 
    return Aig_ManRandom(0) & 0x55555555;               
}
func Saig_SynchRandomTernary() uint32 { 
    var w uint32  = Aig_ManRandom(0);
    return w^((~w)&(w>>1)&0x55555555);        
}
func Saig_SynchTernary(v int32) uint32 {
    assert(v == 0 || v == 1 || v == 3, "v == 0 || v == 1 || v == 3");
    if v != 0 {
        if v==1 {
            return 0x55555555;
        } else {
            return 0xffffffff;
        }
    } else {
        return 0;
    }
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Initializes registers to the ternary state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchSetConstant1(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32) {
    var pObj *Aig_Obj_t;
    var pSim *uint32;
    var w int32;
    pObj = Aig_ManConst1( pAig );
    pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
    for  w = 0; w < nWords; w++  {
        pSim[w] = 0x55555555;
    }
}

/**Function*************************************************************

  Synopsis    [Initializes registers to the ternary state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchInitRegsTernary(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32) {
    var pObj *Aig_Obj_t;
    var pSim *uint32;
    var i, w int32;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        for  w = 0; w < nWords; w++  {
            pSim[w] = 0xffffffff;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Initializes registers to the given binary state.]

  Description [The binary state is stored in pObj->fMarkA.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchInitRegsBinary(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32) {
    var pObj *Aig_Obj_t;
    var pSim *uint32;
    var i, w int32;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        for  w = 0; w < nWords; w++  {
            pSim[w] = Saig_SynchTernary( pObj.fMarkA );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Initializes random binary primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchInitPisRandom(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32) {
    var pObj *Aig_Obj_t;
    var pSim *uint32;
    var i, w int32;
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        for  w = 0; w < nWords; w++  {
            pSim[w] = Saig_SynchRandomBinary();
        }
    }
}

/**Function*************************************************************

  Synopsis    [Initializes random binary primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchInitPisGiven(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32, pValues string) {
    var pObj *Aig_Obj_t;
    var pSim *uint32;
    var i, w int32;
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        for  w = 0; w < nWords; w++  {
            pSim[w] = Saig_SynchTernary( pValues[i] );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchTernarySimulate(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32) {
    var pObj *Aig_Obj_t;
    var pSim0, pSim1, pSim *uint32;
    var i, w int32;
    // simulate nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pSim  = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        pSim0 = (unsigned *)Vec_PtrEntry( vSimInfo, Aig_ObjFaninId0(pObj) );
        pSim1 = (unsigned *)Vec_PtrEntry( vSimInfo, Aig_ObjFaninId1(pObj) );
        if Aig_ObjFaninC0(pObj) && Aig_ObjFaninC1(pObj) {
            for  w = 0; w < nWords; w++  {
                pSim[w] = Saig_SynchAnd( Saig_SynchNot(pSim0[w]), Saig_SynchNot(pSim1[w]) );
            }
        
        } else if !Aig_ObjFaninC0(pObj) && Aig_ObjFaninC1(pObj) {
            for  w = 0; w < nWords; w++  {
                pSim[w] = Saig_SynchAnd( pSim0[w], Saig_SynchNot(pSim1[w]) );
            }
        
        } else if Aig_ObjFaninC0(pObj) && !Aig_ObjFaninC1(pObj) {
            for  w = 0; w < nWords; w++  {
                pSim[w] = Saig_SynchAnd( Saig_SynchNot(pSim0[w]), pSim1[w] );
            }
        
        } else { // if ( !Aig_ObjFaninC0(pObj) && !Aig_ObjFaninC1(pObj) )
        
            for  w = 0; w < nWords; w++  {
                pSim[w] = Saig_SynchAnd( pSim0[w], pSim1[w] );
            }
        }
    }
    }
    // transfer values to register inputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        pSim  = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        pSim0 = (unsigned *)Vec_PtrEntry( vSimInfo, Aig_ObjFaninId0(pObj) );
        if Aig_ObjFaninC0(pObj) {
            for  w = 0; w < nWords; w++  {
                pSim[w] = Saig_SynchNot( pSim0[w] );
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                pSim[w] = pSim0[w];
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchTernaryTransferState(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32) {
    var pObjLi, pObjLo *Aig_Obj_t;
    var pSim0, pSim1 *uint32;
    var i, w int32;
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
        pSim0 = (unsigned *)Vec_PtrEntry( vSimInfo, pObjLi.Id );
        pSim1 = (unsigned *)Vec_PtrEntry( vSimInfo, pObjLo.Id );
        for  w = 0; w < nWords; w++  {
            pSim1[w] = pSim0[w];
        }
    }
}

/**Function*************************************************************

  Synopsis    [Returns the number of Xs in the smallest ternary pattern.]

  Description [Returns the number of this pattern.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchCountX(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32, piPat *int32) int32 {
    var pObj *Aig_Obj_t;
    var pSim *uint32;
    var pCounters *int32;
    var i int32;
    var w int32;
    var b int32;
    var iPatBest, iTernMin int32;
    // count the number of ternary values in each pattern
    // inlining of macro ABC_CALLOC
    pCounters = ((int *) calloc((size_t)(nWords * 16), sizeof(int)));
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        for  w = 0; w < nWords; w++  {
            for  b = 0; b < 16; b++  {
                if ((pSim[w] >> (b << 1)) & 3) == 3 {
                    pCounters[16 * w + b]++; }
            }
        }
    }
    // get the best pattern
    iPatBest = -1;
    iTernMin = 1 + Saig_ManRegNum(pAig);
    for  b = 0; b < 16 * nWords; b++  {
        if iTernMin > pCounters[b] {
            iTernMin = pCounters[b];
            iPatBest = b;
            if iTernMin == 0 {
                break; }
        }
    }
    pCounters = nil;
    *piPat = iPatBest;
    return iTernMin;
}

/**Function*************************************************************

  Synopsis    [Saves the best pattern found and initializes the registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchSavePattern(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, nWords int32, iPat int32, vSequence *Vec_Str_t) int32 {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var pSim *uint32;
    var Counter, Value, i, w int32;
    assert(iPat < 16 * nWords, "iPat < 16 * nWords");
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        Value = (pSim[iPat>>4] >> ((iPat&0xf) << 1)) & 3;
        Vec_StrPush( vSequence, byte(Value) );
//        printf( "%d ", Value );
    }
//    printf( "\n" );
    Counter = 0;
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObjLi.Id );
        Value = (pSim[iPat>>4] >> ((iPat&0xf) << 1)) & 3;
        Counter += (Value == 3);
        // save patern in the same register
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObjLo.Id );
        for  w = 0; w < nWords; w++  {
            pSim[w] = Saig_SynchTernary( Value );
        }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Implement synchronizing sequence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchSequenceRun(pAig *Aig_Man_t, vSimInfo *Vec_Ptr_t, vSequence *Vec_Str_t, fTernary int32) int32 {
    var pSim *uint32;
    var pObj *Aig_Obj_t;
    var Counter, nIters, Value, i int32;
    assert(Vec_StrSize(vSequence) % Saig_ManPiNum(pAig) == 0, "Vec_StrSize(vSequence) % Saig_ManPiNum(pAig) == 0");
    nIters = Vec_StrSize(vSequence) / Saig_ManPiNum(pAig);
    Saig_SynchSetConstant1( pAig, vSimInfo, 1 );
    if fTernary != 0 {
        Saig_SynchInitRegsTernary( pAig, vSimInfo, 1 ); 
    } else {
        Saig_SynchInitRegsBinary( pAig, vSimInfo, 1 ); } 
    for  i = 0; i < nIters; i++  {
        Saig_SynchInitPisGiven( pAig, vSimInfo, 1, Vec_StrArray(vSequence) + i * Saig_ManPiNum(pAig) );
        Saig_SynchTernarySimulate( pAig, vSimInfo, 1 );
        Saig_SynchTernaryTransferState( pAig, vSimInfo, 1 );
    }
    // save the resulting state in the registers
    Counter = 0;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pSim = (unsigned *)Vec_PtrEntry( vSimInfo, pObj.Id );
        Value = pSim[0] & 3;
        assert(Value != 2, "Value != 2");
        Counter += (Value == 3);
        pObj.fMarkA = Value;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Determines synchronizing sequence using ternary simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchSequence(pAig *Aig_Man_t, nWords int32) *Vec_Str_t {
    var nStepsMax int32  = 100;  // the maximum number of simulation steps
    var nTriesMax int32  = 100;  // the maximum number of attempts at each step
    var fVerify int32    =   1;  // verify the resulting pattern
    var vSequence *Vec_Str_t;
    var vSimInfo *Vec_Ptr_t;
    var nTerPrev, nTerCur, nTerCur2 int32;
    var iPatBest, RetValue, s, t int32;
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    // reset random numbers
    Aig_ManRandom( 1 );
    // start the sequence
    vSequence = Vec_StrAlloc( 20 * Saig_ManRegNum(pAig) );
    // create sim info and init registers
    vSimInfo = Vec_PtrAllocSimInfo( Aig_ManObjNumMax(pAig), nWords );
    Saig_SynchSetConstant1( pAig, vSimInfo, nWords );
    // iterate over the timeframes
    nTerPrev = Saig_ManRegNum(pAig);
    Saig_SynchInitRegsTernary( pAig, vSimInfo, nWords );
    for  s = 0; s < nStepsMax && nTerPrev > 0; s++  {
        for  t = 0; t < nTriesMax; t++  {
            Saig_SynchInitPisRandom( pAig, vSimInfo, nWords );
            Saig_SynchTernarySimulate( pAig, vSimInfo, nWords );
            nTerCur = Saig_SynchCountX( pAig, vSimInfo, nWords, &iPatBest );
            if nTerCur < nTerPrev {
                break; }
        }
        if t == nTriesMax {
            break; }
        nTerCur2 = Saig_SynchSavePattern( pAig, vSimInfo, nWords, iPatBest, vSequence );
        assert(nTerCur == nTerCur2, "nTerCur == nTerCur2");
        nTerPrev = nTerCur;
    }
    if nTerPrev > 0 {
        fmt.Printf("Count not initialize %d registers.\n", nTerPrev);
        Vec_PtrFree( vSimInfo );
        Vec_StrFree( vSequence );
        return nil;
    }
    // verify that the sequence is correct
    if fVerify != 0 {
        RetValue = Saig_SynchSequenceRun( pAig, vSimInfo, vSequence, 1 );
        assert(RetValue == 0, "RetValue == 0");
        Aig_ManCleanMarkA( pAig );
    }
    Vec_PtrFree( vSimInfo );
    return vSequence;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG to have constant-0 initial state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManDupInitZero(p *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        pObj.pData = Aig_NotCond( Aig_ObjCreateCi( pNew ), pObj.fMarkA );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        pObj.pData = Aig_ObjCreateCo( pNew, Aig_NotCond( Aig_ObjChild0Copy(pObj), pObj.fMarkA ) );
    }
    Aig_ManSetRegNum( pNew, Saig_ManRegNum(p) );
    assert(Aig_ManNodeNum(pNew) == Aig_ManNodeNum(p), "Aig_ManNodeNum(pNew) == Aig_ManNodeNum(p)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Determines synchronizing sequence using ternary simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_SynchSequenceApply(pAig *Aig_Man_t, nWords int32, fVerbose int32) *Aig_Man_t {
    var pAigZero *Aig_Man_t;
    var vSequence *Vec_Str_t;
    var vSimInfo *Vec_Ptr_t;
    var RetValue int32;
    var clk abctime;

clk = Abc_Clock();
    // derive synchronization sequence
    vSequence = Saig_SynchSequence( pAig, nWords );
    if vSequence == nil {
        fmt.Printf("Design 1: Synchronizing sequence is not found. ");
    } else if fVerbose != 0 {
        fmt.Printf("Design 1: Synchronizing sequence of length %4d is found. ", Vec_StrSize(vSequence) / Saig_ManPiNum(pAig)); }
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    
    } else {
        fmt.Printf("\n"); }
    if vSequence == nil {
        fmt.Printf("Quitting synchronization.\n");
        return nil;
    }

    // apply synchronization sequence
    vSimInfo = Vec_PtrAllocSimInfo( Aig_ManObjNumMax(pAig), 1 );
    RetValue = Saig_SynchSequenceRun( pAig, vSimInfo, vSequence, 1 );
    assert(RetValue == 0, "RetValue == 0");
    // duplicate 
    pAigZero = Saig_ManDupInitZero( pAig );
    // cleanup
    Vec_PtrFree( vSimInfo );
    Vec_StrFree( vSequence );
    Aig_ManCleanMarkA( pAig );
    return pAigZero;
}

/**Function*************************************************************

  Synopsis    [Creates SEC miter for two designs without initial state.]

  Description [The designs (pAig1 and pAig2) are assumed to have ternary 
  initial state. Determines synchronizing sequences using ternary simulation.
  Simulates the sequences on both designs to come up with equivalent binary 
  initial states. Create seq miter for the designs starting in these states.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_Synchronize(pAig1 *Aig_Man_t, pAig2 *Aig_Man_t, nWords int32, fVerbose int32) *Aig_Man_t {
    var pAig1z, pAig2z, pMiter *Aig_Man_t;
    var vSeq1, vSeq2 *Vec_Str_t;
    var vSimInfo *Vec_Ptr_t;
    var RetValue int32;
    var clk abctime;
/*
    {
        unsigned u = Saig_SynchRandomTernary();
        unsigned w = Saig_SynchRandomTernary();
        unsigned x = Saig_SynchNot( u );
        unsigned y = Saig_SynchNot( w );
        unsigned z = Saig_SynchAnd( x, y );

        Extra_PrintBinary( stdout, &u, 32 );  printf( "\n" );
        Extra_PrintBinary( stdout, &w, 32 );  printf( "\n" );  printf( "\n" );
        Extra_PrintBinary( stdout, &x, 32 );  printf( "\n" );
        Extra_PrintBinary( stdout, &y, 32 );  printf( "\n" );  printf( "\n" );
        Extra_PrintBinary( stdout, &z, 32 );  printf( "\n" );
    }
*/
    // report statistics
    if fVerbose != 0 {
        fmt.Printf("Design 1: ");
        Aig_ManPrintStats( pAig1 );
        fmt.Printf("Design 2: ");
        Aig_ManPrintStats( pAig2 );
    }

    // synchronize the first design
    clk = Abc_Clock();
    vSeq1 = Saig_SynchSequence( pAig1, nWords );
    if vSeq1 == nil {
        fmt.Printf("Design 1: Synchronizing sequence is not found. ");
    } else if fVerbose != 0 {
        fmt.Printf("Design 1: Synchronizing sequence of length %4d is found. ", Vec_StrSize(vSeq1) / Saig_ManPiNum(pAig1)); }
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    
    } else {
        fmt.Printf("\n"); }

    // synchronize the first design
    clk = Abc_Clock();
    vSeq2 = Saig_SynchSequence( pAig2, nWords );
    if vSeq2 == nil {
        fmt.Printf("Design 2: Synchronizing sequence is not found. ");
    } else if fVerbose != 0 {
        fmt.Printf("Design 2: Synchronizing sequence of length %4d is found. ", Vec_StrSize(vSeq2) / Saig_ManPiNum(pAig2)); }
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    
    } else {
        fmt.Printf("\n"); }

    // quit if one of the designs cannot be synchronized
    if vSeq1 == nil || vSeq2 == nil {
        fmt.Printf("Quitting synchronization.\n");
        if vSeq1 { Vec_StrFree( vSeq1 ); }
        if vSeq2 != nil { Vec_StrFree( vSeq2 ); }
        return nil;
    }
    clk = Abc_Clock();
    vSimInfo = Vec_PtrAllocSimInfo( Abc_MaxInt( Aig_ManObjNumMax(pAig1), Aig_ManObjNumMax(pAig2) ), 1 );

    // process Design 1
    RetValue = Saig_SynchSequenceRun( pAig1, vSimInfo, vSeq1, 1 );
    assert(RetValue == 0, "RetValue == 0");
    RetValue = Saig_SynchSequenceRun( pAig1, vSimInfo, vSeq2, 0 );
    assert(RetValue == 0, "RetValue == 0");

    // process Design 2
    RetValue = Saig_SynchSequenceRun( pAig2, vSimInfo, vSeq2, 1 );
    assert(RetValue == 0, "RetValue == 0");

    // duplicate designs
    pAig1z = Saig_ManDupInitZero( pAig1 );
    pAig2z = Saig_ManDupInitZero( pAig2 );
    pMiter = Saig_ManCreateMiter( pAig1z, pAig2z, 0 );
    Aig_ManCleanup( pMiter );
    Aig_ManStop( pAig1z );
    Aig_ManStop( pAig2z );

    // cleanup
    Vec_PtrFree( vSimInfo );
    Vec_StrFree( vSeq1 );
    Vec_StrFree( vSeq2 );
    Aig_ManCleanMarkA( pAig1 );
    Aig_ManCleanMarkA( pAig2 );

    if fVerbose != 0 {
        fmt.Printf("Miter of the synchronized designs is constructed.         ");
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    return pMiter;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

