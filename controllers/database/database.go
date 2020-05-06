package database

import (
	"fmt"

	"github.com/ArchieSpinos/migrate_rds_dbs/awsresources"
	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/services"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
)

func SetupRepl(c *gin.Context) {
	var replicationRequest dbs.ReplicationRequest
	if err := c.ShouldBindJSON(&replicationRequest); err != nil {
		dbErr := errors.NewBadRequestError("invalid json body")
		c.JSON(dbErr.Status, dbErr)
		return
	}

	serviceDBs, err := services.PreFlightCheck(replicationRequest)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	binLogRetentionResult, err := services.EnableBinLogRetention(replicationRequest)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}
	fmt.Print("bin log retention query result: s%", binLogRetentionResult) // remove after test

	awsSession, err := awsresources.CreateSession()
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	dbClusters, err := services.RDSDescribeCluster(awsSession, replicationRequest)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	restoreClusterInput := services.RDSDescribeToStruct(replicationRequest, dbClusters)

	restoredDB, err := services.RDSRestoreCluster(awsSession, restoreClusterInput)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	createInstanceInput := services.RDSCreateInstanceToStruct(restoredDB)

	rdsInstance, err := services.RDSCreateInstance(awsSession, createInstanceInput)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	// //mock CreateDBInstanceOutput
	// var instance = "dev-migrate-temp-instance"
	// var address = "dev-migrate-temp-instance.cmnsml8q1eeo.eu-west-1.rds.amazonaws.com"
	// rdsInstance := rds.CreateDBInstanceOutput{
	// 	DBInstance: &rds.DBInstance{
	// 		DBInstanceIdentifier: &instance,
	// 		Endpoint: &rds.Endpoint{
	// 			Address: &address,
	// 		},
	// 	},
	// }

	// binLogFile, binLogPos, err := services.RDSDescribeEvents(awsSession, rdsInstance)
	// if err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	if err := dbs.MysqlDumpExec(replicationRequest, aws.StringValue(rdsInstance.DBInstance.Endpoint.Address), serviceDBs); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := dbs.MysqlRestore(replicationRequest); err != nil {
		c.JSON(err.Status, err)
		return
	}

}
