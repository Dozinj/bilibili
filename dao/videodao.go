package dao

import (
	"bilibili/module"
	"bilibili/tool"
	"errors"
	"fmt"
)


type VideoDao struct {

}

func (vd VideoDao)FindVideo(newVideo *module.Video)(*module.Video,error){
	if newVideo==nil{
		err:=errors.New("newVideo is  nil pointer")
		return nil,err
	}

	oldVideo:=new(module.Video)
	err:=tool.DBEngine.Where(newVideo).First(oldVideo).Error
	if err!=nil{
		return nil, err
	}
	return oldVideo,nil
}


func (vd VideoDao)FindLastVideo()(*module.Video,error){
	var lastVideo module.Video
	err:=tool.DBEngine.Model(&module.Video{}).Last(&lastVideo).Error
	return &lastVideo,err
}

func (vd VideoDao)CreateVideo(video *module.Video)error{
	return tool.DBEngine.Create(video).Error
}


func (vd VideoDao)UpdateVideos(oldVideo,newVideo *module.Video)error{
	return tool.DBEngine.Model(&module.Video{}).Where(oldVideo).Update(newVideo).Error
}


//评论相关
func (vd VideoDao)CreateVideoComment(videoComment *module.VideoComment)error{
	return tool.DBEngine.Create(videoComment).Error
}

func (vd VideoDao)FindVideoComment(newVideoComment *module.VideoComment)(*[]module.VideoComment,error){
	var oldVideoComment []module.VideoComment
	err:=tool.DBEngine.Where(newVideoComment).Find(&oldVideoComment).Error
	return &oldVideoComment,err
}

//search 模糊查询
func (vd VideoDao)FindByLikeSearch(KeyWords string)(*[]module.Video,error){
	oldVideo:=new([]module.Video)
	err:=tool.DBEngine.Where(fmt.Sprintf(" title like '%%%s'",KeyWords )).Find(oldVideo).Error
	return oldVideo,err
}