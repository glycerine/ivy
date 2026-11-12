package abc

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

// MVVariable owns optional symbolic value labels; an absent variable is binary.
type MVVariable struct {
	Values int
	Names  []string
}

func Abc_NtkStartMvVars(n *Network) {
	if n.mvVars != nil {
		panic("MV attributes already started")
	}
	n.mvVars = make(map[*Object]*MVVariable)
}
func Abc_NtkFreeMvVars(n *Network) { n.mvVars = nil }
func Abc_NtkSetMvVarValues(o *Object, count int32) {
	if count < 2 {
		panic("MV domain must have at least two values")
	}
	if count == 2 || o.network.mvVars[o] != nil {
		return
	}
	if o.network.mvVars == nil {
		Abc_NtkStartMvVars(o.network)
	}
	o.network.mvVars[o] = &MVVariable{Values: int(count)}
}
func Abc_ObjMvVar(o *Object) *MVVariable { return o.network.mvVars[o] }
func Abc_ObjMvVarNum(o *Object) int32 {
	if v := Abc_ObjMvVar(o); v != nil {
		return int32(v.Values)
	}
	return 2
}
func Abc_StringGetNumber(text *string) int32 {
	end := 0
	for end < len(*text) && (*text)[end] >= '0' && (*text)[end] <= '9' {
		end++
	}
	if end == 0 {
		panic("expected decimal number")
	}
	value, err := strconv.ParseInt((*text)[:end], 10, 32)
	if err != nil {
		panic(err)
	}
	*text = (*text)[end:]
	return int32(value)
}
func nodeMVTable(o *Object) (*MVTable, error) {
	domains := make([]int, len(o.fanins)+1)
	for i, e := range o.fanins {
		domains[i] = int(Abc_ObjMvVarNum(e.Object))
	}
	if len(o.fanouts) != 1 {
		return nil, fmt.Errorf("MV node must drive one net")
	}
	domains[len(o.fanins)] = int(Abc_ObjMvVarNum(o.fanouts[0]))
	switch data := o.Data.(type) {
	case *MVTable:
		if len(data.DomainSizes) != len(domains) {
			return nil, fmt.Errorf("MV table arity mismatch")
		}
		for i, size := range domains {
			if data.DomainSizes[i] != size {
				return nil, fmt.Errorf("MV table domain mismatch")
			}
		}
		return ParseMVTable(data.String(), domains)
	case string:
		return ParseMVTable(data, domains)
	default:
		return nil, fmt.Errorf("node has no normalized MV cover")
	}
}
func mvTermSignal(m *Abc_Aig_t, term MVTerm, index int, values [][]Edge, domain int) Edge {
	result := m.Network.One().Not()
	if term.Any {
		return m.Network.One()
	}
	if term.EqualInput >= 0 {
		for value := 0; value < domain; value++ {
			result = Abc_AigOr(m, result, Abc_AigAnd(m, values[index][value], values[term.EqualInput][value]))
		}
		return result
	}
	for _, value := range term.Values {
		result = Abc_AigOr(m, result, values[index][value])
	}
	return result
}
func Abc_NodeStrashBlifMv(n *Network, o *Object) int32 {
	table, err := nodeMVTable(o)
	if err != nil {
		return 0
	}
	m := Abc_AigAlloc(n)
	zero := n.One().Not()
	size := table.DomainSizes[len(table.DomainSizes)-1]
	result := make([]Edge, size)
	for i := range result {
		result[i] = zero
	}
	values := make([][]Edge, len(o.fanins))
	for i, e := range o.fanins {
		values[i] = e.Object.MVEncoding
		if len(values[i]) != table.DomainSizes[i] {
			return 0
		}
	}
	if len(o.fanins) == 0 && len(table.Cubes) > 0 && table.Cubes[0].Output.Any {
		if size == 2 {
			pi := n.NewObject(ObjectPI)
			pi.SetName(fmt.Sprintf("free_var_n%d", pi.ID))
			result[1] = pi.Signal()
			result[0] = result[1].Not()
		} else {
			result[0] = n.One()
		}
		o.fanouts[0].MVEncoding = result
		return 1
	}
	for _, cube := range table.Cubes {
		condition := n.One()
		for i, input := range cube.Inputs {
			condition = Abc_AigAnd(m, condition, mvTermSignal(m, input, i, values, table.DomainSizes[i]))
		}
		for value := 0; value < size; value++ {
			factor := n.One()
			if cube.Output.EqualInput >= 0 {
				factor = values[cube.Output.EqualInput][value]
			} else if !termMatches(cube.Output, value, make([]int, len(values))) {
				continue
			}
			result[value] = Abc_AigOr(m, result[value], Abc_AigAnd(m, condition, factor))
		}
	}
	if table.Default != nil {
		def := *table.Default
		missing := n.One()
		for value, signal := range result {
			if def.EqualInput < 0 && !def.Any && len(def.Values) == 1 && def.Values[0] == value {
				continue
			}
			missing = Abc_AigAnd(m, missing, signal.Not())
		}
		if def.EqualInput >= 0 {
			for value := range result {
				result[value] = Abc_AigOr(m, result[value], Abc_AigAnd(m, missing, values[def.EqualInput][value]))
			}
		} else {
			for value := range result {
				if termMatches(def, value, nil) {
					result[value] = missing
				}
			}
		}
	}
	o.fanouts[0].MVEncoding = result
	return 1
}
func Abc_NtkConvertAssignName(o, net *Object, index int32) {
	o.SetName(fmt.Sprintf("%s[%d]", net.Name(), index))
}
func mvBits(size int) int { return bits.Len(uint(size - 1)) }
func mvNet(o *Object, input bool) *Object {
	if input {
		if len(o.fanouts) != 1 || o.fanouts[0].Kind != ObjectNet {
			panic("MV CI requires one net")
		}
		return o.fanouts[0]
	}
	if len(o.fanins) != 1 || o.fanins[0].Object.Kind != ObjectNet {
		panic("MV CO requires one net")
	}
	return o.fanins[0].Object
}
func Abc_NtkStrashBlifMv(old *Network) *Network {
	for _, box := range old.Boxes() {
		if box.Kind != ObjectLatch {
			return nil
		}
	}
	order, err := old.Topological(old.Objects())
	if err != nil {
		return nil
	}
	defer func() {
		for _, net := range old.ObjectsOfKind(ObjectNet) {
			net.MVEncoding = nil
		}
	}()
	n := NewNetwork(old.Name)
	n.Spec = old.Spec
	m := Abc_AigAlloc(n)
	maxValues := 2
	for _, net := range old.ObjectsOfKind(ObjectNet) {
		if size := int(Abc_ObjMvVarNum(net)); size > maxValues {
			maxValues = size
		}
		net.MVEncoding = nil
	}
	var registerInputs, registerOutputs []*Object
	for _, kind := range []ObjectKind{ObjectPI, ObjectBO} {
		for _, ci := range old.ObjectsOfKind(kind) {
			net := mvNet(ci, true)
			size := int(Abc_ObjMvVarNum(net))
			encodingBits := make([]Edge, mvBits(size))
			for bit := range encodingBits {
				o := n.NewObject(kind)
				encodingBits[bit] = o.Signal()
				if maxValues == 2 {
					o.SetName(net.Name())
				} else {
					Abc_NtkConvertAssignName(o, net, int32(bit))
				}
				if kind == ObjectBO {
					registerOutputs = append(registerOutputs, o)
				}
			}
			net.MVEncoding = make([]Edge, size)
			for value := 0; value < size; value++ {
				signal := n.One()
				for bit, input := range encodingBits {
					if value&(1<<bit) == 0 {
						input = input.Not()
					}
					signal = Abc_AigAnd(m, signal, input)
				}
				net.MVEncoding[value] = signal
			}
		}
	}
	for _, o := range order {
		if o.Kind == ObjectNode && Abc_NodeStrashBlifMv(n, o) == 0 {
			return nil
		}
	}
	for _, kind := range []ObjectKind{ObjectPO, ObjectBI} {
		for _, co := range old.ObjectsOfKind(kind) {
			net := mvNet(co, false)
			size := int(Abc_ObjMvVarNum(net))
			if len(net.MVEncoding) != size {
				return nil
			}
			for bit := 0; bit < mvBits(size); bit++ {
				signal := n.One().Not()
				for value, input := range net.MVEncoding {
					if value&(1<<bit) != 0 {
						signal = Abc_AigOr(m, signal, input)
					}
				}
				o := n.NewObject(kind)
				Abc_ObjAddFanin(o, signal)
				if maxValues == 2 {
					o.SetName(net.Name())
				} else {
					Abc_NtkConvertAssignName(o, net, int32(bit))
				}
				if kind == ObjectBI {
					registerInputs = append(registerInputs, o)
				}
			}
		}
	}
	if len(registerInputs) != len(registerOutputs) {
		return nil
	}
	for i, input := range registerInputs {
		latch := n.NewObject(ObjectLatch)
		latch.Data = ABC_INIT_ZERO
		Abc_ObjAddFanin(latch, input)
		Abc_ObjAddFanin(registerOutputs[i], latch)
	}
	if len(registerInputs) > 0 {
		var ids []int
		for _, pi := range n.ObjectsOfKind(ObjectPI) {
			if strings.HasPrefix(pi.Name(), "free_var_") {
				ids = append(ids, pi.ID)
			}
		}
		for _, pi := range n.ObjectsOfKind(ObjectPI) {
			if !strings.HasPrefix(pi.Name(), "free_var_") {
				ids = append(ids, pi.ID)
			}
		}
		if err := n.SetTerminalOrder(ObjectPI, ids); err != nil {
			panic(err)
		}
		n.cis = append(n.ObjectsOfKind(ObjectPI), n.ObjectsOfKind(ObjectBO)...)
	}
	for _, net := range old.ObjectsOfKind(ObjectNet) {
		net.MVEncoding = nil
	}
	Abc_AigCleanup(m)
	if n.Check() != nil {
		return nil
	}
	return n
}

// Compact MV covers below encode each binary input by one char, with the last
// char the output and newline the cube terminator. This is the original cost API.
func Abc_NodeConvertSopToMvSop(nVars int32, off, on []uint32) string {
	if nVars < 0 || nVars > 16 {
		panic("packed cubes support zero to sixteen variables")
	}
	var out strings.Builder
	if len(off) == 0 || len(on) == 0 {
		out.WriteString(strings.Repeat("-", int(nVars)))
		if len(on) > 0 {
			out.WriteByte('1')
		} else {
			out.WriteByte('0')
		}
		out.WriteByte('\n')
		return out.String()
	}
	for polarity, cubes := range [][]uint32{off, on} {
		for _, cube := range cubes {
			for variable := int32(0); variable < nVars; variable++ {
				switch (cube >> (2 * variable)) & 3 {
				case 0:
					out.WriteByte('-')
				case 1:
					out.WriteByte('0')
				case 2:
					out.WriteByte('1')
				default:
					panic("invalid packed cube literal")
				}
			}
			out.WriteByte(byte('0' + polarity))
			out.WriteByte('\n')
		}
	}
	return out.String()
}
func Abc_NodeEvalMvCostInternal(nVars int32, values []int32, cover string) int32 {
	if nVars < 0 || len(values) != int(nVars)+1 {
		panic("MV cost domain arity mismatch")
	}
	return int32(strings.Count(strings.TrimRight(cover, "\x00"), "\n"))
}
func Abc_NodeEvalMvCost(nVars int32, off, on []uint32) int32 {
	values := make([]int32, nVars+1)
	for i := range values {
		values[i] = 2
	}
	return Abc_NodeEvalMvCostInternal(nVars, values, Abc_NodeConvertSopToMvSop(nVars, off, on))
}

func copyMVVariable(from, to *Object) {
	if variable := Abc_ObjMvVar(from); variable != nil {
		if to.network.mvVars == nil {
			Abc_NtkStartMvVars(to.network)
		}
		to.network.mvVars[to] = &MVVariable{Values: variable.Values, Names: append([]string(nil), variable.Names...)}
	}
}
func cloneBoundaryNet(n *Network, old *Object) *Object {
	if copy := getCopy(old); copy.Object != nil {
		return copy.Object
	}
	net := n.NewObject(ObjectNet)
	if name, ok := old.network.names.Lookup(int32(old.ID)); ok {
		net.SetName(name)
	}
	copyMVVariable(old, net)
	setCopy(old, net.Signal())
	return net
}
func Abc_NtkSkeletonBlifMv(old *Network) *Network {
	for _, box := range old.Boxes() {
		if box.Kind != ObjectLatch {
			return nil
		}
	}
	clearCopies(old)
	n := copyNetworkMetadata(old)
	n.Barriers = 0
	n.Design = nil
	box := n.NewObject(ObjectWhitebox)
	for _, pi := range old.ObjectsOfKind(ObjectPI) {
		dup := Abc_NtkDupObj(n, pi, 0)
		net := cloneBoundaryNet(n, mvNet(pi, true))
		Abc_ObjAddFanin(net, dup)
	}
	for _, po := range old.ObjectsOfKind(ObjectPO) {
		dup := Abc_NtkDupObj(n, po, 0)
		Abc_ObjAddFanin(dup, cloneBoundaryNet(n, mvNet(po, false)))
	}
	for _, latch := range old.ObjectsOfKind(ObjectLatch) {
		if len(latch.fanins) != 1 || len(latch.fanouts) != 1 {
			return nil
		}
		dup := Abc_NtkDupBox(n, latch, 0)
		in, out := latch.fanins[0].Object, latch.fanouts[0]
		Abc_ObjAddFanin(dup.fanins[0].Object, cloneBoundaryNet(n, mvNet(in, false)))
		Abc_ObjAddFanin(cloneBoundaryNet(n, mvNet(out, true)), dup.fanouts[0])
	}
	encoded := map[*Object]bool{}
	for _, ci := range old.Inputs() {
		net := mvNet(ci, true)
		size := int(Abc_ObjMvVarNum(net))
		encoded[net] = true
		for bit := 0; bit < mvBits(size); bit++ {
			zero := MVTerm{EqualInput: -1, Values: []int{0}}
			table := &MVTable{DomainSizes: []int{size, 2}, Default: &zero}
			for value := 0; value < size; value++ {
				if value&(1<<bit) != 0 {
					table.Cubes = append(table.Cubes, MVCube{Inputs: []MVTerm{{EqualInput: -1, Values: []int{value}}}, Output: MVTerm{EqualInput: -1, Values: []int{1}}})
				}
			}
			node := n.NewObject(ObjectNode)
			node.Data = table
			Abc_ObjAddFanin(node, getCopy(net))
			bitNet := netForDriver(n, node.Signal())
			term := n.NewObject(ObjectBI)
			Abc_ObjAddFanin(term, bitNet)
			Abc_ObjAddFanin(box, term)
		}
	}
	for _, co := range old.Outputs() {
		net := mvNet(co, false)
		if encoded[net] {
			continue
		}
		encoded[net] = true
		size := int(Abc_ObjMvVarNum(net))
		count := mvBits(size)
		zero := MVTerm{EqualInput: -1, Values: []int{0}}
		domains := make([]int, count+1)
		for i := range domains {
			domains[i] = 2
		}
		domains[count] = size
		table := &MVTable{DomainSizes: domains, Default: &zero}
		for value := 0; value < size; value++ {
			cube := MVCube{Output: MVTerm{EqualInput: -1, Values: []int{value}}}
			for bit := 0; bit < count; bit++ {
				cube.Inputs = append(cube.Inputs, MVTerm{EqualInput: -1, Values: []int{(value >> bit) & 1}})
			}
			table.Cubes = append(table.Cubes, cube)
		}
		node := n.NewObject(ObjectNode)
		node.Data = table
		for bit := 0; bit < count; bit++ {
			term := n.NewObject(ObjectBO)
			Abc_ObjAddFanin(term, box)
			Abc_ObjAddFanin(node, netForDriver(n, term.Signal()))
		}
		Abc_ObjAddFanin(getCopy(net).Object, node)
	}
	if err := n.Check(); err != nil {
		return nil
	}
	return n
}
func Abc_NtkInsertBlifMv(base, logic *Network) *Network {
	if len(logic.Boxes()) != 0 {
		return nil
	}
	skeleton := Abc_NtkSkeletonBlifMv(base)
	if skeleton == nil {
		return nil
	}
	box := skeleton.ObjectsOfKind(ObjectWhitebox)[0]
	pis, pos := logic.ObjectsOfKind(ObjectPI), logic.ObjectsOfKind(ObjectPO)
	if len(box.fanins) != len(pis) || len(box.fanouts) != len(pos) {
		return nil
	}
	order, err := logic.Topological(logic.Objects())
	if err != nil {
		return nil
	}
	clearCopies(logic)
	setCopy(logic.constant, skeleton.One())
	for i, pi := range pis {
		net := mvNet(pi, true)
		setCopy(net, box.fanins[i].Object.fanins[0])
		setCopy(pi, getCopy(net))
	}
	for _, o := range order {
		if o.Kind == ObjectNode {
			dup := Abc_NtkDupObj(skeleton, o, 0)
			for _, fanin := range o.fanins {
				signal := getCopy(fanin.Object)
				if signal.Object == nil {
					return nil
				}
				signal.Complement = signal.Complement != fanin.Complement
				Abc_ObjAddFanin(dup, signal)
			}
			setCopy(o, dup.Signal())
		} else if o.Kind == ObjectNet && getCopy(o).Object == nil {
			if len(o.fanins) != 1 {
				return nil
			}
			driver := getCopy(o.fanins[0].Object)
			if driver.Object == nil {
				return nil
			}
			driver.Complement = driver.Complement != o.fanins[0].Complement
			net := netForDriver(skeleton, driver)
			copyMVVariable(o, net)
			setCopy(o, net.Signal())
		}
	}
	outputs := append([]*Object(nil), box.fanouts...)
	for i, term := range outputs {
		mapped := getCopy(mvNet(pos[i], false))
		if mapped.Object == nil || len(term.fanouts) != 1 {
			return nil
		}
		net := term.fanouts[0]
		for _, consumer := range net.Fanouts() {
			Abc_ObjPatchFanin(consumer, net, mapped)
		}
		Abc_NtkDeleteObj(net)
		Abc_NtkDeleteObj(term)
	}
	inputs := append([]Edge(nil), box.fanins...)
	for _, term := range inputs {
		Abc_NtkDeleteObj(term.Object)
	}
	Abc_NtkDeleteObj(box)
	if err := skeleton.Check(); err != nil {
		return nil
	}
	return skeleton
}
func boolCoverToMV(text string, nVars int) (*MVTable, error) {
	domains := make([]int, nVars+1)
	for i := range domains {
		domains[i] = 2
	}
	table := &MVTable{DomainSizes: domains}
	polarity := -1
	for _, line := range strings.Split(strings.TrimRight(text, "\x00"), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pattern, output := "", ""
		if nVars == 0 && len(fields) == 1 {
			output = fields[0]
		} else if len(fields) == 2 {
			pattern, output = fields[0], fields[1]
		} else {
			return nil, fmt.Errorf("invalid SOP row")
		}
		if len(pattern) != nVars || (output != "0" && output != "1") {
			return nil, fmt.Errorf("invalid SOP arity/polarity")
		}
		value := int(output[0] - '0')
		if polarity >= 0 && polarity != value {
			return nil, fmt.Errorf("mixed SOP polarities")
		}
		polarity = value
		cube := MVCube{Output: MVTerm{EqualInput: -1, Values: []int{value}}}
		for _, literal := range pattern {
			term := MVTerm{EqualInput: -1}
			switch literal {
			case '-':
				term.Any = true
			case '0', '1':
				term.Values = []int{int(literal - '0')}
			default:
				return nil, fmt.Errorf("invalid SOP literal")
			}
			cube.Inputs = append(cube.Inputs, term)
		}
		table.Cubes = append(table.Cubes, cube)
	}
	if polarity < 0 {
		return nil, fmt.Errorf("empty SOP")
	}
	table.Default = &MVTerm{EqualInput: -1, Values: []int{1 - polarity}}
	return table, nil
}
func Abc_NtkConvertToBlifMv(n *Network) int32 {
	converted := map[*Object]*MVTable{}
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		count := len(o.fanins)
		domains := make([]int, count+1)
		for i := range domains {
			domains[i] = 2
		}
		if table, ok := o.Data.(*MVTable); ok {
			converted[o] = table
			continue
		}
		if text, ok := o.Data.(string); ok && text != aigANDMarker {
			table, err := boolCoverToMV(text, count)
			if err != nil {
				return 0
			}
			converted[o] = table
			continue
		}
		if count > 20 {
			return 0
		}
		var fn func([]bool) bool
		if f, ok := o.Data.(func([]bool) bool); ok {
			fn = f
		} else if o.Data == aigANDMarker {
			fn = func(v []bool) bool { return v[0] && v[1] }
		} else {
			return 0
		}
		zero := MVTerm{EqualInput: -1, Values: []int{0}}
		table := &MVTable{DomainSizes: domains, Default: &zero}
		values := make([]bool, count)
		for pattern := 0; pattern < (1 << count); pattern++ {
			for i := range values {
				values[i] = pattern&(1<<i) != 0
			}
			if !fn(values) {
				continue
			}
			cube := MVCube{Output: MVTerm{EqualInput: -1, Values: []int{1}}}
			for _, value := range values {
				literal := 0
				if value {
					literal = 1
				}
				cube.Inputs = append(cube.Inputs, MVTerm{EqualInput: -1, Values: []int{literal}})
			}
			table.Cubes = append(table.Cubes, cube)
		}
		converted[o] = table
	}
	for o, table := range converted {
		o.Data = table
	}
	n.Function = uint8(ABC_FUNC_BLIFMV)
	return 1
}
