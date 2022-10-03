package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullInt16 the nullable int16 type
type NullInt16 struct {
	value   int16
	nonnull bool
}

// NewNullInt16 create a `nonnull` NullInt16 pointer
func NewNullInt16(value int16) *NullInt16 {
	return &NullInt16{
		value:   value,
		nonnull: true,
	}
}

// MakeNullInt16 create a `nonnull` NullInt16
func MakeNullInt16(value int16) NullInt16 {
	return *NewNullInt16(value)
}

// NullInt16FromString create a `nonnull` NullInt16 from string
func NullInt16FromString(s string) (*NullInt16, error) {
	i := &NullInt16{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable int16 object is null
func (i *NullInt16) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable int16 object is not null
func (i *NullInt16) NotNull() bool {
	return i.nonnull
}

// Get get NullInt16 internal value
func (i *NullInt16) Get() int16 {
	return i.value
}

// Set modify NullInt16 value, become nonnull
func (i *NullInt16) Set(value int16) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullInt16) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullInt16) UnmarshalString(str string) error {
	n, err := strconv.ParseInt(str, 10, 16)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = int16(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullInt16) Scan(value interface{}) error {
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
func (i NullInt16) Value() (driver.Value, error) {
	if i.nonnull {
		return int64(i.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullInt16) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullInt16) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
