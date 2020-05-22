# devops-migrate-mysql-db
This tool perfoms the steps of [AWS Replication between Aurora clusters documentation](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Replication.MySQL.html).

Skim through both guides first to understand the process and actors involved.

### :white_check_mark: Prerequisites

- mysql 5.7 client 
- aws profile with RDS permissions:
    - RestoreCluster
    - DescribeClusters
    - CreateDBInstance
    - DescribeEvents
    - DeleteDBInstance
    - DeleteDBCluster

:rocket: The application starts by running: 

`go run main.go`

listens on localhost:8080 and exposes three routes which must be called in order of below appearance:

### bootstrap

`POST /database/bootstrap`

Takes care of all steps to setup replication between source and target clusters. ie: Set binlog retention, clone source cluster, dump/restore all databases from clone, bootstrap replication.

This call needs to persist some return objects to your local disk in `/tmp/migrated-db-name/` to be used in cleanup call later, so do not remove this directory until you have finished with the migration process.

Once you get a 200 move on to the next call.

### :exclamation: Gotchas
Dumping and restoring databases can take a long time depending on the size, and if during this time the call exits, for whatever reason, there is no way to start at where it left off. A manual cleanup, deleting temporary cluster, and beginning from the start is the only way in this version.

### Replication status

`POST /database/seconds_behind_master`

After cloning the cluster and depending on how long the dump/restore took, the master might have moved on substantially from the slave, and the later need to chatch up (reply all transacional log from master).
Call this endpoint as many times as needed to check on the replication status. 
At this stage at least in sandbox, a switchover of the application to the new mysql endpoint can take place, and readonly tests (because cluster is still a slave) can be executed. 

Once you get a 200 it's time to move on to the final call, but before you do read carefully..

### :sos: Promote slave (point of no return)

`POST /database/promote_slave` 

Calling this endpoind will break the replication, promote the slave to master, and drop the migrated database from source cluster (!so switching application back to source cluster is not an option). Switchover for the application, regarding mysql endpoint, must take place right after this call.

### Body payload

All three calls expect the below JSON body payload

```
{
	"source_db_name": "source-db-name",
	"source_db_user": "source-user",
	"source_db_password": "source-password",
	"source_db_host": "source-master-fqdn",
	"source_cluster_id": "source-cluster-name",
	"source_db_region": "source-cluster-region",
	"caller_aws_profile": "aws-profile",
	"replica_user_pass": "password-to-be-set-for-repl-user-in-source-cluster",
	"dest_db_user": "destination-user",
	"dest_db_password": "destination-password",
	"dest_db_host": "destination-master-fqdn"
}
```

### :no_entry: Steps not involved

Migrating user accounts. User accounts should best be created again in the target cluster.
Service accounts for the migrated database will be added in the next iteration.