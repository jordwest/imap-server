package types

import (
	"testing"
)

func TestCombine(t *testing.T) {
	c1 := CombineFlags(FlagSeen, FlagDraft|FlagDeleted)

	if c1 != (FlagSeen | FlagDraft | FlagDeleted) {
		t.Errorf("Combined flags do not match")
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
