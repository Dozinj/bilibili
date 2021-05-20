package service

import (
	"bilibili/dao"
	"bilibili/module"
	"fmt"
)

type VideoService struct {

}
var vd dao.VideoDao
func (vs VideoService)FindVideo(newVideo *module.Video)(*module.Video,error) {
	oldVideo,err:=vd.FindVideo(newVideo)
	if err!=nil{
		fmt.Println("find video err:",err)
		return nil, err
	}
	return oldVideo,nil
}

func (vs VideoService)FindLastVideo()(*module.Video,error){
	return vd.FindLastVideo()
}

func (vs VideoService)CreateVideo(video *module.Video)error{
	return vd.CreateVideo(video)
}

func (vs VideoService)UpdateVideo(oldVideo,newVideo *module.Video)error{
	return vd.UpdateVideos(oldVideo,newVideo)
}

func (vs VideoService)CreateVideoComment(videoComment *module.VideoComment)error{
	return vd.CreateVideoComment(videoComment)
}

func (vs VideoService)FindVideoComment(newVideoComment *module.VideoComment)(*[]module.VideoComment,error){
	return vd.FindVideoComment(newVideoComment)
}