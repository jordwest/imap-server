package types

import "testing"

func TestCombine(t *testing.T) {
	c1 := CombineFlags(FlagSeen, FlagDraft|FlagDeleted)

	if c1 != (FlagSeen | FlagDraft | FlagDeleted) {
		t.Errorf("Combined flags do not match")
	}
}

func TestSetReset(t *testing.T) {
	c1 := (FlagSeen | FlagDraft | FlagDeleted)
	c1 = c1.ResetFlags(FlagDraft)

	expected := (FlagSeen | FlagDeleted)
	if c1 != expected {
		t.Errorf("Expected %d, Actual %d", expected, c1)
	}

	// What if we try to remove a flag that already doesn't exist?
	c1 = c1.ResetFlags(FlagDraft)
	if c1 != expected {
		t.Errorf("Expected %d, Actual %d", expected, c1)
	}

}

func TestHasFlags(t *testing.T) {
	c1 := (FlagSeen | FlagDraft | FlagDeleted)

	if !c1.HasFlags(FlagSeen) {
		t.Errorf("HasFlags should return true")
	}

	if !c1.HasFlags(FlagSeen | FlagDraft | FlagDeleted) {
		t.Errorf("HasFlags should return true")
	}

	if c1.HasFlags(FlagSeen | FlagDraft | FlagDeleted | FlagRecent) {
		t.Errorf("HasFlags should return false")
	}

}

func TestFlagsFromString(t *testing.T) {
	c1 := FlagsFromString("\\Deleted \\Seen")

	expected := Flags(FlagDeleted | FlagSeen)

	if c1 != expected {
		t.Errorf("Expected %d, Actual %d", expected, c1)
	}
}
