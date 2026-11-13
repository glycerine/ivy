package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// cec4Pattern owns the temporary assignments which the C implementation
// stores in the GIA object's mark bits while constructing one counterexample.
type cec4Pattern struct {
	graph    *gia.Gia_Man_t
	values   []int8 // -1 unassigned; otherwise 0 or 1
	literals []int32
}

func (p *cec4Pattern) fan(id int32, which int) (int32, int32) {
	obj := gia.Gia_ManObj(p.graph, id)
	if which == 0 {
		return gia.Gia_ObjFaninId0(obj, id), gia.Gia_ObjFaninC0(obj)
	}
	return gia.Gia_ObjFaninId1(obj, id), gia.Gia_ObjFaninC1(obj)
}
func (p *cec4Pattern) has(id int32, which int, value int32) bool {
	fan, complement := p.fan(id, which)
	return int32(p.values[fan]) == value^complement
}
func (p *cec4Pattern) implied(id int32, which int, value int32) bool {
	fan, complement := p.fan(id, which)
	obj := gia.Gia_ManObj(p.graph, fan)
	if gia.Gia_ObjIsAnd(obj) == 0 || p.values[fan] != -1 {
		return false
	}
	if value^complement == 1 {
		return p.has(fan, 0, 1) && p.has(fan, 1, 1)
	}
	return p.has(fan, 0, 0) || p.has(fan, 1, 0)
}
func (p *cec4Pattern) generate(id, value int32) bool {
	if p.values[id] != -1 {
		panic("cec4: revisited assigned node")
	}
	p.values[id] = int8(value)
	obj := gia.Gia_ManObj(p.graph, id)
	if gia.Gia_ObjIsCi(obj) != 0 {
		p.literals = append(p.literals, id<<1|value)
		return true
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("cec4: unsupported pattern node")
	}
	a, ca := p.fan(id, 0)
	b, cb := p.fan(id, 1)
	if gia.Gia_ObjIsXor(obj) != 0 {
		assignedA, assignedB := p.values[a] != -1, p.values[b] != -1
		if assignedA && assignedB {
			return value == int32(p.values[a])^int32(p.values[b])
		}
		if assignedA {
			return p.generate(b, value^int32(p.values[a]))
		}
		if assignedB {
			return p.generate(a, value^int32(p.values[b]))
		}
		first := int32(util.Abc_Random(0) & 1)
		if !p.generate(a, first) {
			return false
		}
		want := value ^ first
		if p.values[b] != -1 {
			return int32(p.values[b]) == want
		}
		return p.generate(b, want)
	}
	if value == 1 {
		if p.has(id, 0, 0) || p.has(id, 1, 0) {
			return false
		}
		if !p.has(id, 0, 1) && !p.generate(a, 1^ca) {
			return false
		}
		return p.has(id, 1, 1) || p.generate(b, 1^cb)
	}
	if p.has(id, 0, 1) && p.has(id, 1, 1) {
		return false
	}
	if p.has(id, 0, 0) || p.has(id, 1, 0) {
		return true
	}
	if p.has(id, 0, 1) {
		return p.generate(b, cb)
	}
	if p.has(id, 1, 1) {
		return p.generate(a, ca)
	}
	switch {
	case p.implied(id, 0, 0):
		return p.generate(a, ca)
	case p.implied(id, 1, 0):
		return p.generate(b, cb)
	case p.implied(id, 0, 1):
		return p.generate(b, cb)
	case p.implied(id, 1, 1):
		return p.generate(a, ca)
	case util.Abc_Random(0)&1 != 0:
		return p.generate(b, cb)
	default:
		return p.generate(a, ca)
	}
}

// Cec4_ManGeneratePatternOne returns a partial PI assignment satisfying both
// requested object values, or zero when this non-backtracking attempt fails.
func Cec4_ManGeneratePatternOne(g *gia.Gia_Man_t, repr, reprValue, candidate, candidateValue int32, pattern, visited *vec.Vec_Int_t) int32 {
	if candidate <= 0 {
		panic("cec4: invalid candidate")
	}
	if repr == 0 && reprValue != 0 {
		return 0
	}
	vec.Vec_IntClear(pattern)
	vec.Vec_IntClear(visited)
	p := cec4Pattern{graph: g, values: make([]int8, gia.Gia_ManObjNum(g))}
	for i := range p.values {
		p.values[i] = -1
	}
	ok := (repr == 0 || p.generate(repr, reprValue))
	if ok {
		if p.values[candidate] != -1 {
			ok = int32(p.values[candidate]) == candidateValue
		} else {
			ok = p.generate(candidate, candidateValue)
		}
	}
	for _, lit := range p.literals {
		vec.Vec_IntPush(pattern, lit)
	}
	for id, value := range p.values {
		if value != -1 {
			vec.Vec_IntPush(visited, int32(id))
		}
	}
	if ok {
		return 1
	}
	return 0
}
