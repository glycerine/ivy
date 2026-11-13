package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSemiformalSourceDiagnostics(t *testing.T) {
	if os.Getenv("CEC_SEMIFORMAL_CHILD") == "1" {
		g := gia.Gia_ManStart(20)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		reg := gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendAnd(g, a, b)
		y := gia.Gia_ManAppendAnd(g, a, reg)
		gia.Gia_ManAppendCo(g, a)
		gia.Gia_ManAppendCo(g, reg)
		gia.Gia_ManSetRegNum(g, 1)
		g.StartRepresentatives()
		gia.Gia_ObjSetRepr(g, y>>1, x>>1)
		gia.Gia_ObjSetNext(g, x>>1, y>>1)
		var pars Cec_ParSmf_t
		Cec_ManSmfSetDefaultParams(&pars)
		pars.nWords = 1
		pars.nFrames = 1
		pars.nRounds = 1
		if os.Getenv("CEC_ZERO_ROUNDS") == "1" {
			pars.nRounds = 0
		}
		pars.nMinOutputs = 1
		Cec_ManSeqSemiformal(g, &pars)
		return
	}
	for _, blocked := range []bool{false, true} {
		dir := t.TempDir()
		if blocked {
			for _, name := range []string{"gore.aig", "gsrm.aig"} {
				if err := os.Mkdir(filepath.Join(dir, name), 0700); err != nil {
					t.Fatal(err)
				}
			}
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestSemiformalSourceDiagnostics$")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "CEC_SEMIFORMAL_CHILD=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("blocked=%v child: %v\n%s", blocked, err, out)
		}
		if blocked {
			for _, name := range []string{"gore.aig", "gsrm.aig"} {
				if !strings.Contains(string(out), "Gia_AigerWrite(): Cannot open the output file \""+name+"\".") {
					t.Fatalf("missing %s diagnostic: %s", name, out)
				}
			}
		}
		if !blocked {
			for _, name := range []string{"gore.aig", "gsrm.aig"} {
				graph, err := gia.Gia_AigerRead(filepath.Join(dir, name), 0, 1, 0)
				if err != nil || graph == nil {
					t.Fatalf("invalid %s: %v", name, err)
				}
				if gia.Gia_ManPiNum(graph) != 2 || name == "gore.aig" && gia.Gia_ManRegNum(graph) != 1 || gia.Gia_ManRegNum(graph) > 1 {
					t.Fatalf("%s interface PI=%d regs=%d", name, gia.Gia_ManPiNum(graph), gia.Gia_ManRegNum(graph))
				}
			}
		}
	}
	if _, err := os.Stat("/dev/full"); err == nil {
		dir := t.TempDir()
		for _, name := range []string{"gore.aig", "gsrm.aig"} {
			if err := os.Symlink("/dev/full", filepath.Join(dir, name)); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestSemiformalSourceDiagnostics$")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "CEC_SEMIFORMAL_CHILD=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("write-failure child: %v %s", err, out)
		}
		if strings.Contains(string(out), "Cannot open the output file") {
			t.Fatalf("write failure misreported as open failure: %s", out)
		}
	}
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestSemiformalSourceDiagnostics$")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CEC_SEMIFORMAL_CHILD=1", "CEC_ZERO_ROUNDS=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("zero-round child: %v %s", err, out)
	}
	for _, name := range []string{"gore.aig", "gsrm.aig"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("zero-round output %s exists: %v", name, err)
		}
	}

}

func TestSemiformalSATConstantOutputs(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, 1)
	info := vec.Vec_PtrAllocSimInfo(2, 1)
	var patterns int32
	status := Cec_ManSatSolveSeq(info, g, nil, 0, &patterns)
	if got := vec.Vec_StrArray(status); len(got) != 2 || got[0] != 1 || got[1] != 0 {
		t.Fatalf("statuses %v", got)
	}
}

func TestSemiformalSequentialPatternTraversal(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	root := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManSetPhase(g)
	p := Cec_ManSatCreate(g, nil)
	if Cec_ManSatCheckNode(p, gia.Gia_ManObj(g, root>>1)) != 0 {
		t.Fatal("expected SAT")
	}
	info := vec.Vec_PtrAllocSimInfo(2, 1)
	gia.Gia_ManIncrementTravId(g)
	Cec_ManSatSolveSeq_rec(p, g, gia.Gia_ManObj(g, root>>1), info, 1, 0)
	if p.nCexLits != 2 || gia.Gia_ObjIsTravIdCurrent(g, gia.Gia_ManObj(g, a>>1)) == 0 || gia.Gia_ObjIsTravIdCurrent(g, gia.Gia_ManObj(g, b>>1)) == 0 {
		t.Fatal("missing CI traversal state")
	}
	Cec_ManSatSolveSeq_rec(p, g, gia.Gia_ManObj(g, root>>1), info, 1, 0)
	if p.nCexLits != 2 {
		t.Fatal("revisited assigned cone")
	}
}
