package dal

import (
	"database/sql/driver"
	"golang.org/x/crypto/bcrypt"
)

// UID user id
type UID string

type Password string

func (e Password) Value() (driver.Value, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(e), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return string(hashed), nil
}
