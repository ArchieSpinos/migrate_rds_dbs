package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

// BootstrapReplication boostraps transactional replication by creating
// replication user and setting binlog retention on source cluster.
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

// SetupReplication bootstraps transactional replication at target cluster.
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

// PreFlightCheck checks that none of the source cluster databases exist in the target
// because that would cause the transactional replication to override the target ones.
func PreFlightCheck(request dbs.ReplicationRequest, pathGlobal string) (serviceDBsSource []string, serviceDBsDest []string, err *errors.DBErr) {

	var (
		listQuery  = "show databases;"
		sourceDBs  = dbs.QueryResult{}
		destDBs    = dbs.QueryResult{}
		result     []string
		systemsDBs = []string{"mysql", "performance_schema", "information_schema", "sys"}
	)

	if _, err := os.Stat(pathGlobal); err == nil {
		return nil, nil, errors.NewInternalServerError(fmt.Sprintf("The dump path %s already exists. You need to delete it first.", pathGlobal))
	}

	if err := sourceDBs.MultiQuery(request, listQuery, true); err != nil {
		return nil, nil, err
	}
	if err := destDBs.MultiQuery(request, listQuery, false); err != nil {
		return nil, nil, err
	}

	serviceDBsSource = removeSystemDBs(sourceDBs, systemsDBs)
	serviceDBsDest = removeSystemDBs(destDBs, systemsDBs)

	for _, sourceV := range serviceDBsSource {
		for _, destV := range serviceDBsDest {
			if sourceV == destV {
				result = append(result, sourceV)
			}
		}
	}
	if len(result) > 0 {
		return nil, nil, errors.NewInternalServerError(fmt.Sprintf("The following source host databases exist in destination: %v. RDS transactional replication will migrate all databases so those existing in destination will be overwritten. Cannot continue", result))
	}
	return serviceDBsSource, serviceDBsDest, nil
}

// RDSDescribeToStruct creates an rds.RestoreDBClusterToPointInTimeInput that will be used to create
// a temp RDS cluster to dump databases from.
func RDSDescribeToStruct(replicationRequest dbs.ReplicationRequest, dscrOutput *rds.DescribeDBClustersOutput) rds.RestoreDBClusterToPointInTimeInput {
	var (
		DBClusterIdentifier     = "migrate-temp-" + replicationRequest.SourceDBName
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
		SourceDBClusterIdentifier:   &replicationRequest.SourceClusterID,
		VpcSecurityGroupIds:         VpcSecurityGroupIdsList,
		UseLatestRestorableTime:     &LatestRestorableTime,
	}
}

// RDSCreateInstanceToStruct creates an rds.RestoreDBClusterToPointInTimeInput that will be used to create
// a temp RDS instance to dump databases from.
func RDSCreateInstanceToStruct(restoredDB *rds.RestoreDBClusterToPointInTimeOutput, replicationRequest dbs.ReplicationRequest) rds.CreateDBInstanceInput {
	var (
		DBInstanceClassInput      = "db.r4.large"
		DBInstanceIdentifierInput = "migrate-temp-instance-" + replicationRequest.SourceDBName
		DBEngine                  = "aurora-mysql"
	)

	return rds.CreateDBInstanceInput{
		DBClusterIdentifier:  restoredDB.DBCluster.DBClusterIdentifier,
		DBInstanceClass:      &DBInstanceClassInput,
		DBInstanceIdentifier: &DBInstanceIdentifierInput,
		Engine:               &DBEngine,
	}
}

// RDSDescribeCluster retrieves information about the temp RDS cluster used to dump
// databases from.
func RDSDescribeCluster(awsSession *session.Session, replicationRequest dbs.ReplicationRequest) (*rds.DescribeDBClustersOutput, *errors.DBErr) {
	var (
		sourceClusterID = replicationRequest.SourceClusterID
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

func RDSDescribeInstance(awsSession *session.Session, instance rds.CreateDBInstanceOutput) *errors.DBErr {
	var (
		rdsSvc        = rds.New(awsSession)
		instanceInput = rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: instance.DBInstance.DBInstanceIdentifier,
		}
		tries = 0
	)
	for tries < 36 {
		rdsAddress, err := rdsSvc.DescribeDBInstances(&instanceInput)
		if err != nil {
			return errors.NewBadRequestError(fmt.Sprintf("failed to describe RDS instance: %s",
				err.Error()))
		} else if aws.StringValue(rdsAddress.DBInstances[0].Endpoint.Address) != "" {
			return nil
		} else if tries < 36 {
			time.Sleep(5 * time.Second)
			tries++
		} else {
			return errors.NewBadRequestError(fmt.Sprintf("failed to retrieve RDS instance fqdn"))
		}
	}
	return nil
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

func RDSWaitUntilInstanceAvailable(awsSession *session.Session, dbInstanceOutput *rds.CreateDBInstanceOutput) *errors.DBErr {
	var (
		rdsSvc = rds.New(awsSession)
		input  = rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: dbInstanceOutput.DBInstance.DBInstanceIdentifier,
		}
	)
	fmt.Println("in wait function")
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1800)*time.Second)
	if err := rdsSvc.WaitUntilDBInstanceAvailableWithContext(ctx, &input); err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("RDS instance did not become available in a timely manner: %s",
			err.Error()))
	}
	return nil
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
	for tries < 720 {
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
		if tries < 720 {
			time.Sleep(5 * time.Second)
			tries++
		} else {
			return nil, nil, errors.NewBadRequestError(fmt.Sprintf("failed to retrieve binlog position and file"))
		}
	}
	return
}
