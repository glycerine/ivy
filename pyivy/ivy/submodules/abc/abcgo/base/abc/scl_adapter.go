package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"strconv"
)

// ConfigureSCL connects timing to this network's actual mapped Object.Data.
// Topology remains in Network; only library assignments are staged by SCL.
func (n *Network) ConfigureSCL(library *scl.SC_Lib, gates *mio.Mio_Library_t) error {
	if n.Type != uint8(ABC_NTK_LOGIC) || n.Function != uint8(ABC_FUNC_MAP) {
		return fmt.Errorf("SCL requires mapped logic")
	}
	if err := n.Check(); err != nil {
		return err
	}
	if len(n.ObjectsOfKind(ObjectNet)) > 0 {
		return fmt.Errorf("SCL mapped logic cannot contain explicit nets")
	}
	if library == nil || gates == nil {
		return fmt.Errorf("SCL libraries are missing")
	}
	mapping := &scl.NetworkMapping{Lib: library, Mio: gates, MioGates: make([]*mio.Mio_Gate_t, len(n.objects))}
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		gate, ok := o.Data.(*mio.Mio_Gate_t)
		if !ok || gate == nil {
			return fmt.Errorf("object %d has no MIO gate", o.ID)
		}
		if int(mio.Mio_GateReadPinNum(gate)) != len(o.fanins) {
			return fmt.Errorf("object %d gate arity mismatch", o.ID)
		}
		mapping.MioGates[o.ID] = gate
	}
	n.FunctionManager = gates
	n.timingMapping = mapping
	for _, o := range n.Objects() {
		for _, edge := range o.fanins {
			n.phaseTracking = n.phaseTracking || edge.Complement
		}
	}
	return nil
}
func (n *Network) SclName() string { return n.Name }
func (n *Network) SclObjectIDs() []int {
	objects := n.Objects()
	ids := make([]int, len(objects))
	for i, o := range objects {
		ids[i] = o.ID
	}
	return ids
}
func (n *Network) SclInputs() []int {
	ids := make([]int, len(n.cis))
	for i, o := range n.cis {
		ids[i] = o.ID
	}
	return ids
}
func (n *Network) SclOutputs() []int {
	ids := make([]int, len(n.cos))
	for i, o := range n.cos {
		ids[i] = o.ID
	}
	return ids
}
func (n *Network) SclFanins(id int) []int {
	o := n.Object(id)
	if o == nil {
		return nil
	}
	ids := make([]int, len(o.fanins))
	for i, edge := range o.fanins {
		ids[i] = edge.Object.ID
	}
	return ids
}
func (n *Network) SclFanouts(id int) []int {
	o := n.Object(id)
	if o == nil {
		return nil
	}
	ids := make([]int, len(o.fanouts))
	for i, fanout := range o.fanouts {
		ids[i] = fanout.ID
	}
	return ids
}
func (n *Network) SclIsNode(id int) bool    { o := n.Object(id); return o != nil && o.Kind == ObjectNode }
func (n *Network) SclIsBarrier(id int) bool { o := n.Object(id); return o != nil && o.Barrier }
func (n *Network) SclObjectName(id int) string {
	o := n.Object(id)
	if o == nil {
		return ""
	}
	if name := o.Name(); name != "" {
		return name
	}
	return strconv.Itoa(id)
}
func (n *Network) SclMapping() *scl.NetworkMapping {
	if n.Function != uint8(ABC_FUNC_MAP) || n.Type != uint8(ABC_NTK_LOGIC) {
		return nil
	}
	mapping, _ := n.timingMapping.(*scl.NetworkMapping)
	if mapping == nil {
		gates, ok := n.FunctionManager.(*mio.Mio_Library_t)
		if !ok {
			return nil
		}
		mapping = &scl.NetworkMapping{Mio: gates}
		n.timingMapping = mapping
	}
	for len(mapping.MioGates) < len(n.objects) {
		mapping.MioGates = append(mapping.MioGates, nil)
	}
	// An external change to Object.Data invalidates previously derived cell IDs.
	// During SCL's explicit commit, use the staged vector directly instead.
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		gate, ok := o.Data.(*mio.Mio_Gate_t)
		if !ok || gate == nil {
			return nil
		}
		if mapping.MioGates[o.ID] != gate {
			mapping.MioGates[o.ID] = gate
			mapping.GateIDs = nil
		}
	}
	return mapping
}
func (n *Network) SclCommitMapping() error {
	mapping, ok := n.timingMapping.(*scl.NetworkMapping)
	if !ok || mapping.Mio == nil {
		return fmt.Errorf("SCL mapping is unavailable")
	}
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		if o.ID >= len(mapping.MioGates) || mapping.MioGates[o.ID] == nil {
			return fmt.Errorf("object %d has no staged MIO gate", o.ID)
		}
		if int(mio.Mio_GateReadPinNum(mapping.MioGates[o.ID])) != len(o.fanins) {
			return fmt.Errorf("object %d staged gate arity mismatch", o.ID)
		}
	}
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if !o.Barrier {
			o.Data = mapping.MioGates[o.ID]
		}
	}
	n.FunctionManager = mapping.Mio
	n.Function = uint8(ABC_FUNC_MAP)
	return nil
}
func (n *Network) SclPatchFanin(id, pin, newID int) error {
	o, target := n.Object(id), n.Object(newID)
	if o == nil || target == nil || pin < 0 || pin >= len(o.fanins) {
		return fmt.Errorf("invalid SCL fanin patch")
	}
	return n.PatchFanin(o, pin, Edge{Object: target, Complement: o.fanins[pin].Complement})
}
func (n *Network) SclDeleteNode(id int) error {
	o := n.Object(id)
	if o == nil || o.Kind != ObjectNode {
		return fmt.Errorf("invalid SCL node deletion")
	}
	if err := n.DeleteObject(o); err != nil {
		return err
	}
	if mapping, ok := n.timingMapping.(*scl.NetworkMapping); ok {
		if id < len(mapping.GateIDs) {
			mapping.GateIDs[id] = -1
		}
		if id < len(mapping.MioGates) {
			mapping.MioGates[id] = nil
		}
	}
	return nil
}
func (n *Network) SclHasPhases() bool {
	if n.phaseTracking {
		return true
	}
	for _, o := range n.Objects() {
		for _, edge := range o.fanins {
			if edge.Complement {
				return true
			}
		}
	}
	return false
}
func (n *Network) SclFaninPhase(id, pin int) bool {
	o := n.Object(id)
	return o != nil && pin >= 0 && pin < len(o.fanins) && o.fanins[pin].Complement
}
func (n *Network) SclToggleFaninPhase(id, pin int) error {
	o := n.Object(id)
	if !n.SclHasPhases() || o == nil || pin < 0 || pin >= len(o.fanins) {
		return fmt.Errorf("invalid SCL phase toggle")
	}
	n.phaseTracking = true
	o.fanins[pin].Complement = !o.fanins[pin].Complement
	clear(n.hashed)
	return nil
}
func (n *Network) SclSetPhases(enabled bool) error {
	if !enabled {
		for _, o := range n.Objects() {
			for _, edge := range o.fanins {
				if edge.Complement {
					return fmt.Errorf("cannot discard unmaterialized complemented edges")
				}
			}
		}
	}
	n.phaseTracking = enabled
	return nil
}
func (n *Network) SclCreateNode(fanins []int, gateID int32) (int, error) {
	mapping := n.SclMapping()
	if mapping == nil || mapping.Lib == nil || mapping.Mio == nil || gateID < 0 || int(gateID) >= len(mapping.Lib.Cells) {
		return -1, fmt.Errorf("invalid SCL gate")
	}
	cell := mapping.Lib.Cells[gateID]
	if cell == nil || cell.Inputs != len(fanins) {
		return -1, fmt.Errorf("SCL gate arity mismatch")
	}
	gate := mio.Mio_LibraryReadGateByName(mapping.Mio, cell.Name, "")
	if gate == nil {
		return -1, fmt.Errorf("MIO gate %s is unavailable", cell.Name)
	}
	sources := make([]*Object, len(fanins))
	for i, id := range fanins {
		sources[i] = n.Object(id)
		if sources[i] == nil || sources[i].IsCO() {
			return -1, fmt.Errorf("invalid SCL gate input %d", id)
		}
	}
	o := n.NewObject(ObjectNode)
	o.Data = gate
	for _, source := range sources {
		if err := n.AddFanin(o, source.Signal()); err != nil {
			panic(err)
		}
	}
	for len(mapping.MioGates) <= o.ID {
		mapping.MioGates = append(mapping.MioGates, nil)
	}
	mapping.MioGates[o.ID] = gate
	if len(mapping.GateIDs) > 0 {
		for len(mapping.GateIDs) <= o.ID {
			mapping.GateIDs = append(mapping.GateIDs, -1)
		}
		mapping.GateIDs[o.ID] = gateID
	}
	return o.ID, nil
}
func (n *Network) sclClone(topological bool) (scl.TimingNetwork, error) {
	if err := n.Check(); err != nil {
		return nil, err
	}
	mapping := n.SclMapping()
	if mapping == nil {
		return nil, fmt.Errorf("mapped network has no SCL assignment")
	}
	var order []*Object
	if topological {
		order = append(order, n.cis...)
		nodes, err := n.Topological(n.ObjectsOfKind(ObjectNode))
		if err != nil {
			return nil, err
		}
		for _, o := range nodes {
			if o.Kind == ObjectNode {
				order = append(order, o)
			}
		}
		order = append(order, n.cos...)
	}
	order = append(order, n.Objects()...)
	clone := copyNetworkMetadata(n)
	clone.phaseTracking = n.phaseTracking
	copies := map[*Object]*Object{n.constant: clone.constant}
	reverse := map[*Object]*Object{clone.constant: n.constant}
	for _, o := range order {
		if copies[o] != nil {
			continue
		}
		copy := clone.NewObject(o.Kind)
		copy.Data = o.Data
		copy.Barrier = o.Barrier
		copy.Level = o.Level
		copy.Phase = o.Phase
		copyMVVariable(o, copy)
		if name, ok := n.names.Lookup(int32(o.ID)); ok {
			copy.SetName(name)
		}
		copies[o] = copy
		reverse[copy] = o
	}
	for _, o := range n.Objects() {
		copy := copies[o]
		for _, edge := range o.fanins {
			if err := clone.AddFanin(copy, Edge{Object: copies[edge.Object], Complement: edge.Complement}); err != nil {
				return nil, err
			}
		}
		copy.Choice = copies[o.Choice]
	}
	clone.cis = nil
	for _, o := range n.cis {
		clone.cis = append(clone.cis, copies[o])
	}
	clone.cos = nil
	for _, o := range n.cos {
		clone.cos = append(clone.cos, copies[o])
	}
	clone.boxes = nil
	for _, o := range n.boxes {
		clone.boxes = append(clone.boxes, copies[o])
	}
	if n.CILevel != nil {
		levels := make(map[*Object]int)
		for _, o := range n.cis {
			levels[copies[o]] = n.CILevel(o)
		}
		clone.CILevel = func(o *Object) int { return levels[o] }
	}
	copied := &scl.NetworkMapping{Lib: mapping.Lib, Mio: mapping.Mio, MioGates: make([]*mio.Mio_Gate_t, len(clone.objects))}
	if len(mapping.GateIDs) > 0 {
		copied.GateIDs = make([]int32, len(clone.objects))
		for i := range copied.GateIDs {
			copied.GateIDs[i] = -1
		}
	}
	for copy, original := range reverse {
		if original.ID < len(mapping.MioGates) {
			copied.MioGates[copy.ID] = mapping.MioGates[original.ID]
		}
		if len(copied.GateIDs) > 0 && original.ID < len(mapping.GateIDs) {
			copied.GateIDs[copy.ID] = mapping.GateIDs[original.ID]
		}
	}
	clone.timingMapping = copied
	return clone, clone.Check()
}
func (n *Network) SclClone() (scl.TimingNetwork, error)            { return n.sclClone(false) }
func (n *Network) SclCloneTopological() (scl.TimingNetwork, error) { return n.sclClone(true) }
