// Package test contains ABC's empty test-package extension hooks.
package test

import abcmain "github.com/glycerine/abcgo/base/main"

// Test_Init faithfully preserves the empty initialization hook in
// src/base/test/test.c. The original hook registers no commands or state.
func Test_Init(_ *abcmain.Frame) {}

// Test_End faithfully preserves the empty shutdown hook in the same source.
// Both hooks accept nil because the original implementations never use pAbc.
func Test_End(_ *abcmain.Frame) {}
