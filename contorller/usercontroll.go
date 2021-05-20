package contorller

import (
	"bilibili/module"
	param2 "bilibili/param"
	"bilibili/service"
	"bilibili/tool"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)


var us service.UserService
//用户等录
func Loginpw(c *gin.Context) {
	loginName := c.PostForm("loginName")
	password := c.PostForm("password")

	if loginName == "" {
		tool.PrintFalse(c, "请输入注册时用的邮箱或者手机号呀")
		return
	} else if password == "" {
		tool.PrintFalse(c, "喵，你没输入密码么？")
		return
	}

	newuser:=new(module.User)
	newuser.Name=loginName  //通过name 查询

	olduser, err := us.FindUser(newuser)
	if err != nil ||!us.Decode_md5(password,olduser){ //解密成功返回true
		tool.PrintFalse(c, "用户名或密码错误")
		return
	}

	//创建token
	tokenJwt := tool.NewJWT(olduser.Uid,"token")
	refreshTokenJwt:= tool.NewJWT(olduser.Uid,"refreshToken")

	//在redis中记录颁发最近token时间  使得只有最新的token才有效
	//key 为唯一标识符Uid  val为生成时间
	tool.RedisSet(strconv.Itoa(tokenJwt.Payload.Id),tokenJwt.Payload.Iat)

	//refreshToken 过期时间较长可以存入Mysql数据库
	newuser.RefreshTokenIat=refreshTokenJwt.Payload.Iat
	err=us.ChangeUser(olduser,newuser)
	if err!=nil{
		fmt.Println("refresh token iat failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":       "true",
		"data":         "",
		"token":        tokenJwt.Token,
		"uid":          olduser.Uid,
		"refreshToken": refreshTokenJwt.Token,
	})
}

//注册用户
func Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	phone := c.PostForm("phone")
	verify_code := c.PostForm("verify_code")

	if username == "" {
		tool.PrintFalse(c, "用户名不能为空")
		return
	} else if len(username) > 15 {
		tool.PrintFalse(c, "用户名太长了")
		return
	}

	if len(password) < 6 {
		tool.PrintFalse(c, "密码不能小于6个字符")
		return
	} else if len(password) > 16 {
		tool.PrintFalse(c, "密码不能大于16个字符")
		return
	}

	//查看user是否存在
	newuser:=new(module.User)
	newuser.Name=username

	user, _ := us.FindUser(newuser)
	if phone == "" {
		tool.PrintFalse(c, "手机号不可为空")
		return
	} else if user != nil && user.Phone == phone {
		tool.PrintFalse(c, "该手机号已经被注测")
		return
	}

	if verify_code == "" {
		tool.PrintFalse(c, "请输入验证码")
		return
	}

	//在redis中查找,并验证验证码
	code:=tool.RedisGet(phone)
	if code != verify_code {
		tool.PrintFalse(c, "验证码错误")
	}else {

		//md5 密码加密
		pwdSalt,pwdHash:=us.Encrypt_md5(password)
		//注册成功添加表信息
		user1 := module.User{
			Name:     username,
			PwdHash: pwdHash,
			PwdSalt: pwdSalt,
			Phone:    phone,
			RegDate: time.Unix(time.Now().Unix(),0).Format("2006-01-02 15:04:05"),
			Coins: 6,
			Exp: 0,
			LastViewTime: time.Now(),
			LastCoinTime: time.Now(),
		}
		err:=us.CreateUser(&user1)
		if err!=nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "true",
			"data":   "注册成功",
		})
	}
}

//短信登录
func Loginsms(c *gin.Context) {
	phone := c.PostForm("phone")
	verify_code := c.PostForm("verify_code")

	if phone == "" {
		tool.PrintFalse(c, "手机号不能为空哦")
		return
	} else if verify_code == "" {
		tool.PrintFalse(c, "短信验证码不能为空")
		return
	}
	//获取uid+检测用户是否注册
	newuser := new(module.User)
	newuser.Phone = phone
	olduser, err := us.FindUser(newuser)
	if err != nil {
		tool.PrintFalse(c, "找不到该用户")
		return
	}

	//从redis中读取验证码
	code := tool.RedisGet(phone)

	if code != verify_code {
		tool.PrintFalse(c, "验证码错误")
	} else {
		//登录成功创建Token
		//创建token
		tokenJwt := tool.NewJWT(olduser.Uid, "token")
		refreshTokenJwt := tool.NewJWT(olduser.Uid, "refreshToken")

		//在redis中记录颁发最近token时间  使得只有最新的token才有效
		//key 为唯一标识符Uid  val为生成时间
		tool.RedisSet(strconv.Itoa(tokenJwt.Payload.Id), tokenJwt.Payload.Iat)

		//refreshToken 过期时间较长可以存入Mysql数据库
		newuser.RefreshTokenIat = refreshTokenJwt.Payload.Iat
		err = us.ChangeUser(olduser, newuser)

		if err != nil {
			fmt.Println("refresh token iat failed")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "true",
			"data":         "",
			"token":        tokenJwt.Token,
			"uid":          olduser.Uid,
			"refreshToken": refreshTokenJwt.Token,
		})
	}
}


func GetUid(c *gin.Context){
	uid:=c.Param("uid")
	//正则表达式判断输入的Uid是否为正整数
	regl:=regexp.MustCompile(`^[1-9]\d*$`)
	if regl==nil{
		fmt.Println("regexp.MustCompile err")
	}
	result:=regl.FindAllStringSubmatch(uid,-1)
	if result==nil{
		tool.PrintFalse(c,"UID 无效")
		return
	}

	//从用户表中查找Uid
	intUid,_:=strconv.Atoi(uid)
	newuser:=new(module.User)
	newuser.Uid=intUid
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"UID 无效")
		return
	}

	//将结构体信息转到map中
	str,err:=json.Marshal(olduser)
	if err!=nil{
		fmt.Println("json marshal err:",err)
		return
	}
	var data map[string]interface{}
	err=json.Unmarshal(str,&data)
	if err!=nil{
		fmt.Println("json unmarshal err:",err)
		return
	}
	fmt.Println(data)

	//删除表中一些键值对
	delete(data,"pwd_hash")
	delete(data,"pwd_salt")
	delete(data,"incr_exp_by_coins")
	delete(data,"last_check_in_date")
	delete(data,"last_like_time")
	delete(data,"last_view_time")
	delete(data,"phone")
	delete(data,"refresh_token_iat")

	//添加saves  先找出saveId
	newUserWithVideo:=new(module.UserWithVideo)
	newUserWithVideo.Uid=uid
	newUserWithVideo.CoinsId=""
	newUserWithVideo.LikeId=""

	oldUserWithVideoSlice,_:=us.FindLikes(newUserWithVideo) //返回的是结构体数组指针
	data["saves"]=*oldUserWithVideoSlice

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":data,
	})
}

//修改密码
func ChangePw(c *gin.Context){
	account:=c.PostForm("account") //没写账号邮箱号,先默认只有手机号
	code:=c.PostForm("code")
	new_pw:=c.PostForm("new_password")

	if account==""{
		tool.PrintFalse(c,"账号为空")
		return
	}

	newuser:=new(module.User)
	newuser.Phone=account
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"账号不存在")
		return
	}

	if code==""{
		tool.PrintFalse(c, "验证码为空")
		return
	}

	//比对验证码
	smscode:=tool.RedisGet(account)
	if smscode != code {
		tool.PrintFalse(c, "验证码错误")
		return
	}


	if len(new_pw)<6{
		tool.PrintFalse(c,"密码不能小于6个字符")
	}else if len(new_pw)>16{
		tool.PrintFalse(c,"密码不能大于16个字符")
	}else {
		//修改密码----数据库中储存的为密码哈希
		pwdSalt, pwdHash := us.Encrypt_md5(new_pw)
		newuser.PwdSalt = pwdSalt
		newuser.PwdHash = pwdHash
		err = us.ChangeUser(olduser, newuser)

		if err!=nil {
			tool.PrintFalse(c, "系统修改密码失败")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "true",
			"data":   "",
		})

	}
}

//修改密保手机
func ChangePhone(c *gin.Context) {
	var param param2.Param
	err := c.ShouldBind(&param)
	if err != nil {
		fmt.Println("should bind err:", err)
		return
	}

	//先验证token
	_,err=tool.ParseToken(param.Token,c)
	if err!=nil{
		return
	}


	if param.Old_account == "" {
		tool.PrintFalse(c, "原账号为空")
		return
	}

	newuser := new(module.User)
	newuser.Phone = param.Old_account


	olduser,err:= us.FindUser(newuser)
	if err != nil {
		tool.PrintFalse(c, "原账号不存在")
		return
	}

	if param.Old_code == "" {
		tool.PrintFalse(c, "验证码为空")
		return
	}

	//先验证旧手机发过来的验证码
	smscode:=tool.RedisGet(param.Old_account)
	if smscode != param.Old_code {
		tool.PrintFalse(c, "验证码错误")
		return
	}

	if param.New_phone==""{
		tool.PrintFalse(c,"新手机号为空")
		return
	}else if len(param.New_phone)!=11 {
		tool.PrintFalse(c,"新手机号无效")
		return
	}

	//验证新手机号是否已使用
	newuser.Phone=param.New_phone
	_,err=us.FindUser(newuser)
	if err==nil{
		tool.PrintFalse(c,"新手机号已被使用")
		return
	}

	if param.New_code == "" {
		tool.PrintFalse(c, "新手机号验证码为空")
		return
	}

	//比对新手机的验证码
	smscode=tool.RedisGet(param.New_phone)
	if smscode != param.New_code {
		tool.PrintFalse(c, "新手机号验证码错误")
		return
	}
	//更改密保手机
	err=us.ChangeUser(olduser,newuser)  //olduser 此时为old_account 对应的那个对象
	if err!=nil{
		tool.PrintFalse(c,"修改密保手机失败")
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"",
	})
}

//修改用户名
func ChangeUsername(c *gin.Context){
	token:=c.PostForm("token")
	newUsername:=c.PostForm("new_username")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	//fmt.Println("len new_username",len([]rune(new_username)))  //test
	if newUsername==""{
		tool.PrintFalse(c,"昵称不可为空")
		return
	}else if len([]rune(newUsername))>15{  //username 可能为中文
		tool.PrintFalse(c,"昵称太长了")
		return
	}

	newuser:=new(module.User)
	newuser.Uid=jwt.Payload.Id
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"找不到该用户信息")
		return
	}
	if newUsername==olduser.Name{
		tool.PrintFalse(c,"请输入一个新的名字")
		return
	}

	if olduser.Coins<6{
		tool.PrintFalse(c,"硬币不足")
		return
	}

	//用户名更改与硬币的扣除应该开启事务  保证同时完成或者回滚
	newuser=new(module.User)
	if olduser.Coins==6{
		newuser.Coins=-1  //使用Struct的时候，只会更新Struct中这些非空的字段
	}else {
		newuser.Coins = olduser.Coins - 6 //
	}
	newuser.Name=newUsername
	err=us.CreateAnimals(olduser,newuser)
	if err!=nil{
		tool.PrintFalse(c,"系统修改失败 服务器错误")
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"",
	})
}


//修改个性签名
func ChangeStatement(c *gin.Context){
	token:=c.PostForm("token")
	new_statement:=c.PostForm("new_statement")

	//解析token
	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	if new_statement==""{
		tool.PrintFalse(c,"签名不能为空")//？？？？？
		return
	}
	//fmt.Println("len new_username",len([]rune(new_statement)))  //test
	if len([]rune(new_statement))>15{
		tool.PrintFalse(c,"签名太长了")
		return
	}

	newuser:=new(module.User)
	newuser.Uid=jwt.Payload.Id
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"找不到该用户信息")
		return
	}

	newuser.Statement=new_statement
	err=us.ChangeUser(olduser,newuser)
	if err!=nil{
		tool.PrintFalse(c,"修改用户签名失败")
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"",
	})
}


//修改性别信息
func ChangeGender(c *gin.Context){
	token:=c.PostForm("token")
	new_gender:=c.PostForm("new_gender")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}


	//不准设置奇奇怪怪的性别
	gender:=make([]string,3)
	gender[0]="男"
	gender[1]="女"
	gender[2]="秘密"

	flag:=false
	for _,v:=range gender{
		if new_gender==v{
			flag=true
			break
		}
	}
	if flag==false{
		tool.PrintFalse(c,"无效的性别")
		return
	}

	//修改性别
	newuser:=new(module.User)
	newuser.Uid=jwt.Payload.Id
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"找不到该用户信息")
		return
	}

	newuser.Gender=new_gender
	err=us.ChangeUser(olduser,newuser)
	if err!=nil{
		tool.PrintFalse(c,"服务器修改性别失败")
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"",
	})
}

//修改生日信息
func ChangeBirth(c *gin.Context){
	token:=c.PostForm("token")
	newBirth:=c.PostForm("new_birth")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}


	/*format:=time.Now().Format("2006-01-02") //format为当日的日期并且会将年月日之间的横杠识别
	//fmt.Println("format:",format) test*/
	timeLayout:="2006-01-02"

	_,err=time.Parse(timeLayout,newBirth)
	if err!=nil{
		fmt.Println("time parse err:",err)
		tool.PrintFalse(c,"日期格式错误")
		return
	}

	// new_birth为待转化为时间戳的字符串
	loc, _ := time.LoadLocation("Local")    //获取时区
	tmp, _ := time.ParseInLocation(timeLayout, newBirth, loc)
	timestamp := tmp.Unix()    //转化为时间戳 类型是int64
	if timestamp>time.Now().Unix(){
		tool.PrintFalse(c,"出生日期无效")
		return
	}

	//开始修改出生日期
	newuser:=new(module.User)
	newuser.Uid=jwt.Payload.Id
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"找不到该用户信息")
		return
	}


	newuser.Birthday=newBirth
	err= us.ChangeUser(olduser,newuser)
	if err!=nil{
		tool.PrintFalse(c,"服务器修改出生日期失败")
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"",
	})
}

//日常签到
func CheckIn(c *gin.Context){
	token:=c.PostForm("token")
	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	newuser:=new(module.User)
	newuser.Uid=jwt.Payload.Id
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"找不到该用户信息")
		return
	}

	format:=time.Now().Format("2006-01-02")
	format+=" 00:00:00"  //今日零点时间
	timeLayout := "2006-01-02 15:04:05" //定义模板
	loc, _ := time.LoadLocation("Local")    //获取时区
	tmpCheck, _ := time.ParseInLocation(timeLayout,olduser.LastCheckInDate,loc)
	tmpToday,_:=time.ParseInLocation(timeLayout,format,loc)

	//上一次签到的时间戳大于今日零点时分的时间戳
	if tmpCheck.Unix()>tmpToday.Unix(){
		tool.PrintFalse(c,"ALREADY_DONE")
		return
	}
	//fmt.Println("tmpdate:",time.Unix(tmpCheck.Unix(),0).Format(timeLayout))


	//签到奖励+更新签到时间----多步同时进行需开启事务
	newuser.LastCheckInDate=time.Unix(time.Now().Unix(),0).Format(timeLayout)
	newuser.Exp=olduser.Exp+5
	newuser.Coins=olduser.Coins+1

	err=us.CreateAnimals(olduser,newuser)
	if err!=nil{
		fmt.Println("change err:",err)
		tool.PrintFalse(c,"系统错误 事务回滚")
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"SUCCESS",
	})
}

//获取日常任务
func GetDaily(c *gin.Context){
	token:=c.Query("token")
	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	newuser:=new(module.User)
	newuser.Uid=jwt.Payload.Id
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"找不到该用户信息")
		return
	}

	//获取当前日期
	format:=time.Now().Format("2006-01-02")
	format+=" 00:00:00"
	fmt.Println("Today :",format) //test
	timeLayout:="2006-01-02 15:04:05"
	loc,_:=time.LoadLocation("Local")

	tmp,_:=time.ParseInLocation(timeLayout,format,loc)
	tmpCheckIn,_:=time.ParseInLocation(timeLayout,olduser.LastCheckInDate,loc)

	checkInFlag:=true
	if tmpCheckIn.Unix()<tmp.Unix(){ //上次签到时间小于今日零点时间戳
		checkInFlag=false
	}

	CoinsFlag :=false
	if olduser.Coins>=0&&olduser.Coins<=50{
		CoinsFlag=true
	}

	ViewFlag :=false
	date,_:=time.Parse("2006-01-02 15:04:05",time.Now().Format("2006-01-02 00:00:00"))
	if olduser.LastViewTime.Unix()>date.Unix(){
		ViewFlag=true
	}

	data:=make(map[string]interface{})
	data["check-in"]=checkInFlag
	data["view"]=ViewFlag  //当天是否观看视频
	data["coin"]=CoinsFlag  //coin number 在0到50之间为true

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":data,
	})

}


func FollowUser(c *gin.Context){
	token:= c.PostForm("token")
	uid:=c.PostForm("uid")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil {
		return
	}


	//正则表达式判断输入的Uid是否为正整数
	regl:=regexp.MustCompile(`^[1-9]\d*$`)
	if regl==nil{
		fmt.Println("regexp.MustCompile err")
	}
	result:=regl.FindAllStringSubmatch(uid,-1)
	fmt.Println("result:",result)//test
	if result==nil {//若其中夹杂有非正整数字符将无法匹配出
		tool.PrintFalse(c, "uid 无效")
		return
	}

	newuser:=new(module.User)
	newuser.Uid=jwt.Payload.Id
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		tool.PrintFalse(c,"找不到该用户信息")
		return
	}

	intUid,_:=strconv.Atoi(uid)
	if intUid==olduser.Uid{
		tool.PrintFalse(c,"你时刻都在关注自己")
		return
	}

	//此时再去从数据库检查uid的合法性
	newFollower:=new(module.User)
	newFollower.Uid=intUid
	followuser,err:=us.FindUser(newFollower)
	if err!=nil{
		fmt.Println("find follow user err:",err)
		tool.PrintFalse(c,"uid 无效")
		return
	}

	//取消关注-----先关注才能取关---判断follow_id是否已经存在
	err=us.FindFollowID(strconv.Itoa(olduser.Uid),uid)
	if err==nil { //能找到该用户--之前关注过
		err = us.DeleteFollows(strconv.Itoa(olduser.Uid), uid)
		if err != nil {
			log.Println(err)
			return
		}

		newuser.Followers = olduser.Followers - 1
		err = us.ChangeUser(olduser, newuser)
		if err != nil {
			log.Println(err)
			return
		}

		newFollower.Followings = olduser.Followings - 1
		err = us.ChangeUser(olduser, newuser)
		if err != nil {
			log.Println(err)
			return
		}

		tool.PrintFalse(c, "取关成功")
		return
	}

	fmt.Println("err:",followuser.Uid,err)
	//若没关注过将uid写入follow表+user表
	err=us.InsertFollows(strconv.Itoa(olduser.Uid),uid,followuser.Name)
	if err!=nil{
		tool.PrintFalse(c,"服务器err---关注失败")
		return
	}

	newuser.Followers=olduser.Followers+1
	err=us.ChangeUser(olduser,newuser)
	if err!=nil{
		log.Println(err)
		return
	}

	newFollower.Followings=olduser.Followings+1
	err=us.ChangeUser(olduser,newuser)
	if err!=nil{
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"关注成功",
	})
}

func GetFollows(c *gin.Context){
	a:=c.Query("a") //A用户ID
	b:=c.Query("b") //B用户ID

	//判断参数有效性
	//正则表达式判断输入的Uid是否为正整数
	regl:=regexp.MustCompile(`^[1-9]\d*$`)
	if regl==nil{
		fmt.Println("regexp.MustCompile err")
	}
	result1:=regl.FindAllStringSubmatch(a,-1)
	result2:=regl.FindAllStringSubmatch(b,-1)

	if result1==nil||result2==nil||a==b{//若其中夹杂有非正整数字符将无法匹配出
		tool.PrintFalse(c, "uid 无效")
		fmt.Println(a,b,result1,result2)
		return
	}

	//在user表中匹配
	newuser:=new(module.User)
	inta,_:=strconv.Atoi(a)
	intb,_:=strconv.Atoi(b)
	newuser.Uid=inta
	_,err1:=us.FindUser(newuser)

	newuser.Uid=intb
	_,err2:=us.FindUser(newuser)
	if err1!=nil||err2!=nil{
		tool.PrintFalse(c, "uid 无效")
		return
	}

	err1=us.FindFollowID(b,a)//将a作为follow_id查询
	err2=us.FindFollowID(a,b)//将b作为follow_id查询
	if err1==nil&&err2==nil{  //AB之间互相关注
		c.JSON(http.StatusOK,gin.H{
			"status": "true",
			"data":2,
		})
	}else if err1==nil&&err2!=nil{//B关注了A
		c.JSON(http.StatusOK,gin.H{
			"status": "true",
			"data":-1,
		})
	}else if err1!=nil&&err2==nil{//A关注了B
		c.JSON(http.StatusOK,gin.H{
			"status": "true",
			"data":1,
		})
	}else {  //AB 互不关注
		c.JSON(http.StatusOK,gin.H{
			"status": "true",
			"data":0,
		})
	}
}


func CheckUsername(c *gin.Context){
	username:=c.Param("username")
	if username==""{
		tool.PrintFalse(c,"请告诉我你的昵称吧")
		return
	}else if len([]rune(username))>14{
		tool.PrintFalse(c,"昵称过长")
		return
	}

	//数据库查找username
	newuser:=new(module.User)
	newuser.Name=username
	_,err:=ud.FindUser(newuser)
	if err==nil{
		tool.PrintFalse(c,"昵称已存在")
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status": "true",
		"data":   "",
	})

}

func CheckPhone(c *gin.Context){
	phone:= c.Param("phone")
	if phone==""{
		tool.PrintFalse(c,"请告诉我你的手机号吧")
		return
	}else if len(phone)!=11{
		tool.PrintFalse(c,"手机号不合法")
		return
	}

	regl:=regexp.MustCompile(`^[0-9]\d*$`)
	if regl==nil{
		fmt.Println("regexp.MustCompile err")
	}
	result:=regl.FindAllStringSubmatch(phone,-1)
	if result==nil{
		tool.PrintFalse(c,"手机号不合法")
		return
	}

	newuser:=new(module.User)
	newuser.Phone=phone
	_,err:=us.FindUser(newuser)
	if err==nil{
		tool.PrintFalse(c,"手机号已被使用")
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status": "true",
		"data":   "",
	})
}


func ChangeAvtar(c *gin.Context){
	token:=c.PostForm("token")
	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	newuser:=new(module.User)
	newuser.Uid=jwt.Payload.Id
	olduser,err:=us.FindUser(newuser)
	if err!=nil{
		return
	}

	file,fileheader,err:=c.Request.FormFile("avatar")//获取请求文件
	if err!=nil{
		fmt.Println("get form file err:",err)
		return
	}
	if fileheader.Size==0{//文件大于2MB
		tool.PrintFalse(c,"头像无效")
		return
	}else if fileheader.Size>(2<<20){
		tool.PrintFalse(c,"头像文件过大")
		return
	}

	//上传文件
	var oss service.OssService
	format:=fileheader.Filename //fileHeader type
	format=strings.ToLower(format)
	if format!="jpg"&&format!="png"{
		tool.PrintFalse(c,"图片格式不兼容")
		return
	}

	cfg:=tool.GetCfg().Oss

	filePath:=cfg.AvatarUrl+strconv.Itoa(olduser.Uid)+"."+format
	err=oss.UploadAvtar(filePath,file)
	if err!=nil{
		tool.PrintFalse(c,"图片上传失败")
		return
	}

	//将图片地址存入数据库
	newuser.Avatar=filePath
	err=us.ChangeUser(olduser,newuser)
	if err!=nil{
		log.Println("数据库存入失败:",err)
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"上传成功",
	})

}



