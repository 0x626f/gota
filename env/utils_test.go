package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnv_Set(t *testing.T) {
	t.Setenv("TEST_VAR", "hello")

	if got := GetEnv("TEST_VAR"); got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestGetEnv_Missing_NoDefault(t *testing.T) {
	if got := GetEnv("TEST_VAR_MISSING_XYZ"); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestGetEnv_Missing_WithDefault(t *testing.T) {
	if got := GetEnv("TEST_VAR_MISSING_XYZ", "fallback"); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}

func TestGetEnv_Set_IgnoresDefault(t *testing.T) {
	t.Setenv("TEST_VAR", "real")

	if got := GetEnv("TEST_VAR", "fallback"); got != "real" {
		t.Errorf("got %q, want %q", got, "real")
	}
}

func TestGetEnv_EmptyValue_ReturnsEmpty(t *testing.T) {
	t.Setenv("TEST_VAR", "")

	if got := GetEnv("TEST_VAR", "fallback"); got != "" {
		t.Errorf("got %q, want empty string (var is set but empty)", got)
	}
}

func TestSetEnv(t *testing.T) {
	t.Setenv("TEST_SET_VAR", "")

	if err := SetEnv("TEST_SET_VAR", "value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("TEST_SET_VAR"); got != "value" {
		t.Errorf("got %q, want %q", got, "value")
	}
}

func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.env")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_BasicKeyValue(t *testing.T) {
	path := writeEnvFile(t, "LOAD_HOST=localhost\nLOAD_PORT=5432\n")

	if err := Load(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("LOAD_HOST"); got != "localhost" {
		t.Errorf("LOAD_HOST: got %q, want %q", got, "localhost")
	}
	if got := os.Getenv("LOAD_PORT"); got != "5432" {
		t.Errorf("LOAD_PORT: got %q, want %q", got, "5432")
	}
}

func TestLoad_SkipsComments(t *testing.T) {
	path := writeEnvFile(t, "# this is a comment\nLOAD_NAME=alice\n")

	if err := Load(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("LOAD_NAME"); got != "alice" {
		t.Errorf("LOAD_NAME: got %q, want %q", got, "alice")
	}
}

func TestLoad_SkipsLinesWithoutSeparator(t *testing.T) {
	path := writeEnvFile(t, "NOEQUALSSIGN\nLOAD_VALID=yes\n")

	if err := Load(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("NOEQUALSSIGN"); got != "" {
		t.Errorf("NOEQUALSSIGN should not be set, got %q", got)
	}
	if got := os.Getenv("LOAD_VALID"); got != "yes" {
		t.Errorf("LOAD_VALID: got %q, want %q", got, "yes")
	}
}

func TestLoad_ValueWithEqualsSign(t *testing.T) {
	path := writeEnvFile(t, "LOAD_TOKEN=abc=def\n")

	if err := Load(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("LOAD_TOKEN"); got != "abc=def" {
		t.Errorf("LOAD_TOKEN: got %q, want %q", got, "abc=def")
	}
}

func TestLoad_MultipleFiles(t *testing.T) {
	p1 := writeEnvFile(t, "LOAD_FIRST=one\n")
	p2 := writeEnvFile(t, "LOAD_SECOND=two\n")

	if err := Load(p1, p2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("LOAD_FIRST"); got != "one" {
		t.Errorf("LOAD_FIRST: got %q, want %q", got, "one")
	}
	if got := os.Getenv("LOAD_SECOND"); got != "two" {
		t.Errorf("LOAD_SECOND: got %q, want %q", got, "two")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nonexistent.env")
	if err := Load(missing); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	path := writeEnvFile(t, "")

	if err := Load(path); err != nil {
		t.Fatalf("unexpected error for empty file: %v", err)
	}
}
