package amap

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLibertyQuotedFormulaAndRoundTrip(t *testing.T) {
	source, err := os.ReadFile("test_vectors/library.lib")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := ParseAMAPLibertyTree("library.lib", string(source))
	if err != nil {
		t.Fatal(err)
	}
	genlib, err := LibertyGenlib(tree)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(genlib, "GATE NAND2 1.25 Z=!(A*B);") || strings.Contains(genlib, "GATE SKIP") || strings.Contains(genlib, "GATE DFF") {
		t.Fatal(genlib)
	}
	var output bytes.Buffer
	if err = Amap_LibertyPrintLibertyItem(&output, tree, Amap_LibertyRoot(tree), 0); err != nil {
		t.Fatal(err)
	}
	copy, err := ParseAMAPLibertyTree("copy.lib", output.String())
	if err != nil {
		t.Fatal(err)
	}
	again, err := LibertyGenlib(copy)
	if err != nil || again != genlib {
		t.Fatalf("round trip: %v\n%s", err, again)
	}
}

func TestLibertyMalformed(t *testing.T) {
	for _, source := range []string{`library(x){`, `library(x){ area:; }`, `library(x){ function:"oops; }`, `library(x){ /*oops`, `cell(x){}`, `library(x){} library(y){}`} {
		if _, err := ParseAMAPLibertyTree("bad.lib", source); err == nil {
			t.Errorf("accepted %q", source)
		} else if !strings.Contains(err.Error(), "bad.lib") {
			t.Errorf("missing location: %v", err)
		}
	}
}

func TestLibertyQuotedCommentPunctuation(t *testing.T) {
	tree, err := ParseAMAPLibertyTree("quoted.lib", `library(x){ description:"a /* literal */ // marker; (x)"; }`)
	if err != nil {
		t.Fatal(err)
	}
	item := amapLibertyChild(tree, Amap_LibertyRoot(tree), "description")
	if got := Amap_LibertyGetStringFormula(tree, item.Head); got != "a /* literal */ // marker; (x)" {
		t.Fatal(got)
	}
}

func TestLibertyBuildSubrangeKeepsSourceSpans(t *testing.T) {
	const source = "ignored prefix\nlibrary(x){ area:1; }suffix"
	tree := &Amap_Tree_t{pFileName: "sub.lib", pContents: source}
	position := int32(strings.Index(source, "library"))
	end := int32(strings.Index(source, "suffix"))
	if Amap_LibertyBuildItem(tree, &position, end) != 0 {
		t.Fatal(tree.pError)
	}
	if position != end || tree.pContents != source || Amap_LibertyRoot(tree).iLine != 2 {
		t.Fatal("source coordinates lost")
	}
	if got := Amap_LibertyGetString(tree, Amap_LibertyRoot(tree).Key); got != "library" {
		t.Fatal(got)
	}
}
