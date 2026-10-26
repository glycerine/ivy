package abc

// Public utility queries from src/misc/zlib/zutil.c.

const (
	zlibPinnedVersion      = "1.2.5"
	zlibPinnedCompileFlags = uint64(169)
)

var zlibErrorMessages = [...]string{
	"incompatible version", // -6
	"buffer error",         // -5
	"insufficient memory",  // -4
	"data error",           // -3
	"stream error",         // -2
	"file error",           // -1
	"",                     // 0
	"stream end",           // 1
	"need dictionary",      // 2
}

// ZlibVersion implements zlibVersion for the pinned bundled source.
func ZlibVersion() string { return zlibPinnedVersion }

// ZlibCompileFlags implements zlibCompileFlags for the pinned native reference
// configuration. It is compatibility metadata, not a description of Go types.
func ZlibCompileFlags() uint64 { return zlibPinnedCompileFlags }

// ZlibError implements zError for its defined code range. The native routine
// indexes outside static storage for other codes; Go reports validity instead.
func ZlibError(code int) (message string, valid bool) {
	if code < -6 || code > 2 {
		return "", false
	}
	return zlibErrorMessages[code+6], true
}
