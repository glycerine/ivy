package isolate

import (
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// --- helpers ---

func mkConst(name string) *lg.Symbol {
	return lg.NewSymbol(name, lg.Boolean)
}

func mkSort(name string) lg.Sort {
	return &lg.UninterpretedSort{Name: name}
}

func mkModule() *module.Module {
	m := module.New()
	return m
}

func mkModuleWithActions(acts map[string]actions.Action) *module.Module {
	m := mkModule()
	for name, act := range acts {
		m.Actions[name] = act
	}
	return m
}

// --- IsolateRole ---

func TestIsolateRoleString(t *testing.T) {
	tests := []struct {
		role IsolateRole
		want string
	}{
		{RoleVerified, "verified"},
		{RolePresent, "present"},
		{RoleOpaque, "opaque"},
		{IsolateRole(99), "IsolateRole(99)"},
	}
	for _, tt := range tests {
		got := tt.role.String()
		if got != tt.want {
			t.Errorf("IsolateRole(%d).String() = %q, want %q", int(tt.role), got, tt.want)
		}
	}
}

// --- ComponentInfo ---

func TestNewComponentInfo(t *testing.T) {
	ci := NewComponentInfo("mycomp", RoleVerified)
	if ci.Name != "mycomp" {
		t.Errorf("Name = %q, want %q", ci.Name, "mycomp")
	}
	if ci.Role != RoleVerified {
		t.Errorf("Role = %v, want RoleVerified", ci.Role)
	}
	if ci.Actions == nil {
		t.Error("Actions map should be initialized")
	}
}

// --- LookupAction ---

func TestLookupAction(t *testing.T) {
	seq := actions.NewSequence()
	m := mkModuleWithActions(map[string]actions.Action{"foo": seq})

	act, err := LookupAction(m, "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if act != seq {
		t.Error("returned wrong action")
	}
}

func TestLookupActionNotFound(t *testing.T) {
	m := mkModule()
	_, err := LookupAction(m, "nonexistent")
	if err == nil {
		t.Error("expected error for missing action")
	}
}

// TestLookupActionBadType removed: mod.Actions is now map[string]module.Action,
// so non-Action values cannot be inserted.

// --- SummarizeAction ---

func TestSummarizeActionBasic(t *testing.T) {
	// An action with formals and a body.
	body := actions.NewSequence(actions.NewAssumeAction(mkConst("p")))
	body.SetFormalParams([]*lg.Symbol{mkConst("x")})
	body.SetFormalReturns([]*lg.Symbol{mkConst("r")})

	summarized := SummarizeAction(body)

	// Should preserve formals.
	if len(summarized.GetFormalParams()) != 1 {
		t.Errorf("formal params len = %d, want 1", len(summarized.GetFormalParams()))
	}
	if len(summarized.GetFormalReturns()) != 1 {
		t.Errorf("formal returns len = %d, want 1", len(summarized.GetFormalReturns()))
	}

	// Body should be empty (or just havoc actions).
	seq, ok := summarized.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected *Sequence, got %T", summarized)
	}
	// "r" is not in params, so no havoc.
	if len(seq.Elems) != 0 {
		t.Errorf("expected 0 children (no in/out overlap), got %d", len(seq.Elems))
	}
}

func TestSummarizeActionWithInOutParams(t *testing.T) {
	body := actions.NewSequence()
	p := mkConst("x")
	body.SetFormalParams([]*lg.Symbol{p})
	body.SetFormalReturns([]*lg.Symbol{p}) // same param is both in and out

	isoCfg := module.NewIsolateConfig()
	isoCfg.IsolateMode = "check"

	summarized := SummarizeAction(body, isoCfg)
	seq, ok := summarized.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected *Sequence, got %T", summarized)
	}
	// x is in both params and returns, should be havoced.
	if len(seq.Elems) != 1 {
		t.Errorf("expected 1 havoc child, got %d", len(seq.Elems))
	}
}

func TestSummarizeActionNonCheckMode(t *testing.T) {
	body := actions.NewSequence()
	p := mkConst("x")
	body.SetFormalParams([]*lg.Symbol{p})
	body.SetFormalReturns([]*lg.Symbol{p})

	isoCfg := module.NewIsolateConfig()
	isoCfg.IsolateMode = "test"

	summarized := SummarizeAction(body, isoCfg)
	seq, ok := summarized.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected *Sequence, got %T", summarized)
	}
	// In test mode, no havoc.
	if len(seq.Elems) != 0 {
		t.Errorf("expected 0 children in test mode, got %d", len(seq.Elems))
	}
}

// --- EmptyClone ---

func TestEmptyClone(t *testing.T) {
	body := actions.NewSequence(actions.NewAssumeAction(mkConst("p")))
	body.SetFormalParams([]*lg.Symbol{mkConst("x")})

	clone := EmptyClone(body)
	seq, ok := clone.(*actions.Sequence)
	if !ok {
		t.Fatalf("expected *Sequence, got %T", clone)
	}
	if len(seq.Elems) != 0 {
		t.Errorf("empty clone should have 0 children, got %d", len(seq.Elems))
	}
	if len(clone.GetFormalParams()) != 1 {
		t.Errorf("empty clone should preserve formal params")
	}
}

// --- CanonAct ---

func TestCanonAct(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"ext:foo", "foo"},
		{"ext:a.b.c", "a.b.c"},
		{"foo", "foo"},
		{"", ""},
		{"ext:", ""},
	}
	for _, tt := range tests {
		got := CanonAct(tt.input)
		if got != tt.want {
			t.Errorf("CanonAct(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Ancestors ---

func TestAncestors(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"a.b.c", []string{"a.b.c", "a.b", "a"}},
		{"a", []string{"a"}},
		{"x.y", []string{"x.y", "x"}},
	}
	for _, tt := range tests {
		got := Ancestors(tt.name, ".")
		if len(got) != len(tt.want) {
			t.Errorf("Ancestors(%q) len = %d, want %d", tt.name, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("Ancestors(%q)[%d] = %q, want %q", tt.name, i, got[i], tt.want[i])
			}
		}
	}
}

// --- Presentable ---

func TestPresentable(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"fml:x", "x"},
		{"abc", "abc"},
		{"a:b:c", "c"},
		{"", ""},
	}
	for _, tt := range tests {
		got := Presentable(tt.input)
		if got != tt.want {
			t.Errorf("Presentable(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- ClassifyComponents ---

func TestClassifyComponents(t *testing.T) {
	m := mkModule()
	m.Hierarchy["this"] = map[string]bool{"a": true, "b": true, "c": true}
	m.Hierarchy["a"] = map[string]bool{"x": true}

	verified := map[string]bool{"a": true}
	present := map[string]bool{"a": true, "b": true}

	roles := ClassifyComponents(m, verified, present)

	if roles["a"] != RoleVerified {
		t.Errorf("a should be RoleVerified, got %v", roles["a"])
	}
	if roles["b"] != RolePresent {
		t.Errorf("b should be RolePresent, got %v", roles["b"])
	}
	if roles["c"] != RoleOpaque {
		t.Errorf("c should be RoleOpaque, got %v", roles["c"])
	}
	// Child of verified component.
	if roles["a.x"] != RoleOpaque {
		// a.x is not in verified or present itself, so it's opaque.
		t.Errorf("a.x should be RoleOpaque (not explicitly listed), got %v", roles["a.x"])
	}
}

// --- GetIsolateInfo ---

func TestGetIsolateInfo(t *testing.T) {
	m := mkModule()
	verified, present := GetIsolateInfo(m, []string{"a", "b"}, []string{"c"}, "impl")

	if !verified["a"] || !verified["b"] {
		t.Error("a and b should be verified")
	}
	if !verified["a.impl"] || !verified["b.impl"] {
		t.Error("a.impl and b.impl should be verified")
	}
	if !present["a"] || !present["b"] || !present["c"] {
		t.Error("a, b, c should all be present")
	}
	if verified["c"] {
		t.Error("c should NOT be verified")
	}
}

// --- StartsWithSome / StartsWithEqSome ---

func TestStartsWithEqSome(t *testing.T) {
	m := mkModule()
	prefixes := map[string]bool{"a": true}

	// "a" is directly in prefixes.
	if !StartsWithEqSome("a", prefixes, m, nil) {
		t.Error("a should match")
	}

	// "a.b" is a child of "a".
	if !StartsWithEqSome("a.b", prefixes, m, nil) {
		t.Error("a.b should match via parent")
	}

	// "b" is not in prefixes or under one.
	if StartsWithEqSome("b", prefixes, m, nil) {
		t.Error("b should not match")
	}
}

func TestStartsWithSomePrivates(t *testing.T) {
	m := mkModule()
	m.Privates["a.b"] = true
	prefixes := map[string]bool{"a": true}

	// "a.b.c" has parent "a.b" which is private, so should not match.
	if StartsWithEqSome("a.b.c", prefixes, m, nil) {
		t.Error("a.b.c should not match because a.b is private")
	}
}

func TestStartsWithSomeImplMap(t *testing.T) {
	m := mkModule()
	prefixes := map[string]bool{"real": true}
	implMap := map[string]string{"alias": "real"}

	if !StartsWithEqSome("alias", prefixes, m, implMap) {
		t.Error("alias should match via implementation map")
	}
}

// --- AddMixins ---

func TestAddMixinsNoMixins(t *testing.T) {
	m := mkModule()
	seq := actions.NewSequence()
	result := AddMixins(m, "foo", seq, nil)
	if result != seq {
		t.Error("with no mixins, should return original action")
	}
}

// --- IsolateComponent ---

func TestIsolateComponentEmpty(t *testing.T) {
	m := mkModule()
	err := IsolateComponent(m, "", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsolateComponentNotFound(t *testing.T) {
	m := mkModule()
	err := IsolateComponent(m, "nonexistent", nil, nil, nil)
	if err == nil {
		t.Error("expected error for undefined isolate")
	}
}

func TestIsolateComponentFound(t *testing.T) {
	m := mkModule()
	m.Isolates["test_iso"] = &ast.IsolateDef{}
	err := IsolateComponent(m, "test_iso", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- GetCallsMods ---

func TestGetCallsMods(t *testing.T) {
	// An action that calls "bar" and assigns to "x".
	call := actions.NewCallAction(mkConst("bar"))
	assign := actions.NewAssignAction(mkConst("x"), mkConst("val"))
	seq := actions.NewSequence(call, assign)

	calls, mods := GetCallsMods(seq)

	if len(calls) != 1 || calls[0] != "bar" {
		t.Errorf("calls = %v, want [bar]", calls)
	}
	if len(mods) != 1 || mods[0] != "x" {
		t.Errorf("mods = %v, want [x]", mods)
	}
}

func TestGetCallsModsHavoc(t *testing.T) {
	havoc := actions.NewHavocAction(mkConst("y"))
	seq := actions.NewSequence(havoc)

	_, mods := GetCallsMods(seq)
	if len(mods) != 1 || mods[0] != "y" {
		t.Errorf("mods = %v, want [y]", mods)
	}
}

func TestGetCallsModsEmpty(t *testing.T) {
	seq := actions.NewSequence()
	calls, mods := GetCallsMods(seq)
	if len(calls) != 0 {
		t.Errorf("calls should be empty, got %v", calls)
	}
	if len(mods) != 0 {
		t.Errorf("mods should be empty, got %v", mods)
	}
}

// --- HasSideEffect ---

func TestHasSideEffectNoEffect(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()
	seq := actions.NewSequence()
	actionMap := map[string]actions.Action{"foo": seq}

	if HasSideEffect(m, "foo", actionMap) {
		t.Error("empty sequence should have no side effect")
	}
}

func TestHasSideEffectWithAssert(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()
	assertAct := actions.NewAssertAction(mkConst("p"))
	seq := actions.NewSequence(assertAct)
	actionMap := map[string]actions.Action{"foo": seq}

	if !HasSideEffect(m, "foo", actionMap) {
		t.Error("action with assert should have side effect")
	}
}

func TestHasSideEffectWithSigModification(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()
	m.Sig.Symbols["x"] = &il.SymbolEntry{Name: "x", Sort: lg.Boolean}

	assign := actions.NewAssignAction(mkConst("x"), mkConst("val"))
	seq := actions.NewSequence(assign)
	actionMap := map[string]actions.Action{"foo": seq}

	if !HasSideEffect(m, "foo", actionMap) {
		t.Error("assignment to sig symbol should have side effect")
	}
}

func TestHasSideEffectThroughCall(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()
	m.Sig.Symbols["x"] = &il.SymbolEntry{Name: "x", Sort: lg.Boolean}

	assign := actions.NewAssignAction(mkConst("x"), mkConst("val"))
	barSeq := actions.NewSequence(assign)

	call := actions.NewCallAction(mkConst("bar"))
	fooSeq := actions.NewSequence(call)

	actionMap := map[string]actions.Action{
		"foo": fooSeq,
		"bar": barSeq,
	}

	if !HasSideEffect(m, "foo", actionMap) {
		t.Error("side effect through call should be detected")
	}
}

// --- ActionCallGraph ---

func TestActionCallGraph(t *testing.T) {
	m := mkModule()
	call1 := actions.NewCallAction(mkConst("b"))
	call2 := actions.NewCallAction(mkConst("c"))
	m.Actions["a"] = actions.NewSequence(call1, call2)
	m.Actions["b"] = actions.NewSequence()
	m.Actions["c"] = actions.NewSequence()

	graph := ActionCallGraph(m)
	if len(graph["a"]) != 2 {
		t.Errorf("a should call 2 actions, got %d", len(graph["a"]))
	}
	if len(graph["b"]) != 0 {
		t.Errorf("b should call 0 actions, got %d", len(graph["b"]))
	}
}

// --- TransitiveCallees ---

func TestTransitiveCallees(t *testing.T) {
	graph := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {},
		"d": {},
	}
	result := TransitiveCallees("a", graph)

	expected := map[string]bool{"a": true, "b": true, "c": true, "d": true}
	for name := range expected {
		if !result[name] {
			t.Errorf("%s should be in transitive callees", name)
		}
	}
}

func TestTransitiveCalleesWithCycle(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"a"}, // cycle
	}
	result := TransitiveCallees("a", graph)
	if !result["a"] || !result["b"] {
		t.Error("cycle should be handled without infinite loop")
	}
}

// --- CollectSortDestructors ---

func TestCollectSortDestructors(t *testing.T) {
	m := mkModule()
	destrSort, _ := lg.NewFunctionSort(mkSort("MySort"), lg.Boolean)
	destr := lg.NewSymbol("get_field", destrSort)
	m.SortDestructors["MySort"] = []*lg.Symbol{destr}

	result := make(map[string]bool)
	CollectSortDestructors(m, "MySort", result, make(map[string]bool))

	if !result["get_field"] {
		t.Error("get_field should be in destructors")
	}
}

func TestCollectSortDestructorsWithVariants(t *testing.T) {
	m := mkModule()
	m.Variants["Base"] = []lg.Sort{mkSort("Variant1")}

	destrSort, _ := lg.NewFunctionSort(mkSort("Variant1"), lg.Boolean)
	destr := lg.NewSymbol("v1_field", destrSort)
	m.SortDestructors["Variant1"] = []*lg.Symbol{destr}

	result := make(map[string]bool)
	CollectSortDestructors(m, "Base", result, make(map[string]bool))

	if !result["v1_field"] {
		t.Error("v1_field should be collected from variant")
	}
}

// --- StripSort ---

func TestStripSort(t *testing.T) {
	// Function sort: A * B -> C, strip 1 param.
	sortA := mkSort("A")
	sortB := mkSort("B")
	sortC := mkSort("C")
	fs, err := lg.NewFunctionSort(sortA, sortB, sortC)
	if err != nil {
		t.Fatal(err)
	}

	result := StripSort(fs, 1)
	rfs, ok := result.(*lg.FunctionSort)
	if !ok {
		t.Fatalf("expected FunctionSort, got %T", result)
	}
	if rfs.Arity() != 1 {
		t.Errorf("arity = %d, want 1", rfs.Arity())
	}
}

func TestStripSortAllDomain(t *testing.T) {
	// Function sort: A -> B, strip 1 param.
	fs, err := lg.NewFunctionSort(mkSort("A"), mkSort("B"))
	if err != nil {
		t.Fatal(err)
	}

	result := StripSort(fs, 1)
	// Should return just the range sort.
	if _, ok := result.(*lg.UninterpretedSort); !ok {
		t.Errorf("expected UninterpretedSort, got %T", result)
	}
}

func TestStripSortZero(t *testing.T) {
	fs, _ := lg.NewFunctionSort(mkSort("A"), mkSort("B"))
	result := StripSort(fs, 0)
	if result != fs {
		t.Error("strip 0 should return original")
	}
}

func TestStripSortNonFunction(t *testing.T) {
	s := mkSort("X")
	result := StripSort(s, 1)
	if result != s {
		t.Error("stripping non-function sort should return original")
	}
}

// --- StripMapLookup ---

func TestStripMapLookup(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()

	sm := StripMap{
		"server": {"s"},
	}

	// "server.foo" starts with "server.", should match.
	result := StripMapLookup("server.foo", sm, m)
	if len(result) != 1 || result[0] != "s" {
		t.Errorf("StripMapLookup(server.foo) = %v, want [s]", result)
	}

	// "client.foo" does not match.
	result = StripMapLookup("client.foo", sm, m)
	if result != nil {
		t.Errorf("StripMapLookup(client.foo) = %v, want nil", result)
	}
}

func TestStripMapLookupGlobalParam(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()
	m.Attributes["x.global_parameter"] = true

	sm := StripMap{"x": {"p"}}
	result := StripMapLookup("x", sm, m)
	if result != nil {
		t.Error("global parameter should return nil")
	}
}

func TestStripMapLookupSort(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()
	m.Sig.Sorts["mysort"] = mkSort("mysort")

	sm := StripMap{"mysort": {"p"}}
	result := StripMapLookup("mysort", sm, m)
	if result != nil {
		t.Error("sort name should return nil")
	}
}

// --- StripIsolate ---

func TestStripIsolateEmpty(t *testing.T) {
	m := mkModule()
	err := StripIsolate(m, StripMap{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStripIsolateStripsFormalParams(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()

	act := actions.NewSequence()
	act.SetFormalParams([]*lg.Symbol{mkConst("s"), mkConst("x")})
	act.SetFormalReturns([]*lg.Symbol{mkConst("r")})
	m.Actions["server.do"] = act

	sm := StripMap{"server": {"s"}}
	err := StripIsolate(m, sm, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	strippedIface := m.Actions["server.do"]
	stripped, ok := strippedIface.(actions.Action)
	if !ok {
		t.Fatal("action should still be an Action after stripping")
	}
	fp := stripped.GetFormalParams()
	if len(fp) != 1 || fp[0].Name != "x" {
		t.Errorf("formal params after strip = %v, want [x]", fp)
	}
}

// --- StripSortFromModule ---

func TestStripSortFromModule(t *testing.T) {
	m := mkModule()
	m.Sig = il.NewSig()
	m.Sig.Sorts["mysort"] = mkSort("mysort")
	m.SortOrder = []string{"bool", "mysort", "int"}
	m.SortDestructors["mysort"] = []*lg.Symbol{mkConst("d")}
	m.DestructorSorts["mysort"] = mkSort("mysort")

	err := StripSortFromModule(m, "mysort")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := m.Sig.Sorts["mysort"]; ok {
		t.Error("sort should be removed from signature")
	}
	if _, ok := m.SortDestructors["mysort"]; ok {
		t.Error("sort destructors should be removed")
	}
	if _, ok := m.DestructorSorts["mysort"]; ok {
		t.Error("destructor sorts should be removed")
	}
	for _, s := range m.SortOrder {
		if s == "mysort" {
			t.Error("sort should be removed from sort order")
		}
	}
}

// --- CheckInterference (stubbed) ---

func TestCheckInterferenceDisabled(t *testing.T) {
	m := mkModule()
	m.Cfg.IsolateCfg.DoCheckInterference = false

	err := CheckInterference(m, nil, nil)
	if err != nil {
		t.Errorf("should return nil when disabled: %v", err)
	}
}

func TestCheckInterferenceEnabled(t *testing.T) {
	m := mkModule()
	m.Cfg.IsolateCfg.DoCheckInterference = true

	err := CheckInterference(m, nil, nil)
	if err != nil {
		t.Errorf("stubbed version should return nil: %v", err)
	}
}

// --- ConeOfInfluenceFilter (stubbed) ---

func TestConeOfInfluenceFilterDisabled(t *testing.T) {
	m := mkModule()
	m.Cfg.IsolateCfg.ConeOfInfluence = false

	err := ConeOfInfluenceFilter(m, nil)
	if err != nil {
		t.Errorf("should return nil when disabled: %v", err)
	}
}

// --- Fuzz ---

func FuzzCanonAct(f *testing.F) {
	f.Add("ext:foo")
	f.Add("foo")
	f.Add("")
	f.Add("ext:")
	f.Add("ext:ext:nested")
	f.Add("a.b.c")
	f.Add("ext:a.b.c")

	f.Fuzz(func(t *testing.T, name string) {
		result := CanonAct(name)

		// Result should never start with "ext:" (single strip).
		if len(name) >= 4 && name[:4] == "ext:" {
			if result != name[4:] {
				t.Errorf("CanonAct(%q) = %q, want %q", name, result, name[4:])
			}
		} else {
			if result != name {
				t.Errorf("CanonAct(%q) = %q, want %q", name, result, name)
			}
		}

		// Ancestors should not panic.
		anc := Ancestors(result, ".")
		if len(anc) == 0 {
			t.Error("Ancestors should return at least one element")
		}

		// Presentable should not panic.
		_ = Presentable(result)
	})
}

func FuzzSummarizeAction(f *testing.F) {
	f.Add(0, 0, true)
	f.Add(1, 0, true)
	f.Add(2, 1, true)
	f.Add(0, 0, false)

	f.Fuzz(func(t *testing.T, nParams, nReturns int, checkMode bool) {
		if nParams < 0 || nParams > 10 || nReturns < 0 || nReturns > 10 {
			return
		}

		isoCfg := module.NewIsolateConfig()
		if checkMode {
			isoCfg.IsolateMode = "check"
		} else {
			isoCfg.IsolateMode = "test"
		}

		body := actions.NewSequence()
		params := make([]*lg.Symbol, nParams)
		for i := range params {
			params[i] = lg.NewSymbol("p"+string(rune('a'+i)), lg.Boolean)
		}
		returns := make([]*lg.Symbol, nReturns)
		for i := range returns {
			returns[i] = lg.NewSymbol("r"+string(rune('a'+i)), lg.Boolean)
		}
		body.SetFormalParams(params)
		body.SetFormalReturns(returns)

		// Should not panic.
		result := SummarizeAction(body, isoCfg)
		if result == nil {
			t.Error("SummarizeAction returned nil")
		}

		// Formals should be preserved.
		if len(result.GetFormalParams()) != nParams {
			t.Errorf("params: got %d, want %d", len(result.GetFormalParams()), nParams)
		}
		if len(result.GetFormalReturns()) != nReturns {
			t.Errorf("returns: got %d, want %d", len(result.GetFormalReturns()), nReturns)
		}
	})
}
