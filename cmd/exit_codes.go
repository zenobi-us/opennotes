package cmd

import "errors"

const (
	ExitCodeGeneral  = 1
	ExitCodeNotFound = 2
	ExitCodeConflict = 3
)

type codedError struct {
	code int
	err  error
}

func (e *codedError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *codedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func withExitCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return &codedError{code: code, err: err}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var coded *codedError
	if errors.As(err, &coded) && coded != nil && coded.code > 0 {
		return coded.code
	}

	return ExitCodeGeneral
}
