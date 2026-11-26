package sfm

// This file holds the network helpers used by the Go-native mapped-network
// optimizer (sfm_dec_*.go, sfmArea.go, sfmTim.go, sfm_mapped.go). They are not
// part of the C sfm engine used by "&mfs" and "mfs2"; that engine is the
// line-by-line port in sfmCore.go, sfmNtk.go, sfmWin.go, sfmSat.go and
// sfmCnf.go and does not use any of these helpers.

import (
	"fmt"

	"github.com/glycerine/abcgo/bool/kit"
)

// NewNetwork builds a network for the mapped-network optimizer. Truth rows
// contain expanded 64-bit truth words; fanin zero is the least significant bit.
func NewNetwork(fanins [][]int, pis, pos int, truths [][]uint64, fixed, empty []bool) (*Sfm_Ntk_t, error) {
	n := len(fanins)
	if pis < 0 || pos < 0 || pis+pos > n || len(truths) != n {
		return nil, fmt.Errorf("sfm: invalid interface dimensions")
	}
	p := &Sfm_Ntk_t{Pis: pis, Pos: pos, Params: DefaultParams(), Fanins: make([][]int, n), Truths: make([][]uint64, n), Fixed: make([]bool, n), Empty: make([]bool, n), Deleted: make([]bool, n), Mapping: make([]*Gate, n)}
	p.nObjs = n
	p.nNodes = n - pis - pos
	p.pPars = &p.Params
	if fixed != nil && len(fixed) != n || empty != nil && len(empty) != n {
		return nil, fmt.Errorf("sfm: invalid object flags")
	}
	copy(p.Fixed, fixed)
	copy(p.Empty, empty)
	for id := range fanins {
		p.Fanins[id] = append([]int(nil), fanins[id]...)
		p.Truths[id] = append([]uint64(nil), truths[id]...)
		if id >= pis && id < n-pos && len(fanins[id]) < 6 && len(truths[id]) > 0 {
			p.Truths[id][0] = stretch(truths[id][0], len(fanins[id]))
		}
	}
	if err := p.refresh(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Sfm_Ntk_t) topology() ([]int, error) {
	n := len(p.Fanins)
	if p.Pis < 0 || p.Pos < 0 || p.Pis+p.Pos > n || len(p.Truths) != n || len(p.Fixed) != n || len(p.Empty) != n || len(p.Deleted) != n {
		return nil, fmt.Errorf("sfm: invalid storage")
	}
	state := make([]byte, n)
	var order []int
	var rec func(int) error
	rec = func(id int) error {
		if id < 0 || id >= n || p.Deleted[id] {
			return fmt.Errorf("sfm: invalid fanin %d", id)
		}
		if state[id] == 1 {
			return fmt.Errorf("sfm: cycle at%d", id)
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		fs := p.Fanins[id]
		if id < p.Pis {
			if len(fs) != 0 || p.Fixed[id] {
				return fmt.Errorf("sfm: invalid PI")
			}
		} else if Sfm_ObjIsPo(p, id) {
			if len(fs) != 1 || p.Fixed[id] {
				return fmt.Errorf("sfm: invalid PO")
			}
		} else if len(fs) > 16 || len(p.Truths[id]) < truthWords(len(fs)) {
			return fmt.Errorf("sfm: invalid local truth at%d", id)
		}
		for _, f := range fs {
			if f >= n-p.Pos {
				return fmt.Errorf("sfm: PO used as fanin")
			}
			if err := rec(f); err != nil {
				return err
			}
		}
		state[id] = 2
		order = append(order, id)
		return nil
	}
	for id := range p.Fanins {
		if !p.Deleted[id] {
			if err := rec(id); err != nil {
				return nil, err
			}
		}
	}
	return order, nil
}

// Check validates a mapped-optimizer network.
func (p *Sfm_Ntk_t) Check() error { _, err := p.topology(); return err }

func (p *Sfm_Ntk_t) refresh() error {
	order, err := p.topology()
	if err != nil {
		return err
	}
	n := len(p.Fanins)
	p.nObjs = n
	p.nNodes = n - p.Pis - p.Pos
	p.Fanouts = make([][]int, n)
	p.Levels = make([]int, n)
	p.Reverse = make([]int, n)
	for _, id := range order {
		level := 0
		for _, f := range p.Fanins[id] {
			p.Fanouts[f] = append(p.Fanouts[f], id)
			level = max(level, p.Levels[f])
		}
		if !p.Empty[id] {
			level++
		}
		p.Levels[id] = level
	}
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		level := 0
		for _, fo := range p.Fanouts[id] {
			level = max(level, p.Reverse[fo])
		}
		if !p.Empty[id] {
			level++
		}
		p.Reverse[id] = level
	}
	return nil
}

// Evaluate simulates the network on one input assignment and returns the
// PO values. It accepts networks of both the C engine (Sfm_NtkConstruct)
// and the mapped optimizer (NewNetwork).
func (p *Sfm_Ntk_t) Evaluate(inputs []bool) ([]bool, error) {
	if len(inputs) != p.Pis {
		return nil, fmt.Errorf("sfm: input width mismatch")
	}
	if p.vTruths != nil {
		return p.evaluateEngine(inputs), nil
	}
	order, err := p.topology()
	if err != nil {
		return nil, err
	}
	values := make([]bool, len(p.Fanins))
	copy(values, inputs)
	for _, id := range order {
		if Sfm_ObjIsPi(p, id) {
			continue
		}
		if Sfm_ObjIsPo(p, id) {
			values[id] = values[p.Fanins[id][0]]
			continue
		}
		a := 0
		for i, f := range p.Fanins[id] {
			if values[f] {
				a |= 1 << uint(i)
			}
		}
		values[id] = bit(p.Truths[id], a)
	}
	return append([]bool(nil), values[len(values)-p.Pos:]...), nil
}

// evaluateEngine simulates a C-engine network through Sfm_NodeReadFanins
// and Sfm_NodeReadTruth. Objects need not be in topological order after
// resubstitution, so the evaluation is recursive.
func (p *Sfm_Ntk_t) evaluateEngine(inputs []bool) []bool {
	n := len(p.Fanins)
	values := make([]int8, n) // 0 unknown, 1 false, 2 true
	var rec func(int) bool
	rec = func(id int) bool {
		if values[id] != 0 {
			return values[id] == 2
		}
		var v bool
		switch {
		case Sfm_ObjIsPi(p, id):
			v = inputs[id]
		case Sfm_ObjIsPo(p, id):
			v = rec(Sfm_NodeReadFanins(p, id)[0])
		default:
			a := 0
			for i, f := range Sfm_NodeReadFanins(p, id) {
				if rec(f) {
					a |= 1 << uint(i)
				}
			}
			v = bit(Sfm_NodeReadTruth(p, id), a)
		}
		values[id] = 1
		if v {
			values[id] = 2
		}
		return v
	}
	out := make([]bool, p.Pos)
	for i := range out {
		out[i] = rec(n - p.Pos + i)
	}
	return out
}

func truthWords(vars int) int {
	if vars <= 6 {
		return 1
	}
	return 1 << uint(vars-6)
}
func bit(truth []uint64, a int) bool { return truth[a>>6]&(uint64(1)<<uint(a&63)) != 0 }
func set(truth []uint64, a int)      { truth[a>>6] |= uint64(1) << uint(a&63) }
func stretch(t uint64, vars int) uint64 {
	if vars >= 6 {
		return t
	}
	n := 1 << uint(vars)
	t &= (uint64(1) << uint(n)) - 1
	for size := n; size < 64; size *= 2 {
		t |= t << uint(size)
	}
	return t
}

// sfmTruthToClauses derives both ISOP covers and emits implications to the
// local output variable nVars. Clauses use ordinary var<<1|negative literals.
func sfmTruthToClauses(truth uint64, large []uint64, vars int) [][]int {
	if vars < 0 || vars > 16 {
		panic("sfm: invalid truth support")
	}
	words := []uint64{stretch(truth, vars)}
	if vars > 6 {
		if len(large) < truthWords(vars) {
			panic("sfm: short truth")
		}
		words = large[:truthWords(vars)]
	}
	all0, all1 := true, true
	for _, w := range words {
		all0 = all0 && w == 0
		all1 = all1 && w == ^uint64(0)
	}
	if all0 || all1 {
		lit := vars << 1
		if all0 {
			lit ^= 1
		}
		return [][]int{{lit}}
	}
	var clauses [][]int
	for c := 0; c < 2; c++ {
		t := make([]uint32, 2*len(words))
		for i, w := range words {
			if c != 0 {
				w = ^w
			}
			t[2*i], t[2*i+1] = uint32(w), uint32(w>>32)
		}
		var cover []uint32
		if kit.Kit_TruthIsop(t, int32(vars), &cover, 0) != 0 {
			panic("sfm: ISOP failed")
		}
		for _, cube := range cover {
			var clause []int
			for v := 0; v < vars; v++ {
				literal := (cube >> uint(2*v)) & 3
				if literal == 1 {
					clause = append(clause, v<<1)
				} else if literal == 2 {
					clause = append(clause, v<<1|1)
				}
			}
			clause = append(clause, vars<<1|c)
			clauses = append(clauses, clause)
		}
	}
	return clauses
}

func sfmTranslateClauses(cnf [][]int, faninMap []int, pivot int) [][]int {
	r := make([][]int, len(cnf))
	for i, c := range cnf {
		r[i] = make([]int, len(c))
		for j, l := range c {
			v := faninMap[l>>1]
			r[i][j] = v<<1 | (l & 1)
			if v == pivot {
				r[i][j] ^= 1
			}
		}
	}
	return r
}
