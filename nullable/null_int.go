package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullInt the nullable int type
type NullInt struct {
	value   int
	nonnull bool
}

// NewNullInt create a `nonnull` NullInt pointer
func NewNullInt(value int) *NullInt {
	return &NullInt{
		value:   value,
		nonnull: true,
	}
}

// MakeNullInt create a `nonnull` NullInt
func MakeNullInt(value int) NullInt {
	return *NewNullInt(value)
}

// NullIntFromString create a `nonnull` NullInt from string
func NullIntFromString(s string) (*NullInt, error) {
	i := &NullInt{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable int object is null
func (i *NullInt) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable int object is not null
func (i *NullInt) NotNull() bool {
	return i.nonnull
}

// Get get NullInt internal value
func (i *NullInt) Get() int {
	return i.value
}

// Set modify NullInt value, become nonnull
func (i *NullInt) Set(value int) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullInt) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullInt) UnmarshalString(str string) error {
	n, err := strconv.ParseInt(str, 10, strconv.IntSize)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = int(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullInt) Scan(value interface{}) error {
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
func (i NullInt) Value() (driver.Value, error) {
	if i.nonnull {
		return int64(i.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullInt) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullInt) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}

	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
