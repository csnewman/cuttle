package parser

import "fmt"

type SrcError struct {
	tk    *Token
	inner error
}

func newSrcError(tk *Token, inner error) *SrcError {
	return &SrcError{
		tk:    tk,
		inner: inner,
	}
}

func wrapSrcError(tk *Token, format string, a ...any) *SrcError {
	return newSrcError(tk, fmt.Errorf(format, a...)) //nolint:goerr113
}

func (e *SrcError) Error() string {
	return fmt.Sprintf("%v-%v: %v", e.tk.Start, e.tk.End, e.inner)
}
