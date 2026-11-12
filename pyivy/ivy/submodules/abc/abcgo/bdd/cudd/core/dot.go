package core

import (
	"fmt"
	"io"
	"sort"
	"strconv"
)

// WriteDOT writes BDD or ADD roots, preserving complemented-edge information.
func WriteDOT(w io.Writer, dd *DdManager, roots []*DdNode, inputNames, outputNames []string) error {
	if w == nil || dd == nil || len(roots) == 0 {
		return fmt.Errorf("cudd: DOT export needs a manager, roots, and writer")
	}
	seen := make(map[*DdNode]bool)
	var nodes []*DdNode
	var visit func(*DdNode) error
	visit = func(edge *DdNode) error {
		if !dd.valid(edge) && !dd.validAdd(edge) {
			return fmt.Errorf("cudd: invalid DOT root or child")
		}
		node := Cudd_Regular(edge)
		if seen[node] {
			return nil
		}
		seen[node] = true
		if !Cudd_IsConstant(node) {
			if err := visit(node.T); err != nil {
				return err
			}
			if err := visit(node.E); err != nil {
				return err
			}
		}
		nodes = append(nodes, node)
		return nil
	}
	for _, root := range roots {
		if err := visit(root); err != nil {
			return err
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	if _, err := fmt.Fprintln(w, "digraph DD {\n  rankdir=TB;"); err != nil {
		return err
	}
	for _, node := range nodes {
		label := strconv.FormatFloat(node.Value, 'g', -1, 64)
		if !Cudd_IsConstant(node) {
			i := int(node.Index)
			label = fmt.Sprintf("x%d", i)
			if i < len(inputNames) && inputNames[i] != "" {
				label = inputNames[i]
			}
		}
		if _, err := fmt.Fprintf(w, "  n%d [label=%s,shape=box];\n", node.ID, strconv.Quote(label)); err != nil {
			return err
		}
	}
	for _, node := range nodes {
		if Cudd_IsConstant(node) {
			continue
		}
		for i, child := range []*DdNode{node.T, node.E} {
			style, label := "solid", "1"
			if i == 1 {
				style, label = "dashed", "0"
			}
			if Cudd_IsComplement(child) {
				style = "dotted"
			}
			if _, err := fmt.Fprintf(w, "  n%d -> n%d [style=%s,label=%q];\n", node.ID, Cudd_Regular(child).ID, style, label); err != nil {
				return err
			}
		}
	}
	for i, root := range roots {
		name := fmt.Sprintf("f%d", i)
		if i < len(outputNames) && outputNames[i] != "" {
			name = outputNames[i]
		}
		style := "solid"
		if Cudd_IsComplement(root) {
			style = "dotted"
		}
		if _, err := fmt.Fprintf(w, "  out%d [label=%s,shape=plaintext];\n  out%d -> n%d [style=%s];\n", i, strconv.Quote(name), i, Cudd_Regular(root).ID, style); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "}")
	return err
}
