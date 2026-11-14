package bmc

import (
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/misc/util"
)

func bmcBridgeWriter(w io.Writer) io.Writer {
	if w == nil {
		return os.Stdout
	}
	return w
}

func writeBmcBridgeProgress(w io.Writer, output, frame int32) error {
	message := []byte(fmt.Sprintf("property: safe<%d>\nbug-free-depth: %d\n", output, frame))
	if util.Gia_ManToBridgeProgress(w, int32(len(message)), message) == 0 {
		return fmt.Errorf("writing BMC bridge progress")
	}
	return nil
}

// writeBmcBridgeResult writes one terminal property result. For SAT results,
// output selects the property's ID in the serialized CEX without mutating the
// model reused for other outputs.
func writeBmcBridgeResult(w io.Writer, result int32, cex *util.Abc_Cex_t, output int32) error {
	if result == 0 {
		if cex == nil {
			return fmt.Errorf("missing BMC bridge counterexample")
		}
		copy := *cex
		copy.IPo = output
		if util.Gia_ManFromBridgeCex(w, &copy) != nil {
			return fmt.Errorf("writing BMC bridge counterexample")
		}
		return nil
	}
	if util.Gia_ManToBridgeResult(w, result, nil, output) == 0 {
		return fmt.Errorf("writing BMC bridge result")
	}
	return nil
}
