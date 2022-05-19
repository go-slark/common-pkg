package xerror

import (
	"fmt"
	"github.com/pkg/errors"
)

type customError struct {
	Code int
	Msg  string
	error
}

func (e customError) Error() string {
	return fmt.Sprintf("code:%d, reason:%s, err:%v", e.Code, e.Msg, e.error)
}

func NewError(code int, msg string, err error) error {
	return customError{
		Code:  code,
		Msg:   msg,
		error: err,
	}
}

func GetErr(err error) *customError {
	e := &customError{}
	errors.As(err, e)
	return e
}
