package bac

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAbcNamStrFindOrAddLimUsesParserSpan(t *testing.T) {
	names := util.Abc_NamStart(8, 8)
	defer util.Abc_NamDeref(names)

	start := "signal[7:0]tail"
	if got := Abc_NamStrFindOrAddLim(names, start, len(start)-len("tail"), nil); got == 0 {
		t.Fatal("name insertion returned reserved ID zero")
	} else if name := util.Abc_NamStr(names, got); name != "signal[7:0]" {
		t.Fatalf("interned name = %q, want %q", name, "signal[7:0]")
	}
}

func TestVecStrCompatibilityHelpers(t *testing.T) {
	var value vec.Vec_Str_t
	Vec_StrPush(&value, 'x')
	Vec_StrPush(&value, 'y')
	if got := Vec_StrEntryLast(&value); got != 'y' {
		t.Fatalf("last entry = %q, want y", got)
	}
	Vec_StrClear(&value)
	if got := Vec_StrSize(&value); got != 0 {
		t.Fatalf("size after clear = %d, want 0", got)
	}
}
