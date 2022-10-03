package nullable

import (
	"github.com/pkg/errors"
)

func ScanErrorWrapper(typeName string, value interface{}) error {
	return errors.Errorf("Failed to Scan %s value: %v", typeName, value)
}
