package wlc

import "fmt"

/**CFile****************************************************************

  FileName    [wlcSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Performs sequential simulation of word-level network.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 04, 2015.]

  Revision    [$Id: wlcSim.c,v 1.00 2015/06/04 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_ObjSim(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, p.nSimWords * iObj );
}
func Wlc_ObjSimPi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSim *word  = Wlc_ObjSim( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSim[w] = Gia_ManRandomW( 0 );
    }
}
func Wlc_ObjSimRo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSimRo *word  = Wlc_ObjSim( p, iObj );
    var pSimRi *word  = Wlc_ObjSim( p, Gia_ObjRoToRiId(p, iObj) );
    for  w = 0; w < p.nSimWords; w++  {
        pSimRo[w] = pSimRi[w];
    }
}
func Wlc_ObjSimCo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSimCo *word   = Wlc_ObjSim( p, iObj );
    var pSimDri *word  = Wlc_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] = ~pSimDri[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] =  pSimDri[w];
        } }
}
func Wlc_ObjSimAnd(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSim *word   = Wlc_ObjSim( p, iObj );
    var pSim0 *word  = Wlc_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSim1 *word  = Wlc_ObjSim( p, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & ~pSim1[w];
        }
    } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & pSim1[w];
        }
    } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & ~pSim1[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & pSim1[w];
        } }
}


/**Function*************************************************************

  Synopsis    [Performs simulation of a word-level network.]

  Description [Returns vRes, a 2D array of simulation information for 
  the output of each bit of each object listed in vNodes. In particular, 
  Vec_Ptr_t * vSimObj = (Vec_Ptr_t *)Vec_PtrEntry(vRes, iObj) and
  Vec_Ptr_t * vSimObjBit = (Vec_Ptr_t *)Vec_PtrEntry(vSimObj, iBit)
  are arrays containing the simulation info for each object (vSimObj) 
  and for each output bit of this object (vSimObjBit). Alternatively,
  Vec_Ptr_t * vSimObjBit = Vec_VecEntryEntry( (Vec_Vec_t *)vRes, iObj, iBit ).
  The output bitwidth of an object is Wlc_ObjRange( Wlc_NtkObj(pNtk, iObj) ).
  Simulation information is binary data constaining the given number (nWords)
  of 64-bit machine words for the given number (nFrames) of consecutive 
  timeframes.  The total number of timeframes is nWords * nFrames for 
  each bit of each object.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_NtkDeleteSim(p *Vec_Ptr_t) {
    var pInfo *word; var i, k int32;
    var vVec *Vec_Vec_t  = (Vec_Vec_t *)p;
    Vec_VecForEachEntry( word *, vVec, pInfo, i, k )
        ABC_FREE( pInfo );
    Vec_VecFree( vVec );
}
func Wlc_NtkSimulate(p *Wlc_Ntk_t, vNodes *Vec_Int_t, nWords int32, nFrames int32) *Vec_Ptr_t {
    var pObj *Gia_Obj_t; 
    var vOne, vRes *Vec_Ptr_t;
    var pGia *Gia_Man_t  = Wlc_NtkBitBlast( p, nil );
    var pWlcObj *Wlc_Obj_t;
    var f, i, k, w, nBits, Counter int32;
    // allocate simulation info for one timeframe
    Vec_WrdFreeP( &pGia.vSims );
    pGia.vSims = Vec_WrdStart( Gia_ManObjNum(pGia) * nWords );
    pGia.nSimWords = nWords;
    // allocate resulting simulation info
    vRes = Vec_PtrAlloc( Vec_IntSize(vNodes) );
    // inlining of macro Wlc_NtkForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((pWlcObj) = Wlc_NtkObj(p, Vec_IntEntry(vNodes, i))), 1); i++  {
        nBits = Wlc_ObjRange(pWlcObj);
        vOne = Vec_PtrAlloc( nBits );
        for  k = 0; k < nBits; k++  {
            Vec_PtrPush( vOne, ABC_CALLOC(word, nWords * nFrames) );
        }
        Vec_PtrPush( vRes, vOne ); 
    }
    // perform simulation (const0 and flop outputs are already initialized)
    Gia_ManRandomW( 1 );
    for  f = 0; f < nFrames; f++  {
        Gia_ManForEachObj1( pGia, pObj, i )
        {
            if Gia_ObjIsAnd(pObj) {
                Wlc_ObjSimAnd( pGia, i );
            } else if Gia_ObjIsCo(pObj) {
                Wlc_ObjSimCo( pGia, i );
            } else if Gia_ObjIsPi(pGia, pObj) {
                Wlc_ObjSimPi( pGia, i );
            } else if Gia_ObjIsRo(pGia, pObj) {
                Wlc_ObjSimRo( pGia, i ); }
        }
        // collect simulation data
        // inlining of macro Wlc_NtkForEachObjVec
        for  i = 0; (i < Vec_IntSize(vNodes)) && (((pWlcObj) = Wlc_NtkObj(p, Vec_IntEntry(vNodes, i))), 1); i++  {
            var nBits int32  = Wlc_ObjRange(pWlcObj);
            var iFirst int32  = Vec_IntEntry( &p.vCopies, Wlc_ObjId(p, pWlcObj) );
            for  k = 0; k < nBits; k++  {
                var iLit int32  = Vec_IntEntry( &p.vBits, iFirst + k );
                var pInfo *word  = (word*)Vec_VecEntryEntry( (Vec_Vec_t *)vRes, i, k );
                if iLit == -1 {
                    Counter++;
                    for  w = 0; w < nWords; w++  {
                        pInfo[f * nWords + w] = 0;
                    }
                
                } else {
                    var pInfoObj *word  = Wlc_ObjSim( pGia, Abc_Lit2Var(iLit) );
                    for  w = 0; w < nWords; w++  {
                        pInfo[f * nWords + w] = Abc_LitIsCompl(iLit) ? ~pInfoObj[w] : pInfoObj[w];
                    }
                }
            }
        }
        if f == 0 && (Counter != 0) {
            fmt.Printf("Replaced %d dangling internal bits with constant 0.\n", Counter); }
    }
    Vec_WrdFreeP( &pGia.vSims );
    pGia.nSimWords = 0;
    Gia_ManStop( pGia );
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description [This testing procedure assumes that the WLC network has 
  one node, which is a multiplier. It simulates the node and checks the 
  word-level interpretation of the bit-level simulation info to make sure 
  that it indeed represents multiplication.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_NtkSimulatePrint(p *Wlc_Ntk_t, vNodes *Vec_Int_t, vRes *Vec_Ptr_t, nWords int32, nFrames int32) {
    var pWlcObj *Wlc_Obj_t; 
    var f, w, b, i, k, iPat int32;
    for  f = 0; f < nFrames; f++, printf("\n")  {
      for  w = 0; w < nWords; w++  {
        for  b = 0; b < 64; b++, iPat++, printf("\n")  {
            // inlining of macro Wlc_NtkForEachObjVec
            for  i = 0; (i < Vec_IntSize(vNodes)) && (((pWlcObj) = Wlc_NtkObj(p, Vec_IntEntry(vNodes, i))), 1); i++  {
                var nBits int32  = Wlc_ObjRange(pWlcObj);
                for  k = nBits-1; k >= 0; k--  {
                    var pInfo *word  = (word*)Vec_VecEntryEntry( (Vec_Vec_t *)vRes, i, k );
                    fmt.Printf("%d", Abc_InfoHasBit((unsigned *)pInfo, iPat));
                }
                fmt.Printf(" ");
            }
        }
      }
    }
}
func Wlc_NtkSimulateTest(p *Wlc_Ntk_t) {
    var nWords int32  = 2;
    var nFrames int32  = 2;
    var vRes *Vec_Ptr_t;
    var vNodes *Vec_Int_t  = Vec_IntAlloc( 3 );
    Vec_IntPush( vNodes, 1 );
    Vec_IntPush( vNodes, 2 );
    Vec_IntPush( vNodes, 3 );
    vRes = Wlc_NtkSimulate( p, vNodes, nWords, nFrames );
    Wlc_NtkSimulatePrint( p, vNodes, vRes, nWords, nFrames );
    Wlc_NtkDeleteSim( vRes );
    Vec_IntFree( vNodes );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

