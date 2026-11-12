package abc

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

func ntkCopyNode(dst *Network, o *Object) Edge {
	if Abc_NtkIsStrash(o.network) != 0 {
		return dst.AndEdges(ntkCopiedEdge(o.fanins[0]), ntkCopiedEdge(o.fanins[1]))
	}
	copy := Abc_NtkDupObj(dst, o, 0)
	for _, e := range o.fanins {
		Abc_ObjAddFanin(copy, ntkCopiedEdge(e))
	}
	return copy.Signal()
}
func ntkCopyPart(src *Network, roots, boundary []*Object, allInputs bool) *Network {
	dst := ntkFreshLike(src, Abc_NtkType_t(src.Type), Abc_NtkFunc_t(src.Function))
	ntkClearCopies(src)
	setCopy(src.constant, dst.One())
	stop := make(map[*Object]bool)
	for _, o := range boundary {
		stop[o] = true
	}
	seen := make(map[*Object]bool)
	var order []*Object
	var visit func(*Object)
	visit = func(o *Object) {
		if seen[o] {
			return
		}
		seen[o] = true
		if stop[o] || o.IsCI() || o == src.constant {
			return
		}
		for _, e := range o.fanins {
			visit(e.Object)
		}
		if o.Kind == ObjectNode || o.Kind == ObjectNet {
			order = append(order, o)
		}
	}
	for _, o := range roots {
		visit(o)
	}
	if boundary == nil {
		boundary = src.cis
	}
	for _, o := range boundary {
		if allInputs || seen[o] {
			pi := dst.NewObject(ObjectPI)
			pi.SetName(o.Name())
			setCopy(o, pi.Signal())
		}
	}
	for _, o := range order {
		if o.Kind == ObjectNet {
			setCopy(o, ntkCopiedEdge(o.fanins[0]))
		} else {
			setCopy(o, ntkCopyNode(dst, o))
		}
	}
	for _, o := range roots {
		edge := getCopy(o)
		if o.IsCO() {
			edge = ntkCopiedEdge(o.fanins[0])
		}
		po := dst.NewObject(ObjectPO)
		po.SetName(o.Name())
		Abc_ObjAddFanin(po, edge)
	}
	return dst
}
func Abc_NtkCreateCone(src *Network, node *Object, name string, all int32) *Network {
	dst := ntkCopyPart(src, []*Object{node}, nil, all != 0)
	dst.Name = src.Name + "_" + name
	dst.kindOrder[ObjectPO][0].SetName(name)
	return dst
}
func Abc_NtkCreateConeArray(src *Network, roots []*Object, all int32) *Network {
	dst := ntkCopyPart(src, roots, nil, all != 0)
	dst.Name = src.Name + "_part"
	return dst
}
func Abc_NtkAppendToCone(dst, src *Network, roots []*Object) {
	if Abc_NtkIsStrash(dst) == 0 || Abc_NtkIsStrash(src) == 0 {
		panic("abc: append cone requires strashed networks")
	}
	nodes := Abc_NtkDfsNodes(src, roots, int32(len(roots)))
	setCopy(src.constant, dst.One())
	byName := make(map[string]*Object)
	for _, o := range dst.cis {
		byName[o.Name()] = o
	}
	used := make(map[*Object]bool)
	var visit func(*Object)
	visit = func(o *Object) {
		if used[o] {
			return
		}
		used[o] = true
		if o.IsCI() {
			return
		}
		for _, e := range o.fanins {
			visit(e.Object)
		}
	}
	for _, o := range roots {
		visit(o)
	}
	for _, ci := range src.cis {
		if !used[ci] {
			continue
		}
		copy := byName[ci.Name()]
		if copy == nil {
			copy = dst.NewObject(ObjectPI)
			copy.SetName(ci.Name())
			byName[ci.Name()] = copy
		}
		setCopy(ci, copy.Signal())
	}
	for _, o := range nodes {
		setCopy(o, ntkCopyNode(dst, o))
	}
}
func Abc_NtkCreateMffc(src *Network, node *Object, name string) *Network {
	var cone, support []*Object
	Abc_NodeDeref_rec(node)
	Abc_NodeMffcConeSupp(node, &cone, &support)
	Abc_NodeRef_rec(node)
	dst := ntkCopyPart(src, []*Object{node}, support, true)
	dst.Name = src.Name + "_" + name
	dst.kindOrder[ObjectPO][0].SetName(name)
	return dst
}
func ntkStrashNode(dst *Network, o *Object) Edge {
	if Abc_NtkHasAig(o.network) != 0 {
		return Abc_ConvertAigToAig(dst, o)
	}
	if text, ok := o.Data.(string); ok {
		count := int(Abc_SopGetVarNum(text))
		if count != len(o.fanins) {
			panic("abc: SOP arity mismatch")
		}
		if Abc_SopIsExorType(text) != 0 {
			result := dst.One().Not()
			for _, e := range o.fanins {
				result = ntkXor(dst, result, ntkCopiedEdge(e))
			}
			if Abc_SopGetPhase(text) == 0 {
				result = result.Not()
			}
			return result
		}
		result := dst.One().Not()
		for _, cube := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
			term := dst.One()
			for i := 0; i < count; i++ {
				if cube[i] == '-' {
					continue
				}
				edge := ntkCopiedEdge(o.fanins[i])
				if cube[i] == '0' {
					edge = edge.Not()
				}
				term = dst.AndEdges(term, edge)
			}
			result = dst.AndEdges(result.Not(), term.Not()).Not()
		}
		if Abc_SopGetPhase(text) == 0 {
			result = result.Not()
		}
		return result
	}
	panic("abc: target strashing requires Hop or SOP functions")
}
func Abc_NtkCreateTarget(src *Network, roots []*Object, values []int32) *Network {
	if len(roots) != len(values) {
		panic("abc: target arity mismatch")
	}
	ntkClearCopies(src)
	dst := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	dst.Name = src.Name
	setCopy(src.constant, dst.One())
	for _, o := range src.cis {
		copy := dst.NewObject(ObjectPI)
		copy.SetName(o.Name())
		setCopy(o, copy.Signal())
	}
	for _, o := range Abc_NtkDfsNodes(src, roots, int32(len(roots))) {
		setCopy(o, ntkStrashNode(dst, o))
	}
	result := dst.One()
	for i, o := range roots {
		edge := getCopy(o)
		if o.IsCO() {
			edge = ntkCopiedEdge(o.fanins[0])
		}
		if values[i] == 0 {
			edge = edge.Not()
		}
		result = dst.AndEdges(result, edge)
	}
	po := dst.NewObject(ObjectPO)
	po.SetName("miter")
	Abc_ObjAddFanin(po, result)
	return dst
}
func Abc_NtkCreateFromNode(src *Network, node *Object) *Network {
	support := make([]*Object, len(node.fanins))
	for i, e := range node.fanins {
		support[i] = e.Object
	}
	dst := ntkCopyPart(src, []*Object{node}, support, true)
	dst.Name = node.Name()
	return dst
}
func Abc_NtkCreateWithNode(sop string) *Network {
	n := Abc_NtkCreateFromSops("ex", []string{sop})
	for i, pi := range n.kindOrder[ObjectPI] {
		name := string(rune('a' + i))
		if len(n.kindOrder[ObjectPI]) >= 26 {
			name = string(rune('a'+i%26)) + string(rune('0'+i/26))
		}
		pi.SetName(name)
	}
	n.kindOrder[ObjectPO][0].SetName("F")
	return n
}
func Abc_NtkCreateFromSops(name string, sops []string) *Network {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	n.Name = name
	if len(sops) == 0 {
		return n
	}
	count := int(Abc_SopGetVarNum(sops[0]))
	for i := 0; i < count; i++ {
		pi := n.NewObject(ObjectPI)
		pi.SetName(fmt.Sprintf("pi%d", i))
	}
	for i, sop := range sops {
		arity := int(Abc_SopGetVarNum(sop))
		if arity > count {
			panic("abc: inconsistent SOP supports")
		}
		node := n.NewObject(ObjectNode)
		node.Data = sop
		for _, pi := range n.kindOrder[ObjectPI][:arity] {
			Abc_ObjAddFanin(node, pi)
		}
		po := n.NewObject(ObjectPO)
		po.SetName(fmt.Sprintf("po%d", i))
		Abc_ObjAddFanin(po, node)
	}
	return n
}

func Abc_NtkDouble(src *Network) *Network {
	if Abc_NtkIsLogic(src) == 0 {
		panic("abc: network doubling requires logic")
	}
	dst := ntkFreshLike(src, Abc_NtkType_t(src.Type), Abc_NtkFunc_t(src.Function))
	dst.Name = src.Name + "_2x"
	for part := 1; part <= 2; part++ {
		ntkClearCopies(src)
		setCopy(src.constant, dst.One())
		for _, o := range src.Objects() {
			if o == src.constant {
				continue
			}
			copy := Abc_NtkDupObj(dst, o, 0)
			copy.Barrier = o.Barrier
			setCopy(o, copy.Signal())
			if o.IsCI() || o.IsCO() {
				copy.SetName(fmt.Sprintf("%d_%s", part, o.Name()))
			}
		}
		for _, o := range src.Objects() {
			if o == src.constant {
				continue
			}
			for _, e := range o.fanins {
				Abc_ObjAddFanin(o.Copy, ntkCopiedEdge(e))
			}
		}
	}
	Abc_NtkOrderCisCos(dst)
	return dst
}
func Abc_NtkBottom(src *Network, level int32) *Network {
	if Abc_NtkIsLogic(src) == 0 || Abc_NtkLatchNum(src) != 0 {
		panic("abc: bottom extraction requires combinational logic")
	}
	dst := ntkFreshLike(src, Abc_NtkType_t(src.Type), Abc_NtkFunc_t(src.Function))
	dst.Name = src.Name + "_bot"
	ntkClearCopies(src)
	setCopy(src.constant, dst.One())
	for _, o := range src.kindOrder[ObjectPI] {
		Abc_NtkDupObj(dst, o, 1)
	}
	for _, o := range src.kindOrder[ObjectNode] {
		if o.Level <= int(level) {
			Abc_NtkDupObj(dst, o, 0)
		}
	}
	for _, o := range src.Objects() {
		copy := getCopy(o)
		if copy.Object != nil && o != src.constant {
			for _, e := range o.fanins {
				if getCopy(e.Object).Object != nil {
					Abc_ObjAddFanin(copy.Object, ntkCopiedEdge(e))
				}
			}
		}
		if copy.Object == nil {
			for _, e := range o.fanins {
				if e.Object.Kind == ObjectNode && getCopy(e.Object).Object != nil {
					po := dst.NewObject(ObjectPO)
					Abc_ObjAddFanin(po, ntkCopiedEdge(e))
				}
			}
		}
	}
	return dst
}
func Abc_NtkAttachBottom(top, bottom *Network) *Network {
	if bottom == nil {
		panic("abc: nil bottom network")
	}
	if top == nil {
		return bottom
	}
	inputs := top.kindOrder[ObjectPI]
	outputs := bottom.kindOrder[ObjectPO]
	if len(outputs) != len(inputs) || Abc_NtkLatchNum(top) != 0 || Abc_NtkLatchNum(bottom) != 0 || Abc_NtkPiNum(bottom) >= Abc_NtkPiNum(top) {
		panic("abc: incompatible composition interfaces")
	}
	var buffers []*Object
	for _, pi := range inputs {
		buf := Abc_NtkCreateNodeBuf(top, nil)
		Abc_ObjTransferFanout(pi, buf)
		buffers = append(buffers, buf)
	}
	for _, pi := range inputs[len(bottom.kindOrder[ObjectPI]):] {
		Abc_NtkDeleteObj(pi)
	}
	setCopy(bottom.constant, top.One())
	for i, pi := range bottom.kindOrder[ObjectPI] {
		setCopy(pi, inputs[i].Signal())
	}
	for _, o := range Abc_NtkDfs(bottom, 0) {
		setCopy(o, ntkCopyNode(top, o))
	}
	for i, po := range outputs {
		Abc_ObjAddFanin(buffers[i], ntkCopiedEdge(po.fanins[0]))
	}
	Abc_NtkDelete(bottom)
	return top
}
func Abc_NtkFixNonDrivenNets(n *Network) {
	if len(n.kindOrder[ObjectNode]) == 0 && len(n.boxes) == 0 {
		return
	}
	if net := Abc_NtkFindNet(n, "[_c1_]"); net != nil && len(net.fanins) == 0 {
		Abc_ObjAddFanin(net, Abc_NtkCreateNodeConst1(n))
	}
	var fixed []string
	for _, net := range n.kindOrder[ObjectNet] {
		if len(net.fanins) == 0 {
			Abc_ObjAddFanin(net, Abc_NtkCreateNodeConst0(n))
			fixed = append(fixed, net.Name())
		}
	}
	if len(fixed) > 0 {
		fmt.Printf("Warning: Constant-0 drivers added to %d non-driven nets in network %q\n", len(fixed), n.Name)
	}
}
func ntkChangeKind(o *Object, kind ObjectKind) {
	n := o.network
	n.kindOrder[o.Kind] = removeObjectOrder(n.kindOrder[o.Kind], o)
	o.Kind = kind
	n.kindOrder[kind] = append(n.kindOrder[kind], o)
}
func Abc_NtkMakeComb(n *Network, remove int32) {
	if Abc_NtkLatchNum(n) == 0 {
		return
	}
	if Abc_NtkIsNetlist(n) != 0 || Abc_NtkHasOnlyLatchBoxes(n) == 0 {
		panic("abc: combinational conversion requires latch-only logic")
	}
	for _, latch := range append([]*Object(nil), n.boxes...) {
		Abc_NtkDeleteObj(latch)
	}
	for _, bo := range n.ObjectsOfKind(ObjectBO) {
		ntkChangeKind(bo, ObjectPI)
	}
	for _, bi := range n.ObjectsOfKind(ObjectBI) {
		if remove != 0 {
			Abc_NtkDeleteObj(bi)
		} else {
			ntkChangeKind(bi, ObjectPO)
		}
	}
	Abc_NtkOrderCisCos(n)
	if remove != 0 {
		ntkSweepDangling(n)
	}
}
func Abc_NtkMakeSeq(n *Network, count int32) {
	if len(n.boxes) != 0 || count < 0 {
		panic("abc: sequential conversion requires combinational network")
	}
	pis, pos := n.ObjectsOfKind(ObjectPI), n.ObjectsOfKind(ObjectPO)
	if int(count) >= len(pis) || int(count) >= len(pos) {
		fmt.Println("The number of latches is more or equal than the number of PIs or POs.")
		return
	}
	for i := 0; i < int(count); i++ {
		out, in := pis[len(pis)-int(count)+i], pos[len(pos)-int(count)+i]
		ntkChangeKind(out, ObjectBO)
		ntkChangeKind(in, ObjectBI)
		latch := n.NewObject(ObjectLatch)
		Abc_ObjAddFanin(latch, in)
		Abc_ObjAddFanin(out, latch)
		Abc_LatchSetInit0(latch)
	}
	Abc_NtkOrderCisCos(n)
}
func ntkSweepDangling(n *Network) {
	for {
		changed := false
		for _, o := range n.ObjectsOfKind(ObjectNode) {
			if len(o.fanouts) == 0 {
				Abc_NtkDeleteObj(o)
				changed = true
			}
		}
		if !changed {
			return
		}
	}
}
func Abc_NtkMakeOnePo(src *Network, output, count int32) *Network {
	if output < 0 || output >= Abc_NtkPoNum(src) {
		fmt.Println("PO index is incorrect.")
		return nil
	}
	dst := Abc_NtkDup(src)
	if count < 1 {
		count = 1
	}
	for i, po := range dst.ObjectsOfKind(ObjectPO) {
		if i < int(output) || i >= int(output+count) {
			Abc_NtkDeleteObj(po)
		}
	}
	Abc_NtkOrderCisCos(dst)
	if Abc_NtkIsStrash(dst) != 0 {
		Abc_AigCleanup(dst.FunctionManager.(*Abc_Aig_t))
	}
	return dst
}
func Abc_NtkTrim(n *Network) *Network {
	for _, po := range n.ObjectsOfKind(ObjectPO) {
		driver := po.fanins[0].Object
		remove := len(driver.fanins) == 0 || len(driver.fanins) == 1 && driver.fanins[0].Object.Kind == ObjectPI
		if remove {
			Abc_NtkDeleteObj(po)
			if driver != n.constant && driver.Kind == ObjectNode && len(driver.fanouts) == 0 {
				Abc_NtkDeleteObj_rec(driver, 1)
			}
		}
	}
	for _, pi := range n.ObjectsOfKind(ObjectPI) {
		if len(pi.fanouts) == 0 {
			Abc_NtkDeleteObj(pi)
		}
	}
	Abc_NtkOrderCisCos(n)
	return Abc_NtkDup(n)
}
func Abc_NtkDropSatOutputs(n *Network, cexes []any, verbose int32) {
	pos := n.ObjectsOfKind(ObjectPO)
	if len(pos) != len(cexes) {
		panic("abc: counterexample vector differs from outputs")
	}
	count := 0
	for i, cex := range cexes {
		if cex != nil {
			Abc_NtkDropOneOutput(n, int32(i), 1, 0)
			count++
		}
	}
	if verbose != 0 {
		fmt.Printf("Logic cones of %d POs have been replaced by constant 0.\n", count)
	}
	Abc_AigCleanup(n.FunctionManager.(*Abc_Aig_t))
}
func Abc_NtkDropOneOutput(n *Network, index, skip, one int32) {
	po := n.kindOrder[ObjectPO][index]
	edge := n.One()
	if one == 0 {
		edge = edge.Not()
	}
	if err := n.PatchFanin(po, 0, edge); err != nil {
		panic(err)
	}
	if skip == 0 {
		Abc_AigCleanup(n.FunctionManager.(*Abc_Aig_t))
	}
}
func Abc_NtkSwapOneOutput(n *Network, index int32) {
	if index == 0 {
		return
	}
	a, b := n.kindOrder[ObjectPO][0], n.kindOrder[ObjectPO][index]
	x, y := a.fanins[0], b.fanins[0]
	if err := n.PatchFanin(a, 0, y); err != nil {
		panic(err)
	}
	if err := n.PatchFanin(b, 0, x); err != nil {
		panic(err)
	}
}
func Abc_NtkRemovePo(n *Network, index, removeZero int32) {
	po := n.kindOrder[ObjectPO][index]
	if po.fanins[0].Object == n.constant && abcBool(po.fanins[0].Complement) == removeZero {
		Abc_NtkDeleteObj(po)
	}
}
func Abc_NtkReadFlopPerm(name string, count int32) []int32 {
	file, err := os.Open(name)
	if err != nil {
		fmt.Printf("Cannot open input file %q.\n", name)
		return nil
	}
	defer file.Close()
	var perm []int32
	seen := make(map[int]bool)
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		text := scan.Text()
		if text == "" || text[0] == ' ' || text[0] == '\r' {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) == 0 {
			continue
		}
		v, err := strconv.Atoi(fields[0])
		if err != nil || v < 0 || v >= int(count) || seen[v] {
			return nil
		}
		seen[v] = true
		perm = append(perm, int32(v))
	}
	if scan.Err() != nil || len(perm) != int(count) {
		return nil
	}
	return perm
}
func Abc_NtkPermute(n *Network, inputs, outputs, flops int32, file string) {
	if len(n.boxes) != len(n.kindOrder[ObjectLatch]) {
		panic("abc: permutation requires latch-only boxes")
	}
	orders := [][]*Object{append([]*Object(nil), n.kindOrder[ObjectPI]...), append([]*Object(nil), n.kindOrder[ObjectPO]...), append([]*Object(nil), n.boxes...)}
	flags := []int32{inputs, outputs, flops}
	var fixed []int32
	if file != "" {
		fixed = Abc_NtkReadFlopPerm(file, int32(len(n.boxes)))
		if fixed == nil {
			return
		}
		flags = []int32{0, 0, 0}
	}
	n.ObjectPermutation = nil
	for group, order := range orders {
		perm := make([]int, len(order))
		for i := range perm {
			perm[i] = i
		}
		if flags[group] != 0 {
			for i := range order {
				k := rand.Intn(len(order))
				order[i], order[k] = order[k], order[i]
				perm[i], perm[k] = perm[k], perm[i]
			}
		}
		inverse := make([]int32, len(perm))
		for i, p := range perm {
			inverse[p] = int32(i)
		}
		if group == 2 && fixed != nil {
			for position, original := range fixed {
				inverse[original] = int32(position)
			}
		}
		n.ObjectPermutation = append(n.ObjectPermutation, inverse...)
		orders[group] = order
	}
	n.kindOrder[ObjectPI], n.kindOrder[ObjectPO] = orders[0], orders[1]
	n.boxes = orders[2]
	n.kindOrder[ObjectLatch] = append([]*Object(nil), n.boxes...)
	Abc_NtkOrderCisCos(n)
}
func Abc_NodeCompareByFanoutCount(a, b **Object) int32 {
	if len((*a).fanouts) > len((*b).fanouts) {
		return -1
	}
	if len((*a).fanouts) < len((*b).fanouts) {
		return 1
	}
	return int32(strings.Compare((*a).Name(), (*b).Name()))
}
func Abc_NtkPermutePiUsingFanout(n *Network) {
	sort.SliceStable(n.kindOrder[ObjectPI], func(i, j int) bool {
		a, b := n.kindOrder[ObjectPI][i], n.kindOrder[ObjectPI][j]
		return Abc_NodeCompareByFanoutCount(&a, &b) < 0
	})
	Abc_NtkOrderCisCos(n)
}
func Abc_NtkUnpermute(n *Network) {
	if n.ObjectPermutation == nil {
		fmt.Println("Abc_NtkUnpermute(): Initial permutation is not available.")
		return
	}
	groups := [][]*Object{n.kindOrder[ObjectPI], n.kindOrder[ObjectPO], n.boxes}
	offset := 0
	for group, order := range groups {
		if offset+len(order) > len(n.ObjectPermutation) {
			panic("abc: invalid saved permutation")
		}
		restored := make([]*Object, len(order))
		for i := range restored {
			index := n.ObjectPermutation[offset+i]
			if index < 0 || index >= int32(len(order)) {
				panic("abc: invalid saved permutation index")
			}
			restored[i] = order[index]
		}
		groups[group] = restored
		offset += len(order)
	}
	n.kindOrder[ObjectPI], n.kindOrder[ObjectPO], n.boxes = groups[0], groups[1], groups[2]
	n.kindOrder[ObjectLatch] = append([]*Object(nil), n.boxes...)
	n.ObjectPermutation = nil
	Abc_NtkOrderCisCos(n)
}
func Abc_NtkNodeDup(src *Network, limit, verbose int32) *Network {
	if limit < 2 {
		panic("abc: fanout splitting limit must be at least two")
	}
	dst := Abc_NtkDup(src)
	for {
		var nodes []*Object
		for _, o := range dst.kindOrder[ObjectNode] {
			if len(o.fanouts) >= int(limit) {
				nodes = append(nodes, o)
			}
		}
		if len(nodes) == 0 {
			return dst
		}
		for _, o := range nodes {
			copy := Abc_NtkDupObj(dst, o, 0)
			for _, e := range o.fanins {
				Abc_ObjAddFanin(copy, e)
			}
			fanouts := o.Fanouts()
			for _, fanout := range fanouts[:int(limit)/2] {
				Abc_ObjPatchFanin(fanout, o, copy)
			}
		}
		if verbose != 0 {
			fmt.Printf("Duplicated %d nodes.\n", len(nodes))
		}
	}
}

// NtkLiteralNetwork accepts the native GIA literal interface without coupling
// ABC network ownership to a particular GIA representation.
type NtkLiteralNetwork interface {
	ObjectCount() int
	InputLiterals() []int32
	OutputLiterals() []int32
	Gate(int32) (int32, int32, int32, int)
}

func Abc_NtkCreateFromGias(name string, graphs []NtkLiteralNetwork, multi NtkLiteralNetwork) *Network {
	if multi != nil {
		graphs = []NtkLiteralNetwork{multi}
	}
	dst := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	dst.Name = name
	if len(graphs) == 0 {
		return dst
	}
	for range graphs[0].InputLiterals() {
		dst.NewObject(ObjectPI)
	}
	for _, g := range graphs {
		if len(g.InputLiterals()) != len(dst.kindOrder[ObjectPI]) || multi == nil && len(g.OutputLiterals()) != 1 {
			panic("abc: GIA interface differs")
		}
		copies := make([]Edge, g.ObjectCount())
		copies[0] = dst.One().Not()
		for i, lit := range g.InputLiterals() {
			copies[lit>>1] = dst.kindOrder[ObjectPI][i].Signal()
		}
		var copyLit func(int32) Edge
		copyLit = func(lit int32) Edge {
			id := lit >> 1
			if id < 0 || id >= int32(len(copies)) {
				panic("abc: invalid GIA literal")
			}
			edge := copies[id]
			if edge.Object == nil {
				a, b, c, kind := g.Gate(id)
				switch kind {
				case 2:
					edge = dst.AndEdges(copyLit(a), copyLit(b))
				case 3:
					edge = ntkXor(dst, copyLit(a), copyLit(b))
				case 4:
					control, x, y := copyLit(a), copyLit(b), copyLit(c)
					edge = dst.AndEdges(dst.AndEdges(control, x).Not(), dst.AndEdges(control.Not(), y).Not()).Not()
				case 5, 6:
					edge = copyLit(a)
				default:
					panic("abc: unsupported GIA gate")
				}
				copies[id] = edge
			}
			if lit&1 != 0 {
				edge = edge.Not()
			}
			return edge
		}
		for _, lit := range g.OutputLiterals() {
			Abc_ObjAddFanin(dst.NewObject(ObjectPO), copyLit(lit))
		}
	}
	for i, o := range dst.kindOrder[ObjectPI] {
		o.SetName(fmt.Sprintf("pi%d", i))
	}
	for i, o := range dst.kindOrder[ObjectPO] {
		o.SetName(fmt.Sprintf("po%d", i))
	}
	return dst
}
