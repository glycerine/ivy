package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/hop"
	baseabc "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dar"
)

const TIM_TEST_BOX_RATIO = 200

type Vec_Flt_t = vec.Vec_Flt_t
type Tim_Man_t = tim.Tim_Man_t

func Abc_NodeIsWhiteBox(obj *Abc_Obj_t) int32 {
	if obj == nil || baseabc.Abc_ObjIsNode(obj) == 0 {
		return 0
	}
	n := len(obj.Fanins())
	return bool32(obj.ID%TIM_TEST_BOX_RATIO == 0 && n > 0 && n < 10)
}

func Abc_NtkTestTimNodeStrash_rec(g *Gia_Man_t, obj *Hop_Obj_t) {
	if obj == nil || hop.Hop_IsComplement(obj) != 0 {
		panic("TIM local AIG recursion expects a regular node")
	}
	if hop.Hop_ObjIsNode(obj) == 0 || hop.Hop_ObjIsMarkA(obj) != 0 {
		return
	}
	Abc_NtkTestTimNodeStrash_rec(g, hop.Hop_ObjFanin0(obj))
	Abc_NtkTestTimNodeStrash_rec(g, hop.Hop_ObjFanin1(obj))
	a, b := hop.Hop_ObjChild0CopyI(obj), hop.Hop_ObjChild1CopyI(obj)
	hop.Hop_ObjSetIData(obj, gia.Gia_ManHashAnd(g, a, b))
	if hop.Hop_ObjIsMarkA(obj) != 0 {
		panic("loop in local Hop AIG")
	}
	hop.Hop_ObjSetMarkA(obj)
}

func Abc_NtkTestTimNodeStrash(g *Gia_Man_t, node *Abc_Obj_t) int32 {
	if node == nil || baseabc.Abc_ObjIsNode(node) == 0 || baseabc.Abc_NtkIsAigLogic(node.Network()) == 0 {
		panic("Abc_NtkTestTimNodeStrash requires an AIG logic node")
	}
	man, ok := node.Network().FunctionManager.(*hop.Hop_Man_t)
	if !ok || man == nil {
		panic("AIG logic network has no Hop manager")
	}
	root, ok := node.Data.(*hop.Hop_Obj_t)
	if !ok || root == nil {
		panic("AIG logic node has no local Hop function")
	}
	regular := hop.Hop_Regular(root)
	if len(node.Fanins()) == 0 || regular == hop.Hop_ManConst1(man) {
		if hop.Hop_IsComplement(root) != 0 {
			return 0
		}
		return 1
	}
	for i, e := range node.Fanins() {
		if e.Object == nil {
			panic("nil local-function fanin")
		}
		hop.Hop_ObjSetIData(hop.Hop_IthVar(man, int32(i)), int32(e.Object.Temp))
	}
	Abc_NtkTestTimNodeStrash_rec(g, regular)
	hop.Hop_ConeUnmark_rec(regular)
	lit := hop.Hop_ObjIData(regular)
	if hop.Hop_IsComplement(root) != 0 {
		lit ^= 1
	}
	return lit
}

func Abc_NtkTestTimCollectCone_rec(obj *Abc_Obj_t, nodes *Vec_Ptr_t) {
	if obj == nil || baseabc.Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(obj)
	if obj.IsCI() {
		return
	}
	if obj.Kind != baseabc.ObjectNode {
		panic("TIM cone contains a non-node object")
	}
	for _, e := range obj.Fanins() {
		Abc_NtkTestTimCollectCone_rec(e.Object, nodes)
	}
	vec.Vec_PtrPush(nodes, obj)
}
func Abc_NtkTestTimCollectCone(n *Abc_Ntk_t, obj *Abc_Obj_t) *Vec_Ptr_t {
	cone := vec.Vec_PtrAlloc(1000)
	if obj != nil {
		if obj.Kind != baseabc.ObjectNode || baseabc.Abc_NodeIsTravIdCurrent(obj) != 0 {
			panic("invalid or already visited TIM cone root")
		}
		Abc_NtkTestTimCollectCone_rec(obj, cone)
		vec.Vec_PtrPop(cone)
		return cone
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(n); i++ {
		co := baseabc.Abc_NtkCo(n, i)
		if len(co.Fanins()) > 0 {
			Abc_NtkTestTimCollectCone_rec(co.Fanins()[0].Object, cone)
		}
	}
	return cone
}

func Abc_NtkTestCreateArrivals(n int32) *Vec_Flt_t {
	v := vec.Vec_FltAlloc(int(n))
	for i := int32(0); i < n; i++ {
		vec.Vec_FltPush(v, float32(i%10))
	}
	return v
}
func Abc_NtkTestCreateRequired(n int32) *Vec_Flt_t {
	v := vec.Vec_FltAlloc(int(n))
	for i := int32(0); i < n; i++ {
		vec.Vec_FltPush(v, 100+float32(i))
	}
	return v
}

type timBoxLibrary struct{ lib *ifpkg.If_LibBox_t }

func (l timBoxLibrary) BoxCount() int {
	if l.lib == nil {
		return 0
	}
	count := 0
	for _, b := range l.lib.Boxes() {
		if int(b.ID())+1 > count {
			count = int(b.ID()) + 1
		}
	}
	return count
}
func (l timBoxLibrary) BoxTiming(id int) (int, int, bool, []float32) {
	if l.lib == nil {
		return 0, 0, false, nil
	}
	b := ifpkg.If_LibBoxReadBox(l.lib, int32(id))
	if b == nil {
		return 0, 0, false, nil
	}
	d := make([]float32, int(b.Inputs()*b.Outputs()))
	for o := int32(0); o < b.Outputs(); o++ {
		for i := int32(0); i < b.Inputs(); i++ {
			d[int(o*b.Inputs()+i)] = float32(b.Delay(o, i))
		}
	}
	return int(b.Inputs()), int(b.Outputs()), b.BlackBox(), d
}

func Abc_NtkTestTimDeriveGia(n *Abc_Ntk_t, verbose int32) *Gia_Man_t {
	if n == nil || baseabc.Abc_NtkIsTopo(n) == 0 {
		panic("TIM derivation requires a topological network")
	}
	baseabc.Abc_NtkFillTemp(n)
	ciNum, coNum := baseabc.Abc_NtkCiNum(n), baseabc.Abc_NtkCoNum(n)
	nodes := n.ObjectsOfKind(baseabc.ObjectNode)
	white := make(map[*Abc_Obj_t]bool)
	curPi, curPo, maxInputs := ciNum, coNum, 0
	for _, o := range nodes {
		white[o] = Abc_NodeIsWhiteBox(o) != 0
		if !white[o] {
			continue
		}
		ins := int32(len(o.Fanins()))
		if int(ins) > maxInputs {
			maxInputs = int(ins)
		}
		curPi++
		curPo += ins
		if verbose != 0 {
			fmt.Printf("Selecting node %6d as white boxes with %d inputs and %d output.\n", o.ID, ins, 1)
		}
	}
	g := gia.Gia_ManStart(baseabc.Abc_NtkObjNumMax(n))
	holes := gia.Gia_ManStart(1000)
	for i := int32(0); i < curPi; i++ {
		gia.Gia_ManAppendCi(g)
	}
	for i := 0; i < int32(maxInputs); i++ {
		gia.Gia_ManAppendCi(holes)
	}
	gia.Gia_ManHashAlloc(g)
	gia.Gia_ManHashAlloc(holes)
	manager := tim.Tim_ManStart(int(curPi), int(curPo))
	if c := n.Object(0); c != nil && c.Kind == baseabc.ObjectConst1 {
		c.Temp = 1
	}
	for i := int32(0); i < ciNum; i++ {
		baseabc.Abc_NtkCi(n, i).Temp = gia.Gia_ObjToLit(g, gia.Gia_ManCi(g, i))
	}
	var coLits, holeLits []int32
	curPi, curPo = ciNum, 0
	for _, o := range nodes {
		if !white[o] {
			o.Temp = Abc_NtkTestTimNodeStrash(g, o)
			continue
		}
		ins := int32(len(o.Fanins()))
		boxID := ins
		tim.Tim_ManCreateBox(manager, int(curPo), int(ins), int(curPi), 1, int(boxID), 0)
		curPo += ins
		for k, e := range o.Fanins() {
			coLits = append(coLits, e.Object.Temp)
			e.Object.Temp = gia.Gia_ObjToLit(holes, gia.Gia_ManCi(holes, int32(k)))
			if k >= maxInputs {
				panic("whitebox exceeds hole input capacity")
			}
		}
		holeLits = append(holeLits, Abc_NtkTestTimNodeStrash(holes, o))
		o.Temp = gia.Gia_ObjToLit(g, gia.Gia_ManCi(g, curPi))
		curPi++
	}
	for i := int32(0); i < coNum; i++ {
		co := baseabc.Abc_NtkCo(n, i)
		if len(co.Fanins()) == 0 {
			panic("TIM CO has no driver")
		}
		gia.Gia_ManAppendCo(g, co.Fanins()[0].Object.Temp)
	}
	for _, lit := range coLits {
		gia.Gia_ManAppendCo(g, lit)
	}
	for _, lit := range holeLits {
		gia.Gia_ManAppendCo(holes, lit)
	}
	if curPi != gia.Gia_ManPiNum(g) || curPo+baseabc.Abc_NtkPoNum(n) != gia.Gia_ManPoNum(g) {
		panic("TIM GIA terminal count mismatch")
	}
	gia.Gia_ManHashStop(g)
	gia.Gia_ManSetRegNum(g, 0)
	gia.Gia_ManHashStop(holes)
	gia.Gia_ManSetRegNum(holes, 0)
	g0, h0 := g, holes
	g = gia.Gia_ManCleanup(g0)
	gia.Gia_ManStop(g0)
	holes = gia.Gia_ManCleanup(h0)
	gia.Gia_ManStop(h0)
	arr := Abc_NtkTestCreateArrivals(baseabc.Abc_NtkPiNum(n))
	req := Abc_NtkTestCreateRequired(baseabc.Abc_NtkPoNum(n))
	lib, _ := abcmain.Abc_FrameReadLibBox().(*ifpkg.If_LibBox_t)
	tim.Tim_ManPrint(manager)
	tim.Tim_ManCreate(manager, timBoxLibrary{lib}, vec.Vec_FltArray(arr), vec.Vec_FltArray(req))
	tim.Tim_ManPrint(manager)
	vec.Vec_FltFree(arr)
	vec.Vec_FltFree(req)
	if err := gia.Gia_AigerWrite(holes, "holes00.aig", 0, 0, 0); err != nil {
		fmt.Printf("Could not write holes00.aig: %v\n", err)
	}
	gia.Gia_ManSetBoxData(g, holes, manager)
	return g
}

func Abc_NtkTestTimPerformSynthesis(p *Gia_Man_t, choices int32) *Gia_Man_t {
	if p == nil {
		return nil
	}
	graph := gia.Gia_ManToAig(p, 0)
	var result any
	var err error
	if choices != 0 {
		params := &dar.ChoiceOptions{}
		result, err = dar.Dar_ManChoiceNew(graph, params)
	} else {
		result, err = dar.Dar_ManCompress2(graph, true, true, true, false, false)
	}
	if err != nil {
		panic(err)
	}
	outGraph := result
	// Conversion bridges are the native graph path used by the source synthesis routines.
	newGia := gia.Gia_ManFromAig(outGraph.(*aig.Graph))
	if manager, extra := gia.Gia_ManBoxData(p); manager != nil || extra != nil {
		if cloner, ok := manager.(interface{ CloneTiming() any }); ok {
			manager = cloner.CloneTiming()
		}
		if extra != nil {
			extra = gia.Gia_ManDup(extra)
		}
		gia.Gia_ManSetTimingData(newGia, manager, extra)
	}
	return newGia
}

func Gia_ManVerifyChoices(p *Gia_Man_t) {
	if !gia.Gia_ManHasReprClasses(p) {
		return
	}
	problem := false
	seen := make(map[int32]bool)
	for repr := int32(0); repr < gia.Gia_ManObjNum(p); repr++ {
		if gia.Gia_ObjIsHead(p, repr) == 0 {
			continue
		}
		for id := gia.Gia_ObjNext(p, repr); id > 0; id = gia.Gia_ObjNext(p, id) {
			if id == repr {
				fmt.Printf("Member %d of choice class %d is a representative.\n", id, repr)
				problem = true
			}
			if seen[id] {
				fmt.Printf("Node %d participates in more than one choice node.\n", id)
				problem = true
			}
			seen[id] = true
		}
	}
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 && gia.Gia_ObjIsCo(o) == 0 {
			continue
		}
		a := gia.Gia_ObjFaninId0(o, id)
		if gia.Gia_ObjHasRepr(p, a) != 0 {
			fmt.Printf("Fanin 0 of object %d has a repr.\n", id)
			problem = true
		}
		if gia.Gia_ObjIsAnd(o) != 0 {
			b := gia.Gia_ObjFaninId1(o, id)
			if gia.Gia_ObjHasRepr(p, b) != 0 {
				fmt.Printf("Fanin 1 of AND node %d has a repr.\n", id)
				problem = true
			}
		}
	}
	_ = problem
}

func Gia_ManReverseClasses(p *Gia_Man_t, nowIncreasing int32) {
	if !gia.Gia_ManHasReprClasses(p) {
		return
	}
	heads := []int32{}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsHead(p, id) != 0 {
			heads = append(heads, id)
		}
	}
	for _, head := range heads {
		members := []int32{head}
		for id := gia.Gia_ObjNext(p, head); id > 0; id = gia.Gia_ObjNext(p, id) {
			members = append(members, id)
		}
		if len(members) < 2 {
			continue
		}
		last := members[len(members)-1]
		for _, id := range members {
			if nowIncreasing != 0 {
				if head >= id && id != head {
					panic("choice class is not in increasing order")
				}
				if id == last {
					gia.Gia_ObjUnsetRepr(p, id)
				} else {
					gia.Gia_ObjSetReprRev(p, id, last)
				}
			} else {
				if head <= id && id != head {
					panic("choice class is not in decreasing order")
				}
				if id == last {
					gia.Gia_ObjUnsetRepr(p, id)
				} else {
					gia.Gia_ObjSetRepr(p, id, last)
				}
			}
		}
		for i := range members {
			next := int32(0)
			if i > 0 {
				next = members[i-1]
			}
			gia.Gia_ObjSetNext(p, members[i], next)
		}
	}
}

func Abc_NtkTestTimByWritingFile(p *Gia_Man_t, name string) {
	if p == nil {
		return
	}
	has := gia.Gia_ManHasReprClasses(p)
	if has {
		Gia_ManVerifyChoices(p)
		Gia_ManReverseClasses(p, 0)
	}
	if err := gia.Gia_AigerWrite(p, name, 0, 0, 0); err != nil {
		fmt.Printf("Could not write %s: %v\n", name, err)
	}
	if has {
		Gia_ManReverseClasses(p, 1)
	}
	q, err := gia.Gia_AigerRead(name, 0, 1, 1)
	if err != nil {
		fmt.Printf("Could not read %s: %v\n", name, err)
		return
	}
	if gia.Gia_ManHasReprClasses(q) {
		Gia_ManVerifyChoices(q)
		Gia_ManReverseClasses(q, 1)
	}
	if gia.Gia_ManCompare(p, q) != 0 {
		fmt.Println("Verification successful.")
	}
	gia.Gia_ManStopTimingData(q)
	gia.Gia_ManStop(q)
}

func Abc_NtkTestTim(n *Abc_Ntk_t, verbose int32) {
	if n == nil || baseabc.Abc_NtkIsLogic(n) == 0 {
		panic("TIM test requires a logic network")
	}
	baseabc.Abc_NtkToAig(n)
	g := Abc_NtkTestTimDeriveGia(n, verbose)
	manager, _ := gia.Gia_ManBoxData(g)
	tm, _ := manager.(*tim.Tim_Man_t)
	boxes := 0
	if tm != nil {
		boxes = tim.Tim_ManBoxNum(tm)
	}
	fmt.Printf("Created GIA manager for network with %d white boxes.\n", boxes)
	if verbose != 0 && tm != nil {
		tim.Tim_ManPrint(tm)
	}
	Abc_NtkTestTimByWritingFile(g, "test1.aig")
	newGia := Abc_NtkTestTimPerformSynthesis(g, 0)
	Abc_NtkTestTimByWritingFile(newGia, "test2.aig")
	gia.Gia_ManStopTimingData(g)
	gia.Gia_ManStop(g)
	gia.Gia_ManStopTimingData(newGia)
	gia.Gia_ManStop(newGia)
}
