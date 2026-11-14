package ivy

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestIvyManagerCopyAccessors(t *testing.T) {
	manager := Ivy_ManStart()
	defer Ivy_ManStop(manager)
	copy := vec.Vec_PtrAlloc(1)
	Ivy_ManSetCopy(manager, copy)
	if got := Ivy_ManReadCopy(manager); got != copy {
		t.Fatal("copy accessor changed the temporary payload")
	}
	Ivy_ManSetCopy(manager, nil)
	if Ivy_ManReadCopy(manager) != nil {
		t.Fatal("copy accessor did not clear the temporary payload")
	}
	vec.Vec_PtrFree(copy)
}
