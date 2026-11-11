package startingpoint

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runLibraryProgram(t *testing.T, source string) (string, []LibraryChange) {
	t.Helper()
	draft, changes, err := RewriteStandardLibrary(source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = parser.ParseFile(token.NewFileSet(), "generated.go", draft, parser.AllErrors); err != nil {
		t.Fatalf("%v\n%s", err, draft)
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	if err = os.WriteFile(file, []byte(draft), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", file)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s\n%s", err, output, draft)
	}
	return string(output), changes
}

func TestStdlibNumericRuntime(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import "fmt"
func main() {
 for _, text := range []string{"  +12junk", "-17tail", "abc", "+", "2147483647", "\t42\n"} { fmt.Println(atoi(text)); }
 fmt.Println(atoll("-9223372036854775807rest"));
 fmt.Println(abs(-4), isdigit('8'), isalpha('!'), isspace('\n'), tolower('A'), toupper(-1));
 fmt.Println(sqrt(9), pow(2,3), fmin(0.0/1.0, 4));
}`)
	want := "12\n-17\n0\n0\n2147483647\n42\n-9223372036854775807\n4 1 0 1 97 -1\n3 8 0\n"
	if output != want {
		t.Fatalf("got %q want %q", output, want)
	}
}

func TestStdlibFileRuntime(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import ("fmt"; "os")
func main() {
 dir, err := os.MkdirTemp("", "stdlib-fixture"); if err != nil { panic(err) }; defer os.RemoveAll(dir)
 path := dir + "/data"
 var file *os.File
 file = fopen(path, "wb"); if file == nil { panic("open") }
 fputs("abc\ndefgh\nz", file);
 fmt.Println(fclose(file));
 file = fopen(path, "rb");
 var buffer [5]byte
 for fgets(buffer, 5, file) != nil {
  fmt.Printf("%q:%d\n", string(buffer[:strlen(buffer)]), strlen(buffer));
 }
 fmt.Println(fclose(file));
 file = fopen(path, "ab"); fputc('!',file); fclose(file);
 file = fopen(path, "rb"); fmt.Println(fseek(file, -1, SEEK_END), ftell(file), fgetc(file), fgetc(file)); fclose(file);
 fmt.Println(fopen(dir + "/missing", "r") == nil);
 fmt.Println(rename(path, path+"2"), remove(path+"2"));
}`)
	want := "0\n\"abc\\n\":4\n\"defg\":4\n\"h\\n\":2\n\"z\":1\n0\n0 11 33 -1\ntrue\n0 0\n"
	if output != want {
		t.Fatalf("got %q want %q", output, want)
	}
}

func TestStdlibMemoryRuntime(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import "fmt"
func main() {
 var a [8]byte; var b [8]byte
 memset(a, 'x', 4); memcpy(b, a, 4);
 fmt.Println(memcmp(a,b,4), strlen(a));
 var text string = "hi"
 strcat(text, "!"); fmt.Println(text, strcmp(text,"hi!"), strncmp("abcd","abxy",2));
}`)
	if output != "0 4\nhi! 0 0\n" {
		t.Fatal(output)
	}
}

func TestStdlibProtectionAndReview(t *testing.T) {
	source := `package p
func f() {
 // fopen("x", "w")
 s := "atoi(100)"
 obj.atoi(s)
 file = fopen(s, mode)
 fgets(raw, n, file)
 malloc(n)
 strstr(s, "x")
}`
	draft, changes, err := RewriteStandardLibrary(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`// fopen("x", "w")`, `"atoi(100)"`, `obj.atoi(s)`, `fopen(s, mode)`, `fgets(raw, n, file)`, `malloc(n)`, `strstr(s, "x")`} {
		if !strings.Contains(draft, want) {
			t.Fatalf("lost %q:\n%s", want, draft)
		}
	}
	if strings.Contains(draft, "import ") {
		t.Fatal("unused imports for retained calls", draft)
	}
	reviews := 0
	for _, c := range changes {
		if c.Status == "review" {
			reviews++
		}
		if c.Location.Line < 1 || c.Location.Column < 1 || c.Coordinates != "pre-library Go draft" {
			t.Fatal(c)
		}
	}
	if reviews != 4 {
		t.Fatalf("review count %d", reviews)
	}
}

func TestStdlibAliasesAndNestedCalls(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import (f "fmt"; st "strconv")
func main() {
 _ = st.IntSize
 strings := "local"
 f.Println(strings, abs(atoi("-12foo")))
}`)
	if output != "local 12\n" {
		t.Fatal(output)
	}
}

func TestStdlibShadowedNames(t *testing.T) {
	source := `package p
func atoi(x string) int { return 7 }
func f(stdout *os.File) { atoi("9"); fprintf(stdout, "%d", 1); }
`
	draft, _, err := RewriteStandardLibrary(source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(draft, `atoi("9")`) || strings.Contains(draft, "ParseInt") || strings.Contains(draft, "os.Stdout") {
		t.Fatal(draft)
	}
}

func TestStdlibArgumentEvaluationAndNesting(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import "fmt"
var calls int
func next() string { calls++; return " -4x" }
func main() { fmt.Println(abs(atoi(next())), calls); }
`)
	if output != "4 1\n" {
		t.Fatal(output)
	}
}

func TestStdlibFormatAndFailures(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import "fmt"
func main() {
 printf("%ld %u %zu\n", int64(2), uint32(3), uint64(4));
 fmt.Println(printf("x"));
}`)
	if output != "2 3 4\nx1\n" {
		t.Fatal(output)
	}
	draft, changes, err := RewriteStandardLibrary("package p\nfunc f() { printf(\"%n\", &n); }\n")
	if err != nil || !strings.Contains(draft, "printf(") || len(changes) != 1 || changes[0].Status != "review" || strings.Contains(draft, "import ") {
		t.Fatal(draft, changes, err)
	}
}

func TestStdlibDifferentialCConsumer(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for reference run")
	}
	source := `#include <stdio.h>
#include <stdlib.h>
#include <string.h>
int main(int argc, char *argv[]) {
    FILE *file = fopen(argv[1], "rb");
    char buffer[5];
    if (file == NULL) return 2;
    printf("%d %d\n", atoi(" +12tail"), atoi("-8!"));
    while (fgets(buffer, 5, file) != NULL) {
        printf("%zu:%s|", strlen(buffer), buffer);
    }
    fclose(file);
    return 0;
}`
	dir := t.TempDir()
	cfile := filepath.Join(dir, "reference.c")
	exe := filepath.Join(dir, "reference")
	input := filepath.Join(dir, "input")
	if err = os.WriteFile(cfile, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(cc, cfile, "-o", exe).CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, output)
	}
	r, err := AnalyzeWithOptions("consumer.c", source, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	gofile := filepath.Join(dir, "main.go")
	if err = os.WriteFile(gofile, []byte(r.Draft), 0644); err != nil {
		t.Fatal(err)
	}
	for _, data := range []string{"", "abc\ndefgh\nz", "123456789", "\n\n", "a\x00b\n"} {
		if err = os.WriteFile(input, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
		want, err := exec.Command(exe, input).CombinedOutput()
		if err != nil {
			t.Fatal(err, string(want))
		}
		cmd := exec.Command("go", "run", gofile, input)
		cmd.Env = append(os.Environ(), "GOWORK=off")
		got, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v\n%s\n%s", err, got, r.Draft)
		}
		if string(got) != string(want) {
			t.Fatalf("input %q: Go %q != C %q", data, got, want)
		}
	}
}

func TestStdlibConditionCallsAreNotDefinitions(t *testing.T) {
	draft, changes, err := RewriteStandardLibrary("package p\nfunc f() { if system(command) { }\nif isdigit(c) { } }")
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 || changes[0].Name != "system" || changes[0].Status != "review" || !strings.Contains(draft, "spC >= '0'") {
		t.Fatal(draft, changes)
	}
}

func TestStdlibFloatWidths(t *testing.T) {
	r, err := AnalyzeWithOptions("math.c", `double root(double x) { return sqrt(x); }
float rootf(float x) { return sqrtf(x); }
int main(void) { printf("%.1f %.1f\n", root(9), rootf(16)); return 0; }`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "func root(x float64) float64") || !strings.Contains(r.Draft, "func rootf(x float32) float32") {
		t.Fatal(r.Draft)
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	if err = os.WriteFile(file, []byte(r.Draft), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", file)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil || string(output) != "3.0 4.0\n" {
		t.Fatal(string(output), err, r.Draft)
	}
}

func TestLibraryCatalogUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, rule := range LibraryRules() {
		if seen[rule.Name] || rule.Name == "" || rule.Header == "" || rule.Notes == "" {
			t.Fatal(rule)
		}
		if rule.Mode != "convert" && rule.Mode != "conditional" && rule.Mode != "review" {
			t.Fatal(rule)
		}
		seen[rule.Name] = true
	}
	for _, name := range []string{"fopen", "atoi", "fgets", "sqrtf", "sqrtl", "strtod", "qsort", "feof"} {
		if !seen[name] {
			t.Fatal("missing", name)
		}
	}
}

func TestUnknownSlicesAreNotAssumedBytes(t *testing.T) {
	draft, changes, err := RewriteStandardLibrary("package p\nfunc f() { memcpy(unknown[:], other[:], n); }")
	if err != nil || !strings.Contains(draft, "memcpy(") || len(changes) != 1 || changes[0].Status != "review" {
		t.Fatal(draft, changes, err)
	}
}

func TestLibraryMetadataUsesFileContext(t *testing.T) {
	r, err := Analyze("x.c", "char buffer[8];\nint length(void) { return strlen(buffer); }\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "IndexByte") || !strings.Contains(r.Functions[0].Draft, "IndexByte") {
		t.Fatal(r.Draft, r.Functions)
	}
	if len(r.Library) != 1 || r.Library[0].SourceLine == "" {
		t.Fatal(r.Library)
	}
}

func TestStandardStatusConstants(t *testing.T) {
	draft, _, err := RewriteStandardLibrary("package p\nfunc f() { if fgetc(stdin) == EOF { exit(EXIT_SUCCESS); }; exit(EXIT_FAILURE); }")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(draft, "EXIT_") || strings.Contains(draft, "== EOF") || !strings.Contains(draft, "os.Stdin") {
		t.Fatal(draft)
	}
	kept, _, err := RewriteStandardLibrary("package p\nconst EOF = -3\nfunc f() { use(EOF); }")
	if err != nil || !strings.Contains(kept, "use(EOF)") {
		t.Fatal(kept, err)
	}
}

func TestStdlibNegativeByteConversion(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import ("fmt"; "os")
func main() {
 file, err := os.CreateTemp("", "byte-conversion"); if err != nil { panic(err) }; defer os.Remove(file.Name()); defer file.Close()
 fmt.Println(fputc(-1,file)); rewind(file); fmt.Println(fgetc(file));
 var buffer [2]byte; memset(buffer,-1,2); fmt.Println(buffer[0],buffer[1]);
}`)
	if output != "255\n255\n255 255\n" {
		t.Fatal(output)
	}
}
