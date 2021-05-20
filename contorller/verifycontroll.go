package contorller

import (
	"bilibili/tool"
	"fmt"
	"github.com/gin-gonic/gin"
	"bilibili/dao"
	"bilibili/module"
	"net/http"
	"regexp"
	"bilibili/service"
	"strconv"
	"strings"
)
var ud dao.Userdao

var ss service.SmsService
func RefreshToken(c *gin.Context){
	refreshToken:=c.Query("refreshtoken")
	refreshTokenJwt,err:=tool.ParseRefreshToken(refreshToken,c)
	if err!=nil{
		return
	}
	////判断是否为最新的鉴权令牌
	newuser:=new(module.User)
	newuser.Uid= refreshTokenJwt.Payload.Id
	olduser,err:=us.FindUser(newuser)

	if err!=nil||olduser.RefreshTokenIat!=refreshTokenJwt.Payload.Iat{
		tool.PrintFalse(c,"refreshToken不正确或系统错误")
		return
	}


	//refreshToken 解析成功------生成新的token
	tokenJwt:=tool.NewJWT(refreshTokenJwt.Payload.Id,"token")
	if tokenJwt.Header.Typ!="token" {
		fmt.Println("服务器生成token错误")
		return
	}

	//将token写入redis
	tool.RedisSet(strconv.Itoa(tokenJwt.Payload.Id),tokenJwt.Payload.Iat)

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":tokenJwt.Token,
	})
}




func SmsHandle(c *gin.Context){
	phone:=c.PostForm("phone")
	if phone==""{
		tool.PrintFalse(c,"手机号不可为空")
		return
	}else if len(phone)!=11{
		tool.PrintFalse(c,"手机号不合法")
		return
	}

	//判断这11个字符是否都是0-9之间的数字
	regl:=regexp.MustCompile(`^[0-9]\d*$`)
	if regl==nil{
		fmt.Println("regexp.MustCompile err")
	}
	result:=regl.FindAllStringSubmatch(phone,-1)
	if result==nil{
		tool.PrintFalse(c,"手机号不合法")
		return
	}

	//解析Url
	url:=c.Request.URL.Path
	str:=strings.Split(url,"/")
	path:=str[len(str)-1]//获取末端路径

	newuser:=new(module.User)
	newuser.Phone=phone
	_,err:=us.FindUser(newuser)


	//目前仅有三个接口
	pathslice:="register login general"
	if !strings.Contains(pathslice,path){
		return
	}
	//检测手机号是否存在
	if path=="register"&&err==nil{
		tool.PrintFalse(c,"手机号已被使用")
		return
	}

	if path=="login"&&err!=nil{
		tool.PrintFalse(c,"手机号未被注册")
		return
	}

	//发送验证码
	flag:=ss.SendCode(phone)
	if !flag{
		fmt.Println("服务器发送短信失败 ")
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"",
	})
}
