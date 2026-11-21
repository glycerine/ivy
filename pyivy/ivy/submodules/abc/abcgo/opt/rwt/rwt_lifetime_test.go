package rwt

import (
	"runtime"
	"sync"
	"testing"
)

func TestStoppedManagerReleasesCanonicalTables(t *testing.T) {
	a, b := Rwt_ManStart(1), Rwt_ManStart(1)
	defer Rwt_ManStop(b)
	Rwt_ManGlobalStop()
	if len(a.puCanons) != 65536 || &a.puCanons[0] != &b.puCanons[0] {
		t.Fatal("live managers lost their canonical tables")
	}
	fresh := Rwt_ManStart(1)
	defer Rwt_ManStop(fresh)
	defer Rwt_ManGlobalStop()
	if &fresh.puCanons[0] == &a.puCanons[0] || fresh.puCanons[0x6996] != a.puCanons[0x6996] {
		t.Fatal("restarted frame lost the canonical function or reused the old table")
	}
	Rwt_ManStop(a)
	if a.puCanons != nil || a.pPhases != nil || a.pPerms != nil || a.pMap != nil {
		t.Error("stopped manager retains the frame's canonical tables")
	}
	if len(b.pMap) != 65536 || b.pMap[1] != 1 {
		t.Fatal("stopping one manager cleared a live manager's tables")
	}
}

func TestConcurrentFrameStopAndManagerStart(t *testing.T) {
	finish, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			select {
			case <-finish:
				return
			default:
				Rwt_ManGlobalStop()
				runtime.Gosched()
			}
		}
	}()
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 12 {
				p := Rwt_ManStart(1)
				if len(p.puCanons) != 65536 || len(p.pPhases) != 65536 || len(p.pPerms) != 65536 || len(p.pMap) != 65536 {
					t.Error("a live manager captured an empty frame")
					Rwt_ManStop(p)
					return
				}
				Rwt_ManStop(p)
			}
		})
	}
	wg.Wait()
	close(finish)
	<-stopped
	Rwt_ManGlobalStop()
}
