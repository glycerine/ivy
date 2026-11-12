package bsat

import "fmt"

/**CFile****************************************************************

  FileName    [satProof.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Proof manipulation procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satProof.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/vec/vecSet.h"
//#include "aig/aig/aig.h"
//#include "satTruth.h"

//ABC_NAMESPACE_IMPL_START

/*
    Proof is represented as a vector of records.
    A resolution record is represented by a handle (an offset in this vector).
    A resolution record entry is <size><label><ant1><ant2>...<antN>
    Label is initialized to 0. Root clauses are given by their handles.
    They are marked by bitshifting by 2 bits up and setting the LSB to 1
*/

   
type satset struct {
    learnt uint32; /* C bitfield: 1 bits; review packing and truncation */
    mark uint32; /* C bitfield: 1 bits; review packing and truncation */
    partA uint32; /* C bitfield: 1 bits; review packing and truncation */
    nEnts uint32; /* C bitfield: 29 bits; review packing and truncation */
    Id      int32;
    pEnts [0]int32;
}

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//static inline satset* Proof_ClauseRead  ( Vec_Int_t* p, int h )     { assert( h > 0 );     return satset_read( (veci *)p, h );    }
//static inline satset* Proof_ClauseRead  ( Vec_Int_t* p, int h )     { assert( h > 0 );     return (satset *)Vec_IntEntryP( p, h );}
func Proof_NodeRead(p *Vec_Set_t, h int32) *satset { assert(h > 0, "h > 0");     return (satset*)Vec_SetEntry( p, h );  }
func Proof_NodeWordNum(nEnts int32) int32 { assert(nEnts > 0, "nEnts > 0"); return 1 + ((nEnts + 1) >> 1);         }

func Proof_ClauseSetEnts(p *Vec_Set_t, h int32, nEnts int32) { Proof_NodeRead(p, h).nEnts = nEnts;             }

// iterating through nodes in the proof
// Original macro:
// #define Proof_ForeachClauseVec( pVec, p, pNode, i )          \
//     for ( i = 0; (i < Vec_IntSize(pVec)) && ((pNode) = Proof_ClauseRead(p, Vec_IntEntry(pVec,i))); i++ )

// Original macro:
// #define Proof_ForeachNodeVec( pVec, p, pNode, i )            \
//     for ( i = 0; (i < Vec_IntSize(pVec)) && ((pNode) = Proof_NodeRead(p, Vec_IntEntry(pVec,i))); i++ )

// Original macro:
// #define Proof_ForeachNodeVec1( pVec, p, pNode, i )            \
//     for ( i = 1; (i < Vec_IntSize(pVec)) && ((pNode) = Proof_NodeRead(p, Vec_IntEntry(pVec,i))); i++ )


// iterating through fanins of a proof node
// Original macro:
// #define Proof_NodeForeachFanin( pProof, pNode, pFanin, i )        \
//     for ( i = 0; (i < (int)pNode->nEnts) && (((pFanin) = (pNode->pEnts[i] & 1) ? NULL : Proof_NodeRead(pProof, pNode->pEnts[i] >> 2)), 1); i++ )

/*
#define Proof_NodeForeachLeaf( pClauses, pNode, pLeaf, i )   \
    for ( i = 0; (i < (int)pNode->nEnts) && (((pLeaf) = (pNode->pEnts[i] & 1) ? Proof_ClauseRead(pClauses, pNode->pEnts[i] >> 2) : NULL), 1); i++ )
#define Proof_NodeForeachFaninLeaf( pProof, pClauses, pNode, pFanin, i )    \
    for ( i = 0; (i < (int)pNode->nEnts) && ((pFanin) = (pNode->pEnts[i] & 1) ? Proof_ClauseRead(pClauses, pNode->pEnts[i] >> 2) : Proof_NodeRead(pProof, pNode->pEnts[i] >> 2)); i++ )
*/

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Cleans collected resultion nodes belonging to the proof.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Proof_CleanCollected(vProof *Vec_Set_t, vUsed *Vec_Int_t) {
    var pNode *satset;
    var hNode int32;
    // inlining of macro Proof_ForeachNodeVec
    for  hNode = 0; (hNode < Vec_IntSize(vUsed)) && ((pNode) = Proof_NodeRead(vProof, Vec_IntEntry(vUsed,hNode))); hNode++  {
        pNode.Id = 0;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Proof_CollectUsed_iter(vProof *Vec_Set_t, hNode int32, vUsed *Vec_Int_t, vStack *Vec_Int_t) {
    var pNext *satset;
    var pNode *satset = Proof_NodeRead( vProof, hNode );
    var i int32;
    if pNode.Id {
        return; }
    // start with node
    pNode.Id = 1;
    Vec_IntPush( vStack, hNode << 1 );
    // perform DFS search
    for  Vec_IntSize(vStack)  {
        hNode = Vec_IntPop( vStack );
        if hNode & 1 { // extracted second time
        
            Vec_IntPush( vUsed, hNode >> 1 );
            continue;
        }
        // extracted first time        
        Vec_IntPush( vStack, hNode ^ 1 ); // add second time
        // add its anticedents        ;
        pNode = Proof_NodeRead( vProof, hNode >> 1 );
        // inlining of macro Proof_NodeForeachFanin
        for  i = 0; (i < int32(pNode.nEnts)) && (((pNext) = (pNode.pEnts[i] & 1) ? nil : Proof_NodeRead(vProof, pNode.pEnts[i] >> 2)), 1); i++  {
            if (pNext != nil) && !pNext.Id {
                pNext.Id = 1;
                Vec_IntPush( vStack, (pNode.pEnts[i] >> 2) << 1 ); // add first time
            }
        }
    }
}
func Proof_CollectUsedIter(vProof *Vec_Set_t, vRoots *Vec_Int_t, fSort int32) *Vec_Int_t {
    var fVerify int32;
    var vUsed, vStack *Vec_Int_t;
    var clk abctime  = Abc_Clock();
    var i, Entry, iPrev int32;
    vUsed = Vec_IntAlloc( 1000 );
    vStack = Vec_IntAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRoots)) && (((Entry) = Vec_IntEntry(vRoots, i)), 1); i++  {
        if Entry >= 0 {
            Proof_CollectUsed_iter( vProof, Entry, vUsed, vStack ); }
    }
    Vec_IntFree( vStack );
//    Abc_PrintTime( 1, "Iterative clause collection time", Abc_Clock() - clk );
    clk = Abc_Clock();
    Abc_MergeSort( Vec_IntArray(vUsed), Vec_IntSize(vUsed) );
//    Abc_PrintTime( 1, "Postprocessing with sorting time", Abc_Clock() - clk );
    // verify topological order
    if fVerify != 0 {
        iPrev = 0;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vUsed)) && (((Entry) = Vec_IntEntry(vUsed, i)), 1); i++  {
            if iPrev >= Entry {
                fmt.Printf("Out of topological order!!!\n"); }
            assert(iPrev < Entry, "iPrev < Entry");
            iPrev = Entry;
        }
    }
    return vUsed;
}

/**Function*************************************************************

  Synopsis    [Recursively collects useful proof nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Proof_CollectUsed_rec(vProof *Vec_Set_t, hNode int32, vUsed *Vec_Int_t) {
    var pNext *satset;
    var pNode *satset = Proof_NodeRead( vProof, hNode );
    var i int32;
    if pNode.Id {
        return; }
    pNode.Id = 1;
    // inlining of macro Proof_NodeForeachFanin
    for  i = 0; (i < int32(pNode.nEnts)) && (((pNext) = (pNode.pEnts[i] & 1) ? nil : Proof_NodeRead(vProof, pNode.pEnts[i] >> 2)), 1); i++  {
        if (pNext != nil) && !pNext.Id {
            Proof_CollectUsed_rec( vProof, pNode.pEnts[i] >> 2, vUsed ); }
    }
    Vec_IntPush( vUsed, hNode );
}
func Proof_CollectUsedRec(vProof *Vec_Set_t, vRoots *Vec_Int_t) *Vec_Int_t {
    var vUsed *Vec_Int_t;
    var i, Entry int32;
    vUsed = Vec_IntAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRoots)) && (((Entry) = Vec_IntEntry(vRoots, i)), 1); i++  {
        if Entry >= 0 {
            Proof_CollectUsed_rec( vProof, Entry, vUsed ); }
    }
    return vUsed;
}

/**Function*************************************************************

  Synopsis    [Recursively visits useful proof nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Proof_MarkUsed_rec(vProof *Vec_Set_t, hNode int32) int32 {
    var pNext *satset;
    var pNode *satset = Proof_NodeRead( vProof, hNode );
    var i int32;
    var Counter int32 = 1;
    if pNode.Id {
        return 0; }
    pNode.Id = 1;
    // inlining of macro Proof_NodeForeachFanin
    for  i = 0; (i < int32(pNode.nEnts)) && (((pNext) = (pNode.pEnts[i] & 1) ? nil : Proof_NodeRead(vProof, pNode.pEnts[i] >> 2)), 1); i++  {
        if (pNext != nil) && !pNext.Id {
            Counter += Proof_MarkUsed_rec( vProof, pNode.pEnts[i] >> 2 ); }
    }
    return Counter;
}
func Proof_MarkUsedRec(vProof *Vec_Set_t, vRoots *Vec_Int_t) int32 {
    var i, Entry, Counter int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRoots)) && (((Entry) = Vec_IntEntry(vRoots, i)), 1); i++  {
        if Entry >= 0 {
            Counter += Proof_MarkUsed_rec( vProof, Entry ); }
    }
    return Counter;
}



  
/**Function*************************************************************

  Synopsis    [Checks the validity of the check point.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
void Sat_ProofReduceCheck_rec( Vec_Set_t * vProof, Vec_Int_t * vClauses, satset * pNode, int hClausePivot, Vec_Ptr_t * vVisited )
{
    satset * pFanin;
    int k;
    if ( pNode->Id )
        return;
    pNode->Id = -1;
    Proof_NodeForeachFaninLeaf( vProof, vClauses, pNode, pFanin, k )
        if ( (pNode->pEnts[k] & 1) == 0 ) // proof node
            Sat_ProofReduceCheck_rec( vProof, vClauses, pFanin, hClausePivot, vVisited );
        else // problem clause
            assert( (pNode->pEnts[k] >> 2) < hClausePivot );
    Vec_PtrPush( vVisited, pNode );
}
void Sat_ProofReduceCheckOne( sat_solver2 * s, int iLearnt, Vec_Ptr_t * vVisited )
{
    Vec_Set_t * vProof   = (Vec_Set_t *)&s->Proofs;
    Vec_Int_t * vClauses = (Vec_Int_t *)&s->clauses;
    Vec_Int_t * vRoots   = (Vec_Int_t *)&s->claProofs;
    int hProofNode = Vec_IntEntry( vRoots, iLearnt );
    satset * pNode = Proof_NodeRead( vProof, hProofNode );
    Sat_ProofReduceCheck_rec( vProof, vClauses, pNode, s->hClausePivot, vVisited );
}
void Sat_ProofReduceCheck( sat_solver2 * s )
{
    Vec_Ptr_t * vVisited;
    satset * c;
    int h, i = 1;
    vVisited = Vec_PtrAlloc( 1000 );
    sat_solver_foreach_learnt( s, c, h )
        if ( h < s->hLearntPivot )
            Sat_ProofReduceCheckOne( s, i++, vVisited );
    Vec_PtrForEachEntry( satset *, vVisited, c, i )
        c->Id = 0;
    Vec_PtrFree( vVisited );
}
*/

/**Function*************************************************************

  Synopsis    [Reduces the proof to contain only roots and their children.]

  Description [The result is updated proof and updated roots.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
void Sat_ProofReduce2( sat_solver2 * s )
{
    Vec_Set_t * vProof   = (Vec_Set_t *)&s->Proofs;
    Vec_Int_t * vRoots   = (Vec_Int_t *)&s->claProofs;
    Vec_Int_t * vClauses = (Vec_Int_t *)&s->clauses;

    int fVerbose = 0;
    Vec_Int_t * vUsed;
    satset * pNode, * pFanin, * pPivot;
    int i, k, hTemp;
    abctime clk = Abc_Clock();
    static abctime TimeTotal = 0;

    // collect visited nodes
    vUsed = Proof_CollectUsedIter( vProof, vRoots, 1 );

    // relabel nodes to use smaller space
    Vec_SetShrinkS( vProof, 2 );
    Proof_ForeachNodeVec( vUsed, vProof, pNode, i )
    {
        pNode->Id = Vec_SetAppendS( vProof, 2+pNode->nEnts );
        Proof_NodeForeachFaninLeaf( vProof, vClauses, pNode, pFanin, k )
            if ( (pNode->pEnts[k] & 1) == 0 ) // proof node
                pNode->pEnts[k] = (pFanin->Id << 2) | (pNode->pEnts[k] & 2);
            else // problem clause
                assert( (int*)pFanin >= Vec_IntArray(vClauses) && (int*)pFanin < Vec_IntArray(vClauses)+Vec_IntSize(vClauses) );
    }
    // update roots
    Proof_ForeachNodeVec1( vRoots, vProof, pNode, i )
        Vec_IntWriteEntry( vRoots, i, pNode->Id );
    // determine new pivot
    assert( s->hProofPivot >= 1 && s->hProofPivot <= Vec_SetHandCurrent(vProof) );
    pPivot = Proof_NodeRead( vProof, s->hProofPivot );
    s->hProofPivot = Vec_SetHandCurrentS(vProof);
    // compact the nodes
    Proof_ForeachNodeVec( vUsed, vProof, pNode, i )
    {
        hTemp = pNode->Id; pNode->Id = 0;
        memmove( Vec_SetEntry(vProof, hTemp), pNode, sizeof(word)*Proof_NodeWordNum(pNode->nEnts) );
        if ( pPivot && pPivot <= pNode )
        { 
            s->hProofPivot = hTemp;
            pPivot = NULL;
        }
    }
    Vec_SetWriteEntryNum( vProof, Vec_IntSize(vUsed) );
    Vec_IntFree( vUsed );

    // report the result
    if ( fVerbose )
    {
        printf( "\n" );
        printf( "The proof was reduced from %6.2f MB to %6.2f MB (by %6.2f %%)  ", 
            1.0 * Vec_SetMemory(vProof) / (1<<20), 1.0 * Vec_SetMemoryS(vProof) / (1<<20), 
            100.0 * (Vec_SetMemory(vProof) - Vec_SetMemoryS(vProof)) / Vec_SetMemory(vProof) );
        TimeTotal += Abc_Clock() - clk;
        Abc_PrintTime( 1, "Time", TimeTotal );
    }
    Vec_SetShrink( vProof, Vec_SetHandCurrentS(vProof) );
//    Sat_ProofReduceCheck( s );
} 
*/


func Sat_ProofCheck0(vProof *Vec_Set_t) {
    var pNode, pFanin *satset;
    var i, j, k, nSize int32;
    // inlining of macro Vec_SetForEachEntry
    for  i = 0; i <= vProof.iPage; i++  {                               for  j = 2; j < Vec_SetLimit(vProof.pPages[i]) && ((pNode) = (satset *)(vProof.pPages[i] + (j))); j += nSize  {
        nSize = Vec_SetWordNum( 2 + pNode.nEnts );
        // inlining of macro Proof_NodeForeachFanin
        for  k = 0; (k < int32(pNode.nEnts)) && (((pFanin) = (pNode.pEnts[k] & 1) ? nil : Proof_NodeRead(vProof, pNode.pEnts[k] >> 2)), 1); k++  {
            assert((pNode.pEnts[k] >> 2), "(pNode.pEnts[k] >> 2)");
        }
    }
    }
}

var spStatic_Sat_ProofReduce_TimeTotal abctime;
func Sat_ProofReduce(vProof *Vec_Set_t, pRoots any, hProofPivot int32) int32 {
//    Vec_Set_t * vProof   = (Vec_Set_t *)&s->Proofs;
//    Vec_Int_t * vRoots   = (Vec_Int_t *)&s->claProofs;
    var vRoots *Vec_Int_t    = (Vec_Int_t *)pRoots;
//    Vec_Int_t * vClauses = (Vec_Int_t *)&s->clauses;
    var fVerbose int32;
    var vUsed *Vec_Ptr_t;
    var pNode, pFanin, pPivot *satset;
    var i, j, k, hTemp, nSize int32;
    var clk abctime  = Abc_Clock();
    
    var RetValue int32;
//Sat_ProofCheck0( vProof );

    // collect visited nodes
    nSize = Proof_MarkUsedRec( vProof, vRoots );
    vUsed = Vec_PtrAlloc( nSize );
//Sat_ProofCheck0( vProof );

    // relabel nodes to use smaller space
    Vec_SetShrinkS( vProof, 2 );
    // inlining of macro Vec_SetForEachEntry
    for  i = 0; i <= vProof.iPage; i++  {                               for  j = 2; j < Vec_SetLimit(vProof.pPages[i]) && ((pNode) = (satset *)(vProof.pPages[i] + (j))); j += nSize  {
        nSize = Vec_SetWordNum( 2 + pNode.nEnts );
        if pNode.Id == 0 { 
            continue; }
        pNode.Id = Vec_SetAppendS( vProof, 2 + pNode.nEnts );
        assert(pNode.Id > 0, "pNode.Id > 0");
        Vec_PtrPush( vUsed, pNode );
        // update fanins
        // inlining of macro Proof_NodeForeachFanin
        for  k = 0; (k < int32(pNode.nEnts)) && (((pFanin) = (pNode.pEnts[k] & 1) ? nil : Proof_NodeRead(vProof, pNode.pEnts[k] >> 2)), 1); k++  {
            if (pNode.pEnts[k] & 1) == 0 { // proof node
            
                assert(pFanin.Id > 0, "pFanin.Id > 0");
                pNode.pEnts[k] = (pFanin.Id << 2) | (pNode.pEnts[k] & 2);
            }
        }
//            else // problem clause
//                assert( (int*)pFanin >= Vec_IntArray(vClauses) && (int*)pFanin < Vec_IntArray(vClauses)+Vec_IntSize(vClauses) );
    }
    }
    // update roots
    // inlining of macro Proof_ForeachNodeVec1
    for  i = 1; (i < Vec_IntSize(vRoots)) && ((pNode) = Proof_NodeRead(vProof, Vec_IntEntry(vRoots,i))); i++  {
        Vec_IntWriteEntry( vRoots, i, pNode.Id );
    }
    // determine new pivot
    assert(hProofPivot >= 1 && hProofPivot <= Vec_SetHandCurrent(vProof), "hProofPivot >= 1 && hProofPivot <= Vec_SetHandCurrent(vProof)");
    pPivot = Proof_NodeRead( vProof, hProofPivot );
    RetValue = Vec_SetHandCurrentS(vProof);
    // compact the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vUsed)) && (((pNode) = (satset *)Vec_PtrEntry(vUsed, i)), 1); i++  {
        hTemp = pNode.Id; pNode.Id = 0;
        memmove( Vec_SetEntry(vProof, hTemp), pNode, sizeof(word)*Proof_NodeWordNum(pNode.nEnts) );
        if (pPivot != nil) && pPivot <= pNode {
            RetValue = hTemp;
            pPivot = nil;
        }
    }
    Vec_SetWriteEntryNum( vProof, Vec_PtrSize(vUsed) );
    Vec_PtrFree( vUsed );

    // report the result
    if fVerbose != 0 {
        fmt.Printf("\n");
        fmt.Printf("The proof was reduced from %6.2f MB to %6.2f MB (by %6.2f %%)  ", 1.0 * Vec_SetMemory(vProof) / (1<<20), 1.0 * Vec_SetMemoryS(vProof) / (1<<20), 100.0 * (Vec_SetMemory(vProof) - Vec_SetMemoryS(vProof)) / Vec_SetMemory(vProof));
        spStatic_Sat_ProofReduce_TimeTotal += Abc_Clock() - clk;
        Abc_PrintTime( 1, "Time", spStatic_Sat_ProofReduce_TimeTotal );
    }
    Vec_SetShrink( vProof, Vec_SetHandCurrentS(vProof) );
    Vec_SetShrinkLimits( vProof );
//    Sat_ProofReduceCheck( s );
//Sat_ProofCheck0( vProof );

    return RetValue;
}

     

                                                                        

                                            

                                                                          
               
                

                

                                                                        
                                                                                          
 
               
                                     
                               
                                          
                                          
                                                 
         
                                      
                    
         
                     
     
                                                      
                 
     
                    
     
                                                             
                 
     
                         
                          
                                           
                            
                                          
                                         
                                               
                                          
                                         
                                                     
                                  
                                                                    
                           
                               
                                    
             
 

                                                                        

                                                

                
               
                

                

                                                                        
                                                                                                           
 
                  
                   
                                                       
                       
                                               
                           
                                                 
 

                                                                        

                                                

                
               
                

                

                                                                        
                                      
 
                                                    
                                                   
                                                                  
                          
                               
                                           
                                         
                               
                          
                      
               
                              
                                                       
                                          
                               
                                 
                                   
                                                  
     
                    
                                                                                     
                                                
         
                                                                                          
                                                                                
                                                         
         
                          
                  
     
                         
                      
                                          
                               
                                                                                          
                           
                                                                                             
        
                                                     
                                                  
              
                             
                         
 
      

/**Function*************************************************************

  Synopsis    [Collects nodes belonging to the UNSAT core.]

  Description [The resulting array contains 1-based IDs of root clauses.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_ProofCollectCore(vProof *Vec_Set_t, vUsed *Vec_Int_t) *Vec_Int_t {
    var vCore *Vec_Int_t;
    var pNode, pFanin *satset;
    var pBitMap *uint32;
    var i, k, MaxCla int32;
    // find the largest number
    // inlining of macro Proof_ForeachNodeVec
    for  i = 0; (i < Vec_IntSize(vUsed)) && ((pNode) = Proof_NodeRead(vProof, Vec_IntEntry(vUsed,i))); i++  {
        // inlining of macro Proof_NodeForeachFanin
        for  k = 0; (k < int32(pNode.nEnts)) && (((pFanin) = (pNode.pEnts[k] & 1) ? nil : Proof_NodeRead(vProof, pNode.pEnts[k] >> 2)), 1); k++  {
            if pFanin == nil {
                MaxCla = Abc_MaxInt( MaxCla, pNode.pEnts[k] >> 2 ); }
        }
    }
    // allocate bitmap
    // inlining of macro ABC_CALLOC
    pBitMap = ((unsigned *) calloc((size_t)(Abc_BitWordNum(MaxCla) + 1), sizeof(unsigned)));
    // collect leaves
    vCore = Vec_IntAlloc( 1000 );
    // inlining of macro Proof_ForeachNodeVec
    for  i = 0; (i < Vec_IntSize(vUsed)) && ((pNode) = Proof_NodeRead(vProof, Vec_IntEntry(vUsed,i))); i++  {
        // inlining of macro Proof_NodeForeachFanin
        for  k = 0; (k < int32(pNode.nEnts)) && (((pFanin) = (pNode.pEnts[k] & 1) ? nil : Proof_NodeRead(vProof, pNode.pEnts[k] >> 2)), 1); k++  {
            if pFanin == nil {
                var Entry int32  = (pNode.pEnts[k] >> 2);
                if Abc_InfoHasBit(pBitMap, Entry) {
                    continue; }
                Abc_InfoSetBit(pBitMap, Entry);
                Vec_IntPush( vCore, Entry );
            }
        }
    }
    // inlining of macro ABC_FREE
    ((pBitMap) ? (free((char *) (pBitMap)), (pBitMap) = 0) : 0);
//    Vec_IntUniqify( vCore );
    return vCore;
}

     

                                                                        

                                                              

                
               
                

                

                                                                        
                                                                           
 
               
                        
                             
                        
                                                  
                                      
                                        
                                             
                                                                          
                        
                                             
     
                                              
                                             
                               
                   
     
                                                      
                                            
                         
                        
                           
             
                    
                                                                                                                   
                    
                                                                                                                  
                    
                                                                                                                  
                    
                                                                                                                                                             
                    
                                                                                                                             
 

                                                                        

                                                  

                                                               
               
                

                

                                                                        
                                                               
 
                                                     
                                                    
                                                  
                                                                  
                                                       
                             
                     
                            
                                      
                               
                      
                                                                                         
                          
                      
                    

                                                  

                            
                                                       
                                                    
                                                   
                                               

                                              
                                          
                                              
                                               

                    
                                 
                                               
                                                  
                                

                                                          
                                                   
                                                       
     
                           
         
                                         
                                                  
                                                                        
                                                                                                    
         
            
                                         
                                   
                                            
                                       
     
                           

                                                              
                                                   
     
                                
                                   
                                                                        
         
                                                            
                         
                                                          
                                                            
                                                                              
                
                                                                               
         
                                   
                                       
     
                      
                                                         
                                  
                           

                            
                                                       
                                                 
                                                   
                      
              
                         
                         
                             
                
 


                                                                        

                                                  

                                                               
               
                

                

                                                                        
                                                                    
 
                                                     
                                                    
                                                  
                                                                  
                                                       
                             
                     
                                       
                                                    
                                            
                                      
                                       
                          
                      
                    

                                                  

                            
                                                       
                                                    
                                                     
                                               

                                              
                                          
                                              
                                               

                    
                                 

                                                          
                                                   
                                                       
     
                           
         
                                           
                                         
                                                  
                                                                        
                                                                                                      
                                                                                                   
         
            
                                           
                                        
                                   
                                            
                                         
                                                
     
                           

                                                              
                                                   
     
                                
                                   
                                                                        
         
                                                              
                                                             
                         
                                                            
                                                                                                              
                                                            
                                                                                
                                                                                                            
                
                                                                                 
                                                                                                             
         
                                   
                                         
                                                
     
                      
                                                         
                                    
                             

                            
                                                       
                                                 
                                                   
                      
              
                         
                         
                             
                        
                       
                
 

      

/**Function*************************************************************

  Synopsis    [Computes UNSAT core.]

  Description [The result is the array of root clause indexes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Proof_DeriveCore(vProof *Vec_Set_t, hRoot int32) any {
    var Roots Vec_Int_t = Vec_Int_t{ 1, 1, &hRoot };
    var vRoots *Vec_Int_t = &Roots;
    var vCore, vUsed *Vec_Int_t;
    if hRoot == -1 {
        return nil; }
    // collect visited clauses
    vUsed = Proof_CollectUsedIter( vProof, vRoots, 0 );
    // collect core clauses 
    vCore = Sat_ProofCollectCore( vProof, vUsed );
    Vec_IntFree( vUsed );
    Vec_IntSort( vCore, 1 );
    return vCore;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

