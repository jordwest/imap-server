package types

import (
	"testing"
)

func TestCombine(t *testing.T) {
	c1 := CombineFlags(Seen, Draft|Deleted)

	if c1 != (Seen | Draft | Deleted) {
		t.Errorf("Combined flags do not match")
	}
}

func TestHasFlags(t *testing.T) {
	c1 := (Seen | Draft | Deleted)

	if !c1.HasFlags(Seen) {
		t.Errorf("HasFlags should return true")
	}

	if !c1.HasFlags(Seen | Draft | Deleted) {
		t.Errorf("HasFlags should return true")
	}

	if c1.HasFlags(Seen | Draft | Deleted | Recent) {
		t.Errorf("HasFlags should return false")
	}

}
