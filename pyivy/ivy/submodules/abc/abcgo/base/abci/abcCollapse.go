package abci

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [abcCollapse.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Collapsing the network into two-levels.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcCollapse.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/gia/gia.h"
//#include "misc/vec/vecWec.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

#ifdef ABC_USE_CUDD

extern int Abc_NodeSupport( DdNode * bFunc, Vec_Str_t * vSupport, int nVars );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Makes nodes minimum base.]

  Description [Returns the number of changed nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Abc_NodeMinimumBase2( Abc_Obj_t * pNode )
{
    Vec_Str_t * vSupport;
    Vec_Ptr_t * vFanins;
    DdNode * bTemp;
    int i, nVars;

    assert( Abc_NtkIsBddLogic(pNode->pNtk) );
    assert( Abc_ObjIsNode(pNode) );

    // compute support
    vSupport = Vec_StrAlloc( 10 );
    nVars = Abc_NodeSupport( Cudd_Regular(pNode->pData), vSupport, Abc_ObjFaninNum(pNode) );
    if ( nVars == Abc_ObjFaninNum(pNode) )
    {
        Vec_StrFree( vSupport );
        return 0;
    }

    // add fanins
    vFanins = Vec_PtrAlloc( Abc_ObjFaninNum(pNode) );
    Abc_NodeCollectFanins( pNode, vFanins );
    Vec_IntClear( &pNode->vFanins );
    for ( i = 0; i < vFanins->nSize; i++ )
        if ( vSupport->pArray[i] != 0 ) // useful
            Vec_IntPush( &pNode->vFanins, Abc_ObjId((Abc_Obj_t *)vFanins->pArray[i]) );
    assert( nVars == Abc_ObjFaninNum(pNode) );

    // update the function of the node
    pNode->pData = Extra_bddRemapUp( (DdManager *)pNode->pNtk->pManFunc, bTemp = (DdNode *)pNode->pData );   Cudd_Ref( (DdNode *)pNode->pData );
    Cudd_RecursiveDeref( (DdManager *)pNode->pNtk->pManFunc, bTemp );
    Vec_PtrFree( vFanins );
    Vec_StrFree( vSupport );
    return 1;
}
int Abc_NtkMinimumBase2( Abc_Ntk_t * pNtk )
{
    Abc_Obj_t * pNode, * pFanin;
    int i, k, Counter;
    assert( Abc_NtkIsBddLogic(pNtk) );
    // remove all fanouts
    Abc_NtkForEachObj( pNtk, pNode, i )
        Vec_IntClear( &pNode->vFanouts );
    // add useful fanins
    Counter = 0;
    Abc_NtkForEachNode( pNtk, pNode, i )
        Counter += Abc_NodeMinimumBase2( pNode );
    // add fanouts
    Abc_NtkForEachObj( pNtk, pNode, i )
        Abc_ObjForEachFanin( pNode, pFanin, k )
            Vec_IntPush( &pFanin->vFanouts, Abc_ObjId(pNode) );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collapses the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
Abc_Obj_t * Abc_NodeFromGlobalBdds( Abc_Ntk_t * pNtkNew, DdManager * dd, DdNode * bFunc, int fReverse )
{
    Abc_Obj_t * pNodeNew, * pTemp;
    int i;
    // create a new node
    pNodeNew = Abc_NtkCreateNode( pNtkNew );
    // add the fanins in the order, in which they appear in the reordered manager
    Abc_NtkForEachCi( pNtkNew, pTemp, i )
        Abc_ObjAddFanin( pNodeNew, Abc_NtkCi(pNtkNew, fReverse ? Abc_NtkCiNum(pNtkNew)-1-dd->invperm[i] : dd->invperm[i]) );
    // transfer the function
    pNodeNew->pData = Extra_TransferLevelByLevel( dd, (DdManager *)pNtkNew->pManFunc, bFunc );  Cudd_Ref( (DdNode *)pNodeNew->pData );
    return pNodeNew;
}
Abc_Ntk_t * Abc_NtkFromGlobalBdds( Abc_Ntk_t * pNtk, int fReverse )
{
    ProgressBar * pProgress;
    Abc_Ntk_t * pNtkNew;
    Abc_Obj_t * pNode, * pDriver, * pNodeNew;
    DdManager * dd = (DdManager *)Abc_NtkGlobalBddMan( pNtk );
    int i;

    // extract don't-care and compute ISOP
    if ( pNtk->pExdc )
    {
        DdManager * ddExdc = NULL;
        DdNode * bBddMin, * bBddDc, * bBddL, * bBddU;
        assert( Abc_NtkIsStrash(pNtk->pExdc) );
        assert( Abc_NtkCoNum(pNtk->pExdc) == 1 );
        // compute the global BDDs
        if ( Abc_NtkBuildGlobalBdds(pNtk->pExdc, 10000000, 1, 1, 0, 0) == NULL )
            return NULL;
        // transfer tot the same manager
        ddExdc = (DdManager *)Abc_NtkGlobalBddMan( pNtk->pExdc );
        bBddDc = (DdNode *)Abc_ObjGlobalBdd(Abc_NtkCo(pNtk->pExdc, 0));
        bBddDc = Cudd_bddTransfer( ddExdc, dd, bBddDc );  Cudd_Ref( bBddDc );
        Abc_NtkFreeGlobalBdds( pNtk->pExdc, 1 );
        // minimize the output
        Abc_NtkForEachCo( pNtk, pNode, i )
        {
            bBddMin = (DdNode *)Abc_ObjGlobalBdd(pNode);
            // derive lower and uppwer bound
            bBddL = Cudd_bddAnd( dd, bBddMin,           Cudd_Not(bBddDc) );  Cudd_Ref( bBddL );
            bBddU = Cudd_bddAnd( dd, Cudd_Not(bBddMin), Cudd_Not(bBddDc) );  Cudd_Ref( bBddU );
            Cudd_RecursiveDeref( dd, bBddMin );
            // compute new one
            bBddMin = Cudd_bddIsop( dd, bBddL, Cudd_Not(bBddU) );            Cudd_Ref( bBddMin );
            Cudd_RecursiveDeref( dd, bBddL );
            Cudd_RecursiveDeref( dd, bBddU );
            // update global BDD
            Abc_ObjSetGlobalBdd( pNode, bBddMin );
            //Extra_bddPrint( dd, bBddMin ); printf( "\n" );
        }
        Cudd_RecursiveDeref( dd, bBddDc );
    }

    // start the new network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_BDD );
    // make sure the new manager has the same number of inputs
    Cudd_bddIthVar( (DdManager *)pNtkNew->pManFunc, dd->size-1 );
    // process the POs
    pProgress = Extra_ProgressBarStart( stdout, Abc_NtkCoNum(pNtk) );
    Abc_NtkForEachCo( pNtk, pNode, i )
    {
        Extra_ProgressBarUpdate( pProgress, i, NULL );
        pDriver = Abc_ObjFanin0(pNode);
        if ( Abc_ObjIsCi(pDriver) && !strcmp(Abc_ObjName(pNode), Abc_ObjName(pDriver)) )
        {
            Abc_ObjAddFanin( pNode->pCopy, pDriver->pCopy );
            continue;
        }
        pNodeNew = Abc_NodeFromGlobalBdds( pNtkNew, dd, (DdNode *)Abc_ObjGlobalBdd(pNode), fReverse );
        Abc_ObjAddFanin( pNode->pCopy, pNodeNew );

    }
    Extra_ProgressBarStop( pProgress );
    return pNtkNew;
}
void Abc_NtkDumpVariableOrder( Abc_Ntk_t * pNtk )
{
    DdManager * dd = (DdManager *)Abc_NtkGlobalBddMan( pNtk );
    FILE * pFile = fopen( "order.txt", "wb" ); int i;
    for ( i = 0; i < dd->size; i++ )
        fprintf( pFile, "%d ", dd->invperm[i] );
    fprintf( pFile, "\n" );
    fclose( pFile );
}
Abc_Ntk_t * Abc_NtkCollapse( Abc_Ntk_t * pNtk, int fBddSizeMax, int fDualRail, int fReorder, int fReverse, int fDumpOrder, int fVerbose )
{
    Abc_Ntk_t * pNtkNew;
    abctime clk = Abc_Clock();

    assert( Abc_NtkIsStrash(pNtk) );
    // compute the global BDDs
    if ( Abc_NtkBuildGlobalBdds(pNtk, fBddSizeMax, 1, fReorder, fReverse, fVerbose) == NULL )
        return NULL;
    if ( fVerbose )
    {
        DdManager * dd = (DdManager *)Abc_NtkGlobalBddMan( pNtk );
        printf( "Shared BDD size = %6d nodes.  ", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd) );
        ABC_PRT( "BDD construction time", Abc_Clock() - clk );
    }
    if ( fDumpOrder ) 
        Abc_NtkDumpVariableOrder( pNtk );

    // create the new network
    pNtkNew = Abc_NtkFromGlobalBdds( pNtk, fReverse );
    Abc_NtkFreeGlobalBdds( pNtk, 1 );
    if ( pNtkNew == NULL )
        return NULL;

    // make the network minimum base
    Abc_NtkMinimumBase2( pNtkNew );

    if ( pNtk->pExdc )
        pNtkNew->pExdc = Abc_NtkDup( pNtk->pExdc );

    // make sure that everything is okay
    if ( !Abc_NtkCheck( pNtkNew ) )
    {
        printf( "Abc_NtkCollapse: The network check has failed.\n" );
        Abc_NtkDelete( pNtkNew );
        return NULL;
    }
    return pNtkNew;
}


#else

Abc_Ntk_t * Abc_NtkCollapse( Abc_Ntk_t * pNtk, int fBddSizeMax, int fDualRail, int fReorder, int fReverse, int fDumpOrder, int fVerbose )
{
    return NULL;
}

#endif


     

                                                                        

                                                                            

                
               
                

                

                                                                        
                                                               
 
                     
                                                                                              
                            
                                     
                                      
                                                               
                                                               
                                                           
                                                           
                                                               
 
                                                                            
 
                     
                      
                              
                                
                                                
                                                
                             
                     
                                   
                                        
                                                            
                            
                                   
                                   
                                                              
                                                         
                                  
                      
                                          
                         
                
 
                                                                                                                                             
 
                     
                              
                                                                                                                            
                                                             
                   
                                                                                                        
                                                                                   
                        
                       
                    
                                             
                            
                   
                                                      
                 
 

                                                                        

                                                         

                
               
                

                

                                                                        
                                                                  
 
                              
                             
                                                                
                                      
                                                          
                                        
                                                                          
                                                                          
                                                             
                   
                                                                     
                  
 

                                                                        

                                                                             

                
               
                

                

                                                                        
                                                               
 
                                             
                   
                  
                    
                 
                                                          
                   
                  
                    
                 
              
 
                                                                        
 
                             
                                                            
                                      
     
                                                                                   
                                     
     
                                   
                                                                                           
                                                                      
              
                                                          
                                                               
                     
                  
 

                                                                        

                                     

                
               
                

                

                                                                        
                                                                                                                                                                       
 
                         
                     
               
                              
                                                                                             
                       
                    
                        
                                            
                 
                                                    
                                            
                                                                 
                            
                                                                                             
                        
                    
 
                                                                                                                                 
 
                            
                        
                                             
                                                     
                                     
                       
           

                                        
                                                   
                     

                                  
                                                    
                                     
                                                    
                                       
                       
     
                  
                                                             
                                                       
                                                                   
                                                                                              
                                    
         
                                                                                                   
                                                                                  
                                     
                                    
                                  
                        
         
                                 
     
                                   
                                                    
                                      
                         
                                                          
                                                

                            
                                                                    
                            
                                                     
                                        
                                                          
                                                                
                      
                                                                     
                                        
                                                          
     
                                                      
                                       
                                                                                        
         
                                                                                       
                     
         
                                   
         
                                                    
                                                                                    
                                                                                                                            
                                                      
                     
         
                                             
         
                                                    
                                                                                                                          
                                                      
                     
         
                                                                                                                                                                                            
                               
         
                                     
                           
                  
         
                                                  
     
                               
                            
                              
                          
                                       
                   
 
                                                                                                                                    
 
                        
                                    
                             
                                                                                                
                          
                    
                      
                                                   
                                        
                                   
     
                                                                        
                                 
                    
     
                   
 

      



// extern Vec_Wec_t * Gia_ManCreateCoSupps( Gia_Man_t * p, int fVerbose );

// extern int         Gia_ManCoLargestSupp( Gia_Man_t * p, Vec_Wec_t * vSupps );

// extern Vec_Wec_t * Gia_ManIsoStrashReduceInt( Gia_Man_t * p, Vec_Wec_t * vSupps, int fVerbose );


/**Function*************************************************************

  Synopsis    [Derives GIA for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkClpGia_rec(pNew *Gia_Man_t, pNode *Abc_Obj_t) int32 {
    var iLit0, iLit1 int32;
    if pNode.iTemp >= 0 {
        return pNode.iTemp; }
    assert(Abc_ObjIsNode( pNode ), "Abc_ObjIsNode( pNode )");
    iLit0 = Abc_NtkClpGia_rec( pNew, Abc_ObjFanin0(pNode) );
    iLit1 = Abc_NtkClpGia_rec( pNew, Abc_ObjFanin1(pNode) );
    iLit0 = Abc_LitNotCond( iLit0, Abc_ObjFaninC0(pNode) );
    iLit1 = Abc_LitNotCond( iLit1, Abc_ObjFaninC1(pNode) );
    return (pNode.iTemp = Gia_ManAppendAnd(pNew, iLit0, iLit1));
}
func Abc_NtkClpGia(pNtk *Abc_Ntk_t) *Gia_Man_t {
    var i, iLit int32;
    var pNew *Gia_Man_t;
    var pNode *Abc_Obj_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( pNtk.pName );
    pNew.pSpec = Abc_UtilStrsav( pNtk.pSpec );
    Abc_NtkForEachObj( pNtk, pNode, i )
        pNode.iTemp = -1;
    Abc_AigConst1(pNtk).iTemp = 1;
    Abc_NtkForEachCi( pNtk, pNode, i )
        pNode.iTemp = Gia_ManAppendCi(pNew);
    Abc_NtkForEachCo( pNtk, pNode, i )
    {
        iLit = Abc_NtkClpGia_rec( pNew, Abc_ObjFanin0(pNode) );
        iLit = Abc_LitNotCond( iLit, Abc_ObjFaninC0(pNode) );
        Gia_ManAppendCo( pNew, iLit );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Minimize SOP by removing redundant variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Original macro:
// #define Abc_NtkSopForEachCube( pSop, nVars, pCube )  for ( pCube = (pSop); *pCube; pCube += (nVars) + 3 )


func Abc_NtkCollapseReduce(vSop *Vec_Str_t, vSupp *Vec_Int_t, vClass *Vec_Int_t, vSupps *Vec_Wec_t) int32 {
    var j int32;
    var i int32;
    var k int32;
    var iCo int32;
    var iVar int32;
    var nVars int32 = Vec_IntSize(vSupp);
    var pCube string;
    var pSop string = Vec_StrArray(vSop);
    var vPres *Vec_Int_t;
    if Vec_StrSize(vSop) == 4 { // constant
    
        Vec_IntForEachEntry( vClass, iCo, i )
            Vec_IntClear( Vec_WecEntry(vSupps, iCo) );
        return 1;
    }
    vPres = Vec_IntStart( nVars );
    // inlining of macro Abc_NtkSopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
        for  k = 0; k < nVars; k++  {
            if pCube[k] != '-' {
                Vec_IntWriteEntry( vPres, k, 1 ); }
        }
    }
    if Vec_IntCountZero(vPres) == 0 {
        Vec_IntFree( vPres );
        return 0;
    }
    // reduce cubes
    // inlining of macro Abc_NtkSopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
        for  k = 0; k < nVars + 3; k++  {
            if k >= nVars || Vec_IntEntry(vPres, k) {
                Vec_StrWriteEntry( vSop, j++, pCube[k] ); }
        }
    }
    Vec_StrWriteEntry( vSop, j++, '\0' );
    Vec_StrShrink( vSop, j );
    // reduce support
    Vec_IntForEachEntry( vClass, iCo, i )
    {
        j = 0;
        vSupp = Vec_WecEntry( vSupps, iCo );
        Vec_IntForEachEntry( vSupp, iVar, k )
            if ( Vec_IntEntry(vPres, k) )
                Vec_IntWriteEntry( vSupp, j++, iVar );
        Vec_IntShrink( vSupp, j );
    }
    Vec_IntFree( vPres );
//    if ( Vec_IntSize(vSupp) != Abc_SopGetVarNum(Vec_StrArray(vSop)) )
//        printf( "Mismatch!!!\n" );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Derives SAT solver for one output from the shared CNF.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkClpDeriveSatSolver(pCnf *Cnf_Dat_t, iCoObjId int32, vSupp *Vec_Int_t, vAnds *Vec_Int_t, vMap *Vec_Int_t, ppSat1 **sat_solver, ppSat2 **sat_solver, ppSat3 **sat_solver) *sat_solver {
    var i int32;
    var k int32;
    var iObj int32;
    var status int32;
    var nVars int32 = 2;
//    int i, k, iObj, status, nVars = 1;
    var vLits *Vec_Int_t  = Vec_IntAlloc( 16 );
    var pSat *sat_solver  = sat_solver_new();
    if ppSat1 { *ppSat1 = sat_solver_new(); }
    if ppSat2 { *ppSat2 = sat_solver_new(); }
    if ppSat3 { *ppSat3 = sat_solver_new(); }
    // assign SAT variable numbers
    Vec_IntWriteEntry( vMap, iCoObjId, nVars++ );
    Vec_IntForEachEntry( vSupp, iObj, k )
        Vec_IntWriteEntry( vMap, iObj, nVars++ );
    Vec_IntForEachEntry( vAnds, iObj, k )
        if ( pCnf.pObj2Clause[iObj] != -1 )
            Vec_IntWriteEntry( vMap, iObj, nVars++ );
//    Vec_IntForEachEntry( vSupp, iObj, k )
//        Vec_IntWriteEntry( vMap, iObj, nVars++ );
    // create clauses for the internal nodes and for the output
    sat_solver_setnvars( pSat, nVars );
    if ppSat1 { sat_solver_setnvars( *ppSat1, nVars ); }
    if ppSat2 { sat_solver_setnvars( *ppSat2, nVars ); }
    if ppSat3 { sat_solver_setnvars( *ppSat3, nVars ); }
    Vec_IntPush( vAnds, iCoObjId );
    Vec_IntForEachEntry( vAnds, iObj, k )
    {
        var iClaBeg int32;
        var iClaEnd int32;
        var pLit *int32;
        if pCnf.pObj2Clause[iObj] == -1 {
            continue; }
        iClaBeg = pCnf.pObj2Clause[iObj];
        iClaEnd = iClaBeg + pCnf.pObj2Count[iObj];
        assert(iClaBeg < iClaEnd, "iClaBeg < iClaEnd");
        for  i = iClaBeg; i < iClaEnd; i++  {
            Vec_IntClear( vLits );
            for  pLit = pCnf.pClauses[i]; pLit < pCnf.pClauses[i+1]; pLit++  {
                Vec_IntPush( vLits, Abc_Lit2LitV(Vec_IntArray(vMap), *pLit) );
            }
            status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits)+Vec_IntSize(vLits) );
            assert((status != 0), "status");
            (void) status;
            if ppSat1 { sat_solver_addclause( *ppSat1, Vec_IntArray(vLits), Vec_IntArray(vLits)+Vec_IntSize(vLits) ); }
            if ppSat2 { sat_solver_addclause( *ppSat2, Vec_IntArray(vLits), Vec_IntArray(vLits)+Vec_IntSize(vLits) ); }
            if ppSat3 { sat_solver_addclause( *ppSat3, Vec_IntArray(vLits), Vec_IntArray(vLits)+Vec_IntSize(vLits) ); }
        }
    }
    Vec_IntPop( vAnds );
    Vec_IntFree( vLits );
    assert(nVars == sat_solver_nvars(pSat), "nVars == sat_solver_nvars(pSat)");
    return pSat;
}

/**Function*************************************************************

  Synopsis    [Computes SOPs for each output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkClpGiaOne(p *Gia_Man_t, iCo int32, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, vSupp *Vec_Int_t, fVerbose int32, vClass *Vec_Int_t, vSupps *Vec_Wec_t) *Vec_Str_t {
    var vSop *Vec_Str_t;
    var clk abctime  = Abc_Clock();
    extern Vec_Str_t * Bmc_CollapseOneOld( Gia_Man_t * p, int nCubeLim, int nBTLimit, int fCanon, int fReverse, int fVerbose );
    var pGia *Gia_Man_t   = Gia_ManDupCones( p, &iCo, 1, 1 );
    if fVerbose != 0 {
        fmt.Printf("Output %4d:  Supp = %4d. Cone =%6d.\n", iCo, Vec_IntSize(vSupp), Gia_ManAndNum(pGia)); }
    vSop = Bmc_CollapseOneOld( pGia, nCubeLim, nBTLimit, fCanon, fReverse, fVerbose );
    Gia_ManStop( pGia );
    if vSop == nil {
        return nil; }
    Abc_NtkCollapseReduce( vSop, vSupp, vClass, vSupps );
    if fVerbose != 0 {
        fmt.Printf("Supp new = %4d. Sop = %4d.  ", Vec_IntSize(vSupp), Vec_StrSize(vSop)/(Vec_IntSize(vSupp) +3)); }
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
    return vSop; 
}
func Abc_NtkClpGiaOne2(pCnf *Cnf_Dat_t, p *Gia_Man_t, iCo int32, nCubeLim int32, nBTLimit int32, fCanon int32, fReverse int32, vSupp *Vec_Int_t, vMap *Vec_Int_t, fVerbose int32, vClass *Vec_Int_t, vSupps *Vec_Wec_t) *Vec_Str_t {
    var vSop *Vec_Str_t;
    var pSat *sat_solver;
    var pSat1 *sat_solver = nil;
    var pSat2 *sat_solver = nil;
    var pSat3 *sat_solver = nil;
    var pObj *Gia_Obj_t;
    var clk abctime  = Abc_Clock();
    extern Vec_Str_t * Bmc_CollapseOne_int( sat_solver * pSat, int nVars, int nCubeLim, int nBTLimit, int fCanon, int fReverse, int fVerbose );
    extern Vec_Str_t * Bmc_CollapseOne_int2( sat_solver * pSat, sat_solver * pSat2, int nVars, int nCubeLim, int nBTLimit, int fCanon, int fReverse, int fVerbose );
    extern Vec_Str_t * Bmc_CollapseOne_int3( sat_solver * pSat, sat_solver * pSat1, sat_solver * pSat2, sat_solver * pSat3, int nVars, int nCubeLim, int nBTLimit, int fCanon, int fReverse, int fVerbose );
    var i int32;
    var iCoObjId int32 = Gia_ObjId( p, Gia_ManCo(p, iCo) );
    var vAnds *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vSuppObjs *Vec_Int_t  = Vec_IntAlloc( 100 );
    Gia_ManForEachCiVec( vSupp, p, pObj, i )
        Vec_IntPush( vSuppObjs, Gia_ObjId(p, pObj) );
    Gia_ManIncrementTravId( p );
    Gia_ManCollectAnds( p, &iCoObjId, 1, vAnds, nil );
    assert(Vec_IntSize(vAnds) > 0, "Vec_IntSize(vAnds) > 0");
//    pSat = Abc_NtkClpDeriveSatSolver( pCnf, iCoObjId, vSuppObjs, vAnds, vMap, &pSat1, &pSat2, &pSat3 );
    pSat = Abc_NtkClpDeriveSatSolver( pCnf, iCoObjId, vSuppObjs, vAnds, vMap, nil, nil, nil );
    Vec_IntFree( vSuppObjs );
    if fVerbose != 0 {
        fmt.Printf("Output %4d:  Supp = %4d. Cone =%6d.\n", iCo, Vec_IntSize(vSupp), Vec_IntSize(vAnds)); }
//    vSop = Bmc_CollapseOne_int3( pSat, pSat1, pSat2, pSat3, Vec_IntSize(vSupp), nCubeLim, nBTLimit, fCanon, fReverse, fVerbose );
//    vSop = Bmc_CollapseOne_int2( pSat, pSat1, Vec_IntSize(vSupp), nCubeLim, nBTLimit, fCanon, fReverse, fVerbose );
    vSop = Bmc_CollapseOne_int( pSat, Vec_IntSize(vSupp), nCubeLim, nBTLimit, fCanon, fReverse, fVerbose );
    sat_solver_delete( pSat );
    if pSat1 != nil { sat_solver_delete( pSat1 ); }
    if pSat2 != nil { sat_solver_delete( pSat2 ); }
    if pSat3 != nil { sat_solver_delete( pSat3 ); }
    Vec_IntFree( vAnds );
    if vSop == nil {
        return nil; }
    Abc_NtkCollapseReduce( vSop, vSupp, vClass, vSupps );
    if fVerbose != 0 {
        fmt.Printf("Supp new = %4d. Sop = %4d.  ", Vec_IntSize(vSupp), Vec_StrSize(vSop)/(Vec_IntSize(vSupp) +3)); }
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
    return vSop; 
}
func Abc_GiaDeriveSops(pNtkNew *Abc_Ntk_t, p *Gia_Man_t, vSupps *Vec_Wec_t, nCubeLim int32, nBTLimit int32, nCostMax int32, fCanon int32, fReverse int32, fCnfShared int32, fVerbose int32) *Vec_Ptr_t {
    var pProgress *ProgressBar;
    var clk abctime  = Abc_Clock();
    var vSops *Vec_Ptr_t = nil;
    var vSopsRepr *Vec_Ptr_t;
    var vReprs, vClass, vReprSuppSizes *Vec_Int_t;
    var i int32;
    var k int32;
    var Entry int32;
    var iCo int32;
    var pOrder *int32;
    var vClasses *Vec_Wec_t;
    var pCnf *Cnf_Dat_t  = nil;
    var vMap *Vec_Int_t  = nil;
    // derive classes of outputs
    vClasses = Gia_ManIsoStrashReduceInt( p, vSupps, 0 );
    if fVerbose != 0 {
        fmt.Printf("Considering %d (out of %d) outputs. ", Vec_WecSize(vClasses), Gia_ManCoNum(p));
        Abc_PrintTime( 1, "Reduction time", Abc_Clock() - clk );
    }
    // derive representatives
    vReprs = Vec_WecCollectFirsts( vClasses );
    vReprSuppSizes = Vec_IntAlloc( Vec_IntSize(vReprs) );
    Vec_IntForEachEntry( vReprs, Entry, i )
        Vec_IntPush( vReprSuppSizes, Vec_IntSize(Vec_WecEntry(vSupps, Entry)) );
    pOrder = Abc_MergeSortCost( Vec_IntArray(vReprSuppSizes), Vec_IntSize(vReprSuppSizes) );
    Vec_IntFree( vReprSuppSizes );
    // consider SOPs for representatives
    if fCnfShared != 0 {
        vMap = Vec_IntStartFull( Gia_ManObjNum(p) );
        pCnf = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 1, 0, 0, 0 );
    }
    vSopsRepr = Vec_PtrStart( Vec_IntSize(vReprs) );
    pProgress = Extra_ProgressBarStart( os.Stdout, Vec_IntSize(vReprs) );
    Extra_ProgressBarUpdate( pProgress, 0, nil );
    for  i = 0; i < Vec_IntSize(vReprs); i++  {
        var iEntry int32         = pOrder[Vec_IntSize(vReprs) - 1 - i];
        var iCoThis int32        = Vec_IntEntry( vReprs, iEntry );
        var vSupp *Vec_Int_t  = Vec_WecEntry( vSupps, iCoThis );
        var vSop *Vec_Str_t;
        if Vec_IntSize(vSupp) < 2 {
            Vec_PtrWriteEntry( vSopsRepr, iEntry, (void *)(ABC_PTRINT_T)1 );
            continue;
        }
        if (fCnfShared != 0) && !(fCanon != 0) {
            vSop = Abc_NtkClpGiaOne2( pCnf, p, iCoThis, nCubeLim, nBTLimit, fCanon, fReverse, vSupp, vMap, i ? 0 : fVerbose, Vec_WecEntry(vClasses, iEntry), vSupps );
        } else {
            vSop = Abc_NtkClpGiaOne( p, iCoThis, nCubeLim, nBTLimit, fCanon, fReverse, vSupp, i ? 0 : fVerbose, Vec_WecEntry(vClasses, iEntry), vSupps ); }
        if vSop == nil {
            goto finish; }
        assert(Vec_IntSize( Vec_WecEntry(vSupps, iCoThis) ) == Abc_SopGetVarNum(Vec_StrArray(vSop)), "Vec_IntSize( Vec_WecEntry(vSupps, iCoThis) ) == Abc_SopGetVarNum(Vec_StrArray(vSop))");
        Extra_ProgressBarUpdate( pProgress, i, nil );
        Vec_PtrWriteEntry( vSopsRepr, iEntry, Abc_SopRegister( (Mem_Flex_t *)pNtkNew.pManFunc, Vec_StrArray(vSop) ) );
        Vec_StrFree( vSop );
    }
    Extra_ProgressBarStop( pProgress );
    if fCnfShared != 0 {
        Cnf_DataFree( pCnf );
        Vec_IntFree( vMap );
    }
    // derive SOPs for each output
    vSops = Vec_PtrStart( Gia_ManCoNum(p) );
    Vec_WecForEachLevel ( vClasses, vClass, i )
        Vec_IntForEachEntry( vClass, iCo, k )
            Vec_PtrWriteEntry( vSops, iCo, Vec_PtrEntry(vSopsRepr, i) );
    assert(Vec_PtrCountZero(vSops) == 0, "Vec_PtrCountZero(vSops) == 0");
/*
    // verify
    for ( i = 0; i < Gia_ManCoNum(p); i++ )
    {
        Vec_Int_t * vSupp = Vec_WecEntry( vSupps, i );
        char * pSop = (char *)Vec_PtrEntry( vSops, i );
        assert( Vec_IntSize(vSupp) == Abc_SopGetVarNum(pSop) );
    }
*/
    // cleanup
finish:
    ABC_FREE( pOrder );
    Vec_IntFree( vReprs );
    Vec_WecFree( vClasses );
    Vec_PtrFree( vSopsRepr );
    return vSops;
}
func Abc_NtkFromSopsInt(pNtk *Abc_Ntk_t, nCubeLim int32, nBTLimit int32, nCostMax int32, fCanon int32, fReverse int32, fCnfShared int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pGia *Gia_Man_t;
    var vSupps *Vec_Wec_t;
    var vSupp *Vec_Int_t;
    var vSops *Vec_Ptr_t;
    var pNode, pNodeNew, pDriver *Abc_Obj_t;
    var i, k, iCi int32; 
    pGia    = Abc_NtkClpGia( pNtk );
    vSupps  = Gia_ManCreateCoSupps( pGia, fVerbose );
    // check the largest output
    if nCubeLim > 0 && nCostMax > 0 {
        var iCoMax int32    = Gia_ManCoLargestSupp( pGia, vSupps );
        var iObjMax int32   = Gia_ObjId( pGia, Gia_ManCo(pGia, iCoMax) );
        var nSuppMax int32  = Vec_IntSize( Vec_WecEntry(vSupps, iCoMax) );
        var nNodeMax int32  = Gia_ManConeSize( pGia, &iObjMax, 1 );
        var Cost word  = (word)nNodeMax * (word)nSuppMax * (word)nCubeLim;
        if Cost > (word)nCostMax {
            fmt.Printf("Cost of the largest output cone exceeded the limit (%d * %d * %d  >  %d).\n", nNodeMax, nSuppMax, nCubeLim, nCostMax);
            Gia_ManStop( pGia );
            Vec_WecFree( vSupps );
            return nil;
        }
    }
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP );
    vSops   = Abc_GiaDeriveSops( pNtkNew, pGia, vSupps, nCubeLim, nBTLimit, nCostMax, fCanon, fReverse, fCnfShared, fVerbose );
    Gia_ManStop( pGia );
    if vSops == nil {
        Vec_WecFree( vSupps );
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    Abc_NtkForEachCo( pNtk, pNode, i )
    {
        pDriver = Abc_ObjFanin0(pNode);
        if Abc_ObjIsCi(pDriver) && !(int32(strings.Compare(Abc_ObjName(pNode), Abc_ObjName(pDriver))) != 0) {
            Abc_ObjAddFanin( pNode.pCopy, pDriver.pCopy );
            continue;
        }
        if Abc_ObjIsCi(pDriver) {
            pNodeNew = Abc_NtkCreateNode( pNtkNew );
            Abc_ObjAddFanin( pNodeNew, pDriver.pCopy ); 
            pNodeNew.pData = Abc_SopRegister( (Mem_Flex_t *)pNtkNew.pManFunc, Abc_ObjFaninC0(pNode) ? "0 1\n" : "1 1\n" );
            Abc_ObjAddFanin( pNode.pCopy, pNodeNew );
            continue;
        }
        if pDriver == Abc_AigConst1(pNtk) {
            pNodeNew = Abc_NtkCreateNode( pNtkNew );
            pNodeNew.pData = Abc_SopRegister( (Mem_Flex_t *)pNtkNew.pManFunc, Abc_ObjFaninC0(pNode) ? " 0\n" : " 1\n" );
            Abc_ObjAddFanin( pNode.pCopy, pNodeNew );
            continue;
        }
        pNodeNew = Abc_NtkCreateNode( pNtkNew );
        vSupp = Vec_WecEntry( vSupps, i );
        Vec_IntForEachEntry( vSupp, iCi, k )
            Abc_ObjAddFanin( pNodeNew, Abc_NtkCi(pNtkNew, iCi) );
        pNodeNew.pData = Abc_SopRegister( (Mem_Flex_t *)pNtkNew.pManFunc, (char*)Vec_PtrEntry( vSops, i ) );
        assert(pNodeNew.pData != (void *)(ABC_PTRINT_T)1, "pNodeNew.pData != (void *)(ABC_PTRINT_T)1");
        Abc_ObjAddFanin( pNode.pCopy, pNodeNew );
    }
    Vec_WecFree( vSupps );
    Vec_PtrFree( vSops );
    Abc_NtkSortSops( pNtkNew );
    return pNtkNew;
}
func Abc_NtkCollapseSat(pNtk *Abc_Ntk_t, nCubeLim int32, nBTLimit int32, nCostMax int32, fCanon int32, fReverse int32, fCnfShared int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pNtkNew = Abc_NtkFromSopsInt( pNtk, nCubeLim, nBTLimit, nCostMax, fCanon, fReverse, fCnfShared, fVerbose );
    if pNtkNew == nil {
        return nil; }
    if pNtk.pExdc {
        pNtkNew.pExdc = Abc_NtkDup( pNtk.pExdc ); }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkCollapseSat: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

