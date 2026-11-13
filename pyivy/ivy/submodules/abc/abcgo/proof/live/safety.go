package live

import (
	"fmt"
	"math"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/proof/pdr"
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

// CheckSafety uses ABC's PDR engine. A nonzero MaxStates is unsupported.
func CheckSafety(graph *aig.Graph, output int, options SafetyOptions) SafetyResult {
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
	return decodeSafetyWitness(graph, output, result.Outputs[0].Witness, int32(registers(graph))+constraintLatchCount(constraints))
}

// decodeSafetyWitness validates dimensions, zero initial state and every
// transition before exposing an engine-supplied counterexample to callers.
func decodeSafetyWitness(graph *aig.Graph, output int, cex *util.Abc_Cex_t, expectedRegs int32) SafetyResult {
	if cex == nil || cex.IFrame < 0 || cex.NRegs != expectedRegs || cex.NPis != int32(trueInputs(graph)) {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: invalid PDR witness")}
	}
	expectedBits := int64(cex.NRegs) + (int64(cex.IFrame)+1)*int64(cex.NPis)
	if expectedBits < 0 || expectedBits != int64(cex.NBits) || int64(len(cex.PData))*32 < expectedBits {
		return SafetyResult{Status: -1, Error: fmt.Errorf("live: truncated or inconsistent PDR witness")}
	}
	for i := int32(0); i < cex.NRegs; i++ {
		if util.Abc_InfoHasBit(cex.PData, i) {
			return SafetyResult{Status: -1, Error: fmt.Errorf("live: nonzero PDR initial state")}
		}
	}
	constraints := int(aig.Aig_ManConstrNum(graph))
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
