package dsd

/**CFile****************************************************************

  FileName    [dsdLocal.c]

  PackageName [DSD: Disjoint-support decomposition package.]

  Synopsis    [Deriving the local function of the DSD node.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 8.0. Started - September 22, 2003.]

  Revision    [$Id: dsdLocal.c,v 1.0 2002/22/09 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dsdInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                      STATIC VARIABLES                            ///
////////////////////////////////////////////////////////////////////////

// static DdNode * Extra_dsdRemap( DdManager * dd, DdNode * bFunc, st__table * pCache,
//     int * pVar2Form, int * pForm2Var, DdNode * pbCube0[], DdNode * pbCube1[] );

// static DdNode * Extra_bddNodePointedByCube( DdManager * dd, DdNode * bF, DdNode * bC );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the local function of the DSD node. ]

  Description [The local function is computed using the global function 
  of the node and the global functions of the formal inputs. The resulting
  local function is mapped using the topmost N variables of the manager.
  The number of variables N is equal to the number of formal inputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_TreeGetPrimeFunction(dd *DdManager, pNode *Dsd_Node_t) *DdNode {
    var pForm2Var *int32;   // the mapping of each formal input into its first var
    var pVar2Form *int32;   // the mapping of each var into its formal inputs
    var i int32;
    var iVar int32;
    var iLev int32;
    var pPermute *int32;
    var pbCube0, pbCube1 **DdNode;
    var bFunc, bRes, bTemp *DdNode;
    var pCache *st__table;
    
    pPermute  = ABC_ALLOC( int, dd.size );
    pVar2Form = ABC_ALLOC( int, dd.size );
    pForm2Var = ABC_ALLOC( int, dd.size );

    pbCube0 = ABC_ALLOC( DdNode *, dd.size );
    pbCube1 = ABC_ALLOC( DdNode *, dd.size );

    // remap the global function in such a way that
    // the support variables of each formal input are adjacent
    iLev = 0;
    for  i = 0; i < pNode.nDecs; i++  {
        pForm2Var[i] = dd.invperm[i];
        // inlining of macro cuddT
        for  bTemp = pNode.pDecs[i].S; bTemp != b1; bTemp = ((bTemp).type.kids.T)  {
            iVar = dd.invperm[iLev];
            pPermute[bTemp.index] = iVar;
            pVar2Form[iVar] = i;
            iLev++;
        }

        // collect the cubes representing each assignment
        // inlining of macro Cudd_Not
        pbCube0[i] = Extra_bddGetOneCube( dd, ((DdNode *)((ptrint)(pNode.pDecs[i].G) ^ 01)) );
        Cudd_Ref( pbCube0[i] );
        pbCube1[i] = Extra_bddGetOneCube( dd, pNode.pDecs[i].G );
        Cudd_Ref( pbCube1[i] );
    }

    // remap the function
    bFunc = Cudd_bddPermute( dd, pNode.G, pPermute ); Cudd_Ref( bFunc );
    // remap the cube
    for  i = 0; i < pNode.nDecs; i++  {
        pbCube0[i] = Cudd_bddPermute( dd, bTemp = pbCube0[i], pPermute ); Cudd_Ref( pbCube0[i] );
        Cudd_RecursiveDeref( dd, bTemp );
        pbCube1[i] = Cudd_bddPermute( dd, bTemp = pbCube1[i], pPermute ); Cudd_Ref( pbCube1[i] );
        Cudd_RecursiveDeref( dd, bTemp );
    }

    // remap the function
    pCache = st__init_table( st__ptrcmp, st__ptrhash);;
    bRes = Extra_dsdRemap( dd, bFunc, pCache, pVar2Form, pForm2Var, pbCube0, pbCube1 );  Cudd_Ref( bRes );
    st__free_table( pCache );

    Cudd_RecursiveDeref( dd, bFunc );
    for  i = 0; i < pNode.nDecs; i++  {
        Cudd_RecursiveDeref( dd, pbCube0[i] );
        Cudd_RecursiveDeref( dd, pbCube1[i] );
    }
/*
////////////
    // permute the function once again
    // in such a way that i-th var stood for i-th formal input
    for ( i = 0; i < dd->size; i++ )
        pPermute[i] = -1;
    for ( i = 0; i < pNode->nDecs; i++ )
        pPermute[dd->invperm[i]] = i;
    bRes = Cudd_bddPermute( dd, bTemp = bRes, pPermute ); Cudd_Ref( bRes );
    Cudd_RecursiveDeref( dd, bTemp );
////////////
*/
    ABC_FREE(pPermute);
    ABC_FREE(pVar2Form);
    ABC_FREE(pForm2Var);
    ABC_FREE(pbCube0);
    ABC_FREE(pbCube1);

    Cudd_Deref( bRes );
    return bRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_dsdRemap(dd *DdManager, bF *DdNode, pCache *st__table, pVar2Form *int32, pForm2Var *int32, pbCube0 []*DdNode, pbCube1 []*DdNode) *DdNode {
    var bFR, bF0, bF1 *DdNode;
    var bRes0, bRes1, bRes *DdNode;
    var iForm int32;
    
    // inlining of macro Cudd_Regular
    bFR = ((DdNode *)((ptruint)(bF) & ~01));
    // inlining of macro cuddIsConstant
    if  (bFR).index == CUDD_CONST_INDEX {
        return bF; }

    // check the hash-table
    if bFR.ref != 1 {
        if st__lookup( pCache, (char *)bF, (char **)&bRes ) {
            return bRes; }
    }

    // get the formal input
    iForm = pVar2Form[bFR.index];

    // get the nodes pointed to by the cube
    bF0 = Extra_bddNodePointedByCube( dd, bF, pbCube0[iForm] );
    bF1 = Extra_bddNodePointedByCube( dd, bF, pbCube1[iForm] );

    // call recursively for these nodes
    bRes0 = Extra_dsdRemap( dd, bF0, pCache, pVar2Form, pForm2Var, pbCube0, pbCube1 ); Cudd_Ref( bRes0 );
    bRes1 = Extra_dsdRemap( dd, bF1, pCache, pVar2Form, pForm2Var, pbCube0, pbCube1 ); Cudd_Ref( bRes1 );

    // derive the result using ITE
    bRes = Cudd_bddIte( dd, dd.vars[ pForm2Var[iForm] ], bRes1, bRes0 ); Cudd_Ref( bRes );
    Cudd_RecursiveDeref( dd, bRes0 );
    Cudd_RecursiveDeref( dd, bRes1 );

    // add to the hash table
    if bFR.ref != 1 {
        st__insert( pCache, (char *)bF, (char *)bRes ); }
    Cudd_Deref( bRes );
    return bRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_bddNodePointedByCube(dd *DdManager, bF *DdNode, bC *DdNode) *DdNode {
    var bFR, bCR *DdNode;
    var bF0, bF1 *DdNode;
    var bC0, bC1 *DdNode;
    var LevelF, LevelC int32;

    assert(bC != b0, "bC != b0");
    if bC == b1 {
        return bF; }

//    bRes = cuddCacheLookup2( dd, Extra_bddNodePointedByCube, bF, bC );
//    if ( bRes )
//      return bRes;
    // there is no need for caching because this operation is very fast
    // there will no gain reusing the results of this operations
    // instead, it will flush CUDD cache of other useful entries


    // inlining of macro Cudd_Regular
    bFR = ((DdNode *)((ptruint)(bF) & ~01)); 
    // inlining of macro Cudd_Regular
    bCR = ((DdNode *)((ptruint)(bC) & ~01)); 
    // inlining of macro cuddIsConstant
    assert(!((bFR).index == CUDD_CONST_INDEX), "!/*__starting_point_inline_macro:cuddIsConstant__*/((bFR).index == CUDD_CONST_INDEX)");

    LevelF = dd.perm[bFR.index];
    LevelC = dd.perm[bCR.index];

    if LevelF <= LevelC {
        if bFR != bF {
            // inlining of macro Cudd_Not
            // inlining of macro cuddE
            bF0 = ((DdNode *)((ptrint)(((bFR).type.kids.E)) ^ 01));
            // inlining of macro Cudd_Not
            // inlining of macro cuddT
            bF1 = ((DdNode *)((ptrint)(((bFR).type.kids.T)) ^ 01));
        
        } else {
            // inlining of macro cuddE
            bF0 = ((bFR).type.kids.E);
            // inlining of macro cuddT
            bF1 = ((bFR).type.kids.T);
        }
    
    } else {
        bF0 = bF1 = bF;
    }

    if LevelC <= LevelF {
        if bCR != bC {
            // inlining of macro Cudd_Not
            // inlining of macro cuddE
            bC0 = ((DdNode *)((ptrint)(((bCR).type.kids.E)) ^ 01));
            // inlining of macro Cudd_Not
            // inlining of macro cuddT
            bC1 = ((DdNode *)((ptrint)(((bCR).type.kids.T)) ^ 01));
        
        } else {
            // inlining of macro cuddE
            bC0 = ((bCR).type.kids.E);
            // inlining of macro cuddT
            bC1 = ((bCR).type.kids.T);
        }
    
    } else {
        bC0 = bC1 = bC;
    }

    assert(bC0 == b0 || bC1 == b0, "bC0 == b0 || bC1 == b0");
    if bC0 == b0 {
        return Extra_bddNodePointedByCube( dd, bF1, bC1 ); }
    return Extra_bddNodePointedByCube( dd, bF0, bC0 );
}

     

                                                                        

                

                
               
                

                

                                                                        
                                                                                   
 
                                                                      
          
                      
                                  

                    
                                         
                                      

                                                               
                                         

                                                                                
                                                                                           

                         
                                               
                                 
                                        
                                                                                
     
                                                                                               
                                                                                             
                                          

                                                                                               
                                                                                             
                                          

                                            

                                                            
                                                                                                     
                                                                
                                                                                                                
                                         
                                         
     

                 
     
                                                       
                                                                   
                                            
                                                                  
                                                     

                                                                                            
                                         
     

                           
                    
 

      

////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

