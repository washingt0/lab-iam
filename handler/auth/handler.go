package auth

import (
	"lab/iam/application/auth"
	"lab/iam/middleware"
	"lab/iam/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/washingt0/oops"
)

func Router(r *gin.RouterGroup) {
	r.POST("login", login)
	r.GET("session", middleware.ValidateJWT(), session)
	r.DELETE("session", middleware.ValidateJWT(), logout)
}

func login(c *gin.Context) {
	var (
		loginReq auth.Login
		err      error
		token    *string
	)

	c.Set("UID", "public")

	if err = c.ShouldBindJSON(&loginReq); err != nil {
		oops.GinHandleError(c, err, http.StatusBadRequest)
		return
	}

	loginReq.UserAgent = utils.GetStringPointer(c.GetHeader("User-Agent"))
	loginReq.IPAddress = utils.GetStringPointer(c.ClientIP())

	if loginReq.IPAddress == nil {
		loginReq.IPAddress = utils.GetStringPointer("0.0.0.0")
	}

	loginReq.IPLocation = utils.GetStringPointer("unknown")

	if token, err = auth.TryLogin(c.Copy(), &loginReq); err != nil {
		oops.GinHandleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"kind":  "Bearer",
	})
}

func logout(c *gin.Context) {}

func session(c *gin.Context) {
	var (
		sess *middleware.Session
		ok   bool
	)

	if sess, ok = c.Value("USession").(*middleware.Session); !ok {
		oops.GinHandleError(c, oops.ThrowError("Invalid Session", nil), http.StatusForbidden)
		return
	}

	c.JSON(http.StatusOK, sess)
}
