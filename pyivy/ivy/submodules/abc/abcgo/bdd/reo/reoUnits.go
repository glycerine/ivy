package reo

// Port of src/bdd/reo/reoUnits.c: procedures which support internal data structures.
//
// Units live in one slice grown unitChunk entries at a time and are named
// by index. Index 0 is not handed out: uref 0 is NULL. C stops with
// "Memory manager ran out of memory!" (and then dereferences NULL) after
// nMemChunksAlloc chunks; Go allocates further chunks instead.

// reoUnitsGetNextUnit extracts the next unit from the free unit list.
func reoUnitsGetNextUnit(p *reo_man) uref {
	// check there are stil units to extract
	if p.pUnitFreeList == 0 {
		reoUnitsAddToFreeUnitList(p)
	}
	// extract the next unit from the linked list
	pUnit := p.pUnitFreeList
	p.pUnitFreeList = p.U(pUnit).Next
	p.nUnitsUsed++
	return pUnit
}

// reoUnitsRecycleUnit returns the unit to the free unit list.
func reoUnitsRecycleUnit(p *reo_man, pUnit uref) {
	p.U(pUnit).Next = p.pUnitFreeList
	p.pUnitFreeList = pUnit
	p.nUnitsUsed--
}

// reoUnitsRecycleUnitList returns the list of units to the free unit list.
func reoUnitsRecycleUnitList(p *reo_man, pPlane *reo_plane) {
	var pTail uref
	if pPlane.pHead == 0 {
		return
	}
	// find the tail
	for pUnit := pPlane.pHead; pUnit != 0; pUnit = p.U(pUnit).Next {
		pTail = pUnit
	}
	p.U(pTail).Next = p.pUnitFreeList
	p.pUnitFreeList = pPlane.pHead
	*pPlane = reo_plane{}
}

// reoUnitsStopDispenser stops the unit dispenser.
func reoUnitsStopDispenser(p *reo_man) {
	p.units = nil
	p.nMemChunks = 0
	p.pUnitFreeList = 0
}

// reoUnitsAddUnitToPlane adds one unit to the list of units which constitutes the plane.
func reoUnitsAddUnitToPlane(p *reo_man, pPlane *reo_plane, pUnit uref) {
	if pPlane.pHead == 0 {
		pPlane.pHead = pUnit
		p.U(pUnit).Next = 0
	} else {
		p.U(pUnit).Next = pPlane.pHead
		pPlane.pHead = pUnit
	}
	pPlane.statsNodes++
}

// reoUnitsAddToFreeUnitList allocates the next chunk of units.
func reoUnitsAddToFreeUnitList(p *reo_man) {
	assert(p.pUnitFreeList == 0, "p->pUnitFreeList == NULL")
	// allocate the next chunk
	base := len(p.units)
	p.units = append(p.units, make([]reo_unit, unitChunk)...)
	chunk := p.units[base:]
	first := 0
	if base == 0 {
		first = 1 // unit 0 is NULL
	}
	// split chunks into list-connected units
	for c := first; c < unitChunk-1; c++ {
		chunk[c].Next = uref(base+c+1) << 1
	}
	// set the last pointer to NULL
	chunk[unitChunk-1].Next = 0
	p.nMemChunks++
	p.pUnitFreeList = uref(base+first) << 1
}
