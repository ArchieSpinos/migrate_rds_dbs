package dbs

import (
	"fmt"

	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
)

// const (
// 	queryShowDatabases = "show databases;"
// )

func (result *QueryResult) List(dbcon ReplRequest, query string) *errors.DBErr {
	sourceSQLClient, err := SourceInitConnection(dbcon)
	stmt, err := sourceSQLClient.Prepare(query)
	if err != nil {
		return errors.NewInternalServerError(err.Error())
	}
	defer stmt.Close()
	listResult, err := stmt.Query()
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to execute query: %s",
			err.Error()))
	}
	var row string
	for listResult.Next() {
		if err := listResult.Scan(&row); err != nil {
			return errors.NewInternalServerError(fmt.Sprintf("error when retrieving list of dbs's: %s",
				err.Error()))
		}
		fmt.Printf("this %s", row)
		*result = append(*result, row)
		//fmt.Println(len(result))
	}
	return nil
}
