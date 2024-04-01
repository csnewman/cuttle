package main

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/csnewman/cuttle/internal/parser"
)

func main() {
	file, err := os.Open("example.sql")
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

	parser.Parse(file, logger)
}
