package database

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/ArchieSpinos/migrate_rds_dbs/awsresources"
	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/persist"
	"github.com/ArchieSpinos/migrate_rds_dbs/services"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/gin-gonic/gin"
)

func PromoteSlave(c *gin.Context) {
	var replicationRequest dbs.ReplicationRequest
	awsSession, err := awsresources.CreateSession()
	if err != nil {
		c.JSON(err.Status, err)
		return
	}
	if err := c.ShouldBindJSON(&replicationRequest); err != nil {
		dbErr := errors.NewBadRequestError("invalid json body")
		c.JSON(dbErr.Status, dbErr)
		return
	}
	var (
		pathGlobal             = "/tmp/" + replicationRequest.SourceDBName + "/"
		createDBInstanceOutput = rds.CreateDBInstanceOutput{}
		serviceDBsDest         []string
	)

	if err := persist.Load(pathGlobal+"rdsInstance", createDBInstanceOutput); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := persist.Load(pathGlobal+"serviceDBsDest", serviceDBsDest); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := services.PromoteSlave(replicationRequest); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := services.RDSDeleteInstance(awsSession, createDBInstanceOutput); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := services.CleanUpDBs(replicationRequest, serviceDBsDest); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := persist.DeletePath(pathGlobal); err != nil {
		c.JSON(err.Status, err)
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Cleanup has been complete. %s database has been dropped from %s and %s temp RDS cluster instance has been deleted", replicationRequest.SourceDBName, replicationRequest.SourceClusterID, aws.StringValue(createDBInstanceOutput.DBInstance.DBInstanceArn)))
}
