package util

import "time"

type BasicValue interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | bool | float32 | float64 | string | time.Time
}

func LiteralValuePtr[T BasicValue](v T) *T {
	return &v
}
