package services

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
)

// CheckSlaveStatus checks that slace in mysql transactional replication
// has caught up with master after initial restore.
func CheckSlaveStatus(request dbs.ReplicationRequest) ([]string, *errors.DBErr) {
	var (
		out                   bytes.Buffer
		stderr                bytes.Buffer
		replicationValidation []string
	)

	execute := `SHOW SLAVE STATUS\G`
	cmd := exec.Command("mysql", "-h", request.DestHost, "-u", request.DestUser, "-p"+request.DestPassword, "-e", execute)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("Error retrieving slave status: %s", stderr.String()))
	}
	split := strings.Split(out.String(), "\n")
	replicationValidation = append(replicationValidation, strings.TrimSpace(split[1]), strings.TrimSpace(split[19]), strings.TrimSpace(split[20]), strings.TrimSpace(split[33]), strings.TrimSpace(split[43]), strings.TrimSpace(split[45]))

	for i, k := range replicationValidation {
		if (i == 0 && k != "Slave_IO_State: Waiting for master to send event") ||
			(i == 1 && k != "Last_Errno: 0") ||
			(i == 2 && k != "Last_Error:") ||
			(i == 3 && k != "Seconds_Behind_Master: 0") ||
			(i == 4 && k != "SQL_Delay: 0") ||
			(i == 5 && k != "Slave_SQL_Running_State: Slave has read all relay log; waiting for more updates") {
			return nil, errors.NewInternalServerError(fmt.Sprintf("Transactional replication is not yet ready. Check status: %s", replicationValidation))
		}
	}
	return replicationValidation, nil
}
