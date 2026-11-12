// Package mfs minimizes multi-input logic using window observability don't-cares,
// SAT resubstitution, divisor counterexamples and bi-decomposition.
package mfs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

const MFS_FANIN_MAX = 12

type Mfs_Par_t struct {
	WinTfoLevs, FanoutsMax, DepthMax, WinMax, GrowthLevel                                             int
	BTLimit                                                                                           int64
	RrOnly, Resub, Area, MoreEffort, SwapEdge, OneHotness, Delay, Power, GiaSat, Verbose, VeryVerbose bool
}

func Abc_NtkMfsParsDefault(p *Mfs_Par_t) {
	*p = Mfs_Par_t{WinTfoLevs: 2, FanoutsMax: 30, DepthMax: 20, WinMax: 300, BTLimit: 5000, Resub: true}
}
func DefaultParams() Mfs_Par_t { var p Mfs_Par_t; Abc_NtkMfsParsDefault(&p); return p }

// Literal names a node and optional output inversion; node zero is constant one.
type Literal uint32

func (e Literal) ID() int      { return int(e >> 1) }
func (e Literal) Not() Literal { return e ^ 1 }

type Node struct {
	ID             int
	Input, Deleted bool
	Fanins         []int
	Truth          []uint32
	Level          int
	local          *aig.Graph
}

// Network owns stable node IDs. Fanins refer to regular nodes; inversion is in
// each local truth table. Outputs may be complemented. Register terminals, when
// present, are trailing CIs/COs and are optimized as combinational boundaries.
type Network struct {
	Name, Spec string
	Nodes      []Node
	Inputs     []int
	Outputs    []Literal
	Registers  int
	// Care outputs are clauses required true. A window uses clauses whose entire
	// CI support lies in that window, as the original source does.
	Care *aig.Graph
}

func NewNetwork() *Network { return &Network{Nodes: []Node{{ID: 0, Truth: []uint32{^uint32(0)}}}} }
func (n *Network) AddInput() int {
	id := len(n.Nodes)
	n.Nodes = append(n.Nodes, Node{ID: id, Input: true})
	n.Inputs = append(n.Inputs, id)
	return id
}
func truthWords(vars int) int {
	if vars <= 5 {
		return 1
	}
	return 1 << uint(vars-5)
}
func truthBit(t []uint32, a int) bool { return t[a>>5]&(uint32(1)<<uint(a&31)) != 0 }
func setBit(t []uint32, a int)        { t[a>>5] |= uint32(1) << uint(a&31) }
func normalizeTruth(t []uint32, vars int) []uint32 {
	if vars < 0 || vars > MFS_FANIN_MAX || len(t) < truthWords(vars) {
		panic("mfs: invalid local truth dimensions")
	}
	r := append([]uint32(nil), t[:truthWords(vars)]...)
	if vars < 5 {
		mask := (1 << uint(vars)) - 1
		for i := 0; i < 32; i++ {
			if truthBit(t, i&mask) {
				setBit(r, i)
			} else {
				r[0] &^= 1 << uint(i)
			}
		}
	}
	return r
}
func (n *Network) AddNode(fanins []int, truth []uint32) int {
	for _, id := range fanins {
		if id < 0 || id >= len(n.Nodes) || n.Nodes[id].Deleted {
			panic("mfs: invalid fanin")
		}
	}
	id := len(n.Nodes)
	node := Node{ID: id, Fanins: append([]int(nil), fanins...), Truth: normalizeTruth(truth, len(fanins))}
	node.local = truthGraph(node.Truth, len(fanins))
	for _, f := range fanins {
		node.Level = max(node.Level, n.Nodes[f].Level+1)
	}
	n.Nodes = append(n.Nodes, node)
	return id
}
func (n *Network) AddOutput(e Literal) {
	if e.ID() >= len(n.Nodes) || n.Nodes[e.ID()].Deleted {
		panic("mfs: invalid output")
	}
	n.Outputs = append(n.Outputs, e)
}
func (n *Network) topology() ([]int, error) {
	if n == nil || len(n.Nodes) == 0 || n.Nodes[0].Input || n.Nodes[0].Deleted || len(n.Nodes[0].Fanins) != 0 || len(n.Nodes[0].Truth) != 1 || n.Nodes[0].Truth[0] != ^uint32(0) {
		return nil, fmt.Errorf("mfs: invalid constant")
	}
	state := make([]byte, len(n.Nodes))
	var order []int
	var rec func(int) error
	rec = func(id int) error {
		if id < 0 || id >= len(n.Nodes) || n.Nodes[id].Deleted {
			return fmt.Errorf("mfs: invalid node %d", id)
		}
		if state[id] == 1 {
			return fmt.Errorf("mfs: combinational cycle at %d", id)
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		v := &n.Nodes[id]
		if v.ID != id || len(v.Fanins) > MFS_FANIN_MAX {
			return fmt.Errorf("mfs: invalid node shape %d", id)
		}
		if v.Input {
			if len(v.Fanins) != 0 {
				return fmt.Errorf("mfs: input has fanins")
			}
		} else if len(v.Truth) < truthWords(len(v.Fanins)) {
			return fmt.Errorf("mfs: truth too short at %d", id)
		}
		for _, f := range v.Fanins {
			if err := rec(f); err != nil {
				return err
			}
		}
		state[id] = 2
		order = append(order, id)
		return nil
	}
	for id, v := range n.Nodes {
		if !v.Deleted {
			if err := rec(id); err != nil {
				return nil, err
			}
		}
	}
	seen := map[int]bool{}
	for _, id := range n.Inputs {
		if id <= 0 || id >= len(n.Nodes) || !n.Nodes[id].Input || n.Nodes[id].Deleted || seen[id] {
			return nil, fmt.Errorf("mfs: invalid input list")
		}
		seen[id] = true
	}
	for id, v := range n.Nodes {
		if v.Input && !v.Deleted && !seen[id] {
			return nil, fmt.Errorf("mfs: unlisted input")
		}
	}
	for _, e := range n.Outputs {
		if e.ID() >= len(n.Nodes) || n.Nodes[e.ID()].Deleted {
			return nil, fmt.Errorf("mfs: invalid output")
		}
	}
	if n.Registers < 0 || n.Registers > len(n.Inputs) || n.Registers > len(n.Outputs) {
		return nil, fmt.Errorf("mfs: invalid register partition")
	}
	if n.Care != nil {
		if n.Care.InputCount() != len(n.Inputs) {
			return nil, fmt.Errorf("mfs: care input count mismatch")
		}
		if err := n.Care.Check(); err != nil {
			return nil, err
		}
	}
	return order, nil
}
func (n *Network) Check() error { _, e := n.topology(); return e }
func (n *Network) Evaluate(inputs []bool) ([]bool, error) {
	order, err := n.topology()
	if err != nil {
		return nil, err
	}
	if len(inputs) != len(n.Inputs) {
		return nil, fmt.Errorf("mfs: input width mismatch")
	}
	values := make([]bool, len(n.Nodes))
	values[0] = true
	for i, id := range n.Inputs {
		values[id] = inputs[i]
	}
	for _, id := range order {
		v := n.Nodes[id]
		if v.Input || id == 0 {
			continue
		}
		a := 0
		for i, f := range v.Fanins {
			if values[f] {
				a |= 1 << uint(i)
			}
		}
		values[id] = truthBit(v.Truth, a)
	}
	out := make([]bool, len(n.Outputs))
	for i, e := range n.Outputs {
		out[i] = values[e.ID()] != (e&1 != 0)
	}
	return out, nil
}
