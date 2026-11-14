package bac

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestBlifAndVerilogNameReadersKeepTheirOwnTokenRules(t *testing.T) {
	names := util.Abc_NamStart(16, 16)
	verilog := &Psr_Man_t{pStrs: names, pCur: "alpha.beta \x00"}
	blif := &Psr_Man_t{pStrs: names, pCur: "alpha.beta \x00"}

	if got := Psr_ManReadName(verilog); got == 0 || verilog.pCur != ".beta \x00" {
		t.Fatalf("Verilog name parse got id %d and cursor %q", got, verilog.pCur)
	}
	if got := Psr_BlifReadName(blif); got == 0 || blif.pCur != " \x00" {
		t.Fatalf("BLIF name parse got id %d and cursor %q", got, blif.pCur)
	}
}

func TestParserWhitespaceRulesRemainFormatSpecific(t *testing.T) {
	if Psr_CharIsSpace('\n') == 0 {
		t.Fatal("Verilog whitespace must include newline")
	}
	if Psr_BlifCharIsSpace('\n') != 0 {
		t.Fatal("BLIF token spacing must stop at newline")
	}
}
