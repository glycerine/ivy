package fxch

import "fmt"

/**CFile****************************************************************

  FileName    [ FxchSCHashTable.c ]

  PackageName [ Fast eXtract with Cube Hashing (FXCH) ]

  Synopsis    [ Sub-cubes hash table implementation ]

  Author      [ Bruno Schmitt - boschmitt at inf.ufrgs.br ]

  Affiliation [ UFRGS ]

  Date        [ Ver. 1.0. Started - March 6, 2016. ]

  Revision    []

***********************************************************************/
//#include "Fxch.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
func MurmurHash3_x86_32(key any, len int32, seed uint32_t, out any) {
    uint8_t* data = (uint8_t*)key;
    int nblocks = len / 4;

    var h1 uint32_t  = seed;

    uint32_t c1 = 0xcc9e2d51;
    uint32_t c2 = 0x1b873593;

    uint8_t * tail;
    var k1 uint32_t;

    //----------
    // body

    uint32_t * blocks = (uint32_t *)(data + nblocks*4);
    var i int32;

    for i = -nblocks; i; i++ {
        var k1 uint32_t  = blocks[i];

        k1 *= c1;
        k1 = (k1 << 15) | (k1 >> (32 - 15));
        k1 *= c2;

        h1 ^= k1;
        h1 = (h1 << 13) | (h1 >> (32 - 13));
        h1 = h1*5+0xe6546b64;
    }

    //----------
    // tail

    tail = (uint8_t*)(data + nblocks*4);

    k1 = 0;

    switch len & 3 {
        case 3: k1 ^= tail[2] << 16;
        fallthrough
case 2: k1 ^= tail[1] << 8;
        fallthrough
case 1: k1 ^= tail[0];
              k1 *= c1; k1 = (k1 << 15) | (k1 >> (32 - 15)); k1 *= c2; h1 ^= k1;
    };

    //----------
    // finalization

    h1 ^= len;

    h1 ^= h1 >> 16;
    h1 *= 0x85ebca6b;
    h1 ^= h1 >> 13;
    h1 *= 0xc2b2ae35;
    h1 ^= h1 >> 16;

    *(uint32_t*)out = h1;
}

func Fxch_SCHashTableCreate(pFxchMan *Fxch_Man_t, nEntries int32) *Fxch_SCHashTable_t {
    // inlining of macro ABC_CALLOC
    var pSCHashTable *Fxch_SCHashTable_t  = &Fxch_SCHashTable_t{};
    var nBits int32  = Abc_Base2Log( nEntries + 1 );


    pSCHashTable.pFxchMan = pFxchMan;
    pSCHashTable.SizeMask = (1 << nBits) - 1;
    // inlining of macro ABC_CALLOC
    pSCHashTable.pBins = ((Fxch_SCHashTable_Entry_t *) calloc((size_t)(pSCHashTable.SizeMask + 1), sizeof(Fxch_SCHashTable_Entry_t)));

    return pSCHashTable;
}

func Fxch_SCHashTableDelete(pSCHashTable *Fxch_SCHashTable_t) {
    var i uint32;
    for  i = 0; i <= pSCHashTable.SizeMask; i++  {
        // inlining of macro ABC_FREE
        ((pSCHashTable.pBins[i].vSCData) ? (free((char *) (pSCHashTable.pBins[i].vSCData)), (pSCHashTable.pBins[i].vSCData) = 0) : 0);
    }
    Vec_IntErase( &pSCHashTable.vSubCube0 );
    Vec_IntErase( &pSCHashTable.vSubCube1 );
    // inlining of macro ABC_FREE
    ((pSCHashTable.pBins) ? (free((char *) (pSCHashTable.pBins)), (pSCHashTable.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pSCHashTable) ? (free((char *) (pSCHashTable)), (pSCHashTable) = 0) : 0);
}

static inline Fxch_SCHashTable_Entry_t* Fxch_SCHashTableBin( Fxch_SCHashTable_t* pSCHashTable,
                                                             unsigned int SubCubeID )
{
    return pSCHashTable.pBins + (SubCubeID & pSCHashTable.SizeMask);
}

func Fxch_SCHashTableEntryCompare(pSCHashTable *Fxch_SCHashTable_t, vCubes *Vec_Wec_t, pSCData0 *Fxch_SubCube_t, pSCData1 *Fxch_SubCube_t) int32 {
    var vCube0 *Vec_Int_t = Vec_WecEntry( vCubes, pSCData0.iCube );
    var vCube1 *Vec_Int_t = Vec_WecEntry( vCubes, pSCData1.iCube );

    var pOutputID0 *int32 = Vec_IntEntryP( pSCHashTable.pFxchMan.vOutputID, pSCData0.iCube * pSCHashTable.pFxchMan.nSizeOutputID );
    var pOutputID1 *int32 = Vec_IntEntryP( pSCHashTable.pFxchMan.vOutputID, pSCData1.iCube * pSCHashTable.pFxchMan.nSizeOutputID );
    var i, Result int32;

    if !Vec_IntSize( vCube0 ) ||
         !Vec_IntSize( vCube1 ) ||
         Vec_IntEntry( vCube0, 0 ) != Vec_IntEntry( vCube1, 0 ) ||
         pSCData0.Id != pSCData1.Id {
        return 0; }

    for  i = 0; i < pSCHashTable.pFxchMan.nSizeOutputID && Result == 0; i++  {
        Result = ( pOutputID0[i] & pOutputID1[i] );
    }

    if Result == 0 {
        return 0; }

    Vec_IntClear( &pSCHashTable.vSubCube0 );
    Vec_IntClear( &pSCHashTable.vSubCube1 );

    if pSCData0.iLit1 > 0 && pSCData1.iLit1 > 0 &&
         ( Vec_IntEntry( vCube0, pSCData0.iLit0 ) == Vec_IntEntry( vCube1, pSCData1.iLit0 ) ||
           Vec_IntEntry( vCube0, pSCData0.iLit0 ) == Vec_IntEntry( vCube1, pSCData1.iLit1 ) ||
           Vec_IntEntry( vCube0, pSCData0.iLit1 ) == Vec_IntEntry( vCube1, pSCData1.iLit0 ) ||
           Vec_IntEntry( vCube0, pSCData0.iLit1 ) == Vec_IntEntry( vCube1, pSCData1.iLit1 ) ) {
        return 0; }

    if pSCData0.iLit0 > 0 {
        Vec_IntAppendSkip( &pSCHashTable.vSubCube0, vCube0, pSCData0.iLit0 );
    } else {
        Vec_IntAppend( &pSCHashTable.vSubCube0, vCube0 ); }

    if pSCData1.iLit0 > 0 {
        Vec_IntAppendSkip( &pSCHashTable.vSubCube1, vCube1, pSCData1.iLit0 );
    } else {
        Vec_IntAppend( &pSCHashTable.vSubCube1, vCube1 ); }

    if pSCData0.iLit1 > 0 {
        Vec_IntDrop( &pSCHashTable.vSubCube0,
                       pSCData0.iLit0 < pSCData0.iLit1 ? pSCData0.iLit1 - 1 : pSCData0.iLit1 ); }

    if pSCData1.iLit1 > 0 {
        Vec_IntDrop( &pSCHashTable.vSubCube1,
                       pSCData1.iLit0 < pSCData1.iLit1 ? pSCData1.iLit1 - 1 : pSCData1.iLit1 ); }

    return Vec_IntEqual( &pSCHashTable.vSubCube0, &pSCHashTable.vSubCube1 );
}

func Fxch_SCHashTableInsert(pSCHashTable *Fxch_SCHashTable_t, vCubes *Vec_Wec_t, SubCubeID uint32_t, iCube uint32_t, iLit0 uint32_t, iLit1 uint32_t, fUpdate byte) int32 {
    var iNewEntry int32;
    var Pairs int32;
    var BinID uint32_t;
    var pBin *Fxch_SCHashTable_Entry_t;
    var pNewEntry *Fxch_SubCube_t;
    var iEntry int32;

    MurmurHash3_x86_32( ( void* ) &SubCubeID, sizeof( int ), 0x9747b28c, &BinID);
    pBin = Fxch_SCHashTableBin( pSCHashTable, BinID );

    if pBin.vSCData == nil {
        // inlining of macro ABC_CALLOC
        pBin.vSCData = ((Fxch_SubCube_t *) calloc((size_t)(16), sizeof(Fxch_SubCube_t)));
        pBin.Size = 0;
        pBin.Cap = 16;
    
    } else if pBin.Size == pBin.Cap {
        assert(pBin.Cap <= 0xAAAA, "pBin.Cap <= 0xAAAA");
        pBin.Cap = ( pBin.Cap >> 1 ) * 3;
        // inlining of macro ABC_REALLOC
        
        if pBin.vSCData {
            pBin.vSCData = ((Fxch_SubCube_t *) realloc((char *)(pBin.vSCData), sizeof(Fxch_SubCube_t) * (size_t)(pBin.Cap)));
        } else {
            pBin.vSCData = ((Fxch_SubCube_t *) malloc(sizeof(Fxch_SubCube_t) * (size_t)(pBin.Cap)));
        }
    }

    iNewEntry = pBin.Size++;
    pBin.vSCData[iNewEntry].Id = SubCubeID;
    pBin.vSCData[iNewEntry].iCube = iCube;
    pBin.vSCData[iNewEntry].iLit0 = iLit0;
    pBin.vSCData[iNewEntry].iLit1 = iLit1;
    pSCHashTable.nEntries++;

    if pBin.Size == 1 {
        return 0; }

    pNewEntry = &( pBin.vSCData[iNewEntry] );
    for  iEntry = 0; iEntry < int32(pBin.Size) - 1; iEntry++  {
        var pEntry *Fxch_SubCube_t  = &( pBin.vSCData[iEntry] );
        var pOutputID0 *int32  = Vec_IntEntryP( pSCHashTable.pFxchMan.vOutputID, pEntry.iCube * pSCHashTable.pFxchMan.nSizeOutputID );
        var pOutputID1 *int32  = Vec_IntEntryP( pSCHashTable.pFxchMan.vOutputID, pNewEntry.iCube * pSCHashTable.pFxchMan.nSizeOutputID );
        var Result int32;
        var Base int32;
        var iNewDiv int32 = -1;
        var i int32;
        var z int32;

        if (pEntry.iLit1 != 0 && pNewEntry.iLit1 == 0) || (pEntry.iLit1 == 0 && pNewEntry.iLit1 != 0) {
            continue; }

        if !(Fxch_SCHashTableEntryCompare( pSCHashTable, vCubes, pEntry, pNewEntry ) != 0) {
            continue; }

        if ( pEntry.iLit0 == 0 ) || ( pNewEntry.iLit0 == 0 ) {
            var vCube0 *Vec_Int_t = Fxch_ManGetCube( pSCHashTable.pFxchMan, pEntry.iCube );
            var vCube1 *Vec_Int_t = Fxch_ManGetCube( pSCHashTable.pFxchMan, pNewEntry.iCube );

            if Vec_IntSize( vCube0 ) > Vec_IntSize( vCube1 ) {
                Vec_IntPush( pSCHashTable.pFxchMan.vSCC, pEntry.iCube );
                Vec_IntPush( pSCHashTable.pFxchMan.vSCC, pNewEntry.iCube );
            
            } else {
                Vec_IntPush( pSCHashTable.pFxchMan.vSCC, pNewEntry.iCube );
                Vec_IntPush( pSCHashTable.pFxchMan.vSCC, pEntry.iCube );
            }

            continue;
        }

        Base = Fxch_DivCreate( pSCHashTable.pFxchMan, pEntry, pNewEntry );

        if Base < 0 {
            continue; }

        for  i = 0; i < pSCHashTable.pFxchMan.nSizeOutputID; i++  {
            Result += Fxch_CountOnes( pOutputID0[i] & pOutputID1[i] );
        }

        for  z = 0; z < Result; z++  {
            iNewDiv = Fxch_DivAdd( pSCHashTable.pFxchMan, fUpdate, 0, Base );
        }

        Vec_WecPush( pSCHashTable.pFxchMan.vDivCubePairs, iNewDiv, pEntry.iCube );
        Vec_WecPush( pSCHashTable.pFxchMan.vDivCubePairs, iNewDiv, pNewEntry.iCube );

        Pairs++;
    }

    return Pairs;
}

func Fxch_SCHashTableRemove(pSCHashTable *Fxch_SCHashTable_t, vCubes *Vec_Wec_t, SubCubeID uint32_t, iCube uint32_t, iLit0 uint32_t, iLit1 uint32_t, fUpdate byte) int32 {
    var iEntry int32;
    var Pairs int32;
    var BinID uint32_t;
    var pBin *Fxch_SCHashTable_Entry_t;
    var pEntry *Fxch_SubCube_t;
    var idx int32;

    MurmurHash3_x86_32( ( void* ) &SubCubeID, sizeof( int ), 0x9747b28c, &BinID);

    pBin = Fxch_SCHashTableBin( pSCHashTable, BinID );

    if pBin.Size == 1 {
        pBin.Size = 0;
        return 0;
    }

    for  iEntry = 0; iEntry < int32(pBin.Size); iEntry++  {
        if pBin.vSCData[iEntry].iCube == iCube {
            break; }
    }

    assert(( iEntry != int32(pBin.Size) ) && ( pBin.Size != 0 ), "( iEntry != int32(pBin.Size) ) && ( pBin.Size != 0 )");

    pEntry = &( pBin.vSCData[iEntry] );
    for  idx = 0; idx < int32(pBin.Size); idx++  {
    if idx != iEntry {
        var Base int32;
        var iDiv int32 = -1;

        var i, z, iCube0, iCube1 int32;

        var pNextEntry *Fxch_SubCube_t  = &( pBin.vSCData[idx] );
        var vDivCubePairs *Vec_Int_t;
        var pOutputID0 *int32  = Vec_IntEntryP( pSCHashTable.pFxchMan.vOutputID, pEntry.iCube * pSCHashTable.pFxchMan.nSizeOutputID );
        var pOutputID1 *int32  = Vec_IntEntryP( pSCHashTable.pFxchMan.vOutputID, pNextEntry.iCube * pSCHashTable.pFxchMan.nSizeOutputID );
        var Result int32;

        if (pEntry.iLit1 != 0 && pNextEntry.iLit1 == 0) || (pEntry.iLit1 == 0 && pNextEntry.iLit1 != 0) {
            continue; }

        if !Fxch_SCHashTableEntryCompare( pSCHashTable, vCubes, pEntry, pNextEntry )
             || pEntry.iLit0 == 0
             || pNextEntry.iLit0 == 0 {
            continue; }

        Base = Fxch_DivCreate( pSCHashTable.pFxchMan, pNextEntry, pEntry );

        if Base < 0 {
            continue; }

        for  i = 0; i < pSCHashTable.pFxchMan.nSizeOutputID; i++  {
            Result += Fxch_CountOnes( pOutputID0[i] & pOutputID1[i] );
        }

        for  z = 0; z < Result; z++  {
            iDiv = Fxch_DivRemove( pSCHashTable.pFxchMan, fUpdate, 0, Base );
        }

        vDivCubePairs = Vec_WecEntry( pSCHashTable.pFxchMan.vDivCubePairs, iDiv );
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vDivCubePairs)) && (((iCube0) = Vec_IntEntry(vDivCubePairs, i)), 1) && (((iCube1) = Vec_IntEntry(vDivCubePairs, i+1)), 1); i += 2  {
            if ( iCube0 == int32(pNextEntry.iCube) && iCube1 == int32(pEntry.iCube) )  ||
                 ( iCube0 == int32(pEntry.iCube) && iCube1 == int32(pNextEntry.iCube) ) {
                Vec_IntDrop( vDivCubePairs, i+1 );
                Vec_IntDrop( vDivCubePairs, i );
            }
        }
        if Vec_IntSize( vDivCubePairs ) == 0 {
            Vec_IntErase( vDivCubePairs ); }

        Pairs++;
    }
    }

    memmove(pBin.vSCData + iEntry, pBin.vSCData + iEntry + 1, (pBin.Size - iEntry - 1) * sizeof(*pBin.vSCData));
    pBin.Size -= 1;

    return Pairs;
}

func Fxch_SCHashTableMemory(pHashTable *Fxch_SCHashTable_t) unsigned int {
    unsigned int Memory = sizeof ( Fxch_SCHashTable_t );

    Memory += sizeof( Fxch_SubCube_t ) * ( pHashTable.SizeMask + 1 );

    return Memory;
}

func Fxch_SCHashTablePrint(pHashTable *Fxch_SCHashTable_t) {
    var Memory int32;
    fmt.Printf("SubCube Hash Table at %p\n", ( void* )pHashTable);
    fmt.Printf("%20s %20s\n", "nEntries", "Memory Usage (MB)");

    Memory = Fxch_SCHashTableMemory( pHashTable );
    fmt.Printf("%20d %18.2f\n", pHashTable.nEntries, ( float64(   Memory) / 1048576 ));
}
////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
