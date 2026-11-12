package sfm

import "fmt"

func NewNetwork(fanins [][]int, pis, pos int, truths [][]uint64, fixed, empty []bool) (*Sfm_Ntk_t, error) {
	n := len(fanins)
	if pis < 0 || pos < 0 || pis+pos > n || len(truths) != n {
		return nil, fmt.Errorf("sfm: invalid interface dimensions")
	}
	p := &Sfm_Ntk_t{Pis: pis, Pos: pos, Params: DefaultParams(), Fanins: make([][]int, n), Truths: make([][]uint64, n), Fixed: make([]bool, n), Empty: make([]bool, n), Deleted: make([]bool, n), Mapping: make([]*Gate, n)}
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
	p.cnfs = Sfm_CreateCnf(p)
	return p, nil
}
func Sfm_NtkConstruct(fanins [][]int, pis, pos int, fixed, empty []bool, truths []uint64, starts []int, large []uint64) (*Sfm_Ntk_t, error) {
	if len(truths) != len(fanins) {
		return nil, fmt.Errorf("sfm: truth count mismatch")
	}
	rows := make([][]uint64, len(fanins))
	for id, fs := range fanins {
		if len(fs) <= 6 {
			rows[id] = []uint64{truths[id]}
		} else {
			if len(starts) != len(fanins) || starts[id] < 0 || starts[id]+truthWords(len(fs)) > len(large) {
				return nil, fmt.Errorf("sfm: invalid extended truth offset")
			}
			rows[id] = large[starts[id] : starts[id]+truthWords(len(fs))]
		}
	}
	return NewNetwork(fanins, pis, pos, rows, fixed, empty)
}
func Sfm_ConstructNetwork(fanins [][]int, pis, pos int) (*Sfm_Ntk_t, error) {
	truth := make([][]uint64, len(fanins))
	for i := range truth {
		truth[i] = []uint64{0}
	}
	return NewNetwork(fanins, pis, pos, truth, nil, nil)
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
func (p *Sfm_Ntk_t) Check() error { _, err := p.topology(); return err }
func (p *Sfm_Ntk_t) refresh() error {
	order, err := p.topology()
	if err != nil {
		return err
	}
	n := len(p.Fanins)
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
func (p *Sfm_Ntk_t) Evaluate(inputs []bool) ([]bool, error) {
	if len(inputs) != p.Pis {
		return nil, fmt.Errorf("sfm: input width mismatch")
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
func Sfm_NtkPrepare(p *Sfm_Ntk_t) {
	p.LevelMax = 0
	for _, l := range p.Levels {
		p.LevelMax = max(p.LevelMax, l)
	}
	p.LevelMax += p.Params.GrowthLevel
	p.nodes = nil
	p.divs = nil
	p.roots = nil
	p.tfo = nil
	p.order = nil
	p.witnesses = nil
}
func Sfm_NtkFree(p *Sfm_Ntk_t)                        { *p = Sfm_Ntk_t{} }
func Sfm_NodeReadFanins(p *Sfm_Ntk_t, id int) []int   { return append([]int(nil), p.Fanins[id]...) }
func Sfm_NodeReadTruth(p *Sfm_Ntk_t, id int) []uint64 { return append([]uint64(nil), p.Truths[id]...) }
func Sfm_NodeReadFixed(p *Sfm_Ntk_t, id int) bool     { return p.Fixed[id] }
func Sfm_NodeReadUsed(p *Sfm_Ntk_t, id int) bool {
	return !p.Deleted[id] && (len(p.Fanins[id]) > 0 || len(p.Fanouts[id]) > 0)
}
func Sfm_NtkUpdate(p *Sfm_Ntk_t, id, pin, replacement int, truth uint64) {
	if !Sfm_ObjIsNode(p, id) || pin < 0 || pin >= len(p.Fanins[id]) {
		panic("sfm: invalid replacement")
	}
	old := p.Fanins[id][pin]
	previous := append([]int(nil), p.Fanins[id]...)
	if truth == 0 || truth == ^uint64(0) {
		p.Fanins[id] = nil
	} else {
		p.Fanins[id] = append(p.Fanins[id][:pin], p.Fanins[id][pin+1:]...)
		if replacement >= 0 {
			p.Fanins[id] = append(p.Fanins[id], replacement)
		}
	}
	p.Truths[id] = []uint64{stretch(truth, len(p.Fanins[id]))}
	if err := p.refresh(); err != nil {
		panic(err)
	}
	var prune func(int)
	prune = func(f int) {
		if !Sfm_ObjIsNode(p, f) || p.Fixed[f] || p.Deleted[f] || len(p.Fanouts[f]) > 0 {
			return
		}
		p.Deleted[f] = true
		fs := p.Fanins[f]
		p.Fanins[f] = nil
		for _, fi := range fs {
			for i, fo := range p.Fanouts[fi] {
				if fo == f {
					p.Fanouts[fi] = append(p.Fanouts[fi][:i], p.Fanouts[fi][i+1:]...)
					break
				}
			}
			prune(fi)
		}
	}
	for _, f := range previous {
		prune(f)
	}
	_ = old
	if err := p.refresh(); err != nil {
		panic(err)
	}
	p.cnfs = Sfm_CreateCnf(p)
}
