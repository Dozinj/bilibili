package tool

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Sms SmsConfig `json:"sms"`
	Oss OssConfig `json:"oss"`

}
type SmsConfig struct {
	SignName   string  `json:"sign_name"`
	TemplateCode   string   `json:"template_code"`
	AppKey   string  `json:"app_key"`
	AppSecret   string `json:"app_secret"`
	RegionId   string   `json:"region_id"`
}

type OssConfig struct {
	EndPoint string `json:"end_point"`
	AppKey   string  `json:"app_key"`
	AppSecret   string `json:"app_secret"`
	AvatarBucket string `json:"avatar_bucket"`
	AvatarUrl string `json:"avatar_url"`
	VideosBucket string `json:"videos_bucket"`
	VideosUrl  string `json:"videos_url"`
}

var _cfg *Config

func GetCfg()*Config{
	return _cfg
}

func ParseConfig(path string)(*Config,error){
	file,err:=os.Open(path)
	if err!=nil{
		log.Fatalln(err)
	}

	defer file.Close()
	err= json.NewDecoder(bufio.NewReader(file)).Decode(&_cfg)
	if err!=nil{
		log.Fatalln(err)
	}
	return _cfg,err
}