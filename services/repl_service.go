package services

import (
	"fmt"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

func EnableBinLogRetention(dbconn dbs.ReplRequest) (*dbs.QueryResult, *errors.DBErr) {
	//var queries = []string{"CALL mysql.rds_set_configuration('binlog retention hours', 144);", "CALL mysql.rds_show_configuration;"}
	//var query = "CALL mysql.rds_set_configuration('binlog retention hours', 144);"
	var query = "show databases;"
	//var query = "CALL mysql.rds_show_configuration;"
	result := &dbs.QueryResult{}
	if err := result.MultiQueryLogRetention(dbconn, query); err != nil {
		return nil, err
	}
	return result, nil
}

func RDSDescribeToStruct(replReq dbs.ReplRequest, dscrOutput *rds.DescribeDBClustersOutput) rds.RestoreDBClusterToPointInTimeInput {
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

func RDSDescribeCluster(awsSession *session.Session, replreq dbs.ReplRequest) (*rds.DescribeDBClustersOutput, *errors.DBErr) {
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
