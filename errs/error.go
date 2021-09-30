package errs

import (
	"errors"
	"fmt"
)

type Errs struct {
	Code int32
	Msg  error
}

func (e Errs) Error() string {
	return fmt.Sprintf(`{error_code: %d, error_msg: %s}`, e.Code, e.Msg)
}

func (e Errs) Unwrap() error {
	return e.Msg
}

func (e Errs) Is(target error) bool {
	t, ok := target.(Errs)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func New(code int32, msg string) error {
	return Errs{Code: code, Msg: errors.New(msg)}
}

func Wrap(target error, errs ...error) error {
	if len(errs) < 1 {
		return target
	}

	var e Errs
	if errors.As(target, &e) {
		ee := fmt.Errorf("%s, wrap: %w", e.Msg, errs[0])
		for i := range errs {
			if i == 0 {
				continue
			}
			ee = fmt.Errorf("%v, wrap: %w", ee, errs[i])
		}
		return Errs{Code: e.Code, Msg: fmt.Errorf("%w", ee)}
	}

	ee := errs[0]
	for i := range errs {
		if i == 0 {
			continue
		}
		ee = fmt.Errorf("%v, wrap: %w", ee, errs[i])
	}
	return fmt.Errorf("%v, wrap: %w", target, ee)
}

func Parse(err error, defaultCode int32) Errs {
	var e Errs
	if errors.As(err, &e) {
		return Errs{Code: e.Code, Msg: e.Msg}
	}
	return Errs{Code: defaultCode, Msg: err}
}

func MustParse(err error) Errs {
	var e Errs
	if errors.As(err, &e) {
		return Errs{Code: e.Code, Msg: e.Msg}
	}
	panic("err is not instance of Errs")
}
