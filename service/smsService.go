package service

import (
	"bilibili/tool"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"log"
	"math/rand"
	"time"
)
type SmsService struct {

}

//发送验证码
func (ss SmsService)SendCode(phone string) bool{
	cfg:=tool.GetCfg().Sms
	//1.产生一个验证码
	code:=fmt.Sprintf("%06v",rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))

	//2.调用阿里云sdk，完成调用
	client,err:=dysmsapi.NewClientWithAccessKey(cfg.RegionId,cfg.AppKey,cfg.AppSecret)
	if err!=nil{
		log.Fatal(err)
		return false
	}

	request := dysmsapi.CreateSendSmsRequest() //获取请求结构体
	request.Scheme = "https"  //请求类型
	request.SignName=cfg.SignName
	request.TemplateCode=cfg.TemplateCode
	request.PhoneNumbers=phone

	//将结构体转化为json字符串
	par, err := json.Marshal(map[string]interface{}{
		"code": code,
	})
	request.TemplateParam = string(par)

	response,err:=client.SendSms(request)
	fmt.Println(response)
	if err!=nil{
		log.Fatal(err)
		return false
	}

	//3.接收返回结果并判断返回状态
	//短信验证码发送成功
	if response.Code=="OK"{
		//将验证码保存到redis数据库中
		tool.RedisSet(phone,code)
	}
	return true
}

