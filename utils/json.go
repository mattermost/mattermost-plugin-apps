package utils

import (
	"encoding/json"
)

func ToJSON(in any) string {
	bb, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return string(bb)
}

func Pretty(in any) string {
	bb, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return ""
	}
	return string(bb)
}

func Remarshal(dst, src any) {
	data, _ := json.Marshal(src)
	_ = json.Unmarshal(data, dst)
}
