package io

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [ioReadBlif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to read BLIF files.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadBlif.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

             // all reading info
type Io_ReadBlif_t struct {
    // general info about file
    pFileName                string;    // the name of the file
    pReader  *Extra_FileReader_t;      // the input file reader
    // current processing info
    pNtkMaster           *Abc_Ntk_t;   // the primary network
    pNtkCur           *Abc_Ntk_t;      // the primary network
    LineCur                  int32;      // the line currently parsed
    // temporary storage for tokens
    vTokens           *Vec_Ptr_t;      // the current tokens
    vNewTokens           *Vec_Ptr_t;   // the temporary storage for the tokens
    vCubes           *Vec_Str_t;       // the temporary storage for the tokens
    // timing information
    vInArrs           *Vec_Int_t;      // input arrival
    vOutReqs           *Vec_Int_t;     // output required
    vInDrives           *Vec_Int_t;    // input drive
    vOutLoads           *Vec_Int_t;    // output load
    DefInArrRise                float32; // input arrival default
    DefInArrFall                float32; // input arrival default
    DefOutReqRise                float32;// output required default
    DefOutReqFall                float32;// output required default
    DefInDriRise                float32; // input drive default
    DefInDriFall                float32; // input drive default
    DefOutLoadRise                float32;// output load default
    DefOutLoadFall                float32;// output load default
    fHaveDefInArr                  int32;  // provided in the file
    fHaveDefOutReq                  int32; // provided in the file
    fHaveDefInDri                  int32;  // provided in the file
    fHaveDefOutLoad                  int32;// provided in the file
    // the error message
    Output                *os.File;       // the output stream
    sError [1000]byte; // the error string generated during parsing
    fError                  int32;       // set to 1 when error occurs
}

// static Io_ReadBlif_t * Io_ReadBlifFile( char * pFileName );

// static void Io_ReadBlifFree( Io_ReadBlif_t * p );

// static void Io_ReadBlifPrintErrorMessage( Io_ReadBlif_t * p );

// static Vec_Ptr_t * Io_ReadBlifGetTokens( Io_ReadBlif_t * p );

// static char * Io_ReadBlifCleanName( char * pName );


// static Abc_Ntk_t * Io_ReadBlifNetwork( Io_ReadBlif_t * p );

// static Abc_Ntk_t * Io_ReadBlifNetworkOne( Io_ReadBlif_t * p );

// static int Io_ReadBlifNetworkInputs( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkOutputs( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkLatch( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkNames( Io_ReadBlif_t * p, Vec_Ptr_t ** pvTokens );

// static int Io_ReadBlifNetworkGate( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkSubcircuit( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkInputArrival( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkOutputRequired( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkDefaultInputArrival( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkDefaultOutputRequired( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkInputDrive( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkOutputLoad( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkDefaultInputDrive( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkDefaultOutputLoad( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkAndGateDelay( Io_ReadBlif_t * p, Vec_Ptr_t * vTokens );

// static int Io_ReadBlifNetworkConnectBoxes( Io_ReadBlif_t * p, Abc_Ntk_t * pNtkMaster );

// static int Io_ReadBlifCreateTiming( Io_ReadBlif_t * p, Abc_Ntk_t * pNtkMaster );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads the (hierarchical) network from the BLIF file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlif(pFileName string, fCheck int32) *Abc_Ntk_t {
    var p *Io_ReadBlif_t;
    var pNtk *Abc_Ntk_t;

    // start the file
    p = Io_ReadBlifFile( pFileName );
    if p == nil {
        return nil; }

    // read the hierarchical network
    pNtk = Io_ReadBlifNetwork( p );
    if pNtk == nil {
        Io_ReadBlifFree( p );
        return nil;
    }
    pNtk.pSpec = Extra_UtilStrsav( pFileName );
    Io_ReadBlifCreateTiming( p, pNtk );
    Io_ReadBlifFree( p );

    // make sure that everything is okay with the network structure
    if (fCheck != 0) && !Abc_NtkCheckRead( pNtk ) {
        fmt.Printf("Io_ReadBlif: The network check has failed.\n");
        Abc_NtkDelete( pNtk );
        return nil;
    }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Iteratively reads several networks in the hierarchical design.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetwork(p *Io_ReadBlif_t) *Abc_Ntk_t {
    var pNtk, pNtkMaster *Abc_Ntk_t;

    // read the name of the master network
    p.vTokens = Io_ReadBlifGetTokens(p);
    if p.vTokens == nil || int32(strings.Compare((char *)p.vTokens.pArray[0], ".model")) {
        p.LineCur = 0;
        sprintf( p.sError, "Wrong input file format." );
        Io_ReadBlifPrintErrorMessage( p );
        return nil;
    }

    // read networks (with EXDC) 
    pNtkMaster = nil;
    for  p.vTokens  {
        // read the network and its EXDC if present
        pNtk = Io_ReadBlifNetworkOne( p );
        if pNtk == nil {
            break; }
        if p.vTokens && int32(strings.Compare((char *)p.vTokens.pArray[0], ".exdc")) == 0 {
            pNtk.pExdc = Io_ReadBlifNetworkOne( p );
            if pNtk.pExdc == nil {
                break; }
            Abc_NtkFinalizeRead( pNtk.pExdc );
        }
        // add this network as part of the hierarchy
        if pNtkMaster == nil { // no master network so far
        
            p.pNtkMaster = pNtkMaster = pNtk;
            continue;
        }
/*
        // make sure hierarchy does not have the network with this name
        if ( pNtkMaster->tName2Model && stmm_is_member( pNtkMaster->tName2Model, pNtk->pName ) )
        {
            p->LineCur = 0;
            sprintf( p->sError, "Model %s is multiply defined in the file.", pNtk->pName );
            Io_ReadBlifPrintErrorMessage( p );
            Abc_NtkDelete( pNtk );
            Abc_NtkDelete( pNtkMaster );
            pNtkMaster = NULL;
            return NULL;
        }
        // add the network to the hierarchy
        if ( pNtkMaster->tName2Model == NULL )
            pNtkMaster->tName2Model = stmm_init_table((int (*)(void))strcmp, (int (*)(void))stmm_strhash);
        stmm_insert( pNtkMaster->tName2Model, pNtk->pName, (char *)pNtk );
*/
    }
/*
    // if there is a hierarchy, connect the boxes
    if ( pNtkMaster && pNtkMaster->tName2Model )
    {
        if ( Io_ReadBlifNetworkConnectBoxes( p, pNtkMaster ) )
        {
            Abc_NtkDelete( pNtkMaster );
            return NULL;
        }
    }
    else 
*/
    if !p.fError {
        Abc_NtkFinalizeRead( pNtkMaster ); }
    // return the master network
    return pNtkMaster;
}

/**Function*************************************************************

  Synopsis    [Reads one (main or exdc) network from the BLIF file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkOne(p *Io_ReadBlif_t) *Abc_Ntk_t {
    var pProgress *ProgressBar  = nil;
    var pNtk *Abc_Ntk_t;
    var pDirective string;
    var iLine, fTokensReady, fStatus int32;

    // make sure the tokens are present
    assert(p.vTokens != nil, "p.vTokens != nil");

    // create the new network
    p.pNtkCur = pNtk = Abc_NtkAlloc( ABC_NTK_NETLIST, ABC_FUNC_SOP, 1 );
    // read the model name
    if int32(strings.Compare((char *)p.vTokens.pArray[0], ".model")) == 0 {
        var pToken, pPivot string;
        if Vec_PtrSize(p.vTokens) != 2 {
            p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
            sprintf( p.sError, "The .model line does not have exactly two entries." );
            Io_ReadBlifPrintErrorMessage( p );
            return nil;
        }
        for  pPivot = pToken = (char *)Vec_PtrEntry(p.vTokens, 1); *pToken; pToken++  {
            if *pToken == '/' || *pToken == '\\' {
                pPivot = pToken+1; }
        }
        pNtk.pName = Extra_UtilStrsav( pPivot );
    
    } else if int32(strings.Compare((char *)p.vTokens.pArray[0], ".exdc")) != 0 {
        fmt.Printf("%s: File parsing skipped after line %d (\"%s\").\n", p.pFileName, Extra_FileReaderGetLineNumber(p.pReader, 0), (char*)p.vTokens.pArray[0]);
        Abc_NtkDelete(pNtk);
        p.pNtkCur = nil;
        return nil;
    }

    // read the inputs/outputs
    if p.pNtkMaster == nil {
        pProgress = Extra_ProgressBarStart( os.Stdout, Extra_FileReaderGetFileSize(p.pReader) ); }
    fTokensReady = fStatus = 0;
    for  iLine = 0; fTokensReady || (p.vTokens = Io_ReadBlifGetTokens(p)); iLine++  {
        if p.pNtkMaster == nil && iLine % 1000 == 0 {
            Extra_ProgressBarUpdate( pProgress, Extra_FileReaderGetCurPosition(p.pReader), nil ); }

        // consider different line types
        fTokensReady = 0;
        pDirective = (char *)p.vTokens.pArray[0];
        if !(int32(strings.Compare(pDirective, ".names")) != 0) { fStatus = Io_ReadBlifNetworkNames( p, &p.vTokens ); fTokensReady = 1; 
        } else if !(int32(strings.Compare(pDirective, ".gate")) != 0) {
            fStatus = Io_ReadBlifNetworkGate( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".latch")) != 0) {
            fStatus = Io_ReadBlifNetworkLatch( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".inputs")) != 0) {
            fStatus = Io_ReadBlifNetworkInputs( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".outputs")) != 0) {
            fStatus = Io_ReadBlifNetworkOutputs( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".input_arrival")) != 0) {
            fStatus = Io_ReadBlifNetworkInputArrival( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".output_required")) != 0) {
            fStatus = Io_ReadBlifNetworkOutputRequired( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".default_input_arrival")) != 0) {
            fStatus = Io_ReadBlifNetworkDefaultInputArrival( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".default_output_required")) != 0) {
            fStatus = Io_ReadBlifNetworkDefaultOutputRequired( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".input_drive")) != 0) {
            fStatus = Io_ReadBlifNetworkInputDrive( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".output_load")) != 0) {
            fStatus = Io_ReadBlifNetworkOutputLoad( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".default_input_drive")) != 0) {
            fStatus = Io_ReadBlifNetworkDefaultInputDrive( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".default_output_load")) != 0) {
            fStatus = Io_ReadBlifNetworkDefaultOutputLoad( p, p.vTokens );
        } else if !(int32(strings.Compare(pDirective, ".and_gate_delay")) != 0) {
            fStatus = Io_ReadBlifNetworkAndGateDelay( p, p.vTokens );
//        else if ( !strcmp( pDirective, ".subckt" ) )
//            fStatus = Io_ReadBlifNetworkSubcircuit( p, p->vTokens );
        } else if !(int32(strings.Compare(pDirective, ".exdc")) != 0) {
            break;
        } else if !(int32(strings.Compare(pDirective, ".end")) != 0) {
            p.vTokens = Io_ReadBlifGetTokens(p);
            break;
        
        } else if !(int32(strings.Compare(pDirective, ".blackbox")) != 0) {
            pNtk.ntkType = ABC_NTK_NETLIST;
            pNtk.ntkFunc = ABC_FUNC_BLACKBOX;
            Mem_FlexStop( (Mem_Flex_t *)pNtk.pManFunc, 0 );
            pNtk.pManFunc = nil;
        
        } else {
            fmt.Printf("%s (line %d): Skipping directive \"%s\".\n", p.pFileName, Extra_FileReaderGetLineNumber(p.pReader, 0), pDirective); }
        if p.vTokens == nil { // some files do not have ".end" in the end
            break; }
        if fStatus == 1 {
            Extra_ProgressBarStop( pProgress );
            Abc_NtkDelete( pNtk );
            return nil;
        }
    }
    if p.pNtkMaster == nil {
        Extra_ProgressBarStop( pProgress ); }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkInputs(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var i int32;
    for  i = 1; i < vTokens.nSize; i++  {
        Io_ReadCreatePi( p.pNtkCur, (char *)vTokens.pArray[i] );
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkOutputs(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var i int32;
    for  i = 1; i < vTokens.nSize; i++  {
        Io_ReadCreatePo( p.pNtkCur, (char *)vTokens.pArray[i] );
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkLatch(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 { 
    var pNtk *Abc_Ntk_t  = p.pNtkCur;
    var pLatch *Abc_Obj_t;
    var ResetValue int32;
    if vTokens.nSize < 3 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "The .latch line does not have enough tokens." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // create the latch
    pLatch = Io_ReadCreateLatch( pNtk, (char *)vTokens.pArray[1], (char *)vTokens.pArray[2] );
    // get the latch reset value
    if vTokens.nSize == 3 {
        Abc_LatchSetInitDc( pLatch );
    } else {
        ResetValue = strconv.Atoi((char *)vTokens.pArray[vTokens.nSize-1]);
        if ResetValue != 0 && ResetValue != 1 && ResetValue != 2 {
            p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
            sprintf( p.sError, "The .latch line has an unknown reset value (%s).", (char*)vTokens.pArray[3] );
            Io_ReadBlifPrintErrorMessage( p );
            return 1;
        }
        if ResetValue == 0 {
            Abc_LatchSetInit0( pLatch );
        } else if ResetValue == 1 {
            Abc_LatchSetInit1( pLatch );
        } else if ResetValue == 2 {
            Abc_LatchSetInitDc( pLatch ); }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkNames(p *Io_ReadBlif_t, pvTokens **Vec_Ptr_t) int32 {
    var vTokens *Vec_Ptr_t  = *pvTokens;
    var pNtk *Abc_Ntk_t  = p.pNtkCur;
    var pNode *Abc_Obj_t;
    var pToken string;
    var Char byte;
    var ppNames *string;
    var nFanins, nNames int32;

    // create a new node and add it to the network
    if vTokens.nSize < 2 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "The .names line has less than two tokens." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }

    // create the node
    ppNames = (char **)vTokens.pArray + 1;
    nNames  = vTokens.nSize - 2;
    pNode   = Io_ReadCreateNode( pNtk, ppNames[nNames], ppNames, nNames );

    // derive the functionality of the node
    p.vCubes.nSize = 0;
    nFanins = vTokens.nSize - 2;
    if nFanins == 0 {
        for  vTokens = Io_ReadBlifGetTokens(p)  {
            pToken = (char *)vTokens.pArray[0];
            if pToken[0] == '.' {
                break; }
            // read the cube
            if vTokens.nSize != 1 {
                p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
                sprintf( p.sError, "The number of tokens in the constant cube is wrong." );
                Io_ReadBlifPrintErrorMessage( p );
                return 1;
            }
            // create the cube
            Char = ((char *)vTokens.pArray[0])[0];
            Vec_StrPush( p.vCubes, ' ' );
            Vec_StrPush( p.vCubes, Char );
            Vec_StrPush( p.vCubes, '\n' );
        }
    
    } else {
        for  vTokens = Io_ReadBlifGetTokens(p)  {
            pToken = (char *)vTokens.pArray[0];
            if pToken[0] == '.' {
                break; }
            // read the cube
            if vTokens.nSize != 2 {
                p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
                sprintf( p.sError, "The number of tokens in the cube is wrong." );
                Io_ReadBlifPrintErrorMessage( p );
                return 1;
            }
            // create the cube
            Vec_StrPrintStr( p.vCubes, (char *)vTokens.pArray[0] );
            // check the char 
            Char = ((char *)vTokens.pArray[1])[0];
            if Char != '0' && Char != '1' && Char != 'x' && Char != 'n' {
                p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
                sprintf( p.sError, "The output character in the constant cube is wrong." );
                Io_ReadBlifPrintErrorMessage( p );
                return 1;
            }
            Vec_StrPush( p.vCubes, ' ' );
            Vec_StrPush( p.vCubes, Char );
            Vec_StrPush( p.vCubes, '\n' );
        }
    }
    // if there is nothing there
    if p.vCubes.nSize == 0 {
        // create an empty cube
        Vec_StrPush( p.vCubes, ' ' );
        Vec_StrPush( p.vCubes, '0' );
        Vec_StrPush( p.vCubes, '\n' );
    }
    Vec_StrPush( p.vCubes, 0 );

    // set the pointer to the functionality of the node
    Abc_ObjSetData( pNode, Abc_SopRegister((Mem_Flex_t *)pNtk.pManFunc, p.vCubes.pArray) );

    // check the size
    if Abc_ObjFaninNum(pNode) != Abc_SopGetVarNum((char *)Abc_ObjData(pNode)) {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "The number of fanins (%d) of node %s is different from SOP size (%d).", 
                    Abc_ObjFaninNum(pNode), Abc_ObjName(Abc_ObjFanout(pNode,0)), Abc_SopGetVarNum((char *)Abc_ObjData(pNode)) );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }

    // return the last array of tokens
    *pvTokens = vTokens;
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifReorderFormalNames(vTokens *Vec_Ptr_t, pGate *Mio_Gate_t, pTwin *Mio_Gate_t) int32 {
    var pGatePin *Mio_Pin_t;
    var pName, pNamePin string;
    var i, k, nSize, Length int32;
    nSize = Vec_PtrSize(vTokens);
    if pTwin == nil {
        if nSize - 3 != Mio_GateReadPinNum(pGate) {
            return 0; }
    
    } else {
        if nSize - 3 != Mio_GateReadPinNum(pGate) && nSize - 4 != Mio_GateReadPinNum(pGate) {
            return 0; }
    }
    // check if the names are in order
    for  pGatePin = Mio_GateReadPins(pGate), i = 0; (pGatePin != nil); pGatePin = Mio_PinReadNext(pGatePin), i++  {
        pNamePin = Mio_PinReadName(pGatePin);
        Length = len(pNamePin);
        pName = (char *)Vec_PtrEntry(vTokens, i+2);
        if !strncmp( pNamePin, pName, Length ) && pName[Length] == '=' {
            continue; }
        break;
    }
    if pTwin == nil {
        if i == Mio_GateReadPinNum(pGate) {
            return 1; }
        // reorder the pins
        for  pGatePin = Mio_GateReadPins(pGate), i = 0; (pGatePin != nil); pGatePin = Mio_PinReadNext(pGatePin), i++  {
            pNamePin = Mio_PinReadName(pGatePin);
            Length = len(pNamePin);
            for  k = 2; k < nSize; k++  {
                pName = (char *)Vec_PtrEntry(vTokens, k);
                if !strncmp( pNamePin, pName, Length ) && pName[Length] == '=' {
                    Vec_PtrPush( vTokens, pName );
                    break;
                }
            }
        }
        pNamePin = Mio_GateReadOutName(pGate);
        Length = len(pNamePin);
        for  k = 2; k < nSize; k++  {
            pName = (char *)Vec_PtrEntry(vTokens, k);
            if !strncmp( pNamePin, pName, Length ) && pName[Length] == '=' {
                Vec_PtrPush( vTokens, pName );
                break;
            }
        }
        if Vec_PtrSize(vTokens) - nSize != nSize - 2 {
            return 0; }
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = nSize; (k < Vec_PtrSize(vTokens)) && (((pName) = (char *)Vec_PtrEntry(vTokens, k)), 1); k++  {
            Vec_PtrWriteEntry( vTokens, k - nSize + 2, pName );
        }
        Vec_PtrShrink( vTokens, nSize );
    
    } else {
        if i != Mio_GateReadPinNum(pGate) { // expect the correct order of input pins in the network with twin gates
            return 0; }
        // check the last two entries
        if nSize - 3 == Mio_GateReadPinNum(pGate) { // only one output is available
        
            pNamePin = Mio_GateReadOutName(pGate);
            Length = len(pNamePin);
            pName = (char *)Vec_PtrEntry(vTokens, nSize - 1);
            if !strncmp( pNamePin, pName, Length ) && pName[Length] == '=' { // the last entry is pGate
            
                Vec_PtrPush( vTokens, nil );
                return 1;
            }
            pNamePin = Mio_GateReadOutName(pTwin);
            Length = len(pNamePin);
            pName = (char *)Vec_PtrEntry(vTokens, nSize - 1);
            if !strncmp( pNamePin, pName, Length ) && pName[Length] == '=' { // the last entry is pTwin
            
                pName = (char *)Vec_PtrPop( vTokens );
                Vec_PtrPush( vTokens, nil );
                Vec_PtrPush( vTokens, pName );
                return 1;
            }
            return 0;
        }
        if nSize - 4 == Mio_GateReadPinNum(pGate) { // two outputs are available
        
            pNamePin = Mio_GateReadOutName(pGate);
            Length = len(pNamePin);
            pName = (char *)Vec_PtrEntry(vTokens, nSize - 2);
            if !(!strncmp( pNamePin, pName, Length ) && pName[Length] == '=') {
                return 0; }
            pNamePin = Mio_GateReadOutName(pTwin);
            Length = len(pNamePin);
            pName = (char *)Vec_PtrEntry(vTokens, nSize - 1);
            if !(!strncmp( pNamePin, pName, Length ) && pName[Length] == '=') {
                return 0; }
            return 1;
        }
        assert(false, "0");
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkGate(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pGenlib *Mio_Library_t; 
    var pGate *Mio_Gate_t;
    var pNode *Abc_Obj_t;
    var ppNames *string;
    var i, nNames int32;

    // check that the library is available
    pGenlib = (Mio_Library_t *)Abc_FrameReadLibGen();
    if pGenlib == nil {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "The current library is not available." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }

    // create a new node and add it to the network
    if vTokens.nSize < 2 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "The .gate line has less than two tokens." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }

    // get the gate
    pGate = Mio_LibraryReadGateByName( pGenlib, (char *)vTokens.pArray[1], nil );
    if pGate == nil {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Cannot find gate \"%s\" in the library.", (char*)vTokens.pArray[1] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }

    // if this is the first line with gate, update the network type
    if Abc_NtkNodeNum(p.pNtkCur) == 0 {
        assert(p.pNtkCur.ntkFunc == ABC_FUNC_SOP, "p.pNtkCur.ntkFunc == ABC_FUNC_SOP");
        p.pNtkCur.ntkFunc = ABC_FUNC_MAP;
        Mem_FlexStop( (Mem_Flex_t *)p.pNtkCur.pManFunc, 0 );
        p.pNtkCur.pManFunc = pGenlib;
    }

    // reorder the formal inputs to be in the same order as in the gate
    if !(Io_ReadBlifReorderFormalNames( vTokens, pGate, Mio_GateReadTwin(pGate) ) != 0) {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Mismatch in the fanins of gate \"%s\".", (char*)vTokens.pArray[1] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }


    // remove the formal parameter names
    for  i = 2; i < vTokens.nSize; i++  {
        vTokens.pArray[i] = Io_ReadBlifCleanName( (char *)vTokens.pArray[i] );
        if vTokens.pArray[i] == nil {
            p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
            sprintf( p.sError, "Invalid gate input assignment." );
            Io_ReadBlifPrintErrorMessage( p );
            return 1;
        }
    }

    // create the node
    if Mio_GateReadTwin(pGate) == nil {
        nNames  = vTokens.nSize - 3;
        ppNames = (char **)vTokens.pArray + 2;
        pNode   = Io_ReadCreateNode( p.pNtkCur, ppNames[nNames], ppNames, nNames );
        Abc_ObjSetData( pNode, pGate );
    
    } else {
        nNames  = vTokens.nSize - 4;
        ppNames = (char **)vTokens.pArray + 2;
        assert(ppNames[nNames] != nil || ppNames[nNames+1] != nil, "ppNames[nNames] != nil || ppNames[nNames+1] != nil");
        if ppNames[nNames] {
            pNode   = Io_ReadCreateNode( p.pNtkCur, ppNames[nNames], ppNames, nNames );
            Abc_ObjSetData( pNode, pGate );
        }
        if ppNames[nNames+1] {
            pNode   = Io_ReadCreateNode( p.pNtkCur, ppNames[nNames+1], ppNames, nNames );
            Abc_ObjSetData( pNode, Mio_GateReadTwin(pGate) );
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Creates a multi-input multi-output box in the hierarchical design.]

  Description []
               
  SideEffects [] 

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkSubcircuit(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pBox *Abc_Obj_t;
    var vNames *Vec_Ptr_t;
    var pName string;
    var i int32;

    // create a new node and add it to the network
    if vTokens.nSize < 3 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "The .subcircuit line has less than three tokens." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }

    // store the names of formal/actual inputs/outputs of the box
    vNames = Vec_PtrAlloc( 10 );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vTokens)) && (((pName) = (char *)Vec_PtrEntry(vTokens, i)), 1); i++  {
//        Vec_PtrPush( vNames, Abc_NtkRegisterName(p->pNtkCur, pName) );
        Vec_PtrPush( vNames, Extra_UtilStrsav(pName) );
    }  // memory leak!!!

    // create a new box and add it to the network
    pBox = Abc_NtkCreateBlackbox( p.pNtkCur );
    // set the pointer to the node names
    Abc_ObjSetData( pBox, vNames );
    // remember the line of the file
    pBox.pCopy = (Abc_Obj_t *)(ABC_PTRINT_T)Extra_FileReaderGetLineNumber(p.pReader, 0);
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifCleanName(pName string) string {
    var i, Length int32;
    Length = len(pName);
    for  i = 0; i < Length; i++  {
        if pName[i] == '=' {
            return pName + i + 1; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkInputArrival(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pNet *Abc_Obj_t;
    var pFoo1, pFoo2 string;
    var TimeRise, TimeFall float64;
 
    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".input_arrival", 14 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".input_arrival\", 14 ) == 0");
    if vTokens.nSize != 4 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments on .input_arrival line." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    pNet = Abc_NtkFindNet( p.pNtkCur, (char *)vTokens.pArray[1] );
    if pNet == nil {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Cannot find object corresponding to %s on .input_arrival line.", (char*)vTokens.pArray[1] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    TimeRise = strtod( (char *)vTokens.pArray[2], &pFoo1 );
    TimeFall = strtod( (char *)vTokens.pArray[3], &pFoo2 );
    if *pFoo1 != '\0' || *pFoo2 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s %s) for rise or fall time on .input_arrival line.", (char*)vTokens.pArray[2], (char*)vTokens.pArray[3] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
    //Abc_NtkTimeSetArrival( p->pNtkCur, Abc_ObjFanin0(pNet)->Id, (float)TimeRise, (float)TimeFall );
    Vec_IntPush( p.vInArrs, Abc_ObjFanin0(pNet).Id );
    Vec_IntPush( p.vInArrs, Abc_Float2Int(float32(TimeRise)) );
    Vec_IntPush( p.vInArrs, Abc_Float2Int(float32(TimeFall)) );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkOutputRequired(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pNet *Abc_Obj_t;
    var pFoo1, pFoo2 string;
    var TimeRise, TimeFall float64;
 
    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".output_required", 16 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".output_required\", 16 ) == 0");
    if vTokens.nSize != 4 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments on .output_required line." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    pNet = Abc_NtkFindNet( p.pNtkCur, (char *)vTokens.pArray[1] );
    if pNet == nil {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Cannot find object corresponding to %s on .output_required line.", (char*)vTokens.pArray[1] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    TimeRise = strtod( (char *)vTokens.pArray[2], &pFoo1 );
    TimeFall = strtod( (char *)vTokens.pArray[3], &pFoo2 );
    if *pFoo1 != '\0' || *pFoo2 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s %s) for rise or fall time on .output_required line.", (char*)vTokens.pArray[2], (char*)vTokens.pArray[3] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
//    Abc_NtkTimeSetRequired( p->pNtkCur, Abc_ObjFanout0(pNet)->Id, (float)TimeRise, (float)TimeFall );
    Vec_IntPush( p.vOutReqs, Abc_ObjFanout0(pNet).Id );
    Vec_IntPush( p.vOutReqs, Abc_Float2Int(float32(TimeRise)) );
    Vec_IntPush( p.vOutReqs, Abc_Float2Int(float32(TimeFall)) );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkDefaultInputArrival(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pFoo1, pFoo2 string;
    var TimeRise, TimeFall float64;

    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".default_input_arrival", 23 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".default_input_arrival\", 23 ) == 0");
    if vTokens.nSize != 3 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments on .default_input_arrival line." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    TimeRise = strtod( (char *)vTokens.pArray[1], &pFoo1 );
    TimeFall = strtod( (char *)vTokens.pArray[2], &pFoo2 );
    if *pFoo1 != '\0' || *pFoo2 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s %s) for rise or fall time on .default_input_arrival line.", (char*)vTokens.pArray[1], (char*)vTokens.pArray[2] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
    //Abc_NtkTimeSetDefaultArrival( p->pNtkCur, (float)TimeRise, (float)TimeFall );
    p.DefInArrRise = float32(TimeRise);
    p.DefInArrFall = float32(TimeFall);
    p.fHaveDefInArr = 1;
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkDefaultOutputRequired(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pFoo1, pFoo2 string;
    var TimeRise, TimeFall float64;

    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".default_output_required", 25 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".default_output_required\", 25 ) == 0");
    if vTokens.nSize != 3 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments on .default_output_required line." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    TimeRise = strtod( (char *)vTokens.pArray[1], &pFoo1 );
    TimeFall = strtod( (char *)vTokens.pArray[2], &pFoo2 );
    if *pFoo1 != '\0' || *pFoo2 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s %s) for rise or fall time on .default_output_required line.", (char*)vTokens.pArray[1], (char*)vTokens.pArray[2] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
//    Abc_NtkTimeSetDefaultRequired( p->pNtkCur, (float)TimeRise, (float)TimeFall );
    p.DefOutReqRise = float32(TimeRise);
    p.DefOutReqFall = float32(TimeFall);
    p.fHaveDefOutReq = 1;
    return 0;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadFindCiId(pNtk *Abc_Ntk_t, pObj *Abc_Obj_t) int32 {
    var pTemp *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pTemp) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if pTemp == pObj {
            return i; }
    }
    return -1;
}
func Io_ReadBlifNetworkInputDrive(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pNet *Abc_Obj_t;
    var pFoo1, pFoo2 string;
    var TimeRise, TimeFall float64;
 
    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".input_drive", 12 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".input_drive\", 12 ) == 0");
    if vTokens.nSize != 4 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments on .input_drive line." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    pNet = Abc_NtkFindNet( p.pNtkCur, (char *)vTokens.pArray[1] );
    if pNet == nil {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Cannot find object corresponding to %s on .input_drive line.", (char*)vTokens.pArray[1] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    TimeRise = strtod( (char *)vTokens.pArray[2], &pFoo1 );
    TimeFall = strtod( (char *)vTokens.pArray[3], &pFoo2 );
    if *pFoo1 != '\0' || *pFoo2 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s %s) for rise or fall time on .input_drive line.", (char*)vTokens.pArray[2], (char*)vTokens.pArray[3] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
    //Abc_NtkTimeSetInputDrive( p->pNtkCur, Io_ReadFindCiId(p->pNtkCur, Abc_NtkObj(p->pNtkCur, Abc_ObjFanin0(pNet)->Id)), (float)TimeRise, (float)TimeFall );
    Vec_IntPush( p.vInDrives, Io_ReadFindCiId(p.pNtkCur, Abc_NtkObj(p.pNtkCur, Abc_ObjFanin0(pNet).Id)) );
    Vec_IntPush( p.vInDrives, Abc_Float2Int(float32(TimeRise)) );
    Vec_IntPush( p.vInDrives, Abc_Float2Int(float32(TimeFall)) );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadFindCoId(pNtk *Abc_Ntk_t, pObj *Abc_Obj_t) int32 {
    var pTemp *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTemp) = Abc_NtkPo(pNtk, i)), 1); i++  {
        if pTemp == pObj {
            return i; }
    }
    return -1;
}
func Io_ReadBlifNetworkOutputLoad(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pNet *Abc_Obj_t;
    var pFoo1, pFoo2 string;
    var TimeRise, TimeFall float64;
 
    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".output_load", 12 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".output_load\", 12 ) == 0");
    if vTokens.nSize != 4 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments on .output_load line." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    pNet = Abc_NtkFindNet( p.pNtkCur, (char *)vTokens.pArray[1] );
    if pNet == nil {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Cannot find object corresponding to %s on .output_load line.", (char*)vTokens.pArray[1] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    TimeRise = strtod( (char *)vTokens.pArray[2], &pFoo1 );
    TimeFall = strtod( (char *)vTokens.pArray[3], &pFoo2 );
    if *pFoo1 != '\0' || *pFoo2 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s %s) for rise or fall time on .output_load line.", (char*)vTokens.pArray[2], (char*)vTokens.pArray[3] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
//    Abc_NtkTimeSetOutputLoad( p->pNtkCur, Io_ReadFindCoId(p->pNtkCur, Abc_NtkObj(p->pNtkCur, Abc_ObjFanout0(pNet)->Id)), (float)TimeRise, (float)TimeFall );
    Vec_IntPush( p.vOutLoads, Io_ReadFindCoId(p.pNtkCur, Abc_NtkObj(p.pNtkCur, Abc_ObjFanout0(pNet).Id)) );
    Vec_IntPush( p.vOutLoads, Abc_Float2Int(float32(TimeRise)) );
    Vec_IntPush( p.vOutLoads, Abc_Float2Int(float32(TimeFall)) );
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkDefaultInputDrive(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pFoo1, pFoo2 string;
    var TimeRise, TimeFall float64;

    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".default_input_drive", 21 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".default_input_drive\", 21 ) == 0");
    if vTokens.nSize != 3 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments on .default_input_drive line." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    TimeRise = strtod( (char *)vTokens.pArray[1], &pFoo1 );
    TimeFall = strtod( (char *)vTokens.pArray[2], &pFoo2 );
    if *pFoo1 != '\0' || *pFoo2 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s %s) for rise or fall time on .default_input_drive line.", (char*)vTokens.pArray[1], (char*)vTokens.pArray[2] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
//    Abc_NtkTimeSetDefaultInputDrive( p->pNtkCur, (float)TimeRise, (float)TimeFall );
    p.DefInDriRise = float32(TimeRise);
    p.DefInDriFall = float32(TimeFall);
    p.fHaveDefInDri = 1;
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkDefaultOutputLoad(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pFoo1, pFoo2 string;
    var TimeRise, TimeFall float64;

    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".default_output_load", 21 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".default_output_load\", 21 ) == 0");
    if vTokens.nSize != 3 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments on .default_output_load line." );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    TimeRise = strtod( (char *)vTokens.pArray[1], &pFoo1 );
    TimeFall = strtod( (char *)vTokens.pArray[2], &pFoo2 );
    if *pFoo1 != '\0' || *pFoo2 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s %s) for rise or fall time on .default_output_load line.", (char*)vTokens.pArray[1], (char*)vTokens.pArray[2] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
//    Abc_NtkTimeSetDefaultOutputLoad( p->pNtkCur, (float)TimeRise, (float)TimeFall );
    p.DefOutLoadRise = float32(TimeRise);
    p.DefOutLoadFall = float32(TimeFall);
    p.fHaveDefOutLoad = 1;
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkAndGateDelay(p *Io_ReadBlif_t, vTokens *Vec_Ptr_t) int32 {
    var pFoo1 string;
    var AndGateDelay float64;

    // make sure this is indeed the .inputs line
    assert(strncmp( (char *)vTokens.pArray[0], ".and_gate_delay", 25 ) == 0, "strncmp( (char *)vTokens.pArray[0], \".and_gate_delay\", 25 ) == 0");
    if vTokens.nSize != 2 {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Wrong number of arguments (%d) on .and_gate_delay line (should be 1).", vTokens.nSize-1 );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    AndGateDelay = strtod( (char *)vTokens.pArray[1], &pFoo1 );
    if *pFoo1 != '\0' {
        p.LineCur = Extra_FileReaderGetLineNumber(p.pReader, 0);
        sprintf( p.sError, "Bad value (%s) for AND gate delay in on .and_gate_delay line line.", (char*)vTokens.pArray[1] );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }
    // set timing info
    p.pNtkCur.AndGateDelay = float32(AndGateDelay);
    return 0;
}

/**Function*************************************************************

  Synopsis    [Prints the error message including the file name and line number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifPrintErrorMessage(p *Io_ReadBlif_t) {
    p.fError = 1;
    if p.LineCur == 0 { // the line number is not given
        fmt.Fprintf(p.Output, "%s: %s\n", p.pFileName, p.sError);
    } else { // print the error message with the line number
        fmt.Fprintf(p.Output, "%s (line %d): %s\n", p.pFileName, p.LineCur, p.sError); }
}

/**Function*************************************************************

  Synopsis    [Gets the tokens taking into account the line breaks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifGetTokens(p *Io_ReadBlif_t) *Vec_Ptr_t {
    var vTokens *Vec_Ptr_t;
    var pLastToken string;
    var i int32;

    // get rid of the old tokens
    if p.vNewTokens.nSize > 0 {
        for  i = 0; i < p.vNewTokens.nSize; i++  {
            p.vNewTokens.pArray[i] = nil;
        }
        p.vNewTokens.nSize = 0;
    }

    // get the new tokens
    vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p.pReader);
    if vTokens == nil {
        return vTokens; }

    // check if there is a transfer to another line
    pLastToken = (char *)vTokens.pArray[vTokens.nSize - 1];
    if pLastToken[ len(pLastToken)-1 ] != '\\' {
        return vTokens; }

    // remove the slash
    pLastToken[ len(pLastToken)-1 ] = 0;
    if pLastToken[0] == 0 {
        vTokens.nSize--; }
    // load them into the new array
    for  i = 0; i < vTokens.nSize; i++  {
        Vec_PtrPush( p.vNewTokens, Extra_UtilStrsav((char *)vTokens.pArray[i]) );
    }

    // load as long as there is the line break
    for {
        // get the new tokens
        vTokens = (Vec_Ptr_t *)Extra_FileReaderGetTokens(p.pReader);
        if vTokens.nSize == 0 {
            return p.vNewTokens; }
        // check if there is a transfer to another line
        pLastToken = (char *)vTokens.pArray[vTokens.nSize - 1];
        if pLastToken[ len(pLastToken)-1 ] == '\\' {
            // remove the slash
            pLastToken[ len(pLastToken)-1 ] = 0;
            if pLastToken[0] == 0 {
                vTokens.nSize--; }
            // load them into the new array
            for  i = 0; i < vTokens.nSize; i++  {
                Vec_PtrPush( p.vNewTokens, Extra_UtilStrsav((char *)vTokens.pArray[i]) );
            }
            continue;
        }
        // otherwise, load them and break
        for  i = 0; i < vTokens.nSize; i++  {
            Vec_PtrPush( p.vNewTokens, Extra_UtilStrsav((char *)vTokens.pArray[i]) );
        }
        break;
    }
    return p.vNewTokens;
}

/**Function*************************************************************

  Synopsis    [Starts the reading data structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifFile(pFileName string) *Io_ReadBlif_t {
    var pReader *Extra_FileReader_t;
    var p *Io_ReadBlif_t;

    // start the reader
    pReader = Extra_FileReaderAlloc( pFileName, "#", "\n\r", " \t" );

    if pReader == nil {
        return nil; }

    // start the reading data structure
    // inlining of macro ABC_ALLOC
    p = ((Io_ReadBlif_t *) malloc(sizeof(Io_ReadBlif_t) * (size_t)(1)));
    *p = Io_ReadBlif_t{};
    p.pFileName  = pFileName;
    p.pReader    = pReader;
    p.Output     = os.Stdout;
    p.vNewTokens = Vec_PtrAlloc( 100 );
    p.vCubes     = Vec_StrAlloc( 100 );
    p.vInArrs    = Vec_IntAlloc( 100 );
    p.vOutReqs   = Vec_IntAlloc( 100 );
    p.vInDrives  = Vec_IntAlloc( 100 );
    p.vOutLoads  = Vec_IntAlloc( 100 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Frees the data structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifFree(p *Io_ReadBlif_t) {
    Extra_FileReaderFree( p.pReader );
    Vec_PtrFree( p.vNewTokens );
    Vec_StrFree( p.vCubes );
    Vec_IntFree( p.vInArrs );
    Vec_IntFree( p.vOutReqs );
    Vec_IntFree( p.vInDrives );
    Vec_IntFree( p.vOutLoads );
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Connect one box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkConnectBoxesOneBox(p *Io_ReadBlif_t, pBox *Abc_Obj_t, tName2Model *stmm_table) int32 {
    var pNames *Vec_Ptr_t;
    var pNtkModel *Abc_Ntk_t;
    var pObj, pNet *Abc_Obj_t;
    var pName string = nil;
    var pActual string;
    var i int32;
    var Length int32;
    var Start int32 = -1;

    // get the model for this box
    pNames = (Vec_Ptr_t *)pBox.pData;
    if !stmm_lookup( tName2Model, (char *)Vec_PtrEntry(pNames, 0), (char **)&pNtkModel ) {
        p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
        sprintf( p.sError, "Cannot find the model for subcircuit %s.", (char*)Vec_PtrEntry(pNames, 0) );
        Io_ReadBlifPrintErrorMessage( p );
        return 1;
    }

    // create the fanins of the box
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtkModel)) && (((pObj) = Abc_NtkPi(pNtkModel, i)), 1); i++  {
        pObj.pCopy = nil;
    }
    if Abc_NtkPiNum(pNtkModel) == 0 {
        Start = 1;
    } else {
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = 1; (i < Vec_PtrSize(pNames)) && (((pName) = (char *)Vec_PtrEntry(pNames, i)), 1); i++  {
            pActual = Io_ReadBlifCleanName(pName);
            if pActual == nil {
                p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
                sprintf( p.sError, "Cannot parse formal/actual name pair \"%s\".", pName );
                Io_ReadBlifPrintErrorMessage( p );
                return 1;
            }
            Length = pActual - pName - 1;
            pName[Length] = 0;
            // find the PI net with this name
            pObj = Abc_NtkFindNet( pNtkModel, pName );
            if pObj == nil {
                p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
                sprintf( p.sError, "Cannot find formal input \"%s\" as an PI of model \"%s\".", pName, (char*)Vec_PtrEntry(pNames, 0) );
                Io_ReadBlifPrintErrorMessage( p );
                return 1;
            }
            // get the PI
            pObj = Abc_ObjFanin0(pObj);
            // quit if this is not a PI net
            if !Abc_ObjIsPi(pObj) {
                pName[Length] = '=';
                Start = i;
                break;
            }
            // remember the actual name in the net
            if pObj.pCopy != nil {
                p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
                sprintf( p.sError, "Formal input \"%s\" is used more than once.", pName );
                Io_ReadBlifPrintErrorMessage( p );
                return 1;
            }
            pObj.pCopy = (Abc_Obj_t *)pActual;
            // quit if we processed all PIs
            if i == Abc_NtkPiNum(pNtkModel) {
                Start = i+1;
                break;
            }
        }
    }
    // create the fanins of the box
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtkModel)) && (((pObj) = Abc_NtkPi(pNtkModel, i)), 1); i++  {
        pActual = (char *)pObj.pCopy;
        if pActual == nil {
            p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
            sprintf( p.sError, "Formal input \"%s\" of model %s is not driven.", pName, (char*)Vec_PtrEntry(pNames, 0) );
            Io_ReadBlifPrintErrorMessage( p );
            return 1;
        }
        pNet = Abc_NtkFindOrCreateNet( pBox.pNtk, pActual );
        Abc_ObjAddFanin( pBox, pNet );
    }
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtkModel)) && (((pObj) = Abc_NtkPi(pNtkModel, i)), 1); i++  {
        pObj.pCopy = nil;
    }

    // create the fanouts of the box
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtkModel)) && (((pObj) = Abc_NtkPo(pNtkModel, i)), 1); i++  {
        pObj.pCopy = nil;
    }
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Start; (i < Vec_PtrSize(pNames)) && (((pName) = (char *)Vec_PtrEntry(pNames, i)), 1); i++  {
        pActual = Io_ReadBlifCleanName(pName);
        if pActual == nil {
            p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
            sprintf( p.sError, "Cannot parse formal/actual name pair \"%s\".", pName );
            Io_ReadBlifPrintErrorMessage( p );
            return 1;
        }
        Length = pActual - pName - 1;
        pName[Length] = 0;
        // find the PO net with this name
        pObj = Abc_NtkFindNet( pNtkModel, pName );
        if pObj == nil {
            p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
            sprintf( p.sError, "Cannot find formal output \"%s\" as an PO of model \"%s\".", pName, (char*)Vec_PtrEntry(pNames, 0) );
            Io_ReadBlifPrintErrorMessage( p );
            return 1;
        }
        // get the PO
        pObj = Abc_ObjFanout0(pObj);
        if pObj.pCopy != nil {
            p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
            sprintf( p.sError, "Formal output \"%s\" is used more than once.", pName );
            Io_ReadBlifPrintErrorMessage( p );
            return 1;
        }
        pObj.pCopy = (Abc_Obj_t *)pActual;
    }
    // create the fanouts of the box
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtkModel)) && (((pObj) = Abc_NtkPo(pNtkModel, i)), 1); i++  {
        pActual = (char *)pObj.pCopy;
        if pActual == nil {
            p.LineCur = int32((ABC_PTRINT_T))pBox.pCopy;
            sprintf( p.sError, "Formal output \"%s\" of model %s is not driven.", pName, (char*)Vec_PtrEntry(pNames, 0) );
            Io_ReadBlifPrintErrorMessage( p );
            return 1;
        }
        pNet = Abc_NtkFindOrCreateNet( pBox.pNtk, pActual );
        Abc_ObjAddFanin( pNet, pBox );
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtkModel)) && (((pObj) = Abc_NtkPo(pNtkModel, i)), 1); i++  {
        pObj.pCopy = nil;
    }

    // remove the array of names, assign the pointer to the model
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize((Vec_Ptr_t *)pBox.pData)) && (((pName) = (char *)Vec_PtrEntry((Vec_Ptr_t *)pBox.pData, i)), 1); i++  {
        pName = nil;
    }
    Vec_PtrFree( (Vec_Ptr_t *)pBox.pData );
    pBox.pData = pNtkModel;
    return 0;
}

/**Function*************************************************************

  Synopsis    [Connect the boxes in the hierarchy of networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifNetworkConnectBoxesOne(p *Io_ReadBlif_t, pNtk *Abc_Ntk_t, tName2Model *stmm_table) int32 {
    var pBox *Abc_Obj_t;
    var i int32;
    // go through the boxes
    // inlining of macro Abc_NtkForEachBlackbox
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsBlackbox(pBox) {} else
        if Io_ReadBlifNetworkConnectBoxesOneBox( p, pBox, tName2Model ) != 0 {
            return 1; }
    }
    Abc_NtkFinalizeRead( pNtk );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Creates timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_ReadBlifCreateTiming(p *Io_ReadBlif_t, pNtk *Abc_Ntk_t) int32 {
    var Id, Rise, Fall, i int32;

    // set timing info
    //Abc_NtkTimeSetDefaultArrival( p->pNtkCur, (float)TimeRise, (float)TimeFall );
//    p->DefInArrRise = (float)TimeRise;
//    p->DefInArrFall = (float)TimeFall;
    if p.fHaveDefInArr {
    Abc_NtkTimeSetDefaultArrival( pNtk, p.DefInArrRise, p.DefInArrFall ); }
    // set timing info
    //Abc_NtkTimeSetDefaultRequired( p->pNtkCur, (float)TimeRise, (float)TimeFall );
//    p->DefOutReqRise = (float)TimeRise;
//    p->DefOutReqFall = (float)TimeFall;
    if p.fHaveDefOutReq {
    Abc_NtkTimeSetDefaultRequired( pNtk, p.DefOutReqRise, p.DefOutReqFall ); }
    // set timing info
    //Abc_NtkTimeSetDefaultInputDrive( p->pNtkCur, (float)TimeRise, (float)TimeFall );
//    p->DefInDriRise = (float)TimeRise;
//    p->DefInDriFall = (float)TimeFall;
    if p.fHaveDefInDri {
    Abc_NtkTimeSetDefaultInputDrive( pNtk, p.DefInDriRise, p.DefInDriFall ); }
    // set timing info
    //Abc_NtkTimeSetDefaultOutputLoad( p->pNtkCur, (float)TimeRise, (float)TimeFall );
//    p->DefOutLoadRise = (float)TimeRise;
//    p->DefOutLoadFall = (float)TimeFall;
    if p.fHaveDefOutLoad {
    Abc_NtkTimeSetDefaultOutputLoad( pNtk, p.DefOutLoadRise, p.DefOutLoadFall ); }

    // set timing info
    //Abc_NtkTimeSetArrival( p->pNtkCur, Abc_ObjFanin0(pNet)->Id, (float)TimeRise, (float)TimeFall );
//    Vec_IntPush( p->vInArrs, Abc_ObjFanin0(pNet)->Id );
//    Vec_IntPush( p->vInArrs, Abc_Float2Int((float)TimeRise) );
//    Vec_IntPush( p->vInArrs, Abc_Float2Int((float)TimeFall) );
    // inlining of macro Vec_IntForEachEntryTriple
    for  i = 0; (i+2 < Vec_IntSize(p.vInArrs)) && (((Id) = Vec_IntEntry(p.vInArrs, i)), 1) && (((Rise) = Vec_IntEntry(p.vInArrs, i+1)), 1) && (((Fall) = Vec_IntEntry(p.vInArrs, i+2)), 1); i += 3  {
        Abc_NtkTimeSetArrival( pNtk, Id, Abc_Int2Float(Rise), Abc_Int2Float(Fall) );
    }
    // set timing info
    //Abc_NtkTimeSetRequired( p->pNtkCur, Abc_ObjFanout0(pNet)->Id, (float)TimeRise, (float)TimeFall );
//    Vec_IntPush( p->vOutReqs, Abc_ObjFanout0(pNet)->Id );
//    Vec_IntPush( p->vOutReqs, Abc_Float2Int((float)TimeRise) );
//    Vec_IntPush( p->vOutReqs, Abc_Float2Int((float)TimeFall) );
    // inlining of macro Vec_IntForEachEntryTriple
    for  i = 0; (i+2 < Vec_IntSize(p.vOutReqs)) && (((Id) = Vec_IntEntry(p.vOutReqs, i)), 1) && (((Rise) = Vec_IntEntry(p.vOutReqs, i+1)), 1) && (((Fall) = Vec_IntEntry(p.vOutReqs, i+2)), 1); i += 3  {
        Abc_NtkTimeSetRequired( pNtk, Id, Abc_Int2Float(Rise), Abc_Int2Float(Fall) );
    }
    // set timing info
    //Abc_NtkTimeSetInputDrive( p->pNtkCur, Io_ReadFindCiId(p->pNtkCur, Abc_NtkObj(p->pNtkCur, Abc_ObjFanin0(pNet)->Id)), (float)TimeRise, (float)TimeFall );
//    Vec_IntPush( p->vInDrives, Io_ReadFindCiId(p->pNtkCur, Abc_NtkObj(p->pNtkCur, Abc_ObjFanin0(pNet)->Id)) );
//    Vec_IntPush( p->vInDrives, Abc_Float2Int((float)TimeRise) );
//    Vec_IntPush( p->vInDrives, Abc_Float2Int((float)TimeFall) );
    // inlining of macro Vec_IntForEachEntryTriple
    for  i = 0; (i+2 < Vec_IntSize(p.vInDrives)) && (((Id) = Vec_IntEntry(p.vInDrives, i)), 1) && (((Rise) = Vec_IntEntry(p.vInDrives, i+1)), 1) && (((Fall) = Vec_IntEntry(p.vInDrives, i+2)), 1); i += 3  {
        Abc_NtkTimeSetInputDrive( pNtk, Id, Abc_Int2Float(Rise), Abc_Int2Float(Fall) );
    }
    // set timing info
    //Abc_NtkTimeSetOutputLoad( p->pNtkCur, Io_ReadFindCoId(p->pNtkCur, Abc_NtkObj(p->pNtkCur, Abc_ObjFanout0(pNet)->Id)), (float)TimeRise, (float)TimeFall );
//    Vec_IntPush( p->vOutLoads, Io_ReadFindCoId(p->pNtkCur, Abc_NtkObj(p->pNtkCur, Abc_ObjFanout0(pNet)->Id)) );
//    Vec_IntPush( p->vOutLoads, Abc_Float2Int((float)TimeRise) );
//    Vec_IntPush( p->vOutLoads, Abc_Float2Int((float)TimeFall) );
    // inlining of macro Vec_IntForEachEntryTriple
    for  i = 0; (i+2 < Vec_IntSize(p.vOutLoads)) && (((Id) = Vec_IntEntry(p.vOutLoads, i)), 1) && (((Rise) = Vec_IntEntry(p.vOutLoads, i+1)), 1) && (((Fall) = Vec_IntEntry(p.vOutLoads, i+2)), 1); i += 3  {
        Abc_NtkTimeSetOutputLoad( pNtk, Id, Abc_Int2Float(Rise), Abc_Int2Float(Fall) );
    }
    return 1;
}

     

                                                                        

                                                               

                
               
                

                

                                                                        
                                                                               
 
                         
                     
                 
                                 
                                                                                      
                 
                             
                                                                             
                                                                                    
                     
             
 

      

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

