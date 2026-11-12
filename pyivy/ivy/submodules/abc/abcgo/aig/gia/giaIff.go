package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaIff.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Hierarchical mapping of AIG with white boxes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaIff.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "map/if/if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Iff_Man_t struct {
    pGia      *Gia_Man_t;       // mapped GIA
    pLib    *If_LibLut_t;       // LUT library
    nLutSize             int32;   // LUT size
    nDegree             int32;    // degree 
    vTimes      *Vec_Flt_t;     // arrival times
    Vec_Int_t *     vMatch[4];  // matches
}

func Iff_ObjTimeId(p *Iff_Man_t, iObj int32) float32 { return Vec_FltEntry( p.vTimes, iObj );                       }
func Iff_ObjTime(p *Iff_Man_t, pObj *Gia_Obj_t) float32 { return Iff_ObjTimeId( p, Gia_ObjId(p.pGia, pObj) );          }
func Iff_ObjSetTimeId(p *Iff_Man_t, iObj int32, Time float32) { Vec_FltWriteEntry( p.vTimes, iObj, Time );                   }
func Iff_ObjSetTime(p *Iff_Man_t, pObj *Gia_Obj_t, Time float32) { Iff_ObjSetTimeId( p, Gia_ObjId(p.pGia, pObj), Time );        }

func Iff_ObjMatchId(p *Iff_Man_t, iObj int32, Type int32) int32 { return Vec_IntEntry( p.vMatch[Type], iObj );                 }
func Iff_ObjMatch(p *Iff_Man_t, pObj *Gia_Obj_t, Type int32) int32 { return Iff_ObjMatchId( p, Gia_ObjId(p.pGia, pObj), Type );   }
func Iff_ObjSetMatchId(p *Iff_Man_t, iObj int32, Type int32, Match int32) { Vec_IntWriteEntry( p.vMatch[Type], iObj, Match );            }
func Iff_ObjSetMatch(p *Iff_Man_t, pObj *Gia_Obj_t, Type int32, Match int32) { Iff_ObjSetMatchId( p, Gia_ObjId(p.pGia, pObj), Type, Match );}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIffStart(pGia *Gia_Man_t) *Iff_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Iff_Man_t  = &Iff_Man_t{};
    p.vTimes    = Vec_FltStartFull( Gia_ManObjNum(pGia) );
    p.vMatch[2] = Vec_IntStartFull( Gia_ManObjNum(pGia) );
    p.vMatch[3] = Vec_IntStartFull( Gia_ManObjNum(pGia) );
    return p;
}
func Gia_ManIffStop(p *Iff_Man_t) {
    Vec_FltFree( p.vTimes );
    Vec_IntFree( p.vMatch[2] );
    Vec_IntFree( p.vMatch[3] );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Count the number of unique fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IffObjCount(pGia *Gia_Man_t, iObj int32, iFaninSkip2 int32, iFaninSkip3 int32) int32 {
    var i, iFanin, Count int32;
    Gia_ManIncrementTravId( pGia );
    // inlining of macro Gia_LutForEachFanin
    for  i = 0; i < Gia_ObjLutSize(pGia,iObj) && ((iFanin = Gia_ObjLutFanins(pGia,iObj)[i]),1); i++  {
        if iFanin == iFaninSkip2 || iFanin == iFaninSkip3 {
            continue; }
        if Gia_ObjIsTravIdCurrentId( pGia, iFanin ) {
            continue; }
        Gia_ObjSetTravIdCurrentId( pGia, iFanin );
        Count++;
    }
    if iFaninSkip2 >= 0 {
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(pGia,iFaninSkip2) && ((iFanin = Gia_ObjLutFanins(pGia,iFaninSkip2)[i]),1); i++  {
            if iFanin == iFaninSkip3 {
                continue; }
            if Gia_ObjIsTravIdCurrentId( pGia, iFanin ) {
                continue; }
            Gia_ObjSetTravIdCurrentId( pGia, iFanin );
            Count++;
        }
    }
    if iFaninSkip3 >= 0 {
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(pGia,iFaninSkip3) && ((iFanin = Gia_ObjLutFanins(pGia,iFaninSkip3)[i]),1); i++  {
            if iFanin == iFaninSkip2 {
                continue; }
            if Gia_ObjIsTravIdCurrentId( pGia, iFanin ) {
                continue; }
            Gia_ObjSetTravIdCurrentId( pGia, iFanin );
            Count++;
        }
    }
    return Count;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IffObjTimeOne(p *Iff_Man_t, iObj int32, iFaninSkip2 int32, iFaninSkip3 int32) float32 {
    var i, iFanin int32;
    var DelayMax float32  = -ABC_INFINITY;
    // inlining of macro Gia_LutForEachFanin
    for  i = 0; i < Gia_ObjLutSize(p.pGia,iObj) && ((iFanin = Gia_ObjLutFanins(p.pGia,iObj)[i]),1); i++  {
        if iFanin != iFaninSkip2 && iFanin != iFaninSkip3 && DelayMax < Iff_ObjTimeId(p, iFanin) {
            DelayMax = Iff_ObjTimeId(p, iFanin); }
    }
    assert(i == Gia_ObjLutSize(p.pGia, iObj), "i == Gia_ObjLutSize(p.pGia, iObj)");
    if iFaninSkip2 == -1 {
        return DelayMax; }
    // inlining of macro Gia_LutForEachFanin
    for  i = 0; i < Gia_ObjLutSize(p.pGia,iFaninSkip2) && ((iFanin = Gia_ObjLutFanins(p.pGia,iFaninSkip2)[i]),1); i++  {
        if iFanin != iFaninSkip3 && DelayMax < Iff_ObjTimeId(p, iFanin) {
            DelayMax = Iff_ObjTimeId(p, iFanin); }
    }
    if iFaninSkip3 == -1 {
        return DelayMax; }
    // inlining of macro Gia_LutForEachFanin
    for  i = 0; i < Gia_ObjLutSize(p.pGia,iFaninSkip3) && ((iFanin = Gia_ObjLutFanins(p.pGia,iFaninSkip3)[i]),1); i++  {
        if iFanin != iFaninSkip2 && DelayMax < Iff_ObjTimeId(p, iFanin) {
            DelayMax = Iff_ObjTimeId(p, iFanin); }
    }
    assert(DelayMax >= 0, "DelayMax >= 0");
    return DelayMax;
}
func Gia_IffObjTimeTwo(p *Iff_Man_t, iObj int32, piFanin *int32, DelayMin float32) float32 {
    var i, iFanin, nSize int32;
    var This float32;
    *piFanin = -1;
    // inlining of macro Gia_LutForEachFanin
    for  i = 0; i < Gia_ObjLutSize(p.pGia,iObj) && ((iFanin = Gia_ObjLutFanins(p.pGia,iObj)[i]),1); i++  {
        if Gia_ObjIsCi(Gia_ManObj(p.pGia, iFanin)) {
            continue; }
        This = Gia_IffObjTimeOne( p, iObj, iFanin, -1 );
        nSize = Gia_IffObjCount( p.pGia, iObj, iFanin, -1 );
        assert(nSize <= p.pLib.LutMax, "nSize <= p.pLib.LutMax");
        This += p.pLib.pLutDelays[nSize][0];
        if DelayMin > This {
            DelayMin = This;
            *piFanin = iFanin;
        }
    }
    return DelayMin;
}
func Gia_IffObjTimeThree(p *Iff_Man_t, iObj int32, piFanin *int32, piFanin2 *int32, DelayMin float32) float32 {
    var i, k, iFanin, iFanin2, nSize int32;
    var This float32;
    *piFanin = -1;
    *piFanin2 = -1;
    // inlining of macro Gia_LutForEachFanin
    for  i = 0; i < Gia_ObjLutSize(p.pGia,iObj) && ((iFanin = Gia_ObjLutFanins(p.pGia,iObj)[i]),1); i++  {
    // inlining of macro Gia_LutForEachFanin
    for  k = 0; k < Gia_ObjLutSize(p.pGia,iObj) && ((iFanin2 = Gia_ObjLutFanins(p.pGia,iObj)[k]),1); k++  {
        if iFanin == iFanin2 {
            continue; }
        if Gia_ObjIsCi(Gia_ManObj(p.pGia, iFanin)) {
            continue; }
        if Gia_ObjIsCi(Gia_ManObj(p.pGia, iFanin2)) {
            continue; }
        This  = Gia_IffObjTimeOne( p, iObj, iFanin, iFanin2 );
        nSize = Gia_IffObjCount( p.pGia, iObj, iFanin, iFanin2 );
        assert(nSize <= p.pLib.LutMax, "nSize <= p.pLib.LutMax");
        This += p.pLib.pLutDelays[nSize][0];
        if DelayMin > This {
            DelayMin  = This;
            *piFanin  = iFanin;
            *piFanin2 = iFanin2;
        }
    }
    }
    return DelayMin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIffPerform(pGia *Gia_Man_t, pLib *If_LibLut_t, pTime *Tim_Man_t, nLutSize int32, nDegree int32) *Iff_Man_t {
    var p *Iff_Man_t;
    var pObj *Gia_Obj_t;
    var iObj, iFanin, iFanin1, iFanin2 int32;
    var CountAll, Count2, Count3 int32;
    var arrTime1 float32;
    var arrTime2 float32;
    var arrTime3 float32;
    var arrMax float32 = -ABC_INFINITY; 
    assert(nDegree == 2 || nDegree == 3, "nDegree == 2 || nDegree == 3");
    // start the mapping manager and set its parameters
    p = Gia_ManIffStart( pGia );
    p.pGia     = pGia;
    p.pLib     = pLib;
    p.nLutSize = nLutSize;
    p.nDegree  = nDegree;
    // compute arrival times of each node
    Iff_ObjSetTimeId( p, 0, 0 );
    Tim_ManIncrementTravId( pTime );
    // inlining of macro Gia_ManForEachObj1
    for  iObj = 1; (iObj < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, iObj)); iObj++  {
        if Gia_ObjIsAnd(pObj) {
            if !Gia_ObjIsLut(pGia, iObj) {
                continue; }
            CountAll++;
            // compute arrival times of LUT inputs
            arrTime1 = Gia_IffObjTimeOne( p, iObj, -1, -1 );
            arrTime1 += p.pLib.pLutDelays[Gia_ObjLutSize(pGia, iObj)][0];
            // compute arrival times of LUT pairs
            arrTime2 = Gia_IffObjTimeTwo( p, iObj, &iFanin, arrTime1 );
            if nDegree == 2 {
                // set arrival times
                Iff_ObjSetTimeId( p, iObj, arrTime2 );
                if arrTime2 < arrTime1 {
                    Iff_ObjSetMatchId( p, iObj, 2, iFanin );
Count2++; }                
            
            } else if nDegree == 3 {                
                // compute arrival times of LUT triples
                arrTime3 = Gia_IffObjTimeThree( p, iObj, &iFanin1, &iFanin2, arrTime2 );
                // set arrival times
                Iff_ObjSetTimeId( p, iObj, arrTime3 );
                if arrTime3 == arrTime1 {
                    continue; }
                if arrTime3 == arrTime2 {
                    Iff_ObjSetMatchId( p, iObj, 2, iFanin );
Count2++;
                } else {
                    assert(arrTime3 < arrTime2, "arrTime3 < arrTime2");
                    Iff_ObjSetMatchId( p, iObj, 2, iFanin1 );
                    Iff_ObjSetMatchId( p, iObj, 3, iFanin2 );
Count3++;
                }
            
            } else { assert(false, "0"); }
        
        } else if Gia_ObjIsCi(pObj) {
            arrTime1 = Tim_ManGetCiArrival( pTime, Gia_ObjCioId(pObj) );
            Iff_ObjSetTime( p, pObj, arrTime1 );
        
        } else if Gia_ObjIsCo(pObj) {
            arrTime1 = Iff_ObjTimeId( p, Gia_ObjFaninId0p(pGia, pObj) );
            Tim_ManSetCoArrival( pTime, Gia_ObjCioId(pObj), arrTime1 );
            Iff_ObjSetTime( p, pObj, arrTime1 );
            arrMax = Abc_MaxFloat( arrMax, arrTime1 );
        
        } else { assert(false, "0"); }
    }
    fmt.Printf("Max delay = %.2f.  Count1 = %d.  Count2 = %d.  Count3 = %d.\n", arrMax, CountAll - Count2 - Count3, Count2, Count3);
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Gia_ManIffSelect_rec(p *Iff_Man_t, iObj int32, vPacking *Vec_Int_t) {
    var i, iFanin, iFaninSkip2, iFaninSkip3 int32;
    if Gia_ObjIsTravIdCurrentId( p.pGia, iObj ) {
        return; }
    Gia_ObjSetTravIdCurrentId( p.pGia, iObj );
    assert(Gia_ObjIsLut(p.pGia, iObj), "Gia_ObjIsLut(p.pGia, iObj)");
    iFaninSkip2 = Iff_ObjMatchId(p, iObj, 2);
    iFaninSkip3 = Iff_ObjMatchId(p, iObj, 3);
    if iFaninSkip2 == -1 {
        assert(iFaninSkip3 == -1, "iFaninSkip3 == -1");
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p.pGia,iObj) && ((iFanin = Gia_ObjLutFanins(p.pGia,iObj)[i]),1); i++  {
            Gia_ManIffSelect_rec( p, iFanin, vPacking );
        }
        Vec_IntPush( vPacking, 1 );
        Vec_IntPush( vPacking, iObj );
    
    } else if iFaninSkip3 == -1 {
        assert(iFaninSkip2 > 0, "iFaninSkip2 > 0");
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p.pGia,iObj) && ((iFanin = Gia_ObjLutFanins(p.pGia,iObj)[i]),1); i++  {
            if iFanin != iFaninSkip2 {
                Gia_ManIffSelect_rec( p, iFanin, vPacking ); }
        }
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p.pGia,iFaninSkip2) && ((iFanin = Gia_ObjLutFanins(p.pGia,iFaninSkip2)[i]),1); i++  {
            Gia_ManIffSelect_rec( p, iFanin, vPacking );
        }
        Vec_IntPush( vPacking, 2 );
        Vec_IntPush( vPacking, iFaninSkip2 );
        Vec_IntPush( vPacking, iObj );
    
    } else {
        assert(iFaninSkip2 > 0 && iFaninSkip3 > 0, "iFaninSkip2 > 0 && iFaninSkip3 > 0");
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p.pGia,iObj) && ((iFanin = Gia_ObjLutFanins(p.pGia,iObj)[i]),1); i++  {
            if iFanin != iFaninSkip2 && iFanin != iFaninSkip3 {
                Gia_ManIffSelect_rec( p, iFanin, vPacking ); }
        }
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p.pGia,iFaninSkip2) && ((iFanin = Gia_ObjLutFanins(p.pGia,iFaninSkip2)[i]),1); i++  {
            if iFanin != iFaninSkip3 {
                Gia_ManIffSelect_rec( p, iFanin, vPacking ); }
        }
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p.pGia,iFaninSkip3) && ((iFanin = Gia_ObjLutFanins(p.pGia,iFaninSkip3)[i]),1); i++  {
            if iFanin != iFaninSkip2 {
                Gia_ManIffSelect_rec( p, iFanin, vPacking ); }
        }
        Vec_IntPush( vPacking, 3 );
        Vec_IntPush( vPacking, iFaninSkip2 );
        Vec_IntPush( vPacking, iFaninSkip3 );
        Vec_IntPush( vPacking, iObj );
    }
    Vec_IntAddToEntry( vPacking, 0, 1 );
}
func Gia_ManIffSelect(p *Iff_Man_t) *Vec_Int_t {
    var vPacking *Vec_Int_t;
    var pObj *Gia_Obj_t; var i int32;
    vPacking = Vec_IntAlloc( Gia_ManObjNum(p.pGia) );
    Vec_IntPush( vPacking, 0 );
    // mark const0 and PIs
    Gia_ManIncrementTravId( p.pGia );
    Gia_ObjSetTravIdCurrentId( p.pGia, 0 );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((pObj) = Gia_ManCi(p.pGia, i)); i++  {
        Gia_ObjSetTravIdCurrent( p.pGia, pObj );
    }
    // recursively collect internal nodes
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Gia_ManIffSelect_rec( p, Gia_ObjFaninId0p(p.pGia, pObj), vPacking );
    }
    return vPacking;
}

/**Function*************************************************************

  Synopsis    [This command performs hierarhical mapping.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Gia_ManIffTest(pGia *Gia_Man_t, pLib *If_LibLut_t, fVerbose int32) {
    var p *Iff_Man_t;
    var pTemp *Tim_Man_t  = nil;
    var nDegree int32  = -1;
    var nLutSize int32  = Gia_ManLutSizeMax( pGia );
    if nLutSize <= 4 {
        nLutSize = 4;
        if pLib.LutMax == 7 {
            nDegree = 2;
        } else if pLib.LutMax == 10 {
            nDegree = 3;
        } else { fmt.Printf("LUT library for packing 4-LUTs should have 7 or 10 inputs.\n"); return; }
    
    } else if nLutSize <= 6 {
        nLutSize = 6;
        if pLib.LutMax == 11 {
            nDegree = 2;
        } else if pLib.LutMax == 16 {
            nDegree = 3;
        } else { fmt.Printf("LUT library for packing 6-LUTs should have 11 or 16 inputs.\n"); return; }
    
    } else {
        fmt.Printf("The LUT size is more than 6.\n");
        return;
    }
    if fVerbose != 0 {
        fmt.Printf("Performing %d-clustering with %d-LUTs:\n", nDegree, nLutSize); }
    // create timing manager
    if pGia.pManTime == nil {
        pGia.pManTime = pTemp = Tim_ManStart( Gia_ManCiNum(pGia), Gia_ManCoNum(pGia) ); }
    // perform timing computation
    p = Gia_ManIffPerform( pGia, pLib, (Tim_Man_t *)pGia.pManTime, nLutSize, nDegree );
    // remove timing manager
    if pGia.pManTime == pTemp {
        pGia.pManTime = nil; }
    Tim_ManStopP( (Tim_Man_t **)&pTemp );
    // derive clustering
    Vec_IntFreeP( &pGia.vPacking );
    pGia.vPacking = Gia_ManIffSelect( p );
    Gia_ManIffStop( p );
    // print statistics
    if fVerbose != 0 {
        Gia_ManPrintPackingStats( pGia ); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

