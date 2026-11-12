package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This is implementation of qsort in MiniSat.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func num_cmp1(x *int32, y *int32) int32 { if (*x) < (*y) {
    return -1;
} else {
    if (*x) > (*y) {
        return 1;
    } else {
        return 0;
    }
} }
func num_cmp2(x *int32, y *int32) int32 { return (*x) < (*y); }
func selectionsort(array *int32, size int32, comp func(any, any) int32) {
    var i, j, best_i int32;
    var tmp int32;
    for i = 0; i < size-1; i++{
        best_i = i;
        for j = i+1; j < size; j++{
            if comp(array + j, array + best_i) {
                best_i = j; }
        }
        tmp = array[i]; array[i] = array[best_i]; array[best_i] = tmp;
    }
}
func sort_rec(array *int32, size int32, comp func(any, any) int32) {
    if size <= 15 {
        selectionsort(array, size, comp);
    } else{
        var pivot int32  = array[size/2];
        var tmp int32;
        var i int32  = -1;
        var j int32  = size;
        for {
            do i++; while(comp(array + i, &pivot));
            do j--; while(comp(&pivot, array + j));
            if i >= j { break; }
            tmp = array[i]; array[i] = array[j]; array[j] = tmp;
        }
        sort_rec(array    , i     , comp);
        sort_rec(&array[i], size-i, comp);
    }
}
func minisat_sort(array *int32, size int32, comp func(any, any) int32) {
    sort_rec(array,size,comp);
}


/**Function*************************************************************

  Synopsis    [This is implementation of qsort in MiniSat.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func selectionsort2(array *int32, size int32) {
    var i, j, best_i int32;
    var tmp int32;
    for i = 0; i < size-1; i++{
        best_i = i;
        for j = i+1; j < size; j++{
            if array[j] < array[best_i] {
                best_i = j; }
        }
        tmp = array[i]; array[i] = array[best_i]; array[best_i] = tmp;
    }
}
func sort_rec2(array *int32, size int32) {
    if size <= 15 {
        selectionsort2(array, size);
    } else{
        var pivot int32  = array[size/2];
        var tmp int32;
        var i int32  = -1;
        var j int32  = size;
        for {
            do i++; while(array[i] < pivot);
            do j--; while(pivot < array[j]);
            if i >= j { break; }
            tmp = array[i]; array[i] = array[j]; array[j] = tmp;
        }
        sort_rec2(array    , i     );
        sort_rec2(&array[i], size-i);
    }
}
func minisat_sort2(array *int32, size int32) {
    sort_rec2(array,size);
}

/**Function*************************************************************

  Synopsis    [This is implementation of qsort in MiniSat.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SortGetTest(nSize int32) *int32 {
    var i int32;
    var pArray *int32;
    srand( 0 );
    // inlining of macro ABC_ALLOC
    pArray = ((int *) malloc(sizeof(int) * (size_t)(nSize)));
    for  i = 0; i < nSize; i++  {
        pArray[i] = rand();
    }
    return pArray;
}
func Gia_SortVerifySorted(pArray *int32, nSize int32) {
    var i int32;
    for  i = 1; i < nSize; i++  {
        assert(pArray[i-1] <= pArray[i], "pArray[i-1] <= pArray[i]");
    }
}
func Gia_SortTest() {
    var nSize int32  = 100000000;
    var pArray *int32;
    var clk abctime  = Abc_Clock();

    fmt.Printf("Sorting %d integers\n", nSize);
    pArray = Gia_SortGetTest( nSize );
clk = Abc_Clock();
    qsort( pArray, (size_t)nSize, 4, (int (*)(void *, void *)) num_cmp1 );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("qsort  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    Gia_SortVerifySorted( pArray, nSize );
    // inlining of macro ABC_FREE
    ((pArray) ? (free((char *) (pArray)), (pArray) = 0) : 0);

    pArray = Gia_SortGetTest( nSize );
clk = Abc_Clock();
    minisat_sort( pArray, nSize, (int (*)(void *, void *)) num_cmp2 );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("minisat"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    Gia_SortVerifySorted( pArray, nSize );
    // inlining of macro ABC_FREE
    ((pArray) ? (free((char *) (pArray)), (pArray) = 0) : 0);

    pArray = Gia_SortGetTest( nSize );
clk = Abc_Clock();
    minisat_sort2( pArray, nSize );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("minisat with inlined comparison"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    Gia_SortVerifySorted( pArray, nSize );
    // inlining of macro ABC_FREE
    ((pArray) ? (free((char *) (pArray)), (pArray) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [This is implementation of qsort in MiniSat.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func selectionsort3(array *float32, perm *int32, size int32) {
    var tmpf float32;
    var tmpi int32;
    var i, j, best_i int32;
    for i = 0; i < size-1; i++{
        best_i = i;
        for j = i+1; j < size; j++{
            if array[j] < array[best_i] {
                best_i = j; }
        }
        tmpf = array[i]; array[i] = array[best_i]; array[best_i] = tmpf;
        tmpi = perm[i];  perm[i]  = perm[best_i];  perm[best_i]  = tmpi;
    }
}
func sort_rec3(array *float32, perm *int32, size int32) {
    if size <= 15 {
        selectionsort3(array, perm, size);
    } else{
        var pivot float32  = array[size/2];
        var tmpf float32;
        var tmpi int32;
        var i int32  = -1;
        var j int32  = size;
        for {
            do i++; while(array[i] < pivot);
            do j--; while(pivot < array[j]);
            if i >= j { break; }
            tmpf = array[i]; array[i] = array[j]; array[j] = tmpf;
            tmpi = perm[i];  perm[i]  = perm[j];  perm[j]  = tmpi;
        }
        sort_rec3(array    , perm,     i     );
        sort_rec3(&array[i], &perm[i], size-i);
    }
}
func minisat_sort3(array *float32, perm *int32, size int32) {
    sort_rec3(array, perm, size);
}

/**Function*************************************************************

  Synopsis    [Sorts the array of floating point numbers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SortFloats(pArray *float32, pPerm *int32, nSize int32) *int32 {
    var i int32;
    if pPerm == nil {
        // inlining of macro ABC_ALLOC
        pPerm = ((int *) malloc(sizeof(int) * (size_t)(nSize)));
        for  i = 0; i < nSize; i++  {
            pPerm[i] = i;
        }
    }
    minisat_sort3( pArray, pPerm, nSize );
//    for ( i = 1; i < nSize; i++ )
//        assert( pArray[i-1] <= pArray[i] );
    return pPerm;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

