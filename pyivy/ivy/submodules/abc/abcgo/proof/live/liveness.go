package live

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"strings"
)

const (
	FULL_BIERE_MODE int32 = iota
	IGNORE_LIVENESS_KEEP_SAFETY_MODE
	IGNORE_SAFETY_KEEP_LIVENESS_MODE
	IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE
	FULL_BIERE_ONE_LOOP_MODE
)

type TransformOptions struct {
	Mode                                                    int32
	Liveness, Fairness, SafetyAssertions, SafetyAssumptions []aig.Edge
	ShadowRegisters                                         []int32 // nil means every original register
	Restartable                                             bool
	LTL                                                     []string
}

func liveSignal(p *aig.Graph, e aig.Edge) aig.Edge {
	node := p.Node(aig.Aig_Regular(e))
	if node.Kind == aig.Output {
		return node.Fanins[0] ^ aig.Edge(e&1)
	}
	return e
}
func (c *graphCopy) signal(e aig.Edge) aig.Edge { return c.mapEdge(liveSignal(c.source, e)) }

// LivenessToSafety implements the Biere saved-state/seen-event construction.
// All generated outputs use safety convention: true means a violation.
func LivenessToSafety(network *Network, options TransformOptions) (*Network, error) {
	return livenessToSafety(network, options, false, false)
}

func livenessToSafety(network *Network, options TransformOptions, source, sourceOneStep bool) (*Network, error) {
	if network == nil || network.Graph == nil {
		return nil, fmt.Errorf("live: missing network")
	}
	p := network.Graph
	if err := p.Check(); err != nil {
		return nil, fmt.Errorf("live: invalid graph: %w", err)
	}
	regs, pis, pos := registers(p), trueInputs(p), trueOutputs(p)
	for _, group := range [][]aig.Edge{options.Liveness, options.Fairness, options.SafetyAssertions, options.SafetyAssumptions} {
		for _, edge := range group {
			if aig.Aig_ObjIsNone(p, edge) {
				return nil, fmt.Errorf("live: property signal is not in the input graph")
			}
		}
	}

	if options.Mode < 0 || options.Mode > 4 {
		return nil, fmt.Errorf("live: invalid transformation mode")
	}
	oneStep := options.Mode == 3 || options.Mode == 4
	// The proof-oriented C one-step constructor returns its input AIG when
	// there are no registers; the separate simulation constructor does not.
	if oneStep && !options.Restartable && regs == 0 {
		return network, nil
	}
	doLive := options.Mode != 1
	doSafety := options.Mode == 0 || options.Mode == 1 || options.Mode == 4
	copyGraph := newGraphCopy(p, "_l2s")
	if source {
		copyGraph = newSourceGraphCopy(p, "_l2s")
	}
	dst := copyGraph.destination
	names := []string{}
	next := []aig.Edge{}
	registerNames := []string{}
	for i, input := range p.Inputs()[:pis] {
		image := aig.Aig_ObjCreateCi(dst)
		copyGraph.mapping[input.NodeID()] = image
		aig.Aig_ObjSetCopy(p, input, image)
		names = append(names, inputName(network, i))
	}
	save := aig.InvalidEdge
	if doLive {
		save = aig.Aig_ObjCreateCi(dst)
		names = append(names, "SAVE_BIERE")
	}
	originalRegisters := make([]aig.Edge, regs)
	for r, input := range p.Inputs()[pis:] {
		image := aig.Aig_ObjCreateCi(dst)
		copyGraph.mapping[input.NodeID()] = image
		aig.Aig_ObjSetCopy(p, input, image)
		originalRegisters[r] = image
		registerNames = append(registerNames, inputName(network, pis+r))
	}
	saved, savedNext, capture := aig.InvalidEdge, aig.InvalidEdge, aig.InvalidEdge
	if doLive && !oneStep {
		saved = aig.Aig_ObjCreateCi(dst)
		registerNames = append(registerNames, "SAVED_LO")
		savedNext = aig.Aig_Or(dst, save, saved)
		if options.Restartable {
			capture = save
		} else {
			capture = aig.Aig_And(dst, save, saved.Not())
		}
	}
	copyGraph.copySourceNodes()
	for r := 0; r < regs; r++ {
		next = append(next, copyGraph.signal(p.Outputs()[pos+r]))
	}
	mapSignals := func(signals []aig.Edge) []aig.Edge {
		result := make([]aig.Edge, len(signals))
		for i, e := range signals {
			result[i] = copyGraph.signal(e)
		}
		return result
	}
	live, fair := mapSignals(options.Liveness), mapSignals(options.Fairness)
	assertions, assumptions := mapSignals(options.SafetyAssertions), mapSignals(options.SafetyAssumptions)
	badOutputs, outputNames := []aig.Edge{}, []string{}
	if doSafety && !(sourceOneStep && len(assertions) == 0) {
		bad := aig.Edge(1)
		if len(assertions) > 0 {
			bad = aig.Aig_And(dst, all(dst, assertions).Not(), all(dst, assumptions))
		}
		badOutputs = append(badOutputs, bad)
		outputNames = append(outputNames, "assert_safety_l2s")
	}
	if doLive {
		loop := save
		seenLive, seenFair := live, fair
		if !oneStep {
			next = append(next, savedNext)
			compare := aig.Edge(0)
			shadows := options.ShadowRegisters
			if shadows == nil {
				shadows = make([]int32, regs)
				for i := range shadows {
					shadows[i] = int32(i)
				}
			}
			used := make(map[int32]bool)
			for _, r := range shadows {
				if r < 0 || int(r) >= regs || used[r] {
					return nil, fmt.Errorf("live: invalid or repeated shadow register %d", r)
				}
				used[r] = true
				shadow := aig.Aig_ObjCreateCi(dst)
				registerNames = append(registerNames, inputName(network, pis+int(r))+"__SHADOW")
				next = append(next, aig.Aig_Mux(dst, capture, originalRegisters[r], shadow))
				compare = aig.Aig_And(dst, compare, aig.Aig_Exor(dst, originalRegisters[r], shadow).Not())
			}
			loop = aig.Aig_And(dst, saved, compare)
			active := aig.Aig_Or(dst, save, saved)
			monitor := func(signal aig.Edge, name string) aig.Edge {
				seen := aig.Aig_ObjCreateCi(dst)
				registerNames = append(registerNames, name)
				previous := seen
				if options.Restartable {
					previous = aig.Aig_And(dst, save.Not(), seen)
				}
				next = append(next, aig.Aig_Or(dst, previous, aig.Aig_And(dst, signal, active)))
				return seen
			}
			seenLive = make([]aig.Edge, len(live))
			for i, e := range live {
				seenLive[i] = monitor(e, fmt.Sprintf("liveness%d__SEEN", i))
			}
			seenFair = make([]aig.Edge, len(fair))
			for i, e := range fair {
				seenFair[i] = monitor(e, fmt.Sprintf("fairness%d__SEEN", i))
			}
			for _, formula := range options.LTL {
				node, err := ParseFormula(formula)
				if err != nil {
					return nil, err
				}
				if isWellFormed(node) == 0 {
					return nil, fmt.Errorf("live: formula is outside Boolean combinations of GF")
				}
				if checkSignalNameExistence(network, node) == 0 {
					return nil, fmt.Errorf("live: LTL formula names an unknown output")
				}

				// Populate all source PO copies before binding formula leaves.
				for i := 0; i < pos; i++ {
					copyGraph.signal(p.Outputs()[i])
				}
				populateBoolWithAigNodePtr(network, p, dst, node)
				var signals []aig.Edge
				var groups [][]*ltlNode
				populateAigPointerUnitGF(dst, node, &signals, &groups)
				for i, signal := range signals {
					seen := monitor(signal, fmt.Sprintf("ltl%d__GF", i))
					for _, gf := range groups[i] {
						setAIGNodePtrOfGloballyNode(gf, seen)
					}
				}
				holds := buildLogicFromLTLNode(dst, node)
				bad := aig.Aig_And(dst, loop, holds.Not())
				if !source {
					bad = aig.Aig_And(dst, bad, all(dst, seenFair))
				}
				badOutputs = append(badOutputs, bad)
				outputNames = append(outputNames, "ltl_l2s")
			}
		} else {
			compare := aig.Edge(0)
			for r := 0; r < regs; r++ {
				compare = aig.Aig_And(dst, compare, aig.Aig_Exor(dst, originalRegisters[r], next[r]).Not())
			}
			loop = aig.Aig_And(dst, save, compare)
			if len(options.LTL) > 0 {
				return nil, fmt.Errorf("live: one-step LTL construction requires general loop mode")
			}
		}
		if len(options.LTL) == 0 {
			bad := aig.Aig_And(dst, loop, aig.Aig_And(dst, all(dst, seenFair), all(dst, seenLive).Not()))
			badOutputs = append(badOutputs, bad)
			outputNames = append(outputNames, "assert_liveness_l2s")
		}
	}
	for _, e := range badOutputs {
		aig.Aig_ObjCreateCo(dst, e)
	}
	// Native constraints denote forbidden behavior and must follow properties.
	constraints := int(aig.Aig_ManConstrNum(p))
	for i := pos - constraints; i < pos; i++ {
		aig.Aig_ObjCreateCo(dst, copyGraph.signal(p.Outputs()[i]))
		outputNames = append(outputNames, outputName(network, i))
	}
	for i, e := range next {
		aig.Aig_ObjCreateCo(dst, e)
		outputNames = append(outputNames, registerNames[i]+"__next")
	}
	aig.Aig_ManSetRegNum(dst, int32(len(next)))
	aig.Aig_ManSetConstrNum(dst, int32(constraints))
	names = append(names, registerNames...)
	return &Network{Graph: dst, InputNames: names, OutputNames: outputNames}, nil
}

// sourceNetwork retains network names and properties while using the AIG
// explicitly supplied to the C-shaped entry points.
func sourceNetwork(n *Network, p *aig.Graph) *Network {
	if n == nil || p == nil {
		panic("live: missing network or source AIG")
	}
	copy := *n
	copy.Graph = p
	return &copy
}

func LivenessToSafetyTransformation(mode int32, n *Network, p *aig.Graph, live, fair, assertions, assumptions []aig.Edge) *aig.Graph {
	result, err := livenessToSafety(sourceNetwork(n, p), TransformOptions{Mode: mode, Liveness: live, Fairness: fair, SafetyAssertions: assertions, SafetyAssumptions: assumptions}, true, false)
	if err != nil {
		panic(err)
	}
	return result.Graph
}
func LivenessToSafetyTransformationAbs(mode int32, n *Network, p *aig.Graph, flops []int32, live, fair, assertions, assumptions []aig.Edge) *aig.Graph {
	result, err := livenessToSafety(sourceNetwork(n, p), TransformOptions{Mode: mode, ShadowRegisters: flops, Liveness: live, Fairness: fair, SafetyAssertions: assertions, SafetyAssumptions: assumptions}, true, false)
	if err != nil {
		panic(err)
	}
	return result.Graph
}
func LivenessToSafetyTransformationOneStepLoop(mode int32, n *Network, p *aig.Graph, live, fair, assertions, assumptions []aig.Edge) *aig.Graph {
	if p != nil && registers(p) == 0 {
		return p
	}
	if mode == 0 {
		mode = 4
	}
	if mode == 2 {
		mode = 3
	}
	result, err := livenessToSafety(sourceNetwork(n, p), TransformOptions{Mode: mode, Liveness: live, Fairness: fair, SafetyAssertions: assertions, SafetyAssumptions: assumptions}, true, true)
	if err != nil {
		panic(err)
	}
	return result.Graph
}
func validSourceLTL(n *Network, formulas []string) ([]string, error) {
	valid := make([]string, 0, len(formulas))
	for _, formula := range formulas {
		node, err := ParseFormula(formula)
		if err == nil && isWellFormed(node) != 0 && checkSignalNameExistence(n, node) != 0 {
			valid = append(valid, formula)
		}
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("live: no valid LTL properties")
	}
	return valid, nil
}
func LivenessToSafetyTransformationWithLTL(mode int32, n *Network, p *aig.Graph, live, fair, assertions, assumptions []aig.Edge) *aig.Graph {
	n = sourceNetwork(n, p)
	formulas := n.LTLProperties
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		var err error
		formulas, err = validSourceLTL(n, formulas)
		if err != nil {
			panic(err)
		}
	}
	result, err := livenessToSafety(n, TransformOptions{Mode: mode, LTL: formulas, Liveness: live, Fairness: fair, SafetyAssertions: assertions, SafetyAssumptions: assumptions}, true, false)
	if err != nil {
		panic(err)
	}
	return result.Graph
}
func selectOutputs(n *Network, prefix string) []aig.Edge {
	signals := []aig.Edge{}
	for i := 0; i < propertyOutputs(n.Graph); i++ {
		if strings.HasPrefix(outputName(n, i), prefix) {
			signals = append(signals, n.Graph.Outputs()[i])
		}
	}
	return signals
}
func populateLivenessVector(n *Network, p *aig.Graph) []aig.Edge {
	return selectOutputs(sourceNetwork(n, p), "assert_fair")
}
func populateFairnessVector(n *Network, p *aig.Graph) []aig.Edge {
	return selectOutputs(sourceNetwork(n, p), "assume_fair")
}
func selectOutputAliases(n *Network, prefixes ...string) []aig.Edge {
	result := []aig.Edge{}
	for i := 0; i < propertyOutputs(n.Graph); i++ {
		for _, prefix := range prefixes {
			if strings.HasPrefix(outputName(n, i), prefix) {
				result = append(result, n.Graph.Outputs()[i])
				break
			}
		}
	}
	return result
}
func populateSafetyAssertionVector(n *Network, p *aig.Graph) []aig.Edge {
	return selectOutputAliases(sourceNetwork(n, p), "assert_safety", "Assert")
}
func populateSafetyAssumptionVector(n *Network, p *aig.Graph) []aig.Edge {
	return selectOutputAliases(sourceNetwork(n, p), "assume_safety", "Assume")
}
func prepareFlopVector(p *aig.Graph, length int32) []int32 {
	if length < 0 || int(length) > registers(p) {
		panic("live: abstraction flop count")
	}
	result := make([]int32, length)
	for i := range result {
		result[i] = int32(i)
	}
	return result
}
func Aig_ManCiCleanupBiere(p *aig.Graph) int32 { return aig.Aig_ManCiNum(p) - int32(len(p.Inputs())) } // Native graph interfaces maintain their counts on every edit.
func Aig_ManCoCleanupBiere(p *aig.Graph) int32 { return aig.Aig_ManCoNum(p) - int32(len(p.Outputs())) }
func getPoIndex(p *aig.Graph, pivot aig.Edge) int32 {
	for i, e := range p.Outputs()[:trueOutputs(p)] {
		if e == pivot {
			return int32(i)
		}
	}
	return -1
}
func retrieveTruePiName(n *Network, old, new *aig.Graph, pivot aig.Edge) string {
	for i, e := range new.Inputs()[:trueInputs(new)] {
		if e == pivot {
			if i < trueInputs(old) {
				return inputName(n, i)
			}
			return "SAVE_BIERE"
		}
	}
	return "UNKNOWN"
}
func retrieveLOName(n *Network, old, new *aig.Graph, pivot aig.Edge, live, fair []aig.Edge) string {
	index := -1
	for i, e := range new.Inputs()[trueInputs(new):] {
		if e == pivot {
			index = i
			break
		}
	}
	r := registers(old)
	if index < 0 {
		return "UNKNOWN"
	}
	if index < r {
		return inputName(n, trueInputs(old)+index)
	}
	if index == r {
		return "SAVED_LO"
	}
	if index < 2*r+1 {
		return inputName(n, trueInputs(old)+index-r-1) + "__SHADOW"
	}
	return fmt.Sprintf("monitor%d", index-2*r-1)
}
func updateNewNetworkNameManager(n *Network, p *aig.Graph, piNames, loNames []string) {
	n.InputNames = append(append([]string(nil), piNames...), loNames...)
}
func commandTransform(frame *Frame, argv []string, ltl, simulation, abstraction bool) int32 {
	if frame.Network == nil {
		return 1
	}
	options := TransformOptions{}
	onlySafety, onlyLiveness, oneStep := false, false, false
	args := argv
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		args = args[1:]
	}
	if simulation {
		// The C simulation command calls getopt only once.
		oneStep = len(args) > 0 && strings.HasPrefix(args[0], "-1")
	} else {
		flags := []string{}
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") || len(arg) < 2 {
				return 1
			}
			for _, flag := range arg[1:] {
				flags = append(flags, "-"+string(flag))
			}
		}
		for _, flag := range flags {
			switch flag {
			case "-s":
				if onlyLiveness || onlySafety {
					return 1
				}
				onlyLiveness = true
			case "-l":
				if onlyLiveness || onlySafety {
					return 1
				}
				onlySafety = true
			case "-1":
				if oneStep && !onlySafety {
					panic("live: repeated -1 option")
				}
				oneStep = true
			case "-h":
				out, _ := frame.writers()
				fmt.Fprintln(out, "l2s [-s|-l] [-1]")
				return 1
			default:
				return 1
			}
		}
	}

	if simulation {
		onlyLiveness = true
	}
	if onlySafety && onlyLiveness {
		return 1
	}
	if onlySafety {
		options.Mode = 1
	} else if oneStep {
		options.Mode = 4
		if onlyLiveness {
			options.Mode = 3
		}
	} else if onlyLiveness {
		options.Mode = 2
	}
	n := frame.Network
	if abstraction && options.ShadowRegisters == nil {
		options.ShadowRegisters = prepareFlopVector(n.Graph, int32(registers(n.Graph)/2))
	}
	options.Liveness = populateLivenessVector(n, n.Graph)
	options.Fairness = populateFairnessVector(n, n.Graph)
	options.SafetyAssertions = populateSafetyAssertionVector(n, n.Graph)
	options.SafetyAssumptions = populateSafetyAssumptionVector(n, n.Graph)
	options.Restartable = simulation
	if ltl && !oneStep {
		options.LTL = append([]string(nil), frame.LTLProperties...)
		if len(options.LTL) == 0 {
			options.LTL = n.LTLProperties
		}
		if options.Mode == FULL_BIERE_MODE || options.Mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
			var err error
			options.LTL, err = validSourceLTL(n, options.LTL)
			if err != nil {
				_, errors := frame.writers()
				fmt.Fprintln(errors, err)
				return 1
			}
		}
	}
	result, err := livenessToSafety(n, options, true, oneStep && !simulation)
	if err != nil {
		_, errors := frame.writers()
		fmt.Fprintln(errors, err)
		return 1
	}
	frame.Network = result
	return 0
}
func Abc_CommandAbcLivenessToSafety(frame *Frame, argc int32, argv []string) int32 {
	return commandTransform(frame, argv, false, false, false)
}
func Abc_CommandAbcLivenessToSafetyAbstraction(frame *Frame, argc int32, argv []string) int32 {
	return commandTransform(frame, argv, false, false, true)
}
func Abc_CommandAbcLivenessToSafetyWithLTL(frame *Frame, argc int32, argv []string) int32 {
	return commandTransform(frame, argv, true, false, false)
}
