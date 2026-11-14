package abci

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec"
)

type Gia_Obj_t = gia.Gia_Obj_t

// Abc_ConvertHopToGia_rec1 converts the marked-once HOP DAG rooted at obj.
func Abc_ConvertHopToGia_rec1(p *Gia_Man_t, obj *Hop_Obj_t) {
	if obj == nil || hop.Hop_IsComplement(obj) != 0 {
		panic("Abc_ConvertHopToGia_rec1 requires a regular HOP object")
	}
	if hop.Hop_ObjIsNode(obj) == 0 || hop.Hop_ObjIsMarkA(obj) != 0 {
		return
	}
	Abc_ConvertHopToGia_rec1(p, hop.Hop_ObjFanin0(obj))
	Abc_ConvertHopToGia_rec1(p, hop.Hop_ObjFanin1(obj))
	if hop.Hop_ObjIsMarkA(obj) != 0 {
		panic("Abc_ConvertHopToGia_rec1 detected a HOP cycle")
	}
	lit := gia.Gia_ManHashAnd(p, hop.Hop_ObjChild0CopyI(obj), hop.Hop_ObjChild1CopyI(obj))
	hop.Hop_ObjSetIData(obj, lit)
	hop.Hop_ObjSetMarkA(obj)
}

// Abc_ConvertHopToGia_rec2 clears the temporary HOP marks set by conversion.
func Abc_ConvertHopToGia_rec2(obj *Hop_Obj_t) {
	if obj == nil || hop.Hop_IsComplement(obj) != 0 {
		panic("Abc_ConvertHopToGia_rec2 requires a regular HOP object")
	}
	if hop.Hop_ObjIsNode(obj) == 0 || hop.Hop_ObjIsMarkA(obj) == 0 {
		return
	}
	Abc_ConvertHopToGia_rec2(hop.Hop_ObjFanin0(obj))
	Abc_ConvertHopToGia_rec2(hop.Hop_ObjFanin1(obj))
	if hop.Hop_ObjIsMarkA(obj) == 0 {
		panic("Abc_ConvertHopToGia_rec2 detected a HOP cycle")
	}
	hop.Hop_ObjClearMarkA(obj)
}

func Abc_ConvertHopToGia(p *Gia_Man_t, root *Hop_Obj_t) int32 {
	if root == nil || hop.Hop_IsComplement(root) != 0 {
		panic("Abc_ConvertHopToGia requires a regular HOP root")
	}
	if hop.Hop_ObjIsConst1(root) != 0 {
		return 1
	}
	Abc_ConvertHopToGia_rec1(p, root)
	Abc_ConvertHopToGia_rec2(root)
	return hop.Hop_ObjIData(root)
}

// Abc_NtkAigToGiaOne adds one AIG-logic network to p using vMap for CIs.
func Abc_NtkAigToGiaOne(p *Gia_Man_t, network *Abc_Ntk_t, vMap *Vec_Int_t) {
	if abc.Abc_NtkIsAigLogic(network) == 0 {
		panic("Abc_NtkAigToGiaOne requires AIG logic")
	}
	hopMan, ok := network.FunctionManager.(*hop.Hop_Man_t)
	if !ok || hopMan == nil {
		panic("AIG logic network has no HOP function manager")
	}
	hop.Hop_ObjSetIData(hop.Hop_ManConst1(hopMan), 1)
	abc.Abc_NtkCleanCopy(network)
	for i := int32(0); i < abc.Abc_NtkCiNum(network); i++ {
		abc.Abc_ObjSetTemp(abc.Abc_NtkCi(network, i), gia.Gia_ManCiLit(p, vec.Vec_IntEntry(vMap, int(i))))
	}
	for _, node := range abc.Abc_NtkDfs(network, 1) {
		for k, fanin := range node.Fanins() {
			if int32(k) >= hop.Hop_ManPiNum(hopMan) {
				panic("AIG node has more fanins than its HOP function")
			}
			pi := hop.Hop_ManPi(hopMan, int32(k))
			hop.Hop_ObjSetIData(pi, abc.Abc_ObjTemp(fanin.Object)^boolToLit(fanin.Complement))
		}
		root, ok := abc.Abc_ObjData(node).(*hop.Hop_Obj_t)
		if !ok || root == nil {
			panic("AIG node has no HOP function")
		}
		regular := hop.Hop_Regular(root)
		if hop.Hop_DagSize(regular) > 0 {
			Abc_ConvertHopToGia(p, regular)
		}
		lit := hop.Hop_ObjIData(regular)
		if hop.Hop_IsComplement(root) != 0 {
			lit ^= 1
		}
		abc.Abc_ObjSetTemp(node, lit)
	}
	for i := int32(0); i < abc.Abc_NtkCoNum(network); i++ {
		co := abc.Abc_NtkCo(network, i)
		fanins := co.Fanins()
		if len(fanins) != 1 {
			panic("AIG CO must have one fanin")
		}
		gia.Gia_ManAppendCo(p, abc.Abc_ObjTemp(fanins[0].Object)^boolToLit(fanins[0].Complement))
	}
}

func boolToLit(complement bool) int32 {
	if complement {
		return 1
	}
	return 0
}

// Abc_NtkAigToGiaTwo structurally hashes two AIG networks into shared GIA.
func Abc_NtkAigToGiaTwo(first, second *Abc_Ntk_t, byName int32) *Gia_Man_t {
	if abc.Abc_NtkIsAigLogic(first) == 0 || abc.Abc_NtkIsAigLogic(second) == 0 {
		panic("Abc_NtkAigToGiaTwo requires AIG logic networks")
	}
	map1 := vec.Vec_IntStartNatural(int(abc.Abc_NtkCiNum(first)))
	map2 := vec.Vec_IntAlloc(int(abc.Abc_NtkCiNum(second)))
	index := int32(0)
	if byName != 0 {
		for i := int32(0); i < abc.Abc_NtkCiNum(first); i++ {
			abc.Abc_ObjSetTemp(abc.Abc_NtkCi(first, i), index)
			index++
		}
		common := int32(0)
		for i := int32(0); i < abc.Abc_NtkCiNum(second); i++ {
			ci := abc.Abc_NtkCi(second, i)
			matched := abc.Abc_NtkFindCi(first, abc.Abc_ObjName(ci))
			if matched != nil {
				vec.Vec_IntPush(map2, abc.Abc_ObjTemp(matched))
				common++
			} else {
				vec.Vec_IntPush(map2, index)
				index++
			}
		}
		fmt.Printf("Matched %d vars by name.", common)
		if common != abc.Abc_NtkCiNum(first) {
			fmt.Printf(" Netlist1 has %d unmatched vars.", abc.Abc_NtkCiNum(first)-common)
		}
		if common != abc.Abc_NtkCiNum(second) {
			fmt.Printf(" Netlist2 has %d unmatched vars.", abc.Abc_NtkCiNum(second)-common)
		}
		fmt.Println()
	} else {
		map2 = vec.Vec_IntStartNatural(int(abc.Abc_NtkCiNum(second)))
		index = int32(max(abc.Abc_NtkCiNum(first), abc.Abc_NtkCiNum(second)))
		fmt.Printf("Matched %d vars by order.", min(abc.Abc_NtkCiNum(first), abc.Abc_NtkCiNum(second)))
		if abc.Abc_NtkCiNum(first) < abc.Abc_NtkCiNum(second) {
			fmt.Printf(" The last %d vars of Netlist2 are unmatched vars.", abc.Abc_NtkCiNum(second)-abc.Abc_NtkCiNum(first))
		}
		if abc.Abc_NtkCiNum(first) > abc.Abc_NtkCiNum(second) {
			fmt.Printf(" The last %d vars of Netlist1 are unmatched vars.", abc.Abc_NtkCiNum(first)-abc.Abc_NtkCiNum(second))
		}
		fmt.Println()
	}
	p := gia.Gia_ManStart(10000)
	gia.Gia_ManSetNameSpec(p, abc.Abc_NtkName(first), abc.Abc_NtkSpec(first))
	for i := int32(0); i < index; i++ {
		gia.Gia_ManAppendCi(p)
	}
	gia.Gia_ManHashAlloc(p)
	Abc_NtkAigToGiaOne(p, first, map1)
	Abc_NtkAigToGiaOne(p, second, map2)
	gia.Gia_ManHashStop(p)
	vec.Vec_IntFree(map1)
	vec.Vec_IntFree(map2)
	gia.Gia_ManCreateValueRefs(p)
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		obj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(obj) != 0 && gia.Gia_ObjValue(obj) == 0 {
			gia.Gia_ManAppendCo(p, i<<1)
		}
	}
	return p
}

func Abc_NtkCollectAddOne(networkID, objectID, giaLit int32, graph *Gia_Man_t, giaToClass *Vec_Int_t, nexts []*Vec_Int_t) {
	objID := util.Abc_Lit2Var(giaLit)
	repr := gia.Gia_ObjReprSelf(graph, objID)
	compl := util.Abc_LitIsCompl(giaLit) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(graph, repr)) ^ gia.Gia_ObjPhase(gia.Gia_ManObj(graph, objID))
	added := util.Abc_Var2Lit(util.Abc_Var2Lit(objectID, compl), networkID)
	entry := vec.Vec_IntEntry(giaToClass, int(repr))
	vec.Vec_IntWriteEntry(nexts[networkID], int(objectID), entry)
	vec.Vec_IntWriteEntry(giaToClass, int(repr), added)
}

// Abc_NtkCollectEquivClasses emits packed count-prefixed equivalence classes.
func Abc_NtkCollectEquivClasses(networks []*Abc_Ntk_t, graph *Gia_Man_t) *Vec_Int_t {
	if len(networks) != 2 {
		panic("equivalence collection requires exactly two networks")
	}
	classes := vec.Vec_IntAlloc(1000)
	class := vec.Vec_IntAlloc(100)
	giaToClass := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(graph)))
	nexts := []*Vec_Int_t{
		vec.Vec_IntStartFull(int(abc.Abc_NtkObjNumMax(networks[0]))),
		vec.Vec_IntStartFull(int(abc.Abc_NtkObjNumMax(networks[1]))),
	}
	for i := int32(0); i < abc.Abc_NtkCiNum(networks[0]); i++ {
		o := abc.Abc_NtkCi(networks[0], i)
		Abc_NtkCollectAddOne(0, abc.Abc_ObjId(o), abc.Abc_ObjTemp(o), graph, giaToClass, nexts)
	}
	for n := int32(0); n < 2; n++ {
		for _, o := range networks[n].ObjectsOfKind(abc.ObjectNode) {
			Abc_NtkCollectAddOne(n, abc.Abc_ObjId(o), abc.Abc_ObjTemp(o), graph, giaToClass, nexts)
		}
	}
	for i := 0; i < vec.Vec_IntSize(giaToClass); i++ {
		entry := vec.Vec_IntEntry(giaToClass, i)
		vec.Vec_IntClear(class)
		for entry >= 0 {
			vec.Vec_IntPush(class, entry)
			entry = vec.Vec_IntEntry(nexts[entry&1], int(entry>>2))
		}
		if vec.Vec_IntSize(class) < 2 {
			continue
		}
		vec.Vec_IntReverseOrder(class)
		compl := 2 & vec.Vec_IntEntry(class, 0)
		for k := 0; k < vec.Vec_IntSize(class); k++ {
			vec.Vec_IntWriteEntry(class, k, vec.Vec_IntEntry(class, k)^compl)
		}
		vec.Vec_IntPush(classes, int32(vec.Vec_IntSize(class)))
		vec.Vec_IntAppend(classes, class)
	}
	vec.Vec_IntFree(giaToClass)
	vec.Vec_IntFree(nexts[0])
	vec.Vec_IntFree(nexts[1])
	vec.Vec_IntFree(class)
	return classes
}

func Abc_NtkDumpEquivFile(fileName string, classes *Vec_Int_t, networks []*Abc_Ntk_t) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Cannot open file %s for writing.\n", fileName)
		return
	}
	defer file.Close()
	fmt.Fprintf(file, "# Node equivalences computed by ABC for networks \"%s\" and \"%s\" on %s\n\n", abc.Abc_NtkName(networks[0]), abc.Abc_NtkName(networks[1]), extra.Extra_TimeStamp())
	for i, classID := 0, int32(0); i < vec.Vec_IntSize(classes); classID++ {
		count := int(vec.Vec_IntEntry(classes, i))
		for k := 0; k < count; k++ {
			entry := vec.Vec_IntEntry(classes, i+1+k)
			network := networks[entry&1]
			object := abc.Abc_NtkObj(network, entry>>2)
			prefix := ""
			if entry&2 != 0 {
				prefix = "NOT:"
			}
			fmt.Fprintf(file, "%d:%s:%s%s\n", classID+1, abc.Abc_NtkName(network), prefix, abc.Abc_ObjName(object))
		}
		fmt.Fprintln(file)
		i += 1 + count
	}
}

func Abc_NtkDumpEquiv(networks []*Abc_Ntk_t, fileName string, conflicts, byName, verbose int32) {
	if len(networks) != 2 {
		panic("equivalence dump requires exactly two networks")
	}
	graph := Abc_NtkAigToGiaTwo(networks[0], networks[1], byName)
	if verbose != 0 {
		fmt.Printf("Computing equivalences for networks \"%s\" and \"%s\" with conflict limit %d.\n", abc.Abc_NtkName(networks[0]), abc.Abc_NtkName(networks[1]), conflicts)
	}
	temporary := cec.Gia_ManComputeGiaEquivs(graph, conflicts, verbose)
	gia.Gia_ManStop(temporary)
	if verbose != 0 {
		gia.Gia_ManPrintStats(graph, nil)
	}
	classes := Abc_NtkCollectEquivClasses(networks, graph)
	gia.Gia_ManStop(graph)
	Abc_NtkDumpEquivFile(fileName, classes, networks)
	vec.Vec_IntFree(classes)
}
