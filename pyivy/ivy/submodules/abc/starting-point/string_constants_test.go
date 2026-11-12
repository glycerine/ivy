package startingpoint

import (
	"strings"
	"testing"
)

func TestStringDefines(t *testing.T) {
	r := checkedDraft(t, `#define x "some string"
#define EMPTY ""
#define ESCAPED "quote: \"; slash: \\; newline: \n"
#define NUMBER_TEXT "1.15u"
#define JOIN "first" /* between */ "second" // tail
#define NOTE /* before */ "value" /* after */
`, Options{Package: "main"})
	for _, want := range []string{
		`const x = "some string"`, `const EMPTY = ""`,
		`const NUMBER_TEXT = "1.15u"`,
		`const JOIN = "first" + /* between */ "second" // tail`,
		`const NOTE = /* before */ "value" /* after */`,
	} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	output, _ := runLibraryProgram(t, r.Draft+`func main() { println(x, JOIN, len(EMPTY), NUMBER_TEXT) }
`)
	if output != "some string firstsecond 0 1.15u\n" {
		t.Fatal(output)
	}
}

func TestStringDefineBoundaries(t *testing.T) {
	source := `#define FUNCTION(x) "text"
#define EXPR "text" + suffix
#define CHAR 'x'
#define WIDE L"text"
// #define COMMENTED "text"
`
	got, err := integerDefines(source)
	if err != nil || got != source {
		t.Fatal(got, err)
	}
	r, err := Analyze("x.c", `#if UNKNOWN
#define OPAQUE "unchanged"
#endif
void f(void) {
#define LOCAL "local value"
}
`)
	if err != nil || !strings.Contains(r.Draft, `#define OPAQUE "unchanged"`) || !strings.Contains(r.Draft, `const LOCAL = "local value"`) || !strings.Contains(r.Functions[0].Draft, `const LOCAL = "local value"`) {
		t.Fatal(r.Draft, r.Functions, err)
	}
}
