package abc

import (
	"fmt"
	"strconv"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// latchPredecessor follows a BO -> latch -> BI chain without entering logic.
func latchPredecessor(latch *Object) *Object {
	if latch == nil || latch.Kind != ObjectLatch || len(latch.fanins) != 1 {
		return nil
	}
	input := latch.fanins[0].Object
	if input.Kind != ObjectBI || len(input.fanins) != 1 {
		return nil
	}
	output := input.fanins[0].Object
	if output.Kind != ObjectBO || len(output.fanins) != 1 {
		return nil
	}
	predecessor := output.fanins[0].Object
	if predecessor.Kind != ObjectLatch {
		return nil
	}
	return predecessor
}

func Abc_NtkLatchIsSelfFeed_rec(latch, root *Object) int32 {
	seen := make(map[*Object]bool)
	for latch != nil && !seen[latch] {
		if latch == root {
			return 1
		}
		seen[latch] = true
		latch = latchPredecessor(latch)
	}
	return 0
}

func Abc_NtkLatchIsSelfFeed(latch *Object) int32 {
	if latch == nil || latch.Kind != ObjectLatch {
		panic("expected latch")
	}
	return Abc_NtkLatchIsSelfFeed_rec(latchPredecessor(latch), latch)
}

func Abc_NtkCountSelfFeedLatches(n *Network) int32 {
	var count int32
	for _, latch := range n.ObjectsOfKind(ObjectLatch) {
		count += Abc_NtkLatchIsSelfFeed(latch)
	}
	return count
}

func Abc_NtkRemoveSelfFeedLatches(n *Network) int32 {
	var count int32
	for _, latch := range n.ObjectsOfKind(ObjectLatch) {
		if Abc_NtkLatchIsSelfFeed(latch) == 0 {
			continue
		}
		input := latch.fanins[0].Object
		constant := n.One()
		if Abc_NtkIsStrash(n) == 0 {
			constant = Abc_NtkCreateNodeConst1(n).Signal()
		}
		constant.Complement = input.fanins[0].Complement
		if err := n.PatchFanin(input, 0, constant); err != nil {
			panic(err)
		}
		count++
	}
	return count
}

func Abc_NtkAddLatch(n *Network, driver *Object, init Abc_InitType_t) *Object {
	output := n.NewObject(ObjectBO)
	latch := n.NewObject(ObjectLatch)
	input := n.NewObject(ObjectBI)
	output.SetName(latch.Name() + "_lo")
	input.SetName(latch.Name() + "_li")
	Abc_ObjAddFanin(output, latch)
	Abc_ObjAddFanin(latch, input)
	if driver != nil {
		Abc_ObjAddFanin(input, driver)
	}
	setLatchInit(latch, init)
	return output
}

func Abc_NtkLatchPipe(n *Network, stages int32) {
	if stages <= 0 {
		return
	}
	total := int(stages) * len(n.ObjectsOfKind(ObjectPI))
	width := len(strconv.Itoa(total))
	for i, pi := range n.ObjectsOfKind(ObjectPI) {
		fanouts := pi.Fanouts()
		driver := pi
		for stage := 0; stage < int(stages); stage++ {
			driver = Abc_NtkAddLatch(n, driver, ABC_INIT_DC)
			latch := driver.fanins[0].Object
			latch.SetName(fmt.Sprintf("LL%0*d", width, i*int(stages)+stage))
		}
		for _, fanout := range fanouts {
			Abc_ObjPatchFanin(fanout, pi, driver)
		}
	}
	Abc_NtkOrderCisCos(n)
}

func Abc_NtkCollectLatchValues(n *Network) *vec.Vec_Int_t {
	latches := n.ObjectsOfKind(ObjectLatch)
	values := vec.Vec_IntAlloc(len(latches))
	for _, latch := range latches {
		vec.Vec_IntPush(values, Abc_LatchIsInit1(latch))
	}
	return values
}

func Abc_NtkCollectLatchValuesStr(n *Network) string {
	latches := n.ObjectsOfKind(ObjectLatch)
	values := make([]byte, len(latches))
	for i, latch := range latches {
		switch {
		case Abc_LatchIsInit0(latch) != 0:
			values[i] = '0'
		case Abc_LatchIsInit1(latch) != 0:
			values[i] = '1'
		case Abc_LatchIsInitDc(latch) != 0:
			values[i] = 'x'
		default:
			panic("latch initialization is not 0, 1, or don't-care")
		}
	}
	return string(values)
}

func Abc_NtkInsertLatchValues(n *Network, values *vec.Vec_Int_t) {
	latches := n.ObjectsOfKind(ObjectLatch)
	if values != nil && vec.Vec_IntSize(values) != len(latches) {
		panic("latch initialization count mismatch")
	}
	for i, latch := range latches {
		init := ABC_INIT_DC
		if values != nil {
			init = ABC_INIT_ZERO
			if vec.Vec_IntEntry(values, i) != 0 {
				init = ABC_INIT_ONE
			}
		}
		setLatchInit(latch, init)
	}
}

func Abc_NtkNodeConvertToMux(n *Network, control, whenTrue, whenFalse, mux *Object) {
	if Abc_NtkIsLogic(n) == 0 || len(mux.fanins) != 0 {
		panic("expected empty logic node")
	}
	for _, input := range []*Object{control, whenTrue, whenFalse} {
		Abc_ObjAddFanin(mux, input)
	}
	switch Abc_NtkFunc_t(n.Function) {
	case ABC_FUNC_SOP:
		mux.Data = "11- 1\n0-1 1\n"
	case ABC_FUNC_BDD:
		dd := n.FunctionManager.(*core.DdManager)
		mux.Data = core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1), core.Cudd_bddIthVar(dd, 2))
		core.Cudd_Ref(mux.Data.(*core.DdNode))
	case ABC_FUNC_AIG:
		manager := n.FunctionManager.(*hop.Hop_Man_t)
		mux.Data = hop.Hop_Mux(manager, hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1), hop.Hop_IthVar(manager, 2))
	default:
		panic("unsupported logic function representation")
	}
}

func Abc_NtkConvertDcLatches(n *Network) {
	var latches []*Object
	for _, latch := range n.ObjectsOfKind(ObjectLatch) {
		if Abc_LatchIsInitDc(latch) != 0 {
			latches = append(latches, latch)
		}
	}
	if len(latches) == 0 {
		return
	}
	control := Abc_NtkAddLatch(n, Abc_NtkCreateNodeConst1(n), ABC_INIT_ZERO)
	for _, latch := range latches {
		setLatchInit(latch, ABC_INIT_ZERO)
		output := latch.fanouts[0]
		for _, fanout := range output.fanouts {
			if fanout.IsCO() {
				output.SetName(latch.Name() + "_lo")
				break
			}
		}
		input := n.NewObject(ObjectPI)
		input.SetName(latch.Name() + "_pi")
		mux := n.NewObject(ObjectNode)
		Abc_ObjTransferFanout(output, mux)
		Abc_NtkNodeConvertToMux(n, control, output, input, mux)
	}
	Abc_NtkOrderCisCos(n)
	fmt.Printf("The number of converted latches with DC values = %d.\n", len(latches))
}

// Abc_NtkConvertOnehot replaces a binary register bank with one latch per state.
func Abc_NtkConvertOnehot(n *Network) *Network {
	if Abc_NtkIsLogic(n) == 0 {
		panic("one-hot conversion requires a logic network")
	}
	latches := n.ObjectsOfKind(ObjectLatch)
	if len(latches) == 0 {
		return Abc_NtkDup(n)
	}
	if len(latches) > 16 {
		fmt.Printf("Cannot reencode %d flops because it will lead to 2^%d states.\n", len(latches), len(latches))
		return nil
	}
	initial := 0
	for i, latch := range latches {
		if Abc_LatchIsInitDc(latch) != 0 {
			fmt.Printf("Cannot process logic network with don't-care init values. Run \"zero\".\n")
			return nil
		}
		if Abc_LatchIsInit1(latch) != 0 {
			initial |= 1 << i
		}
	}
	if Abc_NtkToSop(n, -1, 1<<30) == 0 {
		return nil
	}
	result := Abc_NtkStartFromNoLatches(n, ABC_NTK_LOGIC, ABC_FUNC_SOP)
	states := make([]*Object, 1<<len(latches))
	for state := range states {
		init := ABC_INIT_ZERO
		if state == initial {
			init = ABC_INIT_ONE
		}
		states[state] = Abc_NtkAddLatch(result, nil, init)
		states[state].fanins[0].Object.SetName(fmt.Sprintf("latch%d", state))
	}
	signals := map[*Object]Edge{n.constant: result.One()}
	for _, pi := range n.ObjectsOfKind(ObjectPI) {
		signals[pi] = getCopy(pi)
	}
	for bit, latch := range latches {
		decoder := result.NewObject(ObjectNode)
		for state, output := range states {
			if state&(1<<bit) != 0 {
				Abc_ObjAddFanin(decoder, output)
			}
		}
		decoder.Data = Abc_SopCreateOr(nil, int32(len(decoder.fanins)), nil)
		signals[latch.fanouts[0]] = decoder.Signal()
		setCopy(latch.fanouts[0], decoder.Signal())
	}
	order, err := n.Topological(n.cos)
	if err != nil {
		panic(err)
	}
	for _, object := range order {
		if _, ok := signals[object]; ok {
			continue
		}
		if object.IsCO() {
			edge := object.fanins[0]
			signal := signals[edge.Object]
			if edge.Complement {
				signal = signal.Not()
			}
			signals[object] = signal
			continue
		}
		copy := Abc_NtkDupObj(result, object, 1)
		for _, edge := range object.fanins {
			signal, ok := signals[edge.Object]
			if !ok {
				panic("missing one-hot conversion fanin")
			}
			if edge.Complement {
				signal = signal.Not()
			}
			Abc_ObjAddFanin(copy, signal)
		}
		signals[object] = copy.Signal()
	}
	for _, po := range n.ObjectsOfKind(ObjectPO) {
		Abc_ObjAddFanin(po.Copy, signals[po])
	}
	for state, output := range states {
		decoder := result.NewObject(ObjectNode)
		complement := make([]int32, len(latches))
		for bit, latch := range latches {
			Abc_ObjAddFanin(decoder, signals[latch.fanins[0].Object])
			if state&(1<<bit) == 0 {
				complement[bit] = 1
			}
		}
		decoder.Data = Abc_SopCreateAnd(nil, int32(len(latches)), complement)
		Abc_ObjAddFanin(output.fanins[0].Object.fanins[0].Object, decoder)
	}
	Abc_NtkOrderCisCos(result)
	if err := result.Check(); err != nil {
		panic(err)
	}
	return result
}

// Native one-hot groups already contain latch indices rather than name pointers.
func Abc_NtkConverLatchNamesIntoNumbers(n *Network) *vec.Vec_Ptr_t {
	if n.Onehots == nil {
		return nil
	}
	result := vec.Vec_PtrAlloc(len(n.Onehots))
	for _, group := range n.Onehots {
		numbers := vec.Vec_IntAlloc(len(group))
		for _, index := range group {
			if index >= 0 && int(index) < len(n.ObjectsOfKind(ObjectLatch)) {
				vec.Vec_IntPush(numbers, index)
			}
		}
		if vec.Vec_IntSize(numbers) > 1 {
			vec.Vec_PtrPush(result, numbers)
		}
	}
	return result
}

func Abc_NtkTransformBack(old, result *Network, controls *vec.Vec_Ptr_t, classes *vec.Vec_Int_t) {
	oldOutputs, newOutputs := old.ObjectsOfKind(ObjectPO), result.ObjectsOfKind(ObjectPO)
	if len(oldOutputs) != len(newOutputs) {
		panic("retiming changed primary output count")
	}
	for i, output := range oldOutputs {
		output.Copy = newOutputs[i]
	}
	for i := 0; i < vec.Vec_PtrSize(controls); i++ {
		output := vec.Vec_PtrEntry(controls, i).(*Object)
		if output.network != old || output.Kind != ObjectPO {
			panic("control is not an original primary output")
		}
		vec.Vec_PtrWriteEntry(controls, i, output.Copy)
	}
	latches := result.ObjectsOfKind(ObjectLatch)
	if len(latches) != vec.Vec_IntSize(classes) {
		panic("latch class count mismatch")
	}
	for i, latch := range latches {
		class := vec.Vec_IntEntry(classes, i)
		if class == -1 {
			continue
		}
		if class < 0 || int(class) >= vec.Vec_PtrSize(controls) {
			panic("invalid latch class")
		}
		input := latch.fanins[0].Object
		control := vec.Vec_PtrEntry(controls, int(class)).(*Object).fanins[0]
		mux := result.NewObject(ObjectNode)
		Abc_ObjAddFanin(mux, control)
		Abc_ObjAddFanin(mux, input.fanins[0])
		Abc_ObjAddFanin(mux, latch.fanouts[0])
		mux.Data = "0-1 1\n11- 1\n"
		if err := result.PatchFanin(input, 0, mux.Signal()); err != nil {
			panic(err)
		}
	}
	for i := 0; i < vec.Vec_PtrSize(controls); i++ {
		Abc_NtkDeleteObj(vec.Vec_PtrEntry(controls, i).(*Object))
	}
}

func Abc_NtkVerifyCex(n *Network, trace *util.Abc_Cex_t) int32 {
	pis, latches := n.ObjectsOfKind(ObjectPI), n.ObjectsOfKind(ObjectLatch)
	if trace == nil || trace.IFrame < 0 || int(trace.NPis) != len(pis) || int(trace.NRegs) < len(latches) || trace.NBits != trace.NRegs+(trace.IFrame+1)*trace.NPis || int64(len(trace.PData))*32 < int64(trace.NBits) {
		panic("invalid network counterexample dimensions")
	}
	inputs := make(map[*Object]bool, len(pis)+len(latches))
	for i, latch := range latches {
		inputs[latch.fanouts[0]] = util.Abc_InfoHasBit(trace.PData, int32(i))
	}
	outputs := n.ObjectsOfKind(ObjectPO)
	failed := int32(-1)
	for frame := int32(0); frame <= trace.IFrame; frame++ {
		for i, pi := range pis {
			inputs[pi] = util.Abc_InfoHasBit(trace.PData, trace.NRegs+frame*trace.NPis+int32(i))
		}
		if frame == trace.IFrame {
			for i, output := range outputs {
				value, err := n.Evaluate(output.Signal(), inputs)
				if err != nil {
					panic(err)
				}
				if value {
					failed = int32(i)
					break
				}
			}
		}
		next := make([]bool, len(latches))
		for i, latch := range latches {
			var err error
			next[i], err = n.Evaluate(latch.fanins[0], inputs)
			if err != nil {
				panic(err)
			}
		}
		for i, latch := range latches {
			inputs[latch.fanouts[0]] = next[i]
		}
	}
	return failed
}
