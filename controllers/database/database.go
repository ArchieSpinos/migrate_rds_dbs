package database

import (
	"net/http"

	"github.com/ArchieSpinos/migrate_rds_dbs/domain/dbs"
	"github.com/ArchieSpinos/migrate_rds_dbs/services"
	"github.com/ArchieSpinos/migrate_rds_dbs/utils/errors"
	"github.com/gin-gonic/gin"
)

func SetupRepl(c *gin.Context) {
	var dbcon dbs.ReplRequest
	if err := c.ShouldBindJSON(&dbcon); err != nil {
		dbErr := errors.NewBadRequestError("invalid json body")
		c.JSON(dbErr.Status, dbErr)
		return
	}
	binLogRetentionResult, listErr := services.EnableBinLogRetention(dbcon)
	if listErr != nil {
		c.JSON(listErr.Status, listErr)
		return
	}
	c.JSON(http.StatusOK, binLogRetentionResult)
}
