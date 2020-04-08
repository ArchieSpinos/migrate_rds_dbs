package services

import (
	"fmt"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
)

func EnableBinLogRetention(dbconn dbs.ReplRequest) (*dbs.QueryResult, *errors.DBErr) {
	const (
		queryBinLogRetention = "CALL mysql.rds_set_configuration('binlog retention hours', 144); CALL mysql.rds_show_configuration;"
	)
	result := &dbs.QueryResult{}
	if err := result.List(dbconn, queryBinLogRetention); err != nil {
		return nil, err
	}
	fmt.Println(result)
	return result, nil
}
