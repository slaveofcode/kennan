package content

import "encoding/json"

func IsJSON(b []byte) bool {
	var j map[string]interface{}
	return json.Unmarshal(b, &j) == nil
}
