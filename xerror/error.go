package xerror

import (
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	UnknownReason = ""
	UnknownCode   = 1000
)

type customError struct {
	Status
	error
}

func (e customError) Error() string {
	return fmt.Sprintf("code:%d, reason:%s, msg:%v, metadata:%v, err:%v", e.Code, e.Reason, e.Message, e.Metadata, e.error)
}

func NewError(code int, reason, msg string, metadata map[string]string, err error) error {
	return customError{
		Status: Status{
			Code:     int32(code),
			Reason:   reason,
			Message:  msg,
			Metadata: metadata,
		},
		error: err,
	}
}

func GetErr(err error) *customError {
	e := &customError{}
	errors.As(err, e)
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
	err := Clone(e)
	err.error = cause
	return err
}

func (e *customError) WithMetadata(md map[string]string) *customError {
	err := Clone(e)
	err.Metadata = md
	return err
}

func (e *customError) GrpcStatus() *status.Status {
	s, _ := status.New(codes.Code(e.Code), e.Message).
		WithDetails(&errdetails.ErrorInfo{
			Reason:   e.Reason,
			Metadata: e.Metadata,
		})
	return s
}

func New(code int, reason, message string) *customError {
	return &customError{
		Status: Status{
			Code:    int32(code),
			Message: message,
			Reason:  reason,
		},
	}
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

func Clone(err *customError) *customError {
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
		ret := New(
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
	return New(UnknownCode, UnknownReason, err.Error())
}
