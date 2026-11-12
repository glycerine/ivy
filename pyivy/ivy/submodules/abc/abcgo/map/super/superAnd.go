package super

// Exhaustive AND2/inverter enumeration from superAnd.c. Explicit edge phases
// replace tagged pointers, and owned strings replace fixed recursion buffers.
import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Super2_Edge struct {
	Gate     *Super2_Gate_t
	Inverted bool
}
type Super2_Gate_t struct {
	Truth    uint32
	Variable int
	One, Two Super2_Edge
}
type Super2_Lib_t struct {
	Inputs, Levels int
	Gates          []*Super2_Gate_t
}
type Super2_Man_t struct {
	Tried int64
	table map[uint32]*Super2_Gate_t
}

func Super2_Regular(edge Super2_Edge) *Super2_Gate_t { return edge.Gate }
func Super2_IsComplement(edge Super2_Edge) bool      { return edge.Inverted }
func Super2_Not(edge Super2_Edge) Super2_Edge        { edge.Inverted = !edge.Inverted; return edge }
func Super2_NotCond(edge Super2_Edge, condition bool) Super2_Edge {
	if condition {
		return Super2_Not(edge)
	}
	return edge
}
func Super2_ManStart() *Super2_Man_t { return &Super2_Man_t{table: make(map[uint32]*Super2_Gate_t)} }
func Super2_ManStop(p *Super2_Man_t) {
	if p != nil {
		p.table = nil
	}
}
func Super2_LibStart() *Super2_Lib_t { return &Super2_Lib_t{} }
func Super2_LibStop(p *Super2_Lib_t) {
	if p != nil {
		p.Gates = nil
	}
}
func Super2_LibDup(p *Super2_Lib_t) *Super2_Lib_t {
	return &Super2_Lib_t{Inputs: p.Inputs, Levels: p.Levels, Gates: append([]*Super2_Gate_t(nil), p.Gates...)}
}
func Super2_LibAddGate(p *Super2_Lib_t, g *Super2_Gate_t) { p.Gates = append(p.Gates, g) }
func (p *Super2_Lib_t) mask() uint32                      { return SUPER_MASK(int32(1 << uint(p.Inputs))) }
func (p *Super2_Lib_t) canonical(truth uint32) uint32 {
	truth &= p.mask()
	if truth&(1<<uint((1<<uint(p.Inputs))-1)) != 0 {
		truth ^= p.mask()
	}
	return truth
}
func Super2_LibFirst(manager *Super2_Man_t, inputs int32) *Super2_Lib_t {
	if inputs < 1 || inputs > 5 {
		panic("super2: input count must be 1..5")
	}
	p := &Super2_Lib_t{Inputs: int(inputs), Gates: []*Super2_Gate_t{{Variable: -1}}}
	for variable := 0; variable < int(inputs); variable++ {
		g := &Super2_Gate_t{Variable: variable}
		for mint := 0; mint < 1<<uint(inputs); mint++ {
			if mint&(1<<uint(variable)) != 0 {
				g.Truth |= 1 << uint(mint)
			}
		}
		p.Gates = append(p.Gates, g)
	}
	return p
}
func Super2_LibCompute(manager *Super2_Man_t, current *Super2_Lib_t) *Super2_Lib_t {
	next := Super2_LibDup(current)
	manager.table = make(map[uint32]*Super2_Gate_t)
	for _, g := range next.Gates {
		manager.table[next.canonical(g.Truth)] = g
	}
	manager.Tried = int64(len(next.Gates))
	maxFunctions := uint64(1) << uint((1<<uint(current.Inputs))-1)
	if uint64(len(next.Gates)) == maxFunctions {
		return next
	}
	for i, a := range current.Gates {
		for _, b := range current.Gates[:i] {
			for phase := 0; phase < 4; phase++ {
				left, right := a.Truth, b.Truth
				if phase&1 != 0 {
					left ^= next.mask()
				}
				if phase&2 != 0 {
					right ^= next.mask()
				}
				truth := left & right
				key := next.canonical(truth)
				manager.Tried++
				if _, exists := manager.table[key]; exists {
					continue
				}
				gate := &Super2_Gate_t{Truth: truth, Variable: -1, One: Super2_Edge{a, phase&1 != 0}, Two: Super2_Edge{b, phase&2 != 0}}
				manager.table[key] = gate
				next.Gates = append(next.Gates, gate)
			}
		}
	}
	return next
}
func GenerateAND(inputs, levels int) (*Super2_Lib_t, error) {
	if inputs < 1 || inputs > 5 || levels < 0 {
		return nil, fmt.Errorf("super2: invalid input/level count")
	}
	manager := Super2_ManStart()
	p := Super2_LibFirst(manager, int32(inputs))
	for level := 0; level < levels; level++ {
		p = Super2_LibCompute(manager, p)
		p.Levels = level + 1
	}
	return p, nil
}
func Super2_LibCompareGates(p *Super2_Lib_t, a, b *Super2_Gate_t) int {
	aTruth, bTruth := p.canonical(a.Truth), p.canonical(b.Truth)
	if aTruth < bTruth {
		return -1
	}
	if aTruth > bTruth {
		return 1
	}
	return 0
}
func Super2_LibWriteCompare(a, b string) int {
	clean := func(s string) string {
		var result strings.Builder
		for _, character := range s {
			if character >= 'A' {
				result.WriteRune(character)
			}
		}
		return result.String()
	}
	x, y := clean(a), clean(b)
	if x == y {
		return 0
	}
	if x < y {
		return 1
	}
	return -1
}
func Super2_LibWriteGate_rec(g *Super2_Gate_t, inverted bool, level int32) string {
	if g.One.Gate == nil {
		if g.Variable < 0 {
			if inverted {
				return "1$"
			}
			return "0$"
		}
		letter := 'a' + rune(g.Variable)
		if inverted {
			letter = 'A' + rune(g.Variable)
		}
		return string(letter)
	}
	if level <= 0 {
		panic("super2: insufficient formula depth")
	}
	left := Super2_LibWriteGate_rec(g.One.Gate, g.One.Inverted, level-1)
	right := Super2_LibWriteGate_rec(g.Two.Gate, g.Two.Inverted, level-1)
	if g.One.Inverted != g.Two.Inverted {
		if !g.One.Inverted {
			left, right = right, left
		}
	} else if Super2_LibWriteCompare(left, right) != 1 {
		left, right = right, left
	}
	if inverted {
		return "<" + left + right + ">"
	}
	return "(" + left + right + ")"
}
func Super2_LibWriteGate(w io.Writer, p *Super2_Lib_t, g *Super2_Gate_t) error {
	inverted := g.Truth != p.canonical(g.Truth)
	_, err := fmt.Fprintf(w, "%0*b   %s\n", 1<<uint(p.Inputs), p.canonical(g.Truth), Super2_LibWriteGate_rec(g, inverted, int32(p.Levels)))
	return err
}
func (p *Super2_Lib_t) Write(w io.Writer) error {
	gates := append([]*Super2_Gate_t(nil), p.Gates...)
	sort.SliceStable(gates, func(i, j int) bool { return Super2_LibCompareGates(p, gates[i], gates[j]) < 0 })
	if _, err := fmt.Fprintf(w, "# AND2/INV supergates.\n# Inputs=%d Levels=%d\n\n%6d\n", p.Inputs, p.Levels, len(gates)); err != nil {
		return err
	}
	for _, g := range gates {
		if err := Super2_LibWriteGate(w, p, g); err != nil {
			return err
		}
	}
	return nil
}
func Super2_LibWrite(p *Super2_Lib_t) error {
	var b bytes.Buffer
	if err := p.Write(&b); err != nil {
		return err
	}
	return os.WriteFile(fmt.Sprintf("superI%dL%d", p.Inputs, p.Levels), b.Bytes(), 0644)
}
func Super2_Precompute(inputs, levels, verbose int32) error {
	p, err := GenerateAND(int(inputs), int(levels))
	if err != nil {
		return err
	}
	if verbose != 0 {
		fmt.Printf("AND2 supergates: inputs=%d levels=%d gates=%d\n", inputs, levels, len(p.Gates))
	}
	return Super2_LibWrite(p)
}
