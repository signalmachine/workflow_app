package envload

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultDotEnvPath = ".env"

// LoadDefaultIfPresent loads .env from the current working directory without
// overriding variables already present in the process environment.
func LoadDefaultIfPresent() error {
	return LoadFileIfPresent(defaultDotEnvPath)
}

// LoadFileIfPresent loads key/value pairs from path when it exists.
func LoadFileIfPresent(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		key, value, ok, err := parseEnvLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("parse %s line %d: %w", path, lineNumber, err)
		}
		if !ok {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s from %s: %w", key, path, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan %s: %w", path, err)
	}
	return nil
}

func parseEnvLine(line string) (key string, value string, ok bool, err error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false, nil
	}

	if strings.HasPrefix(trimmed, "export ") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "export "))
	}

	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) != 2 {
		return "", "", false, errors.New("expected KEY=VALUE")
	}

	key = strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", false, errors.New("missing key")
	}
	value = strings.TrimSpace(parts[1])

	if len(value) >= 2 {
		if value[0] == '"' && value[len(value)-1] == '"' {
			unquoted, unquoteErr := strconv.Unquote(value)
			if unquoteErr != nil {
				return "", "", false, fmt.Errorf("unquote value for %s: %w", key, unquoteErr)
			}
			value = unquoted
		} else if value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}
	}

	return key, value, true, nil
}
