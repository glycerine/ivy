package abcmain

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/misc/extra"
)

// Extra_ProgressBarStart returns NULL unless the global frame's
// "progressbar" flag is enabled (Abc_FrameShowProgress), so batch runs
// such as "abc -c ..." print no progress bars by default.
func TestProgressBarFollowsFrameFlag(t *testing.T) {
	f := Abc_FrameGetGlobalFrame()
	f.DeleteFlag("progressbar")
	var out bytes.Buffer
	if p := extra.Extra_ProgressBarStart(&out, 10); p != nil || out.Len() != 0 {
		t.Fatalf("progress bar shown without the progressbar flag: %q", out.String())
	}
	f.SetFlag("progressbar", "")
	defer f.DeleteFlag("progressbar")
	p := extra.Extra_ProgressBarStart(&out, 10)
	if p == nil {
		t.Fatal("progress bar hidden with the progressbar flag set")
	}
	extra.Extra_ProgressBarStop(p)
}
