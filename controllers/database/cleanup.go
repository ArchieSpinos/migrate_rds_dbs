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
	if err := c.ShouldBindJSON(&replicationRequest); err != nil {
		dbErr := errors.NewBadRequestError("invalid json body")
		c.JSON(dbErr.Status, dbErr)
		return
	}
	var (
		pathGlobal               = "/tmp/" + replicationRequest.SourceDBName + "/"
		describeDBInstanceOutput = &rds.DescribeDBInstancesOutput{}
		serviceDBsDest           = &[]string{}
	)
	awsSession, err := awsresources.CreateSession(replicationRequest.AwsRegion, replicationRequest.AwsProfile)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := persist.Load(pathGlobal+"describeInstance", describeDBInstanceOutput); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := persist.Load(pathGlobal+"serviceDBsDest", &serviceDBsDest); err != nil {
		c.JSON(err.Status, err)
		return
	}

	serviceDBsDestString := services.SlicePointerToSlice(serviceDBsDest)

	if err := services.PromoteSlave(replicationRequest); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := services.RDSDeleteInstance(awsSession, *describeDBInstanceOutput); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := services.CleanUpDBs(replicationRequest, serviceDBsDestString); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := persist.DeletePath(pathGlobal); err != nil {
		c.JSON(err.Status, err)
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Cleanup has been complete. %s database has been dropped from %s and %s temp RDS cluster instance has been deleted", replicationRequest.SourceDBName, replicationRequest.SourceClusterID, aws.StringValue(describeDBInstanceOutput.DBInstances[0].DBInstanceIdentifier)))
}
