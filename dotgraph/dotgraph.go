// Package dotgraph provides DOT graph generation and layout for Ivy visualizations.
// This is a port of Python's ivy_graphviz.py (197 lines) and dot_layout.py (286 lines).
//
// Instead of depending on pydot/pygraphviz (Python-specific), this package
// generates DOT format strings directly and can optionally invoke the
// graphviz `dot` command for layout.
package dotgraph

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Graph represents a DOT graph.
type Graph struct {
	Name       string
	Directed   bool
	Strict     bool
	Nodes      []*Node
	Edges      []*Edge
	Subgraphs []*Graph
	Attrs      map[string]string
	NodeAttrs  map[string]string
	EdgeAttrs  map[string]string
}

// Node represents a node in the graph.
type Node struct {
	Name  string
	Attrs map[string]string
}

// Edge represents an edge in the graph.
type Edge struct {
	Src   string
	Dst   string
	Attrs map[string]string
}

// NewGraph creates a new DOT graph.
func NewGraph(name string, directed bool) *Graph {
	return &Graph{
		Name:      name,
		Directed:  directed,
		Attrs:     make(map[string]string),
		NodeAttrs: make(map[string]string),
		EdgeAttrs: make(map[string]string),
	}
}

// NewDigraph creates a new directed graph.
func NewDigraph(name string) *Graph {
	return NewGraph(name, true)
}

// AddNode adds a node to the graph.
func (g *Graph) AddNode(name string, attrs map[string]string) *Node {
	n := &Node{Name: name, Attrs: attrs}
	if n.Attrs == nil {
		n.Attrs = make(map[string]string)
	}
	g.Nodes = append(g.Nodes, n)
	return n
}

// AddEdge adds an edge to the graph.
func (g *Graph) AddEdge(src, dst string, attrs map[string]string) *Edge {
	e := &Edge{Src: src, Dst: dst, Attrs: attrs}
	if e.Attrs == nil {
		e.Attrs = make(map[string]string)
	}
	g.Edges = append(g.Edges, e)
	return e
}

// AddSubgraph adds a subgraph.
func (g *Graph) AddSubgraph(sub *Graph) {
	g.Subgraphs = append(g.Subgraphs, sub)
}

// ToDOT generates the DOT format string.
func (g *Graph) ToDOT() string {
	var b strings.Builder
	g.writeDOT(&b, 0)
	return b.String()
}

// WriteDOT writes the DOT format to a writer.
func (g *Graph) WriteDOT(w io.Writer) error {
	_, err := w.Write([]byte(g.ToDOT()))
	return err
}

func (g *Graph) writeDOT(b *strings.Builder, indent int) {
	prefix := strings.Repeat("  ", indent)

	// Graph type
	if g.Strict {
		b.WriteString(prefix + "strict ")
	}
	if g.Directed {
		b.WriteString(prefix + "digraph")
	} else {
		b.WriteString(prefix + "graph")
	}
	if g.Name != "" {
		b.WriteString(" " + quoteID(g.Name))
	}
	b.WriteString(" {\n")

	// Graph attributes
	for k, v := range g.Attrs {
		fmt.Fprintf(b, "%s  %s=%s;\n", prefix, k, quoteID(v))
	}
	if len(g.NodeAttrs) > 0 {
		b.WriteString(prefix + "  node [")
		writeAttrs(b, g.NodeAttrs)
		b.WriteString("];\n")
	}
	if len(g.EdgeAttrs) > 0 {
		b.WriteString(prefix + "  edge [")
		writeAttrs(b, g.EdgeAttrs)
		b.WriteString("];\n")
	}

	// Subgraphs
	for _, sub := range g.Subgraphs {
		sub.writeDOT(b, indent+1)
	}

	// Nodes
	for _, n := range g.Nodes {
		b.WriteString(prefix + "  " + quoteID(n.Name))
		if len(n.Attrs) > 0 {
			b.WriteString(" [")
			writeAttrs(b, n.Attrs)
			b.WriteString("]")
		}
		b.WriteString(";\n")
	}

	// Edges
	edgeOp := " -> "
	if !g.Directed {
		edgeOp = " -- "
	}
	for _, e := range g.Edges {
		b.WriteString(prefix + "  " + quoteID(e.Src) + edgeOp + quoteID(e.Dst))
		if len(e.Attrs) > 0 {
			b.WriteString(" [")
			writeAttrs(b, e.Attrs)
			b.WriteString("]")
		}
		b.WriteString(";\n")
	}

	b.WriteString(prefix + "}\n")
}

func writeAttrs(b *strings.Builder, attrs map[string]string) {
	first := true
	for k, v := range attrs {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%s=%s", k, quoteID(v))
		first = false
	}
}

func quoteID(s string) string {
	// Quote if contains special characters
	if needsQuoting(s) {
		return "\"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
	}
	return s
}

func needsQuoting(s string) bool {
	for _, c := range s {
		if !isIDChar(c) {
			return true
		}
	}
	return len(s) == 0
}

func isIDChar(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_'
}

// -----------------------------------------------------------------------
// Layout — invoke graphviz for layout computation
// -----------------------------------------------------------------------

// Layout runs the graphviz `dot` command on the graph and returns
// the output in the specified format (e.g., "svg", "png", "json").
func (g *Graph) Layout(format string) ([]byte, error) {
	dot := g.ToDOT()
	cmd := exec.Command("dot", "-T"+format)
	cmd.Stdin = strings.NewReader(dot)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("graphviz dot command failed: %w", err)
	}
	return out, nil
}

// LayoutSVG returns the SVG representation of the graph.
func (g *Graph) LayoutSVG() (string, error) {
	data, err := g.Layout("svg")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// -----------------------------------------------------------------------
// DOT layout parsing helpers (from dot_layout.py)
// -----------------------------------------------------------------------

// BoundingBox represents a rectangle in graph coordinates.
type BoundingBox struct {
	X, Y, W, H float64
}

// Point represents a 2D point.
type Point struct {
	X, Y float64
}

// LayoutNode holds layout information for a node.
type LayoutNode struct {
	Name string
	Pos  Point
	Size Point
}

// LayoutEdge holds layout information for an edge.
type LayoutEdge struct {
	Src    string
	Dst    string
	Points []Point
	Label  string
}

// GraphLayout holds the complete layout result.
type GraphLayout struct {
	BBox  BoundingBox
	Nodes []LayoutNode
	Edges []LayoutEdge
}
