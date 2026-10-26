package abc

// Go-only validation for the structural Verilog parser.

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func verilogConnectionsEqual(left, right []VerilogConnection) bool {
	return slices.EqualFunc(left, right, func(a, b VerilogConnection) bool {
		return a.Formal == b.Formal && slices.Equal(a.Actual, b.Actual)
	})
}

func TestVerilogStreamAccessOwnershipAndLifecycle(t *testing.T) {
	input := []byte("a\n bc,def")
	stream, err := NewVerilogStream("memory.v", input, VerilogLimits{MaxBytes: 64, MaxWordBytes: 8})
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 'z'
	if name, _ := stream.FileName(); name != "memory.v" {
		t.Fatalf("name %q", name)
	}
	if size, _ := stream.FileSize(); size != 9 {
		t.Fatalf("size %d", size)
	}
	if got, _ := stream.Scan(); got != 'a' {
		t.Fatalf("ownership got %q", got)
	}
	if got, _ := stream.Pop(); got != 'a' {
		t.Fatalf("pop got %q", got)
	}
	if err = stream.SkipChars("\n "); err != nil {
		t.Fatal(err)
	}
	if line, _ := stream.Line(); line != 2 {
		t.Fatalf("line %d", line)
	}
	if position, _ := stream.Position(); position != 3 {
		t.Fatalf("position %d", position)
	}
	word, err := stream.Word(",")
	if err != nil || word != "bc" {
		t.Fatalf("word=%q err=%v", word, err)
	}
	if err = stream.SkipToChars("d"); err != nil {
		t.Fatal(err)
	}
	if got, _ := stream.Scan(); got != 'd' {
		t.Fatalf("scan got %q", got)
	}
	if err = stream.Close(); err != nil {
		t.Fatal(err)
	}
	if stream.OK() {
		t.Fatal("closed stream is OK")
	}
	if _, err = stream.Scan(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed error %v", err)
	}
	if err = stream.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestVerilogFormulaPrecedenceMuxReductionAndNames(t *testing.T) {
	expr, names, err := ParseVerilogFormula("~a & b | c ^ (d ? e : f)", VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(names, []string{"a", "b", "c", "d", "e", "f"}) {
		t.Fatalf("names %v", names)
	}
	values := map[string]bool{"a": false, "b": true, "c": true, "d": true, "e": false, "f": true}
	got, err := expr.Eval(values)
	want := ((!values["a"]) && values["b"]) || (values["c"] != values["e"])
	if err != nil || got != want {
		t.Fatalf("eval got=%v want=%v err=%v", got, want, err)
	}
	clone := expr.Clone()
	clone.Args[0] = &VerilogExpr{Kind: VerilogExprConst}
	gotOriginal, _ := expr.Eval(values)
	if gotOriginal != want {
		t.Fatal("clone mutated source")
	}

	reduction, reducedNames, err := ParseVerilogReduction("~&{a,b,c}", VerilogLimits{})
	if err != nil || !slices.Equal(reducedNames, []string{"a", "b", "c"}) {
		t.Fatalf("reduction names=%v err=%v", reducedNames, err)
	}
	value, err := reduction.Eval(map[string]bool{"a": true, "b": true, "c": true})
	if err != nil || value {
		t.Fatalf("reduction value=%v err=%v", value, err)
	}
	escaped, escapedNames, err := ParseVerilogFormula("\\a.b  ^ x[3]", VerilogLimits{})
	if err != nil || escaped == nil || !slices.Equal(escapedNames, []string{"a.b", "x[3]"}) {
		t.Fatalf("escaped names=%v err=%v", escapedNames, err)
	}
}

func TestStructuralVerilogCompleteDesign(t *testing.T) {
	source := `
// child precedes its use
module child(a,b,y);
 input a,b;
 output y;
 assign y = a ^ b;
endmodule

module top(a,b,clk,y,q,bus);
 input a,b,clk;
 output y,q;
 output [1:0] bus;
 reg q;
 wire n1,n2,n3;
 and g0(n1,a,b);
 not (n2,n1);
 assign y = n2 ? a : b;
 assign bus[0] = a;
 assign bus[1] = b;
 always @(posedge clk) begin
   q <= n1;
 end
 initial q = 1'b0;
 child u0(.a(a), .b(b), .y(n3));
endmodule
`
	design, err := ParseStructuralVerilog("complete.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(design.Modules) != 2 {
		t.Fatalf("modules %d", len(design.Modules))
	}
	top, ok := design.Module("top")
	if !ok {
		t.Fatal("missing top")
	}
	if len(top.Gates) != 2 || len(top.Assignments) != 3 || len(top.Latches) != 1 || len(top.Instances) != 1 {
		t.Fatalf("counts gates=%d assigns=%d latches=%d instances=%d", len(top.Gates), len(top.Assignments), len(top.Latches), len(top.Instances))
	}
	if top.Latches[0].Clock != "clk" || top.Latches[0].Initial != 0 {
		t.Fatalf("latch %+v", top.Latches[0])
	}
	if _, ok = top.Nets["bus[0]"]; !ok {
		t.Fatal("missing vector bit 0")
	}
	if _, ok = top.Nets["bus[1]"]; !ok {
		t.Fatal("missing vector bit 1")
	}
	if top.Nets["y"].Driver != "assign" || top.Nets["n2"].Driver != "not" {
		t.Fatalf("drivers y=%q n2=%q", top.Nets["y"].Driver, top.Nets["n2"].Driver)
	}
	if !verilogConnectionsEqual(top.Instances[0].Connections, []VerilogConnection{{Formal: "a", Actual: []string{"a"}}, {Formal: "b", Actual: []string{"b"}}, {Formal: "y", Actual: []string{"n3"}}}) {
		t.Fatalf("connections %+v", top.Instances[0].Connections)
	}
	if top.Nets["n3"].Driver != "instance u0" {
		t.Fatalf("instance driver %q", top.Nets["n3"].Driver)
	}
	clone := design.Clone()
	clone.Modules[1].Nets["a"].Driver = "changed"
	clone.Modules[1].Gates[0].Inputs[0] = "changed"
	clone.Modules[1].Instances[0].Connections[0].Actual[0] = "changed"
	if top.Nets["a"].Driver == "changed" || top.Gates[0].Inputs[0] == "changed" || top.Instances[0].Connections[0].Actual[0] == "changed" {
		t.Fatal("clone aliases design")
	}
}

func TestStructuralVerilogPositionalBlackboxAndConstants(t *testing.T) {
	source := `module top(a,y,bus); input a; output y; output [2:0] bus; assign bus = 3'b10x; assign #1 y = a; mystery box0(a,bus[2:1],y); endmodule`
	design, err := ParseStructuralVerilog("blackbox.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	module, _ := design.Module("top")
	if len(module.Instances) != 1 || module.Instances[0].Module != "mystery" || len(module.Instances[0].Connections) != 3 || !slices.Equal(module.Instances[0].Connections[1].Actual, []string{"bus[2]", "bus[1]"}) {
		t.Fatalf("instances %+v", module.Instances)
	}
	if len(module.Assignments) != 4 {
		t.Fatalf("vector assignments %+v", module.Assignments)
	}
	gotBits := make([]bool, 3)
	for i := 0; i < 3; i++ {
		gotBits[i] = module.Assignments[i].Expression.Value
	}
	if !slices.Equal(gotBits, []bool{false, false, true}) {
		t.Fatalf("vector values %v", gotBits)
	}
	bits, err := ParseVerilogBinaryConstant("5'b10x01", 5)
	if err != nil || !slices.Equal(bits, []bool{true, false, false, false, true}) {
		t.Fatalf("bits=%v err=%v", bits, err)
	}
}

func TestStructuralVerilogSourceLatchSemantics(t *testing.T) {
	source := `module m(d,q,r); input d; output q,r; reg q,r;
dff decorative_name(q,d);
always @(posedge undeclared_event or negedge reset) r <= d;
endmodule`
	design, err := ParseStructuralVerilog("latches.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	module, ok := design.Module("m")
	if !ok || len(module.Latches) != 2 {
		t.Fatalf("module=%v latches=%v", ok, module.Latches)
	}
	if got := module.Latches[0]; got.Output != "q" || got.Input != "d" || got.Clock != "" || got.Initial != 0 {
		t.Fatalf("standard dff semantics: %+v", got)
	}
	// The native parser skips event-control contents instead of treating them
	// as a combinational dependency. The Go representation keeps the spelling
	// as metadata without requiring a corresponding declared net.
	if got := module.Latches[1]; got.Output != "r" || got.Input != "d" || got.Clock != "undeclared_event" || got.Initial != -1 {
		t.Fatalf("always semantics: %+v", got)
	}
}

func TestStructuralVerilogDefinedBoxAllowsOpenInput(t *testing.T) {
	source := `module child(a,y); input a; output y; assign y=a; endmodule
module top(y); output y; child u(.a(),.y(y)); endmodule`
	design, err := ParseStructuralVerilog("open.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	top, ok := design.Module("top")
	if !ok || len(top.Instances) != 1 || len(top.Instances[0].Connections) != 2 {
		t.Fatalf("top=%v instances=%+v", ok, top.Instances)
	}
	if actual := top.Instances[0].Connections[0].Actual; len(actual) != 0 {
		t.Fatalf("open input retained actuals: %v", actual)
	}
	if terminals := top.Instances[0].Terminals; !slices.Equal(terminals, []VerilogTerminal{{Formal: "a", Kind: VerilogInput}, {Formal: "y", Actual: "y", Kind: VerilogOutput}}) {
		t.Fatalf("open input terminals %+v", terminals)
	}
	if top.Nets["y"].Driver != "instance u" {
		t.Fatalf("output driver %q", top.Nets["y"].Driver)
	}
}

func TestStructuralVerilogMappedTwinGatePermutation(t *testing.T) {
	library, err := NewVerilogGateLibrary([]VerilogMappedCell{{Name: "DUAL", Inputs: []string{"A", "B"}, Output: "Y", TwinOutput: "Z"}})
	if err != nil {
		t.Fatal(err)
	}
	source := `module top(a,b,y,z); input a,b; output y,z; DUAL g(.B(b),.Y(y),.A(a),.Z(z)); endmodule`
	design, err := ParseStructuralVerilogWithOptions("mapped.v", []byte(source), VerilogParseOptions{Mapped: true, GateLibrary: library, Check: true})
	if err != nil {
		t.Fatal(err)
	}
	top, _ := design.Module("top")
	if !design.Mapped || !design.Finalized || len(top.Gates) != 1 {
		t.Fatalf("design mapped=%v finalized=%v gates=%+v", design.Mapped, design.Finalized, top.Gates)
	}
	gate := top.Gates[0]
	if !gate.Mapped || gate.Cell != "DUAL" || gate.Name != "g" || gate.Output != "y" || gate.TwinOutput != "z" || !slices.Equal(gate.Inputs, []string{"a", "b"}) || !slices.Equal(gate.PinPermutation, []int{1, 0}) {
		t.Fatalf("mapped gate %+v", gate)
	}
	if top.Nets["y"].Driver != "mapped DUAL" || top.Nets["z"].Driver != "mapped Z" {
		t.Fatalf("drivers y=%q z=%q", top.Nets["y"].Driver, top.Nets["z"].Driver)
	}

	clone := design.Clone()
	clone.Modules[0].Gates[0].PinPermutation[0] = 0
	clone.Modules[0].Gates[0].Inputs[0] = "changed"
	if !slices.Equal(gate.PinPermutation, []int{1, 0}) || gate.Inputs[0] != "a" {
		t.Fatal("mapped gate clone aliases source")
	}
	if parsed, parseErr := ParseStructuralVerilogWithOptions("unmapped.v", []byte(source), VerilogParseOptions{GateLibrary: library, Check: true}); parsed != nil || !errors.Is(parseErr, ErrUnsupported) {
		t.Fatalf("unmapped design=%v err=%v", parsed, parseErr)
	}
	if _, libraryErr := NewVerilogGateLibrary([]VerilogMappedCell{{Name: "BAD", Inputs: []string{"A", "A"}, Output: "Y"}}); !errors.Is(libraryErr, ErrDimension) {
		t.Fatalf("duplicate pin error %v", libraryErr)
	}
	if _, libraryErr := NewVerilogGateLibrary([]VerilogMappedCell{{Name: "X", Output: "Y"}, {Name: "X", Output: "Z"}}); !errors.Is(libraryErr, ErrVerilogDuplicate) {
		t.Fatalf("duplicate cell error %v", libraryErr)
	}
	// Public catalog snapshots must not mutate the lookup-owned cell data.
	library.Cells[0].Name = "changed"
	library.Cells[0].Inputs[0] = "changed"
	if reparsed, parseErr := ParseStructuralVerilogWithOptions("mapped-again.v", []byte(source), VerilogParseOptions{Mapped: true, GateLibrary: library, Check: true}); parseErr != nil || reparsed.Modules[0].Gates[0].Cell != "DUAL" {
		t.Fatalf("library snapshot mutation changed lookup: design=%v err=%v", reparsed, parseErr)
	}
}

func TestStructuralVerilogUndefinedBoxInferenceAndTerminals(t *testing.T) {
	source := `module top(a,y); input a; output y; mystery u(.i(a),.o(y)); endmodule`
	design, err := ParseStructuralVerilogWithOptions("undefined.v", []byte(source), VerilogParseOptions{Check: true, EmitHierarchyReport: true})
	if err != nil {
		t.Fatal(err)
	}
	box, ok := design.Module("mystery")
	if !ok || !box.Undefined || !box.Finalized || box.InstancesOf != 1 || !slices.Equal(box.Ports, []string{"i", "o"}) {
		t.Fatalf("undefined module ok=%v box=%+v", ok, box)
	}
	if box.Nets["i"].Kind != VerilogInput || box.Nets["o"].Kind != VerilogOutput {
		t.Fatalf("directions i=%v o=%v", box.Nets["i"].Kind, box.Nets["o"].Kind)
	}
	top, _ := design.Module("top")
	wantTerminals := []VerilogTerminal{{Formal: "i", Actual: "a", Kind: VerilogInput}, {Formal: "o", Actual: "y", Kind: VerilogOutput}}
	if !slices.Equal(top.Instances[0].Terminals, wantTerminals) || top.Nets["y"].Driver != "instance u" {
		t.Fatalf("terminals=%+v driver=%q", top.Instances[0].Terminals, top.Nets["y"].Driver)
	}
	if len(design.Warnings) != 1 || !strings.Contains(design.Warnings[0], "mystery") || !strings.Contains(design.HierarchyReport, "mystery undefbox instantiated 1 times") {
		t.Fatalf("warnings=%v report=%q", design.Warnings, design.HierarchyReport)
	}
}

func TestStructuralVerilogDefinedBoxTerminalsAndHierarchy(t *testing.T) {
	source := `module child(a,y); input [1:0] a; output y; assign y=a[0]; endmodule
module top(a,y); input [1:0] a; output y; child u(.y(y),.a(a)); endmodule`
	design, err := ParseStructuralVerilogWithOptions("hierarchy.v", []byte(source), VerilogParseOptions{Check: true, EmitHierarchyReport: true})
	if err != nil {
		t.Fatal(err)
	}
	child, _ := design.Module("child")
	top, _ := design.Module("top")
	if child.InstancesOf != 1 || len(top.Instances[0].Terminals) != 3 {
		t.Fatalf("instances=%d terminals=%+v", child.InstancesOf, top.Instances[0].Terminals)
	}
	want := []VerilogTerminal{{Formal: "y", Actual: "y", Kind: VerilogOutput}, {Formal: "a", Actual: "a[1]", Kind: VerilogInput}, {Formal: "a", Actual: "a[0]", Kind: VerilogInput, Bit: 1}}
	if !slices.Equal(top.Instances[0].Terminals, want) {
		t.Fatalf("terminals=%+v want=%+v", top.Instances[0].Terminals, want)
	}
	if !strings.Contains(design.HierarchyReport, "child logicbox instantiated 1 times") || !strings.Contains(design.HierarchyReport, "top logicbox instantiated 0 times") {
		t.Fatalf("report %q", design.HierarchyReport)
	}
	for _, invalid := range []string{
		`module child(a,y); input a; output y; assign y=a; endmodule module top(y); output y; child u(.y(y)); endmodule`,
		`module child(a,y); input a; output y; assign y=a; endmodule module top(a,y); input a; output y; child u(a); endmodule`,
	} {
		if parsed, parseErr := ParseStructuralVerilog("missing-terminal.v", []byte(invalid), VerilogLimits{}); parsed != nil || parseErr == nil {
			t.Fatalf("missing terminal design=%v err=%v", parsed, parseErr)
		}
	}
	// A named formal output may be deliberately omitted; the native parser
	// attaches it to an anonymous temporary net.
	if parsed, parseErr := ParseStructuralVerilog("open-output.v", []byte(`module child(a,y); input a; output y; assign y=a; endmodule module top(a); input a; child u(.a(a)); endmodule`), VerilogLimits{}); parseErr != nil || parsed == nil {
		t.Fatalf("open output design=%v err=%v", parsed, parseErr)
	} else if top, _ := parsed.Module("top"); !slices.Equal(top.Instances[0].Terminals, []VerilogTerminal{{Formal: "a", Actual: "a", Kind: VerilogInput}, {Formal: "y", Kind: VerilogOutput}}) {
		t.Fatalf("open output terminals %+v", top.Instances[0].Terminals)
	}
}

func TestStructuralVerilogDefinedBoxBitFormals(t *testing.T) {
	source := `module child(a,y); input [1:0] a; output [1:0] y; assign y[0]=a[0]; assign y[1]=a[1]; endmodule
module top(x0,x1,o0); input x0,x1; output o0; child u(.a[1](x1),.y[0](o0),.a[0](x0)); endmodule`
	design, err := ParseStructuralVerilog("bit-formals.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	top, _ := design.Module("top")
	want := []VerilogTerminal{
		{Formal: "a[1]", Actual: "x1", Kind: VerilogInput},
		{Formal: "y[0]", Actual: "o0", Kind: VerilogOutput},
		{Formal: "a[0]", Actual: "x0", Kind: VerilogInput},
		{Formal: "y", Kind: VerilogOutput, Bit: 1},
	}
	if !slices.Equal(top.Instances[0].Terminals, want) {
		t.Fatalf("terminals=%+v want=%+v", top.Instances[0].Terminals, want)
	}
	overlap := `module child(a,y); input [1:0] a; output y; assign y=a[0]; endmodule
module top(a,y); input [1:0] a; output y; child u(.a(a),.a[0](a[0]),.y(y)); endmodule`
	if parsed, parseErr := ParseStructuralVerilog("overlap.v", []byte(overlap), VerilogLimits{}); parsed != nil || !errors.Is(parseErr, ErrVerilogDuplicate) {
		t.Fatalf("overlap design=%v err=%v", parsed, parseErr)
	}
}

func TestStructuralVerilogDefinedBoxPositionalFlattening(t *testing.T) {
	source := `module child(a,y); input [1:0] a; output y; assign y=a[0]; endmodule
module top(x0,x1,y); input x0,x1; output y; child u({x1,x0},y); endmodule`
	design, err := ParseStructuralVerilog("positional-vector.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	top, _ := design.Module("top")
	want := []VerilogTerminal{
		{Formal: "a[0]", Actual: "x1", Kind: VerilogInput},
		{Formal: "a[1]", Actual: "x0", Kind: VerilogInput},
		{Formal: "y", Actual: "y", Kind: VerilogOutput},
	}
	if !slices.Equal(top.Instances[0].Terminals, want) || top.Nets["y"].Driver != "instance u" {
		t.Fatalf("terminals=%+v driver=%q", top.Instances[0].Terminals, top.Nets["y"].Driver)
	}
}

func TestStructuralVerilogFinalizationChecksAndUndefinedConflicts(t *testing.T) {
	undriven := []byte(`module m(y); output y; endmodule`)
	if design, err := ParseStructuralVerilog("undriven.v", undriven, VerilogLimits{}); design != nil || !errors.Is(err, ErrVerilogUndefined) {
		t.Fatalf("checked design=%v err=%v", design, err)
	}
	design, err := ParseStructuralVerilogWithOptions("unchecked.v", undriven, VerilogParseOptions{})
	if err != nil || !design.Finalized || !design.Modules[0].Finalized {
		t.Fatalf("unchecked design=%+v err=%v", design, err)
	}

	mixed := []byte(`module top(a,y); input a; output y; wire n; assign n=a; mystery u0(.p(y)); mystery u1(.p(n)); endmodule`)
	design, err = ParseStructuralVerilog("mixed.v", mixed, VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	box, _ := design.Module("mystery")
	if box.Nets["p"].Kind != VerilogInput {
		t.Fatalf("mixed formal direction %v", box.Nets["p"].Kind)
	}
	top, _ := design.Module("top")
	if top.Nets["y"].Driver != "constant 0" || len(design.Warnings) < 2 {
		t.Fatalf("mixed finalization driver=%q warnings=%v", top.Nets["y"].Driver, design.Warnings)
	}
	if design, err = ParseStructuralVerilog("limited.v", []byte(`module top(y); output y; mystery u(.o(y)); endmodule`), VerilogLimits{MaxModules: 1}); design != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("undefined module limit design=%v err=%v", design, err)
	}
	if design, err = ParseStructuralVerilog("limited-nets.v", []byte(`module top(); mystery u(.a(1'b0),.b(1'b1)); endmodule`), VerilogLimits{MaxNets: 1}); design != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("undefined net limit design=%v err=%v", design, err)
	}

	cancelSource := []byte(`module top(a,y); input a; output y; mystery u(.i(a),.o(y)); endmodule`)
	baselineCalls := 0
	if _, err = ParseStructuralVerilogWithOptions("count.v", cancelSource, VerilogParseOptions{Check: true, Limits: VerilogLimits{Stop: func() bool { baselineCalls++; return false }}}); err != nil {
		t.Fatal(err)
	}
	if baselineCalls < 3 {
		t.Fatalf("unexpected stop checks %d", baselineCalls)
	}
	calls, cutoff := 0, baselineCalls-1
	design, err = ParseStructuralVerilogWithOptions("cancel-finalize.v", cancelSource, VerilogParseOptions{Check: true, Limits: VerilogLimits{Stop: func() bool { calls++; return calls >= cutoff }}})
	if design != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("late cancellation calls=%d/%d design=%v err=%v", calls, cutoff, design, err)
	}
}

func TestStructuralVerilogNativeNonDrivenFinalization(t *testing.T) {
	source := []byte(`module m(a,y,z); input a; output y,z; assign y=a; endmodule`)
	design, err := ParseStructuralVerilog("nondriven-with-logic.v", source, VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	module, _ := design.Module("m")
	if module.Nets["z"].Driver != "constant 0" || len(module.Assignments) != 2 || len(design.Warnings) != 1 || !strings.Contains(design.Warnings[0], "z") {
		t.Fatalf("module=%+v assignments=%+v warnings=%v", module.Nets["z"], module.Assignments, design.Warnings)
	}

	open, err := ParseStructuralVerilog("empty-undefined.v", []byte(`module top(); mystery u(.o()); endmodule`), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	box, _ := open.Module("mystery")
	if box.Nets["o"].Kind != VerilogOutput {
		t.Fatalf("empty undefined formal kind %v", box.Nets["o"].Kind)
	}

	inout, err := ParseStructuralVerilogWithOptions("inout.v", []byte(`module io(p); inout p; endmodule`), VerilogParseOptions{Check: true, EmitHierarchyReport: true})
	if err != nil || !strings.Contains(inout.HierarchyReport, "pi=1 po=1") {
		t.Fatalf("inout report=%q err=%v", inout.HierarchyReport, err)
	}
}

func TestStructuralVerilogMultipleContinuousAssignments(t *testing.T) {
	source := []byte(`module m(a,b,y,z); input a,b; output y,z; assign y=a, z=b; endmodule`)
	design, err := ParseStructuralVerilog("multi-assign.v", source, VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	module, _ := design.Module("m")
	if len(module.Assignments) != 2 || module.Assignments[0].Output != "y" || module.Assignments[1].Output != "z" || module.Nets["y"].Driver != "assign" || module.Nets["z"].Driver != "assign" {
		t.Fatalf("assignments=%+v y=%q z=%q", module.Assignments, module.Nets["y"].Driver, module.Nets["z"].Driver)
	}
}

func TestStructuralVerilogSourceReductionAssignment(t *testing.T) {
	source := []byte(`module m(a,b,y); input a,b; output y; assign {y} = ~&{a,b}; endmodule`)
	design, err := ParseStructuralVerilog("reduction.v", source, VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	module, _ := design.Module("m")
	if len(module.Assignments) != 1 || !slices.Equal(module.Assignments[0].Inputs, []string{"a", "b"}) {
		t.Fatalf("assignment=%+v", module.Assignments)
	}
	value, err := module.Assignments[0].Expression.Eval(map[string]bool{"a": true, "b": true})
	if err != nil || value {
		t.Fatalf("value=%v err=%v", value, err)
	}
}

func TestStructuralVerilogDefinedBoxExpandsBinaryActual(t *testing.T) {
	source := `module child(a,y); input [3:0] a; output y; assign y=a[0]; endmodule
module top(y); output y; child u(.a({4'b10x1}),.y(y)); endmodule`
	design, err := ParseStructuralVerilog("constant-actual.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	top, _ := design.Module("top")
	if got := top.Instances[0].Connections[0].Actual; !slices.Equal(got, []string{"1'b1", "1'b0", "1'b0", "1'b1"}) {
		t.Fatalf("constant actual expansion %v", got)
	}
}

func TestStructuralVerilogToolOpenActual(t *testing.T) {
	source := `module sink(a); input a; endmodule
module top(); sink u(.a(Open_unused)); endmodule`
	design, err := ParseStructuralVerilog("open-tool.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatal(err)
	}
	top, _ := design.Module("top")
	if len(top.Instances) != 1 || len(top.Instances[0].Connections[0].Actual) != 0 {
		t.Fatalf("open actual %+v", top.Instances)
	}
}

func TestStructuralVerilogFileAndStreamFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "small.v")
	data := []byte("module m(a); input a; endmodule")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	stream, err := OpenVerilogStream(path, VerilogLimits{MaxBytes: 128})
	if err != nil {
		t.Fatal(err)
	}
	if size, _ := stream.FileSize(); size != len(data) {
		t.Fatalf("size %d", size)
	}
	design, err := ParseStructuralVerilogFile(path, VerilogLimits{MaxBytes: 128})
	if err != nil || len(design.Modules) != 1 {
		t.Fatalf("design=%v err=%v", design, err)
	}
	if _, err = ParseStructuralVerilogFile(path, VerilogLimits{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit error %v", err)
	}
}

func TestVerilogErrorsLimitsCancellationAndAtomicFailure(t *testing.T) {
	formulaCases := []string{"", "a &", "(a", "a b", "a ? b", "a : b", "!"}
	for _, input := range formulaCases {
		if expr, names, err := ParseVerilogFormula(input, VerilogLimits{}); err == nil || expr != nil || names != nil {
			t.Fatalf("formula %q expr=%v names=%v err=%v", input, expr, names, err)
		}
	}
	if _, _, err := ParseVerilogReduction("&{}", VerilogLimits{}); !errors.Is(err, ErrVerilogSyntax) {
		t.Fatalf("empty reduction %v", err)
	}
	if _, _, err := ParseVerilogReduction("&{a,b}", VerilogLimits{MaxBytes: 3}); !errors.Is(err, ErrLimit) {
		t.Fatalf("reduction byte limit %v", err)
	}
	if _, _, err := ParseVerilogReduction("~&{a,b}", VerilogLimits{MaxExprNodes: 3}); !errors.Is(err, ErrLimit) {
		t.Fatalf("reduction node limit %v", err)
	}
	if _, _, err := ParseVerilogReduction("&{a,b}", VerilogLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("reduction cancellation %v", err)
	}
	if _, _, err := ParseVerilogFormula("a&b", VerilogLimits{MaxExprNodes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit %v", err)
	}
	if _, _, err := ParseVerilogFormula("a", VerilogLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	if _, err := NewVerilogStream("x", []byte("large"), VerilogLimits{MaxBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatalf("stream limit %v", err)
	}
	stream, _ := NewVerilogStream("x", []byte("abcdef"), VerilogLimits{MaxWordBytes: 2})
	if _, err := stream.Word(","); !errors.Is(err, ErrLimit) {
		t.Fatalf("word limit %v", err)
	}
	calls := 0
	stream, _ = NewVerilogStream("x", []byte("abcdef"), VerilogLimits{Stop: func() bool {
		calls++
		return calls > 3
	}})
	if _, err := stream.Word(","); !errors.Is(err, ErrCancelled) {
		t.Fatalf("stream cancellation %v", err)
	}

	badDesigns := []struct {
		text   string
		target error
	}{
		{"", ErrVerilogSyntax},
		{"wire a;", ErrVerilogSyntax},
		{"module m(a); input a; input a; endmodule", nil},
		{"module m(a); input a; output a; endmodule", ErrVerilogDuplicate},
		{"module m(a,y); input a; output y; assign y = missing; endmodule", ErrVerilogUndefined},
		{"module m(a,y); input a; output y; assign y=a; buf(y,a); endmodule", ErrVerilogDriven},
		{"module m(a); input a; /*", ErrVerilogSyntax},
		{"module m(a); input [99:0] a; endmodule", ErrLimit},
	}
	for _, test := range badDesigns {
		design, err := ParseStructuralVerilog("bad.v", []byte(test.text), VerilogLimits{MaxVectorBits: 8})
		if test.target == nil {
			if err != nil || design == nil {
				t.Fatalf("duplicate same kind text=%q design=%v err=%v", test.text, design, err)
			}
			continue
		}
		if design != nil || !errors.Is(err, test.target) {
			t.Fatalf("text=%q design=%v err=%v want=%v", test.text, design, err, test.target)
		}
	}
	if design, err := ParseStructuralVerilog("x", []byte("module m(); endmodule"), VerilogLimits{MaxStatements: 1, Stop: func() bool { return true }}); design != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel design=%v err=%v", design, err)
	}
	if bits, err := ParseVerilogBinaryConstant("4'hf", 8); bits != nil || !errors.Is(err, ErrVerilogSyntax) {
		t.Fatalf("constant bits=%v err=%v", bits, err)
	}
	if bits, err := ParseVerilogBinaryConstant("9'b000000000", 8); bits != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("constant limit bits=%v err=%v", bits, err)
	}
}

func independentFormulaValue(a, b, c, d bool, selector int) bool {
	switch selector % 6 {
	case 0:
		return a && b
	case 1:
		return (a != b) || c
	case 2:
		return (!a) && (b || c)
	case 3:
		if a {
			return b
		}
		return c
	case 4:
		return (a || b) != (c && d)
	default:
		if a != b {
			return !c
		}
		return d
	}
}

func formulaForSelector(selector int) string {
	switch selector % 6 {
	case 0:
		return "a & b"
	case 1:
		return "a ^ b | c"
	case 2:
		return "~a & (b | c)"
	case 3:
		return "a ? b : c"
	case 4:
		return "(a | b) ^ (c & d)"
	default:
		return "(a ^ b) ? ~c : d"
	}
}

func TestVerilogFormulaModelCampaign(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5e71f1ed))
	for execution := 0; execution < 5000; execution++ {
		selector := rng.Intn(24)
		a, b, c, d := rng.Intn(2) != 0, rng.Intn(2) != 0, rng.Intn(2) != 0, rng.Intn(2) != 0
		expr, _, err := ParseVerilogFormula(formulaForSelector(selector), VerilogLimits{})
		if err != nil {
			t.Fatal(err)
		}
		got, err := expr.Eval(map[string]bool{"a": a, "b": b, "c": c, "d": d})
		if err != nil {
			t.Fatal(err)
		}
		if want := independentFormulaValue(a, b, c, d, selector); got != want {
			t.Fatalf("execution=%d selector=%d got=%v want=%v", execution, selector, got, want)
		}
	}
}

func verilogModelName(prefix byte, index int) string {
	return string([]byte{prefix, byte('0' + index)})
}

func undefinedDirectionCase(seed uint64) (string, map[string]VerilogSignalKind) {
	rng := rand.New(rand.NewSource(int64(seed)))
	count := 2 + rng.Intn(5)
	permutation := rng.Perm(count)
	kinds := make(map[string]VerilogSignalKind, count)
	inputs, outputs, actuals := make([]string, 0, count), make([]string, 0, count), make([]string, count)
	for index := 0; index < count; index++ {
		formal, actual := verilogModelName('f', index), verilogModelName('n', index)
		actuals[index] = actual
		if rng.Intn(2) == 0 {
			kinds[formal] = VerilogInput
			inputs = append(inputs, actual)
		} else {
			kinds[formal] = VerilogOutput
			outputs = append(outputs, actual)
		}
	}
	var source strings.Builder
	source.WriteString("module top(")
	source.WriteString(strings.Join(actuals, ","))
	source.WriteString(");")
	if len(inputs) != 0 {
		source.WriteString(" input ")
		source.WriteString(strings.Join(inputs, ","))
		source.WriteByte(';')
	}
	if len(outputs) != 0 {
		source.WriteString(" output ")
		source.WriteString(strings.Join(outputs, ","))
		source.WriteByte(';')
	}
	source.WriteString(" mystery u(")
	for position, index := range permutation {
		if position != 0 {
			source.WriteByte(',')
		}
		source.WriteByte('.')
		source.WriteString(verilogModelName('f', index))
		source.WriteByte('(')
		source.WriteString(actuals[index])
		source.WriteByte(')')
	}
	source.WriteString("); endmodule")
	return source.String(), kinds
}

func checkUndefinedDirectionCase(t *testing.T, seed uint64) {
	t.Helper()
	source, want := undefinedDirectionCase(seed)
	design, err := ParseStructuralVerilog("direction-model.v", []byte(source), VerilogLimits{})
	if err != nil {
		t.Fatalf("seed=%d source=%q err=%v", seed, source, err)
	}
	box, ok := design.Module("mystery")
	if !ok || !box.Undefined || len(box.Ports) != len(want) {
		t.Fatalf("seed=%d box=%+v", seed, box)
	}
	for formal, kind := range want {
		if net := box.Nets[formal]; net == nil || net.Kind != kind {
			t.Fatalf("seed=%d formal=%s got=%+v want=%v source=%q", seed, formal, net, kind, source)
		}
	}
}

func mappedPermutationCase(seed uint64) (string, []int) {
	rng := rand.New(rand.NewSource(int64(seed ^ 0x6d6170706564)))
	permutation := rng.Perm(4)
	var source strings.Builder
	source.WriteString("module top(a,b,c,d,y,z); input a,b,c,d; output y,z; QUAD g(")
	for position, pin := range permutation {
		if position != 0 {
			source.WriteByte(',')
		}
		source.WriteString(".P")
		source.WriteByte(byte('0' + pin))
		source.WriteByte('(')
		source.WriteByte(byte('a' + pin))
		source.WriteByte(')')
	}
	source.WriteString(",.Y(y),.Z(z)); endmodule")
	return source.String(), permutation
}

func checkMappedPermutationCase(t *testing.T, library *VerilogGateLibrary, seed uint64) {
	t.Helper()
	source, permutation := mappedPermutationCase(seed)
	design, err := ParseStructuralVerilogWithOptions("mapped-model.v", []byte(source), VerilogParseOptions{Mapped: true, GateLibrary: library, Check: true})
	if err != nil {
		t.Fatalf("seed=%d source=%q err=%v", seed, source, err)
	}
	gate := design.Modules[0].Gates[0]
	if !slices.Equal(gate.Inputs, []string{"a", "b", "c", "d"}) || !slices.Equal(gate.PinPermutation, permutation) || gate.Output != "y" || gate.TwinOutput != "z" {
		t.Fatalf("seed=%d gate=%+v permutation=%v", seed, gate, permutation)
	}
}

func TestVerilogBoxAndMappedGateModelCampaign(t *testing.T) {
	library, err := NewVerilogGateLibrary([]VerilogMappedCell{{Name: "QUAD", Inputs: []string{"P0", "P1", "P2", "P3"}, Output: "Y", TwinOutput: "Z"}})
	if err != nil {
		t.Fatal(err)
	}
	for execution := uint64(0); execution < 5000; execution++ {
		seed := execution*0x9e3779b97f4a7c15 + 0x766572696c6f67
		checkUndefinedDirectionCase(t, seed)
		checkMappedPermutationCase(t, library, seed)
	}
}

func FuzzVerilogFormulaModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x123456789abcdef0} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		selector := int(seed % 6)
		a := seed&(1<<8) != 0
		b := seed&(1<<17) != 0
		c := seed&(1<<31) != 0
		d := seed&(1<<47) != 0
		expr, _, err := ParseVerilogFormula(formulaForSelector(selector), VerilogLimits{})
		if err != nil {
			t.Fatal(err)
		}
		got, err := expr.Eval(map[string]bool{"a": a, "b": b, "c": c, "d": d})
		if err != nil {
			t.Fatal(err)
		}
		if want := independentFormulaValue(a, b, c, d, selector); got != want {
			t.Fatalf("seed=%d got=%v want=%v", seed, got, want)
		}
	})
}

func FuzzVerilogUndefinedDirectionModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x123456789abcdef0} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		checkUndefinedDirectionCase(t, seed)
	})
}

func FuzzVerilogMappedPermutationModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0xfedcba9876543210} {
		f.Add(seed)
	}
	library, err := NewVerilogGateLibrary([]VerilogMappedCell{{Name: "QUAD", Inputs: []string{"P0", "P1", "P2", "P3"}, Output: "Y", TwinOutput: "Z"}})
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		checkMappedPermutationCase(t, library, seed)
	})
}

func FuzzStructuralVerilogRejectOrValidate(f *testing.F) {
	for _, seed := range []string{"module m(a); input a; endmodule", "", "module m(a,y); input a; output y; assign y=~a; endmodule"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 4096 {
			t.Skip()
		}
		design, err := ParseStructuralVerilog("fuzz.v", []byte(text), VerilogLimits{MaxBytes: 4096, MaxWordBytes: 256, MaxTokens: 1024, MaxModules: 16, MaxNets: 256, MaxStatements: 256, MaxExprNodes: 256, MaxVectorBits: 64})
		if err != nil {
			if design != nil {
				t.Fatal("failed parse returned partial design")
			}
			return
		}
		if design == nil || len(design.Modules) == 0 {
			t.Fatal("successful parse has no modules")
		}
		for _, module := range design.Modules {
			if module == nil || module.Name == "" || module.Nets == nil {
				t.Fatal("invalid module")
			}
			for _, name := range module.NetOrder {
				net := module.Nets[name]
				if net == nil || net.Name != name {
					t.Fatal("invalid net order")
				}
			}
		}
	})
}

func TestVerilogSourceFilesBeginWithPackage(t *testing.T) {
	for _, name := range []string{"base_ver.go", "base_ver_test.go"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(data), "package abc\n") {
			t.Fatalf("%s has pre-package boilerplate", name)
		}
	}
}
