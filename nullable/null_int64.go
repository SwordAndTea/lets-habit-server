package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullInt64 the nullable int64 type
type NullInt64 struct {
	value   int64
	nonnull bool
}

// NewNullInt64 create a `nonnull` NullInt64 pointer
func NewNullInt64(value int64) *NullInt64 {
	return &NullInt64{
		value:   value,
		nonnull: true,
	}
}

// MakeNullInt64 create a `nonnull` NullInt64
func MakeNullInt64(value int64) NullInt64 {
	return *NewNullInt64(value)
}

// NullInt64FromString create a `nonnull` NullInt64 from string
func NullInt64FromString(s string) (*NullInt64, error) {
	i := &NullInt64{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable int64 object is null
func (i *NullInt64) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable int64 object is not null
func (i *NullInt64) NotNull() bool {
	return i.nonnull
}

// Get get NullInt64 internal value
func (i *NullInt64) Get() int64 {
	return i.value
}

// Set modify NullInt64 value, become nonnull
func (i *NullInt64) Set(value int64) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullInt64) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullInt64) UnmarshalString(str string) error {
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = n
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullInt64) Scan(value interface{}) error {
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
func (i NullInt64) Value() (driver.Value, error) {
	if i.nonnull {
		return i.value, nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullInt64) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullInt64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
