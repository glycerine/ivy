// Wire capacitance and pin-load accounting from sclLoad.c.
package scl

import (
	"fmt"
	"math"
	"sort"
)

func sclFinite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
func Abc_SclFindWireCaps(model *SC_WireLoad, maxFanout int) ([]float32, error) {
	if model == nil || maxFanout < 0 || len(model.Fanout) != len(model.Length) || !sclFinite(model.Cap) || !sclFinite(model.Slope) || model.Cap < 0 || model.Slope < 0 {
		return nil, fmt.Errorf("SCL: invalid wire-load model")
	}
	type point struct {
		fanout int
		cap    float32
	}
	points := make([]point, len(model.Fanout))
	seen := make(map[int]bool)
	for i, count := range model.Fanout {
		if count < 0 || seen[int(count)] || !sclFinite(model.Length[i]) || model.Length[i] < 0 {
			return nil, fmt.Errorf("SCL: invalid wire-load table")
		}
		seen[int(count)] = true
		points[i] = point{int(count), model.Length[i] * model.Cap}
		maxFanout = max(maxFanout, int(count))
	}
	sort.Slice(points, func(i, j int) bool { return points[i].fanout < points[j].fanout })
	caps := make([]float32, maxFanout+1)
	if len(points) == 0 {
		return caps, nil
	}
	// Fanout zero has zero wire length unless explicitly specified.
	previous := point{}
	for _, current := range points {
		if current.fanout > previous.fanout {
			for count := previous.fanout + 1; count <= current.fanout; count++ {
				caps[count] = previous.cap + (current.cap-previous.cap)*float32(count-previous.fanout)/float32(current.fanout-previous.fanout)
			}
		} else {
			caps[current.fanout] = current.cap
		}
		previous = current
	}
	for count := previous.fanout + 1; count <= maxFanout; count++ {
		caps[count] = previous.cap + model.Cap*model.Slope*float32(count-previous.fanout)
	}
	return caps, nil
}
func Abc_SclFindWireLoad(caps []float32, fans int) float32 {
	if len(caps) == 0 {
		return 0
	}
	if fans < 0 {
		panic("SCL: negative fanout")
	}
	return caps[min(fans, len(caps)-1)]
}
func Abc_SclAddWireLoad(p *SC_Man, id int, subtract int32) {
	load := Abc_SclFindWireLoad(p.WireCaps, len(p.Network.SclFanouts(id)))
	if subtract != 0 {
		load = -load
	}
	p.Loads[id].Rise += load
	p.Loads[id].Fall += load
}
func Abc_SclComputeLoad(p *SC_Man) error {
	outputs := make(map[int]bool)
	for _, id := range p.Network.SclOutputs() {
		outputs[id] = true
	}
	for _, id := range p.Network.SclObjectIDs() {
		if !outputs[id] {
			p.Loads[id] = SC_Pair{}
		}
	}
	for _, id := range sclMappedNodes(p.Network) {
		cell := Abc_SclObjCell(p, id)
		if cell == nil {
			return fmt.Errorf("SCL: object %d lacks a cell", id)
		}
		fans := p.Network.SclFanins(id)
		if len(fans) != cell.Inputs {
			return fmt.Errorf("SCL: object %d cell fanin mismatch", id)
		}
		for pin, fanin := range fans {
			p.Loads[fanin].Rise += cell.Pins[pin].RiseCap
			p.Loads[fanin].Fall += cell.Pins[pin].FallCap
		}
	}
	for _, id := range p.Network.SclOutputs() {
		fans := p.Network.SclFanins(id)
		if len(fans) != 1 {
			return fmt.Errorf("SCL: output %d needs one driver", id)
		}
		SC_PairAdd(&p.Loads[fans[0]], &p.Loads[id])
	}
	// Barriers represent ideal wires, not physical library gates.
	for i := len(p.order) - 1; i >= 0; i-- {
		id := p.order[i]
		if p.Network.SclIsBarrier(id) {
			fans := p.Network.SclFanins(id)
			if len(fans) != 1 {
				return fmt.Errorf("SCL: barrier %d needs one driver", id)
			}
			SC_PairAdd(&p.Loads[fans[0]], &p.Loads[id])
		}
	}
	if p.WireLoad != nil {
		maxFanout := 0
		for _, id := range p.Network.SclObjectIDs() {
			maxFanout = max(maxFanout, len(p.Network.SclFanouts(id)))
		}
		caps, err := Abc_SclFindWireCaps(p.WireLoad, maxFanout)
		if err != nil {
			return err
		}
		p.WireCaps = caps
		for _, id := range sclMappedNodes(p.Network) {
			Abc_SclAddWireLoad(p, id, 0)
		}
		for _, id := range p.Network.SclInputs() {
			Abc_SclAddWireLoad(p, id, 0)
		}
	}
	var total float64
	count := 0
	for _, id := range sclMappedNodes(p.Network) {
		total += float64(SC_PairAve(&p.Loads[id]))
		count++
	}
	for _, id := range p.Network.SclInputs() {
		total += float64(SC_PairAve(&p.Loads[id]))
		count++
	}
	p.EstLoadAve = 0
	if count > 0 {
		p.EstLoadAve = float32(total / float64(count))
	}
	return nil
}
func Abc_SclUpdateLoad(p *SC_Man, id int, old, new *SC_Cell) error {
	fans := p.Network.SclFanins(id)
	if old == nil || new == nil || old.Inputs != len(fans) || new.Inputs != len(fans) {
		return fmt.Errorf("SCL: resizing changes gate arity")
	}
	for i, fanin := range fans {
		p.Loads[fanin].Rise += new.Pins[i].RiseCap - old.Pins[i].RiseCap
		p.Loads[fanin].Fall += new.Pins[i].FallCap - old.Pins[i].FallCap
	}
	return nil
}
func Abc_SclUpdateLoadSplit(p *SC_Man, buffer, fanout int) error {
	fans := p.Network.SclFanins(buffer)
	cell := Abc_SclObjCell(p, fanout)
	if len(fans) != 1 || cell == nil {
		return fmt.Errorf("SCL: invalid buffer bypass")
	}
	matches := 0
	for pin, fanin := range p.Network.SclFanins(fanout) {
		if fanin != buffer {
			continue
		}
		matches++
		cap := cell.Pins[pin]
		p.Loads[buffer].Rise -= cap.RiseCap
		p.Loads[buffer].Fall -= cap.FallCap
		p.Loads[fans[0]].Rise += cap.RiseCap
		p.Loads[fans[0]].Fall += cap.FallCap
	}
	if matches == 0 {
		return fmt.Errorf("SCL: buffer does not drive fanout")
	}
	return nil
}
