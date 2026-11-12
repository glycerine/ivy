package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func boxFanins(o *Nwk_Obj_t) []*Nwk_Obj_t {
	p := o.Manager
	if o.Type != NWK_OBJ_CI || p.Timing == nil {
		return o.Fanins
	}
	box := tim.Tim_ManBoxForCi(p.Timing, int(o.PioID))
	if box < 0 {
		return nil
	}
	start, n := tim.Tim_ManBoxInputFirst(p.Timing, box), tim.Tim_ManBoxInputNum(p.Timing, box)
	return p.Outputs[start : start+n]
}
func Nwk_ManVerifyTopoOrder(p *Nwk_Man_t) int32 {
	seen := map[*Nwk_Obj_t]bool{}
	for _, o := range p.Objects {
		if o == nil {
			continue
		}
		for _, n := range boxFanins(o) {
			if !seen[n] {
				fmt.Fprintf(writer(p), "Node %d has fanin %d outside topological order.\n", o.ID, n.ID)
				return 0
			}
		}
		seen[o] = true
	}
	return 1
}
func Nwk_ManLevelBackup(p *Nwk_Man_t) int32 {
	if Nwk_ManVerifyTopoOrder(p) == 0 {
		panic("nwk: expected topological objects")
	}
	var timing *Tim_Man_t
	if p.Timing != nil {
		timing = tim.Tim_ManDup(p.Timing, 1)
		defer tim.Tim_ManStop(timing)
		tim.Tim_ManIncrementTravId(timing)
	}
	var maximum int32
	for _, o := range p.Objects {
		if o == nil {
			continue
		}
		switch o.Type {
		case NWK_OBJ_CI:
			o.Level = 0
			if timing != nil {
				o.Level = int32(tim.Tim_ManGetCiArrival(timing, int(o.PioID)))
			}
		case NWK_OBJ_CO:
			o.Level = o.Fanins[0].Level
			if timing != nil {
				tim.Tim_ManSetCoArrival(timing, int(o.PioID), float32(o.Level))
			}
			maximum = max(maximum, o.Level)
		case NWK_OBJ_NODE:
			o.Level = 0
			for _, fanin := range o.Fanins {
				o.Level = max(o.Level, fanin.Level)
			}
			o.Level++
		default:
			panic("nwk: invalid object in level traversal")
		}
	}
	return maximum
}
func Nwk_ManLevel_rec(o *Nwk_Obj_t) {
	if Nwk_ObjIsTravIdCurrent(o) != 0 {
		return
	}
	Nwk_ObjSetTravIdCurrent(o)
	var level int32
	for _, n := range boxFanins(o) {
		Nwk_ManLevel_rec(n)
		level = max(level, n.Level)
	}
	if o.Type == NWK_OBJ_NODE && len(o.Fanins) > 0 || o.Type == NWK_OBJ_CI && len(boxFanins(o)) > 0 {
		level++
	}
	o.Level = level
}
func Nwk_ManLevel(p *Nwk_Man_t) int32 {
	for _, o := range p.Objects {
		if o != nil {
			o.Level = 0
		}
	}
	Nwk_ManIncrementTravId(p)
	var level int32
	for _, o := range p.Outputs {
		if Nwk_ObjIsPo(o) != 0 {
			Nwk_ManLevel_rec(o)
			level = max(level, o.Level)
		}
	}
	for _, o := range p.Inputs {
		Nwk_ManLevel_rec(o)
		level = max(level, o.Level)
	}
	return level
}
func Nwk_ManLevelMax(p *Nwk_Man_t) int32 {
	var level int32
	for _, o := range p.Outputs {
		if Nwk_ObjIsPo(o) != 0 {
			level = max(level, o.Level)
		}
	}
	return level
}
func Nwk_ManLevelize(p *Nwk_Man_t) *Vec_Vec_t {
	v := vec.Vec_VecStart(int(Nwk_ManLevelMax(p)) + 1)
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_NODE {
			vec.Vec_VecPush(v, int(o.Level), o)
		}
	}
	return v
}
func Nwk_ManDfs_rec(o *Nwk_Obj_t, v *Vec_Ptr_t) {
	if Nwk_ObjIsTravIdCurrent(o) != 0 {
		return
	}
	Nwk_ObjSetTravIdCurrent(o)
	for _, n := range o.Fanins {
		Nwk_ManDfs_rec(n, v)
	}
	vec.Vec_PtrPush(v, o)
}
func Nwk_ManDfs(p *Nwk_Man_t) *Vec_Ptr_t {
	Nwk_ManIncrementTravId(p)
	v := vec.Vec_PtrAlloc(100)
	for _, o := range p.Objects {
		if o == nil {
			continue
		}
		if o.Type == NWK_OBJ_CI {
			Nwk_ObjSetTravIdCurrent(o)
			vec.Vec_PtrPush(v, o)
		} else if o.Type == NWK_OBJ_CO {
			Nwk_ManDfs_rec(o, v)
		}
	}
	return v
}
func Nwk_ManDfsNodes_rec(o *Nwk_Obj_t, v *Vec_Ptr_t) {
	if Nwk_ObjIsTravIdCurrent(o) != 0 {
		return
	}
	Nwk_ObjSetTravIdCurrent(o)
	if o.Type == NWK_OBJ_CI {
		return
	}
	if o.Type == NWK_OBJ_CO {
		o = o.Fanins[0]
	}
	for _, n := range o.Fanins {
		Nwk_ManDfsNodes_rec(n, v)
	}
	vec.Vec_PtrPush(v, o)
}
func Nwk_ManDfsNodes(p *Nwk_Man_t, roots []*Nwk_Obj_t, n int32) *Vec_Ptr_t {
	Nwk_ManIncrementTravId(p)
	v := vec.Vec_PtrAlloc(100)
	for _, o := range roots[:n] {
		if o.Type == NWK_OBJ_CO {
			o = o.Fanins[0]
		}
		Nwk_ManDfsNodes_rec(o, v)
	}
	return v
}
func boxFanouts(o *Nwk_Obj_t) []*Nwk_Obj_t {
	p := o.Manager
	if o.Type != NWK_OBJ_CO || p.Timing == nil {
		return o.Fanouts
	}
	box := tim.Tim_ManBoxForCo(p.Timing, int(o.PioID))
	if box < 0 {
		return nil
	}
	first, count := tim.Tim_ManBoxOutputFirst(p.Timing, box), tim.Tim_ManBoxOutputNum(p.Timing, box)
	return p.Inputs[first : first+count]
}
func Nwk_ManDfsReverse_rec(o *Nwk_Obj_t, v *Vec_Ptr_t) {
	if Nwk_ObjIsTravIdCurrent(o) != 0 {
		return
	}
	Nwk_ObjSetTravIdCurrent(o)
	for _, n := range boxFanouts(o) {
		Nwk_ManDfsReverse_rec(n, v)
	}
	vec.Vec_PtrPush(v, o)
}
func Nwk_ManDfsReverse(p *Nwk_Man_t) *Vec_Ptr_t {
	Nwk_ManIncrementTravId(p)
	v := vec.Vec_PtrAlloc(100)
	for _, o := range p.Inputs {
		if Nwk_ObjIsPi(o) != 0 {
			Nwk_ManDfsReverse_rec(o, v)
		}
	}
	for _, o := range p.Objects {
		if o != nil && o.Type == NWK_OBJ_NODE && len(o.Fanins) == 0 && Nwk_ObjIsTravIdCurrent(o) == 0 {
			vec.Vec_PtrPush(v, o)
		}
	}
	return v
}
func Nwk_ManSupportNodes_rec(o *Nwk_Obj_t, v *Vec_Ptr_t) {
	if Nwk_ObjIsTravIdCurrent(o) != 0 {
		return
	}
	Nwk_ObjSetTravIdCurrent(o)
	if o.Type == NWK_OBJ_CI {
		vec.Vec_PtrPush(v, o)
		return
	}
	for _, n := range o.Fanins {
		Nwk_ManSupportNodes_rec(n, v)
	}
}
func Nwk_ManSupportNodes(p *Nwk_Man_t, roots []*Nwk_Obj_t, n int32) *Vec_Ptr_t {
	Nwk_ManIncrementTravId(p)
	v := vec.Vec_PtrAlloc(100)
	for _, o := range roots[:n] {
		Nwk_ManSupportNodes_rec(o, v)
	}
	return v
}
func Nwk_ManSupportSum(p *Nwk_Man_t) {
	var sum int
	for _, o := range p.Outputs {
		v := Nwk_ManSupportNodes(p, []*Nwk_Obj_t{o}, 1)
		sum += vec.Vec_PtrSize(v)
	}
	fmt.Fprintf(writer(p), "Support sum = %d.\n", sum)
}

// MFFC dereferencing uses temporary counts; adjacency slices remain intact.
func Nwk_ObjDeref_rec(o *Nwk_Obj_t) int32 {
	p := o.Manager
	if p.mffcRefs == nil {
		p.mffcRefs = map[*Nwk_Obj_t]int{}
		for _, n := range p.Objects {
			if n != nil {
				p.mffcRefs[n] = len(n.Fanouts)
			}
		}
	}
	if o.Type == NWK_OBJ_CI {
		return 0
	}
	count := int32(1)
	for _, n := range o.Fanins {
		if p.mffcRefs[n] <= 0 {
			panic("nwk: MFFC reference underflow")
		}
		p.mffcRefs[n]--
		if p.mffcRefs[n] == 0 {
			count += Nwk_ObjDeref_rec(n)
		}
	}
	return count
}
func Nwk_ObjRef_rec(o *Nwk_Obj_t) int32 {
	p := o.Manager
	if p.mffcRefs == nil {
		panic("nwk: MFFC ref without deref")
	}
	var restore func(*Nwk_Obj_t) int32
	restore = func(n *Nwk_Obj_t) int32 {
		if n.Type == NWK_OBJ_CI {
			return 0
		}
		count := int32(1)
		for _, fanin := range n.Fanins {
			previous := p.mffcRefs[fanin]
			p.mffcRefs[fanin]++
			if previous == 0 {
				count += restore(fanin)
			}
		}
		return count
	}
	count := restore(o)
	p.mffcRefs = nil
	return count
}
func Nwk_ObjMffcLabel_rec(o *Nwk_Obj_t, top int32) {
	if top == 0 && (o.Type == NWK_OBJ_CI || o.Manager.mffcRefs[o] > 0) || Nwk_ObjIsTravIdCurrent(o) != 0 {
		return
	}
	Nwk_ObjSetTravIdCurrent(o)
	for _, n := range o.Fanins {
		Nwk_ObjMffcLabel_rec(n, 0)
	}
}
func Nwk_ObjMffcLabel(o *Nwk_Obj_t) int32 {
	a := Nwk_ObjDeref_rec(o)
	Nwk_ManIncrementTravId(o.Manager)
	Nwk_ObjMffcLabel_rec(o, 1)
	b := Nwk_ObjRef_rec(o)
	o.Manager.mffcRefs = nil
	if a != b {
		panic("nwk: MFFC imbalance")
	}
	return a
}
