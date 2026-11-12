package mf

import "fmt"

type coarseGate struct {
	a, b, c int32
	kind    int
}

// Coarsened owns the ordered, cleaned real-XOR/MUX graph produced by the
// Gia_ManDupMuxes pipeline. Its interface literals follow GIA (zero=false).
type Coarsened struct {
	gates           []coarseGate
	inputs, outputs []int32
	siblings        []int32
	registers       int
	name, spec      string
}

func (n *Coarsened) ObjectCount() int        { return len(n.gates) }
func (n *Coarsened) InputLiterals() []int32  { return append([]int32(nil), n.inputs...) }
func (n *Coarsened) OutputLiterals() []int32 { return append([]int32(nil), n.outputs...) }
func (n *Coarsened) Gate(id int32) (int32, int32, int32, int) {
	g := n.gates[id]
	return g.a, g.b, g.c, g.kind
}
func (n *Coarsened) Sibling(id int32) int32 { return n.siblings[id] }
func (n *Coarsened) MuxesEnabled() bool     { return true }
func newCoarsened() *Coarsened {
	return &Coarsened{gates: []coarseGate{{kind: 0}}, siblings: []int32{0}}
}
func (n *Coarsened) append(g coarseGate) int32 {
	lit := int32(len(n.gates)) * 2
	n.gates = append(n.gates, g)
	n.siblings = append(n.siblings, 0)
	if g.kind == 1 {
		n.inputs = append(n.inputs, lit)
	}
	return lit
}

type coarseHash struct {
	n     *Coarsened
	table map[coarseGate]int32
}

func (h *coarseHash) intern(g coarseGate) int32 {
	if lit, ok := h.table[g]; ok {
		return lit
	}
	lit := h.n.append(g)
	h.table[g] = lit
	return lit
}
func (h *coarseHash) and(a, b int32) int32 {
	if a < 2 {
		if a == 1 {
			return b
		}
		return 0
	}
	if b < 2 {
		if b == 1 {
			return a
		}
		return 0
	}
	if a == b {
		return a
	}
	if a == b^1 {
		return 0
	}
	if a > b {
		a, b = b, a
	}
	return h.intern(coarseGate{a: a, b: b, kind: 2})
}
func (h *coarseHash) or(a, b int32) int32 { return h.and(a^1, b^1) ^ 1 }
func (h *coarseHash) xor(a, b int32) int32 {
	if a < 2 {
		return b ^ a
	}
	if b < 2 {
		return a ^ b
	}
	if a == b {
		return 0
	}
	if a == b^1 {
		return 1
	}
	if a < b {
		a, b = b, a
	}
	phase := (a ^ b) & 1
	a &^= 1
	b &^= 1
	return h.intern(coarseGate{a: a, b: b, kind: 3}) ^ phase
}
func (h *coarseHash) mux(control, then, otherwise int32) int32 {
	if control < 2 {
		if control == 1 {
			return then
		}
		return otherwise
	}
	if otherwise < 2 {
		if otherwise == 1 {
			return h.or(control^1, then)
		}
		return h.and(control, then)
	}
	if then < 2 {
		if then == 1 {
			return h.or(control, otherwise)
		}
		return h.and(control^1, otherwise)
	}
	if otherwise == then {
		return otherwise
	}
	if control == otherwise || control == then^1 {
		return h.and(otherwise, then)
	}
	if control == then || control == otherwise^1 {
		return h.or(otherwise, then)
	}
	if otherwise>>1 == then>>1 {
		return h.xor(control, otherwise)
	}
	if otherwise > then {
		otherwise, then = then, otherwise
		control ^= 1
	}
	phase := then & 1
	if phase != 0 {
		then ^= 1
		otherwise ^= 1
	}
	return h.intern(coarseGate{a: control, b: then, c: otherwise, kind: 4}) ^ phase
}
func coarseInternal(kind int) bool { return kind >= 2 && kind <= 5 }
func coarseFanin(g coarseGate) []int32 {
	switch g.kind {
	case 2, 3:
		return []int32{g.a, g.b}
	case 4:
		return []int32{g.c, g.b, g.a}
	case 5, 6:
		return []int32{g.a}
	}
	return nil
}

// coarseRecognize follows Gia_ObjRecognizeMux's four cases in their original
// precedence order. Data values are inverted because the root is !p0 & !p1.
func coarseRecognize(gates []coarseGate, root int) (control, then, otherwise int32, xor bool, ok bool) {
	g := gates[root]
	if g.kind != 2 || g.a&1 == 0 || g.b&1 == 0 {
		return
	}
	p, q := gates[g.a>>1], gates[g.b>>1]
	if p.kind != 2 && p.kind != 5 || q.kind != 2 && q.kind != 5 {
		return
	}
	if p.kind == 5 {
		p.b = p.a
	}
	if q.kind == 5 {
		q.b = q.a
	}
	if p.a>>1 == q.a>>1 && p.b>>1 == q.b>>1 && (p.a^q.a)&1 != 0 && (p.b^q.b)&1 != 0 {
		return 0, p.a, p.b, true, true
	}
	left, right := [2]int32{p.a, p.b}, [2]int32{q.a, q.b}
	for _, pair := range [][2]int{{1, 1}, {0, 0}, {0, 1}, {1, 0}} {
		i, j := pair[0], pair[1]
		if left[i]>>1 != right[j]>>1 || (left[i]^right[j])&1 == 0 {
			continue
		}
		if left[i]&1 != 0 {
			return right[j], right[1-j] ^ 1, left[1-i] ^ 1, false, true
		}
		return left[i], left[1-i] ^ 1, right[1-j] ^ 1, false, true
	}
	return
}

// Coarsen ports Gia_ManDupMuxes and its cleanup, without importing GIA's
// parent package. Limit zero still recognizes XORs; the source default is 3.
func Coarsen(network Network, limit int) (*Coarsened, error) {
	if network == nil || network.ObjectCount() < 1 || limit < 0 {
		return nil, fmt.Errorf("mf: invalid coarsening input or limit")
	}
	if n, ok := network.(interface{ MuxesEnabled() bool }); ok && n.MuxesEnabled() {
		return nil, fmt.Errorf("mf: coarsening requires an AIG without real MUX storage")
	}
	count := network.ObjectCount()
	gates := make([]coarseGate, count)
	siblings := make([]int32, count)
	refs := make([]int, count)
	coCount := 0
	for i := range gates {
		a, b, c, kind := network.Gate(int32(i))
		if kind == 2 && a == b {
			kind = 5
		}
		g := coarseGate{a, b, c, kind}
		if i == 0 && kind != 0 || i != 0 && kind == 0 {
			return nil, fmt.Errorf("mf: noncanonical constant at %d", i)
		}
		if kind != 0 && kind != 1 && kind != 2 && kind != 5 && kind != 6 {
			return nil, fmt.Errorf("mf: coarsening expects plain AIG gates, got kind %d at %d", kind, i)
		}
		if kind == 6 {
			coCount++
		}
		for _, lit := range coarseFanin(g) {
			if lit < 0 || int(lit>>1) >= i || gates[lit>>1].kind == 6 {
				return nil, fmt.Errorf("mf: invalid coarsening fanin %d at %d", lit, i)
			}
			refs[lit>>1]++
		}
		if provider, ok := network.(interface{ Sibling(int32) int32 }); ok {
			sib := provider.Sibling(int32(i))
			if sib < 0 || int(sib) >= i && sib != 0 {
				return nil, fmt.Errorf("mf: invalid coarsening sibling %d at %d", sib, i)
			}
			siblings[i] = sib
		}
		gates[i] = g
	}
	seen := make(map[int32]bool)
	for _, lit := range network.InputLiterals() {
		if lit <= 0 || lit&1 != 0 || int(lit>>1) >= count || gates[lit>>1].kind != 1 || seen[lit] {
			return nil, fmt.Errorf("mf: invalid coarsening CI %d", lit)
		}
		seen[lit] = true
	}
	inputIndex := 0
	inputLits := network.InputLiterals()
	for i, g := range gates {
		if g.kind == 1 {
			if inputIndex >= len(inputLits) || inputLits[inputIndex] != int32(i)*2 {
				return nil, fmt.Errorf("mf: CIs must follow object order")
			}
			inputIndex++
		}
		if g.kind == 1 && !seen[int32(i)*2] {
			return nil, fmt.Errorf("mf: unlisted coarsening CI %d", i)
		}
	}
	outputs := append([]int32(nil), network.OutputLiterals()...)
	for i, lit := range outputs {
		if lit < 0 || int(lit>>1) >= count {
			return nil, fmt.Errorf("mf: invalid coarsening output")
		}
		if gates[lit>>1].kind == 6 {
			outputs[i] = gates[lit>>1].a ^ (lit & 1)
		}
		if coCount == 0 {
			refs[outputs[i]>>1]++
		}
	}
	if coCount != 0 {
		if len(outputs) != coCount {
			return nil, fmt.Errorf("mf: CO list size does not match explicit objects")
		}
		coIndex := 0
		for _, g := range gates {
			if g.kind == 6 {
				if outputs[coIndex] != g.a {
					return nil, fmt.Errorf("mf: CO list does not match explicit object order")
				}
				coIndex++
			}
		}
	}
	registers := 0
	if provider, ok := network.(interface{ RegisterCount() int }); ok {
		registers = provider.RegisterCount()
		if registers < 0 || registers > len(network.InputLiterals()) || registers > len(outputs) {
			return nil, fmt.Errorf("mf: invalid register count %d", registers)
		}
	}
	n := newCoarsened()
	n.registers = registers
	if named, ok := network.(interface{ NetworkName() string }); ok {
		n.name = named.NetworkName()
	}
	if specified, ok := network.(interface{ NetworkSpec() string }); ok {
		n.spec = specified.NetworkSpec()
	}
	h := coarseHash{n: n, table: make(map[coarseGate]int32)}
	copies := make([]int32, count)
	copyLit := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
	for i := 1; i < count; i++ {
		g := gates[i]
		switch g.kind {
		case 1:
			copies[i] = n.append(coarseGate{kind: 1})
		case 6:
			copies[i] = n.append(coarseGate{a: copyLit(g.a), kind: 6})
		case 5:
			copies[i] = n.append(coarseGate{a: copyLit(g.a), kind: 5})
		case 2:
			control, then, otherwise, xor, ok := coarseRecognize(gates, i)
			if !ok || siblings[g.a>>1] != 0 || siblings[g.b>>1] != 0 {
				copies[i] = h.and(copyLit(g.a), copyLit(g.b))
			} else if xor {
				copies[i] = h.xor(copyLit(then), copyLit(otherwise))
			} else if refs[g.a>>1]+refs[g.b>>1] > limit {
				copies[i] = h.and(copyLit(g.a), copyLit(g.b))
			} else {
				copies[i] = h.mux(copyLit(control), copyLit(then), copyLit(otherwise))
			}
		}
		if sib := siblings[i]; sib != 0 {
			node, alternative := copies[i]>>1, copies[sib]>>1
			if node > alternative && coarseInternal(n.gates[node].kind) && coarseInternal(n.gates[alternative].kind) {
				n.siblings[node] = alternative
			}
		}
	}
	for _, lit := range outputs {
		driver := copyLit(lit)
		n.outputs = append(n.outputs, driver)
		if coCount == 0 {
			n.append(coarseGate{a: driver, kind: 6})
		}
	}
	return n.cleanup(), nil
}

// Cleanup marks from every CO and buffer fanin, follows sibling chains, and
// then copies retained objects in original order (not DFS order).
func (n *Coarsened) cleanup() *Coarsened {
	used := make([]bool, len(n.gates))
	var mark func(int32)
	mark = func(id int32) {
		if used[id] {
			return
		}
		used[id] = true
		for _, lit := range coarseFanin(n.gates[id]) {
			mark(lit >> 1)
		}
		if sib := n.siblings[id]; sib != 0 {
			mark(sib)
		}
	}
	for id, g := range n.gates {
		if g.kind == 5 || g.kind == 6 {
			mark(int32(id))
		}
	}
	out := newCoarsened()
	out.registers = n.registers
	out.name, out.spec = n.name, n.spec
	copies := make([]int32, len(n.gates))
	for id := 1; id < len(n.gates); id++ {
		g := n.gates[id]
		if coarseInternal(g.kind) && g.kind != 5 && !used[id] {
			continue
		}
		copyLit := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
		switch g.kind {
		case 2, 3:
			g.a, g.b = copyLit(g.a), copyLit(g.b)
		case 4:
			g.a, g.b, g.c = copyLit(g.a), copyLit(g.b), copyLit(g.c)
		case 5, 6:
			g.a = copyLit(g.a)
		}
		copies[id] = out.append(g)
	}
	for id, sib := range n.siblings {
		if sib != 0 && copies[id] != 0 && copies[sib] != 0 {
			out.siblings[copies[id]>>1] = copies[sib] >> 1
		}
	}
	for _, lit := range n.outputs {
		out.outputs = append(out.outputs, copies[lit>>1]^(lit&1))
	}
	return out
}
func (n *Coarsened) RegisterCount() int { return n.registers }

func (n *Coarsened) NetworkName() string { return n.name }
func (n *Coarsened) NetworkSpec() string { return n.spec }
