package goivy

import (
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const rtlBankAddressBits = 4

type rtlMemory struct {
	symbol    *Const
	address   int
	data      int
	clock     string
	update    rtlUpdate
	updateDef *IvyDefinition
	writes    []rtlWrite
	bank      []rtlNet
}

type rtlGuard struct {
	expr   Expr
	invert bool
}

type rtlWrite struct {
	guards  []rtlGuard
	address Expr
	data    Expr
	depth   int
}

func (t *rtlTranslator) memorySymbol(symbol *Const) bool {
	return IsFunctionSort(symbol.CSort) && !t.defNames[symbol.Name] &&
		!IsInterpretedSymbol(t.mod.Sig, symbol) && !SymbolIsPolymorphic(symbol.Name) &&
		t.inDesign(t.owner(symbol.Name))
}

func (t *rtlTranslator) memories(object string) []*Const {
	symbols := make(map[string]*Const)
	for _, symbol := range t.state[object].All() {
		if IsFunctionSort(symbol.CSort) {
			symbols[symbol.Name] = symbol
		}
	}
	add := func(expr Expr) {
		for _, node := range UsedSymbolsAST(expr).All() {
			if symbol, ok := node.(*Const); ok && t.owner(symbol.Name) == object && t.memorySymbol(symbol) {
				symbols[symbol.Name] = symbol
			}
		}
	}
	for _, def := range t.defs.All() {
		if def.owner == object {
			add(def.def.Rhs)
		}
	}
	for _, update := range t.updates[object].All() {
		for _, def := range update.clauses.Defs {
			add(def.Rhs)
		}
	}
	result := make([]*Const, 0, len(symbols))
	for _, name := range sortedRTLKeys(symbols) {
		result = append(result, symbols[name])
	}
	return result
}

func (t *rtlTranslator) memoryDimensions(symbol *Const) (address, data int, err error) {
	s := symbol.CSort.(*LogicFunctionSort)
	if s.Arity() != 1 {
		return 0, 0, fmt.Errorf("RTL array %s must be one-dimensional", symbol.Name)
	}
	address, err = t.width(s.Domain()[0])
	if err != nil {
		return 0, 0, err
	}
	data, err = t.width(s.Range())
	return
}

func rtlReadAt(expr Expr, index Expr) *Const {
	call, ok := expr.(*Apply)
	if !ok || len(call.Terms) != 1 || !call.Terms[0].Equal(index) {
		return nil
	}
	return GetAppRep(call)
}

func rtlUses(expr Expr, variable Expr) bool {
	if expr.Equal(variable) {
		return true
	}
	for _, arg := range NodeArgs(expr) {
		if rtlUses(arg, variable) {
			return true
		}
	}
	return false
}

func rtlIndexTest(expr, index Expr) Expr {
	for {
		args := NodeArgs(expr)
		switch expr.(type) {
		case *LogicAnd, *LogicOr:
			if len(args) == 1 {
				expr = args[0]
				continue
			}
		}
		break
	}
	if eq, ok := expr.(*Eq); ok {
		if eq.T1.Equal(index) && !rtlUses(eq.T2, index) {
			return eq.T2
		}
		if eq.T2.Equal(index) && !rtlUses(eq.T1, index) {
			return eq.T1
		}
	}
	return nil
}

// rtlPointWrites recovers the point writes from a symbolic functional update.
// The last assignment appears outermost and wins on an address collision.
func rtlPointWrites(expr, index Expr, symbol *Const, defs map[string]*IvyDefinition, guards []rtlGuard, depth int, writes *[]rtlWrite) error {
	if from := rtlReadAt(expr, index); from != nil {
		if from.Name == symbol.Name {
			return nil
		}
		if def := defs[from.Name]; def != nil {
			body, err := rtlInline(def, []Expr{index})
			if err != nil {
				return err
			}
			return rtlPointWrites(body, index, symbol, defs, guards, depth, writes)
		}
	}
	if conditional, ok := expr.(*LogicIte); ok {
		if address := rtlIndexTest(conditional.Cond, index); address != nil {
			data, err := Substitute(conditional.Then, map[NodeKey]Expr{Key(index): address})
			if err != nil {
				return err
			}
			*writes = append(*writes, rtlWrite{guards: guards, address: address, data: data, depth: depth})
			return rtlPointWrites(conditional.Else, index, symbol, defs, guards, depth+1, writes)
		}
		if !rtlUses(conditional.Cond, index) {
			branch := func(invert bool, body Expr) error {
				path := append(slices.Clone(guards), rtlGuard{expr: conditional.Cond, invert: invert})
				return rtlPointWrites(body, index, symbol, defs, path, depth, writes)
			}
			if err := branch(false, conditional.Then); err != nil {
				return err
			}
			return branch(true, conditional.Else)
		}
	}
	return fmt.Errorf("RTL array %s update is not a point write: %s", symbol.Name, expr)
}

func rtlExclusive(a, b []rtlGuard) bool {
	for _, x := range a {
		for _, y := range b {
			if x.invert != y.invert && x.expr.Equal(y.expr) {
				return true
			}
		}
	}
	return false
}

func rtlMatchGuards(a, b []rtlGuard) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		if !slices.ContainsFunc(b, func(y rtlGuard) bool {
			return x.invert == y.invert && x.expr.Equal(y.expr)
		}) {
			return false
		}
	}
	return true
}

func rtlMergeWrite(a, b rtlWrite) (rtlWrite, bool) {
	if !a.address.Equal(b.address) || !a.data.Equal(b.data) {
		return rtlWrite{}, false
	}
	for i, x := range a.guards {
		for j, y := range b.guards {
			if x.invert == y.invert || !x.expr.Equal(y.expr) {
				continue
			}
			xrest := append(slices.Clone(a.guards[:i]), a.guards[i+1:]...)
			yrest := append(slices.Clone(b.guards[:j]), b.guards[j+1:]...)
			if rtlMatchGuards(xrest, yrest) {
				a.guards, a.depth = xrest, max(a.depth, b.depth)
				return a, true
			}
		}
	}
	return rtlWrite{}, false
}

func rtlWritePorts(writes []rtlWrite) [][]rtlWrite {
	for merged := true; merged; {
		merged = false
	search:
		for i, a := range writes {
			for j := i + 1; j < len(writes); j++ {
				if combined, ok := rtlMergeWrite(a, writes[j]); ok {
					writes[i] = combined
					writes = slices.Delete(writes, j, j+1)
					merged = true
					break search
				}
			}
		}
	}
	sort.SliceStable(writes, func(i, j int) bool { return writes[i].depth > writes[j].depth })
	var groups [][]rtlWrite
	for _, write := range writes {
		port := 0
		for i, group := range groups {
			if slices.ContainsFunc(group, func(other rtlWrite) bool { return !rtlExclusive(write.guards, other.guards) }) {
				port = i + 1
			}
		}
		if port == len(groups) {
			groups = append(groups, nil)
		}
		groups[port] = append(groups[port], write)
	}
	return groups
}

func (t *rtlTranslator) declareMemories(m *rtlModule) error {
	for _, symbol := range t.memories(m.name) {
		address, data, err := t.memoryDimensions(symbol)
		if err != nil {
			return err
		}
		mem := &rtlMemory{symbol: symbol, address: address, data: data}
		for _, clock := range t.clocks {
			update := t.updates[m.name].Get(clock)
			if def := update.defs[ActionNewName(symbol.Name)]; def != nil {
				if mem.updateDef != nil {
					return fmt.Errorf("RTL array %s is written on more than one clock", symbol.Name)
				}
				mem.clock, mem.update, mem.updateDef = clock, update, def
			}
		}
		if mem.updateDef != nil {
			def := mem.updateDef
			if args := NodeArgs(def.Lhs); len(args) != 1 {
				return fmt.Errorf("invalid RTL update for %s", symbol.Name)
			} else if err := rtlPointWrites(def.Rhs, args[0], symbol, mem.update.defs, nil, 0, &mem.writes); err != nil {
				if address > rtlBankAddressBits {
					return err
				}
				name := rtlLocal(m.name, symbol.Name)
				for i := 0; i < 1<<address; i++ {
					mem.bank = append(mem.bank, m.wire(rtlPublic(fmt.Sprintf("%s_%d", name, i)), data, ""))
				}
			}
		}
		if mem.bank == nil {
			if address >= strconv.IntSize-1 {
				return fmt.Errorf("RTL memory %s has %d address bits", symbol.Name, address)
			}
			fmt.Fprintf(&m.header, "  memory width %d size %d %s\n", data, 1<<address, rtlPublic(rtlLocal(m.name, symbol.Name)))
		}
		m.memories = append(m.memories, mem)
	}
	return nil
}

func rtlMemoryID(m *rtlModule, symbol *Const) string {
	return strconv.Quote(rtlPublic(rtlLocal(m.name, symbol.Name)))
}

func (e *rtlEmitter) readMemory(symbol *Const, args []Expr) (rtlNet, error) {
	var mem *rtlMemory
	for _, candidate := range e.module.memories {
		if candidate.symbol.Name == symbol.Name {
			mem = candidate
			break
		}
	}
	if mem == nil || len(args) != 1 {
		return rtlNet{}, fmt.Errorf("no local RTL memory for %s", symbol.Name)
	}
	address, err := e.expr(args[0])
	if err != nil {
		return rtlNet{}, err
	}
	m := e.module
	if mem.bank == nil {
		data := m.net(mem.data)
		m.cell("$memrd", m.fresh("$memrd$"+symbol.Name), []rtlParam{{"MEMID", rtlMemoryID(m, symbol)},
			{"ABITS", mem.address}, {"WIDTH", mem.data}, {"CLK_ENABLE", 0}, {"CLK_POLARITY", 0}, {"TRANSPARENT", 0}},
			rtlPort{"CLK", rtlNet{"1'x", 1}}, rtlPort{"EN", rtlNet{"1'x", 1}},
			rtlPort{"ADDR", address}, rtlPort{"DATA", data})
		return data, nil
	}
	if value, ok := rtlLiteral(args[0]); ok {
		return mem.bank[int(value.Uint64())&(len(mem.bank)-1)], nil
	}
	if _, named := m.wires[address.name]; !named {
		net := m.net(address.width)
		m.connect(net, address)
		address = net
	}
	current := slices.Clone(mem.bank)
	for bit := 0; len(current) > 1; bit++ {
		selectBit := rtlNet{fmt.Sprintf("%s [%d]", address.name, bit), 1}
		for i := 0; i < len(current)/2; i++ {
			current[i] = rtlMux(m, selectBit, current[2*i+1], current[2*i])
		}
		current = current[:len(current)/2]
	}
	return current[0], nil
}

func (t *rtlTranslator) emitMemories(m *rtlModule) error {
	for _, memory := range m.memories {
		if err := t.emitMemory(m, memory); err != nil {
			return err
		}
	}
	return nil
}

func (t *rtlTranslator) emitMemory(m *rtlModule, memory *rtlMemory) error {
	init := t.updates[m.name].Get("init")
	def := init.defs[ActionNewName(memory.symbol.Name)]
	var fill Expr
	var points []rtlWrite
	if def != nil {
		args := NodeArgs(def.Lhs)
		if len(args) != 1 {
			return fmt.Errorf("invalid RTL initializer for %s", memory.symbol.Name)
		}
		index := args[0]
		var err error
		fill, err = t.memoryInit(def.Rhs, index, memory.symbol, init.defs, &points)
		if err != nil {
			return err
		}
	}
	if memory.bank != nil {
		return t.emitBank(m, memory, init.defs, fill, points)
	}
	if err := t.emitMemoryInit(m, memory, fill, points); err != nil {
		return err
	}
	e := t.emitter(m, memory.update.defs)
	for port, writes := range rtlWritePorts(memory.writes) {
		if err := e.writeMemoryPort(memory, port, writes); err != nil {
			return err
		}
	}
	return nil
}

func rtlLiteral(expr Expr) (*big.Int, bool) {
	if IsTrue(expr) {
		return big.NewInt(1), true
	}
	if IsFalse(expr) {
		return new(big.Int), true
	}
	if constant, ok := expr.(*Const); ok && IsNumeral(constant) {
		if value, ok := new(big.Int).SetString(constant.Name, 0); ok {
			return value, true
		}
		return new(big.Int).SetString(constant.Name, 10)
	}
	return nil, false
}

func (t *rtlTranslator) memoryInit(expr, index Expr, symbol *Const, defs map[string]*IvyDefinition, points *[]rtlWrite) (Expr, error) {
	if from := rtlReadAt(expr, index); from != nil {
		if def := defs[from.Name]; def != nil {
			body, err := rtlInline(def, []Expr{index})
			if err != nil {
				return nil, err
			}
			return t.memoryInit(body, index, symbol, defs, points)
		}
		if from.Name == symbol.Name {
			return nil, nil
		}
		if def, exists := t.defs.Get2(from.Name); exists {
			body, err := rtlInline(def.def, []Expr{index})
			if err != nil {
				return nil, err
			}
			return t.memoryInit(body, index, symbol, defs, points)
		}
		if slices.ContainsFunc(t.mod.InputWires, func(input *Const) bool { return input.Name == from.Name }) ||
			!t.memoryReferencedOutsideInit(from.Name) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot translate initialization of array %s from undefined %s used elsewhere", symbol.Name, from.Name)
	}
	if conditional, ok := expr.(*LogicIte); ok {
		if address := rtlIndexTest(conditional.Cond, index); address != nil {
			*points = append(*points, rtlWrite{address: address, data: conditional.Then})
			return t.memoryInit(conditional.Else, index, symbol, defs, points)
		}
		return nil, fmt.Errorf("unsupported conditional initializer for RTL array %s", symbol.Name)
	}
	if rtlUses(expr, index) {
		return nil, fmt.Errorf("unsupported indexed initializer for RTL array %s", symbol.Name)
	}
	return expr, nil
}

func (t *rtlTranslator) memoryReferencedOutsideInit(name string) bool {
	mentions := func(expr Expr) bool {
		for _, node := range UsedSymbolsAST(expr).All() {
			if ExprName(node) == name {
				return true
			}
		}
		return false
	}
	for _, definition := range t.defs.All() {
		if mentions(definition.def.Rhs) {
			return true
		}
	}
	for _, updates := range t.updates {
		for _, clock := range t.clocks {
			update, ok := updates.Get2(clock)
			if ok {
				for _, def := range update.clauses.Defs {
					if mentions(def.Rhs) {
						return true
					}
				}
				for _, formula := range update.clauses.Fmlas {
					if mentions(formula) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (t *rtlTranslator) emitBank(m *rtlModule, memory *rtlMemory, init map[string]*IvyDefinition, fill Expr, points []rtlWrite) error {
	e := t.emitter(m, memory.update.defs)
	initializer := t.emitter(m, init)
	for i, q := range memory.bank {
		reset := rtlInt(0, memory.data)
		value := fill
		for _, point := range points {
			address, ok := rtlLiteral(point.address)
			if !ok {
				return fmt.Errorf("nonconstant initializer address for %s", memory.symbol.Name)
			}
			if int(address.Uint64())&(len(memory.bank)-1) == i {
				value = point.data
				break
			}
		}
		if value != nil {
			r, err := initializer.expr(value)
			if err != nil {
				return err
			}
			reset = r
		}
		index := NewConst(strconv.Itoa(i), memory.symbol.CSort.(*LogicFunctionSort).Domain()[0])
		data, err := e.inline(memory.updateDef, []Expr{index})
		if err != nil {
			return err
		}
		d := rtlMux(m, m.wires[rtlPublic(rtlReset)], reset, data)
		m.cell("$dff", m.fresh("$dff$"+memory.symbol.Name), []rtlParam{{"WIDTH", q.width}, {"CLK_POLARITY", 1}},
			rtlPort{"CLK", m.wires[rtlPublic(memory.clock)]}, rtlPort{"D", d}, rtlPort{"Q", q})
	}
	return nil
}

func (t *rtlTranslator) emitMemoryInit(m *rtlModule, memory *rtlMemory, fill Expr, points []rtlWrite) error {
	priority := 0
	emit := func(address, data rtlNet, words int) {
		enable := rtlNet{fmt.Sprintf("%d'%s", memory.data, strings.Repeat("1", memory.data)), memory.data}
		m.cell("$meminit_v2", m.fresh("$meminit$"+memory.symbol.Name),
			[]rtlParam{{"MEMID", rtlMemoryID(m, memory.symbol)}, {"ABITS", memory.address},
				{"WIDTH", memory.data}, {"WORDS", words}, {"PRIORITY", priority}},
			rtlPort{"ADDR", address}, rtlPort{"DATA", data}, rtlPort{"EN", enable})
		priority++
	}
	if fill != nil {
		value, ok := rtlLiteral(fill)
		if !ok {
			return fmt.Errorf("nonconstant initializer for RTL array %s", memory.symbol.Name)
		}
		words := 1 << memory.address
		binary := strings.SplitN(rtlBits(value, memory.data).name, "'", 2)[1]
		emit(rtlInt(0, memory.address), rtlNet{fmt.Sprintf("%d'%s", words*memory.data, strings.Repeat(binary, words)), words * memory.data}, words)
	}
	for i := len(points) - 1; i >= 0; i-- {
		address, addressOK := rtlLiteral(points[i].address)
		data, dataOK := rtlLiteral(points[i].data)
		if !addressOK || !dataOK {
			return fmt.Errorf("nonconstant initializer for RTL array %s", memory.symbol.Name)
		}
		emit(rtlBits(address, memory.address), rtlBits(data, memory.data), 1)
	}
	return nil
}

func (e *rtlEmitter) writeMemoryPort(memory *rtlMemory, port int, writes []rtlWrite) error {
	m := e.module
	var enable, address, data rtlNet
	for i, write := range writes {
		condition := rtlInt(1, 1)
		for j, guard := range write.guards {
			current, err := e.expr(guard.expr)
			if err != nil {
				return err
			}
			if guard.invert {
				current = rtlNot(m, current)
			}
			if j == 0 {
				condition = current
			} else {
				condition = rtlBinary(m, "$and", condition, current, 1)
			}
		}
		addr, err := e.expr(write.address)
		if err != nil {
			return err
		}
		value, err := e.expr(write.data)
		if err != nil {
			return err
		}
		if i == 0 {
			enable, address, data = condition, addr, value
		} else {
			enable = rtlBinary(m, "$or", enable, condition, 1)
			address = rtlMux(m, condition, addr, address)
			data = rtlMux(m, condition, value, data)
		}
	}
	if memory.data != 1 {
		enable = rtlNet{"{ " + strings.Repeat(enable.name+" ", memory.data) + "}", memory.data}
	}
	mask := "0'0"
	if port > 0 {
		mask = fmt.Sprintf("%d'%s", port, strings.Repeat("1", port))
	}
	m.cell("$memwr_v2", m.fresh("$memwr$"+memory.symbol.Name),
		[]rtlParam{{"MEMID", rtlMemoryID(m, memory.symbol)}, {"ABITS", memory.address},
			{"WIDTH", memory.data}, {"PORTID", port}, {"PRIORITY_MASK", mask}, {"CLK_ENABLE", "1'1"}, {"CLK_POLARITY", "1'1"}},
		rtlPort{"ADDR", address}, rtlPort{"DATA", data}, rtlPort{"EN", enable}, rtlPort{"CLK", m.wires[rtlPublic(memory.clock)]})
	return nil
}
