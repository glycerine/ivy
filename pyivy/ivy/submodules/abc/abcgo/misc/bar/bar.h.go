package bar

const BAR_PROGRESS_USE = 1

// Bar_ProgressUpdate performs the same throttling as the original inline wrapper.
// The check is synchronized by the implementation rather than type-punning the
// progress object's first field as an int pointer.
func Bar_ProgressUpdate(p *Bar_Progress_t, current int32, label string) {
	Bar_ProgressUpdate_int(p, current, label)
}
