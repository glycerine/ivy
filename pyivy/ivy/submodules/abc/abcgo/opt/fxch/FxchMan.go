package fxch

import "fmt"

/**CFile****************************************************************

  FileName    [ FxchMan.c ]

  PackageName [ Fast eXtract with Cube Hashing (FXCH) ]

  Synopsis    [ Fxch Manager implementation ]

  Author      [ Bruno Schmitt - boschmitt at inf.ufrgs.br ]

  Affiliation [ UFRGS ]

  Date        [ Ver. 1.0. Started - March 6, 2016. ]

  Revision    []

***********************************************************************/
//#include "Fxch.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                LOCAL FUNCTIONS DEFINITIONS                       ///
////////////////////////////////////////////////////////////////////////
static inline int Fxch_ManSCAddRemove( Fxch_Man_t* pFxchMan,
                                       unsigned int SubCubeID,
                                       unsigned int iCube,
                                       unsigned int iLit0,
                                       unsigned int iLit1,
                                       char fAdd,
                                       char fUpdate )
{
    var ret int32;

    if fAdd != 0 {
        ret = Fxch_SCHashTableInsert( pFxchMan.pSCHashTable, pFxchMan.vCubes,
                                      SubCubeID,
                                      iCube, iLit0, iLit1, fUpdate );
    
    } else {
        ret = Fxch_SCHashTableRemove( pFxchMan.pSCHashTable, pFxchMan.vCubes,
                                      SubCubeID,
                                      iCube, iLit0, iLit1, fUpdate );
    }

    return ret;
}


func Fxch_ManDivSingleCube(pFxchMan *Fxch_Man_t, iCube int32, fAdd int32, fUpdate int32) int32 {
    var vCube *Vec_Int_t  = Vec_WecEntry( pFxchMan.vCubes, iCube );
    var i int32;
    var k int32;
    var Lit0 int32;
    var Lit1 int32;
    var fSingleCube int32 = 1;
    var fBase int32;

    if Vec_IntSize( vCube ) < 2 {
        return 0; }

    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(vCube)) && (((Lit0) = Vec_IntEntry(vCube, i)), 1); i++  {
    // inlining of macro Vec_IntForEachEntryStart
    for  k = (i + 1); (k < Vec_IntSize(vCube)) && (((Lit1) = Vec_IntEntry(vCube, k)), 1); k++  {
        var pOutputID *int32;
        var nOnes int32;
        var j int32;
        var z int32;
        assert(Lit0 < Lit1, "Lit0 < Lit1");

        Vec_IntClear( pFxchMan.vCubeFree );
        Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( Abc_LitNot( Lit0 ), 0 ) );
        Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( Abc_LitNot( Lit1 ), 1 ) );

        pOutputID = Vec_IntEntryP( pFxchMan.vOutputID, iCube * pFxchMan.nSizeOutputID );
        nOnes = 0;

        for  j = 0; j < pFxchMan.nSizeOutputID; j++  {
            nOnes += Fxch_CountOnes( pOutputID[j] );
        }

        if nOnes == 0 {
            nOnes = 1; }

        if fAdd != 0 {
            for  z = 0; z < nOnes; z++  {
                Fxch_DivAdd( pFxchMan, fUpdate, fSingleCube, fBase );
            }
            pFxchMan.nPairsS++;
        
        } else {
            for  z = 0; z < nOnes; z++  {
                Fxch_DivRemove( pFxchMan, fUpdate, fSingleCube, fBase );
            }
            pFxchMan.nPairsS--;
        }
    }
    }

    return Vec_IntSize( vCube ) * ( Vec_IntSize( vCube ) - 1 ) / 2;
}

func Fxch_ManDivDoubleCube(pFxchMan *Fxch_Man_t, iCube int32, fAdd int32, fUpdate int32) {
    var vLitHashKeys *Vec_Int_t = pFxchMan.vLitHashKeys;
    var vCube *Vec_Int_t = Vec_WecEntry( pFxchMan.vCubes, iCube );
    var SubCubeID, iLit0, Lit0 int32;

    // inlining of macro Vec_IntForEachEntryStart
    for  iLit0 = 1; (iLit0 < Vec_IntSize(vCube)) && (((Lit0) = Vec_IntEntry(vCube, iLit0)), 1); iLit0++  {
        SubCubeID += Vec_IntEntry( vLitHashKeys, Lit0 );
    }

    Fxch_ManSCAddRemove( pFxchMan, SubCubeID,
                         iCube, 0, 0,
                         byte(fAdd), byte(fUpdate) );

    // inlining of macro Vec_IntForEachEntryStart
    for  iLit0 = 1; (iLit0 < Vec_IntSize(vCube)) && (((Lit0) = Vec_IntEntry(vCube, iLit0)), 1); iLit0++  {
        /* 1 Lit remove */
        SubCubeID -= Vec_IntEntry( vLitHashKeys, Lit0 );

        pFxchMan.nPairsD += Fxch_ManSCAddRemove( pFxchMan, SubCubeID,
                                                  iCube, iLit0, 0,
                                                  byte(fAdd), byte(fUpdate) );

        if Vec_IntSize( vCube ) >= 3 {
            var Lit1, iLit1 int32;

            // inlining of macro Vec_IntForEachEntryStart
            for  iLit1 = iLit0 + 1; (iLit1 < Vec_IntSize(vCube)) && (((Lit1) = Vec_IntEntry(vCube, iLit1)), 1); iLit1++  {
                /* 2 Lit remove */
                SubCubeID -= Vec_IntEntry( vLitHashKeys, Lit1 );

                pFxchMan.nPairsD += Fxch_ManSCAddRemove( pFxchMan, SubCubeID,
                                                          iCube, iLit0, iLit1,
                                                          byte(fAdd), byte(fUpdate) );

                SubCubeID += Vec_IntEntry( vLitHashKeys, Lit1 );
            }
        }

        SubCubeID += Vec_IntEntry( vLitHashKeys, Lit0 );
    }
}

func Fxch_ManCompressCubes(vCubes *Vec_Wec_t, vLit2Cube *Vec_Int_t) {
    var i, k, CubeId int32;

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLit2Cube)) && (((CubeId) = Vec_IntEntry(vLit2Cube, i)), 1); i++  {
        if Vec_IntSize(Vec_WecEntry(vCubes, CubeId)) > 0 {
            Vec_IntWriteEntry( vLit2Cube, k++, CubeId ); }
    }

    Vec_IntShrink( vLit2Cube, k );
}

////////////////////////////////////////////////////////////////////////
///                     PUBLIC INTERFACE                             ///
////////////////////////////////////////////////////////////////////////
func Fxch_ManAlloc(vCubes *Vec_Wec_t) *Fxch_Man_t {
    // inlining of macro ABC_CALLOC
    var pFxchMan *Fxch_Man_t  = &Fxch_Man_t{};

    pFxchMan.vCubes = vCubes;
    pFxchMan.nCubesInit = Vec_WecSize( vCubes );

    pFxchMan.pDivHash = Hsh_VecManStart( 1024 );
    pFxchMan.vDivWeights = Vec_FltAlloc( 1024 );
    pFxchMan.vDivCubePairs = Vec_WecAlloc( 1024 );

    pFxchMan.vCubeFree = Vec_IntAlloc( 4 );
    pFxchMan.vDiv = Vec_IntAlloc( 4 );

    pFxchMan.vCubesS = Vec_IntAlloc( 128 );
    pFxchMan.vPairs = Vec_IntAlloc( 128 );
    pFxchMan.vCubesToUpdate = Vec_IntAlloc( 64 );
    pFxchMan.vCubesToRemove = Vec_IntAlloc( 64 );
    pFxchMan.vSCC = Vec_IntAlloc( 64 );

    return pFxchMan;
}

func Fxch_ManFree(pFxchMan *Fxch_Man_t) {
    Vec_WecFree( pFxchMan.vLits );
    Vec_IntFree( pFxchMan.vLitCount );
    Vec_IntFree( pFxchMan.vLitHashKeys );
    Hsh_VecManStop( pFxchMan.pDivHash );
    Vec_FltFree( pFxchMan.vDivWeights );
    Vec_QueFree( pFxchMan.vDivPrio );
    Vec_WecFree( pFxchMan.vDivCubePairs );
    Vec_IntFree( pFxchMan.vLevels );

    Vec_IntFree( pFxchMan.vCubeFree );
    Vec_IntFree( pFxchMan.vDiv );

    Vec_IntFree( pFxchMan.vCubesS );
    Vec_IntFree( pFxchMan.vPairs );
    Vec_IntFree( pFxchMan.vCubesToUpdate );
    Vec_IntFree( pFxchMan.vCubesToRemove );
    Vec_IntFree( pFxchMan.vSCC );

    pFxchMan = nil;
}

func Fxch_ManMapLiteralsIntoCubes(pFxchMan *Fxch_Man_t, nVars int32) {

    var vCube *Vec_Int_t;
    var i, k, Lit, Count int32;

    pFxchMan.nVars = 0;
    pFxchMan.nLits = 0;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pFxchMan.vCubes)) && (((vCube) = Vec_WecEntry(pFxchMan.vCubes, i)), 1); i++  {
        assert(Vec_IntSize(vCube) > 0, "Vec_IntSize(vCube) > 0");
        pFxchMan.nVars = Abc_MaxInt( pFxchMan.nVars, Vec_IntEntry( vCube, 0 ) );
        pFxchMan.nLits += Vec_IntSize(vCube) - 1;
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
            pFxchMan.nVars = Abc_MaxInt( pFxchMan.nVars, Abc_Lit2Var( Lit ) );
        }
    }

    assert(pFxchMan.nVars < nVars, "pFxchMan.nVars < nVars");
    pFxchMan.nVars = nVars;

    /* Count the number of time each literal appears on the SOP */
    pFxchMan.vLitCount = Vec_IntStart( 2 * pFxchMan.nVars );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pFxchMan.vCubes)) && (((vCube) = Vec_WecEntry(pFxchMan.vCubes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
            Vec_IntAddToEntry( pFxchMan.vLitCount, Lit, 1 );
        }
    }

    /* Allocate space to the array of arrays wich maps Literals into cubes which uses them */
    pFxchMan.vLits = Vec_WecStart( 2 * pFxchMan.nVars );
    // inlining of macro Vec_IntForEachEntry
    for  Lit = 0; (Lit < Vec_IntSize(pFxchMan.vLitCount)) && (((Count) = Vec_IntEntry(pFxchMan.vLitCount, Lit)), 1); Lit++  {
        Vec_IntGrow( Vec_WecEntry( pFxchMan.vLits, Lit ), Count );
    }

    /* Maps Literals into cubes which uses them */
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pFxchMan.vCubes)) && (((vCube) = Vec_WecEntry(pFxchMan.vCubes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
            Vec_WecPush( pFxchMan.vLits, Lit, i );
        }
    }
}

func Fxch_ManGenerateLitHashKeys(pFxchMan *Fxch_Man_t) {
    var i int32;
    /* Generates the random number which will be used for hashing cubes */
    Gia_ManRandom( 1 );
    pFxchMan.vLitHashKeys = Vec_IntAlloc( ( 2 * pFxchMan.nVars ) );
    for  i = 0; i < (2 * pFxchMan.nVars); i++  {
        Vec_IntPush( pFxchMan.vLitHashKeys, Gia_ManRandom(0) & 0x3FFFFFF );
    }
}

func Fxch_ManSCHashTablesInit(pFxchMan *Fxch_Man_t) {
    var vCubes *Vec_Wec_t  = pFxchMan.vCubes;
    var vCube *Vec_Int_t;
    var iCube, nTotalHashed int32;

    // inlining of macro Vec_WecForEachLevel
    for  iCube = 0; (iCube < Vec_WecSize(vCubes)) && (((vCube) = Vec_WecEntry(vCubes, iCube)), 1); iCube++  {
        var nLits int32 = Vec_IntSize( vCube ) - 1;
        var nSubCubes int32;
        if nLits <= 2 {
            nSubCubes = nLits + 1;
        } else {
            nSubCubes = ( nLits * nLits + nLits ) / 2;
        }

        nTotalHashed += nSubCubes + 1;
    }

    pFxchMan.pSCHashTable = Fxch_SCHashTableCreate( pFxchMan, nTotalHashed );
}

func Fxch_ManSCHashTablesFree(pFxchMan *Fxch_Man_t) {
    Fxch_SCHashTableDelete( pFxchMan.pSCHashTable );
}

func Fxch_ManDivCreate(pFxchMan *Fxch_Man_t) {
    var vCube *Vec_Int_t;
    var Weight float32;
    var fAdd int32 = 1;
    var fUpdate int32;
    var iCube int32;

    // inlining of macro Vec_WecForEachLevel
    for  iCube = 0; (iCube < Vec_WecSize(pFxchMan.vCubes)) && (((vCube) = Vec_WecEntry(pFxchMan.vCubes, iCube)), 1); iCube++  {
        Fxch_ManDivSingleCube( pFxchMan, iCube, fAdd, fUpdate );
        Fxch_ManDivDoubleCube( pFxchMan, iCube, fAdd, fUpdate );
    }

    pFxchMan.vDivPrio = Vec_QueAlloc( Vec_FltSize( pFxchMan.vDivWeights ) );
    Vec_QueSetPriority( pFxchMan.vDivPrio, Vec_FltArrayP( pFxchMan.vDivWeights ) );
    // inlining of macro Vec_FltForEachEntry
    for  iCube = 0; (iCube < Vec_FltSize(pFxchMan.vDivWeights)) && (((Weight) = Vec_FltEntry(pFxchMan.vDivWeights, iCube)), 1); iCube++  {
        if Weight > 0.0 {
            Vec_QuePush( pFxchMan.vDivPrio, iCube ); }
    }
}

/* Level Computation */
func Fxch_ManComputeLevelDiv(pFxchMan *Fxch_Man_t, vCubeFree *Vec_Int_t) int32 {
    var i, Lit, Level int32;

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCubeFree)) && (((Lit) = Vec_IntEntry(vCubeFree, i)), 1); i++  {
        Level = Abc_MaxInt( Level, Vec_IntEntry( pFxchMan.vLevels, Abc_Lit2Var( Abc_Lit2Var( Lit ) ) ) );
    }

    return Abc_MinInt( Level, 800 );
}

func Fxch_ManComputeLevelCube(pFxchMan *Fxch_Man_t, vCube *Vec_Int_t) int32 {
    var k, Lit, Level int32;

    // inlining of macro Vec_IntForEachEntryStart
    for  k = 1; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
        Level = Abc_MaxInt( Level, Vec_IntEntry( pFxchMan.vLevels, Abc_Lit2Var( Lit ) ) );
    }

    return Level;
}

func Fxch_ManComputeLevel(pFxchMan *Fxch_Man_t) {
    var vCube *Vec_Int_t;
    var i, iVar, iFirst int32;

    iVar = Vec_IntEntry( Vec_WecEntry( pFxchMan.vCubes, 0 ), 0 );
    pFxchMan.vLevels = Vec_IntStart( pFxchMan.nVars );

    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(pFxchMan.vCubes)) && (((vCube) = Vec_WecEntry(pFxchMan.vCubes, i)), 1); i++  {
        if iVar != Vec_IntEntry( vCube, 0 ) {
            Vec_IntAddToEntry( pFxchMan.vLevels, iVar, i - iFirst );
            iVar = Vec_IntEntry( vCube, 0 );
            iFirst = i;
        }
        Vec_IntUpdateEntry( pFxchMan.vLevels, iVar, Fxch_ManComputeLevelCube( pFxchMan, vCube ) );
    }
}

/* Extract divisor from single-cubes */
func Fxch_ManExtractDivFromCube(pFxchMan *Fxch_Man_t, Lit0 int32, Lit1 int32, iVarNew int32) {
    var i, iCube0 int32;
    var vLitP *Vec_Int_t  = Vec_WecEntry( pFxchMan.vLits, Vec_WecSize( pFxchMan.vLits ) - 2 );

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pFxchMan.vCubesS)) && (((iCube0) = Vec_IntEntry(pFxchMan.vCubesS, i)), 1); i++  {
        var RetValue int32;
        var vCube0 *Vec_Int_t;

        vCube0 = Fxch_ManGetCube( pFxchMan, iCube0 );
        RetValue  = Vec_IntRemove1( vCube0, Abc_LitNot( Lit0 ) );
        RetValue += Vec_IntRemove1( vCube0, Abc_LitNot( Lit1 ) );
        assert(RetValue == 2, "RetValue == 2");

        Vec_IntPush( vCube0, Abc_Var2Lit( iVarNew, 0 ) );
        Vec_IntPush( vLitP, Vec_WecLevelId( pFxchMan.vCubes, vCube0 ) );
        Vec_IntPush( pFxchMan.vCubesToUpdate, iCube0 );

        pFxchMan.nLits--;
    }
}

/* Extract divisor from cube pairs */
func Fxch_ManExtractDivFromCubePairs(pFxchMan *Fxch_Man_t, iVarNew int32) {
    /* For each pair (Ci, Cj) */
    var iCube0, iCube1, i int32;


    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(pFxchMan.vPairs)) && (((iCube0) = Vec_IntEntry(pFxchMan.vPairs, i)), 1) && (((iCube1) = Vec_IntEntry(pFxchMan.vPairs, i+1)), 1); i += 2  {
        var j, Lit, RetValue, fCompl int32;
        var pOutputID0, pOutputID1 *int32;

        var vCube *Vec_Int_t = nil;
        var vCube0 *Vec_Int_t = Fxch_ManGetCube( pFxchMan, iCube0 );
        var vCube1 *Vec_Int_t = Fxch_ManGetCube( pFxchMan, iCube1 );
        var vCube0Copy *Vec_Int_t = Vec_IntDup( vCube0 );
        var vCube1Copy *Vec_Int_t = Vec_IntDup( vCube1 );

        RetValue  = Fxch_DivRemoveLits( vCube0Copy, vCube1Copy, pFxchMan.vDiv, &fCompl );

        assert(RetValue == Vec_IntSize( pFxchMan.vDiv ), "RetValue == Vec_IntSize( pFxchMan.vDiv )");

        pFxchMan.nLits -= Vec_IntSize( pFxchMan.vDiv ) + Vec_IntSize( vCube1 ) - 2;

        /* Identify type of Extraction */
        pOutputID0 = Vec_IntEntryP( pFxchMan.vOutputID, iCube0 * pFxchMan.nSizeOutputID );
        pOutputID1 = Vec_IntEntryP( pFxchMan.vOutputID, iCube1 * pFxchMan.nSizeOutputID );
        RetValue = 1;
        for  j = 0; j < pFxchMan.nSizeOutputID && (RetValue != 0); j++  {
            RetValue = ( pOutputID0[j] == pOutputID1[j] );
        }

        /* Exact Extractraion */
        if RetValue != 0 {
            Vec_IntClear( vCube0 );
            Vec_IntAppend( vCube0, vCube0Copy );
            vCube = vCube0;

            Vec_IntPush( pFxchMan.vCubesToUpdate, iCube0 );
            Vec_IntClear( vCube1 );

            /* Update Lit -> Cube mapping */
            // inlining of macro Vec_IntForEachEntry
            for  j = 0; (j < Vec_IntSize(pFxchMan.vDiv)) && (((Lit) = Vec_IntEntry(pFxchMan.vDiv, j)), 1); j++  {
                Vec_IntRemove( Vec_WecEntry( pFxchMan.vLits, Abc_Lit2Var( Lit ) ),
                               Vec_WecLevelId( pFxchMan.vCubes, vCube0 ) );
                Vec_IntRemove( Vec_WecEntry( pFxchMan.vLits, Abc_LitNot( Abc_Lit2Var( Lit ) ) ),
                               Vec_WecLevelId( pFxchMan.vCubes, vCube0 ) );
            }

        
        /* Unexact Extraction */
        } else {
            for  j = 0; j < pFxchMan.nSizeOutputID; j++  {
                pFxchMan.pTempOutputID[j] = ( pOutputID0[j] & pOutputID1[j] );
            }

            /* Create new cube */
            vCube = Vec_WecPushLevel( pFxchMan.vCubes );
            Vec_IntAppend( vCube, vCube0Copy );
            Vec_IntPushArray( pFxchMan.vOutputID, pFxchMan.pTempOutputID, pFxchMan.nSizeOutputID );
            Vec_IntPush( pFxchMan.vCubesToUpdate, Vec_WecLevelId( pFxchMan.vCubes, vCube ) );

            /* Update Lit -> Cube mapping */
            // inlining of macro Vec_IntForEachEntryStart
            for  j = 1; (j < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, j)), 1); j++  {
                Vec_WecPush( pFxchMan.vLits, Lit, Vec_WecLevelId( pFxchMan.vCubes, vCube ) );
            }

            /*********************************************************/
            RetValue = 0;
            for  j = 0; j < pFxchMan.nSizeOutputID; j++  {
                pFxchMan.pTempOutputID[j] = pOutputID0[j];
                RetValue |= ( pOutputID0[j] & ~( pOutputID1[j] ) );
                pOutputID0[j] &= ~( pOutputID1[j] );
            }

            if RetValue != 0 {
                Vec_IntPush( pFxchMan.vCubesToUpdate, iCube0 );
            } else {
                Vec_IntClear( vCube0 ); }

            /*********************************************************/
            RetValue = 0;
            for  j = 0; j < pFxchMan.nSizeOutputID; j++  {
                RetValue |= ( pOutputID1[j] & ~( pFxchMan.pTempOutputID[j] ) );
                pOutputID1[j] &= ~( pFxchMan.pTempOutputID[j] );
            }

            if RetValue != 0 {
                Vec_IntPush( pFxchMan.vCubesToUpdate, iCube1 );
            } else {
                Vec_IntClear( vCube1 ); }

        }
        Vec_IntFree( vCube0Copy );
        Vec_IntFree( vCube1Copy );

        if iVarNew != 0 {
            var vLitP *Vec_Int_t = Vec_WecEntry( pFxchMan.vLits, Vec_WecSize( pFxchMan.vLits ) - 2 );
            var vLitN *Vec_Int_t = Vec_WecEntry( pFxchMan.vLits, Vec_WecSize( pFxchMan.vLits ) - 1 );

            assert((vCube != nil), "vCube");
            if Vec_IntSize( pFxchMan.vDiv ) == 2 || (fCompl != 0) {
                Vec_IntPush( vCube, Abc_Var2Lit( iVarNew, 1 ) );
                Vec_IntPush( vLitN, Vec_WecLevelId( pFxchMan.vCubes, vCube ) );
                Vec_IntSort( vLitN, 0 );
            
            } else {
                Vec_IntPush( vCube, Abc_Var2Lit( iVarNew, 0 ) );
                Vec_IntPush( vLitP, Vec_WecLevelId( pFxchMan.vCubes, vCube ) );
                Vec_IntSort( vLitP, 0 );
            }
        }
    }

    return;
}

func Fxch_ManCreateCube(pFxchMan *Fxch_Man_t, Lit0 int32, Lit1 int32) int32 {
    var Level, iVarNew, j int32;
    var vCube0, vCube1 *Vec_Int_t;

    /* Create a new variable */
    iVarNew = pFxchMan.nVars;
    pFxchMan.nVars++;

    /* Clear temporary outputID vector */
    for  j = 0; j < pFxchMan.nSizeOutputID; j++  {
        pFxchMan.pTempOutputID[j] = 0;
    }

    /* Create new Lit hash keys */
    Vec_IntPush( pFxchMan.vLitHashKeys, Gia_ManRandom(0) & 0x3FFFFFF );
    Vec_IntPush( pFxchMan.vLitHashKeys, Gia_ManRandom(0) & 0x3FFFFFF );

    /* Create new Cube */
    vCube0 = Vec_WecPushLevel( pFxchMan.vCubes );
    Vec_IntPush( vCube0, iVarNew );
    Vec_IntPushArray( pFxchMan.vOutputID, pFxchMan.pTempOutputID, pFxchMan.nSizeOutputID );

    if Vec_IntSize( pFxchMan.vDiv ) == 2 {
        if Lit0 > Lit1 {
            // inlining of macro ABC_SWAP
             var t int32  = Lit0; Lit0 = Lit1; Lit1 = t; };

        Vec_IntPush( vCube0, Abc_LitNot( Lit0 ) );
        Vec_IntPush( vCube0, Abc_LitNot( Lit1 ) );
        Level = 1 + Fxch_ManComputeLevelCube( pFxchMan, vCube0 );
    
    } else {
        var i int32;
        var Lit int32;

        vCube1 = Vec_WecPushLevel( pFxchMan.vCubes );
        Vec_IntPush( vCube1, iVarNew );
        Vec_IntPushArray( pFxchMan.vOutputID, pFxchMan.pTempOutputID, pFxchMan.nSizeOutputID );

        vCube0 = Fxch_ManGetCube( pFxchMan, Vec_WecSize( pFxchMan.vCubes ) - 2 );
        Fxch_DivSepareteCubes( pFxchMan.vDiv, vCube0, vCube1 );
        Level = 2 + Abc_MaxInt( Fxch_ManComputeLevelCube( pFxchMan, vCube0 ),
                                Fxch_ManComputeLevelCube( pFxchMan, vCube1 ) );

        Vec_IntPush( pFxchMan.vCubesToUpdate, Vec_WecLevelId( pFxchMan.vCubes, vCube0 ) );
        Vec_IntPush( pFxchMan.vCubesToUpdate, Vec_WecLevelId( pFxchMan.vCubes, vCube1 ) );

        /* Update Lit -> Cube mapping */
        // inlining of macro Vec_IntForEachEntryStart
        for  i = 1; (i < Vec_IntSize(vCube0)) && (((Lit) = Vec_IntEntry(vCube0, i)), 1); i++  {
            Vec_WecPush( pFxchMan.vLits, Lit, Vec_WecLevelId( pFxchMan.vCubes, vCube0 ) );
        }

        // inlining of macro Vec_IntForEachEntryStart
        for  i = 1; (i < Vec_IntSize(vCube1)) && (((Lit) = Vec_IntEntry(vCube1, i)), 1); i++  {
            Vec_WecPush( pFxchMan.vLits, Lit, Vec_WecLevelId( pFxchMan.vCubes, vCube1 ) );
        }
    }
    assert(Vec_IntSize( pFxchMan.vLevels ) == iVarNew, "Vec_IntSize( pFxchMan.vLevels ) == iVarNew");
    Vec_IntPush( pFxchMan.vLevels, Level );

    pFxchMan.nLits += Vec_IntSize( pFxchMan.vDiv );

    /* Create new literals */
    Vec_WecPushLevel( pFxchMan.vLits );
    Vec_WecPushLevel( pFxchMan.vLits );

    return iVarNew;
}

func Fxch_ManUpdate(pFxchMan *Fxch_Man_t, iDiv int32) {
    var i int32;
    var iCube0 int32;
    var iCube1 int32;
    var Lit0 int32 = -1;
    var Lit1 int32 = -1;
    var iVarNew int32;

    var vCube0, vCube1, vDivCubePairs *Vec_Int_t;

    /* Get the selected candidate (divisor) */
    Vec_IntClear( pFxchMan.vDiv );
    Vec_IntAppend( pFxchMan.vDiv, Hsh_VecReadEntry( pFxchMan.pDivHash, iDiv ) );

    /* Find cubes associated with the divisor */
    Vec_IntClear( pFxchMan.vCubesS );
    if Vec_IntSize( pFxchMan.vDiv ) == 2 {
        Lit0 = Abc_Lit2Var( Vec_IntEntry( pFxchMan.vDiv, 0 ) );
        Lit1 = Abc_Lit2Var( Vec_IntEntry( pFxchMan.vDiv, 1 ) );
        assert(Lit0 >= 0 && Lit1 >= 0, "Lit0 >= 0 && Lit1 >= 0");

        Fxch_ManCompressCubes( pFxchMan.vCubes, Vec_WecEntry( pFxchMan.vLits, Abc_LitNot( Lit0 ) ) );
        Fxch_ManCompressCubes( pFxchMan.vCubes, Vec_WecEntry( pFxchMan.vLits, Abc_LitNot( Lit1 ) ) );
        Vec_IntTwoRemoveCommon( Vec_WecEntry( pFxchMan.vLits, Abc_LitNot( Lit0 ) ),
                                Vec_WecEntry( pFxchMan.vLits, Abc_LitNot( Lit1 ) ),
                                pFxchMan.vCubesS );
    }

    /* Find pairs associated with the divisor */
    Vec_IntClear( pFxchMan.vPairs );
    vDivCubePairs = Vec_WecEntry( pFxchMan.vDivCubePairs, iDiv );
    Vec_IntAppend( pFxchMan.vPairs, vDivCubePairs );
    Vec_IntErase( vDivCubePairs );

    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(pFxchMan.vPairs)) && (((iCube0) = Vec_IntEntry(pFxchMan.vPairs, i)), 1) && (((iCube1) = Vec_IntEntry(pFxchMan.vPairs, i+1)), 1); i += 2  {
        assert(Fxch_ManGetLit( pFxchMan, iCube0, 0) == Fxch_ManGetLit( pFxchMan, iCube1, 0), "Fxch_ManGetLit( pFxchMan, iCube0, 0) == Fxch_ManGetLit( pFxchMan, iCube1, 0)");
        if iCube0 > iCube1 {
            Vec_IntSetEntry( pFxchMan.vPairs, i, iCube1);
            Vec_IntSetEntry( pFxchMan.vPairs, i+1, iCube0);
        }
    }

    Vec_IntUniqifyPairs( pFxchMan.vPairs );
    assert(Vec_IntSize( pFxchMan.vPairs ) % 2 == 0, "Vec_IntSize( pFxchMan.vPairs ) % 2 == 0");

    /* subtract cost of single-cube divisors */
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pFxchMan.vCubesS)) && (((iCube0) = Vec_IntEntry(pFxchMan.vCubesS, i)), 1); i++  {
        Fxch_ManDivSingleCube( pFxchMan, iCube0, 0, 1);

        if Vec_WecEntryEntry( pFxchMan.vCubes, iCube0, 0 ) == 0 {
            Fxch_ManDivDoubleCube( pFxchMan, iCube0, 0, 1 ); }
    }

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pFxchMan.vPairs)) && (((iCube0) = Vec_IntEntry(pFxchMan.vPairs, i)), 1); i++  {
        Fxch_ManDivSingleCube( pFxchMan, iCube0, 0, 1);

        if Vec_WecEntryEntry( pFxchMan.vCubes, iCube0, 0 ) == 0 {
            Fxch_ManDivDoubleCube( pFxchMan, iCube0, 0, 1 ); }
    }

    Vec_IntClear( pFxchMan.vCubesToUpdate );
    if Fxch_DivIsNotConstant1( pFxchMan.vDiv ) {
        iVarNew = Fxch_ManCreateCube( pFxchMan, Lit0, Lit1 );
        Fxch_ManExtractDivFromCube( pFxchMan, Lit0, Lit1, iVarNew );
        Fxch_ManExtractDivFromCubePairs( pFxchMan, iVarNew );
    
    } else {
        Fxch_ManExtractDivFromCubePairs( pFxchMan, 0 ); }

    assert(Vec_IntSize( pFxchMan.vCubesToUpdate ), "Vec_IntSize( pFxchMan.vCubesToUpdate )");

    /* Add cost */
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pFxchMan.vCubesToUpdate)) && (((iCube0) = Vec_IntEntry(pFxchMan.vCubesToUpdate, i)), 1); i++  {
        Fxch_ManDivSingleCube( pFxchMan, iCube0, 1, 1 );

        if Vec_WecEntryEntry( pFxchMan.vCubes, iCube0, 0 ) == 0 {
            Fxch_ManDivDoubleCube( pFxchMan, iCube0, 1, 1 ); }
    }

    /* Deal with SCC */
    if Vec_IntSize( pFxchMan.vSCC ) {
        Vec_IntUniqifyPairs( pFxchMan.vSCC );
        assert(Vec_IntSize( pFxchMan.vSCC ) % 2 == 0, "Vec_IntSize( pFxchMan.vSCC ) % 2 == 0");

        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(pFxchMan.vSCC)) && (((iCube0) = Vec_IntEntry(pFxchMan.vSCC, i)), 1) && (((iCube1) = Vec_IntEntry(pFxchMan.vSCC, i+1)), 1); i += 2  {
            var j int32;
            var RetValue int32 = 1;
            var pOutputID0 *int32  = Vec_IntEntryP( pFxchMan.vOutputID, iCube0 * pFxchMan.nSizeOutputID );
            var pOutputID1 *int32  = Vec_IntEntryP( pFxchMan.vOutputID, iCube1 * pFxchMan.nSizeOutputID );
            vCube0 = Vec_WecEntry( pFxchMan.vCubes, iCube0 );
            vCube1 = Vec_WecEntry( pFxchMan.vCubes, iCube1 );

            if !Vec_WecIntHasMark( vCube0 ) {
                Fxch_ManDivSingleCube( pFxchMan, iCube0, 0, 1 );
                Fxch_ManDivDoubleCube( pFxchMan, iCube0, 0, 1 );
                Vec_WecIntSetMark( vCube0 );
            }

            if !Vec_WecIntHasMark( vCube1 ) {
                Fxch_ManDivSingleCube( pFxchMan, iCube1, 0, 1 );
                Fxch_ManDivDoubleCube( pFxchMan, iCube1, 0, 1 );
                Vec_WecIntSetMark( vCube1 );
            }

            if Vec_IntSize( vCube0 ) == Vec_IntSize( vCube1 ) {
                for  j = 0; j < pFxchMan.nSizeOutputID; j++  {
                    pOutputID1[j] |= pOutputID0[j];
                    pOutputID0[j] = 0;
                }
                Vec_IntClear( Vec_WecEntry( pFxchMan.vCubes, iCube0 ) );
                Vec_WecIntXorMark( vCube0 );
                continue;
            }

            for  j = 0; j < pFxchMan.nSizeOutputID && (RetValue != 0); j++  {
                RetValue = ( pOutputID0[j] == pOutputID1[j] );
            }

            if RetValue != 0 {
                Vec_IntClear( Vec_WecEntry( pFxchMan.vCubes, iCube0 ) );
                Vec_WecIntXorMark( vCube0 );
            
            } else {
                RetValue = 0;
                for  j = 0; j < pFxchMan.nSizeOutputID; j++  {
                    RetValue |= ( pOutputID0[j] & ~( pOutputID1[j] ) );
                    pOutputID0[j] &= ~( pOutputID1[j] );
                }

                if RetValue == 0 {
                    Vec_IntClear( Vec_WecEntry( pFxchMan.vCubes, iCube0 ) );
                    Vec_WecIntXorMark( vCube0 );
                }
            }
        }

        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(pFxchMan.vSCC)) && (((iCube0) = Vec_IntEntry(pFxchMan.vSCC, i)), 1) && (((iCube1) = Vec_IntEntry(pFxchMan.vSCC, i+1)), 1); i += 2  {
            vCube0 = Vec_WecEntry( pFxchMan.vCubes, iCube0 );
            vCube1 = Vec_WecEntry( pFxchMan.vCubes, iCube1 );

            if Vec_WecIntHasMark( vCube0 ) {
                Fxch_ManDivSingleCube( pFxchMan, iCube0, 1, 1 );
                Fxch_ManDivDoubleCube( pFxchMan, iCube0, 1, 1 );
                Vec_WecIntXorMark( vCube0 );
            }

            if Vec_WecIntHasMark( vCube1 ) {
                Fxch_ManDivSingleCube( pFxchMan, iCube1, 1, 1 );
                Fxch_ManDivDoubleCube( pFxchMan, iCube1, 1, 1 );
                Vec_WecIntXorMark( vCube1 );
            }
        }

        Vec_IntClear( pFxchMan.vSCC );
    }

    pFxchMan.nExtDivs++;
}

/* Print */
func Fxch_ManPrintDivs(pFxchMan *Fxch_Man_t) {
    var iDiv int32;

    for  iDiv = 0; iDiv < Vec_FltSize( pFxchMan.vDivWeights ); iDiv++  {
        Fxch_DivPrint( pFxchMan, iDiv );
    }
}

func Fxch_ManPrintStats(pFxchMan *Fxch_Man_t) {
    fmt.Printf("Cubes =%8d  ", Vec_WecSizeUsed( pFxchMan.vCubes ));
    fmt.Printf("Lits  =%8d  ", Vec_WecSizeUsed( pFxchMan.vLits ));
    fmt.Printf("Divs  =%8d  ", Hsh_VecSize( pFxchMan.pDivHash ));
    fmt.Printf("Divs+ =%8d  ", Vec_QueSize( pFxchMan.vDivPrio ));
    fmt.Printf("Extr  =%7d  \n", pFxchMan.nExtDivs);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
