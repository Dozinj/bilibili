package main

import (
	"bilibili/routers"
	"bilibili/tool"
	"log"
)

func main() {
	//解析配置文件
	_,err:=tool.ParseConfig("./config/app.json")
	if err!=nil{
		panic(err)
	}

	err= tool.InitMysql()
	if err!=nil{
		panic(err)
	}
	tool.CreateTable()
	defer tool.DBClose()
	tool.GoRedisConn()//连接redis

	log.SetFlags(log.Lshortfile|log.Ltime)
	r:=routers.RegisterRouter()
	_ = r.Run()
}




