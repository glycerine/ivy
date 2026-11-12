package ver

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [verCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Parses several flavors of structural Verilog.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 19, 2006.]

  Revision    [$Id: verCore.c,v 1.00 2006/08/19 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ver.h"
//#include "map/mio/mio.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// types of verilog signals
 type Ver_SignalType_t int32
const ( 
    VER_SIG_NONE Ver_SignalType_t = 0

    VER_SIG_INPUT Ver_SignalType_t = VER_SIG_NONE + 1

    VER_SIG_OUTPUT Ver_SignalType_t = VER_SIG_INPUT + 1

    VER_SIG_INOUT Ver_SignalType_t = VER_SIG_OUTPUT + 1

    VER_SIG_REG Ver_SignalType_t = VER_SIG_INOUT + 1

    VER_SIG_WIRE Ver_SignalType_t = VER_SIG_REG + 1

)

 

// types of verilog gates
 type Ver_GateType_t int32
const ( 
    VER_GATE_AND Ver_GateType_t = 0

    VER_GATE_OR Ver_GateType_t = VER_GATE_AND + 1

    VER_GATE_XOR Ver_GateType_t = VER_GATE_OR + 1

    VER_GATE_BUF Ver_GateType_t = VER_GATE_XOR + 1

    VER_GATE_NAND Ver_GateType_t = VER_GATE_BUF + 1

    VER_GATE_NOR Ver_GateType_t = VER_GATE_NAND + 1

    VER_GATE_XNOR Ver_GateType_t = VER_GATE_NOR + 1

    VER_GATE_NOT Ver_GateType_t = VER_GATE_XNOR + 1

)

 

// static Ver_Man_t * Ver_ParseStart( char * pFileName, Abc_Des_t * pGateLib );

// static void Ver_ParseStop( Ver_Man_t * p );

// static void Ver_ParseFreeData( Ver_Man_t * p );

// static void Ver_ParseInternal( Ver_Man_t * p );

// static int  Ver_ParseModule( Ver_Man_t * p );

// static int  Ver_ParseSignal( Ver_Man_t * p, Abc_Ntk_t * pNtk, Ver_SignalType_t SigType );

// static int  Ver_ParseAlways( Ver_Man_t * p, Abc_Ntk_t * pNtk );

// static int  Ver_ParseInitial( Ver_Man_t * p, Abc_Ntk_t * pNtk );

// static int  Ver_ParseAssign( Ver_Man_t * p, Abc_Ntk_t * pNtk );

// static int  Ver_ParseGateStandard( Ver_Man_t * pMan, Abc_Ntk_t * pNtk, Ver_GateType_t GateType );

// static int  Ver_ParseFlopStandard( Ver_Man_t * pMan, Abc_Ntk_t * pNtk );

// static int  Ver_ParseGate( Ver_Man_t * p, Abc_Ntk_t * pNtk, Mio_Gate_t * pGate );

// static int  Ver_ParseBox( Ver_Man_t * pMan, Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkBox );

// static int  Ver_ParseConnectBox( Ver_Man_t * pMan, Abc_Obj_t * pBox );

// static int  Ver_ParseAttachBoxes( Ver_Man_t * pMan );


// static Abc_Obj_t * Ver_ParseCreatePi( Abc_Ntk_t * pNtk, char * pName );

// static Abc_Obj_t * Ver_ParseCreatePo( Abc_Ntk_t * pNtk, char * pName );

// static Abc_Obj_t * Ver_ParseCreateLatch( Abc_Ntk_t * pNtk, Abc_Obj_t * pNetLI, Abc_Obj_t * pNetLO );

// static Abc_Obj_t * Ver_ParseCreateInv( Abc_Ntk_t * pNtk, Abc_Obj_t * pNet );


// static void Ver_ParseRemoveSuffixTable( Ver_Man_t * pMan );


func Ver_NtkIsDefined(pNtkBox *Abc_Ntk_t) int32 { assert(pNtkBox.pName, "pNtkBox.pName");     return Abc_NtkPiNum(pNtkBox) || Abc_NtkPoNum(pNtkBox);  }
func Ver_ObjIsConnected(pObj *Abc_Obj_t) int32 { assert(Abc_ObjIsBox(pObj), "Abc_ObjIsBox(pObj)"); return Abc_ObjFaninNum(pObj) || Abc_ObjFanoutNum(pObj); }

var glo_fMapped int32; // this is bad!

      
type Ver_Bundle_t struct {
    pNameFormal           string;   // the name of the formal net
    vNetsActual      *Vec_Ptr_t;   // the vector of actual nets (MSB to LSB)
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start parser.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseStart(pFileName string, pGateLib *Abc_Des_t) *Ver_Man_t {
    var p *Ver_Man_t;
    // inlining of macro ABC_ALLOC
    p = ((Ver_Man_t *) malloc(sizeof(Ver_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Ver_Man_t) );
    p.pFileName = pFileName;
    p.pReader   = Ver_StreamAlloc( pFileName );
    if p.pReader == nil {
        // inlining of macro ABC_FREE
        ((p) ? (free((char *) (p)), (p) = 0) : 0);
        return nil;
    }
    p.Output    = os.Stdout;
    p.vNames    = Vec_PtrAlloc( 100 );
    p.vStackFn  = Vec_PtrAlloc( 100 );
    p.vStackOp  = Vec_IntAlloc( 100 );
    p.vPerm     = Vec_IntAlloc( 100 );
    // create the design library and assign the technology library
    p.pDesign   = Abc_DesCreate( pFileName );
    p.pDesign.pLibrary = pGateLib;
    // derive library from SCL
//    if ( Abc_FrameReadLibScl() )
//        Abc_SclInstallGenlib( Abc_FrameReadLibScl(), 0, 0, 0 );
    p.pDesign.pGenlib = Abc_FrameReadLibGen();
    return p;
}

/**Function*************************************************************

  Synopsis    [Stop parser.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseStop(p *Ver_Man_t) {
    if p.pProgress {
        Extra_ProgressBarStop( p.pProgress ); }
    Ver_StreamFree( p.pReader );
    Vec_PtrFree( p.vNames   );
    Vec_PtrFree( p.vStackFn );
    Vec_IntFree( p.vStackOp );
    Vec_IntFree( p.vPerm );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
 
/**Function*************************************************************

  Synopsis    [File parser.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseFile(pFileName string, pGateLib *Abc_Des_t, fCheck int32, fUseMemMan int32) *Abc_Des_t {
    var p *Ver_Man_t;
    var pDesign *Abc_Des_t;
    // start the parser
    p = Ver_ParseStart( pFileName, pGateLib );
    p.fMapped    = glo_fMapped;
    p.fCheck     = fCheck;
    p.fUseMemMan = fUseMemMan;
    if glo_fMapped != 0 {
        Hop_ManStop((Hop_Man_t *)p.pDesign.pManFunc);
        p.pDesign.pManFunc = nil;
    }
    // parse the file
    Ver_ParseInternal( p );
    // save the result
    pDesign = p.pDesign;
    p.pDesign = nil;
    // stop the parser
    Ver_ParseStop( p );
    return pDesign;
}

/**Function*************************************************************

  Synopsis    [File parser.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseInternal(pMan *Ver_Man_t) {
    var pNtk *Abc_Ntk_t;
    var pToken string;
    var i int32;

    // preparse the modeles
    pMan.pProgress = Extra_ProgressBarStart( os.Stdout, Ver_StreamGetFileSize(pMan.pReader) );
    for {
        // get the next token
        pToken = Ver_ParseGetName( pMan );
        if pToken == nil {
            break; }
        if int32(strings.Compare(pToken, "module")) != 0 {
            sprintf( pMan.sError, "Cannot read \"module\" directive." );
            Ver_ParsePrintErrorMessage( pMan );
            return;
        }
        // parse the module
        if !(Ver_ParseModule(pMan) != 0) {
            return; }
    }
    Extra_ProgressBarStop( pMan.pProgress );
    pMan.pProgress = nil;

    // process defined and undefined boxes
    if !(Ver_ParseAttachBoxes( pMan ) != 0) {
        return; }

    // connect the boxes and check
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        // fix the dangling nets
        Abc_NtkFinalizeRead( pNtk );
        // check the network for correctness
        if pMan.fCheck && !Abc_NtkCheckRead( pNtk ) {
            pMan.fTopLevel = 1;
            sprintf( pMan.sError, "The network check has failed for network %s.", pNtk.pName );
            Ver_ParsePrintErrorMessage( pMan );
            return;
        }
    }
}

/**Function*************************************************************

  Synopsis    [File parser.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseFreeData(p *Ver_Man_t) {
    if p.pDesign {
        Abc_DesFree( p.pDesign, nil );
        p.pDesign = nil;
    }
}

/**Function*************************************************************

  Synopsis    [Prints the error message including the file name and line number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParsePrintErrorMessage(p *Ver_Man_t) {
    p.fError = 1;
    if p.fTopLevel { // the line number is not given
        fmt.Fprintf(p.Output, "%s: %s\n", p.pFileName, p.sError);
    } else { // print the error message with the line number
        fmt.Fprintf(p.Output, "%s (line %d): %s\n", p.pFileName, Ver_StreamGetLineNumber(p.pReader), p.sError); }
    // free the data
    Ver_ParseFreeData( p );
}

/**Function*************************************************************

  Synopsis    [Finds the network by name or create a new blackbox network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseFindOrCreateNetwork(pMan *Ver_Man_t, pName string) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    // check if the network exists
    if pNtkNew = Abc_DesFindModelByName( pMan.pDesign, pName ) {
        return pNtkNew; }
//printf( "Creating network %s.\n", pName );
    // create new network
    pNtkNew = Abc_NtkAlloc( ABC_NTK_NETLIST, ABC_FUNC_BLACKBOX, pMan.fUseMemMan );
    pNtkNew.pName = Extra_UtilStrsav( pName );
    pNtkNew.pSpec = nil;
    // add module to the design
    Abc_DesAddModel( pMan.pDesign, pNtkNew );
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Finds the network by name or create a new blackbox network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseFindNet(pNtk *Abc_Ntk_t, pName string) *Abc_Obj_t {
    var pObj *Abc_Obj_t;
    if pObj = Abc_NtkFindNet(pNtk, pName) {
        return pObj; }
    if !int32(strings.Compare(pName, "1\'b0")) || !int32(strings.Compare(pName, "1\'bx")) {
        return Abc_NtkFindOrCreateNet( pNtk, "1\'b0" ); }
    if !int32(strings.Compare(pName, "1\'b1")) {
        return Abc_NtkFindOrCreateNet( pNtk, "1\'b1" ); }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the blackbox type into a different one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseConvertNetwork(pMan *Ver_Man_t, pNtk *Abc_Ntk_t, fMapped int32) int32 {
    if fMapped != 0 {
        // convert from the blackbox into the network with local functions representated by AIGs
        if pNtk.ntkFunc == ABC_FUNC_BLACKBOX {
            // change network type
            assert(pNtk.pManFunc == nil, "pNtk.pManFunc == nil");
            pNtk.ntkFunc = ABC_FUNC_MAP;
            pNtk.pManFunc = pMan.pDesign.pGenlib;
        
        } else if pNtk.ntkFunc != ABC_FUNC_MAP {
            sprintf( pMan.sError, "The network %s appears to have both gates and assign statements. Currently such network are not allowed. One way to fix this problem might be to replace assigns by buffers from the library.", pNtk.pName );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
    
    } else {
        // convert from the blackbox into the network with local functions representated by AIGs
        if pNtk.ntkFunc == ABC_FUNC_BLACKBOX {
            // change network type
            assert(pNtk.pManFunc == nil, "pNtk.pManFunc == nil");
            pNtk.ntkFunc = ABC_FUNC_AIG;
            pNtk.pManFunc = pMan.pDesign.pManFunc;
        
        } else if pNtk.ntkFunc != ABC_FUNC_AIG {
            sprintf( pMan.sError, "The network %s appears to have both gates and assign statements. Currently such network are not allowed. One way to fix this problem might be to replace assigns by buffers from the library.", pNtk.pName );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses one Verilog module.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseModule(pMan *Ver_Man_t) int32 {
    var pGate *Mio_Gate_t;
    var p *Ver_Stream_t  = pMan.pReader;
    var pNtk, pNtkTemp *Abc_Ntk_t;
    var pWord string;
    var Symbol byte;
    var RetValue int32;

    // get the network name
    pWord = Ver_ParseGetName( pMan );

    // get the network with this name
    pNtk = Ver_ParseFindOrCreateNetwork( pMan, pWord );

    // make sure we stopped at the opening parenthesis
    if Ver_StreamPopChar(p) != '(' {
        sprintf( pMan.sError, "Cannot find \"(\" after \"module\" in network %s.", pNtk.pName );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // skip to the end of parentheses
    for {
        if Ver_ParseGetName( pMan ) == nil {
            return 0; }
        Symbol = Ver_StreamPopChar(p);
    
if !( Symbol == ',' ) { break }
} 
    assert(Symbol == ')', "Symbol == ')'");
    if !Ver_ParseSkipComments( pMan ) {
        return 0; }
    Symbol = Ver_StreamPopChar(p);
    if Symbol != ';' {
        sprintf( pMan.sError, "Expected closing parenthesis after \"module\"." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // parse the inputs/outputs/registers/wires/inouts
    for {
        Extra_ProgressBarUpdate( pMan.pProgress, Ver_StreamGetCurPosition(p), nil );
        pWord = Ver_ParseGetName( pMan );
        if pWord == nil {
            return 0; }
        if !(int32(strings.Compare(pWord, "input")) != 0) {
            RetValue = Ver_ParseSignal( pMan, pNtk, VER_SIG_INPUT );
        } else if !(int32(strings.Compare(pWord, "output")) != 0) {
            RetValue = Ver_ParseSignal( pMan, pNtk, VER_SIG_OUTPUT );
        } else if !(int32(strings.Compare(pWord, "reg")) != 0) {
            RetValue = Ver_ParseSignal( pMan, pNtk, VER_SIG_REG );
        } else if !(int32(strings.Compare(pWord, "wire")) != 0) {
            RetValue = Ver_ParseSignal( pMan, pNtk, VER_SIG_WIRE );
        } else if !(int32(strings.Compare(pWord, "inout")) != 0) {
            RetValue = Ver_ParseSignal( pMan, pNtk, VER_SIG_INOUT );
        } else { 
            break; }
        if RetValue == 0 {
            return 0; }
    }

    // parse the remaining statements
    for {
        Extra_ProgressBarUpdate( pMan.pProgress, Ver_StreamGetCurPosition(p), nil );

        if !(int32(strings.Compare(pWord, "and")) != 0) {
            RetValue = Ver_ParseGateStandard( pMan, pNtk, VER_GATE_AND );
        } else if !(int32(strings.Compare(pWord, "or")) != 0) {
            RetValue = Ver_ParseGateStandard( pMan, pNtk, VER_GATE_OR );
        } else if !(int32(strings.Compare(pWord, "xor")) != 0) {
            RetValue = Ver_ParseGateStandard( pMan, pNtk, VER_GATE_XOR );
        } else if !(int32(strings.Compare(pWord, "buf")) != 0) {
            RetValue = Ver_ParseGateStandard( pMan, pNtk, VER_GATE_BUF );
        } else if !(int32(strings.Compare(pWord, "nand")) != 0) {
            RetValue = Ver_ParseGateStandard( pMan, pNtk, VER_GATE_NAND );
        } else if !(int32(strings.Compare(pWord, "nor")) != 0) {
            RetValue = Ver_ParseGateStandard( pMan, pNtk, VER_GATE_NOR );
        } else if !(int32(strings.Compare(pWord, "xnor")) != 0) {
            RetValue = Ver_ParseGateStandard( pMan, pNtk, VER_GATE_XNOR );
        } else if !(int32(strings.Compare(pWord, "not")) != 0) {
            RetValue = Ver_ParseGateStandard( pMan, pNtk, VER_GATE_NOT );

        } else if !(int32(strings.Compare(pWord, "dff")) != 0) {
            RetValue = Ver_ParseFlopStandard( pMan, pNtk );

        } else if !(int32(strings.Compare(pWord, "assign")) != 0) {
            RetValue = Ver_ParseAssign( pMan, pNtk );
        } else if !(int32(strings.Compare(pWord, "always")) != 0) {
            RetValue = Ver_ParseAlways( pMan, pNtk );
        } else if !(int32(strings.Compare(pWord, "initial")) != 0) {
            RetValue = Ver_ParseInitial( pMan, pNtk );
        } else if !(int32(strings.Compare(pWord, "endmodule")) != 0) {
            break;
        } else if pMan.pDesign.pGenlib && (pGate = Mio_LibraryReadGateByName((Mio_Library_t *)pMan.pDesign.pGenlib, pWord, nil)) { // current design
            RetValue = Ver_ParseGate( pMan, pNtk, pGate );
//        else if ( pMan->pDesign->pLibrary && st__lookup(pMan->pDesign->pLibrary->tModules, pWord, (char**)&pNtkTemp) ) // gate library
//            RetValue = Ver_ParseGate( pMan, pNtkTemp );
        } else if !(int32(strings.Compare(pWord, "wire")) != 0) {
            RetValue = Ver_ParseSignal( pMan, pNtk, VER_SIG_WIRE );
        } else { // assume this is the box used in the current design
        
            pNtkTemp = Ver_ParseFindOrCreateNetwork( pMan, pWord );
            RetValue = Ver_ParseBox( pMan, pNtk, pNtkTemp );
        }
        if RetValue == 0 {
            return 0; }
        // skip the comments
        if !Ver_ParseSkipComments( pMan ) {
            return 0; }
        // get new word
        pWord = Ver_ParseGetName( pMan );
        if pWord == nil {
            return 0; }
    }

    // convert from the blackbox into the network with local functions representated by AIGs
    if pNtk.ntkFunc == ABC_FUNC_BLACKBOX {
        if Abc_NtkNodeNum(pNtk) > 0 || Abc_NtkBoxNum(pNtk) > 0 {
            if !(Ver_ParseConvertNetwork( pMan, pNtk, pMan.fMapped ) != 0) {
                return 0; }
        
        } else {
            var pObj, pBox, pTerm *Abc_Obj_t;
            var i int32; 
            pBox = Abc_NtkCreateBlackbox(pNtk);
            // inlining of macro Abc_NtkForEachPi
            for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
                pTerm = Abc_NtkCreateBi(pNtk);
                Abc_ObjAddFanin( pTerm, Abc_ObjFanout0(pObj) );
                Abc_ObjAddFanin( pBox, pTerm );
            }
            // inlining of macro Abc_NtkForEachPo
            for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
                pTerm = Abc_NtkCreateBo(pNtk);
                Abc_ObjAddFanin( pTerm, pBox );
                Abc_ObjAddFanin( Abc_ObjFanin0(pObj), pTerm );
            }
        }
    }

    // remove the table if needed
    Ver_ParseRemoveSuffixTable( pMan );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Lookups the suffix of the signal of the form [m:n].]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseLookupSuffix(pMan *Ver_Man_t, pWord string, pnMsb *int32, pnLsb *int32) int32 {
    var Value uint32;
    *pnMsb = *pnLsb = -1;
    if pMan.tName2Suffix == nil {
        return 1; }
    if ! st__lookup( pMan.tName2Suffix, (char *)pWord, (char **)&Value ) {
        return 1; }
    *pnMsb = (Value >> 8) & 0xff;
    *pnLsb = Value & 0xff;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Lookups the suffix of the signal of the form [m:n].]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseInsertsSuffix(pMan *Ver_Man_t, pWord string, nMsb int32, nLsb int32) int32 {
    var Value uint32;
    if pMan.tName2Suffix == nil {
        pMan.tName2Suffix = st__init_table( st__strcmp, st__strhash ); }
    // inlining of macro st__is_member
    if  st__lookup(pMan.tName2Suffix,pWord,(char **) 0) {
        return 1; }
    assert(nMsb >= 0 && nMsb < 128, "nMsb >= 0 && nMsb < 128");
    assert(nLsb >= 0 && nLsb < 128, "nLsb >= 0 && nLsb < 128");
    Value = (nMsb << 8) | nLsb;
    st__insert( pMan.tName2Suffix, Extra_UtilStrsav(pWord), (char *)(ABC_PTRUINT_T)Value );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Lookups the suffic of the signal of the form [m:n].]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseRemoveSuffixTable(pMan *Ver_Man_t) {
    var gen *st__generator;
    var pKey, pValue string;
    if pMan.tName2Suffix == nil {
        return; }
    // inlining of macro st__foreach_item
    for gen= st__init_gen(pMan.tName2Suffix); st__gen(gen,(char **)&pKey,(char **)&pValue) || ( st__free_gen(gen),0); {
        // inlining of macro ABC_FREE
        ((pKey) ? (free((char *) (pKey)), (pKey) = 0) : 0);
    }
    st__free_table( pMan.tName2Suffix );
    pMan.tName2Suffix = nil;
}

/**Function*************************************************************

  Synopsis    [Determine signal prefix of the form [Beg:End].]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseSignalPrefix(pMan *Ver_Man_t, ppWord *string, pnMsb *int32, pnLsb *int32) int32 {
    var pWord string = *ppWord;
    var pTemp string;
    var nMsb, nLsb int32;
    assert(pWord[0] == '[', "pWord[0] == '['");
    // get the beginning
    nMsb = strconv.Atoi(pWord + 1);
    // find the splitter
    for  *pWord && *pWord != ':' && *pWord != ']'  {
        pWord++;
    }
    if *pWord == 0 {
        sprintf( pMan.sError, "Cannot find closing bracket in this line." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    if *pWord == ']' {
        nLsb = nMsb;
    } else {
        assert(*pWord == ':', "*pWord == ':'");
        nLsb = strconv.Atoi(pWord + 1);
        // find the closing parenthesis
        for  *pWord && *pWord != ']'  {
            pWord++;
        }
        if *pWord == 0 {
            sprintf( pMan.sError, "Cannot find closing bracket in this line." );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        assert(*pWord == ']', "*pWord == ']'");
        pWord++;

        // fix the case when \<name> follows after [] without space
        if *pWord == '\\' {
            pWord++;
            pTemp = pWord;
            for  *pTemp && *pTemp != ' '  {
                pTemp++;
            }
            if *pTemp == ' ' {
                *pTemp = 0; }
        }
    }
    assert(nMsb >= 0 && nLsb >= 0, "nMsb >= 0 && nLsb >= 0");
    // return
    *ppWord = pWord;
    *pnMsb = nMsb;
    *pnLsb = nLsb;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Determine signal suffix of the form [m:n].]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseSignalSuffix(pMan *Ver_Man_t, pWord string, pnMsb *int32, pnLsb *int32) int32 {
    var pCur string;
    var Length int32;
    Length = len(pWord);
    assert(pWord[Length-1] == ']', "pWord[Length-1] == ']'");
    // walk backward
    for  pCur = pWord + Length - 2; pCur != pWord; pCur--  {
        if *pCur == ':' || *pCur == '[' {
            break; }
    }
    if pCur == pWord {
        sprintf( pMan.sError, "Cannot find opening bracket in signal name %s.", pWord );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    if *pCur == '[' {
        *pnMsb = *pnLsb = strconv.Atoi(pCur+1);
        *pCur = 0;
        return 1;
    }
    assert(*pCur == ':', "*pCur == ':'");
    // get the end of the interval
    *pnLsb = strconv.Atoi(pCur+1);
    // find the beginning
    for  pCur = pWord + Length - 2; pCur != pWord; pCur--  {
        if *pCur == '[' {
            break; }
    }
    if pCur == pWord {
        sprintf( pMan.sError, "Cannot find opening bracket in signal name %s.", pWord );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    assert(*pCur == '[', "*pCur == '['");
    // get the beginning of the interval
    *pnMsb = strconv.Atoi(pCur+1);
    // cut the word
    *pCur = 0;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns the values of constant bits.]

  Description [The resulting bits are in MSB to LSB order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseConstant(pMan *Ver_Man_t, pWord string) int32 {
    var nBits, i int32;
    assert(pWord[0] >= '1' && pWord[1] <= '9', "pWord[0] >= '1' && pWord[1] <= '9'");
    nBits = strconv.Atoi(pWord);
    // find the next symbol \'
    for  *pWord && *pWord != '\''  {
        pWord++;
    }
    if *pWord == 0 {
        sprintf( pMan.sError, "Cannot find symbol \' in the constant." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    assert(*pWord == '\'', "*pWord == '\\''");
    pWord++;
    if *pWord != 'b' {
        sprintf( pMan.sError, "Currently can only handle binary constants." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    pWord++;
    // scan the bits
    Vec_PtrClear( pMan.vNames );
    for  i = 0; i < nBits; i++  {
      if pWord[i] != '0' && pWord[i] != '1' && pWord[i] != 'x' {
         sprintf( pMan.sError, "Having problem parsing the binary constant." );
         Ver_ParsePrintErrorMessage( pMan );
         return 0;
      }
      if pWord[i] == 'x' { 
          Vec_PtrPush( pMan.vNames, (void *)0 );
      } else { 
          Vec_PtrPush( pMan.vNames, (void *)(ABC_PTRUINT_T)(pWord[i]-'0') ); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description [The signals are added in the order from LSB to MSB.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseSignal(pMan *Ver_Man_t, pNtk *Abc_Ntk_t, SigType Ver_SignalType_t) int32 {
    var p *Ver_Stream_t  = pMan.pReader;
    var Buffer string // was [1000]byte in C

    var Symbol byte;
    var pWord string;
    var nMsb, nLsb, Bit, Limit, i int32;
    nMsb = nLsb = -1;
    for {
        // get the next word
        pWord = Ver_ParseGetName( pMan );
        if pWord == nil {
            return 0; }

        if !(int32(strings.Compare(pWord, "wire")) != 0) {
            continue; }

        // check if the range is specified
        if pWord[0] == '[' && !pMan.fNameLast {
            assert(nMsb == -1 && nLsb == -1, "nMsb == -1 && nLsb == -1");
            Ver_ParseSignalPrefix( pMan, &pWord, &nMsb, &nLsb );
            // check the case when there is space between bracket and the next word
            if *pWord == 0 {
                // get the signal name
                pWord = Ver_ParseGetName( pMan );
                if pWord == nil {
                    return 0; }
            }
        }

        // create signals
        if nMsb == -1 && nLsb == -1 {
            if SigType == VER_SIG_INPUT || SigType == VER_SIG_INOUT {
                Ver_ParseCreatePi( pNtk, pWord ); }
            if SigType == VER_SIG_OUTPUT || SigType == VER_SIG_INOUT {
                Ver_ParseCreatePo( pNtk, pWord ); }
            if SigType == VER_SIG_WIRE || SigType == VER_SIG_REG {
                Abc_NtkFindOrCreateNet( pNtk, pWord ); }
        
        } else {
            assert(nMsb >= 0 && nLsb >= 0, "nMsb >= 0 && nLsb >= 0");
            // add to the hash table
            Ver_ParseInsertsSuffix( pMan, pWord, nMsb, nLsb );
            // add signals from Lsb to Msb
            if nMsb > nLsb {
                Limit = nMsb - nLsb + 1;
            } else {
                Limit = nLsb - nMsb + 1;
            }
            for  i, Bit = 0, nLsb; i < Limit; i, Bit = i+1, nMsb > nLsb ? Bit + 1: Bit - 1   {
//                sprintf( Buffer, "%s[%d]", pWord, Bit );
                if Limit > 1 {
                    Buffer = fmt.Sprintf("%s[%d]", pWord, Bit);
                } else {
                    Buffer = fmt.Sprintf("%s", pWord); }
                if SigType == VER_SIG_INPUT || SigType == VER_SIG_INOUT {
                    Ver_ParseCreatePi( pNtk, Buffer ); }
                if SigType == VER_SIG_OUTPUT || SigType == VER_SIG_INOUT {
                    Ver_ParseCreatePo( pNtk, Buffer ); }
                if SigType == VER_SIG_WIRE || SigType == VER_SIG_REG {
                    Abc_NtkFindOrCreateNet( pNtk, Buffer ); }
            }
        }

        Symbol = Ver_StreamPopChar(p);
        if Symbol == ',' {
            continue; }
        if Symbol == ';' {
            return 1; }
        break;
    }
    sprintf( pMan.sError, "Cannot parse signal line (expected , or ;)." );
    Ver_ParsePrintErrorMessage( pMan );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseAlways(pMan *Ver_Man_t, pNtk *Abc_Ntk_t) int32 {
    var p *Ver_Stream_t  = pMan.pReader;
    var pNet, pNet2 *Abc_Obj_t;
    var fStopAfterOne int32;
    var pWord, pWord2 string;
    var Symbol byte;
    // parse the directive 
    pWord = Ver_ParseGetName( pMan );
    if pWord == nil {
        return 0; }
    if pWord[0] == '@' {
        Ver_StreamSkipToChars( p, ")" );
        Ver_StreamPopChar(p);
        // parse the directive 
        pWord = Ver_ParseGetName( pMan );
        if pWord == nil {
            return 0; }
    }
    // decide how many statements to parse
    fStopAfterOne = 0;
    if int32(strings.Compare(pWord, "begin")) != 0 {
        fStopAfterOne = 1; }
    // iterate over the initial states
    for {
        if !(fStopAfterOne != 0) {
            // get the name of the output signal
            pWord = Ver_ParseGetName( pMan );
            if pWord == nil {
                return 0; }
            // look for the end of directive
            if !(int32(strings.Compare(pWord, "end")) != 0) {
                break; }
        }
        // get the fanout net
        pNet = Ver_ParseFindNet( pNtk, pWord );
        if pNet == nil {
            sprintf( pMan.sError, "Cannot read the always statement for %s (output wire is not defined).", pWord );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        // get the equality sign
        Symbol = Ver_StreamPopChar(p);
        if Symbol != '<' && Symbol != '=' {
            sprintf( pMan.sError, "Cannot read the assign statement for %s (expected <= or =).", pWord );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        if Symbol == '<' {
            Ver_StreamPopChar(p); }
        // skip the comments
        if !Ver_ParseSkipComments( pMan ) {
            return 0; }
        // get the second name
        pWord2 = Ver_ParseGetName( pMan );
        if pWord2 == nil {
            return 0; }
        // check if the name is complemented
        if pWord2[0] == '~' {
            pNet2 = Ver_ParseFindNet( pNtk, pWord2+1 );
            pNet2 = Ver_ParseCreateInv( pNtk, pNet2 );
        
        } else {
            pNet2 = Ver_ParseFindNet( pNtk, pWord2 ); }
        if pNet2 == nil {
            sprintf( pMan.sError, "Cannot read the always statement for %s (input wire is not defined).", pWord2 );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        // create the latch
        Ver_ParseCreateLatch( pNtk, pNet2, pNet );
        // remove the last symbol
        Symbol = Ver_StreamPopChar(p);
        assert(Symbol == ';', "Symbol == ';'");
        // quit if only one directive
        if fStopAfterOne != 0 {
            break; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseInitial(pMan *Ver_Man_t, pNtk *Abc_Ntk_t) int32 {
    var p *Ver_Stream_t  = pMan.pReader;
    var pNode, pNet *Abc_Obj_t;
    var fStopAfterOne int32;
    var pWord, pEquation string;
    var Symbol byte;
    // parse the directive 
    pWord = Ver_ParseGetName( pMan );
    if pWord == nil {
        return 0; }
    // decide how many statements to parse
    fStopAfterOne = 0;
    if int32(strings.Compare(pWord, "begin")) != 0 {
        fStopAfterOne = 1; }
    // iterate over the initial states
    for {
        if !(fStopAfterOne != 0) {
            // get the name of the output signal
            pWord = Ver_ParseGetName( pMan );
            if pWord == nil {
                return 0; }
            // look for the end of directive
            if !(int32(strings.Compare(pWord, "end")) != 0) {
                break; }
        }
        // get the fanout net
        pNet = Ver_ParseFindNet( pNtk, pWord );
        if pNet == nil {
            sprintf( pMan.sError, "Cannot read the initial statement for %s (output wire is not defined).", pWord );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        // get the equality sign
        Symbol = Ver_StreamPopChar(p);
        if Symbol != '<' && Symbol != '=' {
            sprintf( pMan.sError, "Cannot read the assign statement for %s (expected <= or =).", pWord );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        if Symbol == '<' {
            Ver_StreamPopChar(p); }
        // skip the comments
        if !Ver_ParseSkipComments( pMan ) {
            return 0; }
        // get the second name
        pEquation = Ver_StreamGetWord( p, ";" );
        if pEquation == nil {
            return 0; }
        // find the corresponding latch
        if Abc_ObjFaninNum(pNet) == 0 {
            sprintf( pMan.sError, "Cannot find the latch to assign the initial value." );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        pNode = Abc_ObjFanin0(Abc_ObjFanin0(pNet));
        assert(Abc_ObjIsLatch(pNode), "Abc_ObjIsLatch(pNode)");
        // set the initial state
        if !int32(strings.Compare(pEquation, "0")) || !int32(strings.Compare(pEquation, "1\'b0")) {
            Abc_LatchSetInit0( pNode );
        } else if !int32(strings.Compare(pEquation, "1")) || !int32(strings.Compare(pEquation, "1\'b1")) {
            Abc_LatchSetInit1( pNode );
//        else if ( !strcmp(pEquation, "2") )
//            Abc_LatchSetInitDc( pNode );
        } else {
            sprintf( pMan.sError, "Incorrect initial value of the latch %s.", Abc_ObjName(pNet) );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        // remove the last symbol
        Symbol = Ver_StreamPopChar(p);
        assert(Symbol == ';', "Symbol == ';'");
        // quit if only one directive
        if fStopAfterOne != 0 {
            break; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseAssign(pMan *Ver_Man_t, pNtk *Abc_Ntk_t) int32 {
    var Buffer [1000]byte;
    var Buffer2 [2000]byte;
    var p *Ver_Stream_t  = pMan.pReader;
    var pNode, pNet *Abc_Obj_t;
    var pWord, pName, pEquation string;
    var pFunc *Hop_Obj_t;
    var Symbol byte;
    var i, Bit, Limit, Length, fReduction int32;
    var nMsb, nLsb int32;

//    if ( Ver_StreamGetLineNumber(p) == 2756 )
//    {
//        int x = 0;
//    }

    // convert from the blackbox into the network with local functions representated by AIGs
    if !(Ver_ParseConvertNetwork( pMan, pNtk, pMan.fMapped ) != 0) {
        return 0; }

    for {
        // get the name of the output signal
        pWord = Ver_ParseGetName( pMan );
        if pWord == nil {
            return 0; }
        if int32(strings.Compare(pWord, "#1")) == 0 {
            continue; }
        // check for vector-inputs
        if !(Ver_ParseLookupSuffix( pMan, pWord, &nMsb, &nLsb ) != 0) {
            return 0; }
        // handle special case of constant assignment
        if nMsb > nLsb {
            Limit = nMsb - nLsb + 1;
        } else {
            Limit = nLsb - nMsb + 1;
        }
        if nMsb >= 0 && nLsb >= 0 && Limit > 1 {
            // save the fanout name
            if !int32(strings.Compare(pWord, "1\'h0")) {
                strcpy( Buffer, "1\'b0" );
            } else if !int32(strings.Compare(pWord, "1\'h1")) {
                strcpy( Buffer, "1\'b1" );
            } else {
                strcpy( Buffer, pWord ); }
            // get the equality sign
            if Ver_StreamPopChar(p) != '=' {
                sprintf( pMan.sError, "Cannot read the assign statement for %s (expected equality sign).", pWord );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
            // get the constant
            pWord = Ver_ParseGetName( pMan );
            if pWord == nil {
                return 0; }
            // check if it is indeed a constant
            if !(pWord[0] >= '0' && pWord[0] <= '9') {
                sprintf( pMan.sError, "Currently can only assign vector-signal \"%s\" to be a constant.", Buffer );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }

            // get individual bits of the constant
            if !(Ver_ParseConstant( pMan, pWord ) != 0) {
                return 0; }
            // check that the constant has the same size
            if nMsb > nLsb {
                Limit = nMsb - nLsb + 1;
            } else {
                Limit = nLsb - nMsb + 1;
            }
            if Limit != Vec_PtrSize(pMan.vNames) {
                sprintf( pMan.sError, "The constant size (%d) is different from the signal\"%s\" size (%d).", 
                                    Vec_PtrSize(pMan.vNames), Buffer, Limit );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
            // iterate through the bits
            for  i, Bit = 0, nLsb; i < Limit; i, Bit = i+1, nMsb > nLsb ? Bit + 1: Bit - 1   {
                // get the fanin net
                if Vec_PtrEntry( pMan.vNames, Limit-1-i ) {
                    pNet = Ver_ParseFindNet( pNtk, "1\'b1" );
                } else {
                    pNet = Ver_ParseFindNet( pNtk, "1\'b0" ); }
                assert(pNet != nil, "pNet != nil");

                // create the buffer
                pNode = Abc_NtkCreateNodeBuf( pNtk, pNet );

                // get the fanout net
                sprintf( Buffer2, "%s[%d]", Buffer, Bit );
                pNet = Ver_ParseFindNet( pNtk, Buffer2 );
                if pNet == nil {
                    sprintf( pMan.sError, "Cannot read the assign statement for %s (output wire is not defined).", pWord );
                    Ver_ParsePrintErrorMessage( pMan );
                    return 0;
                }
                Abc_ObjAddFanin( pNet, pNode );
            }
            // go to the end of the line
            Ver_ParseSkipComments( pMan );
        
        } else {
            // consider the case of reduction operations
            fReduction = 0;
            if pWord[0] == '{' && !pMan.fNameLast {
                fReduction = 1; }
            if fReduction != 0 {
                pWord++;
                pWord[len(pWord)-1] = 0;
                assert(pWord[0] != '\\', "pWord[0] != '\\\\'");
            }
            // get the fanout net
            pNet = Ver_ParseFindNet( pNtk, pWord );
            if pNet == nil {
                sprintf( pMan.sError, "Cannot read the assign statement for %s (output wire is not defined).", pWord );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
            // get the equality sign
            if Ver_StreamPopChar(p) != '=' {
                sprintf( pMan.sError, "Cannot read the assign statement for %s (expected equality sign).", pWord );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
            // skip the comments
            if !Ver_ParseSkipComments( pMan ) {
                return 0; }
            // get the second name
            if fReduction != 0 {
                pEquation = Ver_StreamGetWord( p, ";" );
            } else {
                pEquation = Ver_StreamGetWord( p, ",;" ); }
            if pEquation == nil {
                sprintf( pMan.sError, "Cannot read the equation for %s.", Abc_ObjName(pNet) );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }

            // consider the case of mapped network
            Vec_PtrClear( pMan.vNames );
            if pMan.fMapped {
                if !int32(strings.Compare(pEquation, "1\'b0")) {
                    pFunc = (Hop_Obj_t *)Mio_LibraryReadConst0((Mio_Library_t *)Abc_FrameReadLibGen());
                } else if !int32(strings.Compare(pEquation, "1\'b1")) {
                    pFunc = (Hop_Obj_t *)Mio_LibraryReadConst1((Mio_Library_t *)Abc_FrameReadLibGen());
                } else {
                    // "assign foo = \bar ;"
                    if *pEquation == '\\' {
                        pEquation++;
                        pEquation[len(pEquation) - 1] = 0;
                    }
                    if Ver_ParseFindNet(pNtk, pEquation) == nil {
                        sprintf( pMan.sError, "Cannot read Verilog with non-trivial assignments in the mapped netlist." );
                        Ver_ParsePrintErrorMessage( pMan );
                        return 0;
                    }
                    Vec_PtrPush( pMan.vNames, (void *)(ABC_PTRUINT_T)len(pEquation) );
                    Vec_PtrPush( pMan.vNames, pEquation );
                    // get the buffer
                    pFunc = (Hop_Obj_t *)Mio_LibraryReadBuf((Mio_Library_t *)Abc_FrameReadLibGen());
                    if pFunc == nil {
                        sprintf( pMan.sError, "Reading assign statement for node %s has failed because the genlib library has no buffer.", Abc_ObjName(pNet) );
                        Ver_ParsePrintErrorMessage( pMan );
                        return 0;
                    }
                }
            
            } else {
                if !int32(strings.Compare(pEquation, "0")) || !int32(strings.Compare(pEquation, "1\'b0")) || !int32(strings.Compare(pEquation, "1\'bx")) {
                    pFunc = Hop_ManConst0((Hop_Man_t *)pNtk.pManFunc);
                } else if !int32(strings.Compare(pEquation, "1")) || !int32(strings.Compare(pEquation, "1\'b1")) {
                    pFunc = Hop_ManConst1((Hop_Man_t *)pNtk.pManFunc);
                } else if fReduction != 0 {
                    pFunc = (Hop_Obj_t *)Ver_FormulaReduction( pEquation, pNtk.pManFunc, pMan.vNames, pMan.sError );  
                } else {
                    pFunc = (Hop_Obj_t *)Ver_FormulaParser( pEquation, pNtk.pManFunc, pMan.vNames, pMan.vStackFn, pMan.vStackOp, pMan.sError ); }  
                if pFunc == nil {
                    Ver_ParsePrintErrorMessage( pMan );
                    return 0;
                }
            }

            // create the node with the given inputs
            pNode = Abc_NtkCreateNode( pNtk );
            pNode.pData = pFunc;
            Abc_ObjAddFanin( pNet, pNode );
            // connect to fanin nets
            for  i = 0; i < Vec_PtrSize(pMan.vNames)/2; i++  {
                // get the name of this signal
                Length = int32((ABC_PTRUINT_T))Vec_PtrEntry( pMan.vNames, 2*i );
                pName  = (char *)Vec_PtrEntry( pMan.vNames, 2*i + 1 );
                pName[Length] = 0;
                // try name
//                pNet = Ver_ParseFindNet( pNtk, pName );
                if !int32(strings.Compare(pName, "1\'h0")) {
                    pNet = Ver_ParseFindNet( pNtk, "1\'b0" );
                } else if !int32(strings.Compare(pName, "1\'h1")) {
                    pNet = Ver_ParseFindNet( pNtk, "1\'b1" );
                } else {
                    pNet = Ver_ParseFindNet( pNtk, pName ); }
                // find the corresponding net
                if pNet == nil {
                    sprintf( pMan.sError, "Cannot read the assign statement for %s (input wire %s is not defined).", pWord, pName );
                    Ver_ParsePrintErrorMessage( pMan );
                    return 0;
                }
                Abc_ObjAddFanin( pNode, pNet );
            }
        }

        Symbol = Ver_StreamPopChar(p);
        if Symbol == ',' {
            continue; }
        if Symbol == ';' {
            return 1; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseGateStandard(pMan *Ver_Man_t, pNtk *Abc_Ntk_t, GateType Ver_GateType_t) int32 {
    extern void Ver_StreamMove( Ver_Stream_t * p );
    var p *Ver_Stream_t  = pMan.pReader;
    var pNet, pNode *Abc_Obj_t;
    var pWord string;
    var Symbol byte;

    // convert from the blackbox into the network with local functions representated by AIGs
    if !(Ver_ParseConvertNetwork( pMan, pNtk, pMan.fMapped ) != 0) {
        return 0; }
    Ver_StreamMove( p );

    // this is gate name - throw it away
    if Ver_StreamPopChar(p) != '(' {
        sprintf( pMan.sError, "Cannot parse a standard gate (expected opening parenthesis)." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    Ver_ParseSkipComments( pMan );

    // create the node
    pNode = Abc_NtkCreateNode( pNtk );

    // parse pairs of formal/actural inputs
    for {
        // parse the output name
        pWord = Ver_ParseGetName( pMan );
        if pWord == nil {
            return 0; }
        // get the net corresponding to this output
        pNet = Ver_ParseFindNet( pNtk, pWord );
        if pNet == nil {
            sprintf( pMan.sError, "Net is missing in gate %s.", pWord );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        // if this is the first net, add it as an output
        if Abc_ObjFanoutNum(pNode) == 0 {
            Abc_ObjAddFanin( pNet, pNode );
        } else {
            Abc_ObjAddFanin( pNode, pNet ); }
        // check if it is the end of gate
        Ver_ParseSkipComments( pMan );
        Symbol = Ver_StreamPopChar(p);
        if Symbol == ')' {
            break; }
        // skip comma
        if Symbol != ',' {
            sprintf( pMan.sError, "Cannot parse a standard gate %s (expected closing parenthesis).", Abc_ObjName(Abc_ObjFanout0(pNode)) );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        Ver_ParseSkipComments( pMan );
    }
    if (GateType == VER_GATE_BUF || GateType == VER_GATE_NOT) && Abc_ObjFaninNum(pNode) != 1 {
        sprintf( pMan.sError, "Buffer or interver with multiple fanouts %s (currently not supported).", Abc_ObjName(Abc_ObjFanout0(pNode)) );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // check if it is the end of gate
    Ver_ParseSkipComments( pMan );
    if Ver_StreamPopChar(p) != ';' {
        sprintf( pMan.sError, "Cannot read standard gate %s (expected closing semicolumn).", Abc_ObjName(Abc_ObjFanout0(pNode)) );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    // add logic function
    if GateType == VER_GATE_AND || GateType == VER_GATE_NAND {
        pNode.pData = Hop_CreateAnd( (Hop_Man_t *)pNtk.pManFunc, Abc_ObjFaninNum(pNode) );
    } else if GateType == VER_GATE_OR || GateType == VER_GATE_NOR {
        pNode.pData = Hop_CreateOr( (Hop_Man_t *)pNtk.pManFunc, Abc_ObjFaninNum(pNode) );
    } else if GateType == VER_GATE_XOR || GateType == VER_GATE_XNOR {
        pNode.pData = Hop_CreateExor( (Hop_Man_t *)pNtk.pManFunc, Abc_ObjFaninNum(pNode) );
    } else if GateType == VER_GATE_BUF || GateType == VER_GATE_NOT {
        pNode.pData = Hop_CreateAnd( (Hop_Man_t *)pNtk.pManFunc, Abc_ObjFaninNum(pNode) ); }
    if GateType == VER_GATE_NAND || GateType == VER_GATE_NOR || GateType == VER_GATE_XNOR || GateType == VER_GATE_NOT {
        pNode.pData = Hop_Not( (Hop_Obj_t *)pNode.pData ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseFlopStandard(pMan *Ver_Man_t, pNtk *Abc_Ntk_t) int32 {
    var p *Ver_Stream_t  = pMan.pReader;
    var pNetLi, pNetLo, pLatch *Abc_Obj_t;
    var pWord string;
    var Symbol byte;

    // convert from the blackbox into the network with local functions representated by AIGs
    if !(Ver_ParseConvertNetwork( pMan, pNtk, pMan.fMapped ) != 0) {
        return 0; }

    // this is gate name - throw it away
    if Ver_StreamPopChar(p) != '(' {
        sprintf( pMan.sError, "Cannot parse a standard gate (expected opening parenthesis)." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    Ver_ParseSkipComments( pMan );

    // parse the output name
    pWord = Ver_ParseGetName( pMan );
    if pWord == nil {
        return 0; }
    // get the net corresponding to this output
    pNetLo = Ver_ParseFindNet( pNtk, pWord );
    if pNetLo == nil {
        sprintf( pMan.sError, "Net is missing in gate %s.", pWord );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // check if it is the end of gate
    Ver_ParseSkipComments( pMan );
    Symbol = Ver_StreamPopChar(p);
    if Symbol == ')' {
        sprintf( pMan.sError, "Cannot parse the flop." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    // skip comma
    if Symbol != ',' {
        sprintf( pMan.sError, "Cannot parse the flop." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    Ver_ParseSkipComments( pMan );

    // parse the output name
    pWord = Ver_ParseGetName( pMan );
    if pWord == nil {
        return 0; }
    // get the net corresponding to this output
    pNetLi = Ver_ParseFindNet( pNtk, pWord );
    if pNetLi == nil {
        sprintf( pMan.sError, "Net is missing in gate %s.", pWord );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // check if it is the end of gate
    Ver_ParseSkipComments( pMan );
    Symbol = Ver_StreamPopChar(p);
    if Symbol != ')' {
        sprintf( pMan.sError, "Cannot parse the flop." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // check if it is the end of gate
    Ver_ParseSkipComments( pMan );
    if Ver_StreamPopChar(p) != ';' {
        sprintf( pMan.sError, "Cannot parse the flop." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // create the latch
    pLatch = Ver_ParseCreateLatch( pNtk, pNetLi, pNetLo );
    Abc_LatchSetInit0( pLatch );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns the index of the given pin the gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_FindGateInput(pGate *Mio_Gate_t, pName string) int32 {
    var pGatePin *Mio_Pin_t;
    var i int32;
    for  i = 0, pGatePin = Mio_GateReadPins(pGate); pGatePin != nil; pGatePin = Mio_PinReadNext(pGatePin), i++  {
        if int32(strings.Compare(pName, Mio_PinReadName(pGatePin))) == 0 {
            return i; }
    }
    if int32(strings.Compare(pName, Mio_GateReadOutName(pGate))) == 0 {
        return i; }
    if Mio_GateReadTwin(pGate) && int32(strings.Compare(pName, Mio_GateReadOutName(Mio_GateReadTwin(pGate)))) == 0 {
        return i+1; }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseGate(pMan *Ver_Man_t, pNtk *Abc_Ntk_t, pGate *Mio_Gate_t) int32 {
    var p *Ver_Stream_t  = pMan.pReader;
    var pNetActual *Abc_Obj_t;
    var pNode *Abc_Obj_t;
    var pNode2 *Abc_Obj_t = nil;
    var pWord string;
    var Symbol byte;
    var Input int32;
    var i int32;
    var nFanins int32 = Mio_GateReadPinNum(pGate);

    // convert from the blackbox into the network with local functions representated by gates
    if 1 != pMan.fMapped {
        sprintf( pMan.sError, "The network appears to be mapped. Use \"r -m\" to read mapped Verilog." );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // update the network type if needed
    if !(Ver_ParseConvertNetwork( pMan, pNtk, 1 ) != 0) {
        return 0; }

    // parse the directive and set the pointers to the PIs/POs of the gate
    pWord = Ver_ParseGetName( pMan );
    if pWord == nil {
        return 0; }
    // this is gate name - throw it away
    if Ver_StreamPopChar(p) != '(' {
        sprintf( pMan.sError, "Cannot parse gate %s (expected opening parenthesis).", Mio_GateReadName(pGate) );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    Ver_ParseSkipComments( pMan );

    // start the node
    pNode = Abc_NtkCreateNode( pNtk );
    pNode.pData = pGate;
    if Mio_GateReadTwin(pGate) {
        pNode2 = Abc_NtkCreateNode( pNtk );
        pNode2.pData = Mio_GateReadTwin(pGate);
    }
    // parse pairs of formal/actural inputs
    Vec_IntClear( pMan.vPerm );
    for {
        // process one pair of formal/actual parameters
        if Ver_StreamPopChar(p) != '.' {
            sprintf( pMan.sError, "Cannot parse gate %s (expected .).", Mio_GateReadName(pGate) );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }

        // parse the formal name
        pWord = Ver_ParseGetName( pMan );
        if pWord == nil {
            return 0; }

        // find the corresponding pin of the gate
        Input = Ver_FindGateInput( pGate, pWord );
        if Input == -1 {
            sprintf( pMan.sError, "Formal input name %s cannot be found in the gate %s.", pWord, Mio_GateReadOutName(pGate) );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }

        // open the parenthesis
        if Ver_StreamPopChar(p) != '(' {
            sprintf( pMan.sError, "Cannot formal parameter %s of gate %s (expected opening parenthesis).", pWord, Mio_GateReadName(pGate) );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }

        // parse the actual name
        pWord = Ver_ParseGetName( pMan );
        if pWord == nil {
            return 0; }
        // check if the name is complemented
        assert(pWord[0] != '~', "pWord[0] != '~'");
/*
        fCompl = (pWord[0] == '~');
        if ( fCompl )
        {
            fComplUsed = 1;
            pWord++;
            if ( pNtk->pData == NULL )
                pNtk->pData = Extra_MmFlexStart();
        }
*/
        // get the actual net
        pNetActual = Ver_ParseFindNet( pNtk, pWord );
        if pNetActual == nil {
            sprintf( pMan.sError, "Actual net %s is missing.", pWord );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }

        // close the parenthesis
        if Ver_StreamPopChar(p) != ')' {
            sprintf( pMan.sError, "Cannot formal parameter %s of gate %s (expected closing parenthesis).", pWord, Mio_GateReadName(pGate) );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }

        // add the fanin
        if Input < nFanins {
            Vec_IntPush( pMan.vPerm, Input );
            Abc_ObjAddFanin( pNode, pNetActual ); // fanin
            if pNode2 != nil {
                Abc_ObjAddFanin( pNode2, pNetActual ); } // fanin
        
        } else if Input == nFanins {
            Abc_ObjAddFanin( pNetActual, pNode ); // fanout
        } else if Input == nFanins + 1 {
            Abc_ObjAddFanin( pNetActual, pNode2 ); // fanout
        } else {
            assert(false, "0"); }

        // check if it is the end of gate
        Ver_ParseSkipComments( pMan );
        Symbol = Ver_StreamPopChar(p);
        if Symbol == ')' {
            break; }

        // skip comma
        if Symbol != ',' {
            sprintf( pMan.sError, "Cannot formal parameter %s of gate %s (expected closing parenthesis).", pWord, Mio_GateReadName(pGate) );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        Ver_ParseSkipComments( pMan );
    }

    // check that the gate as the same number of input
    if !(Abc_ObjFaninNum(pNode) == nFanins && Abc_ObjFanoutNum(pNode) == 1) {
        sprintf( pMan.sError, "Parsing of gate %s has failed.", Mio_GateReadName(pGate) );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // check if it is the end of gate
    Ver_ParseSkipComments( pMan );
    if Ver_StreamPopChar(p) != ';' {
        sprintf( pMan.sError, "Cannot read gate %s (expected closing semicolumn).", Mio_GateReadName(pGate) );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    
    // check if we need to permute the inputs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(pMan.vPerm)) && (((Input) = Vec_IntEntry(pMan.vPerm, i)), 1); i++  {
        if Input != i {
            break; }
    }
    if i < Vec_IntSize(pMan.vPerm) {
        // add the fanin numnbers to the end of the permuation array
        for  i = 0; i < nFanins; i++  {
            Vec_IntPush( pMan.vPerm, Abc_ObjFaninId(pNode, i) );
        }
        // write the fanin numbers into their corresponding places (according to the gate) 
        for  i = 0; i < nFanins; i++  {
            Vec_IntWriteEntry( &pNode.vFanins, Vec_IntEntry(pMan.vPerm, i), Vec_IntEntry(pMan.vPerm, i+nFanins) );
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseBox(pMan *Ver_Man_t, pNtk *Abc_Ntk_t, pNtkBox *Abc_Ntk_t) int32 {
    var Buffer string // was [1000]byte in C

    var p *Ver_Stream_t  = pMan.pReader;
    var pBundle *Ver_Bundle_t;
    var vBundles *Vec_Ptr_t;
    var pNetActual *Abc_Obj_t; 
    var pNode *Abc_Obj_t;
    var pWord string;
    var Symbol byte;
    var fCompl, fFormalIsGiven int32;
    var i, k, Bit, Limit, nMsb, nLsb, fQuit, flag int32;

    // gate the name of the box
    pWord = Ver_ParseGetName( pMan );
    if pWord == nil {
        return 0; }

    // create a box with this name
    pNode = Abc_NtkCreateBlackbox( pNtk );
    pNode.pData = pNtkBox;
    Abc_ObjAssignName( pNode, pWord, nil );

    // continue parsing the box
    if Ver_StreamPopChar(p) != '(' {
        sprintf( pMan.sError, "Cannot parse box %s (expected opening parenthesis).", Abc_ObjName(pNode) );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }
    Ver_ParseSkipComments( pMan );
 
    // parse pairs of formal/actual inputs
    vBundles = Vec_PtrAlloc( 16 );
    pNode.pCopy = (Abc_Obj_t *)vBundles;
    for {
        // allocate the bundle (formal name + array of actual nets)
        // inlining of macro ABC_ALLOC
        pBundle = ((Ver_Bundle_t *) malloc(sizeof(Ver_Bundle_t) * (size_t)(1)));
        pBundle.pNameFormal = nil;
        pBundle.vNetsActual = Vec_PtrAlloc( 4 );
        Vec_PtrPush( vBundles, pBundle );

        // process one pair of formal/actual parameters
        fFormalIsGiven = 0;
        if Ver_StreamScanChar(p) == '.' {
            fFormalIsGiven = 1;
            if Ver_StreamPopChar(p) != '.' {
                sprintf( pMan.sError, "Cannot parse box %s (expected .).", Abc_ObjName(pNode) );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }

            // parse the formal name
            pWord = Ver_ParseGetName( pMan );
            if pWord == nil {
                return 0; }

            // save the name
            pBundle.pNameFormal = Extra_UtilStrsav( pWord );

            // open the parenthesis
            if Ver_StreamPopChar(p) != '(' {
                sprintf( pMan.sError, "Cannot formal parameter %s of box %s (expected opening parenthesis).", pWord, Abc_ObjName(pNode));
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
            Ver_ParseSkipComments( pMan );
        }

        // check if this is the beginning of {} expression
        Symbol = Ver_StreamScanChar(p);

        // consider the case of vector-inputs
        if Symbol == '{' {
            // skip this char
            Ver_StreamPopChar(p);

            // read actual names
            i = 0;
            fQuit = 0;
            for {
                // parse the formal name
                Ver_ParseSkipComments( pMan );
                pWord = Ver_ParseGetName( pMan );
                if pWord == nil {
                    return 0; }

                // check if the last char is a closing brace
                if pWord[len(pWord)-1] == '}' {
                    pWord[len(pWord)-1] = 0;
                    fQuit = 1;
                }
                if pWord[0] == 0 {
                    break; }

                // check for constant
                if pWord[0] >= '1' && pWord[0] <= '9' {
                    if !(Ver_ParseConstant( pMan, pWord ) != 0) {
                        return 0; }
                    // add constant MSB to LSB
                    for  k = 0; k < Vec_PtrSize(pMan.vNames); k, i = k+1, i+1  {
                        // get the actual net
                        Buffer = fmt.Sprintf("1\'b%d", int32((Vec_PtrEntry(pMan.vNames,k) != nil)));
                        pNetActual = Ver_ParseFindNet( pNtk, Buffer );
                        if pNetActual == nil {
                            sprintf( pMan.sError, "Actual net \"%s\" is missing in gate \"%s\".", Buffer, Abc_ObjName(pNode) );
                            Ver_ParsePrintErrorMessage( pMan );
                            return 0;
                        }
                        Vec_PtrPush( pBundle.vNetsActual, pNetActual );
                    }
                
                } else {
                    // get the suffix of the form [m:n]
                    if pWord[len(pWord)-1] == ']' && !pMan.fNameLast {
                        Ver_ParseSignalSuffix( pMan, pWord, &nMsb, &nLsb );
                    } else {
                        Ver_ParseLookupSuffix( pMan, pWord, &nMsb, &nLsb ); }

                    // generate signals
                    if nMsb == -1 && nLsb == -1 {
                        // get the actual net
                        pNetActual = Ver_ParseFindNet( pNtk, pWord );
                        if pNetActual == nil {
                            if !strncmp(pWord, "Open_", 5) ||
                                !strncmp(pWord, "dct_unconnected", 15) { 
                                pNetActual = Abc_NtkCreateNet( pNtk );
                            } else {
                                sprintf( pMan.sError, "Actual net \"%s\" is missing in box \"%s\".", pWord, Abc_ObjName(pNode) );
                                Ver_ParsePrintErrorMessage( pMan );
                                return 0;
                            }
                        }
                        Vec_PtrPush( pBundle.vNetsActual, pNetActual );
                        i++;
                    
                    } else {
                        // go from MSB to LSB
                        assert(nMsb >= 0 && nLsb >= 0, "nMsb >= 0 && nLsb >= 0");
                        if nMsb > nLsb {
                            Limit = nMsb - nLsb + 1;
                        } else {
                            Limit = nLsb - nMsb + 1;
                        }  
                        for  Bit, k = nMsb, Limit - 1; k >= 0; Bit = (nMsb > nLsb ? Bit - 1: Bit + 1), k--, i++  {
                            // get the actual net
                            Buffer = fmt.Sprintf("%s[%d]", pWord, Bit);
                            pNetActual = Ver_ParseFindNet( pNtk, Buffer );
                            if pNetActual == nil {
                                if !strncmp(pWord, "Open_", 5) ||
                                    !strncmp(pWord, "dct_unconnected", 15) { 
                                    pNetActual = Abc_NtkCreateNet( pNtk );
                                } else {
                                    sprintf( pMan.sError, "Actual net \"%s\" is missing in box \"%s\".", pWord, Abc_ObjName(pNode) );
                                    Ver_ParsePrintErrorMessage( pMan );
                                    return 0;
                                }
                            }
                            Vec_PtrPush( pBundle.vNetsActual, pNetActual );
                        }
                    }
                }

                if fQuit {
                    break; }

                // skip comma
                Ver_ParseSkipComments( pMan );
                Symbol = Ver_StreamPopChar(p);
                if Symbol == '}' {
                    break; }
                if Symbol != ',' {
                    sprintf( pMan.sError, "Cannot parse formal parameter %s of gate %s (expected comma).", pWord, Abc_ObjName(pNode) );
                    Ver_ParsePrintErrorMessage( pMan );
                    return 0;
                }
            }
        
        } else {
            // get the next word
            pWord = Ver_ParseGetName( pMan );
            if pWord == nil {
                return 0; }
            // consider the case of empty name
            fCompl = 0;
            if pWord[0] == 0 {
                pNetActual = Abc_NtkCreateNet( pNtk );
                Vec_PtrPush( pBundle.vNetsActual, Abc_ObjNotCond( pNetActual, fCompl ) );
            
            } else {
                // get the actual net
                flag=0;
                pNetActual = Ver_ParseFindNet( pNtk, pWord );
                if pNetActual == nil {
                    Ver_ParseLookupSuffix( pMan, pWord, &nMsb, &nLsb );
                    if nMsb == -1 && nLsb == -1 {
                        Ver_ParseSignalSuffix( pMan, pWord, &nMsb, &nLsb );
                        if nMsb == -1 && nLsb == -1 {
                            if !strncmp(pWord, "Open_", 5) ||
                                !strncmp(pWord, "dct_unconnected", 15) {
                                pNetActual = Abc_NtkCreateNet( pNtk );
                                Vec_PtrPush( pBundle.vNetsActual, pNetActual );
                             
                            } else {
                                sprintf( pMan.sError, "Actual net \"%s\" is missing in box \"%s\".", pWord, Abc_ObjName(pNode) );
                                Ver_ParsePrintErrorMessage( pMan );
                                return 0;
                            }
                         
                        } else {
                            flag=1;
                        }
                     
                    } else {
                        flag=1;
                    }
                    if flag != 0 {
                        if nMsb > nLsb {
                            Limit = nMsb - nLsb + 1;
                        } else {
                            Limit = nLsb - nMsb + 1;
                        }  
                        for  Bit, k = nMsb, Limit - 1; k >= 0; Bit = (nMsb > nLsb ? Bit - 1: Bit + 1), k-- {
                            // get the actual net
                            Buffer = fmt.Sprintf("%s[%d]", pWord, Bit);
                            pNetActual = Ver_ParseFindNet( pNtk, Buffer );
                            if pNetActual == nil {
                                if !strncmp(pWord, "Open_", 5) ||
                                    !strncmp(pWord, "dct_unconnected", 15) { 
                                    pNetActual = Abc_NtkCreateNet( pNtk );
                                } else {
                                    sprintf( pMan.sError, "Actual net \"%s\" is missing in box \"%s\".", pWord, Abc_ObjName(pNode) );
                                    Ver_ParsePrintErrorMessage( pMan );
                                    return 0;
                                }
                            }
                            Vec_PtrPush( pBundle.vNetsActual, pNetActual );
                        }
                    }
                 
                } else {
                    Vec_PtrPush( pBundle.vNetsActual, Abc_ObjNotCond( pNetActual, fCompl ) );
                }
            }
        }

        if fFormalIsGiven != 0 {
            // close the parenthesis
            Ver_ParseSkipComments( pMan );
            if Ver_StreamPopChar(p) != ')' {
                sprintf( pMan.sError, "Cannot parse formal parameter %s of box %s (expected closing parenthesis).", pWord, Abc_ObjName(pNode) );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
            Ver_ParseSkipComments( pMan );
        }

        // check if it is the end of gate
        Symbol = Ver_StreamPopChar(p);
        if Symbol == ')' {
            break; }
        // skip comma
        if Symbol != ',' {
            sprintf( pMan.sError, "Cannot parse formal parameter %s of box %s (expected comma).", pWord, Abc_ObjName(pNode) );
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        Ver_ParseSkipComments( pMan );
    }

    // check if it is the end of gate
    Ver_ParseSkipComments( pMan );
    if Ver_StreamPopChar(p) != ';' {
        sprintf( pMan.sError, "Cannot read box %s (expected closing semicolumn).", Abc_ObjName(pNode) );
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    return 1;
}

/**Function*************************************************************

  Synopsis    [Connects one box to the network]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseFreeBundle(pBundle *Ver_Bundle_t) {
    // inlining of macro ABC_FREE
    ((pBundle.pNameFormal) ? (free((char *) (pBundle.pNameFormal)), (pBundle.pNameFormal) = 0) : 0);
    Vec_PtrFree( pBundle.vNetsActual );
    // inlining of macro ABC_FREE
    ((pBundle) ? (free((char *) (pBundle)), (pBundle) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Connects one box to the network]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseConnectBox(pMan *Ver_Man_t, pBox *Abc_Obj_t) int32 {
    var vBundles *Vec_Ptr_t  = (Vec_Ptr_t *)pBox.pCopy;
    var pNtk *Abc_Ntk_t  = pBox.pNtk;
    var pNtkBox *Abc_Ntk_t  = (Abc_Ntk_t *)pBox.pData;
    var pTerm, pTermNew, pNetAct *Abc_Obj_t;
    var pBundle *Ver_Bundle_t;
    var pNameFormal string;
    var i, k, j, iBundle, Length int32;

    assert(!(Ver_ObjIsConnected(pBox) != 0), "!Ver_ObjIsConnected(pBox)");
    assert((Ver_NtkIsDefined(pNtkBox) != 0), "Ver_NtkIsDefined(pNtkBox)");
    assert(!Abc_NtkHasBlackbox(pNtkBox) || Abc_NtkBoxNum(pNtkBox) == 1, "!Abc_NtkHasBlackbox(pNtkBox) || Abc_NtkBoxNum(pNtkBox) == 1");

/*
    // clean the PI/PO nets
    Abc_NtkForEachPi( pNtkBox, pTerm, i )
        Abc_ObjFanout0(pTerm)->pCopy = NULL;
    Abc_NtkForEachPo( pNtkBox, pTerm, i )
        Abc_ObjFanin0(pTerm)->pCopy = NULL;
*/

    // check the number of actual nets is the same as the number of formal nets
    if Vec_PtrSize(vBundles) > Abc_NtkPiNum(pNtkBox) + Abc_NtkPoNum(pNtkBox) {
        sprintf( pMan.sError, "The number of actual IOs (%d) is bigger than the number of formal IOs (%d) when instantiating network %s in box %s.", 
                    Vec_PtrSize(vBundles), Abc_NtkPiNum(pNtkBox) + Abc_NtkPoNum(pNtkBox), pNtkBox.pName, Abc_ObjName(pBox) );
        // free the bundling
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, k)), 1); k++  {
            Ver_ParseFreeBundle( pBundle );
        }
        Vec_PtrFree( vBundles );
        pBox.pCopy = nil;
        Ver_ParsePrintErrorMessage( pMan );
        return 0;
    }

    // check if some of them do not have formal names
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, k)), 1); k++  {
        if pBundle.pNameFormal == nil {
            break; }
    }
    if k < Vec_PtrSize(vBundles) {
        fmt.Printf("Warning: The instance %s of network %s will be connected without using formal names.\n", pNtkBox.pName, Abc_ObjName(pBox));
        // add all actual nets in the bundles
        iBundle = 0;
        // inlining of macro Vec_PtrForEachEntry
        for  j = 0; (j < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, j)), 1); j++  {
            iBundle += Vec_PtrSize(pBundle.vNetsActual);
        }

        // check the number of actual nets is the same as the number of formal nets
        if iBundle != Abc_NtkPiNum(pNtkBox) + Abc_NtkPoNum(pNtkBox) {
            sprintf( pMan.sError, "The number of actual IOs (%d) is different from the number of formal IOs (%d) when instantiating network %s in box %s.", 
                            Vec_PtrSize(vBundles), Abc_NtkPiNum(pNtkBox) + Abc_NtkPoNum(pNtkBox), pNtkBox.pName, Abc_ObjName(pBox) );
            // free the bundling
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, k)), 1); k++  {
                Ver_ParseFreeBundle( pBundle );
            }
            Vec_PtrFree( vBundles );
            pBox.pCopy = nil;
            Ver_ParsePrintErrorMessage( pMan );
            return 0;
        }
        // connect bundles in the natural order
        iBundle = 0;
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtkBox)) && (((pTerm) = Abc_NtkPi(pNtkBox, i)), 1); i++  {
            pBundle = (Ver_Bundle_t *)Vec_PtrEntry( vBundles, iBundle++ );
            // the bundle is found - add the connections - using order LSB to MSB
            // inlining of macro Vec_PtrForEachEntryReverse
            for  k = Vec_PtrSize(pBundle.vNetsActual) - 1; (k >= 0) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, k)), 1); k--  {
                pTermNew = Abc_NtkCreateBi( pNtk );
                Abc_ObjAddFanin( pBox, pTermNew );
                Abc_ObjAddFanin( pTermNew, pNetAct );
                i++;
            }
            i--;
        }
        // create fanins of the box
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtkBox)) && (((pTerm) = Abc_NtkPo(pNtkBox, i)), 1); i++  {
            pBundle = (Ver_Bundle_t *)Vec_PtrEntry( vBundles, iBundle++ );
            // the bundle is found - add the connections - using order LSB to MSB
            // inlining of macro Vec_PtrForEachEntryReverse
            for  k = Vec_PtrSize(pBundle.vNetsActual) - 1; (k >= 0) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, k)), 1); k--  {
                pTermNew = Abc_NtkCreateBo( pNtk );
                Abc_ObjAddFanin( pTermNew, pBox );
                Abc_ObjAddFanin( pNetAct, pTermNew );
                i++;
            }
            i--;
        }

        // free the bundling
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, k)), 1); k++  {
            Ver_ParseFreeBundle( pBundle );
        }
        Vec_PtrFree( vBundles );
        pBox.pCopy = nil;
        return 1;
    }

    // bundles arrive in any order - but inside each bundle the order is MSB to LSB
    // make sure every formal PI has a corresponding net
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtkBox)) && (((pTerm) = Abc_NtkPi(pNtkBox, i)), 1); i++  {
        // get the name of this formal net
        pNameFormal = Abc_ObjName( Abc_ObjFanout0(pTerm) );
        // try to find the bundle with this formal net
        pBundle = nil;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, k)), 1); k++  {
            if !(int32(strings.Compare(pBundle.pNameFormal, pNameFormal)) != 0) {
                break; }
        }
        assert(pBundle != nil, "pBundle != nil");
        // if the bundle is not found, try without parentheses
        if k == Vec_PtrSize(vBundles) {
            pBundle = nil;
            Length = len(pNameFormal);
            if pNameFormal[Length-1] == ']' {
                // find the opening brace
                for  Length--; Length >= 0; Length--  {
                    if pNameFormal[Length] == '[' {
                        break; }
                }
                // compare names before brace
                if Length > 0 {
                    // inlining of macro Vec_PtrForEachEntry
                    for  j = 0; (j < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, j)), 1); j++  {
                        if !strncmp(pBundle.pNameFormal, pNameFormal, Length) && int32(len(pBundle.pNameFormal)) == Length {
                            break; }
                    }
                    if j == Vec_PtrSize(vBundles) {
                        pBundle = nil; }
                }
            }
            if pBundle == nil {
                sprintf( pMan.sError, "Cannot find an actual net for the formal net %s when instantiating network %s in box %s.", 
                                    pNameFormal, pNtkBox.pName, Abc_ObjName(pBox) );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
        }
        // the bundle is found - add the connections - using order LSB to MSB
        // inlining of macro Vec_PtrForEachEntryReverse
        for  k = Vec_PtrSize(pBundle.vNetsActual) - 1; (k >= 0) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, k)), 1); k--  {
            pTermNew = Abc_NtkCreateBi( pNtk );
            Abc_ObjAddFanin( pBox, pTermNew );
            Abc_ObjAddFanin( pTermNew, pNetAct );
            i++;
        }
        i--;
    }

    // connect those formal POs that do have nets
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtkBox)) && (((pTerm) = Abc_NtkPo(pNtkBox, i)), 1); i++  {
        // get the name of this PI
        pNameFormal = Abc_ObjName( Abc_ObjFanin0(pTerm) );
        // try to find this formal net in the bundle
        pBundle = nil;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, k)), 1); k++  {
            if !(int32(strings.Compare(pBundle.pNameFormal, pNameFormal)) != 0) {
                break; }
        }
        assert(pBundle != nil, "pBundle != nil");
        // if the name is not found, try without parentheses
        if k == Vec_PtrSize(vBundles) {
            pBundle = nil;
            Length = len(pNameFormal);
            if pNameFormal[Length-1] == ']' {
                // find the opening brace
                for  Length--; Length >= 0; Length--  {
                    if pNameFormal[Length] == '[' {
                        break; }
                }
                // compare names before brace
                if Length > 0 {
                    // inlining of macro Vec_PtrForEachEntry
                    for  j = 0; (j < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, j)), 1); j++  {
                        if !strncmp(pBundle.pNameFormal, pNameFormal, Length) && int32(len(pBundle.pNameFormal)) == Length { 
                            break; }
                    }
                    if j == Vec_PtrSize(vBundles) {
                        pBundle = nil; }
                }
            }
            if pBundle == nil {
                var Buffer string // was [1000]byte in C

//                printf( "Warning: The formal output %s is not driven when instantiating network %s in box %s.", 
//                    pNameFormal, pNtkBox->pName, Abc_ObjName(pBox) );
                pTermNew = Abc_NtkCreateBo( pNtk );
                Buffer = fmt.Sprintf("_temp_net%d", Abc_ObjId(pTermNew));
                pNetAct = Abc_NtkFindOrCreateNet( pNtk, Buffer );
                Abc_ObjAddFanin( pTermNew, pBox );
                Abc_ObjAddFanin( pNetAct, pTermNew );
                continue;
            }
        }
        // the bundle is found - add the connections
        // inlining of macro Vec_PtrForEachEntryReverse
        for  k = Vec_PtrSize(pBundle.vNetsActual) - 1; (k >= 0) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, k)), 1); k--  {
            if !int32(strings.Compare(Abc_ObjName(pNetAct), "1\'b0")) || !int32(strings.Compare(Abc_ObjName(pNetAct), "1\'b1")) {
                sprintf( pMan.sError, "It looks like formal output %s is driving a constant net (%s) when instantiating network %s in box %s.", 
                                    pBundle.pNameFormal, Abc_ObjName(pNetAct), pNtkBox.pName, Abc_ObjName(pBox) );
                // free the bundling
                // inlining of macro Vec_PtrForEachEntry
                for  k = 0; (k < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, k)), 1); k++  {
                    Ver_ParseFreeBundle( pBundle );
                }
                Vec_PtrFree( vBundles );
                pBox.pCopy = nil;
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
            pTermNew = Abc_NtkCreateBo( pNtk );
            Abc_ObjAddFanin( pTermNew, pBox );
            Abc_ObjAddFanin( pNetAct, pTermNew );
            i++;
        }
        i--;
    }

    // free the bundling
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, k)), 1); k++  {
        Ver_ParseFreeBundle( pBundle );
    }
    Vec_PtrFree( vBundles );
    pBox.pCopy = nil;
    return 1;
}


/**Function*************************************************************

  Synopsis    [Connects the defined boxes.]

  Description [Returns 2 if there are any undef boxes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseConnectDefBoxes(pMan *Ver_Man_t) int32 {
    var pNtk *Abc_Ntk_t;
    var pBox *Abc_Obj_t;
    var i int32;
    var k int32;
    var RetValue int32 = 1;
    // go through all the modules
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        // go through all the boxes of this module
        // inlining of macro Abc_NtkForEachBox
        for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, k)), 1); k++  {
            if Abc_ObjIsLatch(pBox) {
                continue; }
            // skip internal boxes of the blackboxes
            if pBox.pData == nil {
                continue; }
            // if the network is undefined, it will be connected later
            if !Ver_NtkIsDefined((Abc_Ntk_t *)pBox.pData) {
                RetValue = 2;
                continue;
            }
            // connect the box
            if !(Ver_ParseConnectBox( pMan, pBox ) != 0) {
                return 0; }
            // if the network is a true blackbox, skip
            if Abc_NtkHasBlackbox((Abc_Ntk_t *)pBox.pData) {
                continue; }
            // convert the box to the whitebox
            Abc_ObjBlackboxToWhitebox( pBox );
        }
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Collects the undef boxes and maps them into their instances.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseCollectUndefBoxes(pMan *Ver_Man_t) *Vec_Ptr_t {
    var vUndefs *Vec_Ptr_t;
    var pNtk, pNtkBox *Abc_Ntk_t;
    var pBox *Abc_Obj_t;
    var i, k int32;
    // clear the module structures
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        pNtk.pData = nil;
    }
    // go through all the blackboxes
    vUndefs = Vec_PtrAlloc( 16 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachBlackbox
        for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, k)), 1); k++  {           if !Abc_ObjIsBlackbox(pBox) {} else {
            pNtkBox = (Abc_Ntk_t *)pBox.pData;
            if pNtkBox == nil {
                continue; }
            if Ver_NtkIsDefined(pNtkBox) != 0 {
                continue; }
            if pNtkBox.pData == nil {
                // save the box
                Vec_PtrPush( vUndefs, pNtkBox );
                pNtkBox.pData = Vec_PtrAlloc( 16 );
            }
            // save the instance
            Vec_PtrPush( (Vec_Ptr_t *)pNtkBox.pData, pBox );
        }
        }
    }
    return vUndefs;
}

/**Function*************************************************************

  Synopsis    [Reports how many times each type of undefined box occurs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseReportUndefBoxes(pMan *Ver_Man_t) {
    var pNtk *Abc_Ntk_t;
    var pBox *Abc_Obj_t;
    var i, k, nBoxes int32;
    // clean 
    nBoxes = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        pNtk.fHiePath = 0;
        if !(Ver_NtkIsDefined(pNtk) != 0) {
            nBoxes++; }
    }
    // count
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachBlackbox
        for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, k)), 1); k++  {           if !Abc_ObjIsBlackbox(pBox) {} else
            if pBox.pData && !Ver_NtkIsDefined((Abc_Ntk_t *)pBox.pData) {
                ((Abc_Ntk_t *)pBox.pData).fHiePath++; }
        }
    }
    // print the stats
    fmt.Printf("Warning: The design contains %d undefined object types interpreted as blackboxes:\n", nBoxes);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        if !(Ver_NtkIsDefined(pNtk) != 0) {
            fmt.Printf("%s (%d)  ", Abc_NtkName(pNtk), pNtk.fHiePath); }
    }
    fmt.Printf("\n");
    // clean 
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        pNtk.fHiePath = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if there are non-driven nets.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseCheckNondrivenNets(vUndefs *Vec_Ptr_t) int32 {
    var pNtk *Abc_Ntk_t;
    var pBundle *Ver_Bundle_t;
    var pBox, pNet *Abc_Obj_t;
    var i, k, j, m int32;
    // go through undef box types
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vUndefs)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(vUndefs, i)), 1); i++  {
        // go through instances of this type
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize((Vec_Ptr_t *)pNtk.pData)) && (((pBox) = (Abc_Obj_t *)Vec_PtrEntry((Vec_Ptr_t *)pNtk.pData, k)), 1); k++  {
            // go through the bundles of this instance
            // inlining of macro Vec_PtrForEachEntryReverse
            for  j = Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy) - 1; (j >= 0) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry((Vec_Ptr_t *)pBox.pCopy, j)), 1); j--  {
                // go through the actual nets of this bundle
                if pBundle != nil {
                // inlining of macro Vec_PtrForEachEntry
                for  m = 0; (m < Vec_PtrSize(pBundle.vNetsActual)) && (((pNet) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, m)), 1); m++  {
                    if Abc_ObjFaninNum(pNet) == 0 { // non-driven
                        if int32(strings.Compare(Abc_ObjName(pNet), "1\'b0")) && int32(strings.Compare(Abc_ObjName(pNet), "1\'b1")) { // diff from a const
                            return 1; } }
                } }
            }
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Checks if formal nets with the given name are driven in any of the instances of undef boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseFormalNetsAreDriven(pNtk *Abc_Ntk_t, pNameFormal string) int32 {
    var pBundle *Ver_Bundle_t  = nil;
    var pBox, pNet *Abc_Obj_t;
    var k, j, m int32;
    // go through instances of this type
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize((Vec_Ptr_t *)pNtk.pData)) && (((pBox) = (Abc_Obj_t *)Vec_PtrEntry((Vec_Ptr_t *)pNtk.pData, k)), 1); k++  {
        // find a bundle with the given name in this instance
        // inlining of macro Vec_PtrForEachEntryReverse
        for  j = Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy) - 1; (j >= 0) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry((Vec_Ptr_t *)pBox.pCopy, j)), 1); j--  {
            if (pBundle != nil) && !(int32(strings.Compare(pBundle.pNameFormal, pNameFormal)) != 0) {
                break; }
        }
        // skip non-driven bundles
        if j == Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy) {
            continue; }
        // check if all nets are driven in this bundle
        assert((pBundle != nil), "pBundle"); // Verify that pBundle was assigned to.
        // inlining of macro Vec_PtrForEachEntry
        for  m = 0; (m < Vec_PtrSize(pBundle.vNetsActual)) && (((pNet) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, m)), 1); m++  {
            if Abc_ObjFaninNum(pNet) > 0 {
                return 1; }
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the non-driven bundle that is given distance from the end.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseGetNondrivenBundle(pNtk *Abc_Ntk_t, Counter int32) *Ver_Bundle_t {
    var pBundle *Ver_Bundle_t;
    var pBox, pNet *Abc_Obj_t;
    var k, m int32;
    // go through instances of this type
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize((Vec_Ptr_t *)pNtk.pData)) && (((pBox) = (Abc_Obj_t *)Vec_PtrEntry((Vec_Ptr_t *)pNtk.pData, k)), 1); k++  {
        if Counter >= Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy) {
            continue; }
        // get the bundle given distance away
        pBundle = (Ver_Bundle_t *)Vec_PtrEntry( (Vec_Ptr_t *)pBox.pCopy, Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy) - 1 - Counter );
        if pBundle == nil {
            continue; }
        // go through the actual nets of this bundle
        // inlining of macro Vec_PtrForEachEntry
        for  m = 0; (m < Vec_PtrSize(pBundle.vNetsActual)) && (((pNet) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, m)), 1); m++  {
            if !Abc_ObjFaninNum(pNet) && !(Ver_ParseFormalNetsAreDriven(pNtk, pBundle.pNameFormal) != 0) { // non-driven
                return pBundle; }
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Drives the bundle in the given undef box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseDriveFormal(pMan *Ver_Man_t, pNtk *Abc_Ntk_t, pBundle0 *Ver_Bundle_t) int32 {
    var Buffer string // was [200]byte in C

    var pName string;
    var pBundle *Ver_Bundle_t  = nil;
    var pBox, pTerm, pTermNew, pNetAct, pNetFormal *Abc_Obj_t;
    var k, j, m int32;

    // drive this net in the undef box
    // inlining of macro Vec_PtrForEachEntry
    for  m = 0; (m < Vec_PtrSize(pBundle0.vNetsActual)) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle0.vNetsActual, m)), 1); m++  {
        // create the formal net
        if Vec_PtrSize(pBundle0.vNetsActual) == 1 {
            Buffer = fmt.Sprintf("%s", pBundle0.pNameFormal);
        } else {
            Buffer = fmt.Sprintf("%s[%d]", pBundle0.pNameFormal, m); }
        assert(Abc_NtkFindNet( pNtk, Buffer ) == nil, "Abc_NtkFindNet( pNtk, Buffer ) == nil");
        pNetFormal = Abc_NtkFindOrCreateNet( pNtk, Buffer );
        // connect it to the box
        pTerm = Abc_NtkCreateBo( pNtk );
        assert(Abc_NtkBoxNum(pNtk) <= 1, "Abc_NtkBoxNum(pNtk) <= 1");
        if Abc_NtkBoxNum(pNtk) {
            pBox = Abc_NtkBox(pNtk,0);
        } else {
            pBox = Abc_NtkCreateBlackbox(pNtk);
        }
        Abc_ObjAddFanin( Abc_NtkCreatePo(pNtk), pNetFormal );
        Abc_ObjAddFanin( pNetFormal, pTerm );
        Abc_ObjAddFanin( pTerm, pBox );
    }

    // go through instances of this type
    pName = Extra_UtilStrsav(pBundle0.pNameFormal);
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize((Vec_Ptr_t *)pNtk.pData)) && (((pBox) = (Abc_Obj_t *)Vec_PtrEntry((Vec_Ptr_t *)pNtk.pData, k)), 1); k++  {
        // find a bundle with the given name in this instance
        // inlining of macro Vec_PtrForEachEntryReverse
        for  j = Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy) - 1; (j >= 0) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry((Vec_Ptr_t *)pBox.pCopy, j)), 1); j--  {
            if (pBundle != nil) && !(int32(strings.Compare(pBundle.pNameFormal, pName)) != 0) {
                break; }
        }
        // skip non-driven bundles
        if j == Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy) {
            continue; }
        // check if any nets are driven in this bundle
        assert((pBundle != nil), "pBundle"); // Verify pBundle was assigned to.
        // inlining of macro Vec_PtrForEachEntry
        for  m = 0; (m < Vec_PtrSize(pBundle.vNetsActual)) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, m)), 1); m++  {
            if Abc_ObjFaninNum(pNetAct) > 0 {
                sprintf( pMan.sError, "Missing specification of the I/Os of undefined box \"%s\".", Abc_NtkName(pNtk) );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
        }
        // drive the nets by the undef box
        // inlining of macro Vec_PtrForEachEntryReverse
        for  m = Vec_PtrSize(pBundle.vNetsActual) - 1; (m >= 0) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, m)), 1); m--  {
            pTermNew = Abc_NtkCreateBo( pNetAct.pNtk );
            Abc_ObjAddFanin( pTermNew, pBox );
            Abc_ObjAddFanin( pNetAct, pTermNew );
        }
        // remove the bundle
        Ver_ParseFreeBundle( pBundle ); pBundle = nil;
        Vec_PtrWriteEntry( (Vec_Ptr_t *)pBox.pCopy, j, nil );
    }
    // inlining of macro ABC_FREE
    ((pName) ? (free((char *) (pName)), (pName) = 0) : 0);
    return 1;
}


/**Function*************************************************************

  Synopsis    [Drives the bundle in the given undef box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseDriveInputs(pMan *Ver_Man_t, vUndefs *Vec_Ptr_t) int32 {
    var Buffer string // was [200]byte in C

    var pBundle *Ver_Bundle_t;
    var pNtk *Abc_Ntk_t;
    var pBox, pBox2, pTerm, pTermNew, pNetFormal, pNetAct *Abc_Obj_t;
    var i int32;
    var k int32;
    var j int32;
    var m int32;
    var CountCur int32;
    var CountTotal int32 = -1;
    // iterate through the undef boxes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vUndefs)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(vUndefs, i)), 1); i++  {
        // count the number of unconnected bundles for instances of this type of box
        CountTotal = -1;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize((Vec_Ptr_t *)pNtk.pData)) && (((pBox) = (Abc_Obj_t *)Vec_PtrEntry((Vec_Ptr_t *)pNtk.pData, k)), 1); k++  {
            CountCur = 0;
            // inlining of macro Vec_PtrForEachEntry
            for  j = 0; (j < Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry((Vec_Ptr_t *)pBox.pCopy, j)), 1); j++  {
                CountCur += (pBundle != nil);
            }
            if CountTotal == -1 {
                CountTotal = CountCur;
            } else if CountTotal != CountCur {
                sprintf( pMan.sError, "The number of formal inputs (%d) is different from the expected one (%d) when instantiating network %s in box %s.", 
                                    CountCur, CountTotal, pNtk.pName, Abc_ObjName(pBox) );
                Ver_ParsePrintErrorMessage( pMan );
                return 0;
            }
        }

        // create formals
        pBox = (Abc_Obj_t *)Vec_PtrEntry( (Vec_Ptr_t *)pNtk.pData, 0 );
        // inlining of macro Vec_PtrForEachEntry
        for  j = 0; (j < Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry((Vec_Ptr_t *)pBox.pCopy, j)), 1); j++  {
            if pBundle == nil {
                continue; }
            // inlining of macro Vec_PtrForEachEntry
            for  m = 0; (m < Vec_PtrSize(pBundle.vNetsActual)) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, m)), 1); m++  {
                // find create the formal net
                if Vec_PtrSize(pBundle.vNetsActual) == 1 {
                    Buffer = fmt.Sprintf("%s", pBundle.pNameFormal);
                } else {
                    Buffer = fmt.Sprintf("%s[%d]", pBundle.pNameFormal, m); }
                assert(Abc_NtkFindNet( pNtk, Buffer ) == nil, "Abc_NtkFindNet( pNtk, Buffer ) == nil");
                pNetFormal = Abc_NtkFindOrCreateNet( pNtk, Buffer );
                // connect
                pTerm = Abc_NtkCreateBi( pNtk );
                assert(Abc_NtkBoxNum(pNtk) <= 1, "Abc_NtkBoxNum(pNtk) <= 1");
                if Abc_NtkBoxNum(pNtk) {
                    pBox2 = Abc_NtkBox(pNtk,0);
                } else {
                    pBox2 = Abc_NtkCreateBlackbox(pNtk);
                }
                Abc_ObjAddFanin( pNetFormal, Abc_NtkCreatePi(pNtk) );
                Abc_ObjAddFanin( pTerm, pNetFormal );
                Abc_ObjAddFanin( pBox2, pTerm );
            }
        }

        // go through all the boxes
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize((Vec_Ptr_t *)pNtk.pData)) && (((pBox) = (Abc_Obj_t *)Vec_PtrEntry((Vec_Ptr_t *)pNtk.pData, k)), 1); k++  {
            // go through all the bundles
            // inlining of macro Vec_PtrForEachEntry
            for  j = 0; (j < Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry((Vec_Ptr_t *)pBox.pCopy, j)), 1); j++  {
                if pBundle == nil {
                    continue; }
                // drive the nets by the undef box
                // inlining of macro Vec_PtrForEachEntryReverse
                for  m = Vec_PtrSize(pBundle.vNetsActual) - 1; (m >= 0) && (((pNetAct) = (Abc_Obj_t *)Vec_PtrEntry(pBundle.vNetsActual, m)), 1); m--  {
                    pTermNew = Abc_NtkCreateBi( pNetAct.pNtk );
                    Abc_ObjAddFanin( pBox, pTermNew );
                    Abc_ObjAddFanin( pTermNew, pNetAct );
                }
                // remove the bundle
                Ver_ParseFreeBundle( pBundle );
                Vec_PtrWriteEntry( (Vec_Ptr_t *)pBox.pCopy, j, nil );
            }

            // free the bundles
            Vec_PtrFree( (Vec_Ptr_t *)pBox.pCopy );
            pBox.pCopy = nil;
        }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Returns the max size of any undef box.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseMaxBoxSize(vUndefs *Vec_Ptr_t) int32 {
    var pNtk *Abc_Ntk_t;
    var pBox *Abc_Obj_t;
    var i, k, nMaxSize int32;
    // go through undef box types
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vUndefs)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(vUndefs, i)), 1); i++  {
        // go through instances of this type
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize((Vec_Ptr_t *)pNtk.pData)) && (((pBox) = (Abc_Obj_t *)Vec_PtrEntry((Vec_Ptr_t *)pNtk.pData, k)), 1); k++  {
            // check the number of bundles of this instance
            if nMaxSize < Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy) {
                nMaxSize = Vec_PtrSize((Vec_Ptr_t *)pBox.pCopy); }
        }
    }
    return nMaxSize;
}

/**Function*************************************************************

  Synopsis    [Prints the comprehensive report into a log file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParsePrintLog(pMan *Ver_Man_t) {
    var pNtk, pNtkBox *Abc_Ntk_t;
    var pBox *Abc_Obj_t;
    var pFile *os.File;
    var pNameGeneric string;
    var Buffer string // was [1000]byte in C

    var i, k, Count1 int32;

    // open the log file
    pNameGeneric = Extra_FileNameGeneric( pMan.pFileName );
    Buffer = fmt.Sprintf("%s.log", pNameGeneric);
    // inlining of macro ABC_FREE
    ((pNameGeneric) ? (free((char *) (pNameGeneric)), (pNameGeneric) = 0) : 0);
    pFile = fopen( Buffer, "w" );

    // count the total number of instances and how many times they occur
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        pNtk.fHieVisited = 0;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachBox
        for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, k)), 1); k++  {
            if Abc_ObjIsLatch(pBox) {
                continue; }
            pNtkBox = (Abc_Ntk_t *)pBox.pData;
            if pNtkBox == nil {
                continue; }
            pNtkBox.fHieVisited++;
        }
    }
    // print each box and its stats
    fmt.Fprintf(pFile, "The hierarhical design %s contains %d modules:\n", pMan.pFileName, Vec_PtrSize(pMan.pDesign.vModules));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        fmt.Fprintf(pFile, "%-50s : ", Abc_NtkName(pNtk));
        if !(Ver_NtkIsDefined(pNtk) != 0) {
            fmt.Fprintf(pFile, "undefbox");
        } else if Abc_NtkHasBlackbox(pNtk) {
            fmt.Fprintf(pFile, "blackbox");
        } else {
            fmt.Fprintf(pFile, "logicbox"); }
        fmt.Fprintf(pFile, " instantiated %6d times ", pNtk.fHieVisited);
//        fprintf( pFile, "\n   " );
        fmt.Fprintf(pFile, " pi = %4d", Abc_NtkPiNum(pNtk));
        fmt.Fprintf(pFile, " po = %4d", Abc_NtkPoNum(pNtk));
        fmt.Fprintf(pFile, " nd = %8d", Abc_NtkNodeNum(pNtk));
        fmt.Fprintf(pFile, " lat = %6d", Abc_NtkLatchNum(pNtk));
        fmt.Fprintf(pFile, " box = %6d", Abc_NtkBoxNum(pNtk)-Abc_NtkLatchNum(pNtk));
        fmt.Fprintf(pFile, "\n");
        Count1 += (Abc_NtkPoNum(pNtk) == 1);
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
        pNtk.fHieVisited = 0;
    }
    fmt.Fprintf(pFile, "The number of modules with one output = %d (%.2f %%).\n", Count1, 100.0 * Count1/Vec_PtrSize(pMan.pDesign.vModules)); 

    // report instances with dangling outputs
    if Vec_PtrSize(pMan.pDesign.vModules) > 1 {
        var vBundles *Vec_Ptr_t;
        var pBundle *Ver_Bundle_t;
        var j, nActNets, Counter int32;
        // count the number of instances with dangling outputs
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
            // inlining of macro Abc_NtkForEachBox
            for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, k)), 1); k++  {
                if Abc_ObjIsLatch(pBox) {
                    continue; }
                vBundles = (Vec_Ptr_t *)pBox.pCopy;
                pNtkBox = (Abc_Ntk_t *)pBox.pData;
                if pNtkBox == nil {
                    continue; }
                if !(Ver_NtkIsDefined(pNtkBox) != 0) {
                    continue; }
                // count the number of actual nets
                nActNets = 0;
                // inlining of macro Vec_PtrForEachEntry
                for  j = 0; (j < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, j)), 1); j++  {
                    nActNets += Vec_PtrSize(pBundle.vNetsActual);
                }
                // the box is defined and will be connected
                if nActNets != Abc_NtkPiNum(pNtkBox) + Abc_NtkPoNum(pNtkBox) {
                    Counter++; }
            }
        }
        if Counter == 0 {
            fmt.Fprintf(pFile, "The outputs of all box instances are connected.\n");
        } else {
            fmt.Fprintf(pFile, "\n");
            fmt.Fprintf(pFile, "The outputs of %d box instances are not connected:\n", Counter);
            // enumerate through the boxes
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(pMan.pDesign.vModules)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(pMan.pDesign.vModules, i)), 1); i++  {
                // inlining of macro Abc_NtkForEachBox
                for  k = 0; (k < Vec_PtrSize((pNtk).vBoxes)) && (((pBox) = Abc_NtkBox(pNtk, k)), 1); k++  {
                    if Abc_ObjIsLatch(pBox) {
                        continue; }
                    vBundles = (Vec_Ptr_t *)pBox.pCopy;
                    pNtkBox = (Abc_Ntk_t *)pBox.pData;
                    if pNtkBox == nil {
                        continue; }
                    if !(Ver_NtkIsDefined(pNtkBox) != 0) {
                        continue; }
                    // count the number of actual nets
                    nActNets = 0;
                    // inlining of macro Vec_PtrForEachEntry
                    for  j = 0; (j < Vec_PtrSize(vBundles)) && (((pBundle) = (Ver_Bundle_t *)Vec_PtrEntry(vBundles, j)), 1); j++  {
                        nActNets += Vec_PtrSize(pBundle.vNetsActual);
                    }
                    // the box is defined and will be connected
                    if nActNets != Abc_NtkPiNum(pNtkBox) + Abc_NtkPoNum(pNtkBox) {
                        fmt.Fprintf(pFile, "In module \"%s\" instance \"%s\" of box \"%s\" has different numbers of actual/formal nets (%d/%d).\n", Abc_NtkName(pNtk), Abc_ObjName(pBox), Abc_NtkName(pNtkBox), nActNets, Abc_NtkPiNum(pNtkBox) + Abc_NtkPoNum(pNtkBox)); }
                }
            }
        }
    }
    _ = (pFile).Close();
    printf( "Hierarchy statistics can be found in log file \"%s\".\n", Buffer );
}


/**Function*************************************************************

  Synopsis    [Attaches the boxes to the network.]

  Description [This procedure is called after the design is parsed. 
  At that point, all the defined models have their PIs present. 
  They are connected first. Next undef boxes are processed (if present).
  Iteratively, one bundle is selected to be driven by the undef boxes in such 
  a way that there is no conflict (if it is driven by an instance of the box,
  no other net will be driven twice by the same formal net of some other instance 
  of the same box). In the end, all the remaining nets that cannot be driven 
  by the undef boxes are connected to the undef boxes as inputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseAttachBoxes(pMan *Ver_Man_t) int32 {
    var fPrintLog int32;
    var pNtk *Abc_Ntk_t  = nil;
    var pBundle *Ver_Bundle_t;
    var vUndefs *Vec_Ptr_t;
    var i, RetValue, Counter, nMaxBoxSize int32;

    // print the log file
    if (fPrintLog != 0) && pMan.pDesign.vModules && Vec_PtrSize(pMan.pDesign.vModules) > 1 {
        Ver_ParsePrintLog( pMan ); }

    // connect defined boxes
    RetValue = Ver_ParseConnectDefBoxes( pMan );
    if RetValue < 2 {
        return RetValue; }

    // report the boxes
    Ver_ParseReportUndefBoxes( pMan );

    // collect undef box types and their actual instances
    vUndefs = Ver_ParseCollectUndefBoxes( pMan );
    assert(Vec_PtrSize( vUndefs ) > 0, "Vec_PtrSize( vUndefs ) > 0");

    // go through all undef box types
    Counter = 0;
    nMaxBoxSize = Ver_ParseMaxBoxSize( vUndefs );
    for  (Ver_ParseCheckNondrivenNets(vUndefs) != 0) && Counter < nMaxBoxSize  {
        // go through undef box types
        pBundle = nil;
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vUndefs)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(vUndefs, i)), 1); i++  {
            if pBundle = Ver_ParseGetNondrivenBundle( pNtk, Counter ) {
                break; }
        }
        if pBundle == nil {
            Counter++;
            continue;
        }
        // drive this bundle by this box
        if !(Ver_ParseDriveFormal( pMan, pNtk, pBundle ) != 0) {
            return 0; }
    }

    // make all the remaining bundles the drivers of undefs
    if !(Ver_ParseDriveInputs( pMan, vUndefs ) != 0) {
        return 0; }

    // cleanup
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vUndefs)) && (((pNtk) = (Abc_Ntk_t *)Vec_PtrEntry(vUndefs, i)), 1); i++  {
        Vec_PtrFree( (Vec_Ptr_t *)pNtk.pData );
        pNtk.pData = nil;
    }
    Vec_PtrFree( vUndefs ); 
    return 1;
}


/**Function*************************************************************

  Synopsis    [Creates PI terminal and net.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseCreatePi(pNtk *Abc_Ntk_t, pName string) *Abc_Obj_t {
    var pNet, pTerm *Abc_Obj_t;
    // get the PI net
//    pNet  = Ver_ParseFindNet( pNtk, pName );
//    if ( pNet )
//        printf( "Warning: PI \"%s\" appears twice in the list.\n", pName );
    pNet  = Abc_NtkFindOrCreateNet( pNtk, pName );
    // add the PI node
    pTerm = Abc_NtkCreatePi( pNtk );
    Abc_ObjAddFanin( pNet, pTerm );
    return pTerm;
}

/**Function*************************************************************

  Synopsis    [Creates PO terminal and net.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseCreatePo(pNtk *Abc_Ntk_t, pName string) *Abc_Obj_t {
    var pNet, pTerm *Abc_Obj_t;
    // get the PO net
//    pNet  = Ver_ParseFindNet( pNtk, pName );
//    if ( pNet && Abc_ObjFaninNum(pNet) == 0 )
//        printf( "Warning: PO \"%s\" appears twice in the list.\n", pName );
    pNet  = Abc_NtkFindOrCreateNet( pNtk, pName );
    // add the PO node
    pTerm = Abc_NtkCreatePo( pNtk );
    Abc_ObjAddFanin( pTerm, pNet );
    return pTerm;
}

/**Function*************************************************************

  Synopsis    [Create a latch with the given input/output.]

  Description [By default, the latch value is a don't-care.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseCreateLatch(pNtk *Abc_Ntk_t, pNetLI *Abc_Obj_t, pNetLO *Abc_Obj_t) *Abc_Obj_t {
    var pLatch, pTerm *Abc_Obj_t;
    // add the BO terminal
    pTerm = Abc_NtkCreateBi( pNtk );
    Abc_ObjAddFanin( pTerm, pNetLI );
    // add the latch box
    pLatch = Abc_NtkCreateLatch( pNtk );
    Abc_ObjAddFanin( pLatch, pTerm  );
    // add the BI terminal
    pTerm = Abc_NtkCreateBo( pNtk );
    Abc_ObjAddFanin( pTerm, pLatch );
    // get the LO net
    Abc_ObjAddFanin( pNetLO, pTerm );
    // set latch name
    Abc_ObjAssignName( pLatch, Abc_ObjName(pNetLO), "L" );
    Abc_LatchSetInitDc( pLatch );
    return pLatch;
}

/**Function*************************************************************

  Synopsis    [Creates inverter and returns its net.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseCreateInv(pNtk *Abc_Ntk_t, pNet *Abc_Obj_t) *Abc_Obj_t {
    var pObj *Abc_Obj_t;
    pObj = Abc_NtkCreateNodeInv( pNtk, pNet );
    pNet = Abc_NtkCreateNet( pNtk );
    Abc_ObjAddFanin( pNet, pObj );
    return pNet;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
