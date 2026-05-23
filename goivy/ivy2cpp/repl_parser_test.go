package ivy2cpp

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestParserCoverageMatrixMatchesPython(t *testing.T) {
	rows := readParserMatrix(t)
	expected := []string{
		"bool",
		"int",
		"nat",
		"strlit",
		"bv",
		"strbv",
		"intbv",
		"named_enum",
		"numeric_enum",
		"range",
		"destructor_record",
		"variant",
		"native",
		"positional_function",
	}
	if len(rows) != len(expected) {
		t.Fatalf("parser matrix has %d rows, want %d: %#v", len(rows), len(expected), rows)
	}
	for _, kind := range expected {
		row, ok := rows[kind]
		if !ok {
			t.Fatalf("parser matrix missing kind %q", kind)
		}
		for _, key := range []string{"python", "go"} {
			if strings.TrimSpace(row[key]) == "" {
				t.Fatalf("parser matrix row %q missing %s: %#v", kind, key, row)
			}
		}
		if row["status"] != "covered" {
			t.Fatalf("parser matrix row %q status=%q, want covered", kind, row["status"])
		}
	}
}

func readParserMatrix(t *testing.T) map[string]map[string]string {
	t.Helper()
	data, err := os.ReadFile("testdata/parser_matrix.yaml")
	if err != nil {
		t.Fatalf("read parser matrix: %v", err)
	}
	rows := map[string]map[string]string{}
	current := ""
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- kind:") {
			current = strings.TrimSpace(strings.TrimPrefix(line, "- kind:"))
			rows[current] = map[string]string{"kind": current}
			continue
		}
		if current == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("bad parser matrix line %q", raw)
		}
		rows[current][strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return rows
}

func TestParserBitvectorWidthN(t *testing.T) {
	tests := []struct {
		width int
		want  []string
	}{
		{width: 1, want: []string{"typedef unsigned word;", "_arg<unsigned>(args, 0, 2)"}},
		{width: 2, want: []string{"typedef unsigned word;", "_arg<unsigned>(args, 0, 4)"}},
		{width: 8, want: []string{"typedef unsigned word;", "_arg<unsigned>(args, 0, 256)"}},
		{width: 16, want: []string{"typedef unsigned word;", "_arg<unsigned>(args, 0, 65536)"}},
		{width: 32, want: []string{"typedef unsigned word;", "_arg<unsigned>(args, 0, 4294967296)"}},
		{width: 64, want: []string{"typedef unsigned long long word;", "_arg<unsigned long long>(args, 0, 0)"}},
		{width: 128, want: []string{"typedef unsigned __int128 word;", "_arg<unsigned __int128>(args, 0, 0)"}},
		{width: 256, want: []string{"typedef ivy_uint<256> word;", "_arg<bvparse256::word>(args, 0, 0)", "template <> ivy_uint<256> _arg<ivy_uint<256>>"}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("bv%d", tt.width), func(t *testing.T) {
			className := fmt.Sprintf("bvparse%d", tt.width)
			mod := compileIvySource(t, fmt.Sprintf(`#lang ivy1.7
type word
interpret word -> bv[%d]
action echo(x:word) returns (out:word) = {
    out := x
}
export echo
`, tt.width))
			out, err := Generate(mod, Config{Target: "repl", ClassName: className})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			text := out.Header + "\n" + out.Impl
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Fatalf("bv[%d] missing %q:\nheader:\n%s\nimpl:\n%s", tt.width, want, out.Header, out.Impl)
				}
			}
			compileGeneratedCPP(t, out)
		})
	}
}

func TestParserNamedEnumByNameAndNumericEnumByNumber(t *testing.T) {
	named := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
action echo(c:color) returns (out:color) = {
    out := c
}
export echo
`)
	namedOut, err := Generate(named, Config{Target: "repl", ClassName: "enumparse"})
	if err != nil {
		t.Fatalf("Generate named enum: %v", err)
	}
	for _, want := range []string{
		"enumparse::color _arg<enumparse::color>(std::vector<ivy_value> &args, unsigned idx, long long bound) {",
		`if (arg.atom == "red") return enumparse::red;`,
		`if (arg.atom == "green") return enumparse::green;`,
		`if (arg.atom == "blue") return enumparse::blue;`,
		"_arg<enumparse::color>(args, 0, 3)",
	} {
		if !strings.Contains(namedOut.Impl, want) {
			t.Fatalf("named enum output missing %q:\n%s", want, namedOut.Impl)
		}
	}
	compileGeneratedCPP(t, namedOut)

	numeric := compileIvySource(t, `#lang ivy1.7
type digit = {0, 1, 2}
action echo(d:digit) returns (out:digit) = {
    out := d
}
export echo
`)
	numericOut, err := Generate(numeric, Config{Target: "repl", ClassName: "numenum"})
	if err != nil {
		t.Fatalf("Generate numeric enum: %v", err)
	}
	text := numericOut.Header + "\n" + numericOut.Impl
	if !strings.Contains(text, "_arg<int>(args, 0, 3)") {
		t.Fatalf("numeric enum should route through primitive int parser:\n%s", text)
	}
	for _, bad := range []string{"enum digit", "_arg<numenum::digit>"} {
		if strings.Contains(text, bad) {
			t.Fatalf("numeric enum should not emit named-enum parser %q:\n%s", bad, text)
		}
	}
	compileGeneratedCPP(t, numericOut)
}

func TestParserRangeBoundsRespected(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type slot = {2..7}
action echo(s:slot) returns (out:slot) = {
    out := s
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "rangeparse"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + "\n" + out.Impl
	for _, want := range []string{
		"typedef unsigned slot;",
		"_arg<unsigned>(args, 0, 8)",
		`std::cerr << "line " << lineno << ":" << err.pos << ": " << err.txt << " bad value" << std::endl;`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("range parser output missing %q:\n%s", want, text)
		}
	}
	runtime := readSupportHeader(t, "ivy_value.hpp")
	if !strings.Contains(runtime, "res < 0 || res >= bound") {
		t.Fatalf("runtime integer _arg<T> lost Python-style bound check:\n%s", runtime)
	}
	if want := strings.TrimSpace(readTestdata(t, "testdata/python_errors/range_out_of_bounds.txt")); want == "" {
		t.Fatalf("captured Python range error fixture is empty")
	}
	compileGeneratedCPP(t, out)
}

func TestParserStringWithEscapes(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strlit
action echo(s:text) returns (out:text) = {
    out := s
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "stringparse"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "_arg<__strlit>(args, 0, 0)") {
		t.Fatalf("strlit parser should use runtime __strlit _arg:\n%s", out.Impl)
	}
	replRuntime := readSupportHeader(t, "ivy_repl.hpp")
	for _, want := range []string{
		`c = (c == 'n') ? 10 : (c == 'r') ? 13 : (c == 't') ? 9 : c;`,
		"res.atom.push_back(c);",
	} {
		if !strings.Contains(replRuntime, want) {
			t.Fatalf("parse_value escape support missing %q:\n%s", want, replRuntime)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestParserDestructorRecord(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type leaf
destructor shade(L:leaf) : color
type cell
destructor child(C:cell) : leaf
destructor ok(C:cell) : bool
action echo(c:cell) returns (out:cell) = {
    out := c
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "recordparse"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"template <> recordparse::cell _arg<recordparse::cell>(std::vector<ivy_value> &args, unsigned idx, long long bound) {",
		`if (arg.fields[i].atom == "child") {`,
		`res.child = _arg<recordparse::leaf>(tmp_args, 0, 0);`,
		`throw out_of_bounds("in field child: " + err.txt, err.pos);`,
		"template <> recordparse::leaf _arg<recordparse::leaf>(std::vector<ivy_value> &args, unsigned idx, long long bound) {",
		`res.shade = _arg<recordparse::color>(tmp_args, 0, 2);`,
		`else throw out_of_bounds("unexpected field: " + arg.fields[i].atom, arg.fields[i].pos);`,
		`throw out_of_bounds("expected struct", args[idx].pos);`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("destructor parser missing %q:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestParserVariantDiscriminator(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type msg
type color = {red, green}
variant request of msg = struct {
    shade : color
}
variant ack of msg
action echo(m:msg) returns (out:msg) = {
    out := m
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "variantparse"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"template <> variantparse::msg _arg<variantparse::msg>(std::vector<ivy_value> &args, unsigned idx, long long bound) {",
		`throw out_of_bounds("unexpected value for sort msg: " + args[idx].atom, args[idx].pos);`,
		`if (args[idx].fields[0].atom == "request") return variantparse::msg(0, new variantparse::msg::twrap<variantparse::request>(_arg<variantparse::request>(args[idx].fields[0].fields, 0, 0)));`,
		`if (args[idx].fields[0].atom == "ack") return variantparse::msg(1, new variantparse::msg::twrap<variantparse::ack>(_arg<variantparse::ack>(args[idx].fields[0].fields, 0, 0)));`,
		`throw out_of_bounds("unexpected field sort msg: " + args[idx].fields[0].atom, args[idx].pos);`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("variant parser missing %q:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestParserNativeSortRoutesToUserProvided(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
<<< header
struct TokenBase {
    TokenBase() : value(0) {}
    int value;
};
>>>
type token
interpret token -> <<< TokenBase >>>
<<< impl
template <>
nativeparse::token _arg<nativeparse::token>(std::vector<ivy_value> &args, unsigned idx, long long bound) {
    nativeparse::token res;
    res.value = _arg<int>(args, idx, bound);
    return res;
}
>>>
action set(t:token) = {}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "nativeparse"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	dispatch := "ivy.set(_arg<nativeparse::token>(args, 0, 0));"
	userParser := "nativeparse::token _arg<nativeparse::token>(std::vector<ivy_value> &args, unsigned idx, long long bound) {"
	if !strings.Contains(out.Impl, dispatch) {
		t.Fatalf("native dispatch missing user-parser route %q:\n%s", dispatch, out.Impl)
	}
	parserIdx := strings.Index(out.Impl, userParser)
	readerIdx := strings.Index(out.Impl, "class nativeparse_cmd_reader : public stdin_reader")
	if parserIdx < 0 || readerIdx < 0 || parserIdx > readerIdx {
		t.Fatalf("native _arg specialization should be emitted before cmd_reader use; parserIdx=%d readerIdx=%d\n%s", parserIdx, readerIdx, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestParserRangeBoundsErrorMatchesPythonSlow(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to run generated parser binary")
	}
	mod := compileIvySource(t, `#lang ivy1.7
type slot = {2..7}
action echo(s:slot) returns (out:slot) = {
    out := s
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "rangeparse"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	_, stderr := runGeneratedReplSlow(t, out, "echo(8)\n")
	want := strings.TrimSpace(readTestdata(t, "testdata/python_errors/range_out_of_bounds.txt"))
	if got := strings.TrimSpace(stderr); got != want {
		t.Fatalf("range parser stderr mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func runGeneratedReplSlow(t *testing.T, out *Output, input string) (string, string) {
	t.Helper()
	path, err := BuildOutput(out, t.TempDir())
	if err != nil {
		if isMissingZ3ToolchainError(err) {
			t.Skip(err.Error())
		}
		t.Fatalf("build generated repl: %v", err)
	}
	cmd := exec.Command(path)
	cmd.Stdin = strings.NewReader(input)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run generated repl: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	return stdout.String(), stderr.String()
}

func readTestdata(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
