package lucky

import (
	"fmt"
	spTime "time"
)

/**CFile****************************************************************

  FileName    [luckyInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Semi-canonical form computation package.]

  Synopsis    [Internal declarations.]

  Author      [Jake]

  Date        [Started - August 2012]

***********************************************************************/




//#include <stdio.h>
//#include <stdlib.h>
//#include <assert.h>
//#include <string.h>
//#include <math.h>
//#include <time.h>

// comment out this line to run Lucky Code outside of ABC
#define _RUNNING_ABC_

                    

//#include "misc/util/abc_global.h"
//#include "lucky.h"

     

                                  
                                
                                
                              
                               
                
               
              
                                                     
                                                                       
                       

              
 
             
                   
                
                   
                   
                
                    
          

      


//ABC_NAMESPACE_HEADER_START

typedef struct  
{
    int      nVars;
    int      nWords;
    int      nFuncs;
    word **  pFuncs;
}
Abc_TtStore_t;

typedef struct 
{
    int direction; 
    int position;
}
 varInfo;


typedef struct 
{
    varInfo* posArray;
    int* realArray;
    int varN;
    int positionToSwap1;
    int positionToSwap2;
}
 swapInfo;


var spStatic_TimePrint_timeBegin int32;
func TimePrint(Message string) {
    
    var time float64  = 1.0*(Abc_Clock() - spStatic_TimePrint_timeBegin)/int64(spTime.Second) ;
    if Message != nil {
        fmt.Printf("%s = %f sec.\n", Message, time); }
    spStatic_TimePrint_timeBegin = Abc_Clock();
}

func CompareWords(x word, y word) int32 {
    if x > y {
        return 1; }
    if x < y {
        return -1; }
    return 0;
}

func luckyMin(x int32, y int32) int32 { if x < y {
    return x;
} else {
    return y;
} }
func luckyMax(x int32, y int32) int32 { if x < y {
    return y;
} else {
    return x;
} }


// extern  int             memCompare(word* x, word*  y, int nVars);

// extern  int             Kit_TruthWordNum_64bit( int nVars );

// extern  Abc_TtStore_t * setTtStore(char * pFileInput);

// extern  void            Abc_TruthStoreFree( Abc_TtStore_t * p );

// extern  void            Kit_TruthChangePhase_64bit( word * pInOut, int nVars, int iVar );

// extern  void            Kit_TruthNot_64bit(word * pIn, int nVars );

// extern  void            Kit_TruthCopy_64bit( word * pOut, word * pIn, int nVars );

// extern  void            Kit_TruthSwapAdjacentVars_64bit( word * pInOut, int nVars, int iVar );

// extern  int             Kit_TruthCountOnes_64bit( word* pIn, int nVars );

// extern  void            simpleMinimal(word* x, word* pAux,word* minimal, permInfo* pi, int nVars);

// extern  permInfo *      setPermInfoPtr(int var);

// extern  void            freePermInfoPtr(permInfo* x);

// extern  void            Kit_TruthSemiCanonicize_Yasha_simple( word* pInOut, int nVars, int * pStore );

// extern  unsigned        Kit_TruthSemiCanonicize_Yasha( word* pInOut, int nVars, char * pCanonPerm);

// extern  unsigned        Kit_TruthSemiCanonicize_Yasha1( word* pInOut, int nVars, char * pCanonPerm, int * pStore);

// extern  word            luckyCanonicizer_final_fast_6Vars(word InOut, int* pStore, char* pCanonPerm, unsigned* pCanonPhase);

// extern  word            luckyCanonicizer_final_fast_6Vars1(word InOut, int* pStore, char* pCanonPerm, unsigned* pCanonPhase);

// extern  unsigned        adjustInfoAfterSwap(char* pCanonPerm, unsigned uCanonPhase, int iVar, unsigned info);

// extern  void            resetPCanonPermArray_6Vars(char* x);

// extern  void            swap_ij( word* f,int totalVars, int varI, int varJ);



//ABC_NAMESPACE_HEADER_END


