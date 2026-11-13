package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
	"time"
)

func cecVerifyOptions(p *Cec_ParCec_t) core.Options {
	options := core.DefaultCECOptions()
	options.ConflictLimit = int64(p.nBTLimit)
	if p.TimeLimit > 0 {
		options.Deadline = time.Now().Add(time.Duration(p.TimeLimit) * time.Second)
	}
	return options
}
func cecRecordResult(graph *Gia_Man_t, params *Cec_ParCec_t, result core.Result, err error) int32 {
	params.Error = err
	params.iOutFail = -1
	if err != nil {
		return -1
	}
	if result.Counterexample != nil {
		cex := result.Counterexample
		params.iOutFail = int32(cex.Output)
		if params.Error = graph.SetCombinationalCounterexample(cex.Inputs, cex.Output); params.Error != nil {
			return -1
		}
	}
	return int32(result.Status)
}

func Cec_ManVerify(graph *Gia_Man_t, params *Cec_ParCec_t) int32 {
	if params == nil {
		params = new(Cec_ParCec_t)
		Cec_ManCecSetDefaultParams(params)
	}
	// ABC handles phase and simple-input cases before selecting either SAT path.
	if status := Cec_ManHandleSpecialCases(graph, params); status != -1 {
		return status
	}
	if params.Error != nil {
		return -1
	}
	return cecVerifyPipeline(graph, params, cecOriginalSweep, cecVerifyFra)
}
func Cec_ManVerifySimple(graph *Gia_Man_t) int32 {
	params := new(Cec_ParCec_t)
	Cec_ManCecSetDefaultParams(params)
	if graph.RegisterCount() != 0 || len(graph.OutputLiterals()) != 2 {
		return -1
	}
	return Cec_ManVerify(graph, params)
}
func Cec_ManVerifyTwo(left, right *Gia_Man_t, verbose int32) int32 {
	if left == nil || right == nil {
		return -1
	}
	left.ClearCombinationalCounterexample()
	if _, err := core.NewModel(left); err != nil {
		return -1
	}
	if _, err := core.NewModel(right); err != nil {
		return -1
	}
	if len(left.InputLiterals()) != len(right.InputLiterals()) || len(left.OutputLiterals()) != len(right.OutputLiterals()) {
		return -1
	}
	if left == right {
		return 1
	}
	miter := gia.Gia_ManMiter(left, right, 0, 1, 0, 0, verbose)
	if miter == nil {
		return -1
	}
	defer gia.Gia_ManStop(miter)
	var params Cec_ParCec_t
	Cec_ManCecSetDefaultParams(&params)
	params.fVerbose = verbose
	status := Cec_ManVerify(miter, &params)
	if status != 0 {
		return status
	}
	witness := miter.CombinationalCounterexample()
	if witness == nil || int(witness.NPis) != len(left.InputLiterals()) || int64(len(witness.PData))*32 < int64(witness.NPis) {
		return -1
	}
	inputs := make([]bool, witness.NPis)
	for i := range inputs {
		inputs[i] = witness.PData[i>>5]&(uint32(1)<<uint(i&31)) != 0
	}
	a, errA := left.Evaluate(inputs)
	b, errB := right.Evaluate(inputs)
	if errA != nil || errB != nil || len(a) != len(b) {
		return -1
	}
	for i := range a {
		if a[i] != b[i] {
			if left.SetCombinationalCounterexample(inputs, i) != nil {
				return -1
			}
			return 0
		}
	}
	return -1
}

// Cec_ManTransformPattern copies a source-style 0/1 input vector into the
// manager's counterexample. A nil vector means the all-zero pattern.
func Cec_ManTransformPattern(graph *Gia_Man_t, output int32, values []int32) {
	inputs := make([]bool, len(graph.InputLiterals()))
	if values != nil && len(values) != len(inputs) {
		panic("cec: invalid pattern dimensions")
	}
	for i, v := range values {
		inputs[i] = v != 0
	}
	if err := graph.SetCombinationalCounterexample(inputs, int(output)); err != nil {
		panic(err)
	}
}

func Cec_ManVerifyOld(graph *Gia_Man_t, verbose int32, failed *int32, started abctime, silent int32) int32 {
	status, output, err := cecVerifyFra(graph, verbose)
	if failed != nil {
		*failed = output
	}
	if err != nil {
		status = -1
	}
	if silent == 0 {
		fmt.Printf("CEC status: %d\n", status)
	}
	return status
}

// Cec_ManHandleSpecialCases checks zero-input phases and primary-input-only
// differences without invoking SAT. -1 means that a nontrivial proof remains.
func Cec_ManHandleSpecialCases(graph *Gia_Man_t, params *Cec_ParCec_t) int32 {
	params.Error = nil
	params.iOutFail = -1
	if graph == nil {
		params.Error = fmt.Errorf("cec: nil network")
		return -1
	}
	graph.ClearCombinationalCounterexample()
	model, err := core.NewModel(graph)
	if err != nil {
		params.Error = err
		return -1
	}
	outputs := graph.OutputLiterals()
	if len(outputs)%2 != 0 {
		params.Error = fmt.Errorf("cec: odd dual-output miter size")
		return -1
	}
	inputs := make([]bool, len(graph.InputLiterals()))
	values, err := model.Evaluate(inputs)
	if err != nil {
		params.Error = err
		return -1
	}
	inputIndex := make(map[int32]int, len(inputs))
	for i, lit := range graph.InputLiterals() {
		inputIndex[lit>>1] = i
	}
	for i := 0; i < len(outputs); i += 2 {
		mismatch := values[i] != values[i+1]
		a, aInput := inputIndex[outputs[i]>>1]
		_, bInput := inputIndex[outputs[i+1]>>1]
		if !mismatch && outputs[i]>>1 != outputs[i+1]>>1 && (aInput || bInput) && (aInput || outputs[i]>>1 == 0) && (bInput || outputs[i+1]>>1 == 0) {
			if aInput {
				inputs[a] = true
			} else {
				inputs[inputIndex[outputs[i+1]>>1]] = true
			}
			mismatch = true
		}
		if mismatch {
			params.iOutFail = int32(i / 2)
			params.Error = graph.SetCombinationalCounterexample(inputs, i/2)
			if params.Error != nil {
				return -1
			}
			return 0
		}
	}
	for id := 1; id < graph.ObjectCount(); id++ {
		_, _, _, kind := graph.Gate(int32(id))
		if kind == core.And || kind == core.Xor || kind == core.Mux {
			return -1
		}
	}
	// With only constants, inputs and unary aliases, zero plus the input basis
	// exhaustively distinguishes all possible output functions.
	for bit := range inputs {
		inputs[bit] = true
		values, err = model.Evaluate(inputs)
		if err != nil {
			params.Error = err
			return -1
		}
		for i := 0; i < len(outputs); i += 2 {
			if values[i] != values[i+1] {
				params.iOutFail = int32(i / 2)
				params.Error = graph.SetCombinationalCounterexample(inputs, i/2)
				if params.Error != nil {
					return -1
				}
				return 0
			}
		}
		inputs[bit] = false
	}
	return 1
}
