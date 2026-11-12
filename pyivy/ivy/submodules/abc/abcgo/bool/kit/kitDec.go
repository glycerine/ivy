package kit

import "math/bits"

type Kit_ManDec_t struct {
	nVarsMax                int32
	vTruthVars, vTruthNodes [][]uint32
	nVarsIn                 int32
	vLutsIn                 []uint32
	vSuppIn                 []int32
	ATimeIn                 [64]byte
	pTruthIn, pTruthOut     []uint32
	nVarsOut                int32
	Order                   [32]byte
	vLutsOut                []uint32
	vSuppOut                []int32
	ATimeOut                [64]byte
}

func Kit_DecOuputArrival(nVars int32, luts []uint32, times []byte) int32 {
	return int32(times[int(nVars)+len(luts)-1])
}
func Kit_ManDecStart(nVars int32) *Kit_ManDec_t {
	if nVars < 0 || nVars > 20 {
		panic("kit: decomposition supports 0..20 variables")
	}
	words := Kit_TruthWordNum(nVars)
	p := &Kit_ManDec_t{nVarsMax: nVars, pTruthIn: make([]uint32, words), pTruthOut: make([]uint32, words), vTruthVars: make([][]uint32, nVars), vTruthNodes: make([][]uint32, 64)}
	for i := range p.vTruthVars {
		p.vTruthVars[i] = make([]uint32, words)
		Kit_TruthIthVar(p.vTruthVars[i], nVars, int32(i))
	}
	for i := range p.vTruthNodes {
		p.vTruthNodes[i] = make([]uint32, words)
	}
	return p
}
func Kit_ManDecStop(p *Kit_ManDec_t) {
	if p != nil {
		*p = Kit_ManDec_t{}
	}
}
func Kit_DecComputeOuputArrival(nVars int32, supports []int32, lutSize int32, in, out []byte) int32 {
	if lutSize <= 0 || len(supports) == 0 || len(supports)%int(lutSize) != 0 {
		panic("kit: malformed LUT supports")
	}
	count := len(supports) / int(lutSize)
	if len(in) < int(nVars) || len(out) < int(nVars)+count {
		panic("kit: short arrival-time buffer")
	}
	copy(out[:nVars], in[:nVars])
	for v := 0; v < count; v++ {
		delay := byte(0)
		for _, index := range supports[v*int(lutSize) : (v+1)*int(lutSize)] {
			if index < 0 || index >= nVars+int32(v) {
				panic("kit: unordered LUT support")
			}
			delay = max(delay, out[index])
		}
		if delay == 255 {
			panic("kit: arrival-time overflow")
		}
		out[int(nVars)+v] = delay + 1
	}
	return int32(out[int(nVars)+count-1])
}
func Kit_DecComputeTruthOne(lutSize int32, lut []uint32, nVars int32, inputs [][]uint32, temp, result []uint32) {
	if lutSize < 0 || lutSize > 16 || len(inputs) < int(lutSize) {
		panic("kit: invalid LUT arity")
	}
	Kit_TruthClear(result, nVars)
	for assignment := int32(0); assignment < 1<<uint(lutSize); assignment++ {
		if Kit_TruthHasBit(lut, assignment) == 0 {
			continue
		}
		Kit_TruthFill(temp, nVars)
		for v := int32(0); v < lutSize; v++ {
			if assignment&(1<<uint(v)) != 0 {
				Kit_TruthAnd(temp, temp, inputs[v], nVars)
			} else {
				Kit_TruthSharp(temp, temp, inputs[v], nVars)
			}
		}
		Kit_TruthOr(result, result, temp, nVars)
	}
}
func Kit_DecComputeTruth(p *Kit_ManDec_t, nVars int32, supports []int32, lutSize int32, luts []uint32, result []uint32) {
	if nVars < 0 || nVars > p.nVarsMax || lutSize <= 0 || len(supports) == 0 || len(supports)%int(lutSize) != 0 {
		panic("kit: malformed LUT network")
	}
	count := len(supports) / int(lutSize)
	words := Kit_TruthWordNum(lutSize)
	if len(luts) != count*int(words) || count >= len(p.vTruthNodes) {
		panic("kit: LUT network exceeds storage")
	}
	inputs := make([][]uint32, lutSize)
	temp := make([]uint32, Kit_TruthWordNum(nVars))
	for v := 0; v < count; v++ {
		for i, index := range supports[v*int(lutSize) : (v+1)*int(lutSize)] {
			if index < 0 || index >= nVars+int32(v) {
				panic("kit: unordered LUT network")
			}
			if index < nVars {
				inputs[i] = p.vTruthVars[index]
			} else {
				inputs[i] = p.vTruthNodes[index-nVars]
			}
		}
		out := p.vTruthNodes[v]
		if v == count-1 {
			out = result
		}
		Kit_DecComputeTruthOne(lutSize, luts[v*int(words):(v+1)*int(words)], nVars, inputs, temp, out)
	}
}

// Kit_DecComputePattern labels equal cofactors of the top lutSize variables.
// The C draft had incorrect cofactor offsets and an incomplete uniqueness test;
// this implementation performs the intended equality partition explicitly.
func Kit_DecComputePattern(nVars int32, truth []uint32, lutSize int32, pattern []int32) int32 {
	if lutSize < 0 || lutSize > 6 || lutSize >= nVars || len(pattern) < 1<<uint(lutSize) {
		panic("kit: invalid pattern dimensions")
	}
	width := int32(1) << uint(nVars-lutSize)
	representatives := []int32{}
	equal := func(a, b int32) bool {
		for i := int32(0); i < width; i++ {
			if Kit_TruthHasBit(truth, a*width+i) != Kit_TruthHasBit(truth, b*width+i) {
				return false
			}
		}
		return true
	}
	for assignment := int32(0); assignment < 1<<uint(lutSize); assignment++ {
		index := int32(-1)
		for i, representative := range representatives {
			if equal(assignment, representative) {
				index = int32(i)
				break
			}
		}
		if index < 0 {
			index = int32(len(representatives))
			representatives = append(representatives, assignment)
		}
		pattern[assignment] = index
	}
	return int32(len(representatives))
}

// Shared variables partition the cofactor-label map into regions needing at
// most one extra bound-function bit. Search the smallest shared subset (<=6
// variables), with deterministic ties. The original recursive draft lacked
// its base case and never inspected Pattern; exhaustive search makes this
// intended partition operation explicit and verifiable.
func Kit_DecComputeShared_rec(pattern, vars []int32, nVars int32, shared []int32, first int32) int32 {
	if nVars < 0 || nVars > 6 || len(pattern) < 1<<uint(nVars) || len(vars) < int(nVars) {
		panic("kit: invalid shared-variable problem")
	}
	best := -1
	for size := 0; size <= int(nVars); size++ {
		for subset := 0; subset < 1<<uint(nVars); subset++ {
			if bits.OnesCount(uint(subset)) != size {
				continue
			}
			valid := true
			groups := map[int]map[int32]bool{}
			for v := int32(0); v < first && v < nVars; v++ {
				if subset&(1<<uint(v)) != 0 {
					valid = false
				}
			}
			if !valid {
				continue
			}
			for assignment := 0; assignment < 1<<uint(nVars); assignment++ {
				key := assignment & subset
				labels := groups[key]
				if labels == nil {
					labels = map[int32]bool{}
					groups[key] = labels
				}
				labels[pattern[assignment]] = true
				if len(labels) > 2 {
					valid = false
					break
				}
			}
			if valid {
				best = subset
				break
			}
		}
		if best >= 0 {
			break
		}
	}
	if best < 0 {
		return 0
	}
	count := int32(0)
	for i := int32(0); i < nVars; i++ {
		if best&(1<<uint(i)) != 0 {
			if count >= int32(len(shared)) {
				panic("kit: short shared-variable result")
			}
			shared[count] = vars[i]
			count++
		}
	}
	return count
}
func Kit_DecComputeShared(pattern []int32, lutSize int32, shared []int32) int32 {
	vars := make([]int32, lutSize)
	for i := range vars {
		vars[i] = int32(i)
	}
	return Kit_DecComputeShared_rec(pattern, vars, lutSize, shared, 0)
}
