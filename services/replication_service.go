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

func BootstrapReplication(request dbs.ReplicationRequest) *errors.DBErr {
	var (
		dropReplicaUser = "DROP USER 'repl_user'@'%';"
		replicaUser     = "CREATE USER 'repl_user'@'%' IDENTIFIED BY '" + request.ReplicaUserPass + "';"
		grantUser       = "GRANT REPLICATION CLIENT, REPLICATION SLAVE ON *.* TO 'repl_user'@'%';"
		setBinLog       = "CALL mysql.rds_set_configuration('binlog retention hours', 144);"
	)
	var queries = []string{
		dropReplicaUser,
		replicaUser,
		grantUser,
		setBinLog,
	}
	for _, query := range queries {
		result := &dbs.QueryResult{}
		if err := result.MultiQuery(request, query, true); err != nil {
			return err
		}
	}
	return nil
}

func SetupReplication(request dbs.ReplicationRequest, binLogFile string, binLogPos string) *errors.DBErr {
	var (
		setMaster        = "CALL mysql.rds_set_external_master ('" + request.SourceHost + "', 3306,'repl_user', '" + request.ReplicaUserPass + "', '" + binLogFile + "', " + binLogPos + ", 0);"
		startReplication = "CALL mysql.rds_start_replication;"
	)
	var queries = []string{
		setMaster,
		startReplication,
	}
	for _, query := range queries {
		result := &dbs.QueryResult{}
		if err := result.MultiQuery(request, query, false); err != nil {
			return err
		}
	}
	return nil
}

func CreateDestDatabase(request dbs.ReplicationRequest) (*dbs.QueryResult, *errors.DBErr) {
	var query = fmt.Sprintf("create database %s;", request.SourceDBName)
	result := &dbs.QueryResult{}
	if err := result.MultiQuery(request, query, false); err != nil {
		return nil, err
	}
	return result, nil
}

func removeSystemDBs(allDBs []string, systemDBs []string) (userDBs []string) {
	for _, v := range allDBs {
		for ks, vs := range systemDBs {
			if v == vs {
				break
			} else if ks < len(systemDBs)-1 {
				continue
			} else {
				userDBs = append(userDBs, v)
			}
		}
	}
	return userDBs
}

func PreFlightCheck(request dbs.ReplicationRequest) (serviceDBs []string, err *errors.DBErr) {
	// get source dbs excluding system, get target dbs excluding systemic, loop all source, if one exists in target fail and stop

	var (
		listQuery  = "show databases;"
		sourceDBs  = dbs.QueryResult{}
		destDBs    = dbs.QueryResult{}
		result     []string
		systemsDBs = []string{"mysql", "performance_schema", "information_schema", "sys"}
	)

	if err := sourceDBs.MultiQuery(request, listQuery, true); err != nil {
		return nil, err
	}
	if err := destDBs.MultiQuery(request, listQuery, false); err != nil {
		return nil, err
	}

	serviceDBsource := removeSystemDBs(sourceDBs, systemsDBs)
	serviceDBdest := removeSystemDBs(destDBs, systemsDBs)

	for _, sourceV := range serviceDBsource {
		for _, destV := range serviceDBdest {
			if sourceV == destV {
				result = append(result, sourceV)
			}
		}
	}
	if len(result) > 0 {
		return nil, errors.NewInternalServerError(fmt.Sprintf("The following source host databases exist in destination: %v. RDS transactional replication will migrate all databases so those existing in destination will be overwritten. Cannot continue", result))
	}
	return serviceDBsource, nil
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
		rdsSvc           = rds.New(awsSession)
		tries            = 0
		sourceType       = "db-instance"
		duration   int64 = 7200
		input            = rds.DescribeEventsInput{
			SourceIdentifier: instance.DBInstance.DBInstanceIdentifier,
			SourceType:       &sourceType,
			Duration:         &duration,
		}
	)
	for tries < 180 {
		events, describeErr := rdsSvc.DescribeEvents(&input)
		if describeErr != nil {
			return nil, nil, errors.NewBadRequestError(fmt.Sprintf("failed to describe RDS instance events: %s",
				describeErr.Error()))
		}
		for _, v := range events.Events {
			strMessage := aws.StringValue(v.Message)
			if strings.Contains(strMessage, "Binlog position from crash recovery") {
				s := strings.Fields(strMessage)
				return &s[len(s)-2], &s[len(s)-1], nil
			}
		}
		if tries < 180 {
			time.Sleep(5 * time.Second)
			tries++
		} else {
			return nil, nil, errors.NewBadRequestError(fmt.Sprintf("failed to retrieve binlog position and file"))
		}
	}
	return
}
