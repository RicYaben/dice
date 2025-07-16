package sdk

import "encoding/json"

func AsJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
