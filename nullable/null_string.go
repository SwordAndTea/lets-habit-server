package nullable

import (
	"database/sql/driver"
	"encoding/json"
)

// NullString the nullable string type
type NullString struct {
	value   string
	nonnull bool
}

// NewNullString create a `nonnull` NullString
func NewNullString(value string) *NullString {
	return &NullString{value, true}
}

// MakeNullString create a `nonnull` NullString
func MakeNullString(value string) NullString {
	return *NewNullString(value)
}

// Null judge that nullable string object is null
func (s *NullString) Null() bool {
	return !s.nonnull
}

// NotNull judge that nullable string object is not null
func (s *NullString) NotNull() bool {
	return s.nonnull
}

// Get get NullString internal value
func (s *NullString) Get() string {
	return s.value
}

// Set modify NullString value, become nonnull
func (s *NullString) Set(value string) {
	s.value = value
	s.nonnull = true
}

// SetNull set self to null
func (s *NullString) SetNull() {
	s.value = ""
	s.nonnull = false
}

// UnmarshalString implementation for param.StringUnmarshaler
func (s *NullString) UnmarshalString(str string) error {
	strLen := len(str)
	if strLen > 1 && str[0] == '"' && str[strLen-1] == '"' {
		s.value = str[1 : strLen-1]
		s.nonnull = true
		return nil
	}
	s.value = str
	s.nonnull = true
	return nil
}

// Scan sql.Scanner interface implementation
func (s *NullString) Scan(value interface{}) error {
	if value == nil {
		s.value = ""
		s.nonnull = false
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return ScanErrorWrapper("NullString", value)
	}

	return s.UnmarshalString(string(bytes))
}

// Value driver.Valuer interface implementation
func (s NullString) Value() (driver.Value, error) {
	if s.nonnull {
		return s.value, nil
	}
	return nil, nil
}

// MarshalJSON json.Marshaler interface implementation
func (s *NullString) MarshalJSON() ([]byte, error) {
	if s.nonnull {
		return json.Marshal(s.value)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON json.Unmarshaler interface implementation
func (s *NullString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		s.value = ""
		s.nonnull = false
		return nil
	}
	s.nonnull = true
	return json.Unmarshal(data, &s.value)
}
