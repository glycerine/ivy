package util

// Temporary-file and formatting utilities from utilFile.c. CreateTemp avoids
// the C implementation's close/unlink/reopen race while retaining suffixes.
import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func realTimeAbs() uint64 { return uint64(time.Now().UnixNano()) }
func tmpFile(prefix, suffix string, outName *string) (*os.File, error) {
	file, err := os.CreateTemp(filepath.Dir(prefix), filepath.Base(prefix)+"*"+suffix)
	if outName != nil {
		*outName = ""
		if err == nil {
			*outName = file.Name()
		}
	}
	return file, err
}
func vnsprintf(format string, args []any) string { return fmt.Sprintf(format, args...) }
func nsprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }
