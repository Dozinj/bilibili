package contorller

import (
	"bilibili/module"
	param2 "bilibili/param"
	"bilibili/service"
	"bilibili/tool"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)
var vs service.VideoService
var oss service.OssService


//获取请求视频----传回视频地址
func GetVideo(c *gin.Context){
	VideoId:=c.Query("video_id")
	if VideoId==""{
		tool.PrintFalse(c,"视频 ID 不可为空")
		return
	}

	newVideo:=new(module.Video)
	intVideoId,_:=strconv.Atoi(VideoId)
	newVideo.Id=intVideoId

	oldVideo,err:=vs.FindVideo(newVideo)
	if err!=nil{
		tool.PrintFalse(c,"视频 ID 无效")
		return
	}

	var data map[string]interface{}
	info,err:=json.Marshal(*oldVideo)
	if err!=nil{
		fmt.Println("json.Marshal err:",err)
		return
	}

	err=json.Unmarshal(info,&data)
	if err!=nil{
		fmt.Println("json.Unmarshal err:",err)
		return
	}

	//将label转为字符串切片
	buf:=new(bytes.Buffer)
	buf.Write(oldVideo.Labels)//先向缓冲器中写入信息
	err=gob.NewDecoder(buf).Decode(&(oldVideo.Labels))
	if err!=nil&&err.Error()!="EOF"{
		log.Println(err)
		return
	}

	//从弹幕表中找到该视频弹幕切片
	newDanMaKu:=new(module.Danmaku)
	newDanMaKu.VideoId=VideoId

	var ds service.DanmakuService
	oldDanMakuSlice,err:=ds.FindDanMaKu(newDanMaKu)  //
	if err!=nil{
		log.Println(err)
		return
	}
	data["Danmakus"]=*oldDanMakuSlice

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":data,
	})
}


//视频投稿
func PostVideo(c *gin.Context){
	var postVideo param2.PostVideo
	err:=c.ShouldBind(&postVideo)
	if err!=nil{
		log.Println("bind struct err:",err)
		return
	}

	//解析token
	jwt,err:=tool.ParseToken(postVideo.Token,c)
	if err!=nil{
		return
	}

	//标题
	if postVideo.Title==""{
		tool.PrintFalse(c,"标题不可为空")
		return
	}else if len([]rune(postVideo.Title))>80{
		tool.PrintFalse(c,"标题过长")
		return
	}

	//视频时长
	length:=strings.Split(postVideo.Length,":")
	if len(length)!=2{
		tool.PrintFalse(c,"时长无效")
		return
	}

	//分区channel
	filehander,err:=os.Open("./channel.md")
	if err!=nil{
		log.Println("open channel err:",err)
		return
	}

	defer filehander.Close()
	filebyte,err:=ioutil.ReadAll(filehander)
	if err!=nil{
		log.Println("read channel err:",err)
		return
	}

	reg:=regexp.MustCompile(`\d{4}`)
	if reg==nil{
		log.Println("正则表达式编译失败")
		return
	}
	channelSlice:=reg.FindAllStringSubmatch(string(filebyte),-1)
	flag:=false
	for _,v:=range channelSlice{
		if postVideo.Channel==v[1]{
			flag=true
			break
		}
	}
	if flag==false{
		tool.PrintFalse(c,"分区无效")
		return
	}

	//标签-----数组转 json 字符串

	labelJson:=c.PostForm("label")
	var label []string
	err=json.Unmarshal([]byte(labelJson),&label)
	if err!=nil{
		log.Println(err)
		return
	}

	//判断每个标签长度
	for _,v:=range label{
		if utf8.RuneCountInString(v)>10{
			tool.PrintFalse(c,"标签过长")
			return
		}
	}

	//标签去重  ---通过Map主键唯一性
	afterStr:=make([]string,0)
	tmpMap:=make(map[string]interface{})

	for _,label:=range label{
		if _,ok:=tmpMap[label];!ok{
			afterStr=append(afterStr,label)
			tmpMap[label]=nil
		}
	}

	//去重后标签数不能多余10个
	if len(afterStr)==0||len(afterStr)>10{
		tool.PrintFalse(c,"标签无效")
		return
	}
	//视频描述
	if utf8.RuneCountInString(postVideo.Description)>250{
		tool.PrintFalse(c,"简介过长")
		return
	}


	//上传视频
	videofile,videoheader,err:=c.Request.FormFile("video")
	if err!=nil{
		log.Println("get video err:",err)
		return
	}
	if videoheader.Size==0{
		tool.PrintFalse(c,"视频不可为空")
		return
	}else if videoheader.Size>(2048<<20){
		tool.PrintFalse(c,"视频体积不可大于 2GB")
		return
	}

	//视频格式 .......

	cfg:=tool.GetCfg().Oss
	//上传视频----
	lastVideo,err:=vs.FindLastVideo()
	if err!=nil{
		log.Println("find video err:",err)
		return
	}
	filepath:= cfg.VideosUrl+strconv.Itoa(lastVideo.Id+1)+"."+videoheader.Filename

	err=oss.UploadVideo(filepath,videofile)
	if err!=nil{
		tool.PrintFalse(c,"视频上传失败")
		return
	}

	//上传封面----videoId++
	coverfile,coverheader,err:=c.Request.FormFile("cover")
	if err!=nil{
		log.Println("get cover err:",err)
		return
	}

	if coverheader.Size==0{
		tool.PrintFalse(c,"封面不可为空")
		return
	}else if coverheader.Size>(2<<20){
		tool.PrintFalse(c,"封面体积不可大于 2MB")
		return
	}

	if coverheader.Filename!="jpg"&&coverheader.Filename!="png"{
		tool.PrintFalse(c,"封面格式无效")
		return
	}

	//上传封面

	avterpath:= cfg.AvatarUrl +strconv.Itoa(lastVideo.Id+1)+"."+coverheader.Filename

	err=oss.UploadAvtar(avterpath,coverfile)
	if err!=nil{
		tool.PrintFalse(c,"封面上传失败")
		return
	}

	timeLayout:="2006-01-02 15:04:05"
	datetime:=time.Unix(time.Now().Unix(),0).Format(timeLayout)

	//将标签二进制流存入数据表
	labelBuf:=new(bytes.Buffer)
	err=gob.NewEncoder(labelBuf).Encode(&label)
	if err!=nil{
		log.Println(err)
		return
	}

	//将video 存入数据库
	newVideo:=module.Video{
		Video: filepath,
		Cover: avterpath,
		Title: postVideo.Title,
		Length: postVideo.Length,
		Channel: postVideo.Channel,
		Description: postVideo.Description,
		Author: jwt.Payload.Id,
		Labels: labelBuf.Bytes(), //二进制数组
		Time:datetime,
	}
	err=vs.CreateVideo(&newVideo)
	if err!=nil{
		tool.PrintFalse(c,"保存视频失败")
		return
	}

	//用户投稿视频存入用户表中
	newUser:=new(module.User)
	newUser.Uid=jwt.Payload.Id

	oldUser,err:=us.FindUser(newUser)
	if err!=nil{
		log.Println("find user err:",err)
		return
	}

	
	videoBuf:=new(bytes.Buffer)
	err=binary.Write(videoBuf,binary.LittleEndian,newVideo)  //适用于结构体->[]byte
	if err!=nil{
		log.Println(err)
		return
	}
	newUser.Videos=append(oldUser.Videos,videoBuf.Bytes()...)//此合并方法只支持两个参数
	
	err=us.ChangeUser(oldUser,newUser)
	if err!=nil{
		log.Println("update user []video err:",err)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":lastVideo.Id,    //视频id
	})
}

//获取视频弹幕信息
func GetDanmaku(c *gin.Context){
	videoId:=c.Query("video_id")
	if videoId==""{
		tool.PrintFalse(c,"视频 ID 不可为空")
		return
	}

	newDanMaKu:=new(module.Danmaku)
	newDanMaKu.VideoId=videoId

	var ds service.DanmakuService
	oldDanMakuSlice,err:=ds.FindDanMaKu(newDanMaKu) //返回该视频的全部弹幕
	if err!=nil{
		tool.PrintFalse(c,"视频 ID 无效")
		return
	}

	for _,oldDanMaku:=range *oldDanMakuSlice{
		dataByte, err := json.Marshal(oldDanMaku)
		if err != nil {
			log.Println("json marshal err:", err)
			return
		}

		var data map[string]interface{}

		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			log.Println("json unmarshal err:", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "true",
			"data":   data,
		})
	}
}

//上传弹幕
func SendDanMaku(c *gin.Context){  //application/json
	var DanMaku param2.ParamDanMaku
	//err:=c.ShouldBindJSON(&DanMaku)
	err:=c.ShouldBind(&DanMaku)
	if err!=nil{
		log.Println(err)
		return
	}

	jwt,err:=tool.ParseToken(DanMaku.Token,c)
	if err!=nil{
		return
	}

	if DanMaku.Value==""{
		tool.PrintFalse(c,"弹幕不可为空")
		return
	}else if utf8.RuneCountInString(DanMaku.Value)>100{
		tool.PrintFalse(c,"弹幕太长")
		return
	}

	//videoID
	oldVideo,err:=vs.FindLastVideo()
	if err!=nil{
		log.Println("find video err:",err)
		return
	}
	intVideId,_:=strconv.Atoi(DanMaku.VideoId)
	if oldVideo.Id<intVideId||DanMaku.VideoId==""{
		tool.PrintFalse(c,"参数无效")
		return
	}

	//color
	if DanMaku.Color==""{
		tool.PrintFalse(c,"参数无效")
		return
	}

	//type
	str:="scroll,top,bottom"
	strSlice:=strings.Split(str,",")
	flag:=false
	for _,v:=range strSlice{
		if DanMaku.Type==v{
			flag=true
			break
		}
	}

	if flag==false{
		tool.PrintFalse(c,"参数无效")
		return
	}

	//location
	if DanMaku.Location==""{
		tool.PrintFalse(c,"参数无效")
		return
	}

	timeLayout:="2006-01-02 15:04:05"
	datetime:=time.Unix(time.Now().Unix(),0).Format(timeLayout)

	var DanmakuModule=module.Danmaku{
		VideoId: DanMaku.VideoId,
		UserId:strconv.Itoa(jwt.Payload.Id),
		Value: DanMaku.Value,
		Color: DanMaku.Color,
		Type: DanMaku.Type,
		Time: datetime,
		Location: DanMaku.Location,
	}

	var ds service.DanmakuService
	err=ds.CreateDanMaKu(&DanmakuModule)

	if err!=nil{
		log.Println("create danmaku err:",err)
		tool.PrintFalse(c,"保存弹幕失败")
		return
	}

	dataByte,err:=json.Marshal(DanmakuModule)
	if err!=nil{
		log.Println("json marshal err:",err)
		return
	}

	var data map[string]interface{}
	err=json.Unmarshal(dataByte,&data)
	if err!=nil{
		log.Println("json unmarshal err:",err)
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":data,
	})
}


//给视频点赞
func PostLike(c *gin.Context) {
	token := c.PostForm("token")
	videoId := c.PostForm("video_id")

	jwt, err := tool.ParseToken(token, c)
	if err != nil {
		return
	}

	//查找视频ID是否存在
	newVideo := new(module.Video)
	intVideoId, _ := strconv.Atoi(videoId)
	newVideo.Id = intVideoId

	if videoId == "" {
		tool.PrintFalse(c, "视频ID无效")
		return
	}
	oldVideo, err:= vs.FindVideo(newVideo)
	if err != nil {
		tool.PrintFalse(c, "视频ID无效")
		return
	}

	//查找Like表
	newLikes := new(module.UserWithVideo)
	newLikes.Uid = strconv.Itoa(jwt.Payload.Id)
	newLikes.LikeId = videoId

	_, err = us.FindLikes(newLikes)
	if err == nil {
		//视频已点赞
		err = us.DeleteLikes(newLikes)
		if err != nil {
			log.Println("delete likes err:", err)
			return
		}

		//视频喜欢数减一
		newVideo.Likes=oldVideo.Likes-1
		fmt.Println("video likes",newVideo.Likes)
		err=vs.UpdateVideo(oldVideo,newVideo)
		if err != nil {
			log.Println("update video likes err:", err)
			return
		}
		tool.PrintFalse(c, "取消点赞成功")
		return

	}
	//视频未点赞
	err = us.CreateLikes(newLikes)
	if err != nil {
		log.Println("create likes err:", err)
		return
	}

	newVideo.Likes=oldVideo.Likes+1//视频喜欢数加一
	fmt.Println("video likes",newVideo.Likes)

	err=vs.UpdateVideo(oldVideo,newVideo)
	if err != nil {
		log.Println("update video likes err:", err)
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"点赞成功",
	})
}

//获取点赞信息
func GetLikes(c *gin.Context){
	token:=c.Query("token")
	videoId:=c.Query("video_id")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	//查找视频ID是否存在
	newVideo := new(module.Video)
	intVideoId, _ := strconv.Atoi(videoId)
	newVideo.Id = intVideoId

	if videoId == "" {
		tool.PrintFalse(c, "视频ID无效")
		return
	}
	_, err = vs.FindVideo(newVideo)
	if err != nil {
		tool.PrintFalse(c, "视频ID无效")
		return
	}

	newLikes:=new(module.UserWithVideo)
	newLikes.Uid=strconv.Itoa(jwt.Payload.Id)
	newLikes.LikeId=videoId

	_,err=us.FindLikes(newLikes)
	if err!=nil{
		tool.PrintFalse(c,"未点赞")
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"已点赞",
	})

}


//视频投币
func PostCoins(c *gin.Context){
	token:=c.PostForm("token")
	videoId:= c.PostForm("video_id")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	//查找视频ID是否存在
	newVideo := new(module.Video)
	intVideoId, _ := strconv.Atoi(videoId)
	newVideo.Id = intVideoId

	if videoId == "" {
		tool.PrintFalse(c, "视频ID无效")
		return
	}
	oldVideo, err := vs.FindVideo(newVideo)
	if err != nil {
		tool.PrintFalse(c, "视频ID无效")
		return
	}

	//视频作者是自己
	if oldVideo.Author==jwt.Payload.Id{
		tool.PrintFalse(c,"不能给自己投币哦")
		return
	}

	newUser:=new(module.User)
	newUser.Uid=jwt.Payload.Id
	oldUser,err:=us.FindUser(newUser)

	if err!=nil{
		log.Println("find user err:",err)
		return
	}else if oldUser.Coins<=0{
		tool.PrintFalse(c,"硬币不足")
		return
	}


	//开启事务处理
	tx:=tool.DBEngine.Begin()

	//每投币一次，投币者硬币-1,经验+10，每日上限50；
	newUser.LastCoinTime=time.Now()
	newUser.Coins=oldUser.Coins-1


	date,_:=time.Parse("2006-01-02 15:04:05",time.Now().Format("2006-01-02 00:00:00"))
	if date.Unix()>oldUser.LastCoinTime.Unix(){
		//当日凌晨清零前日的IncrExpByCoins
		newUser.IncrExpByCoinsTimes=1
		newUser.Exp=oldUser.Exp+10
	} else if date.Unix()<oldUser.LastCoinTime.Unix()&&oldUser.IncrExpByCoinsTimes<5{
		newUser.IncrExpByCoinsTimes=oldUser.IncrExpByCoinsTimes+1
		newUser.Exp=oldUser.Exp+10
	}

	err=us.ChangeUserByAffairs(tx,oldUser,newUser)
	if err!=nil{
		tx.Rollback()
		tool.PrintFalse(c,"false")
		return
	}

	newAuthor:=new(module.User)
	newAuthor.Uid=oldVideo.Author

	oldAuthor,err:=us.FindUser(newAuthor)
	if err!=nil{
		log.Println("find author err:",err)
		return
	}

	//UP主的视频每被投币一次，经验+1,硬币+1，每日无上限
	newAuthor.Coins=oldAuthor.Coins+1
	newAuthor.Exp=oldAuthor.Exp+1

	err=us.ChangeUserByAffairs(tx,oldAuthor,newAuthor)
	if err!=nil{
		tx.Rollback()
		tool.PrintFalse(c,"false") //投币失败
		return
	}

	//数据更新成功，事务提交
	tx.Commit()

	//记录投币记录
	newLikes:=new(module.UserWithVideo)
	newLikes.Uid=strconv.Itoa(jwt.Payload.Id)
	newLikes.CoinsId=videoId
	_,err=us.FindLikes(newLikes)

	if err!=nil{
		log.Println("not put coins before")
		err=us.CreateLikes(newLikes)

		if err!=nil{
			log.Println("create likes err:",err)
			return
		}
	}

	//在video表中记录投币
	newVideo.Coins=oldVideo.Coins+1
	err=vs.UpdateVideo(oldVideo,newVideo)
	if err!=nil{
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"true",  //投币成功
	})

}

//获取投币状态
func GetCoins(c *gin.Context){
	token:=c.Query("token")
	videoId:=c.Query("video_id")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	if videoId==""{
		tool.PrintFalse(c,"视频 ID无效")
		return
	}
	newVideo:=new(module.Video)
	intVideoID,_:=strconv.Atoi(videoId)
	newVideo.Id=intVideoID

	_,err=vs.FindVideo(newVideo)
	if err!=nil{
		tool.PrintFalse(c,"视频 ID无效")
		return
	}

	//查看投币状态
	newLikes:=new(module.UserWithVideo)
	newLikes.Uid=strconv.Itoa(jwt.Payload.Id)
	newLikes.CoinsId=videoId

	_,err=us.FindLikes(newLikes)
	if err!=nil{
		log.Println(err)
		tool.PrintFalse(c,"未投币")
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"已投币",
	})
}


//收藏视频
func PostSave(c *gin.Context){
	//user表中 对应save []video
	token:=c.PostForm("token")
	videoId:= c.PostForm("video_id")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	if videoId==""{
		tool.PrintFalse(c,"视频 ID无效")
		return
	}
	newVideo:=new(module.Video)
	intVideoID,_:=strconv.Atoi(videoId)
	newVideo.Id=intVideoID

	oldVideo,err:=vs.FindVideo(newVideo)
	if err!=nil{
		tool.PrintFalse(c,"视频 ID无效")
		return
	}

	//查看是否收藏过该视频
	newUserWithVideo:=new(module.UserWithVideo)
	newUserWithVideo.Uid=strconv.Itoa(jwt.Payload.Id)
	newUserWithVideo.SaveId=videoId

	_,err=us.FindLikes(newUserWithVideo)
	if err==nil{//视频已收藏->取消视频收藏
		newUserWithVideo.CoinsId=""
		err=us.DeleteLikes(newUserWithVideo)
		if err!=nil{
			log.Println(err)
			return
		}

		//video表中视频收藏数减一
		newVideo.Saves=oldVideo.Saves-1
		err=vs.UpdateVideo(oldVideo,newVideo)
		if err!=nil{
			log.Println(err)
			return
		}
		tool.PrintFalse(c,"取消收藏成功")
		return
	}

	//视频未收藏->点击收藏
	err=us.CreateLikes(newUserWithVideo)
	if err!=nil{
		log.Println(err)
		return
	}

	//视频收藏数加1
	newVideo.Saves=oldVideo.Saves+1
	err=vs.UpdateVideo(oldVideo,newVideo)
	if err!=nil{
		log.Println(err)
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"收藏成功",
	})
}


func PostView(c *gin.Context){
	token:=c.PostForm("token")
	videoId:= c.PostForm("video_id")


	if videoId==""{
		tool.PrintFalse(c,"视频 ID无效")
		return
	}
	newVideo:=new(module.Video)
	intVideoID,_:=strconv.Atoi(videoId)
	newVideo.Id=intVideoID

	oldVideo,err:=vs.FindVideo(newVideo)
	if err!=nil{
		tool.PrintFalse(c,"视频 ID无效")
		return
	}

	//(视频播放数 +1；
	if token==""{
		newVideo.Views=oldVideo.Views+1
		err=vs.UpdateVideo(oldVideo,newVideo)

		if err!=nil{
			log.Println("update video err:",err)
			return
		}

		c.JSON(http.StatusOK,gin.H{
			"status":"true",
			"data":"",
		})
		return
	}


	//如果带有Token则解析token，且浏览用户经验+5，每日上限5
	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}


	newUser:=new(module.User)
	newUser.Uid=jwt.Payload.Id

	oldUser,err:=us.FindUser(newUser)
	if err!=nil{
		log.Println(err)
		return
	}

	tx:=tool.DBEngine.Begin() //开启事务

	newUser.TotalViews=oldUser.TotalViews+1
	date,_:=time.Parse("2006-01-02 15:04:05",time.Now().Format("2006-01-02 00:00:00"))  //当日零点时间戳
	newUser.LastViewTime=date

	if date.Unix()>time.Now().Unix(){
		newUser.Exp=oldUser.Exp+5 //今日首次浏览
	}

	err=us.ChangeUserByAffairs(tx,oldUser,newUser)
	if err!=nil{
		log.Println("change user err:",err)
		tx.Rollback()
	}
	tx.Commit()

	if date.Unix()>time.Now().Unix(){
		c.JSON(http.StatusOK,gin.H{
			"status":"true",
			"data":"SUCCESS",
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"ALREADY_DONE",  //用户今日已浏览
	})
}

func PostShare(c *gin.Context){
	videoId:= c.PostForm("video_id")

	if videoId==""{
		tool.PrintFalse(c,"视频 ID 不可为空")
		return
	}
	newVideo:=new(module.Video)
	intVideoID,_:=strconv.Atoi(videoId)
	newVideo.Id=intVideoID

	oldVideo,err:=vs.FindVideo(newVideo)
	if err!=nil{
		tool.PrintFalse(c,"视频 ID无效")
		return
	}

	newVideo.Shares=oldVideo.Shares+1
	err=vs.UpdateVideo(oldVideo,newVideo)
	if err!=nil{
		log.Println("update video err:",err)
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":"",
	})
}

//提交评论
func PostComment(c *gin.Context){
	//评论点赞接口
	token:=c.PostForm("token")
	videoId:=c.PostForm("video_id")
	comment:=c.PostForm("comment")

	jwt,err:=tool.ParseToken(token,c)
	if err!=nil{
		return
	}

	if videoId==""{
		tool.PrintFalse(c,"视频 ID 不可为空")
		return
	}

	intVideoId,_:=strconv.Atoi(videoId)
	newVideo:=new(module.Video)
	newVideo.Id=intVideoId

	_,err=vs.FindVideo(newVideo)
	if err!=nil{
		log.Println("find video err:",err)
		tool.PrintFalse(c,"视频 ID 无效")
		return
	}

	if comment==""{
		tool.PrintFalse(c,"评论内容不可为空")
		return
	}else if utf8.RuneCountInString(comment)>1024{
		tool.PrintFalse(c,"评论内容过长")
		return
	}

	dateTime:=time.Unix(time.Now().Unix(),0).Format("2006-01-02 15:04:05")
	newVideoCommet:=&module.VideoComment{
		VideoId: videoId,
		UserId: strconv.Itoa(jwt.Payload.Id),
		Value: comment,
		Time: dateTime,
		Likes: 0,
	}

	err=vs.CreateVideoComment(newVideoCommet)
	if err!=nil{
		log.Println("create comment err:",err)
		return
	}

	oldVideoCommet,err:=vs.FindVideoComment(newVideoCommet)
	if err!=nil{
		log.Println("find err:",err)
		return
	}
	data:=make(map[string]interface{})

	dataByte,err:=json.Marshal(oldVideoCommet)
	if err!=nil{
		log.Println("marshal err:",err)
		return
	}

	err=json.Unmarshal(dataByte,&data)
	if err!=nil{
		log.Println("unmarshal err:",err)
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"status":"true",
		"data":data,
	})

}


func GetComment(c *gin.Context) {
	videoId := c.Query("video_id")
	if videoId == "" {
		tool.PrintFalse(c, "视频 ID 不可为空")
		return
	}

	intVideoId, _ := strconv.Atoi(videoId)
	newVideo := new(module.Video)
	newVideo.Id = intVideoId

	_, err := vs.FindVideo(newVideo)
	if err != nil {
		log.Println("find video err:", err)
		tool.PrintFalse(c, "视频 ID 无效")
		return
	}

	newVideoCommet := new(module.VideoComment)
	newVideoCommet.VideoId = videoId

	oldVideoCommetSlice, err := vs.FindVideoComment(newVideoCommet)  //获取的是结构体数组指针
	if err != nil {
		log.Println("find err:", err)
		return
	}
	//fmt.Printf("%#v\n",*oldVideoCommetSlice)
	for _,oldVideoComment:=range *oldVideoCommetSlice {
		data := make(map[string]interface{})
		dataByte, err := json.Marshal(oldVideoComment)
		if err != nil {
			log.Println("marshal err:", err)
			return
		}

		err = json.Unmarshal(dataByte, &data)
		if err != nil {
			log.Println("unmarshal err:", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":   data,
			"status":"true",
		})
	}
}




