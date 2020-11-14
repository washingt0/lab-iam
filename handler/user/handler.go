package user

import (
	"lab/iam/application/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/washingt0/oops"
)

// Router performs all routing for user
func Router(r *gin.RouterGroup) {
	r.POST("", create)
	r.GET("about", get)
	r.GET("sessions", sessions)
}

func create(c *gin.Context) {
	var (
		in  user.User
		err error
		id  *string
	)

	if err = c.ShouldBindJSON(&in); err != nil {
		oops.GinHandleError(c, err, http.StatusBadRequest)
		return
	}

	if id, err = user.Create(c.Copy(), &in); err != nil {
		oops.GinHandleError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

func get(c *gin.Context) {}

func sessions(c *gin.Context) {}
