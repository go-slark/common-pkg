package xerror

import (
	"fmt"
	"github.com/pkg/errors"
)

type customError struct {
	Code     int
	Msg      string
	Metadata map[string]interface{}
	error
}

func (e customError) Error() string {
	return fmt.Sprintf("code:%d, reason:%s, metadata:%v, err:%v", e.Code, e.Msg, e.Metadata, e.error)
}

func NewError(code int, msg string, metadata map[string]interface{}, err error) error {
	return customError{
		Code:     code,
		Msg:      msg,
		Metadata: metadata,
		error:    err,
	}
}

func GetErr(err error) *customError {
	e := &customError{}
	errors.As(err, e)
	return e
}
