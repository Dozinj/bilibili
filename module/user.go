package module

import "time"

type User struct {
	Uid int `gorm:"primary_key;auto_increment" json:"uid"`//i
	Name string  `gorm:"unique" json:"name"`//i
	PwdSalt string `json:"pwd_salt"`//
	PwdHash string `json:"pwd_hash"`//
	Phone string `json:"phone"`//
	RefreshTokenIat string `json:"token_iat"`   //token_iat 记录的是refreshToken生成的时间
	Avatar  string `json:"avatar"`  //用户头像----头像地址  i
	RegDate string `json:"reg_date"`  //注册时间  i
	Statement string `json:"statement"`  //个性签名  i
	Exp int `json:"exp"`  //经验  i
	Coins int `json:"coins"`//  i
	BCoins int `json:"b_coins"`//  i
	Birthday string `json:"birthday"`//  i
	Gender string `json:"gender"` //性别   i
	Videos []byte `json:"videos"`// []Video；投稿视频数组（切片）  i
	Followers int  `json:"followers"`// 关注数   i
	Followings int `json:"followings"`// 粉丝数   i
	TotalViews int `json:"total_views"`//总浏览数   i
	TotalLikes int `json:"total_likes"`//总喜欢数    i
	LastCheckInDate string `json:"last_check_in_date"` //最后一次签到时间
	IncrExpByCoinsTimes int `json:"incr_exp_by_coins"`  //当日通过投币增加经验的次数
	LastCoinTime time.Time `json:"last_like_time"`  //最后一次投币时间
	LastViewTime time.Time `json:"last_view_time"`//最后一次浏览时间
}

type Follows struct {
	Id int `json:"id"`
	Uid string `gorm:"not null;" json:"uid"` //关注者id
	FollowId string `gorm:"not null;" json:"follow_id"`  //被关注者id
	FollowName string `gorm:"not null;" json:"follow_name"`
}

type UserWithVideo struct {
	Id int `json:"id"`
	Uid string  `json:"uid"`
	LikeId string `json:"follow_id"`  //喜欢的视频id
	CoinsId string `json:"coins_id"`   //投币的视频id
	SaveId string `json:"save_id"`
}

