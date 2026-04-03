package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDuration supports "h", "m", "s" and also "d" (days).
func ParseDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Handle "Xd" (days)
	if strings.HasSuffix(value, "d") {
		numStr := strings.TrimSuffix(value, "d")
		days, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid day format: %v", err)
		}
		return time.Hour * 24 * time.Duration(days), nil
	}

	// Fallback to standard Go durations like "15m", "1h", etc.
	return time.ParseDuration(value)
}
