package lucky

import "fmt"

/**CFile****************************************************************

  FileName    [lucky.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Semi-canonical form computation package.]

  Synopsis    [Truth table minimization procedures.]

  Author      [Jake]

  Date        [Started - August 2012]

***********************************************************************/

//#include "luckyInt.h"

//ABC_NAMESPACE_IMPL_START


func memCompare(x *word, y *word, nVars int32) int32 {
    var i int32;
    for i=Kit_TruthWordNum_64bit( nVars )-1;i>=0;i-- {
        if x[i]==y[i] {
            continue;
        } else if x[i]>y[i] {
            return 1;
        } else {
            return -1; }
    }
    return 0;
}
///////sort Word* a///////////////////////////////////////////////////////////////////////////////////////////////////////
 func compareWords1(a any, b any) int32 {
     if *(word*)a > *(word*)b {
         return 1;
     } else {
         if *(word*)a < *(word*)b {
             return -1;
         } else {
             return 0;
         } }
     
 }
 
 func sortAndUnique1(a *word, p *Abc_TtStore_t) {
     var i int32;
     var count int32 = 1;
     var WordsN int32 = p.nFuncs;
     var tempWord word;
     qsort(a,(size_t)WordsN,sizeof(word),compareWords1);
     tempWord = a[0];
     for i=1;i<WordsN;i++ {  
         if tempWord != a[i] {
             a[count] = a[i];
             tempWord = a[i];
             count++;
         }
     }
     p.nFuncs = count;
}
//////////sort Word** a//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func compareWords2(x *any, y *any) int32 {
    
    if **(word**)x > **(word**)y {
        return 1;
    } else if **(word**)x < **(word**)y {
        return -1;
    } else {
        return 0; }   
    
}
func compareWords(a *any, b *any) int32 {
     if memcmp(*(word**)a,*(word**)b,sizeof(word)*1) > 0 {
         return 1;
     } else {
         if memcmp(*(word**)a,*(word**)b,sizeof(word)*1) < 0 {
             return -1;
         } else {
             return 0;
         } }       
}
func compareWords3(x *any, y *any) int32 {
    return memCompare(*(word**)x, *(word**)y, 16);   
} 
func sortAndUnique(a **word, p *Abc_TtStore_t) {
     var i int32;
     var count int32 = 1;
     var WordsPtrN int32 = p.nFuncs;
     var tempWordPtr *word;
     qsort(a,WordsPtrN,sizeof(word*),(int(*)(void *,void *))compareWords3);
     tempWordPtr = a[0];
     for i=1;i<WordsPtrN;i++ {   
         if memcmp(a[i],tempWordPtr,sizeof(word)*(p.nWords)) != 0 {
             a[count] = a[i];
             tempWordPtr = a[i];
             count++;
         }
     }
     p.nFuncs = count;
}
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

typedef struct
{
     int totalCycles;
     int maxNCycles;
     int minNCycles;
     
}
cycleCtr;
func setCycleCtrPtr() *cycleCtr {
     var x *cycleCtr  = (cycleCtr*) malloc(sizeof(cycleCtr));
     x.totalCycles=0;
     x.maxNCycles=0;
     x.minNCycles=111111111;
     return x;
}
func freeCycleCtr(x *cycleCtr) {
    free(x);
}
func makeArray(p *Abc_TtStore_t) **word {
    var i int32;
    var a **word;
    a = (word**)malloc(sizeof(word*)*(p.nFuncs));
    for i=0;i<p.nFuncs;i++ {
        a[i] = (word*)malloc(sizeof(word)*(p.nWords));
        memcpy(a[i],p.pFuncs[i],sizeof(word)*(p.nWords));

    }
    return a;
}
func freeArray(a **word, p *Abc_TtStore_t) {
    var i int32;
    for i=0;i<p.nFuncs;i++ {
        free(a[i]);
    }
    free(a);
}

func makeArrayB(a **word, nFuncs int32) *word {
    var i int32;
    var b *word;
    b = (word*)malloc(sizeof(word)*(nFuncs));
    for i=0;i<nFuncs;i++ {
        b[i] = a[i][0];
    }

    return b;
}
func freeArrayB(b *word) {
    free(b);
}

////////////////////////////////////////////////////////////////////////////////////////

// if highest bit in F ( all ones min term ) is one => inverse 
// if pInOnt changed(minimized) by function return 1 if not 0
// int minimalInitialFlip_propper(word* pInOut, word* pDuplicat, int  nVars)
// {
//  word oneWord=1;
//  Kit_TruthCopy_64bit( pDuplicat, pInOut, nVars );
//  Kit_TruthNot_64bit( pDuplicat, nVars );
//  if( memCompare(pDuplicat,pInOut,nVars) == -1)
//  {
//      Kit_TruthCopy_64bit(pInOut, pDuplicat, nVars );
//      return 1;
//  }
//  return 0;
// }
// int minimalFlip(word* pInOut, word* pMinimal, word* PDuplicat, int  nVars) 
// {
//  int i;
//  int blockSize = Kit_TruthWordNum_64bit( nVars )*sizeof(word);
//  memcpy(pMinimal, pInOut, blockSize);
//  memcpy(PDuplicat, pInOut, blockSize);
//  for(i=0;i<nVars;i++)
//  {
//      Kit_TruthChangePhase_64bit( pInOut, nVars, i );
//      if( memCompare(pMinimal,pInOut,nVars) == 1)
//          memcpy(pMinimal, pInOut, blockSize);
//      memcpy(pInOut,PDuplicat,blockSize);
//  }
//  memcpy(pInOut,pMinimal,blockSize);
//  if(memCompare(pMinimal,PDuplicat,nVars) == 0)
//      return 0;
//  else
//      return 1;
// }
// int minimalSwap(word* pInOut, word* pMinimal, word* PDuplicat, int  nVars) 
// {
//  int i;  
//  int blockSize = Kit_TruthWordNum_64bit( nVars )*sizeof(word);
//  memcpy(pMinimal, pInOut, blockSize);
//  memcpy(PDuplicat, pInOut, blockSize);
//  for(i=0;i<nVars-1;i++)
//  {
//      Kit_TruthSwapAdjacentVars_64bit( pInOut, nVars, i );
//      if(memCompare(pMinimal,pInOut,nVars) == 1)
//          memcpy(pMinimal, pInOut, blockSize);
//      memcpy(pInOut,PDuplicat,blockSize);
//  }
//  memcpy(pInOut,pMinimal,blockSize);
//  if(memCompare(pMinimal,PDuplicat,nVars) == 0)
//      return 0;
//  else
//      return 1;
// }
// 
// void luckyCanonicizer(word* pInOut, word* pAux, word* pAux1, int  nVars, cycleCtr* cCtr)
// {
//  int counter=1, cycles=0;   
//  assert( nVars <= 16 );
//  while(counter>0 )   //  && cycles < 10 if we wanna limit cycles
//  {
//      counter=0;
//      counter += minimalInitialFlip(pInOut, nVars);
//      counter += minimalFlip(pInOut, pAux, pAux1, nVars);
//      counter += minimalSwap(pInOut, pAux, pAux1, nVars);
//      cCtr->totalCycles++;
//      cycles++;
//  }
//  if(cycles < cCtr->minNCycles)
//      cCtr->minNCycles = cycles;
//  else if(cycles > cCtr->maxNCycles)
//      cCtr->maxNCycles = cycles;
// }
//  runs paralel F and ~F in luckyCanonicizer
// void luckyCanonicizer2(word* pInOut, word* pAux, word* pAux1, word* temp, int  nVars)
// {
//  int nWords = Kit_TruthWordNum_64bit( nVars );
//  int counter=1, nOnes;   
//  assert( nVars <= 16 );
//  nOnes = Kit_TruthCountOnes_64bit(pInOut, nVars);
//  
//     if ( (nOnes*2 == nWords * 32) )
//     { 
//      Kit_TruthCopy_64bit( temp, pInOut, nVars );
//      Kit_TruthNot_64bit( temp, nVars );
//         luckyCanonicizer1_simple(pInOut, pAux, pAux1, nVars);
//      luckyCanonicizer1_simple(temp, pAux, pAux1, nVars);
//      if( memCompare(temp,pInOut,nVars) == -1)        
//          Kit_TruthCopy_64bit(pInOut, temp, nVars );  
//      return;     
//     }
//  while(counter>0 )   //  && cycles < 10 if we wanna limit cycles
//  {
//      counter=0;
//      counter += minimalInitialFlip_propper(pInOut, pAux, nVars);
//      counter += minimalFlip1(pInOut, pAux, pAux1, nVars);
//      counter += minimalSwap1(pInOut, pAux, pAux1, nVars);
//  }
// }
// same as luckyCanonicizer + cycleCtr stutistics 
// void luckyCanonicizer1(word* pInOut, word* pAux, word* pAux1, int  nVars, cycleCtr* cCtr)
// {
//  int counter=1, cycles=0;   
//  assert( nVars <= 16 );
//  while(counter>0 )   //  && cycles < 10 if we wanna limit cycles
//  {
//      counter=0;
//      counter += minimalInitialFlip1(pInOut, nVars);
//      counter += minimalFlip1(pInOut, pAux, pAux1, nVars);
//      counter += minimalSwap1(pInOut, pAux, pAux1, nVars);
//      cCtr->totalCycles++;
//      cycles++;
//  }
//  if(cycles < cCtr->minNCycles)
//      cCtr->minNCycles = cycles;
//  else if(cycles > cCtr->maxNCycles)
//      cCtr->maxNCycles = cycles;
// }
// luckyCanonicizer 

func printCCtrInfo(cCtr *cycleCtr, nFuncs int32) {
    fmt.Printf("maxNCycles = %d\n", cCtr.maxNCycles);
    fmt.Printf("minNCycles = %d\n", cCtr.minNCycles);
    fmt.Printf("average NCycles = %.3f\n", cCtr.totalCycles/float64(nFuncs));
}
////////////////////////////////////////New Faster versoin/////////////////////////////////////////////////////////

// if highest bit in F ( all ones min term ) is one => inverse 
// returns: if pInOnt changed(minimized) by function return 1 if not 0
func minimalInitialFlip1(pInOut *word, nVars int32) int32 {
    var oneWord word =1;
    if (pInOut[Kit_TruthWordNum_64bit( nVars ) -1]>>63) & oneWord {
        Kit_TruthNot_64bit( pInOut, nVars );
        return 1;
    }
    return 0;
}

// compare F with  F1 = (F with changed phase in one of the vars).
// keeps smaller.
// same for all vars in F.
// returns: if pInOnt changed(minimized) by function return 1 if not 0
func minimalFlip1(pInOut *word, pMinimal *word, PDuplicat *word, nVars int32) int32 {
    var i int32;
    var blockSize int32  = Kit_TruthWordNum_64bit( nVars )*sizeof(word);
    memcpy(pMinimal, pInOut, (size_t)blockSize);
    memcpy(PDuplicat, pInOut, (size_t)blockSize);
    Kit_TruthChangePhase_64bit( pInOut, nVars, 0 );
    for i=1;i<nVars;i++ {
        if memCompare(pMinimal,pInOut,nVars) == 1 {
            memcpy(pMinimal, pInOut, (size_t)blockSize);
            Kit_TruthChangePhase_64bit( pInOut, nVars, i );
        
        } else {
            memcpy(pInOut, pMinimal, (size_t)blockSize);
            Kit_TruthChangePhase_64bit( pInOut, nVars, i );
        }
    }
    if memCompare(pMinimal,pInOut,nVars) == -1 {
        memcpy(pInOut, pMinimal, (size_t)blockSize); }
    if memcmp(pInOut,PDuplicat,(size_t)blockSize) == 0 {
        return 0;
    } else {
        return 1; }
}
// compare F with  F1 = (F with swapped two adjacent vars).
// keeps smaller.
// same for all vars in F.
// returns: if pInOnt changed(minimized) by function return 1 if not 0
func minimalSwap1(pInOut *word, pMinimal *word, PDuplicat *word, nVars int32) int32 {
    var i int32;
    var blockSize int32  = Kit_TruthWordNum_64bit( nVars )*sizeof(word);
    memcpy(pMinimal, pInOut, (size_t)blockSize);
    memcpy(PDuplicat, pInOut, (size_t)blockSize);
    Kit_TruthSwapAdjacentVars_64bit( pInOut, nVars, 0 );
    for i=1;i<nVars-1;i++ {
        if memCompare(pMinimal,pInOut,nVars) == 1 {
            memcpy(pMinimal, pInOut, (size_t)blockSize);
            Kit_TruthSwapAdjacentVars_64bit( pInOut, nVars, i );
        
        } else {
            memcpy(pInOut, pMinimal, (size_t)blockSize);
            Kit_TruthSwapAdjacentVars_64bit( pInOut, nVars, i );
        }
    }
    if memCompare(pMinimal,pInOut,nVars) == -1 {
        memcpy(pInOut, pMinimal, (size_t)blockSize); }
    if memcmp(pInOut,PDuplicat,(size_t)blockSize) == 0 {
        return 0;
    } else {
        return 1; }
}

// if highest bit in F ( all ones min term ) is one => inverse 
// returns: if pInOnt changed(minimized) by function return 1 if not 0
func minimalInitialFlip(pInOut *word, nVars int32, p_uCanonPhase *uint32) int32 {
    var oneWord word =1;
    if (pInOut[Kit_TruthWordNum_64bit( nVars ) -1]>>63) & oneWord {
        Kit_TruthNot_64bit( pInOut, nVars );
        *p_uCanonPhase ^= (1 << nVars);
        return 1;
    }
    return 0;
}

// compare F with  F1 = (F with changed phase in one of the vars).
// keeps smaller.
// same for all vars in F.
// returns: if pInOnt changed(minimized) by function return 1 if not 0
func minimalFlip(pInOut *word, pMinimal *word, PDuplicat *word, nVars int32, p_uCanonPhase *uint32) int32 {
    var i int32;
    var minTemp uint32  = *p_uCanonPhase;
    var blockSize int32  = Kit_TruthWordNum_64bit( nVars )*sizeof(word);
    memcpy(pMinimal, pInOut, (size_t)blockSize);
    memcpy(PDuplicat, pInOut, (size_t)blockSize);       //////////////need one more unsigned!!!!!!!!!!!!!
    Kit_TruthChangePhase_64bit( pInOut, nVars, 0 );
    *p_uCanonPhase ^= uint32(1);
    for i=1;i<nVars;i++ {
        if memCompare(pMinimal,pInOut,nVars) == 1 {
            memcpy(pMinimal, pInOut, (size_t)blockSize);
            minTemp = *p_uCanonPhase;
        
        } else {
            memcpy(pInOut, pMinimal, (size_t)blockSize);
            *p_uCanonPhase = minTemp;
        }
        Kit_TruthChangePhase_64bit( pInOut, nVars, i );
        *p_uCanonPhase ^= (1 << i);
    }
    if memCompare(pMinimal,pInOut,nVars) == -1 {
        memcpy(pInOut, pMinimal, (size_t)blockSize);
        *p_uCanonPhase = minTemp;
    }
    if memcmp(pInOut,PDuplicat,(size_t)blockSize) == 0 { 
        return 0;
    } else {
        return 1; }
}

// swaps iVar and iVar+1 elements in pCanonPerm ant p_uCanonPhase
func swapInfoAdjacentVars(iVar int32, pCanonPerm string, p_uCanonPhase *uint32) {
    var Temp byte  = pCanonPerm[iVar];
    pCanonPerm[iVar] = pCanonPerm[iVar+1];
    pCanonPerm[iVar+1] = Temp;
    
    // if the polarity of variables is different, swap them
    if ((*p_uCanonPhase & (1 << iVar)) > 0) != ((*p_uCanonPhase & (1 << (iVar+1))) > 0) {
        *p_uCanonPhase ^= (1 << iVar);
        *p_uCanonPhase ^= (1 << (iVar+1));
    }
            
}


// compare F with  F1 = (F with swapped two adjacent vars).
// keeps smaller.
// same for all vars in F.
// returns: if pInOnt changed(minimized) by function return 1 if not 0


/*
// this version is buggy and is fixed below
int minimalSwap(word* pInOut, word* pMinimal, word* PDuplicat, int  nVars, char * pCanonPerm, char * tempArray, unsigned* p_uCanonPhase) 
{
    int i;  
    int blockSizeWord = Kit_TruthWordNum_64bit( nVars )*sizeof(word);
    int blockSizeChar = nVars *sizeof(char);
    memcpy(pMinimal, pInOut, blockSizeWord);
    memcpy(PDuplicat, pInOut, blockSizeWord);
    memcpy(tempArray, pCanonPerm, blockSizeChar);  
    Kit_TruthSwapAdjacentVars_64bit( pInOut, nVars, 0 );
    swapInfoAdjacentVars(0, pCanonPerm, p_uCanonPhase);
    for(i=1;i<nVars-1;i++)
    {
        if( memCompare(pMinimal,pInOut,nVars) == 1)
        {
            memcpy(pMinimal, pInOut, blockSizeWord);
            memcpy(tempArray, pCanonPerm, blockSizeChar);
        }
        else
        {
            memcpy(pInOut, pMinimal, blockSizeWord);
            memcpy(pCanonPerm, tempArray, blockSizeChar);   
        }
        Kit_TruthSwapAdjacentVars_64bit( pInOut, nVars, i );
        swapInfoAdjacentVars(i, pCanonPerm, p_uCanonPhase);
    }
    if( memCompare(pMinimal,pInOut,nVars) == -1)
    {
        memcpy(pInOut, pMinimal, blockSizeWord);
        memcpy(pCanonPerm, tempArray, blockSizeChar);
    }
    if(memcmp(pInOut,PDuplicat,blockSizeWord) == 0)
        return 0;
    else
        return 1;
}
*/

func minimalSwap(pInOut *word, pMinimal *word, PDuplicat *word, nVars int32, pCanonPerm string, tempArray string, p_uCanonPhase *uint32) int32 {
    var i int32;  
    var blockSizeWord int32  = Kit_TruthWordNum_64bit( nVars )*sizeof(word);
    var blockSizeChar int32  = nVars *sizeof(char);
    var TempuCanonPhase uint32  = *p_uCanonPhase;
    memcpy(pMinimal, pInOut, (size_t)blockSizeWord);
    memcpy(PDuplicat, pInOut, (size_t)blockSizeWord);
    memcpy(tempArray, pCanonPerm, (size_t)blockSizeChar);  
    Kit_TruthSwapAdjacentVars_64bit( pInOut, nVars, 0 );
    swapInfoAdjacentVars(0, pCanonPerm, p_uCanonPhase);
    for i=1;i<nVars-1;i++ {
        if memCompare(pMinimal,pInOut,nVars) == 1 {
            memcpy(pMinimal, pInOut, (size_t)blockSizeWord);
            memcpy(tempArray, pCanonPerm, (size_t)blockSizeChar);
            TempuCanonPhase = *p_uCanonPhase;
            
        
        } else {
            memcpy(pInOut, pMinimal, (size_t)blockSizeWord);
            memcpy(pCanonPerm, tempArray, (size_t)blockSizeChar);
            *p_uCanonPhase = TempuCanonPhase;
        }
        Kit_TruthSwapAdjacentVars_64bit( pInOut, nVars, i );
        swapInfoAdjacentVars(i, pCanonPerm, p_uCanonPhase);
    }
    if memCompare(pMinimal,pInOut,nVars) == -1 {
        memcpy(pInOut, pMinimal, (size_t)blockSizeWord);
        memcpy(pCanonPerm, tempArray, (size_t)blockSizeChar);
        *p_uCanonPhase = TempuCanonPhase;
    }
    if memcmp(pInOut,PDuplicat,(size_t)blockSizeWord) == 0 {
        return 0;
    } else {
        return 1; }
}


//////////////// functions below just for Alan if he want to double check my program//////////////////////////////////
/////////////////You need swap_ij function or analogical one//////////////////////////////////////////////////////////
/*
void swapAndFlip(word* pAfter, int nVars, int iVarInPosition, int jVar, char * pCanonPerm, unsigned* pUCanonPhase)
{
    int Temp;
    swap_ij(pAfter, nVars, iVarInPosition, jVar);
    
    Temp = pCanonPerm[iVarInPosition];
    pCanonPerm[iVarInPosition] = pCanonPerm[jVar];
    pCanonPerm[jVar] = Temp;
    
    if ( ((*pUCanonPhase & (1 << iVarInPosition)) > 0) != ((*pUCanonPhase & (1 << jVar)) > 0) )
    {
        *pUCanonPhase ^= (1 << iVarInPosition);
        *pUCanonPhase ^= (1 << jVar);
    }
    if(*pUCanonPhase>>iVarInPosition & (unsigned)1 == 1)
        Kit_TruthChangePhase_64bit( pAfter, nVars, iVarInPosition );
    
}
int luckyCheck(word* pAfter, word* pBefore, int nVars, char * pCanonPerm, unsigned uCanonPhase)
{
    int i,j;
    char tempChar;
    for(j=0;j<nVars;j++)
    {
        tempChar = 'a'+ j;
        for(i=j;i<nVars;i++)
        {
            if(tempChar != pCanonPerm[i])
                continue;
            swapAndFlip(pAfter , nVars, j, i, pCanonPerm, &uCanonPhase);
            break;
        }
    }
    if(uCanonPhase>>nVars & (unsigned)1 == 1)
        Kit_TruthNot_64bit(pAfter, nVars );
    if(memcmp(pAfter, pBefore, Kit_TruthWordNum_64bit( nVars )*sizeof(word)) == 0)
        return 0;
    else
        return 1;
}
*/
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


func luckyCanonicizer(pInOut *word, pAux *word, pAux1 *word, nVars int32, pCanonPerm string, tempArray string, p_uCanonPhase *uint32) {
    var counter int32 =1;
    assert(nVars <= 16, "nVars <= 16");
    for counter>0  {   //  && cycles < 10 if we wanna limit cycles
    
        counter=0;
        counter += minimalInitialFlip(pInOut, nVars, p_uCanonPhase);
        counter += minimalFlip(pInOut, pAux, pAux1, nVars, p_uCanonPhase);
        counter += minimalSwap(pInOut, pAux, pAux1, nVars, pCanonPerm, tempArray, p_uCanonPhase);   
    }
}
// tries to find minimal F till F at the beginning of the loop is the same as at the end - irreducible 
func luckyCanonicizer1_simple(pInOut *word, pAux *word, pAux1 *word, nVars int32, pCanonPerm string, CanonPhase uint32) uint32 {
    var counter int32 =1;
    assert(nVars <= 16, "nVars <= 16");
    for counter>0  {   //  && cycles < 10 if we wanna limit cycles
    
        counter=0;
        counter += minimalInitialFlip1(pInOut, nVars);
        counter += minimalFlip1(pInOut, pAux, pAux1, nVars);
        counter += minimalSwap1(pInOut, pAux, pAux1, nVars);    
    }
    return CanonPhase;
}

func luckyCanonicizer_final(pInOut *word, pAux *word, pAux1 *word, nVars int32) {
    Kit_TruthSemiCanonicize_Yasha_simple( pInOut, nVars, nil );
    luckyCanonicizer1_simple(pInOut, pAux, pAux1, nVars, nil, 0);
}

// this procedure calls internal canoniziers
// it returns canonical phase (as return value) and canonical permutation (in pCanonPerm)
func Kit_TruthSemiCanonicize_new_internal(pInOut *word, nVars int32, pCanonPerm string) uint32 {
    var pAux, pAux1 [1024]word;
    var tempArray [16]byte;
    var CanonPhase uint32;
    assert(nVars <= 16, "nVars <= 16");
    CanonPhase = Kit_TruthSemiCanonicize_Yasha( pInOut, nVars, pCanonPerm );
    luckyCanonicizer(pInOut, pAux, pAux1, nVars, pCanonPerm, tempArray, &CanonPhase);
    return CanonPhase;
}

// this procedure is translates truth table from 'unsingeds' into 'words' 
func Kit_TruthSemiCanonicize_new(pInOut *uint32, pAux *uint32, nVars int32, pCanonPerm string) uint32 {
    var Result uint32;
    if nVars < 6 {
         var Temp word  = ((word)pInOut[0] << 32) | (word)pInOut[0];
         Result = Kit_TruthSemiCanonicize_new_internal( &Temp, nVars, pCanonPerm );
         pInOut[0] = uint32(Temp);
    
    } else {
         Result = Kit_TruthSemiCanonicize_new_internal( (word *)pInOut, nVars, pCanonPerm ); }
    return Result;
}



// compile main() procedure only if running outside of ABC environment
#ifndef _RUNNING_ABC_

int main () 
{
//  char * pFileInput  = "nonDSDfunc06var1M.txt";
//  char * pFileInput1 = "partDSDfunc06var1M.txt";
//  char * pFileInput2 = "fullDSDfunc06var1M.txt";

//  char * pFileInput  = "nonDSDfunc10var100K.txt";
//  char * pFileInput1 = "partDSDfunc10var100K.txt";
//  char * pFileInput2 = "fullDSDfunc10var100K.txt";

//  char * pFileInput = "partDSDfunc12var100K.txt";
//  char * pFileInput  = "nonDSDfunc12var100K.txt";
//  char * pFileInput1 = "partDSDfunc12var100K.txt";
//  char * pFileInput2 = "fullDSDfunc12var100K.txt";

//  char * pFileInput  = "nonDSDfunc14var10K.txt";
//  char * pFileInput1 = "partDSDfunc14var10K.txt";
//  char * pFileInput2 = "fullDSDfunc14var10K.txt";

    char * pFileInput  = "nonDSDfunc16var10K.txt";
    char * pFileInput1 = "partDSDfunc16var10K.txt";
    char * pFileInput2 = "fullDSDfunc16var10K.txt";

    int i, j, tempNF;
    char** charArray;
    word** a, ** b;
    Abc_TtStore_t* p;
    word * pAux, * pAux1;
    int * pStore;
//  cycleCtr* cCtr;
    charArray = (char**)malloc(sizeof(char*)*3);

    charArray[0] = pFileInput;
    charArray[1] = pFileInput1;
    charArray[2] = pFileInput2;
    for(j=0;j<3;j++)
    {
        p = setTtStore(charArray[j]);   
//      p = setTtStore(pFileInput); 
        a = makeArray(p);
        b = makeArray(p);   
//      cCtr = setCycleCtrPtr();

        pAux = (word*)malloc(sizeof(word)*(p->nWords));
        pAux1 = (word*)malloc(sizeof(word)*(p->nWords));    
        pStore = (int*)malloc(sizeof(int)*(p->nVars));
        printf("In %s Fs at start = %d\n",charArray[j],p->nFuncs);
        
        tempNF = p->nFuncs;

        TimePrint("start");     
        for(i=0;i<p->nFuncs;i++)        
            luckyCanonicizer_final(a[i], pAux, pAux1, p->nVars, pStore);        
        TimePrint("done with A");

        sortAndUnique(a, p);
        printf("F left in A final = %d\n",p->nFuncs);
        freeArray(a,p);
        TimePrint("Done with sort");
        

// delete data-structures   
        free(pAux);
        free(pAux1);    
        free(pStore);
//      freeCycleCtr(cCtr);
        Abc_TruthStoreFree( p );
    }
    return 0;
}

#endif



//ABC_NAMESPACE_IMPL_END
