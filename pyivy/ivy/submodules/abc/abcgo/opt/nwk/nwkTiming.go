package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"sort"
)

const timingInfinity float32 = 1e20

func Nwk_ManCleanTiming(p *Nwk_Man_t) {
	for _, o := range p.Objects {
		if o != nil {
			o.Arrival = 0
			o.Required = timingInfinity
			o.Slack = 0
		}
	}
}
func Nwk_ManDelayTraceSortPins(o *Nwk_Obj_t, permutation []int32, arrivals []float32) {
	for i, n := range o.Fanins {
		permutation[i] = int32(i)
		arrivals[i] = n.Arrival
	}
	sort.SliceStable(permutation[:len(o.Fanins)], func(i, j int) bool { return arrivals[permutation[i]] > arrivals[permutation[j]] })
}
func Nwk_ManWhereIsPin(o, fanin *Nwk_Obj_t, permutation []int32) int32 {
	for i, j := range permutation[:len(o.Fanins)] {
		if o.Fanins[j] == fanin {
			return int32(i)
		}
	}
	return -1
}
func pinAssignment(o *Nwk_Obj_t, sorted bool) []int {
	r := make([]int, len(o.Fanins))
	for i := range r {
		r[i] = i
	}
	if sorted && o.Manager.Library != nil && o.Manager.Library.VariablePinDelays() {
		sort.SliceStable(r, func(i, j int) bool { return o.Fanins[r[i]].Arrival > o.Fanins[r[j]].Arrival })
	}
	return r
}
func pinDelay(o *Nwk_Obj_t, pin int) float32 {
	if o.Type == NWK_OBJ_CO {
		return 0
	}
	if o.Manager.Library == nil {
		return 1
	}
	if !o.Manager.Library.VariablePinDelays() {
		pin = 0
	}
	return o.Manager.Library.PinDelay(int32(len(o.Fanins)), int32(pin))
}
func Nwk_NodeComputeArrival(o *Nwk_Obj_t, sorted int32) float32 {
	if o.Type == NWK_OBJ_CI {
		return o.Arrival
	}
	if o.Type == NWK_OBJ_CO {
		return o.Fanins[0].Arrival
	}
	if len(o.Fanins) == 0 {
		return 0
	}
	arrival := float32(-timingInfinity)
	for pin, index := range pinAssignment(o, sorted != 0) {
		arrival = max(arrival, o.Fanins[index].Arrival+pinDelay(o, pin))
	}
	return arrival
}
func Nwk_NodeComputeRequired(o *Nwk_Obj_t, sorted int32) float32 {
	if o.Type == NWK_OBJ_CO {
		return o.Required
	}
	required := float32(timingInfinity)
	for _, fanout := range o.Fanouts {
		for pin, index := range pinAssignment(fanout, sorted != 0) {
			if fanout.Fanins[index] == o {
				required = min(required, fanout.Required-pinDelay(fanout, pin))
			}
		}
	}
	return required
}
func Nwk_NodePropagateRequired(o *Nwk_Obj_t, sorted int32) float32 {
	var result float32
	for pin, index := range pinAssignment(o, sorted != 0) {
		result = o.Required - pinDelay(o, pin)
		o.Fanins[index].Required = min(o.Fanins[index].Required, result)
	}
	return result
}
func Nwk_ManDelayTraceLut(p *Nwk_Man_t) float32 {
	if p.Library != nil && p.Library.LutMax < Nwk_ManGetFaninMax(p) {
		fmt.Fprintln(writer(p), "Maximum LUT size is smaller than network fanin count.")
		return -timingInfinity
	}
	Nwk_ManCleanTiming(p)
	if p.Timing != nil {
		tim.Tim_ManIncrementTravId(p.Timing)
	}
	order := vectorNodes(Nwk_ManDfs(p))
	for _, o := range order {
		arrival := Nwk_NodeComputeArrival(o, 1)
		if o.Type == NWK_OBJ_CI && p.Timing != nil {
			arrival = tim.Tim_ManGetCiArrival(p.Timing, int(o.PioID))
		}
		if o.Type == NWK_OBJ_CO && p.Timing != nil {
			tim.Tim_ManSetCoArrival(p.Timing, int(o.PioID), arrival)
		}
		o.Arrival = arrival
	}
	latest := float32(-timingInfinity)
	for _, o := range p.Outputs {
		if Nwk_ObjIsPo(o) != 0 {
			latest = max(latest, o.Arrival)
		}
	}
	if latest == -timingInfinity {
		latest = 0
	}
	if p.Timing != nil {
		tim.Tim_ManIncrementTravId(p.Timing)
		tim.Tim_ManInitPoRequiredAll(p.Timing, latest)
	} else {
		for _, o := range p.Outputs {
			o.Required = latest
		}
	}
	for _, o := range vectorNodes(Nwk_ManDfsReverse(p)) {
		switch o.Type {
		case NWK_OBJ_NODE:
			Nwk_NodePropagateRequired(o, 1)
		case NWK_OBJ_CI:
			if p.Timing != nil {
				tim.Tim_ManSetCiRequired(p.Timing, int(o.PioID), o.Required)
			}
		case NWK_OBJ_CO:
			if p.Timing != nil {
				o.Required = tim.Tim_ManGetCoRequired(p.Timing, int(o.PioID))
			}
			o.Fanins[0].Required = min(o.Fanins[0].Required, o.Required)
		}
		o.Slack = max(float32(0), o.Required-o.Arrival)
	}
	return latest
}
func Nwk_ManVerifyTiming(p *Nwk_Man_t) int32 {
	for _, o := range p.Objects {
		if o == nil {
			continue
		}
		a, r := Nwk_NodeComputeArrival(o, 1), Nwk_NodeComputeRequired(o, 1)
		if o.Type != NWK_OBJ_CI && math.Abs(float64(a-o.Arrival)) > .01 || o.Type != NWK_OBJ_CO && math.Abs(float64(r-o.Required)) > .01 {
			return 0
		}
	}
	return 1
}
func Nwk_ManDelayTracePrint(p *Nwk_Man_t) {
	for _, o := range p.Objects {
		if o != nil {
			fmt.Fprintf(writer(p), "%d: arrival=%g required=%g slack=%g\n", o.ID, o.Arrival, o.Required, o.Slack)
		}
	}
}
func Nwk_NodeUpdateAddToQueue(v *Vec_Ptr_t, o *Nwk_Obj_t, current int32, arrival int32) {
	nodes := vectorNodes(v)
	for _, n := range nodes {
		if n == o {
			return
		}
	}
	nodes = append(nodes, o)
	sort.SliceStable(nodes[int(current+1):], func(i, j int) bool {
		a, b := nodes[int(current+1)+i], nodes[int(current+1)+j]
		if arrival != 0 {
			return a.Level < b.Level
		}
		return a.Level > b.Level
	})
	vec.Vec_PtrClear(v)
	for _, n := range nodes {
		vec.Vec_PtrPush(v, n)
	}
}
func Nwk_NodeUpdateArrival(o *Nwk_Obj_t) {
	p := o.Manager
	if p.Timing != nil {
		tim.Tim_ManIncrementTravId(p.Timing)
	}
	queue := []*Nwk_Obj_t{o}
	for len(queue) > 0 {
		sort.SliceStable(queue, func(i, j int) bool { return queue[i].Level < queue[j].Level })
		n := queue[0]
		queue = queue[1:]
		arrival := Nwk_NodeComputeArrival(n, 1)
		if p.Timing != nil && n.Type == NWK_OBJ_CI {
			arrival = tim.Tim_ManGetCiArrival(p.Timing, int(n.PioID))
		}
		if absTime(arrival-n.Arrival) < .001 {
			continue
		}
		n.Arrival = arrival
		n.Slack = max(float32(0), n.Required-arrival)
		if p.Timing != nil && n.Type == NWK_OBJ_CO {
			box := tim.Tim_ManBoxForCo(p.Timing, int(n.PioID))
			if box >= 0 {
				if tim.Tim_ManIsCoTravIdCurrent(p.Timing, int(n.PioID)) {
					tim.Tim_ManSetPreviousTravIdBoxInputs(p.Timing, box)
				}
				tim.Tim_ManSetCoArrival(p.Timing, int(n.PioID), arrival)
				tim.Tim_ManSetCurrentTravIdBoxInputs(p.Timing, box)
			}
		}
		queue = append(queue, boxFanouts(n)...)
	}
}
func Nwk_NodeUpdateRequired(o *Nwk_Obj_t) {
	p := o.Manager
	if p.Timing != nil {
		tim.Tim_ManIncrementTravId(p.Timing)
	}
	queue := []*Nwk_Obj_t{}
	queue = append(queue, o.Fanins...)
	for len(queue) > 0 {
		sort.SliceStable(queue, func(i, j int) bool { return queue[i].Level > queue[j].Level })
		n := queue[0]
		queue = queue[1:]
		required := Nwk_NodeComputeRequired(n, 1)
		if p.Timing != nil && n.Type == NWK_OBJ_CO {
			required = tim.Tim_ManGetCoRequired(p.Timing, int(n.PioID))
		}
		if absTime(required-n.Required) < .001 {
			continue
		}
		n.Required = required
		n.Slack = max(float32(0), required-n.Arrival)
		if p.Timing != nil && n.Type == NWK_OBJ_CI {
			box := tim.Tim_ManBoxForCi(p.Timing, int(n.PioID))
			if box >= 0 {
				if tim.Tim_ManIsCiTravIdCurrent(p.Timing, int(n.PioID)) {
					tim.Tim_ManSetPreviousTravIdBoxOutputs(p.Timing, box)
				}
				tim.Tim_ManSetCiRequired(p.Timing, int(n.PioID), required)
				tim.Tim_ManSetCurrentTravIdBoxOutputs(p.Timing, box)
			}
		}
		queue = append(queue, boxFanins(n)...)
	}
}
func absTime(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
func Nwk_ObjLevelNew(o *Nwk_Obj_t) int32 {
	var level int32
	for _, n := range boxFanins(o) {
		level = max(level, n.Level)
	}
	if o.Type == NWK_OBJ_NODE && len(o.Fanins) > 0 || o.Type == NWK_OBJ_CI && len(boxFanins(o)) > 0 {
		level++
	}
	return level
}
func Nwk_ManUpdateLevel(o *Nwk_Obj_t) {
	queue := []*Nwk_Obj_t{o}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		level := Nwk_ObjLevelNew(n)
		if n.Level != level {
			n.Level = level
			queue = append(queue, boxFanouts(n)...)
		}
	}
}
func Nwk_ManVerifyLevel(p *Nwk_Man_t) int32 {
	for _, o := range p.Objects {
		if o != nil && o.Type != NWK_OBJ_CI && o.Level != Nwk_ObjLevelNew(o) {
			return 0
		}
	}
	return 1
}
func Nwk_ManUpdate(old, new *Nwk_Obj_t, levels *Vec_Vec_t) {
	p := old.Manager
	Nwk_ObjReplace(old, new)
	Nwk_ManUpdateLevel(new)
	Nwk_ManDelayTraceLut(p)
	if levels != nil {
		replacement := Nwk_ManLevelize(p)
		*levels = *replacement
	}
}
