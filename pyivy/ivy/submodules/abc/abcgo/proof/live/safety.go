package live

import (
	"fmt"
	"math"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/proof/pdr"
	"github.com/glycerine/abcgo/sat/bsat"
)

type SafetyOptions struct {
	MaxStates     int
	ConflictLimit int64
	Deadline      time.Time
}
type SafetyResult struct {
	Status int32
	Inputs [][]bool
	States [][]bool
	Error  error
}
type SafetyChecker func(*aig.Graph, int, SafetyOptions) SafetyResult

func constraintLatchCount(n int) int32 {
	if n > 0 {
		return 1
	}
	return 0
}

// CheckSafety uses ABC's PDR engine. An explicit state budget selects the
// bounded reachable-image checker, because PDR has no equivalent state limit.
func CheckSafety(graph *aig.Graph, output int, options SafetyOptions) SafetyResult {
	if options.MaxStates > 0 {
		return checkSafetyImages(graph, output, options)
	}
	return CheckSafetyPDR(graph, output, options)
}

// CheckSafetyPDR checks a single output using ABC's PDR engine.
func CheckSafetyPDR(graph *aig.Graph, output int, options SafetyOptions) SafetyResult {
	if graph == nil || output < 0 || output >= trueOutputs(graph)-int(aig.Aig_ManConstrNum(graph)) {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: invalid safety output")}
	}
	if err := graph.Check(); err != nil {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: invalid graph: %w", err)}
	}
	if options.MaxStates != 0 || options.ConflictLimit < 0 || options.ConflictLimit > math.MaxInt32 {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: unsupported PDR limit")}
	}
	if !options.Deadline.IsZero() && !time.Now().Before(options.Deadline) {
		return SafetyResult{Status: -1}
	}
	c := newGraphCopy(graph, "_pdr")
	c.copyInputs()
	constraints := int(aig.Aig_ManConstrNum(graph))
	var invalid aig.Edge
	bad := c.mapEdge(outputDriver(graph, output))
	if constraints > 0 {
		// Constraint outputs are asserted for forbidden inputs. Once a
		// violation occurs, an invalid-history latch disables future bads.
		invalid = c.destination.NewInput()
		invalidNow := invalid
		for i := trueOutputs(graph) - constraints; i < trueOutputs(graph); i++ {
			invalidNow = c.destination.OR(invalidNow, c.mapEdge(outputDriver(graph, i)))
		}
		bad = c.destination.AND(bad, invalidNow.Not())
		aig.Aig_ObjCreateCo(c.destination, bad)
		for i := trueOutputs(graph); i < len(graph.Outputs()); i++ {
			aig.Aig_ObjCreateCo(c.destination, c.mapEdge(outputDriver(graph, i)))
		}
		aig.Aig_ObjCreateCo(c.destination, invalidNow)
	} else {
		aig.Aig_ObjCreateCo(c.destination, bad)
		for i := trueOutputs(graph); i < len(graph.Outputs()); i++ {
			aig.Aig_ObjCreateCo(c.destination, c.mapEdge(outputDriver(graph, i)))
		}
	}
	aig.Aig_ManSetRegNum(c.destination, int32(registers(graph))+constraintLatchCount(constraints))
	opts := pdr.DefaultOptions()
	opts.Silent = true
	opts.ConflictLimit = int32(options.ConflictLimit)
	if !options.Deadline.IsZero() {
		opts.Stop = func(int32) bool { return !time.Now().Before(options.Deadline) }
	}
	result, err := pdr.Solve(c.destination, &opts)
	if err != nil {
		return SafetyResult{Status: -1, Error: err}
	}
	if result.Status != pdr.Failed {
		return SafetyResult{Status: int32(result.Status)}
	}
	cex := result.Outputs[0].Witness
	if cex == nil || cex.IFrame < 0 || cex.NRegs != int32(registers(graph))+constraintLatchCount(constraints) || cex.NPis != int32(trueInputs(graph)) {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: invalid PDR witness")}
	}
	states, inputs := make([][]bool, 0, int(cex.IFrame)+1), make([][]bool, 0, int(cex.IFrame)+1)
	state := make([]bool, registers(graph))
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		values := make([]bool, trueInputs(graph))
		for i := range values {
			values[i] = util.Abc_InfoHasBit(cex.PData, cex.NRegs+frame*cex.NPis+int32(i))
		}
		states = append(states, append([]bool(nil), state...))
		inputs = append(inputs, values)
		outputs, err := graph.Evaluate(append(append([]bool(nil), values...), state...))
		if err != nil {
			return SafetyResult{Status: -1, Error: err}
		}
		for i := trueOutputs(graph) - constraints; i < trueOutputs(graph); i++ {
			if outputs[i] {
				return SafetyResult{Status: -1, Error: fmt.Errorf("live: PDR witness violates a constraint")}
			}
		}
		if frame == cex.IFrame && !outputs[output] {
			return SafetyResult{Status: -1, Error: fmt.Errorf("live: PDR witness does not fail selected output")}
		}
		state = append([]bool(nil), outputs[trueOutputs(graph):]...)
	}
	return SafetyResult{Status: 0, Inputs: inputs, States: states}
}

// checkSafetyImages explores reachable state images with an incremental transition
// CNF. SAT blocks complete next-state images, rather than enumerating inputs.
// A completed frontier proves safety; resource limits explicitly return unknown.
func checkSafetyImages(graph *aig.Graph, output int, options SafetyOptions) SafetyResult {
	if graph == nil || output < 0 || output >= trueOutputs(graph)-int(aig.Aig_ManConstrNum(graph)) {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: invalid safety output")}
	}
	if err := graph.Check(); err != nil {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: invalid graph: %w", err)}
	}
	if options.ConflictLimit < 0 {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: negative SAT conflict limit")}
	}
	if options.MaxStates <= 0 {
		options.MaxStates = 100000
	}
	solver := bsat.NewSolver()
	solver.SetLimits(options.ConflictLimit, 0)
	if !options.Deadline.IsZero() {
		solver.SetDeadline(options.Deadline)
	}
	nodes := graph.Nodes()
	for range nodes {
		solver.NewVariable()
	}
	literal := func(edge aig.Edge) bsat.Literal { return bsat.Literal(edge) }
	solver.AddClause(bsat.Positive(0))
	for id, node := range nodes {
		if id == 0 || aig.Aig_ObjIsNone(graph, aig.Edge(id)<<1) {
			continue
		}
		z := bsat.Positive(int32(id))
		a, b := literal(node.Fanins[0]), literal(node.Fanins[1])
		switch node.Kind {
		case aig.And:
			solver.AddClause(a^1, b^1, z)
			solver.AddClause(a, z^1)
			solver.AddClause(b, z^1)
		case aig.ExclusiveOr:
			solver.AddClause(a^1, b^1, z^1)
			solver.AddClause(a, b, z^1)
			solver.AddClause(a, b^1, z)
			solver.AddClause(a^1, b, z)
		case aig.Buffer, aig.Output:
			solver.AddClause(a^1, z)
			solver.AddClause(a, z^1)
		}
	}
	inputs := graph.Inputs()
	pis, regs, pos := trueInputs(graph), registers(graph), trueOutputs(graph)
	type reached struct {
		state, inputs []bool
		parent        int
	}
	queue := []reached{{state: make([]bool, regs), parent: -1}}
	key := func(state []bool) string {
		bytes := make([]byte, (len(state)+7)/8)
		for i, value := range state {
			if value {
				bytes[i/8] |= 1 << uint(i%8)
			}
		}
		return string(bytes)
	}
	visited := map[string]bool{key(queue[0].state): true}
	readEdge := func(edge aig.Edge) bool {
		value, ok := solver.Value(int32(edge.NodeID()))
		if !ok {
			panic("live: missing SAT model")
		}
		return value != edge.Complemented()
	}
	readInputs := func() []bool {
		values := make([]bool, pis)
		for i, e := range inputs[:pis] {
			values[i] = readEdge(e)
		}
		return values
	}
	for head := 0; head < len(queue); head++ {
		if !options.Deadline.IsZero() && time.Now().After(options.Deadline) {
			return SafetyResult{Status: -1}
		}
		current := queue[head]
		assumptions := make([]bsat.Literal, 0, regs+2)
		for r, value := range current.state {
			lit := literal(inputs[pis+r])
			if !value {
				lit ^= 1
			}
			assumptions = append(assumptions, lit)
		}
		for i := pos - int(aig.Aig_ManConstrNum(graph)); i < pos; i++ {
			assumptions = append(assumptions, literal(outputDriver(graph, i))^1)
		}
		query := append(append([]bsat.Literal(nil), assumptions...), literal(outputDriver(graph, output)))
		status := solver.Solve(query...)
		if status == bsat.Unknown {
			return SafetyResult{Status: -1}
		}
		if status == bsat.Satisfiable {
			trace := [][]bool{readInputs()}
			states := [][]bool{append([]bool(nil), current.state...)}
			for index := head; queue[index].parent >= 0; index = queue[index].parent {
				trace = append(trace, append([]bool(nil), queue[index].inputs...))
				states = append(states, append([]bool(nil), queue[queue[index].parent].state...))
			}
			for i, j := 0, len(trace)-1; i < j; i, j = i+1, j-1 {
				trace[i], trace[j] = trace[j], trace[i]
				states[i], states[j] = states[j], states[i]
			}
			return SafetyResult{Status: 0, Inputs: trace, States: states}
		}
		activation := bsat.Positive(solver.NewVariable())
		assumptions = append(assumptions, activation)
		for {
			status = solver.Solve(assumptions...)
			if status == bsat.Unknown {
				return SafetyResult{Status: -1}
			}
			if status == bsat.Unsatisfiable {
				break
			}
			next := make([]bool, regs)
			block := []bsat.Literal{activation ^ 1}
			for r := range next {
				edge := outputDriver(graph, pos+r)
				next[r] = readEdge(edge)
				lit := literal(edge)
				if next[r] {
					lit ^= 1
				}
				block = append(block, lit)
			}
			pattern := readInputs()
			stateKey := key(next)
			if !visited[stateKey] {
				if len(queue) >= options.MaxStates {
					return SafetyResult{Status: -1}
				}
				visited[stateKey] = true
				queue = append(queue, reached{state: next, inputs: pattern, parent: head})
			}
			solver.AddClause(block...)
		}
	}
	return SafetyResult{Status: 1}
}
