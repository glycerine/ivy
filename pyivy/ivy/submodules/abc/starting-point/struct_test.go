package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestTypedefStructPair(t *testing.T) {
	source := `typedef struct heap_t_ heap_t;
struct heap_t_ {
    vec_int_t *indices;
    vec_uint_t *data;
    vec_act_t *weights;
};`
	r, err := Analyze("heap/heap.h", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"type heap_t struct {", "    indices *vec_int_t;\n", "    data *vec_uint_t;\n", "    weights *vec_act_t;\n"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal(r.Draft)
		}
	}
	if strings.Contains(r.Draft, "typedef") || strings.Contains(r.Draft, "heap_t_") {
		t.Fatal(r.Draft)
	}
	checkedDraft(t, source+"\ntype vec_int_t struct {}\ntype vec_uint_t struct {}\ntype vec_act_t struct {}\n", Options{Package: "p"})
}

func TestStructCommentsAndSelfPointer(t *testing.T) {
	checkedDraft(t, `typedef /* alias note */ struct node_ node;
// between declarations
struct /* tag note */ node_ {
    unsigned*data; // data comment
    int /* count note */ count;
    struct node_ *next;
};`, Options{Package: "p"})
}

func TestComplexStructFieldsRemainVisible(t *testing.T) {
	s := `typedef struct s_ s;
struct s_ { int a[4]; unsigned bits:2; int x, y; };`
	r, err := Analyze("p/file.h", s)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"type s struct", "a [4]int32;", "bits uint32;", "x, y int32;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal(r.Draft)
		}
	}
	count := 0
	for _, issue := range r.Issues {
		if issue.Category == "struct-field-review" {
			count++
		}
	}
	if count != 3 {
		t.Fatal(r.Issues)
	}
}

func TestStructPairDoesNotCrossDirective(t *testing.T) {
	s := "typedef struct s_ s;\n#if FLAG\nstruct s_ { int value; };\n#endif\n"
	r, err := Analyze("p/file.h", s)
	if err != nil {
		t.Fatal(err)
	}
	if r.Draft != "package p\n\ntype s = s_;\n#if FLAG\nstruct s_ { int value; };\n#endif\n" {
		t.Fatal("crossed configuration boundary", r.Draft)
	}
}

func TestStandaloneStructAndSeparateAlias(t *testing.T) {
	checkedDraft(t, `typedef struct node_ node;
 var unrelated int32
 struct node_
 {
  // fields
  int count; // count note
  struct node_ *next;
  int **chunks;
 };`, Options{Package: "p"})
}

func TestAigMemoryStructs(t *testing.T) {
	source, err := os.ReadFile("../src/aig/aig/aigMem.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("aigMem.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"type Aig_MmFixed_t_ struct {", "nEntrySize", "*string", "// the size of one entry", "// the allocated memory"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q", want)
		}
	}
	tokens, _ := Lex(string(source))
	for _, tok := range tokens {
		if tok.Kind == "comment" && !strings.Contains(r.Draft, tok.Text) {
			t.Fatalf("lost comment %q", tok.Text)
		}
	}
	if strings.Contains(r.Draft, "int           nEntrySize") {
		t.Fatal("untranslated field")
	}
	alias, err := Analyze("aig.h", "typedef struct Aig_MmFixed_t_ Aig_MmFixed_t;")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(alias.Draft, "type Aig_MmFixed_t = Aig_MmFixed_t_;") {
		t.Fatal(alias.Draft)
	}
}

func TestPairedStructBracePlacement(t *testing.T) {
	checkedDraft(t, "typedef struct pair_ pair;\nstruct pair_\n{\n int value;\n};", Options{Package: "p"})
	checkedDraft(t, "typedef struct pair_ pair;\nstruct pair_ // keep header comment\n{\n int value;\n};", Options{Package: "p"})
	checkedDraft(t, "struct lone_ // keep standalone comment\n{\n int value;\n};", Options{Package: "p"})
	source, err := os.ReadFile("../src/template.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("template.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Xyz_ParTry_t", "Xyz_ManTry_t"} {
		if !strings.Contains(r.Draft, "type "+name+" struct {") {
			t.Fatalf("brace not joined for %s", name)
		}
	}
}
