package io

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [ioReadEdif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedure to read ISCAS benchmarks in EDIF.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadEdif.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Abc_Ntk_t * Io_ReadEdifNetwork( Extra_FileReader_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads the network from an EDIF file.]

  Description [Works only for the ISCAS benchmarks.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadEdif(pFileName string, fCheck int32) *Abc_Ntk_t {
    var p *Extra_FileReader_t;
    var pNtk *Abc_Ntk_t;

    fmt.Printf("Currently this parser does not work!\n");
    return nil;

    // start the file
    p = Extra_FileReaderAlloc( pFileName, "#", "\n\r", " \t()" );
    if p == nil {
        return nil; }

    // read the network
    pNtk = Io_ReadEdifNetwork( p );
    Extra_FileReaderFree( p );
    if pNtk == nil {
        return nil; }

    // make sure that everything is okay with the network structure
    if (fCheck != 0) && !Abc_NtkCheckRead( pNtk ) {
        fmt.Printf("Io_ReadEdif: The network check has failed.\n");
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
func Io_ReadEdifNetwork(p *Extra_FileReader_t) *Abc_Ntk_t {
    var pProgress *ProgressBar;
    var vTokens *Vec_Ptr_t;
    var pNtk *Abc_Ntk_t;
    var pNet, pObj, pFanout *Abc_Obj_t;
    var pGateName, pNetName string;
    var fTokensReady, iLine, i int32;

    // read the first line
    vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
    if int32(strings.Compare((char *)vTokens.pArray[0], "edif")) != 0 {
        fmt.Printf("%s: Wrong input file format.\n", Extra_FileReaderGetFileName(p));
        return nil;
    }
    
    // allocate the empty network
    pNtk = Abc_NtkStartRead( Extra_FileReaderGetFileName(p) );

    // go through the lines of the file
    fTokensReady = 0;
    pProgress = Extra_ProgressBarStart( os.Stdout, Extra_FileReaderGetFileSize(p) );
    for  iLine = 1; fTokensReady || (vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p)); iLine++  {
        Extra_ProgressBarUpdate( pProgress, Extra_FileReaderGetCurPosition(p), nil );

        // get the type of the line
        fTokensReady = 0;
        if int32(strings.Compare((char *)vTokens.pArray[0], "instance")) == 0 { 
            pNetName = (char *)vTokens.pArray[1];
            pNet = Abc_NtkFindOrCreateNet( pNtk, pNetName );
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            pGateName = (char *)vTokens.pArray[1];
            if strncmp( pGateName, "Flip", 4 ) == 0 {
                pObj = Abc_NtkCreateLatch( pNtk );
                Abc_LatchSetInit0( pObj );
            
            } else {
                pObj = Abc_NtkCreateNode( pNtk );
//                pObj->pData = Abc_NtkRegisterName( pNtk, pGateName );
                pObj.pData = Extra_UtilStrsav( pGateName ); // memory leak!!!
            }
            Abc_ObjAddFanin( pNet, pObj );
        
        } else if int32(strings.Compare((char *)vTokens.pArray[0], "net")) == 0 {
            pNetName = (char *)vTokens.pArray[1];
            if int32(strings.Compare(pNetName, "CK")) == 0 || int32(strings.Compare(pNetName, "RESET")) == 0 {
                continue; }
            if int32(strings.Compare(pNetName + len(pNetName) - 4, "_out")) == 0 {
                pNetName[len(pNetName) - 4] = 0; }
            pNet = Abc_NtkFindNet( pNtk, pNetName );
            assert(pNet, "pNet");
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            for  int32(strings.Compare((char *)vTokens.pArray[0], "portRef")) == 0  {
                if int32(strings.Compare(pNetName, (char *)vTokens.pArray[3])) != 0 {
                    pFanout = Abc_NtkFindNet( pNtk, (char *)vTokens.pArray[3] );
                    Abc_ObjAddFanin( Abc_ObjFanin0(pFanout), pNet );
                }
                vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            }
            fTokensReady = 1;
        
        } else if int32(strings.Compare((char *)vTokens.pArray[0], "library")) == 0 {
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            for  int32(strings.Compare((char *)vTokens.pArray[0], "port")) == 0  {
                pNetName = (char *)vTokens.pArray[1];
                if int32(strings.Compare(pNetName, "CK")) == 0 || int32(strings.Compare(pNetName, "RESET")) == 0 {
                    vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
                    continue;
                }
                if int32(strings.Compare(pNetName + len(pNetName) - 3, "_PO")) == 0 {
                    pNetName[len(pNetName) - 3] = 0; }
                if int32(strings.Compare((char *)vTokens.pArray[3], "INPUT")) == 0 {
                    Io_ReadCreatePi( pNtk, (char *)vTokens.pArray[1] );
                } else if int32(strings.Compare((char *)vTokens.pArray[3], "OUTPUT")) == 0 {
                    Io_ReadCreatePo( pNtk, (char *)vTokens.pArray[1] );
                } else {
                    fmt.Printf("%s (line %d): Wrong interface specification.\n", Extra_FileReaderGetFileName(p), iLine);
                    Abc_NtkDelete( pNtk );
                    return nil;
                }
                vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p);
            }
        
        } else if int32(strings.Compare((char *)vTokens.pArray[0], "design")) == 0 {
            pNtk.pName = nil; 
            pNtk.pName = (char *)Extra_UtilStrsav( (char *)vTokens.pArray[3] );
            break;
        }
    }
    Extra_ProgressBarStop( pProgress );

    // assign logic functions
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if strncmp( (char *)pObj.pData, "And", 3 ) == 0 {
            Abc_ObjSetData( pObj, Abc_SopCreateAnd((Mem_Flex_t *)pNtk.pManFunc, Abc_ObjFaninNum(pObj), nil) );
        } else if strncmp( (char *)pObj.pData, "Or", 2 ) == 0 {
            Abc_ObjSetData( pObj, Abc_SopCreateOr((Mem_Flex_t *)pNtk.pManFunc, Abc_ObjFaninNum(pObj), nil) );
        } else if strncmp( (char *)pObj.pData, "Nand", 4 ) == 0 {
            Abc_ObjSetData( pObj, Abc_SopCreateNand((Mem_Flex_t *)pNtk.pManFunc, Abc_ObjFaninNum(pObj)) );
        } else if strncmp( (char *)pObj.pData, "Nor", 3 ) == 0 {
            Abc_ObjSetData( pObj, Abc_SopCreateNor((Mem_Flex_t *)pNtk.pManFunc, Abc_ObjFaninNum(pObj)) );
        } else if strncmp( (char *)pObj.pData, "Exor", 4 ) == 0 {
            Abc_ObjSetData( pObj, Abc_SopCreateXor((Mem_Flex_t *)pNtk.pManFunc, Abc_ObjFaninNum(pObj)) );
        } else if strncmp( (char *)pObj.pData, "Exnor", 5 ) == 0 {
            Abc_ObjSetData( pObj, Abc_SopCreateNxor((Mem_Flex_t *)pNtk.pManFunc, Abc_ObjFaninNum(pObj)) );
        } else if strncmp( (char *)pObj.pData, "Inv", 3 ) == 0 {
            Abc_ObjSetData( pObj, Abc_SopCreateInv((Mem_Flex_t *)pNtk.pManFunc) );
        } else if strncmp( (char *)pObj.pData, "Buf", 3 ) == 0 {
            Abc_ObjSetData( pObj, Abc_SopCreateBuf((Mem_Flex_t *)pNtk.pManFunc) );
        } else {
            fmt.Printf("%s: Unknown gate type \"%s\".\n", Extra_FileReaderGetFileName(p), (char*)pObj.pData);
            Abc_NtkDelete( pNtk );
            return nil;
        }
    }
    }
    // check if constants have been added
//    if ( pNet = Abc_NtkFindNet( pNtk, "VDD" ) )
//        Io_ReadCreateConst( pNtk, "VDD", 1 );
//    if ( pNet = Abc_NtkFindNet( pNtk, "GND" ) )
//        Io_ReadCreateConst( pNtk, "GND", 0 );

    Abc_NtkFinalizeRead( pNtk );
    return pNtk;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_IMPL_END

