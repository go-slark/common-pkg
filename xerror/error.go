package xerror

import (
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	UnknownReason = "UNKNOWN_REASON"
	UnknownCode   = 600

	ParamValid     = "PARAM_VALID"
	ParamValidCode = 601

	FormatInvalid     = "FORMAT_INVALID"
	FormatInvalidCode = 602

	Panic     = "PANIC"
	PanicCode = 603
)

type customError struct {
	Status
	Surplus interface{}
	clone   bool
	error
}

func (e customError) Error() string {
	return fmt.Sprintf("code:%d, reason:%s, msg:%v, metadata:%v, surplus:%v, err:%v", e.Code, e.Reason, e.Message, e.Metadata, e.Surplus, e.error)
}

func NewError(code int, reason, msg string) *customError {
	return &customError{
		Status: Status{
			Code:    int32(code),
			Reason:  reason,
			Message: msg,
		},
	}
}

func GetErr(err error) *customError {
	e := &customError{
		Status: Status{
			Code: UnknownCode,
		},
		error: err,
	}
	errors.As(err, &e)
	return e
}

// grpc error

func (e *customError) Unwrap() error {
	return e.error
}

func (e *customError) Is(err error) bool {
	if se := new(customError); errors.As(err, &se) {
		return se.Code == e.Code && se.Reason == e.Reason
	}
	return false
}

func (e *customError) WithError(cause error) *customError {
	err := clone(e)
	err.error = fmt.Errorf("%+v", cause)
	return err
}

func (e *customError) WithMetadata(md map[string]string) *customError {
	err := clone(e)
	err.Metadata = md
	return err
}

func (e *customError) WithSurplus(surplus interface{}) *customError {
	err := clone(e)
	err.Surplus = surplus
	return err
}

func (e *customError) GRPCStatus() *status.Status {
	s, _ := status.New(codes.Code(e.Code), e.Message).
		WithDetails(&errdetails.ErrorInfo{
			//Reason:   e.Reason,
			Reason:   fmt.Sprintf("%+v", e.error),
			Metadata: e.Metadata,
		})
	return s
}

func Code(err error) int {
	if err == nil {
		return 200
	}
	return int(FromError(err).Code)
}

func Reason(err error) string {
	if err == nil {
		return UnknownReason
	}
	return FromError(err).Reason
}

func clone(err *customError) *customError {
	if err.clone {
		return err
	}
	err.clone = true
	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &customError{
		error: err.error,
		Status: Status{
			Code:     err.Code,
			Reason:   err.Reason,
			Message:  err.Message,
			Metadata: metadata,
		},
		Surplus: err.Surplus,
	}
}

func FromError(err error) *customError {
	if err == nil {
		return nil
	}
	if se := new(customError); errors.As(err, &se) {
		return se
	}
	gs, ok := status.FromError(err)
	if ok {
		ret := NewError(
			int(gs.Code()),
			UnknownReason,
			gs.Message(),
		)
		for _, detail := range gs.Details() {
			switch d := detail.(type) {
			case *errdetails.ErrorInfo:
				ret.Reason = d.Reason
				return ret.WithMetadata(d.Metadata)
			}
		}
		return ret
	}
	return NewError(UnknownCode, UnknownReason, err.Error())
}
