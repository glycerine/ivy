package webui

// Full port of ivy_ev_viewer.py — event trace viewer for displaying
// verification event traces in the web UI.

import (
	"fmt"
	"sort"
	"strings"
)

// TraceEvent represents a single event in a verification trace
// (Python: ev.Event).
type TraceEvent struct {
	Text    string        `json:"text"`
	Subs    []*TraceEvent `json:"subs,omitempty"`
	Address string        `json:"address,omitempty"`
}

// NewTraceEvent creates a new trace event.
func NewTraceEvent(text string) *TraceEvent {
	return &TraceEvent{Text: text}
}

// AddSub adds a sub-event to this event.
func (e *TraceEvent) AddSub(sub *TraceEvent) {
	e.Subs = append(e.Subs, sub)
}

// HasSubs returns true if this event has sub-events.
func (e *TraceEvent) HasSubs() bool {
	return len(e.Subs) > 0
}

// EventTraceViewer manages the display and navigation of event traces
// (Python: classes EventTree, EventNoteBook, PatternList).
type EventTraceViewer struct {
	// Sheets holds named event trace sheets.
	Sheets map[string]*EventSheet

	// CurrentSheet is the name of the currently active sheet.
	CurrentSheet string

	// Patterns holds saved search patterns.
	Patterns []string

	sheetCounter int
}

// EventSheet represents one sheet/tab of events.
type EventSheet struct {
	Name   string        `json:"name"`
	Label  string        `json:"label"`
	Events []*TraceEvent `json:"events"`
}

// NewEventTraceViewer creates a new empty viewer.
func NewEventTraceViewer() *EventTraceViewer {
	return &EventTraceViewer{
		Sheets: make(map[string]*EventSheet),
	}
}

// NewSheet creates a new sheet from a list of events
// (Python: EventNoteBook.new_sheet).
func (v *EventTraceViewer) NewSheet(events []*TraceEvent) string {
	name := fmt.Sprintf("sht%d", v.sheetCounter)
	label := fmt.Sprintf("Sheet %d", v.sheetCounter)
	v.sheetCounter++

	// Assign addresses to events.
	for i, ev := range events {
		assignAddresses(ev, fmt.Sprintf("%d", i))
	}

	v.Sheets[name] = &EventSheet{
		Name:   name,
		Label:  label,
		Events: events,
	}
	v.CurrentSheet = name
	return name
}

// assignAddresses recursively assigns hierarchical addresses to events.
func assignAddresses(ev *TraceEvent, addr string) {
	ev.Address = addr
	for i, sub := range ev.Subs {
		assignAddresses(sub, fmt.Sprintf("%s/%d", addr, i))
	}
}

// GetSheet returns the named sheet, or nil if not found.
func (v *EventTraceViewer) GetSheet(name string) *EventSheet {
	return v.Sheets[name]
}

// CurrentEvents returns the events from the current sheet.
func (v *EventTraceViewer) CurrentEvents() []*TraceEvent {
	if sheet, ok := v.Sheets[v.CurrentSheet]; ok {
		return sheet.Events
	}
	return nil
}

// SheetNames returns sorted sheet names.
func (v *EventTraceViewer) SheetNames() []string {
	var names []string
	for n := range v.Sheets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Menus returns the event viewer menu structure
// (Python: EventTree.menus).
func (v *EventTraceViewer) Menus() []MenuDef {
	return []MenuDef{
		{
			Type:  "menu",
			Label: "Events",
			Items: []MenuItem{
				{Type: "button", Label: "Filter...", Action: "filter"},
				{Type: "button", Label: "Find reverse...", Action: "find_reverse"},
			},
		},
	}
}

// LookupEvent finds an event by its hierarchical address
// (Python: lookup function).
func LookupEvent(events []*TraceEvent, addr string) *TraceEvent {
	parts := strings.SplitN(addr, "/", 2)
	if len(parts) == 0 || len(events) == 0 {
		return nil
	}
	var idx int
	_, err := fmt.Sscanf(parts[0], "%d", &idx)
	if err != nil || idx < 0 || idx >= len(events) {
		return nil
	}
	if len(parts) == 1 {
		return events[idx]
	}
	return LookupEvent(events[idx].Subs, parts[1])
}

// FilterEvents filters events matching a pattern
// (Python: EventTree.do_filter).
func FilterEvents(events []*TraceEvent, pattern string) []*TraceEvent {
	pats, err := parseEventPatterns(pattern)
	if err != nil {
		return nil
	}
	var result []*TraceEvent
	for _, ev := range events {
		if matchesAnyEventPattern(ev, pats) {
			result = append(result, ev)
		}
		// Also search sub-events recursively.
		result = append(result, filterEventsParsed(ev.Subs, pats)...)
	}
	return result
}

func filterEventsParsed(events []*TraceEvent, pats []*parsedEvent) []*TraceEvent {
	var result []*TraceEvent
	for _, ev := range events {
		if matchesAnyEventPattern(ev, pats) {
			result = append(result, ev)
		}
		result = append(result, filterEventsParsed(ev.Subs, pats)...)
	}
	return result
}

// matchesPattern checks if an event matches the Python event-pattern language.
func matchesPattern(ev *TraceEvent, pattern string) bool {
	pats, err := parseEventPatterns(pattern)
	if err != nil {
		return false
	}
	return matchesAnyEventPattern(ev, pats)
}

// FindEvent searches for an event matching a pattern, starting from an anchor
// (Python: EventTree.do_find).
func FindEvent(events []*TraceEvent, pattern string, reverse bool) (*TraceEvent, string) {
	return FindEventFrom(events, pattern, reverse, "")
}

// FindEventFrom searches for an event matching pattern from an optional anchor.
// The anchor traversal mirrors ivy_ev_parser.EventFwdGen/EventRevGen: the
// selected anchor itself is skipped, then search proceeds forward or backward.
func FindEventFrom(events []*TraceEvent, pattern string, reverse bool, anchor string) (*TraceEvent, string) {
	pats, err := parseEventPatterns(pattern)
	if err != nil {
		return nil, ""
	}
	var flat []flatEntry
	if reverse {
		flat = eventRevEntries(events, anchor)
	} else {
		flat = eventFwdEntries(events, anchor)
	}
	for _, fe := range flat {
		if matchesAnyEventPattern(fe.event, pats) {
			return fe.event, fe.addr
		}
	}
	return nil, ""
}

type parsedEvent struct {
	rep      string
	args     []eventValue
	children []*parsedEvent
}

func (e *parsedEvent) match(pat *parsedEvent, binding map[string]string) bool {
	if e == nil || pat == nil {
		return false
	}
	if pat.rep != "*" && !strings.HasPrefix(e.rep, pat.rep) {
		return false
	}
	for i := 0; i < len(e.args) && i < len(pat.args); i++ {
		if !matchEventValue(e.args[i], pat.args[i], binding) {
			return false
		}
	}
	return true
}

type eventValue interface {
	key() string
	text() string
}

type eventSymbol struct {
	name string
}

func (s eventSymbol) key() string {
	return "S:" + s.name
}

func (s eventSymbol) text() string {
	return s.name
}

type eventApp struct {
	rep  string
	args []eventValue
}

func (a eventApp) key() string {
	var parts []string
	for _, arg := range a.args {
		parts = append(parts, arg.key())
	}
	return "A:" + a.rep + "(" + strings.Join(parts, ",") + ")"
}

func (a eventApp) text() string {
	return a.rep + "(" + eventValueListText(a.args) + ")"
}

type eventList struct {
	elems []eventValue
}

func (l eventList) key() string {
	var parts []string
	for _, elem := range l.elems {
		parts = append(parts, elem.key())
	}
	return "L:[" + strings.Join(parts, ",") + "]"
}

func (l eventList) text() string {
	return "[" + eventValueListText(l.elems) + "]"
}

type eventDict struct {
	entries map[string]eventValue
}

func (d eventDict) key() string {
	var keys []string
	for k := range d.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+":"+d.entries[k].key())
	}
	return "D:{" + strings.Join(parts, ",") + "}"
}

func (d eventDict) text() string {
	var keys []string
	for k := range d.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+":"+d.entries[k].text())
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (e *parsedEvent) text() string {
	if e == nil {
		return ""
	}
	if len(e.args) == 0 {
		return e.rep
	}
	return e.rep + "(" + eventValueListText(e.args) + ")"
}

func eventValueListText(values []eventValue) string {
	var parts []string
	for _, value := range values {
		parts = append(parts, value.text())
	}
	return strings.Join(parts, ",")
}

func matchesAnyEventPattern(ev *TraceEvent, pats []*parsedEvent) bool {
	actual, err := parseTraceEventText(ev.Text)
	if err != nil {
		return false
	}
	for _, pat := range pats {
		if actual.match(pat, make(map[string]string)) {
			return true
		}
	}
	return false
}

func matchEventValue(actual eventValue, pat eventValue, binding map[string]string) bool {
	switch p := pat.(type) {
	case eventSymbol:
		if p.name == "*" {
			return true
		}
		if strings.HasPrefix(p.name, "$") {
			key := actual.key()
			if old, ok := binding[p.name]; ok {
				return old == key
			}
			binding[p.name] = key
			return true
		}
		a, ok := actual.(eventSymbol)
		return ok && a.name == p.name
	case eventApp:
		a, ok := actual.(eventApp)
		if !ok {
			_, wildcard := actual.(eventSymbol)
			return wildcard && p.rep == "*"
		}
		if p.rep != "*" && a.rep != p.rep {
			return false
		}
		for i := 0; i < len(a.args) && i < len(p.args); i++ {
			if !matchEventValue(a.args[i], p.args[i], binding) {
				return false
			}
		}
		return true
	case eventList:
		if s, ok := actual.(eventSymbol); ok && s.name == "*" {
			return true
		}
		a, ok := actual.(eventList)
		if !ok || len(a.elems) != len(p.elems) {
			return false
		}
		for i := range a.elems {
			if !matchEventValue(a.elems[i], p.elems[i], binding) {
				return false
			}
		}
		return true
	case eventDict:
		if s, ok := actual.(eventSymbol); ok && s.name == "*" {
			return true
		}
		a, ok := actual.(eventDict)
		if !ok {
			return false
		}
		for k, v := range p.entries {
			av, ok := a.entries[k]
			if !ok || !matchEventValue(av, v, binding) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

type flatEntry struct {
	event *TraceEvent
	addr  string
}

func flattenEvents(events []*TraceEvent, prefix string) []flatEntry {
	var result []flatEntry
	for i, ev := range events {
		addr := fmt.Sprintf("%d", i)
		if prefix != "" {
			addr = prefix + "/" + addr
		}
		result = append(result, flatEntry{event: ev, addr: addr})
		result = append(result, flattenEvents(ev.Subs, addr)...)
	}
	return result
}

func eventFwdEntries(events []*TraceEvent, anchor string) []flatEntry {
	if anchor == "" {
		return flattenEvents(events, "")
	}
	return eventFwdEntriesRec(events, strings.Split(anchor, "/"), "")
}

func eventFwdEntriesRec(events []*TraceEvent, anchorParts []string, prefix string) []flatEntry {
	if len(anchorParts) == 0 || anchorParts[0] == "" {
		return flattenEvents(events, prefix)
	}
	idx := parseAddressIndex(anchorParts[0])
	if idx < 0 || idx >= len(events) {
		return flattenEvents(events, prefix)
	}
	var result []flatEntry
	if len(anchorParts) > 1 {
		childPrefix := joinEventAddress(prefix, idx)
		result = append(result, eventFwdEntriesRec(events[idx].Subs, anchorParts[1:], childPrefix)...)
	}
	for i := idx + 1; i < len(events); i++ {
		addr := joinEventAddress(prefix, i)
		result = append(result, flatEntry{event: events[i], addr: addr})
		result = append(result, flattenEvents(events[i].Subs, addr)...)
	}
	return result
}

func eventRevEntries(events []*TraceEvent, anchor string) []flatEntry {
	if anchor == "" {
		flat := flattenEvents(events, "")
		for i, j := 0, len(flat)-1; i < j; i, j = i+1, j-1 {
			flat[i], flat[j] = flat[j], flat[i]
		}
		return flat
	}
	return eventRevEntriesRec(events, strings.Split(anchor, "/"), "")
}

func eventRevEntriesRec(events []*TraceEvent, anchorParts []string, prefix string) []flatEntry {
	if len(anchorParts) == 0 || anchorParts[0] == "" {
		return reverseFlattenEvents(events, prefix)
	}
	idx := parseAddressIndex(anchorParts[0])
	if idx < 0 || idx >= len(events) {
		return reverseFlattenEvents(events, prefix)
	}
	var result []flatEntry
	if len(anchorParts) > 1 {
		childPrefix := joinEventAddress(prefix, idx)
		result = append(result, eventRevEntriesRec(events[idx].Subs, anchorParts[1:], childPrefix)...)
	}
	for i := idx - 1; i >= 0; i-- {
		addr := joinEventAddress(prefix, i)
		result = append(result, reverseFlattenEvents(events[i].Subs, addr)...)
		result = append(result, flatEntry{event: events[i], addr: addr})
	}
	return result
}

func reverseFlattenEvents(events []*TraceEvent, prefix string) []flatEntry {
	var result []flatEntry
	for i := len(events) - 1; i >= 0; i-- {
		addr := joinEventAddress(prefix, i)
		result = append(result, reverseFlattenEvents(events[i].Subs, addr)...)
		result = append(result, flatEntry{event: events[i], addr: addr})
	}
	return result
}

func parseAddressIndex(part string) int {
	var idx int
	if _, err := fmt.Sscanf(part, "%d", &idx); err != nil {
		return -1
	}
	return idx
}

func joinEventAddress(prefix string, idx int) string {
	if prefix == "" {
		return fmt.Sprintf("%d", idx)
	}
	return fmt.Sprintf("%s/%d", prefix, idx)
}

type eventPatternParser struct {
	input string
	pos   int
}

func parseEventPatterns(s string) ([]*parsedEvent, error) {
	p := &eventPatternParser{input: s}
	return p.parseEvents(0)
}

// ParseTraceEvents parses a Python .iev-style event trace into web trace nodes.
func ParseTraceEvents(s string) ([]*TraceEvent, error) {
	pats, err := parseEventPatterns(s)
	if err != nil {
		return nil, err
	}
	events := parsedEventsToTraceEvents(pats)
	for i, ev := range events {
		assignAddresses(ev, fmt.Sprintf("%d", i))
	}
	return events, nil
}

func parsedEventsToTraceEvents(pats []*parsedEvent) []*TraceEvent {
	events := make([]*TraceEvent, 0, len(pats))
	for _, pat := range pats {
		ev := NewTraceEvent(pat.text())
		ev.Subs = parsedEventsToTraceEvents(pat.children)
		events = append(events, ev)
	}
	return events
}

func (p *eventPatternParser) parseEvents(stop byte) ([]*parsedEvent, error) {
	var events []*parsedEvent
	for {
		p.skipSpaceAndSemis()
		if p.eof() {
			if stop != 0 {
				return nil, fmt.Errorf("expected %c", stop)
			}
			break
		}
		if stop != 0 && p.peek() == stop {
			p.pos++
			break
		}
		if stop == 0 && p.peek() == '}' {
			return nil, fmt.Errorf("unexpected }")
		}
		ev, err := p.parseEvent()
		if err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	if len(events) == 0 && stop == 0 {
		return nil, fmt.Errorf("empty event pattern")
	}
	return events, nil
}

func parseTraceEventText(s string) (*parsedEvent, error) {
	pats, err := parseEventPatterns(s)
	if err != nil {
		return nil, err
	}
	return pats[0], nil
}

func (p *eventPatternParser) parseEvent() (*parsedEvent, error) {
	p.skipSpace()
	if !p.eof() && (p.peek() == '>' || p.peek() == '<') {
		p.pos++
		p.skipSpace()
	}
	rep, err := p.parseSymbol()
	if err != nil {
		return nil, err
	}
	ev := &parsedEvent{rep: rep}
	p.skipSpace()
	if !p.eof() && p.peek() == '(' {
		args, err := p.parseParenArgs()
		if err != nil {
			return nil, err
		}
		ev.args = args
	}
	p.skipSpace()
	if !p.eof() && p.peek() == '{' {
		p.pos++
		children, err := p.parseEvents('}')
		if err != nil {
			return nil, err
		}
		ev.children = children
	}
	return ev, nil
}

func (p *eventPatternParser) parseParenArgs() ([]eventValue, error) {
	p.pos++
	p.skipSpace()
	if !p.eof() && p.peek() == ')' {
		p.pos++
		return nil, nil
	}
	args, err := p.parseValueList(')')
	if err != nil {
		return nil, err
	}
	if p.eof() || p.peek() != ')' {
		return nil, fmt.Errorf("expected )")
	}
	p.pos++
	return args, nil
}

func (p *eventPatternParser) parseValueList(end byte) ([]eventValue, error) {
	var args []eventValue
	for {
		p.skipSpace()
		if p.eof() || p.peek() == end {
			break
		}
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		args = append(args, v)
		p.skipSpace()
		if !p.eof() && p.peek() == ',' {
			p.pos++
			continue
		}
		break
	}
	return args, nil
}

func (p *eventPatternParser) parseValue() (eventValue, error) {
	p.skipSpace()
	if p.eof() {
		return nil, fmt.Errorf("unexpected end of value")
	}
	switch p.peek() {
	case '[':
		return p.parseListValue()
	case '{':
		return p.parseDictValue()
	default:
		sym, err := p.parseSymbol()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if !p.eof() && p.peek() == '(' {
			args, err := p.parseParenArgs()
			if err != nil {
				return nil, err
			}
			return eventApp{rep: sym, args: args}, nil
		}
		return eventSymbol{name: sym}, nil
	}
}

func (p *eventPatternParser) parseListValue() (eventValue, error) {
	p.pos++
	args, err := p.parseValueList(']')
	if err != nil {
		return nil, err
	}
	if p.eof() || p.peek() != ']' {
		return nil, fmt.Errorf("expected ]")
	}
	p.pos++
	return eventList{elems: args}, nil
}

func (p *eventPatternParser) parseDictValue() (eventValue, error) {
	p.pos++
	entries := make(map[string]eventValue)
	for {
		p.skipSpace()
		if !p.eof() && p.peek() == '}' {
			p.pos++
			return eventDict{entries: entries}, nil
		}
		key, err := p.parseSymbol()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if p.eof() || p.peek() != ':' {
			return nil, fmt.Errorf("expected :")
		}
		p.pos++
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		entries[key] = value
		p.skipSpace()
		if !p.eof() && p.peek() == ',' {
			p.pos++
		}
	}
}

func (p *eventPatternParser) parseSymbol() (string, error) {
	p.skipSpace()
	if p.eof() {
		return "", fmt.Errorf("expected symbol")
	}
	if p.peek() == '"' {
		start := p.pos
		p.pos++
		for !p.eof() && p.peek() != '"' {
			p.pos++
		}
		if p.eof() {
			return "", fmt.Errorf("unterminated quoted symbol")
		}
		p.pos++
		return p.input[start:p.pos], nil
	}
	if p.peek() == '*' {
		p.pos++
		return "*", nil
	}
	start := p.pos
	for !p.eof() && isEventSymbolChar(p.peek()) {
		p.pos++
	}
	if start == p.pos {
		return "", fmt.Errorf("expected symbol")
	}
	return p.input[start:p.pos], nil
}

func (p *eventPatternParser) skipBalanced(open, close byte) error {
	if p.eof() || p.peek() != open {
		return fmt.Errorf("expected %c", open)
	}
	depth := 0
	for !p.eof() {
		ch := p.peek()
		p.pos++
		switch ch {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return nil
			}
		case '"':
			for !p.eof() && p.peek() != '"' {
				p.pos++
			}
			if !p.eof() {
				p.pos++
			}
		}
	}
	return fmt.Errorf("unclosed %c", open)
}

func (p *eventPatternParser) skipSpaceAndSemis() {
	for !p.eof() {
		ch := p.peek()
		if ch != ';' && ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' {
			return
		}
		p.pos++
	}
}

func (p *eventPatternParser) skipSpace() {
	for !p.eof() {
		ch := p.peek()
		if ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' {
			return
		}
		p.pos++
	}
}

func (p *eventPatternParser) eof() bool {
	return p.pos >= len(p.input)
}

func (p *eventPatternParser) peek() byte {
	return p.input[p.pos]
}

func isEventSymbolChar(ch byte) bool {
	return ch == '_' || ch == '.' || ch == '$' || ch == '-' ||
		(ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9')
}

// FormatEvent formats a single event for display.
func FormatEvent(ev *TraceEvent) string {
	if ev == nil {
		return ""
	}
	return ev.Text
}

// FormatTrace formats an entire trace as an indented string
// (Python: not directly ported but useful for web rendering).
func FormatTrace(events []*TraceEvent) string {
	var sb strings.Builder
	formatTraceIndented(&sb, events, 0)
	return sb.String()
}

func formatTraceIndented(sb *strings.Builder, events []*TraceEvent, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, ev := range events {
		sb.WriteString(prefix)
		sb.WriteString(ev.Text)
		sb.WriteString("\n")
		if ev.HasSubs() {
			formatTraceIndented(sb, ev.Subs, indent+1)
		}
	}
}

// AddPattern adds a search pattern to the saved list.
func (v *EventTraceViewer) AddPattern(pattern string) {
	v.Patterns = append(v.Patterns, pattern)
}

// RemovePattern removes a pattern by index.
func (v *EventTraceViewer) RemovePattern(idx int) error {
	if idx < 0 || idx >= len(v.Patterns) {
		return fmt.Errorf("pattern index %d out of range", idx)
	}
	v.Patterns = append(v.Patterns[:idx], v.Patterns[idx+1:]...)
	return nil
}

// ClearPatterns removes all saved patterns.
func (v *EventTraceViewer) ClearPatterns() {
	v.Patterns = nil
}

// SavePatterns returns the patterns as a newline-separated string.
func (v *EventTraceViewer) SavePatterns() string {
	return strings.Join(v.Patterns, "\n")
}

// LoadPatterns loads patterns from a newline-separated string.
func (v *EventTraceViewer) LoadPatterns(data string) {
	if data == "" {
		return
	}
	lines := strings.Split(strings.TrimSpace(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			v.Patterns = append(v.Patterns, line)
		}
	}
}
