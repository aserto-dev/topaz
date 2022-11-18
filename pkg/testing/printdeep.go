package testing

import (
	"encoding/json"
)

// PrettyPrintObject returns the JSON representation
// of any object; only useful for debugging.
func PrettyPrintObject(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}
