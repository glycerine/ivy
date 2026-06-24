import { type SourceFile } from "./diagnostics.js";

export function stubSourcePackageFiles(importPath: string): SourceFile[] | undefined {
  switch (importPath) {
    case "github.com/klauspost/cpuid/v2":
      return [{
        filename: "gojr:stub/github.com/klauspost/cpuid/v2/cpuid.go",
        source: klauspostCpuidV2StubSource
      }];
    case "runtime/pprof":
      return [{
        filename: "gojr:stub/runtime/pprof/pprof.go",
        source: runtimePprofStubSource
      }];
    default:
      return undefined;
  }
}

export function isStubSourcePackageImport(importPath: string): boolean {
  return stubSourcePackageFiles(importPath) !== undefined;
}

export function isStubSourcePackageStandardLibrary(importPath: string): boolean {
  return importPath === "runtime/pprof";
}

const klauspostCpuidV2StubSource = `package cpuid

// FeatureID is the ID of a specific CPU feature. Go-junior deliberately
// exposes a deterministic JavaScript target profile rather than probing native
// hardware.
type FeatureID int

const (
  UNKNOWN FeatureID = -1
  AVX2 FeatureID = 8
  AVX512F FeatureID = 16
)

type CPUInfo struct{}

var CPU CPUInfo

func (c CPUInfo) Supports(features ...FeatureID) bool {
  return false
}
`;

const runtimePprofStubSource = `package pprof

import (
  "context"
  "io"
)

type LabelSet struct{}

type Profile struct{}

func NewProfile(name string) *Profile {
  panic("gojr error: runtime/pprof.NewProfile not implemented")
}

func Lookup(name string) *Profile {
  panic("gojr error: runtime/pprof.Lookup not implemented")
}

func Profiles() []*Profile {
  panic("gojr error: runtime/pprof.Profiles not implemented")
}

func (p *Profile) Name() string {
  panic("gojr error: runtime/pprof.(*Profile).Name not implemented")
}

func (p *Profile) Count() int {
  panic("gojr error: runtime/pprof.(*Profile).Count not implemented")
}

func (p *Profile) Add(value any, skip int) {
  panic("gojr error: runtime/pprof.(*Profile).Add not implemented")
}

func (p *Profile) Remove(value any) {
  panic("gojr error: runtime/pprof.(*Profile).Remove not implemented")
}

func (p *Profile) WriteTo(w io.Writer, debug int) error {
  panic("gojr error: runtime/pprof.(*Profile).WriteTo not implemented")
}

func WriteHeapProfile(w io.Writer) error {
  panic("gojr error: runtime/pprof.WriteHeapProfile not implemented")
}

func StartCPUProfile(w io.Writer) error {
  panic("gojr error: runtime/pprof.StartCPUProfile not implemented")
}

func StopCPUProfile() {
  panic("gojr error: runtime/pprof.StopCPUProfile not implemented")
}

func WithLabels(ctx context.Context, labels LabelSet) context.Context {
  panic("gojr error: runtime/pprof.WithLabels not implemented")
}

func Labels(args ...string) LabelSet {
  panic("gojr error: runtime/pprof.Labels not implemented")
}

func Label(ctx context.Context, key string) (string, bool) {
  panic("gojr error: runtime/pprof.Label not implemented")
}

func ForLabels(ctx context.Context, f func(key, value string) bool) {
  panic("gojr error: runtime/pprof.ForLabels not implemented")
}

func SetGoroutineLabels(ctx context.Context) {
  panic("gojr error: runtime/pprof.SetGoroutineLabels not implemented")
}

func Do(ctx context.Context, labels LabelSet, f func(context.Context)) {
  panic("gojr error: runtime/pprof.Do not implemented")
}
`;
