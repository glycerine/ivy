package startingpoint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMechanicalControlFlowTogether(t *testing.T) {
	source := `enum { FIRST, SECOND, THIRD };
int f(int n) {
 int result = 0;
 switch(n) {
 case FIRST:
 case SECOND: result += 2;
 case THIRD: result += 3; break;
 default: result = 9;
 }
 do { result--; } while(result > 5);
 for(1) { break; }
 if (++result == 4) result += 10;
 return (int)result;
}
int main(void) { printf("%d %d %d\n",f(0),f(2),f(3)); return 0; }
`
	r, err := AnalyzeWithOptions("x.c", source, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "5 3 6\n" {
		t.Fatal(output, r.Draft)
	}
	if strings.Contains(r.Draft, "do {") || strings.Contains(r.Draft, "switch(") || strings.Contains(r.Draft, "++result") {
		t.Fatal(r.Draft)
	}
}

func TestDoContinueRetainedAndComments(t *testing.T) {
	source := "do { if (x) continue; x++; } while (x < 4);"
	got, err := lowerDoWhile(source)
	if err != nil || got != source {
		t.Fatal(got, err)
	}
	source = `// p->field and for 1
func macro() { p->field = 1; }
var s = "p->field"`
	got, err = mechanicalDraft(source)
	if err != nil || !strings.Contains(got, "p.field = 1") || !strings.Contains(got, `"p->field"`) || !strings.Contains(got, "// p->field") {
		t.Fatal(got, err)
	}
}

func TestSharedTypedefIndex(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	out := filepath.Join(root, "out")
	putSource(t, src, "types.h", "typedef long abctime;\ntypedef struct { int x; } widget;\n")
	putSource(t, src, "consumer.c", "void f(void) { abctime clk; widget *p; x * y; }")
	_, err := TranslateTree(src, out, Options{})
	if err != nil {
		t.Fatal(err)
	}
	draft, err := os.ReadFile(filepath.Join(out, "consumer.go"))
	for _, want := range []string{"var clk abctime;", "var p *widget;", "x * y;"} {
		if err != nil || !strings.Contains(string(draft), want) {
			t.Fatal(string(draft), err)
		}
	}
}

func TestTruthAndCommaRuntime(t *testing.T) {
	source := `int f(int n, int *p) {
 int i, j, total=0;
 if (n && p) total = 1;
 if (!n) total = 2;
 for(i=0,j=0; i<3; i++,j++) total += i+j;
 (total += 1, total += 2);
 return total;
}
int main(void) { int value=0; printf("%d %d\n",f(1,&value),f(0,NULL)); return 0; }
`
	r, err := AnalyzeWithOptions("x.c", source, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "10 11\n" {
		t.Fatal(output, r.Draft)
	}
	// Dependent assignments cannot become simultaneous Go assignments.
	got, ok := independentAssignments("i=1,j=i")
	if ok || got != "i=1,j=i" {
		t.Fatal(got, ok)
	}
}

func TestSingleObjectAllocation(t *testing.T) {
	source := `typedef struct node node;
struct node { int value; };
node *create(void) { node *p; p=ABC_CALLOC(node,1); return p; }
int main(void) { node *p; p=create(); printf("%d\n",p->value); return 0; }
`
	r, err := AnalyzeWithOptions("x.c", source, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "0\n" || !strings.Contains(r.Draft, "new(node)") {
		t.Fatal(output, r.Draft)
	}
}

func TestStaticStorageAggregateAndBitfield(t *testing.T) {
	source := `typedef struct item item;
struct item { unsigned int value : 3; };
int next(void) { static int counter = 0; counter++; return counter; }
int f(void) { static int values[3] = {1,2,3}; return values[0]+values[2]; }
int main(void) { printf("%d %d %d\n",next(),next(),f()); return 0; }
`
	r, err := AnalyzeWithOptions("x.c", source, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "1 2 4\n" || !strings.Contains(r.Draft, "value uint32;") || !strings.Contains(r.Draft, "C bitfield: 3 bits") {
		t.Fatal(output, r.Draft)
	}
}

func TestWholeStructReset(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
import "fmt"
type item struct { value int }
func clear(p *item) { memset(p,0,sizeof(item)); }
func main() { p:=new(item);p.value=9;clear(p);fmt.Println(p.value) }
`)
	if output != "0\n" {
		t.Fatal(output)
	}
	draft, _, err := RewriteStandardLibrary("package p\nfunc clear(p *unknown) { memset(p,0,sizeof(unknown)); }")
	if err != nil || !strings.Contains(draft, "memset(p,0,sizeof(unknown))") {
		t.Fatal(draft, err)
	}
}

func TestExpandedExpressionMacroAndFields(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `#define ADD(a,b) ((a)+(b))
typedef struct item item;
struct item { int data[2]; int x,y; };
int main(void) { item *p; p=ABC_CALLOC(item,1); p->data[1]=ADD(2,3); printf("%d\n",p->data[1]); return 0; }
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "5\n" || !strings.Contains(r.Draft, "((2)+(3))") {
		t.Fatal(output, r.Draft)
	}
}

func TestStaticMetadataCollision(t *testing.T) {
	r, err := Analyze("x.c", "int spStatic_f_count;\nint f(void) { static int count=0; count++; return count; }\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Functions) != 1 || !strings.Contains(r.Draft, "spStatic_f_count_") || !strings.Contains(r.Functions[0].Draft, "spStatic_f_count_") {
		t.Fatal(r.Draft, r.Functions)
	}
}

func TestLocalEnumAndTagReferences(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `enum tag { A,B };
typedef enum tag alias;
int f(enum tag value) { enum { X,Y }; return (int)value + Y; }
int main(void) { printf("%d\n",f(B)); return 0; }
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "2\n" || !strings.Contains(r.Draft, "type alias") {
		t.Fatal(output, r.Draft)
	}
}

func TestEnumIntegerSuffixes(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `enum { A=1u, B, C=0x10UL, D };
int main(void) { printf("%d %d %d %d\n",A,B,C,D); return 0; }
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "1 2 16 17\n" {
		t.Fatal(output, r.Draft)
	}
}

func TestStaticParameterShadowRetained(t *testing.T) {
	source := "func f(count int32) { { static int count=0; count++; } use(count); }"
	got, err := liftStaticLocals(source)
	if err != nil || got != source {
		t.Fatal(got, err)
	}
}

func TestSwitchNestedBreakAndFallthrough(t *testing.T) {
	source := `int f(int x) { int result=0; switch(x) {
 case 0: if (x==0) { result=7; break; } result=9;
 case 1: result+=3; break;
 default: result=4;
 } return result; }
int main(void) { printf("%d %d %d\n",f(0),f(1),f(2)); return 0; }
`
	r, err := AnalyzeWithOptions("x.c", source, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "7 3 4\n" {
		t.Fatal(output, r.Draft)
	}
}

func TestPrefixIncrementOperandBoundary(t *testing.T) {
	for _, source := range []string{"if ++p.field == 2 {}", "if ++a[i] == 2 {}", "if + + n {}"} {
		got, err := mechanicalSyntax(source)
		if err != nil || got != source {
			t.Fatal(got, err)
		}
	}
}

func TestConfiguredTreeFeatureInteraction(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	out := filepath.Join(root, "out")
	putSource(t, src, "types.h", `#ifndef TYPES_H
#define TYPES_H
#if WIN32
typedef int WindowsOnly;
#else
typedef struct record record;
struct record { unsigned flag:1; int a[2]; };
enum { LOW,HIGH };
#endif
#endif
`)
	putSource(t, src, "consumer.c", `record *create(void) { record *p; p=ABC_ALLOC(record,1); return p; }
int calc(int n) {
 static int calls=0;
 record *p; int i,j;
 p=create(); p->flag=1;
 assert(p);
 for(i=0,j=0;i<2;i++,j++) p->a[i]=j;
 switch(n) { case LOW: case HIGH: n+=2; case 2: n+=3; break; default: n=0; }
 calls++;
 return (int)(calls+n+p->a[1]);
}
int main(void) { printf("%d %d\n",calc(0),calc(1)); return 0; }
`)
	report, err := TranslateTree(src, out, Options{Package: "main", Defines: map[string]int64{"WIN32": 0}})
	if err != nil || report.Fallbacks != 0 || len(report.HelperFiles) != 1 {
		t.Fatal(report, err)
	}
	files, err := filepath.Glob(filepath.Join(out, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", append([]string{"run"}, files...)...)
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil || string(output) != "7 9\n" {
		t.Fatal(string(output), err)
	}
}

func TestDynamicLocalStaticRemainsLazy(t *testing.T) {
	for _, source := range []string{
		"func f() int32 { static int count=next(); return count; }",
		"func f(n int32) int32 { static int count=n; return count; }",
		"func f(n int32) { static int *p = &n; use(p); }",
	} {
		got, err := liftStaticLocals(source)
		if err != nil || got != source {
			t.Fatal(got, err)
		}
	}
}
