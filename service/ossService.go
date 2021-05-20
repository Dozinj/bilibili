package service

import (
	"bilibili/tool"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"mime/multipart"
)
type OssService struct {

}


func (os OssService)UploadAvtar(filePath string,file multipart.File)error{
	cfg:=tool.GetCfg().Oss
	//创建OSSClient实例
	client,err:=oss.New(cfg.EndPoint,cfg.AppKey,cfg.AppSecret)
	if err!=nil{
		return err
	}

	// 获取存储空间
	bucket,err:=client.Bucket(cfg.AvatarBucket)
	if err!=nil{
		return err
	}

	//上传阿里云
	return bucket.PutObject(filePath,file)
}


func (os OssService)UploadVideo(filePath string,file multipart.File)error{
	cfg:=tool.GetCfg().Oss

	//创建OSSClient实例
	client,err:=oss.New(cfg.EndPoint,cfg.AppKey,cfg.AppSecret)
	if err!=nil{
		return err
	}

	// 获取存储空间
	bucket,err:=client.Bucket(cfg.VideosBucket)
	if err!=nil{
		return err
	}

	//上传阿里云
	return bucket.PutObject(filePath,file)
}
