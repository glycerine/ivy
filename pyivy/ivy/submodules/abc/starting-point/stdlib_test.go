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

func TestStdlibFormatAndFailures(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import "fmt"
func main() {
 printf("%ld %u %zu\n", int64(2), uint32(3), uint64(4));
 printf("x");
}`)
	if output != "2 3 4\nx" {
		t.Fatal(output)
	}
	draft, changes, err := RewriteStandardLibrary("package p\nfunc f() { printf(\"%n\", &n); }\n")
	if err != nil || !strings.Contains(draft, "printf(") || len(changes) != 1 || changes[0].Status != "review" || strings.Contains(draft, "import ") {
		t.Fatal(draft, changes, err)
	}
}

func TestStdlibConditionCallsAreNotDefinitions(t *testing.T) {
	draft, changes, err := RewriteStandardLibrary("package p\nfunc f() { if system(command) { }\nif isdigit(c) { } }")
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 2 || changes[0].Name != "system" || changes[0].Status != "review" || !strings.Contains(draft, "isdigit(c)") {
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
	if !strings.Contains(r.Draft, "strlen(buffer)") || !strings.Contains(r.Functions[0].Draft, "strlen(buffer)") {
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

// Former runtime-adapter fixtures now leave semantic repairs to the next pass.
func TestStdlibNeverIntroducesClosures(t *testing.T) {
	fixtures := []string{
		`package main
import "fmt"
func main() {
 for _, text := range []string{"  +12junk", "-17tail", "abc", "+", "2147483647", "\t42\n"} { fmt.Println(atoi(text)); }
 fmt.Println(atoll("-9223372036854775807rest"));
 fmt.Println(abs(-4), isdigit('8'), isalpha('!'), isspace('\n'), tolower('A'), toupper(-1));
 fmt.Println(sqrt(9), pow(2,3), fmin(0.0/1.0, 4));
}`,
		`package main
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
}`,
		`package main
import "fmt"
func main() {
 var a [8]byte; var b [8]byte
 memset(a, 'x', 4); memcpy(b, a, 4);
 fmt.Println(memcmp(a,b,4), strlen(a));
 var text string = "hi"
 strcat(text, "!"); fmt.Println(text, strcmp(text,"hi!"), strncmp("abcd","abxy",2));
}`,
		`package main
import (f "fmt"; st "strconv")
func main() {
 _ = st.IntSize
 strings := "local"
 f.Println(strings, abs(atoi("-12foo")))
}`,
		`package main
import "fmt"
var calls int
func next() string { calls++; return " -4x" }
func main() { fmt.Println(abs(atoi(next())), calls); }
`,
		`package main
import ("fmt"; "os")
func main() {
 file, err := os.CreateTemp("", "byte-conversion"); if err != nil { panic(err) }; defer os.Remove(file.Name()); defer file.Close()
 fmt.Println(fputc(-1,file)); rewind(file); fmt.Println(fgetc(file));
 var buffer [2]byte; memset(buffer,-1,2); fmt.Println(buffer[0],buffer[1]);
}`,
	}
	for _, source := range fixtures {
		draft, _, err := RewriteStandardLibrary(source)
		if err != nil {
			t.Fatal(err)
		}
		if anonymousFunctions(draft) != 0 {
			t.Fatalf("introduced closure: %s", draft)
		}
	}
}
func TestPrintfExpressionRemainsDirect(t *testing.T) {
	source := `package p
func ABC_PRT(a int, t int) { (printf("%s = ", a), printf("%d", t)) }
func count() { n := printf("x"); use(n) }
`
	draft, _, err := RewriteStandardLibrary(source)
	if err != nil {
		t.Fatal(err)
	}
	if anonymousFunctions(draft) != 0 || strings.Count(draft, "fmt.Printf(") != 3 || !strings.Contains(draft, `n := fmt.Printf("x")`) {
		t.Fatal(draft)
	}
}
