package goivy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/xtracer"
)

func testCtx() *UpdateContext {
	return &UpdateContext{
		Domain: New(),
		PVars:  nil,
		ActCfg: NewActionsConfig(),
	}
}

func captureActionUpdateStdout(t *testing.T, fn func()) string {
	t.Helper()
	if !xtracer.Enabled {
		t.Skip("xtrace disabled at build time")
	}

	oldStdout := os.Stdout
	oldSuppressed := xtracer.Suppressed
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outCh <- buf.String()
	}()

	os.Stdout = w
	xtracer.Suppressed = false
	closedW := false
	defer func() {
		os.Stdout = oldStdout
		xtracer.Suppressed = oldSuppressed
		if !closedW {
			_ = w.Close()
		}
		_ = r.Close()
	}()
	fn()
	_ = w.Close()
	closedW = true
	out := <-outCh
	return out
}

// --- AssumeAction ---

func TestAssumeActionUpdate(t *testing.T) {
	fmla := NewConst("p", Boolean)
	a := NewAssumeAction(fmla)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("AssumeAction should modify nothing, got %v", u.Modified)
	}
	if u.TR == nil || u.TR.IsFalse() {
		t.Error("AssumeAction TR should be the formula, not false")
	}
	if !u.Pre.IsFalse() {
		t.Error("AssumeAction Pre should be false (never fails)")
	}
}

func TestAssumeActionUpdateTrue(t *testing.T) {
	a := NewAssumeAction(True)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if !u.TR.IsTrue() {
		t.Errorf("AssumeAction(true) TR should be true, got %s", u.TR)
	}
}

func TestSubgoalActionIntUpdateMatchesAssertActionLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_module as im, logic as lg

im.module = im.Module()
updated, tr, pre = act.SubgoalAction(lg.false).int_update(im.module, {})
print(json.dumps({
    "tr_fmlas": len(tr.fmlas),
    "pre_fmlas": len(pre.fmlas),
    "pre_is_false": pre.is_false(),
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python SubgoalAction int_update oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		TRFmlas    int  `json:"tr_fmlas"`
		PreFmlas   int  `json:"pre_fmlas"`
		PreIsFalse bool `json:"pre_is_false"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python SubgoalAction int_update oracle %q: %v", out, err)
	}

	u := IntUpdate(NewSubgoalAction(False), testCtx())
	if len(u.TR.Fmlas) != want.TRFmlas || len(u.Pre.Fmlas) != want.PreFmlas || u.Pre.IsFalse() != want.PreIsFalse {
		t.Fatalf("Go SubgoalAction IntUpdate differs from Python AssertAction inheritance\nwant TR fmlas=%d Pre fmlas=%d PreIsFalse=%v\ngot  TR fmlas=%d Pre fmlas=%d PreIsFalse=%v",
			want.TRFmlas, want.PreFmlas, want.PreIsFalse,
			len(u.TR.Fmlas), len(u.Pre.Fmlas), u.Pre.IsFalse())
	}
}

func TestAssignFieldActionRejectsNonBinaryRelationLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act
from ivy import ivy_logic as lg
from ivy import ivy_module as im

mod = im.Module()
sort_s = lg.UninterpretedSort('S')
obj = lg.Symbol('o', sort_s)
val = lg.Symbol('v', sort_s)
field = lg.Symbol('bad', lg.RelationSort([sort_s]))
action = act.AssignFieldAction(obj, field, val)

try:
    action.action_update(mod, {})
    res = {"errored": False, "type": "", "message": ""}
except Exception as err:
    res = {"errored": True, "type": type(err).__name__, "message": str(err)}

print(json.dumps(res, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python AssignFieldAction oracle failed: %v\n%s", err, out)
	}
	var want struct {
		Errored bool   `json:"errored"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python AssignFieldAction oracle %q: %v", out, err)
	}
	if !want.Errored || want.Type != "IvyError" || !strings.Contains(want.Message, "field bad must be a binary relation") {
		t.Fatalf("python AssignFieldAction oracle changed: %+v", want)
	}

	sortS := actionsMkSort("S")
	field := NewConst("bad", LogicRelationSort([]Sort{sortS}))
	obj := NewConst("o", sortS)
	val := NewConst("v", sortS)
	action := NewAssignFieldAction(field, obj, val)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("AssignFieldAction accepted a unary relation field; Python raised IvyError")
		}
		if !strings.Contains(fmt.Sprint(r), "field bad must be a binary relation") {
			t.Fatalf("wrong panic for invalid field update\nwant substring %q\ngot: %v", "field bad must be a binary relation", r)
		}
	}()
	_ = action.ActionUpdate(testCtx())
}

// --- AssertAction ---

func TestAssertActionUpdate(t *testing.T) {
	fmla := NewConst("p", Boolean)
	a := NewAssertAction(fmla)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("AssertAction should modify nothing, got %v", u.Modified)
	}
	if !u.TR.IsTrue() {
		t.Error("AssertAction TR should be true")
	}
	// Pre should be the dual (negation) of p
	if u.Pre == nil || u.Pre.IsFalse() {
		t.Error("AssertAction Pre should be the negated formula")
	}
	// Pre is a Clauses with one formula: Not(p). Access it directly.
	if len(u.Pre.Fmlas) != 1 {
		t.Fatalf("AssertAction Pre should have 1 formula, got %d", len(u.Pre.Fmlas))
	}
	not, ok := u.Pre.Fmlas[0].(*LogicNot)
	if !ok {
		t.Fatalf("AssertAction Pre should contain Not(p), got %T: %s", u.Pre.Fmlas[0], u.Pre.Fmlas[0])
	}
	if c, ok := not.Body.(*Const); !ok || c.Name != "p" {
		t.Errorf("AssertAction Pre should contain Not(p), got %s", u.Pre.Fmlas[0])
	}
}

// --- AssignAction ---

func TestAssignActionRejectsUnboundRHSVariablesLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act
from ivy import ivy_logic as lg
from ivy import ivy_module as im

mod = im.Module()
sort_s = lg.UninterpretedSort('S')
x = lg.Variable('X', sort_s)
y = lg.Variable('Y', sort_s)
p = lg.Symbol('p', lg.RelationSort([sort_s]))
q = lg.Symbol('q', lg.RelationSort([sort_s]))
action = act.AssignAction(p(x), q(y))

try:
    action.action_update(mod, {})
    res = {"errored": False, "type": "", "message": ""}
except Exception as err:
    res = {"errored": True, "type": type(err).__name__, "message": str(err)}

print(json.dumps(res, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python AssignAction variable oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var want struct {
		Errored bool   `json:"errored"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &want); err != nil {
		t.Fatalf("decode python AssignAction variable oracle %q: %v", out, err)
	}
	if !want.Errored || want.Type != "IvyError" || !strings.Contains(want.Message, "multiply assigned: p") {
		t.Fatalf("python AssignAction variable oracle changed: %+v", want)
	}

	sortS := actionsMkSort("S")
	x, _ := NewVariable("X", sortS)
	y, _ := NewVariable("Y", sortS)
	p := NewConst("p", LogicRelationSort([]Sort{sortS}))
	q := NewConst("q", LogicRelationSort([]Sort{sortS}))
	lhs, _ := NewApply(p, x)
	rhs, _ := NewApply(q, y)
	action := NewAssignAction(lhs, rhs)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("AssignAction accepted RHS variable not bound by LHS; Python raised IvyError")
		}
		if !strings.Contains(fmt.Sprint(r), "multiply assigned: p") {
			t.Fatalf("wrong panic for unbound RHS variable\nwant substring %q\ngot: %v", "multiply assigned: p", r)
		}
	}()
	_ = action.ActionUpdate(testCtx())
}

func TestAssignActionRejectsSortMismatchLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act
from ivy import ivy_logic as lg
from ivy import ivy_module as im

mod = im.Module()
sort_s = lg.UninterpretedSort('S')
x = lg.Symbol('x', sort_s)
action = act.AssignAction(x, lg.And())

try:
    action.action_update(mod, {})
    res = {"errored": False, "type": "", "message": ""}
except Exception as err:
    res = {"errored": True, "type": type(err).__name__, "message": str(err)}

print(json.dumps(res, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python AssignAction sort oracle failed: %v\n%s", err, out)
	}
	var want struct {
		Errored bool   `json:"errored"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python AssignAction sort oracle %q: %v", out, err)
	}
	if !want.Errored || want.Type != "IvyError" || !strings.Contains(want.Message, "sort mismatch in assignment to x") {
		t.Fatalf("python AssignAction sort oracle changed: %+v", want)
	}

	sortS := actionsMkSort("S")
	x := NewConst("x", sortS)
	action := NewAssignAction(x, True)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("AssignAction accepted individual/boolean sort mismatch; Python raised IvyError")
		}
		if !strings.Contains(fmt.Sprint(r), "sort mismatch in assignment to x") {
			t.Fatalf("wrong panic for assignment sort mismatch\nwant substring %q\ngot: %v", "sort mismatch in assignment to x", r)
		}
	}()
	_ = action.ActionUpdate(testCtx())
}

func TestAssignActionHierarchyUsesPythonRuntimeKeySemantics(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act
from ivy import ivy_logic as lg
from ivy import ivy_module as im

mod = im.Module()
im.module = mod
sort_s = lg.UninterpretedSort('S')
X = lg.Variable('X', sort_s)
p = lg.Symbol('p', lg.RelationSort([sort_s]))
q = lg.Symbol('q', lg.RelationSort([sort_s]))
mod.add_to_hierarchy('p.a')
mod.add_to_hierarchy('p.b')
updated, clauses, pre = act.AssignAction(p(X), q(X)).action_update(mod, {})
text = str(clauses)
print(json.dumps({
    "mods": [s.name for s in updated],
    "mentions_child": ("p.a" in text) or ("p.b" in text) or ("new_p.a" in text) or ("new_p.b" in text),
    "mentions_parent_new": "new_p" in text,
    "mentions_rhs": "q" in text,
}, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python hierarchical AssignAction oracle failed: %v\n%s", err, out)
	}
	var want struct {
		Mods              []string `json:"mods"`
		MentionsChild     bool     `json:"mentions_child"`
		MentionsParentNew bool     `json:"mentions_parent_new"`
		MentionsRHS       bool     `json:"mentions_rhs"`
	}
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python hierarchical AssignAction oracle %q: %v", out, err)
	}
	if fmt.Sprint(want.Mods) != "[p]" || want.MentionsChild || !want.MentionsParentNew || !want.MentionsRHS {
		t.Fatalf("python hierarchical AssignAction oracle changed: %+v", want)
	}

	sortS := actionsMkSort("S")
	x, _ := NewVariable("X", sortS)
	p := NewConst("p", LogicRelationSort([]Sort{sortS}))
	q := NewConst("q", LogicRelationSort([]Sort{sortS}))
	mod := New()
	mod.AddToHierarchy("p.a")
	mod.AddToHierarchy("p.b")
	u := NewAssignAction(MustApply(p, x), MustApply(q, x)).ActionUpdate(&UpdateContext{
		Domain: mod,
		PVars:  nil,
		ActCfg: NewActionsConfig(),
	})

	var gotMods []string
	for _, m := range u.Modified {
		gotMods = append(gotMods, m.Name)
	}
	gotText := u.TR.String()
	gotMentionsChild := strings.Contains(gotText, "p.a") || strings.Contains(gotText, "p.b") ||
		strings.Contains(gotText, "new_p.a") || strings.Contains(gotText, "new_p.b")
	gotMentionsParentNew := strings.Contains(gotText, "new_p")
	gotMentionsRHS := strings.Contains(gotText, "q")

	if fmt.Sprint(gotMods) != fmt.Sprint(want.Mods) ||
		gotMentionsChild != want.MentionsChild ||
		gotMentionsParentNew != want.MentionsParentNew ||
		gotMentionsRHS != want.MentionsRHS {
		t.Fatalf("Go hierarchical AssignAction differs from Python runtime behavior\nwant mods=%v child=%v parentNew=%v rhs=%v\ngot  mods=%v child=%v parentNew=%v rhs=%v\nTR: %s",
			want.Mods, want.MentionsChild, want.MentionsParentNew, want.MentionsRHS,
			gotMods, gotMentionsChild, gotMentionsParentNew, gotMentionsRHS, gotText)
	}
}

func TestAssignActionSimple(t *testing.T) {
	// x := y where both are constants with TopS sort
	x := NewConst("x", TopS)
	y := NewConst("y", TopS)
	a := NewAssignAction(x, y)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 1 || u.Modified[0].Name != "x" {
		t.Errorf("AssignAction should modify [x], got %v", u.Modified)
	}
	// TR should contain an equivalence new_x = y
	trStr := u.TR.String()
	if !strings.Contains(trStr, "new_x") {
		t.Errorf("AssignAction TR should reference new_x, got %s", trStr)
	}
	if !u.Pre.IsFalse() {
		t.Error("AssignAction Pre should be false")
	}
}

func TestAssignActionWithArgs(t *testing.T) {
	// f(a) := b where f is a function constant
	fSort, _ := NewFunctionSort(TopS, TopS)
	f := NewConst("f", fSort)
	aConst := NewConst("a", TopS)
	bConst := NewConst("b", TopS)
	lhs := MustApply(f, aConst)
	a := NewAssignAction(lhs, bConst)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 1 || u.Modified[0].Name != "f" {
		t.Errorf("should modify [f], got %v", u.Modified)
	}
	// TR should have an ITE: at index a, use b; elsewhere keep f
	trStr := u.TR.String()
	if !strings.Contains(trStr, "new_f") {
		t.Errorf("TR should reference new_f, got %s", trStr)
	}
}

// --- HavocAction ---

func TestHavocActionUpdate(t *testing.T) {
	x := NewConst("x", TopS)
	a := NewHavocAction(x)
	ctx := testCtx()
	u := a.ActionUpdate(ctx)
	if len(u.Modified) != 1 || u.Modified[0].Name != "x" {
		t.Errorf("HavocAction should modify [x], got %v", u.Modified)
	}
	if !u.Pre.IsFalse() {
		t.Error("HavocAction Pre should be false")
	}
}

// --- NativeAction ---

func TestNativeActionUpdate(t *testing.T) {
	code := NewConst("code", TopS)
	a := NewNativeAction(code)
	ctx := testCtx()
	u := a.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("NativeAction should modify nothing, got %v", u.Modified)
	}
	if !u.TR.IsTrue() {
		t.Error("NativeAction TR should be true")
	}
}

// --- Sequence ---

func TestSequenceIntUpdate(t *testing.T) {
	// Sequence of two assumes: assume p; assume q
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	seq := NewSequence(NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	u := seq.IntUpdate(ctx)
	// Should modify nothing
	if len(u.Modified) != 0 {
		t.Errorf("Sequence of assumes should modify nothing, got %v", u.Modified)
	}
	// TR should contain both p and q conjoined
	if u.TR.IsTrue() {
		t.Error("Sequence TR should not be just true")
	}
}

func TestSequenceAssignAssume(t *testing.T) {
	// x := y; assume p — should modify [x]
	x := NewConst("x", TopS)
	y := NewConst("y", TopS)
	p := NewConst("p", Boolean)
	seq := NewSequence(NewAssignAction(x, y), NewAssumeAction(p))
	ctx := testCtx()
	u := seq.IntUpdate(ctx)
	hasX := false
	for _, m := range u.Modified {
		if m.Name == "x" {
			hasX = true
		}
	}
	if !hasX {
		t.Errorf("Sequence should modify x, got %v", u.Modified)
	}
}

// --- ChoiceAction ---

func TestChoiceActionIntUpdate(t *testing.T) {
	// choice { assume p } or { assume q }
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	ch := NewChoiceActionOn(NewActionsConfig(), NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	u := ch.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("Choice of assumes should modify nothing, got %v", u.Modified)
	}
}

// --- IfAction ---

func TestIfActionIntUpdate(t *testing.T) {
	cond := NewConst("c", Boolean)
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	ifAct := NewIfAction(cond, NewAssumeAction(p), NewAssumeAction(q))
	ctx := testCtx()
	u := ifAct.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("If of assumes should modify nothing, got %v", u.Modified)
	}
}

// --- LocalAction ---

func TestLocalActionIntUpdate(t *testing.T) {
	x := NewConst("x", TopS)
	y := NewConst("y", TopS)
	// local x { x := y }
	asgn := NewAssignAction(x, y)
	local := NewLocalActionOn(NewActionsConfig(), "test", x, asgn)
	ctx := testCtx()
	u := local.IntUpdate(ctx)
	// x should be hidden — not in modified list
	for _, m := range u.Modified {
		if m.Name == "x" {
			t.Error("Local should hide x from modified list")
		}
	}
}

// --- BindOldsAction ---

func TestBindOldsActionIntUpdate(t *testing.T) {
	p := NewConst("p", Boolean)
	inner := NewAssumeAction(p)
	bindOlds := NewBindOldsAction(inner)
	ctx := testCtx()
	u := bindOlds.IntUpdate(ctx)
	if len(u.Modified) != 0 {
		t.Errorf("BindOlds of assume should modify nothing, got %v", u.Modified)
	}
}

// --- CallAction ---

func TestCallActionIntUpdateSubstitutesBeforeExitTrace(t *testing.T) {
	cfg := NewActionsConfig()
	mod := New()
	mod.Actions.Set("callee", NewSequence())
	call := NewCallActionOn(cfg, NewConst("callee", TopS))
	ctx := &UpdateContext{Domain: mod, ActCfg: cfg}

	var update *Update
	out := captureActionUpdateStdout(t, func() {
		update = call.IntUpdate(ctx)
	})

	if update == nil {
		t.Fatal("CallAction.IntUpdate returned nil")
	}
	if len(update.Modified) != 0 {
		t.Fatalf("empty callee should modify nothing, got %v", update.Modified)
	}
	if !update.TR.IsTrue() {
		t.Fatalf("empty callee TR should be true, got %s", update.TR.Canon())
	}
	if !update.Pre.IsFalse() {
		t.Fatalf("empty callee Pre should be false, got %s", update.Pre.Canon())
	}

	subIdx := strings.Index(out, "actions.substitute_constants_action ENTER type=Sequence nargs=0")
	if subIdx < 0 {
		t.Fatalf("CallAction.IntUpdate did not substitute the callee action before returning; output:\n%s", out)
	}
	exitIdx := strings.Index(out, "actions.CallAction.int_update EXIT")
	if exitIdx < 0 {
		t.Fatalf("CallAction.IntUpdate did not emit EXIT trace; output:\n%s", out)
	}
	if subIdx > exitIdx {
		t.Fatalf("substitution trace must precede CallAction EXIT; output:\n%s", out)
	}
	wrapperSeqIdx := strings.Index(out[subIdx:], "actions.Sequence.int_update ENTER")
	if wrapperSeqIdx < 0 {
		t.Fatalf("CallAction.IntUpdate must recurse into the Python wrapper Sequence after substitution; output:\n%s", out)
	}
	if subIdx+wrapperSeqIdx > exitIdx {
		t.Fatalf("wrapper Sequence trace must precede CallAction EXIT; output:\n%s", out)
	}
}

func TestCallActionIntUpdateActualsAndHideFormals(t *testing.T) {
	cfg := NewActionsConfig()
	mod := New()
	sortT := &UninterpretedSort{Name: "t"}

	formalIn := NewConst("p", sortT)
	formalOut := NewConst("r", sortT)
	actualIn := NewConst("p", sortT) // Same name forces capture-avoiding formal rename.
	actualOut := NewConst("out", sortT)
	state := NewConst("state", sortT)

	callee := NewSequence(NewAssignAction(state, formalIn))
	callee.SetFormalParams([]*Const{formalIn})
	callee.SetFormalReturns([]*Const{formalOut})
	mod.Actions.Set("callee", callee)

	call := NewCallActionOn(cfg, MustApply(NewConst("callee", TopS), actualIn), actualOut)
	ctx := &UpdateContext{Domain: mod, ActCfg: cfg}
	update := call.IntUpdate(ctx)
	if update == nil {
		t.Fatal("CallAction.IntUpdate returned nil")
	}

	modified := make(map[string]bool)
	for _, sym := range update.Modified {
		modified[sym.Name] = true
	}
	for _, name := range []string{"state", "out"} {
		if !modified[name] {
			t.Fatalf("CallAction.IntUpdate modified=%v, want %s after body/output assignments", modified, name)
		}
	}
	for _, hidden := range []string{"p", "p_a", "r"} {
		if modified[hidden] {
			t.Fatalf("CallAction.IntUpdate leaked hidden formal %q in modified=%v", hidden, modified)
		}
	}
}

func TestDistinctObjRenamingDuplicateStructuralFormalLastWinsLikePython(t *testing.T) {
	arrSort := &UninterpretedSort{Name: "arr"}
	tSort := &UninterpretedSort{Name: "t"}
	firstA := NewConst("fml:a", arrSort)
	v := NewConst("fml:v", tSort)
	secondA := NewConst("fml:a", arrSort)

	renaming := distinctObjRenaming([]*Const{firstA, v, secondA}, map[string]bool{})

	if len(renaming) != 2 {
		t.Fatalf("duplicate structural formals should collapse to two keys, got %d", len(renaming))
	}
	gotA := renaming[Key(firstA)]
	if gotA == nil {
		t.Fatal("missing renaming for fml:a")
	}
	if gotA.Name != "fml:a_a" {
		t.Fatalf("duplicate formal should keep Python's later dict value fml:a_a, got %q", gotA.Name)
	}
	gotV := renaming[Key(v)]
	if gotV == nil || gotV.Name != "fml:v" {
		t.Fatalf("non-conflicting formal should keep identity name fml:v, got %#v", gotV)
	}
}

// --- GetUpdate ---

func TestGetUpdateHidesFormals(t *testing.T) {
	x := NewConst("fml:x", TopS)
	y := NewConst("y", TopS)
	asgn := NewAssignAction(x, y)
	asgn.SetFormalParams([]*Const{x})
	ctx := testCtx()
	u := GetUpdate(asgn, ctx)
	// x should be hidden from the modified list
	for _, m := range u.Modified {
		if m.Name == "fml:x" {
			t.Error("GetUpdate should hide formal params from modified")
		}
	}
}

// --- hideFormals __prefix tests ---

// formulaContainsName checks if a formula references a constant with the given name.
func formulaContainsName(node Expr, name string) bool {
	if node == nil {
		return false
	}
	if c, ok := node.(*Const); ok {
		return c.Name == name
	}
	if app, ok := node.(*Apply); ok {
		if formulaContainsName(app.Func, name) {
			return true
		}
	}
	for _, ch := range node.Children() {
		if formulaContainsName(ch, name) {
			return true
		}
	}
	return false
}

// formulaContainsPrefix checks if any constant name in the formula starts with prefix.
func formulaContainsPrefix(node Expr, prefix string) bool {
	if node == nil {
		return false
	}
	if c, ok := node.(*Const); ok {
		return strings.HasPrefix(c.Name, prefix)
	}
	if app, ok := node.(*Apply); ok {
		if formulaContainsPrefix(app.Func, prefix) {
			return true
		}
	}
	for _, ch := range node.Children() {
		if formulaContainsPrefix(ch, prefix) {
			return true
		}
	}
	return false
}

// collectConstNames returns all constant names in a formula.
func collectConstNames(node Expr, out map[string]bool) {
	if node == nil {
		return
	}
	if c, ok := node.(*Const); ok {
		out[c.Name] = true
		return
	}
	if app, ok := node.(*Apply); ok {
		collectConstNames(app.Func, out)
	}
	for _, ch := range node.Children() {
		collectConstNames(ch, out)
	}
}

// TestHideFormalsRenamesWithDoubleUnderscore verifies that hideFormals
// (via Hide → ExistQuantClauses) renames formal params and their
// new_ versions with __ prefix. This is the core mechanism that should
// produce __new_loc:wr from new_loc:wr in the fragment checker.
func TestHideFormalsRenamesWithDoubleUnderscore(t *testing.T) {
	// Create symbols with a non-TopS sort (UninterpretedSort).
	// This is critical — the original bug was that TopS-keyed lookups
	// failed to match constants with real sorts.
	mySort := &UninterpretedSort{Name: "mytype"}
	loc := NewConst("loc", mySort)
	val := NewConst("val", mySort)

	// Create "loc := val" assignment — this produces new_loc in the TR
	asgn := NewAssignAction(loc, val)
	asgn.SetFormalParams([]*Const{loc})

	ctx := testCtx()
	// Step 1: IntUpdate — should produce new_loc in TR
	u := IntUpdate(asgn, ctx)
	trFormula := u.TRNode()
	t.Logf("After IntUpdate, TR formula: %v", trFormula)

	names := make(map[string]bool)
	collectConstNames(trFormula, names)
	t.Logf("After IntUpdate, constant names: %v", names)

	if !formulaContainsName(trFormula, "new_loc") {
		t.Logf("IntUpdate TR does not contain new_loc (may use different structure)")
	}

	// Step 2: BindOldsAction
	u = BindOldsUpdate(u)

	// Step 3: hideFormals — should rename loc and new_loc with __ prefix
	u = hideFormals(asgn, u)
	trAfterHide := u.TRNode()

	names2 := make(map[string]bool)
	collectConstNames(trAfterHide, names2)
	t.Logf("After hideFormals, constant names: %v", names2)

	// The bare "loc" should NOT appear (it was hidden)
	if formulaContainsName(trAfterHide, "loc") {
		t.Error("After hideFormals, bare 'loc' should be renamed with __ prefix")
	}

	// The bare "new_loc" should NOT appear (it was hidden)
	if formulaContainsName(trAfterHide, "new_loc") {
		t.Error("After hideFormals, bare 'new_loc' should be renamed with __ prefix")
	}

	// There should be some __-prefixed name
	if !formulaContainsPrefix(trAfterHide, "__") {
		t.Error("After hideFormals, expected at least one __-prefixed constant")
	}
}

// TestHideFormalsWithTopSSort verifies that Hide works even when the
// formal param has TopS sort (the original code's assumption).
func TestHideFormalsWithTopSSort(t *testing.T) {
	loc := NewConst("loc", TopS)
	x := NewConst("x", TopS)
	asgn := NewAssignAction(loc, x)
	asgn.SetFormalParams([]*Const{loc})

	ctx := testCtx()
	u := IntUpdate(asgn, ctx)
	u = BindOldsUpdate(u)
	u = hideFormals(asgn, u)
	trAfterHide := u.TRNode()

	names := make(map[string]bool)
	collectConstNames(trAfterHide, names)
	t.Logf("After hideFormals (TopS), constant names: %v", names)

	if formulaContainsName(trAfterHide, "loc") {
		t.Error("After hideFormals, bare 'loc' should be renamed")
	}
	if formulaContainsName(trAfterHide, "new_loc") {
		t.Error("After hideFormals, bare 'new_loc' should be renamed")
	}
}

// TestTransrelHideDirectly tests Hide directly with a
// FunctionSort symbol to verify ExistQuantClauses works with real sorts.
func TestTransrelHideDirectly(t *testing.T) {
	mySort := &UninterpretedSort{Name: "mytype"}
	loc := NewConst("loc", mySort)
	newLoc := NewConst("new_loc", mySort)

	// Build a simple update with loc in Modified and new_loc in TR
	eq, _ := NewEq(newLoc, loc)
	tr := &LogicAnd{Terms: []Expr{eq}}

	u := &Update{
		Modified: []*Const{loc},
		TR:       FormulaToClauses(tr, nil),
		Pre:      FormulaToClauses(False, nil),
	}

	// Hide loc — should also hide new_loc, both renamed with __
	hidden := Hide([]*Const{loc}, u)
	trFormula := hidden.TRNode()

	names := make(map[string]bool)
	collectConstNames(trFormula, names)
	t.Logf("After Hide (FunctionSort), constant names: %v", names)

	if formulaContainsName(trFormula, "loc") {
		t.Errorf("After Hide, bare 'loc' should be renamed to __loc, got names: %v", names)
	}
	if formulaContainsName(trFormula, "new_loc") {
		t.Errorf("After Hide, bare 'new_loc' should be renamed to __new_loc, got names: %v", names)
	}
	if !formulaContainsPrefix(trFormula, "__") {
		t.Errorf("After Hide, expected __-prefixed constants, got names: %v", names)
	}
}

// --- NullUpdate ---

func TestNullUpdate(t *testing.T) {
	u := NullUpdate()
	if u.Modified == nil {
		t.Error("NullUpdate Modified should be non-nil empty slice")
	}
	if len(u.Modified) != 0 {
		t.Errorf("NullUpdate Modified should be empty, got %v", u.Modified)
	}
}

// --- Helper functions ---

func TestEquivASTIndividual(t *testing.T) {
	x := NewConst("x", TopS)
	y := NewConst("y", TopS)
	eq := equivAST(x, y)
	if _, ok := eq.(*Eq); !ok {
		t.Errorf("equivAST for individuals should return Eq, got %T", eq)
	}
}

func TestEquivASTBoolean(t *testing.T) {
	p := NewConst("p", Boolean)
	q := NewConst("q", Boolean)
	result := equivAST(p, q)
	if _, ok := result.(*LogicAnd); !ok {
		t.Errorf("equivAST for booleans should return And, got %T: %s", result, result)
	}
}

func TestDualFormula(t *testing.T) {
	p := NewConst("p", Boolean)
	dual := DualFormula(p, nil, nil)
	not, ok := dual.(*LogicNot)
	if !ok {
		t.Fatalf("DualFormula of constant should be Not, got %T: %s", dual, dual)
	}
	if c, ok := not.Body.(*Const); !ok || c.Name != "p" {
		t.Errorf("DualFormula body should be p, got %s", not.Body)
	}
}

func TestSkolemizeFormula(t *testing.T) {
	v, _ := NewVariable("X", TopS)
	body := v
	ex, _ := NewExists([]*LogicVariable{v}, body)
	result := SkolemizeFormula(ex, nil, nil)
	// Should replace X with __sk__X
	if c, ok := result.(*Const); ok {
		if c.Name != "__sk__X" {
			t.Errorf("expected __sk__X, got %s", c.Name)
		}
	} else {
		t.Errorf("expected Const after skolemization, got %T: %s", result, result)
	}
}

func TestConjoin(t *testing.T) {
	if !actionsUpdateIsTrue(conjoin(True, True)) {
		t.Error("conjoin(true, true) should be true")
	}
	p := NewConst("p", Boolean)
	result := conjoin(True, p)
	if result != p {
		t.Errorf("conjoin(true, p) should be p, got %s", result)
	}
	if !actionsUpdateIsFalse(conjoin(False, p)) {
		t.Error("conjoin(false, p) should be false")
	}
}

func TestDisjoin(t *testing.T) {
	if !actionsUpdateIsFalse(disjoin(False, False)) {
		t.Error("disjoin(false, false) should be false")
	}
	p := NewConst("p", Boolean)
	result := disjoin(False, p)
	if result != p {
		t.Errorf("disjoin(false, p) should be p, got %s", result)
	}
	if !actionsUpdateIsTrue(disjoin(True, p)) {
		t.Error("disjoin(true, p) should be true")
	}
}
