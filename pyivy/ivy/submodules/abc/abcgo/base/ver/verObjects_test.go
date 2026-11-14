package ver

import (
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func TestNetworkAndBoxPredicates(t *testing.T) {
	n := abc.NewNetwork("module")
	if Ver_NtkIsDefined(n) != 0 {
		t.Fatal("empty network defined")
	}
	pi := abc.Abc_NtkCreatePi(n)
	if Ver_NtkIsDefined(n) != 1 {
		t.Fatal("network with PI undefined")
	}
	box := abc.Abc_NtkCreateWhitebox(n)
	if Ver_ObjIsConnected(box) != 0 {
		t.Fatal("new box connected")
	}
	abc.Abc_ObjAddFanin(box, pi)
	if Ver_ObjIsConnected(box) != 1 {
		t.Fatal("box fanin not detected")
	}
}
