package mapper

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"math"
)

const (
	MAP_FLOAT_LARGE = float32(math.MaxFloat32 / 10)
	MAP_FLOAT_SMALL = float32(1e-3)
	MAP_NO_VAR      = float32(-9999)
)

func boolToInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func cutNext(cut *Map_Cut_t) *Map_Cut_t {
	if cut == nil {
		return nil
	}
	return cut.pNext
}

func nodeDelay(manager *Map_Man_t, node *Map_Node_t) float32 {
	index := int(node.Num)
	if manager == nil || index < 0 || index >= len(manager.pNodeDelays) {
		return 0
	}
	return manager.pNodeDelays[index]
}

func incInt(value *int32) int32 { old := *value; *value++; return old }
func decInt(value *int32) int32 { *value--; return *value }

// Pointer tagging in the C mapper uses the low address bit as a phase. Go
// pointers cannot be tagged safely, so a complemented reference is an
// explicit, allocation-local wrapper around its regular object.
func Map_IsComplement(node *Map_Node_t) bool { return node != nil && node.complemented }

func Map_Regular(node *Map_Node_t) *Map_Node_t {
	if node == nil {
		return nil
	}
	if node.regular != nil {
		return node.regular
	}
	return node
}

func Map_Not(node *Map_Node_t) *Map_Node_t {
	if node == nil {
		return nil
	}
	base := Map_Regular(node)
	if node.complemented {
		return base
	}
	if base.neg == nil {
		base.neg = &Map_NodeStruct_t_{regular: base, complemented: true}
	}
	return base.neg
}

func Map_NotCond(node *Map_Node_t, condition bool) *Map_Node_t {
	if condition {
		return Map_Not(node)
	}
	return node
}

func Map_CutIsComplement(cut *Map_Cut_t) bool { return cut != nil && cut.complemented }

func Map_CutRegular(cut *Map_Cut_t) *Map_Cut_t {
	if cut == nil {
		return nil
	}
	if cut.regular != nil {
		return cut.regular
	}
	return cut
}

func Map_CutNot(cut *Map_Cut_t) *Map_Cut_t {
	if cut == nil {
		return nil
	}
	base := Map_CutRegular(cut)
	if cut.complemented {
		return base
	}
	if base.neg == nil {
		base.neg = &Map_CutStruct_t_{regular: base, complemented: true}
	}
	return base.neg
}

func Map_CutNotCond(cut *Map_Cut_t, condition bool) *Map_Cut_t {
	if condition {
		return Map_CutNot(cut)
	}
	return cut
}

func Map_NodeReadRef(node *Map_Node_t) int32 { return Map_Regular(node).nRefs }

func Map_NodeRef(node *Map_Node_t) { Map_Regular(node).nRefs++ }

// Compatibility names for the corresponding ABC packages.
type ProgressBar = extra.ProgressBar
type Mio_PinPhase_t = mio.Mio_PinPhase_t

const MAP_FULL = ^uint32(0)
const ABC_INFINITY = util.ABC_INFINITY
const MIO_PHASE_INV = mio.MIO_PHASE_INV
const MIO_PHASE_NONINV = mio.MIO_PHASE_NONINV

var Abc_Clock = util.Abc_Clock
var Abc_Print = util.Abc_Print
var Extra_ProgressBarStart = extra.Extra_ProgressBarStart
var Extra_ProgressBarUpdate = extra.Extra_ProgressBarUpdate
var Extra_ProgressBarStop = extra.Extra_ProgressBarStop
var Extra_PrintBinary = extra.Extra_PrintBinary
var Extra_ReadBinary = extra.Extra_ReadBinary
var Extra_MmFlexEntryFetch = extra.Extra_MmFlexEntryFetch
var Mio_GateReadName = mio.Mio_GateReadName
var Mio_GateReadArea = mio.Mio_GateReadArea
var Mio_GateReadPins = mio.Mio_GateReadPins
var Mio_GateReadPinNum = mio.Mio_GateReadPinNum
var Mio_PinReadNext = mio.Mio_PinReadNext
var Mio_PinReadPhase = mio.Mio_PinReadPhase
var Mio_PinReadDelayBlockRise = mio.Mio_PinReadDelayBlockRise
var Mio_PinReadDelayBlockFall = mio.Mio_PinReadDelayBlockFall
var Mio_PinReadDelayBlockMax = mio.Mio_PinReadDelayBlockMax
var Mio_DeriveTruthTable = mio.Mio_DeriveTruthTable
var Mio_GateReadProfile = mio.Mio_GateReadProfile
var Mio_GateReadProfile2 = mio.Mio_GateReadProfile2
var Mio_GateIncProfile2 = mio.Mio_GateIncProfile2
var Mio_GateDecProfile2 = mio.Mio_GateDecProfile2
var Mio_LibraryCleanProfile2 = mio.Mio_LibraryCleanProfile2
var Mio_LibraryReadName = mio.Mio_LibraryReadName
var Mio_LibraryReadGateByName = mio.Mio_LibraryReadGateByName

var Scl_ConIsRunning = scl.Scl_ConIsRunning
var Scl_ConHasOutReqs = scl.Scl_ConHasOutReqs

func Scl_ConGetOutReqFloat(i int32) float32 { return scl.Scl_ConGetOutReqFloat(int(i)) }

func Scl_ConGetInArrFloat(i int32) float32 { return scl.Scl_ConGetInArrFloat(int(i)) }
