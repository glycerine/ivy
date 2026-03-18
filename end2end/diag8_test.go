package end2end

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/isolate"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"
	"github.com/glycerine/goivy/ast"
)

func TestDiagnostic8(t *testing.T) {
	data, _ := os.ReadFile(filepath.Join(dataDir(), "client_server.ivy"))
	src := string(data)
	version := lexer.Version{1, 7}
	p := parser.New(src, version)
	decls, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	mod := module.New()
	mod.Sig = il.NewSig()
	topCtx := compiler.CollectActions(decls)
	c := compiler.NewFromModule(mod)
	c.TopCtx = topCtx

	ds := compiler.NewDomainSetup(c)
	if err := ds.ProcessDecls(decls); err != nil {
		t.Fatalf("pass 1: %v", err)
	}

	fmt.Printf("After pass 1:\n")
	fmt.Printf("  Actions: %d\n", len(mod.Actions))
	for k, v := range mod.Actions {
		_, isAct := v.(actions.Action)
		fmt.Printf("    %q: %T, Action=%v\n", k, v, isAct)
	}
	fmt.Printf("  LabeledConjs: %d\n", len(mod.LabeledConjs))
	fmt.Printf("  Exports: %d\n", len(mod.Exports))
	fmt.Printf("  Mixins: %d\n", len(mod.Mixins))
	for k, v := range mod.Mixins {
		fmt.Printf("    %q: %d\n", k, len(v))
	}

	// Now simulate CreateIsolate
	compiler.FixConstructors(mod)
	cs := compiler.NewConjSetup(c)
	cs.ProcessDecls(decls)
	as := compiler.NewARGSetup(c)
	as.ProcessDecls(decls)
	compiler.CreateSortOrder(mod)
	compiler.CreateConstructorSchemata(mod)
	compiler.AttachProofs(mod)
	compiler.CheckDefinitions(mod)
	compiler.CheckPropertiesPass(mod)
	compiler.CreateConjActions(mod)
	compiler.HandleTemporals(mod)

	if _, ok := mod.Isolates["this"]; !ok {
		isol := &ast.IsolateDef{
			Elems:    []ast.Node{ast.NewAtom("this"), ast.NewAtom("this")},
			WithArgs: 0,
		}
		mod.Isolates["this"] = isol
	}
	
	fmt.Printf("\nBefore CreateIsolate:\n")
	fmt.Printf("  Actions: %d\n", len(mod.Actions))
	for k := range mod.Actions {
		fmt.Printf("    %q\n", k)
	}

	if err := isolate.CreateIsolate("this", mod); err != nil {
		fmt.Printf("CreateIsolate error: %v\n", err)
	}

	fmt.Printf("\nAfter CreateIsolate:\n")
	fmt.Printf("  Actions: %d\n", len(mod.Actions))
	for k := range mod.Actions {
		fmt.Printf("    %q\n", k)
	}
}
