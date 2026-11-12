package dau

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/opt/dau/canon"
	"io"
	"math/bits"
	"strconv"
	"strings"
)

// Dss nodes retain stable addresses; complemented edges are explicit paired
// objects rather than C pointer-tag bits. Fans and Truth own their slices.
type Dss_Obj_t struct {
	ID                int32
	Type              Dau_DsdType_t
	Var               int32
	Fans              []int32
	Truth             []uint64
	Mark0, Mark1      bool
	support           int32 // immutable structural support-size aggregate
	regular, negative *Dss_Obj_t
}
type Dss_Ntk_t struct {
	NVars   int32
	Objects []*Dss_Obj_t
	Root    *Dss_Obj_t
}
type Dss_Fun_t struct {
	Dsd  int32
	Fans []byte
}
type Dss_Ent_t struct {
	Dsd0, Dsd1 int32
	Shared     []byte
	Func       *Dss_Fun_t
}
type Dss_Man_t struct {
	NVars, NonDecLimit     int32
	Objects                []*Dss_Obj_t
	unique                 map[string]int32
	cache                  map[string]*Dss_Ent_t
	CacheHits, CacheMisses [2]int64
}

func Dss_Regular(p *Dss_Obj_t) *Dss_Obj_t {
	if p != nil && p.regular != nil {
		return p.regular
	}
	return p
}
func Dss_Not(p *Dss_Obj_t) *Dss_Obj_t {
	if p == nil {
		panic("nil DSD edge")
	}
	if p.regular != nil {
		return p.regular
	}
	if p.negative == nil {
		p.negative = &Dss_Obj_t{regular: p}
	}
	return p.negative
}
func Dss_NotCond(p *Dss_Obj_t, c int32) *Dss_Obj_t {
	if c&1 != 0 {
		return Dss_Not(p)
	}
	return p
}
func Dss_IsComplement(p *Dss_Obj_t) int32                { return dauBool(p != nil && p.regular != nil) }
func Dss_ObjTruth(p *Dss_Obj_t) []uint64                 { return Dss_Regular(p).Truth }
func Dss_ObjClean(p *Dss_Obj_t)                          { *Dss_Regular(p) = Dss_Obj_t{} }
func Dss_ObjId(p *Dss_Obj_t) int32                       { return Dss_Regular(p).ID }
func Dss_ObjType(p *Dss_Obj_t) Dau_DsdType_t             { return Dss_Regular(p).Type }
func Dss_ObjSuppSize(p *Dss_Obj_t) int32                 { return Dss_Regular(p).support }
func Dss_ObjFaninNum(p *Dss_Obj_t) int32                 { return int32(len(Dss_Regular(p).Fans)) }
func Dss_ObjFaninC(p *Dss_Obj_t, i int32) int32          { return Dss_Regular(p).Fans[i] & 1 }
func Dss_VecObj(p []*Dss_Obj_t, id int32) *Dss_Obj_t     { return p[id] }
func Dss_VecConst0(p []*Dss_Obj_t) *Dss_Obj_t            { return p[0] }
func Dss_VecVar(p []*Dss_Obj_t, v int32) *Dss_Obj_t      { return p[v+1] }
func Dss_VecLitSuppSize(p []*Dss_Obj_t, lit int32) int32 { return p[lit/2].support }
func Dss_Obj2Lit(p *Dss_Obj_t) int32                     { return Dss_Regular(p).ID*2 + Dss_IsComplement(p) }
func Dss_Lit2Obj(p []*Dss_Obj_t, lit int32) *Dss_Obj_t   { return Dss_NotCond(p[lit/2], lit&1) }
func Dss_ObjFanin(p []*Dss_Obj_t, obj *Dss_Obj_t, i int32) *Dss_Obj_t {
	return p[Dss_Regular(obj).Fans[i]/2]
}
func Dss_ObjChild(p []*Dss_Obj_t, obj *Dss_Obj_t, i int32) *Dss_Obj_t {
	return Dss_Lit2Obj(p, Dss_Regular(obj).Fans[i])
}
func Dss_WordCountOnes(x uint32) int32             { return int32(bits.OnesCount32(x)) }
func Dss_Lit2Lit(mapping []int32, lit int32) int32 { return mapping[lit/2] ^ (lit & 1) }
func dssWords(n int32) int {
	if n < 0 || n > DAU_MAX_VAR {
		panic("DSD variable count outside configured range")
	}
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func dssBit(table []uint64, i int) bool { return table[i>>6]>>uint(i&63)&1 != 0 }
func dssSet(table []uint64, i int)      { table[i>>6] |= uint64(1) << uint(i&63) }
func dssStretch(table []uint64, n int32) {
	if n >= 6 {
		return
	}
	width := 1 << uint(n)
	mask := uint64(1)<<uint(width) - 1
	v := table[0] & mask
	for shift := width; shift < 64; shift *= 2 {
		v |= v << uint(shift)
	}
	table[0] = v
}
func dssTruthHex(table []uint64, n int) string {
	digits := ((1 << uint(n)) + 3) / 4
	var b strings.Builder
	for i := digits - 1; i >= 0; i-- {
		value := (table[i/16] >> uint((i%16)*4)) & 15
		b.WriteByte("0123456789ABCDEF"[value])
	}
	return b.String()
}
func dssReadHex(text string, n int32) ([]uint64, error) {
	table := make([]uint64, dssWords(n))
	digits := ((1 << uint(n)) + 3) / 4
	if len(text) > digits {
		return nil, fmt.Errorf("prime truth has too many hex digits")
	}
	for i := 0; i < len(text); i++ {
		v, err := strconv.ParseUint(text[len(text)-1-i:len(text)-i], 16, 4)
		if err != nil {
			return nil, err
		}
		table[i/16] |= v << uint((i%16)*4)
	}
	dssStretch(table, n)
	return table, nil
}
func dssNodeKey(kind Dau_DsdType_t, fans []int32, truth []uint64) string {
	data := make([]byte, 4+4*len(fans)+8*len(truth))
	binary.LittleEndian.PutUint32(data, uint32(kind))
	for i, f := range fans {
		binary.LittleEndian.PutUint32(data[4+4*i:], uint32(f))
	}
	for i, v := range truth {
		binary.LittleEndian.PutUint64(data[4+4*len(fans)+8*i:], v)
	}
	return string(data)
}
func dssAllocate(objects *[]*Dss_Obj_t, kind Dau_DsdType_t, fans, truthVars int32) *Dss_Obj_t {
	if fans < 0 || fans > DAU_MAX_VAR {
		panic("DSD fanin count outside configured range")
	}
	o := &Dss_Obj_t{ID: int32(len(*objects)), Type: kind, Var: 31, Fans: make([]int32, fans)}
	if truthVars > 0 {
		o.Truth = make([]uint64, dssWords(truthVars))
	}
	*objects = append(*objects, o)
	return o
}
func Dss_ObjAllocNtk(p *Dss_Ntk_t, kind Dau_DsdType_t, fans, truthVars int32) *Dss_Obj_t {
	return dssAllocate(&p.Objects, kind, fans, truthVars)
}
func Dss_ObjCreateNtk(p *Dss_Ntk_t, kind Dau_DsdType_t, fans []int32) *Dss_Obj_t {
	truthVars := int32(0)
	if kind == DAU_DSD_PRIME {
		truthVars = int32(len(fans))
	}
	o := Dss_ObjAllocNtk(p, kind, int32(len(fans)), truthVars)
	copy(o.Fans, fans)
	for _, lit := range fans {
		o.support += p.Objects[lit/2].support
	}
	return o
}
func Dss_NtkAlloc(n int32) *Dss_Ntk_t {
	dssWords(n)
	p := &Dss_Ntk_t{NVars: n}
	Dss_ObjAllocNtk(p, DAU_DSD_CONST0, 0, 0)
	for i := int32(0); i < n; i++ {
		o := Dss_ObjAllocNtk(p, DAU_DSD_VAR, 0, 0)
		o.Var = i
		o.support = 1
	}
	return p
}
func Dss_NtkFree(p *Dss_Ntk_t) {
	if p != nil {
		*p = Dss_Ntk_t{}
	}
}

type dssExpr struct {
	kind             Dau_DsdType_t
	variable         int32
	phase            bool
	children         []*dssExpr
	truth            []uint64
	start, open, end int
}
type dssParser struct {
	text     string
	pos      int
	fallback []uint64
}

func (p *dssParser) parse() (*dssExpr, error) {
	start := p.pos
	phase := false
	for p.pos < len(p.text) && p.text[p.pos] == '!' {
		phase = !phase
		p.pos++
	}
	hexStart := p.pos
	for p.pos < len(p.text) && ((p.text[p.pos] >= '0' && p.text[p.pos] <= '9') || (p.text[p.pos] >= 'A' && p.text[p.pos] <= 'F')) {
		p.pos++
	}
	hex := p.text[hexStart:p.pos]
	if p.pos == len(p.text) && len(hex) == 1 && (hex == "0" || hex == "1") {
		return &dssExpr{kind: DAU_DSD_CONST0, phase: phase != (hex == "1"), start: start, open: hexStart, end: p.pos}, nil
	}
	if len(hex) > 0 && p.pos < len(p.text) && p.text[p.pos] != '{' {
		return nil, fmt.Errorf("hex truth must precede a prime block at byte %d", p.pos)
	}
	if p.pos >= len(p.text) {
		return nil, fmt.Errorf("incomplete DSD expression")
	}
	c := p.text[p.pos]
	if c >= 'a' && c <= 'z' {
		p.pos++
		return &dssExpr{kind: DAU_DSD_VAR, variable: int32(c - 'a'), phase: phase, start: start, open: p.pos - 1, end: p.pos}, nil
	}
	close := byte(0)
	kind := DAU_DSD_NONE
	switch c {
	case '(':
		kind = DAU_DSD_AND
		close = ')'
	case '[':
		kind = DAU_DSD_XOR
		close = ']'
	case '<':
		kind = DAU_DSD_MUX
		close = '>'
	case '{':
		kind = DAU_DSD_PRIME
		close = '}'
	default:
		return nil, fmt.Errorf("invalid DSD token %q at byte %d", c, p.pos)
	}
	e := &dssExpr{kind: kind, phase: phase, start: start, open: p.pos}
	p.pos++
	for p.pos < len(p.text) && p.text[p.pos] != close {
		child, err := p.parse()
		if err != nil {
			return nil, err
		}
		e.children = append(e.children, child)
	}
	if p.pos >= len(p.text) {
		return nil, fmt.Errorf("unclosed DSD block at byte %d", e.open)
	}
	p.pos++
	e.end = p.pos
	if len(e.children) < 2 || len(e.children) > DAU_MAX_VAR || kind == DAU_DSD_MUX && len(e.children) != 3 {
		return nil, fmt.Errorf("invalid DSD fanin count %d", len(e.children))
	}
	if kind == DAU_DSD_PRIME {
		var err error
		if hex != "" {
			e.truth, err = dssReadHex(hex, int32(len(e.children)))
			if err != nil {
				return nil, err
			}
		} else {
			if len(p.fallback) < dssWords(int32(len(e.children))) {
				return nil, fmt.Errorf("prime truth is missing")
			}
			e.truth = append([]uint64(nil), p.fallback[:dssWords(int32(len(e.children)))]...)
		}
	}
	return e, nil
}
func dssParse(text string, fallback []uint64) (*dssExpr, error) {
	p := dssParser{text: strings.TrimSpace(text), fallback: fallback}
	e, err := p.parse()
	if err != nil {
		return nil, err
	}
	if p.pos != len(p.text) {
		return nil, fmt.Errorf("trailing DSD data at byte %d", p.pos)
	}
	return e, nil
}
func dssClone(e *dssExpr) *dssExpr {
	c := *e
	c.truth = append([]uint64(nil), e.truth...)
	c.children = make([]*dssExpr, len(e.children))
	for i, child := range e.children {
		c.children[i] = dssClone(child)
	}
	return &c
}
func dssExprString(e *dssExpr) string {
	var b strings.Builder
	if e.phase {
		b.WriteByte('!')
	}
	if e.kind == DAU_DSD_CONST0 {
		if e.phase {
			return "1"
		}
		return "0"
	}
	if e.kind == DAU_DSD_VAR {
		b.WriteByte(byte('a' + e.variable))
		return b.String()
	}
	opens, closes := "???([<{", "???)]>}"
	if e.kind == DAU_DSD_PRIME {
		b.WriteString(dssTruthHex(e.truth, len(e.children)))
	}
	b.WriteByte(opens[e.kind])
	for _, c := range e.children {
		b.WriteString(dssExprString(c))
	}
	b.WriteByte(closes[e.kind])
	return b.String()
}
func ParseDssNetwork(text string, nVars int32, primeTruth []uint64) (*Dss_Ntk_t, error) {
	e, err := dssParse(text, primeTruth)
	if err != nil {
		return nil, err
	}
	p := Dss_NtkAlloc(nVars)
	var build func(*dssExpr) (int32, error)
	build = func(e *dssExpr) (int32, error) {
		phase := dauBool(e.phase)
		if e.kind == DAU_DSD_CONST0 {
			return phase, nil
		}
		if e.kind == DAU_DSD_VAR {
			if e.variable >= nVars {
				return 0, fmt.Errorf("variable outside network support")
			}
			return (e.variable+1)*2 + phase, nil
		}
		fans := make([]int32, len(e.children))
		for i, c := range e.children {
			lit, err := build(c)
			if err != nil {
				return 0, err
			}
			fans[i] = lit
		}
		truth := append([]uint64(nil), e.truth...)
		if e.kind == DAU_DSD_PRIME {
			perm := make([]byte, len(fans))
			canonPhase := canon.Canonicalize(truth, int32(len(fans)), perm)
			old := append([]int32(nil), fans...)
			for i, v := range perm {
				fans[i] = old[v] ^ int32(canonPhase>>uint(i)&1)
			}
			phase ^= int32(canonPhase >> uint(len(fans)) & 1)
		}
		o := Dss_ObjCreateNtk(p, e.kind, fans)
		copy(o.Truth, truth)
		return o.ID*2 + phase, nil
	}
	lit, err := build(e)
	if err != nil {
		return nil, err
	}
	p.Root = Dss_Lit2Obj(p.Objects, lit)
	return p, nil
}
func Dss_NtkCreate(text string, nVars int32, primeTruth []uint64) *Dss_Ntk_t {
	p, err := ParseDssNetwork(text, nVars, primeTruth)
	if err != nil {
		panic(err)
	}
	return p
}
func dssWriteNode(w io.Writer, objects []*Dss_Obj_t, o *Dss_Obj_t, leaf func(*Dss_Obj_t) string) error {
	phase := Dss_IsComplement(o)
	o = Dss_Regular(o)
	if o.Type == DAU_DSD_CONST0 {
		_, err := fmt.Fprint(w, phase)
		return err
	}
	if phase != 0 {
		if _, err := io.WriteString(w, "!"); err != nil {
			return err
		}
	}
	if o.Type == DAU_DSD_VAR {
		_, err := io.WriteString(w, leaf(o))
		return err
	}
	opens, closes := "???([<{", "???)]>}"
	if o.Type == DAU_DSD_PRIME {
		if _, err := io.WriteString(w, dssTruthHex(o.Truth, len(o.Fans))); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "%c", opens[o.Type]); err != nil {
		return err
	}
	for _, lit := range o.Fans {
		if err := dssWriteNode(w, objects, Dss_Lit2Obj(objects, lit), leaf); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "%c", closes[o.Type])
	return err
}
