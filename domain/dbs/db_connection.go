package dbs

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func SourceInitConnection(request ReplicationRequest) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
		request.SourceUser,
		request.SourcePassword,
		request.SourceHost,
		request.SourceDBName,
	)
	appDB, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		//log(err)
		return nil, err
	}

	if err = appDB.Ping(); err != nil {
		//log(err)
		return nil, err
	}
	log.Println("database connection successfully configured")
	return appDB, nil
}
