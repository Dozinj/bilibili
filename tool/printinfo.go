package tool

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func PrintFalse(c *gin.Context,data string){
	c.JSON(http.StatusOK,gin.H{
		"status":"false",
		"data":data,
	})
}

