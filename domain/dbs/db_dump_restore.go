package dbs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/ArchieSpinos/migrate_rds_dbs/persist"

	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
)

func MysqlDumpExec(request ReplicationRequest, restoredInstanceDNS string, serviceDBs []string, pathGlobal string) *errors.DBErr {
	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)
	fmt.Println(restoredInstanceDNS)

	if err := persist.CreatePath(pathGlobal + "sqldumps"); err != nil {
		return err
	}

	for _, serviceDB := range serviceDBs {
		dumpFile := pathGlobal + "sqldumps/" + serviceDB + ".sql"
		cmd := exec.Command("mysqldump", "--databases", serviceDB, "--single-transaction", "--set-gtid-purged=OFF", "--compress", "--order-by-primary", "-r", dumpFile, "-h", restoredInstanceDNS, "-u", request.SourceUser, "-p"+request.SourcePassword)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return errors.NewInternalServerError(fmt.Sprintf("Error dumping all source host databases: %s", stderr.String()))
		}
	}
	return nil
}

func MysqlRestore(request ReplicationRequest, pathGlobal string) *errors.DBErr {
	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)
	files, err := ioutil.ReadDir(pathGlobal + "/sqldumps")
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Error listing dump files: %s", err.Error()))
	}

	for _, v := range files {
		execute := "source " + pathGlobal + "/sqldumps/" + v.Name()
		cmd := exec.Command("mysql", "-h", request.DestHost, "-u", request.DestUser, "-p"+request.DestPassword, "-e", execute)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return errors.NewInternalServerError(fmt.Sprintf("Error restoring database: %s", stderr.String()))
		}
	}
	return nil
}
