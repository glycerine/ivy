package abci

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/proof/pdr"
	"github.com/glycerine/abcgo/sat/bmc"
)

func TestProversExhaustiveTwoInputFunctions(t *testing.T) {
	for _, engine := range []string{"SAT", "miter", "Ivy", "ACB"} {
		for truth := uint64(0); truth < 16; truth++ {
			t.Run(fmt.Sprintf("%s/%x", engine, truth), func(t *testing.T) {
				n := truthNetwork(t, 2, truth)
				defer func() { abc.Abc_NtkDelete(n) }()
				var status int32
				var parameters Prove_Params_t
				Prove_ParamsSetDefault(&parameters)
				switch engine {
				case "SAT":
					status = Abc_NtkMiterSat(n, 0, 0, 0, nil, nil)
				case "miter":
					status = Abc_NtkMiterProve(&n, &parameters)
				case "Ivy":
					status = Abc_NtkIvyProve(&n, &parameters)
				case "ACB":
					g := Abc_NtkToDar(n, 0, 0)
					p := gia.Gia_ManFromAigSimple(g)
					witness := acbProveMiter(p)
					gia.Gia_ManStop(p)
					aig.Aig_ManStop(g)
					if (witness != nil) != (truth != 0) {
						t.Fatalf("truth=%x witness=%v", truth, witness)
					}
					return
				}
				want := int32(0)
				if truth == 0 {
					want = 1
				}
				if status != want {
					t.Fatalf("status=%d want=%d", status, want)
				}
				if status == 0 {
					model := abc.Abc_NtkReadModel(n)
					if len(model) != 2 {
						t.Fatalf("model=%v", model)
					}
					bits := int(model[0] | model[1]<<1)
					if bits < 0 || bits > 3 || truth>>bits&1 == 0 {
						t.Fatalf("invalid witness %v for %x", model, truth)
					}
				}
			})
		}
	}
}

func TestIvyProveReturnsReplacementForLogicInput(t *testing.T) {
	for _, truth := range []uint64{0, 1, 8, 15} {
		n := truthNetwork(t, 2, truth)
		logic := Abc_NtkMulti(n, 0, 100, 1, 0, 0, 0)
		abc.Abc_NtkDelete(n)
		if logic == nil {
			t.Fatal("logic conversion failed")
		}
		var parameters Prove_Params_t
		Prove_ParamsSetDefault(&parameters)
		status := Abc_NtkIvyProve(&logic, &parameters)
		if abc.Abc_NtkIsStrash(logic) == 0 || len(logic.Inputs()) != 2 || len(logic.Outputs()) != 1 {
			t.Fatalf("prover returned deleted or unconverted network for truth %x", truth)
		}
		if (status == 1) != (truth == 0) {
			t.Fatalf("truth=%x status=%d", truth, status)
		}
		if status == 0 && len(abc.Abc_NtkReadModel(logic)) != 2 {
			t.Fatal("witness missing from returned network")
		}
		abc.Abc_NtkDelete(logic)
	}
}

func sequentialMiter(t *testing.T, bad bool) *abc.Network {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	n.Name = "sequential"
	pi := abc.Abc_NtkCreatePi(n)
	pi.SetName("input")
	state := abc.Abc_NtkAddLatch(n, nil, abc.ABC_INIT_ZERO)
	latch := abc.Abc_ObjFanin0(state)
	li := abc.Abc_ObjFanin0(latch)
	state.SetName("state")
	latch.SetName("reg")
	li.SetName("next")
	next := n.One().Not()
	if bad {
		next = pi.Signal()
	}
	abc.Abc_ObjAddFanin(li, next)
	po := abc.Abc_NtkCreatePo(n)
	po.SetName("bad")
	abc.Abc_ObjAddFanin(po, state)
	abc.Abc_NtkOrderCisCos(n)
	return n
}

func TestSequentialVerificationAdaptersPreserveWitnesses(t *testing.T) {
	for _, engine := range []string{"bmc", "bmc2", "bmc3", "pdr"} {
		for _, bad := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/bad=%v", engine, bad), func(t *testing.T) {
				n := sequentialMiter(t, bad)
				defer abc.Abc_NtkDelete(n)
				var status int32
				switch engine {
				case "bmc", "bmc2":
					algorithm := int32(0)
					if engine == "bmc" {
						algorithm = 1
					}
					var frames int32
					status = Abc_NtkDarBmc(n, 0, 4, 100000, 2000, 0, 100000, 1000000, 0, algorithm, 0, 0, 0, &frames, 0)
				case "bmc3":
					var p bmc.Saig_ParBmc_t
					bmc.Saig_ParBmcSetDefaultParams(&p)
					p.NFramesMax, p.FSilent = 4, 1
					status = Abc_NtkDarBmc3(n, &p, 0)
				case "pdr":
					var p pdr.Pdr_Par_t
					pdr.Pdr_ManSetDefaultParams(&p)
					p.FSilent = 1
					status = Abc_NtkDarPdr(n, &p)
				}
				if bad && status != 0 || !bad && status == 0 {
					t.Fatalf("bad=%v status=%d", bad, status)
				}
				if !bad && (engine == "pdr" || engine == "bmc3") && status != 1 {
					t.Fatalf("safe finite-state property not proved: %d", status)
				}
				cex := abc.Abc_NtkReadSeqModel(n)
				if !bad {
					if cex != nil {
						t.Fatal("spurious counterexample")
					}
					return
				}
				if cex == nil || cex.IFrame != 1 || cex.IPo != 0 {
					t.Fatalf("bad witness metadata: %+v", cex)
				}
				g := Abc_NtkToDar(n, 0, 1)
				defer aig.Aig_ManStop(g)
				if saig.Saig_ManVerifyCex(g, cex) != 1 {
					t.Fatal("witness does not replay")
				}
			})
		}
	}
}
