package dau

import (
	"fmt"
	"strings"
)

// Dau_DsdArrayNode is the private dauArray.c bitfield record. Its distinct Go
// name avoids colliding with dauDsd.c's unrelated file-local Dau_Dsd_t manager.
type Dau_DsdArrayNode struct {
	iVar, nFans, Depth, Offset, Data int32
	Type                             Dau_DsdType_t
	fCompl, fUnused                  int32
}

func Dau_DsdClean(p *Dau_DsdArrayNode) { *p = Dau_DsdArrayNode{} }
func Dau_DsdCountAnd(nodes []Dau_DsdArrayNode) int32 {
	costs := [7]int32{0, 0, 0, 1, 3, 3, 3}
	sum := int32(0)
	for _, n := range nodes {
		if n.Type == DAU_DSD_NONE {
			break
		}
		sum += costs[n.Type]
	}
	return sum
}
func Dau_DsdConstruct(text string, store []Dau_DsdArrayNode) int32 {
	if len(store) < 2 {
		panic("short DAU DSD array")
	}
	if Dau_DsdIsConst(text) != 0 {
		store[0] = Dau_DsdArrayNode{Type: DAU_DSD_CONST0, fCompl: Dau_DsdIsConst1(text)}
		store[1] = Dau_DsdArrayNode{}
		return 1
	}
	var levels [DAU_MAX_VAR]int
	var close [DAU_MAX_VAR]byte
	depth, n, compl := -1, 0, int32(0)
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '!' {
			compl ^= 1
			continue
		}
		if c == ')' || c == ']' || c == '>' || c == '}' {
			if depth < 0 || close[depth] != c || compl != 0 {
				panic("unbalanced DAU DSD array expression")
			}
			depth--
			if depth >= 0 {
				parent, child := &store[levels[depth]], &store[levels[depth+1]]
				parent.nFans++
				parent.Data = min(parent.Data, child.Data)
			}
			continue
		}
		if n+1 >= len(store) {
			panic("DAU DSD array capacity exceeded")
		}
		q := &store[n]
		*q = Dau_DsdArrayNode{Data: 31, fCompl: compl}
		compl = 0
		index := n
		n++
		if c >= 'a' && c <= 'z' {
			q.Type = DAU_DSD_VAR
			q.iVar = int32(c - 'a')
			q.Depth = int32(depth + 1)
			if depth >= 0 {
				parent := &store[levels[depth]]
				parent.nFans++
				parent.Data = min(parent.Data, q.iVar)
			}
			continue
		}
		var closing byte
		switch c {
		case '(':
			q.Type = DAU_DSD_AND
			closing = ')'
		case '[':
			q.Type = DAU_DSD_XOR
			closing = ']'
		case '<':
			q.Type = DAU_DSD_MUX
			closing = '>'
		case '{':
			q.Type = DAU_DSD_PRIME
			closing = '}'
		default:
			panic("invalid DAU DSD array token")
		}
		depth++
		if depth >= DAU_MAX_VAR {
			panic("DAU DSD array nesting exceeds variable bound")
		}
		levels[depth] = index
		close[depth] = closing
		q.Depth = int32(depth)
	}
	if depth != -1 || compl != 0 || n == 0 {
		panic("incomplete DAU DSD array expression")
	}
	store[n] = Dau_DsdArrayNode{}
	return int32(n)
}
func DauDsdArrayString(nodes []Dau_DsdArrayNode) string {
	if len(nodes) == 0 {
		return ""
	}
	if nodes[0].Type == DAU_DSD_CONST0 {
		return fmt.Sprint(nodes[0].fCompl)
	}
	open := [7]byte{0, 0, 0, '(', '[', '<', '{'}
	close := [7]byte{0, 0, 0, ')', ']', '>', '}'}
	var result strings.Builder
	var closings [DAU_MAX_VAR + 1]byte
	var visits [DAU_MAX_VAR + 1]int32
	depth := 0
	visits[0] = 1
	for _, node := range nodes {
		if node.Type == DAU_DSD_NONE {
			break
		}
		if node.fCompl != 0 {
			result.WriteByte('!')
		}
		if node.Type == DAU_DSD_VAR {
			result.WriteByte(byte('a' + node.iVar))
			for depth > 0 {
				visits[depth]--
				if visits[depth] != 0 {
					break
				}
				result.WriteByte(closings[depth])
				depth--
			}
		} else {
			depth++
			visits[depth] = node.nFans
			result.WriteByte(open[node.Type])
			result.WriteByte(byte('a' + node.Data))
			fmt.Fprint(&result, node.Depth)
			closings[depth] = close[node.Type]
		}
	}
	if depth != 0 {
		panic("incomplete DAU DSD array fanins")
	}
	return result.String()
}
func Dau_DsdPrint(nodes []Dau_DsdArrayNode) { fmt.Println(DauDsdArrayString(nodes)) }
func Dau_DsdDepth(nodes []Dau_DsdArrayNode) {
	if len(nodes) == 0 || nodes[0].Type == DAU_DSD_CONST0 {
		return
	}
	var visits [DAU_MAX_VAR + 1]int32
	visits[0] = 1
	depth := 0
	for i := range nodes {
		p := &nodes[i]
		if p.Type == DAU_DSD_NONE {
			break
		}
		p.Depth = int32(depth)
		if p.Type == DAU_DSD_VAR {
			for depth > 0 {
				visits[depth]--
				if visits[depth] != 0 {
					break
				}
				depth--
			}
		} else {
			depth++
			visits[depth] = p.nFans
		}
	}
	if depth != 0 {
		panic("incomplete DAU DSD array")
	}
}
func Dau_DsdRemoveUseless(nodes []Dau_DsdArrayNode) {
	if len(nodes) == 0 || nodes[0].Type == DAU_DSD_CONST0 {
		return
	}
	var levels [DAU_MAX_VAR + 1]int
	var visits [DAU_MAX_VAR + 1]int32
	visits[0] = 1
	depth := 0
	changed := false
	for i := range nodes {
		p := &nodes[i]
		if p.Type == DAU_DSD_NONE {
			break
		}
		p.Depth = int32(depth)
		if p.Type == DAU_DSD_VAR {
			for depth > 0 {
				visits[depth]--
				if visits[depth] != 0 {
					break
				}
				depth--
			}
			continue
		}
		flatten := false
		if depth > 0 {
			parent := &nodes[levels[depth-1]]
			flatten = parent.Type == DAU_DSD_XOR && p.Type == DAU_DSD_XOR || parent.Type == DAU_DSD_AND && p.Type == DAU_DSD_AND && p.fCompl == 0
			if flatten {
				parent.nFans += p.nFans - 1
				visits[depth] += p.nFans - 1
				if p.Type == DAU_DSD_XOR {
					parent.fCompl ^= p.fCompl
				}
				p.fUnused = 1
				changed = true
			}
		}
		if !flatten {
			levels[depth] = i
			depth++
			visits[depth] = p.nFans
		}
	}
	if depth != 0 {
		panic("incomplete DAU DSD array")
	}
	if changed {
		write := 0
		for _, n := range nodes {
			if n.Type == DAU_DSD_NONE {
				break
			}
			if n.fUnused == 0 {
				nodes[write] = n
				write++
			}
		}
		nodes[write] = Dau_DsdArrayNode{}
	}
}
func Dau_DsdTest22() {
	nodes := make([]Dau_DsdArrayNode, 2*DAU_MAX_VAR)
	Dau_DsdConstruct("[(abd)[cf](f(sg(he)))]", nodes)
	Dau_DsdPrint(nodes)
	Dau_DsdRemoveUseless(nodes)
	Dau_DsdPrint(nodes)
}
