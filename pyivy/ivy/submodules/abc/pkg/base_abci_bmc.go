package abc

// IVY timeframe/FRAIG integration from src/base/abci/abcBmc.c.

type ABCIvyBMCOptions struct {
	Frames, MaxFrameObjects, MaxMappingEntries int
	Initialize                                 bool
	MaxReportEntries                           int
	Functional                                 ABCFunctionalOptions
	Fraig                                      IvyFraigOptions
	Stop                                       func() bool
}

type ABCIvyBMCFrameStats struct {
	Frame, StrashedNodes, FraigNodes int
}

// ABCIvyBMCResult owns the combinational timeframe expansion and its FRAIG.
// Structured counts replace the source's unconditional stdout diagnostics.
type ABCIvyBMCResult struct {
	Frames         *AigFrameExpansion
	Fraig          IvyFraigResult
	FrameNodeCount int
	FraigNodeCount int
	PerFrame       []ABCIvyBMCFrameStats
}

func aigToIvyForFraig(graph *AigManager, maximum, maxWork int, stop func() bool) (*IvyManager, []IvyEdge, []bool, error) {
	if graph == nil || maximum < 0 || maxWork < 0 {
		return nil, nil, nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, nil, nil, err
	}
	if maximum == 0 {
		maximum = 1 << 20
	}
	if maxWork == 0 {
		maxWork = 1 << 24
	}
	if len(graph.objects) > maximum {
		return nil, nil, nil, ErrLimit
	}
	result, err := NewIvyManager(IvyOptions{MaxObjects: maximum, MaxWork: maxWork, Cancelled: stop})
	if err != nil {
		return nil, nil, nil, err
	}
	mapping := make([]IvyEdge, len(graph.objects))
	ready := make([]bool, len(graph.objects))
	mapping[0], ready[0] = IvyConst1, true
	read := func(literal AigLit) (IvyEdge, error) {
		if !literal.Valid() || uint64(literal.ID()) >= uint64(len(mapping)) || !ready[literal.ID()] {
			return 0, ErrInvariant
		}
		return mapping[literal.ID()].NotIf(literal.Complemented()), nil
	}
	work := 0
	for id, object := range graph.objects {
		if id == 0 || object.Type == AigNone {
			continue
		}
		work++
		if work > maxWork {
			return nil, nil, nil, ErrLimit
		}
		if stop != nil && stop() {
			return nil, nil, nil, ErrCancelled
		}
		var edge IvyEdge
		switch object.Type {
		case AigCI:
			edge, err = result.CreatePI()
		case AigAND, AigXOR:
			var a, b IvyEdge
			a, err = read(object.Fanin0)
			if err == nil {
				b, err = read(object.Fanin1)
			}
			if err == nil && object.Type == AigAND {
				edge, err = result.And(a, b)
			} else if err == nil {
				edge, err = result.Xor(a, b)
			}
		case AigBuffer, AigCO:
			edge, err = read(object.Fanin0)
			if err == nil && object.Type == AigBuffer {
				edge, err = result.CreateBuffer(edge)
			} else if err == nil {
				_, err = result.CreatePO(edge)
			}
		default:
			err = ErrUnsupported
		}
		if err != nil {
			return nil, nil, nil, err
		}
		mapping[id], ready[id] = edge, true
	}
	if err = result.Check(); err != nil {
		return nil, nil, nil, err
	}
	return result, mapping, ready, nil
}

// ABCIvyBMC converts a concrete ABC network, constructs the requested
// timeframes, and performs full-support IVY FRAIG reduction. It preserves the
// source routine's behavior: this is a bounded-frame reduction/reporting pass,
// not a property-result API. Initialize selects constant-zero frame-zero latch
// state; otherwise initial state is symbolic and final latch inputs are emitted.
func ABCIvyBMC(network *ABCNetwork, options ABCIvyBMCOptions) (ABCIvyBMCResult, error) {
	var result ABCIvyBMCResult
	if network == nil || options.Frames <= 0 || options.MaxFrameObjects < 0 || options.MaxMappingEntries < 0 || options.MaxReportEntries < 0 {
		return result, ErrDimension
	}
	functional := options.Functional
	if functional.Stop == nil {
		functional.Stop = options.Stop
	}
	converted, err := ABCNetworkToAIG(network, functional)
	if err != nil {
		return result, err
	}
	expansion, err := converted.Graph.ExpandFrames(AigFramesOptions{
		Frames: options.Frames, Initialize: options.Initialize, Outputs: true, Registers: !options.Initialize,
		MaxObjects: options.MaxFrameObjects, MaxMappingEntries: options.MaxMappingEntries, Stop: options.Stop,
	})
	if err != nil {
		return result, err
	}
	fraigOptions := options.Fraig
	if fraigOptions.Stop == nil {
		fraigOptions.Stop = options.Stop
	}
	if fraigOptions.Parameters == (IvyFraigParameters{}) {
		fraigOptions.Parameters = DefaultIvyFraigParameters()
	}
	// Abc_NtkBmc sets infinite node SAT effort, disables proof mode and printing.
	fraigOptions.Parameters.NodeBacktrackLimit = -1
	fraigOptions.Parameters.Prove = false
	fraigOptions.Parameters.Verbose = false
	frameIvy, frameToIvy, frameReady, err := aigToIvyForFraig(expansion.Graph(), fraigOptions.MaxObjects, fraigOptions.MaxWork, options.Stop)
	if err != nil {
		return result, err
	}
	fraig, err := frameIvy.Fraig(fraigOptions)
	if err != nil {
		return result, err
	}
	maxReport := options.MaxReportEntries
	if maxReport == 0 {
		maxReport = 1 << 20
	}
	if options.Frames > maxReport {
		return result, ErrLimit
	}
	perFrame := make([]ABCIvyBMCFrameStats, options.Frames)
	for frame := 0; frame < options.Frames; frame++ {
		if options.Stop != nil && options.Stop() {
			return ABCIvyBMCResult{}, ErrCancelled
		}
		strashed := make(map[uint32]bool)
		reduced := make(map[uint32]bool)
		for id, object := range converted.Graph.objects {
			if object.Type != AigAND {
				continue
			}
			literal, mapErr := expansion.Object(AigID(id), frame)
			if mapErr != nil {
				continue
			}
			frameID := literal.ID()
			if uint64(frameID) >= uint64(len(frameToIvy)) || !frameReady[frameID] {
				return ABCIvyBMCResult{}, ErrInvariant
			}
			strashed[uint32(frameID)] = true
			ivyID := frameToIvy[frameID].ID()
			if uint64(ivyID) >= uint64(len(fraig.Mapping)) || !fraig.MappingValid[ivyID] {
				continue
			}
			targetID := fraig.Mapping[ivyID].ID()
			if targetID != 0 && uint64(targetID) < uint64(len(fraig.Graph.objects)) {
				type_ := fraig.Graph.objects[targetID].Type
				if type_ == IvyAND || type_ == IvyEXOR || type_ == IvyBuffer {
					reduced[targetID] = true
				}
			}
		}
		perFrame[frame] = ABCIvyBMCFrameStats{Frame: frame, StrashedNodes: len(strashed), FraigNodes: len(reduced)}
	}
	return ABCIvyBMCResult{
		Frames: expansion, Fraig: fraig, FrameNodeCount: expansion.Graph().AndCount(),
		FraigNodeCount: fraig.Graph.NodeCount(), PerFrame: perFrame,
	}, nil
}
