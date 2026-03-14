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
	var result []*TraceEvent
	for _, ev := range events {
		if matchesPattern(ev, pattern) {
			result = append(result, ev)
		}
		// Also search sub-events recursively.
		subMatches := FilterEvents(ev.Subs, pattern)
		result = append(result, subMatches...)
	}
	return result
}

// matchesPattern checks if an event text contains the pattern.
func matchesPattern(ev *TraceEvent, pattern string) bool {
	return strings.Contains(ev.Text, pattern)
}

// FindEvent searches for an event matching a pattern, starting from an anchor
// (Python: EventTree.do_find).
func FindEvent(events []*TraceEvent, pattern string, reverse bool) (*TraceEvent, string) {
	flat := flattenEvents(events, "")
	if reverse {
		// Search in reverse order.
		for i := len(flat) - 1; i >= 0; i-- {
			if matchesPattern(flat[i].event, pattern) {
				return flat[i].event, flat[i].addr
			}
		}
	} else {
		for _, fe := range flat {
			if matchesPattern(fe.event, pattern) {
				return fe.event, fe.addr
			}
		}
	}
	return nil, ""
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
