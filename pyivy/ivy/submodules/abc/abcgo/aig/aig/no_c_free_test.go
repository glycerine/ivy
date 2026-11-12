package aig

import (
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Scan lexically rather than searching source bytes: retained source comments
// and quoted original snippets are useful documentation and are not live calls.
func liveCFreeCalls(name string, source []byte) []token.Position {
	set := token.NewFileSet()
	file := set.AddFile(name, -1, len(source))
	var lex scanner.Scanner
	lex.Init(file, source, func(token.Position, string) {}, scanner.ScanComments)
	var previous token.Token
	var word string
	var position token.Pos
	var found []token.Position
	for {
		pos, kind, text := lex.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.COMMENT {
			continue
		}
		if kind == token.LPAREN && previous == token.IDENT && (word == "free" || word == "ABC_FREE") {
			found = append(found, set.Position(position))
		}
		previous, word, position = kind, text, pos
	}
	return found
}
func TestNoLiveCFreeCalls(t *testing.T) {
	_, source, _, _ := runtime.Caller(0)
	directory := filepath.Dir(source)
	files, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, position := range liveCFreeCalls(name, data) {
			t.Errorf("live C allocation cleanup call remains: %s", position)
		}
	}
}
func TestCFreeRegressionScanProtectsTrivia(t *testing.T) {
	source := `// free(p); ABC_FREE(p);
/* free(p); */
var text = "free(p); ABC_FREE(p)"
var raw = ` + "`free(p);`" + `
free /* keep */ (p);
ABC_FREE(
p);
`
	calls := liveCFreeCalls("fixture.go", []byte(source))
	if len(calls) != 2 {
		t.Fatal(calls)
	}
	if !strings.Contains(calls[0].String(), "fixture.go:5:") || !strings.Contains(calls[1].String(), "fixture.go:6:") {
		t.Fatal(calls)
	}
}
