package bsat

/**CFile****************************************************************

  FileName    [satProof2.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Proof logging.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: satProof2.h,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/vec/vec.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

   
type Prf_Man_t struct {
    iFirst             int32;    // first learned clause with proof
    iFirst2             int32;   // first learned clause with proof
    nWords             int32;    // the number of proof words
    pInfo           *word;     // pointer to the current proof
    vInfo      *Vec_Wrd_t;     // proof information
    vSaved      *Vec_Int_t;    // IDs of saved clauses
    vId2Pr      *Vec_Int_t;    // mapping proof IDs of problem clauses into bitshifts (user's array)
}

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

func Prf_BitWordNum(nWidth int32) int32 { return (nWidth >> 6) + ((nWidth & 63) > 0);                           }
func Prf_ManSize(p *Prf_Man_t) int32 { return Vec_WrdSize( p.vInfo ) / p.nWords;                           }
func Prf_ManClearNewInfo(p *Prf_Man_t) { var w int32; for  w = 0; w < p.nWords; w++  { Vec_WrdPush( p.vInfo, 0 ); }  }
func Prf_ManClauseInfo(p *Prf_Man_t, Id int32) *word { return Vec_WrdEntryP( p.vInfo, Id * p.nWords );                     }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Prf_ManAlloc() *Prf_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Prf_Man_t{};
    p.iFirst  = -1;
    p.iFirst2 = -1;
    p.vInfo   = Vec_WrdAlloc( 1000 );
    p.vSaved  = Vec_IntAlloc( 1000 );
    return p;
}
func Prf_ManStop(p *Prf_Man_t) {
    if p == nil {
        return; }
    Vec_IntFree( p.vSaved );
    Vec_WrdFree( p.vInfo );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Prf_ManStopP(p **Prf_Man_t) {
    Prf_ManStop( *p );
    *p = nil;
}
func Prf_ManMemory(p *Prf_Man_t) float64 {
    return Vec_WrdMemory(p.vInfo) + Vec_IntMemory(p.vSaved) + sizeof(Prf_Man_t);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Prf_ManRestart(p *Prf_Man_t, vId2Pr *Vec_Int_t, iFirst int32, nWidth int32) {
    assert(p.iFirst == -1, "p.iFirst == -1");
    p.iFirst = iFirst;
    p.nWords = Prf_BitWordNum( nWidth );
    p.vId2Pr = vId2Pr;
    p.pInfo  = nil;
    Vec_WrdClear( p.vInfo );
}
func Prf_ManGrow(p *Prf_Man_t, nWidth int32) {
    var vInfoNew *Vec_Wrd_t;
    var i, w, nSize, nWordsNew int32;
    assert(p.iFirst >= 0, "p.iFirst >= 0");
    assert(p.pInfo == nil, "p.pInfo == nil");
    if nWidth < 64 * p.nWords {
        return; }
    // get word count after resizing
    nWordsNew = Abc_MaxInt( Prf_BitWordNum(nWidth), 2 * p.nWords );
    // remap the entries
    nSize = Prf_ManSize( p );
    vInfoNew = Vec_WrdAlloc( (nSize + 1000) * nWordsNew );
    for  i = 0; i < nSize; i++  {
        p.pInfo = Prf_ManClauseInfo( p, i );
        for  w = 0; w < p.nWords; w++  {
            Vec_WrdPush( vInfoNew, p.pInfo[w] );
        }
        for  ; w < nWordsNew; w++  {
            Vec_WrdPush( vInfoNew, 0 );
        }
    }
    Vec_WrdFree( p.vInfo );
    p.vInfo = vInfoNew;
    p.nWords = nWordsNew;
    p.pInfo = nil;

}
func Prf_ManShrink(p *Prf_Man_t, iClause int32) {
    assert(p.iFirst >= 0, "p.iFirst >= 0");
    assert(iClause - p.iFirst >= 0, "iClause - p.iFirst >= 0");
    assert(iClause - p.iFirst < Prf_ManSize(p), "iClause - p.iFirst < Prf_ManSize(p)"); 
    Vec_WrdShrink( p.vInfo, (iClause - p.iFirst) * p.nWords );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Prf_ManAddSaved(p *Prf_Man_t, i int32, iNew int32) {
    assert(p.iFirst >= 0, "p.iFirst >= 0");
    if i < p.iFirst {
        return; }
    if Vec_IntSize(p.vSaved) == 0 {
        assert(p.iFirst2 == -1, "p.iFirst2 == -1");
        p.iFirst2 = iNew;
    }
    Vec_IntPush( p.vSaved, i );
}
func Prf_ManCompact(p *Prf_Man_t, iNew int32) {
    var i, w, k, Entry, nSize int32;
    assert(p.iFirst >= 0, "p.iFirst >= 0");
    assert(p.pInfo == nil, "p.pInfo == nil");
    nSize = Prf_ManSize( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vSaved)) && (((Entry) = Vec_IntEntry(p.vSaved, i)), 1); i++  {
        assert(Entry - p.iFirst >= 0 && Entry - p.iFirst < nSize, "Entry - p.iFirst >= 0 && Entry - p.iFirst < nSize");
        p.pInfo = Prf_ManClauseInfo( p, Entry - p.iFirst );
        for  w = 0; w < p.nWords; w++  {
            Vec_WrdWriteEntry( p.vInfo, k++, p.pInfo[w] );
        }
    }
    Vec_WrdShrink( p.vInfo, k );
    Vec_IntClear( p.vSaved );
    p.pInfo = nil;
    // update first
    if p.iFirst2 == -1 {
        p.iFirst = iNew;
    } else {
        p.iFirst = p.iFirst2; }
    p.iFirst2 = -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Prf_ManChainResolve(p *Prf_Man_t, c *clause) {
    assert(p.iFirst >= 0, "p.iFirst >= 0");
    assert(p.pInfo != nil, "p.pInfo != nil");
    // add to proof info
    if c.lrn { // learned clause
    
        if clause_id(c) >= p.iFirst {
            var pProofStart *word;
            var w int32;
            assert(clause_id(c) - p.iFirst >= 0, "clause_id(c) - p.iFirst >= 0");
            assert(clause_id(c) - p.iFirst < Prf_ManSize(p), "clause_id(c) - p.iFirst < Prf_ManSize(p)"); 
            pProofStart = Prf_ManClauseInfo( p, clause_id(c) - p.iFirst );
            for  w = 0; w < p.nWords; w++  {
                p.pInfo[w] |= pProofStart[w];
            }
        }
    
    } else { // problem clause
      
        if clause_id(c) >= 0 { // has proof ID
        
            var Entry int32;
            if p.vId2Pr == nil {
                Entry = clause_id(c);
            } else {
                Entry = Vec_IntEntry( p.vId2Pr, clause_id(c) ); }
            if Entry >= 0 {
                assert(Entry < 64 * p.nWords, "Entry < 64 * p.nWords");
                Abc_InfoSetBit( (unsigned *)p.pInfo, Entry );
            }
        }
    }
}
func Prf_ManChainStart(p *Prf_Man_t, c *clause) {
    assert(p.iFirst >= 0, "p.iFirst >= 0");
    // prepare info for new clause
    Prf_ManClearNewInfo( p );
    // get pointer to the proof
    assert(p.pInfo == nil, "p.pInfo == nil");
    p.pInfo = Prf_ManClauseInfo( p, Prf_ManSize(p)-1 );
    // add to proof info
    Prf_ManChainResolve( p, c );
}
func Prf_ManChainStop(p *Prf_Man_t) int32 {
    assert(p.pInfo != nil, "p.pInfo != nil");
    p.pInfo = nil;
    return 0;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Prf_ManUnsatCore(p *Prf_Man_t) *Vec_Int_t {
    var vCore *Vec_Int_t;
    var i, Entry int32;
    assert(p.iFirst >= 0, "p.iFirst >= 0");
    assert(p.pInfo == nil, "p.pInfo == nil");
    assert(Prf_ManSize(p) > 0, "Prf_ManSize(p) > 0");
    vCore = Vec_IntAlloc( 64 * p.nWords );
    p.pInfo = Prf_ManClauseInfo( p, Prf_ManSize(p)-1 );
    if p.vId2Pr == nil {
        for  i = 0; i < 64 * p.nWords; i++  {
            if Abc_InfoHasBit( (unsigned *)p.pInfo, i ) {
                Vec_IntPush( vCore, i ); }
        }
    
    } else {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vId2Pr)) && (((Entry) = Vec_IntEntry(p.vId2Pr, i)), 1); i++  {
            if Entry < 0 {
                continue; }
            assert(Entry < 64 * p.nWords, "Entry < 64 * p.nWords");
            if Abc_InfoHasBit( (unsigned *)p.pInfo, Entry ) {
                Vec_IntPush( vCore, i ); }
        }
    }
    p.pInfo = nil;
    Vec_IntSort( vCore, 1 );
    return vCore;
}



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

