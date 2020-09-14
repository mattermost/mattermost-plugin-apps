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
