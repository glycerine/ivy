package bsat

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [satStore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Records the trace of SAT solving in the CNF form.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satStore.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Fetches memory.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManMemoryFetch(p *Sto_Man_t, nBytes int32) string {
    var pMem string;
    if p.pChunkLast == nil || nBytes > p.nChunkSize - p.nChunkUsed {
        // inlining of macro ABC_ALLOC
        pMem = (char *)((char *) malloc(sizeof(char) * (size_t)(p.nChunkSize)));
        *(char **)pMem = p.pChunkLast;
        p.pChunkLast = pMem;
        p.nChunkUsed = sizeof(char *);
    }
    pMem = p.pChunkLast + p.nChunkUsed;
    p.nChunkUsed += nBytes;
    return pMem;
}

/**Function*************************************************************

  Synopsis    [Frees memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManMemoryStop(p *Sto_Man_t) {
    var pMem, pNext string;
    if p.pChunkLast == nil {
        return; }
    for  pMem = p.pChunkLast; (pNext = *(char **)pMem); pMem = pNext  {
        // inlining of macro ABC_FREE
        ((pMem) ? (free((char *) (pMem)), (pMem) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((pMem) ? (free((char *) (pMem)), (pMem) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Reports memory usage in bytes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManMemoryReport(p *Sto_Man_t) int32 {
    var Total int32;
    var pMem, pNext string;
    if p.pChunkLast == nil {
        return 0; }
    Total = p.nChunkUsed; 
    for  pMem = p.pChunkLast; (pNext = *(char **)pMem); pMem = pNext  {
        Total += p.nChunkSize;
    }
    return Total;
}


/**Function*************************************************************

  Synopsis    [Allocate proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManAlloc() *Sto_Man_t {
    var p *Sto_Man_t;
    // allocate the manager
    // inlining of macro ABC_ALLOC
    p = (Sto_Man_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Sto_Man_t))));
    memset( p, 0, sizeof(Sto_Man_t) );
    // memory management
    p.nChunkSize = (1<<16); // use 64K chunks
    return p;    
}

/**Function*************************************************************

  Synopsis    [Deallocate proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManFree(p *Sto_Man_t) {
    Sto_ManMemoryStop( p );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Adds one clause to the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManAddClause(p *Sto_Man_t, pBeg *lit, pEnd *lit) int32 {
    var pClause *Sto_Cls_t;
    var Lit lit;
    var i *lit;
    var j *lit;
    var nSize int32;

    // process the literals
    if pBeg < pEnd {
        // insertion sort
        for  i = pBeg + 1; i < pEnd; i++  {
            Lit = *i;
            for  j = i; j > pBeg && *(j-1) > Lit; j--  {
                *j = *(j-1);
            }
            *j = Lit;
        }
        // make sure there is no duplicated variables
        for  i = pBeg + 1; i < pEnd; i++  {
            if lit_var(*(i-1)) == lit_var(*i) {
                fmt.Printf("The clause contains two literals of the same variable: %d and %d.\n", *(i-1), *i);
                return 0;
            }
        }
        // check the largest var size
        // inlining of macro STO_MAX
        
        if (p.nVars) > (lit_var(*(pEnd-1)) + 1) {
            p.nVars = (p.nVars);
        } else {
            p.nVars = (lit_var(*(pEnd-1)) + 1);
        }
    }

    // get memory for the clause
    nSize = sizeof(Sto_Cls_t) + sizeof(lit) * (pEnd - pBeg);
    nSize = (nSize / sizeof(char*) + ((nSize % sizeof(char*)) > 0)) * sizeof(char*); // added by Saurabh on Sep 3, 2009
    pClause = (Sto_Cls_t *)Sto_ManMemoryFetch( p, nSize );
    memset( pClause, 0, sizeof(Sto_Cls_t) );

    // assign the clause
    pClause.Id = p.nClauses++;
    pClause.nLits = pEnd - pBeg;
    memcpy( pClause.pLits, pBeg, sizeof(lit) * (pEnd - pBeg) );
//    assert( pClause->pLits[0] >= 0 );

    // add the clause to the list
    if p.pHead == nil {
        p.pHead = pClause; }
    if p.pTail == nil {
        p.pTail = pClause;
    } else {
        p.pTail.pNext = pClause;
        p.pTail = pClause;
    }

    // add the empty clause
    if pClause.nLits == 0 {
        if p.pEmpty {
            fmt.Printf("More than one empty clause!\n");
            return 0;
        }
        p.pEmpty = pClause;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Mark all clauses added so far as root clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManMarkRoots(p *Sto_Man_t) {
    var pClause *Sto_Cls_t;
    p.nRoots = 0;
    // inlining of macro Sto_ManForEachClause
    for  pClause = p.pHead; (pClause != nil); pClause = pClause.pNext  {
        pClause.fRoot = 1;
        p.nRoots++;
    }
}

/**Function*************************************************************

  Synopsis    [Mark all clauses added so far as clause of A.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManMarkClausesA(p *Sto_Man_t) {
    var pClause *Sto_Cls_t;
    p.nClausesA = 0;
    // inlining of macro Sto_ManForEachClause
    for  pClause = p.pHead; (pClause != nil); pClause = pClause.pNext  {
        pClause.fA = 1;
        p.nClausesA++;
    }
}

/**Function*************************************************************

  Synopsis    [Returns the literal of the last clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManChangeLastClause(p *Sto_Man_t) int32 {
    var pClause, pPrev *Sto_Cls_t;
    pPrev = nil;
    // inlining of macro Sto_ManForEachClause
    for  pClause = p.pHead; (pClause != nil); pClause = pClause.pNext  {
        pPrev = pClause;
    }
    assert(pPrev != nil, "pPrev != nil");
    assert(pPrev.fA == 1, "pPrev.fA == 1");
    assert(pPrev.nLits == 1, "pPrev.nLits == 1");
    p.nClausesA--;
    pPrev.fA = 0;
    return pPrev.pLits[0] >> 1;
}


/**Function*************************************************************

  Synopsis    [Writes the stored clauses into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManDumpClauses(p *Sto_Man_t, pFileName string) {
    var pFile *os.File;
    var pClause *Sto_Cls_t;
    var i int32;
    // start the file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Error: Cannot open output file (%s).\n", pFileName);
        return;
    }
    // write the data
    fmt.Fprintf(pFile, "p %d %d %d %d\n", p.nVars, p.nClauses, p.nRoots, p.nClausesA);
    // inlining of macro Sto_ManForEachClause
    for  pClause = p.pHead; (pClause != nil); pClause = pClause.pNext  {
        for  i = 0; i < int32(pClause.nLits); i++  {
            fmt.Fprintf(pFile, " %d", lit_print(pClause.pLits[i]));
        }
        fmt.Fprintf(pFile, " 0\n");
    }
//    fprintf( pFile, " 0\n" );
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Reads one literal from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManLoadNumber(pFile *os.File, pNumber *int32) int32 {
    var Char, Number, Sign int32;
    // skip space-like chars
    for {
        Char = fgetc( pFile );
        if Char == -1 {
            return 0; }
    
if !( Char == ' ' || Char == '\t' || Char == '\r' || Char == '\n' ) { break }
} 
    // read the literal
    for {
        // get the next character
        Char = fgetc( pFile );
        if Char == ' ' || Char == '\t' || Char == '\r' || Char == '\n' {
            break; }
        // check that the char is a digit
        if (Char < '0' || Char > '9') && Char != '-' {
            fmt.Printf("Error: Wrong char (%c) in the input file.\n", Char);
            return 0;
        }
        // check if this is a minus
        if Char == '-' {
            Sign = 1;
        } else {
            Number = 10 * Number + Char; }
    }
    // return the number
    *pNumber = Sign? -Number : Number;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Reads CNF from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sto_ManLoadClauses(pFileName string) *Sto_Man_t {
    var pFile *os.File;
    var p *Sto_Man_t;
    var pClause *Sto_Cls_t;
    var pBuffer [1024]byte;
    var nLits, nLitsAlloc, Counter, Number int32;
    var pLits *lit;

    // start the file
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Error: Cannot open input file (%s).\n", pFileName);
        return nil;
    }

    // create the manager
    p = Sto_ManAlloc();

    // alloc the array of literals
    nLitsAlloc = 1024;
    // inlining of macro ABC_ALLOC
    pLits = (lit *)((char *) malloc(sizeof(char) * (size_t)(sizeof(lit) * nLitsAlloc)));

    // read file header
    p.nVars = p.nClauses = p.nRoots = p.nClausesA = 0;
    for  fgets( pBuffer, 1024, pFile )  {
        if pBuffer[0] == 'c' {
            continue; }
        if pBuffer[0] == 'p' {
            sscanf( pBuffer + 1, "%d %d %d %d", &p.nVars, &p.nClauses, &p.nRoots, &p.nClausesA );
            break;
        }
        printf( "Warning: Skipping line: \"%s\"\n", pBuffer );
    }

    // read the clauses
    nLits = 0;
    for  Sto_ManLoadNumber(pFile, &Number) != 0  {
        if Number == 0 {
            var RetValue int32;
            RetValue = Sto_ManAddClause( p, pLits, pLits + nLits );
            assert((RetValue != 0), "RetValue");
            nLits = 0;
            continue;
        }
        if nLits == nLitsAlloc {
            nLitsAlloc *= 2;
            // inlining of macro ABC_REALLOC
            
            if pLits != nil {
                pLits = ((lit *) realloc((char *)(pLits), sizeof(lit) * (size_t)(nLitsAlloc)));
            } else {
                pLits = ((lit *) malloc(sizeof(lit) * (size_t)(nLitsAlloc)));
            }
        }
        pLits[ nLits++ ] = lit_read(Number);
    }
    if nLits > 0 {
        fmt.Printf("Error: The last clause was not saved.\n"); }

    // count clauses
    Counter = 0;
    // inlining of macro Sto_ManForEachClause
    for  pClause = p.pHead; (pClause != nil); pClause = pClause.pNext  {
        Counter++;
    }

    // check the number of clauses
    if p.nClauses != Counter {
        fmt.Printf("Error: The actual number of clauses (%d) is different than declared (%d).\n", Counter, p.nClauses);
        Sto_ManFree( p );
        return nil;
    }

    // inlining of macro ABC_FREE
    ((pLits) ? (free((char *) (pLits)), (pLits) = 0) : 0);
    _ = (pFile).Close();
    return p;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

