package env

import (
	"errors"
	"testing"
	"time"
)

// textUnmarshaler is a test type implementing encoding.TextUnmarshaler
type textUnmarshaler struct {
	Value string
}

func (t *textUnmarshaler) UnmarshalText(text []byte) error {
	t.Value = "unmarshaled:" + string(text)
	return nil
}

type errTextUnmarshaler struct{}

func (e *errTextUnmarshaler) UnmarshalText(_ []byte) error {
	return errors.New("unmarshal failed")
}

func TestUnmarshal_BasicTypes(t *testing.T) {
	type config struct {
		Name    string  `env:"NAME"`
		Age     int     `env:"AGE"`
		Active  bool    `env:"ACTIVE"`
		Score   float64 `env:"SCORE"`
		Balance float32 `env:"BALANCE"`
	}

	t.Setenv("NAME", "Alice")
	t.Setenv("AGE", "30")
	t.Setenv("ACTIVE", "true")
	t.Setenv("SCORE", "9.5")
	t.Setenv("BALANCE", "1.5")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "Alice" {
		t.Errorf("Name: got %q, want %q", cfg.Name, "Alice")
	}
	if cfg.Age != 30 {
		t.Errorf("Age: got %d, want 30", cfg.Age)
	}
	if !cfg.Active {
		t.Errorf("Active: got %v, want true", cfg.Active)
	}
	if cfg.Score != 9.5 {
		t.Errorf("Score: got %f, want 9.5", cfg.Score)
	}
	if cfg.Balance != 1.5 {
		t.Errorf("Balance: got %f, want 1.5", cfg.Balance)
	}
}

func TestUnmarshal_IntVariants(t *testing.T) {
	type config struct {
		I8  int8  `env:"I8"`
		I16 int16 `env:"I16"`
		I32 int32 `env:"I32"`
		I64 int64 `env:"I64"`
	}

	t.Setenv("I8", "127")
	t.Setenv("I16", "32767")
	t.Setenv("I32", "2147483647")
	t.Setenv("I64", "9223372036854775807")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.I8 != 127 {
		t.Errorf("I8: got %d, want 127", cfg.I8)
	}
	if cfg.I16 != 32767 {
		t.Errorf("I16: got %d, want 32767", cfg.I16)
	}
	if cfg.I32 != 2147483647 {
		t.Errorf("I32: got %d, want 2147483647", cfg.I32)
	}
	if cfg.I64 != 9223372036854775807 {
		t.Errorf("I64: got %d, want 9223372036854775807", cfg.I64)
	}
}

func TestUnmarshal_UintVariants(t *testing.T) {
	type config struct {
		U   uint   `env:"U"`
		U8  uint8  `env:"U8"`
		U16 uint16 `env:"U16"`
		U32 uint32 `env:"U32"`
		U64 uint64 `env:"U64"`
	}

	t.Setenv("U", "42")
	t.Setenv("U8", "255")
	t.Setenv("U16", "65535")
	t.Setenv("U32", "4294967295")
	t.Setenv("U64", "18446744073709551615")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.U != 42 {
		t.Errorf("U: got %d, want 42", cfg.U)
	}
	if cfg.U8 != 255 {
		t.Errorf("U8: got %d, want 255", cfg.U8)
	}
	if cfg.U16 != 65535 {
		t.Errorf("U16: got %d, want 65535", cfg.U16)
	}
	if cfg.U32 != 4294967295 {
		t.Errorf("U32: got %d, want 4294967295", cfg.U32)
	}
	if cfg.U64 != 18446744073709551615 {
		t.Errorf("U64: got %d, want 18446744073709551615", cfg.U64)
	}
}

func TestUnmarshal_Duration(t *testing.T) {
	type config struct {
		Timeout time.Duration `env:"TIMEOUT"`
	}

	t.Setenv("TIMEOUT", "5s")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout: got %v, want 5s", cfg.Timeout)
	}
}

func TestUnmarshal_InvalidDuration(t *testing.T) {
	type config struct {
		Timeout time.Duration `env:"TIMEOUT"`
	}

	t.Setenv("TIMEOUT", "notaduration")

	var cfg config
	if err := Unmarshal(&cfg); err == nil {
		t.Fatal("expected error for invalid duration, got nil")
	}
}

func TestUnmarshal_TextUnmarshaler(t *testing.T) {
	type config struct {
		Custom textUnmarshaler `env:"CUSTOM"`
	}

	t.Setenv("CUSTOM", "hello")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Custom.Value != "unmarshaled:hello" {
		t.Errorf("Custom: got %q, want %q", cfg.Custom.Value, "unmarshaled:hello")
	}
}

func TestUnmarshal_TextUnmarshaler_Error(t *testing.T) {
	type config struct {
		Custom errTextUnmarshaler `env:"CUSTOM"`
	}

	t.Setenv("CUSTOM", "hello")

	var cfg config
	if err := Unmarshal(&cfg); err == nil {
		t.Fatal("expected error from TextUnmarshaler, got nil")
	}
}

func TestUnmarshal_StringSlice(t *testing.T) {
	type config struct {
		Tags []string `env:"TAGS"`
	}

	t.Setenv("TAGS", "a,b,c")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"a", "b", "c"}
	if len(cfg.Tags) != len(want) {
		t.Fatalf("Tags len: got %d, want %d", len(cfg.Tags), len(want))
	}
	for i, v := range want {
		if cfg.Tags[i] != v {
			t.Errorf("Tags[%d]: got %q, want %q", i, cfg.Tags[i], v)
		}
	}
}

func TestUnmarshal_IntSlice(t *testing.T) {
	type config struct {
		Ports []int `env:"PORTS"`
	}

	t.Setenv("PORTS", "8080,9090,3000")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{8080, 9090, 3000}
	for i, v := range want {
		if cfg.Ports[i] != v {
			t.Errorf("Ports[%d]: got %d, want %d", i, cfg.Ports[i], v)
		}
	}
}

func TestUnmarshal_SkipsFieldsWithoutTag(t *testing.T) {
	type config struct {
		Name     string `env:"NAME"`
		NoTag    string
		EmptyTag string `env:""`
		Skipped  string `env:"-"`
	}

	t.Setenv("NAME", "Dan")
	t.Setenv("-", "should-be-ignored")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "Dan" {
		t.Errorf("Name: got %q, want %q", cfg.Name, "Dan")
	}
	if cfg.NoTag != "" {
		t.Errorf("NoTag should be empty, got %q", cfg.NoTag)
	}
	if cfg.Skipped != "" {
		t.Errorf("Skipped (tag \"-\") should be empty, got %q", cfg.Skipped)
	}
}

func TestUnmarshal_MissingEnvLeavesZeroValue(t *testing.T) {
	type config struct {
		Name string `env:"NAME_MISSING_XYZ"`
		Port int    `env:"PORT_MISSING_XYZ"`
	}

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "" {
		t.Errorf("Name should be empty, got %q", cfg.Name)
	}
	if cfg.Port != 0 {
		t.Errorf("Port should be 0, got %d", cfg.Port)
	}
}

func TestUnmarshal_InvalidInt(t *testing.T) {
	type config struct {
		Port int `env:"PORT"`
	}

	t.Setenv("PORT", "notanint")

	var cfg config
	if err := Unmarshal(&cfg); err == nil {
		t.Fatal("expected error for invalid int, got nil")
	}
}

func TestUnmarshal_InvalidBool(t *testing.T) {
	type config struct {
		Active bool `env:"ACTIVE"`
	}

	t.Setenv("ACTIVE", "notabool")

	var cfg config
	if err := Unmarshal(&cfg); err == nil {
		t.Fatal("expected error for invalid bool, got nil")
	}
}

func TestUnmarshal_InvalidFloat(t *testing.T) {
	type config struct {
		Score float64 `env:"SCORE"`
	}

	t.Setenv("SCORE", "notafloat")

	var cfg config
	if err := Unmarshal(&cfg); err == nil {
		t.Fatal("expected error for invalid float, got nil")
	}
}

func TestUnmarshal_MultipleErrors(t *testing.T) {
	type config struct {
		Port  int     `env:"PORT"`
		Score float64 `env:"SCORE"`
	}

	t.Setenv("PORT", "bad")
	t.Setenv("SCORE", "bad")

	var cfg config
	err := Unmarshal(&cfg)
	if err == nil {
		t.Fatal("expected errors, got nil")
	}

	var joinedErr interface{ Unwrap() []error }
	if !errors.As(err, &joinedErr) {
		t.Errorf("expected joined errors, got: %v", err)
	}
}

func TestUnmarshal_2DSliceReturnsError(t *testing.T) {
	type config struct {
		Matrix [][]int `env:"MATRIX"`
	}

	t.Setenv("MATRIX", "1,2,3")

	var cfg config
	if err := Unmarshal(&cfg); err == nil {
		t.Fatal("expected error for 2D slice, got nil")
	}
}

func TestUnmarshal_NonStructReturnsError(t *testing.T) {
	var s string
	if err := Unmarshal(&s); err == nil {
		t.Fatal("expected error for non-struct input, got nil")
	}

	var n int
	if err := Unmarshal(&n); err == nil {
		t.Fatal("expected error for non-struct input, got nil")
	}
}

func TestUnmarshal_NestedStruct(t *testing.T) {
	type db struct {
		Host string `env:"DB_HOST"`
		Port int    `env:"DB_PORT"`
	}
	type config struct {
		DB db `env:"db"`
	}

	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DB.Host != "localhost" {
		t.Errorf("DB.Host: got %q, want %q", cfg.DB.Host, "localhost")
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("DB.Port: got %d, want 5432", cfg.DB.Port)
	}
}

func TestUnmarshal_NestedPointerToStruct(t *testing.T) {
	type db struct {
		Host string `env:"DB_HOST"`
	}
	type config struct {
		DB *db `env:""`
	}

	t.Setenv("DB_HOST", "remotehost")

	var cfg config
	if err := Unmarshal(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DB == nil {
		t.Fatal("DB pointer should be allocated, got nil")
	}
	if cfg.DB.Host != "remotehost" {
		t.Errorf("DB.Host: got %q, want %q", cfg.DB.Host, "remotehost")
	}
}
