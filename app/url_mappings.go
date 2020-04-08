package app

import "github.com/ArchieSpinos/migrate_rds_dbs/controllers/database"

func mapUrls() {
	router.POST("/database/setuprepl", database.SetupRepl)
}
