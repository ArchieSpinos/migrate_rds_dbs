package services

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
)

func CheckSlaveStatus(request dbs.ReplicationRequest) (*bytes.Buffer, *errors.DBErr) {
	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)

	execute := `SHOW SLAVE STATUS\G`
	cmd := exec.Command("mysql", "-h", request.DestHost, "-u", request.DestUser, "-p"+request.DestPassword, "-e", execute)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Error retrieving slave status: %s", stderr.String()))
	}
	return &out, nil
}
