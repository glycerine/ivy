package dch

import (
	"fmt"
	"slices"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// Preserve public graph metadata without sharing mutable interface arrays.
// Timing/history objects use explicit clone contracts because their shape is
// owned by callers.
func copyMetadata(destination, source *aig.Graph) error {
	destination.Name, destination.Spec = source.Name, source.Spec
	destination.Time2Quit = source.Time2Quit
	destination.RetimingChanged = source.RetimingChanged
	destination.Assertions, destination.BarrierBuffers = source.Assertions, source.BarrierBuffers
	destination.InputNumbers = slices.Clone(source.InputNumbers)
	destination.PartitionInputs = slices.Clone(source.PartitionInputs)
	destination.PartitionOutputs = slices.Clone(source.PartitionOutputs)
	destination.FlopNumbers = slices.Clone(source.FlopNumbers)
	destination.FlopReprs = slices.Clone(source.FlopReprs)
	destination.OneHotGroups = cloneIntGroups(source.OneHotGroups)
	destination.OneHotInputGroups = cloneIntGroups(source.OneHotInputGroups)
	destination.ClockDomains = slices.Clone(source.ClockDomains)
	for i, domain := range source.ClockDomains {
		destination.ClockDomains[i] = slices.Clone(domain)
	}
	if source.Timing != nil {
		clone, ok := source.Timing.(interface{ CloneTiming() any })
		if !ok {
			return fmt.Errorf("dch: timing metadata must implement CloneTiming")
		}
		destination.Timing = clone.CloneTiming()
	}
	if source.History != nil {
		clone, ok := source.History.(interface{ CloneHistory() any })
		if !ok {
			return fmt.Errorf("dch: history metadata must implement CloneHistory")
		}
		destination.History = clone.CloneHistory()
	}
	aig.Aig_ManSetRegNum(destination, aig.Aig_ManRegNum(source))
	aig.Aig_ManSetConstrNum(destination, aig.Aig_ManConstrNum(source))
	return nil
}

func cloneIntGroups(groups [][]int32) [][]int32 {
	if groups == nil {
		return nil
	}
	result := make([][]int32, len(groups))
	for i, group := range groups {
		result[i] = slices.Clone(group)
	}
	return result
}
