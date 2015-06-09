package util

import (
	"strings"
	"testing"
)

func TestSplitParams(t *testing.T) {
	originalList := []string{
		"BODY[HEADER.FIELDS (From Subject)]",
		"FLAGS",
	}
	params := strings.Join(originalList, " ")
	result := splitParams(params)
	for index, param := range originalList {
		if result[index] != param {
			t.Fatalf("Param %d does not match expected:\n"+
				"\tExpected: %s\n"+
				"\tActual:   %s", index, param, result[index])
		}
	}
	if len(result) > len(originalList) {
		t.Fatalf("Expected %d parameters but %d were split:\n%s",
			len(originalList), len(result), result)
	}
}
