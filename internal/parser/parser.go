package parser

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

var ErrDocAlreadyExists = errors.New("doc comment already exists")

type Unit struct {
	Repositories      map[string]*Repository
	RepositoriesOrder []string
}

type Repository struct {
	Name    string
	Queries []*Query
}

type Query struct {
	Name          string
	Doc           *Doc
	Args          []*Arg
	Cols          []*Col
	Variants      map[string]*Variant
	VariantsOrder []string
}

type Variant struct {
	Name    string
	Content []string
}

type Doc struct {
	Lines []string
}

type Arg struct {
	Name string
	Type string
}

type Col struct {
	Name string
	Type string
}

type parser struct {
	tz     *Tokenizer
	logger *slog.Logger
	queued *Token
	unit   *Unit
}

func Parse(in io.Reader, file string, logger *slog.Logger) (*Unit, error) {
	tz := NewTokenizer(in, file)

	p := &parser{
		tz:     tz,
		logger: logger,
		unit: &Unit{
			Repositories: make(map[string]*Repository),
		},
	}

	if err := p.Parse(); err != nil {
		return nil, err
	}

	return p.unit, nil
}

func (p *parser) next() (*Token, error) {
	if p.queued != nil {
		t := p.queued
		p.queued = nil

		return t, nil
	}

	return p.tz.Next()
}

func (p *parser) queue(t *Token) {
	if p.queued != nil {
		panic("queue not empty")
	}

	p.queued = t
}

func (p *parser) Parse() error {
	// Find start
	var meta *Directive

	for {
		tk, err := p.next()
		if err != nil {
			return fmt.Errorf("failed to find start: %w", err)
		}

		if tk.IsDirective(DirectiveTypeCuttle) {
			meta, err = tk.ParseDirective()
			if err != nil {
				return wrapSrcError(tk, "failed to parse cuttle directive: %w", err)
			}

			break
		}
	}

	p.logger.Debug("Found meta", "meta", meta)

	for {
		tk, err := p.next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return wrapSrcError(tk, "failed to parse token: %w", err)
		}

		if tk.Type != TokenTypeDirective {
			continue
		}

		dir, err := tk.ParseDirective()
		if err != nil {
			return wrapSrcError(tk, "failed to parse top level directive: %w", err)
		}

		if dir.Type == DirectiveTypeMigration {
			if err := p.parseMigration(dir); err != nil {
				return fmt.Errorf("failed to parse migration: %w", err)
			}
		} else if dir.Type == DirectiveTypeRepository {
			if err := p.parseRepository(dir); err != nil {
				return fmt.Errorf("failed to parse repository: %w", err)
			}
		} else {
			return wrapSrcError(tk, "%w: unexpected top level directive: %v", ErrInvalidInput, dir.Type)
		}
	}

	return nil
}

func (p *parser) parseMigration(_ *Directive) error {
	p.logger.Debug("Ignoring migration")

	for {
		tk, err := p.next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return wrapSrcError(tk, "failed to parse migration token: %w", err)
		}

		if tk.Type != TokenTypeDirective {
			continue
		}

		dir, err := tk.ParseDirective()
		if err != nil {
			return wrapSrcError(tk, "failed to parse migration directive: %w", err)
		}

		if dir.Type == DirectiveTypeMigration || dir.Type == DirectiveTypeRepository {
			p.queue(tk)

			break
		}
	}

	return nil
}

func (p *parser) parseRepository(dir *Directive) error {
	name, ok := dir.Values["name"]
	if !ok {
		return wrapSrcError(dir.Token, "%w: no name provided", ErrInvalidInput)
	}

	p.logger.Debug("Parsing repository", "name", name)

	repo, ok := p.unit.Repositories[name]
	if !ok {
		repo = &Repository{
			Name: name,
		}

		p.unit.Repositories[name] = repo
		p.unit.RepositoriesOrder = append(p.unit.RepositoriesOrder, name)
	}

	for {
		tk, err := p.next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return wrapSrcError(tk, "failed to parse repository token: %w", err)
		}

		if tk.Type != TokenTypeDirective {
			continue
		}

		dir, err := tk.ParseDirective()
		if err != nil {
			return wrapSrcError(tk, "failed to parse repository directive: %w", err)
		}

		if dir.Type == DirectiveTypeMigration || dir.Type == DirectiveTypeRepository {
			p.queue(tk)

			break
		}

		if dir.Type == DirectiveTypeQuery {
			query, err := p.parseQuery(dir)
			if err != nil {
				return fmt.Errorf("failed to parse query: %w", err)
			}

			repo.Queries = append(repo.Queries, query)
		} else {
			return wrapSrcError(tk, "%w: unexpected repository directive: %v", ErrInvalidInput, dir.Type)
		}
	}

	return nil
}

func (p *parser) parseQuery(dir *Directive) (*Query, error) {
	query := &Query{
		Variants: make(map[string]*Variant),
	}
	ok := false

	query.Name, ok = dir.Values["name"]
	if !ok {
		return nil, wrapSrcError(dir.Token, "%w: no name provided", ErrInvalidInput)
	}

	p.logger.Debug("Parsing query", "name", query.Name)

	dialect := "generic"

	for {
		tk, err := p.next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, wrapSrcError(tk, "failed to parse query token: %w", err)
		}

		if tk.Type == TokenTypeText {
			variant, ok := query.Variants[dialect]
			if !ok {
				variant = &Variant{
					Name: dialect,
				}

				query.Variants[dialect] = variant
				query.VariantsOrder = append(query.VariantsOrder, dialect)
			}

			variant.Content = append(variant.Content, tk.Content...)

			continue
		}

		if tk.Type != TokenTypeDirective {
			return nil, wrapSrcError(tk, "%w: unexpected query token: %v", ErrInvalidInput, tk.Type)
		}

		dir, err := tk.ParseDirective()
		if err != nil {
			return nil, wrapSrcError(tk, "failed to parse query directive: %w", err)
		}

		if dir.Type == DirectiveTypeMigration || dir.Type == DirectiveTypeRepository || dir.Type == DirectiveTypeQuery {
			p.queue(tk)

			break
		}

		switch dir.Type {
		case DirectiveTypeArg:
			a, err := p.parseArg(dir)
			if err != nil {
				return nil, fmt.Errorf("failed to parse arg: %w", err)
			}

			query.Args = append(query.Args, a)

		case DirectiveTypeCol:
			c, err := p.parseCol(dir)
			if err != nil {
				return nil, fmt.Errorf("failed to parse col: %w", err)
			}

			query.Cols = append(query.Cols, c)

		case DirectiveTypeDoc:
			d, err := p.parseDoc(dir)
			if err != nil {
				return nil, fmt.Errorf("failed to parse doc: %w", err)
			}

			if query.Doc != nil {
				return nil, newSrcError(tk, ErrDocAlreadyExists)
			}

			query.Doc = d

		case DirectiveTypeDialect:
			dialect, ok = dir.Values["name"]
			if !ok {
				return nil, wrapSrcError(dir.Token, "%w: no name provided", ErrInvalidInput)
			}

		default:
			return nil, wrapSrcError(tk, "%w: unexpected query directive: %v", ErrInvalidInput, dir.Type)
		}
	}

	return query, nil
}

func (p *parser) parseArg(dir *Directive) (*Arg, error) {
	arg := &Arg{}
	ok := false

	arg.Name, ok = dir.Values["name"]
	if !ok {
		return nil, wrapSrcError(dir.Token, "%w: no name provided", ErrInvalidInput)
	}

	arg.Type, ok = dir.Values["type"]
	if !ok {
		return nil, wrapSrcError(dir.Token, "%w: no type provided", ErrInvalidInput)
	}

	return arg, nil
}

func (p *parser) parseCol(dir *Directive) (*Col, error) {
	col := &Col{}
	ok := false

	col.Name, ok = dir.Values["name"]
	if !ok {
		return nil, wrapSrcError(dir.Token, "%w: no name provided", ErrInvalidInput)
	}

	col.Type, ok = dir.Values["type"]
	if !ok {
		return nil, wrapSrcError(dir.Token, "%w: no type provided", ErrInvalidInput)
	}

	return col, nil
}

func (p *parser) parseDoc(dir *Directive) (*Doc, error) {
	doc := &Doc{}

	for {
		tk, err := p.next()
		if errors.Is(err, io.EOF) {
			return nil, wrapSrcError(dir.Token, "%w: unexpected eof inside doc block", ErrInvalidInput)
		} else if err != nil {
			return nil, wrapSrcError(tk, "failed to parse doc token: %w", err)
		}

		if tk.Type == TokenTypeText {
			doc.Lines = append(doc.Lines, tk.Content...)

			continue
		}

		if tk.Type != TokenTypeDirective {
			return nil, wrapSrcError(tk, "%w: unexpected doc token: %v", ErrInvalidInput, tk.Type)
		}

		dir, err := tk.ParseDirective()
		if err != nil {
			return nil, wrapSrcError(tk, "failed to parse doc directive: %w", err)
		}

		if dir.Type == DirectiveTypeEnd {
			break
		}

		return nil, wrapSrcError(tk, "%w: unexpected doc directive: %v", ErrInvalidInput, dir.Type)
	}

	return doc, nil
}
