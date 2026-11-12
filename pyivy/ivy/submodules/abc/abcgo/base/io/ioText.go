package io

import stdio "io"

func writeText(w stdio.Writer, text string) error {
	n, err := stdio.WriteString(w, text)
	if err == nil && n != len(text) {
		return stdio.ErrShortWrite
	}
	return err
}
