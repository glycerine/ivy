// Genlib mapping library data, ported from mioInt.h.
package mio

import "github.com/glycerine/abcgo/misc/vec"

type word = uint64
type Vec_Int_t = vec.Vec_Int_t
type Vec_Wrd_t = vec.Vec_Wrd_t
type Vec_Str_t = vec.Vec_Str_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Wec_t = vec.Vec_Wec_t
type Vec_Mem_t = vec.Vec_Mem_t

const (
	MIO_STRING_GATE    = "GATE"
	MIO_STRING_LATCH   = "LATCH"
	MIO_STRING_PIN     = "PIN"
	MIO_STRING_NONINV  = "NONINV"
	MIO_STRING_INV     = "INV"
	MIO_STRING_UNKNOWN = "UNKNOWN"
	MIO_STRING_CONST0  = "CONST0"
	MIO_STRING_CONST1  = "CONST1"
	MIO_FULL           = ^uint32(0)
)

func MIO_MASK(n int32) uint32 {
	if n <= 0 {
		return 0
	}
	if n >= 32 {
		return MIO_FULL
	}
	return MIO_FULL >> uint32(32-n)
}

type Mio_Library_t struct {
	matchingProvider                                                                       MatchingProvider
	pName                                                                                  string
	nGates                                                                                 int32
	ppGates0, ppGatesName                                                                  []*Mio_Gate_t
	pGates, pGate0, pGate1, pGateBuf, pGateInv, pGateNand2, pGateAnd2, pGateNor2, pGateOr2 *Mio_Gate_t
	tName2Gate                                                                             map[string]*Mio_Gate_t
	fPinFilter, fPinPerm, fPinQuick                                                        int32
	vTtMem                                                                                 *Vec_Mem_t
	vTt2Match                                                                              *Vec_Wec_t
	pCells                                                                                 []Mio_Cell2_t
	nCells                                                                                 int32
	vNames                                                                                 *Vec_Ptr_t
	vTruths                                                                                *Vec_Wrd_t
	vTt2Match4, vConfigs                                                                   *Vec_Int_t
	vTtMem2                                                                                [3]*Vec_Mem_t
	vTt2Match2                                                                             [3]*Vec_Int_t
}
type Mio_Gate_t struct {
	pName                            string
	dArea                            float64
	pForm                            string
	pPins                            *Mio_Pin_t
	pOutName                         string
	pLib                             *Mio_Library_t
	pNext, pTwin                     *Mio_Gate_t
	Cell, nInputs, Profile, Profile2 int32
	dDelayMax                        float64
	pSop                             string
	vExpr                            *Vec_Int_t
	uTruth                           uint64
	pTruth                           []uint64
	Value                            int32
}
type Mio_Pin_t struct {
	pName                                                                                                      string
	Phase                                                                                                      Mio_PinPhase_t
	dLoadInput, dLoadMax, dDelayBlockRise, dDelayFanoutRise, dDelayBlockFall, dDelayFanoutFall, dDelayBlockMax float64
	pNext                                                                                                      *Mio_Pin_t
}
type Mio_LibraryStruct_t_ = Mio_Library_t
type Mio_GateStruct_t_ = Mio_Gate_t
type Mio_PinStruct_t_ = Mio_Pin_t
