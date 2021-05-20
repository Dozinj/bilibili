package routers

import (
	"bilibili/contorller"
	"github.com/gin-gonic/gin"
)
func RegisterRouter()*gin.Engine{
	r:=gin.Default()
	Usergroup:=r.Group("/api/user")
	{
		Usergroup.POST("login/pw",contorller.Loginpw)
		Usergroup.POST("register",contorller.Register)
		Usergroup.POST("login/sms",contorller.Loginsms)
		Usergroup.GET("info/:uid",contorller.GetUid)
		Usergroup.PUT("password",contorller.ChangePw)
		Usergroup.PUT("phone",contorller.ChangePhone)
		Usergroup.PUT("username",contorller.ChangeUsername)
		Usergroup.PUT("statement",contorller.ChangeStatement)
		Usergroup.PUT("gender",contorller.ChangeGender)
		Usergroup.PUT("birth",contorller.ChangeBirth)
		Usergroup.PUT("check-in",contorller.CheckIn)
		Usergroup.GET("daily",contorller.GetDaily)
		Usergroup.POST("follow",contorller.FollowUser)
		Usergroup.GET("follow",contorller.GetFollows)
		Usergroup.PUT("avatar",contorller.ChangeAvtar)
	}

	Verifygroup:=r.Group("api/verify")
	{
		Verifygroup.GET("token",contorller.RefreshToken)
		Verifygroup.POST("sms/general",contorller.SmsHandle)
		Verifygroup.POST("sms/register",contorller.SmsHandle)
		Verifygroup.POST("sms/login",contorller.SmsHandle)
	}
	Checkgroup:=r.Group("api/check")
	{
		Checkgroup.GET("username",contorller.CheckUsername)
		Checkgroup.GET("phone",contorller.CheckPhone)
	}
	Vediogroup:=r.Group("api/video")
	{
		Vediogroup.GET("video",contorller.GetVideo)
		Vediogroup.POST("video",contorller.PostVideo)
		Vediogroup.GET("danmaku",contorller.GetDanmaku)
		Vediogroup.POST("danmaku",contorller.SendDanMaku)
		Vediogroup.POST("like",contorller.PostLike)
		Vediogroup.GET("like",contorller.GetLikes)
		Vediogroup.POST("coin",contorller.PostCoins)
		Vediogroup.GET("coin",contorller.GetCoins)
		Vediogroup.POST("save",contorller.PostSave)
		Vediogroup.POST("view",contorller.PostView)
		Vediogroup.POST("share",contorller.PostShare)
		Vediogroup.POST("comment",contorller.PostComment)
		Vediogroup.GET("comment",contorller.GetComment)
	}
	HomeGroup:=r.Group("api/home")
	{
		HomeGroup.GET("search",contorller.GetSearch)
	}
	return r
}
