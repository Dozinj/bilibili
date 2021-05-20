package module

type Danmaku struct {
	Id int `json:"id"`
	VideoId string `json:"video_id"`
	UserId  string `json:"user_id"`  //发送弹幕的用户 ID;
	Value string `json:"value"`  //弹幕内容
	Color string `json:"color"`  //弹幕颜色
	Time string `json:"time"`  //弹幕发送时间
	Type string `json:"type"`  //弹幕类型
	Location string `json:"location"`  //弹幕弹出位置  单位为秒; 例如此处弹幕发送于 1min 54s 处location为 114
}
