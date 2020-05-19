package database

import (
	"fmt"
	"net/http"

	"github.com/ArchieSpinos/migrate_rds_dbs/awsresources"
	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/persist"
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
	var pathGlobal = "/tmp/" + replicationRequest.SourceDBName + "/"

	serviceDBsSource, serviceDBsDest, err := services.PreFlightCheck(replicationRequest, pathGlobal)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	fmt.Println(serviceDBsSource)
	fmt.Println(serviceDBsDest)

	if err := persist.CreatePath(pathGlobal); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := persist.Save(pathGlobal, "serviceDBsDest", serviceDBsDest); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := services.BootstrapReplication(replicationRequest); err != nil {
		c.JSON(err.Status, err)
		return
	}

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

	createInstanceInput := services.RDSCreateInstanceToStruct(restoredDB, replicationRequest)

	rdsInstance, err := services.RDSCreateInstance(awsSession, createInstanceInput)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	// // mock CreateDBInstanceOutput
	// var instance = "migrate-temp-instance-source3"
	// var address = "migrate-temp-instance-source3.cmnsml8q1eeo.eu-west-1.rds.amazonaws.com"
	// rdsInstance := &rds.CreateDBInstanceOutput{
	// 	DBInstance: &rds.DBInstance{
	// 		DBInstanceIdentifier: &instance,
	// 		Endpoint: &rds.Endpoint{
	// 			Address: &address,
	// 		},
	// 	},
	// }

	if err := services.RDSWaitUntilInstanceAvailable(awsSession, rdsInstance); err != nil {
		c.JSON(err.Status, err)
		return
	}

	describeInstance, err := services.RDSWaitForAddress(awsSession, rdsInstance)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := persist.Save(pathGlobal, "describeInstance", describeInstance); err != nil {
		c.JSON(err.Status, err)
		return
	}

	binLogFile, binLogPos, err := services.RDSDescribeEvents(awsSession, rdsInstance)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := dbs.MysqlDumpExec(replicationRequest, aws.StringValue(describeInstance.DBInstances[0].Endpoint.Address), serviceDBsSource, pathGlobal); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := dbs.MysqlRestore(replicationRequest, pathGlobal); err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := services.SetupReplication(replicationRequest, aws.StringValue(binLogFile), aws.StringValue(binLogPos)); err != nil {
		c.JSON(err.Status, err)
		return
	}

	c.JSON(http.StatusOK, "Transactional replication between source and taget has been setup. You now need to monitor with /seconds_behind_master route that `Seconds_Behind_Master` of mysql> show slave status; has reached zero after which you need to coordinate the microservice mysql switchover and call /promote_slave")
}
