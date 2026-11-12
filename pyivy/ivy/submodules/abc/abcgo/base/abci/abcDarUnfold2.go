package abci

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Abc_Ntk_t * Abc_NtkDarFold2( Abc_Ntk_t * pNtk, int fCompl, int fVerbose , int);

func Abc_NtkDarUnfold2(pNtk *Abc_Ntk_t, nFrames int32, nConfs int32, nProps int32, fStruct int32, fOldAlgo int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan *Aig_Man_t;
    var pTemp *Aig_Man_t = nil;
    var typeII_cnt int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    if fStruct != 0{
      assert(false, "0");//pMan = Saig_ManDupUnfoldConstrs( pTemp = pMan );
    }else {
      pMan = Saig_ManDupUnfoldConstrsFunc2( pTemp = pMan, nFrames, nConfs, nProps, fOldAlgo, fVerbose , &typeII_cnt); }
    Aig_ManStop( pTemp );
    if pMan == nil {
        return nil; }
    //    typeII_cnt = pMan->nConstrsTypeII;
    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig.pName = Extra_UtilStrsav(pMan.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pMan.pSpec);
    Aig_ManStop( pMan );

    return pNtkAig;//Abc_NtkDarFold2(pNtkAig, 0, fVerbose, typeII_cnt);
    
    //return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarFold2(pNtk *Abc_Ntk_t, fCompl int32, fVerbose int32, typeII_cnt int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    pMan = Saig_ManDupFoldConstrsFunc2( pTemp = pMan, fCompl, fVerbose, typeII_cnt );
    Aig_ManStop( pTemp );
    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig.pName = Extra_UtilStrsav(pMan.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pMan.pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
}

