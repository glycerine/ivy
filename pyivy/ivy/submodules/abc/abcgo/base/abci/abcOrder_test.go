package abci

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSourceCiOrderAndDuplicateNames(t *testing.T) {
	n := abc.NewNetwork("order")
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")
	support := vec.Vec_PtrAllocArrayCopy([]any{c, a, b}, 3)
	Abc_NtkChangeCiOrder(n, support, 1)
	if got, want := n.Inputs(), []*abc.Object{b, a, c}; !reflect.DeepEqual(got, want) {
		t.Fatalf("reverse order %v, want %v", got, want)
	}
	name := filepath.Join(t.TempDir(), "order")
	if err := os.WriteFile(name, []byte("c a a\n"), 0600); err != nil {
		t.Fatal(err)
	}
	Abc_NtkImplementCiOrder(n, name, 0, 0)
	if got, want := n.Inputs(), []*abc.Object{c, a, a}; !reflect.DeepEqual(got, want) {
		t.Fatalf("duplicate source order %v, want %v", got, want)
	}
	if got, want := n.ObjectsOfKind(abc.ObjectPI), []*abc.Object{c, a, a}; !reflect.DeepEqual(got, want) {
		t.Fatalf("PI order %v, want %v", got, want)
	}
}

func TestFindCiOrderUsesNetworkSupportAndReverse(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, unused, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	root := abc.Abc_AigAnd(n.FunctionManager.(*abc.Abc_Aig_t), c.Signal(), a.Signal())
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, root)
	support := abc.Abc_NtkSupport(n)
	Abc_NtkFindCiOrder(n, 0, 0)
	if got := n.Inputs(); !reflect.DeepEqual(got, support) {
		t.Fatalf("support order %v, want %v", got, support)
	}
	reversed := append([]*abc.Object(nil), support...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	Abc_NtkFindCiOrder(n, 1, 0)
	if got := n.Inputs(); !reflect.DeepEqual(got, reversed) {
		t.Fatalf("reverse support order %v, want %v", got, reversed)
	}
	foundUnused := false
	for _, ci := range n.Inputs() {
		foundUnused = foundUnused || ci == unused
	}
	if len(n.Inputs()) != 3 || !foundUnused {
		t.Fatal("unused CI was not retained")
	}
}

func TestCiOrderTokenUsesCWhitespace(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\talpha\u00a0beta\nsecond\r\n"))
	first, err := abcReadCiOrderToken(r)
	if err != nil || first != "alpha\u00a0beta" {
		t.Fatalf("first token %q, err %v", first, err)
	}
	second, err := abcReadCiOrderToken(r)
	if err != nil || second != "second" {
		t.Fatalf("second token %q, err %v", second, err)
	}
}

func TestImplementCiOrderReversesFileOrder(t *testing.T) {
	n := abc.NewNetwork("order")
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")
	name := filepath.Join(t.TempDir(), "order")
	if err := os.WriteFile(name, []byte("c a b\n"), 0600); err != nil {
		t.Fatal(err)
	}
	Abc_NtkImplementCiOrder(n, name, 1, 0)
	if got, want := n.Inputs(), []*abc.Object{b, a, c}; !reflect.DeepEqual(got, want) {
		t.Fatalf("reversed file order %v, want %v", got, want)
	}
}

func TestImplementCiOrderPrintsCNameMessage(t *testing.T) {
	n := abc.NewNetwork("order")
	a := abc.Abc_NtkCreatePi(n)
	a.SetName("a")
	name := filepath.Join(t.TempDir(), "order")
	if err := os.WriteFile(name, []byte("bad\"name\n"), 0600); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = write
	Abc_NtkImplementCiOrder(n, name, 0, 0)
	_ = write.Close()
	os.Stdout = old
	got, err := io.ReadAll(read)
	_ = read.Close()
	if err != nil {
		t.Fatal(err)
	}
	if want := "Name \"bad\"name\" is not a PI name. Cannot use this order.\n"; string(got) != want {
		t.Fatalf("message %q, want %q", got, want)
	}
}
func TestSourceCiOrderInvalidNameLeavesOrder(t *testing.T) {
	n := abc.NewNetwork("order")
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	a.SetName("a")
	b.SetName("b")
	name := filepath.Join(t.TempDir(), "order")
	if err := os.WriteFile(name, []byte("a missing\n"), 0600); err != nil {
		t.Fatal(err)
	}
	Abc_NtkImplementCiOrder(n, name, 0, 0)
	if got, want := n.Inputs(), []*abc.Object{a, b}; !reflect.DeepEqual(got, want) {
		t.Fatalf("invalid order changed input list: %v", got)
	}
}
