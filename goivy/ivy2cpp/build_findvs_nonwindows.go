//go:build !windows

package ivy2cpp

import "errors"

func defaultFindVS() (vsInfo, error) {
	return vsInfo{}, errors.New("vswhere unavailable on non-Windows")
}
