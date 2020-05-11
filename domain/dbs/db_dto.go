package dbs

type DbConnection struct {
	Name     string `json:"db_name"`
	User     string `json:"db_user"`
	Password string `json:"db_password"`
	Host     string `json:"db_host"`
}

type ReplicationRequest struct {
	SourceUser      string `json:"source_db_user"`
	SourcePassword  string `json:"source_db_password"`
	SourceHost      string `json:"source_db_host"`
	SourceClusterID string `json:"source_cluster_id"`
	SourceDBName    string `json:"source_db_name"`
	MysqlDumpPath   string `json:"mysql_dump_path"`
	ReplicaUserPass string `json:"replica_user_pass"`
	DestName        string `json:"dest_db_name"`
	DestUser        string `json:"dest_db_user"`
	DestPassword    string `json:"dest_db_password"`
	DestHost        string `json:"dest_db_host"`
}

type QueryResult []string

type MysqlShow struct {
	SlaveIOState              string `json:"Slave_IO_State"`
	MasterHost                string `json:"Master_Host"`
	MasterUser                string `json:"Master_User"`
	MasterPort                int    `json:"Master_Port"`
	ConnectRetry              int    `json:"Connect_Retry"`
	MasterLogFile             string `json:"Master_Log_File"`
	ReadMasterLogPos          int    `json:"Read_Master_Log_Pos"`
	RelayLogFile              string `json:"Relay_Log_File"`
	RelayLogPos               int    `json:"Relay_Log_Pos"`
	RelayMasterLogFile        string `json:"Relay_Master_Log_File"`
	SlaveIORunning            string `json:"Slave_IO_Running"`
	SlaveSQLRunning           string `json:"Slave_SQL_Running"`
	ReplicateDoDB             string `json:"Replicate_Do_DB"`
	ReplicateIgnoreDB         string `json:"Replicate_Ignore_DB"`
	ReplicateDoTable          string `json:"Replicate_Do_Table"`
	ReplicateIgnoreTable      string `json:"Replicate_Ignore_Table"`
	ReplicateWildDoTable      string `json:"Replicate_Wild_Do_Table"`
	ReplicateWildIgnoreTable  string `json:"Replicate_Wild_Ignore_Table"`
	LastErrno                 int    `json:"Last_Errno"`
	LastError                 string `json:"Last_Error"`
	SkipCounter               int    `json:"Skip_Counter"`
	ExecMasterLogPos          int    `json:"Exec_Master_Log_Pos"`
	RelayLogSpace             int    `json:"Relay_Log_Space"`
	UntilCondition            string `json:"Until_Condition"`
	UntilLogFile              string `json:"Until_Log_File"`
	UntilLogPos               int    `json:"Until_Log_Pos"`
	MasterSSLAllowed          string `json:"Master_SSL_Allowed"`
	MasterSSLCAFile           string `json:"Master_SSL_CA_File"`
	MasterSSLCAPaths          string `json:"Master_SSL_CA_Path"`
	MasterSSLCert             string `json:"Master_SSL_Cert"`
	MasterSSLCipher           string `json:"Master_SSL_Cipher"`
	MasterSSLKey              string `json:"Master_SSL_Key"`
	SecondsBehindMaster       int    `json:"Seconds_Behind_Master"`
	MasterSSLVerifyServerCert string `json:"Master_SSL_Verify_Server_Cert"`
	LastIOErrno               int    `json:"Last_IO_Errno"`
	LastIOError               string `json:"Last_IO_Error"`
	LastSQLErrno              int    `json:"Last_SQL_Errno"`
	LastSQLError              string `json:"Last_SQL_Error"`
	ReplicateIgnoreServerIds  string `json:"Replicate_Ignore_Server_Ids"`
	MasterServerId            int    `json:"Master_Server_Id"`
	MasterUUID                string `json:"Master_UUID"`
	MasterInfoFile            string `json:"Master_Info_File"`
	SQLDelay                  int    `json:"SQL_Delay"`
	SQLRemainingDelay         string `json:"SQL_Remaining_Delay"`
	SlaveSQLRunningState      string `json:"Slave_SQL_Running_State"`
	MasterRetryCount          int    `json:"Master_Retry_Count"`
	MasterBind                string `json:"Master_Bind"`
	LastIOErrorTimestamp      string `json:"Last_IO_Error_Timestamp"`
	LastSQLErrorTimestamp     string `json:"Last_SQL_Error_Timestamp"`
	MasterSSLCrl              string `json:"Master_SSL_Crl"`
	MasterSSLCrlpath          string `json:"Master_SSL_Crlpath"`
	RetrievedGtidSet          string `json:"Retrieved_Gtid_Set"`
	ExecutedGtidSet           string `json:"Executed_Gtid_Set"`
	AutoPosition              int    `json:"Auto_Position"`
	ReplicateRewriteDB        string `json:"Replicate_Rewrite_DB"`
	ChannelName               string `json:"Channel_Name"`
	MasterTLSVersion          string `json:"Master_TLS_Version"`
}
