package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
)

// NullUint32 the nullable uint32 type
type NullUint32 struct {
	value   uint32
	nonnull bool
}

// NewNullUint32 the nullable uint32 type
func NewNullUint32(value uint32) *NullUint32 {
	return &NullUint32{
		value:   value,
		nonnull: true,
	}
}

// MakeNullUint32 create a `nonnull` NullUint32
func MakeNullUint32(value uint32) NullUint32 {
	return *NewNullUint32(value)
}

func NullUint32FromString(s string) (*NullUint32, error) {
	i := &NullUint32{}
	err := i.UnmarshalString(s)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Null judge that nullable uint32 object is null
func (i *NullUint32) Null() bool {
	return !i.nonnull
}

// NotNull judge that nullable uint32 object is not null
func (i *NullUint32) NotNull() bool {
	return i.nonnull
}

// Get get NullUint32 internal value
func (i *NullUint32) Get() uint32 {
	return i.value
}

// Set modify NullUint32 value, become nonnull
func (i *NullUint32) Set(value uint32) {
	i.value = value
	i.nonnull = true
}

// SetNull set self to null
func (i *NullUint32) SetNull() {
	i.value = 0
	i.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (i *NullUint32) UnmarshalString(str string) error {
	n, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return err
	}
	i.nonnull = true
	i.value = uint32(n)
	return nil
}

// Scan sql.Scanner interface implementation
func (i *NullUint32) Scan(value interface{}) error {
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
func (i NullUint32) Value() (driver.Value, error) {
	if i.nonnull {
		return uint64(i.value), nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (i *NullUint32) MarshalJSON() ([]byte, error) {
	if i.nonnull {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (i *NullUint32) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		i.value = 0
		i.nonnull = false
		return nil
	}
	i.nonnull = true
	return json.Unmarshal(data, &i.value)
}
