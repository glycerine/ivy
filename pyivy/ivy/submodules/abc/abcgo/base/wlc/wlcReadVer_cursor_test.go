package wlc

import "testing"

func TestWlcPrsCursorModelsCStringMovement(t *testing.T) {
	p := &Wlc_Prs_t{pBuffer: "  \\escaped name , tail\x00ignored"}
	c := wlcPrsCursorAt(p, 0)
	nameAt, name := c.findName()
	if name != `\escaped name ` {
		t.Fatalf("name = %q", name)
	}
	comma := nameAt.findSymbol(',')
	if !comma.valid() || comma.byte() != ',' {
		t.Fatalf("comma cursor = %#v", comma)
	}
	if got := comma.advance(1).skipSpaces().rest(); got != "tail" {
		t.Fatalf("rest after comma = %q", got)
	}
}

func TestWlcPrsCursorFindWordAndNull(t *testing.T) {
	p := &Wlc_Prs_t{pBuffer: "  module name"}
	c, found := wlcPrsCursorAt(p, 0).findWord("module")
	if !found || c.rest() != "module name" {
		t.Fatalf("findWord = (%q, %v)", c.rest(), found)
	}
	miss, found := c.findWord("endmodule")
	if found || !miss.valid() || miss.rest() != "" {
		t.Fatalf("missing word cursor = (%q, %v)", miss.rest(), found)
	}
}

func TestWlcPrsCursorModuleAndTableTokens(t *testing.T) {
	p := &Wlc_Prs_t{pBuffer: "module table(input x);\n  2'b01: 4'hA;\nendmodule"}
	line := wlcPrsCursorAt(p, 0)
	module, ok := line.findWord("module")
	if !ok {
		t.Fatal("module keyword was not recognized")
	}
	_, name := module.advance(len("module")).findName()
	if name != "table" {
		t.Fatalf("module name = %q", name)
	}
	tableLine := wlcPrsCursorAt(p, len("module table(input x);\n"))
	firstQuote := tableLine.findSymbol('\'')
	if !firstQuote.valid() || firstQuote.retreat(1).rest()[0] != '2' {
		t.Fatalf("first table width cursor = %q", firstQuote.retreat(1).rest())
	}
	secondQuote := firstQuote.advance(2).findSymbol('\'')
	if !secondQuote.valid() || secondQuote.retreat(1).rest()[0] != '4' {
		t.Fatalf("second table width cursor = %q", secondQuote.retreat(1).rest())
	}
	end, ok := tableLine.findSubstring("endmodule")
	if !ok || end.rest() != "endmodule" {
		t.Fatalf("endmodule cursor = %q, found=%v", end.rest(), ok)
	}
}
