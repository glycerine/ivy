package live

import (
	"bufio"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	SIMPLE_kCS int32 = iota
	kCS_WITH_SAFETY_INVARIANTS
	kCS_WITH_DISCOVER_MONOTONE_SIGNALS
	kCS_WITH_SAFETY_AND_DCS_INVARIANTS
	kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS
)

func readLiveSignal_0(p *aig.Graph, index int32) aig.Edge { return outputDriver(p, int(index)) }
func readLiveSignal_k(p *aig.Graph, index int32) aig.Edge { return outputDriver(p, int(index)) }
func introduceAbsorberLogic(p *aig.Graph, zero, index *int32, nonFirst int32) *aig.Graph {
	if zero == nil || index == nil || *zero < 0 || int(*zero) >= propertyOutputs(p) || nonFirst != 0 && *index < 0 {
		panic("live: invalid absorber indexes")
	}
	c := newGraphCopy(p, "_kCS")
	c.copyInputs()
	state := aig.Aig_ObjCreateCi(c.destination)
	input := c.signal(p.Outputs()[*zero])
	previous := input
	if nonFirst != 0 {
		previous = c.signal(p.Outputs()[*index])
	}
	next := aig.Aig_Or(c.destination, previous.Not(), state)
	property := aig.Aig_Or(c.destination, input, state.Not())
	if nonFirst == 0 {
		*index = int32(propertyOutputs(p))
	}
	for i := 0; i < propertyOutputs(p); i++ {
		driver := c.signal(p.Outputs()[i])
		if int32(i) == *index {
			driver = property
		}
		aig.Aig_ObjCreateCo(c.destination, driver)
	}
	if nonFirst == 0 {
		aig.Aig_ObjCreateCo(c.destination, property)
	}
	c.appendConstraints()
	for r := 0; r < registers(p); r++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[trueOutputs(p)+r]))
	}
	aig.Aig_ObjCreateCo(c.destination, next)
	aig.Aig_ManSetRegNum(c.destination, int32(registers(p)+1))
	c.setConstraints()
	return c.destination
}
func modifyAigToApplySafetyInvar(p *aig.Graph, target, invariant int32) {
	property := aig.Aig_And(p, outputDriver(p, int(target)), outputDriver(p, int(invariant)))
	aig.Aig_ObjPatchFanin0(p, p.Outputs()[target], property)
}
func kLiveProofModel(p *aig.Graph, directive, target, invariant int32) *aig.Graph {
	c := newGraphCopy(p, "_proof")
	c.copyInputs()
	for i := 0; i < propertyOutputs(p); i++ {
		property := c.signal(p.Outputs()[i])
		if int32(i) == target && (directive == 1 || directive == 3 || directive == 4) {
			if invariant < 0 {
				panic("live: safety invariant missing")
			}
			property = aig.Aig_And(c.destination, property, c.signal(p.Outputs()[invariant]))
		}
		aig.Aig_ObjCreateCo(c.destination, property.Not())
	}
	c.appendConstraints()
	for r := 0; r < registers(p); r++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[trueOutputs(p)+r]))
	}
	aig.Aig_ManSetRegNum(c.destination, int32(registers(p)))
	c.setConstraints()
	return c.destination
}
func flipConePdr(p *aig.Graph, directive, target, invariant, absorbers int32) int32 {
	return CheckSafety(kLiveProofModel(p, directive, target, invariant), int(target), SafetyOptions{}).Status
}

func findOutput(n *Network, marker string) int32 {
	for i := 0; i < propertyOutputs(n.Graph); i++ {
		if strings.Contains(outputName(n, i), marker) {
			return int32(i)
		}
	}
	return -1
}
func collectSafetyInvariantPOIndex(n *Network) int32 { return findOutput(n, "csSafetyInvar_") }
func collectUserGivenDisjunctiveMonotoneSignals(n *Network) [][]int32 {
	result := [][]int32{}
	for i := 0; i < propertyOutputs(n.Graph); i++ {
		if strings.Contains(outputName(n, i), "csLevel1Stabil_") {
			result = append(result, []int32{int32(i)})
		}
	}
	return result
}
func getVecOfVecFairness(reader io.Reader) [][]int32 {
	scanner := bufio.NewScanner(reader)
	result := [][]int32{}
	for scanner.Scan() {
		line := scanner.Text()
		if index := strings.Index(line, ":"); index >= 0 {
			line = line[index+1:]
		}
		fields := strings.Fields(line)
		group := []int32{}
		for _, field := range fields {
			value, err := strconv.ParseInt(field, 10, 32)
			if err != nil {
				panic(err)
			}
			group = append(group, int32(value))
		}
		if len(group) > 0 {
			result = append(result, group)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return result
}

type KLiveOptions struct {
	Directive    int32
	MaxAbsorbers int
	Barriers     [][]int32
	Safety       SafetyOptions
	Checker      SafetyChecker
}
type KLiveResult struct {
	Status     int32
	Absorbers  int
	LastSafety SafetyResult
	Graph      *aig.Graph
	Barriers   [][]int32
	Error      error
}

// KLive proves constrained stabilization when some finite absorber depth is
// safe. Failure at a finite depth is not a liveness disproof; exhaustion returns
// unknown and retains that depth's concrete safety counterexample.
func KLive(n *Network, options KLiveOptions) KLiveResult {
	if n == nil || n.Graph == nil {
		return KLiveResult{Status: -1, Error: fmt.Errorf("live: missing k-liveness network")}
	}
	if err := n.Graph.Check(); err != nil {
		return KLiveResult{Status: -1, Error: fmt.Errorf("live: invalid graph: %w", err)}
	}
	if options.Directive < 0 || options.Directive > 4 {
		return KLiveResult{Status: -1, Error: fmt.Errorf("live: invalid k-liveness directive")}
	}
	if findOutput(n, "csLiveTarget_") < 0 && options.Directive != 2 {
		return KLiveResult{Status: -1, Error: fmt.Errorf("live: csLiveTarget_ output missing")}
	}
	if !options.Safety.Deadline.IsZero() && time.Now().After(options.Safety.Deadline) {
		return KLiveResult{Status: -1}
	}
	checker := options.Checker
	if checker == nil {
		checker = CheckSafety
	}
	limit := options.MaxAbsorbers
	if limit <= 0 {
		limit = 500
	}
	barriers := options.Barriers
	if options.Directive == 2 || options.Directive == 3 {
		barriers = FindDisjunctiveMonotone(n, options.Safety, checker)
	} else if options.Directive == 4 && barriers == nil {
		barriers = collectUserGivenDisjunctiveMonotoneSignals(n)
	}
	if options.Directive == 2 {
		return KLiveResult{Status: -1, Barriers: barriers, Graph: n.Graph}
	}
	invariant := collectSafetyInvariantPOIndex(n)
	if (options.Directive == 1 || options.Directive == 3 || options.Directive == 4) && invariant < 0 {
		return KLiveResult{Status: -1, Error: fmt.Errorf("live: csSafetyInvar_ output missing")}
	}
	if (options.Directive == 3 || options.Directive == 4) && len(barriers) > 0 {
		if collectWindowBeginSignalWithDSC(n, n.Graph) < 0 || collectWithinWindowSignalWithDSC(n, n.Graph) < 0 || findPendingSignal(n) < 0 {
			return KLiveResult{Status: -1, Error: fmt.Errorf("live: barrier window/pending outputs missing")}
		}
		for _, group := range barriers {
			if len(group) == 0 {
				return KLiveResult{Status: -1, Error: fmt.Errorf("live: empty barrier group")}
			}
			for _, index := range group {
				if index < 0 || int(index) >= propertyOutputs(n.Graph) {
					return KLiveResult{Status: -1, Error: fmt.Errorf("live: barrier output index out of range")}
				}
			}
		}
	}
	zero, index := int32(-1), int32(-1)
	var working *aig.Graph
	if options.Directive == 3 || options.Directive == 4 {
		working = generateWorkingAigWithDSC(n.Graph, n, &zero, barriers)
	} else {
		working = generateWorkingAig(n.Graph, n, &zero)
	}
	current := introduceAbsorberLogic(working, &zero, &index, 0)
	result := KLiveResult{Status: -1, Graph: current, Barriers: barriers}
	for depth := 1; depth <= limit; depth++ {
		if !options.Safety.Deadline.IsZero() && time.Now().After(options.Safety.Deadline) {
			return result
		}
		proofModel := kLiveProofModel(current, options.Directive, index, invariant)
		result.LastSafety = checker(proofModel, int(index), options.Safety)
		result.Absorbers = depth
		result.Graph = current
		if result.LastSafety.Status == 1 {
			result.Status = 1
			return result
		}
		if result.LastSafety.Error != nil {
			result.Error = result.LastSafety.Error
			return result
		}
		if depth < limit {
			current = introduceAbsorberLogic(current, &zero, &index, int32(depth))
		}
	}
	return result
}
func Abc_CommandCS_kLiveness(frame *Frame, argc int32, argv []string) int32 {
	options := KLiveOptions{Checker: frame.Checker}
	args := argv
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		args = args[1:]
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c":
			options.Directive = 1
		case "-m":
			options.Directive = 2
		case "-C":
			options.Directive = 3
		case "-g":
			options.Directive = 4
		case "-K":
			i++
			if i == len(args) {
				return 1
			}
			value, err := strconv.Atoi(args[i])
			if err != nil || value < 1 {
				return 1
			}
			options.MaxAbsorbers = value
		default:
			return 1
		}
	}
	result := KLive(frame.Network, options)
	out, errors := frame.writers()
	if result.Error != nil {
		fmt.Fprintln(errors, result.Error)
		return 1
	}
	fmt.Fprintf(out, "k = %d, status = %d\n", result.Absorbers, result.Status)
	return 0
}
func Abc_CommandNChooseK(frame *Frame, argc int32, argv []string) int32 {
	k := 1
	if len(argv) > 1 {
		value, err := strconv.Atoi(argv[len(argv)-1])
		if err != nil {
			return 1
		}
		k = value
	}
	if frame.Network == nil || k < 0 || k > registers(frame.Network.Graph) {
		return 1
	}
	result := generateGeneralDisjunctiveTester(frame.Network, frame.Network.Graph, int32(k))
	frame.Network = &Network{Graph: result}
	return 0
}
