package dec

import (
	"fmt"
	"os"
)

func Dec_GraphPrint(file *os.File, g *Dec_Graph_t, names []string, out string) {
	if names == nil {
		names = make([]string, g.nLeaves)
		for i := range names {
			names[i] = fmt.Sprintf("x%d", i)
		}
	}
	if len(names) < int(g.nLeaves) {
		panic("dec: missing input names")
	}
	if out == "" {
		out = "F"
	}
	width := int32(0)
	for _, name := range names {
		if int32(len(name)) > width {
			width = int32(len(name))
		}
	}
	if width > 50 {
		width = 20
	}
	pos := Dec_GraphPrintOutputName(file, out)
	if g.fConst != 0 {
		fmt.Fprintf(file, "Constant %d", 1-g.eRoot.fCompl)
	} else if Dec_GraphIsVar(g) != 0 {
		Dec_GraphPrintGetLeafName(file, int32(g.eRoot.Node), int32(g.eRoot.fCompl), names)
	} else {
		Dec_GraphPrint_rec(file, g, Dec_GraphNode(g, int32(g.eRoot.Node)), int32(g.eRoot.fCompl), names, &pos, width)
	}
	fmt.Fprintln(file)
}
func Dec_GraphPrintGetLeafName(file *os.File, index, compl int32, names []string) int32 {
	name := names[index]
	if compl != 0 {
		name = "!" + name
	}
	fmt.Fprint(file, name)
	return int32(len(name))
}
func Dec_GraphPrintOutputName(file *os.File, name string) int32 {
	if name == "" {
		return 0
	}
	fmt.Fprintf(file, "%6s = ", name)
	return 10
}
func Dec_GraphPrintUpdatePos(file *os.File, pos *int32, width int32) {
	if *pos+width < 77 {
		return
	}
	fmt.Fprint(file, "\n          ")
	*pos = 10
}
func Dec_GraphPrint_rec(file *os.File, g *Dec_Graph_t, n *Dec_Node_t, compl int32, names []string, pos *int32, width int32) {
	if Dec_GraphNodeIsVar(g, n) != 0 {
		*pos += Dec_GraphPrintGetLeafName(file, Dec_GraphNodeInt(g, n), compl, names)
		return
	}
	a, b := Dec_GraphNode(g, int32(n.eEdge0.Node)), Dec_GraphNode(g, int32(n.eEdge1.Node))
	fmt.Fprint(file, "(")
	*pos++
	if compl != 0 {
		Dec_GraphPrint_rec(file, g, a, 1-int32(n.eEdge0.fCompl), names, pos, width)
		fmt.Fprint(file, " + ")
		*pos += 3
		Dec_GraphPrint_rec(file, g, b, 1-int32(n.eEdge1.fCompl), names, pos, width)
	} else {
		Dec_GraphPrint_rec(file, g, a, int32(n.eEdge0.fCompl), names, pos, width)
		Dec_GraphPrint_rec(file, g, b, int32(n.eEdge1.fCompl), names, pos, width)
	}
	fmt.Fprint(file, ")")
	*pos++
}

// The alternate printer uses original OR annotations for a less parenthesized
// factored form while retaining the same Boolean function.
func Dec_GraphPrint2_rec(file *os.File, g *Dec_Graph_t, n *Dec_Node_t, compl int32, names []string, pos *int32, width int32) {
	if Dec_GraphNodeIsVar(g, n) != 0 {
		*pos += Dec_GraphPrintGetLeafName(file, Dec_GraphNodeInt(g, n), compl, names)
		return
	}
	a, b := Dec_GraphNode(g, int32(n.eEdge0.Node)), Dec_GraphNode(g, int32(n.eEdge1.Node))
	printChild := func(child *Dec_Node_t, c uint32, bracket bool) {
		if bracket {
			fmt.Fprint(file, "(")
			*pos++
		}
		Dec_GraphPrint_rec(file, g, child, int32(c), names, pos, width)
		if bracket {
			fmt.Fprint(file, ")")
			*pos++
		}
	}
	printChild(a, n.fCompl0, n.fNodeOr == 0 && a.fNodeOr != 0)
	if n.fNodeOr != 0 {
		fmt.Fprint(file, " + ")
		*pos += 3
	} else {
		fmt.Fprint(file, " ")
		*pos++
	}
	Dec_GraphPrintUpdatePos(file, pos, width)
	printChild(b, n.fCompl1, n.fNodeOr == 0 && b.fNodeOr != 0)
}
