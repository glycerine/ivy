// Package nwk implements ABC's logic-network representation and optimization.
package nwk

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
)

type Nwk_Type_t int32

const (
	NWK_OBJ_NONE Nwk_Type_t = iota
	NWK_OBJ_CI
	NWK_OBJ_CO
	NWK_OBJ_NODE
	NWK_OBJ_LATCH
	NWK_OBJ_VOID
)

type Nwk_Man_t struct {
	Name, Spec                string
	Inputs, Outputs, Objects  []*Nwk_Obj_t
	Hop                       *hop.Hop_Man_t
	Timing                    *tim.Tim_Man_t
	Library                   *if_.If_LibLut_t
	TravID                    int32
	Latches, TruePIs, TruePOs int32
	Output                    io.Writer
	mffcRefs                  map[*Nwk_Obj_t]int
}
type Nwk_Obj_t struct {
	Manager                     *Nwk_Man_t
	Function                    *hop.Hop_Obj_t
	Copy, Next                  any
	Temp                        int32
	Type                        Nwk_Type_t
	Invert, MarkA, MarkB, MarkC bool
	PioID, ID, TravID, Level    int32
	Arrival, Required, Slack    float32
	Fanins, Fanouts             []*Nwk_Obj_t
}
type Nwk_Man_t_ = Nwk_Man_t
type Nwk_Obj_t_ = Nwk_Obj_t
type Hop_Man_t = hop.Hop_Man_t
type Hop_Obj_t = hop.Hop_Obj_t
type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Tim_Man_t = tim.Tim_Man_t
type If_LibLut_t = if_.If_LibLut_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Vec_t = vec.Vec_Vec_t

func bit(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func nodesVector(nodes []*Nwk_Obj_t) *Vec_Ptr_t {
	v := vec.Vec_PtrAlloc(len(nodes))
	for _, n := range nodes {
		vec.Vec_PtrPush(v, n)
	}
	return v
}
func vectorNodes(v *Vec_Ptr_t) []*Nwk_Obj_t {
	r := make([]*Nwk_Obj_t, vec.Vec_PtrSize(v))
	for i := range r {
		r[i] = vec.Vec_PtrEntry(v, i).(*Nwk_Obj_t)
	}
	return r
}
func writer(p *Nwk_Man_t) io.Writer {
	if p.Output == nil {
		return os.Stdout
	}
	return p.Output
}
func Nwk_ManCiNum(p *Nwk_Man_t) int32 { return int32(len(p.Inputs)) }
func Nwk_ManCoNum(p *Nwk_Man_t) int32 { return int32(len(p.Outputs)) }
func Nwk_ManNodeNum(p *Nwk_Man_t) int32 {
	var n int32
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_NODE {
			n++
		}
	}
	return n
}
func Nwk_ManLatchNum(p *Nwk_Man_t) int32 {
	var n int32
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_LATCH {
			n++
		}
	}
	return n
}
func Nwk_ManObjNumMax(p *Nwk_Man_t) int32           { return int32(len(p.Objects)) }
func Nwk_ManCi(p *Nwk_Man_t, i int32) *Nwk_Obj_t    { return p.Inputs[i] }
func Nwk_ManCo(p *Nwk_Man_t, i int32) *Nwk_Obj_t    { return p.Outputs[i] }
func Nwk_ManObj(p *Nwk_Man_t, i int32) *Nwk_Obj_t   { return p.Objects[i] }
func Nwk_ObjId(p *Nwk_Obj_t) int32                  { return p.ID }
func Nwk_ObjPioNum(p *Nwk_Obj_t) int32              { return p.PioID }
func Nwk_ObjFaninNum(p *Nwk_Obj_t) int32            { return int32(len(p.Fanins)) }
func Nwk_ObjFanoutNum(p *Nwk_Obj_t) int32           { return int32(len(p.Fanouts)) }
func Nwk_ObjFanin0(p *Nwk_Obj_t) *Nwk_Obj_t         { return p.Fanins[0] }
func Nwk_ObjFanout0(p *Nwk_Obj_t) *Nwk_Obj_t        { return p.Fanouts[0] }
func Nwk_ObjFanin(p *Nwk_Obj_t, i int32) *Nwk_Obj_t { return p.Fanins[i] }

// The C inline accidentally used index 1 for every fanout; honor the actual index.
func Nwk_ObjFanout(p *Nwk_Obj_t, i int32) *Nwk_Obj_t { return p.Fanouts[i] }
func Nwk_ObjIsNone(p *Nwk_Obj_t) int32               { return bit(p.Type == NWK_OBJ_NONE) }
func Nwk_ObjIsCi(p *Nwk_Obj_t) int32                 { return bit(p.Type == NWK_OBJ_CI) }
func Nwk_ObjIsCo(p *Nwk_Obj_t) int32                 { return bit(p.Type == NWK_OBJ_CO) }
func Nwk_ObjIsNode(p *Nwk_Obj_t) int32               { return bit(p.Type == NWK_OBJ_NODE) }
func Nwk_ObjIsLatch(p *Nwk_Obj_t) int32              { return bit(p.Type == NWK_OBJ_LATCH) }
func Nwk_ObjIsPi(p *Nwk_Obj_t) int32 {
	return bit(p.Type == NWK_OBJ_CI && (p.Manager.Timing == nil || tim.Tim_ManBoxForCi(p.Manager.Timing, int(p.PioID)) == -1))
}
func Nwk_ObjIsPo(p *Nwk_Obj_t) int32 {
	return bit(p.Type == NWK_OBJ_CO && (p.Manager.Timing == nil || tim.Tim_ManBoxForCo(p.Manager.Timing, int(p.PioID)) == -1))
}
func Nwk_ObjIsLi(p *Nwk_Obj_t) int32 {
	return bit(p.Manager.TruePOs != 0 && p.Type == NWK_OBJ_CO && p.PioID >= p.Manager.TruePOs)
}
func Nwk_ObjIsLo(p *Nwk_Obj_t) int32 {
	return bit(p.Manager.TruePIs != 0 && p.Type == NWK_OBJ_CI && p.PioID >= p.Manager.TruePIs)
}
func Nwk_ObjArrival(p *Nwk_Obj_t) float32        { return p.Arrival }
func Nwk_ObjRequired(p *Nwk_Obj_t) float32       { return p.Required }
func Nwk_ObjSlack(p *Nwk_Obj_t) float32          { return p.Slack }
func Nwk_ObjSetArrival(p *Nwk_Obj_t, v float32)  { p.Arrival = v }
func Nwk_ObjSetRequired(p *Nwk_Obj_t, v float32) { p.Required = v }
func Nwk_ObjSetSlack(p *Nwk_Obj_t, v float32)    { p.Slack = v }
func Nwk_ObjLevel(p *Nwk_Obj_t) int32            { return p.Level }
func Nwk_ObjSetLevel(p *Nwk_Obj_t, v int32)      { p.Level = v }
func Nwk_ObjSetTravId(p *Nwk_Obj_t, v int32)     { p.TravID = v }
func Nwk_ObjSetTravIdCurrent(p *Nwk_Obj_t)       { p.TravID = p.Manager.TravID }
func Nwk_ObjSetTravIdPrevious(p *Nwk_Obj_t)      { p.TravID = p.Manager.TravID - 1 }
func Nwk_ObjIsTravIdCurrent(p *Nwk_Obj_t) int32  { return bit(p.TravID == p.Manager.TravID) }
func Nwk_ObjIsTravIdPrevious(p *Nwk_Obj_t) int32 { return bit(p.TravID == p.Manager.TravID-1) }
func Nwk_ManTimeEqual(a, b, e float32) int32     { return bit(a < b+e && b < a+e) }
func Nwk_ManTimeLess(a, b, e float32) int32      { return bit(a < b+e) }
func Nwk_ManTimeMore(a, b, e float32) int32      { return bit(a+e > b) }
