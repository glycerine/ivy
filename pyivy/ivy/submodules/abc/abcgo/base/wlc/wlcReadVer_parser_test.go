package wlc

import (
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestWlcPrsFindRangeAndEscapedName(t *testing.T) {
	var end, beg int32
	if rest := Wlc_PrsFindRange(" [ 7 : 0 ] tail", &end, &beg); rest != " tail" || end != 7 || beg != 0 {
		t.Fatalf("range parse = (%q, %d:%d)", rest, end, beg)
	}
	if rest := Wlc_PrsFindRange("[3]x", &end, &beg); rest != "x" || end != 3 || beg != 3 {
		t.Fatalf("single range parse = (%q, %d:%d)", rest, end, beg)
	}
	if rest := Wlc_PrsFindRange("[3]", &end, &beg); rest != "\x00" || end != 3 || beg != 3 {
		t.Fatalf("terminal single range parse = (%q, %d:%d)", rest, end, beg)
	}
	if rest := Wlc_PrsFindRange("[x:0]", &end, &beg); rest != "" {
		t.Fatalf("malformed range should fail, got %q", rest)
	}
	var name string
	if rest := Wlc_PrsFindName(`  \signal name  , tail`, &name); name != `\signal name` || rest != "  , tail" {
		t.Fatalf("escaped name parse = (%q, %q)", name, rest)
	}
	if rest := Wlc_PrsFindName("tail", &name); name != "tail" || rest != "\x00" {
		t.Fatalf("terminal name cursor = (%q, %q)", name, rest)
	}
	if got := Wlc_PrsSkipSpaces("   "); got != "\x00" {
		t.Fatalf("space skip at terminator = %q", got)
	}
}

func TestWlcPrsReadConstantBinaryAndHex(t *testing.T) {
	p := &Wlc_Prs_t{pFileName: "test.v"}
	v := vec.Vec_IntAlloc(0)
	defer vec.Vec_IntFree(v)
	var width, signed, x int32
	if rest := Wlc_PrsReadConstant(p, "4'sb1010 tail", v, &width, &signed, &x); rest != " tail" {
		t.Fatalf("binary constant remainder = %q", rest)
	}
	if width != 4 || signed != 1 || Vec_IntSize(v) != 1 || uint32(Vec_IntEntry(v, 0))&0xf != 0xa {
		t.Fatalf("binary parse width/signed/words = %d/%d/%v", width, signed, Vec_IntArray(v))
	}
	if rest := Wlc_PrsReadConstant(p, "8'h3 tail", v, &width, &signed, &x); rest != " tail" {
		t.Fatalf("hex constant remainder = %q", rest)
	}
	if width != 8 || x != 0 || uint32(Vec_IntEntry(v, 0))&0xff != 0x03 {
		t.Fatalf("hex parse width/x/value = %d/%d/0x%x", width, x, uint32(Vec_IntEntry(v, 0)))
	}
	if rest := Wlc_PrsReadConstant(p, "8'hx tail", v, &width, &signed, &x); rest != " tail" || x != 1 || Vec_IntEntry(v, 0) != 0 {
		t.Fatalf("x-valued hex parse = (remainder %q, x=%d, value=%d)", rest, x, Vec_IntEntry(v, 0))
	}
}

func TestWlcPrsRemoveCommentsPreservesPairAndEndmodule(t *testing.T) {
	source := "// abc2 metadata\n// Pair: in out\n`timescale 1ns\nmodule top; /* body */ endmodule\n"
	p := &Wlc_Prs_t{pBuffer: source, nFileSize: int32(len(source))}
	if got := Wlc_PrsRemoveComments(p); got != 1 {
		t.Fatalf("remove comments returned %d", got)
	}
	if p.vPoPairs == nil || string(Vec_StrArray(p.vPoPairs)[:Vec_StrSize(p.vPoPairs)]) != "in\x00out\x00" {
		t.Fatalf("PO pairs = %v", Vec_StrArray(p.vPoPairs))
	}
	if got := p.pBuffer; got[len(got)-1:] != ";" {
		t.Fatalf("endmodule terminator = %q", got[len(got)-1:])
	}
}

func TestWlcPrsRemoveCommentsReportsUnterminatedLineComment(t *testing.T) {
	source := "// missing newline"
	p := &Wlc_Prs_t{pBuffer: source, nFileSize: int32(len(source)), vLines: Vec_IntAlloc(1)}
	defer Vec_IntFree(p.vLines)
	if got := Wlc_PrsRemoveComments(p); got != 0 {
		t.Fatalf("unterminated comment returned %d, want 0", got)
	}
	if !strings.Contains(string(p.sError[:]), "Cannot find end-of-line") {
		t.Fatalf("error = %q", string(p.sError[:]))
	}
}

func TestWlcPrsPrepareKeepsSemicolonsInsideEscapedNames(t *testing.T) {
	source := "module top;\nwire \\escaped;name ;\nendmodule\n"
	p := &Wlc_Prs_t{
		pBuffer: source,
		vLines:  Vec_IntAlloc(4),
		vStarts: Vec_IntAlloc(4),
	}
	defer Vec_IntFree(p.vLines)
	defer Vec_IntFree(p.vStarts)
	if got := Wlc_PrsPrepare(p); got != 1 {
		t.Fatalf("prepare returned %d", got)
	}
	if got := Vec_IntSize(p.vStarts); got != 3 {
		t.Fatalf("statement count = %d, want 3", got)
	}
	if got := Wlc_PrsStr(p, Vec_IntEntry(p.vStarts, 1)); !strings.HasPrefix(got, "wire \\escaped;name ") {
		t.Fatalf("escaped declaration line = %q", got)
	}
}

func TestWlcPrsStrtokSourceOrderingAndEscapedName(t *testing.T) {
	spStatic_Wlc_PrsStrtok_last = ""
	defer func() { spStatic_Wlc_PrsStrtok_last = "" }()
	if got := Wlc_PrsStrtok(" , first,,second", ", "); got != "first\x00" {
		t.Fatalf("first token = %q", got)
	}
	if got := Wlc_PrsStrtok("", ", "); got != "second\x00" {
		t.Fatalf("second token = %q", got)
	}
	if got := Wlc_PrsStrtok("", ", "); got != "" {
		t.Fatalf("exhausted token = %q", got)
	}
	if got := Wlc_PrsStrtok(`\escaped name ,tail`, ", "); got != "\\escaped name\x00" {
		t.Fatalf("escaped token = %q", got)
	}
	if got := Wlc_PrsStrtok("", ", "); got != "tail\x00" {
		t.Fatalf("token after escaped name = %q", got)
	}
}
