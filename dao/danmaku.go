package dao

import (
	"bilibili/module"
	"bilibili/tool"
)

type DanmakuDao struct {

}

func (Dd DanmakuDao)FindDanmaku(danmaku *module.Danmaku)(*[]module.Danmaku,error){
	var oldDanmaku []module.Danmaku
	err:=tool.DBEngine.Where(danmaku).Find(&oldDanmaku).Error
	return &oldDanmaku,err
}

//保存弹幕信息
func (Dd DanmakuDao)CreateDanMaKu(danmaku *module.Danmaku)error{
	return tool.DBEngine.Create(danmaku).Error
}

