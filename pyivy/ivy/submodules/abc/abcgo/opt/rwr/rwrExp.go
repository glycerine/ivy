package rwr

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [rwrExp.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [Computation of practically used NN-classes of 4-input cuts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwrExp.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Rwr_Man4_t struct {
    // internal lookups
    nFuncs                int32;           // the number of four-var functions
    unsigned short *   puCanons;         // canonical forms
    pnCounts               *int32;         // the counters of functions in each class
    nConsidered                int32;      // the number of nodes considered
    nClasses                int32;         // the number of NN classes
}

   
type Rwr_Man5_t struct {
    // internal lookups
    tTableNN        *stmm_table;         // the NN canonical forms
    tTableNPN        *stmm_table;        // the NPN canonical forms
}

var s_pManRwrExp4 *Rwr_Man4_t = nil;
var s_pManRwrExp5 *Rwr_Man5_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collects stats about 4-var functions appearing in netlists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_Man4ExploreStart() {
    var p *Rwr_Man4_t;
    // inlining of macro ABC_ALLOC
    p = ((Rwr_Man4_t *) malloc(sizeof(Rwr_Man4_t) * (size_t)(1)));
    memset( p, 0, sizeof(Rwr_Man4_t) );
    // canonical forms
    p.nFuncs    = (1<<16);
    // canonical forms, phases, perms
    Extra_Truth4VarNPN( &p.puCanons, nil, nil, nil );
    // counters
    // inlining of macro ABC_ALLOC
    p.pnCounts  = ((int *) malloc(sizeof(int) * (size_t)(p.nFuncs)));
    memset( p.pnCounts, 0, sizeof(int) * p.nFuncs );
    s_pManRwrExp4 = p;
}

/**Function*************************************************************

  Synopsis    [Collects stats about 4-var functions appearing in netlists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_Man4ExploreCount(uTruth uint32) {
    assert(uTruth < (1<<16), "uTruth < (1<<16)");
    s_pManRwrExp4.pnCounts[ s_pManRwrExp4.puCanons[uTruth] ]++;    
}

/**Function*************************************************************

  Synopsis    [Collects stats about 4-var functions appearing in netlists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_Man4ExplorePrint() {
    var pFile *os.File;
    var i, CountMax, CountWrite, nCuts, nClasses int32;
    var pDistrib *int32;
    var pReprs *int32;
    // find the max number of occurences
    nCuts = nClasses = 0;
    CountMax = 0;
    for  i = 0; i < s_pManRwrExp4.nFuncs; i++  {
        if CountMax < s_pManRwrExp4.pnCounts[i] {
            CountMax = s_pManRwrExp4.pnCounts[i]; }
        nCuts += s_pManRwrExp4.pnCounts[i];
        if s_pManRwrExp4.pnCounts[i] > 0 {
            nClasses++; }
    }
    fmt.Printf("Number of cuts considered       = %8d.\n", nCuts);
    fmt.Printf("Classes occurring at least once = %8d.\n", nClasses);
    // print the distribution of classes
    // inlining of macro ABC_ALLOC
    pDistrib = ((int *) malloc(sizeof(int) * (size_t)(CountMax + 1)));
    // inlining of macro ABC_ALLOC
    pReprs   = ((int *) malloc(sizeof(int) * (size_t)(CountMax + 1)));
    memset( pDistrib, 0, sizeof(int)*(CountMax + 1) );
    for  i = 0; i < s_pManRwrExp4.nFuncs; i++  {
        pDistrib[ s_pManRwrExp4.pnCounts[i] ]++;
        pReprs[ s_pManRwrExp4.pnCounts[i] ] = i;
    }

    fmt.Printf("Occurence = %6d.  Num classes = %4d.  \n", 0, 2288-nClasses);
    for  i = 1; i <= CountMax; i++  {
        if pDistrib[i] {
            fmt.Printf("Occurence = %6d.  Num classes = %4d.  Repr = ", i, pDistrib[i]);
            Extra_PrintBinary( os.Stdout, (unsigned*)&(pReprs[i]), 16 ); 
            fmt.Printf("\n");
        }
    }
    // inlining of macro ABC_FREE
    ((pDistrib) ? (free((char *) (pDistrib)), (pDistrib) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pReprs) ? (free((char *) (pReprs)), (pReprs) = 0) : 0);
    // write into a file all classes above limit (5)
    CountWrite = 0;
    pFile = os.OpenFile("npnclass_stats4.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    for  i = 0; i < s_pManRwrExp4.nFuncs; i++  {
        if s_pManRwrExp4.pnCounts[i] > 0 {
            Extra_PrintHex( pFile, (unsigned *)&i, 4 );
            fmt.Fprintf(pFile, " %10d\n", s_pManRwrExp4.pnCounts[i]);
//            fprintf( pFile, "%d ", i );
            CountWrite++;
        }
    }
    _ = (pFile).Close();
    fmt.Printf("%d classes written into file \"%s\".\n", CountWrite, "npnclass_stats4.txt");
}




/**Function*************************************************************

  Synopsis    [Collects stats about 4-var functions appearing in netlists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_Man5ExploreStart() {
    var p *Rwr_Man5_t;
    // inlining of macro ABC_ALLOC
    p = ((Rwr_Man5_t *) malloc(sizeof(Rwr_Man5_t) * (size_t)(1)));
    memset( p, 0, sizeof(Rwr_Man5_t) );
    p.tTableNN  = stmm_init_table( st__numcmp, st__numhash );
    p.tTableNPN = stmm_init_table( st__numcmp, st__numhash );
    s_pManRwrExp5 = p;

//Extra_PrintHex( stdout, Extra_TruthCanonNPN( 0x0000FFFF, 5 ), 5 );
//printf( "\n" );
}

/**Function*************************************************************

  Synopsis    [Collects stats about 4-var functions appearing in netlists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_Man5ExploreCount(uTruth uint32) {
    var pCounter *int32;
    if !stmm_find_or_add( s_pManRwrExp5.tTableNN, (char *)(ABC_PTRUINT_T)uTruth, (char***)&pCounter ) {
        *pCounter = 0; }
    (*pCounter)++;
}

/**Function*************************************************************

  Synopsis    [Collects stats about 4-var functions appearing in netlists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_Man5ExplorePrint() {
    var pFile *os.File;
    var gen *stmm_generator;
    var i, CountMax, nCuts, Counter int32;
    var pDistrib *int32;
    var pReprs *uint32;
    var uTruth, uTruthC uint32;
    var clk abctime  = Abc_Clock();
    var vClassesNN, vClassesNPN *Vec_Int_t;

    // find the max number of occurences
    nCuts = 0;
    CountMax = 0;
    // inlining of macro stmm_foreach_item
    for gen=stmm_init_gen(s_pManRwrExp5.tTableNN); stmm_gen(gen,(char **)&uTruth,(char **)&Counter) || (stmm_free_gen(gen),0); {
        nCuts += Counter;
        if CountMax < Counter {
            CountMax = Counter; }
    }
    fmt.Printf("Number of cuts considered       = %8d.\n", nCuts);
    // inlining of macro stmm_count
    fmt.Printf("Classes occurring at least once = %8d.\n", ((s_pManRwrExp5.tTableNN).num_entries));
    fmt.Printf("The largest number of occurence = %8d.\n", CountMax);

    // print the distribution of classes
    // inlining of macro ABC_ALLOC
    pDistrib = ((int *) malloc(sizeof(int) * (size_t)(CountMax + 1)));
    // inlining of macro ABC_ALLOC
    pReprs   = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(CountMax + 1)));
    memset( pDistrib, 0, sizeof(int)*(CountMax + 1) );
    // inlining of macro stmm_foreach_item
    for gen=stmm_init_gen(s_pManRwrExp5.tTableNN); stmm_gen(gen,(char **)&uTruth,(char **)&Counter) || (stmm_free_gen(gen),0); {
        assert(Counter <= CountMax, "Counter <= CountMax");
        pDistrib[ Counter ]++;
        pReprs[ Counter ] = uTruth;
    }

    for  i = 1; i <= CountMax; i++  {
        if pDistrib[i] {
            fmt.Printf("Occurence = %6d.  Num classes = %4d.  Repr = ", i, pDistrib[i]);
            Extra_PrintBinary( os.Stdout, pReprs + i, 32 ); 
            fmt.Printf("\n");
        }
    }
    // inlining of macro ABC_FREE
    ((pDistrib) ? (free((char *) (pDistrib)), (pDistrib) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pReprs) ? (free((char *) (pReprs)), (pReprs) = 0) : 0);


    // put them into an array
    // inlining of macro stmm_count
    vClassesNN = Vec_IntAlloc( ((s_pManRwrExp5.tTableNN).num_entries) );
    // inlining of macro stmm_foreach_item
    for gen=stmm_init_gen(s_pManRwrExp5.tTableNN); stmm_gen(gen,(char **)&uTruth,nil) || (stmm_free_gen(gen),0); {
        Vec_IntPush( vClassesNN, int32(uTruth) );
    }
    Vec_IntSortUnsigned( vClassesNN );

    // write into a file all classes
    pFile = os.OpenFile("nnclass_stats5.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClassesNN)) && (((uTruth) = Vec_IntEntry(vClassesNN, i)), 1); i++  {
        if !stmm_lookup( s_pManRwrExp5.tTableNN, (char *)(ABC_PTRUINT_T)uTruth, (char **)&Counter ) {
            assert(false, "0");
        }
        Extra_PrintHex( pFile, &uTruth, 5 );
        fmt.Fprintf(pFile, " %10d\n", Counter);
    }
    _ = (pFile).Close();
    fmt.Printf("%d classes written into file \"%s\".\n", vClassesNN.nSize, "nnclass_stats5.txt");


clk = Abc_Clock();
    // how many NPN classes exist?
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClassesNN)) && (((uTruth) = Vec_IntEntry(vClassesNN, i)), 1); i++  {
        var pCounter *int32;
        uTruthC = Extra_TruthCanonNPN( uTruth, 5 );
        if !stmm_find_or_add( s_pManRwrExp5.tTableNPN, (char *)(ABC_PTRUINT_T)uTruthC, (char***)&pCounter ) {
            *pCounter = 0; }
        if !stmm_lookup( s_pManRwrExp5.tTableNN, (char *)(ABC_PTRUINT_T)uTruth, (char **)&Counter ) {
            assert(false, "0");
        }
        (*pCounter) += Counter;
    }
    // inlining of macro stmm_count
    fmt.Printf("The numbe of NPN classes = %d.\n", ((s_pManRwrExp5.tTableNPN).num_entries));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Computing NPN classes"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;

    // put them into an array
    // inlining of macro stmm_count
    vClassesNPN = Vec_IntAlloc( ((s_pManRwrExp5.tTableNPN).num_entries) );
    // inlining of macro stmm_foreach_item
    for gen=stmm_init_gen(s_pManRwrExp5.tTableNPN); stmm_gen(gen,(char **)&uTruth,nil) || (stmm_free_gen(gen),0); {
        Vec_IntPush( vClassesNPN, int32(uTruth) );
    }
    Vec_IntSortUnsigned( vClassesNPN );

    // write into a file all classes
    pFile = os.OpenFile("npnclass_stats5.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClassesNPN)) && (((uTruth) = Vec_IntEntry(vClassesNPN, i)), 1); i++  {
        if !stmm_lookup( s_pManRwrExp5.tTableNPN, (char *)(ABC_PTRUINT_T)uTruth, (char **)&Counter ) {
            assert(false, "0");
        }
        Extra_PrintHex( pFile, &uTruth, 5 );
        fmt.Fprintf(pFile, " %10d\n", Counter);
    }
    _ = (pFile).Close();
    fmt.Printf("%d classes written into file \"%s\".\n", vClassesNPN.nSize, "npnclass_stats5.txt");


    // can they be uniquely characterized?

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

