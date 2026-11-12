package abc

import (
	"fmt"
	aig "github.com/glycerine/abcgo/aig/aig"
	"sync"
)

// Abc_ObjFaninReal skips netlist nets and box terminals for dependency order.
func Abc_ObjFaninReal(o *Object, index int32) *Object {
	child := o.fanins[index].Object
	if Abc_ObjIsBox(o) != 0 {
		child = child.fanins[0].Object
	}
	child = dfsDriver(child)
	if child.Kind == ObjectBO {
		return child.fanins[0].Object
	}
	return child
}
func dfsBoxesVisit(o *Object, nodes *[]*Object, state map[*Object]byte) error {
	if o.Kind == ObjectPI || o.Kind == ObjectConst1 {
		return nil
	}
	if state[o] == 1 {
		return fmt.Errorf("combinational box cycle at object %d", o.ID)
	}
	if state[o] == 2 || dfsVisited(o) {
		return nil
	}
	state[o] = 1
	dfsMark(o)
	for i := range o.fanins {
		if err := dfsBoxesVisit(Abc_ObjFaninReal(o, int32(i)), nodes, state); err != nil {
			return err
		}
	}
	state[o] = 2
	*nodes = append(*nodes, o)
	return nil
}
func Abc_NtkDfsBoxes_rec(o *Object, nodes *[]*Object) {
	if err := dfsBoxesVisit(o, nodes, make(map[*Object]byte)); err != nil {
		panic(err)
	}
}
func Abc_NtkDfsBoxes(n *Network) []*Object {
	if n.Type != uint8(ABC_NTK_NETLIST) {
		panic("box DFS requires netlist")
	}
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	state := make(map[*Object]byte)
	for _, po := range n.ObjectsOfKind(ObjectPO) {
		if len(po.fanins) != 1 {
			return nil
		}
		if err := dfsBoxesVisit(Abc_ObjFaninReal(po, 0), &nodes, state); err != nil {
			return nil
		}
	}
	return nodes
}
func Abc_NtkDeriveFlatGiaSop(graph *aig.Graph, fanins []aig.Edge, text string) aig.Edge {
	return isoCover(graph, text, fanins)
}
func fillHierarchyTemps(n *Network) {
	for _, o := range n.Objects() {
		o.Temp = -1
	}
}
func graphAndCount(graph *aig.Graph) int32 {
	var count int32
	for _, edge := range graph.LiveEdges() {
		if graph.Node(edge).Kind == aig.And {
			count++
		}
	}
	return count
}
func importHierarchyGraph(dst, source *aig.Graph, inputs []aig.Edge) ([]aig.Edge, error) {
	if len(inputs) != len(source.Inputs()) {
		return nil, fmt.Errorf("cached module input count mismatch")
	}
	copies := map[int]aig.Edge{0: 0}
	for i, edge := range source.Inputs() {
		copies[edge.NodeID()] = inputs[i]
	}
	active := map[int]bool{}
	var visit func(aig.Edge) (aig.Edge, error)
	visit = func(edge aig.Edge) (aig.Edge, error) {
		value, ok := copies[edge.NodeID()]
		if !ok {
			if active[edge.NodeID()] {
				return 0, fmt.Errorf("cached module contains cycle")
			}
			active[edge.NodeID()] = true
			node := source.Node(edge)
			a, err := visit(node.Fanins[0])
			if err != nil {
				return 0, err
			}
			switch node.Kind {
			case aig.And, aig.ExclusiveOr:
				b, err := visit(node.Fanins[1])
				if err != nil {
					return 0, err
				}
				if node.Kind == aig.And {
					value = dst.AND(a, b)
				} else {
					value = dst.XOR(a, b)
				}
			case aig.Buffer, aig.Output:
				value = a
			default:
				return 0, fmt.Errorf("invalid cached module node")
			}
			copies[edge.NodeID()] = value
			delete(active, edge.NodeID())
		}
		if edge.Complemented() {
			value = value.Not()
		}
		return value, nil
	}
	outputs := make([]aig.Edge, len(source.Outputs()))
	for i, edge := range source.Outputs() {
		value, err := visit(edge)
		if err != nil {
			return nil, err
		}
		outputs[i] = value
	}
	return outputs, nil
}

type hierarchyAIG struct {
	enterBox func(*aig.Graph, *Object, []aig.Edge) []aig.Edge
	leaveBox func(*aig.Graph, *Object, []aig.Edge) []aig.Edge
	cached   map[*Network]*aig.Graph
	active   map[*Network]bool
	bottomUp bool
	orders   map[*Network][]*Object
}

func (c *hierarchyAIG) order(n *Network) []*Object {
	if order, ok := c.orders[n]; ok {
		return order
	}
	order := Abc_NtkDfsBoxes(n)
	c.orders[n] = order
	return order
}
func hierarchyNodeSignals(driver *Object, value aig.Edge, signals map[*Object]aig.Edge) {
	driver.Temp = int(value)
	for _, net := range driver.fanouts {
		if net.Kind != ObjectNet {
			continue
		}
		signal := value
		if net.fanins[0].Complement {
			signal = signal.Not()
		}
		signals[net] = signal
		net.Temp = int(signal)
	}
}
func (c *hierarchyAIG) expand(graph *aig.Graph, n *Network, inputs []aig.Edge) ([]aig.Edge, error) {
	if n.Type != uint8(ABC_NTK_NETLIST) || len(n.ObjectsOfKind(ObjectLatch)) > 0 {
		return nil, fmt.Errorf("hierarchy AIG conversion requires combinational netlists")
	}
	if n.Function == uint8(ABC_FUNC_BLACKBOX) {
		return nil, fmt.Errorf("blackbox %s has no Boolean implementation", n.Name)
	}
	if c.active[n] {
		return nil, fmt.Errorf("recursive model %s", n.Name)
	}
	c.active[n] = true
	defer delete(c.active, n)
	pis, pos := n.ObjectsOfKind(ObjectPI), n.ObjectsOfKind(ObjectPO)
	if len(pis) != len(inputs) {
		return nil, fmt.Errorf("model input count mismatch")
	}
	fillHierarchyTemps(n)
	signals := make(map[*Object]aig.Edge)
	hierarchyNodeSignals(n.constant, 0, signals)
	for i, pi := range pis {
		hierarchyNodeSignals(pi, inputs[i], signals)
	}
	order := c.order(n)
	if len(order) == 0 && len(pos) > 0 {
		for _, po := range pos {
			net, err := hieInputNet(po)
			if err != nil {
				return nil, err
			}
			if _, ok := signals[net]; !ok {
				return nil, fmt.Errorf("model has an undriven or cyclic output")
			}
		}
	}
	read := func(edge Edge) (aig.Edge, error) {
		value, ok := signals[edge.Object]
		if !ok {
			return 0, fmt.Errorf("net %s is not available in dependency order", hieNetName(edge.Object))
		}
		if edge.Complement {
			value = value.Not()
		}
		return value, nil
	}
	for _, object := range order {
		if object.Kind == ObjectNode {
			args := make([]aig.Edge, len(object.fanins))
			for i, edge := range object.fanins {
				value, err := read(edge)
				if err != nil {
					return nil, err
				}
				args[i] = value
			}
			value := aig.Edge(0)
			if object.Barrier {
				if len(args) != 1 {
					return nil, fmt.Errorf("barrier arity mismatch")
				}
				value = args[0]
			} else {
				value = localFunctionAIG(graph, object, args)
			}
			hierarchyNodeSignals(object, value, signals)
			continue
		}
		model, ok := object.Data.(*Network)
		if !ok || model == nil {
			return nil, fmt.Errorf("box %d has no model", object.ID)
		}
		args := make([]aig.Edge, len(object.fanins))
		for i, edge := range object.fanins {
			terminal := edge.Object
			if len(terminal.fanins) != 1 {
				return nil, fmt.Errorf("box terminal arity mismatch")
			}
			value, err := read(terminal.fanins[0])
			if err != nil {
				return nil, err
			}
			if edge.Complement {
				value = value.Not()
			}
			args[i] = value
		}
		if c.enterBox != nil {
			args = c.enterBox(graph, object, args)
		}
		var results []aig.Edge
		var err error
		if c.bottomUp {
			cached := c.cached[model]
			if cached == nil {
				if provided, ok := model.Data.(*aig.Graph); ok {
					cached = provided
				} else {
					cached, err = c.derive(model)
				}
			}
			if err == nil {
				results, err = importHierarchyGraph(graph, cached, args)
			}
		} else {
			results, err = c.expand(graph, model, args)
		}
		if err != nil {
			return nil, err
		}
		if c.leaveBox != nil {
			results = c.leaveBox(graph, object, results)
		}
		if len(results) != len(object.fanouts) {
			return nil, fmt.Errorf("box output count mismatch")
		}
		for i, bo := range object.fanouts {
			value := results[i]
			if len(bo.fanins) != 1 {
				return nil, fmt.Errorf("box output terminal arity mismatch")
			}
			if bo.fanins[0].Complement {
				value = value.Not()
			}
			hierarchyNodeSignals(bo, value, signals)
		}
	}
	outputs := make([]aig.Edge, len(pos))
	for i, po := range pos {
		if len(po.fanins) != 1 {
			return nil, fmt.Errorf("model output terminal arity mismatch")
		}
		value, err := read(po.fanins[0])
		if err != nil {
			return nil, err
		}
		outputs[i] = value
		po.Temp = int(value)
	}
	return outputs, nil
}
func (c *hierarchyAIG) derive(n *Network) (*aig.Graph, error) {
	if graph := c.cached[n]; graph != nil {
		return graph, nil
	}
	graph := aig.NewGraph()
	graph.Name, graph.Spec = n.Name, n.Spec
	inputs := make([]aig.Edge, len(n.ObjectsOfKind(ObjectPI)))
	for i := range inputs {
		inputs[i] = graph.NewInput()
	}
	outputs, err := c.expand(graph, n, inputs)
	if err != nil {
		return nil, err
	}
	for _, output := range outputs {
		graph.AddOutput(output)
	}
	aig.Aig_ManCleanup(graph)
	if err := graph.Check(); err != nil {
		return nil, err
	}
	c.cached[n] = graph
	return graph, nil
}
func newHierarchyAIG(bottomUp bool) *hierarchyAIG {
	return &hierarchyAIG{cached: make(map[*Network]*aig.Graph), active: make(map[*Network]bool), orders: make(map[*Network][]*Object), bottomUp: bottomUp}
}
func Abc_NtkDeriveFlatGia_rec(graph *aig.Graph, n *Network) {
	inputs := make([]aig.Edge, len(n.ObjectsOfKind(ObjectPI)))
	for i, pi := range n.ObjectsOfKind(ObjectPI) {
		net, err := hieOutputNet(pi)
		if err != nil || net.Temp < 0 {
			panic("hierarchy input has no assigned literal")
		}
		value := aig.Edge(net.Temp)
		if net.fanins[0].Complement {
			value = value.Not()
		}
		inputs[i] = value
	}
	if _, err := newHierarchyAIG(false).expand(graph, n, inputs); err != nil {
		panic(err)
	}
}
func Abc_NtkDeriveFlatGia(n *Network) *aig.Graph {
	if !hieValidateModels(n) || Abc_NtkCheckRecursive(n) != 0 {
		return nil
	}
	graph, err := newHierarchyAIG(false).derive(n)
	if err != nil {
		return nil
	}
	return graph
}
func Abc_NtkCountAndNodes(order []*Object) int32 {
	var count int32
	for _, object := range order {
		if object.Kind == ObjectNode {
			count++
			continue
		}
		model, ok := object.Data.(*Network)
		if !ok || model == nil {
			continue
		}
		graph, ok := model.Data.(*aig.Graph)
		if !ok {
			graph = Abc_NtkDeriveFlatGia(model)
		}
		if graph != nil {
			count += graphAndCount(graph)
		}
	}
	return count
}
func Abc_NtkDeriveFlatGia2Derive(n *Network, order []*Object) *aig.Graph {
	context := newHierarchyAIG(true)
	context.orders[n] = order
	graph, err := context.derive(n)
	if err != nil {
		return nil
	}
	return graph
}
func Abc_NtkDeriveFlatGia2(n *Network, models []*Network) *aig.Graph {
	if !hieValidateModels(n) || Abc_NtkCheckRecursive(n) != 0 {
		return nil
	}
	context := newHierarchyAIG(true)
	for _, model := range models {
		if _, err := context.derive(model); err != nil {
			return nil
		}
	}
	graph, err := context.derive(n)
	if err != nil {
		return nil
	}
	return graph
}
func collectHierarchy(n *Network, models *[]*Network, state map[*Network]byte) error {
	if state[n] == 2 {
		return nil
	}
	if state[n] == 1 {
		return fmt.Errorf("recursive model %s", n.Name)
	}
	state[n] = 1
	for _, object := range Abc_NtkDfsBoxes(n) {
		if Abc_ObjIsBox(object) == 0 {
			continue
		}
		model, ok := object.Data.(*Network)
		if ok && model != nil && model != n {
			if err := collectHierarchy(model, models, state); err != nil {
				return err
			}
		}
	}
	state[n] = 2
	n.Step = int32(len(*models))
	*models = append(*models, n)
	return nil
}
func Abc_NtkCollectHie_rec(n *Network, models *[]*Network) {
	state := make(map[*Network]byte)
	for _, model := range *models {
		state[model] = 2
	}
	if err := collectHierarchy(n, models, state); err != nil {
		panic(err)
	}
}
func Abc_NtkCollectHie(n *Network) []*Network {
	var models []*Network
	if Abc_NtkCheckRecursive(n) != 0 {
		return nil
	}
	if err := collectHierarchy(n, &models, make(map[*Network]byte)); err != nil {
		return nil
	}
	return models
}
func countHierarchy(n *Network, instances bool, memo map[*Network]float64, active map[*Network]bool) float64 {
	if value, ok := memo[n]; ok {
		return value
	}
	if active[n] {
		panic("recursive hierarchy cannot be counted")
	}
	active[n] = true
	value := float64(0)
	if instances {
		value = 1
	}
	if n.Function != uint8(ABC_FUNC_BLACKBOX) {
		for _, object := range Abc_NtkDfsBoxes(n) {
			if object.Kind == ObjectNode {
				if !instances {
					value++
				}
				continue
			}
			model, ok := object.Data.(*Network)
			if ok && model != nil && model != n {
				value += countHierarchy(model, instances, memo, active)
			}
		}
	}
	delete(active, n)
	memo[n] = value
	if instances {
		n.Step = int32(value)
	} else {
		n.DTemp = value
	}
	return value
}
func Abc_NtkCountInst_rec(n *Network) int32 {
	return int32(countHierarchy(n, true, make(map[*Network]float64), make(map[*Network]bool)))
}
func Abc_NtkCountInst(n *Network) { fmt.Printf("Instances = %10d.\n", Abc_NtkCountInst_rec(n)) }
func Abc_NtkCountNodes_rec(n *Network) float64 {
	return countHierarchy(n, false, make(map[*Network]float64), make(map[*Network]bool))
}
func Abc_NtkCountNodes(n *Network) { fmt.Printf("Nodes = %.0f\n", Abc_NtkCountNodes_rec(n)) }
func Abc_NtkCheckRecursive(n *Network) int32 {
	state := make(map[*Network]byte)
	var visit func(*Network) bool
	visit = func(model *Network) bool {
		if state[model] == 1 {
			fmt.Printf("WARNING: Model %q contains a recursive definition.\n", model.Name)
			return true
		}
		if state[model] == 2 {
			return false
		}
		state[model] = 1
		for _, box := range model.boxes {
			if box.Kind == ObjectLatch {
				continue
			}
			if child, ok := box.Data.(*Network); ok && child != nil {
				if visit(child) {
					return true
				}
			}
		}
		state[model] = 2
		return false
	}
	if visit(n) {
		return 1
	}
	if n.Design != nil {
		for _, model := range n.Design.Modules {
			if visit(model) {
				return 1
			}
		}
	}
	return 0
}

// HierarchyReader is installed by the actual BLIF I/O package to avoid an
// abc↔io dependency cycle. The booleans select multivalued parsing and checking.
type HierarchyReader func(filename string, multiValued, check bool) (*Network, error)

var hierarchyReader struct {
	sync.RWMutex
	read HierarchyReader
}

func SetHierarchyReader(read HierarchyReader) {
	hierarchyReader.Lock()
	hierarchyReader.read = read
	hierarchyReader.Unlock()
}
func ReadHierarchyAIG(filename string, verbose bool) (*aig.Graph, error) {
	hierarchyReader.RLock()
	read := hierarchyReader.read
	hierarchyReader.RUnlock()
	if read == nil {
		return nil, fmt.Errorf("hierarchical BLIF reader service is not installed")
	}
	n, err := read(filename, false, true)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, fmt.Errorf("hierarchical BLIF reader returned no network")
	}
	if verbose {
		Abc_NtkPrintBoxInfo(n)
	}
	if n.Type != uint8(ABC_NTK_NETLIST) {
		return nil, fmt.Errorf("hierarchical reader must return a netlist")
	}
	models := Abc_NtkCollectHie(n)
	if models == nil {
		return nil, fmt.Errorf("hierarchical model dependency cycle")
	}
	graph := Abc_NtkDeriveFlatGia2(n, models)
	if graph == nil {
		return nil, fmt.Errorf("hierarchical Boolean graph derivation failed")
	}
	return graph, nil
}
func Abc_NtkHieCecTest(filename string, verbose int32) *aig.Graph {
	graph, err := ReadHierarchyAIG(filename, verbose != 0)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return graph
}
