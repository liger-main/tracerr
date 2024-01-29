// Package tracerr makes error output more informative.
// It adds stack trace to error and can display error with source fragments.
//
// Check example of output here https://github.com/ztrue/tracerr
package tracerr

import (
	"fmt"
	"runtime"
	"strings"
)

// DefaultCap is a default cap for frames array.
// It can be changed to number of expected frames
// for purpose of performance optimisation.
var DefaultCap = 20

// Error is an error with stack trace.
type Error interface {
	Error() string
	StackTrace() []Frame
	Unwrap() error
}

type errorData struct {
	// err contains original error.
	err error
	// optional additional message
	message string
	// frames contains stack trace of an error.
	frames []Frame
}

// CustomError creates an error with provided frames.
func CustomError(err error, frames []Frame) Error {
	return &errorData{
		err:    err,
		frames: frames,
	}
}

// Errorf creates new error with stacktrace and formatted message.
// Formatting works the same way as in fmt.Errorf.
func Errorf(message string, args ...interface{}) Error {
	return trace(fmt.Errorf(message, args...), "", 2)
}

// New creates new error with stacktrace.
func New(message string) Error {
	return trace(fmt.Errorf(message), "", 2)
}

func Newf(format string, a ...interface{}) Error {
	return trace(fmt.Errorf(format, a...), "", 2)
}

// Wrap adds stacktrace to existing error.
func Wrap(err error, message string) Error {
	if err == nil {
		return nil
	}
	e, ok := err.(Error)
	if ok {
		return e
	}
	return trace(err, message, 2)
}

func Wrapf(err error, format string, a ...interface{}) Error {
	return Wrap(err, fmt.Sprintf(format, a...))
}

// Unwrap returns the original error.
func Unwrap(err error) error {
	if err == nil {
		return nil
	}
	e, ok := err.(Error)
	if !ok {
		return err
	}
	return e.Unwrap()
}

// Error returns error message.
func (e *errorData) Error() string {
	builder := strings.Builder{}
	if e.message != "" {
		builder.WriteString(e.message)
		builder.WriteString("\n")
	}
	builder.WriteString(e.err.Error())
	builder.WriteString("\n")
	isFirstFrame := true
	for _, frame := range e.StackTrace() {
		if !isFirstFrame {
			builder.WriteString("\n")
		}
		isFirstFrame = false
		builder.WriteString("\t")
		builder.WriteString(frame.String())
	}
	return builder.String()
}

// StackTrace returns stack trace of an error.
func (e *errorData) StackTrace() []Frame {
	return e.frames
}

// Unwrap returns the original error.
func (e *errorData) Unwrap() error {
	return e.err
}

// Frame is a single step in stack trace.
type Frame struct {
	// Func contains a function name.
	Func string
	// Line contains a line number.
	Line int
	// Path contains a file path.
	Path string
}

// StackTrace returns stack trace of an error.
// It will be empty if err is not of type Error.
func StackTrace(err error) []Frame {
	e, ok := err.(Error)
	if !ok {
		return nil
	}
	return e.StackTrace()
}

// String formats Frame to string.
func (f Frame) String() string {
	return fmt.Sprintf("%s:%d %s()", f.Path, f.Line, f.Func)
}

func trace(err error, message string, skip int) Error {
	frames := make([]Frame, 0, DefaultCap)
	for {
		pc, path, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		frame := Frame{
			Func: fn.Name(),
			Line: line,
			Path: path,
		}
		frames = append(frames, frame)
		skip++
	}
	return &errorData{
		err:     err,
		message: message,
		frames:  frames,
	}
}
