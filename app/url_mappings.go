package app

import "github.com/ArchieSpinos/migrate_rds_dbs/controllers/database"

func mapUrls() {
	router.POST("/database/setuprepl", database.SetupRepl)
	router.POST("/database/seconds_behind_master", database.SecondsBehindMaster)
	router.POST("/database/promote_slave", database.PromoteSlave)
}
