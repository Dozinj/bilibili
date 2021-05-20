package param

type ParamDanMaku struct {
	Token string `form:"token" json:"token"`
	VideoId string `form:"video_id" json:"video_id"`
	Value string `form:"value" json:"value"`  //弹幕内容
	Color string `form:"color" json:"color"`  //弹幕颜色
	Type string `form:"type" json:"type"`  //弹幕类型
	Location string `form:"location" json:"location"`  //弹幕弹出位置  单位为秒; 例如此处弹幕发送于 1min 54s 处location为 114
}
