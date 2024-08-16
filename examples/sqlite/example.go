package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/strattondev/sqlhttp"
)

func main() {
	os.Remove("./foo.db")

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	sqlStmt := `
	create table foo (id integer not null primary key, name text);
	insert into foo (id, name) VALUES(1, "test");
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	req := sqlhttp.SqlHttpRequest{
		Statements: []sqlhttp.Statement{
			{
				Statement: "select * from foo where id = 1",
			},
			{
				Statement: "select * from foo where id = 2",
			},
			{
				Statement: "select * from foo where id = ?",
				Params:    []any{"1"},
			},
			{
				Statement: "select name from foo where id = ?",
				Params:    []any{"1"},
			},
			{
				Statement: "select id, name from foo where id = ?",
				Params:    []any{1},
			},
			{
				Statement: "insert into foo (id, name) VALUES(2, \"another name\");",
			},
			{
				Statement: "select id, name from foo where id = ?",
				Params:    []any{2},
			},
			{
				Statement: "insert into foo (id, name) VALUES(?, ?);",
				Params:    []any{3, "third name"},
			},
			{
				Statement: "select id, name from foo where id = ?",
				Params:    []any{3},
			},
			{
				Statement: "update foo set name = ? where id = ?",
				Params:    []any{"third name updated", 3},
			},
			{
				Statement: "select id, name from foo where id = ?",
				Params:    []any{3},
			},
		},
	}

	res, err := sqlhttp.NewSqlHttp(db).Request(req)

	if err != nil {
		panic(err)
	}

	for i, r := range res {
		slog.Info("", "i", i, "res", string(r))
	}
}
