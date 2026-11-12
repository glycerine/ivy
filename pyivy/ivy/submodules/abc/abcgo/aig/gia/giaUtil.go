package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
const NUMBER1 uint32 =  3716960521
const NUMBER2 uint32 =  2174103536

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates a sequence or random numbers.]

  Description []
               
  SideEffects []

  SeeAlso     [http://www.codeproject.com/KB/recipes/SimpleRNG.aspx]

***********************************************************************/
func Gia_ManRandom(fReset int32) uint32 {
    static unsigned int m_z = NUMBER1;
    static unsigned int m_w = NUMBER2;
    if fReset != 0 {
        m_z = NUMBER1;
        m_w = NUMBER2;
    }
    m_z = 36969 * (m_z & 65535) + (m_z >> 16);
    m_w = 18000 * (m_w & 65535) + (m_w >> 16);
    return (m_z << 16) + m_w;
}
func Gia_ManRandomW(fReset int32) word { 
    return ((word)Gia_ManRandom(fReset) << 32) | ((word)Gia_ManRandom(fReset) << 0);
}



/**Function*************************************************************

  Synopsis    [Creates random info for the primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRandomInfo(vInfo *Vec_Ptr_t, iInputStart int32, iWordStart int32, iWordStop int32) {
    var pInfo *uint32;
    var i, w int32;
    Vec_PtrForEachEntryStart( unsigned *, vInfo, pInfo, i, iInputStart )
        for ( w = iWordStart; w < iWordStop; w++ )
            pInfo[w] = Gia_ManRandom(0);
}


/**Function*************************************************************

  Synopsis    [Returns the time stamp.]

  Description [The file should be closed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gia_TimeStamp_Buffer [100]byte;
func Gia_TimeStamp() string {
    
    var TimeStamp string;
    var ltime int64;
    // get the current time
    time( &ltime );
    TimeStamp = asctime( localtime( &ltime ) );
    TimeStamp[ len(TimeStamp) - 1 ] = 0;
    strcpy( spStatic_Gia_TimeStamp_Buffer, TimeStamp );
    return spStatic_Gia_TimeStamp_Buffer;
}

/**Function*************************************************************

  Synopsis    [Returns the composite name of the file.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gia_FileNameGenericAppend_Buffer [1000]byte;
func Gia_FileNameGenericAppend(pBase string, pSuffix string) string {
    
    var pDot string;
    strcpy( spStatic_Gia_FileNameGenericAppend_Buffer, pBase );
    if pDot = strrchr( spStatic_Gia_FileNameGenericAppend_Buffer, '.' ) {
        *pDot = 0; }
    strcat( spStatic_Gia_FileNameGenericAppend_Buffer, pSuffix );
    if (pDot = strrchr( spStatic_Gia_FileNameGenericAppend_Buffer, '\\' )) || (pDot = strrchr( spStatic_Gia_FileNameGenericAppend_Buffer, '/' )) {
        return pDot+1; }
    return spStatic_Gia_FileNameGenericAppend_Buffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanMark01(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        pObj.fMark0 = pObj.fMark1 = 0;
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSetMark0(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        pObj.fMark0 = 1;
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanMark0(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        pObj.fMark0 = 0;
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckMark0(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSetMark1(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        pObj.fMark1 = 1;
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanMark1(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        pObj.fMark1 = 0;
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckMark1(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        assert(pObj.fMark1 == 0, "pObj.fMark1 == 0");
}

/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanValue(p *Gia_Man_t) {
    var i int32;
    for  i = 0; i < p.nObjs; i++  {
        p.pObjs[i].Value = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFillValue(p *Gia_Man_t) {
    var i int32;
    for  i = 0; i < p.nObjs; i++  {
        p.pObjs[i].Value = ~0;
    }
}

/**Function*************************************************************

  Synopsis    [Sets the phase of one object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjSetPhase(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if Gia_ObjIsAnd(pObj) {
        var fPhase0 int32  = Gia_ObjPhase(Gia_ObjFanin0(pObj)) ^ Gia_ObjFaninC0(pObj);
        var fPhase1 int32  = Gia_ObjPhase(Gia_ObjFanin1(pObj)) ^ Gia_ObjFaninC1(pObj);
        if Gia_ObjIsMux(p, pObj) {
            var fPhase2 int32  = Gia_ObjPhase(Gia_ObjFanin2(p, pObj)) ^ Gia_ObjFaninC2(p, pObj);
            pObj.fPhase = (fPhase2 && fPhase1) || (!fPhase2 && fPhase0);
        
        } else if Gia_ObjIsXor(pObj) {
            pObj.fPhase = fPhase0 ^ fPhase1;
        } else {
            pObj.fPhase = fPhase0 & fPhase1; }
    
    } else if Gia_ObjIsCo(pObj) {
        pObj.fPhase = (Gia_ObjPhase(Gia_ObjFanin0(pObj)) ^ Gia_ObjFaninC0(pObj));
    } else {
        pObj.fPhase = 0; }
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSetPhase(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        Gia_ObjSetPhase( p, pObj );
}
func Gia_ManSetPhasePattern(p *Gia_Man_t, vCiValues *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManCiNum(p) == Vec_IntSize(vCiValues), "Gia_ManCiNum(p) == Vec_IntSize(vCiValues)");
    Gia_ManForEachObj( p, pObj, i )
        if ( Gia_ObjIsCi(pObj) )
            pObj.fPhase = Vec_IntEntry( vCiValues, Gia_ObjCioId(pObj) );
        else
            Gia_ObjSetPhase( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSetPhase1(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachCi( p, pObj, i )
        pObj.fPhase = 1;
    Gia_ManForEachObj( p, pObj, i )
        if ( !Gia_ObjIsCi(pObj) )
            Gia_ObjSetPhase( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Sets phases of the internal nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanPhase(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        pObj.fPhase = 0;
}

/**Function*************************************************************

  Synopsis    [Returns the number of COs whose value is 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckCoPhase(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    Gia_ManForEachCo( p, pObj, i )
        Counter += pObj.fPhase;
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Prepares copies for the model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanLevels(p *Gia_Man_t, Size int32) {
    if p.vLevels == nil {
        p.vLevels = Vec_IntAlloc( Size ); }
    Vec_IntFill( p.vLevels, Size, 0 );
}
/**Function*************************************************************

  Synopsis    [Prepares copies for the model.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanTruth(p *Gia_Man_t) {
    if p.vTruths == nil {
        p.vTruths = Vec_IntAlloc( Gia_ManObjNum(p) ); }
    Vec_IntFill( p.vTruths, Gia_ManObjNum(p), -1 );
}

/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManLevelNum(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManCleanLevels( p, Gia_ManObjNum(p) );
    p.nLevels = 0;
    Gia_ManForEachObj( p, pObj, i )
    {
        if !p.fGiaSimple && Gia_ObjIsBuf(pObj) {
            Gia_ObjSetBufLevel( p, pObj );
        } else if Gia_ObjIsAnd(pObj) {
            Gia_ObjSetGateLevel( p, pObj );
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjSetCoLevel( p, pObj );
        } else {
            Gia_ObjSetLevel( p, pObj, 0 ); }
        p.nLevels = Abc_MaxInt( p.nLevels, Gia_ObjLevel(p, pObj) );
    }
    return p.nLevels;
}
func Gia_ManLevelRNum(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManCleanLevels( p, Gia_ManObjNum(p) );
    p.nLevels = 0;
    Gia_ManForEachObjReverse( p, pObj, i )
    {
        if !p.fGiaSimple && Gia_ObjIsBuf(pObj) {
            Gia_ObjUpdateLevelId( p, Gia_ObjFaninId0(pObj, i), Gia_ObjLevel(p, pObj) );
        } else if Gia_ObjIsAnd(pObj) {
            Gia_ObjUpdateLevelId( p, Gia_ObjFaninId0(pObj, i), 1+Gia_ObjLevel(p, pObj) );
            Gia_ObjUpdateLevelId( p, Gia_ObjFaninId1(pObj, i), 1+Gia_ObjLevel(p, pObj) );
        
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjUpdateLevelId( p, Gia_ObjFaninId0(pObj, i), 1 );
        } else {
            p.nLevels = Abc_MaxInt( p.nLevels, Gia_ObjLevel(p, pObj) ); }
    }
    return p.nLevels;
}
func Gia_ManLevelAve(p *Gia_Man_t) float32 {
    var pObj *Gia_Obj_t;
    var i, Ave int32;
    assert(p.vLevels, "p.vLevels");
    Gia_ManForEachCo( p, pObj, i )
        Ave += Gia_ObjLevel(p, pObj);
    return float32(Ave) / Gia_ManCoNum(p);
}

/**Function*************************************************************

  Synopsis    [Assigns levels using CI level information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGetCiLevels(p *Gia_Man_t) *Vec_Int_t {
    var vCiLevels *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    if p.vLevels == nil {
        return nil; }
    vCiLevels = Vec_IntAlloc( Gia_ManCiNum(p) );
    Gia_ManForEachCi( p, pObj, i )
        Vec_IntPush( vCiLevels, Gia_ObjLevel(p, pObj) );
    return vCiLevels;
}
func Gia_ManSetLevels(p *Gia_Man_t, vCiLevels *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t;
    var i int32;
    if vCiLevels == nil {
        return Gia_ManLevelNum( p ); }
    assert(Vec_IntSize(vCiLevels) == Gia_ManCiNum(p), "Vec_IntSize(vCiLevels) == Gia_ManCiNum(p)");
    Gia_ManCleanLevels( p, Gia_ManObjNum(p) );
    p.nLevels = 0;
    Gia_ManForEachCi( p, pObj, i )
    {
        Gia_ObjSetLevel( p, pObj, Vec_IntEntry(vCiLevels, i) );
        p.nLevels = Abc_MaxInt( p.nLevels, Gia_ObjLevel(p, pObj) );
    }
    Gia_ManForEachObj( p, pObj, i )
    {
        if Gia_ObjIsAnd(pObj) {
            Gia_ObjSetGateLevel( p, pObj );
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjSetCoLevel( p, pObj );
        } else { continue; }
        p.nLevels = Abc_MaxInt( p.nLevels, Gia_ObjLevel(p, pObj) );
    }
    return p.nLevels;
}

/**Function*************************************************************

  Synopsis    [Compute reverse levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManReverseLevel(p *Gia_Man_t) *Vec_Int_t {
    var vLevelRev *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    vLevelRev = Vec_IntStart( Gia_ManObjNum(p) );
    Gia_ManForEachAndReverse( p, pObj, i )
    {
        var LevelR int32  = Vec_IntEntry( vLevelRev, i );
        if Gia_ObjIsMux(p, pObj) {
            Vec_IntUpdateEntry( vLevelRev, Gia_ObjFaninId0(pObj, i), LevelR + 2 );
            Vec_IntUpdateEntry( vLevelRev, Gia_ObjFaninId1(pObj, i), LevelR + 2 );
            Vec_IntUpdateEntry( vLevelRev, Gia_ObjFaninId2(p, i), LevelR + 2 );
        
        } else if Gia_ObjIsXor(pObj) {
            Vec_IntUpdateEntry( vLevelRev, Gia_ObjFaninId0(pObj, i), LevelR + 2 );
            Vec_IntUpdateEntry( vLevelRev, Gia_ObjFaninId1(pObj, i), LevelR + 2 );
        
        } else if Gia_ObjIsBuf(pObj) {
            Vec_IntUpdateEntry( vLevelRev, Gia_ObjFaninId0(pObj, i), LevelR );
        
        } else {
            Vec_IntUpdateEntry( vLevelRev, Gia_ObjFaninId0(pObj, i), LevelR + 1 );
            Vec_IntUpdateEntry( vLevelRev, Gia_ObjFaninId1(pObj, i), LevelR + 1 );
        }
    }
    return vLevelRev;
}

/**Function*************************************************************

  Synopsis    [Compute required levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRequiredLevel(p *Gia_Man_t) *Vec_Int_t {
    var vRequired *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, LevelMax int32;
    vRequired = Gia_ManReverseLevel( p );
    Gia_ManForEachCi( p, pObj, i )
        LevelMax = Abc_MaxInt( LevelMax, Vec_IntEntry(vRequired, Gia_ObjId(p, pObj)) );
    Gia_ManForEachObj( p, pObj, i )
        Vec_IntWriteEntry( vRequired, i, LevelMax - Vec_IntEntry(vRequired, i) );
    return vRequired;
}

/**Function*************************************************************

  Synopsis    [Compute slacks measured using the number of AIG levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManComputeSlacks(p *Gia_Man_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var i int32;
    var nLevels int32 = Gia_ManLevelNum( p );
    var vLevelR *Vec_Int_t  = Gia_ManReverseLevel( p );
    var vSlacks *Vec_Int_t  = Vec_IntAlloc( Gia_ManObjNum(p) );
    Gia_ManForEachObj( p, pObj, i )
        Vec_IntPush( vSlacks, nLevels - Gia_ObjLevelId(p, i) - Vec_IntEntry(vLevelR, i) );
    assert(Vec_IntSize(vSlacks) == Gia_ManObjNum(p), "Vec_IntSize(vSlacks) == Gia_ManObjNum(p)");
    Vec_IntFree( vLevelR );
    return vSlacks;
}

/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCreateValueRefs(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
    {
        pObj.Value = 0;
        if Gia_ObjIsAnd(pObj) {
            Gia_ObjFanin0(pObj).Value++;
            if !Gia_ObjIsBuf(pObj) {
                Gia_ObjFanin1(pObj).Value++; }
        
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjFanin0(pObj).Value++; }
    }
}

/**Function*************************************************************

  Synopsis    [Assigns references.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Assigns references.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCreateMuxRefs(p *Gia_Man_t) *int32 {
    var pObj, pCtrl, pFan0, pFan1 *Gia_Obj_t;
    var i int32;
    var pMuxRefs *int32;
    pMuxRefs = ABC_CALLOC( int, Gia_ManObjNum(p) );
    Gia_ManForEachObj( p, pObj, i )
    {
        if Gia_ObjRecognizeExor( pObj, &pFan0, &pFan1 ) != 0 {
            continue; }
        if !(Gia_ObjIsMuxType(pObj) != 0) {
            continue; }
        pCtrl = Gia_ObjRecognizeMux( pObj, &pFan0, &pFan1 );
        pMuxRefs[ Gia_ObjId(p, Gia_Regular(pCtrl)) ]++;
    }
    return pMuxRefs;
}

/**Function*************************************************************

  Synopsis    [Computes the maximum frontier size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManBfsForCrossCut(p *Gia_Man_t) *Vec_Int_t {
    var vNodes *Vec_Int_t  = Vec_IntAlloc( Gia_ManObjNum(p) );
    var vLevels *Vec_Vec_t  = Gia_ManLevelize( p );
    var vObjs *Vec_Ptr_t;
    var pObj *Gia_Obj_t;
    var i, k int32;
    Vec_VecForEachLevel( vLevels, vObjs, i )
        Vec_PtrForEachEntry( Gia_Obj_t *, vObjs, pObj, k )
            Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
    Vec_VecFree( vLevels );
    return vNodes;
}

func Gia_ManDfsForCrossCut_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vNodes *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
        return;
    }
    if Gia_ObjIsCo(pObj) {
        Gia_ObjFanin0(pObj).Value++;
        Gia_ManDfsForCrossCut_rec( p, Gia_ObjFanin0(pObj), vNodes );
        Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ObjFanin0(pObj).Value++;
    Gia_ObjFanin1(pObj).Value++;
    Gia_ManDfsForCrossCut_rec( p, Gia_ObjFanin0(pObj), vNodes );
    Gia_ManDfsForCrossCut_rec( p, Gia_ObjFanin1(pObj), vNodes );
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
}
func Gia_ManDfsForCrossCut(p *Gia_Man_t, fReverse int32) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManCleanValue( p );
    vNodes = Vec_IntAlloc( Gia_ManObjNum(p) );
    Gia_ManIncrementTravId( p );
    if fReverse != 0 {
        Gia_ManForEachCoReverse( p, pObj, i )
            if ( !Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) )
                Gia_ManDfsForCrossCut_rec( p, pObj, vNodes );
    
    } else {
        Gia_ManForEachCo( p, pObj, i )
            if ( !Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) )
                Gia_ManDfsForCrossCut_rec( p, pObj, vNodes );
    }
    return vNodes;
}
func Gia_ManCrossCut(p *Gia_Man_t, fReverse int32) int32 {
    var vNodes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, nCutCur, nCutMax int32;
    vNodes = Gia_ManDfsForCrossCut( p, fReverse );
    //vNodes = Gia_ManBfsForCrossCut( p );
    Gia_ManForEachObjVec( vNodes, p, pObj, i )
    {
        if pObj.Value {
            nCutCur++; }
        if nCutMax < nCutCur {
            nCutMax = nCutCur; }
        if Gia_ObjIsAnd(pObj) {
            if --Gia_ObjFanin0(pObj).Value == 0 {
                nCutCur--; }
            if --Gia_ObjFanin1(pObj).Value == 0 {
                nCutCur--; }
        
        } else if Gia_ObjIsCo(pObj) {
            if --Gia_ObjFanin0(pObj).Value == 0 {
                nCutCur--; }
        }
    }
    Vec_IntFree( vNodes );
    Gia_ManForEachObj( p, pObj, i )
        assert(pObj.Value == 0, "pObj.Value == 0");
    return nCutMax;
}


/**Function*************************************************************

  Synopsis    [Collects PO Ids into one array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectPoIds(p *Gia_Man_t) *Vec_Int_t {
    var vStart *Vec_Int_t;
    var Entry, i int32;
    vStart = Vec_IntAlloc( Gia_ManPoNum(p) );
    Vec_IntForEachEntryStop( p.vCos, Entry, i, Gia_ManPoNum(p) )
        Vec_IntPush( vStart, Entry );
    return vStart;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjIsMuxType(pNode *Gia_Obj_t) int32 {
    var pNode0, pNode1 *Gia_Obj_t;
    // check that the node is regular
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    // if the node is not AND, this is not MUX
    if !Gia_ObjIsAnd(pNode) || Gia_ObjIsBuf(pNode) {
        return 0; }
    // if the children are not complemented, this is not MUX
    if !Gia_ObjFaninC0(pNode) || !Gia_ObjFaninC1(pNode) {
        return 0; }
    // get children
    pNode0 = Gia_ObjFanin0(pNode);
    pNode1 = Gia_ObjFanin1(pNode);
    // if the children are not ANDs, this is not MUX
    if !Gia_ObjIsAnd(pNode0) || !Gia_ObjIsAnd(pNode1) {
        return 0; }
    // otherwise the node is MUX iff it has a pair of equal grandchildren
    return (Gia_ObjFanin0(pNode0) == Gia_ObjFanin0(pNode1) && (Gia_ObjFaninC0(pNode0) ^ Gia_ObjFaninC0(pNode1))) || 
           (Gia_ObjFanin0(pNode0) == Gia_ObjFanin1(pNode1) && (Gia_ObjFaninC0(pNode0) ^ Gia_ObjFaninC1(pNode1))) ||
           (Gia_ObjFanin1(pNode0) == Gia_ObjFanin0(pNode1) && (Gia_ObjFaninC1(pNode0) ^ Gia_ObjFaninC0(pNode1))) ||
           (Gia_ObjFanin1(pNode0) == Gia_ObjFanin1(pNode1) && (Gia_ObjFaninC1(pNode0) ^ Gia_ObjFaninC1(pNode1)));
}


/**Function*************************************************************

  Synopsis    [Recognizes what nodes are inputs of the EXOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjRecognizeExor(pObj *Gia_Obj_t, ppFan0 **Gia_Obj_t, ppFan1 **Gia_Obj_t) int32 {
    var p0, p1 *Gia_Obj_t;
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    if !Gia_ObjIsAnd(pObj) || Gia_ObjIsBuf(pObj) {
        return 0; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    p0 = Gia_ObjChild0(pObj);
    p1 = Gia_ObjChild1(pObj);
    if !Gia_IsComplement(p0) || !Gia_IsComplement(p1) {
        return 0; }
    p0 = Gia_Regular(p0);
    p1 = Gia_Regular(p1);
    if !Gia_ObjIsAnd(p0) || !Gia_ObjIsAnd(p1) {
        return 0; }
    if Gia_ObjFanin0(p0) != Gia_ObjFanin0(p1) || Gia_ObjFanin1(p0) != Gia_ObjFanin1(p1) {
        return 0; }
    if Gia_ObjFaninC0(p0) == Gia_ObjFaninC0(p1) || Gia_ObjFaninC1(p0) == Gia_ObjFaninC1(p1) {
        return 0; }
    if ppFan0 { *ppFan0 = Gia_ObjChild0(p0); }
    if ppFan1 { *ppFan1 = Gia_ObjChild1(p0); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Recognizes what nodes are control and data inputs of a MUX.]

  Description [If the node is a MUX, returns the control variable C.
  Assigns nodes T and E to be the then and else variables of the MUX. 
  Node C is never complemented. Nodes T and E can be complemented.
  This function also recognizes EXOR/NEXOR gates as MUXes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjRecognizeMux(pNode *Gia_Obj_t, ppNodeT **Gia_Obj_t, ppNodeE **Gia_Obj_t) *Gia_Obj_t {
    var pNode0, pNode1 *Gia_Obj_t;
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    assert((Gia_ObjIsMuxType(pNode) != 0), "Gia_ObjIsMuxType(pNode)");
    // get children
    pNode0 = Gia_ObjFanin0(pNode);
    pNode1 = Gia_ObjFanin1(pNode);

    // find the control variable
    if Gia_ObjFanin1(pNode0) == Gia_ObjFanin1(pNode1) && (Gia_ObjFaninC1(pNode0) ^ Gia_ObjFaninC1(pNode1)) {
//        if ( FrGia_IsComplement(pNode1->p2) )
        if Gia_ObjFaninC1(pNode0) { // pNode2->p2 is positive phase of C
            *ppNodeT = Gia_Not(Gia_ObjChild0(pNode1));//pNode2->p1);
            *ppNodeE = Gia_Not(Gia_ObjChild0(pNode0));//pNode1->p1);
            return Gia_ObjChild1(pNode1);//pNode2->p2;
        
        } else { // pNode1->p2 is positive phase of C
            *ppNodeT = Gia_Not(Gia_ObjChild0(pNode0));//pNode1->p1);
            *ppNodeE = Gia_Not(Gia_ObjChild0(pNode1));//pNode2->p1);
            return Gia_ObjChild1(pNode0);//pNode1->p2;
        }
    
    } else if Gia_ObjFanin0(pNode0) == Gia_ObjFanin0(pNode1) && (Gia_ObjFaninC0(pNode0) ^ Gia_ObjFaninC0(pNode1)) {
//        if ( FrGia_IsComplement(pNode1->p1) )
        if Gia_ObjFaninC0(pNode0) { // pNode2->p1 is positive phase of C
            *ppNodeT = Gia_Not(Gia_ObjChild1(pNode1));//pNode2->p2);
            *ppNodeE = Gia_Not(Gia_ObjChild1(pNode0));//pNode1->p2);
            return Gia_ObjChild0(pNode1);//pNode2->p1;
        
        } else { // pNode1->p1 is positive phase of C
            *ppNodeT = Gia_Not(Gia_ObjChild1(pNode0));//pNode1->p2);
            *ppNodeE = Gia_Not(Gia_ObjChild1(pNode1));//pNode2->p2);
            return Gia_ObjChild0(pNode0);//pNode1->p1;
        }
    
    } else if Gia_ObjFanin0(pNode0) == Gia_ObjFanin1(pNode1) && (Gia_ObjFaninC0(pNode0) ^ Gia_ObjFaninC1(pNode1)) {
//        if ( FrGia_IsComplement(pNode1->p1) )
        if Gia_ObjFaninC0(pNode0) { // pNode2->p2 is positive phase of C
            *ppNodeT = Gia_Not(Gia_ObjChild0(pNode1));//pNode2->p1);
            *ppNodeE = Gia_Not(Gia_ObjChild1(pNode0));//pNode1->p2);
            return Gia_ObjChild1(pNode1);//pNode2->p2;
        
        } else { // pNode1->p1 is positive phase of C
            *ppNodeT = Gia_Not(Gia_ObjChild1(pNode0));//pNode1->p2);
            *ppNodeE = Gia_Not(Gia_ObjChild0(pNode1));//pNode2->p1);
            return Gia_ObjChild0(pNode0);//pNode1->p1;
        }
    
    } else if Gia_ObjFanin1(pNode0) == Gia_ObjFanin0(pNode1) && (Gia_ObjFaninC1(pNode0) ^ Gia_ObjFaninC0(pNode1)) {
//        if ( FrGia_IsComplement(pNode1->p2) )
        if Gia_ObjFaninC1(pNode0) { // pNode2->p1 is positive phase of C
            *ppNodeT = Gia_Not(Gia_ObjChild1(pNode1));//pNode2->p2);
            *ppNodeE = Gia_Not(Gia_ObjChild0(pNode0));//pNode1->p1);
            return Gia_ObjChild0(pNode1);//pNode2->p1;
        
        } else { // pNode1->p2 is positive phase of C
            *ppNodeT = Gia_Not(Gia_ObjChild0(pNode0));//pNode1->p1);
            *ppNodeE = Gia_Not(Gia_ObjChild1(pNode1));//pNode2->p2);
            return Gia_ObjChild1(pNode0);//pNode1->p2;
        }
    }
    assert(false, "0"); // this is not MUX
    return nil;
}
func Gia_ObjRecognizeMuxLits(p *Gia_Man_t, pNode *Gia_Obj_t, iLitT *int32, iLitE *int32) int32 {
    var pNodeT, pNodeE *Gia_Obj_t;
    var pCtrl *Gia_Obj_t  = Gia_ObjRecognizeMux( pNode, &pNodeT, &pNodeE );
    assert(pCtrl != nil, "pCtrl != nil");
    *iLitT = Gia_Obj2Lit( p, pNodeT );
    *iLitE = Gia_Obj2Lit( p, pNodeE );
    return Gia_Obj2Lit( p, pCtrl );
}


/**Function*************************************************************

  Synopsis    [Dereferences the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_NodeDeref_rec(p *Gia_Man_t, pNode *Gia_Obj_t) int32 {
    var pFanin *Gia_Obj_t;
    var Counter int32;
    if Gia_ObjIsCi(pNode) {
        return 0; }
    assert(Gia_ObjIsAnd(pNode), "Gia_ObjIsAnd(pNode)");
    pFanin = Gia_ObjFanin0(pNode);
    assert(Gia_ObjRefNum(p, pFanin) > 0, "Gia_ObjRefNum(p, pFanin) > 0");
    if Gia_ObjRefDec(p, pFanin) == 0 {
        Counter += Gia_NodeDeref_rec( p, pFanin ); }
    pFanin = Gia_ObjFanin1(pNode);
    assert(Gia_ObjRefNum(p, pFanin) > 0, "Gia_ObjRefNum(p, pFanin) > 0");
    if Gia_ObjRefDec(p, pFanin) == 0 {
        Counter += Gia_NodeDeref_rec( p, pFanin ); }
    return Counter + 1;
}

/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_NodeRef_rec(p *Gia_Man_t, pNode *Gia_Obj_t, fMark int32) int32 {
    var pFanin *Gia_Obj_t;
    var Counter int32;
    if Gia_ObjIsCi(pNode) {
        return 0; }
    assert(Gia_ObjIsAnd(pNode), "Gia_ObjIsAnd(pNode)");
    if fMark != 0 { Gia_ObjSetTravIdCurrent(p, pNode); }
    pFanin = Gia_ObjFanin0(pNode);
    if Gia_ObjRefInc(p, pFanin) == 0 {
        Counter += Gia_NodeRef_rec( p, pFanin, fMark ); }
    pFanin = Gia_ObjFanin1(pNode);
    if Gia_ObjRefInc(p, pFanin) == 0 {
        Counter += Gia_NodeRef_rec( p, pFanin, fMark ); }
    return Counter + 1;
}


/**Function*************************************************************

  Synopsis    [References the node's MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPoMffcSize(p *Gia_Man_t) int32 {
    Gia_ManCreateRefs( p );
    return Gia_NodeDeref_rec( p, Gia_ObjFanin0(Gia_ManPo(p, 0)) );
}

/**Function*************************************************************

  Synopsis    [Returns the number of internal nodes in the MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_NodeMffcSize(p *Gia_Man_t, pNode *Gia_Obj_t) int32 {
    var ConeSize1, ConeSize2 int32;
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    assert(Gia_ObjIsCand(pNode), "Gia_ObjIsCand(pNode)");
    ConeSize1 = Gia_NodeDeref_rec( p, pNode );
    ConeSize2 = Gia_NodeRef_rec( p, pNode, 0 );
    assert(ConeSize1 == ConeSize2, "ConeSize1 == ConeSize2");
    assert(ConeSize1 >= 0, "ConeSize1 >= 0");
    return ConeSize1;
}
func Gia_NodeMffcSizeMark(p *Gia_Man_t, pNode *Gia_Obj_t) int32 {
    var ConeSize1, ConeSize2 int32;
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    assert(Gia_ObjIsCand(pNode), "Gia_ObjIsCand(pNode)");
    ConeSize1 = Gia_NodeDeref_rec( p, pNode );
    Gia_ManIncrementTravId( p );
    ConeSize2 = Gia_NodeRef_rec( p, pNode, 1 );
    assert(ConeSize1 == ConeSize2, "ConeSize1 == ConeSize2");
    assert(ConeSize1 >= 0, "ConeSize1 >= 0");
    return ConeSize1;
}

/**Function*************************************************************

  Synopsis    [Returns the number of internal nodes in the MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_NodeCollect_rec(p *Gia_Man_t, pNode *Gia_Obj_t, vSupp *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pNode) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pNode);
    if Gia_ObjRefNum(p, pNode) || Gia_ObjIsCi(pNode) {
        Vec_IntPush( vSupp, Gia_ObjId(p, pNode) );
        return;
    }
    assert(Gia_ObjIsAnd(pNode), "Gia_ObjIsAnd(pNode)");
    Gia_NodeCollect_rec( p, Gia_ObjFanin0(pNode), vSupp );
    Gia_NodeCollect_rec( p, Gia_ObjFanin1(pNode), vSupp );
}
func Gia_NodeMffcSizeSupp(p *Gia_Man_t, pNode *Gia_Obj_t, vSupp *Vec_Int_t) int32 {
    var ConeSize1, ConeSize2 int32;
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    assert(Gia_ObjIsAnd(pNode), "Gia_ObjIsAnd(pNode)");
    Vec_IntClear( vSupp );
    Gia_ManIncrementTravId( p );
    ConeSize1 = Gia_NodeDeref_rec( p, pNode );
    Gia_NodeCollect_rec( p, Gia_ObjFanin0(pNode), vSupp );
    Gia_NodeCollect_rec( p, Gia_ObjFanin1(pNode), vSupp );
    ConeSize2 = Gia_NodeRef_rec( p, pNode, 0 );
    assert(ConeSize1 == ConeSize2, "ConeSize1 == ConeSize2");
    assert(ConeSize1 >= 0, "ConeSize1 >= 0");
    return ConeSize1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_NodeMffcMapping_rec(p *Gia_Man_t, iObj int32, vMapping *Vec_Int_t, vSupp *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t; var i int32;
var iNode int32;
var Count int32 = 1;
    if !(iObj != 0) || Vec_IntEntry(vMapping, iObj) {
        return 0; }
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return 0; }
    Gia_NodeMffcSizeSupp( p, pObj, vSupp );
    Vec_IntSort( vSupp, 0 );
    Vec_IntWriteEntry( vMapping, iObj, Vec_IntSize(vMapping) );
    Vec_IntPush( vMapping, Vec_IntSize(vSupp) );
    Vec_IntAppend( vMapping, vSupp );
    Vec_IntPush( vMapping, iObj );
    Vec_IntForEachEntry( vSupp, iNode, i )
        Count += Gia_NodeMffcMapping_rec( p, iNode, vMapping, vSupp );
    return Count;
}
func Gia_NodeMffcMapping(p *Gia_Man_t) int32 {
    var i, Id, Count int32;
    var pRefsOld *int32;
    var vMapping *Vec_Int_t;
    var vSupp *Vec_Int_t = Vec_IntAlloc( 100 );
    vMapping = Vec_IntAlloc( 2 * Gia_ManObjNum(p) );
    Vec_IntFill( vMapping, Gia_ManObjNum(p), 0 );
    pRefsOld = p.Refs; p.Refs = nil;
    Gia_ManCreateRefs( p ); 
    Gia_ManForEachCoDriverId( p, Id, i )
        Count += Gia_NodeMffcMapping_rec( p, Id, vMapping, vSupp );
    p.Refs = pRefsOld;
    Vec_IntFree( vSupp );
    p.vMapping = vMapping;
    //printf( "Mapping is %.2fx larger than AIG manager.\n", 1.0*Vec_IntSize(vMapping)/Gia_ManObjNum(p) );
    return Count;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManHasDangling(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    Gia_ManForEachObj( p, pObj, i )
    {
        pObj.fMark0 = 0;
        if Gia_ObjIsCo(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
        } else if Gia_ObjIsMux(p, pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
            Gia_ObjFanin1(pObj).fMark0 = 1;
            Gia_ObjFanin2(p, pObj).fMark0 = 1;
        
        } else if Gia_ObjIsAnd(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
            Gia_ObjFanin1(pObj).fMark0 = 1;
        }
    }
    Gia_ManForEachAnd( p, pObj, i )
        Counter += !pObj.fMark0;
    Gia_ManCleanMark0( p );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMarkDangling(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    Gia_ManForEachObj( p, pObj, i )
    {
        pObj.fMark0 = 0;
        if Gia_ObjIsAnd(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
            Gia_ObjFanin1(pObj).fMark0 = 1;
        
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1; }
    }
    Gia_ManForEachAnd( p, pObj, i )
        Counter += !pObj.fMark0;
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGetDangling(p *Gia_Man_t) *Vec_Int_t {
    var vDangles *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
    {
        pObj.fMark0 = 0;
        if Gia_ObjIsAnd(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
            Gia_ObjFanin1(pObj).fMark0 = 1;
        
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1; }
    }
    vDangles = Vec_IntAlloc( 100 );
    Gia_ManForEachAnd( p, pObj, i )
        if ( !pObj.fMark0 )
            Vec_IntPush( vDangles, i );
    Gia_ManCleanMark0( p );
    return vDangles;
}
/**Function*************************************************************

  Synopsis    [Verbose printing of the AIG node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjPrint(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if pObj == nil {
        fmt.Printf("Object is NULL.");
        return;
    }
    if Gia_IsComplement(pObj) {
        fmt.Printf("Compl ");
        pObj = Gia_Not(pObj);
    }
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    fmt.Printf("Obj %4d : ", Gia_ObjId(p, pObj));
    if Gia_ObjIsConst0(pObj) {
        fmt.Printf("constant 0");
    } else if Gia_ObjIsPi(p, pObj) {
        fmt.Printf("PI");
    } else if Gia_ObjIsPo(p, pObj) {
        fmt.Printf("PO( %4d%s )", Gia_ObjFaninId0p(p, pObj), (Gia_ObjFaninC0(pObj)? "\'" : " "));
    } else if Gia_ObjIsCi(pObj) {
        fmt.Printf("RO( %4d%s )", Gia_ObjFaninId0p(p, Gia_ObjRoToRi(p, pObj)), (Gia_ObjFaninC0(Gia_ObjRoToRi(p, pObj))? "\'" : " "));
    } else if Gia_ObjIsCo(pObj) {
        fmt.Printf("RI( %4d%s )", Gia_ObjFaninId0p(p, pObj), (Gia_ObjFaninC0(pObj)? "\'" : " "));
//    else if ( Gia_ObjIsBuf(pObj) )
//        printf( "BUF( %d%s )", Gia_ObjFaninId0p(p, pObj), (Gia_ObjFaninC0(pObj)? "\'" : " ") );
    } else if Gia_ObjIsXor(pObj) {
        fmt.Printf("XOR( %4d%s, %4d%s )", Gia_ObjFaninId0p(p, pObj), (Gia_ObjFaninC0(pObj)? "\'" : " "), Gia_ObjFaninId1p(p, pObj), (Gia_ObjFaninC1(pObj)? "\'" : " "));
    } else if Gia_ObjIsMuxId(p, Gia_ObjId(p, pObj)) {
        fmt.Printf("MUX( %4d%s, %4d%s, %4d%s )", Gia_ObjFaninId2p(p, pObj), (Gia_ObjFaninC2(p, pObj)? "\'" : " "), Gia_ObjFaninId1p(p, pObj), (Gia_ObjFaninC1(pObj)? "\'" : " "), Gia_ObjFaninId0p(p, pObj), (Gia_ObjFaninC0(pObj)? "\'" : " "));
    } else {
        fmt.Printf("AND( %4d%s, %4d%s )", Gia_ObjFaninId0p(p, pObj), (Gia_ObjFaninC0(pObj)? "\'" : " "), Gia_ObjFaninId1p(p, pObj), (Gia_ObjFaninC1(pObj)? "\'" : " ")); }
    if p.Refs {
        fmt.Printf(" (refs = %3d)", Gia_ObjRefNum(p, pObj)); }
    if pObj.fMark0 {
        fmt.Printf(" mark0"); }
    if pObj.fMark1 {
        fmt.Printf(" mark1"); }
    if Gia_ManHasMapping(p) && Gia_ObjIsLut(p, Gia_ObjId(p, pObj)) {
        var i, iFan int32;
        fmt.Printf(" Cut = { ");
        Gia_LutForEachFanin( p, Gia_ObjId(p, pObj), iFan, i )
            fmt.Printf("%d ", iFan);
        fmt.Printf("}");
    }
    if Gia_ManHasMapping2(p) && Gia_ObjIsLut2(p, Gia_ObjId(p, pObj)) {
        var i, iFan int32;
        fmt.Printf(" Cut = { ");
        Gia_LutForEachFanin2( p, Gia_ObjId(p, pObj), iFan, i )
            fmt.Printf("%d ", iFan);
        fmt.Printf("}");
    }
    fmt.Printf("\n");
/*
    if ( p->pRefs )
    {
        Gia_Obj_t * pFanout;
        int i;
        int iFan = -1; // Suppress "might be used uninitialized"
        printf( "\nFanouts:\n" );
        Gia_ObjForEachFanout( p, pObj, pFanout, iFan, i )
        {
            printf( "    " );
            printf( "Node %4d : ", Gia_ObjId(pFanout) );
            if ( Gia_ObjIsPo(pFanout) )
                printf( "PO( %4d%s )", Gia_ObjFanin0(pFanout)->Id, (Gia_ObjFaninC0(pFanout)? "\'" : " ") );
            else if ( Gia_ObjIsBuf(pFanout) )
                printf( "BUF( %d%s )", Gia_ObjFanin0(pFanout)->Id, (Gia_ObjFaninC0(pFanout)? "\'" : " ") );
            else
                printf( "AND( %4d%s, %4d%s )", 
                    Gia_ObjFanin0(pFanout)->Id, (Gia_ObjFaninC0(pFanout)? "\'" : " "), 
                    Gia_ObjFanin1(pFanout)->Id, (Gia_ObjFaninC1(pFanout)? "\'" : " ") );
            printf( "\n" );
        }
        return;
    }
*/
}
func Gia_ManPrint(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    fmt.Printf("GIA manager has %d ANDs, %d XORs, %d MUXes.\n", Gia_ManAndNum(p) - Gia_ManXorNum(p) - Gia_ManMuxNum(p), Gia_ManXorNum(p), Gia_ManMuxNum(p)); 
    Gia_ManForEachObj( p, pObj, i )
        Gia_ObjPrint( p, pObj );
}
func Gia_ManPrintCo_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if Gia_ObjIsAnd(pObj) {
        Gia_ManPrintCo_rec( p, Gia_ObjFanin0(pObj) );
        Gia_ManPrintCo_rec( p, Gia_ObjFanin1(pObj) );
        if Gia_ObjIsMux(p, pObj) {
            Gia_ManPrintCo_rec( p, Gia_ObjFanin2(p, pObj) ); }
    }
    Gia_ObjPrint( p, pObj );
}
func Gia_ManPrintCo(p *Gia_Man_t, pObj *Gia_Obj_t) {
    assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
    fmt.Printf("TFI cone of CO number %d:\n", Gia_ObjCioId(pObj));
    Gia_ManPrintCo_rec( p, Gia_ObjFanin0(pObj) );
    Gia_ObjPrint( p, pObj );
}

func Gia_ManPrintCollect_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vNodes *Vec_Int_t) {
    if Vec_IntFind(vNodes, Gia_ObjId(p, pObj)) >= 0 {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManPrintCollect_rec( p, Gia_ObjFanin0(pObj), vNodes );
    Gia_ManPrintCollect_rec( p, Gia_ObjFanin1(pObj), vNodes );
    if Gia_ObjIsMux(p, pObj) {
        Gia_ManPrintCollect_rec( p, Gia_ObjFanin2(p, pObj), vNodes ); }
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
}
func Gia_ManPrintCone(p *Gia_Man_t, pObj *Gia_Obj_t, pLeaves *int32, nLeaves int32, vNodes *Vec_Int_t) {
    var i int32;
    Vec_IntClear( vNodes );
    for  i = 0; i < nLeaves; i++  {
        Vec_IntPush( vNodes, pLeaves[i] );
    }
    Gia_ManPrintCollect_rec( p, pObj, vNodes );
    fmt.Printf("GIA logic cone for node %d:\n", Gia_ObjId(p, pObj));
    Gia_ManForEachObjVec( vNodes, p, pObj, i )
        Gia_ObjPrint( p, pObj );
}
func Gia_ManPrintConeMulti(p *Gia_Man_t, vObjs *Vec_Int_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Vec_IntClear( vNodes );
    Vec_IntAppend( vNodes, vLeaves );
    Gia_ManForEachObjVec( vObjs, p, pObj, i )
        Gia_ManPrintCollect_rec( p, pObj, vNodes );
    fmt.Printf("GIA logic cone for %d nodes:\n", Vec_IntSize(vObjs));
    Gia_ManForEachObjVec( vNodes, p, pObj, i )
        Gia_ObjPrint( p, pObj );
}

func Gia_ManPrintCollect2_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vNodes *Vec_Int_t) {
    if Vec_IntFind(vNodes, Gia_ObjId(p, pObj)) >= 0 {
        return; }
    if Gia_ObjIsCo(pObj) || Gia_ObjIsAnd(pObj) {
        Gia_ManPrintCollect2_rec( p, Gia_ObjFanin0(pObj), vNodes ); }
    if Gia_ObjIsAnd(pObj) {
        Gia_ManPrintCollect2_rec( p, Gia_ObjFanin1(pObj), vNodes ); }
    if Gia_ObjIsMux(p, pObj) {
        Gia_ManPrintCollect2_rec( p, Gia_ObjFanin2(p, pObj), vNodes ); }
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
}
func Gia_ManPrintCone2(p *Gia_Man_t, pObj *Gia_Obj_t) {
    var vNodes *Vec_Int_t;
    var i int32;
    vNodes = Vec_IntAlloc( 100 );
    Gia_ManPrintCollect2_rec( p, pObj, vNodes );
    fmt.Printf("GIA logic cone for node %d:\n", Gia_ObjId(p, pObj));
    Gia_ManForEachObjVec( vNodes, p, pObj, i )
        Gia_ObjPrint( p, pObj );
    Vec_IntFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Complements the constraint outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInvertConstraints(pAig *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    if Gia_ManConstrNum(pAig) == 0 {
        return; }
    Gia_ManForEachPo( pAig, pObj, i )
        if ( i >= Gia_ManPoNum(pAig) - Gia_ManConstrNum(pAig) )
            Gia_ObjFlipFaninC0( pObj );
}
func Gia_ManInvertPos(pAig *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachPo( pAig, pObj, i )
        Gia_ObjFlipFaninC0( pObj );
}

/**Function*************************************************************

  Synopsis    [Testing the speedup due to grouping POs into batches.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectObjs_rec(p *Gia_Man_t, iObjId int32, vObjs *Vec_Int_t, Limit int32) {
    var pObj *Gia_Obj_t;
    if Vec_IntSize(vObjs) == Limit {
        return; }
    if Gia_ObjIsTravIdCurrentId(p, iObjId) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObjId);
    pObj = Gia_ManObj( p, iObjId );
    if Gia_ObjIsAnd(pObj) {
        Gia_ManCollectObjs_rec( p, Gia_ObjFaninId0p(p, pObj), vObjs, Limit );
        if Vec_IntSize(vObjs) == Limit {
            return; }
        Gia_ManCollectObjs_rec( p, Gia_ObjFaninId1p(p, pObj), vObjs, Limit );
        if Vec_IntSize(vObjs) == Limit {
            return; }
    }
    Vec_IntPush( vObjs, iObjId );
}
func Gia_ManComputePoTruthTables(p *Gia_Man_t, nBytesMax int32) *uint32 {
    var nVars int32  = Gia_ManPiNum(p);
    var nTruthWords int32  = Abc_TruthWordNum( nVars );
    var nTruths int32  = nBytesMax / (sizeof(unsigned) * nTruthWords);
    var nTotalNodes, nRounds int32;
    var vObjs *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var clk abctime  = Abc_Clock();
    var i int32;
    fmt.Printf("Var = %d. Words = %d. Truths = %d.\n", nVars, nTruthWords, nTruths);
    vObjs = Vec_IntAlloc( nTruths );
    Gia_ManIncrementTravId( p );
    Gia_ManForEachPo( p, pObj, i )
    {
        Gia_ManCollectObjs_rec( p, Gia_ObjFaninId0p(p, pObj), vObjs, nTruths );
        if Vec_IntSize(vObjs) == nTruths {
            nRounds++;
//            printf( "%d ", i );
            nTotalNodes += Vec_IntSize( vObjs );
            Vec_IntClear( vObjs );
            Gia_ManIncrementTravId( p );
        }
    }
//    printf( "\n" );
    nTotalNodes += Vec_IntSize( vObjs );
    Vec_IntFree( vObjs );

    fmt.Printf("Rounds = %d. Objects = %d. Total = %d.   ", nRounds, Gia_ManObjNum(p), nTotalNodes);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    return nil;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if the manager are structural identical.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCompare(p1 *Gia_Man_t, p2 *Gia_Man_t) int32 {
    var pObj1, pObj2 *Gia_Obj_t;
    var i int32;
    if Gia_ManObjNum(p1) != Gia_ManObjNum(p2) {
        fmt.Printf("AIGs have different number of objects.\n");
        return 0;
    }
    Gia_ManCleanValue( p1 );
    Gia_ManCleanValue( p2 );
    Gia_ManForEachObj( p1, pObj1, i )
    {
        pObj2 = Gia_ManObj( p2, i );
        if memcmp( pObj1, pObj2, sizeof(Gia_Obj_t) ) {
            fmt.Printf("Objects %d are different.\n", i);
            return 0;
        }
        if p1.pReprs && p2.pReprs {
            if memcmp( &p1.pReprs[i], &p2.pReprs[i], sizeof(Gia_Rpr_t) ) {
                fmt.Printf("Representatives of objects %d are different.\n", i);
                return 0;
            }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Marks nodes that appear as faninis of other nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMarkFanoutDrivers(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
    {
        pObj.fMark0 = 0;
        if Gia_ObjIsAnd(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
            Gia_ObjFanin1(pObj).fMark0 = 1;
        
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1; }
    }
}


/**Function*************************************************************

  Synopsis    [Swaps PO number 0 with PO number i.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSwapPos(p *Gia_Man_t, i int32) {
    var Lit0, LitI int32;
    assert(i >= 0 && i < Gia_ManPoNum(p), "i >= 0 && i < Gia_ManPoNum(p)");
    if i == 0 {
        return; }
    Lit0 = Gia_ObjFaninLit0p( p, Gia_ManPo(p, 0) );
    LitI = Gia_ObjFaninLit0p( p, Gia_ManPo(p, i) );
    Gia_ManPatchCoDriver( p, 0, LitI );
    Gia_ManPatchCoDriver( p, i, Lit0 );
}

/**Function*************************************************************

  Synopsis    [Save/load value from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSaveValue(p *Gia_Man_t) *Vec_Int_t {
    var vValues *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    vValues = Vec_IntAlloc( Gia_ManObjNum(p) );
    Gia_ManForEachObj( p, pObj, i )
        Vec_IntPush( vValues, pObj.Value );
    return vValues;
}
func Gia_ManLoadValue(p *Gia_Man_t, vValues *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
        pObj.Value = Vec_IntEntry(vValues, i);
}


/**Function*************************************************************

  Synopsis    [Returns the array containing the first fanout of each object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFirstFanouts(p *Gia_Man_t) *Vec_Int_t {
    var vFans *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManForEachObj( p, pObj, i )
    {
        if Gia_ObjIsAnd(pObj) {
            if Vec_IntEntry(vFans, Gia_ObjFaninId0p(p, pObj)) == 0 {
                Vec_IntWriteEntry(vFans, Gia_ObjFaninId0p(p, pObj), i); }
            if Vec_IntEntry(vFans, Gia_ObjFaninId1p(p, pObj)) == 0 {
                Vec_IntWriteEntry(vFans, Gia_ObjFaninId1p(p, pObj), i); }
            if Gia_ObjIsMuxId(p, i) && Vec_IntEntry(vFans, Gia_ObjFaninId2p(p, pObj)) == 0 {
                Vec_IntWriteEntry(vFans, Gia_ObjFaninId2p(p, pObj), i); }
        
        } else if Gia_ObjIsCo(pObj) {
            if Vec_IntEntry(vFans, Gia_ObjFaninId0p(p, pObj)) == 0 {
                Vec_IntWriteEntry(vFans, Gia_ObjFaninId0p(p, pObj), i); }
        }
    }
    return vFans;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManHasChoices_very_old(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter1, Counter2 int32;
    var nFailNoRepr int32;
    var nFailHaveRepr int32;
    var nChoiceNodes int32;
    var nChoices int32;
    if p.pReprs == nil || p.pNexts == nil {
        return 0; }
    // check if there are any representatives
    Gia_ManForEachObj( p, pObj, i )
    {
        if Gia_ObjReprObj( p, Gia_ObjId(p, pObj) ) {
//            printf( "%d ", i );
            Counter1++;
        }
//        if ( Gia_ObjNext( p, Gia_ObjId(p, pObj) ) )
//            Counter2++;
    }
//    printf( "\n" );
    Gia_ManForEachObj( p, pObj, i )
    {
//        if ( Gia_ObjReprObj( p, Gia_ObjId(p, pObj) ) )
//            Counter1++;
        if Gia_ObjNext( p, Gia_ObjId(p, pObj) ) {
//            printf( "%d ", i );
            Counter2++;
        }
    }
//    printf( "\n" );
    if Counter1 == 0 {
        fmt.Printf("Warning: AIG has repr data-strucure but not reprs.\n");
        return 0;
    }
    fmt.Printf("%d nodes have reprs.\n", Counter1);
    fmt.Printf("%d nodes have nexts.\n", Counter2);
    // check if there are any internal nodes without fanout
    // make sure all nodes without fanout have representatives
    // make sure all nodes with fanout have no representatives
    p.Refs = nil;
    Gia_ManCreateRefs( p );
    Gia_ManForEachAnd( p, pObj, i )
    {
        if Gia_ObjRefNum(p, pObj) == 0 {
            if Gia_ObjReprObj( p, Gia_ObjId(p, pObj) ) == nil {
                nFailNoRepr++;
            } else {
                nChoices++; }
        
        } else {
            if Gia_ObjReprObj( p, Gia_ObjId(p, pObj) ) != nil {
                nFailHaveRepr++; }
            if Gia_ObjNextObj( p, Gia_ObjId(p, pObj) ) != nil {
                nChoiceNodes++; }
        }
        if Gia_ObjReprObj( p, i ) {
            assert(Gia_ObjRepr(p, i) < i, "Gia_ObjRepr(p, i) < i"); }
    }
    if nChoices == 0 {
        return 0; }
    if nFailNoRepr != 0 {
        fmt.Printf("Gia_ManHasChoices_very_old(): Error: %d internal nodes have no fanout and no repr.\n", nFailNoRepr);
//        return 0;
    }
    if nFailHaveRepr != 0 {
        fmt.Printf("Gia_ManHasChoices_very_old(): Error: %d internal nodes have both fanout and repr.\n", nFailHaveRepr);
//        return 0;
    }
//    printf( "Gia_ManHasChoices_very_old(): AIG has %d choice nodes with %d choices.\n", nChoiceNodes, nChoices );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Proving multi-output properties.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGroupProve(pInit *Gia_Man_t, pCommLine string, nGroupSize int32, fVerbose int32) *Vec_Int_t {
    var pAbc *Abc_Frame_t  = Abc_FrameReadGlobalFrame();
    var p *Gia_Man_t  = Gia_ManDup( pInit );
    var pGroup *Gia_Man_t;
    var vOuts *Vec_Int_t;
    var vOutMap *Vec_Int_t;
    var vCexes *Vec_Ptr_t;
    var i, k, nGroupCur, nGroups int32;
    var clk, timeComm abctime;
    var timeStart abctime  = Abc_Clock();
    // pre-conditions
    assert(nGroupSize > 0, "nGroupSize > 0");
    assert(pCommLine != nil, "pCommLine != nil");
    assert(p.nConstrs == 0, "p.nConstrs == 0");
    Abc_Print( 1, "RUNNING MultiProve: Group size = %d. Command line = \"%s\".\n", nGroupSize, pCommLine );
    // create output map
    vOuts   = Vec_IntStartNatural( Gia_ManPoNum(p) );
    vOutMap = Vec_IntAlloc( Gia_ManPoNum(p) );
    vCexes  = Vec_PtrAlloc( Gia_ManPoNum(p) );
    nGroups = Gia_ManPoNum(p) / nGroupSize + int32(((Gia_ManPoNum(p) % nGroupSize) > 0));
    for  i = 0; i < nGroups; i++  {
        // derive the group
        if (i + 1) * nGroupSize < Gia_ManPoNum(p) {
            nGroupCur = nGroupSize;
        } else {
            nGroupCur = Gia_ManPoNum(p) - i * nGroupSize;
        }
        pGroup = Gia_ManDupCones( p, Vec_IntArray(vOuts) + i * nGroupSize, nGroupCur, 0 );
        Abc_Print( 1, "GROUP %4d : %4d <= PoId < %4d : ", i, i * nGroupSize, i * nGroupSize + nGroupCur );
        // set the current GIA
        Abc_FrameUpdateGia( pAbc, pGroup );
        // solve the group
        clk = Abc_Clock();
        Cmd_CommandExecute( pAbc, pCommLine );
        timeComm += Abc_Clock() - clk;
        // get the solution status
        if nGroupSize == 1 {
            Vec_IntPush( vOutMap, Abc_FrameReadProbStatus(pAbc) );
            Vec_PtrPush( vCexes, Abc_FrameReadCex(pAbc) );
        
        } else { // if ( nGroupSize > 1 )  
        
            var vStatusCur *Vec_Int_t  = Abc_FrameReadPoStatuses( pAbc );
            var vCexesCur *Vec_Ptr_t  = Abc_FrameReadCexVec( pAbc );
            assert(vStatusCur != nil, "vStatusCur != nil"); // only works for "bmc3" and "pdr"
//            assert( vCexesCur != NULL );
            for  k = 0; k < nGroupCur; k++  {
                Vec_IntPush( vOutMap, Vec_IntEntry(vStatusCur, k) );
                Vec_PtrPush( vCexes, vCexesCur ? Vec_PtrEntry(vCexesCur, k) : nil );
            }
        }
    }
    assert(Vec_PtrSize(vCexes) == Gia_ManPoNum(p), "Vec_PtrSize(vCexes) == Gia_ManPoNum(p)");
    assert(Vec_IntSize(vOutMap) == Gia_ManPoNum(p), "Vec_IntSize(vOutMap) == Gia_ManPoNum(p)");
    // set CEXes
    if Vec_PtrCountZero(vCexes) < Vec_PtrSize(vCexes) {
        Abc_FrameReplaceCexVec( pAbc, &vCexes );
    } else {  // there is no CEXes
        Vec_PtrFree( vCexes ); }
    // report the result
    Abc_Print( 1, "SUMMARY:  " );
    Abc_Print( 1, "Properties = %6d. ", Gia_ManPoNum(p) );
    Abc_Print( 1, "UNSAT = %6d. ",      Vec_IntCountEntry(vOutMap, 1) );
    Abc_Print( 1, "SAT = %6d. ",        Vec_IntCountEntry(vOutMap, 0) );
    Abc_Print( 1, "UNDEC = %6d. ",      Vec_IntCountEntry(vOutMap, -1) );
    Abc_Print( 1, "\n" );
    Abc_PrintTime( 1, "Command time", timeComm );
    Abc_PrintTime( 1, "Total time  ", Abc_Clock() - timeStart );
    // cleanup
    Vec_IntFree( vOuts );
    Gia_ManStop( p );
    return vOutMap;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPoXSim(p *Gia_Man_t, nFrames int32, fVerbose int32) *Vec_Int_t {
    var vRes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var f int32;
    var k int32;
    var nLeft int32 = Gia_ManPoNum(p);
    vRes = Vec_IntAlloc( Gia_ManPoNum(p) );
    Vec_IntFill( vRes, Gia_ManPoNum(p), nFrames );
    Gia_ObjTerSimSet0( Gia_ManConst0(p) );
    Gia_ManForEachRi( p, pObj, k )
        Gia_ObjTerSimSet0( pObj );
    for  f = 0; f < nFrames; f++  {
        Gia_ManForEachPi( p, pObj, k )
            Gia_ObjTerSimSetX( pObj );
        Gia_ManForEachRo( p, pObj, k )
            Gia_ObjTerSimRo( p, pObj );
        Gia_ManForEachAnd( p, pObj, k )
            Gia_ObjTerSimAnd( pObj );
        Gia_ManForEachCo( p, pObj, k )
            Gia_ObjTerSimCo( pObj );
        if fVerbose != 0 {
            Gia_ManForEachPo( p, pObj, k )
                Gia_ObjTerSimPrint( pObj );
            fmt.Printf("\n");
        }
        Gia_ManForEachPo( p, pObj, k )
            if ( Vec_IntEntry(vRes, k) == nFrames && Gia_ObjTerSimGetX(pObj) )
                Vec_IntWriteEntry(vRes, k, f);
nLeft--;
        if nLeft == 0 {
            break; }
    }
    if fVerbose != 0 {
        if nLeft == 0 {
            fmt.Printf("Simulation converged after %d frames.\n", f+1);
        } else {
            fmt.Printf("Simulation terminated after %d frames.\n", nFrames); }
    }        
//    Vec_IntPrint( vRes );
    return vRes;
}

const MAX_LUT_SIZE = 8
typedef struct Gia_MapLut_t_
{
    int        Type;          // node type: PI=1, PO=2, LUT=3
    int        Out;           // ID
    int        StartId;       // -1
    int        nFans;         // fanin count
    float      Delay;         // 0.0
    int        pFans[MAX_LUT_SIZE];  // fanin IDs
    unsigned   pTruth[MAX_LUT_SIZE<6?1:(1<<(MAX_LUT_SIZE-5))]; // the truth table
}
 Gia_MapLut_t;

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AigerWriteLut(p *Gia_Man_t, pFileName string) {
    var pObj *Gia_Obj_t;
    var i, k, iFan, iLut int32;
    var LutSizeMax int32  = Gia_ManLutSizeMax( p );
    var nUints int32  = Abc_TruthWordNum(LutSizeMax);
    var nLuts int32  = 1 + Gia_ManCiNum(p) + Gia_ManCoNum(p) + Gia_ManLutNum(p);
    var pLuts *Gia_MapLut_t  = ABC_CALLOC( Gia_MapLut_t, nLuts );
    var vTruths *Vec_Wrd_t  = Vec_WrdStart( Gia_ManObjNum(p) );
    assert(LutSizeMax <= 6, "LutSizeMax <= 6");
    // set obj numbers
    // constant
    pLuts.Type = 3;
    memset( pLuts.pTruth, 0xFF, sizeof(unsigned) * nUints );
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = pLuts[iLut].Out = Abc_Var2Lit( iLut, 0 );
    iLut++;
    // inputs
    Gia_ManForEachCi( p, pObj, i )
    {
        pLuts[iLut].Type = 1;
        memset( pLuts[iLut].pTruth, 0xAA, sizeof(unsigned) * nUints );
        pObj.Value = pLuts[iLut].Out = Abc_Var2Lit( iLut, 0 );
        iLut++;
    }
    // nodes
    Gia_ManForEachObj( p, pObj, i )
        if i && Gia_ObjIsLut(p, i) {
            pLuts[iLut].Type = 3;
            Gia_LutForEachFanin( p, i, iFan, k )
                pLuts[iLut].pFans[k] = Gia_ManObj(p, iFan).Value;
            pLuts[iLut].nFans = k;
            *(word *)pLuts[iLut].pTruth = Gia_LutComputeTruth6(p, i, vTruths);
            pObj.Value = pLuts[iLut].Out = Abc_Var2Lit( iLut, 0 );
            iLut++;
        }
    // outputs
    Gia_ManForEachCo( p, pObj, i )
    {
        pLuts[iLut].Type = 2;
        pLuts[iLut].pFans[0] = Gia_ObjFanin0(pObj).Value;
        if Gia_ObjFaninC0(pObj) ^ Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) {
            memset( pLuts[iLut].pTruth, 0x55, sizeof(unsigned) * nUints );
        } else {
            memset( pLuts[iLut].pTruth, 0xAA, sizeof(unsigned) * nUints ); }
        pLuts[iLut].nFans = 1;
        pObj.Value = pLuts[iLut].Out = Abc_Var2Lit( iLut, 0 );
        iLut++;
    }
    assert(iLut == nLuts, "iLut == nLuts");
    // dump into a file
    {
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        } else {
            var nSize1 int32  = nLuts * sizeof(Gia_MapLut_t);
            var nSize2 int32  = fwrite( pLuts, 1, nSize1, pFile );
            assert(nSize1 == nSize2, "nSize1 == nSize2");
            fmt.Printf("Successfully dumped %d bytes of binary data.\n", nSize1);
        }
        _ = (pFile).Close();
    }
    pLuts = nil;
    Vec_WrdFree( vTruths );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_DumpLutSizeDistrib(p *Gia_Man_t, pFileName string) {
    var pTable *os.File  = os.OpenFile(pFileName, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    var i int32;
    var Counts [10]int32 ;
    Gia_ManForEachLut( p, i )
        if ( Gia_ObjLutSize(p, i) > 0 && Gia_ObjLutSize(p, i) < 10 )
            Counts[ Gia_ObjLutSize(p, i) ]++;
    fmt.Fprintf(pTable, "%s", p.pName);
    for  i = 1; i < 10; i++  {
        fmt.Fprintf(pTable, " %d", Counts[i]);
    }
    fmt.Fprintf(pTable, "\n");
    _ = (pTable).Close();
}

/**Function*************************************************************

  Synopsis    [Check if two logic cones have overlap.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckSuppMark_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if pObj.fMark0 {
        return; }
    pObj.fMark0 = 1;
    if Gia_ObjIsCi(pObj) {
        return; }
    Gia_ManCheckSuppMark_rec( p, Gia_ObjFanin0(pObj) );
    Gia_ManCheckSuppMark_rec( p, Gia_ObjFanin1(pObj) );
}
func Gia_ManCheckSuppUnmark_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if !pObj.fMark0 {
        return; }
    pObj.fMark0 = 0;
    if Gia_ObjIsCi(pObj) {
        return; }
    Gia_ManCheckSuppUnmark_rec( p, Gia_ObjFanin0(pObj) );
    Gia_ManCheckSuppUnmark_rec( p, Gia_ObjFanin1(pObj) );
}
func Gia_ManCheckSupp_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if pObj.fMark0 {
        return 1; }
    if Gia_ObjIsCi(pObj) {
        return 0; }
    if Gia_ManCheckSupp_rec( p, Gia_ObjFanin0(pObj) ) != 0 {
        return 1; }
    return Gia_ManCheckSupp_rec( p, Gia_ObjFanin1(pObj) );
}
func Gia_ManCheckSuppOverlap(p *Gia_Man_t, iNode1 int32, iNode2 int32) int32 {
    var Result int32;
    if iNode1 == 0 || iNode2 == 0 {
        return 0; }
    Gia_ManCheckSuppMark_rec( p, Gia_ManObj(p, iNode1) );
    Result = Gia_ManCheckSupp_rec( p, Gia_ManObj(p, iNode2) );
    Gia_ManCheckSuppUnmark_rec( p, Gia_ManObj(p, iNode1) );
    return Result;
}


/**Function*************************************************************

  Synopsis    [Count PIs with fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountPisWithFanout(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t; 
    var i, Count int32;
    Gia_ManForEachCi( p, pObj, i )
        pObj.fMark0 = 0;
    Gia_ManForEachAnd( p, pObj, i )
    {
        Gia_ObjFanin0(pObj).fMark0 = 1;
        Gia_ObjFanin1(pObj).fMark0 = 1;
    }
    Gia_ManForEachCo( p, pObj, i )
        Gia_ObjFanin0(pObj).fMark0 = 1;
    Gia_ManForEachCi( p, pObj, i )
        Count += pObj.fMark0;
    Gia_ManForEachObj( p, pObj, i )
        pObj.fMark0 = 0;
    return Count;
}

/**Function*************************************************************

  Synopsis    [Count POs driven by non-zero driver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountPosWithNonZeroDrivers(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t; 
    var i, Count int32;
    Gia_ManForEachCo( p, pObj, i )
        Count += Gia_ObjFaninLit0(pObj, Gia_ObjId(p, pObj)) != 0;
    return Count;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUpdateCopy(vCopy *Vec_Int_t, p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i, iLit int32;
    Vec_IntForEachEntry( vCopy, iLit, i )
    {
        if iLit == -1 {
            continue; }
        pObj = Gia_ManObj( p, Abc_Lit2Var(iLit) );
        if !~pObj.Value {
            Vec_IntWriteEntry( vCopy, i, -1 );
        } else {
            Vec_IntWriteEntry( vCopy, i, Abc_LitNotCond(pObj.Value, Abc_LitIsCompl(iLit)) ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupWithMuxPos(p *Gia_Man_t) *Gia_Man_t {
    var vPoints *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vQuads *Vec_Int_t   = Vec_IntAlloc( 1000 );
    var vHeads *Vec_Bit_t   = Vec_BitStart( Gia_ManObjNum(p) );
    var vDatas *Vec_Bit_t   = Vec_BitStart( Gia_ManObjNum(p) );
    var pObj, pCtrl, pData0, pData1 *Gia_Obj_t;
    var pNew *Gia_Man_t  = Gia_ManDup( p ); var i, iObj int32;
    assert(Gia_ManRegNum(pNew) == 0, "Gia_ManRegNum(pNew) == 0");
    Gia_ManForEachAnd( pNew, pObj, i )
    {
        if !(Gia_ObjIsMuxType(pObj) != 0) {
            continue; }
        pCtrl  = Gia_ObjRecognizeMux( pObj, &pData1, &pData0 );
        pCtrl  = Gia_Regular(pCtrl);
        pData0 = Gia_Regular(pData0);
        pData1 = Gia_Regular(pData1);
        Vec_IntPushTwo( vQuads, Gia_ObjId(pNew, pObj),   Gia_ObjId(pNew, pCtrl)  );
        Vec_IntPushTwo( vQuads, Gia_ObjId(pNew, pData0), Gia_ObjId(pNew, pData1) );
        Vec_BitWriteEntry( vHeads, Gia_ObjId(pNew, pObj),   1 );
        Vec_BitWriteEntry( vDatas, Gia_ObjId(pNew, pData0), 1 );
        Vec_BitWriteEntry( vDatas, Gia_ObjId(pNew, pData1), 1 );
    }
    Gia_ManForEachCo( pNew, pObj, i )
        Gia_ObjFanin0(pObj).fMark0 = 1;
    for  i = 0; i < Vec_IntSize(vQuads)/4; i++  {
        var iObj int32    = Vec_IntEntry( vQuads, 4*i+0 );
        var iCtrl int32   = Vec_IntEntry( vQuads, 4*i+1 );
        var iData0 int32  = Vec_IntEntry( vQuads, 4*i+2 );
        var iData1 int32  = Vec_IntEntry( vQuads, 4*i+3 );
        if (Vec_BitEntry(vHeads, iObj)   && Vec_BitEntry(vDatas, iObj))   || 
             (Vec_BitEntry(vHeads, iData0) && Vec_BitEntry(vDatas, iData0)) || 
             (Vec_BitEntry(vHeads, iData1) && Vec_BitEntry(vDatas, iData1)) {
            var pObj *Gia_Obj_t    = Gia_ManObj( p, iObj );
            var pCtrl *Gia_Obj_t   = Gia_ManObj( p, iCtrl );
            var pData0 *Gia_Obj_t  = Gia_ManObj( p, iData0 );
            var pData1 *Gia_Obj_t  = Gia_ManObj( p, iData1 );
            if Gia_ObjIsAnd(pObj)   && !pObj.fMark0 {  Vec_IntPush( vPoints, iObj   ); }
            if Gia_ObjIsAnd(pCtrl)  && !pCtrl.fMark0 {  Vec_IntPush( vPoints, iCtrl  ); }
            if Gia_ObjIsAnd(pData0) && !pData0.fMark0 {  Vec_IntPush( vPoints, iData0 ); }
            if Gia_ObjIsAnd(pData1) && !pData1.fMark0 {  Vec_IntPush( vPoints, iData1 ); }
        }
    }
    Gia_ManCleanMark0( pNew );
    Vec_IntUniqify( vPoints );
    Vec_IntForEachEntry( vPoints, iObj, i )
        Gia_ManAppendCo( pNew, Abc_Var2Lit(iObj, 0) );
    Vec_IntFree( vPoints );
    Vec_IntFree( vQuads );
    Vec_BitFree( vHeads );
    Vec_BitFree( vDatas );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Collect distance info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRingAdd(p *Gia_Man_t, iObj int32, vRes *Vec_Int_t, vDists *Vec_Int_t, Dist int32) {
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    Vec_IntWriteEntry( vDists, iObj, Dist );
    Vec_IntPush( vRes, iObj );
}
func Gia_ManCollectRing(p *Gia_Man_t, vStart *Vec_Int_t, vRes *Vec_Int_t, vDists *Vec_Int_t) {
    var i, k, iObj, iFan int32;
    Vec_IntForEachEntry( vStart, iObj, i )
    {
        var Weight int32  = Vec_IntEntry( vDists, iObj );
        var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj);
        assert(Weight > 0, "Weight > 0");
        if Gia_ObjIsAnd(pObj) {
            Gia_ManRingAdd( p, Gia_ObjFaninId0(pObj, iObj), vRes, vDists, Weight + 1*!Gia_ObjIsBuf(Gia_ObjFanin0(pObj)) );
            Gia_ManRingAdd( p, Gia_ObjFaninId1(pObj, iObj), vRes, vDists, Weight + 1*!Gia_ObjIsBuf(Gia_ObjFanin1(pObj)) );
        }
        Gia_ObjForEachFanoutStaticId( p, iObj, iFan, k )
            Gia_ManRingAdd( p, iFan, vRes, vDists, Weight + 1*!Gia_ObjIsBuf(Gia_ManObj(p, iFan)) );
    }
}
func Gia_ManComputeDistanceInt(p *Gia_Man_t, iTarg int32, vObjs *Vec_Int_t, fVerbose int32) *Vec_Int_t {
    var i, iObj int32;
    var vDists, vStart, vNexts *Vec_Int_t;
    vStart = Vec_IntAlloc( 100 );
    vNexts = Vec_IntAlloc( 100 );
    vDists = Vec_IntStart( Gia_ManObjNum(p) );
    Gia_ManIncrementTravId( p );
    if vObjs {
        Vec_IntForEachEntry( vObjs, iObj, i )
        {
            Gia_ObjSetTravIdCurrentId(p, iObj);
            Vec_IntWriteEntry( vDists, iObj, 1 );
            Vec_IntPush( vStart, iObj );
        }
    
    } else {
        Gia_ObjSetTravIdCurrentId(p, iTarg);
        Vec_IntWriteEntry( vDists, iTarg, 1 );
        Vec_IntPush( vStart, iTarg );
    }
    for  i = 0; ; i++  {
        if fVerbose != 0 {
            fmt.Printf("Ring %2d : %6d\n", i, Vec_IntSize(vDists)-Vec_IntCountZero(vDists)); }        
        Gia_ManCollectRing( p, vStart, vNexts, vDists );
        if Vec_IntSize(vNexts) == 0 {
            break; }
        Vec_IntClear( vStart );
        ABC_SWAP( Vec_Int_t, *vStart, *vNexts );
    }
    Vec_IntFree( vStart );
    Vec_IntFree( vNexts );
    return vDists;
}
func Gia_ManComputeDistance(p *Gia_Man_t, iObj int32, vObjs *Vec_Int_t, fVerbose int32) *Vec_Int_t {
    var vDists *Vec_Int_t;
    if p.vFanoutNums {
        vDists = Gia_ManComputeDistanceInt( p, iObj, vObjs, fVerbose );
    } else {
        Gia_ManStaticFanoutStart( p );
        vDists = Gia_ManComputeDistanceInt( p, iObj, vObjs, fVerbose );
        Gia_ManStaticFanoutStop( p );
    }
    return vDists;
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ComputeTest() {
    var pStart string;
    var Line [1000]byte; var Total float32;
    var pFileName string  = "data.txt";
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil { fmt.Printf("Input file \"%s\" cannot be opened.\n", pFileName); return; }
    for  fgets( Line, 1000, pFile ) != nil  {
        if !strstr(Line, "xxx") {
            continue; }
        if !strstr(Line, "yyy") {
            continue; }
        //printf( "%s", Line );
        pStart = strstr(Line, "zzz");
        if pStart == nil {
            continue; }
        //printf( "%s", pStart + 4 );
        Total += -atof( pStart + 4 );
    }
    fmt.Printf("Total = %.2f\n", Total);
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

