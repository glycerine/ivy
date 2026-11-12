package res

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

// Res_WndStrash rebuilds a resynthesis window as a structurally hashed AIG.
// Temporary object copies are restored before returning, leaving no stale
// mapping state on the source network.
func Res_WndStrash(p *Res_Win_t) *abc.Network {
	network, err := resWndStrash(p)
	if err != nil {
		panic(err)
	}
	return network
}

func resWndStrash(p *Res_Win_t) (*abc.Network, error) {
	if p == nil || p.pNode == nil {
		return nil, fmt.Errorf("res: window has no center node")
	}
	source := p.pNode.Network()
	if source == nil {
		return nil, fmt.Errorf("res: center node has no source network")
	}
	if _, ok := source.FunctionManager.(*hop.Hop_Man_t); !ok {
		return nil, fmt.Errorf("res: source network has no Hop function manager")
	}
	if p.vRoots == nil || p.vLeaves == nil || p.vBranches == nil || p.vNodes == nil || p.vDivs == nil {
		return nil, fmt.Errorf("res: window vectors must be initialized")
	}

	network := abc.NewNetwork("window")
	network.Type = uint8(abc.ABC_NTK_STRASH)
	network.Function = uint8(abc.ABC_FUNC_AIG)
	manager := abc.Abc_AigAlloc(network)
	network.FunctionManager = manager

	copyEdges := make(map[*abc.Object]abc.Edge)
	type copyState struct {
		object *abc.Object
		edge   abc.Edge
	}
	previousCopyEdges := make(map[*abc.Object]copyState)
	bind := func(old *abc.Object, edge abc.Edge) error {
		if old == nil || edge.Object == nil {
			return fmt.Errorf("res: cannot map a nil window object")
		}
		if _, seen := previousCopyEdges[old]; !seen {
			previousCopyEdges[old] = copyState{object: old.Copy, edge: old.CopyEdge}
		}
		old.SetCopyEdge(edge)
		copyEdges[old] = edge
		return nil
	}
	defer func() {
		for old, state := range previousCopyEdges {
			old.Copy = state.object
			old.CopyEdge = state.edge
		}
	}()

	copyObject := func(value any) (*abc.Object, error) {
		old, ok := value.(*abc.Object)
		if !ok || old == nil {
			return nil, fmt.Errorf("res: window vector entry has type %T, want *abc.Object", value)
		}
		if old.Network() != source {
			return nil, fmt.Errorf("res: window object %d belongs to another source network", old.ID)
		}
		return old, nil
	}
	addInputs := func(objects *vec.Vec_Ptr_t) error {
		for i := 0; i < vec.Vec_PtrSize(objects); i++ {
			old, err := copyObject(vec.Vec_PtrEntry(objects, i))
			if err != nil {
				return err
			}
			if _, alreadyMapped := copyEdges[old]; alreadyMapped {
				continue
			}
			pi := network.NewObject(abc.ObjectPI)
			if err := bind(old, pi.Signal()); err != nil {
				return err
			}
		}
		return nil
	}
	if err := addInputs(p.vLeaves); err != nil {
		return nil, err
	}
	if err := addInputs(p.vBranches); err != nil {
		return nil, err
	}

	for i := 0; i < vec.Vec_PtrSize(p.vNodes); i++ {
		old, err := copyObject(vec.Vec_PtrEntry(p.vNodes, i))
		if err != nil {
			return nil, err
		}
		if _, isInput := copyEdges[old]; isInput {
			continue
		}
		if err := bind(old, abc.Abc_ConvertAigToAig(network, old)); err != nil {
			return nil, err
		}
	}

	rootOriginals := make([]abc.Edge, vec.Vec_PtrSize(p.vRoots))
	for i := range rootOriginals {
		root, err := copyObject(vec.Vec_PtrEntry(p.vRoots, i))
		if err != nil {
			return nil, err
		}
		mapped, ok := copyEdges[root]
		if !ok {
			return nil, fmt.Errorf("res: root %d was not included in the window nodes", root.ID)
		}
		rootOriginals[i] = mapped
	}

	center, ok := copyEdges[p.pNode]
	if !ok {
		return nil, fmt.Errorf("res: center node %d was not included in the window nodes", p.pNode.ID)
	}
	if err := bind(p.pNode, center.Not()); err != nil {
		return nil, err
	}

	// Rebuild only nodes in the center's transitive fanout, stopping at output
	// boundaries and the configured level limit, as the original sweep did.
	limit := p.pNode.Level + int(p.nWinTfoMax)
	tfo := resWindowTFO(p.pNode, limit)
	for i := 0; i < vec.Vec_PtrSize(p.vNodes); i++ {
		old, err := copyObject(vec.Vec_PtrEntry(p.vNodes, i))
		if err != nil {
			return nil, err
		}
		if old == p.pNode || !tfo[old] {
			continue
		}
		if err := bind(old, abc.Abc_ConvertAigToAig(network, old)); err != nil {
			return nil, err
		}
	}

	pairs := make([]abc.Edge, 0, 2*len(rootOriginals))
	for i := 0; i < vec.Vec_PtrSize(p.vRoots); i++ {
		root, _ := copyObject(vec.Vec_PtrEntry(p.vRoots, i))
		current, ok := copyEdges[root]
		if !ok {
			return nil, fmt.Errorf("res: updated root %d has no AIG copy", root.ID)
		}
		pairs = append(pairs, rootOriginals[i], current)
	}
	miter := abc.Abc_AigMiter(manager, &pairs, 0)
	resAddOutput(network, miter)

	center, _ = copyEdges[p.pNode]
	resAddOutput(network, center)
	for _, fanin := range p.pNode.Fanins() {
		mapped, ok := copyEdges[fanin.Object]
		if !ok {
			return nil, fmt.Errorf("res: center fanin %d has no AIG copy", fanin.Object.ID)
		}
		if fanin.Complement {
			mapped = mapped.Not()
		}
		resAddOutput(network, mapped)
	}
	for i := 0; i < vec.Vec_PtrSize(p.vDivs); i++ {
		divisor, err := copyObject(vec.Vec_PtrEntry(p.vDivs, i))
		if err != nil {
			return nil, err
		}
		mapped, ok := copyEdges[divisor]
		if !ok {
			return nil, fmt.Errorf("res: divisor %d has no AIG copy", divisor.ID)
		}
		resAddOutput(network, mapped)
	}
	if err := network.Check(); err != nil {
		return nil, fmt.Errorf("res: rebuilt window AIG is invalid: %w", err)
	}
	if abc.Abc_AigCheck(manager) == 0 {
		return nil, fmt.Errorf("res: rebuilt window AIG violates strashing invariants")
	}
	return network, nil
}

func resAddOutput(network *abc.Network, driver abc.Edge) {
	output := network.NewObject(abc.ObjectPO)
	if err := network.AddFanin(output, driver); err != nil {
		panic(err)
	}
}

func resWindowTFO(center *abc.Object, levelLimit int) map[*abc.Object]bool {
	visited := make(map[*abc.Object]bool)
	stack := []*abc.Object{center}
	for len(stack) > 0 {
		last := len(stack) - 1
		object := stack[last]
		stack = stack[:last]
		if object == nil || visited[object] || object.IsCO() || object.Level > levelLimit {
			continue
		}
		visited[object] = true
		stack = append(stack, object.Fanouts()...)
	}
	return visited
}
