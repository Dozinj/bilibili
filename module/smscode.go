package module

//存放用户手机号，验证码等
type SmsCode struct {
	//biz_id 阿里云发送短信的业务代码结构
	Id int64 `json:"id"`
	Phone string `json:"phone"`
	BizId string `json:"biz_id"`
	Code string `json:"code"`
	CreateTime int64 `json:"create_time"`
}
