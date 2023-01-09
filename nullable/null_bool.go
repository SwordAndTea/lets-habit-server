package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullBool the nullable bool type
type NullBool struct {
	value   bool
	nonnull bool
}

// NewNullBool create a `nonnull` NullBool pointer
func NewNullBool(value bool) *NullBool {
	return &NullBool{
		value:   value,
		nonnull: true,
	}
}

// MakeNullBool create a `nonnull` NullBool
func MakeNullBool(value bool) NullBool {
	return *NewNullBool(value)
}

// NullBoolFromString create a `nonnull` NullBool from string
func NullBoolFromString(s string) (*NullBool, error) {
	i := &NullBool{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable bool object is null
func (b *NullBool) Null() bool {
	return !b.nonnull
}

// NotNull judge that nullable int object is not null
func (b *NullBool) NotNull() bool {
	return b.nonnull
}

// Get get NullBool internal value
func (b *NullBool) Get() bool {
	return b.value
}

// Set modify NullBool value, become nonnull
func (b *NullBool) Set(value bool) {
	b.value = value
	b.nonnull = true
}

// SetNull set self to null
func (b *NullBool) SetNull() {
	b.value = false
	b.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (b *NullBool) UnmarshalString(str string) error {
	n, err := strconv.ParseBool(str)
	if err != nil {
		return err
	}
	b.value = n
	b.nonnull = true
	return nil
}

// Scan sql.Scanner interface implementation
func (b *NullBool) Scan(value interface{}) error {
	if value == nil {
		b.value = false
		b.nonnull = false
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		intValue, ok := value.(int64)
		if !ok {
			return ScanErrorWrapper("NullBool", value)
		}
		b.value = intValue == 1
		b.nonnull = true
		return nil
	}
	return b.UnmarshalString(string(bytes))
}

// Value driver.Valuer implementation
func (b NullBool) Value() (driver.Value, error) {
	if b.nonnull {
		return b.value, nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (b *NullBool) MarshalJSON() ([]byte, error) {
	if b.nonnull {
		return json.Marshal(b.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (b *NullBool) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		b.value = false
		b.nonnull = false
		return nil
	}

	b.nonnull = true
	return json.Unmarshal(data, &b.value)
}
