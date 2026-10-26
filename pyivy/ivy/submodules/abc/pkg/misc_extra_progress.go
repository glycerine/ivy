package abc

// Owned progress-bar state from src/misc/extra/extraUtilProgress.c.

import (
	"math"
	"strings"
)

const ExtraProgressBarWidth = 78

type ExtraProgressBarOptions struct {
	Enabled  bool
	Batch    bool
	MaxBytes int
}

type ExtraProgressBar struct {
	itemsNext  int
	itemsTotal int
	position   int
	batch      bool
	maxBytes   int
	stopped    bool
}

type ExtraProgressBarState struct {
	ItemsNext, ItemsTotal int
	Position              int
	Stopped               bool
}

func normalizeExtraProgressBarOptions(options ExtraProgressBarOptions) (ExtraProgressBarOptions, error) {
	if options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 20
	}
	return options, nil
}

func extraProgressThreshold(position, total int) int {
	return int((int64(7+position) * int64(total)) / ExtraProgressBarWidth)
}

func renderExtraProgress(position int, label string, batch bool, maxBytes int) (string, error) {
	if batch {
		return "", nil
	}
	index, size := 0, 1
	if label != "" {
		index = len(label) + 1
		size += index
	}
	if index < position {
		size += position - index
		index = position
	}
	if index == position {
		size++
	}
	index++
	if index <= ExtraProgressBarWidth {
		size += ExtraProgressBarWidth - index + 1
	}
	if size > maxBytes {
		return "", ErrLimit
	}
	var builder strings.Builder
	builder.Grow(size)
	index = 0
	if label != "" {
		builder.WriteString(label)
		builder.WriteByte(' ')
		index = len(label) + 1
	}
	for index < position {
		builder.WriteByte('-')
		index++
	}
	if index == position {
		builder.WriteByte('>')
	}
	index++
	for index <= ExtraProgressBarWidth {
		builder.WriteByte(' ')
		index++
	}
	builder.WriteByte('\r')
	return builder.String(), nil
}

func cleanExtraProgress(batch bool, maxBytes int) (string, error) {
	if batch {
		return "", nil
	}
	if ExtraProgressBarWidth+2 > maxBytes {
		return "", ErrLimit
	}
	return strings.Repeat(" ", ExtraProgressBarWidth+1) + "\r", nil
}

// StartExtraProgressBar creates independent progress state and returns the
// initial frame. Disabled progress follows the source's null-handle behavior.
func StartExtraProgressBar(itemsTotal int, options ExtraProgressBarOptions) (*ExtraProgressBar, string, error) {
	options, err := normalizeExtraProgressBarOptions(options)
	if err != nil {
		return nil, "", err
	}
	if !options.Enabled {
		return nil, "", nil
	}
	if itemsTotal < 0 || int64(itemsTotal) > int64(math.MaxInt32) {
		return nil, "", ErrDimension
	}
	bar := &ExtraProgressBar{
		itemsTotal: itemsTotal,
		position:   1,
		batch:      options.Batch,
		maxBytes:   options.MaxBytes,
	}
	bar.itemsNext = extraProgressThreshold(bar.position, itemsTotal)
	frame, err := renderExtraProgress(bar.position, "", bar.batch, bar.maxBytes)
	if err != nil {
		return nil, "", err
	}
	return bar, frame, nil
}

func (bar *ExtraProgressBar) State() (ExtraProgressBarState, error) {
	if bar == nil {
		return ExtraProgressBarState{}, ErrDimension
	}
	return ExtraProgressBarState{ItemsNext: bar.itemsNext, ItemsTotal: bar.itemsTotal, Position: bar.position, Stopped: bar.stopped}, nil
}

// Update advances at most one seven-column source step and returns a frame only
// when the source routine would print one.
func (bar *ExtraProgressBar) Update(itemsCurrent int, label string) (string, bool, error) {
	if bar == nil || itemsCurrent < 0 {
		return "", false, ErrDimension
	}
	if bar.stopped {
		return "", false, ErrInvariant
	}
	if itemsCurrent < bar.itemsNext {
		return "", false, nil
	}
	position, next := bar.position, bar.itemsNext
	if itemsCurrent >= bar.itemsTotal {
		position, next = ExtraProgressBarWidth, math.MaxInt32
	} else {
		position += 7
		next = extraProgressThreshold(position, bar.itemsTotal)
	}
	frame, err := renderExtraProgress(position, label, bar.batch, bar.maxBytes)
	if err != nil {
		return "", false, err
	}
	bar.position, bar.itemsNext = position, next
	return frame, true, nil
}

// Stop returns the source's clearing frame and closes the owned state.
func (bar *ExtraProgressBar) Stop() (string, error) {
	if bar == nil {
		return "", ErrDimension
	}
	if bar.stopped {
		return "", ErrInvariant
	}
	frame, err := cleanExtraProgress(bar.batch, bar.maxBytes)
	if err != nil {
		return "", err
	}
	bar.stopped = true
	return frame, nil
}
