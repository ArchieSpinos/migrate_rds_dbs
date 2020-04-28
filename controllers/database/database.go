package database

import (
	"net/http"

	"github.com/ArchieSpinos/migrate_rds_dbs/awsresources"
	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/services"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/gin-gonic/gin"
)

func SetupRepl(c *gin.Context) {
	var replreq dbs.ReplRequest
	if err := c.ShouldBindJSON(&replreq); err != nil {
		dbErr := errors.NewBadRequestError("invalid json body")
		c.JSON(dbErr.Status, dbErr)
		return
	}
	binLogRetentionResult, listErr := services.EnableBinLogRetention(replreq)
	if listErr != nil {
		c.JSON(listErr.Status, listErr)
		return
	}
	c.JSON(http.StatusOK, binLogRetentionResult)

	awsSession, awsErr := awsresources.CreateSession()
	if awsErr != nil {
		c.JSON(awsErr.Status, awsErr)
		return
	}

	dbClusters, err := services.RDSDescribeCluster(awsSession, replreq)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	restoreClusterInput := services.RDSDescribeToStruct(replreq, dbClusters)

	restoredDB, restoreErr := services.RDSRestoreCluster(awsSession, restoreClusterInput)
	if restoreErr != nil {
		c.JSON(restoreErr.Status, restoreErr)
		return
	}
	// binlogPos = rds.DescribeEvents(restoredDB)

	// result := dbClusters.DBClusters[0]
	c.JSON(http.StatusOK, restoredDB)

}
