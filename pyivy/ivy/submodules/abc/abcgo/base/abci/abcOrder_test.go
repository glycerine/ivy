package abci

import (
	"os"
	"path/filepath"
	"reflect"
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
