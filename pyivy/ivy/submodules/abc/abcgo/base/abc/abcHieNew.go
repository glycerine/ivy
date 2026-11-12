package abc

// The experimental AU hierarchy uses integer literals and owns its topology.
// Go slices replace the original aligned page allocator and address-derived IDs.
import (
	"bufio"
	"fmt"
	aig "github.com/glycerine/abcgo/aig/aig"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

type Au_Type_t uint8

const (
	AU_OBJ_NONE Au_Type_t = iota
	AU_OBJ_CONST0
	AU_OBJ_PI
	AU_OBJ_PO
	AU_OBJ_FAN
	AU_OBJ_FLOP
	AU_OBJ_BOX
	AU_OBJ_NODE
	AU_OBJ_VOID
)

type Au_Obj_t struct {
	Func, Value       int
	Type              Au_Type_t
	Fanins            []int
	id, pio           int
	network           *Au_Ntk_t
	fanouts           []int
	copy              int
	trav              uint64
	regular, negative *Au_Obj_t
}
type Au_Ntk_t struct {
	Name                                   string
	Man                                    *Au_Man_t
	ID                                     int
	Objects                                []*Au_Obj_t
	PIs, POs                               []int
	counts                                 [AU_OBJ_VOID]int
	epoch                                  uint64
	Boxes, Nodes, Ports, Ands, Xors, Muxes float64
}
type Au_Man_t struct {
	Name                      string
	Networks                  []*Au_Ntk_t
	Functions                 []string
	functionIDs               map[string]int
	PortsC0, PortsC1, PortsNC float64
}

func Au_Var2Lit(v, c int) int { return v*2 + c }
func Au_Lit2Var(l int) int    { return l >> 1 }
func Au_LitIsCompl(l int) int { return l & 1 }
func Au_LitNot(l int) int     { return l ^ 1 }
func Au_LitNotCond(l, c int) int {
	if c > 0 {
		return l ^ 1
	}
	return l
}
func Au_LitRegular(l int) int { return l &^ 1 }
func Au_Regular(o *Au_Obj_t) *Au_Obj_t {
	if o != nil && o.regular != nil {
		return o.regular
	}
	return o
}
func Au_Not(o *Au_Obj_t) *Au_Obj_t {
	if o == nil {
		panic("nil AU complement")
	}
	if o.regular != nil {
		return o.regular
	}
	if o.negative == nil {
		o.negative = &Au_Obj_t{regular: o}
	}
	return o.negative
}
func Au_NotCond(o *Au_Obj_t, c int) *Au_Obj_t {
	if c != 0 {
		return Au_Not(o)
	}
	return o
}
func Au_IsComplement(o *Au_Obj_t) int   { return int(abcBool(o != nil && o.regular != nil)) }
func Au_UtilStrsav(s string) string     { return s }
func Au_ManAlloc(name string) *Au_Man_t { return &Au_Man_t{Name: name, Networks: []*Au_Ntk_t{nil}} }
func Au_NtkAlloc(m *Au_Man_t, name string) *Au_Ntk_t {
	n := &Au_Ntk_t{Name: name, Objects: []*Au_Obj_t{nil}, epoch: 1}
	Au_ManAddNtk(m, n)
	return n
}
func Au_ManAddNtk(m *Au_Man_t, n *Au_Ntk_t) {
	if m == nil || n == nil || Au_ManFindNtk(m, n.Name) >= 0 {
		panic("invalid or duplicate AU model")
	}
	n.Man = m
	n.ID = len(m.Networks)
	m.Networks = append(m.Networks, n)
}
func Au_ManFindNtk(m *Au_Man_t, name string) int {
	for i, n := range m.Networks {
		if n != nil && n.Name == name {
			return i
		}
	}
	return -1
}
func Au_ManFindNtkP(m *Au_Man_t, name string) *Au_Ntk_t {
	i := Au_ManFindNtk(m, name)
	if i < 0 {
		return nil
	}
	return m.Networks[i]
}
func Au_NtkFree(n *Au_Ntk_t) {
	if n == nil {
		return
	}
	if n.Man != nil && n.ID < len(n.Man.Networks) && n.Man.Networks[n.ID] == n {
		n.Man.Networks[n.ID] = nil
	}
	for _, o := range n.Objects {
		if o != nil {
			o.network = nil
			o.Fanins = nil
			o.fanouts = nil
		}
	}
	n.Objects = nil
	n.PIs = nil
	n.POs = nil
	n.Man = nil
}
func Au_ManFree(m *Au_Man_t) {
	if m == nil {
		return
	}
	for _, n := range m.Networks {
		if n != nil {
			return
		}
	}
	m.Networks = nil
	m.Functions = nil
	m.functionIDs = nil
}
func Au_ManDelete(m *Au_Man_t) {
	if m == nil {
		return
	}
	for _, n := range m.Networks {
		Au_NtkFree(n)
	}
	Au_ManFree(m)
}
func Au_NtkMemUsage(n *Au_Ntk_t) int {
	total := int(unsafe.Sizeof(*n)) + cap(n.PIs)*int(unsafe.Sizeof(int(0))) + cap(n.POs)*int(unsafe.Sizeof(int(0))) + cap(n.Objects)*int(unsafe.Sizeof((*Au_Obj_t)(nil)))
	for _, o := range n.Objects {
		if o != nil {
			total += int(unsafe.Sizeof(*o)) + (cap(o.Fanins)+cap(o.fanouts))*int(unsafe.Sizeof(int(0)))
		}
	}
	return total
}
func Au_ManMemUsage(m *Au_Man_t) int {
	total := 0
	for _, n := range m.Networks {
		if n != nil {
			total += Au_NtkMemUsage(n)
		}
	}
	return total
}
func Au_ManMemUsageUseful(m *Au_Man_t) int { return Au_ManMemUsage(m) } // Native allocations contain no page padding.
func Au_NtkCleanCopy(n *Au_Ntk_t) {
	for _, o := range n.Objects {
		if o != nil {
			o.copy = -1
		}
	}
}
func Au_NtkNodeNumFunc(n *Au_Ntk_t, f int) int {
	if len(n.Man.Functions) > 0 {
		return 0
	}
	count := 0
	for _, o := range n.Objects {
		if o != nil && o.Type == AU_OBJ_NODE && o.Func == f {
			count++
		}
	}
	return count
}
func Au_NtkPrintStats(n *Au_Ntk_t) {
	fmt.Printf("%-30s: i/o =%6d/%6d lat =%5d nd =%6d box =%5d obj =%7d %.1f MB\n", n.Name, len(n.PIs), len(n.POs), n.counts[AU_OBJ_FLOP], n.counts[AU_OBJ_NODE], n.counts[AU_OBJ_BOX], len(n.Objects)-1, float64(Au_NtkMemUsage(n))/(1<<20))
}
func Au_ManPrintStats(m *Au_Man_t) {
	fmt.Printf("Design %s\n", m.Name)
	for _, n := range m.Networks {
		if n != nil {
			Au_NtkPrintStats(n)
		}
	}
	fmt.Printf("Different functions = %d. Memory = %.1f MB\n", len(m.Functions), float64(Au_ManMemUsage(m))/(1<<20))
}
func Au_NtkAllocObj(n *Au_Ntk_t, fanins, kind int) int {
	if fanins < 0 || kind <= int(AU_OBJ_NONE) || kind >= int(AU_OBJ_VOID) {
		panic("invalid AU object")
	}
	o := &Au_Obj_t{Type: Au_Type_t(kind), Fanins: make([]int, fanins), id: len(n.Objects), network: n, copy: -1}
	n.Objects = append(n.Objects, o)
	n.counts[o.Type]++
	switch o.Type {
	case AU_OBJ_PI:
		o.pio = len(n.PIs)
		n.PIs = append(n.PIs, o.id)
	case AU_OBJ_PO:
		o.pio = len(n.POs)
		n.POs = append(n.POs, o.id)
	}
	return o.id
}
func Au_NtkCreateConst0(n *Au_Ntk_t) int { return Au_NtkAllocObj(n, 0, int(AU_OBJ_CONST0)) }
func Au_NtkCreatePi(n *Au_Ntk_t) int     { return Au_NtkAllocObj(n, 0, int(AU_OBJ_PI)) }
func Au_NtkCreatePo(n *Au_Ntk_t, lit int) int {
	id := Au_NtkAllocObj(n, 1, int(AU_OBJ_PO))
	n.Objects[id].Fanins[0] = lit
	return id
}
func Au_NtkCreateFan(n *Au_Ntk_t, fanin, output, model int) int {
	id := Au_NtkAllocObj(n, 1, int(AU_OBJ_FAN))
	o := n.Objects[id]
	o.Fanins[0] = fanin
	o.pio = output
	o.Func = model
	return id
}
func Au_NtkCreateNode(n *Au_Ntk_t, fanins []int, function int) int {
	id := Au_NtkAllocObj(n, len(fanins), int(AU_OBJ_NODE))
	o := n.Objects[id]
	copy(o.Fanins, fanins)
	o.Func = function
	return id
}
func Au_NtkCreateBox(n *Au_Ntk_t, fanins []int, outputs, model int) int {
	if outputs < 0 || model <= 0 {
		panic("invalid AU box")
	}
	id := Au_NtkAllocObj(n, len(fanins), int(AU_OBJ_BOX))
	o := n.Objects[id]
	copy(o.Fanins, fanins)
	o.Func = model
	for i := 0; i < outputs; i++ {
		o.fanouts = append(o.fanouts, Au_NtkCreateFan(n, id*2, i, model))
	}
	return id
}
func Au_NtkIncrementTravId(n *Au_Ntk_t) {
	n.epoch++
	if n.epoch == 0 {
		for _, o := range n.Objects {
			if o != nil {
				o.trav = 0
			}
		}
		n.epoch = 1
	}
}
func Au_ObjSetTravIdCurrent(o *Au_Obj_t)  { o = Au_Regular(o); o.trav = o.network.epoch }
func Au_ObjSetTravIdPrevious(o *Au_Obj_t) { o = Au_Regular(o); o.trav = o.network.epoch - 1 }
func Au_ObjIsTravIdCurrent(o *Au_Obj_t) int {
	o = Au_Regular(o)
	return int(abcBool(o.trav == o.network.epoch))
}
func Au_ObjIsTravIdPrevious(o *Au_Obj_t) int {
	o = Au_Regular(o)
	return int(abcBool(o.trav == o.network.epoch-1))
}
func Au_ObjSetTravIdCurrentId(n *Au_Ntk_t, id int)    { Au_ObjSetTravIdCurrent(Au_NtkObj(n, id)) }
func Au_ObjIsTravIdCurrentId(n *Au_Ntk_t, id int) int { return Au_ObjIsTravIdCurrent(Au_NtkObj(n, id)) }
func Au_ObjSuppSize_rec(n *Au_Ntk_t, id int) int {
	o := Au_NtkObj(n, id)
	if Au_ObjIsTravIdCurrent(o) != 0 {
		return 0
	}
	Au_ObjSetTravIdCurrent(o)
	if o.Type == AU_OBJ_PI {
		return 1
	}
	total := 0
	for _, lit := range o.Fanins {
		total += Au_ObjSuppSize_rec(n, lit>>1)
	}
	return total
}
func Au_ObjSuppSize(o *Au_Obj_t) int {
	Au_NtkIncrementTravId(o.network)
	return Au_ObjSuppSize_rec(o.network, o.id)
}
func Au_NtkSuppSizeTest(n *Au_Ntk_t) int {
	count := 0
	for _, o := range n.Objects {
		if o != nil && o.Type == AU_OBJ_NODE && Au_ObjSuppSize(o) <= 16 {
			count++
		}
	}
	fmt.Printf("Nodes with small support %d (out of %d)\n", count, n.counts[AU_OBJ_NODE])
	return count
}

func Au_ManReorderModels_rec(n *Au_Ntk_t, order *[]int) {
	seen := map[*Au_Ntk_t]bool{}
	var visit func(*Au_Ntk_t)
	visit = func(n *Au_Ntk_t) {
		if seen[n] {
			return
		}
		seen[n] = true
		for _, o := range n.Objects {
			if o != nil && o.Type == AU_OBJ_BOX {
				child := Au_ObjModel(o)
				if child != nil && child != n {
					visit(child)
				}
			}
		}
		*order = append(*order, n.ID)
	}
	visit(n)
}
func Au_ManReorderModels(m *Au_Man_t, root *Au_Ntk_t) {
	var order []int
	Au_ManReorderModels_rec(root, &order)
	seen := map[int]bool{}
	for _, id := range order {
		seen[id] = true
	}
	order = order[:len(order)-1]
	for _, n := range m.Networks {
		if n != nil && !seen[n.ID] {
			order = append(order, n.ID)
		}
	}
	order = append(order, root.ID)
	reverse := make(map[int]int)
	models := []*Au_Ntk_t{nil}
	for i := len(order) - 1; i >= 0; i-- {
		reverse[order[i]] = len(models)
		models = append(models, m.Networks[order[i]])
	}
	for _, n := range models[1:] {
		n.ID = reverse[n.ID]
		for _, o := range n.Objects {
			if o != nil && (o.Type == AU_OBJ_BOX || o.Type == AU_OBJ_FAN) {
				o.Func = reverse[o.Func]
			}
		}
	}
	m.Networks = models
}
func Au_NtkCheckRecursive(n *Au_Ntk_t) int {
	active, seen := map[*Au_Ntk_t]bool{}, map[*Au_Ntk_t]bool{}
	var visit func(*Au_Ntk_t) bool
	visit = func(n *Au_Ntk_t) bool {
		if active[n] {
			return true
		}
		if seen[n] {
			return false
		}
		seen[n] = true
		active[n] = true
		defer delete(active, n)
		for _, o := range n.Objects {
			if o != nil && o.Type == AU_OBJ_BOX {
				if visit(Au_ObjModel(o)) {
					return true
				}
			}
		}
		return false
	}
	return int(abcBool(visit(n)))
}
func Au_ManCountThings(m *Au_Man_t) {
	active, done := map[*Au_Ntk_t]bool{}, map[*Au_Ntk_t]bool{}
	var count func(*Au_Ntk_t)
	count = func(n *Au_Ntk_t) {
		if done[n] {
			return
		}
		if active[n] {
			panic("recursive AU model")
		}
		active[n] = true
		n.Boxes = float64(n.counts[AU_OBJ_BOX])
		n.Nodes = float64(n.counts[AU_OBJ_NODE])
		n.Ports = float64(len(n.PIs) + len(n.POs))
		n.Ands = float64(Au_NtkNodeNumFunc(n, 1))
		n.Xors = float64(Au_NtkNodeNumFunc(n, 2))
		n.Muxes = float64(Au_NtkNodeNumFunc(n, 3))
		for _, o := range n.Objects {
			if o != nil && o.Type == AU_OBJ_BOX {
				child := Au_ObjModel(o)
				if child == n {
					continue
				}
				count(child)
				n.Boxes += child.Boxes
				n.Nodes += child.Nodes
				n.Ports += child.Ports
				n.Ands += child.Ands
				n.Xors += child.Xors
				n.Muxes += child.Muxes
			}
		}
		delete(active, n)
		done[n] = true
	}
	for _, n := range m.Networks {
		if n != nil {
			count(n)
		}
	}
	root := Au_ManNtkRoot(m)
	fmt.Printf("Total nodes = %.0f. Total instances = %.0f. Total ports = %.0f.\n", root.Nodes, root.Boxes, root.Ports)
}
func Au_NtkCompareNames(a, b *Au_Ntk_t) int { return strings.Compare(a.Name, b.Name) }
func Au_NtkCompareSign(a, b *Au_Ntk_t) int {
	if len(a.PIs) != len(b.PIs) {
		return len(a.PIs) - len(b.PIs)
	}
	return len(a.POs) - len(b.POs)
}
func auPrintModel(n *Au_Ntk_t) {
	fmt.Printf("MODULE %-30s : PI=%6d PO=%6d BB=%6d ND=%6d\n", n.Name, len(n.PIs), len(n.POs), n.counts[AU_OBJ_BOX], n.counts[AU_OBJ_NODE])
}
func Au_ManPrintBoxInfo(n *Au_Ntk_t) {
	for _, model := range n.Man.Networks {
		if model == nil {
			continue
		}
		auPrintModel(model)
		counts := map[int]int{}
		for _, o := range model.Objects {
			if o != nil && o.Type == AU_OBJ_BOX && o.Func != model.ID {
				counts[o.Func]++
			}
		}
		for id := 1; id < len(n.Man.Networks); id++ {
			if counts[id] > 0 {
				fmt.Printf("%15d : %s\n", counts[id], n.Man.Networks[id].Name)
			}
		}
	}
}
func Au_ManPrintBoxInfoSorted(n *Au_Ntk_t) {
	models := append([]*Au_Ntk_t(nil), n.Man.Networks[1:]...)
	sort.SliceStable(models, func(i, j int) bool { return Au_NtkCompareSign(models[i], models[j]) < 0 })
	for _, model := range models {
		auPrintModel(model)
	}
}

func Au_ManName(m *Au_Man_t) string { return m.Name }

func Au_ManNtkNum(m *Au_Man_t) int { return len(m.Networks) - 1 }

func Au_ManNtk(m *Au_Man_t, i int) *Au_Ntk_t { return m.Networks[i] }

func Au_ManNtkRoot(m *Au_Man_t) *Au_Ntk_t { return m.Networks[1] }

func Au_NtkName(n *Au_Ntk_t) string { return n.Name }

func Au_NtkMan(n *Au_Ntk_t) *Au_Man_t { return n.Man }

func Au_NtkObjNumMax(n *Au_Ntk_t) int { return len(n.Objects) }

func Au_NtkObjNum(n *Au_Ntk_t) int { return len(n.Objects) - 1 }

func Au_NtkObj(n *Au_Ntk_t, id int) *Au_Obj_t { return n.Objects[id] }

func Au_NtkPi(n *Au_Ntk_t, i int) *Au_Obj_t { return n.Objects[n.PIs[i]] }

func Au_NtkPo(n *Au_Ntk_t, i int) *Au_Obj_t { return n.Objects[n.POs[i]] }

func Au_NtkObjI(n *Au_Ntk_t, i int) *Au_Obj_t { return n.Objects[i+1] }

func Au_ObjNtk(o *Au_Obj_t) *Au_Ntk_t { return Au_Regular(o).network }

func Au_ObjId(o *Au_Obj_t) int { return Au_Regular(o).id }

func Au_ObjPioNum(o *Au_Obj_t) int { return Au_Regular(o).pio }

func Au_ObjFunc(o *Au_Obj_t) int { return Au_Regular(o).Func }

func Au_ObjModel(o *Au_Obj_t) *Au_Ntk_t {
	return Au_Regular(o).network.Man.Networks[Au_Regular(o).Func]
}

func Au_ObjFaninNum(o *Au_Obj_t) int { return len(Au_Regular(o).Fanins) }

func Au_ObjFaninId(o *Au_Obj_t, i int) int { return Au_Regular(o).Fanins[i] >> 1 }

func Au_ObjFaninC(o *Au_Obj_t, i int) int { return Au_Regular(o).Fanins[i] & 1 }

func Au_ObjFaninLit(o *Au_Obj_t, i int) int { return Au_Regular(o).Fanins[i] }

func Au_ObjFanin(o *Au_Obj_t, i int) *Au_Obj_t { return Au_NtkObj(Au_ObjNtk(o), Au_ObjFaninId(o, i)) }

func Au_BoxFanoutNum(o *Au_Obj_t) int { return len(Au_Regular(o).fanouts) }

func Au_BoxFanoutId(o *Au_Obj_t, i int) int { return Au_Regular(o).fanouts[i] }

func Au_BoxFanout(o *Au_Obj_t, i int) *Au_Obj_t { return Au_NtkObj(Au_ObjNtk(o), Au_BoxFanoutId(o, i)) }

func Au_ObjFanout(o *Au_Obj_t, i int) int { return Au_BoxFanoutId(o, i) }

func Au_ObjCopy(o *Au_Obj_t) int { return Au_Regular(o).copy }

func Au_ObjGetXsim(o *Au_Obj_t) int { return Au_Regular(o).Value }

func Au_NtkPiNum(n *Au_Ntk_t) int { return n.counts[AU_OBJ_PI] }

func Au_NtkPoNum(n *Au_Ntk_t) int { return n.counts[AU_OBJ_PO] }

func Au_NtkFanNum(n *Au_Ntk_t) int { return n.counts[AU_OBJ_FAN] }

func Au_NtkFlopNum(n *Au_Ntk_t) int { return n.counts[AU_OBJ_FLOP] }

func Au_NtkBoxNum(n *Au_Ntk_t) int { return n.counts[AU_OBJ_BOX] }

func Au_NtkNodeNum(n *Au_Ntk_t) int { return n.counts[AU_OBJ_NODE] }

func Au_ObjIsNone(o *Au_Obj_t) int { return int(abcBool(Au_Regular(o).Type == AU_OBJ_NONE)) }

func Au_ObjIsConst0(o *Au_Obj_t) int { return int(abcBool(Au_Regular(o).Type == AU_OBJ_CONST0)) }

func Au_ObjIsPi(o *Au_Obj_t) int { return int(abcBool(Au_Regular(o).Type == AU_OBJ_PI)) }

func Au_ObjIsPo(o *Au_Obj_t) int { return int(abcBool(Au_Regular(o).Type == AU_OBJ_PO)) }

func Au_ObjIsFan(o *Au_Obj_t) int { return int(abcBool(Au_Regular(o).Type == AU_OBJ_FAN)) }

func Au_ObjIsFlop(o *Au_Obj_t) int { return int(abcBool(Au_Regular(o).Type == AU_OBJ_FLOP)) }

func Au_ObjIsBox(o *Au_Obj_t) int { return int(abcBool(Au_Regular(o).Type == AU_OBJ_BOX)) }

func Au_ObjIsNode(o *Au_Obj_t) int { return int(abcBool(Au_Regular(o).Type == AU_OBJ_NODE)) }

func Au_ObjIsTerm(o *Au_Obj_t) int {
	return int(abcBool(Au_Regular(o).Type >= AU_OBJ_PI && Au_Regular(o).Type <= AU_OBJ_FLOP))
}

func Au_ObjFaninId0(o *Au_Obj_t) int { return Au_ObjFaninId(o, 0) }

func Au_ObjFanin0(o *Au_Obj_t) *Au_Obj_t { return Au_ObjFanin(o, 0) }

func Au_ObjFaninC0(o *Au_Obj_t) int { return Au_ObjFaninC(o, 0) }

func Au_ObjGetXsimFan0(o *Au_Obj_t) int { return auXsimFanin(o, 0) }

func Au_ObjFaninId1(o *Au_Obj_t) int { return Au_ObjFaninId(o, 1) }

func Au_ObjFanin1(o *Au_Obj_t) *Au_Obj_t { return Au_ObjFanin(o, 1) }

func Au_ObjFaninC1(o *Au_Obj_t) int { return Au_ObjFaninC(o, 1) }

func Au_ObjGetXsimFan1(o *Au_Obj_t) int { return auXsimFanin(o, 1) }

func Au_ObjFaninId2(o *Au_Obj_t) int { return Au_ObjFaninId(o, 2) }

func Au_ObjFanin2(o *Au_Obj_t) *Au_Obj_t { return Au_ObjFanin(o, 2) }

func Au_ObjFaninC2(o *Au_Obj_t) int { return Au_ObjFaninC(o, 2) }

func Au_ObjGetXsimFan2(o *Au_Obj_t) int { return auXsimFanin(o, 2) }
func Au_ObjSetFaninLit(o *Au_Obj_t, i, lit int) {
	o = Au_Regular(o)
	if lit < 0 || o.Fanins[i] != 0 {
		panic("invalid AU fanin assignment")
	}
	o.Fanins[i] = lit
}
func Au_ObjSetFanin(o *Au_Obj_t, i, id int) { Au_ObjSetFaninLit(o, i, id*2) }
func Au_ObjSetFanout(o *Au_Obj_t, i, id int) {
	o = Au_Regular(o)
	if o.Type != AU_OBJ_BOX || id <= 0 || o.fanouts[i] != 0 {
		panic("invalid AU fanout assignment")
	}
	o.fanouts[i] = id
}
func Au_ObjSetCopy(o *Au_Obj_t, value int) { Au_Regular(o).copy = value }
func Au_ObjSetXsim(o *Au_Obj_t, value int) {
	if value < AU_VAL0 || value > AU_VALX {
		panic("invalid ternary value")
	}
	Au_Regular(o).Value = value
}
func Au_NtkRemapNum(ids []int, number int) int {
	if number < 0 || number>>1 >= len(ids) {
		panic("undefined CBLIF signal")
	}
	return ids[number>>1]*2 + number&1
}
func Au_NtkParseCBlifNum(fanins *[]int, token string, ids []int) {
	parts := strings.FieldsFunc(token, func(c rune) bool { return c == ':' || c == '*' })
	if len(parts) == 0 || len(parts) > 2 {
		panic("invalid CBLIF range")
	}
	number, err := strconv.Atoi(parts[0])
	if err != nil || number < 0 {
		panic("invalid CBLIF literal")
	}
	count := 1
	if len(parts) == 2 {
		count, err = strconv.Atoi(parts[1])
		if err != nil || count < 0 {
			panic("invalid CBLIF range count")
		}
	}
	for i := 0; i < count; i++ {
		lit := number
		if strings.Contains(token, ":") {
			lit += i * 2
		}
		*fanins = append(*fanins, Au_NtkRemapNum(ids, lit))
	}
}

// ParseCBlif reports malformed directives with their original line number.
// Model references are resolved only after parsing all definitions.
func ParseCBlif(reader io.Reader, name string) (root *Au_Ntk_t, err error) {
	m := Au_ManAlloc(name)
	defer func() {
		if failure := recover(); failure != nil {
			root = nil
			err = fmt.Errorf("CBLIF: %v", failure)
		}
	}()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	var n *Au_Ntk_t
	var ids []int
	unresolved := map[*Au_Obj_t]string{}
	line := 0
	fail := func(message string) { panic(fmt.Sprintf("line %d: %s", line, message)) }
	integer := func(s string) int {
		v, e := strconv.Atoi(s)
		if e != nil || v < 0 {
			fail("invalid nonnegative integer")
		}
		return v
	}
	for scanner.Scan() {
		line++
		text := strings.SplitN(scanner.Text(), "#", 2)[0]
		fields := strings.Fields(text)
		if len(fields) == 0 {
			continue
		}
		directive := fields[0]
		args := fields[1:]
		if directive == ".model" {
			if len(args) != 1 {
				fail("model requires one name")
			}
			n = Au_NtkAlloc(m, args[0])
			ids = []int{Au_NtkCreateConst0(n)}
			if root == nil {
				root = n
			}
			continue
		}
		if n == nil {
			fail("directive outside model")
		}
		switch directive {
		case ".inputs":
			if len(args) != 1 {
				fail("inputs requires a count")
			}
			for i := 0; i < integer(args[0]); i++ {
				ids = append(ids, Au_NtkCreatePi(n))
			}
		case ".and", ".xor", ".mux":
			arity, function := 2, 1
			if directive == ".xor" {
				function = 2
			}
			if directive == ".mux" {
				arity, function = 3, 3
			}
			if len(args) != arity {
				fail("gate arity mismatch")
			}
			fanins := make([]int, arity)
			for i, s := range args {
				fanins[i] = Au_NtkRemapNum(ids, integer(s))
			}
			ids = append(ids, Au_NtkCreateNode(n, fanins, function))
		case ".outputs":
			if len(args) < 1 {
				fail("outputs requires a count")
			}
			var fanins []int
			for _, arg := range args[1:] {
				Au_NtkParseCBlifNum(&fanins, arg, ids)
			}
			if len(fanins) != integer(args[0]) {
				fail("output count mismatch")
			}
			for _, lit := range fanins {
				Au_NtkCreatePo(n, lit)
			}
		case ".subckt":
			if len(args) < 3 {
				fail("subckt requires name and port counts")
			}
			var fanins []int
			for _, arg := range args[3:] {
				Au_NtkParseCBlifNum(&fanins, arg, ids)
			}
			if len(fanins) != integer(args[1]) {
				fail("subckt input count mismatch")
			}
			id := Au_NtkCreateBox(n, fanins, integer(args[2]), 1)
			box := n.Objects[id]
			unresolved[box] = args[0]
			ids = append(ids, box.fanouts...)
		case ".end":
			if len(args) != 0 {
				fail("unexpected end arguments")
			}
			n = nil
		default:
			fail("unknown directive " + directive)
		}
	}
	if e := scanner.Err(); e != nil {
		return nil, e
	}
	if root == nil {
		return nil, fmt.Errorf("empty CBLIF design")
	}
	for box, model := range unresolved {
		found := Au_ManFindNtk(m, model)
		if found < 0 {
			return nil, fmt.Errorf("undefined model %q", model)
		}
		child := m.Networks[found]
		if len(box.Fanins) != len(child.PIs) || len(box.fanouts) != len(child.POs) {
			return nil, fmt.Errorf("model %q port count mismatch", model)
		}
		box.Func = found
		for _, id := range box.fanouts {
			box.network.Objects[id].Func = found
		}
	}
	Au_ManReorderModels(m, root)
	return root, nil
}
func Au_NtkParseCBlif(filename string) *Au_Ntk_t {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()
	root, err := ParseCBlif(file, filename)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return root
}

// AU expansion keeps one signal environment per instance, so repeated models
// and complemented actual arguments do not share stale copy values.
func auExpand(g *aig.Graph, n *Au_Ntk_t, inputs []aig.Edge, active map[*Au_Ntk_t]bool) []aig.Edge {
	if len(inputs) != len(n.PIs) {
		panic("AU model input count mismatch")
	}
	if n.Name == "ref_egcd" {
		outputs := make([]aig.Edge, len(n.POs))
		for i := range outputs {
			outputs[i] = g.NewInput()
		}
		return outputs
	}
	if active[n] {
		panic("recursive AU model")
	}
	active[n] = true
	defer delete(active, n)
	values := map[int]aig.Edge{}
	visiting := map[int]bool{}
	for i, id := range n.PIs {
		values[id] = inputs[i]
		n.Objects[id].copy = int(inputs[i])
	}
	var read func(int) aig.Edge
	read = func(lit int) aig.Edge {
		id := lit >> 1
		value, exists := values[id]
		if !exists {
			if visiting[id] {
				panic("cyclic AU net")
			}
			visiting[id] = true
			o := Au_NtkObj(n, id)
			args := func() []aig.Edge {
				v := make([]aig.Edge, len(o.Fanins))
				for i, f := range o.Fanins {
					v[i] = read(f)
				}
				return v
			}
			switch o.Type {
			case AU_OBJ_CONST0:
				value = 1
			case AU_OBJ_PO:
				value = read(o.Fanins[0])
			case AU_OBJ_NODE:
				fanins := args()
				if len(n.Man.Functions) > 0 {
					if o.Func <= 0 || o.Func >= len(n.Man.Functions) {
						panic("undefined AU SOP")
					}
					value = isoCover(g, n.Man.Functions[o.Func], fanins)
				} else {
					switch o.Func {
					case 1:
						if len(fanins) != 2 {
							panic("AU AND arity")
						}
						value = g.AND(fanins[0], fanins[1])
					case 2:
						if len(fanins) != 2 {
							panic("AU XOR arity")
						}
						value = g.XOR(fanins[0], fanins[1])
					case 3:
						if len(fanins) != 3 {
							panic("AU MUX arity")
						}
						value = g.Mux(fanins[0], fanins[1], fanins[2])
					default:
						panic("unknown AU function")
					}
				}
			case AU_OBJ_FAN:
				box := Au_ObjFanin0(o)
				child := Au_ObjModel(box)
				outputs := auExpand(g, child, func() []aig.Edge {
					v := make([]aig.Edge, len(box.Fanins))
					for i, f := range box.Fanins {
						v[i] = read(f)
					}
					return v
				}(), active)
				if len(outputs) != len(box.fanouts) {
					panic("AU box output mismatch")
				}
				for i, fan := range box.fanouts {
					values[fan] = outputs[i]
					n.Objects[fan].copy = int(outputs[i])
				}
				value = values[id]
			case AU_OBJ_PI:
				panic("unbound AU input")
			default:
				panic("unsupported AU object")
			}
			values[id] = value
			o.copy = int(value)
			delete(visiting, id)
		}
		if lit&1 != 0 {
			return value.Not()
		}
		return value
	}
	outputs := make([]aig.Edge, len(n.POs))
	for i, id := range n.POs {
		outputs[i] = read(id * 2)
	}
	return outputs
}
func Au_NtkDeriveFlatGia_rec(g *aig.Graph, n *Au_Ntk_t) {
	inputs := make([]aig.Edge, len(n.PIs))
	for i, id := range n.PIs {
		if n.Objects[id].copy < 0 {
			panic("unassigned AU PI")
		}
		inputs[i] = aig.Edge(n.Objects[id].copy)
	}
	auExpand(g, n, inputs, map[*Au_Ntk_t]bool{})
}
func Au_NtkDeriveFlatGia(n *Au_Ntk_t) *aig.Graph {
	g := aig.NewGraph()
	g.Name = n.Name
	inputs := make([]aig.Edge, len(n.PIs))
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	outputs := auExpand(g, n, inputs, map[*Au_Ntk_t]bool{})
	for _, output := range outputs {
		g.AddOutput(output)
	}
	return aig.Aig_ManCleanup(g)
}

const (
	AU_VAL0 = 1
	AU_VAL1 = 2
	AU_VALX = 3
)

func auValidXsim(v int) {
	if v < AU_VAL0 || v > AU_VALX {
		panic("invalid AU ternary value")
	}
}
func Au_XsimInv(v int) int {
	auValidXsim(v)
	if v == AU_VAL0 {
		return AU_VAL1
	}
	if v == AU_VAL1 {
		return AU_VAL0
	}
	return AU_VALX
}
func Au_XsimAnd(a, b int) int {
	auValidXsim(a)
	auValidXsim(b)
	if a == AU_VAL0 || b == AU_VAL0 {
		return AU_VAL0
	}
	if a == AU_VALX || b == AU_VALX {
		return AU_VALX
	}
	return AU_VAL1
}
func Au_XsimXor(a, b int) int {
	auValidXsim(a)
	auValidXsim(b)
	if a == AU_VALX || b == AU_VALX {
		return AU_VALX
	}
	if a == b {
		return AU_VAL0
	}
	return AU_VAL1
}
func Au_XsimMux(c, t, e int) int {
	auValidXsim(c)
	auValidXsim(t)
	auValidXsim(e)
	if c == AU_VAL0 {
		return e
	}
	if c == AU_VAL1 {
		return t
	}
	if t == e {
		return t
	}
	return AU_VALX
}
func auXsimFanin(o *Au_Obj_t, i int) int {
	v := Au_ObjFanin(o, i).Value
	if Au_ObjFaninC(o, i) != 0 {
		return Au_XsimInv(v)
	}
	return v
}
func auSimulate(n *Au_Ntk_t, inputs []int, active map[*Au_Ntk_t]bool) []int {
	if len(inputs) != len(n.PIs) {
		panic("AU simulation port mismatch")
	}
	if active[n] && n.Name != "ref_egcd" {
		panic("recursive AU model")
	}
	values := map[int]int{}
	busy := map[int]bool{}
	for i, id := range n.PIs {
		auValidXsim(inputs[i])
		values[id] = inputs[i]
		n.Objects[id].Value = inputs[i]
		switch inputs[i] {
		case AU_VAL0:
			n.Man.PortsC0++
		case AU_VAL1:
			n.Man.PortsC1++
		case AU_VALX:
			n.Man.PortsNC++
		}
	}
	if n.Name == "ref_egcd" {
		out := make([]int, len(n.POs))
		for i, id := range n.POs {
			out[i] = AU_VALX
			n.Objects[id].Value = AU_VALX
		}
		return out
	}
	active[n] = true
	defer delete(active, n)
	var read func(int) int
	read = func(lit int) int {
		id := lit >> 1
		value, ok := values[id]
		if !ok {
			if busy[id] {
				panic("cyclic AU simulation net")
			}
			busy[id] = true
			o := n.Objects[id]
			switch o.Type {
			case AU_OBJ_CONST0:
				value = AU_VAL0
			case AU_OBJ_PO:
				value = read(o.Fanins[0])
			case AU_OBJ_NODE:
				v := make([]int, len(o.Fanins))
				for i, f := range o.Fanins {
					v[i] = read(f)
				}
				switch o.Func {
				case 1:
					value = Au_XsimAnd(v[0], v[1])
				case 2:
					value = Au_XsimXor(v[0], v[1])
				case 3:
					value = Au_XsimMux(v[0], v[1], v[2])
				default:
					panic("unknown AU simulation function")
				}
			case AU_OBJ_FAN:
				box := Au_ObjFanin0(o)
				v := make([]int, len(box.Fanins))
				for i, f := range box.Fanins {
					v[i] = read(f)
				}
				results := auSimulate(Au_ObjModel(box), v, active)
				for i, id := range box.fanouts {
					values[id] = results[i]
					n.Objects[id].Value = results[i]
				}
				value = values[id]
			default:
				panic("unsupported AU simulation object")
			}
			delete(busy, id)
			values[id] = value
			o.Value = value
		}
		if lit&1 != 0 {
			return Au_XsimInv(value)
		}
		return value
	}
	outputs := make([]int, len(n.POs))
	for i, id := range n.POs {
		outputs[i] = read(id * 2)
		switch outputs[i] {
		case AU_VAL0:
			n.Man.PortsC0++
		case AU_VAL1:
			n.Man.PortsC1++
		case AU_VALX:
			n.Man.PortsNC++
		}
	}
	return outputs
}
func Au_NtkTerSimulate_rec(n *Au_Ntk_t) {
	inputs := make([]int, len(n.PIs))
	for i, id := range n.PIs {
		inputs[i] = n.Objects[id].Value
	}
	auSimulate(n, inputs, map[*Au_Ntk_t]bool{})
}
func Au_NtkTerSimulate(n *Au_Ntk_t) {
	if len(n.Man.Functions) > 0 {
		panic("AU ternary simulation needs primitive functions")
	}
	n.Man.PortsC0, n.Man.PortsC1, n.Man.PortsNC = 0, 0, 0
	for _, id := range n.PIs {
		n.Objects[id].Value = AU_VALX
	}
	Au_NtkTerSimulate_rec(n)
	fmt.Printf("Const0 ports = %.0f. Const1 ports = %.0f. Non-const ports = %.0f.\n", n.Man.PortsC0, n.Man.PortsC1, n.Man.PortsNC)
}
func (m *Au_Man_t) intern(sop string) int {
	if m.functionIDs == nil {
		m.functionIDs = map[string]int{}
		m.Functions = []string{""}
	}
	if id, ok := m.functionIDs[sop]; ok {
		return id
	}
	id := len(m.Functions)
	m.Functions = append(m.Functions, sop)
	m.functionIDs[sop] = id
	return id
}
func Au_NtkDerive(m *Au_Man_t, old *Network, order []*Object) *Au_Ntk_t {
	n := Au_NtkAlloc(m, old.Name)
	constant := Au_NtkCreateConst0(n)
	values := map[*Object]int{old.constant: constant*2 + 1}
	set := func(o *Object, lit int) {
		values[o] = lit
		o.Temp = lit >> 1
		for _, net := range o.fanouts {
			if net.Kind == ObjectNet && len(net.fanins) == 1 {
				values[net] = Au_LitNotCond(lit, int(abcBool(net.fanins[0].Complement)))
				net.Temp = values[net] >> 1
			}
		}
	}
	read := func(e Edge) int {
		lit, ok := values[e.Object]
		if !ok {
			panic("AU derivation dependency order")
		}
		return Au_LitNotCond(lit, int(abcBool(e.Complement)))
	}
	set(old.constant, constant*2+1)
	for _, pi := range old.ObjectsOfKind(ObjectPI) {
		set(pi, Au_NtkCreatePi(n)*2)
	}
	for _, o := range order {
		if o.Kind == ObjectNode {
			fanins := make([]int, len(o.fanins))
			for i, e := range o.fanins {
				fanins[i] = read(e)
			}
			sop, ok := o.Data.(string)
			if !ok {
				panic("AU derivation needs SOP logic")
			}
			set(o, Au_NtkCreateNode(n, fanins, m.intern(sop))*2)
		} else {
			child := boxModel(o)
			model := Au_ManFindNtk(m, child.Name)
			if model <= 0 {
				panic("AU child model not derived")
			}
			fanins := make([]int, len(o.fanins))
			for i, e := range o.fanins {
				bi := e.Object
				if len(bi.fanins) != 1 {
					panic("AU BI arity")
				}
				fanins[i] = Au_LitNotCond(read(bi.fanins[0]), int(abcBool(e.Complement)))
			}
			id := Au_NtkCreateBox(n, fanins, len(o.fanouts), model)
			box := n.Objects[id]
			for i, bo := range o.fanouts {
				set(bo, Au_LitNotCond(box.fanouts[i]*2, int(abcBool(bo.fanins[0].Complement))))
			}
		}
	}
	for _, po := range old.ObjectsOfKind(ObjectPO) {
		if len(po.fanins) != 1 {
			panic("AU PO arity")
		}
		Au_NtkCreatePo(n, read(po.fanins[0]))
	}
	return n
}
func Au_ManDeriveTest(root *Network) *aig.Graph {
	m := Au_ManAlloc(root.Name)
	defer Au_ManDelete(m)
	var target *Au_Ntk_t
	for _, model := range Abc_NtkCollectHie(root) {
		derived := Au_NtkDerive(m, model, Abc_NtkDfsBoxes(model))
		if model == root {
			target = derived
		}
	}
	Au_ManReorderModels(m, target)
	// Retain the source's optional experimental model selection, with root fallback.
	if selected := Au_ManFindNtkP(m, "path_0_r_x_lhs"); selected != nil {
		target = selected
	}
	return Au_NtkDeriveFlatGia(target)
}
func Abc_NtkHieCecTest2(filename, model string, verbose int32) *aig.Graph {
	root := Au_NtkParseCBlif(filename)
	if root == nil {
		return nil
	}
	defer Au_ManDelete(root.Man)
	if model != "" {
		if selected := Au_ManFindNtkP(root.Man, model); selected != nil {
			root = selected
		}
	}
	if verbose != 0 {
		Au_ManPrintBoxInfo(root)
		Au_ManPrintStats(root.Man)
	}
	return Au_NtkDeriveFlatGia(root)
}
