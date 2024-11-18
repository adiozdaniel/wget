package utils

import (
	"io"
	"time"
)

type ProgressRecoder struct {
	Reader           io.Reader
	Total            int64
	Progress         int64
	ProgressFunction func(int64, int64)
}

type WgetValues struct {
	BackgroudMode   bool   // Flag -B
	OutputFile      string // Flag -O
	OutPutDirectory string // Flag -P
	RateLimitValue  int    //Flag --rate-limit
	Reject          bool
	Exclude         string // Flag exclude || -X
	ConvertLinks    bool   // Flag --convert-links
	Mirror          bool   //Flag --mirror
	Url             string // --- url given
}

// Rate limiter Struct
// RateLimitedReader limits the read speed to a specified rate in bytes per second
type RateLimitedReader struct {
	Reader io.Reader
	Rate   int64 // bytes per second
	Ticker *time.Ticker
}

func WgetInstance() *WgetValues {
	return &WgetValues{
		BackgroudMode:   false,
		OutputFile:      "",
		OutPutDirectory: "",
		RateLimitValue:  0,
		Reject:          false,
		Exclude:         "",
		ConvertLinks:    false,
		Mirror:          false,
	}
}
