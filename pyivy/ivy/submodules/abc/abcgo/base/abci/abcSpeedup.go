package abci

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
)

func speedupLutTimes(n *Abc_Ntk_t) []float32     { return baseabc.Abc_NtkLutTimes(n) }
func speedupTime(o *Abc_Obj_t, slot int) float32 { return speedupLutTimes(o.Network())[3*o.ID+slot] }
func setSpeedupTime(o *Abc_Obj_t, slot int, value float32) {
	speedupLutTimes(o.Network())[3*o.ID+slot] = value
}
func Abc_ObjArrival(o *Abc_Obj_t) float32        { return speedupTime(o, 0) }
func Abc_ObjRequired(o *Abc_Obj_t) float32       { return speedupTime(o, 1) }
func Abc_ObjSlack(o *Abc_Obj_t) float32          { return speedupTime(o, 2) }
func Abc_ObjSetArrival(o *Abc_Obj_t, v float32)  { setSpeedupTime(o, 0, v) }
func Abc_ObjSetRequired(o *Abc_Obj_t, v float32) { setSpeedupTime(o, 1, v) }
func Abc_ObjSetSlack(o *Abc_Obj_t, v float32)    { setSpeedupTime(o, 2, v) }

func Abc_NtkDelayTraceSortPins(node *Abc_Obj_t, perm []int32, delays []float32) {
	n := int(baseabc.Abc_ObjFaninNum(node))
	if len(perm) < n || len(delays) < n {
		panic("delay-trace pin workspace too small")
	}
	for i := 0; i < n; i++ {
		perm[i] = int32(i)
		delays[i] = Abc_ObjArrival(baseabc.Abc_ObjFanin(node, int32(i)))
	}
	for i := 0; i < n-1; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if delays[perm[j]] > delays[perm[best]] {
				best = j
			}
		}
		perm[i], perm[best] = perm[best], perm[i]
	}
}

func speedupLut() *ifpkg.If_LibLut_t {
	p, _ := ifpkg.Abc_FrameReadLibLut().(*ifpkg.If_LibLut_t)
	return p
}
func speedupDelay(lib *ifpkg.If_LibLut_t, n, pin int32) float32 {
	return ifpkg.If_CutLutDelay(lib, n, pin)
}

func Abc_NtkDelayTraceLut(n *Abc_Ntk_t, useLib int32) float32 {
	if baseabc.Abc_NtkIsLogic(n) == 0 {
		panic("delay trace requires a logic network")
	}
	var lib *ifpkg.If_LibLut_t
	if useLib != 0 {
		lib = speedupLut()
	}
	if lib != nil && lib.LutMax < baseabc.Abc_NtkGetFaninMax(n) {
		fmt.Printf("The max LUT size (%d) is less than the max fanin count (%d).\n", lib.LutMax, baseabc.Abc_NtkGetFaninMax(n))
		return -float32(1e9)
	}
	times := make([]float32, 3*baseabc.Abc_NtkObjNumMax(n))
	for i := 0; i < int(baseabc.Abc_NtkObjNumMax(n)); i++ {
		times[3*i+1] = float32(1e9)
	}
	baseabc.Abc_NtkSetLutTimes(n, times)
	for _, node := range baseabc.Abc_NtkDfs(n, 1) {
		arrival := float32(-float32(1e9))
		count := baseabc.Abc_ObjFaninNum(node)
		var perm [32]int32
		var delays [32]float32
		if lib != nil && lib.VariablePinDelays() {
			Abc_NtkDelayTraceSortPins(node, perm[:], delays[:])
		}
		for k := int32(0); k < count; k++ {
			fanin := baseabc.Abc_ObjFanin(node, k)
			pin := int32(0)
			if lib == nil {
				pin = 0
			} else if lib.VariablePinDelays() {
				pin = k
			}
			candidate := Abc_ObjArrival(fanin) + 1
			if lib != nil && lib.VariablePinDelays() {
				candidate = Abc_ObjArrival(baseabc.Abc_ObjFanin(node, perm[k])) + speedupDelay(lib, count, k)
			} else if lib != nil {
				candidate = Abc_ObjArrival(fanin) + speedupDelay(lib, count, pin)
			}
			if candidate > arrival {
				arrival = candidate
			}
		}
		if count == 0 {
			arrival = 0
		}
		Abc_ObjSetArrival(node, arrival)
	}
	arrival := float32(-float32(1e9))
	for i := int32(0); i < baseabc.Abc_NtkCoNum(n); i++ {
		co := baseabc.Abc_NtkCo(n, i)
		if len(co.Fanins()) == 0 {
			continue
		}
		fanin := baseabc.Abc_ObjFanin0(co)
		if Abc_ObjArrival(fanin) > arrival {
			arrival = Abc_ObjArrival(fanin)
		}
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(n); i++ {
		co := baseabc.Abc_NtkCo(n, i)
		if len(co.Fanins()) != 0 {
			f := baseabc.Abc_ObjFanin0(co)
			if Abc_ObjRequired(f) > arrival {
				Abc_ObjSetRequired(f, arrival)
			}
		}
	}
	for _, node := range baseabc.Abc_NtkDfsReverse(n) {
		count := baseabc.Abc_ObjFaninNum(node)
		var perm [32]int32
		var delays [32]float32
		if lib != nil && lib.VariablePinDelays() {
			Abc_NtkDelayTraceSortPins(node, perm[:], delays[:])
		}
		for k := int32(0); k < count; k++ {
			fanin := baseabc.Abc_ObjFanin(node, k)
			pin := int32(0)
			if lib != nil && lib.VariablePinDelays() {
				pin = k
			}
			d := float32(1)
			if lib != nil {
				d = speedupDelay(lib, count, pin)
			}
			target := Abc_ObjRequired(node) - d
			if lib != nil && lib.VariablePinDelays() {
				fanin = baseabc.Abc_ObjFanin(node, perm[k])
			}
			if Abc_ObjRequired(fanin) > target {
				Abc_ObjSetRequired(fanin, target)
			}
		}
		slack := Abc_ObjRequired(node) - Abc_ObjArrival(node)
		if slack < 0 {
			slack = 0
		}
		Abc_ObjSetSlack(node, slack)
	}
	return arrival
}

func Abc_NtkDelayTracePrint(n *Abc_Ntk_t, useLib, verbose int32) {
	steps := baseabc.Abc_NtkLevel(n)
	if useLib != 0 {
		steps = 20
	}
	if steps <= 0 {
		return
	}
	arrival := Abc_NtkDelayTraceLut(n, useLib)
	delta := arrival / float32(steps)
	if delta <= 0 {
		return
	}
	counts := make([]int, steps+1)
	for _, node := range n.ObjectsOfKind(baseabc.ObjectNode) {
		if baseabc.Abc_ObjFaninNum(node) == 0 {
			continue
		}
		bin := int(Abc_ObjSlack(node) / delta)
		if bin < 0 {
			bin = 0
		}
		if bin > int(steps) {
			bin = int(steps)
		}
		counts[bin]++
	}
	fmt.Printf("Max delay = %6.2f. Delay trace using %s model:\n", arrival, abcSelect(useLib, "LUT library", "unit-delay"))
	nodes := 0
	for i := int32(0); i < steps; i++ {
		nodes += counts[i]
		unit, label := i+1, "lev"
		if useLib != 0 {
			unit, label = 5*(i+1), "%"
		}
		denom := max(1, int(baseabc.Abc_NtkNodeNum(n)))
		fmt.Printf("%3d %s : %5d  (%6.2f %%)\n", unit, label, nodes, 100*float64(nodes)/float64(denom))
	}
}

func Abc_AigCheckTfiRec(node, old *Abc_Obj_t) int32 {
	if node == nil || baseabc.Abc_ObjIsCi(node) != 0 {
		return 0
	}
	if node == old {
		return 1
	}
	if baseabc.Abc_NodeIsTravIdCurrent(node) != 0 {
		return 0
	}
	baseabc.Abc_NodeSetTravIdCurrent(node)
	if Abc_AigCheckTfiRec(baseabc.Abc_ObjFanin0(node), old) != 0 || Abc_AigCheckTfiRec(baseabc.Abc_ObjFanin1(node), old) != 0 {
		return 1
	}
	if choice := baseabc.Abc_ObjChoice(node); choice != nil {
		return Abc_AigCheckTfiRec(choice, old)
	}
	return 0
}
func Abc_AigCheckTfi_rec(node, old *Abc_Obj_t) int32 { return Abc_AigCheckTfiRec(node, old) }
func Abc_AigCheckTfi(newNode, old *Abc_Obj_t) int32 {
	baseabc.Abc_NtkIncrementTravId(newNode.Network())
	return Abc_AigCheckTfiRec(newNode, old)
}
func Abc_NtkSpeedupNodeRec(node *Abc_Obj_t, nodes *[]*Abc_Obj_t) int32 {
	if baseabc.Abc_NodeIsTravIdCurrent(node) != 0 {
		return 1
	}
	if baseabc.Abc_ObjIsCi(node) != 0 {
		return 0
	}
	baseabc.Abc_NodeSetTravIdCurrent(node)
	if Abc_NtkSpeedupNodeRec(baseabc.Abc_ObjFanin0(node), nodes) == 0 || Abc_NtkSpeedupNodeRec(baseabc.Abc_ObjFanin1(node), nodes) == 0 {
		return 0
	}
	*nodes = append(*nodes, node)
	return 1
}
func Abc_NtkSpeedupNode_rec(node *Abc_Obj_t, nodes *Vec_Ptr_t) int32 {
	collected := make([]*Abc_Obj_t, 0)
	result := Abc_NtkSpeedupNodeRec(node, &collected)
	for _, current := range collected {
		Vec_PtrPush(nodes, current)
	}
	return result
}

func speedupMap(node *Abc_Obj_t) baseabc.Edge { return baseabc.Abc_ObjCopyEdge(node) }
func Abc_NtkSpeedupNode(ntk, aigNtk *Abc_Ntk_t, node *Abc_Obj_t, leaves, times *Vec_Ptr_t) {
	leafNodes := make([]*Abc_Obj_t, Vec_PtrSize(leaves))
	timeNodes := make([]*Abc_Obj_t, Vec_PtrSize(times))
	for i := range leafNodes {
		leafNodes[i], _ = Vec_PtrEntry(leaves, int32(i)).(*Abc_Obj_t)
	}
	for i := range timeNodes {
		timeNodes[i], _ = Vec_PtrEntry(times, int32(i)).(*Abc_Obj_t)
	}
	for i, a := range leafNodes {
		for j, b := range leafNodes {
			if i != j && speedupMap(a).Object == speedupMap(b).Object {
				return
			}
		}
	}
	var nodes []*Abc_Obj_t
	baseabc.Abc_NtkIncrementTravId(aigNtk)
	baseabc.Abc_NodeSetTravIdCurrent(aigNtk.One().Object)
	for _, leaf := range leafNodes {
		if mapped := speedupMap(leaf); mapped.Object != nil {
			baseabc.Abc_NodeSetTravIdCurrent(mapped.Object)
		}
	}
	root := speedupMap(node)
	if root.Object == nil || Abc_NtkSpeedupNodeRec(root.Object, &nodes) == 0 {
		return
	}
	if len(timeNodes) > 5 {
		return
	}
	nCofs := 1 << len(timeNodes)
	if nCofs > 32 {
		return
	}
	manager := aigNtk.FunctionManager.(*baseabc.Abc_Aig_t)
	cofs := make([]baseabc.Edge, nCofs)
	for assignment := 0; assignment < nCofs; assignment++ {
		for _, leaf := range leafNodes {
			if mapped := speedupMap(leaf); mapped.Object != nil {
				mapped.Object.SetCopyEdge(mapped.Object.Signal())
			}
		}
		for i, timed := range timeNodes {
			var edge baseabc.Edge
			if assignment&(1<<i) != 0 {
				edge = aigNtk.One()
			} else {
				edge = aigNtk.One().Not()
			}
			if mapped := speedupMap(timed); mapped.Object != nil {
				mapped.Object.SetCopyEdge(edge)
			}
		}
		for _, current := range nodes {
			a, b := baseabc.Abc_ObjChild0(current), baseabc.Abc_ObjChild1(current)
			x := a.Object.ReadCopyEdge()
			y := b.Object.ReadCopyEdge()
			if a.Complement {
				x = x.Not()
			}
			if b.Complement {
				y = y.Not()
			}
			current.SetCopyEdge(baseabc.Abc_AigAnd(manager, x, y))
		}
		cofs[assignment] = root.Object.ReadCopyEdge()
		if root.Complement {
			cofs[assignment] = cofs[assignment].Not()
		}
	}
	for k, timed := range timeNodes {
		skip := 1 << k
		for i := 0; i < nCofs; i += 2 * skip {
			selector := speedupMap(timed)
			cofs[i] = baseabc.Abc_AigMux(manager, selector.Object.Signal(), cofs[i+skip], cofs[i])
		}
	}
	repr, alt := speedupMap(node).Object, cofs[0].Object
	if repr != nil && alt != nil && baseabc.Abc_ObjIsConst1(alt) == 0 && repr.Choice == nil && alt.Choice == nil && repr != alt && Abc_AigCheckTfi(alt, repr) == 0 {
		alt.Choice = repr.Choice
		repr.Choice = alt
	}
}

func Abc_NtkDelayTraceTCEdges(n *Abc_Ntk_t, node *Abc_Obj_t, delta float32, useLib int32) uint32 {
	var lib *ifpkg.If_LibLut_t
	if useLib != 0 {
		lib = speedupLut()
	}
	var perm [32]int32
	var delays [32]float32
	if lib != nil && lib.VariablePinDelays() {
		Abc_NtkDelayTraceSortPins(node, perm[:], delays[:])
	}
	var result uint32
	for k := int32(0); k < baseabc.Abc_ObjFaninNum(node); k++ {
		f := baseabc.Abc_ObjFanin(node, k)
		d := float32(1)
		if lib != nil {
			pin := int32(0)
			if lib.VariablePinDelays() {
				pin = k
				f = baseabc.Abc_ObjFanin(node, perm[k])
			}
			d = speedupDelay(lib, baseabc.Abc_ObjFaninNum(node), pin)
		}
		if Abc_ObjRequired(node) < Abc_ObjArrival(f)+d+delta {
			idx := k
			if lib != nil && lib.VariablePinDelays() {
				idx = perm[k]
			}
			result |= uint32(1) << idx
		}
	}
	return result
}

func speedupBuild(n *Abc_Ntk_t, power bool, useLib, percentage, degree, verbose, veryVerbose int32) *Abc_Ntk_t {
	var probs *Vec_Int_t
	var probability []float32
	var delta float32
	if power {
		probs = Abc_NtkPowerEstimate(n, 0)
		raw := vec.Vec_IntArray(probs)
		probability = float32FromVecIntBits(raw)
		if verbose != 0 {
			Abc_NtkPowerPrint(n, probs)
		}
	} else {
		arrival := Abc_NtkDelayTraceLut(n, useLib)
		if useLib != 0 {
			delta = arrival * float32(percentage) / 100
		} else {
			delta = 1
		}
		if verbose != 0 {
			fmt.Printf("Max delay = %.2f. Delta = %.2f. Using %s model. ", arrival, delta, abcSelect(useLib, "LUT library", "unit-delay"))
			if useLib != 0 {
				fmt.Printf("Percentage = %d. ", percentage)
			}
			fmt.Println()
		}
	}
	critical := make([]uint32, baseabc.Abc_NtkObjNumMax(n))
	counter, counterRes := 0, 0
	for _, node := range n.ObjectsOfKind(baseabc.ObjectNode) {
		if !power && Abc_ObjSlack(node) >= delta {
			continue
		}
		if power {
			critical[node.ID] = Abc_NtkPowerCriticalEdges(n, node, .5-float32(percentage)/100, probs)
		} else {
			critical[node.ID] = Abc_NtkDelayTraceTCEdges(n, node, delta, useLib)
		}
	}
	if !power && verbose != 0 {
		zeroSlack, criticalEdges := 0, 0
		for _, node := range n.ObjectsOfKind(baseabc.ObjectNode) {
			for k := int32(0); k < baseabc.Abc_ObjFaninNum(node); k++ {
				fanin := baseabc.Abc_ObjFanin(node, k)
				if baseabc.Abc_ObjIsCi(fanin) == 0 && Abc_ObjSlack(fanin) < delta {
					zeroSlack++
				}
			}
			criticalEdges += bits.OnesCount32(critical[node.ID])
		}
		total := baseabc.Abc_NtkGetTotalFanins(n)
		fmt.Printf("Edges: Total = %7d. 0-slack = %7d. Critical = %7d. Ratio = %4.2f\n", total, zeroSlack, criticalEdges, float64(criticalEdges)/float64(max(1, zeroSlack)))
	}
	result := Abc_NtkStrash(n, 0, 1, 0)
	if result == nil {
		if probs != nil {
			Vec_IntFree(probs)
		}
		return nil
	}
	cries, fanins := Vec_PtrAlloc(16), Vec_PtrAlloc(16)
	defer Vec_PtrFree(cries)
	defer Vec_PtrFree(fanins)
	for _, node := range n.ObjectsOfKind(baseabc.ObjectNode) {
		marked := critical[node.ID]
		count := 0
		for k := int32(0); k < baseabc.Abc_ObjFaninNum(node); k++ {
			f := baseabc.Abc_ObjFanin(node, k)
			if baseabc.Abc_ObjIsCi(f) == 0 && marked&(uint32(1)<<k) != 0 {
				count++
			}
		}
		if count == 0 && veryVerbose == 0 {
			continue
		}
		counter++
		Vec_PtrClear(cries)
		for k := int32(0); k < baseabc.Abc_ObjFaninNum(node); k++ {
			f := baseabc.Abc_ObjFanin(node, k)
			if baseabc.Abc_ObjIsCi(f) != 0 || marked&(uint32(1)<<k) == 0 {
				continue
			}
			for j := int32(0); j < baseabc.Abc_ObjFaninNum(f); j++ {
				g := baseabc.Abc_ObjFanin(f, j)
				if critical[f.ID]&(uint32(1)<<j) != 0 {
					vec.Vec_PtrPushUnique(cries, g)
				}
			}
		}
		nc := Vec_PtrSize(cries)
		if nc == 0 || nc > degree {
			continue
		}
		counterRes++
		Vec_PtrClear(fanins)
		for k := int32(0); k < baseabc.Abc_ObjFaninNum(node); k++ {
			f := baseabc.Abc_ObjFanin(node, k)
			if baseabc.Abc_ObjIsCi(f) != 0 {
				vec.Vec_PtrPushUnique(fanins, f)
			} else {
				for j := int32(0); j < baseabc.Abc_ObjFaninNum(f); j++ {
					vec.Vec_PtrPushUnique(fanins, baseabc.Abc_ObjFanin(f, j))
				}
			}
		}
		if veryVerbose != 0 {
			fmt.Printf("%5d Node %5d : %d %2d %2d  ", counter, node.ID, count, Vec_PtrSize(cries), Vec_PtrSize(fanins))
			for k := int32(0); k < baseabc.Abc_ObjFaninNum(node); k++ {
				fanin := baseabc.Abc_ObjFanin(node, k)
				value := Abc_ObjSlack(fanin)
				if power {
					value = probability[fanin.ID]
				}
				mark := " "
				if marked&(uint32(1)<<k) != 0 {
					mark = "*"
				}
				fmt.Printf("%d(%.2f)%s ", fanin.ID, value, mark)
			}
			fmt.Println()
		}
		values := make([]*Abc_Obj_t, nc)
		for i := range values {
			values[i], _ = Vec_PtrEntry(cries, int32(i)).(*Abc_Obj_t)
		}
		metric := func(o *Abc_Obj_t) float32 {
			if power {
				return probability[o.ID]
			}
			return Abc_ObjSlack(o)
		}
		for i := 1; i < len(values) && i < 3; i++ {
			for j := i; j > 0 && ((!power && metric(values[j-1]) < metric(values[j])) || (power && metric(values[j-1]) > metric(values[j]))); j-- {
				values[j], values[j-1] = values[j-1], values[j]
			}
		}
		for i, v := range values {
			vec.Vec_PtrWriteEntry(cries, i, v)
		}
		Abc_NtkSpeedupNode(n, result, node, fanins, cries)
	}
	if probs != nil {
		Vec_IntFree(probs)
	}
	if verbose != 0 {
		label := "0-slack"
		if power {
			label = "Power-critical"
		}
		fmt.Printf("Nodes: Total = %7d. %s = %7d. Workable = %7d. Ratio = %4.2f\n", baseabc.Abc_NtkNodeNum(n), label, counter, counterRes, float64(counterRes)/float64(max(1, counter)))
	}
	for _, node := range result.ObjectsOfKind(baseabc.ObjectNode) {
		if node.Choice != nil && baseabc.Abc_ObjFanoutNum(node.Choice) > 0 {
			node.Choice = nil
		}
	}
	return result
}
func Abc_NtkSpeedup(n *Abc_Ntk_t, useLib, percentage, degree, verbose, veryVerbose int32) *Abc_Ntk_t {
	return speedupBuild(n, false, useLib, percentage, degree, verbose, veryVerbose)
}

func Abc_NtkPowerEstimate(n *Abc_Ntk_t, probOne int32) *Vec_Int_t {
	v := vec.Vec_IntStart(int(baseabc.Abc_NtkObjNumMax(n)))
	strashed := Abc_NtkStrash(n, 0, 1, 0)
	if strashed == nil {
		return v
	}
	graph := baseabc.Abc_NtkToDar(strashed, 0, boolToInt32(baseabc.Abc_NtkLatchNum(n) > 0))
	switching := gia.Saig_ManComputeSwitchProbs(graph, 48, 16, probOne)
	values := vec.Vec_IntArray(switching)
	for _, obj := range n.Objects() {
		edge := obj.ReadCopyEdge()
		if edge.Object == nil {
			continue
		}
		mapped := edge.Object.ReadCopyEdge()
		if mapped.Object != nil && mapped.Object.Temp >= 0 && mapped.Object.Temp < len(values) {
			vec.Vec_IntWriteEntry(v, obj.ID, values[mapped.Object.Temp])
		}
	}
	Vec_IntFree(switching)
	aig.Aig_ManStop(graph)
	baseabc.Abc_NtkDelete(strashed)
	return v
}
func Abc_NtkPowerPrint(n *Abc_Ntk_t, probs *Vec_Int_t) {
	data := float32FromVecIntBits(vec.Vec_IntArray(probs))
	var total float64
	buckets := [6]int{}
	powers := [6]float64{}
	nodes, edges := 0, 0
	for _, o := range n.Objects() {
		if baseabc.Abc_ObjIsNode(o) == 0 && baseabc.Abc_ObjIsPi(o) == 0 {
			continue
		}
		nodes++
		edges += int(baseabc.Abc_ObjFanoutNum(o))
		p := data[o.ID]
		total += float64(p) * float64(baseabc.Abc_ObjFanoutNum(o))
		b := int(p * 10)
		if b >= 5 {
			b = 5
		}
		buckets[b]++
		powers[b] += float64(p) * float64(baseabc.Abc_ObjFanoutNum(o))
	}
	fmt.Printf("Node  distribution: ")
	for i := 0; i < 6; i++ {
		fmt.Printf("n%d%d = %6.2f%%  ", i, i+1, 100*float64(buckets[i])/float64(max(1, nodes)))
	}
	fmt.Printf("\nPower distribution: ")
	for i := 0; i < 6; i++ {
		fmt.Printf("p%d%d = %6.2f%%  ", i, i+1, 100*powers[i]/maxFloat(1, total))
	}
	fmt.Printf("\nTotal probs = %7.2f. Total edges = %d. Average = %7.2f.\n", total, edges, total/float64(max(1, edges)))
}
func Abc_NtkPowerCriticalEdges(n *Abc_Ntk_t, node *Abc_Obj_t, limit float32, probs *Vec_Int_t) uint32 {
	data := float32FromVecIntBits(vec.Vec_IntArray(probs))
	var result uint32
	for k := int32(0); k < baseabc.Abc_ObjFaninNum(node); k++ {
		if data[baseabc.Abc_ObjFanin(node, k).ID] >= limit {
			result |= uint32(1) << k
		}
	}
	return result
}
func Abc_NtkPowerdown(n *Abc_Ntk_t, useLib, percentage, degree, verbose, veryVerbose int32) *Abc_Ntk_t {
	return speedupBuild(n, true, useLib, percentage, degree, verbose, veryVerbose)
}

func float32FromVecIntBits(values []int32) []float32 {
	result := make([]float32, len(values))
	for i, value := range values {
		result[i] = math.Float32frombits(uint32(value))
	}
	return result
}
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
