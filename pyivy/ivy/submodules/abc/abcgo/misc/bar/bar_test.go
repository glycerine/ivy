package bar

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestProgressThrottleCompleteAndStop(t *testing.T) {
	Bar_SetEnabled(true)
	Bar_SetBatchMode(false)
	var b bytes.Buffer
	p := Bar_ProgressStart(&b, 780)
	first := b.String()
	if len(first) != 80 || !strings.HasPrefix(first, "->") {
		t.Fatal(first)
	}
	Bar_ProgressUpdate(p, 79, "")
	if b.String() != first {
		t.Fatal("unthrottled")
	}
	Bar_ProgressUpdate(p, 80, "work")
	if !strings.Contains(b.String(), "work --->") {
		t.Fatal(b.String())
	}
	Bar_ProgressUpdate(p, 780, "")
	if p.posCur != 78 || p.nItemsNext != 1<<31-1 {
		t.Fatal("not complete")
	}
	Bar_ProgressStop(p)
	if !strings.HasSuffix(b.String(), strings.Repeat(" ", 79)+"\r") {
		t.Fatal("not cleared")
	}
	n := b.Len()
	Bar_ProgressStop(p)
	Bar_ProgressUpdate(p, 1000, "")
	if b.Len() != n {
		t.Fatal("stopped progress wrote")
	}
}
func TestDisabledAndBatch(t *testing.T) {
	Bar_SetEnabled(false)
	var b bytes.Buffer
	if Bar_ProgressStart(&b, 10) != nil {
		t.Fatal("enabled")
	}
	Bar_ProgressUpdate(nil, 1, "")
	Bar_ProgressStop(nil)
	Bar_SetEnabled(true)
	Bar_SetBatchMode(true)
	defer Bar_SetBatchMode(false)
	p := Bar_ProgressStart(&b, 1)
	Bar_ProgressUpdate(p, 1, "")
	Bar_ProgressStop(p)
	if b.Len() != 0 {
		t.Fatal(b.String())
	}
}

type badWriter struct{}

func (badWriter) Write([]byte) (int, error) { return 0, errors.New("broken") }
func TestProgressErrorsAndLongLabel(t *testing.T) {
	p := Bar_ProgressStart(badWriter{}, 0)
	if p.Err() == nil {
		t.Fatal("lost error")
	}
	var b bytes.Buffer
	p = Bar_ProgressStart(&b, 10)
	b.Reset()
	label := strings.Repeat("x", 90)
	Bar_ProgressShow(p, label)
	if b.String() != label+" \r" {
		t.Fatal(b.String())
	}
}
