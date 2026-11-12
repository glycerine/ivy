package abc

import (
	"fmt"
	"sort"
	"time"
)

// The original traversal IDs remain observable, while vectors are Go slices.
// Netlist traversals skip the intervening net when visiting logic objects.
func dfsDriver(o *Object) *Object {
	if o != nil && o.Kind == ObjectNet && len(o.fanins) > 0 {
		return o.fanins[0].Object
	}
	return o
}
func dfsOutput(o *Object) *Object {
	if o == nil || len(o.fanins) == 0 {
		return nil
	}
	return dfsDriver(o.fanins[0].Object)
}
func dfsFanouts(o *Object) []*Object {
	if o.network.Type == uint8(ABC_NTK_NETLIST) && len(o.fanouts) > 0 && o.fanouts[0].Kind == ObjectNet {
		return o.fanouts[0].fanouts
	}
	return o.fanouts
}
func dfsStop(o *Object) bool    { return o.IsCI() || o.Kind == ObjectConst1 }
func dfsVisited(o *Object) bool { return o.travID == o.network.travID }
func dfsMark(o *Object)         { o.travID = o.network.travID }
func dfsRoots(roots []*Object, count int32) []*Object {
	if count < 0 || int(count) > len(roots) {
		panic("invalid DFS root count")
	}
	return roots[:count]
}
func Abc_NtkDfs_rec(o *Object, nodes *[]*Object) {
	if o == nil || dfsVisited(o) {
		return
	}
	dfsMark(o)
	if dfsStop(o) {
		return
	}
	for _, edge := range o.fanins {
		Abc_NtkDfs_rec(dfsDriver(edge.Object), nodes)
	}
	*nodes = append(*nodes, o)
}
func ntkDfs(n *Network, all int32, barriers bool) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	if barriers {
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			if o.Barrier {
				dfsMark(o)
				Abc_NtkDfs_rec(dfsOutput(o), &nodes)
				nodes = append(nodes, o)
			}
		}
	}
	for _, o := range n.cos {
		dfsMark(o)
		Abc_NtkDfs_rec(dfsOutput(o), &nodes)
	}
	if all != 0 {
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			Abc_NtkDfs_rec(o, &nodes)
		}
	}
	return nodes
}
func Abc_NtkDfs(n *Network, all int32) []*Object { return ntkDfs(n, all, true) }
func Abc_NtkDfs2(n *Network) []*Object           { return ntkDfs(n, 0, false) }
func Abc_NtkDfsNodes(n *Network, roots []*Object, count int32) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range dfsRoots(roots, count) {
		if o.IsCO() {
			dfsMark(o)
			o = dfsOutput(o)
		}
		Abc_NtkDfs_rec(o, &nodes)
	}
	return nodes
}
func dfsReverseResult(o *Object) *Object {
	if o.network.Type == uint8(ABC_NTK_NETLIST) && len(o.fanouts) > 0 {
		return o.fanouts[0]
	}
	return o
}
func Abc_NtkDfsReverse_rec(o *Object, nodes *[]*Object) {
	if o == nil || dfsVisited(o) {
		return
	}
	dfsMark(o)
	if o.IsCO() {
		return
	}
	for _, fanout := range dfsFanouts(o) {
		Abc_NtkDfsReverse_rec(fanout, nodes)
	}
	*nodes = append(*nodes, dfsReverseResult(o))
}
func Abc_NtkDfsReverse(n *Network) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range n.cis {
		dfsMark(o)
		for _, fanout := range dfsFanouts(o) {
			Abc_NtkDfsReverse_rec(fanout, &nodes)
		}
	}
	if n.Type != uint8(ABC_NTK_STRASH) {
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			if len(o.fanins) == 0 {
				nodes = append(nodes, o)
			}
		}
	}
	return nodes
}
func dfsLink(levels *[]*Object, o *Object, level int) {
	for len(*levels) <= level {
		*levels = append(*levels, nil)
	}
	o.Copy = (*levels)[level]
	(*levels)[level] = o
}
func Abc_NtkDfsReverseNodes_rec(o *Object, levels *[]*Object) {
	if o == nil || dfsVisited(o) {
		return
	}
	dfsMark(o)
	if o.IsCO() {
		return
	}
	for _, fanout := range dfsFanouts(o) {
		Abc_NtkDfsReverseNodes_rec(fanout, levels)
	}
	dfsLink(levels, o, o.Level)
}
func Abc_NtkDfsReverseNodes(n *Network, roots []*Object, count int32) []*Object {
	Abc_NtkIncrementTravId(n)
	levels := make([]*Object, int(Abc_AigLevel(n))+1)
	for _, o := range dfsRoots(roots, count) {
		if !o.IsCI() {
			panic("reverse DFS root is not a CI")
		}
		dfsMark(o)
		for _, fanout := range dfsFanouts(o) {
			Abc_NtkDfsReverseNodes_rec(fanout, &levels)
		}
	}
	return levels
}
func Abc_NtkDfsReverseNodesContained(n *Network, roots []*Object, count int32) []*Object {
	maximum := int(Abc_NtkLevel(n))
	Abc_NtkIncrementTravId(n)
	levels := make([]*Object, maximum+2)
	queues := make([][]*Object, maximum+2)
	for _, o := range dfsRoots(roots, count) {
		if !o.IsCI() {
			panic("contained DFS root is not a CI")
		}
		if dfsVisited(o) {
			continue
		}
		dfsMark(o)
		dfsLink(&levels, o, 0)
	}
	for o := levels[0]; o != nil; o = o.Copy {
		queues[0] = append(queues[0], o)
	}
	// A transparent barrier may have the same level as its input. Explicit
	// queues keep newly linked nodes reachable even when prepended to that level.
	for level := 0; level <= maximum; level++ {
		for index := 0; index < len(queues[level]); index++ {
			o := queues[level][index]
			for _, fanout := range o.fanouts {
				if dfsVisited(fanout) {
					continue
				}
				contained := true
				for _, edge := range fanout.fanins {
					if !dfsVisited(edge.Object) {
						contained = false
						break
					}
				}
				if !contained {
					continue
				}
				dfsMark(fanout)
				target := fanout.Level
				if fanout.IsCO() {
					target = maximum + 1
				}
				dfsLink(&levels, fanout, target)
				queues[target] = append(queues[target], fanout)
				if fanout.IsCO() {
					fanout.Level = 0
				}
			}
		}
	}
	return levels
}
func Abc_NtkDfsSeq_rec(o *Object, nodes *[]*Object) {
	if o == nil || dfsVisited(o) {
		return
	}
	dfsMark(o)
	for _, edge := range o.fanins {
		Abc_NtkDfsSeq_rec(edge.Object, nodes)
	}
	*nodes = append(*nodes, o)
}
func Abc_NtkDfsSeq(n *Network) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range n.ObjectsOfKind(ObjectPO) {
		Abc_NtkDfsSeq_rec(o, &nodes)
	}
	for _, o := range n.ObjectsOfKind(ObjectPI) {
		Abc_NtkDfsSeq_rec(o, &nodes)
	}
	return nodes
}
func Abc_NtkDfsSeqReverse_rec(o *Object, nodes *[]*Object) {
	if o == nil || dfsVisited(o) {
		return
	}
	dfsMark(o)
	for _, fanout := range o.fanouts {
		Abc_NtkDfsSeqReverse_rec(fanout, nodes)
	}
	*nodes = append(*nodes, o)
}
func Abc_NtkDfsSeqReverse(n *Network) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range n.ObjectsOfKind(ObjectPI) {
		Abc_NtkDfsSeqReverse_rec(o, &nodes)
	}
	for _, o := range n.ObjectsOfKind(ObjectPO) {
		Abc_NtkDfsSeq_rec(o, &nodes)
	}
	return nodes
}

type DFSFrame struct {
	Node *Object
	Next int
}

func Abc_NtkDfs_iter(stack *[]DFSFrame, root *Object, nodes *[]*Object) {
	if root == nil || dfsVisited(root) {
		return
	}
	dfsMark(root)
	if dfsStop(root) {
		return
	}
	*stack = append((*stack)[:0], DFSFrame{Node: root})
	for len(*stack) > 0 {
		index := len(*stack) - 1
		frame := &(*stack)[index]
		if frame.Next == len(frame.Node.fanins) {
			*nodes = append(*nodes, frame.Node)
			*stack = (*stack)[:index]
			continue
		}
		child := dfsDriver(frame.Node.fanins[frame.Next].Object)
		frame.Next++
		if dfsVisited(child) {
			continue
		}
		dfsMark(child)
		if !dfsStop(child) {
			*stack = append(*stack, DFSFrame{Node: child})
		}
	}
}
func Abc_NtkDfsIter(n *Network, all int32) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	var stack []DFSFrame
	for _, o := range n.cos {
		dfsMark(o)
		Abc_NtkDfs_iter(&stack, dfsOutput(o), &nodes)
	}
	if all != 0 {
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			Abc_NtkDfs_iter(&stack, o, &nodes)
		}
	}
	return nodes
}
func Abc_NtkDfsIterNodes(n *Network, roots []*Object) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	var stack []DFSFrame
	for _, o := range roots {
		Abc_NtkDfs_iter(&stack, o, &nodes)
	}
	return nodes
}
func Abc_NtkDfsHie_rec(o *Object, nodes *[]*Object) { Abc_NtkDfsSeq_rec(o, nodes) }
func Abc_NtkDfsHie(n *Network, all int32) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range n.ObjectsOfKind(ObjectPO) {
		Abc_NtkDfsHie_rec(o, &nodes)
	}
	if all != 0 {
		for _, o := range n.Objects() {
			Abc_NtkDfsHie_rec(o, &nodes)
		}
	}
	return nodes
}
func Abc_NtkIsDfsOrdered(n *Network) int32 {
	Abc_NtkIncrementTravId(n)
	for _, o := range n.cis {
		dfsMark(o)
	}
	dfsMark(n.constant)
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		for _, edge := range o.fanins {
			if !dfsVisited(edge.Object) {
				return 0
			}
		}
		seenChoices := make(map[*Object]bool)
		for choice := o.Choice; choice != nil; choice = choice.Choice {
			if seenChoices[choice] {
				return 0
			}
			seenChoices[choice] = true
			if !dfsVisited(choice) {
				return 0
			}
		}
		dfsMark(o)
	}
	return 1
}
func Abc_NtkDfsNets_rec(net *Object, nets *[]*Object) {
	if net == nil || dfsVisited(net) {
		return
	}
	dfsMark(net)
	driver := dfsOutput(net)
	if driver == nil {
		return
	}
	for _, edge := range driver.fanins {
		Abc_NtkDfsNets_rec(edge.Object, nets)
	}
	for _, fanout := range driver.fanouts {
		*nets = append(*nets, fanout)
	}
}
func Abc_NtkDfsNets(n *Network) []*Object {
	Abc_NtkIncrementTravId(n)
	var nets []*Object
	for _, o := range n.cis {
		if len(o.fanouts) > 0 {
			net := o.fanouts[0]
			dfsMark(net)
			nets = append(nets, net)
		}
	}
	for _, o := range n.cos {
		if len(o.fanins) > 0 {
			Abc_NtkDfsNets_rec(o.fanins[0].Object, &nets)
		}
	}
	return nets
}
func Abc_NtkDfsWithBoxes_rec(o *Object, nodes *[]*Object) {
	if o == nil {
		return
	}
	if o.Kind == ObjectBO {
		o = dfsOutput(o)
	}
	if o.Kind == ObjectPI || dfsVisited(o) {
		return
	}
	dfsMark(o)
	for _, edge := range o.fanins {
		child := edge.Object
		if Abc_ObjIsBox(o) != 0 {
			child = dfsOutput(child)
		}
		Abc_NtkDfsWithBoxes_rec(dfsDriver(child), nodes)
	}
	*nodes = append(*nodes, o)
}
func Abc_NtkDfsWithBoxes(n *Network) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range n.ObjectsOfKind(ObjectPO) {
		Abc_NtkDfsWithBoxes_rec(dfsOutput(o), &nodes)
	}
	return nodes
}
func Abc_NtkNodeSupport_rec(o *Object, nodes *[]*Object) {
	if o == nil || dfsVisited(o) {
		return
	}
	dfsMark(o)
	if dfsStop(o) {
		*nodes = append(*nodes, o)
		return
	}
	for _, edge := range o.fanins {
		Abc_NtkNodeSupport_rec(dfsDriver(edge.Object), nodes)
	}
}
func Abc_NtkSupport(n *Network) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range n.cos {
		Abc_NtkNodeSupport_rec(dfsOutput(o), &nodes)
	}
	filtered := nodes[:0]
	for _, o := range nodes {
		if o.IsCI() {
			filtered = append(filtered, o)
		}
	}
	nodes = filtered
	for _, o := range n.cis {
		if !dfsVisited(o) {
			nodes = append(nodes, o)
		}
	}
	return nodes
}
func Abc_NtkNodeSupport(n *Network, roots []*Object, count int32) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range dfsRoots(roots, count) {
		if o.IsCO() {
			o = dfsOutput(o)
		}
		Abc_NtkNodeSupport_rec(o, &nodes)
	}
	return nodes
}
func Abc_NtkNodeSupportInt_rec(o *Object, ids *[]int32) {
	var nodes []*Object
	Abc_NtkNodeSupport_rec(o, &nodes)
	indices := make(map[*Object]int32)
	for i, ci := range o.network.cis {
		indices[ci] = int32(i)
	}
	for _, leaf := range nodes {
		if leaf.IsCI() {
			*ids = append(*ids, indices[leaf])
		}
	}
}
func Abc_NtkNodeSupportInt(n *Network, output int32) []int32 {
	if output < 0 || int(output) >= len(n.cos) {
		return nil
	}
	Abc_NtkIncrementTravId(n)
	var ids []int32
	Abc_NtkNodeSupportInt_rec(dfsOutput(n.cos[output]), &ids)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
func Abc_ObjSuppSize_rec(o *Object) int32 {
	if o == nil || dfsVisited(o) {
		return 0
	}
	dfsMark(o)
	if o.Kind == ObjectPI {
		return 1
	}
	var count int32
	for _, edge := range o.fanins {
		count += Abc_ObjSuppSize_rec(edge.Object)
	}
	return count
}
func Abc_ObjSuppSize(o *Object) int32 {
	Abc_NtkIncrementTravId(o.network)
	return Abc_ObjSuppSize_rec(o)
}
func Abc_NtkSuppSizeTest(n *Network) int32 {
	start := time.Now()
	var count int32
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if Abc_ObjSuppSize(o) <= 16 {
			count++
		}
	}
	fmt.Printf("Nodes with small support %d (out of %d)\nTime = %.2f sec\n", count, len(n.ObjectsOfKind(ObjectNode)), time.Since(start).Seconds())
	return count
}
func Abc_NtkSupportSum(n *Network) {
	count := 0
	for _, o := range n.cos {
		count += len(Abc_NtkNodeSupport(n, []*Object{o}, 1))
	}
	fmt.Printf("Total supports = %d.\n", count)
}
func Abc_AigDfs_rec(o *Object, nodes *[]*Object) {
	if o == nil || dfsVisited(o) {
		return
	}
	dfsMark(o)
	if dfsStop(o) {
		return
	}
	for _, edge := range o.fanins {
		Abc_AigDfs_rec(edge.Object, nodes)
	}
	if o.Choice != nil {
		Abc_AigDfs_rec(o.Choice, nodes)
	}
	*nodes = append(*nodes, o)
}
func Abc_AigDfs(n *Network, all, outputs int32) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	for _, o := range n.cos {
		Abc_AigDfs_rec(dfsOutput(o), &nodes)
		dfsMark(o)
		if outputs != 0 {
			nodes = append(nodes, o)
		}
	}
	if all != 0 {
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			Abc_AigDfs_rec(o, &nodes)
		}
	}
	return nodes
}
func Abc_AigDfsMap(n *Network) []*Object {
	Abc_NtkIncrementTravId(n)
	var nodes []*Object
	split := len(n.cos) - n.Barriers
	if split < 0 {
		panic("too many barriers")
	}
	for _, o := range n.cos[split:] {
		Abc_AigDfs_rec(dfsOutput(o), &nodes)
		dfsMark(o)
		if len(o.fanouts) == 0 || o.fanouts[0].Kind != ObjectLatch {
			panic("barrier output has no latch")
		}
		nodes = append(nodes, o.fanouts[0])
	}
	for _, o := range n.cos[:split] {
		Abc_AigDfs_rec(dfsOutput(o), &nodes)
		dfsMark(o)
	}
	return nodes
}

func Abc_DfsLevelizedTfo_rec(o *Object, levels *[][]*Object) {
	if o == nil || dfsVisited(o) {
		return
	}
	dfsMark(o)
	if o.IsCO() {
		return
	}
	for len(*levels) <= o.Level {
		*levels = append(*levels, nil)
	}
	(*levels)[o.Level] = append((*levels)[o.Level], o)
	for _, fanout := range o.fanouts {
		Abc_DfsLevelizedTfo_rec(fanout, levels)
	}
}
func Abc_DfsLevelized(o *Object, tfi int32) [][]*Object {
	if tfi != 0 {
		panic("levelized TFI is not supported by the original API")
	}
	Abc_NtkIncrementTravId(o.network)
	var levels [][]*Object
	if o.IsCI() {
		dfsMark(o)
		for _, fanout := range o.fanouts {
			Abc_DfsLevelizedTfo_rec(fanout, &levels)
		}
	} else {
		Abc_DfsLevelizedTfo_rec(o, &levels)
	}
	return levels
}
func Abc_NtkLevel_rec(o *Object) int32 {
	if o == nil {
		return 0
	}
	if o.IsCI() || dfsVisited(o) {
		return int32(o.Level)
	}
	dfsMark(o)
	o.Level = 0
	for _, edge := range o.fanins {
		level := int(Abc_NtkLevel_rec(dfsDriver(edge.Object)))
		if level > o.Level {
			o.Level = level
		}
	}
	if len(o.fanins) > 0 && !o.Barrier {
		o.Level++
	}
	return int32(o.Level)
}
func Abc_NtkLevelReverse_rec(o *Object) int32 {
	if o == nil {
		return 0
	}
	if o.IsCO() || dfsVisited(o) {
		return int32(o.Level)
	}
	dfsMark(o)
	o.Level = 0
	for _, fanout := range dfsFanouts(o) {
		level := int(Abc_NtkLevelReverse_rec(fanout))
		if level > o.Level {
			o.Level = level
		}
	}
	if len(o.fanins) > 0 && !o.Barrier {
		o.Level++
	}
	return int32(o.Level)
}
func Abc_NtkLevel(n *Network) int32 {
	for _, o := range n.cis {
		o.Level = 0
		if n.CILevel != nil {
			o.Level = n.CILevel(o)
			if o.Level < 0 {
				o.Level = 0
			}
		}
	}
	Abc_NtkIncrementTravId(n)
	var maximum int32
	if n.Barriers == 0 {
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			level := Abc_NtkLevel_rec(o)
			if level > maximum {
				maximum = level
			}
		}
	} else {
		split := len(n.cos) - n.Barriers
		if split < 0 {
			panic("too many barrier outputs")
		}
		outputs := append(append([]*Object(nil), n.cos[split:]...), n.cos[:split]...)
		for i, o := range outputs {
			driver := dfsOutput(o)
			level := Abc_NtkLevel_rec(driver)
			if level > maximum {
				maximum = level
			}
			if i < n.Barriers {
				if len(o.fanouts) != 1 || len(o.fanouts[0].fanouts) != 1 {
					panic("invalid barrier latch")
				}
				o.fanouts[0].fanouts[0].Level = int(level)
			}
		}
	}
	return maximum
}
func Abc_NtkLevelize(n *Network) [][]*Object {
	levels := make([][]*Object, int(Abc_NtkLevel(n))+1)
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		levels[o.Level] = append(levels[o.Level], o)
	}
	return levels
}
func Abc_NtkLevelReverse(n *Network) int32 {
	for _, o := range n.cos {
		o.Level = 0
	}
	Abc_NtkIncrementTravId(n)
	var maximum int32
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		level := Abc_NtkLevelReverse_rec(o)
		if level > maximum {
			maximum = level
		}
	}
	return maximum
}
func dfsCycleVisit(o *Object, boxes bool) bool {
	if o == nil {
		return true
	}
	n := o.network
	if boxes {
		if o.Kind == ObjectBO {
			o = dfsOutput(o)
		}
		if o == nil || o.Kind == ObjectPI || o.Kind == ObjectLatch || o.Kind == ObjectBlackbox {
			return true
		}
	} else if dfsStop(o) || Abc_ObjIsBox(o) != 0 {
		return true
	}
	if o.travID == n.travID-1 {
		return true
	}
	if dfsVisited(o) {
		fmt.Printf("Network %q contains combinational loop at %q.\n", n.Name, o.Name())
		return false
	}
	dfsMark(o)
	for _, edge := range o.fanins {
		child := edge.Object
		if boxes && Abc_ObjIsBox(o) != 0 {
			child = dfsOutput(child)
		}
		child = dfsDriver(child)
		if !dfsCycleVisit(child, boxes) {
			return false
		}
	}
	if !boxes && o.Choice != nil {
		if !dfsCycleVisit(o.Choice, false) {
			return false
		}
	}
	o.travID = n.travID - 1
	return true
}
func Abc_NtkIsAcyclic_rec(o *Object) int32 { return abcBool(dfsCycleVisit(o, false)) }
func dfsCycleStart(n *Network) {
	Abc_NtkIncrementTravId(n)
	Abc_NtkIncrementTravId(n)
	if n.travID == 1 {
		Abc_NtkIncrementTravId(n)
	}
}
func Abc_NtkIsAcyclic(n *Network) int32 {
	dfsCycleStart(n)
	for _, o := range n.cos {
		if !dfsCycleVisit(dfsOutput(o), false) {
			return 0
		}
	}
	return 1
}
func Abc_NtkIsAcyclicWithBoxes_rec(o *Object) int32 { return abcBool(dfsCycleVisit(o, true)) }
func Abc_NtkIsAcyclicWithBoxes(n *Network) int32 {
	dfsCycleStart(n)
	for _, o := range n.ObjectsOfKind(ObjectPO) {
		if !dfsCycleVisit(dfsOutput(o), true) {
			return 0
		}
	}
	for _, latch := range n.ObjectsOfKind(ObjectLatch) {
		for _, edge := range latch.fanins {
			if !dfsCycleVisit(dfsOutput(edge.Object), true) {
				return 0
			}
		}
	}
	return 1
}
func Abc_NodeSetChoiceLevel_rec(o *Object, maximum int32) int32 {
	if dfsVisited(o) {
		return int32(o.Temp)
	}
	dfsMark(o)
	level := 0
	for _, edge := range o.fanins {
		child := int(Abc_NodeSetChoiceLevel_rec(edge.Object, maximum))
		if child > level {
			level = child
		}
	}
	if len(o.fanins) > 0 {
		level++
	}
	if o.Choice != nil {
		other := int(Abc_NodeSetChoiceLevel_rec(o.Choice, maximum))
		if maximum != 0 && other > level || maximum == 0 && other < level {
			level = other
		}
		seen := make(map[*Object]bool)
		for choice := o.Choice; choice != nil && !seen[choice]; choice = choice.Choice {
			seen[choice] = true
			choice.Temp = level
		}
	}
	o.Temp = level
	return int32(level)
}
func Abc_AigSetChoiceLevels(n *Network) int32 {
	for _, o := range n.Objects() {
		o.Temp = 0
	}
	Abc_NtkIncrementTravId(n)
	for _, o := range n.cis {
		dfsMark(o)
		o.Temp = 0
	}
	dfsMark(n.constant)
	n.constant.Temp = 0
	var maximum int32
	for _, o := range n.cos {
		level := Abc_NodeSetChoiceLevel_rec(dfsOutput(o), 1)
		if level > maximum {
			maximum = level
		}
	}
	return maximum
}
func Abc_AigGetLevelizedOrder(n *Network, collectCIs int32) []*Object {
	Abc_AigSetChoiceLevels(n)
	var nodes []*Object
	if collectCIs != 0 {
		nodes = append(nodes, n.cis...)
	}
	levels := make(map[int][]*Object)
	var keys []int
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if _, ok := levels[o.Temp]; !ok {
			keys = append(keys, o.Temp)
		}
		levels[o.Temp] = append(levels[o.Temp], o)
	}
	sort.Ints(keys)
	for _, level := range keys {
		objects := levels[level]
		for i := len(objects) - 1; i >= 0; i-- {
			if level > 0 || collectCIs != 0 {
				nodes = append(nodes, objects[i])
			}
		}
	}
	return nodes
}
func Abc_ObjSugraphSize(o *Object) int32 {
	if o == nil || dfsStop(o) || len(o.fanouts) > 1 {
		return 0
	}
	count := int32(1)
	for _, edge := range o.fanins {
		count += Abc_ObjSugraphSize(edge.Object)
	}
	return count
}
func dfsXorType(o *Object) bool {
	if len(o.fanins) != 2 || !o.fanins[0].Complement || !o.fanins[1].Complement {
		return false
	}
	a, b := o.fanins[0].Object, o.fanins[1].Object
	if len(a.fanins) != 2 || len(b.fanins) != 2 {
		return false
	}
	left, right := edgePair(a.fanins[0], a.fanins[1]), edgePair(b.fanins[0], b.fanins[1])
	return left[0].Object == right[0].Object && left[1].Object == right[1].Object && left[0].Complement != right[0].Complement && left[1].Complement != right[1].Complement
}
func Abc_NtkPrintSubraphSizes(n *Network) int32 {
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if len(o.fanouts) > 1 && !dfsXorType(o) {
			count := int32(1)
			for _, edge := range o.fanins {
				count += Abc_ObjSugraphSize(edge.Object)
			}
			fmt.Printf("%d(%d) ", count, len(o.fanouts))
		}
	}
	fmt.Println()
	return 1
}
