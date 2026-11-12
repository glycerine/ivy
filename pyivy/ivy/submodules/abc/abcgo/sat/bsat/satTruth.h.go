package bsat

/**CFile****************************************************************

  FileName    [satTruth.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Truth table computation package.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: satTruth.h,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

type Tru_Man_t = Tru_Man_t_;

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

func Tru_ManEqual(pOut *word, pIn *word, nWords int32) int32 { var w int32; for  w = 0; w < nWords; w++  { if pOut[w]!=pIn[w] {  return 0; } }
 return 1; }
func Tru_ManEqual0(pOut *word, nWords int32) int32 { var w int32; for  w = 0; w < nWords; w++  { if pOut[w]!=0 {       return 0; } }
 return 1; }
func Tru_ManEqual1(pOut *word, nWords int32) int32 { var w int32; for  w = 0; w < nWords; w++  { if pOut[w]!=~(word)0 {return 0; } }
 return 1; }
func Tru_ManCopy(pOut *word, pIn *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] =   pIn[w]; }
  return pOut;        }
func Tru_ManClear(pOut *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] = 0; }
         return pOut;        }
func Tru_ManFill(pOut *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] = ~(word)0; }
  return pOut;        }
func Tru_ManNot(pOut *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] = ~pOut[w]; }
  return pOut;        }
func Tru_ManAnd(pOut *word, pIn *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] &=  pIn[w]; }
  return pOut;        }
func Tru_ManOr(pOut *word, pIn *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] |=  pIn[w]; }
  return pOut;        }
func Tru_ManCopyNot(pOut *word, pIn *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] =  ~pIn[w]; }
  return pOut;        }
func Tru_ManAndNot(pOut *word, pIn *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] &= ~pIn[w]; }
  return pOut;        }
func Tru_ManOrNot(pOut *word, pIn *word, nWords int32) *word { var w int32; for  w = 0; w < nWords; w++  { pOut[w] |= ~pIn[w]; }
  return pOut;        }
func Tru_ManCopyNotCond(pOut *word, pIn *word, nWords int32, fCompl int32) *word { if fCompl != 0 {
    return Tru_ManCopyNot(pOut, pIn, nWords);
} else {
    return Tru_ManCopy(pOut, pIn, nWords);
} }
func Tru_ManAndNotCond(pOut *word, pIn *word, nWords int32, fCompl int32) *word { if fCompl != 0 {
    return Tru_ManAndNot(pOut, pIn, nWords);
} else {
    return Tru_ManAnd(pOut, pIn, nWords);
}  }
func Tru_ManOrNotCond(pOut *word, pIn *word, nWords int32, fCompl int32) *word { if fCompl != 0 {
    return Tru_ManOrNot(pOut, pIn, nWords);
} else {
    return Tru_ManOr(pOut, pIn, nWords);
}   }


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////

// extern Tru_Man_t *   Tru_ManAlloc( int nVars );

// extern void          Tru_ManFree( Tru_Man_t * p );

// extern word *        Tru_ManVar( Tru_Man_t * p, int v );

// extern word *        Tru_ManFunc( Tru_Man_t * p, int h );

// extern int           Tru_ManInsert( Tru_Man_t * p, word * pTruth );

//extern int           Tru_ManHandleMax( Tru_Man_t * p );

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

