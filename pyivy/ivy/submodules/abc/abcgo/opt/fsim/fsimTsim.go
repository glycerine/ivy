package fsim

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
)

const (
	FSIM_ZER = 1
	FSIM_ONE = 2
	FSIM_UND = 3
)

type TerState struct {
	Words []uint32
	Next  *TerState
}
type TernaryResult struct {
	States     [][]uint32
	CycleStart int
}

func Aig_XsimNotCond(value, complement int) int {
	if value < 1 || value > 3 {
		panic("fsim: invalid ternary value")
	}
	if complement != 0 && value != FSIM_UND {
		return 3 ^ value
	}
	return value
}
func Aig_XsimAndCond(a, ca, b, cb int) int {
	a, b = Aig_XsimNotCond(a, ca), Aig_XsimNotCond(b, cb)
	if a == FSIM_ZER || b == FSIM_ZER {
		return FSIM_ZER
	}
	if a == FSIM_UND || b == FSIM_UND {
		return FSIM_UND
	}
	return FSIM_ONE
}
func Fsim_ManTerSimInfoGet(info []uint32, index int) int {
	return int(info[index>>4] >> uint((index&15)*2) & 3)
}
func Fsim_ManTerSimInfoSet(info []uint32, index, value int) {
	if index < 0 || value < 1 || value > 3 {
		panic("fsim: invalid packed ternary write")
	}
	shift := uint((index & 15) * 2)
	info[index>>4] = info[index>>4]&^(3<<shift) | uint32(value)<<shift
}
func Fsim_ManTerStateNext(state *TerState) *TerState { return state.Next }
func Fsim_ManTerStateSetNext(state, next *TerState)  { state.Next = next }
func Fsim_ManTerStateCreate(info []uint32, nPIs, nCIs, nWords int) []uint32 {
	if nPIs < 0 || nCIs < nPIs || nWords < (2*(nCIs-nPIs)+31)/32 {
		return nil
	}
	out := make([]uint32, nWords)
	for i := nPIs; i < nCIs; i++ {
		Fsim_ManTerSimInfoSet(out, i-nPIs, Fsim_ManTerSimInfoGet(info, i))
	}
	return out
}
func Fsim_ManTerStateLookup(words []uint32, bins []*TerState) bool {
	key := Fsim_ManTerStateHash(words, len(words), len(bins))
	for entry := bins[key]; entry != nil; entry = entry.Next {
		if slices.Equal(entry.Words, words) {
			return true
		}
	}
	return false
}
func Fsim_ManTerStateInsert(state *TerState, bins []*TerState) {
	key := Fsim_ManTerStateHash(state.Words, len(state.Words), len(bins))
	state.Next = bins[key]
	bins[key] = state
}
func Fsim_ManTerStatePrint(state []uint32, nRegs int) {
	for i := 0; i < nRegs; i++ {
		switch Fsim_ManTerSimInfoGet(state, i) {
		case FSIM_ZER:
			fmt.Print("0")
		case FSIM_ONE:
			fmt.Print("1")
		case FSIM_UND:
			fmt.Print("x")
		default:
			panic("fsim: empty ternary state value")
		}
	}
	fmt.Println()
}
func (p *Manager) ternaryRound(cis []int) ([]int, error) {
	if len(cis) != p.CIs {
		return nil, fmt.Errorf("fsim: wrong ternary CI count")
	}
	frontier := make([]int, p.Frontier)
	if p.constSlot > 0 {
		frontier[p.constSlot%p.Frontier] = FSIM_ONE
	}
	outputs := make([]int, p.COs)
	get := func(lit int) int { return Aig_XsimNotCond(frontier[(lit>>1)%p.Frontier], lit&1) }
	err := p.forEachOperation(func(o Fsim_Obj_t) {
		value := 0
		switch o.Kind {
		case aig.Input:
			value = cis[o.Index]
		case aig.And:
			value = Aig_XsimAndCond(get(o.Fan0), 0, get(o.Fan1), 0)
		case aig.ExclusiveOr:
			a, b := get(o.Fan0), get(o.Fan1)
			value = FSIM_UND
			if a != FSIM_UND && b != FSIM_UND {
				value = FSIM_ZER
				if a != b {
					value = FSIM_ONE
				}
			}
		case aig.Buffer, aig.Output:
			value = get(o.Fan0)
		}
		if o.Kind == aig.Output {
			outputs[o.Index] = value
		} else {
			frontier[o.Node%p.Frontier] = value
		}
	})
	return outputs, err
}
func ternaryKey(state []uint32) string {
	data := make([]byte, len(state)*4)
	for i, w := range state {
		binary.LittleEndian.PutUint32(data[4*i:], w)
	}
	return string(data)
}
func RunTernary(g *aig.Graph, maxStates int) (*TernaryResult, error) {
	if maxStates < 0 {
		return nil, fmt.Errorf("fsim: negative ternary state limit")
	}
	p, err := Fsim_ManCreate(g)
	if err != nil {
		return nil, err
	}
	defer Fsim_ManDelete(p)
	regs := p.CIs - p.PIs
	if regs <= 0 {
		return nil, fmt.Errorf("fsim: ternary sequential simulation needs registers")
	}
	if err = Fsim_ManFront(p, false); err != nil {
		return nil, err
	}
	cis := make([]int, p.CIs)
	for i := range cis {
		cis[i] = FSIM_ZER
		if i < p.PIs {
			cis[i] = FSIM_UND
		}
	}
	pack := func() []uint32 {
		state := make([]uint32, (2*regs+31)/32)
		for r := 0; r < regs; r++ {
			Fsim_ManTerSimInfoSet(state, r, cis[p.PIs+r])
		}
		return state
	}
	initial := pack()
	result := &TernaryResult{States: [][]uint32{initial}, CycleStart: -1}
	seen := map[string]int{ternaryKey(initial): 0}
	for {
		outputs, err := p.ternaryRound(cis)
		if err != nil {
			return nil, err
		}
		for r := 0; r < regs; r++ {
			cis[p.PIs+r] = outputs[p.POs+r]
		}
		state := pack()
		result.States = append(result.States, state)
		key := ternaryKey(state)
		if index, ok := seen[key]; ok {
			result.CycleStart = index
			return result, nil
		}
		seen[key] = len(result.States) - 1
		if maxStates > 0 && len(result.States) >= maxStates {
			return result, fmt.Errorf("fsim: ternary state limit reached before convergence")
		}
	}
}
func Fsim_ManTerSimulate(g *aig.Graph, verbose bool) ([][]uint32, error) {
	result, err := RunTernary(g, 0)
	if err == nil && verbose {
		fmt.Printf("FSIM ternary: states=%d cycle=%d\n", len(result.States), result.CycleStart)
	}
	if result == nil {
		return nil, err
	}
	return result.States, err
}

var statePrimes = [128]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
	1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
	2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
	2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
	3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
	3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
	4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
	4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
	6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
	6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
	7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
	8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

func Fsim_ManTerStateHash(state []uint32, nWords, nBins int) int {
	if nWords < 0 || nWords > len(state) || nBins <= 0 {
		panic("fsim: invalid ternary hash dimensions")
	}
	var hash uint32
	for i, w := range state[:nWords] {
		hash ^= w * statePrimes[i&127]
	}
	return int(hash % uint32(nBins))
}
