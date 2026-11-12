package util

import "fmt"

/**CFile****************************************************************

  FileName    [utilSort.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Merge sort with user-specified cost.]

  Synopsis    [Merge sort with user-specified cost.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Feburary 13, 2011.]

  Revision    [$Id: utilSort.c,v 1.00 2011/02/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//#include <assert.h>

//#include "abc_global.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Merging two lists of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SortMerge(p1Beg *int32, p1End *int32, p2Beg *int32, p2End *int32, pOut *int32) {
    var nEntries int32  = (p1End - p1Beg) + (p2End - p2Beg);
    var pOutBeg *int32  = pOut;
    for  p1Beg < p1End && p2Beg < p2End  {
        if *p1Beg == *p2Beg {
            *pOut++ = *p1Beg++, *pOut++ = *p2Beg++; 
        } else if *p1Beg < *p2Beg {
            *pOut++ = *p1Beg++; 
        } else { // if ( *p1Beg > *p2Beg )
            *pOut++ = *p2Beg++; } 
    }
    for  p1Beg < p1End  {
        *pOut++ = *p1Beg++;
    } 
    for  p2Beg < p2End  {
        *pOut++ = *p2Beg++;
    }
    assert(pOut - pOutBeg == nEntries, "pOut - pOutBeg == nEntries");
}

/**Function*************************************************************

  Synopsis    [Recursive sorting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Sort_rec(pInBeg *int32, pInEnd *int32, pOutBeg *int32) {
    var nSize int32  = pInEnd - pInBeg;
    assert(nSize > 0, "nSize > 0");
    if nSize == 1 {
        return; }
    if nSize == 2 {
         if pInBeg[0] > pInBeg[1] {
             pInBeg[0] ^= pInBeg[1];
             pInBeg[1] ^= pInBeg[0];
             pInBeg[0] ^= pInBeg[1];
         }
    
    } else if nSize < 8 {
        var temp, i, j, best_i int32;
        for  i = 0; i < nSize-1; i++  {
            best_i = i;
            for  j = i+1; j < nSize; j++  {
                if pInBeg[j] < pInBeg[best_i] {
                    best_i = j; }
            }
            temp = pInBeg[i]; 
            pInBeg[i] = pInBeg[best_i]; 
            pInBeg[best_i] = temp;
        }
    
    } else {
        Abc_Sort_rec( pInBeg, pInBeg + nSize/2, pOutBeg );
        Abc_Sort_rec( pInBeg + nSize/2, pInEnd, pOutBeg + nSize/2 );
        Abc_SortMerge( pInBeg, pInBeg + nSize/2, pInBeg + nSize/2, pInEnd, pOutBeg );
        memcpy( pInBeg, pOutBeg, sizeof(int) * nSize );
    }
}

/**Function*************************************************************

  Synopsis    [Returns the sorted array of integers.]

  Description [This procedure is about 10% faster than qsort().]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MergeSort(pInput *int32, nSize int32) {
    var pOutput *int32;
    if nSize < 2 {
        return; }
    pOutput = (int *) malloc( sizeof(int) * nSize );
    Abc_Sort_rec( pInput, pInput + nSize, pOutput );
    free( pOutput );
}



/**Function*************************************************************

  Synopsis    [Merging two lists of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MergeSortCostMerge(p1Beg *int32, p1End *int32, p2Beg *int32, p2End *int32, pOut *int32) {
    var nEntries int32  = (p1End - p1Beg) + (p2End - p2Beg);
    var pOutBeg *int32  = pOut;
    for  p1Beg < p1End && p2Beg < p2End  {
        if p1Beg[1] == p2Beg[1] {
            *pOut++ = *p1Beg++, *pOut++ = *p1Beg++, *pOut++ = *p2Beg++, *pOut++ = *p2Beg++; 
        } else if p1Beg[1] < p2Beg[1] {
            *pOut++ = *p1Beg++, *pOut++ = *p1Beg++; 
        } else { // if ( p1Beg[1] > p2Beg[1] )
            *pOut++ = *p2Beg++, *pOut++ = *p2Beg++; } 
    }
    for  p1Beg < p1End  {
        *pOut++ = *p1Beg++, *pOut++ = *p1Beg++;
    } 
    for  p2Beg < p2End  {
        *pOut++ = *p2Beg++, *pOut++ = *p2Beg++;
    }
    assert(pOut - pOutBeg == nEntries, "pOut - pOutBeg == nEntries");
}

/**Function*************************************************************

  Synopsis    [Recursive sorting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MergeSortCost_rec(pInBeg *int32, pInEnd *int32, pOutBeg *int32) {
    var nSize int32  = (pInEnd - pInBeg)/2;
    assert(nSize > 0, "nSize > 0");
    if nSize == 1 {
        return; }
    if nSize == 2 {
         if pInBeg[1] > pInBeg[3] {
             pInBeg[1] ^= pInBeg[3];
             pInBeg[3] ^= pInBeg[1];
             pInBeg[1] ^= pInBeg[3];
             pInBeg[0] ^= pInBeg[2];
             pInBeg[2] ^= pInBeg[0];
             pInBeg[0] ^= pInBeg[2];
         }
    
    } else if nSize < 8 {
        var temp, i, j, best_i int32;
        for  i = 0; i < nSize-1; i++  {
            best_i = i;
            for  j = i+1; j < nSize; j++  {
                if pInBeg[2*j+1] < pInBeg[2*best_i+1] {
                    best_i = j; }
            }
            temp = pInBeg[2*i]; 
            pInBeg[2*i] = pInBeg[2*best_i]; 
            pInBeg[2*best_i] = temp;
            temp = pInBeg[2*i+1]; 
            pInBeg[2*i+1] = pInBeg[2*best_i+1]; 
            pInBeg[2*best_i+1] = temp;
        }
    
    } else {
        Abc_MergeSortCost_rec( pInBeg, pInBeg + 2*(nSize/2), pOutBeg );
        Abc_MergeSortCost_rec( pInBeg + 2*(nSize/2), pInEnd, pOutBeg + 2*(nSize/2) );
        Abc_MergeSortCostMerge( pInBeg, pInBeg + 2*(nSize/2), pInBeg + 2*(nSize/2), pInEnd, pOutBeg );
        memcpy( pInBeg, pOutBeg, sizeof(int) * 2 * nSize );
    }
}

/**Function*************************************************************

  Synopsis    [Sorting procedure.]

  Description [Returns permutation for the non-decreasing order of costs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MergeSortCost(pCosts *int32, nSize int32) *int32 {
    var i int32;
    var pResult *int32;
    var pInput *int32;
    var pOutput *int32;
    pResult = (int *) calloc( sizeof(int), nSize );
    if nSize < 2 {
        return pResult; }
    pInput  = (int *) malloc( sizeof(int) * 2 * nSize );
    pOutput = (int *) malloc( sizeof(int) * 2 * nSize );
    for  i = 0; i < nSize; i++  {
        pInput[2*i] = i, pInput[2*i+1] = pCosts[i];
    }
    Abc_MergeSortCost_rec( pInput, pInput + 2*nSize, pOutput );
    for  i = 0; i < nSize; i++  {
        pResult[i] = pInput[2*i];
    }
    free( pOutput );
    free( pInput );
    return pResult;
}


// this implementation uses 3x less memory but is 30% slower due to cache misses

       

                                                                        

                                             

                
               
                

                

                                                                        
                                                                                                          
 
                                                     
                         
                                            
     
                                             
                                                    
                                                 
                                
                                                    
                                
     
                           
                            
                           
                           
                                         
 

                                                                        

                                  

                
               
                

                

                                                                        
                                                                                    
 
                                
                        
                     
               
                     
     
                                                   
          
                                    
                                    
                                    
          
     
                         
     
                               
                                       
         
                       
                                           
                                                               
                               
                              
                                        
                                  
         
     
        
     
                                                                          
                                                                                    
                                                                                                     
                                                       
     
 

                                                                        

                                  

                                                                          
               
                

                

                                                                        
                                                  
 
                               
                                                   
                                 
                      
                    
                      
                                                    
                                                                     
                    
                  
 

      




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SortNumCompare(pNum1 *int32, pNum2 *int32) int32 {
    return *pNum1 - *pNum2;
}

/**Function*************************************************************

  Synopsis    [Testing the sorting procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SortTest() {
    var fUseNew int32;
    var i int32;
    var nSize int32 = 50000000;
    var pArray *int32  = (int *)malloc( sizeof(int) * nSize );
    var pPerm *int32;
    var clk abctime;
    // generate numbers
    srand( 1000 );
    for  i = 0; i < nSize; i++  {
        pArray[i] = rand();
    }

    // try sorting
    if fUseNew != 0 {
        var fUseCost int32  = 1;
        if fUseCost != 0 {
            clk = Abc_Clock();
            pPerm = Abc_MergeSortCost( pArray, nSize );
            Abc_PrintTime( 1, "New sort", Abc_Clock() - clk );
            // check
            for  i = 1; i < nSize; i++  {
                assert(pArray[pPerm[i-1]] <= pArray[pPerm[i]], "pArray[pPerm[i-1]] <= pArray[pPerm[i]]");
            }
            free( pPerm );
        
        } else {
            clk = Abc_Clock();
            Abc_MergeSort( pArray, nSize );
            Abc_PrintTime( 1, "New sort", Abc_Clock() - clk );
            // check
            for  i = 1; i < nSize; i++  {
                assert(pArray[i-1] <= pArray[i], "pArray[i-1] <= pArray[i]");
            }
        }
    
    } else {
        clk = Abc_Clock();
        qsort( (void *)pArray, (size_t)nSize, sizeof(int), (int (*)(void *, void *)) Abc_SortNumCompare );
        Abc_PrintTime( 1, "Old sort", Abc_Clock() - clk );
        // check
        for  i = 1; i < nSize; i++  {
            assert(pArray[i-1] <= pArray[i], "pArray[i-1] <= pArray[i]");
        }
    }

    free( pArray );
}




/**Function*************************************************************

  Synopsis    [QuickSort algorithm as implemented by qsort().]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_QuickSort1CompareInc(p1 *word, p2 *word) int32 {
    if uint32((*p1)) < uint32((*p2)) {
        return -1; }
    if uint32((*p1)) > uint32((*p2)) {
        return 1; }
    return 0;
}
func Abc_QuickSort1CompareDec(p1 *word, p2 *word) int32 {
    if uint32((*p1)) > uint32((*p2)) {
        return -1; }
    if uint32((*p1)) < uint32((*p2)) {
        return 1; }
    return 0;
}
func Abc_QuickSort1(pData *word, nSize int32, fDecrease int32) {
    var i, fVerify int32;
    if fDecrease != 0 {
        qsort( (void *)pData, (size_t)nSize, sizeof(word), (int (*)(void *, void *))Abc_QuickSort1CompareDec );
        if fVerify != 0 {
            for  i = 1; i < nSize; i++  {
                assert(uint32(pData[i-1]) >= uint32(pData[i]), "uint32(pData[i-1]) >= uint32(pData[i])");
            } }
    
    } else {
        qsort( (void *)pData, (size_t)nSize, sizeof(word), (int (*)(void *, void *))Abc_QuickSort1CompareInc );
        if fVerify != 0 {
            for  i = 1; i < nSize; i++  {
                assert(uint32(pData[i-1]) <= uint32(pData[i]), "uint32(pData[i-1]) <= uint32(pData[i])");
            } }
    }
}

/**Function*************************************************************

  Synopsis    [QuickSort algorithm based on 2/3-way partitioning.]

  Description [This code is based on the online presentation 
  "QuickSort is Optimal" by Robert Sedgewick and Jon Bentley.
  http://www.sorting-algorithms.com/static/QuicksortIsOptimal.pdf 

  The first 32-bits of the input data contain values to be compared. 
  The last 32-bits contain the user's data. When sorting is finished, 
  the 64-bit words are ordered in the increasing order of their value ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SelectSortInc(pData *word, nSize int32) {
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if uint32(pData[j]) < uint32(pData[best_i]) {
                best_i = j; }
        }
        // inlining of macro ABC_SWAP
        { var t word  = pData[i]; pData[i] = pData[best_i]; pData[best_i] = t; };
    }
}
func Abc_SelectSortDec(pData *word, nSize int32) {
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if uint32(pData[j]) > uint32(pData[best_i]) {
                best_i = j; }
        }
        // inlining of macro ABC_SWAP
        { var t word  = pData[i]; pData[i] = pData[best_i]; pData[best_i] = t; };
    }
}

func Abc_QuickSort2Inc_rec(pData *word, l int32, r int32) {
    var v word  = pData[r];
    var i int32 = l-1;
    var j int32 = r;
    if l >= r {
        return; }
    assert(l < r, "l < r");
    if r - l < 10 {
        Abc_SelectSortInc( pData + l, r - l + 1 );
        return;
    }
    for {
        for  uint32(pData[++i]) < uint32(v)  {; }
        for  uint32(v) < uint32(pData[--j])  {
            if j == l {
                break; }
        }
        if i >= j {
            break; }
        // inlining of macro ABC_SWAP
        { var t word  = pData[i]; pData[i] = pData[j]; pData[j] = t; };
    }
    // inlining of macro ABC_SWAP
    { var t word  = pData[i]; pData[i] = pData[r]; pData[r] = t; }; 
    Abc_QuickSort2Inc_rec( pData, l, i-1 );
    Abc_QuickSort2Inc_rec( pData, i+1, r );
}
func Abc_QuickSort2Dec_rec(pData *word, l int32, r int32) {
    var v word  = pData[r];
    var i int32 = l-1;
    var j int32 = r;
    if l >= r {
        return; }
    assert(l < r, "l < r");
    if r - l < 10 {
        Abc_SelectSortDec( pData + l, r - l + 1 );
        return;
    }
    for {
        for  uint32(pData[++i]) > uint32(v)  {; }
        for  uint32(v) > uint32(pData[--j])  {
            if j == l {
                break; }
        }
        if i >= j {
            break; }
        // inlining of macro ABC_SWAP
        { var t word  = pData[i]; pData[i] = pData[j]; pData[j] = t; };
    }
    // inlining of macro ABC_SWAP
    { var t word  = pData[i]; pData[i] = pData[r]; pData[r] = t; }; 
    Abc_QuickSort2Dec_rec( pData, l, i-1 );
    Abc_QuickSort2Dec_rec( pData, i+1, r );
}

func Abc_QuickSort3Inc_rec(pData *word, l int32, r int32) {
    var v word  = pData[r];
    var k int32;
    var i int32 = l-1;
    var j int32 = r;
    var p int32 = l-1;
    var q int32 = r;
    if l >= r {
        return; }
    assert(l < r, "l < r");
    if r - l < 10 {
        Abc_SelectSortInc( pData + l, r - l + 1 );
        return;
    }
    for {
        for  uint32(pData[++i]) < uint32(v)  {; }
        for  uint32(v) < uint32(pData[--j])  {
            if j == l {
                break; }
        }
        if i >= j {
            break; }
        // inlining of macro ABC_SWAP
        { var t word  = pData[i]; pData[i] = pData[j]; pData[j] = t; };
        // inlining of macro ABC_SWAP
        if uint32(pData[i]) == uint32(v) { p++; { var t word  = pData[p]; pData[p] = pData[i]; pData[i] = t; }; }
        // inlining of macro ABC_SWAP
        if uint32(v) == uint32(pData[j]) { q--; { var t word  = pData[j]; pData[j] = pData[q]; pData[q] = t; }; }
    }
    // inlining of macro ABC_SWAP
    { var t word  = pData[i]; pData[i] = pData[r]; pData[r] = t; }; 
    j = i-1; i = i+1;
    for  k = l; k < p; k, j = k+1, j-1  { 
        // inlining of macro ABC_SWAP
         var t word  = pData[k]; pData[k] = pData[j]; pData[j] = t; };
    for  k = r-1; k > q; k, i = k-1, i+1  { 
        // inlining of macro ABC_SWAP
         var t word  = pData[i]; pData[i] = pData[k]; pData[k] = t; };
    Abc_QuickSort3Inc_rec( pData, l, j );
    Abc_QuickSort3Inc_rec( pData, i, r );
}
func Abc_QuickSort3Dec_rec(pData *word, l int32, r int32) {
    var v word  = pData[r];
    var k int32;
    var i int32 = l-1;
    var j int32 = r;
    var p int32 = l-1;
    var q int32 = r;
    if l >= r {
        return; }
    assert(l < r, "l < r");
    if r - l < 10 {
        Abc_SelectSortDec( pData + l, r - l + 1 );
        return;
    }
    for {
        for  uint32(pData[++i]) > uint32(v)  {; }
        for  uint32(v) > uint32(pData[--j])  {
            if j == l {
                break; }
        }
        if i >= j {
            break; }
        // inlining of macro ABC_SWAP
        { var t word  = pData[i]; pData[i] = pData[j]; pData[j] = t; };
        // inlining of macro ABC_SWAP
        if uint32(pData[i]) == uint32(v) { p++; { var t word  = pData[p]; pData[p] = pData[i]; pData[i] = t; }; }
        // inlining of macro ABC_SWAP
        if uint32(v) == uint32(pData[j]) { q--; { var t word  = pData[j]; pData[j] = pData[q]; pData[q] = t; }; }
    }
    // inlining of macro ABC_SWAP
    { var t word  = pData[i]; pData[i] = pData[r]; pData[r] = t; }; 
    j = i-1; i = i+1;
    for  k = l; k < p; k, j = k+1, j-1  { 
        // inlining of macro ABC_SWAP
         var t word  = pData[k]; pData[k] = pData[j]; pData[j] = t; };
    for  k = r-1; k > q; k, i = k-1, i+1  { 
        // inlining of macro ABC_SWAP
         var t word  = pData[i]; pData[i] = pData[k]; pData[k] = t; };
    Abc_QuickSort3Dec_rec( pData, l, j );
    Abc_QuickSort3Dec_rec( pData, i, r );
}

func Abc_QuickSort2(pData *word, nSize int32, fDecrease int32) {
    var i, fVerify int32;
    if fDecrease != 0 {
        Abc_QuickSort2Dec_rec( pData, 0, nSize - 1 );
        if fVerify != 0 {
            for  i = 1; i < nSize; i++  {
                assert(uint32(pData[i-1]) >= uint32(pData[i]), "uint32(pData[i-1]) >= uint32(pData[i])");
            } }
    
    } else {
        Abc_QuickSort2Inc_rec( pData, 0, nSize - 1 );
        if fVerify != 0 {
            for  i = 1; i < nSize; i++  {
                assert(uint32(pData[i-1]) <= uint32(pData[i]), "uint32(pData[i-1]) <= uint32(pData[i])");
            } }
    }
}
func Abc_QuickSort3(pData *word, nSize int32, fDecrease int32) {
    var i int32;
    var fVerify int32 = 1;
    if fDecrease != 0 {
        Abc_QuickSort2Dec_rec( pData, 0, nSize - 1 );
        if fVerify != 0 {
            for  i = 1; i < nSize; i++  {
                assert(uint32(pData[i-1]) >= uint32(pData[i]), "uint32(pData[i-1]) >= uint32(pData[i])");
            } }
    
    } else {
        Abc_QuickSort2Inc_rec( pData, 0, nSize - 1 );
        if fVerify != 0 {
            for  i = 1; i < nSize; i++  {
                assert(uint32(pData[i-1]) <= uint32(pData[i]), "uint32(pData[i-1]) <= uint32(pData[i])");
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Wrapper around QuickSort to sort entries based on cost.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_QuickSortCostData(pCosts *int32, nSize int32, fDecrease int32, pData *word, pResult *int32) {
    var i int32;
    for  i = 0; i < nSize; i++  {
        pData[i] = ((word)i << 32) | pCosts[i];
    }
    Abc_QuickSort3( pData, nSize, fDecrease );
    for  i = 0; i < nSize; i++  {
        pResult[i] = int32((pData[i] >> 32));
    }
}
func Abc_QuickSortCost(pCosts *int32, nSize int32, fDecrease int32) *int32 {
    // inlining of macro ABC_ALLOC
    var pData *word  = ((word *) malloc(sizeof(word) * (size_t)(nSize)));
    // inlining of macro ABC_ALLOC
    var pResult *int32  = ((int *) malloc(sizeof(int) * (size_t)(nSize)));
    Abc_QuickSortCostData( pCosts, nSize, fDecrease, pData, pResult );
    // inlining of macro ABC_FREE
    ((pData) ? (free((char *) (pData)), (pData) = 0) : 0);
    return pResult;
}

//        extern void Abc_QuickSortTest();
//        Abc_QuickSortTest();

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_QuickSortTest() {
    var nSize int32  = 1000000;
    var fVerbose int32;
    var pData1, pData2 *word;
    var i int32;
    var clk abctime  = Abc_Clock();
    // generate numbers
    // inlining of macro ABC_ALLOC
    pData1 = ((word *) malloc(sizeof(word) * (size_t)(nSize)));
    // inlining of macro ABC_ALLOC
    pData2 = ((word *) malloc(sizeof(word) * (size_t)(nSize)));
    srand( 1111 );
    for  i = 0; i < nSize; i++  {
        pData2[i] = pData1[i] = ((word)i << 32) | rand();
    }
    Abc_PrintTime( 1, "Prepare ", Abc_Clock() - clk );
    // perform sorting
    clk = Abc_Clock();
    Abc_QuickSort3( pData1, nSize, 1 );
    Abc_PrintTime( 1, "Sort new", Abc_Clock() - clk );
    // print the result
    if fVerbose != 0 {
        for  i = 0; i < nSize; i++  {
            fmt.Printf("(%d,%d) ", int32((pData1[i] >> 32)), int32(pData1[i]));
        }
        fmt.Printf("\n");
    }
    // create new numbers
    clk = Abc_Clock();
    Abc_QuickSort1( pData2, nSize, 1 );
    Abc_PrintTime( 1, "Sort old", Abc_Clock() - clk );
    // print the result
    if fVerbose != 0 {
        for  i = 0; i < nSize; i++  {
            fmt.Printf("(%d,%d) ", int32((pData2[i] >> 32)), int32(pData2[i]));
        }
        fmt.Printf("\n");
    }
    // inlining of macro ABC_FREE
    ((pData1) ? (free((char *) (pData1)), (pData1) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pData2) ? (free((char *) (pData2)), (pData2) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

// Creates a sequence of random numbers.
// http://www.codeproject.com/KB/recipes/SimpleRNG.aspx

const NUMBER1 uint32 =  3716960521
const NUMBER2 uint32 =  2174103536

func Abc_Random(fReset int32) uint32 {
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
func Abc_RandomW(fReset int32) word {
    return ((word)Abc_Random(fReset) << 32) | ((word)Abc_Random(fReset) << 0);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

