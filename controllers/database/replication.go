package database

import (
	"fmt"
	"net/http"

	"github.com/ArchieSpinos/migrate_rds_dbs/awsresources"
	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/persist"
	"github.com/ArchieSpinos/migrate_rds_dbs/services"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/aws/aws-sdk-go/service/rds"
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

	serviceDBs, err := services.PreFlightCheck(replicationRequest)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	fmt.Println(serviceDBs) //delete after testing

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

	createInstanceInput := services.RDSCreateInstanceToStruct(restoredDB)

	rdsInstance, err := services.RDSCreateInstance(awsSession, createInstanceInput)
	if err != nil {
		c.JSON(err.Status, err)
		return
	}

	if err := persist.Save(pathGlobal, "rdsInstance", rdsInstance); err != nil {
		c.JSON(err.Status, err)
		return
	}

	obj := &rds.CreateDBInstanceOutput{}

	if err := persist.Load(pathGlobal+"rdsInstance", obj); err != nil {
		c.JSON(err.Status, err)
		return
	}

	fmt.Println(obj)

	//mock CreateDBInstanceOutput
	// var instance = "dev-migrate-temp-instance"
	// var address = "dev-migrate-temp-instance.cmnsml8q1eeo.eu-west-1.rds.amazonaws.com"
	// rdsInstance := &rds.CreateDBInstanceOutput{
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

	// if err := dbs.MysqlDumpExec(replicationRequest, aws.StringValue(rdsInstance.DBInstance.Endpoint.Address), serviceDBs); err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// if err := dbs.MysqlRestore(replicationRequest); err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	// if err := services.SetupReplication(replicationRequest, aws.StringValue(binLogFile), aws.StringValue(binLogPos)); err != nil {
	// 	c.JSON(err.Status, err)
	// 	return
	// }

	c.JSON(http.StatusOK, "Transactional replication between source and taget has been setup. You now need to monitor with /seconds_behind_master route that `Seconds_Behind_Master` of mysql> show slave status; has reached zero after which you need to coordinate the microservice mysql switchover and call /promote_slave")
}
