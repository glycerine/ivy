//go:build !xtracer_off && wasip1

package xtracer

// NormalizeLine is intentionally a no-op in browser/wasip1 builds.
//
// Native xtracer normalizes local filesystem paths to keep textual
// conformance diffs stable. Browser-side goivy does not have the same local
// filesystem layout, and the trace transport should not depend on probing one.
func NormalizeLine(line string) string {
	return line
}
