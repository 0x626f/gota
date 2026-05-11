package json

import "testing"

func TestUnmarshal_PopulatesStructPointer(t *testing.T) {
	type config struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}

	var cfg config
	if err := Unmarshal([]byte(`{"name":"api","port":8080}`), &cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "api" || cfg.Port != 8080 {
		t.Fatalf("Unmarshal did not populate struct: %+v", cfg)
	}
}
