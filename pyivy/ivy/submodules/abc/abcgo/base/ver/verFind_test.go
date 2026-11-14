package ver

import (
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func TestFindNetConstantAliases(t *testing.T) {
	n := abc.NewNetwork("module")
	if Ver_ParseFindNet(n, "missing") != nil {
		t.Fatal("unknown net")
	}
	zero := Ver_ParseFindNet(n, "1'b0")
	if zero == nil || Ver_ParseFindNet(n, "1'bx") != zero {
		t.Fatal("x should share zero net")
	}
	one := Ver_ParseFindNet(n, "1'b1")
	if one == nil || one == zero {
		t.Fatal("constant one net")
	}
	named := abc.Abc_NtkFindOrCreateNet(n, "wire")
	if Ver_ParseFindNet(n, "wire") != named {
		t.Fatal("existing net")
	}
}

func TestFindOrCreateNetworkKeepsModelIdentity(t *testing.T) {
	p := &Ver_Man_t{pDesign: abc.Abc_DesCreate("design")}
	defer Ver_ParseFreeData(p)
	first := Ver_ParseFindOrCreateNetwork(p, "module")
	if first == nil || first.Name != "module" || first.Type != uint8(abc.ABC_NTK_NETLIST) || first.Function != uint8(abc.ABC_FUNC_BLACKBOX) {
		t.Fatal("new blackbox mismatch")
	}
	if Ver_ParseFindOrCreateNetwork(p, "module") != first || len(p.pDesign.Modules) != 1 {
		t.Fatal("existing module not reused")
	}
}
