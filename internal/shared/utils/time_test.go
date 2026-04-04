package utils

import (
	"testing"
	"time"
)

func TestParseDuration_Days(t *testing.T) {
	d, err := ParseDuration("3d")
	if err != nil {
		t.Fatal(err)
	}
	if d != 72*time.Hour {
		t.Fatalf("got %v", d)
	}
}

func TestParseDuration_Standard(t *testing.T) {
	d, err := ParseDuration("15m")
	if err != nil || d != 15*time.Minute {
		t.Fatalf("%v %v", d, err)
	}
}

func TestParseDuration_Empty(t *testing.T) {
	_, err := ParseDuration("")
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = ParseDuration("   ")
	if err == nil {
		t.Fatal("expected error")
	}
}
