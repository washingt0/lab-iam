package main

import (
	"lab/iam/config"
	"lab/iam/database"
	"lab/iam/database/types"
	"lab/iam/logger"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/washingt0/oops"
	"go.uber.org/zap"
)

func main() {
	var (
		err error
	)

	if err = database.OpenDatabases(); err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err = logger.Setup(); err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	if err = httpRouter().Run(config.GetConfig().BindAddress); err != nil {
		zap.L().Error("error while serving application", zap.Error(err))
	}
}

func httpRouter() (r *gin.Engine) {
	r = gin.New()

	r.GET("/", func(c *gin.Context) {
		var (
			tx  types.Transaction
			err error
		)
		c.Set("RID", uuid.New().String())
		c.Set("UID", uuid.New().String())

		if tx, err = database.NewTx(c.Copy(), false); err != nil {
			oops.GinHandleError(c, err, http.StatusBadRequest)
			return
		}

		if _, err = tx.Exec("INSERT INTO t_outgoing_message(queue, payload) VALUES ('user_creation', '{}')"); err != nil {
			oops.GinHandleError(c, err, http.StatusBadRequest)
			return
		}

		if _, err = tx.Exec("UPDATE t_outgoing_message SET sent_at = NOW() WHERE sent_at IS NULL"); err != nil {
			oops.GinHandleError(c, err, http.StatusBadRequest)
			return
		}

		if err = tx.Commit(); err != nil {
			oops.GinHandleError(c, err, http.StatusBadRequest)
			return
		}

		c.JSON(http.StatusOK, gin.H{"top": true})

	})

	return
}
