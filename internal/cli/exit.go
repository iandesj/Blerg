package cli

import "errors"

const (
	ExitSuccess = 0
	ExitError   = 1
	ExitNoMatch = 2
)

type codedError struct {
	code int
	err  error
}

func (e codedError) Error() string {
	return e.err.Error()
}

func (e codedError) Unwrap() error {
	return e.err
}

func exitCodeFor(err error) int {
	var ce codedError
	if errors.As(err, &ce) {
		return ce.code
	}
	return ExitError
}

func withExitCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return codedError{code: code, err: err}
}
