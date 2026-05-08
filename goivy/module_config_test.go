package goivy

import "testing"

func TestConfigBackendDefaultAndNormalization(t *testing.T) {
	cfg := NewConfig()
	if cfg.BackendName != BackendCGo {
		t.Fatalf("NewConfig().BackendName = %q, want %q", cfg.BackendName, BackendCGo)
	}
	if cfg.Backend == nil {
		t.Fatal("NewConfig().Backend is nil")
	}
	if got := cfg.Backend.Z3BackendName(); got != BackendCGo {
		t.Fatalf("NewConfig().Backend.Z3BackendName() = %q, want %q", got, BackendCGo)
	}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: BackendCGo},
		{name: "native", in: "cgo", want: BackendCGo},
		{name: "wazero", in: "wazero", want: BackendWazero},
		{name: "browser", in: "jsbrowser", want: BackendJSBrowser},
		{name: "trim and lower", in: " WAZERO ", want: BackendWazero},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeBackendName(tc.in)
			if err != nil {
				t.Fatalf("NormalizeBackendName(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeBackendName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestConfigBackendRejectsUnknownValue(t *testing.T) {
	if _, err := NormalizeBackendName("native"); err == nil {
		t.Fatal("NormalizeBackendName(\"native\") succeeded, want error")
	}
}

func TestConfigSetBackendNameResolvesInterface(t *testing.T) {
	cfg := NewConfig()
	if err := cfg.SetBackendName(" WAZERO "); err != nil {
		t.Fatalf("SetBackendName(wazero): %v", err)
	}
	if cfg.BackendName != BackendWazero {
		t.Fatalf("BackendName = %q, want %q", cfg.BackendName, BackendWazero)
	}
	if cfg.Backend == nil {
		t.Fatal("Backend is nil after SetBackendName")
	}
	if got := cfg.Backend.Z3BackendName(); got != BackendWazero {
		t.Fatalf("Backend.Z3BackendName() = %q, want %q", got, BackendWazero)
	}
}
