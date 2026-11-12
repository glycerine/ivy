package abcmain

import (
	"fmt"
	"sync"
)

var initializerRegistry struct {
	sync.RWMutex
	list []*Abc_FrameInitializer_t
}

func Abc_FrameAddInitializer(initializer *Abc_FrameInitializer_t) {
	if initializer == nil {
		panic("abcmain: nil initializer")
	}
	initializerRegistry.Lock()
	defer initializerRegistry.Unlock()
	for _, old := range initializerRegistry.list {
		if old == initializer {
			return
		}
	}
	initializerRegistry.list = append(initializerRegistry.list, initializer)
}
func Abc_FrameInit(frame *Frame) {
	frame.mu.RLock()
	active := frame.initialized && !frame.ended
	frame.mu.RUnlock()
	if active {
		return
	}
	frame.lifecycle.Lock()
	defer frame.lifecycle.Unlock()
	frame.mu.Lock()
	if frame.closed {
		frame.mu.Unlock()
		panic("abcmain: closed frame")
	}
	if frame.initialized && !frame.ended {
		frame.mu.Unlock()
		return
	}
	frame.initialized, frame.ended = true, false
	frame.mu.Unlock()
	initializerRegistry.RLock()
	initializers := append([]*Abc_FrameInitializer_t(nil), initializerRegistry.list...)
	initializerRegistry.RUnlock()
	frame.mu.Lock()
	frame.initializers = initializers
	frame.mu.Unlock()
	for _, initializer := range initializers {
		if initializer.Init != nil {
			initializer.Init(frame)
		}
	}
	if err := Libs_Init(frame); err != nil {
		fmt.Fprintln(frame.Err, err)
	}
}
func Abc_FrameEnd(frame *Frame) {
	if frame == nil {
		return
	}
	frame.mu.RLock()
	inactive := !frame.initialized || frame.ended
	frame.mu.RUnlock()
	if inactive {
		return
	}
	frame.lifecycle.Lock()
	defer frame.lifecycle.Unlock()
	frame.mu.Lock()
	if !frame.initialized || frame.ended {
		frame.mu.Unlock()
		return
	}
	frame.ended = true
	initializers := append([]*Abc_FrameInitializer_t(nil), frame.initializers...)
	frame.initializers = nil
	frame.mu.Unlock()
	Libs_End(frame)
	for i := len(initializers) - 1; i >= 0; i-- {
		if initializers[i].Destroy != nil {
			initializers[i].Destroy(frame)
		}
	}
}
