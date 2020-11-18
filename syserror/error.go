package syserror

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Option func(e *SysError)

func WithCode(code codes.Code) Option {
	return func(e *SysError) {
		e.Code = code
	}
}

func WithCustomCode(code int) Option {
	return func(e *SysError) {
		e.Code = codes.Code(code)
	}
}

func WithFields(fields map[string]interface{}) Option {
	return func(e *SysError) {
		e.MemoryValues = fields
	}
}

type SysError struct {
	TraceID      string
	ID           string
	Note         string
	Code         codes.Code
	ErrorAt      string
	Wrapper      []string
	MemoryValues map[string]interface{}
}

func (e *SysError) Error() string {
	return e.String()
}

func (e *SysError) String() string {
	result := "ErrorID: " + e.ID + " TraceID: " + e.TraceID + "\n"

	result += "Error: \n"
	if len(e.Wrapper) != 0 {
		for i := len(e.Wrapper) - 1; i >= 0; i-- {
			result += "    Cause of: " + e.Wrapper[i] + "\n"
		}
	}
	result += "    Cause of: " + e.Note + "\n"

	if e.ErrorAt != "" {
		result += "ErrorAt:\n    " + e.ErrorAt + "\n"
	}
	// 此种写法避免json marshal 需要异常处理。
	if len(e.MemoryValues) != 0 {
		result += "MemoryValues:\n"
		for k, v := range e.MemoryValues {
			result += "    " + k + " ~> " + fmt.Sprintf("%v", v) + "\n"
		}
	}
	return result
}

func (e *SysError) StatusError() error {
	msg := e.SimpleError().Error()
	return status.Error(e.Code, fmt.Sprintf("%s. trace_id:`%s`", msg, e.TraceID))
}

func (e *SysError) SimpleError() error {
	msg := e.Note
	if len(e.Wrapper) > 0 {
		msg = e.Wrapper[0] + ", " + msg
	}
	return errors.New(msg)
}

func New(traceId, id, note string, memoryValues map[string]interface{}) error {
	return NewV2(traceId, id, note, WithFields(memoryValues))
}

func NewV2(traceId, id, note string, options ...Option) error {
	mergedErrorAt := strings.Builder{}
	maxStackDepth := 3
	for i := 0; i < maxStackDepth; i++ {
		_, file, line, _ := runtime.Caller(i + 1)
		mergedErrorAt.WriteString(fmt.Sprintf("%s:%d", file, line))
		if i+1 != maxStackDepth {
			mergedErrorAt.WriteString("\n    ")
		}
	}
	e := &SysError{
		TraceID: traceId,
		ID:      id,
		Code:    codes.Unknown,
		Note:    note,
		ErrorAt: mergedErrorAt.String(),
	}
	for _, opt := range options {
		if opt != nil {
			opt(e)
		}
	}
	return e
}

func Wrap(err error, msg string) error {
	if val, ok := err.(*SysError); ok {
		val.Wrapper = append(val.Wrapper, msg)
		return val
	}
	return errors.Wrap(err, msg)
}

func StatusError(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*SysError); ok {
		return e.StatusError()
	}
	return err
}
