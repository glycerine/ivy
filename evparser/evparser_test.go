package evparser

import (
	"testing"
)

func TestParseEmpty(t *testing.T) {
	evs, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evs) != 0 {
		t.Errorf("expected 0 events, got %d", len(evs))
	}
}

func TestParseSingleEvent(t *testing.T) {
	evs, err := Parse("foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Rep != "foo" {
		t.Errorf("expected rep 'foo', got %q", evs[0].Rep)
	}
	if evs[0].Dir != DirNone {
		t.Errorf("expected DirNone, got %d", evs[0].Dir)
	}
}

func TestParseInOutEvents(t *testing.T) {
	evs, err := Parse("> send < recv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evs) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evs))
	}
	if evs[0].Dir != DirIn || evs[0].Rep != "send" {
		t.Errorf("expected '> send', got dir=%d rep=%q", evs[0].Dir, evs[0].Rep)
	}
	if evs[1].Dir != DirOut || evs[1].Rep != "recv" {
		t.Errorf("expected '< recv', got dir=%d rep=%q", evs[1].Dir, evs[1].Rep)
	}
}

func TestParseEventWithArgs(t *testing.T) {
	evs, err := Parse("call(x,y,z)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if len(evs[0].Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(evs[0].Args))
	}
	if evs[0].Args[0].String() != "x" || evs[0].Args[1].String() != "y" || evs[0].Args[2].String() != "z" {
		t.Errorf("unexpected args: %v", evs[0].Args)
	}
}

func TestParseEventWithChildren(t *testing.T) {
	evs, err := Parse("parent{child1 child2}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if len(evs[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(evs[0].Children))
	}
	if evs[0].Children[0].Rep != "child1" || evs[0].Children[1].Rep != "child2" {
		t.Errorf("unexpected children: %v", evs[0].Children)
	}
}

func TestParseAppValue(t *testing.T) {
	evs, err := Parse("ev(foo(a,b))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	app, ok := evs[0].Args[0].(*App)
	if !ok {
		t.Fatalf("expected App, got %T", evs[0].Args[0])
	}
	if app.Rep != "foo" || len(app.Args) != 2 {
		t.Errorf("expected foo(a,b), got %v", app)
	}
}

func TestParseListValue(t *testing.T) {
	evs, err := Parse("ev([a,b,c])")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lv, ok := evs[0].Args[0].(*ListValue)
	if !ok {
		t.Fatalf("expected ListValue, got %T", evs[0].Args[0])
	}
	if len(lv.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(lv.Items))
	}
}

func TestParseEmptyList(t *testing.T) {
	evs, err := Parse("ev([])")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lv, ok := evs[0].Args[0].(*ListValue)
	if !ok {
		t.Fatalf("expected ListValue, got %T", evs[0].Args[0])
	}
	if len(lv.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(lv.Items))
	}
}

func TestParseDictValue(t *testing.T) {
	evs, err := Parse("ev({x:a,y:b})")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dv, ok := evs[0].Args[0].(*DictValue)
	if !ok {
		t.Fatalf("expected DictValue, got %T", evs[0].Args[0])
	}
	if len(dv.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(dv.Entries))
	}
	if _, ok := dv.Get("x"); !ok {
		t.Error("expected key 'x'")
	}
	if _, ok := dv.Get("y"); !ok {
		t.Error("expected key 'y'")
	}
}

func TestParseEmptyDict(t *testing.T) {
	evs, err := Parse("ev({})")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dv, ok := evs[0].Args[0].(*DictValue)
	if !ok {
		t.Fatalf("expected DictValue, got %T", evs[0].Args[0])
	}
	if len(dv.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(dv.Entries))
	}
}

func TestParseSemicolon(t *testing.T) {
	evs, err := Parse("ev1; ev2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evs) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evs))
	}
}

func TestSymbolMatch(t *testing.T) {
	s := &Symbol{Name: "foo"}
	pat := &Symbol{Name: "foo"}
	if !s.Match(pat, nil) {
		t.Error("expected match")
	}
	wild := &Symbol{Name: "*"}
	if !s.Match(wild, nil) {
		t.Error("expected wildcard match")
	}
	noMatch := &Symbol{Name: "bar"}
	if s.Match(noMatch, nil) {
		t.Error("expected no match")
	}
}

func TestSymbolBindingMatch(t *testing.T) {
	s := &Symbol{Name: "foo"}
	pat := &Symbol{Name: "$x"}
	binding := make(map[string]Value)
	if !s.Match(pat, binding) {
		t.Error("expected binding match")
	}
	if binding["$x"].(*Symbol).Name != "foo" {
		t.Errorf("expected binding $x=foo, got %v", binding["$x"])
	}
	// Second match should check existing binding
	s2 := &Symbol{Name: "foo"}
	if !s2.Match(pat, binding) {
		t.Error("expected second binding match (same value)")
	}
	s3 := &Symbol{Name: "bar"}
	if s3.Match(pat, binding) {
		t.Error("expected binding mismatch")
	}
}

func TestEventMatch(t *testing.T) {
	ev := &Event{Rep: "send", Args: []Value{&Symbol{Name: "x"}}}
	pat := &Event{Rep: "send", Args: []Value{&Symbol{Name: "*"}}}
	if !ev.Match(pat, nil) {
		t.Error("expected event match")
	}
}

func TestEventGen(t *testing.T) {
	child := &Event{Rep: "child"}
	parent := &Event{Rep: "parent", Children: Events{child}}
	evs := EventGen(Events{parent})
	if len(evs) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evs))
	}
	if evs[0].Rep != "parent" || evs[1].Rep != "child" {
		t.Errorf("unexpected order: %v, %v", evs[0].Rep, evs[1].Rep)
	}
}

func TestRoundTripString(t *testing.T) {
	input := "> send(x,y){< recv(z)}"
	evs, err := Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	s := evs.String()
	// Parse the string representation back
	evs2, err := Parse(s)
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}
	if evs2.String() != s {
		t.Errorf("round-trip mismatch:\n  got:  %s\n  want: %s", evs2.String(), s)
	}
}
