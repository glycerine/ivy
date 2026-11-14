package abci

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Abc_FinType_t int32

const (
	ABC_FIN_NONE      Abc_FinType_t = -100
	ABC_FIN_SA0       Abc_FinType_t = -99
	ABC_FIN_SA1       Abc_FinType_t = -98
	ABC_FIN_NEG       Abc_FinType_t = -97
	ABC_FIN_RDOB_AND  Abc_FinType_t = -96
	ABC_FIN_RDOB_NAND Abc_FinType_t = -95
	ABC_FIN_RDOB_OR   Abc_FinType_t = -94
	ABC_FIN_RDOB_NOR  Abc_FinType_t = -93
	ABC_FIN_RDOB_XOR  Abc_FinType_t = -92
	ABC_FIN_RDOB_NXOR Abc_FinType_t = -91
	ABC_FIN_RDOB_NOT  Abc_FinType_t = -90
	ABC_FIN_RDOB_BUFF Abc_FinType_t = -89
	ABC_FIN_RDOB_LAST Abc_FinType_t = -88
)

func Io_ReadFinType(name string) int32 {
	switch name {
	case "SA0":
		return int32(ABC_FIN_SA0)
	case "SA1":
		return int32(ABC_FIN_SA1)
	case "NEG":
		return int32(ABC_FIN_NEG)
	case "RDOB_AND":
		return int32(ABC_FIN_RDOB_AND)
	case "RDOB_NAND":
		return int32(ABC_FIN_RDOB_NAND)
	case "RDOB_OR":
		return int32(ABC_FIN_RDOB_OR)
	case "RDOB_NOR":
		return int32(ABC_FIN_RDOB_NOR)
	case "RDOB_XOR":
		return int32(ABC_FIN_RDOB_XOR)
	case "RDOB_NXOR":
		return int32(ABC_FIN_RDOB_NXOR)
	case "RDOB_NOT":
		return int32(ABC_FIN_RDOB_NOT)
	case "RDOB_BUFF":
		return int32(ABC_FIN_RDOB_BUFF)
	default:
		return int32(ABC_FIN_NONE)
	}
}

func Io_WriteFinType(kind int32) string {
	switch Abc_FinType_t(kind) {
	case ABC_FIN_SA0:
		return "SA0"
	case ABC_FIN_SA1:
		return "SA1"
	case ABC_FIN_NEG:
		return "NEG"
	case ABC_FIN_RDOB_AND:
		return "RDOB_AND"
	case ABC_FIN_RDOB_NAND:
		return "RDOB_NAND"
	case ABC_FIN_RDOB_OR:
		return "RDOB_OR"
	case ABC_FIN_RDOB_NOR:
		return "RDOB_NOR"
	case ABC_FIN_RDOB_XOR:
		return "RDOB_XOR"
	case ABC_FIN_RDOB_NXOR:
		return "RDOB_NXOR"
	case ABC_FIN_RDOB_NOT:
		return "RDOB_NOT"
	case ABC_FIN_RDOB_BUFF:
		return "RDOB_BUFF"
	default:
		return "Unknown"
	}
}

func finLibrary(network *Abc_Ntk_t) *mio.Mio_Library_t {
	if network == nil {
		return nil
	}
	if library, ok := network.FunctionManager.(*mio.Mio_Library_t); ok {
		return library
	}
	if library, ok := network.Data.(*mio.Mio_Library_t); ok {
		return library
	}
	return nil
}

func Io_ReadFinTypeMapped(library *mio.Mio_Library_t, name string) int32 {
	if library == nil {
		return int32(ABC_FIN_NONE)
	}
	gate := mio.Mio_LibraryReadGateByName(library, name, "")
	if gate == nil {
		return int32(ABC_FIN_NONE)
	}
	return mio.Mio_GateReadCell(gate)
}

func Abc_NtkGenFaultList(network *Abc_Ntk_t, fileName string, stuckOnly int32) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("Cannot open file %q for writing.\n", fileName)
		return
	}
	defer file.Close()
	if baseabc.Abc_NtkIsMappedLogic(network) == 0 {
		panic("Abc_NtkGenFaultList requires a mapped logic network")
	}
	library := finLibrary(network)
	count := 1
	for id := int32(0); id < baseabc.Abc_NtkObjNumMax(network); id++ {
		obj := baseabc.Abc_NtkObj(network, id)
		if obj == nil || baseabc.Abc_ObjIsNode(obj) == 0 {
			continue
		}
		gate, _ := obj.Data.(*mio.Mio_Gate_t)
		if gate == nil {
			continue
		}
		for _, fault := range []string{"SA0", "SA1", "NEG"} {
			fmt.Fprintf(file, "%d %s %s\n", count, baseabc.Abc_ObjName(obj), fault)
			count++
		}
		if stuckOnly != 0 || library == nil {
			continue
		}
		for other := mio.Mio_LibraryReadGates(library); other != nil; other = mio.Mio_GateReadNext(other) {
			if other == gate || mio.Mio_GateReadPinNum(other) != mio.Mio_GateReadPinNum(gate) {
				continue
			}
			fmt.Fprintf(file, "%d %s %s\n", count, baseabc.Abc_ObjName(obj), mio.Mio_GateReadName(other))
			count++
		}
	}
	fmt.Printf("Generated fault list %q for network %q with %d nodes and %d %sfaults.\n", fileName, baseabc.Abc_NtkName(network), baseabc.Abc_NtkNodeNum(network), count-1, abcSelect(stuckOnly, "stuck-at ", ""))
}

func Io_ReadFins(network *Abc_Ntk_t, fileName string, verbose int32) *Vec_Int_t {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Cannot open input file %q for reading.\n", fileName)
		return nil
	}
	defer file.Close()
	nameToID := make(map[string]int32)
	for id := int32(0); id < baseabc.Abc_NtkObjNumMax(network); id++ {
		obj := baseabc.Abc_NtkObj(network, id)
		if obj == nil || (baseabc.Abc_ObjIsCi(obj) == 0 && baseabc.Abc_ObjIsNode(obj) == 0) {
			continue
		}
		name := baseabc.Abc_ObjName(obj)
		if _, exists := nameToID[name]; exists {
			fmt.Printf("The same name %q appears twice among CIs and internal nodes.\n", name)
			return nil
		}
		nameToID[name] = id
	}
	pairs := Vec_IntAlloc(1000)
	Vec_IntPush(pairs, -1)
	Vec_IntPush(pairs, -1)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1<<20)
	expectedLine := 1
	library := finLibrary(network)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			break
		}
		line, parseErr := strconv.Atoi(fields[0])
		if parseErr != nil || line != expectedLine {
			fmt.Println("Line numbers are not consecutive. Quitting.")
			Vec_IntFree(pairs)
			return nil
		}
		expectedLine++
		if len(fields) < 3 {
			continue
		}
		id, found := nameToID[fields[1]]
		if !found {
			fmt.Printf("Cannot find object with name %q.\n", fields[1])
			continue
		}
		kind := Io_ReadFinType(fields[2])
		if baseabc.Abc_NtkIsMappedLogic(network) != 0 && kind != int32(ABC_FIN_SA0) && kind != int32(ABC_FIN_SA1) && kind != int32(ABC_FIN_NEG) {
			kind = Io_ReadFinTypeMapped(library, fields[2])
		}
		if kind == int32(ABC_FIN_NONE) {
			fmt.Printf("Cannot read type %q of object %q.\n", fields[2], baseabc.Abc_ObjName(baseabc.Abc_NtkObj(network, id)))
			continue
		}
		Vec_IntPush(pairs, id)
		Vec_IntPush(pairs, kind)
	}
	if err := scanner.Err(); err != nil {
		Vec_IntFree(pairs)
		return nil
	}
	if verbose != 0 {
		for i := int32(2); i+1 < Vec_IntSize(pairs); i += 2 {
			id, kind := Vec_IntEntry(pairs, i), Vec_IntEntry(pairs, i+1)
			fmt.Printf("%-10d%-10s%-10s\n", i/2, baseabc.Abc_ObjName(baseabc.Abc_NtkObj(network, id)), Io_WriteFinType(kind))
		}
	}
	fmt.Printf("Finished reading %d lines from the fault list file %q.\n", expectedLine-1, fileName)
	return pairs
}

func Abc_NtkDetectObjClasses_rec(obj *Abc_Obj_t, mapping *Vec_Int_t, hash *vec.Hsh_VecMan_t, temp *Vec_Int_t) int32 {
	entry := Vec_IntEntry(mapping, int32(obj.ID))
	if entry != -1 {
		return entry
	}
	if baseabc.Abc_ObjIsCi(obj) == 0 && baseabc.Abc_ObjIsNode(obj) == 0 {
		panic("CO-set recursion expects a CI or node")
	}
	if baseabc.Abc_ObjFanoutNum(obj) == 0 {
		Vec_IntWriteEntry(mapping, int32(obj.ID), 0)
		return 0
	}
	entry = Abc_NtkDetectObjClasses_rec(baseabc.Abc_ObjFanout(obj, 0), mapping, hash, temp)
	if baseabc.Abc_ObjFanoutNum(obj) == 1 {
		Vec_IntWriteEntry(mapping, int32(obj.ID), entry)
		return entry
	}
	set := Vec_IntAlloc(16)
	vec.Vec_IntAppend(set, vec.Hsh_VecReadEntry(hash, entry))
	for i := int32(1); i < baseabc.Abc_ObjFanoutNum(obj); i++ {
		entry = Abc_NtkDetectObjClasses_rec(baseabc.Abc_ObjFanout(obj, i), mapping, hash, temp)
		vec.Vec_IntTwoMerge2(set, vec.Hsh_VecReadEntry(hash, entry), temp)
		*set, *temp = *temp, *set
	}
	entry = vec.Hsh_VecManAdd(hash, set)
	Vec_IntWriteEntry(mapping, int32(obj.ID), entry)
	Vec_IntFree(set)
	return entry
}

func Abc_NtkDetectObjClasses(network *Abc_Ntk_t, objects *Vec_Int_t, coSetsOut **Vec_Wec_t) *Vec_Wec_t {
	if network == nil || objects == nil || baseabc.Abc_NtkIsLogic(network) == 0 {
		panic("object-class detection requires a logic network and object IDs")
	}
	hash := vec.Hsh_VecManStart(1000)
	mapping := vec.Vec_IntStartFull(int(baseabc.Abc_NtkObjNumMax(network)))
	empty := Vec_IntAlloc(16)
	if id := vec.Hsh_VecManAdd(hash, empty); id != 0 {
		panic("empty CO set did not receive ID zero")
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(network); i++ {
		co := baseabc.Abc_NtkCo(network, i)
		vec.Vec_IntClear(empty)
		Vec_IntPush(empty, int32(co.ID))
		setID := vec.Hsh_VecManAdd(hash, empty)
		Vec_IntWriteEntry(mapping, int32(co.ID), setID)
	}
	vec.Vec_IntSort(objects, 0)
	temp := Vec_IntAlloc(16)
	for _, id := range vec.Vec_IntArray(objects) {
		obj := baseabc.Abc_NtkObj(network, id)
		if obj != nil {
			Abc_NtkDetectObjClasses_rec(obj, mapping, hash, temp)
		}
	}
	classMap := vec.Vec_IntStartFull(vec.Hsh_VecSize(hash) + 1)
	classes := vec.Vec_WecAlloc(1000)
	for _, id := range vec.Vec_IntArray(objects) {
		setID := Vec_IntEntry(mapping, id)
		if setID < 0 {
			panic("object did not receive a CO-set ID")
		}
		classID := Vec_IntEntry(classMap, setID)
		var class *Vec_Int_t
		if classID == -1 {
			Vec_IntWriteEntry(classMap, setID, int32(vec.Vec_WecSize(classes)))
			class = vec.Vec_WecPushLevel(classes)
		} else {
			class = vec.Vec_WecEntry(classes, int(classID))
		}
		Vec_IntPush(class, id)
	}
	coSets := vec.Vec_WecStart(vec.Vec_WecSize(classes))
	for i, class := range vec.Vec_WecArray(classes) {
		if Vec_IntSize(&class) == 0 {
			continue
		}
		id := Vec_IntEntry(&class, 0)
		setID := Vec_IntEntry(mapping, id)
		vec.Vec_IntAppend(vec.Vec_WecEntry(coSets, i), vec.Hsh_VecReadEntry(hash, setID))
	}
	vec.Hsh_VecManStop(hash)
	Vec_IntFree(classMap)
	Vec_IntFree(mapping)
	Vec_IntFree(empty)
	Vec_IntFree(temp)
	*coSetsOut = coSets
	return classes
}

func Abc_NtkFinMiterCollect_rec(obj *Abc_Obj_t, cis, nodes *Vec_Int_t) {
	if baseabc.Abc_NodeIsTravIdCurrent(obj) != 0 {
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(obj)
	if baseabc.Abc_ObjIsCi(obj) != 0 {
		Vec_IntPush(cis, int32(obj.ID))
		return
	}
	if baseabc.Abc_ObjIsNode(obj) == 0 {
		return
	}
	for _, edge := range obj.Fanins() {
		Abc_NtkFinMiterCollect_rec(edge.Object, cis, nodes)
	}
	Vec_IntPush(nodes, int32(obj.ID))
}

func Abc_NtkFinMiterCollect(network *Abc_Ntk_t, coIDs, cis, nodes *Vec_Int_t) {
	vec.Vec_IntClear(cis)
	vec.Vec_IntClear(nodes)
	baseabc.Abc_NtkIncrementTravId(network)
	for _, id := range vec.Vec_IntArray(coIDs) {
		co := baseabc.Abc_NtkObj(network, id)
		if co != nil && baseabc.Abc_ObjFaninNum(co) > 0 {
			Abc_NtkFinMiterCollect_rec(baseabc.Abc_ObjFanin0(co), cis, nodes)
		}
	}
}

func Abc_ObjFinGateType(node *Abc_Obj_t) int32 {
	sop, _ := node.Data.(string)
	switch sop {
	case "1 1\n":
		return int32(ABC_FIN_RDOB_BUFF)
	case "0 1\n":
		return int32(ABC_FIN_RDOB_NOT)
	case "11 1\n":
		return int32(ABC_FIN_RDOB_AND)
	case "11 0\n":
		return int32(ABC_FIN_RDOB_NAND)
	case "00 0\n":
		return int32(ABC_FIN_RDOB_OR)
	case "00 1\n":
		return int32(ABC_FIN_RDOB_NOR)
	case "01 1\n10 1\n":
		return int32(ABC_FIN_RDOB_XOR)
	case "11 1\n00 1\n":
		return int32(ABC_FIN_RDOB_NXOR)
	default:
		return int32(ABC_FIN_NONE)
	}
}

func Abc_NtkFinCheckTypesOk(network *Abc_Ntk_t) int32 {
	for id := int32(0); id < baseabc.Abc_NtkObjNumMax(network); id++ {
		obj := baseabc.Abc_NtkObj(network, id)
		if obj != nil && baseabc.Abc_ObjIsNode(obj) != 0 && Abc_ObjFinGateType(obj) == int32(ABC_FIN_NONE) {
			return id
		}
	}
	return 0
}

func Abc_NtkFinCheckTypesOk2(network *Abc_Ntk_t) int32 {
	library := finLibrary(network)
	fins := baseabc.Abc_NtkReadFins(network)
	if library == nil || fins == nil {
		return 0
	}
	for i := int32(2); i+1 < Vec_IntSize(fins); i += 2 {
		id, kind := Vec_IntEntry(fins, i), Vec_IntEntry(fins, i+1)
		if kind < 0 {
			continue
		}
		obj := baseabc.Abc_NtkObj(network, id)
		fault := mio.Mio_LibraryReadGateById(library, kind)
		gate, _ := obj.Data.(*mio.Mio_Gate_t)
		if fault == nil || gate == nil || mio.Mio_GateReadPinNum(fault) >= 1 && mio.Mio_GateReadPinNum(gate) != mio.Mio_GateReadPinNum(fault) {
			return id
		}
	}
	return 0
}

func Abc_NtkFinComputeTypes(network *Abc_Ntk_t) *Vec_Int_t {
	types := vec.Vec_IntStart(int(baseabc.Abc_NtkObjNumMax(network)))
	for id := int32(0); id < baseabc.Abc_NtkObjNumMax(network); id++ {
		obj := baseabc.Abc_NtkObj(network, id)
		if obj != nil && baseabc.Abc_ObjIsNode(obj) != 0 {
			Vec_IntWriteEntry(types, id, Abc_ObjFinGateType(obj))
		}
	}
	return types
}

func Abc_NtkFinComputeObjects(pairs *Vec_Int_t, mapOut **Vec_Wec_t, objectCount int32) *Vec_Int_t {
	objects := Vec_IntAlloc(100)
	*mapOut = vec.Vec_WecStart(int(objectCount))
	for i := int32(2); i+1 < Vec_IntSize(pairs); i += 2 {
		id := Vec_IntEntry(pairs, i)
		Vec_IntPush(objects, id)
		vec.Vec_WecPush(*mapOut, int(id), i/2)
	}
	vec.Vec_IntSort(objects, 0)
	vec.Vec_IntUniqify(objects)
	return objects
}

func Abc_NtkFinCreateList(mapping *Vec_Wec_t, class *Vec_Int_t) *Vec_Int_t {
	list := Vec_IntAlloc(100)
	for _, id := range vec.Vec_IntArray(class) {
		vec.Vec_IntAppend(list, vec.Vec_WecEntry(mapping, int(id)))
	}
	return list
}

func Abc_NtkDetectClassesTest2(network *Abc_Ntk_t, verbose, veryVerbose int32) {
	objects := Vec_IntAlloc(baseabc.Abc_NtkCiNum(network))
	for i := int32(0); i < baseabc.Abc_NtkCiNum(network); i++ {
		Vec_IntPush(objects, int32(baseabc.Abc_NtkCi(network, i).ID))
	}
	classes := Abc_NtkDetectObjClasses(network, objects, new(*Vec_Wec_t))
	if verbose != 0 || veryVerbose != 0 {
		vec.Vec_WecPrint(classes, 0)
	}
	Vec_IntFree(objects)
	vec.Vec_WecFree(classes)
}

// Fault-class helpers use the source's one-based fault indices (pair zero is
// the sentinel reserved by Io_ReadFins).
func Abc_NtkFinCountPairs(classes *Vec_Wec_t) int32 {
	count := 0
	for i := 0; i < vec.Vec_WecSize(classes); i++ {
		count += vec.Vec_IntSize(vec.Vec_WecEntry(classes, i)) - 1
	}
	return int32(count)
}

// finGiaLit builds the source gate's Gia expression from the per-variant
// fanin literals. The expression format is the same variable/complement AIG
// encoding consumed by Mio_LibGateSimulateGia in ABC.
func finGiaLit(manager *gia.Gia_Man_t, library *mio.Mio_Library_t, node *Abc_Obj_t, gateType int32, inputs []int32) int32 {
	if gateType == int32(ABC_FIN_SA0) {
		return gia.Gia_ManConst0Lit()
	}
	if gateType == int32(ABC_FIN_SA1) {
		return gia.Gia_ManConst1Lit()
	}
	if gateType == int32(ABC_FIN_NEG) {
		if len(inputs) == 0 {
			panic("NEG fault requires an existing object function")
		}
		return inputs[0] ^ 1
	}
	if gateType >= 0 {
		gate := mio.Mio_LibraryReadGateById(library, gateType)
		if gate == nil {
			panic("mapped fault simulation gate id is absent from the library")
		}
		expression := mio.Mio_GateReadExpr(gate)
		if mio.Exp_IsConst0(expression) != 0 {
			return gia.Gia_ManConst0Lit()
		}
		if mio.Exp_IsConst1(expression) != 0 {
			return gia.Gia_ManConst1Lit()
		}
		readLit := func(lit int32, values []int32) int32 {
			value := values[lit>>1]
			if lit&1 != 0 {
				value ^= 1
			}
			return value
		}
		if mio.Exp_IsLit(expression) != 0 {
			return readLit(Vec_IntEntry(expression, 0), inputs)
		}
		values := append([]int32(nil), inputs...)
		for i := int32(0); i < mio.Exp_NodeNum(expression); i++ {
			left := readLit(Vec_IntEntry(expression, 2*i), values)
			right := readLit(Vec_IntEntry(expression, 2*i+1), values)
			values = append(values, gia.Gia_ManHashAnd(manager, left, right))
		}
		return readLit(Vec_IntEntry(expression, Vec_IntSize(expression)-1), values)
	}
	switch Abc_FinType_t(gateType) {
	case ABC_FIN_RDOB_BUFF:
		return inputs[0]
	case ABC_FIN_RDOB_NOT:
		return inputs[0] ^ 1
	case ABC_FIN_RDOB_AND:
		return gia.Gia_ManHashAnd(manager, inputs[0], inputs[1])
	case ABC_FIN_RDOB_NAND:
		return gia.Gia_ManHashAnd(manager, inputs[0], inputs[1]) ^ 1
	case ABC_FIN_RDOB_OR:
		return gia.Gia_ManHashOr(manager, inputs[0], inputs[1])
	case ABC_FIN_RDOB_NOR:
		return gia.Gia_ManHashOr(manager, inputs[0], inputs[1]) ^ 1
	case ABC_FIN_RDOB_XOR:
		return gia.Gia_ManHashXor(manager, inputs[0], inputs[1])
	case ABC_FIN_RDOB_NXOR:
		return gia.Gia_ManHashXor(manager, inputs[0], inputs[1]) ^ 1
	default:
		panic(fmt.Sprintf("unsupported fault/gate type %d", gateType))
	}
}

func finNodeLiteral(manager *gia.Gia_Man_t, library *mio.Mio_Library_t, node *Abc_Obj_t, types *Vec_Int_t, literals []int32, faultObject int32, faultType int32) int32 {
	inputs := make([]int32, 0, baseabc.Abc_ObjFaninNum(node))
	for _, edge := range node.Fanins() {
		lit := literals[baseabc.Abc_ObjId(edge.Object)]
		if edge.Complement {
			lit ^= 1
		}
		inputs = append(inputs, lit)
	}
	normalType := faultType
	if baseabc.Abc_ObjId(node) != faultObject || faultType == int32(ABC_FIN_NEG) {
		if library != nil {
			gate, _ := node.Data.(*mio.Mio_Gate_t)
			normalType = mio.Mio_GateReadCell(gate)
		} else {
			normalType = Vec_IntEntry(types, baseabc.Abc_ObjId(node))
		}
	}
	if baseabc.Abc_ObjId(node) == faultObject && faultType == int32(ABC_FIN_NEG) {
		return finGiaLit(manager, library, node, normalType, inputs) ^ 1
	}
	return finGiaLit(manager, library, node, normalType, inputs)
}

func finSimGate(kind int32, fanins []bool, library *mio.Mio_Library_t) bool {
	if kind >= 0 {
		gate := mio.Mio_LibraryReadGateById(library, kind)
		if gate == nil {
			panic("mapped fault gate is absent from library")
		}
		index := 0
		for i, value := range fanins {
			if value {
				index |= 1 << i
			}
		}
		truth := mio.Mio_GateReadTruthP(gate)
		return (truth[index>>6]>>(uint(index)&63))&1 != 0
	}
	switch Abc_FinType_t(kind) {
	case ABC_FIN_RDOB_BUFF:
		return fanins[0]
	case ABC_FIN_RDOB_NOT:
		return !fanins[0]
	case ABC_FIN_RDOB_AND:
		return fanins[0] && fanins[1]
	case ABC_FIN_RDOB_NAND:
		return !(fanins[0] && fanins[1])
	case ABC_FIN_RDOB_OR:
		return fanins[0] || fanins[1]
	case ABC_FIN_RDOB_NOR:
		return !(fanins[0] || fanins[1])
	case ABC_FIN_RDOB_XOR:
		return fanins[0] != fanins[1]
	case ABC_FIN_RDOB_NXOR:
		return fanins[0] == fanins[1]
	default:
		panic(fmt.Sprintf("unsupported simulated gate type %d", kind))
	}
}

// finSimOutputs performs the source detector's all-zero/all-one pattern
// pre-refinement for one fault and its structural CO set.
func finSimOutputs(network *Abc_Ntk_t, coSet, ciSet, pairs *Vec_Int_t, fault int32, pattern *Vec_Int_t) []bool {
	faultObject, faultType := Vec_IntEntry(pairs, 2*fault), Vec_IntEntry(pairs, 2*fault+1)
	values := make([]bool, int(baseabc.Abc_NtkObjNumMax(network)))
	for i := int32(0); i < Vec_IntSize(ciSet); i++ {
		ci := baseabc.Abc_NtkObj(network, Vec_IntEntry(ciSet, i))
		value := Vec_IntEntry(pattern, i) != 0
		if baseabc.Abc_ObjId(ci) == faultObject {
			switch Abc_FinType_t(faultType) {
			case ABC_FIN_SA0:
				value = false
			case ABC_FIN_SA1:
				value = true
			case ABC_FIN_NEG:
				value = !value
			default:
				panic("non-stuck-at fault on a CI")
			}
		}
		values[baseabc.Abc_ObjId(ci)] = value
	}
	library := finLibrary(network)
	for _, node := range network.ObjectsOfKind(baseabc.ObjectNode) {
		id := baseabc.Abc_ObjId(node)
		fanins := make([]bool, 0, baseabc.Abc_ObjFaninNum(node))
		for _, edge := range node.Fanins() {
			value := values[baseabc.Abc_ObjId(edge.Object)]
			if edge.Complement {
				value = !value
			}
			fanins = append(fanins, value)
		}
		kind := int32(Abc_ObjFinGateType(node))
		if library != nil {
			gate, _ := node.Data.(*mio.Mio_Gate_t)
			kind = mio.Mio_GateReadCell(gate)
		}
		value := finSimGate(kind, fanins, library)
		if id == faultObject {
			switch Abc_FinType_t(faultType) {
			case ABC_FIN_SA0:
				value = false
			case ABC_FIN_SA1:
				value = true
			case ABC_FIN_NEG:
				value = !value
			default:
				value = finSimGate(faultType, fanins, library)
			}
		}
		values[id] = value
	}
	outputs := make([]bool, Vec_IntSize(coSet))
	for i := int32(0); i < Vec_IntSize(coSet); i++ {
		co := baseabc.Abc_NtkObj(network, Vec_IntEntry(coSet, i))
		edge := co.Fanins()[0]
		outputs[i] = values[baseabc.Abc_ObjId(edge.Object)] != edge.Complement
	}
	return outputs
}

func finSameBits(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type finClassItem struct {
	fault int32
	index int32
}

// finRefineByPattern applies one SAT counterexample or random simulation
// pattern to the source worklist. It splits each class into the members that
// agree with its first member and one appended class of mismatches.
func finRefineByPattern(network *Abc_Ntk_t, coSet, ciSet, pairs *Vec_Int_t, classes *[][]finClassItem, pattern *Vec_Int_t, level, item int) {
	cache := make(map[int32][]bool)
	outputs := func(fault int32) []bool {
		if value, ok := cache[fault]; ok {
			return value
		}
		value := finSimOutputs(network, coSet, ciSet, pairs, fault, pattern)
		cache[fault] = value
		return value
	}
	for i := 0; i < level; i++ {
		class := (*classes)[i]
		if len(class) < 2 {
			continue
		}
		reference := outputs(class[0].fault)
		for _, member := range class[1:] {
			if !finSameBits(reference, outputs(member.fault)) {
				panic("SAT counterexample refined an already-proved fault class")
			}
		}
	}
	for i := level; i < len(*classes); i++ {
		class := (*classes)[i]
		if len(class) < 2 {
			continue
		}
		start := 1
		if i == level {
			start = item
		}
		if start >= len(class) {
			continue
		}
		reference := outputs(class[0].fault)
		for _, member := range class[1:start] {
			if !finSameBits(reference, outputs(member.fault)) {
				panic("SAT counterexample refined a previously retained fault prefix")
			}
		}
		kept := append([]finClassItem(nil), class[:start]...)
		mismatches := make([]finClassItem, 0)
		for _, candidate := range class[start:] {
			if finSameBits(reference, outputs(candidate.fault)) {
				kept = append(kept, candidate)
			} else {
				mismatches = append(mismatches, candidate)
			}
		}
		(*classes)[i] = kept
		if len(mismatches) != 0 {
			*classes = append(*classes, mismatches)
		}
	}
}

// Abc_NtkFinCheckPair builds and solves the source pairwise XOR miter. A nil
// result means proved equivalent; an empty vector means SAT resource limit;
// otherwise the vector contains a counterexample over the window CIs.
func Abc_NtkFinCheckPair(network *Abc_Ntk_t, types, coSet, ciSet, nodeSet *Vec_Int_t, objects, faultTypes [2]int32, scratch *Vec_Int_t) *Vec_Int_t {
	_ = nodeSet
	_ = scratch
	firstObject, firstType := objects[0], faultTypes[0]
	secondObject, secondType := objects[1], faultTypes[1]
	manager := gia.Gia_ManStart(1000)
	gia.Gia_ManHashStart(manager)
	library := finLibrary(network)
	firstLits := make([]int32, int(baseabc.Abc_NtkObjNumMax(network)))
	secondLits := make([]int32, len(firstLits))
	for i := range firstLits {
		firstLits[i], secondLits[i] = -1, -1
	}
	for _, id := range vec.Vec_IntArray(ciSet) {
		ci := baseabc.Abc_NtkObj(network, id)
		objID := baseabc.Abc_ObjId(ci)
		lit := gia.Gia_ManAppendCi(manager)
		firstLits[objID], secondLits[objID] = lit, lit
		if objID == firstObject {
			if firstType == int32(ABC_FIN_NEG) {
				firstLits[objID] = lit ^ 1
			} else {
				firstLits[objID] = finGiaLit(manager, library, ci, firstType, nil)
			}
		}
		if objID == secondObject {
			if secondType == int32(ABC_FIN_NEG) {
				secondLits[objID] = lit ^ 1
			} else {
				secondLits[objID] = finGiaLit(manager, library, ci, secondType, nil)
			}
		}
	}
	for _, id := range vec.Vec_IntArray(nodeSet) {
		node := baseabc.Abc_NtkObj(network, id)
		objID := baseabc.Abc_ObjId(node)
		var typ int32
		if baseabc.Abc_NtkIsMappedLogic(network) != 0 {
			gate, _ := node.Data.(*mio.Mio_Gate_t)
			typ = mio.Mio_GateReadCell(gate)
		} else {
			typ = Vec_IntEntry(types, objID)
		}
		firstTypeHere, secondTypeHere := typ, typ
		if objID == firstObject {
			firstTypeHere = firstType
		}
		if objID == secondObject {
			secondTypeHere = secondType
		}
		firstLits[objID] = finNodeLiteral(manager, library, node, types, firstLits, firstObject, firstTypeHere)
		secondLits[objID] = finNodeLiteral(manager, library, node, types, secondLits, secondObject, secondTypeHere)
	}
	var mismatch int32
	for i := int32(0); i < Vec_IntSize(coSet); i++ {
		co := baseabc.Abc_NtkObj(network, Vec_IntEntry(coSet, i))
		edge := co.Fanins()[0]
		a, b := firstLits[baseabc.Abc_ObjId(edge.Object)], secondLits[baseabc.Abc_ObjId(edge.Object)]
		if edge.Complement {
			a, b = a^1, b^1
		}
		mismatch = gia.Gia_ManHashOr(manager, mismatch, gia.Gia_ManHashXor(manager, a, b))
	}
	gia.Gia_ManAppendCo(manager, mismatch)
	clean := gia.Gia_ManCleanup(manager)
	gia.Gia_ManStop(manager)
	defer gia.Gia_ManStop(clean)
	if gia.Gia_ManAndNum(clean) == 0 && gia.Gia_ObjIsConst0(gia.Gia_ObjFanin0(gia.Gia_ManCo(clean, 0))) != 0 {
		if gia.Gia_ObjFaninC0(gia.Gia_ManCo(clean, 0)) == 0 {
			return nil
		}
		pattern := Vec_IntAlloc(Vec_IntSize(ciSet))
		for i := int32(0); i < Vec_IntSize(ciSet); i++ {
			Vec_IntPush(pattern, 0)
		}
		return pattern
	}
	generated := gia.Mf_ManGenerateCnf(clean, 8, 0, 1, 0, 0)
	if generated == nil {
		return Vec_IntAlloc(0)
	}
	cnfData := cnf.Cnf_DataAlloc(nil, int32(generated.Variables), int32(len(generated.Clauses)), 0)
	cnfData.Clauses = generated.Clauses
	solver := cnf.Cnf_DataWriteIntoSolver(cnfData, 1, 0)
	cnf.Cnf_DataFree(cnfData)
	if solver == nil {
		return nil // root propagation proved the miter contradictory
	}
	status := solver.SolveWithResourceLimits(nil, 10000, 0, 0, 0)
	if status == bsat.Unsatisfiable {
		return nil
	}
	if status != bsat.Satisfiable {
		return Vec_IntAlloc(0)
	}
	pattern := Vec_IntAlloc(Vec_IntSize(ciSet))
	begin := int32(generated.Variables) - int32(gia.Gia_ManPiNum(clean))
	for i := int32(0); i < Vec_IntSize(ciSet); i++ {
		value, ok := solver.Value(begin + i)
		if !ok {
			return Vec_IntAlloc(0)
		}
		if value {
			Vec_IntPush(pattern, 1)
		} else {
			Vec_IntPush(pattern, 0)
		}
	}
	return pattern
}

// finRefineFaultClass mirrors Abc_NtkFinRefinement's mutable class worklist.
// It returns the exact SAT-call count along with the refined source-order
// classes, whose members are one-based indices into vFins.
func finRefineFaultClass(network *Abc_Ntk_t, types, coSet, ciSet, nodeSet, pairs, faultList *Vec_Int_t) ([][]finClassItem, int) {
	classes := [][]finClassItem{{}}
	for i := int32(0); i < Vec_IntSize(faultList); i++ {
		classes[0] = append(classes[0], finClassItem{fault: Vec_IntEntry(faultList, i), index: i})
	}
	for _, bit := range []int32{0, 1} {
		pattern := Vec_IntAlloc(Vec_IntSize(ciSet))
		for i := int32(0); i < Vec_IntSize(ciSet); i++ {
			Vec_IntPush(pattern, bit)
		}
		finRefineByPattern(network, coSet, ciSet, pairs, &classes, pattern, 0, 1)
		Vec_IntFree(pattern)
	}
	nCalls := 0
	for i := 0; i < len(classes); i++ {
		for k := 1; k < len(classes[i]); {
			class := classes[i]
			first, candidate := class[0].fault, class[k].fault
			objects := [2]int32{Vec_IntEntry(pairs, 2*first), Vec_IntEntry(pairs, 2*candidate)}
			faultTypes := [2]int32{Vec_IntEntry(pairs, 2*first+1), Vec_IntEntry(pairs, 2*candidate+1)}
			pattern := Abc_NtkFinCheckPair(network, types, coSet, ciSet, nodeSet, objects, faultTypes, nil)
			nCalls++
			if pattern == nil {
				k++
				continue
			}
			if Vec_IntSize(pattern) == 0 {
				moved := class[k]
				classes[i] = append(class[:k], class[k+1:]...)
				classes = append(classes, []finClassItem{moved})
				Vec_IntFree(pattern)
			} else {
				finRefineByPattern(network, coSet, ciSet, pairs, &classes, pattern, i, k)
				Vec_IntFree(pattern)
				if k < len(classes[i]) && classes[i][k].fault == candidate {
					panic("SAT counterexample did not refine the compared fault")
				}
			}
		}
	}
	return classes, nCalls
}

func Abc_NtkDetectFinClasses(network *Abc_Ntk_t, verbose int32) *Vec_Wec_t {
	fins := baseabc.Abc_NtkReadFins(network)
	if fins == nil {
		fmt.Println("Current network does not have the required info.")
		return nil
	}
	if baseabc.Abc_NtkIsSopLogic(network) == 0 && baseabc.Abc_NtkIsMappedLogic(network) == 0 {
		panic("Abc_NtkDetectFinClasses requires SOP or mapped logic")
	}
	var types *Vec_Int_t
	if baseabc.Abc_NtkIsSopLogic(network) != 0 {
		if id := Abc_NtkFinCheckTypesOk(network); id != 0 {
			fmt.Printf("Current network contains unsupported gate types (for example, see node %q).\n", baseabc.Abc_ObjName(baseabc.Abc_NtkObj(network, id)))
			return nil
		}
		types = Abc_NtkFinComputeTypes(network)
	} else if id := Abc_NtkFinCheckTypesOk2(network); id != 0 {
		fmt.Printf("Current network has mismatch between mapped gate size and fault gate size (for example, see node %q).\n", baseabc.Abc_ObjName(baseabc.Abc_NtkObj(network, id)))
		return nil
	}
	if types != nil {
		defer Vec_IntFree(types)
	}

	// First group fault sites by their structural TFO CO set. Faults at the
	// same object remain distinct entries and are refined by exact equivalence.
	var mapping *Vec_Wec_t
	objects := Abc_NtkFinComputeObjects(fins, &mapping, baseabc.Abc_NtkObjNumMax(network))
	defer Vec_IntFree(objects)
	defer vec.Vec_WecFree(mapping)
	var coSets *Vec_Wec_t
	objectClasses := Abc_NtkDetectObjClasses(network, objects, &coSets)
	defer vec.Vec_WecFree(objectClasses)
	defer vec.Vec_WecFree(coSets)
	results := vec.Vec_WecAlloc(1000)
	for classID := 0; classID < vec.Vec_WecSize(objectClasses); classID++ {
		objectClass := vec.Vec_WecEntry(objectClasses, classID)
		coSet := vec.Vec_WecEntry(coSets, classID)
		faultList := Abc_NtkFinCreateList(mapping, objectClass)
		ciSet, nodeSet := vec.Vec_IntAlloc(1000), vec.Vec_IntAlloc(1000)
		Abc_NtkFinMiterCollect(network, coSet, ciSet, nodeSet)

		classes, nCalls := finRefineFaultClass(network, types, coSet, ciSet, nodeSet, fins, faultList)
		for _, class := range classes {
			if len(class) < 2 {
				continue
			}
			level := vec.Vec_WecPushLevel(results)
			for _, item := range class {
				vec.Vec_IntPush(level, item.fault)
			}
		}
		if verbose != 0 {
			fmt.Printf("Group %4d : Obj =%4d. Fins =%4d. CI =%5d. CO =%5d. Node =%6d. SAT calls =%5d.\n", classID, Vec_IntSize(objectClass), Vec_IntSize(faultList), Vec_IntSize(ciSet), Vec_IntSize(coSet), Vec_IntSize(nodeSet), nCalls)
		}
		Vec_IntFree(faultList)
		Vec_IntFree(ciSet)
		Vec_IntFree(nodeSet)
	}
	for i := 0; i < vec.Vec_WecSize(results); i++ {
		vec.Vec_IntSort(vec.Vec_WecEntry(results, i), 0)
	}
	vec.Vec_WecSortByFirstInt(results, 0)
	return results
}

func Abc_NtkPrintFinResults(classes *Vec_Wec_t) {
	for i := 0; i < vec.Vec_WecSize(classes); i++ {
		class := vec.Vec_WecEntry(classes, i)
		if Vec_IntSize(class) == 0 {
			continue
		}
		first := Vec_IntEntry(class, 0)
		for j := int32(1); j < Vec_IntSize(class); j++ {
			fmt.Printf("%d %d\n", first, Vec_IntEntry(class, j))
		}
	}
}

func Abc_NtkDetectClassesTest(network *Abc_Ntk_t, sequential, verbose, veryVerbose int32) {
	if sequential != 0 {
		Abc_NtkFrameExtend(network)
	}
	classes := Abc_NtkDetectFinClasses(network, verbose)
	if classes == nil {
		return
	}
	fmt.Printf("Computed %d equivalence classes with %d item pairs.\n", vec.Vec_WecSize(classes), Abc_NtkFinCountPairs(classes))
	if veryVerbose != 0 {
		vec.Vec_WecPrint(classes, 1)
	}
	vec.Vec_WecFree(classes)
}

func Abc_NtkFrameExtend(network *Abc_Ntk_t) {
	// Mirror the ABC detector's two-frame expansion for networks whose final
	// inputs and outputs encode latch reset, enable, and signal functions.
	nPisOld := baseabc.Abc_NtkPiNum(network)
	nPosOld := baseabc.Abc_NtkPoNum(network)
	nConstr := network.Constraints
	if nConstr == 0 {
		return
	}
	if nConstr < 0 || nPisOld < nConstr || nPosOld < 4*nConstr {
		panic("Abc_NtkFrameExtend: malformed sequential network")
	}
	nodes := append([]*Abc_Obj_t(nil), network.ObjectsOfKind(baseabc.ObjectNode)...)
	pis := append([]*Abc_Obj_t(nil), network.ObjectsOfKind(baseabc.ObjectPI)...)
	pos := append([]*Abc_Obj_t(nil), network.ObjectsOfKind(baseabc.ObjectPO)...)
	// Duplicate ordinary primary inputs. For each latch input, build the
	// next-state function !(reset) & ((enable & signal) | (!enable & state)).
	for i, pi := range pis {
		if int32(i) < nPisOld-nConstr {
			copy := baseabc.Abc_NtkDupObj(network, pi, 0)
			baseabc.Abc_ObjAssignName(copy, baseabc.Abc_ObjName(pi), "_frame1")
			pi.SetCopyEdge(copy.Signal())
			continue
		}
		start := int(nPosOld + 4*(int32(i)-nPisOld))
		reset := pos[start+1].Fanins()[0].Object
		enable := pos[start+2].Fanins()[0].Object
		signal := pos[start+3].Fanins()[0].Object
		resetN := baseabc.Abc_NtkCreateNodeInv(network, reset)
		enableN := baseabc.Abc_NtkCreateNodeInv(network, enable)
		and0 := baseabc.Abc_NtkCreateNodeAnd(network, []*Abc_Obj_t{enableN, pi})
		and1 := baseabc.Abc_NtkCreateNodeAnd(network, []*Abc_Obj_t{enable, signal})
		mux := baseabc.Abc_NtkCreateNodeOr(network, []*Abc_Obj_t{and0, and1})
		pi.SetCopyEdge(baseabc.Abc_NtkCreateNodeAnd(network, []*Abc_Obj_t{resetN, mux}).Signal())
	}
	// Duplicate the old logic and connect it through the second-frame copies.
	for _, node := range nodes {
		copy := baseabc.Abc_NtkDupObj(network, node, 0)
		node.SetCopyEdge(copy.Signal())
	}
	for _, node := range nodes {
		copy := node.ReadCopyEdge().Object
		for _, edge := range node.Fanins() {
			mapped := edge.Object.ReadCopyEdge()
			if edge.Complement {
				mapped = mapped.Not()
			}
			baseabc.Abc_ObjAddFanin(copy, mapped)
		}
	}
	// Duplicate ordinary outputs and reconnect encoded latch-input outputs to
	// the copies of their original drivers.
	for i, po := range pos {
		if int32(i) < nPosOld-4*nConstr {
			copy := baseabc.Abc_NtkDupObj(network, po, 0)
			baseabc.Abc_ObjAssignName(copy, baseabc.Abc_ObjName(po), "_frame1")
			old := po.Fanins()[0]
			mapped := old.Object.ReadCopyEdge()
			if old.Complement {
				mapped = mapped.Not()
			}
			baseabc.Abc_ObjAddFanin(copy, mapped)
			continue
		}
		old := po.Fanins()[0]
		mapped := old.Object.ReadCopyEdge()
		if old.Complement {
			mapped = mapped.Not()
		}
		baseabc.Abc_ObjPatchFanin(po, old.Object, mapped)
	}
}
