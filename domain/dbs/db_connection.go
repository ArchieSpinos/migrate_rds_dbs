package dbs

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func SourceInitConnection(dbcon ReplRequest) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
		dbcon.SourceUser,
		dbcon.SourcePassword,
		dbcon.SourceHost,
		dbcon.SourceName,
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
	log.Println("database successfully configured")
	return appDB, nil
}
