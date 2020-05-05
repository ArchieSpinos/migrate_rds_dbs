package database

import (
	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/services"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/gin-gonic/gin"
)

func SetupRepl(c *gin.Context) {
	var replicationRequest dbs.ReplicationRequest
	if err := c.ShouldBindJSON(&replicationRequest); err != nil {
		dbErr := errors.NewBadRequestError("invalid json body")
		c.JSON(dbErr.Status, dbErr)
		return
	}

	if err := services.PreFlightCheck(replicationRequest); err != nil {
		c.JSON(err.Status, err)
		return
	}

	// binLogRetentionResult, err := services.EnableBinLogRetention(replicationRequest)
	// if err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }
	// c.JSON(http.StatusOK, binLogRetentionResult)
	// fmt.Print("bin log retention query result: s%", binLogRetentionResult)

	// awsSession, err := awsresources.CreateSession()
	// if err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// dbClusters, err := services.RDSDescribeCluster(awsSession, replicationRequest)
	// if err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// restoreClusterInput := services.RDSDescribeToStruct(replicationRequest, dbClusters)

	// restoredDB, err := services.RDSRestoreCluster(awsSession, restoreClusterInput)
	// if err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// createInstanceInput := services.RDSCreateInstanceToStruct(restoredDB)

	// rdsInstance, err := services.RDSCreateInstance(awsSession, createInstanceInput)
	// if err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// binLogFile, binLogPos, err := services.RDSDescribeEvents(awsSession, rdsInstance)
	// if err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// mysqlDumpFilename, err := dbs.MysqlDump(replicationRequest)
	// if err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// if err := dbs.MysqlRestore(replicationRequest, aws.StringValue(mysqlDumpFilename)); err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// c.JSON(http.StatusOK, rdsInstance)

}
