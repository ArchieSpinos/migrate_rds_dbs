package dbs

type DbConnection struct {
	Name     string `json:"db_name"`
	User     string `json:"db_user"`
	Password string `json:"db_password"`
	Host     string `json:"db_host"`
}

type ReplRequest struct {
	SourceName      string `json:"source_db_name"`
	SourceUser      string `json:"source_db_user"`
	SourcePassword  string `json:"source_db_password"`
	SourceHost      string `json:"source_db_host"`
	SourceClusterID string `json:"source_cluster_id"`
	DestName        string `json:"dest_db_name"`
	DestUser        string `json:"dest_db_user"`
	DestPassword    string `json:"dest_db_password"`
	DestHost        string `json:"dest_db_host"`
}

type QueryResult []string
