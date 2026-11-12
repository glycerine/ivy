package sfm

import "github.com/glycerine/abcgo/map/mio"

// sfmGateFromMio preserves the pin delays needed by later timing updates.
func sfmGateFromMio(g *mio.Mio_Gate_t) *Gate {
	cell := &Gate{Cell: g, Name: mio.Mio_GateReadName(g), Inputs: int(mio.Mio_GateReadPinNum(g)), Truth: []uint64{mio.Mio_GateReadTruth(g)}, Area: int(float32(mio.Mio_GateReadArea(g)) * 1000)}
	for pin := mio.Mio_GateReadPins(g); pin != nil; pin = mio.Mio_PinReadNext(pin) {
		cell.Pins = append(cell.Pins, Pin{Phase: int(mio.Mio_PinReadPhase(pin)), Rise: int(float32(mio.Mio_PinReadDelayBlockRise(pin)) * 1000), Fall: int(float32(mio.Mio_PinReadDelayBlockFall(pin)) * 1000), Load: mio.Mio_PinReadInputLoad(pin)})
	}
	return cell
}

// sfmInsertAreaMatch applies an already-validated mapped replacement. Original
// internal IDs remain stable; POs move to make room for the new mapped nodes.
func sfmInsertAreaMatch(p *Sfm_Ntk_t, pivot int, gates []int32, fanins [][]int32, order, cone []int, handles []*mio.Mio_Gate_t) bool {
	if len(gates) == 0 || len(gates) > 2 || len(gates) != len(fanins) || pivot < p.Pis || pivot >= len(p.Fanins)-p.Pos || p.Deleted[pivot] {
		return false
	}
	removed := make(map[int]bool, len(cone))
	for _, id := range cone {
		if id < p.Pis || id >= len(p.Fanins)-p.Pos || p.Deleted[id] || p.Fixed[id] {
			return false
		}
		removed[id] = true
	}
	if !removed[pivot] {
		return false
	}
	for id := range removed {
		if id != pivot {
			for _, out := range p.Fanouts[id] {
				if !removed[out] {
					return false
				}
			}
		}
	}
	cells := make([]*Gate, len(gates))
	for i, id := range gates {
		if id < 0 || int(id) >= len(handles) || handles[id] == nil {
			return false
		}
		g := handles[id]
		if int(mio.Mio_GateReadPinNum(g)) != len(fanins[i]) {
			return false
		}
		cells[i] = sfmGateFromMio(g)
		if len(cells[i].Pins) != cells[i].Inputs {
			return false
		}
		for _, f := range fanins[i] {
			if f == -1 && i == 1 {
				continue
			}
			if f < 0 || int(f) >= len(order) {
				return false
			}
			id := order[f]
			if id < 0 || id >= len(p.Fanins)-p.Pos || p.Deleted[id] || removed[id] {
				return false
			}
		}
	}
	newID := -1
	for i, pins := range fanins {
		fs := make([]int, len(pins))
		for j, f := range pins {
			if f == -1 {
				fs[j] = newID
			} else {
				fs[j] = order[f]
			}
		}
		newID = p.insertMapped(fs, cells[i])
	}
	patchAll(p, pivot, newID)
	for id := range removed {
		p.Deleted[id] = true
		p.Fanins[id] = nil
		p.Mapping[id] = nil
	}
	if err := p.refresh(); err != nil {
		panic(err)
	}
	p.cnfs = Sfm_CreateCnf(p)
	return true
}
