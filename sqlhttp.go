package sqlhttp

import (
	"database/sql"
	"encoding/json"
	"strings"
)

type SqlHttp struct {
	db *sql.DB
}

type SqlHttpRequest struct {
	Statements []Statement
}

type Statement struct {
	Statement string
	Params    []any
}

func NewSqlHttp(db *sql.DB) *SqlHttp {
	return &SqlHttp{db: db}
}

func (s *SqlHttp) Request(request SqlHttpRequest) ([][]byte, error) {
	responses := make([][]byte, len(request.Statements))

	for i, statement := range request.Statements {
		response, err := s.innerRequest(statement)

		if err != nil {
			return responses, err
		}

		responses[i] = response
	}

	return responses, nil
}

func (s *SqlHttp) innerRequest(statement Statement) ([]byte, error) {
	stmt, err := s.db.Prepare(statement.Statement)

	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(statement.Statement, "insert") || strings.HasPrefix(statement.Statement, "update") {
		return s.i(stmt, statement)
	}

	return s.s(stmt, statement)
}

func (s *SqlHttp) i(stmt *sql.Stmt, statement Statement) ([]byte, error) {
	res, err := stmt.Exec(statement.Params...)

	if err != nil {
		return nil, err
	}

	lii, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	ra, err := res.RowsAffected()

	if err != nil {
		return nil, err
	}

	type InsertResponse struct {
		LastInsertId int64
		RowsAffected int64
	}

	return json.Marshal(InsertResponse{LastInsertId: lii, RowsAffected: ra})
}

func (s *SqlHttp) s(stmt *sql.Stmt, statement Statement) ([]byte, error) {
	rows, err := stmt.Query(statement.Params...)

	if err != nil {
		return nil, err
	}

	return s.rowsToJson(rows)
}

func (s *SqlHttp) rowsToJson(rows *sql.Rows) ([]byte, error) {
	var err error

	columns, err := rows.Columns()

	if err != nil {
		return nil, err
	}

	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}

	return json.Marshal(tableData)
}
