// Package bar provides ABC's throttled text progress indicator.
// Ported from src/misc/bar/bar.c (Alan Mishchenko).
package bar

import (
	"io"
	"strings"
	"sync"
	"sync/atomic"
)

// Frame-independent controls let the command layer supply its progress and
// batch-mode settings without introducing a dependency cycle through base/main.
var progressDisabled atomic.Bool
var batchMode atomic.Bool

func Bar_SetEnabled(enabled bool) { progressDisabled.Store(!enabled) }
func Bar_SetBatchMode(batch bool) { batchMode.Store(batch) }

type Bar_Progress_t struct {
	mu          sync.Mutex
	nItemsNext  int32
	nItemsTotal int32
	posTotal    int32
	posCur      int32
	pFile       io.Writer
	stopped     bool
	err         error
}
type Bar_Progress_t_ = Bar_Progress_t

// Bar_ProgressStart draws the initial indicator. Nil output or disabled progress
// returns nil, which all update/stop functions accept.
func Bar_ProgressStart(output io.Writer, total int32) *Bar_Progress_t {
	if output == nil || progressDisabled.Load() {
		return nil
	}
	p := &Bar_Progress_t{pFile: output, nItemsTotal: total, posTotal: 78, posCur: 1}
	p.nItemsNext = int32(int64(7+p.posCur) * int64(total) / int64(p.posTotal))
	Bar_ProgressShow(p, "")
	return p
}
func Bar_ProgressUpdate_int(p *Bar_Progress_t, current int32, label string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped || current < p.nItemsNext {
		return
	}
	if current >= p.nItemsTotal {
		p.posCur = 78
		p.nItemsNext = 1<<31 - 1
	} else {
		p.posCur += 7
		p.nItemsNext = int32(int64(7+p.posCur) * int64(p.nItemsTotal) / int64(p.posTotal))
	}
	p.show(label)
}
func Bar_ProgressStop(p *Bar_Progress_t) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	p.clean()
	p.stopped = true
}
func Bar_ProgressShow(p *Bar_Progress_t, label string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.stopped {
		p.show(label)
	}
}
func Bar_ProgressClean(p *Bar_Progress_t) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.stopped {
		p.clean()
	}
}
func (p *Bar_Progress_t) write(text string) {
	if p.err != nil {
		return
	}
	_, p.err = io.WriteString(p.pFile, text)
	if f, ok := p.pFile.(interface{ Flush() error }); ok && p.err == nil {
		p.err = f.Flush()
	}
}
func (p *Bar_Progress_t) show(label string) {
	if batchMode.Load() {
		return
	}
	var b strings.Builder
	i := int32(0)
	if label != "" {
		b.WriteString(label)
		b.WriteByte(' ')
		i = int32(len(label) + 1)
	}
	for ; i < p.posCur; i++ {
		b.WriteByte('-')
	}
	if i == p.posCur {
		b.WriteByte('>')
	}
	for i++; i <= p.posTotal; i++ {
		b.WriteByte(' ')
	}
	b.WriteByte('\r')
	p.write(b.String())
}
func (p *Bar_Progress_t) clean() {
	if !batchMode.Load() {
		p.write(strings.Repeat(" ", int(p.posTotal)+1) + "\r")
	}
}

// Err reports a write/flush error encountered while rendering progress.
func (p *Bar_Progress_t) Err() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}
