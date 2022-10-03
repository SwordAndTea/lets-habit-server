package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullInt32 the nullable int32 type
type NullInt32 struct {
	value   int32
	nonnull bool
}

// NewNullInt32 create a `nonnull` NullInt32 pointer
func NewNullInt32(value int32) *NullInt32 {
	return &NullInt32{
		value:   value,
		nonnull: true,
	}
}

// MakeNullInt32 create a `nonnull` NullInt32
func MakeNullInt32(value int32) NullInt32 {
	return *NewNullInt32(value)
}

// NullInt32FromString create a `nonnull` NullInt32 from string
func NullInt32FromString(s string) (*NullInt32, error) {
	i := &NullInt32{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable int32 object is null
func (i *NullInt32) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable int32 object is not null
func (i *NullInt32) NotNull() bool {
	return i.nonnull
}

// Get get NullInt32 internal value
func (i *NullInt32) Get() int32 {
	return i.value
}

// Set modify NullInt32 value, become nonnull
func (i *NullInt32) Set(value int32) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullInt32) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullInt32) UnmarshalString(str string) error {
	n, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = int32(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullInt32) Scan(value interface{}) error {
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
func (i NullInt32) Value() (driver.Value, error) {
	if i.nonnull {
		return int64(i.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullInt32) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullInt32) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
