package io

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestRegisterCexExtensionPreservesMinimizer(t *testing.T) {
	cexServices.RLock()
	original := cexServices.value
	cexServices.RUnlock()
	defer RegisterCexServices(original)
	called := false
	RegisterCexServices(CexServices{Minimize: func(*aig.Graph, *util.Abc_Cex_t, bool, bool, bool, bool, bool) *util.Abc_Cex_t {
		called = true
		return nil
	}})
	RegisterCexExtension(func(*aig.Graph, *util.Abc_Cex_t) *util.Abc_Cex_t { return nil })
	cexServices.RLock()
	services := cexServices.value
	cexServices.RUnlock()
	if services.Extend == nil || services.Minimize == nil {
		t.Fatal("registration lost a service")
	}
	services.Minimize(nil, nil, false, false, false, false, false)
	if !called {
		t.Fatal("minimizer was replaced")
	}
}
