package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

func EnableBinLogRetention(request dbs.ReplicationRequest) (*dbs.QueryResult, *errors.DBErr) {
	//var queries = []string{"CALL mysql.rds_set_configuration('binlog retention hours', 144);", "CALL mysql.rds_show_configuration;"}
	//var query = "CALL mysql.rds_set_configuration('binlog retention hours', 144);"
	var query = "show databases;"
	//var query = "CALL mysql.rds_show_configuration;"
	result := &dbs.QueryResult{}
	if err := result.MultiQuery(request, query); err != nil {
		return nil, err
	}
	return result, nil
}

func CreateDestDatabase(request dbs.ReplicationRequest) (*dbs.QueryResult, *errors.DBErr) {
	var query = fmt.Sprintf("create database %s;", request.SourceDBName)
	result := &dbs.QueryResult{}
	if err := result.MultiQuery(request, query); err != nil {
		return nil, err
	}
	return result, nil
}

func RDSDescribeToStruct(replReq dbs.ReplicationRequest, dscrOutput *rds.DescribeDBClustersOutput) rds.RestoreDBClusterToPointInTimeInput {
	var (
		DBClusterIdentifier     = "dev-migrate-temp" // to do build string from source cluster id
		VpcSecurityGroupIdsList []*string
		LatestRestorableTime    = true
	)
	for _, element := range dscrOutput.DBClusters[0].VpcSecurityGroups {
		VpcSecurityGroupIdsList = append(VpcSecurityGroupIdsList, element.VpcSecurityGroupId)
	}

	return rds.RestoreDBClusterToPointInTimeInput{
		DBClusterIdentifier:         &DBClusterIdentifier,
		DBClusterParameterGroupName: dscrOutput.DBClusters[0].DBClusterParameterGroup,
		DBSubnetGroupName:           dscrOutput.DBClusters[0].DBSubnetGroup,
		SourceDBClusterIdentifier:   &replReq.SourceClusterID,
		VpcSecurityGroupIds:         VpcSecurityGroupIdsList,
		UseLatestRestorableTime:     &LatestRestorableTime,
	}
}

func RDSCreateInstanceToStruct(restoredDB *rds.RestoreDBClusterToPointInTimeOutput) rds.CreateDBInstanceInput {
	var (
		DBInstanceClassInput      = "db.t2.small"
		DBInstanceIdentifierInput = "dev-migrate-temp-instance"
		DBEngine                  = "aurora-mysql"
	)

	return rds.CreateDBInstanceInput{
		DBClusterIdentifier:  restoredDB.DBCluster.DBClusterIdentifier,
		DBInstanceClass:      &DBInstanceClassInput,
		DBInstanceIdentifier: &DBInstanceIdentifierInput,
		Engine:               &DBEngine,
	}
}

func RDSDescribeCluster(awsSession *session.Session, replreq dbs.ReplicationRequest) (*rds.DescribeDBClustersOutput, *errors.DBErr) {
	var (
		sourceClusterID = replreq.SourceClusterID
		rdsSvc          = rds.New(awsSession)
		clusterInput    = rds.DescribeDBClustersInput{
			DBClusterIdentifier: &sourceClusterID,
		}
	)
	DBClusterOutput, err := rdsSvc.DescribeDBClusters(&clusterInput)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to describe RDS Cluster: %s",
			err.Error()))
	}
	return DBClusterOutput, nil
}

func RDSRestoreCluster(awsSession *session.Session, input rds.RestoreDBClusterToPointInTimeInput) (*rds.RestoreDBClusterToPointInTimeOutput, *errors.DBErr) {
	var rdsSvc = rds.New(awsSession)
	DBClusterOutput, err := rdsSvc.RestoreDBClusterToPointInTime(&input)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to restore RDS Cluster: %s",
			err.Error()))
	}
	return DBClusterOutput, nil
}

func RDSCreateInstance(awsSession *session.Session, input rds.CreateDBInstanceInput) (*rds.CreateDBInstanceOutput, *errors.DBErr) {
	var rdsSvc = rds.New(awsSession)
	DBInstanceOutput, err := rdsSvc.CreateDBInstance(&input)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to create DB instance: %s",
			err.Error()))
	}
	return DBInstanceOutput, nil
}

func RDSDescribeEvents(awsSession *session.Session, instance *rds.CreateDBInstanceOutput) (binLogFile *string, binLogPos *string, err *errors.DBErr) {
	var (
		rdsSvc = rds.New(awsSession)
		tries  = 0
		input  = rds.DescribeEventsInput{
			SourceIdentifier: instance.DBInstance.DBInstanceIdentifier,
		}
	)
	for tries < 20 {
		events, err := rdsSvc.DescribeEvents(&input)
		for _, v := range events.Events {
			strMessage := aws.StringValue(v.Message)
			fmt.Println(v.Message)
			if strings.Contains(strMessage, "Binlog position from crash recovery") {
				s := strings.Fields(strMessage)
				return &s[len(s)-2], &s[len(s)-1], nil
			} else if tries < 20 {
				time.Sleep(5 * time.Second)
				tries++
			} else {
				return nil, nil, errors.NewBadRequestError(fmt.Sprintf("failed to create DB instance: %s",
					err.Error()))
			}
		}
	}
	return
}
