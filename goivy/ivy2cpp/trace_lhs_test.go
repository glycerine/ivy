package ivy2cpp

import (
	"strings"
	"testing"
)

func TestTraceLhsRespectsHexNumberFormat(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step(c:color) = {
    saved := c
}
export step
`)
	mod.Attributes["radix"] = "16"
	out, err := Generate(mod, Config{Target: "repl", ClassName: "tracelhshex", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := `__ivy_out << std::hex << std::showbase << "  write(" << "saved" << "," << (c) << ")" << std::endl;`
	if !strings.Contains(out.Impl, want) {
		t.Fatalf("hex write trace missing %q:\n%s", want, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestTraceLhsFunctionApplicationArgs(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {i0, i1}
type color = {red, green}
function slot(I:idx) : color
action step(i:idx,c:color) = {
    slot(i) := c
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "tracefunc", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := `__ivy_out << "  write(" << "slot" << "(" << i << ")" << "," << (c) << ")" << std::endl;`
	if !strings.Contains(out.Impl, want) {
		t.Fatalf("function-application trace missing %q:\n%s", want, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestTraceLhsDecomposesDestructorRecord(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
individual current : cell
action step(c:color) = {
    current.shade := c
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "tracefield", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := `__ivy_out << "  write(" << "current" << ".shade" << "," << (c) << ")" << std::endl;`
	if !strings.Contains(out.Impl, want) {
		t.Fatalf("destructor-field trace missing %q:\n%s", want, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestTraceLhsNestedDestructorChain(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type inner
type outer
destructor child(O:outer) : inner
destructor shade(I:inner) : color
individual root : outer
action step(c:color) = {
    root.child.shade := c
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "tracenestedfield", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := `__ivy_out << "  write(" << "root" << ".child" << ".shade" << "," << (c) << ")" << std::endl;`
	if !strings.Contains(out.Impl, want) {
		t.Fatalf("nested destructor trace missing %q:\n%s", want, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestTraceLhsNamespacedNameSuppressed(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
    var tmp : color := green;
    tmp := red;
    saved := tmp
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "tracelocal", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.Contains(out.Impl, `write(" << "loc:tmp"`) || strings.Contains(out.Impl, `write(" << "loc__tmp"`) {
		t.Fatalf("namespaced local assignment should not emit write trace:\n%s", out.Impl)
	}
	if !strings.Contains(out.Impl, `<< "saved" << "," << (loc__tmp)`) {
		t.Fatalf("control saved assignment trace missing:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestTraceLhsDoesNotEmitWhenTraceOff(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step(c:color) = {
    saved := c
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "traceoff"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.Contains(out.Impl, `__ivy_out << "  write(`) {
		t.Fatalf("write trace emitted with Trace=false:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}
