package tim

import "fmt"

func Tim_ManInitPiArrival(p *Tim_Man_t, i int, delay float32)  { p.pCis[i].timeArr = delay }
func Tim_ManInitPoRequired(p *Tim_Man_t, i int, delay float32) { p.pCos[i].timeReq = delay }
func Tim_ManInitPiArrivalAll(p *Tim_Man_t, delay float32) {
	for i := range p.pCis {
		if p.pCis[i].iObj2Box < 0 {
			Tim_ManInitPiArrival(p, i, delay)
		}
	}
}
func Tim_ManInitPoRequiredAll(p *Tim_Man_t, delay float32) {
	for i := range p.pCos {
		if p.pCos[i].iObj2Box < 0 {
			Tim_ManSetCoRequired(p, i, delay)
		}
	}
}
func Tim_ManSetCoArrival(p *Tim_Man_t, i int, delay float32) {
	o := &p.pCos[i]
	if p.fUseTravId && o.TravId == p.nTravIds {
		panic("tim: CO arrival already updated in this traversal")
	}
	o.timeArr = delay
	o.TravId = p.nTravIds
}
func Tim_ManGetCoArrival(p *Tim_Man_t, i int) float32 { return p.pCos[i].timeArr }
func Tim_ManSetCiRequired(p *Tim_Man_t, i int, delay float32) {
	o := &p.pCis[i]
	if p.fUseTravId && o.TravId == p.nTravIds {
		panic("tim: CI required already updated in this traversal")
	}
	o.timeReq = delay
	o.TravId = p.nTravIds
}
func Tim_ManSetCoRequired(p *Tim_Man_t, i int, delay float32) {
	o := &p.pCos[i]
	if p.fUseTravId && p.nTravIds != 0 && o.TravId == p.nTravIds {
		panic("tim: CO required already updated in this traversal")
	}
	o.timeReq = delay
	o.TravId = p.nTravIds
}

// Arrival propagation computes every output when any output is requested.
func Tim_ManGetCiArrival(p *Tim_Man_t, i int) float32 {
	o := &p.pCis[i]
	if p.fUseTravId && o.TravId == p.nTravIds {
		return o.timeArr
	}
	b := Tim_ManCiBox(p, i)
	if b == nil {
		o.TravId = p.nTravIds
		return o.timeArr
	}
	table := Tim_ManBoxDelayTable(p, b.iBox)
	if table == nil {
		panic("tim: delay table missing")
	}
	b.TravId = p.nTravIds
	if p.fUseTravId {
		for j := 0; j < b.nInputs; j++ {
			if Tim_ManBoxInput(p, b, j).TravId != p.nTravIds {
				fmt.Println("Tim_ManGetCiArrival(): Input arrival times of the box are not up to date!")
			}
		}
	}
	for output := 0; output < b.nOutputs; output++ {
		best := float32(-TIM_ETERNITY)
		for input := 0; input < b.nInputs; input++ {
			delay := table[3+output*b.nInputs+input]
			if delay == -TIM_ETERNITY {
				continue
			}
			if arrival := Tim_ManBoxInput(p, b, input).timeArr + delay; arrival > best {
				best = arrival
			}
		}
		result := Tim_ManBoxOutput(p, b, output)
		result.timeArr = best
		result.TravId = p.nTravIds
	}
	return o.timeArr
}

// Required propagation takes the tightest output constraint for each input.
func Tim_ManGetCoRequired(p *Tim_Man_t, i int) float32 {
	o := &p.pCos[i]
	if p.fUseTravId && o.TravId == p.nTravIds {
		return o.timeReq
	}
	b := Tim_ManCoBox(p, i)
	if b == nil {
		o.TravId = p.nTravIds
		return o.timeReq
	}
	table := Tim_ManBoxDelayTable(p, b.iBox)
	if table == nil {
		panic("tim: delay table missing")
	}
	b.TravId = p.nTravIds
	if p.fUseTravId {
		for j := 0; j < b.nOutputs; j++ {
			if Tim_ManBoxOutput(p, b, j).TravId != p.nTravIds {
				fmt.Printf("Tim_ManGetCoRequired(): Output required times of output %d the box %d are not up to date!\n", j, b.iBox)
			}
		}
	}
	for input := 0; input < b.nInputs; input++ {
		best := float32(TIM_ETERNITY)
		for output := 0; output < b.nOutputs; output++ {
			delay := table[3+output*b.nInputs+input]
			if delay == -TIM_ETERNITY {
				continue
			}
			if required := Tim_ManBoxOutput(p, b, output).timeReq - delay; required < best {
				best = required
			}
		}
		result := Tim_ManBoxInput(p, b, input)
		result.timeReq = best
		result.TravId = p.nTravIds
	}
	return o.timeReq
}
