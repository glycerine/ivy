package abc

// Derived from ABC src/base/abc/{abcFunc.c,abcAig.c,abcNetlist.c,abcMinBase.c}.

import (
	"fmt"
	"slices"
)

// ABCFunctionalOptions bounds stand-alone representation conversion work.
type ABCFunctionalOptions struct {
	MaxOperations     int
	MaxTruthVariables int
	Stop              func() bool
}

func (o ABCFunctionalOptions) normalized() ABCFunctionalOptions {
	if o.MaxOperations <= 0 {
		o.MaxOperations = 1 << 22
	}
	if o.MaxTruthVariables <= 0 {
		o.MaxTruthVariables = 16
	}
	return o
}

type abcFunctionalBudget struct {
	options ABCFunctionalOptions
	used    int
}

func newABCFunctionalBudget(options ABCFunctionalOptions) (*abcFunctionalBudget, error) {
	if options.MaxOperations < 0 || options.MaxTruthVariables < 0 {
		return nil, ErrDimension
	}
	return &abcFunctionalBudget{options: options.normalized()}, nil
}

func (b *abcFunctionalBudget) tick(count int) error {
	if count < 0 || b.used > b.options.MaxOperations-count {
		return ErrLimit
	}
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	b.used += count
	return nil
}

// ABCToAIGResult owns the converted graph and object-to-literal mapping.
type ABCToAIGResult struct {
	Graph            *AigManager
	ObjectToLiteral  []AigLit
	CombinationalIDs []int
}

func abcApplyEdge(literal AigLit, edge ABCObjectEdge) AigLit {
	return literal.NotIf(edge.Complemented)
}

func abcBuildSOPAIG(manager *AigManager, sop ABCSOP, inputs []AigLit, budget *abcFunctionalBudget) (AigLit, error) {
	if sop.IsConst0() {
		return AigConst0, nil
	}
	if sop.IsConst1() {
		return AigConst1, nil
	}
	if len(inputs) != sop.Variables {
		return InvalidAigLit, ErrDimension
	}
	if err := sop.Check(ABCMetadataOptions{MaxVariables: budget.options.MaxTruthVariables}); err != nil {
		return InvalidAigLit, err
	}
	if sop.IsXORType() {
		result := AigConst0
		for _, input := range inputs {
			if err := budget.tick(1); err != nil {
				return InvalidAigLit, err
			}
			var err error
			result, err = manager.Xor(result, input)
			if err != nil {
				return InvalidAigLit, err
			}
		}
		return result.NotIf(sop.Phase == 'n'), nil
	}
	result := AigConst0
	for _, cube := range sop.Cubes {
		product := AigConst1
		for variable, value := range []byte(cube) {
			if value == '-' {
				continue
			}
			if err := budget.tick(1); err != nil {
				return InvalidAigLit, err
			}
			literal := inputs[variable].NotIf(value == '0')
			var err error
			product, err = manager.And(product, literal)
			if err != nil {
				return InvalidAigLit, err
			}
		}
		if err := budget.tick(1); err != nil {
			return InvalidAigLit, err
		}
		var err error
		result, err = manager.Or(result, product)
		if err != nil {
			return InvalidAigLit, err
		}
	}
	return result.NotIf(sop.Phase == '0'), nil
}

func abcLatchTerminals(network *ABCNetwork) (inputs, outputs []int, err error) {
	for _, boxID := range network.boxes {
		box := &network.objects[boxID]
		if !box.Alive || box.Type != ABCObjectLatch {
			return nil, nil, ErrDimension
		}
		if len(box.Fanins) != 1 || len(box.Fanouts) != 1 {
			return nil, nil, ErrInvariant
		}
		bi, bo := box.Fanins[0].ID, box.Fanouts[0]
		if bi < 0 || bi >= len(network.objects) || bo < 0 || bo >= len(network.objects) || network.objects[bi].Type != ABCObjectBI || network.objects[bo].Type != ABCObjectBO {
			return nil, nil, ErrInvariant
		}
		inputs = append(inputs, bi)
		outputs = append(outputs, bo)
	}
	return inputs, outputs, nil
}

func abcFunctionalPostorder(network *ABCNetwork, roots []int, budget *abcFunctionalBudget) ([]int, error) {
	state := make([]uint8, len(network.objects))
	order := make([]int, 0)
	type frame struct{ id, next int }
	for _, root := range roots {
		if root < 0 || root >= len(network.objects) || !network.objects[root].Alive {
			return nil, ErrDimension
		}
		if state[root] == 2 {
			continue
		}
		stack := []frame{{id: root}}
		for len(stack) != 0 {
			top := &stack[len(stack)-1]
			object := &network.objects[top.id]
			if state[top.id] == 0 {
				if err := budget.tick(1); err != nil {
					return nil, err
				}
				state[top.id] = 1
			}
			if abcIsCI(object.Type) || object.Type == ABCObjectConst1 || object.Type == ABCObjectLatch {
				state[top.id] = 2
				stack = stack[:len(stack)-1]
				continue
			}
			if top.next < len(object.Fanins) {
				fanin := object.Fanins[top.next].ID
				top.next++
				if fanin < 0 || fanin >= len(network.objects) || !network.objects[fanin].Alive {
					return nil, ErrInvariant
				}
				if state[fanin] == 1 {
					return nil, fmt.Errorf("functional conversion cycle: %w", ErrInvariant)
				}
				if state[fanin] == 0 {
					stack = append(stack, frame{id: fanin})
				}
				continue
			}
			state[top.id] = 2
			if object.Type == ABCObjectNode || object.Type == ABCObjectNet {
				order = append(order, top.id)
			}
			stack = stack[:len(stack)-1]
		}
	}
	return order, nil
}

// ABCNetworkToAIG implements the concrete SOP/AIG portion of Abc_NtkSopToAig
// and Abc_NtkToAig. Net objects are treated as transparent signals.
func ABCNetworkToAIG(network *ABCNetwork, options ABCFunctionalOptions) (ABCToAIGResult, error) {
	if network == nil || (network.Function != ABCFunctionSOP && network.Function != ABCFunctionAIG) {
		return ABCToAIGResult{}, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return ABCToAIGResult{}, err
	}
	budget, err := newABCFunctionalBudget(options)
	if err != nil {
		return ABCToAIGResult{}, err
	}
	latchInputs, latchOutputs, err := abcLatchTerminals(network)
	if err != nil {
		return ABCToAIGResult{}, err
	}
	manager := NewAigManager()
	mapping := make([]AigLit, len(network.objects))
	for i := range mapping {
		mapping[i] = InvalidAigLit
	}
	if len(network.objects) > 0 && network.objects[0].Alive && network.objects[0].Type == ABCObjectConst1 {
		mapping[0] = AigConst1
	}
	for _, id := range append(slices.Clone(network.pis), latchOutputs...) {
		if err := budget.tick(1); err != nil {
			return ABCToAIGResult{}, err
		}
		literal, createErr := manager.CreateCI()
		if createErr != nil {
			return ABCToAIGResult{}, createErr
		}
		mapping[id] = literal
	}
	roots := make([]int, 0, len(network.pos)+len(latchInputs))
	for _, id := range append(slices.Clone(network.pos), latchInputs...) {
		object := &network.objects[id]
		if len(object.Fanins) != 1 {
			return ABCToAIGResult{}, ErrInvariant
		}
		roots = append(roots, object.Fanins[0].ID)
	}
	order, err := abcFunctionalPostorder(network, roots, budget)
	if err != nil {
		return ABCToAIGResult{}, err
	}
	for _, id := range order {
		object := &network.objects[id]
		if object.Type == ABCObjectNet {
			if len(object.Fanins) != 1 || mapping[object.Fanins[0].ID] == InvalidAigLit {
				return ABCToAIGResult{}, ErrInvariant
			}
			mapping[id] = abcApplyEdge(mapping[object.Fanins[0].ID], object.Fanins[0])
			continue
		}
		inputs := make([]AigLit, len(object.Fanins))
		for i, edge := range object.Fanins {
			if mapping[edge.ID] == InvalidAigLit {
				return ABCToAIGResult{}, ErrInvariant
			}
			inputs[i] = abcApplyEdge(mapping[edge.ID], edge)
		}
		sop, parseErr := ParseABCSOP(object.Function, ABCMetadataOptions{MaxVariables: budget.options.MaxTruthVariables})
		if parseErr != nil {
			return ABCToAIGResult{}, parseErr
		}
		mapping[id], err = abcBuildSOPAIG(manager, sop, inputs, budget)
		if err != nil {
			return ABCToAIGResult{}, err
		}
	}
	for _, id := range append(slices.Clone(network.pos), latchInputs...) {
		edge := network.objects[id].Fanins[0]
		if mapping[edge.ID] == InvalidAigLit {
			return ABCToAIGResult{}, ErrInvariant
		}
		if _, err = manager.CreateCO(abcApplyEdge(mapping[edge.ID], edge)); err != nil {
			return ABCToAIGResult{}, err
		}
	}
	if err = manager.SetRegisterCount(len(latchInputs)); err != nil {
		return ABCToAIGResult{}, err
	}
	if network.Constraints > manager.PrimaryOutputCount() {
		return ABCToAIGResult{}, ErrDimension
	}
	if err = manager.SetConstraintCount(network.Constraints); err != nil {
		return ABCToAIGResult{}, err
	}
	if err = manager.Check(); err != nil {
		return ABCToAIGResult{}, err
	}
	return ABCToAIGResult{Graph: manager, ObjectToLiteral: mapping, CombinationalIDs: order}, nil
}

// AIGToABCNetwork implements the concrete strashed-network duplication path.
func AIGToABCNetwork(graph *AigManager, options ABCFunctionalOptions) (*ABCNetwork, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	budget, err := newABCFunctionalBudget(options)
	if err != nil {
		return nil, err
	}
	limits := ABCNetworkLimits{MaxObjects: max(32, graph.ObjectCount()*3+8), MaxEdges: max(64, graph.ObjectCount()*4+8), Stop: options.Stop}
	network, err := NewABCNetwork(ABCNetworkStrash, ABCFunctionAIG, limits)
	if err != nil {
		return nil, err
	}
	objectMap := make([]int, graph.ObjectCount())
	for i := range objectMap {
		objectMap[i] = -1
	}
	objectMap[0] = 0
	for i := 0; i < graph.PrimaryInputCount(); i++ {
		id, createErr := network.CreatePI()
		if createErr != nil {
			return nil, createErr
		}
		objectMap[graph.cis[i]] = id
	}
	poIDs := make([]int, graph.PrimaryOutputCount())
	for i := range poIDs {
		poIDs[i], err = network.CreatePO()
		if err != nil {
			return nil, err
		}
	}
	biIDs := make([]int, graph.RegisterCount())
	for i := 0; i < graph.RegisterCount(); i++ {
		bi, _, bo, createErr := network.CreateLatchBox(ABCInitZero)
		if createErr != nil {
			return nil, createErr
		}
		biIDs[i] = bi
		objectMap[graph.cis[graph.PrimaryInputCount()+i]] = bo
	}
	for id, object := range graph.objects {
		if err := budget.tick(1); err != nil {
			return nil, err
		}
		if object.Type != AigAND && object.Type != AigBuffer {
			continue
		}
		makeEdge := func(literal AigLit) (ABCObjectEdge, error) {
			mapped := objectMap[literal.ID()]
			if mapped < 0 {
				return ABCObjectEdge{}, ErrInvariant
			}
			return ABCObjectEdge{ID: mapped, Complemented: literal.Complemented()}, nil
		}
		first, edgeErr := makeEdge(object.Fanin0)
		if edgeErr != nil {
			return nil, edgeErr
		}
		var node int
		if object.Type == AigBuffer {
			node, err = network.CreateLogicNode("1 1\n", first)
		} else {
			second, secondErr := makeEdge(object.Fanin1)
			if secondErr != nil {
				return nil, secondErr
			}
			node, err = network.CreateLogicNode("11 1\n", first, second)
		}
		if err != nil {
			return nil, err
		}
		objectMap[id] = node
	}
	for index, coID := range graph.cos {
		object := graph.objects[coID]
		mapped := objectMap[object.Fanin0.ID()]
		if mapped < 0 {
			return nil, ErrInvariant
		}
		var target int
		if index < graph.PrimaryOutputCount() {
			target = poIDs[index]
		} else {
			target = biIDs[index-graph.PrimaryOutputCount()]
		}
		if err = network.AddFanin(target, ABCObjectEdge{ID: mapped, Complemented: object.Fanin0.Complemented()}); err != nil {
			return nil, err
		}
	}
	network.Constraints = graph.ConstraintCount()
	if err = network.Validate(); err != nil {
		return nil, err
	}
	return network, nil
}

func abcSOPProjection(sop ABCSOP, keep []int) (ABCSOP, error) {
	index := make([]int, sop.Variables)
	for i := range index {
		index[i] = -1
	}
	for newIndex, oldIndex := range keep {
		if oldIndex < 0 || oldIndex >= sop.Variables || index[oldIndex] >= 0 {
			return ABCSOP{}, ErrDimension
		}
		index[oldIndex] = newIndex
	}
	result := ABCSOP{Variables: len(keep), Cubes: make([]string, len(sop.Cubes)), Phase: sop.Phase}
	for cubeIndex, cube := range sop.Cubes {
		projected := make([]byte, len(keep))
		for oldIndex, newIndex := range index {
			if newIndex >= 0 {
				projected[newIndex] = cube[oldIndex]
			}
		}
		result.Cubes[cubeIndex] = string(projected)
	}
	return result, nil
}

// ABCMinimizeNodeBase removes unused and duplicate SOP inputs while preserving
// the function under the source fanin phase interpretation.
func ABCMinimizeNodeBase(network *ABCNetwork, nodeID int, options ABCFunctionalOptions) (bool, error) {
	budget, err := newABCFunctionalBudget(options)
	if err != nil {
		return false, err
	}
	return abcMinimizeNodeBase(network, nodeID, budget)
}

func abcMinimizeNodeBase(network *ABCNetwork, nodeID int, budget *abcFunctionalBudget) (bool, error) {
	if network == nil || network.Function != ABCFunctionSOP {
		return false, ErrDimension
	}
	object, err := network.live(nodeID)
	if err != nil || object.Type != ABCObjectNode {
		return false, ErrDimension
	}
	sop, err := ParseABCSOP(object.Function, ABCMetadataOptions{MaxVariables: budget.options.MaxTruthVariables})
	if err != nil || sop.Variables != len(object.Fanins) {
		return false, ErrDimension
	}
	used := make([]bool, sop.Variables)
	if sop.IsXORType() {
		for i := range used {
			used[i] = true
		}
	} else {
		for _, cube := range sop.Cubes {
			for i := range cube {
				used[i] = used[i] || cube[i] != '-'
			}
		}
	}
	keep := make([]int, 0, len(used))
	for i, yes := range used {
		if yes {
			keep = append(keep, i)
		}
	}
	projected, err := abcSOPProjection(sop, keep)
	if err != nil {
		return false, err
	}
	edges := make([]ABCObjectEdge, len(keep))
	for i, oldIndex := range keep {
		edges[i] = object.Fanins[oldIndex]
	}
	uniqueIDs := make([]int, 0, len(edges))
	uniqueIndex := make(map[int]int)
	for _, edge := range edges {
		if _, ok := uniqueIndex[edge.ID]; !ok {
			uniqueIndex[edge.ID] = len(uniqueIDs)
			uniqueIDs = append(uniqueIDs, edge.ID)
		}
	}
	if len(uniqueIDs) < len(edges) {
		if len(uniqueIDs) > budget.options.MaxTruthVariables || len(uniqueIDs) >= 63 {
			return false, ErrLimit
		}
		assignmentCount := uint64(1) << uint(len(uniqueIDs))
		remaining := budget.options.MaxOperations - budget.used
		if remaining < 0 || assignmentCount > uint64(remaining) {
			return false, ErrLimit
		}
		words := make([]uint64, int((assignmentCount+63)/64))
		for assignment := uint64(0); assignment < assignmentCount; assignment++ {
			if err := budget.tick(1); err != nil {
				return false, err
			}
			oldAssignment := uint64(0)
			for variable, edge := range edges {
				value := assignment>>uint(uniqueIndex[edge.ID])&1 != 0
				if edge.Complemented {
					value = !value
				}
				if value {
					oldAssignment |= uint64(1) << uint(variable)
				}
			}
			value, evalErr := projected.Evaluate(oldAssignment)
			if evalErr != nil {
				return false, evalErr
			}
			if value {
				words[assignment>>6] |= uint64(1) << (assignment & 63)
			}
		}
		projected, err = ABCSOPFromTruth(len(uniqueIDs), words, ABCMetadataOptions{MaxVariables: budget.options.MaxTruthVariables})
		if err != nil {
			return false, err
		}
		edges = make([]ABCObjectEdge, len(uniqueIDs))
		for i, id := range uniqueIDs {
			edges[i] = ABCObjectEdge{ID: id}
		}
	}
	text, err := projected.Format(ABCMetadataOptions{MaxVariables: budget.options.MaxTruthVariables})
	if err != nil {
		return false, err
	}
	if slices.Equal(edges, object.Fanins) && text == object.Function {
		return false, nil
	}
	snapshot := network.Clone()
	if err = network.RemoveFanins(nodeID); err == nil {
		for _, edge := range edges {
			if err = budget.tick(1); err != nil {
				break
			}
			if err = network.AddFanin(nodeID, edge); err != nil {
				break
			}
		}
	}
	if err == nil {
		err = network.SetNodeFunction(nodeID, text)
	}
	if err == nil {
		err = network.Validate()
	}
	if err != nil {
		network.restore(snapshot)
		return false, err
	}
	return true, nil
}

// ABCMinimumBase applies ABCMinimizeNodeBase to all live logic nodes atomically.
func ABCMinimumBase(network *ABCNetwork, options ABCFunctionalOptions) (int, error) {
	if network == nil {
		return 0, ErrDimension
	}
	budget, err := newABCFunctionalBudget(options)
	if err != nil {
		return 0, err
	}
	snapshot := network.Clone()
	changed := 0
	for id, object := range network.objects {
		if !object.Alive || object.Type != ABCObjectNode {
			continue
		}
		yes, err := abcMinimizeNodeBase(network, id, budget)
		if err != nil {
			network.restore(snapshot)
			return 0, err
		}
		if yes {
			changed++
		}
	}
	return changed, nil
}

func abcCopyMetadata(from, to *ABCNetwork) {
	to.Name, to.Spec, to.Step, to.Constraints, to.AndGateDelay = from.Name, from.Spec, from.Step, from.Constraints, from.AndGateDelay
}

func abcFunctionalDrivenNetName(network *ABCNetwork, driverID int) (string, error) {
	if driverID < 0 || driverID >= len(network.objects) || !network.objects[driverID].Alive {
		return "", ErrDimension
	}
	name := ""
	count := 0
	for _, fanoutID := range network.objects[driverID].Fanouts {
		if fanoutID < 0 || fanoutID >= len(network.objects) {
			return "", ErrInvariant
		}
		fanout := &network.objects[fanoutID]
		if !fanout.Alive || fanout.Type != ABCObjectNet || len(fanout.Fanins) != 1 || fanout.Fanins[0].ID != driverID {
			continue
		}
		count++
		name = fanout.Name
	}
	if count != 1 {
		return "", ErrInvariant
	}
	return name, nil
}

// ABCLogicToNetlist converts a combinational logic network to explicit nets.
func ABCLogicToNetlist(network *ABCNetwork, options ABCFunctionalOptions) (*ABCNetwork, error) {
	if network == nil || network.Type != ABCNetworkLogic || len(network.boxes) != 0 {
		return nil, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return nil, err
	}
	budget, err := newABCFunctionalBudget(options)
	if err != nil {
		return nil, err
	}
	result, err := NewABCNetwork(ABCNetworkNetlist, network.Function, network.limits)
	if err != nil {
		return nil, err
	}
	abcCopyMetadata(network, result)
	mapping := make([]int, len(network.objects))
	for i := range mapping {
		mapping[i] = -1
	}
	for _, oldID := range network.pis {
		pi, createErr := result.CreatePI()
		if createErr != nil {
			return nil, createErr
		}
		net, createErr := result.CreateNet()
		if createErr == nil {
			createErr = result.AddFanin(net, ABCObjectEdge{ID: pi})
		}
		if createErr != nil {
			return nil, createErr
		}
		mapping[oldID] = net
		if network.objects[oldID].Name != "" {
			if err = result.SetObjectName(net, network.objects[oldID].Name); err != nil {
				return nil, err
			}
		}
	}
	order, err := abcFunctionalPostorder(network, network.pos, budget)
	if err != nil {
		return nil, err
	}
	for _, oldID := range order {
		old := &network.objects[oldID]
		if old.Type != ABCObjectNode {
			continue
		}
		node, createErr := result.CreateNode()
		if createErr != nil {
			return nil, createErr
		}
		for _, edge := range old.Fanins {
			if mapping[edge.ID] < 0 {
				return nil, ErrInvariant
			}
			if createErr = result.AddFanin(node, ABCObjectEdge{ID: mapping[edge.ID], Complemented: edge.Complemented}); createErr != nil {
				return nil, createErr
			}
		}
		if createErr = result.SetNodeFunction(node, old.Function); createErr != nil {
			return nil, createErr
		}
		net, createErr := result.CreateNet()
		if createErr == nil {
			createErr = result.AddFanin(net, ABCObjectEdge{ID: node})
		}
		if createErr != nil {
			return nil, createErr
		}
		mapping[oldID] = net
		if old.Name != "" {
			if err = result.SetObjectName(net, old.Name); err != nil {
				return nil, err
			}
		}
	}
	for _, oldID := range network.pos {
		old := &network.objects[oldID]
		if len(old.Fanins) != 1 || mapping[old.Fanins[0].ID] < 0 {
			return nil, ErrInvariant
		}
		po, createErr := result.CreatePO()
		if createErr == nil {
			createErr = result.AddFanin(po, ABCObjectEdge{ID: mapping[old.Fanins[0].ID], Complemented: old.Fanins[0].Complemented})
		}
		if createErr != nil {
			return nil, createErr
		}
		if old.Name != "" {
			if err = result.SetObjectName(po, old.Name); err != nil {
				return nil, err
			}
		}
	}
	if err = result.Validate(); err != nil {
		return nil, err
	}
	return result, nil
}

// ABCNetlistToLogic removes transparent nets from a combinational netlist.
func ABCNetlistToLogic(network *ABCNetwork, options ABCFunctionalOptions) (*ABCNetwork, error) {
	if network == nil || network.Type != ABCNetworkNetlist || len(network.boxes) != 0 {
		return nil, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return nil, err
	}
	budget, err := newABCFunctionalBudget(options)
	if err != nil {
		return nil, err
	}
	result, err := NewABCNetwork(ABCNetworkLogic, network.Function, network.limits)
	if err != nil {
		return nil, err
	}
	abcCopyMetadata(network, result)
	mapping := make([]ABCObjectEdge, len(network.objects))
	valid := make([]bool, len(network.objects))
	for _, oldID := range network.pis {
		pi, createErr := result.CreatePI()
		if createErr != nil {
			return nil, createErr
		}
		name, nameErr := abcFunctionalDrivenNetName(network, oldID)
		if nameErr != nil {
			return nil, nameErr
		}
		if name != "" {
			if createErr = result.SetObjectName(pi, name); createErr != nil {
				return nil, createErr
			}
		}
		mapping[oldID], valid[oldID] = ABCObjectEdge{ID: pi}, true
	}
	order, err := abcFunctionalPostorder(network, network.pos, budget)
	if err != nil {
		return nil, err
	}
	for _, oldID := range order {
		old := &network.objects[oldID]
		if old.Type == ABCObjectNet {
			if len(old.Fanins) != 1 || !valid[old.Fanins[0].ID] {
				return nil, ErrInvariant
			}
			edge := mapping[old.Fanins[0].ID]
			edge.Complemented = edge.Complemented != old.Fanins[0].Complemented
			mapping[oldID], valid[oldID] = edge, true
			continue
		}
		if old.Type != ABCObjectNode {
			continue
		}
		node, createErr := result.CreateNode()
		if createErr != nil {
			return nil, createErr
		}
		for _, oldEdge := range old.Fanins {
			if !valid[oldEdge.ID] {
				return nil, ErrInvariant
			}
			edge := mapping[oldEdge.ID]
			edge.Complemented = edge.Complemented != oldEdge.Complemented
			if createErr = result.AddFanin(node, edge); createErr != nil {
				return nil, createErr
			}
		}
		if createErr = result.SetNodeFunction(node, old.Function); createErr != nil {
			return nil, createErr
		}
		mapping[oldID], valid[oldID] = ABCObjectEdge{ID: node}, true
		name, nameErr := abcFunctionalDrivenNetName(network, oldID)
		if nameErr != nil {
			return nil, nameErr
		}
		if name != "" {
			if err = result.SetObjectName(node, name); err != nil {
				return nil, err
			}
		}
	}
	for _, oldID := range network.pos {
		old := &network.objects[oldID]
		if len(old.Fanins) != 1 || !valid[old.Fanins[0].ID] {
			return nil, ErrInvariant
		}
		edge := mapping[old.Fanins[0].ID]
		edge.Complemented = edge.Complemented != old.Fanins[0].Complemented
		po, createErr := result.CreatePO()
		if createErr == nil {
			createErr = result.AddFanin(po, edge)
		}
		if createErr != nil {
			return nil, createErr
		}
		if old.Name != "" {
			if err = result.SetObjectName(po, old.Name); err != nil {
				return nil, err
			}
		}
	}
	if err = result.Validate(); err != nil {
		return nil, err
	}
	return result, nil
}
