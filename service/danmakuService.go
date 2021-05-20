package service

import (
	"bilibili/dao"
	"bilibili/module"
)

type DanmakuService struct {

}
var Dd dao.DanmakuDao

func (ds DanmakuService)FindDanMaKu(newDanMaku *module.Danmaku)(*[]module.Danmaku,error){
	return Dd.FindDanmaku(newDanMaku)
}

func (ds DanmakuService)CreateDanMaKu(Danmaku *module.Danmaku)error{
	return Dd.CreateDanMaKu(Danmaku)
}