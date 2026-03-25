// Package json wraps encoding/json and marshals values with two-space
// indentation for human-readable output.
package json

import "encoding/json"

func Marshall(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, &v)
}
