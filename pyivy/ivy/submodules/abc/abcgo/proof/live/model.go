// Package live implements finite-state liveness reductions and recurrence
// analysis on the native AIG graph. True PIs/POs precede latch interfaces.
package live

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"io"
	"os"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Network struct {
	Graph                   *aig.Graph
	InputNames, OutputNames []string
	LTLProperties           []string
}
type Abc_Ntk_t = Network
type Frame struct {
	Network       *Network
	LTLProperties []string
	In            io.Reader
	Out, Err      io.Writer
	Checker       SafetyChecker
}
type Abc_Frame_t = Frame

func (f *Frame) writers() (io.Writer, io.Writer) {
	out, err := f.Out, f.Err
	if out == nil {
		out = os.Stdout
	}
	if err == nil {
		err = os.Stderr
	}
	return out, err
}
func NewNetwork(graph *aig.Graph, inputNames, outputNames []string) (*Network, error) {
	if graph == nil {
		return nil, fmt.Errorf("live: nil graph")
	}
	if err := graph.Check(); err != nil {
		return nil, fmt.Errorf("live: invalid graph: %w", err)
	}
	if len(inputNames) != 0 && len(inputNames) != len(graph.Inputs()) {
		return nil, fmt.Errorf("live: input name count mismatch")
	}
	if len(outputNames) != 0 && len(outputNames) != len(graph.Outputs()) {
		return nil, fmt.Errorf("live: output name count mismatch")
	}
	return &Network{Graph: graph, InputNames: append([]string(nil), inputNames...), OutputNames: append([]string(nil), outputNames...)}, nil
}
func registers(p *aig.Graph) int   { return int(aig.Aig_ManRegNum(p)) }
func trueInputs(p *aig.Graph) int  { return len(p.Inputs()) - registers(p) }
func trueOutputs(p *aig.Graph) int { return len(p.Outputs()) - registers(p) }

// propertyOutputs excludes trailing native constraint-violation outputs.
func propertyOutputs(p *aig.Graph) int { return trueOutputs(p) - int(aig.Aig_ManConstrNum(p)) }
func (c *graphCopy) appendConstraints() {
	for i := propertyOutputs(c.source); i < trueOutputs(c.source); i++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(c.source.Outputs()[i]))
	}
}
func (c *graphCopy) setConstraints() {
	aig.Aig_ManSetConstrNum(c.destination, aig.Aig_ManConstrNum(c.source))
}

func outputDriver(p *aig.Graph, index int) aig.Edge {
	e := p.Outputs()[index]
	nodes := p.Nodes()
	if nodes[e.NodeID()].Kind == aig.Output {
		return nodes[e.NodeID()].Fanins[0] ^ aig.Edge(e&1)
	}
	return e
}
func inputName(n *Network, index int) string {
	if index < len(n.InputNames) && n.InputNames[index] != "" {
		return n.InputNames[index]
	}
	return fmt.Sprintf("pi%d", index)
}
func outputName(n *Network, index int) string {
	if index < len(n.OutputNames) && n.OutputNames[index] != "" {
		return n.OutputNames[index]
	}
	return fmt.Sprintf("po%d", index)
}

type graphCopy struct {
	source, destination *aig.Graph
	mapping             map[int]aig.Edge
	nodes               []aig.Node
}

func newGraphCopy(p *aig.Graph, suffix string) *graphCopy {
	result := aig.Aig_ManStart(int32(len(p.Nodes())))
	result.Name = p.Name + suffix
	aig.Aig_ObjSetCopy(p, 0, 0)
	return &graphCopy{p, result, map[int]aig.Edge{0: 0}, p.Nodes()}
}
func (c *graphCopy) mapEdge(e aig.Edge) aig.Edge { return c.visit(e.NodeID()) ^ aig.Edge(e&1) }
func (c *graphCopy) visit(id int) aig.Edge {
	if e, exists := c.mapping[id]; exists {
		return e
	}
	n := c.nodes[id]
	var result aig.Edge
	switch n.Kind {
	case aig.And:
		result = aig.Aig_And(c.destination, c.mapEdge(n.Fanins[0]), c.mapEdge(n.Fanins[1]))
	case aig.ExclusiveOr:
		result = aig.Aig_Exor(c.destination, c.mapEdge(n.Fanins[0]), c.mapEdge(n.Fanins[1]))
	case aig.Buffer, aig.Output:
		result = c.mapEdge(n.Fanins[0])
	default:
		panic("live: unbound graph input")
	}
	c.mapping[id] = result
	aig.Aig_ObjSetCopy(c.source, aig.Edge(id)<<1, result)
	return result
}
func (c *graphCopy) copyInputs() {
	for _, input := range c.source.Inputs() {
		c.mapping[input.NodeID()] = aig.Aig_ObjCreateCi(c.destination)
		aig.Aig_ObjSetCopy(c.source, input, c.mapping[input.NodeID()])
	}
}
func (c *graphCopy) copyOutputs() {
	for i := range c.source.Outputs() {
		aig.Aig_ObjCreateCo(c.destination, c.mapEdge(outputDriver(c.source, i)))
	}
}
func all(p *aig.Graph, signals []aig.Edge) aig.Edge {
	result := aig.Edge(0)
	for _, e := range signals {
		result = aig.Aig_And(p, result, e)
	}
	return result
}
func disjoin(p *aig.Graph, signals []aig.Edge) aig.Edge {
	result := aig.Edge(1)
	for _, e := range signals {
		result = aig.Aig_Or(p, result, e)
	}
	return result
}
