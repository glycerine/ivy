package abci

import "testing"

func TestExtraUtilGetoptBridgeResetAndOptions(t *testing.T) {
	argv := []string{"abc", "-a", "-n", "17", "file"}
	Extra_UtilGetoptReset()
	if globalUtilOptind != 0 {
		t.Fatalf("reset index = %d, want 0", globalUtilOptind)
	}
	if got := Extra_UtilGetopt(int32(len(argv)), argv, "an:"); got != 'a' {
		t.Fatalf("first option = %d, want 'a'", got)
	}
	if globalUtilOptind != 2 {
		t.Fatalf("index after -a = %d, want 2", globalUtilOptind)
	}
	if got := Extra_UtilGetopt(int32(len(argv)), argv, "an:"); got != 'n' {
		t.Fatalf("second option = %d, want 'n'", got)
	}
	if got := Extra_UtilGetoptArg(); got != "17" {
		t.Fatalf("-n argument = %q, want %q", got, "17")
	}
	if globalUtilOptind != 4 {
		t.Fatalf("index after -n 17 = %d, want 4", globalUtilOptind)
	}
	if got := Extra_UtilGetopt(int32(len(argv)), argv, "an:"); got != -1 {
		t.Fatalf("option after positional argument = %d, want -1", got)
	}
	if globalUtilOptind != 4 {
		t.Fatalf("index after positional argument = %d, want 4", globalUtilOptind)
	}
}

func TestExtraUtilGetoptBridgeEndOfOptions(t *testing.T) {
	argv := []string{"abc", "--", "-a"}
	Extra_UtilGetoptReset()
	if got := Extra_UtilGetopt(int32(len(argv)), argv, "a"); got != -1 {
		t.Fatalf("option at -- = %d, want -1", got)
	}
	if globalUtilOptind != 2 {
		t.Fatalf("index after -- = %d, want 2", globalUtilOptind)
	}
	if got := Extra_UtilGetopt(int32(len(argv)), argv, "a"); got != 'a' {
		t.Fatalf("option after -- on subsequent call = %d, want 'a'", got)
	}
}

func TestExtraUtilGetoptBridgeManualIndexAdjustment(t *testing.T) {
	argv := []string{"abc", "-a", "operand", "-b"}
	Extra_UtilGetoptReset()
	if got := Extra_UtilGetopt(int32(len(argv)), argv, "ab"); got != 'a' {
		t.Fatalf("first option = %d, want 'a'", got)
	}
	globalUtilOptind++ // Simulate a C command consuming argv[2] itself.
	if got := Extra_UtilGetopt(int32(len(argv)), argv, "ab"); got != 'b' {
		t.Fatalf("option after manual index adjustment = %d, want 'b'", got)
	}
	if globalUtilOptind != 4 {
		t.Fatalf("final index = %d, want 4", globalUtilOptind)
	}
}
