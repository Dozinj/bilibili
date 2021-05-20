package tool

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"time"
)

var Client *redis.Client
func GoRedisConn(){
	client:=redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB: 0,  //默认设置索引为0的数据库
	})

	pong,err:=client.Ping().Result()
	if err!=nil{
		log.Fatal(err)
	}

	if pong!="PONG"{
		log.Fatal("客户端连接redis失败")
	}else{
		fmt.Println("客户端已连接服务器")
	}
	Client=client
}

func RedisSet (key,val string){
	result,err:=Client.Set(key,val,10*time.Minute).Result()//缓存时间为10分钟
	if err!=nil{
		fmt.Println("can not set val")
	}
	fmt.Println(result)
}


func RedisGet(key string)string{
	result,err:=Client.Get(key).Result()
	if err!=nil{
		fmt.Println("can not get val")
	}
	return result
}
