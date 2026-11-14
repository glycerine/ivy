package abc

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/misc/nm"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"sort"
	"strconv"
	"strings"
)

func Abc_ObjName(object *Object) string {
	if object == nil || object.network == nil {
		panic("name requires a live object")
	}
	return nm.Nm_ManCreateUniqueName(object.network.names, int32(object.ID))
}
func Abc_ObjAssignName(object *Object, name, suffix string) string {
	if object == nil || object.network == nil {
		panic("name requires a live object")
	}
	return nm.Nm_ManStoreIdName(object.network.names, int32(object.ID), int32(object.Kind), name, suffix)
}
func Abc_ObjDeleteName(object *Object) {
	if object == nil || object.network == nil {
		panic("name deletion requires a live object")
	}
	nm.Nm_ManDeleteIdName(object.network.names, int32(object.ID))
}
func Abc_ObjNamePrefix(object *Object, prefix string) string { return prefix + Abc_ObjName(object) }
func Abc_ObjNameSuffix(object *Object, suffix string) string { return Abc_ObjName(object) + suffix }
func Abc_ObjNameDummy(prefix string, number, digits int32) string {
	return fmt.Sprintf("%s%0*d", prefix, int(digits), number)
}

func nameTransferTerminal(object *Object, output bool) *Object {
	if Abc_NtkIsNetlist(object.network) == 0 {
		return object
	}
	if output {
		if len(object.fanins) != 1 {
			panic("netlist output has no net")
		}
		return object.fanins[0].Object
	}
	if len(object.fanouts) == 0 {
		panic("netlist input has no net")
	}
	return object.fanouts[0]
}
func transferTerminalName(object *Object, output bool) {
	if object.Copy == nil {
		panic("name transfer requires object copy")
	}
	Abc_ObjAssignName(object.Copy, Abc_ObjName(nameTransferTerminal(object, output)), "")
}
func Abc_NtkTrasferNames(old, new *Network) {
	if Abc_NtkPiNum(old) != Abc_NtkPiNum(new) || Abc_NtkPoNum(old) != Abc_NtkPoNum(new) || Abc_NtkBoxNum(old) != Abc_NtkBoxNum(new) {
		panic("network interface mismatch")
	}
	for _, object := range old.cis {
		transferTerminalName(object, false)
	}
	for _, object := range old.cos {
		transferTerminalName(object, true)
	}
	for _, box := range old.boxes {
		if box.Copy == nil {
			panic("name transfer requires box copy")
		}
		Abc_ObjAssignName(box.Copy, Abc_ObjName(box), "")
	}
}
func Abc_NtkTrasferNamesNoLatches(old, new *Network) {
	if Abc_NtkPiNum(old) != Abc_NtkPiNum(new) || Abc_NtkPoNum(old) != Abc_NtkPoNum(new) {
		panic("network interface mismatch")
	}
	for _, object := range old.cis {
		if len(object.fanins) == 0 || object.fanins[0].Object.Kind != ObjectLatch {
			transferTerminalName(object, false)
		}
	}
	for _, object := range old.cos {
		if len(object.fanouts) == 0 || object.fanouts[0].Kind != ObjectLatch {
			transferTerminalName(object, true)
		}
	}
	for _, box := range old.boxes {
		if box.Kind != ObjectLatch {
			if box.Copy == nil {
				panic("name transfer requires box copy")
			}
			Abc_ObjAssignName(box.Copy, Abc_ObjName(box), "")
		}
	}
}
func Abc_NodeGetFaninNames(node *Object) *vec.Vec_Ptr_t {
	names := vec.Vec_PtrAlloc(len(node.fanins))
	for _, edge := range node.fanins {
		vec.Vec_PtrPush(names, Abc_ObjName(edge.Object))
	}
	return names
}
func Abc_NodeGetFakeNames(count int32) *vec.Vec_Ptr_t {
	if count < 0 {
		panic("negative name count")
	}
	names := vec.Vec_PtrAlloc(int(count))
	for i := int32(0); i < count; i++ {
		if count < 26 {
			vec.Vec_PtrPush(names, string([]byte{byte('a' + i)}))
		} else {
			vec.Vec_PtrPush(names, string([]byte{byte('a' + i%26), byte('0' + i/26)}))
		}
	}
	return names
}
func Abc_NodeFreeNames(names *vec.Vec_Ptr_t) {
	if names != nil {
		vec.Vec_PtrFree(names)
	}
}
func Abc_NtkCollectCioNames(n *Network, outputs int32) []string {
	terminals := n.cis
	if outputs != 0 {
		terminals = n.cos
	}
	names := make([]string, len(terminals))
	for i, object := range terminals {
		names[i] = Abc_ObjName(object)
	}
	return names
}

func objectOrderName(object *Object) string {
	if object.Kind == ObjectLatch && len(object.fanouts) > 0 {
		return Abc_ObjName(object.fanouts[0])
	}
	return Abc_ObjName(object)
}
func Abc_NodeCompareNames(first, second **Object) int32 {
	a, b := *first, *second
	if cmp := strings.Compare(objectOrderName(a), objectOrderName(b)); cmp != 0 {
		return int32(cmp)
	}
	if a.ID < b.ID {
		return -1
	}
	if a.ID > b.ID {
		return 1
	}
	return 0
}
func Abc_NtkOrderObjsByName(n *Network, combinational int32) {
	if Abc_NtkHasOnlyLatchBoxes(n) == 0 {
		panic("name ordering requires latch-only boxes")
	}
	less := func(objects []*Object) {
		sort.SliceStable(objects, func(i, j int) bool { return Abc_NodeCompareNames(&objects[i], &objects[j]) < 0 })
	}
	less(n.kindOrder[ObjectPI])
	less(n.kindOrder[ObjectPO])
	if combinational != 0 {
		less(n.boxes)
	}
	n.kindOrder[ObjectLatch] = append(n.kindOrder[ObjectLatch][:0], n.boxes...)
	Abc_NtkOrderCisCos(n)
}
func Abc_NtkNameMan(n *Network, outputs int32) *util.Abc_Nam_t {
	objects := n.cis
	if outputs != 0 {
		objects = n.cos
	}
	manager := util.Abc_NamStart(int32(len(objects)), 24)
	for _, object := range objects {
		util.Abc_NamStrFindOrAdd(manager, Abc_ObjName(object), nil)
	}
	if util.Abc_NamObjNumMax(manager) != int32(len(objects)+1) {
		panic("terminal names must be unique")
	}
	return manager
}
func Abc_NodeCompareIndexes(first, second **Object) int32 {
	if (*first).Temp < (*second).Temp {
		return -1
	}
	if (*first).Temp > (*second).Temp {
		return 1
	}
	return 0
}
func Abc_NtkTransferOrder(old, new *Network) {
	if Abc_NtkPiNum(old) != Abc_NtkPiNum(new) || Abc_NtkPoNum(old) != Abc_NtkPoNum(new) || Abc_NtkLatchNum(old) != Abc_NtkLatchNum(new) || len(old.cis) != len(new.cis) || len(old.cos) != len(new.cos) {
		panic("network interface mismatch")
	}
	ranks := func(reference, target []*Object) map[*Object]int {
		names := make(map[string]int, len(reference))
		for i, o := range reference {
			name := Abc_ObjName(o)
			if _, ok := names[name]; ok {
				panic("duplicate reference terminal name")
			}
			names[name] = i
		}
		order := make(map[*Object]int, len(target))
		seen := make(map[string]bool, len(target))
		for _, o := range target {
			name := Abc_ObjName(o)
			rank, ok := names[name]
			if !ok || seen[name] {
				panic("terminal names do not match")
			}
			seen[name] = true
			order[o] = rank
		}
		return order
	}
	ciOrder, coOrder := ranks(old.cis, new.cis), ranks(old.cos, new.cos)
	sort.SliceStable(new.cis, func(i, j int) bool { return ciOrder[new.cis[i]] < ciOrder[new.cis[j]] })
	sort.SliceStable(new.cos, func(i, j int) bool { return coOrder[new.cos[i]] < coOrder[new.cos[j]] })
	sort.SliceStable(new.kindOrder[ObjectPI], func(i, j int) bool { return ciOrder[new.kindOrder[ObjectPI][i]] < ciOrder[new.kindOrder[ObjectPI][j]] })
	sort.SliceStable(new.kindOrder[ObjectPO], func(i, j int) bool { return coOrder[new.kindOrder[ObjectPO][i]] < coOrder[new.kindOrder[ObjectPO][j]] })
	for _, kind := range []ObjectKind{ObjectBO, ObjectBI} {
		objects := new.cis
		if kind == ObjectBI {
			objects = new.cos
		}
		new.kindOrder[kind] = nil
		for _, object := range objects {
			if object.Kind == kind {
				new.kindOrder[kind] = append(new.kindOrder[kind], object)
			}
		}
	}
}
func Abc_NodeCompareCiCo(old, new *Network) int32 {
	if Abc_NtkPiNum(old) != Abc_NtkPiNum(new) || Abc_NtkPoNum(old) != Abc_NtkPoNum(new) || Abc_NtkLatchNum(old) != Abc_NtkLatchNum(new) || len(old.cis) != len(new.cis) || len(old.cos) != len(new.cos) {
		return 0
	}
	for i, object := range old.cis {
		if Abc_ObjName(object) != Abc_ObjName(new.cis[i]) {
			return 0
		}
	}
	for i, object := range old.cos {
		if Abc_ObjName(object) != Abc_ObjName(new.cos[i]) {
			return 0
		}
	}
	return 1
}

func dummyNameWidth(count int) int32 {
	if count < 2 {
		return int32(count)
	}
	return int32(len(strconv.Itoa(count - 1)))
}
func Abc_NtkAddDummyPiNames(n *Network) {
	objects := n.ObjectsOfKind(ObjectPI)
	width := dummyNameWidth(len(objects))
	for i, object := range objects {
		Abc_ObjAssignName(object, Abc_ObjNameDummy("pi", int32(i), width), "")
	}
}
func Abc_NtkAddDummyPoNames(n *Network) {
	objects := n.ObjectsOfKind(ObjectPO)
	width := dummyNameWidth(len(objects))
	for i, object := range objects {
		Abc_ObjAssignName(object, Abc_ObjNameDummy("po", int32(i), width), "")
	}
}
func Abc_NtkAddDummyBoxNames(n *Network) {
	if Abc_NtkIsNetlist(n) != 0 {
		panic("dummy latch names require a logic network")
	}
	longest := 0
	for _, kind := range []ObjectKind{ObjectPI, ObjectPO} {
		for _, object := range n.ObjectsOfKind(kind) {
			name := Abc_ObjName(object)
			count := 0
			for count < len(name) && name[count] == 'l' {
				count++
			}
			longest = max(longest, count)
		}
	}
	if longest >= 98 {
		panic("latch prefix exceeds ABC limit")
	}
	prefix := strings.Repeat("l", longest+1)
	width := dummyNameWidth(len(n.ObjectsOfKind(ObjectLatch)))
	for i, box := range n.boxes {
		if box.Kind != ObjectLatch {
			continue
		}
		if len(box.fanins) != 1 || len(box.fanouts) == 0 {
			panic("malformed latch")
		}
		Abc_ObjAssignName(box, Abc_ObjNameDummy("l", int32(i), width), "")
		Abc_ObjAssignName(box.fanins[0].Object, Abc_ObjNameDummy(prefix+"i", int32(i), width), "")
		Abc_ObjAssignName(box.fanouts[0], Abc_ObjNameDummy(prefix+"o", int32(i), width), "")
	}
}
func Abc_NtkShortNames(n *Network) {
	nm.Nm_ManFree(n.names)
	n.names = nm.Nm_ManCreate(int32(len(n.cis) + len(n.cos) + len(n.boxes)))
	Abc_NtkAddDummyPiNames(n)
	Abc_NtkAddDummyPoNames(n)
	Abc_NtkAddDummyBoxNames(n)
}
func Abc_NtkCleanNames(n *Network) {
	manager := nm.Nm_ManCreate(int32(len(n.cis) + len(n.cos) + len(n.boxes)))
	for _, objects := range [][]*Object{n.cis, n.cos} {
		for _, object := range objects {
			nm.Nm_ManStoreIdName(manager, int32(object.ID), int32(object.Kind), Abc_ObjName(object), "")
		}
	}
	nm.Nm_ManFree(n.names)
	n.names = manager
}
func Abc_NtkRedirectCiCo(n *Network) {
	count := 0
	for _, output := range n.cos {
		id := nm.Nm_ManFindIdByNameTwoTypes(n.names, Abc_ObjName(output), int32(ObjectPI), int32(ObjectBO))
		if id < 0 || len(output.fanins) != 1 {
			continue
		}
		input := n.Object(int(id))
		driver := output.fanins[0].Object
		if driver == input {
			continue
		}
		if driver == nil || Abc_NodeIsBuf(driver) == 0 || len(driver.fanins) != 1 {
			panic("same-name terminal output must have a buffer driver")
		}
		replacement := input.Signal()
		if driver.fanins[0].Complement {
			replacement = replacement.Not()
		}
		Abc_ObjPatchFanin(output, driver, replacement)
		if len(driver.fanouts) == 0 {
			Abc_NtkDeleteObj(driver)
		}
		count++
	}
	if count != 0 {
		fmt.Printf("Redirected %d POs from buffers to PIs with the same name.\n", count)
	}
}
func Abc_NtkMoveNames(n, old *Network) {
	if n == old || Abc_NtkPiNum(n) != Abc_NtkPiNum(old) || Abc_NtkPoNum(n) != Abc_NtkPoNum(old) || len(n.boxes) != len(old.boxes) {
		panic("incompatible networks for name transfer")
	}
	nm.Nm_ManFree(n.names)
	n.names = nm.Nm_ManCreate(int32(len(n.cis) + len(n.cos) + len(n.boxes)))
	for _, kind := range []ObjectKind{ObjectPI, ObjectPO} {
		for i, object := range n.ObjectsOfKind(kind) {
			Abc_ObjAssignName(object, Abc_ObjName(old.ObjectsOfKind(kind)[i]), "")
		}
	}
	for i, box := range n.boxes {
		if box.Kind == ObjectLatch {
			source := old.boxes[i]
			if source.Kind != ObjectLatch || len(box.fanins) != 1 || len(source.fanins) != 1 || len(box.fanouts) == 0 || len(source.fanouts) == 0 {
				panic("incompatible latch interfaces")
			}
			Abc_ObjAssignName(box.fanins[0].Object, Abc_ObjName(source.fanins[0].Object), "")
			Abc_ObjAssignName(box.fanouts[0], Abc_ObjName(source.fanouts[0]), "")
		}
	}
	Abc_NtkRedirectCiCo(n)
}

func abcNameIDFile(spec string) string {
	slash := strings.LastIndexAny(spec, "/\\")
	generic := spec
	if dot := strings.LastIndexByte(spec, '.'); dot > slash {
		generic = spec[:dot]
	}
	base := spec[slash+1:]
	extension := base
	if dot := strings.LastIndexByte(base, '.'); dot >= 0 {
		extension = base[dot+1:]
	}
	return generic + "_" + extension + "_names.txt"
}

func Abc_NtkStartNameIds(n *Network) {
	if Abc_NtkIsNetlist(n) == 0 || n.NameIDs != nil {
		panic("name IDs require an uninitialized netlist")
	}
	file, err := os.Create(abcNameIDFile(n.Spec))
	if err != nil {
		panic(err)
	}
	ids := make([]int32, len(n.objects))
	counter := int32(1)
	writeName := func(object, net *Object) {
		if _, err := fmt.Fprintf(file, "%s            \n", Abc_ObjName(net)); err != nil {
			panic(err)
		}
		ids[object.ID] = 2 * counter
		counter++
	}
	for _, input := range n.cis {
		if len(input.fanouts) == 0 {
			panic("netlist input has no net")
		}
		writeName(input, input.fanouts[0])
	}
	for _, output := range n.cos {
		if len(output.fanins) != 1 || len(output.fanins[0].Object.fanins) != 1 {
			panic("netlist output has no driver")
		}
		net := output.fanins[0].Object
		driver := net.fanins[0].Object
		if ids[driver.ID] == 0 {
			if len(driver.fanouts) == 0 {
				panic("netlist driver has no net")
			}
			writeName(driver, driver.fanouts[0])
		}
	}
	for _, node := range Abc_NtkDfs(n, 1) {
		if ids[node.ID] == 0 {
			if len(node.fanouts) == 0 {
				panic("netlist node has no output net")
			}
			writeName(node, node.fanouts[0])
		}
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
	for _, output := range n.cos {
		driver := output.fanins[0].Object.fanins[0].Object
		ids[output.ID] = ids[driver.ID]
		ids[driver.ID] = 0
	}
	n.NameIDs = ids
}

func Abc_NtkTransferNameIds(old, new *Network) {
	if old.NameIDs == nil || new.NameIDs != nil {
		panic("invalid name ID transfer state")
	}
	new.NameIDs = make([]int32, len(new.objects))
	for _, object := range old.Objects() {
		if object.ID >= len(old.NameIDs) || old.NameIDs[object.ID] == 0 {
			continue
		}
		copy := getCopy(object)
		if copy.Object == nil {
			continue
		}
		if copy.Object.network != new {
			panic("name ID copy belongs to another network")
		}
		if copy.Object.IsCI() && !object.IsCI() {
			continue
		}
		id := old.NameIDs[object.ID]
		if copy.Complement {
			id ^= 1
		}
		new.NameIDs[copy.Object.ID] = id
	}
}

func Abc_NtkUpdateNameIds(n *Network) {
	if Abc_NtkIsNetlist(n) != 0 || n.NameIDs == nil {
		panic("name ID update requires a converted network")
	}
	path := abcNameIDFile(n.Spec)
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		file.Close()
		panic(err)
	}
	starts := []int64{-1}
	atLineStart := 0
	for i, c := range data {
		if c == '\n' {
			line := data[atLineStart:i]
			space := bytes.IndexByte(line, ' ')
			if space < 0 {
				file.Close()
				panic("malformed name ID line")
			}
			starts = append(starts, int64(atLineStart+space+1))
			atLineStart = i + 1
		}
	}
	for _, object := range n.Objects() {
		if object.ID == 0 || object.ID >= len(n.NameIDs) {
			continue
		}
		id := n.NameIDs[object.ID]
		if id == 0 {
			continue
		}
		index := int(id >> 1)
		if index <= 0 || index >= len(starts) {
			file.Close()
			panic("name ID outside file")
		}
		text := strconv.Itoa(object.ID)
		if id&1 != 0 {
			text = "-" + text
		}
		if _, err := file.WriteAt([]byte(text), starts[index]); err != nil {
			file.Close()
			panic(err)
		}
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("Saved %d names into file %q.\n", len(starts)-1, path)
	n.NameIDs = nil
}
