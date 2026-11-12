package abcmain

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// ModelChecker connects the competition driver to the real AIG/BMC/SEC engines
// supplied by component packages. No verification result is guessed locally.
type ModelChecker interface {
	ReadAiger(string) (any, error)
	RegisterCount(any) int
	BMC(any, int, int, int) (status, depth int32, err error)
	SEC(any, time.Duration, int) (int32, error)
	WriteCounterexample(io.Writer, any) error
	Close(any) error
}
type ModelCheckOptions struct {
	BMCOnly    bool
	SECTimeout time.Duration
}

func RunModelChecking(f *Frame, path string, options ModelCheckOptions) (int32, error) {
	if f.ModelChecker == nil {
		return -1, errors.New("abcmain: model checking requires a ModelChecker adapter")
	}
	model, err := f.ModelChecker.ReadAiger(path)
	if err != nil {
		return -1, err
	}
	defer f.ModelChecker.Close(model)
	status := int32(-1)
	if f.ModelChecker.RegisterCount(model) > 0 {
		status, _, err = f.ModelChecker.BMC(model, 50, 500000, 10000)
		if err != nil {
			return -1, err
		}
	}
	if !options.BMCOnly && status != 0 {
		limit := options.SECTimeout
		if limit == 0 {
			limit = 600 * time.Second
		}
		status, err = f.ModelChecker.SEC(model, limit, 4)
		if err != nil {
			return -1, err
		}
	}
	if status == -1 && f.ModelChecker.RegisterCount(model) > 0 {
		status, _, err = f.ModelChecker.BMC(model, 200, 500000, 10000000)
		if err != nil {
			return -1, err
		}
		if status != 0 {
			status = -1
		}
	}
	switch status {
	case 0:
		fmt.Fprintln(f.Out, "1")
		if err := f.ModelChecker.WriteCounterexample(f.Out, model); err != nil {
			return -1, err
		}
	case 1:
		fmt.Fprintln(f.Out, "0")
	default:
		fmt.Fprintln(f.Out, "2")
	}
	return status, nil
}
func Abc_MCMain(args []string) int32 {
	f := Abc_FrameGetGlobalFrame()
	if len(args) != 2 {
		fmt.Fprintf(f.Err, "usage: %s <file.aig>\n", firstArg(args))
		return 1
	}
	if _, err := os.Stat(args[1]); err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	status, err := RunModelChecking(f, args[1], ModelCheckOptions{})
	if err != nil {
		fmt.Fprintln(f.Err, err)
		return 1
	}
	if status == 0 {
		return 10
	}
	if status == 1 {
		return 20
	}
	return 0
}
