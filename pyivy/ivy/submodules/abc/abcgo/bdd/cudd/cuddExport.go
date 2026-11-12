package cudd

import (
	"fmt"
	"io"
	"os"
	"sort"
)

func Cudd_DumpDot(dd *DdManager, roots []*DdNode, inputNames, outputNames []string, writer io.Writer) int32 {
	if dd == nil || len(roots) == 0 {
		return 0
	}
	if writer == nil {
		writer = os.Stdout
	}
	nodes, ok := collectDumpNodes(dd, roots)
	if !ok {
		return 0
	}
	if _, err := fmt.Fprintln(writer, "digraph DD {"); err != nil {
		return 0
	}
	if _, err := fmt.Fprintln(writer, "  rankdir=TB;"); err != nil {
		return 0
	}
	for _, node := range nodes {
		if Cudd_IsConstant(node) {
			continue
		}
		if _, err := fmt.Fprintf(writer, "  n%d [label=\"%s\",shape=box];\n", node.ID, dumpVarName(inputNames, int(node.Index))); err != nil {
			return 0
		}
	}
	if _, err := fmt.Fprintln(writer, "  one [label=\"1\",shape=box];\n  zero [label=\"0\",shape=box];"); err != nil {
		return 0
	}
	for _, node := range nodes {
		if Cudd_IsConstant(node) {
			continue
		}
		for branch, child := range []*DdNode{Cudd_T(node), Cudd_E(node)} {
			style, label := "solid", "1"
			if branch == 1 {
				style, label = "dashed", "0"
			}
			if Cudd_IsComplement(child) {
				style = "dotted"
			}
			if _, err := fmt.Fprintf(writer, "  n%d -> %s [style=%s,label=\"%s\"];\n", node.ID, dumpNodeName(dd, child), style, label); err != nil {
				return 0
			}
		}
	}
	for i, root := range roots {
		name := dumpOutputName(outputNames, i)
		if _, err := fmt.Fprintf(writer, "  out%d [label=\"%s\",shape=plaintext];\n  out%d -> %s [style=%s];\n", i, name, i, dumpNodeName(dd, root), map[bool]string{true: "dotted", false: "solid"}[Cudd_IsComplement(root)]); err != nil {
			return 0
		}
	}
	if _, err := fmt.Fprintln(writer, "}"); err != nil {
		return 0
	}
	return 1
}

func Cudd_DumpBlif(dd *DdManager, roots []*DdNode, inputNames, outputNames []string, model string, writer io.Writer, multiValued int32) int32 {
	if dd == nil || len(roots) == 0 || multiValued != 0 {
		return 0
	}
	if writer == nil {
		writer = os.Stdout
	}
	if model == "" {
		model = "DD"
	}
	if _, err := fmt.Fprintf(writer, ".model %s\n.inputs", model); err != nil {
		return 0
	}
	for i := range dd.Vars {
		if _, err := fmt.Fprintf(writer, " %s", dumpVarName(inputNames, i)); err != nil {
			return 0
		}
	}
	if _, err := fmt.Fprint(writer, "\n.outputs"); err != nil {
		return 0
	}
	for i := range roots {
		if _, err := fmt.Fprintf(writer, " %s", dumpOutputName(outputNames, i)); err != nil {
			return 0
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return 0
	}
	if Cudd_DumpBlifBody(dd, roots, inputNames, outputNames, writer, multiValued) == 0 {
		return 0
	}
	_, err := fmt.Fprintln(writer, ".end")
	if err != nil {
		return 0
	}
	return 1
}

func Cudd_DumpBlifBody(dd *DdManager, roots []*DdNode, inputNames, outputNames []string, writer io.Writer, multiValued int32) int32 {
	if dd == nil || len(roots) == 0 || writer == nil || multiValued != 0 {
		return 0
	}
	nodes, ok := collectDumpNodes(dd, roots)
	if !ok {
		return 0
	}
	if _, err := fmt.Fprintln(writer, ".names one\n1\n.names zero"); err != nil {
		return 0
	}
	for _, node := range nodes {
		if Cudd_IsConstant(node) {
			continue
		}
		t, e := dumpBlifSignal(dd, Cudd_T(node)), dumpBlifSignal(dd, Cudd_E(node))
		if _, err := fmt.Fprintf(writer, ".names %s %s %s %s\n1 1 - 1\n0 - 1 1\n", dumpVarName(inputNames, int(node.Index)), t, e, dumpNodeName(dd, node)); err != nil {
			return 0
		}
	}
	for _, node := range nodes {
		if Cudd_IsConstant(node) {
			continue
		}
		if _, err := fmt.Fprintf(writer, ".names %s %s\n0 1\n", dumpNodeName(dd, node), dumpNodeName(dd, node)+"_not"); err != nil {
			return 0
		}
	}
	for i, root := range roots {
		signal := dumpBlifSignal(dd, root)
		if _, err := fmt.Fprintf(writer, ".names %s %s\n1 1\n", signal, dumpOutputName(outputNames, i)); err != nil {
			return 0
		}
	}
	return 1
}

func Cudd_DumpDaVinci(dd *DdManager, roots []*DdNode, inputNames, outputNames []string, writer io.Writer) int32 {
	if dd == nil || len(roots) == 0 {
		return 0
	}
	if writer == nil {
		writer = os.Stdout
	}
	if _, err := fmt.Fprint(writer, "["); err != nil {
		return 0
	}
	for i, root := range roots {
		if i != 0 {
			if _, err := fmt.Fprint(writer, ","); err != nil {
				return 0
			}
		}
		name := dumpOutputName(outputNames, i)
		color := "blue"
		if Cudd_IsComplement(root) {
			color = "red"
		}
		if _, err := fmt.Fprintf(writer, "l(%q,n(\"root\",[a(\"OBJECT\",%q)],[e(\"edge\",[a(\"EDGECOLOR\",%q),a(\"_DIR\",\"none\")],", name, name, color); err != nil {
			return 0
		}
		if !dumpDaVinciNode(dd, Cudd_Regular(root), inputNames, writer, map[*DdNode]bool{}) {
			return 0
		}
		if _, err := fmt.Fprint(writer, ")]))"); err != nil {
			return 0
		}
	}
	if _, err := fmt.Fprintln(writer, "]"); err != nil {
		return 0
	}
	return 1
}

func Cudd_DumpDDcal(dd *DdManager, roots []*DdNode, inputNames, outputNames []string, writer io.Writer) int32 {
	if dd == nil || len(roots) == 0 {
		return 0
	}
	if writer == nil {
		writer = os.Stdout
	}
	nodes, ok := collectDumpNodes(dd, roots)
	if !ok {
		return 0
	}
	for i := range dd.Vars {
		if i != 0 {
			if _, err := fmt.Fprint(writer, " * "); err != nil {
				return 0
			}
		}
		if _, err := fmt.Fprint(writer, dumpVarName(inputNames, i)); err != nil {
			return 0
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return 0
	}
	for _, node := range nodes {
		if Cudd_IsConstant(node) {
			if _, err := fmt.Fprintf(writer, "n%d = %g\n", node.ID, node.Value); err != nil {
				return 0
			}
			continue
		}
		t := dumpNodeName(dd, Cudd_T(node))
		e := dumpNodeName(dd, Cudd_E(node))
		if Cudd_IsComplement(Cudd_E(node)) {
			e += "'"
		}
		if _, err := fmt.Fprintf(writer, "n%d = %s * %s + %s' * %s\n", node.ID, dumpVarName(inputNames, int(node.Index)), t, dumpVarName(inputNames, int(node.Index)), e); err != nil {
			return 0
		}
	}
	for i, root := range roots {
		suffix := ""
		if Cudd_IsComplement(root) {
			suffix = "'"
		}
		if _, err := fmt.Fprintf(writer, "%s = %s%s\n", dumpOutputName(outputNames, i), dumpNodeName(dd, root), suffix); err != nil {
			return 0
		}
	}
	if _, err := fmt.Fprint(writer, "["); err != nil {
		return 0
	}
	for i := range roots {
		if i != 0 {
			if _, err := fmt.Fprint(writer, " "); err != nil {
				return 0
			}
		}
		if _, err := fmt.Fprint(writer, dumpOutputName(outputNames, i)); err != nil {
			return 0
		}
	}
	if _, err := fmt.Fprintln(writer, "]"); err != nil {
		return 0
	}
	return 1
}

func Cudd_DumpFactoredForm(dd *DdManager, roots []*DdNode, inputNames, outputNames []string, writer io.Writer) int32 {
	return dumpExpressions(dd, roots, inputNames, outputNames, writer, "factored")
}

func dumpExpressions(dd *DdManager, roots []*DdNode, inputNames, outputNames []string, writer io.Writer, format string) int32 {
	if dd == nil || len(roots) == 0 {
		return 0
	}
	if writer == nil {
		writer = os.Stdout
	}
	for i, root := range roots {
		expression := dumpExpression(dd, root, inputNames, make(map[*DdNode]string))
		if expression == "" {
			return 0
		}
		if _, err := fmt.Fprintf(writer, "%s = %s\n", dumpOutputName(outputNames, i), expression); err != nil {
			return 0
		}
	}
	return 1
}

func dumpDaVinciNode(dd *DdManager, node *DdNode, names []string, writer io.Writer, visited map[*DdNode]bool) bool {
	id := fmt.Sprintf("n%d", node.ID)
	if visited[node] {
		_, err := fmt.Fprintf(writer, "r(%q)", id)
		return err == nil
	}
	visited[node] = true
	if Cudd_IsConstant(node) {
		_, err := fmt.Fprintf(writer, "l(%q,n(\"constant\",[a(\"OBJECT\",%q)],[]))", id, fmt.Sprint(node.Value))
		return err == nil
	}
	if _, err := fmt.Fprintf(writer, "l(%q,n(\"internal\",[a(\"OBJECT\",%q),a(\"_GO\",\"ellipse\")],[e(\"then\",[a(\"EDGECOLOR\",\"blue\"),a(\"_DIR\",\"none\")],", id, dumpVarName(names, int(node.Index))); err != nil {
		return false
	}
	if !dumpDaVinciEdge(dd, Cudd_T(node), names, writer, visited) {
		return false
	}
	color := "green"
	if Cudd_IsComplement(Cudd_E(node)) {
		color = "red"
	}
	if _, err := fmt.Fprintf(writer, "),e(\"else\",[a(\"EDGECOLOR\",%q),a(\"_DIR\",\"none\")],", color); err != nil {
		return false
	}
	if !dumpDaVinciEdge(dd, Cudd_E(node), names, writer, visited) {
		return false
	}
	_, err := fmt.Fprint(writer, ")] ))")
	return err == nil
}

func dumpDaVinciEdge(dd *DdManager, edge *DdNode, names []string, writer io.Writer, visited map[*DdNode]bool) bool {
	if Cudd_IsComplement(edge) {
		if _, err := fmt.Fprint(writer, "l(\"complement\",n(\"not\",[],["); err != nil {
			return false
		}
		if !dumpDaVinciNode(dd, Cudd_Regular(edge), names, writer, visited) {
			return false
		}
		_, err := fmt.Fprint(writer, "]))")
		return err == nil
	}
	return dumpDaVinciNode(dd, edge, names, writer, visited)
}

func dumpExpression(dd *DdManager, node *DdNode, names []string, memo map[*DdNode]string) string {
	if node == Cudd_ReadOne(dd) {
		return "CONST1"
	}
	if node == Cudd_ReadLogicZero(dd) {
		return "CONST0"
	}
	if cached, ok := memo[node]; ok {
		return cached
	}
	regular := Cudd_Regular(node)
	variable := dumpVarName(names, int(regular.Index))
	t := dumpExpression(dd, Cudd_T(regular), names, memo)
	e := dumpExpression(dd, Cudd_E(regular), names, memo)
	result := "((" + variable + " & " + t + ") | (!" + variable + " & " + e + "))"
	if t == "CONST0" {
		result = "(!" + variable + " & " + e + ")"
	} else if e == "CONST0" {
		result = "(" + variable + " & " + t + ")"
	} else if t == "CONST1" {
		result = "(" + variable + " | " + e + ")"
	} else if e == "CONST1" {
		result = "(!" + variable + " | " + t + ")"
	} else if Cudd_T(regular) == Cudd_E(regular) {
		result = t
	}
	if Cudd_IsComplement(node) {
		result = "!(" + result + ")"
	}
	memo[node] = result
	return result
}

func collectDumpNodes(dd *DdManager, roots []*DdNode) ([]*DdNode, bool) {
	seen := map[*DdNode]bool{}
	var nodes []*DdNode
	var visit func(*DdNode) bool
	visit = func(edge *DdNode) bool {
		if edge == nil {
			return false
		}
		node := Cudd_Regular(edge)
		if seen[node] || Cudd_IsConstant(node) {
			return true
		}
		seen[node] = true
		if !visit(Cudd_T(node)) || !visit(Cudd_E(node)) {
			return false
		}
		nodes = append(nodes, node)
		return true
	}
	for _, root := range roots {
		if !visit(root) {
			return nil, false
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return nodes, true
}

func dumpNodeName(dd *DdManager, edge *DdNode) string {
	if edge == Cudd_ReadOne(dd) {
		return "one"
	}
	if edge == Cudd_ReadLogicZero(dd) {
		return "zero"
	}
	return fmt.Sprintf("n%d", Cudd_Regular(edge).ID)
}

func dumpBlifSignal(dd *DdManager, edge *DdNode) string {
	if edge == Cudd_ReadOne(dd) {
		return "one"
	}
	if edge == Cudd_ReadLogicZero(dd) {
		return "zero"
	}
	name := dumpNodeName(dd, edge)
	if Cudd_IsComplement(edge) {
		return name + "_not"
	}
	return name
}

func dumpVarName(names []string, index int) string {
	if index >= 0 && index < len(names) && names[index] != "" {
		return names[index]
	}
	return fmt.Sprintf("x%d", index)
}

func dumpOutputName(names []string, index int) string {
	if index >= 0 && index < len(names) && names[index] != "" {
		return names[index]
	}
	return fmt.Sprintf("f%d", index)
}
