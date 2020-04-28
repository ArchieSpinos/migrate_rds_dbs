package dbs

import (
	"fmt"

	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
)

func (result *QueryResult) LogRetention(dbcon ReplRequest, query string) *errors.DBErr {
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
	var row string
	for listResult.Next() {
		// var message string

		if err := listResult.Scan(&row); err != nil {
			return errors.NewInternalServerError(fmt.Sprintf("error when scanning row: %s",
				err.Error()))
		}
		// if err := json.Unmarshal(row, &message); err != nil {
		// 	return errors.NewInternalServerError(fmt.Sprintf("error when unmarshalling query result to string: %s",
		// 		err.Error()))
		// }
		*result = append(*result, row)
	}
	return nil
}
