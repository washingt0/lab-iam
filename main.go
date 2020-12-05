package main

import (
	"lab/iam/config"
	"lab/iam/database"
	"lab/iam/handler/auth"
	"lab/iam/handler/user"
	"lab/iam/logger"
	"lab/iam/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2"
)

func main() {
	var (
		err error
	)

	if err = logger.Setup(); err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	if err = database.OpenDatabases(); err != nil {
		zap.L().Fatal("error while opening connections", zap.Error(err))
	}
	defer database.Close()

	if err = httpRouter().Run(config.GetConfig().BindAddress); err != nil {
		zap.L().Fatal("error while serving application", zap.Error(err))
	}
}

func httpRouter() (r *gin.Engine) {
	r = gin.New()

	r.Use(
		middleware.RequestIdentifier(),
		middleware.RequestLogger(),
	)

	r.GET("/keys", listKeys)

	v1 := r.Group("v1")

	auth.Router(v1.Group("auth"))

	v1.Use(middleware.ValidateJWT())

	user.Router(v1.Group("user"))

	return
}

func listKeys(c *gin.Context) {
	var (
		keys = config.GetConfig().JWT.Keys
		out  = make([]jose.JSONWebKey, 0, len(keys))
	)

	for i := range keys {
		out = append(out, jose.JSONWebKey{
			Key:       keys[i].PublicKey,
			KeyID:     keys[i].ID,
			Algorithm: "RS256",
			Use:       "sig",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"keys": out,
	})
}
