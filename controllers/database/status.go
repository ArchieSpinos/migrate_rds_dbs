package database

import (
	"fmt"
	"net/http"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/services"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/gin-gonic/gin"
)

func SecondsBehindMaster(c *gin.Context) {
	var replicationRequest dbs.ReplicationRequest
	if err := c.ShouldBindJSON(&replicationRequest); err != nil {
		dbErr := errors.NewBadRequestError("invalid json body")
		c.JSON(dbErr.Status, dbErr)
		return
	}

	result, err := services.CheckSlaveStatus(replicationRequest)
	if err != nil {
		c.JSON(err.Status, err)
	} else {
		c.JSON(http.StatusOK, fmt.Sprintf("Transactional replication has been completed with status: %s", result))
	}
}
