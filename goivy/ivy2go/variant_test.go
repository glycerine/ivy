package ivy2go

import (
	"os/exec"
	"strings"
	"testing"
)

// --- OPEN 063: variant constructors + *> downcast tests --------------

func TestEmitVariant_SuperStructAndConstructor(t *testing.T) {
	mod := compileIvySource(t, `
type t0
variant t1 of t0
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if text == "" {
		t.Fatalf("types.go missing:\n%v", keysOf(out.Files))
	}
	// Super struct with Tag field.
	if !strings.Contains(text, "type T0 struct") {
		t.Errorf("super struct missing, got:\n%s", text)
	}
	if !strings.Contains(text, "Tag int") {
		t.Errorf("Tag field missing, got:\n%s", text)
	}
	// Plain-leaf constructor (no payload arg).
	if !strings.Contains(text, "func NewT0T1() T0 { return T0{Tag: 0} }") {
		t.Errorf("plain-leaf constructor missing, got:\n%s", text)
	}
}

func TestEmitVariant_TwoLeafConstructors(t *testing.T) {
	mod := compileIvySource(t, `
type super
variant leaf_a of super
variant leaf_b of super
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["types.go"]
	if !strings.Contains(text, "NewSuperLeafA") {
		t.Errorf("missing NewSuperLeafA constructor, got:\n%s", text)
	}
	if !strings.Contains(text, "NewSuperLeafB") {
		t.Errorf("missing NewSuperLeafB constructor, got:\n%s", text)
	}
}

func TestEmitActions_PlainVariantLeafUpcastUsesZeroArgConstructor(t *testing.T) {
	mod := compileIvySource(t, `
type msg
variant ack of msg
individual saved : msg
action save_ack = {
	var a : ack;
	saved := a
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	requireHasLineWithAllTerms(t, actions, "s.Saved", "=", "NewMsgAck()")
}

func TestEmitActions_PayloadVariantLeafUpcastKeepsPayloadArgument(t *testing.T) {
	mod := compileIvySource(t, `
type msg
variant req of msg = struct {
	ok : bool
}
individual saved : msg
action save_req = {
	var r : req;
	r.ok := true;
	saved := r
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	actions := out.Files["actions.go"]
	requireHasLineWithAllTerms(t, actions, "s.Saved", "=", "NewMsgReq(loc__r)")
}

func TestEmitActions_NondetVariantSuperUsesLeafConstructorArity(t *testing.T) {
	g := newExprGen(t, `
type msg
variant req of msg = struct {
	ok : bool
}
variant ack of msg
`)
	msg := mustSort(t, g.Mod, "msg")
	g.mkNondetVariantScoped(&g.actions, "out", msg, "choose", 7, "")
	actions := g.Ctx.Actions.GetFile()
	requireHasLineWithAllTerms(t, actions, "out", "=", "NewMsgReq(__nd_v_7_0)")
	requireHasLineWithAllTerms(t, actions, "out", "=", "NewMsgAck()")
}

func TestSmoke_BuildEmittedPlainVariantLeafUpcastAndNondet(t *testing.T) {
	if !SlowGoTest {
		t.Skip("SLOW_GO_TEST not set")
	}
	mod := compileIvySource(t, `
type msg
variant req of msg = struct {
	ok : bool
}
variant ack of msg
individual saved : msg
action save_ack = {
	var a : ack;
	saved := a
}
action choose returns(out: msg) = {
	var m : msg;
	out := m
}
`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "main"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dir := playpenDir(t)
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = pkgDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed:\n%s", string(output))
	}
}
