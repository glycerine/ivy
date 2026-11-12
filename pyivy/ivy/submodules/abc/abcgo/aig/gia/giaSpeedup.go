package gia

import "fmt"

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
//#include "map/if/if.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sorts the pins in the decreasing order of delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sorts the pins in the decreasing order of delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes the arrival times for the given object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Propagates the required times through the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes the delay trace of the given network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDelayTraceLut(p *Gia_Man_t) float32 {
    var fUseSorting int32  = 1;
    var pLutLib *If_LibLut_t  = (If_LibLut_t *)p.pLutLib;
    var vObjs *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var tArrival, tArrivalCur, tRequired, tSlack float32;
    var i, iObj int32;

    // get the library
    if (pLutLib != nil) && pLutLib.LutMax < Gia_ManLutSizeMax(p) {
        fmt.Printf("The max LUT size (%d) is less than the max fanin count (%d).\n", pLutLib.LutMax, Gia_ManLutSizeMax(p));
        return -TIM_ETERNITY;
    }

    // initialize the arrival times
    Gia_ManTimeStart( p );
    Gia_ManLevelNum( p );

    // propagate arrival times
    if p.pManTime {
        Tim_ManIncrementTravId( (Tim_Man_t *)p.pManTime ); }
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if !Gia_ObjIsCi(pObj) && !Gia_ObjIsCo(pObj) && !Gia_ObjIsLut(p, i) {
            continue; }
        tArrival = Gia_ObjComputeArrival( p, i, fUseSorting );
        if Gia_ObjIsCi(pObj) && p.pManTime {
            tArrival = Tim_ManGetCiArrival( (Tim_Man_t *)p.pManTime, Gia_ObjCioId(pObj) );
//printf( "%.3f  ", tArrival );
        }
        if Gia_ObjIsCo(pObj) && p.pManTime {
            Tim_ManSetCoArrival( (Tim_Man_t *)p.pManTime, Gia_ObjCioId(pObj), tArrival ); }
        Gia_ObjSetTimeArrival( p, i, tArrival );
    }

    // get the latest arrival times
    tArrival = -TIM_ETERNITY;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        tArrivalCur = Gia_ObjTimeArrivalObj( p, Gia_ObjFanin0(pObj) );
        Gia_ObjSetTimeArrival( p, Gia_ObjId(p,pObj), tArrivalCur );
        if tArrival < tArrivalCur {
            tArrival = tArrivalCur; }
    }

    // initialize the required times
    if p.pManTime {
        Tim_ManIncrementTravId( (Tim_Man_t *)p.pManTime );
        Tim_ManInitPoRequiredAll( (Tim_Man_t *)p.pManTime, tArrival );
    
    } else {
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ObjSetTimeRequiredObj( p, pObj, tArrival );
        }
    }

    // propagate the required times
    vObjs = Gia_ManOrderReverse( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        pObj = Gia_ManObj(p, iObj);
        if Gia_ObjIsLut(p, iObj) {
            Gia_ObjPropagateRequired( p, iObj, fUseSorting );
        
        } else if Gia_ObjIsCi(pObj) {
            if p.pManTime {
                Tim_ManSetCiRequired( (Tim_Man_t *)p.pManTime, Gia_ObjCioId(pObj), Gia_ObjTimeRequired(p, iObj) ); }
        
        } else if Gia_ObjIsCo(pObj) {
            if p.pManTime {
                tRequired = Tim_ManGetCoRequired( (Tim_Man_t *)p.pManTime, Gia_ObjCioId(pObj) );
                Gia_ObjSetTimeRequired( p, iObj, tRequired );
            }
            if Gia_ObjTimeRequired(p, Gia_ObjFaninId0p(p, pObj)) > Gia_ObjTimeRequired(p, iObj) {
                Gia_ObjSetTimeRequired(p, Gia_ObjFaninId0p(p, pObj), Gia_ObjTimeRequired(p, iObj) ); }
        }

        // set slack for this object
        tSlack = Gia_ObjTimeRequired(p, iObj) - Gia_ObjTimeArrival(p, iObj);
        assert(tSlack + 0.01 > 0.0, "tSlack + 0.01 > 0.0");
        Gia_ObjSetTimeSlack( p, iObj, tSlack < 0.0 ? 0.0 : tSlack );
    }
    Vec_IntFree( vObjs );
    return tArrival;
}

     

                                                                        

                                                               

                
               
                

                

                                                                        
                                                                        
 
                                       
                     
                         
                        
                                       
                  
                            
                                             
                             
                          
     
                                                
         
                                                     
                                                                        
                                                                      
         
     
                                       
     
                                                
         
                                                                      
                                                            
                                                                        
                                                                      
         
     
        
     
                          
         
                                                     
             
                                                                          
                                                                              
                                                                         
                                                                                 
                                                                     
                                                                            
                                                                          
             
         
            
         
                                                    
             
                                                                          
                                                              
                                                                      
                                                                     
                                                                            
                                                                          
             
         
     
                     
 

                                                                        

                                                              

                
               
                

                

                                                                        
                                         
 
             
                              
          
                                   
     
                                                               
                     
                                                       
                                                         
                                                                                       
                                                                                                                  
                                                                   
                                                                                         
                                                                                                                   
                                                                     
     
             
 

      

/**Function*************************************************************

  Synopsis    [Prints the delay trace for the given network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDelayTraceLutPrint(p *Gia_Man_t, fVerbose int32) float32 {
    var pLutLib *If_LibLut_t  = (If_LibLut_t *)p.pLutLib;
    var i int32;
    var Nodes int32;
    var pCounters *int32;
    var tArrival, tDelta, nSteps, Num float32;
    // get the library
    if (pLutLib != nil) && pLutLib.LutMax < Gia_ManLutSizeMax(p) {
        fmt.Printf("The max LUT size (%d) is less than the max fanin count (%d).\n", pLutLib.LutMax, Gia_ManLutSizeMax(p));
        return -ABC_INFINITY;
    }
    // decide how many steps
    if pLutLib != nil {
        nSteps = 20;
    } else {
        nSteps = Gia_ManLutLevel(p, nil);
    }
    // inlining of macro ABC_ALLOC
    pCounters = ((int *) malloc(sizeof(int) * (size_t)(nSteps + 1)));
    memset( pCounters, 0, sizeof(int)*(nSteps + 1) );
    // perform delay trace
    tArrival = Gia_ManDelayTraceLut( p );
    tDelta   = tArrival / nSteps;
    // count how many nodes have slack in the corresponding intervals
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
        if Gia_ObjLutSize(p, i) == 0 {
            continue; }
        Num = Gia_ObjTimeSlack(p, i) / tDelta;
        if Num > nSteps {
            continue; }
        assert(Num >=0 && Num <= nSteps, "Num >=0 && Num <= nSteps");
        pCounters[int32(Num)]++;
    }
    }
    // print the results    
    if fVerbose != 0 {
        fmt.Printf("Max delay = %6.2f. Delay trace using %s model:\n", tArrival, pLutLib? "LUT library" : "unit-delay");
        Nodes = 0;
        for  i = 0; i < nSteps; i++  {
            Nodes += pCounters[i];
            fmt.Printf("%3d %s : %5d  (%6.2f %%)\n", pLutLib? 5*(i+1) : i+1, pLutLib? "%":"lev", Nodes, 100.0*Nodes/Gia_ManLutNum(p));
        }
    }
    // inlining of macro ABC_FREE
    pCounters = nil;
    Gia_ManTimeStop( p );
    return tArrival;
}

/**Function*************************************************************

  Synopsis    [Determines timing-critical edges of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_LutDelayTraceTCEdges(p *Gia_Man_t, iObj int32, tDelta float32) uint32 {
    var pLutLib *If_LibLut_t  = (If_LibLut_t *)p.pLutLib;
    var pPinPerm [32]int32;
    var pPinDelays [32]float32;
    var tRequired float32;
    var pDelays *float32;
    var uResult uint32;
    var k, iFanin int32;
    tRequired = Gia_ObjTimeRequired( p, iObj );
    if pLutLib == nil {
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
            if tRequired < Gia_ObjTimeArrival(p, iFanin) + 1.0 + tDelta {
                uResult |= (1 << k); }
        }
    
    } else if !pLutLib.fVarPinDelays {
        pDelays = pLutLib.pLutDelays[Gia_ObjLutSize(p, iObj)];
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
            if tRequired < Gia_ObjTimeArrival(p, iFanin) + pDelays[0] + tDelta {
                uResult |= (1 << k); }
        }
    
    } else {
        pDelays = pLutLib.pLutDelays[Gia_ObjLutSize(p, iObj)];
        Gia_LutDelayTraceSortPins( p, iObj, pPinPerm, pPinDelays );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
            if tRequired < Gia_ObjTimeArrival( p, Gia_ObjLutFanin(p, iObj,pPinPerm[k])) + pDelays[k] + tDelta {
                uResult |= (1 << pPinPerm[k]); }
        }
    }
    return uResult;
}

/**Function*************************************************************

  Synopsis    [Adds strashed nodes for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpeedupObj_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vNodes *Vec_Int_t) int32 {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return 1; }
    Gia_ObjSetTravIdCurrent( p, pObj );
    if Gia_ObjIsCi(pObj) {
        return 0; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if !(Gia_ManSpeedupObj_rec( p, Gia_ObjFanin0(pObj), vNodes ) != 0) {
        return 0; }
    if !(Gia_ManSpeedupObj_rec( p, Gia_ObjFanin1(pObj), vNodes ) != 0) {
        return 0; }
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds strashed nodes for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpeedupObj(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vLeaves *Vec_Int_t, vTimes *Vec_Int_t) {
    var vNodes *Vec_Int_t;
    var pTemp *Gia_Obj_t  = nil;
    var pCofs [32]int32;
    var nCofs int32;
    var nSkip int32;
    var i int32;
    var k int32;
    var iResult int32;
    var iObj int32;
    // mark the leaves
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vLeaves)) && ((pTemp) = Gia_ManObj(p, Vec_IntEntry(vLeaves,i))); i++  {
        Gia_ObjSetTravIdCurrent( p, pTemp );
    }
    // collect the AIG nodes
    vNodes = Vec_IntAlloc( 100 );
    if !(Gia_ManSpeedupObj_rec( p, pObj, vNodes ) != 0) {
        fmt.Printf("Bad node!!!\n");
        Vec_IntFree( vNodes );
        return;
    }
    // derive cofactors
    nCofs = (1 << Vec_IntSize(vTimes));
    for  i = 0; i < nCofs; i++  {
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vLeaves)) && ((pTemp) = Gia_ManObj(p, Vec_IntEntry(vLeaves,k))); k++  {
            pTemp.Value = Abc_Var2Lit( Gia_ObjId(p, pTemp), 0 );
        }
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vTimes)) && ((pTemp) = Gia_ManObj(p, Vec_IntEntry(vTimes,k))); k++  {
            pTemp.Value = ((i & (1<<k)) != 0);
        }
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vNodes)) && ((pTemp) = Gia_ManObj(p, Vec_IntEntry(vNodes,k))); k++  {
            pTemp.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pTemp), Gia_ObjFanin1Copy(pTemp) );
        }
        pCofs[i] = pTemp.Value;
    }
    Vec_IntFree( vNodes );
    // collect the resulting tree
    // inlining of macro Gia_ManForEachObjVec
    for  k = 0; (k < Vec_IntSize(vTimes)) && ((pTemp) = Gia_ManObj(p, Vec_IntEntry(vTimes,k))); k++  {
        for  nSkip = (1<<k), i = 0; i < nCofs; i += 2*nSkip  {
            pCofs[i] = Gia_ManHashMux( pNew, Gia_ObjToLit(p,pTemp), pCofs[i+nSkip], pCofs[i] );
        }
    }
    // create choice node  (pObj is repr and ppCofs[0] is new)
    iObj    = Gia_ObjId( p, pObj );
    iResult = Abc_Lit2Var( pCofs[0] );
    if iResult <= iObj {
        return; }
    Gia_ObjSetRepr( pNew, iResult, iObj );
    Gia_ObjSetNext( pNew, iResult, Gia_ObjNext(pNew, iObj) );
    Gia_ObjSetNext( pNew, iObj, iResult );
}

/**Function*************************************************************

  Synopsis    [Adds choices to speed up the network by the given percentage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpeedup(p *Gia_Man_t, Percentage int32, Degree int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var vTimeCries, vTimeFanins *Vec_Int_t;
    var iObj, iFanin, iFanin2, nNodesNew int32;
    var tDelta, tArrival float32;
    var i, k, k2, Counter, CounterRes, nTimeCris int32;
    var fUseLutLib int32  = (p.pLutLib != nil);
    var pTempTim any  = nil;
    var puTCEdges *uint32;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    if !(fUseLutLib != 0) && p.pManTime {
        pTempTim = p.pManTime;
        p.pManTime = Tim_ManDup( (Tim_Man_t *)pTempTim, 1 );
    }
    // perform delay trace
    tArrival = Gia_ManDelayTraceLut( p );
    if fUseLutLib != 0 {
        tDelta = tArrival*Percentage/100.0;
    } else {
        tDelta = 1.0;
    }
    if fVerbose != 0 {
        fmt.Printf("Max delay = %.2f. Delta = %.2f. ", tArrival, tDelta);
        fmt.Printf("Using %s model. ", fUseLutLib ? "LUT library" : "unit-delay");
        if fUseLutLib != 0 {
            fmt.Printf("Percentage = %d. ", Percentage); }
        fmt.Printf("\n");
    }
    // mark the timing critical nodes and edges
    // inlining of macro ABC_CALLOC
    puTCEdges = ((unsigned *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(unsigned)));
    // inlining of macro Gia_ManForEachLut
    for  iObj = 1; iObj < Gia_ManObjNum(p); iObj++  { if !Gia_ObjIsLut(p, iObj) {} else {
        if Gia_ObjTimeSlack(p, iObj) >= tDelta {
            continue; }
        puTCEdges[iObj] = Gia_LutDelayTraceTCEdges( p, iObj, tDelta );
    }
    }
    if fVerbose != 0 {
        Counter = CounterRes = 0;
        // inlining of macro Gia_ManForEachLut
        for  iObj = 1; iObj < Gia_ManObjNum(p); iObj++  { if !Gia_ObjIsLut(p, iObj) {} else {
            // inlining of macro Gia_LutForEachFanin
            for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
                if !Gia_ObjIsCi(Gia_ManObj(p, iFanin)) && Gia_ObjTimeSlack(p, iFanin) < tDelta {
                    Counter++; }
            }
            CounterRes += Gia_WordCountOnes( puTCEdges[iObj] );
        }
        }
        fmt.Printf("Edges: Total = %7d. 0-slack = %7d. Critical = %7d. Ratio = %4.2f\n", Gia_ManLutFaninCount(p), Counter, CounterRes, Counter? 1.0*CounterRes/Counter : 0.0);
    }

    // start the resulting network
    pNew = Gia_ManDup( p );
    Gia_ManHashStart( pNew );
    nNodesNew = 1000 + 3 * Gia_ManObjNum(pNew);
    // inlining of macro ABC_CALLOC
    pNew.pNexts = ((int *) calloc((size_t)(nNodesNew), sizeof(int)));
    // inlining of macro ABC_CALLOC
    pNew.pReprs = ((Gia_Rpr_t *) calloc((size_t)(nNodesNew), sizeof(Gia_Rpr_t)));
    for  i = 0; i < nNodesNew; i++  {
        Gia_ObjSetRepr( pNew, i, GIA_VOID );
    }

    // collect nodes to be used for resynthesis
    Counter = CounterRes = 0;
    vTimeCries = Vec_IntAlloc( 16 );
    vTimeFanins = Vec_IntAlloc( 16 );
    // inlining of macro Gia_ManForEachLut
    for  iObj = 1; iObj < Gia_ManObjNum(p); iObj++  { if !Gia_ObjIsLut(p, iObj) {} else {
        if Gia_ObjTimeSlack(p, iObj) >= tDelta {
            continue; }
        // count the number of non-PI timing-critical nodes
        nTimeCris = 0;
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
            if !Gia_ObjIsCi(Gia_ManObj(p, iFanin)) && (puTCEdges[iObj] & (1<<k)) {
                nTimeCris++; }
        }
        if !(fVeryVerbose != 0) && nTimeCris == 0 {
            continue; }
        Counter++;
        // count the total number of timing critical second-generation nodes
        Vec_IntClear( vTimeCries );
        if nTimeCris != 0 {
            // inlining of macro Gia_LutForEachFanin
            for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
                if !Gia_ObjIsCi(Gia_ManObj(p, iFanin)) && (puTCEdges[iObj] & (1<<k)) {
                    // inlining of macro Gia_LutForEachFanin
                    for  k2 = 0; k2 < Gia_ObjLutSize(p,iFanin) && ((iFanin2 = Gia_ObjLutFanins(p,iFanin)[k2]),1); k2++  {
                        if puTCEdges[iFanin] & (1<<k2) {
                            Vec_IntPushUnique( vTimeCries, iFanin2 ); }
                    } }
            }
        }
//        if ( !fVeryVerbose && (Vec_IntSize(vTimeCries) == 0 || Vec_IntSize(vTimeCries) > Degree) )
        if Vec_IntSize(vTimeCries) == 0 || Vec_IntSize(vTimeCries) > Degree {
            continue; }
        CounterRes++;
        // collect second generation nodes
        Vec_IntClear( vTimeFanins );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
            if Gia_ObjIsCi(Gia_ManObj(p, iFanin)) {
                Vec_IntPushUnique( vTimeFanins, iFanin );
            } else {
                // inlining of macro Gia_LutForEachFanin
                for  k2 = 0; k2 < Gia_ObjLutSize(p,iFanin) && ((iFanin2 = Gia_ObjLutFanins(p,iFanin)[k2]),1); k2++  {
                    Vec_IntPushUnique( vTimeFanins, iFanin2 );
                } }                    
        }
        // print the results
        if fVeryVerbose != 0 {
        fmt.Printf("%5d Node %5d : %d %2d %2d  ", Counter, iObj, nTimeCris, Vec_IntSize(vTimeCries), Vec_IntSize(vTimeFanins));
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
            fmt.Printf("%d(%.2f)%s ", iFanin, Gia_ObjTimeSlack(p, iFanin), (puTCEdges[iObj] & (1<<k))? "*":"");
        }
        fmt.Printf("\n");
        }
        // add the node to choices
        if Vec_IntSize(vTimeCries) == 0 || Vec_IntSize(vTimeCries) > Degree {
            continue; }
        // order the fanins in the increasing order of criticalily
        if Vec_IntSize(vTimeCries) > 1 {
            iFanin = Vec_IntEntry( vTimeCries, 0 );
            iFanin2 = Vec_IntEntry( vTimeCries, 1 );
            if Gia_ObjTimeSlack(p, iFanin) < Gia_ObjTimeSlack(p, iFanin2) {
                Vec_IntWriteEntry( vTimeCries, 0, iFanin2 );
                Vec_IntWriteEntry( vTimeCries, 1, iFanin );
            }
        }
        if Vec_IntSize(vTimeCries) > 2 {
            iFanin = Vec_IntEntry( vTimeCries, 1 );
            iFanin2 = Vec_IntEntry( vTimeCries, 2 );
            if Gia_ObjTimeSlack(p, iFanin) < Gia_ObjTimeSlack(p, iFanin2) {
                Vec_IntWriteEntry( vTimeCries, 1, iFanin2 );
                Vec_IntWriteEntry( vTimeCries, 2, iFanin );
            }
            iFanin = Vec_IntEntry( vTimeCries, 0 );
            iFanin2 = Vec_IntEntry( vTimeCries, 1 );
            if Gia_ObjTimeSlack(p, iFanin) < Gia_ObjTimeSlack(p, iFanin2) {
                Vec_IntWriteEntry( vTimeCries, 0, iFanin2 );
                Vec_IntWriteEntry( vTimeCries, 1, iFanin );
            }
        }
        // add choice
        Gia_ManSpeedupObj( pNew, p, Gia_ManObj(p,iObj), vTimeFanins, vTimeCries );
        // quit if the number of nodes is large
        if Gia_ManObjNum(pNew) > nNodesNew - 100 {
            fmt.Printf("Speedup stopped adding choices because there was too many to add.\n");
            break;
        }
    }
    }
    Gia_ManTimeStop( p );
    Vec_IntFree( vTimeCries );
    Vec_IntFree( vTimeFanins );
    // inlining of macro ABC_FREE
    puTCEdges = nil;
    if fVerbose != 0 {
        fmt.Printf("Nodes: Total = %7d. 0-slack = %7d. Workable = %7d. Ratio = %4.2f\n", Gia_ManLutNum(p), Counter, CounterRes, Counter? 1.0*CounterRes/Counter : 0.0); } 
    if pTempTim {
        Tim_ManStop( (Tim_Man_t *)p.pManTime );
        p.pManTime = pTempTim;
    }
    // derive AIG with choices
//Gia_ManPrintStats( pNew, 0 );
    pTemp = Gia_ManEquivToChoices( pNew, 1 );
    Gia_ManStop( pNew );
//Gia_ManPrintStats( pTemp, 0 );
//    pNew = Gia_ManDupOrderDfsChoices( pTemp );
//    Gia_ManStop( pTemp );
//Gia_ManPrintStats( pNew, 0 );
//    return pNew;
    return pTemp;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

