package abci

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec"
)

// Dch_Pars_t retains the C fields used by ABC's DCH callers. The focused
// Dress2 path consumes only nBTLimit and fVerbose; the remaining fields keep
// the shared command API source-shaped.
type Dch_Pars_t struct {
	nWords, nBTLimit, nSatVarMax                          int32
	fSynthesis, fPolarFlip, fSimulateTfo, fPower          int32
	fUseGia, fUseCSat, fUseNew, fLightSynth, fSkipRedSupp int32
	fVerbose, nNodesAhead, nCallsRecycle                  int32
	timeSynth                                             int64
}

func Dch_ManSetDefaultParams(p *Dch_Pars_t) {
	*p = Dch_Pars_t{
		nWords: 8, nBTLimit: 1000, nSatVarMax: 5000,
		fSynthesis: 1, fPolarFlip: 1, fSimulateTfo: 1,
		nNodesAhead: 1000, nCallsRecycle: 100,
	}
}

func Abc_ObjEquivId2ObjId(id int32) int32 { return id >> 2 }
func Abc_ObjEquivId2Polar(id int32) int32 { return (id >> 1) & 1 }
func Abc_ObjEquivId2NtkId(id int32) int32 { return id & 1 }

// Aig_ManCreateDualOutputMiter copies both AIGs into one manager, sharing
// corresponding CIs and interleaving each pair of outputs as in abcDress2.c.
func Aig_ManCreateDualOutputMiter(p1, p2 *aig.Graph) *aig.Graph {
	assert(p1 != nil && p2 != nil, "non-nil AIG managers")
	assert(aig.Aig_ManCiNum(p1) == aig.Aig_ManCiNum(p2), "matching CI counts")
	assert(aig.Aig_ManCoNum(p1) == aig.Aig_ManCoNum(p2), "matching CO counts")
	pNew := aig.Aig_ManStart(aig.Aig_ManObjNumMax(p1) + aig.Aig_ManObjNumMax(p2))
	copyInputs := func(src *aig.Graph, shared []aig.Edge) {
		aig.Aig_ObjSetCopy(src, src.Const1(), pNew.Const1())
		for i, input := range src.Inputs() {
			if shared == nil {
				aig.Aig_ObjSetCopy(src, input, aig.Aig_ObjCreateCi(pNew))
			} else {
				aig.Aig_ObjSetCopy(src, input, shared[i])
			}
		}
	}
	copyNodes := func(src *aig.Graph) {
		for id, node := range src.Nodes() {
			if node.Kind != aig.And {
				continue
			}
			copyEdge := func(edge aig.Edge) aig.Edge {
				mapped := aig.Aig_ObjCopy(src, aig.Aig_Regular(edge))
				assert(mapped != aig.InvalidEdge, "AIG fanin copied before its node")
				if edge.Complemented() {
					mapped = mapped.Not()
				}
				return mapped
			}
			copy := pNew.AND(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			aig.Aig_ObjSetCopy(src, aig.Edge(id)<<1, copy)
		}
	}
	copyInputs(p1, nil)
	shared := pNew.Inputs()
	copyNodes(p1)
	copyInputs(p2, shared)
	copyNodes(p2)
	copyEdge := func(src *aig.Graph, edge aig.Edge) aig.Edge {
		mapped := aig.Aig_ObjCopy(src, aig.Aig_Regular(edge))
		assert(mapped != aig.InvalidEdge, "AIG output driver copied")
		if edge.Complemented() {
			mapped = mapped.Not()
		}
		return mapped
	}
	for i := range p1.Outputs() {
		out1 := p1.Node(p1.Outputs()[i]).Fanins[0]
		out2 := p2.Node(p2.Outputs()[i]).Fanins[0]
		pNew.AddOutput(copyEdge(p1, out1))
		pNew.AddOutput(copyEdge(p2, out2))
	}
	aig.Aig_ManCleanup(pNew)
	return pNew
}

func Abc_NtkDressMapSetPolarity(n *abc.Network) {
	for _, obj := range n.Objects() {
		edge := obj.CopyEdge
		if edge.Object != nil && edge.Object.Kind != abc.ABC_OBJ_NONE {
			obj.Phase = edge.Object.Phase != edge.Complement
		}
	}
}

// Abc_NtkDressMapClasses maps original node IDs to miter representative IDs.
// darManagers carries the source AIG manager needed to resolve the copy chain;
// it is variadic to preserve the original two-argument API for callers that
// only need the allocation/initialization behavior.
func Abc_NtkDressMapClasses(miter *aig.Graph, n *abc.Network, darManagers ...*aig.Graph) *vec.Vec_Int_t {
	vId2Lit := vec.Vec_IntAlloc(0)
	vec.Vec_IntFill(vId2Lit, int(abc.Abc_NtkObjNumMax(n)), -1)
	if len(darManagers) == 0 || darManagers[0] == nil {
		return vId2Lit
	}
	dar := darManagers[0]
	for _, obj := range n.ObjectsOfKind(abc.ABC_OBJ_NODE) {
		strashEdge := obj.CopyEdge
		if strashEdge.Object == nil || strashEdge.Object.Kind == abc.ABC_OBJ_NONE {
			continue
		}
		darEdge := aig.Edge(strashEdge.Object.Temp)
		if darEdge == aig.InvalidEdge || aig.Aig_ObjIsNone(dar, darEdge) {
			continue
		}
		miterNode := aig.Aig_ObjCopy(dar, aig.Aig_Regular(darEdge))
		if miterNode == aig.InvalidEdge || aig.Aig_ObjIsNone(miter, miterNode) {
			continue
		}
		repr := aig.Aig_ObjRepr(miter, miterNode)
		if repr == aig.InvalidEdge {
			repr = miterNode
		}
		vec.Vec_IntWriteEntry(vId2Lit, obj.ID, int32(repr.NodeID()))
	}
	return vId2Lit
}

func Abc_ObjDressClass(result *vec.Vec_Ptr_t, class2num *vec.Vec_Int_t, class int32) *vec.Vec_Int_t {
	assert(class > 0, "class > 0")
	classNumber := vec.Vec_IntEntry(class2num, int(class))
	assert(classNumber != 0, "class number is initialized")
	if classNumber > 0 {
		return vec.Vec_PtrEntry(result, int(classNumber)).(*vec.Vec_Int_t)
	}
	vec.Vec_IntWriteEntry(class2num, int(class), int32(vec.Vec_PtrSize(result)))
	classVector := vec.Vec_IntAlloc(4)
	vec.Vec_PtrPush(result, classVector)
	return classVector
}

func Abc_ObjDressMakeId(n *abc.Network, objectID, networkID int32) int32 {
	obj := abc.Abc_NtkObj(n, objectID)
	phase := int32(0)
	if obj.Phase {
		phase = 1
	}
	return (objectID << 2) | (phase << 1) | networkID
}

func Abc_NtkDressMapIds(miter *aig.Graph, n1, n2 *abc.Network, managers ...*aig.Graph) *vec.Vec_Ptr_t {
	assert(len(managers) >= 2, "both intermediate AIG managers are provided")
	result := vec.Vec_PtrAlloc(1000)
	Abc_NtkDressMapSetPolarity(n1)
	Abc_NtkDressMapSetPolarity(n2)
	id2lit1 := Abc_NtkDressMapClasses(miter, n1, managers[0])
	id2lit2 := Abc_NtkDressMapClasses(miter, n2, managers[1])
	counts0 := vec.Vec_IntStart(int(aig.Aig_ManObjNumMax(miter)))
	counts1 := vec.Vec_IntStart(int(aig.Aig_ManObjNumMax(miter)))
	for i := 0; i < vec.Vec_IntSize(id2lit1); i++ {
		class := vec.Vec_IntEntry(id2lit1, i)
		if class >= 0 {
			vec.Vec_IntAddToEntry(counts0, int(class), 1)
		}
	}
	for i := 0; i < vec.Vec_IntSize(id2lit2); i++ {
		class := vec.Vec_IntEntry(id2lit2, i)
		if class >= 0 {
			vec.Vec_IntAddToEntry(counts1, int(class), 1)
		}
	}
	constantClass := vec.Vec_IntAlloc(100)
	for i := 0; i < vec.Vec_IntSize(id2lit1); i++ {
		if vec.Vec_IntEntry(id2lit1, i) == 0 {
			vec.Vec_IntPush(constantClass, Abc_ObjDressMakeId(n1, int32(i), 0))
		}
	}
	for i := 0; i < vec.Vec_IntSize(id2lit2); i++ {
		if vec.Vec_IntEntry(id2lit2, i) == 0 {
			vec.Vec_IntPush(constantClass, Abc_ObjDressMakeId(n2, int32(i), 1))
		}
	}
	vec.Vec_PtrPush(result, constantClass)
	class2num := vec.Vec_IntAlloc(0)
	vec.Vec_IntFill(class2num, int(aig.Aig_ManObjNumMax(miter)), -1)
	for i := 0; i < vec.Vec_IntSize(id2lit1); i++ {
		class := vec.Vec_IntEntry(id2lit1, i)
		if class > 0 && vec.Vec_IntEntry(counts0, int(class)) != 0 && vec.Vec_IntEntry(counts1, int(class)) != 0 {
			vec.Vec_IntPush(Abc_ObjDressClass(result, class2num, class), Abc_ObjDressMakeId(n1, int32(i), 0))
		}
	}
	for i := 0; i < vec.Vec_IntSize(id2lit2); i++ {
		class := vec.Vec_IntEntry(id2lit2, i)
		if class > 0 && vec.Vec_IntEntry(counts0, int(class)) != 0 && vec.Vec_IntEntry(counts1, int(class)) != 0 {
			vec.Vec_IntPush(Abc_ObjDressClass(result, class2num, class), Abc_ObjDressMakeId(n2, int32(i), 1))
		}
	}
	vec.Vec_IntFree(class2num)
	vec.Vec_IntFree(counts0)
	vec.Vec_IntFree(counts1)
	vec.Vec_IntFree(id2lit1)
	vec.Vec_IntFree(id2lit2)
	return result
}

func Dch_ComputeEquivalences2(miter *aig.Graph, pars *Dch_Pars_t) {
	pGia := gia.Gia_ManFromAigSimple(miter)
	pNew := cec.Cec4_ManSimulateTest3(pGia, pars.nBTLimit, pars.fVerbose)
	assert(gia.Gia_ManObjNum(pGia) == aig.Aig_ManObjNumMax(miter), "simple GIA conversion preserves object IDs")
	for head := int32(1); head < gia.Gia_ManObjNum(pGia); head++ {
		if gia.Gia_ObjIsHead(pGia, head) == 0 {
			continue
		}
		for member := gia.Gia_ObjNext(pGia, head); member > 0; member = gia.Gia_ObjNext(pGia, member) {
			aig.Aig_ObjSetRepr(miter, aig.Aig_ManObj(miter, member), aig.Aig_ManObj(miter, head))
		}
	}
	gia.Gia_ManStop(pNew)
	gia.Gia_ManStop(pGia)
}

func Abc_NtkDressComputeEquivs(n1, n2 *abc.Network, conflictLimit, verbose int32) *vec.Vec_Ptr_t {
	assert(abc.Abc_NtkIsStrash(n1) == 0, "first network is not strashed")
	assert(abc.Abc_NtkIsStrash(n2) == 0, "second network is not strashed")
	strash1 := abc.Abc_NtkStrash(n1, 1, 1, 0)
	strash2 := abc.Abc_NtkStrash(n2, 1, 1, 0)
	man1 := abc.Abc_NtkToDar(strash1, 0, 0)
	man2 := abc.Abc_NtkToDar(strash2, 0, 0)
	miter := Aig_ManCreateDualOutputMiter(man1, man2)
	var pars Dch_Pars_t
	Dch_ManSetDefaultParams(&pars)
	pars.nBTLimit, pars.fVerbose = conflictLimit, verbose
	Dch_ComputeEquivalences2(miter, &pars)
	result := Abc_NtkDressMapIds(miter, n1, n2, man1, man2)
	aig.Aig_ManStop(miter)
	aig.Aig_ManStop(man1)
	aig.Aig_ManStop(man2)
	abc.Abc_NtkDelete(strash1)
	abc.Abc_NtkDelete(strash2)
	return result
}

func Abc_NtkDressPrintEquivs(result *vec.Vec_Ptr_t) {
	for i := 0; i < vec.Vec_PtrSize(result); i++ {
		class := vec.Vec_PtrEntry(result, i).(*vec.Vec_Int_t)
		fmt.Printf("Class %5d : Num =%5d    ", i, vec.Vec_IntSize(class))
		for k := 0; k < vec.Vec_IntSize(class); k++ {
			entry := vec.Vec_IntEntry(class, k)
			polarity := byte('+')
			if Abc_ObjEquivId2Polar(entry) != 0 {
				polarity = '-'
			}
			fmt.Printf("%5d%c%d ", Abc_ObjEquivId2ObjId(entry), polarity, Abc_ObjEquivId2NtkId(entry))
		}
		fmt.Println()
	}
}

func Abc_NtkDressPrintStats(result *vec.Vec_Ptr_t, nodes0, nodes1 int32, elapsed int64) {
	var negAll, posAll [2]int32
	var pairsAll, pairsSame int32
	for i := 0; i < vec.Vec_PtrSize(result); i++ {
		class := vec.Vec_PtrEntry(result, i).(*vec.Vec_Int_t)
		var pos, neg [2]int32
		for k := 0; k < vec.Vec_IntSize(class); k++ {
			entry := vec.Vec_IntEntry(class, k)
			network, phase := Abc_ObjEquivId2NtkId(entry), Abc_ObjEquivId2Polar(entry)
			if phase != 0 {
				neg[network]++
			} else {
				pos[network]++
			}
		}
		for network := range 2 {
			posAll[network] += pos[network]
			negAll[network] += neg[network]
		}
		pairsAll += minDress(neg[0]+pos[0], neg[1]+pos[1])
		pairsSame += minDress(neg[0], neg[1]) + minDress(pos[0], pos[1])
	}
	count0, count1 := negAll[0]+posAll[0], negAll[1]+posAll[1]
	fmt.Printf("Total number of equiv classes                = %7d.\n", vec.Vec_PtrSize(result))
	fmt.Printf("Participating nodes from both networks       = %7d.\n", count0+count1)
	fmt.Printf("Participating nodes from the first network   = %7d. (%7.2f %% of nodes)\n", count0, 100*float64(count0)/float64(nodes0+1))
	fmt.Printf("Participating nodes from the second network  = %7d. (%7.2f %% of nodes)\n", count1, 100*float64(count1)/float64(nodes1+1))
	fmt.Printf("Node pairs (any polarity)                    = %7d. (%7.2f %% of names can be moved)\n", pairsAll, 100*float64(pairsAll)/float64(nodes0+1))
	fmt.Printf("Node pairs (same polarity)                   = %7d. (%7.2f %% of names can be moved)\n", pairsSame, 100*float64(pairsSame)/float64(nodes0+1))
	fmt.Printf("Total runtime =%9.2f sec\n", float64(elapsed)/float64(time.Second))
}

func minDress(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func Abc_NtkDress2Transfer(n0, n1 *abc.Network, result *vec.Vec_Ptr_t, verbose int32) {
	var assigned, inverted int32
	for i := 0; i < vec.Vec_PtrSize(result); i++ {
		class := vec.Vec_PtrEntry(result, i).(*vec.Vec_Int_t)
		var obj0, obj1 *abc.Object
		var phase0, phase1 int32
		for k := 0; k < vec.Vec_IntSize(class); k++ {
			entry := vec.Vec_IntEntry(class, k)
			if Abc_ObjEquivId2NtkId(entry) != 0 {
				obj1 = abc.Abc_NtkObj(n1, Abc_ObjEquivId2ObjId(entry))
				phase1 = Abc_ObjEquivId2Polar(entry)
			} else {
				obj0 = abc.Abc_NtkObj(n0, Abc_ObjEquivId2ObjId(entry))
				phase0 = Abc_ObjEquivId2Polar(entry)
			}
		}
		if obj0 == nil || obj1 == nil || obj0.Name() != "" || obj1.Name() == "" {
			continue
		}
		suffix := ""
		if phase0^phase1 != 0 {
			suffix = "_inv"
			inverted++
		} else {
			assigned++
		}
		abc.Abc_ObjAssignName(obj0, obj1.Name(), suffix)
	}
	if verbose != 0 {
		fmt.Printf("Total number of names assigned  = %5d. (Dir = %5d. Compl = %5d.)\n", assigned+inverted, assigned, inverted)
	}
}

func Abc_NtkDress2(n1, n2 *abc.Network, conflictLimit, verbose int32) {
	start := time.Now()
	result := Abc_NtkDressComputeEquivs(n1, n2, conflictLimit, verbose)
	Abc_NtkDressPrintStats(result, abc.Abc_NtkNodeNum(n1), abc.Abc_NtkNodeNum(n1), time.Since(start).Nanoseconds())
	Abc_NtkDress2Transfer(n1, n2, result, verbose)
	vec.Vec_VecFree(result)
}
