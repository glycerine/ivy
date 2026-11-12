package gia

import (
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNoLiveCFreeCallsInGia(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(file), "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		f := fset.AddFile(path, fset.Base(), len(source))
		var scan scanner.Scanner
		// Remaining generated C syntax is deliberately tolerated: this audit checks
		// executable tokens for frees, never matches comments or string literals.
		scan.Init(f, source, func(token.Position, string) {}, 0)
		previous := token.ILLEGAL
		type scanned struct {
			kind token.Token
			text string
		}
		var recent []scanned
		for {
			pos, tok, lit := scan.Scan()
			if tok == token.EOF {
				break
			}
			if tok == token.IDENT && (lit == "ABC_FREE" || lit == "free") {
				t.Errorf("C free token remains at %s", fset.Position(pos))
			}
			if previous == token.PERIOD && tok == token.IDENT && lit == "pRefs" {
				t.Errorf("legacy manager ref spelling at %s", fset.Position(pos))
			}
			if tok == token.IDENT && lit == "nil" && len(recent) >= 3 {
				selector, field, operator := recent[len(recent)-3], recent[len(recent)-2], recent[len(recent)-1]
				if selector.kind == token.PERIOD && field.kind == token.IDENT && (field.text == "pName" || field.text == "pSpec" || field.text == "pCellStr") && (operator.kind == token.ASSIGN || operator.kind == token.EQL || operator.kind == token.NEQ) {
					t.Errorf("string field used with nil at %s", fset.Position(pos))
				}
			}
			recent = append(recent, scanned{tok, lit})
			if len(recent) > 3 {
				recent = recent[1:]
			}
			previous = tok
		}
	}
}
