package env

import (
	"bufio"
	"bytes"
	_ "embed"
	"os"
	"strconv"
	"strings"
)

//go:embed .env
var envFile []byte

var envMap map[string]string

func init() {
	envMap = make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewReader(envFile))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		envMap[key] = val
		_ = os.Setenv(key, val)
	}
}

func GetString(key, fallback string) string {
	if val, ok := envMap[key]; ok {
		return val
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	if val, ok := envMap[key]; ok {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}

func GetBool(key string, fallback bool) bool {
	if val, ok := envMap[key]; ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}
