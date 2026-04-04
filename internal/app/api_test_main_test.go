package app_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	originalValue, hadOriginal := os.LookupEnv("WORKFLOW_WEB_FRONTEND")
	_ = os.Setenv("WORKFLOW_WEB_FRONTEND", "templates")

	code := m.Run()

	if hadOriginal {
		_ = os.Setenv("WORKFLOW_WEB_FRONTEND", originalValue)
	} else {
		_ = os.Unsetenv("WORKFLOW_WEB_FRONTEND")
	}

	os.Exit(code)
}
