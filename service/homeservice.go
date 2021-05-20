package service

import "bilibili/module"
type HomeService struct {

}

func (hs HomeService)Search(KeyWords string)(*[]module.Video,error){
	return vd.FindByLikeSearch(KeyWords)
}
