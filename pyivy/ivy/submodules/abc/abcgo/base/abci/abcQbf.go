package abci

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [abcQbf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Implementation of a simple QBF solver.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcQbf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "sat/cnf/cnf.h"

//ABC_NAMESPACE_IMPL_START


/*
   Implementation of a simple QBF solver along the lines of
   A. Solar-Lezama, L. Tancau, R. Bodik, V. Saraswat, and S. Seshia, 
   "Combinatorial sketching for finite programs", 12th International 
   Conference on Architectural Support for Programming Languages and 
   Operating Systems (ASPLOS 2006), San Jose, CA, October 2006.
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Abc_NtkModelToVector( Abc_Ntk_t * pNtk, Vec_Int_t * vPiValues );

// static void Abc_NtkVectorClearPars( Vec_Int_t * vPiValues, int nPars );

// static void Abc_NtkVectorClearVars( Abc_Ntk_t * pNtk, Vec_Int_t * vPiValues, int nPars );

// static void Abc_NtkVectorPrintPars( Vec_Int_t * vPiValues, int nPars );

// static void Abc_NtkVectorPrintVars( Abc_Ntk_t * pNtk, Vec_Int_t * vPiValues, int nPars );


// extern int Abc_NtkDSat( Abc_Ntk_t * pNtk, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int nLearnedStart, int nLearnedDelta, int nLearnedPerce, int fAlignPol, int fAndOuts, int fNewSolver, int fVerbose );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Solve the QBF problem EpAx[M(p,x)].]

  Description [Variables p go first, followed by variable x.
  The number of parameters is nPars. The miter is in pNtk.
  The miter expresses EQUALITY of the implementation and spec.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkQbf(pNtk *Abc_Ntk_t, nPars int32, nItersMax int32, fDumpCnf int32, fVerbose int32) {
    var pNtkVer, pNtkSyn, pNtkSyn2, pNtkTemp *Abc_Ntk_t;
    var vPiValues *Vec_Int_t;
    var clkTotal abctime = Abc_Clock();
    var clkS abctime;
    var clkV abctime;
    var nIters, nInputs, RetValue, fFound int32;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(Abc_NtkIsComb(pNtk), "Abc_NtkIsComb(pNtk)");
    assert(Abc_NtkPoNum(pNtk) == 1, "Abc_NtkPoNum(pNtk) == 1");
    assert(nPars > 0 && nPars < Abc_NtkPiNum(pNtk), "nPars > 0 && nPars < Abc_NtkPiNum(pNtk)");
//    assert( Abc_NtkPiNum(pNtk)-nPars < 32 );
    nInputs = Abc_NtkPiNum(pNtk) - nPars;

    if fDumpCnf != 0 {
        // original problem: \exists p \forall x \exists y.  M(p,x,y)
        // negated problem:  \forall p \exists x \exists y. !M(p,x,y)
        extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
        var pMan *Aig_Man_t  = Abc_NtkToDar( pNtk, 0, 0 );
        var pCnf *Cnf_Dat_t  = Cnf_Derive( pMan, 0 );
        var vVarMap, vForAlls, vExists *Vec_Int_t;
        var pObj *Aig_Obj_t;
        var pFileName string;
        var i, Entry int32;
        // create var map
        vVarMap = Vec_IntStart( pCnf.nVars );
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pMan.vCis, i)), 1); i++  {
            if i < nPars {
                Vec_IntWriteEntry( vVarMap, pCnf.pVarNums[Aig_ObjId(pObj)], 1 ); }
        }
        // create various maps
        vForAlls = Vec_IntAlloc( nPars );
        vExists = Vec_IntAlloc( Abc_NtkPiNum(pNtk) - nPars );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vVarMap)) && (((Entry) = Vec_IntEntry(vVarMap, i)), 1); i++  {
            if Entry != 0 {
                Vec_IntPush( vForAlls, i );
            } else {
                Vec_IntPush( vExists, i ); }
        }
        // generate CNF
        pFileName = Extra_FileNameGenericAppend( pNtk.pSpec, ".qdimacs" );
        Cnf_DataWriteIntoFile( pCnf, pFileName, 0, vForAlls, vExists );
        Aig_ManStop( pMan );
        Cnf_DataFree( pCnf );
        Vec_IntFree( vForAlls );
        Vec_IntFree( vExists );
        Vec_IntFree( vVarMap );
        fmt.Printf("The 2QBF formula was written into file \"%s\".\n", pFileName);
        return;
    }

    // initialize the synthesized network with 0000-combination
    vPiValues = Vec_IntStart( Abc_NtkPiNum(pNtk) );

    // create random init value
    {
    var i int32;
    srand( time.Now().Unix() );
    for  i = nPars; i < Abc_NtkPiNum(pNtk); i++  {
        Vec_IntWriteEntry( vPiValues, i, rand() & 1 );
    }
    }

    Abc_NtkVectorClearPars( vPiValues, nPars );
    pNtkSyn = Abc_NtkMiterCofactor( pNtk, vPiValues );
    if fVerbose != 0 {
        fmt.Printf("Iter %2d : ", 0);
        fmt.Printf("AIG = %6d  ", Abc_NtkNodeNum(pNtkSyn));
        Abc_NtkVectorPrintVars( pNtk, vPiValues, nPars );
        fmt.Printf("\n");
    }

    // iteratively solve
    for  nIters = 0; nIters < nItersMax; nIters++  {
        // solve the synthesis instance
clkS = Abc_Clock();
//        RetValue = Abc_NtkMiterSat( pNtkSyn, 0, 0, 0, NULL, NULL );
        RetValue = Abc_NtkDSat( pNtkSyn, (ABC_INT64_T)0, (ABC_INT64_T)0, 0, 0, 0, 1, 0, 0, 0 );
clkS = Abc_Clock() - clkS;
        if RetValue == 0 {
            Abc_NtkModelToVector( pNtkSyn, vPiValues ); }
        if RetValue == 1 {
            break;
        }
        if RetValue == -1 {
            fmt.Printf("Synthesis timed out.\n");
            break;
        }
        // there is a counter-example

        // construct the verification instance
        Abc_NtkVectorClearVars( pNtk, vPiValues, nPars );
        pNtkVer = Abc_NtkMiterCofactor( pNtk, vPiValues );
        // complement the output
        Abc_ObjXorFaninC( Abc_NtkPo(pNtkVer,0), 0 );

        // solve the verification instance
clkV = Abc_Clock();
        RetValue = Abc_NtkMiterSat( pNtkVer, 0, 0, 0, nil, nil );
clkV = Abc_Clock() - clkV;
        if RetValue == 0 {
            Abc_NtkModelToVector( pNtkVer, vPiValues ); }
        Abc_NtkDelete( pNtkVer );
        if RetValue == 1 {
            fFound = 1;
            break;
        }
        if RetValue == -1 {
            fmt.Printf("Verification timed out.\n");
            break;
        }
        // there is a counter-example

        // create a new synthesis network
        Abc_NtkVectorClearPars( vPiValues, nPars );
        pNtkSyn2 = Abc_NtkMiterCofactor( pNtk, vPiValues );
        // add to the synthesis instance
        pNtkSyn = Abc_NtkMiterAnd( pNtkTemp = pNtkSyn, pNtkSyn2, 0, 0 );
        Abc_NtkDelete( pNtkSyn2 );
        Abc_NtkDelete( pNtkTemp );

        if fVerbose != 0 {
            fmt.Printf("Iter %2d : ", nIters+1);
            fmt.Printf("AIG = %6d  ", Abc_NtkNodeNum(pNtkSyn));
            Abc_NtkVectorPrintVars( pNtk, vPiValues, nPars );
            fmt.Printf("  ");
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Syn"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((clkS)))/(float64(int64(time.Second))));;
//            ABC_PRT( "Ver", clkV );
        }
    }
    Abc_NtkDelete( pNtkSyn );
    // report the results
    if fFound != 0 {
        var nZeros int32  = Vec_IntCountZero( vPiValues );
        fmt.Printf("Parameters: ");
        Abc_NtkVectorPrintPars( vPiValues, nPars );
        fmt.Printf("  Statistics: 0=%d 1=%d\n", nZeros, Vec_IntSize(vPiValues) - nZeros);
        fmt.Printf("Solved after %d iterations.  ", nIters);
    
    } else if nIters == nItersMax {
        fmt.Printf("Quit after %d iterations.  ", nItersMax);
    } else {
        fmt.Printf("Implementation does not exist.  "); }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    Vec_IntFree( vPiValues );
}


/**Function*************************************************************

  Synopsis    [Translates model into the vector of values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkModelToVector(pNtk *Abc_Ntk_t, vPiValues *Vec_Int_t) {
    var pModel *int32;
    var i int32;
    pModel = pNtk.pModel;
    for  i = 0; i < Abc_NtkPiNum(pNtk); i++  {
        Vec_IntWriteEntry( vPiValues, i, pModel[i] );
    }
}

/**Function*************************************************************

  Synopsis    [Clears parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVectorClearPars(vPiValues *Vec_Int_t, nPars int32) {
    var i int32;
    for  i = 0; i < nPars; i++  {
        Vec_IntWriteEntry( vPiValues, i, -1 );
    }
}

/**Function*************************************************************

  Synopsis    [Clears variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVectorClearVars(pNtk *Abc_Ntk_t, vPiValues *Vec_Int_t, nPars int32) {
    var i int32;
    for  i = nPars; i < Abc_NtkPiNum(pNtk); i++  {
        Vec_IntWriteEntry( vPiValues, i, -1 );
    }
}

/**Function*************************************************************

  Synopsis    [Clears variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVectorPrintPars(vPiValues *Vec_Int_t, nPars int32) {
    var i int32;
    for  i = 0; i < nPars; i++  {
        fmt.Printf("%d", Vec_IntEntry(vPiValues,i));
    }
}

/**Function*************************************************************

  Synopsis    [Clears variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVectorPrintVars(pNtk *Abc_Ntk_t, vPiValues *Vec_Int_t, nPars int32) {
    var i int32;
    for  i = nPars; i < Abc_NtkPiNum(pNtk); i++  {
        fmt.Printf("%d", Vec_IntEntry(vPiValues,i));
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

