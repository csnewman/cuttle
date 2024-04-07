package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/csnewman/cuttle/internal/parser"
)

func main() {
	path := "example.sql"

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.SourceKey {
				//nolint:forcetypeassert
				source := a.Value.Any().(*slog.Source)
				source.File = strings.TrimPrefix(source.File, "github.com/csnewman/cuttle/")
			}

			return a
		},
	}))

	unit, err := parser.Parse(file, path, logger)

	if err != nil {
		var el *parser.SrcError

		if errors.As(err, &el) {
			fmt.Println()

			for i, s := range el.Token.RawLines {
				fmt.Printf("%v:%v: %v\n", el.Token.Source, el.Token.Start+i, s)
			}

			fmt.Printf("%v:%v-%v: %v\n", el.Token.Source, el.Token.Start, el.Token.End, el.Inner)

			return
		}

		panic(err)
	}

	_ = unit
}