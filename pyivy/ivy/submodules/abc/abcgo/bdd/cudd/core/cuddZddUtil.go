package core

import (
	"fmt"
	"io"
	"unsafe"
)

// Mirrors sizeof(DdNode) from cudd.h (including the exported 64-bit Id).
func cuddNodeSize() uintptr {
	word := unsafe.Sizeof(uintptr(0))
	if word == 8 {
		return 5 * word
	}
	return 6 * word
}
func Cudd_zddNodeSize() uintptr { return cuddNodeSize() }

func Cudd_zddOutput(dd *DdManager) io.Writer {
	if dd == nil || dd.stdout == nil {
		return nil
	}
	return dd.stdout
}

// Cudd_zddCoverPathToString converts the paired ZDD path encoding used for covers.
func Cudd_zddCoverPathToString(dd *DdManager, path []int, str []byte) []byte {
	if dd == nil || dd.ZddSize&1 != 0 || len(path) < dd.ZddSize {
		return nil
	}
	n := dd.ZddSize / 2
	if str == nil {
		str = make([]byte, n+1)
	} else if len(str) < n+1 {
		return nil
	}
	for i := 0; i < n; i++ {
		switch (path[2*i] << 2) | path[2*i+1] {
		case 0, 2, 8, 10:
			str[i] = '-'
		case 1, 9:
			str[i] = '0'
		case 4, 6:
			str[i] = '1'
		default:
			str[i] = '?'
		}
	}
	str[n] = 0
	return str
}

// Cudd_zddDumpDot writes the CUDD ZDD DOT layout, retaining its support and rank rules.
func Cudd_zddDumpDot(dd *DdManager, roots []*DdNode, inames, onames []string, w io.Writer) int {
	if dd == nil || w == nil || len(roots) == 0 || len(dd.InvPermZ) != dd.ZddSize {
		return 0
	}
	support := make([]bool, dd.ZddSize)
	seen := map[*DdNode]bool{}
	nodes := make([]*DdNode, 0)
	var visit func(*DdNode) bool
	visit = func(edge *DdNode) bool {
		if edge == nil {
			return false
		}
		n := Cudd_Regular(edge)
		if seen[n] {
			return true
		}
		seen[n] = true
		nodes = append(nodes, n)
		if !Cudd_IsConstant(n) {
			if int(n.Index) >= len(support) {
				return false
			}
			support[n.Index] = true
			return visit(n.T) && visit(n.E)
		}
		return true
	}
	for _, r := range roots {
		if !visit(r) {
			return 0
		}
	}
	var diff uintptr
	first := uintptr(unsafe.Pointer(roots[0]))
	for _, n := range nodes {
		diff |= first ^ uintptr(unsafe.Pointer(n))
	}
	bits := uint(unsafe.Sizeof(uintptr(0)) * 8)
	mask := ^uintptr(0)
	for i := uint(0); i < bits; i += 4 {
		m := (uintptr(1) << i) - 1
		mask = m
		if diff <= m {
			break
		}
	}
	id := func(n *DdNode) uintptr { return (uintptr(unsafe.Pointer(n)) & mask) / cuddNodeSize() }
	p := func(n *DdNode) string { return fmt.Sprintf("%#x", id(n)) }
	write := func(format string, args ...any) bool { _, e := fmt.Fprintf(w, format, args...); return e == nil }
	if !write("digraph \"ZDD\" {\nsize = \"7.5,10\"\ncenter = true;\nedge [dir = none];\n") {
		return 0
	}
	if !write("{ node [shape = plaintext];\n  edge [style = invis];\n  \"CONST NODES\" [style = invis];\n") {
		return 0
	}
	for level := 0; level < dd.ZddSize; level++ {
		v := dd.InvPermZ[level]
		if v < 0 || v >= len(support) {
			return 0
		}
		if support[v] {
			name := fmt.Sprint(v)
			if v < len(inames) && inames[v] != "" {
				name = inames[v]
			}
			if !write("\" %s \" -> ", name) {
				return 0
			}
		}
	}
	if !write("\"CONST NODES\"; \n}\n{ rank = same; node [shape = box]; edge [style = invis];\n") {
		return 0
	}
	for i := range roots {
		if i < len(onames) && onames[i] != "" {
			if !write("\"  %s  \"", onames[i]) {
				return 0
			}
		} else if !write("\"F%d\"", i) {
			return 0
		}
		if i == len(roots)-1 {
			if !write("; }\n") {
				return 0
			}
		} else if !write(" -> ") {
			return 0
		}
	}
	for level := 0; level < dd.ZddSize; level++ {
		v := dd.InvPermZ[level]
		if v < 0 || v >= len(support) {
			return 0
		}
		if !support[v] {
			continue
		}
		name := fmt.Sprint(v)
		if v < len(inames) && inames[v] != "" {
			name = inames[v]
		}
		if !write("{ rank = same; \" %s \";\n", name) {
			return 0
		}
		for _, n := range nodes {
			if !Cudd_IsConstant(n) && int(n.Index) == v {
				if !write("\"%s\";\n", p(n)) {
					return 0
				}
			}
		}
		if !write("}\n") {
			return 0
		}
	}
	if !write("{ rank = same; \"CONST NODES\";\n{ node [shape = box]; ") {
		return 0
	}
	for _, n := range nodes {
		if Cudd_IsConstant(n) {
			if !write("\"%s\";\n", p(n)) {
				return 0
			}
		}
	}
	if !write("}\n}\n") {
		return 0
	}
	for i, r := range roots {
		if i < len(onames) && onames[i] != "" {
			if !write("\"  %s  \"", onames[i]) {
				return 0
			}
		} else if !write("\"F%d\"", i) {
			return 0
		}
		if !write(" -> \"%s\" [style = solid];\n", p(r)) {
			return 0
		}
	}
	for _, n := range nodes {
		if Cudd_IsConstant(n) {
			continue
		}
		if !write("\"%s\" -> \"%s\";\n\"%s\" -> \"%s\" [style = dashed];\n", p(n), p(Cudd_T(n)), p(n), p(Cudd_E(n))) {
			return 0
		}
	}
	for _, n := range nodes {
		if Cudd_IsConstant(n) {
			if !write("\"%s\" [label = \"%g\"];\n", p(n), n.Value) {
				return 0
			}
		}
	}
	if !write("}\n") {
		return 0
	}
	return 1
}
