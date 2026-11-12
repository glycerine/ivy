package io

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [ioReadEqn.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read equation format files.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadEqn.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Abc_Ntk_t * Io_ReadEqnNetwork( Extra_FileReader_t * p );

// static void        Io_ReadEqnStrCompact( char * pStr );

// static int         Io_ReadEqnStrFind( Vec_Ptr_t * vTokens, char * pName );

// static void        Io_ReadEqnStrCutAt( char * pStr, char * pStop, int fUniqueOnly, Vec_Ptr_t * vTokens );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads the network from a BENCH file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadEqn(pFileName string, fCheck int32) *Abc_Ntk_t {
    var p *Extra_FileReader_t;
    var pNtk *Abc_Ntk_t;

    // start the file
    p = Extra_FileReaderAlloc( pFileName, "#", ";", "=" );
    if p == nil {
        return nil; }

    // read the network
    pNtk = Io_ReadEqnNetwork( p );
    Extra_FileReaderFree( p );
    if pNtk == nil {
        return nil; }

    // make sure that everything is okay with the network structure
    if (fCheck != 0) && !Abc_NtkCheckRead( pNtk ) {
        fmt.Printf("Io_ReadEqn: The network check has failed.\n");
        Abc_NtkDelete( pNtk );
        return nil;
    }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadEqnNetwork(p *Extra_FileReader_t) *Abc_Ntk_t {
    var pProgress *ProgressBar;
    var vTokens *Vec_Ptr_t;
    var vVars *Vec_Ptr_t;
    var pNtk *Abc_Ntk_t;
    var pNode *Abc_Obj_t;
    var pNodeName, pFormula, pFormulaCopy, pVarName string;
    var iLine, i int32;
    
    // allocate the empty network
    pNtk = Abc_NtkAlloc( ABC_NTK_NETLIST, ABC_FUNC_AIG, 1 );
    // set the specs
    pNtk.pName = Extra_FileNameGeneric(Extra_FileReaderGetFileName(p));
    pNtk.pSpec = Extra_UtilStrsav(Extra_FileReaderGetFileName(p));

    // go through the lines of the file
    vVars  = Vec_PtrAlloc( 100 );
    pProgress = Extra_ProgressBarStart( os.Stdout, Extra_FileReaderGetFileSize(p) );
    for  iLine = 0; (vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p)); iLine++  {
        Extra_ProgressBarUpdate( pProgress, Extra_FileReaderGetCurPosition(p), nil );

        // check if the first token contains anything
        Io_ReadEqnStrCompact( (char *)vTokens.pArray[0] );
        if len((char *)vTokens.pArray[0]) == 0 {
            break; }

        // if the number of tokens is different from two, error
        if vTokens.nSize != 2 {
            fmt.Printf("%s: Wrong input file format.\n", Extra_FileReaderGetFileName(p));
            Abc_NtkDelete( pNtk );
            return nil;
        }

        // get the type of the line
        if strncmp( (char *)vTokens.pArray[0], "INORDER", 7 ) == 0 {
            Io_ReadEqnStrCutAt( (char *)vTokens.pArray[1], " \n\r\t", 0, vVars );
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(vVars)) && (((pVarName) = (char *)Vec_PtrEntry(vVars, i)), 1); i++  {
                Io_ReadCreatePi( pNtk, pVarName );
            }
        
        } else if strncmp( (char *)vTokens.pArray[0], "OUTORDER", 8 ) == 0 {
            Io_ReadEqnStrCutAt( (char *)vTokens.pArray[1], " \n\r\t", 0, vVars );
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(vVars)) && (((pVarName) = (char *)Vec_PtrEntry(vVars, i)), 1); i++  {
                Io_ReadCreatePo( pNtk, pVarName );
            }
        
        } else {
            extern Hop_Obj_t * Parse_FormulaParserEqn( FILE * pOutput, char * pFormInit, Vec_Ptr_t * vVarNames, Hop_Man_t * pMan );

            // get hold of the node name and its formula
            pNodeName = (char *)vTokens.pArray[0];
            pFormula  = (char *)vTokens.pArray[1];
            // compact the formula 
            Io_ReadEqnStrCompact( pFormula );

            // consider the case of the constant node
            if pFormula[1] == 0 && (pFormula[0] == '0' || pFormula[0] == '1') {
                pFormulaCopy = nil;
                Vec_PtrClear( vVars );
            
            } else {
                // make a copy of formula for names
                pFormulaCopy = Extra_UtilStrsav( pFormula );
                // find the names of the fanins of this node
                Io_ReadEqnStrCutAt( pFormulaCopy, "!*+()", 1, vVars );
            }
            // create the node
            pNode = Io_ReadCreateNode( pNtk, pNodeName, (char **)Vec_PtrArray(vVars), Vec_PtrSize(vVars) );
            // derive the function
            pNode.pData = Parse_FormulaParserEqn( os.Stdout, pFormula, vVars, (Hop_Man_t *)pNtk.pManFunc );
            // remove the cubes
            // inlining of macro ABC_FREE
            ((pFormulaCopy) ? (free((char *) (pFormulaCopy)), (pFormulaCopy) = 0) : 0);
        }
    }
    Extra_ProgressBarStop( pProgress );
    Vec_PtrFree( vVars );
    Abc_NtkFinalizeRead( pNtk );
    return pNtk;
}



/**Function*************************************************************

  Synopsis    [Compacts the string by throwing away space-like chars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadEqnStrCompact(pStr string) {
    var pCur, pNew string;
    for  pNew = pCur = pStr; *pCur; pCur++  {
        if !(*pCur == ' ' || *pCur == '\n' || *pCur == '\r' || *pCur == '\t') {
            *pNew++ = *pCur; }
    }
    *pNew = 0;
}

/**Function*************************************************************

  Synopsis    [Determines unique variables in the string.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadEqnStrFind(vTokens *Vec_Ptr_t, pName string) int32 {
    var pToken string;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTokens)) && (((pToken) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
        if int32(strings.Compare(pToken, pName)) == 0 {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Cuts the string into pieces using stop chars.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadEqnStrCutAt(pStr string, pStop string, fUniqueOnly int32, vTokens *Vec_Ptr_t) {
    var pToken string;
    Vec_PtrClear( vTokens );
    for  pToken = strtok( pStr, pStop ); pToken; pToken = strtok( nil, pStop )  {
        if !(fUniqueOnly != 0) || Io_ReadEqnStrFind( vTokens, pToken ) == -1 {
            Vec_PtrPush( vTokens, pToken ); }
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_IMPL_END

