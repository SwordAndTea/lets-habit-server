package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullUint8 the nullable uint8 type
type NullUint8 struct {
	value   uint8
	nonnull bool
}

// NewNullUint8 create a `nonnull` NullUint8
func NewNullUint8(value uint8) *NullUint8 {
	return &NullUint8{
		value:   value,
		nonnull: true,
	}
}

// MakeNullUint8 create a `nonnull` NullUint8
func MakeNullUint8(value uint8) NullUint8 {
	return *NewNullUint8(value)
}

// NullUint8FromString create a `nonnull` NullUint8 from string
func NullUint8FromString(s string) (*NullUint8, error) {
	i := &NullUint8{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable uint8 object is null
func (i *NullUint8) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable uint8 object is not null
func (i *NullUint8) NotNull() bool {
	return i.nonnull
}

// Get get NullUint8 internal value
func (i *NullUint8) Get() uint8 {
	return i.value
}

// Set modify NullUint8 value, become nonnull
func (i *NullUint8) Set(value uint8) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullUint8) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullUint8) UnmarshalString(str string) error {
	n, err := strconv.ParseUint(str, 10, 8)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = uint8(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullUint8) Scan(value interface{}) error {
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

// Value driver.Valuer interface implementation
func (i NullUint8) Value() (driver.Value, error) {
	if i.nonnull {
		return uint64(i.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullUint8) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullUint8) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
