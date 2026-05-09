//go:build wasip1

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"unsafe"

	goivy "github.com/glycerine/ivy/goivy"
)

type checkMeta struct {
	Filename string            `json:"filename"`
	Params   map[string]string `json:"params"`
}

var allocations = make(map[uint32][]byte)

func main() {}

//go:wasmexport goivy_check_alloc
func goivyCheckAlloc(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	buf := make([]byte, n)
	ptr := uint32(uintptr(unsafe.Pointer(&buf[0])))
	allocations[ptr] = buf
	return ptr
}

//go:wasmexport goivy_check_free
func goivyCheckFree(ptr uint32, n uint32) {
	_ = n
	delete(allocations, ptr)
}

//go:wasmexport goivy_check_run
func goivyCheckRun(specPtr, specLen, metaPtr, metaLen uint32) (code int32) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n", r)
			code = 2
		}
	}()

	spec := wasmString(specPtr, specLen)
	metaRaw := wasmString(metaPtr, metaLen)
	meta := checkMeta{Filename: "browser_input.ivy"}
	if metaRaw != "" {
		if err := json.Unmarshal([]byte(metaRaw), &meta); err != nil {
			fmt.Fprintf(os.Stderr, "error: bad goivy_check metadata: %v\n", err)
			return 1
		}
	}
	if meta.Filename == "" {
		meta.Filename = "browser_input.ivy"
	}
	if meta.Params == nil {
		meta.Params = make(map[string]string)
	}

	cfg := goivy.NewConfig()
	if err := goivy.ApplyIvyCheckParams(cfg, meta.Params); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if err := goivy.StartSourceWithConfig(meta.Filename, spec, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func wasmString(ptr, n uint32) string {
	if ptr == 0 || n == 0 {
		return ""
	}
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), int(n))
	return string(bytes)
}
