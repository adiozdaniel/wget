package utils

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type RateLimitedReader struct {
	reader     io.Reader
	rateLimit  int64 // bytes per second
	bucket     int64
	lastFilled time.Time
}

func RateLimitValidator(s string) error {
	idx := strings.Index(s, "=")
	if idx == -1 || idx == len(s)-1 {
		return fmt.Errorf("invalid rate limit format.\nUsage: --rate-limit=400k || --rate-limit=2M")
	}

	// Extract value and unit
	val, unit := s[idx+1:len(s)-1], s[len(s)-1:]

	// Normalize unit to lowercase
	unit = strings.ToLower(unit)

	// Ensure the unit is valid
	if unit != "k" && unit != "m" {
		return fmt.Errorf("invalid rate limit unit.\nUsage: --rate-limit=400k || --rate-limit=2M")
	}

	// Convert the numeric part
	if _, err := strconv.Atoi(val); err != nil {
		return fmt.Errorf("invalid numeric value in rate limit: %s", val)
	}

	return nil
}

func parseRateLimit(rateLimit string) (int64, error) {
	if len(rateLimit) < 2 {
		return 0, fmt.Errorf("invalid rate limit")
	}

	multiplier := 1
	switch rateLimit[len(rateLimit)-1] {
	case 'k', 'K':
		multiplier = 1024
		rateLimit = rateLimit[:len(rateLimit)-1]
	case 'M':
		multiplier = 1024 * 1024
		rateLimit = rateLimit[:len(rateLimit)-1]
	}

	rate, err := strconv.Atoi(rateLimit)
	if err != nil {
		return 0, err
	}
	return int64(rate * multiplier), nil
}
func NewRateLimitedReader(reader io.Reader, limit string) *RateLimitedReader {
	// Convert limit to bytes per second (rateLimit)
	rateLimit, _ := parseRateLimit(limit)
	return &RateLimitedReader{reader: reader, rateLimit: rateLimit, lastFilled: time.Now()}
}

func (r *RateLimitedReader) Read(p []byte) (n int, err error) {
	if r.bucket <= 0 {
		time.Sleep(time.Second)
		r.bucket = r.rateLimit
		r.lastFilled = time.Now()
	}

	toRead := int64(len(p))
	if toRead > r.bucket {
		toRead = r.bucket
	}

	n, err = r.reader.Read(p[:toRead])
	r.bucket -= int64(n)

	return n, err
}
