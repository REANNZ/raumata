package internal

import (
	"strconv"
	"strings"
)

func FormatFloat(f float64, prec int, bitSize int) string {
	s := strconv.FormatFloat(f, 'f', prec, bitSize)

	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		if s[len(s)-1] == '.' {
			s = s[:len(s)-1]
		}
	}

	return s
}

func FormatFloat32(f float32, prec int) string {
	return FormatFloat(float64(f), prec, 32)
}
