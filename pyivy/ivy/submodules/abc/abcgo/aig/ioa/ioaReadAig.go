package ioa

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [ioaReadAiger.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read binary AIGER format developed by
  Armin Biere, Johannes Kepler University (http://fmv.jku.at/)]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - December 16, 2006.]

  Revision    [$Id: ioaReadAiger.c,v 1.00 2006/12/16 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioa.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Extracts one unsigned AIG edge from the input buffer.]

  Description [This procedure is a slightly modified version of Armin Biere's
  procedure "unsigned decode (FILE * file)". ]
  
  SideEffects [Updates the current reading position.]

  SeeAlso     []

***********************************************************************/
func Ioa_ReadAigerDecode(ppPos *string) uint32 {
    var x, i uint32;
    unsigned char ch;

//    while ((ch = getnoneofch (file)) & 0x80)
    for (ch = *(*ppPos)++) & 0x80 {
        x |= (ch & 0x7f) << (7 * i++);
    }

    return x | (ch << (7 * i));
}

/**Function*************************************************************

  Synopsis    [Decodes the encoded array of literals.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_WriteDecodeLiterals(ppPos *string, nEntries int32) *Vec_Int_t {
    var vLits *Vec_Int_t;
    var Lit, LitPrev, Diff, i int32;
    vLits = Vec_IntAlloc( nEntries );
    LitPrev = Ioa_ReadAigerDecode( ppPos );
    Vec_IntPush( vLits, LitPrev );
    for  i = 1; i < nEntries; i++  {
//        Diff = Lit - LitPrev;
//        Diff = (Lit < LitPrev)? -Diff : Diff;
//        Diff = ((2 * Diff) << 1) | (int)(Lit < LitPrev);
        Diff = Ioa_ReadAigerDecode( ppPos );
        if Diff & 1 {
            Diff = -(Diff >> 1);
        } else {
            Diff = Diff >> 1;
        }
        Lit  = Diff + LitPrev;
        Vec_IntPush( vLits, Lit );
        LitPrev = Lit;
    }
    return vLits;
}


/**Function*************************************************************

  Synopsis    [Reads the AIG in from the memory buffer.]

  Description [The buffer constains the AIG in AIGER format. The size gives
  the number of bytes in the buffer. The buffer is allocated by the user 
  and not deallocated by this procedure.]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_ReadAigerFromMemory(pContents string, nFileSize int32, fCheck int32) *Aig_Man_t {
    var vLits *Vec_Int_t  = nil;
    var vNodes, vDrivers *Vec_Ptr_t;//, * vTerms;
    var pObj, pNode0, pNode1 *Aig_Obj_t;
    var pNew *Aig_Man_t;
    var nTotal, nInputs, nOutputs, nLatches, nAnds, i int32;//, iTerm, nDigits;
    var nBad, nConstr, nJust, nFair int32;
    var pDrivers, pSymbols, pCur string;//, * pType;
    var uLit0, uLit1, uLit uint32;

    // check if the input file format is correct
    if strncmp(pContents, "aig", 3) != 0 || (pContents[3] != ' ' && pContents[3] != '2') {
        fmt.Fprintf(os.Stdout, "Wrong input file format.\n");
        return nil;
    }

    // read the parameters (M I L O A + B C J F)
    pCur = pContents;         for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of objects
    nTotal = strconv.Atoi(pCur);    for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of inputs
    nInputs = strconv.Atoi(pCur);   for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of latches
    nLatches = strconv.Atoi(pCur);  for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of outputs
    nOutputs = strconv.Atoi(pCur);  for  *pCur != ' '  { pCur++; }
 pCur++;
    // read the number of nodes
    nAnds = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
    if *pCur == ' ' {
        assert(nOutputs == 0, "nOutputs == 0");
        // read the number of properties
        pCur++;
        nBad = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nBad;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nConstr = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nConstr;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nJust = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nJust;
    }
    if *pCur == ' ' {
        // read the number of properties
        pCur++;
        nFair = strconv.Atoi(pCur);     for  *pCur != ' ' && *pCur != '\n'  { pCur++; } 
        nOutputs += nFair;
    }
    if *pCur != '\n' {
        fmt.Fprintf(os.Stdout, "The parameter line is in a wrong format.\n");
        return nil;
    }
    pCur++;

    // check the parameters
    if nTotal != nInputs + nLatches + nAnds {
        fmt.Fprintf(os.Stdout, "The number of objects does not match.\n");
        return nil;
    }
    if nJust || (nFair != 0) {
        fmt.Fprintf(os.Stdout, "Reading AIGER files with liveness properties are currently not supported.\n");
        return nil;
    }

    if nConstr {
        if nConstr == 1 {
            fmt.Fprintf(os.Stdout, "Warning: The last output is interpreted as a constraint.\n");
        } else {
            fmt.Fprintf(os.Stdout, "Warning: The last %d outputs are interpreted as constraints.\n", nConstr); }
    }

    // allocate the empty AIG
    pNew = Aig_ManStart( nAnds );
    pNew.nConstrs = nConstr;

    // prepare the array of nodes
    vNodes = Vec_PtrAlloc( 1 + nInputs + nLatches + nAnds );
    Vec_PtrPush( vNodes, Aig_ManConst0(pNew) );

    // create the PIs
    for  i = 0; i < nInputs + nLatches; i++  {
        pObj = Aig_ObjCreateCi(pNew);    
        Vec_PtrPush( vNodes, pObj );
    }
/*
    // create the POs
    for ( i = 0; i < nOutputs + nLatches; i++ )
    {
        pObj = Aig_ObjCreateCo(pNew);   
    }
*/
    // create the latches
    pNew.nRegs = nLatches;
/*
    nDigits = Ioa_Base10Log( nLatches );
    for ( i = 0; i < nLatches; i++ )
    {
        pObj = Aig_ObjCreateLatch(pNew);
        Aig_LatchSetInit0( pObj );
        pNode0 = Aig_ObjCreateBi(pNew);
        pNode1 = Aig_ObjCreateBo(pNew);
        Aig_ObjAddFanin( pObj, pNode0 );
        Aig_ObjAddFanin( pNode1, pObj );
        Vec_PtrPush( vNodes, pNode1 );
        // assign names to latch and its input
//        Aig_ObjAssignName( pObj, Aig_ObjNameDummy("_L", i, nDigits), NULL );
//        printf( "Creating latch %s with input %d and output %d.\n", Aig_ObjName(pObj), pNode0->Id, pNode1->Id );
    } 
*/ 

    // remember the beginning of latch/PO literals
    pDrivers = pCur;
    if pContents[3] == ' ' { // standard AIGER
    
        // scroll to the beginning of the binary data
        for  i = 0; i < nLatches + nOutputs;  {
            if *pCur++ == '\n' {
                i++; }
        }
    
    } else { // modified AIGER
    
        vLits = Ioa_WriteDecodeLiterals( &pCur, nLatches + nOutputs );
    }

    // create the AND gates
//    pProgress = Bar_ProgressStart( stdout, nAnds );
    for  i = 0; i < nAnds; i++  {
//        Bar_ProgressUpdate( pProgress, i, NULL );
        uLit = ((i + 1 + nInputs + nLatches) << 1);
        uLit1 = uLit  - Ioa_ReadAigerDecode( &pCur );
        uLit0 = uLit1 - Ioa_ReadAigerDecode( &pCur );
//        assert( uLit1 > uLit0 );
        pNode0 = Aig_NotCond( (Aig_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), uLit0 & 1 );
        pNode1 = Aig_NotCond( (Aig_Obj_t *)Vec_PtrEntry(vNodes, uLit1 >> 1), uLit1 & 1 );
        assert(Vec_PtrSize(vNodes) == i + 1 + nInputs + nLatches, "Vec_PtrSize(vNodes) == i + 1 + nInputs + nLatches");
        Vec_PtrPush( vNodes, Aig_And(pNew, pNode0, pNode1) );
    }
//    Bar_ProgressStop( pProgress );

    // remember the place where symbols begin
    pSymbols = pCur;

    // read the latch driver literals
    vDrivers = Vec_PtrAlloc( nLatches + nOutputs );
    if pContents[3] == ' ' { // standard AIGER
    
        pCur = pDrivers;
        for  i = 0; i < nLatches; i++  {
            uLit0 = strconv.Atoi(pCur);  for  *pCur++ != '\n'  {; }
            pNode0 = Aig_NotCond( (Aig_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );//^ (uLit0 < 2) );
            Vec_PtrPush( vDrivers, pNode0 );
        }
        // read the PO driver literals
        for  i = 0; i < nOutputs; i++  {
            uLit0 = strconv.Atoi(pCur);  for  *pCur++ != '\n'  {; }
            pNode0 = Aig_NotCond( (Aig_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );//^ (uLit0 < 2) );
            Vec_PtrPush( vDrivers, pNode0 );
        }

    
    } else {
        // read the latch driver literals
        for  i = 0; i < nLatches; i++  {
            uLit0 = Vec_IntEntry( vLits, i );
            pNode0 = Aig_NotCond( (Aig_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );//^ (uLit0 < 2) );
            Vec_PtrPush( vDrivers, pNode0 );
        }
        // read the PO driver literals
        for  i = 0; i < nOutputs; i++  {
            uLit0 = Vec_IntEntry( vLits, i+nLatches );
            pNode0 = Aig_NotCond( (Aig_Obj_t *)Vec_PtrEntry(vNodes, uLit0 >> 1), (uLit0 & 1) );//^ (uLit0 < 2) );
            Vec_PtrPush( vDrivers, pNode0 );
        }
        Vec_IntFree( vLits );
    }

    // create the POs
    for  i = 0; i < nOutputs; i++  {
        Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Vec_PtrEntry(vDrivers, nLatches + i) );
    }
    for  i = 0; i < nLatches; i++  {
        Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Vec_PtrEntry(vDrivers, i) );
    }
    Vec_PtrFree( vDrivers );

/*
    // read the names if present
    pCur = pSymbols;
    if ( *pCur != 'c' )
    {
        int Counter = 0;
        while ( pCur < pContents + nFileSize && *pCur != 'c' )
        {
            // get the terminal type
            pType = pCur;
            if ( *pCur == 'i' )
                vTerms = pNew->vPis;
            else if ( *pCur == 'l' )
                vTerms = pNew->vBoxes;
            else if ( *pCur == 'o' )
                vTerms = pNew->vPos;
            else
            {
                fprintf( stdout, "Wrong terminal type.\n" );
                return NULL;
            }
            // get the terminal number
            iTerm = atoi( ++pCur );  while ( *pCur++ != ' ' );
            // get the node
            if ( iTerm >= Vec_PtrSize(vTerms) )
            {
                fprintf( stdout, "The number of terminal is out of bound.\n" );
                return NULL;
            }
            pObj = Vec_PtrEntry( vTerms, iTerm );
            if ( *pType == 'l' )
                pObj = Aig_ObjFanout0(pObj);
            // assign the name
            pName = pCur;          while ( *pCur++ != '\n' );
            // assign this name 
            *(pCur-1) = 0;
            Aig_ObjAssignName( pObj, pName, NULL );
            if ( *pType == 'l' )
            {
                Aig_ObjAssignName( Aig_ObjFanin0(pObj), Aig_ObjName(pObj), "L" );
                Aig_ObjAssignName( Aig_ObjFanin0(Aig_ObjFanin0(pObj)), Aig_ObjName(pObj), "_in" );
            }
            // mark the node as named
            pObj->pCopy = (Aig_Obj_t *)Aig_ObjName(pObj);
        } 

        // assign the remaining names
        Aig_ManForEachCi( pNew, pObj, i )
        {
            if ( pObj->pCopy ) continue;
            Aig_ObjAssignName( pObj, Aig_ObjName(pObj), NULL );
            Counter++;
        }
        Aig_ManForEachLatchOutput( pNew, pObj, i )
        {
            if ( pObj->pCopy ) continue;
            Aig_ObjAssignName( pObj, Aig_ObjName(pObj), NULL );
            Aig_ObjAssignName( Aig_ObjFanin0(pObj), Aig_ObjName(pObj), "L" );
            Aig_ObjAssignName( Aig_ObjFanin0(Aig_ObjFanin0(pObj)), Aig_ObjName(pObj), "_in" );
            Counter++;
        }
        Aig_ManForEachCo( pNew, pObj, i )
        {
            if ( pObj->pCopy ) continue;
            Aig_ObjAssignName( pObj, Aig_ObjName(pObj), NULL );
            Counter++;
        }
        if ( Counter )
            printf( "Ioa_ReadAiger(): Added %d default names for nameless I/O/register objects.\n", Counter );
    }
    else
    {
//        printf( "Ioa_ReadAiger(): I/O/register names are not given. Generating short names.\n" );
        Aig_ManShortNames( pNew );
    }
*/
    pCur = pSymbols;
    if pCur + 1 < pContents + nFileSize && *pCur == 'c' {
        pCur++;
        if *pCur == 'n' {
            pCur++;
            // read model name
            // inlining of macro ABC_FREE
            ((pNew.pName) ? (free((char *) (pNew.pName)), (pNew.pName) = 0) : 0);
            pNew.pName = Abc_UtilStrsav( pCur );
        }
    }

    // skipping the comments
    Vec_PtrFree( vNodes );

    // remove the extra nodes
    Aig_ManCleanup( pNew );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(pNew) );

    // update polarity of the additional outputs
    if nBad || nConstr || nJust || (nFair != 0) {
        Aig_ManInvertConstraints( pNew ); }

    // check the result
    if (fCheck != 0) && !Aig_ManCheck( pNew ) {
        fmt.Printf("Ioa_ReadAiger: The network check has failed.\n");
        Aig_ManStop( pNew );
        return nil;
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reads the AIG in the binary AIGER format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ioa_ReadAiger(pFileName string, fCheck int32) *Aig_Man_t {
    var pFile *os.File;
    var pNew *Aig_Man_t;
    var pName, pContents string;
    var nFileSize, RetValue int32;

    // read the file into the buffer
    nFileSize = Ioa_FileSize( pFileName );
    pFile = os.Open(pFileName);
    // inlining of macro ABC_ALLOC
    pContents = ((char *) malloc(sizeof(char) * (size_t)(nFileSize)));
    RetValue = fread( pContents, nFileSize, 1, pFile );
    _ = (pFile).Close();

    pNew = Ioa_ReadAigerFromMemory( pContents, nFileSize, fCheck );
    // inlining of macro ABC_FREE
    ((pContents) ? (free((char *) (pContents)), (pContents) = 0) : 0);
    if pNew != nil {
        pName = Ioa_FileNameGeneric( pFileName );
        // inlining of macro ABC_FREE
        ((pNew.pName) ? (free((char *) (pNew.pName)), (pNew.pName) = 0) : 0);
        pNew.pName = Abc_UtilStrsav( pName );
        pNew.pSpec = Abc_UtilStrsav( pFileName );
        // inlining of macro ABC_FREE
        ((pName) ? (free((char *) (pName)), (pName) = 0) : 0);
    }
    return pNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

