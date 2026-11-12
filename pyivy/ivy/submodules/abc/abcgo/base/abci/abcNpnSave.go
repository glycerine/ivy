package abci

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [abcNpnSave.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface with the FPGA mapping package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: abcNpnSave.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "aig/aig/aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
type Npn_Obj_t = Npn_Obj_t_;
type Npn_Man_t = Npn_Man_t_;

type Npn_Obj_t_ struct {
    uTruth         word;      // truth table
    Count          int32;       // occurrences
    iNext          int32;       // next entry
}
type Npn_Man_t_ struct {
    pBuffer   *Npn_Obj_t;     // all NPN entries
    pBins         *int32;       // hash table
    nBins          int32;       // hash table size
    nBufferSize          int32; // buffer size
    nEntries          int32;    // entry count
}

func Npn_ManObj(p *Npn_Man_t, i int32) *Npn_Obj_t { assert(i < p.nBufferSize, "i < p.nBufferSize"); if i != 0 {
    return p.pBuffer + i;
} else {
    return nil;
}  }
func Npn_ManObjNum(p *Npn_Man_t, pObj *Npn_Obj_t) int32 { assert(p.pBuffer < pObj, "p.pBuffer < pObj");  return pObj - p.pBuffer;          }

var Truth [8]word = [8]word{
    ABC_CONST(0xAAAAAAAAAAAAAAAA),
    ABC_CONST(0xCCCCCCCCCCCCCCCC),
    ABC_CONST(0xF0F0F0F0F0F0F0F0),
    ABC_CONST(0xFF00FF00FF00FF00),
    ABC_CONST(0xFFFF0000FFFF0000),
    ABC_CONST(0xFFFFFFFF00000000),
    ABC_CONST(0x0000000000000000),
    ABC_CONST(0xFFFFFFFFFFFFFFFF),
};

var pNpnMan *Npn_Man_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Npn_TruthPermute_rec_count int32;
func Npn_TruthPermute_rec(pStr string, mid int32, end int32) {
    
    var pTemp string  = Abc_UtilStrsav(pStr);
    var e byte;
    var i int32;
    if mid == end {
        fmt.Printf("%03d: %s\n", spStatic_Npn_TruthPermute_rec_count++, pTemp);
        return ;
    }
    for  i = mid; i <= end; i++  {
        e = pTemp[mid];
        pTemp[mid] = pTemp[i];
        pTemp[i] = e;

        Npn_TruthPermute_rec( pTemp, mid + 1, end );

        e = pTemp[mid];
        pTemp[mid] = pTemp[i];
        pTemp[i] = e;
    }
    // inlining of macro ABC_FREE
    ((pTemp) ? (free((char *) (pTemp)), (pTemp) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthHasVar(t word, v int32) int32 {
    return ((t & Truth[v]) >> (1<<v)) != (t & ~Truth[v]);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthSupport(t word) int32 {
    var v, Supp int32;
    for  v = 0; v < 6; v++  {
        if Npn_TruthHasVar( t, v ) != 0 {
            Supp |= (1 << v); }
    }
    return Supp;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthSuppSize(t word, nVars int32) int32 {
    var v, nSupp int32;
    assert(nVars <= 6, "nVars <= 6");
    for  v = 0; v < nVars; v++  {
        if Npn_TruthHasVar( t, v ) != 0 {
            nSupp++; }
    }
    return nSupp;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthIsMinBase(t word) int32 {
    var Supp int32  = Npn_TruthSupport(t);
    return (Supp & (Supp+1)) == 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthPadWord(uTruth word, nVars int32) word {
    if nVars == 6 {
        return uTruth; }
    if nVars <= 5 {
        uTruth = ((uTruth & ABC_CONST(0x00000000FFFFFFFF)) << 32) | (uTruth & ABC_CONST(0x00000000FFFFFFFF)); }
    if nVars <= 4 {
        uTruth = ((uTruth & ABC_CONST(0x0000FFFF0000FFFF)) << 16) | (uTruth & ABC_CONST(0x0000FFFF0000FFFF)); }
    if nVars <= 3 {
        uTruth = ((uTruth & ABC_CONST(0x00FF00FF00FF00FF)) <<  8) | (uTruth & ABC_CONST(0x00FF00FF00FF00FF)); }
    if nVars <= 2 {
        uTruth = ((uTruth & ABC_CONST(0x0F0F0F0F0F0F0F0F)) <<  4) | (uTruth & ABC_CONST(0x0F0F0F0F0F0F0F0F)); }
    if nVars <= 1 {
        uTruth = ((uTruth & ABC_CONST(0x3333333333333333)) <<  2) | (uTruth & ABC_CONST(0x3333333333333333)); }
    if nVars == 0 {
        uTruth = ((uTruth & ABC_CONST(0x5555555555555555)) <<  1) | (uTruth & ABC_CONST(0x5555555555555555)); }
    return uTruth;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthCountOnes(t word) int32 {
    t =    (t & ABC_CONST(0x5555555555555555)) + ((t>> 1) & ABC_CONST(0x5555555555555555));
    t =    (t & ABC_CONST(0x3333333333333333)) + ((t>> 2) & ABC_CONST(0x3333333333333333));
    t =    (t & ABC_CONST(0x0F0F0F0F0F0F0F0F)) + ((t>> 4) & ABC_CONST(0x0F0F0F0F0F0F0F0F));
    t =    (t & ABC_CONST(0x00FF00FF00FF00FF)) + ((t>> 8) & ABC_CONST(0x00FF00FF00FF00FF));
    t =    (t & ABC_CONST(0x0000FFFF0000FFFF)) + ((t>>16) & ABC_CONST(0x0000FFFF0000FFFF));
    return (t & ABC_CONST(0x00000000FFFFFFFF)) +  (t>>32);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthChangePhase(t word, v int32) word {
    return ((t & Truth[v]) >> (1<<v)) | ((t & ~Truth[v]) << (1<<v));
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthSwapAdjacentVars(t word, v int32) word {
    static word PMasks[5][3] = {
        { ABC_CONST(0x9999999999999999), ABC_CONST(0x2222222222222222), ABC_CONST(0x4444444444444444) },
        { ABC_CONST(0xC3C3C3C3C3C3C3C3), ABC_CONST(0x0C0C0C0C0C0C0C0C), ABC_CONST(0x3030303030303030) },
        { ABC_CONST(0xF00FF00FF00FF00F), ABC_CONST(0x00F000F000F000F0), ABC_CONST(0x0F000F000F000F00) },
        { ABC_CONST(0xFF0000FFFF0000FF), ABC_CONST(0x0000FF000000FF00), ABC_CONST(0x00FF000000FF0000) },
        { ABC_CONST(0xFFFF00000000FFFF), ABC_CONST(0x00000000FFFF0000), ABC_CONST(0x0000FFFF00000000) }
    };
    assert(v < 6, "v < 6");
    return (t & PMasks[v][0]) | ((t & PMasks[v][1]) << (1 << v)) | ((t & PMasks[v][2]) >> (1 << v));
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_TruthCanon(t word, nVars int32, pPhase *int32) word {
    var fUsePolarity int32;
    var fUsePermutation int32;
    var Temp byte;
    var pSigs [13]byte;
    var pCanonPerm [6]byte;
    var v, fChange, CanonPhase int32;
    assert(nVars < 7, "nVars < 7");
    pSigs[12] = Npn_TruthCountOnes( t );
    if pSigs[12] > 32 {
        t = ~t;
        pSigs[12] = 64 - pSigs[12];
        CanonPhase |= (1 << 6);
    }
    if (fUsePolarity != 0) || (fUsePermutation != 0) {
        for  v = 0; v < nVars; v++  {
            pCanonPerm[v] = v;
            pSigs[2*v+1] = Npn_TruthCountOnes( t & Truth[v] );
            pSigs[2*v] = pSigs[12] - pSigs[2*v+1];
        }
    }
    if fUsePolarity != 0 {
        for  v = 0; v < nVars; v++  {
            if pSigs[2*v] >= pSigs[2*v+1] {
                continue; }
            CanonPhase |= (1 << v);
            Temp = pSigs[2*v];
            pSigs[2*v] = pSigs[2*v+1];
            pSigs[2*v+1] = Temp;
            t = Npn_TruthChangePhase( t, v );
        }
    }
    if fUsePermutation != 0 {
        do {
            fChange = 0;
            for  v = 0; v < nVars-1; v++  {
                if fUsePolarity != 0 {
                    if pSigs[2*v] >= pSigs[2*(v+1)] {
                        continue; }
                
                } else {
                    if Abc_MinInt(pSigs[2*v],pSigs[2*v+1]) >= Abc_MinInt(pSigs[2*(v+1)],pSigs[2*(v+1)+1]) {
                        continue; }
                }
                fChange = 1;

                Temp = pCanonPerm[v];
                pCanonPerm[v] = pCanonPerm[v+1];
                pCanonPerm[v+1] = Temp;

                Temp = pSigs[2*v];
                pSigs[2*v] = pSigs[2*(v+1)];
                pSigs[2*(v+1)] = Temp;

                Temp = pSigs[2*v+1];
                pSigs[2*v+1] = pSigs[2*(v+1)+1];
                pSigs[2*(v+1)+1] = Temp;

                t = Npn_TruthSwapAdjacentVars( t, v );
            }
        } while ( fChange );
    }
    if pPhase != nil {
        *pPhase = 0;
        for  v = 0; v < nVars; v++  {
            *pPhase |= (pCanonPerm[v] << (4 * v));
        }
        *pPhase |= (CanonPhase << 24);
    }
    return t;
}


/**Function*************************************************************

  Synopsis    [Computes the hash key.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManHash(p *Npn_Man_t, uTruth word) int32 {
    var Key word  = (uTruth * (word)101) ^ (uTruth * (word)733) ^ (uTruth * (word)1777);
    return int32((Key % (word)p.nBins));
}

/**Function*************************************************************

  Synopsis    [Resizes the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManResize(p *Npn_Man_t) {
    var pEntry, pNext *Npn_Obj_t;
    var pBinsOld, ppPlace *int32;
    var nBinsOld, Counter, i int32;
    var clk abctime;
    assert(p.pBins != nil, "p.pBins != nil");
clk = Abc_Clock();
    // save the old Bins
    pBinsOld = p.pBins;
    nBinsOld = p.nBins;
    // get the new Bins
    p.nBins = Abc_PrimeCudd( 3 * nBinsOld ); 
    // inlining of macro ABC_CALLOC
    p.pBins = ((int *) calloc((size_t)(p.nBins), sizeof(int)));
    // rehash the entries from the old table
    Counter = 1;
    for  i = 0; i < nBinsOld; i++  {
    for  pEntry = Npn_ManObj(p, pBinsOld[i]), 
          pNext = pEntry ? Npn_ManObj(p, pEntry.iNext) : nil; 
          (pEntry != nil); 
          pEntry = pNext, 
          pNext = pEntry ? Npn_ManObj(p, pEntry.iNext) : nil  {
        // get the place where this entry goes 
        ppPlace = p.pBins + Npn_ManHash( p, pEntry.uTruth );
        // add the entry to the list
        pEntry.iNext = *ppPlace;
        *ppPlace = Npn_ManObjNum( p, pEntry );
        Counter++;
    }
    }
    assert(Counter == p.nEntries, "Counter == p.nEntries");
    // inlining of macro ABC_FREE
    ((pBinsOld) ? (free((char *) (pBinsOld)), (pBinsOld) = 0) : 0);
//ABC_PRT( "Hash table resizing time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    [Adds one entry to the table.]

  Description [Increments ref counter by 1.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManAdd(p *Npn_Man_t, uTruth word) *Npn_Obj_t {
    var pEntry *Npn_Obj_t;
    var pPlace *int32;
    var Key int32 = Npn_ManHash( p, uTruth );
    // resize the link storage if needed
    if p.nEntries == p.nBufferSize {
        p.nBufferSize *= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pBuffer {
            p.pBuffer = ((Npn_Obj_t *) realloc((char *)(p.pBuffer), sizeof(Npn_Obj_t) * (size_t)(p.nBufferSize)));
        } else {
            p.pBuffer = ((Npn_Obj_t *) malloc(sizeof(Npn_Obj_t) * (size_t)(p.nBufferSize)));
        }
    }
    // find the entry
    for  pEntry = Npn_ManObj(p, p.pBins[Key]), 
          pPlace = p.pBins + Key; 
          (pEntry != nil); 
          pPlace = &pEntry.iNext, 
          pEntry = Npn_ManObj(p, pEntry.iNext)  {
        if pEntry.uTruth == uTruth {
            pEntry.Count++;
            return pEntry;
        }
    }
    // create new entry
    *pPlace = p.nEntries;
    assert(p.nEntries < p.nBufferSize, "p.nEntries < p.nBufferSize");
    pEntry = Npn_ManObj( p, p.nEntries++ );
    pEntry.uTruth = uTruth;
    pEntry.Count = 1;
    pEntry.iNext = 0;
    // resize the table if needed
    if p.nEntries > 3 * p.nBins {
        Npn_ManResize( p ); }
    return pEntry;
}

/**Function*************************************************************

  Synopsis    [Fills table from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManRead(p *Npn_Man_t, pFileName string) {
    var pBuffer [1000]byte;
    var pToken string;
    var pEntry *Npn_Obj_t;
    var Truth [2]uint32;
    var uTruth word;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        Abc_Print( -1, "Cannot open NPN function file \"%s\".\n", pFileName );
        return;
    }
    // read lines from the file
    for  fgets( pBuffer, 1000, pFile ) != nil  {
        pToken = strtok( pBuffer, " \t\n" );
        if pToken == nil {
            continue; }
        if pToken[0] == '#' {
            continue; }
        if len(pToken) != 16 {
            Abc_Print( 0, "Skipping token %s that does not look like a 16-digit hex number.\n" );
            continue;
        }
        // extract truth table
        Extra_ReadHexadecimal( Truth, pToken, 6 );
        uTruth = (((word)Truth[1]) << 32) | (word)Truth[0];
        // add truth table
        pEntry = Npn_ManAdd( p, uTruth );
        assert(pEntry.Count == 1, "pEntry.Count == 1");
        // read area
        pToken = strtok( nil, " \t\n" );
        pEntry.Count = strconv.Atoi(pToken);
    }
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManCompareEntries(pp1 **Npn_Obj_t, pp2 **Npn_Obj_t) int32 {
    if (*pp1).Count > (*pp2).Count {
        return -1; }
    if (*pp1).Count < (*pp2).Count { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Adds one entry to the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManWrite(p *Npn_Man_t, pFileName string) {
    var vEntries *Vec_Ptr_t;
    var pEntry *Npn_Obj_t;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var i int32;
    if pFile == nil {
        Abc_Print( -1, "Cannot open NPN function file \"%s\".\n", pFileName );
        return;
    }
    vEntries = Vec_PtrAlloc( p.nEntries );
    for  i = 0; i < p.nBins; i++  {
        for  pEntry = Npn_ManObj(p, p.pBins[i]); (pEntry != nil); pEntry = Npn_ManObj(p, pEntry.iNext)  {
            Vec_PtrPush( vEntries, pEntry );
        }
    }
    Vec_PtrSort( vEntries, (int (*)())Npn_ManCompareEntries );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vEntries)) && (((pEntry) = (Npn_Obj_t *)Vec_PtrEntry(vEntries, i)), 1); i++  {
        Extra_PrintHexadecimal( pFile, (unsigned *)&pEntry.uTruth, 6 );
        fmt.Fprintf(pFile, " %d %d\n", pEntry.Count, Npn_TruthSuppSize(pEntry.uTruth, 6));
    }
    _ = (pFile).Close();
    Vec_PtrFree( vEntries );
}

/**Function*************************************************************

  Synopsis    [Creates the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManStart(pFileName string) *Npn_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Npn_Man_t{};
    if pFileName == nil {
        p.nBufferSize = 1000000;
        p.nBufferSize = 100;
        // inlining of macro ABC_ALLOC
        p.pBuffer     = ((Npn_Obj_t *) malloc(sizeof(Npn_Obj_t) * (size_t)(p.nBufferSize)));
        p.nBins       = Abc_PrimeCudd( p.nBufferSize / 2 );
        // inlining of macro ABC_CALLOC
        p.pBins       = ((int *) calloc((size_t)(p.nBins), sizeof(int)));
        p.nEntries    = 1;
    
    } else {
        var pFile *os.File  = os.Open(pFileName);
        if pFile == nil {
            Abc_Print( -1, "Cannot open NPN function file \"%s\".\n", pFileName );
            return nil;
        }
        _ = (pFile).Close();
        p.nBufferSize = 4 * ( Extra_FileSize(pFileName) / 20 );
        // inlining of macro ABC_ALLOC
        p.pBuffer     = ((Npn_Obj_t *) malloc(sizeof(Npn_Obj_t) * (size_t)(p.nBufferSize)));
        p.nBins       = Abc_PrimeCudd( p.nBufferSize / 2 );
        // inlining of macro ABC_CALLOC
        p.pBins       = ((int *) calloc((size_t)(p.nBins), sizeof(int)));
        p.nEntries    = 1;
        Npn_ManRead( p, pFileName );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManStop(p *Npn_Man_t) {
    // inlining of macro ABC_FREE
    ((p.pBuffer) ? (free((char *) (p.pBuffer)), (p.pBuffer) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Cleans the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManClean() {
    if pNpnMan != nil {
        Npn_ManStop( pNpnMan );
        pNpnMan = nil;
    }
}

/**Function*************************************************************

  Synopsis    [Loads functions from a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManLoad(pFileName string) {
//    Npn_TruthPermute_rec( "012345", 0, 5 );
    if pNpnMan != nil {
        Abc_Print( 1, "Removing old table with %d entries.\n", pNpnMan.nEntries );
        Npn_ManStop( pNpnMan );
    }
    pNpnMan = Npn_ManStart( pFileName );
    Abc_Print( 1, "Created new table with %d entries from file \"%s\".\n", pNpnMan.nEntries, pFileName );
}

/**Function*************************************************************

  Synopsis    [Saves functions into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManSave(pFileName string) {
    if pNpnMan == nil {
        Abc_Print( 1, "There is no table with entries.\n" );
        return;
    }
    Npn_ManWrite( pNpnMan, pFileName );
    Abc_Print( 1, "Dumped table with %d entries from file \"%s\".\n", pNpnMan.nEntries, pFileName );
}

/**Function*************************************************************

  Synopsis    [Saves one function into storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Npn_ManSaveOne(puTruth *uint32, nVars int32) {
    var uTruth word  = (((word)puTruth[1]) << 32) | (word)puTruth[0];
    assert(nVars >= 0 && nVars <= 6, "nVars >= 0 && nVars <= 6");
    if pNpnMan == nil {
        Abc_Print( 1, "Creating new table with 0 entries.\n" );
        pNpnMan = Npn_ManStart( nil );
    }
    // skip truth tables that do not depend on some vars
    if !(Npn_TruthIsMinBase( uTruth ) != 0) {
        return; }
    // extend truth table to look like 6-input
    uTruth = Npn_TruthPadWord( uTruth, nVars );
    // semi(!)-NPN-canonize the truth table
    uTruth = Npn_TruthCanon( uTruth, 6, nil );
    // add to storage
    Npn_ManAdd( pNpnMan, uTruth );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

