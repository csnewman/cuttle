package generator

var dialectConfigs = map[string]dialectConfig{
	"generic": {
		VarName: "DialectGeneric",
		IDEName: "sql",
	},
	"sqlite": {
		VarName: "DialectSQLite",
		IDEName: "sqlite",
	},
	"postgres": {
		VarName: "DialectPostgres",
		IDEName: "postgresql",
	},
}

type dialectConfig struct {
	VarName string
	IDEName string
}
