package abci

import (
	"fmt"
	"os"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/map/fpga"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/opt/sim"
)

// fpgaBDDBuilder adapts CUDD's BDD operations to the FPGA mapper's Boolean
// builder interface.
type fpgaBDDBuilder struct{ manager *core.DdManager }

var abcFpgaSimulationRNG = newGlibcRand(1)

func (b fpgaBDDBuilder) ConstantOne() any { return core.Cudd_ReadOne(b.manager) }
func (b fpgaBDDBuilder) Complement(value any) any {
	return core.Cudd_Not(value.(*core.DdNode))
}
func (b fpgaBDDBuilder) Variable(index int) any {
	return core.Cudd_bddIthVar(b.manager, index)
}
func (b fpgaBDDBuilder) And(left, right any) any {
	return core.Cudd_bddAnd(b.manager, left.(*core.DdNode), right.(*core.DdNode))
}

// Abc_NtkFpga runs the FPGA cut mapper on a structurally hashed network and
// returns the mapped BDD network.
func Abc_NtkFpga(ntk *Abc_Ntk_t, delayTarget float32, recovery, switching, latchPaths, verbose int32) *Abc_Ntk_t {
	assert(baseabc.Abc_NtkIsStrash(ntk) != 0, "Abc_NtkIsStrash(ntk)")
	if choices := baseabc.Abc_NtkGetChoiceNum(ntk); choices != 0 {
		fmt.Printf("Performing LUT mapping with %d choices.\n", choices)
	}

	// C initializes fShowSwitching to one, so activity is computed even when
	// the switching-aware mapping option is off.
	showSwitching := int32(1) | switching
	var activity []float32
	if showSwitching != 0 {
		activity = abcFpgaSwitching(ntk)
	}
	manager := Abc_NtkToFpga(ntk, recovery, activity, latchPaths, verbose)
	if manager == nil {
		return nil
	}
	defer fpga.Fpga_ManFree(manager)
	fpga.Fpga_ManSetSwitching(manager, switching)
	fpga.Fpga_ManSetLatchPaths(manager, latchPaths)
	fpga.Fpga_ManSetLatchNum(manager, baseabc.Abc_NtkLatchNum(ntk))
	fpga.Fpga_ManSetDelayTarget(manager, delayTarget)
	if fpga.Fpga_Mapping(manager) == 0 {
		return nil
	}

	mapped := Abc_NtkFromFpga(manager, ntk)
	if mapped == nil {
		return nil
	}
	baseabc.Abc_NtkMinimumBase(mapped)
	if ntk.Exdc != nil {
		mapped.Exdc = baseabc.Abc_NtkDup(ntk.Exdc)
	}
	if baseabc.Abc_NtkCheck(mapped) == 0 {
		fmt.Println("Abc_NtkFpga: The network check has failed.")
		baseabc.Abc_NtkDelete(mapped)
		return nil
	}
	return mapped
}

func abcFpgaSwitching(ntk *Abc_Ntk_t) []float32 {
	const patterns, words = 4096, 4096 / 32
	values := make([]float32, int(baseabc.Abc_NtkObjNumMax(ntk)))
	info := make([][]uint32, len(values))
	for i := range info {
		info[i] = make([]uint32, words)
	}
	for i := int32(0); i < baseabc.Abc_NtkCiNum(ntk); i++ {
		node := baseabc.Abc_NtkCi(ntk, i)
		abcFpgaSetRandom(info[node.ID], abcFpgaSimulationRNG)
	}
	for _, node := range baseabc.Abc_AigDfs(ntk, 1, 0) {
		if node.Kind != baseabc.ObjectNode || len(node.Fanins()) != 2 {
			continue
		}
		fanins := node.Fanins()
		left, right := info[fanins[0].Object.ID], info[fanins[1].Object.ID]
		for w := 0; w < words; w++ {
			a, b := left[w], right[w]
			if fanins[0].Complement {
				a = ^a
			}
			if fanins[1].Complement {
				b = ^b
			}
			info[node.ID][w] = a & b
		}
	}
	for _, node := range ntk.Objects() {
		if node.ID >= len(values) || (node.Kind != baseabc.ObjectPI && node.Kind != baseabc.ObjectBO && node.Kind != baseabc.ObjectNode) {
			continue
		}
		ones := sim.Sim_UtilCountOnes(info[node.ID], words)
		total := float32(patterns)
		values[node.ID] = 2 * float32(ones) / total * (total - float32(ones)) / total
	}
	return values
}

// abcFpgaSetRandom matches SIM_RANDOM_UNSIGNED: three glibc rand() values
// shifted and XORed for each simulation word.
func abcFpgaSetRandom(words []uint32, rng *glibcRand) {
	for i := range words {
		words[i] = (rng.next() << 24) ^ (rng.next() << 12) ^ rng.next()
	}
}

// Abc_NtkToFpga imports the source AIG and its choice chains into an FPGA
// mapping manager. Activity values, when supplied, are indexed by ABC object ID.
func Abc_NtkToFpga(ntk *Abc_Ntk_t, recovery int32, switching []float32, latchPaths, verbose int32) *fpga.Fpga_Man_t {
	assert(baseabc.Abc_NtkIsStrash(ntk) != 0, "Abc_NtkIsStrash(ntk)")
	manager := fpga.Fpga_ManCreate(baseabc.Abc_NtkCiNum(ntk), baseabc.Abc_NtkCoNum(ntk), verbose)
	if manager == nil {
		return nil
	}
	fpga.Fpga_ManSetAreaRecovery(manager, recovery)
	fpga.Fpga_ManSetOutputNames(manager, baseabc.Abc_NtkCollectCioNames(ntk, 1))
	arrivalPtr := Abc_NtkGetCiArrivalFloats(ntk)
	ciCount := int(baseabc.Abc_NtkCiNum(ntk))
	var arrivals []float32
	if ciCount != 0 && arrivalPtr != nil {
		arrivals = arrivalPtr
	}
	if latchPaths != 0 {
		for i := 0; i < int(baseabc.Abc_NtkPiNum(ntk)) && i < len(arrivals); i++ {
			arrivals[i] = -fpga.FPGA_FLOAT_LARGE
		}
	}
	fpga.Fpga_ManSetInputArrivals(manager, arrivals)

	baseabc.Abc_NtkCleanCopy(ntk)
	baseabc.Abc_ObjSetCopyData(baseabc.Abc_AigConst1(ntk).Object, fpga.Fpga_ManReadConst1(manager))
	for i := int32(0); i < baseabc.Abc_NtkCiNum(ntk); i++ {
		node := baseabc.Abc_NtkCi(ntk, i)
		fpgaNode := fpga.Fpga_ManReadInputs(manager)[i]
		baseabc.Abc_ObjSetCopyData(node, fpgaNode)
		if switching != nil && node.ID < len(switching) {
			fpga.Fpga_NodeSetSwitching(fpgaNode, switching[node.ID])
		}
	}

	nodes := baseabc.Abc_AigDfs(ntk, 0, 0)
	progress := extra.Extra_ProgressBarStart(os.Stdout, int32(len(nodes)))
	for i, node := range nodes {
		extra.Extra_ProgressBarUpdate(progress, int32(i), "")
		fanin0, ok0 := baseabc.Abc_ObjCopyData(baseabc.Abc_ObjFanin0(node)).(*fpga.Fpga_Node_t)
		fanin1, ok1 := baseabc.Abc_ObjCopyData(baseabc.Abc_ObjFanin1(node)).(*fpga.Fpga_Node_t)
		assert(ok0 && ok1, "FPGA fanins are mapped in DFS order")
		fpgaNode := fpga.Fpga_NodeAnd(manager,
			fpga.Fpga_NotCond(fanin0, baseabc.Abc_ObjFaninC0(node)),
			fpga.Fpga_NotCond(fanin1, baseabc.Abc_ObjFaninC1(node)))
		assert(baseabc.Abc_ObjCopyData(node) == nil, "node copy data is initially clear")
		baseabc.Abc_ObjSetCopyData(node, fpgaNode)
		if switching != nil && node.ID < len(switching) {
			fpga.Fpga_NodeSetSwitching(fpgaNode, switching[node.ID])
		}
		if baseabc.Abc_AigNodeIsChoice(node) != 0 {
			previous := node
			for choice := node.Choice; choice != nil; choice = choice.Choice {
				previousNode, okPrev := baseabc.Abc_ObjCopyData(previous).(*fpga.Fpga_Node_t)
				choiceNode, okChoice := baseabc.Abc_ObjCopyData(choice).(*fpga.Fpga_Node_t)
				assert(okPrev && okChoice, "choice alternatives are mapped in DFS order")
				fpga.Fpga_NodeSetNextE(previousNode, choiceNode)
				fpga.Fpga_NodeSetRepr(choiceNode, fpgaNode)
				previous = choice
			}
		}
	}
	extra.Extra_ProgressBarStop(progress)
	progress = extra.Extra_ProgressBarStart(os.Stdout, baseabc.Abc_NtkCoNum(ntk))
	for i := int32(0); i < baseabc.Abc_NtkCoNum(ntk); i++ {
		extra.Extra_ProgressBarUpdate(progress, i, "")
		co := baseabc.Abc_NtkCo(ntk, i)
		fanin := baseabc.Abc_ObjFanin0(co)
		mapped, ok := baseabc.Abc_ObjCopyData(fanin).(*fpga.Fpga_Node_t)
		assert(ok, "CO driver has an FPGA mapping")
		fpga.Fpga_ManReadOutputs(manager)[i] = mapped
	}
	return manager
}

// Abc_NtkFromFpga converts mapped cuts into nodes of a BDD logic network.
func Abc_NtkFromFpga(manager *fpga.Fpga_Man_t, ntk *Abc_Ntk_t) *Abc_Ntk_t {
	mapped := baseabc.Abc_NtkStartFrom(ntk, baseabc.ABC_NTK_LOGIC, baseabc.ABC_FUNC_BDD)
	fpga.Fpga_CutsCleanSign(manager)
	fpga.Fpga_ManCleanData0(manager)
	for i := int32(0); i < baseabc.Abc_NtkCiNum(ntk); i++ {
		source := baseabc.Abc_NtkCi(ntk, i)
		fpga.Fpga_NodeSetData0(fpga.Fpga_ManReadInputs(manager)[i], source.Copy)
	}
	constant := baseabc.Abc_NtkCreateNodeConst1(mapped)
	fpga.Fpga_NodeSetData0(fpga.Fpga_ManReadConst1(manager), constant)
	progress := extra.Extra_ProgressBarStart(os.Stdout, baseabc.Abc_NtkCoNum(ntk))
	for i := int32(0); i < baseabc.Abc_NtkCoNum(ntk); i++ {
		co := baseabc.Abc_NtkCo(ntk, i)
		newDriver := Abc_NodeFromFpga_rec(mapped, fpga.Fpga_ManReadOutputs(manager)[i])
		assert(newDriver != nil, "mapped output node exists")
		assert(baseabc.Abc_ObjIsComplement(newDriver) == 0, "mapped output is regular")
		baseabc.Abc_ObjFanin0(co).Copy = newDriver
	}
	extra.Extra_ProgressBarStop(progress)
	baseabc.Abc_NtkFinalize(ntk, mapped)
	if baseabc.Abc_ObjFanoutNum(constant) == 0 {
		baseabc.Abc_NtkDeleteObj(constant)
	}
	dupGates := baseabc.Abc_NtkLogicMakeSimpleCos(mapped, 1)
	if dupGates != 0 && fpga.Fpga_ManReadVerbose(manager) != 0 {
		fmt.Printf("Duplicated %d gates to decouple the CO drivers.\n", dupGates)
	}
	return mapped
}

func Abc_NodeFromFpga_rec(ntk *Abc_Ntk_t, node *fpga.Fpga_Node_t) *Abc_Obj_t {
	assert(fpga.Fpga_IsComplement(node) == 0, "FPGA cut node is regular")
	if known, ok := fpga.Fpga_NodeReadData0(node).(*Abc_Obj_t); ok && known != nil {
		return known
	}
	assert(fpga.Fpga_NodeIsAnd(node) != 0, "nonterminal FPGA node is an AND")
	cut := fpga.Fpga_NodeReadCutBest(node)
	leaves := fpga.Fpga_CutReadLeaves(cut)
	newNode := baseabc.Abc_NtkCreateNode(ntk)
	for _, leaf := range leaves {
		baseabc.Abc_ObjAddFanin(newNode, Abc_NodeFromFpga_rec(ntk, leaf))
	}
	manager, ok := ntk.FunctionManager.(*core.DdManager)
	assert(ok, "mapped network has a CUDD manager")
	function := fpga.Fpga_TruthsCutBdd(fpgaBDDBuilder{manager}, cut).(*core.DdNode)
	core.Cudd_Ref(function)
	baseabc.Abc_ObjSetData(newNode, function)
	fpga.Fpga_NodeSetData0(node, newNode)
	return newNode
}
