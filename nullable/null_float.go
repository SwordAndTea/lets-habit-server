package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullFloat32 the nullable float32 type
type NullFloat32 struct {
	value   float32
	nonnull bool
}

// NewNullFloat32 create a `nonnull` NullFloat32 pointer
func NewNullFloat32(value float32) *NullFloat32 {
	return &NullFloat32{
		value:   value,
		nonnull: true,
	}
}

// MakeNullFloat32 create a `nonnull` NullFloat32
func MakeNullFloat32(value float32) NullFloat32 {
	return *NewNullFloat32(value)
}

// NullFloat32FromString create a `nonnull` NullFloat32 from string
func NullFloat32FromString(s string) (*NullFloat32, error) {
	f := &NullFloat32{}
	err := f.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Null judge that nullable float32 object is null
func (f *NullFloat32) Null() bool {
	return !f.nonnull
}

// NotNull judge that nullable float32 object is not null
func (f *NullFloat32) NotNull() bool {
	return f.nonnull
}

// Get get NullFloat32 internal value
func (f *NullFloat32) Get() float32 {
	return f.value
}

// Set modify NullFloat32 value and null information
func (f *NullFloat32) Set(value float32) {
	f.value = value
	f.nonnull = true
}

// SetNull set self to null
func (f *NullFloat32) SetNull() {
	f.value = 0
	f.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (f *NullFloat32) UnmarshalString(str string) error {
	n, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return err
	}
	f.nonnull = true
	f.value = float32(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (f *NullFloat32) Scan(value interface{}) error {
	if value == nil {
		f.value = 0
		f.nonnull = false
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return ScanErrorWrapper("NullFloat32", value)
	}
	return f.UnmarshalString(string(bytes))
}

// Value driver.Value interface implementation
func (f NullFloat32) Value() (value driver.Value, err error) {
	if f.nonnull {
		return float64(f.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (f *NullFloat32) MarshalJSON() ([]byte, error) {
	if f.nonnull {
		return json.Marshal(f.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (f *NullFloat32) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		f.value = 0
		f.nonnull = false
		return nil
	}
	f.nonnull = true
	return json.Unmarshal(data, &f.value)
}

// NullFloat64 the nullable float64 type
type NullFloat64 struct {
	value   float64
	nonnull bool
}

// NewNullFloat64 create a `nonnull` NullFloat64 pointer
func NewNullFloat64(value float64) *NullFloat64 {
	return &NullFloat64{
		value:   value,
		nonnull: true,
	}
}

// MakeNullFloat64 create a `nonnull` NullFloat64
func MakeNullFloat64(value float64) NullFloat64 {
	return *NewNullFloat64(value)
}

// NullFloat64FromString create a `nonnull` NullFloat64
func NullFloat64FromString(s string) (*NullFloat64, error) {
	f := &NullFloat64{}
	err := f.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Null judge that nullable float64 object is null
func (f *NullFloat64) Null() bool {
	return !f.nonnull
}

// NotNull judge that nullable float64 object is not null
func (f *NullFloat64) NotNull() bool {
	return f.nonnull
}

// Get get NullFloat64 internal value
func (f *NullFloat64) Get() float64 {
	return f.value
}

// Set modify NullFloat64 value and null information
func (f *NullFloat64) Set(value float64) {
	f.value = value
	f.nonnull = true
}

// SetNull set self to null
func (f *NullFloat64) SetNull() {
	f.value = 0
	f.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (f *NullFloat64) UnmarshalString(str string) error {
	n, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	f.nonnull = true
	f.value = n
	return nil
}

// Scan sql.Scanner interface implementation
func (f *NullFloat64) Scan(value interface{}) error {
	if value == nil {
		f.value = 0
		f.nonnull = false
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return ScanErrorWrapper("NullFloat64", value)
	}

	return f.UnmarshalString(string(bytes))
}

// Value driver.Valuer interface implementation
func (f NullFloat64) Value() (value driver.Value, err error) {
	if f.nonnull {
		return f.value, nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (f *NullFloat64) MarshalJSON() ([]byte, error) {
	if f.nonnull {
		return json.Marshal(f.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (f *NullFloat64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		f.value = 0
		f.nonnull = false
		return nil
	}
	f.nonnull = true
	return json.Unmarshal(data, &f.value)
}
