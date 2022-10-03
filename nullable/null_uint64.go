package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullUint64 the nullable uint64 type
type NullUint64 struct {
	value   uint64
	nonnull bool
}

// NewNullUint64 create a `nonnull` NullUint64
func NewNullUint64(value uint64) *NullUint64 {
	return &NullUint64{
		value:   value,
		nonnull: true,
	}
}

// MakeNullUint64 create a `nonnull` NullUint64
func MakeNullUint64(value uint64) NullUint64 {
	return *NewNullUint64(value)
}

// NullUint64FromString create a `nonnull` NullUint64 from string
func NullUint64FromString(s string) (*NullUint64, error) {
	i := &NullUint64{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable uint64 object is null
func (i *NullUint64) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable uint64 object is not null
func (i *NullUint64) NotNull() bool {
	return i.nonnull
}

// Get get NullUint64 internal value
func (i *NullUint64) Get() uint64 {
	return i.value
}

// Set modify NullUint64 value, become nonnull
func (i *NullUint64) Set(value uint64) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullUint64) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullUint64) UnmarshalString(str string) error {
	n, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = n
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullUint64) Scan(value interface{}) error {
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
func (i NullUint64) Value() (driver.Value, error) {
	if i.nonnull {
		return i.value, nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullUint64) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullUint64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
