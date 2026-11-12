package abc

import (
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// WriteNodeBDDVisualization writes a local node's BDD or its equivalent ADD.
func WriteNodeBDDVisualization(w io.Writer, node *Object, complementedEdges bool) error {
	if node == nil || node.network == nil || Abc_NtkIsBddLogic(node.network) == 0 {
		return fmt.Errorf("abc: BDD visualization requires a BDD logic node")
	}
	dd, ok := node.network.FunctionManager.(*core.DdManager)
	if !ok || dd == nil {
		return fmt.Errorf("abc: BDD manager is missing")
	}
	root, ok := node.Data.(*core.DdNode)
	if !ok || root == nil {
		return fmt.Errorf("abc: node BDD is missing")
	}
	if !complementedEdges {
		root = core.Cudd_BddToAdd(dd, root)
		if root == nil {
			return fmt.Errorf("abc: BDD to ADD conversion failed")
		}
	}
	names := make([]string, len(node.fanins))
	for i, edge := range node.fanins {
		names[i] = Abc_ObjName(edge.Object)
	}
	return core.WriteDOT(w, dd, []*core.DdNode{root}, names, []string{Abc_ObjName(node)})
}

// WriteNetworkBDDVisualization constructs global BDDs for a strashed network.
func WriteNetworkBDDVisualization(w io.Writer, n *Network, complementedEdges bool) error {
	if n == nil || Abc_NtkIsStrash(n) == 0 {
		return fmt.Errorf("abc: global BDD visualization requires an AIG")
	}
	dd := core.NewBDDManager(len(n.cis))
	defer core.Cudd_Quit(dd)
	core.Cudd_SetMaxLive(dd, 10000000)
	values := make(map[*Object]*core.DdNode, len(n.objects))
	values[n.constant] = dd.One
	inputNames := make([]string, len(n.cis))
	for i, input := range n.cis {
		values[input] = dd.Vars[i]
		inputNames[i] = Abc_ObjName(input)
	}
	active := make(map[*Object]bool)
	var visit func(*Object) (*core.DdNode, error)
	visit = func(node *Object) (*core.DdNode, error) {
		if value := values[node]; value != nil {
			return value, nil
		}
		if node == nil || node.Kind != ObjectNode || len(node.fanins) != 2 || active[node] {
			return nil, fmt.Errorf("abc: invalid AIG cone")
		}
		active[node] = true
		var inputs [2]*core.DdNode
		for i, edge := range node.fanins {
			value, err := visit(edge.Object)
			if err != nil {
				return nil, err
			}
			inputs[i] = core.Cudd_NotCond(value, edge.Complement)
		}
		delete(active, node)
		value := core.Cudd_bddAnd(dd, inputs[0], inputs[1])
		if value == nil {
			return nil, fmt.Errorf("abc: construction of global BDDs failed")
		}
		values[node] = value
		return value, nil
	}
	roots := make([]*core.DdNode, len(n.cos))
	outputNames := make([]string, len(n.cos))
	for i, output := range n.cos {
		if len(output.fanins) != 1 {
			return fmt.Errorf("abc: output has no driver")
		}
		edge := output.fanins[0]
		value, err := visit(edge.Object)
		if err != nil {
			return err
		}
		value = core.Cudd_NotCond(value, edge.Complement)
		if !complementedEdges {
			value = core.Cudd_BddToAdd(dd, value)
			if value == nil {
				return fmt.Errorf("abc: BDD to ADD conversion failed")
			}
		}
		roots[i] = value
		outputNames[i] = Abc_ObjName(output)
	}
	return core.WriteDOT(w, dd, roots, inputNames, outputNames)
}

func showBDDFile(filename string, write func(io.Writer) error) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cannot open the intermediate file %q.\n", filename)
		return
	}
	if err = write(file); err == nil {
		err = file.Close()
	} else {
		file.Close()
	}
	if err != nil {
		fmt.Printf("Cannot write intermediate file %q: %v\n", filename, err)
		return
	}
	Abc_ShowFile(filename)
}

func Abc_NodeShowBddOne(dd *core.DdManager, root *core.DdNode) {
	showBDDFile("temp.dot", func(w io.Writer) error { return core.WriteDOT(w, dd, []*core.DdNode{root}, nil, nil) })
}

func Abc_NodeShowBdd(node *Object, complementedEdges int32) {
	showBDDFile(Abc_ShowGetFileName(Abc_ObjName(node)), func(w io.Writer) error {
		return WriteNodeBDDVisualization(w, node, complementedEdges != 0)
	})
}

func Abc_NtkShowBdd(n *Network, complementedEdges int32) {
	showBDDFile(Abc_ShowGetFileName(n.Name), func(w io.Writer) error {
		return WriteNetworkBDDVisualization(w, n, complementedEdges != 0)
	})
}
