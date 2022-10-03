package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullInt8 the nullable int8 type
type NullInt8 struct {
	value   int8
	nonnull bool
}

// NewNullInt8 create a `nonnull` NullInt8 pointer
func NewNullInt8(value int8) *NullInt8 {
	return &NullInt8{
		value:   value,
		nonnull: true,
	}
}

// MakeNullInt8 create a `nonnull` NullInt8
func MakeNullInt8(value int8) NullInt8 {
	return *NewNullInt8(value)
}

// NullInt8FromString create a `nonnull` NullInt8 from string
func NullInt8FromString(s string) (*NullInt8, error) {
	i := &NullInt8{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable int8 object is null
func (i *NullInt8) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable int8 object is not null
func (i *NullInt8) NotNull() bool {
	return i.nonnull
}

// Get get NullInt8 internal value
func (i *NullInt8) Get() int8 {
	return i.value
}

// Set modify NullInt8 value, become nonnull
func (i *NullInt8) Set(value int8) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullInt8) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullInt8) UnmarshalString(str string) error {
	n, err := strconv.ParseInt(str, 10, 8)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = int8(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullInt8) Scan(value interface{}) error {
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
func (i NullInt8) Value() (driver.Value, error) {
	if i.nonnull {
		return int64(i.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullInt8) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullInt8) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
