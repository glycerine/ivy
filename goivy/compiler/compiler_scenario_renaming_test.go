package compiler

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/lexer"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/parser"
)

const reusedScenarioTransitionSource = `#lang ivy1.7

type t
relation p(X:t)

action a(x:t) = {
}

export a

scenario {
    -> s0;
    s0 -> s1 : before a(x:t) {
        assume p(x)
    }
    s1 -> s0 : before a(y:t) {
        assume p(y)
    }
}
`

func TestScenarioReusedTransitionRenamesLaterFormals(t *testing.T) {
	result, err := parser.Parse(reusedScenarioTransitionSource, lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("parse Go Ivy source: %v", err)
	}

	mod := module.New()
	mod.Cfg = module.NewConfig()
	if err := IvyCompile(result.Decls, mod, false); err != nil {
		t.Fatalf("IvyCompile: %v", err)
	}

	var mixerName string
	for name := range mod.Actions.All() {
		if strings.HasPrefix(name, "a[before") {
			mixerName = name
			break
		}
	}
	if mixerName == "" {
		t.Fatalf("scenario did not generate a before mixer action; actions=%v", actionNames(mod))
	}

	mixer, _ := mod.Actions.Get2(mixerName)
	got := fmt.Sprint(mixer)

	if strings.Contains(got, "fml:y") {
		t.Fatalf("scenario branch was not renamed to first transition formals; generated mixer %s contains second-branch formal y:\n%s", mixerName, got)
	}
	if !strings.Contains(got, "fml:x") {
		t.Fatalf("scenario mixer %s does not contain expected first-branch formal x:\n%s", mixerName, got)
	}
}

func actionNames(mod *module.Module) []string {
	names := make([]string, 0, mod.Actions.Len())
	for name := range mod.Actions.All() {
		names = append(names, name)
	}
	return names
}
