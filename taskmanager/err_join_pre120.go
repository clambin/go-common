//go:build !go1.20

package taskmanager

import (
	"errors"
	"strings"
)

func joinErrors(errs ...error) error {
	var errorStrings []string

	for _, err := range errs {
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	if len(errorStrings) == 0 {
		return nil
	}

	return errors.New(strings.Join(errorStrings, "\n"))
}
