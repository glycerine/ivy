package abc

// Go-only compatibility tests for the duplicate bar progress implementation.

import (
	"strings"
	"testing"
)

func TestBarProgressFacade(t *testing.T) {
	if BarProgressUse != 1 {
		t.Fatal(BarProgressUse)
	}
	if frame, changed, err := UpdateBarProgress(nil, 10, "ignored"); err != nil || changed || frame != "" {
		t.Fatal(frame, changed, err)
	}
	if frame, err := StopBarProgress(nil); err != nil || frame != "" {
		t.Fatal(frame, err)
	}
	bar, initial, err := StartBarProgress(78, ExtraProgressBarOptions{Enabled: true})
	if err != nil || len(initial) != 80 {
		t.Fatal(len(initial), err)
	}
	if frame, changed, err := UpdateBarProgress(bar, 7, "early"); err != nil || changed || frame != "" {
		t.Fatal(frame, changed, err)
	}
	frame, changed, err := UpdateBarProgress(bar, 8, "step")
	if err != nil || !changed || !strings.HasPrefix(frame, "step --->") {
		t.Fatal(frame, changed, err)
	}
	frame, changed, err = UpdateBarProgress(bar, 78, "done")
	if err != nil || !changed || frame != "done "+strings.Repeat("-", 73)+">\r" {
		t.Fatal(frame, changed, err)
	}
	if frame, err = StopBarProgress(bar); err != nil || frame != strings.Repeat(" ", 79)+"\r" {
		t.Fatal(frame, err)
	}
}
