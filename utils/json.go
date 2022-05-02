package utils

import (
	"encoding/json"
)

func ToJSON(in interface{}) string {
	bb, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return string(bb)
}

func Pretty(in interface{}) string {
	bb, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return ""
	}
	return string(bb)
}

func PrettyBlock(in interface{}) string {
	return "\n```json\n" + Pretty(in) + "\n```\n"
}

func Remarshal(dst, src interface{}) {
	data, _ := json.Marshal(src)
	_ = json.Unmarshal(data, dst)
}
