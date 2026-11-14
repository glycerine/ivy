package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
	"math/rand"
	"time"
)

// SatSweep partitions phase-normalized simulation signatures, proves candidate
// equivalences incrementally, and rebuilds the graph using only proved merges.
// Independently checked SAT witnesses reject candidate pairs; unknown queries
// never install representatives. Register outputs are independent inputs here.
func SatSweep(graph *Gia_Man_t, params *Cec_ParFra_t) (*Gia_Man_t, error) {
	if params == nil {
		params = new(Cec_ParFra_t)
		Cec_ManFraSetDefaultParams(params)
	}
	if params.NWords < 0 || params.NRounds < 0 || params.NItersMax < 0 || params.NBTLimit < 0 || params.TimeLimit < 0 {
		return nil, fmt.Errorf("cec: negative sweeping limit")
	}
	params.iOutFail = -1
	graph.ClearCombinationalCounterexample()
	options := core.DefaultSATOptions()
	options.ConflictLimit = int64(params.NBTLimit)
	options.TreatRegistersAsInputs = true
	if params.TimeLimit > 0 {
		options.Deadline = time.Now().Add(time.Duration(params.TimeLimit) * time.Second)
	}
	proof, err := core.NewManager(graph, options)
	if err != nil {
		return nil, err
	}
	model := proof.Model()
	inputs := make([]bool, model.InputCount())
	phase, err := model.Values(inputs)
	if err != nil {
		return nil, err
	}
	words := int(params.NWords)
	if words < 1 {
		words = 1
	}
	// Each simulation word contributes 32 patterns, matching ABC's word width.
	signatures := make([][]byte, graph.ObjectCount())
	for id := range signatures {
		signatures[id] = make([]byte, words*4)
	}
	random := rand.New(rand.NewSource(1))
	for word := 0; word < words; word++ {
		if !options.Deadline.IsZero() && time.Now().After(options.Deadline) {
			break
		}
		for bit := uint(0); bit < 32; bit++ {
			for i := range inputs {
				inputs[i] = random.Intn(2) != 0
			}
			values, e := model.Values(inputs)
			if e != nil {
				return nil, e
			}
			for id, value := range values {
				if value != phase[id] {
					signatures[id][word*4+int(bit/8)] |= 1 << (bit % 8)
				}
			}
		}
	}
	representatives := make([]int32, len(signatures))
	for id := range representatives {
		representatives[id] = -1
	}
	buckets := make(map[string][]int32)
	levels := sweepLevels(graph)
	if params.NItersMax > 0 {
		for id := int32(0); int(id) < len(signatures); id++ {
			if !options.Deadline.IsZero() && time.Now().After(options.Deadline) {
				break
			}
			if params.NLevelMax > 0 && levels[id] > int(params.NLevelMax) {
				continue
			}
			_, _, _, kind := graph.Gate(id)
			if kind == core.Absent || kind == core.Output {
				continue
			}
			key := string(signatures[id])
			roots := buckets[key]
			matched := false
			for _, root := range roots {
				a, b := id<<1, root<<1
				if phase[id] {
					a ^= 1
				}
				if phase[root] {
					b ^= 1
				}
				result, e := proof.CheckPair(a, b)
				if e != nil {
					return nil, e
				}
				if result.Status == core.Equivalent {
					representatives[id] = root
					matched = true
					break
				}
			}
			if !matched {
				buckets[key] = append(roots, id)
			}
		}
	}
	for id, value := range phase {
		if value {
			gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(graph, int32(id)), 1)
		} else {
			gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(graph, int32(id)), 0)
		}
	}
	graph.StartRepresentatives()
	for id, repr := range representatives {
		if repr >= 0 {
			gia.Gia_ObjSetRepr(graph, int32(id), repr)
		}
	}
	// Rebuild next links in ascending ID order without mutating any source gate.
	for id := len(representatives) - 1; id > 0; id-- {
		if root := representatives[id]; root >= 0 {
			graph.SetNext(int32(id), graph.Next(root))
			graph.SetNext(root, int32(id))
		}
	}
	if params.FCheckMiter != 0 {
		check := options
		check.TreatRegistersAsInputs = true
		var result core.Result
		if params.FDualOut != 0 {
			result, err = core.VerifyPairs(graph, check)
		} else {
			result, err = core.VerifyMiter(graph, check)
		}
		if err != nil {
			return nil, err
		}
		if result.Counterexample != nil {
			cex := result.Counterexample
			params.iOutFail = int32(cex.Output)
			if err = graph.SetCombinationalCounterexample(cex.Inputs, cex.Output); err != nil {
				return nil, err
			}
			return nil, nil
		}
	}
	return sweepRebuild(graph, representatives, phase), nil
}

func Cec_ManSatSweeping(graph *Gia_Man_t, params *Cec_ParFra_t, silent int32) *Gia_Man_t {
	if params == nil {
		params = new(Cec_ParFra_t)
		Cec_ManFraSetDefaultParams(params)
	}
	result, err := cecOriginalSweep(graph, params)
	params.Error = err
	if err != nil && silent == 0 {
		fmt.Printf("CEC sweeping: %v\n", err)
	} else if err == nil && result == nil && params.iOutFail >= 0 && silent == 0 {
		fmt.Printf("Disproved at least one output of the miter (zero-based number %d).\n", params.iOutFail)
	}
	return result
}

func sweepRebuild(source *Gia_Man_t, representatives []int32, phase []bool) *Gia_Man_t {
	graph := aig.NewGraph()
	graph.Name = gia.Gia_ManName(source)
	copies := make(map[int32]aig.Edge)
	copies[0] = graph.Const0()
	for _, lit := range source.InputLiterals() {
		copies[lit>>1] = graph.NewInput()
	}
	var node func(int32) aig.Edge
	var edge func(int32) aig.Edge
	edge = func(lit int32) aig.Edge {
		result := node(lit >> 1)
		if lit&1 != 0 {
			result = result.Not()
		}
		return result
	}
	node = func(id int32) aig.Edge {
		// Representatives may include inputs; preserve the interface even when an
		// input is redundant, while replacing its fanout uses by the proved root.
		if root := representatives[id]; root >= 0 {
			result := node(root)
			if phase[id] != phase[root] {
				result = result.Not()
			}
			return result
		}
		if result, ok := copies[id]; ok {
			return result
		}
		a, b, c, kind := source.Gate(id)
		var result aig.Edge
		switch kind {
		case core.And:
			result = graph.AND(edge(a), edge(b))
		case core.Xor:
			result = graph.NewXor(edge(a), edge(b))
		case core.Mux:
			result = graph.Mux(edge(a), edge(b), edge(c))
		case core.Buffer, core.Output:
			result = edge(a)
		default:
			panic("cec: invalid validated sweep gate")
		}
		copies[id] = result
		return result
	}
	for _, lit := range source.OutputLiterals() {
		graph.AddOutput(edge(lit))
	}
	aig.Aig_ManSetRegNum(graph, int32(source.RegisterCount()))
	aig.Aig_ManSetConstrNum(graph, gia.Gia_ManConstrNum(source))
	return gia.Gia_ManFromAig(graph)
}

func sweepLevels(graph *Gia_Man_t) []int {
	levels := make([]int, graph.ObjectCount())
	for id := range levels {
		levels[id] = -1
	}
	var level func(int32) int
	level = func(id int32) int {
		if levels[id] >= 0 {
			return levels[id]
		}
		a, b, c, kind := graph.Gate(id)
		result := 0
		switch kind {
		case core.And, core.Xor:
			result = max(level(a>>1), level(b>>1)) + 1
		case core.Mux:
			result = max(level(a>>1), level(b>>1), level(c>>1)) + 1
		case core.Buffer, core.Output:
			result = level(a >> 1)
		}
		levels[id] = result
		return result
	}
	for id := range levels {
		level(int32(id))
	}
	return levels
}
