package module

type VideoComment struct {
	Id int `json:"id"`//评论 ID
	VideoId string `json:"video_id"` // 视频 ID
	UserId string `gorm:"not null" json:"user_id"`// 评论者 ID
	Value string `gorm:"not null" json:"value"`// 评论内容
	Time  string `gorm:"not null" json:"time"`// 评论时间，格式：2021-02-06 19:20
	Likes int  `json:"likes"`//赞数
}
