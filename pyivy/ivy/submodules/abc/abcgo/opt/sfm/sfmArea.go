package sfm

import "fmt"

type CellPair struct {
	Gate        int
	Permutation []int
	Gain        int
}

// permuteTruth assigns old pin i to external input permutation[i].
func permuteTruth(t uint64, permutation []int) uint64 {
	var result uint64
	for a := 0; a < 64; a++ {
		old := 0
		for i, v := range permutation {
			if a&(1<<uint(v)) != 0 {
				old |= 1 << uint(i)
			}
		}
		if t&(uint64(1)<<uint(old)) != 0 {
			result |= 1 << uint(a)
		}
	}
	return result
}
func permutations(n int, fn func([]int)) {
	p := make([]int, n)
	used := make([]bool, n)
	var rec func(int)
	rec = func(i int) {
		if i == n {
			fn(p)
			return
		}
		for v := 0; v < n; v++ {
			if !used[v] {
				used[v] = true
				p[i] = v
				rec(i + 1)
				used[v] = false
			}
		}
	}
	rec(0)
}
func flipTruth(t uint64, pin int) uint64 {
	var result uint64
	for a := 0; a < 64; a++ {
		if t&(uint64(1)<<uint(a^(1<<uint(pin)))) != 0 {
			result |= 1 << uint(a)
		}
	}
	return result
}

// Abc_NtkPrecomputeCellPairs preserves the source's input/output complement
// transforms, pin permutation, minimum-area choice and strict tie handling.
// Cells zero/one are the source's reserved constants and have no records.
func Abc_NtkPrecomputeCellPairs(cells []*Gate) [][]CellPair {
	r := make([][]CellPair, len(cells))
	for i := 2; i < len(cells); i++ {
		gate := cells[i]
		n := gate.Inputs
		if n < 1 || n > 6 || len(gate.Truth) == 0 {
			panic("sfm: unsupported polarity cell")
		}
		r[i] = make([]CellPair, n+1)
		for flip := 0; flip <= n; flip++ {
			target := ^gate.Truth[0]
			if flip < n {
				target = flipTruth(gate.Truth[0], flip)
			}
			best := CellPair{Gate: -1}
			bestArea := int(^uint(0) >> 1)
			for k := 2; k < len(cells); k++ {
				candidate := cells[k]
				if candidate.Inputs != n || i != k && candidate.Truth[0] == gate.Truth[0] {
					continue
				}
				permutations(n, func(perm []int) {
					if candidate.Area < bestArea && permuteTruth(candidate.Truth[0], perm) == target {
						bestArea = candidate.Area
						best = CellPair{Gate: k, Permutation: append([]int(nil), perm...), Gain: gate.Area - candidate.Area}
					}
				})
			}
			r[i][flip] = best
		}
	}
	return r
}
func Abc_NtkPrecomputeFirsts(cells []*Gate) []int {
	r := make([]int, len(cells))
	for i := range r {
		r[i] = -1
	}
	offset := 0
	for i := 2; i < len(cells); i++ {
		r[i] = offset
		offset += 3 * (cells[i].Inputs + 1)
	}
	return r
}
func Abc_NtkPrecomputePrint(cells []*Gate, pairs [][]CellPair) int {
	count := 0
	for i, records := range pairs {
		for pin, pair := range records {
			if pair.Gate >= 0 {
				fmt.Printf("%s transform%d -> %s gain%d pins%v\n", cells[i].Name, pin, cells[pair.Gate].Name, pair.Gain, pair.Permutation)
				count++
			}
		}
	}
	return count
}
func duplicate(a []int) bool {
	seen := map[int]bool{}
	for _, v := range a {
		if seen[v] {
			return true
		}
		seen[v] = true
	}
	return false
}
func Abc_ObjHasDupFanins(p *Sfm_Ntk_t, id int) bool  { return duplicate(p.Fanins[id]) }
func Abc_ObjHasDupFanouts(p *Sfm_Ntk_t, id int) bool { return duplicate(p.Fanouts[id]) }
func Abc_NodeCheckFanoutHasFanin(p *Sfm_Ntk_t, node, fanin int) int {
	for i, fo := range p.Fanouts[node] {
		for _, f := range p.Fanins[fo] {
			if f == fanin {
				return i
			}
		}
	}
	return -1
}
func gateIndex(cells []*Gate, g *Gate) int {
	for i, c := range cells {
		if c == g {
			return i
		}
	}
	return -1
}
func isUnary(p *Sfm_Ntk_t, id int, inv bool) bool {
	if !Sfm_ObjIsNode(p, id) || p.Deleted[id] || len(p.Fanins[id]) != 1 {
		return false
	}
	truth := uint64(0xaaaaaaaaaaaaaaaa)
	if inv {
		truth = ^truth
	}
	return p.Truths[id][0] == truth
}
func Abc_ObjChangeEval(p *Sfm_Ntk_t, id int, cells []*Gate, pairs [][]CellPair, invArea int) (int, bool) {
	index := gateIndex(cells, p.Mapping[id])
	if index < 2 || len(p.Fanins[id]) == 0 {
		return 0, false
	}
	record := pairs[index][len(p.Fanins[id])]
	if record.Gate < 0 {
		return 0, false
	}
	inverter := isUnary(p, id, true)
	gain := record.Gain
	if inverter {
		gain = invArea
	}
	need := false
	for _, fo := range p.Fanouts[id] {
		if inverter {
			for _, fi := range p.Fanins[fo] {
				if fi == p.Fanins[id][0] {
					return 0, false
				}
			}
		}
		if Abc_ObjHasDupFanins(p, fo) {
			return 0, false
		}
		if !Sfm_ObjIsNode(p, fo) || isUnary(p, fo, false) {
			need = true
			continue
		}
		if isUnary(p, fo, true) {
			if Abc_NodeCheckFanoutHasFanin(p, fo, id) >= 0 {
				return 0, false
			}
			gain += invArea
			continue
		}
		cell := gateIndex(cells, p.Mapping[fo])
		if cell < 2 {
			return 0, false
		}
		pin := -1
		for i, fi := range p.Fanins[fo] {
			if fi == id {
				pin = i
				break
			}
		}
		if pin < 0 || pairs[cell][pin].Gate < 0 {
			need = true
			continue
		}
		gain += pairs[cell][pin].Gain
	}
	if need {
		gain -= invArea
	}
	return gain, need
}
func applyPair(p *Sfm_Ntk_t, id int, cells []*Gate, record CellPair) {
	fs := make([]int, len(p.Fanins[id]))
	for i, v := range record.Permutation {
		fs[i] = p.Fanins[id][v]
	}
	p.Fanins[id] = fs
	p.Mapping[id] = cells[record.Gate]
	p.Truths[id] = append([]uint64(nil), cells[record.Gate].Truth...)
}

// insertMapped inserts before POs, preserving all existing internal IDs and
// updating PO indices/attributes. The returned ID is a new internal object.
func (p *Sfm_Ntk_t) insertMapped(fanins []int, gate *Gate) int {
	at := len(p.Fanins) - p.Pos
	p.Fanins = append(p.Fanins, nil)
	copy(p.Fanins[at+1:], p.Fanins[at:])
	p.Fanins[at] = append([]int(nil), fanins...)
	p.Truths = append(p.Truths, nil)
	copy(p.Truths[at+1:], p.Truths[at:])
	p.Truths[at] = append([]uint64(nil), gate.Truth...)
	p.Mapping = append(p.Mapping, nil)
	copy(p.Mapping[at+1:], p.Mapping[at:])
	p.Mapping[at] = gate
	for _, flags := range []*[]bool{&p.Fixed, &p.Empty, &p.Deleted} {
		*flags = append(*flags, false)
		copy((*flags)[at+1:], (*flags)[at:])
		(*flags)[at] = false
	}
	if err := p.refresh(); err != nil {
		panic(err)
	}
	return at
}
func patchAll(p *Sfm_Ntk_t, object, replacement int) {
	for id := range p.Fanins {
		if p.Deleted[id] {
			continue
		}
		for i, fi := range p.Fanins[id] {
			if fi == object {
				p.Fanins[id][i] = replacement
			}
		}
	}
}
func Abc_ObjChangePerform(p *Sfm_Ntk_t, id int, cells []*Gate, pairs [][]CellPair, needInv bool) {
	cell := gateIndex(cells, p.Mapping[id])
	outputs := append([]int(nil), p.Fanouts[id]...)
	originalPO := len(p.Fanins) - p.Pos
	inverter := isUnary(p, id, true)
	if inverter {
		replacement := p.Fanins[id][0]
		patchAll(p, id, replacement)
		p.Deleted[id] = true
		p.Fanins[id] = nil
		id = replacement
	} else {
		applyPair(p, id, cells, pairs[cell][len(p.Fanins[id])])
	}
	newInv := -1
	if needInv {
		newInv = p.insertMapped([]int{id}, cells[3])
		for i, fo := range outputs {
			if fo >= originalPO {
				outputs[i]++
			}
		}
	}
	for _, fo := range outputs {
		if p.Deleted[fo] {
			continue
		}
		if !Sfm_ObjIsNode(p, fo) || isUnary(p, fo, false) {
			if newInv < 0 {
				panic("sfm: missing polarity restoration")
			}
			for i, f := range p.Fanins[fo] {
				if f == id {
					p.Fanins[fo][i] = newInv
				}
			}
			continue
		}
		if isUnary(p, fo, true) {
			patchAll(p, fo, id)
			p.Deleted[fo] = true
			p.Fanins[fo] = nil
			continue
		}
		cell := gateIndex(cells, p.Mapping[fo])
		pin := -1
		for i, f := range p.Fanins[fo] {
			if f == id {
				pin = i
				break
			}
		}
		if cell >= 2 && pin >= 0 && pairs[cell][pin].Gate >= 0 {
			applyPair(p, fo, cells, pairs[cell][pin])
		} else {
			if newInv < 0 {
				panic("sfm: missing polarity inverter")
			}
			for i, f := range p.Fanins[fo] {
				if f == id {
					p.Fanins[fo][i] = newInv
				}
			}
		}
	}
	if err := p.refresh(); err != nil {
		panic(err)
	}
	p.cnfs = Sfm_CreateCnf(p)
}
func Abc_NtkChangePerform(p *Sfm_Ntk_t, cells []*Gate, verbose bool) int {
	if len(cells) < 4 {
		panic("sfm: cell library needs source constant/buffer/inverter prefix")
	}
	pairs := Abc_NtkPrecomputeCellPairs(cells)
	gainTotal := 0
	for id := p.Pis; id < len(p.Fanins)-p.Pos; id++ {
		if p.Deleted[id] || p.Fixed[id] || (len(p.Fanins[id]) < 2 && !isUnary(p, id, true)) || Abc_ObjHasDupFanouts(p, id) {
			continue
		}
		gain, need := Abc_ObjChangeEval(p, id, cells, pairs, cells[3].Area)
		if gain > 0 {
			Abc_ObjChangePerform(p, id, cells, pairs, need)
			gainTotal += gain
		}
	}
	if verbose {
		fmt.Printf("Total area gain: %d\n", gainTotal)
	}
	return gainTotal
}
