package extra

import (
	"bytes"
	"strings"
	"testing"
)

func TestProgressLifecycle(t *testing.T) {
	Extra_ProgressBarSetMode(true, false)
	defer Extra_ProgressBarSetMode(true, false)
	var out bytes.Buffer
	p := Extra_ProgressBarStart(&out, 100)
	if got := out.String(); len(got) != 80 || !strings.HasPrefix(got, "->") || !strings.HasSuffix(got, "\r") {
		t.Fatal(got)
	}
	out.Reset()
	Extra_ProgressBarUpdate(p, 0, "")
	if out.Len() != 0 {
		t.Fatal("premature update")
	}
	Extra_ProgressBarUpdate(p, 100, "done")
	if p.posCur != 78 || !strings.Contains(out.String(), ">") {
		t.Fatal("completion")
	}
	out.Reset()
	Extra_ProgressBarStop(p)
	if out.String() != strings.Repeat(" ", 79)+"\r" {
		t.Fatal("clean")
	}
	out.Reset()
	Extra_ProgressBarUpdate(p, 100, "")
	Extra_ProgressBarStop(p)
	if out.Len() != 0 {
		t.Fatal("stopped bar reused")
	}
}
func TestProgressModesAndOverflow(t *testing.T) {
	defer Extra_ProgressBarSetMode(true, false)
	var out bytes.Buffer
	Extra_ProgressBarSetMode(false, false)
	if Extra_ProgressBarStart(&out, 1) != nil {
		t.Fatal("disabled")
	}
	Extra_ProgressBarSetMode(true, true)
	p := Extra_ProgressBarStart(&out, 1<<31-1)
	Extra_ProgressBarUpdate(p, p.nItemsNext, "")
	Extra_ProgressBarStop(p)
	if out.Len() != 0 || p.nItemsNext < 0 {
		t.Fatal("batch/threshold")
	}
	Extra_ProgressBarUpdate(nil, 0, "")
	Extra_ProgressBarStop(nil)
}
