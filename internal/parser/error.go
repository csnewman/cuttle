package parser

import "fmt"

type SrcError struct {
	Token *Token
	Inner error
}

func newSrcError(tk *Token, inner error) *SrcError {
	return &SrcError{
		Token: tk,
		Inner: inner,
	}
}

func wrapSrcError(tk *Token, format string, a ...any) *SrcError {
	return newSrcError(tk, fmt.Errorf(format, a...)) //nolint:goerr113
}

func (e *SrcError) Error() string {
	return fmt.Sprintf("%v-%v: %v", e.Token.Start, e.Token.End, e.Inner)
}
