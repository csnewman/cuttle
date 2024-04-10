# cuttle
Opinionated database toolkit for Go.

**Work in progress**: This is incomplete and does not fully function yet!

**Features:**

- TODO
- Transactions
- Batches
- Repository code generator

## Codegen

Cuttle optionally includes a repository pattern code generator: 

```sql
-- :cuttle version=1
-- :repository name=UsersRepository dialects=sqlite,postgres

-- :query name=InsertUser mode=exec
-- :arg name=username type=string
-- :arg name=password type=string
-- :arg name=role type=string
-- :dialect name=sqlite
INSERT INTO users (username, password, role)
VALUES (?1, ?2, ?3);
-- :dialect name=postgres
INSERT INTO users (username, password, role)
VALUES ($1, $2, $3);
```

Produces:

```go
package main

import (
	"context"
	"github.com/csnewman/cuttle"
)

type UsersRepository interface {
	InsertUser(
		ctx context.Context,
		tx cuttle.WTx,
		username string,
		password string,
		role string,
	) (int64, error)

	InsertUserAsync(
		tx cuttle.AsyncWTx,
		username string,
		password string,
		role string,
		callback cuttle.AsyncHandler[int64],
	)
}

// [...]
func NewUsersRepository(dialect cuttle.Dialect) (UsersRepository, error) {
// [...]
```

## Why not use `database/sql`

TODO

## Drivers

Behind the scenes, cuttle uses the following libraries:

| Database | Library                                                            |
|----------|--------------------------------------------------------------------|
| Postgres | [github.com/jackc/pgx](https://github.com/jackc/pgx)               |
| SQLite   | [github.com/tailscale/sqlite](https://github.com/tailscale/sqlite) |
| Other    | TODO (wrapper around `database/sql`)                               |
