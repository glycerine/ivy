package saig

import "fmt"


int  Saig_ManFilterUsingIndOne2( Aig_Man_t * p, Aig_Man_t * pFrame, sat_solver * pSat, Cnf_Dat_t * pCnf, int nConfs, int nProps, int Counter 
                                 , int type_ /* jlong --  */
                                 )
{
  var pObj *Aig_Obj_t;
  var Lit, status int32;
  pObj = Aig_ManCo( pFrame, Counter*3+type_ ); /* which co */
  Lit  = toLitCond( pCnf.pVarNums[Aig_ObjId(pObj)], 0 );
  status = sat_solver_solve( pSat, &Lit, &Lit + 1, (ABC_INT64_T)nConfs, 0, 0, 0 );
  if status == l_False {    /* unsat */
      return status; }
  if status == l_Undef {
      fmt.Printf("Solver returned undecided.\n");
      return status;
    }
  assert(status == l_True, "status == l_True");
  return status;
}

/**Function*************************************************************

   Synopsis    [Detects constraints functionally.]

   Description []
               
   SideEffects []

   SeeAlso     []

***********************************************************************/
func Saig_ManFilterUsingInd2(p *Aig_Man_t, vCands *Vec_Vec_t, nConfs int32, nProps int32, fVerbose int32) {
  var vNodes *Vec_Ptr_t;
  var pFrames *Aig_Man_t;
  var pSat *sat_solver;
  var pCnf *Cnf_Dat_t;
  var pObj *Aig_Obj_t;
  var i, k, k2, Counter int32;
  /*
    Vec_VecForEachLevel( vCands, vNodes, i )
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, k )
    printf( "%d ", Aig_ObjId(Aig_Regular(pObj)) );
    printf( "\n" );
  */
  // create timeframes 
  //    pFrames = Saig_ManUnrollInd( p );
  pFrames = Saig_ManCreateIndMiter2( p, vCands );
  assert(Aig_ManCoNum(pFrames) == Vec_VecSizeSize(vCands)*3, "Aig_ManCoNum(pFrames) == Vec_VecSizeSize(vCands)*3");
  // start the SAT solver
  pCnf = Cnf_DeriveSimple( pFrames, Aig_ManCoNum(pFrames) );
  pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
  // check candidates
  if fVerbose != 0 {
    fmt.Printf("Filtered cands:  \n"); }
  Counter = 0;
  Vec_VecForEachLevel( vCands, vNodes, i )
    {
      assert(i==0, "i==0");             /* only one item */
      k2 = 0;
      Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, k )
        {
          if Saig_ManFilterUsingIndOne2( p, pFrames, pSat, pCnf, nConfs, nProps, Counter++ , 0) == l_False {
            //            if ( Saig_ManFilterUsingIndOne_old( p, pSat, pCnf, nConfs, pObj ) )
            
              Vec_PtrWriteEntry( vNodes, k2++, pObj );
              if fVerbose != 0 {
                fmt.Printf("%d:%s%d \n", i, Aig_IsComplement(pObj)? "!":"", Aig_ObjId(Aig_Regular(pObj))); }
              fmt.Printf(" type I : %d:%s%d \n", i, Aig_IsComplement(pObj)? "!":"", Aig_ObjId(Aig_Regular(pObj)));
              Vec_PtrPush(p.unfold2_type_I, pObj);
            
          /* jlong -- begin  */
          } else if Saig_ManFilterUsingIndOne2( p, pFrames, pSat, pCnf, nConfs, nProps, Counter-1 , 1) == l_True { /* can be self-conflicting */
            
              if Saig_ManFilterUsingIndOne2( p, pFrames, pSat, pCnf, nConfs, nProps, Counter-1 , 2) == l_False{
                //Vec_PtrWriteEntry( vNodes, k2++, pObj );
                if fVerbose != 0 {
                  fmt.Printf("%d:%s%d  \n", i, Aig_IsComplement(pObj)? "!":"", Aig_ObjId(Aig_Regular(pObj))); }
                fmt.Printf(" type II: %d:%s%d  \n", i, Aig_IsComplement(pObj)? "!":"", Aig_ObjId(Aig_Regular(pObj)));
                Vec_PtrWriteEntry( vNodes, k2++, pObj ); /* add type II constraints */
                Vec_PtrPush(p.unfold2_type_II, pObj);
              }
            }
          /* jlong -- end  */
        }
      Vec_PtrShrink( vNodes, k2 );
    }

  // clean up
  Cnf_DataFree( pCnf );
  sat_solver_delete( pSat );
  if fVerbose != 0 {
    Aig_ManPrintStats( pFrames ); }
  Aig_ManStop( pFrames );
}


/**Function*************************************************************

   Synopsis    [Returns the number of variables implied by the output.]

   Description []
               
   SideEffects []

   SeeAlso     []

***********************************************************************/
func Ssw_ManFindDirectImplications2(p *Aig_Man_t, nFrames int32, nConfs int32, nProps int32, fVerbose int32) *Vec_Vec_t {
  var vCands *Vec_Vec_t  = nil;
  var vNodes *Vec_Ptr_t;
  var pCnf *Cnf_Dat_t;
  var pSat *sat_solver;
  var pFrames *Aig_Man_t;
  var pObj, pRepr, pReprR *Aig_Obj_t;
  var i, f, k, value int32;
  assert(nFrames == 1, "nFrames == 1");
  vCands = Vec_VecAlloc( nFrames );
  assert(nFrames == 1, "nFrames == 1");
  // perform unrolling
  pFrames = Saig_ManUnrollCOI( p, nFrames );
  assert(Aig_ManCoNum(pFrames) == 1, "Aig_ManCoNum(pFrames) == 1");
  // start the SAT solver
  pCnf = Cnf_DeriveSimple( pFrames, 0 );
  pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
  if pSat != nil {
      Aig_ManIncrementTravId( p );
      for  f = 0; f < nFrames; f++  {
          Aig_ManForEachObj( p, pObj, i )
            {
              if !Aig_ObjIsCand(pObj) {
                continue; }
              //--jlong : also use internal nodes as well
              /* if ( !Aig_ObjIsCi(pObj) ) */
              /*   continue; */ 
              if Aig_ObjIsTravIdCurrent(p, pObj) {
                continue; }
              // get the node from timeframes
              pRepr  = p.pObjCopies[nFrames*i + nFrames-1-f];
              pReprR = Aig_Regular(pRepr);
              if pCnf.pVarNums[Aig_ObjId(pReprR)] < 0 {
                continue; }
              //                value = pSat->assigns[ pCnf->pVarNums[Aig_ObjId(pReprR)] ];
              value = sat_solver_get_var_value( pSat, pCnf.pVarNums[Aig_ObjId(pReprR)] );
              if value == l_Undef {
                continue; }
              // label this node as taken
              Aig_ObjSetTravIdCurrent(p, pObj);
              if Saig_ObjIsLo(p, pObj) {
                Aig_ObjSetTravIdCurrent( p, Aig_ObjFanin0(Saig_ObjLoToLi(p, pObj)) ); }
              // remember the node
              Vec_VecPush( vCands, f, Aig_NotCond( pObj, (value == l_True) ^ Aig_IsComplement(pRepr) ) );
              //        printf( "%s%d ", (value == l_False)? "":"!", i );
            }
        }
      //    printf( "\n" );
      sat_solver_delete( pSat );
    }
  Aig_ManStop( pFrames );
  Cnf_DataFree( pCnf );

  if fVerbose != 0 {
      fmt.Printf("Found %3d candidates.\n", Vec_VecSizeSize(vCands));
      Vec_VecForEachLevel( vCands, vNodes, k )
        {
          fmt.Printf("Level %d. Cands  =%d    ", k, Vec_PtrSize(vNodes));
          //            Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
          //                printf( "%d:%s%d ", k, Aig_IsComplement(pObj)? "!":"", Aig_ObjId(Aig_Regular(pObj)) );
          fmt.Printf("\n");
        }
    }

  ABC_FREE( p.pObjCopies );
  /* -- jlong -- this does the SAT proof of the constraints */
  Saig_ManFilterUsingInd2( p, vCands, nConfs, nProps, fVerbose );
  if Vec_VecSizeSize(vCands) {
    fmt.Printf("Found %3d constraints after filtering.\n", Vec_VecSizeSize(vCands)); }
  if fVerbose != 0 {
      Vec_VecForEachLevel( vCands, vNodes, k )
        {
          fmt.Printf("Level %d. Constr =%d    ", k, Vec_PtrSize(vNodes));
          //            Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
          //                printf( "%d:%s%d ", k, Aig_IsComplement(pObj)? "!":"", Aig_ObjId(Aig_Regular(pObj)) );
          fmt.Printf("\n");
        }
    }
  
  return vCands;
}

/**Function*************************************************************

   Synopsis    [Duplicates the AIG while unfolding constraints.]

   Description []
               
   SideEffects []

   SeeAlso     []

***********************************************************************/
func Saig_ManDupUnfoldConstrsFunc2(pAig *Aig_Man_t, nFrames int32, nConfs int32, nProps int32, fOldAlgo int32, fVerbose int32, typeII_cnt *int32) *Aig_Man_t {
  var pNew *Aig_Man_t;
  var vCands *Vec_Vec_t;
  var vNewFlops *Vec_Ptr_t;
  var pObj *Aig_Obj_t;
  var i, k, nNewFlops int32;
  int fCompl = 0 ;
  if fOldAlgo != 0 {
    vCands = Saig_ManDetectConstrFunc( pAig, nFrames, nConfs, nProps, fVerbose );
  } else {
    vCands = Ssw_ManFindDirectImplications2( pAig, nFrames, nConfs, nProps, fVerbose ); }
  if vCands == nil || Vec_VecSizeSize(vCands) == 0 {
      Vec_VecFreeP( &vCands );
      return Aig_ManDupDfs( pAig );
    }
  // create new manager
  pNew = Aig_ManDupWithoutPos( pAig ); /* good */
  pNew.nConstrs = pAig.nConstrs + Vec_VecSizeSize(vCands);
  pNew.nConstrs = pAig.nConstrs + Vec_PtrSize(pAig.unfold2_type_II)
    + Vec_PtrSize(pAig.unfold2_type_I);
  //  pNew->nConstrsTypeII  = Vec_PtrSize(pAig->unfold2_type_II);
  *typeII_cnt = Vec_PtrSize(pAig.unfold2_type_II);

  /* new set of registers */
  
  // add normal POs
  Saig_ManForEachPo( pAig, pObj, i )
    Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
  // create constraint outputs
  vNewFlops = Vec_PtrAlloc( 100 );


  Vec_PtrForEachEntry(Aig_Obj_t * , pAig.unfold2_type_I, pObj, k){
    var x *Aig_Obj_t  = Aig_ObjRealCopy(pObj);
    Aig_ObjCreateCo(pNew, x);
  }
   
  Vec_PtrForEachEntry(Aig_Obj_t * , pAig.unfold2_type_II, pObj, k){
    var type_II_latch *Aig_Obj_t 
      = Aig_ObjCreateCi(pNew); /* will get connected later; */
    var x *Aig_Obj_t  = Aig_ObjRealCopy(pObj);
    
    var n *Aig_Obj_t  = Aig_And(pNew, 
                            Aig_NotCond(type_II_latch, fCompl),
                            Aig_NotCond(x, fCompl));
    Aig_ObjCreateCo(pNew, n);//Aig_Not(n));
  }                   
  
  // add latch outputs
  Saig_ManForEachLi( pAig, pObj, i )
    Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
  
  Vec_PtrForEachEntry(Aig_Obj_t * , pAig.unfold2_type_II, pObj, k){
    var x *Aig_Obj_t  = Aig_ObjRealCopy(pObj);
    Aig_ObjCreateCo(pNew, x);
  }                   

  // add new latch outputs
  nNewFlops = Vec_PtrSize(pAig.unfold2_type_II);
  //assert( nNewFlops == Vec_PtrSize(vNewFlops) );
  Aig_ManSetRegNum( pNew, Aig_ManRegNum(pAig) + nNewFlops );
  fmt.Printf("#reg after unfold2: %d\n", Aig_ManRegNum(pAig) + nNewFlops);
  Vec_VecFreeP( &vCands );
  Vec_PtrFree( vNewFlops );
  return pNew;

}

/**Function*************************************************************

   Synopsis    [Duplicates the AIG while unfolding constraints.]

   Description []
               
   SideEffects []

   SeeAlso     []

***********************************************************************/
func Saig_ManDupFoldConstrsFunc2(pAig *Aig_Man_t, fCompl int32, fVerbose int32, typeII_cnt int32) *Aig_Man_t {
  var pAigNew *Aig_Man_t;
  var pMiter, pFlopOut, pFlopIn, pObj *Aig_Obj_t;
  var i, typeII_cc, type_II int32;
  if Aig_ManConstrNum(pAig) == 0 {
    return Aig_ManDupDfs( pAig ); }
  assert(Aig_ManConstrNum(pAig) < Saig_ManPoNum(pAig), "Aig_ManConstrNum(pAig) < Saig_ManPoNum(pAig)");
  // start the new manager
  pAigNew = Aig_ManStart( Aig_ManNodeNum(pAig) );
  pAigNew.pName = Abc_UtilStrsav( pAig.pName );
  pAigNew.pSpec = Abc_UtilStrsav( pAig.pSpec );
  // map the constant node
  Aig_ManConst1(pAig).pData = Aig_ManConst1( pAigNew );
  // create variables for PIs
  Aig_ManForEachCi( pAig, pObj, i )
    pObj.pData = Aig_ObjCreateCi( pAigNew );
  // add internal nodes of this frame
  Aig_ManForEachNode( pAig, pObj, i )
    pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );

  // OR the constraint outputs
  pMiter = Aig_ManConst0( pAigNew );
  typeII_cc = 0;//typeII_cnt;
  typeII_cnt = 0;  
  type_II = 0;
    
  Saig_ManForEachPo( pAig, pObj, i )
    {
      
      if i < Saig_ManPoNum(pAig)-Aig_ManConstrNum(pAig) {
        continue; }
      if i + typeII_cnt >= Saig_ManPoNum(pAig) {
        type_II = 1;
      }
      /*  now we got the constraint */
      if type_II != 0 {

        var type_II_latch *Aig_Obj_t 
          = Aig_ObjCreateCi(pAigNew); /* will get connected later; */
        pMiter = Aig_Or(pAigNew, pMiter, 
                        Aig_And(pAigNew, 
                                Aig_NotCond(type_II_latch, fCompl),
                                Aig_NotCond( Aig_ObjChild0Copy(pObj), fCompl ) )
                        );
        fmt.Printf("modeling typeII : %d:%s%d \n", i, Aig_IsComplement(pObj)? "!":"", Aig_ObjId(Aig_Regular(pObj)));
      } else { 
        pMiter = Aig_Or( pAigNew, pMiter, Aig_NotCond( Aig_ObjChild0Copy(pObj), fCompl ) ); }
    }

  // create additional flop
  if Saig_ManRegNum(pAig) > 0 {
      pFlopOut = Aig_ObjCreateCi( pAigNew );
      pFlopIn  = Aig_Or( pAigNew, pMiter, pFlopOut );
    
  } else { 
    pFlopIn = pMiter; }

  // create primary output
  Saig_ManForEachPo( pAig, pObj, i )
    {
      if i >= Saig_ManPoNum(pAig)-Aig_ManConstrNum(pAig) {
        continue; }
      pMiter = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_Not(pFlopIn) );
      Aig_ObjCreateCo( pAigNew, pMiter );
    }

  // transfer to register outputs
  {
    /* the same for type I and type II */
    var pObjLi, pObjLo *Aig_Obj_t;

    Saig_ManForEachLiLo( pAig, pObjLi, pObjLo, i )    {
      if i + typeII_cc < Aig_ManRegNum(pAig) {
        var c *Aig_Obj_t  = Aig_Mux(pAigNew, Aig_Not(pFlopIn), 
                               Aig_ObjChild0Copy(pObjLi) ,
                               (Aig_Obj_t*)pObjLo.pData);
        Aig_ObjCreateCo( pAigNew, c);
      } else {
        fmt.Printf("skipping: reg%d\n", i);
        Aig_ObjCreateCo( pAigNew,Aig_ObjChild0Copy(pObjLi));
      }
    }

  }
  if(0)Saig_ManForEachLi( pAig, pObj, i ) {
      Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }
  Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAig) );
    
  type_II = 0;
  Saig_ManForEachPo( pAig, pObj, i )
    {
    
      if i < Saig_ManPoNum(pAig)-Aig_ManConstrNum(pAig) {
        continue; }
      if i + typeII_cnt >= Saig_ManPoNum(pAig) {
        type_II = 1;
      }
      /*  now we got the constraint */
      if type_II != 0 {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj));
        Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAigNew)+1 );
        fmt.Printf("Latch for typeII : %d:%s%d \n", i, Aig_IsComplement(pObj)? "!":"", Aig_ObjId(Aig_Regular(pObj)));
      }
    }

     
  // create additional flop 

  if Saig_ManRegNum(pAig) > 0 {
      Aig_ObjCreateCo( pAigNew, pFlopIn );
      Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAigNew)+1 );
    }
  fmt.Printf("#reg after fold2: %d\n", Aig_ManRegNum(pAigNew));
  // perform cleanup
  Aig_ManCleanup( pAigNew );
  Aig_ManSeqCleanup( pAigNew );
  return pAigNew;
}
