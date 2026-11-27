// Package randaudit guards the ports whose random draws were matched to
// ABC's generators (glibc rand(), Abc_Random, Aig_ManRandom, ...): none of
// them may go back to Go's math/rand.
package randaudit

import (
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

var auditedFiles = []string{
	"../../misc/vec/vecWrd.go",
	"../../proof/ssw/sswRarity.go",
	"../../aig/saig/saigSwitch.go",
	"../../aig/saig/saigConstr2.go",
	"../../aig/saig/saigSynchState.go",
	"../../base/abci/abcBm.go",
	"../../aig/gia/sim_quality_improve.go",
	"../../aig/gia/sim_rare_improve.go",
	"../../aig/ivy/ivyFraigSim.go",
	"../../base/abc/abcNtk.go",
	"../../base/abci/abc.go",
	"../../base/abci/abcCascade.go",
	"../../base/abci/abcStrash.go",
	"../../opt/sim/simUtils.go",
	"../../aig/gia/rex_walk.go",
	"../../bdd/reo/reoShuffle.go",
	"../../aig/saig/saigStrSim.go",
}

func TestAuditedPortsDoNotUseMathRand(t *testing.T) {
	for _, file := range auditedFiles {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if path == "math/rand" || path == "math/rand/v2" {
				t.Errorf("%s imports %s; ABC's generator must be used instead", file, path)
			}
		}
	}
}
