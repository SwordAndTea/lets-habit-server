package dal

import (
	"database/sql/driver"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// UID user id
type UID string

// Password a tool model to represent a hashed encrypted passport
type Password struct {
	Data   string
	Hashed bool
}

func NewRawPassword(data string) *Password {
	return &Password{
		Data:   data,
		Hashed: false,
	}
}

// Scan an implement of sql.Scanner
func (p *Password) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type")
	}
	if p == nil {
		*p = Password{}
	}
	p.Data = string(bytes)
	p.Hashed = true // in db, always store the hashed password
	return nil
}

// Value an implement of driver.Valuer
func (p *Password) Value() (driver.Value, error) {
	if p.Hashed {
		return p.Data, nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(p.Data), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return string(hashed), nil
}

// HashedValue return hashed value
func (p *Password) HashedValue() string {
	if p.Hashed {
		return p.Data
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(p.Data), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashed)
}

type Pagination struct {
	Page     uint
	PageSize uint
}
