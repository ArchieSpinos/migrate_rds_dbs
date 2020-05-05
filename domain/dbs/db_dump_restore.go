package dbs

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/JamesStewy/go-mysqldump"
)

func MysqlDump(request ReplicationRequest) (*string, *errors.DBErr) {
	sourceSQLClient, err := SourceInitConnection(request, true)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to create DB connection: %s:", err.Error()))
	}
	dumpDir := request.MysqlDumpPath // you should create this directory
	dumpFilenameFormat := fmt.Sprintf("%s-20060102T150405", request.SourceDBName)

	// Register database with mysqldump
	dumper, err := mysqldump.Register(sourceSQLClient, dumpDir, dumpFilenameFormat)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Error registering database: %s", err.Error()))
	}

	// Dump database to file
	resultFilename, err := dumper.Dump()
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Error dumping database: %s", err.Error()))
	}
	fmt.Printf("File is saved to %s", resultFilename)

	// Close dumper and connected database
	dumper.Close()
	return aws.String(resultFilename), nil
}

func MysqlRestore(request ReplicationRequest, mysqlDumpFilename string) *errors.DBErr {
	cmd := exec.Command("mysql", "-h", request.DestHost, "-u", request.DestUser, "-p"+request.DestPassword,
		"-D", request.SourceDBName, "-e", "source", mysqlDumpFilename)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("Error restoring database: %s", stderr.String()))
	}
	return nil
}
