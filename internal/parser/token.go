package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

var ErrInvalidInput = errors.New("invalid input")

type TokenType string

const (
	TokenTypeDirective = "directive"
	TokenTypeText      = "text"
)

type DirectiveType string

const (
	DirectiveTypeCuttle     = "cuttle"
	DirectiveTypeEnd        = "end"
	DirectiveTypeMigration  = "migration"
	DirectiveTypeApply      = "apply"
	DirectiveTypeStep       = "step"
	DirectiveTypeRevert     = "revert"
	DirectiveTypeRepository = "repository"
	DirectiveTypeQuery      = "query"
	DirectiveTypeArg        = "arg"
	DirectiveTypeDoc        = "doc"
	DirectiveTypeCol        = "col"
)

type Token struct {
	Type     TokenType
	Source   string
	Start    int
	End      int
	Content  []string
	RawLines []string
}

func (t *Token) IsDirective(name DirectiveType) bool {
	if t.Type != TokenTypeDirective {
		return false
	}

	if len(t.Content) > 1 {
		panic("more than 1 content line in directive")
	}

	key, _, _ := strings.Cut(t.Content[0], " ")

	return strings.EqualFold(string(name), key)
}

type Directive struct {
	Token  *Token
	Type   DirectiveType
	Values map[string]string
}

func (t *Token) ParseDirective() (*Directive, error) {
	if t.Type != TokenTypeDirective {
		panic("not directive")
	}

	if len(t.Content) > 1 {
		panic("more than 1 content line in directive")
	}

	rawTy, vals, _ := strings.Cut(t.Content[0], " ")

	ty := DirectiveType(strings.ToLower(rawTy))

	vals = strings.TrimSpace(vals)

	parsedVals := make(map[string]string)
	inQuotes := false
	key := ""

	var sb strings.Builder

	for _, r := range vals {
		if inQuotes {
			if r == '"' {
				parsedVals[key] = sb.String()
				key = ""
				inQuotes = false

				sb.Reset()

				continue
			}

			sb.WriteRune(r)
		} else if unicode.IsSpace(r) {
			if key != "" {
				parsedVals[key] = sb.String()
			}

			sb.Reset()
			key = ""
		} else if r == '=' {
			key = sb.String()
			sb.Reset()
		} else if r == '"' {
			if key == "" {
				return nil, fmt.Errorf("%w: quotes only allowed in values", ErrInvalidInput)
			}

			inQuotes = true
			sb.Reset()
		} else {
			sb.WriteRune(r)
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("%w: unterminated quotes", ErrInvalidInput)
	}

	if sb.Len() > 0 && key == "" {
		key = sb.String()
		sb.Reset()
	}

	if key != "" {
		parsedVals[key] = sb.String()
	}

	return &Directive{
		Token:  t,
		Type:   ty,
		Values: parsedVals,
	}, nil
}

type Tokenizer struct {
	scanner *bufio.Scanner
	queued  *Token
	line    int
}

func NewTokenizer(in io.Reader) *Tokenizer {
	scanner := bufio.NewScanner(in)

	return &Tokenizer{
		scanner: scanner,
		line:    -1,
	}
}

func (t *Tokenizer) Next() (*Token, error) {
	if t.queued != nil {
		tk := t.queued
		t.queued = nil

		return tk, nil
	}

	var text []string

	textStart := -1
	textEnd := -1

	for t.scanner.Scan() {
		t.line++

		line := t.scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "--") {
			cmd := strings.TrimSpace(strings.TrimPrefix(trimmed, "--"))

			if strings.HasPrefix(cmd, ":") {
				cmd = strings.TrimPrefix(cmd, ":")

				tk := &Token{
					Type:     TokenTypeDirective,
					Source:   "",
					Start:    t.line,
					End:      t.line,
					Content:  []string{cmd},
					RawLines: []string{line},
				}

				if len(text) > 0 {
					t.queued = tk

					break
				}

				return tk, nil
			}
		}

		if textStart == -1 {
			textStart = t.line
		}

		textEnd = t.line
		text = append(text, line)
	}

	if len(text) == 0 {
		return nil, io.EOF
	}

	if t.scanner.Err() != nil {
		return nil, t.scanner.Err()
	}

	return &Token{
		Type:     TokenTypeText,
		Source:   "",
		Start:    textStart,
		End:      textEnd,
		Content:  text,
		RawLines: text,
	}, nil
}
