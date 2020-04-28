package dbs

import (
	"fmt"

	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
)

func (result *QueryResult) MultiQueryLogRetention(dbcon ReplRequest, query string) *errors.DBErr {
	sourceSQLClient, err := SourceInitConnection(dbcon)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to create DB connection: %s:", err.Error()))
	}
	stmt, err := sourceSQLClient.Prepare(query)
	defer stmt.Close()
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to prepare sql query: %s:", err.Error()))
	}
	listResult, err := stmt.Query()
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to execute query: %s",
			err.Error()))
	}
	cols, err := listResult.Columns()
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to get columns from query: %s",
			err.Error()))
	}

	rawResult := make([][]byte, len(cols))
	resultTemp := make([]string, len(cols))

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice

	for i := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for listResult.Next() {
		if err := listResult.Scan(dest...); err != nil {
			return errors.NewInternalServerError(fmt.Sprintf("error when scanning row: %s",
				err.Error()))
		}
		for i, raw := range rawResult {
			if raw == nil {
				resultTemp[i] = "\\N"
			} else {
				resultTemp[i] = string(raw)
				*result = append(*result, string(raw))
			}
		}
	}
	return nil
}
