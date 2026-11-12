package sfm

/**CFile****************************************************************

  FileName    [sfmNtk.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Logic network.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmNtk.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_CheckConsistency(vFanins *Vec_Wec_t, nPis int32, nPos int32, vFixed *Vec_Str_t) {
    var vArray *Vec_Int_t;
    var i, k, Fanin int32;
    // check entries
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vFanins)) && (((vArray) = Vec_WecEntry(vFanins, i)), 1); i++  {
        // PIs have no fanins
        if i < nPis {
            assert(Vec_IntSize(vArray) == 0 && Vec_StrEntry(vFixed, i) == byte(0), "Vec_IntSize(vArray) == 0 && Vec_StrEntry(vFixed, i) == byte(0)"); }
        // nodes are in a topo order; POs cannot be fanins
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((Fanin) = Vec_IntEntry(vArray, k)), 1); k++  {
//            assert( Fanin < i && Fanin + nPos < Vec_WecSize(vFanins) );
            assert(Fanin + nPos < Vec_WecSize(vFanins), "Fanin + nPos < Vec_WecSize(vFanins)");
        }
        // POs have one fanout
        if i + nPos >= Vec_WecSize(vFanins) {
            assert(Vec_IntSize(vArray) == 1 && Vec_StrEntry(vFixed, i) == byte(0), "Vec_IntSize(vArray) == 1 && Vec_StrEntry(vFixed, i) == byte(0)"); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_CreateFanout(vFanins *Vec_Wec_t, vFanouts *Vec_Wec_t) {
    var vArray *Vec_Int_t;
    var i, k, Fanin int32;
    // count fanouts
    Vec_WecInit( vFanouts, Vec_WecSize(vFanins) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vFanins)) && (((vArray) = Vec_WecEntry(vFanins, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((Fanin) = Vec_IntEntry(vArray, k)), 1); k++  {
            Vec_WecEntry( vFanouts, Fanin ).nSize++;
        }
    }
    // allocate fanins
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vFanouts)) && (((vArray) = Vec_WecEntry(vFanouts, i)), 1); i++  {
        k = vArray.nSize; vArray.nSize = 0;
        Vec_IntGrow( vArray, k );
    }
    // add fanouts
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vFanins)) && (((vArray) = Vec_WecEntry(vFanins, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((Fanin) = Vec_IntEntry(vArray, k)), 1); k++  {
            Vec_IntPush( Vec_WecEntry( vFanouts, Fanin ), i );
        }
    }
    // verify
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vFanouts)) && (((vArray) = Vec_WecEntry(vFanouts, i)), 1); i++  {
        assert(Vec_IntSize(vArray) == Vec_IntCap(vArray), "Vec_IntSize(vArray) == Vec_IntCap(vArray)");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_ObjLevelNew(vArray *Vec_Int_t, vLevels *Vec_Int_t, fAddLevel int32) int32 {
    var k, Fanin, Level int32;
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vArray)) && (((Fanin) = Vec_IntEntry(vArray, k)), 1); k++  {
        Level = Abc_MaxInt( Level, Vec_IntEntry(vLevels, Fanin) );
    }
    return Level + fAddLevel;
}
func Sfm_CreateLevel(vFanins *Vec_Wec_t, vLevels *Vec_Int_t, vEmpty *Vec_Str_t) {
    var vArray *Vec_Int_t;
    var i int32;
    assert(Vec_IntSize(vLevels) == 0, "Vec_IntSize(vLevels) == 0");
    Vec_IntFill( vLevels, Vec_WecSize(vFanins), 0 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vFanins)) && (((vArray) = Vec_WecEntry(vFanins, i)), 1); i++  {
        Vec_IntWriteEntry( vLevels, i, Sfm_ObjLevelNew(vArray, vLevels, Sfm_ObjAddsLevelArray(vEmpty, i)) );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_ObjLevelNewR(vArray *Vec_Int_t, vLevelsR *Vec_Int_t, fAddLevel int32) int32 {
    var k, Fanout, LevelR int32;
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vArray)) && (((Fanout) = Vec_IntEntry(vArray, k)), 1); k++  {
        LevelR = Abc_MaxInt( LevelR, Vec_IntEntry(vLevelsR, Fanout) );
    }
    return LevelR + fAddLevel;
}
func Sfm_CreateLevelR(vFanouts *Vec_Wec_t, vLevelsR *Vec_Int_t, vEmpty *Vec_Str_t) {
    var vArray *Vec_Int_t;
    var i int32;
    assert(Vec_IntSize(vLevelsR) == 0, "Vec_IntSize(vLevelsR) == 0");
    Vec_IntFill( vLevelsR, Vec_WecSize(vFanouts), 0 );
    // inlining of macro Vec_WecForEachLevelReverse
    for  i = Vec_WecSize(vFanouts)-1; (i >= 0) && (((vArray) = Vec_WecEntry(vFanouts, i)), 1); i--  {
        Vec_IntWriteEntry( vLevelsR, i, Sfm_ObjLevelNewR(vArray, vLevelsR, Sfm_ObjAddsLevelArray(vEmpty, i)) );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkConstruct(vFanins *Vec_Wec_t, nPis int32, nPos int32, vFixed *Vec_Str_t, vEmpty *Vec_Str_t, vTruths *Vec_Wrd_t, vStarts *Vec_Int_t, vTruths2 *Vec_Wrd_t) *Sfm_Ntk_t {
    
    Sfm_CheckConsistency( vFanins, nPis, nPos, vFixed );
    // inlining of macro ABC_CALLOC
    p := &Sfm_Ntk_t{};
    p.nObjs    = Vec_WecSize( vFanins );
    p.nPis     = nPis;
    p.nPos     = nPos;
    p.nNodes   = p.nObjs - p.nPis - p.nPos;
    // user data
    p.vFixed   = vFixed;
    p.vEmpty   = vEmpty;
    p.vTruths  = vTruths;
    p.vFanins  = *vFanins;
    p.vStarts  = vStarts;
    p.vTruths2 = vTruths2;
    vFanins = nil;
    // attributes
    Sfm_CreateFanout( &p.vFanins, &p.vFanouts );
    Sfm_CreateLevel( &p.vFanins, &p.vLevels, vEmpty );
    Sfm_CreateLevelR( &p.vFanouts, &p.vLevelsR, vEmpty );
    Vec_IntFill( &p.vCounts,   p.nObjs,  0 );
    Vec_IntFill( &p.vTravIds,  p.nObjs,  0 );
    Vec_IntFill( &p.vTravIds2, p.nObjs,  0 );
    Vec_IntFill( &p.vId2Var,   2*p.nObjs, -1 );
    Vec_IntFill( &p.vVar2Id,   2*p.nObjs, -1 );
    p.vCover   = Vec_IntAlloc( 1 << 16 );
    p.vCnfs    = Sfm_CreateCnf( p );
    return p;
}
func Sfm_NtkPrepare(p *Sfm_Ntk_t) {
    p.nLevelMax = Vec_IntFindMax(&p.vLevels) + p.pPars.nGrowthLevel;
    p.vNodes    = Vec_IntAlloc( 1000 );
    p.vDivs     = Vec_IntAlloc( 100 );
    p.vRoots    = Vec_IntAlloc( 1000 );
    p.vTfo      = Vec_IntAlloc( 1000 );
    p.vDivCexes = Vec_WrdStart( p.pPars.nWinSizeMax );
    p.vOrder    = Vec_IntAlloc( 100 );
    p.vDivVars  = Vec_IntAlloc( 100 );
    p.vDivIds   = Vec_IntAlloc( 1000 );
    p.vLits     = Vec_IntAlloc( 100 );
    p.vValues   = Vec_IntAlloc( 100 );
    p.vClauses  = Vec_WecAlloc( 100 );
    p.vFaninMap = Vec_IntAlloc( 10 );
    p.pSat      = sat_solver_new();
    sat_solver_setnvars( p.pSat, p.pPars.nWinSizeMax );
}
func Sfm_NtkFree(p *Sfm_Ntk_t) {
    // user data
    Vec_StrFree( p.vFixed );
    Vec_StrFreeP( &p.vEmpty );
    Vec_WrdFree( p.vTruths );
    Vec_WecErase( &p.vFanins );
    Vec_IntFree( p.vStarts );
    Vec_WrdFree( p.vTruths2 );
    // attributes
    Vec_WecErase( &p.vFanouts );
    p.vLevels.pArray = nil;
    p.vLevelsR.pArray = nil;
    p.vCounts.pArray = nil;
    p.vTravIds.pArray = nil;
    p.vTravIds2.pArray = nil;
    p.vId2Var.pArray = nil;
    p.vVar2Id.pArray = nil;
    Vec_WecFree( p.vCnfs );
    Vec_IntFree( p.vCover );
    // other data
    Vec_IntFreeP( &p.vNodes );
    Vec_IntFreeP( &p.vDivs  );
    Vec_IntFreeP( &p.vRoots );
    Vec_IntFreeP( &p.vTfo   );
    Vec_WrdFreeP( &p.vDivCexes );
    Vec_IntFreeP( &p.vOrder );
    Vec_IntFreeP( &p.vDivVars );
    Vec_IntFreeP( &p.vDivIds );
    Vec_IntFreeP( &p.vLits  );
    Vec_IntFreeP( &p.vValues );
    Vec_WecFreeP( &p.vClauses );
    Vec_IntFreeP( &p.vFaninMap );
    if p.pSat { sat_solver_delete( p.pSat ); }
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Performs resubstitution for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkRemoveFanin(p *Sfm_Ntk_t, iNode int32, iFanin int32) {
    var RetValue int32;
    assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)");
    assert(!Sfm_ObjIsPo(p, iFanin), "!Sfm_ObjIsPo(p, iFanin)");
    RetValue = Vec_IntRemove( Sfm_ObjFiArray(p, iNode), iFanin );
    assert((RetValue != 0), "RetValue");
    RetValue = Vec_IntRemove( Sfm_ObjFoArray(p, iFanin), iNode );
    assert((RetValue != 0), "RetValue");
}
func Sfm_NtkAddFanin(p *Sfm_Ntk_t, iNode int32, iFanin int32) {
    if iFanin < 0 {
        return; }
    assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)");
    assert(!Sfm_ObjIsPo(p, iFanin), "!Sfm_ObjIsPo(p, iFanin)");
    assert(Vec_IntFind( Sfm_ObjFiArray(p, iNode), iFanin ) == -1, "Vec_IntFind( Sfm_ObjFiArray(p, iNode), iFanin ) == -1");
    assert(Vec_IntFind( Sfm_ObjFoArray(p, iFanin), iNode ) == -1, "Vec_IntFind( Sfm_ObjFoArray(p, iFanin), iNode ) == -1");
    Vec_IntPush( Sfm_ObjFiArray(p, iNode), iFanin );
    Vec_IntPush( Sfm_ObjFoArray(p, iFanin), iNode );
}
func Sfm_NtkDeleteObj_rec(p *Sfm_Ntk_t, iNode int32) {
    var i, iFanin int32;
    if Sfm_ObjFanoutNum(p, iNode) > 0 || Sfm_ObjIsPi(p, iNode) || Sfm_ObjIsFixed(p, iNode) {
        return; }
    assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)");
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iNode)  && ((iFanin = Sfm_ObjFanin(p, iNode, i)), 1);  i++  {
        var RetValue int32  = Vec_IntRemove( Sfm_ObjFoArray(p, iFanin), iNode );  assert((RetValue != 0), "RetValue");
        Sfm_NtkDeleteObj_rec( p, iFanin );
    }
    Vec_IntClear( Sfm_ObjFiArray(p, iNode) );
    Vec_WrdWriteEntry( p.vTruths, iNode, (word)0 );
}
func Sfm_NtkUpdateLevel_rec(p *Sfm_Ntk_t, iNode int32) {
    var i, iFanout int32;
    var LevelNew int32  = Sfm_ObjLevelNew( Sfm_ObjFiArray(p, iNode), &p.vLevels, Sfm_ObjAddsLevel(p, iNode) );
    if LevelNew == Sfm_ObjLevel(p, iNode) {
        return; }
    Sfm_ObjSetLevel( p, iNode, LevelNew );
    // inlining of macro Sfm_ObjForEachFanout
    for  i = 0; i < Sfm_ObjFanoutNum(p, iNode) && ((iFanout = Sfm_ObjFanout(p, iNode, i)), 1); i++  {
        Sfm_NtkUpdateLevel_rec( p, iFanout );
    }
}
func Sfm_NtkUpdateLevelR_rec(p *Sfm_Ntk_t, iNode int32) {
    var i, iFanin int32;
    var LevelNew int32  = Sfm_ObjLevelNewR( Sfm_ObjFoArray(p, iNode), &p.vLevelsR, Sfm_ObjAddsLevel(p, iNode) );
    if LevelNew == Sfm_ObjLevelR(p, iNode) {
        return; }
    Sfm_ObjSetLevelR( p, iNode, LevelNew );
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iNode)  && ((iFanin = Sfm_ObjFanin(p, iNode, i)), 1);  i++  {
        Sfm_NtkUpdateLevelR_rec( p, iFanin );
    }
}
func Sfm_NtkUpdate(p *Sfm_Ntk_t, iNode int32, f int32, iFaninNew int32, uTruth word) {
    var iFanin int32  = Sfm_ObjFanin( p, iNode, f );
    assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)");
    assert(iFanin != iFaninNew, "iFanin != iFaninNew");
    assert(Sfm_ObjFaninNum(p, iNode) <= 6, "Sfm_ObjFaninNum(p, iNode) <= 6");
    if uTruth == 0 || ~uTruth == 0 {
        // inlining of macro Sfm_ObjForEachFanin
        for  f = 0; f < Sfm_ObjFaninNum(p, iNode)  && ((iFanin = Sfm_ObjFanin(p, iNode, f)), 1);  f++  {
            var RetValue int32  = Vec_IntRemove( Sfm_ObjFoArray(p, iFanin), iNode );  assert((RetValue != 0), "RetValue");
            Sfm_NtkDeleteObj_rec( p, iFanin );
        }
        Vec_IntClear( Sfm_ObjFiArray(p, iNode) );
    
    } else {
        // replace old fanin by new fanin
        Sfm_NtkRemoveFanin( p, iNode, iFanin );
        Sfm_NtkAddFanin( p, iNode, iFaninNew );
        // recursively remove MFFC
        Sfm_NtkDeleteObj_rec( p, iFanin );
    }
    // update logic level
    Sfm_NtkUpdateLevel_rec( p, iNode );
    if iFaninNew != -1 {
        Sfm_NtkUpdateLevelR_rec( p, iFaninNew ); }
    if Sfm_ObjFanoutNum(p, iFanin) > 0 {
        Sfm_NtkUpdateLevelR_rec( p, iFanin ); }
    // update truth table
    Vec_WrdWriteEntry( p.vTruths, iNode, uTruth );
    Sfm_TruthToCnf( uTruth, nil, Sfm_ObjFaninNum(p, iNode), p.vCover, (Vec_Str_t *)Vec_WecEntry(p.vCnfs, iNode) );
}

/**Function*************************************************************

  Synopsis    [Public APIs of this network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NodeReadFanins(p *Sfm_Ntk_t, i int32) *Vec_Int_t {
    return Vec_WecEntry( &p.vFanins, i );
}
func Sfm_NodeReadTruth(p *Sfm_Ntk_t, i int32) *word {
    if Sfm_ObjFaninNum(p, i) <= 6 {
        return Vec_WrdEntryP( p.vTruths, i );
    } else {
        return Vec_WrdEntryP( p.vTruths2, Vec_IntEntry(p.vStarts, i) );
    }
}
func Sfm_NodeReadFixed(p *Sfm_Ntk_t, i int32) int32 {
    return int32(Vec_StrEntry( p.vFixed, i ));
}
func Sfm_NodeReadUsed(p *Sfm_Ntk_t, i int32) int32 {
    return (Sfm_ObjFaninNum(p, i) > 0) || (Sfm_ObjFanoutNum(p, i) > 0);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

