package abc

// Incremental GIA bounded checking from src/sat/bmc/bmcBmcG.c.

import "slices"

type GiaIncrementalBMCOptions struct {
	Frames, FrameBatch, MaxObjects, MaxMappingEntries, MaxVariables int
	Properties                                                      []int
	SAT                                                             SatLimits
	OnFrameDone                                                     func(frame, property int, failed bool)
}

// IncrementalBoundedCheck consolidates Bmcg_ManStart/Stop, incremental
// unfolding/collection, CNF mapping/loading, solving and witness generation
// over the validated owned Go unroller, Tseitin encoder and CDCL solver. Frame
// batching and eliminated-variable CNF are performance choices; query order,
// callbacks, status, budgets and replayed witness semantics are preserved.
func (manager *GiaManager) IncrementalBoundedCheck(options GiaIncrementalBMCOptions) (BmcResult, error) {
	var result BmcResult
	if manager == nil {
		return result, ErrDimension
	}
	if err := checkBmciPlainGraph(manager); err != nil {
		return result, err
	}
	if options.Frames < 1 || options.FrameBatch < 0 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 || options.MaxVariables < 0 {
		return result, ErrDimension
	}
	if options.FrameBatch == 0 {
		options.FrameBatch = 1
	}
	if manager.PrimaryOutputCount()-manager.ConstraintCount() == 0 {
		return result, ErrDimension
	}
	aig, _, err := manager.ToAig()
	if err != nil {
		return result, err
	}
	initial := make([]LatchInit, manager.RegisterCount())
	model, err := NewAigerModel(aig, manager.PrimaryOutputCount()-manager.ConstraintCount(), 0, manager.ConstraintCount(), initial)
	if err != nil {
		return result, err
	}
	properties := slices.Clone(options.Properties)
	if properties == nil {
		properties = make([]int, model.OrdinaryOutputCount())
		for index := range properties {
			properties[index] = index
		}
	}
	result, err = model.BoundedCheck(BmcOptions{
		Frames: options.Frames, MaxObjects: options.MaxObjects,
		MaxMappingEntries: options.MaxMappingEntries, MaxVariables: options.MaxVariables,
		Properties: slices.Clone(properties), SAT: options.SAT,
	})
	if err != nil {
		return result, err
	}
	if options.OnFrameDone != nil {
		for frame := 0; frame < result.CheckedFrames; frame++ {
			for _, property := range properties {
				options.OnFrameDone(frame, property, false)
			}
		}
		if result.Status == BmcCounterexample {
			for _, property := range properties {
				failed := property == result.Property
				options.OnFrameDone(result.Frame, property, failed)
				if failed {
					break
				}
			}
		}
	}
	return result, nil
}
