package controller

import "errors"

func NewUnhandledError(err error) error {
	return unhandledError{e: err}
}

type unhandledError struct {
	e error
}

func IsUnhandledError(err error) bool {
	return errors.Is(err, unhandledError{})
}

func (unhandledError) Is(target error) bool {
	//nolint: errorlint // This check is actually fine.
	_, ok := target.(unhandledError)
	return ok
}

func (err unhandledError) Error() string {
	if err.e == nil {
		return ""
	}

	return err.e.Error()
}

func (err unhandledError) Unwrap() error {
	return err.e
}
