package fpga

import (
	"fmt"
	"time"
)

func (p *Fpga_Man_t) Map() (err error) {
	start := time.Now()
	defer func() { p.Stats.Total = time.Since(start) }()
	p.LastError = nil
	p.Stats = MappingStatistics{}
	if len(p.InputArrivals) != len(p.Inputs) {
		return fmt.Errorf("fpga: input arrival count mismatch")
	}
	for _, arrival := range p.InputArrivals {
		if !finite(arrival) {
			return fmt.Errorf("fpga: non-finite input arrival")
		}
	}
	if interrupted(p) {
		return p.LastError
	}
	if e := p.Library.Validate(); e != nil {
		return e
	}
	if p.Library.LutMax > FPGA_MAX_LEAVES {
		return fmt.Errorf("fpga: maximum cut size is %d", FPGA_MAX_LEAVES)
	}
	if p.Latches < 0 || int(p.Latches) > len(p.Outputs) {
		return fmt.Errorf("fpga: invalid latch count")
	}
	for i, e := range p.Outputs {
		if e == nil || Fpga_Regular(e).manager != p {
			return fmt.Errorf("fpga: invalid output %d", i)
		}
	}
	for _, n := range p.Nodes {
		n.Cuts = nil
		n.Best = nil
		n.EstimatedFanouts = -1
		n.Refs = int32(len(n.fanouts))
		n.Required = FPGA_FLOAT_LARGE
	}
	p.Ands = Fpga_MappingDfs(p, 1)
	Fpga_MappingSetChoiceLevels(p)
	startCuts := time.Now()
	Fpga_MappingCuts(p)
	p.Stats.CutTime = time.Since(startCuts)
	startMatch := time.Now()
	if Fpga_MappingMatches(p, 1) == 0 {
		if p.LastError != nil {
			return p.LastError
		}
		return fmt.Errorf("fpga: no feasible LUT cut")
	}
	p.Stats.MatchTime = time.Since(startMatch)
	if Fpga_MappingPostProcess(p) == 0 {
		return p.LastError
	}

	Fpga_TimeComputeRequiredGlobal(p, 0)
	Fpga_MappingSetUsedCuts(p)
	p.Stats.Nodes = len(p.Ands.Nodes)
	p.Stats.LUTs = len(p.Mapping.Nodes)
	p.Stats.Area = p.AreaGlobal
	p.Stats.Delay = Fpga_TimeComputeArrivalMax(p)
	p.Stats.Switching = Fpga_MappingGetSwitching(p, p.Mapping)
	return nil
}
func Fpga_Mapping(p *Fpga_Man_t) int32 { p.LastError = p.Map(); return boolInt(p.LastError == nil) }

// Fpga_MappingPostProcess performs area-flow and exact area/switching recovery
// on an existing delay-oriented cover.
func Fpga_MappingPostProcess(p *Fpga_Man_t) int32 {
	Fpga_MappingSetRefsAndArea(p)
	if p.AreaRecovery {
		startRecovery := time.Now()
		Fpga_TimeComputeRequiredGlobal(p, 1)

		if Fpga_MappingMatches(p, 0) == 0 {
			return 0
		}
		Fpga_MappingSetRefsAndArea(p)
		Fpga_TimeComputeRequiredGlobal(p, 0)
		if p.Switching {
			if Fpga_MappingMatchesSwitch(p) == 0 {
				return 0
			}
		} else if Fpga_MappingMatchesArea(p) == 0 {
			return 0
		}
		Fpga_MappingSetRefsAndArea(p)
		Fpga_TimePropagateArrival(p)
		p.Stats.RecoveryTime = time.Since(startRecovery)
	}
	return 1
}
