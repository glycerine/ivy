package goivy

import (
	"fmt"
	"math/big"
	"strings"
)

type rtlEmitter struct {
	translator *rtlTranslator
	module     *rtlModule
	object     string
	local      map[string]*IvyDefinition
	cache      map[NodeKey]rtlNet
	active     map[NodeKey]bool
}

func (t *rtlTranslator) emitter(module *rtlModule, defs map[string]*IvyDefinition) *rtlEmitter {
	return &rtlEmitter{translator: t, module: module, object: module.name, local: defs,
		cache: make(map[NodeKey]rtlNet), active: make(map[NodeKey]bool)}
}

func rtlBase(name string) string {
	for {
		original := name
		name = strings.TrimPrefix(strings.TrimPrefix(name, "new_"), "__m_")
		if original == name {
			return name
		}
	}
}

func rtlMux(m *rtlModule, selectNet, whenTrue, whenFalse rtlNet) rtlNet {
	result := m.net(whenTrue.width)
	m.cell("$mux", m.fresh("$mux"), []rtlParam{{"WIDTH", result.width}},
		rtlPort{"A", whenFalse}, rtlPort{"B", whenTrue}, rtlPort{"S", selectNet}, rtlPort{"Y", result})
	return result
}

func rtlBinary(m *rtlModule, operator string, a, b rtlNet, width int) rtlNet {
	y := m.net(width)
	m.cell(operator, m.fresh(operator), []rtlParam{{"A_SIGNED", 0}, {"B_SIGNED", 0},
		{"A_WIDTH", a.width}, {"B_WIDTH", b.width}, {"Y_WIDTH", width}},
		rtlPort{"A", a}, rtlPort{"B", b}, rtlPort{"Y", y})
	return y
}

func rtlNot(m *rtlModule, a rtlNet) rtlNet {
	y := m.net(a.width)
	m.cell("$not", m.fresh("$not"), []rtlParam{{"A_SIGNED", 0}, {"A_WIDTH", a.width}, {"Y_WIDTH", y.width}},
		rtlPort{"A", a}, rtlPort{"Y", y})
	return y
}

func (e *rtlEmitter) expr(expr Expr) (rtlNet, error) {
	key := Key(expr)
	if net, ok := e.cache[key]; ok {
		return net, nil
	}
	if e.active[key] {
		return rtlNet{}, fmt.Errorf("cyclic RTL expression: %s", expr)
	}
	e.active[key] = true
	net, err := e.emitExpr(expr)
	delete(e.active, key)
	if err == nil {
		e.cache[key] = net
	}
	return net, err
}

func (e *rtlEmitter) emitExpr(expr Expr) (rtlNet, error) {
	switch term := expr.(type) {
	case *Const:
		if IsNumeral(term) {
			return e.constant(term)
		}
		if def := e.local[term.Name]; def != nil {
			return e.expr(def.Rhs)
		}
		if wire, ok := e.module.wires[rtlPublic(rtlLocal(e.object, term.Name))]; ok {
			return wire, nil
		}
		if wire, ok := e.module.wires[rtlPublic(rtlLocal(e.object, rtlBase(term.Name)))]; ok {
			return wire, nil
		}
		return rtlNet{}, fmt.Errorf("undeclared hardware signal %s in %s", term.Name, e.object)
	case *LogicIte:
		test, err := e.expr(term.Cond)
		if err != nil {
			return rtlNet{}, err
		}
		yes, err := e.expr(term.Then)
		if err != nil {
			return rtlNet{}, err
		}
		no, err := e.expr(term.Else)
		if err != nil {
			return rtlNet{}, err
		}
		return rtlMux(e.module, test, yes, no), nil
	case *Eq:
		return e.binary("$eq", term.T1, term.T2, 1)
	case *LogicIff:
		return e.binary("$eq", term.T1, term.T2, 1)
	case *LogicNot:
		net, err := e.expr(term.Body)
		if err != nil {
			return rtlNet{}, err
		}
		return rtlNot(e.module, net), nil
	case *LogicAnd:
		return e.reduce("$and", term.Terms, 1)
	case *LogicOr:
		return e.reduce("$or", term.Terms, 0)
	case *LogicImplies:
		return e.expr(&LogicOr{Terms: []Expr{&LogicNot{Body: term.T1}, term.T2}})
	case *Apply:
		return e.application(term)
	default:
		return rtlNet{}, fmt.Errorf("cannot translate RTL expression: %s", term)
	}
}

func (e *rtlEmitter) constant(term *Const) (rtlNet, error) {
	width, err := e.translator.width(term.CSort)
	if err != nil {
		return rtlNet{}, err
	}
	value, ok := new(big.Int).SetString(term.Name, 0)
	if !ok {
		value, ok = new(big.Int).SetString(term.Name, 10)
	}
	if !ok {
		return rtlNet{}, fmt.Errorf("invalid RTL numeral: %s", term.Name)
	}
	return rtlBits(value, width), nil
}

func (e *rtlEmitter) binary(kind string, lhs, rhs Expr, width int) (rtlNet, error) {
	a, err := e.expr(lhs)
	if err != nil {
		return rtlNet{}, err
	}
	b, err := e.expr(rhs)
	if err != nil {
		return rtlNet{}, err
	}
	return rtlBinary(e.module, kind, a, b, width), nil
}

func (e *rtlEmitter) reduce(kind string, args []Expr, identity int64) (rtlNet, error) {
	if len(args) == 0 {
		return rtlInt(identity, 1), nil
	}
	value, err := e.expr(args[0])
	if err != nil {
		return rtlNet{}, err
	}
	for _, arg := range args[1:] {
		next, err := e.expr(arg)
		if err != nil {
			return rtlNet{}, err
		}
		value = rtlBinary(e.module, kind, value, next, 1)
	}
	return value, nil
}

func rtlInline(def *IvyDefinition, actuals []Expr) (Expr, error) {
	formals := NodeArgs(def.Lhs)
	if len(formals) != len(actuals) {
		return nil, fmt.Errorf("invalid hardware function arguments: %s", def.Lhs)
	}
	substitutions := make(map[NodeKey]Expr, len(formals))
	for i, formal := range formals {
		substitutions[Key(formal)] = actuals[i]
	}
	return Substitute(def.Rhs, substitutions)
}

var rtlArithmetic = map[string]string{"+": "$add", "-": "$sub", "*": "$mul", "/": "$div"}
var rtlComparison = map[string]string{"<": "$lt", "<=": "$le", ">": "$gt", ">=": "$ge"}

func (e *rtlEmitter) application(term *Apply) (rtlNet, error) {
	sym := GetAppRep(term)
	if sym == nil {
		return rtlNet{}, fmt.Errorf("unsupported RTL application: %s", term)
	}
	name := sym.Name
	if def, ok := e.translator.defs.Get2(name); ok {
		if e.translator.wires.Get(name) == nil {
			return rtlNet{}, fmt.Errorf("cannot translate non-wire function %s to RTL", name)
		}
		return e.inline(def.def, term.Terms)
	}
	if def := e.local[name]; def != nil {
		return e.inline(def, term.Terms)
	}
	if e.translator.memorySymbol(sym) {
		return e.readMemory(sym, term.Terms)
	}
	if cell := rtlArithmetic[name]; cell != "" && len(term.Terms) == 2 {
		width, err := e.translator.width(term.NodeSort())
		if err != nil {
			return rtlNet{}, err
		}
		return e.binary(cell, term.Terms[0], term.Terms[1], width)
	}
	if cell := rtlComparison[name]; cell != "" && len(term.Terms) == 2 {
		return e.binary(cell, term.Terms[0], term.Terms[1], 1)
	}
	if name == "concat" {
		var pieces []string
		width := 0
		for _, arg := range term.Terms {
			part, err := e.expr(arg)
			if err != nil {
				return rtlNet{}, err
			}
			pieces = append(pieces, part.name)
			width += part.width
		}
		return rtlNet{name: "{ " + strings.Join(pieces, " ") + " }", width: width}, nil
	}
	if strings.HasPrefix(name, "bfe[") && len(term.Terms) == 1 {
		low, high, ok := parseBfeParams(name)
		if !ok {
			return rtlNet{}, fmt.Errorf("invalid RTL bit-field extraction %s", name)
		}
		base, err := e.expr(term.Terms[0])
		if err != nil {
			return rtlNet{}, err
		}
		if low < 0 || high < low || high >= base.width {
			return rtlNet{}, fmt.Errorf("out-of-range RTL bit-field extraction %s of %d bits", name, base.width)
		}
		if _, named := e.module.wires[base.name]; !named {
			temp := e.module.net(base.width)
			e.module.connect(temp, base)
			base = temp
		}
		extract := fmt.Sprintf("%s [%d:%d]", base.name, high, low)
		if high == low {
			extract = fmt.Sprintf("%s [%d]", base.name, low)
		}
		return rtlNet{name: extract, width: high - low + 1}, nil
	}
	return rtlNet{}, fmt.Errorf("unsupported RTL operator: %s", name)
}

func (e *rtlEmitter) inline(def *IvyDefinition, args []Expr) (rtlNet, error) {
	inlined, err := rtlInline(def, args)
	if err != nil {
		return rtlNet{}, err
	}
	return e.expr(inlined)
}
