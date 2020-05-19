package services

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

func SlicePointerToSlice(input *[]string) []string {
	output := append([]string{}, *input...)
	return output
}

// PromoteSlave promotes slave mysql node to master and stops replication.
func PromoteSlave(request dbs.ReplicationRequest) *errors.DBErr {
	var (
		stopReplication = "CALL mysql.rds_stop_replication;"
		resetMaster     = "CALL mysql.rds_reset_external_master;"
	)
	var queries = []string{
		stopReplication,
		resetMaster,
	}
	for _, query := range queries {
		result := &dbs.QueryResult{}
		if err := result.MultiQuery(request, query, false); err != nil {
			return err
		}
	}
	return nil
}

// RDSDeleteInstance deletes temporary RDS instance used to dump databases.
func RDSDeleteInstance(awsSession *session.Session, input rds.DescribeDBInstancesOutput) *errors.DBErr {
	var (
		rdsSvc              = rds.New(awsSession)
		deleteInstanceInput = rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: input.DBInstances[0].DBInstanceIdentifier,
			SkipFinalSnapshot:    aws.Bool(true),
		}
	)

	if _, err := rdsSvc.DeleteDBInstance(&deleteInstanceInput); err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to delete DB instance: %s",
			err.Error()))
	}
	return nil
}

// CleanUpDBs drops databases from target cluster which were created as part of
// transactional replication (since RDS does not allow to choose databases when
// creating replication slaves). It also drops migrated database from source cluster
// and resets binlog retention to NULL from source cluster.
func CleanUpDBs(request dbs.ReplicationRequest, serviceDBsDest []string) *errors.DBErr {
	var (
		listQuery              = "show databases;"
		resetLogRetentionQuery = "CALL mysql.rds_set_configuration('binlog retention hours', NULL);"
		dropDBQuery            string
		allDestDBs             = dbs.QueryResult{}
		dropMigratedDB         = dbs.QueryResult{}
		resetLog               = dbs.QueryResult{}
		systemsDBs             = []string{"mysql", "performance_schema", "information_schema", "sys"}
		tobeRemovedDBs         []string
	)
	serviceDBsDest = append(serviceDBsDest, request.SourceDBName)
	for _, v := range systemsDBs {
		serviceDBsDest = append(serviceDBsDest, v)
	}

	if err := allDestDBs.MultiQuery(request, listQuery, false); err != nil {
		return err
	}

	for _, v := range allDestDBs {
		for i, k := range serviceDBsDest {
			if v == k {
				break
			} else if (v != k) && (i < len(serviceDBsDest)-1) {
				continue
			} else {
				tobeRemovedDBs = append(tobeRemovedDBs)
			}
		}
	}

	for _, db := range tobeRemovedDBs {
		result := &dbs.QueryResult{}
		dropDBQuery = "drop database" + db + ";"
		if err := result.MultiQuery(request, dropDBQuery, false); err != nil {
			return err
		}
	}

	if err := dropMigratedDB.MultiQuery(request, "drop database "+request.SourceDBName+";", true); err != nil {
		return err
	}

	if err := resetLog.MultiQuery(request, resetLogRetentionQuery, true); err != nil {
		return err
	}
	return nil
}
