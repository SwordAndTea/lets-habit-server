package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullUint the nullable uint type
type NullUint struct {
	value   uint
	nonnull bool
}

// NewNullUint create a `nonnull` NullUint pointer
func NewNullUint(value uint) *NullUint {
	return &NullUint{
		value:   value,
		nonnull: true,
	}
}

// MakeNullUint create a `nonnull` NullUint
func MakeNullUint(value uint) NullUint {
	return *NewNullUint(value)
}

// NullUintFromString create a `nonnull` NullUint from string
func NullUintFromString(s string) (*NullUint, error) {
	i := &NullUint{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable uint object is null
func (i *NullUint) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable uint object is not null
func (i *NullUint) NotNull() bool {
	return i.nonnull
}

// Get get NullUint internal value
func (i *NullUint) Get() uint {
	return i.value
}

// Set modify NullUint value, become nonnull
func (i *NullUint) Set(value uint) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullUint) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullUint) UnmarshalString(str string) error {
	n, err := strconv.ParseUint(str, 10, strconv.IntSize)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = uint(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullUint) Scan(value interface{}) error {
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
func (i NullUint) Value() (driver.Value, error) {
	if i.nonnull {
		return uint64(i.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullUint) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullUint) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
