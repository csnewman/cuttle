package cuttle

import (
	"errors"
	"strings"
)

var ErrNoCompatibleDialects = errors.New("no compatible dialects")

type DialectCompatFunc func(dialect Dialect) uint

type Dialect struct {
	Name   string
	Compat DialectCompatFunc
}

func (d Dialect) Is(other Dialect) bool {
	return strings.EqualFold(d.Name, other.Name)
}

func (d Dialect) CompatibleWith(other Dialect) bool {
	if d.Is(other) {
		return true
	}

	return d.Compat(other) != 0
}

func (d Dialect) Select(dialects []Dialect) (int, error) {
	var (
		selectedScore uint
		selected      int
	)

	for i, dialect := range dialects {
		if d.Is(dialect) {
			return i, nil
		}

		score := d.Compat(dialect)

		if score > selectedScore {
			selectedScore = score
			selected = i
		}
	}

	if selectedScore == 0 {
		return -1, ErrNoCompatibleDialects
	}

	return selected, nil
}

var DialectGeneric = Dialect{
	Name: "generic",
	Compat: func(Dialect) uint {
		return 0
	},
}

var DialectSQLite = Dialect{
	Name: "sqlite",
	Compat: func(dialect Dialect) uint {
		if dialect.Is(DialectGeneric) {
			return 100
		}

		return 0
	},
}
