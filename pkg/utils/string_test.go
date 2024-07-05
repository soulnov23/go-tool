package utils

import (
	"testing"
)

func TestFlatten(t *testing.T) {
	recordMap := map[string]any{
		"a": "a",
		"b": "b",
		"c": map[string]any{
			"a": "c_a",
			"b": "c_b",
		},
	}
	t.Log(Stringify(FlattenMap(recordMap)))
}
