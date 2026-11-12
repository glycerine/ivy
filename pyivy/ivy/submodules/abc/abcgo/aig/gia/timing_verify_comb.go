package gia

import (
	"fmt"
	ceccore "github.com/glycerine/abcgo/proof/cec/core"
	"time"
)

// Combinational boxed verification treats every CI and CO, including latch
// terminals, as an ordinary input or output, matching Gia_ManMiter's mode.
type giaCombVerificationNetwork struct{ *Gia_Man_t }

func (giaCombVerificationNetwork) RegisterCount() int { return 0 }

func giaVerifyPreparedCombinational(spec, impl *Gia_Man_t, conflicts, seconds, verbose int32) int32 {
	if conflicts < 0 || seconds < 0 {
		return -1
	}
	options := ceccore.DefaultCECOptions()
	options.ConflictLimit = int64(conflicts)
	if seconds > 0 {
		options.Deadline = time.Now().Add(time.Duration(seconds) * time.Second)
	}
	result, err := ceccore.VerifyTwo(giaCombVerificationNetwork{spec}, giaCombVerificationNetwork{impl}, options)
	if err != nil {
		if verbose != 0 {
			fmt.Printf("Boxed verification failed: %v\n", err)
		}
		return -1
	}
	if result.Status == ceccore.Different && result.Counterexample != nil {
		fmt.Printf("Verification failed for at least one output (%d).\n", result.Counterexample.Output)
	}
	return int32(result.Status)
}
