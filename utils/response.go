package utils

import "github.com/gin-gonic/gin"

func ResponseFormat(c *gin.Context, code int, msg string) {
	c.JSON(200, gin.H{
		"code": code,
		"msg":  msg,
	})
	return
}
