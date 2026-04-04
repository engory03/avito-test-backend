package config

import "testing"

func TestGetEnv_WithValue(t *testing.T) {
	t.Setenv("TEST_CONFIG_KEY_XYZ", "from-env")
	if got := GetEnv("TEST_CONFIG_KEY_XYZ", "fallback"); got != "from-env" {
		t.Fatal(got)
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	if got := GetEnv("VERY_UNLIKELY_ENV_VAR_QQQ999", "fb"); got != "fb" {
		t.Fatal(got)
	}
}
