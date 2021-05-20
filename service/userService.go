package service

import (
	"bilibili/dao"
	"bilibili/module"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/jinzhu/gorm"
	"math/rand"
	"strconv"
	"time"
)
type UserService struct {

}

var ud dao.Userdao
func (us *UserService)CreateUser(user *module.User)error{
	return ud.CreateUser(user)
}


func(us *UserService)FindUser(userpoint *module.User)(*module.User,error){
	user,err:=ud.FindUser(userpoint)
	if err!=nil{
		fmt.Println("find user err:",err)
		return nil, err
	}
	return user,nil
}


func (us *UserService)ChangeUser(olduser,newuser *module.User)error{
	return ud.ChangeUser(olduser,newuser)
}

func (us *UserService)CreateAnimals(olduser,newuser *module.User)error{
	return ud.CreateAnimals(olduser,newuser)
}

func (us UserService)ChangeUserByAffairs(tx *gorm.DB,olduser,newuser *module.User)error{
	return ud.ChangeUserByAffairs(tx,olduser,newuser)
}

func (us *UserService)InsertFollows(uid ,followid,username string)error{
	return ud.InsertFollows(uid,followid,username)
}

func (us *UserService)DeleteFollows(uid,followid string)error{
	return ud.DeleteFollows(uid,followid)
}

func (us *UserService)FindFollowID(uid,followid string)error{
	return ud.FindFollowID(uid,followid)
}

func (us *UserService)FindLikes(newLikes *module.UserWithVideo)(*[]module.UserWithVideo,error) {
	return ud.FindLikes(newLikes)
}

func (us UserService)CreateLikes(newLikes *module.UserWithVideo)error {
	return ud.CreateLikes(newLikes)
}

func (us UserService)DeleteLikes(newLikes *module.UserWithVideo)error{
	return ud.DeleteLikes(newLikes)
}

func (us UserService)UpdateLikses(oldLikes,newLikes *module.UserWithVideo)error{
	return ud.UpdateLikes(oldLikes,newLikes)
}

func(us *UserService)Encrypt_md5(pwd string)(pwdSalt,pwHash string){
	rand.Seed(time.Now().Unix())
	pwdSalt=strconv.Itoa(rand.Intn(1000000))

	md5ctx:=md5.New()
	md5ctx.Write([]byte(pwd))
	md5ctx.Write([]byte(pwdSalt))
	pwHash=hex.EncodeToString(md5ctx.Sum(nil))
	return
}

//传入字符串中加入标识
func (us *UserService)Decode_md5(pwd string,olduser *module.User)bool{//返回pwHash
	pwdSalt:=olduser.PwdSalt

	md5ctx:=md5.New()
	md5ctx.Write([]byte(pwd))
	md5ctx.Write([]byte(pwdSalt))
	pwHash:=hex.EncodeToString(md5ctx.Sum(nil))

	return olduser.PwdHash==pwHash
}

