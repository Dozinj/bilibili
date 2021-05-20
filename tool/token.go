package tool

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}
type Payload struct {
	Iss string `json:"iss"`
	Exp string `json:"exp"`
	Iat string `json:"iat"`
	Id int `json:"id"`
}

type JWT struct {
	Header    Header
	Payload   Payload
	Signature string
	Token     string
}
var Key string="1671416"
//新建token
func NewJWT(uid int,TokenType string) JWT {
	var jwt JWT
	jwt.Header= Header{
		Alg: "HS256",
		Typ: TokenType,
	}

	jwt.Payload = Payload{
		Iss: "bilibili",
		Iat: strconv.FormatInt(time.Now().Unix(), 10),
		Id:  uid,
	}
	//对token类型进行判断
	if TokenType=="token" {
		jwt.Payload.Exp=strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)
	}else if TokenType=="refreshToken"{
		jwt.Payload.Exp=strconv.FormatInt(time.Now().Add(1*time.Hour).Unix(), 10)
	}

	//序列化再Base64加密
	h,_:=json.Marshal(jwt.Header)
	p,_:=json.Marshal(jwt.Payload)

	baseh:=base64.StdEncoding.EncodeToString(h)
	basep:=base64.StdEncoding.EncodeToString(p)

	secret:=baseh+"."+basep
	mac:=hmac.New(sha256.New,[]byte(Key))
	s:=mac.Sum(nil)
	jwt.Signature=base64.StdEncoding.EncodeToString(s)

	jwt.Token=secret+"."+jwt.Signature
	return jwt
}

//check token
func Check(token,key string)(jwt JWT,err error) {
	err = errors.New("token err")
	arr := strings.Split(token, ".")
	if len(arr) < 3 {
		fmt.Println(err)
		return
	}
	baseh := arr[0]
	//base64解密再反序列化
	h, err1 := base64.StdEncoding.DecodeString(baseh)
	if err1 != nil {
		fmt.Println("decode err", err1)
		return
	}

	//token header->jwt.Header
	err = json.Unmarshal(h, &jwt.Header)
	if err != nil {
		fmt.Println("json Unmarshal err:", err)
		return
	}

	basep := arr[1]
	//base64解密再反序列化
	p, err2 := base64.StdEncoding.DecodeString(basep)
	if err2 != nil {
		fmt.Println("decode err", err2)
		return
	}

	err = json.Unmarshal(p, &jwt.Payload)
	if err != nil {
		fmt.Println("json Unmarshal err:", err)
		return
	}

	mac := hmac.New(sha256.New, []byte(key))
	s := mac.Sum(nil)
	serverS := base64.StdEncoding.EncodeToString(s)
	clientS := arr[2]
	if serverS != clientS {
		fmt.Println("token signature err")
		err = errors.New("token signature err")
		return
	}
	jwt.Token = token
	jwt.Signature =clientS
	return jwt,nil
}

//解析Token
func ParseToken(token string,c *gin.Context)(jwt JWT,err error){
	//先验证token
	if token==""{
		err=errors.New("NO_TOKEN_PROVIDED")
		PrintFalse(c,"NO_TOKEN_PROVIDED")
		return
	}

	jwt,err=Check(token,Key)

	if err!=nil||jwt.Header.Typ!="token"{
		err=errors.New("PRASE_TOKEN_ERROR")
		PrintFalse(c,"PRASE_TOKEN_ERROR")
		return
	}

	//判断是否为最新的令牌
	tokenIat:=RedisGet(strconv.Itoa(jwt.Payload.Id))
	if jwt.Payload.Iat!=tokenIat{
		err=errors.New("PRASE_TOKEN_ERROR")
		PrintFalse(c,"PRASE_TOKEN_ERROR")
		return
	}

	exptime,_:=strconv.ParseInt(jwt.Payload.Exp,10,64)
	nowtime:=time.Now().Unix()
	if nowtime>exptime{
		PrintFalse(c,"TOKEN_EXPIRED")
		err=errors.New("TOKEN_EXPIRED")
		return
	}
	return
}


//解析refreshToken
func ParseRefreshToken(token string,c *gin.Context)(jwt JWT,err error){
	//先验证token
	if token==""{
		err=errors.New("NO_REFRESHTOKEN_PROVIDED")
		PrintFalse(c,"refreshToken 为空")
		return
	}

	jwt,err=Check(token,Key)
	if err!=nil||jwt.Header.Typ!="refreshToken"{
		err=errors.New("refreshToken不正确或系统错误")
		PrintFalse(c,"refreshToken不正确或系统错误")
		return
	}

	exptime,_:=strconv.ParseInt(jwt.Payload.Exp,10,64)
	nowtime:=time.Now().Unix()
	if nowtime>exptime{
		PrintFalse(c,"refreshToken失效")
		err=errors.New("TOKEN_EXPIRED")
		return
	}
	return
}

