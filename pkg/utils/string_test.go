package utils

import (
	"testing"

	"github.com/soulnov23/go-tool/pkg/json/jsoniter"
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
	t.Log(jsoniter.Stringify(FlattenMap(recordMap)))
}
