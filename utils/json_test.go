package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemarshal(t *testing.T) {
	type testT struct {
		TestStr       string
		TestInt       int
		TestJSONCamel string `json:"testJsonCamel"`
		TestJSONSnake string `json:"test_json_snake"`
		TestStruct    struct {
			A, B string
		}
		TestArray []int
	}

	var tMap = map[string]interface{}{}
	tStruct := testT{
		TestStr:       "test-str",
		TestInt:       10,
		TestJSONCamel: "test-json-camel",
		TestJSONSnake: "test-json-snake",
		TestStruct:    struct{ A, B string }{A: "a", B: "b"},
		TestArray:     []int{0, 1, 2},
	}

	Remarshal(&tMap, tStruct)
	require.EqualValues(t, map[string]interface{}{
		"TestStr":         "test-str",
		"TestInt":         float64(10),
		"testJsonCamel":   "test-json-camel",
		"test_json_snake": "test-json-snake",
		"TestStruct":      map[string]interface{}{"A": "a", "B": "b"},
		"TestArray":       []interface{}{float64(0), float64(1), float64(2)},
	}, tMap)

	ts := testT{}
	Remarshal(&ts, tMap)
	require.EqualValues(t, tStruct, ts)
}
