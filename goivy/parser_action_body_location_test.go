package goivy

import (
	"os"
	"path/filepath"
	"testing"
)

// Python p_top_optimpex_action_symbol_optargs_optreturns_eq_action gives
// a body without a lineno (an empty optactiondef) get_lineno(p,4), the
// action name's line (test/test_array.ivy prints "...in action arr.set"
// with its instantiation reference).
func TestBodilessActionGetsNameLocation(t *testing.T) {
	for _, lang := range []string{"ivy1.6", "ivy1.7"} {
		src := "#lang " + lang + "\n\ntype t\n\naction\n  foo(x:t)\n"
		fn := filepath.Join(t.TempDir(), "a.ivy")
		if err := os.WriteFile(fn, []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
		res, err := ReadModule(fn, false, New().Cfg)
		if err != nil {
			t.Fatal(lang, err)
		}
		found := false
		for _, d := range res.Decls {
			ad, ok := d.(*ActionDecl)
			if !ok {
				continue
			}
			def := ad.Args()[0].(*ActionDef)
			body := def.Args()[1]
			if got := body.GetLineno().Line; got != 6 {
				t.Errorf("%s: body line %d, want 6", lang, got)
			}
			found = true
		}
		if !found {
			t.Fatal(lang, "no action")
		}
	}
}
