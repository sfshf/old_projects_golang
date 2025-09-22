package tools

import (
	"strconv"
	"time"
)

// StringToInt convert from string to int64
func StringToInt(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

// GetTimestamp millisecond
func GetTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}
