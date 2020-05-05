package dbs

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func SourceInitConnection(request ReplicationRequest, source bool) (*sql.DB, error) {
	var (
		host     string
		user     string
		password string
	)
	if source {
		host = request.SourceHost
		user = request.SourceUser
		password = request.SourcePassword
	} else {
		host = request.DestHost
		user = request.DestUser
		password = request.DestPassword
	}

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
		user,
		password,
		host,
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
