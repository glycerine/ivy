package abc

// Source-shaped facade for src/misc/bar/bar.c and bar.h.

const BarProgressUse = 1

type BarProgress = ExtraProgressBar

func StartBarProgress(itemsTotal int, options ExtraProgressBarOptions) (*BarProgress, string, error) {
	return StartExtraProgressBar(itemsTotal, options)
}

// UpdateBarProgress represents both Bar_ProgressUpdate and its _int target.
// A nil handle remains the source's disabled-progress no-op.
func UpdateBarProgress(bar *BarProgress, itemsCurrent int, label string) (string, bool, error) {
	if bar == nil {
		return "", false, nil
	}
	return bar.Update(itemsCurrent, label)
}

// StopBarProgress preserves the source's nil-handle no-op.
func StopBarProgress(bar *BarProgress) (string, error) {
	if bar == nil {
		return "", nil
	}
	return bar.Stop()
}
