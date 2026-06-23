package main

import "testing"

func TestShouldPrintValueSuppressesOnlyActualNil(t *testing.T) {
	tests := []struct {
		name   string
		result evalResult
		want   bool
	}{
		{
			name:   "empty",
			result: evalResult{},
			want:   false,
		},
		{
			name:   "actual nil",
			result: evalResult{Value: "<nil>", ValueIsNil: true},
			want:   false,
		},
		{
			name:   "string containing nil spelling",
			result: evalResult{Value: "<nil>"},
			want:   true,
		},
		{
			name:   "normal value",
			result: evalResult{Value: "42"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPrintValue(tt.result); got != tt.want {
				t.Fatalf("shouldPrintValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
