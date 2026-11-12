package fret

import (
	. "github.com/glycerine/abcgo/base/abc"
	. "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [fretFlow.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Flow-based retiming package.]

  Synopsis    [Max-flow computation.]

  Author      [Aaron Hurst]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2008.]

  Revision    [$Id: fretFlow.c,v 1.00 2008/01/01 00:00:00 ahurst Exp $]

***********************************************************************/

//#include "fretime.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void dfsfast_e_retreat( Abc_Obj_t *pObj );

// static void dfsfast_r_retreat( Abc_Obj_t *pObj );


// Original macro:
// #define FDIST(xn, xe, yn, ye) (FDATA(xn)->xe##_dist == (FDATA(yn)->ye##_dist + 1))


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Fast DFS.]

  Description [Uses sink-distance-histogram heuristic.  May not find all
               flow paths: this occurs in a small number of cases where
               the flow predecessor points to a non-adjacent node and
               the distance ordering is perturbed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func dfsfast_preorder(pNtk *Abc_Ntk_t) {
  var pObj, pNext *Abc_Obj_t;
  var vTimeIn *Vec_Ptr_t;
  var qn *Vec_Ptr_t = Vec_PtrAlloc(Abc_NtkObjNum(pNtk));
  var qe *Vec_Int_t  = Vec_IntAlloc(Abc_NtkObjNum(pNtk));
  var i, j, d, end int32;
  var qpos int32;

  // create reverse timing edges for backward traversal
  if (pManMR->maxDelay) {
    Abc_NtkForEachObj( pNtk, pObj, i ) {
      Vec_PtrForEachEntry( Abc_Obj_t *, FTIMEEDGES(pObj), pNext, j ) {
        vTimeIn = FDATA(pNext)->vNodes;
        if (!vTimeIn) {
          vTimeIn = FDATA(pNext)->vNodes = Vec_PtrAlloc(2);
        }
        Vec_PtrPush(vTimeIn, pObj);
      }
    }
  }

  // clear histogram
  assert(pManMR.vSinkDistHist, "pManMR.vSinkDistHist");
  memset(Vec_IntArray(pManMR.vSinkDistHist), 0, sizeof(int)*Vec_IntSize(pManMR.vSinkDistHist));

  // seed queue : latches, PIOs, and blocks
  Abc_NtkForEachObj( pNtk, pObj, i )
    if Abc_ObjIsPo(pObj) ||
        Abc_ObjIsLatch(pObj) || 
        (pManMR.fIsForward && FTEST(pObj, BLOCK_OR_CONS) & pManMR.constraintMask) {
      Vec_PtrPush(qn, pObj);
      Vec_IntPush(qe, 'r');
      FDATA(pObj).r_dist = 1;
    } else if Abc_ObjIsPi(pObj) || 
               (!pManMR.fIsForward && FTEST(pObj, BLOCK_OR_CONS) & pManMR.constraintMask) {
      Vec_PtrPush(qn, pObj);
      Vec_IntPush(qe, 'e');
      FDATA(pObj).e_dist = 1;
    }

  // until queue is empty...
  for qpos < Vec_PtrSize(qn) {
    pObj = (Abc_Obj_t *)Vec_PtrEntry(qn, qpos);
    assert((pObj != nil), "pObj");
    end = Vec_IntEntry(qe, qpos);
    qpos++;
    
    if end == 'r' {
      d = FDATA(pObj).r_dist;

      // 1. structural edges
      if pManMR.fIsForward {
        Abc_ObjForEachFanin( pObj, pNext, i )
          if !FDATA(pNext).e_dist {
            FDATA(pNext).e_dist = d+1;
            Vec_PtrPush(qn, pNext);
            Vec_IntPush(qe, 'e');
          }
      } else
        Abc_ObjForEachFanout( pObj, pNext, i )
          if !FDATA(pNext).e_dist {
            FDATA(pNext).e_dist = d+1;
            Vec_PtrPush(qn, pNext);
            Vec_IntPush(qe, 'e');
          }

      if d == 1 { continue; }

      // 2. reverse edges (forward retiming only)
      if pManMR.fIsForward {
        Abc_ObjForEachFanout( pObj, pNext, i )
          if !FDATA(pNext).r_dist && !Abc_ObjIsLatch(pNext) {
            FDATA(pNext).r_dist = d+1;
            Vec_PtrPush(qn, pNext);
            Vec_IntPush(qe, 'r');
          }          

      // 3. timimg edges (forward retiming only)
        if (pManMR->maxDelay && FDATA(pObj)->vNodes)
          Vec_PtrForEachEntry(Abc_Obj_t *, FDATA(pObj)->vNodes, pNext, i ) {
            if (!FDATA(pNext)->r_dist) {
              FDATA(pNext)->r_dist = d+1;
              Vec_PtrPush(qn, pNext);
              Vec_IntPush(qe, 'r');
            }
          }
      }
      
    } else { // if 'e'
      if Abc_ObjIsLatch(pObj) { continue; }

      d = FDATA(pObj).e_dist;

      // 1. through node
      if !FDATA(pObj).r_dist {
        FDATA(pObj).r_dist = d+1;
        Vec_PtrPush(qn, pObj);
        Vec_IntPush(qe, 'r');
      }

      // 2. reverse edges (backward retiming only)
      if !pManMR.fIsForward {
        Abc_ObjForEachFanin( pObj, pNext, i )
          if !FDATA(pNext).e_dist && !Abc_ObjIsLatch(pNext) {
            FDATA(pNext).e_dist = d+1;
            Vec_PtrPush(qn, pNext);
            Vec_IntPush(qe, 'e');
          }  

      // 3. timimg edges (backward retiming only)
        if (pManMR->maxDelay && FDATA(pObj)->vNodes)
          Vec_PtrForEachEntry(Abc_Obj_t *, FDATA(pObj)->vNodes, pNext, i ) {
            if (!FDATA(pNext)->e_dist) {
              FDATA(pNext)->e_dist = d+1;
              Vec_PtrPush(qn, pNext);
              Vec_IntPush(qe, 'e');
            }
          }
      }
    }
  }
 
  // free time edges
  if (pManMR->maxDelay) {
    Abc_NtkForEachObj( pNtk, pObj, i ) {
      vTimeIn = FDATA(pObj)->vNodes;
      if (vTimeIn) {
        Vec_PtrFree(vTimeIn);
        FDATA(pObj)->vNodes = 0;
      }
    }
  }

  Abc_NtkForEachObj( pNtk, pObj, i ) {
    Vec_IntAddToEntry(pManMR.vSinkDistHist, FDATA(pObj).r_dist, 1);
    Vec_IntAddToEntry(pManMR.vSinkDistHist, FDATA(pObj).e_dist, 1);

  }

  // printf("\t\tpre-ordered (max depth=%d)\n", d+1);

  // deallocate
  Vec_PtrFree( qn );
  Vec_IntFree( qe );
}

func dfsfast_e(pObj *Abc_Obj_t, pPred *Abc_Obj_t) int32 {
  var i int32;
  var pNext *Abc_Obj_t;
  
  if pManMR.fSinkDistTerminate { return 0; }

  // have we reached the sink?
  if FTEST(pObj, BLOCK_OR_CONS) & pManMR.constraintMask ||
     Abc_ObjIsPi(pObj) {
    assert((pPred != nil), "pPred");
    assert(!pManMR.fIsForward, "!pManMR.fIsForward");
    return 1;
  }

  FSET(pObj, VISITED_E);


  // 1. structural edges
  if (pManMR.fIsForward)
    Abc_ObjForEachFanout( pObj, pNext, i ) {
      if !FTEST(pNext, VISITED_R) &&
          FDIST(pObj, e, pNext, r) &&
          (dfsfast_r(pNext, pPred) != 0) {
        goto found;
      }
    }
  else
    Abc_ObjForEachFanin( pObj, pNext, i ) {
      if !FTEST(pNext, VISITED_R) &&
          FDIST(pObj, e, pNext, r) &&
          (dfsfast_r(pNext, pPred) != 0) {
        goto found;
      }
    }  
  
  if Abc_ObjIsLatch(pObj) {
    goto not_found; }

  // 2. reverse edges (backward retiming only)
  if !pManMR.fIsForward { 
    Abc_ObjForEachFanout( pObj, pNext, i ) {
      if !FTEST(pNext, VISITED_E) &&
          FDIST(pObj, e, pNext, e) &&
          (dfsfast_e(pNext, pPred) != 0) {
        goto found;
      }
    }

  // 3. timing edges (backward retiming only)
    if (pManMR->maxDelay) 
      Vec_PtrForEachEntry(Abc_Obj_t *, FTIMEEDGES(pObj), pNext, i) {
        if (!FTEST(pNext, VISITED_E) &&
            FDIST(pObj, e, pNext, e) &&
            dfsfast_e(pNext, pPred)) {
          goto found;
        }
      }
  }
  
  // unwind
  if FTEST(pObj, FLOW) && 
      !FTEST(pObj, VISITED_R) &&
      FDIST(pObj, e, pObj, r) &&
      (dfsfast_r(pObj, FGETPRED(pObj)) != 0) {

    FUNSET(pObj, FLOW);
    FSETPRED(pObj, nil);
    goto found;
  }
  
 not_found:
  FUNSET(pObj, VISITED_E);
  dfsfast_e_retreat(pObj);
  return 0;
  
 found:
  FUNSET(pObj, VISITED_E);
  return 1;
}

func dfsfast_r(pObj *Abc_Obj_t, pPred *Abc_Obj_t) int32 {
  var i int32;
  var pNext, pOldPred *Abc_Obj_t;

  if pManMR.fSinkDistTerminate { return 0; }


  // have we reached the sink?
  if Abc_ObjIsLatch(pObj) ||
      (pManMR.fIsForward && Abc_ObjIsPo(pObj)) || 
      (pManMR.fIsForward && FTEST(pObj, BLOCK_OR_CONS) & pManMR.constraintMask) {
    assert((pPred != nil), "pPred");
    return 1;
  }

  FSET(pObj, VISITED_R);

  if FTEST(pObj, FLOW) {

    pOldPred = FGETPRED(pObj);
    if (pOldPred != nil) && 
        !FTEST(pOldPred, VISITED_E) &&
        FDIST(pObj, r, pOldPred, e) &&
        (dfsfast_e(pOldPred, pOldPred) != 0) {

      FSETPRED(pObj, pPred);

      goto found;
    }
    
  } else {

    if !FTEST(pObj, VISITED_E) &&
        FDIST(pObj, r, pObj, e) &&
        (dfsfast_e(pObj, pObj) != 0) {

      FSET(pObj, FLOW);
      FSETPRED(pObj, pPred);

      goto found;
    }
  }
 
  // 2. reverse edges (forward retiming only)
  if pManMR.fIsForward {
    Abc_ObjForEachFanin( pObj, pNext, i ) {
      if !FTEST(pNext, VISITED_R) &&
          FDIST(pObj, r, pNext, r) &&
          !Abc_ObjIsLatch(pNext) &&
          (dfsfast_r(pNext, pPred) != 0) {
        goto found;
      }
    }
  
  // 3. timing edges (forward retiming only)
    if (pManMR->maxDelay) 
      Vec_PtrForEachEntry(Abc_Obj_t*, FTIMEEDGES(pObj), pNext, i) {
        if (!FTEST(pNext, VISITED_R) &&
            FDIST(pObj, r, pNext, r) &&
            dfsfast_r(pNext, pPred)) {
          goto found;
        }
      }
  }
  
  FUNSET(pObj, VISITED_R);
  dfsfast_r_retreat(pObj);
  return 0;

 found:
  FUNSET(pObj, VISITED_R);
  return 1;
}

func dfsfast_e_retreat(pObj *Abc_Obj_t) {
  var pNext *Abc_Obj_t;
  var i int32;
  var h *int32;
  var old_dist int32  = FDATA(pObj).e_dist;
  var adj_dist int32;
  var min_dist int32 = MAX_DIST;
  
  // 1. structural edges
  if (pManMR.fIsForward)
    Abc_ObjForEachFanout( pObj, pNext, i ) {
      adj_dist = FDATA(pNext).r_dist;
      if adj_dist != 0 { min_dist = MIN(min_dist, adj_dist); }
    }
  else
    Abc_ObjForEachFanin( pObj, pNext, i ) {
      adj_dist = FDATA(pNext).r_dist;
      if adj_dist != 0 { min_dist = MIN(min_dist, adj_dist); }
    }

  if Abc_ObjIsLatch(pObj) { goto update; }

  // 2. through
  if FTEST(pObj, FLOW) {
    adj_dist = FDATA(pObj).r_dist;
    if adj_dist != 0 { min_dist = MIN(min_dist, adj_dist); }
  }

  // 3. reverse edges (backward retiming only)
  if !pManMR.fIsForward {
    Abc_ObjForEachFanout( pObj, pNext, i ) {
      adj_dist = FDATA(pNext).e_dist;
      if adj_dist != 0 { min_dist = MIN(min_dist, adj_dist); }
    }

  // 4. timing edges (backward retiming only)
    if (pManMR->maxDelay) 
      Vec_PtrForEachEntry(Abc_Obj_t*, FTIMEEDGES(pObj), pNext, i) {
        adj_dist = FDATA(pNext)->e_dist;
        if (adj_dist) min_dist = MIN(min_dist, adj_dist);
      }
  }

 update:
  ++min_dist;
  if min_dist >= MAX_DIST { min_dist = 0; }
  // printf("[%de=%d->%d] ", Abc_ObjId(pObj), old_dist, min_dist+1);
  FDATA(pObj).e_dist = min_dist;

  assert(min_dist < Vec_IntSize(pManMR.vSinkDistHist), "min_dist < Vec_IntSize(pManMR.vSinkDistHist)");
  h = Vec_IntArray(pManMR.vSinkDistHist);
  h[old_dist]--;
  h[min_dist]++;
  if !h[old_dist] {
    pManMR.fSinkDistTerminate = 1;
  }
}

func dfsfast_r_retreat(pObj *Abc_Obj_t) {
  var pNext *Abc_Obj_t;
  var i int32;
  var h *int32;
  var old_dist int32  = FDATA(pObj).r_dist;
  var adj_dist int32;
  var min_dist int32 = MAX_DIST;
  
  // 1. through or pred
  if FTEST(pObj, FLOW) {
    if FGETPRED(pObj) {
      adj_dist = FDATA(FGETPRED(pObj)).e_dist;
      if adj_dist != 0 { min_dist = MIN(min_dist, adj_dist); }
    }
  } else {
    adj_dist = FDATA(pObj).e_dist;
    if adj_dist != 0 { min_dist = MIN(min_dist, adj_dist); }
  }

  // 2. reverse edges (forward retiming only)
  if pManMR.fIsForward {
    Abc_ObjForEachFanin( pObj, pNext, i )
      if !Abc_ObjIsLatch(pNext) {
        adj_dist = FDATA(pNext).r_dist;
        if adj_dist != 0 { min_dist = MIN(min_dist, adj_dist); }
      }

  // 3. timing edges (forward retiming only)
    if (pManMR->maxDelay) 
      Vec_PtrForEachEntry(Abc_Obj_t*, FTIMEEDGES(pObj), pNext, i) {
        adj_dist = FDATA(pNext)->r_dist;
        if (adj_dist) min_dist = MIN(min_dist, adj_dist);
      }
  }

  ++min_dist;
  if min_dist >= MAX_DIST { min_dist = 0; }
  //printf("[%dr=%d->%d] ", Abc_ObjId(pObj), old_dist, min_dist+1);
  FDATA(pObj).r_dist = min_dist;

  assert(min_dist < Vec_IntSize(pManMR.vSinkDistHist), "min_dist < Vec_IntSize(pManMR.vSinkDistHist)");
  h = Vec_IntArray(pManMR.vSinkDistHist);
  h[old_dist]--;
  h[min_dist]++;
  if !h[old_dist] {
    pManMR.fSinkDistTerminate = 1;
  }
}

/**Function*************************************************************

  Synopsis    [Plain DFS.]

  Description [Does not use sink-distance-histogram heuristic.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func dfsplain_e(pObj *Abc_Obj_t, pPred *Abc_Obj_t) int32 {
  var i int32;
  var pNext *Abc_Obj_t;
  
  if FTEST(pObj, BLOCK_OR_CONS) & pManMR.constraintMask || 
      Abc_ObjIsPi(pObj) {
    assert((pPred != nil), "pPred");
    assert(!pManMR.fIsForward, "!pManMR.fIsForward");
    return 1;
  }

  FSET(pObj, VISITED_E);
  
  // printf(" %de\n", Abc_ObjId(pObj));
  
  // 1. structural edges
  if (pManMR.fIsForward)
    Abc_ObjForEachFanout( pObj, pNext, i ) {
      if !FTEST(pNext, VISITED_R) &&
          (dfsplain_r(pNext, pPred) != 0) {
        goto found;
      }
    }
  else
    Abc_ObjForEachFanin( pObj, pNext, i ) {
      if !FTEST(pNext, VISITED_R) &&
          (dfsplain_r(pNext, pPred) != 0) {
        goto found;
      }
    }
  
  if Abc_ObjIsLatch(pObj) {
    return 0; }

  // 2. reverse edges (backward retiming only)
  if !pManMR.fIsForward {
    Abc_ObjForEachFanout( pObj, pNext, i ) {
      if !FTEST(pNext, VISITED_E) &&
          (dfsplain_e(pNext, pPred) != 0) {
        goto found;
      }
    }

  // 3. timing edges (backward retiming only)
    if (pManMR->maxDelay) 
      Vec_PtrForEachEntry(Abc_Obj_t*, FTIMEEDGES(pObj), pNext, i) {
        if (!FTEST(pNext, VISITED_E) &&
            dfsplain_e(pNext, pPred)) {
          goto found;
        }
      }
  }
  
  // unwind
  if FTEST(pObj, FLOW) && 
      !FTEST(pObj, VISITED_R) &&
      (dfsplain_r(pObj, FGETPRED(pObj)) != 0) {
    FUNSET(pObj, FLOW);
    FSETPRED(pObj, nil);
    goto found;
  }
  
  return 0;
  
 found:

  return 1;
}

func dfsplain_r(pObj *Abc_Obj_t, pPred *Abc_Obj_t) int32 {
  var i int32;
  var pNext, pOldPred *Abc_Obj_t;

  // have we reached the sink?
  if Abc_ObjIsLatch(pObj) || 
      (pManMR.fIsForward && Abc_ObjIsPo(pObj)) || 
      (pManMR.fIsForward && FTEST(pObj, BLOCK_OR_CONS) & pManMR.constraintMask) {
    assert((pPred != nil), "pPred");
    return 1;
  }

  FSET(pObj, VISITED_R);

  // printf(" %dr\n", Abc_ObjId(pObj));

  if FTEST(pObj, FLOW) {

    pOldPred = FGETPRED(pObj);
    if (pOldPred != nil) && 
        !FTEST(pOldPred, VISITED_E) &&
        (dfsplain_e(pOldPred, pOldPred) != 0) {

      FSETPRED(pObj, pPred);

      goto found;
    }
    
  } else {

    if !FTEST(pObj, VISITED_E) &&
        (dfsplain_e(pObj, pObj) != 0) {

      FSET(pObj, FLOW);
      FSETPRED(pObj, pPred);

      goto found;
    }
  }
 
  // 2. follow reverse edges
  if pManMR.fIsForward { // forward retiming only
    Abc_ObjForEachFanin( pObj, pNext, i ) {
      if !FTEST(pNext, VISITED_R) &&
          !Abc_ObjIsLatch(pNext) &&
          (dfsplain_r(pNext, pPred) != 0) {
        goto found;
      }
    }
  
  // 3. timing edges (forward only)
    if (pManMR->maxDelay) 
      Vec_PtrForEachEntry(Abc_Obj_t*, FTIMEEDGES(pObj), pNext, i) {
        if (!FTEST(pNext, VISITED_R) &&
            dfsplain_r(pNext, pPred)) {
          goto found;
        }
      }
  }
  
  return 0;

 found:
  return 1;
}
//ABC_NAMESPACE_IMPL_END

