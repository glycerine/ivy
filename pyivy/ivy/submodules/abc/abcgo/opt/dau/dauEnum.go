package dau

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [dauEnum.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Enumeration of decompositions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauEnum.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"

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
var spStatic_Dau_EnumLift_pBuffer [64]byte;
func Dau_EnumLift(pName string, Shift int32) string {
    
    var pTemp string;
    for  pTemp = spStatic_Dau_EnumLift_pBuffer; *pName; pTemp, pName = pTemp+1, pName+1  {
        *pTemp = (*pName >= 'a' && *pName <= 'z') ? *pName + Shift : *pName;
    }
    *pTemp = 0;
    return spStatic_Dau_EnumLift_pBuffer;
}
var spStatic_Dau_EnumLift2_pBuffer [64]byte;
func Dau_EnumLift2(pName string, Shift int32) string {
    
    var pTemp string;
    for  pTemp = spStatic_Dau_EnumLift2_pBuffer; *pName; pTemp, pName = pTemp+1, pName+1  {
        *pTemp = (*pName >= 'a' && *pName <= 'z') ? *pName + Shift : *pName;
    }
    *pTemp = 0;
    return spStatic_Dau_EnumLift2_pBuffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dau_EnumCombineTwo_pBuffer string // was [256]byte in C

func Dau_EnumCombineTwo(vOne *Vec_Ptr_t, fStar int32, fXor int32, pName1 string, pName2 string, Shift2 int32, fCompl1 int32, fCompl2 int32) {
    
    pName2 = Dau_EnumLift( pName2, Shift2 );
    spStatic_Dau_EnumCombineTwo_pBuffer = fmt.Sprintf("%s%c%s%s%s%s%c", fStar?"*":"", fXor?'[':'(', fCompl1?"!":"", pName1[0] == '*' ? pName1 + 1 : pName1, fCompl2?"!":"", pName2[0] == '*' ? pName2 + 1 : pName2, fXor?']':')');
//    printf( "%s ", pBuffer );
    Vec_PtrPush( vOne, Abc_UtilStrsav(spStatic_Dau_EnumCombineTwo_pBuffer) );
}
var spStatic_Dau_EnumCombineThree_pBuffer string // was [256]byte in C

func Dau_EnumCombineThree(vOne *Vec_Ptr_t, fStar int32, pNameC string, pName1 string, pName2 string, Shift1 int32, Shift2 int32, fComplC int32, fCompl1 int32, fCompl2 int32) {
    
    pName1 = Dau_EnumLift( pName1, Shift1 );
    pName2 = Dau_EnumLift2( pName2, Shift2 );
    spStatic_Dau_EnumCombineThree_pBuffer = fmt.Sprintf("%s%c%s%s%s%s%s%s%c", fStar?"*":"", '<', fComplC?"!":"", pNameC[0] == '*' ? pNameC + 1 : pNameC, fCompl1?"!":"", pName1[0] == '*' ? pName1 + 1 : pName1, fCompl2?"!":"", pName2[0] == '*' ? pName2 + 1 : pName2, '>');
//    printf( "%s ", pBuffer );
    Vec_PtrPush( vOne, Abc_UtilStrsav(spStatic_Dau_EnumCombineThree_pBuffer) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_EnumTestDump(vSets *Vec_Ptr_t, pFileName string) { 
    var pFile *os.File;
    var vOne *Vec_Ptr_t;
    var pName string;
    var v, k int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        return; }
    // inlining of macro Vec_PtrForEachEntry
    for  v = 0; (v < Vec_PtrSize(vSets)) && (((vOne) = (Vec_Ptr_t *)Vec_PtrEntry(vSets, v)), 1); v++  {
        fmt.Fprintf(pFile, "VARIABLE NUMBER %d:\n", v);
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vOne)) && (((pName) = (char *)Vec_PtrEntry(vOne, k)), 1); k++  {
            fmt.Fprintf(pFile, "%s\n", pName);
        }
    }
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_EnumTest() {
    var v int32;
    var k int32;
    var nVarMax int32 = 10;
    var vSets *Vec_Ptr_t;
    var vOne *Vec_Ptr_t;
    var pName string;
    // 0 vars
    vSets = Vec_PtrAlloc( 16 );
    Vec_PtrPush( vSets, Vec_PtrAlloc(0) );
    // 1 vars
    vOne = Vec_PtrAlloc( 1 );
    Vec_PtrPush( vOne, Abc_UtilStrsav("*a") );
    Vec_PtrPush( vSets, vOne );
    // 2+ vars
    for  v = 2; v <= nVarMax; v++  {
        var vSetI, vSetJ, vSetK *Vec_Ptr_t;
        var pNameI, pNameJ, pNameK string;
        var i, j, k, i1, j1, k1 int32;
        vOne = Vec_PtrAlloc( 100 );
        for  i = 1; i < v; i++  {
        for  j = i; j < v; j++  {
            if i + j != v {
                continue; }
            vSetI = (Vec_Ptr_t *)Vec_PtrEntry( vSets, i );
            vSetJ = (Vec_Ptr_t *)Vec_PtrEntry( vSets, j );
            // inlining of macro Vec_PtrForEachEntry
            for  i1 = 0; (i1 < Vec_PtrSize(vSetI)) && (((pNameI) = (char *)Vec_PtrEntry(vSetI, i1)), 1); i1++  {
            // inlining of macro Vec_PtrForEachEntry
            for  j1 = 0; (j1 < Vec_PtrSize(vSetJ)) && (((pNameJ) = (char *)Vec_PtrEntry(vSetJ, j1)), 1); j1++  {
                // AND(a,b)
                Dau_EnumCombineTwo( vOne, 0, 0, pNameI, pNameJ, i, 0, 0 );
                // AND(!a,b)
                if pNameI[0] != '*' {
                Dau_EnumCombineTwo( vOne, 0, 0, pNameI, pNameJ, i, 1, 0 ); }
                // AND(a,!b)
                if pNameJ[0] != '*' && !(i == j && i1 == j1) {
                Dau_EnumCombineTwo( vOne, 0, 0, pNameI, pNameJ, i, 0, 1 ); }
                // AND(!a,!b)
                if pNameI[0] != '*' && pNameJ[0] != '*' {
                Dau_EnumCombineTwo( vOne, 0, 0, pNameI, pNameJ, i, 1, 1 ); }
                // XOR(a,b)
                Dau_EnumCombineTwo( vOne, pNameI[0] == '*' || pNameJ[0] == '*', 1, pNameI, pNameJ, i, 0, 0 );
            }
            }
        }
        }
        for  k = 1; k < v; k++  {
        for  i = 1; i < v; i++  {
        for  j = i; j < v; j++  {
            if k + i + j != v {
                continue; }
            vSetK = (Vec_Ptr_t *)Vec_PtrEntry( vSets, k );
            vSetI = (Vec_Ptr_t *)Vec_PtrEntry( vSets, i );
            vSetJ = (Vec_Ptr_t *)Vec_PtrEntry( vSets, j );
            // inlining of macro Vec_PtrForEachEntry
            for  k1 = 0; (k1 < Vec_PtrSize(vSetK)) && (((pNameK) = (char *)Vec_PtrEntry(vSetK, k1)), 1); k1++  {
            // inlining of macro Vec_PtrForEachEntry
            for  i1 = 0; (i1 < Vec_PtrSize(vSetI)) && (((pNameI) = (char *)Vec_PtrEntry(vSetI, i1)), 1); i1++  {
            // inlining of macro Vec_PtrForEachEntry
            for  j1 = 0; (j1 < Vec_PtrSize(vSetJ)) && (((pNameJ) = (char *)Vec_PtrEntry(vSetJ, j1)), 1); j1++  {
                var fStar int32  = pNameI[0] == '*' && pNameJ[0] == '*';

                // MUX(c,a,b)
                Dau_EnumCombineThree( vOne, fStar, pNameK, pNameI, pNameJ, k, k+i, 0, 0, 0 );
                // MUX(c,!a,b)
                if pNameI[0] != '*' {
                Dau_EnumCombineThree( vOne, fStar, pNameK, pNameI, pNameJ, k, k+i, 0, 1, 0 ); }
                // MUX(c,a,!b)
                if pNameJ[0] != '*' && !(i == j && i1 == j1) {
                Dau_EnumCombineThree( vOne, fStar, pNameK, pNameI, pNameJ, k, k+i, 0, 0, 1 ); }

                if pNameK[0] != '*' && !(i == j && i1 == j1) {
                    // MUX(!c,a,b)
                    Dau_EnumCombineThree( vOne, fStar, pNameK, pNameI, pNameJ, k, k+i, 1, 0, 0 );
                    // MUX(!c,!a,b)
                    if pNameI[0] != '*' {
                    Dau_EnumCombineThree( vOne, fStar, pNameK, pNameI, pNameJ, k, k+i, 1, 1, 0 ); }
                    // MUX(!c,a,!b)
                    if pNameJ[0] != '*' {
                    Dau_EnumCombineThree( vOne, fStar, pNameK, pNameI, pNameJ, k, k+i, 1, 0, 1 ); }
                }
            }
            }
            }
        }
        }
        }
        Vec_PtrPush( vSets, vOne );
    }
    Dau_EnumTestDump( vSets, "_npn/npn/dsd10.txt" );

    // inlining of macro Vec_PtrForEachEntry
    for  v = 0; (v < Vec_PtrSize(vSets)) && (((vOne) = (Vec_Ptr_t *)Vec_PtrEntry(vSets, v)), 1); v++  {
        fmt.Printf("VARIABLE NUMBER %d:\n", v);
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vOne)) && (((pName) = (char *)Vec_PtrEntry(vOne, k)), 1); k++  {
            fmt.Printf("%s\n", pName);
        }
        if v == 4 {
            break; }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  v = 0; (v < Vec_PtrSize(vSets)) && (((vOne) = (Vec_Ptr_t *)Vec_PtrEntry(vSets, v)), 1); v++  {
        fmt.Printf("%d=%d ", v, Vec_PtrSize(vOne));
        Vec_PtrFreeFree( vOne );
    }
    Vec_PtrFree( vSets );
    fmt.Printf("\n");
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

