package gia

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Gia_ManVerifyWithBoxes compares a design with its AIGER specification.
// The sequential path uses bounded explicit state exploration and reports
// undecided if it cannot complete within its resource limits.
func Gia_ManVerifyWithBoxes(impl *Gia_Man_t, conflicts, seconds, sequential, dump, verbose int32, specPath string) int32 {
	if conflicts < 0 || seconds < 0 {
		return -1
	}
	spec, design, err := giaPrepareBoxVerification(impl, specPath, sequential)
	if err != nil {
		fmt.Printf("Boxed verification: %v.\n", err)
		return -1
	}
	defer Gia_ManStop(spec)
	defer Gia_ManStop(design)
	if dump != 0 {
		name := specPath
		if name == "" {
			name = impl.pSpec
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		left, right := base+"_spec.aig", base+"_impl.aig"
		if err := Gia_AigerWrite(spec, left, 0, 0, 0); err != nil {
			fmt.Printf("Cannot dump specification: %v\n", err)
			return -1
		}
		if err := Gia_AigerWrite(design, right, 0, 0, 0); err != nil {
			fmt.Printf("Cannot dump implementation: %v\n", err)
			return -1
		}
		fmt.Printf("Dumped two parts of the miter into files %q and %q.\n", left, right)
	}
	if sequential == 0 {
		return giaVerifyPreparedCombinational(spec, design, conflicts, seconds, verbose)
	}
	miter := Gia_ManMiter(spec, design, 0, 0, 1, 0, verbose)
	if miter == nil {
		return -1
	}
	defer Gia_ManStop(miter)
	var deadline time.Time
	if seconds > 0 {
		deadline = time.Now().Add(time.Duration(seconds) * time.Second)
	}
	return giaVerifySequentialMiter(miter, 1<<20, deadline)
}
