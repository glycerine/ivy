package sbd

import "fmt"

/**CFile****************************************************************

  FileName    [sbd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_CountConfigVars(vSet *Vec_Int_t, nVars int32, Degree int32) int32 {
    var i, k, Entry, Entry2, Count, Below int32;
    var Prev int32  = Vec_IntEntry( vSet, 0 );
    Vec_IntForEachEntryStart( vSet, Entry, i, 1 )
    {
        assert(Degree*Prev >= Entry, "Degree*Prev >= Entry");
        if Degree*Prev == Entry {
            Prev = Entry;
            continue;
        }
        Below = nVars; 
        Vec_IntForEachEntryStart( vSet, Entry2, k, i )
            Below += Entry2;
        Count += Below * (Degree*Prev - 1);
        Prev = Entry;
    }
    Count += nVars * Degree*Prev;
    if Vec_IntSum(vSet) < nVars - 1 {
        return 0;
    } else {
        return Count;
    }
}
func Sbd_CountTopos() {
    var nInputs int32   =  9;
    var nNodes int32    = 10;
    var Degree int32    =  3;
    var fVerbose int32  =  1;
    var p *Hsh_VecMan_t  = Hsh_VecManStart( 10000 ); // hash table for arrays
    var vSet *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i, k, e, Start, Stop int32;
    fmt.Printf("Counting topologies for %d inputs and %d degree-%d nodes.\n", nInputs, nNodes, Degree);
    Start = Hsh_VecManAdd( p, vSet );
    for  i = 1; i <= nNodes; i++  {
        Stop = Hsh_VecSize( p );
        for  e = Start; e < Stop; e++  {
            var vTemp *Vec_Int_t  = Hsh_VecReadEntry( p, e );
            Vec_IntClear( vSet );
            Vec_IntAppend( vSet, vTemp );
            for  k = 0; k < Vec_IntSize(vSet); k++  {
                // skip if the number of entries on this level is equal to the number of fanins on the previous level
                if k ? (Vec_IntEntry(vSet, k) == Degree*Vec_IntEntry(vSet, k-1)) : (Vec_IntEntry(vSet, 0) > 0) {
                    continue; }
                Vec_IntAddToEntry( vSet, k, 1 );
                Hsh_VecManAdd( p, vSet );
                Vec_IntAddToEntry( vSet, k, -1 );
            }
            Vec_IntPush( vSet, 1 );
            Hsh_VecManAdd( p, vSet );
        }
        fmt.Printf("Nodes = %2d : This = %8d  All = %8d\n", i, Hsh_VecSize(p) - Stop, Hsh_VecSize(p));
        if fVerbose != 0 {
            for  e = Stop; e < Hsh_VecSize(p); e++  {
                var vTemp *Vec_Int_t  = Hsh_VecReadEntry( p, e );
                fmt.Printf("Params = %3d.  ", Sbd_CountConfigVars(vTemp, nInputs, Degree));
                Vec_IntPrint( vTemp );
            }
            fmt.Printf("\n");
        }
        Start = Stop;
    }
    Vec_IntFree( vSet );
    Hsh_VecManStop( p );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

