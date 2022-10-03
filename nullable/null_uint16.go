package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullUint16 the nullable uint16 type
type NullUint16 struct {
	value   uint16
	nonnull bool
}

// NewNullUint16 create a `nonnull` NullUint16
func NewNullUint16(value uint16) *NullUint16 {
	return &NullUint16{
		value:   value,
		nonnull: true,
	}
}

// MakeNullUint16 create a `nonnull` NullUint16
func MakeNullUint16(value uint16) NullUint16 {
	return *NewNullUint16(value)
}

// NullUint16FromString create a `nonnull` NullUint16 from string
func NullUint16FromString(s string) (*NullUint16, error) {
	i := &NullUint16{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable uint16 object is null
func (i *NullUint16) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable uint16 object is not null
func (i *NullUint16) NotNull() bool {
	return i.nonnull
}

// Get get NullUint16 internal value
func (i *NullUint16) Get() uint16 {
	return i.value
}

// Set modify NullUint16 value, become nonnull
func (i *NullUint16) Set(value uint16) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullUint16) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullUint16) UnmarshalString(str string) error {
	n, err := strconv.ParseUint(str, 10, 16)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = uint16(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullUint16) Scan(value interface{}) error {
	if value == nil {
		i.value = 0
		i.nonnull = false
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return ScanErrorWrapper("NullInt", value)
	}
	return i.UnmarshalString(string(bytes))
}

// Value driver.Valuer implementation
func (i NullUint16) Value() (driver.Value, error) {
	if i.nonnull {
		return uint64(i.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullUint16) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullUint16) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
