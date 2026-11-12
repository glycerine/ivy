// The C util_hack.h was a declaration-only compatibility shim for Extra's CPU
// time, resource limit and allocation-failure hooks. Go callers import the
// owning package directly; declarations and preprocessor aliases need no Go
// runtime definitions. Clock duration helpers live in abc_global.go.
package util
