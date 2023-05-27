//go:build go1.20

package taskmanager

import (
	"errors"
)

func joinErrors(errs ...error) error {
	return errors.Join(errs...)
}
