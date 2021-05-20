package module


type Video struct {
	Id int `json:"id"`//视频 ID
	Video string`json:"video"`// string, 视频地址
	Cover string `json:"cover"`// string, 封面地址
	Title string `json:"title"`// string, 视频标题
	Length string `json:"length"`// string, 视频时长
	Channel string `json:"channel"`// string, 分区，字符串编号，参见`channel.md`
	Description string `json:"description"`// string, 简介
	Author int `json:"author"`// 作者 UID
	Time string `json:"time"`// Time, 上传时间
	Views int `json:"views"`//播放次数
	Likes int`json:"likes"`// 点赞数量
	Coins int`json:"coins"`// 投币数量
	Saves  int`json:"saves"`// 被收藏数量
	Shares int`json:"shares"`// 分享数量
	Labels []byte `json:"labels"` //视频标签
}




