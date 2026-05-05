// Package evparser implements the event trace parser for counterexample display.
// This corresponds to Python's ivy_ev_parser.py.
//
// The grammar (from the PLY parser):
//
//	events : /* empty */ | events event
//	event  : optdir SYMBOL optargs optsubs
//	optdir : /* empty */ | '>' | '<'
//	optsubs: /* empty */ | ';' | '{' events '}'
//	optargs: /* empty */ | '(' list ')'
//	value  : SYMBOL | SYMBOL '(' list ')' | '[' ']' | '[' list ']' | '{' '}' | '{' dict '}'
//	list   : value | list ',' value
//	dict   : SYMBOL ':' value | dict ',' SYMBOL ':' value
//
// SYMBOL matches: [_a-zA-Z0-9.\$\-]+  |  *  |  "..."
package evparser

import (
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------
// AST types
// -----------------------------------------------------------------------

// Value is the interface for all event trace values.
type Value interface {
	Text() string
	String() string
	Subs() []Value
	Match(pattern Value, binding map[string]Value) bool
	Map(fn func(string) Value) Value
}

// Events is a list of Event values (top-level parse result).
type Events []*Event

func (es Events) String() string {
	parts := make([]string, len(es))
	for i, e := range es {
		parts[i] = e.String()
	}
	return strings.Join(parts, " ")
}

// EventDir indicates the direction of an event.
type EventDir int

const (
	DirNone EventDir = iota
	DirIn            // >
	DirOut           // <
)

// Event represents a single event with optional direction, arguments, and children.
type Event struct {
	Dir      EventDir
	Rep      string
	Args     []Value
	Children Events
}

func (e *Event) Text() string {
	res := e.Rep
	if len(e.Args) > 0 {
		parts := make([]string, len(e.Args))
		for i, a := range e.Args {
			parts[i] = a.String()
		}
		res += "(" + strings.Join(parts, ",") + ")"
	}
	return res
}

func (e *Event) Subs() []Value {
	var result []Value
	if len(e.Args) > 0 {
		result = append(result, &ArgList{Args: e.Args})
	}
	for _, c := range e.Children {
		result = append(result, c)
	}
	return result
}

func (e *Event) Match(ev Value, binding map[string]Value) bool {
	other, ok := ev.(*Event)
	if !ok {
		return false
	}
	if other.Rep != "*" && !strings.HasPrefix(e.Rep, other.Rep) {
		return false
	}
	for i := 0; i < len(e.Args) && i < len(other.Args); i++ {
		if !e.Args[i].Match(other.Args[i], binding) {
			return false
		}
	}
	return true
}

func (e *Event) String() string {
	var prefix string
	switch e.Dir {
	case DirIn:
		prefix = "> "
	case DirOut:
		prefix = "< "
	}
	res := prefix + e.Text()
	if len(e.Children) > 0 {
		res += "{" + e.Children.String() + "}"
	}
	return res
}

func (e *Event) Map(fn func(string) Value) Value {
	newArgs := make([]Value, len(e.Args))
	for i, a := range e.Args {
		newArgs[i] = a.Map(fn)
	}
	return &Event{Dir: e.Dir, Rep: e.Rep, Args: newArgs, Children: e.Children}
}

// ArgList wraps event arguments for tree traversal.
type ArgList struct {
	Args []Value
}

func (a *ArgList) Text() string                                 { return "args" }
func (a *ArgList) String() string                               { return a.Text() }
func (a *ArgList) Subs() []Value                                { return a.Args }
func (a *ArgList) Match(v Value, binding map[string]Value) bool { return false }
func (a *ArgList) Map(fn func(string) Value) Value              { return a }

// Symbol represents a simple name value.
type Symbol struct {
	Name string
}

func (s *Symbol) Text() string   { return s.Name }
func (s *Symbol) String() string { return s.Name }
func (s *Symbol) Subs() []Value  { return nil }

func (s *Symbol) Match(v Value, binding map[string]Value) bool {
	if sym, ok := v.(*Symbol); ok {
		if sym.Name == "*" || sym.Name == s.Name {
			return true
		}
		if strings.HasPrefix(sym.Name, "$") {
			if binding != nil {
				if existing, found := binding[sym.Name]; found {
					if es, ok2 := existing.(*Symbol); ok2 {
						return es.Name == s.Name
					}
					return false
				}
				binding[sym.Name] = s
			}
			return true
		}
	}
	return false
}

func (s *Symbol) Map(fn func(string) Value) Value {
	return fn(s.Name)
}

// App represents a function application value: name(args...).
type App struct {
	Rep  string
	Args []Value
}

func (a *App) Text() string {
	res := a.Rep
	if len(a.Args) > 0 {
		parts := make([]string, len(a.Args))
		for i, arg := range a.Args {
			parts[i] = arg.String()
		}
		res += "(" + strings.Join(parts, ",") + ")"
	}
	return res
}

func (a *App) String() string { return a.Text() }
func (a *App) Subs() []Value  { return a.Args }

func (a *App) Match(v Value, binding map[string]Value) bool {
	if sym, ok := v.(*Symbol); ok && sym.Name == "*" {
		return true
	}
	if other, ok := v.(*App); ok {
		if other.Rep != "*" && other.Rep != a.Rep {
			return false
		}
		for i := 0; i < len(a.Args) && i < len(other.Args); i++ {
			if !a.Args[i].Match(other.Args[i], binding) {
				return false
			}
		}
		return true
	}
	return false
}

func (a *App) Map(fn func(string) Value) Value {
	newArgs := make([]Value, len(a.Args))
	for i, arg := range a.Args {
		newArgs[i] = arg.Map(fn)
	}
	return &App{Rep: a.Rep, Args: newArgs}
}

// ListValue represents a list [...].
type ListValue struct {
	Items []Value
}

func (l *ListValue) Text() string  { return l.String() }
func (l *ListValue) Subs() []Value { return l.Items }

func (l *ListValue) String() string {
	parts := make([]string, len(l.Items))
	for i, item := range l.Items {
		parts[i] = item.String()
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (l *ListValue) Match(v Value, binding map[string]Value) bool {
	if sym, ok := v.(*Symbol); ok && sym.Name == "*" {
		return true
	}
	if other, ok := v.(*ListValue); ok {
		if len(other.Items) != len(l.Items) {
			return false
		}
		for i := range l.Items {
			if !l.Items[i].Match(other.Items[i], binding) {
				return false
			}
		}
		return true
	}
	return false
}

func (l *ListValue) Map(fn func(string) Value) Value {
	newItems := make([]Value, len(l.Items))
	for i, item := range l.Items {
		newItems[i] = item.Map(fn)
	}
	return &ListValue{Items: newItems}
}

// DictEntry represents a key:value pair in a dict for tree traversal.
type DictEntry struct {
	Key   string
	Value Value
}

func (d *DictEntry) Text() string                                 { return d.Key + ":" + d.Value.String() }
func (d *DictEntry) String() string                               { return d.Text() }
func (d *DictEntry) Subs() []Value                                { return d.Value.Subs() }
func (d *DictEntry) Match(v Value, binding map[string]Value) bool { return false }
func (d *DictEntry) Map(fn func(string) Value) Value              { return d }

// DictValue represents a dictionary {key:value, ...}.
type DictValue struct {
	Entries map[string]Value
	Order   []string // insertion order
}

func NewDictValue() *DictValue {
	return &DictValue{Entries: make(map[string]Value)}
}

func (d *DictValue) Text() string { return d.String() }
func (d *DictValue) Subs() []Value {
	result := make([]Value, len(d.Order))
	for i, k := range d.Order {
		result[i] = &DictEntry{Key: k, Value: d.Entries[k]}
	}
	return result
}

func (d *DictValue) String() string {
	parts := make([]string, len(d.Order))
	for i, k := range d.Order {
		parts[i] = k + ":" + d.Entries[k].String()
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (d *DictValue) Match(v Value, binding map[string]Value) bool {
	if sym, ok := v.(*Symbol); ok && sym.Name == "*" {
		return true
	}
	if other, ok := v.(*DictValue); ok {
		for _, k := range other.Order {
			myV, exists := d.Entries[k]
			if !exists {
				return false
			}
			if !myV.Match(other.Entries[k], binding) {
				return false
			}
		}
		return true
	}
	return false
}

func (d *DictValue) Map(fn func(string) Value) Value {
	nd := NewDictValue()
	for _, k := range d.Order {
		nd.Entries[k] = d.Entries[k].Map(fn)
		nd.Order = append(nd.Order, k)
	}
	return nd
}

func (d *DictValue) Set(key string, val Value) {
	if _, exists := d.Entries[key]; !exists {
		d.Order = append(d.Order, key)
	}
	d.Entries[key] = val
}

func (d *DictValue) Get(key string) (Value, bool) {
	v, ok := d.Entries[key]
	return v, ok
}

// -----------------------------------------------------------------------
// Event generators for traversal
// -----------------------------------------------------------------------

// EventGen yields all events depth-first.
func EventGen(evs Events) []*Event {
	var result []*Event
	for _, ev := range evs {
		result = append(result, ev)
		result = append(result, EventGen(ev.Children)...)
	}
	return result
}

// AddrEvent is an (address, event) pair for addressed traversal.
type AddrEvent struct {
	Addr  string
	Event *Event
}

// EventRevGen yields (addr, event) pairs in reverse order starting from addr.
func EventRevGen(evs Events, addr string) []AddrEvent {
	return eventRevGenRec(evs, addr, 0)
}

func eventRevGenRec(things Events, addr string, start int) []AddrEvent {
	var result []AddrEvent
	num := len(things)
	if addr != "" {
		parts := strings.SplitN(addr, "/", 2)
		n, _ := strconv.Atoi(parts[0])
		n -= start
		if n >= 0 && n < len(things) {
			thing := things[n]
			if len(parts) == 2 {
				for _, ae := range eventRevGenRec(thing.Children, parts[1], 1) {
					result = append(result, AddrEvent{parts[0] + "/" + ae.Addr, ae.Event})
				}
				result = append(result, AddrEvent{parts[0], thing})
			}
		}
		num = n
	}
	for idx := num - 1; idx >= 0; idx-- {
		thing := things[idx]
		for _, ae := range eventRevGenRec(thing.Children, "", 1) {
			result = append(result, AddrEvent{strconv.Itoa(idx+start) + "/" + ae.Addr, ae.Event})
		}
		result = append(result, AddrEvent{strconv.Itoa(idx + start), thing})
	}
	return result
}

// EventFwdGen yields (addr, event) pairs in forward order starting from addr.
func EventFwdGen(evs Events, addr string) []AddrEvent {
	return eventFwdGenRec(evs, addr, 0)
}

func eventFwdGenRec(things Events, addr string, start int) []AddrEvent {
	var result []AddrEvent
	num := -1
	if addr != "" {
		parts := strings.SplitN(addr, "/", 2)
		n, _ := strconv.Atoi(parts[0])
		n -= start
		num = n
		if n >= 0 && n < len(things) {
			thing := things[n]
			caddr := ""
			if len(parts) == 2 {
				caddr = parts[1]
			}
			for _, ae := range eventFwdGenRec(thing.Children, caddr, 1) {
				result = append(result, AddrEvent{parts[0] + "/" + ae.Addr, ae.Event})
			}
		}
	}
	for idx := num + 1; idx < len(things); idx++ {
		thing := things[idx]
		result = append(result, AddrEvent{strconv.Itoa(idx + start), thing})
		for _, ae := range eventFwdGenRec(thing.Children, "", 1) {
			result = append(result, AddrEvent{strconv.Itoa(idx+start) + "/" + ae.Addr, ae.Event})
		}
	}
	return result
}

// Anchor creates a mapping function that resolves $N.field references.
func Anchor(anchor *Event) func(string) Value {
	return func(s string) Value {
		if strings.HasPrefix(s, "$") {
			path := strings.Split(s[1:], ".")
			num, err := strconv.Atoi(path[0])
			if err != nil {
				return &Symbol{Name: s}
			}
			num-- // 1-based
			if num < 0 || num >= len(anchor.Args) {
				return &Symbol{Name: s} // error case
			}
			var res Value = anchor.Args[num]
			for _, field := range path[1:] {
				if dv, ok := res.(*DictValue); ok {
					if v, found := dv.Get(field); found {
						res = v
						continue
					}
				}
				return &Symbol{Name: s} // error: no such field
			}
			return res
		}
		return &Symbol{Name: s}
	}
}

// Filter yields events matching any pattern.
func Filter(evs []*Event, pats []*Event, anchor *Event) []*Event {
	if anchor != nil {
		fn := Anchor(anchor)
		mapped := make([]*Event, len(pats))
		for i, p := range pats {
			mapped[i] = p.Map(fn).(*Event)
		}
		pats = mapped
	}
	var result []*Event
	for _, e := range evs {
		for _, pat := range pats {
			if e.Match(pat, nil) {
				result = append(result, e)
				break
			}
		}
	}
	return result
}

// Find finds the first (addr, event) pair matching any pattern.
func Find(evs []AddrEvent, pats []*Event, anchor *Event) *AddrEvent {
	if anchor != nil {
		fn := Anchor(anchor)
		mapped := make([]*Event, len(pats))
		for i, p := range pats {
			mapped[i] = p.Map(fn).(*Event)
		}
		pats = mapped
	}
	for _, ae := range evs {
		for _, pat := range pats {
			if ae.Event.Match(pat, nil) {
				return &ae
			}
		}
	}
	return nil
}

// Binding is a pair of (event, variable binding).
type Binding struct {
	Event   *Event
	Binding map[string]Value
}

// Bind yields events matching patterns with free variable bindings.
func Bind(evs []*Event, pats []*Event, anchor *Event) []Binding {
	if anchor != nil {
		fn := Anchor(anchor)
		mapped := make([]*Event, len(pats))
		for i, p := range pats {
			mapped[i] = p.Map(fn).(*Event)
		}
		pats = mapped
	}
	var result []Binding
	for _, e := range evs {
		binding := make(map[string]Value)
		for _, pat := range pats {
			if e.Match(pat, binding) {
				result = append(result, Binding{Event: e, Binding: binding})
				break
			}
		}
	}
	return result
}

// -----------------------------------------------------------------------
// Lexer
// -----------------------------------------------------------------------

type tokenType int

const (
	tokEOF tokenType = iota
	tokSYMBOL
	tokCOMMA
	tokLPAREN
	tokRPAREN
	tokLBR
	tokRBR
	tokLCB
	tokRCB
	tokSEMI
	tokCOLON
	tokGT
	tokLT
)

type token struct {
	typ tokenType
	val string
}

type lexerState struct {
	input  string
	pos    int
	tokens []token
	cur    int
}

func newLexer(input string) *lexerState {
	l := &lexerState{input: input}
	l.tokenize()
	return l
}

func isSymbolChar(c byte) bool {
	return c == '_' || c == '.' || c == '$' || c == '-' || c == '*' ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func (l *lexerState) tokenize() {
	for l.pos < len(l.input) {
		c := l.input[l.pos]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			l.pos++
		case c == ',':
			l.tokens = append(l.tokens, token{tokCOMMA, ","})
			l.pos++
		case c == ';':
			l.tokens = append(l.tokens, token{tokSEMI, ";"})
			l.pos++
		case c == ':':
			l.tokens = append(l.tokens, token{tokCOLON, ":"})
			l.pos++
		case c == '(':
			l.tokens = append(l.tokens, token{tokLPAREN, "("})
			l.pos++
		case c == ')':
			l.tokens = append(l.tokens, token{tokRPAREN, ")"})
			l.pos++
		case c == '[':
			l.tokens = append(l.tokens, token{tokLBR, "["})
			l.pos++
		case c == ']':
			l.tokens = append(l.tokens, token{tokRBR, "]"})
			l.pos++
		case c == '{':
			l.tokens = append(l.tokens, token{tokLCB, "{"})
			l.pos++
		case c == '}':
			l.tokens = append(l.tokens, token{tokRCB, "}"})
			l.pos++
		case c == '>':
			l.tokens = append(l.tokens, token{tokGT, ">"})
			l.pos++
		case c == '<':
			l.tokens = append(l.tokens, token{tokLT, "<"})
			l.pos++
		case c == '"':
			// Quoted string
			end := l.pos + 1
			for end < len(l.input) && l.input[end] != '"' {
				end++
			}
			if end < len(l.input) {
				end++ // include closing quote
			}
			l.tokens = append(l.tokens, token{tokSYMBOL, l.input[l.pos:end]})
			l.pos = end
		case isSymbolChar(c):
			start := l.pos
			for l.pos < len(l.input) && isSymbolChar(l.input[l.pos]) {
				l.pos++
			}
			l.tokens = append(l.tokens, token{tokSYMBOL, l.input[start:l.pos]})
		default:
			// skip illegal characters
			l.pos++
		}
	}
}

func (l *lexerState) peek() token {
	if l.cur >= len(l.tokens) {
		return token{tokEOF, ""}
	}
	return l.tokens[l.cur]
}

func (l *lexerState) next() token {
	t := l.peek()
	if t.typ != tokEOF {
		l.cur++
	}
	return t
}

// The parser is generated by goyacc from ev_grammar.y.
// Run: go generate
// Or: goyacc -o ev_parser.go -p ev ev_grammar.y

//go:generate goyacc -o ev_parser.go -p ev ev_grammar.y
