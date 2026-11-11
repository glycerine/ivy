package cpp2go

import (
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
	for _, want := range []string{"type s struct", "int a[4];", "unsigned bits:2;", "int x, y;"} {
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
	if r.Draft != "package p\n\n"+s {
		t.Fatal("crossed configuration boundary", r.Draft)
	}
}
