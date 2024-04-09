package generator

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/csnewman/cuttle/internal/parser"
	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

const cuttlePkg = "github.com/csnewman/cuttle"

func Generate(unit *parser.Unit, logger *slog.Logger, outPath string) error {
	f := jen.NewFile("main")
	f.HeaderComment("Code generated by " + cuttlePkg + ". DO NOT EDIT")

	g := &Generator{
		logger: logger,
		file:   f,
	}

	g.Generate(unit)

	return f.Save(outPath)
}

type Generator struct {
	logger *slog.Logger
	file   *jen.File
}

func (g *Generator) Generate(unit *parser.Unit) {
	g.file.ImportName(cuttlePkg, "cuttle")

	for _, name := range unit.RepositoriesOrder {
		repo := unit.Repositories[name]

		g.GenerateRepo(repo)
	}
}

func (g *Generator) GenerateRepo(repo *parser.Repository) {
	g.logger.Debug("Generating repository", "name", repo.Name)

	g.file.Type().Id(repo.Name).InterfaceFunc(func(jg *jen.Group) {
		implName := strcase.ToLowerCamel(repo.Name) + "Impl"
		dialectsVar := implName + "Dialects"

		g.file.Line()
		g.file.Type().Id(implName).StructFunc(func(jg *jen.Group) {
			jg.Id("dialect").Qual(cuttlePkg, "Dialect")
			jg.Id("dialectIndex").Qual("", "int")
		})

		g.file.Line()
		g.file.Var().Id(dialectsVar).Op("=").Index().Qual(cuttlePkg, "Dialect").ValuesFunc(func(jg *jen.Group) {
			jg.Line().Qual(cuttlePkg, "DialectGeneric")

			jg.Line()
		})

		g.file.Line()
		g.file.Func().Id("New" + repo.Name).ParamsFunc(func(jg *jen.Group) {
			jg.Id("dialect").Qual(cuttlePkg, "Dialect")
		}).ParamsFunc(func(jg *jen.Group) {
			jg.Qual("", repo.Name)
			jg.Qual("", "error")
		}).BlockFunc(func(jg *jen.Group) {
			jg.Id("selected").Op(",").Id("err").
				Op(":=").
				Id("dialect").Dot("Select").Params(jen.Id(dialectsVar))
			jg.If(jen.Id("err").Op("!=").Id("nil")).BlockFunc(func(jg *jen.Group) {
				jg.Return(jen.Id("nil"), jen.Id("err"))
			})
			jg.Line()
			jg.Return(
				jen.Op("&").Id(implName).Values(jen.Dict{
					jen.Id("dialect"):      jen.Id(dialectsVar).Index(jen.Id("selected")),
					jen.Id("dialectIndex"): jen.Id("selected"),
				}),
				jen.Id("nil"),
			)
		})

		for _, query := range repo.Queries {
			g.generateQuery(repo, query, jg, implName)

			jg.Line()
		}
	})
}

func (g *Generator) generateQuery(repo *parser.Repository, query *parser.Query, jg *jen.Group, implName string) {
	g.logger.Debug("Generating query", "name", query.Name)

	var (
		resultPath string
		resultType string
		txType     string
	)

	resultPath = ""
	resultType = "int64"
	txType = "WTx"

	jg.Id(query.Name).ParamsFunc(func(jg *jen.Group) {
		jg.Line().Id("ctx").Qual("context", "Context")
		jg.Line().Id("tx").Qual(cuttlePkg, txType)

		for _, arg := range query.Args {
			jg.Line().Id(arg.Name).Qual("", arg.Type)
		}

		jg.Line()
	}).ParamsFunc(func(jg *jen.Group) {
		jg.Qual(resultPath, resultType)
		jg.Id("error")
	})

	jg.Line()

	jg.Id(query.Name + "Async").ParamsFunc(func(jg *jen.Group) {
		jg.Line().Id("tx").Qual(cuttlePkg, "Async"+txType)

		for _, arg := range query.Args {
			jg.Line().Id(arg.Name).Qual("", arg.Type)
		}

		jg.Line().Id("callback").Qual(cuttlePkg, "AsyncHandler").Types(
			jen.Qual(resultPath, resultType),
		)

		jg.Line()
	})

	generateStmtSelector := func(jg *jen.Group) {
		jg.Var().Id("cuttleStmt").Id("string")

		var cases []jen.Code

		for i, vname := range query.VariantsOrder {
			variant := query.Variants[vname]

			cases = append(cases, jen.Case(jen.Lit(i)).BlockFunc(func(jg *jen.Group) {
				stmt := ""

				for j, l := range variant.Content {
					l = strings.TrimSpace(l)

					if strings.HasPrefix(l, "--") {
						continue
					}

					if j > 0 {
						stmt += "\n" + l
					} else {
						stmt += l
					}
				}

				stmt = strings.TrimSpace(stmt)
				stmt = strings.TrimSuffix(stmt, ";")
				stmt = strings.TrimSpace(stmt)

				stmt = fmt.Sprintf("/* %v:%v */ %v", repo.Name, query.Name, stmt)

				jg.Comment("language=sql")
				jg.Id("cuttleStmt").Op("=").Custom(jen.Options{
					Open:      "`",
					Close:     "`",
					Separator: "",
					Multi:     false,
				}, jen.Id(stmt))
			}))
		}

		cases = append(cases, jen.Default().Block(jen.Panic(jen.Lit("unknown dialect"))))

		jg.Switch(jen.Id("r").Dot("dialectIndex")).Block(cases...)
	}

	g.file.Line()
	g.file.Func().Params(jen.Id("r").Op("*").Id(implName)).Id(query.Name).
		ParamsFunc(func(jg *jen.Group) {
			jg.Line().Id("ctx").Qual("context", "Context")
			jg.Line().Id("tx").Qual(cuttlePkg, txType)

			for _, arg := range query.Args {
				jg.Line().Id(arg.Name).Qual("", arg.Type)
			}

			jg.Line()
		}).
		ParamsFunc(func(jg *jen.Group) {
			jg.Qual(resultPath, resultType)
			jg.Id("error")
		}).
		BlockFunc(func(jg *jen.Group) {
			generateStmtSelector(jg)

			jg.Line()

			jg.Id("cuttleRes").Op(",").Id("cuttleErr").Op(":=").
				Id("tx").Dot("Exec").
				ParamsFunc(func(jg *jen.Group) {
					jg.Line().Id("ctx")
					jg.Line().Id("cuttleStmt")

					for _, arg := range query.Args {
						jg.Line().Id(arg.Name)
					}

					jg.Line()
				})

			jg.If(jen.Id("cuttleErr").Op("!=").Nil()).
				Block(jen.Return(jen.Id("-1"), jen.Id("cuttleErr")))

			jg.Line()
			jg.Return(jen.Id("cuttleRes").Dot("RowsAffected").Params(), jen.Nil())
		})

	g.file.Line()
	g.file.Func().Params(jen.Id("r").Op("*").Id(implName)).Id(query.Name + "Async").
		ParamsFunc(func(jg *jen.Group) {
			jg.Line().Id("tx").Qual(cuttlePkg, "Async"+txType)

			for _, arg := range query.Args {
				jg.Line().Id(arg.Name).Qual("", arg.Type)
			}

			jg.Line().Id("callback").Qual(cuttlePkg, "AsyncHandler").Types(
				jen.Qual(resultPath, resultType),
			)

			jg.Line()
		}).
		BlockFunc(func(jg *jen.Group) {
			generateStmtSelector(jg)

			jg.Line()
			jg.Id("_").Op("=").Id("cuttleStmt")
		})
}
