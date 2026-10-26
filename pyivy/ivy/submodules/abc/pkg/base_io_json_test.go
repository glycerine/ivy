package abc

// Tests for src/base/io/ioJson.c.

import (
	"bytes"
	"errors"
	"sync"
	"testing"
)

const abcJSONFixture = `{
 "layers" : [
  {
   "class_name" : "Conv2D",
   "config" : {
    "kernel_size" : [ "3" , "3" ],
    "strides" : [ "1" , "1" ],
    "filters" : "16"
   },
   "name" : "conv",
   "inbound_nodes" : [ [ "input" ] ]
  },
  {
   "class_name" : "Activation",
   "config" : { },
   "name" : "relu",
   "inbound_nodes" : [ [ "conv" ] ]
  }
 ]
}`

func TestABCJSONPreprocessParseFormat(t *testing.T) {
	preprocessed, err := PreprocessABCJSON([]byte(`{"a":["b"]}`), IOLimits{})
	if err != nil || string(preprocessed) != ` { "a": [ "b" ]  } ` {
		t.Fatalf("preprocess: %q, %v", preprocessed, err)
	}
	input := []byte(abcJSONFixture)
	document, err := ParseABCJSON(input, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if document.NodeCount() != 12 || document.StringCount() != 16 {
		t.Fatalf("arena sizes: nodes=%d strings=%d", document.NodeCount(), document.StringCount())
	}
	root, err := document.Node(0)
	if err != nil || root.Array || len(root.Entries) != 2 || root.Entries[0].Name != "layers" || root.Entries[1].IsName {
		t.Fatalf("root snapshot: %+v, %v", root, err)
	}
	root.Entries[0].Name = "changed"
	again, _ := document.Node(0)
	if again.Entries[0].Name != "layers" {
		t.Fatal("snapshot aliases document")
	}
	if _, err = document.Node(-1); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative node: %v", err)
	}
	formatted, err := FormatABCJSON(document, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	reparsed, err := ParseABCJSON(formatted, IOLimits{})
	if err != nil {
		t.Fatalf("formatted parse: %v\n%s", err, formatted)
	}
	formattedAgain, err := FormatABCJSON(reparsed, IOLimits{})
	if err != nil || !bytes.Equal(formatted, formattedAgain) {
		t.Fatalf("format is not stable: %v\n%s\n%s", err, formatted, formattedAgain)
	}
	for index := range input {
		input[index] = 'x'
	}
	formatted[0] = 'x'
	owned, err := FormatABCJSON(document, IOLimits{})
	if err != nil || owned[0] != '{' {
		t.Fatalf("ownership: %q, %v", owned, err)
	}
}

func TestABCJSONLayerExtraction(t *testing.T) {
	mappings := map[string]string{
		"InputLayer": "input  ", "Conv2D": "convo  ", "BatchNormalization": "batch  ",
		"Activation": "relu   ", "Add": "eltwise", "MaxPooling2D": "pool   ",
		"GlobalAveragePooling2D": "pool   ", "Dense": "fullcon", "ZeroPadding2D": "pad",
	}
	for input, want := range mappings {
		if got, ok := NNCLayerTypeString(input); !ok || got != want {
			t.Fatalf("mapping %q: %q, %v", input, got, ok)
		}
	}
	if got, ok := NNCLayerTypeString("Softmax"); ok || got != "" {
		t.Fatalf("unknown mapping: %q, %v", got, ok)
	}
	document, err := ParseABCJSON([]byte(abcJSONFixture), IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := ExtractABCJSONLayers(document, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	want := "# Data extracted from JSON file:\n\n  0 : convo     K=33 S=11 C=16 N=conv  input\n  1 : relu      N=relu  conv\n"
	if string(got) != want {
		t.Fatalf("extract mismatch:\n got %q\nwant %q", got, want)
	}
}

func TestABCJSONMalformedLimitsAndCancellation(t *testing.T) {
	malformed := []string{
		"", "scalar", "{", "}", "{ ]", "{ \"a\" }", "{} {}", "[ }",
	}
	for _, input := range malformed {
		if document, err := ParseABCJSON([]byte(input), IOLimits{}); document != nil || !errors.Is(err, ErrFormat) {
			t.Fatalf("malformed %q: document=%v err=%v", input, document, err)
		}
	}
	input := []byte(abcJSONFixture)
	for _, limits := range []IOLimits{
		{MaxInputBytes: len(input) - 1}, {MaxBytes: 1}, {MaxObjects: 1}, {MaxFanins: 1},
		{MaxDepth: 1}, {MaxOperations: 1}, {Stop: func() bool { return true }},
	} {
		if document, err := ParseABCJSON(input, limits); document != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatalf("parse limit %+v: document=%v err=%v", limits, document, err)
		}
	}
	document, err := ParseABCJSON(input, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	formatted, err := FormatABCJSON(document, IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if output, err := FormatABCJSON(document, IOLimits{MaxOutputBytes: len(formatted) - 1}); output != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("format output limit: %q, %v", output, err)
	}
	if output, err := ExtractABCJSONLayers(document, IOLimits{MaxOutputBytes: 1}); output != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("extract output limit: %q, %v", output, err)
	}
	if output, err := RoundTripABCJSON(input, IOLimits{Stop: func() bool { return true }}); output != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("round-trip cancellation: %q, %v", output, err)
	}
	if output, err := PreprocessABCJSON(input, IOLimits{MaxInputBytes: len(input) - 1}); output != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("preprocess limit: %q, %v", output, err)
	}
}

func TestABCJSONConcurrentFrozenReaders(t *testing.T) {
	document, err := ParseABCJSON([]byte(abcJSONFixture), IOLimits{})
	if err != nil {
		t.Fatal(err)
	}
	wantFormat, _ := FormatABCJSON(document, IOLimits{})
	wantExtract, _ := ExtractABCJSONLayers(document, IOLimits{})
	var wait sync.WaitGroup
	errorsFound := make(chan string, 32)
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			formatted, formatErr := FormatABCJSON(document, IOLimits{})
			extracted, extractErr := ExtractABCJSONLayers(document, IOLimits{})
			root, nodeErr := document.Node(0)
			if formatErr != nil || extractErr != nil || nodeErr != nil || !bytes.Equal(formatted, wantFormat) || !bytes.Equal(extracted, wantExtract) || len(root.Entries) != 2 {
				errorsFound <- "frozen reader mismatch"
			}
		}()
	}
	wait.Wait()
	close(errorsFound)
	for message := range errorsFound {
		t.Error(message)
	}
}

func FuzzABCJSONRoundTrip(f *testing.F) {
	for _, seed := range []string{abcJSONFixture, `{}`, `[]`, `[ "a" , { "b" : "c" } ]`, `{ "x" : [ "0" , "1" ] }`} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		limits := IOLimits{MaxInputBytes: 4096, MaxOutputBytes: 16384, MaxObjects: 256, MaxFanins: 256, MaxOperations: 32768, MaxBytes: 65536, MaxDepth: 64}
		if len(input) > limits.MaxInputBytes {
			t.Skip()
		}
		document, err := ParseABCJSON(input, limits)
		if err != nil {
			return
		}
		first, err := FormatABCJSON(document, limits)
		if err != nil {
			if errors.Is(err, ErrLimit) {
				return
			}
			t.Fatal(err)
		}
		again, err := ParseABCJSON(first, limits)
		if err != nil {
			t.Fatalf("formatted output rejected: %v\n%s", err, first)
		}
		second, err := FormatABCJSON(again, limits)
		if err != nil || !bytes.Equal(first, second) {
			t.Fatalf("unstable format: %v\n%s\n%s", err, first, second)
		}
	})
}
