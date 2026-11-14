package ver

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"strings"
)

import "os"

type Abc_Ntk_t = abc.Network
type Abc_Obj_t = abc.Object
type Mio_Gate_t = mio.Mio_Gate_t
type Vec_Ptr_t = vec.Vec_Ptr_t

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

var glo_fMapped int32 // this is bad!

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start parser.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Stop parser.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [File parser.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseFile(pFileName string, pGateLib *abc.Design, fCheck int32, fUseMemMan int32) *abc.Design {
	p := Ver_ParseStart(pFileName, pGateLib)
	if p == nil {
		return nil
	}
	// start the parser
	p.fMapped = glo_fMapped
	p.fCheck = fCheck
	p.fUseMemMan = fUseMemMan
	if glo_fMapped != 0 {
		if p.pDesign.FunctionManager != nil {
			hop.Hop_ManStop(p.pDesign.FunctionManager)
			p.pDesign.FunctionManager = nil
		}
	}
	// parse the file
	Ver_ParseInternal(p)
	// save the result
	pDesign := p.pDesign
	p.pDesign = nil
	// stop the parser
	Ver_ParseStop(p)
	return pDesign
}

/*
*Function*************************************************************

	Synopsis    [File parser.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ver_ParseInternal(pMan *Ver_Man_t) {
	// preparse the modeles
	pMan.pProgress = extra.Extra_ProgressBarStart(os.Stdout, Ver_StreamGetFileSize(pMan.pReader))
	for {
		// get the next token
		pToken := Ver_ParseGetName(pMan)
		if pToken == nil {
			break
		}
		if *pToken != "module" {
			verSetError(pMan, "Cannot read \"module\" directive.")
			Ver_ParsePrintErrorMessage(pMan)
			return
		}
		// parse the module
		if Ver_ParseModule(pMan) == 0 {
			return
		}
	}
	extra.Extra_ProgressBarStop(pMan.pProgress)
	pMan.pProgress = nil

	// process defined and undefined boxes
	if Ver_ParseAttachBoxes(pMan) == 0 {
		return
	}

	// connect the boxes and check
	// inlining of macro Vec_PtrForEachEntry
	for _, pNtk := range pMan.pDesign.Modules {
		// fix the dangling nets
		abc.Abc_NtkFinalizeRead(pNtk)
		// check the network for correctness
		if pMan.fCheck != 0 && abc.Abc_NtkCheckRead(pNtk) == 0 {
			pMan.fTopLevel = 1
			verSetError(pMan, fmt.Sprintf("The network check has failed for network %s.", pNtk.Name))
			Ver_ParsePrintErrorMessage(pMan)
			return
		}
	}
}

/**Function*************************************************************

  Synopsis    [File parser.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Prints the error message including the file name and line number.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Finds the network by name or create a new blackbox network.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Finds the network by name or create a new blackbox network.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts the network from the blackbox type into a different one.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Parses one Verilog module.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseModule(pMan *Ver_Man_t) int32 {
	p := pMan.pReader
	nameToken := Ver_ParseGetName(pMan)
	if nameToken == nil {
		return 0
	}
	n := Ver_ParseFindOrCreateNetwork(pMan, *nameToken)
	fail := func(message string) int32 {
		verSetError(pMan, message)
		Ver_ParsePrintErrorMessage(pMan)
		return 0
	}

	if got := Ver_StreamPopChar(p); got != '(' {
		return fail(fmt.Sprintf("Cannot find \"(\" after \"module\" in network %s.", n.Name))
	}
	// The module header lists names only; declarations below determine their
	// direction and range. Consume the list through its closing parenthesis.
	for {
		if Ver_ParseGetName(pMan) == nil {
			return 0
		}
		if Ver_StreamPopChar(p) != ',' {
			break
		}
	}
	if Ver_ParseSkipComments(pMan) == 0 {
		return 0
	}
	if Ver_StreamPopChar(p) != ';' {
		return fail("Expected closing parenthesis after \"module\".")
	}

	wordToken := Ver_ParseGetName(pMan)
	if wordToken == nil {
		return 0
	}
	word := *wordToken
	for {
		extra.Extra_ProgressBarUpdate(pMan.pProgress, Ver_StreamGetCurPosition(p), "")
		var kind Ver_SignalType_t
		isDeclaration := true
		switch word {
		case "input":
			kind = VER_SIG_INPUT
		case "output":
			kind = VER_SIG_OUTPUT
		case "reg":
			kind = VER_SIG_REG
		case "wire":
			kind = VER_SIG_WIRE
		case "inout":
		default:
			isDeclaration = false
		}
		if !isDeclaration {
			break
		}
		if Ver_ParseSignal(pMan, n, kind) == 0 {
			return 0
		}
		wordToken = Ver_ParseGetName(pMan)
		if wordToken == nil {
			return 0
		}
		word = *wordToken
	}

	for {
		extra.Extra_ProgressBarUpdate(pMan.pProgress, Ver_StreamGetCurPosition(p), "")
		if word == "endmodule" {
			break
		}
		var result int32
		switch word {
		case "and":
			result = Ver_ParseGateStandard(pMan, n, VER_GATE_AND)
		case "or":
			result = Ver_ParseGateStandard(pMan, n, VER_GATE_OR)
		case "xor":
			result = Ver_ParseGateStandard(pMan, n, VER_GATE_XOR)
		case "buf":
			result = Ver_ParseGateStandard(pMan, n, VER_GATE_BUF)
		case "nand":
			result = Ver_ParseGateStandard(pMan, n, VER_GATE_NAND)
		case "nor":
			result = Ver_ParseGateStandard(pMan, n, VER_GATE_NOR)
		case "xnor":
			result = Ver_ParseGateStandard(pMan, n, VER_GATE_XNOR)
		case "not":
			result = Ver_ParseGateStandard(pMan, n, VER_GATE_NOT)
		case "dff":
			result = Ver_ParseFlopStandard(pMan, n)
		case "assign":
			result = Ver_ParseAssign(pMan, n)
		case "always":
			result = Ver_ParseAlways(pMan, n)
		case "initial":
			result = Ver_ParseInitial(pMan, n)
		case "wire":
			result = Ver_ParseSignal(pMan, n, VER_SIG_WIRE)
		default:
			var gate *mio.Mio_Gate_t
			if lib, ok := pMan.pGenlib.(*mio.Mio_Library_t); ok && lib != nil {
				gate = mio.Mio_LibraryReadGateByName(lib, word, "")
			}
			if gate != nil {
				result = Ver_ParseGate(pMan, n, gate)
			} else {
				box := Ver_ParseFindOrCreateNetwork(pMan, word)
				result = Ver_ParseBox(pMan, n, box)
			}
		}
		if result == 0 {
			return 0
		}
		if Ver_ParseSkipComments(pMan) == 0 {
			return 0
		}
		wordToken = Ver_ParseGetName(pMan)
		if wordToken == nil {
			return 0
		}
		word = *wordToken
	}
	if n.Function == uint8(abc.ABC_FUNC_BLACKBOX) {
		if abc.Abc_NtkNodeNum(n) > 0 || abc.Abc_NtkBoxNum(n) > 0 {
			if Ver_ParseConvertNetwork(pMan, n, pMan.fMapped) == 0 {
				return 0
			}
		} else {
			box := abc.Abc_NtkCreateBlackbox(n)
			for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
				obj := abc.Abc_NtkPi(n, i)
				term := abc.Abc_NtkCreateBi(n)
				abc.Abc_ObjAddFanin(term, abc.Abc_ObjFanout0(obj))
				abc.Abc_ObjAddFanin(box, term)
			}
			for i := int32(0); i < abc.Abc_NtkPoNum(n); i++ {
				obj := abc.Abc_NtkPo(n, i)
				term := abc.Abc_NtkCreateBo(n)
				abc.Abc_ObjAddFanin(term, box)
				abc.Abc_ObjAddFanin(abc.Abc_ObjFanin0(obj), term)
			}
		}
	}
	Ver_ParseRemoveSuffixTable(pMan)
	return 1
}

/**Function*************************************************************

  Synopsis    [Lookups the suffix of the signal of the form [m:n].]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Lookups the suffix of the signal of the form [m:n].]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Lookups the suffic of the signal of the form [m:n].]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Determine signal prefix of the form [Beg:End].]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Determine signal suffix of the form [m:n].]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Returns the values of constant bits.]

  Description [The resulting bits are in MSB to LSB order.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description [The signals are added in the order from LSB to MSB.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Returns the index of the given pin the gate.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Parses one directive.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseGate(pMan *Ver_Man_t, pNtk *abc.Network, pGate *mio.Mio_Gate_t) int32 {
	p := pMan.pReader
	nFanins := mio.Mio_GateReadPinNum(pGate)
	if pMan.fMapped != 1 {
		verSetError(pMan, "The network appears to be mapped. Use \"r -m\" to read mapped Verilog.")
		Ver_ParsePrintErrorMessage(pMan)
		return 0
	}
	if Ver_ParseConvertNetwork(pMan, pNtk, 1) == 0 {
		return 0
	}
	word := Ver_ParseGetName(pMan)
	if word == nil {
		return 0
	}
	if Ver_StreamPopChar(p) != '(' {
		verSetError(pMan, fmt.Sprintf("Cannot parse gate %s (expected opening parenthesis).", mio.Mio_GateReadName(pGate)))
		Ver_ParsePrintErrorMessage(pMan)
		return 0
	}
	Ver_ParseSkipComments(pMan)
	node := abc.Abc_NtkCreateNode(pNtk)
	node.Data = pGate
	var node2 *abc.Object
	if twin := mio.Mio_GateReadTwin(pGate); twin != nil {
		node2 = abc.Abc_NtkCreateNode(pNtk)
		node2.Data = twin
	}
	vec.Vec_IntClear(pMan.vPerm)
	for {
		if Ver_StreamPopChar(p) != '.' {
			verSetError(pMan, fmt.Sprintf("Cannot parse gate %s (expected .).", mio.Mio_GateReadName(pGate)))
			Ver_ParsePrintErrorMessage(pMan)
			return 0
		}
		formal := Ver_ParseGetName(pMan)
		if formal == nil {
			return 0
		}
		input := Ver_FindGateInput(pGate, *formal)
		if input == -1 {
			verSetError(pMan, fmt.Sprintf("Formal input name %s cannot be found in the gate %s.", *formal, mio.Mio_GateReadOutName(pGate)))
			Ver_ParsePrintErrorMessage(pMan)
			return 0
		}
		if Ver_StreamPopChar(p) != '(' {
			verSetError(pMan, fmt.Sprintf("Cannot formal parameter %s of gate %s (expected opening parenthesis).", *formal, mio.Mio_GateReadName(pGate)))
			Ver_ParsePrintErrorMessage(pMan)
			return 0
		}
		actualName := Ver_ParseGetName(pMan)
		if actualName == nil {
			return 0
		}
		assert(len(*actualName) > 0 && (*actualName)[0] != '~', "pWord[0] != '~'")
		actual := Ver_ParseFindNet(pNtk, *actualName)
		if actual == nil {
			verSetError(pMan, fmt.Sprintf("Actual net %s is missing.", *actualName))
			Ver_ParsePrintErrorMessage(pMan)
			return 0
		}
		if Ver_StreamPopChar(p) != ')' {
			verSetError(pMan, fmt.Sprintf("Cannot formal parameter %s of gate %s (expected closing parenthesis).", *actualName, mio.Mio_GateReadName(pGate)))
			Ver_ParsePrintErrorMessage(pMan)
			return 0
		}
		if input < nFanins {
			vec.Vec_IntPush(pMan.vPerm, input)
			abc.Abc_ObjAddFanin(node, actual)
			if node2 != nil {
				abc.Abc_ObjAddFanin(node2, actual)
			}
		} else if input == nFanins {
			abc.Abc_ObjAddFanin(actual, node)
		} else if input == nFanins+1 && node2 != nil {
			abc.Abc_ObjAddFanin(actual, node2)
		} else {
			panic("gate output pin index is out of range")
		}
		Ver_ParseSkipComments(pMan)
		symbol := Ver_StreamPopChar(p)
		if symbol == ')' {
			break
		}
		if symbol != ',' {
			verSetError(pMan, fmt.Sprintf("Cannot formal parameter %s of gate %s (expected closing parenthesis).", *actualName, mio.Mio_GateReadName(pGate)))
			Ver_ParsePrintErrorMessage(pMan)
			return 0
		}
		Ver_ParseSkipComments(pMan)
	}
	if abc.Abc_ObjFaninNum(node) != nFanins || abc.Abc_ObjFanoutNum(node) != 1 {
		verSetError(pMan, fmt.Sprintf("Parsing of gate %s has failed.", mio.Mio_GateReadName(pGate)))
		Ver_ParsePrintErrorMessage(pMan)
		return 0
	}
	Ver_ParseSkipComments(pMan)
	if Ver_StreamPopChar(p) != ';' {
		verSetError(pMan, fmt.Sprintf("Cannot read gate %s (expected closing semicolumn).", mio.Mio_GateReadName(pGate)))
		Ver_ParsePrintErrorMessage(pMan)
		return 0
	}
	if vec.Vec_IntSize(pMan.vPerm) > 0 {
		nonIdentity := false
		for i := 0; i < vec.Vec_IntSize(pMan.vPerm); i++ {
			if vec.Vec_IntEntry(pMan.vPerm, i) != int32(i) {
				nonIdentity = true
				break
			}
		}
		if nonIdentity {
			original := node.Fanins()
			for i := 0; i < vec.Vec_IntSize(pMan.vPerm); i++ {
				formalIndex := int(vec.Vec_IntEntry(pMan.vPerm, i))
				if formalIndex < 0 || formalIndex >= len(original) || node.Network().PatchFanin(node, formalIndex, original[i]) != nil {
					panic("invalid mapped-gate input permutation")
				}
			}
		}
	}
	return 1
}

/**Function*************************************************************

  Synopsis    [Connects one box to the network]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Connects one box to the network]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_ParseBox(pMan *Ver_Man_t, pNtk *abc.Network, pNtkBox *abc.Network) int32 {
	p := pMan.pReader
	word := Ver_ParseGetName(pMan)
	if word == nil {
		return 0
	}
	box := abc.Abc_NtkCreateBlackbox(pNtk)
	box.Data = pNtkBox
	abc.Abc_ObjAssignName(box, *word, "")
	fail := func(message string) int32 {
		verSetError(pMan, message)
		Ver_ParsePrintErrorMessage(pMan)
		return 0
	}
	if pMan.boxBundles == nil {
		pMan.boxBundles = make(map[*abc.Object]*vec.Vec_Ptr_t)
	}
	bundles := vec.Vec_PtrAlloc(16)
	pMan.boxBundles[box] = bundles
	if Ver_StreamPopChar(p) != '(' {
		return fail(fmt.Sprintf("Cannot parse box %s (expected opening parenthesis).", abc.Abc_ObjName(box)))
	}
	Ver_ParseSkipComments(pMan)
	for {
		bundle := &Ver_Bundle_t{vNetsActual: vec.Vec_PtrAlloc(4)}
		vec.Vec_PtrPush(bundles, bundle)
		formalGiven := Ver_StreamScanChar(p) == '.'
		if formalGiven {
			Ver_StreamPopChar(p)
			formal := Ver_ParseGetName(pMan)
			if formal == nil {
				return 0
			}
			bundle.pNameFormal = strings.Clone(*formal)
			if Ver_StreamPopChar(p) != '(' {
				return fail(fmt.Sprintf("Cannot formal parameter %s of box %s (expected opening parenthesis).", *formal, abc.Abc_ObjName(box)))
			}
			Ver_ParseSkipComments(pMan)
		}
		appendNamedActual := func(name string) bool {
			actual := Ver_ParseFindNet(pNtk, name)
			if actual == nil && (strings.HasPrefix(name, "Open_") || strings.HasPrefix(name, "dct_unconnected")) {
				actual = abc.Abc_NtkCreateNet(pNtk)
			}
			if actual == nil {
				return false
			}
			vec.Vec_PtrPush(bundle.vNetsActual, actual)
			return true
		}
		if Ver_StreamScanChar(p) == '{' {
			Ver_StreamPopChar(p)
			bitIndex := 0
			for {
				Ver_ParseSkipComments(pMan)
				actualWord := Ver_ParseGetName(pMan)
				if actualWord == nil {
					return 0
				}
				actualName := *actualWord
				last := len(actualName) > 0 && actualName[len(actualName)-1] == '}'
				if last {
					actualName = actualName[:len(actualName)-1]
				}
				if actualName == "" {
					break
				}
				if actualName[0] >= '1' && actualName[0] <= '9' {
					if Ver_ParseConstant(pMan, actualName) == 0 {
						return 0
					}
					for k := 0; k < vec.Vec_PtrSize(pMan.vNames); k++ {
						value := vec.Vec_PtrEntry(pMan.vNames, k)
						bit := 0
						if value != nil {
							bit = 1
						}
						if !appendNamedActual(fmt.Sprintf("1'b%d", bit)) {
							verSetError(pMan, fmt.Sprintf("Actual net \"1'b%d\" is missing in gate \"%s\".", bit, abc.Abc_ObjName(box)))
							Ver_ParsePrintErrorMessage(pMan)
							return 0
						}
						bitIndex++
					}
				} else {
					msb, lsb := int32(-1), int32(-1)
					if strings.HasSuffix(actualName, "]") && pMan.fNameLast == 0 {
						if Ver_ParseSignalSuffix(pMan, &actualName, &msb, &lsb) == 0 {
							return 0
						}
					} else {
						Ver_ParseLookupSuffix(pMan, actualName, &msb, &lsb)
					}
					if msb < 0 && lsb < 0 {
						if !appendNamedActual(actualName) {
							return fail(fmt.Sprintf("Actual net %q is missing in box %q.", actualName, abc.Abc_ObjName(box)))
						}
						bitIndex++
					} else {
						step := int32(1)
						if msb > lsb {
							step = -1
						}
						for bit := msb; ; bit += step {
							name := fmt.Sprintf("%s[%d]", actualName, bit)
							if !appendNamedActual(name) {
								return fail(fmt.Sprintf("Actual net %q is missing in box %q.", actualName, abc.Abc_ObjName(box)))
							}
							bitIndex++
							if bit == lsb {
								break
							}
						}
					}
				}
				if last {
					break
				}
				Ver_ParseSkipComments(pMan)
				symbol := Ver_StreamPopChar(p)
				if symbol == '}' {
					break
				}
				if symbol != ',' {
					return fail(fmt.Sprintf("Cannot parse formal parameter %s of gate %s (expected comma).", actualName, abc.Abc_ObjName(box)))
				}
			}
		} else {
			actualWord := Ver_ParseGetName(pMan)
			if actualWord == nil {
				return 0
			}
			actualName := *actualWord
			if actualName == "" {
				vec.Vec_PtrPush(bundle.vNetsActual, abc.Abc_NtkCreateNet(pNtk))
			} else if actual := Ver_ParseFindNet(pNtk, actualName); actual != nil {
				vec.Vec_PtrPush(bundle.vNetsActual, actual)
			} else {
				msb, lsb := int32(-1), int32(-1)
				Ver_ParseLookupSuffix(pMan, actualName, &msb, &lsb)
				if msb < 0 && lsb < 0 {
					Ver_ParseSignalSuffix(pMan, &actualName, &msb, &lsb)
				}
				if msb >= 0 && lsb >= 0 {
					step := int32(1)
					if msb > lsb {
						step = -1
					}
					for bit := msb; ; bit += step {
						if !appendNamedActual(fmt.Sprintf("%s[%d]", actualName, bit)) {
							return fail(fmt.Sprintf("Actual net %q is missing in box %q.", actualName, abc.Abc_ObjName(box)))
						}
						if bit == lsb {
							break
						}
					}
				} else if strings.HasPrefix(actualName, "Open_") || strings.HasPrefix(actualName, "dct_unconnected") {
					vec.Vec_PtrPush(bundle.vNetsActual, abc.Abc_NtkCreateNet(pNtk))
				} else {
					return fail(fmt.Sprintf("Actual net %q is missing in box %q.", actualName, abc.Abc_ObjName(box)))
				}
			}
		}
		if formalGiven {
			Ver_ParseSkipComments(pMan)
			if Ver_StreamPopChar(p) != ')' {
				return fail(fmt.Sprintf("Cannot parse formal parameter %s of box %s (expected closing parenthesis).", bundle.pNameFormal, abc.Abc_ObjName(box)))
			}
			Ver_ParseSkipComments(pMan)
		}
		symbol := Ver_StreamPopChar(p)
		if symbol == ')' {
			break
		}
		if symbol != ',' {
			return fail(fmt.Sprintf("Cannot parse formal parameter %s of box %s (expected comma).", bundle.pNameFormal, abc.Abc_ObjName(box)))
		}
		Ver_ParseSkipComments(pMan)
	}
	Ver_ParseSkipComments(pMan)
	if Ver_StreamPopChar(p) != ';' {
		return fail(fmt.Sprintf("Cannot read box %s (expected closing semicolumn).", abc.Abc_ObjName(box)))
	}
	return 1
}

func Ver_ParseConnectBox(pMan *Ver_Man_t, box *Abc_Obj_t) int32 {
	model, ok := box.Data.(*abc.Network)
	if !ok || model == nil {
		return 0
	}
	bundles := pMan.boxBundles[box]
	if bundles == nil {
		return 0
	}
	parent := abc.Abc_ObjNtk(box)
	pis, pos := model.ObjectsOfKind(abc.ObjectPI), model.ObjectsOfKind(abc.ObjectPO)
	total := 0
	allNamed := true
	for i := 0; i < vec.Vec_PtrSize(bundles); i++ {
		b := vec.Vec_PtrEntry(bundles, i).(*Ver_Bundle_t)
		total += vec.Vec_PtrSize(b.vNetsActual)
		if b.pNameFormal == "" {
			allNamed = false
		}
	}
	formalCount := len(pis) + len(pos)
	if allNamed && vec.Vec_PtrSize(bundles) > formalCount {
		verSetError(pMan, fmt.Sprintf("The number of actual IOs (%d) is bigger than the number of formal IOs (%d) when instantiating network %s in box %s.", vec.Vec_PtrSize(bundles), formalCount, model.Name, abc.Abc_ObjName(box)))
		Ver_ParsePrintErrorMessage(pMan)
		return 0
	}
	if !allNamed && total != formalCount {
		verSetError(pMan, fmt.Sprintf("The number of actual IOs (%d) is different from the number of formal IOs (%d) when instantiating network %s in box %s.", vec.Vec_PtrSize(bundles), formalCount, model.Name, abc.Abc_ObjName(box)))
		Ver_ParsePrintErrorMessage(pMan)
		return 0
	}
	findBundle := func(name string) *Ver_Bundle_t {
		for i := 0; i < vec.Vec_PtrSize(bundles); i++ {
			b := vec.Vec_PtrEntry(bundles, i).(*Ver_Bundle_t)
			if b.pNameFormal == name {
				return b
			}
		}
		// ABC accepts a formal vector declaration such as data[7:0]
		// when the instance names the vector as data.
		if strings.HasSuffix(name, "]") {
			bracket := strings.LastIndexByte(name, '[')
			if bracket > 0 {
				for i := 0; i < vec.Vec_PtrSize(bundles); i++ {
					b := vec.Vec_PtrEntry(bundles, i).(*Ver_Bundle_t)
					if len(b.pNameFormal) == bracket && strings.HasPrefix(name, b.pNameFormal) {
						return b
					}
				}
			}
		}
		return nil
	}
	cleanup := func() {
		for i := 0; i < vec.Vec_PtrSize(bundles); i++ {
			Ver_ParseFreeBundle(vec.Vec_PtrEntry(bundles, i).(*Ver_Bundle_t))
		}
		vec.Vec_PtrFree(bundles)
		delete(pMan.boxBundles, box)
	}
	connectInput := func(b *Ver_Bundle_t) {
		for i := vec.Vec_PtrSize(b.vNetsActual) - 1; i >= 0; i-- {
			actual := vec.Vec_PtrEntry(b.vNetsActual, i).(*abc.Object)
			bi := abc.Abc_NtkCreateBi(parent)
			abc.Abc_ObjAddFanin(box, bi)
			abc.Abc_ObjAddFanin(bi, actual)
		}
	}
	connectOutput := func(b *Ver_Bundle_t) bool {
		for i := vec.Vec_PtrSize(b.vNetsActual) - 1; i >= 0; i-- {
			actual := vec.Vec_PtrEntry(b.vNetsActual, i).(*abc.Object)
			if abc.Abc_ObjName(actual) == "1'b0" || abc.Abc_ObjName(actual) == "1'b1" {
				verSetError(pMan, fmt.Sprintf("It looks like formal output %s is driving a constant net (%s) when instantiating network %s in box %s.", b.pNameFormal, abc.Abc_ObjName(actual), model.Name, abc.Abc_ObjName(box)))
				cleanup()
				Ver_ParsePrintErrorMessage(pMan)
				return false
			}
			bo := abc.Abc_NtkCreateBo(parent)
			abc.Abc_ObjAddFanin(bo, box)
			abc.Abc_ObjAddFanin(actual, bo)
		}
		return true
	}
	if allNamed {
		for _, pi := range pis {
			name := abc.Abc_ObjName(abc.Abc_ObjFanout0(pi))
			b := findBundle(name)
			if b == nil {
				return failBoxAttach(pMan, fmt.Sprintf("Cannot find an actual net for the formal net %s when instantiating network %s in box %s.", name, model.Name, abc.Abc_ObjName(box)))
			}
			connectInput(b)
		}
		for _, po := range pos {
			name := abc.Abc_ObjName(abc.Abc_ObjFanin0(po))
			b := findBundle(name)
			if b == nil {
				bo := abc.Abc_NtkCreateBo(parent)
				net := abc.Abc_NtkFindOrCreateNet(parent, fmt.Sprintf("_temp_net%d", abc.Abc_ObjId(bo)))
				abc.Abc_ObjAddFanin(bo, box)
				abc.Abc_ObjAddFanin(net, bo)
				continue
			}
			if !connectOutput(b) {
				return 0
			}
		}
	} else {
		index := 0
		for range pis {
			b := vec.Vec_PtrEntry(bundles, index).(*Ver_Bundle_t)
			index++
			connectInput(b)
		}
		for range pos {
			b := vec.Vec_PtrEntry(bundles, index).(*Ver_Bundle_t)
			index++
			if !connectOutput(b) {
				return 0
			}
		}
	}
	cleanup()
	return 1
}

func failBoxAttach(pMan *Ver_Man_t, message string) int32 {
	verSetError(pMan, message)
	Ver_ParsePrintErrorMessage(pMan)
	return 0
}

func Ver_ParseConnectDefBoxes(pMan *Ver_Man_t) int32 {
	retValue := int32(1)
	for _, network := range pMan.pDesign.Modules {
		for _, box := range network.Boxes() {
			if abc.Abc_ObjIsLatch(box) != 0 || box.Data == nil {
				continue
			}
			model, ok := box.Data.(*abc.Network)
			if !ok {
				return failBoxAttach(pMan, fmt.Sprintf("Box %s has an invalid model.", abc.Abc_ObjName(box)))
			}
			if Ver_NtkIsDefined(model) == 0 {
				retValue = 2
				continue
			}
			if Ver_ParseConnectBox(pMan, box) == 0 {
				return 0
			}
			if abc.Abc_NtkHasBlackbox(model) != 0 {
				continue
			}
			abc.Abc_ObjBlackboxToWhitebox(box)
		}
	}
	return retValue
}

/*
*Function*************************************************************

	Synopsis    [Collects the undef boxes and maps them into their instances.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ver_ParseCollectUndefBoxes(pMan *Ver_Man_t) *vec.Vec_Ptr_t {
	// Reuse each undefined model's Data as the temporary instance list, as C does.
	for _, module := range pMan.pDesign.Modules {
		module.Data = nil
	}
	undefs := vec.Vec_PtrAlloc(16)
	for _, module := range pMan.pDesign.Modules {
		for _, box := range module.ObjectsOfKind(abc.ObjectBlackbox) {
			model, ok := box.Data.(*abc.Network)
			if !ok || model == nil || Ver_NtkIsDefined(model) != 0 {
				continue
			}
			instances, ok := model.Data.(*vec.Vec_Ptr_t)
			if !ok || instances == nil {
				instances = vec.Vec_PtrAlloc(16)
				model.Data = instances
				vec.Vec_PtrPush(undefs, model)
			}
			vec.Vec_PtrPush(instances, box)
		}
	}
	return undefs
}

/*
*Function*************************************************************

	Synopsis    [Reports how many times each type of undefined box occurs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Ver_ParseReportUndefBoxes(pMan *Ver_Man_t) {
	instances := make(map[*abc.Network]int)
	undefinedCount := 0
	for _, module := range pMan.pDesign.Modules {
		if Ver_NtkIsDefined(module) == 0 {
			undefinedCount++
		}
		for _, box := range module.ObjectsOfKind(abc.ObjectBlackbox) {
			if model, ok := box.Data.(*abc.Network); ok && model != nil && Ver_NtkIsDefined(model) == 0 {
				instances[model]++
			}
		}
	}
	fmt.Printf("Warning: The design contains %d undefined object types interpreted as blackboxes:\n", undefinedCount)
	for _, module := range pMan.pDesign.Modules {
		if Ver_NtkIsDefined(module) == 0 {
			fmt.Printf("%s (%d)  ", abc.Abc_NtkName(module), instances[module])
		}
	}
	fmt.Printf("\n")
}

func Ver_ParseCheckNondrivenNets(pMan *Ver_Man_t, undefs *vec.Vec_Ptr_t) int32 {
	for i := 0; i < vec.Vec_PtrSize(undefs); i++ {
		model := vec.Vec_PtrEntry(undefs, i).(*abc.Network)
		instances, _ := model.Data.(*vec.Vec_Ptr_t)
		if instances == nil {
			continue
		}
		for j := 0; j < vec.Vec_PtrSize(instances); j++ {
			box := vec.Vec_PtrEntry(instances, j).(*abc.Object)
			bundles := pMan.boxBundles[box]
			for k := vec.Vec_PtrSize(bundles) - 1; k >= 0; k-- {
				bundle, _ := vec.Vec_PtrEntry(bundles, k).(*Ver_Bundle_t)
				if bundle == nil {
					continue
				}
				for m := 0; m < vec.Vec_PtrSize(bundle.vNetsActual); m++ {
					net := vec.Vec_PtrEntry(bundle.vNetsActual, m).(*abc.Object)
					name := abc.Abc_ObjName(net)
					if abc.Abc_ObjFaninNum(net) == 0 && name != "1'b0" && name != "1'b1" {
						return 1
					}
				}
			}
		}
	}
	return 0
}

func Ver_ParseFormalNetsAreDriven(pNtk *abc.Network, formal string, bundlesByBox map[*abc.Object]*vec.Vec_Ptr_t) int32 {
	instances, _ := pNtk.Data.(*vec.Vec_Ptr_t)
	if instances == nil {
		return 0
	}
	for i := 0; i < vec.Vec_PtrSize(instances); i++ {
		box := vec.Vec_PtrEntry(instances, i).(*abc.Object)
		bundles := bundlesByBox[box]
		for j := vec.Vec_PtrSize(bundles) - 1; j >= 0; j-- {
			bundle, _ := vec.Vec_PtrEntry(bundles, j).(*Ver_Bundle_t)
			if bundle == nil || bundle.pNameFormal != formal {
				continue
			}
			for k := 0; k < vec.Vec_PtrSize(bundle.vNetsActual); k++ {
				if abc.Abc_ObjFaninNum(vec.Vec_PtrEntry(bundle.vNetsActual, k).(*abc.Object)) > 0 {
					return 1
				}
			}
			break
		}
	}
	return 0
}

func Ver_ParseGetNondrivenBundle(pNtk *abc.Network, counter int32, pMan *Ver_Man_t) *Ver_Bundle_t {
	instances, _ := pNtk.Data.(*vec.Vec_Ptr_t)
	if instances == nil {
		return nil
	}
	for i := 0; i < vec.Vec_PtrSize(instances); i++ {
		box := vec.Vec_PtrEntry(instances, i).(*abc.Object)
		bundles := pMan.boxBundles[box]
		if int(counter) >= vec.Vec_PtrSize(bundles) {
			continue
		}
		bundle, _ := vec.Vec_PtrEntry(bundles, vec.Vec_PtrSize(bundles)-1-int(counter)).(*Ver_Bundle_t)
		if bundle == nil {
			continue
		}
		for j := 0; j < vec.Vec_PtrSize(bundle.vNetsActual); j++ {
			net := vec.Vec_PtrEntry(bundle.vNetsActual, j).(*abc.Object)
			if abc.Abc_ObjFaninNum(net) == 0 && Ver_ParseFormalNetsAreDriven(pNtk, bundle.pNameFormal, pMan.boxBundles) == 0 {
				return bundle
			}
		}
	}
	return nil
}

func Ver_ParseDriveFormal(pMan *Ver_Man_t, model *abc.Network, selected *Ver_Bundle_t) int32 {
	for i := 0; i < vec.Vec_PtrSize(selected.vNetsActual); i++ {
		name := selected.pNameFormal
		if vec.Vec_PtrSize(selected.vNetsActual) != 1 {
			name = fmt.Sprintf("%s[%d]", name, i)
		}
		if abc.Abc_NtkFindNet(model, name) != nil {
			panic("formal net already exists while driving undefined box")
		}
		formalNet := abc.Abc_NtkFindOrCreateNet(model, name)
		bo := abc.Abc_NtkCreateBo(model)
		box := undefinedModelBox(model)
		po := abc.Abc_NtkCreatePo(model)
		abc.Abc_ObjAddFanin(po, formalNet)
		abc.Abc_ObjAddFanin(formalNet, bo)
		abc.Abc_ObjAddFanin(bo, box)
	}
	instances, _ := model.Data.(*vec.Vec_Ptr_t)
	for i := 0; i < vec.Vec_PtrSize(instances); i++ {
		box := vec.Vec_PtrEntry(instances, i).(*abc.Object)
		bundles := pMan.boxBundles[box]
		bundleIndex := -1
		var bundle *Ver_Bundle_t
		for j := vec.Vec_PtrSize(bundles) - 1; j >= 0; j-- {
			candidate, _ := vec.Vec_PtrEntry(bundles, j).(*Ver_Bundle_t)
			if candidate != nil && candidate.pNameFormal == selected.pNameFormal {
				bundleIndex, bundle = j, candidate
				break
			}
		}
		if bundleIndex < 0 {
			continue
		}
		for j := 0; j < vec.Vec_PtrSize(bundle.vNetsActual); j++ {
			actual := vec.Vec_PtrEntry(bundle.vNetsActual, j).(*abc.Object)
			if abc.Abc_ObjFaninNum(actual) > 0 {
				verSetError(pMan, fmt.Sprintf("Missing specification of the I/Os of undefined box %q.", abc.Abc_NtkName(model)))
				Ver_ParsePrintErrorMessage(pMan)
				return 0
			}
		}
		for j := vec.Vec_PtrSize(bundle.vNetsActual) - 1; j >= 0; j-- {
			actual := vec.Vec_PtrEntry(bundle.vNetsActual, j).(*abc.Object)
			bo := abc.Abc_NtkCreateBo(abc.Abc_ObjNtk(actual))
			abc.Abc_ObjAddFanin(bo, box)
			abc.Abc_ObjAddFanin(actual, bo)
		}
		Ver_ParseFreeBundle(bundle)
		vec.Vec_PtrWriteEntry(bundles, bundleIndex, nil)
	}
	return 1
}

func undefinedModelBox(model *abc.Network) *abc.Object {
	boxes := model.Boxes()
	if len(boxes) > 1 {
		panic("undefined model contains multiple boxes")
	}
	if len(boxes) == 1 {
		return boxes[0]
	}
	return abc.Abc_NtkCreateBlackbox(model)
}

func Ver_ParseDriveInputs(pMan *Ver_Man_t, undefs *vec.Vec_Ptr_t) int32 {
	for i := 0; i < vec.Vec_PtrSize(undefs); i++ {
		model := vec.Vec_PtrEntry(undefs, i).(*abc.Network)
		instances, _ := model.Data.(*vec.Vec_Ptr_t)
		if instances == nil || vec.Vec_PtrSize(instances) == 0 {
			continue
		}
		count := -1
		for j := 0; j < vec.Vec_PtrSize(instances); j++ {
			box := vec.Vec_PtrEntry(instances, j).(*abc.Object)
			bundles := pMan.boxBundles[box]
			current := 0
			for k := 0; k < vec.Vec_PtrSize(bundles); k++ {
				if vec.Vec_PtrEntry(bundles, k) != nil {
					current++
				}
			}
			if count == -1 {
				count = current
			} else if count != current {
				verSetError(pMan, fmt.Sprintf("The number of formal inputs (%d) is different from the expected one (%d) when instantiating network %s in box %s.", current, count, abc.Abc_NtkName(model), abc.Abc_ObjName(box)))
				Ver_ParsePrintErrorMessage(pMan)
				return 0
			}
		}
		first := vec.Vec_PtrEntry(instances, 0).(*abc.Object)
		firstBundles := pMan.boxBundles[first]
		for j := 0; j < vec.Vec_PtrSize(firstBundles); j++ {
			bundle, _ := vec.Vec_PtrEntry(firstBundles, j).(*Ver_Bundle_t)
			if bundle == nil {
				continue
			}
			for k := 0; k < vec.Vec_PtrSize(bundle.vNetsActual); k++ {
				name := bundle.pNameFormal
				if vec.Vec_PtrSize(bundle.vNetsActual) != 1 {
					name = fmt.Sprintf("%s[%d]", name, k)
				}
				if abc.Abc_NtkFindNet(model, name) != nil {
					panic("formal net already exists while driving undefined box inputs")
				}
				formalNet := abc.Abc_NtkFindOrCreateNet(model, name)
				pi := abc.Abc_NtkCreatePi(model)
				bi := abc.Abc_NtkCreateBi(model)
				box := undefinedModelBox(model)
				abc.Abc_ObjAddFanin(formalNet, pi)
				abc.Abc_ObjAddFanin(bi, formalNet)
				abc.Abc_ObjAddFanin(box, bi)
			}
		}
		for j := 0; j < vec.Vec_PtrSize(instances); j++ {
			box := vec.Vec_PtrEntry(instances, j).(*abc.Object)
			bundles := pMan.boxBundles[box]
			for k := 0; k < vec.Vec_PtrSize(bundles); k++ {
				bundle, _ := vec.Vec_PtrEntry(bundles, k).(*Ver_Bundle_t)
				if bundle == nil {
					continue
				}
				for m := vec.Vec_PtrSize(bundle.vNetsActual) - 1; m >= 0; m-- {
					actual := vec.Vec_PtrEntry(bundle.vNetsActual, m).(*abc.Object)
					bi := abc.Abc_NtkCreateBi(abc.Abc_ObjNtk(actual))
					abc.Abc_ObjAddFanin(box, bi)
					abc.Abc_ObjAddFanin(bi, actual)
				}
				Ver_ParseFreeBundle(bundle)
				vec.Vec_PtrWriteEntry(bundles, k, nil)
			}
			vec.Vec_PtrFree(bundles)
			delete(pMan.boxBundles, box)
		}
	}
	return 1
}

func Ver_ParseMaxBoxSize(undefs *vec.Vec_Ptr_t, pMan *Ver_Man_t) int32 {
	maxSize := int32(0)
	for i := 0; i < vec.Vec_PtrSize(undefs); i++ {
		model := vec.Vec_PtrEntry(undefs, i).(*abc.Network)
		instances, _ := model.Data.(*vec.Vec_Ptr_t)
		for j := 0; j < vec.Vec_PtrSize(instances); j++ {
			box := vec.Vec_PtrEntry(instances, j).(*abc.Object)
			if size := vec.Vec_PtrSize(pMan.boxBundles[box]); int32(size) > maxSize {
				maxSize = int32(size)
			}
		}
	}
	return maxSize
}

func Ver_ParsePrintLog(pMan *Ver_Man_t) {
	modules := pMan.pDesign.Modules
	name := extra.Extra_FileNameGeneric(pMan.pFileName) + ".log"
	file, err := os.Create(name)
	if err != nil {
		fmt.Printf("Cannot open hierarchy log file %q: %v\n", name, err)
		return
	}
	defer file.Close()
	instantiated := make(map[*abc.Network]int)
	for _, module := range modules {
		for _, box := range module.Boxes() {
			if abc.Abc_ObjIsLatch(box) != 0 {
				continue
			}
			if model, ok := box.Data.(*abc.Network); ok && model != nil {
				instantiated[model]++
			}
		}
	}
	fmt.Fprintf(file, "The hierarhical design %s contains %d modules:\n", pMan.pFileName, len(modules))
	oneOutput := 0
	for _, module := range modules {
		fmt.Fprintf(file, "%-50s : ", abc.Abc_NtkName(module))
		switch {
		case Ver_NtkIsDefined(module) == 0:
			fmt.Fprint(file, "undefbox")
		case abc.Abc_NtkHasBlackbox(module) != 0:
			fmt.Fprint(file, "blackbox")
		default:
			fmt.Fprint(file, "logicbox")
		}
		fmt.Fprintf(file, " instantiated %6d times ", instantiated[module])
		fmt.Fprintf(file, " pi = %4d", abc.Abc_NtkPiNum(module))
		fmt.Fprintf(file, " po = %4d", abc.Abc_NtkPoNum(module))
		fmt.Fprintf(file, " nd = %8d", abc.Abc_NtkNodeNum(module))
		fmt.Fprintf(file, " lat = %6d", abc.Abc_NtkLatchNum(module))
		fmt.Fprintf(file, " box = %6d\n", abc.Abc_NtkBoxNum(module)-abc.Abc_NtkLatchNum(module))
		if abc.Abc_NtkPoNum(module) == 1 {
			oneOutput++
		}
	}
	pct := 0.0
	if len(modules) != 0 {
		pct = 100 * float64(oneOutput) / float64(len(modules))
	}
	fmt.Fprintf(file, "The number of modules with one output = %d (%.2f %%).\n", oneOutput, pct)
	if len(modules) > 1 {
		countActual := func(module *abc.Network, box *abc.Object) int {
			bundles := pMan.boxBundles[box]
			total := 0
			for i := 0; i < vec.Vec_PtrSize(bundles); i++ {
				b := vec.Vec_PtrEntry(bundles, i).(*Ver_Bundle_t)
				total += vec.Vec_PtrSize(b.vNetsActual)
			}
			return total
		}
		count := 0
		for _, module := range modules {
			for _, box := range module.Boxes() {
				if abc.Abc_ObjIsLatch(box) != 0 {
					continue
				}
				model, ok := box.Data.(*abc.Network)
				if !ok || model == nil || Ver_NtkIsDefined(model) == 0 {
					continue
				}
				if countActual(module, box) != int(abc.Abc_NtkPiNum(model)+abc.Abc_NtkPoNum(model)) {
					count++
				}
			}
		}
		if count == 0 {
			fmt.Fprintln(file, "The outputs of all box instances are connected.")
		} else {
			fmt.Fprintf(file, "\nThe outputs of %d box instances are not connected:\n", count)
			for _, module := range modules {
				for _, box := range module.Boxes() {
					if abc.Abc_ObjIsLatch(box) != 0 {
						continue
					}
					model, ok := box.Data.(*abc.Network)
					if !ok || model == nil || Ver_NtkIsDefined(model) == 0 {
						continue
					}
					actual, formal := countActual(module, box), int(abc.Abc_NtkPiNum(model)+abc.Abc_NtkPoNum(model))
					if actual != formal {
						fmt.Fprintf(file, "In module %q instance %q of box %q has different numbers of actual/formal nets (%d/%d).\n", abc.Abc_NtkName(module), abc.Abc_ObjName(box), abc.Abc_NtkName(model), actual, formal)
					}
				}
			}
		}
	}
	fmt.Printf("Hierarchy statistics can be found in log file %q.\n", name)
}

func Ver_ParseAttachBoxes(pMan *Ver_Man_t) int32 {
	retValue := Ver_ParseConnectDefBoxes(pMan)
	if retValue < 2 {
		return retValue
	}
	Ver_ParseReportUndefBoxes(pMan)
	undefs := Ver_ParseCollectUndefBoxes(pMan)
	if vec.Vec_PtrSize(undefs) == 0 {
		panic("undefined box instances were reported but none were collected")
	}
	counter := int32(0)
	maxSize := Ver_ParseMaxBoxSize(undefs, pMan)
	for Ver_ParseCheckNondrivenNets(pMan, undefs) != 0 && counter < maxSize {
		var selected *Ver_Bundle_t
		var selectedModel *abc.Network
		for i := 0; i < vec.Vec_PtrSize(undefs); i++ {
			model := vec.Vec_PtrEntry(undefs, i).(*abc.Network)
			if selected = Ver_ParseGetNondrivenBundle(model, counter, pMan); selected != nil {
				selectedModel = model
				break
			}
		}
		if selected == nil {
			counter++
			continue
		}
		if Ver_ParseDriveFormal(pMan, selectedModel, selected) == 0 {
			return 0
		}
	}
	if Ver_ParseDriveInputs(pMan, undefs) == 0 {
		return 0
	}
	for i := 0; i < vec.Vec_PtrSize(undefs); i++ {
		model := vec.Vec_PtrEntry(undefs, i).(*abc.Network)
		instances, _ := model.Data.(*vec.Vec_Ptr_t)
		vec.Vec_PtrFree(instances)
		model.Data = nil
	}
	vec.Vec_PtrFree(undefs)
	return 1
}

/**Function*************************************************************

  Synopsis    [Creates PI terminal and net.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Creates PO terminal and net.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Create a latch with the given input/output.]

  Description [By default, the latch value is a don't-care.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Creates inverter and returns its net.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
