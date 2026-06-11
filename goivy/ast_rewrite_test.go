package goivy

import (
	"strings"
	"testing"
)

func TestActionDefRewritePreservesThisSortNode(t *testing.T) {
	cfg := NewAstConfig()
	param := cfg.NewApp(cfg.NewSymbol("base", nil))
	param.ASort = cfg.NewThis()
	ret := cfg.NewApp(cfg.NewSymbol("result", nil))
	ret.ASort = cfg.NewThis()

	action := cfg.NewActionDef(
		cfg.NewAtom("pow"),
		cfg.NewSymbol("skip", nil),
		[]Node{param},
		[]Node{ret},
	)

	rewritten := action.Rewrite(NewAstRewriteSubstPrefix(nil, nil)).(*ActionDef)
	gotParam := rewritten.FormalParams[0].(*App)
	if _, ok := gotParam.ASort.(*This); !ok {
		t.Fatalf("rewritten formal param sort = %T, want *This", gotParam.ASort)
	}
	gotReturn := rewritten.FormalReturns[0].(*App)
	if _, ok := gotReturn.ASort.(*This); !ok {
		t.Fatalf("rewritten formal return sort = %T, want *This", gotReturn.ASort)
	}

	canon := string(rewritten.Canon())
	if !strings.Contains(canon, "aSort:(this)") {
		t.Fatalf("rewritten action canon does not preserve Python This sort spelling: %s", canon)
	}
	if strings.Contains(canon, "aSort:this)") || strings.Contains(canon, "aSort:this]") {
		t.Fatalf("rewritten action canon used bare this sort spelling: %s", canon)
	}
}

func TestActionDefRewriteSubstitutesThisSortNode(t *testing.T) {
	cfg := NewAstConfig()
	param := cfg.NewApp(cfg.NewSymbol("base", nil))
	param.ASort = cfg.NewSymbol("base_t", nil)
	ret := cfg.NewApp(cfg.NewSymbol("result", nil))
	ret.ASort = cfg.NewSymbol("base_t", nil)

	action := cfg.NewActionDef(
		cfg.NewAtom("pow"),
		cfg.NewSymbol("skip", nil),
		[]Node{param},
		[]Node{ret},
	)
	rw := &AstRewriteSubstPrefix{
		Subst:      map[string]string{"base_t": "this"},
		SubstNodes: map[string]Node{"base_t": cfg.NewThis()},
	}

	rewritten := action.Rewrite(rw).(*ActionDef)
	gotParam := rewritten.FormalParams[0].(*App)
	if _, ok := gotParam.ASort.(*This); !ok {
		t.Fatalf("rewritten formal param sort = %T, want *This", gotParam.ASort)
	}
	gotReturn := rewritten.FormalReturns[0].(*App)
	if _, ok := gotReturn.ASort.(*This); !ok {
		t.Fatalf("rewritten formal return sort = %T, want *This", gotReturn.ASort)
	}

	canon := string(rewritten.Canon())
	if !strings.Contains(canon, "aSort:(this)") {
		t.Fatalf("rewritten action canon does not preserve Python This sort spelling: %s", canon)
	}
	if strings.Contains(canon, "aSort:this)") || strings.Contains(canon, "aSort:this]") {
		t.Fatalf("rewritten action canon used bare this sort spelling: %s", canon)
	}
}

func TestSortCanonUnwrapsNullaryAtomButPreservesThis(t *testing.T) {
	cfg := NewAstConfig()

	upper := cfg.NewApp(cfg.NewSymbol("client_id.max", nil))
	upper.ASort = cfg.NewAtom("client_id")
	if got := string(upper.Canon()); !strings.Contains(got, "aSort:client_id") {
		t.Fatalf("nullary atom sort should canon as bare Python sort name: %s", got)
	}

	self := cfg.NewApp(cfg.NewSymbol("base", nil))
	self.ASort = cfg.NewThis()
	if got := string(self.Canon()); !strings.Contains(got, "aSort:(this)") {
		t.Fatalf("This sort should keep Python This object spelling: %s", got)
	}
}
