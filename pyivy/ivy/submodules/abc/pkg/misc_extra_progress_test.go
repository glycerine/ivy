package abc

// Stand-alone recurrence and lifecycle tests for extraUtilProgress.c.

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func extraProgressReference(position, next, total, current int) (int, int, bool) {
	if current < next {
		return position, next, false
	}
	if current >= total {
		return 78, math.MaxInt32, true
	}
	position += 7
	return position, int((int64(7+position) * int64(total)) / 78), true
}

func TestExtraProgressBarFramesAndSchedule(t *testing.T) {
	bar, initial, err := StartExtraProgressBar(100, ExtraProgressBarOptions{Enabled: true})
	if err != nil || initial != "->"+strings.Repeat(" ", 77)+"\r" {
		t.Fatal(len(initial), initial, err)
	}
	state, _ := bar.State()
	if state != (ExtraProgressBarState{ItemsNext: 10, ItemsTotal: 100, Position: 1}) {
		t.Fatal(state)
	}
	if frame, changed, err := bar.Update(9, "waiting"); err != nil || changed || frame != "" {
		t.Fatal(frame, changed, err)
	}
	frame, changed, err := bar.Update(10, "read")
	if err != nil || !changed || !strings.HasPrefix(frame, "read --->") || len(frame) != 80 {
		t.Fatal(len(frame), frame, changed, err)
	}
	state, _ = bar.State()
	if state.Position != 8 || state.ItemsNext != 19 {
		t.Fatal(state)
	}
	frame, changed, err = bar.Update(100, "done")
	if err != nil || !changed || frame != "done "+strings.Repeat("-", 73)+">\r" {
		t.Fatal(len(frame), frame, changed, err)
	}
	clean, err := bar.Stop()
	if err != nil || clean != strings.Repeat(" ", 79)+"\r" {
		t.Fatal(len(clean), clean, err)
	}
	if _, _, err = bar.Update(100, ""); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = bar.Stop(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestExtraProgressBarDisabledBatchLimitsAndAtomicity(t *testing.T) {
	if bar, frame, err := StartExtraProgressBar(10, ExtraProgressBarOptions{}); err != nil || bar != nil || frame != "" {
		t.Fatal(bar, frame, err)
	}
	batch, frame, err := StartExtraProgressBar(10, ExtraProgressBarOptions{Enabled: true, Batch: true, MaxBytes: 1})
	if err != nil || frame != "" {
		t.Fatal(frame, err)
	}
	if frame, changed, err := batch.Update(10, strings.Repeat("x", 1000)); err != nil || !changed || frame != "" {
		t.Fatal(frame, changed, err)
	}
	if frame, err := batch.Stop(); err != nil || frame != "" {
		t.Fatal(frame, err)
	}
	for _, total := range []int{-1, int(int64(math.MaxInt32) + 1)} {
		if _, _, err := StartExtraProgressBar(total, ExtraProgressBarOptions{Enabled: true}); !errors.Is(err, ErrDimension) {
			t.Fatal(total, err)
		}
	}
	if _, _, err := StartExtraProgressBar(10, ExtraProgressBarOptions{Enabled: true, MaxBytes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err := StartExtraProgressBar(10, ExtraProgressBarOptions{Enabled: true, MaxBytes: 79}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	bar, _, _ := StartExtraProgressBar(100, ExtraProgressBarOptions{Enabled: true, MaxBytes: 80})
	before, _ := bar.State()
	if _, _, err := bar.Update(10, strings.Repeat("x", 80)); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	after, _ := bar.State()
	if before != after {
		t.Fatal(before, after)
	}
	if _, _, err := (*ExtraProgressBar)(nil).Update(0, ""); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzExtraProgressBar(f *testing.F) {
	f.Add(uint16(100), []byte{0, 9, 10, 18, 19, 99, 100}, byte(3))
	f.Add(uint16(0), []byte{0}, byte(0))
	f.Fuzz(func(t *testing.T, rawTotal uint16, currents []byte, labelSize byte) {
		total := int(rawTotal)
		bar, initial, err := StartExtraProgressBar(total, ExtraProgressBarOptions{Enabled: true, MaxBytes: 1 << 12})
		if err != nil || len(initial) != 80 {
			t.Fatal(total, len(initial), err)
		}
		position := 1
		next := int((int64(7+position) * int64(total)) / 78)
		label := strings.Repeat("l", int(labelSize%64))
		for _, raw := range currents {
			current := int(raw)
			if total > 0 {
				current = current * (total + 1) / 255
			}
			wantPosition, wantNext, wantChanged := extraProgressReference(position, next, total, current)
			frame, changed, err := bar.Update(current, label)
			if err != nil || changed != wantChanged || (changed && (len(frame) == 0 || frame[len(frame)-1] != '\r')) || (!changed && frame != "") {
				t.Fatal(total, current, position, wantPosition, changed, wantChanged, len(frame), err)
			}
			if wantChanged {
				position, next = wantPosition, wantNext
			}
			state, _ := bar.State()
			if state.Position != position || state.ItemsNext != wantNext {
				t.Fatal(total, current, state, position, wantNext)
			}
		}
	})
}
