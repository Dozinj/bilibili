package contorller

import (
	"bilibili/module"
	"bilibili/service"
	"bilibili/tool"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetSearch(c *gin.Context){
	KeyWords:=c.Query("keywords")

	if KeyWords==""{
		tool.PrintFalse(c,"搜索内容不可为空")
		return
	}
	var hs service.HomeService
	oldVideoSlice,err:=hs.Search(KeyWords)

	if err!=nil{
		fmt.Println(err)
		emptyVideo:=[]module.Video{}

		c.JSON(http.StatusOK,gin.H{
			"status":"true",
			"data":emptyVideo,
		})
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":*oldVideoSlice,
	})

}
