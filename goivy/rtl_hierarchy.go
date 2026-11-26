package goivy

import (
	"fmt"
	"strings"
)

func (t *rtlTranslator) wire(m *rtlModule, symbol *Const, direction string) (rtlNet, error) {
	width, err := t.width(symbol.CSort)
	if err != nil {
		return rtlNet{}, fmt.Errorf("signal %s: %w", symbol.Name, err)
	}
	return m.wire(rtlPublic(rtlLocal(m.name, symbol.Name)), width, direction), nil
}

func (t *rtlTranslator) emitObject(out *strings.Builder, object string) error {
	m := &rtlModule{name: object, wires: make(map[string]rtlNet)}
	inputs, outputs, internal := t.classify(object)
	for _, group := range []struct {
		direction string
		symbols   []*Const
	}{{"input", inputs}, {"input", t.foreignInputs(object)}, {"output", outputs}, {"", internal}} {
		for _, symbol := range group.symbols {
			if _, err := t.wire(m, symbol, group.direction); err != nil {
				return err
			}
		}
	}
	if t.clocked[object] {
		for _, clk := range t.clocks {
			m.wire(rtlPublic(clk), 1, "input")
		}
		m.wire(rtlPublic(rtlReset), 1, "input")
	}
	for _, symbol := range t.state[object].All() {
		if !IsFunctionSort(symbol.CSort) {
			direction := ""
			if t.exported[symbol.Name] {
				direction = "output"
			}
			if _, err := t.wire(m, symbol, direction); err != nil {
				return err
			}
		}
	}

	children := t.children(object)
	for _, child := range children {
		_, childOutputs, _ := t.classify(child)
		for _, symbol := range childOutputs {
			if _, err := t.wire(m, symbol, ""); err != nil {
				return err
			}
		}
		for _, symbol := range t.state[child].All() {
			if t.exported[symbol.Name] {
				if _, err := t.wire(m, symbol, ""); err != nil {
					return err
				}
			}
		}
	}

	if err := t.declareMemories(m); err != nil {
		return err
	}
	e := t.emitter(m, nil)
	for _, child := range children {
		if err := t.emitChild(e, child); err != nil {
			return err
		}
	}
	for _, group := range [][]*Const{outputs, internal} {
		for _, symbol := range group {
			if def, ok := t.defs.Get2(symbol.Name); ok && def.owner == object {
				value, err := e.expr(def.def.Rhs)
				if err != nil {
					return fmt.Errorf("wire %s: %w", symbol.Name, err)
				}
				m.connect(m.wires[rtlPublic(rtlLocal(object, symbol.Name))], value)
			}
		}
	}
	for _, symbol := range t.state[object].All() {
		if !IsFunctionSort(symbol.CSort) {
			if err := t.emitRegister(m, symbol); err != nil {
				return err
			}
		}
	}
	if err := t.emitMemories(m); err != nil {
		return err
	}
	m.write(out)
	for _, child := range children {
		if err := t.emitObject(out, child); err != nil {
			return err
		}
	}
	return nil
}

func (t *rtlTranslator) emitChild(e *rtlEmitter, child string) error {
	m := e.module
	inputs, outputs, _ := t.classify(child)
	var ports []rtlPort
	for _, input := range inputs {
		var signal rtlNet
		if def, ok := t.defs.Get2(input.Name); ok && def.owner == m.name {
			value, err := e.expr(def.def.Rhs)
			if err != nil {
				return fmt.Errorf("wire %s: %w", input.Name, err)
			}
			signal = value
		} else {
			width, err := t.width(input.CSort)
			if err != nil {
				return err
			}
			signal = m.net(width)
		}
		ports = append(ports, rtlPort{rtlLocal(child, input.Name), signal})
	}
	for _, symbol := range append(outputs, t.foreignInputs(child)...) {
		signal, err := t.wire(m, symbol, "")
		if err != nil {
			return err
		}
		ports = append(ports, rtlPort{rtlLocal(child, symbol.Name), signal})
	}
	for _, symbol := range t.state[child].All() {
		if t.exported[symbol.Name] {
			signal := m.wires[rtlPublic(rtlLocal(m.name, symbol.Name))]
			ports = append(ports, rtlPort{rtlLocal(child, symbol.Name), signal})
		}
	}
	if t.clocked[child] {
		for _, clk := range t.clocks {
			ports = append(ports, rtlPort{clk, m.wires[rtlPublic(clk)]})
		}
		ports = append(ports, rtlPort{rtlReset, m.wires[rtlPublic(rtlReset)]})
	}
	m.cell(rtlPublic(child), rtlPublic(rtlLocal(m.name, child)), nil, ports...)
	return nil
}

func (t *rtlTranslator) emitRegister(m *rtlModule, symbol *Const) error {
	if len(t.clocks) == 0 {
		return fmt.Errorf("register %s has no clock", symbol.Name)
	}
	key := ActionNewName(symbol.Name)
	q := m.wires[rtlPublic(rtlLocal(m.name, symbol.Name))]
	data, reset, clock := q, rtlInt(0, q.width), t.clocks[0]
	foundClock := false
	for _, clk := range t.clocks {
		update := t.updates[m.name].Get(clk)
		if def := update.defs[key]; def != nil {
			if foundClock {
				return fmt.Errorf("register %s is updated by both %s and %s", symbol.Name, clock, clk)
			}
			value, err := t.emitter(m, update.defs).expr(def.Rhs)
			if err != nil {
				return fmt.Errorf("clock update of %s: %w", symbol.Name, err)
			}
			data, clock, foundClock = value, clk, true
		}
	}
	init := t.updates[m.name].Get("init")
	if def := init.defs[key]; def != nil {
		value, err := t.emitter(m, init.defs).expr(def.Rhs)
		if err != nil {
			return fmt.Errorf("initializer of %s: %w", symbol.Name, err)
		}
		reset = value
	}
	d := rtlMux(m, m.wires[rtlPublic(rtlReset)], reset, data)
	m.cell("$dff", "$dff$"+symbol.Name, []rtlParam{{"WIDTH", q.width}, {"CLK_POLARITY", 1}},
		rtlPort{"CLK", m.wires[rtlPublic(clock)]}, rtlPort{"D", d}, rtlPort{"Q", q})
	return nil
}
