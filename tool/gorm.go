package tool

import (
	"bilibili/module"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var DBEngine *gorm.DB

func InitMysql()error{
	dsn:="root:123456@tcp(localhost:3306)/bilibili?charset=utf8&parseTime=True&loc=Local"
	db,err:=gorm.Open("mysql",dsn)
	if err!=nil{
		panic(err)
	}
	db.LogMode(true)

	DBEngine=db
	return db.DB().Ping()  //verifies database is still alive
}

func DBClose(){
	_ = DBEngine.Close()
}

func CreateTable(){
	DBEngine.AutoMigrate(&module.Follows{})
	DBEngine.AutoMigrate(&module.Video{})
	DBEngine.AutoMigrate(&module.User{})
	DBEngine.AutoMigrate(&module.VideoComment{})
	DBEngine.AutoMigrate(&module.UserWithVideo{})
	DBEngine.AutoMigrate(&module.Danmaku{})
	DBEngine.AutoMigrate(&module.SmsCode{})
}
