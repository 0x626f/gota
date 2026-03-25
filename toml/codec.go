// Package toml wraps github.com/pelletier/go-toml/v2 and provides
// Marshal/Unmarshal helpers with a consistent interface shared across the
// codec packages.
package toml

import (
	"github.com/pelletier/go-toml/v2"
)

func Marshall(v interface{}) ([]byte, error) {
	return toml.Marshal(v)
}

func Unmarshal(data []byte, v any) error {
	return toml.Unmarshal(data, v)
}
