package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestSingleObjectCalloc(t *testing.T) {
	r, err := Analyze("x.c", `void f(void) { p = ((Xyz_ManTry_t *) calloc((size_t)(1), sizeof(Xyz_ManTry_t))); }`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "p = &Xyz_ManTry_t{};") {
		t.Fatal(r.Draft)
	}
}

func TestSingleObjectCallocForms(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"((T*)calloc(1,sizeof(T)))", "&T{}"},
		{"(T *) calloc((size_t)(((1))), sizeof(T))", "&T{}"},
		{"(((T*)calloc(1UL,(sizeof(T)))))", "&T{}"},
		{"((T*)calloc(1,sizeof(T)))->field", "(&T{})->field"},
		{"((T*)calloc(1,sizeof(T)) + offset)", "(&T{} + offset)"},
		{"((int*)calloc(1,sizeof(int)))", "new(int32)"},
		{"((T*)calloc(2,sizeof(T)))", ""},
		{"((T*)calloc(n,sizeof(T)))", ""},
		{"((T*)calloc(next(),sizeof(T)))", ""},
		{"((T*)calloc(0,sizeof(T)))", ""},
		{"((T*)calloc(1,sizeof(U)))", ""},
		{"((T**)calloc(1,sizeof(T*)))", ""},
		{"((T*)malloc(sizeof(T)))", ""},
		{"calloc(1,sizeof(T))", ""},
		{"((T*)calloc(1,sizeof(T),extra))", ""},
		{`"((T*)calloc(1,sizeof(T)))" /* ((T*)calloc(1,sizeof(T))) */`, ""},
	} {
		got, err := lowerSingleObjectCalloc(tc.source)
		if err != nil {
			t.Fatal(err)
		}
		want := tc.want
		if want == "" {
			want = tc.source
		}
		if got != want {
			t.Fatalf("%s -> %q, want %q", tc.source, got, want)
		}
	}
	got, err := lowerSingleObjectCalloc("((T /* cast */ *) calloc((size_t)(1), sizeof(/* size */ T)))")
	if err != nil || !strings.Contains(got, "&T{}") || strings.Count(got, "/* cast */") != 1 || strings.Count(got, "/* size */") != 1 {
		t.Fatal(err, got)
	}
}

func TestExpandedCallocRuntime(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `#define ABC_CALLOC(type,num) ((type *) calloc((size_t)(num), sizeof(type)))
typedef struct Item Item;
struct Item { int n; };
Item * make_item(void) { Item *p; p = ABC_CALLOC(Item,1); return p; }
int main(void) { Item *p; p = make_item(); printf("%d",p->n); return 0; }
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "p := &Item{};") || !strings.Contains(r.Draft, "// inlining of macro ABC_CALLOC") {
		t.Fatal(r.Draft)
	}
	if len(r.Functions) == 0 || !strings.Contains(r.Functions[0].Draft, "&Item{}") {
		t.Fatal("missing function metadata")
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "0" {
		t.Fatal(output, r.Draft)
	}
}

func TestTemplateSingleObjectCalloc(t *testing.T) {
	path := "../src/template.c"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r, err := AnalyzeWithOptions(path, string(source), Options{IncludePaths: []string{"../src"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "p := &Xyz_ManTry_t{};") {
		t.Fatal(r.Draft)
	}
}
