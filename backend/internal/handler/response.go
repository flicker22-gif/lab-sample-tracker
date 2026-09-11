package handler

import "github.com/gin-gonic/gin"

func ok(data any) gin.H {
	return gin.H{"code": 0, "message": "ok", "data": data}
}

func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"code": status, "message": msg, "data": nil})
}
