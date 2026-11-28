// Package live is the port of ABC's src/proof/live: Biere's liveness to
// safety transformation (l2s, l2ssim, l3s), the LTL parser it uses, and
// Claessen-Sorensson's k-liveness (kcs) with its constraint, arena,
// monotonicity and n-choose-k helpers (nck).
//
// This file holds the declarations the C files share through ABC's
// headers; each other file is the port of the C file of the same name.
package live

import (
	"bufio"
	"io"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	abc "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Abc_Ntk_t = abc.Network
type Abc_Obj_t = abc.Object
type Abc_Frame_t = abcmain.Frame

// Abc_NtkFromAigPhase and Abc_NtkStrash live in base/abci (abcDar.c and
// abcStrash.c), which imports this package; base/abci sets them.
var (
	Abc_NtkFromAigPhase func(pMan *Aig_Man_t) *Abc_Ntk_t
	Abc_NtkStrash       func(pNtk *Abc_Ntk_t, fAllNodes, fCleanup, fRecord int32) *Abc_Ntk_t
)

// Abc_NtkToDar is abcDar.c's conversion (ported in base/abc).
func Abc_NtkToDar(pNtk *Abc_Ntk_t, fExors, fRegisters int32) *Aig_Man_t {
	return abc.Abc_NtkToDar(pNtk, fExors, fRegisters)
}

// NULL for an Aig_Obj_t pointer.
const aigNull = aig.InvalidEdge

// liveExit stands for C's exit(0) inside these commands: the commands
// recover it and return -1, the status that ends the script and the
// session as quit does (exit(0) ends the ABC process).
type liveExit struct{}

func exit0() { panic(liveExit{}) }

// recoverExit turns liveExit into the command status -1.
func recoverExit(status *int32) {
	if r := recover(); r != nil {
		if _, ok := r.(liveExit); !ok {
			panic(r)
		}
		*status = -1
	}
}

// pObj->pData
func aigData(p *Aig_Man_t, pObj Aig_Obj_t) Aig_Obj_t {
	return aig.Aig_ObjCopy(p, aig.Aig_Regular(pObj))
}
func aigSetData(p *Aig_Man_t, pObj, pData Aig_Obj_t) {
	aig.Aig_ObjSetCopy(p, aig.Aig_Regular(pObj), pData)
}

// aigManNodes lists the objects Aig_ManForEachNode visits, and the value
// its loop variable keeps after the loop: the last entry of p->vObjs
// (NULL when that object was deleted). Several C functions read it.
func aigManNodes(p *Aig_Man_t) (nodes []Aig_Obj_t, last Aig_Obj_t) {
	n := aig.Aig_ManObjNumMax(p)
	for i := int32(0); i < n; i++ {
		e := aig.Aig_ManObj(p, i)
		if aig.Aig_ObjIsNone(p, e) {
			continue
		}
		if aig.Aig_ObjIsNode(p, e) {
			nodes = append(nodes, e)
		}
	}
	last = aig.Aig_ManObj(p, n-1)
	if aig.Aig_ObjIsNone(p, last) {
		last = aigNull
	}
	return nodes, last
}

// aigMux is Aig_Mux (aigOper.c); aig.Aig_Mux builds its ANDs in the GCC
// build's order (else branch first).
func aigMux(p *Aig_Man_t, pC, p1, p0 Aig_Obj_t) Aig_Obj_t { return aig.Aig_Mux(p, pC, p1, p0) }

func saigPiNum(p *Aig_Man_t) int32  { return saig.Saig_ManPiNum(p) }
func saigPoNum(p *Aig_Man_t) int32  { return saig.Saig_ManPoNum(p) }
func saigRegNum(p *Aig_Man_t) int32 { return saig.Saig_ManRegNum(p) }

// driverNew is the copy of a CO's driver in the new manager:
// Aig_NotCond( Aig_ObjFanin0(pObj)->pData, Aig_ObjFaninC0(pObj) ).
func driverNew(p *Aig_Man_t, pObj Aig_Obj_t) Aig_Obj_t {
	return aig.Aig_NotCond(aigData(p, aig.Aig_ObjFanin0(p, pObj)), aig.Aig_ObjFaninC0(p, pObj))
}

// driverOld is Aig_NotCond( Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj) ).
func driverOld(p *Aig_Man_t, pObj Aig_Obj_t) Aig_Obj_t {
	return aig.Aig_NotCond(aig.Aig_ObjFanin0(p, pObj), aig.Aig_ObjFaninC0(p, pObj))
}

// copyOf is the image of an old (possibly complemented) object:
// !Aig_IsComplement(x)? Aig_Regular(x)->pData : Aig_Not(Aig_Regular(x)->pData).
func copyOf(p *Aig_Man_t, x Aig_Obj_t) Aig_Obj_t {
	if aig.Aig_IsComplement(x) == 0 {
		return aigData(p, x)
	}
	return aig.Aig_Not(aigData(p, x))
}

func frameNtk(pAbc *Abc_Frame_t) *Abc_Ntk_t {
	n, _ := abcmain.Abc_FrameReadNtk(pAbc).(*Abc_Ntk_t)
	return n
}

// stdinReader is scanf's stdin: the frame's input, or the process's.
var stdinSource io.Reader
var stdinReader *bufio.Reader

func scanfInt(pAbc *Abc_Frame_t) (int32, bool) {
	var src io.Reader = os.Stdin
	if pAbc != nil && pAbc.In != nil {
		src = pAbc.In
	}
	if stdinReader == nil || stdinSource != src {
		stdinSource, stdinReader = src, bufio.NewReader(src)
	}
	r := stdinReader
	// scanf("%d"): skip white space, optional sign, then digits
	c, err := r.ReadByte()
	for err == nil && (c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f') {
		c, err = r.ReadByte()
	}
	if err != nil {
		return 0, false
	}
	neg := false
	if c == '-' || c == '+' {
		neg = c == '-'
		if c, err = r.ReadByte(); err != nil {
			return 0, false
		}
	}
	if c < '0' || c > '9' {
		r.UnreadByte()
		return 0, false
	}
	var v int64
	for err == nil && c >= '0' && c <= '9' {
		v = v*10 + int64(c-'0')
		c, err = r.ReadByte()
	}
	if err == nil {
		r.UnreadByte()
	}
	if neg {
		v = -v
	}
	return int32(v), true
}

// The PI/LO name vectors are globals in liveness.c that keep their value
// from one command to the next for the rest of the ABC process (a
// transformation that returns its input AIG leaves them alone). Here they
// are kept per frame, the Go counterpart of the process.
const liveNamesKey = "proof/live:vecPiNames,vecLoNames"

type liveNames struct{ pis, los []string }

func loadLiveNames(pAbc *Abc_Frame_t) {
	vecPis, vecLos = nil, nil
	vecPiNames, vecLoNames = nil, nil
	if v, ok := pAbc.Resource(liveNamesKey).(liveNames); ok {
		vecPiNames, vecLoNames = v.pis, v.los
	}
}

func saveLiveNames(pAbc *Abc_Frame_t) {
	pAbc.SetResource(liveNamesKey, liveNames{vecPiNames, vecLoNames})
}
