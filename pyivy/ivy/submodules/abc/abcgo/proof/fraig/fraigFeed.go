package fraig

import "fmt"

/**CFile****************************************************************

  FileName    [fraigFeed.c]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Procedures to support the solver feedback.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigFeed.c,v 1.8 2005/07/08 01:01:31 alanmi Exp $]

***********************************************************************/

//#include "fraigInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int   Fraig_FeedBackPrepare( Fraig_Man_t * p, int * pModel, Msat_IntVec_t * vVars );

// static int   Fraig_FeedBackInsert( Fraig_Man_t * p, int nVarsPi );

// static void  Fraig_FeedBackVerify( Fraig_Man_t * p, Fraig_Node_t * pOld, Fraig_Node_t * pNew );


// static void  Fraig_FeedBackCovering( Fraig_Man_t * p, Msat_IntVec_t * vPats );

// static       Fraig_NodeVec_t * Fraig_FeedBackCoveringStart( Fraig_Man_t * pMan );

// static int   Fraig_GetSmallestColumn( int * pHits, int nHits );

// static int   Fraig_GetHittingPattern( unsigned * pSims, int nWords );

// static void  Fraig_CancelCoveredColumns( Fraig_NodeVec_t * vColumns, int * pHits, int iPat );

// static void  Fraig_FeedBackCheckTable( Fraig_Man_t * p );

// static void  Fraig_FeedBackCheckTableF0( Fraig_Man_t * p );

// static void  Fraig_ReallocateSimulationInfo( Fraig_Man_t * p );



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Initializes the feedback information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackInit(p *Fraig_Man_t) {
    p.vCones    = Fraig_NodeVecAlloc( 500 );
    p.vPatsReal = Msat_IntVecAlloc( 1000 ); 
    p.pSimsReal = (unsigned *)Fraig_MemFixedEntryFetch( p.mmSims );
    memset( p.pSimsReal, 0, sizeof(unsigned) * p.nWordsDyna );
    p.pSimsTemp = (unsigned *)Fraig_MemFixedEntryFetch( p.mmSims );
    p.pSimsDiff = (unsigned *)Fraig_MemFixedEntryFetch( p.mmSims );
}

/**Function*************************************************************

  Synopsis    [Processes the feedback from teh solver.]

  Description [Array pModel gives the value of each variable in the SAT 
  solver. Array vVars is the array of integer numbers of variables
  involves in this conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBack(p *Fraig_Man_t, pModel *int32, vVars *Msat_IntVec_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t) {
    var nVarsPi, nWords int32;
    var i int32;
    var clk abctime  = Abc_Clock();

    // get the number of PI vars in the feedback (also sets the PI values)
    nVarsPi = Fraig_FeedBackPrepare( p, pModel, vVars );

    // set the PI values
    nWords = Fraig_FeedBackInsert( p, nVarsPi );
    assert(p.iWordStart + nWords <= p.nWordsDyna, "p.iWordStart + nWords <= p.nWordsDyna");

    // resimulates the words from p->iWordStart to iWordStop
    for  i = 1; i < p.vNodes.nSize; i++  {
        if Fraig_NodeIsAnd(p.vNodes.pArray[i]) {
            Fraig_NodeSimulate( p.vNodes.pArray[i], p.iWordStart, p.iWordStart + nWords, 0 ); }
    }

    if p.fDoSparse {
        Fraig_TableRehashF0( p, 0 ); }

    if !p.fChoicing {
    Fraig_FeedBackVerify( p, pOld, pNew ); }

    // if there is no room left, compress the patterns
    if p.iWordStart + nWords == p.nWordsDyna {
        p.iWordStart = Fraig_FeedBackCompress( p );
    } else {  // otherwise, update the starting word
        p.iWordStart += nWords; }

p.timeFeed += Abc_Clock() - clk;
}

/**Function*************************************************************

  Synopsis    [Get the number and values of the PI variables.]

  Description [Returns the number of PI variables involved in this feedback.
  Fills in the internal presence and value data for the primary inputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackPrepare(p *Fraig_Man_t, pModel *int32, vVars *Msat_IntVec_t) int32 {
    var pNode *Fraig_Node_t;
    var i int32;
    var nVars int32;
    var nVarsPis int32;
    var pVars *int32;

    // clean the presence flag for all PIs
    for  i = 0; i < p.vInputs.nSize; i++  {
        pNode = p.vInputs.pArray[i];
        pNode.fFeedUse = 0;
    }

    // get the variables involved in the feedback
    nVars = Msat_IntVecReadSize(vVars);
    pVars = Msat_IntVecReadArray(vVars);

    // set the values for the present variables
    nVarsPis = 0;
    for  i = 0; i < nVars; i++  {
        pNode = p.vNodes.pArray[ pVars[i] ];
        if !Fraig_NodeIsVar(pNode) {
            continue; }
        // set its value
        pNode.fFeedUse = 1;
        // inlining of macro MSAT_LITSIGN
        pNode.fFeedVal = !((pModel[pVars[i]])&1)     ;
        nVarsPis++;
    }
    return nVarsPis;
}

/**Function*************************************************************

  Synopsis    [Inserts the new simulation patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackInsert(p *Fraig_Man_t, nVarsPi int32) int32 {
    var pNode *Fraig_Node_t;
    var nWords, iPatFlip, nPatFlipLimit, i, w int32;
    var fUseNoPats int32;
    var fUse2Pats int32;

    // get the number of words 
    if fUse2Pats != 0 {
        // inlining of macro FRAIG_NUM_WORDS
        nWords = (((2 * nVarsPi + 1)>>5) + (((2 * nVarsPi + 1)&31) > 0));
    } else if fUseNoPats != 0 {
        nWords = 1;
    } else {
        // inlining of macro FRAIG_NUM_WORDS
        nWords = (((nVarsPi + 1)>>5) + (((nVarsPi + 1)&31) > 0)); }
    // update the number of words if they do not fit into the simulation info
    if nWords > p.nWordsDyna - p.iWordStart {
        nWords = p.nWordsDyna - p.iWordStart; } 
    // determine the bound on the flipping bit
    nPatFlipLimit = nWords * 32 - 2;

    // mark the real pattern
    Msat_IntVecPush( p.vPatsReal, p.iWordStart * 32 ); 
    // record the real pattern
    // inlining of macro Fraig_BitStringSetBit
    ((p.pSimsReal)[(p.iWordStart * 32)>>5] |= (1<<((p.iWordStart * 32) & 31)));

    // set the values at the PIs
    iPatFlip = 1;
    for  i = 0; i < p.vInputs.nSize; i++  {
        pNode = p.vInputs.pArray[i];
        for  w = p.iWordStart; w < p.iWordStart + nWords; w++  {
            if !pNode.fFeedUse {
                pNode.puSimD[w] = FRAIG_RANDOM_UNSIGNED;
            } else if pNode.fFeedVal {
                pNode.puSimD[w] = FRAIG_FULL;
            } else { // if ( !pNode->fFeedVal )
                pNode.puSimD[w] = 0; }
        }

        if fUse2Pats != 0 {
            // flip two patterns
            if pNode.fFeedUse && 2 * iPatFlip < nPatFlipLimit {
                // inlining of macro Fraig_BitStringXorBit
                ((pNode.puSimD + p.iWordStart)[(2 * iPatFlip - 1)>>5] ^= (1<<((2 * iPatFlip - 1) & 31)));
                // inlining of macro Fraig_BitStringXorBit
                ((pNode.puSimD + p.iWordStart)[(2 * iPatFlip)>>5] ^= (1<<((2 * iPatFlip) & 31)));
                // inlining of macro Fraig_BitStringXorBit
                ((pNode.puSimD + p.iWordStart)[(2 * iPatFlip + 1)>>5] ^= (1<<((2 * iPatFlip + 1) & 31)));
                iPatFlip++;
            }
        
        } else if fUseNoPats != 0 {
        
        } else {
            // flip the diagonal
            if pNode.fFeedUse && iPatFlip < nPatFlipLimit {
                // inlining of macro Fraig_BitStringXorBit
                ((pNode.puSimD + p.iWordStart)[(iPatFlip)>>5] ^= (1<<((iPatFlip) & 31)));
                iPatFlip++;
    //            Extra_PrintBinary( stdout, &pNode->puSimD, 45 ); printf( "\n" );
            }
        }
        // clean the use mask
        pNode.fFeedUse = 0;

        // add the info to the D hash value of the PIs
        for  w = p.iWordStart; w < p.iWordStart + nWords; w++  {
            pNode.uHashD ^= pNode.puSimD[w] * s_FraigPrimes[w];
        }

    }
    return nWords;
}


/**Function*************************************************************

  Synopsis    [Checks that the SAT solver pattern indeed distinquishes the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackVerify(p *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t) {
    var fValue1, fValue2, iPat int32;
    iPat   = Msat_IntVecReadEntry( p.vPatsReal, Msat_IntVecReadSize(p.vPatsReal)-1 );
    // inlining of macro Fraig_BitStringHasBit
    fValue1 = ((((pOld.puSimD)[(iPat)>>5]  & (1<<((iPat) & 31))) > 0));
    // inlining of macro Fraig_BitStringHasBit
    fValue2 = ((((pNew.puSimD)[(iPat)>>5]  & (1<<((iPat) & 31))) > 0));
/*
Fraig_PrintNode( p, pOld );
printf( "\n" );
Fraig_PrintNode( p, pNew );
printf( "\n" );
*/
//    assert( fValue1 != fValue2 );
}

/**Function*************************************************************

  Synopsis    [Compress the simulation patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackCompress(p *Fraig_Man_t) int32 {
    var pSims *uint32;
    var uHash uint32;
    var i int32;
    var w int32;
    var t int32;
    var nPats int32;
    var pPats *int32;
    var fPerformChecks int32  = (p.nBTLimit == -1);

    // solve the covering problem
    if fPerformChecks != 0 {
        Fraig_FeedBackCheckTable( p );
        if p.fDoSparse { 
            Fraig_FeedBackCheckTableF0( p ); }
    }

    // solve the covering problem
    Fraig_FeedBackCovering( p, p.vPatsReal );


    // get the number of additional patterns
    nPats = Msat_IntVecReadSize( p.vPatsReal );
    pPats = Msat_IntVecReadArray( p.vPatsReal );
    // get the new starting word
    // inlining of macro FRAIG_NUM_WORDS
    p.iWordStart = (((p.iPatsPerm + nPats)>>5) + (((p.iPatsPerm + nPats)&31) > 0));

    // set the simulation info for the PIs
    for  i = 0; i < p.vInputs.nSize; i++  {
        // get hold of the simulation info for this PI
        pSims = p.vInputs.pArray[i].puSimD;
        // clean the storage for the new patterns
        for  w = p.iWordPerm; w < p.iWordStart; w++  {
            p.pSimsTemp[w] = 0;
        }
        // set the patterns
        for  t = 0; t < nPats; t++  {
            // inlining of macro Fraig_BitStringHasBit
            if  ((pSims)[(pPats[t])>>5]  & (1<<((pPats[t]) & 31))) > 0 {
                // check if this pattern falls into temporary storage
                if p.iPatsPerm + t < p.iWordPerm * 32 {
                    // inlining of macro Fraig_BitStringSetBit
                    ((pSims)[(p.iPatsPerm + t)>>5] |= (1<<((p.iPatsPerm + t) & 31)));
                } else {
                    // inlining of macro Fraig_BitStringSetBit
                    ((p.pSimsTemp)[(p.iPatsPerm + t)>>5] |= (1<<((p.iPatsPerm + t) & 31))); }
            }
        }
        // copy the pattern 
        for  w = p.iWordPerm; w < p.iWordStart; w++  {
            pSims[w] = p.pSimsTemp[w];
        }
        // recompute the hashing info
        uHash = 0;
        for  w = 0; w < p.iWordStart; w++  {
            uHash ^= pSims[w] * s_FraigPrimes[w];
        }
        p.vInputs.pArray[i].uHashD = uHash;
    }

    // update info about the permanently stored patterns
    p.iWordPerm = p.iWordStart;
    p.iPatsPerm += nPats;
    // inlining of macro FRAIG_NUM_WORDS
    assert(p.iWordPerm == (((p.iPatsPerm)>>5) + (((p.iPatsPerm)&31) > 0)), "p.iWordPerm == /*__starting_point_inline_macro:FRAIG_NUM_WORDS__*/(((p.iPatsPerm)>>5) + (((p.iPatsPerm)&31) > 0))");

    // resimulate and recompute the hash values
    for  i = 1; i < p.vNodes.nSize; i++  {
        if Fraig_NodeIsAnd(p.vNodes.pArray[i]) {
            p.vNodes.pArray[i].uHashD = 0;
            Fraig_NodeSimulate( p.vNodes.pArray[i], 0, p.iWordPerm, 0 );
        }
    }

    // double-check that the nodes are still distinguished
    if fPerformChecks != 0 {
        Fraig_FeedBackCheckTable( p ); }

    // rehash the values in the F0 table
    if p.fDoSparse {
        Fraig_TableRehashF0( p, 0 );
        if fPerformChecks != 0 {
            Fraig_FeedBackCheckTableF0( p ); }
    }

    // check if we need to resize the simulation info
    // if less than FRAIG_WORDS_STORE words are left, reallocate simulation info
    if p.iWordPerm + FRAIG_WORDS_STORE > p.nWordsDyna {
        Fraig_ReallocateSimulationInfo( p ); }

    // set the real patterns
    Msat_IntVecClear( p.vPatsReal );
    memset( p.pSimsReal, 0, sizeof(unsigned)*p.nWordsDyna );
    for  w = 0; w < p.iWordPerm; w++  {
        p.pSimsReal[w] = FRAIG_FULL;
    }
    if p.iPatsPerm % 32 > 0 {
        // inlining of macro FRAIG_MASK
        p.pSimsReal[p.iWordPerm-1] = ((~(uint32(0))) >> (32-(p.iPatsPerm % 32))); }
//    printf( "The number of permanent words = %d.\n", p->iWordPerm );
    return p.iWordStart;
}




/**Function*************************************************************

  Synopsis    [Checks the correctness of the functional simulation table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackCovering(p *Fraig_Man_t, vPats *Msat_IntVec_t) {
    var vColumns *Fraig_NodeVec_t;
    var pSims *uint32;
    var pHits *int32;
    var iPat int32;
    var iCol int32;
    var i int32;
    var nOnesTotal, nSolStarting int32;
    var fVeryVerbose int32;

    // collect the pairs to be distinguished
    vColumns = Fraig_FeedBackCoveringStart( p );
    // collect the number of 1s in each simulation vector
    nOnesTotal = 0;
    // inlining of macro ABC_ALLOC
    pHits = ((int *) malloc(sizeof(int) * (size_t)(vColumns.nSize)));
    for  i = 0; i < vColumns.nSize; i++  {
        pSims = (unsigned *)vColumns.pArray[i];
        pHits[i] = Fraig_BitStringCountOnes( pSims, p.iWordStart );
        nOnesTotal += pHits[i];
//        assert( pHits[i] > 0 );
    }
 
    // go through the patterns
    nSolStarting = Msat_IntVecReadSize(vPats);
    for  (iCol = Fraig_GetSmallestColumn( pHits, vColumns.nSize )) != -1  {
        // find the pattern, which hits this column
        iPat = Fraig_GetHittingPattern( (unsigned *)vColumns.pArray[iCol], p.iWordStart );
        // cancel the columns covered by this pattern
        Fraig_CancelCoveredColumns( vColumns, pHits, iPat );
        // save the pattern
        Msat_IntVecPush( vPats, iPat );
    }

    // free the set of columns
    for  i = 0; i < vColumns.nSize; i++  {
        Fraig_MemFixedEntryRecycle( p.mmSims, (char *)vColumns.pArray[i] );
    }

    // print stats related to the covering problem
    if p.fVerbose && (fVeryVerbose != 0) {
        fmt.Printf("%3d\\%3d\\%3d   ", p.nWordsRand, p.nWordsDyna, p.iWordPerm);
        fmt.Printf("Col (pairs) = %5d.  ", vColumns.nSize);
        fmt.Printf("Row (pats) = %5d.  ", p.iWordStart * 32);
        fmt.Printf("Dns = %6.2f %%.  ", vColumns.nSize==0? 0.0 : 100.0 * nOnesTotal / vColumns.nSize / p.iWordStart / 32);
        fmt.Printf("Sol = %3d (%3d).  ", Msat_IntVecReadSize(vPats), nSolStarting);
        fmt.Printf("\n");
    }
    Fraig_NodeVecFree( vColumns );
    pHits = nil;
}


/**Function*************************************************************

  Synopsis    [Checks the correctness of the functional simulation table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackCoveringStart(p *Fraig_Man_t) *Fraig_NodeVec_t {
    var vColumns *Fraig_NodeVec_t;
    var pT *Fraig_HashTable_t  = p.pTableF;
    var pEntF, pEntD *Fraig_Node_t;
    var pSims *uint32;
    var pUnsigned1, pUnsigned2 *uint32;
    var i, k, m, w int32;//, nOnes;

    // start the set of columns
    vColumns = Fraig_NodeVecAlloc( 100 );

    // go through the pairs of nodes to be distinguished
    for  i = 0; i < pT.nBins; i++  {
    // inlining of macro Fraig_TableBinForEachEntryF
    for  pEntF = pT.pBins[i];                                                      (pEntF != nil);                                                             pEntF = pEntF.pNextF  {
        p.vCones.nSize = 0;
        // inlining of macro Fraig_TableBinForEachEntryD
        for  pEntD = pEntF;                                                      (pEntD != nil);                                                             pEntD = pEntD.pNextD  {
            Fraig_NodeVecPush( p.vCones, pEntD );
        }
        if p.vCones.nSize == 1 {
            continue; }
        //////////////////////////////// bug fix by alanmi, September 14, 2006
        if p.vCones.nSize > 20 { 
            continue; }
        ////////////////////////////////

        for  k = 0; k < p.vCones.nSize; k++  {
            for  m = k+1; m < p.vCones.nSize; m++  {
                if !Fraig_CompareSimInfoUnderMask( p.vCones.pArray[k], p.vCones.pArray[m], p.iWordStart, 0, p.pSimsReal ) {
                    continue; }

                // primary simulation patterns (counter-examples) cannot distinguish this pair
                // get memory to store the feasible simulation patterns
                pSims = (unsigned *)Fraig_MemFixedEntryFetch( p.mmSims );
                // find the pattern that distinguish this column, exept the primary ones
                pUnsigned1 = p.vCones.pArray[k].puSimD;
                pUnsigned2 = p.vCones.pArray[m].puSimD;
                for  w = 0; w < p.iWordStart; w++  {
                    pSims[w] = (pUnsigned1[w] ^ pUnsigned2[w]) & ~p.pSimsReal[w];
                }
                // store the pattern
                Fraig_NodeVecPush( vColumns, (Fraig_Node_t *)pSims );
//                nOnes = Fraig_BitStringCountOnes(pSims, p->iWordStart);
//                assert( nOnes > 0 );
            }
        }      
    }
    }

    // if the flag is not set, do not consider sparse nodes in p->pTableF0
    if !p.fDoSparse {
        return vColumns; }

    // recalculate their hash values based on p->pSimsReal
    pT = p.pTableF0;
    for  i = 0; i < pT.nBins; i++  {
    // inlining of macro Fraig_TableBinForEachEntryF
    for  pEntF = pT.pBins[i];                                                      (pEntF != nil);                                                             pEntF = pEntF.pNextF  {
        pSims = pEntF.puSimD;
        pEntF.uHashD = 0;
        for  w = 0; w < p.iWordStart; w++  {
            pEntF.uHashD ^= (pSims[w] & p.pSimsReal[w]) * s_FraigPrimes[w];
        }
    }
    }

    // rehash the table using these values
    Fraig_TableRehashF0( p, 1 );

    // collect the classes of equivalent node pairs
    for  i = 0; i < pT.nBins; i++  {
    // inlining of macro Fraig_TableBinForEachEntryF
    for  pEntF = pT.pBins[i];                                                      (pEntF != nil);                                                             pEntF = pEntF.pNextF  {
        p.vCones.nSize = 0;
        // inlining of macro Fraig_TableBinForEachEntryD
        for  pEntD = pEntF;                                                      (pEntD != nil);                                                             pEntD = pEntD.pNextD  {
            Fraig_NodeVecPush( p.vCones, pEntD );
        }
        if p.vCones.nSize == 1 {
            continue; }

        // primary simulation patterns (counter-examples) cannot distinguish all these pairs
        for  k = 0; k < p.vCones.nSize; k++  {
            for  m = k+1; m < p.vCones.nSize; m++  {
                // get memory to store the feasible simulation patterns
                pSims = (unsigned *)Fraig_MemFixedEntryFetch( p.mmSims );
                // find the patterns that are not distinquished
                pUnsigned1 = p.vCones.pArray[k].puSimD;
                pUnsigned2 = p.vCones.pArray[m].puSimD;
                for  w = 0; w < p.iWordStart; w++  {
                    pSims[w] = (pUnsigned1[w] ^ pUnsigned2[w]) & ~p.pSimsReal[w];
                }
                // store the pattern
                Fraig_NodeVecPush( vColumns, (Fraig_Node_t *)pSims );
//                nOnes = Fraig_BitStringCountOnes(pSims, p->iWordStart);
//                assert( nOnes > 0 );
            }
        }      
    }
    }
    return vColumns;
}

/**Function*************************************************************

  Synopsis    [Selects the column, which has the smallest number of hits.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_GetSmallestColumn(pHits *int32, nHits int32) int32 {
    var i int32;
    var iColMin int32 = -1;
    var nHitsMin int32 = 1000000;
    for  i = 0; i < nHits; i++  {
        // skip covered columns
        if pHits[i] == 0 {
            continue; }
        // take the column if it can only be covered by one pattern
        if pHits[i] == 1 {
            return i; }
        // find the column, which requires the smallest number of patterns
        if nHitsMin > pHits[i] {
            nHitsMin = pHits[i];
            iColMin = i;
        }
    }
    return iColMin;
}

/**Function*************************************************************

  Synopsis    [Select the pattern, which hits this column.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_GetHittingPattern(pSims *uint32, nWords int32) int32 {
    var i, b int32;
    for  i = 0; i < nWords; i++  {
        if pSims[i] == 0 {
            continue; }
        for  b = 0; b < 32; b++  {
            if pSims[i] & (1 << b) {
                return i * 32 + b; }
        }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Cancel covered patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_CancelCoveredColumns(vColumns *Fraig_NodeVec_t, pHits *int32, iPat int32) {
    var pSims *uint32;
    var i int32;
    for  i = 0; i < vColumns.nSize; i++  {
        pSims = (unsigned *)vColumns.pArray[i];
        // inlining of macro Fraig_BitStringHasBit
        if  ((pSims)[(iPat)>>5]  & (1<<((iPat) & 31))) > 0 {
            pHits[i] = 0; }
    }
}


/**Function*************************************************************

  Synopsis    [Checks the correctness of the functional simulation table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackCheckTable(p *Fraig_Man_t) {
    var pT *Fraig_HashTable_t  = p.pTableF;
    var pEntF, pEntD *Fraig_Node_t;
    var i, k, m, nPairs int32;

    nPairs = 0;
    for  i = 0; i < pT.nBins; i++  {
    // inlining of macro Fraig_TableBinForEachEntryF
    for  pEntF = pT.pBins[i];                                                      (pEntF != nil);                                                             pEntF = pEntF.pNextF  {
        p.vCones.nSize = 0;
        // inlining of macro Fraig_TableBinForEachEntryD
        for  pEntD = pEntF;                                                      (pEntD != nil);                                                             pEntD = pEntD.pNextD  {
            Fraig_NodeVecPush( p.vCones, pEntD );
        }
        if p.vCones.nSize == 1 {
            continue; }
        for  k = 0; k < p.vCones.nSize; k++  {
            for  m = k+1; m < p.vCones.nSize; m++  {
                if Fraig_CompareSimInfo( p.vCones.pArray[k], p.vCones.pArray[m], p.iWordStart, 0 ) {
                    fmt.Printf("Nodes %d and %d have the same D simulation info.\n", p.vCones.pArray[k].Num, p.vCones.pArray[m].Num); }
                nPairs++;
            }
        }   
    }
    }
//    printf( "\nThe total of %d node pairs have been verified.\n", nPairs );
}

/**Function*************************************************************

  Synopsis    [Checks the correctness of the functional simulation table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_FeedBackCheckTableF0(p *Fraig_Man_t) {
    var pT *Fraig_HashTable_t  = p.pTableF0;
    var pEntF *Fraig_Node_t;
    var i, k, m, nPairs int32;

    nPairs = 0;
    for  i = 0; i < pT.nBins; i++  {
        p.vCones.nSize = 0;
        // inlining of macro Fraig_TableBinForEachEntryF
        for  pEntF = pT.pBins[i];                                                      (pEntF != nil);                                                             pEntF = pEntF.pNextF  {
            Fraig_NodeVecPush( p.vCones, pEntF );
        }
        if p.vCones.nSize == 1 {
            continue; }
        for  k = 0; k < p.vCones.nSize; k++  {
            for  m = k+1; m < p.vCones.nSize; m++  {
                if Fraig_CompareSimInfo( p.vCones.pArray[k], p.vCones.pArray[m], p.iWordStart, 0 ) {
                    fmt.Printf("Nodes %d and %d have the same D simulation info.\n", p.vCones.pArray[k].Num, p.vCones.pArray[m].Num); }
                nPairs++;
            }
        }   
    }
//    printf( "\nThe total of %d node pairs have been verified.\n", nPairs );
}

/**Function*************************************************************

  Synopsis    [Doubles the size of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ReallocateSimulationInfo(p *Fraig_Man_t) {
    var mmSimsNew *Fraig_MemFixed_t;        // new memory manager for simulation info
    var pNode *Fraig_Node_t;
    var pSimsNew *uint32;
    var uSignOld uint32;
    var i int32;

    // allocate a new memory manager
    p.nWordsDyna *= 2;
    mmSimsNew = Fraig_MemFixedStart( sizeof(unsigned) * (p.nWordsRand + p.nWordsDyna) );

    // set the new data for the constant node
    pNode = p.pConst1;
    pNode.puSimR = (unsigned *)Fraig_MemFixedEntryFetch( mmSimsNew );
    pNode.puSimD = pNode.puSimR + p.nWordsRand;
    memset( pNode.puSimR, 0, sizeof(unsigned) * p.nWordsRand );
    memset( pNode.puSimD, 0, sizeof(unsigned) * p.nWordsDyna );

    // copy the simulation info of the PIs
    for  i = 0; i < p.vInputs.nSize; i++  {
        pNode = p.vInputs.pArray[i];
        // copy the simulation info
        pSimsNew = (unsigned *)Fraig_MemFixedEntryFetch( mmSimsNew );
        memmove( pSimsNew, pNode.puSimR, sizeof(unsigned) * (p.nWordsRand + p.iWordStart) );
        // attach the new info
        pNode.puSimR = pSimsNew;
        pNode.puSimD = pNode.puSimR + p.nWordsRand;
        // signatures remain without changes
    }

    // replace the manager to free up some memory
    Fraig_MemFixedStop( p.mmSims, 0 );
    p.mmSims = mmSimsNew;

    // resimulate the internal nodes (this should lead to the same signatures)
    for  i = 1; i < p.vNodes.nSize; i++  {
        pNode = p.vNodes.pArray[i];
        if !Fraig_NodeIsAnd(pNode) {
            continue; }
        // allocate memory for the simulation info
        pNode.puSimR = (unsigned *)Fraig_MemFixedEntryFetch( mmSimsNew );
        pNode.puSimD = pNode.puSimR + p.nWordsRand;
        // derive random simulation info
        uSignOld = pNode.uHashR;
        pNode.uHashR = 0;
        Fraig_NodeSimulate( pNode, 0, p.nWordsRand, 1 );
        assert(uSignOld == pNode.uHashR, "uSignOld == pNode.uHashR");
        // derive dynamic simulation info
        uSignOld = pNode.uHashD;
        pNode.uHashD = 0;
        Fraig_NodeSimulate( pNode, 0, p.iWordStart, 0 );
        assert(uSignOld == pNode.uHashD, "uSignOld == pNode.uHashD");
    }

    // realloc temporary storage
    p.pSimsReal = (unsigned *)Fraig_MemFixedEntryFetch( mmSimsNew );
    memset( p.pSimsReal, 0, sizeof(unsigned) * p.nWordsDyna );
    p.pSimsTemp = (unsigned *)Fraig_MemFixedEntryFetch( mmSimsNew );
    p.pSimsDiff = (unsigned *)Fraig_MemFixedEntryFetch( mmSimsNew );
}


/**Function*************************************************************

  Synopsis    [Generated trivial counter example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManAllocCounterExample(p *Fraig_Man_t) *int32 {
    var pModel *int32;
    // inlining of macro ABC_ALLOC
    pModel = ((int *) malloc(sizeof(int) * (size_t)(p.vInputs.nSize)));
    memset( pModel, 0, sizeof(int) * p.vInputs.nSize );
    return pModel;
}


/**Function*************************************************************

  Synopsis    [Saves the counter example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManSimulateBitNode_rec(p *Fraig_Man_t, pNode *Fraig_Node_t) int32 {
    var Value0, Value1 int32;
    if Fraig_NodeIsTravIdCurrent( p, pNode ) {
        return pNode.fMark3; }
    Fraig_NodeSetTravIdCurrent( p, pNode );
    // inlining of macro Fraig_Regular
    Value0 = Fraig_ManSimulateBitNode_rec( p, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01))  );
    // inlining of macro Fraig_Regular
    Value1 = Fraig_ManSimulateBitNode_rec( p, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01))  );
    // inlining of macro Fraig_IsComplement
    Value0 ^= ((int32(((ABC_PTRUINT_T) (pNode.p1) & 01))));
    // inlining of macro Fraig_IsComplement
    Value1 ^= ((int32(((ABC_PTRUINT_T) (pNode.p2) & 01))));
    pNode.fMark3 = Value0 & Value1;
    return pNode.fMark3;
}

/**Function*************************************************************

  Synopsis    [Simulates one bit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManSimulateBitNode(p *Fraig_Man_t, pNode *Fraig_Node_t, pModel *int32) int32 {
    var fCompl, RetValue, i int32;
    // set the PI values
    Fraig_ManIncrementTravId( p );
    for  i = 0; i < p.vInputs.nSize; i++  {
        Fraig_NodeSetTravIdCurrent( p, p.vInputs.pArray[i] );
        p.vInputs.pArray[i].fMark3 = pModel[i];
    }
    // perform the traversal
    // inlining of macro Fraig_IsComplement
    fCompl = ((int32(((ABC_PTRUINT_T) (pNode) & 01))));
    // inlining of macro Fraig_Regular
    RetValue = Fraig_ManSimulateBitNode_rec( p, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01))  );
    return fCompl ^ RetValue;
}


/**Function*************************************************************

  Synopsis    [Saves the counter example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManSaveCounterExample(p *Fraig_Man_t, pNode *Fraig_Node_t) *int32 {
    var pModel *int32;
    var iPattern int32;
    var i, fCompl int32;

    // the node can be complemented
    // inlining of macro Fraig_IsComplement
    fCompl = ((int32(((ABC_PTRUINT_T) (pNode) & 01))));
    // because we compare with constant 0, p->pConst1 should also be complemented
    fCompl = !fCompl;

    // derive the model
    pModel = Fraig_ManAllocCounterExample( p );
    // inlining of macro Fraig_Regular
    iPattern = Fraig_FindFirstDiff( p.pConst1, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) , fCompl, p.nWordsRand, 1 );
    if iPattern >= 0 {
        for  i = 0; i < p.vInputs.nSize; i++  {
            // inlining of macro Fraig_BitStringHasBit
            if  ((p.vInputs.pArray[i].puSimR)[(iPattern)>>5]  & (1<<((iPattern) & 31))) > 0 {
                pModel[i] = 1; }
        }
/*
printf( "SAT solver's pattern:\n" );
for ( i = 0; i < p->vInputs->nSize; i++ )
    printf( "%d", pModel[i] );
printf( "\n" );
*/
        assert((Fraig_ManSimulateBitNode( p, pNode, pModel ) != 0), "Fraig_ManSimulateBitNode( p, pNode, pModel )");
        return pModel;
    }
    // inlining of macro Fraig_Regular
    iPattern = Fraig_FindFirstDiff( p.pConst1, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) , fCompl, p.iWordStart, 0 );
    if iPattern >= 0 {
        for  i = 0; i < p.vInputs.nSize; i++  {
            // inlining of macro Fraig_BitStringHasBit
            if  ((p.vInputs.pArray[i].puSimD)[(iPattern)>>5]  & (1<<((iPattern) & 31))) > 0 {
                pModel[i] = 1; }
        }
/*
printf( "SAT solver's pattern:\n" );
for ( i = 0; i < p->vInputs->nSize; i++ )
    printf( "%d", pModel[i] );
printf( "\n" );
*/
        assert((Fraig_ManSimulateBitNode( p, pNode, pModel ) != 0), "Fraig_ManSimulateBitNode( p, pNode, pModel )");
        return pModel;
    }
    pModel = nil;
    return nil;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

