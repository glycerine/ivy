package hop

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSharedPointerVectorInteroperability(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	a, b := Hop_IthVar(p, 0), Hop_IthVar(p, 1)
	root := Hop_And(p, a, b)
	var shared *vec.Vec_Ptr_t = Hop_ManDfs(p)
	if vec.Vec_PtrSize(shared) != 1 || vec.Vec_PtrEntry(shared, 0) != root {
		t.Fatal("DFS vector cannot be consumed through shared API")
	}
	store := Vec_VecAlloc(1)
	defer Vec_VecFree(store)
	Vec_VecPush(store, 2, root)
	var sharedStore *vec.Vec_Vec_t = store
	if vec.Vec_VecEntryEntry(sharedStore, 2, 0) != root {
		t.Fatal("nested vector bridge")
	}
}
func TestCleanupRejectsLevelCounters(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	p.fRefCount = 0
	defer func() {
		if recover() == nil {
			t.Fatal("cleanup accepted level counters as reference counts")
		}
	}()
	Hop_ManCleanup(p)
}
