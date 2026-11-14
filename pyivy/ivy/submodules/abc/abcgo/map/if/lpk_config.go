package if_

// If_ManSetLpkPars initializes the zero-filled IF configuration used by
// Lpk_IfManStart in src/opt/lpk/lpkCore.c. LPK allocates its arrival array
// after starting the IF manager.
func If_ManSetLpkPars(p *If_Par_t, lutSize int32) {
	*p = If_Par_t_{NLutSize: lutSize, NCutsMax: 16, DelayTarget: -1, Epsilon: 0.005, FArea: 1, FTruth: 1}
}

// SetArrivalTime updates one existing arrival slot without changing the rest.
// SetArrivalTimes supplies the array first; indexing an unallocated or short
// array panics, as with an ordinary Go slice.
func (p *If_Par_t_) SetArrivalTime(index int, arrival float32) { p.pTimesArr[index] = arrival }

// SetAreaOnly selects the area-only mapping mode used by LPK's mapping pass.
func (p *If_Par_t_) SetAreaOnly(enabled bool) {
	p.fAreaOnly = 0
	if enabled {
		p.fAreaOnly = 1
	}
}
