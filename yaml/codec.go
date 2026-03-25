// Package yaml wraps gopkg.in/yaml.v3 and provides Marshal/Unmarshal helpers
// with a consistent interface shared across the codec packages.
package yaml

import (
	"gopkg.in/yaml.v3"
)

func Marshall(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
