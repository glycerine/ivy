package startingpoint

import "testing"

func TestCommentIncludesAndMarkers(t *testing.T) {
	source := "#include <assert.h>\n#include <stdio.h>\n#include <string.h>\n\n#include \"../mem.h\"\n\n#include \"misc/util/abc_global.h\"\nABC_NAMESPACE_HEADER_START\n"
	want := "//#include <assert.h>\n//#include <stdio.h>\n//#include <string.h>\n\n//#include \"../mem.h\"\n\n//#include \"misc/util/abc_global.h\"\n//ABC_NAMESPACE_HEADER_START\n"
	got, err := commentScaffolding(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	twice, err := commentScaffolding(got)
	if err != nil || twice != got {
		t.Fatal("not idempotent", twice, err)
	}
}

func TestScaffoldingPreservesCodeAndComments(t *testing.T) {
	source := `// existing comment
/* ABC_NAMESPACE_HEADER_START */
static inline
int
f() {
BODY_MACRO
return 1;
}
unknown_return_type
other();
type known struct {}
known
value;
`
	got, err := commentScaffolding(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != source {
		t.Fatal(got)
	}
}

func TestContinuedInclude(t *testing.T) {
	source := "# include \\\n \"x.h\"\n"
	got, err := commentScaffolding(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != "//# include \\\n// \"x.h\"\n" {
		t.Fatal(got)
	}
}
