package io

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [ioReadBaf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read AIG in the binary format.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadBaf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads the AIG in the binary format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBaf(pFileName string, fCheck int32) *Abc_Ntk_t {
    var pProgress *ProgressBar;
    var pFile *os.File;
    var vNodes *Vec_Ptr_t;
    var pObj, pNode0, pNode1 *Abc_Obj_t;
    var pNtkNew *Abc_Ntk_t;
    var nInputs, nOutputs, nLatches, nAnds, nFileSize, Num, i int32;
    var pContents, pName, pCur string;
    var pBufferNode *uint32;
    var RetValue int32;

    // read the file into the buffer
    nFileSize = Extra_FileSize( pFileName );
    pFile = os.Open(pFileName);
    // inlining of macro ABC_ALLOC
    pContents = ((char *) malloc(sizeof(char) * (size_t)(nFileSize)));
    RetValue = fread( pContents, nFileSize, 1, pFile );
    _ = (pFile).Close();

    // skip the comments (comment lines begin with '#' and end with '\n')
    for  pCur = pContents; *pCur == '#';  {
        for  *pCur++ != '\n'  {; }
    }

    // read the name
    pName = pCur;             for  *pCur++  {; }
    // read the number of inputs
    nInputs = strconv.Atoi(pCur);   for  *pCur++  {; }
    // read the number of outputs
    nOutputs = strconv.Atoi(pCur);  for  *pCur++  {; }
    // read the number of latches
    nLatches = strconv.Atoi(pCur);  for  *pCur++  {; }
    // read the number of nodes
    nAnds = strconv.Atoi(pCur);     for  *pCur++  {; }

    // allocate the empty AIG
    pNtkNew = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtkNew.pName = Extra_UtilStrsav( pName );
    pNtkNew.pSpec = Extra_UtilStrsav( pFileName );

    // prepare the array of nodes
    vNodes = Vec_PtrAlloc( 1 + nInputs + nLatches + nAnds );
    Vec_PtrPush( vNodes, Abc_AigConst1(pNtkNew) );

    // create the PIs
    for  i = 0; i < nInputs; i++  {
        pObj = Abc_NtkCreatePi(pNtkNew);    
        Abc_ObjAssignName( pObj, pCur, nil );  for  *pCur++  {; }
        Vec_PtrPush( vNodes, pObj );
    }
    // create the POs
    for  i = 0; i < nOutputs; i++  {
        pObj = Abc_NtkCreatePo(pNtkNew);   
        Abc_ObjAssignName( pObj, pCur, nil );  for  *pCur++  {; } 
    }
    // create the latches
    for  i = 0; i < nLatches; i++  {
        pObj = Abc_NtkCreateLatch(pNtkNew);
        Abc_ObjAssignName( pObj, pCur, nil );  for  *pCur++  {; } 

        pNode0 = Abc_NtkCreateBi(pNtkNew);
        Abc_ObjAssignName( pNode0, pCur, nil );  for  *pCur++  {; } 

        pNode1 = Abc_NtkCreateBo(pNtkNew);
        Abc_ObjAssignName( pNode1, pCur, nil );  for  *pCur++  {; } 
        Vec_PtrPush( vNodes, pNode1 );

        Abc_ObjAddFanin( pObj, pNode0 );
        Abc_ObjAddFanin( pNode1, pObj );
    }

    // get the pointer to the beginning of the node array
    pBufferNode = (unsigned *)(pContents + (nFileSize - (2 * nAnds + nOutputs + nLatches) * sizeof(int)) );
    // make sure we are at the place where the nodes begin
    if pBufferNode != (unsigned *)pCur {
        // inlining of macro ABC_FREE
        ((pContents) ? (free((char *) (pContents)), (pContents) = 0) : 0);
        Vec_PtrFree( vNodes );
        Abc_NtkDelete( pNtkNew );
        fmt.Printf("Warning: Internal reader error.\n");
        return nil;
    }

    // create the AND gates
    pProgress = Extra_ProgressBarStart( os.Stdout, nAnds );
    for  i = 0; i < nAnds; i++  {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        pNode0 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, pBufferNode[2*i+0] >> 1), pBufferNode[2*i+0] & 1 );
        pNode1 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, pBufferNode[2*i+1] >> 1), pBufferNode[2*i+1] & 1 );
        Vec_PtrPush( vNodes, Abc_AigAnd((Abc_Aig_t *)pNtkNew.pManFunc, pNode0, pNode1) );
    }
    Extra_ProgressBarStop( pProgress );

    // read the POs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtkNew)) && (((pObj) = Abc_NtkCo(pNtkNew, i)), 1); i++  {
        Num = pBufferNode[2*nAnds+i];
        if Abc_ObjFanoutNum(pObj) > 0 && Abc_ObjIsLatch(Abc_ObjFanout0(pObj)) {
            Abc_ObjSetData( Abc_ObjFanout0(pObj), (void *)(ABC_PTRINT_T)(Num & 3) );
            Num >>= 2;
        }
        pNode0 = Abc_ObjNotCond( (Abc_Obj_t *)Vec_PtrEntry(vNodes, Num >> 1), Num & 1 );
        Abc_ObjAddFanin( pObj, pNode0 );
    }
    // inlining of macro ABC_FREE
    ((pContents) ? (free((char *) (pContents)), (pContents) = 0) : 0);
    Vec_PtrFree( vNodes );

    // remove the extra nodes
//    Abc_AigCleanup( (Abc_Aig_t *)pNtkNew->pManFunc );

    // check the result
    if (fCheck != 0) && !Abc_NtkCheckRead( pNtkNew ) {
        fmt.Printf("Io_ReadBaf: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;

}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

