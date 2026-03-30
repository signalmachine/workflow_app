package envload

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileIfPresentSetsMissingValuesOnly(t *testing.T) {
	t.Setenv("ENVLOAD_EXISTING_URL", "postgres://existing")

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "ENVLOAD_EXISTING_URL=postgres://from-file\nENVLOAD_LOADED_URL=postgres://test\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if err := LoadFileIfPresent(path); err != nil {
		t.Fatalf("LoadFileIfPresent() error = %v", err)
	}

	if got := os.Getenv("ENVLOAD_EXISTING_URL"); got != "postgres://existing" {
		t.Fatalf("ENVLOAD_EXISTING_URL = %q, want existing value", got)
	}
	if got := os.Getenv("ENVLOAD_LOADED_URL"); got != "postgres://test" {
		t.Fatalf("ENVLOAD_LOADED_URL = %q, want loaded value", got)
	}
}

func TestLoadFileIfPresentSupportsExportAndQuotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "export APP_LISTEN_ADDR=\"127.0.0.1:18080\"\nOPENAI_MODEL='gpt-4o'\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if err := LoadFileIfPresent(path); err != nil {
		t.Fatalf("LoadFileIfPresent() error = %v", err)
	}

	if got := os.Getenv("APP_LISTEN_ADDR"); got != "127.0.0.1:18080" {
		t.Fatalf("APP_LISTEN_ADDR = %q, want quoted value", got)
	}
	if got := os.Getenv("OPENAI_MODEL"); got != "gpt-4o" {
		t.Fatalf("OPENAI_MODEL = %q, want single-quoted value", got)
	}
}

func TestLoadFileIfPresentIgnoresMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := LoadFileIfPresent(path); err != nil {
		t.Fatalf("LoadFileIfPresent() error = %v, want nil", err)
	}
}
